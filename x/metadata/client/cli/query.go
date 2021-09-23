package cli

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/google/uuid"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/provenance-io/provenance/x/metadata/types"
)

var (
	cmdStart = fmt.Sprintf("%s query metadata", version.AppName)

	includeScope       bool
	includeSessions    bool
	includeRecords     bool
	includeRecordSpecs bool
	includeRequest     bool
)

const all = "all"

// GetQueryCmd returns the top-level command for marker CLI queries.
func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"md"},
		Short:                      "Querying commands for the metadata module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	queryCmd.AddCommand(
		GetMetadataParamsCmd(),
		GetMetadataByIDCmd(),
		GetMetadataGetAllCmd(),
		GetMetadataScopeCmd(),
		GetMetadataSessionCmd(),
		GetMetadataRecordCmd(),
		GetMetadataScopeSpecCmd(),
		GetMetadataContractSpecCmd(),
		GetMetadataRecordSpecCmd(),
		GetOwnershipCmd(),
		GetValueOwnershipCmd(),
		GetOSLocatorCmd(),
	)
	return queryCmd
}

// ------------ Individual Commands ------------

// GetMetadataParamsCmd returns the command handler for metadata parameter querying.
func GetMetadataParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "params [locator]",
		Aliases: []string{"p"},
		Short:   "Query the current metadata parameters",
		Args:    cobra.MaximumNArgs(1),
		Example: fmt.Sprintf("%s params", cmdStart),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return outputParams(cmd)
			}
			arg0 := strings.TrimSpace(args[0])
			if arg0 == "locator" {
				return outputOSLocatorParams(cmd)
			}
			return fmt.Errorf("unknown argument: %s", arg0)
		},
	}

	addIncludeRequestFlag(cmd)
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetMetadataByIDCmd returns the command handler for querying metadata for anything from an id.
func GetMetadataByIDCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get id",
		Aliases: []string{"g"},
		Short:   "Query the current metadata by id",
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
				return outputScope(cmd, id.String(), "", "")
			case types.PrefixSession:
				return outputSessions(cmd, "", id.String(), "", "")
			case types.PrefixRecord:
				return outputRecords(cmd, id.String(), "", "", "")
			case types.PrefixScopeSpecification:
				return outputScopeSpec(cmd, id.String())
			case types.PrefixContractSpecification:
				return outputContractSpec(cmd, id.String())
			case types.PrefixRecordSpecification:
				return outputRecordSpec(cmd, id.String(), "")
			}
			return fmt.Errorf("unexpected address prefix %s", prefix)
		},
	}

	addIncludeScopeFlag(cmd)
	addIncludeSessionsFlag(cmd)
	addIncludeRecordsFlag(cmd)
	addIncludeRecordSpecsFlag(cmd)
	addIncludeRequestFlag(cmd)
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetMetadataGetAllCmd returns the command handler for querying metadata for all entries of a type.
func GetMetadataGetAllCmd() *cobra.Command {
	nonLetterRegex := regexp.MustCompile("[^[:alpha:]]+")
	cmd := &cobra.Command{
		Use:     "all {scopes|sessions|records|scopespecs|contractspecs|recordspecs|locators}",
		Aliases: []string{"a"},
		Short:   "Get all entries of a certain type",
		Long: fmt.Sprintf(`%[1]s all scopes - gets all scopes.
%[1]s all sessions - gets all sessions.
%[1]s all records - gets all records.
%[1]s all scopespecs - gets all scope specifications.
%[1]s all contractspecs - gets all contract specifications.
%[1]s all recordspecs - gets all record specifications.
%[1]s all locators - gets all object store locators.`, cmdStart),
		Example: fmt.Sprintf(`%[1]s all scopes
%[1]s all sessions
%[1]s all records
%[1]s all scopespecs
%[1]s all contractspecs
%[1]s all recordspecs
%[1]s all locators`, cmdStart),
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Smash all the args together. We only want a single "word" anyway.
			input := strings.ToLower(trimSpaceAndJoin(args, ""))
			// Get rid of non-letters
			// This simplifies the switch below, e.g. record-specs becomes recordspecs.
			input = nonLetterRegex.ReplaceAllString(input, "")
			// Make sure it ends with an "s"
			// This simplifies the switch below, e.g. "scope" becomes "scopes"
			if input[len(input)-1:] != "s" {
				input += "s"
			}
			switch input {
			case "scopes":
				return outputScopesAll(cmd)
			case "sessions", "sess":
				return outputSessionsAll(cmd)
			case "records", "recs":
				return outputRecordsAll(cmd)
			case "scopespecs", "scopespecifications":
				return outputScopeSpecsAll(cmd)
			case "contractspecs", "cspecs", "contractspecifications":
				return outputContractSpecsAll(cmd)
			case "recordspecs", "recspecs", "recordspecifications", "recspecifications":
				return outputRecordSpecsAll(cmd)
			case "locators", "locs":
				return outputOSLocatorsAll(cmd)
			}
			return fmt.Errorf("unknown entry type: %s", input)
		},
	}

	addIncludeRequestFlag(cmd)
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "entries")

	return cmd
}

