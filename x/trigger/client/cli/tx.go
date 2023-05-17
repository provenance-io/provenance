package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

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

	txCmd.AddCommand(GetCmdAddTransactionTrigger(), GetCmdAddBlockHeightTrigger(), GetCmdAddBlockTimeTrigger(), GetCmdDestroyTrigger())

	return txCmd
}

func GetCmdAddTransactionTrigger() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-tx-trigger [event.json] [msg.json]",
		Args:    cobra.ExactArgs(2),
		Aliases: []string{"tx"},
		Short:   "Creates a new trigger that fires when a tx event is detected.",
		Long:    strings.TrimSpace(`Creates a new trigger.  This will delay the execution of the provided message until the tx event has occurred`),
		Example: fmt.Sprintf(`$ %[1]s tx trigger create-tx-trigger event.json message.json`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			callerAddr := clientCtx.GetFromAddress()

			event, err := parseEvent(clientCtx.Codec, args[0])
			if err != nil {
				return fmt.Errorf("unable to parse event file : %s", err)
			}

			msgs, err := parseTransactions(clientCtx.Codec, args[1])
			if err != nil {
				return fmt.Errorf("unable to parse msgs file: %s", err)
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

func GetCmdAddBlockHeightTrigger() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-height-trigger [height] [msg.json]",
		Args:    cobra.ExactArgs(2),
		Aliases: []string{"ht", "height"},
		Short:   "Creates a new trigger that fires when a block height is reached",
		Long:    strings.TrimSpace(`Creates a new trigger.  This will delay the execution of the provided message until the block height event has occurred`),
		Example: fmt.Sprintf(`$ %[1]s tx trigger create-height-trigger 500 message.json`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			callerAddr := clientCtx.GetFromAddress()

			height, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid argument : %s", args[0])
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
				&types.BlockHeightEvent{BlockHeight: uint64(height)},
				msgs,
			)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func GetCmdAddBlockTimeTrigger() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-time-trigger [seconds] [msg.json]",
		Args:    cobra.ExactArgs(2),
		Aliases: []string{"tt", "time"},
		Short:   "Creates a new trigger that fires when a block time is reached",
		Long:    strings.TrimSpace(`Creates a new trigger.  This will delay the execution of the provided message until the block time event has occurred`),
		Example: fmt.Sprintf(`$ %[1]s tx trigger create-time-trigger 2006-01-02T15:04:05-04:00 message.json`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			callerAddr := clientCtx.GetFromAddress()

			startTime, err := time.Parse(time.RFC3339, args[0])
			if err != nil {
				return fmt.Errorf("unable to parse time (%v) required format is RFC3339 (%v) , %w", args[0], time.RFC3339, err)
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
				&types.BlockTimeEvent{Time: startTime},
				msgs,
			)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func GetCmdDestroyTrigger() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "destroy-trigger [id]",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"destroy", "d"},
		Short:   "Destroys an existing trigger.",
		Long:    strings.TrimSpace(`Destroys an existing trigger. The trigger will not be destroyable if it has already been detected.`),
		Example: fmt.Sprintf(`$ %[1]s tx trigger destroy-trigger 1`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			callerAddr := clientCtx.GetFromAddress()
			triggerID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid argument : %s", args[0])
			}

			msg := types.NewDestroyTriggerRequest(
				callerAddr.String(),
				uint64(triggerID),
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
