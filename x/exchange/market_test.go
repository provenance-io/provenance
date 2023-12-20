package exchange

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/internal/helpers"
	"github.com/provenance-io/provenance/testutil/assertions"
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
				FeeSellerSettlementFlat: coins("50nnibler,8mfry"),
				FeeSellerSettlementRatios: []FeeRatio{
					{Price: coin(1000, "nnibler"), Fee: coin(1, "nnibler")},
					{Price: coin(300, "mfry"), Fee: coin(1, "mfry")},
				},
				FeeBuyerSettlementFlat: coins("100nnibler,20mfry"),
				FeeBuyerSettlementRatios: []FeeRatio{
					{Price: coin(500, "nnibler"), Fee: coin(1, "nnibler")},
					{Price: coin(500, "nnibler"), Fee: coin(8, "mfry")},
					{Price: coin(150, "mfry"), Fee: coin(1, "mfry")},
					{Price: coin(1, "mfry"), Fee: coin(1, "nnibler")},
				},
				AcceptingOrders:     true,
				AllowUserSettlement: true,
				AccessGrants: []AccessGrant{
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
			name:   "invalid fee create-ask flat",
			market: Market{FeeCreateAskFlat: sdk.Coins{coin(-1, "leela")}},
			expErr: []string{`invalid create-ask flat fee option "-1leela": negative coin amount: -1`},
		},
		{
			name:   "invalid fee create-bid flat",
			market: Market{FeeCreateBidFlat: sdk.Coins{coin(-1, "leela")}},
			expErr: []string{`invalid create-bid flat fee option "-1leela": negative coin amount: -1`},
		},
		{
			name:   "invalid fee seller settlement flat",
			market: Market{FeeSellerSettlementFlat: sdk.Coins{coin(-1, "leela")}},
			expErr: []string{`invalid seller settlement flat fee option "-1leela": negative coin amount: -1`},
		},
		{
			name:   "invalid fee buyer settlement flat",
			market: Market{FeeBuyerSettlementFlat: sdk.Coins{coin(-1, "leela")}},
			expErr: []string{`invalid buyer settlement flat fee option "-1leela": negative coin amount: -1`},
		},
		{
			name:   "invalid seller ratio",
			market: Market{FeeSellerSettlementRatios: []FeeRatio{{Price: coin(0, "fry"), Fee: coin(1, "fry")}}},
			expErr: []string{`seller fee ratio price amount "0fry" must be positive`},
		},
		{
			name:   "invalid buyer ratio",
			market: Market{FeeBuyerSettlementRatios: []FeeRatio{{Price: coin(0, "fry"), Fee: coin(1, "fry")}}},
			expErr: []string{`buyer fee ratio price amount "0fry" must be positive`},
		},
		{
			name: "invalid ratios",
			market: Market{
				FeeSellerSettlementRatios: []FeeRatio{{Price: coin(10, "fry"), Fee: coin(1, "fry")}},
				FeeBuyerSettlementRatios:  []FeeRatio{{Price: coin(100, "leela"), Fee: coin(1, "leela")}},
			},
			expErr: []string{
				`denom "fry" is defined in the seller settlement fee ratios but not buyer`,
				`denom "leela" is defined in the buyer settlement fee ratios but not seller`,
			},
		},
		{
			name:   "invalid access grants",
			market: Market{AccessGrants: []AccessGrant{{Address: "bad_addr", Permissions: AllPermissions()}}},
			expErr: []string{`invalid access grant: invalid address "bad_addr": decoding bech32 failed: invalid separator index -1`},
		},
		{
			name:   "invalid ask required attributes",
			market: Market{ReqAttrCreateAsk: []string{"this-attr-is-bad"}},
			expErr: []string{`invalid create-ask required attribute "this-attr-is-bad"`},
		},
		{
			name:   "invalid bid required attributes",
			market: Market{ReqAttrCreateBid: []string{"this-attr-grrrr"}},
			expErr: []string{`invalid create-bid required attribute "this-attr-grrrr"`},
		},
		{
			name: "multiple errors",
			market: Market{
				MarketDetails:             MarketDetails{Name: strings.Repeat("n", MaxName+1)},
				FeeCreateAskFlat:          sdk.Coins{coin(-1, "leela")},
				FeeCreateBidFlat:          sdk.Coins{coin(-1, "leela")},
				FeeSellerSettlementFlat:   sdk.Coins{coin(-1, "leela")},
				FeeBuyerSettlementFlat:    sdk.Coins{coin(-1, "leela")},
				FeeSellerSettlementRatios: []FeeRatio{{Price: coin(10, "fry"), Fee: coin(1, "fry")}},
				FeeBuyerSettlementRatios:  []FeeRatio{{Price: coin(100, "leela"), Fee: coin(1, "leela")}},
				AccessGrants:              []AccessGrant{{Address: "bad_addr", Permissions: AllPermissions()}},
				ReqAttrCreateAsk:          []string{"this-attr-is-bad"},
				ReqAttrCreateBid:          []string{"this-attr-grrrr"},
			},
			expErr: []string{
				fmt.Sprintf("name length %d exceeds maximum length of %d", MaxName+1, MaxName),
				`invalid create-ask flat fee option "-1leela": negative coin amount: -1`,
				`invalid create-bid flat fee option "-1leela": negative coin amount: -1`,
				`invalid seller settlement flat fee option "-1leela": negative coin amount: -1`,
				`invalid buyer settlement flat fee option "-1leela": negative coin amount: -1`,
				`denom "fry" is defined in the seller settlement fee ratios but not buyer`,
				`denom "leela" is defined in the buyer settlement fee ratios but not seller`,
				`invalid access grant: invalid address "bad_addr": decoding bech32 failed: invalid separator index -1`,
				`invalid create-ask required attribute "this-attr-is-bad"`,
				`invalid create-bid required attribute "this-attr-grrrr"`,
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

			assertions.AssertErrorContents(t, err, tc.expErr, "Market.Validate")
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
			field:   "pinky",
			options: []sdk.Coin{coin(5, "fry"), coin(2, "leela"), coin(0, "farnsworth")},
			expErr:  `invalid pinky option "0farnsworth": amount cannot be zero`,
		},
		{
			name:  "duplicated denoms",
			field: "duppa-duppa-dup",
			options: []sdk.Coin{
				coin(1, "fry"), coin(1, "leela"),
				coin(1, "amy"), coin(2, "amy"),
				coin(1, "farnsworth"),
				coin(2, "fry"), coin(3, "fry")},
			expErr: joinErrs(
				`invalid duppa-duppa-dup option "2amy": denom used in multiple entries`,
				`invalid duppa-duppa-dup option "2fry": denom used in multiple entries`,
			),
		},
		{
			name:    "three options: all bad",
			field:   "carp corp",
			options: []sdk.Coin{coin(0, "fry"), coin(1, "l"), coin(-12, "farnsworth")},
			expErr: joinErrs(
				`invalid carp corp option "0fry": amount cannot be zero`,
				`invalid carp corp option "1l": invalid denom: l`,
				`invalid carp corp option "-12farnsworth": negative coin amount: -12`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = ValidateFeeOptions(tc.field, tc.options)
			}
			require.NotPanics(t, testFunc, "ValidateFeeOptions")

			assertions.AssertErrorValue(t, err, tc.expErr, "ValidateFeeOptions")
		})
	}
}

func TestMarketDetails_Validate(t *testing.T) {
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
			expErr: nameErr(2) + "\n" + descErr(3) + "\n" + urlErr(4) + "\n" + iconErr(5),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.details.Validate()
			}
			require.NotPanics(t, testFunc, "Validate")

			assertions.AssertErrorValue(t, err, tc.expErr, "Validate")
		})
	}
}

func TestValidateFeeRatios(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}

	tests := []struct {
		name         string
		sellerRatios []FeeRatio
		buyerRatios  []FeeRatio
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
			buyerRatios:  []FeeRatio{},
			exp:          "",
		},
		{
			name:         "empty nil",
			sellerRatios: []FeeRatio{},
			buyerRatios:  nil,
			exp:          "",
		},
		{
			name:         "empty empty",
			sellerRatios: []FeeRatio{},
			buyerRatios:  []FeeRatio{},
			exp:          "",
		},
		{
			name: "multiple errors from sellers and buyers",
			sellerRatios: []FeeRatio{
				{Price: coin(1, "fry"), Fee: coin(5, "fry")},
				{Price: coin(0, "leela"), Fee: coin(1, "leela")},
			},
			buyerRatios: []FeeRatio{
				{Price: coin(1, "fry"), Fee: coin(5, "fry")},
				{Price: coin(0, "leela"), Fee: coin(1, "leela")},
			},
			exp: joinErrs(
				`seller fee ratio fee amount "5fry" cannot be greater than price amount "1fry"`,
				`seller fee ratio price amount "0leela" must be positive`,
				`buyer fee ratio fee amount "5fry" cannot be greater than price amount "1fry"`,
				`buyer fee ratio price amount "0leela" must be positive`,
			),
		},
		{
			name: "sellers have price denom that buyers do not",
			sellerRatios: []FeeRatio{
				{Price: coin(100, "leela"), Fee: coin(1, "leela")},
				{Price: coin(500, "fry"), Fee: coin(3, "fry")},
			},
			buyerRatios: []FeeRatio{
				{Price: coin(100, "leela"), Fee: coin(3, "leela")},
			},
			exp: `denom "fry" is defined in the seller settlement fee ratios but not buyer`,
		},
		{
			name: "sellers have two price denoms that buyers do not",
			sellerRatios: []FeeRatio{
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
			sellerRatios: []FeeRatio{
				{Price: coin(100, "leela"), Fee: coin(3, "leela")},
			},
			buyerRatios: []FeeRatio{
				{Price: coin(100, "leela"), Fee: coin(1, "leela")},
				{Price: coin(500, "fry"), Fee: coin(3, "leela")},
			},
			exp: `denom "fry" is defined in the buyer settlement fee ratios but not seller`,
		},
		{
			name:         "buyers have two price denoms that sellers do not",
			sellerRatios: nil,
			buyerRatios: []FeeRatio{
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
			sellerRatios: []FeeRatio{
				{Price: coin(100, "leela"), Fee: coin(1, "leela")},
				{Price: coin(500, "fry"), Fee: coin(3, "fry")},
			},
			buyerRatios: []FeeRatio{
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
			sellerRatios: []FeeRatio{
				{Price: coin(100, "leela"), Fee: coin(1, "leela")},
				{Price: coin(500, "fry"), Fee: coin(3, "fry")},
				{Price: coin(300, "bender"), Fee: coin(7, "bender")},
			},
			buyerRatios: []FeeRatio{
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

			assertions.AssertErrorValue(t, err, tc.exp, "ValidateFeeRatios")
		})
	}
}

func TestValidateSellerFeeRatios(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}

	tests := []struct {
		name   string
		ratios []FeeRatio
		exp    string
	}{
		{
			name:   "nil ratios",
			ratios: nil,
			exp:    "",
		},
		{
			name:   "empty ratios",
			ratios: []FeeRatio{},
			exp:    "",
		},
		{
			name:   "one ratio: different denoms",
			ratios: []FeeRatio{{Price: coin(3, "hermes"), Fee: coin(2, "mom")}},
			exp:    `seller fee ratio price denom "hermes" does not equal fee denom "mom"`,
		},
		{
			name:   "one ratio: same denoms",
			ratios: []FeeRatio{{Price: coin(3, "mom"), Fee: coin(2, "mom")}},
			exp:    "",
		},
		{
			name:   "one ratio: invalid",
			ratios: []FeeRatio{{Price: coin(0, "hermes"), Fee: coin(2, "hermes")}},
			exp:    `seller fee ratio price amount "0hermes" must be positive`,
		},
		{
			name: "two with same denom",
			ratios: []FeeRatio{
				{Price: coin(3, "hermes"), Fee: coin(2, "hermes")},
				{Price: coin(6, "hermes"), Fee: coin(4, "hermes")},
			},
			exp: `seller fee ratio denom "hermes" appears in multiple ratios`,
		},
		{
			name: "three with different denoms",
			ratios: []FeeRatio{
				{Price: coin(30, "leela"), Fee: coin(1, "leela")},
				{Price: coin(5, "fry"), Fee: coin(1, "fry")},
				{Price: coin(100, "professor"), Fee: coin(1, "professor")},
			},
			exp: "",
		},
		{
			name: "multiple errors",
			ratios: []FeeRatio{
				{Price: coin(3, "mom"), Fee: coin(2, "hermes")},
				{Price: coin(0, "hermes"), Fee: coin(2, "hermes")},
				{Price: coin(6, "bender"), Fee: coin(4, "bender")},
				{Price: coin(1, "hermes"), Fee: coin(2, "hermes")},
				{Price: coin(2, "bender"), Fee: coin(1, "bender")},
				// This one is ignored because we've already complained about multiple hermes.
				{Price: coin(30, "hermes"), Fee: coin(2, "hermes")},
			},
			exp: `seller fee ratio price denom "mom" does not equal fee denom "hermes"` + "\n" +
				`seller fee ratio price amount "0hermes" must be positive` + "\n" +
				`seller fee ratio denom "hermes" appears in multiple ratios` + "\n" +
				`seller fee ratio denom "bender" appears in multiple ratios`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = ValidateSellerFeeRatios(tc.ratios)
			}
			require.NotPanics(t, testFunc, "ValidateSellerFeeRatios")

			assertions.AssertErrorValue(t, err, tc.exp, "ValidateSellerFeeRatios")
		})
	}
}

