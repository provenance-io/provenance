package cli_test

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/ghodss/yaml"
	"github.com/gogo/protobuf/proto"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	testnet "github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil"
	"github.com/provenance-io/provenance/x/hold"
	"github.com/provenance-io/provenance/x/hold/client/cli"
	holdkeeper "github.com/provenance-io/provenance/x/hold/keeper"
)

type IntegrationCLITestSuite struct {
	suite.Suite

	cfg     testnet.Config
	testnet *testnet.Network

	addr1 sdk.AccAddress
	addr2 sdk.AccAddress
	addr3 sdk.AccAddress
	addr4 sdk.AccAddress
	addr5 sdk.AccAddress

	addr1Desc string
	addr2Desc string
	addr3Desc string
	addr4Desc string
	addr5Desc string

	addr1Bal sdk.Coins
	addr2Bal sdk.Coins
	addr3Bal sdk.Coins
	addr4Bal sdk.Coins
	addr5Bal sdk.Coins

	addr1Escrow sdk.Coins
	addr2Escrow sdk.Coins
	addr3Escrow sdk.Coins
	addr4Escrow sdk.Coins
	addr5Escrow sdk.Coins

	addr1Spendable sdk.Coins
	addr2Spendable sdk.Coins
	addr3Spendable sdk.Coins
	addr4Spendable sdk.Coins
	addr5Spendable sdk.Coins

	flagAsText     string
	flagAsJSON     string
	flagOffset     string
	flagLimit      string
	flagReverse    string
	flagCountTotal string
	flagPageKey    string
}

func TestIntegrationCLITestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationCLITestSuite))
}

