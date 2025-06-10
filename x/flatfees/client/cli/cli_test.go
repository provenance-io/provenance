package cli_test

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
	"sigs.k8s.io/yaml"

	cmtcli "github.com/cometbft/cometbft/libs/cli"

	sdkmath "cosmossdk.io/math"
	upgradecli "cosmossdk.io/x/upgrade/client/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	testnet "github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authcli "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/gogoproto/proto"

	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil"
	testcli "github.com/provenance-io/provenance/testutil/cli"
	"github.com/provenance-io/provenance/x/flatfees/client/cli"
	"github.com/provenance-io/provenance/x/flatfees/types"
	quarantinecli "github.com/provenance-io/provenance/x/quarantine/client/cli"
)

type CLITestSuite struct {
	suite.Suite

	cfg     testnet.Config
	testnet *testnet.Network

	keyring          keyring.Keyring
	keyringEntries   []testutil.TestKeyringEntry
	accountAddresses []sdk.AccAddress

	accountAddr sdk.AccAddress
	accountKey  *secp256k1.PrivKey

	account2Addr sdk.AccAddress
	account2Key  *secp256k1.PrivKey

	genState types.GenesisState
}

func TestCLITestSuite(t *testing.T) {
	suite.Run(t, new(CLITestSuite))
}