// GetMetadataScopeCmd returns the command handler for metadata scope querying.
func GetMetadataScopeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "scope {scope_id|scope_uuid|session_id|record_id|\"all\"}",
		Aliases: []string{"sc", "scopes"},
		Short:   "Query the current metadata for a scope",
		Long: fmt.Sprintf(`%[1]s scope {scope_id} - gets the scope with the given id.
%[1]s scope {scope_uuid} - gets the scope with the given uuid.
%[1]s scope {session_id} - gets the scope containing the given session.
%[1]s scope {record_id} - gets the scope containing the given record.
%[1]s scope all - gets all scopes.`, cmdStart),
		Args: cobra.ExactArgs(1),
		Example: fmt.Sprintf(`%[1]s scope scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel
%[1]s scope 91978ba2-5f35-459a-86a7-feca1b0512e0
%[1]s scope session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr
%[1]s scope record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3
%[1]s scope all`, cmdStart),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg0 := strings.TrimSpace(args[0])
			if arg0 == all {
				return outputScopesAll(cmd)
			}
			id, idErr := types.MetadataAddressFromBech32(arg0)
			if idErr == nil {
				switch {
				case id.IsScopeAddress():
					return outputScope(cmd, id.String(), "", "")
				case id.IsSessionAddress():
					return outputScope(cmd, "", id.String(), "")
				case id.IsRecordAddress():
					return outputScope(cmd, "", "", id.String())
				}
			}
			return outputScope(cmd, arg0, "", "")
		},
	}

	addIncludeSessionsFlag(cmd)
	addIncludeRecordsFlag(cmd)
	addIncludeRequestFlag(cmd)
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "scopes (all)")

	return cmd
}

// GetMetadataSessionCmd returns the command handler for metadata session querying.
func GetMetadataSessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "session {session_id|{scope_id|scope_uuid} [session_uuid|record_name]|record_id|\"all\"}",
		Aliases: []string{"se", "sessions"},
		Short:   "Query the current metadata for sessions",
		Long: fmt.Sprintf(`%[1]s session {session_id} - gets the session with the given id.
%[1]s session {scope_id} - gets the list of sessions associated with a scope.
%[1]s session {scope_id} {session_uuid} - gets a session with the given scope and session uuid.
%[1]s session {scope_id} {record_name} - gets the session in the given scope containing the given record.
%[1]s session {scope_uuid} - gets the list of sessions associated with a scope.
%[1]s session {scope_uuid} {session_uuid} - gets a session with the given scope uuid and session uuid.
%[1]s session {scope_uuid} {record_name} - gets the session in the given scope containing the given record.
%[1]s session {record_id} - gets the session containing the given record.
%[1]s session all - gets all sessions.`, cmdStart),
		Args: cobra.RangeArgs(1, 2),
		Example: fmt.Sprintf(`%[1]s session session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr
%[1]s session scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel
%[1]s session scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel 5803f8bc-6067-4eb5-951f-2121671c2ec0
%[1]s session scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel recordname
%[1]s session 91978ba2-5f35-459a-86a7-feca1b0512e0
%[1]s session 91978ba2-5f35-459a-86a7-feca1b0512e0 5803f8bc-6067-4eb5-951f-2121671c2ec0
%[1]s session 91978ba2-5f35-459a-86a7-feca1b0512e0 recordname
%[1]s session record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3
%[1]s session all`, cmdStart),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg0 := strings.TrimSpace(args[0])
			if arg0 == all {
				return outputSessionsAll(cmd)
			}
			if len(args) == 2 {
				arg1 := strings.TrimSpace(args[1])
				if sessionUUID, err := uuid.Parse(arg1); err == nil {
					return outputSessions(cmd, arg0, sessionUUID.String(), "", "")
				}
				return outputSessions(cmd, arg0, "", "", arg1)
			}
			id, idErr := types.MetadataAddressFromBech32(arg0)
			if idErr == nil {
				switch {
				case id.IsScopeAddress():
					return outputSessions(cmd, id.String(), "", "", "")
				case id.IsSessionAddress():
					return outputSessions(cmd, "", id.String(), "", "")
				case id.IsRecordAddress():
					return outputSessions(cmd, "", "", id.String(), "")
				}
			}
			return outputSessions(cmd, arg0, "", "", "")
		},
	}

	addIncludeScopeFlag(cmd)
	addIncludeRecordsFlag(cmd)
	addIncludeRequestFlag(cmd)
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "sessions (all)")

	return cmd
}

