package types_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil/assertions"

	. "github.com/provenance-io/provenance/x/flatfees/types"
)

func TestDefaultParams(t *testing.T) {
	pioconfig.SetProvConfig("pineapple")

	var params Params
	testFunc := func() {
		params = DefaultParams()
	}
	require.NotPanics(t, testFunc, "DefaultParams()")

	err := params.Validate()
	assert.NoError(t, err, "params.Validate()")

	assert.Equal(t, "1"+DefaultFeeDefinitionDenom, params.DefaultCost.String(), "DefaultCost")
	assert.Equal(t, "1"+DefaultFeeDefinitionDenom, params.ConversionFactor.DefinitionAmount.String(), "ConversionFactor.DefinitionAmount")
	assert.Equal(t, "1pineapple", params.ConversionFactor.ConvertedAmount.String(), "ConversionFactor.ConvertedAmount")
}

func TestParams_Validate(t *testing.T) {
	coin := func(amt int64, denom string) sdk.Coin {
		return sdk.Coin{
			Denom:  denom,
			Amount: sdkmath.NewInt(amt),
		}
	}

	tests := []struct {
		name   string
		params Params
		expErr string
	}{
		{
			name:   "zero-value",
			params: Params{},
			expErr: "invalid default cost \"<nil>\": invalid denom: ",
		},
		{
			name: "invalid cost",
			params: Params{
				DefaultCost: coin(10, "x"),
				ConversionFactor: ConversionFactor{
					DefinitionAmount: coin(100, "banana"),
					ConvertedAmount:  coin(500, "apple"),
				},
			},
			expErr: "invalid default cost \"10x\": invalid denom: x",
		},
		{
			name: "invalid conversion factor",
			params: Params{
				DefaultCost: coin(10, "banana"),
				ConversionFactor: ConversionFactor{
					DefinitionAmount: coin(10, "banana"),
					ConvertedAmount:  coin(5, "x"),
				},
			},
			expErr: "invalid conversion factor: invalid converted amount \"5x\": invalid denom: x",
		},
		{
			name: "wrong conversion factor base denom",
			params: Params{
				DefaultCost: coin(10, "banana"),
				ConversionFactor: ConversionFactor{
					DefinitionAmount: coin(7, "apple"),
					ConvertedAmount:  coin(3, "banana"),
				},
			},
			expErr: "default cost denom \"banana\" does not equal conversion factor base amount denom \"apple\"",
		},
		{
			name: "valid",
			params: Params{
				DefaultCost: coin(52, "banana"),
				ConversionFactor: ConversionFactor{
					DefinitionAmount: coin(13, "banana"),
					ConvertedAmount:  coin(27, "apple"),
				},
			},
			expErr: "",
		},
		{
			name:   "default",
			params: DefaultParams(),
			expErr: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.params.Validate()
			}
			require.NotPanics(t, testFunc, "params.Validate()")
			assertions.AssertErrorValue(t, err, tc.expErr, "params.Validate() error")
		})
	}
}

func TestParams_DefaultCostCoins(t *testing.T) {
	coin := func(amt int64, denom string) sdk.Coin {
		return sdk.Coin{
			Denom:  denom,
			Amount: sdkmath.NewInt(amt),
		}
	}

	tests := []struct {
		name   string
		params Params
		exp    sdk.Coins
	}{
		{
			name:   "nil default cost",
			params: Params{DefaultCost: sdk.Coin{}},
			exp:    nil,
		},
		{
			name:   "zero default cost",
			params: Params{DefaultCost: coin(0, "banana")},
			exp:    nil,
		},
		{
			name:   "invalid default cost",
			params: Params{DefaultCost: coin(1, "x")},
			exp:    sdk.Coins{coin(1, "x")},
		},
		{
			name:   "normal default cost",
			params: Params{DefaultCost: coin(10, "banana")},
			exp:    sdk.Coins{coin(10, "banana")},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act sdk.Coins
			testFunc := func() {
				act = tc.params.DefaultCostCoins()
			}
			require.NotPanics(t, testFunc, "params.DefaultCostCoins()")
			assert.Equal(t, tc.exp.String(), act.String(), "params.DefaultCostCoins() result")
		})
	}
}

