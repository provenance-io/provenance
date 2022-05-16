package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

type EventCriteria struct {
	eventType      string
	attribute      string
	attributeValue string
}

type EvaluationResult struct {
	eventTypeToSearch string
	attributeKey      string
	shares            int64
	address           sdk.AccAddress // shares to attribute to this address
}

// EvaluateRules takes in a Eligibility criteria and measure it against the events in the context
func (k Keeper) EvaluateRules(ctx sdk.Context, program *types.RewardProgram) error {
	ctx.Logger().Info(fmt.Sprintf("NOTICE: EvaluateRules for RewardProgram: %d", program.GetId()))

	// Check if any of the transactions match the qualifying actions
	for _, qualifyingAction := range program.GetQualifyingActions() {
		var evaluateRes []EvaluationResult
		var err error

		switch actionType := qualifyingAction.GetType().(type) {
		case *types.QualifyingAction_Delegate:
			ctx.Logger().Info(fmt.Sprintf("NOTICE: The Action type is %s", actionType))
			// check the event history
			// for transfers event and make sure there is a sender

			// We probably want to check the criteria on the delegation within here
			evaluateRes, err = k.EvaluateDelegation(ctx, program)
			if err != nil {
				return err
			}
		case *types.QualifyingAction_TransferDelegations:
			ctx.Logger().Info(fmt.Sprintf("NOTICE: The Action type is %s", actionType))
			// check the event history
			// for transfers event and make sure there is a sender
			evaluateRes, err = k.EvaluateTransferAndCheckDelegation(ctx, program)
			if err != nil {
				return err
			}
		default:
			// Skip any unsupported actions
			ctx.Logger().Error(fmt.Sprintf("The Action type %s, cannot be evaluated", actionType))
			continue
		}

		// Record any results
		err = k.RecordShares(ctx, program, evaluateRes)
		if err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) RecordShares(ctx sdk.Context, rewardProgram *types.RewardProgram, evaluateRes []EvaluationResult) error {
	ctx.Logger().Info(fmt.Sprintf("NOTICE: Recording shares for for rewardProgramId=%d, epochId=%d",
		rewardProgram.GetId(), rewardProgram.GetCurrentEpoch()))

	for _, res := range evaluateRes {
		share, err := k.GetShare(ctx, rewardProgram.GetId(), rewardProgram.GetCurrentEpoch(), res.address.String())
		if err != nil {
			return err
		}

		// Share does not exist so create one
		if share.ValidateBasic() != nil {
			ctx.Logger().Info(fmt.Sprintf("NOTICE: Creating new share structure for rewardProgramId=%d, epochId=%d, addr=%s",
				rewardProgram.GetId(), rewardProgram.GetCurrentEpoch(), res.address.String()))

			share = types.NewShare(
				rewardProgram.GetId(),
				rewardProgram.GetCurrentEpoch(),
				string(res.address),
				false,
				rewardProgram.EpochEndTime.Add(time.Duration(rewardProgram.GetShareExpirationOffset())),
				0,
			)
		}
		share.Amount += res.shares
		k.SetShare(ctx, &share)
	}

	return nil
}

func (k Keeper) RecordRewardClaims(ctx sdk.Context, epochNumber uint64, program *types.RewardProgram, distribution types.EpochRewardDistribution, evaluateRes []EvaluationResult) error {
	// get the address from the eval and check if it has delegation
	// it's an array so should be deterministic
	for _, res := range evaluateRes {
		ctx.Logger().Info(fmt.Sprintf("NOTICE: RecordRewardClaims: %v %v %v %v", epochNumber, program, distribution, res))
		// add a share to the final total
		// we know the rewards it so update the epoch reward
		//distribution.TotalShares = distribution.TotalShares + res.shares
		// add it to the claims
		claim, err := k.GetRewardClaim(ctx, res.address.String())
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
					rewardClaimForAddress.EphemeralActionCount = rewardClaimForAddress.EphemeralActionCount + res.shares
					rewardClaimForAddress.TotalShares = rewardClaimForAddress.TotalShares + res.shares
					rewardClaimForAddress.LatestRecordedEpoch = epochNumber
					mutatedSharesPerEpochRewards = append(mutatedSharesPerEpochRewards, rewardClaimForAddress)
					found = true
					// we know the rewards it so update the epoch reward
					distribution.TotalShares = distribution.TotalShares + res.shares
				} else {
					mutatedSharesPerEpochRewards = append(mutatedSharesPerEpochRewards, rewardClaimForAddress)
				}
			}
			if found {
				claim.SharesPerEpochPerReward = mutatedSharesPerEpochRewards
			} else {
				mutatedSharesPerEpochRewards = append(mutatedSharesPerEpochRewards, types.SharesPerEpochPerRewardsProgram{
					RewardProgramId:      program.Id,
					TotalShares:          res.shares,
					EphemeralActionCount: res.shares,
					LatestRecordedEpoch:  epochNumber,
					Claimed:              false,
					Expired:              false,
					TotalRewardClaimed:   sdk.Coin{},
				})
				// we know the rewards it so update the epoch reward
				distribution.TotalShares = distribution.TotalShares + res.shares
			}
			claim.SharesPerEpochPerReward = mutatedSharesPerEpochRewards
			k.SetRewardClaim(ctx, claim)
		} else {
			//set a brand new claim
			var mutatedSharesPerEpochRewards []types.SharesPerEpochPerRewardsProgram
			k.SetRewardClaim(ctx, types.RewardClaim{
				Address: res.address.String(),
				SharesPerEpochPerReward: append(mutatedSharesPerEpochRewards, types.SharesPerEpochPerRewardsProgram{
					RewardProgramId:      program.Id,
					TotalShares:          res.shares,
					EphemeralActionCount: res.shares,
					LatestRecordedEpoch:  epochNumber,
					Claimed:              false,
					Expired:              false,
					TotalRewardClaimed:   sdk.Coin{},
				}),
				Expired: false,
			})
			// we know the rewards it so update the epoch reward
			distribution.TotalShares = distribution.TotalShares + res.shares

		}
	}
	//set total rewards
	k.SetEpochRewardDistribution(ctx, distribution)
	return nil
}

