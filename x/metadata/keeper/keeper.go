package keeper

import (
	"net/url"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/gogoproto/proto"

	"github.com/provenance-io/provenance/x/metadata/types"
)

// Keeper is the concrete state-based API for the metadata module.
type Keeper struct {
	// Key to access the key-value store from sdk.Context
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec

	// To check if accounts exist and set public keys.
	authKeeper AuthKeeper

	// To check granter grantee authorization of messages.
	authzKeeper AuthzKeeper

	// For getting/setting account data.
	attrKeeper AttrKeeper

	// For getting marker accounts
	markerKeeper MarkerKeeper

	// For managing value owners
	bankKeeper BankKeeper
}

// NewKeeper creates new instances of the metadata Keeper.
func NewKeeper(
	cdc codec.BinaryCodec, key storetypes.StoreKey, authKeeper AuthKeeper,
	authzKeeper AuthzKeeper, attrKeeper AttrKeeper, markerKeeper MarkerKeeper,
	bankKeeper bankkeeper.BaseKeeper,
) Keeper {
	return Keeper{
		storeKey:     key,
		cdc:          cdc,
		authKeeper:   authKeeper,
		authzKeeper:  authzKeeper,
		attrKeeper:   attrKeeper,
		markerKeeper: markerKeeper,
		bankKeeper:   NewMDBankKeeper(bankKeeper),
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// VerifyCorrectOwner to determines whether the signer resolves to the owner of the OSLocator record.
func (k Keeper) VerifyCorrectOwner(ctx sdk.Context, ownerAddr sdk.AccAddress) bool {
	stored, found := k.GetOsLocatorRecord(ctx, ownerAddr)
	if !found {
		return false
	}
	return ownerAddr.String() == stored.Owner
}

func (k Keeper) EmitEvent(ctx sdk.Context, event proto.Message) {
	err := ctx.EventManager().EmitTypedEvent(event)
	if err != nil {
		ctx.Logger().Error("unable to emit event", "error", err, "event", event)
	}
}

// unionUnique gets a union of the provided sets of strings without any duplicates.
func (k Keeper) UnionDistinct(sets ...[]string) []string {
	retval := []string{}
	for _, s := range sets {
		for _, v := range s {
			f := false
			for _, r := range retval {
				if r == v {
					f = true
					break
				}
			}
			if !f {
				retval = append(retval, v)
			}
		}
	}
	return retval
}

func (k Keeper) checkValidURI(uri string, ctx sdk.Context) (*url.URL, error) {
	urlToPersist, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	if urlToPersist.Scheme == "" || urlToPersist.Host == "" {
		return nil, types.ErrOSLocatorURIInvalid
	}

	if int(k.GetOSLocatorParams(ctx).MaxUriLength) < len(uri) {
		return nil, types.ErrOSLocatorURIToolong
	}
	return urlToPersist, nil
}
