package rc1

import (
	"fmt"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	icqtypes "github.com/cosmos/ibc-apps/modules/async-icq/v6/types"

	"github.com/provenance-io/provenance/app/keepers"
	"github.com/provenance-io/provenance/app/upgrades"
	"github.com/provenance-io/provenance/x/exchange"
	ibchookstypes "github.com/provenance-io/provenance/x/ibchooks/types"
)

func UpgradeStrategy(ctx sdk.Context, app upgrades.AppUpgrader, vm module.VersionMap) (module.VersionMap, error) {
	// Migrate all the modules
	newVM, err := upgrades.RunModuleMigrations(ctx, app, vm)
	if err != nil {
		return nil, err
	}

	if err = PerformUpgrade(ctx, app.Keepers()); err != nil {
		return nil, err
	}

	return newVM, nil
}

func PerformUpgrade(ctx sdk.Context, k *keepers.AppKeepers) error {
	// set ibchoooks defaults (no allowed async contracts)
	k.IBCHooksKeeper.SetParams(ctx, ibchookstypes.DefaultParams())

	RemoveInactiveValidatorDelegations(ctx, k)
	SetupICQ(ctx, k)
	UpdateMaxSupply(ctx, k)
	SetExchangeParams(ctx, k)
	return nil
}

// removeInactiveValidatorDelegations unbonds all delegations from inactive validators, triggering their removal from the validator set.
// This should be applied in most upgrades.
func RemoveInactiveValidatorDelegations(ctx sdk.Context, k *keepers.AppKeepers) {
	unbondingTimeParam := k.StakingKeeper.GetParams(ctx).UnbondingTime
	ctx.Logger().Info(fmt.Sprintf("removing all delegations from validators that have been inactive (unbonded) for %d days", int64(unbondingTimeParam.Hours()/24)))
	removalCount := 0
	validators := k.StakingKeeper.GetAllValidators(ctx)
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
				delegations := k.StakingKeeper.GetValidatorDelegations(ctx, valAddress)
				for _, delegation := range delegations {
					ctx.Logger().Info(fmt.Sprintf("undelegate delegator %v from validator %v of all shares (%v)", delegation.DelegatorAddress, validator.OperatorAddress, delegation.GetShares()))
					_, err = k.StakingKeeper.Undelegate(ctx, delegation.GetDelegatorAddr(), valAddress, delegation.GetShares())
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

// setupICQ sets the correct default values for ICQKeeper.
// TODO: Remove with the saffron handlers.
func SetupICQ(ctx sdk.Context, k *keepers.AppKeepers) {
	ctx.Logger().Info("Updating ICQ params")
	k.ICQKeeper.SetParams(ctx, icqtypes.NewParams(true, []string{"/provenance.oracle.v1.Query/Oracle"}))
	ctx.Logger().Info("Done updating ICQ params")
}

// updateMaxSupply sets the value of max supply to the current value of MaxTotalSupply.
// TODO: Remove with the saffron handlers.
func UpdateMaxSupply(ctx sdk.Context, k *keepers.AppKeepers) {
	ctx.Logger().Info("Updating MaxSupply marker param")
	params := k.MarkerKeeper.GetParams(ctx)
	// Populate new param with deprecated param

	params.MaxSupply = math.NewIntFromUint64(params.MaxTotalSupply)
	k.MarkerKeeper.SetParams(ctx, params)
	ctx.Logger().Info("Done updating MaxSupply marker param")
}

// setExchangeParams sets exchange module's params to the defaults.
// TODO: Remove with the saffron handlers.
func SetExchangeParams(ctx sdk.Context, k *keepers.AppKeepers) {
	ctx.Logger().Info("Ensuring exchange module params are set.")
	params := k.ExchangeKeeper.GetParams(ctx)
	if params != nil {
		ctx.Logger().Info("Exchange module params are already defined.")
	} else {
		params = exchange.DefaultParams()
		ctx.Logger().Info("Setting exchange module params to defaults.")
		k.ExchangeKeeper.SetParams(ctx, params)
	}
	ctx.Logger().Info("Done ensuring exchange module params are set.")
}
