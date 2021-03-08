package cli

import (
	"context"
	"fmt"
	"github.com/google/uuid"
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
				return scopeByID(cmd, id)
			case types.PrefixSession:
				return sessionByID(cmd, id)
			case types.PrefixRecord:
				return recordByID(cmd, id)
			case types.PrefixScopeSpecification:
				return scopeSpecByID(cmd, id)
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

// scopeByID outputs a scope looked up by a scope MetadataAddress.
func scopeByID(cmd *cobra.Command, scopeID types.MetadataAddress) error {
	scopeUUID, err := scopeID.ScopeUUID()
	if err != nil {
		return err
	}
	return scopeByUUID(cmd, scopeUUID.String())
}

// scopeByUUID outputs a scope looked up by scope UUID.
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
	if res.Scope == nil {
		return fmt.Errorf("no scope found with uuid %s", scopeUUID)
	}

	return clientCtx.PrintProto(res.Scope)
}

// sessionByID outputs a session looked up by a session MetadataAddress.
func sessionByID(cmd *cobra.Command, sessionID types.MetadataAddress) error {
	scopeUUID, err := sessionID.ScopeUUID()
	if err != nil {
		return err
	}
	sessionUUID, err := sessionID.SessionUUID()
	if err != nil {
		return err
	}
	return sessionByUUIDs(cmd, scopeUUID.String(), sessionUUID.String())
}

// sessionByUUIDs outputs a session looked up by scope UUID and session UUID.
func sessionByUUIDs(cmd *cobra.Command, scopeUUID string, sessionUUID string) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	queryClient := types.NewQueryClient(clientCtx)
	res, err := queryClient.SessionContextByUUID(context.Background(), &types.SessionContextByUUIDRequest{ScopeUuid: scopeUUID, SessionUuid: sessionUUID})
	if err != nil {
		return err
	}
	if res.Sessions == nil || len(res.Sessions) == 0 {
		return fmt.Errorf("no session found with scope uuid %s and session uuid %s", scopeUUID, sessionUUID)
	}
	if len(res.Sessions) == 1 {
		return clientCtx.PrintProto(res.Sessions[0])
	}
	return clientCtx.PrintProto(res)
}

// recordByID outputs a record looked up by a record MetadataAddress.
func recordByID(cmd *cobra.Command, recordID types.MetadataAddress) error {
	scopeUUID, err := recordID.ScopeUUID()
	if err != nil {
		return err
	}
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}
	queryClient := types.NewQueryClient(clientCtx)
	res, err := queryClient.Scope(context.Background(), &types.ScopeRequest{ScopeUuid: scopeUUID.String()})
	if err != nil {
		return err
	}
	var record *types.Record = nil
	for _, r := range res.Records {
		if recordID.Equals(types.RecordMetadataAddress(scopeUUID, r.Name)) {
			record = r
			break
		}
	}
	if record == nil {
		return fmt.Errorf("no records with id %s found in scope with uuid %s", recordID, scopeUUID)
	}
	return clientCtx.PrintProto(record)
}

// recordByScopeUUIDAndName outputs a record looked up by scope UUID ane record name.
func recordByScopeUUIDAndName(cmd *cobra.Command, scopeUUID string, name string) error {
	primaryUUID, err := uuid.Parse(scopeUUID)
	if err != nil {
		return err
	}
	return recordByID(cmd, types.RecordMetadataAddress(primaryUUID, name))
}

// scopeSpecByID outputs a scope specification looked up by a scope specification MetadataAddress.
func scopeSpecByID(cmd *cobra.Command, scopeSpecID types.MetadataAddress) error {
	scopeSpecUUID, err := scopeSpecID.PrimaryUUID()
	if err != nil {
		return err
	}
	return scopeSpecByUUID(cmd, scopeSpecUUID.String())
}

// scopeSpecByUUID outputs a scope specification looked up by specification UUID.
func scopeSpecByUUID(cmd *cobra.Command, scopeSpecUUID string) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}
	queryClient := types.NewQueryClient(clientCtx)
	res, err := queryClient.ScopeSpecification(context.Background(), &types.ScopeSpecificationRequest{SpecificationUuid: scopeSpecUUID})
	if err != nil {
		return err
	}
	if res.ScopeSpecification == nil {
		return fmt.Errorf("no scope specification found with uuid %s", scopeSpecUUID)
	}
	return clientCtx.PrintProto(res.ScopeSpecification)
}