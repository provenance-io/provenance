package cmd_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"

	"github.com/provenance-io/provenance/app"
	simappparams "github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/cmd/provenanced/cmd"
	provconfig "github.com/provenance-io/provenance/cmd/provenanced/config"
	"github.com/provenance-io/provenance/internal/pioconfig"
)

type ConfigTestSuite struct {
	suite.Suite

	Home          string
	ClientContext *client.Context
	ServerContext *server.Context
	Context       *context.Context

	EncodingConfig simappparams.EncodingConfig

	HeaderStrApp    string
	HeaderStrCMT    string
	HeaderStrClient string

	BaseFNApp    string
	BaseFNCMT    string
	baseFNClient string
}

func (s *ConfigTestSuite) SetupTest() {
	s.Home = s.T().TempDir()
	s.T().Logf("%s Home: %s", s.T().Name(), s.Home)

	pioconfig.SetProvenanceConfig("confcoin", 5)

	s.EncodingConfig = app.MakeTestEncodingConfig(s.T())
	clientCtx := client.Context{}.
		WithCodec(s.EncodingConfig.Marshaler).
		WithHomeDir(s.Home)
	clientCtx.Viper = viper.New()
	serverCtx := server.NewContext(clientCtx.Viper, provconfig.DefaultCmtConfig(), log.NewNopLogger())

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

	s.ClientContext = &clientCtx
	s.ServerContext = serverCtx
	s.Context = &ctx

	s.HeaderStrApp = "App"
	s.HeaderStrCMT = "CometBFT"
	s.HeaderStrClient = "Client"

	s.BaseFNApp = "app.toml"
	s.BaseFNCMT = "config.toml"
	s.baseFNClient = "client.toml"

	s.ensureConfigFiles()
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

//
// Test setup above. Test helpers below.
//

func (s *ConfigTestSuite) getConfigCmd() *cobra.Command {
	// What I really need here is a cobra.Command
	// that already has a context.
	// So I'm going to call --help on the config command
	// while setting the context and getting the command back.
	configCmd := cmd.ConfigCmd()
	configCmd.SetArgs([]string{"--help"})
	applyMockIODiscardOutErr(configCmd)
	configCmd, err := configCmd.ExecuteContextC(*s.Context)
	s.Require().NoError(err, "config help command to set context")
	// Now this should work to load the defaults (or files) into the cmd.
	s.Require().NoError(
		provconfig.LoadConfigFromFiles(configCmd),
		"loading config from files",
	)
	return configCmd
}

func (s *ConfigTestSuite) ensureConfigFiles() {
	configCmd := s.getConfigCmd()
	// Extract the individual config objects.
	appConfig, aerr := provconfig.ExtractAppConfig(configCmd)
	s.Require().NoError(aerr, "extracting app config")
	cmtConfig, terr := provconfig.ExtractCmtConfig(configCmd)
	s.Require().NoError(terr, "extracting cometbft config")
	clientConfig, cerr := provconfig.ExtractClientConfig(configCmd)
	s.Require().NoError(cerr, "extracting client config")
	appConfig.MinGasPrices = pioconfig.GetProvenanceConfig().ProvenanceMinGasPrices
	// And then save them.
	provconfig.SaveConfigs(configCmd, appConfig, cmtConfig, clientConfig, false)
}

// executeConfigCmd executes the config command with the provided args, returning the command's output.
func (s *ConfigTestSuite) executeConfigCmd(args ...string) string {
	configCmd := s.getConfigCmd()
	configCmd.SetArgs(args)
	b := applyMockIOOutErr(configCmd)
	err := configCmd.Execute()
	s.Require().NoError(err, "unexpected error executing %s %q", configCmd.Name(), args)
	out, err := io.ReadAll(b)
	s.Require().NoError(err, "unexpected error reading %s %q output", configCmd.Name(), args)
	return string(out)
}

func (s *ConfigTestSuite) makeConfigHeaderLine(t, fn string) string {
	return fmt.Sprintf("%s Config: %s/config/%s (or env)", t, s.Home, fn)
}

func (s *ConfigTestSuite) makeAppConfigHeaderLines() string {
	return s.makeConfigHeaderLine(s.HeaderStrApp, s.BaseFNApp) + "\n----------------"
}

func (s *ConfigTestSuite) makeCMTConfigHeaderLines() string {
	return s.makeConfigHeaderLine(s.HeaderStrCMT, s.BaseFNCMT) + "\n---------------------"
}

func (s *ConfigTestSuite) makeClientConfigHeaderLines() string {
	return s.makeConfigHeaderLine(s.HeaderStrClient, s.baseFNClient) + "\n-------------------"
}

func (s *ConfigTestSuite) makeConfigUpdatedLine(t, fn string) string {
	return fmt.Sprintf("%s Config Updated: %s/config/%s", t, s.Home, fn)
}

func (s *ConfigTestSuite) makeAppConfigUpdateLines() string {
	return s.makeConfigUpdatedLine(s.HeaderStrApp, s.BaseFNApp) + "\n------------------------"
}

func (s *ConfigTestSuite) makeCMTConfigUpdateLines() string {
	return s.makeConfigUpdatedLine(s.HeaderStrCMT, s.BaseFNCMT) + "\n-----------------------------"
}

func (s *ConfigTestSuite) makeClientConfigUpdateLines() string {
	return s.makeConfigUpdatedLine(s.HeaderStrClient, s.baseFNClient) + "\n---------------------------"
}

func (s *ConfigTestSuite) makeConfigDiffHeaderLine(t, fn string) string {
	return fmt.Sprintf("%s Config Differences from Defaults: %s/config/%s (or env)", t, s.Home, fn)
}

func (s *ConfigTestSuite) makeAppDiffHeaderLines() string {
	return s.makeConfigDiffHeaderLine(s.HeaderStrApp, s.BaseFNApp) + "\n------------------------------------------"
}

func (s *ConfigTestSuite) makeCMTDiffHeaderLines() string {
	return s.makeConfigDiffHeaderLine(s.HeaderStrCMT, s.BaseFNCMT) + "\n-----------------------------------------------"
}

func (s *ConfigTestSuite) makeClientDiffHeaderLines() string {
	return s.makeConfigDiffHeaderLine(s.HeaderStrClient, s.baseFNClient) + "\n---------------------------------------------"
}

func (s *ConfigTestSuite) makeTmDeprecatedLines(opt string) string {
	return "The \"" + opt + "\" option is deprecated and will be removed in a future version.\n" +
		"Use one of \"cometbft\", \"comet\", or \"cmt\" instead.\n"
}

func (s *ConfigTestSuite) makeMultiLine(lines ...string) string {
	return strings.Join(lines, "\n") + "\n"
}

func (s *ConfigTestSuite) makeKeyUpdatedLine(key, oldVal, newVal string) string {
	return fmt.Sprintf("%s Was: %s, Is Now: %s", key, oldVal, newVal)
}

func applyMockIODiscardOutErr(c *cobra.Command) *bytes.Buffer {
	b := bytes.NewBufferString("")
	c.SetOut(io.Discard)
	c.SetErr(io.Discard)
	return b
}

func applyMockIOOutErr(c *cobra.Command) *bytes.Buffer {
	b := bytes.NewBufferString("")
	c.SetOut(b)
	c.SetErr(b)
	return b
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
			configCmd := s.getConfigCmd()
			configCmd.SetArgs(tc.args)
			err := configCmd.Execute()
			require.EqualError(t, err, tc.err, "%s %s expected error executing configCmd", configCmd.Name(), tc.args)
		})
	}
}