func TestValidateBuyerFeeRatios(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}

	tests := []struct {
		name   string
		ratios []FeeRatio
		exp    string
	}{
		{
			name:   "nil ratios",
			ratios: nil,
			exp:    "",
		},
		{
			name:   "empty ratios",
			ratios: []FeeRatio{},
			exp:    "",
		},
		{
			name:   "one ratio: different denoms",
			ratios: []FeeRatio{{Price: coin(3, "hermes"), Fee: coin(2, "mom")}},
			exp:    "",
		},
		{
			name:   "one ratio: same denoms",
			ratios: []FeeRatio{{Price: coin(3, "mom"), Fee: coin(2, "mom")}},
			exp:    "",
		},
		{
			name:   "one ratio: invalid",
			ratios: []FeeRatio{{Price: coin(0, "hermes"), Fee: coin(2, "hermes")}},
			exp:    `buyer fee ratio price amount "0hermes" must be positive`,
		},
		{
			name: "duplicate ratio denoms",
			ratios: []FeeRatio{
				{Price: coin(10, "morbo"), Fee: coin(2, "scruffy")},
				{Price: coin(3, "morbo"), Fee: coin(1, "scruffy")},
			},
			exp: `buyer fee ratio pair "morbo" to "scruffy" appears in multiple ratios`,
		},
		{
			name: "two ratios one each way",
			ratios: []FeeRatio{
				{Price: coin(10, "leela"), Fee: coin(2, "scruffy")},
				{Price: coin(2, "scruffy"), Fee: coin(8, "leela")},
			},
			exp: "",
		},
		{
			name: "multiple errors",
			ratios: []FeeRatio{
				{Price: coin(10, "morbo"), Fee: coin(2, "scruffy")},
				{Price: coin(0, "zoidberg"), Fee: coin(1, "amy")},
				{Price: coin(1, "hermes"), Fee: coin(2, "hermes")},
				{Price: coin(3, "morbo"), Fee: coin(1, "scruffy")},
				{Price: coin(0, "zoidberg"), Fee: coin(1, "amy")},
				// This one has a different fee denom, though, so it's checked.
				{Price: coin(1, "zoidberg"), Fee: coin(-1, "fry")},
				// We've already complained about this one, so it doesn't happen again.
				{Price: coin(12, "zoidberg"), Fee: coin(55, "amy")},
			},
			exp: `buyer fee ratio price amount "0zoidberg" must be positive` + "\n" +
				`buyer fee ratio fee amount "2hermes" cannot be greater than price amount "1hermes"` + "\n" +
				`buyer fee ratio pair "morbo" to "scruffy" appears in multiple ratios` + "\n" +
				`buyer fee ratio pair "zoidberg" to "amy" appears in multiple ratios` + "\n" +
				`buyer fee ratio fee amount "-1fry" cannot be negative`,
		},
		{
			name: "two different price denoms to several fee denoms",
			ratios: []FeeRatio{
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

			assertions.AssertErrorValue(t, err, tc.exp, "ValidateBuyerFeeRatios")
		})
	}
}

func TestParseCoin(t *testing.T) {
	tests := []struct {
		name    string
		coinStr string
		expCoin sdk.Coin
		expErr  bool
	}{
		{
			name:    "empty string",
			coinStr: "",
			expErr:  true,
		},
		{
			name:    "no denom",
			coinStr: "12345",
			expErr:  true,
		},
		{
			name:    "no amount",
			coinStr: "nhash",
			expErr:  true,
		},
		{
			name:    "decimal amount",
			coinStr: "123.45banana",
			expErr:  true,
		},
		{
			name:    "zero amount",
			coinStr: "0apple",
			expCoin: sdk.NewInt64Coin("apple", 0),
		},
		{
			name:    "normal",
			coinStr: "500acorn",
			expCoin: sdk.NewInt64Coin("acorn", 500),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var expErr string
			if tc.expErr {
				expErr = fmt.Sprintf("invalid coin expression: %q", tc.coinStr)
			}

			var coin sdk.Coin
			var err error
			testFunc := func() {
				coin, err = ParseCoin(tc.coinStr)
			}
			require.NotPanics(t, testFunc, "ParseCoin(%q)", tc.coinStr)
			assertions.AssertErrorValue(t, err, expErr, "ParseCoin(%q) error", tc.coinStr)
			if !assert.Equal(t, tc.expCoin, coin, "ParseCoin(%q) coin", tc.coinStr) {
				t.Logf("Expected: %s", tc.expCoin)
				t.Logf("  Actual: %s", coin)
			}
		})
	}
}

func TestParseFeeRatio(t *testing.T) {
	ratioStr := func(ratio *FeeRatio) string {
		if ratio == nil {
			return "<nil>"
		}
		return fmt.Sprintf("%q", ratio.String())
	}

	tests := []struct {
		name     string
		ratio    string
		expRatio *FeeRatio
		expErr   string
	}{
		{
			name:   "no colons",
			ratio:  "8banana",
			expErr: "expected exactly one colon",
		},
		{
			name:   "two colons",
			ratio:  "8apple:5banana:3cactus",
			expErr: "expected exactly one colon",
		},
		{
			name:   "one colon: first char",
			ratio:  ":18banana",
			expErr: "price: invalid coin expression: \"\"",
		},
		{
			name:   "one colon: las char",
			ratio:  "33apple:",
			expErr: "fee: invalid coin expression: \"\"",
		},
		{
			name:   "bad price coin",
			ratio:  "1234:5banana",
			expErr: "price: invalid coin expression: \"1234\"",
		},
		{
			name:   "bad fee coin",
			ratio:  "1234apple:banana",
			expErr: "fee: invalid coin expression: \"banana\"",
		},
		{
			name:   "neg price coin",
			ratio:  "-55apple:3banana",
			expErr: "price: invalid coin expression: \"-55apple\"",
		},
		{
			name:     "neg fee coin",
			ratio:    "55apple:-3banana",
			expRatio: nil,
			expErr:   "fee: invalid coin expression: \"-3banana\"",
		},
		{
			name:  "zero price coin",
			ratio: "0apple:21banana",
			expRatio: &FeeRatio{
				Price: sdk.NewInt64Coin("apple", 0),
				Fee:   sdk.NewInt64Coin("banana", 21),
			},
		},
		{
			name:  "zero fee coin",
			ratio: "5apple:0banana",
			expRatio: &FeeRatio{
				Price: sdk.NewInt64Coin("apple", 5),
				Fee:   sdk.NewInt64Coin("banana", 0),
			},
		},
		{
			name:  "same denoms: price more",
			ratio: "30apple:29apple",
			expRatio: &FeeRatio{
				Price: sdk.NewInt64Coin("apple", 30),
				Fee:   sdk.NewInt64Coin("apple", 29),
			},
		},
		{
			name:  "same denoms: price same",
			ratio: "30apple:30apple",
			expRatio: &FeeRatio{
				Price: sdk.NewInt64Coin("apple", 30),
				Fee:   sdk.NewInt64Coin("apple", 30),
			},
		},
		{
			name:  "same denoms: price less",
			ratio: "30apple:31apple",
			expRatio: &FeeRatio{
				Price: sdk.NewInt64Coin("apple", 30),
				Fee:   sdk.NewInt64Coin("apple", 31),
			},
		},
		{
			name:  "diff denoms: price more",
			ratio: "30apple:29banana",
			expRatio: &FeeRatio{
				Price: sdk.NewInt64Coin("apple", 30),
				Fee:   sdk.NewInt64Coin("banana", 29),
			},
		},
		{
			name:  "diff denoms: price same",
			ratio: "30apple:30banana",
			expRatio: &FeeRatio{
				Price: sdk.NewInt64Coin("apple", 30),
				Fee:   sdk.NewInt64Coin("banana", 30),
			},
		},
		{
			name:  "diff denoms: price less",
			ratio: "30apple:31banana",
			expRatio: &FeeRatio{
				Price: sdk.NewInt64Coin("apple", 30),
				Fee:   sdk.NewInt64Coin("banana", 31),
			},
		},
		{
			name:   "price has decimal",
			ratio:  "123.4apple:5banana",
			expErr: "price: invalid coin expression: \"123.4apple\"",
		},
		{
			name:   "fee has decimal",
			ratio:  "123apple:5.6banana",
			expErr: "fee: invalid coin expression: \"5.6banana\"",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.expErr) > 0 {
				tc.expErr = fmt.Sprintf("cannot create FeeRatio from %q: %s", tc.ratio, tc.expErr)
			}

			var ratio *FeeRatio
			var err error
			testFunc := func() {
				ratio, err = ParseFeeRatio(tc.ratio)
			}
			require.NotPanics(t, testFunc, "ParseFeeRatio(%q)", tc.ratio)
			assertions.AssertErrorValue(t, err, tc.expErr, "ParseFeeRatio(%q)", tc.ratio)
			assert.Equal(t, ratioStr(tc.expRatio), ratioStr(ratio), "ParseFeeRatio(%q)", tc.ratio)

			var ratioMust FeeRatio
			testFuncMust := func() {
				ratioMust = MustParseFeeRatio(tc.ratio)
			}
			assertions.RequirePanicEquals(t, testFuncMust, tc.expErr, "MustParseFeeRatio(%q)", tc.ratio)
			if tc.expRatio != nil {
				assert.Equal(t, ratioStr(tc.expRatio), ratioStr(&ratioMust), "MustParseFeeRatio(%q)", tc.ratio)
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

func TestFeeRatiosString(t *testing.T) {
	feeRatio := func(priceAmount int64, priceDenom string, feeAmount int64, feeDenom string) FeeRatio {
		return FeeRatio{
			Price: sdk.Coin{Denom: priceDenom, Amount: sdkmath.NewInt(priceAmount)},
			Fee:   sdk.Coin{Denom: feeDenom, Amount: sdkmath.NewInt(feeAmount)},
		}
	}

	tests := []struct {
		name   string
		ratios []FeeRatio
		exp    string
	}{
		{name: "nil", ratios: nil, exp: ""},
		{name: "empty", ratios: []FeeRatio{}, exp: ""},
		{
			name:   "one entry",
			ratios: []FeeRatio{feeRatio(1, "pdenom", 2, "fdenom")},
			exp:    "1pdenom:2fdenom",
		},
		{
			name: "two entries",
			ratios: []FeeRatio{
				feeRatio(1, "pdenom", 2, "fdenom"),
				feeRatio(3, "qdenom", 4, "gdenom"),
			},
			exp: "1pdenom:2fdenom,3qdenom:4gdenom",
		},
		{
			name: "five entries",
			ratios: []FeeRatio{
				feeRatio(1, "a", 2, "b"),
				feeRatio(3, "c", 4, "d"),
				feeRatio(5, "e", 6, "f"),
				feeRatio(7, "g", 8, "h"),
				feeRatio(9, "i", 10, "j"),
			},
			exp: "1a:2b,3c:4d,5e:6f,7g:8h,9i:10j",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual string
			testFunc := func() {
				actual = FeeRatiosString(tc.ratios)
			}
			require.NotPanics(t, testFunc, "FeeRatiosString")
			assert.Equal(t, tc.exp, actual, "FeeRatiosString result")
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

			assertions.AssertErrorValue(t, err, tc.exp, "Validate")
		})
	}
}

func TestFeeRatio_Equals(t *testing.T) {
	feeRatio := func(priceAmount int64, priceDenom string, feeAmount int64, feeDenom string) FeeRatio {
		return FeeRatio{
			Price: sdk.Coin{Denom: priceDenom, Amount: sdkmath.NewInt(priceAmount)},
			Fee:   sdk.Coin{Denom: feeDenom, Amount: sdkmath.NewInt(feeAmount)},
		}
	}

	tests := []struct {
		name  string
		base  FeeRatio
		other FeeRatio
		exp   bool
	}{
		{
			name:  "empty ratios",
			base:  FeeRatio{},
			other: FeeRatio{},
			exp:   true,
		},
		{
			name:  "zero amounts",
			base:  feeRatio(0, "pdenom", 0, "fdenom"),
			other: feeRatio(0, "pdenom", 0, "fdenom"),
			exp:   true,
		},
		{
			name:  "same price and fee",
			base:  feeRatio(1, "pdenom", 2, "fdenom"),
			other: feeRatio(1, "pdenom", 2, "fdenom"),
			exp:   true,
		},
		{
			name:  "different base price amount",
			base:  feeRatio(3, "pdenom", 2, "fdenom"),
			other: feeRatio(1, "pdenom", 2, "fdenom"),
			exp:   false,
		},
		{
			name:  "different base price denom",
			base:  feeRatio(1, "qdenom", 2, "fdenom"),
			other: feeRatio(1, "pdenom", 2, "fdenom"),
			exp:   false,
		},
		{
			name:  "different base fee amount",
			base:  feeRatio(1, "pdenom", 3, "fdenom"),
			other: feeRatio(1, "pdenom", 2, "fdenom"),
			exp:   false,
		},
		{
			name:  "different base fee denom",
			base:  feeRatio(1, "pdenom", 2, "gdenom"),
			other: feeRatio(1, "pdenom", 2, "fdenom"),
			exp:   false,
		},
		{
			name:  "different other price amount",
			base:  feeRatio(1, "pdenom", 2, "fdenom"),
			other: feeRatio(3, "pdenom", 2, "fdenom"),
			exp:   false,
		},
		{
			name:  "different other price denom",
			base:  feeRatio(1, "pdenom", 2, "fdenom"),
			other: feeRatio(1, "qdenom", 2, "fdenom"),
			exp:   false,
		},
		{
			name:  "different other fee amount",
			base:  feeRatio(1, "pdenom", 2, "fdenom"),
			other: feeRatio(1, "pdenom", 3, "fdenom"),
			exp:   false,
		},
		{
			name:  "different other fee denom",
			base:  feeRatio(1, "pdenom", 2, "fdenom"),
			other: feeRatio(1, "pdenom", 2, "gdenom"),
			exp:   false,
		},
		{
			name:  "negative base price amount",
			base:  feeRatio(-1, "pdenom", 2, "fdenom"),
			other: feeRatio(1, "pdenom", 2, "fdenom"),
			exp:   false,
		},
		{
			name:  "negative other price amount",
			base:  feeRatio(1, "pdenom", 2, "fdenom"),
			other: feeRatio(-1, "pdenom", 2, "fdenom"),
			exp:   false,
		},
		{
			name:  "negative base fee amount",
			base:  feeRatio(1, "pdenom", -2, "fdenom"),
			other: feeRatio(1, "pdenom", 2, "fdenom"),
			exp:   false,
		},
		{
			name:  "negative other fee amount",
			base:  feeRatio(1, "pdenom", 2, "fdenom"),
			other: feeRatio(1, "pdenom", -2, "fdenom"),
			exp:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual bool
			testFunc := func() {
				actual = tc.base.Equals(tc.other)
			}
			require.NotPanics(t, testFunc, "%s.Equals(%s)", tc.base, tc.other)
			assert.Equal(t, tc.exp, actual, "%s.Equals(%s) result", tc.base, tc.other)
		})
	}
}

func TestFeeRatio_ApplyTo(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	bigCoin := func(amount string, denom string) sdk.Coin {
		amt := newInt(t, amount)
		return sdk.Coin{Denom: denom, Amount: amt}
	}
	feeRatio := func(priceAmount int64, priceDenom string, feeAmount int64, feeDenom string) FeeRatio {
		return FeeRatio{
			Price: coin(priceAmount, priceDenom),
			Fee:   coin(feeAmount, feeDenom),
		}
	}

	tests := []struct {
		name   string
		ratio  FeeRatio
		price  sdk.Coin
		exp    sdk.Coin
		expErr string
	}{
		{
			name:   "wrong denom",
			ratio:  feeRatio(1, "pdenom", 1, "fdenom"),
			price:  coin(1, "fdenom"),
			expErr: "cannot apply ratio 1pdenom:1fdenom to price 1fdenom: incorrect price denom",
		},
		{
			name:   "ratio price amount is zero",
			ratio:  feeRatio(0, "pdenom", 1, "fdenom"),
			price:  coin(1, "pdenom"),
			expErr: "cannot apply ratio 0pdenom:1fdenom to price 1pdenom: division by zero",
		},
		{
			name:   "indivisible",
			ratio:  feeRatio(14, "pdenom", 1, "fdenom"),
			price:  coin(7, "pdenom"),
			expErr: "cannot apply ratio 14pdenom:1fdenom to price 7pdenom: price amount cannot be evenly divided by ratio price",
		},
		{
			name:  "price equals ratio price",
			ratio: feeRatio(55, "pdenom", 3, "fdenom"),
			price: coin(55, "pdenom"),
			exp:   coin(3, "fdenom"),
		},
		{
			name:  "three times ratio price",
			ratio: feeRatio(13, "pdenom", 17, "fdenom"),
			price: coin(39, "pdenom"),
			exp:   coin(51, "fdenom"),
		},
		{
			name:  "very big price",
			ratio: feeRatio(1_000_000_000, "pdenom", 3, "fdenom"),
			price: bigCoin("1000000000000000000000", "pdenom"),
			exp:   coin(3_000_000_000_000, "fdenom"),
		},
		{
			name:  "same price and fee denoms",
			ratio: feeRatio(100, "nhash", 1, "nhash"),
			price: coin(5_000_000_000, "nhash"),
			exp:   coin(50_000_000, "nhash"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.expErr) > 0 {
				tc.exp = coin(0, "")
			}
			var actual sdk.Coin
			var err error
			testFunc := func() {
				actual, err = tc.ratio.ApplyTo(tc.price)
			}
			require.NotPanics(t, testFunc, "%s.ApplyTo(%s)", tc.ratio, tc.price)
			assertions.AssertErrorValue(t, err, tc.expErr, "%s.ApplyTo(%s) error", tc.ratio, tc.price)
			assert.Equal(t, tc.exp.String(), actual.String(), "%s.ApplyTo(%s) result", tc.ratio, tc.price)
		})
	}
}

