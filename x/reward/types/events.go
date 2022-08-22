package types

const (
	// The type of event generated when a reward program is created
	EventTypeRewardProgramCreated string = "reward_program_created"
	// The type of event generated when a reward program is started
	EventTypeRewardProgramStarted string = "reward_program_started"
	// The type of event generated when a reward program is ended
	EventTypeRewardProgramFinished string = "reward_program_finished"
	// The type of event generated when a reward program is started
	EventTypeRewardProgramExpired string = "reward_program_expired"
	// The type of event generated when a reward program is ended
	EventTypeRewardProgramEnded string = "reward_program_ended"
	// The type of event generated when a address claims rewards
	EventTypeClaimRewards string = "claim_rewards"
	// The type of event generated when a address claims all their rewards
	EventTypeClaimAllRewards string = "claim_all_rewards"

	AttributeKeyRewardProgramID     string = "reward_program_id"
	AttributeKeyRewardProgramIDs    string = "reward_program_ids"
	AttributeKeyRewardsClaimAddress string = "rewards_claim_address"
)
