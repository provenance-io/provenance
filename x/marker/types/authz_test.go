package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
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

func TestNewMultiAuthorization(t *testing.T) {
	msgTypeURL := sdk.MsgTypeURL(&banktypes.MsgSend{})

	tests := []struct {
		name      string
		msgType   string
		auths     []authz.Authorization
		expectErr bool
	}{
		{
			name:    "valid multi-authorization",
			msgType: msgTypeURL,
			auths: []authz.Authorization{
				&authz.GenericAuthorization{Msg: msgTypeURL},
				&authz.GenericAuthorization{Msg: msgTypeURL},
			},
			expectErr: false,
		},
		{
			name:      "empty message type",
			msgType:   "",
			auths:     []authz.Authorization{&authz.GenericAuthorization{Msg: msgTypeURL}},
			expectErr: true,
		},
		{
			name:      "no sub-authorizations",
			msgType:   msgTypeURL,
			auths:     []authz.Authorization{},
			expectErr: true,
		},
		{
			name:    "too many sub-authorizations",
			msgType: msgTypeURL,
			auths: func() []authz.Authorization {
				auths := make([]authz.Authorization, MaxSubAuthorizations+1)
				for i := range auths {
					auths[i] = &authz.GenericAuthorization{Msg: msgTypeURL}
				}
				return auths
			}(),
			expectErr: true,
		},
		{
			name:    "mismatched message types",
			msgType: msgTypeURL,
			auths: []authz.Authorization{
				&authz.GenericAuthorization{Msg: msgTypeURL},
				&authz.GenericAuthorization{Msg: "/different.type"},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewMultiAuthorization(tt.msgType, tt.auths...)
			if tt.expectErr {
				require.Error(t, err, "NewMultiAuthorization")
			} else {
				require.NoError(t, err, "NewMultiAuthorization")
			}
		})
	}
}

func TestMultiAuthorizationAccept(t *testing.T) {
	msgTypeURL := sdk.MsgTypeURL(&banktypes.MsgSend{})

	// Create MultiAuthorization
	auth1 := &authz.GenericAuthorization{Msg: msgTypeURL}
	auth2 := &authz.GenericAuthorization{Msg: msgTypeURL}

	multiAuth, err := NewMultiAuthorization(msgTypeURL, auth1, auth2)
	require.NoError(t, err, "NewMultiAuthorization")

	// Setup codec for unpacking
	registry := codectypes.NewInterfaceRegistry()
	authz.RegisterInterfaces(registry)
	RegisterInterfaces(registry)

	err = multiAuth.UnpackInterfaces(registry)
	require.NoError(t, err, "UnpackInterfaces")

	ctx := sdk.Context{}
	msg := &banktypes.MsgSend{
		FromAddress: "pbmos1from",
		ToAddress:   "pbmos1to",
		Amount:      sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(100))),
	}

	// Test accept
	resp, err := multiAuth.Accept(ctx, msg)
	require.NoError(t, err, "multiAuth.Accept")
	require.True(t, resp.Accept)
	require.False(t, resp.Delete)
	require.Nil(t, resp.Updated, "multiAuth.Accept")

	// Test wrong message type
	wrongMsg := &banktypes.MsgMultiSend{}
	_, err = multiAuth.Accept(ctx, wrongMsg)
	require.Error(t, err, "wrongMsg")
}

func TestMultiAuthorizationValidateBasic(t *testing.T) {
	msgTypeURL := sdk.MsgTypeURL(&banktypes.MsgSend{})

	auth1 := &authz.GenericAuthorization{Msg: msgTypeURL}
	auth2 := &authz.GenericAuthorization{Msg: msgTypeURL}

	multiAuth, err := NewMultiAuthorization(msgTypeURL, auth1, auth2)
	require.NoError(t, err, "NewMultiAuthorization")

	err = multiAuth.ValidateBasic()
	require.NoError(t, err, "ValidateBasic")

	// Test invalid
	invalidAuth := &MultiAuthorization{
		MsgTypeUrl:        "",
		SubAuthorizations: []*codectypes.Any{},
	}

	err = invalidAuth.ValidateBasic()
	require.Error(t, err, "invalidAuth.ValidateBasic()")
}

func TestMultiAuthorizationCodec(t *testing.T) {
	registry := codectypes.NewInterfaceRegistry()
	authz.RegisterInterfaces(registry)
	RegisterInterfaces(registry)

	msgTypeURL := sdk.MsgTypeURL(&banktypes.MsgSend{})
	auth1 := &authz.GenericAuthorization{Msg: msgTypeURL}
	auth2 := &authz.GenericAuthorization{Msg: msgTypeURL}

	multiAuth, err := NewMultiAuthorization(msgTypeURL, auth1, auth2)
	require.NoError(t, err, "NewMultiAuthorization")

	any, err := codectypes.NewAnyWithValue(multiAuth)
	require.NoError(t, err, "codectypes.NewAnyWithValue")

	var unpacked authz.Authorization
	err = registry.UnpackAny(any, &unpacked)
	require.NoError(t, err, "UnpackAny")

	unpackedMulti, ok := unpacked.(*MultiAuthorization)
	require.True(t, ok, "unpacked")
	require.Equal(t, multiAuth.MsgTypeUrl, unpackedMulti.MsgTypeUrl)
	require.Len(t, unpackedMulti.SubAuthorizations, 2)
}
