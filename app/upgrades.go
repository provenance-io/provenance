package app

import (
	"context"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	ibctmmigrations "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint/migrations"

	"github.com/provenance-io/provenance/internal/pioconfig"
	flatfeestypes "github.com/provenance-io/provenance/x/flatfees/types"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
)

// appUpgrade is an internal structure for defining all things for an upgrade.
type appUpgrade struct {
	// Added contains names of modules being added during an upgrade.
	Added []string
	// Deleted contains names of modules being removed during an upgrade.
	Deleted []string
	// Renamed contains info on modules being renamed during an upgrade.
	Renamed []storetypes.StoreRename
	// Handler is a function to execute during an upgrade.
	Handler func(sdk.Context, *App, module.VersionMap) (module.VersionMap, error)
}

// upgrades is where we define things that need to happen during an upgrade.
// If no Handler is defined for an entry, a no-op upgrade handler is still registered.
// If there's nothing that needs to be done for an upgrade, there still needs to be an
// entry in this map, but it can just be {}.
//
// If something is happening in the rc upgrade(s) that isn't being applied in the non-rc,
// or vice versa, please add comments explaining why in both entries.
//
// On the same line as the key, there should be a comment indicating the software version.
// Entries should be in chronological/alphabetical order, earliest first.
// I.e. Brand-new colors should be added to the bottom with the rcs first, then the non-rc.
var upgrades = map[string]appUpgrade{
	"alyssum-rc1": { // Upgrade for v1.25.0-rc1.
		Added:   []string{flatfeestypes.StoreKey},
		Deleted: []string{msgfeestypes.StoreKey},
		Handler: func(ctx sdk.Context, app *App, vm module.VersionMap) (module.VersionMap, error) {
			var err error
			if vm, err = runModuleMigrations(ctx, app, vm); err != nil {
				return nil, err
			}
			if err = pruneIBCExpiredConsensusStates(ctx, app); err != nil {
				return nil, err
			}
			removeInactiveValidatorDelegations(ctx, app)
			if err = convertFinishedVestingAccountsToBase(ctx, app); err != nil {
				return nil, err
			}
			if err = setupFlatFees(ctx, app.FlatFeesKeeper); err != nil {
				return nil, err
			}
			return vm, nil
		},
	},
	"alyssum": { // Upgrade for v1.25.0.
		Handler: func(ctx sdk.Context, app *App, vm module.VersionMap) (module.VersionMap, error) {
			var err error
			if vm, err = runModuleMigrations(ctx, app, vm); err != nil {
				return nil, err
			}
			if err = pruneIBCExpiredConsensusStates(ctx, app); err != nil {
				return nil, err
			}
			removeInactiveValidatorDelegations(ctx, app)
			if err = convertFinishedVestingAccountsToBase(ctx, app); err != nil {
				return nil, err
			}
			if err = setupFlatFees(ctx, app.FlatFeesKeeper); err != nil {
				return nil, err
			}
			return vm, nil
		},
	},
}

// InstallCustomUpgradeHandlers sets upgrade handlers for all entries in the upgrades map.
func InstallCustomUpgradeHandlers(app *App) {
	// Register all explicit appUpgrades
	for name, upgrade := range upgrades {
		// If the handler has been defined, add it here, otherwise, use no-op.
		var handler upgradetypes.UpgradeHandler
		if upgrade.Handler == nil {
			handler = func(goCtx context.Context, plan upgradetypes.Plan, versionMap module.VersionMap) (module.VersionMap, error) {
				ctx := sdk.UnwrapSDKContext(goCtx)
				ctx.Logger().Info(fmt.Sprintf("Applying no-op upgrade to %q", plan.Name))
				ctx.EventManager().EmitEvent(sdk.NewEvent("upgrade", sdk.NewAttribute("name", plan.Name)))
				return versionMap, nil
			}
		} else {
			ref := upgrade
			handler = func(goCtx context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
				ctx := sdk.UnwrapSDKContext(goCtx)
				ctx.Logger().Info(fmt.Sprintf("Starting upgrade to %q", plan.Name), "version-map", vm)
				ctx.EventManager().EmitEvent(sdk.NewEvent("upgrade", sdk.NewAttribute("name", plan.Name)))
				newVM, err := ref.Handler(ctx, app, vm)
				if err != nil {
					ctx.Logger().Error(fmt.Sprintf("Failed to upgrade to %q", plan.Name), "error", err)
				} else {
					ctx.Logger().Info(fmt.Sprintf("Successfully upgraded to %q", plan.Name), "version-map", newVM)
				}
				return newVM, err
			}
		}

		app.UpgradeKeeper.SetUpgradeHandler(name, handler)
	}
}

