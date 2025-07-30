package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ledger "github.com/provenance-io/provenance/x/ledger/types"
	"github.com/provenance-io/provenance/x/registry"
)

// Keeper defines the mymodule keeper.
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	schema   collections.Schema

	Ledgers                     collections.Map[string, ledger.Ledger]
	LedgerEntries               collections.Map[collections.Pair[string, string], ledger.LedgerEntry]
	FundTransfersWithSettlement collections.Map[collections.Pair[string, string], ledger.StoredSettlementInstructions]

	// LedgerClasses stores the configuration of all ledgers for a given class of asset.
	LedgerClasses          collections.Map[string, ledger.LedgerClass]
	LedgerClassEntryTypes  collections.Map[collections.Pair[string, int32], ledger.LedgerClassEntryType]
	LedgerClassStatusTypes collections.Map[collections.Pair[string, int32], ledger.LedgerClassStatusType]
	LedgerClassBucketTypes collections.Map[collections.Pair[string, int32], ledger.LedgerClassBucketType]

	BankKeeper
	RegistryKeeper
}

func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, storeService store.KVStoreService, bankKeeper BankKeeper, registryKeeper RegistryKeeper) Keeper {
	sb := collections.NewSchemaBuilder(storeService)

	lk := Keeper{
		cdc:      cdc,
		storeKey: storeKey,

		Ledgers: collections.NewMap(
			sb,
			collections.NewPrefix(ledgerPrefix),
			"ledger",
			collections.StringKey,
			codec.CollValue[ledger.Ledger](cdc),
		),
		LedgerEntries: collections.NewMap(
			sb,
			collections.NewPrefix(entriesPrefix),
			"ledger_entries",
			collections.PairKeyCodec(collections.StringKey, collections.StringKey),
			codec.CollValue[ledger.LedgerEntry](cdc),
		),
		FundTransfersWithSettlement: collections.NewMap(
			sb,
			collections.NewPrefix(fundTransfersWithSettlementPrefix),
			"settlement_instructions",
			collections.PairKeyCodec(collections.StringKey, collections.StringKey),
			codec.CollValue[ledger.StoredSettlementInstructions](cdc),
		),

		// Ledger Class configuration data
		LedgerClasses: collections.NewMap(
			sb,
			collections.NewPrefix(ledgerClassesPrefix),
			"ledger_classes",
			collections.StringKey,
			codec.CollValue[ledger.LedgerClass](cdc),
		),
		LedgerClassEntryTypes: collections.NewMap(
			sb,
			collections.NewPrefix(ledgerClassEntryTypesPrefix),
			"ledger_class_entry_types",
			collections.PairKeyCodec(collections.StringKey, collections.Int32Key),
			codec.CollValue[ledger.LedgerClassEntryType](cdc),
		),
		LedgerClassStatusTypes: collections.NewMap(
			sb,
			collections.NewPrefix(ledgerClassStatusTypesPrefix),
			"ledger_class_status_types",
			collections.PairKeyCodec(collections.StringKey, collections.Int32Key),
			codec.CollValue[ledger.LedgerClassStatusType](cdc),
		),
		LedgerClassBucketTypes: collections.NewMap(
			sb,
			collections.NewPrefix(ledgerClassBucketTypesPrefix),
			"ledger_class_bucket_types",
			collections.PairKeyCodec(collections.StringKey, collections.Int32Key),
			codec.CollValue[ledger.LedgerClassBucketType](cdc),
		),

		BankKeeper:     bankKeeper,
		RegistryKeeper: registryKeeper,
	}
	// Build and set the schema
	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	lk.schema = schema

	return lk
}

func NewLedgerKey(assetClassId string, nftId string) *ledger.LedgerKey {
	return &ledger.LedgerKey{
		AssetClassId: assetClassId,
		NftId:        nftId,
	}
}

func (k Keeper) RequireAuthority(ctx sdk.Context, addr string, key *registry.RegistryKey) error {
	has, err := assertAuthority(ctx, k.RegistryKeeper, addr, key)
	if err != nil {
		return err
	}
	if !has {
		return ledger.NewLedgerCodedError(ledger.ErrCodeUnauthorized, "authority is not the owner or servicer")
	}
	return nil
}

func assertOwner(ctx sdk.Context, k RegistryKeeper, authorityAddr string, ledgerKey *ledger.LedgerKey) error {
	// Check if the authority has ownership of the NFT
	nftOwner := k.GetNFTOwner(ctx, &ledgerKey.AssetClassId, &ledgerKey.NftId)
	if nftOwner == nil || nftOwner.String() != authorityAddr {
		return ledger.NewLedgerCodedError(ledger.ErrCodeUnauthorized, fmt.Sprintf("authority is not the nft owner (%s)", nftOwner.String()))
	}

	return nil
}

