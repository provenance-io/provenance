package types

import (
	"errors"
	fmt "fmt"
	"reflect"
	"strings"
	time "time"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	provenanceconfig "github.com/provenance-io/provenance/internal/pioconfig"
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
	_ RewardAction = (*ActionDelegate)(nil)
	_ RewardAction = &ActionTransfer{}
	_ RewardAction = &ActionVote{}
)

const (
	ActionTypeDelegate = "ActionDelegate"
	ActionTypeTransfer = "ActionTransfer"
	ActionTypeVote     = "ActionVote"
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

// NewEventCriteria Performs a shallow copy of the map
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
	// for dynamic properties like sender we will never know the value, hence a zero byte check will return true too
	return attribute == nil || reflect.DeepEqual(attribute, value)
}

type EvaluationResult struct {
	EventTypeToSearch string
	AttributeKey      string
	Shares            int64
	Address           sdk.AccAddress // shares to attribute to this address
	Validator         sdk.ValAddress // Address of the validator
	Delegator         sdk.AccAddress // Address of the delegator
	Recipient         sdk.AccAddress // Address of the recipient of the Action, specifically Transfer
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
	rewardClaimExpirationOffset uint64,
	qualifyingActions []QualifyingAction,
) RewardProgram {
	return RewardProgram{
		Title:                       title,
		Description:                 description,
		Id:                          id,
		DistributeFromAddress:       distributeFromAddress,
		TotalRewardPool:             totalRewardPool,
		RemainingPoolBalance:        totalRewardPool,
		ClaimedAmount:               sdk.NewInt64Coin(totalRewardPool.Denom, 0),
		MaxRewardByAddress:          maxRewardByAddress,
		ProgramStartTime:            programStartTime,
		ClaimPeriodSeconds:          claimPeriodSeconds,
		ClaimPeriods:                claimPeriods,
		RewardClaimExpirationOffset: rewardClaimExpirationOffset,
		State:                       RewardProgram_PENDING,
		QualifyingActions:           qualifyingActions,
		MinimumRolloverAmount:       sdk.NewInt64Coin(totalRewardPool.Denom, 100_000_000_000),
	}
}

// TODO Test this
func (rp *RewardProgram) IsStarting(ctx sdk.Context) bool {
	blockTime := ctx.BlockTime()
	return rp.State == RewardProgram_PENDING && (blockTime.After(rp.ProgramStartTime) || blockTime.Equal(rp.ProgramStartTime))
}

// TODO Test this
func (rp *RewardProgram) IsEndingClaimPeriod(ctx sdk.Context) bool {
	blockTime := ctx.BlockTime()
	return rp.State == RewardProgram_STARTED && (blockTime.After(rp.ClaimPeriodEndTime) || blockTime.Equal(rp.ClaimPeriodEndTime))
}

// TODO Test this
func (rp *RewardProgram) IsExpiring(ctx sdk.Context) bool {
	blockTime := ctx.BlockTime()
	expireTime := rp.ActualProgramEndTime.Add(time.Second * time.Duration(rp.RewardClaimExpirationOffset))
	return rp.State == RewardProgram_FINISHED && (blockTime.After(expireTime) || blockTime.Equal(expireTime))
}

// TODO Test this
func (rp *RewardProgram) IsEnding(ctx sdk.Context, programBalance sdk.Coin) bool {
	blockTime := ctx.BlockTime()
	isProgramExpired := !rp.GetExpectedProgramEndTime().IsZero() && (blockTime.After(rp.ExpectedProgramEndTime) || blockTime.Equal(rp.ExpectedProgramEndTime))
	canRollover := programBalance.IsGTE(rp.GetMinimumRolloverAmount())
	return rp.State == RewardProgram_STARTED && (isProgramExpired || !canRollover)
}

func (rp *RewardProgram) Validate() error {
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
	if rp.TotalRewardPool.Denom != rp.RemainingPoolBalance.Denom && rp.RemainingPoolBalance.Denom != rp.MaxRewardByAddress.Denom {
		return fmt.Errorf("all denoms must be same for total reward pool (%s) remaining reward pool (%s) and max reward by address (%s)", rp.TotalRewardPool.Denom, rp.RemainingPoolBalance.Denom, rp.MaxRewardByAddress.Denom)
	}
	if rp.TotalRewardPool.Denom != "nhash" {
		return fmt.Errorf("reward program denom must be in %s : %s", rp.TotalRewardPool.Denom, "nhash")
	}

	if len(rp.QualifyingActions) == 0 {
		return errors.New("reward program must have at least one qualifying action")
	}

	return nil
}

