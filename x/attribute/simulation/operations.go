package simulation

import (
	"fmt"
	"math/rand"

	"github.com/google/uuid"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	simappparams "github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/x/attribute/keeper"
	"github.com/provenance-io/provenance/x/attribute/types"
	namekeeper "github.com/provenance-io/provenance/x/name/keeper"
	nametypes "github.com/provenance-io/provenance/x/name/types"
)

// Simulation operation weights constants
const (
	//nolint:gosec // not credentials
	OpWeightMsgAddAttribute = "op_weight_msg_add_attribute"
	//nolint:gosec // not credentials
	OpWeightMsgUpdateAttribute = "op_weight_msg_update_attribute"
	//nolint:gosec // not credentials
	OpWeightMsgDeleteAttribute = "op_weight_msg_delete_attribute"
	//nolint:gosec // not credentials
	OpWeightMsgDeleteDistinctAttribute = "op_weight_msg_delete_distinct_attribute"
	//nolint:gosec // not credentials
	OpWeightMsgSetAccountData = "op_weight_msg_set_account_data"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	simState module.SimulationState, k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, nk namekeeper.Keeper,
) simulation.WeightedOperations {
	var (
		wMsgAddAttribute            int
		wMsgUpdateAttribute         int
		wMsgDeleteAttribute         int
		wMsgDeleteDistinctAttribute int
		wMsgSetAccountDataRequest   int
	)

	simState.AppParams.GetOrGenerate(OpWeightMsgAddAttribute, &wMsgAddAttribute, nil,
		func(_ *rand.Rand) { wMsgAddAttribute = simappparams.DefaultWeightMsgAddAttribute })
	simState.AppParams.GetOrGenerate(OpWeightMsgUpdateAttribute, &wMsgUpdateAttribute, nil,
		func(_ *rand.Rand) { wMsgUpdateAttribute = simappparams.DefaultWeightMsgUpdateAttribute })
	simState.AppParams.GetOrGenerate(OpWeightMsgDeleteAttribute, &wMsgDeleteAttribute, nil,
		func(_ *rand.Rand) { wMsgDeleteAttribute = simappparams.DefaultWeightMsgDeleteAttribute })
	simState.AppParams.GetOrGenerate(OpWeightMsgDeleteDistinctAttribute, &wMsgDeleteDistinctAttribute, nil,
		func(_ *rand.Rand) { wMsgDeleteDistinctAttribute = simappparams.DefaultWeightMsgDeleteDistinctAttribute })
	simState.AppParams.GetOrGenerate(OpWeightMsgSetAccountData, &wMsgSetAccountDataRequest, nil,
		func(_ *rand.Rand) { wMsgSetAccountDataRequest = simappparams.DefaultWeightMsgSetAccountData })

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(wMsgAddAttribute, SimulateMsgAddAttribute(simState, k, ak, bk, nk)),
		simulation.NewWeightedOperation(wMsgUpdateAttribute, SimulateMsgUpdateAttribute(simState, k, ak, bk, nk)),
		simulation.NewWeightedOperation(wMsgDeleteAttribute, SimulateMsgDeleteAttribute(simState, k, ak, bk, nk)),
		simulation.NewWeightedOperation(wMsgDeleteDistinctAttribute, SimulateMsgDeleteDistinctAttribute(simState, k, ak, bk, nk)),
		simulation.NewWeightedOperation(wMsgSetAccountDataRequest, SimulateMsgSetAccountData(simState, k, ak, bk)),
	}
}

// SimulateMsgAddAttribute will add an attribute under an account with a random type.
func SimulateMsgAddAttribute(simState module.SimulationState, _ keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, nk namekeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		randomRecord, simAccount, found, err := getRandomNameRecord(r, ctx, &nk, accs)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgAddAttributeRequest{}), "iterator of existing name records failed"), nil, err
		}
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgAddAttributeRequest{}), "no name records available to create under"), nil, nil
		}

		t := types.AttributeType(r.Intn(9)) //nolint:gosec // G115: r.Intn(9) will always fit in an int32 (implicit cast here).
		msg := types.NewMsgAddAttributeRequest(
			randomRecord.GetAddress(),
			simAccount.Address,
			randomRecord.Name,
			t,
			getRandomValueOfType(r, t),
		)

		return Dispatch(r, app, ctx, simState, ak, bk, simAccount, chainID, msg)
	}
}

// SimulateMsgUpdateAttribute will add an attribute under an account with a random type.
func SimulateMsgUpdateAttribute(simState module.SimulationState, k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, nk namekeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		randomAttribute, simAccount, found, err := getRandomAttribute(r, ctx, k, &nk, accs)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgUpdateAttributeRequest{}), "iterator of existing attributes failed"), nil, err
		}
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgUpdateAttributeRequest{}), "no attributes available to delete"), nil, nil
		}

		t := types.AttributeType(r.Intn(9)) //nolint:gosec // G115: r.Intn(9) will always fit in an int32 (implicit cast here).
		msg := types.NewMsgUpdateAttributeRequest(
			randomAttribute.GetAddress(),
			simAccount.Address,
			randomAttribute.Name,
			randomAttribute.Value,
			getRandomValueOfType(r, t),
			randomAttribute.AttributeType,
			t,
		)

		return Dispatch(r, app, ctx, simState, ak, bk, simAccount, chainID, msg)
	}
}

