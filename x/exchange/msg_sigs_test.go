package exchange_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/protoadapt"

	"cosmossdk.io/x/tx/signing"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/exchange"
)

const (
	emptyAddrErr = "empty address string is not allowed"
	bech32Err    = "decoding bech32 failed: "
)

type HasGetSigners interface {
	GetSigners() []sdk.AccAddress
}

func TestAllMsgsGetSigners(t *testing.T) {
	// getTypeName gets just the type name of the provided thing, e.g. "MsgGovCreateMarketRequest".
	getTypeName := func(thing interface{}) string {
		rv := fmt.Sprintf("%T", thing) // e.g. "*types.MsgGovCreateMarketRequest"
		lastDot := strings.LastIndex(rv, ".")
		if lastDot < 0 || lastDot+1 >= len(rv) {
			return rv
		}
		return rv[lastDot+1:]
	}

	testAddr := sdk.AccAddress("testAddr____________")
	badAddrStr := "badaddr"
	badAddrErr := bech32Err + "invalid bech32 string length 7"

	msgMakers := []func(signer string) sdk.Msg{
		func(signer string) sdk.Msg {
			return &exchange.MsgCreateAskRequest{AskOrder: exchange.AskOrder{Seller: signer}}
		},
		func(signer string) sdk.Msg {
			return &exchange.MsgCreateBidRequest{BidOrder: exchange.BidOrder{Buyer: signer}}
		},
		func(signer string) sdk.Msg {
			return &exchange.MsgCommitFundsRequest{Account: signer}
		},
		func(signer string) sdk.Msg {
			return &exchange.MsgCancelOrderRequest{Signer: signer}
		},
		func(signer string) sdk.Msg {
			return &exchange.MsgFillBidsRequest{Seller: signer}
		},
		func(signer string) sdk.Msg {
			return &exchange.MsgFillAsksRequest{Buyer: signer}
		},
		func(signer string) sdk.Msg {
			return &exchange.MsgMarketSettleRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &exchange.MsgMarketCommitmentSettleRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &exchange.MsgMarketReleaseCommitmentsRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &exchange.MsgMarketSetOrderExternalIDRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &exchange.MsgMarketWithdrawRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &exchange.MsgMarketUpdateDetailsRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &exchange.MsgMarketUpdateEnabledRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &exchange.MsgMarketUpdateAcceptingOrdersRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &exchange.MsgMarketUpdateUserSettleRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &exchange.MsgMarketUpdateAcceptingCommitmentsRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &exchange.MsgMarketUpdateIntermediaryDenomRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &exchange.MsgMarketManagePermissionsRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &exchange.MsgMarketManageReqAttrsRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &exchange.MsgCreatePaymentRequest{Payment: exchange.Payment{Source: signer}}
		},
		func(signer string) sdk.Msg {
			return &exchange.MsgAcceptPaymentRequest{Payment: exchange.Payment{Target: signer}}
		},
		func(signer string) sdk.Msg {
			return &exchange.MsgRejectPaymentRequest{Target: signer}
		},
		func(signer string) sdk.Msg {
			return &exchange.MsgRejectPaymentsRequest{Target: signer}
		},
		func(signer string) sdk.Msg {
			return &exchange.MsgCancelPaymentsRequest{Source: signer}
		},
		func(signer string) sdk.Msg {
			return &exchange.MsgChangePaymentTargetRequest{Source: signer}
		},
		func(signer string) sdk.Msg {
			return &exchange.MsgGovCreateMarketRequest{Authority: signer}
		},
		func(signer string) sdk.Msg {
			return &exchange.MsgGovManageFeesRequest{Authority: signer}
		},
		func(signer string) sdk.Msg {
			return &exchange.MsgGovCloseMarketRequest{Authority: signer}
		},
		func(signer string) sdk.Msg {
			return &exchange.MsgGovUpdateParamsRequest{Authority: signer}
		},
	}

	signerCases := []struct {
		name       string
		msgSigner  string
		expSigners []sdk.AccAddress
		expPanic   string
	}{
		{
			name:      "no signer",
			msgSigner: "",
			expPanic:  emptyAddrErr,
		},
		{
			name:       "good signer",
			msgSigner:  testAddr.String(),
			expSigners: []sdk.AccAddress{testAddr},
		},
		{
			name:      "bad signer",
			msgSigner: badAddrStr,
			expPanic:  badAddrErr,
		},
	}

	type testCase struct {
		name       string
		msg        sdk.Msg
		expSigners []sdk.AccAddress
		expPanic   string
	}

	var tests []testCase
	hasMaker := make(map[string]bool)

	for _, msgMaker := range msgMakers {
		typeName := getTypeName(msgMaker(""))
		hasMaker[typeName] = true
		for _, tc := range signerCases {
			tests = append(tests, testCase{
				name:       typeName + " " + tc.name,
				msg:        msgMaker(tc.msgSigner),
				expSigners: tc.expSigners,
				expPanic:   tc.expPanic,
			})
		}
	}

	encCfg := app.MakeTestEncodingConfig(t)
	sigCtx := encCfg.InterfaceRegistry.SigningContext()

	// Make sure all of the GetSigners() methods behave as expected.
	t.Run("legacy methods", func(t *testing.T) {
		for _, tc := range tests {
			t.Run(tc.name+" legacy", func(t *testing.T) {
				smsg, ok := tc.msg.(HasGetSigners)
				require.True(t, ok, "%T does not have a .GetSigners method.", tc.msg)

				var signers []sdk.AccAddress
				testFunc := func() {
					signers = smsg.GetSigners()
				}

				assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "GetSigners")
				assert.Equal(t, tc.expSigners, signers, "GetSigners")
			})
		}
	})

	// Make sure all of the msgs behave as expected with the new generic GetSigners(msg) function.
	t.Run("generic GetSigners(msg)", func(t *testing.T) {
		for _, tc := range tests {
			t.Run(tc.name+" generic", func(t *testing.T) {
				var expected [][]byte
				if tc.expSigners != nil {
					expected = make([][]byte, len(tc.expSigners))
					for i, signer := range tc.expSigners {
						expected[i] = signer
					}
				}

				var actual [][]byte
				var err error
				testFunc := func() {
					msgV2 := protoadapt.MessageV2Of(tc.msg)
					actual, err = sigCtx.GetSigners(msgV2)
				}
				require.NotPanics(t, testFunc, "sigCtx.GetSigners(msgV2)")
				if len(tc.expPanic) > 0 {
					assert.ErrorContains(t, err, tc.expPanic, "sigCtx.GetSigners(msgV2) error")
				} else {
					assert.NoError(t, err, "sigCtx.GetSigners(msgV2) error")
				}
				assert.Equal(t, expected, actual, "sigCtx.GetSigners(msgV2) result")
			})
		}
	})

	// Make sure all of the GetSigners funcs are tested.
	t.Run("all msgs have test case", func(t *testing.T) {
		for _, msg := range exchange.AllRequestMsgs {
			typeName := getTypeName(msg)
			t.Run(typeName, func(t *testing.T) {
				// If this fails, a maker needs to be defined above for the missing msg type.
				assert.True(t, hasMaker[typeName], "There is not a GetSigners test case for %s", typeName)
			})
		}
	})
}

