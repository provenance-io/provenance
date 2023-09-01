package exchange

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/testutil/assertions"
)

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
	badAddrErr := "decoding bech32 failed: invalid bech32 string length 7"
	emptyAddrErr := "empty address string is not allowed"

	msgMakers := []func(signer string) sdk.Msg{
		func(signer string) sdk.Msg {
			return &MsgGovCreateMarketRequest{Authority: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgGovManageFeesRequest{Authority: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgGovUpdateParamsRequest{Authority: signer}
		},
		// TODO[1658]: Add the rest of the messages once they've actually been defined.
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

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var signers []sdk.AccAddress
			testFunc := func() {
				signers = tc.msg.GetSigners()
			}

			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "GetSigners")
			assert.Equal(t, tc.expSigners, signers, "GetSigners")
		})
	}

	// Make sure all of the GetSigners and GetSignerStrs funcs are tested.
	t.Run("all msgs have test case", func(t *testing.T) {
		for _, msg := range allRequestMsgs {
			typeName := getTypeName(msg)
			// If this fails, a maker needs to be defined above for the missing msg type.
			assert.True(t, hasMaker[typeName], "whether a GetSigners test exists for %s", typeName)
		}
	})
}

func testValidateBasic(t *testing.T, msg sdk.Msg, expErr []string) {
	t.Helper()
	var err error
	testFunc := func() {
		err = msg.ValidateBasic()
	}
	require.NotPanics(t, testFunc, "%T.ValidateBasic()", msg)

	assertions.AssertErrorContents(t, err, expErr, "%T.ValidateBasic() error", msg)
}

// TODO[1658]: func TestMsgCreateAskRequest_ValidateBasic(t *testing.T)

// TODO[1658]: func TestMsgCreateBidRequest_ValidateBasic(t *testing.T)

// TODO[1658]: func TestMsgCancelOrderRequest_ValidateBasic(t *testing.T)

// TODO[1658]: func TestMsgFillBidsRequest_ValidateBasic(t *testing.T)

// TODO[1658]: func TestMsgFillAsksRequest_ValidateBasic(t *testing.T)

// TODO[1658]: func TestMsgMarketSettleRequest_ValidateBasic(t *testing.T)

// TODO[1658]: func TestMsgMarketWithdrawRequest_ValidateBasic(t *testing.T)

// TODO[1658]: func TestMsgMarketUpdateDetailsRequest_ValidateBasic(t *testing.T)

// TODO[1658]: func TestMsgMarketUpdateEnabledRequest_ValidateBasic(t *testing.T)

// TODO[1658]: func TestMsgMarketUpdateUserSettleRequest_ValidateBasic(t *testing.T)

// TODO[1658]: func TestMsgMarketManagePermissionsRequest_ValidateBasic(t *testing.T)

// TODO[1658]: func TestMsgMarketManageReqAttrsRequest_ValidateBasic(t *testing.T)

