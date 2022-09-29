package pioconfig

import (
	"fmt"
)

const (
	// defaultBondDenom is the denomination of coin to use for bond/staking
	// should only be via provConfig variable
	defaultBondDenom = "nhash" // nano-hash
	// defaultFeeDenom is the (default) denomination of coin to use for fees
	defaultFeeDenom = "nhash" // nano-hash
	// defaultMinGasPrices is the (default) minimum gas prices integer value only
	defaultMinGasPrices = 1905
	// DefaultReDnmString is the allowed denom regex expression
	DefaultReDnmString = `[a-zA-Z][a-zA-Z0-9/\-\.]{2,127}`
)

type ProvenanceConfig struct {
	FeeDenom               string
	ProvenanceMinGasPrices string // maps to defaultMinGasPrices in previous code,Node level config that provenance binary set's from appOpts.
	// Current it will mirror MsgFeeFloorGasPrice.
	MsgFeeFloorGasPrice int64 // Msg fee ante handlers and code use this for their calculations, this ***ONLY SETS***
	// the default param(see method DefaultFloorGasPrice), all calculated values are still from msg fee module PARAMS.
	// for that module, if the param is changed via governance then the code will pick the new value.(should pick that up from module param)
	BondDenom     string // Also referred to as Staking Denom sometimes.
	MsgFloorDenom string // MsgFloorDenom should always be the same Fee Denom, but maybe useful for tests.
}

var provConfig *ProvenanceConfig

// SetProvenanceConfig in running the app it is called once from root.go. We decided not to seal it because we have tests,
// which set the Config to test certain msg fee flows.
// But the contract remains that this will be called once from root.go while starting up.
func SetProvenanceConfig(customDenom string, msgFeeFloorGasPrice int64) {
	if len(customDenom) > 0 && customDenom != defaultFeeDenom {
		provConfig = &ProvenanceConfig{
			FeeDenom:               customDenom,
			ProvenanceMinGasPrices: fmt.Sprintf("%v", msgFeeFloorGasPrice) + customDenom,
			MsgFeeFloorGasPrice:    msgFeeFloorGasPrice,
			BondDenom:              customDenom,
			MsgFloorDenom:          customDenom,
		}
	} else {
		provConfig = &ProvenanceConfig{
			FeeDenom: defaultFeeDenom,
			// for backwards compatibility when these flags were not around, nhash will maintain behavior.
			ProvenanceMinGasPrices: fmt.Sprintf("%v", defaultMinGasPrices) + defaultFeeDenom,
			MsgFeeFloorGasPrice:    defaultMinGasPrices,
			BondDenom:              defaultBondDenom,
			MsgFloorDenom:          defaultFeeDenom,
		}
	}
}

// GetProvenanceConfig get ProvenanceConfig
func GetProvenanceConfig() ProvenanceConfig {
	if provConfig != nil {
		return *provConfig
	}
	// Should get empty config if not set, several things in app wiring will fail fast if this not set so not too worried.
	return ProvenanceConfig{}
}
