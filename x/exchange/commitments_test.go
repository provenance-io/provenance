package exchange

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/provenance-io/provenance/testutil/assertions"
)

func TestCommitment_Validate(t *testing.T) {
	tests := []struct {
		name       string
		commitment Commitment
		exp        string
	}{
		{
			name: "bad account",
			commitment: Commitment{
				Account:  "badaccount",
				MarketId: 1,
				Amount:   sdk.NewCoins(sdk.NewInt64Coin("nhash", 1000)),
			},
			exp: "invalid account \"badaccount\": decoding bech32 failed: invalid separator index -1",
		},
		{
			name: "bad market",
			commitment: Commitment{
				Account:  sdk.AccAddress("account_____________").String(),
				MarketId: 0,
				Amount:   sdk.NewCoins(sdk.NewInt64Coin("nhash", 1000)),
			},
			exp: "invalid market id: cannot be zero",
		},
		{
			name: "bad amount denom",
			commitment: Commitment{
				Account:  sdk.AccAddress("account_____________").String(),
				MarketId: 1,
				Amount:   sdk.Coins{sdk.Coin{Denom: "p", Amount: sdk.NewInt(5)}},
			},
			exp: "invalid amount \"5p\": invalid denom: p",
		},
		{
			name: "negative amount",
			commitment: Commitment{
				Account:  sdk.AccAddress("account_____________").String(),
				MarketId: 1,
				Amount:   sdk.Coins{sdk.Coin{Denom: "plum", Amount: sdk.NewInt(-5)}},
			},
			exp: "invalid amount \"-5plum\": coin -5plum amount is not positive",
		},
		{
			name: "okay",
			commitment: Commitment{
				Account:  sdk.AccAddress("account_____________").String(),
				MarketId: 1,
				Amount:   sdk.NewCoins(sdk.NewInt64Coin("cherry", 5000)),
			},
			exp: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.commitment.Validate()
			}
			require.NotPanics(t, testFunc, "commitment.Validate()")
			assertions.AssertErrorValue(t, err, tc.exp, "commitment.Validate() result")
		})
	}
}

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

