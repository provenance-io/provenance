package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/version"

	hold "github.com/provenance-io/provenance/x/hold"
)

// exampleQueryCmdBase is the base command that gets a user to one of the query commands in here.
var exampleQueryCmdBase = fmt.Sprintf("%s query %s", version.AppName, hold.ModuleName)

var exampleQueryAddr1 = sdk.AccAddress("exampleQueryAddr1___")

func QueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        hold.ModuleName,
		Short:                      "Querying commands for the hold module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		QueryCmdGetHolds(),
		QueryCmdGetAllHolds(),
	)

	return cmd
}

func QueryCmdGetHolds() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get <address>",
		Aliases: []string{"get-hold", "on-hold"},
		Short:   "Get the funds that are on hold for an address.",
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

			req := hold.GetHoldsRequest{
				Address: args[0],
			}

			var res *hold.GetHoldsResponse
			queryClient := hold.NewQueryClient(clientCtx)
			res, err = queryClient.GetHolds(cmd.Context(), &req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func QueryCmdGetAllHolds() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "all",
		Aliases: []string{"get-all"},
		Short:   "Get all funds on hold for all accounts",
		Example: fmt.Sprintf("$ %s all", exampleQueryCmdBase),
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			req := hold.GetAllHoldsRequest{}
			req.Pagination, err = client.ReadPageRequestWithPageKeyDecoded(cmd.Flags())
			if err != nil {
				return err
			}

			var res *hold.GetAllHoldsResponse
			queryClient := hold.NewQueryClient(clientCtx)
			res, err = queryClient.GetAllHolds(cmd.Context(), &req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "all holds")

	return cmd
}
