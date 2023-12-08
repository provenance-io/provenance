package simulation

// DONTCOVER

import (
	"math/rand"
	"time"

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
	"github.com/provenance-io/provenance/x/reward/keeper"
	"github.com/provenance-io/provenance/x/reward/types"
)

// Simulation operation weights constants
const (
	OpWeightSubmitCreateRewardsProposal = "op_weight_submit_create_rewards_proposal"
	OpWeightEndRewardsProposal          = "op_weight_submit_end_reward_proposal"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgAddRewardsProgram int
		weightMsgEndRewardProgram  int
	)

	appParams.GetOrGenerate(OpWeightSubmitCreateRewardsProposal, &weightMsgAddRewardsProgram, nil,
		func(_ *rand.Rand) {
			weightMsgAddRewardsProgram = simappparams.DefaultWeightSubmitCreateRewards
		},
	)
	appParams.GetOrGenerate(OpWeightEndRewardsProposal, &weightMsgEndRewardProgram, nil,
		func(_ *rand.Rand) {
			weightMsgEndRewardProgram = simappparams.DefaultWeightSubmitEndRewards
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgAddRewardsProgram,
			SimulateMsgCreateRewardsProgram(k, ak, bk),
		),
		simulation.NewWeightedOperation(
			weightMsgEndRewardProgram,
			SimulateMsgEndRewardsProgram(k, ak, bk),
		),
	}
}

// SimulateMsgCreateRewardsProgram sends of a MsgCreateRewardProgramRequest.
func SimulateMsgCreateRewardsProgram(_ keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		var totalRewardsPool = GenTotalRewardsPool(r)
		var maxRewardsByAddress = GenMaxRewardsByAddress(r)
		var maxActions = MaxActionsFn(r)

		var minActions = MinActionsFn(r)

		minDelegation := sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().BondDenom, int64(minActions))

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
		return simtypes.NoOpMsg(sdk.MsgTypeURL(msg), sdk.MsgTypeURL(msg), "unable to generate fees"), nil, err
	}
	err = testutil.FundAccount(ctx, bk, account.GetAddress(), sdk.NewCoins(sdk.Coin{
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

	return simtypes.NewOperationMsg(msg, true, ""), futures, nil
}

// SimulateMsgEndRewardsProgram sends a MsgEndRewardProgramRequest for a random existing reward program.
func SimulateMsgEndRewardsProgram(k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		rewardProgram := randomRewardProgram(r, ctx, k)
		if rewardProgram == nil {
			return simtypes.NoOpMsg(sdk.MsgTypeURL(&types.MsgEndRewardProgramRequest{}), sdk.MsgTypeURL(&types.MsgEndRewardProgramRequest{}), "unable to find a valid reward program"), nil, nil
		}
		var simAccount simtypes.Account
		var found bool
		addr, err := sdk.AccAddressFromBech32(rewardProgram.DistributeFromAddress)
		if err != nil {
			// should just noit be possible and panic on the test
			panic(err)
		}
		simAccount, found = simtypes.FindAccount(accs, addr)
		if !found {
			return simtypes.NoOpMsg(sdk.MsgTypeURL(&types.MsgEndRewardProgramRequest{}), sdk.MsgTypeURL(&types.MsgEndRewardProgramRequest{}), "creator of rewards program account does not exist"), nil, nil
		}
		msg := types.NewMsgEndRewardProgramRequest(rewardProgram.Id, rewardProgram.DistributeFromAddress)
		return Dispatch(r, app, ctx, ak, bk, simAccount, chainID, msg, nil)
	}
}

func randomRewardProgram(r *rand.Rand, ctx sdk.Context, k keeper.Keeper) *types.RewardProgram {
	var rewardPrograms []types.RewardProgram
	err := k.IterateRewardPrograms(ctx, func(rewardProgram types.RewardProgram) (stop bool, err error) {
		rewardPrograms = append(rewardPrograms, rewardProgram)
		return false, nil
	})
	if err != nil {
		// sim tests should fail if iterator errors
		panic(err)
	}
	if len(rewardPrograms) == 0 {
		return nil
	}
	idx := r.Intn(len(rewardPrograms))
	return &rewardPrograms[idx]
}
