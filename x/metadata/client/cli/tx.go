package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/provenance-io/provenance/x/metadata/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	uuid "github.com/google/uuid"
)

const (
	FlagSigners = "signers"
)

// NewTxCmd is the top-level command for attribute CLI transactions.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"m"},
		Short:                      "Transaction commands for the metadata module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		AddScopeCmd(),
		RemoveScopeCmd(),

		AddRecordCmd(),
		RemoveRecordCmd(),

		AddOsLocatorCmd(),
		RemoveOsLocatorCmd(),
		ModifyOsLocatorCmd(),

		AddContractSpecificationCmd(),
		RemoveContractSpecificationCmd(),

		AddRecordSpecificationCmd(),
		RemoveRecordSpecificationCmd(),
	)

	return txCmd
}

// AddScopeCmd creates a command for adding a metadata scope.
func AddScopeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-scope [scope-uuid] [spec-id] [owner-addresses] [data-access] [value-owner-address]",
		Short: "Add/Update a metadata scope to the provenance blockchain",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			scopeUUID, err := uuid.Parse(args[0])
			if err != nil {
				fmt.Printf("Invalid uuid for scope uuid: %s", args[0])
				return err
			}
			specUUID, err := uuid.Parse(args[1])
			if err != nil {
				fmt.Printf("Invalid uuid for specification uuid: %s", args[0])
				return err
			}

			specID := types.ScopeSpecMetadataAddress(specUUID)

			ownerAddresses := strings.Split(args[2], ",")
			owners := make([]types.Party, len(ownerAddresses))
			for i, ownerAddr := range ownerAddresses {
				owners[i] = types.Party{Address: ownerAddr, Role: types.PartyType_PARTY_TYPE_OWNER}
			}
			dataAccess := strings.Split(args[3], ",")
			valueOwnerAddress := args[4]

			signers, err := parseSigners(cmd, &clientCtx)
			if err != nil {
				return err
			}

			scope := *types.NewScope(
				types.ScopeMetadataAddress(scopeUUID),
				specID,
				owners,
				dataAccess,
				valueOwnerAddress)

			if err := scope.ValidateBasic(); err != nil {
				fmt.Printf("Failed to validate scope %s : %v", scope.String(), err)
				return err
			}

			msg := types.NewMsgAddScopeRequest(scope, signers)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	addSignerFlagCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// RemoveScopeCmd creates a command for removing a scope.
func RemoveScopeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-scope [scope-address] [signers]",
		Short: "Remove a metadata scope to the provenance blockchain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			scopeUUID, err := uuid.Parse(args[0])
			if err != nil {
				fmt.Printf("Invalid uuid for scope id: %s", args[0])
				return err
			}

			scopeMetaAddress := types.ScopeMetadataAddress(scopeUUID)

			signers, err := parseSigners(cmd, &clientCtx)
			if err != nil {
				return err
			}

			deleteScope := *types.NewMsgDeleteScopeRequest(scopeMetaAddress, signers)
			if err := deleteScope.ValidateBasic(); err != nil {
				fmt.Printf("Failed to validate remove scope %s : %v", deleteScope.String(), err)
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &deleteScope)
		},
	}

	addSignerFlagCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// RemoveMetadataScopeCmd creates a command for removing a scope.
func AddOsLocatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-locator [owner] [uri]",
		Short: "Add a uri to an owner address on the provenance blockchain",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			if _, errAddr := sdk.AccAddressFromBech32(args[0]); errAddr != nil {
				fmt.Printf("failed to add locator for a given owner address, invalid address: %s\n", args[0])
				return fmt.Errorf("invalid address: %w", errAddr)
			}

			objectStoreLocator := types.ObjectStoreLocator{
				LocatorUri: args[1], Owner: args[0],
			}

			addOSLocator := *types.NewMsgBindOSLocatorRequest(objectStoreLocator)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &addOSLocator)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// RemoveOsLocatorCmd creates a command for removing a os locator
func RemoveOsLocatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-locator [owner] [uri]",
		Short: "Remove an os locator already associated owner address on the provenance blockchain",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			if _, errAddr := sdk.AccAddressFromBech32(args[0]); errAddr != nil {
				fmt.Printf("failed to remove locator for a given owner address, invalid address: %s\n", args[0])
				return fmt.Errorf("invalid address: %w", errAddr)
			}

			objectStoreLocator := types.ObjectStoreLocator{
				LocatorUri: args[1], Owner: args[0],
			}

			deleteOSLocator := *types.NewMsgDeleteOSLocatorRequest(objectStoreLocator)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &deleteOSLocator)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// ModifyOsLocatorCmd creates a command for modifying os locator
func ModifyOsLocatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "modify-locator [owner] [uri]",
		Short: "Modify a uri already associated owner address on the provenance blockchain",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			if _, errAddr := sdk.AccAddressFromBech32(args[0]); errAddr != nil {
				fmt.Printf("failed to add locator for a given owner address, invalid address: %s\n", args[0])
				return fmt.Errorf("invalid address: %w", errAddr)
			}
			if err != nil {
				fmt.Printf("Invalid uuid for scope id: %s", args[0])
				return err
			}

			objectStoreLocator := types.ObjectStoreLocator{
				LocatorUri: args[1], Owner: args[0],
			}

			modifyOSLocator := *types.NewMsgModifyOSLocatorRequest(objectStoreLocator)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &modifyOSLocator)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// AddContractSpecificationCmd creates a command to add/update contract specifications
func AddContractSpecificationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-contract-specification [contractspec-id] [owners] [parties-involved] [source-value] [classname] [description-name] [description] [website-url] [icon-url]",
		Short: "Add/Update metadata contract specification on the provenance blockchain",
		Long: `Add/Update metadata contract specification on the provenance blockchain
[contractspec-id]   - contract specification metaaddress
[owners]            - comma delimited list of bech32 owner addresses
[parties-involved]  - comma delimited list of party types.  Accepted values: originator,servicer,investor,custodian,owner,affiliate,omnibus,provenance
[source-value]      - source identifier of type hash or resourceid
[classname]         - name of contract specification
[description-name]* - description name identifier 
[description]*      - description text
[website-url]*      - address of website
[icon-url]*         - address to a image to be used as an icon
* - are optional values		
`,
		Args: cobra.RangeArgs(5, 9),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			specificationID, err := types.MetadataAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			partiesInvolved := parsePartyTypes(args[2])
			description := parseDescription(args[5:])
			contractSpecification := types.ContractSpecification{SpecificationId: specificationID,
				Description:     description,
				OwnerAddresses:  strings.Split(args[1], ","),
				PartiesInvolved: partiesInvolved,
				ClassName:       args[4],
			}
			sourceValue := args[3]
			var recordID types.MetadataAddress
			recordID, err = types.MetadataAddressFromBech32(sourceValue)
			if err != nil {
				contractSpecification.Source = &types.ContractSpecification_Hash{
					Hash: sourceValue,
				}
			} else {
				contractSpecification.Source = &types.ContractSpecification_ResourceId{
					ResourceId: recordID,
				}
			}

			signers, err := parseSigners(cmd, &clientCtx)
			if err != nil {
				return err
			}

			msg := types.NewMsgAddContractSpecificationRequest(contractSpecification, signers)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	addSignerFlagCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// AddRecordCmd creates a command to add/update records
func AddRecordCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-record [scope-id] record-spec-id] [name] [process] [inputs] [outputs] [parties-involved] [session-id]",
		Short: "Add/Update metadata record to the provenance blockchain",
		Long: fmt.Sprintf(`Add/Update metadata record to the provenance blockchain.
[scope-id]         - scope metaaddress for the record
[record-spec-id]   - associated record specification metaaddress
[name]             - record name
[process]          - comma delimited structure of process name, id (hash or bech32 address), and method: Example: processname,hashvalue,method
[inputs]           - semicolon delimited list of input structures.  Example: name,soure-value(hash or metaaddress),typename,status(proposed,record);...
[outputs]          - semicolon delimited list of outputs structures. Example: hash-value,status(pass,skip,fail);...
[parties-involved] - comma delimited list of party types.  Accepted values: originator,servicer,investor,custodian,owner,affiliate,omnibus,provenance
[session-id]       - optional - session metaaddress, if not provided it will be created

Example: 
$ %s tx metadata add-record recspec1qh... recordname myprocessname,myhashvalue input1name,input1hashvalue,input1typename,proposed;... output1hash,pass;... session123...
`, version.AppName),
		Args: cobra.RangeArgs(6, 7),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			recordSpecID, err := types.MetadataAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			process, err := parseProcess(args[2])
			if err != nil {
				return err
			}
			inputs, err := parseRecordInputs(args[3])
			if err != nil {
				return err
			}
			outputs, err := parseRecordOutputs(args[4])
			if err != nil {
				return err
			}
			parties := parsePartyTypes(args[5])

			signers, err := parseSigners(cmd, &clientCtx)
			if err != nil {
				return err
			}

			sessionID := types.MetadataAddress{}
			if len(args) == 7 {
				sessionID, err = types.MetadataAddressFromBech32(args[0])
				if err != nil {
					return err
				}
			}

			record := types.Record{
				SpecificationId: recordSpecID,
				Process:         *process,
				Inputs:          inputs,
				Outputs:         outputs,
				SessionId:       sessionID,
			}

			msg := *types.NewMsgAddRecordRequest(sessionID, record, parties, signers)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	addSignerFlagCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// parseProcess parses a comma seperated structure of name, processid(hash or metaaddress), method.  name,hashvalue,methodnam;...