// GetMetadataRecordCmd returns the command handler for metadata record querying.
func GetMetadataRecordCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "record {record_id|{session_id|scope_id|scope_uuid} [record_name]|\"all\"}",
		Aliases: []string{"r", "records"},
		Short:   "Query the current metadata for records",
		Long: fmt.Sprintf(`%[1]s record {record_id} - gets the record with the given id.
%[1]s record {session_id} - gets the list of records associated with a session id.
%[1]s record {session_id} {record_name} - gets the record with the given name the scope containing that session.
%[1]s record {scope_id} - gets the list of records associated with a scope id.
%[1]s record {scope_id} {record_name} - gets the record with the given name from the given scope.
%[1]s record {scope_uuid} - gets the list of records associated with a scope uuid.
%[1]s record {scope_uuid} {record_name} - gets the record with the given name from the given scope.
%[1]s record all - all records.`, cmdStart),
		Args: cobra.MinimumNArgs(1),
		Example: fmt.Sprintf(`%[1]s record record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3
%[1]s record session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr
%[1]s record session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr recordname
%[1]s record scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel
%[1]s record scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel recordname
%[1]s record 91978ba2-5f35-459a-86a7-feca1b0512e0
%[1]s record 91978ba2-5f35-459a-86a7-feca1b0512e0 recordname
%[1]s record all`, cmdStart),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg0 := strings.TrimSpace(args[0])
			if arg0 == all {
				return outputRecordsAll(cmd)
			}
			name := ""
			if len(args) > 1 {
				name = trimSpaceAndJoin(args[1:], " ")
			}
			id, idErr := types.MetadataAddressFromBech32(arg0)
			if idErr == nil {
				switch {
				case id.IsRecordAddress():
					return outputRecords(cmd, id.String(), "", "", name)
				case id.IsScopeAddress():
					return outputRecords(cmd, "", id.String(), "", name)
				case id.IsSessionAddress():
					return outputRecords(cmd, "", "", id.String(), name)
				}
			}
			return outputRecords(cmd, "", arg0, "", name)
		},
	}

	addIncludeScopeFlag(cmd)
	addIncludeSessionsFlag(cmd)
	addIncludeRequestFlag(cmd)
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "records (all)")

	return cmd
}

// GetMetadataScopeSpecCmd returns the command handler for metadata scope specification querying.
func GetMetadataScopeSpecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "scopespec {scope_spec_id|scope_spec_uuid|\"all\"}",
		Aliases: []string{"ss", "scopespecification", "scopespecs", "scopespecifications"},
		Short:   "Query the current metadata for a scope specification",
		Long: fmt.Sprintf(`%[1]s scopespec {scope_spec_id} - gets the scope specification for that a given id.
%[1]s scopespec {scope_spec_uuid} - gets the scope specification for a given uuid.
%[1]s scopespec all - gets all the scope specifications.`, cmdStart),
		Example: fmt.Sprintf(`%[1]s scopespec scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m
%[1]s scopespec dc83ea70-eacd-40fe-9adf-1cf6148bf8a2
%[1]s scopespec all`, cmdStart),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg0 := strings.TrimSpace(args[0])
			if arg0 == all {
				return outputScopeSpecsAll(cmd)
			}
			return outputScopeSpec(cmd, arg0)
		},
	}

	addIncludeRequestFlag(cmd)
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "scope specifications (all)")

	return cmd
}

