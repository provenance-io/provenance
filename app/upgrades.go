package app

import (
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	msgfeetypes "github.com/provenance-io/provenance/x/msgfees/types"
)

var (
	noopHandler = func(ctx sdk.Context, plan upgradetypes.Plan, versionMap module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Applying no-op upgrade plan for release " + plan.Name)
		return versionMap, nil
	}
)

type appUpgradeHandler = func(sdk.Context, *App, upgradetypes.Plan) (module.VersionMap, error)

// appUpgrade is an internal structure for defining all things for an upgrade.
type appUpgrade struct {
	Added   []string
	Deleted []string
	Renamed []storetypes.StoreRename
	Handler appUpgradeHandler
}

var handlers = map[string]appUpgrade{
	"quicksilver-rc1": { // upgrade for v1.15.0-rc2
		Handler: func(ctx sdk.Context, app *App, plan upgradetypes.Plan) (module.VersionMap, error) {
			app.MarkerKeeper.RemoveIsSendEnabledEntries(ctx)
			app.AttributeKeeper.PopulateAddressAttributeNameTable(ctx)
			versionMap := app.UpgradeKeeper.GetModuleVersionMap(ctx)
			ctx.Logger().Info("Starting migrations. This may take a significant amount of time to complete. Do not restart node.")
			return app.mm.RunMigrations(ctx, app.configurator, versionMap)
		},
	},
	"quicksilver-rc2": {}, // upgrade for v1.15.0-rc3
	"quicksilver": { // upgrade for v1.15.0
		Handler: func(ctx sdk.Context, app *App, plan upgradetypes.Plan) (module.VersionMap, error) {
			app.MarkerKeeper.RemoveIsSendEnabledEntries(ctx)
			app.AttributeKeeper.PopulateAddressAttributeNameTable(ctx)
			versionMap := app.UpgradeKeeper.GetModuleVersionMap(ctx)
			ctx.Logger().Info("Starting migrations. This may take a significant amount of time to complete. Do not restart node.")
			return app.mm.RunMigrations(ctx, app.configurator, versionMap)
		},
	},
	"rust-rc1": { // upgrade for v1.16.0-rc1,
		Handler: func(ctx sdk.Context, app *App, plan upgradetypes.Plan) (module.VersionMap, error) {
			versionMap, err := runModuleMigrations(ctx, app)
			if err != nil {
				return nil, err
			}

			// We only need to call AddGovV1SubmitFee on testnet.
			err = AddGovV1SubmitFee(ctx, app)
			if err != nil {
				return nil, err
			}

			err = RemoveP8eMemorializeContractFee(ctx, app)
			if err != nil {
				return nil, err
			}

			return versionMap, nil
		},
	},
	"rust": { // upgrade for v1.16.0,
		Handler: func(ctx sdk.Context, app *App, plan upgradetypes.Plan) (module.VersionMap, error) {
			versionMap, err := runModuleMigrations(ctx, app)
			if err != nil {
				return nil, err
			}

			// No need to call AddGovV1SubmitFee in here as mainnet already has it defined.

			err = RemoveP8eMemorializeContractFee(ctx, app)
			if err != nil {
				return nil, err
			}

			return versionMap, nil
		},
	},
	// TODO - Add new upgrade definitions here.
}

// InstallCustomUpgradeHandlers sets upgrade handlers for all entries in the handlers map.
func InstallCustomUpgradeHandlers(app *App) {
	// Register all explicit appUpgrades
	for name, upgrade := range handlers {
		// If the handler has been defined, add it here, otherwise, use no-op.
		var handler upgradetypes.UpgradeHandler
		if upgrade.Handler == nil {
			handler = noopHandler
		} else {
			ref := upgrade
			handler = func(ctx sdk.Context, plan upgradetypes.Plan, versionMap module.VersionMap) (module.VersionMap, error) {
				vM, err := ref.Handler(ctx, app, plan)
				if err != nil {
					ctx.Logger().Info(fmt.Sprintf("Failed to upgrade to: %s with err: %v", plan.Name, err))
				} else {
					ctx.Logger().Info(fmt.Sprintf("Successfully upgraded to: %s with version map: %v", plan.Name, vM))
				}
				return vM, err
			}
		}
		app.UpgradeKeeper.SetUpgradeHandler(name, handler)
	}
}

// GetUpgradeStoreLoader creates an StoreLoader for use in an upgrade.
// Returns nil if no upgrade info is found or the upgrade doesn't need a store loader.
func GetUpgradeStoreLoader(app *App, info upgradetypes.Plan) baseapp.StoreLoader {
	upgrade, found := handlers[info.Name]
	if !found {
		return nil
	}

	if len(upgrade.Renamed) == 0 && len(upgrade.Deleted) == 0 && len(upgrade.Added) == 0 {
		app.Logger().Info("No store upgrades required",
			"plan", info.Name,
			"height", info.Height,
		)
		return nil
	}

	storeUpgrades := storetypes.StoreUpgrades{
		Added:   upgrade.Added,
		Renamed: upgrade.Renamed,
		Deleted: upgrade.Deleted,
	}
	app.Logger().Info("Store upgrades",
		"plan", info.Name,
		"height", info.Height,
		"upgrade.added", storeUpgrades.Added,
		"upgrade.deleted", storeUpgrades.Deleted,
		"upgrade.renamed", storeUpgrades.Renamed,
	)
	return upgradetypes.UpgradeStoreLoader(info.Height, &storeUpgrades)
}

