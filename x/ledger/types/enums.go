package types

import (
	"encoding/json"
	"fmt"
	"strings"
)

// UnmarshalJSON implements json.Unmarshaler for DayCount
func (d *DayCount) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		// Convert string to enum value
		str = strings.ToUpper(str)
		if value, exists := DayCount_value[str]; exists {
			*d = DayCount(value)
			return nil
		}
		// Try without DAY_COUNT_ prefix
		if value, exists := DayCount_value["DAY_COUNT_"+str]; exists {
			*d = DayCount(value)
			return nil
		}
		return fmt.Errorf("unknown DayCount string value: %s", str)
	}

	// Try to unmarshal as integer
	var num int32
	if err := json.Unmarshal(data, &num); err == nil {
		*d = DayCount(num)
		return nil
	}

	return fmt.Errorf("DayCount must be a string or integer, got: %s", string(data))
}

// UnmarshalJSON implements json.Unmarshaler for InterestAccrual
func (i *InterestAccrual) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		// Convert string to enum value
		str = strings.ToUpper(str)
		if value, exists := InterestAccrual_value[str]; exists {
			*i = InterestAccrual(value)
			return nil
		}
		// Try without INTEREST_ACCRUAL_ prefix
		if value, exists := InterestAccrual_value["INTEREST_ACCRUAL_"+str]; exists {
			*i = InterestAccrual(value)
			return nil
		}
		return fmt.Errorf("unknown InterestAccrual string value: %s", str)
	}

	// Try to unmarshal as integer
	var num int32
	if err := json.Unmarshal(data, &num); err == nil {
		*i = InterestAccrual(num)
		return nil
	}

	return fmt.Errorf("InterestAccrual must be a string or integer, got: %s", string(data))
}

// UnmarshalJSON implements json.Unmarshaler for PaymentFrequency
func (p *PaymentFrequency) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		// Convert string to enum value
		str = strings.ToUpper(str)
		if value, exists := PaymentFrequency_value[str]; exists {
			*p = PaymentFrequency(value)
			return nil
		}
		// Try without LEDGER_PAYMENT_FREQUENCY_ prefix
		if value, exists := PaymentFrequency_value["LEDGER_PAYMENT_FREQUENCY_"+str]; exists {
			*p = PaymentFrequency(value)
			return nil
		}
		return fmt.Errorf("unknown PaymentFrequency string value: %s", str)
	}

	// Try to unmarshal as integer
	var num int32
	if err := json.Unmarshal(data, &num); err == nil {
		*p = PaymentFrequency(num)
		return nil
	}

	return fmt.Errorf("PaymentFrequency must be a string or integer, got: %s", string(data))
}
