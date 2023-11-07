package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
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
  --%-22s comma-separate each id; this flag can be provided multiple times

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
  --%-21s comma-separate each id; this flag can be provided multiple times

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
  --%-21s comma-separate each id; this flag can be provided multiple times
  --%-21s comma-separate each id; this flag can be provided multiple times

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

// CmdTxMarketUpdateEnabled creates the market-enabled sub-command for the exchange tx command.
func CmdTxMarketUpdateEnabled() *cobra.Command {
	cmd := &cobra.Command{
		Use: fmt.Sprintf("market-enabled {--%s|--%s} <admin> --%s <market id> {--%s|--%s}",
			flags.FlagFrom, FlagAdmin, FlagMarket, FlagEnable, FlagDisable,
		),
		Aliases: []string{"market-update-enabled", "update-enabled"},
		Short:   "Change whether a market is accepting orders",
		Long: fmt.Sprintf(`Change whether a market is accepting orders.

If --%s <admin> is provided, that is used as the admin.
If no --%s is provided, but the --%s flag was, the governance module account is used as the admin.
Otherwise the --%s account address is used as the admin.
An admin is required.

The --%s <market id> flag is required.
One of --%s or --%s must be provided, but not both.
`,
			FlagAdmin,
			FlagAdmin, flags.FlagFrom,
			FlagAuthority,

			FlagMarket,
			FlagEnable, FlagDisable,
		),
		Example: fmt.Sprintf(`%[1]s --%[2]s %[3]s --%[4]s %[5]s --%[6]s
%[1]s --%[2]s %[3]s --%[4]s %[5]s --%[7]s`,
			txCmdStart+" market-enabled",
			flags.FlagFrom, ExampleAddr1,
			FlagMarket, "3",
			FlagEnable, FlagDisable,
		),
		Args: cobra.NoArgs,
		RunE: genericTxRunE(MakeMsgMarketUpdateEnabled),
	}

	AddFlagsMsgMarketUpdateEnabled(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxMarketUpdateUserSettle creates the market-user-settle sub-command for the exchange tx command.
func CmdTxMarketUpdateUserSettle() *cobra.Command {
	cmd := &cobra.Command{
		Use: fmt.Sprintf("market-user-settle {--%s|--%s} <admin> --%s <market id> {--%s|--%s}",
			flags.FlagFrom, FlagAdmin, FlagMarket, FlagEnable, FlagDisable,
		),
		Aliases: []string{"market-update-user-settle", "update-user-settle"},
		Short:   "Change whether a market allows settlements initiated by users",
		Long: fmt.Sprintf(`Change whether a market allows settlements initiated by users.

If --%s <admin> is provided, that is used as the admin.
If no --%s is provided, but the --%s flag was, the governance module account is used as the admin.
Otherwise the --%s account address is used as the admin.
An admin is required.

The --%s <market id> flag is required.
One of --%s or --%s must be provided, but not both.
`,
			FlagAdmin,
			FlagAdmin, flags.FlagFrom,
			FlagAuthority,

			FlagMarket,
			FlagEnable, FlagDisable,
		),
		Example: fmt.Sprintf(`%[1]s --%[2]s %[3]s --%[4]s %[5]s --%[6]s
%[1]s --%[2]s %[3]s --%[4]s %[5]s --%[7]s`,
			txCmdStart+" market-user-settle",
			flags.FlagFrom, ExampleAddr1,
			FlagMarket, "3",
			FlagEnable, FlagDisable,
		),
		Args: cobra.NoArgs,
		RunE: genericTxRunE(MakeMsgMarketUpdateUserSettle),
	}

	AddFlagsMsgMarketUpdateUserSettle(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxMarketManagePermissions creates the market-permissions sub-command for the exchange tx command.
func CmdTxMarketManagePermissions() *cobra.Command {
	cmd := &cobra.Command{
		Use: fmt.Sprintf("market-permissions {--%s|--%s} <admin> --%s <market id> "+
			"[--%s <addresses>] [--%s <access grants>] [--%s <access grants>]",
			flags.FlagFrom, FlagAdmin, FlagMarket,
			FlagRevokeAll, FlagRevoke, FlagGrant,
		),
		Aliases: []string{"market-manage-permissions", "permissions", "market-perms", "market-manage-perms", "perms"},
		Short:   "Update the account permissions for a market",
		Long: fmt.Sprintf(`Update the account permissions for a market

If --%s <admin> is provided, that is used as the admin.
If no --%s is provided, but the --%s flag was, the governance module account is used as the admin.
Otherwise the --%s account address is used as the admin.
An admin is required.

The --%s <market id> flag is required.

The following flags are optional (but at least one must be provided):
  --%-23s Separate each address with commas. This flag can be provided multiple times.
  --%-23s Separate each <access grant> with commas. This flag can be provided multiple times.
  --%-23s Separate each <access grant> with commas. This flag can be provided multiple times.

An <access grant> has the format "<address>:<permissions>"
In <permissions>, separate each entry with a + (plus), - (dash), or . (period).
An <access grant> of "<address>:all" will have all of the permissions.

Example <access grant>: %s:settle+update

Valid permissions entries: %s
The full Permission enum names are also valid.
`,
			FlagAdmin,
			FlagAdmin, flags.FlagFrom,
			FlagAuthority,

			FlagMarket,

			FlagRevokeAll+" <addresses>",
			FlagRevoke+" <access grants>", // = 22 characters
			FlagGrant+" <access grants>",

			ExampleAddr1,
			SimplePerms(),
		),
		Example: fmt.Sprintf(`%[1]s --%s %s --%s 3 --%s %s
%[1]s --%s %s --%s 3 --%s %s:settle+set-ids %s %s:settle+set-ids
%[1]s --%s --%s 3 --%s %s:all --generate-only
`,
			txCmdStart+" market-permissions",
			flags.FlagFrom, ExampleAddr1, FlagMarket, FlagRevokeAll, ExampleAddr2,
			FlagAdmin, ExampleAddr1, FlagMarket, FlagRevoke, ExampleAddr3, FlagGrant, ExampleAddr4,
			FlagAuthority, FlagMarket, FlagGrant, ExampleAddr5,
		),
		Args: cobra.NoArgs,
		RunE: genericTxRunE(MakeMsgMarketManagePermissions),
	}

	AddFlagsMsgMarketManagePermissions(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxMarketManageReqAttrs creates the market-req-attrs sub-command for the exchange tx command.
func CmdTxMarketManageReqAttrs() *cobra.Command {
	cmd := &cobra.Command{
		Use: fmt.Sprintf("market-req-attrs {--%s|--%s} <admin> --%s <market id> "+
			"[--%s <attrs>] [--%s <attrs>] [--%s <attrs>] [--%s <attrs>]",
			flags.FlagFrom, FlagAdmin, FlagMarket,
			FlagAskAdd, FlagAskRemove, FlagBidAdd, FlagBidRemove,
		),
		Aliases: []string{"market-manage-req-attrs", "manage-req-attrs", "req-attrs", "market-required-attributes",
			"market-manage-required-attributes", "manage-required-attributes", "required-attributes",
		},
		Short: "Manage the attributes required to create orders in a market",
		Long: fmt.Sprintf(`Manage the attributes required to create orders in a market.

If --%s <admin> is provided, that is used as the admin.
If no --%s is provided, but the --%s flag was, the governance module account is used as the admin.
Otherwise the --%s account address is used as the admin.
An admin is required.

The --%s <market id> flag is required.

The following flags are optional (but at least one must be provided):
  --%-19s Separate each entry with commas. This flag can be provided multiple times.
  --%-19s Separate each entry with commas. This flag can be provided multiple times.
  --%-19s Separate each entry with commas. This flag can be provided multiple times.
  --%-19s Separate each entry with commas. This flag can be provided multiple times.
`,
			FlagAdmin,
			FlagAdmin, flags.FlagFrom,
			FlagAuthority,

			FlagMarket,

			FlagAskAdd+" <attrs>", // = 18 characters
			FlagAskRemove+" <attrs>",
			FlagBidAdd+" <attrs>",
			FlagBidRemove+" <attrs>",
		),
		Example: fmt.Sprintf(`%[1]s --%s %s --%s 3 --%s 'ask.example' --%s '*.example'
%[1]s --%s %s --%s 3 --%s '*.buyer.example' --%s 'bid.example'`,
			txCmdStart+" market-req-attrs",
			flags.FlagFrom, ExampleAddr1, FlagMarket, FlagAskRemove, FlagAskAdd,
			FlagAdmin, ExampleAddr1, FlagMarket, FlagBidAdd, FlagBidRemove,
		),
		Args: cobra.NoArgs,
		RunE: genericTxRunE(MakeMsgMarketManageReqAttrs),
	}

	AddFlagsMsgMarketManageReqAttrs(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdTxGovCreateMarket creates the create-market sub-command for the exchange tx command.
func CmdTxGovCreateMarket() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-market",
		Aliases: []string{"gov-create-market"},
		Short:   "Submit a governance proposal to create a market",
		Long: fmt.Sprintf(`Submit a governance proposal to create a market.

The following flags define the Market fields and are all optional:
  --%-35s %s
  --%-35s %s
  --%-35s %s
  --%-35s %s
  --%-35s %s
  --%-35s %s
  --%-35s %s
  --%-35s %s
  --%-35s %s
  --%-35s %s
  --%-35s %s
  --%-35s %s
  --%-35s %s
  --%-35s %s
  --%-35s %s
  --%-35s %s

If a flag is repeatable, multiple entries can be separated by commas and/or the flag can be provided multiple times.

An <access grant> has the format "<address>:<permissions>"
In <permissions>, separate each entry with a + (plus), - (dash), or . (period).
An <access grant> of "<address>:all" will have all of the permissions.
Example <access grant>: %s:settle+update

Valid permissions entries: %s
The full Permission enum names are also valid.
`,
			FlagMarket+" <market id>", "",
			FlagName+" <name>", fmt.Sprintf("Max %d characters", exchange.MaxName),
			FlagDescription+" <description>", fmt.Sprintf("Max %d characters", exchange.MaxDescription),
			FlagURL+" <website url>", fmt.Sprintf("Max %d characters", exchange.MaxWebsiteURL),
			FlagIcon+" <icon uri>", fmt.Sprintf("Max %d characters", exchange.MaxIconURI),
			FlagCreateAsk+" <flat fee options>", "Repeatable Coin strings, e.g. 10nhash",
			FlagCreateBid+" <flat fee options>", "Repeatable Coin strings, e.g. 10nhash",
			FlagSellerFlat+" <flat fee options>", "Repeatable Coin strings, e.g. 10nhash",
			FlagSellerRatios+" <fee ratios>", "Repeatable FeeRatio strings, e.g. 100nhash:1nhash",
			FlagBuyerFlat+" <flat fee options>", "Repeatable Coin strings, e.g. 10nhash",
			FlagBuyerRatios+" <fee ratios>", "Repeatable FeeRatio strings, e.g. 100nhash:1nhash",
			FlagAcceptingOrders, "",
			FlagAllowUserSettle, "",
			FlagAccessGrants+" <access grants>", fmt.Sprintf("Repeatable <access grant> strings, e.g. %s:settle", ExampleAddr1),
			FlagReqAttrAsk+" <required attributes>", "Repeatable", // = 34 characters
			FlagReqAttrBid+" <required attributes>", "Repeatable",

			ExampleAddr1, SimplePerms(),
		),
		Args: cobra.NoArgs,
		RunE: govTxRunE(MakeMsgGovCreateMarket),
	}

	govcli.AddGovPropFlagsToCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)
	AddFlagsMsgGovCreateMarket(cmd)
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
