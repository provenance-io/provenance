package cli_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/exchange"
	"github.com/provenance-io/provenance/x/exchange/client/cli"
)

var exampleStart = version.AppName + " query exchange dummy"

// queryMakerTestDef is the definition of a query maker func to be tested.
//
// R is the type that is returned by the maker.
type queryMakerTestDef[R any] struct {
	// makerName is the name of the maker func being tested.
	makerName string
	// maker is the query request maker func being tested.
	maker func(flagSet *pflag.FlagSet, args []string) (*R, error)
	// setup is the command setup func that sets up a command so it has what's needed by the maker.
	setup func(cmd *cobra.Command)
}

// queryMakerTestCase is a test case for a query maker func.
//
// R is the type that is returned by the maker.
type queryMakerTestCase[R any] struct {
	// name is a name for this test case.
	name string
	// flags are the strings the give to FlagSet before it's provided to the maker.
	flags []string
	// args are the strings to supply as args to the maker.
	args []string
	// expReq is the expected result of the maker.
	expReq *R
	// expErr is the expected error string. An empty string indicates the error should be nil.
	expErr string
}

// runQueryMakerTest runs a test case for a query maker func.
//
// R is the type that is returned by the maker.
func runQueryMakerTest[R any](t *testing.T, td queryMakerTestDef[R], tc queryMakerTestCase[R]) {
	cmd := &cobra.Command{
		Use: "dummy",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("this dummy command should not have been executed")
		},
	}
	td.setup(cmd)

	err := cmd.Flags().Parse(tc.flags)
	require.NoError(t, err, "cmd.Flags().Parse(%q)", tc.flags)

	var req *R
	testFunc := func() {
		req, err = td.maker(cmd.Flags(), tc.args)
	}
	require.NotPanics(t, testFunc, td.makerName)
	assertions.AssertErrorValue(t, err, tc.expErr, "%s error", td.makerName)
	assert.Equal(t, tc.expReq, req, "%s request", td.makerName)
}

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
			cli.FlagMarket: {required: {"true"}},
			cli.FlagPrice:  {required: {"true"}},
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

