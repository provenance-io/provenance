package types

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func testAddress() sdk.AccAddress {
	addr := secp256k1.GenPrivKey().PubKey().Address()
	return sdk.AccAddress(addr)
}

func TestHasPermission(t *testing.T) {
	emptyGrant := NewAccessGrant(testAddress(), AccessList{})
	has := emptyGrant.HasAccess(Access_Mint)
	require.False(t, has)

	cases := []struct {
		permission Access
		expectHas  bool
	}{
		{Access_Mint, true},
		{Access_Burn, true},
		{Access_Deposit, true},
		{Access_Unknown, false},
	}
	roleAddr := testAddress()
	roleGrant := NewAccessGrant(nil, AccessList{Access_Mint, Access_Burn, Access_Deposit})

	require.Error(t, roleGrant.Validate(), "an empty address is not valid")
	require.False(t, roleGrant.HasAccess(Access_Burn), "empty address can not have permissions")
	roleGrant.Address = roleAddr.String()
	require.Equal(t, roleGrant.GetAddress(), roleAddr)
	require.Equal(t, roleGrant.String(), fmt.Sprintf("AccessGrant: %s [mint, burn, deposit]", roleAddr.String()))

	require.NoError(t, roleGrant.Validate(), "should be valid after address is assigned.")
	require.True(t, roleGrant.HasAccess(Access_Burn), "permission should be given with a correct address")

	// Check for grants in collection against unexpected address (should be no permissions)
	require.False(t, GrantsForAddress(MustGetMarkerAddress("foo"), *roleGrant).HasAccess(Access_Mint))

	// Check for grants in collection against correct address, should match our expected list above
	require.True(t, GrantsForAddress(roleAddr, *roleGrant).HasAccess(Access_Mint))
	require.False(t, GrantsForAddress(roleAddr, *roleGrant).HasAccess(Access_Admin))

	grants := roleGrant.GetAccessList()
	require.Equal(t, len(grants), 3)
	require.True(t, roleGrant.Address == roleAddr.String())

	for i, tc := range cases {
		has = roleGrant.HasAccess(tc.permission)
		require.Equal(t, tc.expectHas, has, "test case #%d", i)
	}
}

func TestAccessByName(t *testing.T) {
	tests := []struct {
		name string
		exp  Access
	}{
		{name: "unspecified", exp: Access_Unknown},
		{name: "mint", exp: Access_Mint},
		{name: "burn", exp: Access_Burn},
		{name: "deposit", exp: Access_Deposit},
		{name: "withdraw", exp: Access_Withdraw},
		{name: "delete", exp: Access_Delete},
		{name: "admin", exp: Access_Admin},
		{name: "transfer", exp: Access_Transfer},
		{name: "force_transfer", exp: Access_ForceTransfer},
		{name: "forcetransfer", exp: Access_ForceTransfer},
		{name: "bur", exp: Access_Unknown},
	}

	capFirst := func(str string) string {
		return strings.ToUpper(str[0:1]) + strings.ToLower(str[1:])
	}
	capEveryOther := func(str string) string {
		var rv strings.Builder
		for i, c := range []byte(str) {
			if i%2 == 0 {
				rv.Write(bytes.ToLower([]byte{c}))
			} else {
				rv.Write(bytes.ToUpper([]byte{c}))
			}
		}
		return rv.String()
	}

	for _, tc := range tests {
		opts := []string{
			strings.ToLower(tc.name),
			strings.ToUpper(tc.name),
			capFirst(tc.name),
			capEveryOther(tc.name),
			"access_" + strings.ToLower(tc.name),
			"ACCESS_" + strings.ToUpper(tc.name),
			"Access_" + capFirst(tc.name),
			capEveryOther("access_" + tc.name),
		}
		var names []string
		for _, opt := range opts {
			names = append(names, opt, " "+opt, opt+" ")
		}

		for _, name := range names {
			t.Run(name, func(t *testing.T) {
				act := AccessByName(name)
				assert.Equal(t, tc.exp.String(), act.String(), "AccessByName(%q)", name)
			})
		}
	}
}

