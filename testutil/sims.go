package testutil

import (
	"bytes"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/testutil/assertions"
)

// LogOperationMsg outputs an OperationMsg to test logs. The provided msg and args are included first in the output.
func LogOperationMsg(t *testing.T, operationMsg simtypes.OperationMsg, msg string, args ...interface{}) {
	msgFmt := "%q"
	if len(bytes.TrimSpace(operationMsg.Msg)) == 0 {
		msgFmt = "    %q"
	}
	fmtLines := []string{
		fmt.Sprintf(msg, args...),
		"operationMsg.Route:   %q",
		"operationMsg.Name:    %q",
		"operationMsg.Comment: %q",
		"operationMsg.OK:      %t",
		"operationMsg.Msg: " + msgFmt,
	}
	t.Logf(strings.Join(fmtLines, "\n"),
		operationMsg.Route, operationMsg.Name, operationMsg.Comment, operationMsg.OK, string(operationMsg.Msg),
	)
}

// SimTestHelper contains various things needed to check most of the simulation-related stuff.
type SimTestHelper struct {
	// T is the testing.T to use for the tests.
	T *testing.T
	// Ctx is the context to provide to the various simulation-related functions.
	Ctx sdk.Context
	// R is the randomizer to provide to the various simulation-related functions.
	R *rand.Rand
	// App is the app object used in some of the simulation-related functions, and also used to create test accounts.
	App *app.App
	// Accs are the accounts to provide to the various simulation-related funtions.
	Accs []simtypes.Account
}

// NewSimTestHelper creates a new SimTestHelper with the provided aspects.
// You'll probably want to call either WithAccs or WithTestingAccounts on the result to finish setting it up.
func NewSimTestHelper(t *testing.T, ctx sdk.Context, r *rand.Rand, app *app.App) *SimTestHelper {
	return &SimTestHelper{
		T:   t,
		Ctx: ctx,
		R:   r,
		App: app,
	}
}

// WithT returns a copy of this SimTestHelper, but with a different T.
func (a SimTestHelper) WithT(t *testing.T) *SimTestHelper {
	a.T = t
	return &a
}

// WithR returns a copy of this SimTestHelper, but with a different R.
func (a SimTestHelper) WithR(r *rand.Rand) *SimTestHelper {
	a.R = r
	return &a
}

// WithCtx returns a copy of this SimTestHelper, but with a different Ctx.
func (a SimTestHelper) WithCtx(ctx sdk.Context) *SimTestHelper {
	a.Ctx = ctx
	return &a
}

// WithApp returns a copy of this SimTestHelper, but with a different App.
func (a SimTestHelper) WithApp(app *app.App) *SimTestHelper {
	a.App = app
	return &a
}

// WithAccs returns a copy of this SimTestHelper, but with the provided Accs.
func (a SimTestHelper) WithAccs(accs []simtypes.Account) *SimTestHelper {
	a.Accs = accs
	return &a
}

// WithTestingAccounts returns a copy of this SimTestHelper, but with n randomized Accs.
func (a SimTestHelper) WithTestingAccounts(n int) *SimTestHelper {
	a.Accs = GenerateTestingAccounts(a.T, a.Ctx, a.App, a.R, n)
	return &a
}

// GenerateTestingAccounts generates n new accounts, creates them (in state) and gives them 1 million power worth of bond tokens.
func GenerateTestingAccounts(t *testing.T, ctx sdk.Context, app *app.App, r *rand.Rand, n int) []simtypes.Account {
	if n <= 0 {
		return nil
	}
	t.Helper()

	initAmt := sdk.TokensFromConsensusPower(1_000_000, sdk.DefaultPowerReduction)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	accs := simtypes.RandomAccounts(r, n)
	// add coins to the accounts
	for i, account := range accs {
		acc := app.AccountKeeper.NewAccountWithAddress(ctx, account.Address)
		app.AccountKeeper.SetAccount(ctx, acc)
		err := testutil.FundAccount(ctx, app.BankKeeper, account.Address, initCoins)
		require.NoError(t, err, "[%d]: FundAccount", i)
	}

	return accs
}

// ExpectedWeightedOp is the various aspects of a simulation.WeightedOperation to check.
type ExpectedWeightedOp struct {
	// Weight is the expected WeightedOperation.Weight.
	Weight int
	// Route is the expected WeightedOperation.Route.
	Route string
	// MsgType is a Msg with the same type that is expected to be returned by the WeightedOperation.Op() function.
	// No assertion is made that the returned Msg equals this message, only that they both have the same result from sdk.MsgTypeURL.
	MsgType sdk.Msg
}

