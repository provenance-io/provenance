package types_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/x/feegrant"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"

	"github.com/provenance-io/provenance/testutil"

	. "github.com/provenance-io/provenance/x/marker/types"
)

func TestAllMsgsGetSigners(t *testing.T) {
	msgMakers := []testutil.MsgMaker{
		func(signer string) sdk.Msg { return &MsgFinalizeRequest{Administrator: signer} },
		func(signer string) sdk.Msg { return &MsgActivateRequest{Administrator: signer} },
		func(signer string) sdk.Msg { return &MsgCancelRequest{Administrator: signer} },
		func(signer string) sdk.Msg { return &MsgDeleteAccessRequest{Administrator: signer} },
		func(signer string) sdk.Msg { return &MsgMintRequest{Administrator: signer} },
		func(signer string) sdk.Msg { return &MsgBurnRequest{Administrator: signer} },
		func(signer string) sdk.Msg { return &MsgAddAccessRequest{Administrator: signer} },
		func(signer string) sdk.Msg { return &MsgDeleteRequest{Administrator: signer} },
		func(signer string) sdk.Msg { return &MsgWithdrawRequest{Administrator: signer} },
		func(signer string) sdk.Msg { return &MsgAddMarkerRequest{FromAddress: signer} },
		func(signer string) sdk.Msg { return &MsgTransferRequest{Administrator: signer} },
		func(signer string) sdk.Msg { return &MsgIbcTransferRequest{Administrator: signer} },
		func(signer string) sdk.Msg { return &MsgSetDenomMetadataRequest{Administrator: signer} },
		func(signer string) sdk.Msg { return &MsgGrantAllowanceRequest{Administrator: signer} },
		func(signer string) sdk.Msg { return &MsgRevokeGrantAllowanceRequest{Administrator: signer} },
		func(signer string) sdk.Msg { return &MsgAddFinalizeActivateMarkerRequest{FromAddress: signer} },
		func(signer string) sdk.Msg { return &MsgSupplyIncreaseProposalRequest{Authority: signer} },
		func(signer string) sdk.Msg { return &MsgSupplyDecreaseProposalRequest{Authority: signer} },
		func(signer string) sdk.Msg { return &MsgUpdateRequiredAttributesRequest{TransferAuthority: signer} },
		func(signer string) sdk.Msg { return &MsgUpdateForcedTransferRequest{Authority: signer} },
		func(signer string) sdk.Msg { return &MsgSetAccountDataRequest{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgUpdateSendDenyListRequest{Authority: signer} },
		func(signer string) sdk.Msg { return &MsgAddNetAssetValuesRequest{Administrator: signer} },
		func(signer string) sdk.Msg { return &MsgSetAdministratorProposalRequest{Authority: signer} },
		func(signer string) sdk.Msg { return &MsgRemoveAdministratorProposalRequest{Authority: signer} },
		func(signer string) sdk.Msg { return &MsgChangeStatusProposalRequest{Authority: signer} },
		func(signer string) sdk.Msg { return &MsgWithdrawEscrowProposalRequest{Authority: signer} },
		func(signer string) sdk.Msg { return &MsgSetDenomMetadataProposalRequest{Authority: signer} },
		func(signer string) sdk.Msg { return &MsgUpdateParamsRequest{Authority: signer} },
	}

	testutil.RunGetSignersTests(t, AllRequestMsgs, msgMakers, nil)
}

func TestMsgGrantAllowance(t *testing.T) {
	cdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
	addr, _ := sdk.AccAddressFromBech32("cosmos1aeuqja06474dfrj7uqsvukm6rael982kk89mqr")
	addr2, _ := sdk.AccAddressFromBech32("cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl")
	testcoin := sdk.NewCoins(sdk.NewInt64Coin("testcoin", 100))
	threeHours := time.Now().Add(3 * time.Hour)
	basic := &feegrant.BasicAllowance{
		SpendLimit: testcoin,
		Expiration: &threeHours,
	}

	cases := map[string]struct {
		denom         string
		grantee       sdk.AccAddress
		administrator sdk.AccAddress
		allowance     *feegrant.BasicAllowance
		valid         bool
	}{
		"valid": {
			denom:         "testcoin",
			grantee:       addr,
			administrator: addr2,
			allowance:     basic,
			valid:         true,
		},
		"no grantee": {
			administrator: addr2,
			denom:         "testcoin",
			grantee:       sdk.AccAddress{},
			allowance:     basic,
			valid:         false,
		},
		"no administrator": {
			administrator: sdk.AccAddress{},
			denom:         "testcoin",
			grantee:       addr,
			allowance:     basic,
			valid:         false,
		},
		"no denom": {
			administrator: sdk.AccAddress{},
			denom:         "",
			grantee:       addr,
			allowance:     basic,
			valid:         false,
		},
		"grantee == administrator": {
			denom:         "testcoin",
			grantee:       addr,
			administrator: addr,
			allowance:     basic,
			valid:         true,
		},
	}

	for _, tc := range cases {
		msg, err := NewMsgGrantAllowance(tc.denom, tc.administrator, tc.grantee, tc.allowance)
		require.NoError(t, err)
		err = msg.ValidateBasic()

		if tc.valid {
			require.NoError(t, err)

			allowance, err := msg.GetFeeAllowanceI()
			require.NoError(t, err)
			require.Equal(t, tc.allowance, allowance)

			err = msg.UnpackInterfaces(cdc)
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}
	}
}

func TestMsgRevokeGrantAllowance(t *testing.T) {
	addr, _ := sdk.AccAddressFromBech32("cosmos1aeuqja06474dfrj7uqsvukm6rael982kk89mqr")
	addr2, _ := sdk.AccAddressFromBech32("cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl")

	testCases := []struct {
		name          string
		denom         string
		grantee       sdk.AccAddress
		administrator sdk.AccAddress
		expErr        string
	}{
		{
			name:          "valid",
			denom:         "testcoin",
			grantee:       addr,
			administrator: addr2,
			expErr:        "",
		},
		{
			name:          "no grantee",
			denom:         "testcoin",
			grantee:       sdk.AccAddress{},
			administrator: addr2,
			expErr:        "missing grantee address",
		},
		{
			name:          "no administrator",
			denom:         "testcoin",
			grantee:       addr,
			administrator: sdk.AccAddress{},
			expErr:        "missing administrator address",
		},
		{
			name:          "no denom",
			denom:         "",
			grantee:       addr,
			administrator: sdk.AccAddress{},
			expErr:        "missing marker denom",
		},
		{
			name:          "grantee == administrator",
			denom:         "testcoin",
			grantee:       addr,
			administrator: addr,
			expErr:        "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewMsgRevokeGrantAllowance(tc.denom, tc.administrator, tc.grantee)

			// Check constructor output
			expected := &MsgRevokeGrantAllowanceRequest{
				Denom:         tc.denom,
				Administrator: tc.administrator.String(),
				Grantee:       tc.grantee.String(),
			}
			require.Equal(t, expected, got, "NewMsgRevokeGrantAllowance constructor output mismatch")

			err := got.ValidateBasic()
			if tc.expErr == "" {
				require.NoError(t, err, "ValidateBasic unexpected error in case: %s", tc.name)
			} else {
				require.Error(t, err, "ValidateBasic expected error in case: %s", tc.name)
				require.Contains(t, err.Error(), tc.expErr, "ValidateBasic error mismatch in case: %s", tc.name)
			}
		})
	}
}