func (s *IntegrationCLITestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")
	pioconfig.SetProvenanceConfig("mota", 0)
	s.cfg = testutil.DefaultTestNetworkConfig()
	s.cfg.NumValidators = 1
	s.cfg.ChainID = antewrapper.SimAppChainID
	s.cfg.TimeoutCommit = 500 * time.Millisecond

	newCoins := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		s.Require().NoError(err, "ParseCoinsNormalized(%q)", coins)
		return rv
	}
	// newAmounts creates the balance, hold, and spendable amounts.
	newAmounts := func(name, plusBalance, escrowAmount string) (sdk.Coins, sdk.Coins, sdk.Coins) {
		balance := sdk.Coins{sdk.NewInt64Coin(s.cfg.BondDenom, 1_000_000_000)}
		if len(plusBalance) > 0 {
			balance = balance.Add(newCoins(plusBalance)...)
		}
		inEscrow := balance
		if escrowAmount != "all" {
			inEscrow = newCoins(escrowAmount)
		}
		// manually doing this subtraction because we need the zero entries.
		hasNeg := false
		var spendable sdk.Coins
		for _, balCoin := range balance {
			escHas, escCoin := inEscrow.Find(balCoin.Denom)
			spendCoin := sdk.Coin{Denom: balCoin.Denom, Amount: balCoin.Amount}
			if escHas && !escCoin.Amount.IsZero() {
				spendCoin.Amount = balCoin.Amount.Sub(escCoin.Amount)
			}
			hasNeg = hasNeg || spendCoin.IsNegative()
			spendable = append(spendable, spendCoin)
		}
		s.Require().False(hasNeg, "%s spendable went negative: %q - %q = %q", name, balance, inEscrow, spendable)
		return balance, inEscrow, spendable
	}

	s.addr1 = sdk.AccAddress("cli_test_address_1__")
	s.addr2 = sdk.AccAddress("cli_test_address_2__")
	s.addr3 = sdk.AccAddress("cli_test_address_3__")
	s.addr4 = sdk.AccAddress("cli_test_address_4__")
	s.addr5 = sdk.AccAddress("cli_test_address_5__")

	// addr1 characteristics:
	// - No hold of bond denom.
	// - Partial hold of one denom.
	// - More than max uint64 of one denom both in hold and spendable.
	// - One denom fully on hold.
	s.addr1Desc = "addr with large amounts"
	addr1Plus := "15banana,5000000000000000000000hugecoin,1xenon"
	addr1Esrow := "5banana,2000000000000000000000hugecoin,1xenon"
	s.addr1Bal, s.addr1Escrow, s.addr1Spendable = newAmounts("addr1", addr1Plus, addr1Esrow)

	// addr2 characteristics:
	// - One extra denom.
	// - Nothing on hold.
	s.addr2Desc = "addr with extra denom no hold"
	s.addr2Bal, s.addr2Escrow, s.addr2Spendable = newAmounts("addr2", "99banana", "")

	// addr3 characteristics:
	// - All funds on hold.
	s.addr3Desc = "addr with all funds on hold"
	s.addr3Bal, s.addr3Escrow, s.addr3Spendable = newAmounts("addr3", "55acorn,12banana", "all")

	// addr4 characteristics:
	// - Most of one denom on hold.
	// - A little of the bond denom on hold.
	// - None of another denom on hold.
	s.addr4Desc = "addr with only a little on hold"
	s.addr4Bal, s.addr4Escrow, s.addr4Spendable = newAmounts("addr4", "93acorn,9carrot", "90acorn,30000"+s.cfg.BondDenom)

	// addr5 characteristics:
	// - Only has bond denom.
	// - Nothing on hold.
	s.addr5Desc = "addr without holds and only bond denom"
	s.addr5Bal, s.addr5Escrow, s.addr5Spendable = newAmounts("addr5", "", "")

	s.flagAsText = fmt.Sprintf("--%s=text", tmcli.OutputFlag)
	s.flagAsJSON = fmt.Sprintf("--%s=json", tmcli.OutputFlag)
	s.flagOffset = "--" + flags.FlagOffset
	s.flagLimit = "--" + flags.FlagLimit
	s.flagReverse = "--" + flags.FlagReverse
	s.flagCountTotal = "--" + flags.FlagCountTotal
	s.flagPageKey = "--" + flags.FlagPageKey

	// Add the accounts to the auth module gen state.
	var authGen authtypes.GenesisState
	err := s.cfg.Codec.UnmarshalJSON(s.cfg.GenesisState[authtypes.ModuleName], &authGen)
	s.Require().NoError(err, "UnmarshalJSON auth gen state")
	newAccounts, err := authtypes.PackAccounts(authtypes.GenesisAccounts{
		authtypes.NewBaseAccount(s.addr1, nil, 0, 1),
		authtypes.NewBaseAccount(s.addr2, nil, 0, 1),
		authtypes.NewBaseAccount(s.addr3, nil, 0, 1),
		authtypes.NewBaseAccount(s.addr4, nil, 0, 1),
		authtypes.NewBaseAccount(s.addr5, nil, 0, 1),
	})
	s.Require().NoError(err, "PackAccounts")
	authGen.Accounts = append(authGen.Accounts, newAccounts...)
	s.cfg.GenesisState[authtypes.ModuleName], err = s.cfg.Codec.MarshalJSON(&authGen)
	s.Require().NoError(err, "MarshalJSON auth gen state")

	// Give each of them balances.
	var bankGen banktypes.GenesisState
	err = s.cfg.Codec.UnmarshalJSON(s.cfg.GenesisState[banktypes.ModuleName], &bankGen)
	s.Require().NoError(err, "UnmarshalJSON bank gen state")
	bankGen.Balances = append(bankGen.Balances,
		banktypes.Balance{Address: s.addr1.String(), Coins: s.addr1Bal},
		banktypes.Balance{Address: s.addr2.String(), Coins: s.addr2Bal},
		banktypes.Balance{Address: s.addr3.String(), Coins: s.addr3Bal},
		banktypes.Balance{Address: s.addr4.String(), Coins: s.addr4Bal},
		banktypes.Balance{Address: s.addr5.String(), Coins: s.addr5Bal},
	)
	s.cfg.GenesisState[banktypes.ModuleName], err = s.cfg.Codec.MarshalJSON(&bankGen)
	s.Require().NoError(err, "MarshalJSON bank gen state")

	// Place some of their stuff on hold.
	var escrowGen hold.GenesisState
	err = s.cfg.Codec.UnmarshalJSON(s.cfg.GenesisState[hold.ModuleName], &escrowGen)
	s.Require().NoError(err, "UnmarshalJSON auth gen state")
	escrowGen.Escrows = append(escrowGen.Escrows,
		&hold.AccountEscrow{Address: s.addr1.String(), Amount: s.addr1Escrow},
		&hold.AccountEscrow{Address: s.addr2.String(), Amount: s.addr2Escrow},
		&hold.AccountEscrow{Address: s.addr3.String(), Amount: s.addr3Escrow},
		&hold.AccountEscrow{Address: s.addr4.String(), Amount: s.addr4Escrow},
		&hold.AccountEscrow{Address: s.addr5.String(), Amount: s.addr5Escrow},
	)
	s.cfg.GenesisState[hold.ModuleName], err = s.cfg.Codec.MarshalJSON(&escrowGen)
	s.Require().NoError(err, "MarshalJSON hold gen state")

	s.testnet, err = testnet.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err, "creating testnet")

	_, err = s.testnet.WaitForHeight(1)
	s.Require().NoError(err, "waiting for height 1")
}

