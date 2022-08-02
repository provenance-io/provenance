package simulation

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simappparams "github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/x/reward/keeper"
	"github.com/provenance-io/provenance/x/reward/types"
	"math/rand"
	"time"
)

// Simulation operation weights constants
const (
	//nolint:gosec // not credentials
	OpWeightSubmitCreateRewardsProposal = "op_weight_submit_create_rewards_proposal"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgAddRewardsProgram int
	)

	appParams.GetOrGenerate(cdc, OpWeightSubmitCreateRewardsProposal, &weightMsgAddRewardsProgram, nil,
		func(_ *rand.Rand) {
			weightMsgAddRewardsProgram = simappparams.DefaultWeightSubmitCreateRewards
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgAddRewardsProgram,
			SimulateMsgCreateRewardsProgram(k, ak, bk),
		),
	}
}

func SimulateMsgCreateRewardsProgram(k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		var totalRewardsPool = GenTotalRewardsPool(r)
		var maxRewardsByAddress = GenMaxRewardsByAddress(r)
		var maxActions = MaxActionsFn(r)

		var minActions = MinActionsFn(r)

		minDelegation := sdk.NewInt64Coin(pioconfig.DefaultBondDenom, int64(minActions))

		now := time.Now()
		claimPeriods := uint64(r.Intn(100))
		maxRolloverPeriods := uint64(r.Intn(10))
		expireClaimPeriods := uint64(r.Intn(100))

		msg := types.NewMsgCreateRewardProgramRequest("title",
			"description",
			simAccount.Address.String(),
			totalRewardsPool,
			maxRewardsByAddress,
			now.Add(5*time.Second),
			claimPeriods,
			1,
			maxRolloverPeriods,
			expireClaimPeriods,
			[]types.QualifyingAction{
				{
					Type: &types.QualifyingAction_Vote{
						Vote: &types.ActionVote{
							MinimumActions:          minActions,
							MaximumActions:          maxActions,
							MinimumDelegationAmount: minDelegation,
						},
					},
				},
			})
		return Dispatch(r, app, ctx, ak, bk, simAccount, chainID, msg, nil)
	}
}

// Dispatch sends an operation to the chain using a given account/funds on account for fees.  Failures on the server side
// are handled as no-op msg operations with the error string as the status/response.
func Dispatch(
	r *rand.Rand,
	app *baseapp.BaseApp,
	ctx sdk.Context,
	ak authkeeper.AccountKeeperI,
	bk bankkeeper.Keeper,
	from simtypes.Account,
	chainID string,
	msg sdk.Msg,
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
		return simtypes.NoOpMsg(types.ModuleName, fmt.Sprintf("%T", msg), "unable to generate fees"), nil, err
	}
	err = simapp.FundAccount(bk, ctx, account.GetAddress(), sdk.NewCoins(sdk.Coin{
		Denom:  pioconfig.DefaultBondDenom,
		Amount: sdk.NewInt(1_000_000_000_000_000),
	}))
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

	return simtypes.NewOperationMsg(msg, true, "", &codec.ProtoCodec{}), futures, nil
}
