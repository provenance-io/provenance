package keeper_test

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/x/feegrant"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/exchange"
	markerkeeper "github.com/provenance-io/provenance/x/marker/keeper"
	"github.com/provenance-io/provenance/x/marker/types"
	"github.com/provenance-io/provenance/x/quarantine"
)

// setNewAccount updates the account's number, then stores the account.
func setNewAccount(app *simapp.App, ctx sdk.Context, acc sdk.AccountI) sdk.AccountI {
	newAcc := app.AccountKeeper.NewAccount(ctx, acc)
	app.AccountKeeper.SetAccount(ctx, newAcc)
	return newAcc
}

// getAllMarkerHolders gets all the accounts holding a given denom, and the amount they each hold.
func getAllMarkerHolders(t *testing.T, ctx context.Context, app *simapp.App, denom string) []types.Balance {
	req := &banktypes.QueryDenomOwnersRequest{
		Denom:      denom,
		Pagination: &query.PageRequest{Limit: 10000},
	}
	resp, err := app.BankKeeper.DenomOwners(ctx, req)
	require.NoError(t, err, "BankKeeper.DenomOwners(%q)", denom)
	if len(resp.DenomOwners) == 0 {
		return nil
	}
	rv := make([]types.Balance, len(resp.DenomOwners))
	for i, owner := range resp.DenomOwners {
		rv[i].Address = owner.Address
		rv[i].Coins = sdk.Coins{owner.Balance}
	}
	return rv
}

func TestAccountMapperGetSet(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)

	addr := types.MustGetMarkerAddress("testcoin")
	user := testUserAddress("test")

	// no account before its created
	acc := app.AccountKeeper.GetAccount(ctx, addr)
	require.Nil(t, acc)

	// create account and check default values
	acc = types.NewEmptyMarkerAccount("testcoin", user.String(), nil)
	mac, ok := acc.(types.MarkerAccountI)
	require.True(t, ok)
	require.NotNil(t, mac)
	require.Equal(t, addr, mac.GetAddress())
	require.EqualValues(t, nil, mac.GetPubKey())

	// NewAccount doesn't call Set, so it's still nil
	require.Nil(t, app.AccountKeeper.GetAccount(ctx, addr))

	// set some values on the account and save it
	require.NoError(t, mac.GrantAccess(types.NewAccessGrant(user, []types.Access{types.Access_Mint, types.Access_Admin})))

	setNewAccount(app, ctx, acc)

	// check the new values
	acc = app.AccountKeeper.GetAccount(ctx, addr)
	require.NotNil(t, acc)
	mac, ok = acc.(types.MarkerAccountI)
	require.True(t, ok)
	require.True(t, mac.AddressHasAccess(user, types.Access_Admin))

	// add something to the send deny list just to verify removal
	app.MarkerKeeper.AddSendDeny(ctx, addr, addr)

	app.MarkerKeeper.RemoveMarker(ctx, mac)

	// marker should not exist in send deny list
	require.Empty(t, app.MarkerKeeper.GetSendDenyList(ctx, addr), "should not have entries in send deny list")

	// getting account after delete should be nil
	acc = app.AccountKeeper.GetAccount(ctx, addr)
	require.Nil(t, acc)

	require.Empty(t, getAllMarkerHolders(t, ctx, app, "testcoin"))

	// check for error on invaid marker denom
	_, err := app.MarkerKeeper.GetMarkerByDenom(ctx, "doesntexist")
	require.Error(t, err, "marker does not exist, should error")
}

func TestExistingAccounts(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)
	server := markerkeeper.NewMsgServerImpl(app.MarkerKeeper)

	addr := types.MustGetMarkerAddress("testcoin")
	pubkey := secp256k1.GenPrivKey().PubKey()
	user := testUserAddress("testcoin")
	manager := testUserAddress("manager")
	existingBalance := sdk.NewInt64Coin("coin", 1000)

	// prefund the marker address so an account gets created before the marker does.
	newAcc := setNewAccount(app, ctx, app.AccountKeeper.NewAccount(ctx, authtypes.NewBaseAccount(user, pubkey, 0, 0)))
	require.NoError(t, testutil.FundAccount(ctx, app.BankKeeper, addr, sdk.NewCoins(existingBalance)), "funding account")
	require.Equal(t, existingBalance, app.BankKeeper.GetBalance(ctx, addr, "coin"), "account balance must be set")

	// Creating a marker over an account with zero sequence is fine.
	_, err := server.AddMarker(ctx, types.NewMsgAddMarkerRequest("testcoin", sdkmath.NewInt(30), user, manager, types.MarkerType_Coin, true, true, false, []string{}, 0, 0))
	require.NoError(t, err, "should allow a marker over existing account that has not signed anything.")

	// existing coin balance must still be present
	require.Equal(t, existingBalance, app.BankKeeper.GetBalance(ctx, addr, "coin"), "account balances must be preserved")

	// Creating a marker over an existing marker fails.
	_, err = server.AddMarker(ctx, types.NewMsgAddMarkerRequest("testcoin", sdkmath.NewInt(30), user, manager, types.MarkerType_Coin, true, true, false, []string{}, 0, 0))
	require.Error(t, err, "fails because marker already exists")

	// replace existing test account with a new copy that has a positive sequence number
	err = newAcc.SetSequence(10)
	require.NoError(t, err, "newAcc.SetSequence(10)")
	app.AccountKeeper.SetAccount(ctx, newAcc)

	// Creating a marker over an existing account with a positive sequence number fails.
	_, err = server.AddMarker(ctx, types.NewMsgAddMarkerRequest("testcoin", sdkmath.NewInt(30), user, manager, types.MarkerType_Coin, true, true, false, []string{}, 0, 0))
	require.Error(t, err, "should not allow creation over and existing account with a positive sequence number.")
}

func TestUnrestrictedDenoms(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)
	server := markerkeeper.NewMsgServerImpl(app.MarkerKeeper)

	user := testUserAddress("test")

	// Require a long unrestricted denom
	app.MarkerKeeper.SetParams(ctx, types.Params{UnrestrictedDenomRegex: "[a-z]{12,20}"})
	_, err := server.AddMarker(ctx, types.NewMsgAddMarkerRequest("tooshort", sdkmath.NewInt(30), user, user, types.MarkerType_Coin, true, true, false, []string{}, 0, 0))
	require.Error(t, err, "fails with unrestricted denom length fault")
	require.Equal(t, fmt.Errorf("invalid denom [tooshort] (fails unrestricted marker denom validation [a-z]{12,20})"), err, "should fail with denom restriction")

	_, err = server.AddMarker(ctx, types.NewMsgAddMarkerRequest("itslongenough", sdkmath.NewInt(30), user, user, types.MarkerType_Coin, true, true, false, []string{}, 0, 0))
	require.NoError(t, err, "should allow a marker with a sufficiently long denom")

	// Set to an empty string (returns to default expression)
	app.MarkerKeeper.SetParams(ctx, types.Params{UnrestrictedDenomRegex: ""})
	_, err = server.AddMarker(ctx, types.NewMsgAddMarkerRequest("short", sdkmath.NewInt(30), user, user, types.MarkerType_Coin, true, true, false, []string{}, 0, 0))
	// succeeds now as the default unrestricted denom expression allows any valid denom (minimum length is 2)
	require.NoError(t, err, "should allow any valid denom with a min length of two")
}

func TestAccountKeeperReader(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)

	addr := types.MustGetMarkerAddress("testcoin")
	user := testUserAddress("test")
	// create account and check default values
	mac := types.NewEmptyMarkerAccount(
		"testcoin",
		user.String(),
		[]types.AccessGrant{*types.NewAccessGrant(user, []types.Access{types.Access_Mint})})

	require.NoError(t, app.MarkerKeeper.AddMarkerAccount(ctx, mac))

	m, err := app.MarkerKeeper.GetMarkerByDenom(ctx, "testcoin")
	require.NoError(t, err)
	require.NotNil(t, m)
	require.EqualValues(t, m.GetDenom(), "testcoin")
	require.EqualValues(t, m.GetAddress(), addr)

	m, err = app.MarkerKeeper.GetMarker(ctx, addr)
	require.NoError(t, err)
	require.NotNil(t, m)
	require.EqualValues(t, m.GetDenom(), "testcoin")
	require.EqualValues(t, m.GetAddress(), addr)

	count := 0
	app.MarkerKeeper.IterateMarkers(ctx, func(record types.MarkerAccountI) bool {
		require.EqualValues(t, record.GetDenom(), "testcoin")
		count++
		return false
	})
	require.EqualValues(t, count, 1)
}

func TestManageAccess(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)

	addr := types.MustGetMarkerAddress("testcoin")
	// Easiest way to create a valid bech32 address for testing.
	user1 := testUserAddress("test1")
	user2 := testUserAddress("test2")
	admin := testUserAddress("admin")

	// create account and check default values
	mac := types.NewEmptyMarkerAccount("testcoin",
		user1.String(),
		[]types.AccessGrant{*types.NewAccessGrant(user1, []types.Access{types.Access_Burn}),
			*types.NewAccessGrant(admin, []types.Access{types.Access_Admin})})

	require.NoError(t, mac.SetManager(user1))
	require.NoError(t, mac.SetSupply(sdk.NewInt64Coin(mac.Denom, 1)))
	require.NoError(t, app.MarkerKeeper.AddMarkerAccount(ctx, mac))
	require.NoError(t, app.MarkerKeeper.SetNetAssetValue(ctx, mac, types.NewNetAssetValue(sdk.NewInt64Coin(types.UsdDenom, 1), 1), "test"))

	// Initial, should not have access
	m, err := app.MarkerKeeper.GetMarkerByDenom(ctx, "testcoin")
	require.NoError(t, err)
	require.NotNil(t, m)
	require.False(t, m.AddressHasAccess(user2, types.Access_Burn))

	// Grant access and check (succeeds on a Proposed marker without Admin grant)
	require.NoError(t,
		app.MarkerKeeper.AddAccess(
			ctx, user1, "testcoin", types.NewAccessGrant(user2, []types.Access{types.Access_Mint, types.Access_Delete})),
	)

	// Grant access fails for caller that is not the manager of a proposed marker
	require.Error(t, app.MarkerKeeper.AddAccess(
		ctx, user2, "testcoin", types.NewAccessGrant(user2, []types.Access{types.Access_Burn})))
	require.Error(t, app.MarkerKeeper.RemoveAccess(ctx, user2, "testcoin", user1))

	m, err = app.MarkerKeeper.GetMarker(ctx, addr)
	require.NoError(t, err)
	require.NotNil(t, m)
	require.True(t, m.AddressHasAccess(user2, types.Access_Mint))
	require.False(t, m.AddressHasAccess(user2, types.Access_Burn))
	require.False(t, m.AddressHasAccess(user2, types.Access_Admin))
	require.True(t, m.AddressHasAccess(user2, types.Access_Delete))
	require.False(t, m.AddressHasAccess(user2, types.Access_Withdraw))

	// Remove access and check
	require.NoError(t, app.MarkerKeeper.RemoveAccess(ctx, user1, "testcoin", user2))

	m, err = app.MarkerKeeper.GetMarker(ctx, addr)
	require.NoError(t, err)
	require.NotNil(t, m)
	require.False(t, m.AddressHasAccess(user2, types.Access_Mint))
	require.False(t, m.AddressHasAccess(user2, types.Access_Burn))
	require.False(t, m.AddressHasAccess(user2, types.Access_Admin))
	require.False(t, m.AddressHasAccess(user2, types.Access_Delete))
	require.False(t, m.AddressHasAccess(user2, types.Access_Withdraw))

	// Finalize marker and check permission enforcement.
	require.NoError(t, app.MarkerKeeper.FinalizeMarker(ctx, user1, m.GetDenom()))
	_, err = app.MarkerKeeper.GetMarker(ctx, addr)
	require.NoError(t, err)

	// Manager can make changes to grants for finalized markers
	require.NoError(t, app.MarkerKeeper.RemoveAccess(ctx, user1, "testcoin", user1))
	require.NoError(t, app.MarkerKeeper.AddAccess(ctx, user1, "testcoin",
		types.NewAccessGrant(user1, []types.Access{types.Access_Burn})))

	// Unauthorized user can not manipulate finalized marker grants
	require.Error(t, app.MarkerKeeper.RemoveAccess(ctx, user2, "testcoin", user1))

	// Admin can make changes to grants for finalized markers
	require.NoError(t, app.MarkerKeeper.AddAccess(ctx, admin, "testcoin",
		types.NewAccessGrant(user2, []types.Access{types.Access_Mint, types.Access_Delete})))
	_, err = app.MarkerKeeper.GetMarker(ctx, addr)
	require.NoError(t, err)

	// User2 can adjust supply up/down for a finalized marker
	require.NoError(t, app.MarkerKeeper.MintCoin(ctx, user2, sdk.NewInt64Coin("testcoin", 1)))
	require.NoError(t, app.MarkerKeeper.BurnCoin(ctx, user1, sdk.NewInt64Coin("testcoin", 1)))

	// Cancel marker and check permission enforcement.
	require.NoError(t, app.MarkerKeeper.CancelMarker(ctx, user2, "testcoin"))

	// Admin cannot make changes to grants for cancelled markers
	require.Error(t, app.MarkerKeeper.AddAccess(ctx, admin, "testcoin",
		types.NewAccessGrant(user2, []types.Access{types.Access_Burn})))
	require.Error(t, app.MarkerKeeper.RemoveAccess(ctx, admin, "testcoin", user2))

	// Load the marker one last time and verify our permission records are consistent and correct
	m, err = app.MarkerKeeper.GetMarker(ctx, addr)
	require.NoError(t, err)

	require.True(t, m.AddressHasAccess(admin, types.Access_Admin))
	require.True(t, m.AddressHasAccess(user1, types.Access_Burn))
	require.True(t, m.AddressHasAccess(user2, types.Access_Mint))
	require.True(t, m.AddressHasAccess(user2, types.Access_Delete))

	require.EqualValues(t, 1, len(m.AddressListForPermission(types.Access_Delete)))
	require.EqualValues(t, 1, len(m.AddressListForPermission(types.Access_Burn)))
	require.EqualValues(t, 1, len(m.AddressListForPermission(types.Access_Admin)))
	require.EqualValues(t, 0, len(m.AddressListForPermission(types.Access_Deposit)))
	require.EqualValues(t, 1, len(m.AddressListForPermission(types.Access_Mint)))
	require.EqualValues(t, 0, len(m.AddressListForPermission(types.Access_Withdraw)))
}

func TestCancelProposedByManager(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)

	addr := types.MustGetMarkerAddress("testcoin")
	// Easiest way to create a valid bech32 address for testing.
	user1 := testUserAddress("test1")
	user2 := testUserAddress("test2")
	admin := testUserAddress("admin")

	// create account and check default values
	mac := types.NewEmptyMarkerAccount("testcoin",
		user1.String(),
		[]types.AccessGrant{*types.NewAccessGrant(user1, []types.Access{types.Access_Burn}),
			*types.NewAccessGrant(admin, []types.Access{types.Access_Admin})})

	require.NoError(t, mac.SetManager(user1))
	require.NoError(t, mac.SetSupply(sdk.NewInt64Coin(mac.Denom, 1)))
	require.NoError(t, app.MarkerKeeper.AddMarkerAccount(ctx, mac))

	m, err := app.MarkerKeeper.GetMarker(ctx, addr)
	require.NoError(t, err)
	// user1 and user2 will not have been assigned delete
	require.False(t, m.AddressHasAccess(user1, types.Access_Delete))
	require.False(t, m.AddressHasAccess(user2, types.Access_Delete))

	// Delete marker (fails, marker is not cancelled)
	require.Error(t, app.MarkerKeeper.DeleteMarker(ctx, user1, "testcoin"), "can only delete markeraccounts in the Cancelled status")

	// Cancel marker and check permission enforcement. (expect fail, no access)
	require.Error(t, app.MarkerKeeper.CancelMarker(ctx, user2, "testcoin"))

	// Cancel marker and check permission enforcement. (succeeds for manager)
	require.NoError(t, app.MarkerKeeper.CancelMarker(ctx, user1, "testcoin"))

	// Delete marker and check permission enforcement. (expect fail, no access)
	require.Error(t, app.MarkerKeeper.DeleteMarker(ctx, user2, "testcoin"), "does not have ACCESS_DELETE on testcoin markeraccount")

	// Delete marker and check permission enforcement. (succeeds for manager)
	require.NoError(t, app.MarkerKeeper.DeleteMarker(ctx, user1, "testcoin"))
}

