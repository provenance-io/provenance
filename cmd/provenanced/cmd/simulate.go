package cmd

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
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
		Short: "Simulate transaction and return estimated costs with possible msg fees.",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			theTx, err := authclient.ReadTxFromFile(clientCtx, args[0])
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
