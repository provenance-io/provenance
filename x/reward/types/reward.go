package types

import (
	"gopkg.in/yaml.v2"
)

func NewRewardProgram() RewardProgram {
	return RewardProgram{}
}

func (rp *RewardProgram) ValidateBasic() error {
	return nil
}

func (rp *RewardProgram) String() string {
	out, _ := yaml.Marshal(rp)
	return string(out)
}

func NewRewardClaim() RewardClaim {
	return RewardClaim{}
}

func (rc *RewardClaim) ValidateBasic() error {
	return nil
}

func (rc *RewardClaim) String() string {
	out, _ := yaml.Marshal(rc)
	return string(out)
}

func NewEpochRewardDistribution() EpochRewardDistribution {
	return EpochRewardDistribution{}
}

func (erd *EpochRewardDistribution) ValidateBasic() error {
	return nil
}

func (erd *EpochRewardDistribution) String() string {
	out, _ := yaml.Marshal(erd)
	return string(out)
}

func NewEligibilityCriteria() EligibilityCriteria {
	return EligibilityCriteria{}
}

func (ec *EligibilityCriteria) ValidateBasic() error {
	return nil
}

func (ec *EligibilityCriteria) String() string {
	out, _ := yaml.Marshal(ec)
	return string(out)
}
