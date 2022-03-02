package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewEventMsgs(totalCalls map[string]uint64, totalFees map[string]sdk.Coin) *EventMsgFees {
	events := make([]EventMsgFee, len(totalCalls))
	i := 0
	for typeURL, count := range totalCalls {
		total := totalFees[typeURL]
		events[i] = EventMsgFee{
			MsgType: typeURL,
			Count:   fmt.Sprintf("%v", count),
			Total:   total.String(),
		}
		i++
	}

	return &EventMsgFees{
		MsgFees: events,
	}
}
