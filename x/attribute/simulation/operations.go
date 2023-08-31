package simulation

import (
	"fmt"
	"math/rand"

	"github.com/google/uuid"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	appParams simtypes.AppParams, cdc codec.JSONCodec, k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, nk namekeeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgAddAttribute            int
		weightMsgUpdateAttribute         int
		weightMsgDeleteAttribute         int
		weightMsgDeleteDistinctAttribute int
		weightMsgSetAccountDataRequest   int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgAddAttribute, &weightMsgAddAttribute, nil,
		func(_ *rand.Rand) {
			weightMsgAddAttribute = simappparams.DefaultWeightMsgAddAttribute
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgUpdateAttribute, &weightMsgUpdateAttribute, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateAttribute = simappparams.DefaultWeightMsgUpdateAttribute
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgDeleteAttribute, &weightMsgDeleteAttribute, nil,
		func(_ *rand.Rand) {
			weightMsgDeleteAttribute = simappparams.DefaultWeightMsgDeleteAttribute
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgDeleteDistinctAttribute, &weightMsgDeleteDistinctAttribute, nil,
		func(_ *rand.Rand) {
			weightMsgDeleteDistinctAttribute = simappparams.DefaultWeightMsgDeleteDistinctAttribute
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgSetAccountData, &weightMsgSetAccountDataRequest, nil,
		func(_ *rand.Rand) {
			weightMsgSetAccountDataRequest = simappparams.DefaultWeightMsgSetAccountData
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgAddAttribute,
			SimulateMsgAddAttribute(k, ak, bk, nk),
		),
		simulation.NewWeightedOperation(
			weightMsgUpdateAttribute,
			SimulateMsgUpdateAttribute(k, ak, bk, nk),
		),
		simulation.NewWeightedOperation(
			weightMsgDeleteAttribute,
			SimulateMsgDeleteAttribute(k, ak, bk, nk),
		),
		simulation.NewWeightedOperation(
			weightMsgDeleteDistinctAttribute,
			SimulateMsgDeleteDistinctAttribute(k, ak, bk, nk),
		),
		simulation.NewWeightedOperation(
			weightMsgSetAccountDataRequest,
			SimulateMsgSetAccountData(k, ak, bk),
		),
	}
}

// SimulateMsgAddAttribute will add an attribute under an account with a random type.
func SimulateMsgAddAttribute(_ keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, nk namekeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		randomRecord, simAccount, found, err := getRandomNameRecord(r, ctx, &nk, accs)
		if err != nil {
			return simtypes.NoOpMsg(sdk.MsgTypeURL(&types.MsgAddAttributeRequest{}), sdk.MsgTypeURL(&types.MsgAddAttributeRequest{}), "iterator of existing name records failed"), nil, err
		}
		if !found {
			return simtypes.NoOpMsg(sdk.MsgTypeURL(&types.MsgAddAttributeRequest{}), sdk.MsgTypeURL(&types.MsgAddAttributeRequest{}), "no name records available to create under"), nil, nil
		}

		t := types.AttributeType(r.Intn(9))
		msg := types.NewMsgAddAttributeRequest(
			randomRecord.GetAddress(),
			simAccount.Address,
			randomRecord.Name,
			t,
			getRandomValueOfType(r, t),
		)

		return Dispatch(r, app, ctx, ak, bk, simAccount, chainID, msg)
	}
}

// SimulateMsgUpdateAttribute will add an attribute under an account with a random type.
func SimulateMsgUpdateAttribute(k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, nk namekeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		randomAttribute, simAccount, found, err := getRandomAttribute(r, ctx, k, &nk, accs)
		if err != nil {
			return simtypes.NoOpMsg(sdk.MsgTypeURL(&types.MsgUpdateAttributeRequest{}), sdk.MsgTypeURL(&types.MsgUpdateAttributeRequest{}), "iterator of existing attributes failed"), nil, err
		}
		if !found {
			return simtypes.NoOpMsg(sdk.MsgTypeURL(&types.MsgUpdateAttributeRequest{}), sdk.MsgTypeURL(&types.MsgUpdateAttributeRequest{}), "no attributes available to delete"), nil, nil
		}

		t := types.AttributeType(r.Intn(9))
		msg := types.NewMsgUpdateAttributeRequest(
			randomAttribute.GetAddress(),
			simAccount.Address,
			randomAttribute.Name,
			randomAttribute.Value,
			getRandomValueOfType(r, t),
			randomAttribute.AttributeType,
			t,
		)

		return Dispatch(r, app, ctx, ak, bk, simAccount, chainID, msg)
	}
}

// SimulateMsgDeleteAttribute will dispatch a delete attribute operation against a random record
func SimulateMsgDeleteAttribute(k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, nk namekeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		randomAttribute, simAccount, found, err := getRandomAttribute(r, ctx, k, &nk, accs)
		if err != nil {
			return simtypes.NoOpMsg(sdk.MsgTypeURL(&types.MsgDeleteAttributeRequest{}), sdk.MsgTypeURL(&types.MsgDeleteAttributeRequest{}), "iterator of existing attributes failed"), nil, err
		}
		if !found {
			return simtypes.NoOpMsg(sdk.MsgTypeURL(&types.MsgDeleteAttributeRequest{}), sdk.MsgTypeURL(&types.MsgDeleteAttributeRequest{}), "no attributes available to delete"), nil, nil
		}

		msg := types.NewMsgDeleteAttributeRequest(randomAttribute.Address, simAccount.Address, randomAttribute.Name)

		return Dispatch(r, app, ctx, ak, bk, simAccount, chainID, msg)
	}
}

// SimulateMsgDeleteDistinctAttribute will dispatch a delete attribute operation against a random record
func SimulateMsgDeleteDistinctAttribute(k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, nk namekeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		randomAttribute, simAccount, found, err := getRandomAttribute(r, ctx, k, &nk, accs)
		if err != nil {
			return simtypes.NoOpMsg(sdk.MsgTypeURL(&types.MsgDeleteDistinctAttributeRequest{}), sdk.MsgTypeURL(&types.MsgDeleteDistinctAttributeRequest{}), "iterator of existing attributes failed"), nil, err
		}
		if !found {
			return simtypes.NoOpMsg(sdk.MsgTypeURL(&types.MsgDeleteDistinctAttributeRequest{}), sdk.MsgTypeURL(&types.MsgDeleteDistinctAttributeRequest{}), "no attributes available to delete distinct"), nil, nil
		}

		msg := types.NewMsgDeleteDistinctAttributeRequest(randomAttribute.Address, simAccount.Address, randomAttribute.Name, randomAttribute.Value)

		return Dispatch(r, app, ctx, ak, bk, simAccount, chainID, msg)
	}
}

// SimulateMsgSetAccountData will dispatch a set account data operation for a random account.
func SimulateMsgSetAccountData(k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// 1 in 10 chance that the value stays "".
		// 9 in 10 chance that it will be between 1 and MaxValueLen characters.
		value := ""
		if r.Intn(10) != 0 {
			maxLen := uint(k.GetMaxValueLength(ctx))
			if maxLen > 500 {
				maxLen = 500
			}
			strLen := r.Intn(int(maxLen)) + 1
			value = simtypes.RandStringOfLength(r, strLen)
		}

		acc, _ := simtypes.RandomAcc(r, accs)

		msg := &types.MsgSetAccountDataRequest{
			Value:   value,
			Account: acc.Address.String(),
		}

		return Dispatch(r, app, ctx, ak, bk, acc, chainID, msg)
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
