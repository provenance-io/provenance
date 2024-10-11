package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wasmcli "github.com/CosmWasm/wasmd/x/wasm/client/cli"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"

	"github.com/provenance-io/provenance/cmd/provenanced/config"
	"github.com/provenance-io/provenance/internal"
)

func TestIAVLConfig(t *testing.T) {
	require.Equal(t, getIAVLCacheSize(simtestutil.EmptyAppOptions{}), cast.ToInt(serverconfig.DefaultConfig().IAVLCacheSize))
}

type testOpts map[string]interface{}

var _ servertypes.AppOptions = (*testOpts)(nil)

func (o testOpts) Get(key string) interface{} {
	return o[key]
}

type panicOpts struct{}

var _ servertypes.AppOptions = (*panicOpts)(nil)

func (o panicOpts) Get(key string) interface{} {
	panic(fmt.Errorf("panic forced while getting option %q", key))
}

func TestWarnAboutSettings(t *testing.T) {
	keyChainID := flags.FlagChainID
	keySkipTimeoutCommit := "consensus.skip_timeout_commit"
	keyTimeoutCommit := "consensus.timeout_commit"

	upperLimit := config.DefaultConsensusTimeoutCommit + 2*time.Second
	overUpperLimit := upperLimit + 1*time.Millisecond
	mainnetChainID := "pio-mainnet-1"
	notMainnetChainID := "not-" + mainnetChainID

	tooHighMsg := func(timeoutCommit time.Duration) string {
		return fmt.Sprintf("ERR Your consensus.timeout_commit config value is too high and should be decreased to at most %q. The recommended value is %q. Your current value is %q.",
			upperLimit, config.DefaultConsensusTimeoutCommit, timeoutCommit)
	}

	mainnetOpts := func(skipTimeoutCommit bool, timeoutCommit time.Duration) testOpts {
		return testOpts{
			keyChainID:           mainnetChainID,
			keySkipTimeoutCommit: skipTimeoutCommit,
			keyTimeoutCommit:     timeoutCommit,
		}
	}
	notMainnetOpts := func(skipTimeoutCommit bool, timeoutCommit time.Duration) testOpts {
		return testOpts{
			keyChainID:           notMainnetChainID,
			keySkipTimeoutCommit: skipTimeoutCommit,
			keyTimeoutCommit:     timeoutCommit,
		}
	}

	tests := []struct {
		name      string
		appOpts   servertypes.AppOptions
		expLogged []string
	}{
		{
			name:      "mainnet not skipped value over upper limit",
			appOpts:   mainnetOpts(false, overUpperLimit),
			expLogged: []string{tooHighMsg(overUpperLimit)},
		},
		{
			name:    "mainnet not skipped value at limit",
			appOpts: mainnetOpts(false, upperLimit),
		},
		{
			name:    "mainnet not skipped value at default",
			appOpts: mainnetOpts(false, config.DefaultConsensusTimeoutCommit),
		},
		{
			name:    "mainnet not skipped value at zero",
			appOpts: mainnetOpts(false, 0),
		},
		{
			name:    "mainnet skipped value over upper limit",
			appOpts: mainnetOpts(true, overUpperLimit),
		},
		{
			name:    "mainnet skipped value at upper limit",
			appOpts: mainnetOpts(true, upperLimit),
		},
		{
			name:    "mainnet skipped value at default",
			appOpts: mainnetOpts(true, config.DefaultConsensusTimeoutCommit),
		},
		{
			name:    "mainnet skipped value at zero",
			appOpts: mainnetOpts(true, 0),
		},
		{
			name:    "not mainnet not skipped value over upper limit",
			appOpts: notMainnetOpts(false, overUpperLimit),
		},
		{
			name:    "not mainnet not skipped value at upper limit",
			appOpts: notMainnetOpts(false, upperLimit),
		},
		{
			name:    "not mainnet not skipped value at default",
			appOpts: notMainnetOpts(false, config.DefaultConsensusTimeoutCommit),
		},
		{
			name:    "not mainnet not skipped value at zero",
			appOpts: notMainnetOpts(false, 0),
		},
		{
			name:    "not mainnet skipped value over upper limit",
			appOpts: notMainnetOpts(true, overUpperLimit),
		},
		{
			name:    "not mainnet skipped value at upper limit",
			appOpts: notMainnetOpts(true, upperLimit),
		},
		{
			name:    "not mainnet skipped value at default",
			appOpts: notMainnetOpts(true, config.DefaultConsensusTimeoutCommit),
		},
		{
			name:    "not mainnet skipped value at zero",
			appOpts: notMainnetOpts(true, 0),
		},
		{
			name:    "timeout commit opt not a duration",
			appOpts: testOpts{keyChainID: mainnetChainID, keySkipTimeoutCommit: false, keyTimeoutCommit: "nope"},
		},
		{
			name:    "empty opts",
			appOpts: testOpts{},
		},
		{
			name:    "only chain-id mainnet",
			appOpts: testOpts{keyChainID: mainnetChainID},
		},
		{
			name:    "only chain-id not mainnet",
			appOpts: testOpts{keyChainID: mainnetChainID},
		},
		{
			name:    "mainnet not skipped no timeout commit opt",
			appOpts: testOpts{keyChainID: mainnetChainID, keySkipTimeoutCommit: false},
		},
		{
			name:    "panic from getter",
			appOpts: panicOpts{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// If expLogged isn't defined, we're expecting zero logged lines.
			if tc.expLogged == nil {
				tc.expLogged = make([]string, 0)
			}

			// Create a logger wrapped around our own buffer so we can see what was logged.
			var buf bytes.Buffer
			lw := zerolog.ConsoleWriter{
				Out:          &buf,
				NoColor:      true,
				PartsExclude: []string{"time"}, // Without this, each line starts with "<nil> "
			}
			// Error log lines will start with "ERR ".
			// Info log lines will start with "INF ".
			// Debug log lines are omitted, but would start with "DBG ".
			logger := log.NewCustomLogger(zerolog.New(lw).Level(zerolog.InfoLevel))

			// Make sure the function never panics.
			require.NotPanics(t, func() { warnAboutSettings(logger, tc.appOpts) }, "warnAboutSettings")

			// Make sure the logged output is as expected.
			out := buf.String()
			// Splitting it into lines and comparing the string slices makes test failure output nicer than
			// if we just compare two multi-line strings.
			loggedLines := strings.Split(out, "\n")
			// Get rid of last "line" if it's just an empty string. Happens when out ends with "\n".
			if len(loggedLines[len(loggedLines)-1]) == 0 {
				loggedLines = loggedLines[:len(loggedLines)-1]
			}
			assert.Equal(t, tc.expLogged, loggedLines, "lines logged during warnAboutSettings")
		})
	}
}