func TestFeeRatio_ApplyToLoosely(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	bigCoin := func(amount string, denom string) sdk.Coin {
		amt := newInt(t, amount)
		return sdk.Coin{Denom: denom, Amount: amt}
	}
	feeRatio := func(priceAmount int64, priceDenom string, feeAmount int64, feeDenom string) FeeRatio {
		return FeeRatio{
			Price: coin(priceAmount, priceDenom),
			Fee:   coin(feeAmount, feeDenom),
		}
	}

	tests := []struct {
		name   string
		ratio  FeeRatio
		price  sdk.Coin
		exp    sdk.Coin
		expErr string
	}{
		{
			name:   "wrong denom",
			ratio:  feeRatio(1, "pdenom", 1, "fdenom"),
			price:  coin(1, "fdenom"),
			expErr: "cannot apply ratio 1pdenom:1fdenom to price 1fdenom: incorrect price denom",
		},
		{
			name:   "ratio price amount is zero",
			ratio:  feeRatio(0, "pdenom", 1, "fdenom"),
			price:  coin(1, "pdenom"),
			expErr: "cannot apply ratio 0pdenom:1fdenom to price 1pdenom: division by zero",
		},
		{
			name:  "price amount is less than ratio price amount",
			ratio: feeRatio(14, "pdenom", 3, "fdenom"),
			price: coin(7, "pdenom"),
			exp:   coin(2, "fdenom"), // 7 * 3 / 14 = 1.5 => 2
		},
		{
			name:  "price amount is not evenly divisible by ratio",
			ratio: feeRatio(7, "pdenom", 3, "fdenom"),
			price: coin(71, "pdenom"),
			exp:   coin(31, "fdenom"), // 71 * 3 / 7 = 30.4 => 31
		},
		{
			name:  "price equals ratio price",
			ratio: feeRatio(55, "pdenom", 3, "fdenom"),
			price: coin(55, "pdenom"),
			exp:   coin(3, "fdenom"),
		},
		{
			name:  "three times ratio price",
			ratio: feeRatio(13, "pdenom", 17, "fdenom"),
			price: coin(39, "pdenom"),
			exp:   coin(51, "fdenom"),
		},
		{
			name:  "very big price",
			ratio: feeRatio(1_000_000_000, "pdenom", 3, "fdenom"),
			price: bigCoin("1000000000000000000000", "pdenom"),
			exp:   coin(3_000_000_000_000, "fdenom"),
		},
		{
			name:  "same price and fee denoms",
			ratio: feeRatio(100, "nhash", 1, "nhash"),
			price: coin(5_000_000_000, "nhash"),
			exp:   coin(50_000_000, "nhash"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.expErr) > 0 {
				tc.exp = coin(0, "")
			}
			var actual sdk.Coin
			var err error
			testFunc := func() {
				actual, err = tc.ratio.ApplyToLoosely(tc.price)
			}
			require.NotPanics(t, testFunc, "%s.ApplyToLoosely(%s)", tc.ratio, tc.price)
			assertions.AssertErrorValue(t, err, tc.expErr, "%s.ApplyToLoosely(%s) error", tc.ratio, tc.price)
			assert.Equal(t, tc.exp.String(), actual.String(), "%s.ApplyToLoosely(%s) result", tc.ratio, tc.price)
		})
	}
}

func TestIntersectionOfFeeRatios(t *testing.T) {
	feeRatio := func(priceAmount int64, priceDenom string, feeAmount int64, feeDenom string) FeeRatio {
		return FeeRatio{
			Price: sdk.Coin{Denom: priceDenom, Amount: sdkmath.NewInt(priceAmount)},
			Fee:   sdk.Coin{Denom: feeDenom, Amount: sdkmath.NewInt(feeAmount)},
		}
	}

	tests := []struct {
		name     string
		options1 []FeeRatio
		options2 []FeeRatio
		expected []FeeRatio
	}{
		{name: "nil nil", options1: nil, options2: nil, expected: nil},
		{name: "nil empty", options1: nil, options2: []FeeRatio{}, expected: nil},
		{name: "empty nil", options1: []FeeRatio{}, options2: nil, expected: nil},
		{name: "empty empty", options1: []FeeRatio{}, options2: []FeeRatio{}, expected: nil},
		{
			name:     "one nil",
			options1: []FeeRatio{feeRatio(1, "spicy", 2, "lemon")},
			options2: nil,
			expected: nil,
		},
		{
			name:     "nil one",
			options1: nil,
			options2: []FeeRatio{feeRatio(1, "spicy", 2, "lemon")},
			expected: nil,
		},
		{
			name:     "one one equal",
			options1: []FeeRatio{feeRatio(1, "spicy", 2, "lemon")},
			options2: []FeeRatio{feeRatio(1, "spicy", 2, "lemon")},
			expected: []FeeRatio{feeRatio(1, "spicy", 2, "lemon")},
		},
		{
			name:     "one one diff first price amount",
			options1: []FeeRatio{feeRatio(3, "spicy", 2, "lemon")},
			options2: []FeeRatio{feeRatio(1, "spicy", 2, "lemon")},
			expected: nil,
		},
		{
			name:     "one one diff first price denom",
			options1: []FeeRatio{feeRatio(1, "bland", 2, "lemon")},
			options2: []FeeRatio{feeRatio(1, "spicy", 2, "lemon")},
			expected: nil,
		},
		{
			name:     "one one diff first fee amount",
			options1: []FeeRatio{feeRatio(1, "spicy", 3, "lemon")},
			options2: []FeeRatio{feeRatio(1, "spicy", 2, "lemon")},
			expected: nil,
		},
		{
			name:     "one one diff first fee denom",
			options1: []FeeRatio{feeRatio(1, "spicy", 2, "grape")},
			options2: []FeeRatio{feeRatio(1, "spicy", 2, "lemon")},
			expected: nil,
		},
		{
			name:     "one one diff second price amount",
			options1: []FeeRatio{feeRatio(1, "spicy", 2, "lemon")},
			options2: []FeeRatio{feeRatio(3, "spicy", 2, "lemon")},
			expected: nil,
		},
		{
			name:     "one one diff second price denom",
			options1: []FeeRatio{feeRatio(1, "spicy", 2, "lemon")},
			options2: []FeeRatio{feeRatio(1, "bland", 2, "lemon")},
			expected: nil,
		},
		{
			name:     "one one diff second fee amount",
			options1: []FeeRatio{feeRatio(1, "spicy", 2, "lemon")},
			options2: []FeeRatio{feeRatio(1, "spicy", 3, "lemon")},
			expected: nil,
		},
		{
			name:     "one one diff second fee denom",
			options1: []FeeRatio{feeRatio(1, "spicy", 2, "lemon")},
			options2: []FeeRatio{feeRatio(1, "spicy", 2, "bland")},
			expected: nil,
		},
		{
			name: "three three two common",
			options1: []FeeRatio{
				feeRatio(1, "lamp", 2, "bag"),
				feeRatio(3, "keys", 4, "phone"),
				feeRatio(5, "fan", 6, "bottle"),
			},
			options2: []FeeRatio{
				feeRatio(3, "kays", 4, "phone"),
				feeRatio(5, "fan", 6, "bottle"),
				feeRatio(1, "lamp", 2, "bag"),
			},
			expected: []FeeRatio{
				feeRatio(1, "lamp", 2, "bag"),
				feeRatio(5, "fan", 6, "bottle"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []FeeRatio
			testFunc := func() {
				actual = IntersectionOfFeeRatios(tc.options1, tc.options2)
			}
			require.NotPanics(t, testFunc, "IntersectionOfFeeRatios")
			assert.Equal(t, tc.expected, actual, "IntersectionOfFeeRatios result")
		})
	}
}

func TestContainsFeeRatio(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	ratio := func(price, fee sdk.Coin) FeeRatio {
		return FeeRatio{Price: price, Fee: fee}
	}

	tests := []struct {
		name   string
		vals   []FeeRatio
		toFind FeeRatio
		exp    bool
	}{
		{
			name:   "nil vals",
			vals:   nil,
			toFind: ratio(coin(1, "pear"), coin(2, "fig")),
			exp:    false,
		},
		{
			name:   "empty vals",
			vals:   []FeeRatio{},
			toFind: ratio(coin(1, "pear"), coin(2, "fig")),
			exp:    false,
		},
		{
			name:   "one val, same to find",
			vals:   []FeeRatio{ratio(coin(1, "pear"), coin(2, "fig"))},
			toFind: ratio(coin(1, "pear"), coin(2, "fig")),
			exp:    true,
		},
		{
			name:   "one val, diff price amount",
			vals:   []FeeRatio{ratio(coin(1, "pear"), coin(2, "fig"))},
			toFind: ratio(coin(3, "pear"), coin(2, "fig")),
			exp:    false,
		},
		{
			name:   "one val, diff price denom",
			vals:   []FeeRatio{ratio(coin(1, "pear"), coin(2, "fig"))},
			toFind: ratio(coin(1, "prune"), coin(2, "fig")),
			exp:    false,
		},
		{
			name:   "one val, diff fee amount",
			vals:   []FeeRatio{ratio(coin(1, "pear"), coin(2, "fig"))},
			toFind: ratio(coin(1, "pear"), coin(3, "fig")),
			exp:    false,
		},
		{
			name:   "one val, diff fee denom",
			vals:   []FeeRatio{ratio(coin(1, "pear"), coin(2, "fig"))},
			toFind: ratio(coin(1, "pear"), coin(2, "grape")),
			exp:    false,
		},
		{
			name:   "one bad val, same",
			vals:   []FeeRatio{ratio(coin(-5, ""), coin(-6, ""))},
			toFind: ratio(coin(-5, ""), coin(-6, "")),
			exp:    true,
		},
		{
			name: "three vals, not to find",
			vals: []FeeRatio{
				ratio(coin(1, "pear"), coin(2, "fig")),
				ratio(coin(1, "prune"), coin(3, "fig")),
				ratio(coin(1, "plum"), coin(4, "fig")),
			},
			toFind: ratio(coin(1, "pineapple"), coin(5, "fig")),
			exp:    false,
		},
		{
			name: "three vals, first",
			vals: []FeeRatio{
				ratio(coin(1, "pear"), coin(2, "fig")),
				ratio(coin(1, "prune"), coin(3, "fig")),
				ratio(coin(1, "plum"), coin(4, "fig")),
			},
			toFind: ratio(coin(1, "pear"), coin(2, "fig")),
			exp:    true,
		},
		{
			name: "three vals, second",
			vals: []FeeRatio{
				ratio(coin(1, "pear"), coin(2, "fig")),
				ratio(coin(1, "prune"), coin(3, "fig")),
				ratio(coin(1, "plum"), coin(4, "fig")),
			},
			toFind: ratio(coin(1, "prune"), coin(3, "fig")),
			exp:    true,
		},
		{
			name: "three vals, third",
			vals: []FeeRatio{
				ratio(coin(1, "pear"), coin(2, "fig")),
				ratio(coin(1, "prune"), coin(3, "fig")),
				ratio(coin(1, "plum"), coin(4, "fig")),
			},
			toFind: ratio(coin(1, "plum"), coin(4, "fig")),
			exp:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual bool
			tesFunc := func() {
				actual = ContainsFeeRatio(tc.vals, tc.toFind)
			}
			require.NotPanics(t, tesFunc, "ContainsFeeRatio(%q, %q)", FeeRatiosString(tc.vals), tc.toFind)
			assert.Equal(t, tc.exp, actual, "ContainsFeeRatio(%q, %q)", FeeRatiosString(tc.vals), tc.toFind)
		})
	}
}

