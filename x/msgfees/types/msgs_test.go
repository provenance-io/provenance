package types

import (
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
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

func TestMsgAddMsgFeeProposalRequestValidateBasic(t *testing.T) {
	authority := sdk.AccAddress("input111111111111111").String()

	cases := []struct {
		name     string
		msg      *MsgAddMsgFeeProposalRequest
		errorMsg string
	}{
		{
			"Empty type error",
			NewMsgAddMsgFeeProposalRequest("", sdk.NewCoin("hotdog", sdk.NewInt(10)), "", "", authority),
			"msg type is empty",
		},
		{
			"Invalid fee amounts",
			NewMsgAddMsgFeeProposalRequest("msgType", sdk.NewCoin("hotdog", sdk.NewInt(0)), "", "", authority),
			"invalid fee amount",
		},
		{
			"Invalid proposal recipient address",
			NewMsgAddMsgFeeProposalRequest("msgType", sdk.NewCoin("hotdog", sdk.NewInt(10)), "invalid", "", authority),
			"decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			"Invalid proposal invalid basis points for address",
			NewMsgAddMsgFeeProposalRequest("msgType", sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", "10001", authority),
			"recipient basis points can only be between 0 and 10,000 : 10001",
		},
		{
			"Valid proposal without recipient",
			NewMsgAddMsgFeeProposalRequest("msgType", sdk.NewCoin("hotdog", sdk.NewInt(10)), "", "", authority),
			"",
		},
		{
			"Valid proposal with recipient",
			NewMsgAddMsgFeeProposalRequest("msgType", sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", "10000", authority),
			"",
		},
		{
			"Valid proposal with recipient without defined bips",
			NewMsgAddMsgFeeProposalRequest("msgType", sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", "", authority),
			"",
		},
		{
			"Valid proposal with recipient with defined bips",
			NewMsgAddMsgFeeProposalRequest("msgType", sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", "10", authority),
			"",
		},
		{
			"invalid authority",
			NewMsgAddMsgFeeProposalRequest("msgType", sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", "10", ""),
			"empty address string is not allowed",
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
			"Empty type error",
			NewMsgUpdateMsgFeeProposalRequest("", sdk.NewCoin("hotdog", sdk.NewInt(10)), "", "", authority),
			"msg type is empty",
		},
		{
			"Invalid fee amounts",
			NewMsgUpdateMsgFeeProposalRequest(msgType, sdk.NewCoin("hotdog", sdk.NewInt(0)), "", "", authority),
			"invalid fee amount",
		},
		{
			"Invalid proposal recipient address",
			NewMsgUpdateMsgFeeProposalRequest(msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "invalid", "50", authority),
			"decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			"Invalid proposal invalid basis points for address",
			NewMsgUpdateMsgFeeProposalRequest(msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", "10001", authority),
			"recipient basis points can only be between 0 and 10,000 : 10001",
		},
		{
			"Valid proposal without recipient",
			NewMsgUpdateMsgFeeProposalRequest(msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "", "", authority),
			"",
		},
		{
			"Valid proposal with recipient without defined bips",
			NewMsgUpdateMsgFeeProposalRequest(msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", "", authority),
			"",
		},
		{
			"Valid proposal with recipient with defined bips",
			NewMsgUpdateMsgFeeProposalRequest(msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", "10", authority),
			"",
		},
		{
			"invalid authority",
			NewMsgUpdateMsgFeeProposalRequest(msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", "10", ""),
			"empty address string is not allowed",
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
			"valid message",
			NewMsgRemoveMsgFeeProposalRequest(msgType, authority),
			"",
		},
		{
			"empty message type",
			NewMsgRemoveMsgFeeProposalRequest("", authority),
			"msg type is empty",
		},
		{
			"invalid authority",
			NewMsgRemoveMsgFeeProposalRequest(msgType, ""),
			"empty address string is not allowed",
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
			"Empty type error",
			NewMsgUpdateNhashPerUsdMilProposalRequest(0, authority),
			"nhash per usd mil must be greater than 0",
		},
		{
			"Valid proposal",
			NewMsgUpdateNhashPerUsdMilProposalRequest(70, authority),
			"",
		},
		{
			"invalid authority",
			NewMsgUpdateNhashPerUsdMilProposalRequest(70, ""),
			"empty address string is not allowed",
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
			"invalid denom",
			NewMsgUpdateConversionFeeDenomProposalRequest("??", authority),
			"invalid denom: ??",
		},
		{
			"valid message",
			NewMsgUpdateConversionFeeDenomProposalRequest("hotdog", authority),
			"",
		},
		{
			"invalid authority",
			NewMsgUpdateConversionFeeDenomProposalRequest("hotdog", ""),
			"empty address string is not allowed",
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
