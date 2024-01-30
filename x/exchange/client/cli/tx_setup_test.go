package cli_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/exchange"
	"github.com/provenance-io/provenance/x/exchange/client/cli"
)

// "cosmos1geex7m2pv3j8yetnwd047h6lta047h6ls98cgw" = sdk.AccAddress("FromAddress_________").String()
// "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn" = cli.AuthorityAddr.String()

// txMakerTestDef is the definition of a tx maker func to be tested.
//
// R is the type of the sdk.Msg returned by the maker.
type txMakerTestDef[R sdk.Msg] struct {
	// makerName is the name of the maker func being tested.
	makerName string
	// maker is the tx request maker func being tested.
	maker func(clientCtx client.Context, flagSet *pflag.FlagSet, args []string) (R, error)
	// setup is the command setup func that sets up a command so it has what's needed by the maker.
	setup func(cmd *cobra.Command)
}

// txMakerTestCase is a test case for a tx maker func.
//
// R is the type of the sdk.Msg returned by the maker.
type txMakerTestCase[R sdk.Msg] struct {
	// name is a name for this test case.
	name string
	// clientCtx is the client context to provided to the maker.
	clientCtx client.Context
	// flags are the strings the give to FlagSet before it's provided to the maker.
	flags []string
	// args are the strings to supply as args to the maker.
	args []string
	// expMsg is the expected Msg result.
	expMsg R
	// expErr is the expected error string. An empty string indicates the error should be nil.
	expErr string
}

// runTxMakerTestCase runs a test case for a tx maker func.
//
// R is the type of the sdk.Msg returned by the maker.
func runTxMakerTestCase[R sdk.Msg](t *testing.T, td txMakerTestDef[R], tc txMakerTestCase[R]) {
	cmd := &cobra.Command{
		Use: "dummy",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("this dummy command should not have been executed")
		},
	}
	cmd.Flags().String(flags.FlagFrom, "", "The from address")
	td.setup(cmd)

	err := cmd.Flags().Parse(tc.flags)
	require.NoError(t, err, "cmd.Flags().Parse(%q)", tc.flags)

	var msg R
	testFunc := func() {
		msg, err = td.maker(tc.clientCtx, cmd.Flags(), tc.args)
	}
	require.NotPanics(t, testFunc, td.makerName)
	assertions.AssertErrorValue(t, err, tc.expErr, "%s error", td.makerName)
	assert.Equal(t, tc.expMsg, msg, "%s msg", td.makerName)
}

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
			cli.FlagMarket: {required: {"true"}},
			cli.FlagAssets: {required: {"true"}},
			cli.FlagPrice:  {required: {"true"}},
		},
		expInUse: []string{
			"--seller", "--market <market id>", "--assets <assets>", "--price <price>",
			"[--settlement-fee <seller settlement flat fee>]", "[--partial]",
			"[--external-id <external id>]", "[--creation-fee <creation fee>]",
			cli.ReqSignerDesc(cli.FlagSeller),
		},
	})
}

func TestMakeMsgCreateAsk(t *testing.T) {
	td := txMakerTestDef[*exchange.MsgCreateAskRequest]{
		makerName: "MakeMsgCreateAsk",
		maker:     cli.MakeMsgCreateAsk,
		setup:     cli.SetupCmdTxCreateAsk,
	}

	tests := []txMakerTestCase[*exchange.MsgCreateAskRequest]{
		{
			name:      "a couple errors",
			clientCtx: client.Context{FromAddress: sdk.AccAddress("FromAddress_________")},
			flags:     []string{"--assets", "nope", "--creation-fee", "123"},
			expMsg: &exchange.MsgCreateAskRequest{
				AskOrder: exchange.AskOrder{Seller: sdk.AccAddress("FromAddress_________").String()},
			},
			expErr: joinErrs(
				"error parsing --assets as a coin: invalid coin expression: \"nope\"",
				"missing required --price flag",
				"error parsing --creation-fee as a coin: invalid coin expression: \"123\"",
			),
		},
		{
			name: "all fields",
			flags: []string{
				"--seller", "someaddr", "--market", "4",
				"--assets", "10apple", "--price", "55plum",
				"--settlement-fee", "5fig", "--partial",
				"--external-id", "uuid", "--creation-fee", "6grape",
			},
			expMsg: &exchange.MsgCreateAskRequest{
				AskOrder: exchange.AskOrder{
					MarketId:                4,
					Seller:                  "someaddr",
					Assets:                  sdk.NewInt64Coin("apple", 10),
					Price:                   sdk.NewInt64Coin("plum", 55),
					SellerSettlementFlatFee: &sdk.Coin{Denom: "fig", Amount: sdkmath.NewInt(5)},
					AllowPartial:            true,
					ExternalId:              "uuid",
				},
				OrderCreationFee: &sdk.Coin{Denom: "grape", Amount: sdkmath.NewInt(6)},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runTxMakerTestCase(t, td, tc)
		})
	}
}

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
			cli.FlagMarket: {required: {"true"}},
			cli.FlagAssets: {required: {"true"}},
			cli.FlagPrice:  {required: {"true"}},
		},
		expInUse: []string{
			"--buyer", "--market <market id>", "--assets <assets>", "--price <price>",
			"[--settlement-fee <seller settlement flat fee>]", "[--partial]",
			"[--external-id <external id>]", "[--creation-fee <creation fee>]",
			cli.ReqSignerDesc(cli.FlagBuyer),
		},
	})
}

func TestMakeMsgCreateBid(t *testing.T) {
	td := txMakerTestDef[*exchange.MsgCreateBidRequest]{
		makerName: "MakeMsgCreateBid",
		maker:     cli.MakeMsgCreateBid,
		setup:     cli.SetupCmdTxCreateBid,
	}

	tests := []txMakerTestCase[*exchange.MsgCreateBidRequest]{
		{
			name:      "a couple errors",
			clientCtx: client.Context{FromAddress: sdk.AccAddress("FromAddress_________")},
			flags:     []string{"--assets", "nope", "--creation-fee", "123"},
			expMsg: &exchange.MsgCreateBidRequest{
				BidOrder: exchange.BidOrder{Buyer: sdk.AccAddress("FromAddress_________").String()},
			},
			expErr: joinErrs(
				"error parsing --assets as a coin: invalid coin expression: \"nope\"",
				"missing required --price flag",
				"error parsing --creation-fee as a coin: invalid coin expression: \"123\"",
			),
		},
		{
			name: "all fields",
			flags: []string{
				"--buyer", "someaddr", "--market", "4",
				"--assets", "10apple", "--price", "55plum",
				"--settlement-fee", "5fig", "--partial",
				"--external-id", "uuid", "--creation-fee", "6grape",
			},
			expMsg: &exchange.MsgCreateBidRequest{
				BidOrder: exchange.BidOrder{
					MarketId:            4,
					Buyer:               "someaddr",
					Assets:              sdk.NewInt64Coin("apple", 10),
					Price:               sdk.NewInt64Coin("plum", 55),
					BuyerSettlementFees: sdk.Coins{sdk.Coin{Denom: "fig", Amount: sdkmath.NewInt(5)}},
					AllowPartial:        true,
					ExternalId:          "uuid",
				},
				OrderCreationFee: &sdk.Coin{Denom: "grape", Amount: sdkmath.NewInt(6)},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runTxMakerTestCase(t, td, tc)
		})
	}
}

