package simulation

import (
	"fmt"
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
	channelkeeper "github.com/cosmos/ibc-go/v8/modules/core/04-channel/keeper"

	simappparams "github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/x/oracle/keeper"
	"github.com/provenance-io/provenance/x/oracle/types"
)

// Simulation operation weights constants
const (
	//nolint:gosec // not credentials
	OpWeightMsgUpdateOracle = "op_weight_msg_update_oracle"
	//nolint:gosec // not credentials
	OpWeightMsgSendOracleQuery = "op_weight_msg_send_oracle_query"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	simState module.SimulationState, k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.Keeper, ck channelkeeper.Keeper,
) simulation.WeightedOperations {
	var (
		wMsgUpdateOracle    int
		wMsgSendOracleQuery int
	)

	simState.AppParams.GetOrGenerate(OpWeightMsgUpdateOracle, &wMsgUpdateOracle, nil,
		func(_ *rand.Rand) { wMsgUpdateOracle = simappparams.DefaultWeightUpdateOracle })
	simState.AppParams.GetOrGenerate(OpWeightMsgSendOracleQuery, &wMsgSendOracleQuery, nil,
		func(_ *rand.Rand) { wMsgSendOracleQuery = simappparams.DefaultWeightSendOracleQuery })

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(wMsgUpdateOracle, SimulateMsgUpdateOracle(simState, k, ak, bk)),
		simulation.NewWeightedOperation(wMsgSendOracleQuery, SimulateMsgSendQueryOracle(simState, k, ak, bk, ck)),
	}
}

// SimulateMsgUpdateOracle sends a MsgUpdateOracle.
func SimulateMsgUpdateOracle(simState module.SimulationState, _ keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		raccs, err := RandomAccs(r, accs, uint64(len(accs)))
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgUpdateOracleRequest{}), err.Error()), nil, nil
		}

		// 50% chance to be from the module's authority
		from := raccs[0]
		to := raccs[1]

		// TODO[1760]: Submit this as a gov prop and also return futures for the votes.
		msg := types.NewMsgUpdateOracle(from.Address.String(), to.Address.String())

		return Dispatch(r, app, ctx, simState, from, chainID, msg, ak, bk, nil)
	}
}

// SimulateMsgSendQueryOracle sends a MsgSendQueryOracle.
func SimulateMsgSendQueryOracle(simState module.SimulationState, _ keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.Keeper, ck channelkeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		raccs, err := RandomAccs(r, accs, 1)

		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgSendQueryOracleRequest{}), err.Error()), nil, nil
		}
		addr := raccs[0]

		channel, err := randomChannel(r, ctx, ck)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgSendQueryOracleRequest{}), err.Error()), nil, nil
		}
		query := randomQuery(r)

		msg := types.NewMsgSendQueryOracle(addr.Address.String(), channel, query)
		return Dispatch(r, app, ctx, simState, addr, chainID, msg, ak, bk, nil)
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

func randomChannel(r *rand.Rand, ctx sdk.Context, ck channelkeeper.Keeper) (string, error) {
	channels := ck.GetAllChannels(ctx)
	if len(channels) == 0 {
		return "", fmt.Errorf("cannot get random channel because none exist")
	}
	idx := r.Intn(len(channels))
	return channels[idx].String(), nil
}

func randomQuery(r *rand.Rand) []byte {
	queryType := randIntBetween(r, 0, 3)
	var query string
	switch queryType {
	case 0:
		query = ""
	case 1:
		query = "{}"
	case 2:
		query = "{\"version\":{}}"
	default:
		query = "xyz"
	}

	return []byte(query)
}
