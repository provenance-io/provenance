package keeper

import (
	"fmt"
	"sort"
	"strings"

	"cosmossdk.io/core/store"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	ledger "github.com/provenance-io/provenance/x/ledger/types"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
	"github.com/provenance-io/provenance/x/registry"
)

var _ Keeper = (*BaseKeeper)(nil)

type Keeper interface {
}

// Keeper defines the mymodule keeper.
type BaseKeeper struct {
	BaseViewKeeper
	BaseConfigKeeper
	BaseEntriesKeeper
	BaseFundTransferKeeper
}

var (
	ledgerPrefix                      = []byte{0x01}
	entriesPrefix                     = []byte{0x02}
	ledgerClassesPrefix               = []byte{0x03}
	ledgerClassEntryTypesPrefix       = []byte{0x04}
	ledgerClassStatusTypesPrefix      = []byte{0x05}
	ledgerClassBucketTypesPrefix      = []byte{0x06}
	fundTransfersPrefix               = []byte{0x07} // Reserved for future use...
	fundTransfersWithSettlementPrefix = []byte{0x08}
)

const (
	ledgerKeyHrp  = "ledger"
	settlementHrp = "settlement"
)

// NewKeeper returns a new mymodule Keeper.
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, storeService store.KVStoreService, bankKeeper BankKeeper, registryKeeper RegistryKeeper, metadataKeeper MetaDataKeeper) BaseKeeper {
	viewKeeper := NewBaseViewKeeper(cdc, storeKey, storeService, registryKeeper, metadataKeeper)

	return BaseKeeper{
		BaseViewKeeper: viewKeeper,
		BaseConfigKeeper: BaseConfigKeeper{
			BaseViewKeeper: viewKeeper,
			BankKeeper:     bankKeeper,
		},
		BaseEntriesKeeper: BaseEntriesKeeper{
			BaseViewKeeper: viewKeeper,
		},
		BaseFundTransferKeeper: BaseFundTransferKeeper{
			BankKeeper:     bankKeeper,
			BaseViewKeeper: viewKeeper,
		},
	}
}

// Combine the asset class id and nft id into a bech32 string.
// Using bech32 here just allows us a readable identifier for the ledger.
func LedgerKeyToString(key *ledger.LedgerKey) (*string, error) {
	joined := strings.Join([]string{key.AssetClassId, key.NftId}, ":")

	b32, err := bech32.ConvertAndEncode(ledgerKeyHrp, []byte(joined))
	if err != nil {
		return nil, err
	}

	return &b32, nil
}

func StringToLedgerKey(s string) (*ledger.LedgerKey, error) {
	hrp, b, err := bech32.DecodeAndConvert(s)
	if err != nil {
		return nil, err
	}

	if hrp != ledgerKeyHrp {
		return nil, fmt.Errorf("invalid hrp: %s", hrp)
	}

	parts := strings.Split(string(b), ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid key: %s", s)
	}

	return &ledger.LedgerKey{
		AssetClassId: parts[0],
		NftId:        parts[1],
	}, nil
}

func RequireAuthority(ctx sdk.Context, rk RegistryKeeper, addr string, key *registry.RegistryKey) error {
	has, err := assertAuthority(ctx, rk, addr, key)
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
		return ledger.NewLedgerCodedError(ledger.ErrCodeUnauthorized, fmt.Sprintf("authority (%s) is not the nft owner (%s)", authorityAddr, nftOwner.String()))
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

	lk := &ledger.LedgerKey{
		AssetClassId: rk.AssetClassId,
		NftId:        rk.NftId,
	}

	if registryEntry == nil {
		err = assertOwner(ctx, k, authorityAddr, lk)
		if err != nil {
			return false, err
		}

		return true, nil
	} else {
		// Since the authority doesn't have the servicer role, let's see if there is any servicer set. If there is, we'll return an error
		// so that only the assigned servicer can append entries.
		var servicerRegistered bool = false
		for _, role := range registryEntry.Roles {
			if role.Role == registry.RegistryRole_REGISTRY_ROLE_SERVICER {
				// Note that there is a registered servicer since we allow the owner to be the servicer if there is a registry without one.
				servicerRegistered = true
				for _, address := range role.Addresses {
					// Check if the authority is the servicer
					if address == authorityAddr {
						return true, nil
					}
				}

				return false, ledger.NewLedgerCodedError(ledger.ErrCodeUnauthorized, "registered servicer")
			}
		}

		if !servicerRegistered {
			err = assertOwner(ctx, k, authorityAddr, lk)
			if err != nil {
				return false, err
			}

			// The authority owns the asset, and there is no registered servicer
			return true, nil
		}
	}

	// Default to false if the authority is not the owner or servicer
	return false, nil
}

