package cli_test

import (
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

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

func (s *CmdTestSuite) TestCmdTxCancelOrder() {
	tests := []txCmdTestCase{
		{
			name:     "no order id",
			args:     []string{"cancel-order", "--from", s.addr2.String()},
			expInErr: []string{"no <order id> provided"},
		},
		{
			name: "order does not exist",
			args: []string{"cancel", "18446744073709551615", "--from", s.addr2.String()},
			expInRawLog: []string{"failed to execute message", "invalid request",
				"order 18446744073709551615 does not exist"},
			expectedCode: 18,
		},
		{
			name: "order exists",
			preRun: func() ([]string, func(txResponse *sdk.TxResponse)) {
				newOrder := exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
					MarketId: 5,
					Seller:   s.addr2.String(),
					Assets:   sdk.NewInt64Coin("apple", 100),
					Price:    sdk.NewInt64Coin("peach", 150),
				})
				orderID := s.createOrder(newOrder, nil)
				orderIDStr := orderIDStringer(orderID)
				followup := func(_ *sdk.TxResponse) {
					s.assertGetOrder(orderIDStr, nil)
				}
				return []string{"--order", orderIDStr}, followup
			},
			args:         []string{"cancel", "--from", s.addr2.String()},
			expectedCode: 0,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runTxCmdTestCase(tc)
		})
	}
}

func (s *CmdTestSuite) TestCmdTxFillBids() {
	tests := []txCmdTestCase{
		{
			name:     "no bids",
			args:     []string{"fill-bids", "--from", s.addr3.String(), "--market", "5", "--assets", "100apple"},
			expInErr: []string{"required flag(s) \"bids\" not set"},
		},
		{
			name: "ask order id provided",
			preRun: func() ([]string, func(txResponse *sdk.TxResponse)) {
				askOrder := exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
					MarketId: 5,
					Seller:   s.addr2.String(),
					Assets:   sdk.NewInt64Coin("apple", 100),
					Price:    sdk.NewInt64Coin("peach", 150),
				})
				orderID := s.createOrder(askOrder, &sdk.Coin{Denom: "peach", Amount: sdkmath.NewInt(10)})
				return []string{"--bids", orderIDStringer(orderID)}, nil
			},
			args:         []string{"fill-bids", "--from", s.addr3.String(), "--market", "5", "--assets", "100apple"},
			expInRawLog:  []string{"failed to execute message", "invalid request", "is type ask: expected bid"},
			expectedCode: 18,
		},
		{
			name: "two bids",
			preRun: func() ([]string, func(txResponse *sdk.TxResponse)) {
				creationFee := sdk.NewInt64Coin("peach", 10)
				bid2 := exchange.NewOrder(1).WithBid(&exchange.BidOrder{
					MarketId:            5,
					Buyer:               s.addr2.String(),
					Assets:              sdk.NewInt64Coin("apple", 1000),
					Price:               sdk.NewInt64Coin("peach", 1500),
					BuyerSettlementFees: sdk.NewCoins(sdk.NewInt64Coin("peach", 65)),
				})
				bid3 := exchange.NewOrder(1).WithBid(&exchange.BidOrder{
					MarketId:            5,
					Buyer:               s.addr3.String(),
					Assets:              sdk.NewInt64Coin("apple", 500),
					Price:               sdk.NewInt64Coin("peach", 1000),
					BuyerSettlementFees: sdk.NewCoins(sdk.NewInt64Coin("peach", 60)),
				})
				bid2ID := s.createOrder(bid2, &creationFee)
				bid3ID := s.createOrder(bid3, &creationFee)

				preBalsAddr2 := s.queryBankBalances(s.addr2.String())
				preBalsAddr3 := s.queryBankBalances(s.addr3.String())
				preBalsAddr4 := s.queryBankBalances(s.addr4.String())

				expBals := []banktypes.Balance{
					s.adjustBalance(preBalsAddr2, bid2),
					s.adjustBalance(preBalsAddr3, bid3),
					{
						Address: s.addr4.String(),
						Coins: preBalsAddr4.Add(bid2.GetPrice()).Add(bid3.GetPrice()).
							Sub(bid2.GetAssets()).Sub(bid3.GetAssets()).Sub(s.bondCoins(10)...),
					},
				}

				args := []string{"--bids", orderIDStringer(bid2ID) + "," + orderIDStringer(bid3ID)}
				return args, s.assertBalancesFollowup(expBals)
			},
			args:         []string{"fill-bids", "--from", s.addr4.String(), "--market", "5", "--assets", "1500apple"},
			expectedCode: 0,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runTxCmdTestCase(tc)
		})
	}
}

