package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/testutil"
	"github.com/stretchr/testify/require"

	"github.com/provenance-io/provenance/x/ibchooks/types"
	. "github.com/provenance-io/provenance/x/ibchooks/types"
)

func TestAllMsgsGetSigners(t *testing.T) {
	msgMakers := []testutil.MsgMaker{
		func(signer string) sdk.Msg { return &MsgEmitIBCAck{Sender: signer} },
		func(signer string) sdk.Msg { return &MsgUpdateParamsRequest{Authority: signer} },
	}

	testutil.RunGetSignersTests(t, AllRequestMsgs, msgMakers, nil)
}

func TestNewMsgUpdateParamsRequest(t *testing.T) {
	authority := sdk.AccAddress("authority").String()
	validContract := sdk.AccAddress("valid______________").String()
	invalidContract := "invalid_contract"

	tests := []struct {
		name      string
		contracts []string
		authority string
		expErr    string
	}{
		{
			name:      "valid request",
			contracts: []string{validContract, validContract},
			authority: authority,
		},
		{
			name:      "invalid contract address",
			contracts: []string{validContract, invalidContract},
			authority: authority,
			expErr:    `invalid contract address: "invalid_contract": decoding bech32 failed: invalid separator index -1`,
		},
		{
			name:      "invalid authority address",
			contracts: []string{validContract, validContract},
			authority: "invalid_authority",
			expErr:    `invalid authority address: "invalid_authority": decoding bech32 failed: invalid separator index -1`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgUpdateParamsRequest(tc.contracts, tc.authority)
			err := msg.ValidateBasic()
			if tc.expErr != "" {
				require.EqualError(t, err, tc.expErr, "MsgUpdateParamsRequest.ValidateBasic expected error message: %s, but got: %s", tc.expErr, err)
			} else {
				require.NoError(t, err, "MsgUpdateParamsRequest.ValidateBasic expected no error, but got: %s", err)
			}
		})
	}
}
