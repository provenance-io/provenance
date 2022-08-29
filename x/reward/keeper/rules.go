package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	"github.com/provenance-io/provenance/x/reward/types"
)

func (k Keeper) ProcessTransactions(ctx sdk.Context) {
	// Get all Active Reward Programs
	rewardPrograms, err := k.GetAllActiveRewardPrograms(ctx)
	if err != nil {
		ctx.Logger().Error(err.Error())
		return
	}

	// Grant shares for qualifying actions
	for index := range rewardPrograms {
		// Go through all the reward programs
		actions, err := k.DetectQualifyingActions(ctx, &rewardPrograms[index])
		if err != nil {
			ctx.Logger().Error(err.Error())
			continue
		}

		// Record any results
		err = k.RewardShares(ctx, &rewardPrograms[index], actions)
		if err != nil {
			ctx.Logger().Error(err.Error())
		}
	}
}

// DetectQualifyingActions takes in the RewardProgram and checks if any of the qualifying actions is found in the event history
func (k Keeper) DetectQualifyingActions(ctx sdk.Context, program *types.RewardProgram) ([]types.EvaluationResult, error) {
	ctx.Logger().Info(fmt.Sprintf("EvaluateRules for RewardProgram: %d", program.GetId()))
	results := []types.EvaluationResult(nil)

	// Check if any of the transactions are qualifying actions
	for _, supportedAction := range program.GetQualifyingActions() {
		// Get the appropriate RewardAction
		// If it's not supported we skip it
		action, err := supportedAction.GetRewardAction(ctx)
		if err != nil {
			ctx.Logger().Info(err.Error())
			continue
		}

		// Build all the qualifying actions from the abci events
		actions, err := k.FindQualifyingActions(ctx, action)
		if err != nil {
			return nil, err
		}

		// Process actions and get the results
		actions = k.ProcessQualifyingActions(ctx, program, action, actions)
		results = append(results, actions...)
	}

	return results, nil
}

func (k Keeper) ProcessQualifyingActions(ctx sdk.Context, program *types.RewardProgram, processor types.RewardAction, actions []types.EvaluationResult) []types.EvaluationResult {
	successfulActions := []types.EvaluationResult(nil)
	if program == nil || processor == nil || actions == nil {
		return successfulActions
	}

	for _, action := range actions {
		state, err := k.GetRewardAccountState(ctx, program.GetId(), program.GetCurrentClaimPeriod(), action.Address.String())
		if err != nil {
			continue
		}
		if state.Validate() != nil {
			state = types.NewRewardAccountState(program.GetId(), program.GetCurrentClaimPeriod(), action.Address.String(), 0, map[string]uint64{})
		}

		if !processor.PreEvaluate(ctx, k, state) {
			k.SetRewardAccountState(ctx, state)
			continue
		}
		if !processor.Evaluate(ctx, k, state, action) {
			k.SetRewardAccountState(ctx, state)
			continue
		}
		state.ActionCounter[processor.ActionType()]++
		if !processor.PostEvaluate(ctx, k, state) {
			k.SetRewardAccountState(ctx, state)
			continue
		}

		successfulActions = append(successfulActions, action)
		k.SetRewardAccountState(ctx, state)
	}

	return successfulActions
}

func (k Keeper) RewardShares(ctx sdk.Context, rewardProgram *types.RewardProgram, evaluateRes []types.EvaluationResult) error {
	ctx.Logger().Info(fmt.Sprintf("Recording shares for for rewardProgramId=%d, claimPeriod=%d",
		rewardProgram.GetId(), rewardProgram.GetCurrentClaimPeriod()))

	if rewardProgram == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "reward program cannot be nil")
	}

	// get the ClaimPeriodRewardDistribution
	claimPeriodRewardDistribution, err := k.GetClaimPeriodRewardDistribution(ctx, rewardProgram.GetCurrentClaimPeriod(), rewardProgram.GetId())

	if err != nil {
		return err
	}

	if claimPeriodRewardDistribution.Validate() != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "invalid claim period reward distribution.")
	}

	for _, res := range evaluateRes {
		state, err := k.GetRewardAccountState(ctx, rewardProgram.GetId(), rewardProgram.GetCurrentClaimPeriod(), res.Address.String())
		if state.Validate() != nil {
			ctx.Logger().Error(fmt.Sprintf("Account state does not exist for RewardProgram: %d, ClaimPeriod: %d, Address: %s. Skipping...",
				rewardProgram.GetId(), rewardProgram.GetCurrentClaimPeriod(), res.Address.String()))
			continue
		}
		if err != nil {
			return err
		}

		state.SharesEarned += uint64(res.Shares)
		k.SetRewardAccountState(ctx, state)
		// we know the rewards, so update the claim period reward
		claimPeriodRewardDistribution.TotalShares += res.Shares
	}

	// set total claim period rewards distribution.
	k.SetClaimPeriodRewardDistribution(ctx, claimPeriodRewardDistribution)

	return nil
}

// IterateABCIEvents Iterates through all the ABCIEvents that match the eventCriteria.
// Nil criteria means to iterate over everything.
func (k Keeper) IterateABCIEvents(ctx sdk.Context, criteria *types.EventCriteria, action func(string, *map[string][]byte) error) error {
	for _, event := range ctx.EventManager().GetABCIEventHistory() {
		event := event
		ctx.Logger().Info(fmt.Sprintf("events type is %s", event.Type))

		// Event type must match the criteria
		// nil criteria is considered to match everything
		if criteria != nil && !criteria.MatchesEvent(event.Type) {
			continue
		}

		// Convert the attributes into a map
		attributes := make(map[string][]byte)
		for _, attribute := range event.Attributes {
			attributes[string(attribute.Key)] = attribute.Value
		}

		valid := true
		if criteria != nil {
			// Ensure each attribute matches the required criteria
			// If a single attribute does not match then we don't continue with the event
			eventCriteria := criteria.Events[event.Type]
			for key := range eventCriteria.Attributes {
				valid = eventCriteria.MatchesAttribute(key, attributes[key])
				if !valid {
					break
				}
			}
		}
		if !valid {
			continue
		}

		err := action(event.Type, &attributes)
		if err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) FindQualifyingActions(ctx sdk.Context, action types.RewardAction) ([]types.EvaluationResult, error) {
	result := ([]types.EvaluationResult)(nil)
	builder := action.GetBuilder()

	err := k.IterateABCIEvents(ctx, builder.GetEventCriteria(), func(eventType string, attributes *map[string][]byte) error {
		// Add the event to the builder
		err := builder.AddEvent(eventType, attributes)
		if err != nil {
			return err
		}

		// Not finished building skip attempting to build
		if !builder.CanBuild() {
			return nil
		}

		// Attempt to build
		action, err := builder.BuildAction()
		if err != nil {
			return err
		}
		result = append(result, action)

		builder.Reset()

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (k Keeper) GetAccountKeeper() types.AccountKeeper {
	return k.authkeeper
}

func (k Keeper) GetStakingKeeper() types.StakingKeeper {
	return k.stakingKeeper
}

// SetStakingKeeper only used in tests
func (k *Keeper) SetStakingKeeper(newKeeper types.StakingKeeper) {
	k.stakingKeeper = newKeeper
}

// SetAccountKeeper only used in tests
func (k *Keeper) SetAccountKeeper(newKeeper authkeeper.AccountKeeper) {
	k.authkeeper = newKeeper
}
