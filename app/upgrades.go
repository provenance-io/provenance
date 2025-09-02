package app

import (
	"bytes"
	"compress/gzip"
	"context"
	"embed"
	"encoding/json"
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
	ledgerTypes "github.com/provenance-io/provenance/x/ledger/types"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
	registryTypes "github.com/provenance-io/provenance/x/registry/types"
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
	"bouvardia-rc1": { // Upgrade for v1.26.0-rc1.
		Added:   []string{flatfeestypes.StoreKey, registryTypes.StoreKey, ledgerTypes.StoreKey},
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
			if err = importLedgerData(ctx, app.LedgerKeeper); err != nil {
				return nil, err
			}
			return vm, nil
		},
	},
	"bouvardia": { // Upgrade for v1.26.0.
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
			if err = importLedgerData(ctx, app.LedgerKeeper); err != nil {
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
	for _, addr := range addrs {
		// We don't care about the error here because it's already logged, and we want to move on.
		_ = app.HoldKeeper.UnlockVestingAccount(ctx, addr)
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
// Part of the bouvardia upgrade.
type FlatFeesKeeper interface {
	SetParams(ctx sdk.Context, params flatfeestypes.Params) error
	SetMsgFee(ctx sdk.Context, msgFee flatfeestypes.MsgFee) error
}

// setupFlatFees defines the flatfees module params and msg costs.
// Part of the bouvardia upgrade.
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
// Part of the bouvardia upgrade.
func feeDefCoin(amount int64) sdk.Coin {
	return sdk.NewInt64Coin(flatfeestypes.DefaultFeeDefinitionDenom, amount)
}

// MakeFlatFeesParams returns the params to give the flatfeees module.
// Part of the bouvardia upgrade.
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
// Part of the bouvardia upgrade.
func MakeFlatFeesCosts() []*flatfeestypes.MsgFee {
	return []*flatfeestypes.MsgFee{
		// Free Msg types. These are gov-prop-only Msg types. A gov prop costs $2.00 + the cost of each msg in it.
		// So even though these msgs are free, it'll still cost $2.00 to submit one.
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
		flatfeestypes.NewMsgFee("/provenance.hold.v1.MsgUnlockVestingAccountsRequest"),
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

		// Msgs that cost $0.005.
		flatfeestypes.NewMsgFee("/provenance.metadata.v1.MsgAddNetAssetValuesRequest", feeDefCoin(5)),

		// Msgs that cost $0.01.
		flatfeestypes.NewMsgFee("/provenance.ledger.v1.MsgDestroyRequest", feeDefCoin(10)),
		flatfeestypes.NewMsgFee("/provenance.marker.v1.MsgRevokeGrantAllowanceRequest", feeDefCoin(10)),

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
		flatfeestypes.NewMsgFee("/cosmos.nft.v1beta1.MsgSend", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/ibc.core.channel.v1.MsgAcknowledgement", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/ibc.core.channel.v1.MsgRecvPacket", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/provenance.exchange.v1.MsgAcceptPaymentRequest", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/provenance.exchange.v1.MsgCommitFundsRequest", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/provenance.exchange.v1.MsgMarketReleaseCommitmentsRequest", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/provenance.exchange.v1.MsgMarketTransferCommitmentRequest", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/provenance.exchange.v1.MsgRejectPaymentRequest", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/provenance.exchange.v1.MsgRejectPaymentsRequest", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/provenance.ledger.v1.MsgAddClassBucketTypeRequest", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/provenance.ledger.v1.MsgAddClassEntryTypeRequest", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/provenance.ledger.v1.MsgAddClassStatusTypeRequest", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/provenance.ledger.v1.MsgUpdateBalancesRequest", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/provenance.ledger.v1.MsgUpdateInterestRateRequest", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/provenance.ledger.v1.MsgUpdateMaturityDateRequest", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/provenance.ledger.v1.MsgUpdatePaymentRequest", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/provenance.ledger.v1.MsgUpdateStatusRequest", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/provenance.marker.v1.MsgActivateRequest", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/provenance.marker.v1.MsgAddNetAssetValuesRequest", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/provenance.marker.v1.MsgFinalizeRequest", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/provenance.marker.v1.MsgGrantAllowanceRequest", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/provenance.marker.v1.MsgSetDenomMetadataRequest", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/provenance.registry.v1.MsgGrantRole", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/provenance.registry.v1.MsgRevokeRole", feeDefCoin(50)),
		flatfeestypes.NewMsgFee("/provenance.registry.v1.MsgUnregisterNFT", feeDefCoin(50)),

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
		flatfeestypes.NewMsgFee("/provenance.ledger.v1.MsgAppendRequest", feeDefCoin(100)),
		flatfeestypes.NewMsgFee("/provenance.ledger.v1.MsgTransferFundsWithSettlementRequest", feeDefCoin(100)),
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
		flatfeestypes.NewMsgFee("/provenance.asset.v1.MsgCreateAssetClass", feeDefCoin(500)),
		flatfeestypes.NewMsgFee("/provenance.metadata.v1.MsgMigrateValueOwnerRequest", feeDefCoin(500)),
		flatfeestypes.NewMsgFee("/provenance.metadata.v1.MsgUpdateValueOwnersRequest", feeDefCoin(500)),
		flatfeestypes.NewMsgFee("/provenance.metadata.v1.MsgWriteContractSpecificationRequest", feeDefCoin(500)),
		flatfeestypes.NewMsgFee("/provenance.metadata.v1.MsgWriteRecordSpecificationRequest", feeDefCoin(500)),
		flatfeestypes.NewMsgFee("/provenance.metadata.v1.MsgWriteScopeSpecificationRequest", feeDefCoin(500)),
		flatfeestypes.NewMsgFee("/provenance.metadata.v1.MsgWriteSessionRequest", feeDefCoin(500)),

		// Msgs that cost $1.00.
		flatfeestypes.NewMsgFee("/cosmos.group.v1.MsgSubmitProposal", feeDefCoin(1000)),
		flatfeestypes.NewMsgFee("/cosmwasm.wasm.v1.MsgStoreCode", feeDefCoin(1000)),
		flatfeestypes.NewMsgFee("/provenance.ledger.v1.MsgCreateRequest", feeDefCoin(1000)),
		flatfeestypes.NewMsgFee("/provenance.metadata.v1.MsgAddContractSpecToScopeSpecRequest", feeDefCoin(1000)),
		flatfeestypes.NewMsgFee("/provenance.metadata.v1.MsgDeleteContractSpecFromScopeSpecRequest", feeDefCoin(1000)),
		flatfeestypes.NewMsgFee("/provenance.metadata.v1.MsgWriteScopeRequest", feeDefCoin(1000)),

		// Msgs that cost $1.50.
		flatfeestypes.NewMsgFee("/ibc.core.channel.v1.MsgChannelOpenTry", feeDefCoin(1500)),
		flatfeestypes.NewMsgFee("/ibc.core.connection.v1.MsgConnectionOpenConfirm", feeDefCoin(1500)),
		flatfeestypes.NewMsgFee("/ibc.core.connection.v1.MsgConnectionOpenTry", feeDefCoin(1500)),

		// Msgs that cost $2.00.
		flatfeestypes.NewMsgFee("/cosmos.gov.v1.MsgSubmitProposal", feeDefCoin(2000)),
		flatfeestypes.NewMsgFee("/cosmos.gov.v1beta1.MsgSubmitProposal", feeDefCoin(2000)),

		// Msgs that cost $3.00.
		flatfeestypes.NewMsgFee("/provenance.asset.v1.MsgCreatePool", feeDefCoin(3000)),
		flatfeestypes.NewMsgFee("/provenance.asset.v1.MsgCreateSecuritization", feeDefCoin(3000)),
		flatfeestypes.NewMsgFee("/provenance.asset.v1.MsgCreateTokenization", feeDefCoin(3000)),
		flatfeestypes.NewMsgFee("/provenance.marker.v1.MsgAddFinalizeActivateMarkerRequest", feeDefCoin(3000)),
		flatfeestypes.NewMsgFee("/provenance.marker.v1.MsgAddMarkerRequest", feeDefCoin(3000)),

		// Msgs that cost $5.00.
		flatfeestypes.NewMsgFee("/provenance.ledger.v1.MsgCreateLedgerClassRequest", feeDefCoin(5000)),

		// Msgs that cost $17.00.
		flatfeestypes.NewMsgFee("/provenance.registry.v1.MsgRegisterNFT", feeDefCoin(17000)),

		// Msgs that cost $18.00.
		flatfeestypes.NewMsgFee("/provenance.asset.v1.MsgCreateAsset", feeDefCoin(18000)),
	}
}

//go:embed upgrade_data/*
var upgradeDataFS embed.FS

// LedgerKeeper has the ledger keeper methods needed for creating ledgers and entries.
type LedgerKeeper interface {
	ImportLedgerClasses(ctx sdk.Context, state *ledgerTypes.GenesisState)
	ImportLedgerClassEntryTypes(ctx sdk.Context, state *ledgerTypes.GenesisState)
	ImportLedgerClassStatusTypes(ctx sdk.Context, state *ledgerTypes.GenesisState)
	ImportLedgerClassBucketTypes(ctx sdk.Context, state *ledgerTypes.GenesisState)
	ImportLedgers(ctx sdk.Context, state *ledgerTypes.GenesisState)
	ImportLedgerEntries(ctx sdk.Context, state *ledgerTypes.GenesisState)
	ImportStoredSettlementInstructions(ctx sdk.Context, state *ledgerTypes.GenesisState)
}

// importLedgerData creates ledgers and entries from embedded genesis data using streaming.
func importLedgerData(ctx sdk.Context, lk LedgerKeeper) error {
	ctx.Logger().Info("Starting streaming import of ledger data.")

	// Process the gzipped genesis file using streaming
	if err := streamImportLedgerData(ctx, lk); err != nil {
		return fmt.Errorf("failed to stream import ledger data: %w", err)
	}

	ctx.Logger().Info("Completed streaming import of ledger data.")
	return nil
}

// streamImportLedgerData processes the gzipped genesis file using streaming for memory efficiency.
func streamImportLedgerData(ctx sdk.Context, lk LedgerKeeper) error {
	filePath := "upgrade_data/bouvardia_ledger_genesis.json.gz"

	// Read the gzipped file data
	data, err := upgradeDataFS.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Create gzip reader for streaming decompression.
	reader := bytes.NewReader(data)
	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader for %s: %w", filePath, err)
	}
	defer gzReader.Close()

	// Use JSON decoder for streaming JSON parsing.
	decoder := json.NewDecoder(gzReader)

	// Expect the start of a JSON object.
	token, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("failed to read JSON token from %s: %w", filePath, err)
	}

	if delim, ok := token.(json.Delim); !ok || delim != '{' {
		return fmt.Errorf("expected JSON object start '{' in %s, got %v", filePath, token)
	}

	// Process each field in the GenesisState object.
	for decoder.More() {
		// Get the field name.
		fieldToken, err := decoder.Token()
		if err != nil {
			return fmt.Errorf("failed to read field name from %s: %w", filePath, err)
		}

		fieldName, ok := fieldToken.(string)
		if !ok {
			return fmt.Errorf("expected field name string in %s, got %v", filePath, fieldToken)
		}

		// Process each field based on its name.
		if err := processGenesisField(ctx, lk, decoder, fieldName); err != nil {
			return fmt.Errorf("failed to process field %s in %s: %w", fieldName, filePath, err)
		}
	}

	// Expect the end of the JSON object.
	token, err = decoder.Token()
	if err != nil {
		return fmt.Errorf("failed to read JSON token from %s: %w", filePath, err)
	}

	if delim, ok := token.(json.Delim); !ok || delim != '}' {
		return fmt.Errorf("expected JSON object end '}' in %s, got %v", filePath, token)
	}

	return nil
}

// processGenesisField processes a single field from the GenesisState JSON.
func processGenesisField(ctx sdk.Context, lk LedgerKeeper, decoder *json.Decoder, fieldName string) error {
	switch fieldName {
	case "ledgerClasses":
		var ledgerClasses []ledgerTypes.LedgerClass
		if err := decoder.Decode(&ledgerClasses); err != nil {
			return fmt.Errorf("failed to decode ledgerClasses: %w", err)
		}
		genesis := &ledgerTypes.GenesisState{LedgerClasses: ledgerClasses}
		lk.ImportLedgerClasses(ctx, genesis)
		ctx.Logger().Info("Imported ledger classes", "count", len(ledgerClasses))

	case "ledgerClassEntryTypes":
		var entryTypes []ledgerTypes.GenesisLedgerClassEntryType
		if err := decoder.Decode(&entryTypes); err != nil {
			return fmt.Errorf("failed to decode ledgerClassEntryTypes: %w", err)
		}
		genesis := &ledgerTypes.GenesisState{LedgerClassEntryTypes: entryTypes}
		lk.ImportLedgerClassEntryTypes(ctx, genesis)
		ctx.Logger().Info("Imported ledger class entry types", "count", len(entryTypes))

	case "ledgerClassStatusTypes":
		var statusTypes []ledgerTypes.GenesisLedgerClassStatusType
		if err := decoder.Decode(&statusTypes); err != nil {
			return fmt.Errorf("failed to decode ledgerClassStatusTypes: %w", err)
		}
		genesis := &ledgerTypes.GenesisState{LedgerClassStatusTypes: statusTypes}
		lk.ImportLedgerClassStatusTypes(ctx, genesis)
		ctx.Logger().Info("Imported ledger class status types", "count", len(statusTypes))

	case "ledgerClassBucketTypes":
		var bucketTypes []ledgerTypes.GenesisLedgerClassBucketType
		if err := decoder.Decode(&bucketTypes); err != nil {
			return fmt.Errorf("failed to decode ledgerClassBucketTypes: %w", err)
		}
		genesis := &ledgerTypes.GenesisState{LedgerClassBucketTypes: bucketTypes}
		lk.ImportLedgerClassBucketTypes(ctx, genesis)
		ctx.Logger().Info("Imported ledger class bucket types", "count", len(bucketTypes))

	case "ledgers":
		var ledgers []ledgerTypes.GenesisLedger
		if err := decoder.Decode(&ledgers); err != nil {
			return fmt.Errorf("failed to decode ledgers: %w", err)
		}
		genesis := &ledgerTypes.GenesisState{Ledgers: ledgers}
		lk.ImportLedgers(ctx, genesis)
		ctx.Logger().Info("Imported ledgers", "count", len(ledgers))

	case "ledgerEntries":
		var entries []ledgerTypes.GenesisLedgerEntry
		if err := decoder.Decode(&entries); err != nil {
			return fmt.Errorf("failed to decode ledgerEntries: %w", err)
		}
		genesis := &ledgerTypes.GenesisState{LedgerEntries: entries}
		lk.ImportLedgerEntries(ctx, genesis)
		ctx.Logger().Info("Imported ledger entries", "count", len(entries))

	case "settlementInstructions":
		var settlements []ledgerTypes.GenesisStoredSettlementInstructions
		if err := decoder.Decode(&settlements); err != nil {
			return fmt.Errorf("failed to decode settlementInstructions: %w", err)
		}
		genesis := &ledgerTypes.GenesisState{SettlementInstructions: settlements}
		lk.ImportStoredSettlementInstructions(ctx, genesis)
		ctx.Logger().Info("Imported settlement instructions", "count", len(settlements))

	default:
		// Skip unknown fields by decoding and discarding them
		var value interface{}
		if err := decoder.Decode(&value); err != nil {
			return fmt.Errorf("failed to skip unknown field %s: %w", fieldName, err)
		}
		ctx.Logger().Info("Skipped unknown field", "field", fieldName)
	}

	return nil
}
