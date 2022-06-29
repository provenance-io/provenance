package types

import (
	"errors"
	fmt "fmt"
	"reflect"
	"strings"
	time "time"

	// "github.com/gogo/protobuf/proto"
	"gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultStartingRewardProgramID is 1
const DefaultStartingRewardProgramID uint64 = 1

// Constants pertaining to a RewardProgram object
const (
	MaxDescriptionLength int = 10000
	MaxTitleLength       int = 140
	DayInSeconds         int = 60 * 60 * 24
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
		ActionType() string
		Evaluate(ctx sdk.Context, provider KeeperProvider, state RewardAccountState, event EvaluationResult) bool
		PreEvaluate(ctx sdk.Context, provider KeeperProvider, state *RewardAccountState)
		PostEvaluate(ctx sdk.Context, provider KeeperProvider, state *RewardAccountState)
		GetBuilder() ActionBuilder
	}
)

// ============ Shared structs ============

type ABCIEvent struct {
	Type       string
	Attributes map[string][]byte
}

type EventCriteria struct {
	Events map[string]ABCIEvent
}

// Performs a shallow copy of the map
// We are assuming this takes ownership of events
func NewEventCriteria(events []ABCIEvent) *EventCriteria {
	criteria := EventCriteria{}
	if len(events) == 0 {
		return &criteria
	}
	criteria.Events = make(map[string]ABCIEvent)
	for _, event := range events {
		criteria.Events[event.Type] = event
	}
	return &criteria
}

func (ec *EventCriteria) MatchesEvent(eventType string) bool {
	// If we have no Events then we match everything
	if ec.Events == nil {
		return true
	}

	// If we don't have the event then we don't match it
	_, exists := ec.Events[eventType]
	return exists
}

func (ec *ABCIEvent) MatchesAttribute(name string, value []byte) bool {
	attribute, exists := ec.Attributes[name]
	if !exists {
		return false
	}
	return attribute == nil || reflect.DeepEqual(attribute, value)
}

type EvaluationResult struct {
	EventTypeToSearch string
	AttributeKey      string
	Shares            int64
	Address           sdk.AccAddress // shares to attribute to this address
	Validator         sdk.ValAddress // Address of the validator
	Delegator         sdk.AccAddress // Address of the delegator
}

// ============ Reward Program ============

func NewRewardProgram(
	title string,
	description string,
	id uint64,
	distributeFromAddress string,
	totalRewardPool sdk.Coin,
	maxRewardByAddress sdk.Coin,
	programStartTime time.Time,
	claimPeriodSeconds uint64,
	claimPeriods uint64,
	shareExpirationOffset uint64,
	qualifyingActions []QualifyingAction,
) RewardProgram {
	return RewardProgram{
		Title:                 title,
		Description:           description,
		Id:                    id,
		DistributeFromAddress: distributeFromAddress,
		TotalRewardPool:       totalRewardPool,
		RemainingPoolBalance:  totalRewardPool,
		MaxRewardByAddress:    maxRewardByAddress,
		ProgramStartTime:      programStartTime,
		ClaimPeriodSeconds:    claimPeriodSeconds,
		ClaimPeriods:          claimPeriods,
		ShareExpirationOffset: shareExpirationOffset,
		State:                 RewardProgram_PENDING,
		QualifyingActions:     qualifyingActions,
		MinimumRolloverAmount: sdk.NewInt64Coin("nhash", 100_000_000_000),
	}
}

func (rp *RewardProgram) IsStarting(ctx sdk.Context) bool {
	blockTime := ctx.BlockTime()
	return rp.State == RewardProgram_PENDING && (blockTime.After(rp.ProgramStartTime) || blockTime.Equal(rp.ProgramStartTime))
}

func (rp *RewardProgram) IsEndingClaimPeriod(ctx sdk.Context) bool {
	blockTime := ctx.BlockTime()
	return rp.State == RewardProgram_STARTED && (blockTime.After(rp.ClaimPeriodEndTime) || blockTime.Equal(rp.ClaimPeriodEndTime))
}

func (rp *RewardProgram) IsEnding(ctx sdk.Context, programBalance sdk.Coin) bool {
	blockTime := ctx.BlockTime()
	isProgramExpired := !rp.GetExpectedProgramEndTime().IsZero() && (blockTime.After(rp.ExpectedProgramEndTime) || blockTime.Equal(rp.ExpectedProgramEndTime))
	canRollover := programBalance.IsGTE(rp.GetMinimumRolloverAmount())
	return rp.State == RewardProgram_STARTED && (isProgramExpired || !canRollover)
}

