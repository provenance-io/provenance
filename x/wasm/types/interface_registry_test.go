package types

import (
	"testing"

	wasmv1 "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v1beta1 "github.com/provenance-io/provenance/x/wasm"
	"github.com/stretchr/testify/require"
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
			anyMsg, err := types.NewAnyWithValue(tt.msg)
			require.NoError(t, err)
			require.Equal(t, tt.typeURL, anyMsg.TypeUrl)
			// Unpack back
			var unpacked sdk.Msg
			err = cdc.UnpackAny(anyMsg, &unpacked)
			require.NoError(t, err)
			require.NotNil(t, unpacked)
			require.IsType(t, tt.msg, unpacked)
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
	require.Error(t, err)
	require.Contains(t, err.Error(), "no registered implementations")
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
			require.NoError(t, err)
			require.Equal(t, tt.typeURL, anyMsg.TypeUrl)

			// Unpack back
			var unpacked sdk.Msg
			err = cdc.UnpackAny(anyMsg, &unpacked)
			require.NoError(t, err)
			require.NotNil(t, unpacked)
			require.IsType(t, tt.msg, unpacked)

			// Check field values (optional but good for validation)
			if v1msg, ok := tt.msg.(*wasmv1.MsgExecuteContract); ok {
				unpackedV1, _ := unpacked.(*wasmv1.MsgExecuteContract)
				require.Equal(t, v1msg.Sender, unpackedV1.Sender)
				require.Equal(t, v1msg.Contract, unpackedV1.Contract)
			}
			if beta1msg, ok := tt.msg.(*v1beta1.MsgExecuteContract); ok {
				unpackedBeta1, _ := unpacked.(*v1beta1.MsgExecuteContract)
				require.Equal(t, beta1msg.Sender, unpackedBeta1.Sender)
				require.Equal(t, beta1msg.Contract, unpackedBeta1.Contract)
			}
		})
	}
}
