package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/provenance-io/provenance/x/metadata/types"
)

// GetQueryCmd returns the top-level command for marker CLI queries.
func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the metadata module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	queryCmd.AddCommand(
		GetMetadataParamsCmd(),
		GetMetadataScopeCmd(),
	)
	return queryCmd
}

// GetMetadataEntryCmd returns the command handler for querying metadata for anything from an id.
func GetMetadataEntryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [id]",
		Short: "Query the current metadata by id",
		Args:  cobra.ExactArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the current metadata module by id:
Example:
$ %s query metadata get scope1qzxcpvj6czy5g354dews3nlruxjsahhnsp

Any metadata address type is allowed and the appropriate entry will be returned.
`,
				version.AppName,
			)),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := types.MetadataAddressFromBech32(strings.ToLower(strings.TrimSpace(args[0])))
			if err != nil {
				return err
			}
			addressType, err := id.Prefix()
			if err != nil {
				return err
			}

			switch addressType {
			case types.PrefixScope:
				scopeUUID, err := id.ScopeUUID()
				if err != nil {
					return err
				}
				return scopeByUUID(cmd, scopeUUID.String())
			case types.PrefixSession:
				// TODO: Session lookup
			case types.PrefixRecord:
				// TODO: Record lookup
			case types.PrefixScopeSpecification:
				// TODO: Scope spec lookup
			case types.PrefixContractSpecification:
				// TODO: Contract spec lookup
			case types.PrefixRecordSpecification:
				// TODO: Record spec lookup
			}
			return fmt.Errorf("unexpected address type %s", addressType)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetMetadataParamsCmd returns the command handler for metadata parameter querying.
func GetMetadataParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the current metadata parameters",
		Args:  cobra.NoArgs,
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the current metadata module parameters:
Example:
$ %s query metadata params
`,
				version.AppName,
			)),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetMetadataScopeCmd returns the command handler for metadata scope querying.
func GetMetadataScopeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scope [id]",
		Short: "Query the current metadata for scope",
		Args:  cobra.ExactArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the current metadata module parameters:
Example:
$ %s query metadata scope 123e4567-e89b-12d3-a456-426614174000
`,
				version.AppName,
			)),
		RunE: func(cmd *cobra.Command, args []string) error {
			return scopeByUUID(cmd, strings.ToLower(strings.TrimSpace(args[0])))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func scopeByUUID(cmd *cobra.Command, scopeUUID string) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	queryClient := types.NewQueryClient(clientCtx)
	res, err := queryClient.Scope(context.Background(), &types.ScopeRequest{ScopeUuid: scopeUUID})
	if err != nil {
		return err
	}

	return clientCtx.PrintProto(res.Scope)
}
