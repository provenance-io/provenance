package hold

import (
	"errors"
	"fmt"
)

func DefaultGenesisState() *GenesisState {
	return &GenesisState{}
}

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
