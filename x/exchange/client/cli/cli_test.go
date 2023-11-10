package cli_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	testnet "github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil"
	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/exchange"
	"github.com/provenance-io/provenance/x/exchange/client/cli"
)

type CmdTestSuite struct {
	suite.Suite

	cfg          testnet.Config
	testnet      *testnet.Network
	keyring      keyring.Keyring
	keyringDir   string
	accountAddrs []sdk.AccAddress

	addr0 sdk.AccAddress
	addr1 sdk.AccAddress
	addr2 sdk.AccAddress
	addr3 sdk.AccAddress
	addr4 sdk.AccAddress
	addr5 sdk.AccAddress
	addr6 sdk.AccAddress
	addr7 sdk.AccAddress
	addr8 sdk.AccAddress
	addr9 sdk.AccAddress
}

func TestCmdTestSuite(t *testing.T) {
	suite.Run(t, new(CmdTestSuite))
}

func (s *CmdTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")
	pioconfig.SetProvenanceConfig("", 0)
	s.cfg = testutil.DefaultTestNetworkConfig()
	s.cfg.NumValidators = 1
	s.cfg.ChainID = antewrapper.SimAppChainID
	s.cfg.TimeoutCommit = 500 * time.Millisecond

	s.generateAccountsWithKeyring(10)
	s.addr0 = s.accountAddrs[0]
	s.addr1 = s.accountAddrs[1]
	s.addr2 = s.accountAddrs[2]
	s.addr3 = s.accountAddrs[3]
	s.addr4 = s.accountAddrs[4]
	s.addr5 = s.accountAddrs[5]
	s.addr6 = s.accountAddrs[6]
	s.addr7 = s.accountAddrs[7]
	s.addr8 = s.accountAddrs[8]
	s.addr9 = s.accountAddrs[9]

	// Add accounts to auth gen state.
	var authGen authtypes.GenesisState
	err := s.cfg.Codec.UnmarshalJSON(s.cfg.GenesisState[authtypes.ModuleName], &authGen)
	s.Require().NoError(err, "UnmarshalJSON auth gen state")
	genAccs := make(authtypes.GenesisAccounts, len(s.accountAddrs))
	for i, addr := range s.accountAddrs {
		genAccs[i] = authtypes.NewBaseAccount(addr, nil, 0, 1)
	}
	newAccounts, err := authtypes.PackAccounts(genAccs)
	s.Require().NoError(err, "PackAccounts")
	authGen.Accounts = append(authGen.Accounts, newAccounts...)
	s.cfg.GenesisState[authtypes.ModuleName], err = s.cfg.Codec.MarshalJSON(&authGen)
	s.Require().NoError(err, "MarshalJSON auth gen state")

	// Add balances to bank gen state.
	balance := sdk.NewCoins(
		sdk.NewInt64Coin(s.cfg.BondDenom, 1_000_000_000),
		sdk.NewInt64Coin("apple", 1_000_000_000),
		sdk.NewInt64Coin("peach", 1_000_000_000),
	)
	var bankGen banktypes.GenesisState
	err = s.cfg.Codec.UnmarshalJSON(s.cfg.GenesisState[banktypes.ModuleName], &bankGen)
	s.Require().NoError(err, "UnmarshalJSON bank gen state")
	for _, addr := range s.accountAddrs {
		bankGen.Balances = append(bankGen.Balances, banktypes.Balance{Address: addr.String(), Coins: balance})
	}
	s.cfg.GenesisState[banktypes.ModuleName], err = s.cfg.Codec.MarshalJSON(&bankGen)
	s.Require().NoError(err, "MarshalJSON bank gen state")

	// Add some markets to the exchange gen state.
	var exchangeGen exchange.GenesisState
	err = s.cfg.Codec.UnmarshalJSON(s.cfg.GenesisState[exchange.ModuleName], &exchangeGen)
	s.Require().NoError(err, "UnmarshalJSON exchange gen state")
	exchangeGen.Params = exchange.DefaultParams()
	exchangeGen.Markets = append(exchangeGen.Markets,
		exchange.Market{
			MarketId: 3,
			MarketDetails: exchange.MarketDetails{
				Name:        "Market Three",
				Description: "The third market (or is it?). It only has ask/seller fees.",
			},
			FeeCreateAskFlat:        []sdk.Coin{sdk.NewInt64Coin("peach", 10)},
			FeeSellerSettlementFlat: []sdk.Coin{sdk.NewInt64Coin("peach", 50)},
			FeeSellerSettlementRatios: []exchange.FeeRatio{
				{Price: sdk.NewInt64Coin("peach", 100), Fee: sdk.NewInt64Coin("peach", 1)},
			},
			AcceptingOrders:     true,
			AllowUserSettlement: true,
			AccessGrants: []exchange.AccessGrant{
				{Address: s.addr1.String(), Permissions: exchange.AllPermissions()},
			},
		},
		exchange.Market{
			MarketId: 5,
			MarketDetails: exchange.MarketDetails{
				Name:        "Market Five",
				Description: "Market the Fifth. It only has bid/buyer fees.",
			},
			FeeCreateBidFlat:       []sdk.Coin{sdk.NewInt64Coin("peach", 10)},
			FeeBuyerSettlementFlat: []sdk.Coin{sdk.NewInt64Coin("peach", 50)},
			FeeBuyerSettlementRatios: []exchange.FeeRatio{
				{Price: sdk.NewInt64Coin("peach", 100), Fee: sdk.NewInt64Coin("peach", 1)},
				{Price: sdk.NewInt64Coin("peach", 100), Fee: sdk.NewInt64Coin(s.cfg.BondDenom, 3)},
			},
			AcceptingOrders:     true,
			AllowUserSettlement: true,
			AccessGrants: []exchange.AccessGrant{
				{Address: s.addr1.String(), Permissions: exchange.AllPermissions()},
			},
		},
		exchange.Market{
			MarketId: 420,
			MarketDetails: exchange.MarketDetails{
				Name:        "THE Market",
				Description: "It's coming; you know it. It has all the fees.",
			},
			FeeCreateAskFlat:        []sdk.Coin{sdk.NewInt64Coin("peach", 20)},
			FeeCreateBidFlat:        []sdk.Coin{sdk.NewInt64Coin("peach", 25)},
			FeeSellerSettlementFlat: []sdk.Coin{sdk.NewInt64Coin("peach", 100)},
			FeeSellerSettlementRatios: []exchange.FeeRatio{
				{Price: sdk.NewInt64Coin("peach", 75), Fee: sdk.NewInt64Coin("peach", 1)},
			},
			FeeBuyerSettlementFlat: []sdk.Coin{sdk.NewInt64Coin("peach", 105)},
			FeeBuyerSettlementRatios: []exchange.FeeRatio{
				{Price: sdk.NewInt64Coin("peach", 50), Fee: sdk.NewInt64Coin("peach", 1)},
				{Price: sdk.NewInt64Coin("peach", 50), Fee: sdk.NewInt64Coin(s.cfg.BondDenom, 3)},
			},
			AcceptingOrders:     true,
			AllowUserSettlement: true,
			AccessGrants: []exchange.AccessGrant{
				{Address: s.addr1.String(), Permissions: exchange.AllPermissions()},
			},
		},
	)
	s.cfg.GenesisState[exchange.ModuleName], err = s.cfg.Codec.MarshalJSON(&exchangeGen)
	s.Require().NoError(err, "MarshalJSON exchange gen state")

	// And fire it all up!!
	s.testnet, err = testnet.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err, "testnet.New(...)")

	_, err = s.testnet.WaitForHeight(1)
	s.Require().NoError(err, "s.testnet.WaitForHeight(1)")
}