func TestNewMsgFee(t *testing.T) {
	coin := func(amt int64, denom string) sdk.Coin {
		return sdk.Coin{
			Denom:  denom,
			Amount: sdkmath.NewInt(amt),
		}
	}
	cz := func(coins ...sdk.Coin) []sdk.Coin {
		return coins
	}

	msgTypeURL := sdk.MsgTypeURL(&MsgUpdateMsgFeesRequest{}) // "/provenance.flatfees.v1.MsgUpdateMsgFeesRequest"

	tests := []struct {
		name string
		url  string
		cost []sdk.Coin
		exp  *MsgFee
	}{
		{
			name: "empty msg type url",
			url:  "",
			cost: cz(coin(10, "banana")),
			exp:  &MsgFee{MsgTypeUrl: "", Cost: cz(coin(10, "banana"))},
		},
		{
			name: "no costs",
			url:  msgTypeURL,
			cost: nil,
			exp:  &MsgFee{MsgTypeUrl: msgTypeURL, Cost: nil},
		},
		{
			name: "one cost",
			url:  msgTypeURL + "x",
			cost: cz(coin(10, "banana")),
			exp:  &MsgFee{MsgTypeUrl: msgTypeURL + "x", Cost: cz(coin(10, "banana"))},
		},
		{
			name: "two costs: wrong order",
			url:  msgTypeURL,
			cost: cz(coin(10, "banana"), coin(5, "apple")),
			exp:  &MsgFee{MsgTypeUrl: msgTypeURL, Cost: sdk.Coins{coin(5, "apple"), coin(10, "banana")}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act *MsgFee
			testFunc := func() {
				act = NewMsgFee(tc.url, tc.cost...)
			}
			require.NotPanics(t, testFunc, "NewMsgFee")
			ok := assert.Equal(t, tc.exp.MsgTypeUrl, act.MsgTypeUrl, "MsgTypeUrl")
			ok = assert.Equal(t, tc.exp.Cost.String(), act.Cost.String(), "Cost") && ok
			if ok {
				assert.Equal(t, tc.exp, act, "NewMsgFee result")
			}
		})
	}
}

func TestMsgFee_String(t *testing.T) {
	coin := func(amt int64, denom string) sdk.Coin {
		return sdk.Coin{
			Denom:  denom,
			Amount: sdkmath.NewInt(amt),
		}
	}
	cz := func(coins ...sdk.Coin) []sdk.Coin {
		return coins
	}

	tests := []struct {
		name   string
		msgFee *MsgFee
		exp    string
	}{
		{
			name:   "nil",
			msgFee: nil,
			exp:    "<nil>",
		},
		{
			name:   "zero-value",
			msgFee: &MsgFee{},
			exp:    `""=<free>`,
		},
		{
			name:   "url without cost",
			msgFee: &MsgFee{MsgTypeUrl: "abc"},
			exp:    "abc=<free>",
		},
		{
			name:   "cost without url",
			msgFee: &MsgFee{Cost: cz(coin(3, "banana"))},
			exp:    `""=3banana`,
		},
		{
			name: "url with 1 coin cost",
			msgFee: &MsgFee{
				MsgTypeUrl: "dance",
				Cost:       cz(coin(2, "avocado")),
			},
			exp: "dance=2avocado",
		},
		{
			name: "url with 2 coin cost",
			msgFee: &MsgFee{
				MsgTypeUrl: "dance",
				Cost:       cz(coin(2, "avocado"), coin(3, "banana")),
			},
			exp: "dance=2avocado,3banana",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act string
			testFunc := func() {
				act = tc.msgFee.String()
			}
			require.NotPanics(t, testFunc, "msgFee.String()")
			assert.Equal(t, tc.exp, act, "msgFee.String() result")
		})
	}
}

