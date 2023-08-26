package exchange

import (
	"fmt"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is the name of the exchange module.
	ModuleName = "exchange"

	// StoreKey is the store key string for the exchange module.
	StoreKey = ModuleName
)

// GetMarketAddress returns the module account address for the given marketID.
func GetMarketAddress(marketID uint32) sdk.AccAddress {
	return sdk.AccAddress(crypto.AddressHash([]byte(fmt.Sprintf("%s/%d", ModuleName, marketID))))
}
