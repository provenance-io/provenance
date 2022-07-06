package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

type ActionBuilder interface {
	GetEventCriteria() *EventCriteria
	AddEvent(eventType string, attributes *map[string][]byte) error
	CanBuild() bool
	BuildAction() (EvaluationResult, error)
	Reset()
}

// ============ TransferActionBuilder ============
type TransferActionBuilder struct {
	Sender sdk.AccAddress
}

func (b *TransferActionBuilder) GetEventCriteria() *EventCriteria {
	return NewEventCriteria([]ABCIEvent{
		{
			Type:       banktypes.EventTypeTransfer,
			Attributes: map[string][]byte{},
		},
	})
}

func (b *TransferActionBuilder) AddEvent(eventType string, attributes *map[string][]byte) error {
	if eventType == banktypes.EventTypeTransfer {
		// Update the result with the senders address
		address := (*attributes)[banktypes.AttributeKeySender]
		address, err := sdk.AccAddressFromBech32(string(address))
		if err != nil {
			return err
		}
		b.Sender = address
	}
	return nil
}

func (b *TransferActionBuilder) CanBuild() bool {
	return !b.Sender.Empty()
}

func (b *TransferActionBuilder) BuildAction() (EvaluationResult, error) {
	return EvaluationResult{}, nil
}

func (b *TransferActionBuilder) Reset() {
}

// ============ DelegateActionBuilder ============
type DelegateActionBuilder struct {
	Validator sdk.ValAddress
	Delegator sdk.AccAddress
}

func (b *DelegateActionBuilder) GetEventCriteria() *EventCriteria {
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

func (b *DelegateActionBuilder) AddEvent(eventType string, attributes *map[string][]byte) error {
	switch eventType {
	case "delegate":
		address := (*attributes)["validator"]
		validator, err := sdk.ValAddressFromBech32(string(address))
		if err != nil {
			return err
		}
		b.Validator = validator
	case "create_validator":
		address := (*attributes)["validator"]
		validator, err := sdk.ValAddressFromBech32(string(address))
		if err != nil {
			return err
		}
		b.Validator = validator
	case "message":
		// Update the last result to have the delegator's address
		address := (*attributes)["sender"]
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

// ============ VoteActionBuilder ============

type VoteActionBuilder struct {
	Voter sdk.AccAddress
}

func (v *VoteActionBuilder) CanBuild() bool {
	return !v.Voter.Empty()
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
}

func (v *VoteActionBuilder) GetEventCriteria() *EventCriteria {
	return NewEventCriteria([]ABCIEvent{
		{
			Type:       sdk.EventTypeMessage,
			Attributes: map[string][]byte{sdk.AttributeKeyModule: []byte(types.AttributeValueCategory)},
		},
	})
}

func (v *VoteActionBuilder) AddEvent(eventType string, attributes *map[string][]byte) error {
	if eventType == sdk.EventTypeMessage {
		// get the action
		action := (*attributes)[sdk.AttributeKeyAction]
		// accounts for legacy proto message for voting and newer msg type
		if string(action) == sdk.MsgTypeURL(&types.MsgVote{}) || string(action) == sdk.MsgTypeURL(&types.MsgVoteWeighted{}) {
			// Update the result with the voters address
			address := (*attributes)[banktypes.AttributeKeySender]
			address, err := sdk.AccAddressFromBech32(string(address))
			if err != nil {
				return err
			}
			v.Voter = address
		}
	}

	return nil
}
