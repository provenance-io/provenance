package cmd_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/cmd/provenanced/cmd"
	provconfig "github.com/provenance-io/provenance/cmd/provenanced/config"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltest "github.com/cosmos/cosmos-sdk/x/genutil/client/testutil"
)

type ConfigTestSuite struct {
	suite.Suite

	Home    string
	Context *context.Context

	HeaderStrApp    string
	HeaderStrTM     string
	HeaderStrClient string

	BaseFNApp    string
	BaseFNTM     string
	baseFNClient string
}

func (s *ConfigTestSuite) SetupTest() {
	s.Home = s.T().TempDir()
	tmConfig, err := genutiltest.CreateDefaultTendermintConfig(s.Home)
	s.Require().NoError(err, "creating default tendermint config")

	tMbm := module.NewBasicManager(genutil.AppModuleBasic{})
	appCodec := simapp.MakeTestEncodingConfig().Marshaler
	err = genutiltest.ExecInitCmd(tMbm, s.Home, appCodec)
	s.Require().NoError(err, "unexpected error calling genutiltest init command")

	logger := log.NewNopLogger()
	serverCtx := server.NewContext(viper.New(), tmConfig, logger)
	serverCtx.Viper.Set(server.FlagMinGasPrices, fmt.Sprintf("1905%s", app.DefaultFeeDenom))

	clientCtx := client.Context{}.WithCodec(appCodec).WithHomeDir(s.Home).WithViper("")
	clientCtx, err = provconfig.ReadFromClientConfig(clientCtx)
	s.Require().NoError(err, "setting up client context")

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)
	s.Context = &ctx

	s.HeaderStrApp = "App"
	s.HeaderStrTM = "Tendermint"
	s.HeaderStrClient = "Client"

	s.BaseFNApp = "app.toml"
	s.BaseFNTM = "config.toml"
	s.baseFNClient = "client.toml"
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

//
// Test setup above. Test helpers below.
//

func (s ConfigTestSuite) makeConfigHeaderLine(t, fn string) string {
	return fmt.Sprintf("%s Config: %s/config/%s", t, s.Home, fn)
}

func (s ConfigTestSuite) makeAppConfigHeaderLines() string {
	return s.makeConfigHeaderLine(s.HeaderStrApp, s.BaseFNApp) + "\n----------------"
}

func (s ConfigTestSuite) makeTMConfigHeaderLines() string {
	return s.makeConfigHeaderLine(s.HeaderStrTM, s.BaseFNTM) + "\n-----------------------"
}

func (s ConfigTestSuite) makeClientConfigHeaderLines() string {
	return s.makeConfigHeaderLine(s.HeaderStrClient, s.baseFNClient) + "\n-------------------"
}

func (s ConfigTestSuite) makeConfigUpdatedLine(t, fn string) string {
	return fmt.Sprintf("%s Config Updated: %s/config/%s", t, s.Home, fn)
}

func (s ConfigTestSuite) makeAppConfigUpdateLines() string {
	return s.makeConfigUpdatedLine(s.HeaderStrApp, s.BaseFNApp) + "\n------------------------"
}

func (s ConfigTestSuite) makeTMConfigUpdateLines() string {
	return s.makeConfigUpdatedLine(s.HeaderStrTM, s.BaseFNTM) + "\n-------------------------------"
}

func (s ConfigTestSuite) makeClientConfigUpdateLines() string {
	return s.makeConfigUpdatedLine(s.HeaderStrClient, s.baseFNClient) + "\n---------------------------"
}

func (s ConfigTestSuite) makeConfigDiffHeaderLine(t, fn string) string {
	return fmt.Sprintf("%s Config Differences from Defaults: %s/config/%s", t, s.Home, fn)
}

func (s ConfigTestSuite) makeAppDiffHeaderLines() string {
	return s.makeConfigDiffHeaderLine(s.HeaderStrApp, s.BaseFNApp) + "\n------------------------------------------"
}

func (s ConfigTestSuite) makeTMDiffHeaderLines() string {
	return s.makeConfigDiffHeaderLine(s.HeaderStrTM, s.BaseFNTM) + "\n-------------------------------------------------"
}

func (s ConfigTestSuite) makeClientDiffHeaderLines() string {
	return s.makeConfigDiffHeaderLine(s.HeaderStrClient, s.baseFNClient) + "\n---------------------------------------------"
}

func (s ConfigTestSuite) makeMultiLine(lines ...string) string {
	return strings.Join(lines, "\n") + "\n"
}