func TestMintBurnCoins(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)
	app.MarkerKeeper.SetParams(ctx, types.DefaultParams())
	addr := types.MustGetMarkerAddress("testcoin")
	user := testUserAddress("test")

	// fail for an unknown coin.
	require.Error(t, app.MarkerKeeper.MintCoin(ctx, user, sdk.NewInt64Coin("testcoin", 100)))
	require.Error(t, app.MarkerKeeper.BurnCoin(ctx, user, sdk.NewInt64Coin("testcoin", 100)))

	// create account and check default values
	mac := types.NewEmptyMarkerAccount("testcoin", user.String(), []types.AccessGrant{*types.NewAccessGrant(user,
		[]types.Access{types.Access_Mint, types.Access_Burn, types.Access_Withdraw, types.Access_Delete, types.Access_Deposit})})
	require.NoError(t, mac.SetManager(user))
	require.NoError(t, mac.SetSupply(sdk.NewInt64Coin("testcoin", 1000)))

	require.NoError(t, app.MarkerKeeper.AddMarkerAccount(ctx, mac))
	require.NoError(t, app.MarkerKeeper.SetNetAssetValue(ctx, mac, types.NewNetAssetValue(sdk.NewInt64Coin(types.UsdDenom, 1), 1), "test"))

	// Should not fail for a non-active/finalized coin, must be able to adjust supply amount to match any existing
	require.NoError(t, app.MarkerKeeper.MintCoin(ctx, user, sdk.NewInt64Coin("testcoin", 100)))
	require.NoError(t, app.MarkerKeeper.BurnCoin(ctx, user, sdk.NewInt64Coin("testcoin", 100)))

	// Moves to finalized, mints required supply, moves to active status.
	require.NoError(t, app.MarkerKeeper.FinalizeMarker(ctx, user, "testcoin"))
	require.NoError(t, app.MarkerKeeper.ActivateMarker(ctx, user, "testcoin"))

	// Load the created marker
	m, err := app.MarkerKeeper.GetMarker(ctx, addr)
	require.NoError(t, err)
	require.EqualValues(t, m.GetSupply(), sdk.NewInt64Coin("testcoin", 1000))
	// entire supply should have been allocated to markeracount
	require.EqualValues(t, app.MarkerKeeper.GetEscrow(ctx, m).AmountOf("testcoin"), sdkmath.NewInt(1000))

	// perform a successful mint (and check)
	require.NoError(t, app.MarkerKeeper.MintCoin(ctx, user, sdk.NewInt64Coin("testcoin", 100)))
	m, err = app.MarkerKeeper.GetMarker(ctx, addr)
	require.NoError(t, err)
	require.EqualValues(t, m.GetSupply(), sdk.NewInt64Coin("testcoin", 1100))
	require.EqualValues(t, app.MarkerKeeper.GetEscrow(ctx, m), sdk.NewCoins(sdk.NewInt64Coin("testcoin", 1100)))

	// perform a successful burn (and check)
	require.NoError(t, app.MarkerKeeper.BurnCoin(ctx, user, sdk.NewInt64Coin("testcoin", 100)))
	m, err = app.MarkerKeeper.GetMarker(ctx, addr)
	require.NoError(t, err)
	require.EqualValues(t, m.GetSupply(), sdk.NewInt64Coin("testcoin", 1000))

	// Fail for burn too much (exceed supply)
	require.Error(t, app.MarkerKeeper.BurnCoin(ctx, user, sdk.NewInt64Coin("testcoin", 10000)))

	// check that supply remains unchanged after above error
	m, err = app.MarkerKeeper.GetMarker(ctx, addr)
	require.NoError(t, err)
	require.EqualValues(t, m.GetSupply(), sdk.NewInt64Coin("testcoin", 1000))

	// Check that our marker account is currently holding all the minted coins after above mint/burn
	require.EqualValues(t, app.MarkerKeeper.GetEscrow(ctx, m), sdk.NewCoins(sdk.NewInt64Coin("testcoin", 1000)))

	// move coin out of the marker and into a user account
	require.NoError(t, app.MarkerKeeper.WithdrawCoins(ctx, user, user, "testcoin",
		sdk.NewCoins(sdk.NewInt64Coin("testcoin", 50))))

	// verify user has the withdrawn coins
	require.EqualValues(t, app.BankKeeper.GetBalance(ctx, user, "testcoin").Amount, sdkmath.NewInt(50))

	// verify marker account has remaining coins
	m, err = app.MarkerKeeper.GetMarker(ctx, addr)
	require.NoError(t, err)
	require.EqualValues(t, app.MarkerKeeper.GetEscrow(ctx, m).AmountOf("testcoin"), sdkmath.NewInt(950))

	// Fail for burn too much (exceed marker account holdings)
	require.Error(t, app.MarkerKeeper.BurnCoin(ctx, user, sdk.NewInt64Coin("testcoin", 1000)))
	// Fails because a user is holding some of the supply
	require.Error(t, app.MarkerKeeper.CancelMarker(ctx, user, "testcoin"))

	// two a user and the marker
	require.Equal(t, 2, len(getAllMarkerHolders(t, ctx, app, "testcoin")))

	// put the coins back in the types.
	require.NoError(t, app.BankKeeper.SendCoins(ctx, user, addr, sdk.NewCoins(sdk.NewInt64Coin("testcoin", 50))))

	// one, only the marker
	require.Equal(t, 1, len(getAllMarkerHolders(t, ctx, app, "testcoin")))

	// succeeds because marker has all its supply
	require.NoError(t, app.MarkerKeeper.CancelMarker(ctx, user, "testcoin"))

	// verify status is cancelled
	m, err = app.MarkerKeeper.GetMarker(ctx, addr)
	require.NoError(t, err)
	require.EqualValues(t, types.StatusCancelled, m.GetStatus())

	// succeeds on a cancelled marker (no-op)
	require.NoError(t, app.MarkerKeeper.CancelMarker(ctx, user, "testcoin"))

	// Set an escrow balance
	require.NoError(t, testutil.FundAccount(ctx, app.BankKeeper, addr, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1))), "funding account")
	// Fails because there are coins in escrow.
	require.Error(t, app.MarkerKeeper.DeleteMarker(ctx, user, "testcoin"))

	// Remove escrow balance from account
	require.NoError(t, app.BankKeeper.SendCoinsFromAccountToModule(types.WithTransferAgents(ctx, user), addr, "mint", sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.OneInt()))), "sending coins to module")

	// Succeeds because the bond denom coin was removed.
	require.NoError(t, app.MarkerKeeper.DeleteMarker(ctx, user, "testcoin"))

	// none, marker has been deleted
	require.Equal(t, 0, len(getAllMarkerHolders(t, ctx, app, "testcoin")))

	// verify status is destroyed and supply is zero.
	m, err = app.MarkerKeeper.GetMarker(ctx, addr)
	require.NoError(t, err)
	require.EqualValues(t, types.StatusDestroyed, m.GetStatus())
	require.EqualValues(t, m.GetSupply().Amount, sdkmath.ZeroInt())

	// supply module should also indicate a zero supply
	require.EqualValues(t, app.BankKeeper.GetSupply(ctx, "testcoin").Amount, sdkmath.ZeroInt())
}

func TestWithdrawCoins(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.NewContext(false)

	addrManager := sdk.AccAddress("addrManager_________")
	addrNoWithdraw := sdk.AccAddress("addrNoWithdraw______")
	addrOnlyWithdraw := sdk.AccAddress("addrOnlyWithdraw____")
	addrWithDep := sdk.AccAddress("addrWithDep_________")
	addr1 := sdk.AccAddress("addr1_______________")
	addr2 := sdk.AccAddress("addr2_______________")
	addr3 := sdk.AccAddress("addr3_______________")

	denomMain := "mackenzie"
	denomCoin := "norman"
	denomToDeposit := "dana"
	denomInactive := "indigo"
	denomNoMarker := "noah"

	allAccessExcept := func(addr sdk.AccAddress, perms ...types.Access) types.AccessGrant {
		rv := types.AccessGrant{
			Address:     addr.String(),
			Permissions: nil,
		}
		for permVal := range types.Access_name {
			if permVal == 0 {
				continue
			}
			perm := types.Access(permVal)
			keep := true
			for _, ignore := range perms {
				if perm == ignore {
					keep = false
					break
				}
			}
			if keep {
				rv.Permissions = append(rv.Permissions, perm)
			}
		}
		sort.Slice(rv.Permissions, func(i, j int) bool {
			return rv.Permissions[i] < rv.Permissions[j]
		})
		return rv
	}
	allAccess := func(addr sdk.AccAddress) types.AccessGrant {
		return allAccessExcept(addr)
	}
	accessOnly := func(addr sdk.AccAddress, perms ...types.Access) types.AccessGrant {
		return *types.NewAccessGrant(addr, perms)
	}
	markerAddr := func(denom string) sdk.AccAddress {
		rv, err := types.MarkerAddress(denom)
		require.NoError(t, err, "MarkerAddress(%q)", denom)
		return rv
	}
	fundAcct := func(addr sdk.AccAddress, amount sdk.Coins) {
		err := testutil.FundAccount(types.WithBypass(ctx), app.BankKeeper, addr, amount)
		require.NoError(t, err, "FundAccount(%q, %q)", string(addr), amount)
	}
	setupMarker := func(denom string, marker *types.MarkerAccount) sdk.AccAddress {
		addr := markerAddr(denom)
		marker.BaseAccount = authtypes.NewBaseAccountWithAddress(addr)
		marker.Denom = denom
		if marker.Supply.IsNil() {
			marker.Supply = sdkmath.NewInt(1_000_000_000)
		}
		if marker.Status == 0 {
			marker.Status = types.StatusProposed
		}
		if marker.MarkerType == 0 {
			marker.MarkerType = types.MarkerType_RestrictedCoin
		}
		if len(marker.Manager) == 0 {
			marker.Manager = addrManager.String()
		}
		managerHasAccess := false
		for _, grant := range marker.AccessControl {
			if grant.Address == marker.Manager {
				managerHasAccess = true
				break
			}
		}
		if !managerHasAccess {
			marker.AccessControl = append([]types.AccessGrant{allAccess(addrManager)}, marker.AccessControl...)
		}

		return addr
	}
	createMarker := func(denom string, marker *types.MarkerAccount) sdk.AccAddress {
		addr := setupMarker(denom, marker)
		err := app.MarkerKeeper.SetNetAssetValue(ctx, marker, types.NewNetAssetValue(sdk.NewInt64Coin(types.UsdDenom, 1), 1), "test")
		require.NoError(t, err, "SetNetAssetValue %q", denom)
		err = app.MarkerKeeper.AddFinalizeAndActivateMarker(ctx, marker)
		require.NoError(t, err, "AddFinalizeAndActivateMarker %q", denom)
		return addr
	}
	createProposedMarker := func(denom string, marker *types.MarkerAccount) sdk.AccAddress {
		addr := setupMarker(denom, marker)
		err := app.MarkerKeeper.AddMarkerAccount(ctx, marker)
		require.NoError(t, err, "AddMarkerAccount %q", denom)
		return addr
	}
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.NewInt64Coin(denom, amount)
	}
	noAccessErr := func(addr sdk.AccAddress, role types.Access, denom string) string {
		mAddr, err := types.MarkerAddress(denom)
		require.NoError(t, err, "MarkerAddress(%q)", denom)
		return fmt.Sprintf("%s does not have %s on %s marker (%s)", addr, role, denom, mAddr)
	}

	markerMain := &types.MarkerAccount{
		AccessControl: []types.AccessGrant{
			allAccessExcept(addrNoWithdraw, types.Access_Withdraw),
			accessOnly(addrOnlyWithdraw, types.Access_Withdraw),
			accessOnly(addrWithDep, types.Access_Withdraw),
		},
		SupplyFixed:            true,
		AllowGovernanceControl: true,
	}
	addrMarkerMain := createMarker(denomMain, markerMain)
	fundAcct(addrMarkerMain, sdk.NewCoins(coin(1_000_000, denomCoin), coin(1_000_000, denomNoMarker)))

	markerCoin := &types.MarkerAccount{
		AccessControl:          []types.AccessGrant{allAccessExcept(addrManager, types.Access_Transfer, types.Access_ForceTransfer)},
		MarkerType:             types.MarkerType_Coin,
		SupplyFixed:            true,
		AllowGovernanceControl: true,
	}
	addrMarkerCoin := createMarker(denomCoin, markerCoin)

	markerToDeposit := &types.MarkerAccount{
		AccessControl: []types.AccessGrant{
			accessOnly(addrOnlyWithdraw, types.Access_Withdraw),
			accessOnly(addrWithDep, types.Access_Deposit),
		},
		SupplyFixed:            true,
		AllowGovernanceControl: true,
	}
	addrMarkerToDeposit := createMarker(denomToDeposit, markerToDeposit)

	markerInactive := &types.MarkerAccount{
		SupplyFixed:            true,
		AllowGovernanceControl: true,
	}
	createProposedMarker(denomInactive, markerInactive)

	// Marker needs:
	// a non-restricted marker
	// a restricted marker with funds to withdraw
	// 		Need admin without withdraw
	//      Need admin with withdraw but not deposit on another.
	//      Need admin with withdraw and deposit on another.
	// a restricted marker to send funds to.
	//      Need admin with deposit and also withdraw on main.
	//      Need admin with withdraw on both, but not deposit on this one.

	tests := []struct {
		name                  string
		bankKeeper            *WrappedBankKeeper
		caller                sdk.AccAddress
		recipient             sdk.AccAddress
		denom                 string
		coins                 sdk.Coins
		expErr                string
		expEventTo            sdk.AccAddress
		expCallerGetsCoins    bool
		expRecipientGetsCoins bool
	}{
		{
			name:      "no marker",
			caller:    addrManager,
			recipient: addr1,
			denom:     denomNoMarker,
			coins:     sdk.NewCoins(coin(5, denomNoMarker)),
			expErr:    "marker not found for " + denomNoMarker + ": marker " + denomNoMarker + " not found for address: " + markerAddr(denomNoMarker).String(),
		},
		{
			name:      "no withdraw access",
			caller:    addrNoWithdraw,
			recipient: addr2,
			denom:     denomMain,
			coins:     sdk.NewCoins(coin(5, denomMain)),
			expErr:    noAccessErr(addrNoWithdraw, types.Access_Withdraw, denomMain),
		},
		{
			name:                  "to a coin marker",
			caller:                addrManager,
			recipient:             addrMarkerCoin,
			denom:                 denomMain,
			coins:                 sdk.NewCoins(coin(3, denomMain)),
			expEventTo:            addrMarkerCoin,
			expRecipientGetsCoins: true,
		},
		{
			name:      "to a restricted marker: admin does not have deposit on it",
			caller:    addrOnlyWithdraw,
			recipient: addrMarkerToDeposit,
			denom:     denomMain,
			coins:     sdk.NewCoins(coin(6, denomMain)),
			expErr:    noAccessErr(addrOnlyWithdraw, types.Access_Deposit, denomToDeposit),
		},
		{
			name:                  "to a restricted marker: admin has deposit on it",
			caller:                addrWithDep,
			recipient:             addrMarkerToDeposit,
			denom:                 denomMain,
			coins:                 sdk.NewCoins(coin(7, denomMain)),
			expEventTo:            addrMarkerToDeposit,
			expRecipientGetsCoins: true,
		},
		{
			name:      "marker is not active",
			caller:    addrManager,
			recipient: addr3,
			denom:     denomInactive,
			coins:     sdk.NewCoins(coin(4, denomInactive)),
			expErr:    "cannot withdraw marker created coins from a marker that is not in Active status",
		},
		{
			name:       "to addr blocked by bank module",
			bankKeeper: NewWrappedBankKeeper().WithExtraBlockedAddrs(addr2),
			caller:     addrManager,
			recipient:  addr2,
			denom:      denomMain,
			coins:      sdk.NewCoins(coin(14, denomMain)),
			expErr:     addr2.String() + " is not allowed to receive funds",
		},
		{
			name:       "error from send",
			bankKeeper: NewWrappedBankKeeper().WithSendCoinsErrs("some random error"),
			caller:     addrOnlyWithdraw,
			recipient:  addr1,
			denom:      denomMain,
			coins:      sdk.NewCoins(coin(12, denomMain)),
			expErr:     "some random error",
		},
		{
			name:               "no recipient provided",
			caller:             addrOnlyWithdraw,
			recipient:          nil,
			denom:              denomMain,
			coins:              sdk.NewCoins(coin(12, denomMain), coin(3, denomCoin), coin(77, denomNoMarker)),
			expEventTo:         addrOnlyWithdraw,
			expCallerGetsCoins: true,
		},
		{
			name:                  "recipient is caller",
			caller:                addrOnlyWithdraw,
			recipient:             addrOnlyWithdraw,
			denom:                 denomMain,
			coins:                 sdk.NewCoins(coin(33, denomMain)),
			expEventTo:            addrOnlyWithdraw,
			expCallerGetsCoins:    true,
			expRecipientGetsCoins: true,
		},
		{
			name:                  "recipient is not caller",
			caller:                addrOnlyWithdraw,
			recipient:             addr3,
			denom:                 denomMain,
			coins:                 sdk.NewCoins(coin(27, denomNoMarker)),
			expEventTo:            addr3,
			expRecipientGetsCoins: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var expEvents sdk.Events
			if len(tc.expEventTo) > 0 {
				tev := &types.EventMarkerWithdraw{
					Coins:         tc.coins.String(),
					Denom:         tc.denom,
					Administrator: tc.caller.String(),
					ToAddress:     tc.expEventTo.String(),
				}
				event, err := sdk.TypedEventToEvent(tev)
				require.NoError(t, err, "TypedEventToEvent(%#v)", tev)
				expEvents = append(expEvents, event)
			}

			ctx = app.NewContext(false)
			var callerOrigBals, recipientOrigBals sdk.Coins
			var callerExpBals, recipientExpBals sdk.Coins
			if len(tc.caller) > 0 {
				callerOrigBals = app.BankKeeper.GetAllBalances(ctx, tc.caller)
				callerExpBals = callerOrigBals
				if tc.expCallerGetsCoins {
					callerExpBals = callerExpBals.Add(tc.coins...)
				}
			}
			if len(tc.recipient) > 0 {
				recipientOrigBals = app.BankKeeper.GetAllBalances(ctx, tc.recipient)
				recipientExpBals = recipientOrigBals
				if tc.expRecipientGetsCoins {
					recipientExpBals = recipientExpBals.Add(tc.coins...)
				}
			}

			kpr := app.MarkerKeeper
			if tc.bankKeeper != nil {
				kpr = kpr.WithBankKeeper(tc.bankKeeper.WithParent(app.BankKeeper))
			}

			em := sdk.NewEventManager()
			ctx = ctx.WithEventManager(em)
			var err error
			testFunc := func() {
				err = kpr.WithdrawCoins(ctx, tc.caller, tc.recipient, tc.denom, tc.coins)
			}
			require.NotPanics(t, testFunc, "WithdrawCoins(%q, %q, %q, %q)",
				string(tc.caller), string(tc.recipient), tc.denom, tc.coins)
			assertions.AssertErrorValue(t, err, tc.expErr, "WithdrawCoins(%q, %q, %q, %q) error",
				string(tc.caller), string(tc.recipient), tc.denom, tc.coins)
			actEvents := em.Events()
			assertions.AssertEventsContains(t, expEvents, actEvents, "WithdrawCoins(%q, %q, %q, %q) events",
				string(tc.caller), string(tc.recipient), tc.denom, tc.coins)

			var callerActBals, recipientActBals sdk.Coins
			if len(tc.caller) > 0 {
				callerActBals = app.BankKeeper.GetAllBalances(ctx, tc.caller)
			}
			if len(tc.recipient) > 0 {
				recipientActBals = app.BankKeeper.GetAllBalances(ctx, tc.recipient)
			}
			assert.Equal(t, callerExpBals.String(), callerActBals.String(), "caller balances after WithdrawCoins(%q, %q, %q, %q)",
				string(tc.caller), string(tc.recipient), tc.denom, tc.coins)
			assert.Equal(t, recipientExpBals.String(), recipientActBals.String(), "recipient balances after WithdrawCoins(%q, %q, %q, %q)",
				string(tc.caller), string(tc.recipient), tc.denom, tc.coins)
		})
	}
}

