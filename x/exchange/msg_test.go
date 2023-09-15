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
			return &MsgCreateAskRequest{AskOrder: AskOrder{Seller: signer}}
		},
		func(signer string) sdk.Msg {
			return &MsgCreateBidRequest{BidOrder: BidOrder{Buyer: signer}}
		},
		func(signer string) sdk.Msg {
			return &MsgCancelOrderRequest{Owner: signer}
		},
		// TODO[1658]: Add MsgFillBidsRequest once it's actually been defined.
		// TODO[1658]: Add MsgFillAsksRequest once it's actually been defined.
		// TODO[1658]: Add MsgMarketSettleRequest once it's actually been defined.
		func(signer string) sdk.Msg {
			return &MsgMarketWithdrawRequest{Administrator: signer}
		},
		// TODO[1658]: Add MsgMarketUpdateDetailsRequest once it's actually been defined.
		// TODO[1658]: Add MsgMarketUpdateEnabledRequest once it's actually been defined.
		// TODO[1658]: Add MsgMarketUpdateUserSettleRequest once it's actually been defined.
		// TODO[1658]: Add MsgMarketManagePermissionsRequest once it's actually been defined.
		// TODO[1658]: Add MsgMarketManageReqAttrsRequest once it's actually been defined.
		func(signer string) sdk.Msg {
			return &MsgGovCreateMarketRequest{Authority: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgGovManageFeesRequest{Authority: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgGovUpdateParamsRequest{Authority: signer}
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
			t.Run(typeName, func(t *testing.T) {
				// If this fails, a maker needs to be defined above for the missing msg type.
				assert.True(t, hasMaker[typeName], "There is not a GetSigners test case for %s", typeName)
			})
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

func TestMsgCreateAskRequest_ValidateBasic(t *testing.T) {
	tests := []struct {
		name   string
		msg    MsgCreateAskRequest
		expErr []string
	}{
		{
			name: "empty ask order",
			msg:  MsgCreateAskRequest{},
			expErr: []string{
				"invalid market id: ",
				"invalid seller: ",
				"invalid price: ",
				"invalid assets: ",
			},
		},
		{
			name: "invalid fees",
			msg: MsgCreateAskRequest{
				AskOrder: AskOrder{
					MarketId: 1,
					Seller:   sdk.AccAddress("seller______________").String(),
					Assets:   sdk.NewCoins(sdk.NewInt64Coin("banana", 99)),
					Price:    sdk.NewInt64Coin("acorn", 12),
				},
				OrderCreationFee: &sdk.Coin{Denom: "cactus", Amount: sdkmath.NewInt(-3)},
			},
			expErr: []string{"invalid order creation fee: negative coin amount: -3"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

func TestMsgCreateBidRequest_ValidateBasic(t *testing.T) {
	tests := []struct {
		name   string
		msg    MsgCreateBidRequest
		expErr []string
	}{
		{
			name: "empty ask order",
			msg:  MsgCreateBidRequest{},
			expErr: []string{
				"invalid market id: ",
				"invalid buyer: ",
				"invalid price: ",
				"invalid assets: ",
			},
		},
		{
			name: "invalid fees",
			msg: MsgCreateBidRequest{
				BidOrder: BidOrder{
					MarketId: 1,
					Buyer:    sdk.AccAddress("buyer_______________").String(),
					Assets:   sdk.NewCoins(sdk.NewInt64Coin("banana", 99)),
					Price:    sdk.NewInt64Coin("acorn", 12),
				},
				OrderCreationFee: &sdk.Coin{Denom: "cactus", Amount: sdkmath.NewInt(-3)},
			},
			expErr: []string{"invalid order creation fee: negative coin amount: -3"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

func TestMsgCancelOrderRequest_ValidateBasic(t *testing.T) {
	tests := []struct {
		name   string
		msg    MsgCancelOrderRequest
		expErr []string
	}{
		{
			name: "control",
			msg: MsgCancelOrderRequest{
				Owner:   sdk.AccAddress("owner_______________").String(),
				OrderId: 1,
			},
			expErr: nil,
		},
		{
			name: "missing owner",
			msg: MsgCancelOrderRequest{
				Owner:   "",
				OrderId: 1,
			},
			expErr: []string{"invalid owner: ", "empty address string is not allowed"},
		},
		{
			name: "invalid owner",
			msg: MsgCancelOrderRequest{
				Owner:   "notgonnawork",
				OrderId: 1,
			},
			expErr: []string{"invalid owner: ", "decoding bech32 failed: invalid separator index -1"},
		},
		{
			name: "order 0",
			msg: MsgCancelOrderRequest{
				Owner:   sdk.AccAddress("valid_owner_________").String(),
				OrderId: 0,
			},
			expErr: []string{"invalid order id: cannot be zero"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

// TODO[1658]: func TestMsgFillBidsRequest_ValidateBasic(t *testing.T)

// TODO[1658]: func TestMsgFillAsksRequest_ValidateBasic(t *testing.T)

// TODO[1658]: func TestMsgMarketSettleRequest_ValidateBasic(t *testing.T)

func TestMsgMarketWithdrawRequest_ValidateBasic(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	goodAdminAddr := sdk.AccAddress("Administrator_______").String()
	goodToAddr := sdk.AccAddress("ToAddress___________").String()
	goodCoins := sdk.NewCoins(coin(3, "acorns"))

	tests := []struct {
		name   string
		msg    MsgMarketWithdrawRequest
		expErr []string
	}{
		{
			name: "control",
			msg: MsgMarketWithdrawRequest{
				Administrator: goodAdminAddr,
				MarketId:      1,
				ToAddress:     goodToAddr,
				Amount:        goodCoins,
			},
			expErr: nil,
		},
		{
			name: "no administrator",
			msg: MsgMarketWithdrawRequest{
				Administrator: "",
				MarketId:      1,
				ToAddress:     goodToAddr,
				Amount:        goodCoins,
			},
			expErr: []string{`invalid administrator ""`, "empty address string is not allowed"},
		},
		{
			name: "bad administrator",
			msg: MsgMarketWithdrawRequest{
				Administrator: "notright",
				MarketId:      1,
				ToAddress:     goodToAddr,
				Amount:        goodCoins,
			},
			expErr: []string{`invalid administrator "notright"`, "decoding bech32 failed"},
		},
		{
			name: "market id zero",
			msg: MsgMarketWithdrawRequest{
				Administrator: goodAdminAddr,
				MarketId:      0,
				ToAddress:     goodToAddr,
				Amount:        goodCoins,
			},
			expErr: []string{"invalid market id", "cannot be zero"},
		},
		{
			name: "no to address",
			msg: MsgMarketWithdrawRequest{
				Administrator: goodAdminAddr,
				MarketId:      1,
				ToAddress:     "",
				Amount:        goodCoins,
			},
			expErr: []string{`invalid to address ""`, "empty address string is not allowed"},
		},
		{
			name: "bad to address",
			msg: MsgMarketWithdrawRequest{
				Administrator: goodAdminAddr,
				MarketId:      1,
				ToAddress:     "notright",
				Amount:        goodCoins,
			},
			expErr: []string{`invalid to address "notright"`, "decoding bech32 failed"},
		},
		{
			name: "invalid denom in amount",
			msg: MsgMarketWithdrawRequest{
				Administrator: goodAdminAddr,
				MarketId:      1,
				ToAddress:     goodToAddr,
				Amount:        sdk.Coins{coin(3, "x")},
			},
			expErr: []string{`invalid amount "3x"`, "invalid denom: x"},
		},
		{
			name: "negative amount",
			msg: MsgMarketWithdrawRequest{
				Administrator: goodAdminAddr,
				MarketId:      1,
				ToAddress:     goodToAddr,
				Amount:        sdk.Coins{coin(-1, "negcoin")},
			},
			expErr: []string{`invalid amount "-1negcoin"`, "coin -1negcoin amount is not positive"},
		},
		{
			name: "empty amount",
			msg: MsgMarketWithdrawRequest{
				Administrator: goodAdminAddr,
				MarketId:      1,
				ToAddress:     goodToAddr,
				Amount:        sdk.Coins{},
			},
			expErr: []string{`invalid amount ""`, "cannot be zero"},
		},
		{
			name: "zero coin amount",
			msg: MsgMarketWithdrawRequest{
				Administrator: goodAdminAddr,
				MarketId:      1,
				ToAddress:     goodToAddr,
				Amount:        sdk.Coins{coin(0, "acorn"), coin(0, "banana")},
			},
			expErr: []string{`invalid amount "0acorn,0banana"`, "coin 0acorn amount is not positive"},
		},
		{
			name: "multiple errors",
			msg:  MsgMarketWithdrawRequest{},
			expErr: []string{
				"invalid administrator",
				"invalid market id",
				"invalid to address",
				"invalid amount",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

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
		FeeSellerSettlementFlat: []sdk.Coin{sdk.NewInt64Coin("nhash", 50)},
		FeeSellerSettlementRatios: []FeeRatio{
			{Price: sdk.NewInt64Coin("nhash", 100), Fee: sdk.NewInt64Coin("nhash", 1)},
		},
		FeeBuyerSettlementFlat: []sdk.Coin{sdk.NewInt64Coin("nhash", 100)},
		FeeBuyerSettlementRatios: []FeeRatio{
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
			name: "invalid add seller settlement flat",
			msg: MsgGovManageFeesRequest{
				Authority:                  authority,
				AddFeeSellerSettlementFlat: []sdk.Coin{coin(0, "nhash")},
			},
			expErr: []string{`invalid seller settlement flat fee to add option "0nhash": amount cannot be zero`},
		},
		{
			name: "same add and remove seller settlement flat",
			msg: MsgGovManageFeesRequest{
				Authority:                     authority,
				AddFeeSellerSettlementFlat:    []sdk.Coin{coin(1, "nhash")},
				RemoveFeeSellerSettlementFlat: []sdk.Coin{coin(1, "nhash")},
			},
			expErr: []string{"cannot add and remove the same seller settlement flat fee options: 1nhash"},
		},
		{
			name: "invalid add seller settlement ratio",
			msg: MsgGovManageFeesRequest{
				Authority:                    authority,
				AddFeeSellerSettlementRatios: []FeeRatio{ratio(1, "nhash", 2, "nhash")},
			},
			expErr: []string{`seller fee ratio fee amount "2nhash" cannot be greater than price amount "1nhash"`},
		},
		{
			name: "same add and remove seller settlement ratio",
			msg: MsgGovManageFeesRequest{
				Authority:                       authority,
				AddFeeSellerSettlementRatios:    []FeeRatio{ratio(2, "nhash", 1, "nhash")},
				RemoveFeeSellerSettlementRatios: []FeeRatio{ratio(2, "nhash", 1, "nhash")},
			},
			expErr: []string{"cannot add and remove the same seller settlement fee ratios: 2nhash:1nhash"},
		},
		{
			name: "invalid add buyer settlement flat",
			msg: MsgGovManageFeesRequest{
				Authority:                 authority,
				AddFeeBuyerSettlementFlat: []sdk.Coin{coin(0, "nhash")},
			},
			expErr: []string{`invalid buyer settlement flat fee to add option "0nhash": amount cannot be zero`},
		},
		{
			name: "same add and remove buyer settlement flat",
			msg: MsgGovManageFeesRequest{
				Authority:                    authority,
				AddFeeBuyerSettlementFlat:    []sdk.Coin{coin(1, "nhash")},
				RemoveFeeBuyerSettlementFlat: []sdk.Coin{coin(1, "nhash")},
			},
			expErr: []string{"cannot add and remove the same buyer settlement flat fee options: 1nhash"},
		},
		{
			name: "invalid add buyer settlement ratio",
			msg: MsgGovManageFeesRequest{
				Authority:                   authority,
				AddFeeBuyerSettlementRatios: []FeeRatio{ratio(1, "nhash", 2, "nhash")},
			},
			expErr: []string{`buyer fee ratio fee amount "2nhash" cannot be greater than price amount "1nhash"`},
		},
		{
			name: "same add and remove buyer settlement ratio",
			msg: MsgGovManageFeesRequest{
				Authority:                      authority,
				AddFeeBuyerSettlementRatios:    []FeeRatio{ratio(2, "nhash", 1, "nhash")},
				RemoveFeeBuyerSettlementRatios: []FeeRatio{ratio(2, "nhash", 1, "nhash")},
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
				AddFeeSellerSettlementFlat:      []sdk.Coin{coin(0, "nhash")},
				RemoveFeeSellerSettlementFlat:   []sdk.Coin{coin(0, "nhash")},
				AddFeeSellerSettlementRatios:    []FeeRatio{ratio(1, "nhash", 2, "nhash")},
				RemoveFeeSellerSettlementRatios: []FeeRatio{ratio(1, "nhash", 2, "nhash")},
				AddFeeBuyerSettlementFlat:       []sdk.Coin{coin(0, "nhash")},
				RemoveFeeBuyerSettlementFlat:    []sdk.Coin{coin(0, "nhash")},
				AddFeeBuyerSettlementRatios:     []FeeRatio{ratio(1, "nhash", 2, "nhash")},
				RemoveFeeBuyerSettlementRatios:  []FeeRatio{ratio(1, "nhash", 2, "nhash")},
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
			name: "one add fee seller settlement flat",
			msg:  MsgGovManageFeesRequest{AddFeeSellerSettlementFlat: oneCoin},
			exp:  true,
		},
		{
			name: "one remove fee seller settlement flat",
			msg:  MsgGovManageFeesRequest{RemoveFeeSellerSettlementFlat: oneCoin},
			exp:  true,
		},
		{
			name: "one add fee seller settlement ratio",
			msg:  MsgGovManageFeesRequest{AddFeeSellerSettlementRatios: oneRatio},
			exp:  true,
		},
		{
			name: "one remove fee seller settlement ratio",
			msg:  MsgGovManageFeesRequest{RemoveFeeSellerSettlementRatios: oneRatio},
			exp:  true,
		},
		{
			name: "one add fee buyer settlement flat",
			msg:  MsgGovManageFeesRequest{AddFeeBuyerSettlementFlat: oneCoin},
			exp:  true,
		},
		{
			name: "one remove fee buyer settlement flat",
			msg:  MsgGovManageFeesRequest{RemoveFeeBuyerSettlementFlat: oneCoin},
			exp:  true,
		},
		{
			name: "one add fee buyer settlement ratio",
			msg:  MsgGovManageFeesRequest{AddFeeBuyerSettlementRatios: oneRatio},
			exp:  true,
		},
		{
			name: "one remove fee buyer settlement ratio",
			msg:  MsgGovManageFeesRequest{RemoveFeeBuyerSettlementRatios: oneRatio},
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
