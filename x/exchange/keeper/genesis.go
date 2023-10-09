package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/exchange"
)

// InitGenesis writes the provided genesis state to the state store.
func (k Keeper) InitGenesis(ctx sdk.Context, genState *exchange.GenesisState) {
	if genState == nil {
		return
	}

	k.SetParams(ctx, genState.Params)

	store := k.getStore(ctx)
	for _, market := range genState.Markets {
		k.initMarket(ctx, store, market)
	}

	setLastAutoMarketID(store, genState.LastMarketId)

	var addrs []string
	amounts := make(map[string]sdk.Coins)

	for i, order := range genState.Orders {
		if err := k.setOrderInStore(store, order); err != nil {
			panic(fmt.Errorf("failed to store Orders[%d]: %w", i, err))
		}
		addr := order.GetOwner()
		if _, known := amounts[addr]; !known {
			addrs = append(addrs, addr)
			amounts[addr] = nil
		}
		amounts[addr] = amounts[addr].Add(order.GetHoldAmount()...)
	}

	// Make sure all the needed funds have holds on them. These should have been placed during initialization of the hold module.
	for _, addr := range addrs {
		for _, reqAmt := range amounts[addr] {
			holdAmt, err := k.holdKeeper.GetHoldCoin(ctx, sdk.MustAccAddressFromBech32(addr), reqAmt.Denom)
			if err != nil {
				panic(fmt.Errorf("failed to look up amount of %q on hold for %s", reqAmt.Denom, addr))
			}
			if holdAmt.Amount.LT(reqAmt.Amount) {
				panic(fmt.Errorf("account %s should have at least %q on hold (due to exchange orders), but only has %q", addr, reqAmt, holdAmt))
			}
		}
	}
}

// ExportGenesis creates a genesis state from the current state store.
func (k Keeper) ExportGenesis(ctx sdk.Context) *exchange.GenesisState {
	genState := &exchange.GenesisState{
		Params:       k.GetParams(ctx),
		LastMarketId: getLastAutoMarketID(k.getStore(ctx)),
	}

	k.IterateMarkets(ctx, func(market *exchange.Market) bool {
		genState.Markets = append(genState.Markets, *market)
		return false
	})

	err := k.IterateOrders(ctx, func(order *exchange.Order) bool {
		genState.Orders = append(genState.Orders, *order)
		return false
	})
	if err != nil {
		k.logErrorf(ctx, "error (ignored) while reading orders: %v", err)
	}

	return genState
}
