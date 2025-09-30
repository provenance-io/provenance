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

	var holdAddrs []string
	holdAmounts := make(map[string]sdk.Coins)
	recordHold := func(addr string, amount sdk.Coins) {
		if !amount.IsZero() {
			if _, known := holdAmounts[addr]; !known {
				holdAddrs = append(holdAddrs, addr)
				holdAmounts[addr] = nil
			}
			holdAmounts[addr] = holdAmounts[addr].Add(amount...)
		}
	}

	var maxOrderID uint64
	for i, order := range genState.Orders {
		if err := k.setOrderInStore(store, order); err != nil {
			panic(fmt.Errorf("failed to store Orders[%d]: %w", i, err))
		}
		recordHold(order.GetOwner(), order.GetHoldAmount())
		if order.OrderId > maxOrderID {
			maxOrderID = order.OrderId
		}
	}

	if genState.LastOrderId < maxOrderID {
		panic(fmt.Errorf("last order id %d is less than largest order id %d", genState.LastOrderId, maxOrderID))
	}
	setLastOrderID(store, genState.LastOrderId)

	for i, com := range genState.Commitments {
		addr, err := sdk.AccAddressFromBech32(com.Account)
		if err != nil {
			panic(fmt.Errorf("failed to convert commitments[%d].Account=%q to AccAddress: %w", i, com.Account, err))
		}
		addCommitmentAmount(store, com.MarketId, addr, com.Amount)
		recordHold(com.Account, com.Amount)
	}

	for i := range genState.Payments {
		payment := &genState.Payments[i]
		err := k.createPaymentInStore(store, payment)
		if err != nil {
			panic(fmt.Errorf("failed to store Payments[%d]: %w", i, err))
		}
		recordHold(payment.Source, payment.SourceAmount)
	}

	// Make sure all the needed funds have holds on them. These should have been placed during initialization of the hold module.
	for _, addr := range holdAddrs {
		for _, reqAmt := range holdAmounts[addr] {
			holdAmt, err := k.holdKeeper.GetHoldCoin(ctx, sdk.MustAccAddressFromBech32(addr), reqAmt.Denom)
			if err != nil {
				panic(fmt.Errorf("failed to look up amount of %q on hold for %s: %w", reqAmt.Denom, addr, err))
			}
			if holdAmt.Amount.LT(reqAmt.Amount) {
				panic(fmt.Errorf("account %s should have at least %q on hold (due to the exchange module), but only has %q", addr, reqAmt, holdAmt))
			}
		}
	}
}

// ExportGenesis creates a genesis state from the current state store.
func (k Keeper) ExportGenesis(ctx sdk.Context) *exchange.GenesisState {
	store := k.getStore(ctx)
	genState := &exchange.GenesisState{
		Params:       k.GetParams(ctx),
		LastMarketId: getLastAutoMarketID(store),
		LastOrderId:  getLastOrderID(store),
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

	k.IterateCommitments(ctx, func(commitment exchange.Commitment) bool {
		genState.Commitments = append(genState.Commitments, commitment)
		return false
	})

	k.IteratePayments(ctx, func(payment *exchange.Payment) bool {
		genState.Payments = append(genState.Payments, *payment)
		return false
	})

	return genState
}
