package types

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/assert"
)

func TestSplitCoinByBips(t *testing.T) {
	splitCases := []struct {
		testName              string
		splitCoin             sdk.Coin
		bips                  uint32
		expectedRecipientCoin sdk.Coin
		expectedFeePayoutCoin sdk.Coin
		expectedErrorMsg      string
	}{
		{
			"Should all go to recipient",
			sdk.NewInt64Coin(NhashDenom, 10),
			10_000,
			sdk.NewInt64Coin(NhashDenom, 10),
			sdk.NewInt64Coin(NhashDenom, 0),
			"",
		},
		{
			"Should all go to fee payout",
			sdk.NewInt64Coin(NhashDenom, 10),
			0,
			sdk.NewInt64Coin(NhashDenom, 0),
			sdk.NewInt64Coin(NhashDenom, 10),
			"",
		},
		{
			"Should error on invalid bips value",
			sdk.NewInt64Coin(NhashDenom, 10),
			10_001,
			sdk.NewInt64Coin(NhashDenom, 0),
			sdk.NewInt64Coin(NhashDenom, 10),
			"invalid: 10001: invalid bips amount",
		},
		{
			"Both Recipient and FeePayout should equal on split of even number",
			sdk.NewInt64Coin(NhashDenom, 10),
			5_000,
			sdk.NewInt64Coin(NhashDenom, 5),
			sdk.NewInt64Coin(NhashDenom, 5),
			"",
		},
		{
			"Recipient will get floor of calc",
			sdk.NewInt64Coin(NhashDenom, 9),
			5_000,
			sdk.NewInt64Coin(NhashDenom, 4),
			sdk.NewInt64Coin(NhashDenom, 5),
			"",
		},
		{
			"FeePayout should get the remaining amount",
			sdk.NewInt64Coin(NhashDenom, 1),
			5_000,
			sdk.NewInt64Coin(NhashDenom, 0),
			sdk.NewInt64Coin(NhashDenom, 1),
			"",
		},
		{
			"Recipient should receive bips amount calculated to 25",
			sdk.NewInt64Coin(NhashDenom, 100),
			2_500,
			sdk.NewInt64Coin(NhashDenom, 25),
			sdk.NewInt64Coin(NhashDenom, 75),
			"",
		},
		{
			"Recipient should receive bips amount calculated to 25 with truncation remainder going to fee module",
			sdk.NewInt64Coin(NhashDenom, 101),
			2_500,
			sdk.NewInt64Coin(NhashDenom, 25),
			sdk.NewInt64Coin(NhashDenom, 76),
			"",
		},
	}

	for _, tc := range splitCases {
		actualRecipientCoin, actualFeePayoutCoin, err := SplitCoinByBips(tc.splitCoin, tc.bips)
		if len(tc.expectedErrorMsg) == 0 {
			assert.NoError(t, err)
			assert.True(t, tc.expectedRecipientCoin.Equal(actualRecipientCoin), fmt.Sprintf("RecipientCoin for bips: %v not equal expected: %s actual: %s name: %s", tc.bips, tc.expectedRecipientCoin.String(), actualRecipientCoin, tc.testName))
			assert.True(t, tc.expectedFeePayoutCoin.Equal(actualFeePayoutCoin), fmt.Sprintf("FeePayoutCoin not equal expected: %s actual: %s name: %s", tc.expectedFeePayoutCoin.String(), actualFeePayoutCoin.String(), tc.testName))
		} else {
			assert.EqualError(t, err, tc.expectedErrorMsg, tc.testName)
		}
	}
}
