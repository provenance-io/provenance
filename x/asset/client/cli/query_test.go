package cli_test

import (
	"strings"

	"github.com/provenance-io/provenance/x/asset/client/cli"
	"github.com/provenance-io/provenance/x/asset/types"
)

// TestGetQueryCmd tests the GetQueryCmd function.
func (s *CmdTestSuite) TestGetQueryCmd() {
	cmd := cli.GetQueryCmd()
	s.Require().NotNil(cmd, "GetQueryCmd should not return nil")
	s.Require().Equal(types.ModuleName, cmd.Use, "command use should be module name")
	s.Require().Equal("Querying commands for the asset module", cmd.Short, "command short description")
	s.Require().True(cmd.DisableFlagParsing, "flag parsing should be disabled")
	s.Require().Equal(2, cmd.SuggestionsMinimumDistance, "suggestions minimum distance")

	// Check that all expected subcommands are present
	expectedSubCommands := []string{
		"asset",
		"assets",
		"class",
		"classes",
	}

	subCommands := cmd.Commands()
	s.Require().GreaterOrEqual(len(subCommands), len(expectedSubCommands), "should have all expected subcommands")

	// Extract the first word from each command's Use string
	subCommandNames := make(map[string]bool)
	for _, subCmd := range subCommands {
		// The Use string may contain arguments like "asset <class-id> <asset-id>"
		// Extract just the command name (first word)
		parts := strings.Fields(subCmd.Use)
		if len(parts) > 0 {
			subCommandNames[parts[0]] = true
		}
	}

	for _, expectedCmd := range expectedSubCommands {
		s.Assert().True(subCommandNames[expectedCmd], "expected subcommand %s to exist", expectedCmd)
	}
}
