package types

import (
	"errors"
	fmt "fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gopkg.in/yaml.v2"
)

const (
	TypeMsgCreateRewardProgramRequest = "create_reward_program"
	TypeMsgClaimRewardRequest         = "claim_reward"
)

// Compile time interface checks.
var _ sdk.Msg = &MsgCreateRewardProgramRequest{}

// NewMsgCreateRewardProgramRequest creates a new create reward program request
func NewMsgCreateRewardProgramRequest(
	title string,
	description string,
	distributeFromAddress string,
	totalRewardPool sdk.Coin,
	maxRewardPerClaimAddress sdk.Coin,
	programStartTime time.Time,
	rewardPeriodDays uint64,
	claimPeriodDays uint64,
	expireDays uint64,
	qualifyingAction []QualifyingAction,
) *MsgCreateRewardProgramRequest {
	return &MsgCreateRewardProgramRequest{
		Title:                    title,
		Description:              description,
		DistributeFromAddress:    distributeFromAddress,
		TotalRewardPool:          totalRewardPool,
		MaxRewardPerClaimAddress: maxRewardPerClaimAddress,
		ProgramStartTime:         programStartTime,
		RewardPeriodDays:         rewardPeriodDays,
		ClaimPeriodDays:          claimPeriodDays,
		ExpireDays:               expireDays,
		QualifyingActions:        qualifyingAction,
	}
}

// Route implements Msg
func (msg MsgCreateRewardProgramRequest) Route() string { return ModuleName }

// Type implements Msg
func (msg MsgCreateRewardProgramRequest) Type() string { return TypeMsgCreateRewardProgramRequest }

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgCreateRewardProgramRequest) ValidateBasic() error {
	title := msg.GetTitle()
	if len(strings.TrimSpace(title)) == 0 {
		return errors.New("reward program title cannot be blank")
	}
	if len(title) > MaxTitleLength {
		return fmt.Errorf("reward program title is longer than max length of %d", MaxTitleLength)
	}
	description := msg.GetDescription()
	if len(description) == 0 {
		return errors.New("reward program description cannot be blank")
	}
	if len(description) > MaxDescriptionLength {
		return fmt.Errorf("reward program description is longer than max length of %d", MaxDescriptionLength)
	}
	if _, err := sdk.AccAddressFromBech32(msg.DistributeFromAddress); err != nil {
		return fmt.Errorf("invalid address for rewards program distribution from address: %w", err)
	}
	startTime := msg.ProgramStartTime.UTC()
	if startTime.Hour() != 0 || startTime.Minute() != 0 || startTime.Second() != 0 || startTime.Nanosecond() != 0 {
		return fmt.Errorf("invalid time must be of date only: %v", startTime)
	}
	if !msg.TotalRewardPool.IsPositive() {
		return fmt.Errorf("reward program requires total reward pool to be positive: %v", msg.TotalRewardPool)
	}
	if !msg.MaxRewardPerClaimAddress.IsPositive() {
		return fmt.Errorf("reward program requires positive max reward by address: %v", msg.MaxRewardPerClaimAddress)
	}
	if msg.TotalRewardPool.Denom != msg.MaxRewardPerClaimAddress.Denom {
		return fmt.Errorf("coin denoms differ %v : %v", msg.TotalRewardPool.Denom, msg.MaxRewardPerClaimAddress.Denom)
	}
	if msg.MaxRewardPerClaimAddress.Amount.GT(msg.TotalRewardPool.Amount) {
		return fmt.Errorf("max claims per address cannot be larger than pool %v : %v", msg.MaxRewardPerClaimAddress.Amount, msg.TotalRewardPool.Amount)
	}
	if msg.RewardPeriodDays < 1 || msg.ClaimPeriodDays < 1 || msg.ExpireDays < 1 {
		return fmt.Errorf("reward period days (%v), claim period days (%v), and expire days (%v) must be larger than 0", msg.RewardPeriodDays, msg.ClaimPeriodDays, msg.ExpireDays)
	}
	if msg.RewardPeriodDays%msg.ClaimPeriodDays != 0 {
		return fmt.Errorf("reward period days (%v) must be multiple of claim period days (%v)", msg.RewardPeriodDays, msg.ClaimPeriodDays)
	}
	if len(msg.QualifyingActions) == 0 {
		return fmt.Errorf("reward program must contain qualifying actions")
	}
	// TODO validate basic should be on top level action
	// for _, action :=  range msg.QualifyingActions {
	// 	action.Vali
	// }
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgCreateRewardProgramRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners indicates that the message must have been signed by the parent.
func (msg MsgCreateRewardProgramRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.DistributeFromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes encodes the message for signing
func (msg MsgCreateRewardProgramRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// NewMsgClaimRewardRequest creates a new reward claim request
func NewMsgClaimRewardRequest(
	rewardProgramID uint64,
	rewardAddress string,
) *MsgClaimRewardRequest {
	return &MsgClaimRewardRequest{
		RewardProgramId: rewardProgramID,
		RewardAddress:   rewardAddress,
	}
}

// Route implements Msg
func (msg MsgClaimRewardRequest) Route() string { return ModuleName }

// Type implements Msg
func (msg MsgClaimRewardRequest) Type() string { return TypeMsgClaimRewardRequest }

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgClaimRewardRequest) ValidateBasic() error {
	if msg.RewardProgramId < 1 {
		return fmt.Errorf("invalid rewards program id : %d", msg.RewardProgramId)
	}
	if _, err := sdk.AccAddressFromBech32(msg.RewardAddress); err != nil {
		return fmt.Errorf("invalid reward address : %w", err)
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgClaimRewardRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners indicates that the message must have been signed by the parent.
func (msg MsgClaimRewardRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.RewardAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}
