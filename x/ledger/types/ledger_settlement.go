package types

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Validate validates the FundTransferWithSettlement type
func (ft *FundTransferWithSettlement) Validate() error {
	if ft.Key == nil {
		return fmt.Errorf("key cannot be nil")
	}

	var errs []error
	if err := ft.Key.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("key: %w", err))
	}

	// Validate that the correlation ID is valid
	if err := lenCheck(ft.LedgerEntryCorrelationId, 1, MaxLenCorrelationID); err != nil {
		errs = append(errs, fmt.Errorf("ledger_entry_correlation_id: %w", err))
	}

	// Validate that the settlement instructions are valid
	for _, si := range ft.SettlementInstructions {
		if err := si.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("settlement_instructions: %w", err))
		}
	}

	return errors.Join(errs...)
}

// Validate validates the SettlementInstruction type
func (si *SettlementInstruction) Validate() error {
	var errs []error
	if err := si.Amount.Validate(); err != nil {
		errs = append(errs, err)
	}
	if si.Amount.IsNegative() {
		errs = append(errs, fmt.Errorf("amount: must be a non-negative integer"))
	}

	if err := lenCheck(si.Memo, 0, MaxLenMemo); err != nil {
		errs = append(errs, fmt.Errorf("memo: %w", err))
	}

	// Validate recipient address format
	if _, err := sdk.AccAddressFromBech32(si.RecipientAddress); err != nil {
		errs = append(errs, fmt.Errorf("recipient_address: %w", err))
	}

	// Validate status enum
	if err := si.Status.Validate(); err != nil {
		errs = append(errs, NewErrCodeInvalidField("status", "invalid funding transfer status"))
	}

	return errors.Join(errs...)
}

// UnmarshalJSON implements json.Unmarshaler for FundingTransferStatus.
func (s *FundingTransferStatus) UnmarshalJSON(data []byte) error {
	value, err := enumUnmarshalJSON(data, FundingTransferStatus_value, FundingTransferStatus_name)
	if err != nil {
		return err
	}
	*s = FundingTransferStatus(value)
	return nil
}

// Validate returns an error if this FundingTransferStatus isn't a defined enum entry.
func (s FundingTransferStatus) Validate() error {
	return enumValidateExists(s, FundingTransferStatus_name)
}