func TestMarkerGetters(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)

	user := testUserAddress("test")
	mac := types.NewEmptyMarkerAccount("testcoin",
		user.String(),
		[]types.AccessGrant{*types.NewAccessGrant(user, []types.Access{types.Access_Deposit})})
	require.NoError(t, mac.SetManager(user))
	require.NoError(t, app.MarkerKeeper.AddMarkerAccount(ctx, mac))

	mac = types.NewEmptyMarkerAccount("secondcoin",
		user.String(),
		[]types.AccessGrant{*types.NewAccessGrant(user, []types.Access{types.Access_Deposit})})
	require.NoError(t, mac.SetManager(user))
	require.NoError(t, app.MarkerKeeper.AddMarkerAccount(ctx, mac))

	var err error
	var m types.MarkerAccountI
	m, err = app.MarkerKeeper.GetMarkerByDenom(ctx, "testcoin")
	require.NoError(t, err)
	require.NotNil(t, m)

	m, err = app.MarkerKeeper.GetMarkerByDenom(ctx, "secondcoin")
	require.NoError(t, err)
	require.NotNil(t, m)

	count := 0
	app.MarkerKeeper.IterateMarkers(ctx, func(record types.MarkerAccountI) bool {
		count++
		return false
	})
	require.Equal(t, 2, count)

	// Could do more in-depth checking, but if both markers are returned that is the expected behavior
}

func TestInsufficientExisting(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)

	pubkey := secp256k1.GenPrivKey().PubKey()
	user := sdk.AccAddress(pubkey.Address())

	// setup an existing account with an existing balance (and matching supply)
	existingSupply := sdk.NewInt64Coin("testcoin", 10000)
	setNewAccount(app, ctx, authtypes.NewBaseAccount(user, pubkey, 0, 0))

	require.NoError(t, testutil.FundAccount(ctx, app.BankKeeper, user, sdk.NewCoins(existingSupply)), "funding account")

	//prevSupply := app.BankKeeper.GetSupply(ctx, "testcoin")
	//app.BankKeeper.SetSupply(ctx, banktypes.NewSupply(prevSupply.Amount.Add(existingSupply.Amount)))

	// create account and check default values
	mac := types.NewEmptyMarkerAccount("testcoin", user.String(), []types.AccessGrant{*types.NewAccessGrant(user,
		[]types.Access{types.Access_Mint, types.Access_Burn, types.Access_Withdraw, types.Access_Delete})})
	require.NoError(t, mac.SetManager(user))
	require.NoError(t, mac.SetSupply(sdk.NewInt64Coin("testcoin", 1000)))

	require.NoError(t, app.MarkerKeeper.AddMarkerAccount(ctx, mac))
	require.NoError(t, app.MarkerKeeper.SetNetAssetValue(ctx, mac, types.NewNetAssetValue(sdk.NewInt64Coin(types.UsdDenom, 1), 1), "test"))

	// insufficient supply to cover existing
	require.Error(t, app.MarkerKeeper.FinalizeMarker(ctx, user, "testcoin"))

	// move supply up high enough
	require.NoError(t, app.MarkerKeeper.MintCoin(ctx, user, sdk.NewInt64Coin("testcoin", 9001)))
	// no error now...
	require.NoError(t, app.MarkerKeeper.FinalizeMarker(ctx, user, "testcoin"))
	require.NoError(t, app.MarkerKeeper.ActivateMarker(ctx, user, "testcoin"))

	var err error
	var m types.MarkerAccountI
	m, err = app.MarkerKeeper.GetMarkerByDenom(ctx, "testcoin")
	require.NoError(t, err)
	require.NotNil(t, m)

	// Amount that was minted shal be 1
	require.EqualValues(t, 1, app.MarkerKeeper.GetEscrow(ctx, m).AmountOf("testcoin").Int64())
	// Amount of the total supply shall be 10001
	require.EqualValues(t, 10001, m.GetSupply().Amount.Int64())
}

func TestImplictControl(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)

	setAcc := func(addr sdk.AccAddress, sequence uint64) {
		acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
		require.NoError(t, acc.SetSequence(sequence), "%s.SetSequence(%d)", string(addr), sequence)
		app.AccountKeeper.SetAccount(ctx, acc)
	}

	user := testUserAddress("test")
	user2 := testUserAddress("test2")
	setAcc(user, 1)
	setAcc(user2, 1)

	// create account and check default values
	mac := types.NewEmptyMarkerAccount("testcoin", user.String(), []types.AccessGrant{*types.NewAccessGrant(user,
		[]types.Access{types.Access_Mint, types.Access_Burn, types.Access_Withdraw, types.Access_Delete})})

	mac.MarkerType = types.MarkerType_RestrictedCoin
	require.NoError(t, mac.SetManager(user))
	require.NoError(t, mac.SetSupply(sdk.NewInt64Coin("testcoin", 1000)))

	require.NoError(t, app.MarkerKeeper.AddMarkerAccount(ctx, mac))
	require.NoError(t, app.MarkerKeeper.SetNetAssetValue(ctx, mac, types.NewNetAssetValue(sdk.NewInt64Coin(types.UsdDenom, 1), 1), "test"))

	// Moves to finalized, mints required supply, moves to active status.
	require.NoError(t, app.MarkerKeeper.FinalizeMarker(ctx, user, "testcoin"))
	// No send enabled flag enforced yet, default is allowed
	require.True(t, app.BankKeeper.IsSendEnabledDenom(ctx, "testcoin"))
	require.NoError(t, app.MarkerKeeper.ActivateMarker(ctx, user, "testcoin"))

	// Activated restricted coins should be added to send enable, restriction checks are verified further in the stack, verify is true now
	require.True(t, app.BankKeeper.IsSendEnabledDenom(ctx, "testcoin"))

	// Must fail because user2 does not have any access
	require.Error(t, app.MarkerKeeper.AddAccess(ctx, user2, "testcoin", types.NewAccessGrant(user2,
		[]types.Access{types.Access_Mint, types.Access_Delete})))

	// Move 100% of the supply into user2.
	require.NoError(t, app.MarkerKeeper.WithdrawCoins(ctx, user, user2, "testcoin",
		sdk.NewCoins(sdk.NewInt64Coin("testcoin", 1000))))

	// Succeeds now because user2 is holding all of the testcoin supply.
	require.NoError(t, app.MarkerKeeper.AddAccess(ctx, user2, "testcoin",
		types.NewAccessGrant(user2, []types.Access{types.Access_Mint, types.Access_Delete, types.Access_Transfer})))

	// succeeds for a user with transfer rights
	require.NoError(t, app.MarkerKeeper.TransferCoin(ctx, user2, user, user2, sdk.NewInt64Coin("testcoin", 10)))
	// fails if the admin user does not have transfer authority
	require.Error(t, app.MarkerKeeper.TransferCoin(ctx, user, user2, user, sdk.NewInt64Coin("testcoin", 10)))

	// validate authz when 'from' is different from 'admin'
	granter := user
	grantee := user2
	now := ctx.BlockHeader().Time
	require.NotNil(t, now, "now")
	exp1Hour := now.Add(time.Hour)
	a := types.NewMarkerTransferAuthorization(sdk.NewCoins(sdk.NewInt64Coin("testcoin", 10)), []sdk.AccAddress{})

	// fails when admin user (grantee without authz permissions) has transfer authority
	require.Error(t, app.MarkerKeeper.TransferCoin(ctx, granter, user, grantee, sdk.NewInt64Coin("testcoin", 5)))
	// succeeds when admin user (grantee with authz permissions) has transfer authority
	require.NoError(t, app.AuthzKeeper.SaveGrant(ctx, grantee, granter, a, &exp1Hour))
	require.NoError(t, app.MarkerKeeper.TransferCoin(ctx, granter, user, grantee, sdk.NewInt64Coin("testcoin", 5)))
	// succeeds when admin user (grantee with authz permissions) has transfer authority (transfer remaining balance)
	require.NoError(t, app.MarkerKeeper.TransferCoin(ctx, granter, user, grantee, sdk.NewInt64Coin("testcoin", 5)))
	// fails when admin user (grantee with authz permissions) and transfer authority has transferred all coin ^^^ (grant has now been deleted)
	require.Error(t, app.MarkerKeeper.TransferCoin(ctx, granter, user, grantee, sdk.NewInt64Coin("testcoin", 5)))

	// validate authz when with allow list set
	now = ctx.BlockHeader().Time
	require.NotNil(t, now, "now")
	exp1Hour = now.Add(time.Hour)
	a = types.NewMarkerTransferAuthorization(sdk.NewCoins(sdk.NewInt64Coin("testcoin", 10)), []sdk.AccAddress{user})
	require.NoError(t, app.AuthzKeeper.SaveGrant(ctx, grantee, granter, a, &exp1Hour))
	// fails when admin user (grantee with authz permissions) has transfer authority but receiver is not on allowed list
	require.Error(t, app.MarkerKeeper.TransferCoin(ctx, granter, user2, grantee, sdk.NewInt64Coin("testcoin", 5)))
	// succeeds when admin user (grantee with authz permissions) has transfer authority with receiver is on allowed list
	require.NoError(t, app.MarkerKeeper.TransferCoin(ctx, granter, user, grantee, sdk.NewInt64Coin("testcoin", 5)))
}