func TestFixDebugPubkeyRawTypeFlagCanary(t *testing.T) {
	// If this test fails, something has changed with the debug pubkey-raw command flags.
	// Either fixDebugPubkeyRawTypeFlag isn't needed or else it should be tweaked.
	cmd := debug.PubkeyRawCmd()
	typeFlag := cmd.Flags().Lookup("type")
	require.NotNil(t, typeFlag, "--type flag")
	assert.Equal(t, "t", typeFlag.Shorthand, "--type flag shorthand")
}

func TestFixTxWasmInstantiate2AliasesCanary(t *testing.T) {
	// If this test fails, something has changed with the wasmd library.
	// Either fixTxWasmInstantiate2Aliases isn't needed or else it should be tweaked.
	txWasmCmd := wasmcli.GetTxCmd()
	i1Cmd, _, err := txWasmCmd.Find([]string{"instantiate"})
	require.NoError(t, err, "Finding the instantiate sub-command")
	require.NotNil(t, i1Cmd, "The instantiate sub-command")
	i2Cmd, _, err := txWasmCmd.Find([]string{"instantiate2"})
	require.NoError(t, err, "Finding the instantiate2 sub-command")
	require.NotNil(t, i2Cmd, "The instantiate2 sub-command")

	i1Aliases := i1Cmd.Aliases
	i2Aliases := i2Cmd.Aliases
	assert.Equal(t, i1Aliases, i2Aliases, "instantiate2 aliases")
}

func TestCmdsWithPersistentPreRun(t *testing.T) {
	// Our root command has a PersistentPreRunE. If a sub-command has one, the root one will not
	// be run for that sub-command and any of its children. That is almost certainly not good.
	// So, if this test fails, we'll need to add something that changes the sub-command's version to
	// a composed version of what it's got and the root command's.

	exp := []string{"provenanced"}

	rootCmd, _ := NewRootCmd(false)
	havePerPre := findCmdsWithPersistentPreRun(rootCmd)

	assert.Equal(t, exp, havePerPre, "commands that have a persistent pre-run function")
}

