package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/google/uuid"

	simappparams "github.com/provenance-io/provenance/app/params"

	keeper "github.com/provenance-io/provenance/x/attribute/keeper"
	types "github.com/provenance-io/provenance/x/attribute/types"
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
	}
}

// SimulateMsgAddAttribute will add an attribute under an account with a random type.
func SimulateMsgAddAttribute(k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, nk namekeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		var records []nametypes.NameRecord
		if err := nk.IterateRecords(ctx, nametypes.NameKeyPrefix, func(record nametypes.NameRecord) error {
			records = append(records, record)
			return nil
		}); err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgAddAttribute, "iterator of existing name records failed"), nil, err
		}

		if len(records) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgAddAttribute, "no name records available to create under"), nil, nil
		}

		found := false
		var simAccount simtypes.Account
		var randomRecord nametypes.NameRecord

		for !found {
			randomRecord = records[r.Intn(len(records))]
			simAccount, found = simtypes.FindAccount(accs, mustGetAddress(randomRecord.Address))
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
		var attributes []types.Attribute
		if err := k.IterateRecords(ctx, types.AttributeKeyPrefix, func(attribute types.Attribute) error {
			attributes = append(attributes, attribute)
			return nil
		}); err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUpdateAttribute, "iterator of existing attributes failed"), nil, err
		}

		if len(attributes) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUpdateAttribute, "no attributes available to delete"), nil, nil
		}

		randomAttribute := attributes[r.Intn(len(attributes))]

		// the name associated with this attribute may no longer exist so use the attribute account as a backup account for "owner"
		var ownerAddress string
		nr, err := nk.GetRecordByName(ctx, randomAttribute.Name)
		if err == nil {
			ownerAddress = nr.Address
		} else {
			ownerAddress = randomAttribute.Address
		}

		t := types.AttributeType(r.Intn(9))
		simAccount, _ := simtypes.FindAccount(accs, mustGetAddress(ownerAddress))
		msg := types.NewMsgUpdateAttributeRequest(
			randomAttribute.GetAddress(),
			mustGetAddress(ownerAddress),
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
		var attributes []types.Attribute
		if err := k.IterateRecords(ctx, types.AttributeKeyPrefix, func(attribute types.Attribute) error {
			attributes = append(attributes, attribute)
			return nil
		}); err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDeleteAttribute, "iterator of existing attributes failed"), nil, err
		}

		if len(attributes) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDeleteAttribute, "no attributes available to delete"), nil, nil
		}

		randomAttribute := attributes[r.Intn(len(attributes))]

		// the name associated with this attribute may no longer exist so use the attribute account as a backup account for "owner"
		var ownerAddress string
		nr, err := nk.GetRecordByName(ctx, randomAttribute.Name)
		if err == nil {
			ownerAddress = nr.Address
		} else {
			ownerAddress = randomAttribute.Address
		}

		simAccount, _ := simtypes.FindAccount(accs, mustGetAddress(ownerAddress))
		msg := types.NewMsgDeleteAttributeRequest(randomAttribute.Address, mustGetAddress(ownerAddress), randomAttribute.Name)

		return Dispatch(r, app, ctx, ak, bk, simAccount, chainID, msg)
	}
}

// SimulateMsgDeleteDistinctAttribute will dispatch a delete attribute operation against a random record
func SimulateMsgDeleteDistinctAttribute(k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, nk namekeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		var attributes []types.Attribute
		if err := k.IterateRecords(ctx, types.AttributeKeyPrefix, func(attribute types.Attribute) error {
			attributes = append(attributes, attribute)
			return nil
		}); err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDeleteDistinctAttribute, "iterator of existing attributes failed"), nil, err
		}

		if len(attributes) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDeleteDistinctAttribute, "no attributes available to delete distinct"), nil, nil
		}

		randomAttribute := attributes[r.Intn(len(attributes))]

		// the name associated with this attribute may no longer exist so use the attribute account as a backup account for "owner"
		var ownerAddress string
		nr, err := nk.GetRecordByName(ctx, randomAttribute.Name)
		if err == nil {
			ownerAddress = nr.Address
		} else {
			ownerAddress = randomAttribute.Address
		}

		simAccount, _ := simtypes.FindAccount(accs, mustGetAddress(ownerAddress))
		msg := types.NewMsgDeleteDistinctAttributeRequest(randomAttribute.Address, mustGetAddress(ownerAddress), randomAttribute.Name, randomAttribute.Value)

		return Dispatch(r, app, ctx, ak, bk, simAccount, chainID, msg)
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

func mustGetAddress(addr string) sdk.AccAddress {
	a, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		panic(err)
	}
	return a
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
		return simtypes.NoOpMsg(types.ModuleName, fmt.Sprintf("%T", msg), "unable to generate mock tx"), nil, err
	}

	_, _, err = app.Deliver(txGen.TxEncoder(), tx)
	if err != nil {
		return simtypes.NoOpMsg(types.ModuleName, fmt.Sprintf("%T", msg), err.Error()), nil, nil
	}

	return simtypes.NewOperationMsg(msg, true, "", &codec.ProtoCodec{}), nil, nil
}
