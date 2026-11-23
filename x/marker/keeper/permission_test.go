package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	simapp "github.com/provenance-io/provenance/app"
	markerkeeper "github.com/provenance-io/provenance/x/marker/keeper"
	"github.com/provenance-io/provenance/x/marker/types"
)

func TestSetAccess(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)
	k := app.MarkerKeeper
	markerAddr := types.MustGetMarkerAddress("setaccess")
	addr := testUserAddress("addr")

	tests := []struct {
		name  string
		perms types.AccessList
		exp   types.AccessList
	}{
		{
			name:  "one permission",
			perms: types.AccessList{types.Access_Mint},
			exp:   types.AccessList{types.Access_Mint},
		},
		{
			name:  "unsorted with duplicates: stored sorted and deduped",
			perms: types.AccessList{types.Access_Withdraw, types.Access_Mint, types.Access_Burn, types.Access_Mint},
			exp:   types.AccessList{types.Access_Mint, types.Access_Burn, types.Access_Withdraw},
		},
		{
			name:  "replaces what was there",
			perms: types.AccessList{types.Access_Deposit},
			exp:   types.AccessList{types.Access_Deposit},
		},
		{
			name:  "empty: deletes the entry",
			perms: types.AccessList{},
			exp:   nil,
		},
		{
			name:  "nil when there's no entry: no error",
			perms: nil,
			exp:   nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := k.SetAccess(ctx, markerAddr, addr, tc.perms)
			require.NoError(t, err, "SetAccess(%v)", tc.perms)
			act := k.GetAccess(ctx, markerAddr, addr)
			assert.Equal(t, tc.exp, act, "GetAccess after SetAccess(%v)", tc.perms)
		})
	}

	t.Run("input slice is not changed", func(t *testing.T) {
		perms := types.AccessList{types.Access_Withdraw, types.Access_Mint, types.Access_Mint}
		orig := append(types.AccessList{}, perms...)
		require.NoError(t, k.SetAccess(ctx, markerAddr, addr, perms), "SetAccess")
		assert.Equal(t, orig, perms, "perms after SetAccess")
	})
}

func TestHasAccessAndRevokeAccessEntry(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)
	k := app.MarkerKeeper
	markerAddr := types.MustGetMarkerAddress("hasaccess")
	addr := testUserAddress("addr")
	other := testUserAddress("other")

	require.NoError(t, k.SetAccess(ctx, markerAddr, addr, types.AccessList{types.Access_Mint, types.Access_Burn}), "SetAccess addr")
	require.NoError(t, k.SetAccess(ctx, markerAddr, other, types.AccessList{types.Access_Mint}), "SetAccess other")

	assert.True(t, k.HasAccess(ctx, markerAddr, addr, types.Access_Mint), "addr has mint")
	assert.True(t, k.HasAccess(ctx, markerAddr, addr, types.Access_Burn), "addr has burn")
	assert.False(t, k.HasAccess(ctx, markerAddr, addr, types.Access_Admin), "addr has admin")
	assert.False(t, k.HasAccess(ctx, markerAddr, testUserAddress("nobody"), types.Access_Mint), "nobody has mint")

	require.NoError(t, k.RevokeAccessEntry(ctx, markerAddr, addr), "RevokeAccessEntry addr")
	assert.Empty(t, k.GetAccess(ctx, markerAddr, addr), "addr access after revoke")
	assert.Equal(t, types.AccessList{types.Access_Mint}, k.GetAccess(ctx, markerAddr, other), "other access after revoking addr")
	require.NoError(t, k.RevokeAccessEntry(ctx, markerAddr, addr), "RevokeAccessEntry addr again")
}

