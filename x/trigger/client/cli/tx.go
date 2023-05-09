package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/provenance-io/provenance/x/trigger/types"
)

func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"t"},
		Short:                      "Transaction commands for the trigger module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(GetCmdAddTrigger())

	return txCmd
}

func GetCmdAddTrigger() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create [event.json] [msg.json]",
		Args:    cobra.ExactArgs(2),
		Aliases: []string{"c"},
		Short:   "Creates a new trigger",
		Long:    strings.TrimSpace(`Creates a new trigger.  This will delay the execution of the provided message until the event has occurred`),
		Example: fmt.Sprintf(`$ %[1]s tx trigger create message.json`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			callerAddr := clientCtx.GetFromAddress()
			if err != nil {
				return fmt.Errorf("invalid argument : %s", args[0])
			}

			event, err := parseEvent(clientCtx.Codec, args[0])
			if err != nil {
				return fmt.Errorf("unable to parse file : %s", err)
			}

			msgs, err := parseTransactions(clientCtx.Codec, args[1])
			if err != nil {
				return fmt.Errorf("unable to parse file : %s", err)
			}
			if len(msgs) == 0 {
				return fmt.Errorf("no actions added to trigger")
			}

			msg := types.NewCreateTriggerRequest(
				callerAddr.String(),
				event,
				msgs,
			)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

type Action struct {
	// Msgs defines a sdk.Msg proto-JSON-encoded as Any.
	Action json.RawMessage `json:"message,omitempty"`
}

// parseSubmitProposal reads and parses the proposal.
func parseTransactions(cdc codec.Codec, path string) ([]sdk.Msg, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var action Action
	err = json.Unmarshal(contents, &action)
	if err != nil {
		return nil, err
	}

	var msg sdk.Msg
	err = cdc.UnmarshalInterfaceJSON(action.Action, &msg)
	if err != nil {
		return nil, err
	}

	return []sdk.Msg{msg}, nil
}

type TransactionEvent struct {
	// Msgs defines an array of sdk.Msgs proto-JSON-encoded as Anys.
	Attributes []Attribute `json:"attributes,omitempty"`
	Name       string      `json:"name"`
}

type Attribute struct {
	// Msgs defines an array of sdk.Msgs proto-JSON-encoded as Anys.
	Name  string `json:"name"`
	Value string `json:"value"`
}

// parseSubmitProposal reads and parses the proposal.
func parseEvent(cdc codec.Codec, path string) (*types.TransactionEvent, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var event TransactionEvent
	err = json.Unmarshal(contents, &event)
	if err != nil {
		return nil, err
	}

	newEvent := types.TransactionEvent{
		Name: event.Name,
	}
	for _, attr := range event.Attributes {
		newEvent.Attributes = append(newEvent.Attributes, types.Attribute{
			Name:  attr.Name,
			Value: attr.Value,
		})
	}

	return &newEvent, nil
}
