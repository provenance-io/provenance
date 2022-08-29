package types

import (
	"fmt"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

var (
	_ ActionBuilder = &TransferActionBuilder{}
	_ ActionBuilder = &DelegateActionBuilder{}
	_ ActionBuilder = &VoteActionBuilder{}
)

type ActionBuilder interface {
	GetEventCriteria() *EventCriteria
	AddEvent(eventType string, attributes *map[string][]byte) error
	CanBuild() bool
	BuildAction() (EvaluationResult, error)
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
		return EvaluationResult{}, nil
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
			Attributes: map[string][]byte{sdk.AttributeKeyModule: []byte(types.AttributeValueCategory)},
		},
		{
			Type:       sdk.EventTypeMessage,
			Attributes: map[string][]byte{sdk.AttributeKeyAction: nil},
		},
	})
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
		if string(action) != sdk.MsgTypeURL(&types.MsgVote{}) && string(action) != sdk.MsgTypeURL(&types.MsgVoteWeighted{}) {
			return nil
		}

		v.Voted = true
	}

	return nil
}
