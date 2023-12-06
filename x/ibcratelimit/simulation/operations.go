package simulation

import (
	"fmt"
	"math/rand"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	simappparams "github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/x/ibcratelimit"
	"github.com/provenance-io/provenance/x/ibcratelimit/keeper"
)

// Simulation operation weights constants
const (
	//nolint:gosec // not credentials
	OpWeightMsgUpdateParams = "op_weight_msg_update_params"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgUpdateParams int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgUpdateParams, &weightMsgUpdateParams, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateParams = simappparams.DefaultWeightGovUpdateParams
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgUpdateParams,
			SimulateMsgGovUpdateParams(k, ak, bk),
		),
	}
}

// SimulateMsgGovUpdateParams sends a MsgUpdateParams.
func SimulateMsgGovUpdateParams(_ keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		raccs, err := RandomAccs(r, accs, uint64(len(accs)))
		if err != nil {
			return simtypes.NoOpMsg(sdk.MsgTypeURL(&ibcratelimit.MsgGovUpdateParamsRequest{}), sdk.MsgTypeURL(&ibcratelimit.MsgGovUpdateParamsRequest{}), err.Error()), nil, nil
		}

		// 50% chance to be from the module's authority
		from := raccs[0]
		to := raccs[1]

		msg := ibcratelimit.NewMsgGovUpdateParamsRequest(from.Address.String(), to.Address.String())

		return Dispatch(r, app, ctx, from, chainID, msg, ak, bk, nil)
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
		Amount: sdkmath.NewInt(1_000_000_000_000_000),
	}))
	if err != nil {
		return simtypes.NoOpMsg(sdk.MsgTypeURL(msg), sdk.MsgTypeURL(msg), "unable to fund account"), nil, err
	}
	txGen := simappparams.MakeTestEncodingConfig().TxConfig
	tx, err := simtestutil.GenSignedMockTx(
		r,
		txGen,
		[]sdk.Msg{msg},
		fees,
		simtestutil.DefaultGenTxGas,
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