func (s *IntegrationCLITestSuite) TearDownSuite() {
	testutil.CleanUp(s.testnet, s.T())
}

// queryCmdTestCase is a test case struct that provides common functionality in these tests.
type queryCmdTestCase struct {
	// name is the name of the test.
	name string
	// cmd is the command to run. Always create a fresh one for each test.
	cmd *cobra.Command
	// args are the arguments to provide with the command.
	args []string
	// expErr is expected to be in an error.
	// This is expected to be in the output as well.
	expErr string
	// expOut is the entire expected output string.
	// If providing this, you probably don't need expInOut.
	expOut string
	// expInOut is substrings that are expected to be in the output.
	// Provide this without an expOut if you only care about a portion of the output.
	expInOut []string
}

// assertQueryCmdTestCase executes the query command and asserts that the error and output are as expected.
func (s *IntegrationCLITestSuite) assertQueryCmdTestCase(tc queryCmdTestCase) bool {
	s.T().Helper()
	cmdName := tc.cmd.Name()
	var outStr string
	defer func() {
		if s.T().Failed() {
			s.T().Logf("Command: %s\nArgs: %q\nOutput:\n%s", cmdName, tc.args, outStr)
		}
	}()

	clientCtx := s.testnet.Validators[0].ClientCtx
	out, err := clitestutil.ExecTestCLICmd(clientCtx, tc.cmd, tc.args)
	outStr = out.String()

	rv := true

	if len(tc.expErr) > 0 {
		rv = s.Assert().ErrorContains(err, tc.expErr, "%s error", cmdName) && rv
		rv = s.Assert().Contains(outStr, tc.expErr, "%s error in the output") && rv
	} else {
		rv = s.Assert().NoError(err, "%s error", cmdName) && rv
	}

	if len(tc.expOut) > 0 {
		rv = s.Assert().Equal(tc.expOut, outStr, "%s output", cmdName) && rv
	}

	for _, exp := range tc.expInOut {
		rv = s.Assert().Contains(outStr, exp, "%s output", cmdName) && rv
	}

	return rv
}

// asJSON returns the provided proto message converted to a json string for expected output.
func (s *IntegrationCLITestSuite) asJSON(p proto.Message) string {
	rv, err := s.cfg.Codec.MarshalJSON(p)
	s.Require().NoError(err, "MarshalJSON(%T)", p)
	return string(rv) + "\n"
}

// asJSON returns the provided proto message converted to a yaml string for expected output.
func (s *IntegrationCLITestSuite) asYAML(p proto.Message) string {
	j, err := s.cfg.Codec.MarshalJSON(p)
	s.Require().NoError(err, "MarshalJSON(%T)", p)
	rv, err := yaml.JSONToYAML(j)
	s.Require().NoError(err, "JSONToYAML(%T)", p)
	return string(rv)
}

func (s *IntegrationCLITestSuite) TestQueryCmd() {
	cmdGen := func() *cobra.Command {
		return cli.QueryCmd()
	}
	resp := func(amount sdk.Coins) *hold.GetEscrowResponse {
		return &hold.GetEscrowResponse{Amount: amount}
	}
	respAll := &hold.GetAllEscrowResponse{
		Escrows: []*hold.AccountEscrow{
			{Address: s.addr1.String(), Amount: s.addr1Escrow},
			// addr2 doesn't have anything on hold.
			{Address: s.addr3.String(), Amount: s.addr3Escrow},
			{Address: s.addr4.String(), Amount: s.addr4Escrow},
			// addr5 doesn't have anything on hold.
		},
		Pagination: &query.PageResponse{
			NextKey: nil,
			Total:   0,
		},
	}

	tests := []queryCmdTestCase{
		{
			name:   "get",
			args:   []string{"get", s.addr1.String(), s.flagAsText},
			expOut: s.asYAML(resp(s.addr1Escrow)),
		},
		{
			name:   "all",
			args:   []string{"all", s.flagAsText},
			expOut: s.asYAML(respAll),
		},
		{
			name:   "get-all",
			args:   []string{"get-all", s.flagAsText},
			expOut: s.asYAML(respAll),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.cmd = cmdGen()
			s.assertQueryCmdTestCase(tc)
		})
	}
}

