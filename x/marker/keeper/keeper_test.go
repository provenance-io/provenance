package keeper_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simapp "github.com/provenance-io/provenance/app"
	markerkeeper "github.com/provenance-io/provenance/x/marker/keeper"
	"github.com/provenance-io/provenance/x/marker/types"
	rewardtypes "github.com/provenance-io/provenance/x/reward/types"
)

func TestAccountMapperGetSet(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

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

	app.AccountKeeper.SetAccount(ctx, mac)

	// check the new values
	acc = app.AccountKeeper.GetAccount(ctx, addr)
	require.NotNil(t, acc)
	mac, ok = acc.(types.MarkerAccountI)
	require.True(t, ok)
	require.True(t, mac.AddressHasAccess(user, types.Access_Admin))

	app.MarkerKeeper.RemoveMarker(ctx, mac)

	// getting account after delete should be nil
	acc = app.AccountKeeper.GetAccount(ctx, addr)
	require.Nil(t, acc)

	require.Empty(t, app.MarkerKeeper.GetAllMarkerHolders(ctx, "testcoin"))

	// check for error on invaid marker denom
	_, err := app.MarkerKeeper.GetMarkerByDenom(ctx, "doesntexist")
	require.Error(t, err, "marker does not exist, should error")
}

func TestExistingAccounts(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	server := markerkeeper.NewMsgServerImpl(app.MarkerKeeper)

	addr := types.MustGetMarkerAddress("testcoin")
	pubkey := secp256k1.GenPrivKey().PubKey()
	user := testUserAddress("testcoin")
	manager := testUserAddress("manager")
	existingBalance := sdk.NewCoin("coin", sdk.NewInt(1000))

	// prefund the marker address so an account gets created before the marker does.
	app.AccountKeeper.SetAccount(ctx, authtypes.NewBaseAccount(user, pubkey, 0, 0))
	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, addr, sdk.NewCoins(existingBalance)), "funding account")
	require.Equal(t, existingBalance, app.BankKeeper.GetBalance(ctx, addr, "coin"), "account balance must be set")

	// Creating a marker over an account with zero sequence is fine.
	_, err := server.AddMarker(sdk.WrapSDKContext(ctx), types.NewMsgAddMarkerRequest("testcoin", sdk.NewInt(30), user, manager, types.MarkerType_Coin, true, true, false, []string{}, 0, 0))
	require.NoError(t, err, "should allow a marker over existing account that has not signed anything.")

	// existing coin balance must still be present
	require.Equal(t, existingBalance, app.BankKeeper.GetBalance(ctx, addr, "coin"), "account balances must be preserved")

	// Creating a marker over an existing marker fails.
	_, err = server.AddMarker(sdk.WrapSDKContext(ctx), types.NewMsgAddMarkerRequest("testcoin", sdk.NewInt(30), user, manager, types.MarkerType_Coin, true, true, false, []string{}, 0, 0))
	require.Error(t, err, "fails because marker already exists")

	// replace existing test account with a new copy that has a positive sequence number
	app.AccountKeeper.SetAccount(ctx, authtypes.NewBaseAccount(user, pubkey, 0, 10))

	// Creating a marker over an existing account with a positive sequence number fails.
	_, err = server.AddMarker(sdk.WrapSDKContext(ctx), types.NewMsgAddMarkerRequest("testcoin", sdk.NewInt(30), user, manager, types.MarkerType_Coin, true, true, false, []string{}, 0, 0))
	require.Error(t, err, "should not allow creation over and existing account with a positive sequence number.")
}

