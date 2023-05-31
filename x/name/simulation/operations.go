package simulation

import (
	"math/rand"
	"strings"

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
	//nolint:gosec // not credentials
	OpWeightMsgBindName = "op_weight_msg_bind_name"
	//nolint:gosec // not credentials
	OpWeightMsgDeleteName = "op_weight_msg_delete_name"
	//nolint:gosec // not credentials
	OpWeightMsgModifyName = "op_weight_msg_modify_name"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper,
) simulation.WeightedOperations {
	var (
		weightMsgBindName   int
		weightMsgDeleteName int
		weightMsgModifyName int
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

	appParams.GetOrGenerate(cdc, OpWeightMsgModifyName, &weightMsgModifyName, nil,
		func(_ *rand.Rand) {
			weightMsgModifyName = simappparams.DefaultWeightMsgModifyName
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
		simulation.NewWeightedOperation(
			weightMsgModifyName,
			SimulateMsgModifyName(k, ak, bk),
		),
	}
}

// SimulateMsgBindName will bind a NAME under an existing name using a 40% probability of restricting it.
func SimulateMsgBindName(k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		parentRecord, parentOwner, found, err := getRandomRecord(r, ctx, k, accs, true)
		if err != nil {
			return simtypes.NoOpMsg(sdk.MsgTypeURL(&types.MsgBindNameRequest{}), sdk.MsgTypeURL(&types.MsgBindNameRequest{}), "iterator of existing records failed"), nil, err
		}
		if !found {
			return simtypes.NoOpMsg(sdk.MsgTypeURL(&types.MsgBindNameRequest{}), sdk.MsgTypeURL(&types.MsgBindNameRequest{}), "no name records available to create under"), nil, nil
		}

		newRecordName := simtypes.RandStringOfLength(r, r.Intn(10)+2)
		newRecordOwner := parentOwner
		if !parentRecord.Restricted {
			newRecordOwner, _ = simtypes.RandomAcc(r, accs)
		}
		newRecordRestricted := r.Intn(9) < 4
		newRecord := types.NewNameRecord(newRecordName, newRecordOwner.Address, newRecordRestricted)
		msg := types.NewMsgBindNameRequest(newRecord, parentRecord)

		return Dispatch(r, app, ctx, ak, bk, parentOwner, chainID, msg)
	}
}

// SimulateMsgDeleteName will dispatch a delete name operation against a random name record
func SimulateMsgDeleteName(k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		randomRecord, simAccount, found, err := getRandomRecord(r, ctx, k, accs, false)
		if err != nil {
			return simtypes.NoOpMsg(sdk.MsgTypeURL(&types.MsgDeleteNameRequest{}), sdk.MsgTypeURL(&types.MsgDeleteNameRequest{}), "iterator of existing records failed"), nil, err
		}
		if !found {
			return simtypes.NoOpMsg(sdk.MsgTypeURL(&types.MsgDeleteNameRequest{}), sdk.MsgTypeURL(&types.MsgDeleteNameRequest{}), "no name records available to delete"), nil, nil
		}

		msg := types.NewMsgDeleteNameRequest(randomRecord)

		return Dispatch(r, app, ctx, ak, bk, simAccount, chainID, msg)
	}
}

// SimulateMsgModifyName will dispatch a modify name operation against a random name record
func SimulateMsgModifyName(k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		randomRecord, simAccount, found, err := getRandomRecord(r, ctx, k, accs, true)
		if err != nil {
			return simtypes.NoOpMsg(sdk.MsgTypeURL(&types.MsgModifyNameRequest{}), sdk.MsgTypeURL(&types.MsgModifyNameRequest{}), "iterator of existing records failed"), nil, err
		}
		if !found {
			return simtypes.NoOpMsg(sdk.MsgTypeURL(&types.MsgModifyNameRequest{}), sdk.MsgTypeURL(&types.MsgModifyNameRequest{}), "no name records available to modify"), nil, nil
		}

		newOwner, _ := simtypes.RandomAcc(r, accs)
		restrict := r.Intn(9) < 4
		msg := types.NewMsgModifyNameRequest(simAccount.Address.String(), randomRecord.Name, newOwner.Address, restrict)

		return Dispatch(r, app, ctx, ak, bk, simAccount, chainID, msg)
	}
}

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
) (
	simtypes.OperationMsg,
	[]simtypes.FutureOperation,
	error,
) {
	account := ak.GetAccount(ctx, from.Address)
	spendable := bk.SpendableCoins(ctx, account.GetAddress())

	fees, err := simtypes.RandomFees(r, ctx, spendable)
	if err != nil {
		return simtypes.NoOpMsg(sdk.MsgTypeURL(msg), sdk.MsgTypeURL(msg), "unable to generate fees"), nil, err
	}

	txGen := simappparams.MakeTestEncodingConfig().TxConfig
	tx, err := helpers.GenSignedMockTx(
		r,
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
		return simtypes.NoOpMsg(sdk.MsgTypeURL(msg), sdk.MsgTypeURL(msg), "unable to generate mock tx"), nil, err
	}

	_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
	if err != nil {
		return simtypes.NoOpMsg(sdk.MsgTypeURL(msg), sdk.MsgTypeURL(msg), err.Error()), nil, nil
	}

	return simtypes.NewOperationMsg(msg, true, "", &codec.ProtoCodec{}), nil, nil
}

// getRandomRecord finds a random record owned by a known account.
// An error is only returned if there was a problem iterating records.
func getRandomRecord(r *rand.Rand, ctx sdk.Context, k keeper.Keeper, accs []simtypes.Account, rootOK bool) (types.NameRecord, simtypes.Account, bool, error) {
	var randomRecord types.NameRecord
	var simAccount simtypes.Account

	var records []types.NameRecord
	err := k.IterateRecords(ctx, types.NameKeyPrefix, func(record types.NameRecord) error {
		if rootOK || strings.Contains(record.Name, ".") {
			records = append(records, record)
		}
		return nil
	})
	if err != nil || len(records) == 0 {
		return randomRecord, simAccount, false, err
	}

	r.Shuffle(len(records), func(i, j int) {
		records[i], records[j] = records[j], records[i]
	})

	found := false
	for _, randomRecord = range records {
		simAccount, found = simtypes.FindAccount(accs, sdk.MustAccAddressFromBech32(randomRecord.Address))
		if found {
			break
		}
	}

	return randomRecord, simAccount, found, nil
}
