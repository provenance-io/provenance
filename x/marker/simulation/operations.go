package simulation

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"

	"github.com/cosmos/cosmos-sdk/simapp"

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
	//nolint:gosec // not credentials
	OpWeightMsgAddMarker = "op_weight_msg_add_marker"
	//nolint:gosec // not credentials
	OpWeightMsgChangeStatus = "op_weight_msg_change_status"
	//nolint:gosec // not credentials
	OpWeightMsgAddAccess = "op_weight_msg_add_access"
)

/*

AddAccess
DeleteAccess

Withdraw

Mint
Burn
Transfer

SetDenomMetadata
*/

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgAddMarker    int
		weightMsgChangeStatus int
		weightMsgAddAccess    int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgAddMarker, &weightMsgAddMarker, nil,
		func(_ *rand.Rand) {
			weightMsgAddMarker = simappparams.DefaultWeightMsgAddMarker
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgChangeStatus, &weightMsgChangeStatus, nil,
		func(_ *rand.Rand) {
			weightMsgChangeStatus = simappparams.DefaultWeightMsgChangeStatus
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgAddAccess, &weightMsgAddAccess, nil,
		func(_ *rand.Rand) {
			weightMsgAddAccess = simappparams.DefaultWeightMsgAddAccess
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgAddMarker,
			SimulateMsgAddMarker(k, ak, bk),
		),
		simulation.NewWeightedOperation(
			weightMsgChangeStatus,
			SimulateMsgChangeStatus(k, ak, bk),
		),
		simulation.NewWeightedOperation(
			weightMsgAddAccess,
			SimulateMsgAddAccess(k, ak, bk),
		),
	}
}

// SimulateMsgAddMarker will bind a NAME under an existing name using a 40% probability of restricting it.
func SimulateMsgAddMarker(k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		mgrAccount, _ := simtypes.RandomAcc(r, accs)
		denom := randomUnrestrictedDenom(r, k.GetUnrestrictedDenomRegex(ctx))
		msg := types.NewMsgAddMarkerRequest(
			denom,
			sdk.NewIntFromUint64(randomUint64(r, k.GetMaxTotalSupply(ctx))),
			simAccount.Address,
			mgrAccount.Address,
			types.MarkerType(r.Intn(2)+1), // coin or restricted_coin
			r.Intn(2) > 0,                 // fixed supply
			r.Intn(2) > 0,                 // allow gov
		)

		return Dispatch(r, app, ctx, ak, bk, simAccount, chainID, msg, nil)
	}
}

func SimulateMsgChangeStatus(k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		m := randomMarker(r, ctx, k)
		if m == nil {
			return simtypes.NoOpMsg(types.ModuleName, "ChangeStatus", "unable to get marker for status change"), nil, nil
		}
		var simAccount simtypes.Account
		var found bool
		var msg sdk.Msg
		switch m.GetStatus() {
		// 50% chance of (re-)issuing a finalize or a 50/50 chance to cancel/activate.
		case types.StatusProposed, types.StatusFinalized:
			if r.Intn(10) < 5 {
				msg = types.NewMsgFinalizeRequest(m.GetDenom(), m.GetManager())
			} else {
				if r.Intn(10) < 5 {
					msg = types.NewMsgCancelRequest(m.GetDenom(), simAccount.Address)
				} else {
					msg = types.NewMsgActivateRequest(m.GetDenom(), m.GetManager())
				}
			}
			simAccount, found = simtypes.FindAccount(accs, m.GetManager())
			if !found {
				return simtypes.NoOpMsg(types.ModuleName, fmt.Sprintf("%T", msg), "manager account does not exist"), nil, nil
			}
		case types.StatusActive:
			accounts := m.AddressListForPermission(types.Access_Delete)
			if len(accounts) < 1 {
				return simtypes.NoOpMsg(types.ModuleName, types.TypeCancelRequest, "no account has cancel access"), nil, nil
			}
			simAccount, _ = simtypes.FindAccount(accs, accounts[0])
			msg = types.NewMsgCancelRequest(m.GetDenom(), simAccount.Address)
		case types.StatusCancelled:
			accounts := m.AddressListForPermission(types.Access_Delete)
			if len(accounts) < 1 {
				return simtypes.NoOpMsg(types.ModuleName, types.TypeDeleteRequest, "no account has delete access"), nil, nil
			}
			simAccount, _ = simtypes.FindAccount(accs, accounts[0])
			msg = types.NewMsgDeleteRequest(m.GetDenom(), simAccount.Address)
		}

		return Dispatch(r, app, ctx, ak, bk, simAccount, chainID, msg, nil)
	}
}