func TestAccessListByNames(t *testing.T) {
	cases := []struct {
		name        string
		accessNames string
		permissions AccessList
		expectEqual bool
		expectValid bool
	}{
		{"Single value", "mint", AccessList{Access_Mint}, true, true},
		{"Single unknown value", "foo", AccessList{Access_Unknown}, true, false},
		{"Single explicit value", "ACCESS_MINT", AccessList{Access_Mint}, true, true},

		{"Multiple values", "mint,burn", AccessList{Access_Mint, Access_Burn}, true, true},
		{"Multiple values spaced", " mint, burn ", AccessList{Access_Mint, Access_Burn}, true, true},
		{"Multiple unknown values", "foo,bar,baz", AccessList{Access_Unknown, Access_Unknown, Access_Unknown}, true, false},
	}
	for i, tc := range cases {
		i, tc := i, tc
		t.Run(tc.name, func(t *testing.T) {
			result := AccessListByNames(tc.accessNames)
			if tc.expectEqual {
				require.Equal(t, tc.permissions, result, "test case #%d: %s", i, tc.name)
			} else {
				require.NotEqual(t, tc.permissions, result, "test case #%d: %s", i, tc.name)
			}
			if tc.expectValid {
				require.NoError(t, validateAccess(result), "test case #%d: %s", i, tc.name)
			} else {
				require.Error(t, validateAccess(result), "test case #%d: %s", i, tc.name)
			}
		})
	}
}

func TestAccessOneOf(t *testing.T) {
	cases := []struct {
		name        string
		permission  Access
		permissions AccessList
		expectPass  bool
	}{
		{"no permissions", Access_Burn, AccessList{}, false},
		{"valid permission single", Access_Mint, AccessList{Access_Mint}, true},
		{"invalid permission single", Access_Mint, AccessList{Access_Burn}, false},
		{"valid permission many", Access_Mint, AccessList{Access_Mint, Access_Deposit, Access_Admin}, true},
		{"invalid permission many", Access_Unknown, AccessList{Access_Mint, Access_Deposit}, false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectPass != tc.permission.IsOneOf(tc.permissions...) {
				require.Fail(t, "failed %s", tc.name)
			}
		})
	}
}

func TestValidatePermissions(t *testing.T) {
	cases := []struct {
		name        string
		permissions AccessList
		expectPass  bool
	}{
		{"no permissions", AccessList{}, true},
		{"valid permission", AccessList{Access_Mint}, true},
		{"invalid and valid permission", AccessList{Access_Deposit, Access_Unknown}, false},
	}

	for i, tc := range cases {
		i, tc := i, tc
		t.Run(tc.name, func(t *testing.T) {
			err := validateAccess(tc.permissions)
			if tc.expectPass {
				require.NoError(t, err, "test case #%d: %s", i, tc.name)
			} else {
				require.Error(t, err, "test case #%d: %s", i, tc.name)
			}
		})
	}
}

func TestAddRemovePermissions(t *testing.T) {
	roleGrant := NewAccessGrant(MustGetMarkerAddress("test"), AccessList{Access_Mint, Access_Burn, Access_Deposit})

	require.True(t, roleGrant.HasAccess(Access_Mint))
	require.NoError(t, roleGrant.RemoveAccess(Access_Mint))
	require.Error(t, roleGrant.RemoveAccess(Access_Mint), "permission should be removed already")
	require.False(t, roleGrant.HasAccess(Access_Mint))
	require.NoError(t, roleGrant.AddAccess(Access_Mint))
	require.Error(t, roleGrant.AddAccess(Access_Mint), "permission exists already")
	require.True(t, roleGrant.HasAccess(Access_Mint))

	require.Error(t, roleGrant.AddAccess(Access_Unknown))
	require.Error(t, roleGrant.RemoveAccess(Access_Unknown))
}

func TestMergeAddRemovePermissions(t *testing.T) {
	roleAddr := MustGetMarkerAddress("test")
	otherAddr := MustGetMarkerAddress("other")
	roleGrant := NewAccessGrant(roleAddr, AccessList{Access_Mint, Access_Burn, Access_Deposit})

	require.True(t, roleGrant.HasAccess(Access_Mint))
	require.NoError(t, roleGrant.MergeAdd(*NewAccessGrant(roleAddr, AccessList{Access_Mint, Access_Admin})))
	require.True(t, roleGrant.HasAccess(Access_Mint))
	require.True(t, roleGrant.HasAccess(Access_Admin))

	require.NoError(t, roleGrant.MergeRemove(*NewAccessGrant(roleAddr, AccessList{Access_Admin})))
	require.True(t, roleGrant.HasAccess(Access_Mint))
	require.False(t, roleGrant.HasAccess(Access_Admin))

	// Expect faults merging in grants for other addresses
	require.Error(t, roleGrant.MergeAdd(*NewAccessGrant(otherAddr, AccessList{Access_Mint, Access_Admin})))
	require.Error(t, roleGrant.MergeRemove(*NewAccessGrant(otherAddr, AccessList{Access_Mint, Access_Admin})))
}
