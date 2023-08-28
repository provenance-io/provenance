package exchange

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMarket_Validate(t *testing.T) {
	coins := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "sdk.ParseCoinsNormalized(%q)", coins)
		return rv
	}
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}

	addr1 := sdk.AccAddress("addr1_______________").String()
	addr2 := sdk.AccAddress("addr2_______________").String()

	tests := []struct {
		name   string
		market Market
		expErr []string
	}{
		{
			name:   "zero value",
			market: Market{},
			expErr: nil,
		},
		{
			// A little bit of everything that should all be valid.
			name: "control",
			market: Market{
				MarketId:                1,
				MarketDetails:           MarketDetails{Name: "Test Market", Description: "Just a market for testing."},
				FeeCreateAskFlat:        coins("10nnibbler,5mfry"),
				FeeCreateBidFlat:        coins("15nnibbler,5mfry"),
				FeeSettlementSellerFlat: coins("50nnibler,8mfry"),
				FeeSettlementSellerRatios: []*FeeRatio{
					{Price: coin(1000, "nnibler"), Fee: coin(1, "nnibler")},
					{Price: coin(300, "mfry"), Fee: coin(1, "mfry")},
				},
				FeeSettlementBuyerFlat: coins("100nnibler,20mfry"),
				FeeSettlementBuyerRatios: []*FeeRatio{
					{Price: coin(500, "nnibler"), Fee: coin(1, "nnibler")},
					{Price: coin(500, "nnibler"), Fee: coin(8, "mfry")},
					{Price: coin(150, "mfry"), Fee: coin(1, "mfry")},
					{Price: coin(1, "mfry"), Fee: coin(1, "nnibler")},
				},
				AcceptingOrders:     true,
				AllowUserSettlement: true,
				AccessGrants: []*AccessGrant{
					{Address: addr1, Permissions: AllPermissions()},
					{Address: addr2, Permissions: []Permission{Permission_settle}},
				},
				ReqAttrCreateAsk: []string{"kyc.ask.path", "*.ask.some.other.path"},
				ReqAttrCreateBid: []string{"kyc.bid.path", "*.bid.some.other.path"},
			},
			expErr: nil,
		},
		{
			name:   "market id 0",
			market: Market{MarketId: 0},
			expErr: nil,
		},
		{
			name:   "invalid market details",
			market: Market{MarketDetails: MarketDetails{Name: strings.Repeat("n", MaxName+1)}},
			expErr: []string{fmt.Sprintf("name length %d exceeds maximum length of %d", MaxName+1, MaxName)},
		},
		{
			name:   "invalid fee create ask flat",
			market: Market{FeeCreateAskFlat: sdk.Coins{coin(-1, "leela")}},
			expErr: []string{`invalid create ask flat fee option "-1leela": negative coin amount: -1`},
		},
		{
			name:   "invalid fee create bid flat",
			market: Market{FeeCreateBidFlat: sdk.Coins{coin(-1, "leela")}},
			expErr: []string{`invalid create bid flat fee option "-1leela": negative coin amount: -1`},
		},
		{
			name:   "invalid fee settlement seller flat",
			market: Market{FeeSettlementSellerFlat: sdk.Coins{coin(-1, "leela")}},
			expErr: []string{`invalid settlement seller flat fee option "-1leela": negative coin amount: -1`},
		},
		{
			name:   "invalid fee settlement buyer flat",
			market: Market{FeeSettlementBuyerFlat: sdk.Coins{coin(-1, "leela")}},
			expErr: []string{`invalid settlement buyer flat fee option "-1leela": negative coin amount: -1`},
		},
		{
			name:   "invalid seller ratio",
			market: Market{FeeSettlementSellerRatios: []*FeeRatio{{Price: coin(0, "fry"), Fee: coin(1, "fry")}}},
			expErr: []string{`seller fee ratio price amount "0fry" must be positive`},
		},
		{
			name:   "invalid buyer ratio",
			market: Market{FeeSettlementBuyerRatios: []*FeeRatio{{Price: coin(0, "fry"), Fee: coin(1, "fry")}}},
			expErr: []string{`buyer fee ratio price amount "0fry" must be positive`},
		},
		{
			name: "invalid ratios",
			market: Market{
				FeeSettlementSellerRatios: []*FeeRatio{{Price: coin(10, "fry"), Fee: coin(1, "fry")}},
				FeeSettlementBuyerRatios:  []*FeeRatio{{Price: coin(100, "leela"), Fee: coin(1, "leela")}},
			},
			expErr: []string{
				`denom "fry" is defined in the seller settlement fee ratios but not buyer`,
				`denom "leela" is defined in the buyer settlement fee ratios but not seller`,
			},
		},
		{
			name:   "invalid access grants",
			market: Market{AccessGrants: []*AccessGrant{{Address: "bad_addr", Permissions: AllPermissions()}}},
			expErr: []string{"invalid access grant: invalid address: decoding bech32 failed: invalid separator index -1"},
		},
		{
			name:   "invalid ask required attributes",
			market: Market{ReqAttrCreateAsk: []string{"this-attr-is-bad"}},
			expErr: []string{`invalid create ask required attributes: invalid required attribute "this-attr-is-bad"`},
		},
		{
			name:   "invalid bid required attributes",
			market: Market{ReqAttrCreateBid: []string{"this-attr-grrrr"}},
			expErr: []string{`invalid create bid required attributes: invalid required attribute "this-attr-grrrr"`},
		},
		{
			name: "multiple errors",
			market: Market{
				MarketDetails:             MarketDetails{Name: strings.Repeat("n", MaxName+1)},
				FeeCreateAskFlat:          sdk.Coins{coin(-1, "leela")},
				FeeCreateBidFlat:          sdk.Coins{coin(-1, "leela")},
				FeeSettlementSellerFlat:   sdk.Coins{coin(-1, "leela")},
				FeeSettlementBuyerFlat:    sdk.Coins{coin(-1, "leela")},
				FeeSettlementSellerRatios: []*FeeRatio{{Price: coin(10, "fry"), Fee: coin(1, "fry")}},
				FeeSettlementBuyerRatios:  []*FeeRatio{{Price: coin(100, "leela"), Fee: coin(1, "leela")}},
				AccessGrants:              []*AccessGrant{{Address: "bad_addr", Permissions: AllPermissions()}},
				ReqAttrCreateAsk:          []string{"this-attr-is-bad"},
				ReqAttrCreateBid:          []string{"this-attr-grrrr"},
			},
			expErr: []string{
				fmt.Sprintf("name length %d exceeds maximum length of %d", MaxName+1, MaxName),
				`invalid create ask flat fee option "-1leela": negative coin amount: -1`,
				`invalid create bid flat fee option "-1leela": negative coin amount: -1`,
				`invalid settlement seller flat fee option "-1leela": negative coin amount: -1`,
				`invalid settlement buyer flat fee option "-1leela": negative coin amount: -1`,
				`denom "fry" is defined in the seller settlement fee ratios but not buyer`,
				`denom "leela" is defined in the buyer settlement fee ratios but not seller`,
				"invalid access grant: invalid address: decoding bech32 failed: invalid separator index -1",
				`invalid create ask required attributes: invalid required attribute "this-attr-is-bad"`,
				`invalid create bid required attributes: invalid required attribute "this-attr-grrrr"`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.market.Validate()
			}
			require.NotPanics(t, testFunc, "Market.Validate")

			// TODO[1658]: Refactor to testutils.AssertErrorContents(t, err, tc.expErr, "Market.Validate")
			if len(tc.expErr) > 0 {
				if assert.Error(t, err, "Market.Validate") {
					for _, exp := range tc.expErr {
						assert.ErrorContains(t, err, exp, "Market.Validate")
					}
				}
			} else {
				assert.NoError(t, err, "Market.Validate")
			}
		})
	}
}

