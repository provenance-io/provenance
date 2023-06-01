package cli

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	attrcli "github.com/provenance-io/provenance/x/attribute/client/cli"

	"github.com/provenance-io/provenance/x/metadata/types"
)

const (
	FlagSigners            = "signers"
	FlagRequirePartyRollup = "require-party-rollup"
	AddSwitch              = "add"
	RemoveSwitch           = "remove"
)

// NewTxCmd is the top-level command for Metadata CLI transactions.
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
		WriteScopeCmd(),
		RemoveScopeCmd(),
		AddRemoveScopeDataAccessCmd(),
		AddRemoveScopeOwnersCmd(),
		UpdateValueOwnersCmd(),
		MigrateValueOwnerCmd(),

		BindOsLocatorCmd(),
		RemoveOsLocatorCmd(),
		ModifyOsLocatorCmd(),

		WriteScopeSpecificationCmd(),
		RemoveScopeSpecificationCmd(),

		WriteContractSpecificationCmd(),
		RemoveContractSpecificationCmd(),

		AddContractSpecToScopeSpecCmd(),
		RemoveContractSpecFromScopeSpecCmd(),

		WriteRecordSpecificationCmd(),
		RemoveRecordSpecificationCmd(),

		WriteSessionCmd(),

		WriteRecordCmd(),
		RemoveRecordCmd(),

		SetAccountDataCmd(),
	)

	return txCmd
}

// WriteScopeCmd creates a command for adding or updating a metadata scope.
func WriteScopeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "write-scope [scope-id] [spec-id] [owners] [data-access] [value-owner-address] [flags]",
		Short: "Add/Update a metadata scope to the provenance blockchain",
		Long: `Add/Update a metadata scope to the provenance blockchain

[scope-id] is a scope metadata address.
[spec-id] is a scope specification metadata address.
[owners] is a semicolon delimited list of parties.
  Each party must have one of the following formats:
    "[address]" or "[address],[role]" or "[address],[role],opt"
    Default role is "owner".
    "opt" indicates optional = true. Default optional is false.
[data-access] - a comma delimited list of addresses.
[value-owner-address] - an address.
`,
		Example: fmt.Sprintf(`$ %[1]s tx metadata write-scope scope1qzhpuff00wpy2yuf7xr0rp8aucqstsk0cn scopespec1qjpreurq8n7ylc4y5zw6gn255lkqle56sv pb1sh49f6ze3vn7cdl2amh2gnc70z5mten3dpvr42 pb1sh49f6ze3vn7cdl2amh2gnc70z5mten3dpvr42 pb1sh49f6ze3vn7cdl2amh2gnc70z5mten3dpvr42`, version.AppName),
		Args:    cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var scopeID types.MetadataAddress
			scopeID, err = types.MetadataAddressFromBech32(args[0])
			if err != nil {
				return fmt.Errorf("invalid scope id: %w", err)
			}

			var specID types.MetadataAddress
			specID, err = types.MetadataAddressFromBech32(args[1])
			if err != nil {
				return fmt.Errorf("invalid spec id: %w", err)
			}

			owners, err := parseParties(args[2])
			if err != nil {
				return fmt.Errorf("invalid owners: %w", err)
			}
			dataAccess := strings.Split(args[3], ",")
			valueOwnerAddress := args[4]

			signers, err := parseSigners(cmd, &clientCtx)
			if err != nil {
				return err
			}

			requirePartyRollup, err := cmd.Flags().GetBool(FlagRequirePartyRollup)
			if err != nil {
				return err
			}

			scope := *types.NewScope(
				scopeID,
				specID,
				owners,
				dataAccess,
				valueOwnerAddress,
				requirePartyRollup,
			)

			msg := types.NewMsgWriteScopeRequest(scope, signers)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Bool(FlagRequirePartyRollup, false, "Indicates party rollup is required in this scope")
	addSignersFlagToCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// RemoveScopeCmd creates a command for removing a scope.
func RemoveScopeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove-scope [scope-id]",
		Short:   "Remove a metadata scope to the provenance blockchain",
		Example: fmt.Sprintf(`$ %[1]s tx metadata remove-scope scope1qzhpuff00wpy2yuf7xr0rp8aucqstsk0cn`, version.AppName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var scopeID types.MetadataAddress
			scopeID, err = types.MetadataAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			signers, err := parseSigners(cmd, &clientCtx)
			if err != nil {
				return err
			}

			msg := *types.NewMsgDeleteScopeRequest(scopeID, signers)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	addSignersFlagToCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// AddRemoveScopeDataAccessCmd creates a command for either adding or removing an address from a scope's data access list.
func AddRemoveScopeDataAccessCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scope-data-access {add|remove} [scope-id] [data-access]",
		Short: "Add or remove a metadata scope data access on to the provenance blockchain",
		Example: fmt.Sprintf(`$ %[1]s tx metadata scope-data-access add scope1qzhpuff00wpy2yuf7xr0rp8aucqstsk0cn pb1sh49f6ze3vn7cdl2amh2gnc70z5mten3dpvr42
$ %[1]s tx metadata scope-data-access remove scope1qzhpuff00wpy2yuf7xr0rp8aucqstsk0cn pb1sh49f6ze3vn7cdl2amh2gnc70z5mten3dpvr42`, version.AppName),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			removeOrAdd := strings.ToLower(args[0])
			if removeOrAdd != RemoveSwitch && removeOrAdd != AddSwitch {
				return fmt.Errorf("incorrect command %s : required remove or update", removeOrAdd)
			}

			var scopeID types.MetadataAddress
			scopeID, err = types.MetadataAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			if !scopeID.IsScopeAddress() {
				return fmt.Errorf("meta address is not a scope: %s", scopeID.String())
			}

			signers, err := parseSigners(cmd, &clientCtx)
			if err != nil {
				return err
			}

			dataAccess := strings.Split(args[2], ",")
			var msg sdk.Msg
			if removeOrAdd == AddSwitch {
				msg = types.NewMsgAddScopeDataAccessRequest(scopeID, dataAccess, signers)
			} else {
				msg = types.NewMsgDeleteScopeDataAccessRequest(scopeID, dataAccess, signers)
			}
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	addSignersFlagToCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// AddRemoveScopeOwnersCmd creates a command for either adding or removing scope owners.
func AddRemoveScopeOwnersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scope-owners {add|remove} [scope-id] [owner-addresses]",
		Short: "Add or remove a metadata scope owners on to the provenance blockchain",
		Example: fmt.Sprintf(`$ %[1]s tx metadata scope-owners add scope1qzhpuff00wpy2yuf7xr0rp8aucqstsk0cn pb1sh49f6ze3vn7cdl2amh2gnc70z5mten3dpvr42
$ %[1]s tx metadata scope-owners remove scope1qzhpuff00wpy2yuf7xr0rp8aucqstsk0cn pb1sh49f6ze3vn7cdl2amh2gnc70z5mten3dpvr42`, version.AppName),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			removeOrAdd := strings.ToLower(args[0])
			if removeOrAdd != RemoveSwitch && removeOrAdd != AddSwitch {
				return fmt.Errorf("incorrect command %s : required remove or update", removeOrAdd)
			}

			var scopeID types.MetadataAddress
			scopeID, err = types.MetadataAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			if !scopeID.IsScopeAddress() {
				return fmt.Errorf("meta address is not a scope: %s", scopeID.String())
			}

			signers, err := parseSigners(cmd, &clientCtx)
			if err != nil {
				return err
			}

			ownerAddresses := strings.Split(args[2], ",")

			var msg sdk.Msg
			if removeOrAdd == AddSwitch {
				owners := make([]types.Party, len(ownerAddresses))
				for i, ownerAddr := range ownerAddresses {
					owners[i] = types.Party{Address: ownerAddr, Role: types.PartyType_PARTY_TYPE_OWNER}
				}
				msg = types.NewMsgAddScopeOwnerRequest(scopeID, owners, signers)
			} else {
				msg = types.NewMsgDeleteScopeOwnerRequest(scopeID, ownerAddresses, signers)
			}
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	addSignersFlagToCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// UpdateValueOwnersCmd creates a command for updating the value owner of one or more scopes.
func UpdateValueOwnersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update-value-owners <new value owner> <scope id> [<scope id 2> ...]",
		Aliases: []string{"update-value-owner", "uvo"},
		Short:   "Update the value owner of one or more scopes.",
		Example: fmt.Sprintf(`$ %[1]s tx metadata update-value-owners pb1sh49f6ze3vn7cdl2amh2gnc70z5mten3dpvr42 scope1qzhpuff00wpy2yuf7xr0rp8aucqstsk0cn scope1qqg3uff00wpy2yuf7xr0rp8aucqs902xhw`,
			version.AppName),
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgUpdateValueOwnersRequest{}

			msg.ValueOwnerAddress, err = validateAccAddress(args[0], "new value owner")
			if err != nil {
				return err
			}

			msg.ScopeIds = make([]types.MetadataAddress, len(args[1:]))
			for i, arg := range args[1:] {
				msg.ScopeIds[i], err = types.MetadataAddressFromBech32(arg)
				if err == nil && !msg.ScopeIds[i].IsScopeAddress() {
					err = fmt.Errorf("not a scope identifier")
				}
				if err != nil {
					return fmt.Errorf("invalid scope id %d %q: %w", i+1, arg, err)
				}
			}

			msg.Signers, err = parseSigners(cmd, &clientCtx)
			if err != nil {
				return err
			}

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	addSignersFlagToCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MigrateValueOwnerCmd creates a command for migrating the scopes of one value owner to another.
func MigrateValueOwnerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "migrate-value-owner <existing value owner> <proposed value owner>",
		Aliases: []string{"mvo"},
		Short:   "Migrate the scopes of one value owner to another.",
		Example: fmt.Sprintf(`$ %[1]s tx metadata migrate-value-owner pb1sh49f6ze3vn7cdl2amh2gnc70z5mten3dpvr42 pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk`,
			version.AppName),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgMigrateValueOwnerRequest{}
			msg.Existing, err = validateAccAddress(args[0], "existing value owner")
			if err != nil {
				return err
			}

			msg.Proposed, err = validateAccAddress(args[1], "proposed value owner")
			if err != nil {
				return err
			}

			msg.Signers, err = parseSigners(cmd, &clientCtx)
			if err != nil {
				return err
			}

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	addSignersFlagToCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// BindOsLocatorCmd creates a command for binding an owner to uri in the object store.
func BindOsLocatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bind-locator [owner] [uri]",
		Short:   "Bind a uri to an owner address on the provenance blockchain",
		Example: fmt.Sprintf(`$ %[1]s tx metadata bind-locator pb1sh49f6ze3vn7cdl2amh2gnc70z5mten3dpvr42 "http://foo.com"`, version.AppName),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			if _, errAddr := sdk.AccAddressFromBech32(args[0]); errAddr != nil {
				fmt.Printf("failed to bind locator for a given owner address, invalid address: %s\n", args[0])
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

// RemoveOsLocatorCmd creates a command for removing an object store locator entry.
func RemoveOsLocatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove-locator [owner] [uri]",
		Short:   "Remove an os locator already associated owner address on the provenance blockchain",
		Example: fmt.Sprintf(`$ %[1]s tx metadata remove-locator pb1sh49f6ze3vn7cdl2amh2gnc70z5mten3dpvr42 "http://foo.com"`, version.AppName),
		Args:    cobra.ExactArgs(2),
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

// ModifyOsLocatorCmd creates a command to modify the object store locator uri for an owner.
func ModifyOsLocatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "modify-locator [owner] [uri]",
		Short:   "Modify a uri already associated owner address on the provenance blockchain",
		Example: fmt.Sprintf(`$ %[1]s tx metadata modify-locator pb1sh49f6ze3vn7cdl2amh2gnc70z5mten3dpvr42 "http://foo2.com"`, version.AppName),
		Args:    cobra.ExactArgs(2),
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

// WriteScopeSpecificationCmd creates a command for adding scope specificiation
func WriteScopeSpecificationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "write-scope-specification [specification-id] [owner-addresses] [responsible-parties] [contract-specification-ids] [description-name, optional] [description, optional] [website-url, optional] [icon-url, optional]",
		Short:   "Add/Update metadata scope specification to the provenance blockchain",
		Example: fmt.Sprintf(`$ %[1]s tx metadata write-scope-specification scopespec1qn7jh3jvw4gytq9r5x770e8yj74s9t479r pb1sh49f6ze3vn7cdl2amh2gnc70z5mten3dpvr42 "owner" contractspec1q0mnck3mqh75mg9qvykqq0jxzs2struaa8`, version.AppName),
		Args:    cobra.RangeArgs(4, 8),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			specificationID, err := types.MetadataAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			cSpecIds := strings.Split(args[3], ",")
			contractSpecIds := make([]types.MetadataAddress, len(cSpecIds))
			for i, cid := range cSpecIds {
				contractSpecIds[i], err = types.MetadataAddressFromBech32(cid)
				if err != nil {
					return err
				}
			}

			signers, err := parseSigners(cmd, &clientCtx)
			if err != nil {
				return err
			}

			partyTypes, err := parsePartyTypes(args[2])
			if err != nil {
				return err
			}

			scopeSpec := types.ScopeSpecification{
				SpecificationId: specificationID,
				OwnerAddresses:  strings.Split(args[1], ","),
				Description:     parseDescription(args[4:]),
				PartiesInvolved: partyTypes,
				ContractSpecIds: contractSpecIds,
			}

			msg := types.NewMsgWriteScopeSpecificationRequest(scopeSpec, signers)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	addSignersFlagToCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// WriteContractSpecificationCmd creates a command to add/update contract specifications
func WriteContractSpecificationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "write-contract-specification [contractspec-id] [owners] [parties-involved] [source-value] [classname] [description-name, optional] [description, optional] [website-url, optional] [icon-url, optional]",
		Short: "Add/Update metadata contract specification on the provenance blockchain",
		Long: `Add/Update metadata contract specification on the provenance blockchain
contractspec-id    - contract specification metaaddress
owners             - comma delimited list of bech32 owner addresses
parties-involved   - comma delimited list of party types.  Accepted values: originator,servicer,investor,custodian,owner,affiliate,omnibus,provenance
source-value       - source identifier of type hash or resourceid
classname          - name of contract specification
description-name   - description name identifier (optional)
description        - description text (optional, can only be provided with a description-name)
website-url        - address of website (optional, can only be provided with a description)
icon-url           - address to a image to be used as an icon (optional, can only be provided with an website-url)`,
		Example: fmt.Sprintf(`$ %[1]s tx metadata write-contract-specification contractspec1q0w6ys5g6jm509v2830374aprsrq260w62 pb1sh49f6ze3vn7cdl2amh2gnc70z5mten3dpvr42 "owner" "hashvalue" "myclassname" --from=mykey`, version.AppName),
		Args:    cobra.RangeArgs(5, 9),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			specificationID, err := types.MetadataAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			partiesInvolved, err := parsePartyTypes(args[2])
			if err != nil {
				return err
			}
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

			msg := types.NewMsgWriteContractSpecificationRequest(contractSpecification, signers)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	addSignersFlagToCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// AddContractSpecToScopeSpecCmd creates an add contract spec to scope spec command
func AddContractSpecToScopeSpecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-contract-spec-to-scope-spec [contract-specification-id] [scope-specification-id]",
		Short:   "Add an existing contract specification to a scope specification",
		Example: fmt.Sprintf(`$ %[1]s tx metadata add-contract-spec-to-scope-spec contractspec1q0w6ys5g6jm509v2830374aprsrq260w62 scopespec1qjpreurq8n7ylc4y5zw6gn255lkqle56sv --from=mykey`, version.AppName),
		Aliases: []string{"acstss"},
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			contractSpecID, err := types.MetadataAddressFromBech32(args[0])
			if err != nil || !contractSpecID.IsContractSpecificationAddress() {
				return fmt.Errorf("invalid contract specification id : %s", args[0])
			}
			scopeSpecID, err := types.MetadataAddressFromBech32(args[1])
			if err != nil || !scopeSpecID.IsScopeSpecificationAddress() {
				return fmt.Errorf("invalid scope specification id : %s", args[1])
			}

			signers, err := parseSigners(cmd, &clientCtx)
			if err != nil {
				return err
			}

			msg := types.NewMsgAddContractSpecToScopeSpecRequest(contractSpecID, scopeSpecID, signers)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	addSignersFlagToCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// WriteSessionCmd creates a command for writing a session.
func WriteSessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "write-session {session-id|{scope-id|scope-uuid} session-uuid} [contract-spec-id] [parties] [name] [context, optional]",
		Short: "Add/Update metadata sessioon to the provenance blockchain",
		Long: `Add/Update metadata session to the provenance blockchain.
session-id        - a bech32 address string for this session
scope-id          - a bech32 address string for the scope this session belongs to
scope-uuid        - a UUID string representing the uuid of the scope this session belongs to
session-uuid      - a UUID string representing the uuid for this session
  The above arguments are used to define the session-id. They must be provided in one of these forms:
    session-id
    scope-id session-uuid
    scope-uuid session-uuid
contract-spec-id  - a bech32 address string for the contract specification that applies to this session
parties-involved  - semicolon delimited list of party structures(address,role). Accepted roles: originator,servicer,investor,custodian,owner,affiliate,omnibus,provenance
name              - a name for this session
context           - a base64 encoded string of the bytes that represent the session context (optional)`,
		Example: fmt.Sprintf(`$ %[1]s tx metadata write-session \
91978ba2-5f35-459a-86a7-feca1b0512e0 5803f8bc-6067-4eb5-951f-2121671c2ec0 \
contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn \
pb1sh49f6ze3vn7cdl2amh2gnc70z5mten3dpvr42,owner \
io.prov.contracts.example.HelloWorldContract

$ %[1]s tx metadata write-session \
session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr \
contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn \
pb1sh49f6ze3vn7cdl2amh2gnc70z5mten3dpvr42,owner \
io.prov.contracts.example.HelloWorldContract \
ChFIRUxMTyBQUk9WRU5BTkNFIQ==`, version.AppName),
		Args: cobra.RangeArgs(4, 6),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, ctxErr := client.GetClientTxContext(cmd)
			if ctxErr != nil {
				return ctxErr
			}

			var scopeUUID uuid.UUID
			var sessionID types.MetadataAddress
			var cSpecID types.MetadataAddress
			var parties []types.Party
			var name string
			var context []byte
			var signers []string
			var err error

			argsLeft := args

			// Parse: {session-id|{scope-id|scope-uuid} session-uuid}
			arg0AsID, asIDErr := types.MetadataAddressFromBech32(argsLeft[0])
			if asIDErr == nil {
				switch {
				case arg0AsID.IsSessionAddress():
					sessionID = arg0AsID
				case arg0AsID.IsScopeAddress():
					scopeUUID, _ = arg0AsID.ScopeUUID()
				default:
					return fmt.Errorf("invalid address type in argument [%s]", arg0AsID)
				}
			} else {
				arg0AsUUID, asUUIDerr := uuid.Parse(argsLeft[0])
				if asUUIDerr != nil {
					return fmt.Errorf("argument [%s] is neither a bech32 address (%v) nor UUID (%v)", argsLeft[0], asIDErr, asUUIDerr) //nolint:errorlint // Can't wrap two errors at once.
				}
				scopeUUID = arg0AsUUID
			}
			argsLeft = argsLeft[1:]
			if len(sessionID) == 0 {
				sUUID, sUUIDErr := uuid.Parse(argsLeft[0])
				if sUUIDErr != nil {
					return fmt.Errorf("invalid session uuid as argument [%s]: %w", argsLeft[0], sUUIDErr)
				}
				sessionID = types.SessionMetadataAddress(scopeUUID, sUUID)
				argsLeft = argsLeft[1:]
			}

			if len(argsLeft) < 3 {
				return fmt.Errorf("not enough arguments (expected >= %d, found %d)", len(args)-len(argsLeft)+3, len(args))
			}

			// arguments left: {contract-specification-id} {parties} {name} and possibly context stuff.
			cSpecID, err = types.MetadataAddressFromBech32(argsLeft[0])
			if err != nil {
				return fmt.Errorf("invalid contract specification id [%s]: %w", argsLeft[0], err)
			}
			parties, err = parseParties(argsLeft[1])
			if err != nil {
				return err
			}
			name = argsLeft[2]
			argsLeft = argsLeft[3:]

			// Handle the optional context stuff.
			if len(argsLeft) > 0 {
				context, err = base64.StdEncoding.DecodeString(argsLeft[0])
				if err != nil {
					return fmt.Errorf("invalid context: %w", err)
				}
				argsLeft = argsLeft[1:]
			}

			// Make sure there aren't any leftover/unused arguments
			if len(argsLeft) > 0 {
				return fmt.Errorf("too many arguments (expected <= %d, found %d)", len(args)-len(argsLeft), len(args))
			}

			signers, err = parseSigners(cmd, &clientCtx)
			if err != nil {
				return err
			}

			session := types.Session{
				SessionId:       sessionID,
				SpecificationId: cSpecID,
				Parties:         parties,
				Name:            name,
				Context:         context,
			}
			writeSessionMsg := types.NewMsgWriteSessionRequest(session, signers)
			err = writeSessionMsg.ValidateBasic()
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), writeSessionMsg)
		},
	}

	addSignersFlagToCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// WriteRecordCmd creates a command to add/update records
func WriteRecordCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "write-record [scope-id] [record-spec-id] [name] [process] [inputs] [outputs] [parties-involved] {contract-spec-id|session-id}",
		Short: "Add/Update metadata record to the provenance blockchain",
		Long: `Add/Update metadata record to the provenance blockchain.
scope-id          - scope metaaddress for the record
record-spec-id    - associated record specification metaaddress
name              - record name
process           - comma delimited structure of process name, id (hash or bech32 address), and method: Example: processname,hashvalue,method
inputs            - semicolon delimited list of input structures.  Example: name,soure-value(hash or metaaddress),typename,status(proposed,record);...
outputs           - semicolon delimited list of outputs structures. Example: hash-value,status(pass,skip,fail);...
parties-involved  - semicolon delimited list of party structures(address,role). Accepted roles: originator,servicer,investor,custodian,owner,affiliate,omnibus,provenance
contract-spec-id  - a bech32 address string for a contract specification - If provided, a new session will be created using this contract specification
session-id        - a bech32 address string for the session this record belongs to
  Either a contract-spec-id or a session-id must be provided (but not both).
  If a contract-spec-id is provided, a new session will be created using it as the specification for the session, and the record will be part of that session.
  If a session-id is provided, the record will be part of that session (a new session is NOT created).`,
		Example: fmt.Sprintf(`$ %[1]s tx metadata write-record scope1qp... \
recspec1qh... \
recordname \
myprocessname,myhashvalue \
input1name,input1hashvalue,input1typename,proposed;... \
output1hash,pass;... \
userid,owner;... \
session123...
$ %[1]s tx metadata write-record scope1qp... \
recspec1qh... \
recordname \
myprocessname,myhashvalue \
input1name,input1hashvalue,input1typename,proposed;... \
output1hash,pass;... \
userid,owner;... \
contractspec123... \
contractspec-name
`, version.AppName),
		Args: cobra.ExactArgs(8),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			scopeID, err := types.MetadataAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			recordSpecID, err := types.MetadataAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			name := args[2]

			process, err := parseProcess(args[3])
			if err != nil {
				return err
			}
			inputs, err := parseRecordInputs(args[4])
			if err != nil {
				return err
			}
			outputs, err := parseRecordOutputs(args[5])
			if err != nil {
				return err
			}

			parties, err := parseParties(args[6])
			if err != nil {
				return err
			}

			signers, err := parseSigners(cmd, &clientCtx)
			if err != nil {
				return err
			}

			record := types.Record{
				Name:            name,
				SpecificationId: recordSpecID,
				Process:         process,
				Inputs:          inputs,
				Outputs:         outputs,
			}

			contractOrSessionID, err := types.MetadataAddressFromBech32(args[7])
			if err != nil {
				return err
			}
			var sessionID types.MetadataAddress
			var writeSessionMsg *types.MsgWriteSessionRequest
			switch {
			case contractOrSessionID.IsSessionAddress():
				record.SessionId = contractOrSessionID
			case contractOrSessionID.IsContractSpecificationAddress():
				scopeUUID, _ := scopeID.ScopeUUID()
				sessionID = types.SessionMetadataAddress(scopeUUID, uuid.New())
				record.SessionId = sessionID
				session := types.Session{
					SessionId:       sessionID,
					SpecificationId: contractOrSessionID,
					Parties:         parties,
				}
				writeSessionMsg = types.NewMsgWriteSessionRequest(session, signers)
				err = writeSessionMsg.ValidateBasic()
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("id must be a contract or session id: %s", contractOrSessionID.String())
			}
			msg := *types.NewMsgWriteRecordRequest(record, nil, "", signers, parties)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			if writeSessionMsg != nil {
				return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), writeSessionMsg, &msg)
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	addSignersFlagToCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// WriteRecordSpecificationCmd creates a command for writing a record specification.
func WriteRecordSpecificationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "write-record-specification [specification-id] [name] [input-specifications] [type-name] [result-types] [responsible-parties]",
		Short: "Add/Update metadata record specification to the provenance blockchain",
		Long: `Add/Update metadata record specification to the provenance blockchain.
specification-id      - record specification metaaddress
name                  - record name
input-specifications  - semi-colon delimited list of input specifications <name>,<type-name>,<source-value>
type-name             - contract specification type name
result-types          - result definition type. Accepted values: proposed, record, record_list
responsible-parties   - comma delimited list of party types.  Accepted values: originator,servicer,investor,custodian,owner,affiliate,omnibus,provenance`,
		Example: fmt.Sprintf(`$ %[1]s tx metadata write-record-specification recspec1qh... \
recordname \
inputname1,typename1,hashvalue; \
inputename2,typename2,<recordmetaaddress> \
record_list \
owner,originator`, version.AppName),
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

			inputs, err := parseInputSpecifications(args[2])
			if err != nil {
				return err
			}

			resultType := types.DefinitionType(types.DefinitionType_value[fmt.Sprintf("DEFINITION_TYPE_%s", strings.ToUpper(args[4]))])
			partyTypes, err := parsePartyTypes(args[5])
			if err != nil {
				return err
			}
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

			msg := *types.NewMsgWriteRecordSpecificationRequest(recordSpecification, signers)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	addSignersFlagToCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// RemoveScopeSpecificationCmd creates a command to remove scope specification
func RemoveScopeSpecificationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove-scope-specification [specification-id]",
		Short:   "Remove scope specification from the provenance blockchain",
		Example: fmt.Sprintf(`$ %[1]s tx metadata remove-scope-specification scope1qzhpuff00wpy2yuf7xr0rp8aucqstsk0cn --from=mykey`, version.AppName),
		Args:    cobra.ExactArgs(1),
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

			msg := *types.NewMsgDeleteScopeSpecificationRequest(specificationID, signers)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	addSignersFlagToCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// RemoveContractSpecificationCmd creates a command to remove a contract specification
func RemoveContractSpecificationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove-contract-specification [specification-id]",
		Short:   "Removes a contract specification on the provenance blockchain",
		Example: fmt.Sprintf(`$ %[1]s tx metadata remove-contract-specification scope1qzhpuff00wpy2yuf7xr0rp8aucqstsk0cn --from=mykey`, version.AppName),
		Args:    cobra.ExactArgs(1),
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

	addSignersFlagToCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// RemoveContractSpecFromScopeSpecCmd removes a contract spec from scope spec command
func RemoveContractSpecFromScopeSpecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove-contract-spec-from-scope-spec [contract-specification-id] [scope-specification-id]",
		Short:   "Remove an existing contract specification from a scope specification",
		Example: fmt.Sprintf(`$ %[1]s tx metadata remove-contract-spec-from-scope-spec pb1sh49f6ze3vn7cdl2amh2gnc70z5mten3dpvr42 contractspec1qvvwn8p3x8gy2cd6d09e4whxmlhs6af72k --from=mykey`, version.AppName),
		Aliases: []string{"rcsfss"},
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			contractSpecID, err := types.MetadataAddressFromBech32(args[0])
			if err != nil || !contractSpecID.IsContractSpecificationAddress() {
				return fmt.Errorf("invalid contract specification id : %s", args[0])
			}
			scopeSpecID, err := types.MetadataAddressFromBech32(args[1])
			if err != nil || !scopeSpecID.IsScopeSpecificationAddress() {
				return fmt.Errorf("invalid scope specification id : %s", args[1])
			}

			signers, err := parseSigners(cmd, &clientCtx)
			if err != nil {
				return err
			}

			msg := types.NewMsgDeleteContractSpecFromScopeSpecRequest(contractSpecID, scopeSpecID, signers)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	addSignersFlagToCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// RemoveRecordCmd creates a command to remove a contract specification
func RemoveRecordCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove-record [record-id]",
		Short:   "Remove record specification from the provenance blockchain",
		Example: fmt.Sprintf(`$ %[1]s tx metadata remove-record record1qtjqgzrza7h5w8a4amnk9ru9s7236qz42yxp5uejah5tje7c6l0pwue0yn3 --from=mykey`, version.AppName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var recordID types.MetadataAddress
			recordID, err = types.MetadataAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			signers, err := parseSigners(cmd, &clientCtx)
			if err != nil {
				return err
			}
			msg := *types.NewMsgDeleteRecordRequest(recordID, signers)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	addSignersFlagToCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// RemoveRecordSpecificationCmd creates  a command to remove a record specification
func RemoveRecordSpecificationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove-record-specification [specification-id]",
		Short:   "Remove record specification from the provenance blockchain",
		Example: fmt.Sprintf(`$ %[1]s tx metadata remove-record-specification recspec1q4wuhel8td05784pwx6gjqcpz8r0rtq2nzhkhq59fkgty06kz0y5smvfv5p --from=mykey`, version.AppName),
		Args:    cobra.ExactArgs(1),
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

	addSignersFlagToCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// SetAccountDataCmd creates a command for setting account data for a metadata address.
func SetAccountDataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "account-data <address> " + attrcli.AccountDataFlagsUse,
		Aliases: []string{"accountdata", "ad"},
		Short:   "Set a metadata address' account data to either the value provided or the contents of the file provided",
		Example: fmt.Sprintf(`$ %[1]s tx metadata account-data %[2]s --%[3]s "This is some account data."
$ %[1]s tx metadata account-data %[2]s --%[4]s account-data.json
$ %[1]s tx metadata account-data %[2]s --%[5]s
`,
			version.AppName, "scope1qzhpuff00wpy2yuf7xr0rp8aucqstsk0cn", attrcli.FlagValue, attrcli.FlagFile, attrcli.FlagDelete),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgSetAccountDataRequest{}

			msg.MetadataAddr, err = types.MetadataAddressFromBech32(args[0])
			if err != nil {
				return fmt.Errorf("invalid metadata address %q: %w", args[0], err)
			}

			msg.Value, err = attrcli.ReadAccountDataFlags(cmd.Flags())
			if err != nil {
				return err
			}

			msg.Signers, err = parseSigners(cmd, &clientCtx)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	attrcli.AddAccountDataFlagsToCmd(cmd)
	addSignersFlagToCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// addSignersFlagToCmd adds the standard --signers flag to a command.
// See also: parseSigners.
func addSignersFlagToCmd(cmd *cobra.Command) {
	cmd.Flags().String(FlagSigners, "", "comma delimited list of bech32 addresses")
}

// parseSigners checks signers flag for signers, else uses the from address
// See also: addSignersFlagToCmd
func parseSigners(cmd *cobra.Command, client *client.Context) ([]string, error) {
	flagSet := cmd.Flags()
	if flagSet.Changed(FlagSigners) {
		signerList, _ := flagSet.GetString(FlagSigners)
		if len(signerList) > 0 {
			signers := strings.Split(signerList, ",")
			for _, signer := range signers {
				_, err := sdk.AccAddressFromBech32(signer)
				if err != nil {
					return nil, fmt.Errorf("invalid signer address %q: %w", signer, err)
				}
			}
			return signers, nil
		}
	}
	return []string{client.GetFromAddress().String()}, nil
}

// validateAccAddress makes sure the provided addr is a valid bech32.
// If not, an error is returned indicating the argName field.
// If it's valid, it's returned as the first arg.
func validateAccAddress(addr, argName string) (string, error) {
	_, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return "", fmt.Errorf("invalid %s %q: %w", argName, addr, err)
	}
	return addr, nil
}
