package simulation

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	simappparams "github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/x/marker/keeper"
	"github.com/provenance-io/provenance/x/marker/types"
)

// Simulation operation weights constants
const (
	//nolint:gosec // not credentials
	OpWeightMsgAddMarker = "op_weight_msg_add_marker"
	//nolint:gosec // not credentials
	OpWeightMsgChangeStatus = "op_weight_msg_change_status"
	//nolint:gosec // not credentials
	OpWeightMsgAddAccess = "op_weight_msg_add_access"
	//nolint:gosec // not credentials
	OpWeightMsgAddActivateFinalizeMarker = "op_weight_msg_add_finalize_activate_marker"
	//nolint:gosec // not credentials
	OpWeightMsgAddMarkerProposal = "op_weight_msg_add_marker_proposal"
	//nolint:gosec // not credentials
	OpWeightMsgSetAccountData = "op_weight_msg_set_account_data"
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
	appParams simtypes.AppParams, cdc codec.JSONCodec, protoCodec *codec.ProtoCodec,
	k keeper.Keeper, ak authkeeper.AccountKeeper, bk bankkeeper.Keeper, gk types.GovKeeper, attrk types.AttrKeeper,
) simulation.WeightedOperations {
	args := &WeightedOpsArgs{
		AppParams:  appParams,
		JSONCodec:  cdc,
		ProtoCodec: protoCodec,
		AK:         ak,
		BK:         bk,
		GK:         gk,
		AttrK:      attrk,
	}

	var (
		weightMsgAddMarker                 int
		weightMsgChangeStatus              int
		weightMsgAddAccess                 int
		weightMsgAddFinalizeActivateMarker int
		weightMsgAddMarkerProposal         int
		weightMsgSetAccountData            int
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

	appParams.GetOrGenerate(cdc, OpWeightMsgAddActivateFinalizeMarker, &weightMsgAddFinalizeActivateMarker, nil,
		func(_ *rand.Rand) {
			weightMsgAddFinalizeActivateMarker = simappparams.DefaultWeightMsgAddFinalizeActivateMarker
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgAddMarkerProposal, &weightMsgAddMarkerProposal, nil,
		func(_ *rand.Rand) {
			weightMsgAddMarkerProposal = simappparams.DefaultWeightMsgAddMarkerProposal
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgSetAccountData, &weightMsgSetAccountData, nil,
		func(_ *rand.Rand) {
			weightMsgSetAccountData = simappparams.DefaultWeightMsgSetAccountData
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
		simulation.NewWeightedOperation(
			weightMsgAddFinalizeActivateMarker,
			SimulateMsgAddFinalizeActivateMarker(k, ak, bk),
		),
		simulation.NewWeightedOperation(
			weightMsgAddMarkerProposal,
			SimulateMsgAddMarkerProposal(k, args),
		),
		simulation.NewWeightedOperation(
			weightMsgSetAccountData,
			SimulateMsgSetAccountData(k, args),
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
			r.Intn(2) > 0,                 // allow forced transfer
			[]string{},
			0,
			0,
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
			simAccount, found = randomAccWithAccess(r, m, accs, types.Access_Delete)
			if !found {
				return simtypes.NoOpMsg(sdk.MsgTypeURL(&types.MsgCancelRequest{}), sdk.MsgTypeURL(&types.MsgCancelRequest{}), "no account has cancel access"), nil, nil
			}
			msg = types.NewMsgCancelRequest(m.GetDenom(), simAccount.Address)
		case types.StatusCancelled:
			simAccount, found = randomAccWithAccess(r, m, accs, types.Access_Delete)
			if !found {
				return simtypes.NoOpMsg(sdk.MsgTypeURL(&types.MsgDeleteRequest{}), sdk.MsgTypeURL(&types.MsgDeleteRequest{}), "no account has delete access"), nil, nil
			}
			msg = types.NewMsgDeleteRequest(m.GetDenom(), simAccount.Address)
		case types.StatusDestroyed:
			return simtypes.NoOpMsg(types.ModuleName, "ChangeStatus", "marker status is destroyed"), nil, nil
		default:
			return simtypes.NoOpMsg("marker", "", "unknown marker status"), nil, fmt.Errorf("unknown marker status: %#v", m)
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
			return simtypes.NoOpMsg(sdk.MsgTypeURL(&types.MsgAddAccessRequest{}), sdk.MsgTypeURL(&types.MsgAddAccessRequest{}), "unable to get marker for access change"), nil, nil
		}
		if !m.GetManager().Equals(sdk.AccAddress{}) {
			simAccount, _ = simtypes.FindAccount(accs, m.GetManager())
		}
		grants := randomAccessGrants(r, accs, 100, m.GetMarkerType())
		msg := types.NewMsgAddAccessRequest(m.GetDenom(), simAccount.Address, grants[0])
		return Dispatch(r, app, ctx, ak, bk, simAccount, chainID, msg, nil)
	}
}

// SimulateMsgAddFinalizeActivateMarker will bind a NAME under an existing name using a 40% probability of restricting it.
func SimulateMsgAddFinalizeActivateMarker(k keeper.Keeper, ak authkeeper.AccountKeeperI, bk bankkeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		mgrAccount, _ := simtypes.RandomAcc(r, accs)
		denom := randomUnrestrictedDenom(r, k.GetUnrestrictedDenomRegex(ctx))
		markerType := types.MarkerType(r.Intn(2) + 1) // coin or restricted_coin
		// random access grants
		grants := randomAccessGrants(r, accs, 100, markerType)
		msg := types.NewMsgAddFinalizeActivateMarkerRequest(
			denom,
			sdk.NewIntFromUint64(randomUint64(r, k.GetMaxTotalSupply(ctx))),
			simAccount.Address,
			mgrAccount.Address,
			markerType,
			r.Intn(2) > 0, // fixed supply
			r.Intn(2) > 0, // allow gov
			r.Intn(2) > 0, // allow forced transfer
			[]string{},
			grants,
			0,
			0,
		)

		if msg.MarkerType != types.MarkerType_RestrictedCoin {
			msg.AllowForcedTransfer = false
		}

		return Dispatch(r, app, ctx, ak, bk, simAccount, chainID, msg, nil)
	}
}

func SimulateMsgAddMarkerProposal(k keeper.Keeper, args *WeightedOpsArgs) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		denom := randomUnrestrictedDenom(r, k.GetUnrestrictedDenomRegex(ctx))

		markerStatus := types.MarkerStatus(r.Intn(3) + 1)
		markerType := types.MarkerType(r.Intn(2) + 1)
		msg := &types.MsgAddMarkerRequest{
			Amount: sdk.Coin{
				Denom:  denom,
				Amount: sdk.NewIntFromUint64(randomUint64(r, k.GetMaxTotalSupply(ctx))),
			},
			Manager:                simAccount.Address.String(),
			FromAddress:            k.GetAuthority(),
			Status:                 markerStatus,
			MarkerType:             markerType,
			AccessList:             []types.AccessGrant{{Address: simAccount.Address.String(), Permissions: randomAccessTypes(r, markerType)}},
			SupplyFixed:            r.Intn(2) > 0,
			AllowGovernanceControl: true,
			AllowForcedTransfer:    r.Intn(2) > 0,
			RequiredAttributes:     nil,
		}
		if msg.Status == types.StatusActive {
			msg.Manager = ""
		}
		if msg.MarkerType != types.MarkerType_RestrictedCoin {
			msg.AllowForcedTransfer = false
		}

		// Get the governance min deposit needed
		govMinDep := sdk.NewCoins(args.GK.GetDepositParams(ctx).MinDeposit...)

		sender, _ := simtypes.RandomAcc(r, accs)

		msgArgs := &SendGovMsgArgs{
			WeightedOpsArgs: *args,
			R:               r,
			App:             app,
			Ctx:             ctx,
			Accs:            accs,
			ChainID:         chainID,
			Sender:          sender,
			Msg:             msg,
			Deposit:         govMinDep,
			Comment:         "marker",
		}

		skip, opMsg, err := SendGovMsg(msgArgs)

		if skip || err != nil {
			return opMsg, nil, err
		}

		proposalID, err := args.GK.GetProposalID(ctx)
		if err != nil {
			return opMsg, nil, err
		}

		votingPeriod := args.GK.GetVotingParams(ctx).VotingPeriod
		fops := make([]simtypes.FutureOperation, len(accs))
		for i, acct := range accs {
			whenVote := ctx.BlockHeader().Time.Add(time.Duration(r.Int63n(int64(votingPeriod.Seconds()))) * time.Second)
			fops[i] = simtypes.FutureOperation{
				BlockTime: whenVote,
				Op:        OperationMsgVote(args, acct, proposalID, govtypes.OptionYes, msgArgs.Comment),
			}
		}

		return opMsg, fops, nil
	}
}

func SimulateMsgSetAccountData(k keeper.Keeper, args *WeightedOpsArgs) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msg := &types.MsgSetAccountDataRequest{}

		marker, signer := randomMarkerWithAccessSigner(r, ctx, k, accs, types.Access_Deposit)
		if marker == nil {
			return simtypes.NoOpMsg(sdk.MsgTypeURL(msg), sdk.MsgTypeURL(msg), "unable to find marker with a deposit signer"), nil, nil
		}

		msg.Denom = marker.GetDenom()
		msg.Signer = signer.Address.String()

		// 1 in 10 chance that the value stays "".
		// 9 in 10 chance that it will be between 1 and MaxValueLen characters.
		if r.Intn(10) != 0 {
			maxLen := uint(args.AttrK.GetMaxValueLength(ctx))
			if maxLen > 500 {
				maxLen = 500
			}
			strLen := r.Intn(int(maxLen)) + 1
			msg.Value = simtypes.RandStringOfLength(r, strLen)
		}

		return Dispatch(r, app, ctx, args.AK, args.BK, signer, chainID, msg, nil)
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
	// fund account with nhash for additional fees, if the account exists (100m stake)
	if sdk.MsgTypeURL(msg) == "/provenance.marker.v1.MsgAddMarkerRequest" && ak.GetAccount(ctx, account.GetAddress()) != nil {
		err = testutil.FundAccount(bk, ctx, account.GetAddress(), sdk.NewCoins(sdk.Coin{
			Denom:  "stake",
			Amount: sdk.NewInt(100_000_000_000_000),
		}))
		if err != nil {
			return simtypes.NoOpMsg(sdk.MsgTypeURL(msg), sdk.MsgTypeURL(msg), "unable to fund account with additional fee"), nil, err
		}
		fees = fees.Add(sdk.Coin{
			Denom:  "stake",
			Amount: sdk.NewInt(100_000_000_000_000),
		})
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

// randomAccessGrants generates random access grants for randomly selected accounts.
// Each account has a 30% chance of being chosen with a max of limit.
func randomAccessGrants(r *rand.Rand, accs []simtypes.Account, limit int, markerType types.MarkerType) (grants []types.AccessGrant) {
	// select random number of accounts ...
	for i := 0; i < len(accs); i++ {
		if r.Intn(10) < 3 {
			continue
		}
		// for each of the accounts selected, add a random set of permissions.
		grants = append(grants, *types.NewAccessGrant(accs[i].Address, randomAccessTypes(r, markerType)))
		if len(grants) >= limit {
			return
		}
	}
	return
}

// randomAccessTypes builds a list of access rights with a 40% chance of including each one
func randomAccessTypes(r *rand.Rand, markerType types.MarkerType) (result []types.Access) {
	access := []string{"mint", "burn", "deposit", "withdraw", "delete", "admin"}
	if markerType == types.MarkerType_RestrictedCoin {
		access = append(access, "transfer")
	}
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

func randomMarkerWithAccessSigner(r *rand.Rand, ctx sdk.Context, k keeper.Keeper, accs []simtypes.Account, access types.Access) (types.MarkerAccountI, simtypes.Account) {
	var markers []types.MarkerAccountI
	k.IterateMarkers(ctx, func(marker types.MarkerAccountI) (stop bool) {
		markers = append(markers, marker)
		return false
	})
	if len(markers) == 0 {
		return nil, simtypes.Account{}
	}

	r.Shuffle(len(markers), func(i, j int) {
		markers[i], markers[j] = markers[j], markers[i]
	})

	for _, marker := range markers {
		acc, found := randomAccWithAccess(r, marker, accs, access)
		if found {
			return marker, acc
		}
	}

	return nil, simtypes.Account{}
}

func randomAccWithAccess(r *rand.Rand, marker types.MarkerAccountI, accs []simtypes.Account, access types.Access) (simtypes.Account, bool) {
	addrs := marker.AddressListForPermission(access)

	if len(addrs) == 0 {
		return simtypes.Account{}, false
	}

	r.Shuffle(len(addrs), func(i, j int) {
		addrs[i], addrs[j] = addrs[j], addrs[i]
	})

	for _, addr := range addrs {
		acc, found := simtypes.FindAccount(accs, addr)
		if found {
			return acc, true
		}
	}

	return simtypes.Account{}, false
}

func randomInt63(r *rand.Rand, max int64) (result int64) {
	if max == 0 {
		return 0
	}
	return r.Int63n(max)
}

// randomUint64 gets a random uint64 between 0 and max (inclusive): [0, max].
func randomUint64(r *rand.Rand, max uint64) uint64 {
	if max == 0 {
		return 0
	}
	// Max int64 is 9,223,372,036,854,775,807.
	// If the provided max is less than that, we'll just use r.Int63n.
	// Otherwise, we'll use an infinite loop until r.Uint64() returns something small enough.
	// This way, if the max is small (e.g. 2), we don't sit in this loop forever.
	if max < 9_223_372_036_854_775_807 {
		// Using max+1 because we want max to be possible.
		return uint64(r.Int63n(int64(max + 1)))
	}
	// Not using modulo here because that increases the chances of the low numbers and reduces the chances of bigger ones.
	result := r.Uint64()
	for result > max {
		result = r.Uint64()
	}
	return result
}

// WeightedOpsArgs holds all the args provided to WeightedOperations so that they can be passed on later more easily.
type WeightedOpsArgs struct {
	AppParams  simtypes.AppParams
	JSONCodec  codec.JSONCodec
	ProtoCodec *codec.ProtoCodec
	AK         authkeeper.AccountKeeper
	BK         bankkeeper.Keeper
	GK         types.GovKeeper
	AttrK      types.AttrKeeper
}

// SendGovMsgArgs holds all the args available and needed for sending a gov msg.
type SendGovMsgArgs struct {
	WeightedOpsArgs

	R       *rand.Rand
	App     *baseapp.BaseApp
	Ctx     sdk.Context
	Accs    []simtypes.Account
	ChainID string

	Sender  simtypes.Account
	Msg     sdk.Msg
	Deposit sdk.Coins
	Comment string
}

// SendGovMsg sends a msg as a gov prop.
// It returns whether to skip the rest, an operation message, and any error encountered.
func SendGovMsg(args *SendGovMsgArgs) (bool, simtypes.OperationMsg, error) {
	msgType := sdk.MsgTypeURL(args.Msg)

	spendableCoins := args.BK.SpendableCoins(args.Ctx, args.Sender.Address)
	if spendableCoins.Empty() {
		return true, simtypes.NoOpMsg(types.ModuleName, msgType, "sender has no spendable coins"), nil
	}

	_, hasNeg := spendableCoins.SafeSub(args.Deposit...)
	if hasNeg {
		return true, simtypes.NoOpMsg(types.ModuleName, msgType, "sender has insufficient balance to cover deposit"), nil
	}

	msgAny, err := codectypes.NewAnyWithValue(args.Msg)
	if err != nil {
		return true, simtypes.NoOpMsg(types.ModuleName, msgType, "wrapping MsgAddMarkerProposalRequest as Any"), err
	}

	govMsg := &govtypes.MsgSubmitProposal{
		Messages:       []*codectypes.Any{msgAny},
		InitialDeposit: args.Deposit,
		Proposer:       args.Sender.Address.String(),
		Metadata:       "",
	}

	txCtx := simulation.OperationInput{
		R:               args.R,
		App:             args.App,
		TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
		Cdc:             args.ProtoCodec,
		Msg:             govMsg,
		MsgType:         govMsg.Type(),
		CoinsSpentInMsg: govMsg.InitialDeposit,
		Context:         args.Ctx,
		SimAccount:      args.Sender,
		AccountKeeper:   args.AK,
		Bankkeeper:      args.BK,
		ModuleName:      types.ModuleName,
	}

	opMsg, _, err := simulation.GenAndDeliverTxWithRandFees(txCtx)
	if opMsg.Comment == "" {
		opMsg.Comment = args.Comment
	}

	return err != nil, opMsg, err
}

// OperationMsgVote returns an operation that casts a yes vote on a gov prop from an account.
func OperationMsgVote(args *WeightedOpsArgs, voter simtypes.Account, govPropID uint64, vote govtypes.VoteOption, comment string) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msg := govtypes.NewMsgVote(voter.Address, govPropID, vote, "")

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             args.ProtoCodec,
			Msg:             msg,
			MsgType:         msg.Type(),
			CoinsSpentInMsg: sdk.Coins{},
			Context:         ctx,
			SimAccount:      voter,
			AccountKeeper:   args.AK,
			Bankkeeper:      args.BK,
			ModuleName:      types.ModuleName,
		}

		opMsg, fops, err := simulation.GenAndDeliverTxWithRandFees(txCtx)
		if opMsg.Comment == "" {
			opMsg.Comment = comment
		}

		return opMsg, fops, err
	}
}
