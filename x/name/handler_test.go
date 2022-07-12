package name_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256r1"
	"github.com/golang/protobuf/proto"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/name"
	"github.com/provenance-io/provenance/x/name/keeper"
	nametypes "github.com/provenance-io/provenance/x/name/types"
)

func TestInvalidMsg(t *testing.T) {
	k := keeper.Keeper{}
	h := name.NewHandler(k)

	res, err := h(sdk.NewContext(nil, tmproto.Header{}, false, nil), testdata.NewTestMsg())
	require.Error(t, err)
	require.Nil(t, res)

	_, _, log := sdkerrors.ABCIInfo(err, false)
	require.True(t, strings.Contains(log, "unrecognized name message type"))
}

func containsMessage(result *sdk.Result, msg proto.Message) bool {
	events := result.GetEvents().ToABCIEvents()
	for _, event := range events {
		typeEvent, _ := sdk.ParseTypedEvent(event)
		if assert.ObjectsAreEqual(msg, typeEvent) {
			return true
		}
	}
	return false
}

//  create name record
func TestCreateName(t *testing.T) {
	priv1 := secp256k1.GenPrivKey()
	addr1 := sdk.AccAddress(priv1.PubKey().Address())
	priv2, _ := secp256r1.GenPrivKey()
	addr2 := sdk.AccAddress(priv2.PubKey().Address())

	tests := []struct {
		name          string
		expectedError error
		msg           *nametypes.MsgBindNameRequest
		expectedEvent proto.Message
	}{
		{
			name:          "create name record",
			msg:           nametypes.NewMsgBindNameRequest(nametypes.NewNameRecord("new", addr2, false), nametypes.NewNameRecord("example.name", addr1, false)),
			expectedError: nil,
			expectedEvent: nametypes.NewEventNameBound(addr2.String(), "new.example.name", false),
		},
		{
			name:          "create bad name record",
			msg:           nametypes.NewMsgBindNameRequest(nametypes.NewNameRecord("new", addr2, false), nametypes.NewNameRecord("foo.name", addr1, false)),
			expectedError: sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, nametypes.ErrNameNotBound.Error()),
		},
	}

	acc1 := &authtypes.BaseAccount{
		Address: addr1.String(),
	}
	acc2 := &authtypes.BaseAccount{
		Address: addr2.String(),
	}
	accs := authtypes.GenesisAccounts{acc1, acc2}
	app := simapp.SetupWithGenesisAccounts(accs)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	var nameData nametypes.GenesisState
	nameData.Bindings = append(nameData.Bindings, nametypes.NewNameRecord("name", addr1, false))
	nameData.Bindings = append(nameData.Bindings, nametypes.NewNameRecord("example.name", addr1, false))
	nameData.Params.AllowUnrestrictedNames = false
	nameData.Params.MaxNameLevels = 16
	nameData.Params.MinSegmentLength = 2
	nameData.Params.MaxSegmentLength = 16

	app.NameKeeper.InitGenesis(ctx, nameData)

	app.NameKeeper = keeper.NewKeeper(app.AppCodec(), app.GetKey(nametypes.ModuleName), app.GetSubspace(nametypes.ModuleName))
	handler := name.NewHandler(app.NameKeeper)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			response, err := handler(ctx, tc.msg)
			if tc.expectedError != nil {
				require.EqualError(t, err, tc.expectedError.Error())
			} else {
				require.NoError(t, err)
			}
			if tc.expectedEvent != nil {
				result := containsMessage(response, tc.expectedEvent)
				require.True(t, result, fmt.Sprintf("Expected typed event was not found: %v", tc.expectedEvent))
			}
		})
	}
}

//  delete name record
func TestDeleteName(t *testing.T) {
	priv1 := secp256k1.GenPrivKey()
	addr1 := sdk.AccAddress(priv1.PubKey().Address())

	tests := []struct {
		name          string
		expectedError error
		msg           *nametypes.MsgDeleteNameRequest
		expectedEvent proto.Message
	}{
		{
			name:          "delete name record",
			msg:           nametypes.NewMsgDeleteNameRequest(nametypes.NewNameRecord("example.name", addr1, false)),
			expectedError: nil,
			expectedEvent: nametypes.NewEventNameUnbound(addr1.String(), "example.name", false),
		},
		{
			name:          "create bad name record",
			msg:           nametypes.NewMsgDeleteNameRequest(nametypes.NewNameRecord("example.name", addr1, false)),
			expectedError: sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "name does not exist"),
		},
	}

	acc1 := &authtypes.BaseAccount{
		Address: addr1.String(),
	}
	accs := authtypes.GenesisAccounts{acc1}
	app := simapp.SetupWithGenesisAccounts(accs)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	var nameData nametypes.GenesisState
	nameData.Bindings = append(nameData.Bindings, nametypes.NewNameRecord("name", addr1, false))
	nameData.Bindings = append(nameData.Bindings, nametypes.NewNameRecord("example.name", addr1, false))
	nameData.Params.AllowUnrestrictedNames = false
	nameData.Params.MaxNameLevels = 16
	nameData.Params.MinSegmentLength = 2
	nameData.Params.MaxSegmentLength = 16

	app.NameKeeper.InitGenesis(ctx, nameData)

	app.NameKeeper = keeper.NewKeeper(app.AppCodec(), app.GetKey(nametypes.ModuleName), app.GetSubspace(nametypes.ModuleName))
	handler := name.NewHandler(app.NameKeeper)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			response, err := handler(ctx, tc.msg)
			if tc.expectedError != nil {
				require.EqualError(t, err, tc.expectedError.Error())
			} else {
				require.NoError(t, err)
			}
			if tc.expectedEvent != nil {
				result := containsMessage(response, tc.expectedEvent)
				require.True(t, result, fmt.Sprintf("Expected typed event was not found: %v", tc.expectedEvent))
			}
		})
	}
}
