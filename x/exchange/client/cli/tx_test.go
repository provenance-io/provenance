package cli_test

import (
	"bytes"
	"fmt"
	"sort"

	"golang.org/x/exp/maps"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/provenance-io/provenance/x/exchange"
	"github.com/provenance-io/provenance/x/exchange/client/cli"
)

var (
	// invReqCode is the TxResponse code for an ErrInvalidRequest.
	invReqCode = sdkerrors.ErrInvalidRequest.ABCICode()
	// invSigCode is the TxResponse code for an ErrInvalidSigner.
	invSigCode = govtypes.ErrInvalidSigner.ABCICode()
	// insFeeCode is the TxResponse code for an ErrInsufficientFunds.
	insFeeCode = sdkerrors.ErrInsufficientFunds.ABCICode()
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
			expectedCode: invReqCode,
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
				return nil, s.createOrderFollowup(expOrder)
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
			expectedCode: invReqCode,
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
				return nil, s.createOrderFollowup(expOrder)
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

func (s *CmdTestSuite) TestCmdTxCommitFunds() {
	tests := []txCmdTestCase{
		{
			name:     "cmd error",
			args:     []string{"commit", "--market", "5", "--amount", "10apple"},
			expInErr: []string{"at least one of the flags in the group [from account] is required"},
		},
		{
			name: "insufficient creation fee",
			args: []string{"commit-funds", "--market", "5",
				"--amount", "1000apple", "--creation-fee", "14peach",
				"--from", s.addr2.String(),
			},
			expInRawLog: []string{"failed to execute message", "invalid request",
				"insufficient commitment creation fee: \"14peach\" is less than required amount \"15peach\""},
			expectedCode: invReqCode,
		},
		{
			name: "okay",
			preRun: func() ([]string, func(*sdk.TxResponse)) {
				toCommit := sdk.NewCoins(sdk.NewInt64Coin("apple", 1000))
				fee := sdk.NewInt64Coin("peach", 15)
				tag := "committal-0EE2A6EC"
				args := []string{"--amount", toCommit.String(), "--creation-fee", fee.String(), "--tag", tag}

				addr2Bals := s.queryBankBalances(s.addr2.String())
				expBals := []banktypes.Balance{{
					Address: s.addr2.String(),
					Coins:   addr2Bals.Sub(fee, s.bondCoin(10)),
				}}

				addr2Spendable := s.queryBankSpendableBalances(s.addr2.String())
				expSpend := []banktypes.Balance{{
					Address: s.addr2.String(),
					Coins:   addr2Spendable.Sub(fee, s.bondCoin(10)).Sub(toCommit...),
				}}

				expEvent := exchange.NewEventFundsCommitted(s.addr2.String(), 5, toCommit, tag)
				expEvents := sdk.Events{s.untypeEvent(expEvent)}
				s.markAttrsIndexed(expEvents)

				fup := s.composeFollowups(
					s.assertBalancesFollowup(expBals),
					s.assertSpendableBalancesFollowup(expSpend),
					s.assertEventsContains(expEvents),
				)
				return args, fup
			},
			args:         []string{"commit", "--market", "5", "--from", s.addr2.String()},
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
			expectedCode: invReqCode,
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

				return []string{"--order", orderIDStr}, s.getOrderFollowup(orderIDStr, nil)
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
			expectedCode: invReqCode,
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
			expectedCode: invReqCode,
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
			expectedCode: invReqCode,
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

func (s *CmdTestSuite) TestCmdTxMarketCommitmentSettle() {
	tests := []txCmdTestCase{
		{
			name: "cmd error",
			args: []string{"market-commitment-settle",
				"--from", s.addr1.String(),
				"--inputs", s.addr8.String() + ":10apple",
				"--outputs", s.addr9.String() + ":10apple",
			},
			expInErr: []string{"at least one of the flags in the group [file market] is required"},
		},
		{
			name: "market does not exist",
			args: []string{"commitment-settle",
				"--from", s.addr9.String(),
				"--market", "419",
				"--inputs", s.addr8.String() + ":10apple",
				"--outputs", s.addr9.String() + ":10apple",
			},
			expInErr: nil,
			expInRawLog: []string{"failed to execute message", "invalid request",
				"account " + s.addr9.String() + " does not have permission to settle commitments for market 419",
			},
			expectedCode: invReqCode,
		},
		{
			name: "insufficient fees",
			preRun: func() ([]string, func(*sdk.TxResponse)) {
				var marketID uint32 = 3
				amount := sdk.NewCoins(sdk.NewInt64Coin("apple", 41))
				// 41apple * 8cherry/1apple = 328cherry
				// 328cherry * 100<fee> = 32800<fee>
				// 32800<fee> * 50/20000 = 82<fee>
				tag := "iliketomoveitmoveit"
				args := []string{
					"--market", fmt.Sprintf("%d", marketID),
					"--inputs", fmt.Sprintf("%s:%s", s.addr5, amount),
					"--outputs", fmt.Sprintf("%s:%s", s.addr6, amount),
					"--tag", tag,
				}

				s.commitFunds(s.addr5, marketID, amount, nil)

				balsAddr5 := s.queryBankBalances(s.addr5.String())
				balsAddr6 := s.queryBankBalances(s.addr6.String())
				spendBalsAddr5 := s.queryBankSpendableBalances(s.addr5.String())
				spendBalsAddr6 := s.queryBankSpendableBalances(s.addr6.String())

				expBals := []banktypes.Balance{
					{Address: s.addr5.String(), Coins: balsAddr5},
					{Address: s.addr6.String(), Coins: balsAddr6},
				}
				expSpendBals := []banktypes.Balance{
					{Address: s.addr5.String(), Coins: spendBalsAddr5},
					{Address: s.addr6.String(), Coins: spendBalsAddr6},
				}

				fups := s.composeFollowups(
					s.assertBalancesFollowup(expBals),
					s.assertSpendableBalancesFollowup(expSpendBals),
				)
				return args, fups
			},
			args: []string{"market-settle-commitments", "--from", s.addr1.String()},
			expInRawLog: []string{"insufficient funds",
				"negative balance after sending coins to accounts and fee collector",
				"remainingFees: \"10stake\", sentCoins: \"82nhash\"",
			},
			expectedCode: insFeeCode,
		},
		{
			name: "settlement done",
			preRun: func() ([]string, func(*sdk.TxResponse)) {
				var marketID uint32 = 3
				amount := sdk.NewCoins(sdk.NewInt64Coin("apple", 41))
				// 41apple * 8cherry/1apple = 328cherry
				// 328cherry * 100<fee> = 32800<fee>
				// 32800<fee> * 50/20000 = 82<fee>
				tag := "iliketomoveitmoveit"
				args := []string{
					"--market", fmt.Sprintf("%d", marketID),
					"--inputs", fmt.Sprintf("%s:%s", s.addr5, amount),
					"--outputs", fmt.Sprintf("%s:%s", s.addr6, amount),
					"--tag", tag,
				}

				s.commitFunds(s.addr5, marketID, amount, nil)

				balsAddr5 := s.queryBankBalances(s.addr5.String())
				balsAddr6 := s.queryBankBalances(s.addr6.String())
				spendBalsAddr5 := s.queryBankSpendableBalances(s.addr5.String())
				spendBalsAddr6 := s.queryBankSpendableBalances(s.addr6.String())

				expBals := []banktypes.Balance{
					{Address: s.addr5.String(), Coins: balsAddr5.Sub(amount...)},
					{Address: s.addr6.String(), Coins: balsAddr6.Add(amount...)},
				}
				expSpendBals := []banktypes.Balance{
					{Address: s.addr5.String(), Coins: spendBalsAddr5},
					{Address: s.addr6.String(), Coins: spendBalsAddr6},
				}
				expEvents := sdk.Events{
					s.untypeEvent(exchange.NewEventCommitmentReleased(s.addr5.String(), marketID, amount, tag)),
					s.untypeEvent(exchange.NewEventFundsCommitted(s.addr6.String(), marketID, amount, tag)),
				}
				s.markAttrsIndexed(expEvents)

				fups := s.composeFollowups(
					s.assertBalancesFollowup(expBals),
					s.assertSpendableBalancesFollowup(expSpendBals),
					s.assertEventsContains(expEvents),
				)
				return args, fups
			},
			args:         []string{"settle-commitments", "--from", s.addr1.String()},
			addedFees:    s.feeCoins(82),
			expectedCode: 0,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runTxCmdTestCase(tc)
		})
	}
}

func (s *CmdTestSuite) TestCmdTxMarketReleaseCommitments() {
	tests := []txCmdTestCase{
		{
			name: "cmd error",
			args: []string{"market-release-commitments", "--from", s.addr1.String(),
				"--release", s.addr9.String() + ":10apple"},
			expInErr: []string{"required flag(s) \"market\" not set"},
		},
		{
			name: "no permission",
			args: []string{"release-commitments", "--from", s.addr9.String(), "--market", "419",
				"--release-all", s.addr9.String()},
			expInRawLog: []string{"failed to execute message", "invalid request",
				"account " + s.addr9.String() + " does not have permission to release commitments for market 419",
			},
			expectedCode: invReqCode,
		},
		{
			name: "funds released",
			preRun: func() ([]string, func(*sdk.TxResponse)) {
				var marketID uint32 = 3
				toRelease := sdk.NewCoins(sdk.NewInt64Coin("apple", 44), sdk.NewInt64Coin("peach", 76))
				tag := "letitgo"
				addr := s.addr2
				args := []string{
					"--market", fmt.Sprintf("%d", marketID),
					"--release", fmt.Sprintf("%s:%s", addr, toRelease),
					"--tag", tag,
				}

				curSpend := s.queryBankSpendableBalances(addr.String())
				s.commitFunds(addr, marketID, toRelease, nil)
				expSpendBals := []banktypes.Balance{{Address: addr.String(), Coins: curSpend.Sub(s.bondCoin(10))}}

				expEvent := exchange.NewEventCommitmentReleased(addr.String(), marketID, toRelease, tag)
				expEvents := sdk.Events{s.untypeEvent(expEvent)}
				s.markAttrsIndexed(expEvents)

				fup := s.composeFollowups(
					s.assertSpendableBalancesFollowup(expSpendBals),
					s.assertEventsContains(expEvents),
				)
				return args, fup
			},
			args:         []string{"market-release-commitments", "--from", s.addr3.String()},
			expectedCode: 0,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runTxCmdTestCase(tc)
		})
	}
}

func (s *CmdTestSuite) TestCmdTxMarketSetOrderExternalID() {
	tests := []txCmdTestCase{
		{
			name: "no market id",
			args: []string{"market-set-external-id", "--from", s.addr1.String(),
				"--order", "10", "--external-id", "FD6A9038-E15F-4309-ADA6-1AAC3B09DD3E"},
			expInErr: []string{"required flag(s) \"market\" not set"},
		},
		{
			name: "does not have permission",
			preRun: func() ([]string, func(txResponse *sdk.TxResponse)) {
				newOrder := exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
					MarketId:   5,
					Seller:     s.addr7.String(),
					Assets:     sdk.NewInt64Coin("apple", 100),
					Price:      sdk.NewInt64Coin("peach", 100),
					ExternalId: "0A66B2C8-40EF-457A-95B8-5B1D41D020F9",
				})
				orderID := s.createOrder(newOrder, nil)
				orderIDStr := orderIDStringer(orderID)

				return []string{"--order", orderIDStr}, s.getOrderFollowup(orderIDStr, newOrder)
			},
			args: []string{"set-external-id", "--market", "5", "--from", s.addr7.String(),
				"--external-id", "984C9430-7E5E-461A-8468-1F067E26CBE9"},
			expInRawLog: []string{"failed to execute message", "invalid request",
				"account " + s.addr7.String() + " does not have permission to set external ids on orders for market 5",
			},
			expectedCode: invReqCode,
		},
		{
			name: "external id updated",
			preRun: func() ([]string, func(txResponse *sdk.TxResponse)) {
				newOrder := exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
					MarketId:   5,
					Seller:     s.addr7.String(),
					Assets:     sdk.NewInt64Coin("apple", 100),
					Price:      sdk.NewInt64Coin("peach", 100),
					ExternalId: "C0CC7021-A28B-4312-92C9-78DFADC68799",
				})
				orderID := s.createOrder(newOrder, nil)
				orderIDStr := orderIDStringer(orderID)
				newOrder.GetAskOrder().ExternalId = "FF1C3210-D015-4EF8-A397-139E98602365"

				return []string{"--order", orderIDStr}, s.getOrderFollowup(orderIDStr, newOrder)
			},
			args: []string{"market-set-order-external-id", "--from", s.addr1.String(), "--market", "5",
				"--external-id", "FF1C3210-D015-4EF8-A397-139E98602365"},
			expectedCode: 0,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runTxCmdTestCase(tc)
		})
	}
}