func TestCreatePaymentGetSignersFunc(t *testing.T) {
	encCfg := app.MakeTestEncodingConfig(t)
	sigCtx := encCfg.InterfaceRegistry.SigningContext()
	opts := &signing.Options{
		AddressCodec:          sigCtx.AddressCodec(),
		ValidatorAddressCodec: sigCtx.ValidatorAddressCodec(),
	}

	tests := []struct {
		name      string
		fieldName string
		msg       sdk.Msg
		expAddrs  [][]byte
		expInErr  []string
	}{
		{
			name:      "msg without a payment field",
			fieldName: "whatever",
			msg:       &exchange.MsgCreateAskRequest{},
			expInErr:  []string{"no payment field found in provenance.exchange.v1.MsgCreateAskRequest"},
		},
		{
			name:      "no such field in the payment",
			fieldName: "no_such_thing",
			msg:       &exchange.MsgCreatePaymentRequest{},
			expInErr:  []string{"no payment.no_such_thing field found in provenance.exchange.v1.MsgCreatePaymentRequest"},
		},
		{
			name:      "field is not a string",
			fieldName: "source_amount",
			msg:       &exchange.MsgCreatePaymentRequest{},
			expInErr: []string{"panic (recovered) getting provenance.exchange.v1.MsgCreatePaymentRequest.payment.source_amount as a signer",
				"interface conversion", "not string"},
		},
		{
			name:      "field is empty",
			fieldName: "source",
			msg:       &exchange.MsgCreatePaymentRequest{},
			expInErr:  []string{"error decoding payment.source address \"\"", emptyAddrErr},
		},
		{
			name:      "invalid bech32",
			fieldName: "source",
			msg:       &exchange.MsgCreatePaymentRequest{Payment: exchange.Payment{Source: "not_an_address"}},
			expInErr:  []string{"error decoding payment.source address \"not_an_address\"", bech32Err},
		},
		{
			name:      "all good: create and source",
			fieldName: "source",
			msg: &exchange.MsgCreatePaymentRequest{
				Payment: exchange.Payment{Source: sdk.AccAddress("source_address______").String()},
			},
			expAddrs: [][]byte{[]byte("source_address______")},
		},
		{
			name:      "all good: accept and target",
			fieldName: "target",
			msg: &exchange.MsgAcceptPaymentRequest{
				Payment: exchange.Payment{Target: sdk.AccAddress("target_address______").String()},
			},
			expAddrs: [][]byte{[]byte("target_address______")},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var getSignersFn signing.GetSignersFunc
			testMaker := func() {
				getSignersFn = exchange.CreatePaymentGetSignersFunc(opts, tc.fieldName)
			}
			require.NotPanics(t, testMaker, "CreatePaymentGetSignersFunc(%q)", tc.fieldName)

			var actAddrs [][]byte
			var actErr error
			testGetter := func() {
				msgV2 := protoadapt.MessageV2Of(tc.msg)
				actAddrs, actErr = getSignersFn(msgV2)
			}
			require.NotPanics(t, testGetter, "custom GetSigners function")
			assertions.AssertErrorContents(t, actErr, tc.expInErr, "custom GetSigners function error")
			assert.Equal(t, tc.expAddrs, actAddrs, "custom GetSigners function addresses")
		})
	}
}

