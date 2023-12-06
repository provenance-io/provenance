package rc1

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	provenance "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/app/upgrades"
	attributekeeper "github.com/provenance-io/provenance/x/attribute/keeper"
	attributetypes "github.com/provenance-io/provenance/x/attribute/types"
	msgfeetypes "github.com/provenance-io/provenance/x/msgfees/types"
)

func UpgradeStrategy(ctx sdk.Context, app *provenance.App, vm module.VersionMap) (module.VersionMap, error) {
	// Migrate all the modules
	newVM, err := upgrades.RunModuleMigrations(ctx, app, vm)
	if err != nil {
		return nil, err
	}

	RemoveInactiveValidatorDelegations(ctx, app)

	err = SetAccountDataNameRecord(ctx, app.AccountKeeper, &app.NameKeeper)
	if err != nil {
		return nil, err
	}

	// We only need to call addGovV1SubmitFee on testnet.
	AddGovV1SubmitFee(ctx, app)

	RemoveP8eMemorializeContractFee(ctx, app)

	FixNameIndexEntries(ctx, app)

	return newVM, nil
}

// removeInactiveValidatorDelegations unbonds all delegations from inactive validators, triggering their removal from the validator set.
// This should be applied in most upgrades.
func RemoveInactiveValidatorDelegations(ctx sdk.Context, app *provenance.App) {
	unbondingTimeParam := app.StakingKeeper.GetParams(ctx).UnbondingTime
	ctx.Logger().Info(fmt.Sprintf("removing all delegations from validators that have been inactive (unbonded) for %d days", int64(unbondingTimeParam.Hours()/24)))
	removalCount := 0
	validators := app.StakingKeeper.GetAllValidators(ctx)
	for _, validator := range validators {
		if validator.IsUnbonded() {
			inactiveDuration := ctx.BlockTime().Sub(validator.UnbondingTime)
			if inactiveDuration >= unbondingTimeParam {
				ctx.Logger().Info(fmt.Sprintf("validator %v has been inactive (unbonded) for %d days and will be removed", validator.OperatorAddress, int64(inactiveDuration.Hours()/24)))
				valAddress, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
				if err != nil {
					ctx.Logger().Error(fmt.Sprintf("invalid operator address: %s: %v", validator.OperatorAddress, err))
					continue
				}
				delegations := app.StakingKeeper.GetValidatorDelegations(ctx, valAddress)
				for _, delegation := range delegations {
					ctx.Logger().Info(fmt.Sprintf("undelegate delegator %v from validator %v of all shares (%v)", delegation.DelegatorAddress, validator.OperatorAddress, delegation.GetShares()))
					_, err = app.StakingKeeper.Undelegate(ctx, delegation.GetDelegatorAddr(), valAddress, delegation.GetShares())
					if err != nil {
						ctx.Logger().Error(fmt.Sprintf("failed to undelegate delegator %s from validator %s: %v", delegation.GetDelegatorAddr().String(), valAddress.String(), err))
						continue
					}
				}
				removalCount++
			}
		}
	}
	ctx.Logger().Info(fmt.Sprintf("a total of %d inactive (unbonded) validators have had all their delegators removed", removalCount))
}

// setAccountDataNameRecord makes sure the account data name record exists, is restricted,
// and is owned by the attribute module. An error is returned if it fails to make it so.
// TODO: Remove with the rust handlers.
func SetAccountDataNameRecord(ctx sdk.Context, accountK attributetypes.AccountKeeper, nameK attributetypes.NameKeeper) (err error) {
	return attributekeeper.EnsureModuleAccountAndAccountDataNameRecord(ctx, accountK, nameK)
}

// removeP8eMemorializeContractFee removes the message fee for the now-non-existent MsgP8eMemorializeContractRequest.
// TODO: Remove with the rust handlers.
func RemoveP8eMemorializeContractFee(ctx sdk.Context, app *provenance.App) {
	typeURL := "/provenance.metadata.v1.MsgP8eMemorializeContractRequest"

	ctx.Logger().Info(fmt.Sprintf("Removing message fee for %q if one exists.", typeURL))
	// Get the existing fee for log output, but ignore any errors so we try to delete the entry either way.
	fee, _ := app.MsgFeesKeeper.GetMsgFee(ctx, typeURL)
	// At the time of writing this, the only error that RemoveMsgFee can return is ErrMsgFeeDoesNotExist.
	// So ignore any error here and just use fee != nil for the different log messages.
	_ = app.MsgFeesKeeper.RemoveMsgFee(ctx, typeURL)
	if fee == nil {
		ctx.Logger().Info(fmt.Sprintf("Message fee for %q already does not exist. Nothing to do.", typeURL))
	} else {
		ctx.Logger().Info(fmt.Sprintf("Successfully removed message fee for %q with amount %q.", fee.MsgTypeUrl, fee.AdditionalFee.String()))
	}
}

// fixNameIndexEntries fixes the name module's address to name index entries.
// TODO: Remove with the rust handlers.
func FixNameIndexEntries(ctx sdk.Context, app *provenance.App) {
	ctx.Logger().Info("Fixing name module store index entries.")
	app.NameKeeper.DeleteInvalidAddressIndexEntries(ctx)
	ctx.Logger().Info("Done fixing name module store index entries.")
}

// addGovV1SubmitFee adds a msg-fee for the gov v1 MsgSubmitProposal if there isn't one yet.
// TODO: Remove with the rust handlers.
func AddGovV1SubmitFee(ctx sdk.Context, app *provenance.App) {
	typeURL := sdk.MsgTypeURL(&govtypesv1.MsgSubmitProposal{})

	ctx.Logger().Info(fmt.Sprintf("Creating message fee for %q if it doesn't already exist.", typeURL))
	// At the time of writing this, the only way GetMsgFee returns an error is if it can't unmarshall state.
	// If that's the case for the v1 entry, we want to fix it anyway, so we just ignore any error here.
	fee, _ := app.MsgFeesKeeper.GetMsgFee(ctx, typeURL)
	// If there's already a fee for it, do nothing.
	if fee != nil {
		ctx.Logger().Info(fmt.Sprintf("Message fee for %q already exists with amount %q. Nothing to do.", fee.MsgTypeUrl, fee.AdditionalFee.String()))
		return
	}

	// Copy the fee from the beta entry if it exists, otherwise, just make it fresh.
	betaTypeURL := sdk.MsgTypeURL(&govtypesv1beta1.MsgSubmitProposal{})
	// Here too, if there's an error getting the beta fee, just ignore it.
	betaFee, _ := app.MsgFeesKeeper.GetMsgFee(ctx, betaTypeURL)
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

	// At the time of writing this, SetMsgFee always returns nil.
	_ = app.MsgFeesKeeper.SetMsgFee(ctx, *fee)
	ctx.Logger().Info(fmt.Sprintf("Successfully set fee for %q with amount %q.", fee.MsgTypeUrl, fee.AdditionalFee.String()))
}
