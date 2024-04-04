package exchange

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/internal/helpers"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil/assertions"
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
			return &MsgCreateAskRequest{AskOrder: AskOrder{Seller: signer}}
		},
		func(signer string) sdk.Msg {
			return &MsgCreateBidRequest{BidOrder: BidOrder{Buyer: signer}}
		},
		func(signer string) sdk.Msg {
			return &MsgCommitFundsRequest{Account: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgCancelOrderRequest{Signer: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgFillBidsRequest{Seller: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgFillAsksRequest{Buyer: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgMarketSettleRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgMarketCommitmentSettleRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgMarketReleaseCommitmentsRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgMarketSetOrderExternalIDRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgMarketWithdrawRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgMarketUpdateDetailsRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgMarketUpdateEnabledRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgMarketUpdateAcceptingOrdersRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgMarketUpdateUserSettleRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgMarketUpdateAcceptingCommitmentsRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgMarketUpdateIntermediaryDenomRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgMarketManagePermissionsRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgMarketManageReqAttrsRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgCreatePaymentRequest{Payment: Payment{Source: signer}}
		},
		func(signer string) sdk.Msg {
			return &MsgAcceptPaymentRequest{Payment: Payment{Target: signer}}
		},
		func(signer string) sdk.Msg {
			return &MsgRejectPaymentRequest{Target: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgRejectPaymentsRequest{Target: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgCancelPaymentsRequest{Source: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgChangePaymentTargetRequest{Source: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgGovCreateMarketRequest{Authority: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgGovManageFeesRequest{Authority: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgGovCloseMarketRequest{Authority: signer}
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
			smsg, ok := tc.msg.(HasGetSigners)
			require.True(t, ok, "%T does not have a .GetSigners method.")

			var signers []sdk.AccAddress
			testFunc := func() {
				signers = smsg.GetSigners()
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
		err = helpers.ValidateBasic(msg)
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
					Assets:   sdk.NewInt64Coin("banana", 99),
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
			name: "empty bid order",
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
					Assets:   sdk.NewInt64Coin("banana", 99),
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

func TestMsgCommitFundsRequest_ValidateBasic(t *testing.T) {
	tests := []struct {
		name   string
		msg    MsgCommitFundsRequest
		expErr []string
	}{
		{
			name: "okay",
			msg: MsgCommitFundsRequest{
				Account:  sdk.AccAddress("account_____________").String(),
				MarketId: 1,
				Amount:   sdk.Coins{sdk.NewInt64Coin("cherry", 52)},
			},
			expErr: nil,
		},
		{
			name: "okay with optional fields",
			msg: MsgCommitFundsRequest{
				Account:     sdk.AccAddress("account_____________").String(),
				MarketId:    3,
				Amount:      sdk.Coins{sdk.NewInt64Coin("cherry", 52)},
				CreationFee: &sdk.Coin{Denom: "fig", Amount: sdkmath.NewInt(8)},
				EventTag:    "just-some-tag",
			},
			expErr: nil,
		},
		{
			name: "no account",
			msg: MsgCommitFundsRequest{
				Account:  "",
				MarketId: 1,
				Amount:   sdk.Coins{sdk.NewInt64Coin("cherry", 52)},
			},
			expErr: []string{"invalid account \"\": " + emptyAddrErr},
		},
		{
			name: "bad account",
			msg: MsgCommitFundsRequest{
				Account:  "badaccountstring",
				MarketId: 1,
				Amount:   sdk.Coins{sdk.NewInt64Coin("cherry", 52)},
			},
			expErr: []string{"invalid account \"badaccountstring\": " + bech32Err},
		},
		{
			name: "market zero",
			msg: MsgCommitFundsRequest{
				Account:  sdk.AccAddress("account_____________").String(),
				MarketId: 0,
				Amount:   sdk.Coins{sdk.NewInt64Coin("cherry", 52)},
			},
			expErr: []string{"invalid market id: cannot be zero"},
		},
		{
			name: "nil amount",
			msg: MsgCommitFundsRequest{
				Account:  sdk.AccAddress("account_____________").String(),
				MarketId: 1,
				Amount:   nil,
			},
			expErr: []string{"invalid amount \"\": cannot be zero"},
		},
		{
			name: "empty amount",
			msg: MsgCommitFundsRequest{
				Account:  sdk.AccAddress("account_____________").String(),
				MarketId: 1,
				Amount:   sdk.Coins{},
			},
			expErr: []string{"invalid amount \"\": cannot be zero"},
		},
		{
			name: "bad amount",
			msg: MsgCommitFundsRequest{
				Account:  sdk.AccAddress("account_____________").String(),
				MarketId: 1,
				Amount:   sdk.Coins{sdk.Coin{Denom: "cherry", Amount: sdkmath.NewInt(-3)}},
			},
			expErr: []string{"invalid amount \"-3cherry\": coin -3cherry amount is not positive"},
		},
		{
			name: "bad creation fee",
			msg: MsgCommitFundsRequest{
				Account:     sdk.AccAddress("account_____________").String(),
				MarketId:    1,
				Amount:      sdk.Coins{sdk.NewInt64Coin("cherry", 52)},
				CreationFee: &sdk.Coin{Denom: "fig", Amount: sdkmath.NewInt(-1)},
			},
			expErr: []string{"invalid creation fee \"-1fig\": negative coin amount: -1"},
		},
		{
			name: "bad event tag",
			msg: MsgCommitFundsRequest{
				Account:  sdk.AccAddress("account_____________").String(),
				MarketId: 1,
				Amount:   sdk.Coins{sdk.NewInt64Coin("cherry", 52)},
				EventTag: strings.Repeat("p", 100) + "x",
			},
			expErr: []string{"invalid event tag \"ppppp...ppppx\" (length 101): exceeds max length 100"},
		},
		{
			name: "multiple errors",
			msg: MsgCommitFundsRequest{
				CreationFee: &sdk.Coin{Denom: "fig", Amount: sdkmath.NewInt(-1)},
				EventTag:    strings.Repeat("p", 100) + "x",
			},
			expErr: []string{
				"invalid account \"\": " + emptyAddrErr,
				"invalid market id: cannot be zero",
				"invalid amount \"\": cannot be zero",
				"invalid creation fee \"-1fig\": negative coin amount: -1",
				"invalid event tag \"ppppp...ppppx\" (length 101): exceeds max length 100",
			},
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
				Signer:  sdk.AccAddress("signer______________").String(),
				OrderId: 1,
			},
			expErr: nil,
		},
		{
			name: "missing owner",
			msg: MsgCancelOrderRequest{
				Signer:  "",
				OrderId: 1,
			},
			expErr: []string{"invalid signer: ", emptyAddrErr},
		},
		{
			name: "invalid owner",
			msg: MsgCancelOrderRequest{
				Signer:  "notgonnawork",
				OrderId: 1,
			},
			expErr: []string{"invalid signer: ", bech32Err + "invalid separator index -1"},
		},
		{
			name: "order 0",
			msg: MsgCancelOrderRequest{
				Signer:  sdk.AccAddress("valid_signer________").String(),
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

func TestMsgFillBidsRequest_ValidateBasic(t *testing.T) {
	coin := func(amount int64, denom string) *sdk.Coin {
		return &sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	seller := sdk.AccAddress("seller______________").String()

	tests := []struct {
		name   string
		msg    MsgFillBidsRequest
		expErr []string
	}{
		{
			name: "control",
			msg: MsgFillBidsRequest{
				Seller:                  seller,
				MarketId:                1,
				TotalAssets:             sdk.Coins{*coin(3, "acorn")},
				BidOrderIds:             []uint64{1, 2, 3},
				SellerSettlementFlatFee: coin(2, "banana"),
				AskOrderCreationFee:     coin(8, "cactus"),
			},
			expErr: nil,
		},
		{
			name: "empty seller",
			msg: MsgFillBidsRequest{
				Seller:      "",
				MarketId:    1,
				TotalAssets: sdk.Coins{*coin(3, "acorn")},
				BidOrderIds: []uint64{1},
			},
			expErr: []string{"invalid seller", emptyAddrErr},
		},
		{
			name: "bad seller",
			msg: MsgFillBidsRequest{
				Seller:      "not-an-address",
				MarketId:    1,
				TotalAssets: sdk.Coins{*coin(3, "acorn")},
				BidOrderIds: []uint64{1},
			},
			expErr: []string{"invalid seller", bech32Err},
		},
		{
			name: "market id zero",
			msg: MsgFillBidsRequest{
				Seller:      seller,
				MarketId:    0,
				TotalAssets: sdk.Coins{*coin(3, "acorn")},
				BidOrderIds: []uint64{1},
			},
			expErr: []string{"invalid market id", "cannot be zero"},
		},
		{
			name: "nil total assets",
			msg: MsgFillBidsRequest{
				Seller:      seller,
				MarketId:    1,
				TotalAssets: nil,
				BidOrderIds: []uint64{1},
			},
			expErr: []string{"invalid total assets", "cannot be zero"},
		},
		{
			name: "empty total assets",
			msg: MsgFillBidsRequest{
				Seller:      seller,
				MarketId:    1,
				TotalAssets: sdk.Coins{},
				BidOrderIds: []uint64{1},
			},
			expErr: []string{"invalid total assets", "cannot be zero"},
		},
		{
			name: "invalid total assets",
			msg: MsgFillBidsRequest{
				Seller:      seller,
				MarketId:    1,
				TotalAssets: sdk.Coins{*coin(-1, "acorn")},
				BidOrderIds: []uint64{1},
			},
			expErr: []string{"invalid total assets", "coin -1acorn amount is not positive"},
		},
		{
			name: "nil order ids",
			msg: MsgFillBidsRequest{
				Seller:      seller,
				MarketId:    1,
				TotalAssets: sdk.Coins{*coin(1, "acorn")},
				BidOrderIds: nil,
			},
			expErr: []string{"no bid order ids provided"},
		},
		{
			name: "order id zero",
			msg: MsgFillBidsRequest{
				Seller:      seller,
				MarketId:    1,
				TotalAssets: sdk.Coins{*coin(1, "acorn")},
				BidOrderIds: []uint64{0},
			},
			expErr: []string{"invalid bid order ids: cannot contain order id zero"},
		},
		{
			name: "duplicate order ids",
			msg: MsgFillBidsRequest{
				Seller:      seller,
				MarketId:    1,
				TotalAssets: sdk.Coins{*coin(1, "acorn")},
				BidOrderIds: []uint64{1, 2, 1},
			},
			expErr: []string{"duplicate bid order ids provided: [1]"},
		},
		{
			name: "invalid seller settlement flat fee",
			msg: MsgFillBidsRequest{
				Seller:                  seller,
				MarketId:                1,
				TotalAssets:             sdk.Coins{*coin(1, "acorn")},
				BidOrderIds:             []uint64{1},
				SellerSettlementFlatFee: coin(-1, "catan"),
			},
			expErr: []string{"invalid seller settlement flat fee", "negative coin amount: -1"},
		},
		{
			name: "seller settlement flat fee with zero amount",
			msg: MsgFillBidsRequest{
				Seller:                  seller,
				MarketId:                1,
				TotalAssets:             sdk.Coins{*coin(1, "acorn")},
				BidOrderIds:             []uint64{1},
				SellerSettlementFlatFee: coin(0, "catan"),
			},
			expErr: []string{"invalid seller settlement flat fee", "catan amount cannot be zero"},
		},
		{
			name: "invalid order creation fee",
			msg: MsgFillBidsRequest{
				Seller:              seller,
				MarketId:            1,
				TotalAssets:         sdk.Coins{*coin(1, "acorn")},
				BidOrderIds:         []uint64{1},
				AskOrderCreationFee: coin(-1, "catan"),
			},
			expErr: []string{"invalid ask order creation fee", "negative coin amount: -1"},
		},
		{
			name: "order creation fee with zero amount",
			msg: MsgFillBidsRequest{
				Seller:              seller,
				MarketId:            1,
				TotalAssets:         sdk.Coins{*coin(1, "acorn")},
				BidOrderIds:         []uint64{1},
				AskOrderCreationFee: coin(0, "catan"),
			},
			expErr: []string{"invalid ask order creation fee", "catan amount cannot be zero"},
		},
		{
			name: "multiple errors",
			msg: MsgFillBidsRequest{
				Seller:                  "",
				MarketId:                0,
				TotalAssets:             nil,
				BidOrderIds:             nil,
				SellerSettlementFlatFee: coin(0, "catan"),
				AskOrderCreationFee:     coin(-1, "catan"),
			},
			expErr: []string{
				"invalid seller",
				"invalid market id",
				"invalid total assets",
				"no bid order ids provided",
				"invalid seller settlement flat fee",
				"invalid ask order creation fee",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

func TestMsgFillAsksRequest_ValidateBasic(t *testing.T) {
	coin := func(amount int64, denom string) *sdk.Coin {
		return &sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	buyer := sdk.AccAddress("buyer_______________").String()

	tests := []struct {
		name   string
		msg    MsgFillAsksRequest
		expErr []string
	}{
		{
			name: "control",
			msg: MsgFillAsksRequest{
				Buyer:               buyer,
				MarketId:            1,
				TotalPrice:          *coin(3, "acorn"),
				AskOrderIds:         []uint64{1, 2, 3},
				BuyerSettlementFees: sdk.Coins{*coin(2, "banana")},
				BidOrderCreationFee: coin(8, "cactus"),
			},
			expErr: nil,
		},
		{
			name: "empty buyer",
			msg: MsgFillAsksRequest{
				Buyer:       "",
				MarketId:    1,
				TotalPrice:  *coin(3, "acorn"),
				AskOrderIds: []uint64{1},
			},
			expErr: []string{"invalid buyer", emptyAddrErr},
		},
		{
			name: "bad buyer",
			msg: MsgFillAsksRequest{
				Buyer:       "not-an-address",
				MarketId:    1,
				TotalPrice:  *coin(3, "acorn"),
				AskOrderIds: []uint64{1},
			},
			expErr: []string{"invalid buyer", bech32Err},
		},
		{
			name: "market id zero",
			msg: MsgFillAsksRequest{
				Buyer:       buyer,
				MarketId:    0,
				TotalPrice:  *coin(3, "acorn"),
				AskOrderIds: []uint64{1},
			},
			expErr: []string{"invalid market id", "cannot be zero"},
		},
		{
			name: "invalid total price",
			msg: MsgFillAsksRequest{
				Buyer:       buyer,
				MarketId:    1,
				TotalPrice:  *coin(-1, "acorn"),
				AskOrderIds: []uint64{1},
			},
			expErr: []string{"invalid total price", "negative coin amount: -1"},
		},
		{
			name: "nil order ids",
			msg: MsgFillAsksRequest{
				Buyer:       buyer,
				MarketId:    1,
				TotalPrice:  *coin(1, "acorn"),
				AskOrderIds: nil,
			},
			expErr: []string{"no ask order ids provided"},
		},
		{
			name: "order id zero",
			msg: MsgFillAsksRequest{
				Buyer:       buyer,
				MarketId:    1,
				TotalPrice:  *coin(1, "acorn"),
				AskOrderIds: []uint64{0},
			},
			expErr: []string{"invalid ask order ids: cannot contain order id zero"},
		},
		{
			name: "duplicate order ids",
			msg: MsgFillAsksRequest{
				Buyer:       buyer,
				MarketId:    1,
				TotalPrice:  *coin(1, "acorn"),
				AskOrderIds: []uint64{1, 2, 1},
			},
			expErr: []string{"duplicate ask order ids provided: [1]"},
		},
		{
			name: "invalid buyer settlement flat fee",
			msg: MsgFillAsksRequest{
				Buyer:               buyer,
				MarketId:            1,
				TotalPrice:          *coin(1, "acorn"),
				AskOrderIds:         []uint64{1},
				BuyerSettlementFees: sdk.Coins{*coin(-1, "catan")},
			},
			expErr: []string{"invalid buyer settlement fees", "coin -1catan amount is not positive"},
		},
		{
			name: "buyer settlement flat fee with zero amount",
			msg: MsgFillAsksRequest{
				Buyer:               buyer,
				MarketId:            1,
				TotalPrice:          *coin(1, "acorn"),
				AskOrderIds:         []uint64{1},
				BuyerSettlementFees: sdk.Coins{*coin(0, "catan")},
			},
			expErr: []string{"invalid buyer settlement fees", "coin 0catan amount is not positive"},
		},
		{
			name: "invalid order creation fee",
			msg: MsgFillAsksRequest{
				Buyer:               buyer,
				MarketId:            1,
				TotalPrice:          *coin(1, "acorn"),
				AskOrderIds:         []uint64{1},
				BidOrderCreationFee: coin(-1, "catan"),
			},
			expErr: []string{"invalid bid order creation fee", "negative coin amount: -1"},
		},
		{
			name: "order creation fee with zero amount",
			msg: MsgFillAsksRequest{
				Buyer:               buyer,
				MarketId:            1,
				TotalPrice:          *coin(1, "acorn"),
				AskOrderIds:         []uint64{1},
				BidOrderCreationFee: coin(0, "catan"),
			},
			expErr: []string{"invalid bid order creation fee", "catan amount cannot be zero"},
		},
		{
			name: "multiple errors",
			msg: MsgFillAsksRequest{
				Buyer:               "",
				MarketId:            0,
				TotalPrice:          sdk.Coin{},
				AskOrderIds:         nil,
				BuyerSettlementFees: sdk.Coins{*coin(0, "catan")},
				BidOrderCreationFee: coin(-1, "catan"),
			},
			expErr: []string{
				"invalid buyer",
				"invalid market id",
				"invalid total price",
				"no ask order ids provided",
				"invalid buyer settlement fees",
				"invalid bid order creation fee",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

func TestMsgMarketSettleRequest_ValidateBasic(t *testing.T) {
	admin := sdk.AccAddress("admin_address_______").String()

	tests := []struct {
		name   string
		msg    MsgMarketSettleRequest
		expErr []string
	}{
		{
			name: "control",
			msg: MsgMarketSettleRequest{
				Admin:         admin,
				MarketId:      1,
				AskOrderIds:   []uint64{1, 3, 5},
				BidOrderIds:   []uint64{2, 4, 6},
				ExpectPartial: true,
			},
			expErr: nil,
		},
		{
			name: "no admin",
			msg: MsgMarketSettleRequest{
				Admin:       "",
				MarketId:    1,
				AskOrderIds: []uint64{1},
				BidOrderIds: []uint64{2},
			},
			expErr: []string{`invalid administrator ""`, emptyAddrErr},
		},
		{
			name: "bad admin",
			msg: MsgMarketSettleRequest{
				Admin:       "badbadadmin",
				MarketId:    1,
				AskOrderIds: []uint64{1},
				BidOrderIds: []uint64{2},
			},
			expErr: []string{`invalid administrator "badbadadmin"`, bech32Err},
		},
		{
			name: "market id zero",
			msg: MsgMarketSettleRequest{
				Admin:       admin,
				MarketId:    0,
				AskOrderIds: []uint64{1},
				BidOrderIds: []uint64{2},
			},
			expErr: []string{"invalid market id", "cannot be zero"},
		},
		{
			name: "nil ask orders",
			msg: MsgMarketSettleRequest{
				Admin:       admin,
				MarketId:    1,
				AskOrderIds: nil,
				BidOrderIds: []uint64{2},
			},
			expErr: []string{"no ask order ids provided"},
		},
		{
			name: "ask orders has a zero",
			msg: MsgMarketSettleRequest{
				Admin:       admin,
				MarketId:    1,
				AskOrderIds: []uint64{1, 3, 0, 5},
				BidOrderIds: []uint64{2},
			},
			expErr: []string{"invalid ask order ids", "cannot contain order id zero"},
		},
		{
			name: "duplicate ask orders ids",
			msg: MsgMarketSettleRequest{
				Admin:       admin,
				MarketId:    1,
				AskOrderIds: []uint64{1, 3, 1, 5, 5},
				BidOrderIds: []uint64{2},
			},
			expErr: []string{"duplicate ask order ids provided: [1 5]"},
		},
		{
			name: "nil bid orders",
			msg: MsgMarketSettleRequest{
				Admin:       admin,
				MarketId:    1,
				AskOrderIds: []uint64{1},
				BidOrderIds: nil,
			},
			expErr: []string{"no bid order ids provided"},
		},
		{
			name: "bid orders has a zero",
			msg: MsgMarketSettleRequest{
				Admin:       admin,
				MarketId:    1,
				AskOrderIds: []uint64{1},
				BidOrderIds: []uint64{2, 0, 4, 6},
			},
			expErr: []string{"invalid bid order ids", "cannot contain order id zero"},
		},
		{
			name: "duplicate bid orders ids",
			msg: MsgMarketSettleRequest{
				Admin:       admin,
				MarketId:    1,
				AskOrderIds: []uint64{1},
				BidOrderIds: []uint64{2, 4, 2, 6, 6},
			},
			expErr: []string{"duplicate bid order ids provided: [2 6]"},
		},
		{
			name: "same orders in both lists",
			msg: MsgMarketSettleRequest{
				Admin:       admin,
				MarketId:    1,
				AskOrderIds: []uint64{1, 3, 5, 6},
				BidOrderIds: []uint64{2, 4, 6, 3},
			},
			expErr: []string{"order ids duplicated as both bid and ask: [3 6]"},
		},
		{
			name: "multiple errors",
			msg: MsgMarketSettleRequest{
				Admin:       "",
				MarketId:    0,
				AskOrderIds: nil,
				BidOrderIds: nil,
			},
			expErr: []string{
				"invalid administrator",
				"invalid market id",
				"no ask order ids provided",
				"no bid order ids provided",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

func TestMsgMarketCommitmentSettleRequest_ValidateBasic(t *testing.T) {
	type testCase struct {
		name         string
		msg          MsgMarketCommitmentSettleRequest
		expErrInReq  []string // for errors that only happen when requireInputs = true.
		expErrInOpt  []string // for errors that only happen when requireInputs = false.
		expErrAlways []string // for errors that happen regardless of the requireInputs value.
	}
	getExpErr := func(tc testCase, requireInputs bool) []string {
		if requireInputs {
			return append(tc.expErrAlways, tc.expErrInReq...)
		}
		return append(tc.expErrAlways, tc.expErrInOpt...)
	}
	toAccAddr := func(str string) string {
		return sdk.AccAddress(str + strings.Repeat("_", 20-len(str))).String()
	}
	goodAA := func(account, amount string) AccountAmount {
		addr := toAccAddr(account)
		amt, err := sdk.ParseCoinsNormalized(amount)
		require.NoError(t, err, "ParseCoinsNormalized(%q)", amount)
		return AccountAmount{Account: addr, Amount: amt}
	}
	goodNAV := func(assets, price string) NetAssetPrice {
		rv := NetAssetPrice{}
		var err error
		rv.Assets, err = sdk.ParseCoinNormalized(assets)
		require.NoError(t, err, "ParseCoinNormalized(%q) (assets)", assets)
		rv.Price, err = sdk.ParseCoinNormalized(price)
		require.NoError(t, err, "ParseCoinNormalized(%q) (price)", price)
		return rv
	}
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	coins := func(amount int64, denom string) sdk.Coins {
		return sdk.Coins{{Denom: denom, Amount: sdkmath.NewInt(amount)}}
	}

	tests := []testCase{
		{
			name: "control",
			msg: MsgMarketCommitmentSettleRequest{
				Admin:    toAccAddr("admin"),
				MarketId: 1,
				Inputs:   []AccountAmount{goodAA("input0", "13cherry")},
				Outputs:  []AccountAmount{goodAA("output0", "13cherry")},
			},
		},
		{
			name: "okay: with all optional fields",
			msg: MsgMarketCommitmentSettleRequest{
				Admin:    toAccAddr("admin"),
				MarketId: 1,
				Inputs:   []AccountAmount{goodAA("input0", "13cherry")},
				Outputs:  []AccountAmount{goodAA("output0", "13cherry")},
				Fees:     []AccountAmount{goodAA("fee0", "7fig")},
				Navs:     []NetAssetPrice{goodNAV("13cherry", "1990musd")},
				EventTag: "you're it",
			},
		},
		{
			name: "no admin",
			msg: MsgMarketCommitmentSettleRequest{
				Admin:    "",
				MarketId: 1,
				Inputs:   []AccountAmount{goodAA("input0", "13cherry")},
				Outputs:  []AccountAmount{goodAA("output0", "13cherry")},
			},
			expErrInReq: []string{"invalid administrator \"\": " + emptyAddrErr},
		},
		{
			name: "bad admin",
			msg: MsgMarketCommitmentSettleRequest{
				Admin:    "badbadadmin",
				MarketId: 1,
				Inputs:   []AccountAmount{goodAA("input0", "13cherry")},
				Outputs:  []AccountAmount{goodAA("output0", "13cherry")},
			},
			expErrInReq: []string{"invalid administrator \"badbadadmin\": " + bech32Err},
		},
		{
			name: "market zero",
			msg: MsgMarketCommitmentSettleRequest{
				Admin:    toAccAddr("admin"),
				MarketId: 0,
				Inputs:   []AccountAmount{goodAA("input0", "13cherry")},
				Outputs:  []AccountAmount{goodAA("output0", "13cherry")},
			},
			expErrAlways: []string{"invalid market id: cannot be zero"},
		},
		{
			name: "no inputs",
			msg: MsgMarketCommitmentSettleRequest{
				Admin:    toAccAddr("admin"),
				MarketId: 1,
				Outputs:  []AccountAmount{goodAA("output0", "13cherry")},
			},
			expErrInReq: []string{"no inputs provided"},
			expErrInOpt: []string{"input total \"\" does not equal output total \"13cherry\""},
		},
		{
			name: "bad inputs",
			msg: MsgMarketCommitmentSettleRequest{
				Admin:    toAccAddr("admin"),
				MarketId: 1,
				Inputs: []AccountAmount{
					{Account: toAccAddr("input0"), Amount: coins(-3, "cherry")},
					goodAA("input1", "8cherry"),
					{Account: "badinput2addr", Amount: coins(4, "cherry")},
					{Account: toAccAddr("input3"), Amount: nil},
				},
			},
			expErrInReq: []string{"no outputs provided"},
			expErrAlways: []string{
				"inputs[0]: invalid amount \"-3cherry\": coin -3cherry amount is not positive",
				"inputs[2]: invalid account \"badinput2addr\": " + bech32Err,
				"inputs[3]: invalid amount \"\": cannot be zero",
			},
		},
		{
			name: "no outputs",
			msg: MsgMarketCommitmentSettleRequest{
				Admin:    toAccAddr("admin"),
				MarketId: 1,
				Inputs:   []AccountAmount{goodAA("input0", "13cherry")},
			},
			expErrInReq: []string{"no outputs provided"},
			expErrInOpt: []string{"input total \"13cherry\" does not equal output total \"\""},
		},
		{
			name: "bad outputs",
			msg: MsgMarketCommitmentSettleRequest{
				Admin:    toAccAddr("admin"),
				MarketId: 1,
				Outputs: []AccountAmount{
					{Account: toAccAddr("output0"), Amount: coins(-3, "cherry")},
					goodAA("output1", "8cherry"),
					{Account: "badoutput2addr", Amount: coins(4, "cherry")},
					{Account: toAccAddr("output3"), Amount: sdk.Coins{}},
				},
			},
			expErrInReq: []string{"no inputs provided"},
			expErrAlways: []string{
				"outputs[0]: invalid amount \"-3cherry\": coin -3cherry amount is not positive",
				"outputs[2]: invalid account \"badoutput2addr\": " + bech32Err,
				"outputs[3]: invalid amount \"\": cannot be zero",
			},
		},
		{
			name: "input output amounts not equal",
			msg: MsgMarketCommitmentSettleRequest{
				Admin:    toAccAddr("admin"),
				MarketId: 1,
				Inputs:   []AccountAmount{goodAA("input0", "13cherry"), goodAA("input1", "36cherry")},
				Outputs:  []AccountAmount{goodAA("output0", "25cherry"), goodAA("output1", "23cherry")},
			},
			expErrAlways: []string{"input total \"49cherry\" does not equal output total \"48cherry\""},
		},
		{
			name: "bad fees",
			msg: MsgMarketCommitmentSettleRequest{
				Admin:    toAccAddr("admin"),
				MarketId: 1,
				Inputs:   []AccountAmount{goodAA("input0", "13cherry")},
				Outputs:  []AccountAmount{goodAA("output0", "13cherry")},
				Fees: []AccountAmount{
					{Account: "badfee0addr", Amount: coins(4, "cherry")},
					goodAA("fee1", "8cherry"),
					{Account: toAccAddr("fee2"), Amount: coins(-3, "cherry")},
					{Account: toAccAddr("fee3"), Amount: nil},
				},
			},
			expErrAlways: []string{
				"fees[0]: invalid account \"badfee0addr\": " + bech32Err,
				"fees[2]: invalid amount \"-3cherry\": coin -3cherry amount is not positive",
				"fees[3]: invalid amount \"\": cannot be zero",
			},
		},
		{
			name: "bad navs",
			msg: MsgMarketCommitmentSettleRequest{
				Admin:    toAccAddr("admin"),
				MarketId: 1,
				Inputs:   []AccountAmount{goodAA("input0", "13cherry")},
				Outputs:  []AccountAmount{goodAA("output0", "13cherry")},
				Navs: []NetAssetPrice{
					{Assets: coin(-2, "cherry"), Price: coin(87, "nhash")},
					goodNAV("13cherry", "1990musd"),
					{Assets: coin(57, "cherry"), Price: coin(52, "x")},
				},
			},
			expErrAlways: []string{
				"navs[0]: invalid assets \"-2cherry\": negative coin amount: -2",
				"navs[2]: invalid price \"52x\": invalid denom: x",
			},
		},
		{
			name: "bad event tag",
			msg: MsgMarketCommitmentSettleRequest{
				Admin:    toAccAddr("admin"),
				MarketId: 1,
				Inputs:   []AccountAmount{goodAA("input0", "13cherry")},
				Outputs:  []AccountAmount{goodAA("output0", "13cherry")},
				EventTag: "p" + strings.Repeat("b", 99) + "f",
			},
			expErrAlways: []string{"invalid event tag \"pbbbb...bbbbf\" (length 101): exceeds max length 100"},
		},
		{
			name: "multiple errors",
			msg: MsgMarketCommitmentSettleRequest{
				Admin:    "",
				MarketId: 0,
				Inputs:   []AccountAmount{{Account: toAccAddr("input0"), Amount: coins(-3, "cherry")}},
				Outputs:  []AccountAmount{{Account: toAccAddr("output0"), Amount: coins(-3, "cherry")}},
				Fees:     []AccountAmount{{Account: "badfee2addr", Amount: coins(4, "cherry")}},
				Navs:     []NetAssetPrice{{Assets: coin(-2, "cherry"), Price: coin(87, "nhash")}},
				EventTag: "p" + strings.Repeat("b", 99) + "f",
			},
			expErrInReq: []string{
				"invalid administrator \"\": " + emptyAddrErr,
			},
			expErrAlways: []string{
				"invalid market id: cannot be zero",
				"inputs[0]: invalid amount \"-3cherry\": coin -3cherry amount is not positive",
				"outputs[0]: invalid amount \"-3cherry\": coin -3cherry amount is not positive",
				"fees[0]: invalid account \"badfee2addr\": " + bech32Err,
				"navs[0]: invalid assets \"-2cherry\": negative coin amount: -2",
				"invalid event tag \"pbbbb...bbbbf\" (length 101): exceeds max length 100",
			},
		},
	}

	for _, tc := range tests {
		for _, requireInputs := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s: Validate(%t)", tc.name, requireInputs), func(t *testing.T) {
				expErr := getExpErr(tc, requireInputs)
				var err error
				testFunc := func() {
					err = tc.msg.Validate(requireInputs)
				}
				require.NotPanics(t, testFunc, "%T.Validate(%t)", tc.msg, requireInputs)
				assertions.AssertErrorContents(t, err, expErr, "%T.Validate(%t) error", tc.msg, requireInputs)
			})
		}

		t.Run(fmt.Sprintf("%s: ValidateBasic()", tc.name), func(t *testing.T) {
			expErr := getExpErr(tc, true)
			testValidateBasic(t, &tc.msg, expErr)
		})
	}
}

func TestMsgMarketReleaseCommitmentsRequest_ValidateBasic(t *testing.T) {
	toAccAddr := func(str string) string {
		return sdk.AccAddress(str + strings.Repeat("_", 20-len(str))).String()
	}

	tests := []struct {
		name   string
		msg    MsgMarketReleaseCommitmentsRequest
		expErr []string
	}{
		{
			name: "control",
			msg: MsgMarketReleaseCommitmentsRequest{
				Admin:     toAccAddr("admin"),
				MarketId:  1,
				ToRelease: []AccountAmount{{Account: toAccAddr("torelease0")}},
			},
		},
		{
			name: "control with optional fields",
			msg: MsgMarketReleaseCommitmentsRequest{
				Admin:     toAccAddr("admin"),
				MarketId:  1,
				ToRelease: []AccountAmount{{Account: toAccAddr("torelease0")}},
				EventTag:  "tagtagtag",
			},
		},
		{
			name: "no admin",
			msg: MsgMarketReleaseCommitmentsRequest{
				Admin:     "",
				MarketId:  1,
				ToRelease: []AccountAmount{{Account: toAccAddr("torelease0")}},
			},
			expErr: []string{"invalid administrator \"\": " + emptyAddrErr},
		},
		{
			name: "bad admin",
			msg: MsgMarketReleaseCommitmentsRequest{
				Admin:     "badbadadmin",
				MarketId:  1,
				ToRelease: []AccountAmount{{Account: toAccAddr("torelease0")}},
			},
			expErr: []string{"invalid administrator \"badbadadmin\": " + bech32Err},
		},
		{
			name: "market zero",
			msg: MsgMarketReleaseCommitmentsRequest{
				Admin:     toAccAddr("admin"),
				MarketId:  0,
				ToRelease: []AccountAmount{{Account: toAccAddr("torelease0")}},
			},
			expErr: []string{"invalid market id: cannot be zero"},
		},
		{
			name: "nil to release",
			msg: MsgMarketReleaseCommitmentsRequest{
				Admin:     toAccAddr("admin"),
				MarketId:  1,
				ToRelease: nil,
			},
			expErr: []string{"nothing to release"},
		},
		{
			name: "empty to release",
			msg: MsgMarketReleaseCommitmentsRequest{
				Admin:     toAccAddr("admin"),
				MarketId:  1,
				ToRelease: []AccountAmount{},
			},
			expErr: []string{"nothing to release"},
		},
		{
			name: "bad to release entries",
			msg: MsgMarketReleaseCommitmentsRequest{
				Admin:    toAccAddr("admin"),
				MarketId: 1,
				ToRelease: []AccountAmount{
					{Account: "torelease0"},
					{Account: toAccAddr("torelease1")},
					{Account: toAccAddr("torelease2"), Amount: sdk.Coins{sdk.Coin{Denom: "cherry", Amount: sdkmath.NewInt(-1)}}},
				},
			},
			expErr: []string{
				"to release[0]: invalid account \"torelease0\": " + bech32Err,
				"to release[2]: invalid amount \"-1cherry\": coin -1cherry amount is not positive",
			},
		},
		{
			name: "bad event tag",
			msg: MsgMarketReleaseCommitmentsRequest{
				Admin:     toAccAddr("admin"),
				MarketId:  1,
				ToRelease: []AccountAmount{{Account: toAccAddr("torelease0")}},
				EventTag:  "abcd" + strings.Repeat("M", 93) + "wxyz",
			},
			expErr: []string{"invalid event tag \"abcdM...Mwxyz\" (length 101): exceeds max length 100"},
		},
		{
			name: "multiple errors",
			msg: MsgMarketReleaseCommitmentsRequest{
				Admin:     "",
				MarketId:  0,
				ToRelease: nil,
				EventTag:  "abcd" + strings.Repeat("M", 93) + "wxyz",
			},
			expErr: []string{
				"invalid administrator \"\": " + emptyAddrErr,
				"invalid market id: cannot be zero",
				"nothing to release",
				"invalid event tag \"abcdM...Mwxyz\" (length 101): exceeds max length 100",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

func TestMsgMarketSetOrderExternalIDRequest_ValidateBasic(t *testing.T) {
	admin := sdk.AccAddress("admin_address_______").String()

	tests := []struct {
		name   string
		msg    MsgMarketSetOrderExternalIDRequest
		expErr []string
	}{
		{
			name: "control",
			msg: MsgMarketSetOrderExternalIDRequest{
				Admin:      admin,
				MarketId:   5,
				OrderId:    12,
				ExternalId: "twenty-eight",
			},
			expErr: nil,
		},
		{
			name: "no admin",
			msg: MsgMarketSetOrderExternalIDRequest{
				Admin:      "",
				MarketId:   5,
				OrderId:    12,
				ExternalId: "twenty-eight",
			},
			expErr: []string{"invalid administrator \"\"", emptyAddrErr},
		},
		{
			name: "bad admin",
			msg: MsgMarketSetOrderExternalIDRequest{
				Admin:      "badadmin",
				MarketId:   5,
				OrderId:    12,
				ExternalId: "twenty-eight",
			},
			expErr: []string{"invalid administrator \"badadmin\"", bech32Err},
		},
		{
			name: "market zero",
			msg: MsgMarketSetOrderExternalIDRequest{
				Admin:      admin,
				MarketId:   0,
				OrderId:    12,
				ExternalId: "twenty-eight",
			},
			expErr: []string{"invalid market id: cannot be zero"},
		},
		{
			name: "empty external id",
			msg: MsgMarketSetOrderExternalIDRequest{
				Admin:      admin,
				MarketId:   5,
				OrderId:    12,
				ExternalId: "",
			},
			expErr: nil,
		},
		{
			name: "external id at max length",
			msg: MsgMarketSetOrderExternalIDRequest{
				Admin:      admin,
				MarketId:   5,
				OrderId:    12,
				ExternalId: strings.Repeat("y", MaxExternalIDLength),
			},
			expErr: nil,
		},
		{
			name: "external id more than max length",
			msg: MsgMarketSetOrderExternalIDRequest{
				Admin:      admin,
				MarketId:   5,
				OrderId:    12,
				ExternalId: strings.Repeat("r", MaxExternalIDLength+1),
			},
			expErr: []string{
				fmt.Sprintf("invalid external id %q (length %d): max length %d",
					"rrrrr...rrrrr", MaxExternalIDLength+1, MaxExternalIDLength),
			},
		},
		{
			name: "order zero",
			msg: MsgMarketSetOrderExternalIDRequest{
				Admin:      admin,
				MarketId:   5,
				OrderId:    0,
				ExternalId: "twenty-eight",
			},
			expErr: []string{"invalid order id: cannot be zero"},
		},
		{
			name: "multiple errors",
			msg: MsgMarketSetOrderExternalIDRequest{
				Admin:      "",
				MarketId:   0,
				OrderId:    0,
				ExternalId: strings.Repeat("V", MaxExternalIDLength+1),
			},
			expErr: []string{
				"invalid administrator \"\"", emptyAddrErr,
				"invalid market id: cannot be zero",
				fmt.Sprintf("invalid external id %q (length %d): max length %d",
					"VVVVV...VVVVV", MaxExternalIDLength+1, MaxExternalIDLength),
				"invalid order id: cannot be zero",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

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
				Admin:     goodAdminAddr,
				MarketId:  1,
				ToAddress: goodToAddr,
				Amount:    goodCoins,
			},
			expErr: nil,
		},
		{
			name: "no administrator",
			msg: MsgMarketWithdrawRequest{
				Admin:     "",
				MarketId:  1,
				ToAddress: goodToAddr,
				Amount:    goodCoins,
			},
			expErr: []string{`invalid administrator ""`, emptyAddrErr},
		},
		{
			name: "bad administrator",
			msg: MsgMarketWithdrawRequest{
				Admin:     "notright",
				MarketId:  1,
				ToAddress: goodToAddr,
				Amount:    goodCoins,
			},
			expErr: []string{`invalid administrator "notright"`, bech32Err},
		},
		{
			name: "market id zero",
			msg: MsgMarketWithdrawRequest{
				Admin:     goodAdminAddr,
				MarketId:  0,
				ToAddress: goodToAddr,
				Amount:    goodCoins,
			},
			expErr: []string{"invalid market id", "cannot be zero"},
		},
		{
			name: "no to address",
			msg: MsgMarketWithdrawRequest{
				Admin:     goodAdminAddr,
				MarketId:  1,
				ToAddress: "",
				Amount:    goodCoins,
			},
			expErr: []string{`invalid to address ""`, emptyAddrErr},
		},
		{
			name: "bad to address",
			msg: MsgMarketWithdrawRequest{
				Admin:     goodAdminAddr,
				MarketId:  1,
				ToAddress: "notright",
				Amount:    goodCoins,
			},
			expErr: []string{`invalid to address "notright"`, bech32Err},
		},
		{
			name: "invalid denom in amount",
			msg: MsgMarketWithdrawRequest{
				Admin:     goodAdminAddr,
				MarketId:  1,
				ToAddress: goodToAddr,
				Amount:    sdk.Coins{coin(3, "x")},
			},
			expErr: []string{`invalid amount "3x"`, "invalid denom: x"},
		},
		{
			name: "negative amount",
			msg: MsgMarketWithdrawRequest{
				Admin:     goodAdminAddr,
				MarketId:  1,
				ToAddress: goodToAddr,
				Amount:    sdk.Coins{coin(-1, "negcoin")},
			},
			expErr: []string{`invalid amount "-1negcoin"`, "coin -1negcoin amount is not positive"},
		},
		{
			name: "empty amount",
			msg: MsgMarketWithdrawRequest{
				Admin:     goodAdminAddr,
				MarketId:  1,
				ToAddress: goodToAddr,
				Amount:    sdk.Coins{},
			},
			expErr: []string{`invalid amount ""`, "cannot be zero"},
		},
		{
			name: "zero coin amount",
			msg: MsgMarketWithdrawRequest{
				Admin:     goodAdminAddr,
				MarketId:  1,
				ToAddress: goodToAddr,
				Amount:    sdk.Coins{coin(0, "acorn"), coin(0, "banana")},
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

func TestMsgMarketUpdateDetailsRequest_ValidateBasic(t *testing.T) {
	admin := sdk.AccAddress("admin_______________").String()
	tooLongErr := func(field string, max int) string {
		return fmt.Sprintf("%s length %d exceeds maximum length of %d", field, max+1, max)
	}

	tests := []struct {
		name   string
		msg    MsgMarketUpdateDetailsRequest
		expErr []string
	}{
		{
			name: "control",
			msg: MsgMarketUpdateDetailsRequest{
				Admin:    admin,
				MarketId: 1,
				MarketDetails: MarketDetails{
					Name:        "MyMarket",
					Description: "This is my own market only for me.",
					WebsiteUrl:  "https://example.com",
					IconUri:     "https://example.com/icon",
				},
			},
			expErr: nil,
		},
		{
			name: "empty admin",
			msg: MsgMarketUpdateDetailsRequest{
				Admin:         "",
				MarketId:      1,
				MarketDetails: MarketDetails{},
			},
			expErr: []string{`invalid administrator ""`, emptyAddrErr},
		},
		{
			name: "invalid admin",
			msg: MsgMarketUpdateDetailsRequest{
				Admin:         "notvalidadmin",
				MarketId:      1,
				MarketDetails: MarketDetails{},
			},
			expErr: []string{`invalid administrator "notvalidadmin"`, bech32Err},
		},
		{
			name: "market id zero",
			msg: MsgMarketUpdateDetailsRequest{
				Admin:         admin,
				MarketId:      0,
				MarketDetails: MarketDetails{},
			},
			expErr: []string{"invalid market id", "cannot be zero"},
		},
		{
			name: "name too long",
			msg: MsgMarketUpdateDetailsRequest{
				Admin:    admin,
				MarketId: 1,
				MarketDetails: MarketDetails{
					Name: strings.Repeat("p", MaxName+1),
				},
			},
			expErr: []string{tooLongErr("name", MaxName)},
		},
		{
			name: "description too long",
			msg: MsgMarketUpdateDetailsRequest{
				Admin:    admin,
				MarketId: 1,
				MarketDetails: MarketDetails{
					Description: strings.Repeat("d", MaxDescription+1),
				},
			},
			expErr: []string{tooLongErr("description", MaxDescription)},
		},
		{
			name: "website_url too long",
			msg: MsgMarketUpdateDetailsRequest{
				Admin:    admin,
				MarketId: 1,
				MarketDetails: MarketDetails{
					WebsiteUrl: strings.Repeat("w", MaxWebsiteURL+1),
				},
			},
			expErr: []string{tooLongErr("website_url", MaxWebsiteURL)},
		},
		{
			name: "icon_uri too long",
			msg: MsgMarketUpdateDetailsRequest{
				Admin:    admin,
				MarketId: 1,
				MarketDetails: MarketDetails{
					IconUri: strings.Repeat("i", MaxIconURI+1),
				},
			},
			expErr: []string{tooLongErr("icon_uri", MaxIconURI)},
		},
		{
			name: "multiple errors",
			msg: MsgMarketUpdateDetailsRequest{
				Admin:    "",
				MarketId: 0,
				MarketDetails: MarketDetails{
					Name:        strings.Repeat("p", MaxName+1),
					Description: strings.Repeat("d", MaxDescription+1),
					WebsiteUrl:  strings.Repeat("w", MaxWebsiteURL+1),
					IconUri:     strings.Repeat("i", MaxIconURI+1),
				},
			},
			expErr: []string{
				"invalid administrator",
				"invalid market id",
				tooLongErr("name", MaxName),
				tooLongErr("description", MaxDescription),
				tooLongErr("website_url", MaxWebsiteURL),
				tooLongErr("icon_uri", MaxIconURI),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

func TestMsgMarketUpdateEnabledRequest_ValidateBasic(t *testing.T) {
	msg := MsgMarketUpdateEnabledRequest{}
	expErr := []string{"the MarketUpdateEnabled endpoint has been replaced by the MarketUpdateAcceptingOrders endpoint"}
	testValidateBasic(t, &msg, expErr)
}

func TestMsgMarketUpdateAcceptingOrdersRequest_ValidateBasic(t *testing.T) {
	admin := sdk.AccAddress("admin_______________").String()

	tests := []struct {
		name   string
		msg    MsgMarketUpdateAcceptingOrdersRequest
		expErr []string
	}{
		{
			name: "control: true",
			msg: MsgMarketUpdateAcceptingOrdersRequest{
				Admin:           admin,
				MarketId:        1,
				AcceptingOrders: true,
			},
			expErr: nil,
		},
		{
			name: "control: false",
			msg: MsgMarketUpdateAcceptingOrdersRequest{
				Admin:           admin,
				MarketId:        1,
				AcceptingOrders: true,
			},
			expErr: nil,
		},
		{
			name: "empty admin",
			msg: MsgMarketUpdateAcceptingOrdersRequest{
				Admin:    "",
				MarketId: 1,
			},
			expErr: []string{
				`invalid administrator ""`, emptyAddrErr,
			},
		},
		{
			name: "bad admin",
			msg: MsgMarketUpdateAcceptingOrdersRequest{
				Admin:    "badadmin",
				MarketId: 1,
			},
			expErr: []string{
				`invalid administrator "badadmin"`, bech32Err,
			},
		},
		{
			name: "market id zero",
			msg: MsgMarketUpdateAcceptingOrdersRequest{
				Admin:    admin,
				MarketId: 0,
			},
			expErr: []string{
				"invalid market id", "cannot be zero",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

func TestMsgMarketUpdateUserSettleRequest_ValidateBasic(t *testing.T) {
	admin := sdk.AccAddress("admin_______________").String()

	tests := []struct {
		name   string
		msg    MsgMarketUpdateUserSettleRequest
		expErr []string
	}{
		{
			name: "control: true",
			msg: MsgMarketUpdateUserSettleRequest{
				Admin:               admin,
				MarketId:            1,
				AllowUserSettlement: true,
			},
			expErr: nil,
		},
		{
			name: "control: false",
			msg: MsgMarketUpdateUserSettleRequest{
				Admin:               admin,
				MarketId:            1,
				AllowUserSettlement: true,
			},
			expErr: nil,
		},
		{
			name: "empty admin",
			msg: MsgMarketUpdateUserSettleRequest{
				Admin:    "",
				MarketId: 1,
			},
			expErr: []string{
				`invalid administrator ""`, emptyAddrErr,
			},
		},
		{
			name: "bad admin",
			msg: MsgMarketUpdateUserSettleRequest{
				Admin:    "badadmin",
				MarketId: 1,
			},
			expErr: []string{
				`invalid administrator "badadmin"`, bech32Err,
			},
		},
		{
			name: "market id zero",
			msg: MsgMarketUpdateUserSettleRequest{
				Admin:    admin,
				MarketId: 0,
			},
			expErr: []string{
				"invalid market id", "cannot be zero",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

func TestMsgMarketUpdateAcceptingCommitmentsRequest_ValidateBasic(t *testing.T) {
	tests := []struct {
		name   string
		msg    MsgMarketUpdateAcceptingCommitmentsRequest
		expErr []string
	}{
		{
			name: "control: false",
			msg: MsgMarketUpdateAcceptingCommitmentsRequest{
				Admin:                sdk.AccAddress("admin_______________").String(),
				MarketId:             1,
				AcceptingCommitments: false,
			},
		},
		{
			name: "control: true",
			msg: MsgMarketUpdateAcceptingCommitmentsRequest{
				Admin:                sdk.AccAddress("admin_______________").String(),
				MarketId:             1,
				AcceptingCommitments: true,
			},
		},
		{
			name: "no admin",
			msg: MsgMarketUpdateAcceptingCommitmentsRequest{
				Admin:    "",
				MarketId: 1,
			},
			expErr: []string{"invalid administrator \"\": " + emptyAddrErr},
		},
		{
			name: "bad admin",
			msg: MsgMarketUpdateAcceptingCommitmentsRequest{
				Admin:    "notanadminaddr",
				MarketId: 1,
			},
			expErr: []string{"invalid administrator \"notanadminaddr\": " + bech32Err},
		},
		{
			name: "market zero",
			msg: MsgMarketUpdateAcceptingCommitmentsRequest{
				Admin:    sdk.AccAddress("admin_______________").String(),
				MarketId: 0,
			},
			expErr: []string{"invalid market id: cannot be zero"},
		},
		{
			name: "multiple errors",
			msg:  MsgMarketUpdateAcceptingCommitmentsRequest{},
			expErr: []string{
				"invalid administrator \"\": " + emptyAddrErr,
				"invalid market id: cannot be zero",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

func TestMsgMarketUpdateIntermediaryDenomRequest_ValidateBasic(t *testing.T) {
	tests := []struct {
		name   string
		msg    MsgMarketUpdateIntermediaryDenomRequest
		expErr []string
	}{
		{
			name: "control",
			msg: MsgMarketUpdateIntermediaryDenomRequest{
				Admin:             sdk.AccAddress("admin_______________").String(),
				MarketId:          1,
				IntermediaryDenom: "musd",
			},
		},
		{
			name: "no admin",
			msg: MsgMarketUpdateIntermediaryDenomRequest{
				Admin:             "",
				MarketId:          1,
				IntermediaryDenom: "musd",
			},
			expErr: []string{"invalid administrator \"\": " + emptyAddrErr},
		},
		{
			name: "bad admin",
			msg: MsgMarketUpdateIntermediaryDenomRequest{
				Admin:             "notanadminaddr",
				MarketId:          1,
				IntermediaryDenom: "musd",
			},
			expErr: []string{"invalid administrator \"notanadminaddr\": " + bech32Err},
		},
		{
			name: "market zero",
			msg: MsgMarketUpdateIntermediaryDenomRequest{
				Admin:             sdk.AccAddress("admin_______________").String(),
				MarketId:          0,
				IntermediaryDenom: "musd",
			},
			expErr: []string{"invalid market id: cannot be zero"},
		},
		{
			name: "no intermediary denom",
			msg: MsgMarketUpdateIntermediaryDenomRequest{
				Admin:             sdk.AccAddress("admin_______________").String(),
				MarketId:          1,
				IntermediaryDenom: "",
			},
		},
		{
			name: "invalid intermediary denom",
			msg: MsgMarketUpdateIntermediaryDenomRequest{
				Admin:             sdk.AccAddress("admin_______________").String(),
				MarketId:          1,
				IntermediaryDenom: "x",
			},
			expErr: []string{"invalid intermediary denom: invalid denom: x"},
		},
		{
			name: "multiple errors",
			msg: MsgMarketUpdateIntermediaryDenomRequest{
				Admin:             "",
				MarketId:          0,
				IntermediaryDenom: "x",
			},
			expErr: []string{
				"invalid administrator \"\": " + emptyAddrErr,
				"invalid market id: cannot be zero",
				"invalid intermediary denom: invalid denom: x",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

func TestMsgMarketManagePermissionsRequest_ValidateBasic(t *testing.T) {
	goodAdminAddr := sdk.AccAddress("goodAdminAddr_______").String()
	goodAddr1 := sdk.AccAddress("goodAddr1___________").String()
	goodAddr2 := sdk.AccAddress("goodAddr2___________").String()
	goodAddr3 := sdk.AccAddress("goodAddr3___________").String()

	tests := []struct {
		name   string
		msg    MsgMarketManagePermissionsRequest
		expErr []string
	}{
		{
			name: "control",
			msg: MsgMarketManagePermissionsRequest{
				Admin:     goodAdminAddr,
				MarketId:  1,
				RevokeAll: []string{goodAddr1},
				ToRevoke:  []AccessGrant{{Address: goodAddr2, Permissions: []Permission{Permission_settle}}},
				ToGrant:   []AccessGrant{{Address: goodAddr3, Permissions: []Permission{Permission_cancel}}},
			},
			expErr: nil,
		},
		{
			name: "empty admin",
			msg: MsgMarketManagePermissionsRequest{
				Admin:     "",
				MarketId:  1,
				RevokeAll: []string{goodAddr1},
			},
			expErr: []string{`invalid administrator ""`, emptyAddrErr},
		},
		{
			name: "invalid admin",
			msg: MsgMarketManagePermissionsRequest{
				Admin:     "bad1admin",
				MarketId:  1,
				RevokeAll: []string{goodAddr1},
			},
			expErr: []string{`invalid administrator "bad1admin"`, bech32Err},
		},
		{
			name: "market id zero",
			msg: MsgMarketManagePermissionsRequest{
				Admin:     goodAdminAddr,
				MarketId:  0,
				RevokeAll: []string{goodAddr1},
			},
			expErr: []string{"invalid market id", "cannot be zero"},
		},
		{
			name: "no updates",
			msg: MsgMarketManagePermissionsRequest{
				Admin:    goodAdminAddr,
				MarketId: 1,
			},
			expErr: []string{"no updates"},
		},
		{
			name: "two invalid addresses in revoke all",
			msg: MsgMarketManagePermissionsRequest{
				Admin:     goodAdminAddr,
				MarketId:  1,
				RevokeAll: []string{"bad1addr", "bad2addr"},
			},
			expErr: []string{
				`invalid revoke-all address "bad1addr": ` + bech32Err,
				`invalid revoke-all address "bad2addr": ` + bech32Err,
			},
		},
		{
			name: "two invalid to-revoke entries",
			msg: MsgMarketManagePermissionsRequest{
				Admin:    goodAdminAddr,
				MarketId: 1,
				ToRevoke: []AccessGrant{
					{Address: "badaddr", Permissions: []Permission{Permission_withdraw}},
					{Address: goodAddr1, Permissions: []Permission{Permission_unspecified}},
				},
			},
			expErr: []string{
				`invalid to-revoke access grant: invalid address "badaddr"`,
				"invalid to-revoke access grant: permission is unspecified for " + goodAddr1,
			},
		},
		{
			name: "two addrs in both revoke-all and to-revoke",
			msg: MsgMarketManagePermissionsRequest{
				Admin:     goodAdminAddr,
				MarketId:  1,
				RevokeAll: []string{goodAddr1, goodAddr2, goodAddr3},
				ToRevoke: []AccessGrant{
					{Address: goodAddr2, Permissions: []Permission{Permission_update}},
					{Address: goodAddr1, Permissions: []Permission{Permission_permissions}},
				},
			},
			expErr: []string{
				"address " + goodAddr2 + " appears in both the revoke-all and to-revoke fields",
				"address " + goodAddr1 + " appears in both the revoke-all and to-revoke fields",
			},
		},
		{
			name: "two invalid to-grant entries",
			msg: MsgMarketManagePermissionsRequest{
				Admin:    goodAdminAddr,
				MarketId: 1,
				ToGrant: []AccessGrant{
					{Address: "badaddr", Permissions: []Permission{Permission_withdraw}},
					{Address: goodAddr1, Permissions: []Permission{Permission_unspecified}},
				},
			},
			expErr: []string{
				`invalid to-grant access grant: invalid address "badaddr"`,
				"invalid to-grant access grant: permission is unspecified for " + goodAddr1,
			},
		},
		{
			name: "revoke all for two addrs and some for one, then add some for all three",
			msg: MsgMarketManagePermissionsRequest{
				Admin:     goodAdminAddr,
				MarketId:  1,
				RevokeAll: []string{goodAddr1, goodAddr2},
				ToRevoke:  []AccessGrant{{Address: goodAddr3, Permissions: []Permission{Permission_settle}}},
				ToGrant: []AccessGrant{
					{Address: goodAddr1, Permissions: []Permission{Permission_settle, Permission_update}},
					{Address: goodAddr2, Permissions: []Permission{Permission_cancel, Permission_permissions}},
					{Address: goodAddr3, Permissions: []Permission{Permission_withdraw, Permission_attributes}},
				},
			},
			expErr: nil,
		},
		{
			name: "revoke and grant the same permission for two addresses",
			msg: MsgMarketManagePermissionsRequest{
				Admin:    goodAdminAddr,
				MarketId: 1,
				ToRevoke: []AccessGrant{
					{Address: goodAddr2, Permissions: []Permission{Permission_permissions, Permission_cancel, Permission_attributes}},
					{Address: goodAddr1, Permissions: []Permission{Permission_settle, Permission_withdraw, Permission_update}},
				},
				ToGrant: []AccessGrant{
					{Address: goodAddr1, Permissions: []Permission{Permission_withdraw}},
					{Address: goodAddr2, Permissions: []Permission{Permission_attributes, Permission_permissions}},
				},
			},
			expErr: []string{
				"address " + goodAddr1 + " has both revoke and grant \"withdraw\"",
				"address " + goodAddr2 + " has both revoke and grant \"attributes\"",
				"address " + goodAddr2 + " has both revoke and grant \"permissions\"",
			},
		},
		{
			// We allow this because it can be fixed via gov prop if no one is left that can manage permissions.
			name: "admin in revoke-all",
			msg: MsgMarketManagePermissionsRequest{
				Admin:     goodAdminAddr,
				MarketId:  1,
				RevokeAll: []string{goodAddr1, goodAdminAddr, goodAddr2},
			},
			expErr: nil,
		},
		{
			name: "admin revoking all their permissions except permissions",
			msg: MsgMarketManagePermissionsRequest{
				Admin:    goodAdminAddr,
				MarketId: 1,
				ToRevoke: []AccessGrant{
					{
						Address: goodAdminAddr,
						Permissions: []Permission{Permission_settle, Permission_cancel,
							Permission_withdraw, Permission_update, Permission_attributes},
					},
				},
			},
			expErr: nil,
		},
		{
			// We allow this because it can be fixed via gov prop if no one is left that can manage permissions.
			name: "admin revoking own ability to manage permissions",
			msg: MsgMarketManagePermissionsRequest{
				Admin:    goodAdminAddr,
				MarketId: 1,
				ToRevoke: []AccessGrant{
					{Address: goodAddr1, Permissions: []Permission{Permission_permissions}},
					{Address: goodAdminAddr, Permissions: []Permission{Permission_permissions}},
					{Address: goodAddr2, Permissions: []Permission{Permission_permissions}},
				},
			},
			expErr: nil,
		},
		{
			name: "multiple errs",
			msg: MsgMarketManagePermissionsRequest{
				Admin:     "",
				MarketId:  0,
				RevokeAll: []string{"bad-revoke-addr"},
				ToRevoke:  []AccessGrant{{Address: goodAddr1, Permissions: []Permission{Permission_unspecified}}},
				ToGrant:   []AccessGrant{{Address: "bad-grant-addr", Permissions: []Permission{Permission_withdraw}}},
			},
			expErr: []string{
				"invalid administrator \"\"",
				"invalid market id",
				"invalid revoke-all address \"bad-revoke-addr\"",
				"invalid to-revoke access grant: permission is unspecified for " + goodAddr1,
				`invalid to-grant access grant: invalid address "bad-grant-addr"`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

func TestMsgMarketManagePermissionsRequest_HasUpdates(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgMarketManagePermissionsRequest
		exp  bool
	}{
		{
			name: "empty",
			msg:  MsgMarketManagePermissionsRequest{},
			exp:  false,
		},
		{
			name: "empty except for admin",
			msg: MsgMarketManagePermissionsRequest{
				Admin: "admin",
			},
			exp: false,
		},
		{
			name: "one revoke all",
			msg: MsgMarketManagePermissionsRequest{
				RevokeAll: []string{"revoke_all"},
			},
			exp: true,
		},
		{
			name: "one to revoke",
			msg: MsgMarketManagePermissionsRequest{
				ToRevoke: []AccessGrant{{Address: "to_revoke"}},
			},
			exp: true,
		},
		{
			name: "one to grant",
			msg: MsgMarketManagePermissionsRequest{
				ToGrant: []AccessGrant{{Address: "to_grant"}},
			},
			exp: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual bool
			testFunc := func() {
				actual = tc.msg.HasUpdates()
			}
			require.NotPanics(t, testFunc, "%T.HasUpdates()", tc.msg)
			assert.Equal(t, tc.exp, actual, "%T.HasUpdates()", tc.msg)
		})
	}
}

func TestMsgMarketManageReqAttrsRequest_ValidateBasic(t *testing.T) {
	goodAdmin := sdk.AccAddress("goodAdmin___________").String()

	tests := []struct {
		name   string
		msg    MsgMarketManageReqAttrsRequest
		expErr []string
	}{
		{
			name: "no admin",
			msg: MsgMarketManageReqAttrsRequest{
				Admin:          "",
				MarketId:       1,
				CreateAskToAdd: []string{"abc"},
			},
			expErr: []string{"invalid administrator", emptyAddrErr},
		},
		{
			name: "bad admin",
			msg: MsgMarketManageReqAttrsRequest{
				Admin:          "not1valid",
				MarketId:       1,
				CreateAskToAdd: []string{"abc"},
			},
			expErr: []string{"invalid administrator", bech32Err},
		},
		{
			name: "market id zero",
			msg: MsgMarketManageReqAttrsRequest{
				Admin:          goodAdmin,
				CreateAskToAdd: []string{"abc"},
			},
			expErr: []string{"invalid market id: cannot be zero"},
		},
		{
			name: "no updates",
			msg: MsgMarketManageReqAttrsRequest{
				Admin:    goodAdmin,
				MarketId: 1,
			},
			expErr: []string{"no updates"},
		},
		{
			name: "invalid create ask to add entry",
			msg: MsgMarketManageReqAttrsRequest{
				Admin:          goodAdmin,
				MarketId:       1,
				CreateAskToAdd: []string{"in-valid-attr"},
			},
			expErr: []string{"invalid create-ask to add required attribute \"in-valid-attr\""},
		},
		{
			name: "invalid create ask to remove entry",
			msg: MsgMarketManageReqAttrsRequest{
				Admin:             goodAdmin,
				MarketId:          1,
				CreateAskToRemove: []string{"in-valid-attr"},
			},
		},
		{
			name: "invalid create bid to add entry",
			msg: MsgMarketManageReqAttrsRequest{
				Admin:          goodAdmin,
				MarketId:       1,
				CreateBidToAdd: []string{"in-valid-attr"},
			},
			expErr: []string{"invalid create-bid to add required attribute \"in-valid-attr\""},
		},
		{
			name: "invalid create bid to remove entry",
			msg: MsgMarketManageReqAttrsRequest{
				Admin:             goodAdmin,
				MarketId:          1,
				CreateBidToRemove: []string{"in-valid-attr"},
			},
		},
		{
			name: "invalid create commitment to add entry",
			msg: MsgMarketManageReqAttrsRequest{
				Admin:                 goodAdmin,
				MarketId:              1,
				CreateCommitmentToAdd: []string{"in-valid-attr"},
			},
			expErr: []string{"invalid create-commitment to add required attribute \"in-valid-attr\""},
		},
		{
			name: "invalid create commitment to remove entry",
			msg: MsgMarketManageReqAttrsRequest{
				Admin:                    goodAdmin,
				MarketId:                 1,
				CreateCommitmentToRemove: []string{"in-valid-attr"},
			},
		},
		{
			name: "add and remove same create ask entry",
			msg: MsgMarketManageReqAttrsRequest{
				Admin:             goodAdmin,
				MarketId:          1,
				CreateAskToAdd:    []string{"abc", "def", "ghi"},
				CreateAskToRemove: []string{"jkl", "abc"},
			},
			expErr: []string{"cannot add and remove the same create-ask required attributes \"abc\""},
		},
		{
			name: "add and remove same create bid entry",
			msg: MsgMarketManageReqAttrsRequest{
				Admin:             goodAdmin,
				MarketId:          1,
				CreateBidToAdd:    []string{"abc", "def", "ghi"},
				CreateBidToRemove: []string{"jkl", "abc"},
			},
			expErr: []string{"cannot add and remove the same create-bid required attributes \"abc\""},
		},
		{
			name: "add and remove same create commitment entry",
			msg: MsgMarketManageReqAttrsRequest{
				Admin:                    goodAdmin,
				MarketId:                 1,
				CreateCommitmentToAdd:    []string{"abc", "def", "ghi"},
				CreateCommitmentToRemove: []string{"jkl", "abc"},
			},
			expErr: []string{"cannot add and remove the same create-commitment required attributes \"abc\""},
		},
		{
			name: "add to create-ask the same as remove from create-bid",
			msg: MsgMarketManageReqAttrsRequest{
				Admin:             goodAdmin,
				MarketId:          1,
				CreateAskToAdd:    []string{"abc", "def", "ghi"},
				CreateBidToRemove: []string{"jkl", "abc"},
			},
		},
		{
			name: "add to create-bid the same as remove from create-ask",
			msg: MsgMarketManageReqAttrsRequest{
				Admin:             goodAdmin,
				MarketId:          1,
				CreateBidToAdd:    []string{"abc", "def", "ghi"},
				CreateAskToRemove: []string{"jkl", "abc"},
			},
		},
		{
			name: "add one to and remove one from each",
			msg: MsgMarketManageReqAttrsRequest{
				Admin:                    goodAdmin,
				MarketId:                 1,
				CreateAskToAdd:           []string{"to-add.ask"},
				CreateAskToRemove:        []string{"to-remove.ask"},
				CreateBidToAdd:           []string{"to-add.bid"},
				CreateBidToRemove:        []string{"to-remove.bid"},
				CreateCommitmentToAdd:    []string{"to-add.com"},
				CreateCommitmentToRemove: []string{"to-remove.com"},
			},
		},
		{
			name: "no admin and no market id and no updates",
			msg:  MsgMarketManageReqAttrsRequest{},
			expErr: []string{
				"invalid administrator",
				"invalid market id",
				"no updates",
			},
		},
		{
			name: "multiple errors",
			msg: MsgMarketManageReqAttrsRequest{
				Admin:                    "not1valid",
				MarketId:                 0,
				CreateAskToAdd:           []string{"bad-ask-attr", "dup-ask"},
				CreateAskToRemove:        []string{"dup-ask"},
				CreateBidToAdd:           []string{"bad-bid-attr", "dup-bid"},
				CreateBidToRemove:        []string{"dup-bid"},
				CreateCommitmentToAdd:    []string{"bad-com-attr", "dup-com"},
				CreateCommitmentToRemove: []string{"dup-com"},
			},
			expErr: []string{
				"invalid administrator",
				"invalid market id",
				"invalid create-ask to add required attribute \"bad-ask-attr\"",
				"cannot add and remove the same create-ask required attributes \"dup-ask\"",
				"invalid create-bid to add required attribute \"bad-bid-attr\"",
				"cannot add and remove the same create-bid required attributes \"dup-bid\"",
				"invalid create-commitment to add required attribute \"bad-com-attr\"",
				"cannot add and remove the same create-commitment required attributes \"dup-com\"",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

func TestMsgMarketManageReqAttrsRequest_HasUpdates(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgMarketManageReqAttrsRequest
		exp  bool
	}{
		{
			name: "empty",
			msg:  MsgMarketManageReqAttrsRequest{},
			exp:  false,
		},
		{
			name: "empty except for admin",
			msg: MsgMarketManageReqAttrsRequest{
				Admin: "admin",
			},
			exp: false,
		},
		{
			name: "one ask to add",
			msg: MsgMarketManageReqAttrsRequest{
				CreateAskToAdd: []string{"ask_to_add"},
			},
			exp: true,
		},
		{
			name: "one ask to remove",
			msg: MsgMarketManageReqAttrsRequest{
				CreateAskToRemove: []string{"ask_to_remove"},
			},
			exp: true,
		},
		{
			name: "one bid to add",
			msg: MsgMarketManageReqAttrsRequest{
				CreateBidToAdd: []string{"bid_to_add"},
			},
			exp: true,
		},
		{
			name: "one bid to remove",
			msg: MsgMarketManageReqAttrsRequest{
				CreateBidToRemove: []string{"bid_to_remove"},
			},
			exp: true,
		},
		{
			name: "one commitment to add",
			msg: MsgMarketManageReqAttrsRequest{
				CreateCommitmentToAdd: []string{"commitment_to_add"},
			},
			exp: true,
		},
		{
			name: "one commitment to remove",
			msg: MsgMarketManageReqAttrsRequest{
				CreateCommitmentToRemove: []string{"commitment_to_remove"},
			},
			exp: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual bool
			testFunc := func() {
				actual = tc.msg.HasUpdates()
			}
			require.NotPanics(t, testFunc, "%T.HasUpdates()", tc.msg)
			assert.Equal(t, tc.exp, actual, "%T.HasUpdates()", tc.msg)
		})
	}
}

func TestMsgCreatePaymentRequest_ValidateBasic(t *testing.T) {
	tests := []struct {
		name   string
		msg    MsgCreatePaymentRequest
		expErr []string
	}{
		{
			name:   "valid payment",
			msg:    MsgCreatePaymentRequest{Payment: ValidPayment},
			expErr: nil,
		},
		{
			name: "no target",
			msg: MsgCreatePaymentRequest{Payment: Payment{
				Source:       ValidPayment.Source,
				SourceAmount: ValidPayment.SourceAmount,
				Target:       "",
				TargetAmount: ValidPayment.TargetAmount,
				ExternalId:   ValidPayment.ExternalId,
			}},
			expErr: nil,
		},
		{
			name: "no source amount",
			msg: MsgCreatePaymentRequest{Payment: Payment{
				Source:       ValidPayment.Source,
				SourceAmount: nil,
				Target:       ValidPayment.Target,
				TargetAmount: ValidPayment.TargetAmount,
				ExternalId:   ValidPayment.ExternalId,
			}},
			expErr: nil,
		},
		{
			name: "no target amount",
			msg: MsgCreatePaymentRequest{Payment: Payment{
				Source:       ValidPayment.Source,
				SourceAmount: ValidPayment.SourceAmount,
				Target:       ValidPayment.Target,
				TargetAmount: nil,
				ExternalId:   ValidPayment.ExternalId,
			}},
			expErr: nil,
		},
		{
			name: "no external id",
			msg: MsgCreatePaymentRequest{Payment: Payment{
				Source:       ValidPayment.Source,
				SourceAmount: ValidPayment.SourceAmount,
				Target:       ValidPayment.Target,
				TargetAmount: ValidPayment.TargetAmount,
				ExternalId:   "",
			}},
			expErr: nil,
		},
		{
			name: "invalid payment",
			msg: MsgCreatePaymentRequest{Payment: Payment{
				Source:       "",
				SourceAmount: ValidPayment.SourceAmount,
				Target:       "notgoodeither",
				TargetAmount: ValidPayment.TargetAmount,
				ExternalId:   ValidPayment.ExternalId,
			}},
			expErr: []string{
				"invalid source \"\": empty address string is not allowed",
				"invalid target \"notgoodeither\": decoding bech32 failed: invalid separator index -1",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

func TestMsgAcceptPaymentRequest_ValidateBasic(t *testing.T) {
	tests := []struct {
		name   string
		msg    MsgAcceptPaymentRequest
		expErr []string
	}{
		{
			name:   "valid payment",
			msg:    MsgAcceptPaymentRequest{Payment: ValidPayment},
			expErr: nil,
		},
		{
			name: "invalid payment",
			msg: MsgAcceptPaymentRequest{Payment: Payment{
				Source:       "",
				SourceAmount: ValidPayment.SourceAmount,
				Target:       "notgoodeither",
				TargetAmount: ValidPayment.TargetAmount,
				ExternalId:   ValidPayment.ExternalId,
			}},
			expErr: []string{
				"invalid source \"\": empty address string is not allowed",
				"invalid target \"notgoodeither\": decoding bech32 failed: invalid separator index -1",
			},
		},
		{
			name: "no target",
			msg: MsgAcceptPaymentRequest{Payment: Payment{
				Source:       ValidPayment.Source,
				SourceAmount: ValidPayment.SourceAmount,
				Target:       "",
				TargetAmount: ValidPayment.TargetAmount,
				ExternalId:   ValidPayment.ExternalId,
			}},
			expErr: []string{"invalid target: empty address string is not allowed"},
		},
		{
			name: "no source amount",
			msg: MsgAcceptPaymentRequest{Payment: Payment{
				Source:       ValidPayment.Source,
				SourceAmount: nil,
				Target:       ValidPayment.Target,
				TargetAmount: ValidPayment.TargetAmount,
				ExternalId:   ValidPayment.ExternalId,
			}},
			expErr: nil,
		},
		{
			name: "no target amount",
			msg: MsgAcceptPaymentRequest{Payment: Payment{
				Source:       ValidPayment.Source,
				SourceAmount: ValidPayment.SourceAmount,
				Target:       ValidPayment.Target,
				TargetAmount: nil,
				ExternalId:   ValidPayment.ExternalId,
			}},
			expErr: nil,
		},
		{
			name: "no external id",
			msg: MsgAcceptPaymentRequest{Payment: Payment{
				Source:       ValidPayment.Source,
				SourceAmount: ValidPayment.SourceAmount,
				Target:       ValidPayment.Target,
				TargetAmount: ValidPayment.TargetAmount,
				ExternalId:   "",
			}},
			expErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

func TestMsgRejectPaymentRequest_ValidateBasic(t *testing.T) {
	validMsg := MsgRejectPaymentRequest{
		Target:     sdk.AccAddress("Target______________").String(),
		Source:     sdk.AccAddress("Source______________").String(),
		ExternalId: "just-some-id:1D250135-D735-42E7-9DC3-FDB374DE2604",
	}

	tests := []struct {
		name   string
		msg    MsgRejectPaymentRequest
		expErr []string
	}{
		{
			name:   "valid message",
			msg:    validMsg,
			expErr: nil,
		},
		{
			name: "no target",
			msg: MsgRejectPaymentRequest{
				Target:     "",
				Source:     validMsg.Source,
				ExternalId: validMsg.ExternalId,
			},
			expErr: []string{"invalid target \"\": empty address string is not allowed"},
		},
		{
			name: "invalid target",
			msg: MsgRejectPaymentRequest{
				Target:     "reallybad",
				Source:     validMsg.Source,
				ExternalId: validMsg.ExternalId,
			},
			expErr: []string{"invalid target \"reallybad\": decoding bech32 failed: invalid separator index -1"},
		},
		{
			name: "no source",
			msg: MsgRejectPaymentRequest{
				Target:     validMsg.Target,
				Source:     "",
				ExternalId: validMsg.ExternalId,
			},
			expErr: []string{"invalid source \"\": empty address string is not allowed"},
		},
		{
			name: "invalid source",
			msg: MsgRejectPaymentRequest{
				Target:     validMsg.Target,
				Source:     "alsoverybad",
				ExternalId: validMsg.ExternalId,
			},
			expErr: []string{"invalid source \"alsoverybad\": decoding bech32 failed: invalid separator index -1"},
		},
		{
			name: "no external id",
			msg: MsgRejectPaymentRequest{
				Target:     validMsg.Target,
				Source:     validMsg.Source,
				ExternalId: "",
			},
			expErr: nil,
		},
		{
			name: "invalid external id",
			msg: MsgRejectPaymentRequest{
				Target:     validMsg.Target,
				Source:     validMsg.Source,
				ExternalId: "w" + strings.Repeat("o", MaxExternalIDLength) + "w",
			},
			expErr: []string{fmt.Sprintf("invalid external id %q (length %d): max length %d",
				"woooo...oooow", MaxExternalIDLength+2, MaxExternalIDLength)},
		},
		{
			name: "multiple errors",
			msg: MsgRejectPaymentRequest{
				Target:     "reallybad",
				Source:     "",
				ExternalId: "w" + strings.Repeat("o", MaxExternalIDLength) + "w",
			},
			expErr: []string{
				"invalid target \"reallybad\": decoding bech32 failed: invalid separator index -1",
				"invalid source \"\": empty address string is not allowed",
				fmt.Sprintf("invalid external id %q (length %d): max length %d",
					"woooo...oooow", MaxExternalIDLength+2, MaxExternalIDLength),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

func TestMsgRejectPaymentsRequest_ValidateBasic(t *testing.T) {
	target := sdk.AccAddress("target______________").String()
	source0 := sdk.AccAddress("source0_____________").String()
	source1 := sdk.AccAddress("source1_____________").String()
	source2 := sdk.AccAddress("source2_____________").String()

	tests := []struct {
		name   string
		msg    MsgRejectPaymentsRequest
		expErr []string
	}{
		{
			name: "valid: one source",
			msg: MsgRejectPaymentsRequest{
				Target:  target,
				Sources: []string{source0},
			},
			expErr: nil,
		},
		{
			name: "valid: three sources",
			msg: MsgRejectPaymentsRequest{
				Target:  target,
				Sources: []string{source0, source1, source2},
			},
			expErr: nil,
		},
		{
			name: "no target",
			msg: MsgRejectPaymentsRequest{
				Target:  "",
				Sources: []string{source0},
			},
			expErr: []string{"invalid target \"\": empty address string is not allowed"},
		},
		{
			name: "invalid target",
			msg: MsgRejectPaymentsRequest{
				Target:  "oopsnotgonnawork",
				Sources: []string{source0},
			},
			expErr: []string{"invalid target \"oopsnotgonnawork\": decoding bech32 failed: invalid separator index -1"},
		},
		{
			name: "nil sources",
			msg: MsgRejectPaymentsRequest{
				Target:  target,
				Sources: nil,
			},
			expErr: []string{"at least one source is required"},
		},
		{
			name: "empty sources",
			msg: MsgRejectPaymentsRequest{
				Target:  target,
				Sources: []string{},
			},
			expErr: []string{"at least one source is required"},
		},
		{
			name: "one source: empty string",
			msg: MsgRejectPaymentsRequest{
				Target:  target,
				Sources: []string{""},
			},
			expErr: []string{"invalid sources[0] \"\": empty address string is not allowed"},
		},
		{
			name: "one source: invalid",
			msg: MsgRejectPaymentsRequest{
				Target:  target,
				Sources: []string{"diditagain"},
			},
			expErr: []string{"invalid sources[0] \"diditagain\": decoding bech32 failed: invalid separator index -1"},
		},
		{
			name: "three sources: all same",
			msg: MsgRejectPaymentsRequest{
				Target:  target,
				Sources: []string{source1, source1, source1},
			},
			expErr: []string{"invalid sources: duplicate entry " + source1},
		},
		{
			name: "three sources: same first and third",
			msg: MsgRejectPaymentsRequest{
				Target:  target,
				Sources: []string{source0, source1, source0},
			},
			expErr: []string{"invalid sources: duplicate entry " + source0},
		},
		{
			name: "three sources: invalid first",
			msg: MsgRejectPaymentsRequest{
				Target:  target,
				Sources: []string{"thebadzero", source1, source2},
			},
			expErr: []string{"invalid sources[0] \"thebadzero\": decoding bech32 failed: invalid separator index -1"},
		},
		{
			name: "three sources: invalid second",
			msg: MsgRejectPaymentsRequest{
				Target:  target,
				Sources: []string{source0, "thebadone", source2},
			},
			expErr: []string{"invalid sources[1] \"thebadone\": decoding bech32 failed: invalid separator index -1"},
		},
		{
			name: "three sources: invalid third",
			msg: MsgRejectPaymentsRequest{
				Target:  target,
				Sources: []string{source0, source1, "thebadtwo"},
			},
			expErr: []string{"invalid sources[2] \"thebadtwo\": decoding bech32 failed: invalid separator index -1"},
		},
		{
			name: "multiple errors",
			msg: MsgRejectPaymentsRequest{
				Target:  "",
				Sources: []string{"thebadzero", source1, "thebadtwo"},
			},
			expErr: []string{
				"invalid target \"\": empty address string is not allowed",
				"invalid sources[0] \"thebadzero\": decoding bech32 failed: invalid separator index -1",
				"invalid sources[2] \"thebadtwo\": decoding bech32 failed: invalid separator index -1",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

func TestMsgCancelPaymentsRequest_ValidateBasic(t *testing.T) {
	source := sdk.AccAddress("source______________").String()

	tests := []struct {
		name   string
		msg    MsgCancelPaymentsRequest
		expErr []string
	}{
		{
			name: "valid",
			msg: MsgCancelPaymentsRequest{
				Source:      source,
				ExternalIds: []string{"valideid"},
			},
			expErr: nil,
		},
		{
			name: "no source",
			msg: MsgCancelPaymentsRequest{
				Source:      "",
				ExternalIds: []string{"eid"},
			},
			expErr: []string{"invalid source \"\": empty address string is not allowed"},
		},
		{
			name: "invalid source",
			msg: MsgCancelPaymentsRequest{
				Source:      "notanaddr",
				ExternalIds: []string{"eid"},
			},
			expErr: []string{"invalid source \"notanaddr\": decoding bech32 failed: invalid separator index -1"},
		},
		{
			name: "nil external ids",
			msg: MsgCancelPaymentsRequest{
				Source:      source,
				ExternalIds: nil,
			},
			expErr: []string{"at least one external id is required"},
		},
		{
			name: "empty external ids",
			msg: MsgCancelPaymentsRequest{
				Source:      source,
				ExternalIds: []string{},
			},
			expErr: []string{"at least one external id is required"},
		},
		{
			name: "one external id: empty",
			msg: MsgCancelPaymentsRequest{
				Source:      source,
				ExternalIds: []string{""},
			},
			expErr: nil,
		},
		{
			name: "one external id: invalid",
			msg: MsgCancelPaymentsRequest{
				Source:      source,
				ExternalIds: []string{strings.Repeat("w", MaxExternalIDLength+1)},
			},
			expErr: []string{fmt.Sprintf("invalid external id %q (length %d): max length %d",
				"wwwww...wwwww", MaxExternalIDLength+1, MaxExternalIDLength)},
		},
		{
			name: "three external ids: all same",
			msg: MsgCancelPaymentsRequest{
				Source:      source,
				ExternalIds: []string{"same", "same", "same"},
			},
			expErr: []string{"invalid external ids: duplicate entry \"same\""},
		},
		{
			name: "three external ids: same first and third",
			msg: MsgCancelPaymentsRequest{
				Source:      source,
				ExternalIds: []string{"twin", "other", "twin"},
			},
			expErr: []string{"invalid external ids: duplicate entry \"twin\""},
		},
		{
			name: "three external ids: invalid first",
			msg: MsgCancelPaymentsRequest{
				Source: source,
				ExternalIds: []string{
					strings.Repeat("0", MaxExternalIDLength+1),
					strings.Repeat("1", MaxExternalIDLength),
					strings.Repeat("2", MaxExternalIDLength),
				},
			},
			expErr: []string{fmt.Sprintf("invalid external ids[0]: invalid external id %q (length %d): max length %d",
				"00000...00000", MaxExternalIDLength+1, MaxExternalIDLength)},
		},
		{
			name: "three external ids: invalid second",
			msg: MsgCancelPaymentsRequest{
				Source: source,
				ExternalIds: []string{
					strings.Repeat("0", MaxExternalIDLength),
					strings.Repeat("1", MaxExternalIDLength+1),
					strings.Repeat("2", MaxExternalIDLength),
				},
			},
			expErr: []string{fmt.Sprintf("invalid external ids[1]: invalid external id %q (length %d): max length %d",
				"11111...11111", MaxExternalIDLength+1, MaxExternalIDLength)},
		},
		{
			name: "three external ids: invalid third",
			msg: MsgCancelPaymentsRequest{
				Source: source,
				ExternalIds: []string{
					strings.Repeat("0", MaxExternalIDLength),
					strings.Repeat("1", MaxExternalIDLength),
					strings.Repeat("2", MaxExternalIDLength+1),
				},
			},
			expErr: []string{fmt.Sprintf("invalid external ids[2]: invalid external id %q (length %d): max length %d",
				"22222...22222", MaxExternalIDLength+1, MaxExternalIDLength)},
		},
		{
			name: "multiple errors",
			msg: MsgCancelPaymentsRequest{
				Source: "",
				ExternalIds: []string{
					"eid0",
					"(^=~" + strings.Repeat("-", MaxExternalIDLength-7) + "~=^)",
					"eid1",
					"eid0",
				},
			},
			expErr: []string{
				"invalid source \"\": empty address string is not allowed",
				fmt.Sprintf("invalid external ids[1]: invalid external id %q (length %d): max length %d",
					"(^=~-...-~=^)", MaxExternalIDLength+1, MaxExternalIDLength),
				"invalid external ids: duplicate entry \"eid0\"",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

func TestMsgChangePaymentTargetRequest_ValidateBasic(t *testing.T) {
	source := sdk.AccAddress("source______________").String()
	newTarget := sdk.AccAddress("newTarget___________").String()
	eid := "my|1932DA1E-5469-4E47-BCBA-2589877A4860"

	tests := []struct {
		name   string
		msg    MsgChangePaymentTargetRequest
		expErr []string
	}{
		{
			name: "valid",
			msg: MsgChangePaymentTargetRequest{
				Source:     source,
				ExternalId: eid,
				NewTarget:  newTarget,
			},
			expErr: nil,
		},
		{
			name: "no source",
			msg: MsgChangePaymentTargetRequest{
				Source:     "",
				ExternalId: eid,
				NewTarget:  newTarget,
			},
			expErr: []string{"invalid source \"\": empty address string is not allowed"},
		},
		{
			name: "invalid source",
			msg: MsgChangePaymentTargetRequest{
				Source:     "justkidding",
				ExternalId: eid,
				NewTarget:  newTarget,
			},
			expErr: []string{"invalid source \"justkidding\": decoding bech32 failed: invalid separator index -1"},
		},
		{
			name: "empty external id",
			msg: MsgChangePaymentTargetRequest{
				Source:     source,
				ExternalId: "",
				NewTarget:  newTarget,
			},
			expErr: nil,
		},
		{
			name: "invalid external id",
			msg: MsgChangePaymentTargetRequest{
				Source:     source,
				ExternalId: strings.Repeat("e", MaxExternalIDLength+1),
				NewTarget:  newTarget,
			},
			expErr: []string{fmt.Sprintf("invalid external id %q (length %d): max length %d",
				"eeeee...eeeee", MaxExternalIDLength+1, MaxExternalIDLength)},
		},
		{
			name: "empty new target",
			msg: MsgChangePaymentTargetRequest{
				Source:     source,
				ExternalId: eid,
				NewTarget:  "",
			},
			expErr: nil,
		},
		{
			name: "invalid new target",
			msg: MsgChangePaymentTargetRequest{
				Source:     source,
				ExternalId: eid,
				NewTarget:  "mistakenaddr",
			},
			expErr: []string{"invalid new target \"mistakenaddr\": decoding bech32 failed: invalid separator index -1"},
		},
		{
			name: "multiple errors",
			msg: MsgChangePaymentTargetRequest{
				Source:     "",
				ExternalId: strings.Repeat("e", MaxExternalIDLength+1),
				NewTarget:  "mistakenaddr",
			},
			expErr: []string{
				"invalid source \"\": empty address string is not allowed",
				fmt.Sprintf("invalid external id %q (length %d): max length %d",
					"eeeee...eeeee", MaxExternalIDLength+1, MaxExternalIDLength),
				"invalid new target \"mistakenaddr\": decoding bech32 failed: invalid separator index -1",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

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
			expErr: []string{"invalid authority", emptyAddrErr},
		},
		{
			name: "bad authority",
			msg: MsgGovCreateMarketRequest{
				Authority: "bad",
				Market:    validMarket,
			},
			expErr: []string{"invalid authority", bech32Err},
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
			expErr: []string{"invalid authority", "no updates", "market id cannot be zero"},
		},
		{
			name: "no authority",
			msg: MsgGovManageFeesRequest{
				Authority:           "",
				AddFeeCreateAskFlat: []sdk.Coin{coin(1, "nhash")},
			},
			expErr: []string{"invalid authority", emptyAddrErr},
		},
		{
			name: "bad authority",
			msg: MsgGovManageFeesRequest{
				Authority:           "bad",
				AddFeeCreateAskFlat: []sdk.Coin{coin(1, "nhash")},
			},
			expErr: []string{"invalid authority", bech32Err},
		},
		{
			name: "no market id",
			msg: MsgGovManageFeesRequest{
				Authority:           authority,
				AddFeeCreateAskFlat: []sdk.Coin{coin(1, "nhash")},
			},
			expErr: []string{"market id cannot be zero"},
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
			expErr: []string{"cannot add and remove the same create-ask flat fee options 1nhash"},
		},
		{
			name: "invalid add create-commitment flat",
			msg: MsgGovManageFeesRequest{
				Authority:                  authority,
				AddFeeCreateCommitmentFlat: []sdk.Coin{coin(0, "nhash")},
			},
			expErr: []string{`invalid create-commitment flat fee to add option "0nhash": amount cannot be zero`},
		},
		{
			name: "same add and remove create-commitment flat",
			msg: MsgGovManageFeesRequest{
				Authority:                     authority,
				AddFeeCreateCommitmentFlat:    []sdk.Coin{coin(1, "nhash")},
				RemoveFeeCreateCommitmentFlat: []sdk.Coin{coin(1, "nhash")},
			},
			expErr: []string{"cannot add and remove the same create-commitment flat fee options 1nhash"},
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
			expErr: []string{"cannot add and remove the same create-bid flat fee options 1nhash"},
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
			expErr: []string{"cannot add and remove the same seller settlement flat fee options 1nhash"},
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
			expErr: []string{"cannot add and remove the same seller settlement fee ratios 2nhash:1nhash"},
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
			expErr: []string{"cannot add and remove the same buyer settlement flat fee options 1nhash"},
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
			expErr: []string{"cannot add and remove the same buyer settlement fee ratios 2nhash:1nhash"},
		},
		{
			name: "set fee commitment settlement bips too high",
			msg: MsgGovManageFeesRequest{
				Authority:                      authority,
				SetFeeCommitmentSettlementBips: 10_001,
			},
			expErr: []string{"invalid commitment settlement bips 10001: exceeds max of 10000"},
		},
		{
			name: "set fee commitment settlement bips with unset",
			msg: MsgGovManageFeesRequest{
				Authority:                        authority,
				SetFeeCommitmentSettlementBips:   1,
				UnsetFeeCommitmentSettlementBips: true,
			},
			expErr: []string{"invalid commitment settlement bips 1: must be zero when unset_fee_commitment_settlement_bips is true"},
		},
		{
			name: "multiple errors",
			msg: MsgGovManageFeesRequest{
				Authority:                        "",
				AddFeeCreateAskFlat:              []sdk.Coin{coin(0, "nhash")},
				RemoveFeeCreateAskFlat:           []sdk.Coin{coin(0, "nhash")},
				AddFeeCreateBidFlat:              []sdk.Coin{coin(0, "nhash")},
				RemoveFeeCreateBidFlat:           []sdk.Coin{coin(0, "nhash")},
				AddFeeSellerSettlementFlat:       []sdk.Coin{coin(0, "nhash")},
				RemoveFeeSellerSettlementFlat:    []sdk.Coin{coin(0, "nhash")},
				AddFeeSellerSettlementRatios:     []FeeRatio{ratio(1, "nhash", 2, "nhash")},
				RemoveFeeSellerSettlementRatios:  []FeeRatio{ratio(1, "nhash", 2, "nhash")},
				AddFeeBuyerSettlementFlat:        []sdk.Coin{coin(0, "nhash")},
				RemoveFeeBuyerSettlementFlat:     []sdk.Coin{coin(0, "nhash")},
				AddFeeBuyerSettlementRatios:      []FeeRatio{ratio(1, "nhash", 2, "nhash")},
				RemoveFeeBuyerSettlementRatios:   []FeeRatio{ratio(1, "nhash", 2, "nhash")},
				AddFeeCreateCommitmentFlat:       []sdk.Coin{coin(0, "nhash")},
				RemoveFeeCreateCommitmentFlat:    []sdk.Coin{coin(0, "nhash")},
				SetFeeCommitmentSettlementBips:   12345,
				UnsetFeeCommitmentSettlementBips: true,
			},
			expErr: []string{
				"invalid authority", emptyAddrErr,
				"market id cannot be zero",
				`invalid create-ask flat fee to add option "0nhash": amount cannot be zero`,
				"cannot add and remove the same create-ask flat fee options 0nhash",
				`invalid create-bid flat fee to add option "0nhash": amount cannot be zero`,
				"cannot add and remove the same create-bid flat fee options 0nhash",
				`invalid create-commitment flat fee to add option "0nhash": amount cannot be zero`,
				"cannot add and remove the same create-commitment flat fee options 0nhash",
				`invalid seller settlement flat fee to add option "0nhash": amount cannot be zero`,
				"cannot add and remove the same seller settlement flat fee options 0nhash",
				`seller fee ratio fee amount "2nhash" cannot be greater than price amount "1nhash"`,
				"cannot add and remove the same seller settlement fee ratios 1nhash:2nhash",
				`invalid buyer settlement flat fee to add option "0nhash": amount cannot be zero`,
				"cannot add and remove the same buyer settlement flat fee options 0nhash",
				`buyer fee ratio fee amount "2nhash" cannot be greater than price amount "1nhash"`,
				"cannot add and remove the same buyer settlement fee ratios 1nhash:2nhash",
				"invalid commitment settlement bips 12345: exceeds max of 10000",
				"invalid commitment settlement bips 12345: must be zero when unset_fee_commitment_settlement_bips is true",
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
		{
			name: "one add fee create-commitment flat",
			msg:  MsgGovManageFeesRequest{AddFeeCreateCommitmentFlat: oneCoin},
			exp:  true,
		},
		{
			name: "one remove fee create-commitment flat",
			msg:  MsgGovManageFeesRequest{RemoveFeeCreateCommitmentFlat: oneCoin},
			exp:  true,
		},
		{
			name: "set fee commitment settlement bips",
			msg:  MsgGovManageFeesRequest{SetFeeCommitmentSettlementBips: 1},
			exp:  true,
		},
		{
			name: "unset fee commitment settlement bips",
			msg:  MsgGovManageFeesRequest{UnsetFeeCommitmentSettlementBips: true},
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

func TestMsgGovCloseMarketRequest_ValidateBasic(t *testing.T) {
	tests := []struct {
		name   string
		msg    MsgGovCloseMarketRequest
		expErr []string
	}{
		{
			name: "control",
			msg: MsgGovCloseMarketRequest{
				Authority: sdk.AccAddress("authority___________").String(),
				MarketId:  1,
			},
		},
		{
			name: "no authority",
			msg: MsgGovCloseMarketRequest{
				Authority: "",
				MarketId:  1,
			},
			expErr: []string{"invalid authority \"\": " + emptyAddrErr},
		},
		{
			name: "bad authority",
			msg: MsgGovCloseMarketRequest{
				Authority: "notanauthorityaddr",
				MarketId:  1,
			},
			expErr: []string{"invalid authority \"notanauthorityaddr\": " + bech32Err},
		},
		{
			name: "market zero",
			msg: MsgGovCloseMarketRequest{
				Authority: sdk.AccAddress("authority___________").String(),
				MarketId:  0,
			},
			expErr: []string{"invalid market id: cannot be zero"},
		},
		{
			name: "multiple errors",
			msg: MsgGovCloseMarketRequest{
				Authority: "",
				MarketId:  0,
			},
			expErr: []string{
				"invalid authority \"\": " + emptyAddrErr,
				"invalid market id: cannot be zero",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

func TestMsgGovUpdateParamsRequest_ValidateBasic(t *testing.T) {
	pioconfig.SetProvenanceConfig("", 0)
	authority := sdk.AccAddress("authority___________").String()

	tests := []struct {
		name   string
		msg    MsgGovUpdateParamsRequest
		expErr []string
	}{
		{
			name:   "zero value",
			msg:    MsgGovUpdateParamsRequest{},
			expErr: []string{"invalid authority", emptyAddrErr},
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
			expErr: []string{"invalid authority", emptyAddrErr},
		},
		{
			name: "bad authority",
			msg: MsgGovUpdateParamsRequest{
				Authority: "bad",
				Params:    *DefaultParams(),
			},
			expErr: []string{"invalid authority", bech32Err},
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