// GetMetadataContractSpecCmd returns the command handler for metadata contract specification querying.
func GetMetadataContractSpecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "contractspec {contract_spec_id|contract_spec_uuid|record_spec_id|\"all\"}",
		Aliases: []string{"cs", "contractspecification", "contractspecs", "contractspecifications"},
		Short:   "Query the current metadata for a contract specification",
		Long: fmt.Sprintf(`%[1]s contractspec {contract_spec_id} - gets the contract specification for that a given id.
%[1]s contractspec {contract_spec_uuid} - gets the contract specification for a given uuid.
%[1]s contractspec {record_spec_id} - gets the contract specification associated with that record spec id.
%[1]s contractspec all - gets all the contract specifications.`, cmdStart),
		Args: cobra.ExactArgs(1),
		Example: fmt.Sprintf(`%[1]s contractspec contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn
%[1]s contractspec def6bc0a-c9dd-4874-948f-5206e6060a84
%[1]s contractspec recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44
%[1]s contractspec all`, cmdStart),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg0 := strings.TrimSpace(args[0])
			if arg0 == all {
				return outputContractSpecsAll(cmd)
			}
			return outputContractSpec(cmd, arg0)
		},
	}

	addIncludeRecordSpecsFlag(cmd)
	addIncludeRequestFlag(cmd)
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "contract specifications (all)")

	return cmd
}

// GetMetadataRecordSpecCmd returns the command handler for metadata record specification querying.
func GetMetadataRecordSpecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "recordspec {rec_spec_id|{contract_spec_id|contract_spec_uuid} [record_name]|\"all\"}",
		Aliases: []string{"rs", "recordspecs", "recspec", "recspecs", "recordspecification", "recordspecifications"},
		Short:   "Query the current metadata for record specifications",
		Long: fmt.Sprintf(`%[1]s recordspec {rec_spec_id} - gets the record specification for a given id.
%[1]s recordspec {contract_spec_id} - gets the list of record specifications for the given contract specification.
%[1]s recordspec {contract_spec_id} {record_name} - gets the record specification for a given contract specification and record name.
%[1]s recordspec {contract_spec_uuid} - gets the list of record specifications for the given contract specification.
%[1]s recordspec {contract_spec_uuid} {record_name} - gets the record specification for a given contract specification and record name.
%[1]s recordspec all - gets all the record specifications`, cmdStart),
		Args: cobra.MinimumNArgs(1),
		Example: fmt.Sprintf(`%[1]s recordspec recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44
%[1]s recordspec contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn
%[1]s recordspec contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn recordname
%[1]s recordspec def6bc0a-c9dd-4874-948f-5206e6060a84
%[1]s recordspec def6bc0a-c9dd-4874-948f-5206e6060a84 recordname
%[1]s recordspec all`, cmdStart),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg0 := strings.TrimSpace(args[0])
			if arg0 == all {
				return outputRecordSpecsAll(cmd)
			}
			name := ""
			if len(args) > 1 {
				name = trimSpaceAndJoin(args[1:], " ")
			}
			if len(name) == 0 {
				return outputRecordSpecsForContractSpec(cmd, arg0)
			}
			return outputRecordSpec(cmd, arg0, name)
		},
	}

	addIncludeRequestFlag(cmd)
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "record specifications (all)")

	return cmd
}

// GetOwnershipCmd returns the command handler for metadata entry querying by owner address
func GetOwnershipCmd() *cobra.Command {
	// Note: Once we get queries for ownership of things other than scopes,
	// we can add a 2nd argument to this command to restrict the search to one specific type.
	// E.g. Use: "owner {address} [scope|session|record|scopespec|contractspec|recordspec]"
	cmd := &cobra.Command{
		Use:     "owner address",
		Aliases: []string{"o", "ownership"},
		Short:   "Query the current metadata for entries owned by an address",
		Long:    fmt.Sprintf(`%[1]s owner {address} - gets a list of scope uuids owned by the provided address.`, cmdStart),
		Args:    cobra.ExactArgs(1),
		Example: fmt.Sprintf(`%[1]s owner pb1sh49f6ze3vn7cdl2amh2gnc70z5mten3dpvr42`, cmdStart),
		RunE: func(cmd *cobra.Command, args []string) error {
			address := strings.TrimSpace(args[0])
			if len(address) == 0 {
				return fmt.Errorf("empty address")
			}
			return outputOwnership(cmd, address)
		},
	}

	addIncludeRequestFlag(cmd)
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "scopes")

	return cmd
}

