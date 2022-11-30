package types_test

import (
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

func TestMarkerTransferAuthorization(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	authorization := NewMarkerTransferAuthorization(sdk.NewCoins(coin1000), []sdk.AccAddress{})

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
		sendAuth := NewMarkerTransferAuthorization(sdk.NewCoins(coin500), []sdk.AccAddress{})
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

func TestMarkerTransferAuthorizationValidateBasic(t *testing.T) {
	addr1, _ := sdk.AccAddressFromBech32("cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck")

	cases := []struct {
		name     string
		msg      *MarkerTransferAuthorization
		errorMsg string
	}{
		{
			"valid msg with empty allow list",
			NewMarkerTransferAuthorization(sdk.NewCoins(coin500), []sdk.AccAddress{}),
			"",
		},
		{
			"valid msg without non-empty allow list",
			NewMarkerTransferAuthorization(sdk.NewCoins(coin500), []sdk.AccAddress{addr1}),
			"",
		},
		{
			"invalid msg with duplicate allow list",
			NewMarkerTransferAuthorization(sdk.NewCoins(coin500), []sdk.AccAddress{addr1, addr1}),
			"all allow list addresses must be unique: duplicate entry",
		},
	}
	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if len(tc.errorMsg) > 0 {
				require.Error(t, err)
				require.Equal(t, tc.errorMsg, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
