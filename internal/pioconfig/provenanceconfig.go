package pioconfig

import (
	"fmt"
)

const (
	// defaultFeeDenom is the (default) denomination of coin to use for fees
	defaultFeeDenom = "nhash" // nano-hash
	// defaultMinGasPricesAmount is the (default) minimum gas prices integer value only
	defaultMinGasPricesAmount = 0
	// DefaultReDnmString is the allowed denom regex expression
	DefaultReDnmString = `[a-zA-Z][a-zA-Z0-9/\-\.]{2,127}`

	// SimAppChainID hardcoded chainID for simulation.
	// Copied from cosmossdk.io/simapp/sim_test.go. We used to use this directly, but now its in a _test.go file.
	SimAppChainID = "simulation-app"
)

var provConfig *ProvConfig

type ProvConfig struct {
	// FeeDenom is the denom used to pay fees.
	FeeDenom string
	// BondDenom is the denom used for staking and delegation.
	BondDenom string
	// ProvMinGasPrices is the coin string of the minimum gas prices a node is allowed to accept.
	ProvMinGasPrices string
}

// SetProvConfig defines some config stuff specific to the Provenance blockchain.
// See also: GetProvConfig.
func SetProvConfig(feeAndBondDenom string) {
	if len(feeAndBondDenom) == 0 {
		feeAndBondDenom = defaultFeeDenom
	}
	provConfig = &ProvConfig{
		FeeDenom:         feeAndBondDenom,
		BondDenom:        feeAndBondDenom,
		ProvMinGasPrices: fmt.Sprintf("%d%s", defaultMinGasPricesAmount, feeAndBondDenom),
	}
}

// GetProvConfig get the current ProvConfig.
// See also: SetProvConfig.
func GetProvConfig() ProvConfig {
	if provConfig != nil {
		return *provConfig
	}
	return ProvConfig{
		FeeDenom:         defaultFeeDenom,
		BondDenom:        defaultFeeDenom,
		ProvMinGasPrices: fmt.Sprintf("%d%s", defaultMinGasPricesAmount, defaultFeeDenom),
	}
}