func (s *CmdTestSuite) TearDownSuite() {
	testutil.CleanUp(s.testnet, s.T())
}

// generateAccountsWithKeyring creates a keyring and adds a number of keys to it.
// The s.keyringDir, s.keyring, and s.accountAddrs are all set in here.
// The GetClientCtx function returns a context that knows about this keyring.
func (s *CmdTestSuite) generateAccountsWithKeyring(number int) {
	path := hd.CreateHDPath(118, 0, 0).String()
	s.keyringDir = s.T().TempDir()
	var err error
	s.keyring, err = keyring.New(s.T().Name(), "test", s.keyringDir, nil, s.cfg.Codec)
	s.Require().NoError(err, "keyring.New(...)")

	s.accountAddrs = make([]sdk.AccAddress, number)
	for i := range s.accountAddrs {
		keyId := fmt.Sprintf("test_key_%v", i)
		var info *keyring.Record
		info, _, err = s.keyring.NewMnemonic(keyId, keyring.English, path, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
		s.Require().NoError(err, "[%d] s.keyring.NewMnemonic(...)", i)
		s.accountAddrs[i], err = info.GetAddress()
		s.Require().NoError(err, "[%d] getting keyring address", i)
	}
}

// GetClientCtx get a client context that knows about the suite's keyring.
func (s *CmdTestSuite) GetClientCtx() client.Context {
	return s.testnet.Validators[0].ClientCtx.
		WithKeyringDir(s.keyringDir).
		WithKeyring(s.keyring)
}

// txCmdTestCase is a test case for a TX command.
type txCmdTestCase struct {
	// name is a name for this test case.
	name string
	// cmdGen is a generator for the command to execute. If not set, CmdTx is used.
	cmdGen func() *cobra.Command
	// args are the arguments to provide to the command.
	args []string
	// expInErr are strings to expect in an error message (and the output).
	expInErr []string
	// expectedCode is the code expected from the Tx.
	expectedCode uint32
	// followup is any further checks to do.
	followup func(cmdOutput string)
}

// RunTxCmdTestCase runs a txCmdTestCase by executing the command and checking the result.
func (s *CmdTestSuite) RunTxCmdTestCase(tc txCmdTestCase) {
	var cmd *cobra.Command
	if tc.cmdGen != nil {
		cmd = tc.cmdGen()
	} else {
		cmd = cli.CmdTx()
	}

	cmdName := cmd.Name()
	var outBz []byte
	defer func() {
		if s.T().Failed() {
			s.T().Logf("Command: %s\nArgs: %q\nOutput\n%s", cmdName, tc.args, string(outBz))
		}
	}()

	clientCtx := s.GetClientCtx()
	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
	outBz = out.Bytes()

	s.AssertErrorContents(err, tc.expInErr, "ExecTestCLICmd error")
	for _, exp := range tc.expInErr {
		s.Assert().Contains(string(outBz), exp, "command output should contain:\n%q", exp)
	}

	if len(tc.expInErr) == 0 && err == nil {
		var resp sdk.TxResponse
		err = clientCtx.Codec.UnmarshalJSON(outBz, &resp)
		if s.Assert().NoError(err, "UnmarshalJSON(command output) error") {
			if s.Assert().Equalf(int(tc.expectedCode), int(resp.Code), "response code") {
				s.T().Logf("TxResponse:\n%v", resp)
			}
		}
	}

	if tc.followup != nil {
		tc.followup(string(outBz))
	}
}

// queryCmdTestCase is a test case of a query command.
type queryCmdTestCase struct {
	// name is a name for this test case.
	name string
	// cmdGen is a generator for the command to execute. If not set, CmdQuery is used.
	cmdGen func() *cobra.Command
	// args are the arguments to provide to the command.
	args []string
	// expInErr are strings to expect in an error message (and output).
	expInErr []string
	// expInOut are strings to expect in the output.
	expInOut []string
	// expOut is the expected full output. Leave empty to skip this check.
	expOut string
}

// RunQueryCmdTestCase runs a queryCmdTestCase by executing the command and checking the result.
func (s *CmdTestSuite) RunQueryCmdTestCase(tc queryCmdTestCase) {
	var cmd *cobra.Command
	if tc.cmdGen != nil {
		cmd = tc.cmdGen()
	} else {
		cmd = cli.CmdTx()
	}

	cmdName := cmd.Name()
	var outStr string
	defer func() {
		if s.T().Failed() {
			s.T().Logf("Command: %s\nArgs: %q\nOutput\n%s", cmdName, tc.args, outStr)
		}
	}()

	clientCtx := s.GetClientCtx()
	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
	outStr = out.String()

	s.AssertErrorContents(err, tc.expInErr, "ExecTestCLICmd error")
	for _, exp := range tc.expInErr {
		if !s.Assert().Contains(outStr, exp, "command output (error)") {
			s.T().Logf("Not found: %q", exp)
		}
	}

	for _, exp := range tc.expInOut {
		if !s.Assert().Contains(outStr, exp, "command output:\n%q", exp) {
			s.T().Logf("Not found: %q", exp)
		}
	}

	if len(tc.expOut) > 0 {
		s.Assert().Equal(tc.expOut, outStr, "command output string")
	}
}

// AssertErrorContents is a wrapper for assertions.AssertErrorContents using this suite's T().
func (s *CmdTestSuite) AssertErrorContents(theError error, contains []string, msgAndArgs ...interface{}) bool {
	return assertions.AssertErrorContents(s.T(), theError, contains, msgAndArgs...)
}

// joinErrs joins the provided error strings matching to how errors.Join does.
func joinErrs(errs ...string) string {
	return strings.Join(errs, "\n")
}

// toStringSlice applies the stringer to each value and returns a slice with the results.
func toStringSlice[T any](vals []T, stringer func(T) string) []string {
	if vals == nil {
		return nil
	}
	rv := make([]string, len(vals))
	for i, val := range vals {
		rv[i] = stringer(val)
	}
	return rv
}

// assertEqualSlices asserts that the two slices are equal; returns true if so.
// If not, the stringer is applied to each entry and the comparison is redone
// using the strings for a more helpful failure message.
func assertEqualSlices[T any](t *testing.T, expected []T, actual []T, stringer func(T) string, message string, args ...interface{}) bool {
	t.Helper()
	if assert.Equalf(t, expected, actual, message, args...) {
		return true
	}
	expStrs := toStringSlice(expected, stringer)
	actStrs := toStringSlice(actual, stringer)
	assert.Equalf(t, expStrs, actStrs, message+" as strings", args...)
	return false
}

// splitStringer makes a string from the provided DenomSplit.
func splitStringer(split exchange.DenomSplit) string {
	return fmt.Sprintf("%s:%d", split.Denom, split.Split)
}

// orderIDStringer converts an order id to a string.
func orderIDStringer(orderID uint64) string {
	return fmt.Sprintf("%d", orderID)
}

// truncate truncates the provided string returning at most length characters.
func truncate(str string, length int) string {
	if len(str) < length-3 {
		return str
	}
	return str[:length-3] + "..."
}

const (
	// mutExc is the annotation type for "mutually exclusive".
	// It equals the cobra.Command.mutuallyExclusive variable.
	mutExc = "cobra_annotation_mutually_exclusive"
	// oneReq is the annotation type for "one required".
	// It equals the cobra.Command.oneRequired variable.
	oneReq = "cobra_annotation_one_required"
	// mutExc is the annotation type for "required".
	required = cobra.BashCompOneRequiredFlag
)

// setupTestCase contains the stuff that runSetupTestCase should check.
type setupTestCase struct {
	// name is the name of the setup func being tested.
	name string
	// setup is the function being tested.
	setup func(cmd *cobra.Command)
	// expFlags is the list of flags expected to be added to the command after setup.
	// The flags.FlagFrom flag is added to the command prior to calling the setup func;
	// it should be included in this list if you want to check its annotations.
	expFlags []string
	// expAnnotations is the annotations expected for each of the expFlags.
	// The map is "flag name" -> "annotation type" -> values
	// The following variables have the annotation type strings: mutExc, oneReq, required.
	// Annotions are only checked on the flags listed in expFlags.
	expAnnotations map[string]map[string][]string
	// expInUse is a set of strings that are expected to be in the command's Use string.
	// Each entry that does not start with a "[" is also checked to not be in the Use wrapped in [].
	expInUse []string
	// expExamples is a set of examples to ensure are on the command.
	// There must be a full line in the command's Example that matches each entry.
	expExamples []string
	// skipArgsCheck true causes the runner to skip the check ensuring that the command's Args func has been set.
	skipArgsCheck bool
}

// runSetupTestCase runs the provided setup func and checks that everything is set up as expected.
func runSetupTestCase(t *testing.T, tc setupTestCase) {
	if tc.expAnnotations == nil {
		tc.expAnnotations = make(map[string]map[string][]string)
	}
	cmd := &cobra.Command{
		Use: "dummy",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("the dummy command should not have been executed")
		},
	}
	cmd.Flags().String(flags.FlagFrom, "", "The from flag")

	testFunc := func() {
		tc.setup(cmd)
	}
	require.NotPanics(t, testFunc, tc.name)

	for i, flagName := range tc.expFlags {
		t.Run(fmt.Sprintf("flag[%d]: --%s", i, flagName), func(t *testing.T) {
			flag := cmd.Flags().Lookup(flagName)
			if assert.NotNil(t, flag, "--%s", flagName) {
				expAnnotations, _ := tc.expAnnotations[flagName]
				actAnnotations := flag.Annotations
				assert.Equal(t, expAnnotations, actAnnotations, "--%s annotations", flagName)
			}
		})
	}

	for i, exp := range tc.expInUse {
		t.Run(fmt.Sprintf("use[%d]: %s", i, truncate(exp, 20)), func(t *testing.T) {
			assert.Contains(t, cmd.Use, exp, "command use after %s", tc.name)
			if exp[0] != '[' {
				assert.NotContains(t, cmd.Use, "["+exp+"]", "command use after %s", tc.name)
			}
		})
	}

	examples := strings.Split(cmd.Example, "\n")
	for i, exp := range tc.expExamples {
		t.Run(fmt.Sprintf("examples[%d]", i), func(t *testing.T) {
			assert.Contains(t, examples, exp, "command examples after %s", tc.name)
		})
	}

	if !tc.skipArgsCheck {
		t.Run("args", func(t *testing.T) {
			assert.NotNil(t, cmd.Args, "command args after %s", tc.name)
		})
	}
}
