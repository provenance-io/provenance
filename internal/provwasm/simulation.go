package provwasm

import (
	"fmt"
	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/golang/protobuf/proto"
	simappparams "github.com/provenance-io/provenance/app/params"
	markersim "github.com/provenance-io/provenance/x/marker/simulation"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	namekeeper "github.com/provenance-io/provenance/x/name/keeper"
	namesim "github.com/provenance-io/provenance/x/name/simulation"
	nametypes "github.com/provenance-io/provenance/x/name/types"
	"io/ioutil"
	"math/rand"
)

const (
	denom = "coinfortestingsmartc" // must be a string of length 20
	namePrefix = "scsnameprefix" // must be a string of length 13
	label = "tutorials"
)

type ProvwasmWrapper struct {
	cdc codec.Codec
	wasm module.AppModuleSimulation
	ak authkeeper.AccountKeeperI
	bk bankkeeper.ViewKeeper
	nk namekeeper.Keeper
	keeper wasm.Keeper
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
	codeBytes, err := ioutil.ReadFile("./sim_contracts/tutorial.wasm")
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
	count := 0
	return []simtypes.WeightedOperation{
		simulation.NewWeightedOperation(
			100,
			SimulateMsgBindName(pw.ak, pw.bk, pw.nk, pw.keeper, &count),
		),
	}
}

// SimulateMsgBindName will bind a NAME under an existing name using a 40% probability of restricting it.
func SimulateMsgBindName(ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, nk namekeeper.Keeper, keeper wasm.Keeper, count *int) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		if *count != 0 {
			return simtypes.NoOpMsg("provwasm", "", "already bound name"), nil, nil
		}
		*count = *count + 1


		node := accs[0]
		consumer := accs[1]
		feebucket := accs[2]
		merchant := accs[3]

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
				namePrefix,
				node.Address,
				false),
			parent)
			//nametypes.NewNameRecord(
			//	parent.Name,
			//	node.Address,
			//	false))

		op, future, err := namesim.Dispatch(r, app, ctx, ak, bk, node, chainID, msg)

		future = append(future, simtypes.FutureOperation{Op: SimulateMsgAddMarker(ak, bk, node), BlockHeight: 2})
		future = append(future, simtypes.FutureOperation{Op: SimulateMsgAddAccess(ak, bk, node), BlockHeight: 3})
		future = append(future, simtypes.FutureOperation{Op: SimulateFinalizeOrActivateMarker(ak, bk, true, node), BlockHeight: 4})
		future = append(future, simtypes.FutureOperation{Op: SimulateFinalizeOrActivateMarker(ak, bk, false, node), BlockHeight: 5})
		future = append(future, simtypes.FutureOperation{Op: SimulateMsgWithdrawRequest(ak, bk, node, consumer), BlockHeight: 6})
		future = append(future, simtypes.FutureOperation{Op: SimulateMsgStoreContract(ak, bk, feebucket), BlockHeight: 6})

		var contractAddr string
		future = append(future, simtypes.FutureOperation{Op: SimulateMsgInitiateContract(ak, bk, feebucket, merchant, parent.Name, &contractAddr), BlockHeight: 7})
		future = append(future, simtypes.FutureOperation{Op: SimulateMsgExecuteContract(ak, bk, consumer, &contractAddr), BlockHeight: 8})

		return op, future, err
	}
}

// SimulateMsgAddMarker will bind a NAME under an existing name using a 40% probability of restricting it.
func SimulateMsgAddMarker(ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, node simtypes.Account) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
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

func SimulateFinalizeOrActivateMarker(ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, finalize bool, node simtypes.Account) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		var msg sdk.Msg
		if finalize {
			msg = markertypes.NewMsgFinalizeRequest(denom, node.Address)
		} else {
			msg = markertypes.NewMsgActivateRequest(denom, node.Address)
		}

		return markersim.Dispatch(r, app, ctx, ak, bk, node, chainID, msg, nil)
	}
}

func SimulateMsgAddAccess(ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, node simtypes.Account) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		accessTypes := []markertypes.Access{markertypes.AccessByName("withdraw")}
		grant := *markertypes.NewAccessGrant(node.Address, accessTypes)
		msg := markertypes.NewMsgAddAccessRequest(denom, node.Address, grant)
		return markersim.Dispatch(r, app, ctx, ak, bk, node, chainID, msg, nil)
	}
}

