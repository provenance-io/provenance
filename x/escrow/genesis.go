package escrow

import (
	"errors"
	"fmt"
)

func DefaultGenesisState() *GenesisState {
	return &GenesisState{}
}

func (g GenesisState) Validate() error {
	var errs []error
	for i, ae := range g.Escrows {
		if err := ae.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("invalid escrows[%d]: %w", i, err))
		}
	}
	return errors.Join(errs...)
}
