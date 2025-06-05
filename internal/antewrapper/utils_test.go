package antewrapper

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1b1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/group"

	"github.com/provenance-io/provenance/internal"
	"github.com/provenance-io/provenance/testutil/assertions"
)

func TestGetGasWanted(t *testing.T) {
	tests := []struct {
		name      string
		feeTx     sdk.FeeTx
		expGas    uint64
		expErr    string
		expInLogs []string
	}{
		{
			name:   "no gas",
			feeTx:  NewMockFeeTx("Martha").WithGas(0),
			expGas: DefaultGasLimit,
			expInLogs: []string{
				"No gas limit provided. Using default.", "returning=500000",
				"method=GetGasWanted", "feeTx.GetGas()=0",
			},
		},
		{
			name:   "no nhash in fee",
			feeTx:  NewMockFeeTx("Frankie").WithGas(74).WithFeeStr(t, "7banana"),
			expGas: 74,
			expInLogs: []string{
				"No nhash in fee. Using provided gas limit.", "returning=74",
				"method=GetGasWanted", "feeTx.GetGas()=74", "feeTx.GetFee()=7banana",
			},
		},
		{
			name:   "gas equals nhash fee amount: 1",
			feeTx:  NewMockFeeTx("Sam").WithGas(1).WithFeeStr(t, "1nhash"),
			expGas: DefaultGasLimit,
			expInLogs: []string{
				"Gas limit equals fee amount. Using default gas limit.", "returning=500000",
				"method=GetGasWanted", "feeTx.GetGas()=1", "feeTx.GetFee()=1nhash",
			},
		},
		{
			name:   "gas equals nhash fee amount: 1,500,000,000",
			feeTx:  NewMockFeeTx("Sam").WithGas(1_500_000_000).WithFeeStr(t, "1500000000nhash"),
			expGas: DefaultGasLimit,
			expInLogs: []string{
				"Gas limit equals fee amount. Using default gas limit.", "returning=500000",
				"method=GetGasWanted", "feeTx.GetGas()=1500000000", "feeTx.GetFee()=1500000000nhash",
			},
		},
		{
			name:   "old gas-prices used: mainnet",
			feeTx:  NewMockFeeTx("Sam").WithGas(5).WithFeeStr(t, "9525nhash"),
			expGas: math.MaxUint64,
			expErr: "old gas-prices value detected; always use 1nhash",
			expInLogs: []string{
				"Gas limit indicates old gas-prices value. Using max uint64.", "returning=18446744073709551615",
				"method=GetGasWanted", "feeTx.GetGas()=5", "feeTx.GetFee()=9525nhash",
			},
		},
		{
			name:   "old gas-prices used: testnet",
			feeTx:  NewMockFeeTx("Billy").WithGas(5).WithFeeStr(t, "95250nhash"),
			expGas: math.MaxUint64,
			expErr: "old gas-prices value detected; always use 1nhash",
			expInLogs: []string{
				"Gas limit indicates old gas-prices value. Using max uint64.", "returning=18446744073709551615",
				"method=GetGasWanted", "feeTx.GetGas()=5", "feeTx.GetFee()=95250nhash",
			},
		},
		{
			name:   "other",
			feeTx:  NewMockFeeTx("Chris").WithGas(64).WithFeeStr(t, "1000000000nhash"),
			expGas: 64,
			expInLogs: []string{
				"Using provided gas limit.", "returning=64",
				"method=GetGasWanted", "feeTx.GetGas()=64", "feeTx.GetFee()=1000000000nhash",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buffer bytes.Buffer
			logger := internal.NewBufferedDebugLogger(&buffer)

			var gas uint64
			var err error
			testFunc := func() {
				gas, err = GetGasWanted(logger, tc.feeTx)
			}
			require.NotPanics(t, testFunc, "GetGasWanted")
			assertions.AssertErrorValue(t, err, tc.expErr, "GetGasWanted error")
			assert.Equal(t, tc.expGas, gas, "GetGasWanted result")

			if t.Failed() {
				return
			}

			logged := buffer.String()
			if len(tc.expInLogs) == 0 {
				assert.Empty(t, logged, "log contents")
			} else {
				for _, exp := range tc.expInLogs {
					assert.Contains(t, logged, exp, "log contents\nExpected: %q", exp)
				}
			}
		})
	}
}

