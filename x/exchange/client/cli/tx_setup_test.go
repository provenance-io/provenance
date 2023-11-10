package cli_test

import (
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/provenance-io/provenance/x/exchange/client/cli"
)

func TestSetupCmdTxCreateAsk(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name:  "SetupCmdTxCreateAsk",
		setup: cli.SetupCmdTxCreateAsk,
		expFlags: []string{
			cli.FlagSeller, cli.FlagMarket, cli.FlagAssets, cli.FlagPrice,
			cli.FlagSettlementFee, cli.FlagPartial, cli.FlagExternalID, cli.FlagCreationFee,
			flags.FlagFrom, // not added by setup, but include so the annotation is checked.
		},
		expAnnotations: map[string]map[string][]string{
			flags.FlagFrom: {oneReq: {flags.FlagFrom + " " + cli.FlagSeller}},
			cli.FlagSeller: {oneReq: {flags.FlagFrom + " " + cli.FlagSeller}},
			cli.FlagMarket: {req: {"true"}},
			cli.FlagAssets: {req: {"true"}},
			cli.FlagPrice:  {req: {"true"}},
		},
		expInUse: []string{
			"--seller", "--market <market id>", "--assets <assets>", "--price <price>",
			"[--settlement-fee <seller settlement flat fee>]", "[--partial]",
			"[--external-id <external id>]", "[--creation-fee <creation fee>]",
			cli.ReqSignerDesc(cli.FlagSeller),
		},
	})
}

// TODO[1701]: func TestMakeMsgCreateAsk(t *testing.T)

func TestSetupCmdTxCreateBid(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name:  "SetupCmdTxCreateBid",
		setup: cli.SetupCmdTxCreateBid,
		expFlags: []string{
			cli.FlagBuyer, cli.FlagMarket, cli.FlagAssets, cli.FlagPrice,
			cli.FlagSettlementFee, cli.FlagPartial, cli.FlagExternalID, cli.FlagCreationFee,
			flags.FlagFrom, // not added by setup, but include so the annotation is checked.
		},
		expAnnotations: map[string]map[string][]string{
			flags.FlagFrom: {oneReq: {flags.FlagFrom + " " + cli.FlagBuyer}},
			cli.FlagBuyer:  {oneReq: {flags.FlagFrom + " " + cli.FlagBuyer}},
			cli.FlagMarket: {req: {"true"}},
			cli.FlagAssets: {req: {"true"}},
			cli.FlagPrice:  {req: {"true"}},
		},
		expInUse: []string{
			"--buyer", "--market <market id>", "--assets <assets>", "--price <price>",
			"[--settlement-fee <seller settlement flat fee>]", "[--partial]",
			"[--external-id <external id>]", "[--creation-fee <creation fee>]",
			cli.ReqSignerDesc(cli.FlagBuyer),
		},
	})
}

// TODO[1701]: func TestMakeMsgCreateBid(t *testing.T)

func TestSetupCmdTxCancelOrder(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name:  "SetupCmdTxCancelOrder",
		setup: cli.SetupCmdTxCancelOrder,
		expFlags: []string{
			cli.FlagSigner, cli.FlagOrder,
			flags.FlagFrom, // not added by setup, but include so the annotation is checked.
		},
		expAnnotations: map[string]map[string][]string{
			flags.FlagFrom: {oneReq: {flags.FlagFrom + " " + cli.FlagSigner}},
			cli.FlagSigner: {oneReq: {flags.FlagFrom + " " + cli.FlagSigner}},
		},
		expInUse: []string{
			"{<order id>|--order <order id>}",
			"{--from|--signer} <signer>",
			cli.ReqSignerDesc(cli.FlagSigner),
			"The <order id> must be provided either as the first argument or using the --order flag, but not both.",
		},
	})
}

// TODO[1701]: func TestMakeMsgCancelOrder(t *testing.T)

func TestSetupCmdTxFillBids(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name:  "SetupCmdTxFillBids",
		setup: cli.SetupCmdTxFillBids,
		expFlags: []string{
			cli.FlagSeller, cli.FlagMarket, cli.FlagAssets,
			cli.FlagBids, cli.FlagSettlementFee, cli.FlagCreationFee,
			flags.FlagFrom, // not added by setup, but include so the annotation is checked.
		},
		expAnnotations: map[string]map[string][]string{
			flags.FlagFrom: {oneReq: {flags.FlagFrom + " " + cli.FlagSeller}},
			cli.FlagSeller: {oneReq: {flags.FlagFrom + " " + cli.FlagSeller}},
			cli.FlagMarket: {req: {"true"}},
			cli.FlagAssets: {req: {"true"}},
			cli.FlagBids:   {req: {"true"}},
		},
		expInUse: []string{
			"{--from|--seller} <seller>", "--market <market id>", "--assets <total assets>",
			"--bids <bid order ids>", "[--settlement-fee <seller settlement flat fee>]",
			"[--creation-fee <ask order creation fee>]",
			cli.ReqSignerDesc(cli.FlagSeller),
			cli.RepeatableDesc,
		},
	})
}

