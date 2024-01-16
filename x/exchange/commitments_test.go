package exchange

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/testutil/assertions"
)

func TestAccountAmount_String(t *testing.T) {
	tests := []struct {
		name string
		val  AccountAmount
		exp  string
	}{
		{
			name: "empty",
			val:  AccountAmount{},
			exp:  `:""`,
		},
		{
			name: "only account",
			val:  AccountAmount{Account: "acct"},
			exp:  `acct:""`,
		},
		{
			name: "only amount",
			val:  AccountAmount{Amount: sdk.NewCoins(sdk.NewInt64Coin("okay", 123))},
			exp:  `:"123okay"`,
		},
		{
			name: "both account and amount",
			val: AccountAmount{
				Account: "justsomeaccount",
				Amount:  sdk.NewCoins(sdk.NewInt64Coin("apple", 72), sdk.NewInt64Coin("banana", 41)),
			},
			exp: `justsomeaccount:"72apple,41banana"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act string
			testFunc := func() {
				act = tc.val.String()
			}
			require.NotPanics(t, testFunc, "%#v.String()", tc.val)
			assert.Equal(t, tc.exp, act, "String() result")
		})
	}
}

func TestAccountAmount_Validate(t *testing.T) {
	tests := []struct {
		name string
		val  AccountAmount
		exp  string
	}{
		{
			name: "okay",
			val: AccountAmount{
				Account: sdk.AccAddress("account_____________").String(),
				Amount:  sdk.NewCoins(sdk.NewInt64Coin("apple", 12)),
			},
			exp: "",
		},
		{
			name: "bad account",
			val: AccountAmount{
				Account: "notanaccount",
				Amount:  sdk.NewCoins(sdk.NewInt64Coin("apple", 12)),
			},
			exp: "invalid account \"notanaccount\": decoding bech32 failed: invalid separator index -1",
		},
		{
			name: "bad amount denom",
			val: AccountAmount{
				Account: sdk.AccAddress("account_____________").String(),
				Amount:  sdk.Coins{sdk.Coin{Denom: "x", Amount: sdk.NewInt(12)}},
			},
			exp: "invalid amount \"12x\": invalid denom: x",
		},
		{
			name: "negative amount",
			val: AccountAmount{
				Account: sdk.AccAddress("account_____________").String(),
				Amount:  sdk.Coins{sdk.Coin{Denom: "negcoin", Amount: sdk.NewInt(-3)}},
			},
			exp: "invalid amount \"-3negcoin\": coin -3negcoin amount is not positive",
		},
		{
			name: "no amount",
			val: AccountAmount{
				Account: sdk.AccAddress("account_____________").String(),
				Amount:  nil,
			},
		},
		{
			name: "zero coin in amount",
			val: AccountAmount{
				Account: sdk.AccAddress("account_____________").String(),
				Amount:  sdk.NewCoins(sdk.Coin{Denom: "zcoin", Amount: sdk.NewInt(0)}),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.val.Validate()
			}
			require.NotPanics(t, testFunc, "%#v.Validate()", tc.val)
			assertions.AssertErrorValue(t, err, tc.exp, "Validate() result")
		})
	}
}

func TestSumAccountAmounts(t *testing.T) {
	coins := func(val string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(val)
		require.NoError(t, err, "ParseCoinsNormalized(%q)", val)
		return rv
	}

	tests := []struct {
		name    string
		entries []AccountAmount
		exp     sdk.Coins
	}{
		{
			name:    "nil entries",
			entries: nil,
			exp:     nil,
		},
		{
			name:    "empty entries",
			entries: []AccountAmount{},
			exp:     nil,
		},
		{
			name:    "one entry",
			entries: []AccountAmount{{Amount: coins("10banana")}},
			exp:     coins("10banana"),
		},
		{
			name: "three entries",
			entries: []AccountAmount{
				{Amount: coins("15apple,51prune")},
				{Amount: coins("2apple,8banana")},
				{Amount: coins("3banana,6cherry")},
			},
			exp: coins("17apple,11banana,6cherry,51prune"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act sdk.Coins
			testFunc := func() {
				act = SumAccountAmounts(tc.entries)
			}
			require.NotPanics(t, testFunc, "SumAccountAmounts")
			assert.Equal(t, tc.exp.String(), act.String(), "SumAccountAmounts result")
		})
	}
}

func TestMarketAmount_String(t *testing.T) {
	tests := []struct {
		name string
		val  MarketAmount
		exp  string
	}{
		{
			name: "empty",
			val:  MarketAmount{},
			exp:  `0:""`,
		},
		{
			name: "only market",
			val:  MarketAmount{MarketId: 8},
			exp:  `8:""`,
		},
		{
			name: "only amount",
			val:  MarketAmount{Amount: sdk.NewCoins(sdk.NewInt64Coin("okay", 123))},
			exp:  `0:"123okay"`,
		},
		{
			name: "both market and amount",
			val: MarketAmount{
				MarketId: 985412,
				Amount:   sdk.NewCoins(sdk.NewInt64Coin("apple", 72), sdk.NewInt64Coin("banana", 41)),
			},
			exp: `985412:"72apple,41banana"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act string
			testFunc := func() {
				act = tc.val.String()
			}
			require.NotPanics(t, testFunc, "%#v.String()", tc.val)
			assert.Equal(t, tc.exp, act, "String() result")
		})
	}
}

