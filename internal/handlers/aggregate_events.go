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
func AggregateEvents(anteEvents []abci.Event, resultEvents []abci.Event) ([]abci.Event, []abci.Event, error) {
	if len(resultEvents) == 0 { // tx failed...fix fee event to have the exact fee charged
		var err error
		var txFee sdk.Coins
		var feeIndex int
		var feeFound, spenderFound bool
		for i, event := range anteEvents {
			if !feeFound && event.Type == sdk.EventTypeTx && string(event.Attributes[0].Key) == sdk.AttributeKeyFee {
				feeFound = true
				feeIndex = i
			}
			// first spent coin event is the coin sent to fee module for tx
			if !spenderFound && event.Type == banktypes.EventTypeCoinSpent && string(event.Attributes[0].Key) == banktypes.AttributeKeySpender && len(anteEvents) >= i+3 {
				txFee, err = sdk.ParseCoinsNormalized(string(event.Attributes[1].Value))
				if err != nil {
					return nil, nil, err
				}
				spenderFound = true
			}
		}
		if feeFound && spenderFound {
			anteEvents[feeIndex].Attributes[0].Value = []byte(txFee.String())
		}
	}

	return anteEvents, resultEvents, nil
}