func TestTransferCoin(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.NewContext(false)

	addrManager := sdk.AccAddress("manager_____________")
	addrTransOnly := sdk.AccAddress("transfer_only_______")
	addrForceTransOnly := sdk.AccAddress("force_transfer_only_")
	addrTransAndForce := sdk.AccAddress("transfer_and_force__")
	addrTransDepWithdraw := sdk.AccAddress("xfer_dep_withdraw___")
	addrNoTrans := sdk.AccAddress("all_but_transfer____")
	addrSeq0 := sdk.AccAddress("addr_w_sequence_zero")
	addr1 := sdk.AccAddress("addr_one____________")
	addr2 := sdk.AccAddress("addr_two____________")
	addr3 := sdk.AccAddress("addr_three__________")
	addr4 := sdk.AccAddress("addr_four___________")

	addrsToFund := []sdk.AccAddress{
		addrTransOnly, addrTransDepWithdraw, addrNoTrans,
		addrSeq0,
		addr1, addr2, addr3, addr4,
	}
	seq1Addrs := []sdk.AccAddress{
		addrManager,
		addrTransOnly, addrTransDepWithdraw, addrNoTrans,
		addr1, addr2, addr3, addr4,
	}

	denomCoin := "normalcoin"
	denomRestricted := "restrictedcoin"
	denomOnlyDeposit := "onlydepositcoin"
	denomOnlyWithdraw := "onlywithdrawcoin"
	denomForceTrans := "jedicoin"
	denomProposed := "propcoin"

	allAccessExcept := func(addr sdk.AccAddress, perms ...types.Access) types.AccessGrant {
		rv := types.AccessGrant{
			Address:     addr.String(),
			Permissions: nil,
		}
		for permVal := range types.Access_name {
			if permVal == 0 {
				continue
			}
			perm := types.Access(permVal)
			keep := true
			for _, ignore := range perms {
				if perm == ignore {
					keep = false
					break
				}
			}
			if keep {
				rv.Permissions = append(rv.Permissions, perm)
			}
		}
		sort.Slice(rv.Permissions, func(i, j int) bool {
			return rv.Permissions[i] < rv.Permissions[j]
		})
		return rv
	}
	allAccess := func(addr sdk.AccAddress) types.AccessGrant {
		return allAccessExcept(addr)
	}
	accessOnly := func(addr sdk.AccAddress, perms ...types.Access) types.AccessGrant {
		return *types.NewAccessGrant(addr, perms)
	}
	markerAddr := func(denom string) sdk.AccAddress {
		rv, err := types.MarkerAddress(denom)
		require.NoError(t, err, "MarkerAddress(%q)", denom)
		return rv
	}
	fundAcct := func(addr sdk.AccAddress, denom string) {
		fundAmt := sdk.NewCoins(sdk.NewInt64Coin(denom, 1_000_000))
		err := app.MarkerKeeper.WithdrawCoins(ctx, addrManager, addr, denom, fundAmt)
		require.NoError(t, err, "WithdrawCoins %s to %q", fundAmt, string(addr))
	}
	setupMarker := func(denom string, marker *types.MarkerAccount) sdk.AccAddress {
		addr := markerAddr(denom)
		marker.BaseAccount = authtypes.NewBaseAccountWithAddress(addr)
		marker.Denom = denom
		if marker.Supply.IsNil() {
			marker.Supply = sdkmath.NewInt(1_000_000_000)
		}
		if marker.Status == 0 {
			marker.Status = types.StatusProposed
		}
		if marker.MarkerType == 0 {
			marker.MarkerType = types.MarkerType_RestrictedCoin
		}
		if len(marker.Manager) == 0 {
			marker.Manager = addrManager.String()
		}
		managerHasAccess := false
		for _, grant := range marker.AccessControl {
			if grant.Address == marker.Manager {
				managerHasAccess = true
				break
			}
		}
		if !managerHasAccess {
			marker.AccessControl = append([]types.AccessGrant{allAccess(addrManager)}, marker.AccessControl...)
		}

		return addr
	}
	createMarker := func(denom string, marker *types.MarkerAccount) sdk.AccAddress {
		addr := setupMarker(denom, marker)

		err := app.MarkerKeeper.SetNetAssetValue(ctx, marker, types.NewNetAssetValue(sdk.NewInt64Coin(types.UsdDenom, 1), 1), "test")
		require.NoError(t, err, "SetNetAssetValue %q", denom)
		err = app.MarkerKeeper.AddFinalizeAndActivateMarker(ctx, marker)
		require.NoError(t, err, "AddFinalizeAndActivateMarker %q", denom)

		for _, fundAddr := range addrsToFund {
			fundAcct(fundAddr, denom)
		}
		return addr
	}
	createProposedMarker := func(denom string, marker *types.MarkerAccount) sdk.AccAddress {
		addr := setupMarker(denom, marker)

		fundAmt := sdk.NewCoins(sdk.NewInt64Coin(denom, 1_000_000))
		for _, fundAddr := range addrsToFund {
			err := testutil.FundAccount(types.WithBypass(ctx), app.BankKeeper, fundAddr, fundAmt)
			require.NoError(t, err, "FundAccount(%q, %q)", string(fundAddr), fundAmt)
		}

		err := app.MarkerKeeper.AddMarkerAccount(ctx, marker)
		require.NoError(t, err, "AddMarkerAccount %q", denom)

		return addr
	}
	setAcc := func(addr sdk.AccAddress, sequence uint64) {
		acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
		require.NoError(t, acc.SetSequence(sequence), "%s.SetSequence(%d)", string(addr), sequence)
		app.AccountKeeper.SetAccount(ctx, acc)
	}
	createGroup := func(admin sdk.AccAddress) sdk.AccAddress {
		msg := &group.MsgCreateGroupWithPolicy{
			Admin:              admin.String(),
			Members:            []group.MemberRequest{{Address: admin.String(), Weight: "1"}},
			GroupPolicyAsAdmin: true,
		}
		err := msg.SetDecisionPolicy(group.NewPercentageDecisionPolicy("0.5", time.Second, time.Second))
		require.NoError(t, err, "SetDecisionPolicy %q", string(admin))

		res, err := app.GroupKeeper.CreateGroupWithPolicy(ctx, msg)
		require.NoError(t, err, "CreateGroupWithPolicy %q", string(admin))

		rv, err := sdk.AccAddressFromBech32(res.GroupPolicyAddress)
		require.NoError(t, err, "AccAddressFromBech32(%q) (GroupPolicyAddress for %q)",
			res.GroupPolicyAddress, string(admin))
		return rv
	}
	noAccessErr := func(addr sdk.AccAddress, role types.Access, denom string) string {
		mAddr, err := types.MarkerAddress(denom)
		require.NoError(t, err, "MarkerAddress(%q)", denom)
		return fmt.Sprintf("%s does not have %s on %s marker (%s)", addr, role, denom, mAddr)
	}

	for _, addr := range seq1Addrs {
		setAcc(addr, 1)
	}
	setAcc(addrSeq0, 0)

	addrGroup := createGroup(addrManager)
	addrsToFund = append(addrsToFund, addrGroup)

	markerCoin := &types.MarkerAccount{
		AccessControl:          []types.AccessGrant{allAccessExcept(addrManager, types.Access_Transfer, types.Access_ForceTransfer)},
		MarkerType:             types.MarkerType_Coin,
		SupplyFixed:            true,
		AllowGovernanceControl: true,
		AllowForcedTransfer:    false,
	}
	markerAddrCoin := createMarker(denomCoin, markerCoin)

	markerRestricted := &types.MarkerAccount{
		AccessControl: []types.AccessGrant{
			accessOnly(addrTransOnly, types.Access_Transfer),
			accessOnly(addrForceTransOnly, types.Access_ForceTransfer),
			accessOnly(addrTransAndForce, types.Access_Transfer, types.Access_ForceTransfer),
			accessOnly(addrTransDepWithdraw, types.Access_Transfer),
			allAccessExcept(addrNoTrans, types.Access_Transfer, types.Access_ForceTransfer),
		},
		SupplyFixed:            true,
		AllowGovernanceControl: true,
		AllowForcedTransfer:    false,
	}
	createMarker(denomRestricted, markerRestricted)

	markerOnlyDep := &types.MarkerAccount{
		AccessControl:          []types.AccessGrant{accessOnly(addrTransDepWithdraw, types.Access_Deposit)},
		SupplyFixed:            true,
		AllowGovernanceControl: true,
		AllowForcedTransfer:    false,
	}
	markerAddrOnlyDep := createMarker(denomOnlyDeposit, markerOnlyDep)

	markerOnlyWithdraw := &types.MarkerAccount{
		AccessControl:          []types.AccessGrant{accessOnly(addrTransDepWithdraw, types.Access_Withdraw)},
		SupplyFixed:            true,
		AllowGovernanceControl: true,
		AllowForcedTransfer:    false,
	}
	markerAddrOnlyWithdraw := createMarker(denomOnlyWithdraw, markerOnlyWithdraw)

	markerForceTrans := &types.MarkerAccount{
		AccessControl: []types.AccessGrant{
			accessOnly(addrTransDepWithdraw, types.Access_Transfer),
			accessOnly(addrTransOnly, types.Access_Transfer),
			accessOnly(addrForceTransOnly, types.Access_ForceTransfer),
			accessOnly(addrTransAndForce, types.Access_Transfer, types.Access_ForceTransfer),
		},
		SupplyFixed:            true,
		AllowGovernanceControl: true,
		AllowForcedTransfer:    true,
	}
	createMarker(denomForceTrans, markerForceTrans)

	markerProposed := &types.MarkerAccount{
		SupplyFixed:            true,
		AllowGovernanceControl: true,
	}
	createProposedMarker(denomProposed, markerProposed)

	// The only-withdraw marker needs to have some coins, because, reasons (needed for test cases).
	fundAcct(markerAddrOnlyWithdraw, denomRestricted)
	fundAcct(markerAddrOnlyWithdraw, denomForceTrans)

	tests := []struct {
		name        string
		authzKeeper *MockAuthzKeeper
		bankKeeper  *WrappedBankKeeper
		from        sdk.AccAddress
		to          sdk.AccAddress
		admin       sdk.AccAddress
		amount      sdk.Coin
		expErr      string
	}{
		{
			name:   "marker not found",
			from:   addr1,
			to:     addr2,
			admin:  addr3,
			amount: sdk.NewInt64Coin("nosuchmarker", 5),
			expErr: "marker not found for nosuchmarker: marker nosuchmarker not found for address: " + markerAddr("nosuchmarker").String(),
		},
		{
			name:   "marker not active",
			from:   addr1,
			to:     addr2,
			admin:  addrManager,
			amount: sdk.NewInt64Coin(denomProposed, 78),
			expErr: "marker status (proposed) is not active, funds cannot be moved",
		},
		{
			name:   "marker not restricted",
			from:   addr1,
			to:     addr2,
			admin:  addr3,
			amount: sdk.NewInt64Coin(denomCoin, 12),
			expErr: "marker type is not restricted_coin, brokered transfer not supported",
		},
		{
			name:   "admin does not have transfer or force transfer",
			from:   addr1,
			to:     addr2,
			admin:  addrNoTrans,
			amount: sdk.NewInt64Coin(denomRestricted, 8),
			expErr: noAccessErr(addrNoTrans, types.Access_Transfer, denomRestricted),
		},
		{
			name:   "going to restricted: admin does not have deposit",
			from:   addrTransOnly,
			to:     markerAddrOnlyDep,
			admin:  addrTransOnly,
			amount: sdk.NewInt64Coin(denomRestricted, 14),
			expErr: noAccessErr(addrTransOnly, types.Access_Deposit, denomOnlyDeposit),
		},
		{
			name:   "going to restricted: admin has deposit",
			from:   addrTransDepWithdraw,
			to:     markerAddrOnlyDep,
			admin:  addrTransDepWithdraw,
			amount: sdk.NewInt64Coin(denomRestricted, 9),
		},
		{
			name:   "going to unrestricted",
			from:   addrTransOnly,
			to:     markerAddrCoin,
			admin:  addrTransOnly,
			amount: sdk.NewInt64Coin(denomRestricted, 17),
		},
		{
			name:        "admin not from: no force transfer: no authz",
			authzKeeper: NewMockAuthzKeeper().WithAuthzHandlerNoAuth(),
			from:        addr4,
			to:          addr3,
			admin:       addrTransAndForce,
			amount:      sdk.NewInt64Coin(denomRestricted, 11),
			expErr:      addrTransAndForce.String() + " account has not been granted authority to withdraw from " + addr4.String() + " account",
		},
		{
			name:        "admin not from: no force transfer: with authz",
			authzKeeper: NewMockAuthzKeeper().WithAuthzHandlerSuccess(),
			from:        addr3,
			to:          addr1,
			admin:       addrTransOnly,
			amount:      sdk.NewInt64Coin(denomRestricted, 23),
		},
		{
			name:        "admin not from: no force transfer: from marker: with withdraw",
			authzKeeper: NewMockAuthzKeeper().WithAuthzHandlerSuccess(),
			from:        markerAddrOnlyWithdraw,
			to:          addr2,
			admin:       addrTransDepWithdraw,
			amount:      sdk.NewInt64Coin(denomRestricted, 2),
		},
		{
			name:        "admin not from: no force transfer access: no authz",
			authzKeeper: NewMockAuthzKeeper().WithAuthzHandlerNoAuth(),
			from:        addr4,
			to:          addr3,
			admin:       addrTransOnly,
			amount:      sdk.NewInt64Coin(denomForceTrans, 11),
			expErr:      addrTransOnly.String() + " account has not been granted authority to withdraw from " + addr4.String() + " account",
		},
		{
			name:        "admin not from: no force transfer access: with authz",
			authzKeeper: NewMockAuthzKeeper().WithAuthzHandlerSuccess(),
			from:        addr3,
			to:          addr1,
			admin:       addrTransOnly,
			amount:      sdk.NewInt64Coin(denomForceTrans, 23),
		},
		{
			name:        "admin not from: no force transfer access: from marker: with withdraw",
			authzKeeper: NewMockAuthzKeeper().WithAuthzHandlerSuccess(),
			from:        markerAddrOnlyWithdraw,
			to:          addr2,
			admin:       addrTransDepWithdraw,
			amount:      sdk.NewInt64Coin(denomForceTrans, 2),
		},
		{
			name:        "admin not from: force transfer okay: no force transfer access",
			authzKeeper: NewMockAuthzKeeper().WithAuthzHandlerNoAuth(),
			from:        addr4,
			to:          addr1,
			admin:       addrTransOnly,
			amount:      sdk.NewInt64Coin(denomForceTrans, 19),
			expErr:      addrTransOnly.String() + " account has not been granted authority to withdraw from " + addr4.String() + " account",
		},
		{
			name:   "admin not from: force transfer okay: only force access",
			from:   addr4,
			to:     addr1,
			admin:  addrForceTransOnly,
			amount: sdk.NewInt64Coin(denomForceTrans, 20),
		},
		{
			name:   "admin not from: force transfer: from is a group account",
			from:   addrGroup,
			to:     addr3,
			admin:  addrTransAndForce,
			amount: sdk.NewInt64Coin(denomForceTrans, 15),
		},
		{
			name:   "admin not from: force transfer: from account has sequence zero",
			from:   addrSeq0,
			to:     addr1,
			admin:  addrTransAndForce,
			amount: sdk.NewInt64Coin(denomForceTrans, 7),
			expErr: "funds are not allowed to be removed from " + addrSeq0.String(),
		},
		{
			name:   "admin not from: force transfer: from okay account",
			from:   addr1,
			to:     addr2,
			admin:  addrTransAndForce,
			amount: sdk.NewInt64Coin(denomForceTrans, 41),
		},
		{
			name:       "to blocked account",
			bankKeeper: NewWrappedBankKeeper().WithExtraBlockedAddrs(addr3),
			from:       addrTransOnly,
			to:         addr3,
			admin:      addrTransOnly,
			amount:     sdk.NewInt64Coin(denomRestricted, 57),
			expErr:     addr3.String() + " is not allowed to receive funds",
		},
		{
			name:       "send error",
			bankKeeper: NewWrappedBankKeeper().WithSendCoinsErrs("nope, not gonna allow that"),
			from:       addr1,
			to:         addr2,
			admin:      addrForceTransOnly,
			amount:     sdk.NewInt64Coin(denomForceTrans, 48),
			expErr:     "nope, not gonna allow that",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.authzKeeper == nil {
				tc.authzKeeper = NewMockAuthzKeeper().WithAuthzHandlerAcceptError("injected authorization.Accept error")
			}
			kpr := app.MarkerKeeper.WithAuthzKeeper(tc.authzKeeper)
			if tc.bankKeeper != nil {
				kpr = kpr.WithBankKeeper(tc.bankKeeper.WithParent(app.BankKeeper))
			}

			origFromBal := app.BankKeeper.GetAllBalances(ctx, tc.from)
			origToBal := app.BankKeeper.GetAllBalances(ctx, tc.to)
			origAdminBal := app.BankKeeper.GetAllBalances(ctx, tc.admin)
			expFromBal, expToBal, expAdminBal := origFromBal, origToBal, origAdminBal
			if len(tc.expErr) == 0 {
				var hasNeg bool
				expFromBal, hasNeg = origFromBal.SafeSub(tc.amount)
				require.False(t, hasNeg, "%q.SafeSub(%q) has negative result", origFromBal, tc.amount)
				expToBal = origToBal.Add(tc.amount)
				switch {
				case tc.admin.Equals(tc.to):
					expAdminBal = expToBal
				case tc.admin.Equals(tc.from):
					expAdminBal = expFromBal
				}
			}

			var expEvents sdk.Events
			if len(tc.expErr) == 0 {
				expEvents = sdk.Events{
					{
						Type: "provenance.marker.v1.EventMarkerTransfer",
						Attributes: []abci.EventAttribute{
							{Key: "administrator", Value: `"` + tc.admin.String() + `"`},
							{Key: "amount", Value: `"` + tc.amount.Amount.String() + `"`},
							{Key: "denom", Value: `"` + tc.amount.Denom + `"`},
							{Key: "from_address", Value: `"` + tc.from.String() + `"`},
							{Key: "to_address", Value: `"` + tc.to.String() + `"`},
						},
					},
				}
			}

			em := sdk.NewEventManager()
			ctx = app.NewContext(false).WithEventManager(em)
			var err error
			testFunc := func() {
				err = kpr.TransferCoin(ctx, tc.from, tc.to, tc.admin, tc.amount)
			}
			require.NotPanics(t, testFunc, "TransferCoin")
			assertions.AssertErrorValue(t, err, tc.expErr, "TransferCoin error")
			actEvents := em.Events()
			assertions.AssertEventsContains(t, expEvents, actEvents, "events emitted during TransferCoin")

			actFromBal := app.BankKeeper.GetAllBalances(ctx, tc.from)
			actToBal := app.BankKeeper.GetAllBalances(ctx, tc.to)
			actAdminBal := app.BankKeeper.GetAllBalances(ctx, tc.admin)
			assert.Equal(t, expFromBal.String(), actFromBal.String(), "from balance after TransferCoin")
			assert.Equal(t, expToBal.String(), actToBal.String(), "to balance after TransferCoin")
			assert.Equal(t, expAdminBal.String(), actAdminBal.String(), "admin balance after TransferCoin")
		})
	}
}