func TestContainsSameFeeRatioDenoms(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	ratio := func(price, fee sdk.Coin) FeeRatio {
		return FeeRatio{Price: price, Fee: fee}
	}

	tests := []struct {
		name   string
		vals   []FeeRatio
		toFind FeeRatio
		exp    bool
	}{
		{
			name:   "nil vals",
			vals:   nil,
			toFind: ratio(coin(1, "plum"), coin(2, "fig")),
			exp:    false,
		},
		{
			name:   "empty vals",
			vals:   []FeeRatio{},
			toFind: ratio(coin(1, "plum"), coin(2, "fig")),
			exp:    false,
		},
		{
			name:   "one val, same",
			vals:   []FeeRatio{ratio(coin(1, "plum"), coin(2, "fig"))},
			toFind: ratio(coin(1, "plum"), coin(2, "fig")),
			exp:    true,
		},
		{
			name:   "one val, diff price amount",
			vals:   []FeeRatio{ratio(coin(1, "plum"), coin(2, "fig"))},
			toFind: ratio(coin(3, "plum"), coin(2, "fig")),
			exp:    true,
		},
		{
			name:   "one val, diff price denom",
			vals:   []FeeRatio{ratio(coin(1, "plum"), coin(2, "fig"))},
			toFind: ratio(coin(1, "prune"), coin(2, "fig")),
			exp:    false,
		},
		{
			name:   "one val, diff fee amount",
			vals:   []FeeRatio{ratio(coin(1, "plum"), coin(2, "fig"))},
			toFind: ratio(coin(1, "plum"), coin(3, "fig")),
			exp:    true,
		},
		{
			name:   "one val, diff fee denom",
			vals:   []FeeRatio{ratio(coin(1, "plum"), coin(2, "fig"))},
			toFind: ratio(coin(1, "plum"), coin(2, "grape")),
			exp:    false,
		},
		{
			name: "three vals, not to find",
			vals: []FeeRatio{
				ratio(coin(1, "prune"), coin(2, "fig")),
				ratio(coin(3, "plum"), coin(4, "grape")),
				ratio(coin(5, "peach"), coin(6, "honeydew")),
			},
			toFind: ratio(coin(7, "pineapple"), coin(8, "jackfruit")),
			exp:    false,
		},
		{
			name: "three vals, first price denom second fee denom",
			vals: []FeeRatio{
				ratio(coin(1, "prune"), coin(2, "fig")),
				ratio(coin(3, "plum"), coin(4, "grape")),
				ratio(coin(5, "peach"), coin(6, "honeydew")),
			},
			toFind: ratio(coin(7, "prune"), coin(8, "grape")),
			exp:    false,
		},
		{
			name: "three vals, first",
			vals: []FeeRatio{
				ratio(coin(1, "prune"), coin(2, "fig")),
				ratio(coin(3, "plum"), coin(4, "grape")),
				ratio(coin(5, "peach"), coin(6, "honeydew")),
			},
			toFind: ratio(coin(7, "prune"), coin(8, "fig")),
			exp:    true,
		},
		{
			name: "three vals, second",
			vals: []FeeRatio{
				ratio(coin(1, "prune"), coin(2, "fig")),
				ratio(coin(3, "plum"), coin(4, "grape")),
				ratio(coin(5, "peach"), coin(6, "honeydew")),
			},
			toFind: ratio(coin(7, "plum"), coin(8, "grape")),
			exp:    true,
		},
		{
			name: "three vals, third",
			vals: []FeeRatio{
				ratio(coin(1, "prune"), coin(2, "fig")),
				ratio(coin(3, "plum"), coin(4, "grape")),
				ratio(coin(5, "peach"), coin(6, "honeydew")),
			},
			toFind: ratio(coin(7, "peach"), coin(8, "honeydew")),
			exp:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual bool
			testFunc := func() {
				actual = ContainsSameFeeRatioDenoms(tc.vals, tc.toFind)
			}
			require.NotPanics(t, testFunc, "ContainsSameFeeRatioDenoms(%q, %q)", FeeRatiosString(tc.vals), tc.toFind)
			assert.Equal(t, tc.exp, actual, "ContainsSameFeeRatioDenoms(%q, %q)", FeeRatiosString(tc.vals), tc.toFind)
		})
	}
}

func TestValidateDisjointFeeRatios(t *testing.T) {
	feeRatio := func(priceAmount int64, priceDenom string, feeAmount int64, feeDenom string) FeeRatio {
		return FeeRatio{
			Price: sdk.Coin{Denom: priceDenom, Amount: sdkmath.NewInt(priceAmount)},
			Fee:   sdk.Coin{Denom: feeDenom, Amount: sdkmath.NewInt(feeAmount)},
		}
	}

	tests := []struct {
		name     string
		field    string
		toAdd    []FeeRatio
		toRemove []FeeRatio
		expErr   string
	}{
		{name: "nil nil", toAdd: nil, toRemove: nil, expErr: ""},
		{name: "nil empty", toAdd: nil, toRemove: []FeeRatio{}, expErr: ""},
		{name: "empty nil", toAdd: []FeeRatio{}, toRemove: nil, expErr: ""},
		{name: "empty empty", toAdd: []FeeRatio{}, toRemove: []FeeRatio{}, expErr: ""},
		{
			name:     "one nil",
			toAdd:    []FeeRatio{feeRatio(1, "spicy", 2, "lemon")},
			toRemove: nil,
			expErr:   "",
		},
		{
			name:     "nil one",
			toAdd:    nil,
			toRemove: []FeeRatio{feeRatio(1, "spicy", 2, "lemon")},
			expErr:   "",
		},
		{
			name:     "one one equal",
			field:    "<fieldname>",
			toAdd:    []FeeRatio{feeRatio(1, "spicy", 2, "lemon")},
			toRemove: []FeeRatio{feeRatio(1, "spicy", 2, "lemon")},
			expErr:   "cannot add and remove the same <fieldname> ratios 1spicy:2lemon",
		},
		{
			name:     "one one diff first price amount",
			toAdd:    []FeeRatio{feeRatio(3, "spicy", 2, "lemon")},
			toRemove: []FeeRatio{feeRatio(1, "spicy", 2, "lemon")},
			expErr:   "",
		},
		{
			name:     "one one diff first price denom",
			toAdd:    []FeeRatio{feeRatio(1, "bland", 2, "lemon")},
			toRemove: []FeeRatio{feeRatio(1, "spicy", 2, "lemon")},
			expErr:   "",
		},
		{
			name:     "one one diff first fee amount",
			toAdd:    []FeeRatio{feeRatio(1, "spicy", 3, "lemon")},
			toRemove: []FeeRatio{feeRatio(1, "spicy", 2, "lemon")},
			expErr:   "",
		},
		{
			name:     "one one diff first fee denom",
			toAdd:    []FeeRatio{feeRatio(1, "spicy", 2, "grape")},
			toRemove: []FeeRatio{feeRatio(1, "spicy", 2, "lemon")},
			expErr:   "",
		},
		{
			name:     "one one diff second price amount",
			toAdd:    []FeeRatio{feeRatio(1, "spicy", 2, "lemon")},
			toRemove: []FeeRatio{feeRatio(3, "spicy", 2, "lemon")},
			expErr:   "",
		},
		{
			name:     "one one diff second price denom",
			toAdd:    []FeeRatio{feeRatio(1, "spicy", 2, "lemon")},
			toRemove: []FeeRatio{feeRatio(1, "bland", 2, "lemon")},
			expErr:   "",
		},
		{
			name:     "one one diff second fee amount",
			toAdd:    []FeeRatio{feeRatio(1, "spicy", 2, "lemon")},
			toRemove: []FeeRatio{feeRatio(1, "spicy", 3, "lemon")},
			expErr:   "",
		},
		{
			name:     "one one diff second fee denom",
			toAdd:    []FeeRatio{feeRatio(1, "spicy", 2, "lemon")},
			toRemove: []FeeRatio{feeRatio(1, "spicy", 2, "bland")},
			expErr:   "",
		},
		{
			name:  "three three two common",
			field: "test field name",
			toAdd: []FeeRatio{
				feeRatio(1, "lamp", 2, "bag"),
				feeRatio(3, "keys", 4, "phone"),
				feeRatio(5, "fan", 6, "bottle"),
			},
			toRemove: []FeeRatio{
				feeRatio(3, "kays", 4, "phone"),
				feeRatio(5, "fan", 6, "bottle"),
				feeRatio(1, "lamp", 2, "bag"),
			},
			expErr: "cannot add and remove the same test field name ratios 1lamp:2bag,5fan:6bottle",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = ValidateDisjointFeeRatios(tc.field, tc.toAdd, tc.toRemove)
			}
			require.NotPanics(t, testFunc, "ValidateDisjointFeeRatios")

			assertions.AssertErrorValue(t, err, tc.expErr, "ValidateDisjointFeeRatios")
		})
	}
}

