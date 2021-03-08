package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/google/uuid"

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
		GetMetadataByIDCmd(),
		GetMetadataScopeCmd(),
		GetMetadataSessionCmd(),
		GetMetadataRecordCmd(),
		GetMetadataScopeSpecCmd(),
		GetMetadataContractSpecCmd(),
		GetMetadataRecordSpecCmd(),
	)
	return queryCmd
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

// GetMetadataByIDCmd returns the command handler for querying metadata for anything from an id.
func GetMetadataByIDCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get id",
		Short:   "Query the current metadata by id",
		Args:    cobra.ExactArgs(1),
		Example: fmt.Sprintf("%s query metadata get scope1qzxcpvj6czy5g354dews3nlruxjsahhnsp", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := types.MetadataAddressFromBech32(strings.TrimSpace(args[0]))
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
				return contractSpecByID(cmd, id)
			case types.PrefixRecordSpecification:
				return recordSpecByID(cmd, id)
			}
			return fmt.Errorf("unexpected address type %s", addressType)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetMetadataScopeCmd returns the command handler for metadata scope querying.
func GetMetadataScopeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "scope uuid",
		Short:   "Query the current metadata for a scope",
		Args:    cobra.ExactArgs(1),
		Example: fmt.Sprintf("%s query metadata scope 123e4567-e89b-12d3-a456-426614174000", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			return scopeByUUID(cmd, strings.TrimSpace(args[0]))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetMetadataSessionCmd returns the command handler for metadata session querying.
func GetMetadataSessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "session scope_uuid session_uuid",
		Short:   "Query the current metadata for a session",
		Args:    cobra.ExactArgs(2),
		Example: fmt.Sprintf("%s query metadata session 123e4567-e89b-12d3-a456-426614174000 123e4567-e89b-12d3-a456-426614174001", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			return sessionByUUIDs(cmd, strings.TrimSpace(args[0]), strings.TrimSpace(args[1]))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetMetadataRecordCmd returns the command handler for metadata record querying.
func GetMetadataRecordCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "record scope_uuid record_name",
		Aliases: []string{"r"},
		Short:   "Query the current metadata for a record",
		Args:    cobra.ExactArgs(2),
		Example: fmt.Sprintf("%s query metadata record 123e4567-e89b-12d3-a456-426614174000 myrecord", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			return recordByScopeUUIDAndName(cmd, strings.TrimSpace(args[0]), strings.TrimSpace(args[1]))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetMetadataScopeSpecCmd returns the command handler for metadata scope specification querying.
func GetMetadataScopeSpecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "scopespec uuid",
		Aliases: []string{"ss"},
		Short:   "Query the current metadata for a scope specification",
		Args:    cobra.ExactArgs(1),
		Example: fmt.Sprintf("%s query metadata scopespec 123e4567-e89b-12d3-a456-426614174000", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			return scopeSpecByUUID(cmd, strings.TrimSpace(args[0]))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetMetadataContractSpecCmd returns the command handler for metadata contract specification querying.
func GetMetadataContractSpecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "contractspec uuid",
		Aliases: []string{"cs"},
		Short:   "Query the current metadata for a contract specification",
		Args:    cobra.ExactArgs(1),
		Example: fmt.Sprintf("%s query metadata contractspec 123e4567-e89b-12d3-a456-426614174000", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			return contractSpecByUUID(cmd, strings.TrimSpace(args[0]))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetMetadataRecordSpecCmd returns the command handler for metadata record specification querying.
func GetMetadataRecordSpecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "recordspec contract_spec_uuid record_name",
		Aliases: []string{"rs"},
		Short:   "Query the current metadata for a record specification",
		Args:    cobra.ExactArgs(1),
		Example: fmt.Sprintf("%s query metadata recordspec 123e4567-e89b-12d3-a456-426614174000 myrecord", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			return recordSpecByContractSpecUUIDAndName(cmd, strings.TrimSpace(args[0]), strings.TrimSpace(args[1]))
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
	res, err := queryClient.SessionContextByUUID(
		context.Background(),
		&types.SessionContextByUUIDRequest{
			ScopeUuid:   scopeUUID,
			SessionUuid: sessionUUID,
		},
	)
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
	scopeSpecUUID, err := scopeSpecID.ScopeSpecUUID()
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

// contractSpecByID outputs a contract specification looked up by a contract specification MetadataAddress.
func contractSpecByID(cmd *cobra.Command, contractSpecID types.MetadataAddress) error {
	contractSpecUUID, err := contractSpecID.ContractSpecUUID()
	if err != nil {
		return err
	}
	return contractSpecByUUID(cmd, contractSpecUUID.String())
}

// contractSpecByUUID outputs a contract specification looked up by specification UUID.
func contractSpecByUUID(cmd *cobra.Command, contractSpecUUID string) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}
	queryClient := types.NewQueryClient(clientCtx)
	res, err := queryClient.ContractSpecification(context.Background(), &types.ContractSpecificationRequest{SpecificationUuid: contractSpecUUID})
	if err != nil {
		return err
	}
	if res.ContractSpecification == nil {
		return fmt.Errorf("no contract specification found with uuid %s", contractSpecUUID)
	}
	return clientCtx.PrintProto(res.ContractSpecification)
}

// recordSpecByID outputs a record specification looked up by a record specification MetadataAddress.
func recordSpecByID(cmd *cobra.Command, recordSpecID types.MetadataAddress) error {
	contractSpecUUID, err := recordSpecID.ContractSpecUUID()
	if err != nil {
		return err
	}
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}
	queryClient := types.NewQueryClient(clientCtx)
	res, err := queryClient.RecordSpecificationsForContractSpecification(
		context.Background(),
		&types.RecordSpecificationsForContractSpecificationRequest{ContractSpecificationUuid: contractSpecUUID.String()},
	)
	if err != nil {
		return err
	}
	var recSpec *types.RecordSpecification = nil
	for _, rs := range res.RecordSpecifications {
		if recordSpecID.Equals(rs.SpecificationId) {
			recSpec = rs
			break
		}
	}
	if recSpec == nil {
		return fmt.Errorf("no record specification found with id %s in contract specification uuid %s", recordSpecID, contractSpecUUID)
	}
	return clientCtx.PrintProto(recSpec)
}

// recordSpecByContractSpecUUIDAndName outputs a record specification looked up by contract specification UUID and record specification name.
func recordSpecByContractSpecUUIDAndName(cmd *cobra.Command, contractSpecUUID string, name string) error {
	primaryUUID, err := uuid.Parse(contractSpecUUID)
	if err != nil {
		return err
	}
	return recordSpecByID(cmd, types.RecordSpecMetadataAddress(primaryUUID, name))
}