func TestMsgFee_Validate(t *testing.T) {
	coin := func(amt int64, denom string) sdk.Coin {
		return sdk.Coin{
			Denom:  denom,
			Amount: sdkmath.NewInt(amt),
		}
	}

	validMsgTypeURL := sdk.MsgTypeURL(&MsgUpdateMsgFeesRequest{}) // "/provenance.flatfees.v1.MsgUpdateMsgFeesRequest"
	validCost := coin(100, sdk.DefaultBondDenom)

	cases := []struct {
		name string
		msg  *MsgFee
		exp  string
	}{
		{
			name: "nil",
			msg:  nil,
			exp:  "nil MsgFee not allowed",
		},

		{
			name: "msg type url empty",
			msg:  NewMsgFee("", validCost),
			exp:  "msg type url cannot be empty",
		},
		{
			name: "msg type url too long",
			msg:  NewMsgFee("ab"+strings.Repeat("x", MaxMsgTypeURLLen-3)+"ba", validCost),
			exp:  "msg type url \"abxxx...xxxba\" length (161) exceeds max length (160)",
		},
		{
			name: "msg type url at max length",
			msg:  NewMsgFee(strings.Repeat("x", MaxMsgTypeURLLen), validCost),
			exp:  "",
		},

		{
			name: "invalid cost",
			msg:  &MsgFee{MsgTypeUrl: validMsgTypeURL, Cost: sdk.Coins{coin(100, "x")}},
			exp:  "invalid " + validMsgTypeURL + " cost \"100x\": invalid denom: x",
		},
		{
			name: "no cost",
			msg:  NewMsgFee(validMsgTypeURL),
			exp:  "",
		},
		{
			name: "cost with two coins",
			msg:  NewMsgFee(validMsgTypeURL, validCost, coin(99, "banana")),
			exp:  "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.msg.Validate()
			}
			require.NotPanics(t, testFunc, "MsgFee.Validate()")
			assertions.AssertErrorValue(t, err, tc.exp, "MsgFee.Validate() error")
		})
	}
}

func TestValidateMsgTypeURL(t *testing.T) {
	tests := []struct {
		name       string
		msgTypeURL string
		expErr     string
	}{
		{
			name:       "empty",
			msgTypeURL: "",
			expErr:     "msg type url cannot be empty",
		},
		{
			name:       "one char",
			msgTypeURL: "x",
		},
		{
			name:       "80 chars",
			msgTypeURL: "/ibc.applications.interchain_accounts.controller.v1.MsgRegisterInterchainAccount",
		},
		{
			name:       "max minus one",
			msgTypeURL: "/" + strings.Repeat("v", MaxMsgTypeURLLen-2),
		},
		{
			name:       "max",
			msgTypeURL: "/" + strings.Repeat("d", MaxMsgTypeURLLen-1),
		},
		{
			name:       "max plus one",
			msgTypeURL: "/" + strings.Repeat("d", MaxMsgTypeURLLen),
			expErr:     "msg type url \"/dddd...ddddd\" length (161) exceeds max length (160)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = ValidateMsgTypeURL(tc.msgTypeURL)
			}
			require.NotPanics(t, testFunc, "ValidateMsgTypeURL(%q)", tc.msgTypeURL)
			assertions.AssertErrorValue(t, err, tc.expErr, "ValidateMsgTypeURL(%q) error", tc.msgTypeURL)
		})
	}
}

