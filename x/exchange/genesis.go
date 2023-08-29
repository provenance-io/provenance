package exchange

import (
	"errors"
	"fmt"
)

func NewGenesisState(params *Params, markets []Market, orders []Order) *GenesisState {
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
			errs = append(errs, fmt.Errorf("invalid params: %w", err))
		}
	}

	marketIDs := make(map[uint32]int)
	for i, market := range g.Markets {
		if market.MarketId == 0 {
			errs = append(errs, fmt.Errorf("invalid market[%d]: market id cannot be zero", i))
			continue
		}

		j, seen := marketIDs[market.MarketId]
		if seen {
			errs = append(errs, fmt.Errorf("invalid market[%d]: duplicate market id %d seen at [%d]", i, market.MarketId, j))
			continue
		}
		marketIDs[market.MarketId] = i

		if err := market.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("invalid market[%d]: %w", i, err))
		}
	}

	orderIDs := make(map[uint64]int)
	for i, order := range g.Orders {
		if order.OrderId != 0 {
			j, seen := orderIDs[order.OrderId]
			if seen {
				errs = append(errs, fmt.Errorf("invalid order[%d]: duplicate order id %d seen at [%d]", i, order.OrderId, j))
				continue
			}
			orderIDs[order.OrderId] = i
		}

		if err := order.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("invalid order[%d]: %w", i, err))
			continue
		}

		_, knownMarket := marketIDs[order.GetMarketID()]
		if !knownMarket {
			errs = append(errs, fmt.Errorf("invalid order[%d]: unknown market id %d", i, order.GetMarketID()))
		}
	}

	return errors.Join(errs...)
}
