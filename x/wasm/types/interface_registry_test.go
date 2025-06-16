package types

import (
	"testing"

	wasmv1 "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1beta1 "github.com/provenance-io/provenance/x/wasm"
)

func Test_InterfaceRegistry_RegisterImplementations(t *testing.T) {
	registry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	tests := []struct {
		name    string
		msg     sdk.Msg
		typeURL string
	}{
		{
			name:    "v1 MsgExecuteContract",
			msg:     &wasmv1.MsgExecuteContract{},
			typeURL: "/cosmwasm.wasm.v1.MsgExecuteContract",
		},
		{
			name:    "v1beta1 MsgExecuteContract",
			msg:     &v1beta1.MsgExecuteContract{},
			typeURL: "/cosmwasm.wasm.v1beta1.MsgExecuteContract",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Pack into Any
			// Pack into Any
			anyMsg, err := types.NewAnyWithValue(tt.msg)
			require.NoError(t, err, "NewAnyWithValue")
			require.Equal(t, tt.typeURL, anyMsg.TypeUrl, "anyMsg.TypeUrl")

			// Unpack it.
			var unpacked sdk.Msg
			err = cdc.UnpackAny(anyMsg, &unpacked)
			require.NoError(t, err, "UnpackAny")
			require.NotNil(t, unpacked, "UnpackAny result")
			require.IsType(t, tt.msg, unpacked, "UnpackAny result")
		})
	}
}

func Test_Decode_UnknownType(t *testing.T) {
	registry := types.NewInterfaceRegistry()

	cdc := codec.NewProtoCodec(registry)

	// Create an Any with a completely unknown type URL
	unknownAny := &types.Any{
		TypeUrl: "/some.unknown.MsgType",
		Value:   []byte("garbage"),
	}

	var msg sdk.Msg
	err := cdc.UnpackAny(unknownAny, &msg)
	require.ErrorContains(t, err, "no registered implementations")
}

func Test_Decode_MsgExecuteContract(t *testing.T) {
	registry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	tests := []struct {
		name    string
		msg     sdk.Msg
		typeURL string
	}{
		{
			name:    "v1 MsgExecuteContract",
			msg:     &wasmv1.MsgExecuteContract{Sender: "cosmos1abc...", Contract: "cosmos1def..."},
			typeURL: "/cosmwasm.wasm.v1.MsgExecuteContract",
		},
		{
			name:    "v1beta1 MsgExecuteContract",
			msg:     &v1beta1.MsgExecuteContract{Sender: "cosmos1xyz...", Contract: "cosmos1uvw..."},
			typeURL: "/cosmwasm.wasm.v1beta1.MsgExecuteContract",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Pack into Any
			anyMsg, err := types.NewAnyWithValue(tt.msg)
			require.NoError(t, err, "NewAnyWithValue failed")
			require.Equal(t, tt.typeURL, anyMsg.TypeUrl, "unexpected type URL")

			// Unpack back
			var unpacked sdk.Msg
			err = cdc.UnpackAny(anyMsg, &unpacked)
			require.NoError(t, err, "UnpackAny failed")
			require.NotNil(t, unpacked, "unpacked message is nil")
			require.IsType(t, tt.msg, unpacked, "unpacked message has unexpected type")

			// Check field values (use assert to continue on failure)
			if v1msg, ok := tt.msg.(*wasmv1.MsgExecuteContract); ok {
				unpackedV1, _ := unpacked.(*wasmv1.MsgExecuteContract)
				assert.Equal(t, v1msg.Sender, unpackedV1.Sender, "Sender mismatch for v1 MsgExecuteContract")
				assert.Equal(t, v1msg.Contract, unpackedV1.Contract, "Contract mismatch for v1 MsgExecuteContract")
			}
			if beta1msg, ok := tt.msg.(*v1beta1.MsgExecuteContract); ok {
				unpackedBeta1, _ := unpacked.(*v1beta1.MsgExecuteContract)
				assert.Equal(t, beta1msg.Sender, unpackedBeta1.Sender, "Sender mismatch for v1beta1 MsgExecuteContract")
				assert.Equal(t, beta1msg.Contract, unpackedBeta1.Contract, "Contract mismatch for v1beta1 MsgExecuteContract")
			}
		})
	}
}

func Test_Decode_v1beta1_MsgFromGenesis(t *testing.T) {
	registry := types.NewInterfaceRegistry()
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&v1beta1.MsgExecuteContract{},
		&wasmv1.MsgExecuteContract{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&govtypes.Proposal{},
	)

	cdc := codec.NewProtoCodec(registry)
	oldMsg := &v1beta1.MsgExecuteContract{
		Sender:   "pbmos1abc...",
		Contract: "pbmos1def...",
		Msg:      []byte(`{"do_something":{}}`),
	}
	msgAny, err := types.NewAnyWithValue(oldMsg)
	require.NoError(t, err)

	genesisProposal := &govtypes.Proposal{
		Id:       1,
		Messages: []*types.Any{msgAny},
	}

	bz, err := cdc.Marshal(genesisProposal)
	require.NoError(t, err)

	var loaded govtypes.Proposal
	err = cdc.Unmarshal(bz, &loaded)
	require.NoError(t, err)

	for _, msg := range loaded.Messages {
		var unpacked sdk.Msg
		err := cdc.UnpackAny(msg, &unpacked)
		require.NoError(t, err)
		_, ok := unpacked.(*v1beta1.MsgExecuteContract)
		require.True(t, ok, "expected v1beta1.MsgExecuteContract")
	}
}
