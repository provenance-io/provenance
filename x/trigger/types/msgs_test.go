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
	assert.Equal(t, expected, trigger)
}

func TestNewDestroyTriggerRequest(t *testing.T) {
	expected := MsgDestroyTriggerRequest{
		Id:        2,
		Authority: "addr",
	}

	request := NewDestroyTriggerRequest(expected.Authority, expected.Id)
	assert.Equal(t, &expected, request)
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
			err:       "invalid id for trigger",
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
				assert.ErrorContains(t, err, tc.err)
			} else {
				assert.NoError(t, err)
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
			assert.Equal(t, tc.signers, signers)
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
			assert.NoError(t, err)
			assert.Equal(t, tc.event, msg.Event.GetCachedValue())
			assert.Equal(t, tc.msgs[0], msg.Actions[0].GetCachedValue())
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
			err := msg.UnpackInterfaces(cdc)
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
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMsgDestroyTriggerRequestGetSigners(t *testing.T) {
	tests := []struct {
		name      string
		authority string
		id        uint64
		panics    bool
	}{
		{
			name:      "valid - success",
			authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
			id:        1,
			panics:    false,
		},
		{
			name:      "invalid - bad addr",
			authority: "badaddr",
			id:        1,
			panics:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := NewDestroyTriggerRequest(tc.authority, tc.id)
			if tc.panics {
				assert.Panics(t, func() {
					msg.GetSigners()
				})
			} else {
				signers := msg.GetSigners()
				assert.Equal(t, []sdk.AccAddress{sdk.MustAccAddressFromBech32(tc.authority)}, signers)
			}
		})
	}
}
