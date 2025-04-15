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

	cmtconfig "github.com/cometbft/cometbft/config"

	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/app"
	cmderrors "github.com/provenance-io/provenance/cmd/errors"
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

// cmdResult contains the info, results, and output of the executed command.
type cmdResult struct {
	Cmd      *cobra.Command
	Home     string
	Stdout   string
	Stderr   string
	Result   error
	ExitCode int
}

func executeRootCmd(t *testing.T, home string, cmdArgs ...string) *cmdResult {
	// Ensure the keyring backend doesn't get changed when we run this command.
	t.Setenv("PIO_TESTNET", "false")

	rv := &cmdResult{Home: home}

	if len(home) > 0 {
		cmdArgs = append([]string{"--home", home}, cmdArgs...)
	}

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
		var srvErrP *cmderrors.ExitCodeError
		var srvErr cmderrors.ExitCodeError
		switch {
		case errors.As(rv.Result, &srvErrP):
			rv.ExitCode = int(*srvErrP)
		case errors.As(rv.Result, &srvErr):
			rv.ExitCode = int(srvErr)
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
func makeDummyCmd(t *testing.T, cdc codec.Codec, home string) *cobra.Command {
	clientCtx := client.Context{}.
		WithCodec(cdc).
		WithHomeDir(home)
	clientCtx.Viper = viper.New()
	serverCtx := server.NewContext(clientCtx.Viper, config.DefaultCmtConfig(), log.NewNopLogger())

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
	dummyCmd.SetContext(ctx)
	require.NoError(t, config.LoadConfigFromFiles(dummyCmd), "LoadConfigFromFiles")
	return dummyCmd
}

func TestPreUpgradeCmd(t *testing.T) {
	origCache := sdk.IsAddrCacheEnabled()
	defer sdk.SetAddrCacheEnabled(origCache)
	sdk.SetAddrCacheEnabled(false)

	pioconfig.SetProvConfig("") // Undo any config changes done in any other tests before.
	encodingConfig := app.MakeTestEncodingConfig(t)
	cdc := encodingConfig.Marshaler

	tmpDir := t.TempDir()

	seenNames := make(map[string]bool)
	// newHome creates a new home directory and saves the configs. Returns full path to home and success.
	newHome := func(t *testing.T, name string, appCfg *serverconfig.Config, cmtCfg *cmtconfig.Config, clientCfg *config.ClientConfig) (string, bool) {
		require.False(t, seenNames[name], "dir name %q created in previous test", name)
		seenNames[name] = true
		home := filepath.Join(tmpDir, name)
		if !assert.NoError(t, os.MkdirAll(home, 0o755), "MkdirAll") {
			return home, false
		}

		dummyCmd := makeDummyCmd(t, cdc, home)
		success := assert.NotPanics(t, func() { config.SaveConfigs(dummyCmd, appCfg, cmtCfg, clientCfg, false) }, "SaveConfigs")
		return home, success
	}
	// newHomePacked creates a new home directory, saves the configs, and packs them. Returns full path to home and success.
	newHomePacked := func(t *testing.T, name string, appCfg *serverconfig.Config, cmtCfg *cmtconfig.Config, clientCfg *config.ClientConfig) (string, bool) {
		home, success := newHome(t, name, appCfg, cmtCfg, clientCfg)
		if !success {
			return home, success
		}

		dummyCmd := makeDummyCmd(t, cdc, home)
		success = assert.NoError(t, config.PackConfig(dummyCmd), "PackConfig")
		return home, success
	}

	appCfgD := config.DefaultAppConfig()
	cmtCfgD := config.DefaultCmtConfig()
	cmtCfgD.Consensus.TimeoutCommit = 3500 * time.Millisecond
	clientCfgD := config.DefaultClientConfig()

	appCfgC := config.DefaultAppConfig()
	appCfgC.API.Enable = true
	appCfgC.API.Swagger = true
	cmtCfgC := config.DefaultCmtConfig()
	cmtCfgC.Consensus.TimeoutCommit = cmtCfgD.Consensus.TimeoutCommit + 123*time.Millisecond
	cmtCfgC.LogLevel = "debug"
	clientCfgC := config.DefaultClientConfig()
	clientCfgC.Output = "json"
	clientCfgC.ChainID = "pio-mainnet-1"
	clientCfgAsync := config.DefaultClientConfig()
	clientCfgAsync.BroadcastMode = "async"

	cmtCfgT := config.DefaultCmtConfig()
	cmtCfgT.Consensus.TimeoutCommit = 777 * time.Second

	successMsg := "pre-upgrade successful"
	updatingBlocksyncMsg := "Updating the broadcast_mode config value to \"sync\" (from \"block\", which is no longer an option)."

	tests := []struct {
		name         string
		setup        func(t *testing.T) (string, func(), bool) // returns home dir, a deferrable and success.
		args         []string
		expExitCode  int
		expInStdout  []string
		expInStderr  []string
		expNot       []string
		expAppCfg    *serverconfig.Config
		expCmtCfg    *cmtconfig.Config
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
			expCmtCfg:    cmtCfgD,
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
				home, success := newHome(t, "defaults_unpacked", appCfgD, cmtCfgD, clientCfgD)
				return home, nil, success
			},
			expExitCode:  0,
			expInStdout:  []string{successMsg},
			expAppCfg:    appCfgD,
			expCmtCfg:    cmtCfgD,
			expClientCfg: clientCfgD,
		},
		{
			name: "packed config with defaults",
			setup: func(t *testing.T) (string, func(), bool) {
				home, success := newHomePacked(t, "defaults_packed", appCfgD, cmtCfgD, clientCfgD)
				return home, nil, success
			},
			expExitCode:  0,
			expInStdout:  []string{successMsg},
			expAppCfg:    appCfgD,
			expCmtCfg:    cmtCfgD,
			expClientCfg: clientCfgD,
		},
		{
			name: "cannot write new file",
			setup: func(t *testing.T) (string, func(), bool) {
				home, success := newHomePacked(t, "mainnet_cannot_write", appCfgC, cmtCfgC, clientCfgC)
				if !success {
					return home, nil, success
				}

				dummyCmd := makeDummyCmd(t, cdc, home)
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
		{
			name: "packed config broadcast block",
			setup: func(t *testing.T) (string, func(), bool) {
				clientCfg := config.DefaultClientConfig()
				clientCfg.BroadcastMode = "block"
				home, success := newHomePacked(t, "packed_broadcast_block", appCfgD, cmtCfgD, clientCfg)
				return home, nil, success
			},
			expExitCode:  0,
			expInStdout:  []string{updatingBlocksyncMsg, successMsg},
			expAppCfg:    appCfgD,
			expCmtCfg:    cmtCfgD,
			expClientCfg: clientCfgD,
		},
		{
			name: "packed config broadcast sync",
			setup: func(t *testing.T) (string, func(), bool) {
				clientCfg := config.DefaultClientConfig()
				clientCfg.BroadcastMode = "sync"
				home, success := newHomePacked(t, "packed_broadcast_sync", appCfgD, cmtCfgD, clientCfg)
				return home, nil, success
			},
			expExitCode:  0,
			expInStdout:  []string{successMsg},
			expNot:       []string{updatingBlocksyncMsg},
			expAppCfg:    appCfgD,
			expCmtCfg:    cmtCfgD,
			expClientCfg: clientCfgD,
		},
		{
			name: "packed config broadcast async",
			setup: func(t *testing.T) (string, func(), bool) {
				clientCfg := config.DefaultClientConfig()
				clientCfg.BroadcastMode = "async"
				home, success := newHomePacked(t, "packed_broadcast_async", appCfgD, cmtCfgD, clientCfg)
				return home, nil, success
			},
			expExitCode:  0,
			expInStdout:  []string{successMsg},
			expNot:       []string{updatingBlocksyncMsg},
			expAppCfg:    appCfgD,
			expCmtCfg:    cmtCfgD,
			expClientCfg: clientCfgAsync,
		},
		{
			name: "unpacked config broadcast block",
			setup: func(t *testing.T) (string, func(), bool) {
				clientCfg := config.DefaultClientConfig()
				clientCfg.BroadcastMode = "block"
				home, success := newHome(t, "unpacked_broadcast_block", appCfgD, cmtCfgD, clientCfg)
				return home, nil, success
			},
			expExitCode:  0,
			expInStdout:  []string{updatingBlocksyncMsg, successMsg},
			expAppCfg:    appCfgD,
			expCmtCfg:    cmtCfgD,
			expClientCfg: clientCfgD,
		},
		{
			name: "unpacked config broadcast sync",
			setup: func(t *testing.T) (string, func(), bool) {
				clientCfg := config.DefaultClientConfig()
				clientCfg.BroadcastMode = "sync"
				home, success := newHome(t, "unpacked_broadcast_sync", appCfgD, cmtCfgD, clientCfg)
				return home, nil, success
			},
			expExitCode:  0,
			expInStdout:  []string{successMsg},
			expNot:       []string{updatingBlocksyncMsg},
			expAppCfg:    appCfgD,
			expCmtCfg:    cmtCfgD,
			expClientCfg: clientCfgD,
		},
		{
			name: "unpacked config broadcast async",
			setup: func(t *testing.T) (string, func(), bool) {
				clientCfg := config.DefaultClientConfig()
				clientCfg.BroadcastMode = "async"
				home, success := newHome(t, "unpacked_broadcast_async", appCfgD, cmtCfgD, clientCfg)
				return home, nil, success
			},
			expExitCode:  0,
			expInStdout:  []string{successMsg},
			expNot:       []string{updatingBlocksyncMsg},
			expAppCfg:    appCfgD,
			expCmtCfg:    cmtCfgD,
			expClientCfg: clientCfgAsync,
		},
		{
			name: "unpacked mainnet timeout commit",
			setup: func(t *testing.T) (string, func(), bool) {
				home, success := newHome(t, "unpacked_mainnet_timeout_commit", appCfgD, cmtCfgT, clientCfgD)
				return home, nil, success
			},
			expExitCode:  0,
			expInStdout:  []string{successMsg},
			expAppCfg:    appCfgD,
			expCmtCfg:    cmtCfgD,
			expClientCfg: clientCfgD,
		},
		{
			name: "packed mainnet timeout commit",
			setup: func(t *testing.T) (string, func(), bool) {
				home, success := newHomePacked(t, "packed_mainnet_timeout_commit", appCfgD, cmtCfgT, clientCfgD)
				return home, nil, success
			},
			expExitCode:  0,
			expInStdout:  []string{successMsg},
			expAppCfg:    appCfgD,
			expCmtCfg:    cmtCfgD,
			expClientCfg: clientCfgD,
		},
		{
			name: "unpacked testnet timeout commit",
			setup: func(t *testing.T) (string, func(), bool) {
				home, success := newHome(t, "unpacked_testnet_timeout_commit", appCfgD, cmtCfgT, clientCfgD)
				return home, nil, success
			},
			args:         []string{"--testnet"},
			expExitCode:  0,
			expInStdout:  []string{successMsg},
			expAppCfg:    appCfgD,
			expCmtCfg:    cmtCfgT,
			expClientCfg: clientCfgD,
		},
		{
			name: "packed testnet timeout commit",
			setup: func(t *testing.T) (string, func(), bool) {
				home, success := newHomePacked(t, "packed_testnet_timeout_commit", appCfgD, cmtCfgT, clientCfgD)
				return home, nil, success
			},
			args:         []string{"--testnet"},
			expExitCode:  0,
			expInStdout:  []string{successMsg},
			expAppCfg:    appCfgD,
			expCmtCfg:    cmtCfgT,
			expClientCfg: clientCfgD,
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

			dummyCmd := makeDummyCmd(t, cdc, home)
			appCfg, err := config.ExtractAppConfig(dummyCmd)
			if assert.NoError(t, err, "ExtractAppConfig") {
				assert.Equal(t, tc.expAppCfg, appCfg, "app config")
			}
			cmtCfg, err := config.ExtractCmtConfig(dummyCmd)
			cmtCfg.SetRoot("")
			if assert.NoError(t, err, "ExtractCmtConfig") {
				assert.Equal(t, tc.expCmtCfg, cmtCfg, "cmt config")
			}
			clientCfg, err := config.ExtractClientConfig(dummyCmd)
			if assert.NoError(t, err, "ExtractClientConfig") {
				assert.Equal(t, tc.expClientCfg, clientCfg)
			}
		})
	}
}