func TestForceTransfer(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)

	setAcc := func(addr sdk.AccAddress, sequence uint64) {
		acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
		require.NoError(t, acc.SetSequence(sequence), "%s.SetSequence(%d)", string(addr), sequence)
		app.AccountKeeper.SetAccount(ctx, acc)
	}

	admin := sdk.AccAddress("admin_account_______")
	other := sdk.AccAddress("other_account_______")
	seq0 := sdk.AccAddress("sequence_0__________")
	setAcc(admin, 1)
	setAcc(other, 1)
	setAcc(seq0, 0)

	// Shorten up the lines making Coins.
	cz := func(coins ...sdk.Coin) sdk.Coins {
		return sdk.NewCoins(coins...)
	}

	accessList := []types.AccessGrant{{
		Address: admin.String(),
		Permissions: []types.Access{
			types.Access_Transfer,
			types.Access_Mint, types.Access_Burn, types.Access_Deposit,
			types.Access_Withdraw, types.Access_Delete, types.Access_Admin,
		},
	}}

	noForceDenom := "noforcecoin"
	noForceCoin := func(amt int64) sdk.Coin {
		return sdk.NewInt64Coin(noForceDenom, amt)
	}
	noForceAddr := types.MustGetMarkerAddress(noForceDenom)
	noForceMac := types.NewMarkerAccount(
		authtypes.NewBaseAccount(noForceAddr, nil, 0, 0),
		noForceCoin(1111),
		admin,
		accessList,
		types.StatusProposed,
		types.MarkerType_RestrictedCoin,
		true,
		true,
		false,
		[]string{},
	)
	require.NoError(t, app.MarkerKeeper.SetNetAssetValue(ctx, noForceMac, types.NewNetAssetValue(sdk.NewInt64Coin(types.UsdDenom, 1), 1), "test"))
	require.NoError(t, app.MarkerKeeper.AddFinalizeAndActivateMarker(ctx, noForceMac),
		"AddFinalizeAndActivateMarker without force transfer")

	wForceDenom := "withforcecoin"
	wForceCoin := func(amt int64) sdk.Coin {
		return sdk.NewInt64Coin(wForceDenom, amt)
	}
	wForceAddr := types.MustGetMarkerAddress(wForceDenom)
	wForceMac := types.NewMarkerAccount(
		authtypes.NewBaseAccount(wForceAddr, nil, 0, 0),
		wForceCoin(2222),
		admin,
		accessList,
		types.StatusProposed,
		types.MarkerType_RestrictedCoin,
		true,
		true,
		true,
		[]string{},
	)
	require.NoError(t, app.MarkerKeeper.SetNetAssetValue(ctx, wForceMac, types.NewNetAssetValue(sdk.NewInt64Coin(types.UsdDenom, 1), 1), "test"))
	require.NoError(t, app.MarkerKeeper.AddFinalizeAndActivateMarker(ctx, wForceMac),
		"AddFinalizeAndActivateMarker with force transfer")

	noForceBal := cz(noForceMac.GetSupply())
	wForceBal := cz(wForceMac.GetSupply())
	var adminBal, otherBal, seq0Bal sdk.Coins
	requireBalances := func(tt *testing.T, desc string) {
		tt.Helper()
		ok := assert.Equal(tt, noForceBal.String(), app.BankKeeper.GetAllBalances(ctx, noForceAddr).String(),
			"%s: no-force-transfer marker balance", desc)
		ok = assert.Equal(tt, wForceBal.String(), app.BankKeeper.GetAllBalances(ctx, wForceAddr).String(),
			"%s: with-force-transfer marker balance", desc) && ok
		ok = assert.Equal(tt, adminBal.String(), app.BankKeeper.GetAllBalances(ctx, admin).String(),
			"%s: admin balance", desc) && ok
		ok = assert.Equal(tt, otherBal.String(), app.BankKeeper.GetAllBalances(ctx, other).String(),
			"%s: other balance", desc) && ok
		ok = assert.Equal(tt, seq0Bal.String(), app.BankKeeper.GetAllBalances(ctx, seq0).String(),
			"%s: sequence 0 balance", desc) && ok
		if !ok {
			tt.FailNow()
		}
	}
	requireBalances(t, "starting")

	// Have the admin withdraw funds of each to the other account.
	toWithdraw := cz(noForceCoin(111))
	otherBal = otherBal.Add(toWithdraw...)
	noForceBal = noForceBal.Sub(toWithdraw...)
	require.NoError(t, app.MarkerKeeper.WithdrawCoins(ctx, admin, other, noForceDenom, toWithdraw),
		"withdraw 500noforceback to other")
	toWithdraw = cz(wForceCoin(222))
	otherBal = otherBal.Add(toWithdraw...)
	wForceBal = wForceBal.Sub(toWithdraw...)
	require.NoError(t, app.MarkerKeeper.WithdrawCoins(ctx, admin, other, wForceDenom, toWithdraw),
		"withdraw 500wforceback to other")
	requireBalances(t, "after withdraws")

	// Have the admin try a transfer of the no-force-transfer from that other account to itself. It should fail.
	assert.EqualError(t, app.MarkerKeeper.TransferCoin(ctx, other, admin, admin, noForceCoin(11)),
		fmt.Sprintf("%s account has not been granted authority to withdraw from %s account", admin, other),
		"transfer of non-force-transfer coin from other account back to admin")
	requireBalances(t, "after failed transfer")

	// Have the admin try a transfer of the force-transfer, but without the force-transfer permission.
	assert.EqualError(t, app.MarkerKeeper.TransferCoin(ctx, other, admin, admin, wForceCoin(7)),
		fmt.Sprintf("%s account has not been granted authority to withdraw from %s account", admin, other),
		"transfer of force-transfer coin by account without force-transfer access")
	requireBalances(t, "after failed force-transfer")

	// Give the admin force transfer permission now.
	addFTGrant := &types.AccessGrant{Address: admin.String(), Permissions: types.AccessList{types.Access_ForceTransfer}}
	require.NoError(t, app.MarkerKeeper.AddAccess(ctx, admin, wForceDenom, addFTGrant),
		"AddAccess to grant admin force-transfer access")

	// Have the admin try a transfer of the w/force transfer from that other account to itself. It should go through.
	transferCoin := wForceCoin(22)
	assert.NoError(t, app.MarkerKeeper.TransferCoin(ctx, other, admin, admin, transferCoin),
		"transfer of force-transferrable coin from other account back to admin")
	otherBal = otherBal.Sub(transferCoin)
	adminBal = adminBal.Add(transferCoin)
	requireBalances(t, "after successful transfer")

	// Fund the sequence 0 account and have the admin try a transfer of the w/force transfer from there to itself. This should fail.
	seq0Bal = cz(wForceCoin(5))
	wForceBal = wForceBal.Sub(seq0Bal...)
	require.NoError(t, app.MarkerKeeper.WithdrawCoins(ctx, admin, seq0, wForceDenom, seq0Bal),
		"withdraw 500wforceback to other")
	requireBalances(t, "funds withdrawn to sequence 0 address")
	assert.EqualError(t, app.MarkerKeeper.TransferCoin(ctx, seq0, admin, admin, wForceCoin(2)),
		fmt.Sprintf("funds are not allowed to be removed from %s", seq0),
		"transfer of force-transfer coin from account with sequence 0 back to admin",
	)
	requireBalances(t, "after failed force transfer")
}

func TestCanForceTransferFrom(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)

	testAddr := func(prefix string) sdk.AccAddress {
		return sdk.AccAddress(prefix + strings.Repeat("_", 20-len(prefix)))
	}

	setAcc := func(addrPrefix string, sequence uint64) sdk.AccAddress {
		addr := testAddr(addrPrefix)
		acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
		require.NoError(t, acc.SetSequence(sequence), "%s.SetSequence(%d)", string(addr), sequence)
		app.AccountKeeper.SetAccount(ctx, acc)
		return addr
	}

	createGroup := func() sdk.AccAddress {
		msg, err := group.NewMsgCreateGroupWithPolicy("cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
			[]group.MemberRequest{
				{
					Address:  "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
					Weight:   "1",
					Metadata: "",
				},
			},
			"", "", true, group.NewPercentageDecisionPolicy("0.5", time.Second, time.Second))
		require.NoError(t, err, "NewMsgCreateGroupWithPolicy")
		res, err := app.GroupKeeper.CreateGroupWithPolicy(ctx, msg)
		require.NoError(t, err, "CreateGroupWithPolicy")

		return sdk.MustAccAddressFromBech32(res.GroupPolicyAddress)
	}

	createMarkerAcc := func(addrPrefix string) sdk.AccAddress {
		addr := testAddr(addrPrefix)
		acc := &types.MarkerAccount{
			BaseAccount: authtypes.NewBaseAccountWithAddress(addr),
			Status:      types.StatusActive,
			Denom:       "whatever",
			Supply:      sdkmath.NewInt(0),
			MarkerType:  types.MarkerType_RestrictedCoin,
		}
		setNewAccount(app, ctx, acc)
		return addr
	}

	createMarketAcc := func(addrPrefix string) sdk.AccAddress {
		addr := testAddr(addrPrefix)
		acc := &exchange.MarketAccount{
			BaseAccount:   authtypes.NewBaseAccountWithAddress(addr),
			MarketId:      97531,
			MarketDetails: exchange.MarketDetails{},
		}
		setNewAccount(app, ctx, acc)
		return addr
	}

	addrNoAcc := testAddr("addrNoAcc")
	addrSeq0 := setAcc("addrSeq0", 0)
	addrSeq1 := setAcc("addrSeq1", 1)
	addrGroup := createGroup()
	addrMarker := createMarkerAcc("addrMarker")
	addrMarket := createMarketAcc("addrMarket")

	tests := []struct {
		name string
		from sdk.AccAddress
		exp  bool
	}{
		{name: "address without an account", from: addrNoAcc, exp: true},
		{name: "address with sequence 0", from: addrSeq0, exp: false},
		{name: "address with sequence 1", from: addrSeq1, exp: true},
		{name: "group address", from: addrGroup, exp: true},
		{name: "marker address", from: addrMarker, exp: true},
		{name: "market address", from: addrMarket, exp: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := app.MarkerKeeper.CanForceTransferFrom(ctx, tc.from)
			assert.Equal(t, tc.exp, actual, "canForceTransferFrom")
		})
	}
}

func TestMarkerFeeGrant(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)
	server := markerkeeper.NewMsgServerImpl(app.MarkerKeeper)

	addr := types.MustGetMarkerAddress("testcoin")
	user := testUserAddress("admin")

	// no account before its created
	acc := app.AccountKeeper.GetAccount(ctx, addr)
	require.Nil(t, acc)

	// create account and check default values
	acc = types.NewEmptyMarkerAccount("testcoin", user.String(), nil)
	mac, ok := acc.(types.MarkerAccountI)
	require.True(t, ok)
	require.NotNil(t, mac)
	require.Equal(t, addr, mac.GetAddress())
	require.EqualValues(t, nil, mac.GetPubKey())

	// NewAccount doesn't call Set, so it's still nil
	require.Nil(t, app.AccountKeeper.GetAccount(ctx, addr))

	// set some values on the account and save it
	require.NoError(t, mac.GrantAccess(types.NewAccessGrant(user, []types.Access{types.Access_Mint, types.Access_Admin})))

	setNewAccount(app, ctx, mac)

	existingSupply := sdk.NewInt64Coin("testcoin", 10000)
	require.NoError(t, testutil.FundAccount(types.WithBypass(ctx), app.BankKeeper, user, sdk.NewCoins(existingSupply)), "funding accont")

	allowance, err := types.NewMsgGrantAllowance(
		"testcoin",
		user,
		testUserAddress("grantee"),
		&feegrant.BasicAllowance{SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("testcoin", 1))})
	require.NoError(t, err, "basic allowance creation failed")
	_, err = server.GrantAllowance(ctx, allowance)
	require.NoError(t, err, "failed to grant basic allowance from admin")
}

// testUserAddress gives a quick way to make a valid test address (no keys though)
func testUserAddress(name string) sdk.AccAddress {
	addr := types.MustGetMarkerAddress(name)
	return addr
}

func TestAddFinalizeActivateMarker(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)
	server := markerkeeper.NewMsgServerImpl(app.MarkerKeeper)

	addr := types.MustGetMarkerAddress("testcoin")
	pubkey := secp256k1.GenPrivKey().PubKey()
	user := testUserAddress("testcoin")
	manager := testUserAddress("manager")
	existingBalance := sdk.NewInt64Coin("coin", 1000)

	// prefund the marker address so an account gets created before the marker does.
	setNewAccount(app, ctx, authtypes.NewBaseAccount(user, pubkey, 0, 0))
	require.NoError(t, testutil.FundAccount(ctx, app.BankKeeper, addr, sdk.NewCoins(existingBalance)), "funding account")
	require.Equal(t, existingBalance, app.BankKeeper.GetBalance(ctx, addr, "coin"), "account balance must be set")

	// Creating a marker over an account with zero sequence is fine.
	// One shot marker creation
	_, err := server.AddFinalizeActivateMarker(ctx, types.NewMsgAddFinalizeActivateMarkerRequest(
		"testcoin",
		sdkmath.NewInt(30),
		user,
		manager,
		types.MarkerType_Coin,
		true,
		true,
		false,
		[]string{},
		[]types.AccessGrant{*types.NewAccessGrant(manager, []types.Access{types.Access_Mint, types.Access_Admin})},
		0,
		0,
	))
	require.NoError(t, err, "should allow a marker over existing account that has not signed anything.")

	// existing coin balance must still be present
	require.Equal(t, existingBalance, app.BankKeeper.GetBalance(ctx, addr, "coin"), "account balances must be preserved")

	m, err := app.MarkerKeeper.GetMarkerByDenom(ctx, "testcoin")
	require.NoError(t, err)
	require.EqualValues(t, m.GetSupply(), sdk.NewInt64Coin("testcoin", 30))
	require.EqualValues(t, m.GetStatus(), types.StatusActive)

	m, err = app.MarkerKeeper.GetMarker(ctx, user)
	require.NoError(t, err)
	require.EqualValues(t, m.GetSupply(), sdk.NewInt64Coin("testcoin", 30))
	require.EqualValues(t, m.GetStatus(), types.StatusActive)

	// Creating a marker over an existing marker fails.
	_, err = server.AddFinalizeActivateMarker(ctx, types.NewMsgAddFinalizeActivateMarkerRequest(
		"testcoin",
		sdkmath.NewInt(30),
		user,
		manager,
		types.MarkerType_Coin,
		true,
		true,
		false,
		[]string{},
		[]types.AccessGrant{*types.NewAccessGrant(manager, []types.Access{types.Access_Mint, types.Access_Admin})},
		0,
		0,
	))
	require.Error(t, err, "fails because marker already exists")

	// Load the created marker
	m, err = app.MarkerKeeper.GetMarker(ctx, user)
	require.NoError(t, err)
	require.EqualValues(t, m.GetSupply(), sdk.NewInt64Coin("testcoin", 30))
	require.EqualValues(t, m.GetStatus(), types.StatusActive)

	// entire supply should have been allocated to marker acount
	require.EqualValues(t, app.MarkerKeeper.GetEscrow(ctx, m).AmountOf("testcoin"), sdkmath.NewInt(30))
}

