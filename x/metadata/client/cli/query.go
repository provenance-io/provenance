package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/gogo/protobuf/proto"

	"github.com/google/uuid"

	"github.com/provenance-io/provenance/x/metadata/types"

	"gopkg.in/yaml.v2"
)

var cmdStart = fmt.Sprintf("%s query metadata", version.AppName)

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
		GetMetadataFullScopeCmd(),
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
		Use:     "params",
		Short:   "Query the current metadata parameters",
		Args:    cobra.NoArgs,
		Example: fmt.Sprintf("%s params", cmdStart),
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
		Use:   "get id",
		Short: "Query the current metadata by id",
		Long: fmt.Sprintf(`%[1]s get {scope_id} - gets the scope with the given scope id.
%[1]s get {session_id} - gets the session with the given session id.
%[1]s get {record_id} - gets the record with the given record id.
%[1]s get {scope_spec_id} - gets the scope specification with the given scope specification id.
%[1]s get {contract_spec_id} - gets the contract specification with the given contract specification id.
%[1]s get {record_spec_id} - gets the record specification with the given record specification id.`, cmdStart),
		Args: cobra.ExactArgs(1),
		Example: fmt.Sprintf(`%[1]s get scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel
%[1]s get session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr
%[1]s get record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3
%[1]s get scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m
%[1]s get contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn
%[1]s get recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44`, cmdStart),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := types.MetadataAddressFromBech32(strings.TrimSpace(args[0]))
			if err != nil {
				return err
			}
			prefix, err := id.Prefix()
			if err != nil {
				return err
			}

			switch prefix {
			case types.PrefixScope:
				return fullScopeByID(cmd, id)
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
			return fmt.Errorf("unexpected address prefix %s", prefix)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetMetadataScopeCmd returns the command handler for metadata scope querying.
func GetMetadataScopeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scope {scope_id|scope_uuid}",
		Short: "Query the current metadata for a scope",
		Long: fmt.Sprintf(`%[1]s scope {scope_id} - gets the scope with the given id.
%[1]s scope {scope_uuid} - gets the scope with the given uuid.`, cmdStart),
		Args: cobra.ExactArgs(1),
		Example: fmt.Sprintf(`%[1]s scope 91978ba2-5f35-459a-86a7-feca1b0512e0
%[1]s scope scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel`, cmdStart),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg0 := strings.TrimSpace(args[0])
			id, idErr := types.MetadataAddressFromBech32(arg0)
			if idErr == nil {
				return scopeByID(cmd, id)
			}
			_, uuidErr := uuid.Parse(arg0)
			if uuidErr == nil {
				return scopeByUUID(cmd, arg0)
			}
			return fmt.Errorf("argument %s is neither a metadata address (%s) nor uuid (%s)", arg0, idErr.Error(), uuidErr.Error())
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetMetadataFullScopeCmd returns the command handler for metadata full-scope querying.
func GetMetadataFullScopeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fullscope {scope_id|scope_uuid}",
		Short: "Query the current metadata for a scope, its sessions, and its records",
		Long: fmt.Sprintf(`%[1]s fullscope {scope_id} - gets a scope, sessions, and records with the given scope id.
%[1]s fullscope {scope_uuid} - gets a scope, sessions, and records with the given scope uuid.`, cmdStart),
		Args: cobra.ExactArgs(1),
		Example: fmt.Sprintf(`%[1]s fullscope 91978ba2-5f35-459a-86a7-feca1b0512e0
%[1]s fullscope scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel`, cmdStart),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg0 := strings.TrimSpace(args[0])
			id, idErr := types.MetadataAddressFromBech32(arg0)
			if idErr == nil {
				return fullScopeByID(cmd, id)
			}
			_, uuidErr := uuid.Parse(arg0)
			if uuidErr == nil {
				return fullScopeByUUID(cmd, arg0)
			}
			return fmt.Errorf("argument %s is neither a metadata address (%s) nor uuid (%s)", arg0, idErr.Error(), uuidErr.Error())
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetMetadataSessionCmd returns the command handler for metadata session querying.
func GetMetadataSessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "session {session_id|scope_id|scope_uuid [session_uuid]}",
		Aliases: []string{"sessions"},
		Short:   "Query the current metadata for sessions",
		Long: fmt.Sprintf(`%[1]s session {session_id} - gets the session with the given id.
%[1]s session {scope_id} - gets the list of sessions associated with a scope.
%[1]s session {scope_uuid} - gets the list of sessions associated with a scope.
%[1]s session {scope_uuid} {session_uuid} - gets a session with the given scope uuid and session uuid.`, cmdStart),
		Args: cobra.RangeArgs(1, 2),
		Example: fmt.Sprintf(`%[1]s session 91978ba2-5f35-459a-86a7-feca1b0512e0 5803f8bc-6067-4eb5-951f-2121671c2ec0
%[1]s session session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr`, cmdStart),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg0 := strings.TrimSpace(args[0])
			if len(args) == 1 {
				id, idErr := types.MetadataAddressFromBech32(arg0)
				if idErr == nil {
					switch {
					case id.IsSessionAddress():
						return sessionByID(cmd, id)
					case id.IsScopeAddress():
						return sessionsByScopeID(cmd, id)
					}
					return fmt.Errorf("unexpected metadata address prefix on %s", id)
				}
				_, uuidErr := uuid.Parse(arg0)
				if uuidErr == nil {
					return sessionsByScopeUUID(cmd, arg0)
				}
				return fmt.Errorf("argument %s is neither a metadata address (%s) nor uuid (%s)", arg0, idErr.Error(), uuidErr.Error())
			} // else there's 2 more arguments.
			_, uuidErr := uuid.Parse(arg0)
			if uuidErr != nil {
				return uuidErr
			}
			arg1 := strings.TrimSpace(args[1])
			_, uuidErr = uuid.Parse(arg1)
			if uuidErr != nil {
				return uuidErr
			}
			return sessionByUUIDs(cmd, arg0, arg1)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetMetadataRecordCmd returns the command handler for metadata record querying.
func GetMetadataRecordCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "record {record_id|session_id|scope_id|scope_uuid [record_name]}",
		Aliases: []string{"r", "records"},
		Short:   "Query the current metadata for records",
		Long: fmt.Sprintf(`%[1]s record {record_id} - gets the record with the given id.
%[1]s record {session_id} - gets the list of records associated with a session id.
%[1]s record {scope_id} - gets the list of records associated with a scope id.
%[1]s record {scope_uuid} - gets the list of records associated with a scope uuid.
%[1]s record {scope_uuid} {record_name} - gets the record with the given name from the given scope.`, cmdStart),
		Args: cobra.MinimumNArgs(1),
		Example: fmt.Sprintf(`%[1]s record 91978ba2-5f35-459a-86a7-feca1b0512e0 recordname
%[1]s record record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3`, cmdStart),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg0 := strings.TrimSpace(args[0])
			if len(args) == 1 {
				id, idErr := types.MetadataAddressFromBech32(arg0)
				if idErr == nil {
					switch {
					case id.IsRecordAddress():
						return recordByID(cmd, id)
					case id.IsSessionAddress():
						return recordsBySessionID(cmd, id)
					case id.IsScopeAddress():
						return recordsByScopeID(cmd, id)
					}
					return fmt.Errorf("unexpected metadata address prefix on %s", id)
				}
				_, uuidErr := uuid.Parse(arg0)
				if uuidErr == nil {
					return recordsByScopeUUID(cmd, arg0)
				}
				return fmt.Errorf("argument %s is neither a metadata address (%s) nor uuid (%s)", arg0, idErr.Error(), uuidErr.Error())
			} // else there's 2 more arguments.
			_, uuidErr := uuid.Parse(arg0)
			if uuidErr != nil {
				return uuidErr
			}
			arg1 := trimSpaceAndJoin(args[1:], " ")
			if len(arg1) == 0 {
				return recordsByScopeUUID(cmd, arg0)
			}
			return recordByScopeUUIDAndName(cmd, arg0, arg1)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetMetadataScopeSpecCmd returns the command handler for metadata scope specification querying.
func GetMetadataScopeSpecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "scopespec {scope_spec_id|scope_spec_uuid}",
		Aliases: []string{"ss", "scopespecification"},
		Short:   "Query the current metadata for a scope specification",
		Long: fmt.Sprintf(`%[1]s scopespec {scope_spec_id} - gets the scope specification for that a given id.
%[1]s scopespec {scope_spec_uuid} - gets the scope specification for a given uuid.`, cmdStart),
		Args: cobra.ExactArgs(1),
		Example: fmt.Sprintf(`%[1]s scopespec dc83ea70-eacd-40fe-9adf-1cf6148bf8a2
%[1]s scopespec scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m`, cmdStart),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg0 := strings.TrimSpace(args[0])
			id, idErr := types.MetadataAddressFromBech32(arg0)
			if idErr == nil {
				return scopeSpecByID(cmd, id)
			}
			_, uuidErr := uuid.Parse(arg0)
			if uuidErr == nil {
				return scopeSpecByUUID(cmd, arg0)
			}
			return fmt.Errorf("argument %s is neither a metadata address (%s) nor uuid (%s)", arg0, idErr.Error(), uuidErr.Error())
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetMetadataContractSpecCmd returns the command handler for metadata contract specification querying.
func GetMetadataContractSpecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "contractspec {contract_spec_id|contract_spec_uuid|record_spec_id}",
		Aliases: []string{"cs", "contractspecification"},
		Short:   "Query the current metadata for a contract specification",
		Long: fmt.Sprintf(`%[1]s contractspec {contract_spec_id} - gets the contract specification for that a given id.
%[1]s contractspec {contract_spec_uuid} - gets the contract specification for a given uuid.
%[1]s contractspec {record_spec_id} - gets the contract specification associated with that record spec id.`, cmdStart),
		Args: cobra.ExactArgs(1),
		Example: fmt.Sprintf(`%[1]s contractspec def6bc0a-c9dd-4874-948f-5206e6060a84
%[1]s contractspec contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn`, cmdStart),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg0 := strings.TrimSpace(args[0])
			id, idErr := types.MetadataAddressFromBech32(arg0)
			if idErr == nil {
				switch {
				case id.IsContractSpecificationAddress():
					return contractSpecByID(cmd, id)
				case id.IsRecordSpecificationAddress():
					return contractSpecByID(cmd, id.GetContractSpecAddress())
				}
				return fmt.Errorf("unexpected metadata address prefix on %s", id)
			}
			_, uuidErr := uuid.Parse(arg0)
			if uuidErr == nil {
				return contractSpecByUUID(cmd, arg0)
			}
			return fmt.Errorf("argument %s is neither a metadata address (%s) nor uuid (%s)", arg0, idErr.Error(), uuidErr.Error())
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetMetadataRecordSpecCmd returns the command handler for metadata record specification querying.
func GetMetadataRecordSpecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "recordspec {rec_spec_id|contract_spec_id|contract_spec_uuid [record_name]}",
		Aliases: []string{"rs", "recordspecs", "recspec", "recspecs", "recordspecification", "recordspecifications"},
		Short:   "Query the current metadata for record specifications",
		Long: fmt.Sprintf(`%[1]s recordspec {rec_spec_id} - gets the record specification for a given id.
%[1]s recordspec {contract_spec_id} - gets the list of record specifications for the given contract specification.
%[1]s recordspec {contract_spec_uuid} - gets the list of record specifications for the given contract specification.
%[1]s recordspec {contract_spec_uuid} {record_name} - gets the record specification for a given contract specification uuid and record name.`, cmdStart),
		Args: cobra.MinimumNArgs(1),
		Example: fmt.Sprintf(`%[1]s recordspec def6bc0a-c9dd-4874-948f-5206e6060a84 recordname
%[1]s recordspec recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44`, cmdStart),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg0 := strings.TrimSpace(args[0])
			if len(args) == 1 {
				id, idErr := types.MetadataAddressFromBech32(arg0)
				if idErr == nil {
					switch {
					case id.IsRecordSpecificationAddress():
						return recordSpecByID(cmd, id)
					case id.IsContractSpecificationAddress():
						return recordSpecsByContractSpecID(cmd, id)
					}
					return fmt.Errorf("unexpected metadata address prefix on %s", id)
				}
				_, uuidErr := uuid.Parse(arg0)
				if uuidErr == nil {
					return recordSpecsByContractSpecUUID(cmd, arg0)
				}
				return fmt.Errorf("argument %s is neither a metadata address (%s) nor uuid (%s)", arg0, idErr.Error(), uuidErr.Error())
			} // else there's 2 more arguments.
			_, uuidErr := uuid.Parse(arg0)
			if uuidErr != nil {
				return uuidErr
			}
			arg1 := trimSpaceAndJoin(args[1:], " ")
			if len(arg1) == 0 {
				return recordSpecsByContractSpecUUID(cmd, arg0)
			}
			return recordSpecByContractSpecUUIDAndName(cmd, arg0, arg1)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// scopeByID outputs a scope looked up by a scope MetadataAddress.
func scopeByID(cmd *cobra.Command, scopeID types.MetadataAddress) error {
	if !scopeID.IsScopeAddress() {
		return fmt.Errorf("id %s is not a scope metadata address", scopeID)
	}
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
	if res == nil || res.Scope == nil {
		return fmt.Errorf("no scope found with uuid %s", scopeUUID)
	}

	return clientCtx.PrintProto(res.Scope)
}

// sessionsByScopeID outputs the sessions for a scope looked up by a scope MetadataAddress.
func sessionsByScopeID(cmd *cobra.Command, scopeID types.MetadataAddress) error {
	if !scopeID.IsScopeAddress() {
		return fmt.Errorf("id %s is not a scope metadata address", scopeID)
	}
	scopeUUID, err := scopeID.ScopeUUID()
	if err != nil {
		return err
	}
	return sessionsByScopeUUID(cmd, scopeUUID.String())
}

// sessionsByScopeUUID outputs the sessions for a scope looked up by scope UUID.
func sessionsByScopeUUID(cmd *cobra.Command, scopeUUID string) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	queryClient := types.NewQueryClient(clientCtx)
	res, err := queryClient.Scope(context.Background(), &types.ScopeRequest{ScopeUuid: scopeUUID})
	if err != nil {
		return err
	}
	if res == nil || res.Sessions == nil || len(res.Sessions) == 0 {
		return fmt.Errorf("no sessions found for scope uuid %s", scopeUUID)
	}

	protoList := make([]proto.Message, len(res.Sessions))
	for i, session := range res.Sessions {
		if session != nil {
			protoList[i] = session
		} else {
			protoList[i] = &types.Session{}
		}
	}
	return printProtoList(clientCtx, protoList)
}