// GetValueOwnershipCmd returns the command handler for metadata scope querying by owner address
func GetValueOwnershipCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "valueowner address",
		Aliases: []string{"vo", "valueownership"},
		Short:   "Query the current metadata for scopes with the provided address as the value owner",
		Long:    fmt.Sprintf(`%[1]s valueowner {address} - gets a list of scope uuids value-owned by the provided address.`, cmdStart),
		Args:    cobra.ExactArgs(1),
		Example: fmt.Sprintf(`%[1]s valueowner pb1sh49f6ze3vn7cdl2amh2gnc70z5mten3dpvr42`, cmdStart),
		RunE: func(cmd *cobra.Command, args []string) error {
			address := strings.TrimSpace(args[0])
			if len(address) == 0 {
				return fmt.Errorf("empty address")
			}
			return outputValueOwnership(cmd, address)
		},
	}

	addIncludeRequestFlag(cmd)
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "scopes")

	return cmd
}

// GetOSLocatorCmd returns the command handler for metadata object store locator querying.
func GetOSLocatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "locator {owner|scope_id|scope_uuid|uri|\"params\"|\"all\"}",
		Aliases: []string{"l", "locators"},
		Short:   "Query the current metadata for object store locators",
		Long: fmt.Sprintf(`%[1]s locator {owner} - gets the object store locator for that owner.
%[1]s locator {scope_id} - gets object store locators for all the owners of that scope.
%[1]s locator {scope_uuid} - gets object store locators for all the owners of that scope.
%[1]s locator {uri} - gets object store locators with that uri.
%[1]s locator params - gets the object store locator params.
%[1]s locator all - gets all object store locators.`, cmdStart),
		Args: cobra.ExactArgs(1),
		Example: fmt.Sprintf(`%[1]s locator pb1sh49f6ze3vn7cdl2amh2gnc70z5mten3dpvr42
%[1]s locator scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel
%[1]s locator 91978ba2-5f35-459a-86a7-feca1b0512e0
%[1]s locator https://provenance.io/
%[1]s locator params
%[1]s locator all`, cmdStart),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg0 := strings.TrimSpace(args[0])
			// First check if it's just the string "params".
			if arg0 == "params" {
				return outputOSLocatorParams(cmd)
			}
			// And then look for "all".
			if arg0 == all {
				return outputOSLocatorsAll(cmd)
			}
			// Now look to see if it's a metadata address.
			_, idErr := types.MetadataAddressFromBech32(arg0)
			if idErr == nil {
				return outputOSLocatorsByScope(cmd, arg0)
			}
			// Okay... maybe check for a generic bech32.
			_, _, bech32Err := bech32.DecodeAndConvert(arg0)
			if bech32Err == nil {
				return outputOSLocator(cmd, arg0)
			}
			// Well maybe a UUID?
			_, uuidErr := uuid.Parse(arg0)
			if uuidErr == nil {
				return outputOSLocatorsByScope(cmd, arg0)
			}
			// Fine... let's just hope it's a URI.
			return outputOSLocatorsByURI(cmd, arg0)
		},
	}

	addIncludeRequestFlag(cmd)
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "locators (all)")

	return cmd
}

// ------------ private funcs for actually querying and outputting ------------

// outputParams calls the Params query and outputs the response.
func outputParams(cmd *cobra.Command) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	queryClient := types.NewQueryClient(clientCtx)
	res, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
	if err != nil {
		return err
	}

	if !includeRequest {
		res.Request = nil
	}

	return clientCtx.PrintProto(res)
}

// outputScope calls the Scope query and outputs the response.
func outputScope(cmd *cobra.Command, scopeID string, sessionAddr string, recordAddr string) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	req := types.ScopeRequest{
		ScopeId:         scopeID,
		SessionAddr:     sessionAddr,
		RecordAddr:      recordAddr,
		IncludeSessions: includeSessions,
		IncludeRecords:  includeRecords,
	}

	queryClient := types.NewQueryClient(clientCtx)
	res, err := queryClient.Scope(context.Background(), &req)
	if err != nil {
		return err
	}

	if !includeRequest {
		res.Request = nil
	}

	return clientCtx.PrintProto(res)
}

