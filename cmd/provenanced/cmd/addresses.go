package cmd

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/provenance-io/provenance/x/metadata/types"
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
		Use:     "decode address",
		Aliases: []string{"d"},
		Short:   "Decode MetadataAddress and display associate IDs and types",
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
			_, cmdErr := fmt.Fprintf(cmd.OutOrStdout(), toOut)
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
		Long: `Encodes metadata uuids to bech32 address for specific type. 
		Types: scope,session,record,contract-specification,scope-specification`,
		Args: cobra.RangeArgs(3, 4),
		RunE: func(cmd *cobra.Command, args []string) error {
			primaryUUID, err := uuid.Parse(args[2])
			if err != nil {
				return err
			}
			switch addrType := args[1]; addrType {
			case "scope":
				if len(args) != 3 {
					return fmt.Errorf("too many arguments for %s address encoder", addrType)
				}
				scopeAddr := types.ScopeMetadataAddress(primaryUUID)
				fmt.Fprint(cmd.OutOrStdout(), scopeAddr.String())
			case "session":
				if len(args) != 4 {
					return fmt.Errorf("missing secondary uuid for type session")
				}
				secondaryUUID, err := uuid.Parse(args[3])
				if err != nil {
					return err
				}
				sessionAddr := types.SessionMetadataAddress(primaryUUID, secondaryUUID)
				fmt.Fprint(cmd.OutOrStdout(), sessionAddr.String())
			case "record":
				if len(args) != 4 {
					return fmt.Errorf("missing name for type record")
				}
				recordAddr := types.RecordMetadataAddress(primaryUUID, args[3])
				fmt.Fprint(cmd.OutOrStdout(), recordAddr.String())
			case "contract-specification":
				if len(args) != 3 {
					return fmt.Errorf("too many arguments for %s address encoder", addrType)
				}
				contractSpecAddr := types.ContractSpecMetadataAddress(primaryUUID)
				fmt.Fprint(cmd.OutOrStdout(), contractSpecAddr.String())
			case "scope-specification":
				if len(args) != 3 {
					return fmt.Errorf("too many arguments for %s address encoder", addrType)
				}
				scopeSpecAddr := types.ScopeSpecMetadataAddress(primaryUUID)
				fmt.Fprint(cmd.OutOrStdout(), scopeSpecAddr.String())
			default:
				return fmt.Errorf("unknown type: %s, Supported types: scope, session, record, contract-specification, scope-specification", addrType)
			}
			return nil
		},
	}
	return cmd
}
