package types

import (
	"testing"
	time "time"

	"github.com/stretchr/testify/assert"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
)

func TestNewTrigger(t *testing.T) {
	authorities := []string{"addr1", "addr2"}
	event := &BlockHeightEvent{}
	msgs := []sdk.Msg{}
	id := uint64(1)

	request := MustNewCreateTriggerRequest(authorities, event, msgs)

	trigger := NewTrigger(id, authorities[0], request.GetEvent(), request.GetActions())
	assert.Equal(t, int(id), int(trigger.Id), "should have correct id for NewTrigger")
	assert.Equal(t, authorities[0], trigger.Owner, "should have correct owner for NewTrigger")
	assert.Equal(t, request.GetEvent(), trigger.Event, "should have correct event for NewTrigger")
	assert.Equal(t, request.GetActions(), trigger.Actions, "should have correct actions for NewTrigger")
}

func TestNewQueuedTrigger(t *testing.T) {
	authorities := []string{"addr"}
	event := &BlockHeightEvent{}
	msgs := []sdk.Msg{}
	id := uint64(1)

	request := MustNewCreateTriggerRequest(authorities, event, msgs)

	trigger := NewTrigger(id, authorities[0], request.GetEvent(), request.GetActions())
	queuedTrigger := NewQueuedTrigger(trigger, time.Time{}, uint64(1))

	assert.Equal(t, trigger, queuedTrigger.Trigger, "should have correct trigger for NewQueuedTrigger")
	assert.Equal(t, time.Time{}, queuedTrigger.Time, "should have correct time for NewQueuedTrigger")
	assert.Equal(t, int(1), int(queuedTrigger.BlockHeight), "should have correct height for NewQueuedTrigger")
}

func TestTransactionEventMatches(t *testing.T) {
	tests := []struct {
		name        string
		event       TransactionEvent
		event2      abci.Event
		shouldMatch bool
	}{
		{
			name:        "valid - two exact events match",
			event:       TransactionEvent{Name: "name", Attributes: []Attribute{{Name: "attr1", Value: "value1"}, {Name: "attr2", Value: "value2"}}},
			event2:      abci.Event{Type: "name", Attributes: []abci.EventAttribute{{Key: "attr1", Value: "value1"}, {Key: "attr2", Value: "value2"}}},
			shouldMatch: true,
		},
		{
			name:        "valid - only specified attributes need to match match",
			event:       TransactionEvent{Name: "name", Attributes: []Attribute{{Name: "attr1", Value: "value1"}}},
			event2:      abci.Event{Type: "name", Attributes: []abci.EventAttribute{{Key: "attr1", Value: "value1"}, {Key: "attr2", Value: "value2"}}},
			shouldMatch: true,
		},
		{
			name:        "valid - no attributes",
			event:       TransactionEvent{Name: "name", Attributes: []Attribute{}},
			event2:      abci.Event{Type: "name", Attributes: []abci.EventAttribute{{Key: "attr1", Value: "value1"}, {Key: "attr2", Value: "value2"}}},
			shouldMatch: true,
		},
		{
			name:        "invalid - event name doesn't match",
			event:       TransactionEvent{Name: "invalid", Attributes: []Attribute{{Name: "attr1", Value: "value1"}, {Name: "attr2", Value: "value2"}}},
			event2:      abci.Event{Type: "name", Attributes: []abci.EventAttribute{{Key: "attr1", Value: "value1"}, {Key: "attr2", Value: "value2"}}},
			shouldMatch: false,
		},
		{
			name:        "invalid - missing attribute",
			event:       TransactionEvent{Name: "name", Attributes: []Attribute{{Name: "attr1", Value: "value1"}, {Name: "attr2", Value: "value2"}}},
			event2:      abci.Event{Type: "name", Attributes: []abci.EventAttribute{{Key: "attr1", Value: "value1"}, {Key: "attr3", Value: "value3"}}},
			shouldMatch: false,
		},
		{
			name:        "invalid - attribute value doesn't match",
			event:       TransactionEvent{Name: "name", Attributes: []Attribute{{Name: "attr1", Value: "value1"}, {Name: "attr2", Value: "value2"}}},
			event2:      abci.Event{Type: "name", Attributes: []abci.EventAttribute{{Key: "attr1", Value: "value3"}, {Key: "attr2", Value: "value2"}}},
			shouldMatch: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.shouldMatch, tc.event.Matches(tc.event2), "should have correct output for Matches")
		})
	}
}

