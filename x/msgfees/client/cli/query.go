package cli

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"

	flag "github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/msgfees/types"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the top-level command for msgfees CLI queries.
func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the msgfees module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	queryCmd.AddCommand(
		AllMsgBasedFeesCmd(),
		GetCmdCalculateMsgBasedFees(),
	)
	return queryCmd
}

// AllMsgBasedFeesCmd is the CLI command for listing all msg based fees.
func AllMsgBasedFeesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "List all the msg based fees on the Provenance Blockchain",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(withPageKeyDecoded(cmd.Flags()))
			if err != nil {
				return err
			}

			var response *types.QueryAllMsgBasedFeesResponse
			if response, err = queryClient.QueryAllMsgBasedFees(
				context.Background(),
				&types.QueryAllMsgBasedFeesRequest{Pagination: pageReq},
			); err != nil {
				fmt.Printf("failed to query msg based fees: %s\n", err.Error())
				return nil
			}
			return clientCtx.PrintProto(response)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "markers")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetCmdCalculateMsgBasedFees() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "simulate [msg_tx_json_file]",
		Args:  cobra.ExactArgs(1),
		Short: "Simulate transaction and return estimated costs with possible msg based fees.",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// sdk.Tx
			theTx, err := ReadTxFromFile(clientCtx, args[0])
			if err != nil {
				return err
			}
			txBytes, err := clientCtx.TxConfig.TxEncoder()(theTx)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			var response *types.CalculateTxFeesResponse
			if response, err = queryClient.CalculateTxFees(
				context.Background(),
				&types.CalculateTxFeesRequest{Tx: txBytes},
			); err != nil {
				fmt.Printf("failed to calculate fees: %s\n", err.Error())
				return nil
			}
			return clientCtx.PrintProto(response)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// Read and decode a StdTx from the given filename.  Can pass "-" to read from stdin.
func ReadTxFromFile(ctx client.Context, filename string) (tx sdk.Tx, err error) {
	var bytes []byte

	if filename == "-" {
		bytes, err = ioutil.ReadAll(os.Stdin)
	} else {
		bytes, err = ioutil.ReadFile(filename)
	}

	if err != nil {
		return
	}

	return ctx.TxConfig.TxJSONDecoder()(bytes)
}

// sdk ReadPageRequest expects binary but we encoded to base64 in our marshaller
func withPageKeyDecoded(flagSet *flag.FlagSet) *flag.FlagSet {
	encoded, err := flagSet.GetString(flags.FlagPageKey)
	if err != nil {
		panic(err.Error())
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		panic(err.Error())
	}
	_ = flagSet.Set(flags.FlagPageKey, string(raw))
	return flagSet
}
