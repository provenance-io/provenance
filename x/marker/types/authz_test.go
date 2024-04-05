package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/testutil/assertions"
	. "github.com/provenance-io/provenance/x/marker/types"
)

var (
	coin1000   = sdk.NewInt64Coin("stake", 1000)
	coin500    = sdk.NewInt64Coin("stake", 500)
	msgTypeURL = "/provenance.marker.v1.MsgTransferRequest"
)

func TestMarkerTransferAuthorization(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)
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
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	addr1 := sdk.AccAddress("addr1_______________")
	addr2 := sdk.AccAddress("addr2_______________")
	addr3 := sdk.AccAddress("addr3_______________")

	cases := []struct {
		name   string
		msg    MarkerTransferAuthorization
		expErr string
	}{
		{
			name:   "nil allow list",
			msg:    MarkerTransferAuthorization{TransferLimit: sdk.NewCoins(coin500)},
			expErr: "",
		},
		{
			name: "empty allow list",
			msg: MarkerTransferAuthorization{
				TransferLimit: sdk.NewCoins(coin500),
				AllowList:     []string{},
			},
			expErr: "",
		},
		{
			name: "non-empty allow list",
			msg: MarkerTransferAuthorization{
				TransferLimit: sdk.NewCoins(coin500),
				AllowList:     []string{addr1.String(), addr2.String(), addr3.String()},
			},
			expErr: "",
		},
		{
			name:   "nil transfer limit",
			msg:    MarkerTransferAuthorization{TransferLimit: nil},
			expErr: "invalid transfer limit: cannot be zero: invalid coins",
		},
		{
			name:   "empty transfer limit",
			msg:    MarkerTransferAuthorization{TransferLimit: sdk.Coins{}},
			expErr: "invalid transfer limit: cannot be zero: invalid coins",
		},
		{
			name:   "transfer limit with invalid denom",
			msg:    MarkerTransferAuthorization{TransferLimit: sdk.Coins{coin(3, "x")}},
			expErr: "invalid transfer limit: invalid denom: x: invalid coins",
		},
		{
			name:   "transfer limit with zero coin",
			msg:    MarkerTransferAuthorization{TransferLimit: sdk.Coins{coin(0, "catcoin")}},
			expErr: "invalid transfer limit: coin 0catcoin amount is not positive: invalid coins",
		},
		{
			name:   "transfer limit with negative coin",
			msg:    MarkerTransferAuthorization{TransferLimit: sdk.Coins{coin(-3, "catcoin")}},
			expErr: "invalid transfer limit: coin -3catcoin amount is not positive: invalid coins",
		},
		{
			name:   "unsorted transfer limit",
			msg:    MarkerTransferAuthorization{TransferLimit: sdk.Coins{coin(10, "banana"), coin(3, "apple")}},
			expErr: "invalid transfer limit: denomination apple is not sorted: invalid coins",
		},
		{
			name: "invalid allow list entry",
			msg: MarkerTransferAuthorization{
				TransferLimit: sdk.NewCoins(coin500),
				AllowList:     []string{addr1.String(), "notgonnawork", addr3.String()},
			},
			expErr: "invalid allow list entry [1] \"notgonnawork\": decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			name: "duplicate allow list entry",
			msg: MarkerTransferAuthorization{
				TransferLimit: sdk.NewCoins(coin500),
				AllowList:     []string{addr1.String(), addr2.String(), addr1.String()},
			},
			expErr: "invalid allow list entry [2] " + addr1.String() + ": duplicate entry",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.msg.ValidateBasic()
			}
			require.NotPanics(t, testFunc, "ValidateBasic")
			assertions.AssertErrorValue(t, err, tc.expErr, "ValidateBasic error")
		})
	}
}