// TODO[1701]: func TestMakeMsgFillBids(t *testing.T)

func TestSetupCmdTxFillAsks(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name:  "SetupCmdTxFillAsks",
		setup: cli.SetupCmdTxFillAsks,
		expFlags: []string{
			cli.FlagBuyer, cli.FlagMarket, cli.FlagPrice,
			cli.FlagAsks, cli.FlagSettlementFee, cli.FlagCreationFee,
			flags.FlagFrom, // not added by setup, but include so the annotation is checked.
		},
		expAnnotations: map[string]map[string][]string{
			flags.FlagFrom: {oneReq: {flags.FlagFrom + " " + cli.FlagBuyer}},
			cli.FlagBuyer:  {oneReq: {flags.FlagFrom + " " + cli.FlagBuyer}},
			cli.FlagMarket: {req: {"true"}},
			cli.FlagPrice:  {req: {"true"}},
			cli.FlagAsks:   {req: {"true"}},
		},
		expInUse: []string{
			"{--from|--buyer} <buyer>", "--market <market id>", "--price <total price>",
			"--asks <ask order ids>", "[--settlement-fee <buyer settlement fees>]",
			"[--creation-fee <bid order creation fee>]",
			cli.ReqSignerDesc(cli.FlagBuyer),
			cli.RepeatableDesc,
		},
	})
}

// TODO[1701]: func TestMakeMsgFillAsks(t *testing.T)

func TestSetupCmdTxMarketSettle(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name:  "SetupCmdTxMarketSettle",
		setup: cli.SetupCmdTxMarketSettle,
		expFlags: []string{
			cli.FlagMarket, cli.FlagAsks, cli.FlagBids, cli.FlagPartial,
		},
		expAnnotations: map[string]map[string][]string{
			cli.FlagMarket: {req: {"true"}},
			cli.FlagAsks:   {req: {"true"}},
			cli.FlagBids:   {req: {"true"}},
		},
		expInUse: []string{
			cli.ReqAdminUse, "--market <market id>",
			"--asks <ask order ids>", "--bids <bid order ids>",
			"[--partial]",
			cli.ReqAdminDesc, cli.RepeatableDesc,
		},
	})
}

// TODO[1701]: func TestMakeMsgMarketSettle(t *testing.T)

func TestSetupCmdTxMarketSetOrderExternalID(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name:  "SetupCmdTxMarketSetOrderExternalID",
		setup: cli.SetupCmdTxMarketSetOrderExternalID,
		expFlags: []string{
			cli.FlagAdmin, cli.FlagAuthority,
			cli.FlagMarket, cli.FlagOrder, cli.FlagExternalID,
			flags.FlagFrom, // not added by setup, but include so the annotation is checked.
		},
		expAnnotations: map[string]map[string][]string{
			flags.FlagFrom: {oneReq: {flags.FlagFrom + " " + cli.FlagAdmin + " " + cli.FlagAuthority}},
			cli.FlagAdmin: {
				mutExc: {cli.FlagAdmin + " " + cli.FlagAuthority},
				oneReq: {flags.FlagFrom + " " + cli.FlagAdmin + " " + cli.FlagAuthority},
			},
			cli.FlagAuthority: {
				mutExc: {cli.FlagAdmin + " " + cli.FlagAuthority},
				oneReq: {flags.FlagFrom + " " + cli.FlagAdmin + " " + cli.FlagAuthority},
			},
			cli.FlagMarket: {req: {"true"}},
			cli.FlagOrder:  {req: {"true"}},
		},
		expInUse: []string{
			cli.ReqAdminUse, "--market <market id>", "--order <order id>",
			"[--external-id <external id>]",
			cli.ReqAdminDesc,
		},
	})
}

// TODO[1701]: func TestMakeMsgMarketSetOrderExternalID(t *testing.T)