func (rp *RewardProgram) ValidateBasic() error {
	title := rp.GetTitle()
	if len(strings.TrimSpace(title)) == 0 {
		return errors.New("reward program title cannot be blank")
	}
	if len(title) > MaxTitleLength {
		return fmt.Errorf("reward program title is longer than max length of %d", MaxTitleLength)
	}
	description := rp.GetDescription()
	if len(description) == 0 {
		return errors.New("reward program description cannot be blank")
	}
	if len(description) > MaxDescriptionLength {
		return fmt.Errorf("reward program description is longer than max length of %d", MaxDescriptionLength)
	}
	if rp.Id < 1 {
		return errors.New("reward program id must be larger than 0")
	}
	if _, err := sdk.AccAddressFromBech32(rp.DistributeFromAddress); err != nil {
		return fmt.Errorf("invalid address for rewards program distribution from address: %w", err)
	}
	if !rp.TotalRewardPool.IsPositive() {
		return fmt.Errorf("reward program requires coins: %v", rp.TotalRewardPool)
	}
	if !rp.MaxRewardByAddress.IsPositive() {
		return fmt.Errorf("reward program requires positive max reward by address: %v", rp.MaxRewardByAddress)
	}
	if rp.ClaimPeriods < 1 {
		return errors.New("reward program number of claim periods must be larger than 0")
	}

	return nil
}

func (rp *RewardProgram) String() string {
	out, _ := yaml.Marshal(rp)
	return string(out)
}

// ============ Account State ============

func NewRewardAccountState(rewardProgramID, rewardClaimPeriodID uint64, address string) RewardAccountState {
	return RewardAccountState{
		RewardProgramId: rewardProgramID,
		ClaimPeriodId:   rewardClaimPeriodID,
		Address:         address,
		ActionCounter:   0,
	}
}

func (s *RewardAccountState) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(s.Address); err != nil {
		return fmt.Errorf("invalid address for share address: %w", err)
	}
	if id := s.GetRewardProgramId(); id == 0 {
		return fmt.Errorf("reward program id must be greater than 0")
	}
	return nil
}

func (s *RewardAccountState) String() string {
	out, _ := yaml.Marshal(s)
	return string(out)
}

// ============ Share ============

func NewShare(rewardProgramID, claimPeriodID uint64, address string, amount int64) Share {
	return Share{
		RewardProgramId: rewardProgramID,
		ClaimPeriodId:   claimPeriodID,
		Address:         address,
		Amount:          amount,
	}
}

func (s *Share) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(s.Address); err != nil {
		return fmt.Errorf("invalid address for share address: %w", err)
	}
	if id := s.GetRewardProgramId(); id == 0 {
		return fmt.Errorf("invalid reward program id")
	}
	return nil
}

func (s *Share) String() string {
	out, _ := yaml.Marshal(s)
	return string(out)
}

// ============ Claim Period Reward Distribution ============

func NewClaimPeriodRewardDistribution(claimPeriodID uint64, rewardProgramID uint64, rewardsPool, totalRewardsPoolForClaimPeriod sdk.Coin, totalShares int64, claimPeriodEnded bool) ClaimPeriodRewardDistribution {
	return ClaimPeriodRewardDistribution{
		ClaimPeriodId:                  claimPeriodID,
		RewardProgramId:                rewardProgramID,
		RewardsPool:                    rewardsPool,
		TotalRewardsPoolForClaimPeriod: totalRewardsPoolForClaimPeriod,
		TotalShares:                    totalShares,
		ClaimPeriodEnded:               claimPeriodEnded,
	}
}

func (erd *ClaimPeriodRewardDistribution) ValidateBasic() error {
	if erd.ClaimPeriodId <= 0 {
		return errors.New("claim reward distribution has invalid claim id")
	}
	if erd.RewardProgramId < 1 {
		return errors.New("claim reward distribution must have a valid reward program id")
	}
	if !erd.TotalRewardsPoolForClaimPeriod.IsPositive() {
		return errors.New("claim reward distribution must have a total reward pool")
	}
	if !erd.RewardsPool.IsPositive() {
		return errors.New("claim reward distribution must have a reward pool")
	}
	return nil
}

func (erd *ClaimPeriodRewardDistribution) String() string {
	out, _ := yaml.Marshal(erd)
	return string(out)
}

// ============ Action Delegate ============

func NewActionDelegate() ActionDelegate {
	return ActionDelegate{}
}

func (ad *ActionDelegate) ValidateBasic() error {
	return nil
}

func (ad *ActionDelegate) ActionType() string {
	return ActionTypeDelegate
}

func (ad *ActionDelegate) GetBuilder() ActionBuilder {
	return &DelegateActionBuilder{}
}

func (ad *ActionDelegate) getTokensFromValidator(ctx sdk.Context, provider KeeperProvider, validatorAddr sdk.ValAddress, delegator sdk.AccAddress) (sdk.Dec, bool) {
	stakingKeeper := provider.GetStakingKeeper()
	delegations := stakingKeeper.GetValidatorDelegations(ctx, validatorAddr)
	delegatorShares := sdk.NewDec(0)
	for _, delegation := range delegations {
		if !delegator.Equals(delegation.GetDelegatorAddr()) {
			continue
		}
		shares := delegation.GetShares()
		delegatorShares = delegatorShares.Add(shares)
	}

	validator, found := stakingKeeper.GetValidator(ctx, validatorAddr)
	if !found {
		return sdk.NewDec(0), found
	}
	tokens := validator.TokensFromShares(delegatorShares)
	return tokens, found
}