// TestPermsAreScopedToMarker makes sure that the functions that walk a marker's permissions
// only ever see (or change) that marker's entries.
func TestPermsAreScopedToMarker(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)
	k := app.MarkerKeeper
	marker1 := types.MustGetMarkerAddress("scoped1")
	marker2 := types.MustGetMarkerAddress("scoped2")
	addr1 := testUserAddress("addr1")
	addr2 := testUserAddress("addr2")
	addr3 := testUserAddress("addr3")

	require.NoError(t, k.SetAccess(ctx, marker1, addr1, types.AccessList{types.Access_Mint, types.Access_Admin}), "SetAccess marker1 addr1")
	require.NoError(t, k.SetAccess(ctx, marker1, addr2, types.AccessList{types.Access_Burn}), "SetAccess marker1 addr2")
	require.NoError(t, k.SetAccess(ctx, marker2, addr2, types.AccessList{types.Access_Mint}), "SetAccess marker2 addr2")
	require.NoError(t, k.SetAccess(ctx, marker2, addr3, types.AccessList{types.Access_Mint}), "SetAccess marker2 addr3")

	t.Run("GetMarkerAccessList", func(t *testing.T) {
		grants, err := k.GetMarkerAccessList(ctx, marker1)
		require.NoError(t, err, "GetMarkerAccessList(marker1)")
		assert.ElementsMatch(t, []types.AccessGrant{
			{Address: addr1.String(), Permissions: types.AccessList{types.Access_Mint, types.Access_Admin}},
			{Address: addr2.String(), Permissions: types.AccessList{types.Access_Burn}},
		}, grants, "GetMarkerAccessList(marker1)")

		grants, err = k.GetMarkerAccessList(ctx, types.MustGetMarkerAddress("nosuchmarker"))
		require.NoError(t, err, "GetMarkerAccessList(nosuchmarker)")
		assert.Empty(t, grants, "GetMarkerAccessList(nosuchmarker)")
	})

	t.Run("AddressesWithAccess", func(t *testing.T) {
		addrs, err := k.AddressesWithAccess(ctx, marker2, types.Access_Mint)
		require.NoError(t, err, "AddressesWithAccess(marker2, mint)")
		assert.ElementsMatch(t, []sdk.AccAddress{addr2, addr3}, addrs, "AddressesWithAccess(marker2, mint)")

		addrs, err = k.AddressesWithAccess(ctx, marker1, types.Access_Mint)
		require.NoError(t, err, "AddressesWithAccess(marker1, mint)")
		assert.Equal(t, []sdk.AccAddress{addr1}, addrs, "AddressesWithAccess(marker1, mint)")

		addrs, err = k.AddressesWithAccess(ctx, marker1, types.Access_Withdraw)
		require.NoError(t, err, "AddressesWithAccess(marker1, withdraw)")
		assert.Empty(t, addrs, "AddressesWithAccess(marker1, withdraw)")
	})

	t.Run("RemoveAllAccessForMarker", func(t *testing.T) {
		require.NoError(t, k.RemoveAllAccessForMarker(ctx, marker1), "RemoveAllAccessForMarker(marker1)")
		grants, err := k.GetMarkerAccessList(ctx, marker1)
		require.NoError(t, err, "GetMarkerAccessList(marker1) after remove")
		assert.Empty(t, grants, "marker1 grants after remove")

		grants, err = k.GetMarkerAccessList(ctx, marker2)
		require.NoError(t, err, "GetMarkerAccessList(marker2) after removing marker1")
		assert.Len(t, grants, 2, "marker2 grants after removing marker1")
	})
}

func TestGetMarkerWithPerms(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)
	k := app.MarkerKeeper
	manager := testUserAddress("manager")
	user := testUserAddress("user")

	mac := types.NewEmptyMarkerAccount("withperms", manager.String(), []types.AccessGrant{
		*types.NewAccessGrant(manager, types.AccessList{types.Access_Admin, types.Access_Mint}),
		*types.NewAccessGrant(user, types.AccessList{types.Access_Withdraw}),
	})
	require.NoError(t, k.AddMarkerAccount(ctx, mac), "AddMarkerAccount")

	plain, err := k.GetMarker(ctx, mac.GetAddress())
	require.NoError(t, err, "GetMarker")
	assert.Empty(t, plain.GetAccessList(), "GetMarker access list")

	expGrants := []types.AccessGrant{
		{Address: manager.String(), Permissions: types.AccessList{types.Access_Mint, types.Access_Admin}},
		{Address: user.String(), Permissions: types.AccessList{types.Access_Withdraw}},
	}

	byAddr, err := k.GetMarkerWithPerms(ctx, mac.GetAddress())
	require.NoError(t, err, "GetMarkerWithPerms")
	assert.ElementsMatch(t, expGrants, byAddr.GetAccessList(), "GetMarkerWithPerms access list")

	byDenom, err := k.GetMarkerByDenomWithPerms(ctx, "withperms")
	require.NoError(t, err, "GetMarkerByDenomWithPerms")
	assert.ElementsMatch(t, expGrants, byDenom.GetAccessList(), "GetMarkerByDenomWithPerms access list")

	_, err = k.GetMarkerByDenomWithPerms(ctx, "nosuchmarker")
	assert.Error(t, err, "GetMarkerByDenomWithPerms(nosuchmarker)")
}