func TestSetupCmdTxMarketWithdraw(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name:  "SetupCmdTxMarketWithdraw",
		setup: cli.SetupCmdTxMarketWithdraw,
		expFlags: []string{
			cli.FlagAdmin, cli.FlagAuthority,
			cli.FlagMarket, cli.FlagTo, cli.FlagAmount,
			flags.FlagFrom, // not added by setup, but include so the annotation is checked.
		},
		expAnnotations: map[string]map[string][]string{
			flags.FlagFrom: {oneReq: {flags.FlagFrom + " " + cli.FlagAdmin + " " + cli.FlagAuthority}},
			cli.FlagAdmin: {
				mutExc: {cli.FlagAdmin + " " + cli.FlagAuthority},
				oneReq: {flags.FlagFrom + " " + cli.FlagAdmin + " " + cli.FlagAuthority},
			},
			cli.FlagAuthority: {
				mutExc: {cli.FlagAdmin + " " + cli.FlagAuthority},
				oneReq: {flags.FlagFrom + " " + cli.FlagAdmin + " " + cli.FlagAuthority},
			},
			cli.FlagMarket: {req: {"true"}},
			cli.FlagTo:     {req: {"true"}},
			cli.FlagAmount: {req: {"true"}},
		},
		expInUse: []string{
			cli.ReqAdminUse, "--market <market id>", "--to <to address>", "--amount <amount>",
			cli.ReqAdminDesc,
		},
	})
}

// TODO[1701]: func TestMakeMsgMarketWithdraw(t *testing.T)

// TODO[1701]: func TestAddFlagsMarketDetails(t *testing.T)

// TODO[1701]: func TestReadFlagsMarketDetails(t *testing.T)

func TestSetupCmdTxMarketUpdateDetails(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name:  "SetupCmdTxMarketUpdateDetails",
		setup: cli.SetupCmdTxMarketUpdateDetails,
		expFlags: []string{
			cli.FlagAdmin, cli.FlagAuthority,
			cli.FlagMarket,
			cli.FlagName, cli.FlagDescription, cli.FlagURL, cli.FlagIcon,
			flags.FlagFrom, // not added by setup, but include so the annotation is checked.
		},
		expAnnotations: map[string]map[string][]string{
			flags.FlagFrom: {oneReq: {flags.FlagFrom + " " + cli.FlagAdmin + " " + cli.FlagAuthority}},
			cli.FlagAdmin: {
				mutExc: {cli.FlagAdmin + " " + cli.FlagAuthority},
				oneReq: {flags.FlagFrom + " " + cli.FlagAdmin + " " + cli.FlagAuthority},
			},
			cli.FlagAuthority: {
				mutExc: {cli.FlagAdmin + " " + cli.FlagAuthority},
				oneReq: {flags.FlagFrom + " " + cli.FlagAdmin + " " + cli.FlagAuthority},
			},
			cli.FlagMarket: {req: {"true"}},
		},
		expInUse: []string{
			cli.ReqAdminUse, "--market <market id>",
			"[--name <name>]", "[--description <description>]", "[--url <website url>]", "[--icon <icon uri>]",
			cli.ReqAdminDesc,
			`All fields of a market's details will be updated.
If you omit an optional flag, that field will be updated to an empty string.`,
		},
	})
}

// TODO[1701]: func TestMakeMsgMarketUpdateDetails(t *testing.T)

func TestSetupCmdTxMarketUpdateEnabled(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name:  "SetupCmdTxMarketUpdateEnabled",
		setup: cli.SetupCmdTxMarketUpdateEnabled,
		expFlags: []string{
			cli.FlagAdmin, cli.FlagAuthority,
			cli.FlagMarket, cli.FlagEnable, cli.FlagDisable,
			flags.FlagFrom, // not added by setup, but include so the annotation is checked.
		},
		expAnnotations: map[string]map[string][]string{
			flags.FlagFrom: {oneReq: {flags.FlagFrom + " " + cli.FlagAdmin + " " + cli.FlagAuthority}},
			cli.FlagAdmin: {
				mutExc: {cli.FlagAdmin + " " + cli.FlagAuthority},
				oneReq: {flags.FlagFrom + " " + cli.FlagAdmin + " " + cli.FlagAuthority},
			},
			cli.FlagAuthority: {
				mutExc: {cli.FlagAdmin + " " + cli.FlagAuthority},
				oneReq: {flags.FlagFrom + " " + cli.FlagAdmin + " " + cli.FlagAuthority},
			},
			cli.FlagMarket: {req: {"true"}},
			cli.FlagEnable: {
				mutExc: {cli.FlagEnable + " " + cli.FlagDisable},
				oneReq: {cli.FlagEnable + " " + cli.FlagDisable},
			},
			cli.FlagDisable: {
				mutExc: {cli.FlagEnable + " " + cli.FlagDisable},
				oneReq: {cli.FlagEnable + " " + cli.FlagDisable},
			},
		},
		expInUse: []string{
			cli.ReqAdminUse, "--market <market id>", cli.ReqEnableDisableUse,
			cli.ReqAdminDesc, cli.ReqEnableDisableDesc,
		},
	})
}