// TODO[1789]: func TestSetupCmdTxCommitFunds(t *testing.T)

// TODO[1789]: func TestMakeMsgCommitFunds(t *testing.T)

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

func TestMakeMsgCancelOrder(t *testing.T) {
	td := txMakerTestDef[*exchange.MsgCancelOrderRequest]{
		makerName: "MakeMsgCancelOrder",
		maker:     cli.MakeMsgCancelOrder,
		setup:     cli.SetupCmdTxCancelOrder,
	}

	tests := []txMakerTestCase[*exchange.MsgCancelOrderRequest]{
		{
			name:   "nothing",
			expMsg: &exchange.MsgCancelOrderRequest{},
			expErr: joinErrs(
				"no <signer> provided",
				"no <order id> provided",
			),
		},
		{
			name:      "from and arg",
			clientCtx: client.Context{FromAddress: sdk.AccAddress("FromAddress_________")},
			args:      []string{"87"},
			expMsg: &exchange.MsgCancelOrderRequest{
				Signer:  sdk.AccAddress("FromAddress_________").String(),
				OrderId: 87,
			},
		},
		{
			name:  "signer and flag",
			flags: []string{"--order", "52", "--signer", "someone"},
			expMsg: &exchange.MsgCancelOrderRequest{
				Signer:  "someone",
				OrderId: 52,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runTxMakerTestCase(t, td, tc)
		})
	}
}

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
			cli.FlagMarket: {required: {"true"}},
			cli.FlagAssets: {required: {"true"}},
			cli.FlagBids:   {required: {"true"}},
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

func TestMakeMsgFillBids(t *testing.T) {
	td := txMakerTestDef[*exchange.MsgFillBidsRequest]{
		makerName: "MakeMsgFillBids",
		maker:     cli.MakeMsgFillBids,
		setup:     cli.SetupCmdTxFillBids,
	}

	tests := []txMakerTestCase[*exchange.MsgFillBidsRequest]{
		{
			name:      "some errors",
			clientCtx: client.Context{FromAddress: sdk.AccAddress("FromAddress_________")},
			flags:     []string{"--assets", "18", "--creation-fee", "apple"},
			expMsg: &exchange.MsgFillBidsRequest{
				Seller: sdk.AccAddress("FromAddress_________").String(),
			},
			expErr: joinErrs(
				"error parsing --assets as coins: invalid coin expression: \"18\"",
				"error parsing --creation-fee as a coin: invalid coin expression: \"apple\"",
			),
		},
		{
			name: "all the flags",
			flags: []string{
				"--market", "10", "--seller", "mike",
				"--assets", "18acorn,5apple", "--bids", "83,52,99",
				"--settlement-fee", "15fig", "--creation-fee", "9grape",
				"--bids", "5",
			},
			expMsg: &exchange.MsgFillBidsRequest{
				Seller:                  "mike",
				MarketId:                10,
				TotalAssets:             sdk.NewCoins(sdk.NewInt64Coin("acorn", 18), sdk.NewInt64Coin("apple", 5)),
				BidOrderIds:             []uint64{83, 52, 99, 5},
				SellerSettlementFlatFee: &sdk.Coin{Denom: "fig", Amount: sdkmath.NewInt(15)},
				AskOrderCreationFee:     &sdk.Coin{Denom: "grape", Amount: sdkmath.NewInt(9)},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runTxMakerTestCase(t, td, tc)
		})
	}
}

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
			cli.FlagMarket: {required: {"true"}},
			cli.FlagPrice:  {required: {"true"}},
			cli.FlagAsks:   {required: {"true"}},
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

func TestMakeMsgFillAsks(t *testing.T) {
	td := txMakerTestDef[*exchange.MsgFillAsksRequest]{
		makerName: "MakeMsgFillAsks",
		maker:     cli.MakeMsgFillAsks,
		setup:     cli.SetupCmdTxFillAsks,
	}

	tests := []txMakerTestCase[*exchange.MsgFillAsksRequest]{
		{
			name:      "some errors",
			clientCtx: client.Context{FromAddress: sdk.AccAddress("FromAddress_________")},
			flags:     []string{"--price", "18", "--creation-fee", "apple"},
			expMsg: &exchange.MsgFillAsksRequest{
				Buyer: sdk.AccAddress("FromAddress_________").String(),
			},
			expErr: joinErrs(
				"error parsing --price as a coin: invalid coin expression: \"18\"",
				"error parsing --creation-fee as a coin: invalid coin expression: \"apple\"",
			),
		},
		{
			name: "all the flags",
			flags: []string{
				"--market", "10", "--buyer", "george",
				"--price", "23apple", "--asks", "41,12,77",
				"--settlement-fee", "15fig", "--creation-fee", "9grape",
				"--asks", "20", "--asks", "987,444,6",
			},
			expMsg: &exchange.MsgFillAsksRequest{
				Buyer:               "george",
				MarketId:            10,
				TotalPrice:          sdk.NewInt64Coin("apple", 23),
				AskOrderIds:         []uint64{41, 12, 77, 20, 987, 444, 6},
				BuyerSettlementFees: sdk.NewCoins(sdk.NewInt64Coin("fig", 15)),
				BidOrderCreationFee: &sdk.Coin{Denom: "grape", Amount: sdkmath.NewInt(9)},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runTxMakerTestCase(t, td, tc)
		})
	}
}

func TestSetupCmdTxMarketSettle(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name:  "SetupCmdTxMarketSettle",
		setup: cli.SetupCmdTxMarketSettle,
		expFlags: []string{
			cli.FlagMarket, cli.FlagAsks, cli.FlagBids, cli.FlagPartial,
		},
		expAnnotations: map[string]map[string][]string{
			cli.FlagMarket: {required: {"true"}},
			cli.FlagAsks:   {required: {"true"}},
			cli.FlagBids:   {required: {"true"}},
		},
		expInUse: []string{
			cli.ReqAdminUse, "--market <market id>",
			"--asks <ask order ids>", "--bids <bid order ids>",
			"[--partial]",
			cli.ReqAdminDesc, cli.RepeatableDesc,
		},
	})
}

