package cli

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/provenance-io/provenance/x/registry"
	"github.com/spf13/cobra"
)

// CmdQuery returns the cli query commands for the registry module
func CmdQuery() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        registry.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", registry.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdQueryRegistryEntry(),
		GetCmdQueryRegistryEntries(),
	)

	return cmd
}

// GetCmdQueryRegistryEntry returns the command for querying a registry entry
func GetCmdQueryRegistryEntry() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "entry [address]",
		Short: "Query a registry entry by address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := registry.NewQueryClient(clientCtx)

			res, err := queryClient.GetRegistryEntry(context.Background(), &registry.QueryGetRegistryEntryRequest{
				Address: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryRegistryEntries returns the command for querying all registry entries
func GetCmdQueryRegistryEntries() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "entries",
		Short: "Query all registry entries",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := registry.NewQueryClient(clientCtx)

			res, err := queryClient.ListRegistryEntries(context.Background(), &registry.QueryListRegistryEntriesRequest{
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "registry entries")

	return cmd
}
