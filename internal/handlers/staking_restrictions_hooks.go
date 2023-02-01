package handlers

import (
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Assert proper implementation of StakingRestrictions
var _ stakingtypes.StakingHooks = StakingRestrictionHooks{}

const (
	// The default concentration of bonded tokens a validator is allowed as a multiple of equal shares
	DefaultConcentration = 5.5 // Amounts to ~8% with 68 validators
	// Maximum Allowed Cap for Bonded stake of any single validator
	DefaultMaxCapPercent = 0.33
	// Minimum Allowed Cap for Bonded stake of any single validator
	DefaultMinCapPercent = 0.05
)

type RestrictionOptions struct {
	MaxConcentrationMultiple float32
	MaxBondedCapPercent      float32
	MinBondedCapPercent      float32
}

// DefaultRestictionOptions are default constraints that prevent single point of failure on validators
var DefaultRestictionOptions = &RestrictionOptions{
	MaxConcentrationMultiple: DefaultConcentration,
	MaxBondedCapPercent:      DefaultMaxCapPercent,
	MinBondedCapPercent:      DefaultMinCapPercent,
}

// UnlimitedRestrictionOptions are used to remove restrictions for validator staking limits from delegations
var UnlimitedRestrictionOptions = &RestrictionOptions{
	MaxConcentrationMultiple: 1.0,
	MaxBondedCapPercent:      1.0,
	MinBondedCapPercent:      1.0,
}

// Hooks wrapper struct for slashing keeper
type StakingRestrictionHooks struct {
	k    *stakingkeeper.Keeper
	opts RestrictionOptions
}

// NewStakingRestrictionHooks configures a hook that validates changes to delegation modifications and
// prevents concentration of voting power beyond configured limits on active validators.
func NewStakingRestrictionHooks(k *stakingkeeper.Keeper, opts RestrictionOptions) StakingRestrictionHooks {
	return StakingRestrictionHooks{k, opts}
}

// Verifies that the delegation would not cause the validator's voting power to exceed our staking distribution limits
func (h StakingRestrictionHooks) AfterDelegationModified(ctx sdktypes.Context, delAddr sdktypes.AccAddress, valAddr sdktypes.ValAddress) error {
	valCount := len(h.k.GetLastValidators(ctx))

	// do not bother with limits on networks this small (or under simulation).
	if valCount < 4 || ctx.ChainID() == helpers.SimAppChainID {
		return nil
	}

	// bond limit is allowed to have a multiple of even shares of network bonded stake.
	maxValidatorPercent := h.opts.MaxConcentrationMultiple * (1.0 / float32(valCount))

	// check the capped bond amount is within the overall range limits.
	if maxValidatorPercent > h.opts.MaxBondedCapPercent {
		maxValidatorPercent = h.opts.MaxBondedCapPercent
	} else if maxValidatorPercent < h.opts.MinBondedCapPercent {
		maxValidatorPercent = h.opts.MinBondedCapPercent
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
				currentBond.BigInt(),
				maxBond.BigInt(),
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
