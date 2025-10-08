package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

// DefaultGenesisState returns the initial set of name -> address bindings.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{}
}

// Validate performs genesis state validation returning an error upon any failure.
// It calls the Validate function for each entry in the GenesisState.
func (genesisState GenesisState) Validate() error {
	// Validate each ledger class.
	for i, ledgerClass := range genesisState.LedgerClasses {
		err := ledgerClass.Validate()
		if err != nil {
			return sdkerrors.ErrInvalidType.Wrapf(err.Error(), "LedgerClass %d validation failed", i)
		}
	}

	// Validate each ledger class entry type.
	for i, entryType := range genesisState.LedgerClassEntryTypes {
		err := entryType.EntryType.Validate()
		if err != nil {
			return sdkerrors.ErrInvalidType.Wrapf(err.Error(), "LedgerClassEntryType %d validation failed", i)
		}
	}

	// Validate each ledger class status type.
	for i, statusType := range genesisState.LedgerClassStatusTypes {
		err := statusType.StatusType.Validate()
		if err != nil {
			return sdkerrors.ErrInvalidType.Wrapf(err.Error(), "LedgerClassStatusType %d validation failed", i)
		}
	}

	// Validate each ledger class bucket type.
	for i, bucketType := range genesisState.LedgerClassBucketTypes {
		err := bucketType.BucketType.Validate()
		if err != nil {
			return sdkerrors.ErrInvalidType.Wrapf(err.Error(), "LedgerClassBucketType %d validation failed", i)
		}
	}

	// Validate each ledger.
	for i, ledgers := range genesisState.Ledgers {
		err := ledgers.Ledger.Validate()
		if err != nil {
			return sdkerrors.ErrInvalidType.Wrapf(err.Error(), "Ledger %d validation failed", i)
		}
	}

	// Validate each ledger entry.
	for i, ledgerEntries := range genesisState.LedgerEntries {
		err := ledgerEntries.Entry.Validate()
		if err != nil {
			return sdkerrors.ErrInvalidType.Wrapf(err.Error(), "LedgerEntry %d validation failed", i)
		}
	}

	// Validate each settlement instruction.
	for i, settlement := range genesisState.SettlementInstructions {
		for j, instruction := range settlement.SettlementInstructions.SettlementInstructions {
			err := instruction.Validate()
			if err != nil {
				return sdkerrors.ErrInvalidType.Wrapf(err.Error(), "SettlementInstruction %d.%d validation failed", i, j)
			}
		}
	}

	return nil
}