func (s *IntegrationCLITestSuite) TestQueryCmdGetEscrow() {
	cmdGen := func() *cobra.Command {
		return cli.QueryCmdGetEscrow()
	}
	resp := func(amount sdk.Coins) *hold.GetEscrowResponse {
		return &hold.GetEscrowResponse{Amount: amount}
	}

	unknownAddr := sdk.AccAddress("unknown_address_____")

	tests := []queryCmdTestCase{
		{
			name:   s.addr1Desc + ": get hold as text",
			args:   []string{s.addr1.String(), s.flagAsText},
			expOut: s.asYAML(resp(s.addr1Escrow)),
		},
		{
			name:   s.addr1Desc + ": get hold as json",
			args:   []string{s.addr1.String(), s.flagAsJSON},
			expOut: s.asJSON(resp(s.addr1Escrow)),
		},
		{
			name:   s.addr2Desc + ": get hold as text",
			args:   []string{s.addr2.String(), s.flagAsText},
			expOut: s.asYAML(resp(s.addr2Escrow)),
		},
		{
			name:   s.addr2Desc + ": get hold as json",
			args:   []string{s.addr2.String(), s.flagAsJSON},
			expOut: s.asJSON(resp(s.addr2Escrow)),
		},
		{
			name:   s.addr3Desc + ": get hold as text",
			args:   []string{s.addr3.String(), s.flagAsText},
			expOut: s.asYAML(resp(s.addr3Escrow)),
		},
		{
			name:   s.addr3Desc + ": get hold as json",
			args:   []string{s.addr3.String(), s.flagAsJSON},
			expOut: s.asJSON(resp(s.addr3Escrow)),
		},
		{
			name:   s.addr4Desc + ": get hold as text",
			args:   []string{s.addr4.String(), s.flagAsText},
			expOut: s.asYAML(resp(s.addr4Escrow)),
		},
		{
			name:   s.addr4Desc + ": get hold as json",
			args:   []string{s.addr4.String(), s.flagAsJSON},
			expOut: s.asJSON(resp(s.addr4Escrow)),
		},
		{
			name:   s.addr5Desc + ": get hold as text",
			args:   []string{s.addr5.String(), s.flagAsText},
			expOut: s.asYAML(resp(s.addr5Escrow)),
		},
		{
			name:   s.addr5Desc + ": get hold as json",
			args:   []string{s.addr5.String(), s.flagAsJSON},
			expOut: s.asJSON(resp(s.addr5Escrow)),
		},
		{
			name:   "unknown address",
			args:   []string{unknownAddr.String(), s.flagAsText},
			expOut: s.asYAML(resp(nil)),
		},
		{
			name:   "bad address",
			args:   []string{"not-an-address"},
			expErr: "decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			name:   "no address",
			args:   []string{},
			expErr: "accepts 1 arg(s), received 0",
		},
		{
			name:   "two args",
			args:   []string{s.addr1.String(), s.addr2.String()},
			expErr: "accepts 1 arg(s), received 2",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.cmd = cmdGen()
			s.assertQueryCmdTestCase(tc)
		})
	}
}