func (rp *RewardProgram) String() string {
	out, _ := yaml.Marshal(rp)
	return string(out)
}

// ============ Account State ============

func NewRewardAccountState(rewardProgramID, rewardClaimPeriodID uint64, address string, shares uint64) RewardAccountState {
	return RewardAccountState{
		RewardProgramId: rewardProgramID,
		ClaimPeriodId:   rewardClaimPeriodID,
		Address:         address,
		ActionCounter:   map[string]uint64{},
		SharesEarned:    shares,
		ClaimStatus:     RewardAccountState_UNCLAIMABLE,
	}
}

func (s *RewardAccountState) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(s.Address); err != nil {
		return fmt.Errorf("invalid address for share address: %w", err)
	}
	if id := s.GetRewardProgramId(); id == 0 {
		return fmt.Errorf("reward program id must be greater than 0")
	}
	if claimID := s.GetClaimPeriodId(); claimID == 0 {
		return fmt.Errorf("claim period id must be greater than 0")
	}

	return nil
}

func (s *RewardAccountState) String() string {
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
	if !erd.TotalRewardsPoolForClaimPeriod.IsPositive() && !erd.TotalRewardsPoolForClaimPeriod.IsZero() {
		return errors.New("claim reward distribution must have a total reward pool")
	}
	if !erd.RewardsPool.IsPositive() {
		return errors.New("claim reward distribution must have a reward pool which is positive")
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
	if ad.MinimumDelegationAmount != nil && ad.MaximumDelegationAmount == nil {

	}
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
	delegation, found := stakingKeeper.GetDelegation(ctx, delegator, validatorAddr)
	if !found {
		return sdk.NewDec(0), found
	}
	validator, found := stakingKeeper.GetValidator(ctx, validatorAddr)
	if !found {
		return sdk.NewDec(0), found
	}
	tokens := validator.TokensFromShares(delegation.GetShares())
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
	actionCounter := state.ActionCounter[ad.ActionType()]

	// TODO Is this correct to truncate the tokens?
	delegatedHash := sdk.NewInt64Coin(provenanceconfig.DefaultBondDenom, tokens.TruncateInt64())
	minDelegation := ad.GetMinimumDelegationAmount()
	maxDelegation := ad.GetMaximumDelegationAmount()
	minPercentile := ad.GetMinimumActiveStakePercentile()
	maxPercentile := ad.GetMaximumActiveStakePercentile()

	hasValidActionCount := actionCounter >= ad.GetMinimumActions() && actionCounter <= ad.GetMaximumActions()
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

func NewActionTransfer() ActionTransfer {
	return ActionTransfer{}
}

func (at *ActionTransfer) ValidateBasic() error {
	return nil
}

func (at *ActionTransfer) GetBuilder() ActionBuilder {
	return &TransferActionBuilder{}
}

func (at *ActionTransfer) String() string {
	out, _ := yaml.Marshal(at)
	return string(out)
}

func (at *ActionTransfer) ActionType() string {
	return ActionTypeTransfer
}

func (at *ActionTransfer) Evaluate(ctx sdk.Context, provider KeeperProvider, state RewardAccountState, event EvaluationResult) bool {
	// get the address that is performing the send
	addressSender := event.Address
	if addressSender == nil {
		return false
	}
	if provider.GetAccountKeeper().GetModuleAddress(authtypes.FeeCollectorName).Equals(event.Recipient) {
		return false
	}
	actionCounter := state.ActionCounter[at.ActionType()]
	hasValidActionCount := actionCounter >= at.GetMinimumActions() && actionCounter <= at.GetMaximumActions()
	if !hasValidActionCount {
		return false
	}

	println(at.MinimumDelegationAmount.String())
	// check delegations if and only if mandated by the Action
	if sdk.NewCoin(provenanceconfig.DefaultBondDenom, sdk.ZeroInt()).IsLT(at.MinimumDelegationAmount) {
		// now check if it has any delegations
		totalDelegations, found := getAllDelegations(ctx, provider, addressSender)
		if !found {
			return false
		}
		if totalDelegations.IsGTE(at.MinimumDelegationAmount) {
			return true
		}
	}
	return true
}

func (at *ActionTransfer) PreEvaluate(ctx sdk.Context, provider KeeperProvider, state *RewardAccountState) {
	// Any action specific thing that we need to do before evaluation
}

func (at *ActionTransfer) PostEvaluate(ctx sdk.Context, provider KeeperProvider, state *RewardAccountState) {
	// Any action specific thing that we need to do after evaluation
}

// ============ Action Vote  ============

func NewActionVote() ActionVote {
	return ActionVote{}
}

func (atd *ActionVote) ValidateBasic() error {
	return nil
}

func (atd *ActionVote) GetBuilder() ActionBuilder {
	return &VoteActionBuilder{}
}

func (atd *ActionVote) String() string {
	out, _ := yaml.Marshal(atd)
	return string(out)
}

func (atd *ActionVote) ActionType() string {
	return ActionTypeVote
}

func (atd *ActionVote) Evaluate(ctx sdk.Context, provider KeeperProvider, state RewardAccountState, event EvaluationResult) bool {
	actionCounter := state.ActionCounter[atd.ActionType()]
	hasValidActionCount := actionCounter >= atd.GetMinimumActions() && actionCounter <= atd.GetMaximumActions()
	if !hasValidActionCount {
		return false
	}

	// get the address that voted
	addressVoting := event.Address
	if atd.MinimumDelegationAmount.IsGTE(sdk.NewCoin(provenanceconfig.DefaultBondDenom, sdk.ZeroInt())) {
		// now check if it has any delegations

		totalDelegations, found := getAllDelegations(ctx, provider, addressVoting)
		if !found {
			return false
		}
		if totalDelegations.IsGTE(atd.MinimumDelegationAmount) {
			return true
		}
	}
	return false
}

func (atd *ActionVote) PreEvaluate(ctx sdk.Context, provider KeeperProvider, state *RewardAccountState) {
	// Any action specific thing that we need to do before evaluation
}

func (atd *ActionVote) PostEvaluate(ctx sdk.Context, provider KeeperProvider, state *RewardAccountState) {
	// Any action specific thing that we need to do after evaluation
}

// ============ Qualifying Action ============

func (qa *QualifyingAction) GetRewardAction(ctx sdk.Context) (RewardAction, error) {
	var action RewardAction

	switch actionType := qa.GetType().(type) {
	case *QualifyingAction_Delegate:
		action = qa.GetDelegate()
	case *QualifyingAction_Transfer:
		action = qa.GetTransfer()
	case *QualifyingAction_Vote:
		action = qa.GetVote()
	default:
		// Skip any unsupported actions
		message := fmt.Sprintf("The Action type %s is not supported", actionType)
		return nil, errors.New(message)
	}

	ctx.Logger().Info(fmt.Sprintf("NOTICE: The Action type is %s", action.ActionType()))

	return action, nil
}

// getAllDelegations pure functions to get delegated coins for an address
// return total coin delegated and boolean to indicate if any delegations are at all present.
func getAllDelegations(ctx sdk.Context, provider KeeperProvider, delegator sdk.AccAddress) (sdk.Coin, bool) {
	stakingKeeper := provider.GetStakingKeeper()
	delegations := stakingKeeper.GetAllDelegatorDelegations(ctx, delegator)
	// if no delegations then return not found
	if len(delegations) == 0 {
		return sdk.NewCoin(provenanceconfig.DefaultBondDenom, sdk.ZeroInt()), false
	}

	sum := sdk.NewCoin(provenanceconfig.DefaultBondDenom, sdk.ZeroInt())

	for _, delegation := range delegations {
		val, found := stakingKeeper.GetValidator(ctx, delegation.GetValidatorAddr())

		if found {
			tokens := val.TokensFromShares(delegation.GetShares()).TruncateInt()
			sum = sum.Add(sdk.NewCoin(provenanceconfig.DefaultBondDenom, tokens))
		}
	}

	if sum.Amount.Equal(sdk.ZeroInt()) {
		return sum, false
	}
	return sum, true
}
