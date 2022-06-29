package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ActionBuilder interface {
	GetEventCriteria() *EventCriteria
	AddEvent(eventType string, attributes *map[string][]byte) error
	CanBuild() bool
	BuildAction() (EvaluationResult, error)
	Reset()
}

// ============ DelegateTransferActionBuilder ============
type DelegateTransferActionBuilder struct {
}

func (b *DelegateTransferActionBuilder) GetEventCriteria() *EventCriteria {
	return NewEventCriteria([]ABCIEvent{
		{
			Type:       "transfer",
			Attributes: map[string][]byte{},
		},
	})
}

func (b *DelegateTransferActionBuilder) AddEvent(eventType string, attributes *map[string][]byte) error {
	return nil
}

func (b *DelegateTransferActionBuilder) CanBuild() bool {
	return false
}

func (b *DelegateTransferActionBuilder) BuildAction() (EvaluationResult, error) {
	return EvaluationResult{}, nil
}

func (b *DelegateTransferActionBuilder) Reset() {
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