func TestMsgIbcTransferRequestValidateBasic(t *testing.T) {
	validAddress := "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"

	cases := []struct {
		name     string
		msg      MsgIbcTransferRequest
		errorMsg string
	}{
		{
			"should fail to validate basic, invalid admin address",
			*NewMsgIbcTransferRequest(
				"notvalidaddress",
				"transfer",
				"channel-1",
				sdk.NewInt64Coin("jackthecat", 1),
				validAddress,
				validAddress,
				clienttypes.NewHeight(1, 1),
				1000,
				"",
			),
			"decoding bech32 failed: invalid separator index -1",
		},
		{
			"should fail to validate basic, invalid ibctransfertypes.MsgTransfer ",
			*NewMsgIbcTransferRequest(
				validAddress,
				"transfer",
				"channel-1",
				sdk.NewInt64Coin("jackthecat", 1),
				"invalid-address",
				validAddress,
				clienttypes.NewHeight(1, 1),
				1000,
				"",
			),
			"string could not be parsed as address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"should succeed",
			*NewMsgIbcTransferRequest(
				validAddress,
				"transfer",
				"channel-1",
				sdk.NewInt64Coin("jackthecat", 1),
				validAddress,
				validAddress,
				clienttypes.NewHeight(1, 1),
				1000,
				"",
			),
			"",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if len(tc.errorMsg) > 0 {
				require.Error(t, err)
				require.Equal(t, tc.errorMsg, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgAddMarkerRequestValidateBasic(t *testing.T) {
	validAddress := sdk.MustAccAddressFromBech32("cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck")

	cases := []struct {
		name     string
		msg      MsgAddMarkerRequest
		errorMsg string
	}{
		{
			name: "should fail on attributes for non restricted coin",
			msg: *NewMsgAddMarkerRequest(
				"hotdog",
				sdkmath.NewInt(100),
				validAddress,
				validAddress,
				MarkerType_Coin,
				true,
				true,
				false,
				[]string{"blah"},
				0,
				0,
			),
			errorMsg: "required attributes are reserved for restricted markers",
		},
		{
			name: "should succeed on attributes for restricted coin",
			msg: *NewMsgAddMarkerRequest(
				"hotdog",
				sdkmath.NewInt(100),
				validAddress,
				validAddress,
				MarkerType_RestrictedCoin,
				true,
				true,
				false,
				[]string{"blah"},
				0,
				0,
			),
			errorMsg: "",
		},
		{
			name: "should succeed on for restricted coin",
			msg: *NewMsgAddMarkerRequest(
				"hotdog",
				sdkmath.NewInt(100),
				validAddress,
				validAddress,
				MarkerType_RestrictedCoin,
				true,
				true,
				false,
				[]string{},
				0,
				0,
			),
			errorMsg: "",
		},
		{
			name: "should succeed on for non-restricted coin",
			msg: *NewMsgAddMarkerRequest(
				"hotdog",
				sdkmath.NewInt(100),
				validAddress,
				validAddress,
				MarkerType_Coin,
				true,
				true,
				false,
				[]string{},
				0,
				0,
			),
			errorMsg: "",
		},
		{
			name: "should fail duplicate entries for req attrs",
			msg: *NewMsgAddMarkerRequest(
				"hotdog",
				sdkmath.NewInt(100),
				validAddress,
				validAddress,
				MarkerType_RestrictedCoin,
				true,
				true,
				false,
				[]string{"foo", "foo"},
				0,
				0,
			),
			errorMsg: "required attribute list contains duplicate entries",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if len(tc.errorMsg) > 0 {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMsgAddFinalizeActivateMarkerRequestValidateBasic(t *testing.T) {
	validAddress := sdk.MustAccAddressFromBech32("cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck")

	cases := []struct {
		name     string
		msg      MsgAddFinalizeActivateMarkerRequest
		errorMsg string
	}{
		{
			name: "should fail on invalid marker",
			msg: MsgAddFinalizeActivateMarkerRequest{
				Amount: sdk.Coin{
					Amount: sdkmath.NewInt(100),
					Denom:  "",
				},
				Manager:                "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck",
				FromAddress:            "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck",
				MarkerType:             MarkerType_Coin,
				SupplyFixed:            true,
				AllowGovernanceControl: true,
				AccessList:             []AccessGrant{*NewAccessGrant(validAddress, []Access{Access_Mint, Access_Admin})},
			},
			errorMsg: "invalid marker denom/total supply: invalid coins",
		},
		{
			name: "should fail on invalid manager address",
			msg: MsgAddFinalizeActivateMarkerRequest{
				Amount:                 sdk.NewInt64Coin("hotdog", 100),
				Manager:                "",
				FromAddress:            "",
				MarkerType:             MarkerType_Coin,
				SupplyFixed:            true,
				AllowGovernanceControl: true,
				AccessList:             []AccessGrant{*NewAccessGrant(validAddress, []Access{Access_Mint, Access_Admin})},
			},
			errorMsg: "empty address string is not allowed",
		},
		{
			name: "should fail on empty access list",
			msg: *NewMsgAddFinalizeActivateMarkerRequest(
				"hotdog",
				sdkmath.NewInt(100),
				validAddress,
				validAddress,
				MarkerType_Coin,
				true,
				true,
				false,
				[]string{},
				[]AccessGrant{},
				0,
				0,
			),
			errorMsg: "since this will activate the marker, must have access list defined",
		},
		{
			name: "should fail on attributes for non restricted coin",
			msg: *NewMsgAddFinalizeActivateMarkerRequest(
				"hotdog",
				sdkmath.NewInt(100),
				validAddress,
				validAddress,
				MarkerType_Coin,
				true,
				true,
				false,
				[]string{"blah"},
				[]AccessGrant{*NewAccessGrant(validAddress, []Access{Access_Mint, Access_Admin})},
				0,
				0,
			),
			errorMsg: "required attributes are reserved for restricted markers",
		},
		{
			name: "should succeed",
			msg: *NewMsgAddFinalizeActivateMarkerRequest(
				"hotdog",
				sdkmath.NewInt(100),
				validAddress,
				validAddress,
				MarkerType_Coin,
				true,
				true,
				false,
				[]string{},
				[]AccessGrant{*NewAccessGrant(validAddress, []Access{Access_Mint, Access_Admin})},
				0,
				0,
			),
			errorMsg: "",
		},
		{
			name: "should succeed for restricted coin with required attributes",
			msg: *NewMsgAddFinalizeActivateMarkerRequest(
				"hotdog",
				sdkmath.NewInt(100),
				validAddress,
				validAddress,
				MarkerType_RestrictedCoin,
				true,
				true,
				false,
				[]string{"blah"},
				[]AccessGrant{*NewAccessGrant(validAddress, []Access{Access_Mint, Access_Admin})},
				0,
				0,
			),
			errorMsg: "",
		},
		{
			name: "should fail when forced tranfers allowed with coin type",
			msg: *NewMsgAddFinalizeActivateMarkerRequest(
				"banana",
				sdkmath.NewInt(500),
				validAddress,
				validAddress,
				MarkerType_Coin,
				true,
				true,
				true,
				[]string{},
				[]AccessGrant{*NewAccessGrant(validAddress, []Access{Access_Mint, Access_Admin})},
				0,
				0,
			),
			errorMsg: "forced transfer is only available for restricted coins",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if len(tc.errorMsg) > 0 {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMsgSupplyIncreaseProposalRequestValidateBasic(t *testing.T) {
	authority := sdk.AccAddress("input111111111111111").String()
	targetAddress := sdk.AccAddress("input22222222222").String()

	testCases := []struct {
		name          string
		amount        sdk.Coin
		targetAddress string
		authority     string
		shouldFail    bool
		expectedError string
	}{
		{
			name: "negative coin amount",
			amount: sdk.Coin{
				Amount: sdkmath.NewInt(-1),
				Denom:  "invalid-denom",
			},
			targetAddress: targetAddress,
			authority:     authority,
			shouldFail:    true,
			expectedError: "negative coin amount: -1",
		},
		{
			name: "invalid target address",
			amount: sdk.Coin{
				Amount: sdkmath.NewInt(100),
				Denom:  "bbq-hotdog",
			},
			targetAddress: "invalidaddress",
			authority:     authority,
			shouldFail:    true,
			expectedError: "decoding bech32 failed: invalid separator index -1",
		},
		{
			name: "valid with target address",
			amount: sdk.Coin{
				Amount: sdkmath.NewInt(100),
				Denom:  "bbq-hotdog",
			},
			targetAddress: targetAddress,
			authority:     authority,
			shouldFail:    false,
		},
		{
			name: "valid without target address",
			amount: sdk.Coin{
				Amount: sdkmath.NewInt(100),
				Denom:  "bbq-hotdog",
			},
			targetAddress: "",
			authority:     authority,
			shouldFail:    false,
		},
		{
			name: "invalid authority",
			amount: sdk.Coin{
				Amount: sdkmath.NewInt(100),
				Denom:  "bbq-hotdog",
			},
			targetAddress: targetAddress,
			authority:     "invalidaddress",
			shouldFail:    true,
			expectedError: "decoding bech32 failed: invalid separator index -1",
		},
		{
			name: "empty authority",
			amount: sdk.Coin{
				Amount: sdkmath.NewInt(100),
				Denom:  "bbq-hotdog",
			},
			targetAddress: targetAddress,
			authority:     "",
			shouldFail:    true,
			expectedError: "empty address string is not allowed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := NewMsgSupplyIncreaseProposalRequest(tc.amount, tc.targetAddress, tc.authority)
			err := msg.ValidateBasic()

			if tc.shouldFail {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMsgUpdateRequiredAttributesRequestValidateBasic(t *testing.T) {
	authority := sdk.AccAddress("input111111111111111").String()

	testCases := []struct {
		name          string
		msg           MsgUpdateRequiredAttributesRequest
		expectedError string
	}{
		{
			name:          "should fail, invalid denom",
			msg:           *NewMsgUpdateRequiredAttributesRequest("#&", sdk.AccAddress(authority), []string{"foo.provenance.io"}, []string{"foo2.provenance.io"}),
			expectedError: "invalid denom: #&",
		},
		{
			name:          "should fail, invalid address",
			msg:           MsgUpdateRequiredAttributesRequest{Denom: "jackthecat", TransferAuthority: "invalid-addrr", AddRequiredAttributes: []string{"foo.provenance.io"}, RemoveRequiredAttributes: []string{"foo2.provenance.io"}},
			expectedError: "decoding bech32 failed: invalid separator index -1",
		},
		{
			name:          "should fail, both add and remove list are empty",
			msg:           *NewMsgUpdateRequiredAttributesRequest("jackthecat", sdk.AccAddress(authority), []string{}, []string{}),
			expectedError: "both add and remove lists cannot be empty",
		},
		{
			name:          "should fail, combined list has duplicate entries",
			msg:           *NewMsgUpdateRequiredAttributesRequest("jackthecat", sdk.AccAddress(authority), []string{"foo.provenance.io"}, []string{"foo.provenance.io"}),
			expectedError: "required attribute lists contain duplicate entries",
		},
		{
			name:          "should fail, add list has duplicate entries",
			msg:           *NewMsgUpdateRequiredAttributesRequest("jackthecat", sdk.AccAddress(authority), []string{"foo.provenance.io", "foo.provenance.io"}, []string{"foo2.provenance.io"}),
			expectedError: "required attribute lists contain duplicate entries",
		},
		{
			name:          "should fail, remove list has duplicate entries",
			msg:           *NewMsgUpdateRequiredAttributesRequest("jackthecat", sdk.AccAddress(authority), []string{"foo.provenance.io"}, []string{"foo2.provenance.io", "foo2.provenance.io"}),
			expectedError: "required attribute lists contain duplicate entries",
		},
		{
			name: "should succeed",
			msg:  *NewMsgUpdateRequiredAttributesRequest("jackthecat", sdk.AccAddress(authority), []string{"foo.provenance.io"}, []string{"foo2.provenance.io"}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if len(tc.expectedError) > 0 {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMsgUpdateForcedTransferRequestValidateBasic(t *testing.T) {
	goodAuthority := sdk.AccAddress("goodAddr____________").String()
	goodDenom := "gooddenom"
	tests := []struct {
		name string
		msg  *MsgUpdateForcedTransferRequest
		exp  string
	}{
		{
			name: "invalid denom",
			msg: &MsgUpdateForcedTransferRequest{
				Denom:               "x",
				AllowForcedTransfer: false,
				Authority:           goodAuthority,
			},
			exp: "invalid denom: x",
		},
		{
			name: "invalid authority",
			msg: &MsgUpdateForcedTransferRequest{
				Denom:               goodDenom,
				AllowForcedTransfer: false,
				Authority:           "x",
			},
			exp: "invalid authority: decoding bech32 failed: invalid bech32 string length 1",
		},
		{
			name: "ok forced transfer true",
			msg: &MsgUpdateForcedTransferRequest{
				Denom:               goodDenom,
				AllowForcedTransfer: true,
				Authority:           goodAuthority,
			},
			exp: "",
		},
		{
			name: "ok forced transfer false",
			msg: &MsgUpdateForcedTransferRequest{
				Denom:               goodDenom,
				AllowForcedTransfer: false,
				Authority:           goodAuthority,
			},
			exp: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if len(tc.exp) > 0 {
				assert.EqualError(t, err, tc.exp, "ValidateBasic error")
			} else {
				assert.NoError(t, err, tc.exp, "ValidateBasic error")
			}
		})
	}
}

func TestMsgSetAccountDataRequestValidateBasic(t *testing.T) {
	addr := sdk.AccAddress("addr________________").String()
	denom := "somedenom"

	tests := []struct {
		name string
		msg  MsgSetAccountDataRequest
		exp  string
	}{
		{
			name: "control",
			msg:  MsgSetAccountDataRequest{Denom: denom, Signer: addr},
			exp:  "",
		},
		{
			name: "no denom",
			msg:  MsgSetAccountDataRequest{Denom: "", Signer: addr},
			exp:  "invalid denom: empty denom string is not allowed",
		},
		{
			name: "invalid denom",
			msg:  MsgSetAccountDataRequest{Denom: "1denomcannotstartwithdigit", Signer: addr},
			exp:  "invalid denom: 1denomcannotstartwithdigit",
		},
		{
			name: "no signer",
			msg:  MsgSetAccountDataRequest{Denom: denom, Signer: ""},
			exp:  "invalid signer: empty address string is not allowed",
		},
		{
			name: "invalid signer",
			msg:  MsgSetAccountDataRequest{Denom: denom, Signer: "not1validsigner"},
			exp:  "invalid signer: decoding bech32 failed: invalid character not part of charset: 105",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if len(tc.exp) > 0 {
				require.EqualErrorf(t, err, tc.exp, "ValidateBasic error")
			} else {
				require.NoError(t, err, "ValidateBasic error")
			}
		})
	}
}

func TestMsgUpdateSendDenyListRequestValidateBasic(t *testing.T) {
	addr := sdk.AccAddress("addr________________").String()
	denom := "somedenom"
	addAddr := sdk.AccAddress("addAddr________________").String()
	removeAddr := sdk.AccAddress("removeAddr________________").String()

	tests := []struct {
		name   string
		msg    MsgUpdateSendDenyListRequest
		expErr string
	}{
		{
			name: "should succeed",
			msg:  MsgUpdateSendDenyListRequest{Denom: denom, RemoveDeniedAddresses: []string{removeAddr}, AddDeniedAddresses: []string{addAddr}, Authority: addr},
		},
		{
			name:   "invalid authority address",
			msg:    MsgUpdateSendDenyListRequest{Denom: denom, RemoveDeniedAddresses: []string{removeAddr}, AddDeniedAddresses: []string{addAddr}, Authority: "invalid-address"},
			expErr: "decoding bech32 failed: invalid separator index -1",
		},
		{
			name:   "both add and remove list are empty",
			msg:    MsgUpdateSendDenyListRequest{Denom: denom, RemoveDeniedAddresses: []string{}, AddDeniedAddresses: []string{}, Authority: addr},
			expErr: "both add and remove lists cannot be empty",
		},
		{
			name:   "invalid authority address",
			msg:    MsgUpdateSendDenyListRequest{Denom: denom, RemoveDeniedAddresses: []string{removeAddr}, AddDeniedAddresses: []string{addAddr}, Authority: "invalid-address"},
			expErr: "decoding bech32 failed: invalid separator index -1",
		},
		{
			name:   "invalid remove address",
			msg:    MsgUpdateSendDenyListRequest{Denom: denom, RemoveDeniedAddresses: []string{"invalid-address"}, AddDeniedAddresses: []string{}, Authority: addr},
			expErr: "decoding bech32 failed: invalid separator index -1",
		},
		{
			name:   "invalid add address",
			msg:    MsgUpdateSendDenyListRequest{Denom: denom, RemoveDeniedAddresses: []string{removeAddr}, AddDeniedAddresses: []string{"invalid-addrs"}, Authority: addr},
			expErr: "decoding bech32 failed: invalid separator index -1",
		},
		{
			name:   "duplicate entries in list",
			msg:    MsgUpdateSendDenyListRequest{Denom: denom, RemoveDeniedAddresses: []string{removeAddr, removeAddr}, AddDeniedAddresses: []string{}, Authority: addr},
			expErr: "denied address lists contain duplicate entries",
		},
		{
			name:   "invalid denom",
			msg:    MsgUpdateSendDenyListRequest{Denom: "1", RemoveDeniedAddresses: []string{removeAddr}, AddDeniedAddresses: []string{addAddr}, Authority: addr},
			expErr: "invalid denom: 1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if len(tc.expErr) > 0 {
				require.EqualErrorf(t, err, tc.expErr, "ValidateBasic error")
			} else {
				require.NoError(t, err, "ValidateBasic error")
			}
		})
	}
}

func TestMsgAddNetAssetValueValidateBasic(t *testing.T) {
	addr := sdk.AccAddress("addr________________").String()
	denom := "somedenom"
	netAssetValue1 := NetAssetValue{Price: sdk.NewInt64Coin("jackthecat", 100), Volume: uint64(100)}
	netAssetValue2 := NetAssetValue{Price: sdk.NewInt64Coin("hotdog", 100), Volume: uint64(100)}
	invalidNetAssetValue := NetAssetValue{Price: sdk.NewInt64Coin("hotdog", 100), Volume: uint64(0)}
	invalidNetAssetValue2 := NetAssetValue{Price: sdk.NewInt64Coin("hotdog", 100), Volume: uint64(1), UpdatedBlockHeight: 1}

	tests := []struct {
		name   string
		msg    MsgAddNetAssetValuesRequest
		expErr string
	}{
		{
			name: "should succeed",
			msg:  MsgAddNetAssetValuesRequest{Denom: denom, NetAssetValues: []NetAssetValue{netAssetValue1, netAssetValue2}, Administrator: addr},
		},
		{
			name:   "block height is set",
			msg:    MsgAddNetAssetValuesRequest{Denom: denom, NetAssetValues: []NetAssetValue{invalidNetAssetValue2}, Administrator: addr},
			expErr: "marker net asset value must not have update height set",
		},
		{
			name:   "validation of net asset value failure",
			msg:    MsgAddNetAssetValuesRequest{Denom: denom, NetAssetValues: []NetAssetValue{invalidNetAssetValue}, Administrator: addr},
			expErr: "marker net asset value volume must be positive value",
		},
		{
			name:   "duplicate net asset values",
			msg:    MsgAddNetAssetValuesRequest{Denom: denom, NetAssetValues: []NetAssetValue{netAssetValue1, netAssetValue2, netAssetValue2}, Administrator: addr},
			expErr: "list of net asset values contains duplicates",
		},
		{
			name:   "invalid denom",
			msg:    MsgAddNetAssetValuesRequest{Denom: "", NetAssetValues: []NetAssetValue{netAssetValue1, netAssetValue2, netAssetValue2}, Administrator: addr},
			expErr: "invalid denom: ",
		},
		{
			name:   "invalid administrator address",
			msg:    MsgAddNetAssetValuesRequest{Denom: denom, NetAssetValues: []NetAssetValue{netAssetValue1, netAssetValue2}, Administrator: "invalid address"},
			expErr: "decoding bech32 failed: invalid character in string: ' '",
		},
		{
			name:   "empty net asset list",
			msg:    MsgAddNetAssetValuesRequest{Denom: denom, NetAssetValues: []NetAssetValue{}, Administrator: addr},
			expErr: "net asset value list cannot be empty",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if len(tc.expErr) > 0 {
				require.EqualErrorf(t, err, tc.expErr, "ValidateBasic error")
			} else {
				require.NoError(t, err, "ValidateBasic error")
			}
		})
	}
}

func TestMsgSupplyDecreaseProposalRequestValidateBasic(t *testing.T) {
	validAddress := "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"
	invalidAddress := "invalidaddr0000"

	testCases := []struct {
		name          string
		authority     string
		amount        sdk.Coin
		expectError   bool
		expectedError string
	}{
		{
			name:          "valid input",
			authority:     validAddress,
			amount:        sdk.NewInt64Coin("testcoin", 100),
			expectError:   false,
			expectedError: "",
		},
		{
			name:          "negative amount",
			authority:     validAddress,
			amount:        sdk.Coin{Denom: "testcoin", Amount: sdkmath.NewInt(-100)},
			expectError:   true,
			expectedError: "amount to decrease must be greater than zero",
		},
		{
			name:          "invalid authority address",
			authority:     invalidAddress,
			amount:        sdk.NewInt64Coin("testcoin", 100),
			expectError:   true,
			expectedError: "decoding bech32 failed: invalid separator index -1",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			msg := MsgSupplyDecreaseProposalRequest{
				Authority: tc.authority,
				Amount:    tc.amount,
			}

			err := msg.ValidateBasic()
			if tc.expectError {
				require.Error(t, err)
				require.EqualError(t, err, tc.expectedError, "ValidateBasic error")
			} else {
				require.NoError(t, err, "ValidateBasic error")
			}
		})
	}
}

func TestMsgSetAdministratorProposalRequestValidateBasic(t *testing.T) {
	validAddress := "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"
	invalidAddress := "invalidaddr0000"

	validAccessGrant := AccessGrant{
		Address:     validAddress,
		Permissions: []Access{Access_Admin},
	}
	invalidAccessGrant := AccessGrant{
		Address:     "invalidaddress",
		Permissions: []Access{Access_Admin},
	}

	testCases := []struct {
		name          string
		denom         string
		accessGrant   []AccessGrant
		authority     string
		expectError   bool
		expectedError string
	}{
		{
			name:        "valid case",
			denom:       "testcoin",
			accessGrant: []AccessGrant{validAccessGrant},
			authority:   validAddress,
			expectError: false,
		},
		{
			name:          "invalid authority address",
			denom:         "testcoin",
			accessGrant:   []AccessGrant{validAccessGrant},
			authority:     invalidAddress,
			expectError:   true,
			expectedError: "decoding bech32 failed: invalid separator index -1",
		},
		{
			name:          "invalid access grant",
			denom:         "testcoin",
			accessGrant:   []AccessGrant{invalidAccessGrant},
			authority:     validAddress,
			expectError:   true,
			expectedError: "invalid access grant for administrator: invalid address: decoding bech32 failed: invalid separator index -1",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			msg := NewMsgSetAdministratorProposalRequest(tc.denom, tc.accessGrant, tc.authority)

			err := msg.ValidateBasic()
			if tc.expectError {
				require.Error(t, err)
				require.EqualError(t, err, tc.expectedError, "ValidateBasic error")
			} else {
				require.NoError(t, err, "ValidateBasic error")
			}
		})
	}
}

func TestMsgRemoveAdministratorProposalRequestValidateBasic(t *testing.T) {
	validAuthority := "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"
	invalidAuthority := "invalidauth0000"

	validRemovedAddress := "cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl"
	invalidRemovedAddress := "invalidremoved000"

	testCases := []struct {
		name           string
		authority      string
		removedAddress []string
		expectError    bool
		expectedError  string
	}{
		{
			name:           "valid case",
			authority:      validAuthority,
			removedAddress: []string{validRemovedAddress},
			expectError:    false,
		},
		{
			name:           "invalid authority address",
			authority:      invalidAuthority,
			removedAddress: []string{validRemovedAddress},
			expectError:    true,
			expectedError:  "decoding bech32 failed: invalid separator index -1",
		},
		{
			name:           "invalid removed address",
			authority:      validAuthority,
			removedAddress: []string{invalidRemovedAddress},
			expectError:    true,
			expectedError:  "administrator account address is invalid: decoding bech32 failed: invalid separator index -1",
		},
		{
			name:           "multiple removed addresses with an invalid one",
			authority:      validAuthority,
			removedAddress: []string{validRemovedAddress, invalidRemovedAddress},
			expectError:    true,
			expectedError:  "administrator account address is invalid: decoding bech32 failed: invalid separator index -1",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			msg := MsgRemoveAdministratorProposalRequest{
				Authority:      tc.authority,
				RemovedAddress: tc.removedAddress,
			}

			err := msg.ValidateBasic()
			if tc.expectError {
				require.Error(t, err)
				require.EqualError(t, err, tc.expectedError, "ValidateBasic error")
			} else {
				require.NoError(t, err, "ValidateBasic error")
			}
		})
	}
}

func TestMsgChangeStatusProposalRequestValidateBasic(t *testing.T) {
	validAuthority := "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"
	invalidAuthority := "invalidauth0000"
	validDenom := "validcoin"
	invalidDenom := "1invalid"

	testCases := []struct {
		name          string
		denom         string
		authority     string
		expectError   bool
		expectedError string
	}{
		{
			name:        "valid case",
			denom:       validDenom,
			authority:   validAuthority,
			expectError: false,
		},
		{
			name:          "invalid authority address",
			denom:         validDenom,
			authority:     invalidAuthority,
			expectError:   true,
			expectedError: "decoding bech32 failed: invalid separator index -1",
		},
		{
			name:          "invalid denom",
			denom:         invalidDenom,
			authority:     validAuthority,
			expectError:   true,
			expectedError: "invalid denom: 1invalid",
		},
		{
			name:          "both authority and denom are invalid",
			denom:         invalidDenom,
			authority:     invalidAuthority,
			expectError:   true,
			expectedError: "invalid denom: 1invalid",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			msg := MsgChangeStatusProposalRequest{
				Denom:     tc.denom,
				Authority: tc.authority,
			}

			err := msg.ValidateBasic()
			if tc.expectError {
				require.Error(t, err)
				require.EqualError(t, err, tc.expectedError, "ValidateBasic error")
			} else {
				require.NoError(t, err, "ValidateBasic error")
			}
		})
	}
}

func TestMsgWithdrawEscrowProposalRequestValidateBasic(t *testing.T) {
	validAuthority := "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"
	invalidAuthority := "invalidauth0000"
	validTargetAddress := "cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl"
	invalidTargetAddress := "invalidtarget0000"
	validDenom := "validcoin"
	validAmount := sdk.NewCoins(sdk.NewInt64Coin(validDenom, 100))
	invalidAmount := sdk.Coins{sdk.Coin{Denom: validDenom, Amount: sdkmath.NewInt(-100)}} // Negative amount

	testCases := []struct {
		name          string
		denom         string
		amount        sdk.Coins
		targetAddress string
		authority     string
		expectError   bool
		expectedError string
	}{
		{
			name:          "valid case",
			denom:         validDenom,
			amount:        validAmount,
			targetAddress: validTargetAddress,
			authority:     validAuthority,
			expectError:   false,
		},
		{
			name:          "invalid amount",
			denom:         validDenom,
			amount:        invalidAmount,
			targetAddress: validTargetAddress,
			authority:     validAuthority,
			expectError:   true,
			expectedError: fmt.Sprintf("amount is invalid: %v", invalidAmount),
		},
		{
			name:          "invalid target address",
			denom:         validDenom,
			amount:        validAmount,
			targetAddress: invalidTargetAddress,
			authority:     validAuthority,
			expectError:   true,
			expectedError: "decoding bech32 failed: invalid separator index -1",
		},
		{
			name:          "invalid authority address",
			denom:         validDenom,
			amount:        validAmount,
			targetAddress: validTargetAddress,
			authority:     invalidAuthority,
			expectError:   true,
			expectedError: "decoding bech32 failed: invalid separator index -1",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			msg := MsgWithdrawEscrowProposalRequest{
				Denom:         tc.denom,
				Amount:        tc.amount,
				TargetAddress: tc.targetAddress,
				Authority:     tc.authority,
			}

			err := msg.ValidateBasic()
			if tc.expectError {
				require.Error(t, err)
				require.EqualError(t, err, tc.expectedError, "ValidateBasic error")
			} else {
				require.NoError(t, err, "ValidateBasic error")
			}
		})
	}
}

func TestMsgSetDenomMetadataProposalRequestValidateBasic(t *testing.T) {
	validAuthority := "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"
	invalidAuthority := "invalidauth0000"
	hotdogDenom := "hotdog"

	validMetadata := banktypes.Metadata{
		Description: "a description",
		DenomUnits: []*banktypes.DenomUnit{
			{Denom: fmt.Sprintf("n%s", hotdogDenom), Exponent: 0, Aliases: []string{fmt.Sprintf("nano%s", hotdogDenom)}},
			{Denom: fmt.Sprintf("u%s", hotdogDenom), Exponent: 3, Aliases: []string{}},
			{Denom: hotdogDenom, Exponent: 9, Aliases: []string{}},
			{Denom: fmt.Sprintf("mega%s", hotdogDenom), Exponent: 15, Aliases: []string{}},
		},
		Base:    fmt.Sprintf("n%s", hotdogDenom),
		Display: hotdogDenom,
		Name:    "hotdogName",
		Symbol:  "WIFI",
	}
	invalidMetadata := banktypes.Metadata{
		Name:        "",
		Description: "Description.",
	}

	testCases := []struct {
		name          string
		metadata      banktypes.Metadata
		authority     string
		expectError   bool
		expectedError string
	}{
		{
			name:        "valid case",
			metadata:    validMetadata,
			authority:   validAuthority,
			expectError: false,
		},
		{
			name:          "invalid metadata",
			metadata:      invalidMetadata,
			authority:     validAuthority,
			expectError:   true,
			expectedError: "name field cannot be blank",
		},
		{
			name:          "invalid authority address",
			metadata:      validMetadata,
			authority:     invalidAuthority,
			expectError:   true,
			expectedError: "decoding bech32 failed: invalid separator index -1",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			msg := MsgSetDenomMetadataProposalRequest{
				Metadata:  tc.metadata,
				Authority: tc.authority,
			}

			err := msg.ValidateBasic()
			if tc.expectError {
				require.Error(t, err)
				require.EqualError(t, err, tc.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgUpdateParamsRequestValidateBasic(t *testing.T) {
	validAuthority := "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"
	invalidAuthority := "invalidaddress"

	testCases := []struct {
		name          string
		msg           MsgUpdateParamsRequest
		expectError   bool
		expectedError string
	}{
		{
			name: "valid case",
			msg: MsgUpdateParamsRequest{
				Authority: validAuthority,
				Params: NewParams(
					true,
					"[a-zA-Z][a-zA-Z0-9\\-\\.]{2,83}",
					sdkmath.NewInt(1000000000000),
				),
			},
			expectError: false,
		},
		{
			name: "invalid regex",
			msg: MsgUpdateParamsRequest{
				Authority: validAuthority,
				Params: NewParams(
					true,
					"^invalidregex$",
					sdkmath.NewInt(1000000000000),
				),
			},
			expectError:   true,
			expectedError: "invalid parameter, validation regex must not contain anchors ^,$",
		},
		{
			name: "invalid authority",
			msg: MsgUpdateParamsRequest{
				Authority: invalidAuthority,
				Params: NewParams(
					true,
					"[a-zA-Z][a-zA-Z0-9\\-\\.]{2,83}",
					sdkmath.NewInt(1000000000000),
				),
			},
			expectError:   true,
			expectedError: "decoding bech32 failed: invalid separator index -1",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.expectError {
				require.Error(t, err, "expected error but got none for case: %s", tc.name)
				require.EqualError(t, err, tc.expectedError, "unexpected error message for case: %s", tc.name)
			} else {
				require.NoError(t, err, "unexpected error for case: %s", tc.name)
			}
		})
	}
}