// TODO[1701]: func TestMakeMsgMarketUpdateEnabled(t *testing.T)

func TestSetupCmdTxMarketUpdateUserSettle(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name:  "SetupCmdTxMarketUpdateUserSettle",
		setup: cli.SetupCmdTxMarketUpdateUserSettle,
		expFlags: []string{
			cli.FlagAdmin, cli.FlagAuthority,
			cli.FlagMarket, cli.FlagEnable, cli.FlagDisable,
			flags.FlagFrom, // not added by setup, but include so the annotation is checked.
		},
		expAnnotations: map[string]map[string][]string{
			flags.FlagFrom: {oneReq: {flags.FlagFrom + " " + cli.FlagAdmin + " " + cli.FlagAuthority}},
			cli.FlagAdmin: {
				mutExc: {cli.FlagAdmin + " " + cli.FlagAuthority},
				oneReq: {flags.FlagFrom + " " + cli.FlagAdmin + " " + cli.FlagAuthority},
			},
			cli.FlagAuthority: {
				mutExc: {cli.FlagAdmin + " " + cli.FlagAuthority},
				oneReq: {flags.FlagFrom + " " + cli.FlagAdmin + " " + cli.FlagAuthority},
			},
			cli.FlagMarket: {req: {"true"}},
			cli.FlagEnable: {
				mutExc: {cli.FlagEnable + " " + cli.FlagDisable},
				oneReq: {cli.FlagEnable + " " + cli.FlagDisable},
			},
			cli.FlagDisable: {
				mutExc: {cli.FlagEnable + " " + cli.FlagDisable},
				oneReq: {cli.FlagEnable + " " + cli.FlagDisable},
			},
		},
		expInUse: []string{
			cli.ReqAdminUse, "--market <market id>", cli.ReqEnableDisableUse,
			cli.ReqAdminDesc, cli.ReqEnableDisableDesc,
		},
	})
}

// TODO[1701]: func TestMakeMsgMarketUpdateUserSettle(t *testing.T)

func TestSetupCmdTxMarketManagePermissions(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name:  "SetupCmdTxMarketManagePermissions",
		setup: cli.SetupCmdTxMarketManagePermissions,
		expFlags: []string{
			cli.FlagAdmin, cli.FlagAuthority,
			cli.FlagMarket, cli.FlagRevokeAll, cli.FlagRevoke, cli.FlagGrant,
			flags.FlagFrom, // not added by setup, but include so the annotation is checked.
		},
		expAnnotations: map[string]map[string][]string{
			flags.FlagFrom: {oneReq: {flags.FlagFrom + " " + cli.FlagAdmin + " " + cli.FlagAuthority}},
			cli.FlagAdmin: {
				mutExc: {cli.FlagAdmin + " " + cli.FlagAuthority},
				oneReq: {flags.FlagFrom + " " + cli.FlagAdmin + " " + cli.FlagAuthority},
			},
			cli.FlagAuthority: {
				mutExc: {cli.FlagAdmin + " " + cli.FlagAuthority},
				oneReq: {flags.FlagFrom + " " + cli.FlagAdmin + " " + cli.FlagAuthority},
			},
			cli.FlagMarket:    {req: {"true"}},
			cli.FlagRevokeAll: {oneReq: {cli.FlagRevokeAll + " " + cli.FlagRevoke + " " + cli.FlagGrant}},
			cli.FlagRevoke:    {oneReq: {cli.FlagRevokeAll + " " + cli.FlagRevoke + " " + cli.FlagGrant}},
			cli.FlagGrant:     {oneReq: {cli.FlagRevokeAll + " " + cli.FlagRevoke + " " + cli.FlagGrant}},
		},
		expInUse: []string{
			cli.ReqAdminUse, "--market <market id>",
			"[--revoke-all <addresses>]", "[--revoke <access grants>]", "[--grant <access grants>]",
			cli.ReqAdminDesc, cli.RepeatableDesc, cli.AccessGrantsDesc,
		},
	})
}

