package ledger

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	EventTypeLedgerCreated       = "ledger_created"
	EventTypeLedgerConfigUpdated = "ledger_config_updated"
	EventTypeLedgerEntryAdded    = "ledger_entry_added"
)

// NewEventLedgerCreated creates a new EventLedgerCreated event
func NewEventLedgerCreated(key *LedgerKey) sdk.Event {
	return sdk.NewEvent(
		EventTypeLedgerCreated,
		sdk.NewAttribute("asset_class_id", key.AssetClassId),
		sdk.NewAttribute("nft_id", key.NftId),
	)
}

// NewEventLedgerConfigUpdated creates a new EventLedgerConfigUpdated event
func NewEventLedgerUpdated(key *LedgerKey) sdk.Event {
	return sdk.NewEvent(
		EventTypeLedgerConfigUpdated,
		sdk.NewAttribute("asset_class_id", key.AssetClassId),
		sdk.NewAttribute("nft_id", key.NftId),
	)
}

// NewEventLedgerEntryAdded creates a new EventLedgerEntryAdded event
func NewEventLedgerEntryAdded(key *LedgerKey, correlationID string) sdk.Event {
	return sdk.NewEvent(
		EventTypeLedgerEntryAdded,
		sdk.NewAttribute("asset_class_id", key.AssetClassId),
		sdk.NewAttribute("nft_id", key.NftId),
		sdk.NewAttribute("correlation_id", correlationID),
	)
}