// The percentile is dictated by its placement in the BondedValidator list
// If there are 5 validators and the first validator matches then that validator is in the top 80%
func (ad *ActionDelegate) getValidatorRankPercentile(ctx sdk.Context, provider KeeperProvider, validator sdk.ValAddress) sdk.Dec {
	validators := provider.GetStakingKeeper().GetBondedValidatorsByPower(ctx)
	numValidators := int64(len(validators))
	rank := numValidators
	for i := int64(0); i < numValidators; i++ {
		v := validators[i]
		validatorString := validator.String()
		if v.OperatorAddress == validatorString {
			rank = i + 1
			break
		}
	}
	placement := sdk.NewDec(numValidators - rank)
	vals := sdk.NewDec(numValidators)
	percentile := placement.Quo(vals)

	return percentile
}

func (ad *ActionDelegate) Evaluate(ctx sdk.Context, provider KeeperProvider, state RewardAccountState, event EvaluationResult) bool {
	validator := event.Validator
	delegator := event.Delegator

	tokens, found := ad.getTokensFromValidator(ctx, provider, validator, delegator)
	if !found {
		return false
	}
	percentile := ad.getValidatorRankPercentile(ctx, provider, validator)

	hasValidActionCount := state.ActionCounter >= ad.GetMinimumActions() && state.ActionCounter <= ad.GetMaximumActions()

	// TODO Is this correct to truncate the tokens?
	delegatedHash := sdk.NewInt64Coin("nhash", tokens.TruncateInt64())
	minDelegation := ad.GetMinimumDelegationAmount()
	maxDelegation := ad.GetMaximumDelegationAmount()
	minPercentile := ad.GetMinimumActiveStakePercentile()
	maxPercentile := ad.GetMaximumActiveStakePercentile()

	hasValidDelegationAmount := delegatedHash.IsGTE(*minDelegation) && (delegatedHash.IsLT(*maxDelegation) || delegatedHash.IsEqual(*maxDelegation))
	hasValidActivePercentile := percentile.GTE(minPercentile) && percentile.LTE(maxPercentile)

	return hasValidActionCount && hasValidDelegationAmount && hasValidActivePercentile
}

func (ad *ActionDelegate) String() string {
	out, _ := yaml.Marshal(ad)
	return string(out)
}

func (ad *ActionDelegate) GetMinimumActiveStakePercentile() sdk.Dec {
	if ad != nil {
		return ad.MinimumActiveStakePercentile
	}
	return sdk.NewDec(0)
}

func (ad *ActionDelegate) GetMaximumActiveStakePercentile() sdk.Dec {
	if ad != nil {
		return ad.MaximumActiveStakePercentile
	}
	return sdk.NewDec(0)
}

func (ad *ActionDelegate) PreEvaluate(ctx sdk.Context, provider KeeperProvider, state *RewardAccountState) {
	// Any action specific thing that we need to do before evaluation
}

func (ad *ActionDelegate) PostEvaluate(ctx sdk.Context, provider KeeperProvider, state *RewardAccountState) {
	// Any action specific thing that we need to do after evaluation
}

// ============ Action Transfer Delegations ============

func NewActionTransferDelegations() ActionTransferDelegations {
	return ActionTransferDelegations{}
}

func (atd *ActionTransferDelegations) ValidateBasic() error {
	return nil
}

func (atd *ActionTransferDelegations) GetBuilder() ActionBuilder {
	return &DelegateTransferActionBuilder{}
}

func (atd *ActionTransferDelegations) String() string {
	out, _ := yaml.Marshal(atd)
	return string(out)
}

func (atd *ActionTransferDelegations) ActionType() string {
	return ActionTypeDelegate
}

func (atd *ActionTransferDelegations) Evaluate(ctx sdk.Context, provider KeeperProvider, state RewardAccountState, event EvaluationResult) bool {
	// TODO execute all the rules for action?
	return false
}

func (atd *ActionTransferDelegations) PreEvaluate(ctx sdk.Context, provider KeeperProvider, state *RewardAccountState) {
	// Any action specific thing that we need to do before evaluation
}

func (atd *ActionTransferDelegations) PostEvaluate(ctx sdk.Context, provider KeeperProvider, state *RewardAccountState) {
	// Any action specific thing that we need to do after evaluation
}

// ============ Qualifying Action ============

func (qa *QualifyingAction) GetRewardAction(ctx sdk.Context) (RewardAction, error) {
	var action RewardAction

	switch actionType := qa.GetType().(type) {
	case *QualifyingAction_Delegate:
		action = qa.GetDelegate()
	case *QualifyingAction_TransferDelegations:
		action = qa.GetTransferDelegations()
	default:
		// Skip any unsupported actions
		message := fmt.Sprintf("The Action type %s is not supported", actionType)
		return nil, errors.New(message)
	}

	ctx.Logger().Info(fmt.Sprintf("NOTICE: The Action type is %s", action.ActionType()))

	return action, nil
}