func TestIsOldGasPrices(t *testing.T) {
	testGasses := []int64{
		// Chosen specifically.
		0, 1, 2, 5, 10, int64(DefaultGasLimit), int64(TxGasLimit),
		// Chosen via random number generator
		355, 854, 955,
		5390, 8651, 9146,
		13_071, 44_062, 92_093,
		187_999, 238_337, 696_213,
		1_117_028, 2_027_998, 3_756_542,
	}

	type testCase struct {
		nhash sdkmath.Int
		gas   sdkmath.Int
		exp   bool
	}
	var tests []testCase

	mults := []sdkmath.Int{sdkmath.NewInt(1905), sdkmath.NewInt(19050)}
	for _, gas := range testGasses {
		gasInt := sdkmath.NewInt(gas)
		for _, mult := range mults {
			nhashInt := gasInt.Mul(mult)
			tests = append(tests,
				testCase{gas: gasInt, nhash: nhashInt, exp: true},
				testCase{gas: gasInt.AddRaw(-1), nhash: nhashInt, exp: false},
				testCase{gas: gasInt.AddRaw(1), nhash: nhashInt, exp: false},
				testCase{gas: gasInt, nhash: nhashInt.AddRaw(-1), exp: false},
				testCase{gas: gasInt, nhash: nhashInt.AddRaw(1), exp: false},
			)
			if gas == 0 {
				break
			}
			tests = append(tests,
				testCase{gas: gasInt.Neg(), nhash: nhashInt, exp: false},
				testCase{gas: gasInt, nhash: nhashInt.Neg(), exp: false},
				testCase{gas: gasInt.Neg(), nhash: nhashInt.Neg(), exp: true},
			)
		}
		if gas != 0 {
			tests = append(tests, testCase{gas: gasInt, nhash: gasInt, exp: false})
		}
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%t=%s, %s", tc.exp, tc.nhash, tc.gas), func(t *testing.T) {
			var act bool
			testFunc := func() {
				act = isOldGasPrices(tc.nhash, tc.gas)
			}
			require.NotPanics(t, testFunc, "isOldGasPrices(%s, %s)", tc.nhash, tc.gas)
			assert.Equal(t, tc.exp, act, "isOldGasPrices(%s, %s) result", tc.nhash, tc.gas)
		})
	}
}

func TestTxGasLimitShouldApply(t *testing.T) {
	tests := []struct {
		name    string
		chainID string
		msgs    []sdk.Msg
		exp     bool
	}{
		{
			name:    "test chain, only gov props",
			chainID: SimAppChainID,
			msgs:    []sdk.Msg{&govv1.MsgSubmitProposal{}},
			exp:     false,
		},
		{
			name:    "other chain, only gov props",
			chainID: "dkjfe7iwxx",
			msgs:    []sdk.Msg{&govv1.MsgSubmitProposal{}},
			exp:     false,
		},
		{
			name:    "test chain, other msgs",
			chainID: "testchain-71",
			msgs:    []sdk.Msg{&banktypes.MsgSend{}},
			exp:     false,
		},
		{
			name:    "mainnet, other msgs",
			chainID: "pio-mainnet-1",
			msgs:    []sdk.Msg{&banktypes.MsgSend{}},
			exp:     true,
		},
		{
			name:    "testnet, other msgs",
			chainID: "pio-testnet-1",
			msgs:    []sdk.Msg{&banktypes.MsgSend{}},
			exp:     true,
		},
		{
			name:    "other chain, other msgs",
			chainID: "coq9jwn3ef",
			msgs:    []sdk.Msg{&banktypes.MsgSend{}},
			exp:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act bool
			testFunc := func() {
				act = txGasLimitShouldApply(tc.chainID, tc.msgs)
			}
			require.NotPanics(t, testFunc, "txGasLimitShouldApply(%q, [%d]", tc.chainID, len(tc.msgs))
			assert.Equal(t, tc.exp, act, "txGasLimitShouldApply(%q, [%d]", tc.chainID, len(tc.msgs))
		})
	}
}