func TestAccountAmount_ValidateWithOptionalAmount(t *testing.T) {
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
				Amount:  sdk.Coins{sdk.Coin{Denom: "zcoin", Amount: sdk.NewInt(0)}},
			},
			exp: "invalid amount \"0zcoin\": coin 0zcoin amount is not positive",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.val.ValidateWithOptionalAmount()
			}
			require.NotPanics(t, testFunc, "%#v.ValidateWithOptionalAmount()", tc.val)
			assertions.AssertErrorValue(t, err, tc.exp, "ValidateWithOptionalAmount() result")
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
			exp: "invalid amount \"\": cannot be zero",
		},
		{
			name: "zero coin in amount",
			val: AccountAmount{
				Account: sdk.AccAddress("account_____________").String(),
				Amount:  sdk.Coins{sdk.Coin{Denom: "zcoin", Amount: sdk.NewInt(0)}},
			},
			exp: "invalid amount \"0zcoin\": coin 0zcoin amount is not positive",
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

func TestSimplifyAccountAmounts(t *testing.T) {
	coins := func(val string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(val)
		require.NoError(t, err, "ParseCoinsNormalized(%q)", val)
		return rv
	}

	tests := []struct {
		name     string
		entries  []AccountAmount
		expected []AccountAmount
	}{
		{
			name:     "nil",
			entries:  nil,
			expected: nil,
		},
		{
			name:     "empty",
			entries:  []AccountAmount{},
			expected: []AccountAmount{},
		},
		{
			name:     "one entry",
			entries:  []AccountAmount{{Account: "one", Amount: coins("1one")}},
			expected: []AccountAmount{{Account: "one", Amount: coins("1one")}},
		},
		{
			name: "two entries: diff addrs",
			entries: []AccountAmount{
				{Account: "addr1", Amount: coins("1onecoin")},
				{Account: "addr2", Amount: coins("2twocoin")},
			},
			expected: []AccountAmount{
				{Account: "addr1", Amount: coins("1onecoin")},
				{Account: "addr2", Amount: coins("2twocoin")},
			},
		},
		{
			name: "two entries: same addrs",
			entries: []AccountAmount{
				{Account: "addr1", Amount: coins("1onecoin")},
				{Account: "addr1", Amount: coins("2twocoin")},
			},
			expected: []AccountAmount{
				{Account: "addr1", Amount: coins("1onecoin,2twocoin")},
			},
		},
		{
			name: "three entries: all diff addrs",
			entries: []AccountAmount{
				{Account: "addr1", Amount: coins("1apple")},
				{Account: "addr2", Amount: coins("2apple")},
				{Account: "addr3", Amount: coins("3apple")},
			},
			expected: []AccountAmount{
				{Account: "addr1", Amount: coins("1apple")},
				{Account: "addr2", Amount: coins("2apple")},
				{Account: "addr3", Amount: coins("3apple")},
			},
		},
		{
			name: "three entries: first second same addr",
			entries: []AccountAmount{
				{Account: "addr1", Amount: coins("1apple")},
				{Account: "addr1", Amount: coins("2apple")},
				{Account: "addr3", Amount: coins("9apple")},
			},
			expected: []AccountAmount{
				{Account: "addr1", Amount: coins("3apple")},
				{Account: "addr3", Amount: coins("9apple")},
			},
		},
		{
			name: "three entries: first third same addr",
			entries: []AccountAmount{
				{Account: "addr1", Amount: coins("1apple,2banana")},
				{Account: "addr3", Amount: coins("7prune")},
				{Account: "addr1", Amount: coins("3apple,5cherry")},
			},
			expected: []AccountAmount{
				{Account: "addr1", Amount: coins("4apple,2banana,5cherry")},
				{Account: "addr3", Amount: coins("7prune")},
			},
		},
		{
			name: "three entries: second third same addr",
			entries: []AccountAmount{
				{Account: "addr3", Amount: coins("9apple")},
				{Account: "addr1", Amount: coins("87banana")},
				{Account: "addr1", Amount: coins("65cherry")},
			},
			expected: []AccountAmount{
				{Account: "addr3", Amount: coins("9apple")},
				{Account: "addr1", Amount: coins("87banana,65cherry")},
			},
		},
		{
			name: "three entries: all same addr",
			entries: []AccountAmount{
				{Account: "addr1", Amount: coins("1apple,2banana")},
				{Account: "addr1", Amount: coins("7prune")},
				{Account: "addr1", Amount: coins("3apple,5cherry")},
			},
			expected: []AccountAmount{
				{Account: "addr1", Amount: coins("4apple,2banana,5cherry,7prune")},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []AccountAmount
			testFunc := func() {
				actual = SimplifyAccountAmounts(tc.entries)
			}
			require.NotPanics(t, testFunc, "SimplifyAccountAmounts")
			assertEqualSlice(t, tc.expected, actual, AccountAmount.String, "SimplifyAccountAmounts result")
		})
	}
}

func TestAccountAmountsToBankInputs(t *testing.T) {
	tests := []struct {
		name     string
		entries  []AccountAmount
		expected []banktypes.Input
	}{
		{
			name:     "no entries",
			entries:  nil,
			expected: []banktypes.Input{},
		},
		{
			name:     "one entry: good",
			entries:  []AccountAmount{{Account: "someacct", Amount: sdk.NewCoins(sdk.NewInt64Coin("cherry", 5000))}},
			expected: []banktypes.Input{{Address: "someacct", Coins: sdk.NewCoins(sdk.NewInt64Coin("cherry", 5000))}},
		},
		{
			name:     "one entry: no address",
			entries:  []AccountAmount{{Account: "", Amount: sdk.NewCoins(sdk.NewInt64Coin("cherry", 5000))}},
			expected: []banktypes.Input{{Address: "", Coins: sdk.NewCoins(sdk.NewInt64Coin("cherry", 5000))}},
		},
		{
			name:     "one entry: no amount",
			entries:  []AccountAmount{{Account: "someacct", Amount: nil}},
			expected: []banktypes.Input{{Address: "someacct", Coins: nil}},
		},
		{
			name: "one entry: bad amount",
			entries: []AccountAmount{
				{Account: "someacct", Amount: sdk.Coins{sdk.Coin{Denom: "cherry", Amount: sdkmath.NewInt(-2)}}},
			},
			expected: []banktypes.Input{
				{Address: "someacct", Coins: sdk.Coins{sdk.Coin{Denom: "cherry", Amount: sdkmath.NewInt(-2)}}},
			},
		},
		{
			name:     "one entry: empty",
			entries:  []AccountAmount{{Account: "", Amount: nil}},
			expected: []banktypes.Input{{Address: "", Coins: nil}},
		},
		{
			name: "three entries",
			entries: []AccountAmount{
				{Account: "addr0", Amount: sdk.NewCoins(sdk.NewInt64Coin("cherry", 23))},
				{Account: "addr1", Amount: sdk.NewCoins(sdk.NewInt64Coin("banana", 99))},
				{Account: "addr2", Amount: sdk.NewCoins(sdk.NewInt64Coin("apple", 42))},
			},
			expected: []banktypes.Input{
				{Address: "addr0", Coins: sdk.NewCoins(sdk.NewInt64Coin("cherry", 23))},
				{Address: "addr1", Coins: sdk.NewCoins(sdk.NewInt64Coin("banana", 99))},
				{Address: "addr2", Coins: sdk.NewCoins(sdk.NewInt64Coin("apple", 42))},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []banktypes.Input
			testFunc := func() {
				actual = AccountAmountsToBankInputs(tc.entries...)
			}
			require.NotPanics(t, testFunc, "AccountAmountsToBankInputs")
			assertEqualSlice(t, tc.expected, actual, bankInputString, "AccountAmountsToBankInputs")
		})
	}
}

func TestAccountAmountsToBankOutputs(t *testing.T) {
	tests := []struct {
		name     string
		entries  []AccountAmount
		expected []banktypes.Output
	}{
		{
			name:     "no entries",
			entries:  nil,
			expected: []banktypes.Output{},
		},
		{
			name:     "one entry: good",
			entries:  []AccountAmount{{Account: "someacct", Amount: sdk.NewCoins(sdk.NewInt64Coin("cherry", 5000))}},
			expected: []banktypes.Output{{Address: "someacct", Coins: sdk.NewCoins(sdk.NewInt64Coin("cherry", 5000))}},
		},
		{
			name:     "one entry: no address",
			entries:  []AccountAmount{{Account: "", Amount: sdk.NewCoins(sdk.NewInt64Coin("cherry", 5000))}},
			expected: []banktypes.Output{{Address: "", Coins: sdk.NewCoins(sdk.NewInt64Coin("cherry", 5000))}},
		},
		{
			name:     "one entry: no amount",
			entries:  []AccountAmount{{Account: "someacct", Amount: nil}},
			expected: []banktypes.Output{{Address: "someacct", Coins: nil}},
		},
		{
			name: "one entry: bad amount",
			entries: []AccountAmount{
				{Account: "someacct", Amount: sdk.Coins{sdk.Coin{Denom: "cherry", Amount: sdkmath.NewInt(-2)}}},
			},
			expected: []banktypes.Output{
				{Address: "someacct", Coins: sdk.Coins{sdk.Coin{Denom: "cherry", Amount: sdkmath.NewInt(-2)}}},
			},
		},
		{
			name:     "one entry: empty",
			entries:  []AccountAmount{{Account: "", Amount: nil}},
			expected: []banktypes.Output{{Address: "", Coins: nil}},
		},
		{
			name: "three entries",
			entries: []AccountAmount{
				{Account: "addr0", Amount: sdk.NewCoins(sdk.NewInt64Coin("cherry", 23))},
				{Account: "addr1", Amount: sdk.NewCoins(sdk.NewInt64Coin("banana", 99))},
				{Account: "addr2", Amount: sdk.NewCoins(sdk.NewInt64Coin("apple", 42))},
			},
			expected: []banktypes.Output{
				{Address: "addr0", Coins: sdk.NewCoins(sdk.NewInt64Coin("cherry", 23))},
				{Address: "addr1", Coins: sdk.NewCoins(sdk.NewInt64Coin("banana", 99))},
				{Address: "addr2", Coins: sdk.NewCoins(sdk.NewInt64Coin("apple", 42))},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []banktypes.Output
			testFunc := func() {
				actual = AccountAmountsToBankOutputs(tc.entries...)
			}
			require.NotPanics(t, testFunc, "AccountAmountsToBankOutputs")
			assertEqualSlice(t, tc.expected, actual, bankOutputString, "AccountAmountsToBankOutputs")
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

// TODO[1789]: func TestBuildCommitmentTransfers(t *testing.T)

func denomSourceMapString(dsm denomSourceMap) string {
	if dsm == nil {
		return "<nil>"
	}

	denoms := make([]string, 0, len(dsm))
	for denom, _ := range dsm {
		denoms = append(denoms, denom)
	}
	sort.Strings(denoms)
	entries := make([]string, len(denoms))
	for i, denom := range denoms {
		entry := make([]string, len(dsm[denom]))
		for j, aa := range dsm[denom] {
			entry[j] = fmt.Sprintf("{%s:%s}", aa.account, aa.int)
		}
		entries[i] = fmt.Sprintf("%s:[%s]", denom, strings.Join(entry, ","))
	}
	return fmt.Sprintf("{%s}", strings.Join(entries, ", "))
}

func TestNewDenomSourceMap(t *testing.T) {
	coins := func(str string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(str)
		require.NoError(t, err, "ParseCoinsNormalized(%q)", str)
		return rv
	}
	newAccountInt := func(account string, amount int64) *accountInt {
		return &accountInt{account: account, int: sdkmath.NewInt(amount)}
	}

	tests := []struct {
		name     string
		entries  []AccountAmount
		expected denomSourceMap
	}{
		{
			name:     "nil entries",
			entries:  nil,
			expected: make(denomSourceMap),
		},
		{
			name:     "empty entries",
			entries:  []AccountAmount{},
			expected: make(denomSourceMap),
		},
		{
			name: "one entry: one denom",
			entries: []AccountAmount{
				{Account: "addr0", Amount: coins("41cherry")},
			},
			expected: denomSourceMap{
				"cherry": []*accountInt{newAccountInt("addr0", 41)},
			},
		},
		{
			name: "one entry: three denoms",
			entries: []AccountAmount{
				{Account: "addr0", Amount: coins("29apple,76banana,41cherry")},
			},
			expected: denomSourceMap{
				"apple":  []*accountInt{newAccountInt("addr0", 29)},
				"banana": []*accountInt{newAccountInt("addr0", 76)},
				"cherry": []*accountInt{newAccountInt("addr0", 41)},
			},
		},
		{
			name: "two entries: one denom",
			entries: []AccountAmount{
				{Account: "addr0", Amount: coins("41cherry")},
				{Account: "addr1", Amount: coins("52cherry")},
			},
			expected: denomSourceMap{
				"cherry": []*accountInt{newAccountInt("addr0", 41), newAccountInt("addr1", 52)},
			},
		},
		{
			name: "two entries: three denoms",
			entries: []AccountAmount{
				{Account: "addr0", Amount: coins("6apple,41cherry")},
				{Account: "addr1", Amount: coins("78banana,52cherry")},
			},
			expected: denomSourceMap{
				"apple":  []*accountInt{newAccountInt("addr0", 6)},
				"banana": []*accountInt{newAccountInt("addr1", 78)},
				"cherry": []*accountInt{newAccountInt("addr0", 41), newAccountInt("addr1", 52)},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual denomSourceMap
			testFunc := func() {
				actual = newDenomSourceMap(tc.entries)
			}
			require.NotPanics(t, testFunc, "newDenomSourceMap")
			if !assert.Equal(t, tc.expected, actual, "newDenomSourceMap result") {
				expStr := denomSourceMapString(tc.expected)
				actStr := denomSourceMapString(actual)
				assert.Equal(t, expStr, actStr, "newDenomSourceMap result as strings")
			}
		})
	}
}

func TestDenomSourceMap_sum(t *testing.T) {
	tests := []struct {
		name     string
		dsm      denomSourceMap
		expected sdk.Coins
	}{
		{
			name:     "empty map",
			dsm:      make(denomSourceMap),
			expected: nil,
		},
		{
			name: "one denom, one addr",
			dsm: denomSourceMap{
				"cherry": []*accountInt{{account: "addr0", int: sdkmath.NewInt(5)}},
			},
			expected: sdk.NewCoins(sdk.NewInt64Coin("cherry", 5)),
		},
		{
			name: "one denom, three addrs",
			dsm: denomSourceMap{
				"cherry": []*accountInt{
					{account: "addr0", int: sdkmath.NewInt(5)},
					{account: "addr1", int: sdkmath.NewInt(9)},
					{account: "addr2", int: sdkmath.NewInt(22)},
				},
			},
			expected: sdk.NewCoins(sdk.NewInt64Coin("cherry", 36)),
		},
		{
			name: "two denoms, one addr",
			dsm: denomSourceMap{
				"banana": []*accountInt{{account: "addr0", int: sdkmath.NewInt(84)}},
				"cherry": []*accountInt{{account: "addr0", int: sdkmath.NewInt(5)}},
			},
			expected: sdk.NewCoins(sdk.NewInt64Coin("banana", 84), sdk.NewInt64Coin("cherry", 5)),
		},
		{
			name: "two denoms, three addrs",
			dsm: denomSourceMap{
				"banana": []*accountInt{
					{account: "addr0", int: sdkmath.NewInt(84)},
					{account: "addr1", int: sdkmath.NewInt(67)},
				},
				"cherry": []*accountInt{
					{account: "addr0", int: sdkmath.NewInt(5)},
					{account: "addr2", int: sdkmath.NewInt(12)},
				},
			},
			expected: sdk.NewCoins(sdk.NewInt64Coin("banana", 151), sdk.NewInt64Coin("cherry", 17)),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sdk.Coins
			testFunc := func() {
				actual = tc.dsm.sum()
			}
			require.NotPanics(t, testFunc, "denomSourceMap.sum()")
			assert.Equal(t, tc.expected.String(), actual.String(), "denomSourceMap.sum() result")
		})
	}
}

func TestDenomSourceMap_useCoin(t *testing.T) {
	tests := []struct {
		name       string
		funds      denomSourceMap
		coin       sdk.Coin
		source     string
		expAmounts []AccountAmount
		expErr     string
		expFunds   denomSourceMap
	}{
		{
			name: "unknown denom",
			funds: denomSourceMap{
				"cherry": []*accountInt{{account: "addr0", int: sdkmath.NewInt(15)}},
				"plum":   []*accountInt{{account: "addr0", int: sdkmath.NewInt(72)}},
			},
			coin:   sdk.NewInt64Coin("banana", 2),
			source: "testSource",
			expErr: "failed to allocate 2banana to testSource: 2 left over",
			expFunds: denomSourceMap{
				"cherry": []*accountInt{{account: "addr0", int: sdkmath.NewInt(15)}},
				"plum":   []*accountInt{{account: "addr0", int: sdkmath.NewInt(72)}},
			},
		},
		{
			name: "one entry: use all",
			funds: denomSourceMap{
				"cherry": []*accountInt{{account: "addr0", int: sdkmath.NewInt(15)}},
				"plum":   []*accountInt{{account: "addr0", int: sdkmath.NewInt(72)}},
			},
			coin:       sdk.NewInt64Coin("cherry", 15),
			source:     "testSource",
			expAmounts: []AccountAmount{{Account: "addr0", Amount: sdk.NewCoins(sdk.NewInt64Coin("cherry", 15))}},
			expFunds: denomSourceMap{
				"plum": []*accountInt{{account: "addr0", int: sdkmath.NewInt(72)}},
			},
		},
		{
			name: "one entry: use some",
			funds: denomSourceMap{
				"cherry": []*accountInt{{account: "addr0", int: sdkmath.NewInt(15)}},
				"plum":   []*accountInt{{account: "addr0", int: sdkmath.NewInt(72)}},
			},
			coin:       sdk.NewInt64Coin("cherry", 9),
			source:     "testSource",
			expAmounts: []AccountAmount{{Account: "addr0", Amount: sdk.NewCoins(sdk.NewInt64Coin("cherry", 9))}},
			expFunds: denomSourceMap{
				"cherry": []*accountInt{{account: "addr0", int: sdkmath.NewInt(6)}},
				"plum":   []*accountInt{{account: "addr0", int: sdkmath.NewInt(72)}},
			},
		},
		{
			name: "one entry: use too much",
			funds: denomSourceMap{
				"cherry": []*accountInt{{account: "addr0", int: sdkmath.NewInt(15)}},
				"plum":   []*accountInt{{account: "addr0", int: sdkmath.NewInt(72)}},
			},
			coin:   sdk.NewInt64Coin("cherry", 16),
			source: "outputs",
			expErr: "failed to allocate 16cherry to outputs: 1 left over",
			expFunds: denomSourceMap{
				"plum": []*accountInt{{account: "addr0", int: sdkmath.NewInt(72)}},
			},
		},
		{
			name: "two entries: use some of first",
			funds: denomSourceMap{
				"cherry": []*accountInt{
					{account: "addr0", int: sdkmath.NewInt(15)},
					{account: "addr1", int: sdkmath.NewInt(34)},
				},
				"plum": []*accountInt{{account: "addr0", int: sdkmath.NewInt(72)}},
			},
			coin:       sdk.NewInt64Coin("cherry", 14),
			source:     "testSource",
			expAmounts: []AccountAmount{{Account: "addr0", Amount: sdk.NewCoins(sdk.NewInt64Coin("cherry", 14))}},
			expFunds: denomSourceMap{
				"cherry": []*accountInt{
					{account: "addr0", int: sdkmath.NewInt(1)},
					{account: "addr1", int: sdkmath.NewInt(34)},
				},
				"plum": []*accountInt{{account: "addr0", int: sdkmath.NewInt(72)}},
			},
		},
		{
			name: "two entries: use exactly first",
			funds: denomSourceMap{
				"cherry": []*accountInt{
					{account: "addr0", int: sdkmath.NewInt(15)},
					{account: "addr1", int: sdkmath.NewInt(34)},
				},
				"plum": []*accountInt{{account: "addr0", int: sdkmath.NewInt(72)}},
			},
			coin:       sdk.NewInt64Coin("cherry", 15),
			source:     "testSource",
			expAmounts: []AccountAmount{{Account: "addr0", Amount: sdk.NewCoins(sdk.NewInt64Coin("cherry", 15))}},
			expFunds: denomSourceMap{
				"cherry": []*accountInt{{account: "addr1", int: sdkmath.NewInt(34)}},
				"plum":   []*accountInt{{account: "addr0", int: sdkmath.NewInt(72)}},
			},
		},
		{
			name: "two entries: use some of second",
			funds: denomSourceMap{
				"cherry": []*accountInt{
					{account: "addr0", int: sdkmath.NewInt(15)},
					{account: "addr1", int: sdkmath.NewInt(34)},
				},
				"plum": []*accountInt{{account: "addr0", int: sdkmath.NewInt(72)}},
			},
			coin:   sdk.NewInt64Coin("cherry", 22),
			source: "testSource",
			expAmounts: []AccountAmount{
				{Account: "addr0", Amount: sdk.NewCoins(sdk.NewInt64Coin("cherry", 15))},
				{Account: "addr1", Amount: sdk.NewCoins(sdk.NewInt64Coin("cherry", 7))},
			},
			expFunds: denomSourceMap{
				"cherry": []*accountInt{{account: "addr1", int: sdkmath.NewInt(27)}},
				"plum":   []*accountInt{{account: "addr0", int: sdkmath.NewInt(72)}},
			},
		},
		{
			name: "two entries: use all of both",
			funds: denomSourceMap{
				"cherry": []*accountInt{
					{account: "addr0", int: sdkmath.NewInt(15)},
					{account: "addr1", int: sdkmath.NewInt(34)},
				},
				"plum": []*accountInt{{account: "addr0", int: sdkmath.NewInt(72)}},
			},
			coin:   sdk.NewInt64Coin("cherry", 49),
			source: "testSource",
			expAmounts: []AccountAmount{
				{Account: "addr0", Amount: sdk.NewCoins(sdk.NewInt64Coin("cherry", 15))},
				{Account: "addr1", Amount: sdk.NewCoins(sdk.NewInt64Coin("cherry", 34))},
			},
			expFunds: denomSourceMap{
				"plum": []*accountInt{{account: "addr0", int: sdkmath.NewInt(72)}},
			},
		},
		{
			name: "two entries: use too much",
			funds: denomSourceMap{
				"cherry": []*accountInt{
					{account: "addr0", int: sdkmath.NewInt(15)},
					{account: "addr1", int: sdkmath.NewInt(34)},
				},
				"plum": []*accountInt{{account: "addr0", int: sdkmath.NewInt(72)}},
			},
			coin:   sdk.NewInt64Coin("cherry", 53),
			source: "inputs",
			expErr: "failed to allocate 53cherry to inputs: 4 left over",
			expFunds: denomSourceMap{
				"plum": []*accountInt{{account: "addr0", int: sdkmath.NewInt(72)}},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actAmounts []AccountAmount
			var err error
			testFunc := func() {
				actAmounts, err = tc.funds.useCoin(tc.coin, tc.source)
			}
			require.NotPanics(t, testFunc, "useCoin(%q)", tc.coin)
			assertions.AssertErrorValue(t, err, tc.expErr, "useCoin(%q) error", tc.coin)
			assertEqualSlice(t, tc.expAmounts, actAmounts, AccountAmount.String, "useCoin(%q) result", tc.coin)
			if !assert.Equal(t, tc.expFunds, tc.funds, "receiver after useCoin(%q)", tc.coin) {
				expStr := denomSourceMapString(tc.expFunds)
				actStr := denomSourceMapString(tc.funds)
				assert.Equal(t, expStr, actStr, "receiver as string after useCoin(%q)", tc.coin)
			}
		})
	}
}

func TestDenomSourceMap_useCoins(t *testing.T) {
	coins := func(str string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(str)
		require.NoError(t, err, "ParseCoinsNormalized(%q)", str)
		return rv
	}

	tests := []struct {
		name       string
		funds      denomSourceMap
		coins      sdk.Coins
		source     string
		expAmounts []AccountAmount
		expErr     string
		expFunds   denomSourceMap
	}{
		{
			name: "one denom: insufficient funds",
			funds: denomSourceMap{
				"cherry": []*accountInt{{account: "addr0", int: sdkmath.NewInt(44)}},
				"pear":   []*accountInt{{account: "addr0", int: sdkmath.NewInt(7000)}},
			},
			coins:  coins("99cherry"),
			source: "testSource",
			expErr: "failed to allocate 99cherry to testSource: 55 left over",
			expFunds: denomSourceMap{
				"pear": []*accountInt{{account: "addr0", int: sdkmath.NewInt(7000)}},
			},
		},
		{
			name: "one denom: no entry",
			funds: denomSourceMap{
				"cherry": []*accountInt{{account: "addr0", int: sdkmath.NewInt(44)}},
				"pear":   []*accountInt{{account: "addr0", int: sdkmath.NewInt(7000)}},
			},
			coins:  coins("99banana"),
			source: "testSource",
			expErr: "failed to allocate 99banana to testSource: 99 left over",
			expFunds: denomSourceMap{
				"cherry": []*accountInt{{account: "addr0", int: sdkmath.NewInt(44)}},
				"pear":   []*accountInt{{account: "addr0", int: sdkmath.NewInt(7000)}},
			},
		},
		{
			name: "two denoms, one account",
			funds: denomSourceMap{
				"banana": []*accountInt{{account: "addr0", int: sdkmath.NewInt(99)}},
				"cherry": []*accountInt{{account: "addr0", int: sdkmath.NewInt(44)}},
				"pear":   []*accountInt{{account: "addr0", int: sdkmath.NewInt(7000)}},
			},
			coins:  coins("90banana,33cherry"),
			source: "testSource",
			expAmounts: []AccountAmount{
				{Account: "addr0", Amount: coins("90banana,33cherry")},
			},
			expFunds: denomSourceMap{
				"banana": []*accountInt{{account: "addr0", int: sdkmath.NewInt(9)}},
				"cherry": []*accountInt{{account: "addr0", int: sdkmath.NewInt(11)}},
				"pear":   []*accountInt{{account: "addr0", int: sdkmath.NewInt(7000)}},
			},
		},
		{
			name: "one denom, two accounts",
			funds: denomSourceMap{
				"cherry": []*accountInt{
					{account: "addr0", int: sdkmath.NewInt(44)},
					{account: "addr1", int: sdkmath.NewInt(56)},
				},
				"pear": []*accountInt{{account: "addr0", int: sdkmath.NewInt(7000)}},
			},
			coins:  coins("99cherry"),
			source: "testSource",
			expAmounts: []AccountAmount{
				{Account: "addr0", Amount: coins("44cherry")},
				{Account: "addr1", Amount: coins("55cherry")},
			},
			expFunds: denomSourceMap{
				"cherry": []*accountInt{{account: "addr1", int: sdkmath.NewInt(1)}},
				"pear":   []*accountInt{{account: "addr0", int: sdkmath.NewInt(7000)}},
			},
		},
		{
			name: "two denoms, two accounts",
			funds: denomSourceMap{
				"banana": []*accountInt{
					{account: "addr0", int: sdkmath.NewInt(20)},
					{account: "addr1", int: sdkmath.NewInt(30)},
				},
				"cherry": []*accountInt{
					{account: "addr1", int: sdkmath.NewInt(44)},
					{account: "addr0", int: sdkmath.NewInt(55)},
				},
				"pear": []*accountInt{{account: "addr0", int: sdkmath.NewInt(7000)}},
			},
			coins:  coins("50banana,99cherry"),
			source: "testSource",
			expAmounts: []AccountAmount{
				{Account: "addr0", Amount: coins("20banana,55cherry")},
				{Account: "addr1", Amount: coins("30banana,44cherry")},
			},
			expFunds: denomSourceMap{
				"pear": []*accountInt{{account: "addr0", int: sdkmath.NewInt(7000)}},
			},
		},
		{
			name: "two denoms: error from first",
			funds: denomSourceMap{
				"banana": []*accountInt{{account: "addr0", int: sdkmath.NewInt(99)}},
				"cherry": []*accountInt{{account: "addr0", int: sdkmath.NewInt(44)}},
				"pear":   []*accountInt{{account: "addr0", int: sdkmath.NewInt(7000)}},
			},
			coins:  coins("100banana,44cherry"),
			source: "outputs",
			expErr: "failed to allocate 100banana to outputs: 1 left over",
			expFunds: denomSourceMap{
				"cherry": []*accountInt{{account: "addr0", int: sdkmath.NewInt(44)}},
				"pear":   []*accountInt{{account: "addr0", int: sdkmath.NewInt(7000)}},
			},
		},
		{
			name: "two denoms: error from second",
			funds: denomSourceMap{
				"banana": []*accountInt{{account: "addr0", int: sdkmath.NewInt(99)}},
				"cherry": []*accountInt{{account: "addr0", int: sdkmath.NewInt(44)}},
				"pear":   []*accountInt{{account: "addr0", int: sdkmath.NewInt(7000)}},
			},
			coins:  coins("99banana,45cherry"),
			source: "inputs",
			expErr: "failed to allocate 45cherry to inputs: 1 left over",
			expFunds: denomSourceMap{
				"pear": []*accountInt{{account: "addr0", int: sdkmath.NewInt(7000)}},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actAmounts []AccountAmount
			var err error
			testFunc := func() {
				actAmounts, err = tc.funds.useCoins(tc.coins, tc.source)
			}
			require.NotPanics(t, testFunc, "useCoins(%q)", tc.coins)
			assertions.AssertErrorValue(t, err, tc.expErr, "useCoins(%q) error", tc.coins)
			assertEqualSlice(t, tc.expAmounts, actAmounts, AccountAmount.String, "useCoins(%q) result", tc.coins)
			if !assert.Equal(t, tc.expFunds, tc.funds, "receiver after useCoins(%q)", tc.coins) {
				expStr := denomSourceMapString(tc.expFunds)
				actStr := denomSourceMapString(tc.funds)
				assert.Equal(t, expStr, actStr, "receiver as string after useCoins(%q)", tc.coins)
			}
		})
	}
}

// TODO[1789]: func TestBuildPrimaryTransfers(t *testing.T)

func TestBuildFeesTransfer(t *testing.T) {
	coins := func(str string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(str)
		require.NoError(t, err, "ParseCoinsNormalized(%q)", str)
		return rv
	}

	tests := []struct {
		name     string
		marketID uint32
		fees     []AccountAmount
		expected *Transfer
	}{
		{
			name:     "nil fees",
			marketID: 3,
			fees:     nil,
			expected: nil,
		},
		{
			name:     "empty fees",
			marketID: 3,
			fees:     []AccountAmount{},
			expected: nil,
		},
		{
			name:     "one fee",
			marketID: 3,
			fees:     []AccountAmount{{Account: "addr0", Amount: coins("13fig")}},
			expected: &Transfer{
				Inputs:  []banktypes.Input{{Address: "addr0", Coins: coins("13fig")}},
				Outputs: []banktypes.Output{{Address: GetMarketAddress(3).String(), Coins: coins("13fig")}},
			},
		},
		{
			name:     "three fees",
			marketID: 8,
			fees: []AccountAmount{
				{Account: "addr0", Amount: coins("13fig")},
				{Account: "addr1", Amount: coins("7fig,8grape")},
				{Account: "addr2", Amount: coins("4fig")},
			},
			expected: &Transfer{
				Inputs: []banktypes.Input{
					{Address: "addr0", Coins: coins("13fig")},
					{Address: "addr1", Coins: coins("7fig,8grape")},
					{Address: "addr2", Coins: coins("4fig")},
				},
				Outputs: []banktypes.Output{{Address: GetMarketAddress(8).String(), Coins: coins("24fig,8grape")}},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual *Transfer
			testFunc := func() {
				actual = buildFeesTransfer(tc.marketID, tc.fees)
			}
			require.NotPanics(t, testFunc, "buildFeesTransfer")
			if !assert.Equal(t, tc.expected, actual, "buildFeesTransfer") {
				expStr := transferString(tc.expected)
				actStr := transferString(actual)
				assert.Equal(t, expStr, actStr, "buildFeesTransfer as string")
			}
		})
	}
}
