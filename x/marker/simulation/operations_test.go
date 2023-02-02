package simulation_test

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/provenance-io/provenance/app"
	simappparams "github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/x/marker/simulation"
	types "github.com/provenance-io/provenance/x/marker/types"
)

type SimTestSuite struct {
	suite.Suite

	ctx sdk.Context
	app *app.App
}

func (suite *SimTestSuite) SetupTest() {
	suite.app = app.Setup(suite.T())
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{})
}

func (suite *SimTestSuite) TestWeightedOperations() {
	cdc := suite.app.AppCodec()
	appParams := make(simtypes.AppParams)

	weightedOps := simulation.WeightedOperations(appParams, cdc, codec.NewProtoCodec(suite.app.InterfaceRegistry()), suite.app.MarkerKeeper,
		suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.GovKeeper,
	)

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accs := suite.getTestingAccounts(r, 3)

	// Note: r is now passed around and used in several places, including in SDK functions.
	//	Since we're seeding it, these tests are deterministic. However, if there are changes
	//	made in the SDK or to the operations, these outcomes can change. To further confuse
	//	things, the operation name is sometimes taken from msg.Type(), and sometimes from
	//	fmt.Sprintf("%T", msg), and sometimes hard-coded. The .Type() function is no longer
	//	part of the Msg interface (though it is part of LegacyMsg). But depending on how the
	//	randomness plays out, it can be either of those. If one of these starts failing on
	//	the operation name, and the actual value is one of the other possibilities for that
	//	operation, it's probably just do to a change in the number of times r is used before
	//	getting to that operation.

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		// Possible names: types.TypeAddMarkerRequest, fmt.Sprintf("%T", &types.MsgAddMarkerRequest{})
		{simappparams.DefaultWeightMsgAddMarker, sdk.MsgTypeURL(&types.MsgAddMarkerRequest{}), sdk.MsgTypeURL(&types.MsgAddMarkerRequest{})},
		// Possible names: "ChangeStatus",
		//	types.TypeActivateRequest, fmt.Sprintf("%T", &types.MsgActivateRequest{}),
		//	types.TypeFinalizeRequest, fmt.Sprintf("%T", &types.MsgFinalizeRequest{}),
		//	types.TypeCancelRequest, fmt.Sprintf("%T", &types.MsgCancelRequest{}),
		//	types.TypeDeleteRequest, fmt.Sprintf("%T", &types.MsgDeleteRequest{}),
		{simappparams.DefaultWeightMsgChangeStatus, sdk.MsgTypeURL(&types.MsgActivateRequest{}), sdk.MsgTypeURL(&types.MsgActivateRequest{})},
		// Possible names: types.TypeAddAccessRequest, fmt.Sprintf("%T", &types.MsgAddAccessRequest{})
		{simappparams.DefaultWeightMsgAddAccess, sdk.MsgTypeURL(&types.MsgAddAccessRequest{}), sdk.MsgTypeURL(&types.MsgAddAccessRequest{})},
		{simappparams.DefaultWeightMsgAddFinalizeActivateMarker, sdk.MsgTypeURL(&types.MsgAddFinalizeActivateMarkerRequest{}), sdk.MsgTypeURL(&types.MsgAddFinalizeActivateMarkerRequest{})},
		{simappparams.DefaultWeightMsgAddMarkerProposal, sdk.MsgTypeURL(&types.MsgAddMarkerProposalRequest{}), sdk.MsgTypeURL(&types.MsgAddMarkerProposalRequest{})},
	}
	for i, w := range weightedOps {
		operationMsg, _, _ := w.Op()(r, suite.app.BaseApp, suite.ctx, accs, "")
		// the following checks are very much dependent from the ordering of the output given
		// by WeightedOperations. if the ordering in WeightedOperations changes some tests
		// will fail
		suite.Require().Equal(expected[i].weight, w.Weight(), "weight should be the same")
		suite.Require().Equal(expected[i].opMsgRoute, operationMsg.Route, "route should be the same")
		suite.Require().Equal(expected[i].opMsgName, operationMsg.Name, "operation Msg name should be the same")
	}
}

