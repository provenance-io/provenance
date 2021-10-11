package provwasm

import (
	"fmt"
	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmsim "github.com/CosmWasm/wasmd/x/wasm/simulation"
	"github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"

	//"fmt"
	//"github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	//"github.com/cosmos/cosmos-sdk/x/simulation"
	"math/rand"
)

type ProvwasmWrapper struct {
	cdc codec.Codec
	wasm module.AppModule
}

func NewProvwasmWrapper(cdc codec.Codec, keeper *wasm.Keeper, validatorSetSource keeper.ValidatorSetSource) *ProvwasmWrapper {
	return &ProvwasmWrapper{
		cdc: cdc,
		wasm: wasm.NewAppModule(cdc, keeper, validatorSetSource),
	}
}

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the wasm module.
func (ProvwasmWrapper) GenerateGenesisState(input *module.SimulationState) {
	fmt.Println("asdf - GenerateGenesisState for provwasm")
	params := wasmsim.RandomParams(input.Rand)
	wasmGenesis := types.GenesisState{
		Params:    params,
		Codes:     nil,
		Contracts: nil, // TODO: add contract specific code here
		Sequences: nil,
		GenMsgs:   nil,
	}

	_, err := input.Cdc.MarshalJSON(&wasmGenesis)
	if err != nil {
		panic(err)
	}

	input.GenState[types.ModuleName] = input.Cdc.MustMarshalJSON(&wasmGenesis)
}

// ProposalContents doesn't return any content functions for governance proposals.
func (ProvwasmWrapper) ProposalContents(simState module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

// RandomizedParams creates randomized bank param changes for the simulator.
func (am ProvwasmWrapper) RandomizedParams(r *rand.Rand) []simtypes.ParamChange {
	params := wasmsim.RandomParams(r)
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, string(types.ParamStoreKeyUploadAccess),
			func(r *rand.Rand) string {
				jsonBz, err := am.cdc.MarshalJSON(&params.CodeUploadAccess)
				if err != nil {
					panic(err)
				}
				return string(jsonBz)
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.ParamStoreKeyInstantiateAccess),
			func(r *rand.Rand) string {
				return fmt.Sprintf("%q", params.CodeUploadAccess.Permission.String())
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.ParamStoreKeyMaxWasmCodeSize),
			func(r *rand.Rand) string {
				return fmt.Sprintf(`"%d"`, params.MaxWasmCodeSize)
			},
		),
	}
}

// RegisterStoreDecoder registers a decoder for supply module's types
func (am ProvwasmWrapper) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am ProvwasmWrapper) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return nil
}
