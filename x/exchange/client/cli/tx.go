package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/provenance-io/provenance/x/exchange"
)

var txCmdStart = fmt.Sprintf("%s tx %s", version.AppName, exchange.ModuleName)

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

// CmdTxCreateAsk creates the create-ask sub-command for the tx command.
func CmdTxCreateAsk() *cobra.Command {
	cmd := &cobra.Command{
		Use: fmt.Sprintf("create-ask --%s <seller> --%s <market id> --%s <assets> --%s <price> "+
			"[--%s <seller settlement flat fee>] [--%s] [--%s <external id>] [--%s <creation fee>]",
			flags.FlagFrom, FlagMarketID, FlagAssets, FlagPrice,
			FlagSettlementFee, FlagAllowPartial, FlagExternalID, FlagCreationFee),
		Aliases: []string{"ask", "create-ask-order"},
		Short:   "Create an ask order",
		Long: fmt.Sprintf(`Create an ask order.

The following flags are required:
  --%-22s (either the account address or keyring name)
  --%-22s e.g. 3
  --%-22s e.g. 10nhash
  --%-22s e.g. 10nhash

The following flags are optional:
  --%-44s e.g. 10nhash
  --%-44s (defaults to false if not provided)
  --%-44s e.g. 5D8759ED-0A1F-4952-9B9C-87A90099ED61
  --%-44s e.g. 10nhash
`,
			flags.FlagFrom+" <seller>",
			FlagMarketID+" <market id>",
			FlagAssets+" <assets>",
			FlagPrice+" <price>",
			FlagSettlementFee+" <seller settlement flat fee>",
			FlagAllowPartial,
			FlagExternalID+" <external id>",
			FlagCreationFee+" <creation fee>",
		),
		Example: fmt.Sprintf(`%s create-ask --%s %s --%s %s --%s %s --%s %s --%s --%s %s`,
			txCmdStart,
			flags.FlagFrom, ExampleAddr,
			FlagAssets, "10atom",
			FlagPrice, "1000000000nhash",
			FlagMarketID, "3",
			FlagAllowPartial,
			FlagCreationFee, "500nhash",
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			flagSet := cmd.Flags()
			flagVals, err := ReadCreateOrderFlags(flagSet)
			if err != nil {
				return err
			}

			if len(flagVals.SettlementFee) > 1 {
				return fmt.Errorf("only one --%s coin type can be provided", FlagSettlementFee)
			}

			msg := &exchange.MsgCreateAskRequest{
				AskOrder: exchange.AskOrder{
					Seller:                  clientCtx.From,
					MarketId:                flagVals.MarketID,
					Assets:                  flagVals.Assets,
					Price:                   flagVals.Price,
					SellerSettlementFlatFee: nil,
					AllowPartial:            flagVals.AllowPartial,
					ExternalId:              flagVals.ExternalID,
				},
				OrderCreationFee: flagVals.CreationFee,
			}
			if len(flagVals.SettlementFee) > 0 {
				msg.AskOrder.SellerSettlementFlatFee = &flagVals.SettlementFee[0]
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, flagSet, msg)
		},
	}

	AddCreateOrderFlags(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxCreateBid creates the create-bid sub-command for the tx command.
func CmdTxCreateBid() *cobra.Command {
	cmd := &cobra.Command{
		Use: fmt.Sprintf("create-bid --%s <buyer> --%s <market id> --%s <assets> --%s <price> "+
			"[--%s <buyer settlement fees>] [--%s] [--%s <external id>] [--%s <creation fee>]",
			flags.FlagFrom, FlagMarketID, FlagAssets, FlagPrice,
			FlagSettlementFee, FlagAllowPartial, FlagExternalID, FlagCreationFee),
		Aliases: []string{"bid", "create-bid-order"},
		Short:   "Create a bid order",
		Long: fmt.Sprintf(`Create a bid order.

The following flags are required:
  --%-22s (either the account address or keyring name)
  --%-22s e.g. 3
  --%-22s e.g. 10nhash
  --%-22s e.g. 10nhash

The following flags are optional:
  --%-39s e.g. 10nhash
  --%-39s (defaults to false if not provided)
  --%-39s e.g. 5D8759ED-0A1F-4952-9B9C-87A90099ED61
  --%-39s e.g. 10nhash
`,
			flags.FlagFrom+" <buyer>",
			FlagMarketID+" <market id>",
			FlagAssets+" <assets>",
			FlagPrice+" <price>",
			FlagSettlementFee+" <buyer settlement fees>",
			FlagAllowPartial,
			FlagExternalID+" <external id>",
			FlagCreationFee+" <creation fee>",
		),
		Example: fmt.Sprintf(`%s create-bid --%s %s --%s %s --%s %s --%s %s --%s --%s %s`,
			txCmdStart,
			flags.FlagFrom, ExampleAddr,
			FlagAssets, "10atom",
			FlagPrice, "1000000000nhash",
			FlagMarketID, "3",
			FlagAllowPartial,
			FlagCreationFee, "500nhash",
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			flagSet := cmd.Flags()
			flagVals, err := ReadCreateOrderFlags(flagSet)
			if err != nil {
				return err
			}

			msg := &exchange.MsgCreateBidRequest{
				BidOrder: exchange.BidOrder{
					Buyer:               clientCtx.From,
					MarketId:            flagVals.MarketID,
					Assets:              flagVals.Assets,
					Price:               flagVals.Price,
					BuyerSettlementFees: flagVals.SettlementFee,
					AllowPartial:        flagVals.AllowPartial,
					ExternalId:          flagVals.ExternalID,
				},
				OrderCreationFee: flagVals.CreationFee,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, flagSet, msg)
		},
	}

	AddCreateOrderFlags(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxCancelOrder TODO
func CmdTxCancelOrder() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, txCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdTxCancelOrder
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxFillBids TODO
func CmdTxFillBids() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, txCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdTxFillBids
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxFillAsks TODO
func CmdTxFillAsks() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "TODO",
		Aliases: []string{"TODO"},
		Short:   "TODO",
		Long:    `TODO`,
		Example: fmt.Sprintf(`%[1]s TODO`, txCmdStart),
		Args:    cobra.ExactArgs(0), // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO[1701]: CmdTxFillAsks
			return nil
		},
	}

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
