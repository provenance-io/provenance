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
func NewEventLedgerCreated(nftId string) sdk.Event {
	return sdk.NewEvent(
		EventTypeLedgerCreated,
		sdk.NewAttribute("nft_id", nftId),
	)
}

// NewEventLedgerConfigUpdated creates a new EventLedgerConfigUpdated event
func NewEventLedgerUpdated(nftId string) sdk.Event {
	return sdk.NewEvent(
		EventTypeLedgerConfigUpdated,
		sdk.NewAttribute("nft_id", nftId),
	)
}

// NewEventLedgerEntryAdded creates a new EventLedgerEntryAdded event
func NewEventLedgerEntryAdded(nftId, correlationID string) sdk.Event {
	return sdk.NewEvent(
		EventTypeLedgerEntryAdded,
		sdk.NewAttribute("nft_id", nftId),
		sdk.NewAttribute("correlation_id", correlationID),
	)
}