func findCmdsWithPersistentPreRun(cmd *cobra.Command) []string {
	var rv []string
	if cmd.PersistentPreRunE != nil || cmd.PersistentPreRun != nil {
		rv = append(rv, cmd.CommandPath())
	}
	for _, subCmd := range cmd.Commands() {
		rv = append(rv, findCmdsWithPersistentPreRun(subCmd)...)
	}
	return rv
}

func TestCmdAliasesAreUnique(t *testing.T) {
	rootCmd, _ := NewRootCmd(false)
	knownProblems := aliasNameMap{
		// Example entry (keep for future known problems).
		//"provenanced tx wasm start": {"instantiate", "instantiate2"},
	}
	assertCommandsAndAliasesAreUnique(t, knownProblems, nil, rootCmd)
}

type aliasNameMap map[string][]string

func assertCommandsAndAliasesAreUnique(t *testing.T, knownProblems, parentAliasNameMap aliasNameMap, cmd *cobra.Command) bool {
	rv := true
	cmdPath := cmd.CommandPath()
	t.Run(cmdPath, func(t *testing.T) {
		if parentAliasNameMap == nil || !cmd.HasParent() {
			return
		}
		toCheck := make([]string, 1, 1+len(cmd.Aliases))
		toCheck[0] = cmd.Name()
		toCheck = appendIfNew(toCheck, cmd.Aliases...)
		parentPath := cmd.Parent().CommandPath()
		for _, alias := range toCheck {
			exp, known := knownProblems[parentPath+" "+alias]
			if !known {
				exp = toCheck[0:1]
			}
			rv = assert.ElementsMatch(t, exp, parentAliasNameMap[alias],
				"The %q command has multiple sub-commands with name or alias %q (A = expected, B = actual)",
				parentPath, alias)
		}
	})

	subCmds := cmd.Commands()
	if len(subCmds) == 0 {
		return rv
	}

	subCmdsOfInterest := make([]*cobra.Command, 0, len(subCmds))
	cmdAliasNameMap := make(aliasNameMap)
	for _, subCmd := range subCmds {
		name := subCmd.Name()
		// Ignore commands without a name sometimes used for line breaks.
		if len(name) == 0 {
			continue
		}
		subCmdsOfInterest = append(subCmdsOfInterest, subCmd)
		cmdAliasNameMap[name] = appendIfNew(cmdAliasNameMap[name], name)
		for _, alias := range subCmd.Aliases {
			cmdAliasNameMap[alias] = appendIfNew(cmdAliasNameMap[alias], name)
		}
	}

	// Run the check on all sub-commands of interest.
	for _, sc := range subCmdsOfInterest {
		rv = assertCommandsAndAliasesAreUnique(t, knownProblems, cmdAliasNameMap, sc) && rv
	}

	return rv
}

func appendIfNew(slice []string, elems ...string) []string {
	for _, elem := range elems {
		has := false
		for _, entry := range slice {
			if entry == elem {
				has = true
				break
			}
		}
		if !has {
			slice = append(slice, elem)
		}
	}
	return slice
}