func TestValidateHasAccess(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)
	k := app.MarkerKeeper
	manager := testUserAddress("manager")
	user1 := testUserAddress("user1")
	user2 := testUserAddress("user2")
	user3 := testUserAddress("user3")

	mac := types.NewEmptyMarkerAccount("validatehas", manager.String(), []types.AccessGrant{
		*types.NewAccessGrant(user2, types.AccessList{types.Access_Mint}),
	})
	require.NoError(t, k.AddMarkerAccount(ctx, mac), "AddMarkerAccount")
	markerAddr := mac.GetAddress()

	t.Run("ValidateHasAccess", func(t *testing.T) {
		assert.NoError(t, k.ValidateHasAccess(ctx, markerAddr, user2, types.Access_Mint), "user2 mint")
		err := k.ValidateHasAccess(ctx, markerAddr, user1, types.Access_Mint)
		assert.EqualError(t, err, fmt.Sprintf("%s does not have ACCESS_MINT on validatehas marker (%s)", user1, markerAddr), "user1 mint")

		// When the marker doesn't exist, the denom is left blank.
		unknown := types.MustGetMarkerAddress("unknown")
		err = k.ValidateHasAccess(ctx, unknown, user1, types.Access_Mint)
		assert.EqualError(t, err, fmt.Sprintf("%s does not have ACCESS_MINT on  marker (%s)", user1, unknown), "unknown marker")
	})

	tests := []struct {
		name   string
		addrs  []sdk.AccAddress
		exp    bool
		expErr string
	}{
		{
			name:   "nil addrs",
			addrs:  nil,
			exp:    false,
			expErr: fmt.Sprintf("none of [] have permission ACCESS_MINT on validatehas marker (%s)", markerAddr),
		},
		{
			name:  "one addr: has it",
			addrs: []sdk.AccAddress{user2},
			exp:   true,
		},
		{
			name:   "one addr: does not have it",
			addrs:  []sdk.AccAddress{user1},
			exp:    false,
			expErr: fmt.Sprintf("%s does not have ACCESS_MINT on validatehas marker (%s)", user1, markerAddr),
		},
		{
			name:  "three addrs: last has it",
			addrs: []sdk.AccAddress{user1, user3, user2},
			exp:   true,
		},
		{
			name:   "two addrs: neither has it",
			addrs:  []sdk.AccAddress{user1, user3},
			exp:    false,
			expErr: fmt.Sprintf("none of [%q %q] have permission ACCESS_MINT on validatehas marker (%s)", user1.String(), user3.String(), markerAddr),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			act := k.AtLeastOneHasAccess(ctx, markerAddr, tc.addrs, types.Access_Mint)
			assert.Equal(t, tc.exp, act, "AtLeastOneHasAccess")
			err := k.ValidateAtLeastOneHasAccess(ctx, markerAddr, tc.addrs, types.Access_Mint)
			if len(tc.expErr) > 0 {
				assert.EqualError(t, err, tc.expErr, "ValidateAtLeastOneHasAccess")
			} else {
				assert.NoError(t, err, "ValidateAtLeastOneHasAccess")
			}
		})
	}
}

func TestMergeAccess(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)
	k := app.MarkerKeeper
	markerAddr := types.MustGetMarkerAddress("mergeaccess")
	existing := testUserAddress("existing")
	newAddr := testUserAddress("new")
	require.NoError(t, k.SetAccess(ctx, markerAddr, existing, types.AccessList{types.Access_Mint, types.Access_Burn}), "SetAccess")

	tests := []struct {
		name   string
		grant  *types.AccessGrant
		exp    types.AccessList
		expErr string
	}{
		{
			name:  "address with no permissions: grant unchanged",
			grant: types.NewAccessGrant(newAddr, types.AccessList{types.Access_Deposit}),
			exp:   types.AccessList{types.Access_Deposit},
		},
		{
			name:  "address with permissions: existing ones are added",
			grant: types.NewAccessGrant(existing, types.AccessList{types.Access_Withdraw}),
			exp:   types.AccessList{types.Access_Withdraw, types.Access_Mint, types.Access_Burn},
		},
		{
			name:  "overlapping permissions: no duplicates",
			grant: types.NewAccessGrant(existing, types.AccessList{types.Access_Mint}),
			exp:   types.AccessList{types.Access_Mint, types.Access_Burn},
		},
		{
			name:   "invalid grant",
			grant:  types.NewAccessGrant(newAddr, types.AccessList{types.Access_Mint, types.Access_Mint}),
			expErr: "access list contains duplicate entry",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := k.MergeAccess(ctx, markerAddr, tc.grant)
			if len(tc.expErr) > 0 {
				assert.ErrorContains(t, err, tc.expErr, "MergeAccess")
				return
			}
			require.NoError(t, err, "MergeAccess")
			assert.ElementsMatch(t, tc.exp, tc.grant.GetAccessList(), "grant permissions after MergeAccess")
		})
	}

	assert.Equal(t, types.AccessList{types.Access_Mint, types.Access_Burn}, k.GetAccess(ctx, markerAddr, existing),
		"stored permissions are not changed by MergeAccess")
}

