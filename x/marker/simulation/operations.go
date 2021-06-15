package simulation

import (
	"math/rand"
	"regexp"
	"strconv"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simappparams "github.com/provenance-io/provenance/app/params"

	keeper "github.com/provenance-io/provenance/x/marker/keeper"
	types "github.com/provenance-io/provenance/x/marker/types"
)

// Simulation operation weights constants
const (
	OpWeightMsgAddMarker    = "op_weight_msg_add_marker"
	OpWeightMsgDeleteMarker = "op_weight_msg_delete_marker"
)

/*
Finalize
Activate
Cancel
Delete


AddAccess
DeleteAccess

Withdraw
AddMarker
	Finalize
	Activate
	Cancel
	Delete

Mint
Burn
Transfer

SetDenomMetadata
*/

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONMarshaler, k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper,
) simulation.WeightedOperations {
	var (
		weightMsgBindName int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgAddMarker, &weightMsgBindName, nil,
		func(_ *rand.Rand) {
			weightMsgBindName = simappparams.DefaultWeightMsgAddMarker
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgBindName,
			SimulateMsgAddMarker(k, ak, bk),
		),
	}
}

// SimulateMsgAddMarker will bind a NAME under an existing name using a 40% probability of restricting it.
func SimulateMsgAddMarker(k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.ViewKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		mgrAccount, _ := simtypes.RandomAcc(r, accs)

		msg := types.NewMsgAddMarkerRequest(
			randomUnrestrictedDenom(r, k.GetUnrestrictedDenomRegex(ctx)),
			sdk.NewInt(r.Int63()),
			simAccount.Address,
			mgrAccount.Address,
			types.MarkerType(r.Intn(1)+1), // coin or restricted_coin
			r.Intn(1) > 0,                 // fixed supply
			r.Intn(1) > 0,                 // allow gov
		)

		return Dispatch(r, app, ctx, ak, bk, simAccount, chainID, msg)
	}
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

func randomUnrestrictedDenom(r *rand.Rand, unrestrictedDenomExp string) string {
	exp := regexp.MustCompile(`\{(\d+),(\d+)\}`)
	matches := exp.FindStringSubmatch(unrestrictedDenomExp)
	if len(matches) != 3 {
		panic("expected two number as range expression in unrestricted denom expression")
	}
	min, _ := strconv.ParseInt(matches[1], 10, 32)
	max, _ := strconv.ParseInt(matches[2], 10, 32)

	return simtypes.RandStringOfLength(r, int(r.Int63n(max-min)+min))
}

func randomAccessGrants(r *rand.Rand, accs []simtypes.Account) (grants []types.AccessGrant) {
	// select random number of accounts ...
	count := r.Intn(len(accs))
	for i := 0; i < count; i++ {
		simaccount, _ := simtypes.RandomAcc(r, accs)
		// for each of the accounts selected .. add a random set of permissions.
		grants = append(grants, *types.NewAccessGrant(simaccount.Address, types.AccessListByNames("mint, burn")))
	}
	// mint, burn, deposit, withdraw, delete, admin, transfer
	return
}

func randomMarker(r *rand.Rand, ctx sdk.Context, k keeper.Keeper) types.MarkerAccountI {
	var markers []types.MarkerAccountI
	k.IterateMarkers(ctx, func(marker types.MarkerAccountI) (stop bool) {
		markers = append(markers, marker)
		return false
	})
	idx := r.Intn(len(markers))
	return markers[idx]
}