func TestValidateFeeOptions(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}

	tests := []struct {
		name    string
		field   string
		options []sdk.Coin
		expErr  string
	}{
		{
			name:    "nil options",
			field:   "",
			options: nil,
			expErr:  "",
		},
		{
			name:    "empty options",
			field:   "",
			options: []sdk.Coin{},
			expErr:  "",
		},
		{
			name:    "one option: good",
			field:   "",
			options: []sdk.Coin{coin(1, "leela")},
			expErr:  "",
		},
		{
			name:    "one option: bad denom",
			field:   "test field one",
			options: []sdk.Coin{coin(1, "%")},
			expErr:  `invalid test field one option "1%": invalid denom: %`,
		},
		{
			name:    "one option: zero amount",
			field:   "zero-amount",
			options: []sdk.Coin{coin(0, "fry")},
			expErr:  `invalid zero-amount option "0fry": amount cannot be zero`,
		},
		{
			name:    "one option: negative amount",
			field:   "i-pay-u",
			options: []sdk.Coin{coin(-1, "nibbler")},
			expErr:  `invalid i-pay-u option "-1nibbler": negative coin amount: -1`,
		},
		{
			name:    "three options: all good",
			field:   "",
			options: []sdk.Coin{coin(5, "fry"), coin(2, "leela"), coin(1, "farnsworth")},
			expErr:  "",
		},
		{
			name:    "three options: bad first",
			field:   "coffee",
			options: []sdk.Coin{coin(0, "fry"), coin(2, "leela"), coin(1, "farnsworth")},
			expErr:  `invalid coffee option "0fry": amount cannot be zero`,
		},
		{
			name:    "three options: bad second",
			field:   "eyeballs",
			options: []sdk.Coin{coin(5, "fry"), coin(0, "leela"), coin(1, "farnsworth")},
			expErr:  `invalid eyeballs option "0leela": amount cannot be zero`,
		},
		{
			name:    "three options: bad third",
			field:   "eyeballs",
			options: []sdk.Coin{coin(5, "fry"), coin(2, "leela"), coin(0, "farnsworth")},
			expErr:  `invalid eyeballs option "0farnsworth": amount cannot be zero`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = ValidateFeeOptions(tc.field, tc.options)
			}
			require.NotPanics(t, testFunc, "ValidateFeeOptions")

			// TODO[1658]: Refactor this to testutils.AssertErrorValue(t, err, tc.expErr, "ValidateFeeOptions")
			if len(tc.expErr) > 0 {
				assert.EqualError(t, err, tc.expErr, "ValidateFeeOptions")
			} else {
				assert.NoError(t, err, "ValidateFeeOptions")
			}
		})
	}
}

func TestMarketDetails_Validate(t *testing.T) {
	joinErrs := func(errs ...string) string {
		return strings.Join(errs, "\n")
	}
	nameErr := func(over int) string {
		return fmt.Sprintf("name length %d exceeds maximum length of %d", MaxName+over, MaxName)
	}
	descErr := func(over int) string {
		return fmt.Sprintf("description length %d exceeds maximum length of %d", MaxDescription+over, MaxDescription)
	}
	urlErr := func(over int) string {
		return fmt.Sprintf("website_url length %d exceeds maximum length of %d", MaxWebsiteURL+over, MaxWebsiteURL)
	}
	iconErr := func(over int) string {
		return fmt.Sprintf("icon_uri length %d exceeds maximum length of %d", MaxIconURI+over, MaxIconURI)
	}

	tests := []struct {
		name    string
		details MarketDetails
		expErr  string
	}{
		{
			name:    "zero value",
			details: MarketDetails{},
			expErr:  "",
		},
		{
			name:    "name at max length",
			details: MarketDetails{Name: strings.Repeat("n", MaxName)},
			expErr:  "",
		},
		{
			name:    "name over max length",
			details: MarketDetails{Name: strings.Repeat("n", MaxName+1)},
			expErr:  nameErr(1),
		},
		{
			name:    "description at max length",
			details: MarketDetails{Description: strings.Repeat("d", MaxDescription)},
			expErr:  "",
		},
		{
			name:    "description over max length",
			details: MarketDetails{Description: strings.Repeat("d", MaxDescription+1)},
			expErr:  descErr(1),
		},
		{
			name:    "website url at max length",
			details: MarketDetails{WebsiteUrl: strings.Repeat("w", MaxWebsiteURL)},
			expErr:  "",
		},
		{
			name:    "website url over max length",
			details: MarketDetails{WebsiteUrl: strings.Repeat("w", MaxWebsiteURL+1)},
			expErr:  urlErr(1),
		},
		{
			name:    "icon uri at max length",
			details: MarketDetails{IconUri: strings.Repeat("i", MaxIconURI)},
			expErr:  "",
		},
		{
			name:    "icon uri over max length",
			details: MarketDetails{IconUri: strings.Repeat("i", MaxIconURI+1)},
			expErr:  iconErr(1),
		},
		{
			name: "all at max",
			details: MarketDetails{
				Name:        strings.Repeat("N", MaxName),
				Description: strings.Repeat("D", MaxDescription),
				WebsiteUrl:  strings.Repeat("W", MaxWebsiteURL),
				IconUri:     strings.Repeat("I", MaxIconURI),
			},
			expErr: "",
		},
		{
			name: "multiple errors",
			details: MarketDetails{
				Name:        strings.Repeat("N", MaxName+2),
				Description: strings.Repeat("D", MaxDescription+3),
				WebsiteUrl:  strings.Repeat("W", MaxWebsiteURL+4),
				IconUri:     strings.Repeat("I", MaxIconURI+5),
			},
			expErr: joinErrs(nameErr(2), descErr(3), urlErr(4), iconErr(5)),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.details.Validate()
			}
			require.NotPanics(t, testFunc, "Validate")

			// TODO[1658]: Refactor to testutils.AssertErrorValue(t, err, tc.expErr, "Validate")
			if len(tc.expErr) > 0 {
				assert.EqualError(t, err, tc.expErr, "Validate")
			} else {
				assert.NoError(t, err, "Validate")
			}
		})
	}
}

func TestValidateFeeRatios(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	joinErrs := func(errs ...string) string {
		return strings.Join(errs, "\n")
	}

	tests := []struct {
		name         string
		sellerRatios []*FeeRatio
		buyerRatios  []*FeeRatio
		exp          string
	}{
		{
			name:         "nil nil",
			sellerRatios: nil,
			buyerRatios:  nil,
			exp:          "",
		},
		{
			name:         "nil empty",
			sellerRatios: nil,
			buyerRatios:  []*FeeRatio{},
			exp:          "",
		},
		{
			name:         "empty nil",
			sellerRatios: []*FeeRatio{},
			buyerRatios:  nil,
			exp:          "",
		},
		{
			name:         "empty empty",
			sellerRatios: []*FeeRatio{},
			buyerRatios:  []*FeeRatio{},
			exp:          "",
		},
		{
			name:         "nil seller entry",
			sellerRatios: []*FeeRatio{nil},
			buyerRatios:  nil,
			exp:          "nil seller fee ratio not allowed",
		},
		{
			name:         "nil buyer entry",
			sellerRatios: nil,
			buyerRatios:  []*FeeRatio{nil},
			exp:          "nil buyer fee ratio not allowed",
		},
		{
			name: "multiple errors from sellers and buyers",
			sellerRatios: []*FeeRatio{
				nil,
				{Price: coin(0, "leela"), Fee: coin(1, "leela")},
			},
			buyerRatios: []*FeeRatio{
				nil,
				{Price: coin(0, "leela"), Fee: coin(1, "leela")},
			},
			exp: joinErrs(
				"nil seller fee ratio not allowed",
				`seller fee ratio price amount "0leela" must be positive`,
				"nil buyer fee ratio not allowed",
				`buyer fee ratio price amount "0leela" must be positive`,
			),
		},
		{
			name: "sellers have price denom that buyers do not",
			sellerRatios: []*FeeRatio{
				{Price: coin(100, "leela"), Fee: coin(1, "leela")},
				{Price: coin(500, "fry"), Fee: coin(3, "fry")},
			},
			buyerRatios: []*FeeRatio{
				{Price: coin(100, "leela"), Fee: coin(3, "leela")},
			},
			exp: `denom "fry" is defined in the seller settlement fee ratios but not buyer`,
		},
		{
			name: "sellers have two price denoms that buyers do not",
			sellerRatios: []*FeeRatio{
				{Price: coin(100, "leela"), Fee: coin(1, "leela")},
				{Price: coin(500, "fry"), Fee: coin(3, "fry")},
			},
			buyerRatios: nil,
			exp: joinErrs(
				`denom "leela" is defined in the seller settlement fee ratios but not buyer`,
				`denom "fry" is defined in the seller settlement fee ratios but not buyer`,
			),
		},
		{
			name: "buyers have price denom that sellers do not",
			sellerRatios: []*FeeRatio{
				{Price: coin(100, "leela"), Fee: coin(3, "leela")},
			},
			buyerRatios: []*FeeRatio{
				{Price: coin(100, "leela"), Fee: coin(1, "leela")},
				{Price: coin(500, "fry"), Fee: coin(3, "leela")},
			},
			exp: `denom "fry" is defined in the buyer settlement fee ratios but not seller`,
		},
		{
			name:         "buyers have two price denoms that sellers do not",
			sellerRatios: nil,
			buyerRatios: []*FeeRatio{
				{Price: coin(100, "leela"), Fee: coin(1, "leela")},
				{Price: coin(500, "fry"), Fee: coin(3, "leela")},
			},
			exp: joinErrs(
				`denom "leela" is defined in the buyer settlement fee ratios but not seller`,
				`denom "fry" is defined in the buyer settlement fee ratios but not seller`,
			),
		},
		{
			name: "two buyers and two sellers and four price denoms",
			sellerRatios: []*FeeRatio{
				{Price: coin(100, "leela"), Fee: coin(1, "leela")},
				{Price: coin(500, "fry"), Fee: coin(3, "fry")},
			},
			buyerRatios: []*FeeRatio{
				{Price: coin(100, "bender"), Fee: coin(1, "leela")},
				{Price: coin(3, "professor"), Fee: coin(500, "fry")},
			},
			exp: joinErrs(
				`denom "leela" is defined in the seller settlement fee ratios but not buyer`,
				`denom "fry" is defined in the seller settlement fee ratios but not buyer`,
				`denom "bender" is defined in the buyer settlement fee ratios but not seller`,
				`denom "professor" is defined in the buyer settlement fee ratios but not seller`,
			),
		},
		{
			name: "three seller ratios and many buyer ratios all legit",
			sellerRatios: []*FeeRatio{
				{Price: coin(100, "leela"), Fee: coin(1, "leela")},
				{Price: coin(500, "fry"), Fee: coin(3, "fry")},
				{Price: coin(300, "bender"), Fee: coin(7, "bender")},
			},
			buyerRatios: []*FeeRatio{
				{Price: coin(10, "leela"), Fee: coin(1, "leela")},
				{Price: coin(11, "leela"), Fee: coin(2, "fry")},
				{Price: coin(12, "leela"), Fee: coin(3, "bender")},
				{Price: coin(1, "leela"), Fee: coin(3, "professor")},
				{Price: coin(50, "fry"), Fee: coin(1, "leela")},
				{Price: coin(51, "fry"), Fee: coin(2, "fry")},
				{Price: coin(52, "fry"), Fee: coin(3, "bender")},
				{Price: coin(1, "fry"), Fee: coin(2, "professor")},
				{Price: coin(30, "bender"), Fee: coin(1, "leela")},
				{Price: coin(31, "bender"), Fee: coin(2, "fry")},
				{Price: coin(32, "bender"), Fee: coin(3, "bender")},
				{Price: coin(1, "bender"), Fee: coin(1, "professor")},
			},
			exp: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = ValidateFeeRatios(tc.sellerRatios, tc.buyerRatios)
			}
			require.NotPanics(t, testFunc, "ValidateFeeRatios")

			// TODO[1658]: Refactor to testutils.AssertErrorValue(t, err, tc.exp, "ValidateFeeRatios")
			if len(tc.exp) > 0 {
				assert.EqualError(t, err, tc.exp, "ValidateFeeRatios")
			} else {
				assert.NoError(t, err, "ValidateFeeRatios")
			}
		})
	}
}