func TestConversionFactor_Validate(t *testing.T) {
	coin := func(amt int64, denom string) sdk.Coin {
		return sdk.Coin{
			Denom:  denom,
			Amount: sdkmath.NewInt(amt),
		}
	}

	tests := []struct {
		name   string
		cf     ConversionFactor
		expErr string
	}{
		{
			name:   "zero-value",
			cf:     ConversionFactor{},
			expErr: "invalid base amount \"<nil>\": invalid denom: ",
		},
		{
			name: "invalid base amount",
			cf: ConversionFactor{
				DefinitionAmount: coin(3, "x"),
				ConvertedAmount:  coin(10, "banana"),
			},
			expErr: "invalid base amount \"3x\": invalid denom: x",
		},
		{
			name: "zero base amount",
			cf: ConversionFactor{
				DefinitionAmount: coin(0, "apple"),
				ConvertedAmount:  coin(10, "banana"),
			},
			expErr: "invalid base amount \"0apple\": cannot be zero",
		},
		{
			name: "invalid converted amount",
			cf: ConversionFactor{
				DefinitionAmount: coin(10, "banana"),
				ConvertedAmount:  coin(4, "x"),
			},
			expErr: "invalid converted amount \"4x\": invalid denom: x",
		},
		{
			name: "zero converted amount",
			cf: ConversionFactor{
				DefinitionAmount: coin(10, "apple"),
				ConvertedAmount:  coin(0, "banana"),
			},
			expErr: "invalid converted amount \"0banana\": cannot be zero",
		},
		{
			name: "same denoms, diff amounts",
			cf: ConversionFactor{
				DefinitionAmount: coin(10, "banana"),
				ConvertedAmount:  coin(11, "banana"),
			},
			expErr: "base amount \"10banana\" and converted amount \"11banana\" cannot have different amounts when the denoms are the same",
		},
		{
			name: "same denoms and amounts",
			cf: ConversionFactor{
				DefinitionAmount: coin(14, "banana"),
				ConvertedAmount:  coin(14, "banana"),
			},
			expErr: "",
		},
		{
			name: "diff denoms, same amounts",
			cf: ConversionFactor{
				DefinitionAmount: coin(14, "banana"),
				ConvertedAmount:  coin(14, "apple"),
			},
			expErr: "",
		},
		{
			name: "diff denoms and amounts",
			cf: ConversionFactor{
				DefinitionAmount: coin(17, "banana"),
				ConvertedAmount:  coin(23, "apple"),
			},
			expErr: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.cf.Validate()
			}
			require.NotPanics(t, testFunc, "ConversionFactor.Validate()")
			assertions.AssertErrorValue(t, err, tc.expErr, "ConversionFactor.Validate() error")
		})
	}
}

func TestConversionFactor_String(t *testing.T) {
	coin := func(amt int64, denom string) sdk.Coin {
		return sdk.Coin{
			Denom:  denom,
			Amount: sdkmath.NewInt(amt),
		}
	}

	tests := []struct {
		name string
		cf   ConversionFactor
		exp  string
	}{
		{
			name: "zero-value",
			cf:   ConversionFactor{},
			exp:  "*<nil>/<nil>",
		},
		{
			name: "nil base, non-nil converted amount",
			cf:   ConversionFactor{ConvertedAmount: coin(10, "banana")},
			exp:  "*<nil>/10banana",
		},
		{
			name: "non-nil base, nil converted amount",
			cf:   ConversionFactor{DefinitionAmount: coin(10, "banana")},
			exp:  "*10banana/<nil>",
		},
		{
			name: "equal amounts",
			cf: ConversionFactor{
				DefinitionAmount: coin(10, "banana"),
				ConvertedAmount:  coin(10, "banana"),
			},
			exp: "*10banana/10banana",
		},
		{
			name: "diff amounts",
			cf: ConversionFactor{
				DefinitionAmount: coin(13, "apple"),
				ConvertedAmount:  coin(29, "pear"),
			},
			exp: "*13apple/29pear",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act string
			testFunc := func() {
				act = tc.cf.String()
			}
			require.NotPanics(t, testFunc, "cf.String()")
			assert.Equal(t, tc.exp, act, "cf.String() result")
		})
	}
}

