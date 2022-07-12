package provwasm

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"

	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
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
)

const (
	denom      = "coinfortestingsmartc" // must be a string of length 20
	namePrefix = "scsnameprefix"        // must be a string of length 13
	label      = "tutorialsc"           // must gbe a string of at least length 10 so that the name module doesn't fail on minlength
)

type Wrapper struct {
	cdc  codec.Codec
	wasm module.AppModuleSimulation
	ak   authkeeper.AccountKeeperI
	bk   bankkeeper.Keeper
	nk   namekeeper.Keeper
}

func NewWrapper(cdc codec.Codec, keeper *wasm.Keeper, validatorSetSource keeper.ValidatorSetSource, ak authkeeper.AccountKeeperI, bk bankkeeper.Keeper, nk namekeeper.Keeper) *Wrapper {
	return &Wrapper{
		cdc:  cdc,
		wasm: wasm.NewAppModule(cdc, keeper, validatorSetSource),
		ak:   ak,
		bk:   bk,
		nk:   nk,
	}
}

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the wasm module.
func (pw Wrapper) GenerateGenesisState(input *module.SimulationState) {
	accessConfig := types.AccessTypeEverybody.With(nil)
	params := types.Params{
		CodeUploadAccess:             accessConfig,
		InstantiateDefaultPermission: accessConfig.Permission,
		MaxWasmCodeSize:              uint64(600 * 1024),
	}

	wasmGenesis := types.GenesisState{
		Params:    params,
		Codes:     nil,
		Contracts: nil,
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
func (pw Wrapper) ProposalContents(simState module.SimulationState) []simtypes.WeightedProposalContent {
	return pw.wasm.ProposalContents(simState)
}

// RandomizedParams returns empty list as the params don't change
func (pw Wrapper) RandomizedParams(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{}
}

// RegisterStoreDecoder registers a decoder for supply module's types
func (pw Wrapper) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
	pw.wasm.RegisterStoreDecoder(sdr)
}

// WeightedOperations returns the all the provwasm operations with their respective weights.
func (pw Wrapper) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	count := 0
	return []simtypes.WeightedOperation{
		simulation.NewWeightedOperation(
			100,
			SimulateMsgBindName(pw.ak, pw.bk, pw.nk, &count),
		),
	}
}

// SimulateMsgBindName will bind a NAME under an existing name
func SimulateMsgBindName(ak authkeeper.AccountKeeperI, bk bankkeeper.Keeper, nk namekeeper.Keeper, count *int) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		if *count != 0 {
			return simtypes.NoOpMsg("provwasm", "", "already bound name"), nil, nil
		}
		*count++

		node := accs[0]
		consumer := accs[1]
		feebucket := accs[2]
		merchant := accs[3]

		var parent nametypes.NameRecord
		err := nk.IterateRecords(ctx, nametypes.NameKeyPrefix, func(record nametypes.NameRecord) error {
			if len(record.Address) > 0 && !strings.Contains(record.Name, ".") {
				parent = record
			}

			return nil
		})

		if err != nil {
			panic(err)
		}

		if len(parent.Name) == 0 {
			panic("no records")
		}

		msg := nametypes.NewMsgBindNameRequest(
			nametypes.NewNameRecord(
				namePrefix,
				node.Address,
				false),
			parent)

		op, future, err2 := namesim.Dispatch(r, app, ctx, ak, bk, node, chainID, msg)

		name := parent.Name

		future = append(future, simtypes.FutureOperation{Op: SimulateMsgAddMarker(ak, bk, nk, node, feebucket, merchant, consumer, name), BlockHeight: int(ctx.BlockHeight()) + 1})

		return op, future, err2
	}
}

// SimulateMsgAddMarker will bind a NAME under an existing name
func SimulateMsgAddMarker(ak authkeeper.AccountKeeperI, bk bankkeeper.Keeper, nk namekeeper.Keeper, node, feebucket, merchant, consumer simtypes.Account, name string) simtypes.Operation {
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

		// fund the node account to do all of these txs
		fundErr := simapp.FundAccount(bk, ctx, node.Address, sdk.NewCoins(sdk.Coin{
			Denom:  "stake",
			Amount: sdk.NewInt(1000000000000000),
		}))

		if fundErr != nil {
			return simtypes.NoOpMsg("provwasm", "", "unable to fund account"), nil, nil
		}

		msg2, ops, err := markersim.Dispatch(r, app, ctx, ak, bk, node, chainID, msg, nil)

		ops = append(ops, simtypes.FutureOperation{Op: SimulateMsgAddAccess(ak, bk, nk, node, feebucket, merchant, consumer, name), BlockHeight: int(ctx.BlockHeight()) + 1})

		return msg2, ops, err
	}
}

func SimulateMsgAddAccess(ak authkeeper.AccountKeeperI, bk bankkeeper.Keeper, nk namekeeper.Keeper, node, feebucket, merchant, consumer simtypes.Account, name string) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		accessTypes := []markertypes.Access{markertypes.AccessByName("withdraw")}
		grant := *markertypes.NewAccessGrant(node.Address, accessTypes)
		msg := markertypes.NewMsgAddAccessRequest(denom, node.Address, grant)
		msg2, ops, err := markersim.Dispatch(r, app, ctx, ak, bk, node, chainID, msg, nil)

		ops = append(ops, simtypes.FutureOperation{Op: SimulateFinalizeMarker(ak, bk, nk, node, feebucket, merchant, consumer, name), BlockHeight: int(ctx.BlockHeight()) + 1})

		return msg2, ops, err
	}
}