func TestAccountUnrestrictedDenoms(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	server := markerkeeper.NewMsgServerImpl(app.MarkerKeeper)

	user := testUserAddress("test")

	// Require a long unrestricted denom
	app.MarkerKeeper.SetParams(ctx, types.Params{UnrestrictedDenomRegex: "[a-z]{12,20}"})
	_, err := server.AddMarker(sdk.WrapSDKContext(ctx), types.NewMsgAddMarkerRequest("tooshort", sdk.NewInt(30), user, user, types.MarkerType_Coin, true, true, false, []string{}, 0, 0))
	require.Error(t, err, "fails with unrestricted denom length fault")
	require.Equal(t, fmt.Errorf("invalid denom [tooshort] (fails unrestricted marker denom validation [a-z]{12,20})"), err, "should fail with denom restriction")

	_, err = server.AddMarker(sdk.WrapSDKContext(ctx), types.NewMsgAddMarkerRequest("itslongenough", sdk.NewInt(30), user, user, types.MarkerType_Coin, true, true, false, []string{}, 0, 0))
	require.NoError(t, err, "should allow a marker with a sufficiently long denom")

	// Set to an empty string (returns to default expression)
	app.MarkerKeeper.SetParams(ctx, types.Params{UnrestrictedDenomRegex: ""})
	_, err = server.AddMarker(sdk.WrapSDKContext(ctx), types.NewMsgAddMarkerRequest("short", sdk.NewInt(30), user, user, types.MarkerType_Coin, true, true, false, []string{}, 0, 0))
	// succeeds now as the default unrestricted denom expression allows any valid denom (minimum length is 2)
	require.NoError(t, err, "should allow any valid denom with a min length of two")
}

func TestAccountKeeperReader(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

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

func TestAccountKeeperManageAccess(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

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
	require.NoError(t, mac.SetSupply(sdk.NewCoin(mac.Denom, sdk.OneInt())))
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
	require.NoError(t, app.MarkerKeeper.MintCoin(ctx, user2, sdk.NewCoin("testcoin", sdk.OneInt())))
	require.NoError(t, app.MarkerKeeper.BurnCoin(ctx, user1, sdk.NewCoin("testcoin", sdk.OneInt())))

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

func TestAccountKeeperCancelProposedByManager(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

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
	require.NoError(t, mac.SetSupply(sdk.NewCoin(mac.Denom, sdk.OneInt())))
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

func TestAccountKeeperMintBurnCoins(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
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
	require.NoError(t, mac.SetSupply(sdk.NewCoin("testcoin", sdk.NewInt(1000))))

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
	require.EqualValues(t, app.MarkerKeeper.GetEscrow(ctx, m).AmountOf("testcoin"), sdk.NewInt(1000))

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
	require.EqualValues(t, app.BankKeeper.GetBalance(ctx, user, "testcoin").Amount, sdk.NewInt(50))

	// verify marker account has remaining coins
	m, err = app.MarkerKeeper.GetMarker(ctx, addr)
	require.NoError(t, err)
	require.EqualValues(t, app.MarkerKeeper.GetEscrow(ctx, m).AmountOf("testcoin"), sdk.NewInt(950))

	// Fail for burn too much (exceed marker account holdings)
	require.Error(t, app.MarkerKeeper.BurnCoin(ctx, user, sdk.NewInt64Coin("testcoin", 1000)))
	// Fails because a user is holding some of the supply
	require.Error(t, app.MarkerKeeper.CancelMarker(ctx, user, "testcoin"))

	// two a user and the marker
	require.Equal(t, 2, len(app.MarkerKeeper.GetAllMarkerHolders(ctx, "testcoin")))

	// put the coins back in the types.
	require.NoError(t, app.BankKeeper.SendCoins(ctx, user, addr, sdk.NewCoins(sdk.NewInt64Coin("testcoin", 50))))

	// one, only the marker
	require.Equal(t, 1, len(app.MarkerKeeper.GetAllMarkerHolders(ctx, "testcoin")))

	// succeeds because marker has all its supply
	require.NoError(t, app.MarkerKeeper.CancelMarker(ctx, user, "testcoin"))

	// verify status is cancelled
	m, err = app.MarkerKeeper.GetMarker(ctx, addr)
	require.NoError(t, err)
	require.EqualValues(t, types.StatusCancelled, m.GetStatus())

	// succeeds on a cancelled marker (no-op)
	require.NoError(t, app.MarkerKeeper.CancelMarker(ctx, user, "testcoin"))

	// Set an escrow balance
	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, addr, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.OneInt()))), "funding account")
	// Fails because there are coins in escrow.
	require.Error(t, app.MarkerKeeper.DeleteMarker(ctx, user, "testcoin"))

	// Remove escrow balance from account
	require.NoError(t, app.BankKeeper.SendCoinsFromAccountToModule(ctx, addr, "mint", sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.OneInt()))), "sending coins to module")

	// Succeeds because the bond denom coin was removed.
	require.NoError(t, app.MarkerKeeper.DeleteMarker(ctx, user, "testcoin"))

	// none, marker has been deleted
	require.Equal(t, 0, len(app.MarkerKeeper.GetAllMarkerHolders(ctx, "testcoin")))

	// verify status is destroyed and supply is zero.
	m, err = app.MarkerKeeper.GetMarker(ctx, addr)
	require.NoError(t, err)
	require.EqualValues(t, types.StatusDestroyed, m.GetStatus())
	require.EqualValues(t, m.GetSupply().Amount, sdk.ZeroInt())

	// supply module should also indicate a zero supply
	require.EqualValues(t, app.BankKeeper.GetSupply(ctx, "testcoin").Amount, sdk.ZeroInt())
}

