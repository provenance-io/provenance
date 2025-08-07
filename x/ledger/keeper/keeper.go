package keeper

import (
	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/provenance-io/provenance/x/ledger/types"
)

// Keeper defines the ledger module keeper that manages the state store configuration
// and provides access to all ledger-related data collections. The keeper serves as
// the central interface for interacting with the ledger module's persistent state.
//
// The keeper organizes data into several collections:
// - Core ledger data (ledgers, entries, settlements)
// - Ledger class configuration (classes, entry types, status types, bucket types)
// - Integration with other modules (bank, registry)
type Keeper struct {
	// Core infrastructure for state management
	cdc      codec.BinaryCodec   // Binary codec for serialization/deserialization
	storeKey storetypes.StoreKey // Store key for the ledger module's state store
	schema   collections.Schema  // Schema defining the structure of all collections

	// Ledgers stores the main ledger instances indexed by ledger ID (bech32 string).
	// Each ledger represents a complete accounting system for a specific asset.
	Ledgers collections.Map[string, types.Ledger]

	// LedgerEntries stores individual entries within ledgers.
	// It is keyed by (ledger_id, entry_id) pair for efficient lookup.
	LedgerEntries collections.Map[collections.Pair[string, string], types.LedgerEntry]

	// FundTransfersWithSettlement stores settlement instructions for fund transfers.
	// It is keyed by (ledger_id, settlement_id) pair.
	FundTransfersWithSettlement collections.Map[collections.Pair[string, string], types.StoredSettlementInstructions]

	// LedgerClasses stores the configuration of all ledgers for a given class of asset.
	// Each ledger class defines the entry types, status types, and bucket types available
	// for ledgers of that class.
	LedgerClasses collections.Map[string, types.LedgerClass]

	// LedgerClassEntryTypes stores the entry type definitions for each ledger class.
	// It is keyed by (class_id, entry_type_id) pair.
	// Entry types define what kinds of transactions can be recorded in ledgers of this class.
	LedgerClassEntryTypes collections.Map[collections.Pair[string, int32], types.LedgerClassEntryType]

	// LedgerClassStatusTypes stores the status type definitions for each ledger class.
	// It is keyed by (class_id, status_type_id) pair.
	// Status types define the possible states that ledger entries can have.
	LedgerClassStatusTypes collections.Map[collections.Pair[string, int32], types.LedgerClassStatusType]

	// LedgerClassBucketTypes stores the bucket type definitions for each ledger class.
	// It is keyed by (class_id, bucket_type_id) pair.
	// Bucket types define how funds are categorized and organized within ledgers.
	LedgerClassBucketTypes collections.Map[collections.Pair[string, int32], types.LedgerClassBucketType]

	BankKeeper     // Provides access to bank module for token transfers and balances
	RegistryKeeper // Provides access to registry module for NFT ownership and role management
}

// NewKeeper creates and configures a new ledger keeper instance.
// This function sets up all the collections that define the ledger module's state store structure.
//
// Parameters:
// - cdc: Binary codec for data serialization
// - storeKey: Store key for the ledger module's state store
// - storeService: KV store service for state persistence
// - bankKeeper: Keeper for bank module integration
// - registryKeeper: Keeper for registry module integration
//
// Returns a fully configured Keeper instance with all collections initialized.
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, storeService store.KVStoreService, bankKeeper BankKeeper, registryKeeper RegistryKeeper) Keeper {
	// Create a schema builder to define the structure of all collections
	sb := collections.NewSchemaBuilder(storeService)

	lk := Keeper{
		cdc:      cdc,
		storeKey: storeKey,

		// Initialize core ledger data collections
		// These collections store the primary ledger functionality

		// Ledgers collection stores complete ledger instances.
		// The key is ledger_id (string) and the value is Ledger (complete ledger configuration and metadata).
		Ledgers: collections.NewMap(
			sb,
			collections.NewPrefix(ledgerPrefix),
			"ledger",
			collections.StringKey,
			codec.CollValue[types.Ledger](cdc),
		),

		// LedgerEntries collection stores individual ledger entries.
		// The key is (ledger_id, entry_id) pair and the value is LedgerEntry (individual transaction or accounting entry).
		LedgerEntries: collections.NewMap(
			sb,
			collections.NewPrefix(entriesPrefix),
			"ledger_entries",
			collections.PairKeyCodec(collections.StringKey, collections.StringKey),
			codec.CollValue[types.LedgerEntry](cdc),
		),

		// FundTransfersWithSettlement collection stores settlement instructions.
		// The key is (ledger_id, settlement_id) pair and the value is StoredSettlementInstructions (settlement configuration and metadata).
		FundTransfersWithSettlement: collections.NewMap(
			sb,
			collections.NewPrefix(fundTransfersWithSettlementPrefix),
			"settlement_instructions",
			collections.PairKeyCodec(collections.StringKey, collections.StringKey),
			codec.CollValue[types.StoredSettlementInstructions](cdc),
		),

		// Initialize ledger class configuration collections
		// These collections define the structure and behavior of different ledger types

		// LedgerClasses collection stores ledger class configurations.
		// The key is class_id (string) and the value is LedgerClass (complete class configuration including entry types, status types, bucket types).
		LedgerClasses: collections.NewMap(
			sb,
			collections.NewPrefix(ledgerClassesPrefix),
			"ledger_classes",
			collections.StringKey,
			codec.CollValue[types.LedgerClass](cdc),
		),

		// LedgerClassEntryTypes collection stores entry type definitions per class.
		// The key is (class_id, entry_type_id) pair and the value is LedgerClassEntryType (entry type configuration and validation rules).
		LedgerClassEntryTypes: collections.NewMap(
			sb,
			collections.NewPrefix(ledgerClassEntryTypesPrefix),
			"ledger_class_entry_types",
			collections.PairKeyCodec(collections.StringKey, collections.Int32Key),
			codec.CollValue[types.LedgerClassEntryType](cdc),
		),

		// LedgerClassStatusTypes collection stores status type definitions per class.
		// The key is (class_id, status_type_id) pair and the value is LedgerClassStatusType (status type configuration and state transitions).
		LedgerClassStatusTypes: collections.NewMap(
			sb,
			collections.NewPrefix(ledgerClassStatusTypesPrefix),
			"ledger_class_status_types",
			collections.PairKeyCodec(collections.StringKey, collections.Int32Key),
			codec.CollValue[types.LedgerClassStatusType](cdc),
		),

		// LedgerClassBucketTypes collection stores bucket type definitions per class.
		// The key is (class_id, bucket_type_id) pair and the value is LedgerClassBucketType (bucket type configuration and categorization rules).
		LedgerClassBucketTypes: collections.NewMap(
			sb,
			collections.NewPrefix(ledgerClassBucketTypesPrefix),
			"ledger_class_bucket_types",
			collections.PairKeyCodec(collections.StringKey, collections.Int32Key),
			codec.CollValue[types.LedgerClassBucketType](cdc),
		),

		// Set module integration dependencies
		BankKeeper:     bankKeeper,
		RegistryKeeper: registryKeeper,
	}

	// Build and set the schema
	// The schema defines the complete structure of all collections and their relationships.
	// This ensures data consistency and provides a contract for the state store layout.
	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	lk.schema = schema

	return lk
}