// outputScopesAll calls the ScopesAllRequest query and outputs the response.
func outputScopesAll(cmd *cobra.Command) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}
	pageReq, e := client.ReadPageRequest(withPageKeyDecoded(cmd.Flags()))
	if e != nil {
		return e
	}
	queryClient := types.NewQueryClient(clientCtx)
	res, err := queryClient.ScopesAll(
		context.Background(),
		&types.ScopesAllRequest{Pagination: pageReq},
	)
	if err != nil {
		return err
	}

	if !includeRequest {
		res.Request = nil
	}

	return clientCtx.PrintProto(res)
}

// outputSessions calls the Sessions query and outputs the response.
func outputSessions(cmd *cobra.Command, scopeID, sessionID, recordID, recordName string) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	req := types.SessionsRequest{
		ScopeId:        scopeID,
		SessionId:      sessionID,
		RecordAddr:     recordID,
		RecordName:     recordName,
		IncludeScope:   includeScope,
		IncludeRecords: includeRecords,
	}

	queryClient := types.NewQueryClient(clientCtx)
	res, err := queryClient.Sessions(context.Background(), &req)
	if err != nil {
		return err
	}
	if res == nil || res.Sessions == nil || len(res.Sessions) == 0 {
		return errors.New("no sessions found")
	}

	if !includeRequest {
		res.Request = nil
	}

	return clientCtx.PrintProto(res)
}

// outputSessionsAll calls the SessionsAll query and outputs the response.
func outputSessionsAll(cmd *cobra.Command) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}
	pageReq, e := client.ReadPageRequest(withPageKeyDecoded(cmd.Flags()))
	if e != nil {
		return e
	}
	queryClient := types.NewQueryClient(clientCtx)
	res, err := queryClient.SessionsAll(
		context.Background(),
		&types.SessionsAllRequest{Pagination: pageReq},
	)
	if err != nil {
		return err
	}

	if !includeRequest {
		res.Request = nil
	}

	return clientCtx.PrintProto(res)
}

// outputRecords calls the Records query and outputs the response.
func outputRecords(cmd *cobra.Command, recordAddr string, scopeID string, sessionID string, name string) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	req := types.RecordsRequest{
		RecordAddr:      recordAddr,
		ScopeId:         scopeID,
		SessionId:       sessionID,
		Name:            name,
		IncludeScope:    includeScope,
		IncludeSessions: includeSessions,
	}

	queryClient := types.NewQueryClient(clientCtx)
	res, err := queryClient.Records(context.Background(), &req)
	if err != nil {
		return err
	}

	if !includeRequest {
		res.Request = nil
	}

	return clientCtx.PrintProto(res)
}

// outputRecordsAll calls the RecordsAll query and outputs the response.
func outputRecordsAll(cmd *cobra.Command) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}
	pageReq, e := client.ReadPageRequest(withPageKeyDecoded(cmd.Flags()))
	if e != nil {
		return e
	}
	queryClient := types.NewQueryClient(clientCtx)
	res, err := queryClient.RecordsAll(
		context.Background(),
		&types.RecordsAllRequest{Pagination: pageReq},
	)
	if err != nil {
		return err
	}

	if !includeRequest {
		res.Request = nil
	}

	return clientCtx.PrintProto(res)
}

// outputOwnership calls the Ownership query and outputs the response.
func outputOwnership(cmd *cobra.Command, address string) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}
	pageReq, e := client.ReadPageRequest(withPageKeyDecoded(cmd.Flags()))
	if e != nil {
		return e
	}
	queryClient := types.NewQueryClient(clientCtx)
	res, err := queryClient.Ownership(
		context.Background(),
		&types.OwnershipRequest{Address: address, Pagination: pageReq},
	)
	if err != nil {
		return err
	}

	if !includeRequest {
		res.Request = nil
	}

	return clientCtx.PrintProto(res)
}

// outputValueOwnership calls the ValueOwnership query and outputs the response.
func outputValueOwnership(cmd *cobra.Command, address string) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}
	pageReq, e := client.ReadPageRequest(withPageKeyDecoded(cmd.Flags()))
	if e != nil {
		return e
	}
	queryClient := types.NewQueryClient(clientCtx)
	res, err := queryClient.ValueOwnership(
		context.Background(),
		&types.ValueOwnershipRequest{Address: address, Pagination: pageReq},
	)
	if err != nil {
		return err
	}

	if !includeRequest {
		res.Request = nil
	}

	return clientCtx.PrintProto(res)
}

