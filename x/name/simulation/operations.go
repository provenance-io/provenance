package simulation

import (
	"math/rand"
	"strings"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	simappparams "github.com/provenance-io/provenance/app/params"
	internalrand "github.com/provenance-io/provenance/internal/rand"
	"github.com/provenance-io/provenance/x/name/keeper"
	"github.com/provenance-io/provenance/x/name/types"
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
	simState module.SimulationState, k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper,
) simulation.WeightedOperations {
	var (
		wMsgBindName   int
		wMsgDeleteName int
		wMsgModifyName int
	)

	simState.AppParams.GetOrGenerate(OpWeightMsgBindName, &wMsgBindName, nil,
		func(_ *rand.Rand) { wMsgBindName = simappparams.DefaultWeightMsgBindName })
	simState.AppParams.GetOrGenerate(OpWeightMsgDeleteName, &wMsgDeleteName, nil,
		func(_ *rand.Rand) { wMsgDeleteName = simappparams.DefaultWeightMsgDeleteName })
	simState.AppParams.GetOrGenerate(OpWeightMsgModifyName, &wMsgModifyName, nil,
		func(_ *rand.Rand) { wMsgModifyName = simappparams.DefaultWeightMsgModifyName })

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(wMsgBindName, SimulateMsgBindName(simState, k, ak, bk)),
		simulation.NewWeightedOperation(wMsgDeleteName, SimulateMsgDeleteName(simState, k, ak, bk)),
		simulation.NewWeightedOperation(wMsgModifyName, SimulateMsgModifyName(simState, k, ak, bk)),
	}
}

// SimulateMsgBindName will bind a NAME under an existing name using a 40% probability of restricting it.
func SimulateMsgBindName(simState module.SimulationState, k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		params := k.GetParams(ctx)
		parentRecord, parentOwner, found, err := getRandomRecord(r, ctx, k, accs, 1, int(params.MaxNameLevels)-1)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgBindNameRequest{}), "iterator of existing records failed"), nil, err
		}
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgBindNameRequest{}), "no name records available to create under"), nil, nil
		}

		nameLen := internalrand.IntBetween(r, int(params.GetMinSegmentLength()), int(params.GetMaxSegmentLength()))
		newRecordName := simtypes.RandStringOfLength(r, nameLen)
		newRecordOwner := parentOwner
		if !parentRecord.Restricted {
			newRecordOwner, _ = simtypes.RandomAcc(r, accs)
		}
		newRecordRestricted := r.Intn(9) < 4
		newRecord := types.NewNameRecord(newRecordName, newRecordOwner.Address, newRecordRestricted)
		msg := types.NewMsgBindNameRequest(newRecord, parentRecord)

		return Dispatch(r, app, ctx, simState, ak, bk, parentOwner, chainID, msg)
	}
}

// SimulateMsgDeleteName will dispatch a delete name operation against a random name record
func SimulateMsgDeleteName(simState module.SimulationState, k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// minSeg = 2 because we don't want to delete any root name records.
		// maxSeg = 1_000_000 because that should be more than any name has.
		// Not doing a min/max params lookup because they can change during the sim and don't apply to this operation.
		randomRecord, simAccount, found, err := getRandomRecord(r, ctx, k, accs, 2, 1_000_000)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgDeleteNameRequest{}), "iterator of existing records failed"), nil, err
		}
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgDeleteNameRequest{}), "no name records available to delete"), nil, nil
		}

		msg := types.NewMsgDeleteNameRequest(randomRecord)

		return Dispatch(r, app, ctx, simState, ak, bk, simAccount, chainID, msg)
	}
}

// SimulateMsgModifyName will dispatch a modify name operation against a random name record
func SimulateMsgModifyName(simState module.SimulationState, k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		params := k.GetParams(ctx)
		randomRecord, simAccount, found, err := getRandomRecord(r, ctx, k, accs, 1, int(params.MaxNameLevels))
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgModifyNameRequest{}), "iterator of existing records failed"), nil, err
		}
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgModifyNameRequest{}), "no name records available to modify"), nil, nil
		}

		newOwner, _ := simtypes.RandomAcc(r, accs)
		restrict := r.Intn(9) < 4
		msg := types.NewMsgModifyNameRequest(simAccount.Address.String(), randomRecord.Name, newOwner.Address, restrict)

		return Dispatch(r, app, ctx, simState, ak, bk, simAccount, chainID, msg)
	}
}

// Dispatch sends an operation to the chain using a given account/funds on account for fees.  Failures on the server side
// are handled as no-op msg operations with the error string as the status/response.
func Dispatch(
	r *rand.Rand,
	app *baseapp.BaseApp,
	ctx sdk.Context,
	simState module.SimulationState,
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
		return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "unable to generate fees"), nil, err
	}

	tx, err := simtestutil.GenSignedMockTx(
		r,
		simState.TxConfig,
		[]sdk.Msg{msg},
		fees,
		simtestutil.DefaultGenTxGas,
		chainID,
		[]uint64{account.GetAccountNumber()},
		[]uint64{account.GetSequence()},
		from.PrivKey,
	)
	if err != nil {
		return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "unable to generate mock tx"), nil, err
	}

	_, _, err = app.SimDeliver(simState.TxConfig.TxEncoder(), tx)
	if err != nil {
		return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), err.Error()), nil, nil
	}

	return simtypes.NewOperationMsg(msg, true, ""), nil, nil
}

// getRandomRecord finds a random record owned by a known account.
// An error is only returned if there was a problem iterating records.
func getRandomRecord(r *rand.Rand, ctx sdk.Context, k keeper.Keeper, accs []simtypes.Account, segmentsMin, segmentsMax int) (types.NameRecord, simtypes.Account, bool, error) {
	var records []types.NameRecord
	err := k.IterateRecords(ctx, types.NameKeyPrefix, func(record types.NameRecord) error {
		segments := strings.Count(record.Name, ".") + 1
		if segmentsMin <= segments && segments <= segmentsMax {
			records = append(records, record)
		}
		return nil
	})
	if err != nil || len(records) == 0 {
		return types.NameRecord{}, simtypes.Account{}, false, err
	}

	r.Shuffle(len(records), func(i, j int) {
		records[i], records[j] = records[j], records[i]
	})

	for _, randomRecord := range records {
		simAccount, found := simtypes.FindAccount(accs, sdk.MustAccAddressFromBech32(randomRecord.Address))
		if found {
			return randomRecord, simAccount, true, nil
		}
	}

	return types.NameRecord{}, simtypes.Account{}, false, nil
}
