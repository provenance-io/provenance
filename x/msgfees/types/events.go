package types

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// EventTypeAssessCustomMsgFee is the event that is emitted when assess custom fee is submitted as msg
	EventTypeAssessCustomMsgFee string = "assess_custom_msg_fee"

	// KeyAttributeAmount is the key for the custom additional amount of fee
	KeyAttributeAmount string = "amount"
	// KeyAttributeRecipient is the key for the optional recipient of the request, if empty the full fee amount is sent to fee module
	KeyAttributeRecipient string = "recipient"
	// KeyAttributeName is the key for the optional name for assess custom fee
	KeyAttributeName string = "name"
)

func NewEventMsgs(totalCalls map[string]uint64, totalFees map[string]sdk.Coins) *EventMsgFees {
	sortedKeys := sortAndReduce(totalCalls, totalFees)
	events := make([]EventMsgFee, len(sortedKeys))
	for i, compositeKey := range sortedKeys {
		msgType, recipient := SplitCompositeKey(compositeKey)

		events[i] = EventMsgFee{
			MsgType:   msgType,
			Count:     fmt.Sprintf("%v", totalCalls[compositeKey]),
			Total:     totalFees[compositeKey].String(),
			Recipient: recipient,
		}
	}

	return &EventMsgFees{
		MsgFees: events,
	}
}

// sortAndReduce returns a sorted list of keys that are contained in both totalCalls and totalFees
func sortAndReduce(totalCalls map[string]uint64, totalFees map[string]sdk.Coins) []string {
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
