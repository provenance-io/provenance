package types

import (
	"encoding/json"
	"fmt"
	"strings"
)

// MustGetMarkerStatus turns the string into a MarkerStatus typed value ... panics if invalid.
func MustGetMarkerStatus(str string) MarkerStatus {
	s, err := MarkerStatusFromString(str)
	if err != nil {
		panic(err)
	}
	return s
}

// MarkerStatusFromString returns a MarkerStatus from a string. It returns an error
// if the string is invalid.
func MarkerStatusFromString(str string) (MarkerStatus, error) {
	switch strings.ToLower(str) {
	case "undefined":
		return StatusUndefined, nil
	case "proposed":
		return StatusProposed, nil
	case "finalized":
		return StatusFinalized, nil
	case "active":
		return StatusActive, nil
	case "cancelled":
		return StatusCancelled, nil
	case "destroyed":
		return StatusDestroyed, nil

	default:
		if val, ok := MarkerStatus_value[str]; ok {
			return MarkerStatus(val), nil
		}
	}

	return MarkerStatus(0xff), fmt.Errorf("'%s' is not a valid marker status", str)
}

// ValidMarkerStatus returns true if the marker status is valid and false otherwise.
func ValidMarkerStatus(markerStatus MarkerStatus) bool {
	_, ok := MarkerStatus_name[int32(markerStatus)]
	return ok && markerStatus != StatusUndefined
}

// Marshal needed for protobuf compatibility.
func (rt MarkerStatus) Marshal() ([]byte, error) {
	return []byte{byte(rt)}, nil
}

// Unmarshal needed for protobuf compatibility.
func (rt *MarkerStatus) Unmarshal(data []byte) error {
	*rt = MarkerStatus(data[0])
	return nil
}

// MarshalJSON using string.
func (rt MarkerStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(MarkerStatus_name[int32(rt)])
}

// UnmarshalJSON decodes from JSON string version of this status
func (rt *MarkerStatus) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	bz2, err := MarkerStatusFromString(s)
	if err != nil {
		return err
	}

	*rt = bz2
	return nil
}

// String implements the Stringer interface.
func (rt MarkerStatus) String() string {
	switch rt {
	case StatusUndefined:
		return "undefined"
	case StatusProposed:
		return "proposed"
	case StatusFinalized:
		return "finalized"
	case StatusActive:
		return "active"
	case StatusCancelled:
		return "cancelled"
	case StatusDestroyed:
		return "destroyed"

	default:
		return ""
	}
}

// Format implements the fmt.Formatter interface.
func (rt MarkerStatus) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(rt.String()))
	default:
		s.Write([]byte(fmt.Sprintf("%v", byte(rt))))
	}
}

// IsOneOf checks to see if this MarkerStatus is equal to one of the provided statuses.
func (rt MarkerStatus) IsOneOf(statuses ...MarkerStatus) bool {
	for _, s := range statuses {
		if rt == s {
			return true
		}
	}
	return false
}