// outputScopeSpec calls the ScopeSpecification query and outputs the response.
func outputScopeSpec(cmd *cobra.Command, specificationID string) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	req := types.ScopeSpecificationRequest{
		SpecificationId: specificationID,
	}

	queryClient := types.NewQueryClient(clientCtx)
	res, err := queryClient.ScopeSpecification(context.Background(), &req)
	if err != nil {
		return err
	}

	if !includeRequest {
		res.Request = nil
	}

	return clientCtx.PrintProto(res)
}

// outputScopeSpecsAll calls the ScopeSpecificationsAll query and outputs the response.
func outputScopeSpecsAll(cmd *cobra.Command) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}
	pageReq, e := client.ReadPageRequest(withPageKeyDecoded(cmd.Flags()))
	if e != nil {
		return e
	}
	queryClient := types.NewQueryClient(clientCtx)
	res, err := queryClient.ScopeSpecificationsAll(
		context.Background(),
		&types.ScopeSpecificationsAllRequest{Pagination: pageReq},
	)
	if err != nil {
		return err
	}

	if !includeRequest {
		res.Request = nil
	}

	return clientCtx.PrintProto(res)
}

// outputContractSpec calls the ContractSpecification query and outputs the response.
func outputContractSpec(cmd *cobra.Command, specificationID string) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	req := types.ContractSpecificationRequest{
		SpecificationId:    specificationID,
		IncludeRecordSpecs: includeRecordSpecs,
	}

	queryClient := types.NewQueryClient(clientCtx)
	res, err := queryClient.ContractSpecification(context.Background(), &req)
	if err != nil {
		return err
	}

	if !includeRequest {
		res.Request = nil
	}

	return clientCtx.PrintProto(res)
}

// outputContractSpecsAll calls the ContractSpecificationsAll query and outputs the response.
func outputContractSpecsAll(cmd *cobra.Command) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}
	pageReq, e := client.ReadPageRequest(withPageKeyDecoded(cmd.Flags()))
	if e != nil {
		return e
	}
	queryClient := types.NewQueryClient(clientCtx)
	res, err := queryClient.ContractSpecificationsAll(
		context.Background(),
		&types.ContractSpecificationsAllRequest{Pagination: pageReq},
	)
	if err != nil {
		return err
	}

	if !includeRequest {
		res.Request = nil
	}

	return clientCtx.PrintProto(res)
}

// outputRecordSpec calls the RecordSpecification query and outputs the response.
func outputRecordSpec(cmd *cobra.Command, specificationID string, name string) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	req := types.RecordSpecificationRequest{
		SpecificationId: specificationID,
		Name:            name,
	}

	queryClient := types.NewQueryClient(clientCtx)
	res, err := queryClient.RecordSpecification(context.Background(), &req)
	if err != nil {
		return err
	}

	if !includeRequest {
		res.Request = nil
	}

	return clientCtx.PrintProto(res)
}

// outputRecordSpecsForContractSpec calls the RecordSpecificationsForContractSpecification query and outputs the response.
func outputRecordSpecsForContractSpec(cmd *cobra.Command, specificationID string) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	req := types.RecordSpecificationsForContractSpecificationRequest{
		SpecificationId: specificationID,
	}

	queryClient := types.NewQueryClient(clientCtx)
	res, err := queryClient.RecordSpecificationsForContractSpecification(context.Background(), &req)
	if err != nil {
		return err
	}

	if !includeRequest {
		res.Request = nil
	}

	return clientCtx.PrintProto(res)
}

// outputRecordSpecsAll calls the RecordSpecificationsAll query and outputs the response.
func outputRecordSpecsAll(cmd *cobra.Command) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}
	pageReq, e := client.ReadPageRequest(withPageKeyDecoded(cmd.Flags()))
	if e != nil {
		return e
	}
	queryClient := types.NewQueryClient(clientCtx)
	res, err := queryClient.RecordSpecificationsAll(
		context.Background(),
		&types.RecordSpecificationsAllRequest{Pagination: pageReq},
	)
	if err != nil {
		return err
	}

	if !includeRequest {
		res.Request = nil
	}

	return clientCtx.PrintProto(res)
}

// outputOSLocatorParams calls the OSLocatorParams query and outputs the response.
func outputOSLocatorParams(cmd *cobra.Command) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}
	queryClient := types.NewQueryClient(clientCtx)
	res, err := queryClient.OSLocatorParams(
		context.Background(),
		&types.OSLocatorParamsRequest{},
	)
	if err != nil {
		return err
	}

	if !includeRequest {
		res.Request = nil
	}

	return clientCtx.PrintProto(res)
}