func TestMakeMsgMarketSettle(t *testing.T) {
	td := txMakerTestDef[*exchange.MsgMarketSettleRequest]{
		makerName: "MakeMsgMarketSettle",
		maker:     cli.MakeMsgMarketSettle,
		setup:     cli.SetupCmdTxMarketSettle,
	}

	tests := []txMakerTestCase[*exchange.MsgMarketSettleRequest]{
		{
			name:  "no admin",
			flags: []string{"--asks", "7", "--bids", "8", "--partial"},
			expMsg: &exchange.MsgMarketSettleRequest{
				AskOrderIds:   []uint64{7},
				BidOrderIds:   []uint64{8},
				ExpectPartial: true,
			},
			expErr: "no <admin> provided",
		},
		{
			name:      "from",
			clientCtx: client.Context{FromAddress: sdk.AccAddress("FromAddress_________")},
			flags: []string{
				"--asks", "15,16,17", "--market", "52", "--bids", "51,52,53",
				"--asks", "8", "--bids", "9"},
			expMsg: &exchange.MsgMarketSettleRequest{
				Admin:       sdk.AccAddress("FromAddress_________").String(),
				MarketId:    52,
				AskOrderIds: []uint64{15, 16, 17, 8},
				BidOrderIds: []uint64{51, 52, 53, 9},
			},
		},
		{
			name:  "authority",
			flags: []string{"--market", "52", "--asks", "91", "--bids", "12,13", "--authority", "--partial"},
			expMsg: &exchange.MsgMarketSettleRequest{
				Admin:         cli.AuthorityAddr.String(),
				MarketId:      52,
				AskOrderIds:   []uint64{91},
				BidOrderIds:   []uint64{12, 13},
				ExpectPartial: true,
			},
		},
		{
			name:  "admin",
			flags: []string{"--market", "14", "--admin", "bob", "--asks", "1,2,3", "--bids", "5"},
			expMsg: &exchange.MsgMarketSettleRequest{
				Admin:       "bob",
				MarketId:    14,
				AskOrderIds: []uint64{1, 2, 3},
				BidOrderIds: []uint64{5},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runTxMakerTestCase(t, td, tc)
		})
	}
}

// TODO[1789]: func TestSetupCmdTxMarketCommitmentSettle(t *testing.T)

// TODO[1789]: func TestMakeMsgMarketCommitmentSettle(t *testing.T)

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
			cli.FlagMarket: {required: {"true"}},
			cli.FlagOrder:  {required: {"true"}},
		},
		expInUse: []string{
			cli.ReqAdminUse, "--market <market id>", "--order <order id>",
			"[--external-id <external id>]",
			cli.ReqAdminDesc,
		},
	})
}

func TestMakeMsgMarketSetOrderExternalID(t *testing.T) {
	td := txMakerTestDef[*exchange.MsgMarketSetOrderExternalIDRequest]{
		makerName: "MakeMsgMarketSetOrderExternalID",
		maker:     cli.MakeMsgMarketSetOrderExternalID,
		setup:     cli.SetupCmdTxMarketSetOrderExternalID,
	}

	tests := []txMakerTestCase[*exchange.MsgMarketSetOrderExternalIDRequest]{
		{
			name:  "no admin",
			flags: []string{"--market", "8", "--order", "7", "--external-id", "markus"},
			expMsg: &exchange.MsgMarketSetOrderExternalIDRequest{
				MarketId: 8, OrderId: 7, ExternalId: "markus",
			},
			expErr: "no <admin> provided",
		},
		{
			name:      "no external id",
			clientCtx: client.Context{FromAddress: sdk.AccAddress("FromAddress_________")},
			flags:     []string{"--market", "4000", "--order", "9001"},
			expMsg: &exchange.MsgMarketSetOrderExternalIDRequest{
				Admin:    sdk.AccAddress("FromAddress_________").String(),
				MarketId: 4000, OrderId: 9001, ExternalId: "",
			},
		},
		{
			name: "all the flags",
			flags: []string{
				"--market", "5", "--order", "100000000000001",
				"--external-id", "one", "--admin", "michelle",
			},
			expMsg: &exchange.MsgMarketSetOrderExternalIDRequest{
				Admin: "michelle", MarketId: 5, OrderId: 100000000000001, ExternalId: "one",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runTxMakerTestCase(t, td, tc)
		})
	}
}

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
			cli.FlagMarket: {required: {"true"}},
			cli.FlagTo:     {required: {"true"}},
			cli.FlagAmount: {required: {"true"}},
		},
		expInUse: []string{
			cli.ReqAdminUse, "--market <market id>", "--to <to address>", "--amount <amount>",
			cli.ReqAdminDesc,
		},
	})
}

func TestMakeMsgMarketWithdraw(t *testing.T) {
	td := txMakerTestDef[*exchange.MsgMarketWithdrawRequest]{
		makerName: "MakeMsgMarketWithdraw",
		maker:     cli.MakeMsgMarketWithdraw,
		setup:     cli.SetupCmdTxMarketWithdraw,
	}

	tests := []txMakerTestCase[*exchange.MsgMarketWithdrawRequest]{
		{
			name:  "some errors",
			flags: []string{"--market", "5", "--to", "annie", "--amount", "bill"},
			expMsg: &exchange.MsgMarketWithdrawRequest{
				Admin: "", MarketId: 5, ToAddress: "annie", Amount: nil,
			},
			expErr: joinErrs(
				"no <admin> provided",
				"error parsing --amount as coins: invalid coin expression: \"bill\"",
			),
		},
		{
			name:      "all fields",
			clientCtx: client.Context{FromAddress: sdk.AccAddress("FromAddress_________")},
			flags:     []string{"--market", "2", "--to", "samantha", "--amount", "52plum,18pear"},
			expMsg: &exchange.MsgMarketWithdrawRequest{
				Admin:    sdk.AccAddress("FromAddress_________").String(),
				MarketId: 2, ToAddress: "samantha",
				Amount: sdk.NewCoins(sdk.NewInt64Coin("plum", 52), sdk.NewInt64Coin("pear", 18)),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runTxMakerTestCase(t, td, tc)
		})
	}
}

func TestAddFlagsMarketDetails(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name:          "AddFlagsMarketDetails",
		setup:         cli.AddFlagsMarketDetails,
		expFlags:      []string{cli.FlagName, cli.FlagDescription, cli.FlagURL, cli.FlagIcon},
		skipArgsCheck: true,
	})
}

func TestReadFlagsMarketDetails(t *testing.T) {
	tests := []struct {
		name       string
		skipSetup  bool
		flags      []string
		def        exchange.MarketDetails
		expDetails exchange.MarketDetails
		expErr     string
	}{
		{
			name:      "no setup",
			skipSetup: true,
			expErr: joinErrs(
				"flag accessed but not defined: name",
				"flag accessed but not defined: description",
				"flag accessed but not defined: url",
				"flag accessed but not defined: icon",
			),
		},
		{
			name:       "just name",
			flags:      []string{"--name", "Richard"},
			expDetails: exchange.MarketDetails{Name: "Richard"},
		},
		{
			name:  "name and url, no defaults",
			flags: []string{"--url", "https://example.com", "--name", "Sally"},
			expDetails: exchange.MarketDetails{
				Name:       "Sally",
				WebsiteUrl: "https://example.com",
			},
		},
		{
			name:  "name and url, with defaults",
			flags: []string{"--url", "https://example.com/new", "--name", "Glen"},
			def: exchange.MarketDetails{
				Name:        "Martha",
				Description: "Some existing Description",
				WebsiteUrl:  "https://old.example.com",
				IconUri:     "https://example.com/icon",
			},
			expDetails: exchange.MarketDetails{
				Name:        "Glen",
				Description: "Some existing Description",
				WebsiteUrl:  "https://example.com/new",
				IconUri:     "https://example.com/icon",
			},
		},
		{
			name: "all fields",
			flags: []string{
				"--name", "Market Eight Dude",
				"--description", "The Little Lebowski",
				"--url", "https://bowling.god",
				"--icon", "https://bowling.god/icon",
			},
			expDetails: exchange.MarketDetails{
				Name:        "Market Eight Dude",
				Description: "The Little Lebowski",
				WebsiteUrl:  "https://bowling.god",
				IconUri:     "https://bowling.god/icon",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use: "dummy",
				RunE: func(cmd *cobra.Command, args []string) error {
					return errors.New("this dummy command should not have been executed")
				},
			}
			if !tc.skipSetup {
				cli.AddFlagsMarketDetails(cmd)
			}

			err := cmd.Flags().Parse(tc.flags)
			require.NoError(t, err, "cmd.Flags().Parse(%q)", tc.flags)

			var details exchange.MarketDetails
			testFunc := func() {
				details, err = cli.ReadFlagsMarketDetails(cmd.Flags(), tc.def)
			}
			require.NotPanics(t, testFunc, "ReadFlagsMarketDetails")
			assertions.AssertErrorValue(t, err, tc.expErr, "ReadFlagsMarketDetails")
			assert.Equal(t, tc.expDetails, details, "ReadFlagsMarketDetails")
		})
	}
}

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
			cli.FlagMarket: {required: {"true"}},
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

