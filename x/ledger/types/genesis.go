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
			return sdkerrors.ErrInvalidType.Wrapf("ledger class %d validation failed: %v", i, err)
		}
	}

	// Validate each ledger class entry type.
	for i, entryType := range genesisState.LedgerClassEntryTypes {
		err := entryType.EntryType.Validate()
		if err != nil {
			return sdkerrors.ErrInvalidType.Wrapf("ledger class entry type %d validation failed: %v", i, err)
		}
	}

	// Validate each ledger class status type.
	for i, statusType := range genesisState.LedgerClassStatusTypes {
		err := statusType.StatusType.Validate()
		if err != nil {
			return sdkerrors.ErrInvalidType.Wrapf("ledger class status Type %d validation failed: %v", i, err)
		}
	}

	// Validate each ledger class bucket type.
	for i, bucketType := range genesisState.LedgerClassBucketTypes {
		err := bucketType.BucketType.Validate()
		if err != nil {
			return sdkerrors.ErrInvalidType.Wrapf("ledger class bucket type %d validation failed: %v", i, err)
		}
	}

	// Validate each ledger.
	for i, ledgers := range genesisState.Ledgers {
		err := ledgers.Ledger.Validate()
		if err != nil {
			return sdkerrors.ErrInvalidType.Wrapf("ledger %d validation failed: %v", i, err)
		}
	}

	// Validate each ledger entry.
	for i, ledgerEntries := range genesisState.LedgerEntries {
		err := ledgerEntries.Entry.Validate()
		if err != nil {
			return sdkerrors.ErrInvalidType.Wrapf("ledger entry %d validation failed: %v", i, err)
		}
	}

	// Validate each settlement instruction.
	for i, settlement := range genesisState.SettlementInstructions {
		for j, instruction := range settlement.SettlementInstructions.SettlementInstructions {
			err := instruction.Validate()
			if err != nil {
				return sdkerrors.ErrInvalidType.Wrapf("settlement instruction %d.%d validation failed: %v", i, j, err)
			}
		}
	}

	return nil
}
