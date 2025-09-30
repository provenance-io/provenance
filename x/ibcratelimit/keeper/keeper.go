package keeper

import (
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/provenance-io/provenance/x/ibcratelimit"
)

// Keeper for the ibcratelimit module
type Keeper struct {
	storeKey           storetypes.StoreKey
	cdc                codec.BinaryCodec
	PermissionedKeeper ibcratelimit.PermissionedKeeper
	authority          string
}

// NewKeeper Creates a new Keeper for the module.
func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	permissionedKeeper ibcratelimit.PermissionedKeeper,
) Keeper {
	return Keeper{
		storeKey:           key,
		cdc:                cdc,
		PermissionedKeeper: permissionedKeeper,
		authority:          authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	}
}

// Logger Creates a new logger for the module.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+ibcratelimit.ModuleName)
}

// GetParams Gets the params for the module.
func (k Keeper) GetParams(ctx sdk.Context) (params ibcratelimit.Params, err error) {
	store := ctx.KVStore(k.storeKey)
	key := ibcratelimit.ParamsKey
	bz := store.Get(key)
	if len(bz) == 0 {
		return ibcratelimit.Params{}, nil
	}
	err = k.cdc.Unmarshal(bz, &params)
	return params, err
}

// SetParams Sets the params for the module.
func (k Keeper) SetParams(ctx sdk.Context, params ibcratelimit.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&params)
	store.Set(ibcratelimit.ParamsKey, bz)
}

// GetContractAddress Gets the current value of the module's contract address.
func (k Keeper) GetContractAddress(ctx sdk.Context) (contract string) {
	params, _ := k.GetParams(ctx)
	return params.ContractAddress
}

// IsContractConfigured Checks if the contract has been configured for the module.
func (k Keeper) IsContractConfigured(ctx sdk.Context) bool {
	params, err := k.GetParams(ctx)
	if err != nil {
		return false
	}
	return params.ContractAddress != ""
}

// GetAuthority gets the authority account address.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// ValidateAuthority returns an error if the provided address is not the authority.
func (k Keeper) ValidateAuthority(addr string) error {
	if k.authority != addr {
		return govtypes.ErrInvalidSigner.Wrapf("expected %q got %q", k.authority, addr)
	}
	return nil
}

// emitEvent emits the provided event and writes any error to the error log.
func (k Keeper) emitEvent(ctx sdk.Context, event proto.Message) {
	err := ctx.EventManager().EmitTypedEvent(event)
	if err != nil {
		k.Logger(ctx).Error("error emitting event %#v: %v", event, err)
	}
}