func TestAccountKeeperGetAll(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

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

func TestAccountInsufficientExisting(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	pubkey := secp256k1.GenPrivKey().PubKey()
	user := sdk.AccAddress(pubkey.Address())

	// setup an existing account with an existing balance (and matching supply)
	existingSupply := sdk.NewCoin("testcoin", sdk.NewInt(10000))
	app.AccountKeeper.SetAccount(ctx, authtypes.NewBaseAccount(user, pubkey, 0, 0))

	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, user, sdk.NewCoins(existingSupply)), "funding account")

	//prevSupply := app.BankKeeper.GetSupply(ctx, "testcoin")
	//app.BankKeeper.SetSupply(ctx, banktypes.NewSupply(prevSupply.Amount.Add(existingSupply.Amount)))

	// create account and check default values
	mac := types.NewEmptyMarkerAccount("testcoin", user.String(), []types.AccessGrant{*types.NewAccessGrant(user,
		[]types.Access{types.Access_Mint, types.Access_Burn, types.Access_Withdraw, types.Access_Delete})})
	require.NoError(t, mac.SetManager(user))
	require.NoError(t, mac.SetSupply(sdk.NewCoin("testcoin", sdk.NewInt(1000))))

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

func TestAccountImplictControl(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

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
	require.NoError(t, mac.SetSupply(sdk.NewCoin("testcoin", sdk.NewInt(1000))))

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
	require.NoError(t, app.MarkerKeeper.TransferCoin(ctx, user2, user, user2, sdk.NewCoin("testcoin", sdk.NewInt(10))))
	// fails if the admin user does not have transfer authority
	require.Error(t, app.MarkerKeeper.TransferCoin(ctx, user, user2, user, sdk.NewCoin("testcoin", sdk.NewInt(10))))

	// validate authz when 'from' is different from 'admin'
	granter := user
	grantee := user2
	now := ctx.BlockHeader().Time
	require.NotNil(t, now, "now")
	exp1Hour := now.Add(time.Hour)
	a := types.NewMarkerTransferAuthorization(sdk.NewCoins(sdk.NewCoin("testcoin", sdk.NewInt(10))), []sdk.AccAddress{})

	// fails when admin user (grantee without authz permissions) has transfer authority
	require.Error(t, app.MarkerKeeper.TransferCoin(ctx, granter, user, grantee, sdk.NewCoin("testcoin", sdk.NewInt(5))))
	// succeeds when admin user (grantee with authz permissions) has transfer authority
	require.NoError(t, app.AuthzKeeper.SaveGrant(ctx, grantee, granter, a, &exp1Hour))
	require.NoError(t, app.MarkerKeeper.TransferCoin(ctx, granter, user, grantee, sdk.NewCoin("testcoin", sdk.NewInt(5))))
	// succeeds when admin user (grantee with authz permissions) has transfer authority (transfer remaining balance)
	require.NoError(t, app.MarkerKeeper.TransferCoin(ctx, granter, user, grantee, sdk.NewCoin("testcoin", sdk.NewInt(5))))
	// fails when admin user (grantee with authz permissions) and transfer authority has transferred all coin ^^^ (grant has now been deleted)
	require.Error(t, app.MarkerKeeper.TransferCoin(ctx, granter, user, grantee, sdk.NewCoin("testcoin", sdk.NewInt(5))))

	// validate authz when with allow list set
	now = ctx.BlockHeader().Time
	require.NotNil(t, now, "now")
	exp1Hour = now.Add(time.Hour)
	a = types.NewMarkerTransferAuthorization(sdk.NewCoins(sdk.NewCoin("testcoin", sdk.NewInt(10))), []sdk.AccAddress{user})
	require.NoError(t, app.AuthzKeeper.SaveGrant(ctx, grantee, granter, a, &exp1Hour))
	// fails when admin user (grantee with authz permissions) has transfer authority but receiver is not on allowed list
	require.Error(t, app.MarkerKeeper.TransferCoin(ctx, granter, user2, grantee, sdk.NewCoin("testcoin", sdk.NewInt(5))))
	// succeeds when admin user (grantee with authz permissions) has transfer authority with receiver is on allowed list
	require.NoError(t, app.MarkerKeeper.TransferCoin(ctx, granter, user, grantee, sdk.NewCoin("testcoin", sdk.NewInt(5))))
}

