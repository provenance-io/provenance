package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"

	clienttypes "github.com/cosmos/ibc-go/v5/modules/core/02-client/types"
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

func TestMsgAssessCustomMsgFeeValidateBasic(t *testing.T) {
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

func TestMsgReflectMarkerValidateBasic(t *testing.T) {
	validAddress := "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"

	cases := []struct {
		name     string
		msg      MsgReflectMarkerRequest
		errorMsg string
	}{
		{
			"should fail to validate basic, invalid admin address",
			*NewMsgReflectMarkerRequest(
				"markder-denom",
				"ibc-denom",
				"connection-id",
				"invalid-address",
			),
			"decoding bech32 failed: invalid separator index -1",
		},
		{
			"should pass",
			*NewMsgReflectMarkerRequest(
				"markder-denom",
				"ibc-denom",
				"connection-id",
				validAddress,
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

func TestMsgIcaReflectMarkerValidateBasic(t *testing.T) {
	validAddress := "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"

	cases := []struct {
		name     string
		msg      MsgIcaReflectMarkerRequest
		errorMsg string
	}{
		{
			"should fail to validate basic, invalid marker status",
			*NewMsgIcaReflectMarkerRequest(
				"marker-denom",
				"ibc-denom",
				"invoker",
				"owner",
				StatusUndefined,
				MarkerType_Unknown,
				[]AccessGrant{},
				false,
			),
			"reflected marker must have active status",
		},
		{
			"should fail to validate basic, invalid marker type",
			*NewMsgIcaReflectMarkerRequest(
				"marker-denom",
				"ibc-denom",
				"invoker",
				"owner",
				StatusActive,
				MarkerType_Unknown,
				[]AccessGrant{},
				false,
			),
			"reflected marker must not have unknown type",
		},
		{
			"should fail to validate basic, invalid invoker address",
			*NewMsgIcaReflectMarkerRequest(
				"marker-denom",
				"ibc-denom",
				"invalid-address",
				"owner",
				StatusActive,
				MarkerType_RestrictedCoin,
				[]AccessGrant{},
				false,
			),
			"decoding bech32 failed: invalid separator index -1",
		},
		{
			"should fail to validate basic, invalid admin address",
			*NewMsgIcaReflectMarkerRequest(
				"marker-denom",
				"ibc-denom",
				"invalid-address",
				"owner",
				StatusActive,
				MarkerType_RestrictedCoin,
				[]AccessGrant{
					{
						"Address",
						[]Access{
							Access_Burn,
						},
					},
				},
				false,
			),
			"access list contains mint and/or burn",
		},
		{
			"should fail to validate basic, invalid permissions",
			*NewMsgIcaReflectMarkerRequest(
				"marker-denom",
				"ibc-denom",
				"invalid-address",
				"owner",
				StatusActive,
				MarkerType_RestrictedCoin,
				[]AccessGrant{
					{
						"Address",
						[]Access{
							Access_Mint,
						},
					},
				},
				false,
			),
			"access list contains mint and/or burn",
		},
		{
			"should fail to validate basic, invalid permissions",
			*NewMsgIcaReflectMarkerRequest(
				"marker-denom",
				"ibc-denom",
				"invalid-address",
				"owner",
				StatusActive,
				MarkerType_RestrictedCoin,
				[]AccessGrant{
					{
						"Address",
						[]Access{
							Access_Mint,
						},
					},
				},
				false,
			),
			"access list contains mint and/or burn",
		},
		{
			"should pass to validate basic",
			*NewMsgIcaReflectMarkerRequest(
				"marker-denom",
				"ibc-denom",
				validAddress,
				"owner",
				StatusActive,
				MarkerType_RestrictedCoin,
				[]AccessGrant{
					{
						"Address",
						[]Access{
							Access_Unknown,
							Access_Deposit,
							Access_Withdraw,
							Access_Delete,
							Access_Admin,
							Access_Transfer,
						},
					},
				},
				false,
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