func TestIsTestnetFlagSet(t *testing.T) {
	envVar := "PIO_TESTNET"

	tests := []struct {
		name    string
		envArgs map[string]string
		args    []string
		exp     bool
	}{
		{
			name: "no env, nil args",
			args: nil,
			exp:  false,
		},
		{
			name: "no env, empty args",
			args: make([]string, 0),
			exp:  false,
		},
		{
			name: "one arg: -t",
			args: []string{"-t"},
			exp:  true,
		},
		{
			name: "one arg: -t=true",
			args: []string{"-t=true"},
			exp:  true,
		},
		{
			name: "one arg: -t=false",
			args: []string{"-t=false"},
			exp:  false,
		},
		{
			name: "one arg: -t=xxx",
			args: []string{"-t=xxx"},
			exp:  false,
		},
		{
			name: "one arg: --testnet",
			args: []string{"--testnet"},
			exp:  true,
		},
		{
			name: "one arg: --testnet=true",
			args: []string{"--testnet=true"},
			exp:  true,
		},
		{
			name: "one arg: --testnet=false",
			args: []string{"--testnet=false"},
			exp:  false,
		},
		{
			name: "one arg: --testnet=xxx",
			args: []string{"--testnet=xxx"},
			exp:  false,
		},
		{
			name: "five args: -t first",
			args: []string{"-t", "two", "three", "four", "five"},
			exp:  true,
		},
		{
			name: "five args: --testnet third",
			args: []string{"one", "two", "--testnet", "four", "five"},
			exp:  true,
		},
		{
			name: "five args: --testnet=true fifth",
			args: []string{"one", "two", "three", "four", "--testnet=true"},
			exp:  true,
		},
		{
			name: "five args: none of interest",
			args: []string{"one", "two", "three", "four", "five"},
			exp:  false,
		},
		{
			name: "two args: --testnet=true --testnet=false",
			args: []string{"--testnet=true", "--testnet=false"},
			exp:  false,
		},
		{
			name: "two args: --testnet=false --testnet",
			args: []string{"--testnet=false", "--testnet"},
			exp:  true,
		},
		{
			name:    "env var: true",
			envArgs: map[string]string{envVar: "true"},
			exp:     true,
		},
		{
			name:    "env var: false",
			envArgs: map[string]string{envVar: "false"},
			exp:     false,
		},
		{
			name:    "env var: xxx",
			envArgs: map[string]string{envVar: "xxx"},
			exp:     false,
		},
		{
			name:    "env var but lower-case",
			envArgs: map[string]string{strings.ToLower(envVar): "true"},
			exp:     false,
		},
		{
			name:    "env var false, args true",
			envArgs: map[string]string{envVar: "false"},
			args:    []string{"-t"},
			exp:     true,
		},
		{
			name:    "env var true, args false",
			envArgs: map[string]string{envVar: "true"},
			args:    []string{"--testnet=false"},
			exp:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envArgs {
				t.Setenv(k, v)
			}
			var actual bool
			testFunc := func() {
				actual = isTestnetFlagSet(tc.args)
			}
			require.NotPanics(t, testFunc, "isTestnetFlagSet")
			assert.Equal(t, tc.exp, actual, "result from isTestnetFlagSet")
		})
	}
}