// createDefaultScopeSpec creates a default scope specification if it doesn't exist
func (k BaseKeeper) createDefaultScopeSpec(ctx sdk.Context, scopeSpecID string, authorityAddr sdk.AccAddress) error {
	// Check if the scope spec already exists
	metadataAddr, err := metadatatypes.MetadataAddressFromBech32(scopeSpecID)
	if err != nil {
		return fmt.Errorf("invalid scope spec ID %s: %w", scopeSpecID, err)
	}

	_, found := k.BaseViewKeeper.MetaDataKeeper.GetScopeSpecification(ctx, metadataAddr)
	if found {
		// Scope spec already exists, no need to create
		return nil
	}

	// Create a default scope specification
	scopeSpec := metadatatypes.ScopeSpecification{
		SpecificationId: metadataAddr,
		Description: &metadatatypes.Description{
			Name:        "Default Ledger Scope Specification",
			Description: "Auto-generated scope specification for ledger bulk import",
			WebsiteUrl:  "",
			IconUrl:     "",
		},
		OwnerAddresses:  []string{authorityAddr.String()},
		PartiesInvolved: []metadatatypes.PartyType{metadatatypes.PartyType_PARTY_TYPE_OWNER},
		ContractSpecIds: []metadatatypes.MetadataAddress{},
	}

	k.BaseViewKeeper.MetaDataKeeper.SetScopeSpecification(ctx, scopeSpec)
	ctx.Logger().Info("Created default scope specification", "scope_spec_id", scopeSpecID)
	return nil
}

// createDefaultScope creates a default scope if it doesn't exist
func (k BaseKeeper) createDefaultScope(ctx sdk.Context, scopeID string, scopeSpecID string, authorityAddr sdk.AccAddress) error {
	// Check if the scope already exists
	scopeMetadataAddr, err := metadatatypes.MetadataAddressFromBech32(scopeID)
	if err != nil {
		return fmt.Errorf("invalid scope ID %s: %w", scopeID, err)
	}

	_, found := k.BaseViewKeeper.MetaDataKeeper.GetScope(ctx, scopeMetadataAddr)
	if found {
		// Scope already exists, no need to create
		return nil
	}

	// Get the scope spec ID
	scopeSpecMetadataAddr, err := metadatatypes.MetadataAddressFromBech32(scopeSpecID)
	if err != nil {
		return fmt.Errorf("invalid scope spec ID %s: %w", scopeSpecID, err)
	}

	// Create a default scope
	scope := metadatatypes.Scope{
		ScopeId:           scopeMetadataAddr,
		SpecificationId:   scopeSpecMetadataAddr,
		Owners:            []metadatatypes.Party{{Address: authorityAddr.String(), Role: metadatatypes.PartyType_PARTY_TYPE_OWNER}, {Address: authorityAddr.String(), Role: metadatatypes.PartyType_PARTY_TYPE_SERVICER}},
		DataAccess:        []string{authorityAddr.String()},
		ValueOwnerAddress: authorityAddr.String(),
	}

	err = k.BaseViewKeeper.MetaDataKeeper.SetScope(ctx, scope)
	if err != nil {
		return fmt.Errorf("failed to create default scope: %w", err)
	}

	ctx.Logger().Info("Created default scope", "scope_id", scopeID, "scope_spec_id", scopeSpecID)
	return nil
}

