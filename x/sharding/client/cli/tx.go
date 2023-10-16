package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/provenance-io/provenance/x/sharding/types"
)

// NewTxCmd is the top-level command for sharding CLI transactions.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Transaction commands for the sharding module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		GetCmdRead(),
		GetCmdWrite(),
	)

	return txCmd
}

// GetCmdRead is a command to test reading
func GetCmdRead() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "read",
		Short:   "Read from the store",
		Long:    "Specify different read types to test gas costs",
		Args:    cobra.ExactArgs(7),
		Example: fmt.Sprintf(`%[1]s tx sharding read`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			iterations, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid iterations %q: %w", args[0], err)
			}
			full, err := strconv.ParseBool(args[1])
			if err != nil {
				return fmt.Errorf("invalid full %q: %w", args[1], err)
			}
			group, err := strconv.ParseBool(args[2])
			if err != nil {
				return fmt.Errorf("invalid group %q: %w", args[2], err)
			}

			owner, err := strconv.ParseBool(args[3])
			if err != nil {
				return fmt.Errorf("invalid owner %q: %w", args[3], err)
			}
			name, err := strconv.ParseBool(args[4])
			if err != nil {
				return fmt.Errorf("invalid name %q: %w", args[4], err)
			}
			color, err := strconv.ParseBool(args[5])
			if err != nil {
				return fmt.Errorf("invalid color %q: %w", args[5], err)
			}
			spots, err := strconv.ParseBool(args[6])
			if err != nil {
				return fmt.Errorf("invalid spots %q: %w", args[6], err)
			}
			sharded := false

			fmt.Printf("iterations: %v\n", iterations)
			fmt.Printf("full: %v\n", full)
			fmt.Printf("group: %v\n", group)
			fmt.Printf("owner: %v\n", owner)
			fmt.Printf("name: %v\n", name)
			fmt.Printf("color: %v\n", color)
			fmt.Printf("spots: %v\n", spots)

			msg := types.NewMsgRead(
				clientCtx.GetFromAddress().String(),
				owner,
				name,
				color,
				spots,
				full,
				group,
				sharded,
				uint64(iterations),
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdWrite is a command to test reading
func GetCmdWrite() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "write",
		Short:   "Write from the store",
		Long:    "Specify different write types to test gas costs",
		Args:    cobra.ExactArgs(7),
		Example: fmt.Sprintf(`%[1]s tx sharding write`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			iterations, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid iterations %q: %w", args[0], err)
			}
			full, err := strconv.ParseBool(args[1])
			if err != nil {
				return fmt.Errorf("invalid full %q: %w", args[1], err)
			}
			group, err := strconv.ParseBool(args[2])
			if err != nil {
				return fmt.Errorf("invalid group %q: %w", args[2], err)
			}

			owner, err := strconv.ParseBool(args[3])
			if err != nil {
				return fmt.Errorf("invalid owner %q: %w", args[3], err)
			}
			name, err := strconv.ParseBool(args[4])
			if err != nil {
				return fmt.Errorf("invalid name %q: %w", args[4], err)
			}
			color, err := strconv.ParseBool(args[5])
			if err != nil {
				return fmt.Errorf("invalid color %q: %w", args[5], err)
			}
			spots, err := strconv.ParseBool(args[6])
			if err != nil {
				return fmt.Errorf("invalid spots %q: %w", args[6], err)
			}
			sharded := false

			fmt.Printf("iterations: %v\n", iterations)
			fmt.Printf("full: %v\n", full)
			fmt.Printf("group: %v\n", group)
			fmt.Printf("owner: %v\n", owner)
			fmt.Printf("name: %v\n", name)
			fmt.Printf("color: %v\n", color)
			fmt.Printf("spots: %v\n", spots)

			msg := types.NewMsgWrite(
				clientCtx.GetFromAddress().String(),
				owner,
				name,
				color,
				spots,
				full,
				group,
				sharded,
				uint64(iterations),
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdUpdate is a command to test reading
func GetCmdUpdate() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update",
		Short:   "Update from the store",
		Long:    "Specify different write types to test gas costs",
		Args:    cobra.ExactArgs(0),
		Example: fmt.Sprintf(`%[1]s tx sharding update`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdate(
				clientCtx.GetFromAddress().String(),
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