// Creating a marker over an existing account with a positive sequence number fails.
func TestInvalidAccount(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)
	pubkey := secp256k1.GenPrivKey().PubKey()
	server := markerkeeper.NewMsgServerImpl(app.MarkerKeeper)
	user := testUserAddress("testcoin")
	manager := testUserAddress("manager")

	// replace existing test account with a new copy that has a positive sequence number
	setNewAccount(app, ctx, authtypes.NewBaseAccount(user, pubkey, 0, 10))

	_, err := server.AddFinalizeActivateMarker(ctx, types.NewMsgAddFinalizeActivateMarkerRequest(
		"testcoin",
		sdkmath.NewInt(30),
		user,
		manager,
		types.MarkerType_Coin,
		true,
		true,
		false,
		[]string{},
		[]types.AccessGrant{*types.NewAccessGrant(manager, []types.Access{types.Access_Mint, types.Access_Admin})},
		0,
		0,
	))
	require.Error(t, err, "should not allow creation over and existing account with a positive sequence number.")
	require.Contains(t, err.Error(), "account at "+user.String()+" is not a marker account: invalid request")
}

func TestAddFinalizeActivateMarkerUnrestrictedDenoms(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)
	server := markerkeeper.NewMsgServerImpl(app.MarkerKeeper)

	user := testUserAddress("test")

	// Require a long unrestricted denom
	app.MarkerKeeper.SetParams(ctx, types.Params{UnrestrictedDenomRegex: "[a-z]{12,20}"})

	_, err := server.AddFinalizeActivateMarker(ctx,
		types.NewMsgAddFinalizeActivateMarkerRequest(
			"tooshort",
			sdkmath.NewInt(30),
			user,
			user,
			types.MarkerType_Coin,
			true,
			true,
			false,
			[]string{},
			[]types.AccessGrant{*types.NewAccessGrant(user, []types.Access{types.Access_Mint, types.Access_Admin})},
			0,
			0,
		))
	require.Error(t, err, "fails with unrestricted denom length fault")
	require.Equal(t, fmt.Errorf("invalid denom [tooshort] (fails unrestricted marker denom validation [a-z]{12,20})"), err, "should fail with denom restriction")

	_, err = server.AddFinalizeActivateMarker(ctx, types.NewMsgAddFinalizeActivateMarkerRequest(
		"itslongenough",
		sdkmath.NewInt(30),
		user,
		user,
		types.MarkerType_Coin,
		true,
		true,
		false,
		[]string{},
		[]types.AccessGrant{*types.NewAccessGrant(user, []types.Access{types.Access_Mint, types.Access_Admin})},
		0,
		0,
	))
	require.NoError(t, err, "should allow a marker with a sufficiently long denom")

	// Set to an empty string (returns to default expression)
	app.MarkerKeeper.SetParams(ctx, types.Params{UnrestrictedDenomRegex: ""})
	_, err = server.AddFinalizeActivateMarker(ctx, types.NewMsgAddFinalizeActivateMarkerRequest(
		"short",
		sdkmath.NewInt(30),
		user,
		user,
		types.MarkerType_Coin,
		true,
		true,
		false,
		[]string{},
		[]types.AccessGrant{*types.NewAccessGrant(user, []types.Access{types.Access_Mint, types.Access_Admin})},
		0,
		0,
	))
	// succeeds now as the default unrestricted denom expression allows any valid denom (minimum length is 2)
	require.NoError(t, err, "should allow any valid denom with a min length of two")
}

func TestAddMarkerViaProposal(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)
	server := markerkeeper.NewMsgServerImpl(app.MarkerKeeper)

	newMsg := func(denom string, amt sdkmath.Int, manager string, status types.MarkerStatus,
		markerType types.MarkerType, access []types.AccessGrant, allowGov bool,
	) *types.MsgAddMarkerRequest {
		return &types.MsgAddMarkerRequest{
			Amount:                 sdk.NewCoin(denom, amt),
			Manager:                manager,
			FromAddress:            app.MarkerKeeper.GetAuthority(),
			Status:                 status,
			MarkerType:             markerType,
			AccessList:             access,
			SupplyFixed:            true,
			AllowGovernanceControl: allowGov,
		}
	}

	user := testUserAddress("test").String()

	coin := types.MarkerType_Coin
	restricted := types.MarkerType_RestrictedCoin

	active := types.StatusActive
	finalized := types.StatusFinalized

	testCases := []struct {
		name string
		prop *types.MsgAddMarkerRequest
		err  error
	}{
		{
			"add marker - valid",
			newMsg("test1", sdkmath.NewInt(100), "", active, coin, []types.AccessGrant{}, true),
			nil,
		},
		{
			"add marker - valid restricted marker",
			newMsg("testrestricted", sdkmath.NewInt(100), "", active, restricted, []types.AccessGrant{}, true),
			nil,
		},
		{
			"add marker - valid no governance",
			newMsg("testnogov", sdkmath.NewInt(100), user, active, coin, []types.AccessGrant{}, false),
			nil,
		},
		{
			"add marker - valid finalized",
			newMsg("pending", sdkmath.NewInt(100), user, finalized, coin, []types.AccessGrant{}, true),
			nil,
		},
		{
			"add marker - already exists",
			newMsg("test1", sdkmath.NewInt(0), "", active, coin, []types.AccessGrant{}, true),
			fmt.Errorf("marker address already exists for cosmos1ku2jzvpkt4ffxxaajyk2r88axk9cr5jqlthcm4: invalid request"),
		},
		{
			"add marker - invalid status",
			newMsg("test2", sdkmath.NewInt(100), "", types.StatusUndefined, coin, []types.AccessGrant{}, true),
			fmt.Errorf("invalid marker status: invalid request"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := server.AddMarker(ctx, tc.prop)
			if tc.err == nil {
				require.NoError(t, err)
				require.Equal(t, res, res)
			} else {
				require.Nil(t, res)
				require.EqualError(t, err, tc.err.Error())
			}
		})
	}
}

func TestMsgSupplyIncreaseProposalRequest(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)
	server := markerkeeper.NewMsgServerImpl(app.MarkerKeeper)

	authority := "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"
	invalidAuthority := testUserAddress("test")
	targetAddress := "cosmos1tnh2q55v8wyygtt9srz5safamzdengsnqeycj3"

	// add markers for test cases
	user := testUserAddress("test")
	govDisabledDenom := types.NewEmptyMarkerAccount(
		"denom-with-gov-disabled",
		user.String(),
		[]types.AccessGrant{})
	govDisabledDenom.AllowGovernanceControl = false

	err := app.MarkerKeeper.AddMarkerAccount(ctx, govDisabledDenom)
	require.NoError(t, err)

	nonActiveDenom := types.NewEmptyMarkerAccount(
		"denom-with-non-active-status",
		user.String(),
		[]types.AccessGrant{})
	nonActiveDenom.Status = types.StatusDestroyed

	err = app.MarkerKeeper.AddMarkerAccount(ctx, nonActiveDenom)
	require.NoError(t, err)

	// exceed max supply
	kneesBees := types.NewEmptyMarkerAccount(
		"knees-bees",
		user.String(),
		[]types.AccessGrant{})
	kneesBees.Status = types.StatusActive

	err = app.MarkerKeeper.AddMarkerAccount(ctx, kneesBees)
	require.NoError(t, err)

	// all good
	beesKnees := types.NewEmptyMarkerAccount(
		"bees-knees",
		user.String(),
		[]types.AccessGrant{})
	beesKnees.Status = types.StatusProposed

	err = app.MarkerKeeper.AddMarkerAccount(ctx, beesKnees)
	require.NoError(t, err)

	testCases := []struct {
		name          string
		amount        sdk.Coin
		targetAddress string
		authority     string
		shouldFail    bool
		expectedError string
	}{
		{
			name: "invalid authority",
			amount: sdk.Coin{
				Amount: sdkmath.NewInt(100),
				Denom:  "invalid-authority-denom",
			},
			targetAddress: targetAddress,
			authority:     invalidAuthority.String(),
			shouldFail:    true,
			expectedError: "expected cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn got cosmos1ppvtnfw30fvpdutnjzuavntqk0a34xfsrlnsg2: expected gov account as only signer for proposal message",
		},
		{
			name: "marker does not exist",
			amount: sdk.Coin{
				Amount: sdkmath.NewInt(100),
				Denom:  "unknown-denom",
			},
			targetAddress: authority,
			authority:     authority,
			shouldFail:    true,
			expectedError: "unknown-denom marker does not exist",
		},
		{
			name: "marker with governance disabled",
			amount: sdk.Coin{
				Amount: sdkmath.NewInt(100),
				Denom:  govDisabledDenom.Denom,
			},
			targetAddress: authority,
			authority:     authority,
			shouldFail:    true,
			expectedError: "denom-with-gov-disabled marker does not allow governance control",
		},
		{
			name: "marker status is not StatusActive",
			amount: sdk.Coin{
				Amount: sdkmath.NewInt(100),
				Denom:  nonActiveDenom.Denom,
			},
			targetAddress: authority,
			authority:     authority,
			shouldFail:    true,
			expectedError: "cannot mint coin for a marker that is not in Active status",
		},
		{
			name: "all good",
			amount: sdk.Coin{
				Amount: sdkmath.NewInt(100),
				Denom:  beesKnees.Denom,
			},
			targetAddress: authority,
			authority:     authority,
			shouldFail:    false,
			expectedError: "",
		},
		{
			name: "supply increase exceeds max supply",
			amount: sdk.Coin{
				Amount: types.StringToBigInt("1000000000000000000000"),
				Denom:  kneesBees.Denom,
			},
			targetAddress: authority,
			authority:     authority,
			shouldFail:    true,
			expectedError: "requested supply 1000000000000000000000 exceeds maximum allowed value 100000000000000000000",
		},
	}

	for _, tc := range testCases {
		res, err := server.SupplyIncreaseProposal(ctx,
			types.NewMsgSupplyIncreaseProposalRequest(tc.amount, tc.targetAddress, tc.authority))

		if tc.shouldFail {
			require.Nil(t, res)
			require.EqualError(t, err, tc.expectedError)

		} else {
			require.NoError(t, err)
			require.Equal(t, res, &types.MsgSupplyIncreaseProposalResponse{})
		}
	}
}

func TestMsgUpdateRequiredAttributesRequest(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)
	server := markerkeeper.NewMsgServerImpl(app.MarkerKeeper)

	authority := authtypes.NewModuleAddress(govtypes.ModuleName)

	transferAuthUser := testUserAddress("test")
	notTransferAuthUser := testUserAddress("test1")

	notRestrictedMarker := types.NewEmptyMarkerAccount(
		"not-restricted-marker",
		transferAuthUser.String(),
		[]types.AccessGrant{})

	err := app.MarkerKeeper.AddMarkerAccount(ctx, notRestrictedMarker)
	require.NoError(t, err)
	reqAttr := []string{"foo.provenance.io", "*.provenance.io", "bar.provenance.io"}
	rMarkerDenom := "restricted-marker"
	rMarkerAcct := authtypes.NewBaseAccount(types.MustGetMarkerAddress(rMarkerDenom), nil, 0, 0)
	app.MarkerKeeper.SetNewMarker(ctx, types.NewMarkerAccount(rMarkerAcct, sdk.NewInt64Coin(rMarkerDenom, 1000), transferAuthUser, []types.AccessGrant{{Address: transferAuthUser.String(), Permissions: []types.Access{types.Access_Transfer}}}, types.StatusFinalized, types.MarkerType_RestrictedCoin, true, true, false, reqAttr))

	testCases := []struct {
		name             string
		updateMsgRequest types.MsgUpdateRequiredAttributesRequest
		expectedReqAttr  []string
		expectedError    string
	}{
		{
			name:             "should fail, cannot find marker",
			updateMsgRequest: *types.NewMsgUpdateRequiredAttributesRequest("blah", transferAuthUser, []string{}, []string{}),
			expectedError:    "marker not found for blah: marker blah not found for address: cosmos1psw3a97ywtr595qa4295lw07cz9665hynnfpee",
		},
		{
			name:             "should fail, marker is not restricted",
			updateMsgRequest: *types.NewMsgUpdateRequiredAttributesRequest(notRestrictedMarker.Denom, transferAuthUser, []string{}, []string{}),
			expectedError:    "marker not-restricted-marker is not a restricted marker",
		},
		{
			name:             "should fail, no transfer authority ",
			updateMsgRequest: *types.NewMsgUpdateRequiredAttributesRequest(rMarkerDenom, notTransferAuthUser, []string{}, []string{}),
			expectedError:    fmt.Sprintf("caller does not have authority to update required attributes %s", notTransferAuthUser.String()),
		},
		{
			name:             "should succeed, has gov transfer authority",
			updateMsgRequest: *types.NewMsgUpdateRequiredAttributesRequest(rMarkerDenom, authority, []string{}, []string{}),
			expectedReqAttr:  reqAttr,
		},
		{
			name:             "should succeed, user has transfer auth",
			updateMsgRequest: *types.NewMsgUpdateRequiredAttributesRequest(rMarkerDenom, transferAuthUser, []string{}, []string{}),
			expectedReqAttr:  reqAttr,
		},
		{
			name:             "should fail, can not normalize remove list entry",
			updateMsgRequest: *types.NewMsgUpdateRequiredAttributesRequest(rMarkerDenom, transferAuthUser, []string{"?$#"}, []string{}),
			expectedError:    "value provided for name is invalid",
		},
		{
			name:             "should fail, can not normalize add list entry",
			updateMsgRequest: *types.NewMsgUpdateRequiredAttributesRequest(rMarkerDenom, transferAuthUser, []string{}, []string{"?$#"}),
			expectedError:    "value provided for name is invalid",
		},
		{
			name:             "should fail, remove value does not exist",
			updateMsgRequest: *types.NewMsgUpdateRequiredAttributesRequest(rMarkerDenom, transferAuthUser, []string{"dne.provenance.io"}, []string{}),
			expectedError:    `attribute "dne.provenance.io" is already not required`,
		},
		{
			name:             "should fail, cannot add duplicate entries",
			updateMsgRequest: *types.NewMsgUpdateRequiredAttributesRequest(rMarkerDenom, transferAuthUser, []string{}, []string{"foo.provenance.io"}),
			expectedError:    `attribute "foo.provenance.io" is already required`,
		},
		{
			name:             "should succeed, to remove element",
			updateMsgRequest: *types.NewMsgUpdateRequiredAttributesRequest(rMarkerDenom, transferAuthUser, []string{"foo.provenance.io"}, []string{}),
			expectedReqAttr:  []string{"*.provenance.io", "bar.provenance.io"},
		},
		{
			name:             "should succeed, to add elements",
			updateMsgRequest: *types.NewMsgUpdateRequiredAttributesRequest(rMarkerDenom, transferAuthUser, []string{}, []string{"foo2.provenance.io", "*.jackthecat.io"}),
			expectedReqAttr:  []string{"*.provenance.io", "bar.provenance.io", "foo2.provenance.io", "*.jackthecat.io"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := server.UpdateRequiredAttributes(ctx,
				&tc.updateMsgRequest)

			if len(tc.expectedError) > 0 {
				assert.Nil(t, res)
				assert.EqualError(t, err, tc.expectedError)

			} else {
				assert.NoError(t, err)
				assert.Equal(t, res, &types.MsgUpdateRequiredAttributesResponse{})
				actualMarker, err := app.MarkerKeeper.GetMarkerByDenom(ctx, tc.updateMsgRequest.Denom)
				require.NoError(t, err)
				assert.ElementsMatch(t, tc.expectedReqAttr, actualMarker.GetRequiredAttributes())
			}
		})
	}
}

