package types

import (
	"testing"
	time "time"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestNewTrigger(t *testing.T) {
	authority := "addr"
	event := &BlockHeightEvent{}
	msgs := []sdk.Msg{}
	id := uint64(1)

	request := NewCreateTriggerRequest(authority, event, msgs)

	trigger := NewTrigger(id, authority, request.GetEvent(), request.GetActions())
	assert.Equal(t, authority, trigger.Owner)
	assert.Equal(t, request.GetEvent(), trigger.Event)
	assert.Equal(t, request.GetActions(), trigger.Actions)
}

func TestNewQueuedTrigger(t *testing.T) {
	authority := "addr"
	event := &BlockHeightEvent{}
	msgs := []sdk.Msg{}
	id := uint64(1)

	request := NewCreateTriggerRequest(authority, event, msgs)

	trigger := NewTrigger(id, authority, request.GetEvent(), request.GetActions())
	queuedTrigger := NewQueuedTrigger(trigger, time.Time{}, uint64(1))

	assert.Equal(t, trigger, queuedTrigger.Trigger)
	assert.Equal(t, time.Time{}, queuedTrigger.Time)
	assert.Equal(t, uint64(1), queuedTrigger.BlockHeight)
}

func TestTransactionEventEquals(t *testing.T) {
	tests := []struct {
		name        string
		event       TransactionEvent
		event2      abci.Event
		shouldMatch bool
	}{
		{
			name:        "valid - two exact events match",
			event:       TransactionEvent{Name: "name", Attributes: []Attribute{{Name: "attr1", Value: "value1"}, {Name: "attr2", Value: "value2"}}},
			event2:      abci.Event{Type: "name", Attributes: []abci.EventAttribute{{Key: []byte("attr1"), Value: []byte("value1")}, {Key: []byte("attr2"), Value: []byte("value2")}}},
			shouldMatch: true,
		},
		{
			name:        "valid - only specified attributes need to match match",
			event:       TransactionEvent{Name: "name", Attributes: []Attribute{{Name: "attr1", Value: "value1"}}},
			event2:      abci.Event{Type: "name", Attributes: []abci.EventAttribute{{Key: []byte("attr1"), Value: []byte("value1")}, {Key: []byte("attr2"), Value: []byte("value2")}}},
			shouldMatch: true,
		},
		{
			name:        "valid - no attributes",
			event:       TransactionEvent{Name: "name", Attributes: []Attribute{}},
			event2:      abci.Event{Type: "name", Attributes: []abci.EventAttribute{{Key: []byte("attr1"), Value: []byte("value1")}, {Key: []byte("attr2"), Value: []byte("value2")}}},
			shouldMatch: true,
		},
		{
			name:        "invalid - event name doesn't match",
			event:       TransactionEvent{Name: "invalid", Attributes: []Attribute{{Name: "attr1", Value: "value1"}, {Name: "attr2", Value: "value2"}}},
			event2:      abci.Event{Type: "name", Attributes: []abci.EventAttribute{{Key: []byte("attr1"), Value: []byte("value1")}, {Key: []byte("attr2"), Value: []byte("value2")}}},
			shouldMatch: false,
		},
		{
			name:        "invalid - missing attribute",
			event:       TransactionEvent{Name: "name", Attributes: []Attribute{{Name: "attr1", Value: "value1"}, {Name: "attr2", Value: "value2"}}},
			event2:      abci.Event{Type: "name", Attributes: []abci.EventAttribute{{Key: []byte("attr1"), Value: []byte("value1")}, {Key: []byte("attr3"), Value: []byte("value3")}}},
			shouldMatch: false,
		},
		{
			name:        "invalid - attribute value doesn't match",
			event:       TransactionEvent{Name: "name", Attributes: []Attribute{{Name: "attr1", Value: "value1"}, {Name: "attr2", Value: "value2"}}},
			event2:      abci.Event{Type: "name", Attributes: []abci.EventAttribute{{Key: []byte("attr1"), Value: []byte("value3")}, {Key: []byte("attr2"), Value: []byte("value2")}}},
			shouldMatch: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.shouldMatch, tc.event.Equals(tc.event2))
		})
	}
}

