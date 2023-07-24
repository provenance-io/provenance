package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/provenance-io/provenance/x/escrow"
)

// exampleQueryCmdBase is the base command that gets a user to one of the query commands in here.
var exampleQueryCmdBase = fmt.Sprintf("%s query %s", version.AppName, escrow.ModuleName)

var exampleQueryAddr1 = sdk.AccAddress("exampleQueryAddr1___")

func QueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        escrow.ModuleName,
		Short:                      "Querying commands for the escrow module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		QueryCmdGetEscrow(),
		QueryCmdGetAllEscrow(),
	)

	return cmd
}

func QueryCmdGetEscrow() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get <address>",
		Aliases: []string{"get-escrow"},
		Short:   "Get the funds that are in escrow for an address.",
		Example: fmt.Sprintf("$ %s get %s", exampleQueryCmdBase, exampleQueryAddr1),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			if _, err = sdk.AccAddressFromBech32(args[0]); err != nil {
				return sdkerrors.ErrInvalidAddress.Wrap(err.Error())
			}

			req := escrow.GetEscrowRequest{
				Address: args[0],
			}

			var res *escrow.GetEscrowResponse
			queryClient := escrow.NewQueryClient(clientCtx)
			res, err = queryClient.GetEscrow(cmd.Context(), &req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func QueryCmdGetAllEscrow() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "all",
		Aliases: []string{"get-all"},
		Short:   "Get all funds in escrow for all accounts",
		Example: fmt.Sprintf("$ %s all", exampleQueryCmdBase),
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			req := escrow.GetAllEscrowRequest{}

			var res *escrow.GetAllEscrowResponse
			queryClient := escrow.NewQueryClient(clientCtx)
			res, err = queryClient.GetAllEscrow(cmd.Context(), &req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