func (s ConfigTestSuite) makeKeyUpdatedLine(key, oldVal, newVal string) string {
	return fmt.Sprintf("%s Was: %s, Is Now: %s", key, oldVal, newVal)
}

//
// Test helpers above. Tests below.
//

func (s *ConfigTestSuite) TestConfigBadArgs() {
	tests := []struct {
		name string
		args []string
		err  string
	}{
		{
			name: "set with nothing else",
			args: []string{"set"},
			err:  "no key/value pairs provided",
		},
		{
			name: "set with odd args",
			args: []string{"set", "output", "text", "banana"},
			err:  "an even number of arguments are required when setting values",
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			b := bytes.NewBufferString("")
			command := cmd.ConfigCmd()
			command.SetArgs(tc.args)
			command.SetOut(b)
			err := command.ExecuteContext(*s.Context)
			require.EqualError(t, err, tc.err, "%s %s expected error executing command", command.Name(), tc.args)
		})
	}
}

func (s *ConfigTestSuite) TestConfigCmdGet() {
	command := cmd.ConfigCmd()

	command.SetArgs([]string{"get"})
	b := bytes.NewBufferString("")
	command.SetOut(b)
	err := command.ExecuteContext(*s.Context)
	s.Require().NoError(err, "%s - unexpected error executing command", command.Name())
	out, err := ioutil.ReadAll(b)
	s.Require().NoError(err, "%s - unexpected error reading command output", command.Name())
	outStr := string(out)

	// Only spot-checking instead of checking the entire 100+ lines exactly because I don't want this to needlessly break
	// if comsos or tendermint adds a new configuration field/section.
	expectedRegexpMatches := []struct {
		name string
		re   *regexp.Regexp
	}{
		// App config header and a few entries.
		{"app header", regexp.MustCompile(`(?m)^App Config: .*/config/` + s.BaseFNApp + `$`)},
		{"app halt-height", regexp.MustCompile(`(?m)^halt-height=0$`)},
		{"app api.swagger", regexp.MustCompile(`(?m)^api.swagger=false$`)},
		{"app grpc.address", regexp.MustCompile(`(?m)^grpc.address="0.0.0.0:9090"$`)},
		{"app telemetry.enabled", regexp.MustCompile(`(?m)^telemetry.enabled=false$`)},
		{"app rosetta.enable", regexp.MustCompile(`(?m)^rosetta.enable=false$`)},

		// Tendermint header and a few entries.
		{"tendermint header", regexp.MustCompile(`(?m)^Tendermint Config: .*/config/` + s.BaseFNTM + `$`)},
		{"tendermint fast_sync", regexp.MustCompile(`(?m)^fast_sync=true$`)},
		{"tendermint consensus.timeout_commit", regexp.MustCompile(`(?m)^consensus.timeout_commit="1s"$`)},
		{"tendermint mempool.size", regexp.MustCompile(`(?m)^mempool.size=5000$`)},
		{"tendermint statesync.trust_period", regexp.MustCompile(`(?m)^statesync.trust_period="168h0m0s"$`)},
		{"tendermint p2p.recv_rate", regexp.MustCompile(`(?m)^p2p.recv_rate=5120000$`)},

		// Client config header all the entries.
		{"client header", regexp.MustCompile(`(?m)^Client Config: .*/config/` + s.baseFNClient + `$`)},
		{"client broadcast-mode", regexp.MustCompile(`(?m)^broadcast-mode="block"$`)},
		{"client chain-id", regexp.MustCompile(`(?m)^chain-id=""$`)},
		{"client keyring-backend", regexp.MustCompile(`(?m)^keyring-backend="test"$`)},
		{"client node", regexp.MustCompile(`(?m)^node="tcp://localhost:26657"$`)},
		{"client output", regexp.MustCompile(`(?m)^output="text"$`)},
	}

	for _, tc := range expectedRegexpMatches {
		s.T().Run(tc.name, func(t *testing.T) {
			isMatch := tc.re.Match(out)
			assert.True(t, isMatch, "`%s` matching:\n%s", tc.re.String(), out)
		})
	}

	s.T().Run("with args get all", func(t *testing.T) {
		command2 := cmd.ConfigCmd()
		args := []string{"get", "all"}
		command2.SetArgs(args)
		b2 := bytes.NewBufferString("")
		command2.SetOut(b2)
		err2 := command2.ExecuteContext(*s.Context)
		require.NoError(t, err2, "%s - unexpected error executing command", command2.Name())
		out2, err2 := ioutil.ReadAll(b2)
		require.NoError(t, err2, "%s - unexpected error reading command output", command2.Name())
		out2Str := string(out2)
		require.Equal(t, outStr, out2Str, "output of no-args vs output of %s", args)
	})

	inOutTests := []struct {
		name string
		args []string
	}{
		{
			name: "get app",
			args: []string{"get", "app"},
		},
		{
			name: "get cosmos",
			args: []string{"get", "cosmos"},
		},
		{
			name: "get config",
			args: []string{"get", "config"},
		},
		{
			name: "get tendermint",
			args: []string{"get", "tendermint"},
		},
		{
			name: "get tm",
			args: []string{"get", "tm"},
		},
		{
			name: "get client",
			args: []string{"get", "client"},
		},
	}

	for _, tc := range inOutTests {
		s.T().Run(tc.name, func(t *testing.T) {
			command2 := cmd.ConfigCmd()
			command2.SetArgs(tc.args)
			b2 := bytes.NewBufferString("")
			command2.SetOut(b2)
			err2 := command2.ExecuteContext(*s.Context)
			require.NoError(t, err2, "%s - unexpected error executing command", command2.Name())
			out2, err2 := ioutil.ReadAll(b2)
			require.NoError(t, err2, "%s - unexpected error reading command output", command2.Name())
			out2Str := string(out2)
			require.Contains(t, outStr, out2Str, "output of no-args vs output of %s", tc.args)
		})
	}
}