func TestNetAssetPrice_String(t *testing.T) {
	tests := []struct {
		name string
		nav  NetAssetPrice
		exp  string
	}{
		{
			name: "empty",
			nav:  NetAssetPrice{},
			exp:  `"<nil>"="<nil>"`,
		},
		{
			name: "only assets",
			nav:  NetAssetPrice{Assets: sdk.NewInt64Coin("apple", 3)},
			exp:  `"3apple"="<nil>"`,
		},
		{
			name: "only price",
			nav:  NetAssetPrice{Price: sdk.NewInt64Coin("plum", 14)},
			exp:  `"<nil>"="14plum"`,
		},
		{
			name: "both assets and price",
			nav: NetAssetPrice{
				Assets: sdk.NewInt64Coin("apple", 42),
				Price:  sdk.NewInt64Coin("plum", 88),
			},
			exp: `"42apple"="88plum"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act string
			testFunc := func() {
				act = tc.nav.String()
			}
			require.NotPanics(t, testFunc, "%#v.String()", tc.nav)
			assert.Equal(t, tc.exp, act, "String() result")
		})
	}
}

func TestNetAssetPrice_Validate(t *testing.T) {
	tests := []struct {
		name string
		nav  NetAssetPrice
		exp  string
	}{
		{
			name: "okay",
			nav: NetAssetPrice{
				Assets: sdk.NewInt64Coin("apple", 15),
				Price:  sdk.NewInt64Coin("plum", 44),
			},
		},
		{
			name: "bad assets denom",
			nav: NetAssetPrice{
				Assets: sdk.Coin{Denom: "x", Amount: sdk.NewInt(16)},
				Price:  sdk.NewInt64Coin("plum", 44),
			},
			exp: "invalid assets \"16x\": invalid denom: x",
		},
		{
			name: "negative assets",
			nav: NetAssetPrice{
				Assets: sdk.Coin{Denom: "apple", Amount: sdk.NewInt(-12)},
				Price:  sdk.NewInt64Coin("plum", 44),
			},
			exp: "invalid assets \"-12apple\": negative coin amount: -12",
		},
		{
			name: "zero assets",
			nav: NetAssetPrice{
				Assets: sdk.Coin{Denom: "apple", Amount: sdk.NewInt(0)},
				Price:  sdk.NewInt64Coin("plum", 44),
			},
			exp: "invalid assets \"0apple\": cannot be zero",
		},
		{
			name: "bad price denom",
			nav: NetAssetPrice{
				Assets: sdk.NewInt64Coin("apple", 15),
				Price:  sdk.Coin{Denom: "y", Amount: sdk.NewInt(16)},
			},
			exp: "invalid price \"16y\": invalid denom: y",
		},
		{
			name: "negative price",
			nav: NetAssetPrice{
				Assets: sdk.NewInt64Coin("apple", 15),
				Price:  sdk.Coin{Denom: "plum", Amount: sdk.NewInt(-12)},
			},
			exp: "invalid price \"-12plum\": negative coin amount: -12",
		},
		{
			name: "zero price",
			nav: NetAssetPrice{
				Assets: sdk.NewInt64Coin("apple", 15),
				Price:  sdk.Coin{Denom: "plum", Amount: sdk.NewInt(0)},
			},
			exp: "invalid price \"0plum\": cannot be zero",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.nav.Validate()
			}
			require.NotPanics(t, testFunc, "%#v.Validate()", tc.nav)
			assertions.AssertErrorValue(t, err, tc.exp, "Validate() result")
		})
	}
}

func TestValidateEventTag(t *testing.T) {
	tests := []struct {
		name     string
		eventTag string
		exp      string
	}{
		{
			name:     "empty string",
			eventTag: "",
			exp:      "",
		},
		{
			name:     "one char",
			eventTag: "a",
			exp:      "",
		},
		{
			name:     "a uuid",
			eventTag: "c356a71d-7d5f-46dd-bf48-74283377ec31",
			exp:      "",
		},
		{
			name:     "max length",
			eventTag: strings.Repeat("p", MaxEventTagLength),
			exp:      "",
		},
		{
			name:     "max length plus one",
			eventTag: strings.Repeat("p", MaxEventTagLength) + "R",
			exp: fmt.Sprintf("invalid event tag %q (length %d): exceeds max length %d",
				"ppppp...ppppR", MaxEventTagLength+1, MaxEventTagLength),
		},
		{
			name:     "really long",
			eventTag: "a" + strings.Repeat("pP", 4_999) + "z",
			exp: fmt.Sprintf("invalid event tag %q (length %d): exceeds max length %d",
				"apPpP...pPpPz", 10_000, MaxEventTagLength),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = ValidateEventTag(tc.eventTag)
			}
			require.NotPanics(t, testFunc, "ValidateEventTag(%q)", tc.eventTag)
			assertions.AssertErrorValue(t, err, tc.exp, "ValidateEventTag(%q) result")
		})
	}
}