func (s *CmdTestSuite) TestCmdTxMarketWithdraw() {
	tests := []txCmdTestCase{
		{
			name: "no market id",
			args: []string{"market-withdraw", "--from", s.addr1.String(),
				"--to", s.addr1.String(), "--amount", "10peach"},
			expInErr: []string{"required flag(s) \"market\" not set"},
		},
		{
			name: "not enough in market account",
			preRun: func() ([]string, func(txResponse *sdk.TxResponse)) {
				market3Addr := exchange.GetMarketAddress(3)
				expBals := []banktypes.Balance{
					{Address: market3Addr.String(), Coins: s.queryBankBalances(market3Addr.String())},
					{Address: s.addr2.String(), Coins: s.queryBankBalances(s.addr2.String())},
				}

				return nil, s.assertBalancesFollowup(expBals)
			},
			args: []string{"market-withdraw", "--from", s.addr1.String(),
				"--market", "3", "--to", s.addr2.String(), "--amount", "50avocado"},
			expInRawLog: []string{"failed to execute message", "invalid request",
				"failed to withdraw 50avocado from market 3",
				"spendable balance 0avocado is smaller than 50avocado",
				"insufficient funds",
			},
			expectedCode: invReqCode,
		},
		{
			name: "success",
			preRun: func() ([]string, func(txResponse *sdk.TxResponse)) {
				amount := sdk.NewInt64Coin("acorn", 50)
				market3Addr := exchange.GetMarketAddress(3)
				s.execBankSend(s.addr8.String(), market3Addr.String(), amount.String())

				preBalsMarket3 := s.queryBankBalances(market3Addr.String())
				preBalsAddr8 := s.queryBankBalances(s.addr8.String())

				expBals := []banktypes.Balance{
					{Address: market3Addr.String(), Coins: preBalsMarket3.Sub(amount)},
					{Address: s.addr8.String(), Coins: preBalsAddr8.Add(amount)},
				}

				return []string{"--amount", amount.String()}, s.assertBalancesFollowup(expBals)
			},
			args:         []string{"withdraw", "--market", "3", "--from", s.addr1.String(), "--to", s.addr8.String()},
			expectedCode: 0,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runTxCmdTestCase(tc)
		})
	}
}