func TestMakeQueryOrderFeeCalc(t *testing.T) {
	td := queryMakerTestDef[exchange.QueryOrderFeeCalcRequest]{
		makerName: "MakeQueryOrderFeeCalc",
		maker:     cli.MakeQueryOrderFeeCalc,
		setup:     cli.SetupCmdQueryOrderFeeCalc,
	}

	fillerCoin := sdk.Coin{Denom: "filler", Amount: sdkmath.NewInt(0)}

	tests := []queryMakerTestCase[exchange.QueryOrderFeeCalcRequest]{
		{
			name:  "no price and bad settlement fees",
			flags: []string{"--market", "3", "--bid", "--settlement-fee", "oops"},
			expReq: &exchange.QueryOrderFeeCalcRequest{
				BidOrder: &exchange.BidOrder{
					MarketId: 3,
					Assets:   fillerCoin,
				},
			},
			expErr: joinErrs(
				"missing required --price flag",
				"error parsing --settlement-fee as coins: invalid coin expression: \"oops\"",
			),
		},
		{
			name:  "ask with two settlement fees",
			flags: []string{"--market", "2", "--ask", "--settlement-fee", "10apple,3banana", "--price", "18pear"},
			expReq: &exchange.QueryOrderFeeCalcRequest{
				AskOrder: &exchange.AskOrder{
					MarketId:                2,
					Assets:                  fillerCoin,
					Price:                   sdk.NewInt64Coin("pear", 18),
					SellerSettlementFlatFee: &sdk.Coin{Denom: "apple", Amount: sdkmath.NewInt(10)},
				},
			},
			expErr: "only one settlement fee coin is allowed for ask orders",
		},
		{
			name:   "bad coins",
			flags:  []string{"--market", "11", "--price", "-3badcoin", "--assets", "noamt", "--settlement-fee", "88x"},
			expReq: &exchange.QueryOrderFeeCalcRequest{},
			expErr: joinErrs(
				"error parsing --assets as a coin: invalid coin expression: \"noamt\"",
				"error parsing --price as a coin: invalid coin expression: \"-3badcoin\"",
				"error parsing --settlement-fee as coins: invalid coin expression: \"88x\""),
		},
		{
			name:  "minimal ask",
			flags: []string{"--ask", "--market", "51", "--price", "66prune"},
			expReq: &exchange.QueryOrderFeeCalcRequest{
				AskOrder: &exchange.AskOrder{
					MarketId: 51,
					Assets:   fillerCoin,
					Price:    sdk.NewInt64Coin("prune", 66),
				},
			},
		},
		{
			name: "full ask",
			flags: []string{
				"--ask", "--seller", "someaddr",
				"--assets", "15apple", "--price", "60plum", "--market", "8",
				"--partial", "--external-id", "outsideid",
				"--settlement-fee", "5fig",
			},
			expReq: &exchange.QueryOrderFeeCalcRequest{
				AskOrder: &exchange.AskOrder{
					MarketId:                8,
					Seller:                  "someaddr",
					Assets:                  sdk.NewInt64Coin("apple", 15),
					Price:                   sdk.NewInt64Coin("plum", 60),
					SellerSettlementFlatFee: &sdk.Coin{Denom: "fig", Amount: sdkmath.NewInt(5)},
					AllowPartial:            true,
					ExternalId:              "outsideid",
				},
			},
		},
		{
			name:  "minimal bid",
			flags: []string{"--bid", "--market", "51", "--price", "66prune"},
			expReq: &exchange.QueryOrderFeeCalcRequest{
				BidOrder: &exchange.BidOrder{
					MarketId: 51,
					Assets:   fillerCoin,
					Price:    sdk.NewInt64Coin("prune", 66),
				},
			},
		},
		{
			name: "full bid",
			flags: []string{
				"--bid", "--buyer", "someaddr",
				"--assets", "15apple", "--price", "60plum", "--market", "8",
				"--partial", "--external-id", "outsideid",
				"--settlement-fee", "5fig",
			},
			expReq: &exchange.QueryOrderFeeCalcRequest{
				BidOrder: &exchange.BidOrder{
					MarketId:            8,
					Buyer:               "someaddr",
					Assets:              sdk.NewInt64Coin("apple", 15),
					Price:               sdk.NewInt64Coin("plum", 60),
					BuyerSettlementFees: sdk.Coins{sdk.Coin{Denom: "fig", Amount: sdkmath.NewInt(5)}},
					AllowPartial:        true,
					ExternalId:          "outsideid",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runQueryMakerTest(t, td, tc)
		})
	}
}

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

func TestMakeQueryGetOrder(t *testing.T) {
	td := queryMakerTestDef[exchange.QueryGetOrderRequest]{
		makerName: "MakeQueryGetOrder",
		maker:     cli.MakeQueryGetOrder,
		setup:     cli.SetupCmdQueryGetOrder,
	}

	tests := []queryMakerTestCase[exchange.QueryGetOrderRequest]{
		{
			name:   "no order id",
			expReq: &exchange.QueryGetOrderRequest{},
			expErr: "no <order id> provided",
		},
		{
			name:   "just order flag",
			flags:  []string{"--order", "15"},
			expReq: &exchange.QueryGetOrderRequest{OrderId: 15},
		},
		{
			name:   "just order id arg",
			args:   []string{"83"},
			expReq: &exchange.QueryGetOrderRequest{OrderId: 83},
		},
		{
			name:   "both order flag and arg",
			flags:  []string{"--order", "15"},
			args:   []string{"83"},
			expReq: &exchange.QueryGetOrderRequest{},
			expErr: "cannot provide <order id> as both an arg (\"83\") and flag (--order 15)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runQueryMakerTest(t, td, tc)
		})
	}
}

func TestSetupCmdQueryGetOrderByExternalID(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name:  "SetupCmdQueryGetOrderByExternalID",
		setup: cli.SetupCmdQueryGetOrderByExternalID,
		expFlags: []string{
			cli.FlagMarket, cli.FlagExternalID,
		},
		expAnnotations: map[string]map[string][]string{
			cli.FlagMarket:     {required: {"true"}},
			cli.FlagExternalID: {required: {"true"}},
		},
		expInUse: []string{
			"--market <market id>", "--external-id <external id>",
		},
		expExamples: []string{
			exampleStart + " --market 3 --external-id 12BD2C9C-9641-4370-A503-802CD7079CAA",
		},
	})
}

