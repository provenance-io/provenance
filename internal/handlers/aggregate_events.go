package handlers

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/internal/antewrapper"
)

// AggregateEvents is called in the baseapp after a transaction has completed
// This is used to modify the events that are emitted for a transaction.
// anteEvents will be populated on failure and success
// resultEvents will only be populated on success
func AggregateEvents(anteEvents []abci.Event, resultEvents []abci.Event) ([]abci.Event, []abci.Event) {
	if len(resultEvents) == 0 { // tx failed...fix fee event to have the exact fee charged
		var txFee []byte
		var feeEventIndex, feeAttributeIndex int
		var feeFound, minFeeFound bool
		for i, event := range anteEvents {
			if event.Type == sdk.EventTypeTx {
				for j, attr := range event.Attributes {
					if !feeFound && string(attr.Key) == sdk.AttributeKeyFee {
						feeEventIndex = i
						feeAttributeIndex = j
						feeFound = true
					}
					if !minFeeFound && string(attr.Key) == antewrapper.AttributeKeyMinFeeCharged {
						txFee = attr.Value
						minFeeFound = true
					}
				}
			}
			if minFeeFound && feeFound {
				break
			}
		}
		if feeFound && minFeeFound {
			anteEvents[feeEventIndex].Attributes[feeAttributeIndex].Value = txFee
		}
	}

	return anteEvents, resultEvents
}