func TestConversionFactor_ConvertCoin(t *testing.T) {
	coin := func(amt int64, denom string) sdk.Coin {
		return sdk.Coin{
			Denom:  denom,
			Amount: sdkmath.NewInt(amt),
		}
	}
	newCF := func(baseAmt int64, baseDenom string, convAmt int64, convDenom string) ConversionFactor {
		return ConversionFactor{
			DefinitionAmount: coin(baseAmt, baseDenom),
			ConvertedAmount:  coin(convAmt, convDenom),
		}
	}

	tests := []struct {
		name string
		cf   ConversionFactor
		arg  sdk.Coin
		exp  sdk.Coin
	}{
		{
			name: "diff denom",
			cf:   newCF(10, "banana", 5, "apple"),
			arg:  coin(14, "plum"),
			exp:  coin(14, "plum"),
		},
		{
			name: "no conversion",
			cf:   newCF(100, "banana", 100, "banana"),
			arg:  coin(3, "banana"),
			exp:  coin(3, "banana"),
		},
		{
			name: "denom change",
			cf:   newCF(100, "banana", 100, "apple"),
			arg:  coin(7, "banana"),
			exp:  coin(7, "apple"),
		},
		{
			name: "remainder: 0",
			cf:   newCF(10, "banana", 4, "apple"),
			arg:  coin(45, "banana"), // 45 * 4 = 180, 180/10 = 18r0 => 18
			exp:  coin(18, "apple"),
		},
		{
			name: "remainder: 1 of 10",
			cf:   newCF(10, "banana", 3, "apple"),
			arg:  coin(57, "banana"), // 57 * 3 = 171, 171/10 = 17r1 => 18
			exp:  coin(18, "apple"),
		},
		{
			name: "remainder: 9 of 10",
			cf:   newCF(10, "banana", 3, "apple"),
			arg:  coin(63, "banana"), // 63 * 3 = 189, 189/10 = 18r9 => 19
			exp:  coin(19, "apple"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act sdk.Coin
			testFunc := func() {
				act = tc.cf.ConvertCoin(tc.arg)
			}
			require.NotPanics(t, testFunc, "%s ConvertCoin(%s)", tc.cf, tc.arg)
			assert.Equal(t, tc.exp, act, "%s ConvertCoin(%s)", tc.cf, tc.arg)
		})
	}
}

func TestConversionFactor_ConvertCoins(t *testing.T) {
	coin := func(amt int64, denom string) sdk.Coin {
		return sdk.Coin{
			Denom:  denom,
			Amount: sdkmath.NewInt(amt),
		}
	}
	cz := func(coins ...sdk.Coin) []sdk.Coin {
		return coins
	}
	newCF := func(baseAmt int64, baseDenom string, convAmt int64, convDenom string) ConversionFactor {
		return ConversionFactor{
			DefinitionAmount: coin(baseAmt, baseDenom),
			ConvertedAmount:  coin(convAmt, convDenom),
		}
	}

	tests := []struct {
		name string
		cf   ConversionFactor
		arg  sdk.Coins
		exp  sdk.Coins
	}{
		{
			name: "nil",
			cf:   newCF(10, "banana", 5, "apple"),
			arg:  nil,
			exp:  nil,
		},
		{
			name: "empty",
			cf:   newCF(10, "banana", 5, "apple"),
			arg:  sdk.Coins{},
			exp:  nil,
		},
		{
			name: "one coin: convertable",
			cf:   newCF(10, "banana", 5, "apple"),
			arg:  cz(coin(30, "banana")),
			exp:  cz(coin(15, "apple")),
		},
		{
			name: "one coin: inconvertable",
			cf:   newCF(10, "banana", 5, "apple"),
			arg:  cz(coin(13, "plum")),
			exp:  cz(coin(13, "plum")),
		},
		{
			name: "two coins: first is convertable",
			cf:   newCF(10, "banana", 5, "apple"),
			arg:  cz(coin(13, "banana"), coin(41, "plum")),
			exp:  cz(coin(7, "apple"), coin(41, "plum")),
		},
		{
			name: "two coins: second is convertable",
			cf:   newCF(10, "banana", 5, "apple"),
			arg:  cz(coin(14, "acorn"), coin(54, "banana")),
			exp:  cz(coin(14, "acorn"), coin(27, "apple")),
		},
		{
			name: "two coins: one converts to the other",
			cf:   newCF(10, "banana", 5, "apple"),
			arg:  cz(coin(16, "apple"), coin(24, "banana")),
			exp:  cz(coin(28, "apple")),
		},
		{
			name: "two coins: both inconvertable",
			cf:   newCF(10, "banana", 5, "apple"),
			arg:  cz(coin(53, "acorn"), coin(2, "plum")),
			exp:  cz(coin(53, "acorn"), coin(2, "plum")),
		},
		{
			name: "no conversion: two coins",
			cf:   newCF(1000, "banana", 1000, "banana"),
			arg:  cz(coin(57, "apple"), coin(117, "banana")),
			exp:  cz(coin(57, "apple"), coin(117, "banana")),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act sdk.Coins
			testFunc := func() {
				act = tc.cf.ConvertCoins(tc.arg)
			}
			require.NotPanics(t, testFunc, "%s ConvertCoins(%s)", tc.cf, tc.arg)
			assert.Equal(t, tc.exp.String(), act.String(), "%s ConvertCoins(%s) result", tc.cf, tc.arg)
		})
	}
}