// recordsByScopeID outputs the records for a scope looked up by a scope MetadataAddress.
func recordsByScopeID(cmd *cobra.Command, scopeID types.MetadataAddress) error {
	if !scopeID.IsScopeAddress() {
		return fmt.Errorf("id %s is not a scope metadata address", scopeID)
	}
	scopeUUID, err := scopeID.ScopeUUID()
	if err != nil {
		return err
	}
	return recordsByScopeUUID(cmd, scopeUUID.String())
}

// recordsByScopeUUID outputs the records for a scope looked up by scope UUID.
func recordsByScopeUUID(cmd *cobra.Command, scopeUUID string) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	queryClient := types.NewQueryClient(clientCtx)
	res, err := queryClient.Scope(context.Background(), &types.ScopeRequest{ScopeUuid: scopeUUID})
	if err != nil {
		return err
	}
	if res == nil || res.Records == nil || len(res.Records) == 0 {
		return fmt.Errorf("no records found for scope uuid %s", scopeUUID)
	}

	protoList := make([]proto.Message, len(res.Records))
	for i, record := range res.Records {
		if record != nil {
			protoList[i] = record
		} else {
			protoList[i] = &types.Record{}
		}
	}
	return printProtoList(clientCtx, protoList)
}

func recordsBySessionID(cmd *cobra.Command, sessionID types.MetadataAddress) error {
	scopeUUID, err := sessionID.ScopeUUID()
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
	if res == nil || res.Records == nil || len(res.Records) == 0 {
		return fmt.Errorf("no records found for scope uuid %s", scopeUUID)
	}

	protoList := []proto.Message{}
	for _, record := range res.Records {
		if record != nil && sessionID.Equals(record.SessionId) {
			protoList = append(protoList, record)
		}
	}
	return printProtoList(clientCtx, protoList)
}