func TestDefineCustomGetSigners(t *testing.T) {
	encCfg := app.MakeTestEncodingConfig(t)
	sigCtx := encCfg.InterfaceRegistry.SigningContext()

	sigOpts := signing.Options{
		AddressCodec:          sigCtx.AddressCodec(),
		ValidatorAddressCodec: sigCtx.ValidatorAddressCodec(),
	}

	testFunc := func() {
		exchange.DefineCustomGetSigners(&sigOpts)
	}
	require.NotPanics(t, testFunc, "DefineCustomGetSigners")
	assert.Len(t, sigOpts.CustomGetSigners, 2, "CustomGetSigners")

	tests := []struct {
		msg sdk.Msg
		exp []byte
	}{
		{
			msg: &exchange.MsgCreatePaymentRequest{Payment: exchange.Payment{Source: sdk.AccAddress("source______________").String()}},
			exp: []byte("source______________"),
		},
		{
			msg: &exchange.MsgAcceptPaymentRequest{Payment: exchange.Payment{Target: sdk.AccAddress("target______________").String()}},
			exp: []byte("target______________"),
		},
	}

	for _, tc := range tests {
		msgV2 := protoadapt.MessageV2Of(tc.msg)
		name := protov2.MessageName(msgV2)
		expected := [][]byte{tc.exp}

		// Make sure the custom entries are added to the map, and that they work as expected.
		t.Run(string(name)+" CustomGetSigners", func(t *testing.T) {
			getSignersFn := sigOpts.CustomGetSigners[name]
			if assert.NotNil(t, getSignersFn, "sigOpts.CustomGetSigners[%q]", name) {
				var actual [][]byte
				var err error
				testGetSigners := func() {
					actual, err = getSignersFn(msgV2)
				}
				require.NotPanics(t, testGetSigners, "getSignersFn", name)
				assert.NoError(t, err, "getSignersFn error", name)
				assert.Equal(t, expected, actual, "getSignersFn result", name)
			}
		})

		// Make sure the custom entries are added to the encoder and that that works as expected.
		t.Run(string(name)+"GetSigners", func(t *testing.T) {
			var actual [][]byte
			var err error
			testGetSigners := func() {
				actual, err = sigCtx.GetSigners(msgV2)
			}
			require.NotPanics(t, testGetSigners, "sigCtx.GetSigners(msg)")
			assert.NoError(t, err, "sigCtx.GetSigners(msg) error")
			assert.Equal(t, expected, actual, "sigCtx.GetSigners(msg) result")
		})
	}
}