func TestValidateGrantForMarker(t *testing.T) {
	user := testUserAddress("user")
	coin := types.NewEmptyMarkerAccount("grantcoin", "", nil)
	restricted := types.NewEmptyMarkerAccount("grantrestricted", "", nil)
	restricted.MarkerType = types.MarkerType_RestrictedCoin
	ibc := types.NewEmptyMarkerAccount("ibc/0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF", "", nil)
	ibc.SupplyFixed = false

	tests := []struct {
		name   string
		marker types.MarkerAccountI
		grant  *types.AccessGrant
		expErr string
	}{
		{
			name:   "coin marker: ok",
			marker: coin,
			grant:  types.NewAccessGrant(user, types.AccessList{types.Access_Mint, types.Access_Burn}),
		},
		{
			name:   "coin marker: transfer not allowed",
			marker: coin,
			grant:  types.NewAccessGrant(user, types.AccessList{types.Access_Transfer}),
			expErr: "invalid access privileges granted: ACCESS_TRANSFER is not supported for marker type MARKER_TYPE_COIN",
		},
		{
			name:   "restricted marker: transfer ok",
			marker: restricted,
			grant:  types.NewAccessGrant(user, types.AccessList{types.Access_Transfer}),
		},
		{
			name:   "ibc marker: deposit ok",
			marker: ibc,
			grant:  types.NewAccessGrant(user, types.AccessList{types.Access_Deposit}),
		},
		{
			name:   "ibc marker: mint not allowed",
			marker: ibc,
			grant:  types.NewAccessGrant(user, types.AccessList{types.Access_Mint}),
			expErr: "invalid ibc denom configuration: ACCESS_MINT is not supported for ibc marker",
		},
		{
			name:   "ibc marker: burn not allowed",
			marker: ibc,
			grant:  types.NewAccessGrant(user, types.AccessList{types.Access_Burn}),
			expErr: "invalid ibc denom configuration: ACCESS_BURN is not supported for ibc marker",
		},
		{
			name:   "permissions for the marker itself",
			marker: coin,
			grant:  types.NewAccessGrant(coin.GetAddress(), types.AccessList{types.Access_Mint}),
			expErr: "permissions cannot be granted to 'grantcoin' marker account: [ACCESS_MINT]",
		},
		{
			name:   "no permissions for the marker itself",
			marker: coin,
			grant:  types.NewAccessGrant(coin.GetAddress(), nil),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := markerkeeper.ValidateGrantForMarker(tc.marker, tc.grant)
			if len(tc.expErr) > 0 {
				assert.EqualError(t, err, tc.expErr, "ValidateGrantForMarker")
			} else {
				assert.NoError(t, err, "ValidateGrantForMarker")
			}
		})
	}
}

