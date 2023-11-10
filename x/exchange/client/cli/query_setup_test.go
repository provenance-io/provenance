package cli_test

import (
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/provenance-io/provenance/x/exchange/client/cli"
)

var exampleStart = version.AppName + " query exchange dummy"

func TestSetupCmdQueryOrderFeeCalc(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name:  "SetupCmdQueryOrderFeeCalc",
		setup: cli.SetupCmdQueryOrderFeeCalc,
		expFlags: []string{
			cli.FlagAsk, cli.FlagBid, cli.FlagMarket,
			cli.FlagSeller, cli.FlagBuyer, cli.FlagAssets, cli.FlagPrice,
			cli.FlagSettlementFee, cli.FlagPartial, cli.FlagExternalID,
		},
		expAnnotations: map[string]map[string][]string{
			cli.FlagAsk: {
				mutExc: {cli.FlagAsk + " " + cli.FlagBid, cli.FlagAsk + " " + cli.FlagBuyer},
				oneReq: {cli.FlagAsk + " " + cli.FlagBid},
			},
			cli.FlagBid: {
				mutExc: {cli.FlagAsk + " " + cli.FlagBid, cli.FlagBid + " " + cli.FlagSeller},
				oneReq: {cli.FlagAsk + " " + cli.FlagBid},
			},
			cli.FlagBuyer:  {mutExc: {cli.FlagBuyer + " " + cli.FlagSeller, cli.FlagAsk + " " + cli.FlagBuyer}},
			cli.FlagSeller: {mutExc: {cli.FlagBuyer + " " + cli.FlagSeller, cli.FlagBid + " " + cli.FlagSeller}},
			cli.FlagMarket: {req: {"true"}},
			cli.FlagPrice:  {req: {"true"}},
		},
		expInUse: []string{
			cli.ReqAskBidUse, "--market <market id>", "--price <price>",
			cli.ReqAskBidDesc,
		},
		expExamples: []string{
			exampleStart + " --ask --market 3 --price 10nhash",
			exampleStart + " --bid --market 3 --price 10nhash",
		},
	})
}

// TODO[1701]: func TestMakeQueryOrderFeeCalc(t *testing.T)

func TestSetupCmdQueryGetOrder(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name:     "SetupCmdQueryGetOrder",
		setup:    cli.SetupCmdQueryGetOrder,
		expFlags: []string{cli.FlagOrder},
		expInUse: []string{
			"{<order id>|--order <order id>}",
			"An <order id> is required as either an arg or flag, but not both.",
		},
		expExamples: []string{
			exampleStart + " 8",
			exampleStart + " --order 8",
		},
	})
}

// TODO[1701]: func TestMakeQueryGetOrder(t *testing.T)

func TestSetupCmdQueryGetOrderByExternalID(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name:  "SetupCmdQueryGetOrderByExternalID",
		setup: cli.SetupCmdQueryGetOrderByExternalID,
		expFlags: []string{
			cli.FlagMarket, cli.FlagExternalID,
		},
		expAnnotations: map[string]map[string][]string{
			cli.FlagMarket:     {req: {"true"}},
			cli.FlagExternalID: {req: {"true"}},
		},
		expInUse: []string{
			"--market <market id>", "--external-id <external id>",
		},
		expExamples: []string{
			exampleStart + " --market 3 --external-id 12BD2C9C-9641-4370-A503-802CD7079CAA",
		},
	})
}

// TODO[1701]: func TestMakeQueryGetOrderByExternalID(t *testing.T)

func TestSetupCmdQueryGetMarketOrders(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name:  "SetupCmdQueryGetMarketOrders",
		setup: cli.SetupCmdQueryGetMarketOrders,
		expFlags: []string{
			flags.FlagPage, flags.FlagPageKey, flags.FlagOffset,
			flags.FlagLimit, flags.FlagCountTotal, flags.FlagReverse,
			cli.FlagMarket, cli.FlagAsks, cli.FlagBids, cli.FlagOrder,
		},
		expAnnotations: map[string]map[string][]string{
			cli.FlagAsks: {mutExc: {cli.FlagAsks + " " + cli.FlagBids}},
			cli.FlagBids: {mutExc: {cli.FlagAsks + " " + cli.FlagBids}},
		},
		expInUse: []string{
			"{<market id>|--market <market id>}", cli.OptAsksBidsUse,
			"[--order <after order id>", "[pagination flags]",
			"A <market id> is required as either an arg or flag, but not both.",
			cli.OptAsksBidsDesc,
		},
		expExamples: []string{
			exampleStart + " 3 --asks",
			exampleStart + " --market 1 --order 15 --limit 10",
		},
	})
}

// TODO[1701]: func TestMakeQueryGetMarketOrders(t *testing.T)