func (s *ConfigTestSuite) TestConfigGetMulti() {
	tests := []struct {
		name     string
		keys     []string
		expected string
	}{
		{
			name: "three app config keys",
			keys: []string{"min-retain-blocks", "rosetta.retries", "grpc.address"},
			expected: s.makeMultiLine(
				s.makeAppConfigHeaderLines(),
				`min-retain-blocks=0`,
				`grpc.address="0.0.0.0:9090"`,
				`rosetta.retries=3`,
				""),
		},
		{
			name: "three tendermint config keys",
			keys: []string{"p2p.send_rate", "genesis_file", "consensus.timeout_propose"},
			expected: s.makeMultiLine(
				s.makeTMConfigHeaderLines(),
				`genesis_file="config/genesis.json"`,
				`consensus.timeout_propose="3s"`,
				`p2p.send_rate=5120000`,
				""),
		},
		{
			name: "three client config keys",
			keys: []string{"keyring-backend", "broadcast-mode", "output"},
			expected: s.makeMultiLine(
				s.makeClientConfigHeaderLines(),
				`broadcast-mode="block"`,
				`keyring-backend="test"`,
				`output="text"`,
				""),
		},
		{
			name: "two from each",
			keys: []string{"rpc.cors_allowed_origins", "pruning", "node", "rosetta.offline", "chain-id", "priv_validator_state_file"},
			expected: s.makeMultiLine(
				s.makeAppConfigHeaderLines(),
				`pruning="default"`,
				`rosetta.offline=false`,
				"",
				s.makeTMConfigHeaderLines(),
				`priv_validator_state_file="data/priv_validator_state.json"`,
				`rpc.cors_allowed_origins=[]`,
				"",
				s.makeClientConfigHeaderLines(),
				`chain-id=""`,
				`node="tcp://localhost:26657"`,
				""),
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			args := []string{"get"}
			args = append(args, tc.keys...)
			b := bytes.NewBufferString("")

			command := cmd.ConfigCmd()
			command.SetArgs(args)
			command.SetOut(b)
			err := command.ExecuteContext(*s.Context)
			require.NoError(t, err, "%s %s - unexpected error executing command", command.Name(), args)
			out, err := ioutil.ReadAll(b)
			require.NoError(t, err, "%s %s - unexpected error reading command output", command.Name(), args)
			outStr := string(out)
			assert.Equal(t, tc.expected, outStr, "%s %s - output", command.Name(), args)
		})
	}

	s.T().Run("three found two missing", func(t *testing.T) {
		expectedError := "2 configuration keys not found: bananas, pears"
		expected := s.makeMultiLine(
			s.makeAppConfigHeaderLines(),
			`api.enable=false`,
			`api.swagger=false`,
			"",
			s.makeClientConfigHeaderLines(),
			`output="text"`,
			"",
		) + "Error: " + expectedError + "\n"
		args := []string{"get", "bananas", "api.enable", "pears", "api.swagger", "output"}
		b := bytes.NewBufferString("")

		command := cmd.ConfigCmd()
		command.SetArgs(args)
		command.SetOut(b)
		err := command.ExecuteContext(*s.Context)
		require.NoError(t, err, "%s %s - expected error executing command", command.Name(), args)
		out, err := ioutil.ReadAll(b)
		require.NoError(t, err, "%s %s - unexpected error reading command output", command.Name(), args)
		outStr := string(out)
		assert.Equal(t, expected, outStr, "%s %s - output", command.Name(), args)
	})

	s.T().Run("two found one missing", func(t *testing.T) {
		expectedError := "1 configuration key not found: cannot.find.me"
		expected := s.makeMultiLine(
			s.makeAppConfigHeaderLines(),
			`grpc.enable=true`,
			"",
			s.makeTMConfigHeaderLines(),
			`consensus.create_empty_blocks_interval="0s"`,
			"",
		) + "Error: " + expectedError + "\n"
		args := []string{"get", "cannot.find.me", "consensus.create_empty_blocks_interval", "grpc.enable"}
		b := bytes.NewBufferString("")

		command := cmd.ConfigCmd()
		command.SetArgs(args)
		command.SetOut(b)
		err := command.ExecuteContext(*s.Context)
		require.NoError(t, err, "%s %s - expected error executing command", command.Name(), args)
		out, err := ioutil.ReadAll(b)
		require.NoError(t, err, "%s %s - unexpected error reading command output", command.Name(), args)
		outStr := string(out)
		assert.Equal(t, expected, outStr, "%s %s - output", command.Name(), args)
	})
}

