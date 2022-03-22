package types

import (
	"errors"
	fmt "fmt"

	"gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"

	epochtypes "github.com/provenance-io/provenance/x/epoch/types"
)

func NewRewardProgram(
	id uint64,
	distributeFromAddress string,
	coin sdk.Coin,
	epoch epochtypes.EpochInfo,
	startEpoch uint64,
	numberEpochs uint64,
	eligibilityCriteria EligibilityCriteria,
) RewardProgram {
	return RewardProgram{
		Id:                    id,
		DistributeFromAddress: distributeFromAddress,
		Coin:                  &coin,
		Epoch:                 &epoch,
		StartEpoch:            startEpoch,
		NumberEpochs:          numberEpochs,
		EligibilityCriteria:   &eligibilityCriteria,
	}
}

func (rp *RewardProgram) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(rp.DistributeFromAddress); err != nil {
		return fmt.Errorf("invalid address for rewards program distribution from address: %w", err)
	}
	if rp.Epoch == nil {
		return errors.New("epoch info cannot be null for rewards program")
	}
	// TODO once validate basic is implemented on epoch
	// if err = rp.Epoch.ValidateBasic(); err != nil {
	// 	return fmt.Errorf("epoch info is not valid: %w", err)
	// }
	if rp.EligibilityCriteria == nil {
		return errors.New("eligibility criteria info cannot be null for rewards program")
	}
	if err := rp.EligibilityCriteria.ValidateBasic(); err != nil {
		return fmt.Errorf("eligibility criteria is not valid: %w", err)
	}

	return nil
}

func (rp *RewardProgram) String() string {
	out, _ := yaml.Marshal(rp)
	return string(out)
}

func NewRewardClaim(address string, sharesPerEpochPerRewardsProgram []*SharesPerEpochPerRewardsProgram) RewardClaim {
	return RewardClaim{
		Address:                 address,
		SharesPerEpochPerReward: sharesPerEpochPerRewardsProgram,
	}
}

func (rc *RewardClaim) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(rc.Address); err != nil {
		return fmt.Errorf("invalid address for reward claim address: %w", err)
	}
	return nil
}

func (rc *RewardClaim) String() string {
	out, _ := yaml.Marshal(rc)
	return string(out)
}

func NewEpochRewardDistribution(epochId string, rewardProgramId uint64, totalRewardsPool *sdk.Coin, total_shares uint64) EpochRewardDistribution {
	return EpochRewardDistribution{
		EpochId:          epochId,
		RewardProgramId:  rewardProgramId,
		TotalRewardsPool: totalRewardsPool,
		TotalShares:      total_shares,
	}
}

func (erd *EpochRewardDistribution) ValidateBasic() error {
	if len(erd.EpochId) < 1 {
		errors.New("epoch reward distribution must have a epoch id")
	}
	if erd.RewardProgramId < 1 {
		errors.New("epoch reward distribution must have a valid reward program id")
	}
	if erd.TotalRewardsPool == nil {
		errors.New("epoch reward distribution must have a reward pool")
	}
	return nil
}

func (erd *EpochRewardDistribution) String() string {
	out, _ := yaml.Marshal(erd)
	return string(out)
}

func NewEligibilityCriteria(name string, action isEligibilityCriteria_Action) EligibilityCriteria {
	return EligibilityCriteria{Name: name, Action: action}
}

func (ec *EligibilityCriteria) ValidateBasic() error {
	if len(ec.Name) < 1 {
		return errors.New("eligibility criteria must have a name")
	}
	if ec.Action == nil {
		return errors.New("eligibility criteria must have an action")
	}
	return nil
}

func (ec *EligibilityCriteria) String() string {
	out, _ := yaml.Marshal(ec)
	return string(out)
}

func NewActionDelegate(minimum int64, maximum int64) ActionDelegate {
	return ActionDelegate{Minimum: minimum, Maximum: maximum}
}

func (ad *ActionDelegate) ValidateBasic() error {
	if ad.Minimum < 0 || ad.Maximum < 0 {
		return fmt.Errorf("rewards action delegate minimum (%d) and maximum (%d) must be greater than 0", ad.Minimum, ad.Maximum)
	}
	return nil
}

func (ad *ActionDelegate) String() string {
	out, _ := yaml.Marshal(ad)
	return string(out)
}

func NewActionTransferDelegations(minimum int64, maximum int64) ActionTransferDelegations {
	return ActionTransferDelegations{Minimum: minimum, Maximum: maximum}
}

func (atd *ActionTransferDelegations) ValidateBasic() error {
	if atd.Minimum < 0 || atd.Maximum < 0 {
		return fmt.Errorf("rewards action delegate minimum (%d) and maximum (%d) must be greater than 0", atd.Minimum, atd.Maximum)
	}
	return nil
}

func (atd *ActionTransferDelegations) String() string {
	out, _ := yaml.Marshal(atd)
	return string(out)
}