func TestValidateSellerFeeRatios(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	joinErrs := func(errs ...string) string {
		return strings.Join(errs, "\n")
	}

	tests := []struct {
		name   string
		ratios []*FeeRatio
		exp    string
	}{
		{
			name:   "nil ratios",
			ratios: nil,
			exp:    "",
		},
		{
			name:   "empty ratios",
			ratios: []*FeeRatio{},
			exp:    "",
		},
		{
			name:   "one ratio: nil",
			ratios: []*FeeRatio{nil},
			exp:    "nil seller fee ratio not allowed",
		},
		{
			name:   "one ratio: different denoms",
			ratios: []*FeeRatio{{Price: coin(3, "hermes"), Fee: coin(2, "mom")}},
			exp:    `seller fee ratio price denom "hermes" does not equal fee denom "mom"`,
		},
		{
			name:   "one ratio: same denoms",
			ratios: []*FeeRatio{{Price: coin(3, "mom"), Fee: coin(2, "mom")}},
			exp:    "",
		},
		{
			name:   "one ratio: invalid",
			ratios: []*FeeRatio{{Price: coin(0, "hermes"), Fee: coin(2, "hermes")}},
			exp:    `seller fee ratio price amount "0hermes" must be positive`,
		},
		{
			name: "two with same denom",
			ratios: []*FeeRatio{
				{Price: coin(3, "hermes"), Fee: coin(2, "hermes")},
				{Price: coin(6, "hermes"), Fee: coin(4, "hermes")},
			},
			exp: `seller fee ratio denom "hermes" appears in multiple ratios`,
		},
		{
			name: "three with diffrent denoms",
			ratios: []*FeeRatio{
				{Price: coin(30, "leela"), Fee: coin(1, "leela")},
				{Price: coin(5, "fry"), Fee: coin(1, "fry")},
				{Price: coin(100, "professor"), Fee: coin(1, "professor")},
			},
			exp: "",
		},
		{
			name: "multiple errors",
			ratios: []*FeeRatio{
				{Price: coin(3, "mom"), Fee: coin(2, "hermes")},
				{Price: coin(0, "hermes"), Fee: coin(2, "hermes")},
				{Price: coin(6, "bender"), Fee: coin(4, "bender")},
				nil,
				{Price: coin(1, "hermes"), Fee: coin(2, "hermes")},
				{Price: coin(2, "bender"), Fee: coin(1, "bender")},
				// This one is ignored because we've already complained about multiple hermes.
				{Price: coin(30, "hermes"), Fee: coin(2, "hermes")},
			},
			exp: joinErrs(
				`seller fee ratio price denom "mom" does not equal fee denom "hermes"`,
				`seller fee ratio price amount "0hermes" must be positive`,
				"nil seller fee ratio not allowed",
				`seller fee ratio denom "hermes" appears in multiple ratios`,
				`seller fee ratio denom "bender" appears in multiple ratios`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = ValidateSellerFeeRatios(tc.ratios)
			}
			require.NotPanics(t, testFunc, "ValidateSellerFeeRatios")

			// TODO[1658]: Refactor to testutils.AssertErrorValue(t, err, tc.exp, "ValidateSellerFeeRatios")
			if len(tc.exp) > 0 {
				assert.EqualError(t, err, tc.exp, "ValidateSellerFeeRatios")
			} else {
				assert.NoError(t, err, "ValidateSellerFeeRatios")
			}
		})
	}
}

func TestValidateBuyerFeeRatios(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	joinErrs := func(errs ...string) string {
		return strings.Join(errs, "\n")
	}

	tests := []struct {
		name   string
		ratios []*FeeRatio
		exp    string
	}{
		{
			name:   "nil ratios",
			ratios: nil,
			exp:    "",
		},
		{
			name:   "empty ratios",
			ratios: []*FeeRatio{},
			exp:    "",
		},
		{
			name:   "one ratio: nil",
			ratios: []*FeeRatio{nil},
			exp:    "nil buyer fee ratio not allowed",
		},
		{
			name:   "one ratio: different denoms",
			ratios: []*FeeRatio{{Price: coin(3, "hermes"), Fee: coin(2, "mom")}},
			exp:    "",
		},
		{
			name:   "one ratio: same denoms",
			ratios: []*FeeRatio{{Price: coin(3, "mom"), Fee: coin(2, "mom")}},
			exp:    "",
		},
		{
			name:   "one ratio: invalid",
			ratios: []*FeeRatio{{Price: coin(0, "hermes"), Fee: coin(2, "hermes")}},
			exp:    `buyer fee ratio price amount "0hermes" must be positive`,
		},
		{
			name: "duplicate ratio denoms",
			ratios: []*FeeRatio{
				{Price: coin(10, "morbo"), Fee: coin(2, "scruffy")},
				{Price: coin(3, "morbo"), Fee: coin(1, "scruffy")},
			},
			exp: `buyer fee ratio pair "morbo" to "scruffy" appears in multiple ratios`,
		},
		{
			name: "two ratios one each way",
			ratios: []*FeeRatio{
				{Price: coin(10, "leela"), Fee: coin(2, "scruffy")},
				{Price: coin(2, "scruffy"), Fee: coin(8, "leela")},
			},
			exp: "",
		},
		{
			name: "multiple errors",
			ratios: []*FeeRatio{
				{Price: coin(10, "morbo"), Fee: coin(2, "scruffy")},
				{Price: coin(0, "zoidberg"), Fee: coin(1, "amy")},
				{Price: coin(1, "hermes"), Fee: coin(2, "hermes")},
				nil,
				{Price: coin(3, "morbo"), Fee: coin(1, "scruffy")},
				{Price: coin(0, "zoidberg"), Fee: coin(1, "amy")},
				// This one has a different fee denom, though, so it's checked.
				{Price: coin(1, "zoidberg"), Fee: coin(-1, "fry")},
				// We've already complained about this one, so it doesn't happen again.
				{Price: coin(12, "zoidberg"), Fee: coin(55, "amy")},
			},
			exp: joinErrs(
				`buyer fee ratio price amount "0zoidberg" must be positive`,
				`buyer fee ratio fee amount "2hermes" cannot be greater than price amount "1hermes"`,
				"nil buyer fee ratio not allowed",
				`buyer fee ratio pair "morbo" to "scruffy" appears in multiple ratios`,
				`buyer fee ratio pair "zoidberg" to "amy" appears in multiple ratios`,
				`buyer fee ratio fee amount "-1fry" cannot be negative`,
			),
		},
		{
			name: "two different price denoms to several fee denoms",
			ratios: []*FeeRatio{
				{Price: coin(100, "fry"), Fee: coin(1, "fry")},
				{Price: coin(1000, "fry"), Fee: coin(1, "professor")},
				{Price: coin(1, "fry"), Fee: coin(1, "leela")},
				{Price: coin(25, "fry"), Fee: coin(4, "bender")},
				{Price: coin(10, "leela"), Fee: coin(1, "fry")},
				{Price: coin(100, "leela"), Fee: coin(1, "professor")},
				{Price: coin(1000, "leela"), Fee: coin(1, "leela")},
				{Price: coin(35, "leela"), Fee: coin(2, "bender")},
			},
			exp: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = ValidateBuyerFeeRatios(tc.ratios)
			}
			require.NotPanics(t, testFunc, "ValidateBuyerFeeRatios")

			// TODO[1658]: Refactor to testutils.AssertErrorValue(t, err, tc.exp, "ValidateBuyerFeeRatios")
			if len(tc.exp) > 0 {
				assert.EqualError(t, err, tc.exp, "ValidateBuyerFeeRatios")
			} else {
				assert.NoError(t, err, "ValidateBuyerFeeRatios")
			}
		})
	}
}

