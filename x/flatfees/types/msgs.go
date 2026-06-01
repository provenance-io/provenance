package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AllRequestMsgs defines all the Msg*Request messages.
var AllRequestMsgs = []sdk.Msg{
	(*MsgUpdateParamsRequest)(nil),
	(*MsgUpdateConversionFactorRequest)(nil),
	(*MsgUpdateMsgFeesRequest)(nil),
	(*MsgAddOracleAddressRequest)(nil),
	(*MsgRemoveOracleAddressRequest)(nil),
}

// ValidateBasic performs basic validation for MsgUpdateParamsRequest.
func (m MsgUpdateParamsRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return fmt.Errorf("invalid authority: %w", err)
	}
	if err := m.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	return nil
}

// ValidateBasic performs basic validation for MsgUpdateConversionFactorRequest.
func (m MsgUpdateConversionFactorRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return fmt.Errorf("invalid authority: %w", err)
	}
	if err := m.ConversionFactor.Validate(); err != nil {
		return fmt.Errorf("invalid conversion factor: %w", err)
	}
	return nil
}

// ValidateBasic performs basic validation for MsgUpdateMsgFeesRequest.
func (m MsgUpdateMsgFeesRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return fmt.Errorf("invalid authority: %w", err)
	}

	if len(m.ToSet) == 0 && len(m.ToUnset) == 0 {
		return fmt.Errorf("at least one entry to set or unset must be provided: empty request")
	}

	seenSet := make(map[string]int)
	for i, msgFee := range m.ToSet {
		if err := msgFee.Validate(); err != nil {
			return fmt.Errorf("invalid ToSet[%d]: %w", i, err)
		}
		if j, ok := seenSet[msgFee.MsgTypeUrl]; ok {
			return fmt.Errorf("duplicate msg type url %q found in ToSet[%d] and ToSet[%d]", msgFee.MsgTypeUrl, j, i)
		}
		seenSet[msgFee.MsgTypeUrl] = i
	}

	seenUnset := make(map[string]int)
	for i, url := range m.ToUnset {
		if err := ValidateMsgTypeURL(url); err != nil {
			return fmt.Errorf("invalid ToUnset[%d]: %w", i, err)
		}
		if j, ok := seenUnset[url]; ok {
			return fmt.Errorf("duplicate msg type url %q found in ToUnset[%d] and ToUnset[%d]", url, j, i)
		}
		if j, ok := seenSet[url]; ok {
			return fmt.Errorf("duplicate msg type url %q found in ToSet[%d] and ToUnset[%d]", url, j, i)
		}
		seenUnset[url] = i
	}

	return nil
}

// ValidateBasic performs basic validation on MsgAddOracleAddressRequest
func (m MsgAddOracleAddressRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return fmt.Errorf("invalid authority: %w", err)
	}
	if _, err := sdk.AccAddressFromBech32(m.OracleAddress); err != nil {
		return fmt.Errorf("invalid oracle address: %w", err)
	}
	return nil
}

// ValidateBasic performs basic validation on MsgRemoveOracleAddressRequest
func (m MsgRemoveOracleAddressRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return fmt.Errorf("invalid authority: %w", err)
	}
	if _, err := sdk.AccAddressFromBech32(m.OracleAddress); err != nil {
		return fmt.Errorf("invalid oracle address: %w", err)
	}
	return nil
}
