package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/provenance-io/provenance/x/metadata/types"
)

var (
	cmdStart = fmt.Sprintf("%s metaaddress", version.AppName)
)

// GetQueryCmd is the top-level command for name CLI queries.
func AddMetaAddressCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        "metaaddress",
		Aliases:                    []string{"ma"},
		Short:                      "Decode/Encode Metaaddresses commands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	queryCmd.AddCommand(
		AddMetaAddressEncoder(),
		AddMetaAddressDecoder(),
	)

	return queryCmd
}

// AddMetaAddressDecoder returns metadata address parser cobra Command.
func AddMetaAddressDecoder() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "decode [address]",
		Aliases: []string{"d"},
		Short:   "Decode MetadataAddress and display associate IDs and types",
		Example: fmt.Sprintf("%[1]s decode scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel", cmdStart),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			addr, parseErr := types.MetadataAddressFromBech32(args[0])
			if parseErr != nil {
				return parseErr
			}
			addrDetails := addr.GetDetails()
			var toOut string
			switch {
			case addr.IsScopeAddress():
				toOut = fmt.Sprintf(`Type: Scope
Scope UUID: %s
`, addrDetails.PrimaryUUID)
			case addr.IsSessionAddress():
				toOut = fmt.Sprintf(`Type: Session
Scope Id: %s
Scope UUID: %s
Session UUID: %s
`, addrDetails.ParentAddress, addrDetails.PrimaryUUID, addrDetails.SecondaryUUID)
			case addr.IsRecordAddress():
				toOut = fmt.Sprintf(`Type: Record
Scope Id: %s
Scope UUID: %s
Name Hash (hex): %s
`, addrDetails.ParentAddress, addrDetails.PrimaryUUID, addrDetails.NameHashHex)
			case addr.IsScopeSpecificationAddress():
				toOut = fmt.Sprintf(`Type: Scope Specification
Scope Specification UUID: %s
`, addrDetails.PrimaryUUID)
			case addr.IsContractSpecificationAddress():
				toOut = fmt.Sprintf(`Type: Contract Specification
Contract Specification UUID: %s
`, addrDetails.PrimaryUUID)
			case addr.IsRecordSpecificationAddress():
				toOut = fmt.Sprintf(`Type: Record Specification
Contract Specification Id: %s
Contract Specification UUID: %s
Name Hash (hex): %s
`, addrDetails.ParentAddress, addrDetails.PrimaryUUID, addrDetails.NameHashHex)
			default:
				toOut = fmt.Sprintf(`Type: UNKNOWN
prefix: %s
primary UUID: %s
secondary UUID: %s
Name Hash (hex): %s
Excess (hex): %s
`, addrDetails.Prefix, addrDetails.PrimaryUUID, addrDetails.SecondaryUUID, addrDetails.NameHashHex, addrDetails.ExcessHex)
			}
			_, cmdErr := fmt.Fprint(cmd.OutOrStdout(), toOut)
			return cmdErr
		},
	}
	return cmd
}

// AddMetaAddressEncoder returns metadata address encoder cobra Command.
func AddMetaAddressEncoder() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "encode [type] [uuid] [uuid|name]",
		Short: "Encodes metadata uuids to bech32 address for specific type",
		Long: fmt.Sprintf(`Encodes metadata uuids to bech32 address for specific type.

%[1]s encode type uuid [uuid|name]

Types: scope session record scope-specification contract-specification record-specification

The type and first uuid argument are required.
The third [uuid|name] argument is either required or forbidden based on the type.

These types forbid a third argument: scope scope-specification contract-specification
These types require a third argument: session record record-specification
This type requires the third argument to be a UUID: session`, cmdStart),
		Example: fmt.Sprintf(`%[1]s encode scope 91978ba2-5f35-459a-86a7-feca1b0512e0
%[1]s encode session 91978ba2-5f35-459a-86a7-feca1b0512e0 5803f8bc-6067-4eb5-951f-2121671c2ec0
%[1]s encode record 91978ba2-5f35-459a-86a7-feca1b0512e0 recordname
%[1]s encode scope-specification dc83ea70-eacd-40fe-9adf-1cf6148bf8a2
%[1]s encode contract-specification def6bc0a-c9dd-4874-948f-5206e6060a84
%[1]s encode record-specification def6bc0a-c9dd-4874-948f-5206e6060a84 recordname`, cmdStart),
		Args: cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			addrType := strings.ToLower(regexp.MustCompile("[^[:alpha:]]+").ReplaceAllString(args[0], ""))
			primaryUUID, err := uuid.Parse(args[1])
			if err != nil {
				return err
			}
			var uuidOrNameArg string
			argsLen := len(args)
			if argsLen == 3 {
				uuidOrNameArg = strings.TrimSpace(args[2])
				if len(uuidOrNameArg) == 0 {
					argsLen--
				}
			}
			var addr types.MetadataAddress
			switch addrType {
			case "scope":
				if argsLen != 2 {
					return fmt.Errorf("too many arguments for %s address encoder", addrType)
				}
				addr = types.ScopeMetadataAddress(primaryUUID)
			case "session":
				if argsLen != 3 {
					return fmt.Errorf("not enough arguments for %s address encoder", addrType)
				}
				secondaryUUID, err := uuid.Parse(uuidOrNameArg)
				if err != nil {
					return err
				}
				addr = types.SessionMetadataAddress(primaryUUID, secondaryUUID)
			case "record":
				if argsLen != 3 {
					return fmt.Errorf("not enough arguments for %s address encoder", addrType)
				}
				addr = types.RecordMetadataAddress(primaryUUID, uuidOrNameArg)
			case "scopespecification", "scopespec":
				if argsLen != 2 {
					return fmt.Errorf("too many arguments for %s address encoder", "scope-specification")
				}
				addr = types.ScopeSpecMetadataAddress(primaryUUID)
			case "contractspecification", "contractspec", "cspec":
				if argsLen != 2 {
					return fmt.Errorf("too many arguments for %s address encoder", "contract-specification")
				}
				addr = types.ContractSpecMetadataAddress(primaryUUID)
			case "recordspecification", "recordspec", "recspec":
				if argsLen != 3 {
					return fmt.Errorf("not enough arguments for %s address encoder", "record-specification")
				}
				addr = types.RecordSpecMetadataAddress(primaryUUID, uuidOrNameArg)
			default:
				return fmt.Errorf("unknown type: %s, Supported types: scope session record scope-specification contract-specification record-specification", args[0])
			}
			_, cmdErr := fmt.Fprintf(cmd.OutOrStdout(), "%s\n", addr)
			return cmdErr
		},
	}
	return cmd
}
