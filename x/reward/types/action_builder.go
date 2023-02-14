package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	_ ActionBuilder = &TransferActionBuilder{}
	_ ActionBuilder = &DelegateActionBuilder{}
	_ ActionBuilder = &VoteActionBuilder{}
)

// ActionBuilder defines functions used to collect events to check against specific actions.
type ActionBuilder interface {
	// GetEventCriteria returns the event criteria for this action.
	GetEventCriteria() *EventCriteria
	// AddEvent adds an event to this builder in preparation for checking events against the criteria.
	AddEvent(eventType string, attributes *map[string][]byte) error
	// CanBuild returns true if this builder has enough event information for an evaluation result.
	CanBuild() bool
	// BuildAction builds the action and returns an EvaluationResult.
	// This should only be called when CanBuild returns true.
	// This should return an error if CanBuild is false or some other error is encountered.
	BuildAction() (EvaluationResult, error)
	// Reset clears out any previous event data, returning the builder to it's initial state.
	Reset()
}

type TransferActionBuilder struct {
	Sender    sdk.AccAddress
	Action    string
	Recipient sdk.AccAddress
}

func (b *TransferActionBuilder) GetEventCriteria() *EventCriteria {
	return NewEventCriteria([]ABCIEvent{
		{
			Type: sdk.EventTypeMessage,
			Attributes: map[string][]byte{
				sdk.AttributeKeyAction: []byte(sdk.MsgTypeURL(&banktypes.MsgSend{})),
			},
		},
		{
			Type:       banktypes.EventTypeTransfer,
			Attributes: map[string][]byte{},
		},
	})
}

func (b *TransferActionBuilder) AddEvent(eventType string, attributes *map[string][]byte) error {
	switch eventType {
	case sdk.EventTypeMessage:
		// Update the result with the senders address
		address := (*attributes)[banktypes.AttributeKeySender]
		action := (*attributes)[sdk.AttributeKeyAction]
		if address != nil {
			address, err := sdk.AccAddressFromBech32(string(address))
			if err != nil {
				return err
			}
			b.Sender = address
		}

		if action != nil {
			b.Action = string(action)
		}
	case banktypes.EventTypeTransfer:
		// Update the result with the senders address
		address := (*attributes)[banktypes.AttributeKeySender]
		addressSenderAddr, err := sdk.AccAddressFromBech32(string(address))
		if err != nil {
			return err
		}
		b.Sender = addressSenderAddr

		addressRecipient := (*attributes)[banktypes.AttributeKeyRecipient]
		addressRecipientAddr, errFromParsingRecipientAddress := sdk.AccAddressFromBech32(string(addressRecipient))
		if errFromParsingRecipientAddress != nil {
			return errFromParsingRecipientAddress
		}
		b.Recipient = addressRecipientAddr
	}
	return nil
}

func (b *TransferActionBuilder) CanBuild() bool {
	return !b.Sender.Empty() && !(len(b.Action) == 0) && !b.Recipient.Empty()
}

func (b *TransferActionBuilder) BuildAction() (EvaluationResult, error) {
	if !b.CanBuild() {
		return EvaluationResult{}, fmt.Errorf("missing sender or recipient or action from transfer action")
	}

	result := EvaluationResult{
		Shares:    1,
		Address:   b.Sender,
		Recipient: b.Recipient,
	}

	return result, nil
}

func (b *TransferActionBuilder) Reset() {
	b.Sender = sdk.AccAddress{}
}

type DelegateActionBuilder struct {
	Validator sdk.ValAddress
	Delegator sdk.AccAddress
}

func (b *DelegateActionBuilder) GetEventCriteria() *EventCriteria {
	return NewEventCriteria([]ABCIEvent{
		{
			Type: sdk.EventTypeMessage,
			Attributes: map[string][]byte{
				sdk.AttributeKeyModule: []byte(stakingtypes.ModuleName),
			},
		},
		{
			Type:       stakingtypes.EventTypeDelegate,
			Attributes: map[string][]byte{},
		},
		{
			Type:       stakingtypes.EventTypeCreateValidator,
			Attributes: map[string][]byte{},
		},
	})
}

