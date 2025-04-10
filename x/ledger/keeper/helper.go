// validation.go contains data logic validations across the LedgerKeeper and other module Keepers
package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func getAddress(s *string) (sdk.AccAddress, error) {
	addr, err := sdk.AccAddressFromBech32(*s)
	if err != nil || addr == nil {
		return nil, NewLedgerCodedError(ErrCodeInvalidField, "nft_address")
	}

	return addr, nil
}