func (k Keeper) EvaluateTransferAndCheckDelegation(ctx sdk.Context, rewardProgram *types.RewardProgram) ([]EvaluationResult, error) {
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

func (k Keeper) IterateABCIEvents(ctx sdk.Context, eventCriteria *EventCriteria, action func(*abci.Event, *abci.EventAttribute) error) error {
	for _, event := range ctx.EventManager().GetABCIEventHistory() {
		ctx.Logger().Info(fmt.Sprintf("events type is %s", event.Type))

		// Event types must match or wildcard/empty must be present
		if event.Type != eventCriteria.eventType && eventCriteria.eventType != "" {
			continue
		}

		for _, attribute := range event.Attributes {
			ctx.Logger().Info(fmt.Sprintf("event attribute is %s attribute_key:attribute_value  %s:%s", event.Type, attribute.Key, attribute.Value))

			// Attribute names must match or wildcard/empty must be present
			if eventCriteria.attribute != string(attribute.Key) && eventCriteria.attribute != "" {
				continue
			}

			// Attribute values must match or wildcard/empty must be present
			if eventCriteria.attributeValue != string(attribute.Value) && eventCriteria.attributeValue != "" {
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

func (k Keeper) EvaluateDelegation(ctx sdk.Context, rewardProgram *types.RewardProgram) ([]EvaluationResult, error) {
	events, err := k.GetDelegationEvents(ctx)
	if err != nil {
		return nil, err
	}

	results := ([]EvaluationResult)(nil)
	for _, res := range events {
		state, err := k.GetAccountState(ctx, rewardProgram.GetId(), rewardProgram.GetCurrentEpoch(), string(res.address))
		if err != nil {
			continue
		}

		state.ActionCounter += 1
		k.SetAccountState(ctx, &state)

		// We append to the new list only if
		// If we match the condition then append to the results

		results = append(results, res)
	}

	return results, err
}

// EvaluateDelegateEvents
// unfortunately delegate events appear like this
//10:38PM INF events type is message
//10:38PM INF event attribute is message attribute_key:attribute_value  module:staking
//10:38PM INF event attribute is message attribute_key:attribute_value  sender:tp1hsrqwfypd3w3py2kw7fajw4fzfqwv5qdel6vf6
func (k Keeper) GetDelegationEvents(ctx sdk.Context) ([]EvaluationResult, error) {
	result := ([]EvaluationResult)(nil)
	eventCriteria := EventCriteria{
		eventType:      "message",
		attribute:      "staking",
		attributeValue: "sender",
	}

	err := k.IterateABCIEvents(ctx, &eventCriteria, func(event *abci.Event, attribute *abci.EventAttribute) error {
		// really not possible to get an error but could happen i guess
		address, err := searchValue(event.Attributes, string(attribute.Key))

		//TODO check this address has a delegation
		if err != nil {
			return err
		}
		result = append(result, EvaluationResult{
			eventTypeToSearch: event.Type,
			attributeKey:      string(attribute.Key),
			shares:            1,
			address:           address,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (k Keeper) GetTransferEvents(ctx sdk.Context) ([]EvaluationResult, error) {
	result := ([]EvaluationResult)(nil)
	eventCriteria := EventCriteria{
		eventType: "transfer",
		attribute: "sender",
	}

	err := k.IterateABCIEvents(ctx, &eventCriteria, func(event *abci.Event, attribute *abci.EventAttribute) error {
		// really not possible to get an error but could happen i guess
		address, err := sdk.AccAddressFromBech32(string(attribute.Value))

		//TODO check this address has a delegation
		if err != nil {
			return err
		}

		result = append(result, EvaluationResult{
			eventTypeToSearch: event.Type,
			attributeKey:      string(attribute.Key),
			shares:            1,
			address:           address,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func searchValue(attributes []abci.EventAttribute, attributeKey string) (sdk.AccAddress, error) {
	for _, y := range attributes {
		if attributeKey == string(y.Key) {
			// really not possible to get an error but could happen i guess
			address, err := sdk.AccAddressFromBech32(string(y.Value))
			//TODO check this address has a delegation
			if err != nil {
				return nil, err
			}
			return address, err
		}
	}
	return nil, nil
}
