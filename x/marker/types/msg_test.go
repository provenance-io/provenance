package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
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