func TestGetAuthority(t *testing.T) {
	app := simapp.Setup(t)
	require.Equal(t, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn", app.MarkerKeeper.GetAuthority())
}

func TestClearSendDeny(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)

	type SendDenyPair struct {
		marker   sdk.AccAddress
		sendDeny sdk.AccAddress
	}

	tests := []struct {
		name   string
		pairs  []SendDenyPair
		marker sdk.AccAddress
	}{
		{
			name:   "non existant marker",
			pairs:  []SendDenyPair{},
			marker: sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
		},
		{
			name: "single entry for marker",
			pairs: []SendDenyPair{
				{
					marker:   sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
					sendDeny: sdk.AccAddress("cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"),
				},
			},
			marker: sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
		},
		{
			name: "multiple entry for marker",
			pairs: []SendDenyPair{
				{
					marker:   sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
					sendDeny: sdk.AccAddress("cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"),
				},
				{
					marker:   sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
					sendDeny: sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
				},
			},
			marker: sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
		},
		{
			name: "multiple markers with multiple entries",
			pairs: []SendDenyPair{
				{
					marker:   sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
					sendDeny: sdk.AccAddress("cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"),
				},
				{
					marker:   sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
					sendDeny: sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
				},
				{
					marker:   sdk.AccAddress("cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"),
					sendDeny: sdk.AccAddress("cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"),
				},
				{
					marker:   sdk.AccAddress("cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"),
					sendDeny: sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
				},
			},
			marker: sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, pair := range tc.pairs {
				app.MarkerKeeper.AddSendDeny(ctx, pair.marker, pair.sendDeny)
			}

			app.MarkerKeeper.ClearSendDeny(ctx, tc.marker)
			list := app.MarkerKeeper.GetSendDenyList(ctx, tc.marker)
			assert.Empty(t, list, "should remove all entries from send deny list")

			for _, pair := range tc.pairs {
				app.MarkerKeeper.RemoveSendDeny(ctx, pair.marker, pair.sendDeny)
			}
		})
	}
}

func TestGetSendDenyList(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)

	type SendDenyPair struct {
		marker   sdk.AccAddress
		sendDeny sdk.AccAddress
	}

	tests := []struct {
		name     string
		pairs    []SendDenyPair
		marker   sdk.AccAddress
		expected []sdk.AccAddress
	}{
		{
			name:     "non existant marker",
			pairs:    []SendDenyPair{},
			marker:   sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
			expected: []sdk.AccAddress{},
		},
		{
			name: "single entry for marker",
			pairs: []SendDenyPair{
				{
					marker:   sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
					sendDeny: sdk.AccAddress("cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"),
				},
			},
			marker: sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
			expected: []sdk.AccAddress{
				sdk.AccAddress("cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"),
			},
		},
		{
			name: "multiple entry for marker",
			pairs: []SendDenyPair{
				{
					marker:   sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
					sendDeny: sdk.AccAddress("cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"),
				},
				{
					marker:   sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
					sendDeny: sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
				},
			},
			marker: sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
			expected: []sdk.AccAddress{
				sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
				sdk.AccAddress("cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"),
			},
		},
		{
			name: "multiple markers with multiple entries",
			pairs: []SendDenyPair{
				{
					marker:   sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
					sendDeny: sdk.AccAddress("cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"),
				},
				{
					marker:   sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
					sendDeny: sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
				},
				{
					marker:   sdk.AccAddress("cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"),
					sendDeny: sdk.AccAddress("cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"),
				},
				{
					marker:   sdk.AccAddress("cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"),
					sendDeny: sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
				},
			},
			marker: sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
			expected: []sdk.AccAddress{
				sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
				sdk.AccAddress("cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, pair := range tc.pairs {
				app.MarkerKeeper.AddSendDeny(ctx, pair.marker, pair.sendDeny)
			}

			list := app.MarkerKeeper.GetSendDenyList(ctx, tc.marker)
			assert.Equal(t, tc.expected, list, "should return the correct send deny entries for a marker")

			for _, pair := range tc.pairs {
				app.MarkerKeeper.RemoveSendDeny(ctx, pair.marker, pair.sendDeny)
			}
		})
	}
}

func TestAddRemoveSendDeny(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)

	type SendDenyPair struct {
		marker   sdk.AccAddress
		sendDeny sdk.AccAddress
	}

	tests := []struct {
		name  string
		pairs []SendDenyPair
	}{
		{
			name: "new marker and send deny pair",
			pairs: []SendDenyPair{
				{
					marker:   sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
					sendDeny: sdk.AccAddress("cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"),
				},
			},
		},
		{
			name: "duplicate marker and send deny pair",
			pairs: []SendDenyPair{
				{
					marker:   sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
					sendDeny: sdk.AccAddress("cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"),
				},
				{
					marker:   sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
					sendDeny: sdk.AccAddress("cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"),
				},
			},
		},
		{
			name: "multiple senders for marker",
			pairs: []SendDenyPair{
				{
					marker:   sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
					sendDeny: sdk.AccAddress("cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"),
				},
				{
					marker:   sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
					sendDeny: sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
				},
			},
		},
		{
			name: "multiple markers",
			pairs: []SendDenyPair{
				{
					marker:   sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
					sendDeny: sdk.AccAddress("cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"),
				},
				{
					marker:   sdk.AccAddress("cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"),
					sendDeny: sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, pair := range tc.pairs {
				app.MarkerKeeper.AddSendDeny(ctx, pair.marker, pair.sendDeny)
			}

			for _, pair := range tc.pairs {
				isSendDeny := app.MarkerKeeper.IsSendDeny(ctx, pair.marker, pair.sendDeny)
				require.True(t, isSendDeny, "should have entry for added pair")
			}

			for _, pair := range tc.pairs {
				app.MarkerKeeper.RemoveSendDeny(ctx, pair.marker, pair.sendDeny)
			}

			for _, pair := range tc.pairs {
				isSendDeny := app.MarkerKeeper.IsSendDeny(ctx, pair.marker, pair.sendDeny)
				require.False(t, isSendDeny, "should not have entry for removed pair")
			}
		})
	}
}

func TestAddSetNetAssetValues(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.NewContext(false)

	admin := sdk.AccAddress("admin_______________")

	markerAddr := func(denom string) sdk.AccAddress {
		rv, err := types.MarkerAddress(denom)
		require.NoError(t, err, "MarkerAddress(%q)", denom)
		return rv
	}
	makeMarker := func(denom string) types.MarkerAccountI {
		markerAcc := &types.MarkerAccount{
			BaseAccount:            authtypes.NewBaseAccountWithAddress(markerAddr(denom)),
			Manager:                admin.String(),
			Status:                 types.StatusProposed,
			Denom:                  denom,
			Supply:                 sdkmath.NewInt(1000),
			MarkerType:             types.MarkerType_RestrictedCoin,
			SupplyFixed:            true,
			AllowGovernanceControl: true,
		}
		testFunc := func() error {
			return app.MarkerKeeper.AddMarkerAccount(ctx, markerAcc)
		}
		assertions.RequireNotPanicsNoError(t, testFunc, "AddMarkerAccount %q", denom)
		return markerAcc
	}
	coin := func(str string) sdk.Coin {
		rv, err := sdk.ParseCoinNormalized(str)
		require.NoError(t, err, "ParseCoinsNormalized(%q)", str)
		return rv
	}
	navEvent := func(denom string, price string, volume uint64, source string) sdk.Event {
		tev := types.NewEventSetNetAssetValue(denom, coin(price), volume, source)
		rv, err := sdk.TypedEventToEvent(tev)
		require.NoError(t, err, "TypedEventToEvent %q, %s, %d %q", denom, price, volume, source)
		return rv
	}
	newNav := func(price string, volume uint64) types.NetAssetValue {
		return types.NetAssetValue{Price: coin(price), Volume: volume}
	}

	blueMarker := makeMarker("blue")
	redMarker := makeMarker("red")
	yellowMarker := makeMarker("yellow")
	whiteMarker := makeMarker("white")

	tests := []struct {
		name      string
		marker    types.MarkerAccountI
		navs      []types.NetAssetValue
		source    string
		expErr    string
		expEvents sdk.Events
		expNavs   []types.NetAssetValue
	}{
		{
			name:   "nil navs",
			marker: blueMarker,
			navs:   nil,
			source: "billie",
		},
		{
			name:   "empty navs",
			marker: yellowMarker,
			navs:   nil,
			source: "billie",
		},
		{
			name:   "price denom equals marker denom",
			marker: blueMarker,
			navs:   []types.NetAssetValue{newNav("3blue", 3)},
			source: "devin",
			expErr: "net asset value denom cannot match marker denom \"blue\"",
		},
		{
			name:   "price marker does not exist: invalid nav",
			marker: redMarker,
			navs:   []types.NetAssetValue{newNav("4purple", 0)},
			source: "jesse",
			expErr: "net asset value denom does not exist: marker purple not found for address: " + markerAddr("purple").String(),
		},
		{
			name:      "price marker does not exist: valid nav",
			marker:    blueMarker,
			navs:      []types.NetAssetValue{newNav("4purple", 1)},
			source:    "lennon",
			expErr:    "net asset value denom does not exist: marker purple not found for address: " + markerAddr("purple").String(),
			expEvents: sdk.Events{navEvent("blue", "4purple", 1, "lennon")},
		},
		{
			name:   "price marker exists: invalid nav",
			marker: yellowMarker,
			navs:   []types.NetAssetValue{newNav("4red", 0)},
			source: "remy",
			expErr: "cannot set net asset value: marker net asset value volume must be positive value",
		},
		{
			name:      "volume greater than supply",
			marker:    redMarker,
			navs:      []types.NetAssetValue{newNav("3blue", 1001)},
			source:    "val",
			expEvents: sdk.Events{navEvent("red", "3blue", 1001, "val")},
			expNavs:   []types.NetAssetValue{newNav("3blue", 1001)},
		},
		{
			name:      "one nav: success",
			marker:    yellowMarker,
			navs:      []types.NetAssetValue{newNav("3blue", 17)},
			source:    "harper",
			expEvents: sdk.Events{navEvent("yellow", "3blue", 17, "harper")},
			expNavs:   []types.NetAssetValue{newNav("3blue", 17)},
		},
		{
			name:   "usd nav: zero volume",
			marker: blueMarker,
			navs:   []types.NetAssetValue{newNav("5"+types.UsdDenom, 0)},
			source: "tony",
			expErr: "cannot set net asset value: marker net asset value volume must be positive value",
		},
		{
			name:      "usd nav: volume greater than supply",
			marker:    blueMarker,
			navs:      []types.NetAssetValue{newNav("55"+types.UsdDenom, 1005)},
			source:    "wynne",
			expEvents: sdk.Events{navEvent("blue", "55"+types.UsdDenom, 1005, "wynne")},
			expNavs:   []types.NetAssetValue{newNav("55"+types.UsdDenom, 1005)},
		},
		{
			name:      "usd nav: success",
			marker:    blueMarker,
			navs:      []types.NetAssetValue{newNav("55"+types.UsdDenom, 1000)},
			source:    "cody",
			expEvents: sdk.Events{navEvent("blue", "55"+types.UsdDenom, 1000, "cody")},
			expNavs:   []types.NetAssetValue{newNav("55"+types.UsdDenom, 1000)},
		},
		{
			name:   "three navs: no errors",
			marker: whiteMarker,
			navs:   []types.NetAssetValue{newNav("7blue", 2), newNav("15red", 66), newNav("400yellow", 89)},
			source: "jordan",
			expEvents: sdk.Events{
				navEvent("white", "7blue", 2, "jordan"),
				navEvent("white", "15red", 66, "jordan"),
				navEvent("white", "400yellow", 89, "jordan"),
			},
			expNavs: []types.NetAssetValue{newNav("7blue", 2), newNav("15red", 66), newNav("400yellow", 89)},
		},
		{
			name:   "three navs: error on first",
			marker: whiteMarker,
			navs:   []types.NetAssetValue{newNav("7blue", 0), newNav("167red", 66), newNav("377yellow", 89)},
			source: "knox",
			expErr: "cannot set net asset value: marker net asset value volume must be positive value",
			expEvents: sdk.Events{
				// no blue event because the nav is invalid.
				navEvent("white", "167red", 66, "knox"),
				navEvent("white", "377yellow", 89, "knox"),
			},
			expNavs: []types.NetAssetValue{newNav("167red", 66), newNav("377yellow", 89)},
		},
		{
			name:   "three navs: error on second",
			marker: whiteMarker,
			navs:   []types.NetAssetValue{newNav("14blue", 2), newNav("15red", 0), newNav("403yellow", 89)},
			source: "max",
			expErr: "cannot set net asset value: marker net asset value volume must be positive value",
			expEvents: sdk.Events{
				navEvent("white", "14blue", 2, "max"),
				// no red event because the nav is invalid.
				navEvent("white", "403yellow", 89, "max"),
			},
			expNavs: []types.NetAssetValue{newNav("14blue", 2), newNav("403yellow", 89)},
		},
		{
			name:   "three navs: error on third",
			marker: whiteMarker,
			navs:   []types.NetAssetValue{newNav("788blue", 14), newNav("215red", 3), newNav("470white", 14)},
			source: "palmer",
			expErr: "net asset value denom cannot match marker denom \"white\"",
			expEvents: sdk.Events{
				navEvent("white", "788blue", 14, "palmer"),
				navEvent("white", "215red", 3, "palmer"),
				// no white event because it's the same denom as the marker.
			},
			expNavs: []types.NetAssetValue{newNav("788blue", 14), newNav("215red", 3)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			em := sdk.NewEventManager()
			ctx = ctx.WithEventManager(em)
			expHeight := ctx.BlockHeight()
			var err error
			testFunc := func() {
				err = app.MarkerKeeper.AddSetNetAssetValues(ctx, tc.marker, tc.navs, tc.source)
			}

			require.NotPanics(t, testFunc, "AddSetNetAssetValues")
			assertions.AssertErrorValue(t, err, tc.expErr, "AddSetNetAssetValues error")
			actEvents := em.Events()
			assertions.AssertEqualEvents(t, tc.expEvents, actEvents, "events emitted during AddSetNetAssetValues")

			for i, expNav := range tc.expNavs {
				actNav, navErr := app.MarkerKeeper.GetNetAssetValue(ctx, tc.marker.GetDenom(), expNav.Price.Denom)
				if assert.NoError(t, navErr, "[%d]: GetNetAssetValue(%q, %q)", i, tc.marker.GetDenom(), expNav.Price.Denom) {
					assert.Equal(t, expNav.Price.String(), actNav.Price.String(),
						"[%d]: %s:%s nav Price", i, tc.marker.GetDenom(), expNav.Price.Denom)
					assert.Equal(t, fmt.Sprintf("%d", expNav.Volume), fmt.Sprintf("%d", actNav.Volume),
						"[%d]: %s:%s nav Volume", i, tc.marker.GetDenom(), expNav.Price.Denom)
					assert.Equal(t, fmt.Sprintf("%d", expHeight), fmt.Sprintf("%d", actNav.UpdatedBlockHeight),
						"[%d]: %s:%s nav UpdatedBlockHeight", i, tc.marker.GetDenom(), expNav.Price.Denom)
				}
			}
		})
	}
}

