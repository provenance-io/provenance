package exchange

import (
	"errors"
	"fmt"
)

func NewGenesisState(params *Params, markets []*Market, orders []*Order) *GenesisState {
	return &GenesisState{
		Params:  params,
		Markets: markets,
		Orders:  orders,
	}
}

func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultParams(), nil, nil)
}

func (g GenesisState) Validate() error {
	var errs []error

	if g.Params != nil {
		if err := g.Params.Validate(); err != nil {
			errs = append(errs, err)
		}
	}

	for i, market := range g.Markets {
		if err := market.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("invalid market[%d]: %w", i, err))
		}
	}

	for i, order := range g.Orders {
		if err := order.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("invalid order[%d]: %w", i, err))
		}
	}

	return errors.Join(errs...)
}