func SimulateMsgWithdrawRequest(ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, node simtypes.Account, customer simtypes.Account) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		coins := []sdk.Coin{{
			denom,
			sdk.NewIntFromUint64(1000000),
		}}
		msg := markertypes.NewMsgWithdrawRequest(node.Address, customer.Address, denom, coins)
		return markersim.Dispatch(r, app, ctx, ak, bk, node, chainID, msg, nil)
	}
}

func SimulateMsgStoreContract(ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, feebucket simtypes.Account) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		code, err := ioutil.ReadFile("./sim_contracts/tutorial.wasm")

		if err != nil {
			panic(err)
		}

		msg := &types.MsgStoreCode{
			Sender: feebucket.Address.String(),
			WASMByteCode: code,
		}

		msg2, ops, _, instantiateErr := Dispatch(r, app, ctx, ak, bk, feebucket, chainID, msg, nil)

		return msg2, ops, instantiateErr
	}
}

func SimulateMsgInitiateContract(ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, feebucket, merchant simtypes.Account, name string, contractAddr *string) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		//m := fmt.Sprintf(`{ "contract_name": "%s.%s.%s", "purchase_denom": "%s", "merchant_address": "%s", "fee_percent": "0.10" }`, label, namePrefix, name, denom, merchant.Address.String())
		m := fmt.Sprintf(`{ "contract_name": "%s.%s", "purchase_denom": "%s", "merchant_address": "%s", "fee_percent": "0.10" }`, label, name, denom, merchant.Address.String())

		// hmm, we currently have access to 0 funds...
		amountStr := fmt.Sprintf("0%s", denom)
		amount, err := sdk.ParseCoinsNormalized(amountStr)

		if err != nil {
			fmt.Println("Err on Parse Coins Normalized")
			panic(err)
		}

		msg := &types.MsgInstantiateContract{
			Sender: feebucket.Address.String(),
			Admin: feebucket.Address.String(),
			CodeID: 1,
			Label: label,
			Msg: []byte(m),
			Funds: amount, // do we need funds?
		}

		msg2, ops, sdkResponse, instantiateErr := Dispatch(r, app, ctx, ak, bk, feebucket, chainID, msg, nil)

		// get the contract address for use when executing the contract
		var protoResult sdk.TxMsgData
		err3 := proto.Unmarshal(sdkResponse.Data, &protoResult)

		if err3 != nil {
			panic(err3)
		}

		var pInstResp types.MsgInstantiateContractResponse
		err4 := pInstResp.Unmarshal(protoResult.Data[0].Data)

		if err4 == nil {
			*contractAddr = pInstResp.Address
		}


		return msg2, ops, instantiateErr
	}
}

func SimulateMsgExecuteContract(ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, consumer simtypes.Account, contractAddr *string) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		amount, err := sdk.ParseCoinsNormalized(fmt.Sprintf("100%s", denom))

		if err != nil {
			panic(err)
		}

		msg := &types.MsgExecuteContract{
			Sender: consumer.Address.String(),
			Funds: amount,
			Contract: *contractAddr,
			Msg: []byte("{\"purchase\":{\"id\":\"12345\"}}"),
		}

		msg2, ops, _, err2 := Dispatch(r, app, ctx, ak, bk, consumer, chainID, msg, nil)
		return msg2, ops, err2
	}
}


// Dispatch sends an operation to the chain using a given account/funds on account for fees.  Failures on the server side
// are handled as no-op msg operations with the error string as the status/response.
// Ideally this would live in wasmd
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
	*sdk.Result,
	error,
) {
	account := ak.GetAccount(ctx, from.Address)
	spendable := bk.SpendableCoins(ctx, account.GetAddress())

	fees, err := simtypes.RandomFees(r, ctx, spendable)

	if err != nil {
		panic("no fees")
	}

	txGen := simappparams.MakeTestEncodingConfig().TxConfig
	tx, err := helpers.GenTx(
		txGen,
		[]sdk.Msg{msg},
		fees,
		helpers.DefaultGenTxGas*10, // storing a contract requires more gas than most txs
		chainID,
		[]uint64{account.GetAccountNumber()},
		[]uint64{account.GetSequence()},
		from.PrivKey,
	)
	if err != nil {
		panic(err)
	}

	_, sdkResponse, err2 := app.Deliver(txGen.TxEncoder(), tx)
	if err2 != nil {
		panic(err2)
	}

	return simtypes.NewOperationMsg(msg, true, "", &codec.ProtoCodec{}), futures, sdkResponse, nil
}
