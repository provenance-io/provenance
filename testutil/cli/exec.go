package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	sdkcli "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/testutil/queries"
)

// TxExecutor helps facilitate the execution and testing of a CLI command.
//
// The command will be executed when either .Execute() or .AssertExecute() are called.
// The former will halt the test upon failure, the latter will allow test execution to continue.
//
// The error returned from the command is tested against ExpErr, ExpErrMsg, and/or ExpInErrMsg.
// If none of those are set, the error form the command must be nil.
//
// If the command did not return an error, the Tx is queried, so that we can check the actual result.
// That means that this will block for at least one block while it waits for the tx to be processed.
//
// The result code and raw log are only checked if a Tx result is available.
// It's considered a failure if the command did not return an error, but the Tx result is not available.
type TxExecutor struct {
	// Name is the name of the test. It's not actually used in here, but included
	// in case you want to use []TxExecutor to define your test cases.
	Name string
	// Cmd is the cobra.Command to execute.
	Cmd *cobra.Command
	// Args are all the arguments to provide to the command.
	Args []string

	// ExpErr, if true, the command must return an error.
	// If ExpErr is false, and both ExpErrMsg and ExpInErrMsg are empty, the command must NOT return an error.
	ExpErr bool
	// ExpErrMsg, if not empty, the command must return an error that equals this provided string.
	// If ExpErr is false, and both ExpErrMsg and ExpInErrMsg are empty, the command must NOT return an error.
	ExpErrMsg string
	// ExpInErrMsg, if not empty, the command must return an error that contains each of the provided strings.
	// If ExpErr is false, and both ExpErrMsg and ExpInErrMsg are empty, the command must NOT return an error.
	ExpInErrMsg []string

	// ExpCode is the expected response code from the Tx.
	ExpCode uint32

	// ExpRawLog, if not empty, the TxResponse.RawLog must equal this string.
	// If both ExpRawLog and ExpInRawLog are empty, the TxResponse.RawLog is ignored.
	ExpRawLog string
	// ExpInRawLog, if not empty, the TxResponse.RawLog must contain each of the provided strings.
	// If both ExpRawLog and ExpInRawLog are empty, the TxResponse.RawLog is ignored.
	ExpInRawLog []string
}

// NewTxExecutor creates a new TxExecutor with the provided command and args.
func NewTxExecutor(cmd *cobra.Command, args []string) TxExecutor {
	return TxExecutor{
		Cmd:  cmd,
		Args: args,
	}
}

// WithName returns a copy of this TxExecutor that has the provided Name.
func (c TxExecutor) WithName(name string) TxExecutor {
	c.Name = name
	return c
}

// WithCmd returns a copy of this TxExecutor that has the provided Cmd.
func (c TxExecutor) WithCmd(cmd *cobra.Command) TxExecutor {
	c.Cmd = cmd
	return c
}

// WithArgs returns a copy of this TxExecutor that has the provided Args.
func (c TxExecutor) WithArgs(args []string) TxExecutor {
	c.Args = args
	return c
}

// WithExpErr returns a copy of this TxExecutor that has the provided ExpErr.
func (c TxExecutor) WithExpErr(expErr bool) TxExecutor {
	c.ExpErr = expErr
	return c
}

// WithExpErrMsg returns a copy of this TxExecutor that has the provided ExpErrMsg.
func (c TxExecutor) WithExpErrMsg(expErrMsg string) TxExecutor {
	c.ExpErrMsg = expErrMsg
	return c
}

// WithExpInErrMsg returns a copy of this TxExecutor that has the provided ExpInErrMsg.
func (c TxExecutor) WithExpInErrMsg(expInErrMsg []string) TxExecutor {
	c.ExpInErrMsg = expInErrMsg
	return c
}

// WithExpCode returns a copy of this TxExecutor that has the provided ExpCode.
func (c TxExecutor) WithExpCode(expCode uint32) TxExecutor {
	c.ExpCode = expCode
	return c
}

