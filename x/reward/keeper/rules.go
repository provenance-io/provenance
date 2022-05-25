package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

// EvaluateRules takes in a Eligibility criteria and measure it against the events in the context
func (k Keeper) DetectQualifyingActions(ctx sdk.Context, program *types.RewardProgram) ([]types.EvaluationResult, error) {
	ctx.Logger().Info(fmt.Sprintf("NOTICE: EvaluateRules for RewardProgram: %d", program.GetId()))
	qualifyingActions := []types.EvaluationResult(nil)

	// Check if any of the transactions are qualifying actions
	for _, supportedAction := range program.GetQualifyingActions() {
		// Get the appropriate RewardAction
		// If it's unsupported we skip it
		action, err := supportedAction.GetRewardAction(ctx)
		if err != nil {
			ctx.Logger().Info(err.Error())
			continue
		}

		// Get all events that the qualyfing action type uses
		eventCriteria := action.GetEventCriteria()
		abciEvents, err := k.GetMatchingEvents(ctx, eventCriteria)
		if err != nil {
			return nil, err
		}

		// Obtain the events that were successfully evaluted
		actions := k.FindQualifyingActions(ctx, program, action, abciEvents)
		qualifyingActions = append(qualifyingActions, actions...)
	}

	return qualifyingActions, nil
}

func (k Keeper) FindQualifyingActions(ctx sdk.Context, program *types.RewardProgram, action types.RewardAction, abciEvents []types.EvaluationResult) []types.EvaluationResult {
	detectedEvents := []types.EvaluationResult(nil)
	if program == nil || action == nil || abciEvents == nil {
		return detectedEvents
	}

	for _, event := range abciEvents {
		// Get the AccountState for the triggering account
		var state types.AccountState
		state, err := k.GetAccountState(ctx, program.GetId(), program.GetCurrentEpoch(), event.Address.String())
		if err != nil {
			continue
		}
		state.ActionCounter++
		k.SetAccountState(ctx, &state)

		if action.Evaluate(ctx, k, state, event) {
			detectedEvents = append(detectedEvents, event)
		}
	}

	return detectedEvents
}

func (k Keeper) RewardShares(ctx sdk.Context, rewardProgram *types.RewardProgram, evaluateRes []types.EvaluationResult) error {
	ctx.Logger().Info(fmt.Sprintf("NOTICE: Recording shares for for rewardProgramId=%d, epochId=%d",
		rewardProgram.GetId(), rewardProgram.GetCurrentEpoch()))

	err := rewardProgram.ValidateBasic()
	if rewardProgram == nil || err != nil {
		return err
	}

	for _, res := range evaluateRes {
		share, err := k.GetShare(ctx, rewardProgram.GetId(), rewardProgram.GetCurrentEpoch(), res.Address.String())
		if err != nil {
			return err
		}

		// Share does not exist so create one
		if share.ValidateBasic() != nil {
			ctx.Logger().Info(fmt.Sprintf("NOTICE: Creating new share structure for rewardProgramId=%d, epochId=%d, addr=%s",
				rewardProgram.GetId(), rewardProgram.GetCurrentEpoch(), res.Address.String()))

			share = types.NewShare(
				rewardProgram.GetId(),
				rewardProgram.GetCurrentEpoch(),
				res.Address.String(),
				false,
				rewardProgram.EpochEndTime.Add(time.Duration(rewardProgram.GetShareExpirationOffset())),
				0,
			)
		}
		share.Amount += res.Shares
		k.SetShare(ctx, &share)
	}

	return nil
}