// AssertWeightedOperations tests that the results of a WeightedOperations function are as expected.
// The returned bool is true iff everything is as expected.
func (a SimTestHelper) AssertWeightedOperations(expected []ExpectedWeightedOp, wOpsFn func() simulation.WeightedOperations) bool {
	a.T.Helper()
	var wOps simulation.WeightedOperations
	testWOpsFn := func() {
		wOps = wOpsFn()
	}
	if !assert.NotPanics(a.T, testWOpsFn, "Executing the provided WeightedOperations function.") {
		return false
	}

	expNames := make([]string, len(expected))
	for i, exp := range expected {
		expNames[i] = sdk.MsgTypeURL(exp.MsgType)
	}

	// Run all the ops and get the operation messages and their names.
	opMsgs := make([]simtypes.OperationMsg, len(wOps))
	actNames := make([]string, len(wOps))
	for i, w := range wOps {
		testFunc := func() {
			opMsgs[i], _, _ = w.Op()(a.R, a.App.BaseApp, a.Ctx, a.Accs, "")
		}
		if !assert.NotPanics(a.T, testFunc, "executing OperationMsg[%d]", i) {
			return false
		}
		actNames[i] = opMsgs[i].Name
	}

	// First, make sure the op names are as expected since a failure there probably means the rest will fail.
	// And it's probably easier to address when you've got a nice list comparison of names and their orderings.
	if !assert.Equal(a.T, expNames, actNames, "operation message names") {
		return false
	}

	// Now assert that each entry was as expected.
	ok := true
	for i := range expected {
		a.T.Run(fmt.Sprintf("%d: %s", i, expNames[i]), func(t *testing.T) {
			LogOperationMsg(t, opMsgs[i], "WeightedOperation [%d]", i)
			ok = assert.Equal(t, expected[i].Weight, wOps[i].Weight(), "weightedOps[%d].Weight", i) && ok
			ok = assert.Equal(t, expected[i].Route, opMsgs[i].Route, "weightedOps[%d] operationMsg.Route", i) && ok
			ok = assert.Equal(t, expNames[i], opMsgs[i].Name, "weightedOps[%d] operationMsg.Name", i) && ok
		})
	}

	return ok
}

// ExpectedProposalMsg is the various aspects of a simulation.WeightedProposalMsg to check.
type ExpectedProposalMsg struct {
	// Key is the expected WeightedProposalMsg.AppParamsKey()
	Key string
	// Weight is the expected WeightedProposalMsg.DefaultWeight()
	Weight int
	// MsgType is a Msg with the same type that is expected to be returned by the WeightedProposalMsg.MsgSimulatorFn().
	// No assertion is made that the returned Msg equals this message, only that they both have the same result from sdk.MsgTypeURL.
	MsgType sdk.Msg
}

// AssertProposalMsgs tests that the results of a ProposalMsgs function are as expected.
// The returned bool is true iff everything is as expected.
func (a SimTestHelper) AssertProposalMsgs(expected []ExpectedProposalMsg, proposalMsgsFn func() []simtypes.WeightedProposalMsg) bool {
	a.T.Helper()
	var wOps []simtypes.WeightedProposalMsg
	testWOps := func() {
		wOps = proposalMsgsFn()
	}
	if !assert.NotPanics(a.T, testWOps, "Executing the provided ProposalMsgs function.") {
		return false
	}

	expNames := make([]string, len(expected))
	for i, exp := range expected {
		expNames[i] = sdk.MsgTypeURL(exp.MsgType)
	}
	actNames := make([]string, len(wOps))
	propMsgs := make([]sdk.Msg, len(wOps))
	ok := true
	for i, act := range wOps {
		testFunc := func() {
			propMsgs[i] = act.MsgSimulatorFn()(a.R, a.Ctx, a.Accs)
		}
		if !assert.NotPanics(a.T, testFunc, "executing MsgSimulatorFn [%d]", i) {
			ok = false
			actNames[i] = "<panic>"
		} else {
			actNames[i] = sdk.MsgTypeURL(propMsgs[i])
		}
	}

	// First, make sure the op names are as expected since a failure there probably means the rest will fail.
	// And it's probably easier to address when you've got a nice list comparison of names and their orderings.
	if !assert.Equal(a.T, expNames, actNames, "operation message names") {
		return false
	}

	// Now assert that each entry was as expected.
	for i, exp := range expected {
		a.T.Run(fmt.Sprintf("%d: %s", i, exp.Key), func(t *testing.T) {
			ok = assert.Equal(t, exp.Key, wOps[i].AppParamsKey(), "AppParamsKey()") && ok
			ok = assert.Equal(t, exp.Weight, wOps[i].DefaultWeight(), "DefaultWeight()") && ok
			ok = assert.NotNil(t, propMsgs[i], "MsgSimulatorFn result") && ok
		})
	}

	return ok
}