func TestMakeMsgMarketUpdateDetails(t *testing.T) {
	td := txMakerTestDef[*exchange.MsgMarketUpdateDetailsRequest]{
		makerName: "MakeMsgMarketUpdateDetails",
		maker:     cli.MakeMsgMarketUpdateDetails,
		setup:     cli.SetupCmdTxMarketUpdateDetails,
	}

	tests := []txMakerTestCase[*exchange.MsgMarketUpdateDetailsRequest]{
		{
			name:  "no admin",
			flags: []string{"--market", "8", "--name", "Lynne"},
			expMsg: &exchange.MsgMarketUpdateDetailsRequest{
				MarketId:      8,
				MarketDetails: exchange.MarketDetails{Name: "Lynne"},
			},
			expErr: "no <admin> provided",
		},
		{
			name:      "just name and description",
			clientCtx: client.Context{FromAddress: sdk.AccAddress("FromAddress_________")},
			flags:     []string{"--market", "9002", "--name", "River", "--description", "The person, not the water."},
			expMsg: &exchange.MsgMarketUpdateDetailsRequest{
				Admin:    sdk.AccAddress("FromAddress_________").String(),
				MarketId: 9002,
				MarketDetails: exchange.MarketDetails{
					Name:        "River",
					Description: "The person, not the water.",
				},
			},
		},
		{
			name: "all fields",
			flags: []string{
				"--market", "14", "--authority", "--name", "Ashley",
				"--icon", "https://example.com/ashley/icon",
				"--url", "https://example.com/ashley",
				"--description", "The best market out there.",
			},
			expMsg: &exchange.MsgMarketUpdateDetailsRequest{
				Admin:    cli.AuthorityAddr.String(),
				MarketId: 14,
				MarketDetails: exchange.MarketDetails{
					Name:        "Ashley",
					Description: "The best market out there.",
					WebsiteUrl:  "https://example.com/ashley",
					IconUri:     "https://example.com/ashley/icon",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runTxMakerTestCase(t, td, tc)
		})
	}
}

func TestSetupCmdTxMarketUpdateAcceptingOrders(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name:  "SetupCmdTxMarketUpdateAcceptingOrders",
		setup: cli.SetupCmdTxMarketUpdateAcceptingOrders,
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
			cli.FlagMarket: {required: {"true"}},
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

func TestMakeMsgMarketUpdateAcceptingOrders(t *testing.T) {
	td := txMakerTestDef[*exchange.MsgMarketUpdateAcceptingOrdersRequest]{
		makerName: "MakeMsgMarketUpdateAcceptingOrders",
		maker:     cli.MakeMsgMarketUpdateAcceptingOrders,
		setup:     cli.SetupCmdTxMarketUpdateAcceptingOrders,
	}

	tests := []txMakerTestCase[*exchange.MsgMarketUpdateAcceptingOrdersRequest]{
		{
			name:   "some errors",
			flags:  []string{"--market", "56"},
			expMsg: &exchange.MsgMarketUpdateAcceptingOrdersRequest{MarketId: 56},
			expErr: joinErrs(
				"no <admin> provided",
				"exactly one of --enable or --disable must be provided",
			),
		},
		{
			name:      "enable",
			clientCtx: client.Context{FromAddress: sdk.AccAddress("FromAddress_________")},
			flags:     []string{"--enable", "--market", "4"},
			expMsg: &exchange.MsgMarketUpdateAcceptingOrdersRequest{
				Admin:           sdk.AccAddress("FromAddress_________").String(),
				MarketId:        4,
				AcceptingOrders: true,
			},
		},
		{
			name:      "disable",
			clientCtx: client.Context{FromAddress: sdk.AccAddress("FromAddress_________")},
			flags:     []string{"--admin", "Blake", "--market", "94", "--disable"},
			expMsg: &exchange.MsgMarketUpdateAcceptingOrdersRequest{
				Admin:           "Blake",
				MarketId:        94,
				AcceptingOrders: false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runTxMakerTestCase(t, td, tc)
		})
	}
}

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
			cli.FlagMarket: {required: {"true"}},
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

func TestMakeMsgMarketUpdateUserSettle(t *testing.T) {
	td := txMakerTestDef[*exchange.MsgMarketUpdateUserSettleRequest]{
		makerName: "MakeMsgMarketUpdateUserSettle",
		maker:     cli.MakeMsgMarketUpdateUserSettle,
		setup:     cli.SetupCmdTxMarketUpdateUserSettle,
	}

	tests := []txMakerTestCase[*exchange.MsgMarketUpdateUserSettleRequest]{
		{
			name:   "some errors",
			flags:  []string{"--market", "56"},
			expMsg: &exchange.MsgMarketUpdateUserSettleRequest{MarketId: 56},
			expErr: joinErrs(
				"no <admin> provided",
				"exactly one of --enable or --disable must be provided",
			),
		},
		{
			name:      "enable",
			clientCtx: client.Context{FromAddress: sdk.AccAddress("FromAddress_________")},
			flags:     []string{"--enable", "--market", "4"},
			expMsg: &exchange.MsgMarketUpdateUserSettleRequest{
				Admin:               sdk.AccAddress("FromAddress_________").String(),
				MarketId:            4,
				AllowUserSettlement: true,
			},
		},
		{
			name:      "disable",
			clientCtx: client.Context{FromAddress: sdk.AccAddress("FromAddress_________")},
			flags:     []string{"--admin", "Blake", "--market", "94", "--disable"},
			expMsg: &exchange.MsgMarketUpdateUserSettleRequest{
				Admin:               "Blake",
				MarketId:            94,
				AllowUserSettlement: false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runTxMakerTestCase(t, td, tc)
		})
	}
}

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
			cli.FlagMarket:    {required: {"true"}},
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

func TestMakeMsgMarketManagePermissions(t *testing.T) {
	td := txMakerTestDef[*exchange.MsgMarketManagePermissionsRequest]{
		makerName: "MakeMsgMarketManagePermissions",
		maker:     cli.MakeMsgMarketManagePermissions,
		setup:     cli.SetupCmdTxMarketManagePermissions,
	}

	accessGrant := func(addr string, perms ...exchange.Permission) exchange.AccessGrant {
		return exchange.AccessGrant{Address: addr, Permissions: perms}
	}
	tests := []txMakerTestCase[*exchange.MsgMarketManagePermissionsRequest]{
		{
			name:  "some errors",
			flags: []string{"--revoke", "addr8:oops", "--revoke", "Ryan", "--market", "1", "--grant", ":settle"},
			expMsg: &exchange.MsgMarketManagePermissionsRequest{
				MarketId:  1,
				RevokeAll: []string{},
				ToRevoke:  []exchange.AccessGrant{},
				ToGrant:   []exchange.AccessGrant{},
			},
			expErr: joinErrs(
				"no <admin> provided",
				"could not parse permissions for \"addr8\" from \"oops\": invalid permission: \"oops\"",
				"could not parse \"Ryan\" as an <access grant>: expected format <address>:<permissions>",
				"invalid <access grant> \":settle\": both an <address> and <permissions> are required",
			),
		},
		{
			name:      "just a revoke",
			clientCtx: client.Context{FromAddress: sdk.AccAddress("FromAddress_________")},
			flags: []string{
				"--market", "6", "--revoke", "alan:settle+update"},
			expMsg: &exchange.MsgMarketManagePermissionsRequest{
				Admin:     sdk.AccAddress("FromAddress_________").String(),
				MarketId:  6,
				RevokeAll: []string{},
				ToRevoke: []exchange.AccessGrant{
					accessGrant("alan", exchange.Permission_settle, exchange.Permission_update),
				},
				ToGrant: nil,
			},
		},
		{
			name: "all fields",
			flags: []string{
				"--market", "123", "--admin", "Frankie", "--revoke-all", "Freddie,Fritz,Forrest",
				"--revoke", "Dylan:settle,Devin:update", "--revoke-all", "Finn",
				"--grant", "Sam:permissions+update", "--revoke", "Dave:setids",
				"--grant", "Skylar:all,Fritz:all",
			},
			expMsg: &exchange.MsgMarketManagePermissionsRequest{
				Admin:     "Frankie",
				MarketId:  123,
				RevokeAll: []string{"Freddie", "Fritz", "Forrest", "Finn"},
				ToRevoke: []exchange.AccessGrant{
					accessGrant("Dylan", exchange.Permission_settle),
					accessGrant("Devin", exchange.Permission_update),
					accessGrant("Dave", exchange.Permission_set_ids),
				},
				ToGrant: []exchange.AccessGrant{
					accessGrant("Sam", exchange.Permission_permissions, exchange.Permission_update),
					accessGrant("Skylar", exchange.AllPermissions()...),
					accessGrant("Fritz", exchange.AllPermissions()...),
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runTxMakerTestCase(t, td, tc)
		})
	}
}

func TestSetupCmdTxMarketManageReqAttrs(t *testing.T) {
	tc := setupTestCase{
		name:  "SetupCmdTxMarketManageReqAttrs",
		setup: cli.SetupCmdTxMarketManageReqAttrs,
		expFlags: []string{
			cli.FlagAdmin, cli.FlagAuthority, cli.FlagMarket,
			cli.FlagAskAdd, cli.FlagAskRemove, cli.FlagBidAdd, cli.FlagBidRemove,
			cli.FlagCommitmentAdd, cli.FlagCommitmentRemove,
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
			cli.FlagMarket: {required: {"true"}},
		},
		expInUse: []string{
			cli.ReqAdminUse, "--market <market id>",
			"[--ask-add <attrs>]", "[--ask-remove <attrs>]",
			"[--bid-add <attrs>]", "[--bid-remove <attrs>]",
			"[--commitment-add <attrs>]", "[--commitment-remove <attrs>]",
			cli.ReqAdminDesc, cli.RepeatableDesc,
		},
	}

	oneReqFlags := []string{
		cli.FlagAskAdd, cli.FlagAskRemove, cli.FlagBidAdd, cli.FlagBidRemove,
		cli.FlagCommitmentAdd, cli.FlagCommitmentRemove,
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

func TestMakeMsgMarketManageReqAttrs(t *testing.T) {
	td := txMakerTestDef[*exchange.MsgMarketManageReqAttrsRequest]{
		makerName: "MakeMsgMarketManageReqAttrs",
		maker:     cli.MakeMsgMarketManageReqAttrs,
		setup:     cli.SetupCmdTxMarketManageReqAttrs,
	}

	tests := []txMakerTestCase[*exchange.MsgMarketManageReqAttrsRequest]{
		{
			name:  "no admin",
			flags: []string{"--market", "41", "--bid-add", "*.kyc"},
			expMsg: &exchange.MsgMarketManageReqAttrsRequest{
				MarketId:                 41,
				CreateAskToAdd:           []string{},
				CreateAskToRemove:        []string{},
				CreateBidToAdd:           []string{"*.kyc"},
				CreateBidToRemove:        []string{},
				CreateCommitmentToAdd:    []string{},
				CreateCommitmentToRemove: []string{},
			},
			expErr: "no <admin> provided",
		},
		{
			name:      "all fields",
			clientCtx: client.Context{FromAddress: sdk.AccAddress("FromAddress_________")},
			flags: []string{
				"--market", "44444",
				"--ask-add", "def.abc,*.xyz", "--ask-remove", "uvw.xyz",
				"--bid-add", "ghi.abc,*.xyz", "--bid-remove", "rst.xyz",
				"--commitment-add", "jkl.abc,*.xyz", "--commitment-remove", "opq.xyz",
			},
			expMsg: &exchange.MsgMarketManageReqAttrsRequest{
				Admin:                    sdk.AccAddress("FromAddress_________").String(),
				MarketId:                 44444,
				CreateAskToAdd:           []string{"def.abc", "*.xyz"},
				CreateAskToRemove:        []string{"uvw.xyz"},
				CreateBidToAdd:           []string{"ghi.abc", "*.xyz"},
				CreateBidToRemove:        []string{"rst.xyz"},
				CreateCommitmentToAdd:    []string{"jkl.abc", "*.xyz"},
				CreateCommitmentToRemove: []string{"opq.xyz"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runTxMakerTestCase(t, td, tc)
		})
	}
}

func TestSetupCmdTxGovCreateMarket(t *testing.T) {
	tc := setupTestCase{
		name:  "SetupCmdTxGovCreateMarket",
		setup: cli.SetupCmdTxGovCreateMarket,
		expFlags: []string{
			cli.FlagAuthority,
			cli.FlagMarket, cli.FlagName, cli.FlagDescription, cli.FlagURL, cli.FlagIcon,
			cli.FlagCreateAsk, cli.FlagCreateBid, cli.FlagCreateCommitment,
			cli.FlagSellerFlat, cli.FlagSellerRatios, cli.FlagBuyerFlat, cli.FlagBuyerRatios,
			cli.FlagAcceptingOrders, cli.FlagAllowUserSettle, cli.FlagAcceptingCommitments, cli.FlagAccessGrants,
			cli.FlagReqAttrAsk, cli.FlagReqAttrBid, cli.FlagReqAttrCommitment,
			cli.FlagBips, cli.FlagDenom,
			cli.FlagProposal,
		},
		expInUse: []string{
			"[--authority <authority>]", "[--market <market id>]",
			"[--name <name>]", "[--description <description>]", "[--url <website url>]", "[--icon <icon uri>]",
			"[--create-ask <coins>]", "[--create-bid <coins>]", "[--create-commitment <coins>]",
			"[--seller-flat <coins>]", "[--seller-ratios <fee ratios>]",
			"[--buyer-flat <coins>]", "[--buyer-ratios <fee ratios>]",
			"[--accepting-orders]", "[--allow-user-settle]", "[--accepting-commitments]",
			"[--access-grants <access grants>]",
			"[--req-attr-ask <attrs>]", "[--req-attr-bid <attrs>]", "[--req-attr-commitment <attrs>]",
			"[--bips <bips>]", "[--denom <denom>]",
			"[--proposal <json filename>",
			cli.AuthorityDesc, cli.RepeatableDesc, cli.AccessGrantsDesc, cli.FeeRatioDesc,
			cli.ProposalFileDesc(&exchange.MsgGovCreateMarketRequest{}),
		},
	}

	oneReqFlags := []string{
		cli.FlagMarket, cli.FlagName, cli.FlagDescription, cli.FlagURL, cli.FlagIcon,
		cli.FlagCreateAsk, cli.FlagCreateBid, cli.FlagCreateCommitment,
		cli.FlagSellerFlat, cli.FlagSellerRatios, cli.FlagBuyerFlat, cli.FlagBuyerRatios,
		cli.FlagAcceptingOrders, cli.FlagAllowUserSettle, cli.FlagAcceptingCommitments, cli.FlagAccessGrants,
		cli.FlagReqAttrAsk, cli.FlagReqAttrBid, cli.FlagReqAttrCommitment,
		cli.FlagBips, cli.FlagDenom,
		cli.FlagProposal,
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

func TestMakeMsgGovCreateMarket(t *testing.T) {
	td := txMakerTestDef[*exchange.MsgGovCreateMarketRequest]{
		makerName: "MakeMsgGovCreateMarket",
		maker:     cli.MakeMsgGovCreateMarket,
		setup:     cli.SetupCmdTxGovCreateMarket,
	}

	tdir := t.TempDir()
	propFN := filepath.Join(tdir, "manage-fees-prop.json")
	fileMsg := &exchange.MsgGovCreateMarketRequest{
		Authority: cli.AuthorityAddr.String(),
		Market: exchange.Market{
			MarketId: 3,
			MarketDetails: exchange.MarketDetails{
				Name:        "A Name",
				Description: "A description.",
				WebsiteUrl:  "A URL",
				IconUri:     "An Icon",
			},
			FeeCreateAskFlat:        []sdk.Coin{sdk.NewInt64Coin("apple", 1)},
			FeeCreateBidFlat:        []sdk.Coin{sdk.NewInt64Coin("banana", 2)},
			FeeSellerSettlementFlat: []sdk.Coin{sdk.NewInt64Coin("cherry", 3)},
			FeeSellerSettlementRatios: []exchange.FeeRatio{
				{Price: sdk.NewInt64Coin("grape", 110), Fee: sdk.NewInt64Coin("grape", 10)},
			},
			FeeBuyerSettlementFlat: []sdk.Coin{sdk.NewInt64Coin("date", 4)},
			FeeBuyerSettlementRatios: []exchange.FeeRatio{
				{Price: sdk.NewInt64Coin("kiwi", 111), Fee: sdk.NewInt64Coin("kiwi", 11)},
			},
			AcceptingOrders:     true,
			AllowUserSettlement: true,
			AccessGrants: []exchange.AccessGrant{
				{
					Address:     sdk.AccAddress("ag1_________________").String(),
					Permissions: []exchange.Permission{2},
				},
			},
			ReqAttrCreateAsk: []string{"ask.create"},
			ReqAttrCreateBid: []string{"bid.create"},

			AcceptingCommitments:     true,
			FeeCreateCommitmentFlat:  []sdk.Coin{sdk.NewInt64Coin("elderberry", 5)},
			CommitmentSettlementBips: 84,
			IntermediaryDenom:        "fig",
			ReqAttrCreateCommitment:  []string{"commitment.create"},
		},
	}
	prop := newGovProp(t, fileMsg)
	tx := newTx(t, prop)
	writeFileAsJson(t, propFN, tx)

	tests := []txMakerTestCase[*exchange.MsgGovCreateMarketRequest]{
		{
			name: "several errors",
			flags: []string{
				"--create-ask", "nope", "--seller-ratios", "8apple",
				"--access-grants", "addr8:set", "--accepting-orders",
			},
			expMsg: &exchange.MsgGovCreateMarketRequest{
				Authority: cli.AuthorityAddr.String(),
				Market: exchange.Market{
					FeeCreateAskFlat:          []sdk.Coin{},
					FeeSellerSettlementRatios: []exchange.FeeRatio{},
					AcceptingOrders:           true,
					AccessGrants:              []exchange.AccessGrant{},
				},
			},
			expErr: joinErrs(
				"invalid coin expression: \"nope\"",
				"cannot create FeeRatio from \"8apple\": expected exactly one colon",
				"could not parse permissions for \"addr8\" from \"set\": invalid permission: \"set\"",
			),
		},
		{
			name:      "all fields",
			clientCtx: client.Context{FromAddress: sdk.AccAddress("FromAddress_________")},
			flags: []string{
				"--market", "18",
				"--create-ask", "10fig", "--create-bid", "5grape", "--create-commitment", "7honeydew",
				"--seller-flat", "12fig", "--seller-ratios", "100prune:1prune",
				"--buyer-flat", "17fig", "--buyer-ratios", "88plum:3plum",
				"--accepting-orders", "--allow-user-settle", "--accepting-commitments",
				"--access-grants", "addr1:settle+cancel", "--access-grants", "addr2:update+permissions",
				"--req-attr-ask", "seller.kyc", "--req-attr-bid", "buyer.kyc", "--req-attr-commitment", "com.kyc",
				"--name", "Special market", "--description", "This market is special.",
				"--url", "https://example.com", "--icon", "https://example.com/icon",
				"--access-grants", "addr3:all",
				"--bips", "47", "--denom", "raisin",
			},
			expMsg: &exchange.MsgGovCreateMarketRequest{
				Authority: cli.AuthorityAddr.String(),
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

					AcceptingCommitments:     true,
					FeeCreateCommitmentFlat:  []sdk.Coin{sdk.NewInt64Coin("honeydew", 7)},
					CommitmentSettlementBips: 47,
					IntermediaryDenom:        "raisin",
					ReqAttrCreateCommitment:  []string{"com.kyc"},
				},
			},
		},
		{
			name:      "proposal flag",
			clientCtx: clientContextWithCodec(client.Context{FromAddress: sdk.AccAddress("FromAddress_________")}),
			flags:     []string{"--proposal", propFN},
			expMsg:    fileMsg,
		},
		{
			name:      "proposal flag with others",
			clientCtx: clientContextWithCodec(client.Context{FromAddress: sdk.AccAddress("FromAddress_________")}),
			flags:     []string{"--proposal", propFN, "--market", "22"},
			expMsg: &exchange.MsgGovCreateMarketRequest{
				Authority: fileMsg.Authority,
				Market: exchange.Market{
					MarketId:                  22,
					MarketDetails:             fileMsg.Market.MarketDetails,
					FeeCreateAskFlat:          fileMsg.Market.FeeCreateAskFlat,
					FeeCreateBidFlat:          fileMsg.Market.FeeCreateBidFlat,
					FeeSellerSettlementFlat:   fileMsg.Market.FeeSellerSettlementFlat,
					FeeSellerSettlementRatios: fileMsg.Market.FeeSellerSettlementRatios,
					FeeBuyerSettlementFlat:    fileMsg.Market.FeeBuyerSettlementFlat,
					FeeBuyerSettlementRatios:  fileMsg.Market.FeeBuyerSettlementRatios,
					AcceptingOrders:           fileMsg.Market.AcceptingOrders,
					AllowUserSettlement:       fileMsg.Market.AllowUserSettlement,
					AccessGrants:              fileMsg.Market.AccessGrants,
					ReqAttrCreateAsk:          fileMsg.Market.ReqAttrCreateAsk,
					ReqAttrCreateBid:          fileMsg.Market.ReqAttrCreateBid,
					AcceptingCommitments:      fileMsg.Market.AcceptingCommitments,
					FeeCreateCommitmentFlat:   fileMsg.Market.FeeCreateCommitmentFlat,
					CommitmentSettlementBips:  fileMsg.Market.CommitmentSettlementBips,
					IntermediaryDenom:         fileMsg.Market.IntermediaryDenom,
					ReqAttrCreateCommitment:   fileMsg.Market.ReqAttrCreateCommitment,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runTxMakerTestCase(t, td, tc)
		})
	}
}

func TestSetupCmdTxGovManageFees(t *testing.T) {
	tc := setupTestCase{
		name:  "SetupCmdTxGovManageFees",
		setup: cli.SetupCmdTxGovManageFees,
		expFlags: []string{
			cli.FlagAuthority, cli.FlagMarket,
			cli.FlagAskAdd, cli.FlagAskRemove, cli.FlagBidAdd, cli.FlagBidRemove,
			cli.FlagSellerFlatAdd, cli.FlagSellerFlatRemove, cli.FlagSellerRatiosAdd, cli.FlagSellerRatiosRemove,
			cli.FlagBuyerFlatAdd, cli.FlagBuyerFlatRemove, cli.FlagBuyerRatiosAdd, cli.FlagBuyerRatiosRemove,
			cli.FlagCommitmentAdd, cli.FlagCommitmentRemove, cli.FlagBips, cli.FlagUnsetBips,
			cli.FlagProposal,
		},
		expAnnotations: map[string]map[string][]string{
			cli.FlagMarket: {required: {"true"}},
		},
		expInUse: []string{
			"--market <market id>", "[--authority <authority>]",
			"[--ask-add <coins>]", "[--ask-remove <coins>]",
			"[--bid-add <coins>]", "[--bid-remove <coins>]",
			"[--commitment-add <coins>]", "[--commitment-remove <coins>]",
			"[--seller-flat-add <coins>]", "[--seller-flat-remove <coins>]",
			"[--seller-ratios-add <fee ratios>]", "[--seller-ratios-remove <fee ratios>]",
			"[--buyer-flat-add <coins>]", "[--buyer-flat-remove <coins>]",
			"[--buyer-ratios-add <fee ratios>]", "[--buyer-ratios-remove <fee ratios>]",
			"[--bips <bips>]", "[--unset-bips]",
			"[--proposal <json filename>",
			cli.AuthorityDesc, cli.RepeatableDesc, cli.FeeRatioDesc,
			cli.ProposalFileDesc(&exchange.MsgGovManageFeesRequest{}),
		},
	}

	oneReqFlags := []string{
		cli.FlagAskAdd, cli.FlagAskRemove, cli.FlagBidAdd, cli.FlagBidRemove,
		cli.FlagSellerFlatAdd, cli.FlagSellerFlatRemove, cli.FlagSellerRatiosAdd, cli.FlagSellerRatiosRemove,
		cli.FlagBuyerFlatAdd, cli.FlagBuyerFlatRemove, cli.FlagBuyerRatiosAdd, cli.FlagBuyerRatiosRemove,
		cli.FlagCommitmentAdd, cli.FlagCommitmentRemove, cli.FlagBips, cli.FlagUnsetBips,
		cli.FlagProposal,
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

func TestMakeMsgGovManageFees(t *testing.T) {
	td := txMakerTestDef[*exchange.MsgGovManageFeesRequest]{
		makerName: "MakeMsgGovManageFees",
		maker:     cli.MakeMsgGovManageFees,
		setup:     cli.SetupCmdTxGovManageFees,
	}

	tdir := t.TempDir()
	propFN := filepath.Join(tdir, "manage-fees-prop.json")
	fileMsg := &exchange.MsgGovManageFeesRequest{
		Authority:                     cli.AuthorityAddr.String(),
		MarketId:                      101,
		AddFeeCreateAskFlat:           []sdk.Coin{sdk.NewInt64Coin("apple", 5)},
		RemoveFeeCreateAskFlat:        []sdk.Coin{sdk.NewInt64Coin("acorn", 6)},
		AddFeeCreateBidFlat:           []sdk.Coin{sdk.NewInt64Coin("banana", 7)},
		RemoveFeeCreateBidFlat:        []sdk.Coin{sdk.NewInt64Coin("blueberry", 8)},
		AddFeeSellerSettlementFlat:    []sdk.Coin{sdk.NewInt64Coin("cherry", 9)},
		RemoveFeeSellerSettlementFlat: []sdk.Coin{sdk.NewInt64Coin("cantaloupe", 10)},
		AddFeeSellerSettlementRatios: []exchange.FeeRatio{
			{Price: sdk.NewInt64Coin("grape", 100), Fee: sdk.NewInt64Coin("grape", 1)},
		},
		RemoveFeeSellerSettlementRatios: []exchange.FeeRatio{
			{Price: sdk.NewInt64Coin("grapefruit", 101), Fee: sdk.NewInt64Coin("grapefruit", 2)},
		},
		AddFeeBuyerSettlementFlat:    []sdk.Coin{sdk.NewInt64Coin("date", 11)},
		RemoveFeeBuyerSettlementFlat: []sdk.Coin{sdk.NewInt64Coin("damson", 12)},
		AddFeeBuyerSettlementRatios: []exchange.FeeRatio{
			{Price: sdk.NewInt64Coin("kiwi", 102), Fee: sdk.NewInt64Coin("kiwi", 3)},
		},
		RemoveFeeBuyerSettlementRatios: []exchange.FeeRatio{
			{Price: sdk.NewInt64Coin("keylime", 104), Fee: sdk.NewInt64Coin("keylime", 4)},
		},
		AddFeeCreateCommitmentFlat:     []sdk.Coin{sdk.NewInt64Coin("lemon", 13)},
		RemoveFeeCreateCommitmentFlat:  []sdk.Coin{sdk.NewInt64Coin("lime", 14)},
		SetFeeCommitmentSettlementBips: 15,
	}
	prop := newGovProp(t, fileMsg)
	tx := newTx(t, prop)
	writeFileAsJson(t, propFN, tx)

	tests := []txMakerTestCase[*exchange.MsgGovManageFeesRequest]{
		{
			name: "multiple errors",
			flags: []string{
				"--ask-add", "15", "--buyer-flat-remove", "noamt",
			},
			expMsg: &exchange.MsgGovManageFeesRequest{
				Authority:                    cli.AuthorityAddr.String(),
				AddFeeCreateAskFlat:          []sdk.Coin{},
				RemoveFeeBuyerSettlementFlat: []sdk.Coin{},
			},
			expErr: joinErrs(
				"invalid coin expression: \"15\"",
				"invalid coin expression: \"noamt\"",
			),
		},
		{
			name:      "all fields",
			clientCtx: client.Context{FromAddress: sdk.AccAddress("FromAddress_________")},
			flags: []string{
				"--market", "55",
				"--ask-add", "18fig", "--ask-remove", "15fig", "--ask-add", "5grape",
				"--bid-add", "17fig", "--bid-remove", "14fig",
				"--seller-flat-add", "55prune", "--seller-flat-remove", "54prune",
				"--seller-ratios-add", "101prune:7prune", "--seller-ratios-remove", "101prune:3prune",
				"--buyer-flat-add", "59prune", "--buyer-flat-remove", "57prune",
				"--buyer-ratios-add", "107prune:1prune", "--buyer-ratios-remove", "43prune:2prune",
				"--commitment-add", "20lychee", "--commitment-remove", "21lingonberry",
				"--bips", "87", "--unset-bips",
			},
			expMsg: &exchange.MsgGovManageFeesRequest{
				Authority:                     cli.AuthorityAddr.String(),
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
				AddFeeCreateCommitmentFlat:       []sdk.Coin{sdk.NewInt64Coin("lychee", 20)},
				RemoveFeeCreateCommitmentFlat:    []sdk.Coin{sdk.NewInt64Coin("lingonberry", 21)},
				SetFeeCommitmentSettlementBips:   87,
				UnsetFeeCommitmentSettlementBips: true,
			},
		},
		{
			name:      "proposal flag",
			clientCtx: clientContextWithCodec(client.Context{FromAddress: sdk.AccAddress("FromAddress_________")}),
			flags:     []string{"--proposal", propFN},
			expMsg:    fileMsg,
		},
		{
			name:      "proposal flag plus others",
			clientCtx: clientContextWithCodec(client.Context{FromAddress: sdk.AccAddress("FromAddress_________")}),
			flags:     []string{"--market", "5", "--proposal", propFN},
			expMsg: &exchange.MsgGovManageFeesRequest{
				Authority:                        fileMsg.Authority,
				MarketId:                         5,
				AddFeeCreateAskFlat:              fileMsg.AddFeeCreateAskFlat,
				RemoveFeeCreateAskFlat:           fileMsg.RemoveFeeCreateAskFlat,
				AddFeeCreateBidFlat:              fileMsg.AddFeeCreateBidFlat,
				RemoveFeeCreateBidFlat:           fileMsg.RemoveFeeCreateBidFlat,
				AddFeeSellerSettlementFlat:       fileMsg.AddFeeSellerSettlementFlat,
				RemoveFeeSellerSettlementFlat:    fileMsg.RemoveFeeSellerSettlementFlat,
				AddFeeSellerSettlementRatios:     fileMsg.AddFeeSellerSettlementRatios,
				RemoveFeeSellerSettlementRatios:  fileMsg.RemoveFeeSellerSettlementRatios,
				AddFeeBuyerSettlementFlat:        fileMsg.AddFeeBuyerSettlementFlat,
				RemoveFeeBuyerSettlementFlat:     fileMsg.RemoveFeeBuyerSettlementFlat,
				AddFeeBuyerSettlementRatios:      fileMsg.AddFeeBuyerSettlementRatios,
				RemoveFeeBuyerSettlementRatios:   fileMsg.RemoveFeeBuyerSettlementRatios,
				AddFeeCreateCommitmentFlat:       fileMsg.AddFeeCreateCommitmentFlat,
				RemoveFeeCreateCommitmentFlat:    fileMsg.RemoveFeeCreateCommitmentFlat,
				SetFeeCommitmentSettlementBips:   fileMsg.SetFeeCommitmentSettlementBips,
				UnsetFeeCommitmentSettlementBips: fileMsg.UnsetFeeCommitmentSettlementBips,
			},
			expErr: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runTxMakerTestCase(t, td, tc)
		})
	}
}

func TestSetupCmdTxGovUpdateParams(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name:  "SetupCmdTxGovUpdateParams",
		setup: cli.SetupCmdTxGovUpdateParams,
		expFlags: []string{
			cli.FlagAuthority, cli.FlagDefault, cli.FlagSplit,
		},
		expAnnotations: map[string]map[string][]string{
			cli.FlagDefault: {required: {"true"}},
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

func TestMakeMsgGovUpdateParams(t *testing.T) {
	td := txMakerTestDef[*exchange.MsgGovUpdateParamsRequest]{
		makerName: "MakeMsgGovUpdateParams",
		maker:     cli.MakeMsgGovUpdateParams,
		setup:     cli.SetupCmdTxGovUpdateParams,
	}

	tests := []txMakerTestCase[*exchange.MsgGovUpdateParamsRequest]{
		{
			name:  "some errors",
			flags: []string{"--split", "jack,14"},
			expMsg: &exchange.MsgGovUpdateParamsRequest{
				Authority: cli.AuthorityAddr.String(),
				Params:    exchange.Params{DenomSplits: []exchange.DenomSplit{}},
			},
			expErr: joinErrs(
				"invalid denom split \"jack\": expected format <denom>:<amount>",
				"invalid denom split \"14\": expected format <denom>:<amount>",
			),
		},
		{
			name:      "no splits",
			clientCtx: client.Context{FromAddress: sdk.AccAddress("FromAddress_________")},
			flags:     []string{"--default", "501"},
			expMsg: &exchange.MsgGovUpdateParamsRequest{
				Authority: cli.AuthorityAddr.String(),
				Params:    exchange.Params{DefaultSplit: 501},
			},
		},
		{
			name:      "all fields",
			clientCtx: client.Context{FromAddress: sdk.AccAddress("FromAddress_________")},
			flags: []string{
				"--split", "banana:99", "--default", "105",
				"--authority", "Jeff", "--split", "apple:333,plum:555"},
			expMsg: &exchange.MsgGovUpdateParamsRequest{
				Authority: "Jeff",
				Params: exchange.Params{
					DefaultSplit: 105,
					DenomSplits: []exchange.DenomSplit{
						{Denom: "banana", Split: 99},
						{Denom: "apple", Split: 333},
						{Denom: "plum", Split: 555},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runTxMakerTestCase(t, td, tc)
		})
	}
}