// Assert that the authority address is either the registered servicer, or the owner of the NFT if there is no registered servicer.
func assertAuthority(ctx sdk.Context, k RegistryKeeper, authorityAddr string, rk *registry.RegistryKey) (bool, error) {
	// Get the registry entry for the NFT to determine if the authority has the servicer role.
	registryEntry, err := k.GetRegistry(ctx, rk)
	if err != nil {
		return false, err
	}

	lk := NewLedgerKey(rk.AssetClassId, rk.NftId)

	// If there is no registry entry, the authority is the owner.
	if registryEntry == nil {
		err = assertOwner(ctx, k, authorityAddr, lk)
		if err != nil {
			return false, err
		}

		return true, nil
	}

	// Since the authority doesn't have the servicer role, let's see if there is any servicer set. If there is, we'll return an error
	// so that only the assigned servicer can append entries.
	for _, role := range registryEntry.Roles {
		if role.Role == registry.RegistryRole_REGISTRY_ROLE_SERVICER {
			for _, address := range role.Addresses {
				// Check if the authority is the servicer
				if address == authorityAddr {
					return true, nil
				}
			}

			// Since there is a registered servicer, the owner is not authorized.
			return false, ledger.NewLedgerCodedError(ledger.ErrCodeUnauthorized, "registered servicer")
		}
	}

	// Since there isn't a registered servicer, let's see if the authority is the owner.
	err = assertOwner(ctx, k, authorityAddr, lk)
	if err == nil {
		// The authority owns the asset, and there is no registered servicer
		return true, nil

	}

	return false, err
}

// TODO move this to init genesis.
// BulkImportLedgerData imports ledger data from genesis state. This function assumes that ledger classes, status
// types, entry types, and bucket types are already created before calling this function.
func (k Keeper) BulkImportLedgerData(ctx sdk.Context, authorityAddr sdk.AccAddress, genesisState ledger.GenesisState) error {
	ctx.Logger().Info("Starting bulk import of ledger data",
		"ledger_to_entries", len(genesisState.LedgerToEntries))

	// Import ledgers and their entries
	for _, ledgerToEntries := range genesisState.LedgerToEntries {
		var maintainerAddr sdk.AccAddress
		var ledgerClassId string

		// Determine the ledger class ID and get maintainer address
		if ledgerToEntries.Ledger != nil {
			// If we have a ledger object, use its ledger class ID
			ledgerClassId = ledgerToEntries.Ledger.LedgerClassId
		} else {
			// If we don't have a ledger object, get it from the existing ledger
			existingLedger, err := k.GetLedger(ctx, ledgerToEntries.LedgerKey)
			if err != nil {
				return fmt.Errorf("failed to get existing ledger: %w", err)
			}
			if existingLedger == nil {
				return fmt.Errorf("ledger %s does not exist and no ledger object provided", ledgerToEntries.LedgerKey.NftId)
			}
			ledgerClassId = existingLedger.LedgerClassId
		}

		// Get the maintainer address from the ledger class
		ledgerClass, err := k.GetLedgerClass(ctx, ledgerClassId)
		if err != nil {
			return fmt.Errorf("failed to get ledger class %s: %w", ledgerClassId, err)
		}
		if ledgerClass == nil {
			return fmt.Errorf("ledger class %s not found - ensure it is created before bulk import", ledgerClassId)
		}

		maintainerAddr, err = sdk.AccAddressFromBech32(ledgerClass.MaintainerAddress)
		if err != nil {
			return fmt.Errorf("invalid maintainer address %s: %w", ledgerClass.MaintainerAddress, err)
		}

		// Create the ledger only if it doesn't already exist
		if ledgerToEntries.Ledger != nil && !k.HasLedger(ctx, ledgerToEntries.Ledger.Key) {
			if err := k.AddLedger(ctx, maintainerAddr, *ledgerToEntries.Ledger); err != nil {
				return fmt.Errorf("failed to create ledger: %w", err)
			}
			ctx.Logger().Info("Created ledger", "nft_id", ledgerToEntries.Ledger.Key.NftId, "asset_class", ledgerToEntries.Ledger.Key.AssetClassId)
		} else if ledgerToEntries.Ledger != nil {
			ctx.Logger().Info("Ledger already exists, skipping creation", "nft_id", ledgerToEntries.Ledger.Key.NftId, "asset_class", ledgerToEntries.Ledger.Key.AssetClassId)
		}

		// If the ledger doesn't exist, we can't add entries to it
		if !k.HasLedger(ctx, ledgerToEntries.LedgerKey) {
			return fmt.Errorf("ledger %s does not exist", ledgerToEntries.LedgerKey.NftId)
		}

		// Add ledger entries
		if len(ledgerToEntries.Entries) > 0 {
			entries := make([]*ledger.LedgerEntry, len(ledgerToEntries.Entries))
			for i, entry := range ledgerToEntries.Entries {
				entries[i] = entry
			}

			if err := k.AppendEntries(ctx, maintainerAddr, ledgerToEntries.LedgerKey, entries); err != nil {
				return fmt.Errorf("failed to append entries for ledger key %s: %w", ledgerToEntries.LedgerKey.NftId, err)
			}
			ctx.Logger().Info("Added ledger entries", "ledger_key", ledgerToEntries.LedgerKey.NftId, "count", len(entries))
		}
	}

	ctx.Logger().Info("Successfully completed bulk import of ledger data",
		"ledger_to_entries", len(genesisState.LedgerToEntries))

	return nil
}