func (s *ConfigTestSuite) TestConfigChanged() {
	allEqual := func(t string) string {
		return fmt.Sprintf("All %s config values equal the default config values.", t)
	}
	expectedAppOutLines := []string{
		s.makeAppDiffHeaderLines(),
		`minimum-gas-prices="1905nhash" (default="")`,
		"",
	}
	expectedTMOutLines := []string{
		s.makeTMDiffHeaderLines(),
		allEqual("tendermint"),
		"",
	}
	expectedClientOutLines := []string{
		s.makeClientDiffHeaderLines(),
		allEqual("client"),
		"",
	}
	expectedAllOutLines := []string{}
	expectedAllOutLines = append(expectedAllOutLines, expectedAppOutLines...)
	expectedAllOutLines = append(expectedAllOutLines, expectedTMOutLines...)
	expectedAllOutLines = append(expectedAllOutLines, expectedClientOutLines...)
	expectedAppOut := s.makeMultiLine(expectedAppOutLines...)
	expectedTMOut := s.makeMultiLine(expectedTMOutLines...)
	expectedClientOut := s.makeMultiLine(expectedClientOutLines...)
	expectedAll := s.makeMultiLine(expectedAllOutLines...)

	equalAllTests := []struct {
		name string
		args []string
		out  string
	}{
		{"changed", []string{"changed"}, expectedAll},
		{"changed all", []string{"changed", "all"}, expectedAll},
		{"changed app", []string{"changed", "app"}, expectedAppOut},
		{"changed cosmos", []string{"changed", "cosmos"}, expectedAppOut},
		{"changed config", []string{"changed", "config"}, expectedTMOut},
		{"changed tm", []string{"changed", "tm"}, expectedTMOut},
		{"changed tendermint", []string{"changed", "tendermint"}, expectedTMOut},
		{"changed client", []string{"changed", "client"}, expectedClientOut},
		{
			name: "changed output",
			args: []string{"changed", "output"},
			out: s.makeMultiLine(
				s.makeClientDiffHeaderLines(),
				`output="text" (same as default)`,
				"",
			),
		},
	}
	for _, tc := range equalAllTests {
		s.T().Run(tc.name, func(t *testing.T) {
			b := bytes.NewBufferString("")
			command := cmd.ConfigCmd()
			command.SetArgs(tc.args)
			command.SetOut(b)
			err := command.ExecuteContext(*s.Context)
			require.NoError(t, err, "%s %s - unexpected error executing command", command.Name(), tc.args)
			out, err := ioutil.ReadAll(b)
			require.NoError(t, err, "%s %s - unexpected error reading command output", command.Name(), tc.args)
			outStr := string(out)
			assert.Equal(t, tc.out, outStr, "%s %s - output", command.Name(), tc.args)
		})
	}
}

