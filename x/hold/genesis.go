package escrow

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
	for i, ae := range g.Escrows {
		if ae == nil {
			errs = append(errs, fmt.Errorf("invalid escrows[%d]: cannot be nil", i))
			continue
		}
		if err := ae.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("invalid escrows[%d]: %w", i, err))
			continue
		}
		j, seen := addrs[ae.Address]
		if seen {
			errs = append(errs, fmt.Errorf("invalid escrows[%d]: duplicate address also at index %d", i, j))
		} else {
			addrs[ae.Address] = i
		}
	}
	return errors.Join(errs...)
}
