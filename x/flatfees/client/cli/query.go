package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

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