// TestSimulateMsgAddMarker tests the normal scenario of a valid message of type TypeAddMarkerRequest.
// Abnormal scenarios, where the message is created by an errors, are not tested here.
func (suite *SimTestSuite) TestSimulateMsgAddMarker() {

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 3)

	// begin a new block
	suite.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: suite.app.LastBlockHeight() + 1, AppHash: suite.app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgAddMarker(suite.app.MarkerKeeper, suite.app.AccountKeeper, suite.app.BankKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg types.MsgAddMarkerRequest
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	suite.Require().True(operationMsg.OK, operationMsg.String())
	suite.Require().Equal(sdk.MsgTypeURL(&msg), operationMsg.Name)
	suite.Require().Equal(sdk.MsgTypeURL(&msg), operationMsg.Route)
	suite.Require().Len(futureOperations, 0)
}

// TestSimulateMsgAddActivateFinalizeMarker tests the normal scenario of a valid message of type TypeAddActivateFinalizeMarkerRequest.
// Abnormal scenarios, where the message is created by an errors, are not tested here.
func (suite *SimTestSuite) TestSimulateMsgAddActivateFinalizeMarker() {

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 3)

	// begin a new block
	suite.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: suite.app.LastBlockHeight() + 1, AppHash: suite.app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgAddFinalizeActivateMarker(suite.app.MarkerKeeper, suite.app.AccountKeeper, suite.app.BankKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg types.MsgAddFinalizeActivateMarkerRequest
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	suite.Require().True(operationMsg.OK, operationMsg.String())
	suite.Require().Equal(sdk.MsgTypeURL(&msg), operationMsg.Name)
	suite.Require().Equal(sdk.MsgTypeURL(&msg), operationMsg.Route)
	suite.Require().Len(futureOperations, 0)
}

func (suite *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)

	initAmt := sdk.TokensFromConsensusPower(1000000, sdk.DefaultPowerReduction)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, account.Address)
		suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
		err := testutil.FundAccount(suite.app.BankKeeper, suite.ctx, account.Address, initCoins)
		suite.Require().NoError(err)
	}

	return accounts
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}

