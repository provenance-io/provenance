package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"testing"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
)

var (
	coin1000 = sdk.NewCoin("stake", sdk.NewInt(1000))
	coin500  = sdk.NewCoin("stake", sdk.NewInt(500))
)

func TestMarkerTransferAuthorization(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	authorization := NewMarkerTransferAuthorization(sdk.NewCoins(coin1000))

	t.Log("verify authorization returns valid method name")
	require.Equal(t, authorization.MsgTypeURL(), "/provenance.marker.v1.MsgTransferRequest")
	require.NoError(t, authorization.ValidateBasic())
	send := &MsgTransferRequest{Amount: coin500}

	t.Log("verify updated authorization returns remaining spent limit")
	resp, err := authorization.Accept(ctx, send)
	require.NoError(t, err)
	require.False(t, resp.Delete)
	require.NotNil(t, resp.Updated)
	sendAuth := NewMarkerTransferAuthorization(sdk.NewCoins(coin500))
	require.Equal(t, sendAuth.String(), resp.Updated.String())

	t.Log("expect updated authorization delete after spending remaining amount")
	resp, err = resp.Updated.Accept(ctx, send)
	require.NoError(t, err)
	require.True(t, resp.Delete)
	require.NotNil(t, resp.Updated)

	t.Log("verify invalid transfer type")
	sendInvalid := &MsgBurnRequest{Amount: coin500}
	resp, err = authorization.Accept(ctx, sendInvalid)
	require.Error(t, err)
	require.Nil(t, resp.Updated)
}
