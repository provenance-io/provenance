package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

var (
	_ authz.Authorization = &MarkerTransferAuthorization{}
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
