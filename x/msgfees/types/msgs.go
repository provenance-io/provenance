package types

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (

	// AssessCustomMsgFeeBips is the hardcoded value for bips the recipient will receive while remainder will go to fee module
	// 5,000 bips is 50:50 to recipient and fee module
	AssessCustomMsgFeeBips = 5_000

	TypeAssessCustomMsgFee = "assess_custom_msg_fee"
)

// Compile time interface checks.
var (
	_ sdk.Msg = &MsgAssessCustomMsgFeeRequest{}
)

func NewMsgAssessCustomMsgFeeRequest(
	name string,
	amount sdk.Coin,
	recipient string,
	from string,
) MsgAssessCustomMsgFeeRequest {
	return MsgAssessCustomMsgFeeRequest{
		Name:      name,
		Amount:    amount,
		Recipient: recipient,
		From:      from,
	}
}

func (msg MsgAssessCustomMsgFeeRequest) ValidateBasic() error {
	if len(msg.Recipient) != 0 {
		_, err := sdk.AccAddressFromBech32(msg.Recipient)
		if err != nil {
			return err
		}
	}
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return err
	}
	if msg.Amount.Denom != UsdDenom && msg.Amount.Denom != NhashDenom {
		return fmt.Errorf("denom must be in usd or nhash : %s", msg.Amount.Denom)
	}
	if !msg.Amount.IsPositive() {
		return errors.New("amount must be greater than zero")
	}
	return nil
}

func (msg MsgAssessCustomMsgFeeRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgAssessCustomMsgFeeRequest) GetSignerBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// Route returns the module route
func (msg MsgAssessCustomMsgFeeRequest) Route() string {
	return ModuleName
}

// Type returns the type name for this msg
func (msg MsgAssessCustomMsgFeeRequest) Type() string {
	return TypeAssessCustomMsgFee
}
