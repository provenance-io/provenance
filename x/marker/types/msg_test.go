package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"

	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
)

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

			addrSlice := msg.GetSigners()
			require.Equal(t, tc.administrator.String(), addrSlice[0].String())

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

func TestMsgIbcTransferRequestValidateBasic(t *testing.T) {
	validAddress := "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"

	cases := []struct {
		name     string
		msg      MsgIbcTransferRequest
		errorMsg string
	}{
		{
			"should fail to validate basic, invalid admin address",
			*NewIbcMsgTransferRequest(
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
			*NewIbcMsgTransferRequest(
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
			*NewIbcMsgTransferRequest(
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
				sdk.NewInt(100),
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
				sdk.NewInt(100),
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
				sdk.NewInt(100),
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
				sdk.NewInt(100),
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
				sdk.NewInt(100),
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
					Amount: math.NewInt(100),
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
				sdk.NewInt(100),
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
				sdk.NewInt(100),
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
				sdk.NewInt(100),
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
				sdk.NewInt(100),
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
				sdk.NewInt(500),
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

func TestMsgSupplyIncreaseProposalRequestGetSigners(t *testing.T) {
	authority := sdk.AccAddress("input111111111111111")
	targetAddress := sdk.AccAddress("input22222222222")
	amount :=
		sdk.Coin{
			Amount: math.NewInt(100),
			Denom:  "chocolate",
		}

	msg := NewMsgSupplyIncreaseProposalRequest(amount, targetAddress.String(), authority.String())
	res := msg.GetSigners()
	exp := []sdk.AccAddress{authority}
	require.Equal(t, exp, res, "GetSigners")
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
				Amount: math.NewInt(-1),
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
				Amount: math.NewInt(100),
				Denom:  "bbq-hotdog",
			},
			targetAddress: "",
			authority:     authority,
			shouldFail:    true,
			expectedError: "empty address string is not allowed",
		},
		{
			name: "invalid authority",
			amount: sdk.Coin{
				Amount: math.NewInt(100),
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

func TestMsgUpdateForcedTransferRequestGetSigners(t *testing.T) {
	t.Run("good authority", func(t *testing.T) {
		msg := MsgUpdateForcedTransferRequest{
			Authority: sdk.AccAddress("good_address________").String(),
		}
		exp := []sdk.AccAddress{sdk.AccAddress("good_address________")}

		var signers []sdk.AccAddress
		testFunc := func() {
			signers = msg.GetSigners()
		}
		require.NotPanics(t, testFunc, "GetSigners")
		assert.Equal(t, exp, signers, "GetSigners")
	})

	t.Run("bad authority", func(t *testing.T) {
		msg := MsgUpdateForcedTransferRequest{
			Authority: "bad_address________",
		}

		testFunc := func() {
			_ = msg.GetSigners()
		}
		require.PanicsWithError(t, "decoding bech32 failed: invalid separator index -1", testFunc, "GetSigners")
	})
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

func TestMsgSetAccountDataRequestGetSigners(t *testing.T) {
	t.Run("good signer", func(t *testing.T) {
		msg := MsgSetAccountDataRequest{
			Signer: sdk.AccAddress("good_address________").String(),
		}
		exp := []sdk.AccAddress{sdk.AccAddress("good_address________")}

		var signers []sdk.AccAddress
		testFunc := func() {
			signers = msg.GetSigners()
		}
		require.NotPanics(t, testFunc, "GetSigners")
		assert.Equal(t, exp, signers, "GetSigners")
	})

	t.Run("bad signer", func(t *testing.T) {
		msg := MsgSetAccountDataRequest{
			Signer: "bad_address________",
		}

		testFunc := func() {
			_ = msg.GetSigners()
		}
		require.PanicsWithError(t, "decoding bech32 failed: invalid separator index -1", testFunc, "GetSigners")
	})
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

func TestMsgUpdateSendDenyListRequestGetSigners(t *testing.T) {
	t.Run("good signer", func(t *testing.T) {
		msg := MsgUpdateSendDenyListRequest{
			Authority: sdk.AccAddress("good_address________").String(),
		}
		exp := []sdk.AccAddress{sdk.AccAddress("good_address________")}

		var signers []sdk.AccAddress
		testFunc := func() {
			signers = msg.GetSigners()
		}
		require.NotPanics(t, testFunc, "GetSigners")
		assert.Equal(t, exp, signers, "GetSigners")
	})

	t.Run("bad signer", func(t *testing.T) {
		msg := MsgUpdateSendDenyListRequest{
			Authority: "bad_address________",
		}

		testFunc := func() {
			_ = msg.GetSigners()
		}
		require.PanicsWithError(t, "decoding bech32 failed: invalid separator index -1", testFunc, "GetSigners")
	})
}

func TestMsgAddNetAssetValueValidateBasic(t *testing.T) {
	addr := sdk.AccAddress("addr________________").String()
	denom := "somedenom"
	netAssetValue1 := NetAssetValue{Value: sdk.NewInt64Coin("jackthecat", 100), Volume: uint64(100)}
	netAssetValue2 := NetAssetValue{Value: sdk.NewInt64Coin("hotdog", 100), Volume: uint64(100)}
	invalidNetAssetValue := NetAssetValue{Value: sdk.NewInt64Coin("hotdog", 100), Volume: uint64(0)}
	invalidNetAssetValue2 := NetAssetValue{Value: sdk.NewInt64Coin("hotdog", 100), Volume: uint64(1), UpdatedBlockHeight: 1}

	tests := []struct {
		name   string
		msg    MsgAddNetAssetValueRequest
		expErr string
	}{
		{
			name: "should succeed",
			msg:  MsgAddNetAssetValueRequest{Denom: denom, NetAssetValues: []NetAssetValue{netAssetValue1, netAssetValue2}, Administrator: addr},
		},
		{
			name:   "block height is set",
			msg:    MsgAddNetAssetValueRequest{Denom: denom, NetAssetValues: []NetAssetValue{invalidNetAssetValue2}, Administrator: addr},
			expErr: "marker net asset value must not have update height set",
		},
		{
			name:   "validation of net asset value failure",
			msg:    MsgAddNetAssetValueRequest{Denom: denom, NetAssetValues: []NetAssetValue{invalidNetAssetValue}, Administrator: addr},
			expErr: "marker net asset value volume must be positive value",
		},
		{
			name:   "duplicate net asset values",
			msg:    MsgAddNetAssetValueRequest{Denom: denom, NetAssetValues: []NetAssetValue{netAssetValue1, netAssetValue2, netAssetValue2}, Administrator: addr},
			expErr: "list of net asset values contains duplicates",
		},
		{
			name:   "invalid denom",
			msg:    MsgAddNetAssetValueRequest{Denom: "", NetAssetValues: []NetAssetValue{netAssetValue1, netAssetValue2, netAssetValue2}, Administrator: addr},
			expErr: "invalid denom: ",
		},
		{
			name:   "invalid administrator address",
			msg:    MsgAddNetAssetValueRequest{Denom: denom, NetAssetValues: []NetAssetValue{netAssetValue1, netAssetValue2}, Administrator: "invalid address"},
			expErr: "decoding bech32 failed: invalid character in string: ' '",
		},
		{
			name:   "empty net asset list",
			msg:    MsgAddNetAssetValueRequest{Denom: denom, NetAssetValues: []NetAssetValue{}, Administrator: addr},
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

func TestMsgAddNetAssetValueRequestGetSigners(t *testing.T) {
	t.Run("good signer", func(t *testing.T) {
		msg := MsgAddNetAssetValueRequest{
			Administrator: sdk.AccAddress("good_address________").String(),
		}
		exp := []sdk.AccAddress{sdk.AccAddress("good_address________")}

		var signers []sdk.AccAddress
		testFunc := func() {
			signers = msg.GetSigners()
		}
		require.NotPanics(t, testFunc, "GetSigners")
		assert.Equal(t, exp, signers, "GetSigners")
	})

	t.Run("bad signer", func(t *testing.T) {
		msg := MsgAddNetAssetValueRequest{
			Administrator: "bad_address________",
		}

		testFunc := func() {
			_ = msg.GetSigners()
		}
		require.PanicsWithError(t, "decoding bech32 failed: invalid separator index -1", testFunc, "GetSigners")
	})
}