// TODO[1701]: func TestMakeMsgMarketManagePermissions(t *testing.T)

func TestSetupCmdTxMarketManageReqAttrs(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name:  "SetupCmdTxMarketManageReqAttrs",
		setup: cli.SetupCmdTxMarketManageReqAttrs,
		expFlags: []string{
			cli.FlagAdmin, cli.FlagAuthority, cli.FlagMarket,
			cli.FlagAskAdd, cli.FlagAskRemove, cli.FlagBidAdd, cli.FlagBidRemove,
			flags.FlagFrom, // not added by setup, but include so the annotation is checked.
		},
		expAnnotations: map[string]map[string][]string{
			flags.FlagFrom: {oneReq: {flags.FlagFrom + " " + cli.FlagAdmin + " " + cli.FlagAuthority}},
			cli.FlagAdmin: {
				mutExc: {cli.FlagAdmin + " " + cli.FlagAuthority},
				oneReq: {flags.FlagFrom + " " + cli.FlagAdmin + " " + cli.FlagAuthority},
			},
			cli.FlagAuthority: {
				mutExc: {cli.FlagAdmin + " " + cli.FlagAuthority},
				oneReq: {flags.FlagFrom + " " + cli.FlagAdmin + " " + cli.FlagAuthority},
			},
			cli.FlagMarket:    {req: {"true"}},
			cli.FlagAskAdd:    {oneReq: {cli.FlagAskAdd + " " + cli.FlagAskRemove + " " + cli.FlagBidAdd + " " + cli.FlagBidRemove}},
			cli.FlagAskRemove: {oneReq: {cli.FlagAskAdd + " " + cli.FlagAskRemove + " " + cli.FlagBidAdd + " " + cli.FlagBidRemove}},
			cli.FlagBidAdd:    {oneReq: {cli.FlagAskAdd + " " + cli.FlagAskRemove + " " + cli.FlagBidAdd + " " + cli.FlagBidRemove}},
			cli.FlagBidRemove: {oneReq: {cli.FlagAskAdd + " " + cli.FlagAskRemove + " " + cli.FlagBidAdd + " " + cli.FlagBidRemove}},
		},
		expInUse: []string{
			cli.ReqAdminUse, "--market <market id>",
			"[--ask-add <attrs>]", "[--ask-remove <attrs>]",
			"[--bid-add <attrs>]", "[--bid-remove <attrs>]",
			cli.ReqAdminDesc, cli.RepeatableDesc,
		},
	})
}

// TODO[1701]: func TestMakeMsgMarketManageReqAttrs(t *testing.T)

func TestSetupCmdTxGovCreateMarket(t *testing.T) {
	tc := setupTestCase{
		name:  "SetupCmdTxGovCreateMarket",
		setup: cli.SetupCmdTxGovCreateMarket,
		expFlags: []string{
			cli.FlagAuthority,
			cli.FlagMarket, cli.FlagName, cli.FlagDescription, cli.FlagURL, cli.FlagIcon,
			cli.FlagCreateAsk, cli.FlagCreateBid,
			cli.FlagSellerFlat, cli.FlagSellerRatios, cli.FlagBuyerFlat, cli.FlagBuyerRatios,
			cli.FlagAcceptingOrders, cli.FlagAllowUserSettle, cli.FlagAccessGrants,
			cli.FlagReqAttrAsk, cli.FlagReqAttrBid,
		},
		expInUse: []string{
			"[--authority <authority>]", "[--market <market id>]",
			"[--name <name>]", "[--description <description>]", "[--url <website url>]", "[--icon <icon uri>]",
			"[--create-ask <coins>]", "[--create-bid <coins>]",
			"[--seller-flat <coins>]", "[--seller-ratios <fee ratios>]",
			"[--buyer-flat <coins>]", "[--buyer-ratios <fee ratios>]",
			"[--accepting-orders]", "[--allow-user-settle]",
			"[--access-grants <access grants>]",
			"[--req-attr-ask <attrs>]", "[--req-attr-bid <attrs>]",
			cli.AuthorityDesc, cli.RepeatableDesc, cli.AccessGrantsDesc, cli.FeeRatioDesc,
		},
	}

	oneReqFlags := []string{
		cli.FlagMarket, cli.FlagName, cli.FlagDescription, cli.FlagURL, cli.FlagIcon,
		cli.FlagCreateAsk, cli.FlagCreateBid,
		cli.FlagSellerFlat, cli.FlagSellerRatios, cli.FlagBuyerFlat, cli.FlagBuyerRatios,
		cli.FlagAcceptingOrders, cli.FlagAllowUserSettle, cli.FlagAccessGrants,
		cli.FlagReqAttrAsk, cli.FlagReqAttrBid,
	}
	oneReqVal := strings.Join(oneReqFlags, " ")
	if tc.expAnnotations == nil {
		tc.expAnnotations = make(map[string]map[string][]string)
	}
	for _, name := range oneReqFlags {
		if tc.expAnnotations[name] == nil {
			tc.expAnnotations[name] = make(map[string][]string)
		}
		tc.expAnnotations[name][oneReq] = []string{oneReqVal}
	}

	runSetupTestCase(t, tc)
}