func (s *CmdTestSuite) TestCmdTxFillAsks() {
	tests := []txCmdTestCase{
		{
			name:     "no asks",
			args:     []string{"fill-asks", "--from", s.addr3.String(), "--market", "5", "--price", "100peach"},
			expInErr: []string{"required flag(s) \"asks\" not set"},
		},
		{
			name: "bid order id provided",
			preRun: func() ([]string, func(txResponse *sdk.TxResponse)) {
				bidOrder := exchange.NewOrder(1).WithBid(&exchange.BidOrder{
					MarketId: 3,
					Buyer:    s.addr2.String(),
					Assets:   sdk.NewInt64Coin("apple", 100),
					Price:    sdk.NewInt64Coin("peach", 150),
				})
				orderID := s.createOrder(bidOrder, &sdk.Coin{Denom: "peach", Amount: sdkmath.NewInt(10)})
				return []string{"--asks", orderIDStringer(orderID)}, nil
			},
			args:         []string{"fill-asks", "--from", s.addr3.String(), "--market", "3", "--price", "150peach"},
			expInRawLog:  []string{"failed to execute message", "invalid request", "is type bid: expected ask"},
			expectedCode: 18,
		},
		{
			name: "two asks",
			preRun: func() ([]string, func(txResponse *sdk.TxResponse)) {
				ask2 := exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
					MarketId: 5,
					Seller:   s.addr2.String(),
					Assets:   sdk.NewInt64Coin("apple", 1000),
					Price:    sdk.NewInt64Coin("peach", 1500),
				})
				ask3 := exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
					MarketId: 5,
					Seller:   s.addr3.String(),
					Assets:   sdk.NewInt64Coin("apple", 500),
					Price:    sdk.NewInt64Coin("peach", 1000),
				})
				ask2ID := s.createOrder(ask2, nil)
				ask3ID := s.createOrder(ask3, nil)

				preBalsAddr2 := s.queryBankBalances(s.addr2.String())
				preBalsAddr3 := s.queryBankBalances(s.addr3.String())
				preBalsAddr4 := s.queryBankBalances(s.addr4.String())

				expBals := []banktypes.Balance{
					s.adjustBalance(preBalsAddr2, ask2),
					s.adjustBalance(preBalsAddr3, ask3),
					{
						Address: s.addr4.String(),
						Coins: preBalsAddr4.Add(ask2.GetAssets()).Add(ask3.GetAssets()).
							Sub(ask2.GetPrice()).Sub(ask3.GetPrice()).
							Sub(sdk.NewInt64Coin("peach", 85)).Sub(s.bondCoins(10)...),
					},
				}

				args := []string{"--asks", orderIDStringer(ask2ID) + "," + orderIDStringer(ask3ID)}
				return args, s.assertBalancesFollowup(expBals)
			},
			args: []string{"fill-asks", "--from", s.addr4.String(), "--market", "5",
				"--price", "2500peach", "--settlement-fee", "75peach", "--creation-fee", "10peach"},
			expectedCode: 0,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runTxCmdTestCase(tc)
		})
	}
}

func (s *CmdTestSuite) TestCmdTxMarketSettle() {
	tests := []txCmdTestCase{
		{
			name:     "no asks",
			args:     []string{"market-settle", "--from", s.addr1.String(), "--market", "5", "--bids", "112,113"},
			expInErr: []string{"required flag(s) \"asks\" not set"},
		},
		{
			name: "endpoint error",
			args: []string{"market-settle", "--from", s.addr9.String(), "--market", "419", "--bids", "18446744073709551614", "--asks", "18446744073709551615"},
			expInRawLog: []string{"failed to execute message", "invalid request",
				"account " + s.addr9.String() + " does not have permission to settle orders for market 419",
			},
			expectedCode: 18,
		},
		{
			name: "two asks two bids",
			preRun: func() ([]string, func(txResponse *sdk.TxResponse)) {
				creationFee := sdk.NewInt64Coin("peach", 10)
				ask5 := exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
					MarketId: 5,
					Seller:   s.addr5.String(),
					Assets:   sdk.NewInt64Coin("apple", 1000),
					Price:    sdk.NewInt64Coin("peach", 1500),
				})
				ask6 := exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
					MarketId: 5,
					Seller:   s.addr6.String(),
					Assets:   sdk.NewInt64Coin("apple", 500),
					Price:    sdk.NewInt64Coin("peach", 1000),
				})
				bid7 := exchange.NewOrder(1).WithBid(&exchange.BidOrder{
					MarketId:            5,
					Buyer:               s.addr7.String(),
					Assets:              sdk.NewInt64Coin("apple", 700),
					Price:               sdk.NewInt64Coin("peach", 1300),
					BuyerSettlementFees: sdk.NewCoins(sdk.NewInt64Coin("peach", 63)),
				})
				bid8 := exchange.NewOrder(1).WithBid(&exchange.BidOrder{
					MarketId:            5,
					Buyer:               s.addr8.String(),
					Assets:              sdk.NewInt64Coin("apple", 800),
					Price:               sdk.NewInt64Coin("peach", 1200),
					BuyerSettlementFees: sdk.NewCoins(sdk.NewInt64Coin("peach", 62)),
				})
				ask5ID := s.createOrder(ask5, nil)
				ask6ID := s.createOrder(ask6, nil)
				bid7ID := s.createOrder(bid7, &creationFee)
				bid8ID := s.createOrder(bid8, &creationFee)

				preBalsAddr5 := s.queryBankBalances(s.addr5.String())
				preBalsAddr6 := s.queryBankBalances(s.addr6.String())
				preBalsAddr7 := s.queryBankBalances(s.addr7.String())
				preBalsAddr8 := s.queryBankBalances(s.addr8.String())

				expBals := []banktypes.Balance{
					s.adjustBalance(preBalsAddr5, ask5),
					s.adjustBalance(preBalsAddr6, ask6),
					s.adjustBalance(preBalsAddr7, bid7),
					s.adjustBalance(preBalsAddr8, bid8),
				}

				args := []string{
					"--asks", orderIDStringer(ask5ID) + "," + orderIDStringer(ask6ID),
					"--bids", orderIDStringer(bid7ID) + "," + orderIDStringer(bid8ID),
				}
				return args, s.assertBalancesFollowup(expBals)
			},
			args:         []string{"settle", "--from", s.addr1.String(), "--market", "5"},
			expectedCode: 0,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runTxCmdTestCase(tc)
		})
	}
}

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
