package types

import (
	"context"
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

var (
	_ authz.Authorization = &MarkerTransferAuthorization{}
	_ authz.Authorization = &MultiAuthorization{}
)

const (
	// MaxSubAuthorizations defines the maximum number of sub-authorizations allowed
	MaxSubAuthorizations = 10
	// MinSubAuthorizations defines the minimum number of sub-authorizations required
	MinSubAuthorizations = 2
)

// NewMarkerTransferAuthorization creates a new MarkerTransferAuthorization object.
func NewMarkerTransferAuthorization(transferLimit sdk.Coins, allowed []sdk.AccAddress) *MarkerTransferAuthorization {
	allowedAddrs := toBech32Addresses(allowed)
	return &MarkerTransferAuthorization{
		TransferLimit: transferLimit,
		AllowList:     allowedAddrs,
	}
}

// MsgTypeURL implements Authorization.MsgTypeURL.
func (a MarkerTransferAuthorization) MsgTypeURL() string {
	return sdk.MsgTypeURL(&MsgTransferRequest{})
}

// Accept implements Authorization.Accept.
func (a MarkerTransferAuthorization) Accept(_ context.Context, msg sdk.Msg) (authz.AcceptResponse, error) {
	switch msg := msg.(type) {
	case *MsgTransferRequest:
		toAddress := msg.ToAddress
		limitLeft, isNegative := a.DecreaseTransferLimit(msg.Amount)
		if isNegative {
			return authz.AcceptResponse{}, sdkerrors.ErrInsufficientFunds.Wrap("requested amount is more than spend limit")
		}
		shouldDelete := false
		if limitLeft.IsZero() {
			shouldDelete = true
		}

		isAddrExists := false
		allowedList := a.GetAllowList()

		for _, addr := range allowedList {
			if addr == toAddress {
				isAddrExists = true
				break
			}
		}

		if len(allowedList) > 0 && !isAddrExists {
			return authz.AcceptResponse{}, sdkerrors.ErrUnauthorized.Wrapf("cannot send to %s address", toAddress)
		}

		return authz.AcceptResponse{Accept: true, Delete: shouldDelete, Updated: &MarkerTransferAuthorization{TransferLimit: limitLeft}}, nil
	default:
		return authz.AcceptResponse{}, sdkerrors.ErrInvalidType.Wrap("type mismatch")
	}
}

// ValidateBasic implements Authorization.ValidateBasic.
func (a MarkerTransferAuthorization) ValidateBasic() error {
	if err := a.TransferLimit.Validate(); err != nil {
		return sdkerrors.ErrInvalidCoins.Wrapf("invalid transfer limit: %v", err)
	}
	if a.TransferLimit.IsZero() {
		return sdkerrors.ErrInvalidCoins.Wrap("invalid transfer limit: cannot be zero")
	}

	found := make(map[string]bool, 0)
	for i, addr := range a.AllowList {
		if _, err := sdk.AccAddressFromBech32(addr); err != nil {
			return sdkerrors.ErrInvalidAddress.Wrapf("invalid allow list entry [%d] %q: %v", i, addr, err)
		}
		if found[addr] {
			return ErrDuplicateEntry.Wrapf("invalid allow list entry [%d] %s", i, addr)
		}
		found[addr] = true
	}

	return nil
}

// DecreaseTransferLimit will return the decreased transfer limit and if it is negative
func (a MarkerTransferAuthorization) DecreaseTransferLimit(amount sdk.Coin) (sdk.Coins, bool) {
	return a.TransferLimit.SafeSub(amount)
}

func toBech32Addresses(allowed []sdk.AccAddress) []string {
	if len(allowed) == 0 {
		return nil
	}

	allowedAddrs := make([]string, len(allowed))
	for i, addr := range allowed {
		allowedAddrs[i] = addr.String()
	}
	return allowedAddrs
}

// NewMultiAuthorization creates a new MultiAuthorization
func NewMultiAuthorization(msgTypeURL string, subAuthorizations ...authz.Authorization) (*MultiAuthorization, error) {
	if msgTypeURL == "" {
		return nil, fmt.Errorf("message type URL cannot be empty")
	}
	if len(subAuthorizations) < MinSubAuthorizations {
		return nil, fmt.Errorf("must have at least %d sub-authorization", MinSubAuthorizations)
	}
	if len(subAuthorizations) > MaxSubAuthorizations {
		return nil, fmt.Errorf("cannot have more than %d sub-authorizations", MaxSubAuthorizations)
	}
	// Convert to Any and validate
	anyAuths := make([]*codectypes.Any, len(subAuthorizations))
	for i, auth := range subAuthorizations {
		if auth == nil {
			return nil, fmt.Errorf("sub-authorization %d cannot be nil", i)
		}
		if auth.MsgTypeURL() != msgTypeURL {
			return nil, fmt.Errorf("sub-authorization %d has different msg type URL: expected %s, got %s",
				i, msgTypeURL, auth.MsgTypeURL())
		}
		if _, isMulti := auth.(*MultiAuthorization); isMulti {
			return nil, fmt.Errorf("cannot have nested MultiAuthorization")
		}
		any, err := codectypes.NewAnyWithValue(auth)
		if err != nil {
			return nil, fmt.Errorf("failed to pack sub-authorization %d: %w", i, err)
		}
		anyAuths[i] = any
	}
	return &MultiAuthorization{
		MsgTypeUrl:        msgTypeURL,
		SubAuthorizations: anyAuths,
	}, nil
}

// MsgTypeURL implements the Authorization interface
func (m MultiAuthorization) MsgTypeURL() string {
	return m.MsgTypeUrl
}

// Accept implements the Authorization interface
func (m MultiAuthorization) Accept(ctx context.Context, msg sdk.Msg) (authz.AcceptResponse, error) {
	if m.MsgTypeURL() != sdk.MsgTypeURL(msg) {
		return authz.AcceptResponse{}, sdkerrors.ErrInvalidType.Wrapf("message type %s does not match authorization type %s", sdk.MsgTypeURL(msg), m.MsgTypeURL())
	}
	if len(m.SubAuthorizations) == 0 {
		return authz.AcceptResponse{}, fmt.Errorf("no sub-authorizations available")
	}
	// Process all sub-authorizations
	updatedSubAuths := make([]*codectypes.Any, len(m.SubAuthorizations))
	hasUpdates := false
	hasDeletes := false
	for i, anyAuth := range m.SubAuthorizations {
		if anyAuth == nil {
			return authz.AcceptResponse{}, fmt.Errorf("sub-authorization %d is nil", i)
		}
		cached := anyAuth.GetCachedValue()
		if cached == nil {
			return authz.AcceptResponse{}, fmt.Errorf("sub-authorization %d not unpacked", i)
		}
		subAuth, ok := cached.(authz.Authorization)
		if !ok {
			return authz.AcceptResponse{}, fmt.Errorf("sub-authorization %d is not an Authorization", i)
		}
		resp, err := subAuth.Accept(ctx, msg)
		if err != nil {
			return authz.AcceptResponse{}, fmt.Errorf("sub-authorization %d rejected: %w", i, err)
		}
		if !resp.Accept {
			return authz.AcceptResponse{Accept: false}, nil
		}
		if resp.Delete {
			hasDeletes = true
		}
		if resp.Updated != nil {
			hasUpdates = true
			updatedAny, err := codectypes.NewAnyWithValue(resp.Updated)
			if err != nil {
				return authz.AcceptResponse{}, fmt.Errorf("failed to pack updated sub-authorization %d: %w", i, err)
			}
			updatedSubAuths[i] = updatedAny
		} else {
			updatedSubAuths[i] = anyAuth
		}
	}
	if hasDeletes {
		return authz.AcceptResponse{Accept: true, Delete: true}, nil
	}
	if hasUpdates {
		updatedMultiAuth := &MultiAuthorization{
			MsgTypeUrl:        m.MsgTypeUrl,
			SubAuthorizations: updatedSubAuths,
		}
		return authz.AcceptResponse{Accept: true, Updated: updatedMultiAuth}, nil
	}
	return authz.AcceptResponse{Accept: true}, nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m *MultiAuthorization) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	if m == nil {
		return fmt.Errorf("MultiAuthorization is nil")
	}
	for i, anyAuth := range m.SubAuthorizations {
		if anyAuth == nil {
			return fmt.Errorf("sub-authorization %d is nil", i)
		}
		var subAuth authz.Authorization
		if err := unpacker.UnpackAny(anyAuth, &subAuth); err != nil {
			return fmt.Errorf("failed to unpack sub-authorization %d: %w", i, err)
		}
	}
	return nil
}

