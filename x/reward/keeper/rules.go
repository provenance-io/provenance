package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

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
			ctx.Logger().Error(err.Error())
			continue
		}

		// Get all events that the qualyfing action type uses
		eventCriteria := action.GetEventCriteria()
		abciEvents, err := k.GetMatchingEvents(ctx, &eventCriteria)
		if err != nil {
			return nil, err
		}

		// Obtain the events that were successfully evaluted
		detectedActions := k.FindQualifyingActions(ctx, program, action, abciEvents)
		qualifyingActions = append(qualifyingActions, detectedActions...)
	}

	return qualifyingActions, nil
}

func (k Keeper) FindQualifyingActions(ctx sdk.Context, program *types.RewardProgram, action types.RewardAction, abciEvents []types.EvaluationResult) []types.EvaluationResult {
	detectedEvents := []types.EvaluationResult(nil)

	for _, event := range abciEvents {
		// Get the AccountState for the triggering account
		var state types.AccountState
		state, err := k.GetAccountState(ctx, program.GetId(), program.GetCurrentEpoch(), string(event.Address))
		if err != nil {
			continue
		}
		state.ActionCounter++
		k.SetAccountState(ctx, &state)

		// We want to evaluate it
		// If it passes then it should be added to the list
		if action.Evaluate(ctx, state) {
			detectedEvents = append(detectedEvents, event)
		}
	}

	return detectedEvents
}

func (k Keeper) RewardShares(ctx sdk.Context, rewardProgram *types.RewardProgram, evaluateRes []types.EvaluationResult) error {
	ctx.Logger().Info(fmt.Sprintf("NOTICE: Recording shares for for rewardProgramId=%d, epochId=%d",
		rewardProgram.GetId(), rewardProgram.GetCurrentEpoch()))

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
				string(res.Address),
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

func (k Keeper) EvaluateTransferAndCheckDelegation(ctx sdk.Context, rewardProgram *types.RewardProgram) ([]types.EvaluationResult, error) {
	result, err := k.GetTransferEvents(ctx)
	if err != nil {
		return nil, err
	}

	// TODO Apply actions to each account based on events
	// TODO Filter out results that don't meet criteria.

	/*for _, s := range result {
		if len(k.CheckActiveDelegations(ctx, s.address)) > 0 {
			result = append(result, s)
		}
	}*/

	return result, nil
}

func (k Keeper) IterateABCIEvents(ctx sdk.Context, eventCriteria *types.EventCriteria, action func(*abci.Event, *abci.EventAttribute) error) error {
	for _, event := range ctx.EventManager().GetABCIEventHistory() {
		event := event
		ctx.Logger().Info(fmt.Sprintf("events type is %s", event.Type))

		// Event types must match or wildcard/empty must be present
		if event.Type != eventCriteria.EventType && eventCriteria.EventType != "" {
			continue
		}

		for _, attribute := range event.Attributes {
			attribute := attribute
			ctx.Logger().Info(fmt.Sprintf("event attribute is %s attribute_key:attribute_value  %s:%s", event.Type, attribute.Key, attribute.Value))

			// Attribute names must match or wildcard/empty must be present
			if eventCriteria.Attribute != string(attribute.Key) && eventCriteria.Attribute != "" {
				continue
			}

			// Attribute values must match or wildcard/empty must be present
			if eventCriteria.AttributeValue != string(attribute.Value) && eventCriteria.AttributeValue != "" {
				continue
			}

			err := action(&event, &attribute)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// EvaluateDelegateEvents
// unfortunately delegate events appear like this
//10:38PM INF events type is message
//10:38PM INF event attribute is message attribute_key:attribute_value  module:staking
//10:38PM INF event attribute is message attribute_key:attribute_value  sender:tp1hsrqwfypd3w3py2kw7fajw4fzfqwv5qdel6vf6
func (k Keeper) GetDelegationEvents(ctx sdk.Context) ([]types.EvaluationResult, error) {
	result := ([]types.EvaluationResult)(nil)
	eventCriteria := types.EventCriteria{
		EventType:      "message",
		Attribute:      "staking",
		AttributeValue: "sender",
	}

	err := k.IterateABCIEvents(ctx, &eventCriteria, func(event *abci.Event, attribute *abci.EventAttribute) error {
		// really not possible to get an error but could happen i guess
		address, err := sdk.AccAddressFromBech32(string(attribute.Value))

		// TODO check this address has a delegation
		if err != nil {
			return err
		}
		result = append(result, types.EvaluationResult{
			EventTypeToSearch: event.Type,
			AttributeKey:      string(attribute.Key),
			Shares:            1,
			Address:           address,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (k Keeper) GetTransferEvents(ctx sdk.Context) ([]types.EvaluationResult, error) {
	result := ([]types.EvaluationResult)(nil)
	eventCriteria := types.EventCriteria{
		EventType: "transfer",
		Attribute: "sender",
	}

	err := k.IterateABCIEvents(ctx, &eventCriteria, func(event *abci.Event, attribute *abci.EventAttribute) error {
		// really not possible to get an error but could happen i guess
		address, err := sdk.AccAddressFromBech32(string(attribute.Value))

		// TODO check this address has a delegation
		if err != nil {
			return err
		}

		result = append(result, types.EvaluationResult{
			EventTypeToSearch: event.Type,
			AttributeKey:      string(attribute.Key),
			Shares:            1,
			Address:           address,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (k Keeper) GetMatchingEvents(ctx sdk.Context, eventCriteria *types.EventCriteria) ([]types.EvaluationResult, error) {
	result := ([]types.EvaluationResult)(nil)

	err := k.IterateABCIEvents(ctx, eventCriteria, func(event *abci.Event, attribute *abci.EventAttribute) error {
		// really not possible to get an error but could happen i guess
		address, err := sdk.AccAddressFromBech32(string(attribute.Value))

		// TODO check this address has a delegation
		if err != nil {
			return err
		}

		result = append(result, types.EvaluationResult{
			EventTypeToSearch: event.Type,
			AttributeKey:      string(attribute.Key),
			Shares:            1,
			Address:           address,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

/*func searchValue(attributes []abci.EventAttribute, attributeKey string) (sdk.AccAddress, error) {
	for _, y := range attributes {
		if attributeKey == string(y.Key) {
			// really not possible to get an error but could happen i guess
			address, err := sdk.AccAddressFromBech32(string(y.Value))
			// TODO check this address has a delegation
			if err != nil {
				return nil, err
			}
			return address, err
		}
	}
	return nil, nil
}*/
