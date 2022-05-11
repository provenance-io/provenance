package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/provenance-io/provenance/x/reward/types"
)

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
			evaluateRes, err = k.EvaluateDelegation(ctx, &program.EligibilityCriteria)
			if err != nil {
				return err
			}
		case *types.QualifyingAction_TransferDelegations:
			ctx.Logger().Info(fmt.Sprintf("NOTICE: The Action type is %s", actionType))
			// check the event history
			// for transfers event and make sure there is a sender
			evaluateRes, err = k.EvaluateTransferAndCheckDelegation(ctx, &program.EligibilityCriteria)
			if err != nil {
				return err
			}
		default:
			// Skip any unsupported actions
			ctx.Logger().Error(fmt.Sprintf("The Action type %s, cannot be evaluated", actionType))
			continue
		}

		// Record any results
		err = k.GrantRewardProgramShares(ctx, program, evaluateRes)
		if err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) GrantRewardProgramShares(ctx sdk.Context, program *types.RewardProgram, evaluateRes []EvaluationResult) error {
	for _, res := range evaluateRes {
		ctx.Logger().Info(fmt.Sprintf("NOTICE: RecordRewardClaims: %v %v", program, res))
		program.Shares = append(program.Shares, types.Share{
			Address: "",
			Claimed: false,
			// TODO This needs to be set to the correct expiration time. Either universally or needs to be included in EvaluationResult
			ExpireTime: timestamppb.Now().AsTime(),
			Amount:     res.shares,
		})
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

func (k Keeper) EvaluateTransferAndCheckDelegation(ctx sdk.Context, criteria *types.EligibilityCriteria) ([]EvaluationResult, error) {
	result := ([]EvaluationResult)(nil)
	/*evaluateRes, err := k.EvaluateSearchEvents(ctx, "transfer", "sender")
	if err != nil {
		return nil, err
	}
	for _, s := range evaluateRes {
		if len(k.CheckActiveDelegations(ctx, s.address)) > 0 {
			result = append(result, s)
		}
	}*/
	return result, nil
}

func (k Keeper) EvaluateDelegation(ctx sdk.Context, criteria *types.EligibilityCriteria) ([]EvaluationResult, error) {
	evaluateRes, err := k.EvaluateDelegateEvents(ctx, "message", "staking", "sender")
	return evaluateRes, err
}

func (k Keeper) EvaluateSearchEvents(ctx sdk.Context, eventTypeToSearch string, attributeKey string) ([]EvaluationResult, error) {
	result := ([]EvaluationResult)(nil)
	for _, s := range ctx.EventManager().GetABCIEventHistory() {
		ctx.Logger().Info(fmt.Sprintf("events type is %s", s.Type))
		if s.Type == eventTypeToSearch {
			// now look for the attribute
			for _, y := range s.Attributes {
				ctx.Logger().Info(fmt.Sprintf("event attribute is %s attribute_key:attribute_value  %s:%s", s.Type, y.Key, y.Value))
				//4:24PM INF events type is transfer
				//4:24PM INF event attribute is transfer attribute_key:attribute_value  recipient:tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt
				//4:24PM INF event attribute is transfer attribute_key:attribute_value  sender:tp1sha7e07l5knw4vdw2vgc3k06gd0fscz9r32yv6
				//4:24PM INF event attribute is transfer attribute_key:attribute_value  amount:76200000000000nhash
				if attributeKey == string(y.Key) {

					// really not possible to get an error but could happen i guess
					address, err := sdk.AccAddressFromBech32(string(y.Value))

					//TODO check this address has a delegation
					if err != nil {
						return nil, err
					}
					result = append(result, EvaluationResult{
						eventTypeToSearch: eventTypeToSearch,
						attributeKey:      string(y.Key),
						shares:            1,
						address:           address,
					})
				}
			}
		}
	}

	return result, nil

}

// EvaluateDelegateEvents
// unfortunately delegate events appear like this
//10:38PM INF events type is message
//10:38PM INF event attribute is message attribute_key:attribute_value  module:staking
//10:38PM INF event attribute is message attribute_key:attribute_value  sender:tp1hsrqwfypd3w3py2kw7fajw4fzfqwv5qdel6vf6
func (k Keeper) EvaluateDelegateEvents(ctx sdk.Context, eventTypeToSearch string, attributeValue string, attributeKey string) ([]EvaluationResult, error) {
	result := ([]EvaluationResult)(nil)
	for _, s := range ctx.EventManager().GetABCIEventHistory() {
		ctx.Logger().Info(fmt.Sprintf("events type is %s", s.Type))
		if s.Type == eventTypeToSearch {
			// now look for the attribute
			for _, y := range s.Attributes {
				ctx.Logger().Info(fmt.Sprintf("event attribute is %s attribute_key:attribute_value  %s:%s", s.Type, y.Key, y.Value))

				if attributeValue == string(y.Value) {

					// really not possible to get an error but could happen i guess
					address, err := searchValue(s.Attributes, attributeKey)

					//TODO check this address has a delegation
					if err != nil {
						return nil, err
					}
					result = append(result, EvaluationResult{
						eventTypeToSearch: eventTypeToSearch,
						attributeKey:      string(y.Key),
						shares:            1,
						address:           address,
					})
				}
			}
		}
	}

	return result, nil

}

func (k Keeper) GetAllActiveRewardPrograms(ctx sdk.Context) ([]types.RewardProgram, error) {
	var rewardPrograms []types.RewardProgram
	// get all the rewards programs
	err := k.IterateRewardPrograms(ctx, func(rewardProgram types.RewardProgram) (stop bool) {
		if !rewardProgram.Finished && rewardProgram.Started {
			rewardPrograms = append(rewardPrograms, rewardProgram)
		}
		return false
	})
	return rewardPrograms, err
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
