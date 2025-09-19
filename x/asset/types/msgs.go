package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AllRequestMsgs lists all Msg types for interface registration.
var AllRequestMsgs = []sdk.Msg{
	(*MsgBurnAsset)(nil),
	(*MsgCreateAsset)(nil),
	(*MsgCreateAssetClass)(nil),
	(*MsgCreatePool)(nil),
	(*MsgCreateTokenization)(nil),
	(*MsgCreateSecuritization)(nil),
}

func (msg MsgCreateAsset) ValidateBasic() error {
	if err := msg.Asset.Validate(); err != nil {
		return NewErrCodeInvalidField("asset", err.Error())
	}

	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return NewErrCodeInvalidField("owner", err.Error())
	}

	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		return NewErrCodeInvalidField("signer", err.Error())
	}

	return nil
}

func (msg MsgCreateAssetClass) ValidateBasic() error {
	if err := msg.AssetClass.Validate(); err != nil {
		return NewErrCodeInvalidField("asset_class", err.Error())
	}

	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		return NewErrCodeInvalidField("signer", err.Error())
	}

	return nil
}

func (msg MsgCreatePool) ValidateBasic() error {
	if err := msg.Pool.Validate(); err != nil {
		return NewErrCodeInvalidField("pool", err.Error())
	}

	if len(msg.Assets) == 0 {
		return NewErrCodeMissingField("assets")
	}

	seen := make(map[AssetKey]int)
	for i, asset := range msg.Assets {
		if err := asset.Validate(); err != nil {
			return NewErrCodeInvalidField(fmt.Sprintf("asset[%d]", i), err.Error())
		}
		if j, found := seen[*asset]; found {
			return NewErrCodeInvalidField("assets", fmt.Sprintf("duplicate asset at index %d and %d", j, i))
		}
		seen[*asset] = i
	}

	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		return NewErrCodeInvalidField("signer", err.Error())
	}

	return nil
}

func (msg MsgCreateTokenization) ValidateBasic() error {
	if err := msg.Token.Validate(); err != nil {
		return NewErrCodeInvalidField("token", err.Error())
	}

	if err := msg.Asset.Validate(); err != nil {
		return NewErrCodeInvalidField("asset", err.Error())
	}

	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		return NewErrCodeInvalidField("signer", err.Error())
	}

	return nil
}

func (msg MsgCreateSecuritization) ValidateBasic() error {
	if msg.Id == "" {
		return NewErrCodeMissingField("id")
	}

	if len(msg.Pools) == 0 {
		return NewErrCodeMissingField("pools")
	}

	for i, pool := range msg.Pools {
		if len(pool) == 0 {
			return NewErrCodeInvalidField(fmt.Sprintf("pools[%d]", i), "cannot be empty")
		}
	}

	if len(msg.Tranches) == 0 {
		return NewErrCodeMissingField("tranches")
	}

	for i, tranche := range msg.Tranches {
		if tranche == nil {
			return NewErrCodeInvalidField(fmt.Sprintf("tranches[%d]", i), "cannot be nil")
		}
		if err := tranche.Validate(); err != nil {
			return NewErrCodeInvalidField(fmt.Sprintf("tranches[%d]", i), err.Error())
		}
	}

	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		return NewErrCodeInvalidField("signer", err.Error())
	}

	return nil
}

// ValidateBasic performs basic validation on the MsgBurnAsset message.
func (msg MsgBurnAsset) ValidateBasic() error {
	if msg.Asset.ClassId == "" {
		return NewErrCodeMissingField("class_id")
	}

	if msg.Asset.Id == "" {
		return NewErrCodeMissingField("id")
	}

	if err := msg.Asset.Validate(); err != nil {
		return NewErrCodeInvalidField("asset", err.Error())
	}

	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		return NewErrCodeInvalidField("signer", err.Error())
	}

	return nil
}