func TestIsTestChainID(t *testing.T) {
	tests := []struct {
		chainID string
		exp     bool
	}{
		{exp: true, chainID: ""},
		{exp: true, chainID: "simapp-unit-testing"}, // SimAppChainID
		{exp: true, chainID: "simulation-app"},      // pioconfig.SimAppChainID
		{exp: true, chainID: "testchain"},
		{exp: true, chainID: "testchain-xyz"},
		{exp: true, chainID: "testchain5"},
		{exp: false, chainID: "test-chain"},
		{exp: false, chainID: "x"},
		{exp: false, chainID: "simapp-vnit-testing"},
		{exp: false, chainID: "simvlation-app"},
		{exp: false, chainID: "pio-testnet-1"},
		{exp: false, chainID: "pio-mainnet-1"},
	}

	for _, tc := range tests {
		name := tc.chainID
		if len(name) == 0 {
			name = "empty"
		}
		t.Run(name, func(t *testing.T) {
			var act bool
			testFunc := func() {
				act = isTestChainID(tc.chainID)
			}
			require.NotPanics(t, testFunc, "isTestChainID(%q)", tc.chainID)
			assert.Equal(t, tc.exp, act, "isTestChainID(%q) result", tc.chainID)
		})
	}
}

func TestIsOnlyGovProps(t *testing.T) {
	tests := []struct {
		name string
		msgs []sdk.Msg
		exp  bool
	}{
		{name: "nil", msgs: nil, exp: false},
		{name: "empty", msgs: []sdk.Msg{}, exp: false},

		{
			name: "1 msg: gov v1 submit prop",
			msgs: []sdk.Msg{&govv1.MsgSubmitProposal{}},
			exp:  true,
		},
		{
			name: "1 msg: gov v1 beta1 submit prop",
			msgs: []sdk.Msg{&govv1b1.MsgSubmitProposal{}},
			exp:  true,
		},
		{
			name: "1 msg: bank send",
			msgs: []sdk.Msg{&banktypes.MsgSend{}},
			exp:  false,
		},

		{
			name: "2 msgs: both gov v1 submit prop",
			msgs: []sdk.Msg{&govv1.MsgSubmitProposal{}, &govv1.MsgSubmitProposal{}},
			exp:  true,
		},
		{
			name: "2 msgs: both gov v1beta1 submit prop",
			msgs: []sdk.Msg{&govv1b1.MsgSubmitProposal{}, &govv1b1.MsgSubmitProposal{}},
			exp:  true,
		},
		{
			name: "2 msgs: both bank send",
			msgs: []sdk.Msg{&banktypes.MsgSend{}, &banktypes.MsgSend{}},
			exp:  false,
		},

		{
			name: "2 msgs: v1 prop, v1beta1 prop",
			msgs: []sdk.Msg{&govv1.MsgSubmitProposal{}, &govv1b1.MsgSubmitProposal{}},
			exp:  true,
		},
		{
			name: "2 msgs: v1beta1 prop, v1 prop",
			msgs: []sdk.Msg{&govv1b1.MsgSubmitProposal{}, &govv1.MsgSubmitProposal{}},
			exp:  true,
		},
		{
			name: "2 msgs: v1 prop, bank send",
			msgs: []sdk.Msg{&govv1.MsgSubmitProposal{}, &banktypes.MsgSend{}},
			exp:  false,
		},
		{
			name: "2 msgs: bank send, v1 prop",
			msgs: []sdk.Msg{&banktypes.MsgSend{}, &govv1.MsgSubmitProposal{}},
			exp:  false,
		},
		{
			name: "2 msgs: v1beta1 prop, bank send",
			msgs: []sdk.Msg{&govv1b1.MsgSubmitProposal{}, &banktypes.MsgSend{}},
			exp:  false,
		},
		{
			name: "2 msgs: bank send, v1beta1 prop",
			msgs: []sdk.Msg{&banktypes.MsgSend{}, &govv1b1.MsgSubmitProposal{}},
			exp:  false,
		},
		{
			name: "10 msgs: all gov props",
			msgs: []sdk.Msg{
				&govv1.MsgSubmitProposal{}, &govv1b1.MsgSubmitProposal{},
				&govv1b1.MsgSubmitProposal{}, &govv1.MsgSubmitProposal{},
				&govv1.MsgSubmitProposal{}, &govv1b1.MsgSubmitProposal{},
				&govv1.MsgSubmitProposal{}, &govv1.MsgSubmitProposal{},
				&govv1b1.MsgSubmitProposal{}, &govv1b1.MsgSubmitProposal{},
			},
			exp: true,
		},
		{
			name: "10 msgs: all gov props except one",
			msgs: []sdk.Msg{
				&govv1.MsgSubmitProposal{}, &govv1b1.MsgSubmitProposal{},
				&govv1b1.MsgSubmitProposal{}, &govv1.MsgSubmitProposal{},
				&govv1.MsgSubmitProposal{}, &govv1b1.MsgSubmitProposal{},
				&banktypes.MsgSend{}, &govv1.MsgSubmitProposal{},
				&govv1b1.MsgSubmitProposal{}, &govv1b1.MsgSubmitProposal{},
			},
			exp: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act bool
			testFunc := func() {
				act = isOnlyGovProps(tc.msgs)
			}
			require.NotPanics(t, testFunc, "isOnlyGovProps(%#v)", tc.msgs)
			assert.Equal(t, tc.exp, act, "isOnlyGovProps(%#v) result", tc.msgs)
		})
	}
}