func TestMakeQueryGetOrderByExternalID(t *testing.T) {
	td := queryMakerTestDef[exchange.QueryGetOrderByExternalIDRequest]{
		makerName: "MakeQueryGetOrderByExternalID",
		maker:     cli.MakeQueryGetOrderByExternalID,
		setup:     cli.SetupCmdQueryGetOrderByExternalID,
	}

	tests := []queryMakerTestCase[exchange.QueryGetOrderByExternalIDRequest]{
		{
			name:  "normal use",
			flags: []string{"--external-id", "myid", "--market", "15"},
			expReq: &exchange.QueryGetOrderByExternalIDRequest{
				MarketId:   15,
				ExternalId: "myid",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runQueryMakerTest(t, td, tc)
		})
	}
}

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

func TestMakeQueryGetMarketOrders(t *testing.T) {
	td := queryMakerTestDef[exchange.QueryGetMarketOrdersRequest]{
		makerName: "MakeQueryGetMarketOrders",
		maker:     cli.MakeQueryGetMarketOrders,
		setup:     cli.SetupCmdQueryGetMarketOrders,
	}

	defaultPageReq := &query.PageRequest{
		Key:   []byte{},
		Limit: 100,
	}
	tests := []queryMakerTestCase[exchange.QueryGetMarketOrdersRequest]{
		{
			name: "no market id",
			expReq: &exchange.QueryGetMarketOrdersRequest{
				Pagination: defaultPageReq,
			},
			expErr: "no <market id> provided",
		},
		{
			name:  "just market id flag",
			flags: []string{"--market", "1"},
			expReq: &exchange.QueryGetMarketOrdersRequest{
				MarketId:   1,
				Pagination: defaultPageReq,
			},
		},
		{
			name: "just market id arg",
			args: []string{"1"},
			expReq: &exchange.QueryGetMarketOrdersRequest{
				MarketId:   1,
				Pagination: defaultPageReq,
			},
		},
		{
			name:  "both market id flag and arg",
			flags: []string{"--market", "1"},
			args:  []string{"1"},
			expReq: &exchange.QueryGetMarketOrdersRequest{
				Pagination: defaultPageReq,
			},
			expErr: "cannot provide <market id> as both an arg (\"1\") and flag (--market 1)",
		},
		{
			name: "all opts asks",
			flags: []string{
				"--asks", "--order", "12", "--limit", "63",
				"--offset", "42", "--count-total",
			},
			args: []string{"7"},
			expReq: &exchange.QueryGetMarketOrdersRequest{
				MarketId:     7,
				OrderType:    "ask",
				AfterOrderId: 12,
				Pagination: &query.PageRequest{
					Key:        []byte{},
					Offset:     42,
					Limit:      63,
					CountTotal: true,
				},
			},
		},
		{
			name: "all opts bids",
			flags: []string{
				"--order", "88", "--limit", "25", "--page-key", "AAAAAAAAAKA=",
				"--market", "444", "--reverse", "--bids",
			},
			expReq: &exchange.QueryGetMarketOrdersRequest{
				MarketId:     444,
				OrderType:    "bid",
				AfterOrderId: 88,
				Pagination: &query.PageRequest{
					Key:     []byte{0, 0, 0, 0, 0, 0, 0, 160},
					Limit:   25,
					Reverse: true,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runQueryMakerTest(t, td, tc)
		})
	}
}

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

func TestMakeQueryGetOwnerOrders(t *testing.T) {
	td := queryMakerTestDef[exchange.QueryGetOwnerOrdersRequest]{
		makerName: "MakeQueryGetOwnerOrders",
		maker:     cli.MakeQueryGetOwnerOrders,
		setup:     cli.SetupCmdQueryGetOwnerOrders,
	}

	defaultPageReq := &query.PageRequest{
		Key:   []byte{},
		Limit: 100,
	}
	tests := []queryMakerTestCase[exchange.QueryGetOwnerOrdersRequest]{
		{
			name: "no owner",
			expReq: &exchange.QueryGetOwnerOrdersRequest{
				Pagination: defaultPageReq,
			},
			expErr: "no <owner> provided",
		},
		{
			name:  "just owner flag",
			flags: []string{"--owner", "someaddr"},
			expReq: &exchange.QueryGetOwnerOrdersRequest{
				Owner:      "someaddr",
				Pagination: defaultPageReq,
			},
		},
		{
			name: "just owner arg",
			args: []string{"otheraddr"},
			expReq: &exchange.QueryGetOwnerOrdersRequest{
				Owner:      "otheraddr",
				Pagination: defaultPageReq,
			},
		},
		{
			name:  "both owner flag and arg",
			flags: []string{"--owner", "someaddr"},
			args:  []string{"otheraddr"},
			expReq: &exchange.QueryGetOwnerOrdersRequest{
				Pagination: defaultPageReq,
			},
			expErr: "cannot provide <owner> as both an arg (\"otheraddr\") and flag (--owner \"someaddr\")",
		},
		{
			name: "all opts asks",
			flags: []string{
				"--asks", "--order", "12", "--limit", "63",
				"--offset", "42", "--count-total",
			},
			args: []string{"otheraddr"},
			expReq: &exchange.QueryGetOwnerOrdersRequest{
				Owner:        "otheraddr",
				OrderType:    "ask",
				AfterOrderId: 12,
				Pagination: &query.PageRequest{
					Key:        []byte{},
					Offset:     42,
					Limit:      63,
					CountTotal: true,
				},
			},
		},
		{
			name: "all opts bids",
			flags: []string{
				"--order", "88", "--limit", "25", "--page-key", "AAAAAAAAAKA=",
				"--owner", "myself", "--reverse", "--bids",
			},
			expReq: &exchange.QueryGetOwnerOrdersRequest{
				Owner:        "myself",
				OrderType:    "bid",
				AfterOrderId: 88,
				Pagination: &query.PageRequest{
					Key:     []byte{0, 0, 0, 0, 0, 0, 0, 160},
					Limit:   25,
					Reverse: true,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runQueryMakerTest(t, td, tc)
		})
	}
}

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

func TestMakeQueryGetAssetOrders(t *testing.T) {
	td := queryMakerTestDef[exchange.QueryGetAssetOrdersRequest]{
		makerName: "MakeQueryGetAssetOrders",
		maker:     cli.MakeQueryGetAssetOrders,
		setup:     cli.SetupCmdQueryGetAssetOrders,
	}

	defaultPageReq := &query.PageRequest{
		Key:   []byte{},
		Limit: 100,
	}
	tests := []queryMakerTestCase[exchange.QueryGetAssetOrdersRequest]{
		{
			name: "no denom",
			expReq: &exchange.QueryGetAssetOrdersRequest{
				Pagination: defaultPageReq,
			},
			expErr: "no <asset> provided",
		},
		{
			name:  "just denom flag",
			flags: []string{"--denom", "mycoin"},
			expReq: &exchange.QueryGetAssetOrdersRequest{
				Asset:      "mycoin",
				Pagination: defaultPageReq,
			},
		},
		{
			name: "just denom arg",
			args: []string{"yourcoin"},
			expReq: &exchange.QueryGetAssetOrdersRequest{
				Asset:      "yourcoin",
				Pagination: defaultPageReq,
			},
		},
		{
			name:  "both denom flag and arg",
			flags: []string{"--denom", "mycoin"},
			args:  []string{"yourcoin"},
			expReq: &exchange.QueryGetAssetOrdersRequest{
				Pagination: defaultPageReq,
			},
			expErr: "cannot provide <asset> as both an arg (\"yourcoin\") and flag (--denom \"mycoin\")",
		},
		{
			name: "all opts asks",
			flags: []string{
				"--asks", "--order", "12", "--limit", "63",
				"--offset", "42", "--count-total",
			},
			args: []string{"yourcoin"},
			expReq: &exchange.QueryGetAssetOrdersRequest{
				Asset:        "yourcoin",
				OrderType:    "ask",
				AfterOrderId: 12,
				Pagination: &query.PageRequest{
					Key:        []byte{},
					Offset:     42,
					Limit:      63,
					CountTotal: true,
				},
			},
		},
		{
			name: "all opts bids",
			flags: []string{
				"--order", "88", "--limit", "25", "--page-key", "AAAAAAAAAKA=",
				"--denom", "mycoin", "--reverse", "--bids",
			},
			expReq: &exchange.QueryGetAssetOrdersRequest{
				Asset:        "mycoin",
				OrderType:    "bid",
				AfterOrderId: 88,
				Pagination: &query.PageRequest{
					Key:     []byte{0, 0, 0, 0, 0, 0, 0, 160},
					Limit:   25,
					Reverse: true,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runQueryMakerTest(t, td, tc)
		})
	}
}

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

func TestMakeQueryGetAllOrders(t *testing.T) {
	td := queryMakerTestDef[exchange.QueryGetAllOrdersRequest]{
		makerName: "MakeQueryGetAllOrders",
		maker:     cli.MakeQueryGetAllOrders,
		setup:     cli.SetupCmdQueryGetAllOrders,
	}

	tests := []queryMakerTestCase[exchange.QueryGetAllOrdersRequest]{
		{
			name: "no flags",
			expReq: &exchange.QueryGetAllOrdersRequest{
				Pagination: &query.PageRequest{
					Key:   []byte{},
					Limit: 100,
				},
			},
		},
		{
			name:  "some pagination flags",
			flags: []string{"--limit", "5", "--reverse", "--page-key", "AAAAAAAAAKA="},
			expReq: &exchange.QueryGetAllOrdersRequest{
				Pagination: &query.PageRequest{
					Key:     []byte{0, 0, 0, 0, 0, 0, 0, 160},
					Limit:   5,
					Reverse: true,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runQueryMakerTest(t, td, tc)
		})
	}
}

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

func TestMakeQueryGetMarket(t *testing.T) {
	td := queryMakerTestDef[exchange.QueryGetMarketRequest]{
		makerName: "MakeQueryGetMarket",
		maker:     cli.MakeQueryGetMarket,
		setup:     cli.SetupCmdQueryGetMarket,
	}

	tests := []queryMakerTestCase[exchange.QueryGetMarketRequest]{
		{
			name:   "no market",
			expReq: &exchange.QueryGetMarketRequest{},
			expErr: "no <market id> provided",
		},
		{
			name:   "just flag",
			flags:  []string{"--market", "2"},
			expReq: &exchange.QueryGetMarketRequest{MarketId: 2},
		},
		{
			name:   "just arg",
			args:   []string{"1000"},
			expReq: &exchange.QueryGetMarketRequest{MarketId: 1000},
		},
		{
			name:   "both arg and flag",
			flags:  []string{"--market", "2"},
			args:   []string{"1000"},
			expReq: &exchange.QueryGetMarketRequest{},
			expErr: "cannot provide <market id> as both an arg (\"1000\") and flag (--market 2)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runQueryMakerTest(t, td, tc)
		})
	}
}

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

func TestMakeQueryGetAllMarkets(t *testing.T) {
	td := queryMakerTestDef[exchange.QueryGetAllMarketsRequest]{
		makerName: "MakeQueryGetAllMarkets",
		maker:     cli.MakeQueryGetAllMarkets,
		setup:     cli.SetupCmdQueryGetAllMarkets,
	}

	tests := []queryMakerTestCase[exchange.QueryGetAllMarketsRequest]{
		{
			name: "no flags",
			expReq: &exchange.QueryGetAllMarketsRequest{
				Pagination: &query.PageRequest{
					Key:   []byte{},
					Limit: 100,
				},
			},
		},
		{
			name:  "some pagination flags",
			flags: []string{"--limit", "5", "--reverse", "--page-key", "AAAAAAAAAKA="},
			expReq: &exchange.QueryGetAllMarketsRequest{
				Pagination: &query.PageRequest{
					Key:     []byte{0, 0, 0, 0, 0, 0, 0, 160},
					Limit:   5,
					Reverse: true,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runQueryMakerTest(t, td, tc)
		})
	}
}

func TestSetupCmdQueryParams(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name:        "SetupCmdQueryParams",
		setup:       cli.SetupCmdQueryParams,
		expExamples: []string{exampleStart},
	})
}

func TestMakeQueryParams(t *testing.T) {
	td := queryMakerTestDef[exchange.QueryParamsRequest]{
		makerName: "MakeQueryParams",
		maker:     cli.MakeQueryParams,
		setup:     cli.SetupCmdQueryParams,
	}

	tests := []queryMakerTestCase[exchange.QueryParamsRequest]{
		{
			name:   "normal",
			expReq: &exchange.QueryParamsRequest{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runQueryMakerTest(t, td, tc)
		})
	}
}

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

func TestMakeQueryValidateCreateMarket(t *testing.T) {
	td := queryMakerTestDef[exchange.QueryValidateCreateMarketRequest]{
		makerName: "MakeQueryValidateCreateMarket",
		maker:     cli.MakeQueryValidateCreateMarket,
		setup:     cli.SetupCmdQueryValidateCreateMarket,
	}

	tests := []queryMakerTestCase[exchange.QueryValidateCreateMarketRequest]{
		{
			name: "several errors",
			flags: []string{
				"--create-ask", "nope", "--seller-ratios", "8apple",
				"--access-grants", "addr8:set", "--accepting-orders",
			},
			expReq: &exchange.QueryValidateCreateMarketRequest{
				CreateMarketRequest: &exchange.MsgGovCreateMarketRequest{
					Authority: cli.AuthorityAddr.String(),
					Market: exchange.Market{
						FeeCreateAskFlat:          []sdk.Coin{},
						FeeSellerSettlementRatios: []exchange.FeeRatio{},
						AcceptingOrders:           true,
						AccessGrants:              []exchange.AccessGrant{},
						ReqAttrCreateAsk:          []string{},
						ReqAttrCreateBid:          []string{},
					},
				},
			},
			expErr: joinErrs(
				"invalid coin expression: \"nope\"",
				"cannot create FeeRatio from \"8apple\": expected exactly one colon",
				"could not parse permissions for \"addr8\" from \"set\": invalid permission: \"set\"",
			),
		},
		{
			name: "all fields",
			flags: []string{
				"--authority", "otherauth", "--market", "18",
				"--create-ask", "10fig", "--create-bid", "5grape",
				"--seller-flat", "12fig", "--seller-ratios", "100prune:1prune",
				"--buyer-flat", "17fig", "--buyer-ratios", "88plum:3plum",
				"--accepting-orders", "--allow-user-settle",
				"--access-grants", "addr1:settle+cancel", "--access-grants", "addr2:update+permissions",
				"--req-attr-ask", "seller.kyc", "--req-attr-bid", "buyer.kyc",
				"--name", "Special market", "--description", "This market is special.",
				"--url", "https://example.com", "--icon", "https://example.com/icon",
				"--access-grants", "addr3:all",
			},
			expReq: &exchange.QueryValidateCreateMarketRequest{
				CreateMarketRequest: &exchange.MsgGovCreateMarketRequest{
					Authority: "otherauth",
					Market: exchange.Market{
						MarketId: 18,
						MarketDetails: exchange.MarketDetails{
							Name:        "Special market",
							Description: "This market is special.",
							WebsiteUrl:  "https://example.com",
							IconUri:     "https://example.com/icon",
						},
						FeeCreateAskFlat:        []sdk.Coin{sdk.NewInt64Coin("fig", 10)},
						FeeCreateBidFlat:        []sdk.Coin{sdk.NewInt64Coin("grape", 5)},
						FeeSellerSettlementFlat: []sdk.Coin{sdk.NewInt64Coin("fig", 12)},
						FeeSellerSettlementRatios: []exchange.FeeRatio{
							{Price: sdk.NewInt64Coin("prune", 100), Fee: sdk.NewInt64Coin("prune", 1)},
						},
						FeeBuyerSettlementFlat: []sdk.Coin{sdk.NewInt64Coin("fig", 17)},
						FeeBuyerSettlementRatios: []exchange.FeeRatio{
							{Price: sdk.NewInt64Coin("plum", 88), Fee: sdk.NewInt64Coin("plum", 3)},
						},
						AcceptingOrders:     true,
						AllowUserSettlement: true,
						AccessGrants: []exchange.AccessGrant{
							{
								Address:     "addr1",
								Permissions: []exchange.Permission{exchange.Permission_settle, exchange.Permission_cancel},
							},
							{
								Address:     "addr2",
								Permissions: []exchange.Permission{exchange.Permission_update, exchange.Permission_permissions},
							},
							{
								Address:     "addr3",
								Permissions: exchange.AllPermissions(),
							},
						},
						ReqAttrCreateAsk: []string{"seller.kyc"},
						ReqAttrCreateBid: []string{"buyer.kyc"},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runQueryMakerTest(t, td, tc)
		})
	}
}

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

func TestMakeQueryValidateMarket(t *testing.T) {
	td := queryMakerTestDef[exchange.QueryValidateMarketRequest]{
		makerName: "MakeQueryValidateMarket",
		maker:     cli.MakeQueryValidateMarket,
		setup:     cli.SetupCmdQueryValidateMarket,
	}

	tests := []queryMakerTestCase[exchange.QueryValidateMarketRequest]{
		{
			name:   "no market",
			expReq: &exchange.QueryValidateMarketRequest{},
			expErr: "no <market id> provided",
		},
		{
			name:   "just flag",
			flags:  []string{"--market", "2"},
			expReq: &exchange.QueryValidateMarketRequest{MarketId: 2},
		},
		{
			name:   "just arg",
			args:   []string{"1000"},
			expReq: &exchange.QueryValidateMarketRequest{MarketId: 1000},
		},
		{
			name:   "both arg and flag",
			flags:  []string{"--market", "2"},
			args:   []string{"1000"},
			expReq: &exchange.QueryValidateMarketRequest{},
			expErr: "cannot provide <market id> as both an arg (\"1000\") and flag (--market 2)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runQueryMakerTest(t, td, tc)
		})
	}
}

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
			cli.FlagMarket: {required: {"true"}},
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

func TestMakeQueryValidateManageFees(t *testing.T) {
	td := queryMakerTestDef[exchange.QueryValidateManageFeesRequest]{
		makerName: "MakeQueryValidateManageFees",
		maker:     cli.MakeQueryValidateManageFees,
		setup:     cli.SetupCmdQueryValidateManageFees,
	}

	tests := []queryMakerTestCase[exchange.QueryValidateManageFeesRequest]{
		{
			name: "multiple errors",
			flags: []string{
				"--ask-add", "15", "--buyer-flat-remove", "noamt",
			},
			expReq: &exchange.QueryValidateManageFeesRequest{
				ManageFeesRequest: &exchange.MsgGovManageFeesRequest{
					Authority:                    cli.AuthorityAddr.String(),
					AddFeeCreateAskFlat:          []sdk.Coin{},
					RemoveFeeBuyerSettlementFlat: []sdk.Coin{},
				},
			},
			expErr: joinErrs(
				"invalid coin expression: \"15\"",
				"invalid coin expression: \"noamt\"",
			),
		},
		{
			name: "all fields",
			flags: []string{
				"--authority", "respect", "--market", "55",
				"--ask-add", "18fig", "--ask-remove", "15fig", "--ask-add", "5grape",
				"--bid-add", "17fig", "--bid-remove", "14fig",
				"--seller-flat-add", "55prune", "--seller-flat-remove", "54prune",
				"--seller-ratios-add", "101prune:7prune", "--seller-ratios-remove", "101prune:3prune",
				"--buyer-flat-add", "59prune", "--buyer-flat-remove", "57prune",
				"--buyer-ratios-add", "107prune:1prune", "--buyer-ratios-remove", "43prune:2prune",
			},
			expReq: &exchange.QueryValidateManageFeesRequest{
				ManageFeesRequest: &exchange.MsgGovManageFeesRequest{
					Authority:                     "respect",
					MarketId:                      55,
					AddFeeCreateAskFlat:           []sdk.Coin{sdk.NewInt64Coin("fig", 18), sdk.NewInt64Coin("grape", 5)},
					RemoveFeeCreateAskFlat:        []sdk.Coin{sdk.NewInt64Coin("fig", 15)},
					AddFeeCreateBidFlat:           []sdk.Coin{sdk.NewInt64Coin("fig", 17)},
					RemoveFeeCreateBidFlat:        []sdk.Coin{sdk.NewInt64Coin("fig", 14)},
					AddFeeSellerSettlementFlat:    []sdk.Coin{sdk.NewInt64Coin("prune", 55)},
					RemoveFeeSellerSettlementFlat: []sdk.Coin{sdk.NewInt64Coin("prune", 54)},
					AddFeeSellerSettlementRatios: []exchange.FeeRatio{
						{Price: sdk.NewInt64Coin("prune", 101), Fee: sdk.NewInt64Coin("prune", 7)},
					},
					RemoveFeeSellerSettlementRatios: []exchange.FeeRatio{
						{Price: sdk.NewInt64Coin("prune", 101), Fee: sdk.NewInt64Coin("prune", 3)},
					},
					AddFeeBuyerSettlementFlat:    []sdk.Coin{sdk.NewInt64Coin("prune", 59)},
					RemoveFeeBuyerSettlementFlat: []sdk.Coin{sdk.NewInt64Coin("prune", 57)},
					AddFeeBuyerSettlementRatios: []exchange.FeeRatio{
						{Price: sdk.NewInt64Coin("prune", 107), Fee: sdk.NewInt64Coin("prune", 1)},
					},
					RemoveFeeBuyerSettlementRatios: []exchange.FeeRatio{
						{Price: sdk.NewInt64Coin("prune", 43), Fee: sdk.NewInt64Coin("prune", 2)},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runQueryMakerTest(t, td, tc)
		})
	}
}
