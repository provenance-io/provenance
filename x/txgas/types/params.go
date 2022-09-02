package types

import (
	"fmt"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Default parameter namespace
const (
	DefaultTxGasLimit = 4_000_000
)

// Parameter store keys
var (
	ParamStoreKeyTxGasLimit = []byte("TxGasLimit")
)

type Params struct {
	TxGasLimit uint64 `protobuf:"varint,1,opt,name=tx_gas_limit,json=txGasLimit,proto3" json:"tx_gas_limit,omitempty" yaml:"tx_gas_limit"`
}

func (params *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyTxGasLimit, &params.TxGasLimit, validateTxGasLimit),
	}
}

func validateTxGasLimit(i interface{}) error {
	_, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
