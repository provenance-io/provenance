package helpers

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// MustValAddressFromBech32 calls sdk.ValAddressFromBech32 and panics on error.
func MustValAddressFromBech32(operatorAddr string) sdk.ValAddress {
	addr, err := sdk.ValAddressFromBech32(operatorAddr)
	if err != nil {
		panic(err)
	}
	return addr
}

// GetOperatorAddr returns the validator's operator address.
func GetOperatorAddr(val stakingtypes.ValidatorI) (sdk.ValAddress, error) {
	return sdk.ValAddressFromBech32(val.GetOperator())
}

// MustGetOperatorAddr returns the validator's operator address and panics on ierror.
func MustGetOperatorAddr(val stakingtypes.ValidatorI) sdk.ValAddress {
	return MustValAddressFromBech32(val.GetOperator())
}

// ValidateBasic calls the ValidateBasic on the provided msg if it has that method, otherwise returns nil.
func ValidateBasic(msg sdk.Msg) error {
	vmsg, ok := msg.(sdk.HasValidateBasic)
	if ok {
		return vmsg.ValidateBasic()
	}
	return nil
}