func TestIsGovProp(t *testing.T) {
	tests := []struct {
		msg sdk.Msg
		exp bool
	}{
		{exp: true, msg: &govv1.MsgSubmitProposal{}},
		{exp: true, msg: &govv1b1.MsgSubmitProposal{}},
		{exp: false, msg: &govv1.MsgCancelProposal{}},
		{exp: false, msg: &govv1.MsgVote{}},
		{exp: false, msg: &group.MsgSubmitProposal{}},
		{exp: false, msg: &banktypes.MsgSend{}},
	}

	for _, tc := range tests {
		t.Run(sdk.MsgTypeURL(tc.msg), func(t *testing.T) {
			var act bool
			testFunc := func() {
				act = isGovProp(tc.msg)
			}
			require.NotPanics(t, testFunc, "isGovProp(%q)", tc.msg)
			assert.Equal(t, tc.exp, act, "isGovProp(%q) result", tc.msg)
		})
	}
}

func TestValidateFeeAmount(t *testing.T) {
	cz := func(coins string) sdk.Coins {
		// Not using ParseCoinNormalized or sdk.NewCoin (etc.) so that illegal coins can be made.
		if len(coins) == 0 {
			return nil
		}
		coinStrs := strings.Split(coins, ",")
		rv := make(sdk.Coins, len(coinStrs))
		for n, coin := range coinStrs {
			if len(coin) == 0 || coin == "nil" {
				continue
			}
			var amtStr, denom string
			for i, c := range coin {
				if !unicode.IsDigit(c) && !(i == 0 && c == '-') {
					amtStr = coin[:i]
					denom = coin[i:]
					break
				}
			}
			amt, ok := sdkmath.NewIntFromString(amtStr)
			require.True(t, ok, "sdkmath.NewIntFromString(%q) from %q and %q", amtStr, coin, coins)
			rv[n] = sdk.Coin{Denom: denom, Amount: amt}
		}
		return rv
	}

	tests := []struct {
		name     string
		required sdk.Coins
		provided sdk.Coins
		expErr   string
	}{
		{
			name:     "invalid provided: negative amount",
			required: cz("12nhash"),
			provided: cz("13acorn,-5banana,12cherry"),
			expErr:   "fee provided \"13acorn,-5banana,12cherry\" is invalid: coin banana amount is not positive: insufficient fee",
		},
		{
			name:     "invalid provided: nil coin",
			required: cz("12nhash"),
			provided: cz("13acorn,,12cherry"),
			expErr:   "fee provided \"13acorn,<nil>,12cherry\" is invalid: invalid denom: : insufficient fee",
		},
		{name: "nil, nil", required: nil, provided: nil},
		{name: "nil, empty", required: nil, provided: sdk.Coins{}},
		{name: "empty, nil", required: sdk.Coins{}, provided: nil},
		{name: "empty, empty", required: sdk.Coins{}, provided: sdk.Coins{}},
		{name: "nil, 1 coin", required: nil, provided: cz("5nhash")},
		{name: "empty, 1 coin", required: sdk.Coins{}, provided: cz("5nhash")},

		{
			name:     "required 1 coin: nil provided",
			required: cz("7apple"),
			provided: nil,
			expErr:   "fee required: \"7apple\", fee provided: \"\", short by \"7apple\": insufficient fee",
		},
		{
			name:     "required 1 coin: empty provided",
			required: cz("6apple"),
			provided: sdk.Coins{},
			expErr:   "fee required: \"6apple\", fee provided: \"\", short by \"6apple\": insufficient fee",
		},
		{
			name:     "required 1 coin: provided other denom",
			required: cz("6apple"),
			provided: cz("3plum"),
			expErr:   "fee required: \"6apple\", fee provided: \"3plum\", short by \"6apple\": insufficient fee",
		},
		{
			name:     "required 1 coin: provided is less",
			required: cz("12345apple"),
			provided: cz("12344apple"),
			expErr:   "fee required: \"12345apple\", fee provided: \"12344apple\", short by \"1apple\": insufficient fee",
		},
		{
			name:     "required 1 coin: provided is same",
			required: cz("400apple"),
			provided: cz("400apple"),
		},
		{
			name:     "required 1 coin: provided is same plus a zero coin",
			required: cz("400apple"),
			provided: cz("400apple,0banana"),
		},
		{
			name:     "required 1 coin: provided is same with other coins",
			required: cz("400apple"),
			provided: cz("400apple,12banana,15plum"),
		},
		{
			name:     "required 1 coin: provided is more",
			required: cz("99banana"),
			provided: cz("100banana"),
		},
		{
			name:     "required 1 coin: provided is more plus a zero coin",
			required: cz("99banana"),
			provided: cz("0apple,100banana"),
		},
		{
			name:     "required 1 coin: provided is more plus other coins",
			required: cz("99banana"),
			provided: cz("4apple,100banana,77cherry"),
		},

		{
			name:     "required 2 coins: provided has neither denom",
			required: cz("1apple,2banana"),
			provided: cz("1cherry,2durian"),
			expErr:   "fee required: \"1apple,2banana\", fee provided: \"1cherry,2durian\", short by \"1apple,2banana\": insufficient fee",
		},
		{
			name:     "required 2 coins: provided only first",
			required: cz("5apple,12banana"),
			provided: cz("5apple"),
			expErr:   "fee required: \"5apple,12banana\", fee provided: \"5apple\", short by \"12banana\": insufficient fee",
		},
		{
			name:     "required 2 coins: provided only second",
			required: cz("5apple,12banana"),
			provided: cz("12banana"),
			expErr:   "fee required: \"5apple,12banana\", fee provided: \"12banana\", short by \"5apple\": insufficient fee",
		},
		{
			name:     "required 2 coins: provided less on both",
			required: cz("12apple,73banana"),
			provided: cz("11apple,72banana"),
			expErr:   "fee required: \"12apple,73banana\", fee provided: \"11apple,72banana\", short by \"1apple,1banana\": insufficient fee",
		},
		{
			name:     "required 2 coins: provided less, same",
			required: cz("12apple,73banana"),
			provided: cz("11apple,73banana"),
			expErr:   "fee required: \"12apple,73banana\", fee provided: \"11apple,73banana\", short by \"1apple\": insufficient fee",
		},
		{
			name:     "required 2 coins: provided less, more",
			required: cz("12apple,73banana"),
			provided: cz("11apple,74banana"),
			expErr:   "fee required: \"12apple,73banana\", fee provided: \"11apple,74banana\", short by \"1apple\": insufficient fee",
		},
		{
			name:     "required 2 coins: provided same, less",
			required: cz("12apple,73banana"),
			provided: cz("12apple,72banana"),
			expErr:   "fee required: \"12apple,73banana\", fee provided: \"12apple,72banana\", short by \"1banana\": insufficient fee",
		},
		{
			name:     "required 2 coins: provided same on both",
			required: cz("12apple,73banana"),
			provided: cz("12apple,73banana"),
		},
		{
			name:     "required 2 coins: provided same on both plus a zero coin",
			required: cz("12apple,73banana"),
			provided: cz("12apple,73banana,0cherry"),
		},
		{
			name:     "required 2 coins: provided same on both plus other coins",
			required: cz("12apple,73banana"),
			provided: cz("12apple,73banana,5cherry,31durian"),
		},
		{
			name:     "required 2 coins: provided same, more",
			required: cz("12apple,73banana"),
			provided: cz("12apple,74banana"),
		},
		{
			name:     "required 2 coins: provided more, less",
			required: cz("12apple,73banana"),
			provided: cz("13apple,72banana"),
			expErr:   "fee required: \"12apple,73banana\", fee provided: \"13apple,72banana\", short by \"1banana\": insufficient fee",
		},
		{
			name:     "required 2 coins: provided more, same",
			required: cz("12apple,73banana"),
			provided: cz("13apple,73banana"),
		},
		{
			name:     "required 2 coins: provided more on both",
			required: cz("12apple,73banana"),
			provided: cz("13apple,74banana"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = validateFeeAmount(tc.required, tc.provided)
			}
			require.NotPanics(t, testFunc, "validateFeeAmount(%s, %s)", tc.required, tc.provided)
			assertions.AssertErrorValue(t, err, tc.expErr, "validateFeeAmount(%s, %s) error", tc.required, tc.provided)
		})
	}
}