func (s *ConfigTestSuite) TestConfigSetValidation() {
	tests := []struct {
		name string
		args []string
		out  string
	}{
		{
			name: "set app fails validation",
			args: []string{"set", "minimum-gas-prices", ""},
			out:  `App config validation error: set min gas price in app.toml or flag or env variable: error in app.toml [cosmos/cosmos-sdk@v0.43.0/types/errors/errors.go:269]`,
		},
		{
			name: "set tendermint fails validation",
			args: []string{"set", "log_format", "crazy"},
			out:  `Tendermint config validation error: unknown log_format (must be 'plain' or 'json')`,
		},
		{
			name: "set client fails validation",
			args: []string{"set", "output", "csv"},
			out:  `Client config validation error: unknown output (must be 'text' or 'json')`,
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			expected := fmt.Sprintf("%s\n%s\n",
				tc.out,
				"Error: one or more issues encountered; no configuration values have been updated")
			b := bytes.NewBufferString("")
			command := cmd.ConfigCmd()
			command.SetArgs(tc.args)
			command.SetOut(b)
			err := command.ExecuteContext(*s.Context)
			require.NoError(t, err, "%s %s unexpected error executing command", command.Name(), tc.args)
			out, rerr := ioutil.ReadAll(b)
			require.NoError(t, rerr, "%s %s unexpected error reading output", command.Name(), tc.args)
			outStr := string(out)
			assert.Equal(t, expected, outStr, "%s %s output", command.Name(), tc.args)
		})
	}
}

func (s *ConfigTestSuite) TestConfigCmdSet() {
	reAppConfigUpdated := regexp.MustCompile(`(?m)^App Config Updated: .*/config/` + s.BaseFNApp + `$`)
	reTMConfigUpdated := regexp.MustCompile(`(?m)^Tendermint Config Updated: .*/config/` + s.BaseFNTM + `$`)
	reClientConfigUpdated := regexp.MustCompile(`(?m)^Client Config Updated: .*/config/` + s.baseFNClient + `$`)

	positiveTests := []struct {
		name    string
		oldVal  string
		newVal  string
		toMatch []*regexp.Regexp
	}{
		// app fields
		{
			name:    "api.enable",
			oldVal:  `false`,
			newVal:  `true`,
			toMatch: []*regexp.Regexp{reAppConfigUpdated},
		},
		{
			name:    "min-retain-blocks",
			oldVal:  `0`,
			newVal:  `5`,
			toMatch: []*regexp.Regexp{reAppConfigUpdated},
		},
		{
			name:    "api.max-open-connections",
			oldVal:  `1000`,
			newVal:  `999`,
			toMatch: []*regexp.Regexp{reAppConfigUpdated},
		},
		{
			name:    "state-sync.snapshot-keep-recent",
			oldVal:  `2`,
			newVal:  `3`,
			toMatch: []*regexp.Regexp{reAppConfigUpdated},
		},
		{
			name:    "telemetry.service-name",
			oldVal:  `""`,
			newVal:  `"banana"`,
			toMatch: []*regexp.Regexp{reAppConfigUpdated},
		},

		// tendermint fields
		{
			name:    "filter_peers",
			oldVal:  `false`,
			newVal:  `true`,
			toMatch: []*regexp.Regexp{reTMConfigUpdated},
		},
		{
			name:    "proxy_app",
			oldVal:  `"tcp://127.0.0.1:26658"`,
			newVal:  `"tcp://localhost:26658"`,
			toMatch: []*regexp.Regexp{reTMConfigUpdated},
		},
		{
			name:    "consensus.timeout_commit",
			oldVal:  `"1s"`,
			newVal:  `"2s"`,
			toMatch: []*regexp.Regexp{reTMConfigUpdated},
		},
		{
			name:    "mempool.cache_size",
			oldVal:  `10000`,
			newVal:  `10005`,
			toMatch: []*regexp.Regexp{reTMConfigUpdated},
		},
		{
			name:    "rpc.cors_allowed_methods",
			oldVal:  `["HEAD", "GET", "POST"]`,
			newVal:  `["POST", "HEAD", "GET"]`,
			toMatch: []*regexp.Regexp{reTMConfigUpdated},
		},

		// Client fields
		{
			name:    "chain-id",
			oldVal:  `""`,
			newVal:  `"new-chain"`,
			toMatch: []*regexp.Regexp{reClientConfigUpdated},
		},
		{
			name:    "node",
			oldVal:  `"tcp://localhost:26657"`,
			newVal:  `"tcp://127.0.0.1:26657"`,
			toMatch: []*regexp.Regexp{reClientConfigUpdated},
		},
		{
			name:    "output",
			oldVal:  `"text"`,
			newVal:  `"json"`,
			toMatch: []*regexp.Regexp{reClientConfigUpdated},
		},
		{
			name:    "broadcast-mode",
			oldVal:  `"block"`,
			newVal:  `"sync"`,
			toMatch: []*regexp.Regexp{reClientConfigUpdated},
		},
		{
			name:    "keyring-backend",
			oldVal:  `"test"`,
			newVal:  `"os"`,
			toMatch: []*regexp.Regexp{reClientConfigUpdated},
		},
	}

	for _, tc := range positiveTests {
		s.T().Run(tc.name+" (with set arg)", func(t *testing.T) {
			expectedInOut := s.makeKeyUpdatedLine(tc.name, tc.oldVal, tc.newVal)
			args := []string{"set", tc.name, strings.Trim(tc.newVal, "\"")}
			command := cmd.ConfigCmd()
			command.SetArgs(args)
			b := bytes.NewBufferString("")
			command.SetOut(b)
			err := command.ExecuteContext(*s.Context)
			require.NoError(t, err, "%s %s - unexpected error in execution", command.Name(), args)
			out, rerr := ioutil.ReadAll(b)
			require.NoError(t, rerr, "%s %s - unexpected error in reading", command.Name(), args)
			outStr := string(out)
			for _, re := range tc.toMatch {
				isMatch := re.Match(out)
				assert.True(t, isMatch, "`%s` matching:\n%s", re.String(), outStr)
			}
			assert.Contains(t, outStr, expectedInOut, "update line")
		})
	}
}

