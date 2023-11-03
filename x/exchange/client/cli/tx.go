package cli

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/provenance-io/provenance/x/exchange"
)

var txCmdStart = fmt.Sprintf("%s tx %s", version.AppName, exchange.ModuleName)

// CmdTx creates the tx command (and sub-commands) for the exchange module.
func CmdTx() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        exchange.ModuleName,
		Aliases:                    []string{"ex"},
		Short:                      "Transaction commands for the exchange module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdTxCreateAsk(),
		CmdTxCreateBid(),
		CmdTxCancelOrder(),
		CmdTxFillBids(),
		CmdTxFillAsks(),
		CmdTxMarketSettle(),
		CmdTxMarketSetOrderExternalID(),
		CmdTxMarketWithdraw(),
		CmdTxMarketUpdateDetails(),
		CmdTxMarketUpdateEnabled(),
		CmdTxMarketUpdateUserSettle(),
		CmdTxMarketManagePermissions(),
		CmdTxMarketManageReqAttrs(),
		CmdTxGovCreateMarket(),
		CmdTxGovManageFees(),
		CmdTxGovUpdateParams(),
	)

	return cmd
}

// CmdTxCreateAsk creates the create-ask sub-command for the exchange tx command.
func CmdTxCreateAsk() *cobra.Command {
	cmd := &cobra.Command{
		Use: fmt.Sprintf("create-ask --%s <seller> --%s <market id> --%s <assets> --%s <price> "+
			"[--%s <seller settlement flat fee>] [--%s] [--%s <external id>] [--%s <creation fee>]",
			flags.FlagFrom, FlagMarket, FlagAssets, FlagPrice,
			FlagSettlementFee, FlagAllowPartial, FlagExternalID, FlagCreationFee,
		),
		Aliases: []string{"ask", "create-ask-order"},
		Short:   "Create an ask order",
		Long: fmt.Sprintf(`Create an ask order.

The following flags are required:
  --%-19s (either the account address or keyring name)
  --%-19s e.g. 3
  --%-19s e.g. 10nhash
  --%-19s e.g. 10nhash

The following flags are optional:
  --%-44s e.g. 10nhash
  --%-44s (defaults to false if not provided)
  --%-44s e.g. 5D8759ED-0A1F-4952-9B9C-87A90099ED61
  --%-44s e.g. 10nhash
`,
			flags.FlagFrom+" <seller>",
			FlagMarket+" <market id>", // = 18 characters
			FlagAssets+" <assets>",
			FlagPrice+" <price>",
			FlagSettlementFee+" <seller settlement flat fee>", // = 43 characters
			FlagAllowPartial,
			FlagExternalID+" <external id>",
			FlagCreationFee+" <creation fee>",
		),
		Example: fmt.Sprintf(`%s --%s %s --%s %s --%s %s --%s %s --%s --%s %s`,
			txCmdStart+" create-ask",
			flags.FlagFrom, ExampleAddr,
			FlagAssets, "10atom",
			FlagPrice, "1000000000nhash",
			FlagMarket, "3",
			FlagAllowPartial,
			FlagCreationFee, "500nhash",
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &exchange.MsgCreateAskRequest{
				AskOrder: exchange.AskOrder{
					Seller: clientCtx.FromAddress.String(),
				},
			}

			flagSet := cmd.Flags()
			errs := make([]error, 7)
			msg.AskOrder.MarketId, errs[0] = ReadFlagMarket(flagSet)
			msg.AskOrder.Assets, errs[1] = ReadFlagAssets(flagSet)
			msg.AskOrder.Price, errs[2] = ReadFlagPrice(flagSet)
			msg.AskOrder.SellerSettlementFlatFee, errs[3] = ReadFlagSettlementFeeCoin(flagSet)
			msg.AskOrder.AllowPartial, errs[4] = ReadFlagAllowPartial(flagSet)
			msg.AskOrder.ExternalId, errs[5] = ReadFlagExternalID(flagSet)
			msg.OrderCreationFee, errs[6] = ReadFlagCreationFee(flagSet)
			err = errors.Join(errs...)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, flagSet, msg)
		},
	}

	AddFlagMarket(cmd, true)
	AddFlagAssets(cmd)
	AddFlagPrice(cmd)
	AddFlagSettlementFee(cmd)
	AddFlagAllowPartial(cmd)
	AddFlagExternalID(cmd)
	AddFlagCreationFee(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxCreateBid creates the create-bid sub-command for the exchange tx command.
func CmdTxCreateBid() *cobra.Command {
	cmd := &cobra.Command{
		Use: fmt.Sprintf("create-bid --%s <buyer> --%s <market id> --%s <assets> --%s <price> "+
			"[--%s <buyer settlement fees>] [--%s] [--%s <external id>] [--%s <creation fee>]",
			flags.FlagFrom, FlagMarket, FlagAssets, FlagPrice,
			FlagSettlementFee, FlagAllowPartial, FlagExternalID, FlagCreationFee,
		),
		Aliases: []string{"bid", "create-bid-order"},
		Short:   "Create a bid order",
		Long: fmt.Sprintf(`Create a bid order.

The following flags are required:
  --%-19s (either the account address or keyring name)
  --%-19s e.g. 3
  --%-19s e.g. 10nhash
  --%-19s e.g. 10nhash

The following flags are optional:
  --%-39s e.g. 10nhash
  --%-39s (defaults to false if not provided)
  --%-39s e.g. 5D8759ED-0A1F-4952-9B9C-87A90099ED61
  --%-39s e.g. 10nhash
`,
			flags.FlagFrom+" <buyer>",
			FlagMarket+" <market id>", // = 18 characters
			FlagAssets+" <assets>",
			FlagPrice+" <price>",
			FlagSettlementFee+" <buyer settlement fees>", // = 38 characters
			FlagAllowPartial,
			FlagExternalID+" <external id>",
			FlagCreationFee+" <creation fee>",
		),
		Example: fmt.Sprintf(`%s --%s %s --%s %s --%s %s --%s %s --%s --%s %s`,
			txCmdStart+" create-bid",
			flags.FlagFrom, ExampleAddr,
			FlagAssets, "10atom",
			FlagPrice, "1000000000nhash",
			FlagMarket, "3",
			FlagAllowPartial,
			FlagCreationFee, "500nhash",
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &exchange.MsgCreateBidRequest{
				BidOrder: exchange.BidOrder{
					Buyer: clientCtx.FromAddress.String(),
				},
			}

			flagSet := cmd.Flags()
			errs := make([]error, 7)
			msg.BidOrder.MarketId, errs[0] = ReadFlagMarket(flagSet)
			msg.BidOrder.Assets, errs[1] = ReadFlagAssets(flagSet)
			msg.BidOrder.Price, errs[2] = ReadFlagPrice(flagSet)
			msg.BidOrder.BuyerSettlementFees, errs[3] = ReadFlagSettlementFeeCoins(flagSet)
			msg.BidOrder.AllowPartial, errs[4] = ReadFlagAllowPartial(flagSet)
			msg.BidOrder.ExternalId, errs[5] = ReadFlagExternalID(flagSet)
			msg.OrderCreationFee, errs[6] = ReadFlagCreationFee(flagSet)
			err = errors.Join(errs...)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, flagSet, msg)
		},
	}

	AddFlagMarket(cmd, true)
	AddFlagAssets(cmd)
	AddFlagPrice(cmd)
	AddFlagSettlementFee(cmd)
	AddFlagAllowPartial(cmd)
	AddFlagExternalID(cmd)
	AddFlagCreationFee(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxCancelOrder creates the cancel-order sub-command for the exchange tx command.
func CmdTxCancelOrder() *cobra.Command {
	cmd := &cobra.Command{
		Use:     fmt.Sprintf("cancel-order {<order id>|--order <order id>} --%s <signer>", flags.FlagFrom),
		Aliases: []string{"cancel"},
		Short:   "Cancel an order",
		Long: `Cancel an order.

The <order id> must be provided either as the first argument or using the --order flag, but not both.
`,
		Example: fmt.Sprintf(`%[1]s 5 --%[2]s %[3]s
%[1]s --%[2]s %[3]s --%[4]s 5
`,
			txCmdStart+" cancel-order", flags.FlagFrom, ExampleAddr, FlagOrder,
		),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &exchange.MsgCancelOrderRequest{
				Signer: clientCtx.FromAddress.String(),
			}

			flagSet := cmd.Flags()
			msg.OrderId, err = ReadFlagOrder(flagSet)
			if err != nil {
				return err
			}

			if len(args) > 0 && len(args[0]) > 0 {
				var orderID uint64
				orderID, err = strconv.ParseUint(args[0], 10, 64)
				if err != nil {
					return fmt.Errorf("could not convert <order id> arg %q to uint64: %w", args[0], err)
				}
				if msg.OrderId != 0 && orderID != 0 && msg.OrderId != orderID {
					return fmt.Errorf("cannot provide an <order id> as both an arg (%d) and flag (--%s %d)",
						orderID, FlagOrder, msg.OrderId)
				}
				msg.OrderId = orderID
			}

			if msg.OrderId == 0 {
				return errors.New("no <order id> provided")
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, flagSet, msg)
		},
	}

	AddFlagOrder(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxFillBids creates the fill-bids sub-command for the exchange tx command.
func CmdTxFillBids() *cobra.Command {
	cmd := &cobra.Command{
		Use: fmt.Sprintf("fill-bids --%s <seller> --%s <market id> --%s <total assets> --%s <bid order ids> "+
			"[--%s <seller settlement flat fee>] [--%s <ask order creation fee>]",
			flags.FlagFrom, FlagMarket, FlagAssets, FlagBids, FlagSettlementFee, FlagCreationFee,
		),
		Short: "Fill one or more bid orders",
		Long: fmt.Sprintf(`Fill one or more bid orders.

The following flags are required:
  --%-22s (either the account address or keyring name)
  --%-22s e.g. 3
  --%-22s e.g. 10nhash
  --%-22s comma-separate each id; this flag can also be provided multiple times

The following flags are optional:
  --%-44s e.g. 10nhash
  --%-44s e.g. 10nhash
`,
			flags.FlagFrom+" <seller>",
			FlagMarket+" <market id>",
			FlagAssets+" <total assets>", // = 21 characters.
			FlagBids+" <bid order ids>",
			FlagSettlementFee+" <seller settlement flat fee>", // = 43 characters.
			FlagCreationFee+" <ask order creation fee>",
		),
		Example: fmt.Sprintf(`%[1]s --%[2]s %[3]s --%[4]s %[5]s --%[6]s %[7]s --%[8]s %[9]s
%[1]s --%[2]s %[3]s --%[4]s %[5]s --%[6]s %[7]s --%[8]s %[10]s --%[8]s %[11]s
`,
			txCmdStart+" fill-bids",
			flags.FlagFrom, ExampleAddr,
			FlagMarket, "3",
			FlagAssets, "10atom",
			FlagBids, "1,2,3", "1", "2,3",
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &exchange.MsgFillBidsRequest{
				Seller: clientCtx.FromAddress.String(),
			}
			errs := make([]error, 5)
			flagSet := cmd.Flags()
			msg.MarketId, errs[0] = ReadFlagMarket(flagSet)
			msg.TotalAssets, errs[1] = ReadFlagTotalAssets(flagSet)
			msg.BidOrderIds, errs[2] = ReadFlagBids(flagSet)
			msg.SellerSettlementFlatFee, errs[3] = ReadFlagSettlementFeeCoin(flagSet)
			msg.AskOrderCreationFee, errs[4] = ReadFlagCreationFee(flagSet)
			err = errors.Join(errs...)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, flagSet, msg)
		},
	}

	AddFlagMarket(cmd, true)
	AddFlagTotalAssets(cmd)
	AddFlagBids(cmd)
	AddFlagSettlementFee(cmd)
	AddFlagCreationFee(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxFillAsks creates the fill-asks sub-command for the exchange tx command.
func CmdTxFillAsks() *cobra.Command {
	cmd := &cobra.Command{
		Use: fmt.Sprintf("fill-asks --%s <buyer> --%s <market id> --%s <total price> --%s <ask order ids> "+
			"[--%s <buyer settlement fees>] [--%s <bid order creation fee>]",
			flags.FlagFrom, FlagMarket, FlagPrice, FlagAsks, FlagSettlementFee, FlagCreationFee,
		),
		Short: "Fill one or more ask orders",
		Long: fmt.Sprintf(`Fill one or more ask orders.

The following flags are required:
  --%-21s (either the account address or keyring name)
  --%-21s e.g. 3
  --%-21s e.g. 10nhash
  --%-21s comma-separate each id; this flag can also be provided multiple times

The following flags are optional:
  --%-39s e.g. 10nhash
  --%-39s e.g. 10nhash
`,
			flags.FlagFrom+" <buyer>",
			FlagMarket+" <market id>",
			FlagPrice+" <total price>",
			FlagAsks+" <ask order ids>",                  // = 20 characters.
			FlagSettlementFee+" <buyer settlement fees>", // = 38 characters.
			FlagCreationFee+" <bid order creation fee>",
		),
		Example: fmt.Sprintf(`%[1]s --%[2]s %[3]s --%[4]s %[5]s --%[6]s %[7]s --%[8]s %[9]s
%[1]s --%[2]s %[3]s --%[4]s %[5]s --%[6]s %[7]s --%[8]s %[10]s --%[8]s %[11]s
`,
			txCmdStart+" fill-asks",
			flags.FlagFrom, ExampleAddr,
			FlagMarket, "3",
			FlagAssets, "10atom",
			FlagBids, "1,2,3", "1", "2,3",
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &exchange.MsgFillAsksRequest{
				Buyer: clientCtx.FromAddress.String(),
			}
			errs := make([]error, 5)
			flagSet := cmd.Flags()
			msg.MarketId, errs[0] = ReadFlagMarket(flagSet)
			msg.TotalPrice, errs[1] = ReadFlagPrice(flagSet)
			msg.AskOrderIds, errs[2] = ReadFlagAsks(flagSet)
			msg.BuyerSettlementFees, errs[3] = ReadFlagSettlementFeeCoins(flagSet)
			msg.BidOrderCreationFee, errs[4] = ReadFlagCreationFee(flagSet)
			err = errors.Join(errs...)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, flagSet, msg)
		},
	}

	AddFlagMarket(cmd, true)
	AddFlagTotalPrice(cmd)
	AddFlagAsks(cmd)
	AddFlagSettlementFee(cmd)
	AddFlagCreationFee(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxMarketSettle TODO
func CmdTxMarketSettle() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, txCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdTxMarketSettle
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxMarketSetOrderExternalID TODO
func CmdTxMarketSetOrderExternalID() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, txCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdTxMarketSetOrderExternalID
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxMarketWithdraw TODO
func CmdTxMarketWithdraw() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, txCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdTxMarketWithdraw
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxMarketUpdateDetails TODO
func CmdTxMarketUpdateDetails() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, txCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdTxMarketUpdateDetails
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxMarketUpdateEnabled TODO
func CmdTxMarketUpdateEnabled() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, txCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdTxMarketUpdateEnabled
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxMarketUpdateUserSettle TODO
func CmdTxMarketUpdateUserSettle() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, txCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdTxMarketUpdateUserSettle
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxMarketManagePermissions TODO
func CmdTxMarketManagePermissions() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, txCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdTxMarketManagePermissions
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxMarketManageReqAttrs TODO
func CmdTxMarketManageReqAttrs() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, txCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdTxMarketManageReqAttrs
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxGovCreateMarket TODO
func CmdTxGovCreateMarket() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, txCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdTxGovCreateMarket
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxGovManageFees TODO
func CmdTxGovManageFees() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, txCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdTxGovManageFees
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxGovUpdateParams TODO
func CmdTxGovUpdateParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, txCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdTxGovUpdateParams
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
