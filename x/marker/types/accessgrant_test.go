package types

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
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

func TestAccessByString(t *testing.T) {
	cases := []struct {
		name        string
		accessNames string
		permissions AccessList
		expectEqual bool
		expectValid bool
	}{
		{
            name: "Single value",
            accessNames:  "mint",
            permissions:   AccessList{Access_Mint},
            expectEqual:    true,
            expectValid:     true},
		{
            name: "Single unknown value",
            accessNames: "foo",
            permissions:  AccessList{Access_Unknown},
            expectEqual:   true,
            expectValid:    false},
		{
            name: "Single explicit value",
            accessNames:  "ACCESS_MINT",
            permissions:   AccessList{Access_Mint},
            expectEqual:    true,
            expectValid:     true
            },

		{
            name: "Multiple values",
            accessNames:  "mint,burn",
            permissions:   AccessList{Access_Mint, Access_Burn},
            expectEqual:    true,
            expectValid:    true},
		{
            name: "Multiple values spaced",
            accessNames:  " mint, burn ",
            permissions:   AccessList{Access_Mint,  Access_Burn},
            expectEqual:    true,
            expectValid:     true},
		{
            name: "Multiple unknown values",
            accessNames: "foo,bar,baz",
            permissions:  AccessList{Access_Unknown, Access_Unknown, Access_Unknown},
            expectEqual:  true,
            expectValid:  false
        },
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
		{
            name: "no permissions",
            permission:  Access_Burn,
            permissions:   AccessList{},
            expectPass:    false},
		{
            name: "valid permission single",
            permission:  Access_Mint,
            permissions:   AccessList{Access_Mint},
            expectPass:    true},
		{
            name: "invalid permission single",
            permission:  Access_Mint,
            permissions:   AccessList{Access_Burn},
            expectPass:    false},
		{
            name: "valid permission many",
            permission:  Access_Mint,
            permissions:   AccessList{Access_Mint, Access_Deposit, Access_Admin},
            expectPass:    true},
		{
            name: "invalid permission many",
            permission:  Access_Unknown,
            permissions:   AccessList{Access_Mint, Access_Deposit},
            expectPass:    false},
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
		{
		name: "no permissions",
		permissions:  AccessList{},
		expectPass:   true},
		{
		name: "valid permission",
		permissions:  AccessList{Access_Mint},
		expectPass:   true},
		{
		name: "invalid and valid permission",
		permissions: AccessList{Access_Deposit,
		expectPass:  Access_Unknown}, false},
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