// TestSimulateMsgAddMarkerProposal tests the normal scenario of a valid message of type MsgAddMarkerProposalRequest.
// Abnormal scenarios, where the message is created by an errors, are not tested here.
func (suite *SimTestSuite) TestSimulateMsgAddMarkerProposal() {

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	wopArgs := s.getWeightedOpsArgs()
	voteType := sdk.MsgTypeURL(&govv1.MsgVote{})

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var op simtypes.Operation
			testFunc := func() {
				op = simulation.SimulateGovMsgUpdateParams(&wopArgs)
			}
			s.Require().NotPanics(testFunc, "SimulateGovMsgUpdateParams")
			var opMsg simtypes.OperationMsg
			var fops []simtypes.FutureOperation
			var err error
			testOp := func() {
				opMsg, fops, err = op(tc.r, s.app.BaseApp, s.ctx, tc.accs, chainID)
			}
			s.Require().NotPanics(testOp, "SimulateGovMsgUpdateParams op execution")
			testutil.AssertErrorContents(s.T(), err, tc.expInErr, "op error")
			s.Assert().Equal(tc.expOpMsgOK, opMsg.OK, "op msg ok")
			s.Assert().Equal(tc.expOpMsgRoute, opMsg.Route, "op msg route")
			s.Assert().Equal(tc.expOpMsgName, opMsg.Name, "op msg name")
			s.Assert().Equal(tc.expOpMsgComment, opMsg.Comment, "op msg comment")
			if !tc.expOpMsgOK && !opMsg.OK {
				s.Assert().Empty(fops, "future ops")
			}
			if tc.expOpMsgOK && opMsg.OK {
				s.Assert().Equal(len(tc.accs), len(fops), "number of future ops")
				// If we were expecting it to be okay, and it was, run all the future ops too.
				// Some of them might fail (due to being sanctioned),
				// but all the ones that went through should be YES votes.
				maxBlockTime := s.ctx.BlockHeader().Time.Add(votingPeriod)
				prop := s.getLastGovProp()
				s.Assert().Equal(govMinDep.String(), sdk.NewCoins(prop.TotalDeposit...).String(), "prop deposit")
				preVotes := s.app.GovKeeper.GetVotes(s.ctx, prop.Id)
				// There shouldn't be any votes yet.
				if !s.Assert().Empty(preVotes, "votes before running future ops") {
					for i, fop := range fops {
						s.Assert().LessOrEqual(fop.BlockTime, maxBlockTime, "future op %d block time", i+1)
						s.Assert().Equal(0, fop.BlockHeight, "future op %d block height", i+1)
						var fopMsg simtypes.OperationMsg
						var ffops []simtypes.FutureOperation
						testFop := func() {
							fopMsg, ffops, err = fop.Op(rand.New(rand.NewSource(1)), s.app.BaseApp, s.ctx, tc.accs, chainID)
						}
						if !s.Assert().NotPanics(testFop, "future op %d execution", i+1) {
							continue
						}
						if err != nil {
							s.T().Logf("future op %d returned an error, but that's kind of expected: %v", i+1, err)
							continue
						}
						if !fopMsg.OK {
							s.T().Logf("future op %d returned not okay, but that's kind of expected: %q", i+1, fopMsg.Comment)
							continue
						}
						s.Assert().Empty(ffops, "future ops returned by future op %d", i+1)
						s.Assert().Equal(voteType, fopMsg.Name, "future op %d msg name", i+1)
						s.Assert().Equal(tc.expOpMsgComment, fopMsg.Comment, "future op %d msg comment", i+1)
					}
					// Now there should be some votes.
					postVotes := s.app.GovKeeper.GetVotes(s.ctx, prop.Id)
					for i, vote := range postVotes {
						if s.Assert().Len(vote.Options, 1, "vote %d options count", i+1) {
							s.Assert().Equal(govv1.OptionYes, vote.Options[0].Option, "vote %d option", i+1)
							s.Assert().Equal("1.000000000000000000", vote.Options[1].Weight, "vote %d weight", i+1)
						}
					}
				}
				// Now, get the message and check its content.
				msgs, err := prop.GetMsgs()
				if s.Assert().NoError(err, "getting messages from the proposal") {
					if s.Assert().Len(msgs, 1, "number of messages in the proposal") {
						msg, ok := msgs[0].(*sanction.MsgUpdateParams)
						if s.Assert().True(ok, "could not cast prop msg to MsgUpdateParams") {
							if !s.Assert().Equal(tc.expParams, msg.Params, "params in gov prop") && tc.expParams != nil && msg.Params != nil {
								s.Assert().Equal(tc.expParams.ImmediateSanctionMinDeposit.String(),
									msg.Params.ImmediateSanctionMinDeposit.String(),
									"ImmediateSanctionMinDeposit")
								s.Assert().Equal(tc.expParams.ImmediateUnsanctionMinDeposit.String(),
									msg.Params.ImmediateUnsanctionMinDeposit.String(),
									"ImmediateUnsanctionMinDeposit")
							}
						}
					}
				}
			}
			s.nextBlock()
		})
	}
}

// getWeightedOpsArgs creates a standard WeightedOpsArgs.
func (s *SimTestSuite) getWeightedOpsArgs() simulation.WeightedOpsArgs {
	return simulation.WeightedOpsArgs{
		AppParams:  make(simtypes.AppParams),
		JSONCodec:  s.app.AppCodec(),
		ProtoCodec: codec.NewProtoCodec(s.app.InterfaceRegistry()),
		AK:         s.app.AccountKeeper,
		BK:         s.app.BankKeeper,
		GK:         s.app.GovKeeper,
	}
}
