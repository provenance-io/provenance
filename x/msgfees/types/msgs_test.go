package types

import (
	"testing"

	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"

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
			name:     "should succeed to validate basic, usd coin type",
			msg:      NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin(UsdDenom, 10), validAddress, validAddress, ""),
			errorMsg: "",
		},
		{
			name:     "should succeed to validate basic, nhash coin type",
			msg:      NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin(UsdDenom, 10), validAddress, validAddress, ""),
			errorMsg: "",
		},
		{
			name:     "should succeed to validate basic, non positive coin",
			msg:      NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin(UsdDenom, 0), validAddress, validAddress, ""),
			errorMsg: "amount must be greater than zero",
		},
		{
			name:     "should succeed to validate basic, without recipient",
			msg:      NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin(UsdDenom, 10), "", validAddress, ""),
			errorMsg: "",
		},
		{
			name:     "should succeed to validate basic, with bips set",
			msg:      NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin(UsdDenom, 10), validAddress, validAddress, "5000"),
			errorMsg: "",
		}, {
			name:     "should fail to validate basic, invalid address",
			msg:      NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin(UsdDenom, 10), "invalid", validAddress, ""),
			errorMsg: "decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			name:     "should fail to validate basic, invalid from address",
			msg:      NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin("nhash", 10), "", "invalid", ""),
			errorMsg: "decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			name:     "should fail to validate basic, invalid bips string",
			msg:      NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin("nhash", 10), "", validAddress, "invalid_bips"),
			errorMsg: `strconv.ParseUint: parsing "invalid_bips": invalid syntax`,
		},
		{
			name:     "should fail to validate basic, invalid bips string too small",
			msg:      NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin("nhash", 10), "", validAddress, "-1"),
			errorMsg: `strconv.ParseUint: parsing "-1": invalid syntax`,
		},
		{
			name:     "should fail to validate basic, invalid bips string too high",
			msg:      NewMsgAssessCustomMsgFeeRequest("shortname", sdk.NewInt64Coin("nhash", 10), "", validAddress, "10001"),
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

func TestMsgAddMsgFeeProposalRequestValidateBasic(t *testing.T) {
	authority := sdk.AccAddress("input111111111111111").String()

	cases := []struct {
		name     string
		msg      *MsgAddMsgFeeProposalRequest
		errorMsg string
	}{
		{
			name:     "Empty type error",
			msg:      NewMsgAddMsgFeeProposalRequest("", sdk.NewCoin("hotdog", sdk.NewInt(10)), "", "", authority),
			errorMsg: "msg type is empty",
		},
		{
			name:     "Invalid fee amounts",
			msg:      NewMsgAddMsgFeeProposalRequest("msgType", sdk.NewCoin("hotdog", sdk.NewInt(0)), "", "", authority),
			errorMsg: "invalid fee amount",
		},
		{
			name:     "Invalid proposal recipient address",
			msg:      NewMsgAddMsgFeeProposalRequest("msgType", sdk.NewCoin("hotdog", sdk.NewInt(10)), "invalid", "", authority),
			errorMsg: "decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			name:     "Invalid proposal invalid basis points for address",
			msg:      NewMsgAddMsgFeeProposalRequest("msgType", sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", "10001", authority),
			errorMsg: "recipient basis points can only be between 0 and 10,000 : 10001",
		},
		{
			name:     "Valid proposal without recipient",
			msg:      NewMsgAddMsgFeeProposalRequest("msgType", sdk.NewCoin("hotdog", sdk.NewInt(10)), "", "", authority),
			errorMsg: "",
		},
		{
			name:     "Valid proposal with recipient",
			msg:      NewMsgAddMsgFeeProposalRequest("msgType", sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", "10000", authority),
			errorMsg: "",
		},
		{
			name:     "Valid proposal with recipient without defined bips",
			msg:      NewMsgAddMsgFeeProposalRequest("msgType", sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", "", authority),
			errorMsg: "",
		},
		{
			name:     "Valid proposal with recipient with defined bips",
			msg:      NewMsgAddMsgFeeProposalRequest("msgType", sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", "10", authority),
			errorMsg: "",
		},
		{
			name:     "invalid authority",
			msg:      NewMsgAddMsgFeeProposalRequest("msgType", sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", "10", ""),
			errorMsg: "empty address string is not allowed",
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

func TestMsgAddMsgFeeProposalRequestGetSigners(t *testing.T) {
	authority := sdk.AccAddress("input111111111111111")
	msg := NewMsgAddMsgFeeProposalRequest("msgType", sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", "10", authority.String())
	res := msg.GetSigners()
	require.Equal(t, 1, len(res))
	require.True(t, authority.Equals(res[0]))
}

func TestMsgUpdateMsgFeeProposalRequestValidateBasic(t *testing.T) {
	msgType := sdk.MsgTypeURL(&metadatatypes.MsgWriteRecordRequest{})
	authority := sdk.AccAddress("input111111111111111").String()

	cases := []struct {
		name     string
		msg      *MsgUpdateMsgFeeProposalRequest
		errorMsg string
	}{
		{
			name:     "Empty type error",
			msg:      NewMsgUpdateMsgFeeProposalRequest("", sdk.NewCoin("hotdog", sdk.NewInt(10)), "", "", authority),
			errorMsg: "msg type is empty",
		},
		{
			name:     "Invalid fee amounts",
			msg:      NewMsgUpdateMsgFeeProposalRequest(msgType, sdk.NewCoin("hotdog", sdk.NewInt(0)), "", "", authority),
			errorMsg: "invalid fee amount",
		},
		{
			name:     "Invalid proposal recipient address",
			msg:      NewMsgUpdateMsgFeeProposalRequest(msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "invalid", "50", authority),
			errorMsg: "decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			name:     "Invalid proposal invalid basis points for address",
			msg:      NewMsgUpdateMsgFeeProposalRequest(msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", "10001", authority),
			errorMsg: "recipient basis points can only be between 0 and 10,000 : 10001",
		},
		{
			name:     "Valid proposal without recipient",
			msg:      NewMsgUpdateMsgFeeProposalRequest(msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "", "", authority),
			errorMsg: "",
		},
		{
			name:     "Valid proposal with recipient without defined bips",
			msg:      NewMsgUpdateMsgFeeProposalRequest(msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", "", authority),
			errorMsg: "",
		},
		{
			name:     "Valid proposal with recipient with defined bips",
			msg:      NewMsgUpdateMsgFeeProposalRequest(msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", "10", authority),
			errorMsg: "",
		},
		{
			name:     "invalid authority",
			msg:      NewMsgUpdateMsgFeeProposalRequest(msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", "10", ""),
			errorMsg: "empty address string is not allowed",
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

func TestMsgUpdateMsgFeeProposalRequestGetSigners(t *testing.T) {
	authority := sdk.AccAddress("input111111111111111")
	msg := NewMsgUpdateMsgFeeProposalRequest("msgType", sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", "10", authority.String())
	res := msg.GetSigners()
	require.Equal(t, 1, len(res))
	require.True(t, authority.Equals(res[0]))
}

func TestMsgRemoveMsgFeeProposalRequestValidateBasic(t *testing.T) {
	msgType := sdk.MsgTypeURL(&metadatatypes.MsgWriteRecordRequest{})

	authority := sdk.AccAddress("input111111111111111").String()

	cases := []struct {
		name     string
		msg      *MsgRemoveMsgFeeProposalRequest
		errorMsg string
	}{
		{
			name:     "valid message",
			msg:      NewMsgRemoveMsgFeeProposalRequest(msgType, authority),
			errorMsg: "",
		},
		{
			name:     "empty message type",
			msg:      NewMsgRemoveMsgFeeProposalRequest("", authority),
			errorMsg: "msg type is empty",
		},
		{
			name:     "invalid authority",
			msg:      NewMsgRemoveMsgFeeProposalRequest(msgType, ""),
			errorMsg: "empty address string is not allowed",
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

func TestMsgRemoveMsgFeeProposalRequestGetSigners(t *testing.T) {
	authority := sdk.AccAddress("input111111111111111")
	msg := NewMsgRemoveMsgFeeProposalRequest("msgtype", authority.String())
	res := msg.GetSigners()
	require.Equal(t, 1, len(res))
	require.True(t, authority.Equals(res[0]))
}

func TestMsgUpdateNhashPerUsdMilProposalRequestValidateBasic(t *testing.T) {
	authority := sdk.AccAddress("input111111111111111").String()

	cases := []struct {
		name     string
		msg      *MsgUpdateNhashPerUsdMilProposalRequest
		errorMsg string
	}{
		{
			name:     "Empty type error",
			msg:      NewMsgUpdateNhashPerUsdMilProposalRequest(0, authority),
			errorMsg: "nhash per usd mil must be greater than 0",
		},
		{
			name:     "Valid proposal",
			msg:      NewMsgUpdateNhashPerUsdMilProposalRequest(70, authority),
			errorMsg: "",
		},
		{
			name:     "invalid authority",
			msg:      NewMsgUpdateNhashPerUsdMilProposalRequest(70, ""),
			errorMsg: "empty address string is not allowed",
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

func TestMsgUpdateNhashPerUsdMilProposalRequestGetSigners(t *testing.T) {
	authority := sdk.AccAddress("input111111111111111")
	msg := NewMsgUpdateNhashPerUsdMilProposalRequest(70, authority.String())
	res := msg.GetSigners()
	require.Equal(t, 1, len(res))
	require.True(t, authority.Equals(res[0]))
}

func TestUpdateConversionFeeDenomProposalRequestValidateBasic(t *testing.T) {
	authority := sdk.AccAddress("input111111111111111").String()

	cases := []struct {
		name     string
		msg      *MsgUpdateConversionFeeDenomProposalRequest
		errorMsg string
	}{
		{
			name:     "invalid denom",
			msg:      NewMsgUpdateConversionFeeDenomProposalRequest("??", authority),
			errorMsg: "invalid denom: ??",
		},
		{
			name:     "valid message",
			msg:      NewMsgUpdateConversionFeeDenomProposalRequest("hotdog", authority),
			errorMsg: "",
		},
		{
			name:     "invalid authority",
			msg:      NewMsgUpdateConversionFeeDenomProposalRequest("hotdog", ""),
			errorMsg: "empty address string is not allowed",
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

func TestUpdateConversionFeeDenomProposalRequestGetSigners(t *testing.T) {
	authority := sdk.AccAddress("input111111111111111")
	msg := NewMsgUpdateConversionFeeDenomProposalRequest("some-denom", authority.String())
	res := msg.GetSigners()
	require.Equal(t, 1, len(res))
	require.True(t, authority.Equals(res[0]))
}