func TestConversionFactor_ConvertMsgFee(t *testing.T) {
	coin := func(amt int64, denom string) sdk.Coin {
		return sdk.Coin{
			Denom:  denom,
			Amount: sdkmath.NewInt(amt),
		}
	}
	cz := func(coins ...sdk.Coin) []sdk.Coin {
		return coins
	}
	newCF := func(baseAmt int64, baseDenom string, convAmt int64, convDenom string) ConversionFactor {
		return ConversionFactor{
			DefinitionAmount: coin(baseAmt, baseDenom),
			ConvertedAmount:  coin(convAmt, convDenom),
		}
	}

	tests := []struct {
		name   string
		cf     ConversionFactor
		msgFee *MsgFee
		exp    *MsgFee
	}{
		{
			name:   "nil",
			cf:     newCF(10, "banana", 5, "apple"),
			msgFee: nil,
			exp:    nil,
		},
		{
			name:   "zero-value",
			cf:     newCF(10, "banana", 5, "apple"),
			msgFee: &MsgFee{},
			exp:    &MsgFee{},
		},
		{
			name:   "one coin: convertable",
			cf:     newCF(10, "banana", 5, "apple"),
			msgFee: &MsgFee{MsgTypeUrl: "abcdef", Cost: cz(coin(43, "banana"))},
			exp:    &MsgFee{MsgTypeUrl: "abcdef", Cost: cz(coin(22, "apple"))},
		},
		{
			name:   "one coin: incovertable",
			cf:     newCF(10, "banana", 5, "apple"),
			msgFee: &MsgFee{MsgTypeUrl: "ghijk", Cost: cz(coin(75, "apple"))},
			exp:    &MsgFee{MsgTypeUrl: "ghijk", Cost: cz(coin(75, "apple"))},
		},
		{
			name:   "two coins: one is convertable",
			cf:     newCF(10, "banana", 5, "apple"),
			msgFee: &MsgFee{MsgTypeUrl: "elmeno", Cost: cz(coin(58, "banana"), coin(75, "plum"))},
			exp:    &MsgFee{MsgTypeUrl: "elmeno", Cost: cz(coin(29, "apple"), coin(75, "plum"))},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act *MsgFee
			testFunc := func() {
				act = tc.cf.ConvertMsgFee(tc.msgFee)
			}
			require.NotPanics(t, testFunc, "%s ConvertMsgFee(%s)", tc.cf, tc.msgFee)
			assert.Equal(t, tc.exp, act, "%s ConvertMsgFee(%s) result\nExpected: %s\n  Actual: %s",
				tc.cf, tc.msgFee, tc.exp, act)
		})
	}
}
