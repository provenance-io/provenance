package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simappparams "github.com/provenance-io/provenance/app/params"

	keeper "github.com/provenance-io/provenance/x/name/keeper"
	types "github.com/provenance-io/provenance/x/name/types"
)

// Simulation operation weights constants
const (
	OpWeightMsgBindName   = "op_weight_msg_bind_name"
	OpWeightMsgDeleteName = "op_weight_msg_delete_name"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONMarshaler, k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper,
) simulation.WeightedOperations {
	var (
		weightMsgBindName   int
		weightMsgDeleteName int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgBindName, &weightMsgBindName, nil,
		func(_ *rand.Rand) {
			weightMsgBindName = simappparams.DefaultWeightMsgBindName
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgDeleteName, &weightMsgDeleteName, nil,
		func(_ *rand.Rand) {
			weightMsgDeleteName = simappparams.DefaultWeightMsgDeleteName
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgBindName,
			SimulateMsgBindName(k, ak, bk),
		),
		simulation.NewWeightedOperation(
			weightMsgDeleteName,
			SimulateMsgDeleteName(k, ak, bk),
		),
	}
}

func SimulateMsgBindName(k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		var records []types.NameRecord
		k.IterateRecords(ctx, types.NameKeyPrefix, func(record types.NameRecord) error {
			records = append(records, record)
			return nil
		})

		if len(records) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBindNameRequest, "no name records available to create under"), nil, nil
		}

		parent := records[r.Intn(len(records))]
		if parent.Restricted && parent.Address != simAccount.Address.String() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBindNameRequest, "parent name record is restricted, not current owner"), nil, nil
		} else {
			// tell the system we are the owner/signer of the parent record
			parent.Address = simAccount.Address.String()
		}

		restrict := r.Intn(9) < 1
		msg := types.NewMsgBindNameRequest(types.NewNameRecord(simtypes.RandStringOfLength(r, r.Intn(10)+2), simAccount.Address, restrict), parent)

		return Dispatch(r, app, ctx, ak, bk, simAccount, chainID, msg)
	}
}

func SimulateMsgDeleteName(k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		var records []types.NameRecord
		k.IterateRecords(ctx, types.NameKeyPrefix, func(record types.NameRecord) error {
			records = append(records, record)
			return nil
		})

		if len(records) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDeleteNameRequest, "no name records available to delete"), nil, nil
		}

		randomRecord := records[r.Intn(len(records))]

		if simAccount.Address.String() != randomRecord.Address {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDeleteNameRequest, "name record does not belong to user"), nil, nil
		}

		msg := types.NewMsgDeleteNameRequest(randomRecord)

		return Dispatch(r, app, ctx, ak, bk, simAccount, chainID, msg)
	}
}

func Dispatch(
	r *rand.Rand,
	app *baseapp.BaseApp,
	ctx sdk.Context,
	ak authkeeper.AccountKeeperI,
	bk bankkeeper.ViewKeeper,
	from simtypes.Account,
	chainID string,
	msg sdk.Msg,
) (
	simtypes.OperationMsg,
	[]simtypes.FutureOperation,
	error,
) {
	account := ak.GetAccount(ctx, from.Address)
	spendable := bk.SpendableCoins(ctx, account.GetAddress())

	fees, err := simtypes.RandomFees(r, ctx, spendable)
	if err != nil {
		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to generate fees"), nil, err
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
		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to generate mock tx"), nil, err
	}

	_, _, err = app.Deliver(txGen.TxEncoder(), tx)
	if err != nil {
		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), err.Error()), nil, nil
	}

	return simtypes.NewOperationMsg(msg, true, ""), nil, nil
}
