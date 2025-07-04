package hold_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/testutil"
	"github.com/provenance-io/provenance/testutil/assertions"

	. "github.com/provenance-io/provenance/x/hold"
)

func TestAllMsgsGetSigners(t *testing.T) {
	msgMakers := []testutil.MsgMaker{
		func(signer string) sdk.Msg { return &MsgUnlockVestingAccountsRequest{Authority: signer} },
	}

	testutil.RunGetSignersTests(t, AllRequestMsgs, msgMakers, nil)
}

func TestMsgUnlockVestingAccountsRequest(t *testing.T) {
	addrAuth := sdk.AccAddress("addrAuth____________").String()
	addr1 := sdk.AccAddress("addr1_______________").String()
	addr2 := sdk.AccAddress("addr2_______________").String()
	addr3 := sdk.AccAddress("addr3_______________").String()
	addr4 := sdk.AccAddress("addr4_______________").String()
	addr5 := sdk.AccAddress("addr5_______________").String()

	tests := []struct {
		name   string
		msg    MsgUnlockVestingAccountsRequest
		expErr string
	}{
		{
			name: "okay: five addrs",
			msg: MsgUnlockVestingAccountsRequest{
				Authority: addrAuth,
				Addresses: []string{addr1, addr2, addr3, addr4, addr5},
			},
			expErr: "",
		},
		{
			name:   "empty authority",
			msg:    MsgUnlockVestingAccountsRequest{Authority: "", Addresses: []string{addr1}},
			expErr: "invalid authority address \"\": empty address string is not allowed: invalid address",
		},
		{
			name:   "invalid authority",
			msg:    MsgUnlockVestingAccountsRequest{Authority: "willnotwork", Addresses: []string{addr1}},
			expErr: "invalid authority address \"willnotwork\": decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			name:   "nil addresses",
			msg:    MsgUnlockVestingAccountsRequest{Authority: addrAuth, Addresses: nil},
			expErr: "addresses list cannot be empty: invalid request",
		},
		{
			name:   "empty addresses",
			msg:    MsgUnlockVestingAccountsRequest{Authority: addrAuth, Addresses: []string{}},
			expErr: "addresses list cannot be empty: invalid request",
		},
		{
			name:   "one addr: invalid",
			msg:    MsgUnlockVestingAccountsRequest{Authority: addrAuth, Addresses: []string{"stillnogood"}},
			expErr: "invalid addresses[0] \"stillnogood\": decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			name:   "two addrs: first invalid",
			msg:    MsgUnlockVestingAccountsRequest{Authority: addrAuth, Addresses: []string{"ohohno", addr2}},
			expErr: "invalid addresses[0] \"ohohno\": decoding bech32 failed: invalid bech32 string length 6: invalid address",
		},
		{
			name:   "two addrs: second invalid",
			msg:    MsgUnlockVestingAccountsRequest{Authority: addrAuth, Addresses: []string{addr1, "saywhat"}},
			expErr: "invalid addresses[1] \"saywhat\": decoding bech32 failed: invalid bech32 string length 7: invalid address",
		},
		{
			name:   "two addrs: same",
			msg:    MsgUnlockVestingAccountsRequest{Authority: addrAuth, Addresses: []string{addr1, addr1}},
			expErr: "duplicate address \"" + addr1 + "\" at addresses[0] and [1]: invalid request",
		},
		{
			name: "five addrs: same first and last",
			msg: MsgUnlockVestingAccountsRequest{
				Authority: addrAuth,
				Addresses: []string{addr5, addr2, addr3, addr4, addr5},
			},
			expErr: "duplicate address \"" + addr5 + "\" at addresses[0] and [4]: invalid request",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.msg.ValidateBasic()
			}
			require.NotPanics(t, testFunc, "%T.ValidateBasic()", tc.msg)
			assertions.AssertErrorValue(t, err, tc.expErr, "%T.ValidateBasic() error", tc.msg)
		})
	}
}