func (s *ConfigTestSuite) TestConfigCmdGet() {
	noArgsOut := s.executeConfigCmd("get")

	s.Run("all fields and defaults", func() {
		expectedAll := s.makeMultiLine(
			s.makeAppConfigHeaderLines(),
			`app-db-backend=""
halt-height=0
halt-time=0
iavl-cache-size=781250
iavl-disable-fastnode=true
index-events=[]
inter-block-cache=true
min-retain-blocks=0
minimum-gas-prices="5confcoin"
pruning="default"
pruning-interval="0"
pruning-keep-recent="0"
query-gas-limit=0
api.address="tcp://localhost:1317"
api.enable=false
api.enabled-unsafe-cors=false
api.max-open-connections=1000
api.rpc-max-body-bytes=1000000
api.rpc-read-timeout=10
api.rpc-write-timeout=0
api.swagger=false
grpc-web.enable=true
grpc.address="localhost:9090"
grpc.enable=true
grpc.max-recv-msg-size=10485760
grpc.max-send-msg-size=2147483647
mempool.max-txs=5000
state-sync.snapshot-interval=0
state-sync.snapshot-keep-recent=2
streaming.abci.keys=[]
streaming.abci.plugin=""
streaming.abci.stop-node-on-err=true
telemetry.datadog-hostname=""
telemetry.enable-hostname=false
telemetry.enable-hostname-label=false
telemetry.enable-service-label=false
telemetry.enabled=false
telemetry.global-labels=[]
telemetry.metrics-sink=""
telemetry.prometheus-retention-time=0
telemetry.service-name=""
telemetry.statsd-addr=""`,
			"",
			s.makeCMTConfigHeaderLines(),
			`abci="socket"
db_backend="goleveldb"
db_dir="data"
filter_peers=false
genesis_file="config/genesis.json"
log_format="plain"
log_level="info"
moniker="F0796.localdomain"
node_key_file="config/node_key.json"
priv_validator_key_file="config/priv_validator_key.json"
priv_validator_laddr=""
priv_validator_state_file="data/priv_validator_state.json"
proxy_app="tcp://127.0.0.1:26658"
version="0.38.7"
blocksync.version="v0"
consensus.create_empty_blocks=true
consensus.create_empty_blocks_interval="0s"
consensus.double_sign_check_height=0
consensus.peer_gossip_sleep_duration="100ms"
consensus.peer_query_maj23_sleep_duration="2s"
consensus.skip_timeout_commit=false
consensus.timeout_commit="1.5s"
consensus.timeout_precommit="1s"
consensus.timeout_precommit_delta="500ms"
consensus.timeout_prevote="1s"
consensus.timeout_prevote_delta="500ms"
consensus.timeout_propose="3s"
consensus.timeout_propose_delta="500ms"
consensus.wal_file="data/cs.wal/wal"
instrumentation.max_open_connections=3
instrumentation.namespace="cometbft"
instrumentation.prometheus=false
instrumentation.prometheus_listen_addr=":26660"
mempool.broadcast=true
mempool.cache_size=10000
mempool.experimental_max_gossip_connections_to_non_persistent_peers=0
mempool.experimental_max_gossip_connections_to_persistent_peers=0
mempool.keep-invalid-txs-in-cache=false
mempool.max_batch_bytes=0
mempool.max_tx_bytes=1048576
mempool.max_txs_bytes=1073741824
mempool.recheck=true
mempool.size=5000
mempool.type="flood"
mempool.wal_dir=""
p2p.addr_book_file="config/addrbook.json"
p2p.addr_book_strict=true
p2p.allow_duplicate_ip=false
p2p.dial_timeout="3s"
p2p.external_address=""
p2p.flush_throttle_timeout="100ms"
p2p.handshake_timeout="20s"
p2p.laddr="tcp://0.0.0.0:26656"
p2p.max_num_inbound_peers=40
p2p.max_num_outbound_peers=10
p2p.max_packet_msg_payload_size=1024
p2p.persistent_peers=""
p2p.persistent_peers_max_dial_period="0s"
p2p.pex=true
p2p.private_peer_ids=""
p2p.recv_rate=5120000
p2p.seed_mode=false
p2p.seeds=""
p2p.send_rate=5120000
p2p.unconditional_peer_ids=""
rpc.cors_allowed_headers=["Origin", "Accept", "Content-Type", "X-Requested-With", "X-Server-Time"]
rpc.cors_allowed_methods=["HEAD", "GET", "POST"]
rpc.cors_allowed_origins=[]
rpc.experimental_close_on_slow_client=false
rpc.experimental_subscription_buffer_size=200
rpc.experimental_websocket_write_buffer_size=200
rpc.grpc_laddr=""
rpc.grpc_max_open_connections=900
rpc.laddr="tcp://127.0.0.1:26657"
rpc.max_body_bytes=1000000
rpc.max_header_bytes=1048576
rpc.max_open_connections=900
rpc.max_subscription_clients=100
rpc.max_subscriptions_per_client=5
rpc.pprof_laddr=""
rpc.timeout_broadcast_tx_commit="10s"
rpc.tls_cert_file=""
rpc.tls_key_file=""
rpc.unsafe=false
statesync.chunk_fetchers=4
statesync.chunk_request_timeout="10s"
statesync.discovery_time="15s"
statesync.enable=false
statesync.rpc_servers=[]
statesync.temp_dir=""
statesync.trust_hash=""
statesync.trust_height=0
statesync.trust_period="168h0m0s"
storage.discard_abci_responses=false
tx_index.indexer="null"
tx_index.psql-conn=""`,
			"",
			s.makeClientConfigHeaderLines(),
			`broadcast-mode="block"
chain-id=""
keyring-backend="test"
node="tcp://localhost:26657"
output="text"`,
			"",
		)

		// This test compares the no-args output to a previously known result.
		// If this test fails unexpectedly, we'll probably want to discuss how those changes
		// affect us, and whether we'll need to make further changes to accommodate the updates.
		s.Assert().Equal(expectedAll, noArgsOut)
	})

	s.Run("get all", func() {
		allOutStr := s.executeConfigCmd("get", "all")
		s.Assert().Equal(noArgsOut, allOutStr, "output of get vs output of get all")
	})

	// Test that the various sub-sections are in the no-args output.
	var cmtOut string
	for _, opt := range []string{"app", "cosmos", "config", "cometbft", "comet", "cmt", "client"} {
		args := []string{"get", opt}
		s.Run(strings.Join(args, " "), func() {
			outStr := s.executeConfigCmd(args...)
			s.Assert().Contains(noArgsOut, outStr, "output of get vs output of %q", args)
			if opt == "cometbft" {
				cmtOut = outStr
			}
		})
	}

	// Test that the deprecated tendermint and tm options have both the deprecation message and the section content.
	for _, opt := range []string{"tendermint", "tm"} {
		args := []string{"get", opt}
		s.Run(strings.Join(args, " "), func() {
			expTMOut := s.makeTmDeprecatedLines(opt) + cmtOut
			outStr := s.executeConfigCmd(args...)
			s.Assert().Equal(expTMOut, outStr)
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
			keys: []string{"min-retain-blocks", "mempool.max-txs", "grpc.address"},
			expected: s.makeMultiLine(
				s.makeAppConfigHeaderLines(),
				`min-retain-blocks=0`,
				`grpc.address="localhost:9090"`,
				`mempool.max-txs=5000`,
				""),
		},
		{
			name: "three cometbft config keys",
			keys: []string{"p2p.send_rate", "genesis_file", "consensus.timeout_propose"},
			expected: s.makeMultiLine(
				s.makeCMTConfigHeaderLines(),
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
			keys: []string{"rpc.cors_allowed_origins", "pruning", "node", "api.enable", "chain-id", "priv_validator_state_file"},
			expected: s.makeMultiLine(
				s.makeAppConfigHeaderLines(),
				`pruning="default"`,
				`api.enable=false`,
				"",
				s.makeCMTConfigHeaderLines(),
				`priv_validator_state_file="data/priv_validator_state.json"`,
				`rpc.cors_allowed_origins=[]`,
				"",
				s.makeClientConfigHeaderLines(),
				`chain-id=""`,
				`node="tcp://localhost:26657"`,
				""),
		},
		{
			name: "one category that is in two files",
			keys: []string{"mempool"},
			expected: s.makeMultiLine(
				s.makeAppConfigHeaderLines(),
				`mempool.max-txs=5000`,
				"",
				s.makeCMTConfigHeaderLines(),
				`mempool.broadcast=true`,
				`mempool.cache_size=10000`,
				`mempool.experimental_max_gossip_connections_to_non_persistent_peers=0`,
				`mempool.experimental_max_gossip_connections_to_persistent_peers=0`,
				`mempool.keep-invalid-txs-in-cache=false`,
				`mempool.max_batch_bytes=0`,
				`mempool.max_tx_bytes=1048576`,
				`mempool.max_txs_bytes=1073741824`,
				`mempool.recheck=true`,
				`mempool.size=5000`,
				`mempool.type="flood"`,
				`mempool.wal_dir=""`,
				""),
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			args := []string{"get"}
			args = append(args, tc.keys...)

			configCmd := s.getConfigCmd()
			configCmd.SetArgs(args)
			b := applyMockIOOutErr(configCmd)
			err := configCmd.Execute()
			require.NoError(t, err, "%s %s - unexpected error executing configCmd", configCmd.Name(), args)
			out, err := io.ReadAll(b)
			require.NoError(t, err, "%s %s - unexpected error reading configCmd output", configCmd.Name(), args)
			outStr := string(out)
			assert.Equal(t, tc.expected, outStr, "%s %s - output", configCmd.Name(), args)
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

		configCmd := s.getConfigCmd()
		configCmd.SetArgs(args)
		b := applyMockIOOutErr(configCmd)
		err := configCmd.Execute()
		require.NoError(t, err, "%s %s - expected error executing configCmd", configCmd.Name(), args)
		out, err := io.ReadAll(b)
		require.NoError(t, err, "%s %s - unexpected error reading configCmd output", configCmd.Name(), args)
		outStr := string(out)
		assert.Equal(t, expected, outStr, "%s %s - output", configCmd.Name(), args)
	})

	s.T().Run("two found one missing", func(t *testing.T) {
		expectedError := "1 configuration key not found: cannot.find.me"
		expected := s.makeMultiLine(
			s.makeAppConfigHeaderLines(),
			`grpc.enable=true`,
			"",
			s.makeCMTConfigHeaderLines(),
			`consensus.create_empty_blocks_interval="0s"`,
			"",
		) + "Error: " + expectedError + "\n"
		args := []string{"get", "cannot.find.me", "consensus.create_empty_blocks_interval", "grpc.enable"}

		configCmd := s.getConfigCmd()
		configCmd.SetArgs(args)
		b := applyMockIOOutErr(configCmd)
		err := configCmd.Execute()
		require.NoError(t, err, "%s %s - expected error executing configCmd", configCmd.Name(), args)
		out, err := io.ReadAll(b)
		require.NoError(t, err, "%s %s - unexpected error reading configCmd output", configCmd.Name(), args)
		outStr := string(out)
		assert.Equal(t, expected, outStr, "%s %s - output", configCmd.Name(), args)
	})
}

func (s *ConfigTestSuite) TestConfigChanged() {
	allEqual := func(t string) string {
		return fmt.Sprintf("All %s config values equal the default config values.", t)
	}
	expectedAppOutLines := []string{
		s.makeAppDiffHeaderLines(),
		allEqual("app"),
		"",
	}
	expectedCMTOutLines := []string{
		s.makeCMTDiffHeaderLines(),
		allEqual("cometbft"),
		"",
	}
	expectedClientOutLines := []string{
		s.makeClientDiffHeaderLines(),
		allEqual("client"),
		"",
	}
	var expectedAllOutLines []string
	expectedAllOutLines = append(expectedAllOutLines, expectedAppOutLines...)
	expectedAllOutLines = append(expectedAllOutLines, expectedCMTOutLines...)
	expectedAllOutLines = append(expectedAllOutLines, expectedClientOutLines...)
	expectedAppOut := s.makeMultiLine(expectedAppOutLines...)
	expectedCMTOut := s.makeMultiLine(expectedCMTOutLines...)
	expectedClientOut := s.makeMultiLine(expectedClientOutLines...)
	expectedAll := s.makeMultiLine(expectedAllOutLines...)

	equalAllTests := []struct {
		args []string
		out  string
	}{
		{args: []string{"changed"}, out: expectedAll},
		{args: []string{"changed", "all"}, out: expectedAll},
		{args: []string{"changed", "app"}, out: expectedAppOut},
		{args: []string{"changed", "cosmos"}, out: expectedAppOut},
		{args: []string{"changed", "config"}, out: expectedCMTOut},
		{args: []string{"changed", "tm"}, out: s.makeTmDeprecatedLines("tm") + expectedCMTOut},
		{args: []string{"changed", "tendermint"}, out: s.makeTmDeprecatedLines("tendermint") + expectedCMTOut},
		{args: []string{"changed", "cometbft"}, out: expectedCMTOut},
		{args: []string{"changed", "comet"}, out: expectedCMTOut},
		{args: []string{"changed", "cmt"}, out: expectedCMTOut},
		{args: []string{"changed", "client"}, out: expectedClientOut},
		{
			args: []string{"changed", "output"},
			out: s.makeMultiLine(
				s.makeClientDiffHeaderLines(),
				`output="text" (same as default)`,
				"",
			),
		},
		{
			args: []string{"changed", "mempool"},
			out: s.makeMultiLine(
				s.makeAppDiffHeaderLines(),
				`mempool.max-txs=5000 (same as default)`,
				"",
				s.makeCMTDiffHeaderLines(),
				`mempool.broadcast=true (same as default)`,
				`mempool.cache_size=10000 (same as default)`,
				`mempool.experimental_max_gossip_connections_to_non_persistent_peers=0 (same as default)`,
				`mempool.experimental_max_gossip_connections_to_persistent_peers=0 (same as default)`,
				`mempool.keep-invalid-txs-in-cache=false (same as default)`,
				`mempool.max_batch_bytes=0 (same as default)`,
				`mempool.max_tx_bytes=1048576 (same as default)`,
				`mempool.max_txs_bytes=1073741824 (same as default)`,
				`mempool.recheck=true (same as default)`,
				`mempool.size=5000 (same as default)`,
				`mempool.type="flood" (same as default)`,
				`mempool.wal_dir="" (same as default)`,
				"",
			),
		},
	}

	for _, tc := range equalAllTests {
		s.Run(strings.Join(tc.args, " "), func() {
			outStr := s.executeConfigCmd(tc.args...)
			s.Assert().Equal(tc.out, outStr)
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
			name: "set cometbft fails validation",
			args: []string{"set", "log_format", "crazy"},
			out:  `CometBFT config validation error: unknown log_format (must be 'plain' or 'json')`,
		},
		{
			name: "set client fails validation",
			args: []string{"set", "output", "csv"},
			out:  `Client config validation error: unknown output (must be 'text' or 'json')`,
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			expected := "Error: one or more issues encountered; no configuration values have been updated"
			configCmd := s.getConfigCmd()
			configCmd.SetArgs(tc.args)
			b := applyMockIOOutErr(configCmd)
			err := configCmd.Execute()
			require.NoError(t, err, "%s %s unexpected error executing configCmd", configCmd.Name(), tc.args)
			out, rerr := io.ReadAll(b)
			require.NoError(t, rerr, "%s %s unexpected error reading output", configCmd.Name(), tc.args)
			outStr := string(out)
			assert.True(t, strings.Contains(outStr, expected), "%s %s output", configCmd.Name(), tc.args)
		})
	}
}

func (s *ConfigTestSuite) TestConfigCmdSet() {
	reAppConfigUpdated := regexp.MustCompile(`(?m)^App Config Updated: .*/config/` + s.BaseFNApp + `$`)
	reCMTConfigUpdated := regexp.MustCompile(`(?m)^CometBFT Config Updated: .*/config/` + s.BaseFNCMT + `$`)
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

		// cometbft fields
		{
			name:    "filter_peers",
			oldVal:  `false`,
			newVal:  `true`,
			toMatch: []*regexp.Regexp{reCMTConfigUpdated},
		},
		{
			name:    "proxy_app",
			oldVal:  `"tcp://127.0.0.1:26658"`,
			newVal:  `"tcp://localhost:26658"`,
			toMatch: []*regexp.Regexp{reCMTConfigUpdated},
		},
		{
			name:    "consensus.timeout_commit",
			oldVal:  fmt.Sprintf("%q", provconfig.DefaultConsensusTimeoutCommit),
			newVal:  `"2s"`,
			toMatch: []*regexp.Regexp{reCMTConfigUpdated},
		},
		{
			name:    "mempool.cache_size",
			oldVal:  `10000`,
			newVal:  `10005`,
			toMatch: []*regexp.Regexp{reCMTConfigUpdated},
		},
		{
			name:    "rpc.cors_allowed_methods",
			oldVal:  `["HEAD", "GET", "POST"]`,
			newVal:  `["POST", "HEAD", "GET"]`,
			toMatch: []*regexp.Regexp{reCMTConfigUpdated},
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
			configCmd := s.getConfigCmd()
			configCmd.SetArgs(args)
			b := applyMockIOOutErr(configCmd)
			err := configCmd.Execute()
			require.NoError(t, err, "%s %s - unexpected error in execution", configCmd.Name(), args)
			out, rerr := io.ReadAll(b)
			require.NoError(t, rerr, "%s %s - unexpected error in reading", configCmd.Name(), args)
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
			name: "two cometbft entries",
			args: []string{"set", "log_format", "json", "consensus.timeout_commit", "950ms"},
			out: s.makeMultiLine(
				s.makeCMTConfigUpdateLines(),
				s.makeKeyUpdatedLine("log_format", `"plain"`, `"json"`),
				s.makeKeyUpdatedLine("consensus.timeout_commit", fmt.Sprintf("%q", provconfig.DefaultConsensusTimeoutCommit), `"950ms"`),
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
			args: []string{"set",
				"consensus.timeout_commit", "951ms",
				"api.swagger", "true",
				"telemetry.service-name", "blocky2",
				"node", "tcp://localhost:26657",
				"output", "text",
				"log_format", "plain"},
			out: s.makeMultiLine(
				s.makeAppConfigUpdateLines(),
				s.makeKeyUpdatedLine("api.swagger", "false", "true"),
				s.makeKeyUpdatedLine("telemetry.service-name", `"blocky"`, `"blocky2"`),
				"",
				s.makeCMTConfigUpdateLines(),
				s.makeKeyUpdatedLine("log_format", `"json"`, `"plain"`),
				s.makeKeyUpdatedLine("consensus.timeout_commit", `"950ms"`, `"951ms"`),
				"",
				s.makeClientConfigUpdateLines(),
				s.makeKeyUpdatedLine("node", `"tcp://127.0.0.1:26657"`, `"tcp://localhost:26657"`),
				s.makeKeyUpdatedLine("output", `"json"`, `"text"`),
				""),
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			configCmd := s.getConfigCmd()
			configCmd.SetArgs(tc.args)
			b := applyMockIOOutErr(configCmd)
			err := configCmd.Execute()
			require.NoError(t, err, "%s %s - unexpected error in execution", configCmd.Name(), tc.args)
			out, rerr := io.ReadAll(b)
			require.NoError(t, rerr, "%s %s - unexpected error in reading", configCmd.Name(), tc.args)
			outStr := string(out)
			assert.Equal(t, tc.out, outStr, "%s %s output", configCmd.Name(), tc.args)
		})
	}
}

func (s *ConfigTestSuite) TestPackUnpack() {
	s.T().Run("pack", func(t *testing.T) {
		expectedPacked := map[string]string{}
		expectedPackedJSON, jerr := json.MarshalIndent(expectedPacked, "", "  ")
		require.NoError(t, jerr, "making expected json")
		expectedPackedJSONStr := string(expectedPackedJSON)
		configCmd := s.getConfigCmd()
		args := []string{"pack"}
		configCmd.SetArgs(args)
		b := applyMockIOOutErr(configCmd)
		err := configCmd.Execute()
		require.NoError(t, err, "%s %s - unexpected error in execution", configCmd.Name(), args)
		out, rerr := io.ReadAll(b)
		require.NoError(t, rerr, "%s %s - unexpected error in reading", configCmd.Name(), args)
		outStr := string(out)

		assert.Contains(t, outStr, expectedPackedJSONStr, "packed json")
		packedFile := provconfig.GetFullPathToPackedConf(configCmd)

		assert.Contains(t, outStr, packedFile, "packed filename")
		assert.True(t, provconfig.FileExists(packedFile), "file exists: packed")
		appFile := provconfig.GetFullPathToAppConf(configCmd)
		assert.Contains(t, outStr, appFile, "app filename")
		assert.False(t, provconfig.FileExists(appFile), "file exists: app")
		cmtFile := provconfig.GetFullPathToAppConf(configCmd)
		assert.Contains(t, outStr, cmtFile, "cometbft filename")
		assert.False(t, provconfig.FileExists(cmtFile), "file exists: cometbft")
		clientFile := provconfig.GetFullPathToAppConf(configCmd)
		assert.Contains(t, outStr, clientFile, "client filename")
		assert.False(t, provconfig.FileExists(clientFile), "file exists: client")
	})

	s.T().Run("unpack", func(t *testing.T) {
		configCmd := s.getConfigCmd()
		args := []string{"unpack"}
		configCmd.SetArgs(args)
		b := applyMockIOOutErr(configCmd)
		err := configCmd.Execute()
		require.NoError(t, err, "%s %s - unexpected error in execution", configCmd.Name(), args)
		out, rerr := io.ReadAll(b)
		require.NoError(t, rerr, "%s %s - unexpected error in reading", configCmd.Name(), args)
		outStr := string(out)

		packedFile := provconfig.GetFullPathToPackedConf(configCmd)
		assert.Contains(t, outStr, packedFile, "packed filename")
		assert.False(t, provconfig.FileExists(packedFile), "file exists: packed")
		appFile := provconfig.GetFullPathToAppConf(configCmd)
		assert.Contains(t, outStr, appFile, "app filename")
		assert.True(t, provconfig.FileExists(appFile), "file exists: app")
		cmtFile := provconfig.GetFullPathToAppConf(configCmd)
		assert.Contains(t, outStr, cmtFile, "cometbft filename")
		assert.True(t, provconfig.FileExists(cmtFile), "file exists: cometbft")
		clientFile := provconfig.GetFullPathToAppConf(configCmd)
		assert.Contains(t, outStr, clientFile, "client filename")
		assert.True(t, provconfig.FileExists(clientFile), "file exists: client")
	})
}

func (s *ConfigTestSuite) TestEmptyPackedConfigHasDefaultMinGas() {
	expected := provconfig.DefaultAppConfig().MinGasPrices
	s.Require().NotEqual("", expected, "default MinGasPrices")

	// Pack the config, then rewrite the file to be an empty object.
	pcmd := s.getConfigCmd()
	pcmd.SetArgs([]string{"pack"})
	err := pcmd.Execute()
	s.Require().NoError(err, "pack the config")
	s.Require().NoError(os.WriteFile(provconfig.GetFullPathToPackedConf(pcmd), []byte("{}"), 0o644), "writing empty packed config")

	// Now read the config and check that the min gas prices are the default that we want.
	ncmd := s.getConfigCmd()
	err = provconfig.LoadConfigFromFiles(ncmd)
	s.Require().NoError(err, "LoadConfigFromFiles")

	appConfig, err := provconfig.ExtractAppConfig(ncmd)
	s.Require().NoError(err, "ExtractAppConfig")
	actual := appConfig.MinGasPrices
	s.Assert().Equal(expected, actual, "MinGasPrices")
}

func (s *ConfigTestSuite) TestUpdate() {
	// Write a dumb version of each file that is missing almost everything, but has an easily identifiable comment.
	customComment := "# There's no way this is comment is in the real config."
	uFileField := "bananas"
	uFileValue := "no it's not"
	uFileContents := fmt.Sprintf("%s\n%s = %q\n", customComment, uFileField, uFileValue)
	dbBackend := "bananas"
	tFileContents := fmt.Sprintf("%s\n%s = %q\n", customComment, "db_backend", dbBackend)
	minGasPrices := "not a lot"
	aFileContents := fmt.Sprintf("%s\n%s = %q\n", customComment, "minimum-gas-prices", minGasPrices)
	chainId := "this-will-never-work"
	cFileContents := fmt.Sprintf("%s\n%s = %q\n", customComment, "chain-id", chainId)
	configCmd := s.getConfigCmd()
	configDir := provconfig.GetFullPathToConfigDir(configCmd)
	uFile := provconfig.GetFullPathToUnmanagedConf(configCmd)
	tFile := provconfig.GetFullPathToCmtConf(configCmd)
	aFile := provconfig.GetFullPathToAppConf(configCmd)
	cFile := provconfig.GetFullPathToClientConf(configCmd)
	s.Require().NoError(os.MkdirAll(configDir, 0o755), "making config dir")
	s.Require().NoError(os.WriteFile(uFile, []byte(uFileContents), 0o644), "writing unmanaged config")
	s.Require().NoError(os.WriteFile(tFile, []byte(tFileContents), 0o644), "writing cometbft config")
	s.Require().NoError(os.WriteFile(aFile, []byte(aFileContents), 0o644), "writing app config")
	s.Require().NoError(os.WriteFile(cFile, []byte(cFileContents), 0o644), "writing client config")

	// Load the config from files.
	// Usually this is done in the pre-run handler, but that's defined in the root command,
	// so this one doesn't know about it.
	err := provconfig.LoadConfigFromFiles(configCmd)
	s.Require().NoError(err, "loading config from files")

	// Run the command!
	args := []string{"update"}
	configCmd.SetArgs(args)
	err = configCmd.Execute()
	s.Require().NoError(err, "%s %s - unexpected error in execution", configCmd.Name(), args)

	s.Run("unmanaged file is unchanged", func() {
		actualUFileContents, err := os.ReadFile(uFile)
		s.Require().NoError(err, "reading unmanaged file: %s", uFile)
		s.Assert().Equal(uFileContents, string(actualUFileContents), "unmanaged file contents")
	})

	s.Run("cometbft file has been updated", func() {
		actualTFileContents, err := os.ReadFile(tFile)
		s.Require().NoError(err, "reading cometbft file: %s", tFile)
		s.Assert().NotEqual(tFileContents, string(actualTFileContents), "cometbft file contents")
		lines := strings.Split(string(actualTFileContents), "\n")
		s.Assert().Greater(len(lines), 2, "number of lines in cometbft file.")
	})

	s.Run("app file has been updated", func() {
		actualAFileContents, err := os.ReadFile(aFile)
		s.Require().NoError(err, "reading app file: %s", aFile)
		s.Assert().NotEqual(aFileContents, string(actualAFileContents), "app file contents")
		lines := strings.Split(string(actualAFileContents), "\n")
		s.Assert().Greater(len(lines), 2, "number of lines in app file.")
	})

	s.Run("client file has been updated", func() {
		actualCFileContents, err := os.ReadFile(cFile)
		s.Require().NoError(err, "reading client file: %s", cFile)
		s.Assert().NotEqual(aFileContents, string(actualCFileContents), "client file contents")
		lines := strings.Split(string(actualCFileContents), "\n")
		s.Assert().Greater(len(lines), 2, "number of lines in client file.")
	})

	err = provconfig.LoadConfigFromFiles(configCmd)
	s.Require().NoError(err, "loading config from files")

	s.Run("cometbft db_backend value unchanged", func() {
		cmtConfig, err := provconfig.ExtractCmtConfig(configCmd)
		s.Require().NoError(err, "ExtractCmtConfig")
		actual := cmtConfig.DBBackend
		s.Assert().Equal(dbBackend, actual, "DBBackend")
	})

	s.Run("app minimum-gas-prices value unchanged", func() {
		appConfig, err := provconfig.ExtractAppConfig(configCmd)
		s.Require().NoError(err, "ExtractAppConfig")
		actual := appConfig.MinGasPrices
		s.Assert().Equal(minGasPrices, actual, "MinGasPrices")
	})

	s.Run("client chain-id value unchanged", func() {
		clientConfig, err := provconfig.ExtractClientConfig(configCmd)
		s.Require().NoError(err, "ExtractClientConfig")
		actual := clientConfig.ChainID
		s.Assert().Equal(chainId, actual, "ChainID")
	})

	s.Run("unmanaged config entry still applies", func() {
		ctx := client.GetClientContextFromCmd(configCmd)
		vpr := ctx.Viper
		actual := vpr.GetString(uFileField)
		s.Assert().Equal(uFileValue, actual, "unmanaged config entry")
	})
}