// ensureScopeAndScopeSpecExist ensures that both scope and scope spec exist, creating defaults if they don't
func (k BaseKeeper) ensureScopeAndScopeSpecExist(ctx sdk.Context, assetClassID string, nftID string, authorityAddr sdk.AccAddress) error {
	ctx.Logger().Error("Ensuring scope and scope spec exist", "scope_spec_id", assetClassID, "nft_id", nftID)

	// Check if this is a metadata scope (starts with "scope1" or "scopespec1")
	if strings.HasPrefix(assetClassID, "scopespec1") {
		// This is a scope spec ID, ensure it exists
		if err := k.createDefaultScopeSpec(ctx, assetClassID, authorityAddr); err != nil {
			// If the error is due to invalid checksum, skip this scope spec
			if strings.Contains(err.Error(), "invalid checksum") {
				ctx.Logger().Info("Skipping scope spec creation due to invalid checksum", "scope_spec_id", assetClassID)
				return nil
			}
			return err
		}
	}

	// If the NFT ID is also a scope ID, ensure it exists
	if strings.HasPrefix(nftID, "scope1") {
		if err := k.createDefaultScope(ctx, nftID, assetClassID, authorityAddr); err != nil {
			// If the error is due to invalid checksum, skip this scope
			if strings.Contains(err.Error(), "invalid checksum") {
				ctx.Logger().Info("Skipping scope creation due to invalid checksum", "scope_id", nftID)
				return nil
			}
			return err
		}
	}

	return nil
}

