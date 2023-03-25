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
		    name: 	"should succeed to validate with no recipient",
		    msg: 	NewMsgFee(sdk.MsgTypeURL(&MsgAssessCustomMsgFeeRequest{}), sdk.NewInt64Coin(sdk.DefaultBondDenom, 100), "", DefaultMsgFeeBips),
		    errorMsg: 	"",
		},
		{
			name:"should succeed to validate with a recipient and valid split",
			msg:NewMsgFee(sdk.MsgTypeURL(&MsgAssessCustomMsgFeeRequest{}), sdk.NewInt64Coin(sdk.DefaultBondDenom, 100), validAddress, DefaultMsgFeeBips),
			errorMsg:"",
		},
		{
			name:"should fail to validate from invalid recipient address",
			msg:NewMsgFee(sdk.MsgTypeURL(&MsgAssessCustomMsgFeeRequest{}), sdk.NewInt64Coin(sdk.DefaultBondDenom, 100), "invalid", DefaultMsgFeeBips),
			errorMsg:"decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			name: "should fail to validate from invalid bip point amount too large",
			msg: NewMsgFee(sdk.MsgTypeURL(&MsgAssessCustomMsgFeeRequest{}), sdk.NewInt64Coin(sdk.DefaultBondDenom, 100), validAddress, 10_001),
			errorMsg: "recipient basis points can only be between 0 and 10,000 : 10001",
		},
		{
			name: "should fail to validate from invalid msg type url",
			msg: NewMsgFee("", sdk.NewInt64Coin(sdk.DefaultBondDenom, 100), validAddress, DefaultMsgFeeBips),
			errorMsg: "invalid msg type url",
		},
		{
			name: "should succeed to validate with no recipient",
			msg: NewMsgFee(sdk.MsgTypeURL(&MsgAssessCustomMsgFeeRequest{}), sdk.NewInt64Coin(sdk.DefaultBondDenom, 0), "", DefaultMsgFeeBips),
			errorMsg: "invalid fee amount",
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