func TestGetFeePayerUsingFeeGrant(t *testing.T) {
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "ParseCoinsNormalized(%q)", coins)
		return rv
	}

	addr := func(base string) sdk.AccAddress {
		if len(base) < 20 {
			base += strings.Repeat("_", 20-len(base))
		}
		return sdk.AccAddress(base)
	}

	tests := []struct {
		name     string
		fk       *MockFeegrantKeeper
		feeTx    sdk.FeeTx
		amount   sdk.Coins
		msgs     []sdk.Msg
		expPayer sdk.AccAddress
		expUsed  bool
		expErr   string
		expCall  bool
	}{
		{
			name:     "granter is nil",
			fk:       NewMockFeegrantKeeper(),
			feeTx:    NewMockFeeTx("a").WithFeeGranter(nil).WithFeePayer(addr("payer")),
			amount:   cz("5nhash"),
			msgs:     []sdk.Msg{&banktypes.MsgSend{}},
			expPayer: addr("payer"),
			expUsed:  false,
		},
		{
			name:     "amount zero",
			fk:       NewMockFeegrantKeeper(),
			feeTx:    NewMockFeeTx("a").WithFeeGranter(addr("granter")).WithFeePayer(addr("payer")),
			amount:   nil,
			msgs:     []sdk.Msg{&banktypes.MsgSend{}},
			expPayer: addr("granter"),
			expUsed:  true,
		},
		{
			name:   "error from UseGrantedFees",
			fk:     NewMockFeegrantKeeper().WithUseGrantedFees("kaplow"),
			feeTx:  NewMockFeeTx("a").WithFeeGranter(addr("granter")).WithFeePayer(addr("payer")),
			amount: cz("1000000000nhash"),
			msgs:   []sdk.Msg{&banktypes.MsgSend{}, &govv1.MsgVote{}},
			expErr: "failed to use fee grant: " +
				"granter: " + addr("granter").String() + ", " +
				"grantee: " + addr("payer").String() + ", " +
				"fee: \"1000000000nhash\", " +
				"msgs: [\"/cosmos.bank.v1beta1.MsgSend\" \"/cosmos.gov.v1.MsgVote\"]" +
				": kaplow",
			expCall: true,
		},
		{
			name:     "grant used",
			fk:       NewMockFeegrantKeeper(),
			feeTx:    NewMockFeeTx("a").WithFeeGranter(addr("granter")).WithFeePayer(addr("payer")),
			amount:   cz("1000000000nhash"),
			msgs:     []sdk.Msg{&banktypes.MsgSend{}},
			expPayer: addr("granter"),
			expUsed:  true,
			expCall:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var expCall *UseGrantedFeesArgs
			if tc.expCall {
				expCall = NewUseGrantedFeesArgs(tc.feeTx.FeeGranter(), tc.feeTx.FeePayer(), tc.amount, tc.msgs)
			}

			var actPayer sdk.AccAddress
			var actUsed bool
			var actErr error
			testFunc := func() {
				actPayer, actUsed, actErr = getFeePayerUsingFeeGrant(sdk.Context{}, tc.fk, tc.feeTx, tc.amount, tc.msgs)
			}
			require.NotPanics(t, testFunc, "getFeePayerUsingFeeGrant")
			assertions.AssertErrorValue(t, actErr, tc.expErr, "getFeePayerUsingFeeGrant error")
			assert.Equal(t, tc.expPayer, actPayer, "getFeePayerUsingFeeGrant payer")
			assert.Equal(t, tc.expUsed, actUsed, "getFeePayerUsingFeeGrant used bool")
			if tc.fk != nil {
				tc.fk.AssertUseGrantedFeesCall(t, expCall)
			}
		})
	}
}

