package types

import (
	"fmt"

	"github.com/provenance-io/provenance/internal/pioconfig"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// DefaultFloorGasPrice to differentiate between base fee and additional fee when additional fee is in same denom as default base denom i.e nhash
// cannot be a const unfortunately because it's a custom type.
func DefaultFloorGasPrice() sdk.Coin {
	return sdk.Coin{
		Amount: sdk.NewInt(pioconfig.GetProvenanceConfig().MsgFeeFloorGasPrice),
		Denom:  pioconfig.GetProvenanceConfig().MsgFloorDenom,
	}
}

func DefaultMsgFee() sdk.Coin {
	return sdk.Coin{
		Amount: sdk.NewInt(1_000_000_000), // for now each tx takes 1hash or equivalent
		Denom:  pioconfig.GetProvenanceConfig().MsgFloorDenom,
	}
}

var DefaultNhashPerUsdMil = uint64(25_000_000)

var (
	// ParamStoreKeyFloorGasPrice if msg fees are paid in the same denom as base default gas is paid, then use this to differentiate between base price
	// and additional fees.
	ParamStoreKeyFloorGasPrice      = []byte("FloorGasPrice")
	ParamStoreKeyNhashPerUsdMil     = []byte("NhashPerUsdMil")
	ParamStoreKeyConversionFeeDenom = []byte("ConversionFeeDenom")
	ParamStoreKeyDefaultMsgFee      = []byte("DefaultMsgFee")
)

// ParamKeyTable for marker module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new parameter object
func NewParams(
	floorGasPrice sdk.Coin,
	nhashPerUsdMil uint64,
	conversionFeeDenom string,
	defaultMsgFee sdk.Coin,
) Params {
	return Params{
		FloorGasPrice:      floorGasPrice,
		NhashPerUsdMil:     nhashPerUsdMil,
		ConversionFeeDenom: conversionFeeDenom,
		DefaultMsgFee:      defaultMsgFee,
	}
}

// ParamSetPairs - Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyFloorGasPrice, &p.FloorGasPrice, validateCoinParam),
		paramtypes.NewParamSetPair(ParamStoreKeyNhashPerUsdMil, &p.NhashPerUsdMil, validateNhashPerUsdMilParam),
		paramtypes.NewParamSetPair(ParamStoreKeyConversionFeeDenom, &p.ConversionFeeDenom, validateConversionFeeDenomParam),
		paramtypes.NewParamSetPair(ParamStoreKeyDefaultMsgFee, &p.DefaultMsgFee, validateCoinParam),
	}
}

// DefaultParams is the default parameter configuration for the bank module
func DefaultParams() Params {
	return NewParams(
		DefaultFloorGasPrice(),
		DefaultNhashPerUsdMil,
		pioconfig.GetProvenanceConfig().FeeDenom,
		DefaultMsgFee(),
	)
}

// Equal returns true if the given value is equivalent to the current instance of params
func (p *Params) Equal(that interface{}) bool {
	if that == nil {
		return p == nil
	}

	that1, ok := that.(*Params)
	if !ok {
		that2, ok := that.(Params)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return p == nil
	} else if p == nil {
		return false
	}
	return true
}

func validateCoinParam(i interface{}) error {
	coin, ok := i.(sdk.Coin)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// validate appropriate Coin
	if coin.Validate() != nil {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateNhashPerUsdMilParam(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateConversionFeeDenomParam(i interface{}) error {
	_, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}