func parseProcess(cliDelimitedValue string) (*types.Process, error) {
	values := strings.Split(cliDelimitedValue, ",")
	if len(values) != 3 {
		return nil, fmt.Errorf("invalid number of values for process: %v", len(values))
	}

	process := types.Process{
		Name:   values[0],
		Method: values[2],
	}
	processID, err := types.MetadataAddressFromBech32(values[1])
	if err != nil {
		process.ProcessId = &types.Process_Address{Address: string(processID)}
	} else {
		process.ProcessId = &types.Process_Hash{Hash: values[0]}
	}
	return &process, nil

}

// parseRecordInputs parses a list of semicolon, comma delimited input structure name,soure-value(hash or metaaddress),typename,status(proposed,record);...
func parseRecordInputs(cliDelimitedValue string) ([]types.RecordInput, error) {
	delimitedInputs := strings.Split(cliDelimitedValue, ";")
	inputs := make([]types.RecordInput, len(delimitedInputs))
	for i, delimitedInput := range delimitedInputs {
		values := strings.Split(delimitedInput, ",")
		if len(values) != 4 {
			return nil, fmt.Errorf("invalid number of values for record input: %v", len(values))
		}
		inputs[i] = types.RecordInput{
			Name:     values[0],
			TypeName: values[2],
			Status:   types.RecordInputStatus(types.RecordInputStatus_value[fmt.Sprintf("RECORD_INPUT_STATUS_%s", strings.ToUpper(values[3]))]),
		}
		sourceValue := values[1]
		recordID, err := types.MetadataAddressFromBech32(sourceValue)
		if err != nil {
			inputs[i].Source = &types.RecordInput_Hash{Hash: sourceValue}
		} else {
			inputs[i].Source = &types.RecordInput_RecordId{RecordId: recordID}
		}
	}
	return inputs, nil
}

// parseRecordOutputs parses a list of semicolon, comma delimited output structures hash,status(pass,skip,fail);...
func parseRecordOutputs(cliDelimitedValue string) ([]types.RecordOutput, error) {
	delimitedOutputs := strings.Split(cliDelimitedValue, ";")
	outputs := make([]types.RecordOutput, len(delimitedOutputs))
	for i, delimitedOutput := range delimitedOutputs {
		values := strings.Split(delimitedOutput, ",")
		if len(values) != 2 {
			return nil, fmt.Errorf("invalid number of values for record output: %v", len(values))
		}
		outputs[i] = types.RecordOutput{
			Hash:   values[0],
			Status: types.ResultStatus(types.ResultStatus_value[fmt.Sprintf("ResultStatus_RESULT_STATUS_%s", strings.ToUpper(values[1]))]),
		}
	}
	return outputs, nil
}

// AddRecordSpecificationCmd creates a command to add/update record specifications
func AddRecordSpecificationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-record-specification [specification-id] [name] [input-specifications] [type-name] [result-types] [responsible-parties]",
		Short: "Add/Update metadata record specification to the provenance blockchain",
		Long: fmt.Sprintf(`Add/Update metadata record specification to the provenance blockchain.
[specification-id]     - record specification metaaddress
[name]                 - record name
[input-specifications] - semi-colon delimited list of input specifications <name>,<type-name>,<source-value>
[type-name]            - contract specification type name
[result-types]         - result definition type.  Accepted values: proposed,record,record_list
[responsible-parties]  - comma delimited list of party types.  Accepted values: originator,servicer,investor,custodian,owner,affiliate,omnibus,provenance

Example: 
$ %s tx metadata add-record-specification recspec1qh... recordname inputname1,typename1,hashvalue;inputename2,typename2,<recordmetaaddress> record_list owner,originator

`, version.AppName),
		Args: cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			specificationID, err := types.MetadataAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			recordName := args[1]

			inputs, err := parseInputSpecification(args[2])
			if err != nil {
				return err
			}

			resultType := types.DefinitionType(types.DefinitionType_value[fmt.Sprintf("DEFINITION_TYPE_%s", strings.ToUpper(args[4]))])
			partyTypes := parsePartyTypes(args[5])
			signers, err := parseSigners(cmd, &clientCtx)
			if err != nil {
				return err
			}

			recordSpecification := types.RecordSpecification{
				SpecificationId:    specificationID,
				Name:               recordName,
				Inputs:             inputs,
				TypeName:           args[3],
				ResultType:         resultType,
				ResponsibleParties: partyTypes,
			}

			msg := *types.NewMsgAddRecordSpecificationRequest(recordSpecification, signers)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	addSignerFlagCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// parseInputSpecification converts cli delimited argument and converts it to InputSpecifications
