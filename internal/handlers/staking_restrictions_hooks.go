package handlers

import (
	"cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Hooks wrapper struct for slashing keeper
type StakingRestrictionHooks struct {
	k *stakingkeeper.Keeper
}

const (
	// The concentration of bonded tokens a validator is allowed as a multiple of equal shares
	MaxConcentration = 8.0
	// Maximum Allowed Cap for Bonded stake of any single validator
	MaxBondedCapPercent = 0.33
	// Minimum Allowed Cap for Bonded stake of any single validator
	MinBondedCapPercent = 0.05
)

var _ stakingtypes.StakingHooks = StakingRestrictionHooks{}

func NewStakingRestrictionHooks(k *stakingkeeper.Keeper) StakingRestrictionHooks {
	return StakingRestrictionHooks{k}
}

// Verifies that the delegation would not cause the validator's voting power to exceed our staking distribution limits
func (h StakingRestrictionHooks) AfterDelegationModified(ctx sdktypes.Context, delAddr sdktypes.AccAddress, valAddr sdktypes.ValAddress) error {

	valCount := len(h.k.GetLastValidators(ctx))

	// bond limit is allowed to have a multiple of even shares of network bonded stake.
	maxValidatorPercent := MaxConcentration * (1.0 / float64(valCount))

	// do not bother with limits on networks this small.
	if valCount < 4 {
		return nil
	}

	// check the capped bond amount is within the overall range limits.
	if maxValidatorPercent > MaxBondedCapPercent {
		maxValidatorPercent = MaxBondedCapPercent
	} else if maxValidatorPercent < MinBondedCapPercent {
		maxValidatorPercent = MinBondedCapPercent
	}

	oldPower := h.k.GetLastValidatorPower(ctx, valAddr)
	validator, found := h.k.GetValidator(ctx, valAddr)
	if found {
		power := validator.GetConsensusPower(h.k.PowerReduction(ctx))
		currentBond := h.k.TotalBondedTokens(ctx)
		maxBond := currentBond.Quo(math.NewInt(100)).MulRaw(int64(maxValidatorPercent * 100))

		// if the power of this validator is increasing and it is over the maximum bonded token amount then we error out this transaction.
		if power > oldPower && validator.GetBondedTokens().GT(maxBond) {
			return sdkerrors.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"validator bonded tokens of %d exceeds max of %d (%.1f%%) for %d validators",
				currentBond.Int64(),
				maxBond.Int64(),
				maxValidatorPercent*100,
				valCount,
			)
		}
	}

	return nil
}

// Implements sdktypes.ValidatorHooks
func (h StakingRestrictionHooks) BeforeDelegationCreated(ctx sdktypes.Context, _ sdktypes.AccAddress, _ sdktypes.ValAddress) error {
	return nil
}

// Implements sdktypes.ValidatorHooks
func (h StakingRestrictionHooks) AfterValidatorBonded(_ sdktypes.Context, _ sdktypes.ConsAddress, _ sdktypes.ValAddress) error {
	return nil
}

// Implements sdktypes.ValidatorHooks
func (h StakingRestrictionHooks) AfterValidatorRemoved(_ sdktypes.Context, _ sdktypes.ConsAddress, _ sdktypes.ValAddress) error {
	return nil
}

// Implements sdktypes.ValidatorHooks
func (h StakingRestrictionHooks) AfterValidatorCreated(_ sdktypes.Context, valAddr sdktypes.ValAddress) error {
	return nil
}

// Implements sdktypes.ValidatorHooks
func (h StakingRestrictionHooks) AfterValidatorBeginUnbonding(_ sdktypes.Context, _ sdktypes.ConsAddress, _ sdktypes.ValAddress) error {
	return nil
}

// Implements sdktypes.ValidatorHooks
func (h StakingRestrictionHooks) BeforeValidatorModified(_ sdktypes.Context, _ sdktypes.ValAddress) error {
	return nil
}

// Implements sdktypes.ValidatorHooks
func (h StakingRestrictionHooks) BeforeDelegationSharesModified(_ sdktypes.Context, _ sdktypes.AccAddress, _ sdktypes.ValAddress) error {
	return nil
}

// Implements sdktypes.ValidatorHooks
func (h StakingRestrictionHooks) BeforeDelegationRemoved(_ sdktypes.Context, _ sdktypes.AccAddress, _ sdktypes.ValAddress) error {
	return nil
}

// Implements sdktypes.ValidatorHooks
func (h StakingRestrictionHooks) BeforeValidatorSlashed(_ sdktypes.Context, _ sdktypes.ValAddress, _ sdktypes.Dec) error {
	return nil
}
