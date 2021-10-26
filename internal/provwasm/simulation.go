package provwasm

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmd/x/wasm/types"

	"io/ioutil"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

type ProvwasmWrapper struct {
	cdc codec.Codec
	wasm module.AppModuleSimulation
}

func NewProvwasmWrapper(cdc codec.Codec, keeper *wasm.Keeper, validatorSetSource keeper.ValidatorSetSource) *ProvwasmWrapper {

	return &ProvwasmWrapper{
		cdc: cdc,
		wasm: wasm.NewAppModule(cdc, keeper, validatorSetSource),
	}
}

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the wasm module.
func (pw ProvwasmWrapper) GenerateGenesisState(input *module.SimulationState) {
	codeBytes, err := ioutil.ReadFile("/Users/fredkneeland/code/provenance/tutorial.wasm")
	if err != nil {
		panic("failed to read file")
	}

	codes := make([]types.Code, 1)
	codes[0] = types.Code{
		CodeID: 1,
		CodeInfo: types.CodeInfoFixture(types.WithSHA256CodeHash(codeBytes)),
		CodeBytes: codeBytes,
	}

	contracts := make([]types.Contract, 1)
	contracts[0] = types.Contract{
		ContractAddress: input.Accounts[0].Address.String(),
		ContractInfo:    types.ContractInfoFixture(func(c *types.ContractInfo) { c.CodeID = 1 }, types.OnlyGenesisFields),
	}

	wasmGenesis := types.GenesisState{
		Params:    types.DefaultParams(),
		Codes:     codes,
		Contracts: contracts,
		Sequences: []types.Sequence{
			{IDKey: types.KeyLastCodeID, Value: 2},
			{IDKey: types.KeyLastInstanceID, Value: 2},
		},
		GenMsgs:   nil,
	}

	_, err = input.Cdc.MarshalJSON(&wasmGenesis)
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
func (pw ProvwasmWrapper) RandomizedParams(r *rand.Rand) []simtypes.ParamChange {
	return pw.wasm.RandomizedParams(r)
}

// RegisterStoreDecoder registers a decoder for supply module's types
func (pw ProvwasmWrapper) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
}

// WeightedOperations returns the all the provwasm operations with their respective weights.
func (pw ProvwasmWrapper) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {

	// Okay... so I probably want to return the basic opperations here??? idk...


	// I will need the separate accounts all created and with necessary gas fees as well as the necessary currency created?  I may want to use a different smart contract... lol

	return nil
}
