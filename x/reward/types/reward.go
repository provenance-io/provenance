package types

import (
	"errors"
	fmt "fmt"

	// "github.com/gogo/protobuf/proto"
	"gopkg.in/yaml.v2"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/gogo/protobuf/proto"
)

var (
	_ RewardAction = &ActionDelegate{}
	_ RewardAction = &ActionTransferDelegations{}
)

const (
	ActionTypeDelegate            = "ActionDelegate"
	ActionTypeTransferDelegations = "ActionTransferDelegations"
)

type (
	// RewardAction defines the interface that actions need to implement
	RewardAction interface {
		proto.Message

		ActionType() string
		IsEligible() error
	}
)

func NewRewardProgram(
	id uint64,
	distributeFromAddress string,
	coin sdk.Coin,
	epochId string,
	startEpoch uint64,
	numberEpochs uint64,
	eligibilityCriteria EligibilityCriteria,
) RewardProgram {
	return RewardProgram{
		Id:                    id,
		DistributeFromAddress: distributeFromAddress,
		Coin:                  coin,
		EpochId:               epochId,
		StartEpoch:            startEpoch,
		NumberEpochs:          numberEpochs,
		EligibilityCriteria:   &eligibilityCriteria,
	}
}

func (rp *RewardProgram) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(rp.DistributeFromAddress); err != nil {
		return fmt.Errorf("invalid address for rewards program distribution from address: %w", err)
	}
	if len(rp.EpochId) == 0 {
		return errors.New("epoch id cannot be empty")
	}
	if rp.EligibilityCriteria == nil {
		return errors.New("eligibility criteria info cannot be null for rewards program")
	}
	if err := rp.EligibilityCriteria.ValidateBasic(); err != nil {
		return fmt.Errorf("eligibility criteria is not valid: %w", err)
	}
	if !rp.Coin.IsPositive() {
		return fmt.Errorf("reward program requires coins: %v", rp.Coin)
	}

	return nil
}

func (rp *RewardProgram) String() string {
	out, _ := yaml.Marshal(rp)
	return string(out)
}

func NewRewardClaim(address string, sharesPerEpochPerRewardsProgram []SharesPerEpochPerRewardsProgram) RewardClaim {
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

func NewEpochRewardDistribution(epochId string, rewardProgramId uint64, totalRewardsPool sdk.Coin, totalShares int64) EpochRewardDistribution {
	return EpochRewardDistribution{
		EpochId:          epochId,
		RewardProgramId:  rewardProgramId,
		TotalRewardsPool: totalRewardsPool,
		TotalShares:      totalShares,
	}
}

func (erd *EpochRewardDistribution) ValidateBasic() error {
	if len(erd.EpochId) < 1 {
		return errors.New("epoch reward distribution must have a epoch id")
	}
	if erd.RewardProgramId < 1 {
		return errors.New("epoch reward distribution must have a valid reward program id")
	}
	if !erd.TotalRewardsPool.IsPositive() {
		return errors.New("epoch reward distribution must have a reward pool")
	}
	return nil
}

func (erd *EpochRewardDistribution) String() string {
	out, _ := yaml.Marshal(erd)
	return string(out)
}

func NewEligibilityCriteria(name string, action RewardAction) EligibilityCriteria {
	ec := EligibilityCriteria{Name: name}
	err := ec.SetAction(action)
	if err != nil {
		panic(err)
	}
	return ec
}

func (ec *EligibilityCriteria) SetAction(rewardAction RewardAction) error {
	if rewardAction == nil {
		ec.Action = nil
		return nil
	}
	any, err := codectypes.NewAnyWithValue(rewardAction)
	if err == nil {
		ec.Action = any
	}
	return err
}

func (ec *EligibilityCriteria) GetAction() RewardAction {
	if ec.Action == nil {
		return nil
	}
	content, ok := ec.Action.GetCachedValue().(RewardAction)
	if !ok {
		return nil
	}
	return content
}

func (ec *EligibilityCriteria) ValidateBasic() error {
	if len(ec.Name) < 1 {
		return errors.New("eligibility criteria must have a name")
	}
	if ec.Action == nil {
		return errors.New("eligibility criteria must have an action")
	}
	//_, ok := ec.Action.GetCachedValue().(RewardAction)
	//if !ok {
	//	return errors.New("eligibility criteria action must implement RewardAction interface")
	//}
	return nil
}

func (ec *EligibilityCriteria) String() string {
	out, _ := yaml.Marshal(ec)
	return string(out)
}

func NewActionDelegate(ephemeral_count int64, maximum int64) ActionDelegate {
	return ActionDelegate{EphemeralCount: ephemeral_count, Maximum: maximum}
}

func (ad *ActionDelegate) ValidateBasic() error {
	if ad.EphemeralCount < 0 || ad.Maximum < 0 {
		return fmt.Errorf("rewards action delegate ephemeral count (%d) and maximum (%d) must be greater than 0", ad.EphemeralCount, ad.Maximum)
	}
	return nil
}

func (ad *ActionDelegate) ActionType() string {
	return ActionTypeDelegate
}

func (ad *ActionDelegate) IsEligible() error {
	// TODO execute all the rules for action?
	return nil
}

func (ad *ActionDelegate) String() string {
	out, _ := yaml.Marshal(ad)
	return string(out)
}

func NewActionTransferDelegations(ephemeral_count int64, maximum int64) ActionTransferDelegations {
	return ActionTransferDelegations{EphemeralCount: ephemeral_count, Maximum: maximum}
}

func (atd *ActionTransferDelegations) ValidateBasic() error {
	if atd.EphemeralCount < 0 || atd.Maximum < 0 {
		return fmt.Errorf("rewards action delegate ephemeral_count (%d) and maximum (%d) must be greater than 0", atd.EphemeralCount, atd.Maximum)
	}
	return nil
}

func (atd *ActionTransferDelegations) String() string {
	out, _ := yaml.Marshal(atd)
	return string(out)
}

func (atd *ActionTransferDelegations) ActionType() string {
	return ActionTypeDelegate
}

func (atd *ActionTransferDelegations) IsEligible() error {
	// TODO execute all the rules for action?
	return nil
}

func NewSharesPerEpochPerRewardsProgram(
	rewardProgramId uint64,
	totalShares int64,
	epochId string,
	latestRecordedEpoch uint64,
	claimed bool,
	expirationHeight uint64,
	expired bool,
	totalRewardsClaimed sdk.Coin,
) SharesPerEpochPerRewardsProgram {
	return SharesPerEpochPerRewardsProgram{
		RewardProgramId:     rewardProgramId,
		TotalShares:         totalShares,
		LatestRecordedEpoch: latestRecordedEpoch,
		Claimed:             claimed,
		Expired:             expired,
		TotalRewardClaimed:  totalRewardsClaimed,
	}
}

func (apeprp *SharesPerEpochPerRewardsProgram) ValidateBasice() error {
	if apeprp.RewardProgramId < 1 {
		return errors.New("shares per epoch must have a valid reward program id")
	}
	if apeprp.LatestRecordedEpoch < 1 {
		return errors.New("latest recorded epoch cannot be less than 1")
	}
	// TODO need more?
	return nil
}

func (apeprp *SharesPerEpochPerRewardsProgram) String() string {
	out, _ := yaml.Marshal(apeprp)
	return string(out)
}
