package name_test

import (
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/name"
	"github.com/provenance-io/provenance/x/name/keeper"
	nametypes "github.com/provenance-io/provenance/x/name/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
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

//  create name record
func TestCreateName(t *testing.T) {
	priv1 := secp256k1.GenPrivKey()
	addr1 := sdk.AccAddress(priv1.PubKey().Address())
	priv2 := secp256k1.GenPrivKey()
	addr2 := sdk.AccAddress(priv2.PubKey().Address())

	tests := []struct {
		name          string
		expectedError error
		msg           *nametypes.MsgBindNameRequest
		expectedEvent *nametypes.EventNameBound
	}{
		{
			name:          "create name record",
			msg:           nametypes.NewMsgBindNameRequest(nametypes.NewNameRecord("new", addr2, false), nametypes.NewNameRecord("example.name", addr1, false)),
			expectedError: nil,
			expectedEvent: &nametypes.EventNameBound{
				Address: addr2.String(),
				Name:    "new.example.name",
			},
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
				require.Equal(t, 1, len(response.GetEvents().ToABCIEvents()))
				msg1, _ := sdk.ParseTypedEvent(response.GetEvents().ToABCIEvents()[0])
				require.Equal(t, tc.expectedEvent, msg1)
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
		expectedEvent *nametypes.EventNameUnbound
	}{
		{
			name:          "delete name record",
			msg:           nametypes.NewMsgDeleteNameRequest(nametypes.NewNameRecord("example.name", addr1, false)),
			expectedError: nil,
			expectedEvent: &nametypes.EventNameUnbound{
				Address: addr1.String(),
				Name:    "example.name",
			},
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
				require.Equal(t, 1, len(response.GetEvents().ToABCIEvents()))
				msg1, _ := sdk.ParseTypedEvent(response.GetEvents().ToABCIEvents()[0])
				require.Equal(t, tc.expectedEvent, msg1)
			}
		})
	}
}
