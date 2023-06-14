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

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdksim "github.com/cosmos/cosmos-sdk/simapp"
	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/cmd/provenanced/cmd"
	"github.com/provenance-io/provenance/cmd/provenanced/config"
	"github.com/provenance-io/provenance/internal/pioconfig"
	tmconfig "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/log"
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

// CmdResult contains the info, results, and output of the an executed command.
type CmdResult struct {
	Cmd      *cobra.Command
	Home     string
	Stdout   string
	Stderr   string
	Result   error
	ExitCode int
}

func executeRootCmd(t *testing.T, home string, cmdArgs ...string) *CmdResult {
	rv := &CmdResult{Home: home}

	origHome := app.DefaultNodeHome
	defer func() {
		app.DefaultNodeHome = origHome
	}()
	t.Logf("Setting DefaultNodeHome = %q", home)
	app.DefaultNodeHome = home

	// Use our own buffers for the output so we can capture it.
	var outBuf, errBuf bytes.Buffer
	rv.Cmd, _ = cmd.NewRootCmd(false)
	rv.Cmd.SetOut(&outBuf)
	rv.Cmd.SetErr(&errBuf)
	rv.Cmd.SetArgs(cmdArgs)

	t.Logf("Executing: %s", strings.Join(append([]string{rv.Cmd.Name()}, cmdArgs...), " "))
	// This is similar logic to main.go, but just capturing the exit code instead of calling os.Exit.
	if rv.Result = cmd.Execute(rv.Cmd); rv.Result != nil {
		var srvErrP *server.ErrorCode
		var srvErr server.ErrorCode
		switch {
		case errors.As(rv.Result, &srvErrP):
			rv.ExitCode = srvErrP.Code
		case errors.As(rv.Result, &srvErr):
			rv.ExitCode = srvErr.Code
		default:
			os.Exit(1)
		}
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

func executePreUpgradeCmd(t *testing.T, home string, cmdArgs ...string) *CmdResult {
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
	app.SetConfig(true, true)
	pioconfig.SetProvenanceConfig("", 0)

	appCfgD := config.DefaultAppConfig()
	tmCfgD := config.DefaultTmConfig()
	clientCfgD := config.DefaultClientConfig()

	appCfgC := config.DefaultAppConfig()
	appCfgC.API.Enable = true
	appCfgC.API.Swagger = true
	tmCfgC := config.DefaultTmConfig()
	tmCfgC.Consensus.TimeoutCommit = 1 * time.Second
	tmCfgC.LogLevel = "debug"
	clientCfgC := config.DefaultClientConfig()
	clientCfgC.Output = "json"

	tmCfgFixed := config.DefaultTmConfig()
	tmCfgFixed.LogLevel = "debug"

	var homeDefaultsUnpacked, homeDefaultsPacked string
	var homeChangedUnpacked, homeChangedPacked string
	var homeDNE string

	tmpDir := t.TempDir()

	t.Run("setup: defaults unpacked", func(t *testing.T) {
		home := filepath.Join(tmpDir, "defaults_unpacked")
		homeDefaultsUnpacked = home
		t.Logf("Creating: %s", home)
		require.NoError(t, os.MkdirAll(home, 0o755), "MkdirAll")

		t.Logf("Saving configs")
		dummyCmd := makeDummyCmd(t, home)
		require.NotPanics(t, func() { config.SaveConfigs(dummyCmd, appCfgD, tmCfgD, clientCfgD, false) }, "SaveConfigs")
		logDir(t, home)
	})

	t.Run("setup: defaults packed", func(t *testing.T) {
		home := filepath.Join(tmpDir, "defaults_packed")
		homeDefaultsPacked = home
		t.Logf("Creating: %s", home)
		require.NoError(t, os.MkdirAll(home, 0o755), "MkdirAll")

		t.Logf("Saving configs")
		dummyCmd := makeDummyCmd(t, home)
		require.NotPanics(t, func() { config.SaveConfigs(dummyCmd, appCfgD, tmCfgD, clientCfgD, false) }, "SaveConfigs")
		logDir(t, home)

		t.Logf("Packing configs")
		dummyCmd = makeDummyCmd(t, home)
		require.NoError(t, config.PackConfig(dummyCmd), "PackConfig")
		logDir(t, home)
	})

	t.Run("setup: changed unpacked", func(t *testing.T) {
		home := filepath.Join(tmpDir, "changed_unpacked")
		homeChangedUnpacked = home
		t.Logf("Creating: %s", home)
		require.NoError(t, os.MkdirAll(home, 0o755), "MkdirAll")

		t.Logf("Saving configs")
		dummyCmd := makeDummyCmd(t, home)
		require.NotPanics(t, func() { config.SaveConfigs(dummyCmd, appCfgC, tmCfgC, clientCfgC, false) }, "SaveConfigs")
		logDir(t, home)
	})

	t.Run("setup: changed packed", func(t *testing.T) {
		home := filepath.Join(tmpDir, "changed_packed")
		homeChangedPacked = home
		t.Logf("Creating: %s", home)
		require.NoError(t, os.MkdirAll(home, 0o755), "MkdirAll")

		t.Logf("Saving configs")
		dummyCmd := makeDummyCmd(t, home)
		require.NotPanics(t, func() { config.SaveConfigs(dummyCmd, appCfgC, tmCfgC, clientCfgC, false) }, "SaveConfigs")
		logDir(t, home)

		t.Logf("Packing configs")
		dummyCmd = makeDummyCmd(t, home)
		require.NoError(t, config.PackConfig(dummyCmd), "PackConfig")
		logDir(t, home)
	})

	t.Run("setup: home does not exist", func(t *testing.T) {
		homeDNE = filepath.Join(tmpDir, "dne")
		t.Logf("Not creating: %s", homeDNE)

		logDir(t, homeDNE)
	})

	success := "pre-upgrade successful"

	tests := []struct {
		name         string
		home         string
		args         []string
		expExitCode  int
		expInStdout  []string
		expInStderr  []string
		expAppCfg    *serverconfig.Config
		expTmCfg     *tmconfig.Config
		expClientCfg *config.ClientConfig
	}{
		{
			name:         "home dir does not exist yet",
			home:         homeDNE,
			expExitCode:  0,
			expInStdout:  []string{success},
			expAppCfg:    appCfgD,
			expTmCfg:     tmCfgD,
			expClientCfg: clientCfgD,
		},
		{
			name:         "unpacked config with defaults",
			home:         homeDefaultsUnpacked,
			expExitCode:  0,
			expInStdout:  []string{success},
			expAppCfg:    appCfgD,
			expTmCfg:     tmCfgD,
			expClientCfg: clientCfgD,
		},
		{
			name:         "packed config with defaults",
			home:         homeDefaultsPacked,
			expExitCode:  0,
			expInStdout:  []string{success},
			expAppCfg:    appCfgD,
			expTmCfg:     tmCfgD,
			expClientCfg: clientCfgD,
		},
		{
			name:        "unpacked config with changes",
			home:        homeChangedUnpacked,
			expExitCode: 0,
			expInStdout: []string{
				`Updating consensus.timeout_commit config value to "5s" (from "1s")`,
				success,
			},
			expAppCfg:    appCfgC,
			expTmCfg:     tmCfgFixed,
			expClientCfg: clientCfgC,
		},
		{
			name:        "packed config with changes",
			home:        homeChangedPacked,
			expExitCode: 0,
			expInStdout: []string{
				`Updating consensus.timeout_commit config value to "5s" (from "1s")`,
				success,
			},
			expAppCfg:    appCfgC,
			expTmCfg:     tmCfgFixed,
			expClientCfg: clientCfgC,
		},
		{
			name:        "arg provided",
			home:        homeDNE,
			args:        []string{"arg1"},
			expExitCode: 30,
			expInStderr: []string{"expected 0 args, received 1"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := executePreUpgradeCmd(t, tc.home, tc.args...)
			assert.Equal(t, tc.expExitCode, res.ExitCode, "exit code")

			assertContainsAll(t, res.Stdout, tc.expInStdout, "stdout")
			if len(tc.expInStderr) == 0 {
				assert.Empty(t, tc.expInStderr, "stderr")
			} else {
				assertContainsAll(t, res.Stderr, tc.expInStderr, "stderr")
			}

			if res.ExitCode != 0 {
				return
			}

			dummyCmd := makeDummyCmd(t, tc.home)
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