func TestFeeRatio_String(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}

	tests := []struct {
		name string
		r    FeeRatio
		exp  string
	}{
		{
			name: "zero value",
			r:    FeeRatio{},
			exp:  "<nil>:<nil>",
		},
		{
			name: "same denoms",
			r:    FeeRatio{Price: coin(3, "zapp"), Fee: coin(5, "zapp")},
			exp:  "3zapp:5zapp",
		},
		{
			name: "different denoms",
			r:    FeeRatio{Price: coin(10, "kif"), Fee: coin(5, "zapp")},
			exp:  "10kif:5zapp",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual string
			testFunc := func() {
				actual = tc.r.String()
			}
			require.NotPanics(t, testFunc, "FeeRatio.String()")
			assert.Equal(t, tc.exp, actual, "FeeRatio.String()")
		})
	}
}

func TestFeeRatio_Validate(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}

	tests := []struct {
		name string
		r    FeeRatio
		exp  string
	}{
		{
			name: "zero price amount",
			r:    FeeRatio{Price: coin(0, "fry"), Fee: coin(0, "leela")},
			exp:  `price amount "0fry" must be positive`,
		},
		{
			name: "negative price amount",
			r:    FeeRatio{Price: coin(-1, "fry"), Fee: coin(0, "leela")},
			exp:  `price amount "-1fry" must be positive`,
		},
		{
			name: "negative fee amount",
			r:    FeeRatio{Price: coin(1, "fry"), Fee: coin(-1, "leela")},
			exp:  `fee amount "-1leela" cannot be negative`,
		},
		{
			name: "same price and fee",
			r:    FeeRatio{Price: coin(1, "fry"), Fee: coin(1, "fry")},
			exp:  "",
		},
		{
			name: "same denom fee greater",
			r:    FeeRatio{Price: coin(1, "fry"), Fee: coin(2, "fry")},
			exp:  `fee amount "2fry" cannot be greater than price amount "1fry"`,
		},
		{
			name: "same denom price greater",
			r:    FeeRatio{Price: coin(2, "fry"), Fee: coin(1, "fry")},
			exp:  "",
		},
		{
			name: "different denoms fee amount greater",
			r:    FeeRatio{Price: coin(1, "fry"), Fee: coin(2, "leela")},
			exp:  "",
		},
		{
			name: "different denoms price amount greater",
			r:    FeeRatio{Price: coin(2, "fry"), Fee: coin(1, "leela")},
			exp:  "",
		},
		{
			name: "different denoms same amounts",
			r:    FeeRatio{Price: coin(1, "fry"), Fee: coin(1, "leela")},
			exp:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.r.Validate()

			// TODO[1658]: Refactor to use testutils.AssertErrorValue(t, err, tc.exp, "Validate")
			if len(tc.exp) > 0 {
				assert.EqualError(t, err, tc.exp, "Validate")
			} else {
				assert.NoError(t, err, "Validate")
			}
		})
	}
}

func TestValidateAccessGrants(t *testing.T) {
	joinErrs := func(errs ...string) string {
		return strings.Join(errs, "\n")
	}
	addrDup := sdk.AccAddress("duplicate_address___").String()
	addr1 := sdk.AccAddress("address_1___________").String()
	addr2 := sdk.AccAddress("address_2___________").String()
	addr3 := sdk.AccAddress("address_3___________").String()

	tests := []struct {
		name   string
		grants []*AccessGrant
		exp    string
	}{
		{
			name:   "nil grants",
			grants: nil,
			exp:    "",
		},
		{
			name:   "empty grants",
			grants: []*AccessGrant{},
			exp:    "",
		},
		{
			name:   "nil entry",
			grants: []*AccessGrant{nil},
			exp:    "nil access grant not allowed",
		},
		{
			name: "duplicate address",
			grants: []*AccessGrant{
				{Address: addrDup, Permissions: []Permission{Permission_settle}},
				{Address: addrDup, Permissions: []Permission{Permission_cancel}},
			},
			exp: sdk.AccAddress("duplicate_address___").String() + " appears in multiple access grant entries",
		},
		{
			name: "three entries: all valid",
			grants: []*AccessGrant{
				{Address: addr1, Permissions: AllPermissions()},
				{Address: addr2, Permissions: AllPermissions()},
				{Address: addr3, Permissions: AllPermissions()},
			},
			exp: "",
		},
		{
			name: "three entries: invalid first",
			grants: []*AccessGrant{
				{Address: addr1, Permissions: []Permission{-1}},
				{Address: addr2, Permissions: AllPermissions()},
				{Address: addr3, Permissions: AllPermissions()},
			},
			exp: "invalid access grant: permission -1 does not exist for " + addr1,
		},
		{
			name: "three entries: invalid second",
			grants: []*AccessGrant{
				{Address: addr1, Permissions: AllPermissions()},
				{Address: addr2, Permissions: []Permission{-1}},
				{Address: addr3, Permissions: AllPermissions()},
			},
			exp: "invalid access grant: permission -1 does not exist for " + addr2,
		},
		{
			name: "three entries: invalid second",
			grants: []*AccessGrant{
				{Address: addr1, Permissions: AllPermissions()},
				{Address: addr2, Permissions: AllPermissions()},
				{Address: addr3, Permissions: []Permission{-1}},
			},
			exp: "invalid access grant: permission -1 does not exist for " + addr3,
		},
		{
			name: "three entries: only valid first",
			grants: []*AccessGrant{
				{Address: addr1, Permissions: AllPermissions()},
				{Address: addr2, Permissions: []Permission{0}},
				{Address: addr3, Permissions: []Permission{-1}},
			},
			exp: joinErrs(
				"invalid access grant: permission is unspecified for "+addr2,
				"invalid access grant: permission -1 does not exist for "+addr3,
			),
		},
		{
			name: "three entries: only valid second",
			grants: []*AccessGrant{
				{Address: addr1, Permissions: []Permission{0}},
				{Address: addr2, Permissions: AllPermissions()},
				{Address: addr3, Permissions: []Permission{-1}},
			},
			exp: joinErrs(
				"invalid access grant: permission is unspecified for "+addr1,
				"invalid access grant: permission -1 does not exist for "+addr3,
			),
		},
		{
			name: "three entries: only valid third",
			grants: []*AccessGrant{
				{Address: addr1, Permissions: []Permission{0}},
				{Address: addr2, Permissions: []Permission{-1}},
				{Address: addr3, Permissions: AllPermissions()},
			},
			exp: joinErrs(
				"invalid access grant: permission is unspecified for "+addr1,
				"invalid access grant: permission -1 does not exist for "+addr2,
			),
		},
		{
			name: "three entries: all same address",
			grants: []*AccessGrant{
				{Address: addrDup, Permissions: AllPermissions()},
				{Address: addrDup, Permissions: AllPermissions()},
				{Address: addrDup, Permissions: AllPermissions()},
			},
			exp: addrDup + " appears in multiple access grant entries",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateAccessGrants(tc.grants)

			// TODO[1658]: Refactor to use testutils.AssertErrorValue(t, err, tc.exp, "ValidateAccessGrants")
			if len(tc.exp) > 0 {
				assert.EqualError(t, err, tc.exp, "ValidateAccessGrants")
			} else {
				assert.NoError(t, err, "ValidateAccessGrants")
			}
		})
	}
}

