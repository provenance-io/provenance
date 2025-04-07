package ledger

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	EventTypeLedgerCreated       = "ledger_created"
	EventTypeLedgerEntryAdded    = "ledger_entry_added"
	EventTypeBalanceUpdated      = "ledgerbalance_updated"
	EventTypeLedgerConfigUpdated = "ledger_config_updated"
)

// NewEventLedgerCreated creates a new EventLedgerCreated event
func NewEventLedgerCreated(nftAddress, denom string) sdk.Event {
	return sdk.NewEvent(
		EventTypeLedgerCreated,
		sdk.NewAttribute("nft_address", nftAddress),
		sdk.NewAttribute("denom", denom),
	)
}

// NewEventLedgerEntryAdded creates a new EventLedgerEntryAdded event
func NewEventLedgerEntryAdded(nftAddress, entryUUID string, entryType int32, postedDate, effectiveDate time.Time, amount string) sdk.Event {
	return sdk.NewEvent(
		EventTypeLedgerEntryAdded,
		sdk.NewAttribute("nft_address", nftAddress),
		sdk.NewAttribute("entry_uuid", entryUUID),
		sdk.NewAttribute("entry_type", string(entryType)),
		sdk.NewAttribute("posted_date", postedDate.Format(time.RFC3339)),
		sdk.NewAttribute("effective_date", effectiveDate.Format(time.RFC3339)),
		sdk.NewAttribute("amount", amount),
	)
}

// NewEventBalanceUpdated creates a new EventBalanceUpdated event
func NewEventBalanceUpdated(nftAddress, principalBalance, interestBalance, otherBalance string) sdk.Event {
	return sdk.NewEvent(
		EventTypeBalanceUpdated,
		sdk.NewAttribute("nft_address", nftAddress),
		sdk.NewAttribute("principal_balance", principalBalance),
		sdk.NewAttribute("interest_balance", interestBalance),
		sdk.NewAttribute("other_balance", otherBalance),
	)
}

// NewEventLedgerConfigUpdated creates a new EventLedgerConfigUpdated event
func NewEventLedgerConfigUpdated(nftAddress, denom, previousDenom string) sdk.Event {
	return sdk.NewEvent(
		EventTypeLedgerConfigUpdated,
		sdk.NewAttribute("nft_address", nftAddress),
		sdk.NewAttribute("denom", denom),
		sdk.NewAttribute("previous_denom", previousDenom),
	)
}