func (s *CmdTestSuite) TestCmdTxMarketUpdateDetails() {
	tests := []txCmdTestCase{
		{
			name:     "no market",
			args:     []string{"market-details", "--from", s.addr1.String(), "--name", "Notgonnawork"},
			expInErr: []string{"required flag(s) \"market\" not set"},
		},
		{
			name: "market does not exist",
			args: []string{"market-update-details", "--market", "419",
				"--from", s.addr1.String(), "--name", "No Such Market"},
			expInRawLog: []string{"failed to execute message", "invalid request",
				"account " + s.addr1.String() + " does not have permission to update market 419",
			},
			expectedCode: invReqCode,
		},
		{
			name: "success",
			preRun: func() ([]string, func(txResponse *sdk.TxResponse)) {
				market3 := s.getMarket("3")
				if len(market3.MarketDetails.IconUri) == 0 {
					market3.MarketDetails.IconUri = "https://example.com/3/icon"
				}
				market3.MarketDetails.IconUri += "?7A9AF177=true"

				args := make([]string, 0, 8)
				if len(market3.MarketDetails.Name) > 0 {
					args = append(args, "--name", market3.MarketDetails.Name)
				}
				if len(market3.MarketDetails.Description) > 0 {
					args = append(args, "--description", market3.MarketDetails.Description)
				}
				if len(market3.MarketDetails.WebsiteUrl) > 0 {
					args = append(args, "--url", market3.MarketDetails.WebsiteUrl)
				}
				if len(market3.MarketDetails.IconUri) > 0 {
					args = append(args, "--icon", market3.MarketDetails.IconUri)
				}

				return args, s.getMarketFollowup("3", market3)
			},
			args:         []string{"market-update-details", "--from", s.addr1.String(), "--market", "3"},
			expectedCode: 0,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runTxCmdTestCase(tc)
		})
	}
}