func TestSetupCmdQueryGetOwnerOrders(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name:  "SetupCmdQueryGetOwnerOrders",
		setup: cli.SetupCmdQueryGetOwnerOrders,
		expFlags: []string{
			flags.FlagPage, flags.FlagPageKey, flags.FlagOffset,
			flags.FlagLimit, flags.FlagCountTotal, flags.FlagReverse,
			cli.FlagOwner, cli.FlagAsks, cli.FlagBids, cli.FlagOrder,
		},
		expAnnotations: map[string]map[string][]string{
			cli.FlagAsks: {mutExc: {cli.FlagAsks + " " + cli.FlagBids}},
			cli.FlagBids: {mutExc: {cli.FlagAsks + " " + cli.FlagBids}},
		},
		expInUse: []string{
			"{<owner>|--owner <owner>}", cli.OptAsksBidsUse,
			"[--order <after order id>", "[pagination flags]",
			"An <owner> is required as either an arg or flag, but not both.",
			cli.OptAsksBidsDesc,
		},
		expExamples: []string{
			exampleStart + " " + cli.ExampleAddr + " --bids",
			exampleStart + " --owner " + cli.ExampleAddr + " --asks --order 15 --limit 10",
		},
	})
}

// TODO[1701]: func TestMakeQueryGetOwnerOrders(t *testing.T)

func TestSetupCmdQueryGetAssetOrders(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name:  "SetupCmdQueryGetAssetOrders",
		setup: cli.SetupCmdQueryGetAssetOrders,
		expFlags: []string{
			flags.FlagPage, flags.FlagPageKey, flags.FlagOffset,
			flags.FlagLimit, flags.FlagCountTotal, flags.FlagReverse,
			cli.FlagDenom, cli.FlagAsks, cli.FlagBids, cli.FlagOrder,
		},
		expAnnotations: map[string]map[string][]string{
			cli.FlagAsks: {mutExc: {cli.FlagAsks + " " + cli.FlagBids}},
			cli.FlagBids: {mutExc: {cli.FlagAsks + " " + cli.FlagBids}},
		},
		expInUse: []string{
			"{<asset>|--denom <asset>}", cli.OptAsksBidsUse,
			"[--order <after order id>", "[pagination flags]",
			"An <asset> is required as either an arg or flag, but not both.",
			cli.OptAsksBidsDesc,
		},
		expExamples: []string{
			exampleStart + " nhash --asks",
			exampleStart + " --denom nhash --order 15 --limit 10",
		},
	})
}

// TODO[1701]: func TestMakeQueryGetAssetOrders(t *testing.T)

func TestSetupCmdQueryGetAllOrders(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name:  "SetupCmdQueryGetAllOrders",
		setup: cli.SetupCmdQueryGetAllOrders,
		expFlags: []string{
			flags.FlagPage, flags.FlagPageKey, flags.FlagOffset,
			flags.FlagLimit, flags.FlagCountTotal, flags.FlagReverse,
		},
		expInUse: []string{"[pagination flags]"},
		expExamples: []string{
			exampleStart + " --limit 10",
			exampleStart + " --reverse",
		},
	})
}

// TODO[1701]: func TestMakeQueryGetAllOrders(t *testing.T)

func TestSetupCmdQueryGetMarket(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name:     "SetupCmdQueryGetMarket",
		setup:    cli.SetupCmdQueryGetMarket,
		expFlags: []string{cli.FlagMarket},
		expInUse: []string{
			"{<market id>|--market <market id>}",
			"A <market id> is required as either an arg or flag, but not both.",
		},
		expExamples: []string{
			exampleStart + " 3",
			exampleStart + " --market 1",
		},
	})
}

// TODO[1701]: func TestMakeQueryGetMarket(t *testing.T)

func TestSetupCmdQueryGetAllMarkets(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name:  "SetupCmdQueryGetAllMarkets",
		setup: cli.SetupCmdQueryGetAllMarkets,
		expFlags: []string{
			flags.FlagPage, flags.FlagPageKey, flags.FlagOffset,
			flags.FlagLimit, flags.FlagCountTotal, flags.FlagReverse,
		},
		expInUse: []string{"[pagination flags]"},
		expExamples: []string{
			exampleStart + " --limit 10",
			exampleStart + " --reverse",
		},
	})
}

// TODO[1701]: func TestMakeQueryGetAllMarkets(t *testing.T)

func TestSetupCmdQueryParams(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name:        "SetupCmdQueryParams",
		setup:       cli.SetupCmdQueryParams,
		expExamples: []string{exampleStart},
	})
}

// TODO[1701]: func TestMakeQueryParams(t *testing.T)

func TestSetupCmdQueryValidateCreateMarket(t *testing.T) {
	tc := setupTestCase{
		name:  "SetupCmdQueryValidateCreateMarket",
		setup: cli.SetupCmdQueryValidateCreateMarket,
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

// TODO[1701]: func TestMakeQueryValidateCreateMarket(t *testing.T)

func TestSetupCmdQueryValidateMarket(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name:     "SetupCmdQueryValidateMarket",
		setup:    cli.SetupCmdQueryValidateMarket,
		expFlags: []string{cli.FlagMarket},
		expInUse: []string{
			"{<market id>|--market <market id>}",
			"A <market id> is required as either an arg or flag, but not both.",
		},
		expExamples: []string{
			exampleStart + " 3",
			exampleStart + " --market 1",
		},
	})
}

// TODO[1701]: func TestMakeQueryValidateMarket(t *testing.T)

func TestSetupCmdQueryValidateManageFees(t *testing.T) {
	tc := setupTestCase{
		name:  "SetupCmdQueryValidateManageFees",
		setup: cli.SetupCmdQueryValidateManageFees,
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

// TODO[1701]: func TestMakeQueryValidateManageFees(t *testing.T)