// TODO[1701]: func TestMakeMsgGovCreateMarket(t *testing.T)

func TestSetupCmdTxGovManageFees(t *testing.T) {
	tc := setupTestCase{
		name:  "SetupCmdTxGovManageFees",
		setup: cli.SetupCmdTxGovManageFees,
		expFlags: []string{
			cli.FlagAuthority, cli.FlagMarket,
			cli.FlagAskAdd, cli.FlagAskRemove, cli.FlagBidAdd, cli.FlagBidRemove,
			cli.FlagSellerFlatAdd, cli.FlagSellerFlatRemove, cli.FlagSellerRatiosAdd, cli.FlagSellerRatiosRemove,
			cli.FlagBuyerFlatAdd, cli.FlagBuyerFlatRemove, cli.FlagBuyerRatiosAdd, cli.FlagBuyerRatiosRemove,
		},
		expAnnotations: map[string]map[string][]string{
			cli.FlagMarket: {req: {"true"}},
		},
		expInUse: []string{
			"--market <market id>", "[--authority <authority>]",
			"[--ask-add <coins>]", "[--ask-remove <coins>]",
			"[--bid-add <coins>]", "[--bid-remove <coins>]",
			"[--seller-flat-add <coins>]", "[--seller-flat-remove <coins>]",
			"[--seller-ratios-add <fee ratios>]", "[--seller-ratios-remove <fee ratios>]",
			"[--buyer-flat-add <coins>]", "[--buyer-flat-remove <coins>]",
			"[--buyer-ratios-add <fee ratios>]", "[--buyer-ratios-remove <fee ratios>]",
			cli.AuthorityDesc, cli.RepeatableDesc, cli.FeeRatioDesc,
		},
	}

	oneReqFlags := []string{
		cli.FlagAskAdd, cli.FlagAskRemove, cli.FlagBidAdd, cli.FlagBidRemove,
		cli.FlagSellerFlatAdd, cli.FlagSellerFlatRemove, cli.FlagSellerRatiosAdd, cli.FlagSellerRatiosRemove,
		cli.FlagBuyerFlatAdd, cli.FlagBuyerFlatRemove, cli.FlagBuyerRatiosAdd, cli.FlagBuyerRatiosRemove,
	}
	oneReqVal := strings.Join(oneReqFlags, " ")
	if tc.expAnnotations == nil {
		tc.expAnnotations = make(map[string]map[string][]string)
	}
	for _, name := range oneReqFlags {
		if tc.expAnnotations[name] == nil {
			tc.expAnnotations[name] = make(map[string][]string)
		}
		tc.expAnnotations[name][oneReq] = []string{oneReqVal}
	}

	runSetupTestCase(t, tc)
}

// TODO[1701]: func TestMakeMsgGovManageFees(t *testing.T)

func TestSetupCmdTxGovUpdateParams(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name:  "SetupCmdTxGovUpdateParams",
		setup: cli.SetupCmdTxGovUpdateParams,
		expFlags: []string{
			cli.FlagAuthority, cli.FlagDefault, cli.FlagSplit,
		},
		expAnnotations: map[string]map[string][]string{
			cli.FlagDefault: {req: {"true"}},
		},
		expInUse: []string{
			"--default <amount>", "[--split <splits>]", "[--authority <authority>]",
			cli.AuthorityDesc, cli.RepeatableDesc,
			`A <split> has the format "<denom>:<amount>".
An <amount> is in basis points and is limited to 0 to 10,000 (both inclusive).

Example <split>: nhash:500`,
		},
	})
}

// TODO[1701]: func TestMakeMsgGovUpdateParams(t *testing.T)