func (b *DelegateActionBuilder) AddEvent(eventType string, attributes *map[string][]byte) error {
	switch eventType {
	case stakingtypes.EventTypeDelegate:
		address := (*attributes)[stakingtypes.AttributeKeyValidator]
		validator, err := sdk.ValAddressFromBech32(string(address))
		if err != nil {
			return err
		}
		b.Validator = validator
	case stakingtypes.EventTypeCreateValidator:
		address := (*attributes)[stakingtypes.AttributeKeyValidator]
		validator, err := sdk.ValAddressFromBech32(string(address))
		if err != nil {
			return err
		}
		b.Validator = validator
	case sdk.EventTypeMessage:
		// Update the last result to have the delegator's address
		address := (*attributes)[banktypes.AttributeKeySender]
		address, err := sdk.AccAddressFromBech32(string(address))
		if err != nil {
			return err
		}
		b.Delegator = address
	}

	return nil
}

func (b *DelegateActionBuilder) CanBuild() bool {
	return !b.Validator.Empty() && !b.Delegator.Empty()
}

func (b *DelegateActionBuilder) BuildAction() (EvaluationResult, error) {
	if !b.CanBuild() {
		return EvaluationResult{}, fmt.Errorf("missing delegator or validator from delegate action")
	}

	result := EvaluationResult{
		Shares:    1,
		Address:   b.Delegator,
		Delegator: b.Delegator,
		Validator: b.Validator,
	}

	return result, nil
}

func (b *DelegateActionBuilder) Reset() {
	b.Validator = sdk.ValAddress{}
	b.Delegator = sdk.AccAddress{}
}

type VoteActionBuilder struct {
	Voter sdk.AccAddress
	Voted bool
}

func (v *VoteActionBuilder) CanBuild() bool {
	return !v.Voter.Empty() && v.Voted
}

func (v *VoteActionBuilder) BuildAction() (EvaluationResult, error) {
	if !v.CanBuild() {
		return EvaluationResult{}, fmt.Errorf("missing voter address from vote action")
	}

	result := EvaluationResult{
		Shares:  1,
		Address: v.Voter,
	}

	return result, nil
}

func (v *VoteActionBuilder) Reset() {
	v.Voter = sdk.AccAddress{}
	v.Voted = false
}

func (v *VoteActionBuilder) GetEventCriteria() *EventCriteria {
	return NewEventCriteria([]ABCIEvent{
		{
			Type:       sdk.EventTypeMessage,
			Attributes: map[string][]byte{sdk.AttributeKeyModule: []byte(govtypes.AttributeValueCategory)},
		},
		{
			Type:       sdk.EventTypeMessage,
			Attributes: map[string][]byte{sdk.AttributeKeyAction: nil},
		},
	})
}

// govVoteMsgURLs the MsgURLs of all the governance module's vote messages.
// Use getGovVoteMsgURLs() instead of using this variable directly.
var govVoteMsgURLs []string

// getGovVoteMsgURLs returns govVoteMsgURLs, but first sets it if it hasn't yet been set.
func getGovVoteMsgURLs() []string {
	// Checking for nil here (as opposed to len == 0) because we only want to set it
	// if it hasn't been set yet.
	if govVoteMsgURLs == nil {
		// sdk.MsgTypeURL sometimes uses reflection and/or proto registration.
		// So govVoteMsgURLs is only set when it's finally needed in the hopes
		// that everything's wired up as needed by then.
		govVoteMsgURLs = []string{
			sdk.MsgTypeURL(&govtypesv1.MsgVote{}),
			sdk.MsgTypeURL(&govtypesv1.MsgVoteWeighted{}),
			sdk.MsgTypeURL(&govtypesv1beta1.MsgVote{}),
			sdk.MsgTypeURL(&govtypesv1beta1.MsgVoteWeighted{}),
		}
	}
	return govVoteMsgURLs
}

func (v *VoteActionBuilder) AddEvent(eventType string, attributes *map[string][]byte) error {
	if _, ok := (*attributes)[sdk.AttributeKeyModule]; ok {
		address := (*attributes)[banktypes.AttributeKeySender]
		address, err := sdk.AccAddressFromBech32(string(address))
		if err != nil {
			return err
		}
		v.Voter = address
	} else if action, ok := (*attributes)[sdk.AttributeKeyAction]; ok {
		a := string(action)
		for _, m := range getGovVoteMsgURLs() {
			if a == m {
				v.Voted = true
				break
			}
		}
	}

	return nil
}
