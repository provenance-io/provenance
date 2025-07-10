package types_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gogoproto/proto"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/testutil/assertions"
	. "github.com/provenance-io/provenance/x/marker/types"
)

func TestMarkerTransferAuthorization(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)

	msgTypeURL := "/provenance.marker.v1.MsgTransferRequest"
	coin1000 := sdk.NewInt64Coin("stake", 1000)
	coin500 := sdk.NewInt64Coin("stake", 500)
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
	coin500 := sdk.NewInt64Coin("stake", 500)
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
			t.Run(tt.name, func(t *testing.T) {
				ma, err := NewMultiAuthorization(tt.msgType, tt.auths...)
				require.NoError(t, err, "NewMultiAuthorization should not fail on packing")

				err = ma.ValidateBasic()
				if tt.expectErr {
					require.Error(t, err, "ValidateBasic")
				} else {
					require.NoError(t, err, "ValidateBasic")
				}
			})
		})
	}
}

// mockAuthorization satisfies the authz.Authorization interface and does what it's told.
type mockAuthorization struct {
	msgTypeURL string

	AcceptCalled bool

	RespAccept  bool
	RespDelete  bool
	RespUpdated bool
	RespErr     string
}

var _ authz.Authorization = (*mockAuthorization)(nil)

// newMockAuthorization creates a new mock authorization.
func newMockAuthorization() *mockAuthorization {
	return &mockAuthorization{}
}

// ToAccept sets this mockAuthorization up to return Accept = true in the accept response.
func (a *mockAuthorization) WithMsgTypeURL(msgTypeURL string) *mockAuthorization {
	a.msgTypeURL = msgTypeURL
	return a
}

// ToAccept sets this mockAuthorization up to return Accept = true in the accept response.
func (a *mockAuthorization) WithMsgType(msgType proto.Message) *mockAuthorization {
	a.msgTypeURL = sdk.MsgTypeURL(msgType)
	return a
}

// ToAccept sets this mockAuthorization up to return Accept = true in the accept response.
func (a *mockAuthorization) ToAccept() *mockAuthorization {
	a.RespAccept = true
	return a
}

// ToAccept sets this mockAuthorization up to return Delete = true in the accept response.
func (a *mockAuthorization) ToDelete() *mockAuthorization {
	a.RespDelete = true
	return a
}

// ToAccept sets this mockAuthorization up to return itself in the accept response's Updated field.
func (a *mockAuthorization) ToUpdate() *mockAuthorization {
	a.RespUpdated = true
	return a
}

func (a *mockAuthorization) WasCalled() *mockAuthorization {
	a.AcceptCalled = true
	return a
}

// ToAccept sets this mockAuthorization up to the provided error from Accept.
func (a *mockAuthorization) WithError(err string) *mockAuthorization {
	a.RespErr = err
	return a
}

// Accept just returns everything it was defined to return.
func (a *mockAuthorization) Accept(_ context.Context, _ sdk.Msg) (authz.AcceptResponse, error) {
	a.AcceptCalled = true
	resp := authz.AcceptResponse{
		Accept: a.RespAccept,
		Delete: a.RespDelete,
	}
	if a.RespUpdated {
		resp.Updated = a
	}
	var err error
	if len(a.RespErr) > 0 {
		err = errors.New(a.RespErr)
	}
	return resp, err
}

// MsgTypeURL returns "mockAuthorization". Satisfies the authz.Authorization interface.
func (a *mockAuthorization) MsgTypeURL() string {
	return a.msgTypeURL
}

// ValidateBasic returns nil. Satisfies the authz.Authorization interface.
func (a *mockAuthorization) ValidateBasic() error {
	return nil
}

// Reset does nothing. Satisfies the authz.Authorization interface.
func (a *mockAuthorization) Reset() {}

// String returns a string representation of this mockAuthorization.
func (a *mockAuthorization) String() string {
	return fmt.Sprintf("mockAuthorization{msgTypeURL=%q,Accept=%t,Delete=%t,Update=%t,Err=%q}",
		a.msgTypeURL, a.RespAccept, a.RespDelete, a.RespUpdated, a.RespErr)
}