func SimulateMsgAddAccess(k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		m := randomMarker(r, ctx, k)
		if m == nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeAddAccessRequest, "unable to get marker for access change"), nil, nil
		}
		if !m.GetManager().Equals(sdk.AccAddress{}) {
			simAccount, _ = simtypes.FindAccount(accs, m.GetManager())
		}
		grants := randomAccessGrants(r, accs, 100)
		msg := types.NewMsgAddAccessRequest(m.GetDenom(), simAccount.Address, grants[0])
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
	// fund account with nhash for additional fees, if the account exists (100m stake)
	if sdk.MsgTypeURL(msg) == "/provenance.marker.v1.MsgAddMarkerRequest" && ak.GetAccount(ctx, account.GetAddress()) != nil {
		err = simapp.FundAccount(bk, ctx, account.GetAddress(), sdk.NewCoins(sdk.Coin{
			Denom:  "stake",
			Amount: sdk.NewInt(1_000_000_000_000_000),
		}))
		fees = fees.Add(sdk.Coin{
			Denom:  "stake",
			Amount: sdk.NewInt(100_000_000_000_000),
		})
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, fmt.Sprintf("%T", msg), "unable to fund account with additional fee"), nil, err
		}
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

	return simtypes.NewOperationMsg(msg, true, "", &codec.ProtoCodec{}), futures, nil
}

func randomUnrestrictedDenom(r *rand.Rand, unrestrictedDenomExp string) string {
	exp := regexp.MustCompile(`\{(\d+),(\d+)\}`)
	matches := exp.FindStringSubmatch(unrestrictedDenomExp)
	if len(matches) != 3 {
		panic("expected two number as range expression in unrestricted denom expression")
	}
	min, _ := strconv.ParseInt(matches[1], 10, 32)
	max, _ := strconv.ParseInt(matches[2], 10, 32)

	return simtypes.RandStringOfLength(r, int(randomInt63(r, max-min)+min))
}

// build
func randomAccessGrants(r *rand.Rand, accs []simtypes.Account, limit int) (grants []types.AccessGrant) {
	// select random number of accounts ...
	for i := 0; i < len(accs); i++ {
		if r.Intn(10) < 3 {
			continue
		}
		if i >= limit {
			return
		}
		// for each of the accounts selected .. add a random set of permissions.
		grants = append(grants, *types.NewAccessGrant(accs[i].Address, randomAccessTypes(r)))
	}
	return
}

// builds a list of access rights with a 50:50 chance of including each one
func randomAccessTypes(r *rand.Rand) (result []types.Access) {
	access := []string{"mint", "burn", "deposit", "withdraw", "delete", "admin"}
	for i := 0; i < len(access); i++ {
		if r.Intn(10) < 4 {
			result = append(result, types.AccessByName(access[i]))
		}
	}
	return
}

func randomMarker(r *rand.Rand, ctx sdk.Context, k keeper.Keeper) types.MarkerAccountI {
	var markers []types.MarkerAccountI
	k.IterateMarkers(ctx, func(marker types.MarkerAccountI) (stop bool) {
		markers = append(markers, marker)
		return false
	})
	if len(markers) == 0 {
		return nil
	}
	idx := r.Intn(len(markers))
	return markers[idx]
}

func randomInt63(r *rand.Rand, max int64) (result int64) {
	if max == 0 {
		return 0
	}
	return r.Int63n(max)
}

func randomUint64(r *rand.Rand, max uint64) (result uint64) {
	if max == 0 {
		return 0
	}
	for {
		result = r.Uint64()
		if result < max {
			break
		}
	}
	return
}
