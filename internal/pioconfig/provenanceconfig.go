package pioconfig

import (
	"fmt"
	"sync"
)

var lock = &sync.Mutex{}

const (
	// DefaultBondDenom is the denomination of coin to use for bond/staking
	// should only be via provConfig variable
	defaultBondDenom = "nhash" // nano-hash
	// DefaultFeeDenom is the denomination of coin to use for fees
	defaultFeeDenom = "nhash" // nano-hash
	// DefaultMinGasPrices is the minimum gas prices integer value only.
	defaultMinGasPrices = 1905
	// DefaultReDnmString is the allowed denom regex expression
	DefaultReDnmString = `[a-zA-Z][a-zA-Z0-9/\-\.]{2,127}`
)

type ProvenanceConfig struct {
	FeeDenom               string
	ProvenanceMinGasPrices string // Node level config that provenance binary can set and enforce across the board
	// e.g. provenance enforces 1905nhash across the board, also for now it will mirror MsgFeeFloorGasPrice.
	MsgFeeFloorGasPrice int64 // Msg fee ante handlers and code use this for their calculations, this ***ONLY SETS***
	// the default param(see method DefaultFloorGasPrice), all calculated values are still from msg fee module PARAMS.
	// for that module, if the param is changed via governance then the code will pick the new value.(should pick that up from module param)
	BondDenom string // Also referred to as Staking Denom sometimes.
}

var provConfig *ProvenanceConfig

// SetProvenanceConfig in running the app it is called once from root.go. We decided not to seal it because we have tests,
// which set the Config to test certain msg fee flows.
// But the contract remains that this will be called once from root.go while starting up.
func SetProvenanceConfig(customDenom string, msgFeeFloorGasPrice int64) {
	// custom denom (e.g. vspn) to be used in custom zones, if not passed in will default to nhash,
	// to preserve backwards compatible behaviour.
	if len(customDenom) > 0 {
		provConfig = &ProvenanceConfig{
			FeeDenom:               customDenom,
			ProvenanceMinGasPrices: fmt.Sprintf("%v", msgFeeFloorGasPrice) + customDenom,
			MsgFeeFloorGasPrice:    msgFeeFloorGasPrice,
			BondDenom:              customDenom,
		}
	} else {
		provConfig = &ProvenanceConfig{
			FeeDenom: defaultFeeDenom,
			// for backwards compatibility when these flags were not around, nhash will maintain behaviour.
			ProvenanceMinGasPrices: fmt.Sprintf("%v", defaultMinGasPrices) + defaultFeeDenom,
			MsgFeeFloorGasPrice:    defaultMinGasPrices,
			BondDenom:              defaultBondDenom,
		}
	}
}

// GetProvenanceConfig get ProvenanceConfig
func GetProvenanceConfig() ProvenanceConfig {
	// check that config is set
	if provConfig == nil {
		panic("Config should have been set explicitly.")
	}
	return *provConfig
}