// fullScopeByID outputs a scope, its sessions, and its records, looked up by a scope MetadataAddress.
func fullScopeByID(cmd *cobra.Command, scopeID types.MetadataAddress) error {
	if !scopeID.IsScopeAddress() {
		return fmt.Errorf("id %s is not a scope metadata address", scopeID)
	}
	scopeUUID, err := scopeID.ScopeUUID()
	if err != nil {
		return err
	}
	return fullScopeByUUID(cmd, scopeUUID.String())
}

// fullScopeByUUID outputs a scope, its sessions, and its records, looked up by scope UUID.
func fullScopeByUUID(cmd *cobra.Command, scopeUUID string) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	queryClient := types.NewQueryClient(clientCtx)
	res, err := queryClient.Scope(context.Background(), &types.ScopeRequest{ScopeUuid: scopeUUID})
	if err != nil {
		return err
	}
	if res == nil {
		return fmt.Errorf("no scope found with uuid %s", scopeUUID)
	}

	return clientCtx.PrintProto(res)
}

// sessionByID outputs a session looked up by a session MetadataAddress.
func sessionByID(cmd *cobra.Command, sessionID types.MetadataAddress) error {
	if !sessionID.IsSessionAddress() {
		return fmt.Errorf("id %s is not a session metadata address", sessionID)
	}
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
	if !recordID.IsRecordAddress() {
		return fmt.Errorf("id %s is not a record metadata address", recordID)
	}
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
	if !scopeSpecID.IsScopeSpecificationAddress() {
		return fmt.Errorf("id %s is not a scope specification metadata address", scopeSpecID)
	}
	scopeSpecUUID, err := scopeSpecID.ScopeSpecUUID()
	if err != nil {
		return err
	}
	return scopeSpecByUUID(cmd, scopeSpecUUID.String())
}

