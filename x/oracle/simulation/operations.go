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
	channelkeeper "github.com/cosmos/ibc-go/v6/modules/core/04-channel/keeper"
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
	appParams simtypes.AppParams, cdc codec.JSONCodec, k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.Keeper, ck channelkeeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgUpdateOracle    int
		weightMsgSendOracleQuery int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgUpdateOracle, &weightMsgUpdateOracle, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateOracle = simappparams.DefaultWeightUpdateOracle
		},
	)
	appParams.GetOrGenerate(cdc, OpWeightMsgSendOracleQuery, &weightMsgSendOracleQuery, nil,
		func(_ *rand.Rand) {
			weightMsgSendOracleQuery = simappparams.DefaultWeightSendOracleQuery
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgUpdateOracle,
			SimulateMsgUpdateOracle(k, ak, bk),
		),
		simulation.NewWeightedOperation(
			weightMsgSendOracleQuery,
			SimulateMsgSendQueryOracle(k, ak, bk, ck),
		),
	}
}

// SimulateMsgCreateTrigger sends a MsgUpdateOracle.
func SimulateMsgUpdateOracle(_ keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		raccs, err := RandomAccs(r, accs, 2)
		if err != nil {
			return simtypes.NoOpMsg(sdk.MsgTypeURL(&types.MsgUpdateOracleRequest{}), sdk.MsgTypeURL(&types.MsgUpdateOracleRequest{}), err.Error()), nil, nil
		}
		from := raccs[0]
		to := raccs[1]

		msg := types.NewMsgUpdateOracle(from.Address.String(), to.Address.String())

		return Dispatch(r, app, ctx, from, chainID, msg, ak, bk, nil)
	}
}

// SimulateMsgSendQueryOracle sends a MsgSendQueryOracle.
func SimulateMsgSendQueryOracle(k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.Keeper, ck channelkeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		raccs, err := RandomAccs(r, accs, 1)
		if err != nil {
			return simtypes.NoOpMsg(sdk.MsgTypeURL(&types.MsgSendQueryOracleRequest{}), sdk.MsgTypeURL(&types.MsgSendQueryOracleRequest{}), err.Error()), nil, nil
		}
		addr := raccs[0]

		channel, err := randomChannel(r, ctx, ck)
		if err != nil {
			return simtypes.NoOpMsg(sdk.MsgTypeURL(&types.MsgSendQueryOracleRequest{}), sdk.MsgTypeURL(&types.MsgSendQueryOracleRequest{}), err.Error()), nil, nil
		}
		query := []byte("{}")

		msg := types.NewMsgSendQueryOracle(addr.Address.String(), channel, query)
		return Dispatch(r, app, ctx, addr, chainID, msg, ak, bk, nil)
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
