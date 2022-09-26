package types

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/internal/pioconfig"

	"github.com/stretchr/testify/assert"
)

func TestSplitCoinByBips(t *testing.T) {
	pioconfig.SetProvenanceConfig("", 0)
	splitCases := []struct {
		name                  string
		splitCoin             sdk.Coin
		bips                  uint32
		expectedRecipientCoin sdk.Coin
		expectedFeePayoutCoin sdk.Coin
		expectedErrorMsg      string
	}{
		{
			name:                  "Should all go to recipient",
			splitCoin:             sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 10),
			bips:                  10_000,
			expectedRecipientCoin: sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 10),
			expectedFeePayoutCoin: sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 0),
			expectedErrorMsg:      "",
		},
		{
			name:                  "Should all go to fee payout",
			splitCoin:             sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 10),
			bips:                  0,
			expectedRecipientCoin: sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 0),
			expectedFeePayoutCoin: sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 10),
			expectedErrorMsg:      "",
		},
		{
			name:                  "Should error on invalid bips value",
			splitCoin:             sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 10),
			bips:                  10_001,
			expectedRecipientCoin: sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 0),
			expectedFeePayoutCoin: sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 10),
			expectedErrorMsg:      "invalid: 10001: invalid bips amount",
		},
		{
			name:                  "Both Recipient and FeePayout should equal on split of even number",
			splitCoin:             sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 10),
			bips:                  5_000,
			expectedRecipientCoin: sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 5),
			expectedFeePayoutCoin: sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 5),
			expectedErrorMsg:      "",
		},
		{
			name:                  "Recipient will get floor of calc",
			splitCoin:             sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 9),
			bips:                  5_000,
			expectedRecipientCoin: sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 4),
			expectedFeePayoutCoin: sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 5),
			expectedErrorMsg:      "",
		},
		{
			name:                  "FeePayout should get the remaining amount",
			splitCoin:             sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 1),
			bips:                  5_000,
			expectedRecipientCoin: sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 0),
			expectedFeePayoutCoin: sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 1),
			expectedErrorMsg:      "",
		},
		{
			name:                  "Recipient should receive bips amount calculated to 25",
			splitCoin:             sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 100),
			bips:                  2_500,
			expectedRecipientCoin: sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 25),
			expectedFeePayoutCoin: sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 75),
			expectedErrorMsg:      "",
		},
		{
			name:                  "Recipient should receive bips amount calculated to 25 with truncation remainder going to fee module",
			splitCoin:             sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 101),
			bips:                  2_500,
			expectedRecipientCoin: sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 25),
			expectedFeePayoutCoin: sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 76),
			expectedErrorMsg:      "",
		},
	}

	for _, tc := range splitCases {
		actualRecipientCoin, actualFeePayoutCoin, err := SplitCoinByBips(tc.splitCoin, tc.bips)
		if len(tc.expectedErrorMsg) == 0 {
			assert.NoError(t, err)
			assert.True(t, tc.expectedRecipientCoin.Equal(actualRecipientCoin), fmt.Sprintf("RecipientCoin for bips: %v not equal expected: %s actual: %s name: %s", tc.bips, tc.expectedRecipientCoin.String(), actualRecipientCoin, tc.name))
			assert.True(t, tc.expectedFeePayoutCoin.Equal(actualFeePayoutCoin), fmt.Sprintf("FeePayoutCoin not equal expected: %s actual: %s name: %s", tc.expectedFeePayoutCoin.String(), actualFeePayoutCoin.String(), tc.name))
		} else {
			assert.EqualError(t, err, tc.expectedErrorMsg, tc.name)
		}
	}
}