// scopeSpecByUUID outputs a scope specification looked up by scope specification UUID.
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
	if !contractSpecID.IsContractSpecificationAddress() {
		return fmt.Errorf("id %s is not a contract specification metadata address", contractSpecID)
	}
	contractSpecUUID, err := contractSpecID.ContractSpecUUID()
	if err != nil {
		return err
	}
	return contractSpecByUUID(cmd, contractSpecUUID.String())
}

// contractSpecByUUID outputs a contract specification looked up by contract specification UUID.
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
	if !recordSpecID.IsRecordSpecificationAddress() {
		return fmt.Errorf("id %s is not a record specification metadata address", recordSpecID)
	}
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

// recordSpecsByContractSpecID outputs the record specifications looked up by a contract specification MetadataAddress.
func recordSpecsByContractSpecID(cmd *cobra.Command, contractSpecID types.MetadataAddress) error {
	if !contractSpecID.IsContractSpecificationAddress() {
		return fmt.Errorf("id %s is not a contract specification metadata address", contractSpecID)
	}
	contractSpecUUID, err := contractSpecID.ContractSpecUUID()
	if err != nil {
		return err
	}
	return recordSpecsByContractSpecUUID(cmd, contractSpecUUID.String())
}

// recordSpecsByContractSpecUUID outputs the record specifications looked up by contract specification UUID.
func recordSpecsByContractSpecUUID(cmd *cobra.Command, contractSpecUUID string) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}
	queryClient := types.NewQueryClient(clientCtx)
	res, err := queryClient.RecordSpecificationsForContractSpecification(
		context.Background(),
		&types.RecordSpecificationsForContractSpecificationRequest{ContractSpecificationUuid: contractSpecUUID},
	)
	if err != nil {
		return err
	}
	if res == nil || res.RecordSpecifications == nil || len(res.RecordSpecifications) == 0 {
		return fmt.Errorf("no contract specification found with uuid %s", contractSpecUUID)
	}

	protoList := make([]proto.Message, len(res.RecordSpecifications))
	for i, recSpec := range res.RecordSpecifications {
		if recSpec != nil {
			protoList[i] = recSpec
		} else {
			protoList[i] = &types.RecordSpecification{}
		}
	}
	return printProtoList(clientCtx, protoList)
}

