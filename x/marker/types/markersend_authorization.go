package types

import (
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authz "github.com/provenance-io/provenance/x/authz/exported"
)


var (
	_ authz.Authorization = &MarkerSendAuthorization{}
)

// NewSendAuthorization creates a new SendAuthorization object.
func NewMarkerSendAuthorization(spendLimit sdk.Coins) *MarkerSendAuthorization {
	return &MarkerSendAuthorization{
		SpendLimit: spendLimit,
	}
}

// MethodName implements Authorization.MethodName.
func (authorization MarkerSendAuthorization) MethodName() string {
	return "/provenance.marker.v1.Msg/Transfer"
}

// Accept implements Authorization.Accept.
func (authorization MarkerSendAuthorization) Accept(ctx sdk.Context, msg sdk.ServiceMsg) (updated authz.Authorization, delete bool, err error) {
	if reflect.TypeOf(msg.Request) == reflect.TypeOf(&MsgTransferRequest{}) {
		msg, ok := msg.Request.(*MsgTransferRequest)
		if ok {
			limitLeft, isNegative := authorization.SpendLimit.SafeSub(sdk.Coins{msg.Amount})
			if isNegative {
				return nil, false, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "requested amount is more than spend limit")
			}
			if limitLeft.IsZero() {
				return nil, true, nil
			}

			return &MarkerSendAuthorization{SpendLimit: limitLeft}, false, nil
		}
	}
	return nil, false, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "type mismatch")
}

// ValidateBasic implements Authorization.ValidateBasic.
func (authorization MarkerSendAuthorization) ValidateBasic() error {
	if authorization.SpendLimit == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "spend limit cannot be nil")
	}
	if !authorization.SpendLimit.IsAllPositive() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "spend limit cannot be negitive")
	}
	return nil
}