func (s *CLITestSuite) SetupSuite() {
	var err error
	s.accountKey = secp256k1.GenPrivKeyFromSecret([]byte("acc2"))
	s.accountAddr, err = sdk.AccAddressFromHexUnsafe(s.accountKey.PubKey().Address().String())
	s.Require().NoError(err, "accountAddr")

	s.account2Key = secp256k1.GenPrivKeyFromSecret([]byte("acc22"))
	s.account2Addr, err = sdk.AccAddressFromHexUnsafe(s.account2Key.PubKey().Address().String())
	s.Require().NoError(err, "account2Addr")

	s.T().Log("setting up integration test suite")
	pioconfig.SetProvConfig("atom")
	govv1.DefaultMinDepositRatio = sdkmath.LegacyZeroDec()

	s.cfg = testutil.DefaultTestNetworkConfig()
	s.cfg.TimeoutCommit = 500 * time.Millisecond
	s.cfg.NumValidators = 1
	s.generateAccountsWithKeyrings(4)

	testutil.MutateGenesisState(s.T(), &s.cfg, banktypes.ModuleName, &banktypes.GenesisState{}, func(bankGenState *banktypes.GenesisState) *banktypes.GenesisState {
		var genBalances []banktypes.Balance
		for i := range s.accountAddresses {
			genBalances = append(genBalances, banktypes.Balance{Address: s.accountAddresses[i].String(), Coins: sdk.NewCoins(
				sdk.NewInt64Coin("nhash", 100_000_000), sdk.NewInt64Coin(s.cfg.BondDenom, 100_000_000),
			).Sort()})
		}
		bankGenState.Params = banktypes.DefaultParams()
		bankGenState.Balances = genBalances
		return bankGenState
	})

	testutil.MutateGenesisState(s.T(), &s.cfg, authtypes.ModuleName, &authtypes.GenesisState{}, func(authData *authtypes.GenesisState) *authtypes.GenesisState {
		var genAccounts []authtypes.GenesisAccount
		authData.Params = authtypes.DefaultParams()
		genAccounts = append(genAccounts, authtypes.NewBaseAccount(s.accountAddresses[0], nil, 3, 0))
		accounts, err := authtypes.PackAccounts(genAccounts)
		s.Require().NoError(err, "should be able to pack accounts for genesis state when setting up suite")
		authData.Accounts = accounts
		return authData
	})

	testutil.MutateGenesisState(s.T(), &s.cfg, types.ModuleName, &types.GenesisState{}, func(flatfeeGen *types.GenesisState) *types.GenesisState {
		flatfeeGen.Params = types.Params{
			DefaultCost: sdk.NewInt64Coin("banana", 10), // 10 banana * 1 stake / 2 banana = 5 stake (default cost).
			ConversionFactor: types.ConversionFactor{
				BaseAmount:      sdk.NewInt64Coin("banana", 2),
				ConvertedAmount: sdk.NewInt64Coin(s.cfg.BondDenom, 1),
			},
		}
		// Note that these are sorted alphabetically here to match the state store.
		flatfeeGen.MsgFees = append(flatfeeGen.MsgFees,
			// MsgSend should remain using the default cost since some tests rely on that being how it's set up.
			// Only the gov prop msg is fee, still gotta pay for the Msgs in it, though.
			types.NewMsgFee("/cosmos.gov.v1.MsgSubmitProposal"),
			// Also make a software upgrade free.
			types.NewMsgFee("/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade"),
			types.NewMsgFee("/cosmos.group.v1.MsgCreateGroup", sdk.NewInt64Coin("banana", 3)),
			types.NewMsgFee("/cosmos.group.v1.MsgCreateGroupPolicy", sdk.NewInt64Coin("banana", 4)),
			types.NewMsgFee("/cosmos.group.v1.MsgCreateGroupWithPolicy", sdk.NewInt64Coin("banana", 5)),
			types.NewMsgFee("/cosmos.group.v1.MsgExec", sdk.NewInt64Coin("banana", 6)),
			types.NewMsgFee("/cosmos.group.v1.MsgLeaveGroup", sdk.NewInt64Coin("banana", 7)),
			types.NewMsgFee("/cosmos.group.v1.MsgSubmitProposal", sdk.NewInt64Coin("banana", 8)),
			types.NewMsgFee("/cosmos.group.v1.MsgUpdateGroupAdmin", sdk.NewInt64Coin("banana", 9)),
			types.NewMsgFee("/cosmos.group.v1.MsgUpdateGroupMembers", sdk.NewInt64Coin("banana", 10)),
			types.NewMsgFee("/cosmos.group.v1.MsgUpdateGroupMetadata", sdk.NewInt64Coin("banana", 11)),
			types.NewMsgFee("/cosmos.group.v1.MsgUpdateGroupPolicyAdmin", sdk.NewInt64Coin("banana", 12)),
			types.NewMsgFee("/cosmos.group.v1.MsgUpdateGroupPolicyDecisionPolicy", sdk.NewInt64Coin("banana", 13)),
			types.NewMsgFee("/cosmos.group.v1.MsgUpdateGroupPolicyMetadata", sdk.NewInt64Coin("banana", 14)),
			types.NewMsgFee("/cosmos.group.v1.MsgVote", sdk.NewInt64Coin("banana", 15)),
			types.NewMsgFee("/cosmos.group.v1.MsgWithdrawProposal", sdk.NewInt64Coin("banana", 16)),
			// Define an expensive one, and a cheaper one.
			types.NewMsgFee("/cosmos.quarantine.v1beta1.MsgOptIn", sdk.NewInt64Coin("banana", 500)),
			types.NewMsgFee("/cosmos.quarantine.v1beta1.MsgOptOut", sdk.NewInt64Coin("banana", 1)),
		)
		// Sort them by MsgTypeURL so that they're in the same order as they will be in state,
		// which makes it easier to identify pagination next keys.
		slices.SortFunc(flatfeeGen.MsgFees, func(a, b *types.MsgFee) int {
			return strings.Compare(a.MsgTypeUrl, b.MsgTypeUrl)
		})
		s.genState = *flatfeeGen
		return flatfeeGen
	})

	s.testnet, err = testnet.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err, "creating testnet")

	s.testnet.Validators[0].ClientCtx = s.testnet.Validators[0].ClientCtx.WithKeyring(s.keyring)
	_, err = testutil.WaitForHeight(s.testnet, 1)
	s.Require().NoError(err, "waiting for height 1")
}

func (s *CLITestSuite) TearDownSuite() {
	testutil.Cleanup(s.testnet, s.T())
}

func (s *CLITestSuite) generateAccountsWithKeyrings(number int) {
	s.keyringEntries, s.keyring = testutil.GenerateTestKeyring(s.T(), number, s.cfg.Codec)
	s.accountAddresses = testutil.GetKeyringEntryAddresses(s.keyringEntries)
}

// asJSONAndYAML returns both the JSON and YAML representations of the provided response.
func (s *CLITestSuite) asJSONAndYAML(resp proto.Message) (string, string) {
	// This reflects how the client.Context.PrintProto method generates JSON and YAML for output.
	jsonBz, err := s.testnet.Validators[0].ClientCtx.Codec.MarshalJSON(resp)
	s.Require().NoError(err, "MarshalJSON(%T)", resp)
	yamlBz, err := yaml.JSONToYAML(jsonBz)
	s.Require().NoError(err, "JSONToYAML(%T)", resp)
	return string(jsonBz) + "\n", string(yamlBz)
}

