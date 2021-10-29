package provwasm

import (
	"fmt"
	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simappparams "github.com/provenance-io/provenance/app/params"
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
	count := 0
	return []simtypes.WeightedOperation{
		simulation.NewWeightedOperation(
			weight,
			SimulateMsgBindName(pw.ak, pw.bk, pw.nk, &count),
		),
		simulation.NewWeightedOperation(
			weight-1,
			SimulateMsgAddMarker(pw.ak, pw.bk, &count),
		),
		simulation.NewWeightedOperation(
			weight-2,
			SimulateFinalizeOrActivateMarker(pw.ak, pw.bk, &count),
		),
		simulation.NewWeightedOperation(
			weight-3,
			SimulateMsgAddAccess(pw.ak, pw.bk, &count),
		),
		simulation.NewWeightedOperation(
			weight-4,
			SimulateMsgWithdrawRequest(pw.ak, pw.bk, &count),
		),
		//simulation.NewWeightedOperation(
		//	weight-4,
		//	SimulateMsgStoreContract(pw.ak, pw.bk, &count),
		//),
		simulation.NewWeightedOperation(
			weight-5,
			SimulateMsgInitiateContract(pw.ak, pw.bk, &count),
		),

	}
}

// SimulateMsgBindName will bind a NAME under an existing name using a 40% probability of restricting it.
func SimulateMsgBindName(ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, nk namekeeper.Keeper, count *int) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		if *count != 0 {
			return simtypes.NoOpMsg("provwasm", "", "already bound name"), nil, nil
		}
		*count = *count + 1
		fmt.Println(count)
		node := accs[0]

		var parent nametypes.NameRecord
		nk.IterateRecords(ctx, nametypes.NameKeyPrefix, func(record nametypes.NameRecord) error {
			parent = record
			return nil
		})

		if len(parent.Name) == 0 {
			panic("no records")
		}

		msg := nametypes.NewMsgBindNameRequest(
			nametypes.NewNameRecord(
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
func SimulateMsgAddMarker(ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, count *int) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		if *count != 1 {
			return simtypes.NoOpMsg("provwasm", "", "already added marker"), nil, nil
		}
		*count = *count + 1
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

func SimulateFinalizeOrActivateMarker(ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, count *int) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		if *count != 3 && *count != 4 {
			return simtypes.NoOpMsg("provwasm", "", "already activated marker"), nil, nil
		}
		*count = *count + 1

		node := accs[0]
		var msg sdk.Msg
		if *count == 4 {
			msg = markertypes.NewMsgFinalizeRequest("purchasecoineightsss", node.Address)
		} else {
			msg = markertypes.NewMsgActivateRequest("purchasecoineightsss", node.Address)
		}

		return markersim.Dispatch(r, app, ctx, ak, bk, node, chainID, msg, nil)
	}
}

func SimulateMsgAddAccess(ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, count *int) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		if *count != 2 {
			return simtypes.NoOpMsg("provwasm", "", "already activated marker"), nil, nil
		}
		*count = *count + 1

		node := accs[0]

		accessTypes := []markertypes.Access{markertypes.AccessByName("withdraw")}
		grant := *markertypes.NewAccessGrant(node.Address, accessTypes)
		msg := markertypes.NewMsgAddAccessRequest("purchasecoineightsss", node.Address, grant)
		return markersim.Dispatch(r, app, ctx, ak, bk, node, chainID, msg, nil)
	}
}

func SimulateMsgWithdrawRequest(ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, count *int) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		if *count != 5 {
			return simtypes.NoOpMsg("provwasm", "", "already activated marker"), nil, nil
		}
		*count = *count + 1

		node := accs[0]
		customer := accs[1]

		coins := []sdk.Coin{{
			"purchasecoineightsss",
			sdk.NewIntFromUint64(1000000),
		}}
		msg := markertypes.NewMsgWithdrawRequest(node.Address, customer.Address, "purchasecoineightsss", coins)
		return markersim.Dispatch(r, app, ctx, ak, bk, node, chainID, msg, nil)
	}
}

