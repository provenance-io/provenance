package cmd_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"

	"github.com/provenance-io/provenance/cmd/provenanced/cmd"
	"github.com/provenance-io/provenance/testutil/assertions"
)

func TestInitCmd(t *testing.T) {
	home := t.TempDir()
	rootCmd, _ := cmd.NewRootCmd(false)
	rootCmd.SetArgs([]string{
		"--home", home,
		"init",        // Test the init cmd
		"simapp-test", // Moniker
		fmt.Sprintf("--%s=%s", cli.FlagOverwrite, "true"), // Overwrite genesis.json, in case it already exists
	})
	rootCmd.SetOut(io.Discard)
	rootCmd.SetErr(io.Discard)

	err := cmd.Execute(rootCmd)
	require.NoError(t, err)
}

func TestGenAutoCompleteCmd(t *testing.T) {
	home := t.TempDir()

	tests := []struct {
		name string
		args []string
		err  string
	}{
		{
			name: "failure - missing arg",
			err:  "accepts 1 arg(s), received 0",
		},
		{
			name: "failure - too many args",
			args: []string{"bash", "fish"},
			err:  "accepts 1 arg(s), received 2",
		},
		{
			name: "failure - invalid shell type",
			args: []string{"badshellname"},
			err:  "shell badshellname is not supported",
		},
		{
			name: "success - works with bash",
			args: []string{"bash"},
		},
		{
			name: "success - works with zsh",
			args: []string{"zsh"},
		},
		{
			name: "success - works with fish",
			args: []string{"fish"},
		},
		{
			name: "success - works with powershell",
			args: []string{"powershell"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := []string{"--home", home, "enable-cli-autocomplete"}
			args = append(args, tc.args...)

			rootCmd, _ := cmd.NewRootCmd(false)
			rootCmd.SetArgs(args)
			rootCmd.SetOut(io.Discard)
			rootCmd.SetErr(io.Discard)

			err := cmd.Execute(rootCmd)
			assertions.AssertErrorValue(t, err, tc.err, "should have the correct output value")
		})
	}
}