func TestAccessGrant_Validate(t *testing.T) {
	addr := sdk.AccAddress("addr________________").String()
	tests := []struct {
		name string
		a    AccessGrant
		exp  string
	}{
		{
			name: "control",
			a:    AccessGrant{Address: addr, Permissions: AllPermissions()},
			exp:  "",
		},
		{
			name: "invalid address",
			a:    AccessGrant{Address: "invalid_address_____", Permissions: []Permission{Permission_settle}},
			exp:  "invalid access grant: invalid address: decoding bech32 failed: invalid separator index -1",
		},
		{
			name: "nil permissions",
			a:    AccessGrant{Address: addr, Permissions: nil},
			exp:  "invalid access grant: no permissions provided for " + addr,
		},
		{
			name: "empty permissions",
			a:    AccessGrant{Address: addr, Permissions: []Permission{}},
			exp:  "invalid access grant: no permissions provided for " + addr,
		},
		{
			name: "duplicate entry",
			a: AccessGrant{
				Address: addr,
				Permissions: []Permission{
					Permission_settle,
					Permission_cancel,
					Permission_settle,
				},
			},
			exp: "invalid access grant: settle appears multiple times for " + addr,
		},
		{
			name: "invalid entry",
			a: AccessGrant{
				Address: addr,
				Permissions: []Permission{
					Permission_withdraw,
					-1,
					Permission_attributes,
				},
			},
			exp: "invalid access grant: permission -1 does not exist for " + addr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.a.Validate()

			// TODO[1658]: Refactor this to testutils.AssertErrorValue(t, err, tc.exp, "Validate")
			if len(tc.exp) > 0 {
				assert.EqualError(t, err, tc.exp, "Validate")
			} else {
				assert.NoError(t, err, "Validate")
			}
		})
	}
}

func TestPermission_SimpleString(t *testing.T) {
	tests := []struct {
		name string
		p    Permission
		exp  string
	}{
		{
			name: "unspecified",
			p:    Permission_unspecified,
			exp:  "unspecified",
		},
		{
			name: "settle",
			p:    Permission_settle,
			exp:  "settle",
		},
		{
			name: "cancel",
			p:    Permission_cancel,
			exp:  "cancel",
		},
		{
			name: "withdraw",
			p:    Permission_withdraw,
			exp:  "withdraw",
		},
		{
			name: "update",
			p:    Permission_update,
			exp:  "update",
		},
		{
			name: "permissions",
			p:    Permission_permissions,
			exp:  "permissions",
		},
		{
			name: "attributes",
			p:    Permission_attributes,
			exp:  "attributes",
		},
		{
			name: "negative 1",
			p:    -1,
			exp:  "-1",
		},
		{
			name: "unknown value",
			p:    99,
			exp:  "99",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.p.SimpleString()
			assert.Equal(t, tc.exp, actual, "%s.SimpleString()", tc.p)
		})
	}
}

func TestPermission_Validate(t *testing.T) {
	tests := []struct {
		name string
		p    Permission
		exp  string
	}{
		{
			name: "unspecified",
			p:    Permission_unspecified,
			exp:  "permission is unspecified",
		},
		{
			name: "settle",
			p:    Permission_settle,
			exp:  "",
		},
		{
			name: "cancel",
			p:    Permission_cancel,
			exp:  "",
		},
		{
			name: "withdraw",
			p:    Permission_withdraw,
			exp:  "",
		},
		{
			name: "update",
			p:    Permission_update,
			exp:  "",
		},
		{
			name: "permissions",
			p:    Permission_permissions,
			exp:  "",
		},
		{
			name: "attributes",
			p:    Permission_attributes,
			exp:  "",
		},
		{
			name: "negative 1",
			p:    -1,
			exp:  "permission -1 does not exist",
		},
		{
			name: "unknown value",
			p:    99,
			exp:  "permission 99 does not exist",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.p.Validate()

			// TODO: Refactor this to testutils.AssertErrorValue(t, err, tc.exp, "Validate()")
			if len(tc.exp) > 0 {
				assert.EqualError(t, err, tc.exp, "Validate()")
			} else {
				assert.NoError(t, err, "Validate()")
			}
		})
	}

	t.Run("all values have a test case", func(t *testing.T) {
		allVals := maps.Keys(Permission_name)
		sort.Slice(allVals, func(i, j int) bool {
			return allVals[i] < allVals[j]
		})

		for _, val := range allVals {
			perm := Permission(val)
			hasTest := false
			for _, tc := range tests {
				if tc.p == perm {
					hasTest = true
					break
				}
			}
			assert.True(t, hasTest, "No test case found that expects the %s permission", perm)
		}
	})
}

func TestAllPermissions(t *testing.T) {
	expected := []Permission{
		Permission_settle,
		Permission_cancel,
		Permission_withdraw,
		Permission_update,
		Permission_permissions,
		Permission_attributes,
	}

	actual := AllPermissions()
	assert.Equal(t, expected, actual, "AllPermissions()")
}

func TestParsePermission(t *testing.T) {
	tests := []struct {
		permission string
		expected   Permission
		expErr     string
	}{
		// Permission_settle
		{permission: "settle", expected: Permission_settle},
		{permission: " settle", expected: Permission_settle},
		{permission: "settle ", expected: Permission_settle},
		{permission: "SETTLE", expected: Permission_settle},
		{permission: "SeTTle", expected: Permission_settle},
		{permission: "permission_settle", expected: Permission_settle},
		{permission: "PERMISSION_SETTLE", expected: Permission_settle},
		{permission: "pERmiSSion_seTTle", expected: Permission_settle},

		// Permission_cancel
		{permission: "cancel", expected: Permission_cancel},
		{permission: " cancel", expected: Permission_cancel},
		{permission: "cancel ", expected: Permission_cancel},
		{permission: "CANCEL", expected: Permission_cancel},
		{permission: "caNCel", expected: Permission_cancel},
		{permission: "permission_cancel", expected: Permission_cancel},
		{permission: "PERMISSION_CANCEL", expected: Permission_cancel},
		{permission: "pERmiSSion_CanCEl", expected: Permission_cancel},

		// Permission_withdraw
		{permission: "withdraw", expected: Permission_withdraw},
		{permission: " withdraw", expected: Permission_withdraw},
		{permission: "withdraw ", expected: Permission_withdraw},
		{permission: "WITHDRAW", expected: Permission_withdraw},
		{permission: "wiTHdRaw", expected: Permission_withdraw},
		{permission: "permission_withdraw", expected: Permission_withdraw},
		{permission: "PERMISSION_WITHDRAW", expected: Permission_withdraw},
		{permission: "pERmiSSion_wIThdrAw", expected: Permission_withdraw},

		// Permission_update
		{permission: "update", expected: Permission_update},
		{permission: " update", expected: Permission_update},
		{permission: "update ", expected: Permission_update},
		{permission: "UPDATE", expected: Permission_update},
		{permission: "uPDaTe", expected: Permission_update},
		{permission: "permission_update", expected: Permission_update},
		{permission: "PERMISSION_UPDATE", expected: Permission_update},
		{permission: "pERmiSSion_UpdAtE", expected: Permission_update},

		// Permission_permissions
		{permission: "permissions", expected: Permission_permissions},
		{permission: " permissions", expected: Permission_permissions},
		{permission: "permissions ", expected: Permission_permissions},
		{permission: "PERMISSIONS", expected: Permission_permissions},
		{permission: "pErmiSSions", expected: Permission_permissions},
		{permission: "permission_permissions", expected: Permission_permissions},
		{permission: "PERMISSION_PERMISSIONS", expected: Permission_permissions},
		{permission: "pERmiSSion_perMIssIons", expected: Permission_permissions},

		// Permission_attributes
		{permission: "attributes", expected: Permission_attributes},
		{permission: " attributes", expected: Permission_attributes},
		{permission: "attributes ", expected: Permission_attributes},
		{permission: "ATTRIBUTES", expected: Permission_attributes},
		{permission: "aTTribuTes", expected: Permission_attributes},
		{permission: "permission_attributes", expected: Permission_attributes},
		{permission: "PERMISSION_ATTRIBUTES", expected: Permission_attributes},
		{permission: "pERmiSSion_attRiButes", expected: Permission_attributes},

		// Permission_unspecified
		{permission: "unspecified", expErr: `invalid permission: "unspecified"`},
		{permission: " unspecified", expErr: `invalid permission: " unspecified"`},
		{permission: "unspecified ", expErr: `invalid permission: "unspecified "`},
		{permission: "UNSPECIFIED", expErr: `invalid permission: "UNSPECIFIED"`},
		{permission: "unsPeCiFied", expErr: `invalid permission: "unsPeCiFied"`},
		{permission: "permission_unspecified", expErr: `invalid permission: "permission_unspecified"`},
		{permission: "PERMISSION_UNSPECIFIED", expErr: `invalid permission: "PERMISSION_UNSPECIFIED"`},
		{permission: "pERmiSSion_uNSpEcifiEd", expErr: `invalid permission: "pERmiSSion_uNSpEcifiEd"`},

		// Invalid
		{permission: "ettle", expErr: `invalid permission: "ettle"`},
		{permission: "settl", expErr: `invalid permission: "settl"`},
		{permission: "set tle", expErr: `invalid permission: "set tle"`},

		{permission: "ancel", expErr: `invalid permission: "ancel"`},
		{permission: "cance", expErr: `invalid permission: "cance"`},
		{permission: "can cel", expErr: `invalid permission: "can cel"`},

		{permission: "ithdraw", expErr: `invalid permission: "ithdraw"`},
		{permission: "withdra", expErr: `invalid permission: "withdra"`},
		{permission: "with draw", expErr: `invalid permission: "with draw"`},

		{permission: "pdate", expErr: `invalid permission: "pdate"`},
		{permission: "updat", expErr: `invalid permission: "updat"`},
		{permission: "upd ate", expErr: `invalid permission: "upd ate"`},

		{permission: "ermissions", expErr: `invalid permission: "ermissions"`},
		{permission: "permission", expErr: `invalid permission: "permission"`},
		{permission: "permission_permission", expErr: `invalid permission: "permission_permission"`},
		{permission: "permis sions", expErr: `invalid permission: "permis sions"`},

		{permission: "ttributes", expErr: `invalid permission: "ttributes"`},
		{permission: "attribute", expErr: `invalid permission: "attribute"`},
		{permission: "attr ibutes", expErr: `invalid permission: "attr ibutes"`},

		{permission: "", expErr: `invalid permission: ""`},
	}

	for _, tc := range tests {
		name := tc.permission
		if len(tc.permission) == 0 {
			name = "empty"
		}
		t.Run(name, func(t *testing.T) {
			perm, err := ParsePermission(tc.permission)

			if len(tc.expErr) > 0 {
				assert.EqualError(t, err, tc.expErr, "ParsePermission(%q) error", tc.permission)
			} else {
				assert.NoError(t, err, "ParsePermission(%q) error", tc.permission)
			}

			assert.Equal(t, tc.expected, perm, "ParsePermission(%q) result", tc.permission)
		})
	}

	t.Run("all values have a test case", func(t *testing.T) {
		allVals := maps.Keys(Permission_name)
		sort.Slice(allVals, func(i, j int) bool {
			return allVals[i] < allVals[j]
		})

		for _, val := range allVals {
			perm := Permission(val)
			hasTest := false
			for _, tc := range tests {
				if tc.expected == perm {
					hasTest = true
					break
				}
			}
			assert.True(t, hasTest, "No test case found that expects the %s permission", perm)
		}
	})
}

