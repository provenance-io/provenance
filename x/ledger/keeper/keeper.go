package keeper

import (
	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
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
	cdc    codec.BinaryCodec  // Binary codec for serialization/deserialization
	schema collections.Schema // Schema defining the structure of all collections

	// Ledgers stores the main ledger instances indexed by ledger ID (bech32 string).
	// Each ledger represents a complete accounting system for a specific asset.
	// The key is ledger_id (string).
	Ledgers collections.Map[string, types.Ledger]

	// LedgerEntries stores individual entries within ledgers.
	// The key is (ledger_id, entry_id) pair.
	LedgerEntries collections.Map[collections.Pair[string, string], types.LedgerEntry]

	// FundTransfersWithSettlement stores settlement instructions for fund transfers.
	// The key is (ledger_id, settlement_id) pair.
	FundTransfersWithSettlement collections.Map[collections.Pair[string, string], types.StoredSettlementInstructions]

	// LedgerClasses stores the configuration of all ledgers for a given class of asset.
	// Each ledger class defines the entry types, status types, and bucket types available
	// for ledgers of that class.
	// The key is class_id (string).
	LedgerClasses collections.Map[string, types.LedgerClass]

	// LedgerClassEntryTypes stores the entry type definitions for each ledger class.
	// Entry types define what kinds of transactions can be recorded in ledgers of this class.
	// The key is (class_id, entry_type_id) pair.
	LedgerClassEntryTypes collections.Map[collections.Pair[string, int32], types.LedgerClassEntryType]

	// LedgerClassStatusTypes stores the status type definitions for each ledger class.
	// Status types define the possible states that ledger entries can have.
	// The key is (class_id, status_type_id) pair.
	LedgerClassStatusTypes collections.Map[collections.Pair[string, int32], types.LedgerClassStatusType]

	// LedgerClassBucketTypes stores the bucket type definitions for each ledger class.
	// Bucket types define how funds are categorized and organized within ledgers.
	// The key is (class_id, bucket_type_id) pair.
	LedgerClassBucketTypes collections.Map[collections.Pair[string, int32], types.LedgerClassBucketType]

	BankKeeper     BankKeeper     // Provides access to bank module for token transfers and balances
	RegistryKeeper RegistryKeeper // Provides access to registry module for NFT ownership and role management
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
func NewKeeper(cdc codec.BinaryCodec, storeService store.KVStoreService, bankKeeper BankKeeper, registryKeeper RegistryKeeper) Keeper {
	// Create a schema builder to define the structure of all collections
	sb := collections.NewSchemaBuilder(storeService)

	lk := Keeper{
		cdc: cdc,

		// Initialize core ledger data collections
		// These collections store the primary ledger functionality

		Ledgers: collections.NewMap(sb,
			collections.NewPrefix(ledgerPrefix), "ledger",
			collections.StringKey,
			codec.CollValue[types.Ledger](cdc)),

		LedgerEntries: collections.NewMap(sb,
			collections.NewPrefix(entriesPrefix), "ledger_entries",
			collections.PairKeyCodec(collections.StringKey, collections.StringKey),
			codec.CollValue[types.LedgerEntry](cdc)),

		FundTransfersWithSettlement: collections.NewMap(sb,
			collections.NewPrefix(fundTransfersWithSettlementPrefix), "settlement_instructions",
			collections.PairKeyCodec(collections.StringKey, collections.StringKey),
			codec.CollValue[types.StoredSettlementInstructions](cdc)),

		// Initialize ledger class configuration collections
		// These collections define the structure and behavior of different ledger types

		LedgerClasses: collections.NewMap(sb,
			collections.NewPrefix(ledgerClassesPrefix), "ledger_classes",
			collections.StringKey,
			codec.CollValue[types.LedgerClass](cdc)),

		LedgerClassEntryTypes: collections.NewMap(sb,
			collections.NewPrefix(ledgerClassEntryTypesPrefix), "ledger_class_entry_types",
			collections.PairKeyCodec(collections.StringKey, collections.Int32Key),
			codec.CollValue[types.LedgerClassEntryType](cdc)),

		LedgerClassStatusTypes: collections.NewMap(sb,
			collections.NewPrefix(ledgerClassStatusTypesPrefix), "ledger_class_status_types",
			collections.PairKeyCodec(collections.StringKey, collections.Int32Key),
			codec.CollValue[types.LedgerClassStatusType](cdc)),

		LedgerClassBucketTypes: collections.NewMap(sb,
			collections.NewPrefix(ledgerClassBucketTypesPrefix), "ledger_class_bucket_types",
			collections.PairKeyCodec(collections.StringKey, collections.Int32Key),
			codec.CollValue[types.LedgerClassBucketType](cdc)),

		// Set module integration dependencies
		BankKeeper:     bankKeeper,
		RegistryKeeper: registryKeeper,
	}

	// Build and set the schema.
	// The schema defines the complete structure of all collections and their relationships.
	// This ensures data consistency and provides a contract for the state store layout.
	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	lk.schema = schema

	return lk
}