// SimulateMsgDeleteAttribute will dispatch a delete attribute operation against a random record
func SimulateMsgDeleteAttribute(simState module.SimulationState, k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, nk namekeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		randomAttribute, simAccount, found, err := getRandomAttribute(r, ctx, k, &nk, accs)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgDeleteAttributeRequest{}), "iterator of existing attributes failed"), nil, err
		}
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgDeleteAttributeRequest{}), "no attributes available to delete"), nil, nil
		}

		msg := types.NewMsgDeleteAttributeRequest(randomAttribute.Address, simAccount.Address, randomAttribute.Name)

		return Dispatch(r, app, ctx, simState, ak, bk, simAccount, chainID, msg)
	}
}

// SimulateMsgDeleteDistinctAttribute will dispatch a delete attribute operation against a random record
func SimulateMsgDeleteDistinctAttribute(simState module.SimulationState, k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, nk namekeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		randomAttribute, simAccount, found, err := getRandomAttribute(r, ctx, k, &nk, accs)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgDeleteDistinctAttributeRequest{}), "iterator of existing attributes failed"), nil, err
		}
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgDeleteDistinctAttributeRequest{}), "no attributes available to delete distinct"), nil, nil
		}

		msg := types.NewMsgDeleteDistinctAttributeRequest(randomAttribute.Address, simAccount.Address, randomAttribute.Name, randomAttribute.Value)

		return Dispatch(r, app, ctx, simState, ak, bk, simAccount, chainID, msg)
	}
}

// SimulateMsgSetAccountData will dispatch a set account data operation for a random account.
func SimulateMsgSetAccountData(simState module.SimulationState, k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// 1 in 10 chance that the value stays "".
		// 9 in 10 chance that it will be between 1 and MaxValueLen characters.
		value := ""
		if r.Intn(10) != 0 {
			maxLen := min(k.GetMaxValueLength(ctx), 500)
			strLen := r.Intn(int(maxLen)) + 1
			value = simtypes.RandStringOfLength(r, strLen)
		}

		acc, _ := simtypes.RandomAcc(r, accs)

		msg := &types.MsgSetAccountDataRequest{
			Value:   value,
			Account: acc.Address.String(),
		}

		return Dispatch(r, app, ctx, simState, ak, bk, acc, chainID, msg)
	}
}

func getRandomValueOfType(r *rand.Rand, t types.AttributeType) []byte {
	switch t {
	case types.AttributeType_Int:
		return []byte(fmt.Sprintf("%d", r.Int31()))
	case types.AttributeType_Bytes:
		return []byte(simtypes.RandStringOfLength(r, int(r.Int31n(20))))
	case types.AttributeType_String:
		return []byte(simtypes.RandStringOfLength(r, int(r.Int31n(20))))
	case types.AttributeType_UUID:
		id, _ := uuid.NewRandomFromReader(r)
		return []byte(id.String())
	case types.AttributeType_Float:
		return []byte(fmt.Sprintf("%f", r.Float32()))
	case types.AttributeType_Uri:
		return []byte("http://www.example.com/")
	case types.AttributeType_JSON:
		return []byte(`{"id":"value"}`)
	case types.AttributeType_Unspecified:
		return nil
	}
	return nil
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
		return simtypes.NoOpMsg(types.ModuleName, fmt.Sprintf("%T", msg), "unable to generate fees"), nil, err
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

// getRandomNameRecord finds a random name record owned by a known account.
// An error is only returned if there was a problem iterating records.
func getRandomNameRecord(r *rand.Rand, ctx sdk.Context, nk types.NameKeeper, accs []simtypes.Account) (nametypes.NameRecord, simtypes.Account, bool, error) {
	var records []nametypes.NameRecord
	err := nk.IterateRecords(ctx, nametypes.NameKeyPrefix, func(record nametypes.NameRecord) error {
		records = append(records, record)
		return nil
	})
	if err != nil || len(records) == 0 {
		return nametypes.NameRecord{}, simtypes.Account{}, false, err
	}

	r.Shuffle(len(records), func(i, j int) {
		records[i], records[j] = records[j], records[i]
	})

	for _, record := range records {
		simAccount, found := simtypes.FindAccount(accs, sdk.MustAccAddressFromBech32(record.Address))
		if found {
			return record, simAccount, true, nil
		}
	}

	return nametypes.NameRecord{}, simtypes.Account{}, false, nil
}

// getRandomAttribute finds a random attribute owned by a known account.
// An error is only returned if there was a problem iterating records.
// The sim account returned is the one that owns the name record for the attribute.
func getRandomAttribute(r *rand.Rand, ctx sdk.Context, k keeper.Keeper, nk types.NameKeeper, accs []simtypes.Account) (types.Attribute, simtypes.Account, bool, error) {
	var attributes []types.Attribute
	err := k.IterateRecords(ctx, types.AttributeKeyPrefix, func(attribute types.Attribute) error {
		attributes = append(attributes, attribute)
		return nil
	})
	if err != nil || len(attributes) == 0 {
		return types.Attribute{}, simtypes.Account{}, false, err
	}

	r.Shuffle(len(attributes), func(i, j int) {
		attributes[i], attributes[j] = attributes[j], attributes[i]
	})

	for _, attr := range attributes {
		nr, err := nk.GetRecordByName(ctx, attr.Name)
		if err == nil {
			simAccount, found := simtypes.FindAccount(accs, sdk.MustAccAddressFromBech32(nr.Address))
			if found {
				return attr, simAccount, true, nil
			}
		}
	}

	return types.Attribute{}, simtypes.Account{}, false, nil
}
