package handlers

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

// AggregateEvents is called in the baseapp after a transaction has completed
// This is used to modify the events that are emitted for a transaction.
// anteEvents will be populated on failure and success
// resultEvents will only be populated on success
func AggregateEvents(anteEvents []abci.Event, resultEvents []abci.Event) ([]abci.Event, []abci.Event) {
	if len(resultEvents) == 0 { // tx failed...fix fee event to have the exact fee charged
		var txFee []byte
		var feeIndex int
		var feeFound, spenderFound bool
		for i, event := range anteEvents {
			if !feeFound && event.Type == sdk.EventTypeTx && string(event.Attributes[0].Key) == sdk.AttributeKeyFee {
				feeFound = true
				feeIndex = i
			}
			// first spent coin event is the coin sent to fee module for tx
			if !spenderFound && event.Type == banktypes.EventTypeCoinSpent && string(event.Attributes[0].Key) == banktypes.AttributeKeySpender {
				txFee = event.Attributes[1].Value
				spenderFound = true
			}
			if spenderFound && feeFound {
				break
			}
		}
		if feeFound && spenderFound {
			anteEvents[feeIndex].Attributes[0].Value = txFee
		}
	}

	return anteEvents, resultEvents
}
