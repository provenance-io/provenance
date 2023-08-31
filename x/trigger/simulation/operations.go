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
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	simappparams "github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/x/trigger/keeper"
	"github.com/provenance-io/provenance/x/trigger/types"
)

// Simulation operation weights constants
const (
	//nolint:gosec // not credentials
	OpWeightMsgCreateTrigger = "op_weight_msg_create_trigger"
	//nolint:gosec // not credentials
	OpWeightMsgDestroyTrigger = "op_weight_msg_destroy_trigger"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgCreateTrigger  int
		weightMsgDestroyTrigger int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgCreateTrigger, &weightMsgCreateTrigger, nil,
		func(_ *rand.Rand) {
			weightMsgCreateTrigger = simappparams.DefaultWeightSubmitCreateTrigger
		},
	)
	appParams.GetOrGenerate(cdc, OpWeightMsgDestroyTrigger, &weightMsgDestroyTrigger, nil,
		func(_ *rand.Rand) {
			weightMsgDestroyTrigger = simappparams.DefaultWeightSubmitDestroyTrigger
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgCreateTrigger,
			SimulateMsgCreateTrigger(k, ak, bk),
		),
		simulation.NewWeightedOperation(
			weightMsgDestroyTrigger,
			SimulateMsgDestroyTrigger(k, ak, bk),
		),
	}
}

// SimulateMsgCreateTrigger sends a MsgCreateTriggerRequest.
func SimulateMsgCreateTrigger(_ keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		now := ctx.BlockTime()
		raccs, err := RandomAccs(r, accs, 2)
		if err != nil {
			return simtypes.NoOpMsg(sdk.MsgTypeURL(&types.MsgCreateTriggerRequest{}), sdk.MsgTypeURL(&types.MsgCreateTriggerRequest{}), err.Error()), nil, nil
		}
		from := raccs[0]
		to := raccs[1]

		msg, err := types.NewCreateTriggerRequest([]string{from.Address.String()}, NewRandomEvent(r, now), []sdk.Msg{NewRandomAction(r, from.Address.String(), to.Address.String())})
		if err != nil {
			return simtypes.NoOpMsg(sdk.MsgTypeURL(msg), sdk.MsgTypeURL(msg), "error creating message"), nil, err
		}

		return Dispatch(r, app, ctx, from, chainID, msg, ak, bk, nil)
	}
}

// SimulateMsgDestroyTrigger sends a MsgDestroyTriggerRequest.
func SimulateMsgDestroyTrigger(k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		trigger := randomTrigger(r, ctx, k)
		if trigger == nil {
			return simtypes.NoOpMsg(sdk.MsgTypeURL(&types.MsgDestroyTriggerRequest{}), sdk.MsgTypeURL(&types.MsgDestroyTriggerRequest{}), "unable to find a valid trigger"), nil, nil
		}
		addr, err := sdk.AccAddressFromBech32(trigger.Owner)
		if err != nil {
			// should not be possible and panic on the test
			panic(err)
		}
		simAccount, found := simtypes.FindAccount(accs, addr)
		if !found {
			return simtypes.NoOpMsg(sdk.MsgTypeURL(&types.MsgDestroyTriggerRequest{}), sdk.MsgTypeURL(&types.MsgDestroyTriggerRequest{}), "creator of trigger does not exist"), nil, nil
		}
		msg := types.NewDestroyTriggerRequest(trigger.GetOwner(), trigger.GetId())
		return Dispatch(r, app, ctx, simAccount, chainID, msg, ak, bk, nil)
	}
}

// Dispatch sends an operation to the chain using a given account/funds on account for fees.  Failures on the server side
// are handled as no-op msg operations with the error string as the status/response.
func Dispatch(
	r *rand.Rand,
	app *baseapp.BaseApp,
	ctx sdk.Context,
	from simtypes.Account,
	chainID string,
	msg sdk.Msg,
	ak authkeeper.AccountKeeperI,
	bk bankkeeper.Keeper,
	futures []simtypes.FutureOperation,
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
	err = testutil.FundAccount(bk, ctx, account.GetAddress(), sdk.NewCoins(sdk.Coin{
		Denom:  pioconfig.GetProvenanceConfig().BondDenom,
		Amount: sdk.NewInt(1_000_000_000_000_000),
	}))
	if err != nil {
		return simtypes.NoOpMsg(sdk.MsgTypeURL(msg), sdk.MsgTypeURL(msg), "unable to fund account"), nil, err
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

	return simtypes.NewOperationMsg(msg, true, "", &codec.ProtoCodec{}), futures, nil
}

func randomTrigger(r *rand.Rand, ctx sdk.Context, k keeper.Keeper) *types.Trigger {
	triggers, err := k.GetAllTriggers(ctx)
	if err != nil {
		// sim tests should fail if iterator errors
		panic(err)
	}
	if len(triggers) == 0 {
		return nil
	}
	idx := r.Intn(len(triggers))
	return &triggers[idx]
}

func RandomAccs(r *rand.Rand, accs []simtypes.Account, count uint64) ([]simtypes.Account, error) {
	if uint64(len(accs)) < count {
		return nil, fmt.Errorf("cannot choose %d accounts because there are only %d", count, len(accs))
	}
	raccs := make([]simtypes.Account, 0, len(accs))
	raccs = append(raccs, accs...)
	r.Shuffle(len(raccs), func(i, j int) {
		raccs[i], raccs[j] = raccs[j], raccs[i]
	})
	return raccs[:count], nil
}
