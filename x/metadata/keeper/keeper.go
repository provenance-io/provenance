package keeper

import (
	"net/url"
	"slices"

	"cosmossdk.io/collections"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/gogoproto/proto"

	"github.com/provenance-io/provenance/x/metadata/types"
)

// indexPresent is the sentinel value stored in all secondary index collections,
var indexPresent = []byte{0x01}

// Keeper is the concrete state-based API for the metadata module.
type Keeper struct {
	// Key to access the key-value store from sdk.Context
	storeService corestore.KVStoreService
	cdc          codec.BinaryCodec
	moduleAddr   sdk.AccAddress
	schema       collections.Schema

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
	// Primary data collections.
	// Key = MetadataAddress[1:] to preserve legacy storage layout
	scopeCollections       collections.Map[[]byte, types.Scope]
	sessionCollections     collections.Map[[]byte, types.Session]
	recordCollections      collections.Map[[]byte, types.Record]
	contractSpecCollection collections.Map[[]byte, types.ContractSpecification]
	scopeSpecs             collections.Map[[]byte, types.ScopeSpecification]
	recordSpecs            collections.Map[[]byte, types.RecordSpecification]
	// OS Locator collection.
	// Key = address.MustLengthPrefix(ownerAddr) — matches existing GetOSLocatorKey layout.
	osLocators collections.Map[[]byte, types.ObjectStoreLocator]
	// OS Locator params — single item stored at the prefix bytes.
	osLocatorParamsCollection collections.Item[types.OSLocatorParams]
	// Net Asset Values.
	// Key = address.MustLengthPrefix(scopeAddr) + []byte(denom) — matches NetAssetValueKey layout.
	netAssetValues collections.Map[[]byte, types.NetAssetValue]

	// Secondary index caches (value = []byte{0x01} matching existing layout).
	// prefix 0x17: key = length_prefix(addr) + scopeID_17bytes
	addressScopeIndex collections.Map[[]byte, []byte]
	// prefix 0x11: key = scopeSpecID_17bytes + scopeID_17bytes
	scopeSpecScopeIndex collections.Map[[]byte, []byte]
	// prefix 0x19: key = length_prefix(addr) + scopeSpecID_17bytes
	addressScopeSpecIndex collections.Map[[]byte, []byte]
	// prefix 0x14: key = contractSpecID_17bytes + scopeSpecID_17bytes
	contractSpecScopeSpecIndex collections.Map[[]byte, []byte]
	// prefix 0x20: key = length_prefix(addr) + contractSpecID_17bytes
	addressContractSpecIndex collections.Map[[]byte, []byte]
}

// NewKeeper creates new instances of the metadata Keeper.
func NewKeeper(
	cdc codec.BinaryCodec, key corestore.KVStoreService, authKeeper AuthKeeper,
	authzKeeper AuthzKeeper, attrKeeper AttrKeeper, markerKeeper MarkerKeeper,
	bankKeeper bankkeeper.BaseKeeper,
) Keeper {
	sb := collections.NewSchemaBuilder(key)

	k := Keeper{
		storeService: key,
		cdc:          cdc,
		moduleAddr:   authtypes.NewModuleAddress(types.ModuleName),
		authKeeper:   authKeeper,
		authzKeeper:  authzKeeper,
		attrKeeper:   attrKeeper,
		markerKeeper: markerKeeper,
		bankKeeper:   NewMDBankKeeper(bankKeeper),

		scopeCollections: collections.NewMap(sb,
			collections.NewPrefix(types.ScopeKeyPrefix),
			"scopes",
			collections.BytesKey,
			codec.CollValue[types.Scope](cdc)),

		sessionCollections: collections.NewMap(sb,
			collections.NewPrefix(types.SessionKeyPrefix),
			"sessions",
			collections.BytesKey,
			codec.CollValue[types.Session](cdc)),

		recordCollections: collections.NewMap(sb,
			collections.NewPrefix(types.RecordKeyPrefix),
			"records",
			collections.BytesKey,
			codec.CollValue[types.Record](cdc)),

		contractSpecCollection: collections.NewMap(sb,
			collections.NewPrefix(types.ContractSpecificationKeyPrefix),
			"contract_specs",
			collections.BytesKey,
			codec.CollValue[types.ContractSpecification](cdc)),

		scopeSpecs: collections.NewMap(sb,
			collections.NewPrefix(types.ScopeSpecificationKeyPrefix),
			"scope_specs",
			collections.BytesKey,
			codec.CollValue[types.ScopeSpecification](cdc)),

		recordSpecs: collections.NewMap(sb,
			collections.NewPrefix(types.RecordSpecificationKeyPrefix),
			"record_specs",
			collections.BytesKey,
			codec.CollValue[types.RecordSpecification](cdc)),

		osLocators: collections.NewMap(sb,
			collections.NewPrefix(types.OSLocatorAddressKeyPrefix),
			"os_locators",
			collections.BytesKey,
			codec.CollValue[types.ObjectStoreLocator](cdc)),

		osLocatorParamsCollection: collections.NewItem(sb,
			collections.NewPrefix(types.OSLocatorParamPrefix),
			"os_locator_params",
			codec.CollValue[types.OSLocatorParams](cdc)),

		netAssetValues: collections.NewMap(sb,
			collections.NewPrefix(types.NetAssetValuePrefix),
			"net_asset_values",
			collections.BytesKey,
			codec.CollValue[types.NetAssetValue](cdc)),

		addressScopeIndex: collections.NewMap(sb,
			collections.NewPrefix(types.AddressScopeCacheKeyPrefix),
			"addr_scope_idx",
			collections.BytesKey,
			collections.BytesValue),

		scopeSpecScopeIndex: collections.NewMap(sb,
			collections.NewPrefix(types.ScopeSpecScopeCacheKeyPrefix),
			"scope_spec_scope_idx",
			collections.BytesKey,
			collections.BytesValue),

		addressScopeSpecIndex: collections.NewMap(sb,
			collections.NewPrefix(types.AddressScopeSpecCacheKeyPrefix),
			"addr_scope_spec_idx",
			collections.BytesKey,
			collections.BytesValue),

		contractSpecScopeSpecIndex: collections.NewMap(sb,
			collections.NewPrefix(types.ContractSpecScopeSpecCacheKeyPrefix),
			"contract_spec_scope_spec_idx",
			collections.BytesKey,
			collections.BytesValue),

		addressContractSpecIndex: collections.NewMap(sb,
			collections.NewPrefix(types.AddressContractSpecCacheKeyPrefix),
			"addr_contract_spec_idx",
			collections.BytesKey,
			collections.BytesValue),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.schema = schema
	return k
}

// mdKey returns the collection key for a MetadataAddress.
func mdKey(addr types.MetadataAddress) []byte {
	if len(addr) == 0 {
		return nil
	}
	return addr[1:]
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
			if !slices.Contains(retval, v) {
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