func TestValidateAddRemoveFeeRatiosWithExisting(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	ratio := func(price, fee sdk.Coin) FeeRatio {
		return FeeRatio{Price: price, Fee: fee}
	}
	noRemErr := func(field, fee string) string {
		return fmt.Sprintf("cannot remove %s ratio fee %q: no such ratio exists", field, fee)
	}
	noAddErr := func(field, fee string) string {
		return fmt.Sprintf("cannot add %s ratio fee %q: ratio with those denoms already exists", field, fee)
	}

	tests := []struct {
		name     string
		field    string
		existing []FeeRatio
		toAdd    []FeeRatio
		toRemove []FeeRatio
		expErr   string
	}{
		{
			name:     "nil existing",
			field:    "DLEIF",
			existing: nil,
			toAdd:    []FeeRatio{ratio(coin(1, "pineapple"), coin(2, "grape"))},
			toRemove: []FeeRatio{ratio(coin(1, "pineapple"), coin(3, "fig"))},
			expErr:   noRemErr("DLEIF", "1pineapple:3fig"),
		},
		{
			name:     "empty existing",
			field:    "DLEIF",
			existing: []FeeRatio{},
			toAdd:    []FeeRatio{ratio(coin(1, "pineapple"), coin(2, "grape"))},
			toRemove: []FeeRatio{ratio(coin(1, "pineapple"), coin(3, "fig"))},
			expErr:   noRemErr("DLEIF", "1pineapple:3fig"),
		},
		{
			name:     "nil toAdd",
			field:    "DLEIF",
			existing: []FeeRatio{ratio(coin(1, "pineapple"), coin(3, "fig"))},
			toAdd:    nil,
			toRemove: []FeeRatio{ratio(coin(1, "pineapple"), coin(3, "fig"))},
			expErr:   "",
		},
		{
			name:     "empty toAdd",
			field:    "DLEIF",
			existing: []FeeRatio{ratio(coin(1, "pineapple"), coin(3, "fig"))},
			toAdd:    []FeeRatio{},
			toRemove: []FeeRatio{ratio(coin(1, "pineapple"), coin(3, "fig"))},
			expErr:   "",
		},
		{
			name:     "nil toRemove",
			field:    "DLEIF",
			existing: []FeeRatio{ratio(coin(1, "pineapple"), coin(3, "fig"))},
			toAdd:    []FeeRatio{ratio(coin(1, "pineapple"), coin(2, "grape"))},
			toRemove: nil,
			expErr:   "",
		},
		{
			name:     "empty toRemove",
			field:    "DLEIF",
			existing: []FeeRatio{ratio(coin(1, "pineapple"), coin(3, "fig"))},
			toAdd:    []FeeRatio{ratio(coin(1, "pineapple"), coin(2, "grape"))},
			toRemove: []FeeRatio{},
			expErr:   "",
		},
		{
			name:  "to remove not in existing, diff price amounts",
			field: "DLEIF",
			existing: []FeeRatio{
				ratio(coin(111, "pineapple"), coin(2, "fig")),
				ratio(coin(333, "pineapple"), coin(4, "grape")),
			},
			toAdd: nil,
			toRemove: []FeeRatio{
				ratio(coin(112, "pineapple"), coin(2, "fig")),
				ratio(coin(334, "pineapple"), coin(4, "grape")),
			},
			expErr: joinErrs(
				noRemErr("DLEIF", "112pineapple:2fig"),
				noRemErr("DLEIF", "334pineapple:4grape"),
			),
		},
		{
			name:  "to remove not in existing, diff price denoms",
			field: "DleiF",
			existing: []FeeRatio{
				ratio(coin(111, "prune"), coin(2, "fig")),
				ratio(coin(333, "plum"), coin(4, "grape")),
			},
			toAdd: nil,
			toRemove: []FeeRatio{
				ratio(coin(111, "plum"), coin(2, "fig")),
				ratio(coin(333, "prune"), coin(4, "grape")),
			},
			expErr: joinErrs(
				noRemErr("DleiF", "111plum:2fig"),
				noRemErr("DleiF", "333prune:4grape"),
			),
		},
		{
			name:  "to remove not in existing, diff fee amounts",
			field: "DleiF",
			existing: []FeeRatio{
				ratio(coin(111, "prune"), coin(2, "fig")),
				ratio(coin(333, "plum"), coin(4, "grape")),
			},
			toAdd: nil,
			toRemove: []FeeRatio{
				ratio(coin(111, "prune"), coin(3, "fig")),
				ratio(coin(333, "plum"), coin(5, "grape")),
			},
			expErr: joinErrs(
				noRemErr("DleiF", "111prune:3fig"),
				noRemErr("DleiF", "333plum:5grape"),
			),
		},
		{
			name:  "to remove not in existing, diff fee denoms",
			field: "DleiF",
			existing: []FeeRatio{
				ratio(coin(111, "prune"), coin(2, "fig")),
				ratio(coin(333, "plum"), coin(4, "grape")),
			},
			toAdd: nil,
			toRemove: []FeeRatio{
				ratio(coin(111, "prune"), coin(2, "grape")),
				ratio(coin(333, "plum"), coin(4, "fig")),
			},
			expErr: joinErrs(
				noRemErr("DleiF", "111prune:2grape"),
				noRemErr("DleiF", "333plum:4fig"),
			),
		},
		{
			name:  "to add already exists",
			field: "yard",
			existing: []FeeRatio{
				ratio(coin(111, "prune"), coin(2, "fig")),
				ratio(coin(333, "plum"), coin(4, "grape")),
				ratio(coin(555, "peach"), coin(6, "honeydew")),
			},
			toAdd: []FeeRatio{
				ratio(coin(321, "prune"), coin(7, "fig")),
				ratio(coin(654, "peach"), coin(8, "honeydew")),
			},
			toRemove: []FeeRatio{ratio(coin(333, "plum"), coin(4, "grape"))},
			expErr: joinErrs(
				noAddErr("yard", "321prune:7fig"),
				noAddErr("yard", "654peach:8honeydew"),
			),
		},
		{
			name:  "remove one add another with same denoms",
			field: "lawn",
			existing: []FeeRatio{
				ratio(coin(111, "prune"), coin(2, "fig")),
				ratio(coin(333, "plum"), coin(4, "grape")),
				ratio(coin(555, "peach"), coin(6, "honeydew")),
			},
			toAdd: []FeeRatio{
				ratio(coin(321, "plum"), coin(1, "grape")),
			},
			toRemove: []FeeRatio{
				ratio(coin(333, "plum"), coin(4, "grape")),
			},
			expErr: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var errs []error
			testFunc := func() {
				errs = ValidateAddRemoveFeeRatiosWithExisting(tc.field, tc.existing, tc.toAdd, tc.toRemove)
			}
			require.NotPanics(t, testFunc, "ValidateAddRemoveFeeRatiosWithExisting")
			err := errors.Join(errs...)
			assertions.AssertErrorValue(t, err, tc.expErr, "ValidateAddRemoveFeeRatiosWithExisting error")
		})
	}
}

func TestValidateRatioDenoms(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	ratio := func(price, fee sdk.Coin) FeeRatio {
		return FeeRatio{Price: price, Fee: fee}
	}
	noBuyerDenomErr := func(denom string) string {
		return fmt.Sprintf("seller settlement fee ratios have price denom %q but there "+
			"are no buyer settlement fee ratios with that price denom", denom)
	}
	noSellerDenomErr := func(denom string) string {
		return fmt.Sprintf("buyer settlement fee ratios have price denom %q but there "+
			"is not a seller settlement fee ratio with that price denom", denom)
	}

	tests := []struct {
		name   string
		seller []FeeRatio
		buyer  []FeeRatio
		expErr string
	}{
		{name: "nil seller, nil buyer", seller: nil, buyer: nil, expErr: ""},
		{name: "empty seller, nil buyer", seller: []FeeRatio{}, buyer: nil, expErr: ""},
		{name: "nil seller, empty buyer", seller: nil, buyer: []FeeRatio{}, expErr: ""},
		{name: "empty seller, empty buyer", seller: []FeeRatio{}, buyer: []FeeRatio{}, expErr: ""},
		{
			name:   "nil seller, 2 buyer",
			seller: nil,
			buyer: []FeeRatio{
				ratio(coin(1, "apple"), coin(2, "fig")),
				ratio(coin(3, "cucumber"), coin(4, "dill")),
			},
			expErr: "",
		},
		{
			name:   "empty seller, 2 buyer",
			seller: []FeeRatio{},
			buyer: []FeeRatio{
				ratio(coin(1, "apple"), coin(2, "fig")),
				ratio(coin(3, "cucumber"), coin(4, "dill")),
			},
			expErr: "",
		},
		{
			name: "2 seller, nil buyer",
			seller: []FeeRatio{
				ratio(coin(1, "apple"), coin(2, "fig")),
				ratio(coin(3, "cucumber"), coin(4, "dill")),
			},
			buyer:  nil,
			expErr: "",
		},
		{
			name: "2 seller, empty buyer",
			seller: []FeeRatio{
				ratio(coin(1, "apple"), coin(2, "fig")),
				ratio(coin(3, "cucumber"), coin(4, "dill")),
			},
			buyer:  []FeeRatio{},
			expErr: "",
		},
		{
			name:   "1 seller, 1 buyer: different",
			seller: []FeeRatio{ratio(coin(1, "apple"), coin(2, "banana"))},
			buyer:  []FeeRatio{ratio(coin(3, "cucumber"), coin(4, "dill"))},
			expErr: joinErrs(noBuyerDenomErr("apple"), noSellerDenomErr("cucumber")),
		},
		{
			name: "2 seller, 2 buyer: all different",
			seller: []FeeRatio{
				ratio(coin(1, "apple"), coin(2, "banana")),
				ratio(coin(3, "cucumber"), coin(4, "dill")),
			},
			buyer: []FeeRatio{
				ratio(coin(5, "eggplant"), coin(6, "fig")),
				ratio(coin(7, "grape"), coin(8, "honeydew")),
			},
			expErr: joinErrs(
				noBuyerDenomErr("apple"),
				noBuyerDenomErr("cucumber"),
				noSellerDenomErr("eggplant"),
				noSellerDenomErr("grape"),
			),
		},
		{
			name:   "1 seller, 2 buyer with same: same",
			seller: []FeeRatio{ratio(coin(1, "apple"), coin(2, "banana"))},
			buyer: []FeeRatio{
				ratio(coin(5, "apple"), coin(6, "fig")),
				ratio(coin(7, "apple"), coin(8, "honeydew")),
			},
			expErr: "",
		},
		{
			name:   "1 seller, 2 buyer with same: different",
			seller: []FeeRatio{ratio(coin(1, "eggplant"), coin(2, "banana"))},
			buyer: []FeeRatio{
				ratio(coin(5, "apple"), coin(6, "fig")),
				ratio(coin(7, "apple"), coin(8, "honeydew")),
			},
			expErr: joinErrs(noBuyerDenomErr("eggplant"), noSellerDenomErr("apple")),
		},
		{
			name:   "1 seller, 2 buyer: 1 buyer missing",
			seller: []FeeRatio{ratio(coin(1, "apple"), coin(2, "banana"))},
			buyer: []FeeRatio{
				ratio(coin(5, "cucumber"), coin(6, "fig")),
				ratio(coin(7, "apple"), coin(8, "honeydew")),
			},
			expErr: noSellerDenomErr("cucumber"),
		},
		{
			name: "2 seller, 1 buyer",
			seller: []FeeRatio{
				ratio(coin(5, "apple"), coin(6, "banana")),
				ratio(coin(7, "cucumber"), coin(8, "dill")),
			},
			buyer:  []FeeRatio{ratio(coin(1, "apple"), coin(2, "grape"))},
			expErr: noBuyerDenomErr("cucumber"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var errs []error
			testFunc := func() {
				errs = ValidateRatioDenoms(tc.seller, tc.buyer)
			}
			require.NotPanics(t, testFunc, "ValidateRatioDenoms(%q, %q)",
				FeeRatiosString(tc.seller), FeeRatiosString(tc.buyer))
			err := errors.Join(errs...)
			assertions.AssertErrorValue(t, err, tc.expErr, "ValidateRatioDenoms(%q, %q)",
				FeeRatiosString(tc.seller), FeeRatiosString(tc.buyer))
		})
	}
}

