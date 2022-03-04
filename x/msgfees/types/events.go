package types

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewEventMsgs(totalCalls map[string]uint64, totalFees map[string]sdk.Coin) *EventMsgFees {
	sortedKeys := sortAndReduce(totalCalls, totalFees)
	events := make([]EventMsgFee, len(sortedKeys))
	for i, typeURL := range sortedKeys {
		events[i] = EventMsgFee{
			MsgType: typeURL,
			Count:   fmt.Sprintf("%v", totalCalls[typeURL]),
			Total:   totalFees[typeURL].String(),
		}
	}

	return &EventMsgFees{
		MsgFees: events,
	}
}

// sortAndReduce returns a sorted list of keys that are contained in both totalCalls and totalFees
func sortAndReduce(totalCalls map[string]uint64, totalFees map[string]sdk.Coin) []string {
	keys := make([]string, 0, len(totalCalls))
	totalAdded := 0
	for k := range totalCalls {
		_, found := totalFees[k]
		if found {
			keys = append(keys, k)
			totalAdded++
		}
	}
	sort.Strings(keys[0:totalAdded])
	return keys[0:totalAdded]
}
