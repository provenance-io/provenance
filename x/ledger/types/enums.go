package types

import (
	"encoding/json"
	"fmt"
	"strings"
)

// enumUnmarshalJSON unmarshals an enum entry from either a JSON string or number.
// As a string, the name prefix is optional. I.e. "PAYMENT_FREQUENCY_DAILY" can be provided as just "DAILY".
// It is not an error to get the _UNSPECIFIED value.
func enumUnmarshalJSON(data []byte, values map[string]int32, names map[int32]string) (int32, error) {
	var namePrefix string
	// Try to unmarshal as string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		// Convert string to enum value
		str = strings.ToUpper(str)
		if value, exists := values[str]; exists {
			return value, nil
		}
		namePrefix = strings.TrimSuffix(names[0], "UNSPECIFIED") // E.g. DAY_COUNT_CONVENTION_
		// Try without enum prefix
		if value, exists := values[namePrefix+str]; exists {
			return value, nil
		}
		return 0, fmt.Errorf("unknown %s string value: %s", strings.ToLower(strings.TrimSuffix(namePrefix, "_")), str)
	}

	// Try to unmarshal as integer
	var num int32
	if err := json.Unmarshal(data, &num); err == nil {
		if _, exists := names[num]; exists {
			return num, nil
		}
		return 0, fmt.Errorf("unknown %s integer value: %d", strings.ToLower(strings.TrimSuffix(names[0], "_UNSPECIFIED")), num)
	}

	return 0, fmt.Errorf("%s must be a string or integer, got: %s", strings.ToLower(strings.TrimSuffix(names[0], "_UNSPECIFIED")), string(data))
}

// enumValidateExists returns an error if the provided value is not contained in the provided names map.
// It does NOT return an error on the zero (_UNSPECIFIED) value.
func enumValidateExists[E ~int32](value E, names map[int32]string) error {
	if _, exists := names[int32(value)]; !exists {
		return fmt.Errorf("unknown %s enum value: %d", strings.ToLower(strings.TrimSuffix(names[0], "_UNSPECIFIED")), value)
	}
	return nil
}
