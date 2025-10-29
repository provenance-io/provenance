package types

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/internal/provutils"
)

// Validate validates the FundTransferWithSettlement type
func (ft *FundTransferWithSettlement) Validate() error {
	if ft == nil {
		return fmt.Errorf("fund_transfer_with_settlement cannot be nil")
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
	if err := validateSlice(ft.SettlementInstructions, "settlement_instructions"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// Validate validates the SettlementInstruction type
func (si *SettlementInstruction) Validate() error {
	if si == nil {
		return fmt.Errorf("settlement_instruction cannot be nil")
	}

	var errs []error
	if err := si.Amount.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("amount: %w", err))
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
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// UnmarshalJSON implements json.Unmarshaler for FundingTransferStatus.
func (s *FundingTransferStatus) UnmarshalJSON(data []byte) error {
	value, err := provutils.EnumUnmarshalJSON(data, FundingTransferStatus_value, FundingTransferStatus_name)
	if err != nil {
		return err
	}
	*s = FundingTransferStatus(value)
	return nil
}

// Validate returns an error if this FundingTransferStatus isn't a defined enum entry.
func (s FundingTransferStatus) Validate() error {
	return provutils.EnumValidateExists(s, FundingTransferStatus_name)
}
