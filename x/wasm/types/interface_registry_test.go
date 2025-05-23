package types

import (
	"fmt"
	"testing"

	wasmv1 "github.com/CosmWasm/wasmd/x/wasm/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	v1beta1 "github.com/provenance-io/provenance/x/wasm"
	"github.com/stretchr/testify/require"
)

func setupTestApp(t *testing.T) (*TestApp, sdk.Context) {
	// Simplified version of app setup for testing
	app := NewTestApp()
	ctx := app.NewContext(false)
	return app, ctx
}

type TestApp struct {
	GovKeeper govKeeperMock
}

func NewTestApp() *TestApp {
	return &TestApp{
		GovKeeper: govKeeperMock{},
	}
}

func (app *TestApp) NewContext(iscCheckTx bool) sdk.Context {
	return sdk.NewContext(nil, tmproto.Header{}, iscCheckTx, nil)
}

type govKeeperMock struct{}

func (g govKeeperMock) SetProposal(ctx sdk.Context, p govtypes.Proposal) {}
func (g govKeeperMock) GetProposal(ctx sdk.Context, id uint64) govtypes.Proposal {
	// Return empty or mock proposal
	return govtypes.Proposal{}
}

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

func Test_Decode_v1beta1_MsgInGovProposal(t *testing.T) {
	app, ctx := setupTestApp(t)
	registry := types.NewInterfaceRegistry()
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&wasmv1.MsgExecuteContract{},  // new type
		&v1beta1.MsgExecuteContract{}, // legacy type
		&govtypes.Proposal{},
	)
	cdc := codec.NewProtoCodec(registry)
	oldMsg := &v1beta1.MsgExecuteContract{
		Sender:   "pbmos1abc...",
		Contract: "pbsmos1def...",
		Msg:      []byte(`{"do_something":{}}`),
	}
	msgAny, err := types.NewAnyWithValue(oldMsg)
	require.NoError(t, err)
	proposal := &govtypes.Proposal{
		Id:       1,
		Messages: []*types.Any{msgAny},
	}
	app.GovKeeper.SetProposal(ctx, *proposal)
	stored := app.GovKeeper.GetProposal(ctx, 1)
	require.NotNil(t, stored)
	for _, msg := range stored.Messages {
		var unpacked sdk.Msg
		err := cdc.UnpackAny(msg, &unpacked)
		require.NoError(t, err)
		switch m := unpacked.(type) {
		case *v1beta1.MsgExecuteContract:
			require.Equal(t, oldMsg.Sender, m.Sender)
			require.Equal(t, oldMsg.Contract, m.Contract)
		default:
			t.Fatalf("unexpected message type: %T", m)
		}
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
	// This mimics what happens when the node starts and loads proposals from genesis
	bz, err := cdc.Marshal(genesisProposal)
	require.NoError(t, err)
	var loaded govtypes.Proposal
	fmt.Printf("marshalled value %s", bz)
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
