package types

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestSplitCoinByPercentage(t *testing.T) {
	splitCases := []struct {
		testName              string
		splitCoin             sdk.Coin
		splitPercent          uint32
		expectedRecipientCoin sdk.Coin
		expectedFeePayoutCoin sdk.Coin
	}{
		{
			"Both Recipient and FeePayout should equal on split of even number",
			sdk.NewInt64Coin(NhashDenom, 10),
			5_000,
			sdk.NewInt64Coin(NhashDenom, 5),
			sdk.NewInt64Coin(NhashDenom, 5),
		},
		{
			"Recipient will get floor of calc",
			sdk.NewInt64Coin(NhashDenom, 9),
			5_000,
			sdk.NewInt64Coin(NhashDenom, 4),
			sdk.NewInt64Coin(NhashDenom, 5),
		},
		{
			"FeePayout should get the remaining amount",
			sdk.NewInt64Coin(NhashDenom, 1),
			5_000,
			sdk.NewInt64Coin(NhashDenom, 0),
			sdk.NewInt64Coin(NhashDenom, 1),
		},
		{
			"Recipient should receive bips amount calculated to 25",
			sdk.NewInt64Coin(NhashDenom, 100),
			2_500,
			sdk.NewInt64Coin(NhashDenom, 25),
			sdk.NewInt64Coin(NhashDenom, 75),
		},
		{
			"Recipient should receive bips amount calculated to 25 with truncation remainder going to fee module",
			sdk.NewInt64Coin(NhashDenom, 101),
			2_500,
			sdk.NewInt64Coin(NhashDenom, 25),
			sdk.NewInt64Coin(NhashDenom, 76),
		},
	}

	for _, tc := range splitCases {
		actualRecipientCoin, actualFeePayoutCoin := SplitCoinByPercentage(tc.splitCoin, tc.splitPercent)
		assert.Equal(t, tc.expectedRecipientCoin, actualRecipientCoin, fmt.Sprintf("RecipientCoin not equal: %s", tc.testName))
		assert.Equal(t, tc.expectedFeePayoutCoin, actualFeePayoutCoin, fmt.Sprintf("FeePayoutCoin not equal: %s", tc.testName))
	}
}
