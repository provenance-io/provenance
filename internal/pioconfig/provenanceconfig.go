package pioconfig

import (
	"fmt"
)

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

func SetProvenanceConfig(customDenom string, msgFeeFloorGasPrice int64) {
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
			// for backwards compatibility when these flags were not around, nhash will maintain behavior.
			ProvenanceMinGasPrices: fmt.Sprintf("%v", defaultMinGasPrices) + defaultFeeDenom,
			MsgFeeFloorGasPrice:    defaultMinGasPrices,
			BondDenom:              defaultBondDenom,
		}
	}
}

func GetProvenanceConfig() ProvenanceConfig {
	if provConfig == nil {
		panic("Provenance config should have been set.")
	}
	return *provConfig
}
