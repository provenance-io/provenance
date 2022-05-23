package types

import (
	"errors"
	fmt "fmt"
	"reflect"
	"strings"
	time "time"

	// "github.com/gogo/protobuf/proto"
	"gopkg.in/yaml.v2"

	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultStartingRewardProgramID is 1
const DefaultStartingRewardProgramID uint64 = 1

// Constants pertaining to a RewardProgram object
const (
	MaxDescriptionLength int = 10000
	MaxTitleLength       int = 140
)

var EpochTypeToSeconds = map[string]uint64{
	"day":   60 * 60 * 24,
	"week":  60 * 60 * 24 * 7,
	"month": 60 * 60 * 24 * 30,
}

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
		Evaluate(ctx sdk.Context, provider KeeperProvider, state AccountState, event EvaluationResult) bool
		GetEventCriteria() *EventCriteria
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
	coin sdk.Coin,
	maxRewardByAddress sdk.Coin,
	programStartTime time.Time,
	epochSeconds uint64,
	numberEpochs uint64,
	eligibilityCriteria EligibilityCriteria,
) RewardProgram {
	return RewardProgram{
		Title:                 title,
		Description:           description,
		Id:                    id,
		DistributeFromAddress: distributeFromAddress,
		Coin:                  coin,
		MaxRewardByAddress:    maxRewardByAddress,
		ProgramStartTime:      programStartTime,
		EpochSeconds:          epochSeconds,
		NumberEpochs:          numberEpochs,
		EligibilityCriteria:   eligibilityCriteria,
		Started:               false,
		Finished:              false,
	}
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
	if err := rp.EligibilityCriteria.ValidateBasic(); err != nil {
		return fmt.Errorf("eligibility criteria is not valid: %w", err)
	}
	if !rp.Coin.IsPositive() {
		return fmt.Errorf("reward program requires coins: %v", rp.Coin)
	}
	if !rp.MaxRewardByAddress.IsPositive() {
		return fmt.Errorf("reward program requires positive max reward by address: %v", rp.MaxRewardByAddress)
	}
	if rp.NumberEpochs < 1 {
		return errors.New("reward program number of epochs must be larger than 0")
	}
	return nil
}

func (rp *RewardProgram) String() string {
	out, _ := yaml.Marshal(rp)
	return string(out)
}

// ============ Reward Program Balance ============

func NewRewardProgramBalance(
	rewardProgramID uint64,
	distributionAddress string,
	balance sdk.Coin,
) RewardProgramBalance {
	return RewardProgramBalance{
		RewardProgramId:     rewardProgramID,
		DistributionAddress: distributionAddress,
		Balance:             balance,
	}
}

func (rpb *RewardProgramBalance) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(rpb.DistributionAddress); err != nil {
		return fmt.Errorf("invalid address for reward program balance: %w", err)
	}
	if rpb.RewardProgramId < 1 {
		return errors.New("reward program id must be larger than 0")
	}
	return nil
}

func (rpb *RewardProgramBalance) String() string {
	out, _ := yaml.Marshal(rpb)
	return string(out)
}

// ============ Reward Claim * Not Used * ============

