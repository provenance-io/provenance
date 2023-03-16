package types

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

// if len(msg.Manager) == 0 && msg.Status == StatusProposed {
// 	return fmt.Errorf("marker manager cannot be empty when creating a proposed marker")
// }
// if msg.Status != StatusFinalized && msg.Status != StatusProposed {
// 	return fmt.Errorf("marker can only be created with a Proposed or Finalized status")
// }
// testCoin := sdk.Coin{
// 	Denom:  msg.Amount.Denom,
// 	Amount: msg.Amount.Amount,
// }
// if !testCoin.IsValid() {
// 	return fmt.Errorf("invalid marker denom/total supply: %w", sdkerrors.ErrInvalidCoins)
// }

func TestMsgAddMarkerRequestValidateBasic(t *testing.T) {
	validAddress := sdk.MustAccAddressFromBech32("cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck")

	cases := []struct {
		name     string
		msg      MsgAddMarkerRequest
		errorMsg string
	}{
		{
			"should fail on attributes for non restricted coin",
			*NewMsgAddMarkerRequest(
				"hotdog",
				sdk.NewInt(100),
				validAddress,
				validAddress,
				MarkerType_Coin,
				true,
				true,
				[]string{"blah"},
			),
			"required attributes are reserved for restricted markers",
		},
		{
			"should succeed on attributes for restricted coin",
			*NewMsgAddMarkerRequest(
				"hotdog",
				sdk.NewInt(100),
				validAddress,
				validAddress,
				MarkerType_RestrictedCoin,
				true,
				true,
				[]string{"blah"},
			),
			"",
		},
		{
			"should succeed on for restricted coin",
			*NewMsgAddMarkerRequest(
				"hotdog",
				sdk.NewInt(100),
				validAddress,
				validAddress,
				MarkerType_RestrictedCoin,
				true,
				true,
				[]string{},
			),
			"",
		},
		{
			"should succeed on for non-restricted coin",
			*NewMsgAddMarkerRequest(
				"hotdog",
				sdk.NewInt(100),
				validAddress,
				validAddress,
				MarkerType_Coin,
				true,
				true,
				[]string{},
			),
			"",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if len(tc.errorMsg) > 0 {
				assert.Error(t, err)
				assert.Equal(t, tc.errorMsg, err.Error())
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
			"should fail on invalid marker",
			MsgAddFinalizeActivateMarkerRequest{
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
			"invalid marker denom/total supply: invalid coins",
		},
		{
			"should fail on invalid manager address",
			MsgAddFinalizeActivateMarkerRequest{
				Amount:                 sdk.NewInt64Coin("hotdog", 100),
				Manager:                "",
				FromAddress:            "",
				MarkerType:             MarkerType_Coin,
				SupplyFixed:            true,
				AllowGovernanceControl: true,
				AccessList:             []AccessGrant{*NewAccessGrant(validAddress, []Access{Access_Mint, Access_Admin})},
			},
			"empty address string is not allowed",
		},
		{
			"should fail on empty access list",
			*NewMsgAddFinalizeActivateMarkerRequest(
				"hotdog",
				sdk.NewInt(100),
				validAddress,
				validAddress,
				MarkerType_Coin,
				true,
				true,
				[]AccessGrant{},
				[]string{},
			),
			"since this will activate the marker, must have access list defined",
		},
		{
			"should fail on attributes for non restricted coin",
			*NewMsgAddFinalizeActivateMarkerRequest(
				"hotdog",
				sdk.NewInt(100),
				validAddress,
				validAddress,
				MarkerType_Coin,
				true,
				true,
				[]AccessGrant{*NewAccessGrant(validAddress, []Access{Access_Mint, Access_Admin})},
				[]string{"blah"},
			),
			"required attributes are reserved for restricted markers",
		},
		{
			"should succeed",
			*NewMsgAddFinalizeActivateMarkerRequest(
				"hotdog",
				sdk.NewInt(100),
				validAddress,
				validAddress,
				MarkerType_Coin,
				true,
				true,
				[]AccessGrant{*NewAccessGrant(validAddress, []Access{Access_Mint, Access_Admin})},
				[]string{},
			),
			"",
		},
		{
			"should succeed for restricted coin with required attributes",
			*NewMsgAddFinalizeActivateMarkerRequest(
				"hotdog",
				sdk.NewInt(100),
				validAddress,
				validAddress,
				MarkerType_RestrictedCoin,
				true,
				true,
				[]AccessGrant{*NewAccessGrant(validAddress, []Access{Access_Mint, Access_Admin})},
				[]string{"blah"},
			),
			"",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if len(tc.errorMsg) > 0 {
				assert.Error(t, err)
				assert.Equal(t, tc.errorMsg, err.Error())
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
		msg := NewMsgSupplyIncreaseProposalRequest(tc.amount, tc.targetAddress, tc.authority)
		err := msg.ValidateBasic()

		if tc.shouldFail {
			require.EqualError(t, err, tc.expectedError)
		} else {
			require.NoError(t, err)
		}
	}
}
