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
func NewEventLedgerCreated(nftAddress, denom string) sdk.Event {
	return sdk.NewEvent(
		EventTypeLedgerCreated,
		sdk.NewAttribute("nft_address", nftAddress),
		sdk.NewAttribute("denom", denom),
	)
}

// NewEventLedgerConfigUpdated creates a new EventLedgerConfigUpdated event
func NewEventLedgerUpdated(nftAddress string) sdk.Event {
	return sdk.NewEvent(
		EventTypeLedgerConfigUpdated,
		sdk.NewAttribute("nft_address", nftAddress),
	)
}

// NewEventLedgerEntryAdded creates a new EventLedgerEntryAdded event
func NewEventLedgerEntryAdded(nftAddress, correlationID string) sdk.Event {
	return sdk.NewEvent(
		EventTypeLedgerEntryAdded,
		sdk.NewAttribute("nft_address", nftAddress),
		sdk.NewAttribute("correlation_id", correlationID),
	)
}
