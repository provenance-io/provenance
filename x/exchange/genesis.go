package exchange

import (
	"errors"
	"fmt"
)

// DefaultGenesisState returns the default genesis state for the exchange module.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{Params: DefaultParams()}
}

func (g GenesisState) Validate() error {
	var errs []error

	if g.Params != nil {
		if err := g.Params.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("invalid params: %w", err))
		}
	}

	marketIDs := make(map[uint32]int, len(g.Markets))
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

	maxOrderID := uint64(0)
	orderIDs := make(map[uint64]int, len(g.Orders))
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

		if order.OrderId > maxOrderID {
			maxOrderID = order.OrderId
		}
	}

	if g.LastOrderId < maxOrderID {
		errs = append(errs, fmt.Errorf("last order id %d is less than the largest id in the provided orders %d",
			g.LastOrderId, maxOrderID))
	}

	// No validation to do on LastMarketId.

	return errors.Join(errs...)
}