func TestGetNetAssetValue(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.NewContext(false)

	admin := sdk.AccAddress("admin_account_______")
	makeMarker := func(denom string, navs ...types.NetAssetValue) types.MarkerAccountI {
		markerAddr := types.MustGetMarkerAddress(denom)
		markerAcc := types.NewMarkerAccount(
			authtypes.NewBaseAccount(markerAddr, nil, 0, 0),
			sdk.NewInt64Coin(denom, 1_000_000_000),
			admin,
			[]types.AccessGrant{{
				Address: admin.String(),
				Permissions: []types.Access{
					types.Access_Transfer,
					types.Access_Mint, types.Access_Burn, types.Access_Deposit,
					types.Access_Withdraw, types.Access_Delete, types.Access_Admin,
				},
			}},
			types.StatusProposed,
			types.MarkerType_RestrictedCoin,
			true,
			true,
			true,
			[]string{},
		)

		require.NoError(t, app.MarkerKeeper.AddSetNetAssetValues(ctx, markerAcc, navs, "initial"), "AddSetNetAssetValues %s", denom)
		require.NoError(t, app.MarkerKeeper.AddFinalizeAndActivateMarker(ctx, markerAcc), "AddFinalizeAndActivateMarker %s", denom)
		return markerAcc
	}

	cherryUsdNav := types.NewNetAssetValue(sdk.NewInt64Coin(types.UsdDenom, 25), 1)
	cherryAcc := makeMarker("cherry", cherryUsdNav)

	appleUsdNav := types.NewNetAssetValue(sdk.NewInt64Coin(types.UsdDenom, 50_000), 1_000_000)
	appleCherryNav := types.NewNetAssetValue(sdk.NewInt64Coin("cherry", 57), 7777)
	appleAcc := makeMarker("apple", appleUsdNav, appleCherryNav)

	require.NoError(t, app.MarkerKeeper.SetNetAssetValue(ctx, appleAcc, appleCherryNav, "test setup"), "AddSetNetAssetValues apple cherry")

	// Put a bad cherry -> durian entry in state.
	app.MarkerKeeper.GetStore(ctx).Set(types.NetAssetValueKey(cherryAcc.GetAddress(), "durian"), []byte{255, 255})

	tests := []struct {
		name        string
		markerDenom string
		priceDenom  string
		expNav      *types.NetAssetValue
		expErr      string
	}{
		{
			name:        "invalid marker denom",
			markerDenom: "x",
			priceDenom:  types.UsdDenom,
			expNav:      nil,
			expErr:      "could not get marker \"x\" address: invalid denom: x",
		},
		{
			name:        "no entry: cherry apple",
			markerDenom: "cherry",
			priceDenom:  "apple",
			expNav:      nil,
			expErr:      "",
		},
		{
			name:        "bad entry: cherry durian",
			markerDenom: "cherry",
			priceDenom:  "durian",
			expNav:      nil,
			expErr:      "could not read nav for marker \"cherry\" with price denom \"durian\": unexpected EOF",
		},
		{
			name:        "good entry: apple usd",
			markerDenom: "apple",
			priceDenom:  types.UsdDenom,
			expNav:      &appleUsdNav,
		},
		{
			name:        "good entry: apple cherry",
			markerDenom: "apple",
			priceDenom:  "cherry",
			expNav:      &appleCherryNav,
		},
		{
			name:        "good entry: cherry usd",
			markerDenom: "cherry",
			priceDenom:  types.UsdDenom,
			expNav:      &cherryUsdNav,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var nav *types.NetAssetValue
			var err error
			testFunc := func() {
				nav, err = app.MarkerKeeper.GetNetAssetValue(ctx, tc.markerDenom, tc.priceDenom)
			}
			require.NotPanics(t, testFunc, "GetNetAssetValue(%q, %q)", tc.markerDenom, tc.priceDenom)
			assertions.AssertErrorValue(t, err, tc.expErr, "GetNetAssetValue(%q, %q) error", tc.markerDenom, tc.priceDenom)
			if tc.expNav == nil {
				assert.Nil(t, nav, "GetNetAssetValue(%q, %q) nav", tc.markerDenom, tc.priceDenom)
			} else if assert.NotNil(t, nav, "GetNetAssetValue(%q, %q) nav", tc.markerDenom, tc.priceDenom) {
				assert.Equal(t, tc.expNav.Price.String(), nav.Price.String(), "nav price")
				assert.Equal(t, tc.expNav.Volume, nav.Volume, "nav volume")
			}
		})
	}
}

func TestIterateAllNetAssetValues(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)

	tests := []struct {
		name       string
		markerNavs []sdk.Coins
		expected   int
	}{
		{
			name:       "should work with no markers",
			markerNavs: []sdk.Coins{},
			expected:   0,
		},
		{
			name: "should work with one marker no usd denom",
			markerNavs: []sdk.Coins{
				sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 1)),
			},
			expected: 1,
		},
		{
			name: "should work with multiple markers no usd denom",
			markerNavs: []sdk.Coins{
				sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 1)),
				sdk.NewCoins(sdk.NewInt64Coin("georgethedog", 2)),
			},
			expected: 2,
		},
		{
			name: "should work with one marker with usd denom",
			markerNavs: []sdk.Coins{
				sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 1), sdk.NewInt64Coin(types.UsdDenom, 2)),
			},
			expected: 2,
		},
		{
			name: "should work with multiple markers with usd denom",
			markerNavs: []sdk.Coins{
				sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 1), sdk.NewInt64Coin(types.UsdDenom, 3)),
				sdk.NewCoins(sdk.NewInt64Coin("georgethedog", 2), sdk.NewInt64Coin(types.UsdDenom, 4)),
			},
			expected: 4,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create the marker
			for i, prices := range tc.markerNavs {
				address := sdk.AccAddress(fmt.Sprintf("marker%d", i))
				marker := types.NewEmptyMarkerAccount(fmt.Sprintf("coin%d", i), address.String(), []types.AccessGrant{})
				marker.Supply = sdkmath.OneInt()
				require.NoError(t, app.MarkerKeeper.AddMarkerAccount(ctx, marker), "AddMarkerAccount() error")

				var navs []types.NetAssetValue
				for _, price := range prices {
					navs = append(navs, types.NewNetAssetValue(price, uint64(1)))
					navAddr := sdk.AccAddress(price.Denom)
					if acc, _ := app.MarkerKeeper.GetMarkerByDenom(ctx, price.Denom); acc == nil {
						navMarker := types.NewEmptyMarkerAccount(price.Denom, navAddr.String(), []types.AccessGrant{})
						navMarker.Supply = sdkmath.OneInt()
						require.NoError(t, app.MarkerKeeper.AddMarkerAccount(ctx, navMarker), "AddMarkerAccount() error")
					}
				}
				require.NoError(t, app.MarkerKeeper.AddSetNetAssetValues(ctx, marker, navs, "AddSetNetAssetValues() error"))
			}

			// Test Logic
			count := 0
			app.MarkerKeeper.IterateAllNetAssetValues(ctx, func(aa sdk.AccAddress, nav types.NetAssetValue) (stop bool) {
				count += 1
				return false
			})
			assert.Equal(t, tc.expected, count, "should iterate the correct number of times")

			// Destroy the marker
			for i, prices := range tc.markerNavs {
				coin := fmt.Sprintf("coin%d", i)
				marker, err := app.MarkerKeeper.GetMarkerByDenom(ctx, coin)
				require.NoError(t, err, "GetMarkerByDenom() error")
				app.MarkerKeeper.RemoveMarker(ctx, marker)

				// We need to remove the nav markers
				for _, price := range prices {
					if navMarker, _ := app.MarkerKeeper.GetMarkerByDenom(ctx, price.Denom); navMarker != nil {
						app.MarkerKeeper.RemoveMarker(ctx, navMarker)
					}
				}
			}
		})
	}
}

func TestReqAttrBypassAddrs(t *testing.T) {
	// Tests both GetReqAttrBypassAddrs and IsReqAttrBypassAddr.
	expectedNames := []string{
		authtypes.FeeCollectorName,
		quarantine.ModuleName,
		govtypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.BondedPoolName,
		stakingtypes.NotBondedPoolName,
	}

	incByte := func(b byte) byte {
		if b == 0xFF {
			return 0x00
		}
		return b + 1
	}

	app := simapp.Setup(t)

	for _, name := range expectedNames {
		t.Run(fmt.Sprintf("get: contains %s", name), func(t *testing.T) {
			expAddr := authtypes.NewModuleAddress(name)
			actual := app.MarkerKeeper.GetReqAttrBypassAddrs()
			assert.Contains(t, actual, expAddr, "GetReqAttrBypassAddrs()")
		})
	}
	t.Run("get: only has expected entries", func(t *testing.T) {
		// This assumes each expectedNames test passed. This is designed to fail if a new entry
		// is added to the list (in app.go). When that happens, update the expectedNames with
		// the new entry so it's harder for it to accidentally go missing.
		actual := app.MarkerKeeper.GetReqAttrBypassAddrs()
		assert.Len(t, actual, len(expectedNames), "GetReqAttrBypassAddrs()")
	})

	t.Run("get: called twice equal but not same", func(t *testing.T) {
		expected := app.MarkerKeeper.GetReqAttrBypassAddrs()
		actual := app.MarkerKeeper.GetReqAttrBypassAddrs()
		if assert.Equal(t, expected, actual, "GetReqAttrBypassAddrs()") {
			if assert.NotSame(t, &expected, &actual, "GetReqAttrBypassAddrs()") {
				for i := range expected {
					assert.NotSame(t, &expected[i], &actual[i], "GetReqAttrBypassAddrs()[%d]", i)
				}
			}
		}
	})

	t.Run("get: changes to result not reflected in next result", func(t *testing.T) {
		actual1 := app.MarkerKeeper.GetReqAttrBypassAddrs()
		origActual100 := actual1[0][0]
		actual1[0][0] = incByte(origActual100)
		actual2 := app.MarkerKeeper.GetReqAttrBypassAddrs()
		actual200 := actual2[0][0]
		assert.Equal(t, origActual100, actual200, "first byte of first address after changing it in an earlier result")
	})

	for _, name := range expectedNames {
		t.Run(fmt.Sprintf("is: %s", name), func(t *testing.T) {
			addr := authtypes.NewModuleAddress(name)
			actual := app.MarkerKeeper.IsReqAttrBypassAddr(addr)
			assert.True(t, actual, "IsReqAttrBypassAddr(NewModuleAddress(%q))", name)
		})
	}

	almostName0 := authtypes.NewModuleAddress(expectedNames[0])
	almostName0[0] = incByte(almostName0[0])

	negativeIsTests := []struct {
		name string
		addr sdk.AccAddress
	}{
		{name: "nil address", addr: nil},
		{name: "empty address", addr: sdk.AccAddress{}},
		{name: "zerod address", addr: make(sdk.AccAddress, 20)},
		{name: "almost " + expectedNames[0], addr: almostName0},
		{name: "short " + expectedNames[0], addr: authtypes.NewModuleAddress(expectedNames[0])[:19]},
		{name: "long " + expectedNames[0], addr: append(authtypes.NewModuleAddress(expectedNames[0]), 0x00)},
	}

	for _, tc := range negativeIsTests {
		t.Run(tc.name, func(t *testing.T) {
			actual := app.MarkerKeeper.IsReqAttrBypassAddr(tc.addr)
			assert.False(t, actual, "IsReqAttrBypassAddr(...)")
		})
	}
}

func TestIsMarkerAccount(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)

	newMarker := func(denom string, status types.MarkerStatus, typ types.MarkerType) sdk.AccAddress {
		addr, err := types.MarkerAddress(denom)
		require.NoError(t, err, "MarkerAddress(%q)", denom)
		marker := &types.MarkerAccount{
			BaseAccount: &authtypes.BaseAccount{Address: addr.String()},
			AccessControl: []types.AccessGrant{{
				Address:     sdk.AccAddress("addr_with_perms_____").String(),
				Permissions: types.AccessList{types.Access_Admin},
			}},
			Status:                 status,
			Denom:                  denom,
			Supply:                 sdkmath.NewInt(1000),
			MarkerType:             typ,
			SupplyFixed:            true,
			AllowGovernanceControl: true,
		}

		require.NotPanics(t, func() {
			app.MarkerKeeper.SetNewMarker(ctx, marker)
		}, "SetNewMarker %q", marker.Denom)
		return addr
	}

	normalAddr := sdk.AccAddress("normal_address______")
	setNewAccount(app, ctx, &authtypes.BaseAccount{Address: normalAddr.String()})

	tests := []struct {
		name string
		addr sdk.AccAddress
		exp  bool
	}{
		{name: "nil address", addr: nil, exp: false},
		{name: "empty address", addr: nil, exp: false},
		{name: "unknown address", addr: sdk.AccAddress("unknown_address_____"), exp: false},
		{name: "normal address", addr: normalAddr, exp: false},
		{
			name: "proposed restricted marker",
			addr: newMarker("proposedrestricted", types.StatusProposed, types.MarkerType_RestrictedCoin),
			exp:  true,
		},
		{
			name: "active restricted marker",
			addr: newMarker("activerestricted", types.StatusActive, types.MarkerType_RestrictedCoin),
			exp:  true,
		},
		{
			name: "proposed coin marker",
			addr: newMarker("proposedcoin", types.StatusProposed, types.MarkerType_Coin),
			exp:  true,
		},
		{
			name: "active coin marker",
			addr: newMarker("activecoin", types.StatusActive, types.MarkerType_Coin),
			exp:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual bool
			testFunc := func() {
				actual = app.MarkerKeeper.IsMarkerAccount(ctx, tc.addr)
			}
			require.NotPanics(t, testFunc, "IsMarkerAccount")
			assert.Equal(t, tc.exp, actual, "result from IsMarkerAccount")
		})
	}
}

// dummyBankKeeper satisfies the types.BankKeeper interface but does nothing.
type dummyBankKeeper struct{}

var _ types.BankKeeper = (*dummyBankKeeper)(nil)

func (d dummyBankKeeper) GetAllBalances(_ context.Context, _ sdk.AccAddress) sdk.Coins { return nil }

func (d dummyBankKeeper) GetBalance(_ context.Context, _ sdk.AccAddress, _ string) sdk.Coin {
	return sdk.Coin{}
}

func (d dummyBankKeeper) GetSupply(_ context.Context, _ string) sdk.Coin { return sdk.Coin{} }

func (d dummyBankKeeper) DenomOwners(_ context.Context, _ *banktypes.QueryDenomOwnersRequest) (*banktypes.QueryDenomOwnersResponse, error) {
	return nil, nil
}

func (d dummyBankKeeper) SendCoins(_ context.Context, _, _ sdk.AccAddress, _ sdk.Coins) error {
	return nil
}

func (d dummyBankKeeper) SendCoinsFromModuleToAccount(_ context.Context, _ string, _ sdk.AccAddress, _ sdk.Coins) error {
	return nil
}

func (d dummyBankKeeper) SendCoinsFromAccountToModule(_ context.Context, _ sdk.AccAddress, _ string, _ sdk.Coins) error {
	return nil
}

func (d dummyBankKeeper) MintCoins(_ context.Context, _ string, _ sdk.Coins) error { return nil }

func (d dummyBankKeeper) BurnCoins(_ context.Context, _ string, _ sdk.Coins) error { return nil }

func (d dummyBankKeeper) AppendSendRestriction(_ banktypes.SendRestrictionFn) {}

func (d dummyBankKeeper) BlockedAddr(_ sdk.AccAddress) bool { return false }

func (d dummyBankKeeper) GetDenomMetaData(_ context.Context, _ string) (banktypes.Metadata, bool) {
	return banktypes.Metadata{}, false
}

func (d dummyBankKeeper) SetDenomMetaData(_ context.Context, _ banktypes.Metadata) {}

func (d dummyBankKeeper) IterateAllBalances(_ context.Context, _ func(sdk.AccAddress, sdk.Coin) bool) {
}

func (d dummyBankKeeper) GetAllSendEnabledEntries(_ context.Context) []banktypes.SendEnabled {
	return nil
}

func TestBypassAddrsLocked(t *testing.T) {
	// This test makes sure that the keeper's copy of reqAttrBypassAddrs
	// isn't changed if the originally provided value is changed.

	addrs := []sdk.AccAddress{
		sdk.AccAddress("addrs[0]____________"),
		sdk.AccAddress("addrs[1]____________"),
		sdk.AccAddress("addrs[2]____________"),
		sdk.AccAddress("addrs[3]____________"),
		sdk.AccAddress("addrs[4]____________"),
	}

	mk := markerkeeper.NewKeeper(nil, nil, nil, &dummyBankKeeper{}, nil, nil, nil, nil, nil, addrs, nil)

	// Now that the keeper has been created using the provided addresses, change the first byte of
	// the first address to something else. Then, get the addresses back from the keeper and make
	// sure that change didn't affect what's in the keeper.
	orig00 := addrs[0][0]
	addrs[0][0] = 'b'
	kAddrs := mk.GetReqAttrBypassAddrs()
	act00 := kAddrs[0][0]
	assert.Equal(t, orig00, act00, "first byte of first address returned by GetReqAttrBypassAddrs")
}
