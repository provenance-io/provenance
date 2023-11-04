package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/staking/client/cli"

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
		Use: fmt.Sprintf("create-ask {--%s|--%s} <seller> --%s <market id> --%s <assets> --%s <price> "+
			"[--%s <seller settlement flat fee>] [--%s] [--%s <external id>] [--%s <creation fee>]",
			flags.FlagFrom, FlagSeller, FlagMarket, FlagAssets, FlagPrice,
			FlagSettlementFee, FlagPartial, FlagExternalID, FlagCreationFee,
		),
		Aliases: []string{"ask", "create-ask-order"},
		Short:   "Create an ask order",
		Long: fmt.Sprintf(`Create an ask order.

If --%s <seller> is provided, that is used as the seller.
If no --%s is provided, the --%s account address is used as the seller.
A seller is required.

The following flags are required:
  --%-19s e.g. 3
  --%-19s e.g. 10nhash
  --%-19s e.g. 10nhash

The following flags are optional:
  --%-44s e.g. 10nhash
  --%-44s (defaults to false if not provided)
  --%-44s e.g. 5D8759ED-0A1F-4952-9B9C-87A90099ED61
  --%-44s e.g. 10nhash
`,
			FlagSeller,
			FlagSeller, flags.FlagFrom,

			FlagMarket+" <market id>", // = 18 characters
			FlagAssets+" <assets>",
			FlagPrice+" <price>",

			FlagSettlementFee+" <seller settlement flat fee>", // = 43 characters
			FlagPartial,
			FlagExternalID+" <external id>",
			FlagCreationFee+" <creation fee>",
		),
		Example: fmt.Sprintf(`%s --%s %s --%s %s --%s %s --%s %s --%s --%s %s`,
			txCmdStart+" create-ask",
			flags.FlagFrom, ExampleAddr1,
			FlagAssets, "10atom",
			FlagPrice, "1000000000nhash",
			FlagMarket, "3",
			FlagPartial,
			FlagCreationFee, "500nhash",
		),
		Args: cobra.NoArgs,
		RunE: genericTxRunE(MakeMsgCreateAsk),
	}

	AddFlagsMsgCreateAsk(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxCreateBid creates the create-bid sub-command for the exchange tx command.
func CmdTxCreateBid() *cobra.Command {
	cmd := &cobra.Command{
		Use: fmt.Sprintf("create-bid {--%s|--%s} <buyer> --%s <market id> --%s <assets> --%s <price> "+
			"[--%s <buyer settlement fees>] [--%s] [--%s <external id>] [--%s <creation fee>]",
			flags.FlagFrom, FlagBuyer, FlagMarket, FlagAssets, FlagPrice,
			FlagSettlementFee, FlagPartial, FlagExternalID, FlagCreationFee,
		),
		Aliases: []string{"bid", "create-bid-order"},
		Short:   "Create a bid order",
		Long: fmt.Sprintf(`Create a bid order.

If --%s <buyer> is provided, that is used as the buyer.
If no --%s is provided, the --%s account address is used as the buyer.
A buyer is required.

The following flags are required:
  --%-19s e.g. 3
  --%-19s e.g. 10nhash
  --%-19s e.g. 10nhash

The following flags are optional:
  --%-39s e.g. 10nhash
  --%-39s (defaults to false if not provided)
  --%-39s e.g. 5D8759ED-0A1F-4952-9B9C-87A90099ED61
  --%-39s e.g. 10nhash
`,
			FlagBuyer,
			FlagBuyer, flags.FlagFrom,

			FlagMarket+" <market id>", // = 18 characters
			FlagAssets+" <assets>",
			FlagPrice+" <price>",

			FlagSettlementFee+" <buyer settlement fees>", // = 38 characters
			FlagPartial,
			FlagExternalID+" <external id>",
			FlagCreationFee+" <creation fee>",
		),
		Example: fmt.Sprintf(`%s --%s %s --%s %s --%s %s --%s %s --%s --%s %s`,
			txCmdStart+" create-bid",
			flags.FlagFrom, ExampleAddr1,
			FlagAssets, "10atom",
			FlagPrice, "1000000000nhash",
			FlagMarket, "3",
			FlagPartial,
			FlagCreationFee, "500nhash",
		),
		Args: cobra.NoArgs,
		RunE: genericTxRunE(MakeMsgCreateBid),
	}

	AddFlagsMsgCreateBid(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxCancelOrder creates the cancel-order sub-command for the exchange tx command.
func CmdTxCancelOrder() *cobra.Command {
	cmd := &cobra.Command{
		Use:     fmt.Sprintf("cancel-order {<order id>|--order <order id>} {--%s|--%s} <signer>", flags.FlagFrom, FlagSigner),
		Aliases: []string{"cancel"},
		Short:   "Cancel an order",
		Long: fmt.Sprintf(`Cancel an order.

The <order id> must be provided either as the first argument or using the --order flag, but not both.

If --%s <signer> is provided, that is used as the signer.
If no --%s is provided, the --%s account address is used as the signer.
A buyer is required.
`,
			FlagSigner,
			FlagSeller, flags.FlagFrom,
		),
		Example: fmt.Sprintf(`%[1]s 5 --%s %s
%[1]s --%s %s --%s 5
`,
			txCmdStart+" cancel-order",
			flags.FlagFrom, ExampleAddr1,
			FlagSigner, ExampleAddr1, FlagOrder,
		),
		Args: cobra.MaximumNArgs(1),
		RunE: genericTxRunE(MakeMsgCancelOrder),
	}

	AddFlagsMsgCancelOrder(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxFillBids creates the fill-bids sub-command for the exchange tx command.
func CmdTxFillBids() *cobra.Command {
	cmd := &cobra.Command{
		Use: fmt.Sprintf("fill-bids {--%s|--%s} <seller> --%s <market id> --%s <total assets> --%s <bid order ids> "+
			"[--%s <seller settlement flat fee>] [--%s <ask order creation fee>]",
			flags.FlagFrom, FlagSeller, FlagMarket, FlagAssets, FlagBids, FlagSettlementFee, FlagCreationFee,
		),
		Short: "Fill one or more bid orders",
		Long: fmt.Sprintf(`Fill one or more bid orders.

If --%s <seller> is provided, that is used as the seller.
If no --%s is provided, the --%s account address is used as the seller.
A seller is required.

The following flags are required:
  --%-22s e.g. 3
  --%-22s e.g. 10nhash
  --%-22s comma-separate each id; this flag can also be provided multiple times

The following flags are optional:
  --%-44s e.g. 10nhash
  --%-44s e.g. 10nhash
`,
			FlagSeller,
			FlagSeller, flags.FlagFrom,

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
			flags.FlagFrom, ExampleAddr1,
			FlagMarket, "3",
			FlagAssets, "10atom",
			FlagBids, "1,2,3", "1", "2,3",
		),
		Args: cobra.NoArgs,
		RunE: genericTxRunE(MakeMsgFillBids),
	}

	AddFlagsMsgFillBids(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxFillAsks creates the fill-asks sub-command for the exchange tx command.
func CmdTxFillAsks() *cobra.Command {
	cmd := &cobra.Command{
		Use: fmt.Sprintf("fill-asks {--%s|--%s} <buyer> --%s <market id> --%s <total price> --%s <ask order ids> "+
			"[--%s <buyer settlement fees>] [--%s <bid order creation fee>]",
			flags.FlagFrom, FlagBuyer, FlagMarket, FlagPrice, FlagAsks, FlagSettlementFee, FlagCreationFee,
		),
		Short: "Fill one or more ask orders",
		Long: fmt.Sprintf(`Fill one or more ask orders.

If --%s <buyer> is provided, that is used as the buyer.
If no --%s is provided, the --%s account address is used as the buyer.
A buyer is required.

The following flags are required:
  --%-21s e.g. 3
  --%-21s e.g. 10nhash
  --%-21s comma-separate each id; this flag can also be provided multiple times

The following flags are optional:
  --%-39s e.g. 10nhash
  --%-39s e.g. 10nhash
`,
			FlagBuyer,
			FlagBuyer, flags.FlagFrom,

			FlagMarket+" <market id>",
			FlagPrice+" <total price>",
			FlagAsks+" <ask order ids>", // = 20 characters.

			FlagSettlementFee+" <buyer settlement fees>", // = 38 characters.
			FlagCreationFee+" <bid order creation fee>",
		),
		Example: fmt.Sprintf(`%[1]s --%[2]s %[3]s --%[4]s %[5]s --%[6]s %[7]s --%[8]s %[9]s
%[1]s --%[2]s %[3]s --%[4]s %[5]s --%[6]s %[7]s --%[8]s %[10]s --%[8]s %[11]s
`,
			txCmdStart+" fill-asks",
			flags.FlagFrom, ExampleAddr1,
			FlagMarket, "3",
			FlagAssets, "10atom",
			FlagBids, "1,2,3", "1", "2,3",
		),
		Args: cobra.NoArgs,
		RunE: genericTxRunE(MakeMsgFillAsks),
	}

	AddFlagsMsgFillAsks(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxMarketSettle creates the market-settle sub-command for the exchange tx command.
func CmdTxMarketSettle() *cobra.Command {
	cmd := &cobra.Command{
		Use: fmt.Sprintf("market-settle {--%s|--%s} <admin> --%s <market id> --%s <ask order ids> --%s <bid order ids> [--%s]",
			flags.FlagFrom, FlagAdmin, FlagMarket, FlagAsks, FlagBids, FlagPartial),
		Aliases: []string{"settle"},
		Short:   "Settle some orders",
		Long: fmt.Sprintf(`Settle some orders.

If --%s <admin> is provided, that is used as the admin.
If no --%s is provided, but the --%s flag was, the governance module account is used as the admin.
Otherwise the --%s account address is used as the admin.
An admin is required.

The following flags are required:
  --%-21s e.g. 3
  --%-21s comma-separate each id; this flag can also be provided multiple times
  --%-21s comma-separate each id; this flag can also be provided multiple times

The --%s flag is optional.
`,
			FlagAdmin,
			FlagAdmin, FlagAuthority,
			flags.FlagFrom,

			FlagMarket+" <market id>",
			FlagAsks+" <ask order ids>", // = 20 characters
			FlagBids+" <bid order ids>",

			FlagPartial,
		),
		Example: fmt.Sprintf(`%s --%s 3 --%s 1,3,6 --%s 2,4,5 --%s %s`,
			txCmdStart+" market-settle",
			FlagMarket, FlagAsks, FlagBids, flags.FlagFrom, ExampleAddr1,
		),
		Args: cobra.NoArgs,
		RunE: genericTxRunE(MakeMsgMarketSettle),
	}

	AddFlagsMsgMarketSettle(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxMarketSetOrderExternalID creates the market-set-external-id sub-command for the exchange tx command.
func CmdTxMarketSetOrderExternalID() *cobra.Command {
	cmd := &cobra.Command{
		Use: fmt.Sprintf("market-set-external-id {--%s|--%s} <admin> --%s <market id> --%s <order id> --%s <new external id>",
			flags.FlagFrom, FlagAdmin, FlagMarket, FlagOrder, FlagExternalID,
		),
		Aliases: []string{"market-set-order-external-id", "set-external-id", "external-id",
			"mseid", "msei", "msoeid", "msoei", "seid", "sei"},
		Short: "Set an order's external id",
		Long: fmt.Sprintf(`Set an order's external id.

If --%s <admin> is provided, that is used as the admin.
If no --%s is provided, but the --%s flag was, the governance module account is used as the admin.
Otherwise the --%s account address is used as the admin.
An admin is required.

The following flags are required:
  --%-30s e.g. 3
  --%-30s e.g. 8

The %s flag is optional.
`,
			FlagAdmin,
			FlagAdmin, flags.FlagFrom,
			FlagAuthority,

			FlagMarket+" <market id>",
			FlagOrder+" <order id>",
			FlagExternalID+" <new external id>", // = 29 characters
		),
		Example: fmt.Sprintf(`%s --%s %s --%s %s --%s %s --%s %s`,
			txCmdStart+" set-order-id",
			FlagMarket, "3",
			FlagOrder, "8",
			FlagExternalID, "D333E919-B1FB-438B-96D0-9797B8DE418A",
			flags.FlagFrom, ExampleAddr1,
		),
		Args: cobra.NoArgs,
		RunE: genericTxRunE(MakeMsgMarketSetOrderExternalID),
	}

	AddFlagsMsgMarketSetOrderExternalID(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxMarketWithdraw creates the market-withdraw sub-command for the exchange tx command.
func CmdTxMarketWithdraw() *cobra.Command {
	cmd := &cobra.Command{
		Use: fmt.Sprintf("market-withdraw {--%s|--%s} <admin> --%s <market id> --%s <to address> --%s <amount>",
			flags.FlagFrom, FlagAdmin, FlagMarket, FlagTo, FlagAmount),
		Aliases: []string{"withdraw"},
		Short:   "Withdraw funds from a market account",
		Long: fmt.Sprintf(`Withdraw funds from a market account.

If --%s <admin> is provided, that is used as the admin.
If no --%s is provided, but the --%s flag was, the governance module account is used as the admin.
Otherwise the --%s account address is used as the admin.
An admin is required.

The following flags are required:
  --%-19s e.g. 3
  --%-19s e.g. %s
  --%-19s e.g. 10nhash
`,
			FlagAdmin,
			FlagAdmin, flags.FlagFrom,
			FlagAuthority,

			FlagMarket+" <market id>", // = 18 characters
			FlagTo+" <to address>", ExampleAddr1,
			FlagAmount+" <amount>",
		),
		Example: fmt.Sprintf(`%s --%s %s --%s %s --%s %s --%s %s`,
			txCmdStart+" market-withdraw",
			FlagMarket, "3",
			FlagAmount, "10nhash",
			FlagTo, ExampleAddr1,
			FlagAdmin, ExampleAddr2,
		),
		Args: cobra.NoArgs,
		RunE: genericTxRunE(MakeMsgMarketWithdraw),
	}

	AddFlagsMsgMarketWithdraw(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxMarketUpdateDetails creates the market-details sub-command for the exchange tx command.
func CmdTxMarketUpdateDetails() *cobra.Command {
	cmd := &cobra.Command{
		Use: fmt.Sprintf("market-details {--%s|--%s} <admin> --%s <market id> "+
			"[--%s <name>] [--%s <description>] [--%s <website url>] [--%s <icon uri>]",
			flags.FlagFrom, FlagAdmin, FlagMarket,
			FlagName, FlagDescription, FlagURL, FlagIcon,
		),
		Aliases: []string{"market-update-details", "update-details", "details"},
		Short:   "Update a market's details",
		Long: fmt.Sprintf(`Update a market's details.

All fields of a market's details will be updated.
If you omit an optional flag, it will be updated to an empty string.

If --%s <admin> is provided, that is used as the admin.
If no --%s is provided, but the --%s flag was, the governance module account is used as the admin.
Otherwise the --%s account address is used as the admin.
An admin is required.

The --%s flag is required.

The following flags are optional and default to an empty string:
  --%-26s (max %d characters)
  --%-26s (max %d characters)
  --%-26s (max %d characters)
  --%-26s (max %d characters)
`,
			FlagAdmin,
			FlagAdmin, flags.FlagFrom,
			FlagAuthority,

			FlagMarket, // = 18 characters

			FlagName+" <name>", exchange.MaxName,
			FlagDescription+" <description>", exchange.MaxDescription, // = 25 characters
			FlagURL+" <website url>", exchange.MaxWebsiteURL,
			FlagIcon+" <icon uri>", exchange.MaxIconURI,
		),
		Example: fmt.Sprintf(`%[1]s --%s %s --%s %s --%s %s --%s %s`,
			txCmdStart+" market-details",
			FlagAdmin, ExampleAddr1,
			FlagMarket, "3",
			FlagName, "'My Better Market'",
			cli.FlagWebsite, "`https://example.com'",
		),
		Args: cobra.NoArgs,
		RunE: genericTxRunE(MakeMsgMarketUpdateDetails),
	}

	AddFlagsMsgMarketUpdateDetails(cmd)
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