func (s *IntegrationCLITestSuite) TestQueryCmdGetAllEscrow() {
	cmdGen := func() *cobra.Command {
		return cli.QueryCmdGetAllEscrow()
	}
	pageKey := func(addr sdk.AccAddress, denom string) []byte {
		return holdkeeper.CreateEscrowCoinKey(addr, denom)[1:]

	}
	pageKeyArg := func(addr sdk.AccAddress, denom string) string {
		return base64.StdEncoding.EncodeToString(pageKey(addr, denom))
	}
	allEntries := []*hold.AccountEscrow{
		{Address: s.addr1.String(), Amount: s.addr1Escrow},
		// addr2 doesn't have anything on hold.
		{Address: s.addr3.String(), Amount: s.addr3Escrow},
		{Address: s.addr4.String(), Amount: s.addr4Escrow},
		// addr5 doesn't have anything on hold.
	}
	resp := func(total uint64, nextKey []byte, escrows ...*hold.AccountEscrow) *hold.GetAllEscrowResponse {
		return &hold.GetAllEscrowResponse{
			Escrows: escrows,
			Pagination: &query.PageResponse{
				NextKey: nextKey,
				Total:   total,
			},
		}
	}

	tests := []queryCmdTestCase{
		{
			name:   "all as text",
			args:   []string{s.flagAsText},
			expOut: s.asYAML(resp(0, nil, allEntries...)),
		},
		{
			name:   "all as json",
			args:   []string{s.flagAsJSON},
			expOut: s.asJSON(resp(0, nil, allEntries...)),
		},
		{
			name:   "offset 1 limit 1",
			args:   []string{s.flagAsText, s.flagOffset, "1", s.flagLimit, "1"},
			expOut: s.asYAML(resp(0, pageKey(s.addr4, "acorn"), allEntries[1])),
		},
		{
			name:   "key for 2nd limit 1",
			args:   []string{s.flagAsText, s.flagPageKey, pageKeyArg(s.addr3, "acorn"), s.flagLimit, "1"},
			expOut: s.asYAML(resp(0, pageKey(s.addr4, "acorn"), allEntries[1])),
		},
		{
			name:   "reversed as text",
			args:   []string{s.flagAsText, s.flagReverse},
			expOut: s.asYAML(resp(0, nil, allEntries[2], allEntries[1], allEntries[0])),
		},
		{
			name:   "limit 1 count total",
			args:   []string{s.flagAsText, s.flagLimit, "1", s.flagCountTotal},
			expOut: s.asYAML(resp(3, pageKey(s.addr3, "acorn"), allEntries[0])),
		},
		{
			name:   "unknown argument",
			args:   []string{"unknown"},
			expErr: "accepts 0 arg(s), received 1",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.cmd = cmdGen()
			s.assertQueryCmdTestCase(tc)
		})
	}
}

// TODO[1607]: Uncomment this TestEscrowRemovedFromSpendable test.
/* commented out until I have a version of the sdk with the GetSpendableBalancesCmd query.
func (s *IntegrationCLITestSuite) TestEscrowRemovedFromSpendable() {
	// The purpose of these tests is to make sure that the bank module is
	// being properly informed of the locked hold funds.
	cmdGen := func() *cobra.Command {
		return bankcli.GetSpendableBalancesCmd()
	}
	resp := func(balances sdk.Coins) *banktypes.QuerySpendableBalancesResponse {
		return &banktypes.QuerySpendableBalancesResponse{
			Balances: balances,
			Pagination: &query.PageResponse{
				NextKey: nil,
				Total:   0,
			},
		}
	}

	tests := []queryCmdTestCase{
		{
			name:   s.addr1Desc + ": get spendable",
			args:   []string{s.addr1.String(), s.flagAsJSON},
			expOut: s.asJSON(resp(s.addr1Spendable)),
		},
		{
			name:   s.addr2Desc + ": get spendable",
			args:   []string{s.addr2.String(), s.flagAsJSON},
			expOut: s.asJSON(resp(s.addr2Spendable)),
		},
		{
			name:   s.addr3Desc + ": get spendable",
			args:   []string{s.addr3.String(), s.flagAsJSON},
			expOut: s.asJSON(resp(s.addr3Spendable)),
		},
		{
			name:   s.addr4Desc + ": get spendable",
			args:   []string{s.addr4.String(), s.flagAsJSON},
			expOut: s.asJSON(resp(s.addr4Spendable)),
		},
		{
			name:   s.addr5Desc + ": get spendable",
			args:   []string{s.addr5.String(), s.flagAsJSON},
			expOut: s.asJSON(resp(s.addr5Spendable)),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.cmd = cmdGen()
			s.assertQueryCmdTestCase(tc)
		})
	}
}
*/
