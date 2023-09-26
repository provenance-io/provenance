package cli

import (
	"github.com/spf13/cobra"

	"github.com/provenance-io/provenance/x/ibcratelimit/osmoutils/osmocli"
	"github.com/provenance-io/provenance/x/ibcratelimit/types"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	cmd := osmocli.QueryIndexCmd(types.ModuleName)

	cmd.AddCommand(
		osmocli.GetParams[*types.ParamsRequest](
			types.ModuleName, types.NewQueryClient),
	)

	return cmd
}