func (s *CmdTestSuite) TestCmdTxMarketUpdateAcceptingOrders() {
	tests := []txCmdTestCase{
		{
			name:     "no market",
			args:     []string{"market-accepting-orders", "--from", s.addr1.String(), "--enable"},
			expInErr: []string{"required flag(s) \"market\" not set"},
		},
		{
			name: "market does not exist",
			args: []string{"market-update-accepting-orders", "--market", "419",
				"--from", s.addr4.String(), "--enable"},
			expInRawLog: []string{"failed to execute message", "invalid request",
				"account " + s.addr4.String() + " does not have permission to update market 419",
			},
			expectedCode: invReqCode,
		},
		{
			name: "disable market",
			preRun: func() ([]string, func(*sdk.TxResponse)) {
				market420 := s.getMarket("420")
				market420.AcceptingOrders = false
				return nil, s.getMarketFollowup("420", market420)
			},
			args:         []string{"update-accepting-orders", "--disable", "--market", "420", "--from", s.addr1.String()},
			expectedCode: 0,
		},
		{
			name: "enable market",
			preRun: func() ([]string, func(*sdk.TxResponse)) {
				market420 := s.getMarket("420")
				market420.AcceptingOrders = true
				return nil, s.getMarketFollowup("420", market420)
			},
			args:         []string{"update-accepting-orders", "--enable", "--market", "420", "--from", s.addr1.String()},
			expectedCode: 0,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runTxCmdTestCase(tc)
		})
	}
}

