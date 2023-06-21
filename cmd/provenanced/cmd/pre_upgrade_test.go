package cmd_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tmconfig "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdksim "github.com/cosmos/cosmos-sdk/simapp"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/cmd/provenanced/cmd"
	"github.com/provenance-io/provenance/cmd/provenanced/config"
	"github.com/provenance-io/provenance/internal/pioconfig"
)

// logIfError logs an error if it's not nil.
// The error is automatically added to the format and args.
// Use this if there's a possible error that we probably don't care about (but might).
func logIfError(t *testing.T, err error, format string, args ...interface{}) {
	if err != nil {
		t.Logf(format+" error: %v", append(args, err)...)
	}
}

func logDir(t *testing.T, path string) {
	contents, err := os.ReadDir(path)
	if os.IsNotExist(err) {
		t.Logf("Directory does not exist: %s", path)
		return
	}
	if err != nil {
		t.Logf("Error reading directory: %v", err)
		return
	}

	var files []string
	var dirs []string
	for _, entry := range contents {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		} else {
			files = append(files, entry.Name())
		}
	}

	if len(dirs) == 0 {
		t.Logf("No subdirectories: %s", path)
	} else {
		t.Logf("Subdirectories: %s\n  %s", path, strings.Join(dirs, "  \n"))
	}
	if len(files) == 0 {
		t.Logf("No files: %s", path)
	} else {
		t.Logf("Directory files: %s\n  %s", path, strings.Join(files, "  \n"))
	}

	for _, file := range files {
		fullPath := filepath.Join(path, file)
		fileContents, err := os.ReadFile(fullPath)
		if err != nil {
			t.Logf("Error reading file: %v", err)
		}
		t.Logf("File Contents: %s\n%s", fullPath, fileContents)
	}

	for _, dir := range dirs {
		logDir(t, filepath.Join(path, dir))
	}
}

// cmdResult contains the info, results, and output of the an executed command.
type cmdResult struct {
	Cmd      *cobra.Command
	Home     string
	Stdout   string
	Stderr   string
	Result   error
	ExitCode int
}

func executeRootCmd(t *testing.T, home string, cmdArgs ...string) *cmdResult {
	rv := &cmdResult{Home: home}

	cmdArgs = append([]string{"--home", home}, cmdArgs...)

	// Use our own buffers for the output so we can capture it.
	var outBuf, errBuf bytes.Buffer
	rv.Cmd, _ = cmd.NewRootCmd(false)
	rv.Cmd.SetOut(&outBuf)
	rv.Cmd.SetErr(&errBuf)
	rv.Cmd.SetArgs(cmdArgs)

	t.Logf("Executing: %s", strings.Join(append([]string{rv.Cmd.Name()}, cmdArgs...), " "))
	// This is similar logic to main.go, but just capturing the exit code instead of calling os.Exit.
	rv.Result = cmd.Execute(rv.Cmd)
	if rv.Result != nil {
		t.Logf("Execution resulting in error: %v", rv.Result)
		var srvErrP *server.ErrorCode
		var srvErr server.ErrorCode
		switch {
		case errors.As(rv.Result, &srvErrP):
			rv.ExitCode = srvErrP.Code
		case errors.As(rv.Result, &srvErr):
			rv.ExitCode = srvErr.Code
		default:
			rv.ExitCode = 1
		}
	} else {
		t.Logf("Execution completed successfully.")
	}
	t.Logf("exit code: %d", rv.ExitCode)

	rv.Stdout = outBuf.String()
	if len(rv.Stdout) == 0 {
		t.Logf("Nothing printed to stdout.")
	} else {
		t.Logf("stdout:\n%s", rv.Stdout)
	}

	rv.Stderr = errBuf.String()
	if len(rv.Stderr) == 0 {
		t.Logf("Nothing printed to stderr.")
	} else {
		t.Logf("stderr:\n%s", rv.Stderr)
	}

	return rv
}

func executePreUpgradeCmd(t *testing.T, home string, cmdArgs ...string) *cmdResult {
	return executeRootCmd(t, home, append([]string{"pre-upgrade"}, cmdArgs...)...)
}

// assertContainsAll asserts that all of the expected entries are substrings of the actual.
// Returns true if that is the case, false = failure.
func assertContainsAll(t *testing.T, actual string, expected []string, msgAndArgs ...interface{}) bool {
	t.Helper()
	missing := make(map[int]bool)
	for i, exp := range expected {
		if !strings.Contains(actual, exp) {
			missing[i] = true
		}
	}

	if len(missing) == 0 {
		return true
	}

	lines := []string{
		fmt.Sprintf("Missing %d expected substring(s).", len(missing)),
		fmt.Sprintf("Actual:\n%s", actual),
		"Expected:",
	}
	for i, exp := range expected {
		result := "( found )"
		if missing[i] {
			result = "(missing)"
		}
		lines = append(lines, fmt.Sprintf("[%d] %s: %q", i, result, exp))
	}

	return assert.Fail(t, strings.Join(lines, "\n"), msgAndArgs...)
}