func parseInputSpecification(cliDelimitedValue string) ([]*types.InputSpecification, error) {
	delimitedInputs := strings.Split(cliDelimitedValue, ";")
	inputs := make([]*types.InputSpecification, len(delimitedInputs))
	for i, delimitedInput := range delimitedInputs {
		values := strings.Split(delimitedInput, ",")
		if len(values) != 3 {
			return nil, fmt.Errorf("invalid number of values for input specification: %v", len(values))
		}
		inputs[i] = &types.InputSpecification{
			Name:     values[0],
			TypeName: values[1],
		}
		sourceValue := values[2]
		recordID, err := types.MetadataAddressFromBech32(sourceValue)
		if err != nil {
			inputs[i].Source = &types.InputSpecification_Hash{Hash: sourceValue}
		} else {
			inputs[i].Source = &types.InputSpecification_RecordId{RecordId: recordID}
		}
	}
	return inputs, nil
}

func addSignerFlagCmd(cmd *cobra.Command) {
	cmd.Flags().String(FlagSigners, "", "comma delimited list of bech32 addresses")
}

// parseSigners checks signers flag for signers, else uses the from address
func parseSigners(cmd *cobra.Command, client *client.Context) ([]string, error) {
	flagSet := cmd.Flags()
	if flagSet.Changed(FlagSigners) {
		signerList, _ := flagSet.GetString(FlagSigners)
		signers := strings.Split(signerList, ",")
		for _, signer := range signers {
			_, err := sdk.AccAddressFromBech32(signer)
			if err != nil {
				fmt.Printf("signer address must be a Bech32 string: %v", err)
				return nil, err
			}
		}
		return signers, nil
	}
	return []string{client.GetFromAddress().String()}, nil
}

func parsePartyTypes(delimitedPartyTypes string) []types.PartyType {
	parties := strings.Split(delimitedPartyTypes, ",")
	partyTypes := make([]types.PartyType, len(parties))
	for i, party := range parties {
		partyValue := types.PartyType_value[fmt.Sprintf("PARTY_TYPE_%s", strings.ToUpper(party))]
		partyTypes[i] = types.PartyType(partyValue)
	}
	return partyTypes
}

// parseDescription hydrates Description from a sorted array name,description,website,icon-url
func parseDescription(cliArgs []string) *types.Description {
	if len(cliArgs) == 0 {
		return nil
	}

	description := types.Description{}
	if len(cliArgs) >= 1 {
		description.Name = cliArgs[0]
	}
	if len(cliArgs) >= 2 {
		description.Description = cliArgs[1]
	}
	if len(cliArgs) >= 3 {
		description.WebsiteUrl = cliArgs[2]
	}
	if len(cliArgs) >= 4 {
		description.IconUrl = cliArgs[3]
	}
	return &description
}

// RemoveContractSpecificationCmd creates a command to remove a contract specification
func RemoveContractSpecificationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-contract-specification [specification-id]",
		Short: "Removes a contract specification on the provenance blockchain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var specificationID types.MetadataAddress
			specificationID, err = types.MetadataAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			signers, err := parseSigners(cmd, &clientCtx)
			if err != nil {
				return err
			}

			msg := types.NewMsgDeleteContractSpecificationRequest(specificationID, signers)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	addSignerFlagCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// RemoveRecordCmd creates a command to remove a contract specification
func RemoveRecordCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-record [record-id]",
		Short: "Remove record specification from the provenance blockchain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var recordId types.MetadataAddress
			recordId, err = types.MetadataAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			signers, err := parseSigners(cmd, &clientCtx)
			if err != nil {
				return err
			}
			msg := *types.NewMsgDeleteRecordRequest(recordId, signers)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	addSignerFlagCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// RemoveRecordSpecificationCmd creates  a command to remove a record specification
func RemoveRecordSpecificationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-record-specification [specification-id]",
		Short: "Remove record specification from the provenance blockchain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var specificationID types.MetadataAddress
			specificationID, err = types.MetadataAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			if !specificationID.IsRecordSpecificationAddress() {
				return fmt.Errorf("invalid contract specification id: %s", args[0])
			}
			signers, err := parseSigners(cmd, &clientCtx)
			if err != nil {
				return err
			}
			msg := *types.NewMsgDeleteRecordSpecificationRequest(specificationID, signers)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	addSignerFlagCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