func TestGetTelemetryGlobalLabels(t *testing.T) {
	newTestOpts := func(enabled, globalLabels interface{}) servertypes.AppOptions {
		return testOpts{
			telEnabledKey:      enabled,
			telGlobalLabelsKey: globalLabels,
		}
	}
	newViperOpts := func(enabled, globalLabels interface{}) servertypes.AppOptions {
		vpr := viper.New()
		if enabled != nil {
			vpr.Set(telEnabledKey, enabled)
		}
		if globalLabels != nil {
			vpr.Set(telGlobalLabelsKey, globalLabels)
		}
		return vpr
	}

	funcLogKV := "function=getTelemetryGlobalLabels"

	tests := []struct {
		name            string
		appOpts         servertypes.AppOptions
		expGlobalLabels [][]string
		expEnabled      bool
		expLogLines     []string
	}{
		{
			name:            "nil appOpts",
			appOpts:         nil,
			expGlobalLabels: nil,
			expEnabled:      false,
			expLogLines:     []string{"DBG No app options available. " + funcLogKV},
		},
		{
			name:            "viper: empty",
			appOpts:         newViperOpts(nil, nil),
			expGlobalLabels: make([][]string, 0),
			expEnabled:      false,
			expLogLines: []string{"DBG Using Telemetry config from app config. " + funcLogKV +
				" " + telEnabledKey + "=false " + telGlobalLabelsKey + "=[]"},
		},
		{
			name:            "viper: only enabled true",
			appOpts:         newViperOpts(true, nil),
			expGlobalLabels: make([][]string, 0),
			expEnabled:      true,
			expLogLines: []string{"DBG Using Telemetry config from app config. " + funcLogKV +
				" " + telEnabledKey + "=true " + telGlobalLabelsKey + "=[]"},
		},
		{
			name:            "viper: only enabled false",
			appOpts:         newViperOpts(false, nil),
			expGlobalLabels: make([][]string, 0),
			expEnabled:      false,
			expLogLines: []string{"DBG Using Telemetry config from app config. " + funcLogKV +
				" " + telEnabledKey + "=false " + telGlobalLabelsKey + "=[]"},
		},
		{
			name:            "viper: only global labels",
			appOpts:         newViperOpts(nil, [][]string{{"arg1", "arg2"}}),
			expGlobalLabels: [][]string{{"arg1", "arg2"}},
			expEnabled:      false,
			expLogLines: []string{"DBG Using Telemetry config from app config. " + funcLogKV +
				" " + telEnabledKey + "=false " + telGlobalLabelsKey + "=[[\"arg1\",\"arg2\"]]"},
		},
		{
			name:            "viper: enabled true and global labels",
			appOpts:         newViperOpts(true, [][]string{{"arg1", "arg2"}, {"arg3", "arg4", "arg5"}}),
			expGlobalLabels: [][]string{{"arg1", "arg2"}, {"arg3", "arg4", "arg5"}},
			expEnabled:      true,
			expLogLines: []string{"DBG Using Telemetry config from app config. " + funcLogKV +
				" " + telEnabledKey + "=true " + telGlobalLabelsKey + "=[[\"arg1\",\"arg2\"],[\"arg3\",\"arg4\",\"arg5\"]]"},
		},
		{
			name:            "viper: enabled false and global labels",
			appOpts:         newViperOpts(false, [][]string{{"arg1", "arg2"}}),
			expGlobalLabels: [][]string{{"arg1", "arg2"}},
			expEnabled:      false,
			expLogLines: []string{"DBG Using Telemetry config from app config. " + funcLogKV +
				" " + telEnabledKey + "=false " + telGlobalLabelsKey + "=[[\"arg1\",\"arg2\"]]"},
		},
		{
			name:            "viper: error parsing",
			appOpts:         newViperOpts(true, struct{}{}),
			expGlobalLabels: nil,
			expEnabled:      true,
			expLogLines: []string{
				"DBG Error parsing app config to get telemetry config (ignoring). " +
					"error=\"1 error(s) decoding:\\n\\n* '" + telGlobalLabelsKey + "[0][0]' expected type 'string', " +
					"got unconvertible type 'struct {}', value: '{}'\" " + funcLogKV,
				"DBG App option " + telGlobalLabelsKey + " is not [][]string, is struct {} " + funcLogKV +
					" " + telGlobalLabelsKey + "={}",
				"DBG Extracted telemetry setup from app options. " + funcLogKV + " " +
					telEnabledKey + "=true " + telGlobalLabelsKey + "=null",
			},
		},
		{
			name:            "not-viper: empty",
			appOpts:         newTestOpts(nil, nil),
			expGlobalLabels: nil,
			expEnabled:      false,
			expLogLines: []string{
				"DBG Extracted telemetry setup from app options. " + funcLogKV + " " +
					telEnabledKey + "=false " + telGlobalLabelsKey + "=null",
			},
		},
		{
			name:            "not-viper: only enabled true",
			appOpts:         newTestOpts(true, nil),
			expGlobalLabels: nil,
			expEnabled:      true,
			expLogLines: []string{
				"DBG Extracted telemetry setup from app options. " + funcLogKV + " " +
					telEnabledKey + "=true " + telGlobalLabelsKey + "=null",
			},
		},
		{
			name:            "not-viper: only enabled false",
			appOpts:         newTestOpts(false, nil),
			expGlobalLabels: nil,
			expEnabled:      false,
			expLogLines: []string{
				"DBG Extracted telemetry setup from app options. " + funcLogKV + " " +
					telEnabledKey + "=false " + telGlobalLabelsKey + "=null",
			},
		},
		{
			name:            "not-viper: only global labels, right type",
			appOpts:         newTestOpts(nil, [][]string{{"val1", "valtwo"}}),
			expGlobalLabels: nil, // not returned because it's not enabled.
			expEnabled:      false,
			expLogLines: []string{
				"DBG Extracted telemetry setup from app options. " + funcLogKV + " " +
					telEnabledKey + "=false " + telGlobalLabelsKey + "=null",
			},
		},
		{
			name:            "not-viper: only global labels, wrong type",
			appOpts:         newTestOpts(nil, struct{}{}),
			expGlobalLabels: nil, // not returned because it's not enabled.
			expEnabled:      false,
			expLogLines: []string{
				"DBG Extracted telemetry setup from app options. " + funcLogKV + " " +
					telEnabledKey + "=false " + telGlobalLabelsKey + "=null",
			},
		},
		{
			name:            "not-viper: enabled true and global labels, right type",
			appOpts:         newTestOpts(true, [][]string{{"val1", "valtwo"}}),
			expGlobalLabels: [][]string{{"val1", "valtwo"}}, // not returned because it's not enabled.
			expEnabled:      true,
			expLogLines: []string{
				"DBG Extracted telemetry setup from app options. " + funcLogKV + " " +
					telEnabledKey + "=true " + telGlobalLabelsKey + "=[[\"val1\",\"valtwo\"]]",
			},
		},
		{
			name:            "not-viper: enabled true and global labels, wrong type",
			appOpts:         newTestOpts(true, struct{}{}),
			expGlobalLabels: nil,
			expEnabled:      true,
			expLogLines: []string{
				"DBG App option " + telGlobalLabelsKey + " is not [][]string, is struct {} " + funcLogKV +
					" " + telGlobalLabelsKey + "={}",
				"DBG Extracted telemetry setup from app options. " + funcLogKV + " " +
					telEnabledKey + "=true " + telGlobalLabelsKey + "=null",
			},
		},
		{
			name:            "not-viper: enabled false and global labels, right type",
			appOpts:         newTestOpts(false, [][]string{{"val1", "valtwo"}}),
			expGlobalLabels: nil, // not returned because it's not enabled.
			expEnabled:      false,
			expLogLines: []string{
				"DBG Extracted telemetry setup from app options. " + funcLogKV + " " +
					telEnabledKey + "=false " + telGlobalLabelsKey + "=null",
			},
		},
		{
			name:            "not-viper: enabled false and global labels, wrong type",
			appOpts:         newTestOpts(false, struct{}{}),
			expGlobalLabels: nil, // not returned because it's not enabled.
			expEnabled:      false,
			expLogLines: []string{
				"DBG Extracted telemetry setup from app options. " + funcLogKV + " " +
					telEnabledKey + "=false " + telGlobalLabelsKey + "=null",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var logBuffer bytes.Buffer
			logger := internal.NewBufferedDebugLogger(&logBuffer)
			var actGlobalLabels [][]string
			var actEnabled bool
			testFunc := func() {
				actGlobalLabels, actEnabled = getTelemetryGlobalLabels(logger, tc.appOpts)
			}
			require.NotPanics(t, testFunc, "getTelemetryGlobalLabels")
			logs := logBuffer.String()
			assert.Equal(t, tc.expGlobalLabels, actGlobalLabels, "global labels from getTelemetryGlobalLabels")
			assert.Equal(t, tc.expEnabled, actEnabled, "enabled bool from getTelemetryGlobalLabels")
			logLines := internal.SplitLogLines(logs)
			if !assert.Equal(t, tc.expLogLines, logLines, "lines logged during getTelemetryGlobalLabels") {
				t.Logf("full logs from getTelemetryGlobalLabels:\n%s", logs)
			}
		})
	}
}

func TestTelemetryKeysCanary(t *testing.T) {
	// This test will hopefully fail if the SDK ever makes a change that would mess up getTelemetryGlobalLabels.
	// So if this test fails, you'll need to update getTelemetryGlobalLabels (and this) to account for the changes.

	// Get all of the app config field names so that we can ensure that the ones we use in
	// getTelemetryGlobalLabels still exist. Also, make sure they have the expected types.
	appCfg := serverconfig.Config{}
	appCfgMap := config.MakeFieldValueMap(appCfg, true)

	if assert.Contains(t, appCfgMap, telEnabledKey, "Field %q does not exist in the app config.", telEnabledKey) {
		expType := "bool"
		actType := fmt.Sprintf("%T", appCfg.Telemetry.Enabled)
		assert.Equal(t, expType, actType, "type of appCfg.Telemetry.Enabled")
	}

	if assert.Contains(t, appCfgMap, telGlobalLabelsKey, "Field %q does not exist in the app config.", telGlobalLabelsKey) {
		expType := "[][]string"
		actType := fmt.Sprintf("%T", appCfg.Telemetry.GlobalLabels)
		assert.Equal(t, expType, actType, "type of appCfg.Telemetry.GlobalLabels")
	}
}