func TestValidateAddRemoveFeeOptions(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}

	tests := []struct {
		name     string
		field    string
		toAdd    []sdk.Coin
		toRemove []sdk.Coin
		expErr   string
	}{
		{name: "nil nil", toAdd: nil, toRemove: nil, expErr: ""},
		{name: "nil empty", toAdd: nil, toRemove: []sdk.Coin{}, expErr: ""},
		{name: "empty nil", toAdd: []sdk.Coin{}, toRemove: nil, expErr: ""},
		{name: "empty empty", toAdd: []sdk.Coin{}, toRemove: []sdk.Coin{}, expErr: ""},
		{
			name:     "one nil",
			toAdd:    []sdk.Coin{coin(1, "finger")},
			toRemove: nil,
			expErr:   "",
		},
		{
			name:     "nil one",
			toAdd:    nil,
			toRemove: []sdk.Coin{coin(1, "finger")},
			expErr:   "",
		},
		{
			name:     "one one same",
			field:    "<fieldname>",
			toAdd:    []sdk.Coin{coin(1, "finger")},
			toRemove: []sdk.Coin{coin(1, "finger")},
			expErr:   "cannot add and remove the same <fieldname> options 1finger",
		},
		{
			name:     "one one different first amount",
			toAdd:    []sdk.Coin{coin(2, "finger")},
			toRemove: []sdk.Coin{coin(1, "finger")},
			expErr:   "",
		},
		{
			name:     "one one different first denom",
			toAdd:    []sdk.Coin{coin(1, "toe")},
			toRemove: []sdk.Coin{coin(1, "finger")},
			expErr:   "",
		},
		{
			name:     "one one different second amount",
			toAdd:    []sdk.Coin{coin(1, "finger")},
			toRemove: []sdk.Coin{coin(2, "finger")},
			expErr:   "",
		},
		{
			name:     "one one different second denom",
			toAdd:    []sdk.Coin{coin(1, "finger")},
			toRemove: []sdk.Coin{coin(1, "toe")},
			expErr:   "",
		},
		{
			name:     "three three two common",
			field:    "body part",
			toAdd:    []sdk.Coin{coin(1, "finger"), coin(2, "toe"), coin(3, "elbow")},
			toRemove: []sdk.Coin{coin(5, "toe"), coin(3, "elbow"), coin(1, "finger")},
			expErr:   "cannot add and remove the same body part options 1finger,3elbow",
		},
		{
			name:     "adding one invalid",
			field:    "badabadabada",
			toAdd:    []sdk.Coin{coin(0, "zerocoin")},
			toRemove: nil,
			expErr:   `invalid badabadabada to add option "0zerocoin": amount cannot be zero`,
		},
		{
			name:     "adding two invalid entries",
			field:    "friendly iterative error lookup device (F.I.E.L.D)",
			toAdd:    []sdk.Coin{coin(-1, "bananas"), coin(3, "okaycoin"), coin(0, "zerocoin"), coin(5, "goodgood")},
			toRemove: nil,
			expErr: joinErrs(
				`invalid friendly iterative error lookup device (F.I.E.L.D) to add option "-1bananas": negative coin amount: -1`,
				`invalid friendly iterative error lookup device (F.I.E.L.D) to add option "0zerocoin": amount cannot be zero`,
			),
		},
		{
			name:     "removing one invalid",
			field:    "",
			toAdd:    nil,
			toRemove: []sdk.Coin{coin(-1, "bananas")},
			expErr:   "",
		},
		{
			name:     "adding and removing one invalid",
			field:    "fruits",
			toAdd:    []sdk.Coin{coin(-1, "bananas")},
			toRemove: []sdk.Coin{coin(-1, "bananas")},
			expErr: joinErrs(
				`invalid fruits to add option "-1bananas": negative coin amount: -1`,
				"cannot add and remove the same fruits options -1bananas",
			),
		},
		{
			name:  "multiple errors",
			field: "merrs!",
			toAdd: []sdk.Coin{
				coin(1, "apple"),
				coin(3, "l"),
				coin(99, "banana"),
				coin(-55, "peach"),
				coin(14, "copycoin"),
				coin(5, "grape"),
			},
			toRemove: []sdk.Coin{
				coin(98, "banana"),
				coin(14, "copycoin"),
			},
			expErr: joinErrs(
				`invalid merrs! to add option "3l": invalid denom: l`,
				`invalid merrs! to add option "-55peach": negative coin amount: -55`,
				"cannot add and remove the same merrs! options 14copycoin",
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = ValidateAddRemoveFeeOptions(tc.field, tc.toAdd, tc.toRemove)
			}
			require.NotPanics(t, testFunc, "ValidateAddRemoveFeeOptions")

			assertions.AssertErrorValue(t, err, tc.expErr, "ValidateAddRemoveFeeOptions")
		})
	}
}

func TestValidateAddRemoveFeeOptionsWithExisting(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	noRemErr := func(field, fee string) string {
		return fmt.Sprintf("cannot remove %s flat fee %q: no such fee exists", field, fee)
	}
	noAddErr := func(field, fee string) string {
		return fmt.Sprintf("cannot add %s flat fee %q: fee with that denom already exists", field, fee)
	}

	tests := []struct {
		name     string
		field    string
		existing []sdk.Coin
		toAdd    []sdk.Coin
		toRemove []sdk.Coin
		expErr   string
	}{
		{
			name:     "nil existing",
			field:    "tree",
			existing: nil,
			toAdd:    []sdk.Coin{coin(1, "apple")},
			toRemove: []sdk.Coin{coin(2, "banana")},
			expErr:   noRemErr("tree", "2banana"),
		},
		{
			name:     "empty existing",
			field:    "tree",
			existing: []sdk.Coin{},
			toAdd:    []sdk.Coin{coin(1, "apple")},
			toRemove: []sdk.Coin{coin(2, "banana")},
			expErr:   noRemErr("tree", "2banana"),
		},
		{
			name:     "nil toAdd",
			field:    "fern",
			existing: []sdk.Coin{coin(2, "banana")},
			toAdd:    nil,
			toRemove: []sdk.Coin{coin(2, "banana")},
			expErr:   "",
		},
		{
			name:     "empty toAdd",
			field:    "fern",
			existing: []sdk.Coin{coin(2, "banana")},
			toAdd:    []sdk.Coin{},
			toRemove: []sdk.Coin{coin(2, "banana")},
			expErr:   "",
		},
		{
			name:     "nil toRemove",
			field:    "yucca",
			existing: []sdk.Coin{coin(2, "apple")},
			toAdd:    []sdk.Coin{coin(2, "banana")},
			toRemove: nil,
			expErr:   "",
		},
		{
			name:     "empty toRemove",
			field:    "yucca",
			existing: []sdk.Coin{coin(2, "apple")},
			toAdd:    []sdk.Coin{coin(2, "banana")},
			toRemove: []sdk.Coin{},
			expErr:   "",
		},
		{
			name:     "to remove not in existing, diff amounts",
			field:    "rose",
			existing: []sdk.Coin{coin(1, "apple"), coin(2, "banana")},
			toAdd:    nil,
			toRemove: []sdk.Coin{coin(3, "apple"), coin(4, "banana")},
			expErr: joinErrs(
				noRemErr("rose", "3apple"),
				noRemErr("rose", "4banana"),
			),
		},
		{
			name:     "to remove not in existing, diff denoms",
			field:    "rose",
			existing: []sdk.Coin{coin(1, "apple"), coin(2, "banana")},
			toAdd:    nil,
			toRemove: []sdk.Coin{coin(1, "cactus"), coin(2, "durian")},
			expErr: joinErrs(
				noRemErr("rose", "1cactus"),
				noRemErr("rose", "2durian"),
			),
		},
		{
			name:     "to remove, in existing",
			field:    "rose",
			existing: []sdk.Coin{coin(1, "apple"), coin(2, "banana")},
			toAdd:    nil,
			toRemove: []sdk.Coin{coin(1, "apple"), coin(2, "banana")},
			expErr:   "",
		},
		{
			name:     "to add denom already exists",
			field:    "tuLip",
			existing: []sdk.Coin{coin(1, "apple"), coin(2, "banana")},
			toAdd:    []sdk.Coin{coin(3, "apple"), coin(4, "banana")},
			toRemove: nil,
			expErr: joinErrs(
				noAddErr("tuLip", "3apple"),
				noAddErr("tuLip", "4banana"),
			),
		},
		{
			name:     "remove and add same denom",
			field:    "whoops",
			existing: []sdk.Coin{coin(1, "apple"), coin(2, "banana")},
			toAdd:    []sdk.Coin{coin(3, "apple")},
			toRemove: []sdk.Coin{coin(1, "apple")},
			expErr:   "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var errs []error
			testFunc := func() {
				errs = ValidateAddRemoveFeeOptionsWithExisting(tc.field, tc.existing, tc.toAdd, tc.toRemove)
			}
			require.NotPanics(t, testFunc, "ValidateAddRemoveFeeOptionsWithExisting")
			err := errors.Join(errs...)
			assertions.AssertErrorValue(t, err, tc.expErr, "ValidateAddRemoveFeeOptionsWithExisting error")
		})
	}
}

