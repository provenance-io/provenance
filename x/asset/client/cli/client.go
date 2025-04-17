package cli

import (
	"github.com/spf13/cobra"
)

// RegisterCommands registers the asset module commands with the provided root command
func RegisterCommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(GetRootCmd())
}
