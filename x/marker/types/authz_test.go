package types_test

import (
	"github.com/provenance-io/provenance/internal/pioconfig"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simapp "github.com/provenance-io/provenance/app"
	. "github.com/provenance-io/provenance/x/marker/types"
)

var (
	coin1000   = sdk.NewCoin("stake", sdk.NewInt(1000))
	coin500    = sdk.NewCoin("stake", sdk.NewInt(500))
	msgTypeURL = "/provenance.marker.v1.MsgTransferRequest"
)

func init() {
	pioconfig.SetProvenanceConfig("", 0)
}

func TestMarkerTransferAuthorization(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	authorization := NewMarkerTransferAuthorization(sdk.NewCoins(coin1000))

	t.Run("verify authorization returns valid method name", func(t *testing.T) {
		require.Equal(t, authorization.MsgTypeURL(), msgTypeURL)
		require.NoError(t, authorization.ValidateBasic())
	})

	t.Run("verify updated authorization returns remaining spent limit", func(t *testing.T) {
		send := &MsgTransferRequest{Amount: coin500}
		resp, err := authorization.Accept(ctx, send)
		require.NoError(t, err)
		require.False(t, resp.Delete)
		require.NotNil(t, resp.Updated)
		sendAuth := NewMarkerTransferAuthorization(sdk.NewCoins(coin500))
		require.Equal(t, sendAuth.String(), resp.Updated.String())
	})

	t.Run("expect updated authorization delete after spending remaining amount", func(t *testing.T) {
		send := &MsgTransferRequest{Amount: coin1000}
		resp, err := authorization.Accept(ctx, send)
		require.NoError(t, err)
		require.True(t, resp.Delete)
		require.NotNil(t, resp.Updated)
	})

	t.Run("verify invalid transfer type", func(t *testing.T) {
		sendInvalid := &MsgBurnRequest{Amount: coin500}
		resp, err := authorization.Accept(ctx, sendInvalid)
		require.Error(t, err)
		require.Nil(t, resp.Updated)
	})
}
