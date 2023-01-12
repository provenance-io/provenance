package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"
)

func TestMsgAssessCustomMsgFeeValidateBasic(t *testing.T) {
	validAddress := "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"
	cases := []struct {
		name     string
		msg      MsgAssessCustomMsgFeeRequest
		errorMsg string
	}{
		{
			"should succeed to validate basic, usd coin type",
			NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin(UsdDenom, 10), validAddress, validAddress, ""),
			"",
		},
		{
			"should succeed to validate basic, nhash coin type",
			NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin(UsdDenom, 10), validAddress, validAddress, ""),
			"",
		},
		{
			"should succeed to validate basic, non positive coin",
			NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin(UsdDenom, 0), validAddress, validAddress, ""),
			"amount must be greater than zero",
		}, {
			"should succeed to validate basic, without recipient",
			NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin(UsdDenom, 10), "", validAddress, ""),
			"",
		}, {
			"should succeed to validate basic, with bips set",
			NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin(UsdDenom, 10), validAddress, validAddress, "5000"),
			"",
		}, {
			"should fail to validate basic, invalid address",
			NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin(UsdDenom, 10), "invalid", validAddress, ""),
			"decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			"should fail to validate basic, invalid from address",
			NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin("nhash", 10), "", "invalid", ""),
			"decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			"should fail to validate basic, invalid bips string",
			NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin("nhash", 10), "", validAddress, "invalid_bips"),
			`strconv.ParseUint: parsing "invalid_bips": invalid syntax`,
		},
		{
			"should fail to validate basic, invalid bips string too small",
			NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin("nhash", 10), "", validAddress, "-1"),
			`strconv.ParseUint: parsing "-1": invalid syntax`,
		},
		{
			"should fail to validate basic, invalid bips string too high",
			NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin("nhash", 10), "", validAddress, "10001"),
			"recipient basis points can only be between 0 and 10,000 : 10001",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if len(tc.errorMsg) > 0 {
				require.EqualError(t, err, tc.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
