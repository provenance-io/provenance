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
			name: "should succeed to validate basic, usd coin type",
			msg: MsgAssessCustomMsgFeeRequest{
				Name:                 "shortname",
				Amount:               sdk.NewInt64Coin(UsdDenom, 10),
				Recipient:            validAddress,
				From:                 validAddress,
				RecipientBasisPoints: "",
			},
			errorMsg: "",
		},
		{
			name: "should succeed to validate basic, nhash coin type",
			msg: MsgAssessCustomMsgFeeRequest{
				Name:                 "shortname",
				Amount:               sdk.NewInt64Coin(UsdDenom, 10),
				Recipient:            validAddress,
				From:                 validAddress,
				RecipientBasisPoints: "",
			},
			errorMsg: "",
		},
		{
			name: "should succeed to validate basic, non positive coin",
			msg: MsgAssessCustomMsgFeeRequest{
				Name:                 "shortname",
				Amount:               sdk.NewInt64Coin(UsdDenom, 0),
				Recipient:            validAddress,
				From:                 validAddress,
				RecipientBasisPoints: "",
			},
			errorMsg: "amount must be greater than zero",
		},
		{
			name: "should succeed to validate basic, without recipient",
			msg: MsgAssessCustomMsgFeeRequest{
				Name:                 "shortname",
				Amount:               sdk.NewInt64Coin(UsdDenom, 10),
				Recipient:            "",
				From:                 validAddress,
				RecipientBasisPoints: "",
			},
			errorMsg: "",
		},
		{
			name: "should succeed to validate basic, with bips set",
			msg: MsgAssessCustomMsgFeeRequest{
				Name:                 "shortname",
				Amount:               sdk.NewInt64Coin(UsdDenom, 10),
				Recipient:            validAddress,
				From:                 validAddress,
				RecipientBasisPoints: "5000",
			},
			errorMsg: "",
		}, {
			name: "should fail to validate basic, invalid address",
			msg: MsgAssessCustomMsgFeeRequest{
				Name:                 "shortname",
				Amount:               sdk.NewInt64Coin(UsdDenom, 10),
				Recipient:            "invalid",
				From:                 validAddress,
				RecipientBasisPoints: "",
			},
			errorMsg: "decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			name: "should fail to validate basic, invalid from address",
			msg: MsgAssessCustomMsgFeeRequest{
				Name:                 "shortname",
				Amount:               sdk.NewInt64Coin("nhash", 10),
				Recipient:            "",
				From:                 "invalid",
				RecipientBasisPoints: "",
			},
			errorMsg: "decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			name: "should fail to validate basic, invalid bips string",
			msg: MsgAssessCustomMsgFeeRequest{
				Name:                 "shortname",
				Amount:               sdk.NewInt64Coin("nhash", 10),
				Recipient:            "",
				From:                 validAddress,
				RecipientBasisPoints: "invalid_bips",
			},
			errorMsg: `strconv.ParseUint: parsing "invalid_bips": invalid syntax`,
		},
		{
			name: "should fail to validate basic, invalid bips string too small",
			msg: MsgAssessCustomMsgFeeRequest{
				Name:                 "shortname",
				Amount:               sdk.NewInt64Coin("nhash", 10),
				Recipient:            "",
				From:                 validAddress,
				RecipientBasisPoints: "-1",
			},
			errorMsg: `strconv.ParseUint: parsing "-1": invalid syntax`,
		},
		{
			name: "should fail to validate basic, invalid bips string too high",
			msg: MsgAssessCustomMsgFeeRequest{
				Name:                 "shortname",
				Amount:               sdk.NewInt64Coin("nhash", 10),
				Recipient:            "",
				From:                 validAddress,
				RecipientBasisPoints: "10001",
			},
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
		msg      MsgAddMsgFeeProposalRequest
		errorMsg string
	}{
		{
			name: "Empty type error",
			msg: MsgAddMsgFeeProposalRequest{
				MsgTypeUrl:           "",
				AdditionalFee:        sdk.NewCoin("hotdog", sdk.NewInt(10)),
				Recipient:            "",
				RecipientBasisPoints: "",
				Authority:            authority,
			},
			errorMsg: "msg type is empty",
		},
		{
			name: "Invalid fee amounts",
			msg: MsgAddMsgFeeProposalRequest{
				MsgTypeUrl:           "msgType",
				AdditionalFee:        sdk.NewCoin("hotdog", sdk.NewInt(0)),
				Recipient:            "",
				RecipientBasisPoints: "",
				Authority:            authority,
			},
			errorMsg: "invalid fee amount",
		},
		{
			name: "Invalid proposal recipient address",
			msg: MsgAddMsgFeeProposalRequest{
				MsgTypeUrl:           "msgType",
				AdditionalFee:        sdk.NewCoin("hotdog", sdk.NewInt(10)),
				Recipient:            "invalid",
				RecipientBasisPoints: "",
				Authority:            authority,
			},
			errorMsg: "decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			name: "Invalid proposal invalid basis points for address",
			msg: MsgAddMsgFeeProposalRequest{
				MsgTypeUrl:           "msgType",
				AdditionalFee:        sdk.NewCoin("hotdog", sdk.NewInt(10)),
				Recipient:            "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27",
				RecipientBasisPoints: "10001",
				Authority:            authority,
			},
			errorMsg: "recipient basis points can only be between 0 and 10,000 : 10001",
		},
		{
			name: "Valid proposal without recipient",
			msg: MsgAddMsgFeeProposalRequest{
				MsgTypeUrl:           "msgType",
				AdditionalFee:        sdk.NewCoin("hotdog", sdk.NewInt(10)),
				Recipient:            "",
				RecipientBasisPoints: "",
				Authority:            authority,
			},
			errorMsg: "",
		},
		{
			name: "Valid proposal with recipient",
			msg: MsgAddMsgFeeProposalRequest{
				MsgTypeUrl:           "msgType",
				AdditionalFee:        sdk.NewCoin("hotdog", sdk.NewInt(10)),
				Recipient:            "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27",
				RecipientBasisPoints: "10000",
				Authority:            authority,
			},
			errorMsg: "",
		},
		{
			name: "Valid proposal with recipient without defined bips",
			msg: MsgAddMsgFeeProposalRequest{
				MsgTypeUrl:           "msgType",
				AdditionalFee:        sdk.NewCoin("hotdog", sdk.NewInt(10)),
				Recipient:            "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27",
				RecipientBasisPoints: "",
				Authority:            authority,
			},
			errorMsg: "",
		},
		{
			name: "Valid proposal with recipient with defined bips",
			msg: MsgAddMsgFeeProposalRequest{
				MsgTypeUrl:           "msgType",
				AdditionalFee:        sdk.NewCoin("hotdog", sdk.NewInt(10)),
				Recipient:            "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27",
				RecipientBasisPoints: "10",
				Authority:            authority,
			},
			errorMsg: "",
		},
		{
			name: "invalid authority",
			msg: MsgAddMsgFeeProposalRequest{
				MsgTypeUrl:           "msgType",
				AdditionalFee:        sdk.NewCoin("hotdog", sdk.NewInt(10)),
				Recipient:            "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27",
				RecipientBasisPoints: "10",
				Authority:            "",
			},
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
	msg := MsgAddMsgFeeProposalRequest{
		MsgTypeUrl:           "msgType",
		AdditionalFee:        sdk.NewCoin("hotdog", sdk.NewInt(10)),
		Recipient:            "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27",
		RecipientBasisPoints: "10",
		Authority:            authority.String(),
	}
	res := msg.GetSigners()
	require.Equal(t, 1, len(res))
	require.True(t, authority.Equals(res[0]))
}

func TestMsgUpdateMsgFeeProposalRequestValidateBasic(t *testing.T) {
	msgType := sdk.MsgTypeURL(&metadatatypes.MsgWriteRecordRequest{})
	authority := sdk.AccAddress("input111111111111111").String()

	cases := []struct {
		name     string
		msg      MsgUpdateMsgFeeProposalRequest
		errorMsg string
	}{
		{
			name: "Empty type error",
			msg: MsgUpdateMsgFeeProposalRequest{
				MsgTypeUrl:           "",
				AdditionalFee:        sdk.NewCoin("hotdog", sdk.NewInt(10)),
				Recipient:            "",
				RecipientBasisPoints: "",
				Authority:            authority,
			},
			errorMsg: "msg type is empty",
		},
		{
			name: "Invalid fee amounts",
			msg: MsgUpdateMsgFeeProposalRequest{
				MsgTypeUrl:           msgType,
				AdditionalFee:        sdk.NewCoin("hotdog", sdk.NewInt(0)),
				Recipient:            "",
				RecipientBasisPoints: "",
				Authority:            authority,
			},
			errorMsg: "invalid fee amount",
		},
		{
			name: "Invalid proposal recipient address",
			msg: MsgUpdateMsgFeeProposalRequest{
				MsgTypeUrl:           msgType,
				AdditionalFee:        sdk.NewCoin("hotdog", sdk.NewInt(10)),
				Recipient:            "invalid",
				RecipientBasisPoints: "50",
				Authority:            authority,
			},
			errorMsg: "decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			name: "Invalid proposal invalid basis points for address",
			msg: MsgUpdateMsgFeeProposalRequest{
				MsgTypeUrl:           msgType,
				AdditionalFee:        sdk.NewCoin("hotdog", sdk.NewInt(10)),
				Recipient:            "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27",
				RecipientBasisPoints: "10001",
				Authority:            authority,
			},
			errorMsg: "recipient basis points can only be between 0 and 10,000 : 10001",
		},
		{
			name: "Valid proposal without recipient",
			msg: MsgUpdateMsgFeeProposalRequest{
				MsgTypeUrl:           msgType,
				AdditionalFee:        sdk.NewCoin("hotdog", sdk.NewInt(10)),
				Recipient:            "",
				RecipientBasisPoints: "",
				Authority:            authority,
			},
			errorMsg: "",
		},
		{
			name: "Valid proposal with recipient without defined bips",
			msg: MsgUpdateMsgFeeProposalRequest{
				MsgTypeUrl:           msgType,
				AdditionalFee:        sdk.NewCoin("hotdog", sdk.NewInt(10)),
				Recipient:            "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27",
				RecipientBasisPoints: "",
				Authority:            authority,
			},
			errorMsg: "",
		},
		{
			name: "Valid proposal with recipient with defined bips",
			msg: MsgUpdateMsgFeeProposalRequest{
				MsgTypeUrl:           msgType,
				AdditionalFee:        sdk.NewCoin("hotdog", sdk.NewInt(10)),
				Recipient:            "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27",
				RecipientBasisPoints: "10",
				Authority:            authority,
			},
			errorMsg: "",
		},
		{
			name: "invalid authority",
			msg: MsgUpdateMsgFeeProposalRequest{
				MsgTypeUrl:           msgType,
				AdditionalFee:        sdk.NewCoin("hotdog", sdk.NewInt(10)),
				Recipient:            "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27",
				RecipientBasisPoints: "10",
				Authority:            "",
			},
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
	msg := MsgUpdateMsgFeeProposalRequest{
		MsgTypeUrl:           "msgType",
		AdditionalFee:        sdk.NewCoin("hotdog", sdk.NewInt(10)),
		Recipient:            "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27",
		RecipientBasisPoints: "10",
		Authority:            authority.String(),
	}
	res := msg.GetSigners()
	require.Equal(t, 1, len(res))
	require.True(t, authority.Equals(res[0]))
}

func TestMsgRemoveMsgFeeProposalRequestValidateBasic(t *testing.T) {
	msgType := sdk.MsgTypeURL(&metadatatypes.MsgWriteRecordRequest{})

	authority := sdk.AccAddress("input111111111111111").String()

	cases := []struct {
		name     string
		msg      MsgRemoveMsgFeeProposalRequest
		errorMsg string
	}{
		{
			name: "valid message",
			msg: MsgRemoveMsgFeeProposalRequest{
				MsgTypeUrl: msgType,
				Authority:  authority,
			},
			errorMsg: "",
		},
		{
			name: "empty message type",
			msg: MsgRemoveMsgFeeProposalRequest{
				MsgTypeUrl: "",
				Authority:  authority,
			},
			errorMsg: "msg type is empty",
		},
		{
			name: "invalid authority",
			msg: MsgRemoveMsgFeeProposalRequest{
				MsgTypeUrl: msgType,
				Authority:  "",
			},
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
	msg := MsgRemoveMsgFeeProposalRequest{
		MsgTypeUrl: "msgtype",
		Authority:  authority.String(),
	}
	res := msg.GetSigners()
	require.Equal(t, 1, len(res))
	require.True(t, authority.Equals(res[0]))
}

func TestMsgUpdateNhashPerUsdMilProposalRequestValidateBasic(t *testing.T) {
	authority := sdk.AccAddress("input111111111111111").String()

	cases := []struct {
		name     string
		msg      MsgUpdateNhashPerUsdMilProposalRequest
		errorMsg string
	}{
		{
			name: "Empty type error",
			msg: MsgUpdateNhashPerUsdMilProposalRequest{
				NhashPerUsdMil: 0,
				Authority:      authority,
			},
			errorMsg: "nhash per usd mil must be greater than 0",
		},
		{
			name: "Valid proposal",
			msg: MsgUpdateNhashPerUsdMilProposalRequest{
				NhashPerUsdMil: 70,
				Authority:      authority,
			},
			errorMsg: "",
		},
		{
			name: "invalid authority",
			msg: MsgUpdateNhashPerUsdMilProposalRequest{
				NhashPerUsdMil: 70,
				Authority:      "",
			},
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
	msg := MsgUpdateNhashPerUsdMilProposalRequest{
		NhashPerUsdMil: 70,
		Authority:      authority.String(),
	}
	res := msg.GetSigners()
	require.Equal(t, 1, len(res))
	require.True(t, authority.Equals(res[0]))
}

func TestUpdateConversionFeeDenomProposalRequestValidateBasic(t *testing.T) {
	authority := sdk.AccAddress("input111111111111111").String()

	cases := []struct {
		name     string
		msg      MsgUpdateConversionFeeDenomProposalRequest
		errorMsg string
	}{
		{
			name: "invalid denom",
			msg: MsgUpdateConversionFeeDenomProposalRequest{
				ConversionFeeDenom: "??",
				Authority:          authority,
			},
			errorMsg: "invalid denom: ??",
		},
		{
			name: "valid message",
			msg: MsgUpdateConversionFeeDenomProposalRequest{
				ConversionFeeDenom: "hotdog",
				Authority:          authority,
			},
			errorMsg: "",
		},
		{
			name: "invalid authority",
			msg: MsgUpdateConversionFeeDenomProposalRequest{
				ConversionFeeDenom: "hotdog",
				Authority:          "",
			},
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
	msg := MsgUpdateConversionFeeDenomProposalRequest{
		ConversionFeeDenom: "some-denom",
		Authority:          authority.String(),
	}
	res := msg.GetSigners()
	require.Equal(t, 1, len(res))
	require.True(t, authority.Equals(res[0]))
}