// GetUpgradeStoreLoader creates an StoreLoader for use in an upgrade.
// Returns nil if no upgrade info is found or the upgrade doesn't need a store loader.
func GetUpgradeStoreLoader(app *App, info upgradetypes.Plan) baseapp.StoreLoader {
	upgrade, found := upgrades[info.Name]
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
func runModuleMigrations(ctx sdk.Context, app *App, vm module.VersionMap) (module.VersionMap, error) {
	// Even if this function is no longer called, do not delete it. Keep it around for the next time it's needed.
	ctx.Logger().Info("Starting module migrations. This may take a significant amount of time to complete. Do not restart node.")
	newVM, err := app.mm.RunMigrations(ctx, app.configurator, vm)
	if err != nil {
		ctx.Logger().Error("Module migrations encountered an error.", "error", err)
		return nil, err
	}
	ctx.Logger().Info("Module migrations completed.")
	return newVM, nil
}

// removeInactiveValidatorDelegations unbonds all delegations from inactive validators, triggering their removal from the validator set.
// This should be applied in most upgrades.
func removeInactiveValidatorDelegations(ctx sdk.Context, app *App) {
	ctx.Logger().Info("Removing inactive validator delegations.")

	sParams, perr := app.StakingKeeper.GetParams(ctx)
	if perr != nil {
		ctx.Logger().Error(fmt.Sprintf("Could not get staking params: %v.", perr))
		return
	}

	unbondingTimeParam := sParams.UnbondingTime
	ctx.Logger().Info(fmt.Sprintf("Threshold: %d days", int64(unbondingTimeParam.Hours()/24)))

	validators, verr := app.StakingKeeper.GetAllValidators(ctx)
	if verr != nil {
		ctx.Logger().Error(fmt.Sprintf("Could not get all validators: %v.", perr))
		return
	}

	removalCount := 0
	for _, validator := range validators {
		if validator.IsUnbonded() {
			inactiveDuration := ctx.BlockTime().Sub(validator.UnbondingTime)
			if inactiveDuration >= unbondingTimeParam {
				ctx.Logger().Info(fmt.Sprintf("Validator %v has been inactive (unbonded) for %d days and will be removed.", validator.OperatorAddress, int64(inactiveDuration.Hours()/24)))
				valAddress, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
				if err != nil {
					ctx.Logger().Error(fmt.Sprintf("Invalid operator address: %s: %v.", validator.OperatorAddress, err))
					continue
				}

				delegations, err := app.StakingKeeper.GetValidatorDelegations(ctx, valAddress)
				if err != nil {
					ctx.Logger().Error(fmt.Sprintf("Could not delegations for validator %s: %v.", valAddress, perr))
					continue
				}

				for _, delegation := range delegations {
					ctx.Logger().Info(fmt.Sprintf("Undelegate delegator %v from validator %v of all shares (%v).", delegation.DelegatorAddress, validator.OperatorAddress, delegation.GetShares()))
					var delAddr sdk.AccAddress
					delegator := delegation.GetDelegatorAddr()
					delAddr, err = sdk.AccAddressFromBech32(delegator)
					if err != nil {
						ctx.Logger().Error(fmt.Sprintf("Failed to undelegate delegator %s from validator %s: could not parse delegator address: %v.", delegator, valAddress.String(), err))
						continue
					}
					_, _, err = app.StakingKeeper.Undelegate(ctx, delAddr, valAddress, delegation.GetShares())
					if err != nil {
						ctx.Logger().Error(fmt.Sprintf("Failed to undelegate delegator %s from validator %s: %v.", delegator, valAddress.String(), err))
						continue
					}
				}
				removalCount++
			}
		}
	}

	ctx.Logger().Info(fmt.Sprintf("A total of %d inactive (unbonded) validators have had all their delegators removed.", removalCount))
}

// pruneIBCExpiredConsensusStates prunes expired consensus states for IBC.
// This should be applied in most upgrades.
func pruneIBCExpiredConsensusStates(ctx sdk.Context, app *App) error {
	ctx.Logger().Info("Pruning expired consensus states for IBC.")
	_, err := ibctmmigrations.PruneExpiredConsensusStates(ctx, app.appCodec, app.IBCKeeper.ClientKeeper)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Unable to prune expired consensus states, error: %s.", err))
		return err
	}
	ctx.Logger().Info("Done pruning expired consensus states for IBC.")
	return nil
}

