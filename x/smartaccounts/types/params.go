package types

import (
	"encoding/json"
	"fmt"
)

// DefaultParams returns default module parameters.
func DefaultParams() Params {
	return Params{
		// enabled by default
		Enabled:              true,
		MaxCredentialAllowed: 10, // Set default max credentials per account
	}
}

// Stringer method for Params.
func (p Params) String() string {
	bz, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return string(bz)
}

// Validate does the sanity check on the params.
func (p Params) Validate() error {
	if p.MaxCredentialAllowed == 0 {
		return fmt.Errorf("max credential allowed must be positive")
	}

	// p.Enabled will always be set so not sure what to validate here. So no-op for now.
	return nil
}

// NewParams creates a new parameter object
func NewParams(
	enable bool,
) Params {
	return Params{
		Enabled: enable,
	}
}