// asJSON returns the JSON representation of the provided response.
func (s *CLITestSuite) asJSON(resp proto.Message) string {
	// This reflects how the client.Context.PrintProto method generates JSON for output.
	jsonBz, err := s.testnet.Validators[0].ClientCtx.Codec.MarshalJSON(resp)
	s.Require().NoError(err, "MarshalJSON(%T)", resp)
	return string(jsonBz) + "\n"
}

// asYAML returns the YAML representation of the provided response.
func (s *CLITestSuite) asYAML(resp proto.Message) string {
	// This reflects how the client.Context.PrintProto method generates YAML for output.
	jsonBz, err := s.testnet.Validators[0].ClientCtx.Codec.MarshalJSON(resp)
	s.Require().NoError(err, "MarshalJSON(%T)", resp)
	yamlBz, err := yaml.JSONToYAML(jsonBz)
	s.Require().NoError(err, "JSONToYAML(%T)", resp)
	return string(yamlBz)
}

// nextKeyFor will return the --next-key for the genState.MsgFees entry with the given index.
func (s *CLITestSuite) nextKeyFor(i int) string {
	return base64.StdEncoding.EncodeToString([]byte(s.genState.MsgFees[i].MsgTypeUrl))
}

// convertMsgFee converts a single msg fee using genState.Params.
func (s *CLITestSuite) convertMsgFee(msgFee *types.MsgFee) *types.MsgFee {
	return s.genState.Params.ConversionFactor.ConvertMsgFee(msgFee)
}

// convertMsgFee converts a slice of msg fees using genState.Params.
func (s *CLITestSuite) convertMsgFees(msgFees []*types.MsgFee) []*types.MsgFee {
	rv := make([]*types.MsgFee, len(msgFees))
	for i, msgFee := range msgFees {
		rv[i] = s.genState.Params.ConversionFactor.ConvertMsgFee(msgFee)
	}
	return rv
}

// subCommand defines the name and aliases to expect in a sub-command.
type subCommand struct {
	name    string
	aliases []string
}

// assertBaseCmd checks that the provided cmd has the correct name and aliases and the provided subCmds.
func (s *CLITestSuite) assertBaseCmd(cmdMaker func() *cobra.Command, subCmds []subCommand) bool {
	s.T().Helper()
	var cmd *cobra.Command
	ok := s.Run("make the cmd", func() {
		testFunc := func() {
			cmd = cmdMaker()
		}
		s.Require().NotPanics(testFunc)
	})

	ok = s.Run("base command", func() {
		if cmd == nil {
			s.T().Skipf("command maker failed")
		}
		s.Assert().Equal("flatfees", cmd.Name(), "cmd.Name()")
		s.Assert().Equal("flatfees", cmd.Use, "cmd.Use")
		s.Assert().ElementsMatch([]string{"fees", "ff"}, cmd.Aliases, "cmd.Aliases")
		s.Assert().Equal(len(subCmds), len(cmd.Commands()), "len(cmd.Commands())")
	}) && ok

	for _, tc := range subCmds {
		ok = s.Run(fmt.Sprintf("sub-command %s", tc.name), func() {
			if cmd == nil {
				s.T().Skipf("command maker failed")
			}
			var subCmd *cobra.Command
			for _, subCmd = range cmd.Commands() {
				if subCmd.Name() == tc.name {
					break
				}
			}
			s.Assert().NotNil(subCmd, "Could not find the %q sub-command under %q", tc.name, cmd.Name())
			s.Assert().ElementsMatch(tc.aliases, subCmd.Aliases, "%q sub-command aliases", tc.name)
		}) && ok
	}

	return ok
}

// reversed returns a copy of this slice with the entries reversed.
func reversed[S ~[]E, E any](s S) S {
	rv := make(S, len(s))
	for i, val := range s {
		rv[len(s)-i-1] = val
	}
	return rv
}

// generateAndSignTx will call the provided txCmd to generate the tx, then it's signed.
// Returns the path to the signed tx file.
// The --yes (FlagSkipConfirmation) and --generate-only (FlagGenerateOnly) flags are added if not already
// present in args. Also, if the args don't have a --from flag, --from s.accountAddresses[0] is added.
func (s *CLITestSuite) generateAndSignTx(tmpDir string, baseFilename string, txCmd *cobra.Command, args ...string) string {
	unsignedTxFilename, from := s.generateUnsignedTx(tmpDir, baseFilename, txCmd, args...)
	return s.signTx(unsignedTxFilename, from)
}