func TestPayFee(t *testing.T) {
	tests := []struct {
		name    string
		bk      *MockBankKeeper
		addr    sdk.AccAddress
		fee     sdk.Coins
		expErr  string
		expCall bool
	}{
		{
			name: "nil fee",
			bk:   NewMockBankKeeper(),
			addr: sdk.AccAddress("zero_fee_addr_______"),
			fee:  nil,
		},
		{
			name: "empty fee",
			bk:   NewMockBankKeeper(),
			addr: sdk.AccAddress("empty_fee_addr______"),
			fee:  sdk.Coins{},
		},
		{
			name: "fee with only a zero coin",
			bk:   NewMockBankKeeper(),
			addr: sdk.AccAddress("zero_coin_addr______"),
			fee:  sdk.Coins{sdk.Coin{Denom: "nhash", Amount: sdkmath.ZeroInt()}},
		},
		{
			name:    "invalid fee",
			bk:      NewMockBankKeeper(),
			addr:    sdk.AccAddress("invalid_fee_addr____"),
			fee:     sdk.Coins{sdk.Coin{Denom: "nhash", Amount: sdkmath.NewInt(-3)}},
			expErr:  "invalid fee amount: -3nhash: insufficient fee",
			expCall: false,
		},
		{
			name:    "error sending",
			bk:      NewMockBankKeeper().WithSendCoinsFromAccountToModule("blamo"),
			addr:    sdk.AccAddress("error_sending_addr__"),
			fee:     sdk.NewCoins(sdk.NewInt64Coin("plum", 5)),
			expErr:  "blamo: account: " + sdk.AccAddress("error_sending_addr__").String() + ": insufficient funds",
			expCall: true,
		},
		{
			name:    "okay",
			bk:      NewMockBankKeeper(),
			addr:    sdk.AccAddress("okay_addr___________"),
			fee:     sdk.NewCoins(sdk.NewInt64Coin("plum", 5)),
			expCall: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var expCall *SendCoinsFromAccountToModuleArgs
			if tc.expCall {
				expCall = NewSendCoinsFromAccountToModuleArgs(tc.addr, authtypes.FeeCollectorName, tc.fee)
			}

			var err error
			testFunc := func() {
				err = PayFee(sdk.Context{}, tc.bk, tc.addr, tc.fee)
			}
			require.NotPanics(t, testFunc, "PayFee")
			assertions.AssertErrorValue(t, err, tc.expErr, "PayFee error")
			tc.bk.AssertSendCoinsFromAccountToModuleCall(t, expCall)
		})
	}
}

