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
	OpWeightMsgAddAttribute    = "op_weight_msg_add_attribute"
	OpWeightMsgDeleteAttribute = "op_weight_msg_delete_attribute"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONMarshaler, k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, nk namekeeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgAddAttribute    int
		weightMsgDeleteAttribute int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgAddAttribute, &weightMsgAddAttribute, nil,
		func(_ *rand.Rand) {
			weightMsgAddAttribute = simappparams.DefaultWeightMsgAddAttribute
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgDeleteAttribute, &weightMsgDeleteAttribute, nil,
		func(_ *rand.Rand) {
			weightMsgDeleteAttribute = simappparams.DefaultWeightMsgDeleteAttriubte
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgAddAttribute,
			SimulateMsgAddAttribute(k, ak, bk, nk),
		),
		simulation.NewWeightedOperation(
			weightMsgDeleteAttribute,
			SimulateMsgDeleteAttribute(k, ak, bk, nk),
		),
	}
}

// SimulateMsgAddAttribute will add an attribute under an account with a random type.
func SimulateMsgAddAttribute(k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper, nk namekeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		//		simAccount, _ := simtypes.RandomAcc(r, accs)

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

		randomRecord := records[r.Intn(len(records))]
		simAccount, _ := simtypes.FindAccount(accs, mustGetAddress(randomRecord.Address))

		t := types.AttributeType(r.Intn(9))
		msg := types.NewMsgAddAttributeRequest(
			mustGetAddress(randomRecord.GetAddress()),
			simAccount.Address,
			randomRecord.Name,
			t,
			getRandomValueOfType(r, t),
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

		nr, err := nk.GetRecordByName(ctx, randomAttribute.Name)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDeleteAttribute, "name record for existing attributes not found"), nil, err
		}

		simAccount, _ := simtypes.FindAccount(accs, mustGetAddress(nr.Address))
		msg := types.NewMsgDeleteAttributeRequest(mustGetAddress(randomAttribute.Address), mustGetAddress(nr.Address), randomAttribute.Name)

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