// runModuleMigrations wraps standard logging around the call to app.mm.RunMigrations.
// In most cases, it should be the first thing done during a migration.
//
// If state is updated prior to this migration, you run the risk of writing state using
// a new format when the migration is expecting all state to be in the old format.
func runModuleMigrations(ctx sdk.Context, app *App) (module.VersionMap, error) {
	// Even if this function is no longer called, do not delete it. Keep it around for the next time it's needed.m
	ctx.Logger().Info("Starting module migrations. This may take a significant amount of time to complete. Do not restart node.")
	versionMap := app.UpgradeKeeper.GetModuleVersionMap(ctx)
	var err error
	versionMap, err = app.mm.RunMigrations(ctx, app.configurator, versionMap)
	if err != nil {
		ctx.Logger().Error("Module migrations encountered an error.", "error", err)
		return nil, err
	}
	ctx.Logger().Info("Module migrations completed.")
	return versionMap, nil
}

// Create a use of runModuleMigrations so that the linter neither complains about it not being used,
// nor complains about a nolint:unused directive that isn't needed because the function is used.
var _ = runModuleMigrations

// AddGovV1SubmitFee adds a msg-fee for the gov v1 MsgSubmitProposal if there isn't one yet.
func AddGovV1SubmitFee(ctx sdk.Context, app *App) (err error) {
	typeURL := sdk.MsgTypeURL(&govtypesv1.MsgSubmitProposal{})
	defer func() {
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("Error encountered while setting message fee for %q", typeURL), "error", err)
		}
	}()

	ctx.Logger().Info(fmt.Sprintf("Creating message fee for %q if it doesn't already exist.", typeURL))
	fee, err := app.MsgFeesKeeper.GetMsgFee(ctx, typeURL)
	if err != nil {
		return fmt.Errorf("error getting existing v1 fee for %q: %w", typeURL, err)
	}
	// If there's already a fee for it, do nothing.
	if fee != nil {
		ctx.Logger().Info(fmt.Sprintf("Message fee for %q already exists with amount %q. Nothing more to do.", fee.MsgTypeUrl, fee.AdditionalFee.String()))
		return nil
	}

	// Copy the fee from the beta entry if it exists, otherwise, just make it fresh.
	betaTypeURL := sdk.MsgTypeURL(&govtypesv1beta1.MsgSubmitProposal{})
	betaFee, err := app.MsgFeesKeeper.GetMsgFee(ctx, betaTypeURL)
	if err != nil {
		return fmt.Errorf("error getting existing v1beta1 message fee for %q: %w", betaTypeURL, err)
	}
	if betaFee != nil {
		fee = betaFee
		fee.MsgTypeUrl = typeURL
		ctx.Logger().Info(fmt.Sprintf("Copying %q fee to %q.", betaTypeURL, fee.MsgTypeUrl))
	} else {
		fee = &msgfeetypes.MsgFee{
			MsgTypeUrl:           typeURL,
			AdditionalFee:        sdk.NewInt64Coin("nhash", 100_000_000_000), // 100 hash
			Recipient:            "",
			RecipientBasisPoints: 0,
		}
		ctx.Logger().Info(fmt.Sprintf("Creating %q fee.", fee.MsgTypeUrl))
	}

	err = app.MsgFeesKeeper.SetMsgFee(ctx, *fee)
	if err != nil {
		return fmt.Errorf("error setting message fee for %q: %w", fee.MsgTypeUrl, err)
	}
	ctx.Logger().Info("Successfully set fee for %q with amount %q.", fee.MsgTypeUrl, fee.AdditionalFee.String())

	return nil
}

// RemoveP8eMemorializeContractFee removes the message fee for the now-non-existent MsgP8eMemorializeContractRequest.
func RemoveP8eMemorializeContractFee(ctx sdk.Context, app *App) (err error) {
	typeURL := "/provenance.metadata.v1.MsgP8eMemorializeContractRequest"
	defer func() {
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("Error encountered while removing message fee for %q", typeURL), "error", err)
		}
	}()

	ctx.Logger().Info(fmt.Sprintf("Removing message fee for %q if one exists.", typeURL))
	// Get the existing fee for log output, but ignore any errors so we try to delete the entry either way.
	fee, _ := app.MsgFeesKeeper.GetMsgFee(ctx, typeURL)
	err = app.MsgFeesKeeper.RemoveMsgFee(ctx, typeURL)
	switch {
	case errors.Is(err, msgfeetypes.ErrMsgFeeDoesNotExist):
		ctx.Logger().Info(fmt.Sprintf("Message fee for %q already does not exist. Nothing more to do.", typeURL))
		err = nil
	case err != nil:
		return fmt.Errorf("error removing message fee for %q: %w", typeURL, err)
	case fee != nil:
		ctx.Logger().Info(fmt.Sprintf("Successfully removed message fee for %q with amount %q.", fee.MsgTypeUrl, fee.AdditionalFee.String()))
	default:
		ctx.Logger().Info(fmt.Sprintf("Successfully removed message fee for %q.", typeURL))
	}

	return nil
}