func (k Keeper) RecordRewardClaims(ctx sdk.Context, epochNumber uint64, program *types.RewardProgram, distribution types.EpochRewardDistribution, evaluateRes []types.EvaluationResult) error {
	// get the address from the eval and check if it has delegation
	// it's an array so should be deterministic
	for _, res := range evaluateRes {
		ctx.Logger().Info(fmt.Sprintf("NOTICE: RecordRewardClaims: %v %v %v %v", epochNumber, program, distribution, res))
		// add a share to the final total
		// we know the rewards it so update the epoch reward
		// distribution.TotalShares = distribution.TotalShares + res.shares
		// add it to the claims
		claim, err := k.GetRewardClaim(ctx, res.Address.String())
		if err != nil {
			return err
		}

		if claim.Address != "" {
			found := false
			var mutatedSharesPerEpochRewards []types.SharesPerEpochPerRewardsProgram
			// set a new claim or add to a claim
			// TODO: Need to do EC checking
			for _, rewardClaimForAddress := range claim.SharesPerEpochPerReward {
				if rewardClaimForAddress.RewardProgramId == program.Id {
					rewardClaimForAddress.EphemeralActionCount += res.Shares
					rewardClaimForAddress.TotalShares += res.Shares
					rewardClaimForAddress.LatestRecordedEpoch = epochNumber
					mutatedSharesPerEpochRewards = append(mutatedSharesPerEpochRewards, rewardClaimForAddress)
					found = true
					// we know the rewards it so update the epoch reward
					distribution.TotalShares += res.Shares
				} else {
					mutatedSharesPerEpochRewards = append(mutatedSharesPerEpochRewards, rewardClaimForAddress)
				}
			}
			if found {
				claim.SharesPerEpochPerReward = mutatedSharesPerEpochRewards
			} else {
				mutatedSharesPerEpochRewards = append(mutatedSharesPerEpochRewards, types.SharesPerEpochPerRewardsProgram{
					RewardProgramId:      program.Id,
					TotalShares:          res.Shares,
					EphemeralActionCount: res.Shares,
					LatestRecordedEpoch:  epochNumber,
					Claimed:              false,
					Expired:              false,
					TotalRewardClaimed:   sdk.Coin{},
				})
				// we know the rewards it so update the epoch reward
				distribution.TotalShares += res.Shares
			}
			claim.SharesPerEpochPerReward = mutatedSharesPerEpochRewards
			k.SetRewardClaim(ctx, claim)
		} else {
			// set a brand new claim
			var mutatedSharesPerEpochRewards []types.SharesPerEpochPerRewardsProgram
			k.SetRewardClaim(ctx, types.RewardClaim{
				Address: res.Address.String(),
				SharesPerEpochPerReward: append(mutatedSharesPerEpochRewards, types.SharesPerEpochPerRewardsProgram{
					RewardProgramId:      program.Id,
					TotalShares:          res.Shares,
					EphemeralActionCount: res.Shares,
					LatestRecordedEpoch:  epochNumber,
					Claimed:              false,
					Expired:              false,
					TotalRewardClaimed:   sdk.Coin{},
				}),
				Expired: false,
			})
			// we know the rewards it so update the epoch reward
			distribution.TotalShares += res.Shares
		}
	}
	// set total rewards
	k.SetEpochRewardDistribution(ctx, distribution)
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

// TODO This is currently written for only the delegate logic. We will want to refactor to accommodate other events
func (k Keeper) GetMatchingEvents(ctx sdk.Context, eventCriteria *types.EventCriteria) ([]types.EvaluationResult, error) {
	result := ([]types.EvaluationResult)(nil)

	err := k.IterateABCIEvents(ctx, eventCriteria, func(eventType string, attributes *map[string][]byte) error {
		// really not possible to get an error but could happen i guess

		// This logic is specific to one type of event
		if eventType == "delegate" {
			address := (*attributes)["validator"]
			validator, err := sdk.ValAddressFromBech32(string(address))
			if err != nil {
				return err
			}
			result = append(result, types.EvaluationResult{
				EventTypeToSearch: eventType,
				AttributeKey:      "address",
				Shares:            1,
				Validator:         validator,
			})
		} else if eventType == "create_validator" {
			address := (*attributes)["validator"]
			validator, err := sdk.ValAddressFromBech32(string(address))
			if err != nil {
				return err
			}
			result = append(result, types.EvaluationResult{
				EventTypeToSearch: eventType,
				AttributeKey:      "address",
				Shares:            1,
				Validator:         validator,
			})
		} else if eventType == "message" {
			// Update the last result to have the delegator's address
			address := (*attributes)["sender"]
			address, err := sdk.AccAddressFromBech32(string(address))
			if err != nil {
				return err
			}

			result[len(result)-1].Address = address
			result[len(result)-1].Delegator = address
		}

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
