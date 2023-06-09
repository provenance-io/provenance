package types

import (
	fmt "fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/stretchr/testify/assert"
)

func TestNewCreateTriggerRequest(t *testing.T) {
	authority := "addr"
	var event TriggerEventI = &BlockHeightEvent{BlockHeight: 1}
	msgs := []sdk.Msg{&MsgDestroyTriggerRequest{Id: 5, Authority: authority}}
	actions, _ := sdktx.SetMsgs(msgs)
	eventAny, _ := codectypes.NewAnyWithValue(event)
	expected := &MsgCreateTriggerRequest{
		Authority: authority,
		Actions:   actions,
		Event:     eventAny,
	}

	trigger := NewCreateTriggerRequest(expected.Authority, event, msgs)
	assert.Equal(t, expected, trigger, "should create the correct request with NewCreateTriggerRequest")
}

func TestNewDestroyTriggerRequest(t *testing.T) {
	expected := MsgDestroyTriggerRequest{
		Id:        2,
		Authority: "addr",
	}

	request := NewDestroyTriggerRequest(expected.Authority, expected.Id)
	assert.Equal(t, &expected, request, "should create the correct request with DestroyTriggerRequest")
}

func TestMsgCreateTriggerRequestValidateBasic(t *testing.T) {
	tests := []struct {
		name      string
		authority string
		event     TriggerEventI
		msgs      []sdk.Msg
		err       string
	}{
		{
			name:      "valid - successful validate",
			authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
			event:     &BlockHeightEvent{},
			msgs:      []sdk.Msg{&MsgDestroyTriggerRequest{Authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", Id: 1}},
			err:       "",
		},
		{
			name:      "invalid - address is not correct format",
			authority: "badaddr",
			event:     &BlockHeightEvent{},
			msgs:      []sdk.Msg{&MsgDestroyTriggerRequest{Authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", Id: 1}},
			err:       "invalid address for trigger authority from address: decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			name:      "invalid - missing actions",
			authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
			event:     &BlockHeightEvent{},
			msgs:      []sdk.Msg{},
			err:       "trigger must contain actions",
		},
		{
			name:      "invalid - actions validate failed",
			authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
			event:     &BlockHeightEvent{},
			msgs:      []sdk.Msg{&MsgDestroyTriggerRequest{Authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"}},
			err:       "msg: 0, err: invalid id for trigger",
		},
		{
			name:      "invalid - event validation failed",
			authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
			event:     &TransactionEvent{},
			msgs:      []sdk.Msg{&MsgDestroyTriggerRequest{Authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", Id: 1}},
			err:       "empty event name",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := NewCreateTriggerRequest(tc.authority, tc.event, tc.msgs)
			err := msg.ValidateBasic()
			if len(tc.err) > 0 {
				assert.EqualError(t, err, tc.err, "should have error in ValidateBasic")
			} else {
				assert.NoError(t, err, "should have no error in successful ValidateBasic")
			}
		})
	}
}

func TestMsgCreateTriggerRequestGetSigners(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgCreateTriggerRequest
		signers []sdk.AccAddress
	}{
		{
			name:    "valid - Get signers returns the correct signers",
			msg:     NewCreateTriggerRequest("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", &BlockHeightEvent{}, []sdk.Msg{}),
			signers: []sdk.AccAddress{sdk.MustAccAddressFromBech32("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			signers := tc.msg.GetSigners()
			assert.Equal(t, tc.signers, signers, "should receive the correct signers from GetSigners")
		})
	}
}

func TestMsgCreateTriggerRequestUnpackInterfaces(t *testing.T) {
	cdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

	tests := []struct {
		name      string
		authority string
		event     TriggerEventI
		msgs      []sdk.Msg
	}{
		{
			name:      "valid - Unpack Interfaces",
			authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
			event:     &BlockHeightEvent{},
			msgs:      []sdk.Msg{&MsgDestroyTriggerRequest{Authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", Id: 1}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := NewCreateTriggerRequest(tc.authority, tc.event, tc.msgs)
			err := msg.UnpackInterfaces(cdc)
			assert.NoError(t, err, "should have no error for UnpackInterfaces")
			assert.Equal(t, tc.event, msg.Event.GetCachedValue(), "should have cached value for Event in UnpackInterfaces")
			assert.Equal(t, tc.msgs[0], msg.Actions[0].GetCachedValue(), "should have cached value for Actions in UnpackInterfaces")
		})
	}
}

func TestMsgCreateTriggerRequestGetTriggerEventI(t *testing.T) {
	cdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

	tests := []struct {
		name      string
		authority string
		event     TriggerEventI
		msgs      []sdk.Msg
		err       string
	}{
		{
			name:      "valid - GetTriggerEventI",
			authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
			event:     &BlockHeightEvent{},
			msgs:      []sdk.Msg{&MsgDestroyTriggerRequest{Authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", Id: 1}},
			err:       "",
		},
		{
			name:      "invalid - Returns error when interface is nil",
			authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
			event:     nil,
			msgs:      []sdk.Msg{&MsgDestroyTriggerRequest{Authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", Id: 1}},
			err:       "event is nil: trigger does not have event",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			event := tc.event
			if event == nil {
				event = &BlockHeightEvent{}
			}
			msg := NewCreateTriggerRequest(tc.authority, event, tc.msgs)
			if tc.event == nil {
				msg.Event = nil
			}
			err := msg.UnpackInterfaces(cdc)
			assert.NoError(t, err, "should have no error for UnpackInterfaces")
			triggerEvent, err := msg.GetTriggerEventI()
			if len(tc.err) == 0 {
				assert.NoError(t, err, "should have no error for GetTriggerEventI after UnpackInterfaces")
				assert.Equal(t, tc.event, triggerEvent, "should have matching events after UnpackInterfaces")
			} else {
				assert.EqualError(t, err, tc.err)
			}

		})
	}
}

func TestMsgDestroyTriggerRequestValidateBasic(t *testing.T) {
	tests := []struct {
		name      string
		authority string
		id        uint64
		err       error
	}{
		{
			name:      "valid - success",
			authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
			id:        1,
			err:       nil,
		},
		{
			name:      "invalid - bad address",
			authority: "badaddr",
			id:        1,
			err:       fmt.Errorf("invalid address for trigger authority from address: decoding bech32 failed: invalid bech32 string length 7"),
		},
		{
			name:      "invalid - bad id",
			authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
			id:        0,
			err:       fmt.Errorf("invalid id for trigger"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := NewDestroyTriggerRequest(tc.authority, tc.id)
			err := msg.ValidateBasic()
			if tc.err != nil {
				assert.Error(t, err, "should receive correct error for failed ValidateBasic")
			} else {
				assert.NoError(t, err, "should receive no error for successful ValidateBasic")
			}
		})
	}
}

func TestMsgDestroyTriggerRequestGetSigners(t *testing.T) {
	tests := []struct {
		name         string
		authority    string
		id           uint64
		panicMessage string
	}{
		{
			name:         "valid - success",
			authority:    "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
			id:           1,
			panicMessage: "",
		},
		{
			name:         "invalid - bad addr",
			authority:    "badaddr",
			id:           1,
			panicMessage: "decoding bech32 failed: invalid bech32 string length 7",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := NewDestroyTriggerRequest(tc.authority, tc.id)
			if len(tc.panicMessage) > 0 {
				assert.PanicsWithError(t, tc.panicMessage, func() {
					msg.GetSigners()
				})
			} else {
				signers := msg.GetSigners()
				assert.Equal(t, []sdk.AccAddress{sdk.MustAccAddressFromBech32(tc.authority)}, signers, "should only contain authority in GetSigners")
			}
		})
	}
}