func TestForceTransfer(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

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
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	setAcc := func(addr sdk.AccAddress, sequence uint64) {
		acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
		require.NoError(t, acc.SetSequence(sequence), "%s.SetSequence(%d)", string(addr), sequence)
		app.AccountKeeper.SetAccount(ctx, acc)
	}

	addrNoAcc := sdk.AccAddress("addrNoAcc___________")
	addrSeq0 := sdk.AccAddress("addrSeq0____________")
	addrSeq1 := sdk.AccAddress("addrSeq1____________")
	setAcc(addrSeq0, 0)
	setAcc(addrSeq1, 1)

	tests := []struct {
		name string
		from sdk.AccAddress
		exp  bool
	}{
		{name: "address without an account", from: addrNoAcc, exp: true},
		{name: "address with sequence 0", from: addrSeq0, exp: false},
		{name: "address with sequence 1", from: addrSeq1, exp: true},
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
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
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

	app.AccountKeeper.SetAccount(ctx, mac)

	existingSupply := sdk.NewCoin("testcoin", sdk.NewInt(10000))
	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, user, sdk.NewCoins(existingSupply)), "funding accont")

	allowance, err := types.NewMsgGrantAllowance(
		"testcoin",
		user,
		testUserAddress("grantee"),
		&feegrant.BasicAllowance{SpendLimit: sdk.NewCoins(sdk.NewCoin("testcoin", sdk.OneInt()))})
	require.NoError(t, err, "basic allowance creation failed")
	_, err = server.GrantAllowance(sdk.WrapSDKContext(ctx), allowance)
	require.NoError(t, err, "failed to grant basic allowance from admin")
}

// testUserAddress gives a quick way to make a valid test address (no keys though)
func testUserAddress(name string) sdk.AccAddress {
	addr := types.MustGetMarkerAddress(name)
	return addr
}

