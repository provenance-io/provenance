package cli

import (
	"github.com/spf13/cobra"

	"github.com/provenance-io/provenance/x/asset/types"
)

// GetRootCmd returns the root command for the asset module
func GetRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Asset module commands",
	}

	cmd.AddCommand(
		GetTxCmd(),
		GetQueryCmd(),
	)

	return cmd
}
