package types

// Validate validates the FundTransferWithSettlement type
func (ft *FundTransferWithSettlement) Validate() error {
	if ft.Key == nil {
		return NewErrCodeMissingField("key")
	}

	if err := ft.Key.Validate(); err != nil {
		return err
	}

	// Validate that the correlation ID is valid
	if err := lenCheck("ledger_entry_correlation_id", ft.LedgerEntryCorrelationId, 1, 50); err != nil {
		return err
	}

	// Validate that the settlement instructions are valid
	for _, si := range ft.SettlementInstructions {
		if err := si.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate validates the SettlementInstruction type
func (si *SettlementInstruction) Validate() error {
	if err := si.Amount.Validate(); err != nil {
		return NewErrCodeInvalidField("amount", err.Error())
	}
	if si.Amount.IsNegative() {
		return NewErrCodeInvalidField("amount", "must be a non-negative integer")
	}

	if err := lenCheck("memo", si.Memo, 0, 50); err != nil {
		return err
	}

	// Validate recipient address format
	if err := validateAccAddress("recipient_address", si.RecipientAddress); err != nil {
		return err
	}

	// Validate status enum
	if _, ok := FundingTransferStatus_name[int32(si.Status)]; !ok {
		return NewErrCodeInvalidField("status", "invalid funding transfer status")
	}

	return nil
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
