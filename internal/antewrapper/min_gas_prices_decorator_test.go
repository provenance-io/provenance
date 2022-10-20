package antewrapper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"testing"
)

type TestTerminator struct {
	isTerminated bool
	ctx          sdk.Context
	tx           sdk.Tx
	simulate     bool
}

func NewTestTerminator() *TestTerminator {
	return &TestTerminator{}
}

func (t *TestTerminator) AnteHandler(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
	t.isTerminated = true
	t.ctx = ctx
	t.tx = tx
	t.simulate = simulate
	return ctx, nil
}

var _ sdk.Tx = &NonFeeTx{}

type NonFeeTx struct{}

func (t NonFeeTx) GetMsgs() []sdk.Msg {
	return nil
}

func (t NonFeeTx) ValidateBasic() error {
	return nil
}

func NewFeeTx(gasLimit uint64, fee sdk.Coins) sdk.Tx {
	return &txtypes.Tx{
		AuthInfo: &txtypes.AuthInfo{
			Fee: &txtypes.Fee{
				Amount:   fee,
				GasLimit: gasLimit,
			},
		},
	}
}

func TestAnteHandle(tt *testing.T) {
	var dummyTx sdk.Tx
	dummyTx = &NonFeeTx{}
	_, ok := dummyTx.(sdk.FeeTx)
	require.False(tt, ok, "NonFeeTx should not implement FeeTx.")

	testSkipMinGasPrices := sdk.NewDecCoins(sdk.NewInt64DecCoin("simfoo", 1000))
	testSkipGas := uint64(5)
	testSkipFee := sdk.NewCoins(sdk.NewInt64Coin("simfoo", 4999))

	tests := []struct {
		name            string
		simulate        bool
		isCheckTx       bool
		minGasPrices    sdk.DecCoins
		tx              sdk.Tx
		expectedInError []string
	}{
		// These three tests demonstrate that the check is properly skipped when it should be.
		// They should have the same minGasPrices
		{
			name:            "skip because simulating",
			simulate:        true,
			isCheckTx:       true,
			minGasPrices:    testSkipMinGasPrices,
			tx:              NewFeeTx(testSkipGas, testSkipFee),
			expectedInError: nil,
		},
		{
			name:            "skip because not isCheckTx",
			simulate:        false,
			isCheckTx:       false,
			minGasPrices:    testSkipMinGasPrices,
			tx:              NewFeeTx(testSkipGas, testSkipFee),
			expectedInError: nil,
		},
		{
			name:            "skip fails when not skipping",
			simulate:        false,
			isCheckTx:       true,
			minGasPrices:    testSkipMinGasPrices,
			tx:              NewFeeTx(testSkipGas, testSkipFee),
			expectedInError: []string{"insufficient fee", "min-gas-prices not met", "got: 4999simfoo", "required: 5000simfoo"},
		},
		// end of skip tests.

		{
			name:            "min gas denom not in fee",
			simulate:        false,
			isCheckTx:       true,
			minGasPrices:    sdk.NewDecCoins(sdk.NewInt64DecCoin("gascoin", 100)),
			tx:              NewFeeTx(5, sdk.NewCoins(sdk.NewInt64Coin("feecoin", 500_000))),
			expectedInError: []string{"insufficient fee", "min-gas-prices not met", "got: 500000feecoin", "required: 500gascoin"},
		},
		{
			name:            "two gas denoms only first in fee not enough",
			simulate:        false,
			isCheckTx:       true,
			minGasPrices:    sdk.NewDecCoins(sdk.NewInt64DecCoin("onefoo", 10), sdk.NewInt64DecCoin("twofoo", 100)),
			tx:              NewFeeTx(5, sdk.NewCoins(sdk.NewInt64Coin("onefoo", 49))),
			expectedInError: []string{"insufficient fee", "min-gas-prices not met", "got: 49onefoo", "required: 50onefoo,500twofoo"},
		},
		{
			name:            "two gas denoms only first in fee barely enough",
			simulate:        false,
			isCheckTx:       true,
			minGasPrices:    sdk.NewDecCoins(sdk.NewInt64DecCoin("threefoo", 10), sdk.NewInt64DecCoin("fourfoo", 100)),
			tx:              NewFeeTx(5, sdk.NewCoins(sdk.NewInt64Coin("threefoo", 50))),
			expectedInError: nil,
		},
		{
			name:            "two gas denoms only second in fee not enough",
			simulate:        false,
			isCheckTx:       true,
			minGasPrices:    sdk.NewDecCoins(sdk.NewInt64DecCoin("fivefoo", 10), sdk.NewInt64DecCoin("sixfoo", 100)),
			tx:              NewFeeTx(5, sdk.NewCoins(sdk.NewInt64Coin("sixfoo", 499))),
			expectedInError: []string{"insufficient fee", "min-gas-prices not met", "got: 499sixfoo", "required: 50fivefoo,500sixfoo"},
		},
		{
			name:            "two gas denoms only second in fee barely enough",
			simulate:        false,
			isCheckTx:       true,
			minGasPrices:    sdk.NewDecCoins(sdk.NewInt64DecCoin("sevenfoo", 10), sdk.NewInt64DecCoin("eightfoo", 100)),
			tx:              NewFeeTx(5, sdk.NewCoins(sdk.NewInt64Coin("eightfoo", 500))),
			expectedInError: nil,
		},
		{
			name:            "two gas denoms not enough both",
			simulate:        false,
			isCheckTx:       true,
			minGasPrices:    sdk.NewDecCoins(sdk.NewInt64DecCoin("ninefoo", 10), sdk.NewInt64DecCoin("tenfoo", 100)),
			tx:              NewFeeTx(10, sdk.NewCoins(sdk.NewInt64Coin("ninefoo", 99), sdk.NewInt64Coin("tenfoo", 999))),
			expectedInError: []string{"insufficient fee", "min-gas-prices not met", "got: 99ninefoo,999tenfoo", "required: 100ninefoo,1000tenfoo"},
		},
		{
			name:            "two gas denoms not enough first",
			simulate:        false,
			isCheckTx:       true,
			minGasPrices:    sdk.NewDecCoins(sdk.NewInt64DecCoin("elevenfoo", 10), sdk.NewInt64DecCoin("twelvefoo", 100)),
			tx:              NewFeeTx(10, sdk.NewCoins(sdk.NewInt64Coin("elevenfoo", 99), sdk.NewInt64Coin("twelvefoo", 1000))),
			expectedInError: nil,
		},
		{
			name:            "two gas denoms not enough second",
			simulate:        false,
			isCheckTx:       true,
			minGasPrices:    sdk.NewDecCoins(sdk.NewInt64DecCoin("thirteenfoo", 10), sdk.NewInt64DecCoin("fourteenfoo", 100)),
			tx:              NewFeeTx(10, sdk.NewCoins(sdk.NewInt64Coin("thirteenfoo", 100), sdk.NewInt64Coin("fourteenfoo", 999))),
			expectedInError: nil,
		},
		{
			name:            "one gas denoms two coins in fee not enough",
			simulate:        false,
			isCheckTx:       true,
			minGasPrices:    sdk.NewDecCoins(sdk.NewInt64DecCoin("fifteenfoo", 10)),
			tx:              NewFeeTx(7, sdk.NewCoins(sdk.NewInt64Coin("fifteenfoo", 69), sdk.NewInt64Coin("sixteenfoo", 420))),
			expectedInError: []string{"insufficient fee", "min-gas-prices not met", "got: 69fifteenfoo,420sixteenfoo", "required: 70fifteenfoo"},
		},
		{
			name:            "one gas denoms two coins in fee enough",
			simulate:        false,
			isCheckTx:       true,
			minGasPrices:    sdk.NewDecCoins(sdk.NewInt64DecCoin("seventeenfoo", 10)),
			tx:              NewFeeTx(1, sdk.NewCoins(sdk.NewInt64Coin("seventeenfoo", 420), sdk.NewInt64Coin("eighteenfoo", 69))),
			expectedInError: nil,
		},
		{
			name:            "one gas denom more than enough fee",
			simulate:        false,
			isCheckTx:       true,
			minGasPrices:    sdk.NewDecCoins(sdk.NewInt64DecCoin("nineteenfoo", 50)),
			tx:              NewFeeTx(2, sdk.NewCoins(sdk.NewInt64Coin("nineteenfoo", 1_000_000))),
			expectedInError: nil,
		},
		{
			name:            "min gas zero and no fee",
			simulate:        false,
			isCheckTx:       true,
			minGasPrices:    sdk.NewDecCoins(sdk.NewInt64DecCoin("twentyfoo", 0), sdk.NewInt64DecCoin("twentybar", 0)),
			tx:              NewFeeTx(100, sdk.NewCoins()),
			expectedInError: nil,
		},
		{
			name:            "decimal min gas rounded up",
			simulate:        false,
			isCheckTx:       true,
			minGasPrices:    sdk.NewDecCoins(sdk.NewDecCoinFromDec("pcoin", sdk.MustNewDecFromStr("0.15"))),
			tx:              NewFeeTx(7, sdk.NewCoins(sdk.NewInt64Coin("pcoin", 1))),
			expectedInError: []string{"required: 2pcoin"},
		},

		// Check cases when not provided a FeeTx.
		{
			name:            "non-fee tx simulating isCheckTx",
			simulate:        true,
			isCheckTx:       true,
			minGasPrices:    sdk.NewDecCoins(sdk.NewInt64DecCoin("acoin", 1)),
			tx:              &NonFeeTx{},
			expectedInError: nil,
		},
		{
			name:            "non-fee tx simulating not isCheckTx",
			simulate:        true,
			isCheckTx:       false,
			minGasPrices:    sdk.NewDecCoins(sdk.NewInt64DecCoin("acoin", 1)),
			tx:              &NonFeeTx{},
			expectedInError: nil,
		},
		{
			name:            "non-fee tx not simulating isCheckTx",
			simulate:        false,
			isCheckTx:       true,
			minGasPrices:    sdk.NewDecCoins(sdk.NewInt64DecCoin("acoin", 1)),
			tx:              &NonFeeTx{},
			expectedInError: []string{"tx parse error", "Tx must be a FeeTx"},
		},
		{
			name:            "non-fee tx not simulating not isCheckTx",
			simulate:        false,
			isCheckTx:       false,
			minGasPrices:    sdk.NewDecCoins(sdk.NewInt64DecCoin("acoin", 1)),
			tx:              &NonFeeTx{},
			expectedInError: []string{"tx parse error", "Tx must be a FeeTx"},
		},
	}

	for _, tc := range tests {
		tt.Run(tc.name, func(t *testing.T) {
			ctx := sdk.NewContext(nil, tmproto.Header{}, tc.isCheckTx, nil).WithMinGasPrices(tc.minGasPrices)
			terminator := NewTestTerminator()
			decorator := antewrapper.NewMinGasPricesDecorator()
			newCtx, err := decorator.AnteHandle(ctx, tc.tx, tc.simulate, terminator.AnteHandler)
			// The context should not have changed along the way.
			assert.Equal(t, ctx, newCtx, "newCtx")
			if len(tc.expectedInError) > 0 {
				for _, exp := range tc.expectedInError {
					assert.ErrorContains(t, err, exp)
				}
				// If we were expecting an error, the terminator should not have been called.
				assert.False(t, terminator.isTerminated, "terminator.isTerminated")
			} else {
				assert.NoError(t, err)
				// If we were not expecting an error, the terminator should have been called with
				// the same arguments provided to the AnteHandle.
				assert.True(t, terminator.isTerminated, "terminator.isTerminated")
				assert.Equal(t, ctx, terminator.ctx, "terminator.ctx")
				assert.Equal(t, tc.tx, terminator.tx, "terminator.tx")
				assert.Equal(t, tc.simulate, terminator.simulate, "terminator.simulate")
			}
		})
	}
}
