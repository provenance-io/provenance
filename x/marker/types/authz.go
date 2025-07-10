package types

import (
	"context"
	"errors"
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

var (
	_ authz.Authorization                = &MarkerTransferAuthorization{}
	_ authz.Authorization                = &MultiAuthorization{}
	_ codectypes.UnpackInterfacesMessage = &MultiAuthorization{}
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
	anyAuths := make([]*codectypes.Any, len(subAuthorizations))
	for i, auth := range subAuthorizations {
		authValue, err := codectypes.NewAnyWithValue(auth)
		if err != nil {
			return nil, sdkerrors.ErrInvalidRequest.Wrapf("failed to pack sub-authorization %d: %s", i, err)
		}
		anyAuths[i] = authValue
	}

	return &MultiAuthorization{
		MsgTypeUrl:        msgTypeURL,
		SubAuthorizations: anyAuths,
	}, nil
}

// MsgTypeURL implements Authorization.MsgTypeURL.
func (m MultiAuthorization) MsgTypeURL() string {
	return m.MsgTypeUrl
}

// Accept implements Authorization.Accept.
func (m MultiAuthorization) Accept(ctx context.Context, msg sdk.Msg) (authz.AcceptResponse, error) {
	if m.MsgTypeURL() != sdk.MsgTypeURL(msg) {
		return authz.AcceptResponse{}, errors.New("message type mismatch")
	}

	anyDeleteRequested := false
	updatedAuths := make([]*codectypes.Any, len(m.SubAuthorizations))
	anyUpdates := false

	for i, anyAuth := range m.SubAuthorizations {
		if anyAuth == nil {
			return authz.AcceptResponse{}, fmt.Errorf("sub-authorization %d is nil", i)
		}

		auth, ok := anyAuth.GetCachedValue().(authz.Authorization)
		if !ok || auth == nil {
			return authz.AcceptResponse{}, fmt.Errorf("sub-authorization %d not unpacked", i)
		}
		resp, err := auth.Accept(ctx, msg)
		if err != nil {
			return authz.AcceptResponse{}, fmt.Errorf("sub-authorization %d was not accepted: %w", i, err)
		}

		// If any sub-authorization is not accepted, the whole thing is not accepted and there's nothing more to do.
		if !resp.Accept {
			return authz.AcceptResponse{Accept: false}, nil
		}

		// If any sub-authorization requests delete, mark for full deletion
		if resp.Delete {
			anyDeleteRequested = true
		}
		// If the sub-authorization needs to be updated, we need to update it in this multi-authorization.
		if resp.Updated != nil {
			updatedAny, err := codectypes.NewAnyWithValue(resp.Updated)
			if err != nil {
				return authz.AcceptResponse{}, fmt.Errorf("failed to pack updated sub-authorization %w", err)
			}
			updatedAuths[i] = updatedAny
			anyUpdates = true
		} else {
			updatedAuths[i] = anyAuth
		}
	}

	rv := authz.AcceptResponse{Accept: true, Delete: anyDeleteRequested}
	if anyUpdates && !anyDeleteRequested {
		rv.Updated = &MultiAuthorization{MsgTypeUrl: m.MsgTypeUrl, SubAuthorizations: updatedAuths}
	}
	return rv, nil
}

// UnpackInterfaces implements codectypes.UnpackInterfacesMessage
func (m *MultiAuthorization) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for i, anyAuth := range m.SubAuthorizations {
		if anyAuth == nil {
			return sdkerrors.ErrInvalidRequest.Wrapf("sub-authorization %d is nil", i)
		}
		var auth authz.Authorization
		if err := unpacker.UnpackAny(anyAuth, &auth); err != nil {
			return sdkerrors.ErrInvalidType.Wrapf("failed to unpack sub-authorization %s", err)
		}
	}
	return nil
}

// ValidateBasic implements Authorization.ValidateBasic.
func (m MultiAuthorization) ValidateBasic() error {
	if m.MsgTypeUrl == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("message type URL cannot be empty")
	}
	authCount := len(m.SubAuthorizations)
	if authCount < MinSubAuthorizations {
		return sdkerrors.ErrInvalidRequest.Wrapf("must have at least %d sub-authorizations, got %d", MinSubAuthorizations, authCount)
	}
	if authCount > MaxSubAuthorizations {
		return sdkerrors.ErrInvalidRequest.Wrapf("cannot have more than %d sub-authorizations, got %d", MaxSubAuthorizations, authCount)
	}
	for i, anyAuth := range m.SubAuthorizations {
		if anyAuth == nil {
			return sdkerrors.ErrInvalidRequest.Wrapf("sub-authorization %d is nil", i)
		}
		if anyAuth.TypeUrl == "" {
			return sdkerrors.ErrInvalidRequest.Wrapf("sub-authorization %d has empty type URL", i)
		}
		// Retrieve unpacked authorization from cached value
		cachedValue := anyAuth.GetCachedValue()
		if cachedValue == nil {
			return sdkerrors.ErrInvalidType.Wrapf("sub-authorization %d has not been unpacked", i)
		}
		auth, ok := cachedValue.(authz.Authorization)
		if !ok {
			return sdkerrors.ErrInvalidType.Wrapf("sub-authorization %d is not an Authorization", i)
		}
		// Type safety check
		if auth.MsgTypeURL() != m.MsgTypeUrl {
			return sdkerrors.ErrInvalidType.Wrapf(
				"sub-authorization %d has msg type %q, expected %q",
				i, auth.MsgTypeURL(), m.MsgTypeUrl,
			)
		}
		// Prevent nested MultiAuthorization
		if _, isMulti := auth.(*MultiAuthorization); isMulti {
			return sdkerrors.ErrInvalidType.Wrapf(
				"nested MultiAuthorization not allowed for sub-authorization %d", i,
			)
		}
		// Recursive validation
		if err := auth.ValidateBasic(); err != nil {
			return sdkerrors.ErrInvalidType.Wrapf(
				"sub-authorization %d failed basic validation: %v", i, err,
			)
		}
	}

	return nil
}