func TestParsePermissions(t *testing.T) {
	tests := []struct {
		name        string
		permissions []string
		expected    []Permission
		expErr      string
	}{
		{
			name:        "nil permissions",
			permissions: nil,
			expected:    nil,
			expErr:      "",
		},
		{
			name:        "empty permissions",
			permissions: nil,
			expected:    nil,
			expErr:      "",
		},
		{
			name:        "one of each permission",
			permissions: []string{"settle", "cancel", "PERMISSION_WITHDRAW", "permission_update", "permissions", "attributes"},
			expected: []Permission{
				Permission_settle,
				Permission_cancel,
				Permission_withdraw,
				Permission_update,
				Permission_permissions,
				Permission_attributes,
			},
		},
		{
			name:        "one bad entry",
			permissions: []string{"settle", "what", "cancel"},
			expected: []Permission{
				Permission_settle,
				Permission_unspecified,
				Permission_cancel,
			},
			expErr: `invalid permission: "what"`,
		},
		{
			name:        "two bad entries",
			permissions: []string{"nope", "withdraw", "notgood"},
			expected: []Permission{
				Permission_unspecified,
				Permission_withdraw,
				Permission_unspecified,
			},
			expErr: `invalid permission: "nope"` + "\n" +
				`invalid permission: "notgood"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			perms, err := ParsePermissions(tc.permissions...)

			// TODO: Refactor to use testutils.AssertErrorValue(t, err, tc.expErr, "ParsePermissions(%q) error", tc.permissions)
			if len(tc.expErr) > 0 {
				assert.EqualError(t, err, tc.expErr, "ParsePermissions(%q) error", tc.permissions)
			} else {
				assert.NoError(t, err, "ParsePermissions(%q) error", tc.permissions)
			}
			assert.Equal(t, tc.expected, perms, "ParsePermissions(%q) result", tc.permissions)
		})
	}
}

func TestValidateReqAttrs(t *testing.T) {
	joinErrs := func(errs ...string) string {
		return strings.Join(errs, "\n")
	}

	tests := []struct {
		name      string
		attrLists [][]string
		exp       string
	}{
		{
			name:      "nil lists",
			attrLists: nil,
			exp:       "",
		},
		{
			name:      "no lists",
			attrLists: [][]string{},
			exp:       "",
		},
		{
			name:      "two empty lists",
			attrLists: [][]string{{}, {}},
			exp:       "",
		},
		{
			name: "one list: three valid entries: normalized",
			attrLists: [][]string{
				{"*.wildcard", "penny.nickel.dime", "*.example.pb"},
			},
			exp: "",
		},
		{
			name: "one list: three valid entries: not normalized",
			attrLists: [][]string{
				{" * . wildcard ", " penny  . nickel .   dime ", " * . example . pb        "},
			},
			exp: "",
		},
		{
			name: "one list: three entries: first invalid",
			attrLists: [][]string{
				{"x.*.wildcard", "penny.nickel.dime", "*.example.pb"},
			},
			exp: `invalid required attribute "x.*.wildcard"`,
		},
		{
			name: "one list: three entries: second invalid",
			attrLists: [][]string{
				{"*.wildcard", "penny.nic kel.dime", "*.example.pb"},
			},
			exp: `invalid required attribute "penny.nic kel.dime"`,
		},
		{
			name: "one list: three entries: third invalid",
			attrLists: [][]string{
				{"*.wildcard", "penny.nickel.dime", "*.ex-am-ple.pb"},
			},
			exp: `invalid required attribute "*.ex-am-ple.pb"`,
		},
		{
			name: "one list: duplicate entries",
			attrLists: [][]string{
				{"*.multi", "*.multi", "*.multi"},
			},
			exp: `duplicate required attribute entry: "*.multi"`,
		},
		{
			name: "one list: duplicate bad entries",
			attrLists: [][]string{
				{"bad.*.example", "bad. * .example"},
			},
			exp: `invalid required attribute "bad.*.example"`,
		},
		{
			name: "one list: multiple problems",
			attrLists: [][]string{
				{
					"one.multi", "x.*.wildcard", "x.*.wildcard", "one.multi", "two.multi",
					"penny.nic kel.dime", "one.multi", "two.multi", "*.ex-am-ple.pb", "two.multi",
				},
			},
			exp: joinErrs(
				`invalid required attribute "x.*.wildcard"`,
				`duplicate required attribute entry: "one.multi"`,
				`invalid required attribute "penny.nic kel.dime"`,
				`duplicate required attribute entry: "two.multi"`,
				`invalid required attribute "*.ex-am-ple.pb"`,
			),
		},
		{
			name: "two lists: second has invalid first",
			attrLists: [][]string{
				{"*.ok", "also.okay.by.me", "this.makes.me.happy"},
				{"x.*.wildcard", "penny.nickel.dime", "*.example.pb"},
			},
			exp: `invalid required attribute "x.*.wildcard"`,
		},
		{
			name: "two lists: second has invalid middle",
			attrLists: [][]string{
				{"*.ok", "also.okay.by.me", "this.makes.me.happy"},
				{"*.wildcard", "penny.nic kel.dime", "*.example.pb"},
			},
			exp: `invalid required attribute "penny.nic kel.dime"`,
		},
		{
			name: "two lists: second has invalid last",
			attrLists: [][]string{
				{"*.ok", "also.okay.by.me", "this.makes.me.happy"},
				{"*.wildcard", "penny.nickel.dime", "*.ex-am-ple.pb"},
			},
			exp: `invalid required attribute "*.ex-am-ple.pb"`,
		},
		{
			name: "two lists: same entry in both but one is not normalized",
			attrLists: [][]string{
				{"this.attr.is.twice"},
				{" This .    Attr . Is . TWice"},
			},
			exp: `duplicate required attribute entry: " This .    Attr . Is . TWice"`,
		},
		{
			name: "two lists: multiple problems",
			attrLists: [][]string{
				{"one.multi", "x.*.wildcard", "x.*.wildcard", "one.multi", "two.multi"},
				{"penny.nic kel.dime", "one.multi", "two.multi", "*.ex-am-ple.pb", "two.multi"},
			},
			exp: joinErrs(
				`invalid required attribute "x.*.wildcard"`,
				`duplicate required attribute entry: "one.multi"`,
				`invalid required attribute "penny.nic kel.dime"`,
				`duplicate required attribute entry: "two.multi"`,
				`invalid required attribute "*.ex-am-ple.pb"`,
			),
		},
		{
			name: "many lists: multiple problems",
			attrLists: [][]string{
				{" one . multi "}, {"x.*.wildcard"}, {"x.*.wildcard"}, {"one.multi"}, {"   two.multi       "},
				{"penny.nic kel.dime"}, {"one.multi"}, {"two.multi"}, {"*.ex-am-ple.pb"}, {"two.multi"},
			},
			exp: joinErrs(
				`invalid required attribute "x.*.wildcard"`,
				`duplicate required attribute entry: "one.multi"`,
				`invalid required attribute "penny.nic kel.dime"`,
				`duplicate required attribute entry: "two.multi"`,
				`invalid required attribute "*.ex-am-ple.pb"`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateReqAttrs(tc.attrLists...)
			// TODO[1658]: Replace this with testutils.AssertErrorValue(t, err, tc.exp, "ValidateReqAttrs")
			if len(tc.exp) > 0 {
				assert.EqualError(t, err, tc.exp, "ValidateReqAttrs")
			} else {
				assert.NoError(t, err, "ValidateReqAttrs")
			}
		})
	}
}

func TestIsValidReqAttr(t *testing.T) {
	tests := []struct {
		name    string
		reqAttr string
		exp     bool
	}{
		{name: "already valid and normalized", reqAttr: "x.y.z", exp: true},
		{name: "already valid but not normalized", reqAttr: " x . y . z ", exp: false},
		{name: "invalid character", reqAttr: "x._y.z", exp: false},
		{name: "just the wildcard", reqAttr: "*", exp: true},
		{name: "just the wildcard not normalized", reqAttr: " * ", exp: false},
		{name: "just star dot", reqAttr: "*.", exp: false},
		{name: "star dot valid", reqAttr: "*.x.y.z", exp: true},
		{name: "star dot valid not normalized", reqAttr: "* . x . y . z", exp: false},
		{name: "star dot invalid", reqAttr: "*.x._y.z", exp: false},
		{name: "empty string", reqAttr: "", exp: false},
		{name: "wildcard in middle", reqAttr: "x.*.y.z", exp: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ok := IsValidReqAttr(tc.reqAttr)
			assert.Equal(t, tc.exp, ok, "IsValidReqAttr(%q)", tc.reqAttr)
		})
	}
}

func TestFindUnmatchedReqAttrs(t *testing.T) {
	tests := []struct {
		name     string
		reqAttrs []string
		accAttrs []string
		exp      []string
	}{
		{
			name:     "nil req attrs",
			reqAttrs: nil,
			accAttrs: []string{"one"},
			exp:      nil,
		},
		{
			name:     "empty req attrs",
			reqAttrs: []string{},
			accAttrs: []string{"one"},
			exp:      nil,
		},
		{
			name:     "one req attr no wildcard: in acc attrs",
			reqAttrs: []string{"one"},
			accAttrs: []string{"one", "two"},
			exp:      nil,
		},
		{
			name:     "one req attr with wildcard: in acc attrs",
			reqAttrs: []string{"*.one"},
			accAttrs: []string{"zero.one", "two"},
			exp:      nil,
		},
		{
			name:     "three req attrs: nil acc attrs",
			reqAttrs: []string{"*.desk", "nickel.dime.quarter", "*.x.y.z"},
			accAttrs: nil,
			exp:      []string{"*.desk", "nickel.dime.quarter", "*.x.y.z"},
		},
		{
			name:     "three req attrs: empty acc attrs",
			reqAttrs: []string{"*.desk", "nickel.dime.quarter", "*.x.y.z"},
			accAttrs: []string{},
			exp:      []string{"*.desk", "nickel.dime.quarter", "*.x.y.z"},
		},
		{
			name:     "three req attrs: only first in acc attrs",
			reqAttrs: []string{"*.desk", "nickel.dime.quarter", "*.x.y.z"},
			accAttrs: []string{"lamp.corner.desk"},
			exp:      []string{"nickel.dime.quarter", "*.x.y.z"},
		},
		{
			name:     "three req attrs: only second in acc attrs",
			reqAttrs: []string{"*.desk", "nickel.dime.quarter", "*.x.y.z"},
			accAttrs: []string{"nickel.dime.quarter"},
			exp:      []string{"*.desk", "*.x.y.z"},
		},
		{
			name:     "three req attrs: only third in acc attrs",
			reqAttrs: []string{"*.desk", "nickel.dime.quarter", "*.x.y.z"},
			accAttrs: []string{"w.x.y.z"},
			exp:      []string{"*.desk", "nickel.dime.quarter"},
		},
		{
			name:     "three req attrs: missing first",
			reqAttrs: []string{"*.desk", "nickel.dime.quarter", "*.x.y.z"},
			accAttrs: []string{"nickel.dime.quarter", "w.x.y.z"},
			exp:      []string{"*.desk"},
		},
		{
			name:     "three req attrs: missing middle",
			reqAttrs: []string{"*.desk", "nickel.dime.quarter", "*.x.y.z"},
			accAttrs: []string{"lamp.corner.desk", "w.x.y.z"},
			exp:      []string{"nickel.dime.quarter"},
		},
		{
			name:     "three req attrs: missing last",
			reqAttrs: []string{"*.desk", "nickel.dime.quarter", "*.x.y.z"},
			accAttrs: []string{"lamp.corner.desk", "nickel.dime.quarter"},
			exp:      []string{"*.x.y.z"},
		},
		{
			name:     "three req attrs: has all",
			reqAttrs: []string{"*.desk", "nickel.dime.quarter", "*.x.y.z"},
			accAttrs: []string{"just.some.first", "w.x.y.z", "other", "lamp.corner.desk", "random.entry", "nickel.dime.quarter", "what.is.this"},
			exp:      nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			unmatched := FindUnmatchedReqAttrs(tc.reqAttrs, tc.accAttrs)
			assert.Equal(t, tc.exp, unmatched, "FindUnmatchedReqAttrs(%q, %q", tc.reqAttrs, tc.accAttrs)
		})
	}
}

func TestHasReqAttrMatch(t *testing.T) {
	tests := []struct {
		name     string
		reqAttr  string
		accAttrs []string
		exp      bool
	}{
		{
			name:     "nil acc attrs",
			reqAttr:  "nickel.dime.quarter",
			accAttrs: nil,
			exp:      false,
		},
		{
			name:     "empty acc attrs",
			reqAttr:  "nickel.dime.quarter",
			accAttrs: []string{},
			exp:      false,
		},
		{
			name:    "no wildcard: not in acc attrs",
			reqAttr: "nickel.dime.quarter",
			accAttrs: []string{
				"xnickel.dime.quarter",
				"nickelx.dime.quarter",
				"nickel.xdime.quarter",
				"nickel.dimex.quarter",
				"nickel.dime.xquarter",
				"nickel.dime.quarterx",
				"penny.nickel.dime.quarter",
				"nickel.dime.quarter.dollar",
			},
			exp: false,
		},
		{
			name:     "no wildcard: only one in acc attrs",
			reqAttr:  "nickel.dime.quarter",
			accAttrs: []string{"nickel.dime.quarter"},
			exp:      true,
		},
		{
			name:    "no wildcard: first in acc attrs",
			reqAttr: "nickel.dime.quarter",
			accAttrs: []string{
				"nickel.dime.quarter",
				"xnickel.dime.quarter",
				"nickelx.dime.quarter",
				"nickel.xdime.quarter",
				"nickel.dimex.quarter",
				"nickel.dime.xquarter",
				"nickel.dime.quarterx",
				"penny.nickel.dime.quarter",
				"nickel.dime.quarter.dollar",
			},
			exp: true,
		},
		{
			name:    "no wildcard: in middle of acc attrs",
			reqAttr: "nickel.dime.quarter",
			accAttrs: []string{
				"xnickel.dime.quarter",
				"nickelx.dime.quarter",
				"nickel.xdime.quarter",
				"nickel.dimex.quarter",
				"nickel.dime.quarter",
				"nickel.dime.xquarter",
				"nickel.dime.quarterx",
				"penny.nickel.dime.quarter",
				"nickel.dime.quarter.dollar",
			},
			exp: true,
		},
		{
			name:    "no wildcard: at end of acc attrs",
			reqAttr: "nickel.dime.quarter",
			accAttrs: []string{
				"xnickel.dime.quarter",
				"nickelx.dime.quarter",
				"nickel.xdime.quarter",
				"nickel.dimex.quarter",
				"nickel.dime.xquarter",
				"nickel.dime.quarterx",
				"penny.nickel.dime.quarter",
				"nickel.dime.quarter.dollar",
				"nickel.dime.quarter",
			},
			exp: true,
		},

		{
			name:    "with wildcard: no match",
			reqAttr: "*.dime.quarter",
			accAttrs: []string{
				"dime.quarter",
				"penny.xdime.quarter",
				"penny.dimex.quarter",
				"penny.dime.xquarter",
				"penny.dime.quarterx",
				"penny.quarter",
				"penny.dime",
			},
			exp: false,
		},
		{
			name:    "with wildcard: matches only entry",
			reqAttr: "*.dime.quarter",
			accAttrs: []string{
				"penny.dime.quarter",
			},
			exp: true,
		},
		{
			name:    "with wildcard: matches first entry",
			reqAttr: "*.dime.quarter",
			accAttrs: []string{
				"penny.dime.quarter",
				"dime.quarter",
				"penny.xdime.quarter",
				"penny.dimex.quarter",
				"penny.dime.xquarter",
				"penny.dime.quarterx",
				"penny.quarter",
				"penny.dime",
			},
			exp: true,
		},
		{
			name:    "with wildcard: matches middle entry",
			reqAttr: "*.dime.quarter",
			accAttrs: []string{
				"dime.quarter",
				"penny.xdime.quarter",
				"penny.dimex.quarter",
				"penny.dime.xquarter",
				"penny.dime.quarter",
				"penny.dime.quarterx",
				"penny.quarter",
				"penny.dime",
			},
			exp: true,
		},
		{
			name:    "with wildcard: matches last entry",
			reqAttr: "*.dime.quarter",
			accAttrs: []string{
				"dime.quarter",
				"penny.xdime.quarter",
				"penny.dimex.quarter",
				"penny.dime.xquarter",
				"penny.dime.quarterx",
				"penny.quarter",
				"penny.dime",
				"penny.dime.quarter",
			},
			exp: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hasMatch := HasReqAttrMatch(tc.reqAttr, tc.accAttrs)
			assert.Equal(t, tc.exp, hasMatch, "HasReqAttrMatch(%q, %q)", tc.reqAttr, tc.accAttrs)
		})
	}
}

func TestIsReqAttrMatch(t *testing.T) {
	tests := []struct {
		name    string
		reqAttr string
		accAttr string
		exp     bool
	}{
		{
			name:    "empty req attr",
			reqAttr: "",
			accAttr: "foo",
			exp:     false,
		},
		{
			name:    "empty acc attr",
			reqAttr: "foo",
			accAttr: "",
			exp:     false,
		},
		{
			name:    "both empty",
			reqAttr: "",
			accAttr: "",
			exp:     false,
		},
		{
			name:    "no wildcard: exact match",
			reqAttr: "penny.dime.quarter",
			accAttr: "penny.dime.quarter",
			exp:     true,
		},
		{
			name:    "no wildcard: opposite order",
			reqAttr: "penny.dime.quarter",
			accAttr: "quarter.dime.penny",
			exp:     false,
		},
		{
			name:    "no wildcard: missing 1st char from 1st name",
			reqAttr: "penny.dime.quarter",
			accAttr: "enny.dime.quarter",
			exp:     false,
		},
		{
			name:    "no wildcard: missing last char from 1st name",
			reqAttr: "penny.dime.quarter",
			accAttr: "penn.dime.quarter",
			exp:     false,
		},
		{
			name:    "no wildcard: missing 1st char from middle name",
			reqAttr: "penny.dime.quarter",
			accAttr: "penny.ime.quarter",
			exp:     false,
		},
		{
			name:    "no wildcard: missing last char from middle name",
			reqAttr: "penny.dime.quarter",
			accAttr: "penny.dim.quarter",
			exp:     false,
		},
		{
			name:    "no wildcard: missing 1st char from last name",
			reqAttr: "penny.dime.quarter",
			accAttr: "penny.dime.uarter",
			exp:     false,
		},
		{
			name:    "no wildcard: missing last char from last name",
			reqAttr: "penny.dime.quarter",
			accAttr: "penny.dime.quarte",
			exp:     false,
		},
		{
			name:    "no wildcard: extra char at start of first name",
			reqAttr: "penny.dime.quarter",
			accAttr: "xpenny.dime.quarter",
			exp:     false,
		},
		{
			name:    "no wildcard: extra char at end of first name",
			reqAttr: "penny.dime.quarter",
			accAttr: "pennyx.dime.quarter",
			exp:     false,
		},
		{
			name:    "no wildcard: extra char at start of middle name",
			reqAttr: "penny.dime.quarter",
			accAttr: "penny.xdime.quarter",
			exp:     false,
		},
		{
			name:    "no wildcard: extra char at end of middle name",
			reqAttr: "penny.dime.quarter",
			accAttr: "penny.dimex.quarter",
			exp:     false,
		},
		{
			name:    "no wildcard: extra char at start of last name",
			reqAttr: "penny.dime.quarter",
			accAttr: "penny.dime.xquarter",
			exp:     false,
		},
		{
			name:    "no wildcard: extra char at end of first name",
			reqAttr: "penny.dime.quarter",
			accAttr: "penny.dime.quarterx",
			exp:     false,
		},
		{
			name:    "no wildcard: extra name at start",
			reqAttr: "penny.dime.quarter",
			accAttr: "mil.penny.dime.quarter",
			exp:     false,
		},
		{
			name:    "no wildcard: extra name at end",
			reqAttr: "penny.dime.quarter",
			accAttr: "penny.dime.quarter.dollar",
			exp:     false,
		},
		{
			name:    "with wildcard: missing 1st char from 1st name",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "enny.dime.quarter",
			exp:     false,
		},
		{
			name:    "with wildcard: missing last char from 1st name",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "penn.dime.quarter",
			exp:     false,
		},
		{
			name:    "with wildcard: missing 1st char from middle name",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "penny.ime.quarter",
			exp:     false,
		},
		{
			name:    "with wildcard: missing last char from middle name",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "penny.dim.quarter",
			exp:     false,
		},
		{
			name:    "with wildcard: missing 1st char from last name",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "penny.dime.uarter",
			exp:     false,
		},
		{
			name:    "with wildcard: missing last char from last name",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "penny.dime.quarte",
			exp:     false,
		},
		{
			name:    "with wildcard: extra char at start of first name",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "xpenny.dime.quarter",
			exp:     false,
		},
		{
			name:    "with wildcard: extra char at end of first name",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "pennyx.dime.quarter",
			exp:     false,
		},
		{
			name:    "with wildcard: extra char at start of middle name",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "penny.xdime.quarter",
			exp:     false,
		},
		{
			name:    "with wildcard: extra char at end of middle name",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "penny.dimex.quarter",
			exp:     false,
		},
		{
			name:    "with wildcard: extra char at start of last name",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "penny.dime.xquarter",
			exp:     false,
		},
		{
			name:    "with wildcard: extra char at end of first name",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "penny.dime.quarterx",
			exp:     false,
		},
		{
			name:    "with wildcard: extra name at start",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "mil.penny.dime.quarter",
			exp:     true,
		},
		{
			name:    "with wildcard: two extra names at start",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "scraps.mil.penny.dime.quarter",
			exp:     true,
		},
		{
			name:    "with wildcard: extra name at start but wrong 1st req name",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "mil.xpenny.dime.quarter",
			exp:     false,
		},
		{
			name:    "with wildcard: only base name",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "penny.dime.quarter",
			exp:     false,
		},
		{
			name:    "with wildcard: extra name at end",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "penny.dime.quarter.dollar",
			exp:     false,
		},
		{
			name:    "with wildcard: extra name at start but wrong base order",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "dollar.quarter.dime.penny",
			exp:     false,
		},
		{
			name:    "star in middle",
			reqAttr: "penny.*.quarter",
			accAttr: "penny.dime.quarter",
			exp:     false,
		},
		{
			name:    "just a star: empty account attribute",
			reqAttr: "*",
			accAttr: "",
			exp:     false,
		},
		{
			name:    "just wildcard: empty account attribute",
			reqAttr: "*.",
			accAttr: "",
			exp:     false,
		},
		{
			name:    "just a star: account attribute has value",
			reqAttr: "*",
			accAttr: "penny.dime.quarter",
			exp:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isMatch := IsReqAttrMatch(tc.reqAttr, tc.accAttr)
			assert.Equal(t, tc.exp, isMatch, "IsReqAttrMatch(%q, %q)", tc.reqAttr, tc.accAttr)
		})
	}
}