func SimulateFinalizeMarker(ak authkeeper.AccountKeeperI, bk bankkeeper.Keeper, nk namekeeper.Keeper, node, feebucket, merchant, consumer simtypes.Account, name string) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msg := markertypes.NewMsgFinalizeRequest(denom, node.Address)

		msg2, ops, err := markersim.Dispatch(r, app, ctx, ak, bk, node, chainID, msg, nil)

		ops = append(ops, simtypes.FutureOperation{Op: SimulateActivateMarker(ak, bk, nk, node, feebucket, merchant, consumer, name), BlockHeight: int(ctx.BlockHeight()) + 1})

		return msg2, ops, err
	}
}

func SimulateActivateMarker(ak authkeeper.AccountKeeperI, bk bankkeeper.Keeper, nk namekeeper.Keeper, node, feebucket, merchant, consumer simtypes.Account, name string) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msg := markertypes.NewMsgActivateRequest(denom, node.Address)

		msg2, ops, err := markersim.Dispatch(r, app, ctx, ak, bk, node, chainID, msg, nil)

		ops = append(ops, simtypes.FutureOperation{Op: SimulateMsgWithdrawRequest(ak, bk, nk, node, feebucket, merchant, consumer, name), BlockHeight: int(ctx.BlockHeight()) + 1})

		return msg2, ops, err
	}
}

func SimulateMsgWithdrawRequest(ak authkeeper.AccountKeeperI, bk bankkeeper.Keeper, nk namekeeper.Keeper, node, feebucket, merchant, consumer simtypes.Account, name string) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		coins := []sdk.Coin{{
			Denom:  denom,
			Amount: sdk.NewIntFromUint64(1000000),
		}}
		msg := markertypes.NewMsgWithdrawRequest(node.Address, consumer.Address, denom, coins)
		msg2, ops, err := markersim.Dispatch(r, app, ctx, ak, bk, node, chainID, msg, nil)

		ops = append(ops, simtypes.FutureOperation{Op: SimulateMsgStoreContract(ak, bk, nk, node, feebucket, merchant, consumer, name), BlockHeight: int(ctx.BlockHeight()) + 1})

		return msg2, ops, err
	}
}

func SimulateMsgStoreContract(ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, nk namekeeper.Keeper, node, feebucket, merchant, consumer simtypes.Account, name string) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		code, err := ioutil.ReadFile("./sim_contracts/tutorial.wasm")

		if err != nil {
			panic(err)
		}

		msg := &types.MsgStoreCode{
			Sender:       feebucket.Address.String(),
			WASMByteCode: code,
		}

		msg2, ops, _, storeErr := Dispatch(r, app, ctx, ak, bk, feebucket, chainID, msg, nil)

		ops = append(ops, simtypes.FutureOperation{Op: SimulateMsgInstantiateContract(ak, bk, nk, node, feebucket, merchant, consumer, name), BlockHeight: int(ctx.BlockHeight()) + 1})

		return msg2, ops, storeErr
	}
}

func SimulateMsgInstantiateContract(ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, nk namekeeper.Keeper, node, feebucket, merchant, consumer simtypes.Account, name string) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		record, err := nk.GetRecordByName(ctx, name)

		if err != nil || record.Address == "" {
			return simtypes.NoOpMsg("provwasm", "", "name record has been removed"), nil, nil
		}

		m := fmt.Sprintf(`{ "contract_name": "%s.%s", "purchase_denom": "%s", "merchant_address": "%s", "fee_percent": "0.10" }`, label, name, denom, merchant.Address.String())
		amountStr := fmt.Sprintf("0%s", denom)
		amount, err := sdk.ParseCoinsNormalized(amountStr)

		if err != nil {
			panic(err)
		}

		msg := &types.MsgInstantiateContract{
			Sender: feebucket.Address.String(),
			Admin:  feebucket.Address.String(),
			CodeID: 1,
			Label:  label,
			Msg:    []byte(m),
			Funds:  amount,
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

		var contractAddr string
		if err4 == nil {
			contractAddr = pInstResp.Address
		}

		ops = append(ops, simtypes.FutureOperation{Op: SimulateMsgExecuteContract(ak, bk, node, consumer, contractAddr), BlockHeight: int(ctx.BlockHeight()) + 1})

		return msg2, ops, instantiateErr
	}
}

func SimulateMsgExecuteContract(ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, node, consumer simtypes.Account, contractAddr string) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		amount, _ := sdk.ParseCoinsNormalized(fmt.Sprintf("100%s", denom))
		coins := bk.SpendableCoins(ctx, consumer.Address)

		if coins.AmountOf(denom).LT(sdk.NewInt(100)) {
			return simtypes.NoOpMsg("provwasm", "", "not enough coins"), nil, nil
		}

		msg := &types.MsgExecuteContract{
			Sender:   consumer.Address.String(),
			Funds:    amount,
			Contract: contractAddr,
			Msg:      []byte("{\"purchase\":{\"id\":\"12345\"}}"),
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