// generateUnsignedTx will call the provided txCmd to generate a tx.
// The --yes (FlagSkipConfirmation) and --generate-only (FlagGenerateOnly) flags are added if not already
// present in args. Also, if the args don't have a --from flag, --from s.accountAddresses[0] is added.
// Returns the path to the unsigned tx file and the --from value.
func (s *CLITestSuite) generateUnsignedTx(tmpDir string, baseFilename string, txCmd *cobra.Command, args ...string) (unsignedTxFilename string, from string) {
	var haveSkipConf, haveGenOnly bool
	for i, arg := range args {
		switch {
		case arg == flags.FlagSkipConfirmation || strings.HasPrefix(arg, flags.FlagSkipConfirmation+"="):
			haveSkipConf = true
		case arg == flags.FlagGenerateOnly || strings.HasPrefix(arg, flags.FlagGenerateOnly+"="):
			haveGenOnly = true
		case arg == flags.FlagFrom && i+1 < len(args):
			from = args[i+1]
		case strings.HasPrefix(arg, flags.FlagFrom+"="):
			from = strings.TrimPrefix(arg, flags.FlagFrom+"=")
		}
	}

	if !haveSkipConf {
		args = append(args, "--"+flags.FlagSkipConfirmation)
	}
	if !haveGenOnly {
		args = append(args, "--"+flags.FlagGenerateOnly)
	}
	if len(from) == 0 {
		from = s.accountAddresses[0].String()
		args = append(args, "--"+flags.FlagFrom, from)
	}

	// Generate the unsigned Tx.
	clientCtx := s.testnet.Validators[0].ClientCtx
	out, err := clitestutil.ExecTestCLICmd(clientCtx, txCmd, args)
	s.Require().NoError(err, "ExecTestCLICmd(%s, %q)", txCmd.Name(), args)
	unsignedTxBz := out.Bytes()

	// Save the unsigned Tx.
	unsignedTxFilename = filepath.Join(tmpDir, fmt.Sprintf("%s.unsigned.json", baseFilename))
	err = os.WriteFile(unsignedTxFilename, unsignedTxBz, 0o666)
	s.Require().NoError(err, "WriteFile(%q, ...): unsigned tx json", unsignedTxFilename)

	return unsignedTxFilename, from
}

// signTx signs a tx using the provided from account. Returns the path to the signed tx file.
func (s *CLITestSuite) signTx(unsignedTxFilename, from string) string {
	// Generate the signed Tx.
	clientCtx := s.testnet.Validators[0].ClientCtx
	signCmd := authcli.GetSignCommand()
	signArgs := []string{
		unsignedTxFilename,
		"--chain-id", clientCtx.ChainID,
		"--" + flags.FlagFrom, from,
	}
	out, err := clitestutil.ExecTestCLICmd(clientCtx, signCmd, signArgs)
	s.Require().NoError(err, "Signing tx using ExecTestCLICmd(%s, %q)", signCmd.Name(), signArgs)
	signedTxBz := out.Bytes()

	// Save the signed Tx.
	signedTxFilename := unsignedTxFilename + ".signed"
	err = os.WriteFile(signedTxFilename, signedTxBz, 0o666)
	s.Require().NoError(err, "WriteFile(%q, ...): signed tx", signedTxFilename)
	return signedTxFilename
}

func (s *CLITestSuite) TestNewTxCmd() {
	// These tests are only for making sure the sub commands are all added and named/aliased as expected.
	subCmds := []subCommand{
		{name: "update", aliases: []string{"costs"}},
		{name: "params"},
	}
	s.assertBaseCmd(cli.NewTxCmd, subCmds)
}

func (s *CLITestSuite) TestNewCmdUpdateParams() {
	tests := []testcli.TxExecutor{
		{
			Name:      "zero args",
			Args:      []string{},
			ExpErrMsg: "accepts 2 arg(s), received 0",
		},
		{
			Name:      "one arg",
			Args:      []string{"15banana"},
			ExpErrMsg: "accepts 2 arg(s), received 1",
		},
		{
			Name:      "three args",
			Args:      []string{"15banana", "1banana", "3stake"},
			ExpErrMsg: "accepts 2 arg(s), received 3",
		},
		{
			Name: "wrong authority",
			Args: []string{"15banana", "1banana=3stake",
				"--authority", s.account2Addr.String(),
			},
			ExpCode:     13,
			ExpInRawLog: []string{s.account2Addr.String(), "expected gov account as only signer for proposal message"},
		},
		{
			Name:      "invalid default cost",
			Args:      []string{"15x", "1banana=3stake"},
			ExpErrMsg: "invalid default cost \"15x\": invalid decimal coin expression: 15x",
		},
		{
			Name:      "invalid conversion factor",
			Args:      []string{"15banana", "1banana==3stake"},
			ExpErrMsg: "invalid conversion factor \"1banana==3stake\": expected exactly one equals sign",
		},
		{
			Name:    "all good",
			Args:    []string{"15banana", "1banana=3stake"},
			ExpCode: 0,
		},
	}

	for _, tc := range tests {
		s.Run(tc.Name, func() {
			if tc.Cmd == nil {
				tc.Cmd = cli.NewCmdUpdateParams()
			}
			tc.Args = append(tc.Args,
				"--title", "Update msg fees", "--summary", "Updates the MsgFees.",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddresses[0].String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 5)).String()),
				fmt.Sprintf("--%s=json", cmtcli.OutputFlag),
			)

			tc.Execute(s.T(), s.testnet)
		})
	}
}