func TestValidateAccessGrantsField(t *testing.T) {
	addrDup := sdk.AccAddress("duplicate_address___").String()
	addr1 := sdk.AccAddress("address_1___________").String()
	addr2 := sdk.AccAddress("address_2___________").String()
	addr3 := sdk.AccAddress("address_3___________").String()

	tests := []struct {
		name   string
		field  string
		grants []AccessGrant
		exp    string
	}{
		{
			name:   "nil grants",
			field:  "<~FIELD~>",
			grants: nil,
			exp:    "",
		},
		{
			name:   "empty grants",
			field:  "<~FIELD~>",
			grants: []AccessGrant{},
			exp:    "",
		},
		{
			name:  "duplicate address: no field",
			field: "",
			grants: []AccessGrant{
				{Address: addrDup, Permissions: []Permission{Permission_settle}},
				{Address: addrDup, Permissions: []Permission{Permission_cancel}},
			},
			exp: sdk.AccAddress("duplicate_address___").String() + " appears in multiple access grant entries",
		},
		{
			name:  "duplicate address: with field",
			field: "<~FIELD~>",
			grants: []AccessGrant{
				{Address: addrDup, Permissions: []Permission{Permission_settle}},
				{Address: addrDup, Permissions: []Permission{Permission_cancel}},
			},
			exp: sdk.AccAddress("duplicate_address___").String() + " appears in multiple <~FIELD~> access grant entries",
		},
		{
			name:  "three entries: all valid",
			field: "<~FIELD~>",
			grants: []AccessGrant{
				{Address: addr1, Permissions: AllPermissions()},
				{Address: addr2, Permissions: AllPermissions()},
				{Address: addr3, Permissions: AllPermissions()},
			},
			exp: "",
		},
		{
			name:  "three entries: invalid first: no field",
			field: "",
			grants: []AccessGrant{
				{Address: addr1, Permissions: []Permission{-1}},
				{Address: addr2, Permissions: AllPermissions()},
				{Address: addr3, Permissions: AllPermissions()},
			},
			exp: "invalid access grant: permission -1 does not exist for " + addr1,
		},
		{
			name:  "three entries: invalid first: with field",
			field: "<~FIELD~>",
			grants: []AccessGrant{
				{Address: addr1, Permissions: []Permission{-1}},
				{Address: addr2, Permissions: AllPermissions()},
				{Address: addr3, Permissions: AllPermissions()},
			},
			exp: "invalid <~FIELD~> access grant: permission -1 does not exist for " + addr1,
		},
		{
			name:  "three entries: invalid second: no field",
			field: "",
			grants: []AccessGrant{
				{Address: addr1, Permissions: AllPermissions()},
				{Address: addr2, Permissions: []Permission{-1}},
				{Address: addr3, Permissions: AllPermissions()},
			},
			exp: "invalid access grant: permission -1 does not exist for " + addr2,
		},
		{
			name:  "three entries: invalid second: with field",
			field: "<~FIELD~>",
			grants: []AccessGrant{
				{Address: addr1, Permissions: AllPermissions()},
				{Address: addr2, Permissions: []Permission{-1}},
				{Address: addr3, Permissions: AllPermissions()},
			},
			exp: "invalid <~FIELD~> access grant: permission -1 does not exist for " + addr2,
		},
		{
			name:  "three entries: invalid third: no field",
			field: "",
			grants: []AccessGrant{
				{Address: addr1, Permissions: AllPermissions()},
				{Address: addr2, Permissions: AllPermissions()},
				{Address: addr3, Permissions: []Permission{-1}},
			},
			exp: "invalid access grant: permission -1 does not exist for " + addr3,
		},
		{
			name:  "three entries: invalid third: with field",
			field: "<~FIELD~>",
			grants: []AccessGrant{
				{Address: addr1, Permissions: AllPermissions()},
				{Address: addr2, Permissions: AllPermissions()},
				{Address: addr3, Permissions: []Permission{-1}},
			},
			exp: "invalid <~FIELD~> access grant: permission -1 does not exist for " + addr3,
		},
		{
			name:  "three entries: only valid first: no field",
			field: "",
			grants: []AccessGrant{
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
			name:  "three entries: only valid first: with field",
			field: "<~FIELD~>",
			grants: []AccessGrant{
				{Address: addr1, Permissions: AllPermissions()},
				{Address: addr2, Permissions: []Permission{0}},
				{Address: addr3, Permissions: []Permission{-1}},
			},
			exp: joinErrs(
				"invalid <~FIELD~> access grant: permission is unspecified for "+addr2,
				"invalid <~FIELD~> access grant: permission -1 does not exist for "+addr3,
			),
		},
		{
			name:  "three entries: only valid second: no field",
			field: "",
			grants: []AccessGrant{
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
			name:  "three entries: only valid second: with field",
			field: "<~FIELD~>",
			grants: []AccessGrant{
				{Address: addr1, Permissions: []Permission{0}},
				{Address: addr2, Permissions: AllPermissions()},
				{Address: addr3, Permissions: []Permission{-1}},
			},
			exp: joinErrs(
				"invalid <~FIELD~> access grant: permission is unspecified for "+addr1,
				"invalid <~FIELD~> access grant: permission -1 does not exist for "+addr3,
			),
		},
		{
			name:  "three entries: only valid third: no field",
			field: "",
			grants: []AccessGrant{
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
			name:  "three entries: only valid third: with field",
			field: "<~FIELD~>",
			grants: []AccessGrant{
				{Address: addr1, Permissions: []Permission{0}},
				{Address: addr2, Permissions: []Permission{-1}},
				{Address: addr3, Permissions: AllPermissions()},
			},
			exp: joinErrs(
				"invalid <~FIELD~> access grant: permission is unspecified for "+addr1,
				"invalid <~FIELD~> access grant: permission -1 does not exist for "+addr2,
			),
		},
		{
			name:  "three entries: all same address: no field",
			field: "",
			grants: []AccessGrant{
				{Address: addrDup, Permissions: AllPermissions()},
				{Address: addrDup, Permissions: AllPermissions()},
				{Address: addrDup, Permissions: AllPermissions()},
			},
			exp: addrDup + " appears in multiple access grant entries",
		},
		{
			name:  "three entries: all same address: with field",
			field: "<~FIELD~>",
			grants: []AccessGrant{
				{Address: addrDup, Permissions: AllPermissions()},
				{Address: addrDup, Permissions: AllPermissions()},
				{Address: addrDup, Permissions: AllPermissions()},
			},
			exp: addrDup + " appears in multiple <~FIELD~> access grant entries",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateAccessGrantsField(tc.field, tc.grants)

			assertions.AssertErrorValue(t, err, tc.exp, "ValidateAccessGrantsField")
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
			exp:  `invalid access grant: invalid address "invalid_address_____": decoding bech32 failed: invalid separator index -1`,
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

			assertions.AssertErrorValue(t, err, tc.exp, "Validate")
		})
	}
}

func TestAccessGrant_ValidateInField(t *testing.T) {
	addr := sdk.AccAddress("addr________________").String()
	tests := []struct {
		name  string
		a     AccessGrant
		field string
		exp   string
	}{
		{
			name:  "control",
			a:     AccessGrant{Address: addr, Permissions: AllPermissions()},
			field: "meadow",
			exp:   "",
		},
		{
			name:  "invalid address: no field",
			a:     AccessGrant{Address: "invalid_address_____", Permissions: []Permission{Permission_settle}},
			field: "",
			exp:   `invalid access grant: invalid address "invalid_address_____": decoding bech32 failed: invalid separator index -1`,
		},
		{
			name:  "invalid address: with field",
			a:     AccessGrant{Address: "invalid_address_____", Permissions: []Permission{Permission_settle}},
			field: "meadow",
			exp:   `invalid meadow access grant: invalid address "invalid_address_____": decoding bech32 failed: invalid separator index -1`,
		},
		{
			name:  "nil permissions: no field",
			a:     AccessGrant{Address: addr, Permissions: nil},
			field: "",
			exp:   "invalid access grant: no permissions provided for " + addr,
		},
		{
			name:  "nil permissions: with field",
			a:     AccessGrant{Address: addr, Permissions: nil},
			field: "meadow",
			exp:   "invalid meadow access grant: no permissions provided for " + addr,
		},
		{
			name:  "empty permissions: no field",
			a:     AccessGrant{Address: addr, Permissions: []Permission{}},
			field: "",
			exp:   "invalid access grant: no permissions provided for " + addr,
		},
		{
			name:  "empty permissions: with field",
			a:     AccessGrant{Address: addr, Permissions: []Permission{}},
			field: "meadow",
			exp:   "invalid meadow access grant: no permissions provided for " + addr,
		},
		{
			name: "duplicate entry: no field",
			a: AccessGrant{
				Address: addr,
				Permissions: []Permission{
					Permission_settle,
					Permission_cancel,
					Permission_settle,
				},
			},
			field: "",
			exp:   "invalid access grant: settle appears multiple times for " + addr,
		},
		{
			name: "duplicate entry: with field",
			a: AccessGrant{
				Address: addr,
				Permissions: []Permission{
					Permission_settle,
					Permission_cancel,
					Permission_settle,
				},
			},
			field: "meadow",
			exp:   "invalid meadow access grant: settle appears multiple times for " + addr,
		},
		{
			name: "invalid entry: no field",
			a: AccessGrant{
				Address: addr,
				Permissions: []Permission{
					Permission_withdraw,
					-1,
					Permission_attributes,
				},
			},
			field: "",
			exp:   "invalid access grant: permission -1 does not exist for " + addr,
		},
		{
			name: "invalid entry: with field",
			a: AccessGrant{
				Address: addr,
				Permissions: []Permission{
					Permission_withdraw,
					-1,
					Permission_attributes,
				},
			},
			field: "meadow",
			exp:   "invalid meadow access grant: permission -1 does not exist for " + addr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.a.ValidateInField(tc.field)

			assertions.AssertErrorValue(t, err, tc.exp, "ValidateInField(%q)", tc.field)
		})
	}
}

func TestAccessGrant_Contains(t *testing.T) {
	tests := []struct {
		name string
		a    AccessGrant
		perm Permission
		exp  bool
	}{
		{
			name: "nil permissions",
			a:    AccessGrant{Permissions: nil},
			perm: 0,
			exp:  false,
		},
		{
			name: "empty permissions",
			a:    AccessGrant{Permissions: []Permission{}},
			perm: 0,
			exp:  false,
		},
		{
			name: "all permissions: checking unspecified",
			a:    AccessGrant{Permissions: AllPermissions()},
			perm: Permission_unspecified,
			exp:  false,
		},
		{
			name: "all permissions: checking settle",
			a:    AccessGrant{Permissions: AllPermissions()},
			perm: Permission_settle,
			exp:  true,
		},
		{
			name: "all permissions: checking cancel",
			a:    AccessGrant{Permissions: AllPermissions()},
			perm: Permission_cancel,
			exp:  true,
		},
		{
			name: "all permissions: checking withdraw",
			a:    AccessGrant{Permissions: AllPermissions()},
			perm: Permission_withdraw,
			exp:  true,
		},
		{
			name: "all permissions: checking update",
			a:    AccessGrant{Permissions: AllPermissions()},
			perm: Permission_update,
			exp:  true,
		},
		{
			name: "all permissions: checking permissions",
			a:    AccessGrant{Permissions: AllPermissions()},
			perm: Permission_permissions,
			exp:  true,
		},
		{
			name: "all permissions: checking attributes",
			a:    AccessGrant{Permissions: AllPermissions()},
			perm: Permission_attributes,
			exp:  true,
		},
		{
			name: "all permissions: checking unknown",
			a:    AccessGrant{Permissions: AllPermissions()},
			perm: 99,
			exp:  false,
		},
		{
			name: "all permissions: checking negative",
			a:    AccessGrant{Permissions: AllPermissions()},
			perm: -22,
			exp:  false,
		},
		{
			name: "only settle: checking unspecified",
			a:    AccessGrant{Permissions: []Permission{Permission_settle}},
			perm: Permission_unspecified,
			exp:  false,
		},
		{
			name: "only settle: checking settle",
			a:    AccessGrant{Permissions: []Permission{Permission_settle}},
			perm: Permission_settle,
			exp:  true,
		},
		{
			name: "only settle: checking cancel",
			a:    AccessGrant{Permissions: []Permission{Permission_settle}},
			perm: Permission_cancel,
			exp:  false,
		},
		{
			name: "only settle: checking withdraw",
			a:    AccessGrant{Permissions: []Permission{Permission_settle}},
			perm: Permission_withdraw,
			exp:  false,
		},
		{
			name: "only settle: checking update",
			a:    AccessGrant{Permissions: []Permission{Permission_settle}},
			perm: Permission_update,
			exp:  false,
		},
		{
			name: "only settle: checking permissions",
			a:    AccessGrant{Permissions: []Permission{Permission_settle}},
			perm: Permission_permissions,
			exp:  false,
		},
		{
			name: "only settle: checking attributes",
			a:    AccessGrant{Permissions: []Permission{Permission_settle}},
			perm: Permission_attributes,
			exp:  false,
		},
		{
			name: "only settle: checking unknown",
			a:    AccessGrant{Permissions: []Permission{Permission_settle}},
			perm: 99,
			exp:  false,
		},
		{
			name: "only settle: checking negative",
			a:    AccessGrant{Permissions: []Permission{Permission_settle}},
			perm: -22,
			exp:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual bool
			testFunc := func() {
				actual = tc.a.Contains(tc.perm)
			}
			require.NotPanics(t, testFunc, "%#v.Contains(%s)", tc.a, tc.perm.SimpleString())
			assert.Equal(t, tc.exp, actual, "%#v.Contains(%s)", tc.a, tc.perm.SimpleString())
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
			name: "set_ids",
			p:    Permission_set_ids,
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

			assertions.AssertErrorValue(t, err, tc.exp, "Validate()")
		})
	}

	t.Run("all values have a test case", func(t *testing.T) {
		allVals := helpers.Keys(Permission_name)
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
		Permission_set_ids,
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

		// Permission_set_ids
		{permission: "set_ids", expected: Permission_set_ids},
		{permission: " set_ids", expected: Permission_set_ids},
		{permission: "set_ids ", expected: Permission_set_ids},
		{permission: "SET_IDS", expected: Permission_set_ids},
		{permission: "Set_Ids", expected: Permission_set_ids},
		{permission: "setids", expected: Permission_set_ids},
		{permission: "setids ", expected: Permission_set_ids},
		{permission: " setids", expected: Permission_set_ids},
		{permission: "setIds", expected: Permission_set_ids},
		{permission: "SetIds", expected: Permission_set_ids},
		{permission: "SETIDS", expected: Permission_set_ids},
		{permission: "permission_set_ids", expected: Permission_set_ids},
		{permission: "PERMISSION_SET_IDS", expected: Permission_set_ids},
		{permission: "peRMissiOn_sEt_iDs", expected: Permission_set_ids},
		{permission: "permission_setids", expected: Permission_set_ids},
		{permission: "PERMISSION_SETIDS", expected: Permission_set_ids},
		{permission: "peRMissiOn_sEtiDs", expected: Permission_set_ids},

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
		allVals := helpers.Keys(Permission_name)
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

			assertions.AssertErrorValue(t, err, tc.expErr, "ParsePermissions(%q) error", tc.permissions)
			assert.Equal(t, tc.expected, perms, "ParsePermissions(%q) result", tc.permissions)
		})
	}
}

