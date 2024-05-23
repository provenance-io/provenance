package simulation

import (
	"math/rand"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	simappparams "github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/internal/helpers"
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
	simState module.SimulationState, k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.Keeper,
) simulation.WeightedOperations {
	var (
		wMsgCreateTrigger  int
		wMsgDestroyTrigger int
	)

	simState.AppParams.GetOrGenerate(OpWeightMsgCreateTrigger, &wMsgCreateTrigger, nil,
		func(_ *rand.Rand) { wMsgCreateTrigger = simappparams.DefaultWeightSubmitCreateTrigger })
	simState.AppParams.GetOrGenerate(OpWeightMsgDestroyTrigger, &wMsgDestroyTrigger, nil,
		func(_ *rand.Rand) { wMsgDestroyTrigger = simappparams.DefaultWeightSubmitDestroyTrigger })

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(wMsgCreateTrigger, SimulateMsgCreateTrigger(simState, k, ak, bk)),
		simulation.NewWeightedOperation(wMsgDestroyTrigger, SimulateMsgDestroyTrigger(simState, k, ak, bk)),
	}
}

// SimulateMsgCreateTrigger sends a MsgCreateTriggerRequest.
func SimulateMsgCreateTrigger(simState module.SimulationState, _ keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		now := ctx.BlockTime()
		raccs, err := helpers.SelectRandomEntries(r, accs, 2)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgCreateTriggerRequest{}), err.Error()), nil, nil
		}
		from := raccs[0]
		to := raccs[1]

		msg, err := types.NewCreateTriggerRequest([]string{from.Address.String()}, NewRandomEvent(r, now), []sdk.Msg{NewRandomAction(r, from.Address.String(), to.Address.String())})
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "error creating message"), nil, err
		}

		return Dispatch(r, app, ctx, simState, from, chainID, msg, ak, bk, nil)
	}
}

// SimulateMsgDestroyTrigger sends a MsgDestroyTriggerRequest.
func SimulateMsgDestroyTrigger(simState module.SimulationState, k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		trigger := randomTrigger(r, ctx, k)
		if trigger == nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgDestroyTriggerRequest{}), "unable to find a valid trigger"), nil, nil
		}
		addr, err := sdk.AccAddressFromBech32(trigger.Owner)
		if err != nil {
			// should not be possible and panic on the test
			panic(err)
		}
		simAccount, found := simtypes.FindAccount(accs, addr)
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgDestroyTriggerRequest{}), "creator of trigger does not exist"), nil, nil
		}
		msg := types.NewDestroyTriggerRequest(trigger.GetOwner(), trigger.GetId())
		return Dispatch(r, app, ctx, simState, simAccount, chainID, msg, ak, bk, nil)
	}
}

// Dispatch sends an operation to the chain using a given account/funds on account for fees.  Failures on the server side
// are handled as no-op msg operations with the error string as the status/response.
func Dispatch(
	r *rand.Rand,
	app *baseapp.BaseApp,
	ctx sdk.Context,
	simState module.SimulationState,
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
		return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "unable to generate fees"), nil, err
	}
	err = testutil.FundAccount(ctx, bk, account.GetAddress(), sdk.NewCoins(sdk.Coin{
		Denom:  pioconfig.GetProvenanceConfig().BondDenom,
		Amount: sdkmath.NewInt(1_000_000_000_000_000),
	}))
	if err != nil {
		return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "unable to fund account"), nil, err
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

	return simtypes.NewOperationMsg(msg, true, ""), futures, nil
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