func (s *ConfigTestSuite) TestConfigSetMulti() {
	tests := []struct {
		name string
		args []string
		out  string
	}{
		{
			name: "two app entries",
			args: []string{"set", "api.enable", "true", "telemetry.service-name", "blocky"},
			out: s.makeMultiLine(
				s.makeAppConfigUpdateLines(),
				s.makeKeyUpdatedLine("api.enable", "false", "true"),
				s.makeKeyUpdatedLine("telemetry.service-name", `""`, `"blocky"`),
				""),
		},
		{
			name: "two tendermint entries",
			args: []string{"set", "log_format", "json", "consensus.timeout_commit", "950ms"},
			out: s.makeMultiLine(
				s.makeTMConfigUpdateLines(),
				s.makeKeyUpdatedLine("log_format", `"plain"`, `"json"`),
				s.makeKeyUpdatedLine("consensus.timeout_commit", `"1s"`, `"950ms"`),
				""),
		},
		{
			name: "two client entries",
			args: []string{"set", "node", "tcp://127.0.0.1:26657", "output", "json"},
			out: s.makeMultiLine(
				s.makeClientConfigUpdateLines(),
				s.makeKeyUpdatedLine("node", `"tcp://localhost:26657"`, `"tcp://127.0.0.1:26657"`),
				s.makeKeyUpdatedLine("output", `"text"`, `"json"`),
				""),
		},
		{
			name: "two of each",
			args: []string{"set", "consensus.timeout_commit", "950ms", "api.enable", "true", "telemetry.service-name", "blocky", "node", "tcp://127.0.0.1:26657", "output", "json", "log_format", "json"},
			out: s.makeMultiLine(
				s.makeAppConfigUpdateLines(),
				s.makeKeyUpdatedLine("api.enable", "false", "true"),
				s.makeKeyUpdatedLine("telemetry.service-name", `""`, `"blocky"`),
				"",
				s.makeTMConfigUpdateLines(),
				s.makeKeyUpdatedLine("log_format", `"plain"`, `"json"`),
				s.makeKeyUpdatedLine("consensus.timeout_commit", `"1s"`, `"950ms"`),
				"",
				s.makeClientConfigUpdateLines(),
				s.makeKeyUpdatedLine("node", `"tcp://localhost:26657"`, `"tcp://127.0.0.1:26657"`),
				s.makeKeyUpdatedLine("output", `"text"`, `"json"`),
				""),
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			b := bytes.NewBufferString("")
			command := cmd.ConfigCmd()
			command.SetArgs(tc.args)
			command.SetOut(b)
			err := command.ExecuteContext(*s.Context)
			require.NoError(t, err, "%s %s - unexpected error in execution", command.Name(), tc.args)
			out, rerr := ioutil.ReadAll(b)
			require.NoError(t, rerr, "%s %s - unexpected error in reading", command.Name(), tc.args)
			outStr := string(out)
			assert.Equal(t, tc.out, outStr, "%s %s output", command.Name(), tc.args)
		})
	}
}