func TestMsgGovCreateMarketRequest_ValidateBasic(t *testing.T) {
	authority := sdk.AccAddress("authority___________").String()

	validMarket := Market{
		MarketId: 1,
		MarketDetails: MarketDetails{
			Name:        "Test Market One",
			Description: "This is the first test market",
		},
		FeeCreateAskFlat:        []sdk.Coin{sdk.NewInt64Coin("nhash", 10)},
		FeeCreateBidFlat:        []sdk.Coin{sdk.NewInt64Coin("nhash", 20)},
		FeeSettlementSellerFlat: []sdk.Coin{sdk.NewInt64Coin("nhash", 50)},
		FeeSettlementSellerRatios: []FeeRatio{
			{Price: sdk.NewInt64Coin("nhash", 100), Fee: sdk.NewInt64Coin("nhash", 1)},
		},
		FeeSettlementBuyerFlat: []sdk.Coin{sdk.NewInt64Coin("nhash", 100)},
		FeeSettlementBuyerRatios: []FeeRatio{
			{Price: sdk.NewInt64Coin("nhash", 100), Fee: sdk.NewInt64Coin("nhash", 1)},
		},
		AcceptingOrders:     true,
		AllowUserSettlement: true,
		AccessGrants: []AccessGrant{
			{
				Address:     sdk.AccAddress("just_some_addr______").String(),
				Permissions: AllPermissions(),
			},
		},
		ReqAttrCreateAsk: []string{"one.attr.pb"},
		ReqAttrCreateBid: []string{"*.attr.pb"},
	}

	tests := []struct {
		name   string
		msg    MsgGovCreateMarketRequest
		expErr []string
	}{
		{
			name:   "zero value",
			msg:    MsgGovCreateMarketRequest{},
			expErr: []string{"invalid authority"},
		},
		{
			name: "control",
			msg: MsgGovCreateMarketRequest{
				Authority: authority,
				Market:    validMarket,
			},
			expErr: nil,
		},
		{
			name: "no authority",
			msg: MsgGovCreateMarketRequest{
				Authority: "",
				Market:    validMarket,
			},
			expErr: []string{"invalid authority", "empty address string is not allowed"},
		},
		{
			name: "bad authority",
			msg: MsgGovCreateMarketRequest{
				Authority: "bad",
				Market:    validMarket,
			},
			expErr: []string{"invalid authority", "decoding bech32 failed"},
		},
		{
			name: "invalid market",
			msg: MsgGovCreateMarketRequest{
				Authority: authority,
				Market: Market{
					FeeCreateAskFlat: []sdk.Coin{{Denom: "badbad", Amount: sdkmath.NewInt(0)}},
				},
			},
			expErr: []string{`invalid create-ask flat fee option "0badbad": amount cannot be zero`},
		},
		{
			name: "multiple errors",
			msg: MsgGovCreateMarketRequest{
				Authority: "",
				Market: Market{
					FeeCreateBidFlat: []sdk.Coin{{Denom: "badbad", Amount: sdkmath.NewInt(0)}},
				},
			},
			expErr: []string{
				"invalid authority",
				`invalid create-bid flat fee option "0badbad": amount cannot be zero`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

func TestMsgGovManageFeesRequest_ValidateBasic(t *testing.T) {
	authority := sdk.AccAddress("authority___________").String()
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	ratio := func(priceAmount int64, priceDenom string, feeAmount int64, feeDenom string) FeeRatio {
		return FeeRatio{Price: coin(priceAmount, priceDenom), Fee: coin(feeAmount, feeDenom)}
	}

	tests := []struct {
		name   string
		msg    MsgGovManageFeesRequest
		expErr []string
	}{
		{
			name:   "zero value",
			msg:    MsgGovManageFeesRequest{},
			expErr: []string{"invalid authority", "no updates"},
		},
		{
			name: "no authority",
			msg: MsgGovManageFeesRequest{
				Authority:           "",
				AddFeeCreateAskFlat: []sdk.Coin{coin(1, "nhash")},
			},
			expErr: []string{"invalid authority", "empty address string is not allowed"},
		},
		{
			name: "bad authority",
			msg: MsgGovManageFeesRequest{
				Authority:           "bad",
				AddFeeCreateAskFlat: []sdk.Coin{coin(1, "nhash")},
			},
			expErr: []string{"invalid authority", "decoding bech32 failed"},
		},
		{
			name: "invalid add create-ask flat",
			msg: MsgGovManageFeesRequest{
				Authority:           authority,
				AddFeeCreateAskFlat: []sdk.Coin{coin(0, "nhash")},
			},
			expErr: []string{`invalid create-ask flat fee to add option "0nhash": amount cannot be zero`},
		},
		{
			name: "same add and remove create-ask flat",
			msg: MsgGovManageFeesRequest{
				Authority:              authority,
				AddFeeCreateAskFlat:    []sdk.Coin{coin(1, "nhash")},
				RemoveFeeCreateAskFlat: []sdk.Coin{coin(1, "nhash")},
			},
			expErr: []string{"cannot add and remove the same create-ask flat fee options: 1nhash"},
		},
		{
			name: "invalid add create-bid flat",
			msg: MsgGovManageFeesRequest{
				Authority:           authority,
				AddFeeCreateBidFlat: []sdk.Coin{coin(0, "nhash")},
			},
			expErr: []string{`invalid create-bid flat fee to add option "0nhash": amount cannot be zero`},
		},
		{
			name: "same add and remove create-bid flat",
			msg: MsgGovManageFeesRequest{
				Authority:              authority,
				AddFeeCreateBidFlat:    []sdk.Coin{coin(1, "nhash")},
				RemoveFeeCreateBidFlat: []sdk.Coin{coin(1, "nhash")},
			},
			expErr: []string{"cannot add and remove the same create-bid flat fee options: 1nhash"},
		},
		{
			name: "invalid add settlement seller flat",
			msg: MsgGovManageFeesRequest{
				Authority:                  authority,
				AddFeeSettlementSellerFlat: []sdk.Coin{coin(0, "nhash")},
			},
			expErr: []string{`invalid seller settlement flat fee to add option "0nhash": amount cannot be zero`},
		},
		{
			name: "same add and remove settlement seller flat",
			msg: MsgGovManageFeesRequest{
				Authority:                     authority,
				AddFeeSettlementSellerFlat:    []sdk.Coin{coin(1, "nhash")},
				RemoveFeeSettlementSellerFlat: []sdk.Coin{coin(1, "nhash")},
			},
			expErr: []string{"cannot add and remove the same seller settlement flat fee options: 1nhash"},
		},
		{
			name: "invalid add settlement seller ratio",
			msg: MsgGovManageFeesRequest{
				Authority:                    authority,
				AddFeeSettlementSellerRatios: []FeeRatio{ratio(1, "nhash", 2, "nhash")},
			},
			expErr: []string{`seller fee ratio fee amount "2nhash" cannot be greater than price amount "1nhash"`},
		},
		{
			name: "same add and remove settlement seller ratio",
			msg: MsgGovManageFeesRequest{
				Authority:                       authority,
				AddFeeSettlementSellerRatios:    []FeeRatio{ratio(2, "nhash", 1, "nhash")},
				RemoveFeeSettlementSellerRatios: []FeeRatio{ratio(2, "nhash", 1, "nhash")},
			},
			expErr: []string{"cannot add and remove the same seller settlement fee ratios: 2nhash:1nhash"},
		},
		{
			name: "invalid add settlement buyer flat",
			msg: MsgGovManageFeesRequest{
				Authority:                 authority,
				AddFeeSettlementBuyerFlat: []sdk.Coin{coin(0, "nhash")},
			},
			expErr: []string{`invalid buyer settlement flat fee to add option "0nhash": amount cannot be zero`},
		},
		{
			name: "same add and remove settlement buyer flat",
			msg: MsgGovManageFeesRequest{
				Authority:                    authority,
				AddFeeSettlementBuyerFlat:    []sdk.Coin{coin(1, "nhash")},
				RemoveFeeSettlementBuyerFlat: []sdk.Coin{coin(1, "nhash")},
			},
			expErr: []string{"cannot add and remove the same buyer settlement flat fee options: 1nhash"},
		},
		{
			name: "invalid add settlement buyer ratio",
			msg: MsgGovManageFeesRequest{
				Authority:                   authority,
				AddFeeSettlementBuyerRatios: []FeeRatio{ratio(1, "nhash", 2, "nhash")},
			},
			expErr: []string{`buyer fee ratio fee amount "2nhash" cannot be greater than price amount "1nhash"`},
		},
		{
			name: "same add and remove settlement buyer ratio",
			msg: MsgGovManageFeesRequest{
				Authority:                      authority,
				AddFeeSettlementBuyerRatios:    []FeeRatio{ratio(2, "nhash", 1, "nhash")},
				RemoveFeeSettlementBuyerRatios: []FeeRatio{ratio(2, "nhash", 1, "nhash")},
			},
			expErr: []string{"cannot add and remove the same buyer settlement fee ratios: 2nhash:1nhash"},
		},
		{
			name: "multiple errors",
			msg: MsgGovManageFeesRequest{
				Authority:                       "",
				AddFeeCreateAskFlat:             []sdk.Coin{coin(0, "nhash")},
				RemoveFeeCreateAskFlat:          []sdk.Coin{coin(0, "nhash")},
				AddFeeCreateBidFlat:             []sdk.Coin{coin(0, "nhash")},
				RemoveFeeCreateBidFlat:          []sdk.Coin{coin(0, "nhash")},
				AddFeeSettlementSellerFlat:      []sdk.Coin{coin(0, "nhash")},
				RemoveFeeSettlementSellerFlat:   []sdk.Coin{coin(0, "nhash")},
				AddFeeSettlementSellerRatios:    []FeeRatio{ratio(1, "nhash", 2, "nhash")},
				RemoveFeeSettlementSellerRatios: []FeeRatio{ratio(1, "nhash", 2, "nhash")},
				AddFeeSettlementBuyerFlat:       []sdk.Coin{coin(0, "nhash")},
				RemoveFeeSettlementBuyerFlat:    []sdk.Coin{coin(0, "nhash")},
				AddFeeSettlementBuyerRatios:     []FeeRatio{ratio(1, "nhash", 2, "nhash")},
				RemoveFeeSettlementBuyerRatios:  []FeeRatio{ratio(1, "nhash", 2, "nhash")},
			},
			expErr: []string{
				"invalid authority", "empty address string is not allowed",
				`invalid create-ask flat fee to add option "0nhash": amount cannot be zero`,
				"cannot add and remove the same create-ask flat fee options: 0nhash",
				`invalid create-bid flat fee to add option "0nhash": amount cannot be zero`,
				"cannot add and remove the same create-bid flat fee options: 0nhash",
				`invalid seller settlement flat fee to add option "0nhash": amount cannot be zero`,
				"cannot add and remove the same seller settlement flat fee options: 0nhash",
				`seller fee ratio fee amount "2nhash" cannot be greater than price amount "1nhash"`,
				"cannot add and remove the same seller settlement fee ratios: 1nhash:2nhash",
				`invalid buyer settlement flat fee to add option "0nhash": amount cannot be zero`,
				"cannot add and remove the same buyer settlement flat fee options: 0nhash",
				`buyer fee ratio fee amount "2nhash" cannot be greater than price amount "1nhash"`,
				"cannot add and remove the same buyer settlement fee ratios: 1nhash:2nhash",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

func TestMsgGovManageFeesRequest_HasUpdates(t *testing.T) {
	oneCoin := []sdk.Coin{{}}
	oneRatio := []FeeRatio{{}}

	tests := []struct {
		name string
		msg  MsgGovManageFeesRequest
		exp  bool
	}{
		{
			name: "empty",
			msg:  MsgGovManageFeesRequest{},
			exp:  false,
		},
		{
			name: "empty except for authority",
			msg: MsgGovManageFeesRequest{
				Authority: "authority",
			},
			exp: false,
		},
		{
			name: "one add fee create-ask flat",
			msg:  MsgGovManageFeesRequest{AddFeeCreateAskFlat: oneCoin},
			exp:  true,
		},
		{
			name: "one remove fee create-ask flat",
			msg:  MsgGovManageFeesRequest{RemoveFeeCreateAskFlat: oneCoin},
			exp:  true,
		},
		{
			name: "one add fee create-bid flat",
			msg:  MsgGovManageFeesRequest{AddFeeCreateBidFlat: oneCoin},
			exp:  true,
		},
		{
			name: "one remove fee create-bid flat",
			msg:  MsgGovManageFeesRequest{RemoveFeeCreateBidFlat: oneCoin},
			exp:  true,
		},
		{
			name: "one add fee settlement seller flat",
			msg:  MsgGovManageFeesRequest{AddFeeSettlementSellerFlat: oneCoin},
			exp:  true,
		},
		{
			name: "one remove fee settlement seller flat",
			msg:  MsgGovManageFeesRequest{RemoveFeeSettlementSellerFlat: oneCoin},
			exp:  true,
		},
		{
			name: "one add fee settlement seller ratio",
			msg:  MsgGovManageFeesRequest{AddFeeSettlementSellerRatios: oneRatio},
			exp:  true,
		},
		{
			name: "one remove fee settlement seller ratio",
			msg:  MsgGovManageFeesRequest{RemoveFeeSettlementSellerRatios: oneRatio},
			exp:  true,
		},
		{
			name: "one add fee settlement buyer flat",
			msg:  MsgGovManageFeesRequest{AddFeeSettlementBuyerFlat: oneCoin},
			exp:  true,
		},
		{
			name: "one remove fee settlement buyer flat",
			msg:  MsgGovManageFeesRequest{RemoveFeeSettlementBuyerFlat: oneCoin},
			exp:  true,
		},
		{
			name: "one add fee settlement buyer ratio",
			msg:  MsgGovManageFeesRequest{AddFeeSettlementBuyerRatios: oneRatio},
			exp:  true,
		},
		{
			name: "one remove fee settlement buyer ratio",
			msg:  MsgGovManageFeesRequest{RemoveFeeSettlementBuyerRatios: oneRatio},
			exp:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual bool
			testFunc := func() {
				actual = tc.msg.HasUpdates()
			}
			require.NotPanics(t, testFunc, "%T.HasUpdates()", tc.msg)
			assert.Equal(t, tc.exp, actual, "%T.HasUpdates() result", tc.msg)
		})
	}
}

func TestMsgGovUpdateParamsRequest_ValidateBasic(t *testing.T) {
	authority := sdk.AccAddress("authority___________").String()

	tests := []struct {
		name   string
		msg    MsgGovUpdateParamsRequest
		expErr []string
	}{
		{
			name:   "zero value",
			msg:    MsgGovUpdateParamsRequest{},
			expErr: []string{"invalid authority", "empty address string is not allowed"},
		},
		{
			name: "default params",
			msg: MsgGovUpdateParamsRequest{
				Authority: authority,
				Params:    *DefaultParams(),
			},
			expErr: nil,
		},
		{
			name: "control",
			msg: MsgGovUpdateParamsRequest{
				Authority: authority,
				Params: Params{
					DefaultSplit: 543,
					DenomSplits: []DenomSplit{
						{Denom: "nhash", Split: 222},
						{Denom: "nusdf", Split: 123},
						{Denom: "musdm", Split: 8},
					},
				},
			},
			expErr: nil,
		},
		{
			name: "no authority",
			msg: MsgGovUpdateParamsRequest{
				Authority: "",
				Params:    *DefaultParams(),
			},
			expErr: []string{"invalid authority", "empty address string is not allowed"},
		},
		{
			name: "bad authority",
			msg: MsgGovUpdateParamsRequest{
				Authority: "bad",
				Params:    *DefaultParams(),
			},
			expErr: []string{"invalid authority", "decoding bech32 failed"},
		},
		{
			name: "bad params",
			msg: MsgGovUpdateParamsRequest{
				Authority: authority,
				Params: Params{
					DefaultSplit: 10_123,
					DenomSplits: []DenomSplit{
						{Denom: "x", Split: 500},
						{Denom: "nhash", Split: 20_000},
					},
				},
			},
			expErr: []string{
				"default split 10123 cannot be greater than 10000",
				"nhash split 20000 cannot be greater than 10000",
			},
		},
		{
			name: "multiple errors",
			msg: MsgGovUpdateParamsRequest{
				Authority: "",
				Params:    Params{DefaultSplit: 10_555},
			},
			expErr: []string{
				"invalid authority",
				"default split 10555 cannot be greater than 10000",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}
