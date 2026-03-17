package keeper

import (
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/provenance-io/provenance/x/ibcratelimit"
)

// Keeper for the ibcratelimit module
type Keeper struct {
	storeService       store.KVStoreService
	cdc                codec.BinaryCodec
	PermissionedKeeper ibcratelimit.PermissionedKeeper
	authority          string
	// Schema is the collections schema for this module's store.
	Schema collections.Schema

	// Params is the single-item collection storing the module parameters.
	ParamsItem collections.Item[ibcratelimit.Params]
}

// NewKeeper Creates a new Keeper for the module.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	permissionedKeeper ibcratelimit.PermissionedKeeper,
) Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		cdc:                cdc,
		storeService:       storeService,
		PermissionedKeeper: permissionedKeeper,
		authority:          authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		ParamsItem: collections.NewItem(
			sb,
			ibcratelimit.ParamsKeyPrefix,
			"params",
			codec.CollValue[ibcratelimit.Params](cdc),
		),
	}

	var err error
	k.Schema, err = sb.Build()
	if err != nil {
		panic(fmt.Errorf("failed to build ibcratelimit collections schema: %w", err))
	}

	return k
}

// Logger Creates a new logger for the module.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+ibcratelimit.ModuleName)
}

// GetParams Gets the params for the module.
func (k Keeper) GetParams(ctx sdk.Context) (params ibcratelimit.Params, err error) {
	params, err = k.ParamsItem.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return ibcratelimit.DefaultParams(), nil
		}
		return ibcratelimit.Params{}, err
	}
	return params, nil
}

// SetParams Sets the params for the module.
func (k Keeper) SetParams(ctx sdk.Context, params ibcratelimit.Params) error {
	return k.ParamsItem.Set(ctx, params)
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
