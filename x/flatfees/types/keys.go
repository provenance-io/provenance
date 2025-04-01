package types

import "cosmossdk.io/collections"

const (
	// ModuleName defines the module name
	ModuleName = "flatfees"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName
)

var (
	// MsgFeeKeyBz prefix bytes for the MsgFee entries.
	MsgFeeKeyBz = []byte{0x00}
	// ParamsKeyBz prefix bytes for the x/flatfees params entry.
	ParamsKeyBz = []byte{0x01}

	// MsgFeeKeyPrefix is the collections prefix for the MsgFee entries.
	MsgFeeKeyPrefix = collections.NewPrefix(MsgFeeKeyBz)
	// ParamsKeyPrefix is the collections prefix for the Params entry.
	ParamsKeyPrefix = collections.NewPrefix(ParamsKeyBz)
)