func (s *CmdTestSuite) TestCmdTxMarketUpdateUserSettle() {
	tests := []txCmdTestCase{
		{
			name:     "no market",
			args:     []string{"market-user-settle", "--from", s.addr1.String(), "--enable"},
			expInErr: []string{"required flag(s) \"market\" not set"},
		},
		{
			name: "market does not exist",
			args: []string{"market-update-user-settle", "--market", "419",
				"--from", s.addr4.String(), "--enable"},
			expInRawLog: []string{"failed to execute message", "invalid request",
				"account " + s.addr4.String() + " does not have permission to update market 419",
			},
			expectedCode: invReqCode,
		},
		{
			name: "disable user settle",
			preRun: func() ([]string, func(*sdk.TxResponse)) {
				market420 := s.getMarket("420")
				market420.AllowUserSettlement = false
				return nil, s.getMarketFollowup("420", market420)
			},
			args:         []string{"update-user-settle", "--disable", "--market", "420", "--from", s.addr1.String()},
			expectedCode: 0,
		},
		{
			name: "enable user settle",
			preRun: func() ([]string, func(*sdk.TxResponse)) {
				market420 := s.getMarket("420")
				market420.AllowUserSettlement = true
				return nil, s.getMarketFollowup("420", market420)
			},
			args:         []string{"update-user-settle", "--enable", "--market", "420", "--from", s.addr1.String()},
			expectedCode: 0,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runTxCmdTestCase(tc)
		})
	}
}

func (s *CmdTestSuite) TestCmdTxMarketUpdateAcceptingCommitments() {
	tests := []txCmdTestCase{
		{
			name:     "no market",
			args:     []string{"market-accepting-commitments", "--from", s.addr1.String(), "--enable"},
			expInErr: []string{"required flag(s) \"market\" not set"},
		},
		{
			name: "market does not exist",
			args: []string{"market-update-accepting-commitments", "--market", "419",
				"--from", s.addr4.String(), "--enable"},
			expInRawLog: []string{"failed to execute message", "invalid request",
				"account " + s.addr4.String() + " does not have permission to update market 419",
			},
			expectedCode: invReqCode,
		},
		{
			name: "disable market",
			preRun: func() ([]string, func(*sdk.TxResponse)) {
				market420 := s.getMarket("420")
				market420.AcceptingCommitments = false
				return nil, s.getMarketFollowup("420", market420)
			},
			args:         []string{"update-market-accepting-commitments", "--disable", "--market", "420", "--from", s.addr1.String()},
			expectedCode: 0,
		},
		{
			name: "enable market",
			preRun: func() ([]string, func(*sdk.TxResponse)) {
				market420 := s.getMarket("420")
				market420.AcceptingCommitments = true
				return nil, s.getMarketFollowup("420", market420)
			},
			args:         []string{"update-accepting-commitments", "--enable", "--market", "420", "--from", s.addr1.String()},
			expectedCode: 0,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runTxCmdTestCase(tc)
		})
	}
}

