package types

import (
	"errors"
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
	var errs []error

	if err := msg.Asset.Validate(); err != nil {
		errs = append(errs, NewErrCodeInvalidField("asset", "%s", err))
	}

	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		errs = append(errs, NewErrCodeInvalidField("owner", "%s", err))
	}

	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		errs = append(errs, NewErrCodeInvalidField("signer", "%s", err))
	}

	return errors.Join(errs...)
}

func (msg MsgCreateAssetClass) ValidateBasic() error {
	var errs []error

	if err := msg.AssetClass.Validate(); err != nil {
		errs = append(errs, NewErrCodeInvalidField("asset_class", "%s", err))
	}

	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		errs = append(errs, NewErrCodeInvalidField("signer", "%s", err))
	}

	return errors.Join(errs...)
}

func (msg MsgCreatePool) ValidateBasic() error {
	var errs []error

	if err := msg.Pool.Validate(); err != nil {
		errs = append(errs, NewErrCodeInvalidField("pool", "%s", err))
	}

	if len(msg.Assets) == 0 {
		errs = append(errs, NewErrCodeInvalidField("assets", "cannot be empty"))
	}

	seen := make(map[AssetKey]int)
	for i, asset := range msg.Assets {
		if err := asset.Validate(); err != nil {
			errs = append(errs, NewErrCodeInvalidField(fmt.Sprintf("assets[%d]", i), "%s", err))
		} else {
			if j, found := seen[*asset]; found {
				errs = append(errs, NewErrCodeInvalidField("assets", "duplicate asset at index %d and %d", j, i))
			} else {
				seen[*asset] = i
			}
		}
	}

	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		errs = append(errs, NewErrCodeInvalidField("signer", "%s", err))
	}

	return errors.Join(errs...)
}

func (msg MsgCreateTokenization) ValidateBasic() error {
	var errs []error

	if err := msg.Token.Validate(); err != nil {
		errs = append(errs, NewErrCodeInvalidField("token", "%s", err))
	}

	if err := msg.Asset.Validate(); err != nil {
		errs = append(errs, NewErrCodeInvalidField("asset", "%s", err))
	}

	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		errs = append(errs, NewErrCodeInvalidField("signer", "%s", err))
	}

	return errors.Join(errs...)
}

func (msg MsgCreateSecuritization) ValidateBasic() error {
	var errs []error

	if msg.Id == "" {
		errs = append(errs, NewErrCodeInvalidField("id", "cannot be empty"))
	}

	if len(msg.Pools) == 0 {
		errs = append(errs, NewErrCodeInvalidField("pools", "cannot be empty"))
	}

	for i, pool := range msg.Pools {
		if len(pool) == 0 {
			errs = append(errs, NewErrCodeInvalidField(fmt.Sprintf("pools[%d]", i), "cannot be empty"))
		}
	}

	if len(msg.Tranches) == 0 {
		errs = append(errs, NewErrCodeInvalidField("tranches", "cannot be empty"))
	}

	for i, tranche := range msg.Tranches {
		if tranche == nil {
			errs = append(errs, NewErrCodeInvalidField(fmt.Sprintf("tranches[%d]", i), "cannot be nil"))
		} else {
			if err := tranche.Validate(); err != nil {
				errs = append(errs, NewErrCodeInvalidField(fmt.Sprintf("tranches[%d]", i), "%s", err))
			}
		}
	}

	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		errs = append(errs, NewErrCodeInvalidField("signer", "%s", err))
	}

	return errors.Join(errs...)
}

// ValidateBasic performs basic validation on the MsgBurnAsset message.
func (msg MsgBurnAsset) ValidateBasic() error {
	var errs []error

	if err := msg.Asset.Validate(); err != nil {
		errs = append(errs, NewErrCodeInvalidField("asset", "%s", err))
	}

	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		errs = append(errs, NewErrCodeInvalidField("signer", "%s", err))
	}

	return errors.Join(errs...)
}