func trimSpaceAndJoin(args []string, sep string) string {
	trimmedArgs := make([]string, len(args))
	for i, arg := range args {
		trimmedArgs[i] = strings.TrimSpace(arg)
	}
	return strings.TrimSpace(strings.Join(trimmedArgs, sep))
}

// printProtoList outputs toPrint to the ctx.Output based on ctx.OutputFormat which is
// either text or json. If text, toPrint will be YAML encoded. Otherwise, toPrint
// will be JSON encoded using ctx.JSONMarshaler. An error is returned upon failure.
// See also: client.Context.PrintProto
func printProtoList(ctx client.Context, toPrint []proto.Message) error {
	// always serialize JSON initially because proto json can't be directly YAML encoded
	result := []byte{}
	result = append(result, []byte("[")...)
	maxI := len(toPrint) - 1
	for i, p := range toPrint {
		out, err := ctx.JSONMarshaler.MarshalJSON(p)
		if err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
		result = append(result, out...)
		if i < maxI {
			result = append(result, []byte(",")...)
		}
	}
	result = append(result, []byte("]")...)

	if ctx.OutputFormat == "text" {
		// handle text format by decoding and re-encoding JSON as YAML
		var j interface{}

		err := json.Unmarshal(result, &j)
		if err != nil {
			return err
		}

		result, err = yaml.Marshal(j)
		if err != nil {
			return err
		}
	}

	writer := ctx.Output
	if writer == nil {
		writer = os.Stdout
	}

	_, err := writer.Write(result)
	if err != nil {
		return err
	}

	if ctx.OutputFormat != "text" {
		// append new-line for formats besides YAML
		_, err = writer.Write([]byte("\n"))
		if err != nil {
			return err
		}
	}

	return nil
}