// makeDummyCmd creates a dummy command with a context in it that can be used to test all the config stuff.
func makeDummyCmd(t *testing.T, home string) *cobra.Command {
	encodingConfig := sdksim.MakeTestEncodingConfig()
	clientCtx := client.Context{}.
		WithCodec(encodingConfig.Codec).
		WithHomeDir(home)
	clientCtx.Viper = viper.New()
	serverCtx := server.NewContext(clientCtx.Viper, config.DefaultTmConfig(), log.NewNopLogger())

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

	dummyCmd := &cobra.Command{
		Use:   "dummy",
		Short: "Just used for testing. Doesn't do anything.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	dummyCmd.SetOut(io.Discard)
	dummyCmd.SetErr(io.Discard)
	dummyCmd.SetArgs([]string{})
	var err error
	dummyCmd, err = dummyCmd.ExecuteContextC(ctx)
	require.NoError(t, err, "dummy command execution")
	require.NoError(t, config.LoadConfigFromFiles(dummyCmd), "LoadConfigFromFiles")
	return dummyCmd
}

func TestPreUpgradeCmd(t *testing.T) {
	app.SetConfig(true, false)
	pioconfig.SetProvenanceConfig("", 0)

	tmpDir := t.TempDir()

	seenNames := make(map[string]bool)
	// newHome creates a new home directory and saves the configs. Returns full path to home and success.
	newHome := func(t *testing.T, name string, appCfg *serverconfig.Config, tmCfg *tmconfig.Config, clientCfg *config.ClientConfig) (string, bool) {
		require.False(t, seenNames[name], "dir name %q created in previous test", name)
		seenNames[name] = true
		home := filepath.Join(tmpDir, name)
		if !assert.NoError(t, os.MkdirAll(home, 0o755), "MkdirAll") {
			return home, false
		}

		dummyCmd := makeDummyCmd(t, home)
		success := assert.NotPanics(t, func() { config.SaveConfigs(dummyCmd, appCfg, tmCfg, clientCfg, false) }, "SaveConfigs")
		return home, success
	}
	// newHomePacked creates a new home directory, saves the configs, and packs them. Returns full path to home and success.
	newHomePacked := func(t *testing.T, name string, appCfg *serverconfig.Config, tmCfg *tmconfig.Config, clientCfg *config.ClientConfig) (string, bool) {
		home, success := newHome(t, name, appCfg, tmCfg, clientCfg)
		if !success {
			return home, success
		}

		dummyCmd := makeDummyCmd(t, home)
		success = assert.NoError(t, config.PackConfig(dummyCmd), "PackConfig")
		return home, success
	}

	appCfgD := config.DefaultAppConfig()
	tmCfgD := config.DefaultTmConfig()
	clientCfgD := config.DefaultClientConfig()

	appCfgC := config.DefaultAppConfig()
	appCfgC.API.Enable = true
	appCfgC.API.Swagger = true
	tmCfgLower := config.DefaultTmConfig()
	tmCfgLower.Consensus.TimeoutCommit = tmCfgD.Consensus.TimeoutCommit - 500*time.Millisecond
	tmCfgLower.LogLevel = "debug"
	tmCfgHigher := config.DefaultTmConfig()
	tmCfgHigher.Consensus.TimeoutCommit = tmCfgD.Consensus.TimeoutCommit + 500*time.Millisecond
	tmCfgHigher.LogLevel = "debug"
	clientCfgC := config.DefaultClientConfig()
	clientCfgC.Output = "json"
	clientCfgMainnetC := config.DefaultClientConfig()
	clientCfgMainnetC.Output = "json"
	clientCfgMainnetC.ChainID = "pio-mainnet-1"

	tmCfgCFixed := config.DefaultTmConfig()
	tmCfgCFixed.LogLevel = "debug"

	successMsg := "pre-upgrade successful"
	updatingMsg := func(old time.Duration) string {
		return fmt.Sprintf("Updating consensus.timeout_commit config value to %q (from %q)", config.DefaultConsensusTimeoutCommit, old)
	}

	tests := []struct {
		name         string
		setup        func(t *testing.T) (string, func(), bool) // returns home dir, a deferrable and success.
		args         []string
		expExitCode  int
		expInStdout  []string
		expInStderr  []string
		expNot       []string
		expAppCfg    *serverconfig.Config
		expTmCfg     *tmconfig.Config
		expClientCfg *config.ClientConfig
	}{
		{
			name: "home dir does not exist yet",
			setup: func(t *testing.T) (string, func(), bool) {
				return filepath.Join(tmpDir, "dne"), nil, true
			},
			expExitCode:  0,
			expInStdout:  []string{successMsg},
			expAppCfg:    appCfgD,
			expTmCfg:     tmCfgD,
			expClientCfg: clientCfgD,
		},
		{
			name: "arg provided",
			setup: func(t *testing.T) (string, func(), bool) {
				return filepath.Join(tmpDir, "dne_with_arg"), nil, true
			},
			args:        []string{"arg1"},
			expExitCode: 30,
			expInStderr: []string{"expected 0 args, received 1"},
		},
		{
			name: "unpacked config with defaults",
			setup: func(t *testing.T) (string, func(), bool) {
				home, success := newHome(t, "defaults_unpacked", appCfgD, tmCfgD, clientCfgD)
				return home, nil, success
			},
			expExitCode:  0,
			expInStdout:  []string{successMsg},
			expAppCfg:    appCfgD,
			expTmCfg:     tmCfgD,
			expClientCfg: clientCfgD,
		},
		{
			name: "packed config with defaults",
			setup: func(t *testing.T) (string, func(), bool) {
				home, success := newHomePacked(t, "defaults_packed", appCfgD, tmCfgD, clientCfgD)
				return home, nil, success
			},
			expExitCode:  0,
			expInStdout:  []string{successMsg},
			expAppCfg:    appCfgD,
			expTmCfg:     tmCfgD,
			expClientCfg: clientCfgD,
		},
		{
			name: "mainnet unpacked config lower than default",
			setup: func(t *testing.T) (string, func(), bool) {
				home, success := newHome(t, "mainnet_unpacked_lower", appCfgC, tmCfgLower, clientCfgMainnetC)
				return home, nil, success
			},
			expExitCode: 0,
			expInStdout: []string{
				updatingMsg(tmCfgLower.Consensus.TimeoutCommit),
				successMsg,
			},
			expAppCfg:    appCfgC,
			expTmCfg:     tmCfgCFixed,
			expClientCfg: clientCfgMainnetC,
		},
		{
			name: "mainnet unpacked config higher than default",
			setup: func(t *testing.T) (string, func(), bool) {
				home, success := newHome(t, "mainnet_unpacked_higher", appCfgC, tmCfgHigher, clientCfgMainnetC)
				return home, nil, success
			},
			expExitCode: 0,
			expInStdout: []string{
				updatingMsg(tmCfgHigher.Consensus.TimeoutCommit),
				successMsg,
			},
			expAppCfg:    appCfgC,
			expTmCfg:     tmCfgCFixed,
			expClientCfg: clientCfgMainnetC,
		},
		{
			name: "mainnet packed config lower than default",
			setup: func(t *testing.T) (string, func(), bool) {
				home, success := newHomePacked(t, "mainnet_packed_lower", appCfgC, tmCfgLower, clientCfgMainnetC)
				return home, nil, success
			},
			expExitCode: 0,
			expInStdout: []string{
				updatingMsg(tmCfgLower.Consensus.TimeoutCommit),
				successMsg,
			},
			expAppCfg:    appCfgC,
			expTmCfg:     tmCfgCFixed,
			expClientCfg: clientCfgMainnetC,
		},
		{
			name: "mainnet packed config higher than default",
			setup: func(t *testing.T) (string, func(), bool) {
				home, success := newHomePacked(t, "mainnet_packed_higher", appCfgC, tmCfgHigher, clientCfgMainnetC)
				return home, nil, success
			},
			expExitCode: 0,
			expInStdout: []string{
				updatingMsg(tmCfgHigher.Consensus.TimeoutCommit),
				successMsg,
			},
			expAppCfg:    appCfgC,
			expTmCfg:     tmCfgCFixed,
			expClientCfg: clientCfgMainnetC,
		},
		{
			name: "mainnet cannot write new file",
			setup: func(t *testing.T) (string, func(), bool) {
				home, success := newHomePacked(t, "mainnet_cannot_write", appCfgC, tmCfgLower, clientCfgMainnetC)
				if !success {
					return home, nil, success
				}

				dummyCmd := makeDummyCmd(t, home)
				unwritableFile := config.GetFullPathToPackedConf(dummyCmd)
				success = assert.NoError(t, os.Chmod(unwritableFile, 0o444), "Chmod")
				deferrable := func() {
					logIfError(t, os.Chmod(unwritableFile, 0o666), "Changing permissions on %s so it can be deleted", unwritableFile)
				}
				return home, deferrable, success
			},
			expExitCode: 30,
			expInStdout: []string{updatingMsg(tmCfgLower.Consensus.TimeoutCommit)},
			expInStderr: []string{"could not update config file(s):", "error saving config file(s):", "permission denied"},
		},
		{
			name: "other net unpacked config lower than default",
			setup: func(t *testing.T) (string, func(), bool) {
				home, success := newHome(t, "other_unpacked_lower", appCfgC, tmCfgLower, clientCfgC)
				return home, nil, success
			},
			expExitCode:  0,
			expInStdout:  []string{successMsg},
			expNot:       []string{"Updating consensus.timeout_commit config value"},
			expAppCfg:    appCfgC,
			expTmCfg:     tmCfgLower,
			expClientCfg: clientCfgC,
		},
		{
			name: "other net unpacked config higher than default",
			setup: func(t *testing.T) (string, func(), bool) {
				home, success := newHome(t, "other_unpacked_higher", appCfgC, tmCfgHigher, clientCfgC)
				return home, nil, success
			},
			expExitCode:  0,
			expInStdout:  []string{successMsg},
			expNot:       []string{"Updating consensus.timeout_commit config value"},
			expAppCfg:    appCfgC,
			expTmCfg:     tmCfgHigher,
			expClientCfg: clientCfgC,
		},
		{
			name: "other net packed config lower than default",
			setup: func(t *testing.T) (string, func(), bool) {
				home, success := newHomePacked(t, "other_packed_lower", appCfgC, tmCfgLower, clientCfgC)
				return home, nil, success
			},
			expExitCode:  0,
			expInStdout:  []string{successMsg},
			expNot:       []string{"Updating consensus.timeout_commit config value"},
			expAppCfg:    appCfgC,
			expTmCfg:     tmCfgLower,
			expClientCfg: clientCfgC,
		},
		{
			name: "other net packed config higher than default",
			setup: func(t *testing.T) (string, func(), bool) {
				home, success := newHomePacked(t, "other_packed_higher", appCfgC, tmCfgHigher, clientCfgC)
				return home, nil, success
			},
			expExitCode:  0,
			expInStdout:  []string{successMsg},
			expNot:       []string{"Updating consensus.timeout_commit config value"},
			expAppCfg:    appCfgC,
			expTmCfg:     tmCfgHigher,
			expClientCfg: clientCfgC,
		},
		{
			name: "other net cannot write new file",
			setup: func(t *testing.T) (string, func(), bool) {
				home, success := newHomePacked(t, "other_cannot_write", appCfgC, tmCfgLower, clientCfgC)
				if !success {
					return home, nil, success
				}

				dummyCmd := makeDummyCmd(t, home)
				unwritableFile := config.GetFullPathToPackedConf(dummyCmd)
				success = assert.NoError(t, os.Chmod(unwritableFile, 0o444), "Chmod")
				deferrable := func() {
					logIfError(t, os.Chmod(unwritableFile, 0o666), "Changing permissions on %s so it can be deleted", unwritableFile)
				}
				return home, deferrable, success
			},
			expExitCode: 30,
			expInStderr: []string{"could not update config file(s):", "error saving config file(s):", "permission denied"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			home, deferrable, ok := tc.setup(t)
			if deferrable != nil {
				defer deferrable()
			}
			if !ok {
				t.Logf("Setup failed.")
				return
			}
			logDir(t, home)

			res := executePreUpgradeCmd(t, home, tc.args...)
			assert.Equal(t, tc.expExitCode, res.ExitCode, "exit code")

			assertContainsAll(t, res.Stdout, tc.expInStdout, "stdout")
			if len(tc.expInStderr) == 0 {
				assert.Empty(t, tc.expInStderr, "stderr")
			} else {
				assertContainsAll(t, res.Stderr, tc.expInStderr, "stderr")
			}
			for _, unexp := range tc.expNot {
				assert.NotContains(t, res.Stdout, unexp, "stdout")
				assert.NotContains(t, res.Stderr, unexp, "stderr")
			}

			if res.ExitCode != 0 {
				return
			}

			dummyCmd := makeDummyCmd(t, home)
			appCfg, err := config.ExtractAppConfig(dummyCmd)
			if assert.NoError(t, err, "ExtractAppConfig") {
				assert.Equal(t, tc.expAppCfg, appCfg, "app config")
			}
			tmCfg, err := config.ExtractTmConfig(dummyCmd)
			tmCfg.SetRoot("")
			if assert.NoError(t, err, "ExtractTmConfig") {
				assert.Equal(t, tc.expTmCfg, tmCfg, "tm config")
			}
			clientCfg, err := config.ExtractClientConfig(dummyCmd)
			if assert.NoError(t, err, "ExtractClientConfig") {
				assert.Equal(t, tc.expClientCfg, clientCfg)
			}
		})
	}
}
