package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	ledger "github.com/provenance-io/provenance/x/ledger/types"
)

// parseEnum parses a string value into an enum using the generated *_value map.
// It normalizes input to be case-insensitive and accepts either - or _ as separators.
// The prefix is optional, e.g., "DAILY" or "PAYMENT_FREQUENCY_DAILY" both work.
func parseEnum[E ~int32](val, flagName, prefix string, valueMap map[string]int32, unspecified E) (E, error) {
	// Normalize input: uppercase and replace dashes with underscores
	normalized := strings.ToUpper(strings.ReplaceAll(val, "-", "_"))

	// Try with prefix first (e.g., "DAY_COUNT_CONVENTION_ACTUAL_365")
	if enumVal, exists := valueMap[normalized]; exists {
		return E(enumVal), nil
	}

	// Try with prefix prepended (e.g., "ACTUAL_365" -> "DAY_COUNT_CONVENTION_ACTUAL_365")
	if enumVal, exists := valueMap[prefix+normalized]; exists {
		return E(enumVal), nil
	}

	return unspecified, fmt.Errorf("invalid --%s: %s", flagName, val)
}

// Flag names used by ledger CLI commands.
const (
	FlagNextPmtDate           = "next-pmt-date"
	FlagNextPmtAmt            = "next-pmt-amt"
	FlagInterestRate          = "interest-rate"
	FlagMaturityDate          = "maturity-date"
	FlagDayCountConvention    = "day-count-convention"
	FlagInterestAccrualMethod = "interest-accrual-method"
	FlagPaymentFrequency      = "payment-frequency"
)

// AddFlagNextPmtDate adds the --next-pmt-date flag to the provided command.
func AddFlagNextPmtDate(cmd *cobra.Command) {
	cmd.Flags().String(FlagNextPmtDate, "", "Next payment date (YYYY-MM-DD)")
}

// ReadFlagNextPmtDate returns the value provided with the --next-pmt-date flag.
func ReadFlagNextPmtDate(flagSet *pflag.FlagSet) (string, error) {
	return flagSet.GetString(FlagNextPmtDate)
}

// AddFlagNextPmtAmt adds the --next-pmt-amt flag to the provided command.
func AddFlagNextPmtAmt(cmd *cobra.Command) {
	cmd.Flags().Int64(FlagNextPmtAmt, 0, "Next payment amount")
}

// ReadFlagNextPmtAmt returns the value provided with the --next-pmt-amt flag.
func ReadFlagNextPmtAmt(flagSet *pflag.FlagSet) (int64, error) {
	return flagSet.GetInt64(FlagNextPmtAmt)
}

// AddFlagInterestRate adds the --interest-rate flag to the provided command.
func AddFlagInterestRate(cmd *cobra.Command) {
	cmd.Flags().Int32(FlagInterestRate, 0, "Interest rate (10000000 = 10.000000%)")
}

// ReadFlagInterestRate returns the value provided with the --interest-rate flag.
func ReadFlagInterestRate(flagSet *pflag.FlagSet) (int32, error) {
	return flagSet.GetInt32(FlagInterestRate)
}

// AddFlagMaturityDate adds the --maturity-date flag to the provided command.
func AddFlagMaturityDate(cmd *cobra.Command) {
	cmd.Flags().String(FlagMaturityDate, "", "Maturity date (YYYY-MM-DD)")
}

// ReadFlagMaturityDate returns the value provided with the --maturity-date flag.
func ReadFlagMaturityDate(flagSet *pflag.FlagSet) (string, error) {
	return flagSet.GetString(FlagMaturityDate)
}

// AddFlagDayCountConvention adds the --day-count-convention flag to the provided command.
func AddFlagDayCountConvention(cmd *cobra.Command) {
	cmd.Flags().String(FlagDayCountConvention, "", "Day count convention (actual-365, actual-360, thirty-360, actual-actual, days-365, days-360)")
}

// ReadFlagDayCountConvention returns the parsed enum value provided with the --day-count-convention flag.
func ReadFlagDayCountConvention(flagSet *pflag.FlagSet) (ledger.DayCountConvention, error) {
	val, err := flagSet.GetString(FlagDayCountConvention)
	if err != nil {
		return ledger.DAY_COUNT_CONVENTION_UNSPECIFIED, err
	}
	if val == "" {
		return ledger.DAY_COUNT_CONVENTION_UNSPECIFIED, nil
	}

	return parseEnum(val, FlagDayCountConvention, "DAY_COUNT_CONVENTION_", ledger.DayCountConvention_value, ledger.DAY_COUNT_CONVENTION_UNSPECIFIED)
}

// AddFlagInterestAccrualMethod adds the --interest-accrual-method flag to the provided command.
func AddFlagInterestAccrualMethod(cmd *cobra.Command) {
	cmd.Flags().String(FlagInterestAccrualMethod, "", "Interest accrual method (simple, compound, daily, monthly, quarterly, annual, continuous)")
}

// ReadFlagInterestAccrualMethod returns the parsed enum value provided with the --interest-accrual-method flag.
func ReadFlagInterestAccrualMethod(flagSet *pflag.FlagSet) (ledger.InterestAccrualMethod, error) {
	val, err := flagSet.GetString(FlagInterestAccrualMethod)
	if err != nil {
		return ledger.INTEREST_ACCRUAL_METHOD_UNSPECIFIED, err
	}
	if val == "" {
		return ledger.INTEREST_ACCRUAL_METHOD_UNSPECIFIED, nil
	}

	// Map short forms to canonical tokens expected by the enum.
	switch strings.ToUpper(val) {
	case "SIMPLE":
		val = "SIMPLE_INTEREST"
	case "COMPOUND":
		val = "COMPOUND_INTEREST"
	case "DAILY":
		val = "DAILY_COMPOUNDING"
	case "MONTHLY":
		val = "MONTHLY_COMPOUNDING"
	case "QUARTERLY":
		val = "QUARTERLY_COMPOUNDING"
	case "ANNUAL", "ANNUALLY":
		val = "ANNUAL_COMPOUNDING"
	case "CONTINUOUS":
		val = "CONTINUOUS_COMPOUNDING"
	}

	return parseEnum(val, FlagInterestAccrualMethod, "INTEREST_ACCRUAL_METHOD_", ledger.InterestAccrualMethod_value, ledger.INTEREST_ACCRUAL_METHOD_UNSPECIFIED)
}

// AddFlagPaymentFrequency adds the --payment-frequency flag to the provided command.
func AddFlagPaymentFrequency(cmd *cobra.Command) {
	cmd.Flags().String(FlagPaymentFrequency, "", "Payment frequency (daily, weekly, monthly, quarterly, annually)")
}

// ReadFlagPaymentFrequency returns the parsed enum value provided with the --payment-frequency flag.
func ReadFlagPaymentFrequency(flagSet *pflag.FlagSet) (ledger.PaymentFrequency, error) {
	val, err := flagSet.GetString(FlagPaymentFrequency)
	if err != nil {
		return ledger.PAYMENT_FREQUENCY_UNSPECIFIED, err
	}
	if val == "" {
		return ledger.PAYMENT_FREQUENCY_UNSPECIFIED, nil
	}

	return parseEnum(val, FlagPaymentFrequency, "PAYMENT_FREQUENCY_", ledger.PaymentFrequency_value, ledger.PAYMENT_FREQUENCY_UNSPECIFIED)
}
