package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewEventMsgs(totalCalls map[string]uint64, totalFees map[string]sdk.Coin) *EventMsgFees {
	events := make([]EventMsgFee, len(totalCalls))
	i := 0
	for typeUrl, count := range totalCalls {
		total := totalFees[typeUrl]
		events[i] = EventMsgFee{
			MsgType: typeUrl,
			Count:   fmt.Sprintf("%v", count),
			Total:   total.String(),
		}
		i++
	}

	return &EventMsgFees{
		MsgFees: events,
	}
}
