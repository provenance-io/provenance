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
	DefaultMinGasPrices = 1905
	// DefaultReDnmString is the allowed denom regex expression
	DefaultReDnmString = `[a-zA-Z][a-zA-Z0-9/\-\.]{2,127}`
)

type ProvenanceConfig struct {
	FeeDenom string
	// Min gas price as a node level config, i.e each node before could set it's gas price in app.toml
	MinGasPrices        string
	MsgFeeFloorGasPrice int64 // Msg fee antehandlers and code use this for their calc's, this only sets the default param
	// for that module, if the param is changed then the code will(should pick that up from module param)
	BondDenom string
	set       bool
}

var provConfig *ProvenanceConfig

func SetProvenanceConfig(customDenom string, msgFeeFloorGasPrice int64) {
	lock.Lock()
	defer lock.Unlock()
	if len(customDenom) > 0 {
		provConfig = &ProvenanceConfig{
			FeeDenom:            customDenom,
			MinGasPrices:        fmt.Sprintf("%v", msgFeeFloorGasPrice) + customDenom,
			MsgFeeFloorGasPrice: msgFeeFloorGasPrice,
			BondDenom:           customDenom,
			set:                 true,
		}
	} else {
		provConfig = &ProvenanceConfig{
			FeeDenom:            defaultFeeDenom,
			MinGasPrices:        fmt.Sprintf("%v", DefaultMinGasPrices) + defaultFeeDenom,
			MsgFeeFloorGasPrice: DefaultMinGasPrices,
			BondDenom:           defaultBondDenom,
			set:                 true,
		}
	}
}

func GetProvenanceConfig() ProvenanceConfig {
	if provConfig == nil {
		SetProvenanceConfig("", DefaultMinGasPrices)
	}
	if !provConfig.set {
		panic("Accessing Provenance config before it is set is not allowed")
	}
	return *provConfig
}
