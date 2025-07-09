package keeper

import (
	"fmt"
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
	defaultStatusTypes := []ledger.LedgerClassStatusType{
		{Id: 1, Code: "IN_REPAYMENT", Description: "In Repayment"},
		{Id: 24, Code: "DEFAULTED", Description: "Defaulted"},
		{Id: 2, Code: "IN_FORECLOSURE", Description: "In Foreclosure"},
		{Id: 3, Code: "FORBEARANCE", Description: "Forbearance"},
		{Id: 4, Code: "DEFERMENT", Description: "Deferment"},
		{Id: 5, Code: "BANKRUPTCY", Description: "Bankruptcy"},
		{Id: 6, Code: "CLOSED", Description: "Closed"},
		{Id: 7, Code: "CANCELLED", Description: "Cancelled"},
		{Id: 8, Code: "SUSPENDED", Description: "Suspended"},
		{Id: 9, Code: "OTHER", Description: "Other"},
	}

	for _, statusType := range defaultStatusTypes {
		if err := k.AddClassStatusType(ctx, authorityAddr, ledgerClassID, statusType); err != nil {
			// Log the error but continue, as the status type might already exist
			ctx.Logger().Info("Failed to add status type (may already exist)", "status_type", statusType.Code, "error", err)
		}
	}

	// Create default entry types
	defaultEntryTypes := []ledger.LedgerClassEntryType{
		{Id: 0, Code: "UNKNOWN", Description: "UNKNOWN"},
		{Id: 1, Code: "DISBURSEMENT", Description: "DISBURSEMENT"},
		{Id: 2, Code: "DRAW_DOWN", Description: "DRAW_DOWN"},
		{Id: 3, Code: "REFUND", Description: "REFUND"},
		{Id: 10, Code: "AUTOMATIC_PAYMENT", Description: "AUTOMATIC_PAYMENT"},
		{Id: 11, Code: "AUTOMATIC_PAYMENT_IN_KIND", Description: "AUTOMATIC_PAYMENT_IN_KIND"},
		{Id: 12, Code: "MANUAL_PAYMENT", Description: "MANUAL_PAYMENT"},
		{Id: 13, Code: "MANUAL_PAYMENT_IN_KIND", Description: "MANUAL_PAYMENT_IN_KIND"},
		{Id: 14, Code: "CREDIT", Description: "CREDIT"},
		{Id: 15, Code: "PAYOFF", Description: "PAYOFF"},
		{Id: 16, Code: "SUSPENDED_PAYMENT", Description: "SUSPENDED_PAYMENT"},
		{Id: 17, Code: "REQUESTED_ALLOCATION_MANUAL_PAYMENT", Description: "REQUESTED_ALLOCATION_MANUAL_PAYMENT"},
		{Id: 18, Code: "INTEREST_PREPAYMENT", Description: "INTEREST_PREPAYMENT"},
		{Id: 19, Code: "EXTERNAL_PAYMENT", Description: "EXTERNAL_PAYMENT"},
		{Id: 20, Code: "CREDIT_SETTLEMENT", Description: "CREDIT_SETTLEMENT"},
		{Id: 30, Code: "FEE", Description: "FEE"},
		{Id: 31, Code: "ORIGINATION_FEE", Description: "ORIGINATION_FEE"},
		{Id: 32, Code: "LATE_FEE", Description: "LATE_FEE"},
		{Id: 33, Code: "NSF_FEE", Description: "NSF_FEE"},
		{Id: 34, Code: "RELEASE_FEE", Description: "RELEASE_FEE"},
		{Id: 35, Code: "SUBORDINATION_FEE", Description: "SUBORDINATION_FEE"},
		{Id: 36, Code: "MANUAL_NOTARY_COST", Description: "MANUAL_NOTARY_COST"},
		{Id: 37, Code: "MANUAL_NOTARY_ATTORNEY_FEE", Description: "MANUAL_NOTARY_ATTORNEY_FEE"},
		{Id: 38, Code: "INTANGIBLE_TAX", Description: "INTANGIBLE_TAX"},
		{Id: 39, Code: "MORTGAGE_TAX", Description: "MORTGAGE_TAX"},
		{Id: 40, Code: "STATE_TRANSFER_TAX", Description: "STATE_TRANSFER_TAX"},
		{Id: 41, Code: "LOCAL_TRANSFER_TAX", Description: "LOCAL_TRANSFER_TAX"},
		{Id: 42, Code: "RECORDATION_TRANSFER_STAMP_TAX", Description: "RECORDATION_TRANSFER_STAMP_TAX"},
		{Id: 43, Code: "FULL_PROPERTY_REPORT", Description: "FULL_PROPERTY_REPORT"},
		{Id: 44, Code: "INSTANT_TITLE", Description: "INSTANT_TITLE"},
		{Id: 45, Code: "PRE_PAYMENT_FEE", Description: "PRE_PAYMENT_FEE"},
		{Id: 50, Code: "ADJUSTMENT", Description: "ADJUSTMENT"},
		{Id: 51, Code: "REVERSAL", Description: "REVERSAL"},
		{Id: 52, Code: "RECALCULATION_ADJUSTMENT", Description: "RECALCULATION_ADJUSTMENT"},
		{Id: 53, Code: "FUNDING_BOUNCE_INTEREST_ADJUSTMENT", Description: "FUNDING_BOUNCE_INTEREST_ADJUSTMENT"},
		{Id: 54, Code: "COLLATERAL_LIQUIDATION_RECALCULATION_ADJUSTMENT", Description: "COLLATERAL_LIQUIDATION_RECALCULATION_ADJUSTMENT"},
		{Id: 60, Code: "COLLATERAL_LIQUIDATION", Description: "COLLATERAL_LIQUIDATION"},
		{Id: 61, Code: "COLLATERAL_SLIPPAGE", Description: "COLLATERAL_SLIPPAGE"},
		{Id: 70, Code: "ESCROW_REFUND", Description: "ESCROW_REFUND"},
		{Id: 71, Code: "ESCROW_PAYMENT", Description: "ESCROW_PAYMENT"},
		{Id: 72, Code: "SERVICER_ADVANCE", Description: "SERVICER_ADVANCE"},
		{Id: 73, Code: "LENDER_PLACED_FLOOD_INSURANCE", Description: "LENDER_PLACED_FLOOD_INSURANCE"},
		{Id: 74, Code: "LENDER_PLACED_HAZARD_INSURANCE", Description: "LENDER_PLACED_HAZARD_INSURANCE"},
		{Id: 75, Code: "ESCROW_PREPAYMENT", Description: "ESCROW_PREPAYMENT"},
		{Id: 76, Code: "INSURANCE_DISBURSEMENT", Description: "INSURANCE_DISBURSEMENT"},
		{Id: 77, Code: "TAX_DISBURSEMENT", Description: "TAX_DISBURSEMENT"},
		{Id: 78, Code: "ESCROW_CLOSEOUT", Description: "ESCROW_CLOSEOUT"},
		{Id: 79, Code: "INTEREST_ON_ESCROW", Description: "INTEREST_ON_ESCROW"},
		{Id: 90, Code: "DEFERMENT", Description: "DEFERMENT"},
		{Id: 91, Code: "RATE_CHANGE", Description: "RATE_CHANGE"},
		{Id: 92, Code: "SETTLEMENT_WRITE_OFF", Description: "SETTLEMENT_WRITE_OFF"},
		{Id: 93, Code: "FORECLOSURE", Description: "FORECLOSURE"},
		{Id: 94, Code: "CLOSING_CREDIT", Description: "CLOSING_CREDIT"},
		{Id: 95, Code: "CLOSING_COST", Description: "CLOSING_COST"},
		{Id: 96, Code: "PRE_PETITION_DEFERMENT", Description: "PRE_PETITION_DEFERMENT"},
		{Id: 97, Code: "BPO", Description: "BPO"},
		{Id: 98, Code: "DRIVE_BY_VALUATION", Description: "DRIVE_BY_VALUATION"},
		{Id: 99, Code: "DIRECT_DEBT_LIABILITY", Description: "DIRECT_DEBT_LIABILITY"},
		{Id: 100, Code: "DIRECT_DEBT_LIABILITY_CREDIT", Description: "DIRECT_DEBT_LIABILITY_CREDIT"},
		{Id: 101, Code: "VOID_DRAW", Description: "VOID_DRAW"},
		{Id: 102, Code: "UPFRONT_ORIGINATION_FEE", Description: "UPFRONT_ORIGINATION_FEE"},
		{Id: 103, Code: "SUB_SERVICER_TRANSACTION", Description: "SUB_SERVICER_TRANSACTION"},
		{Id: 104, Code: "PARTIAL_RELEASE_FEE", Description: "PARTIAL_RELEASE_FEE"},
	}

	for _, entryType := range defaultEntryTypes {
		if err := k.AddClassEntryType(ctx, authorityAddr, ledgerClassID, entryType); err != nil {
			// Log the error but continue, as the entry type might already exist
			ctx.Logger().Info("Failed to add entry type (may already exist)", "entry_type", entryType.Code, "error", err)
		}
	}

	// Create default bucket types
	defaultBucketTypes := []ledger.LedgerClassBucketType{
		{Id: 0, Code: "UNKNOWN", Description: "UNKNOWN"},
		{Id: 1, Code: "FEE", Description: "FEE"},
		{Id: 2, Code: "FLOOD", Description: "FLOOD"},
		{Id: 3, Code: "HAZARD", Description: "HAZARD"},
		{Id: 4, Code: "INTEREST", Description: "INTEREST"},
		{Id: 5, Code: "CAP_ORIGINATION_FEE", Description: "CAP_ORIGINATION_FEE"},
		{Id: 6, Code: "NON_CAP_ORIGINATION_FEE", Description: "NON_CAP_ORIGINATION_FEE"},
		{Id: 7, Code: "PRINCIPAL", Description: "PRINCIPAL"},
		{Id: 8, Code: "PRINCIPAL_OVERPAY", Description: "PRINCIPAL_OVERPAY"},
		{Id: 9, Code: "CREDIT", Description: "CREDIT"},
		{Id: 10, Code: "RECORDING_FEE", Description: "RECORDING_FEE"},
		{Id: 11, Code: "MORTGAGE_INSURANCE", Description: "MORTGAGE_INSURANCE"},
		{Id: 12, Code: "HOMEOWNER_TAXES", Description: "HOMEOWNER_TAXES"},
		{Id: 13, Code: "HOA_FEE", Description: "HOA_FEE"},
		{Id: 14, Code: "LATE_FEE", Description: "LATE_FEE"},
		{Id: 15, Code: "INTEREST_PREPAYMENT", Description: "INTEREST_PREPAYMENT"},
		{Id: 16, Code: "ESCROW", Description: "ESCROW"},
		{Id: 17, Code: "NSF_FEE", Description: "NSF_FEE"},
		{Id: 18, Code: "DEFERRED_INTEREST", Description: "DEFERRED_INTEREST"},
		{Id: 19, Code: "INVESTOR_RECOVERABLE_FEES", Description: "INVESTOR_RECOVERABLE_FEES"},
		{Id: 20, Code: "DEFERRED_PRINCIPAL", Description: "DEFERRED_PRINCIPAL"},
		{Id: 21, Code: "BORROWER_RECOVERABLE_FEES", Description: "BORROWER_RECOVERABLE_FEES"},
		{Id: 22, Code: "RESTRICTED_ESCROW", Description: "RESTRICTED_ESCROW"},
		{Id: 23, Code: "SUSPENSE", Description: "SUSPENSE"},
		{Id: 24, Code: "LENDER_PLACED_FLOOD_INSURANCE", Description: "LENDER_PLACED_FLOOD_INSURANCE"},
		{Id: 25, Code: "LENDER_PLACED_HAZARD_INSURANCE", Description: "LENDER_PLACED_HAZARD_INSURANCE"},
		{Id: 26, Code: "REPORTED_MORT_DSI", Description: "REPORTED_MORT_DSI"},
		{Id: 27, Code: "DEFERRED_INTEREST_V2", Description: "DEFERRED_INTEREST_V2"},
		{Id: 28, Code: "SUBORDINATION_FEE", Description: "SUBORDINATION_FEE"},
		{Id: 29, Code: "SERVICER_ADVANCE_PROPERTY_TAX_REPAYMENT", Description: "SERVICER_ADVANCE_PROPERTY_TAX_REPAYMENT"},
		{Id: 30, Code: "SERVICER_ADVANCE_HOA_REPAYMENT", Description: "SERVICER_ADVANCE_HOA_REPAYMENT"},
		{Id: 31, Code: "SERVICER_ADVANCE_PROPERTY_PRESERVATION", Description: "SERVICER_ADVANCE_PROPERTY_PRESERVATION"},
		{Id: 32, Code: "SERVICER_ADVANCE_DELINQUENCY_EXPENSE", Description: "SERVICER_ADVANCE_DELINQUENCY_EXPENSE"},
		{Id: 33, Code: "SERVICER_ADVANCE_LEGAL_EXPENSE", Description: "SERVICER_ADVANCE_LEGAL_EXPENSE"},
		{Id: 34, Code: "ESCROW_INTEREST", Description: "ESCROW_INTEREST"},
		{Id: 35, Code: "ESCROW_SHORTAGE", Description: "ESCROW_SHORTAGE"},
		{Id: 36, Code: "ESCROW_ADVANCE", Description: "ESCROW_ADVANCE"},
		{Id: 37, Code: "SETTLEMENT_WRITE_OFF", Description: "SETTLEMENT_WRITE_OFF"},
		{Id: 38, Code: "DELINQUENCY_REPAYMENT_PLAN", Description: "DELINQUENCY_REPAYMENT_PLAN"},
		{Id: 39, Code: "PRE_PETITION_PRINCIPAL_DEFERRED", Description: "PRE_PETITION_PRINCIPAL_DEFERRED"},
		{Id: 40, Code: "PRE_PETITION_INTEREST_DEFERRED", Description: "PRE_PETITION_INTEREST_DEFERRED"},
		{Id: 41, Code: "DEFERRAL_FEE", Description: "DEFERRAL_FEE"},
		{Id: 42, Code: "PARTIAL_RELEASE_FEE", Description: "PARTIAL_RELEASE_FEE"},
		{Id: 43, Code: "PRE_PAYMENT_FEE", Description: "PRE_PAYMENT_FEE"},
	}

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