func TestMsgTypeURLs(t *testing.T) {
	tests := []struct {
		name string
		msgs []sdk.Msg
		exp  []string
	}{
		{
			name: "nil",
			msgs: nil,
			exp:  nil,
		},
		{
			name: "empty",
			msgs: []sdk.Msg{},
			exp:  []string{},
		},
		{
			name: "1 msg",
			msgs: []sdk.Msg{&govv1.MsgVote{}},
			exp:  []string{"/cosmos.gov.v1.MsgVote"},
		},
		{
			name: "2 msgs",
			msgs: []sdk.Msg{&banktypes.MsgSend{}, &govv1.MsgDeposit{}},
			exp:  []string{"/cosmos.bank.v1beta1.MsgSend", "/cosmos.gov.v1.MsgDeposit"},
		},
		{
			name: "10 msgs with dups",
			msgs: []sdk.Msg{
				&govv1.MsgUpdateParams{}, &banktypes.MsgSend{},
				&group.MsgSubmitProposal{}, &govv1.MsgSubmitProposal{},
				&banktypes.MsgSend{}, &banktypes.MsgSend{},
				&govv1b1.MsgVoteWeighted{}, &group.MsgExec{},
				&group.MsgExec{}, &banktypes.MsgUpdateParams{},
			},
			exp: []string{
				"/cosmos.gov.v1.MsgUpdateParams", "/cosmos.bank.v1beta1.MsgSend",
				"/cosmos.group.v1.MsgSubmitProposal", "/cosmos.gov.v1.MsgSubmitProposal",
				"/cosmos.bank.v1beta1.MsgSend", "/cosmos.bank.v1beta1.MsgSend",
				"/cosmos.gov.v1beta1.MsgVoteWeighted", "/cosmos.group.v1.MsgExec",
				"/cosmos.group.v1.MsgExec", "/cosmos.bank.v1beta1.MsgUpdateParams",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act []string
			testFunc := func() {
				act = msgTypeURLs(tc.msgs)
			}
			require.NotPanics(t, testFunc, "msgTypeURLs")
			assert.Equal(t, tc.exp, act, "msgTypeURLs result")
		})
	}
}