// ExpectedOp is the various aspects of a simulation.Operation to check.
type ExpectedOp struct {
	// FutureOpCount is the exected length of the  []FutureOperation returned by the Operation.
	FutureOpCount int
	// Err is the error expected from the Operation.
	Err string

	// Route is the expected OperationMsg.Route value (usually the module name).
	Route string
	// EmptyMsg serves two purposes:
	//	1. The OperationMsg.Name is expected to be the sdk.MsgTypeURL of this field.
	//  2. The OperationMsg.Msg will be unmarshalled into this field if possible.
	EmptyMsg sdk.Msg
	// Comment is the expected OperationMsg.Comment value.
	Comment string
	// OK is the expected OperationMsg.OK value.
	OK bool
}

// AssertSimOp tests that a simulation Operation is as expected.
// The returned bool is true iff everything is as expected.
// The msg and future ops are returned even if things aren't as expected.
func (a SimTestHelper) AssertSimOp(expected ExpectedOp, opMaker func() simtypes.Operation, opDesc string) (sdk.Msg, []simtypes.FutureOperation, bool) {
	a.T.Helper()
	var op simtypes.Operation
	testOp := func() {
		op = opMaker()
	}
	if !assert.NotPanics(a.T, testOp, "Running the Operation Maker.") {
		return nil, nil, false
	}
	if !assert.NotNil(a.T, op, "The Operation") {
		return nil, nil, false
	}

	// Run the op, and log it.
	var opMsg simtypes.OperationMsg
	var fOps []simtypes.FutureOperation
	var opErr error
	testFunc := func() {
		opMsg, fOps, opErr = op(a.R, a.App.BaseApp, a.Ctx, a.Accs, "")
	}
	if !assert.NotPanics(a.T, testFunc, "executing the simtypes.Operation(...)") {
		return nil, fOps, false
	}
	LogOperationMsg(a.T, opMsg, opDesc)

	// Check the error and number of future ops.
	ok := true
	ok = assertions.AssertErrorValue(a.T, opErr, expected.Err, "error returned from the simtypes.Operation(...)") && ok
	ok = assert.Len(a.T, fOps, expected.FutureOpCount, "future ops returned from the simtypes.Operation(...)") && ok

	// Attempt to unmarshal the msg now, so it can be available even if we end up returning early.
	// We check the error after the short-circuit because of the case when we're expecting an error,
	// but don't get one. When that happens, the error assertion above will fail, so ok will be false,
	// and we don't really care whether the Msg is the type expected. Checking this error before the
	// short-circuit is likely to just pollute the test output with a failure that isn't helpful.
	// The content of the Msg was logged, so it's available for reference if needed for troubleshooting.
	//
	// Further, there are some cases (e.g. a no-op) where the Msg is expected to be empty.
	// We want to allow for this, so we only make this conversion attempt if it's not empty.
	var msgErr error
	if len(opMsg.Msg) > 0 {
		msgErr = a.App.AppCodec().Unmarshal(opMsg.Msg, expected.EmptyMsg)
	}

	// If the Operation returned an error, or we were expecting one, the rest of the checks are meaningless.
	if opErr != nil || len(expected.Err) > 0 {
		return expected.EmptyMsg, fOps, ok
	}

	// Check the opMsg results.
	ok = assert.NoError(a.T, msgErr, "Unmarshal(opMsg.Msg) error") && ok
	ok = assert.Equal(a.T, expected.OK, opMsg.OK, "opMsg.OK") && ok
	ok = assert.Equal(a.T, sdk.MsgTypeURL(expected.EmptyMsg), opMsg.Name, "opMsg.Name") && ok
	ok = assert.Equal(a.T, expected.Route, opMsg.Route, "opMsg.Route") && ok
	ok = assert.Equal(a.T, expected.Comment, opMsg.Comment, "opMsg.Comment") && ok

	return expected.EmptyMsg, fOps, ok
}

// AssertMsgSimulatorFn tests that a MsgSimulatorFn generates the expected Msg.
// The returned bool is true iff everything is as expected.
// The returned Msg might be nil regardless of the returned bool value.
func (a SimTestHelper) AssertMsgSimulatorFn(expected sdk.Msg, simFnMaker func() simtypes.MsgSimulatorFn) (sdk.Msg, bool) {
	a.T.Helper()
	var msgSimFn simtypes.MsgSimulatorFn
	testMaker := func() {
		msgSimFn = simFnMaker()
	}
	if !assert.NotPanics(a.T, testMaker, "The provided MsgSimulatorFn maker.") {
		return nil, false
	}
	if !assert.NotNil(a.T, msgSimFn, "The MsgSimulatorFn returned by the provided maker.") {
		return nil, false
	}

	var actual sdk.Msg
	testMsgSimFn := func() {
		actual = msgSimFn(a.R, a.Ctx, a.Accs)
	}
	if !assert.NotPanics(a.T, testMsgSimFn, "Executing the MsgSimulatorFn") {
		return actual, false
	}
	return actual, assert.Equal(a.T, expected, actual)
}