func TestAddFinalizeActivateMarker(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	server := markerkeeper.NewMsgServerImpl(app.MarkerKeeper)

	addr := types.MustGetMarkerAddress("testcoin")
	pubkey := secp256k1.GenPrivKey().PubKey()
	user := testUserAddress("testcoin")
	manager := testUserAddress("manager")
	existingBalance := sdk.NewCoin("coin", sdk.NewInt(1000))

	// prefund the marker address so an account gets created before the marker does.
	app.AccountKeeper.SetAccount(ctx, authtypes.NewBaseAccount(user, pubkey, 0, 0))
	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, addr, sdk.NewCoins(existingBalance)), "funding account")
	require.Equal(t, existingBalance, app.BankKeeper.GetBalance(ctx, addr, "coin"), "account balance must be set")

	// Creating a marker over an account with zero sequence is fine.
	// One shot marker creation
	_, err := server.AddFinalizeActivateMarker(sdk.WrapSDKContext(ctx), types.NewMsgAddFinalizeActivateMarkerRequest(
		"testcoin",
		sdk.NewInt(30),
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
	_, err = server.AddFinalizeActivateMarker(sdk.WrapSDKContext(ctx), types.NewMsgAddFinalizeActivateMarkerRequest(
		"testcoin",
		sdk.NewInt(30),
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
	require.EqualValues(t, app.MarkerKeeper.GetEscrow(ctx, m).AmountOf("testcoin"), sdk.NewInt(30))
}

// Creating a marker over an existing account with a positive sequence number fails.
func TestInvalidAccount(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	pubkey := secp256k1.GenPrivKey().PubKey()
	server := markerkeeper.NewMsgServerImpl(app.MarkerKeeper)
	user := testUserAddress("testcoin")
	manager := testUserAddress("manager")

	// replace existing test account with a new copy that has a positive sequence number
	app.AccountKeeper.SetAccount(ctx, authtypes.NewBaseAccount(user, pubkey, 0, 10))

	_, err := server.AddFinalizeActivateMarker(sdk.WrapSDKContext(ctx), types.NewMsgAddFinalizeActivateMarkerRequest(
		"testcoin",
		sdk.NewInt(30),
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
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	server := markerkeeper.NewMsgServerImpl(app.MarkerKeeper)

	user := testUserAddress("test")

	// Require a long unrestricted denom
	app.MarkerKeeper.SetParams(ctx, types.Params{UnrestrictedDenomRegex: "[a-z]{12,20}"})

	_, err := server.AddFinalizeActivateMarker(sdk.WrapSDKContext(ctx),
		types.NewMsgAddFinalizeActivateMarkerRequest(
			"tooshort",
			sdk.NewInt(30),
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

	_, err = server.AddFinalizeActivateMarker(sdk.WrapSDKContext(ctx), types.NewMsgAddFinalizeActivateMarkerRequest(
		"itslongenough",
		sdk.NewInt(30),
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
	_, err = server.AddFinalizeActivateMarker(sdk.WrapSDKContext(ctx), types.NewMsgAddFinalizeActivateMarkerRequest(
		"short",
		sdk.NewInt(30),
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
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	server := markerkeeper.NewMsgServerImpl(app.MarkerKeeper)

	newMsg := func(denom string, amt math.Int, manager string, status types.MarkerStatus,
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
			newMsg("test1", sdk.NewInt(100), "", active, coin, []types.AccessGrant{}, true),
			nil,
		},
		{
			"add marker - valid restricted marker",
			newMsg("testrestricted", sdk.NewInt(100), "", active, restricted, []types.AccessGrant{}, true),
			nil,
		},
		{
			"add marker - valid no governance",
			newMsg("testnogov", sdk.NewInt(100), user, active, coin, []types.AccessGrant{}, false),
			nil,
		},
		{
			"add marker - valid finalized",
			newMsg("pending", sdk.NewInt(100), user, finalized, coin, []types.AccessGrant{}, true),
			nil,
		},
		{
			"add marker - already exists",
			newMsg("test1", sdk.NewInt(0), "", active, coin, []types.AccessGrant{}, true),
			fmt.Errorf("marker address already exists for cosmos1ku2jzvpkt4ffxxaajyk2r88axk9cr5jqlthcm4: invalid request"),
		},
		{
			"add marker - invalid status",
			newMsg("test2", sdk.NewInt(100), "", types.StatusUndefined, coin, []types.AccessGrant{}, true),
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
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
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
				Amount: math.NewInt(100),
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
				Amount: math.NewInt(100),
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
				Amount: math.NewInt(100),
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
				Amount: math.NewInt(100),
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
				Amount: math.NewInt(100),
				Denom:  beesKnees.Denom,
			},
			targetAddress: authority,
			authority:     authority,
			shouldFail:    false,
			expectedError: "",
		},
	}

	for _, tc := range testCases {
		res, err := server.SupplyIncreaseProposal(sdk.WrapSDKContext(ctx),
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
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
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
	app.MarkerKeeper.SetMarker(ctx, types.NewMarkerAccount(rMarkerAcct, sdk.NewInt64Coin(rMarkerDenom, 1000), transferAuthUser, []types.AccessGrant{{Address: transferAuthUser.String(), Permissions: []types.Access{types.Access_Transfer}}}, types.StatusFinalized, types.MarkerType_RestrictedCoin, true, true, false, reqAttr))

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
			res, err := server.UpdateRequiredAttributes(sdk.WrapSDKContext(ctx),
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

func TestReqAttrBypassAddrs(t *testing.T) {
	// Tests both GetReqAttrBypassAddrs and IsReqAttrBypassAddr.
	expectedNames := []string{
		authtypes.FeeCollectorName,
		rewardtypes.ModuleName,
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
			if assert.NotSame(t, expected, actual, "GetReqAttrBypassAddrs()") {
				for i := range expected {
					assert.NotSame(t, expected[i], actual[i], "GetReqAttrBypassAddrs()[%d]", i)
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

// dummyBankKeeper satisfies the types.BankKeeper interface but does nothing.
type dummyBankKeeper struct{}

var _ types.BankKeeper = (*dummyBankKeeper)(nil)

func (d dummyBankKeeper) GetAllBalances(_ sdk.Context, _ sdk.AccAddress) sdk.Coins { return nil }

func (d dummyBankKeeper) GetBalance(_ sdk.Context, _ sdk.AccAddress, denom string) sdk.Coin {
	return sdk.Coin{}
}

func (d dummyBankKeeper) GetSupply(_ sdk.Context, _ string) sdk.Coin { return sdk.Coin{} }

func (d dummyBankKeeper) DenomOwners(_ context.Context, _ *banktypes.QueryDenomOwnersRequest) (*banktypes.QueryDenomOwnersResponse, error) {
	return nil, nil
}

func (d dummyBankKeeper) SendCoins(_ sdk.Context, _, _ sdk.AccAddress, amt sdk.Coins) error {
	return nil
}

func (d dummyBankKeeper) SendCoinsFromModuleToAccount(_ sdk.Context, _ string, _ sdk.AccAddress, _ sdk.Coins) error {
	return nil
}

func (d dummyBankKeeper) SendCoinsFromAccountToModule(_ sdk.Context, _ sdk.AccAddress, _ string, _ sdk.Coins) error {
	return nil
}

func (d dummyBankKeeper) MintCoins(_ sdk.Context, _ string, _ sdk.Coins) error { return nil }

func (d dummyBankKeeper) BurnCoins(_ sdk.Context, _ string, _ sdk.Coins) error { return nil }

func (d dummyBankKeeper) AppendSendRestriction(_ banktypes.SendRestrictionFn) {}

func (d dummyBankKeeper) BlockedAddr(_ sdk.AccAddress) bool { return false }

func (d dummyBankKeeper) GetDenomMetaData(_ sdk.Context, _ string) (banktypes.Metadata, bool) {
	return banktypes.Metadata{}, false
}

func (d dummyBankKeeper) SetDenomMetaData(_ sdk.Context, _ banktypes.Metadata) {}

func (d dummyBankKeeper) IterateAllBalances(_ sdk.Context, _ func(sdk.AccAddress, sdk.Coin) bool) {}

func (d dummyBankKeeper) GetAllSendEnabledEntries(_ sdk.Context) []banktypes.SendEnabled { return nil }

func (d dummyBankKeeper) DeleteSendEnabled(_ sdk.Context, _ string) {}

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

	mk := markerkeeper.NewKeeper(nil, nil, paramtypes.NewSubspace(nil, nil, nil, nil, "test"), nil, &dummyBankKeeper{}, nil, nil, nil, nil, nil, addrs)

	// Now that the keeper has been created using the provided addresses, change the first byte of
	// the first address to something else. Then, get the addresses back from the keeper and make
	// sure that change didn't affect what's in the keeper.
	orig00 := addrs[0][0]
	addrs[0][0] = 'b'
	kAddrs := mk.GetReqAttrBypassAddrs()
	act00 := kAddrs[0][0]
	assert.Equal(t, orig00, act00, "first byte of first address returned by GetReqAttrBypassAddrs")
}