func (s *CmdTestSuite) TestCmdTxMarketUpdateIntermediaryDenom() {
	tests := []txCmdTestCase{
		{
			name:     "cmd error",
			args:     []string{"market-intermediary-denom", "--from", s.addr2.String(), "--denom", "banana"},
			expInErr: []string{"required flag(s) \"market\" not set"},
		},
		{
			name: "no permission",
			args: []string{"update-intermediary-denom", "--from", s.addr2.String(),
				"--denom", "banana", "--market", "421"},
			expInRawLog: []string{"failed to execute message", "invalid request",
				"account " + s.addr2.String() + " does not have permission to update market 421",
			},
			expectedCode: invReqCode,
		},
		{
			name: "updated",
			preRun: func() ([]string, func(*sdk.TxResponse)) {
				expMarket := s.getMarket("421")
				expMarket.IntermediaryDenom = "orange"
				return nil, s.getMarketFollowup("421", expMarket)
			},
			args: []string{"update-market-intermediary-denom", "--from", s.addr1.String(),
				"--denom", "orange", "--market", "421"},
			expectedCode: 0,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runTxCmdTestCase(tc)
		})
	}
}

func (s *CmdTestSuite) TestCmdTxMarketManagePermissions() {
	tests := []txCmdTestCase{
		{
			name:     "no market",
			args:     []string{"market-permissions", "--from", s.addr1.String(), "--revoke-all", s.addr8.String()},
			expInErr: []string{"required flag(s) \"market\" not set"},
		},
		{
			name: "market does not exist",
			args: []string{"market-manage-permissions", "--market", "419",
				"--from", s.addr4.String(), "--revoke-all", s.addr2.String()},
			expInRawLog: []string{"failed to execute message", "invalid request",
				"account " + s.addr4.String() + " does not have permission to manage permissions for market 419",
			},
			expectedCode: invReqCode,
		},
		{
			name: "permissions updated",
			preRun: func() ([]string, func(*sdk.TxResponse)) {
				expPerms := map[int][]exchange.Permission{
					1: exchange.AllPermissions(),
					4: {exchange.Permission_permissions},
				}
				for _, perm := range exchange.AllPermissions() {
					if perm != exchange.Permission_cancel {
						expPerms[2] = append(expPerms[2], perm)
					}
				}

				addrOrder := maps.Keys(expPerms)
				sort.Slice(addrOrder, func(i, j int) bool {
					return bytes.Compare(s.accountAddrs[addrOrder[i]], s.accountAddrs[addrOrder[j]]) < 0
				})

				market3 := s.getMarket("3")
				market3.AccessGrants = []exchange.AccessGrant{}
				for _, addrI := range addrOrder {
					market3.AccessGrants = append(market3.AccessGrants, exchange.AccessGrant{
						Address:     s.accountAddrs[addrI].String(),
						Permissions: expPerms[addrI],
					})
				}

				return nil, s.getMarketFollowup("3", market3)
			},
			args: []string{
				"permissions", "--market", "3", "--from", s.addr1.String(),
				"--revoke-all", s.addr3.String(), "--revoke", s.addr2.String() + ":cancel",
				"--grant", s.addr4.String() + ":permissions",
			},
			expectedCode: 0,
		},
		{
			name: "permissions put back",
			preRun: func() ([]string, func(*sdk.TxResponse)) {
				expPerms := map[int][]exchange.Permission{
					1: exchange.AllPermissions(),
					2: exchange.AllPermissions(),
					3: {exchange.Permission_cancel, exchange.Permission_attributes},
				}

				addrOrder := maps.Keys(expPerms)
				sort.Slice(addrOrder, func(i, j int) bool {
					return bytes.Compare(s.accountAddrs[addrOrder[i]], s.accountAddrs[addrOrder[j]]) < 0
				})

				market3 := s.getMarket("3")
				market3.AccessGrants = []exchange.AccessGrant{}
				for _, addrI := range addrOrder {
					market3.AccessGrants = append(market3.AccessGrants, exchange.AccessGrant{
						Address:     s.accountAddrs[addrI].String(),
						Permissions: expPerms[addrI],
					})
				}

				return nil, s.getMarketFollowup("3", market3)
			},
			args: []string{
				"permissions", "--market", "3", "--from", s.addr4.String(),
				"--revoke-all", s.addr2.String() + "," + s.addr4.String(),
				"--grant", s.addr2.String() + ":all",
				"--grant", s.addr3.String() + ":cancel+attributes",
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

func (s *CmdTestSuite) TestCmdTxMarketManageReqAttrs() {
	tests := []txCmdTestCase{
		{
			name:     "no market",
			args:     []string{"market-req-attrs", "--from", s.addr1.String(), "--ask-add", "*.nope"},
			expInErr: []string{"required flag(s) \"market\" not set"},
		},
		{
			name: "market does not exist",
			args: []string{"market-manage-req-attrs", "--market", "419",
				"--from", s.addr4.String(), "--bid-add", "*.also.nope"},
			expInRawLog: []string{"failed to execute message", "invalid request",
				"account " + s.addr4.String() + " does not have permission to manage required attributes for market 419",
			},
			expectedCode: invReqCode,
		},
		{
			name: "req attrs updated",
			preRun: func() ([]string, func(*sdk.TxResponse)) {
				market420 := s.getMarket("420")
				market420.ReqAttrCreateAsk = []string{"seller.kyc", "*.my.attr"}
				market420.ReqAttrCreateBid = []string{}
				return nil, s.getMarketFollowup("420", market420)
			},
			args: []string{"manage-req-attrs", "--from", s.addr1.String(), "--market", "420",
				"--ask-add", "*.my.attr", "--bid-remove", "buyer.kyc"},
			expectedCode: 0,
		},
		{
			name: "req attrs put back",
			preRun: func() ([]string, func(*sdk.TxResponse)) {
				market420 := s.getMarket("420")
				market420.ReqAttrCreateAsk = []string{"seller.kyc"}
				market420.ReqAttrCreateBid = []string{"buyer.kyc"}
				return nil, s.getMarketFollowup("420", market420)
			},
			args: []string{"manage-market-req-attrs", "--from", s.addr1.String(), "--market", "420",
				"--ask-remove", "*.my.attr", "--bid-add", "buyer.kyc"},
			expectedCode: 0,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runTxCmdTestCase(tc)
		})
	}
}

// TODO[1703]: func (s *CmdTestSuite) TestCmdTxCreatePayment()

// TODO[1703]: func (s *CmdTestSuite) TestCmdTxAcceptPayment()

// TODO[1703]: func (s *CmdTestSuite) TestCmdTxRejectPayment()

// TODO[1703]: func (s *CmdTestSuite) TestCmdTxRejectPayments()

// TODO[1703]: func (s *CmdTestSuite) TestCmdTxCancelPayments()

// TODO[1703]: func (s *CmdTestSuite) TestCmdTxChangePaymentTarget()

func (s *CmdTestSuite) TestCmdTxGovCreateMarket() {
	tests := []txCmdTestCase{
		{
			name:     "cmd error",
			args:     []string{"gov-create-market", "--from", s.addr1.String(), "--create-ask", "bananas"},
			expInErr: []string{"invalid coin expression: \"bananas\""},
		},
		{
			name: "wrong authority",
			args: []string{"create-market", "--from", s.addr2.String(), "--authority", s.addr2.String(), "--name", "Whatever"},
			expInRawLog: []string{"failed to execute message",
				s.addr2.String(), "expected gov account as only signer for proposal message",
			},
			expectedCode: invSigCode,
		},
		{
			name: "prop created",
			preRun: func() ([]string, func(*sdk.TxResponse)) {
				expMsg := &exchange.MsgGovCreateMarketRequest{
					Authority: cli.AuthorityAddr.String(),
					Market: exchange.Market{
						MarketId: 0,
						MarketDetails: exchange.MarketDetails{
							Name:        "My New Market",
							Description: "Market 01E6",
						},
						FeeCreateAskFlat:    sdk.NewCoins(sdk.NewInt64Coin("acorn", 100)),
						FeeCreateBidFlat:    sdk.NewCoins(sdk.NewInt64Coin("acorn", 110)),
						AcceptingOrders:     true,
						AllowUserSettlement: false,
						AccessGrants: []exchange.AccessGrant{
							{Address: s.addr4.String(), Permissions: exchange.AllPermissions()},
						},
					},
				}
				return nil, s.govPropFollowup(expMsg)
			},
			args: []string{"create-market", "--from", s.addr4.String(),
				"--name", "My New Market",
				"--description", "Market 01E6",
				"--create-ask", "100acorn", "--create-bid", "110acorn",
				"--accepting-orders", "--access-grants", s.addr4.String() + ":all",
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

func (s *CmdTestSuite) TestCmdTxGovManageFees() {
	tests := []txCmdTestCase{
		{
			name:     "cmd error",
			args:     []string{"gov-manage-fees", "--from", s.addr1.String(), "--ask-add", "bananas", "--market", "12"},
			expInErr: []string{"invalid coin expression: \"bananas\""},
		},
		{
			name: "wrong authority",
			args: []string{"manage-fees", "--from", s.addr2.String(), "--authority", s.addr2.String(),
				"--ask-add", "99banana", "--market", "12"},
			expInRawLog: []string{"failed to execute message",
				s.addr2.String(), "expected gov account as only signer for proposal message",
			},
			expectedCode: invSigCode,
		},
		{
			name: "prop created",
			preRun: func() ([]string, func(*sdk.TxResponse)) {
				expMsg := &exchange.MsgGovManageFeesRequest{
					Authority:                     cli.AuthorityAddr.String(),
					MarketId:                      419,
					AddFeeCreateAskFlat:           []sdk.Coin{sdk.NewInt64Coin("banana", 99)},
					RemoveFeeSellerSettlementFlat: []sdk.Coin{sdk.NewInt64Coin("acorn", 12)},
					AddFeeBuyerSettlementRatios: []exchange.FeeRatio{
						{Price: sdk.NewInt64Coin("plum", 100), Fee: sdk.NewInt64Coin("plum", 1)},
					},
				}
				return nil, s.govPropFollowup(expMsg)
			},
			args: []string{"update-fees", "--from", s.addr4.String(), "--market", "419",
				"--ask-add", "99banana", "--seller-flat-remove", "12acorn",
				"--buyer-ratios-add", "100plum:1plum",
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

func (s *CmdTestSuite) TestCmdTxGovCloseMarket() {
	tests := []txCmdTestCase{
		{
			name:     "cmd error",
			args:     []string{"gov-close-market"},
			expInErr: []string{"required flag(s) \"market\" not set"},
		},
		{
			name: "wrong authority",
			args: []string{"close-market", "--market", "419",
				"--from", s.addr2.String(), "--authority", s.addr2.String()},
			expInRawLog: []string{"failed to execute message",
				s.addr2.String(), "expected gov account as only signer for proposal message",
			},
			expectedCode: invSigCode,
		},
		{
			name: "prop created",
			preRun: func() ([]string, func(*sdk.TxResponse)) {
				expMsg := &exchange.MsgGovCloseMarketRequest{
					Authority: cli.AuthorityAddr.String(),
					MarketId:  419,
				}
				return nil, s.govPropFollowup(expMsg)
			},
			args:         []string{"close-market", "--market", "419", "--from", s.addr2.String()},
			expectedCode: 0,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runTxCmdTestCase(tc)
		})
	}
}

func (s *CmdTestSuite) TestCmdTxGovUpdateParams() {
	tests := []txCmdTestCase{
		{
			name:     "cmd error",
			args:     []string{"gov-update-params", "--from", s.addr1.String(), "--default", "500", "--split", "eight"},
			expInErr: []string{"invalid denom split \"eight\": expected format <denom>:<amount>"},
		},
		{
			name: "wrong authority",
			args: []string{"gov-params", "--from", s.addr2.String(), "--authority", s.addr2.String(), "--default", "500"},
			expInRawLog: []string{"failed to execute message",
				s.addr2.String(), "expected gov account as only signer for proposal message",
			},
			expectedCode: invSigCode,
		},
		{
			name: "prop created",
			preRun: func() ([]string, func(*sdk.TxResponse)) {
				expMsg := &exchange.MsgGovUpdateParamsRequest{
					Authority: cli.AuthorityAddr.String(),
					Params: exchange.Params{
						DefaultSplit: 777,
						DenomSplits: []exchange.DenomSplit{
							{Denom: "apple", Split: 500},
							{Denom: "acorn", Split: 555},
						},
					},
				}
				return nil, s.govPropFollowup(expMsg)
			},
			args: []string{"params", "--from", s.addr4.String(),
				"--default", "777", "--split", "apple:500", "--split", "acorn:555",
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