func (s *CLITestSuite) TestNewCmdUpdateMsgFees() {
	tests := []testcli.TxExecutor{
		{
			Name:      "no options",
			Args:      nil,
			ExpErrMsg: "at least one entry to set or unset must be provided: empty request",
		},
		{
			Name: "wrong authority",
			Args: []string{
				"--authority", s.account2Addr.String(),
				"--unset", "/provenance.name.v1.MsgBindNameRequest",
			},
			ExpCode:     13,
			ExpInRawLog: []string{s.account2Addr.String(), "expected gov account as only signer for proposal message"},
		},
		{
			Name:      "invalid set format",
			Args:      []string{"--set", "/provenance.name.v1.MsgBindNameRequest==4banana"},
			ExpErrMsg: "invalid set arg \"/provenance.name.v1.MsgBindNameRequest==4banana\": expected exactly one equals sign",
		},
		{
			Name:      "set unknown type url",
			Args:      []string{"--set", "/provenance.name.v1.MsgBindNameRequest2=4banana"},
			ExpErrMsg: "unable to resolve type URL /provenance.name.v1.MsgBindNameRequest2",
		},
		{
			Name:      "unset unknown type url",
			Args:      []string{"--unset", "/provenance.name.v1.MsgBindNameRequest2"},
			ExpErrMsg: "unable to resolve type URL /provenance.name.v1.MsgBindNameRequest2",
		},
		{
			Name:    "one to set",
			Args:    []string{"--set", "/provenance.name.v1.MsgBindNameRequest=4banana"},
			ExpCode: 0,
		},
		{
			Name:    "one to unset",
			Args:    []string{"--unset", "/provenance.name.v1.MsgBindNameRequest"},
			ExpCode: 0,
		},
		{
			Name: "one of each",
			Args: []string{
				"--set", "/provenance.name.v1.MsgModifyNameRequest=5banana,1stake",
				"--unset", "/provenance.name.v1.MsgBindNameRequest",
			},
			ExpCode: 0,
		},
		{
			Name: "three to set",
			Args: []string{
				"--set", "/provenance.name.v1.MsgModifyNameRequest=5banana,1stake",
				"--set", "/provenance.name.v1.MsgBindNameRequest=4banana",
				"--set", "/provenance.name.v1.MsgDeleteNameRequest=",
			},
			ExpCode: 0,
		},
		{
			Name: "three to unset",
			Args: []string{
				"--unset", "/provenance.name.v1.MsgModifyNameRequest",
				"--unset", "/provenance.name.v1.MsgBindNameRequest",
				"--unset", "/provenance.name.v1.MsgDeleteNameRequest",
			},
			ExpCode: 0,
		},
		{
			Name: "two of each",
			Args: []string{
				"--set", "/provenance.name.v1.MsgBindNameRequest=9banana",
				"--unset", "/provenance.name.v1.MsgModifyNameRequest",
				"--unset", "/provenance.name.v1.MsgDeleteNameRequest",
				"--set", "/provenance.attribute.v1.MsgAddAttributeRequest=8banana",
			},
			ExpCode: 0,
		},
	}

	for _, tc := range tests {
		s.Run(tc.Name, func() {
			if tc.Cmd == nil {
				tc.Cmd = cli.NewCmdUpdateMsgFees()
			}
			tc.Args = append(tc.Args,
				"--title", "Update msg fees", "--summary", "Updates the MsgFees.",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddresses[0].String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 5)).String()),
				fmt.Sprintf("--%s=json", cmtcli.OutputFlag),
			)

			tc.Execute(s.T(), s.testnet)
		})
	}
}

