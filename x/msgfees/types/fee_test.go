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
			"FeePayout should have floor go to fee payout on odd number",
			sdk.NewInt64Coin(NhashDenom, 9),
			5_000,
			sdk.NewInt64Coin(NhashDenom, 5),
			sdk.NewInt64Coin(NhashDenom, 4),
		},
		{
			"FeePayout should have receive 0 on split of 1",
			sdk.NewInt64Coin(NhashDenom, 1),
			5_000,
			sdk.NewInt64Coin(NhashDenom, 1),
			sdk.NewInt64Coin(NhashDenom, 0),
		},
	}

	for _, tc := range splitCases {
		actualRecipientCoin, actualFeePayoutCoin := SplitCoinByPercentage(tc.splitCoin, tc.splitPercent)
		assert.Equal(t, tc.expectedRecipientCoin, actualRecipientCoin, fmt.Sprintf("RecipientCoin not equal: %s", tc.testName))
		assert.Equal(t, tc.expectedFeePayoutCoin, actualFeePayoutCoin, fmt.Sprintf("FeePayoutCoin not equal: %s", tc.testName))
	}
}