func TestAttributeMatches(t *testing.T) {
	tests := []struct {
		name        string
		attr1       Attribute
		attr2       abci.EventAttribute
		shouldMatch bool
	}{
		{
			name:        "valid - two exact attributes are equal",
			attr1:       Attribute{Name: "attr", Value: "value"},
			attr2:       abci.EventAttribute{Key: "attr", Value: "value"},
			shouldMatch: true,
		},
		{
			name:        "valid - attribute matches wildcard",
			attr1:       Attribute{Name: "attr", Value: ""},
			attr2:       abci.EventAttribute{Key: "attr", Value: "value"},
			shouldMatch: true,
		},
		{
			name:        "invalid - names don't match",
			attr1:       Attribute{Name: "attr", Value: "value"},
			attr2:       abci.EventAttribute{Key: "blah", Value: "value"},
			shouldMatch: false,
		},
		{
			name:        "invalid - values don't match",
			attr1:       Attribute{Name: "attr", Value: "value"},
			attr2:       abci.EventAttribute{Key: "attr", Value: "blah"},
			shouldMatch: false,
		},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.shouldMatch, tc.attr1.Matches(tc.attr2), "should have correct output for Matches")
		t.Run(tc.name, func(t *testing.T) {})
	}
}

func TestTransactionEventGetEventPrefix(t *testing.T) {
	event := TransactionEvent{Name: "customName"}
	assert.Equal(t, "customName", event.GetEventPrefix(), "should get correct prefix for GetEventPrefix")
}

func TestTransactionEventGetEventOrder(t *testing.T) {
	event := TransactionEvent{Name: "customName"}
	assert.Equal(t, 0, int(event.GetEventOrder()), "should get correct event order")
}

func TestTransactionEventValidate(t *testing.T) {
	tests := []struct {
		name  string
		event TransactionEvent
		err   string
	}{
		{
			name:  "valid - transaction event no attributes",
			event: TransactionEvent{Name: "event", Attributes: []Attribute{}},
			err:   "",
		},
		{
			name:  "valid - transaction event with attributes",
			event: TransactionEvent{Name: "event", Attributes: []Attribute{{Name: "attr", Value: "value"}}},
			err:   "",
		},
		{
			name:  "invalid - empty name",
			event: TransactionEvent{Name: "", Attributes: []Attribute{{Name: "attr", Value: "value"}}},
			err:   "empty event name",
		},
		{
			name:  "invalid - empty attribute name",
			event: TransactionEvent{Name: "event", Attributes: []Attribute{{Name: "", Value: "value"}}},
			err:   "empty attribute name",
		},
		{
			name:  "invalid - empty attribute name with multiple attributes",
			event: TransactionEvent{Name: "event", Attributes: []Attribute{{Name: "", Value: "value"}, {Name: "attr", Value: "value2"}}},
			err:   "empty attribute name",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := tc.event.Validate()
			if len(tc.err) > 0 {
				assert.EqualError(t, res, tc.err, "should have correct error for Validate")
			} else {
				assert.NoError(t, res, "should have no error for successful Validate")
			}
		})
	}
}