func (s *CLITestSuite) TestNewQueryCmd() {
	// These tests are only for making sure the sub commands are all added and named/aliased as expected.
	subCmds := []subCommand{
		{name: "params"},
		{name: "list", aliases: []string{"ls", "l", "all"}},
		{name: "get", aliases: []string{"msgfee", "fee"}},
		{name: "calculate-tx-fees", aliases: []string{"simulate", "calculate-fees", "sim", "calc"}},
	}
	s.assertBaseCmd(cli.NewQueryCmd, subCmds)
}

func (s *CLITestSuite) TestNewCmdGetParams() {
	expResp := &types.QueryParamsResponse{
		Params: s.genState.Params,
	}
	expJSON, expYAML := s.asJSONAndYAML(expResp)

	tests := []testcli.QueryExecutor{
		{
			Name:      "one arg",
			Args:      []string{"oops"},
			ExpErrMsg: "unknown command \"oops\" for \"params\"",
		},
		{
			Name:   "json",
			Args:   []string{"--output", "json"},
			ExpOut: expJSON,
		},
		{
			Name:   "yaml",
			Args:   []string{"--output", "text"},
			ExpOut: expYAML,
		},
	}

	for _, tc := range tests {
		s.Run(tc.Name, func() {
			if tc.Cmd == nil {
				tc.Cmd = cli.NewCmdGetParams()
			}
			tc.Execute(s.T(), s.testnet)
		})
	}
}

func (s *CLITestSuite) TestNewCmdGetAllMsgFees() {
	tests := []testcli.QueryExecutor{
		{
			Name:      "one arg",
			Args:      []string{"oops"},
			ExpErrMsg: "unknown command \"oops\" for \"list\"",
		},
		{
			Name:      "invalid pagination flag",
			Args:      []string{"--limit", "a"},
			ExpErrMsg: "invalid argument \"a\" for \"--limit\" flag: strconv.ParseUint: parsing \"a\": invalid syntax",
		},
		{
			Name: "limit 1 with count",
			Args: []string{"--limit", "1", "--count-total", "--output", "json"},
			ExpOut: s.asJSON(&types.QueryAllMsgFeesResponse{
				MsgFees: s.convertMsgFees(s.genState.MsgFees[0:1]),
				Pagination: &query.PageResponse{
					NextKey: []byte(s.genState.MsgFees[1].MsgTypeUrl),
					Total:   uint64(len(s.genState.MsgFees)),
				},
			}),
		},
		{
			Name: "limit 3 with next key",
			Args: []string{"--limit", "3", "--page-key", s.nextKeyFor(5), "--output", "json"},
			ExpOut: s.asJSON(&types.QueryAllMsgFeesResponse{
				MsgFees: s.convertMsgFees(s.genState.MsgFees[5:8]),
				Pagination: &query.PageResponse{
					NextKey: []byte(s.genState.MsgFees[8].MsgTypeUrl),
				},
			}),
		},
		{
			Name: "limit 3 with next key and no conversion",
			Args: []string{"--limit", "3", "--page-key", s.nextKeyFor(5), "--do-not-convert", "--output", "text"},
			ExpOut: s.asYAML(&types.QueryAllMsgFeesResponse{
				MsgFees: s.genState.MsgFees[5:8],
				Pagination: &query.PageResponse{
					NextKey: []byte(s.genState.MsgFees[8].MsgTypeUrl),
				},
			}),
		},
		{
			Name: "limit 3 with offset",
			Args: []string{"--limit", "3", "--offset", "1", "--output", "text"},
			ExpOut: s.asYAML(&types.QueryAllMsgFeesResponse{
				MsgFees: s.convertMsgFees(s.genState.MsgFees[1:4]),
				Pagination: &query.PageResponse{
					NextKey: []byte(s.genState.MsgFees[4].MsgTypeUrl),
				},
			}),
		},
		{
			Name: "limit 4 reversed",
			Args: []string{"--limit", "4", "--reverse", "--output", "json"},
			ExpOut: s.asJSON(&types.QueryAllMsgFeesResponse{
				MsgFees: s.convertMsgFees(reversed(s.genState.MsgFees)[0:4]),
				Pagination: &query.PageResponse{
					NextKey: []byte(s.genState.MsgFees[len(s.genState.MsgFees)-5].MsgTypeUrl),
				},
			}),
		},
		{
			Name: "limit 5 reversed with next key and no conversion",
			Args: []string{"--limit", "5", "--page-key", s.nextKeyFor(6),
				"--do-not-convert", "--reverse", "--output", "text"},
			ExpOut: s.asYAML(&types.QueryAllMsgFeesResponse{
				MsgFees: reversed(s.genState.MsgFees[2:7]),
				Pagination: &query.PageResponse{
					NextKey: []byte(s.genState.MsgFees[1].MsgTypeUrl),
				},
			}),
		},
	}

	for _, tc := range tests {
		s.Run(tc.Name, func() {
			if tc.Cmd == nil {
				tc.Cmd = cli.NewCmdGetAllMsgFees()
			}
			tc.Execute(s.T(), s.testnet)
		})
	}
}