func TestNormalizeReqAttrs(t *testing.T) {
	tests := []struct {
		name     string
		reqAttrs []string
		expAttrs []string
		expErr   string
	}{
		{
			name:     "whitespace fixed",
			reqAttrs: []string{"  ab  .  cd  .  ef    ", "cd . ef", " * . jk "},
			expAttrs: []string{"ab.cd.ef", "cd.ef", "*.jk"},
		},
		{
			name:     "casing fixed",
			reqAttrs: []string{"AB.cD.Ef", "*.PQ"},
			expAttrs: []string{"ab.cd.ef", "*.pq"},
		},
		{
			name:     "some problems",
			reqAttrs: []string{"ab.c d.ef", "X.*.Y", " a-B-c .d"},
			expAttrs: []string{"ab.c d.ef", "x.*.y", "a-b-c.d"},
			expErr: `invalid attribute "ab.c d.ef"` + "\n" +
				`invalid attribute "X.*.Y"` + "\n" +
				`invalid attribute " a-B-c .d"`,
		},
		{
			name:     "two good one bad",
			reqAttrs: []string{" AB .cd", "l,M.n.o,p", " *.x.Y.z"},
			expAttrs: []string{"ab.cd", "l,m.n.o,p", "*.x.y.z"},
			expErr:   `invalid attribute "l,M.n.o,p"`,
		},
		{
			// Unlike ValidateReqAttrs, this one doesn't care about dups or duplicated errors.
			name:     "duplicated entries",
			reqAttrs: []string{"*.x.y.z", "*.x.y.z", "a.b.*.d", "a.b.*.d"},
			expAttrs: []string{"*.x.y.z", "*.x.y.z", "a.b.*.d", "a.b.*.d"},
			expErr: `invalid attribute "a.b.*.d"` + "\n" +
				`invalid attribute "a.b.*.d"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var attrs []string
			var err error
			testFunc := func() {
				attrs, err = NormalizeReqAttrs(tc.reqAttrs)
			}
			require.NotPanics(t, testFunc, "NormalizeReqAttrs(%q)", tc.reqAttrs)
			assertions.AssertErrorValue(t, err, tc.expErr, "NormalizeReqAttrs(%q) error", tc.reqAttrs)
			assert.Equal(t, tc.expAttrs, attrs, "NormalizeReqAttrs(%q) result", tc.reqAttrs)
		})
	}
}

func TestValidateReqAttrsAreNormalized(t *testing.T) {
	notNormErr := func(field, attr, norm string) string {
		return fmt.Sprintf("%s required attribute %q is not normalized, expected %q", field, attr, norm)
	}

	tests := []struct {
		name   string
		field  string
		attrs  []string
		expErr string
	}{
		{name: "nil attrs", field: "FOILD", attrs: nil, expErr: ""},
		{name: "empty attrs", field: "FOILD", attrs: []string{}, expErr: ""},
		{
			name:   "one attr: normalized",
			field:  "TINFOILD",
			attrs:  []string{"abc.def"},
			expErr: "",
		},
		{
			name:   "one attr: with whitespace",
			field:  "AlFOILD",
			attrs:  []string{" abc.def"},
			expErr: notNormErr("AlFOILD", " abc.def", "abc.def"),
		},
		{
			name:   "one attr: with upper",
			field:  "AlFOILD",
			attrs:  []string{"aBc.def"},
			expErr: notNormErr("AlFOILD", "aBc.def", "abc.def"),
		},
		{
			name:   "one attr: with wildcard, ok",
			field:  "NOFOILD",
			attrs:  []string{"*.abc.def"},
			expErr: "",
		},
		{
			name:   "one attr: with wildcard, bad",
			field:  "AirFOILD",
			attrs:  []string{"*.abc. def"},
			expErr: notNormErr("AirFOILD", "*.abc. def", "*.abc.def"),
		},
		{
			name:   "three attrs: all okay",
			field:  "WhaFoild",
			attrs:  []string{"abc.def", "*.ghi.jkl", "mno.pqr.stu.vwx.yz"},
			expErr: "",
		},
		{
			name:   "three attrs: bad first",
			field:  "Uno1Foild",
			attrs:  []string{"abc. def", "*.ghi.jkl", "mno.pqr.stu.vwx.yz"},
			expErr: notNormErr("Uno1Foild", "abc. def", "abc.def"),
		},
		{
			name:   "three attrs: bad second",
			field:  "Uno2Foild",
			attrs:  []string{"abc.def", "*.ghi.jkl ", "mno.pqr.stu.vwx.yz"},
			expErr: notNormErr("Uno2Foild", "*.ghi.jkl ", "*.ghi.jkl"),
		},
		{
			name:   "three attrs: bad third",
			field:  "Uno3Foild",
			attrs:  []string{"abc.def", "*.ghi.jkl", "mnO.pqr.stu.vwX.yz"},
			expErr: notNormErr("Uno3Foild", "mnO.pqr.stu.vwX.yz", "mno.pqr.stu.vwx.yz"),
		},
		{
			name:  "three attrs: bad first and second",
			field: "TwoFold1",
			attrs: []string{"abc.Def", "* .ghi.jkl", "mno.pqr.stu.vwx.yz"},
			expErr: joinErrs(
				notNormErr("TwoFold1", "abc.Def", "abc.def"),
				notNormErr("TwoFold1", "* .ghi.jkl", "*.ghi.jkl"),
			),
		},
		{
			name:  "three attrs: bad first and third",
			field: "TwoFold2",
			attrs: []string{"abc . def", "*.ghi.jkl", "mno.pqr. stu .vwx.yz"},
			expErr: joinErrs(
				notNormErr("TwoFold2", "abc . def", "abc.def"),
				notNormErr("TwoFold2", "mno.pqr. stu .vwx.yz", "mno.pqr.stu.vwx.yz"),
			),
		},
		{
			name:  "three attrs: bad second and third",
			field: "TwoFold3",
			attrs: []string{"abc.def", "*.ghi.JKl", "mno.pqr.sTu.vwx.yz"},
			expErr: joinErrs(
				notNormErr("TwoFold3", "*.ghi.JKl", "*.ghi.jkl"),
				notNormErr("TwoFold3", "mno.pqr.sTu.vwx.yz", "mno.pqr.stu.vwx.yz"),
			),
		},
		{
			name:  "three attrs: all bad",
			field: "CURSES!",
			attrs: []string{" abc . def ", " * . ghi . jkl ", " mno . pqr . stu . vwx . yz "},
			expErr: joinErrs(
				notNormErr("CURSES!", " abc . def ", "abc.def"),
				notNormErr("CURSES!", " * . ghi . jkl ", "*.ghi.jkl"),
				notNormErr("CURSES!", " mno . pqr . stu . vwx . yz ", "mno.pqr.stu.vwx.yz"),
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = ValidateReqAttrsAreNormalized(tc.field, tc.attrs)
			}
			require.NotPanics(t, testFunc, "ValidateReqAttrsAreNormalized")
			assertions.AssertErrorValue(t, err, tc.expErr, "ValidateReqAttrsAreNormalized error")
		})
	}
}

func TestValidateReqAttrs(t *testing.T) {
	tests := []struct {
		name  string
		field string
		attrs []string
		exp   string
	}{
		{
			name:  "nil list",
			field: "FEYULD",
			attrs: nil,
			exp:   "",
		},
		{
			name:  "empty list",
			field: "FEYULD",
			attrs: []string{},
			exp:   "",
		},
		{
			name:  "three valid entries: normalized",
			field: "FEYULD",
			attrs: []string{"*.wildcard", "penny.nickel.dime", "*.example.pb"},
			exp:   "",
		},
		{
			name:  "three valid entries: not normalized",
			field: "FEYULD",
			attrs: []string{" * . wildcard ", " penny  . nickel .   dime ", " * . example . pb        "},
			exp:   "",
		},
		{
			name:  "three entries: first invalid",
			field: "FEYULD",
			attrs: []string{"x.*.wildcard", "penny.nickel.dime", "*.example.pb"},
			exp:   `invalid FEYULD required attribute "x.*.wildcard"`,
		},
		{
			name:  "three entries: second invalid",
			field: "fee-yield",
			attrs: []string{"*.wildcard", "penny.nic kel.dime", "*.example.pb"},
			exp:   `invalid fee-yield required attribute "penny.nic kel.dime"`,
		},
		{
			name:  "three entries: third invalid",
			field: "fee-yelled",
			attrs: []string{"*.wildcard", "penny.nickel.dime", "*.ex-am-ple.pb"},
			exp:   `invalid fee-yelled required attribute "*.ex-am-ple.pb"`,
		},
		{
			name:  "duplicate entries",
			field: "just some field name thingy",
			attrs: []string{"*.multi", "*.multi", "*.multi"},
			exp:   `duplicate just some field name thingy required attribute "*.multi"`,
		},
		{
			name:  "duplicate bad entries",
			field: "bananas",
			attrs: []string{"bad.*.example", "bad. * .example"},
			exp:   `invalid bananas required attribute "bad.*.example"`,
		},
		{
			name:  "multiple problems",
			field: "♪ but a bit ain't one ♪",
			attrs: []string{
				"one.multi", "x.*.wildcard", "x.*.wildcard", "one.multi", "two.multi",
				"penny.nic kel.dime", "one.multi", "two.multi", "*.ex-am-ple.pb", "two.multi",
			},
			exp: joinErrs(
				`invalid ♪ but a bit ain't one ♪ required attribute "x.*.wildcard"`,
				`duplicate ♪ but a bit ain't one ♪ required attribute "one.multi"`,
				`invalid ♪ but a bit ain't one ♪ required attribute "penny.nic kel.dime"`,
				`duplicate ♪ but a bit ain't one ♪ required attribute "two.multi"`,
				`invalid ♪ but a bit ain't one ♪ required attribute "*.ex-am-ple.pb"`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateReqAttrs(tc.field, tc.attrs)
			assertions.AssertErrorValue(t, err, tc.exp, "ValidateReqAttrs")
		})
	}
}

func TestIntersectionOfAttributes(t *testing.T) {
	tests := []struct {
		name     string
		options1 []string
		options2 []string
		expected []string
	}{
		{name: "nil lists", options1: nil, options2: nil, expected: nil},
		{name: "empty lists", options1: []string{}, options2: []string{}, expected: nil},
		{name: "one item to nil", options1: []string{"one"}, options2: nil, expected: nil},
		{name: "nil to one item", options1: nil, options2: []string{"one"}, expected: nil},
		{name: "one to one: different", options1: []string{"one"}, options2: []string{"two"}, expected: nil},
		{name: "one to one: same", options1: []string{"one"}, options2: []string{"one"}, expected: []string{"one"}},
		{
			name:     "one to three: none common",
			options1: []string{"one"},
			options2: []string{"two", "three", "four"},
			expected: nil,
		},
		{
			name:     "one to three: first",
			options1: []string{"one"},
			options2: []string{"one", "three", "four"},
			expected: []string{"one"},
		},
		{
			name:     "one to three: second",
			options1: []string{"one"},
			options2: []string{"two", "one", "four"},
			expected: []string{"one"},
		},
		{
			name:     "one to three: third",
			options1: []string{"one"},
			options2: []string{"two", "three", "one"},
			expected: []string{"one"},
		},
		{
			name:     "one to three: all equal",
			options1: []string{"one"},
			options2: []string{"one", "one", "one"},
			expected: []string{"one"},
		},
		{
			name:     "three to one: none common",
			options1: []string{"one", "two", "three"},
			options2: []string{"four"},
			expected: nil,
		},
		{
			name:     "three to one: first",
			options1: []string{"one", "two", "three"},
			options2: []string{"one"},
			expected: []string{"one"},
		},
		{
			name:     "three to one: second",
			options1: []string{"one", "two", "three"},
			options2: []string{"two"},
			expected: []string{"two"},
		},
		{
			name:     "three to one: third",
			options1: []string{"one", "two", "three"},
			options2: []string{"three"},
			expected: []string{"three"},
		},
		{
			name:     "three to one: all equal",
			options1: []string{"one", "one", "one"},
			options2: []string{"one"},
			expected: []string{"one"},
		},
		{
			name:     "three to three: none common",
			options1: []string{"one", "two", "three"},
			options2: []string{"four", "five", "six"},
			expected: nil,
		},
		{
			name:     "three to three: one common: first to first",
			options1: []string{"one", "two", "three"},
			options2: []string{"one", "five", "six"},
			expected: []string{"one"},
		},
		{
			name:     "three to three: one common: first to second",
			options1: []string{"one", "two", "three"},
			options2: []string{"four", "one", "six"},
			expected: []string{"one"},
		},
		{
			name:     "three to three: one common: first to third",
			options1: []string{"one", "two", "three"},
			options2: []string{"four", "five", "one"},
			expected: []string{"one"},
		},
		{
			name:     "three to three: one common: second to first",
			options1: []string{"one", "two", "three"},
			options2: []string{"two", "five", "six"},
			expected: []string{"two"},
		},
		{
			name:     "three to three: one common: second to second",
			options1: []string{"one", "two", "three"},
			options2: []string{"four", "two", "six"},
			expected: []string{"two"},
		},
		{
			name:     "three to three: one common: second to third",
			options1: []string{"one", "two", "three"},
			options2: []string{"four", "five", "two"},
			expected: []string{"two"},
		},
		{
			name:     "three to three: one common: third to first",
			options1: []string{"one", "two", "three"},
			options2: []string{"three", "five", "six"},
			expected: []string{"three"},
		},
		{
			name:     "three to three: one common: third to second",
			options1: []string{"one", "two", "three"},
			options2: []string{"four", "three", "six"},
			expected: []string{"three"},
		},
		{
			name:     "three to three: one common: third to third",
			options1: []string{"one", "two", "three"},
			options2: []string{"four", "five", "three"},
			expected: []string{"three"},
		},
		{
			name:     "three to three: two common",
			options1: []string{"one", "two", "three"},
			options2: []string{"two", "five", "one"},
			expected: []string{"one", "two"},
		},
		{
			name:     "three to three: same lists different order",
			options1: []string{"one", "two", "three"},
			options2: []string{"two", "three", "one"},
			expected: []string{"one", "two", "three"},
		},
		{
			name:     "three to three: all equal",
			options1: []string{"one", "one", "one"},
			options2: []string{"one", "one", "one"},
			expected: []string{"one"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []string
			testFunc := func() {
				actual = IntersectionOfAttributes(tc.options1, tc.options2)
			}
			require.NotPanics(t, testFunc, "IntersectionOfAttributes")
			assert.Equal(t, tc.expected, actual, "IntersectionOfAttributes result")
		})
	}
}

func TestValidateAddRemoveReqAttrs(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		toAdd    []string
		toRemove []string
		expErr   string
	}{
		{
			name:     "nil lists",
			field:    "-<({[FIELD]})>-",
			toAdd:    nil,
			toRemove: nil,
		},
		{
			name:     "empty lists",
			field:    "-<({[FIELD]})>-",
			toAdd:    []string{},
			toRemove: []string{},
		},
		{
			name:     "invalid to-add entry",
			field:    "-<({[FIELD]})>-",
			toAdd:    []string{"this has too many spaces"},
			toRemove: nil,
			expErr:   "invalid -<({[FIELD]})>- to add required attribute \"this has too many spaces\"",
		},
		{
			name:     "invalid to-remove entry",
			field:    "-<({[FIELD]})>-",
			toAdd:    nil,
			toRemove: []string{"this has too many spaces"},
		},
		{
			name:     "duplicate to-add entry",
			field:    "-<({[FIELD]})>-",
			toAdd:    []string{"abc", "def", "abc", "abc"},
			toRemove: nil,
			expErr:   "duplicate -<({[FIELD]})>- to add required attribute \"abc\"",
		},
		{
			name:     "same entry in both lists",
			field:    "-<({[FIELD]})>-",
			toAdd:    []string{"one", "two", "three"},
			toRemove: []string{"four", "five", "six", "two", "seven"},
			expErr:   "cannot add and remove the same -<({[FIELD]})>- required attributes \"two\"",
		},
		{
			name:     "two entries in both lists differently cased",
			field:    "-<({[FIELD]})>-",
			toAdd:    []string{"one", "SEvEN", "two", "three"},
			toRemove: []string{"four", "five", "six", "Two", "seven"},
			expErr:   "cannot add and remove the same -<({[FIELD]})>- required attributes \"SEvEN\",\"two\"",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = ValidateAddRemoveReqAttrs(tc.field, tc.toAdd, tc.toRemove)
			}
			require.NotPanics(t, testFunc, "ValidateAddRemoveReqAttrs")
			assertions.AssertErrorValue(t, err, tc.expErr, "ValidateAddRemoveReqAttrs error")
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
		{name: "just the wildcard", reqAttr: "*", exp: false},
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
			name:    "with wildcard: just base",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "penny.dime.quarter",
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
			name:    "just a star",
			reqAttr: "*",
			accAttr: "penny.dime.quarter",
			exp:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isMatch := IsReqAttrMatch(tc.reqAttr, tc.accAttr)
			assert.Equal(t, tc.exp, isMatch, "IsReqAttrMatch(%q, %q)", tc.reqAttr, tc.accAttr)
		})
	}
}
