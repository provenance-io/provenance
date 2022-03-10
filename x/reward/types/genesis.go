package types

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Actions []Action

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		ModuleAccountBalance: sdk.NewCoin(DefaultRewardDenom, sdk.ZeroInt()),
		Params: Params{
			AirdropStartTime:   time.Time{},
			DurationUntilDecay: DefaultDurationUntilDecay, // 2 month
			DurationOfDecay:    DefaultDurationOfDecay,    // 4 months
			RewardDenom:        DefaultRewardDenom,        // uosmo
		},
		RewardRecords: []RewardRecord{},
	}
}

// GetGenesisStateFromAppState returns x/reward GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.JSONCodec, appState map[string]json.RawMessage) *GenesisState {
	var genesisState GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return &genesisState
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	totalRewardable := sdk.Coins{}

	for _, rewardRecord := range gs.RewardRecords {
		totalRewardable = totalRewardable.Add(rewardRecord.InitialRewardAmount...)
	}

	if !totalRewardable.IsEqual(sdk.NewCoins(gs.ModuleAccountBalance)) {
		return ErrIncorrectModuleAccountBalance
	}

	if gs.Params.RewardDenom != gs.ModuleAccountBalance.Denom {
		return fmt.Errorf("denom for module and reward does not match")
	}

	return nil
}