func TestValidateAccessControl(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)
	k := app.MarkerKeeper
	manager := testUserAddress("manager")
	admin1 := testUserAddress("admin1")
	admin2 := testUserAddress("admin2")
	minter := testUserAddress("minter")
	other := testUserAddress("other")

	const errNoAdmin = "a manager is required if there are no accounts with ACCESS_ADMIN and marker is not ACTIVE"
	const errNoMint = "cannot create a marker with zero total supply and no authorization for minting more"

	// newMarker creates a marker (not stored as an account) and stores the given permissions for it.
	denomNum := 0
	newMarker := func(t *testing.T, status types.MarkerStatus, mgr sdk.AccAddress, supply int64, grants ...types.AccessGrant) types.MarkerAccountI {
		t.Helper()
		denomNum++
		mgrStr := ""
		if len(mgr) > 0 {
			mgrStr = mgr.String()
		}
		m := types.NewEmptyMarkerAccount(fmt.Sprintf("vac%d", denomNum), mgrStr, nil)
		m.Status = status
		m.Supply = sdkmath.NewInt(supply)
		for _, g := range grants {
			require.NoError(t, k.SetAccess(ctx, m.GetAddress(), g.GetAddress(), g.Permissions), "SetAccess(%s)", g.Address)
		}
		return m
	}
	grant := func(addr sdk.AccAddress, perms ...types.Access) types.AccessGrant {
		return *types.NewAccessGrant(addr, perms)
	}

	tests := []struct {
		name    string
		status  types.MarkerStatus
		manager sdk.AccAddress
		supply  int64
		grants  []types.AccessGrant
		without []sdk.AccAddress
		expErr  string
	}{
		{
			name:    "proposed, with manager, no perms",
			status:  types.StatusProposed,
			manager: manager,
		},
		{
			name:   "proposed, no manager, no perms",
			status: types.StatusProposed,
			expErr: errNoAdmin,
		},
		{
			name:   "proposed, no manager, has admin",
			status: types.StatusProposed,
			grants: []types.AccessGrant{grant(other, types.Access_Mint), grant(admin1, types.Access_Admin)},
		},
		{
			name:    "proposed, no manager, only admin is excluded",
			status:  types.StatusProposed,
			grants:  []types.AccessGrant{grant(admin1, types.Access_Admin), grant(other, types.Access_Mint)},
			without: []sdk.AccAddress{admin1},
			expErr:  errNoAdmin,
		},
		{
			name:    "proposed, no manager, one of two admins excluded",
			status:  types.StatusProposed,
			grants:  []types.AccessGrant{grant(admin1, types.Access_Admin), grant(admin2, types.Access_Admin)},
			without: []sdk.AccAddress{admin1},
		},
		{
			name:    "proposed, no manager, both admins excluded",
			status:  types.StatusProposed,
			grants:  []types.AccessGrant{grant(admin1, types.Access_Admin), grant(admin2, types.Access_Admin)},
			without: []sdk.AccAddress{admin2, admin1},
			expErr:  errNoAdmin,
		},
		{
			name:    "finalized, zero supply, no mint",
			status:  types.StatusFinalized,
			manager: manager,
			grants:  []types.AccessGrant{grant(admin1, types.Access_Admin)},
			expErr:  errNoMint,
		},
		{
			name:    "finalized, zero supply, has mint",
			status:  types.StatusFinalized,
			manager: manager,
			grants:  []types.AccessGrant{grant(minter, types.Access_Mint)},
		},
		{
			name:    "finalized, zero supply, only minter excluded",
			status:  types.StatusFinalized,
			manager: manager,
			grants:  []types.AccessGrant{grant(minter, types.Access_Mint)},
			without: []sdk.AccAddress{minter},
			expErr:  errNoMint,
		},
		{
			name:    "finalized, with supply, no mint",
			status:  types.StatusFinalized,
			manager: manager,
			supply:  100,
		},
		{
			name:   "finalized, no manager, zero supply, has mint but no admin",
			status: types.StatusFinalized,
			grants: []types.AccessGrant{grant(minter, types.Access_Mint)},
			expErr: errNoAdmin,
		},
		{
			name:   "finalized, no manager, zero supply, has admin but no mint",
			status: types.StatusFinalized,
			grants: []types.AccessGrant{grant(admin1, types.Access_Admin)},
			expErr: errNoMint,
		},
		{
			name:   "finalized, no manager, zero supply, admin and mint on different addresses",
			status: types.StatusFinalized,
			grants: []types.AccessGrant{grant(admin1, types.Access_Admin), grant(minter, types.Access_Mint)},
		},
		{
			name:   "active, no manager, zero supply, no perms",
			status: types.StatusActive,
		},
		{
			name:   "cancelled, no manager, zero supply, no perms",
			status: types.StatusCancelled,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := newMarker(t, tc.status, tc.manager, tc.supply, tc.grants...)
			err := k.ValidateAccessControl(ctx, m, tc.without...)
			if len(tc.expErr) > 0 {
				assert.EqualError(t, err, tc.expErr, "ValidateAccessControl")
			} else {
				assert.NoError(t, err, "ValidateAccessControl")
			}
		})
	}
}