func (s *CLITestSuite) TestNewCmdGetMsgFee() {
	tests := []testcli.QueryExecutor{
		{
			Name:      "no args",
			Args:      []string{},
			ExpErrMsg: "accepts 1 arg(s), received 0",
		},
		{
			Name:      "two args",
			Args:      []string{"oops", "nope"},
			ExpErrMsg: "accepts 1 arg(s), received 2",
		},
		{
			Name: "unknown url",
			Args: []string{"/not.an.actual.Msg"},
			ExpInErrMsg: []string{
				"failed to query msg fee for \"/not.an.actual.Msg\"",
				"code = InvalidArgument",
				"desc = unknown msg type url \"/not.an.actual.Msg\"",
				"invalid request",
			},
			ExpNotInOut: []string{"Usage:", "Aliases:", "Examples:", "Flags:"},
		},
		{
			// If this test fails because there's an amount, switch the index from 0 to one without a cost.
			Name:   "json: free entry",
			Args:   []string{s.genState.MsgFees[0].MsgTypeUrl, "--output", "json"},
			ExpOut: s.asJSON(&types.QueryMsgFeeResponse{MsgFee: s.genState.MsgFees[0]}),
		},
		{
			// If this test fails because there's an amount, switch the index from 0 to one without a cost.
			Name:   "yaml: free entry",
			Args:   []string{s.genState.MsgFees[0].MsgTypeUrl, "--output", "text"},
			ExpOut: s.asYAML(&types.QueryMsgFeeResponse{MsgFee: s.genState.MsgFees[0]}),
		},
		{
			// If this test fails because there's an amount, switch the index from 0 to one without a cost.
			Name:   "free entry without conversion",
			Args:   []string{s.genState.MsgFees[0].MsgTypeUrl, "--output", "json", "--do-not-convert"},
			ExpOut: s.asJSON(&types.QueryMsgFeeResponse{MsgFee: s.genState.MsgFees[0]}),
		},
		{
			Name:   "json: with cost",
			Args:   []string{s.genState.MsgFees[1].MsgTypeUrl, "--output", "json"},
			ExpOut: s.asJSON(&types.QueryMsgFeeResponse{MsgFee: s.convertMsgFee(s.genState.MsgFees[1])}),
		},
		{
			Name:   "yaml: with cost",
			Args:   []string{s.genState.MsgFees[1].MsgTypeUrl, "--output", "text"},
			ExpOut: s.asYAML(&types.QueryMsgFeeResponse{MsgFee: s.convertMsgFee(s.genState.MsgFees[1])}),
		},
		{
			Name:   "json: with cost and no conversion",
			Args:   []string{s.genState.MsgFees[2].MsgTypeUrl, "--output", "json", "--do-not-convert"},
			ExpOut: s.asJSON(&types.QueryMsgFeeResponse{MsgFee: s.genState.MsgFees[2]}),
		},
		{
			Name:   "yaml: with cost and no conversion",
			Args:   []string{s.genState.MsgFees[2].MsgTypeUrl, "--output", "text", "--do-not-convert"},
			ExpOut: s.asYAML(&types.QueryMsgFeeResponse{MsgFee: s.genState.MsgFees[2]}),
		},
		{
			Name: "msg type that uses the default",
			Args: []string{"/cosmos.gov.v1.MsgVote", "--output", "text"},
			ExpOut: s.asYAML(&types.QueryMsgFeeResponse{
				MsgFee: &types.MsgFee{
					MsgTypeUrl: "/cosmos.gov.v1.MsgVote",
					Cost:       sdk.Coins{s.genState.Params.ConversionFactor.ConvertCoin(s.genState.Params.DefaultCost)},
				},
			}),
		},
	}

	for _, tc := range tests {
		s.Run(tc.Name, func() {
			if tc.Cmd == nil {
				tc.Cmd = cli.NewCmdGetMsgFee()
			}
			tc.Execute(s.T(), s.testnet)
		})
	}
}