// outputOSLocator calls the OSLocator query and outputs the response.
func outputOSLocator(cmd *cobra.Command, owner string) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}
	queryClient := types.NewQueryClient(clientCtx)
	res, err := queryClient.OSLocator(
		context.Background(),
		&types.OSLocatorRequest{Owner: owner},
	)
	if err != nil {
		return err
	}

	if !includeRequest {
		res.Request = nil
	}

	return clientCtx.PrintProto(res)
}

// outputOSLocatorsByURI calls the OSLocatorsByURI query and outputs the response.
func outputOSLocatorsByURI(cmd *cobra.Command, uri string) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}
	pageReq, e := client.ReadPageRequest(withPageKeyDecoded(cmd.Flags()))
	if e != nil {
		return e
	}
	queryClient := types.NewQueryClient(clientCtx)
	res, err := queryClient.OSLocatorsByURI(
		context.Background(),
		&types.OSLocatorsByURIRequest{Uri: uri, Pagination: pageReq},
	)
	if err != nil {
		return err
	}

	if !includeRequest {
		res.Request = nil
	}

	return clientCtx.PrintProto(res)
}

// outputOSLocatorsByScope calls the OSLocatorsByScope query and outputs the response.
func outputOSLocatorsByScope(cmd *cobra.Command, scopeID string) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}
	queryClient := types.NewQueryClient(clientCtx)
	res, err := queryClient.OSLocatorsByScope(
		context.Background(),
		&types.OSLocatorsByScopeRequest{ScopeId: scopeID},
	)
	if err != nil {
		return err
	}

	if !includeRequest {
		res.Request = nil
	}

	return clientCtx.PrintProto(res)
}

// outputOSLocatorsAll calls the OSAllLocators query and outputs the response.
func outputOSLocatorsAll(cmd *cobra.Command) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}
	pageReq, e := client.ReadPageRequest(withPageKeyDecoded(cmd.Flags()))
	if e != nil {
		return e
	}
	queryClient := types.NewQueryClient(clientCtx)
	res, err := queryClient.OSAllLocators(
		context.Background(),
		&types.OSAllLocatorsRequest{Pagination: pageReq},
	)
	if err != nil {
		return err
	}

	if !includeRequest {
		res.Request = nil
	}

	return clientCtx.PrintProto(res)
}

// ------------ private generic helper functions ------------

// trimSpaceAndJoin trims leading and trailing whitespace from each arg,
// then joins them using the provided sep string,
// then lastly trims any left over leading and trailing whitespace from that result.
func trimSpaceAndJoin(args []string, sep string) string {
	trimmedArgs := make([]string, len(args))
	for i, arg := range args {
		trimmedArgs[i] = strings.TrimSpace(arg)
	}
	return strings.TrimSpace(strings.Join(trimmedArgs, sep))
}

// addIncludeScopeFlag sets up a command to look for an --include-scope flag.
// The flag value is tied to the includeScope variable.
func addIncludeScopeFlag(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&includeScope, "include-scope", false, "include the scope in the output")
}

// addIncludeSessionsFlag sets up a command to look for an --include-sessions flag.
// The flag value is tied to the includeSessions variable.
func addIncludeSessionsFlag(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&includeSessions, "include-sessions", false, "include sessions in the output")
}

// addIncludeRecordsFlag sets up a command to look for an --include-records flag.
// The flag value is tied to the includeRecords variable.
func addIncludeRecordsFlag(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&includeRecords, "include-records", false, "include records in the output")
}

// addIncludeRecordSpecsFlag sets up a command to look for an --include-record-specs.
// The flag value is tied to the includeRecordSpecs variable.
func addIncludeRecordSpecsFlag(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&includeRecordSpecs, "include-record-specs", false, "include record specs in the output")
}

// addIncludeRequestFlag sets up a command to look for an --include-request.
// The flag value is tied to the includeRequest variable.
func addIncludeRequestFlag(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&includeRequest, "include-request", false, "include the query request in the output")
}

// sdk ReadPageRequest expects binary but we encoded to base64 in our marshaller
func withPageKeyDecoded(flagSet *flag.FlagSet) *flag.FlagSet {
	encoded, err := flagSet.GetString(flags.FlagPageKey)
	if err != nil {
		panic(err.Error())
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		panic(err.Error())
	}
	_ = flagSet.Set(flags.FlagPageKey, string(raw))
	return flagSet
}
