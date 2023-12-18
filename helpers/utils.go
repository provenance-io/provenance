package helpers

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// ExitCodeError contains the exit code for cmd exit.
type ExitCodeError int

func (e ExitCodeError) Error() string {
	return fmt.Sprintf("exit code: %d", e)
}

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

// Keys returns the unordered keys of the given map.
// This is the same as the experimental maps.Keys function.
// As of writing, that isn't in the standard library version yet. Once it is, remove this and switch to that.
func Keys[M ~map[K]V, K comparable, V any](m M) []K {
	rv := make([]K, 0, len(m))
	for k := range m {
		rv = append(rv, k)
	}
	return rv
}
