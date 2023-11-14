package cli_test

import (
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/exchange"
)

func (s *CmdTestSuite) TestCmdTxCreateAsk() {
	tests := []txCmdTestCase{
		{
			name: "cmd error",
			args: []string{"create-ask", "--market", "3",
				"--assets", "10apple", "--price", "20peach",
			},
			expInErr: []string{"at least one of the flags in the group [from seller] is required"},
		},
		{
			name: "insufficient creation fee",
			args: []string{"create-ask", "--market", "3",
				"--assets", "1000apple", "--price", "2000peach",
				"--settlement-fee", "50peach",
				"--creation-fee", "9peach",
				"--from", s.addr2.String(),
			},
			expInRawLog: []string{"failed to execute message", "invalid request",
				"insufficient ask order creation fee: \"9peach\" is less than required amount \"10peach\""},
			expectedCode: 18,
		},
		{
			name: "okay",
			preRun: func() ([]string, func(txResponse *sdk.TxResponse)) {
				expOrder := exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
					MarketId:                3,
					Seller:                  s.addr2.String(),
					Assets:                  sdk.NewInt64Coin("apple", 1000),
					Price:                   sdk.NewInt64Coin("peach", 2000),
					SellerSettlementFlatFee: &sdk.Coin{Denom: "peach", Amount: sdkmath.NewInt(50)},
					AllowPartial:            true,
					ExternalId:              "my-new-ask-order-E2DF6AFE",
				})
				return nil, s.getOrderFollowup(expOrder)
			},
			args: []string{"ask", "--market", "3", "--partial",
				"--assets", "1000apple", "--price", "2000peach",
				"--settlement-fee", "50peach",
				"--creation-fee", "10peach",
				"--from", s.addr2.String(),
				"--external-id", "my-new-ask-order-E2DF6AFE",
			},
			expectedCode: 0,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runTxCmdTestCase(tc)
		})
	}
}

func (s *CmdTestSuite) TestCmdTxCreateBid() {
	tests := []txCmdTestCase{
		{
			name: "cmd error",
			args: []string{"create-bid", "--market", "5",
				"--assets", "10apple", "--price", "20peach",
			},
			expInErr: []string{"at least one of the flags in the group [from buyer] is required"},
		},
		{
			name: "insufficient creation fee",
			args: []string{"create-bid", "--market", "5",
				"--assets", "1000apple", "--price", "2000peach",
				"--settlement-fee", "70peach",
				"--creation-fee", "9peach",
				"--from", s.addr2.String(),
			},
			expInRawLog: []string{"failed to execute message", "invalid request",
				"insufficient bid order creation fee: \"9peach\" is less than required amount \"10peach\""},
			expectedCode: 18,
		},
		{
			name: "okay",
			preRun: func() ([]string, func(txResponse *sdk.TxResponse)) {
				expOrder := exchange.NewOrder(1).WithBid(&exchange.BidOrder{
					MarketId:            5,
					Buyer:               s.addr2.String(),
					Assets:              sdk.NewInt64Coin("apple", 1000),
					Price:               sdk.NewInt64Coin("peach", 2000),
					BuyerSettlementFees: sdk.NewCoins(sdk.NewInt64Coin("peach", 70)),
					AllowPartial:        true,
					ExternalId:          "my-new-bid-order-83A99979",
				})
				return nil, s.getOrderFollowup(expOrder)
			},
			args: []string{"bid", "--market", "5", "--partial",
				"--assets", "1000apple", "--price", "2000peach",
				"--settlement-fee", "70peach",
				"--creation-fee", "10peach",
				"--from", s.addr2.String(),
				"--external-id", "my-new-bid-order-83A99979",
			},
			expectedCode: 0,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runTxCmdTestCase(tc)
		})
	}
}

// TODO[1701]: func (s *CmdTestSuite) TestCmdTxCancelOrder()

// TODO[1701]: func (s *CmdTestSuite) TestCmdTxFillBids()

// TODO[1701]: func (s *CmdTestSuite) TestCmdTxFillAsks()

// TODO[1701]: func (s *CmdTestSuite) TestCmdTxMarketSettle()

// TODO[1701]: func (s *CmdTestSuite) TestCmdTxMarketSetOrderExternalID()

// TODO[1701]: func (s *CmdTestSuite) TestCmdTxMarketWithdraw()

// TODO[1701]: func (s *CmdTestSuite) TestCmdTxMarketUpdateDetails()

// TODO[1701]: func (s *CmdTestSuite) TestCmdTxMarketUpdateEnabled()

// TODO[1701]: func (s *CmdTestSuite) TestCmdTxMarketUpdateUserSettle()

// TODO[1701]: func (s *CmdTestSuite) TestCmdTxMarketManagePermissions()

// TODO[1701]: func (s *CmdTestSuite) TestCmdTxMarketManageReqAttrs()

// TODO[1701]: func (s *CmdTestSuite) TestCmdTxGovCreateMarket()

// TODO[1701]: func (s *CmdTestSuite) TestCmdTxGovManageFees()

// TODO[1701]: func (s *CmdTestSuite) TestCmdTxGovUpdateParams()
