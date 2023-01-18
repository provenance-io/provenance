package types

import (
	"errors"
	fmt "fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Compile time interface checks.
var _ sdk.Msg = &MsgCreateRewardProgramRequest{}
var _ sdk.Msg = &MsgEndRewardProgramRequest{}
var _ sdk.Msg = &MsgClaimRewardsRequest{}
var _ sdk.Msg = &MsgClaimAllRewardsRequest{}

// NewMsgCreateRewardProgramRequest creates a new create reward program request
func NewMsgCreateRewardProgramRequest(
	title string,
	description string,
	distributeFromAddress string,
	totalRewardPool sdk.Coin,
	maxRewardPerClaimAddress sdk.Coin,
	programStartTime time.Time,
	claimPeriods uint64,
	claimPeriodDays uint64,
	maxRolloverClaimPeriods uint64,
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
		ClaimPeriods:             claimPeriods,
		ClaimPeriodDays:          claimPeriodDays,
		MaxRolloverClaimPeriods:  maxRolloverClaimPeriods,
		ExpireDays:               expireDays,
		QualifyingActions:        qualifyingAction,
	}
}

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
	if msg.ClaimPeriods < 1 || msg.ClaimPeriodDays < 1 || msg.ExpireDays < 1 {
		return fmt.Errorf("claim periods (%v), claim period days (%v), and expire days (%v) must be larger than 0", msg.ClaimPeriods, msg.ClaimPeriodDays, msg.ExpireDays)
	}
	if len(msg.QualifyingActions) == 0 {
		return fmt.Errorf("reward program must contain qualifying actions")
	}
	for _, action := range msg.QualifyingActions {
		if err := action.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// GetSigners indicates that the message must have been signed by the parent.
func (msg MsgCreateRewardProgramRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.DistributeFromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// NewMsgEndRewardProgramRequest ends a reward program request
func NewMsgEndRewardProgramRequest(
	rewardProgramID uint64,
	programOwnerAddress string) *MsgEndRewardProgramRequest {
	return &MsgEndRewardProgramRequest{
		RewardProgramId:     rewardProgramID,
		ProgramOwnerAddress: programOwnerAddress,
	}
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgEndRewardProgramRequest) ValidateBasic() error {
	if msg.RewardProgramId < 1 {
		return fmt.Errorf("invalid reward program id: %v", msg.RewardProgramId)
	}
	if _, err := sdk.AccAddressFromBech32(msg.ProgramOwnerAddress); err != nil {
		return fmt.Errorf("invalid address for rewards program distribution from address: %w", err)
	}
	return nil
}

// GetSigners indicates that the message must have been signed by the parent.
func (msg MsgEndRewardProgramRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.ProgramOwnerAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// NewMsgClaimRewardsRequest creates a new reward claim request
func NewMsgClaimRewardsRequest(
	rewardProgramID uint64,
	rewardAddress string,
) *MsgClaimRewardsRequest {
	return &MsgClaimRewardsRequest{
		RewardProgramId: rewardProgramID,
		RewardAddress:   rewardAddress,
	}
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgClaimRewardsRequest) ValidateBasic() error {
	if msg.RewardProgramId < 1 {
		return fmt.Errorf("invalid rewards program id : %d", msg.RewardProgramId)
	}
	if _, err := sdk.AccAddressFromBech32(msg.RewardAddress); err != nil {
		return fmt.Errorf("invalid reward address : %w", err)
	}
	return nil
}

// GetSigners indicates that the message must have been signed by the parent.
func (msg MsgClaimRewardsRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.RewardAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// NewMsgClaimRewardsRequest creates a new claim all request
func NewMsgClaimAllRewardsRequest(
	rewardAddress string,
) *MsgClaimAllRewardsRequest {
	return &MsgClaimAllRewardsRequest{
		RewardAddress: rewardAddress,
	}
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgClaimAllRewardsRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.RewardAddress); err != nil {
		return fmt.Errorf("invalid reward address : %w", err)
	}
	return nil
}

// GetSigners indicates that the message must have been signed by the parent.
func (msg MsgClaimAllRewardsRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.RewardAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}
