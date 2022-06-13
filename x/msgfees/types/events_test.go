package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/assert"
)

func TestNewEventMsgs(t *testing.T) {
	totalCalls := make(map[string]uint64)
	totalFees := make(map[string]sdk.Coins)

	events := NewEventMsgs(totalCalls, totalFees)
	assert.Equal(t, 0, len(events.MsgFees))

	totalCalls["msgfee_typeurl_z"] = 612
	totalFees["msgfee_typeurl_z"] = sdk.NewCoins(sdk.NewCoin("jackthecat", sdk.NewInt(1000)))
	totalCalls["msgfee_typeurl_a"] = 406
	totalFees["msgfee_typeurl_a"] = sdk.NewCoins(sdk.NewCoin("jackthecat", sdk.NewInt(5000)))

	events = NewEventMsgs(totalCalls, totalFees)
	assert.Equal(t, 2, len(events.MsgFees))
	assert.Equal(t, "406", events.MsgFees[0].Count)
	assert.Equal(t, "msgfee_typeurl_a", events.MsgFees[0].MsgType)
	assert.Equal(t, "5000jackthecat", events.MsgFees[0].Total)
	assert.Equal(t, "612", events.MsgFees[1].Count)
	assert.Equal(t, "msgfee_typeurl_z", events.MsgFees[1].MsgType)
	assert.Equal(t, "1000jackthecat", events.MsgFees[1].Total)

	totalFees["not_in_calls_map"] = sdk.NewCoins(sdk.NewCoin("jackthecat", sdk.NewInt(1000)))
	events = NewEventMsgs(totalCalls, totalFees)
	assert.Equal(t, 2, len(events.MsgFees))

	totalCalls["not_in_total_map"] = 1
	events = NewEventMsgs(totalCalls, totalFees)
	assert.Equal(t, 2, len(events.MsgFees))

}