// ValidateBasic implements basic validation
func (m MultiAuthorization) ValidateBasic() error {
	if m.MsgTypeUrl == "" {
		return fmt.Errorf("message type URL cannot be empty")
	}
	authCount := len(m.SubAuthorizations)
	if authCount < MinSubAuthorizations {
		return fmt.Errorf("must have at least %d sub-authorization", MinSubAuthorizations)
	}
	if authCount > MaxSubAuthorizations {
		return fmt.Errorf("cannot have more than %d sub-authorizations", MaxSubAuthorizations)
	}
	// Validate cached sub-authorizations if available
	for i, anyAuth := range m.SubAuthorizations {
		if anyAuth == nil {
			return fmt.Errorf("sub-authorization %d cannot be nil", i)
		}
		if cached := anyAuth.GetCachedValue(); cached != nil {
			subAuth, ok := cached.(authz.Authorization)
			if !ok {
				return fmt.Errorf("sub-authorization %d is not a valid Authorization", i)
			}
			if subAuth.MsgTypeURL() != m.MsgTypeUrl {
				return fmt.Errorf("sub-authorization %d has different msg type URL", i)
			}
			if _, isMulti := subAuth.(*MultiAuthorization); isMulti {
				return fmt.Errorf("cannot have nested MultiAuthorization at %d", i)
			}
			if validator, ok := subAuth.(interface{ ValidateBasic() error }); ok {
				if err := validator.ValidateBasic(); err != nil {
					return fmt.Errorf("sub-authorization %d is invalid: %w", i, err)
				}
			}
		}
	}
	return nil
}