func TestTransactionEventValidateContext(t *testing.T) {
	ctx := sdk.NewContext(nil, cmtproto.Header{Time: time.Now().UTC()}, false, nil)
	ctx = ctx.WithBlockHeight(100)

	tests := []struct {
		name  string
		event TransactionEvent
		err   string
	}{
		{
			name:  "valid - transaction event should always succeed",
			event: TransactionEvent{Name: "event", Attributes: []Attribute{}},
			err:   "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := tc.event.ValidateContext(ctx)
			if len(tc.err) > 0 {
				assert.EqualError(t, res, tc.err, "should have correct error for ValidateContext")
			} else {
				assert.NoError(t, res, "should have no error for successful ValidateContext")
			}
		})
	}
}

func TestBlockHeightEventValidateContext(t *testing.T) {
	ctx := sdk.NewContext(nil, cmtproto.Header{Time: time.Now().UTC()}, false, nil)
	ctx = ctx.WithBlockHeight(100)

	tests := []struct {
		name  string
		event BlockHeightEvent
		err   string
	}{
		{
			name:  "valid - block height event should be valid for future heights",
			event: BlockHeightEvent{BlockHeight: 101},
			err:   "",
		},
		{
			name:  "valid - block height event should be invalid for current height",
			event: BlockHeightEvent{BlockHeight: 100},
			err:   ErrInvalidBlockHeight.Error(),
		},
		{
			name:  "valid - block height event should be invalid for past height",
			event: BlockHeightEvent{BlockHeight: 99},
			err:   ErrInvalidBlockHeight.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := tc.event.ValidateContext(ctx)
			if len(tc.err) > 0 {
				assert.EqualError(t, res, tc.err, "should have correct error for ValidateContext")
			} else {
				assert.NoError(t, res, "should have no error for succesful ValidateContext")
			}
		})
	}
}

func TestBlockTimeEventValidateContext(t *testing.T) {
	now := time.Now().UTC()
	ctx := sdk.NewContext(nil, cmtproto.Header{Time: now}, false, nil)
	ctx = ctx.WithBlockHeight(100)

	tests := []struct {
		name  string
		event BlockTimeEvent
		err   string
	}{
		{
			name:  "valid - block height event should be valid for future time",
			event: BlockTimeEvent{now.Add(time.Hour)},
			err:   "",
		},
		{
			name:  "invalid - block height event should be invalid for current height",
			event: BlockTimeEvent{now},
			err:   ErrInvalidBlockTime.Error(),
		},
		{
			name:  "invalid - block height event should be invalid for past height",
			event: BlockTimeEvent{now.Add(-time.Hour)},
			err:   ErrInvalidBlockTime.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := tc.event.ValidateContext(ctx)
			if len(tc.err) > 0 {
				assert.EqualError(t, res, tc.err, "should have correct error for ValidateContext")
			} else {
				assert.NoError(t, res, "should have no error for successful ValidateContext")
			}
		})
	}
}

func TestBlockHeightEventGetEventPrefix(t *testing.T) {
	event := BlockHeightEvent{}
	assert.Equal(t, BlockHeightPrefix, event.GetEventPrefix(), "should have correct prefix for GetEventPrefix")
}

func TestBlockHeightEventGetEventOrder(t *testing.T) {
	event := BlockHeightEvent{BlockHeight: 77}
	assert.Equal(t, int(77), int(event.GetEventOrder()), "should have correct event order")
}

func TestBlockHeightEventValidate(t *testing.T) {
	event := BlockHeightEvent{}
	assert.Nil(t, event.Validate(), "should always have successful validate")
}

func TestBlockTimeEventGetEventPrefix(t *testing.T) {
	event := BlockTimeEvent{}
	assert.Equal(t, BlockTimePrefix, event.GetEventPrefix(), "should have correct prefix for GetEventPrefix")
}

func TestBlockTimeEventGetEventOrder(t *testing.T) {
	now := time.Now().UTC()
	event := BlockTimeEvent{Time: now}
	assert.Equal(t, int(now.UnixNano()), int(event.GetEventOrder()), "should have correct order")
}

func TestBlockTimeEventValidate(t *testing.T) {
	event := BlockTimeEvent{}
	assert.Nil(t, event.Validate(), "should always have successful validate")
}

