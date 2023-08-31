package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	// TODO[1658]: Finish the Keeper struct.
	authority string
}

func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey) Keeper {
	// TODO[1658]: Finish NewKeeper.
	rv := Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	}
	return rv
}

// GetAuthority gets the address (as bech32) that has governance authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}