// convertFinishedVestingAccountsToBase will turn completed vesting accounts into regular BaseAccounts.
// This should be applied in most upgrades.
func convertFinishedVestingAccountsToBase(ctx sdk.Context, app *App) error {
	ctx.Logger().Info("Converting completed vesting accounts into base accounts.")
	blockTime := ctx.BlockTime().UTC().Unix()
	var updatedAccts []sdk.AccountI
	err := app.AccountKeeper.Accounts.Walk(ctx, nil, func(_ sdk.AccAddress, acctI sdk.AccountI) (stop bool, err error) {
		var baseVestAcct *vesting.BaseVestingAccount
		switch acct := acctI.(type) {
		case *vesting.ContinuousVestingAccount:
			baseVestAcct = acct.BaseVestingAccount
		case *vesting.DelayedVestingAccount:
			baseVestAcct = acct.BaseVestingAccount
		case *vesting.PeriodicVestingAccount:
			baseVestAcct = acct.BaseVestingAccount
		default:
			// We don't care about permanent locked because they never end.
			// Nothing should ever be a *vesting.BaseVestingAccount, so we ignore those too.
			// All other accounts aren't a vesting account, so there's nothing to do with them here.
			return false, nil
		}

		if baseVestAcct.EndTime <= blockTime {
			updatedAccts = append(updatedAccts, baseVestAcct.BaseAccount)
		}
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("error walking accounts: %w", err)
	}

	if len(updatedAccts) == 0 {
		ctx.Logger().Info("No completed vesting accounts found.")
	} else {
		ctx.Logger().Info(fmt.Sprintf("Found %d completed vesting accounts. Updating them now.", len(updatedAccts)))
		for _, acct := range updatedAccts {
			app.AccountKeeper.SetAccount(ctx, acct)
		}
	}

	ctx.Logger().Info("Done converting completed vesting accounts into base accounts.")
	return nil
}

// unlockVestingAccounts will convert the provided addrs from ContinuousVestingAccount to BaseAccount.
// This might be needed later, for another round of unlocks, so we're keeping it around in case.
func unlockVestingAccounts(ctx sdk.Context, app *App, addrs []sdk.AccAddress) {
	ctx.Logger().Info("Unlocking select vesting accounts.")

	ctx.Logger().Info(fmt.Sprintf("Identified %d accounts to unlock.", len(addrs)))
	for i, addr := range addrs {
		acctI := app.AccountKeeper.GetAccount(ctx, addr)
		if acctI == nil {
			ctx.Logger().Info(fmt.Sprintf("[%d/%d]: Cannot unlock account %s: account not found.", i+1, len(addrs), addr))
			continue
		}
		switch acct := acctI.(type) {
		case *vesting.ContinuousVestingAccount:
			app.AccountKeeper.SetAccount(ctx, acct.BaseAccount)
			ctx.Logger().Debug(fmt.Sprintf("[%d/%d]: Unlocked account: %s.", i+1, len(addrs), addr))
		case *vesting.DelayedVestingAccount:
			app.AccountKeeper.SetAccount(ctx, acct.BaseAccount)
			ctx.Logger().Debug(fmt.Sprintf("[%d/%d]: Unlocked account: %s.", i+1, len(addrs), addr))
		case *vesting.PeriodicVestingAccount:
			app.AccountKeeper.SetAccount(ctx, acct.BaseAccount)
			ctx.Logger().Debug(fmt.Sprintf("[%d/%d]: Unlocked account: %s.", i+1, len(addrs), addr))
		case *vesting.PermanentLockedAccount:
			app.AccountKeeper.SetAccount(ctx, acct.BaseAccount)
			ctx.Logger().Debug(fmt.Sprintf("[%d/%d]: Unlocked account: %s.", i+1, len(addrs), addr))
		default:
			ctx.Logger().Info(fmt.Sprintf("[%d/%d]: Cannot unlock account %s: not a vesting account: %T.", i+1, len(addrs), addr, acctI))
		}
	}

	ctx.Logger().Info("Done unlocking select vesting accounts.")
}

// Create a use of the standard helpers so that the linter neither complains about it not being used,
// nor complains about a nolint:unused directive that isn't needed because the function is used.
var (
	_ = runModuleMigrations
	_ = removeInactiveValidatorDelegations
	_ = pruneIBCExpiredConsensusStates
	_ = convertFinishedVestingAccountsToBase
	_ = unlockVestingAccounts
)

// FlatFeesKeeper has the flatfees keeper methods needed for setting up flat fees.
// Part of the alyssum upgrade.
type FlatFeesKeeper interface {
	SetParams(ctx sdk.Context, params flatfeestypes.Params) error
	SetMsgFee(ctx sdk.Context, msgFee flatfeestypes.MsgFee) error
}

// setupFlatFees defines the flatfees module params and msg costs.
// Part of the alyssum upgrade.
func setupFlatFees(ctx sdk.Context, ffk FlatFeesKeeper) error {
	ctx.Logger().Info("Setting up flat fees.")

	params := MakeFlatFeesParams()
	err := ffk.SetParams(ctx, params)
	if err != nil {
		return fmt.Errorf("could not set x/flatfees params: %w", err)
	}

	msgFees := MakeFlatFeesCosts()
	for _, msgFee := range msgFees {
		err = ffk.SetMsgFee(ctx, *msgFee)
		if err != nil {
			return fmt.Errorf("could not set msg fee %s: %w", msgFee, err)
		}
	}

	ctx.Logger().Info("Done setting up flat fees.")
	return nil
}

// feeDefCoin returns a coin in the fee definition denom with the given amount.
// Part of the alyssum upgrade.
func feeDefCoin(amount int64) sdk.Coin {
	return sdk.NewInt64Coin(flatfeestypes.DefaultFeeDefinitionDenom, amount)
}

// MakeFlatFeesParams returns the params to give the flatfeees module.
func MakeFlatFeesParams() flatfeestypes.Params {
	return flatfeestypes.Params{
		DefaultCost: feeDefCoin(150),
		ConversionFactor: flatfeestypes.ConversionFactor{
			// 1 hash = $0.025, so 1000000000nhash = 25musd
			DefinitionAmount: feeDefCoin(25),
			ConvertedAmount:  sdk.NewInt64Coin(pioconfig.GetProvConfig().FeeDenom, 1000000000),
		},
	}
}

// MakeFlatFeesCosts returns the list of MsgFees that we want to set.
// Part of the alyssum upgrade.
func MakeFlatFeesCosts() []*flatfeestypes.MsgFee {
	// TODO[fees]: Identify the new Msgs being added how much we want them to cost, and add them to this list.
	return []*flatfeestypes.MsgFee{
		// Free Msg types. These are gov-prop-only Msg types. A gov prop costs $2.50 + the cost of each msg in it.
		// So even though these msgs are free, it'll still cost $2.50 to submit one.
		flatfeestypes.NewMsgFee("/cosmos.auth.v1beta1.MsgUpdateParams"),
		flatfeestypes.NewMsgFee("/cosmos.bank.v1beta1.MsgSetSendEnabled"),
		flatfeestypes.NewMsgFee("/cosmos.bank.v1beta1.MsgUpdateDenomMetadata"),
		flatfeestypes.NewMsgFee("/cosmos.bank.v1beta1.MsgUpdateParams"),
		flatfeestypes.NewMsgFee("/cosmos.consensus.v1.MsgUpdateParams"),
		flatfeestypes.NewMsgFee("/cosmos.crisis.v1beta1.MsgUpdateParams"),
		flatfeestypes.NewMsgFee("/cosmos.distribution.v1beta1.MsgCommunityPoolSpend"),
		flatfeestypes.NewMsgFee("/cosmos.distribution.v1beta1.MsgUpdateParams"),
		flatfeestypes.NewMsgFee("/cosmos.gov.v1.MsgExecLegacyContent"),
		flatfeestypes.NewMsgFee("/cosmos.gov.v1.MsgUpdateParams"),
		flatfeestypes.NewMsgFee("/cosmos.mint.v1beta1.MsgUpdateParams"),
		flatfeestypes.NewMsgFee("/cosmos.sanction.v1beta1.MsgSanction"),
		flatfeestypes.NewMsgFee("/cosmos.sanction.v1beta1.MsgUnsanction"),
		flatfeestypes.NewMsgFee("/cosmos.sanction.v1beta1.MsgUpdateParams"),
		flatfeestypes.NewMsgFee("/cosmos.slashing.v1beta1.MsgUpdateParams"),
		flatfeestypes.NewMsgFee("/cosmos.staking.v1beta1.MsgUpdateParams"),
		flatfeestypes.NewMsgFee("/cosmos.upgrade.v1beta1.MsgCancelUpgrade"),
		flatfeestypes.NewMsgFee("/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade"),
		flatfeestypes.NewMsgFee("/cosmwasm.wasm.v1.MsgAddCodeUploadParamsAddresses"),
		flatfeestypes.NewMsgFee("/cosmwasm.wasm.v1.MsgPinCodes"),
		flatfeestypes.NewMsgFee("/cosmwasm.wasm.v1.MsgRemoveCodeUploadParamsAddresses"),
		flatfeestypes.NewMsgFee("/cosmwasm.wasm.v1.MsgStoreAndInstantiateContract"),
		flatfeestypes.NewMsgFee("/cosmwasm.wasm.v1.MsgStoreAndMigrateContract"),
		flatfeestypes.NewMsgFee("/cosmwasm.wasm.v1.MsgSudoContract"),
		flatfeestypes.NewMsgFee("/cosmwasm.wasm.v1.MsgUnpinCodes"),
		flatfeestypes.NewMsgFee("/cosmwasm.wasm.v1.MsgUpdateParams"),
		flatfeestypes.NewMsgFee("/ibc.applications.interchain_accounts.controller.v1.MsgUpdateParams"),
		flatfeestypes.NewMsgFee("/ibc.applications.interchain_accounts.host.v1.MsgUpdateParams"),
		flatfeestypes.NewMsgFee("/ibc.applications.transfer.v1.MsgUpdateParams"),
		flatfeestypes.NewMsgFee("/ibc.core.channel.v1.MsgUpdateParams"),
		flatfeestypes.NewMsgFee("/ibc.core.client.v1.MsgUpdateParams"),
		flatfeestypes.NewMsgFee("/ibc.core.connection.v1.MsgUpdateParams"),
		flatfeestypes.NewMsgFee("/icq.v1.MsgUpdateParams"),
		flatfeestypes.NewMsgFee("/provenance.attribute.v1.MsgUpdateParamsRequest"),
		flatfeestypes.NewMsgFee("/provenance.exchange.v1.MsgGovCloseMarketRequest"),
		flatfeestypes.NewMsgFee("/provenance.exchange.v1.MsgGovCreateMarketRequest"),
		flatfeestypes.NewMsgFee("/provenance.exchange.v1.MsgGovManageFeesRequest"),
		flatfeestypes.NewMsgFee("/provenance.exchange.v1.MsgGovUpdateParamsRequest"),
		flatfeestypes.NewMsgFee("/provenance.exchange.v1.MsgUpdateParamsRequest"),
		flatfeestypes.NewMsgFee("/provenance.flatfees.v1.MsgUpdateConversionFactorRequest"),
		flatfeestypes.NewMsgFee("/provenance.flatfees.v1.MsgUpdateMsgFeesRequest"),
		flatfeestypes.NewMsgFee("/provenance.flatfees.v1.MsgUpdateParamsRequest"),
		flatfeestypes.NewMsgFee("/provenance.ibchooks.v1.MsgUpdateParamsRequest"),
		flatfeestypes.NewMsgFee("/provenance.ibcratelimit.v1.MsgGovUpdateParamsRequest"),
		flatfeestypes.NewMsgFee("/provenance.ibcratelimit.v1.MsgUpdateParamsRequest"),
		flatfeestypes.NewMsgFee("/provenance.marker.v1.MsgChangeStatusProposalRequest"),
		flatfeestypes.NewMsgFee("/provenance.marker.v1.MsgRemoveAdministratorProposalRequest"),
		flatfeestypes.NewMsgFee("/provenance.marker.v1.MsgSetAdministratorProposalRequest"),
		flatfeestypes.NewMsgFee("/provenance.marker.v1.MsgSetDenomMetadataProposalRequest"),
		flatfeestypes.NewMsgFee("/provenance.marker.v1.MsgSupplyDecreaseProposalRequest"),
		flatfeestypes.NewMsgFee("/provenance.marker.v1.MsgSupplyIncreaseProposalRequest"),
		flatfeestypes.NewMsgFee("/provenance.marker.v1.MsgUpdateForcedTransferRequest"),
		flatfeestypes.NewMsgFee("/provenance.marker.v1.MsgUpdateParamsRequest"),
		flatfeestypes.NewMsgFee("/provenance.marker.v1.MsgWithdrawEscrowProposalRequest"),
		flatfeestypes.NewMsgFee("/provenance.msgfees.v1.MsgAddMsgFeeProposalRequest"),
		flatfeestypes.NewMsgFee("/provenance.msgfees.v1.MsgAssessCustomMsgFeeRequest"),
		flatfeestypes.NewMsgFee("/provenance.msgfees.v1.MsgRemoveMsgFeeProposalRequest"),
		flatfeestypes.NewMsgFee("/provenance.msgfees.v1.MsgUpdateConversionFeeDenomProposalRequest"),
		flatfeestypes.NewMsgFee("/provenance.msgfees.v1.MsgUpdateMsgFeeProposalRequest"),
		flatfeestypes.NewMsgFee("/provenance.msgfees.v1.MsgUpdateNhashPerUsdMilProposalRequest"),
		flatfeestypes.NewMsgFee("/provenance.name.v1.MsgCreateRootNameRequest"),
		flatfeestypes.NewMsgFee("/provenance.name.v1.MsgUpdateParamsRequest"),
		flatfeestypes.NewMsgFee("/provenance.oracle.v1.MsgUpdateOracleRequest"),

		// Msgs that cost $0.05.
		flatfeestypes.NewMsgFee("/cosmos.authz.v1beta1.MsgExec", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/cosmos.authz.v1beta1.MsgRevoke", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/cosmos.bank.v1beta1.MsgSend", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/cosmos.distribution.v1beta1.MsgDepositValidatorRewardsPool", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/cosmos.distribution.v1beta1.MsgFundCommunityPool", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/cosmos.distribution.v1beta1.MsgSetWithdrawAddress", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommission", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/cosmos.gov.v1.MsgCancelProposal", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/cosmos.gov.v1.MsgDeposit", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/cosmos.gov.v1.MsgVote", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/cosmos.gov.v1.MsgVoteWeighted", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/cosmos.gov.v1beta1.MsgDeposit", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/cosmos.group.v1.MsgExec", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/cosmos.group.v1.MsgWithdrawProposal", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/ibc.core.channel.v1.MsgAcknowledgement", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/ibc.core.channel.v1.MsgRecvPacket", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/provenance.exchange.v1.MsgAcceptPaymentRequest", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/provenance.exchange.v1.MsgCommitFundsRequest", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/provenance.exchange.v1.MsgMarketReleaseCommitmentsRequest", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/provenance.exchange.v1.MsgRejectPaymentRequest", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/provenance.exchange.v1.MsgRejectPaymentsRequest", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/provenance.marker.v1.MsgActivateRequest", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/provenance.marker.v1.MsgFinalizeRequest", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/provenance.marker.v1.MsgGrantAllowanceRequest", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/provenance.marker.v1.MsgSetDenomMetadataRequest", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/provenance.metadata.v1.MsgAddNetAssetValuesRequest", feeDefCoin(50)),

		// Msgs that cost $0.10.
		flatfeestypes.NewMsgFee("/cosmos.authz.v1beta1.MsgGrant", feeDefCoin(100)),
		flatfeestypes.NewMsgFee("/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward", feeDefCoin(100)),
		flatfeestypes.NewMsgFee("/cosmos.feegrant.v1beta1.MsgGrantAllowance", feeDefCoin(100)),
		flatfeestypes.NewMsgFee("/cosmos.gov.v1beta1.MsgVote", feeDefCoin(100)),
		flatfeestypes.NewMsgFee("/cosmos.gov.v1beta1.MsgVoteWeighted", feeDefCoin(100)),
		flatfeestypes.NewMsgFee("/cosmwasm.wasm.v1.MsgMigrateContract", feeDefCoin(100)),
		flatfeestypes.NewMsgFee("/ibc.applications.transfer.v1.MsgTransfer", feeDefCoin(100)),
		flatfeestypes.NewMsgFee("/provenance.attribute.v1.MsgDeleteAttributeRequest", feeDefCoin(100)),
		flatfeestypes.NewMsgFee("/provenance.attribute.v1.MsgUpdateAttributeRequest", feeDefCoin(100)),
		flatfeestypes.NewMsgFee("/provenance.exchange.v1.MsgMarketUpdateAcceptingCommitmentsRequest", feeDefCoin(100)),
		flatfeestypes.NewMsgFee("/provenance.exchange.v1.MsgMarketUpdateAcceptingOrdersRequest", feeDefCoin(100)),
		flatfeestypes.NewMsgFee("/provenance.exchange.v1.MsgMarketUpdateDetailsRequest", feeDefCoin(100)),
		flatfeestypes.NewMsgFee("/provenance.exchange.v1.MsgMarketUpdateEnabledRequest", feeDefCoin(100)),
		flatfeestypes.NewMsgFee("/provenance.exchange.v1.MsgMarketUpdateIntermediaryDenomRequest", feeDefCoin(100)),
		flatfeestypes.NewMsgFee("/provenance.exchange.v1.MsgMarketUpdateUserSettleRequest", feeDefCoin(100)),
		flatfeestypes.NewMsgFee("/provenance.trigger.v1.MsgCreateTriggerRequest", feeDefCoin(100)),

		// The default cost is $0.15. All Msg types not in this list will use the default.

		// Msgs that cost $0.25.
		flatfeestypes.NewMsgFee("/provenance.attribute.v1.MsgAddAttributeRequest", feeDefCoin(250)),
		flatfeestypes.NewMsgFee("/provenance.marker.v1.MsgIbcTransferRequest", feeDefCoin(250)),
		flatfeestypes.NewMsgFee("/provenance.name.v1.MsgBindNameRequest", feeDefCoin(250)),
		flatfeestypes.NewMsgFee("/provenance.oracle.v1.MsgSendQueryOracleRequest", feeDefCoin(250)),

		// Msgs that cost $0.50.
		flatfeestypes.NewMsgFee("/cosmos.staking.v1beta1.MsgBeginRedelegate", feeDefCoin(500)),
		flatfeestypes.NewMsgFee("/cosmwasm.wasm.v1.MsgInstantiateContract", feeDefCoin(500)),
		flatfeestypes.NewMsgFee("/cosmwasm.wasm.v1.MsgInstantiateContract2", feeDefCoin(500)),
		flatfeestypes.NewMsgFee("/ibc.core.client.v1.MsgCreateClient", feeDefCoin(500)),
		flatfeestypes.NewMsgFee("/provenance.marker.v1.MsgAddNetAssetValuesRequest", feeDefCoin(500)),
		flatfeestypes.NewMsgFee("/provenance.metadata.v1.MsgMigrateValueOwnerRequest", feeDefCoin(500)),
		flatfeestypes.NewMsgFee("/provenance.metadata.v1.MsgUpdateValueOwnersRequest", feeDefCoin(500)),
		flatfeestypes.NewMsgFee("/provenance.metadata.v1.MsgWriteContractSpecificationRequest", feeDefCoin(500)),
		flatfeestypes.NewMsgFee("/provenance.metadata.v1.MsgWriteRecordSpecificationRequest", feeDefCoin(500)),
		flatfeestypes.NewMsgFee("/provenance.metadata.v1.MsgWriteScopeSpecificationRequest", feeDefCoin(500)),
		flatfeestypes.NewMsgFee("/provenance.metadata.v1.MsgWriteSessionRequest", feeDefCoin(500)),

		// Msgs that cost $1.00.
		flatfeestypes.NewMsgFee("/cosmos.group.v1.MsgSubmitProposal", feeDefCoin(1000)),
		flatfeestypes.NewMsgFee("/cosmwasm.wasm.v1.MsgStoreCode", feeDefCoin(1000)),
		flatfeestypes.NewMsgFee("/provenance.metadata.v1.MsgAddContractSpecToScopeSpecRequest", feeDefCoin(1000)),
		flatfeestypes.NewMsgFee("/provenance.metadata.v1.MsgDeleteContractSpecFromScopeSpecRequest", feeDefCoin(1000)),
		flatfeestypes.NewMsgFee("/provenance.metadata.v1.MsgWriteScopeRequest", feeDefCoin(1000)),

		// Msgs that cost $1.50.
		flatfeestypes.NewMsgFee("/ibc.core.channel.v1.MsgChannelOpenTry", feeDefCoin(1500)),
		flatfeestypes.NewMsgFee("/ibc.core.connection.v1.MsgConnectionOpenConfirm", feeDefCoin(1500)),
		flatfeestypes.NewMsgFee("/ibc.core.connection.v1.MsgConnectionOpenTry", feeDefCoin(1500)),

		// Msgs that cost $2.50.
		flatfeestypes.NewMsgFee("/cosmos.gov.v1.MsgSubmitProposal", feeDefCoin(2500)),
		flatfeestypes.NewMsgFee("/cosmos.gov.v1beta1.MsgSubmitProposal", feeDefCoin(2500)),
		flatfeestypes.NewMsgFee("/provenance.marker.v1.MsgAddMarkerRequest", feeDefCoin(2500)),
	}
}
