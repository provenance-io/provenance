package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"

	"github.com/provenance-io/provenance/x/flatfees/types"
)

// NewQueryCmd returns the top-level command for x/flatfees CLI queries.
func NewQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"fees", "ff"},
		Short:                      "Querying commands for the x/flatfees module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	queryCmd.AddCommand(
		NewCmdGetParams(),
		NewCmdGetAllMsgFees(),
		NewCmdGetMsgFee(),
		NewCmdCalculateTxFees(),
	)
	return queryCmd
}

const FlagDoNotConvert = "do-not-convert"

// AddFlagDoNotConvert adds the --do-not-convert flag to the command.
func AddFlagDoNotConvert(cmd *cobra.Command) {
	cmd.Flags().Bool(FlagDoNotConvert, false, "Do not convert msg fee cost into the fee denom, return them as they are defined")
}

// ReadFlagDoNotConvert reads the --do-not-convert flag.
func ReadFlagDoNotConvert(flagSet *pflag.FlagSet) (bool, error) {
	return flagSet.GetBool(FlagDoNotConvert)
}

// NewCmdGetParams is the CLI command for listing all params.
func NewCmdGetParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "List the x/flatfees params on the Provenance Blockchain",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			cmd.SilenceUsage = true
			response, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
			if err != nil {
				return fmt.Errorf("failed to query x/flatfees params: %w", err)
			}
			return clientCtx.PrintProto(response)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// NewCmdGetAllMsgFees is the CLI command for listing all msg fees.
func NewCmdGetAllMsgFees() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l", "all"},
		Short:   "List all the msg fees on the Provenance Blockchain",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			flagSet := cmd.Flags()
			pageReq, err := client.ReadPageRequestWithPageKeyDecoded(flagSet)
			if err != nil {
				return err
			}

			req := &types.QueryAllMsgFeesRequest{Pagination: pageReq}
			req.DoNotConvert, err = ReadFlagDoNotConvert(flagSet)
			if err != nil {
				return err
			}

			cmd.SilenceUsage = true
			response, err := queryClient.AllMsgFees(context.Background(), req)
			if err != nil {
				return fmt.Errorf("failed to query msg fees: %w", err)
			}
			return clientCtx.PrintProto(response)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "msgfees")
	flags.AddQueryFlagsToCmd(cmd)
	AddFlagDoNotConvert(cmd)
	return cmd
}

// NewCmdGetMsgFee is the CLI command for looking up a single Msg fee.
func NewCmdGetMsgFee() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get <msg type url>",
		Aliases: []string{"msgfee", "fee"},
		Short:   "Get the msg fee for a specific msg type on the Provenance Blockchain",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			if len(args) == 0 || len(args[0]) == 0 {
				return errors.New("no msg-type-url provided")
			}

			req := &types.QueryMsgFeeRequest{MsgTypeUrl: args[0]}
			req.DoNotConvert, err = ReadFlagDoNotConvert(cmd.Flags())
			if err != nil {
				return err
			}

			cmd.SilenceUsage = true
			response, err := queryClient.MsgFee(context.Background(), req)
			if err != nil {
				return fmt.Errorf("failed to query msg fee for %q: %w", req.MsgTypeUrl, err)
			}
			return clientCtx.PrintProto(response)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	AddFlagDoNotConvert(cmd)
	return cmd
}

func NewCmdCalculateTxFees() *cobra.Command {
	// This cmd is named to match this module's query, but as the alias "simulate" because
	// that's what the command is called under the tx sub-command.
	cmd := &cobra.Command{
		Use:     "calculate-tx-fees [msg_tx_json_file]",
		Aliases: []string{"simulate", "calculate-fees", "sim", "calc"},
		Args:    cobra.ExactArgs(1),
		Short:   "Simulate a Tx and return total fees and estimated gas.",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return fmt.Errorf("error reading tx flags: %w", err)
			}
			queryCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return fmt.Errorf("error reading query flags: %w", err)
			}

			gasAdj, err := cmd.Flags().GetFloat64(flags.FlagGasAdjustment)
			if err != nil {
				return fmt.Errorf("error reading --%s flag: %w", flags.FlagGasAdjustment, err)
			}

			// We've read all the command stuff. Any errors now, aren't usage problems.
			cmd.SilenceUsage = true

			theTx, err := authclient.ReadTxFromFile(clientCtx, args[0])
			if err != nil {
				return fmt.Errorf("error reading tx from file %q: %w", args[0], err)
			}

			txBytes, err := clientCtx.TxConfig.TxEncoder()(theTx)
			if err != nil {
				return fmt.Errorf("error decoding tx bytes: %w", err)
			}

			queryClient := types.NewQueryClient(queryCtx)
			resp, err := queryClient.CalculateTxFees(
				context.Background(),
				&types.QueryCalculateTxFeesRequest{TxBytes: txBytes, GasAdjustment: float32(gasAdj)},
			)
			if err != nil {
				return fmt.Errorf("error calculating fees: %w", err)
			}
			return queryCtx.PrintProto(resp)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	// We also want the contents of AddQueryFlagsToCmd applied, but it has a couple in common with AddTxFlagsToCmd.
	// That makes the second one panic due to duplicate flag. Since AddQueryFlagsToCmd is smaller, we manually apply
	// what's in there, but not in AddTxFlagsToCmd. In both: FlagNode, FlagOutput.
	cmd.Flags().String(flags.FlagGRPC, "", "the gRPC endpoint to use for this chain")
	cmd.Flags().Bool(flags.FlagGRPCInsecure, false, "allow gRPC over insecure channels, if not the server must use TLS")
	cmd.Flags().Int64(flags.FlagHeight, 0, "Use a specific height to query state at (this can error if the node is pruning state)")
	// We're fine without marking FlagChainID as required.

	// Also, AddQueryFlagsToCmd sets the default FlagOutput to text, but AddTxFlagsToCmd defaults to json.
	// So this query defaults to json.

	return cmd
}
