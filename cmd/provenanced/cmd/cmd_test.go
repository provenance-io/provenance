package cmd_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"

	"github.com/provenance-io/provenance/cmd/provenanced/cmd"
)

func TestInitCmd(t *testing.T) {
	rootCmd, _ := cmd.NewRootCmd(true)
	rootCmd.SetArgs([]string{
		"init",        // Test the init cmd
		"simapp-test", // Moniker
		fmt.Sprintf("--%s=%s", cli.FlagOverwrite, "true"), // Overwrite genesis.json, in case it already exists
	})

	err := cmd.Execute(rootCmd)
	require.NoError(t, err)
}
