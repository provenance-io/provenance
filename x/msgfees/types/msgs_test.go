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
			name: "should succeed to validate basic, usd coin type",
			msg: NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin(UsdDenom, 10), validAddress, validAddress, ""),
			errorMsg: "",
		},
		{
			name: "should succeed to validate basic, nhash coin type",
			msg: NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin(UsdDenom, 10), validAddress, validAddress, ""),
			errorMsg: "",
		},
		{
			name: "should succeed to validate basic, non positive coin",
			msg: NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin(UsdDenom, 0), validAddress, validAddress, ""),
			errorMsg: "amount must be greater than zero",
		}, {
			name: "should succeed to validate basic, without recipient",
			msg: NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin(UsdDenom, 10), "", validAddress, ""),
			errorMsg: "",
		}, {
			name: "should succeed to validate basic, with bips set",
			msg: NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin(UsdDenom, 10), validAddress, validAddress, "5000"),
			errorMsg: "",
		}, {
			name: "should fail to validate basic, invalid address",
			msg: NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin(UsdDenom, 10), "invalid", validAddress, ""),
			errorMsg: "decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			name: "should fail to validate basic, invalid from address",
			msg: NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin("nhash", 10), "", "invalid", ""),
			errorMsg: "decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			name: "should fail to validate basic, invalid bips string",
			msg: NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin("nhash", 10), "", validAddress, "invalid_bips"),
			errorMsg: `strconv.ParseUint: parsing "invalid_bips": invalid syntax`,
		},
		{
			name: "should fail to validate basic, invalid bips string too small",
			msg: NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin("nhash", 10), "", validAddress, "-1"),
			errorMsg: `strconv.ParseUint: parsing "-1": invalid syntax`,
		},
		{
			name: "should fail to validate basic, invalid bips string too high",
			msg: NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin("nhash", 10), "", validAddress, "10001"),
			errorMsg: "recipient basis points can only be between 0 and 10,000 : 10001",
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
