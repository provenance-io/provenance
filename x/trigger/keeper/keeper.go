package keeper

import (
	"cosmossdk.io/collections"
	store "cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/trigger/types"
)

type Keeper struct {
	cdc          codec.BinaryCodec
	router       baseapp.IMsgServiceRouter
	StoreService store.KVStoreService
	// Collections Schema
	Schema         collections.Schema
	TriggersMap    collections.Map[uint64, types.Trigger]                         // prefix 0x01
	EventListeners collections.KeySet[collections.Triple[[]byte, uint64, uint64]] // prefix 0x02
	Queue          collections.Map[uint64, types.QueuedTrigger]                   // prefix 0x03
	NextTriggerID  collections.Item[uint64]                                       // prefix 0x05
	// Queue Metadata items for compatibility
	QueueStartIndex collections.Item[uint64] // prefix 0x06
	QueueLength     collections.Item[uint64] // prefix 0x07
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	router baseapp.IMsgServiceRouter,

) Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		cdc:    cdc,
		router: router,
		// Primary trigger storage
		TriggersMap: collections.NewMap(sb, types.TriggerKeyPrefix, "triggers", collections.Uint64Key, codec.CollValue[types.Trigger](cdc)),
		// Event listeners with custom key codec for compatibility
		EventListeners: collections.NewKeySet(sb, types.EventListenerKeyPrefix, "event_listeners", types.EventListenerKeyCodec()),
		// Queue items
		Queue: collections.NewMap(sb, types.QueueKeyPrefix, "queue", collections.Uint64Key, codec.CollValue[types.QueuedTrigger](cdc)),
		// Auto-incrementing trigger IDs
		NextTriggerID: collections.NewItem(sb, types.NextTriggerIDKey, "trigger_id_sequence", collections.Uint64Value),
		// Queue metadata - SEPARATE for backward compatibility
		QueueStartIndex: collections.NewItem(sb, types.QueueStartIndexKey, "queue_start_index", collections.Uint64Value),
		QueueLength:     collections.NewItem(sb, types.QueueLengthKey, "queue_length", collections.Uint64Value),
	}
	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}
