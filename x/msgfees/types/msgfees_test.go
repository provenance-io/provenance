package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestMsgFeeValidate(t *testing.T) {
	validAddress := "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"
	cases := []struct {
		name     string
		msg      MsgFee
		errorMsg string
	}{
		{
			"should succeed to validate with no recipient",
			NewMsgFee(sdk.MsgTypeURL(&MsgAssessCustomMsgFeeRequest{}), sdk.NewInt64Coin(sdk.DefaultBondDenom, 100), "", DefaultMsgFeeSplit),
			"",
		},
		{
			"should succeed to validate with a recipient and valid split",
			NewMsgFee(sdk.MsgTypeURL(&MsgAssessCustomMsgFeeRequest{}), sdk.NewInt64Coin(sdk.DefaultBondDenom, 100), validAddress, DefaultMsgFeeSplit),
			"",
		},
		{
			"should fail to validate from invalid recipient address",
			NewMsgFee(sdk.MsgTypeURL(&MsgAssessCustomMsgFeeRequest{}), sdk.NewInt64Coin(sdk.DefaultBondDenom, 100), "invalid", DefaultMsgFeeSplit),
			"decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			"should fail to validate from invalid split amount",
			NewMsgFee(sdk.MsgTypeURL(&MsgAssessCustomMsgFeeRequest{}), sdk.NewInt64Coin(sdk.DefaultBondDenom, 100), validAddress, 101),
			"split can only be between 0 and 100 : 101",
		},
		{
			"should fail to validate from invalid msg type url",
			NewMsgFee("", sdk.NewInt64Coin(sdk.DefaultBondDenom, 100), validAddress, DefaultMsgFeeSplit),
			"invalid msg type url",
		},
		{
			"should succeed to validate with no recipient",
			NewMsgFee(sdk.MsgTypeURL(&MsgAssessCustomMsgFeeRequest{}), sdk.NewInt64Coin(sdk.DefaultBondDenom, 0), "", DefaultMsgFeeSplit),
			"invalid fee amount",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.Validate()
			if len(tc.errorMsg) > 0 {
				require.EqualError(t, err, tc.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
