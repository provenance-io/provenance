// Formatting functions/structs for ledger queries that happen via the cli. These are necessary due to having most
// of the types stored on chain as integers. We map on the client side to avoid having that complexity in the
// module's keeper.
package cli

import (
	ledger "github.com/provenance-io/provenance/x/ledger/types"
)

// LedgerPlainText represents a ledger in plain text format
type LedgerPlainText struct {
	// Ledger key
	Key *ledger.LedgerKey `json:"key,omitempty"`
	// Status of the ledger
	Status string `json:"status,omitempty"`
	// Next payment date
	NextPmtDate string `json:"next_pmt_date,omitempty"`
	// Next payment amount
	NextPmtAmt string `json:"next_pmt_amt,omitempty"`
	// Interest rate
	InterestRate string `json:"interest_rate,omitempty"`
	// Maturity date
	MaturityDate string `json:"maturity_date,omitempty"`
	// Day count convention for interest
	InterestDayCountConvention ledger.DayCountConvention `json:"interest_day_count_convention,omitempty"`
	// Interest accrual method for interest
	InterestAccrualMethod ledger.InterestAccrualMethod `json:"interest_accrual_method,omitempty"`
	// Payment frequency
	PaymentFrequency ledger.PaymentFrequency `json:"payment_frequency,omitempty"`
}

// LedgerEntryPlainText represents a ledger entry in plain text format
type LedgerEntryPlainText struct {
	// Correlation ID for tracking ledger entries with external systems (max 50 characters)
	CorrelationId string `json:"correlation_id,omitempty"`
	// Sequence number of the ledger entry (less than 100)
	// This field is used to maintain the correct order of entries when multiple entries
	// share the same effective date. Entries are sorted first by effective date, then by sequence.
	Sequence uint32 `json:"sequence,omitempty"`
	// The type of ledger entry specified by the LedgerClassEntryType.id
	Type *ledger.LedgerClassEntryType `json:"type,omitempty"`
	// Posted date
	PostedDate string `json:"posted_date,omitempty"`
	// Effective date
	EffectiveDate string `json:"effective_date,omitempty"`
	// The total amount of the ledger entry
	TotalAmt string `json:"total_amt,omitempty"`
	// The amounts applied to each bucket
	AppliedAmounts []*LedgerBucketAmountPlainText `json:"applied_amounts,omitempty"`
}

// LedgerBucketAmountPlainText represents bucket amounts in plain text format
type LedgerBucketAmountPlainText struct {
	Bucket     *ledger.LedgerClassBucketType `json:"bucket,omitempty"`
	AppliedAmt string                        `json:"applied_amt,omitempty"`
	BalanceAmt string                        `json:"balance_amt,omitempty"`
}

// QueryLedgerEntryResponsePlainText represents the response for ledger entries query in plain text format
type QueryLedgerEntryResponsePlainText struct {
	Entries []*LedgerEntryPlainText `json:"entries,omitempty"`
}