func TestAttributeEquals(t *testing.T) {
	tests := []struct {
		name        string
		attr1       Attribute
		attr2       abci.EventAttribute
		shouldMatch bool
	}{
		{
			name:        "valid - two exact attributes are equal",
			attr1:       Attribute{Name: "attr", Value: "value"},
			attr2:       abci.EventAttribute{Key: []byte("attr"), Value: []byte("value")},
			shouldMatch: true,
		},
		{
			name:        "valid - attribute matches wildcard",
			attr1:       Attribute{Name: "attr", Value: ""},
			attr2:       abci.EventAttribute{Key: []byte("attr"), Value: []byte("value")},
			shouldMatch: true,
		},
		{
			name:        "invalid - names don't match",
			attr1:       Attribute{Name: "attr", Value: "value"},
			attr2:       abci.EventAttribute{Key: []byte("blah"), Value: []byte("value")},
			shouldMatch: false,
		},
		{
			name:        "invalid - values don't match",
			attr1:       Attribute{Name: "attr", Value: "value"},
			attr2:       abci.EventAttribute{Key: []byte("attr"), Value: []byte("blah")},
			shouldMatch: false,
		},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.shouldMatch, tc.attr1.Equals(tc.attr2))
		t.Run(tc.name, func(t *testing.T) {})
	}
}

func TestTransactionEventGetEventPrefix(t *testing.T) {
	event := TransactionEvent{Name: "customName"}
	assert.Equal(t, "customName", event.GetEventPrefix())
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
		tc.event.Validate()
		t.Run(tc.name, func(t *testing.T) {
			res := tc.event.Validate()
			if len(tc.err) > 0 {
				assert.Errorf(t, res, tc.err)
			} else {
				assert.NoError(t, res)
			}
		})
	}
}

func TestBlockHeightEventGetEventPrefix(t *testing.T) {
	event := BlockHeightEvent{}
	assert.Equal(t, BlockHeightPrefix, event.GetEventPrefix())
}

func TestBlockHeightEventValidate(t *testing.T) {
	event := BlockHeightEvent{}
	assert.Nil(t, event.Validate())
}

func TestBlockTimeEventGetEventPrefix(t *testing.T) {
	event := BlockTimeEvent{}
	assert.Equal(t, BlockTimePrefix, event.GetEventPrefix())
}

func TestBlockTimeEventValidate(t *testing.T) {
	event := BlockTimeEvent{}
	assert.Nil(t, event.Validate())
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
			msg := NewCreateTriggerRequest(tc.authority, tc.event, tc.msgs)
			trigger := NewTrigger(uint64(1), tc.authority, msg.Event, msg.Actions)
			err := trigger.UnpackInterfaces(cdc)
			assert.NoError(t, err)
			assert.Equal(t, tc.event, trigger.Event.GetCachedValue())
			assert.Equal(t, tc.msgs[0], trigger.Actions[0].GetCachedValue())
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
			msg := NewCreateTriggerRequest(tc.authority, tc.event, tc.msgs)
			trigger := NewTrigger(uint64(1), tc.authority, msg.Event, msg.Actions)
			queuedTrigger := NewQueuedTrigger(trigger, time.Time{}, uint64(1))
			err := queuedTrigger.UnpackInterfaces(cdc)
			assert.NoError(t, err)
			assert.Equal(t, tc.event, queuedTrigger.Trigger.Event.GetCachedValue())
			assert.Equal(t, tc.msgs[0], queuedTrigger.Trigger.Actions[0].GetCachedValue())
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
			msg := NewCreateTriggerRequest(tc.authority, tc.event, tc.msgs)
			trigger := NewTrigger(uint64(1), tc.authority, msg.Event, msg.Actions)
			err := trigger.UnpackInterfaces(cdc)
			assert.NoError(t, err)
			event, err := msg.GetTriggerEventI()
			if tc.err == nil {
				assert.NoError(t, err)
				assert.Equal(t, tc.event, event)
			} else {
				assert.Error(t, tc.err, err)
			}

		})
	}
}
