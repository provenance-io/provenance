package keeper

import "github.com/cosmos/cosmos-sdk/codec"

// This file is in the keeper package (not keeper_test) so that it can expose
// some private keeper stuff for unit testing.

// WithCodec returns a new Keeper that uses the provided codec. Only for unit tests.
func (k Keeper) WithCodec(cdc codec.Codec) Keeper {
	k.cdc = cdc
	return k
}

// GetCodec returns this Keeper's codec. Only for unit tests.
func (k Keeper) GetCodec() codec.Codec {
	return k.cdc
}
