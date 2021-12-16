package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/provenance-io/provenance/x/msgfees/types"
)

const (
	flagDefaultDenom = "default-denom"
)

func GetCmdPioSimulateTx() *cobra.Command {
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

			defaultDenom, err := cmd.Flags().GetString(flagDefaultDenom)
			if err != nil {
				return err
			}

			gasAdustment, err := cmd.Flags().GetFloat64(flags.FlagGasAdjustment)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			var response *types.CalculateTxFeesResponse
			if response, err = queryClient.CalculateTxFees(
				context.Background(),
				&types.CalculateTxFeesRequest{
					TxBytes:          txBytes,
					DefaultBaseDenom: defaultDenom,
					GasAdjustment:    float32(gasAdustment),
				},
			); err != nil {
				fmt.Printf("failed to calculate fees: %s\n", err.Error())
				return nil
			}
			return clientCtx.PrintProto(response)
		},
	}
	cmd.Flags().String(flagDefaultDenom, "nhash", "Denom used for gas costs")
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
