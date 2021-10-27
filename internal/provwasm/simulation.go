package provwasm

import (
	"fmt"
	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	namekeeper "github.com/provenance-io/provenance/x/name/keeper"

	"io/ioutil"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	markersim "github.com/provenance-io/provenance/x/marker/simulation"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	namesim "github.com/provenance-io/provenance/x/name/simulation"
	nametypes "github.com/provenance-io/provenance/x/name/types"
)

type ProvwasmWrapper struct {
	cdc codec.Codec
	wasm module.AppModuleSimulation
	ak authkeeper.AccountKeeperI
	bk bankkeeper.ViewKeeper
	nk namekeeper.Keeper
}

func NewProvwasmWrapper(cdc codec.Codec, keeper *wasm.Keeper, validatorSetSource keeper.ValidatorSetSource, ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, nk namekeeper.Keeper) *ProvwasmWrapper {

	return &ProvwasmWrapper{
		cdc: cdc,
		wasm: wasm.NewAppModule(cdc, keeper, validatorSetSource),
		ak: ak,
		bk: bk,
		nk: nk,
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
	weight := 10000
	namecount := 0
	acount := 0
	addcount := 0
	return []simtypes.WeightedOperation{
		simulation.NewWeightedOperation(
			weight,
			SimulateMsgBindName(pw.ak, pw.bk, pw.nk, &namecount),
		),
		simulation.NewWeightedOperation(
			weight-1,
			SimulateMsgAddMarker(pw.ak, pw.bk, &addcount),
		),
		simulation.NewWeightedOperation(
			weight-2,
			SimulateFinalizeOrActivateMarker(pw.ak, pw.bk, &acount),
		),
	}
}

// SimulateMsgBindName will bind a NAME under an existing name using a 40% probability of restricting it.
func SimulateMsgBindName(ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, nk namekeeper.Keeper, namecount *int) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		if *namecount > 0 {
			return simtypes.NoOpMsg("provwasm", "", "already bound name"), nil, nil
		}
		*namecount = *namecount + 1
		fmt.Println(namecount)
		node := accs[0]

		var parent nametypes.NameRecord
		nk.IterateRecords(ctx, nametypes.NameKeyPrefix, func(record nametypes.NameRecord) error {
			parent = record
			return nil
		})

		if len(parent.Name) == 0 {
			panic("no records")
		}

		fmt.Println(parent.Name)

		name := simtypes.RandStringOfLength(r, r.Intn(10)+2)
		fmt.Println("name:")
		fmt.Println(name)

		msg := nametypes.NewMsgBindNameRequest(
			nametypes.NewNameRecord(
				//name,
				"sctwoandthree",
				node.Address,
				true),
			nametypes.NewNameRecord(
				parent.Name,
				//"pb",
				node.Address,
				false))

		return namesim.Dispatch(r, app, ctx, ak, bk, node, chainID, msg)
	}
}

// SimulateMsgAddMarker will bind a NAME under an existing name using a 40% probability of restricting it.
func SimulateMsgAddMarker(ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, addcount *int) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		if *addcount > 0 {
			return simtypes.NoOpMsg("provwasm", "", "already added marker"), nil, nil
		}
		*addcount = *addcount + 1
		fmt.Println("-----------------")
		fmt.Println("Simulate add Marker")
		fmt.Println("-----------------")
		node := accs[0]
		// [a-zA-Z][a-zA-Z0-9\\-\\.]{18,21})
		denom := "purchasecoineightsss"
		msg := markertypes.NewMsgAddMarkerRequest(
			denom,
			sdk.NewIntFromUint64(1000000000),
			node.Address,
			node.Address,
			markertypes.MarkerType_Coin,
			true, // fixed supply
			true, // allow gov
		)

		return markersim.Dispatch(r, app, ctx, ak, bk, node, chainID, msg, nil)
	}
}

func SimulateFinalizeOrActivateMarker(ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, markercount *int) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		if *markercount > 1 {
			return simtypes.NoOpMsg("provwasm", "", "already activated marker"), nil, nil
		}
		*markercount = *markercount + 1

		node := accs[0]
		var msg sdk.Msg
		if *markercount == 1 {
			msg = markertypes.NewMsgFinalizeRequest("purchasecoineightsss", node.Address)
		} else {
			msg = markertypes.NewMsgActivateRequest("purchasecoineightsss", node.Address)
		}

		simAccount, found := simtypes.FindAccount(accs, node.Address)
		if !found {
			panic("couldn't find manager of sim account")
		}

		return markersim.Dispatch(r, app, ctx, ak, bk, simAccount, chainID, msg, nil)
	}
}