func NewRewardClaim(address string, sharesPerEpochPerRewardsProgram []SharesPerEpochPerRewardsProgram, expired bool) RewardClaim {
	return RewardClaim{
		Address:                 address,
		SharesPerEpochPerReward: sharesPerEpochPerRewardsProgram,
		Expired:                 expired,
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

// ============ Account State ============

func NewAccountState(rewardProgramID, epochID uint64, address string) AccountState {
	return AccountState{
		RewardProgramId: rewardProgramID,
		EpochId:         epochID,
		Address:         address,
		ActionCounter:   0,
	}
}

func (s *AccountState) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(s.Address); err != nil {
		return fmt.Errorf("invalid address for share address: %w", err)
	}
	if id := s.GetRewardProgramId(); id == 0 {
		return fmt.Errorf("reward program id must be greater than 0")
	}
	return nil
}

func (s *AccountState) String() string {
	out, _ := yaml.Marshal(s)
	return string(out)
}

// ============ Share ============

func NewShare(rewardProgramID, epochID uint64, address string, claimed bool, expireTime time.Time, amount int64) Share {
	return Share{
		RewardProgramId: rewardProgramID,
		EpochId:         epochID,
		Address:         address,
		Claimed:         claimed,
		ExpireTime:      expireTime,
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

// ============ Epoch Reward Distribution ============

func NewEpochRewardDistribution(epochID string, rewardProgramID uint64, totalRewardsPool sdk.Coin, totalShares int64, epochEnded bool) EpochRewardDistribution {
	return EpochRewardDistribution{
		EpochId:          epochID,
		RewardProgramId:  rewardProgramID,
		TotalRewardsPool: totalRewardsPool,
		TotalShares:      totalShares,
		EpochEnded:       epochEnded,
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

// ============ Eligibility Criteria ============

func NewEligibilityCriteria(name string, action RewardAction) EligibilityCriteria {
	ec := EligibilityCriteria{Name: name}
	/*err := ec.SetAction(action)
	if err != nil {
		panic(err)
	}*/
	return ec
}

/*func (ec *EligibilityCriteria) SetAction(rewardAction RewardAction) error {
	any, err := codectypes.NewAnyWithValue(rewardAction)
	if err == nil {
		ec.Action = *any
	}
	return err
}*/

func (ec *EligibilityCriteria) GetAction() RewardAction {
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
	return nil
}

func (ec *EligibilityCriteria) String() string {
	out, _ := yaml.Marshal(ec)
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

func (ad *ActionDelegate) GetEventCriteria() *EventCriteria {
	return NewEventCriteria([]ABCIEvent{
		{
			Type: "message",
			Attributes: map[string][]byte{
				"module": []byte("staking"),
			},
		},
		{
			Type:       "delegate",
			Attributes: map[string][]byte{},
		},
		{
			Type:       "create_validator",
			Attributes: map[string][]byte{},
		},
	})
}

func (ad *ActionDelegate) Evaluate(ctx sdk.Context, provider KeeperProvider, state AccountState, event EvaluationResult) bool {
	validator := event.Validator
	delegator := event.Delegator
	delegations := provider.GetStakingKeeper().GetValidatorDelegations(ctx, validator)

	validatorShares := types.NewDec(0)
	delegatorShares := types.NewDec(0)
	for _, delegation := range delegations {
		validatorShares.Add(delegation.GetShares())
		if !delegator.Equals(delegation.GetDelegatorAddr()) {
			continue
		}
		delegatorShares.Add(delegation.GetShares())
	}

	percentage := float64(validatorShares.BigInt().Uint64()) / float64(validatorShares.BigInt().Uint64()) * 100
	hasValidActionCount := state.ActionCounter >= ad.GetMinimumActions() && state.ActionCounter <= ad.GetMaximumActions()
	hasValidDelegationAmount := delegatorShares.BigInt().Uint64() >= ad.GetMinimumDelegationAmount() && delegatorShares.BigInt().Uint64() <= ad.GetMaximumDelegationAmount()
	hasValidDelegationPercentage := percentage >= ad.GetMinimumDelegationPercentage() && percentage <= ad.GetMaximumDelegationPercentage()

	return hasValidActionCount && hasValidDelegationAmount && hasValidDelegationPercentage
}

func (ad *ActionDelegate) String() string {
	out, _ := yaml.Marshal(ad)
	return string(out)
}

// ============ Action Transfer Delegations ============

func NewActionTransferDelegations() ActionTransferDelegations {
	return ActionTransferDelegations{}
}

func (atd *ActionTransferDelegations) ValidateBasic() error {
	return nil
}

func (atd *ActionTransferDelegations) GetEventCriteria() *EventCriteria {
	return NewEventCriteria([]ABCIEvent{
		{
			Type:       "transfer",
			Attributes: map[string][]byte{},
		},
	})
}

func (atd *ActionTransferDelegations) String() string {
	out, _ := yaml.Marshal(atd)
	return string(out)
}

func (atd *ActionTransferDelegations) ActionType() string {
	return ActionTypeDelegate
}

func (atd *ActionTransferDelegations) Evaluate(ctx sdk.Context, provider KeeperProvider, state AccountState, event EvaluationResult) bool {
	// TODO execute all the rules for action?
	return false
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

// ============ SharesPerEpochPerRewardsProgram * Not used * ============

func NewSharesPerEpochPerRewardsProgram(
	rewardProgramID uint64,
	totalShares int64,
	ephemeralActionCount int64,
	latestRecordedEpoch uint64,
	claimed bool,
	expired bool,
	totalRewardsClaimed sdk.Coin,
) SharesPerEpochPerRewardsProgram {
	return SharesPerEpochPerRewardsProgram{
		RewardProgramId:      rewardProgramID,
		TotalShares:          totalShares,
		EphemeralActionCount: ephemeralActionCount,
		LatestRecordedEpoch:  latestRecordedEpoch,
		Claimed:              claimed,
		Expired:              expired,
		TotalRewardClaimed:   totalRewardsClaimed,
	}
}

func (apeprp *SharesPerEpochPerRewardsProgram) ValidateBasic() error {
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