// WithExpRawLog returns a copy of this TxExecutor that has the provided ExpRawLog.
func (c TxExecutor) WithExpRawLog(expRawLog string) TxExecutor {
	c.ExpRawLog = expRawLog
	return c
}

// WithExpInRawLog returns a copy of this TxExecutor that has the provided ExpInRawLog.
func (c TxExecutor) WithExpInRawLog(expInRawLog []string) TxExecutor {
	c.ExpInRawLog = expInRawLog
	return c
}

// Execute executes the command, requiring everything is as expected.
//
// It is possible for everything to be as expected, and still get a nil TxResponse from this.
//
// To allow test execution to continue on a failure, use AssertExecute.
func (c TxExecutor) Execute(t *testing.T, n *network.Network) *sdk.TxResponse {
	t.Helper()
	rv, ok := c.AssertExecute(t, n)
	if !ok {
		t.FailNow()
	}
	return rv
}

// AssertExecute executes the command, asserting that everything is as expected.
//
// The returned TxResponse is nil if the command did not generate one, which is not dependent on the returned bool.
// The returned bool is true if everything is as expected, false otherwise.
//
// To halt test execution on failure, use Execute.
func (c TxExecutor) AssertExecute(t *testing.T, n *network.Network) (*sdk.TxResponse, bool) {
	t.Helper()
	if !assert.NotNil(t, c.Cmd, "TxExecutor.Cmd cannot be nil") {
		return nil, false
	}
	if !assert.NotEmpty(t, n.Validators, "Network.Validators") {
		return nil, false
	}

	clientCtx := n.Validators[0].ClientCtx
	out, err := sdkcli.ExecTestCLICmd(clientCtx, c.Cmd, c.Args)
	outBz := out.Bytes()
	t.Logf("ExecTestCLICmd %q %q\nOutput:\n%s", c.Cmd.Name(), c.Args, string(outBz))

	// Make sure the error is as expected.
	ok, expNoErr := true, true
	if c.ExpErr {
		ok = assert.Error(t, err, "ExecTestCLICmd error") && ok
		expNoErr = false
	}
	if len(c.ExpErrMsg) > 0 {
		ok = assert.EqualError(t, err, c.ExpErrMsg, "ExecTestCLICmd error") && ok
		expNoErr = false
	}
	if len(c.ExpInErrMsg) > 0 {
		ok = assertions.AssertErrorContents(t, err, c.ExpInErrMsg, "ExecTestCLICmd error") && ok
		expNoErr = false
	}
	if expNoErr {
		ok = assert.NoError(t, err, "ExecTestCLICmd error") && ok
	}

	var txResp sdk.TxResponse
	var gotResp bool
	if err != nil {
		// If there was an error, the output is likely just command help stuff.
		// But just in case it's not, attempt to convert it to a TxResponse.
		err = clientCtx.Codec.UnmarshalJSON(outBz, &txResp)
		gotResp = err == nil
	} else {
		// If there wasn't an error, the account's sequence number was probably updated.
		// So we always want to get the TxResponse in such a case. At the very least,
		// it makes us wait a block, keeping the sequence number up-to-date for future tests.
		txResp, gotResp = queries.AssertGetTxFromResponse(t, n, outBz)
		ok = ok && gotResp
	}

	// If we weren't able to get a response, there's nothing left to check.
	if !gotResp {
		return nil, ok
	}

	// Check the response code.
	ok = assert.Equal(t, int(c.ExpCode), int(txResp.Code), "response Code") && ok

	if len(c.ExpRawLog) > 0 {
		ok = assert.Equal(t, c.ExpRawLog, txResp.RawLog, "response RawLog") && ok
	}
	if len(c.ExpInRawLog) > 0 {
		for _, exp := range c.ExpInRawLog {
			ok = assert.Contains(t, txResp.RawLog, exp, "response RawLog") && ok
		}
	}

	return &txResp, ok
}