// We shouldn't need to store the contract as it is in the Genesis??? Correct???
func SimulateMsgStoreContract(ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, count *int) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		if *count != 6 {
			return simtypes.NoOpMsg("provwasm", "", "already activated marker"), nil, nil
		}
		*count = *count + 1

		feebucket := accs[2]

		code, err := ioutil.ReadFile("/Users/fredkneeland/code/provenance/tutorial.wasm")

		if err != nil {
			panic(err)
		}

		msg := &types.MsgStoreCode{
			Sender: feebucket.Address.String(),
			WASMByteCode: code,
		}

		return Dispatch(r, app, ctx, ak, bk, feebucket, chainID, msg, nil)
	}
}

func SimulateMsgInitiateContract(ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, count *int) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		fmt.Println("---------------")
		fmt.Println("Initiate Contract")
		if *count != 6 {
			return simtypes.NoOpMsg("provwasm", "", "already activated marker"), nil, nil
		}
		*count = *count + 1

		fmt.Println("starting stuffs")

		feebucket := accs[0]
		merchant := accs[1]

		m := fmt.Sprintf(`{ "contract_name": "tutorial.sctwoandthree.oamtciwub", "purchase_denom": "purchasecoineightsss", "merchant_address": "%s", "fee_percent": "0.10" }`, merchant.Address.String())

		msg := &types.MsgInstantiateContract{
			Sender: feebucket.Address.String(),
			Admin: feebucket.Address.String(),
			CodeID: 1,
			Label: "tutorial",
			Msg: []byte(m),
		}

		fmt.Println("ready to send Dispatch")

		defer fmt.Println("AFter func")
		return Dispatch(r, app, ctx, ak, bk, feebucket, chainID, msg, nil)
	}
}

// Ideally this would someday live in wasmd

// Dispatch sends an operation to the chain using a given account/funds on account for fees.  Failures on the server side
// are handled as no-op msg operations with the error string as the status/response.
func Dispatch(
	r *rand.Rand,
	app *baseapp.BaseApp,
	ctx sdk.Context,
	ak authkeeper.AccountKeeperI,
	bk bankkeeper.ViewKeeper,
	from simtypes.Account,
	chainID string,
	msg sdk.Msg,
	futures []simtypes.FutureOperation,
) (
	simtypes.OperationMsg,
	[]simtypes.FutureOperation,
	error,
) {
	account := ak.GetAccount(ctx, from.Address)
	spendable := bk.SpendableCoins(ctx, account.GetAddress())

	//fees := make([]sdk.Coin, 0)
	//
	//for _, c := range spendable {
	//	if !c.Amount.IsZero() {
	//		fees = append(fees, c)
	//	}
	//}
	//
	//if len(fees) == 0 {
	//	panic("no fees")
	//}

	fees, err := simtypes.RandomFees(r, ctx, spendable)
	if err != nil {
		panic("no fees")
		return simtypes.NoOpMsg(types.ModuleName, fmt.Sprintf("%T", msg), "unable to generate fees"), nil, err
	}

	txGen := simappparams.MakeTestEncodingConfig().TxConfig
	tx, err := helpers.GenTx(
		txGen,
		[]sdk.Msg{msg},
		fees,
		helpers.DefaultGenTxGas,
		chainID,
		[]uint64{account.GetAccountNumber()},
		[]uint64{account.GetSequence()},
		from.PrivKey,
	)
	if err != nil {
		panic(err)
		return simtypes.NoOpMsg(types.ModuleName, fmt.Sprintf("%T", msg), "unable to generate mock tx"), nil, err
	}

	_, _, err = app.Deliver(txGen.TxEncoder(), tx)
	if err != nil {
		panic(err)
		return simtypes.NoOpMsg(types.ModuleName, fmt.Sprintf("%T", msg), err.Error()), nil, nil
	}

	return simtypes.NewOperationMsg(msg, true, "", &codec.ProtoCodec{}), futures, nil
}