func (s *CLITestSuite) TestNewCmdCalculateTxFees() {
	tmpDir := s.T().TempDir()
	addrCdc := s.cfg.Codec.InterfaceRegistry().SigningContext().AddressCodec()

	bankSendTx := s.generateAndSignTx(tmpDir, "send", bankcli.NewSendTxCmd(addrCdc),
		s.accountAddresses[0].String(), s.accountAddresses[1].String(), fmt.Sprintf("%d%s", 3, s.cfg.BondDenom),
	)

	upgradePropTx := s.generateAndSignTx(tmpDir, "upgrade", upgradecli.NewCmdSubmitUpgradeProposal(addrCdc),
		"pink",
		"--"+upgradecli.FlagUpgradeHeight, "5000000",
		"--"+upgradecli.FlagNoValidate,
		"--"+govcli.FlagTitle, "The Pink Upgrade",
		"--"+govcli.FlagSummary, "The Pink Upgrade for a new version.",
	)

	optInTx := s.generateAndSignTx(tmpDir, "quar-opt-in", quarantinecli.TxOptInCmd(), s.accountAddresses[0].String())
	optOutTx := s.generateAndSignTx(tmpDir, "quar-opt-out", quarantinecli.TxOptOutCmd(), s.accountAddresses[0].String())

	tests := []testcli.QueryExecutor{
		{
			Name: "default cost, no multiplier, default output",
			Args: []string{bankSendTx},
			// This gas estimate keeps changing between 83359 and 83362.
			// It probably has something to do with whether any of the tx tests have run and updated some state.
			// We don't care about the actual number, though, just that it's not zero.
			// So we'll just check that the field starts with 83.
			ExpInOut: []string{`"total_fees":[{"denom":"stake","amount":"5"}]`, `"estimated_gas":"83`},
		},
		{
			Name:     "default cost, no multiplier, json output",
			Args:     []string{bankSendTx, "--output", "json"},
			ExpInOut: []string{`"total_fees":[{"denom":"stake","amount":"5"}]`, `"estimated_gas":"83`},
		},
		{
			Name:     "default cost, no multiplier, text output",
			Args:     []string{bankSendTx, "--output", "text"},
			ExpInOut: []string{"total_fees:\n- amount: \"5\"\n  denom: stake\n", "estimated_gas: \"83"},
		},
		{
			Name:     "default cost, 1.5 multiplier",
			Args:     []string{bankSendTx, "--gas-adjustment", "1.7", "--output", "json"}, // 83,3xx * 1.7 = 14y,yyy
			ExpInOut: []string{`"total_fees":[{"denom":"stake","amount":"5"}]`, `"estimated_gas":"14`},
		},
		{
			Name:     "default cost, 5.3 multiplier",
			Args:     []string{bankSendTx, "--gas-adjustment", "5.3", "--output", "json"}, // 83,3xx * 5.3 = 44y,yyy.
			ExpInOut: []string{`"total_fees":[{"denom":"stake","amount":"5"}]`, `"estimated_gas":"44`},
		},
		{
			Name: "free msg",
			Args: []string{upgradePropTx, "--output", "json"},
			// This gas estimate keeps changing between 127267 and 127585.
			// With this test, I want to make sure that gas is still being estimated, even if the fee is zero.
			// But I don't care what the estimate actually is, so I'm just going to make sure it starts with a 1.
			ExpInOut: []string{`"total_fees":[]`, `"estimated_gas":"1`},
		},
		{
			Name:     "larger cost: quarantine opt in",
			Args:     []string{optInTx, "--output", "json"}, // 500 banana * 1 stake / 2 banana = 250 stake
			ExpInOut: []string{`"total_fees":[{"denom":"stake","amount":"250"}]`},
		},
		{
			Name:        "make sure the previous quarantine opt in wasn't actually applied",
			Cmd:         quarantinecli.QueryIsQuarantinedCmd(),
			Args:        []string{s.accountAddresses[0].String(), "--output", "text"},
			ExpInOut:    []string{"is_quarantined: false"},
			ExpNotInOut: []string{"is_quarantined: true"},
		},
		{
			Name:     "cheaper cost",
			Args:     []string{optOutTx, "--output", "json"}, // 1 banana * 1 stake / 2 banana = 0.5 stake => 1 stake
			ExpInOut: []string{`"total_fees":[{"denom":"stake","amount":"1"}]`},
		},
	}

	for _, tc := range tests {
		s.Run(tc.Name, func() {
			if tc.Cmd == nil {
				tc.Cmd = cli.NewCmdCalculateTxFees()
			}
			tc.Execute(s.T(), s.testnet)
		})
	}
}
