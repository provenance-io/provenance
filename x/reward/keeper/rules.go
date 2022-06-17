package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

func (k Keeper) ProcessTransactions(ctx sdk.Context) {
	logEvents(ctx)

	// Get all Active Reward Programs
	rewardPrograms, err := k.GetAllActiveRewardPrograms(ctx)
	if err != nil {
		ctx.Logger().Error(err.Error())
		return
	}

	// Grant shares for qualifying actions
	for _, program := range rewardPrograms {
		// Go through all the reward programs
		program := program
		actions, err := k.DetectQualifyingActions(ctx, &program)
		if err != nil {
			ctx.Logger().Error(err.Error())
			continue
		}

		// Record any results
		err = k.RewardShares(ctx, &program, actions)
		if err != nil {
			ctx.Logger().Error(err.Error())
		}
	}
}

// EvaluateRules takes in a Eligibility criteria and measure it against the events in the context
func (k Keeper) DetectQualifyingActions(ctx sdk.Context, program *types.RewardProgram) ([]types.EvaluationResult, error) {
	ctx.Logger().Info(fmt.Sprintf("NOTICE: EvaluateRules for RewardProgram: %d", program.GetId()))
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
		// Get the AccountState for the triggering account
		var state types.AccountState

		// TODO Move this into the action's Pre-evaluate
		state, err := k.GetAccountState(ctx, program.GetId(), program.GetCurrentSubPeriod(), action.Address.String())
		if err != nil {
			continue
		}
		state.ActionCounter++
		k.SetAccountState(ctx, &state)

		// TODO We want to create an Evaluation Result here.
		// TODO We can get the share amount from the action
		if processor.Evaluate(ctx, k, state, action) {
			successfulActions = append(successfulActions, action)
		}

		// TODO Do a PostEvaluate
	}

	return successfulActions
}

func (k Keeper) RewardShares(ctx sdk.Context, rewardProgram *types.RewardProgram, evaluateRes []types.EvaluationResult) error {
	ctx.Logger().Info(fmt.Sprintf("NOTICE: Recording shares for for rewardProgramId=%d, subPeriod=%d",
		rewardProgram.GetId(), rewardProgram.GetCurrentSubPeriod()))

	err := rewardProgram.ValidateBasic()
	if rewardProgram == nil || err != nil {
		return err
	}

	for _, res := range evaluateRes {
		share, err := k.GetShare(ctx, rewardProgram.GetId(), rewardProgram.GetCurrentSubPeriod(), res.Address.String())
		if err != nil {
			return err
		}

		// Share does not exist so create one
		if share.ValidateBasic() != nil {
			ctx.Logger().Info(fmt.Sprintf("NOTICE: Creating new share structure for rewardProgramId=%d, subPeriod=%d, addr=%s",
				rewardProgram.GetId(), rewardProgram.GetCurrentSubPeriod(), res.Address.String()))

			share = types.NewShare(
				rewardProgram.GetId(),
				rewardProgram.GetCurrentSubPeriod(),
				res.Address.String(),
				false,
				rewardProgram.SubPeriodEndTime.Add(time.Duration(rewardProgram.GetShareExpirationOffset())),
				0,
			)
		}
		share.Amount += res.Shares
		k.SetShare(ctx, &share)
	}

	return nil
}

// Iterates through all the ABCIEvents that match the eventCriteria.
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

func (k Keeper) GetDistributionKeeper() types.DistributionKeeper {
	return nil
}

func (k Keeper) GetStakingKeeper() types.StakingKeeper {
	return k.stakingKeeper
}

func (k *Keeper) SetStakingKeeper(newKeeper types.StakingKeeper) {
	k.stakingKeeper = newKeeper
}

// this method is only for testing
func logEvents(ctx sdk.Context) {
	history := ctx.EventManager().GetABCIEventHistory()
	blockTime := ctx.BlockTime()
	ctx.Logger().Info(fmt.Sprintf("NOTICE: Block time: %v Size of events is %d", blockTime, len(history)))

	// check if sub period has ended
	for _, s := range ctx.EventManager().GetABCIEventHistory() {
		// ctx.Logger().Info(fmt.Sprintf("events type is %s", s.Type))
		ctx.Logger().Info(fmt.Sprintf("------- %s -------\n", s.Type))
		for _, y := range s.Attributes {
			ctx.Logger().Info(fmt.Sprintf("%s: %s\n", y.Key, y.Value))
			// ctx.Logger().Info(fmt.Sprintf("event attribute is %s attribute_key:attribute_value  %s:%s", s.Type, y.Key, y.Value))
			//4:24PM INF events type is coin_spent
			//4:24PM INF event attribute is coin_spent attribute_key:attribute_value  spender:tp1sha7e07l5knw4vdw2vgc3k06gd0fscz9r32yv6
			//4:24PM INF event attribute is coin_spent attribute_key:attribute_value  amount:76200000000000nhash
			//4:24PM INF events type is coin_received
			//4:24PM INF event attribute is coin_received attribute_key:attribute_value  receiver:tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt
			//4:24PM INF event attribute is coin_received attribute_key:attribute_value  amount:76200000000000nhash
			//4:24PM INF events type is transfer
			//4:24PM INF event attribute is transfer attribute_key:attribute_value  recipient:tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt
			//4:24PM INF event attribute is transfer attribute_key:attribute_value  sender:tp1sha7e07l5knw4vdw2vgc3k06gd0fscz9r32yv6
			//4:24PM INF event attribute is transfer attribute_key:attribute_value  amount:76200000000000nhash
		}
		ctx.Logger().Info(fmt.Sprintf("------- %s -------\n\n", s.Type))
	}
}
