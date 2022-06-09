package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestMsgAssessCustomMsgFeeValidateBasic(t *testing.T) {
	validAddress := "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"

	cases := map[string]struct {
		msg      MsgAssessCustomMsgFeeRequest
		wantErr  bool
		errorMsg string
	}{
		"should succeed to validate basic, usd coin type": {
			NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin(UsdDenom, 10), validAddress, validAddress),
			false,
			"",
		},
		"should succeed to validate basic, nhash coin type": {
			NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin(UsdDenom, 10), validAddress, validAddress),
			false,
			"",
		},
		"should succeed to validate basic, non positive coin": {
			NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin(UsdDenom, 0), validAddress, validAddress),
			true,
			"amount must be greater than zero",
		},
		"should fail to validate basic, invalid coin type": {
			NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin("jackthecat", 10), validAddress, validAddress),
			true,
			"denom must be in usd or nhash : jackthecat",
		},
		"should succeed to validate basic, without recipient": {
			NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin(UsdDenom, 10), "", validAddress),
			false,
			"",
		},
		"should fail to validate basic, invalid address": {
			NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin(NhashDenom, 10), "invalid", validAddress),
			true,
			"decoding bech32 failed: invalid bech32 string length 7",
		},
		"should fail to validate basic, invalid address from address": {
			NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin(NhashDenom, 10), "", "invalid"),
			true,
			"decoding bech32 failed: invalid bech32 string length 7",
		},
	}

	for n, tc := range cases {
		tc := tc

		t.Run(n, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.wantErr {
				require.Error(t, err)
				require.Equal(t, tc.errorMsg, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
