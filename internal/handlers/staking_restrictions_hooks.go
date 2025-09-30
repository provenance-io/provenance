package handlers

import (
	"context"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/provenance-io/provenance/internal/pioconfig"
)

// Assert proper implementation of StakingRestrictions
var _ stakingtypes.StakingHooks = StakingRestrictionHooks{}

const (
	// DefaultConcentrationMultiple is the default concentration of bonded tokens a validator is allowed as a multiple of equal shares.
	DefaultConcentrationMultiple = 5.5 // Amounts to ~8% with 68 validators
	// DefaultMaxCapPercent is the maximum cap for bonded stake of any single validator.
	DefaultMaxCapPercent = 0.33
	// DefaultMinCapPercent is the minimum cap for bonded stake of any single validator.
	DefaultMinCapPercent = 0.05
)

// RestrictionOptions contains configurable fields related to how the staking restrictions apply.
type RestrictionOptions struct {
	MaxConcentrationMultiple float64
	MaxBondedCapPercent      float64
	MinBondedCapPercent      float64
}

// DefaultRestrictionOptions are default constraints that prevent single point of failure on validators
var DefaultRestrictionOptions = &RestrictionOptions{
	MaxConcentrationMultiple: DefaultConcentrationMultiple,
	MaxBondedCapPercent:      DefaultMaxCapPercent,
	MinBondedCapPercent:      DefaultMinCapPercent,
}

// UnlimitedRestrictionOptions are used to remove restrictions for validator staking limits from delegations
var UnlimitedRestrictionOptions = &RestrictionOptions{
	MaxConcentrationMultiple: 1.0,
	MaxBondedCapPercent:      1.0,
	MinBondedCapPercent:      1.0,
}

// CalcMaxValPct returns the maximum percent (of total bond) a single validator is allowed to have.
func (o RestrictionOptions) CalcMaxValPct(valCount int) float64 {
	rv := o.MaxConcentrationMultiple / float64(valCount)
	if rv >= o.MaxBondedCapPercent {
		return o.MaxBondedCapPercent // This gets returned if valCount == 0.
	}
	if rv <= o.MinBondedCapPercent {
		return o.MinBondedCapPercent
	}
	return rv
}

// CalcMaxValBond calculates the maximum bond allowed for a single validator based on total bonded tokens and max percent.
func CalcMaxValBond(totalBond sdkmath.Int, maxValPct float64) sdkmath.Int {
	// maxValPct is expected to be between 0.05 and 0.33. At 100 validators it will be 0.055.
	// The * 1_000_000 then / by same, essentially tells it to use 6 digits of precision from maxValPct.
	return totalBond.MulRaw(int64(maxValPct * 1_000_000)).QuoRaw(1_000_000)
}

type StakingKeeper interface {
	GetLastValidators(ctx context.Context) (validators []stakingtypes.Validator, err error)
	GetValidator(ctx context.Context, valAddr sdk.ValAddress) (validator stakingtypes.Validator, err error)
	GetLastValidatorPower(ctx context.Context, valAddr sdk.ValAddress) (power int64, err error)
	PowerReduction(ctx context.Context) sdkmath.Int
	TotalBondedTokens(ctx context.Context) (total sdkmath.Int, err error)
}

// StakingRestrictionHooks wrapper struct for staking keeper.
type StakingRestrictionHooks struct {
	k    StakingKeeper
	opts RestrictionOptions
}

// NewStakingRestrictionHooks configures a hook that validates changes to delegation modifications and
// prevents concentration of voting power beyond configured limits on active validators.
func NewStakingRestrictionHooks(k StakingKeeper, opts RestrictionOptions) StakingRestrictionHooks {
	return StakingRestrictionHooks{k, opts}
}

// AfterDelegationModified verifies that the delegation would not cause the validator's voting power to exceed our staking distribution limits.
func (h StakingRestrictionHooks) AfterDelegationModified(ctx context.Context, _ sdk.AccAddress, valAddr sdk.ValAddress) error {
	vals, _ := h.k.GetLastValidators(ctx) // Ignoring error here to treat it as zero validators.
	valCount := len(vals)

	// do not bother with limits on networks this small (or under simulation).
	if valCount < 4 || sdk.UnwrapSDKContext(ctx).ChainID() == pioconfig.SimAppChainID {
		return nil
	}

	validator, err := h.k.GetValidator(ctx, valAddr)
	if err != nil {
		// If we couldn't get the validator, there's nothing we can actually check in here. Since we don't
		// know that the validator is over the limit (or how things got this way), we let this happen.
		return nil
	}

	// If the power of this validator is not increasing, we don't need to check anything.
	oldPower, _ := h.k.GetLastValidatorPower(ctx, valAddr) // Ignoring error here to treat the old power as zero.
	newPower := validator.GetConsensusPower(h.k.PowerReduction(ctx))
	if newPower <= oldPower {
		return nil
	}

	totalBond, _ := h.k.TotalBondedTokens(ctx) // Ignoring error here to treat it as zero.
	maxValPct := h.opts.CalcMaxValPct(valCount)
	maxBond := CalcMaxValBond(totalBond, maxValPct)
	newBond := validator.GetBondedTokens()

	// If the new bond amount is over the max bonded token amount then we error out this transaction.
	if newBond.GT(maxBond) {
		return sdkerrors.ErrInvalidRequest.Wrapf(
			"validator bonded tokens of %s exceeds max of %s (= %.2f%% of %s total across %d validators)",
			newBond, maxBond, maxValPct*100, totalBond, valCount)
	}

	return nil
}

// BeforeDelegationCreated implements sdk.ValidatorHooks; does nothing.
func (h StakingRestrictionHooks) BeforeDelegationCreated(_ context.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	return nil
}

// AfterValidatorBonded implements sdk.ValidatorHooks; does nothing.
func (h StakingRestrictionHooks) AfterValidatorBonded(_ context.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

// AfterValidatorRemoved implements sdk.ValidatorHooks; does nothing.
func (h StakingRestrictionHooks) AfterValidatorRemoved(_ context.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

// AfterValidatorCreated implements sdk.ValidatorHooks; does nothing.
func (h StakingRestrictionHooks) AfterValidatorCreated(_ context.Context, _ sdk.ValAddress) error {
	return nil
}

// AfterValidatorBeginUnbonding implements sdk.ValidatorHooks; does nothing.
func (h StakingRestrictionHooks) AfterValidatorBeginUnbonding(_ context.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

// BeforeValidatorModified implements sdk.ValidatorHooks; does nothing.
func (h StakingRestrictionHooks) BeforeValidatorModified(_ context.Context, _ sdk.ValAddress) error {
	return nil
}

// BeforeDelegationSharesModified implements sdk.ValidatorHooks; does nothing.
func (h StakingRestrictionHooks) BeforeDelegationSharesModified(_ context.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	return nil
}

// BeforeDelegationRemoved implements sdk.ValidatorHooks; does nothing.
func (h StakingRestrictionHooks) BeforeDelegationRemoved(_ context.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	return nil
}

// BeforeValidatorSlashed implements sdk.ValidatorHooks; does nothing.
func (h StakingRestrictionHooks) BeforeValidatorSlashed(_ context.Context, _ sdk.ValAddress, _ sdkmath.LegacyDec) error {
	return nil
}

// AfterUnbondingInitiated implements sdk.ValidatorHooks; does nothing.
func (h StakingRestrictionHooks) AfterUnbondingInitiated(_ context.Context, _ uint64) error {
	return nil
}
