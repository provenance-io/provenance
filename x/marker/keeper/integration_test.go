package keeper_test

import (
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	markerkeeper "github.com/provenance-io/provenance/x/marker/keeper"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

type MarkerSimApp struct {
	*simapp.SimApp

	MarkerKeeper markerkeeper.Keeper
}

var maccPerms map[string][]string

// returns context and app with params set on account keeper
func createTestApp(isCheckTx bool) (*MarkerSimApp, sdk.Context) {
	sim := simapp.Setup(isCheckTx)

	// add module accounts to bank keeper
	maccPerms = simapp.GetMaccPerms()
	maccPerms[markertypes.ModuleName] = []string{authtypes.Burner, authtypes.Minter}
	markertypes.RegisterLegacyAminoCodec(sim.LegacyAmino())

	app := MarkerSimApp{SimApp: sim}

	ctx := app.BaseApp.NewContext(isCheckTx, tmproto.Header{})

	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())

	app.BankKeeper.SetSupply(ctx, banktypes.NewSupply(sdk.NewCoins()))

	app.DistrKeeper = distrkeeper.NewKeeper(
		app.AppCodec(), app.GetKey(distrtypes.StoreKey), app.GetSubspace(distrtypes.ModuleName), app.AccountKeeper, app.BankKeeper, app.StakingKeeper,
		authtypes.FeeCollectorName, app.ModuleAccountAddrs())

	app.MarkerKeeper = markerkeeper.NewKeeper(app.AppCodec(), app.GetKey(markertypes.StoreKey), app.GetSubspace(markertypes.ModuleName), app.AccountKeeper, app.BankKeeper)
	return &app, ctx
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *MarkerSimApp) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// BlacklistedAccAddrs returns all the app's module account addresses black listed for receiving tokens.
func (app *MarkerSimApp) BlacklistedAccAddrs() map[string]bool {
	blacklistedAddrs := make(map[string]bool)
	for acc := range maccPerms {
		blacklistedAddrs[authtypes.NewModuleAddress(acc).String()] = !(acc == distrtypes.ModuleName)
	}

	return blacklistedAddrs
}