// BulkImportLedgerData imports ledger data from genesis state
// This function will create default scopes, scope specs, and ledger classes with all necessary types
// if they don't exist, using the provided authority address.
// It will also add any missing entry types or status types used in the genesis data to the ledger class.
func (k BaseKeeper) BulkImportLedgerData(ctx sdk.Context, authorityAddr sdk.AccAddress, genesisState ledger.GenesisState) error {
	ctx.Logger().Info("Starting bulk import of ledger data",
		"ledger_to_entries", len(genesisState.LedgerToEntries))

	// Second pass: import ledgers and their entries, ensuring all types exist
	for _, ledgerToEntries := range genesisState.LedgerToEntries {
		var maintainerAddr sdk.AccAddress
		var ledgerClassId string

		// Determine the ledger class ID and get maintainer address
		if ledgerToEntries.Ledger != nil {
			// If we have a ledger object, use its ledger class ID
			ledgerClassId = ledgerToEntries.Ledger.LedgerClassId

			// This is a new ledger so we need to ensure all the types are present
			// Ensure scope and scope spec exist if they are metadata addresses
			if err := k.ensureScopeAndScopeSpecExist(ctx, ledgerToEntries.LedgerKey.AssetClassId, ledgerToEntries.LedgerKey.NftId, authorityAddr); err != nil {
				return fmt.Errorf("failed to ensure scope and scope spec exist: %w", err)
			}

			// Ensure ledger class exists with all necessary types
			ctx.Logger().Info("Ensuring ledger class exists",
				"ledger_class_id", ledgerToEntries.Ledger.LedgerClassId,
				"asset_class_id", ledgerToEntries.LedgerKey.AssetClassId,
				"nft_id", ledgerToEntries.LedgerKey.NftId)

			// Validate that AssetClassId is not empty
			if ledgerToEntries.LedgerKey.AssetClassId == "" {
				return fmt.Errorf("asset_class_id is empty for ledger key with nft_id %s", ledgerToEntries.LedgerKey.NftId)
			}

			if err := k.EnsureLedgerClassExists(ctx, ledgerToEntries.Ledger.LedgerClassId, ledgerToEntries.LedgerKey.AssetClassId, authorityAddr); err != nil {
				return fmt.Errorf("failed to ensure ledger class exists: %w", err)
			}

			// Ensure the registry entry exists
			if err := k.EnsureRegistryEntryExists(ctx, ledgerToEntries.LedgerKey.AssetClassId, ledgerToEntries.LedgerKey.NftId, authorityAddr); err != nil {
				return fmt.Errorf("failed to ensure registry entry exists: %w", err)
			}
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
			return fmt.Errorf("ledger class %s not found after creation", ledgerClassId)
		}

		maintainerAddr, err = sdk.AccAddressFromBech32(ledgerClass.MaintainerAddress)
		if err != nil {
			return fmt.Errorf("invalid maintainer address %s: %w", ledgerClass.MaintainerAddress, err)
		}

		// Create the ledger only if it doesn't already exist
		if ledgerToEntries.Ledger != nil && !k.HasLedger(ctx, ledgerToEntries.Ledger.Key) {
			if err := k.CreateLedger(ctx, maintainerAddr, *ledgerToEntries.Ledger); err != nil {
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

// CreateDefaultLedgerClass creates a default ledger class with all necessary types if it doesn't exist
func (k BaseKeeper) CreateDefaultLedgerClass(ctx sdk.Context, ledgerClassID string, assetClassID string, authorityAddr sdk.AccAddress) error {
	ctx.Logger().Info("Creating default ledger class",
		"ledger_class_id", ledgerClassID,
		"asset_class_id", assetClassID,
		"authority_addr", authorityAddr.String())

	// Validate that assetClassID is not empty
	if assetClassID == "" {
		return fmt.Errorf("asset_class_id cannot be empty")
	}

	// Create a default ledger class
	defaultLedgerClass := ledger.LedgerClass{
		LedgerClassId:     ledgerClassID,
		AssetClassId:      assetClassID,
		MaintainerAddress: authorityAddr.String(),
		Denom:             "nhash", // Default denomination
	}

	if err := k.CreateLedgerClass(ctx, authorityAddr, defaultLedgerClass); err != nil {
		return fmt.Errorf("failed to create default ledger class: %w", err)
	}

	ctx.Logger().Info("Created default ledger class", "ledger_class_id", ledgerClassID)

	// Create default status types
	defaultStatusTypes := []ledger.LedgerClassStatusType{}
	
	// Fill in missing entry types for IDs 0-120
	statusTypeMap := make(map[int]ledger.LedgerClassStatusType)
	for _, et := range defaultStatusTypes {
		statusTypeMap[int(et.Id)] = et
	}
	for i := 0; i <= 120; i++ {
		if _, exists := statusTypeMap[i]; !exists {
			defaultStatusTypes = append(defaultStatusTypes, ledger.LedgerClassStatusType{
				Id:          int32(i),
				Code:        fmt.Sprintf("RANDOM_STATUS_%d", i),
				Description: fmt.Sprintf("Random Status %d", i),
			})
		}
	}
	// Sort by ID for consistency
	sort.Slice(defaultStatusTypes, func(i, j int) bool { return defaultStatusTypes[i].Id < defaultStatusTypes[j].Id })

	for _, statusType := range defaultStatusTypes {
		if err := k.AddClassStatusType(ctx, authorityAddr, ledgerClassID, statusType); err != nil {
			// Log the error but continue, as the status type might already exist
			ctx.Logger().Info("Failed to add status type (may already exist)", "status_type", statusType.Code, "error", err)
		}
	}

	// Create default entry types
	defaultEntryTypes := []ledger.LedgerClassEntryType{}
	
	// Fill in missing entry types for IDs 0-120
	entryTypeMap := make(map[int]ledger.LedgerClassEntryType)
	for _, et := range defaultEntryTypes {
		entryTypeMap[int(et.Id)] = et
	}
	for i := 0; i <= 120; i++ {
		if _, exists := entryTypeMap[i]; !exists {
			defaultEntryTypes = append(defaultEntryTypes, ledger.LedgerClassEntryType{
				Id:          int32(i),
				Code:        fmt.Sprintf("RANDOM_ENTRY_%d", i),
				Description: fmt.Sprintf("Random Entry %d", i),
			})
		}
	}
	// Sort by ID for consistency
	sort.Slice(defaultEntryTypes, func(i, j int) bool { return defaultEntryTypes[i].Id < defaultEntryTypes[j].Id })

	for _, entryType := range defaultEntryTypes {
		if err := k.AddClassEntryType(ctx, authorityAddr, ledgerClassID, entryType); err != nil {
			// Log the error but continue, as the entry type might already exist
			ctx.Logger().Info("Failed to add entry type (may already exist)", "entry_type", entryType.Code, "error", err)
		}
	}

	// Create default bucket types
	defaultBucketTypes := []ledger.LedgerClassBucketType{}
	// Fill in missing bucket types for IDs 0-60
	bucketTypeMap := make(map[int]ledger.LedgerClassBucketType)
	for _, bt := range defaultBucketTypes {
		bucketTypeMap[int(bt.Id)] = bt
	}
	for i := 0; i <= 60; i++ {
		defaultBucketTypes = append(defaultBucketTypes, ledger.LedgerClassBucketType{
			Id:          int32(i),
			Code:        fmt.Sprintf("RANDOM_BUCKET_%d", i),
			Description: fmt.Sprintf("Random Bucket %d", i),
		})
	}
	// Sort by ID for consistency
	sort.Slice(defaultBucketTypes, func(i, j int) bool { return defaultBucketTypes[i].Id < defaultBucketTypes[j].Id })

	for _, bucketType := range defaultBucketTypes {
		if err := k.AddClassBucketType(ctx, authorityAddr, ledgerClassID, bucketType); err != nil {
			// Log the error but continue, as the bucket type might already exist
			ctx.Logger().Info("Failed to add bucket type (may already exist)", "bucket_type", bucketType.Code, "error", err)
		}
	}

	ctx.Logger().Info("Created default types for ledger class", "ledger_class_id", ledgerClassID)
	return nil
}

// EnsureLedgerClassExists ensures that a ledger class exists with all necessary types
func (k BaseKeeper) EnsureLedgerClassExists(ctx sdk.Context, ledgerClassID string, assetClassID string, authorityAddr sdk.AccAddress) error {
	// Validate that assetClassID is not empty
	if assetClassID == "" {
		return fmt.Errorf("asset_class_id cannot be empty for ledger class %s", ledgerClassID)
	}

	// Check if the ledger class exists
	ledgerClass, _ := k.GetLedgerClass(ctx, ledgerClassID)
	if ledgerClass == nil {
		// Create default ledger class with all types
		if err := k.CreateDefaultLedgerClass(ctx, ledgerClassID, assetClassID, authorityAddr); err != nil {
			return fmt.Errorf("failed to create default ledger class: %w", err)
		}
	}

	return nil
}

// EnsureRegistryEntryExists ensures that a registry entry exists for the given asset class and NFT
func (k BaseKeeper) EnsureRegistryEntryExists(ctx sdk.Context, assetClassID string, nftID string, authorityAddr sdk.AccAddress) error {
	// Create registry key
	registryKey := &registry.RegistryKey{
		AssetClassId: assetClassID,
		NftId:        nftID,
	}

	// Check if registry entry already exists
	existingRegistry, err := k.BaseViewKeeper.RegistryKeeper.GetRegistry(ctx, registryKey)
	if err != nil {
		return fmt.Errorf("failed to check existing registry: %w", err)
	}

	// If registry entry doesn't exist, create a default one
	if existingRegistry == nil {
		if err := k.BaseViewKeeper.RegistryKeeper.CreateDefaultRegistry(ctx, authorityAddr, registryKey); err != nil {
			return fmt.Errorf("failed to create default registry: %w", err)
		}
		ctx.Logger().Info("Created default registry entry", "asset_class_id", assetClassID, "nft_id", nftID)
	}

	return nil
}
