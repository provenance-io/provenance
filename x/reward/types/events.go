package types

const (
	// The type of event generated when a reward program is created
	EventTypeRewardProgramCreated string = "reward_program_created"
	// The type of event generated when a address claims rewards
	EventTypeClaimRewards string = "claim_rewards"

	AttributeKeyRewardProgramID     string = "reward_program_id"
	AttributeKeyRewardsClaimAddress string = "rewards_claim_address"
)

func NewEventCreateRewardProgram() EventCreateRewardProgram {
	return EventCreateRewardProgram{}
}
