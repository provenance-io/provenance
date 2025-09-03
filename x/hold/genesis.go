package hold

import (
	"errors"
	"fmt"
)

// DefaultGenesisState returns the default genesis state for the hold module.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{}
}

// Validate validates the genesis state of the hold module.
func (g GenesisState) Validate() error {
	addrs := make(map[string]int)
	var errs []error
	for i, ah := range g.Holds {
		if ah == nil {
			errs = append(errs, fmt.Errorf("invalid holds[%d]: cannot be nil", i))
			continue
		}
		if err := ah.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("invalid holds[%d]: %w", i, err))
			continue
		}
		j, seen := addrs[ah.Address]
		if seen {
			errs = append(errs, fmt.Errorf("invalid holds[%d]: duplicate address also at index %d", i, j))
		} else {
			addrs[ah.Address] = i
		}
	}
	return errors.Join(errs...)
}