func TestGetFeeTx(t *testing.T) {
	tests := []struct {
		name   string
		tx     sdk.Tx
		expErr string
	}{
		{
			name:   "nil",
			tx:     nil,
			expErr: "Tx must be a FeeTx: <nil>: tx parse error",
		},
		{
			name:   "not a fee tx",
			tx:     NewNotFeeTx("oops"),
			expErr: "Tx must be a FeeTx: *antewrapper.NotFeeTx: tx parse error",
		},
		{
			name: "fee tx",
			tx:   NewMockFeeTx("yay"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var feeTx sdk.FeeTx
			var err error
			testFunc := func() {
				feeTx, err = GetFeeTx(tc.tx)
			}
			require.NotPanics(t, testFunc, "GetFeeTx")
			assertions.AssertErrorValue(t, err, tc.expErr, "GetFeeTx error")
			if len(tc.expErr) == 0 && err == nil {
				assert.NotNil(t, feeTx, "GetFeeTx result")
				// The easiest way I could think of to make sure that the feeTx returned is the one provided
				// is to give them an ID and a String() that returns that id. But the FeeTx interface does not
				// have a stringer, so I'm using sprintf to invoke .String() on each of them.
				assert.Equal(t, fmt.Sprintf("%s", tc.tx), fmt.Sprintf("%s", feeTx))
			}
		})
	}
}

func TestIsInitGenesis(t *testing.T) {
	tests := []struct {
		name string
		ctx  sdk.Context
		exp  bool
	}{
		{
			name: "zero-val context",
			ctx:  sdk.Context{},
			exp:  true,
		},
		{
			name: "block height = 0",
			ctx:  sdk.Context{}.WithBlockHeight(0),
			exp:  true,
		},
		{
			name: "block height = 1",
			ctx:  sdk.Context{}.WithBlockHeight(1),
			exp:  false,
		},
		{
			name: "block height = 2",
			ctx:  sdk.Context{}.WithBlockHeight(2),
			exp:  false,
		},
		{
			name: "block height = -1",
			ctx:  sdk.Context{}.WithBlockHeight(-1),
			exp:  true,
		},
		{
			name: "block height = 24253450",
			ctx:  sdk.Context{}.WithBlockHeight(24253450),
			exp:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act bool
			testFunc := func() {
				act = isInitGenesis(tc.ctx)
			}
			require.NotPanics(t, testFunc, "isInitGenesis")
			assert.Equal(t, tc.exp, act, "isInitGenesis result")
		})
	}
}
