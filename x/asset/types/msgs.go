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
		return fmt.Errorf("invalid asset: %w", err)
	}

	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return fmt.Errorf("invalid owner: %w", err)
	}

	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		return fmt.Errorf("invalid signer: %w", err)
	}

	return nil
}

func (msg MsgCreateAssetClass) ValidateBasic() error {
	if err := msg.AssetClass.Validate(); err != nil {
		return fmt.Errorf("invalid asset class: %w", err)
	}

	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		return fmt.Errorf("invalid signer: %w", err)
	}

	return nil
}

func (msg MsgCreatePool) ValidateBasic() error {
	if err := msg.Pool.Validate(); err != nil {
		return fmt.Errorf("invalid pool: %w", err)
	}

	if len(msg.Assets) == 0 {
		return fmt.Errorf("assets cannot be empty")
	}

	seen := make(map[AssetKey]int)
	for i, asset := range msg.Assets {
		if err := asset.Validate(); err != nil {
			return fmt.Errorf("invalid asset at index %d: %w", i, err)
		}
		if j, found := seen[*asset]; found {
			return fmt.Errorf("duplicate asset at index %d and %d", j, i)
		}
		seen[*asset] = i
	}

	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		return fmt.Errorf("invalid signer: %w", err)
	}

	return nil
}

func (msg MsgCreateTokenization) ValidateBasic() error {
	if err := msg.Token.Validate(); err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	if err := msg.Asset.Validate(); err != nil {
		return fmt.Errorf("invalid asset: %w", err)
	}

	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		return fmt.Errorf("invalid signer: %w", err)
	}

	return nil
}

func (msg MsgCreateSecuritization) ValidateBasic() error {
	if msg.Id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	if len(msg.Pools) == 0 {
		return fmt.Errorf("pools cannot be empty")
	}

	for i, pool := range msg.Pools {
		if len(pool) == 0 {
			return fmt.Errorf("invalid pool at index %d: cannot be empty", i)
		}
	}

	if len(msg.Tranches) == 0 {
		return fmt.Errorf("tranches cannot be empty")
	}

	for i, tranche := range msg.Tranches {
		if tranche == nil {
			return fmt.Errorf("invalid tranche at index %d: cannot be nil", i)
		}
		if err := tranche.Validate(); err != nil {
			return fmt.Errorf("invalid tranche at index %d: %w", i, err)
		}
	}

	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		return fmt.Errorf("invalid signer: %w", err)
	}

	return nil
}

// ValidateBasic performs basic validation on the MsgBurnAsset message.
func (msg MsgBurnAsset) ValidateBasic() error {
	if msg.Asset.ClassId == "" {
		return fmt.Errorf("class id cannot be empty")
	}

	if msg.Asset.Id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	if err := msg.Asset.Validate(); err != nil {
		return fmt.Errorf("invalid asset: %w", err)
	}

	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		return fmt.Errorf("invalid signer: %w", err)
	}

	return nil
}
