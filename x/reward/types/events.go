package types

const (
	// The type of event generated when account attributes are added.
	EventTypeRewardProgramCreated string = "reward_program_created"
)

func NewEventCreateRewardProgram() EventCreateRewardProgram {
	return EventCreateRewardProgram{}
}