func TestTriggerUnpackInterfaces(t *testing.T) {
	cdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

	tests := []struct {
		name      string
		authority string
		event     TriggerEventI
		msgs      []sdk.Msg
	}{
		{
			name:      "valid - Unpack Trigger Interfaces",
			authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
			event:     &BlockHeightEvent{},
			msgs:      []sdk.Msg{&MsgDestroyTriggerRequest{Authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", Id: 1}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := MustNewCreateTriggerRequest([]string{tc.authority}, tc.event, tc.msgs)
			trigger := NewTrigger(uint64(1), tc.authority, msg.Event, msg.Actions)
			err := trigger.UnpackInterfaces(cdc)
			assert.NoError(t, err, "should not throw an error for UnpackInterfaces")
			assert.Equal(t, tc.event, trigger.Event.GetCachedValue(), "should have correct cached event after UnpackInterfaces")
			assert.Equal(t, tc.msgs[0], trigger.Actions[0].GetCachedValue(), "should have correct cached actions after UnpackInterfaces")
		})
	}
}

func TestQueuedTriggerUnpackInterfaces(t *testing.T) {
	cdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

	tests := []struct {
		name      string
		authority string
		event     TriggerEventI
		msgs      []sdk.Msg
	}{
		{
			name:      "valid - Unpack Trigger Interfaces",
			authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
			event:     &BlockHeightEvent{},
			msgs:      []sdk.Msg{&MsgDestroyTriggerRequest{Authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", Id: 1}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := MustNewCreateTriggerRequest([]string{tc.authority}, tc.event, tc.msgs)
			trigger := NewTrigger(uint64(1), tc.authority, msg.Event, msg.Actions)
			queuedTrigger := NewQueuedTrigger(trigger, time.Time{}, uint64(1))
			err := queuedTrigger.UnpackInterfaces(cdc)
			assert.NoError(t, err, "should have no error in successful UnpackInterfaces")
			assert.Equal(t, tc.event, queuedTrigger.Trigger.Event.GetCachedValue(), "should have correct cached event after UnpackInterfaces")
			assert.Equal(t, tc.msgs[0], queuedTrigger.Trigger.Actions[0].GetCachedValue(), "should have correct cached actions after UnpackInterfaces")
		})
	}
}

func TestTriggerGetTriggerEventI(t *testing.T) {
	cdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

	tests := []struct {
		name      string
		authority string
		event     TriggerEventI
		msgs      []sdk.Msg
		err       error
	}{
		{
			name:      "valid - GetTriggerEventI",
			authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
			event:     &BlockHeightEvent{},
			msgs:      []sdk.Msg{&MsgDestroyTriggerRequest{Authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", Id: 1}},
			err:       nil,
		},
		{
			name:      "invalid - Returns error when interface is nil",
			authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
			event:     nil,
			msgs:      []sdk.Msg{&MsgDestroyTriggerRequest{Authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", Id: 1}},
			err:       ErrNoTriggerEvent.Wrap("failed to get event"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			event := tc.event
			if event == nil {
				event = &BlockHeightEvent{}
			}
			msg := MustNewCreateTriggerRequest([]string{tc.authority}, event, tc.msgs)
			if tc.event == nil {
				msg.Event = nil
			}
			trigger := NewTrigger(uint64(1), tc.authority, msg.Event, msg.Actions)
			err := trigger.UnpackInterfaces(cdc)
			assert.NoError(t, err, "should have no error from UnpackInterfaces")
			triggerEvent, err := msg.GetTriggerEventI()
			if tc.err == nil {
				assert.NoError(t, err, "should not throw error in successful GetTriggerEvent")
				assert.Equal(t, tc.event, triggerEvent, "should have expected event for GetTriggerEventI")
			} else {
				assert.Error(t, tc.err, err, "should have expected error from failing GetTriggerEventI")
			}

		})
	}
}
