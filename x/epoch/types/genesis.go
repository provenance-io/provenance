package types

import (
	"errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

func NewGenesisState(epochs []EpochInfo) *GenesisState {
	return &GenesisState{Epochs: epochs}
}

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis(ctx sdk.Context) *GenesisState {
	epochs := []EpochInfo{
		{
			Identifier:              "week",
			StartHeight:             ctx.BlockHeight(),
			Duration: 				 int64((24*60*60*7)/5), //duration in blocks
			CurrentEpoch:            0,
			CurrentEpochStartHeight: ctx.BlockHeight(),
		},
		{
			Identifier:              "day",
			StartHeight:              ctx.BlockHeight(),
			Duration:                int64((24*60*60)/5),
			CurrentEpoch:            0,
			CurrentEpochStartHeight: ctx.BlockHeight(),
		},
	}
	return NewGenesisState(epochs)
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// TODO: Epochs identifiers should be unique
	epochIdentifiers := map[string]bool{}
	for _, epoch := range gs.Epochs {
		if epoch.Identifier == "" {
			return errors.New("epoch identifier should NOT be empty")
		}
		if epochIdentifiers[epoch.Identifier] {
			return errors.New("epoch identifier should be unique")
		}
		if epoch.Duration == 0 {
			return errors.New("epoch duration should NOT be 0")
		}
		epochIdentifiers[epoch.Identifier] = true
	}
	return nil
}