// ProtoMessage does nothing. Satisfies the authz.Authorization interface.
func (a *mockAuthorization) ProtoMessage() {}

func TestMultiAuthorizationAccept(t *testing.T) {
	newAny := func(v proto.Message) *codectypes.Any {
		rv, err := codectypes.NewAnyWithValue(v)
		require.NoError(t, err, "NewAnyWithValue(%#v)", v)
		return rv
	}
	newMultiAuthz := func(msgTypeURL string, subAuths ...*codectypes.Any) *MultiAuthorization {
		return &MultiAuthorization{
			MsgTypeUrl:        msgTypeURL,
			SubAuthorizations: subAuths,
		}
	}

	msgSend := &banktypes.MsgSend{
		FromAddress: "from_address________",
		ToAddress:   "to_address__________",
		Amount:      sdk.NewCoins(sdk.NewInt64Coin("banana", 99)),
	}
	msgSendTypeURL := sdk.MsgTypeURL(msgSend)

	tests := []struct {
		name       string
		multiAuthz *MultiAuthorization
		msg        sdk.Msg
		expErr     string
		expResp    authz.AcceptResponse
	}{
		{
			name: "wrong msg type",
			multiAuthz: newMultiAuthz(msgSendTypeURL+"2",
				newAny(newMockAuthorization()),
				newAny(newMockAuthorization()),
			),
			msg:     msgSend,
			expErr:  "message type mismatch",
			expResp: authz.AcceptResponse{},
		},
		{
			name: "nil first sub-auth",
			multiAuthz: newMultiAuthz(msgSendTypeURL,
				nil, newAny(newMockAuthorization())),
			msg:     msgSend,
			expErr:  "sub-authorization 0 is nil",
			expResp: authz.AcceptResponse{},
		},
		{
			name: "nil second sub-auth",
			multiAuthz: newMultiAuthz(msgSendTypeURL,
				newAny(newMockAuthorization().ToAccept()), nil),
			msg:     msgSend,
			expErr:  "sub-authorization 1 is nil",
			expResp: authz.AcceptResponse{},
		},
		{
			name: "error from first accept",
			multiAuthz: newMultiAuthz(msgSendTypeURL,
				newAny(newMockAuthorization().WithError("no-accepty")),
				newAny(newMockAuthorization()),
			),
			msg:     msgSend,
			expErr:  "sub-authorization 0 was not accepted: no-accepty",
			expResp: authz.AcceptResponse{},
		},
		{
			name: "error from second accept",
			multiAuthz: newMultiAuthz(msgSendTypeURL,
				newAny(newMockAuthorization().ToAccept()),
				newAny(newMockAuthorization().WithError("what-now-this-time")),
			),
			msg:     msgSend,
			expErr:  "sub-authorization 1 was not accepted: what-now-this-time",
			expResp: authz.AcceptResponse{},
		},
		{
			name: "two: not accept first",
			multiAuthz: newMultiAuthz(msgSendTypeURL,
				newAny(newMockAuthorization()),
				newAny(newMockAuthorization().ToAccept()),
			),
			msg:     msgSend,
			expResp: authz.AcceptResponse{Accept: false},
		},
		{
			name: "two: not accept second",
			multiAuthz: newMultiAuthz(msgSendTypeURL,
				newAny(newMockAuthorization().ToAccept()),
				newAny(newMockAuthorization()),
			),
			msg:     msgSend,
			expResp: authz.AcceptResponse{Accept: false},
		},
		{
			name: "two: both accept",
			multiAuthz: newMultiAuthz(msgSendTypeURL,
				newAny(newMockAuthorization().ToAccept()),
				newAny(newMockAuthorization().ToAccept()),
			),
			msg:     msgSend,
			expResp: authz.AcceptResponse{Accept: true},
		},
		{
			name: "two: delete from first",
			multiAuthz: newMultiAuthz(msgSendTypeURL,
				newAny(newMockAuthorization().ToAccept().ToDelete()),
				newAny(newMockAuthorization().ToAccept()),
			),
			msg:     msgSend,
			expResp: authz.AcceptResponse{Accept: true, Delete: true},
		},
		{
			name: "two: delete from second",
			multiAuthz: newMultiAuthz(msgSendTypeURL,
				newAny(newMockAuthorization().ToAccept()),
				newAny(newMockAuthorization().ToAccept().ToDelete()),
			),
			msg:     msgSend,
			expResp: authz.AcceptResponse{Accept: true, Delete: true},
		},
		{
			name: "two: delete from both",
			multiAuthz: newMultiAuthz(msgSendTypeURL,
				newAny(newMockAuthorization().ToAccept().ToDelete()),
				newAny(newMockAuthorization().ToAccept().ToDelete()),
			),
			msg:     msgSend,
			expResp: authz.AcceptResponse{Accept: true, Delete: true},
		},
		{
			name: "two: delete then not accept",
			multiAuthz: newMultiAuthz(msgSendTypeURL,
				newAny(newMockAuthorization().ToAccept().ToDelete()),
				newAny(newMockAuthorization()),
			),
			msg:     msgSend,
			expResp: authz.AcceptResponse{Accept: false},
		},
		{
			name: "two: not accept then delete",
			multiAuthz: newMultiAuthz(msgSendTypeURL,
				newAny(newMockAuthorization()),
				newAny(newMockAuthorization().ToAccept().ToDelete()),
			),
			msg:     msgSend,
			expResp: authz.AcceptResponse{Accept: false},
		},
		{
			name: "two: delete then update",
			multiAuthz: newMultiAuthz(msgSendTypeURL,
				newAny(newMockAuthorization().ToAccept().ToDelete()),
				newAny(newMockAuthorization().ToAccept().ToUpdate()),
			),
			msg:     msgSend,
			expResp: authz.AcceptResponse{Accept: true, Delete: true},
		},
		{
			name: "two: update then delete",
			multiAuthz: newMultiAuthz(msgSendTypeURL,
				newAny(newMockAuthorization().ToAccept().ToUpdate()),
				newAny(newMockAuthorization().ToAccept().ToDelete()),
			),
			msg:     msgSend,
			expResp: authz.AcceptResponse{Accept: true, Delete: true},
		},
		{
			name: "two: update from first",
			multiAuthz: newMultiAuthz(msgSendTypeURL,
				newAny(newMockAuthorization().ToAccept().ToUpdate()),
				newAny(newMockAuthorization().ToAccept()),
			),
			msg: msgSend,
			expResp: authz.AcceptResponse{
				Accept: true,
				Updated: newMultiAuthz(msgSendTypeURL,
					newAny(newMockAuthorization().ToAccept().ToUpdate().WasCalled()),
					newAny(newMockAuthorization().ToAccept().WasCalled()),
				),
			},
		},
		{
			name: "two: update from second",
			multiAuthz: newMultiAuthz(msgSendTypeURL,
				newAny(newMockAuthorization().ToAccept()),
				newAny(newMockAuthorization().ToAccept().ToUpdate()),
			),
			msg: msgSend,
			expResp: authz.AcceptResponse{
				Accept: true,
				Updated: newMultiAuthz(msgSendTypeURL,
					newAny(newMockAuthorization().ToAccept().WasCalled()),
					newAny(newMockAuthorization().ToAccept().ToUpdate().WasCalled()),
				),
			},
		},
		{
			name: "two: update from both",
			multiAuthz: newMultiAuthz(msgSendTypeURL,
				newAny(newMockAuthorization().ToAccept().ToUpdate()),
				newAny(newMockAuthorization().ToAccept().ToUpdate()),
			),
			msg: msgSend,
			expResp: authz.AcceptResponse{
				Accept: true,
				Updated: newMultiAuthz(msgSendTypeURL,
					// We don't use WasCalled() on this first one since it's not marked for update.
					newAny(newMockAuthorization().ToAccept().ToUpdate().WasCalled()),
					newAny(newMockAuthorization().ToAccept().ToUpdate().WasCalled()),
				),
			},
		},
		{
			name: "three: not accept first",
			multiAuthz: newMultiAuthz(msgSendTypeURL,
				newAny(newMockAuthorization()),
				newAny(newMockAuthorization().ToAccept()),
				newAny(newMockAuthorization().ToAccept()),
			),
			msg:     msgSend,
			expResp: authz.AcceptResponse{Accept: false},
		},
		{
			name: "three: not accept first",
			multiAuthz: newMultiAuthz(msgSendTypeURL,
				newAny(newMockAuthorization().ToAccept()),
				newAny(newMockAuthorization()),
				newAny(newMockAuthorization().ToAccept()),
			),
			msg:     msgSend,
			expResp: authz.AcceptResponse{Accept: false},
		},
		{
			name: "three: not accept third",
			multiAuthz: newMultiAuthz(msgSendTypeURL,
				newAny(newMockAuthorization().ToAccept()),
				newAny(newMockAuthorization().ToAccept()),
				newAny(newMockAuthorization()),
			),
			msg:     msgSend,
			expResp: authz.AcceptResponse{Accept: false},
		},
		{
			name: "three: all accept",
			multiAuthz: newMultiAuthz(msgSendTypeURL,
				newAny(newMockAuthorization().ToAccept()),
				newAny(newMockAuthorization().ToAccept()),
				newAny(newMockAuthorization().ToAccept()),
			),
			msg:     msgSend,
			expResp: authz.AcceptResponse{Accept: true},
		},
		{
			name: "three: update from first",
			multiAuthz: newMultiAuthz(msgSendTypeURL,
				newAny(newMockAuthorization().ToAccept().ToUpdate()),
				newAny(newMockAuthorization().ToAccept()),
				newAny(newMockAuthorization().ToAccept()),
			),
			msg: msgSend,
			expResp: authz.AcceptResponse{
				Accept: true,
				Updated: newMultiAuthz(msgSendTypeURL,
					newAny(newMockAuthorization().ToAccept().ToUpdate().WasCalled()),
					newAny(newMockAuthorization().ToAccept().WasCalled()),
					newAny(newMockAuthorization().ToAccept().WasCalled()),
				),
			},
		},
		{
			name: "three: update from second",
			multiAuthz: newMultiAuthz(msgSendTypeURL,
				newAny(newMockAuthorization().ToAccept()),
				newAny(newMockAuthorization().ToAccept().ToUpdate()),
				newAny(newMockAuthorization().ToAccept()),
			),
			msg: msgSend,
			expResp: authz.AcceptResponse{
				Accept: true,
				Updated: newMultiAuthz(msgSendTypeURL,
					newAny(newMockAuthorization().ToAccept().WasCalled()),
					newAny(newMockAuthorization().ToAccept().ToUpdate().WasCalled()),
					newAny(newMockAuthorization().ToAccept().WasCalled()),
				),
			},
		},
		{
			name: "three: update from third",
			multiAuthz: newMultiAuthz(msgSendTypeURL,
				newAny(newMockAuthorization().ToAccept()),
				newAny(newMockAuthorization().ToAccept()),
				newAny(newMockAuthorization().ToAccept().ToUpdate()),
			),
			msg: msgSend,
			expResp: authz.AcceptResponse{
				Accept: true,
				Updated: newMultiAuthz(msgSendTypeURL,
					newAny(newMockAuthorization().ToAccept().WasCalled()),
					newAny(newMockAuthorization().ToAccept().WasCalled()),
					newAny(newMockAuthorization().ToAccept().ToUpdate().WasCalled()),
				),
			},
		},
		{
			name: "three: update from all",
			multiAuthz: newMultiAuthz(msgSendTypeURL,
				newAny(newMockAuthorization().ToAccept().ToUpdate()),
				newAny(newMockAuthorization().ToAccept().ToUpdate()),
				newAny(newMockAuthorization().ToAccept().ToUpdate()),
			),
			msg: msgSend,
			expResp: authz.AcceptResponse{
				Accept: true,
				Updated: newMultiAuthz(msgSendTypeURL,
					newAny(newMockAuthorization().ToAccept().ToUpdate().WasCalled()),
					newAny(newMockAuthorization().ToAccept().ToUpdate().WasCalled()),
					newAny(newMockAuthorization().ToAccept().ToUpdate().WasCalled()),
				),
			},
		},
		{
			name: "three: delete from first",
			multiAuthz: newMultiAuthz(msgSendTypeURL,
				newAny(newMockAuthorization().ToAccept().ToDelete()),
				newAny(newMockAuthorization().ToAccept()),
				newAny(newMockAuthorization().ToAccept()),
			),
			msg:     msgSend,
			expResp: authz.AcceptResponse{Accept: true, Delete: true},
		},
		{
			name: "three: delete from second",
			multiAuthz: newMultiAuthz(msgSendTypeURL,
				newAny(newMockAuthorization().ToAccept()),
				newAny(newMockAuthorization().ToAccept().ToDelete()),
				newAny(newMockAuthorization().ToAccept()),
			),
			msg:     msgSend,
			expResp: authz.AcceptResponse{Accept: true, Delete: true},
		},
		{
			name: "three: delete from third",
			multiAuthz: newMultiAuthz(msgSendTypeURL,
				newAny(newMockAuthorization().ToAccept()),
				newAny(newMockAuthorization().ToAccept()),
				newAny(newMockAuthorization().ToAccept().ToDelete()),
			),
			msg:     msgSend,
			expResp: authz.AcceptResponse{Accept: true, Delete: true},
		},
		{
			name: "three: update from first and second, delete from third",
			multiAuthz: newMultiAuthz(msgSendTypeURL,
				newAny(newMockAuthorization().ToAccept().ToUpdate()),
				newAny(newMockAuthorization().ToAccept().ToUpdate()),
				newAny(newMockAuthorization().ToAccept().ToDelete()),
			),
			msg:     msgSend,
			expResp: authz.AcceptResponse{Accept: true, Delete: true},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var resp authz.AcceptResponse
			var err error
			testFunc := func() {
				resp, err = tc.multiAuthz.Accept(sdk.Context{}, tc.msg)
			}
			require.NotPanics(t, testFunc, "MultiAuthorization.Accept")
			assertions.AssertErrorValue(t, err, tc.expErr, "MultiAuthorization.Accept error")
			assert.Equal(t, tc.expResp.Accept, resp.Accept, "resp.Accept")
			assert.Equal(t, tc.expResp.Delete, resp.Delete, "resp.Delete")

			require.Equal(t, tc.expResp.Updated == nil, resp.Updated == nil, "resp.Updated == nil")
			if tc.expResp.Updated == nil || resp.Updated == nil {
				return
			}
			expMulti, eOK := tc.expResp.Updated.(*MultiAuthorization)
			actMulti, aOK := resp.Updated.(*MultiAuthorization)
			if !eOK || !aOK {
				// If they're not multi-authorizations, just compare them directly.
				// It's an annoying failure message to try to figure out, but it's better than nothing.
				assert.Equal(t, tc.expResp.Updated, resp.Updated, "resp.Updated")
				return
			}
			assert.Equal(t, expMulti.MsgTypeUrl, actMulti.MsgTypeUrl, "updated MsgTypeUrl")
			require.Equal(t, len(expMulti.SubAuthorizations), len(actMulti.SubAuthorizations), "updated sub-authorizations")
			for i := range expMulti.SubAuthorizations {
				expSub := expMulti.SubAuthorizations[i]
				actSub := actMulti.SubAuthorizations[i]
				assert.Equal(t, expSub, actSub, "[%d/%d]: updated sub-authorizations", i+1, len(expMulti.SubAuthorizations))
			}
		})
	}
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
