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

// NewTxCmd is the top-level command for trigger CLI transactions.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"t"},
		Short:                      "Transaction commands for the trigger module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		GetCmdAddTransactionTrigger(),
		GetCmdAddBlockHeightTrigger(),
		GetCmdAddBlockTimeTrigger(),
		GetCmdDestroyTrigger(),
	)

	return txCmd
}

// GetCmdAddTransactionTrigger is a command to add a trigger for a transaction event.
func GetCmdAddTransactionTrigger() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-tx-trigger <event.json> <msg.json>",
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

			event, err := parseEvent(args[0])
			if err != nil {
				return fmt.Errorf("unable to parse event file: %w", err)
			}

			msgs, err := parseMessages(clientCtx.Codec, args[1])
			if err != nil {
				return fmt.Errorf("unable to parse message file: %w", err)
			}
			if len(msgs) == 0 {
				return fmt.Errorf("no actions added to trigger")
			}

			msg, err := types.NewCreateTriggerRequest(
				[]string{callerAddr.String()},
				event,
				msgs,
			)
			if err != nil {
				return fmt.Errorf("error creating %T: %w", msg, err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdAddBlockHeightTrigger is a command to add a trigger for a block height event.
func GetCmdAddBlockHeightTrigger() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-height-trigger <height> <msg.json>",
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
				return fmt.Errorf("invalid block height %q: %w", args[0], err)
			}

			msgs, err := parseMessages(clientCtx.Codec, args[1])
			if err != nil {
				return fmt.Errorf("unable to parse message file: %w", err)
			}
			if len(msgs) == 0 {
				return fmt.Errorf("no actions added to trigger")
			}

			msg, err := types.NewCreateTriggerRequest(
				[]string{callerAddr.String()},
				&types.BlockHeightEvent{BlockHeight: uint64(height)},
				msgs,
			)
			if err != nil {
				return fmt.Errorf("error creating %T: %w", msg, err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdAddBlockTimeTrigger is a command to add a trigger for a block time event.
func GetCmdAddBlockTimeTrigger() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-time-trigger <seconds> <msg.json>",
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
				return fmt.Errorf("unable to parse time (%v) required format is RFC3339 (%v): %w", args[0], time.RFC3339, err)
			}

			msgs, err := parseMessages(clientCtx.Codec, args[1])
			if err != nil {
				return fmt.Errorf("unable to parse message file: %w", err)
			}
			if len(msgs) == 0 {
				return fmt.Errorf("no actions added to trigger")
			}

			msg, err := types.NewCreateTriggerRequest(
				[]string{callerAddr.String()},
				&types.BlockTimeEvent{Time: startTime.UTC()},
				msgs,
			)
			if err != nil {
				return fmt.Errorf("error creating %T: %w", msg, err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdDestroyTrigger is a command to destroy an existing trigger.
func GetCmdDestroyTrigger() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "destroy-trigger <id>",
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
				return fmt.Errorf("invalid trigger id %q: %w", args[0], err)
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

// parseMessages reads and parses the message.
func parseMessages(cdc codec.Codec, path string) ([]sdk.Msg, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var msg sdk.Msg
	err = cdc.UnmarshalInterfaceJSON(contents, &msg)
	if err != nil {
		return nil, err
	}

	return []sdk.Msg{msg}, nil
}

// parseEvent reads and parses the transaction event from a file.
func parseEvent(path string) (*types.TransactionEvent, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var event types.TransactionEvent
	err = json.Unmarshal(contents, &event)
	if err != nil {
		return nil, err
	}

	return &event, nil
}
