package types

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
)

// NewGenesisState creates new GenesisState object
func NewGenesisState(params Params, entries []*MsgFee) *GenesisState {
	return &GenesisState{
		Params:  params,
		MsgFees: entries,
	}
}

// Validate returns an error if there's something wrong with this genesis state.
func (s *GenesisState) Validate() error {
	if s == nil {
		return errors.New("flatfees genesis state cannot be nil")
	}

	var errs []error
	if err := s.Params.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("invalid flatfees params: %w", err))
	}

	seen := make(map[string]int)
	for i, msgFee := range s.MsgFees {
		if msgFee != nil {
			seen[msgFee.MsgTypeUrl]++
		}
		switch {
		case msgFee == nil || seen[msgFee.MsgTypeUrl] == 1:
			if err := msgFee.Validate(); err != nil {
				errs = append(errs, fmt.Errorf("invalid MsgFees[%d]: %w", i, err))
			}
		case seen[msgFee.MsgTypeUrl] == 2:
			errs = append(errs, fmt.Errorf("duplicate MsgTypeUrl not allowed, %q", msgFee.MsgTypeUrl))
		}
		// If it's at 3 or more, we've already added an error for the msg type url, no need to add more.
	}

	return errors.Join(errs...)
}

// DefaultGenesisState returns default state for x/flatfees module.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:  DefaultParams(),
		MsgFees: []*MsgFee{},
	}
}

// GetGenesisStateFromAppState returns x/flatfees GenesisState given raw application genesis state.
// If the appState doesn't have anything for this module, a zero-value (empty) GenesisState is returned.
func GetGenesisStateFromAppState(cdc codec.JSONCodec, appState map[string]json.RawMessage) (*GenesisState, error) {
	var genesisState GenesisState
	var err error

	if len(appState[ModuleName]) > 0 {
		err = cdc.UnmarshalJSON(appState[ModuleName], &genesisState)
		if err != nil {
			err = fmt.Errorf("could not unmarshal flatfees genesis state: %w", err)
		}
	}

	return &genesisState, err
}
