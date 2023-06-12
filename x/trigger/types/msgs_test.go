package types

import (
	fmt "fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
)

func TestNewCreateTriggerRequest(t *testing.T) {
	authorities := []string{"addr1", "addr2"}
	var event TriggerEventI = &BlockHeightEvent{BlockHeight: 1}
	msgs := []sdk.Msg{&MsgDestroyTriggerRequest{Id: 5, Authority: authorities[0]}}
	actions, _ := sdktx.SetMsgs(msgs)
	eventAny, _ := codectypes.NewAnyWithValue(event)
	expected := &MsgCreateTriggerRequest{
		Authorities: authorities,
		Actions:     actions,
		Event:       eventAny,
	}

	trigger := NewCreateTriggerRequest(expected.Authorities, event, msgs)
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
		name        string
		authorities []string
		event       TriggerEventI
		msgs        []sdk.Msg
		err         string
	}{
		{
			name:        "valid - successful validate",
			authorities: []string{"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"},
			event:       &BlockHeightEvent{},
			msgs:        []sdk.Msg{&MsgDestroyTriggerRequest{Authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", Id: 1}},
			err:         "",
		},
		{
			name:        "invalid - address is not correct format",
			authorities: []string{"badaddr"},
			event:       &BlockHeightEvent{},
			msgs:        []sdk.Msg{&MsgDestroyTriggerRequest{Authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", Id: 1}},
			err:         "invalid address for trigger authority from address: decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			name:        "invalid - missing actions",
			authorities: []string{"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"},
			event:       &BlockHeightEvent{},
			msgs:        []sdk.Msg{},
			err:         "trigger must contain actions",
		},
		{
			name:        "invalid - actions validate failed",
			authorities: []string{"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"},
			event:       &BlockHeightEvent{},
			msgs:        []sdk.Msg{&MsgDestroyTriggerRequest{Authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"}},
			err:         "action: 0: invalid id for trigger",
		},
		{
			name:        "invalid - event validation failed",
			authorities: []string{"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"},
			event:       &TransactionEvent{},
			msgs:        []sdk.Msg{&MsgDestroyTriggerRequest{Authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", Id: 1}},
			err:         "empty event name",
		},
		{
			name:        "invalid - authorities must match",
			authorities: []string{"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"},
			event:       &BlockHeightEvent{},
			msgs:        []sdk.Msg{&MsgDestroyTriggerRequest{Authority: "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqs2m6sx4", Id: 1}},
			err:         "action: 0: signers[0] \"cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqs2m6sx4\" is not a signer of the request message",
		},
		{
			name:        "valid - the action's signer must be in authorities subset",
			authorities: []string{"cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqs2m6sx4", "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"},
			event:       &BlockHeightEvent{},
			msgs:        []sdk.Msg{&MsgDestroyTriggerRequest{Authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", Id: 1}},
			err:         "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := NewCreateTriggerRequest(tc.authorities, tc.event, tc.msgs)
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
			name: "valid - Get signers returns the correct signers",
			msg: NewCreateTriggerRequest(
				[]string{"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqs2m6sx4"},
				&BlockHeightEvent{},
				[]sdk.Msg{},
			),
			signers: []sdk.AccAddress{
				sdk.MustAccAddressFromBech32("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
				sdk.MustAccAddressFromBech32("cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqs2m6sx4"),
			},
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
		name        string
		authorities []string
		event       TriggerEventI
		msgs        []sdk.Msg
	}{
		{
			name:        "valid - Unpack Interfaces",
			authorities: []string{"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"},
			event:       &BlockHeightEvent{},
			msgs:        []sdk.Msg{&MsgDestroyTriggerRequest{Authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", Id: 1}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := NewCreateTriggerRequest(tc.authorities, tc.event, tc.msgs)
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
		name        string
		authorities []string
		event       TriggerEventI
		msgs        []sdk.Msg
		err         string
	}{
		{
			name:        "valid - GetTriggerEventI",
			authorities: []string{"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"},
			event:       &BlockHeightEvent{},
			msgs:        []sdk.Msg{&MsgDestroyTriggerRequest{Authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", Id: 1}},
			err:         "",
		},
		{
			name:        "invalid - Returns error when interface is nil",
			authorities: []string{"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"},
			event:       nil,
			msgs:        []sdk.Msg{&MsgDestroyTriggerRequest{Authority: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", Id: 1}},
			err:         "event is nil: trigger does not have event",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			event := tc.event
			if event == nil {
				event = &BlockHeightEvent{}
			}
			msg := NewCreateTriggerRequest(tc.authorities, event, tc.msgs)
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
