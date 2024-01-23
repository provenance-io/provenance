package keeper_test

import (
	"bytes"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/exchange"
	"github.com/provenance-io/provenance/x/exchange/keeper"
)

func (s *TestSuite) TestKeeper_GetOrder() {
	tests := []struct {
		name     string
		setup    func()
		orderID  uint64
		expOrder *exchange.Order
		expErr   string
	}{
		{
			name:     "empty state",
			orderID:  1,
			expOrder: nil,
			expErr:   "",
		},
		{
			name: "order does not exist",
			setup: func() {
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(1).WithBid(&exchange.BidOrder{
					MarketId: 5,
					Buyer:    sdk.AccAddress("some_buyer__________").String(),
					Assets:   s.coin("20apple"),
					Price:    s.coin("160prune"),
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(3).WithAsk(&exchange.AskOrder{
					MarketId: 5,
					Seller:   sdk.AccAddress("some_seller_________").String(),
					Assets:   s.coin("10apple"),
					Price:    s.coin("80prune"),
				}))
			},
			orderID:  2,
			expOrder: nil,
			expErr:   "",
		},
		{
			name: "unknown type byte",
			setup: func() {
				order := exchange.NewOrder(5).WithAsk(&exchange.AskOrder{
					MarketId:     1,
					Seller:       sdk.AccAddress("the_seller__________").String(),
					Assets:       s.coin("10apple"),
					Price:        s.coin("2pineapple"),
					AllowPartial: true,
					ExternalId:   "justsomeid",
				})
				key, value, err := s.k.GetOrderStoreKeyValue(*order)
				s.Require().NoError(err, "GetOrderStoreKeyValue")
				value[0] = 99
				s.getStore().Set(key, value)
			},
			orderID: 5,
			expErr:  "failed to read order 5: unknown type byte 0x63",
		},
		{
			name: "cannot read ask order",
			setup: func() {
				order := exchange.NewOrder(4).WithAsk(&exchange.AskOrder{
					MarketId:     1,
					Seller:       sdk.AccAddress("the_seller__________").String(),
					Assets:       s.coin("10apple"),
					Price:        s.coin("2pineapple"),
					AllowPartial: true,
					ExternalId:   "justsomeid",
				})
				key, value, err := s.k.GetOrderStoreKeyValue(*order)
				s.Require().NoError(err, "GetOrderStoreKeyValue")
				newValue := bytes.Repeat([]byte{3}, len(value))
				newValue[0] = value[0]
				s.getStore().Set(key, newValue)
			},
			orderID:  4,
			expOrder: nil,
			expErr:   "failed to read order 4: failed to unmarshal ask order: proto: AskOrder: illegal tag 0 (wire type 3)",
		},
		{
			name: "cannot read bid order",
			setup: func() {
				order := exchange.NewOrder(4).WithBid(&exchange.BidOrder{
					MarketId:     1,
					Buyer:        sdk.AccAddress("the_buyer___________").String(),
					Assets:       s.coin("10apple"),
					Price:        s.coin("2pineapple"),
					AllowPartial: true,
					ExternalId:   "justsomeid",
				})
				key, value, err := s.k.GetOrderStoreKeyValue(*order)
				s.Require().NoError(err, "GetOrderStoreKeyValue")
				newValue := bytes.Repeat([]byte{3}, len(value))
				newValue[0] = value[0]
				s.getStore().Set(key, newValue)
			},
			orderID:  4,
			expOrder: nil,
			expErr:   "failed to read order 4: failed to unmarshal bid order: proto: BidOrder: illegal tag 0 (wire type 3)",
		},
		{
			name: "order 1 ask",
			setup: func() {
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
					MarketId:                23,
					Seller:                  sdk.AccAddress("seller_for_order_one").String(),
					Assets:                  s.coin("2apple"),
					Price:                   s.coin("53peach"),
					SellerSettlementFlatFee: s.coinP("10fig"),
					AllowPartial:            true,
					ExternalId:              "externalidfororder1",
				}))
			},
			orderID: 1,
			expOrder: exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
				MarketId:                23,
				Seller:                  sdk.AccAddress("seller_for_order_one").String(),
				Assets:                  s.coin("2apple"),
				Price:                   s.coin("53peach"),
				SellerSettlementFlatFee: s.coinP("10fig"),
				AllowPartial:            true,
				ExternalId:              "externalidfororder1",
			}),
		},
		{
			name: "order 1 bid",
			setup: func() {
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(1).WithBid(&exchange.BidOrder{
					MarketId:            23,
					Buyer:               sdk.AccAddress("buyer_for_order_one_").String(),
					Assets:              s.coin("2apple"),
					Price:               s.coin("53peach"),
					BuyerSettlementFees: s.coins("10fig"),
					AllowPartial:        true,
					ExternalId:          "externalidfororder1",
				}))
			},
			orderID: 1,
			expOrder: exchange.NewOrder(1).WithBid(&exchange.BidOrder{
				MarketId:            23,
				Buyer:               sdk.AccAddress("buyer_for_order_one_").String(),
				Assets:              s.coin("2apple"),
				Price:               s.coin("53peach"),
				BuyerSettlementFees: s.coins("10fig"),
				AllowPartial:        true,
				ExternalId:          "externalidfororder1",
			}),
		},
		{
			name: "order max uint32+1 ask",
			setup: func() {
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(4_294_967_295).WithBid(&exchange.BidOrder{
					MarketId: 3,
					Buyer:    sdk.AccAddress("another_buyer_______").String(),
					Assets:   s.coin("5apple"),
					Price:    s.coin("5peach"),
				}))
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(4_294_967_296).WithAsk(&exchange.AskOrder{
					MarketId: 3,
					Seller:   sdk.AccAddress("another_seller______").String(),
					Assets:   s.coin("2apple"),
					Price:    s.coin("53peach"),
				}))
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(4_294_967_297).WithBid(&exchange.BidOrder{
					MarketId: 3,
					Buyer:    sdk.AccAddress("yet_another_buyer___").String(),
					Assets:   s.coin("7apple"),
					Price:    s.coin("7peach"),
				}))
			},
			orderID: 4_294_967_296,
			expOrder: exchange.NewOrder(4_294_967_296).WithAsk(&exchange.AskOrder{
				MarketId: 3,
				Seller:   sdk.AccAddress("another_seller______").String(),
				Assets:   s.coin("2apple"),
				Price:    s.coin("53peach"),
			}),
		},
		{
			name: "order max uint32+1 bid",
			setup: func() {
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(4_294_967_295).WithAsk(&exchange.AskOrder{
					MarketId: 3,
					Seller:   sdk.AccAddress("another_seller______").String(),
					Assets:   s.coin("5apple"),
					Price:    s.coin("5peach"),
				}))
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(4_294_967_296).WithBid(&exchange.BidOrder{
					MarketId: 3,
					Buyer:    sdk.AccAddress("another_buyer_______").String(),
					Assets:   s.coin("2apple"),
					Price:    s.coin("53peach"),
				}))
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(4_294_967_297).WithAsk(&exchange.AskOrder{
					MarketId: 3,
					Seller:   sdk.AccAddress("yet_another_seller__").String(),
					Assets:   s.coin("7apple"),
					Price:    s.coin("7peach"),
				}))
			},
			orderID: 4_294_967_296,
			expOrder: exchange.NewOrder(4_294_967_296).WithBid(&exchange.BidOrder{
				MarketId: 3,
				Buyer:    sdk.AccAddress("another_buyer_______").String(),
				Assets:   s.coin("2apple"),
				Price:    s.coin("53peach"),
			}),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var order *exchange.Order
			var err error
			testFunc := func() {
				order, err = s.k.GetOrder(s.ctx, tc.orderID)
			}
			s.Require().NotPanics(testFunc, "GetOrder(%d)", tc.orderID)
			s.assertErrorValue(err, tc.expErr, "GetOrder(%d) error", tc.orderID)
			s.Assert().Equal(tc.expOrder, order, "GetOrder(%d) order", tc.orderID)
		})
	}
}

func (s *TestSuite) TestKeeper_GetOrderByExternalID() {
	tests := []struct {
		name       string
		setup      func()
		marketID   uint32
		externalID string
		expOrder   *exchange.Order
		expErr     string
	}{
		{
			name:       "market 0",
			marketID:   0,
			externalID: "something",
			expErr:     "invalid market id: cannot be zero",
		},
		{
			name:       "empty externalID",
			marketID:   1,
			externalID: "",
			expOrder:   nil,
			expErr:     "",
		},
		{
			name:       "externalID too long",
			setup:      nil,
			marketID:   1,
			externalID: strings.Repeat("u", exchange.MaxExternalIDLength+1),
			expOrder:   nil,
			expErr:     "",
		},
		{
			name: "unknown externalID",
			setup: func() {
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(7).WithAsk(&exchange.AskOrder{
					MarketId:   1,
					Seller:     sdk.AccAddress("seller______________").String(),
					Assets:     s.coin("4apple"),
					Price:      s.coin("2plum"),
					ExternalId: "seven",
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(8).WithBid(&exchange.BidOrder{
					MarketId:   1,
					Buyer:      sdk.AccAddress("buyer_______________").String(),
					Assets:     s.coin("5apple"),
					Price:      s.coin("3plum"),
					ExternalId: "eight",
				}))
			},
			marketID:   1,
			externalID: "nine",
			expOrder:   nil,
			expErr:     "",
		},
		{
			name: "two orders in market: first",
			setup: func() {
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(7).WithAsk(&exchange.AskOrder{
					MarketId:   3,
					Seller:     sdk.AccAddress("seller______________").String(),
					Assets:     s.coin("4apple"),
					Price:      s.coin("2plum"),
					ExternalId: "seven",
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(8).WithBid(&exchange.BidOrder{
					MarketId:   3,
					Buyer:      sdk.AccAddress("buyer_______________").String(),
					Assets:     s.coin("5apple"),
					Price:      s.coin("3plum"),
					ExternalId: "eight",
				}))
			},
			marketID:   3,
			externalID: "seven",
			expOrder: exchange.NewOrder(7).WithAsk(&exchange.AskOrder{
				MarketId:   3,
				Seller:     sdk.AccAddress("seller______________").String(),
				Assets:     s.coin("4apple"),
				Price:      s.coin("2plum"),
				ExternalId: "seven",
			}),
		},
		{
			name: "two orders in market: second",
			setup: func() {
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(7).WithAsk(&exchange.AskOrder{
					MarketId:   3,
					Seller:     sdk.AccAddress("seller______________").String(),
					Assets:     s.coin("4apple"),
					Price:      s.coin("2plum"),
					ExternalId: "seven",
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(8).WithBid(&exchange.BidOrder{
					MarketId:   3,
					Buyer:      sdk.AccAddress("buyer_______________").String(),
					Assets:     s.coin("5apple"),
					Price:      s.coin("3plum"),
					ExternalId: "eight",
				}))
			},
			marketID:   3,
			externalID: "eight",
			expOrder: exchange.NewOrder(8).WithBid(&exchange.BidOrder{
				MarketId:   3,
				Buyer:      sdk.AccAddress("buyer_______________").String(),
				Assets:     s.coin("5apple"),
				Price:      s.coin("3plum"),
				ExternalId: "eight",
			}),
		},
		{
			name: "externalID in two markets: first",
			setup: func() {
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(7).WithAsk(&exchange.AskOrder{
					MarketId:   55,
					Seller:     sdk.AccAddress("seller______________").String(),
					Assets:     s.coin("4apple"),
					Price:      s.coin("2plum"),
					ExternalId: "specialorder",
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(8).WithBid(&exchange.BidOrder{
					MarketId:   5,
					Buyer:      sdk.AccAddress("buyer_______________").String(),
					Assets:     s.coin("5apple"),
					Price:      s.coin("3plum"),
					ExternalId: "specialorder",
				}))
			},
			marketID:   55,
			externalID: "specialorder",
			expOrder: exchange.NewOrder(7).WithAsk(&exchange.AskOrder{
				MarketId:   55,
				Seller:     sdk.AccAddress("seller______________").String(),
				Assets:     s.coin("4apple"),
				Price:      s.coin("2plum"),
				ExternalId: "specialorder",
			}),
		},
		{
			name: "externalID in two markets: second",
			setup: func() {
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(7).WithAsk(&exchange.AskOrder{
					MarketId:   55,
					Seller:     sdk.AccAddress("seller______________").String(),
					Assets:     s.coin("4apple"),
					Price:      s.coin("2plum"),
					ExternalId: "specialorder",
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(8).WithBid(&exchange.BidOrder{
					MarketId:   5,
					Buyer:      sdk.AccAddress("buyer_______________").String(),
					Assets:     s.coin("5apple"),
					Price:      s.coin("3plum"),
					ExternalId: "specialorder",
				}))
			},
			marketID:   5,
			externalID: "specialorder",
			expOrder: exchange.NewOrder(8).WithBid(&exchange.BidOrder{
				MarketId:   5,
				Buyer:      sdk.AccAddress("buyer_______________").String(),
				Assets:     s.coin("5apple"),
				Price:      s.coin("3plum"),
				ExternalId: "specialorder",
			}),
		},
		{
			name: "externalID in two markets: neither",
			setup: func() {
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(7).WithAsk(&exchange.AskOrder{
					MarketId:   55,
					Seller:     sdk.AccAddress("seller______________").String(),
					Assets:     s.coin("4apple"),
					Price:      s.coin("2plum"),
					ExternalId: "specialorder",
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(8).WithBid(&exchange.BidOrder{
					MarketId:   5,
					Buyer:      sdk.AccAddress("buyer_______________").String(),
					Assets:     s.coin("5apple"),
					Price:      s.coin("3plum"),
					ExternalId: "specialorder",
				}))
			},
			marketID:   15,
			externalID: "specialorder",
			expOrder:   nil,
			expErr:     "",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var order *exchange.Order
			var err error
			testFunc := func() {
				order, err = s.k.GetOrderByExternalID(s.ctx, tc.marketID, tc.externalID)
			}
			s.Require().NotPanics(testFunc, "GetOrderByExternalID(%d, %q)", tc.marketID, tc.externalID)
			s.assertErrorValue(err, tc.expErr, "GetOrderByExternalID(%d, %q) error", tc.marketID, tc.externalID)
			s.Assert().Equal(tc.expOrder, order, "GetOrderByExternalID(%d, %q) order", tc.marketID, tc.externalID)
		})
	}
}

func (s *TestSuite) TestKeeper_CreateAskOrder() {
	reason := func(orderID uint64) string {
		return fmt.Sprintf("x/exchange: order %d", orderID)
	}

	tests := []struct {
		name         string
		attrKeeper   *MockAttributeKeeper
		bankKeeper   *MockBankKeeper
		holdKeeper   *MockHoldKeeper
		setup        func()
		askOrder     exchange.AskOrder
		creationFee  *sdk.Coin
		expOrderID   uint64
		expErr       string
		expBankCalls BankCalls
		expHoldCalls HoldCalls
	}{
		// Tests that result in errors.
		{
			name: "invalid order",
			askOrder: exchange.AskOrder{
				MarketId: 0,
				Seller:   s.addr1.String(),
				Assets:   s.coin("35apple"),
				Price:    s.coin("10peach"),
			},
			expErr: "invalid market id: cannot be zero",
		},
		{
			name: "market does not exist",
			askOrder: exchange.AskOrder{
				MarketId: 2,
				Seller:   s.addr2.String(),
				Assets:   s.coin("35apple"),
				Price:    s.coin("10peach"),
			},
			expErr: "market 2 does not exist",
		},
		{
			name: "market not accepting orders",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:        2,
					AcceptingOrders: false,
				})
			},
			askOrder: exchange.AskOrder{
				MarketId: 2,
				Seller:   s.addr3.String(),
				Assets:   s.coin("35apple"),
				Price:    s.coin("10peach"),
			},
			expErr: "market 2 is not accepting orders",
		},
		{
			name: "attrs required: does not have",
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(s.addr4, []string{"ccc.bb.aa"}, ""),
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:         7,
					AcceptingOrders:  true,
					ReqAttrCreateAsk: []string{"cc.bb.aa"},
				})
			},
			askOrder: exchange.AskOrder{
				MarketId: 7,
				Seller:   s.addr4.String(),
				Assets:   s.coin("35apple"),
				Price:    s.coin("10peach"),
			},
			expErr: "account " + s.addr4.String() + " is not allowed to create ask orders in market 7",
		},
		{
			name: "creation fee required: not enough",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:         2,
					AcceptingOrders:  true,
					FeeCreateAskFlat: s.coins("3fig"),
				})
			},
			askOrder: exchange.AskOrder{
				MarketId: 2,
				Seller:   s.addr4.String(),
				Assets:   s.coin("35apple"),
				Price:    s.coin("10peach"),
			},
			creationFee: s.coinP("2fig"),
			expErr:      "insufficient ask order creation fee: \"2fig\" is less than required amount \"3fig\"",
		},
		{
			name: "settlement fee required: not enough",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:                9,
					AcceptingOrders:         true,
					FeeSellerSettlementFlat: s.coins("6fig"),
				})
			},
			askOrder: exchange.AskOrder{
				MarketId:                9,
				Seller:                  s.addr4.String(),
				Assets:                  s.coin("35apple"),
				Price:                   s.coin("10peach"),
				SellerSettlementFlatFee: s.coinP("4fig"),
			},
			expErr: "insufficient seller settlement flat fee: \"4fig\" is less than required amount \"6fig\"",
		},
		{
			name: "settlement fee denom same as price: price too low",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:                  77,
					AcceptingOrders:           true,
					FeeSellerSettlementFlat:   s.coins("20peach"),
					FeeSellerSettlementRatios: []exchange.FeeRatio{{Price: s.coin("100peach"), Fee: s.coin("51peach")}},
				})
			},
			askOrder: exchange.AskOrder{
				MarketId:                77,
				Seller:                  s.addr4.String(),
				Assets:                  s.coin("35apple"),
				Price:                   s.coin("41peach"),
				SellerSettlementFlatFee: s.coinP("20peach"),
			},
			expErr: "price 41peach is not more than total required seller settlement fee 41peach = 20peach flat + 21peach ratio",
		},
		{
			name:       "cannot collect creation fee",
			bankKeeper: NewMockBankKeeper().WithSendCoinsResults("oh no, an error"),
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:        1,
					AcceptingOrders: true,
				})
			},
			askOrder: exchange.AskOrder{
				MarketId: 1,
				Seller:   s.addr5.String(),
				Assets:   s.coin("35apple"),
				Price:    s.coin("700peach"),
			},
			creationFee: s.coinP("3fig"),
			expErr: "error collecting ask order creation fee: error transferring 3fig from " +
				s.addr5.String() + " to market 1: oh no, an error",
			expBankCalls: BankCalls{
				SendCoins: []*SendCoinsArgs{{fromAddr: s.addr5, toAddr: s.marketAddr1, amt: s.coins("3fig")}},
			},
		},
		{
			name: "external id already in use in this market",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:        2,
					AcceptingOrders: true,
				})
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(15).WithBid(&exchange.BidOrder{
					MarketId:   2,
					Buyer:      s.addr3.String(),
					Assets:     s.coin("100acorn"),
					Price:      s.coin("30plum"),
					ExternalId: "not-that-random-external-id",
				}))
				keeper.SetLastOrderID(store, 33)
			},
			askOrder: exchange.AskOrder{
				MarketId:   2,
				Seller:     s.addr4.String(),
				Assets:     s.coin("500acorn"),
				Price:      s.coin("45plum"),
				ExternalId: "not-that-random-external-id",
			},
			expErr: "error storing ask order: external id \"not-that-random-external-id\" is " +
				"already in use by order 15: cannot be used for order 34",
		},
		{
			name:       "settlement fee denom same as price: cannot place hold on assets",
			holdKeeper: NewMockHoldKeeper().WithAddHoldResults("nope, this is a test error, sorry"),
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:                3,
					AcceptingOrders:         true,
					FeeSellerSettlementFlat: s.coins("20peach"),
				})
			},
			askOrder: exchange.AskOrder{
				MarketId:                3,
				Seller:                  s.addr1.String(),
				Assets:                  s.coin("22apple"),
				Price:                   s.coin("500peach"),
				SellerSettlementFlatFee: s.coinP("20peach"),
			},
			expErr:       "error placing hold for ask order 1: nope, this is a test error, sorry",
			expBankCalls: BankCalls{},
			expHoldCalls: HoldCalls{
				AddHold: []*AddHoldArgs{{
					addr:   s.addr1,
					funds:  s.coins("22apple"),
					reason: "x/exchange: order 1",
				}},
			},
		},
		{
			name:       "settlement fee denom diff from price: cannot place hold on assets and fee",
			holdKeeper: NewMockHoldKeeper().WithAddHoldResults("nope, this is a test error, sorry"),
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:                3,
					AcceptingOrders:         true,
					FeeSellerSettlementFlat: s.coins("20peach,3fig"),
				})
				keeper.SetLastOrderID(s.getStore(), 5)
			},
			askOrder: exchange.AskOrder{
				MarketId:                3,
				Seller:                  s.addr1.String(),
				Assets:                  s.coin("22apple"),
				Price:                   s.coin("500peach"),
				SellerSettlementFlatFee: s.coinP("3fig"),
			},
			expErr:       "error placing hold for ask order 6: nope, this is a test error, sorry",
			expBankCalls: BankCalls{},
			expHoldCalls: HoldCalls{
				AddHold: []*AddHoldArgs{{
					addr:   s.addr1,
					funds:  s.coins("22apple,3fig"),
					reason: reason(6),
				}},
			},
		},

		// Tests that should not give an error.
		{
			name: "no attrs required",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:         3,
					AcceptingOrders:  true,
					FeeCreateAskFlat: s.coins("10fig"),
				})
				keeper.SetLastOrderID(s.getStore(), 50_000)
			},
			askOrder: exchange.AskOrder{
				MarketId: 3,
				Seller:   s.addr3.String(),
				Assets:   s.coin("100apple"),
				Price:    s.coin("3pineapple"),
			},
			creationFee: s.coinP("12fig"),
			expOrderID:  50_001,
			expBankCalls: BankCalls{
				SendCoins: []*SendCoinsArgs{
					{fromAddr: s.addr3, toAddr: s.marketAddr3, amt: s.coins("12fig")},
				},
				SendCoinsFromAccountToModule: []*SendCoinsFromAccountToModuleArgs{
					{senderAddr: s.marketAddr3, recipientModule: s.feeCollector, amt: s.coins("1fig")},
				},
			},
			expHoldCalls: HoldCalls{
				AddHold: []*AddHoldArgs{{addr: s.addr3, funds: s.coins("100apple"), reason: reason(50_001)}},
			},
		},
		{
			name:       "attrs required: has",
			attrKeeper: NewMockAttributeKeeper().WithGetAllAttributesAddrResult(s.addr2, []string{"dd.cc.bb.aa"}, ""),
			setup: func() {
				s.k.SetParams(s.ctx, &exchange.Params{
					DefaultSplit: 200,
					DenomSplits:  []exchange.DenomSplit{{Denom: "fig", Split: 5000}},
				})
				s.requireCreateMarket(exchange.Market{
					MarketId:         3,
					AcceptingOrders:  true,
					FeeCreateAskFlat: s.coins("20fig"),
					ReqAttrCreateAsk: []string{"*.bb.aa"},
				})
				keeper.SetLastOrderID(s.getStore(), 888)
			},
			askOrder: exchange.AskOrder{
				MarketId: 3,
				Seller:   s.addr2.String(),
				Assets:   s.coin("57apple"),
				Price:    s.coin("3pineapple"),
			},
			creationFee: s.coinP("20fig"),
			expOrderID:  889,
			expBankCalls: BankCalls{
				SendCoins: []*SendCoinsArgs{
					{fromAddr: s.addr2, toAddr: s.marketAddr3, amt: s.coins("20fig")},
				},
				SendCoinsFromAccountToModule: []*SendCoinsFromAccountToModuleArgs{
					{senderAddr: s.marketAddr3, recipientModule: s.feeCollector, amt: s.coins("10fig")},
				},
			},
			expHoldCalls: HoldCalls{
				AddHold: []*AddHoldArgs{{addr: s.addr2, funds: s.coins("57apple"), reason: reason(889)}},
			},
		},
		{
			name: "no fees required",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:        1,
					AcceptingOrders: true,
				})
				keeper.SetLastOrderID(s.getStore(), 1)
			},
			askOrder: exchange.AskOrder{
				MarketId: 1,
				Seller:   s.addr4.String(),
				Assets:   s.coin("33apple"),
				Price:    s.coin("57plum"),
			},
			expOrderID:   2,
			expHoldCalls: HoldCalls{AddHold: []*AddHoldArgs{{addr: s.addr4, funds: s.coins("33apple"), reason: reason(2)}}},
		},
		{
			name: "settlement fee denom same as price: hold okay",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:                1,
					AcceptingOrders:         true,
					FeeSellerSettlementFlat: s.coins("20peach"),
				})
				keeper.SetLastOrderID(s.getStore(), 122)
			},
			askOrder: exchange.AskOrder{
				MarketId:                1,
				Seller:                  s.addr3.String(),
				Assets:                  s.coin("500acorn"),
				Price:                   s.coin("100peach"),
				SellerSettlementFlatFee: s.coinP("20peach"),
			},
			expOrderID:   123,
			expHoldCalls: HoldCalls{AddHold: []*AddHoldArgs{{addr: s.addr3, funds: s.coins("500acorn"), reason: reason(123)}}},
		},
		{
			name: "settlement fee denom diff from price: hold okay",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:                1,
					AcceptingOrders:         true,
					FeeSellerSettlementFlat: s.coins("20peach"),
				})
				keeper.SetLastOrderID(s.getStore(), 999)
			},
			askOrder: exchange.AskOrder{
				MarketId:                1,
				Seller:                  s.addr1.String(),
				Assets:                  s.coin("500acorn"),
				Price:                   s.coin("100papaya"),
				SellerSettlementFlatFee: s.coinP("20peach"),
			},
			expOrderID:   1000,
			expHoldCalls: HoldCalls{AddHold: []*AddHoldArgs{{addr: s.addr1, funds: s.coins("500acorn,20peach"), reason: reason(1000)}}},
		},
		{
			name: "external id in use but in different market",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:        1,
					AcceptingOrders: true,
				})
				s.requireCreateMarket(exchange.Market{
					MarketId:        2,
					AcceptingOrders: true,
				})
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(55).WithBid(&exchange.BidOrder{
					MarketId:   1,
					Buyer:      s.addr1.String(),
					Assets:     s.coin("11acorn"),
					Price:      s.coin("55plum"),
					ExternalId: "unoriginal",
				}))
				keeper.SetLastOrderID(store, 98765)
			},
			askOrder: exchange.AskOrder{
				MarketId:   2,
				Seller:     s.addr1.String(),
				Assets:     s.coin("11acorn"),
				Price:      s.coin("55plum"),
				ExternalId: "unoriginal",
			},
			expOrderID:   98766,
			expHoldCalls: HoldCalls{AddHold: []*AddHoldArgs{{addr: s.addr1, funds: s.coins("11acorn"), reason: reason(98766)}}},
		},
		{
			name: "new external id",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:        2,
					AcceptingOrders: true,
				})
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(55).WithBid(&exchange.BidOrder{
					MarketId:   2,
					Buyer:      s.addr1.String(),
					Assets:     s.coin("11acorn"),
					Price:      s.coin("55plum"),
					ExternalId: "C52B5350-BBD6-48B4-9AA7-2F2197260F9D",
				}))
				keeper.SetLastOrderID(store, 65)
			},
			askOrder: exchange.AskOrder{
				MarketId:   2,
				Seller:     s.addr1.String(),
				Assets:     s.coin("11acorn"),
				Price:      s.coin("55plum"),
				ExternalId: "C52B5350-BBD6-48B4-9AA7-2F2197260F9E",
			},
			expOrderID:   66,
			expHoldCalls: HoldCalls{AddHold: []*AddHoldArgs{{addr: s.addr1, funds: s.coins("11acorn"), reason: reason(66)}}},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var expOrder *exchange.Order
			var expEvents sdk.Events
			if len(tc.expErr) == 0 {
				expOrder = exchange.NewOrder(tc.expOrderID).WithAsk(&tc.askOrder)
				event := exchange.NewEventOrderCreated(expOrder)
				expEvents = append(expEvents, s.untypeEvent(event))
			}

			if tc.attrKeeper == nil {
				tc.attrKeeper = NewMockAttributeKeeper()
			}
			if tc.bankKeeper == nil {
				tc.bankKeeper = NewMockBankKeeper()
			}
			if tc.holdKeeper == nil {
				tc.holdKeeper = NewMockHoldKeeper()
			}
			kpr := s.k.WithAttributeKeeper(tc.attrKeeper).WithBankKeeper(tc.bankKeeper).WithHoldKeeper(tc.holdKeeper)

			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)
			var orderID uint64
			var err error
			testFunc := func() {
				orderID, err = kpr.CreateAskOrder(ctx, tc.askOrder, tc.creationFee)
			}
			s.Require().NotPanics(testFunc, "CreateAskOrder")
			s.assertErrorValue(err, tc.expErr, "CreateAskOrder error")
			s.assertEqualOrderID(tc.expOrderID, orderID, "CreateAskOrder order id")
			actEvents := em.Events()
			s.assertEqualEvents(expEvents, actEvents, "CreateAskOrder events")
			s.assertBankKeeperCalls(tc.bankKeeper, tc.expBankCalls, "CreateAskOrder")
			s.assertHoldKeeperCalls(tc.holdKeeper, tc.expHoldCalls, "CreateAskOrder")

			if len(tc.expErr) > 0 || err != nil {
				return
			}

			order, err := s.k.GetOrder(s.ctx, orderID)
			s.Require().NoError(err, "GetOrder(%d) error (the one just created)", orderID)
			s.Assert().Equal(expOrder, order, "GetOrder(%d) (the one just created)", orderID)
			lastOrderID := keeper.GetLastOrderID(s.getStore())
			s.assertEqualOrderID(tc.expOrderID, lastOrderID, "last order id")
		})
	}
}

func (s *TestSuite) TestKeeper_CreateBidOrder() {
	reason := func(orderID uint64) string {
		return fmt.Sprintf("x/exchange: order %d", orderID)
	}

	tests := []struct {
		name         string
		attrKeeper   *MockAttributeKeeper
		bankKeeper   *MockBankKeeper
		holdKeeper   *MockHoldKeeper
		setup        func()
		bidOrder     exchange.BidOrder
		creationFee  *sdk.Coin
		expOrderID   uint64
		expErr       string
		expBankCalls BankCalls
		expHoldCalls HoldCalls
	}{
		// Tests that result in errors.
		{
			name: "invalid order",
			bidOrder: exchange.BidOrder{
				MarketId: 0,
				Buyer:    s.addr1.String(),
				Assets:   s.coin("35apple"),
				Price:    s.coin("10peach"),
			},
			expErr: "invalid market id: cannot be zero",
		},
		{
			name: "market does not exist",
			bidOrder: exchange.BidOrder{
				MarketId: 2,
				Buyer:    s.addr2.String(),
				Assets:   s.coin("35apple"),
				Price:    s.coin("10peach"),
			},
			expErr: "market 2 does not exist",
		},
		{
			name: "market not accepting orders",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:        2,
					AcceptingOrders: false,
				})
			},
			bidOrder: exchange.BidOrder{
				MarketId: 2,
				Buyer:    s.addr3.String(),
				Assets:   s.coin("35apple"),
				Price:    s.coin("10peach"),
			},
			expErr: "market 2 is not accepting orders",
		},
		{
			name: "attrs required: does not have",
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(s.addr4, []string{"ccc.bb.aa"}, ""),
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:         7,
					AcceptingOrders:  true,
					ReqAttrCreateBid: []string{"cc.bb.aa"},
				})
			},
			bidOrder: exchange.BidOrder{
				MarketId: 7,
				Buyer:    s.addr4.String(),
				Assets:   s.coin("35apple"),
				Price:    s.coin("10peach"),
			},
			expErr: "account " + s.addr4.String() + " is not allowed to create bid orders in market 7",
		},
		{
			name: "creation fee required: not enough",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:         2,
					AcceptingOrders:  true,
					FeeCreateBidFlat: s.coins("3fig"),
				})
			},
			bidOrder: exchange.BidOrder{
				MarketId: 2,
				Buyer:    s.addr4.String(),
				Assets:   s.coin("35apple"),
				Price:    s.coin("10peach"),
			},
			creationFee: s.coinP("2fig"),
			expErr:      "insufficient bid order creation fee: \"2fig\" is less than required amount \"3fig\"",
		},
		{
			name: "only buyer flat: not enough",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:               2,
					AcceptingOrders:        true,
					FeeBuyerSettlementFlat: s.coins("10fig"),
				})
			},
			bidOrder: exchange.BidOrder{
				MarketId:            2,
				Buyer:               s.addr2.String(),
				Assets:              s.coin("100acorn"),
				Price:               s.coin("77plum"),
				BuyerSettlementFees: s.coins("9fig"),
			},
			expErr: s.joinErrs(
				"9fig is less than required flat fee 10fig",
				"required flat fee not satisfied, valid options: 10fig",
				"insufficient buyer settlement fee 9fig",
			),
		},
		{
			name: "only buyer ratio: not enough",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:                 2,
					AcceptingOrders:          true,
					FeeBuyerSettlementRatios: s.ratios("100plum:3plum"),
				})
			},
			bidOrder: exchange.BidOrder{
				MarketId:            2,
				Buyer:               s.addr2.String(),
				Assets:              s.coin("100acorn"),
				Price:               s.coin("400plum"),
				BuyerSettlementFees: s.coins("11plum"),
			},
			expErr: s.joinErrs(
				"11plum is less than required ratio fee 12plum (based on price 400plum and ratio 100plum:3plum)",
				"required ratio fee not satisfied, valid ratios: 100plum:3plum",
				"insufficient buyer settlement fee 11plum",
			),
		},
		{
			name: "buyer flat and ratio: not enough",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:                 2,
					AcceptingOrders:          true,
					FeeBuyerSettlementFlat:   s.coins("10plum"),
					FeeBuyerSettlementRatios: s.ratios("100plum:3plum"),
				})
			},
			bidOrder: exchange.BidOrder{
				MarketId:            2,
				Buyer:               s.addr2.String(),
				Assets:              s.coin("100acorn"),
				Price:               s.coin("400plum"),
				BuyerSettlementFees: s.coins("21plum"),
			},
			expErr: s.joinErrs(
				"21plum is less than combined fee 22plum = 10plum (flat) + 12plum (ratio based on price 400plum)",
				"insufficient buyer settlement fee 21plum",
			),
		},
		{
			name:       "cannot collect creation fee",
			bankKeeper: NewMockBankKeeper().WithSendCoinsResults("oh no, an error"),
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:        1,
					AcceptingOrders: true,
				})
			},
			bidOrder: exchange.BidOrder{
				MarketId: 1,
				Buyer:    s.addr5.String(),
				Assets:   s.coin("35apple"),
				Price:    s.coin("700peach"),
			},
			creationFee: s.coinP("3fig"),
			expErr: "error collecting bid order creation fee: error transferring 3fig from " +
				s.addr5.String() + " to market 1: oh no, an error",
			expBankCalls: BankCalls{
				SendCoins: []*SendCoinsArgs{{fromAddr: s.addr5, toAddr: s.marketAddr1, amt: s.coins("3fig")}},
			},
		},
		{
			name: "external id already in use in this market",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:        2,
					AcceptingOrders: true,
				})
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(15).WithAsk(&exchange.AskOrder{
					MarketId:   2,
					Seller:     s.addr3.String(),
					Assets:     s.coin("100acorn"),
					Price:      s.coin("30plum"),
					ExternalId: "not-that-random-external-id",
				}))
				keeper.SetLastOrderID(store, 33)
			},
			bidOrder: exchange.BidOrder{
				MarketId:   2,
				Buyer:      s.addr4.String(),
				Assets:     s.coin("500acorn"),
				Price:      s.coin("45plum"),
				ExternalId: "not-that-random-external-id",
			},
			expErr: "error storing bid order: external id \"not-that-random-external-id\" is " +
				"already in use by order 15: cannot be used for order 34",
		},
		{
			name:       "no settlement fee: cannot place hold",
			holdKeeper: NewMockHoldKeeper().WithAddHoldResults("injected problem"),
			setup: func() {
				s.k.SetParams(s.ctx, &exchange.Params{DefaultSplit: 0})
				s.requireCreateMarket(exchange.Market{
					MarketId:        3,
					AcceptingOrders: true,
				})
				keeper.SetLastOrderID(s.getStore(), 776)
			},
			bidOrder: exchange.BidOrder{
				MarketId: 3,
				Buyer:    s.addr1.String(),
				Assets:   s.coin("11apple"),
				Price:    s.coin("55peach"),
			},
			creationFee:  s.coinP("3fig"),
			expErr:       "error placing hold for bid order 777: injected problem",
			expBankCalls: BankCalls{SendCoins: []*SendCoinsArgs{{fromAddr: s.addr1, toAddr: s.marketAddr3, amt: s.coins("3fig")}}},
			expHoldCalls: HoldCalls{AddHold: []*AddHoldArgs{{addr: s.addr1, funds: s.coins("55peach"), reason: reason(777)}}},
		},
		{
			name:       "with settlement fee: cannot place hold",
			holdKeeper: NewMockHoldKeeper().WithAddHoldResults("injected problem"),
			setup: func() {
				s.k.SetParams(s.ctx, &exchange.Params{DefaultSplit: 0})
				s.requireCreateMarket(exchange.Market{
					MarketId:        3,
					AcceptingOrders: true,
				})
				keeper.SetLastOrderID(s.getStore(), 83483)
			},
			bidOrder: exchange.BidOrder{
				MarketId:            3,
				Buyer:               s.addr1.String(),
				Assets:              s.coin("11apple"),
				Price:               s.coin("55peach"),
				BuyerSettlementFees: s.coins("2peach,5grape"),
			},
			creationFee:  s.coinP("3fig"),
			expErr:       "error placing hold for bid order 83484: injected problem",
			expBankCalls: BankCalls{SendCoins: []*SendCoinsArgs{{fromAddr: s.addr1, toAddr: s.marketAddr3, amt: s.coins("3fig")}}},
			expHoldCalls: HoldCalls{AddHold: []*AddHoldArgs{{addr: s.addr1, funds: s.coins("5grape,57peach"), reason: reason(83484)}}},
		},

		// Tests that should not give an error.
		{
			name: "no attrs required",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:         3,
					AcceptingOrders:  true,
					FeeCreateBidFlat: s.coins("10fig"),
				})
				keeper.SetLastOrderID(s.getStore(), 50_000)
			},
			bidOrder: exchange.BidOrder{
				MarketId: 3,
				Buyer:    s.addr3.String(),
				Assets:   s.coin("100apple"),
				Price:    s.coin("3pineapple"),
			},
			creationFee: s.coinP("12fig"),
			expOrderID:  50_001,
			expBankCalls: BankCalls{
				SendCoins: []*SendCoinsArgs{
					{fromAddr: s.addr3, toAddr: s.marketAddr3, amt: s.coins("12fig")},
				},
				SendCoinsFromAccountToModule: []*SendCoinsFromAccountToModuleArgs{
					{senderAddr: s.marketAddr3, recipientModule: s.feeCollector, amt: s.coins("1fig")},
				},
			},
			expHoldCalls: HoldCalls{
				AddHold: []*AddHoldArgs{{addr: s.addr3, funds: s.coins("3pineapple"), reason: reason(50_001)}},
			},
		},
		{
			name:       "attrs required: has",
			attrKeeper: NewMockAttributeKeeper().WithGetAllAttributesAddrResult(s.addr2, []string{"dd.cc.bb.aa"}, ""),
			setup: func() {
				s.k.SetParams(s.ctx, &exchange.Params{
					DefaultSplit: 200,
					DenomSplits:  []exchange.DenomSplit{{Denom: "fig", Split: 5000}},
				})
				s.requireCreateMarket(exchange.Market{
					MarketId:         3,
					AcceptingOrders:  true,
					FeeCreateBidFlat: s.coins("20fig"),
					ReqAttrCreateBid: []string{"*.bb.aa"},
				})
				keeper.SetLastOrderID(s.getStore(), 888)
			},
			bidOrder: exchange.BidOrder{
				MarketId: 3,
				Buyer:    s.addr2.String(),
				Assets:   s.coin("57apple"),
				Price:    s.coin("3pineapple"),
			},
			creationFee: s.coinP("20fig"),
			expOrderID:  889,
			expBankCalls: BankCalls{
				SendCoins: []*SendCoinsArgs{
					{fromAddr: s.addr2, toAddr: s.marketAddr3, amt: s.coins("20fig")},
				},
				SendCoinsFromAccountToModule: []*SendCoinsFromAccountToModuleArgs{
					{senderAddr: s.marketAddr3, recipientModule: s.feeCollector, amt: s.coins("10fig")},
				},
			},
			expHoldCalls: HoldCalls{
				AddHold: []*AddHoldArgs{{addr: s.addr2, funds: s.coins("3pineapple"), reason: reason(889)}},
			},
		},
		{
			name: "no fees required",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:        1,
					AcceptingOrders: true,
				})
				keeper.SetLastOrderID(s.getStore(), 1)
			},
			bidOrder: exchange.BidOrder{
				MarketId: 1,
				Buyer:    s.addr4.String(),
				Assets:   s.coin("33apple"),
				Price:    s.coin("57plum"),
			},
			expOrderID:   2,
			expHoldCalls: HoldCalls{AddHold: []*AddHoldArgs{{addr: s.addr4, funds: s.coins("57plum"), reason: reason(2)}}},
		},
		{
			name: "no settlement fee: hold okay",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:         1,
					AcceptingOrders:  true,
					FeeCreateBidFlat: s.coins("5fig"),
				})
				keeper.SetLastOrderID(s.getStore(), 122)
			},
			bidOrder: exchange.BidOrder{
				MarketId: 1,
				Buyer:    s.addr3.String(),
				Assets:   s.coin("500acorn"),
				Price:    s.coin("100peach"),
			},
			creationFee: s.coinP("5fig"),
			expOrderID:  123,
			expBankCalls: BankCalls{
				SendCoins: []*SendCoinsArgs{
					{fromAddr: s.addr3, toAddr: s.marketAddr1, amt: s.coins("5fig")},
				},
				SendCoinsFromAccountToModule: []*SendCoinsFromAccountToModuleArgs{
					{senderAddr: s.marketAddr1, recipientModule: s.feeCollector, amt: s.coins("1fig")},
				},
			},
			expHoldCalls: HoldCalls{AddHold: []*AddHoldArgs{{addr: s.addr3, funds: s.coins("100peach"), reason: reason(123)}}},
		},
		{
			name: "with settlement fee: hold okay",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:                 1,
					AcceptingOrders:          true,
					FeeBuyerSettlementFlat:   s.coins("20fig"),
					FeeBuyerSettlementRatios: []exchange.FeeRatio{{Price: s.coin("100peach"), Fee: s.coin("3peach")}},
				})
				keeper.SetLastOrderID(s.getStore(), 999)
			},
			bidOrder: exchange.BidOrder{
				MarketId:            1,
				Buyer:               s.addr1.String(),
				Assets:              s.coin("500acorn"),
				Price:               s.coin("1000peach"),
				BuyerSettlementFees: s.coins("20fig,30peach"),
			},
			expOrderID:   1000,
			expHoldCalls: HoldCalls{AddHold: []*AddHoldArgs{{addr: s.addr1, funds: s.coins("20fig,1030peach"), reason: reason(1000)}}},
		},
		{
			name: "external id in use but in different market",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:        1,
					AcceptingOrders: true,
				})
				s.requireCreateMarket(exchange.Market{
					MarketId:        2,
					AcceptingOrders: true,
				})
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(55).WithAsk(&exchange.AskOrder{
					MarketId:   1,
					Seller:     s.addr1.String(),
					Assets:     s.coin("11acorn"),
					Price:      s.coin("55plum"),
					ExternalId: "unoriginal",
				}))
				keeper.SetLastOrderID(store, 98765)
			},
			bidOrder: exchange.BidOrder{
				MarketId:   2,
				Buyer:      s.addr1.String(),
				Assets:     s.coin("11acorn"),
				Price:      s.coin("55plum"),
				ExternalId: "unoriginal",
			},
			expOrderID:   98766,
			expHoldCalls: HoldCalls{AddHold: []*AddHoldArgs{{addr: s.addr1, funds: s.coins("55plum"), reason: reason(98766)}}},
		},
		{
			name: "new external id",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:        2,
					AcceptingOrders: true,
				})
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(55).WithAsk(&exchange.AskOrder{
					MarketId:   2,
					Seller:     s.addr1.String(),
					Assets:     s.coin("11acorn"),
					Price:      s.coin("55plum"),
					ExternalId: "C52B5350-BBD6-48B4-9AA7-2F2197260F9D",
				}))
				keeper.SetLastOrderID(store, 65)
			},
			bidOrder: exchange.BidOrder{
				MarketId:   2,
				Buyer:      s.addr1.String(),
				Assets:     s.coin("11acorn"),
				Price:      s.coin("55plum"),
				ExternalId: "C52B5350-BBD6-48B4-9AA7-2F2197260F9E",
			},
			expOrderID:   66,
			expHoldCalls: HoldCalls{AddHold: []*AddHoldArgs{{addr: s.addr1, funds: s.coins("55plum"), reason: reason(66)}}},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var expOrder *exchange.Order
			var expEvents sdk.Events
			if len(tc.expErr) == 0 {
				expOrder = exchange.NewOrder(tc.expOrderID).WithBid(&tc.bidOrder)
				event := exchange.NewEventOrderCreated(expOrder)
				expEvents = append(expEvents, s.untypeEvent(event))
			}

			if tc.attrKeeper == nil {
				tc.attrKeeper = NewMockAttributeKeeper()
			}
			if tc.bankKeeper == nil {
				tc.bankKeeper = NewMockBankKeeper()
			}
			if tc.holdKeeper == nil {
				tc.holdKeeper = NewMockHoldKeeper()
			}
			kpr := s.k.WithAttributeKeeper(tc.attrKeeper).WithBankKeeper(tc.bankKeeper).WithHoldKeeper(tc.holdKeeper)

			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)
			var orderID uint64
			var err error
			testFunc := func() {
				orderID, err = kpr.CreateBidOrder(ctx, tc.bidOrder, tc.creationFee)
			}
			s.Require().NotPanics(testFunc, "CreateBidOrder")
			s.assertErrorValue(err, tc.expErr, "CreateBidOrder error")
			s.assertEqualOrderID(tc.expOrderID, orderID, "CreateBidOrder order id")
			actEvents := em.Events()
			s.assertEqualEvents(expEvents, actEvents, "CreateBidOrder events")
			s.assertBankKeeperCalls(tc.bankKeeper, tc.expBankCalls, "CreateBidOrder")
			s.assertHoldKeeperCalls(tc.holdKeeper, tc.expHoldCalls, "CreateBidOrder")

			if len(tc.expErr) > 0 || err != nil {
				return
			}

			order, err := s.k.GetOrder(s.ctx, orderID)
			s.Require().NoError(err, "error from GetOrder(%d) (the one just created)", orderID)
			s.Assert().Equal(expOrder, order, "GetOrder(%d) (the one just created)", orderID)
			lastOrderID := keeper.GetLastOrderID(s.getStore())
			s.assertEqualOrderID(tc.expOrderID, lastOrderID, "last order id")
		})
	}
}

func (s *TestSuite) TestKeeper_CancelOrder() {
	tests := []struct {
		name         string
		holdKeeper   *MockHoldKeeper
		setup        func() *exchange.Order // should return the order expected to be cancelled.
		orderID      uint64
		signer       string
		expErr       string
		expHoldCalls HoldCalls
	}{
		{
			name: "error getting order",
			setup: func() *exchange.Order {
				order := exchange.NewOrder(3).WithBid(&exchange.BidOrder{
					MarketId: 1,
					Buyer:    s.addr3.String(),
					Assets:   s.coin("50apricot"),
					Price:    s.coin("333prune"),
				})
				key, value, err := s.k.GetOrderStoreKeyValue(*order)
				s.Require().NoError(err, "GetOrderStoreKeyValue")
				value[0] = 9
				s.getStore().Set(key, value)
				return nil
			},
			orderID: 3,
			signer:  s.addr3.String(),
			expErr:  "failed to read order 3: unknown type byte 0x9",
		},
		{
			name:    "order does not exist",
			orderID: 55,
			signer:  s.addr1.String(),
			expErr:  "order 55 does not exist",
		},
		{
			name: "signer not allowed",
			setup: func() *exchange.Order {
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(8).WithBid(&exchange.BidOrder{
					MarketId: 1,
					Buyer:    s.addr3.String(),
					Assets:   s.coin("50apricot"),
					Price:    s.coin("333prune"),
				}))
				return nil
			},
			orderID: 8,
			signer:  s.addr2.String(),
			expErr:  "account " + s.addr2.String() + " does not have permission to cancel order 8",
		},
		{
			name:       "error releasing hold",
			holdKeeper: NewMockHoldKeeper().WithReleaseHoldResults("there's not enough here"),
			setup: func() *exchange.Order {
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(7).WithBid(&exchange.BidOrder{
					MarketId: 1,
					Buyer:    s.addr3.String(),
					Assets:   s.coin("50apricot"),
					Price:    s.coin("333prune"),
				}))
				return nil
			},
			orderID:      7,
			signer:       s.addr3.String(),
			expErr:       "unable to release hold on order 7 funds: there's not enough here",
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{{addr: s.addr3, funds: s.coins("333prune")}}},
		},
		{
			name: "signer can cancel in other market but not this one",
			setup: func() *exchange.Order {
				s.requireCreateMarket(exchange.Market{
					MarketId:        1,
					AcceptingOrders: true,
					AccessGrants:    []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_cancel)},
				})
				s.requireCreateMarket(exchange.Market{
					MarketId:        2,
					AcceptingOrders: true,
					AccessGrants:    []exchange.AccessGrant{s.agCanAllBut(s.addr5, exchange.Permission_cancel)},
				})
				s.requireCreateMarket(exchange.Market{
					MarketId:        3,
					AcceptingOrders: true,
					AccessGrants:    []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_cancel)},
				})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(2).WithBid(&exchange.BidOrder{
					MarketId: 2,
					Buyer:    s.addr3.String(),
					Assets:   s.coin("50apricot"),
					Price:    s.coin("333prune"),
				}))
				return nil
			},
			orderID: 2,
			signer:  s.addr5.String(),
			expErr:  "account " + s.addr5.String() + " does not have permission to cancel order 2",
		},
		{
			name: "signer is ask order seller",
			setup: func() *exchange.Order {
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(51).WithBid(&exchange.BidOrder{
					MarketId:   1,
					Buyer:      s.addr1.String(),
					Assets:     s.coin("50apricot"),
					Price:      s.coin("55plum"),
					ExternalId: "order 51",
				}))
				orderToCancel := exchange.NewOrder(52).WithAsk(&exchange.AskOrder{
					MarketId:                1,
					Seller:                  s.addr1.String(),
					Assets:                  s.coin("50apricot"),
					Price:                   s.coin("55plum"),
					SellerSettlementFlatFee: s.coinP("8fig"),
					ExternalId:              "bananas",
				})
				s.requireSetOrderInStore(store, orderToCancel)
				s.requireSetOrderInStore(store, exchange.NewOrder(53).WithBid(&exchange.BidOrder{
					MarketId:   1,
					Buyer:      s.addr1.String(),
					Assets:     s.coin("6apple"),
					Price:      s.coin("55plum"),
					ExternalId: "order 53",
				}))
				return orderToCancel
			},
			orderID:      52,
			signer:       s.addr1.String(),
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{{addr: s.addr1, funds: s.coins("50apricot,8fig")}}},
		},
		{
			name: "signer is bid order buyer",
			setup: func() *exchange.Order {
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(56).WithAsk(&exchange.AskOrder{
					MarketId:   3,
					Seller:     s.addr4.String(),
					Assets:     s.coin("12apple"),
					Price:      s.coin("55plum"),
					ExternalId: "order 56",
				}))
				orderToCancel := exchange.NewOrder(57).WithBid(&exchange.BidOrder{
					MarketId:            3,
					Buyer:               s.addr4.String(),
					Assets:              s.coin("50apricot"),
					Price:               s.coin("55plum"),
					BuyerSettlementFees: s.coins("8fig"),
					ExternalId:          "whatever",
				})
				s.requireSetOrderInStore(store, orderToCancel)
				s.requireSetOrderInStore(store, exchange.NewOrder(58).WithAsk(&exchange.AskOrder{
					MarketId:   3,
					Seller:     s.addr4.String(),
					Assets:     s.coin("13apple"),
					Price:      s.coin("80plum"),
					ExternalId: "order 58",
				}))
				return orderToCancel
			},
			orderID:      57,
			signer:       s.addr4.String(),
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{{addr: s.addr4, funds: s.coins("8fig,55plum")}}},
		},
		{
			name: "signer is authority",
			setup: func() *exchange.Order {
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(99).WithBid(&exchange.BidOrder{
					MarketId:   1,
					Buyer:      s.addr2.String(),
					Assets:     s.coin("12apple"),
					Price:      s.coin("55plum"),
					ExternalId: "order 99",
				}))
				orderToCancel := exchange.NewOrder(100).WithAsk(&exchange.AskOrder{
					MarketId:                2,
					Seller:                  s.addr3.String(),
					Assets:                  s.coin("12apricot"),
					Price:                   s.coin("73pear"),
					SellerSettlementFlatFee: s.coinP("1pear"),
					ExternalId:              "whatever",
				})
				s.requireSetOrderInStore(store, orderToCancel)
				s.requireSetOrderInStore(store, exchange.NewOrder(101).WithAsk(&exchange.AskOrder{
					MarketId:   3,
					Seller:     s.addr5.String(),
					Assets:     s.coin("13apple"),
					Price:      s.coin("80plum"),
					ExternalId: "order 101",
				}))
				return orderToCancel
			},
			orderID:      100,
			signer:       s.k.GetAuthority(),
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{{addr: s.addr3, funds: s.coins("12apricot")}}},
		},
		{
			name: "signer can cancel in market",
			setup: func() *exchange.Order {
				s.requireCreateMarket(exchange.Market{
					MarketId:        1,
					AcceptingOrders: true,
					AccessGrants:    []exchange.AccessGrant{s.agCanOnly(s.addr1, exchange.Permission_cancel)},
				})
				orderToCancel := exchange.NewOrder(999).WithBid(&exchange.BidOrder{
					MarketId:   1,
					Buyer:      s.addr2.String(),
					Assets:     s.coin("12apple"),
					Price:      s.coin("55plum"),
					ExternalId: "order 999",
				})
				s.requireSetOrderInStore(s.getStore(), orderToCancel)
				return orderToCancel
			},
			orderID:      999,
			signer:       s.addr1.String(),
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{{addr: s.addr2, funds: s.coins("55plum")}}},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			var cancelledOrder *exchange.Order
			if tc.setup != nil {
				cancelledOrder = tc.setup()
			}

			var expEvents sdk.Events
			var expDelKVs []sdk.KVPair
			if cancelledOrder != nil {
				event := exchange.NewEventOrderCancelled(cancelledOrder, tc.signer)
				expEvents = append(expEvents, s.untypeEvent(event))

				expDelKVs = keeper.CreateConstantIndexEntries(*cancelledOrder)
				extIDKV := keeper.CreateMarketExternalIDToOrderEntry(cancelledOrder)
				if extIDKV != nil {
					expDelKVs = append(expDelKVs, *extIDKV)
				}
			}

			if tc.holdKeeper == nil {
				tc.holdKeeper = NewMockHoldKeeper()
			}
			kpr := s.k.WithHoldKeeper(tc.holdKeeper)

			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)
			var err error
			testFunc := func() {
				err = kpr.CancelOrder(ctx, tc.orderID, tc.signer)
			}
			s.Require().NotPanics(testFunc, "CancelOrder(%d, %q)", tc.orderID, tc.signer)
			s.assertErrorValue(err, tc.expErr, "CancelOrder(%d, %q) error", tc.orderID, tc.signer)
			actEvents := em.Events()
			s.assertEqualEvents(expEvents, actEvents, "CancelOrder(%d, %q) events", tc.orderID, tc.signer)
			s.assertHoldKeeperCalls(tc.holdKeeper, tc.expHoldCalls, "CancelOrder(%d, %q)", tc.orderID, tc.signer)

			if err != nil || len(tc.expErr) > 0 {
				return
			}

			order, err := s.k.GetOrder(s.ctx, tc.orderID)
			s.Assert().NoError(err, "GetOrder(%d) error after cancel")
			s.Assert().Nil(order, "GetOrder(%d) order after cancel")
			store := s.getStore()
			for i, kv := range expDelKVs {
				has := store.Has(kv.Key)
				s.Assert().False(has, "[%d]: store.Has(%q) (index entry) after cancel", i, kv.Key)
			}
		})
	}
}

func (s *TestSuite) TestKeeper_SetOrderExternalID() {
	tests := []struct {
		name          string
		setup         func() string // should return the original externalID
		marketID      uint32
		orderID       uint64
		newExternalID string
		expErr        string
	}{
		{
			name:          "new external id too long",
			marketID:      1,
			orderID:       1,
			newExternalID: strings.Repeat("I", exchange.MaxExternalIDLength+1),
			expErr: fmt.Sprintf("invalid external id %q: max length %d",
				strings.Repeat("I", exchange.MaxExternalIDLength+1), exchange.MaxExternalIDLength),
		},
		{
			name: "error getting order",
			setup: func() string {
				key, value, err := s.k.GetOrderStoreKeyValue(*exchange.NewOrder(5).WithAsk(&exchange.AskOrder{
					MarketId: 1,
					Seller:   s.addr1.String(),
					Assets:   s.coin("1apple"),
					Price:    s.coin("1pear"),
				}))
				s.Require().NoError(err, "GetOrderStoreKeyValue")
				value[0] = 9
				s.getStore().Set(key, value)
				return ""
			},
			marketID:      1,
			orderID:       5,
			newExternalID: "something",
			expErr:        "failed to read order 5: unknown type byte 0x9",
		},
		{
			name:          "unknown order",
			marketID:      1,
			orderID:       1,
			newExternalID: "",
			expErr:        "order 1 not found",
		},
		{
			name: "wrong market id",
			setup: func() string {
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(3).WithBid(&exchange.BidOrder{
					MarketId: 1,
					Buyer:    s.addr5.String(),
					Assets:   s.coin("7acai"),
					Price:    s.coin("1papaya"),
				}))
				return ""
			},
			marketID:      2,
			orderID:       3,
			newExternalID: "what",
			expErr:        "order 3 has market id 1, expected 2",
		},
		{
			name: "unchanged external id",
			setup: func() string {
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(3).WithBid(&exchange.BidOrder{
					MarketId:   1,
					Buyer:      s.addr5.String(),
					Assets:     s.coin("7acai"),
					Price:      s.coin("1papaya"),
					ExternalId: "thisisfruity",
				}))
				return ""
			},
			marketID:      1,
			orderID:       3,
			newExternalID: "thisisfruity",
			expErr:        "order 3 already has external id \"thisisfruity\"",
		},
		{
			name: "nothing to nothing",
			setup: func() string {
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(3).WithBid(&exchange.BidOrder{
					MarketId:   1,
					Buyer:      s.addr5.String(),
					Assets:     s.coin("7acai"),
					Price:      s.coin("1papaya"),
					ExternalId: "",
				}))
				return ""
			},
			marketID:      1,
			orderID:       3,
			newExternalID: "",
			expErr:        "order 3 already has external id \"\"",
		},
		{
			name: "new external id already exists in market",
			setup: func() string {
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(3).WithBid(&exchange.BidOrder{
					MarketId:   2,
					Buyer:      s.addr5.String(),
					Assets:     s.coin("7acai"),
					Price:      s.coin("1papaya"),
					ExternalId: "duplicate",
				}))
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(4).WithBid(&exchange.BidOrder{
					MarketId:   2,
					Buyer:      s.addr5.String(),
					Assets:     s.coin("7acai"),
					Price:      s.coin("1papaya"),
					ExternalId: "",
				}))
				return ""
			},
			marketID:      2,
			orderID:       4,
			newExternalID: "duplicate",
			expErr:        "external id \"duplicate\" is already in use by order 3: cannot be used for order 4",
		},
		{
			name: "nothing to something: ask",
			setup: func() string {
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(57).WithAsk(&exchange.AskOrder{
					MarketId:   1,
					Seller:     s.addr3.String(),
					Assets:     s.coin("7acai"),
					Price:      s.coin("1papaya"),
					ExternalId: "",
				}))
				return ""
			},
			marketID:      1,
			orderID:       57,
			newExternalID: "something",
		},
		{
			name: "nothing to something: bid",
			setup: func() string {
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(57).WithBid(&exchange.BidOrder{
					MarketId:   1,
					Buyer:      s.addr2.String(),
					Assets:     s.coin("7acai"),
					Price:      s.coin("1papaya"),
					ExternalId: "",
				}))
				return ""
			},
			marketID:      1,
			orderID:       57,
			newExternalID: "something",
		},
		{
			name: "something to nothing: ask",
			setup: func() string {
				// make sure it's okay to have multiple without an external id.
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(57).WithAsk(&exchange.AskOrder{
					MarketId:   1,
					Seller:     s.addr4.String(),
					Assets:     s.coin("7acai"),
					Price:      s.coin("1papaya"),
					ExternalId: "",
				}))
				oldVal := "changeme"
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(58).WithAsk(&exchange.AskOrder{
					MarketId:   1,
					Seller:     s.addr2.String(),
					Assets:     s.coin("7acai"),
					Price:      s.coin("1papaya"),
					ExternalId: oldVal,
				}))
				return oldVal
			},
			marketID:      1,
			orderID:       58,
			newExternalID: "",
		},
		{
			name: "something to nothing: bid",
			setup: func() string {
				// make sure it's okay to have multiple without an external id.
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(57).WithBid(&exchange.BidOrder{
					MarketId:   1,
					Buyer:      s.addr2.String(),
					Assets:     s.coin("7acai"),
					Price:      s.coin("1papaya"),
					ExternalId: "",
				}))
				oldVal := "changeme"
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(58).WithBid(&exchange.BidOrder{
					MarketId:   1,
					Buyer:      s.addr2.String(),
					Assets:     s.coin("7acai"),
					Price:      s.coin("1papaya"),
					ExternalId: oldVal,
				}))
				return oldVal
			},
			marketID:      1,
			orderID:       58,
			newExternalID: "",
		},
		{
			name: "something to something else: ask",
			setup: func() string {
				oldVal := "alterthis"
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(6).WithAsk(&exchange.AskOrder{
					MarketId:   1,
					Seller:     s.addr3.String(),
					Assets:     s.coin("7acai"),
					Price:      s.coin("1papaya"),
					ExternalId: oldVal,
				}))
				return oldVal
			},
			marketID:      1,
			orderID:       6,
			newExternalID: "consideritaltered",
		},
		{
			name: "something to something else: bid",
			setup: func() string {
				oldVal := "alterthis"
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(6).WithBid(&exchange.BidOrder{
					MarketId:   1,
					Buyer:      s.addr2.String(),
					Assets:     s.coin("7acai"),
					Price:      s.coin("1papaya"),
					ExternalId: oldVal,
				}))
				return oldVal
			},
			marketID:      1,
			orderID:       6,
			newExternalID: "consideritaltered",
		},
		{
			name: "new external id exists but in different market",
			setup: func() string {
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(5).WithBid(&exchange.BidOrder{
					MarketId:   1,
					Buyer:      s.addr2.String(),
					Assets:     s.coin("7acai"),
					Price:      s.coin("1papaya"),
					ExternalId: "sharedval",
				}))
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(6).WithBid(&exchange.BidOrder{
					MarketId:   2,
					Buyer:      s.addr2.String(),
					Assets:     s.coin("7acai"),
					Price:      s.coin("1papaya"),
					ExternalId: "",
				}))
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(7).WithBid(&exchange.BidOrder{
					MarketId:   3,
					Buyer:      s.addr2.String(),
					Assets:     s.coin("7acai"),
					Price:      s.coin("1papaya"),
					ExternalId: "sharedval",
				}))
				return ""
			},
			marketID:      2,
			orderID:       6,
			newExternalID: "sharedval",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			var origExternalID string
			if tc.setup != nil {
				origExternalID = tc.setup()
			}

			var expEvents sdk.Events
			var expOrder *exchange.Order
			if len(tc.expErr) == 0 {
				event := &exchange.EventOrderExternalIDUpdated{
					OrderId:    tc.orderID,
					MarketId:   tc.marketID,
					ExternalId: tc.newExternalID,
				}
				expEvents = append(expEvents, s.untypeEvent(event))

				var err error
				expOrder, err = s.k.GetOrder(s.ctx, tc.orderID)
				s.Require().NoError(err, "GetOrder(%d) before anything", tc.orderID)
				switch {
				case expOrder.IsAskOrder():
					askOrder := expOrder.GetAskOrder()
					askOrder.ExternalId = tc.newExternalID
					expOrder = exchange.NewOrder(expOrder.OrderId).WithAsk(askOrder)
				case expOrder.IsBidOrder():
					bidOrder := expOrder.GetBidOrder()
					bidOrder.ExternalId = tc.newExternalID
					expOrder = exchange.NewOrder(expOrder.OrderId).WithBid(bidOrder)
				}
			}

			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)
			var err error
			testFunc := func() {
				err = s.k.SetOrderExternalID(ctx, tc.marketID, tc.orderID, tc.newExternalID)
			}
			s.Require().NotPanics(testFunc, "SetOrderExternalID(%d, %d, %q)", tc.marketID, tc.orderID, tc.newExternalID)
			s.assertErrorValue(err, tc.expErr, "SetOrderExternalID(%d, %d, %q) error", tc.marketID, tc.orderID, tc.newExternalID)
			actEvents := em.Events()
			s.assertEqualEvents(expEvents, actEvents, "SetOrderExternalID(%d, %d, %q) events", tc.marketID, tc.orderID, tc.newExternalID)

			if err != nil || len(tc.expErr) != 0 {
				return
			}

			order, err := s.k.GetOrder(s.ctx, tc.orderID)
			if s.Assert().NoError(err, "GetOrder(%d) error", tc.orderID) {
				s.Assert().Equal(expOrder, order, "GetOrder(%d) result (after SetOrderExternalID)", tc.orderID)
			}
			oldOrder, err := s.k.GetOrderByExternalID(s.ctx, tc.marketID, origExternalID)
			s.Assert().NoError(err, "error from GetOrderByExternalID(%d, %q) (original ExternalID)", tc.marketID, origExternalID)
			s.Assert().Nil(oldOrder, "result from GetOrderByExternalID(%d, %q) (original ExternalID)", tc.marketID, origExternalID)
		})
	}
}

func (s *TestSuite) TestKeeper_IterateOrders() {
	var orders []*exchange.Order
	getAll := func(order *exchange.Order) bool {
		orders = append(orders, order)
		return false
	}
	stopAfter := func(count int) func(order *exchange.Order) bool {
		return func(order *exchange.Order) bool {
			orders = append(orders, order)
			return len(orders) >= count
		}
	}
	addr := func(prefix string, orderID uint64) sdk.AccAddress {
		return sdk.AccAddress(fmt.Sprintf("%s_%d__________________", prefix, orderID)[:20])
	}
	askOrder := func(orderID uint64) *exchange.Order {
		return exchange.NewOrder(orderID).WithAsk(&exchange.AskOrder{
			MarketId:     uint32(orderID / 10),
			Seller:       addr("seller", orderID).String(),
			Assets:       sdk.NewInt64Coin("apple", int64(orderID)),
			Price:        sdk.NewInt64Coin("papaya", int64(orderID)),
			AllowPartial: orderID%2 == 0,
			ExternalId:   fmt.Sprintf("external%d", orderID),
		})
	}
	bidOrder := func(orderID uint64) *exchange.Order {
		return exchange.NewOrder(orderID).WithBid(&exchange.BidOrder{
			MarketId:     uint32(orderID / 10),
			Buyer:        addr("buyer", orderID).String(),
			Assets:       sdk.NewInt64Coin("apple", int64(orderID)),
			Price:        sdk.NewInt64Coin("papaya", int64(orderID)),
			AllowPartial: orderID%2 == 0,
			ExternalId:   fmt.Sprintf("external%d", orderID),
		})
	}

	tests := []struct {
		name      string
		setup     func()
		cb        func(order *exchange.Order) bool
		expErr    string
		expOrders []*exchange.Order
	}{
		{
			name:      "empty state",
			expOrders: nil,
		},
		{
			name: "one order: ask",
			setup: func() {
				s.requireSetOrderInStore(s.getStore(), askOrder(8))
			},
			expOrders: []*exchange.Order{askOrder(8)},
		},
		{
			name: "one order: bid",
			setup: func() {
				s.requireSetOrderInStore(s.getStore(), bidOrder(8))
			},
			expOrders: []*exchange.Order{bidOrder(8)},
		},
		{
			name: "one order: bad key",
			setup: func() {
				key, value, err := s.k.GetOrderStoreKeyValue(*askOrder(4))
				s.Require().NoError(err, "GetOrderStoreKeyValue")
				s.getStore().Set(s.badKey(key), value)
			},
			expErr: "invalid order store key [0 0 0 0 0 0 4]: length expected to be at least 8",
		},
		{
			name: "one order: bad value",
			setup: func() {
				key, value, err := s.k.GetOrderStoreKeyValue(*askOrder(3))
				s.Require().NoError(err, "GetOrderStoreKeyValue")
				value[0] = 8
				s.getStore().Set(key, value)
			},
			expErr: "failed to read order 3: unknown type byte 0x8",
		},
		{
			name: "five orders, 1 through 5: get all",
			setup: func() {
				store := s.getStore()
				s.requireSetOrderInStore(store, askOrder(1))
				s.requireSetOrderInStore(store, bidOrder(2))
				s.requireSetOrderInStore(store, bidOrder(3))
				s.requireSetOrderInStore(store, askOrder(4))
				s.requireSetOrderInStore(store, askOrder(5))
			},
			expOrders: []*exchange.Order{
				askOrder(1), bidOrder(2), bidOrder(3), askOrder(4), askOrder(5),
			},
		},
		{
			name: "five orders, 1 through 5: get one",
			setup: func() {
				store := s.getStore()
				s.requireSetOrderInStore(store, bidOrder(1))
				s.requireSetOrderInStore(store, bidOrder(2))
				s.requireSetOrderInStore(store, askOrder(3))
				s.requireSetOrderInStore(store, bidOrder(4))
				s.requireSetOrderInStore(store, askOrder(5))
			},
			cb:        stopAfter(1),
			expErr:    "",
			expOrders: []*exchange.Order{bidOrder(1)},
		},
		{
			name: "five orders, 1 through 5: get three",
			setup: func() {
				store := s.getStore()
				s.requireSetOrderInStore(store, bidOrder(1))
				s.requireSetOrderInStore(store, askOrder(2))
				s.requireSetOrderInStore(store, askOrder(3))
				s.requireSetOrderInStore(store, bidOrder(4))
				s.requireSetOrderInStore(store, askOrder(5))
			},
			cb:        stopAfter(3),
			expOrders: []*exchange.Order{bidOrder(1), askOrder(2), askOrder(3)},
		},
		{
			name: "five orders, random: get all",
			setup: func() {
				store := s.getStore()
				s.requireSetOrderInStore(store, bidOrder(57))
				s.requireSetOrderInStore(store, bidOrder(78))
				s.requireSetOrderInStore(store, askOrder(83))
				s.requireSetOrderInStore(store, bidOrder(47))
				s.requireSetOrderInStore(store, askOrder(28))
			},
			expOrders: []*exchange.Order{
				askOrder(28), bidOrder(47), bidOrder(57), bidOrder(78), askOrder(83),
			},
		},
		{
			name: "five orders, random: get one",
			setup: func() {
				store := s.getStore()
				s.requireSetOrderInStore(store, bidOrder(57))
				s.requireSetOrderInStore(store, bidOrder(78))
				s.requireSetOrderInStore(store, askOrder(83))
				s.requireSetOrderInStore(store, bidOrder(47))
				s.requireSetOrderInStore(store, askOrder(28))
			},
			cb:        stopAfter(1),
			expOrders: []*exchange.Order{askOrder(28)},
		},
		{
			name: "five orders, random: get three",
			setup: func() {
				store := s.getStore()
				s.requireSetOrderInStore(store, bidOrder(57))
				s.requireSetOrderInStore(store, bidOrder(78))
				s.requireSetOrderInStore(store, askOrder(83))
				s.requireSetOrderInStore(store, bidOrder(47))
				s.requireSetOrderInStore(store, askOrder(28))
			},
			cb: stopAfter(3),
			expOrders: []*exchange.Order{
				askOrder(28), bidOrder(47), bidOrder(57),
			},
		},
		{
			name: "three orders: second bad",
			setup: func() {
				store := s.getStore()
				s.requireSetOrderInStore(store, bidOrder(6))
				key, value, err := s.k.GetOrderStoreKeyValue(*bidOrder(74))
				s.Require().NoError(err, "GetOrderStoreKeyValue")
				value[0] = 8
				store.Set(key, value)
				s.requireSetOrderInStore(store, askOrder(91))
			},
			expErr:    "failed to read order 74: unknown type byte 0x8",
			expOrders: []*exchange.Order{bidOrder(6), askOrder(91)},
		},
		{
			name: "three orders: all bad",
			setup: func() {
				store := s.getStore()

				key6, value6, err := s.k.GetOrderStoreKeyValue(*askOrder(6))
				s.Require().NoError(err, "GetOrderStoreKeyValue 6")
				value6[0] = 6
				store.Set(key6, value6)

				key74, value74, err := s.k.GetOrderStoreKeyValue(*bidOrder(74))
				s.Require().NoError(err, "GetOrderStoreKeyValue 74")
				value74[0] = 74
				store.Set(key74, value74)

				key91, value91, err := s.k.GetOrderStoreKeyValue(*bidOrder(91))
				s.Require().NoError(err, "GetOrderStoreKeyValue 91")
				value91[0] = 91
				store.Set(key91, value91)
			},
			cb: stopAfter(1), // should never get there.
			expErr: s.joinErrs(
				"failed to read order 6: unknown type byte 0x6",
				"failed to read order 74: unknown type byte 0x4a",
				"failed to read order 91: unknown type byte 0x5b",
			),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			if tc.cb == nil {
				tc.cb = getAll
			}

			orders = nil
			var err error
			testFunc := func() {
				err = s.k.IterateOrders(s.ctx, tc.cb)
			}
			s.Require().NotPanics(testFunc, "IterateOrders")
			s.assertErrorValue(err, tc.expErr, "IterateOrders error")
			s.assertEqualOrders(tc.expOrders, orders, "orders iterated")
		})
	}
}

// orderIterCBArgs are the args provided to an order index iterator.
type orderIterCBArgs struct {
	orderID       uint64
	orderTypeByte byte
}

// orderIDString gets this order id as a string.
func (a orderIterCBArgs) orderIDString() string {
	return fmt.Sprintf("%d", a.orderID)
}

// newOrderIterCBArgs creates a new orderIterCBArgs.
func newOrderIterCBArgs(orderID uint64, orderTypeByte byte) orderIterCBArgs {
	return orderIterCBArgs{
		orderID:       orderID,
		orderTypeByte: orderTypeByte,
	}
}

func (s *TestSuite) TestKeeper_IterateMarketOrders() {
	var seen []orderIterCBArgs
	getAll := func(orderID uint64, orderTypeByte byte) bool {
		seen = append(seen, newOrderIterCBArgs(orderID, orderTypeByte))
		return false
	}
	stopAfter := func(count int) func(orderID uint64, orderTypeByte byte) bool {
		return func(orderID uint64, orderTypeByte byte) bool {
			seen = append(seen, newOrderIterCBArgs(orderID, orderTypeByte))
			return len(seen) >= count
		}
	}

	tests := []struct {
		name     string
		setup    func()
		marketID uint32
		cb       func(orderID uint64, orderTypeByte byte) bool
		expSeen  []orderIterCBArgs
	}{
		{
			name:     "empty state",
			marketID: 3,
			expSeen:  nil,
		},
		{
			name: "no orders in market",
			setup: func() {
				store := s.getStore()
				store.Set(keeper.MakeIndexKeyMarketToOrder(1, 1), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyMarketToOrder(1, 2), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyMarketToOrder(1, 3), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyMarketToOrder(3, 5), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyMarketToOrder(3, 6), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyMarketToOrder(3, 7), []byte{keeper.OrderKeyTypeAsk})
			},
			marketID: 2,
			expSeen:  nil,
		},
		{
			name: "one entry: ask",
			setup: func() {
				store := s.getStore()
				store.Set(keeper.MakeIndexKeyMarketToOrder(1, 1), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyMarketToOrder(1, 2), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyMarketToOrder(1, 3), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyMarketToOrder(2, 4), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyMarketToOrder(3, 5), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyMarketToOrder(3, 6), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyMarketToOrder(3, 7), []byte{keeper.OrderKeyTypeAsk})
			},
			marketID: 2,
			expSeen:  []orderIterCBArgs{newOrderIterCBArgs(4, keeper.OrderKeyTypeAsk)},
		},
		{
			name: "one entry: bid",
			setup: func() {
				store := s.getStore()
				store.Set(keeper.MakeIndexKeyMarketToOrder(1, 1), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyMarketToOrder(1, 2), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyMarketToOrder(1, 3), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyMarketToOrder(2, 4), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyMarketToOrder(3, 5), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyMarketToOrder(3, 6), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyMarketToOrder(3, 7), []byte{keeper.OrderKeyTypeAsk})
			},
			marketID: 2,
			expSeen:  []orderIterCBArgs{newOrderIterCBArgs(4, keeper.OrderKeyTypeBid)},
		},
		{
			name: "one entry no value",
			setup: func() {
				s.getStore().Set(keeper.MakeIndexKeyMarketToOrder(2, 4), []byte{})
			},
			marketID: 2,
			expSeen:  nil,
		},
		{
			name: "one entry bad key",
			setup: func() {
				s.getStore().Set(s.badKey(keeper.MakeIndexKeyMarketToOrder(2, 4)), []byte{keeper.OrderKeyTypeAsk})
			},
			marketID: 2,
			expSeen:  nil,
		},
		{
			name: "five entries, 1 through 5: get all",
			setup: func() {
				store := s.getStore()
				store.Set(keeper.MakeIndexKeyMarketToOrder(4, 1), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyMarketToOrder(4, 2), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyMarketToOrder(4, 3), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyMarketToOrder(4, 4), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyMarketToOrder(4, 5), []byte{keeper.OrderKeyTypeBid})
			},
			marketID: 4,
			expSeen: []orderIterCBArgs{
				newOrderIterCBArgs(1, keeper.OrderKeyTypeBid),
				newOrderIterCBArgs(2, keeper.OrderKeyTypeBid),
				newOrderIterCBArgs(3, keeper.OrderKeyTypeBid),
				newOrderIterCBArgs(4, keeper.OrderKeyTypeAsk),
				newOrderIterCBArgs(5, keeper.OrderKeyTypeBid),
			},
		},
		{
			name: "five entries, 1 through 5: get one",
			setup: func() {
				store := s.getStore()
				store.Set(keeper.MakeIndexKeyMarketToOrder(4, 1), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyMarketToOrder(4, 2), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyMarketToOrder(4, 3), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyMarketToOrder(4, 4), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyMarketToOrder(4, 5), []byte{keeper.OrderKeyTypeAsk})
			},
			marketID: 4,
			cb:       stopAfter(1),
			expSeen:  []orderIterCBArgs{newOrderIterCBArgs(1, keeper.OrderKeyTypeBid)},
		},
		{
			name: "five entries, 1 through 5: get three",
			setup: func() {
				store := s.getStore()
				store.Set(keeper.MakeIndexKeyMarketToOrder(4, 1), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyMarketToOrder(4, 2), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyMarketToOrder(4, 3), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyMarketToOrder(4, 4), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyMarketToOrder(4, 5), []byte{keeper.OrderKeyTypeBid})
			},
			marketID: 4,
			cb:       stopAfter(3),
			expSeen: []orderIterCBArgs{
				newOrderIterCBArgs(1, keeper.OrderKeyTypeAsk),
				newOrderIterCBArgs(2, keeper.OrderKeyTypeBid),
				newOrderIterCBArgs(3, keeper.OrderKeyTypeAsk),
			},
		},
		{
			name: "five entries, random: get all",
			setup: func() {
				store := s.getStore()
				store.Set(keeper.MakeIndexKeyMarketToOrder(7, 44), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyMarketToOrder(7, 96), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyMarketToOrder(7, 75), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyMarketToOrder(7, 3), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyMarketToOrder(7, 56), []byte{keeper.OrderKeyTypeAsk})
			},
			marketID: 7,
			expSeen: []orderIterCBArgs{
				newOrderIterCBArgs(3, keeper.OrderKeyTypeBid),
				newOrderIterCBArgs(44, keeper.OrderKeyTypeAsk),
				newOrderIterCBArgs(56, keeper.OrderKeyTypeAsk),
				newOrderIterCBArgs(75, keeper.OrderKeyTypeBid),
				newOrderIterCBArgs(96, keeper.OrderKeyTypeBid),
			},
		},
		{
			name: "five entries, random: get one",
			setup: func() {
				store := s.getStore()
				store.Set(keeper.MakeIndexKeyMarketToOrder(7, 44), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyMarketToOrder(7, 96), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyMarketToOrder(7, 75), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyMarketToOrder(7, 3), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyMarketToOrder(7, 56), []byte{keeper.OrderKeyTypeAsk})
			},
			marketID: 7,
			cb:       stopAfter(1),
			expSeen:  []orderIterCBArgs{newOrderIterCBArgs(3, keeper.OrderKeyTypeAsk)},
		},
		{
			name: "five entries, random: get three",
			setup: func() {
				store := s.getStore()
				store.Set(keeper.MakeIndexKeyMarketToOrder(7, 44), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyMarketToOrder(7, 96), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyMarketToOrder(7, 75), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyMarketToOrder(7, 3), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyMarketToOrder(7, 56), []byte{keeper.OrderKeyTypeBid})
			},
			marketID: 7,
			cb:       stopAfter(3),
			expSeen: []orderIterCBArgs{
				newOrderIterCBArgs(3, keeper.OrderKeyTypeBid),
				newOrderIterCBArgs(44, keeper.OrderKeyTypeBid),
				newOrderIterCBArgs(56, keeper.OrderKeyTypeBid),
			},
		},
		{
			name: "five entries: two are bad",
			setup: func() {
				store := s.getStore()
				store.Set(keeper.MakeIndexKeyMarketToOrder(27, 1), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyMarketToOrder(27, 2), []byte{})
				store.Set(s.badKey(keeper.MakeIndexKeyMarketToOrder(27, 3)), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyMarketToOrder(27, 4), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyMarketToOrder(27, 5), []byte{keeper.OrderKeyTypeAsk})
			},
			marketID: 27,
			expSeen: []orderIterCBArgs{
				newOrderIterCBArgs(1, keeper.OrderKeyTypeAsk),
				newOrderIterCBArgs(4, keeper.OrderKeyTypeBid),
				newOrderIterCBArgs(5, keeper.OrderKeyTypeAsk),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			if tc.cb == nil {
				tc.cb = getAll
			}

			seen = nil
			testFunc := func() {
				s.k.IterateMarketOrders(s.ctx, tc.marketID, tc.cb)
			}
			s.Require().NotPanics(testFunc, "IterateMarketOrders(%d)", tc.marketID)
			assertEqualSlice(s, tc.expSeen, seen, orderIterCBArgs.orderIDString, "args provided to callback")
		})
	}
}

func (s *TestSuite) TestKeeper_IterateAddressOrders() {
	var seen []orderIterCBArgs
	getAll := func(orderID uint64, orderTypeByte byte) bool {
		seen = append(seen, newOrderIterCBArgs(orderID, orderTypeByte))
		return false
	}
	stopAfter := func(count int) func(orderID uint64, orderTypeByte byte) bool {
		return func(orderID uint64, orderTypeByte byte) bool {
			seen = append(seen, newOrderIterCBArgs(orderID, orderTypeByte))
			return len(seen) >= count
		}
	}

	tests := []struct {
		name    string
		setup   func()
		addr    sdk.AccAddress
		cb      func(orderID uint64, orderTypeByte byte) bool
		expSeen []orderIterCBArgs
	}{
		{
			name:    "empty state",
			addr:    s.addr1,
			expSeen: nil,
		},
		{
			name: "no orders for addr",
			setup: func() {
				store := s.getStore()
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr1, 1), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr1, 2), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr1, 3), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr3, 5), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr3, 6), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr3, 7), []byte{keeper.OrderKeyTypeAsk})
			},
			addr:    s.addr2,
			expSeen: nil,
		},
		{
			name: "one entry: ask",
			setup: func() {
				store := s.getStore()
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr1, 1), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr1, 2), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr1, 3), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr2, 4), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr3, 5), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr3, 6), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr3, 7), []byte{keeper.OrderKeyTypeAsk})
			},
			addr:    s.addr2,
			expSeen: []orderIterCBArgs{newOrderIterCBArgs(4, keeper.OrderKeyTypeAsk)},
		},
		{
			name: "one entry: bid",
			setup: func() {
				store := s.getStore()
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr1, 1), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr1, 2), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr1, 3), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr2, 4), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr3, 5), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr3, 6), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr3, 7), []byte{keeper.OrderKeyTypeAsk})
			},
			addr:    s.addr2,
			expSeen: []orderIterCBArgs{newOrderIterCBArgs(4, keeper.OrderKeyTypeBid)},
		},
		{
			name: "one entry no value",
			setup: func() {
				s.getStore().Set(keeper.MakeIndexKeyAddressToOrder(s.addr1, 4), []byte{})
			},
			addr:    s.addr1,
			expSeen: nil,
		},
		{
			name: "one entry bad key",
			setup: func() {
				s.getStore().Set(s.badKey(keeper.MakeIndexKeyAddressToOrder(s.addr1, 4)), []byte{keeper.OrderKeyTypeAsk})
			},
			addr:    s.addr1,
			expSeen: nil,
		},
		{
			name: "five entries, 1 through 5: get all",
			setup: func() {
				store := s.getStore()
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr4, 1), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr4, 2), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr4, 3), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr4, 4), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr4, 5), []byte{keeper.OrderKeyTypeBid})
			},
			addr: s.addr4,
			expSeen: []orderIterCBArgs{
				newOrderIterCBArgs(1, keeper.OrderKeyTypeBid),
				newOrderIterCBArgs(2, keeper.OrderKeyTypeBid),
				newOrderIterCBArgs(3, keeper.OrderKeyTypeBid),
				newOrderIterCBArgs(4, keeper.OrderKeyTypeAsk),
				newOrderIterCBArgs(5, keeper.OrderKeyTypeBid),
			},
		},
		{
			name: "five entries, 1 through 5: get one",
			setup: func() {
				store := s.getStore()
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr4, 1), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr4, 2), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr4, 3), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr4, 4), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr4, 5), []byte{keeper.OrderKeyTypeAsk})
			},
			addr:    s.addr4,
			cb:      stopAfter(1),
			expSeen: []orderIterCBArgs{newOrderIterCBArgs(1, keeper.OrderKeyTypeBid)},
		},
		{
			name: "five entries, 1 through 5: get three",
			setup: func() {
				store := s.getStore()
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr2, 1), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr2, 2), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr2, 3), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr2, 4), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr2, 5), []byte{keeper.OrderKeyTypeBid})
			},
			addr: s.addr2,
			cb:   stopAfter(3),
			expSeen: []orderIterCBArgs{
				newOrderIterCBArgs(1, keeper.OrderKeyTypeAsk),
				newOrderIterCBArgs(2, keeper.OrderKeyTypeBid),
				newOrderIterCBArgs(3, keeper.OrderKeyTypeAsk),
			},
		},
		{
			name: "five entries, random: get all",
			setup: func() {
				store := s.getStore()
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr5, 44), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr5, 96), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr5, 75), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr5, 3), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr5, 56), []byte{keeper.OrderKeyTypeAsk})
			},
			addr: s.addr5,
			expSeen: []orderIterCBArgs{
				newOrderIterCBArgs(3, keeper.OrderKeyTypeBid),
				newOrderIterCBArgs(44, keeper.OrderKeyTypeAsk),
				newOrderIterCBArgs(56, keeper.OrderKeyTypeAsk),
				newOrderIterCBArgs(75, keeper.OrderKeyTypeBid),
				newOrderIterCBArgs(96, keeper.OrderKeyTypeBid),
			},
		},
		{
			name: "five entries, random: get one",
			setup: func() {
				store := s.getStore()
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr3, 44), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr3, 96), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr3, 75), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr3, 3), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr3, 56), []byte{keeper.OrderKeyTypeAsk})
			},
			addr:    s.addr3,
			cb:      stopAfter(1),
			expSeen: []orderIterCBArgs{newOrderIterCBArgs(3, keeper.OrderKeyTypeAsk)},
		},
		{
			name: "five entries, random: get three",
			setup: func() {
				store := s.getStore()
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr3, 44), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr3, 96), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr3, 75), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr3, 3), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr3, 56), []byte{keeper.OrderKeyTypeBid})
			},
			addr: s.addr3,
			cb:   stopAfter(3),
			expSeen: []orderIterCBArgs{
				newOrderIterCBArgs(3, keeper.OrderKeyTypeBid),
				newOrderIterCBArgs(44, keeper.OrderKeyTypeBid),
				newOrderIterCBArgs(56, keeper.OrderKeyTypeBid),
			},
		},
		{
			name: "five entries: two are bad",
			setup: func() {
				store := s.getStore()
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr3, 1), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr3, 2), []byte{})
				store.Set(s.badKey(keeper.MakeIndexKeyAddressToOrder(s.addr3, 3)), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr3, 4), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyAddressToOrder(s.addr3, 5), []byte{keeper.OrderKeyTypeAsk})
			},
			addr: s.addr3,
			expSeen: []orderIterCBArgs{
				newOrderIterCBArgs(1, keeper.OrderKeyTypeAsk),
				newOrderIterCBArgs(4, keeper.OrderKeyTypeBid),
				newOrderIterCBArgs(5, keeper.OrderKeyTypeAsk),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			if tc.cb == nil {
				tc.cb = getAll
			}

			seen = nil
			testFunc := func() {
				s.k.IterateAddressOrders(s.ctx, tc.addr, tc.cb)
			}
			s.Require().NotPanics(testFunc, "IterateAddressOrders(%s)", s.getAddrName(tc.addr))
			assertEqualSlice(s, tc.expSeen, seen, orderIterCBArgs.orderIDString, "args provided to callback")
		})
	}
}

func (s *TestSuite) TestKeeper_IterateAssetOrders() {
	var seen []orderIterCBArgs
	getAll := func(orderID uint64, orderTypeByte byte) bool {
		seen = append(seen, newOrderIterCBArgs(orderID, orderTypeByte))
		return false
	}
	stopAfter := func(count int) func(orderID uint64, orderTypeByte byte) bool {
		return func(orderID uint64, orderTypeByte byte) bool {
			seen = append(seen, newOrderIterCBArgs(orderID, orderTypeByte))
			return len(seen) >= count
		}
	}

	tests := []struct {
		name       string
		setup      func()
		assetDenom string
		cb         func(orderID uint64, orderTypeByte byte) bool
		expSeen    []orderIterCBArgs
	}{
		{
			name:       "empty state",
			assetDenom: "apple",
			expSeen:    nil,
		},
		{
			name: "no orders for addr",
			setup: func() {
				store := s.getStore()
				store.Set(keeper.MakeIndexKeyAssetToOrder("apple", 1), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAssetToOrder("apple", 2), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAssetToOrder("apple", 3), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAssetToOrder("cactus", 5), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAssetToOrder("cactus", 6), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAssetToOrder("cactus", 7), []byte{keeper.OrderKeyTypeAsk})
			},
			assetDenom: "banana",
			expSeen:    nil,
		},
		{
			name: "one entry: ask",
			setup: func() {
				store := s.getStore()
				store.Set(keeper.MakeIndexKeyAssetToOrder("apple", 1), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAssetToOrder("apple", 2), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAssetToOrder("apple", 3), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAssetToOrder("banana", 4), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAssetToOrder("cactus", 5), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAssetToOrder("cactus", 6), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAssetToOrder("cactus", 7), []byte{keeper.OrderKeyTypeAsk})
			},
			assetDenom: "banana",
			expSeen:    []orderIterCBArgs{newOrderIterCBArgs(4, keeper.OrderKeyTypeAsk)},
		},
		{
			name: "one entry: bid",
			setup: func() {
				store := s.getStore()
				store.Set(keeper.MakeIndexKeyAssetToOrder("apple", 1), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAssetToOrder("apple", 2), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAssetToOrder("apple", 3), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAssetToOrder("banana", 4), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyAssetToOrder("cactus", 5), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAssetToOrder("cactus", 6), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAssetToOrder("cactus", 7), []byte{keeper.OrderKeyTypeAsk})
			},
			assetDenom: "banana",
			expSeen:    []orderIterCBArgs{newOrderIterCBArgs(4, keeper.OrderKeyTypeBid)},
		},
		{
			name: "one entry no value",
			setup: func() {
				s.getStore().Set(keeper.MakeIndexKeyAssetToOrder("banana", 4), []byte{})
			},
			assetDenom: "banana",
			expSeen:    nil,
		},
		{
			name: "one entry bad key",
			setup: func() {
				s.getStore().Set(s.badKey(keeper.MakeIndexKeyAssetToOrder("banana", 4)), []byte{keeper.OrderKeyTypeBid})
			},
			assetDenom: "banana",
			expSeen:    nil,
		},
		{
			name: "five entries, 1 through 5: get all",
			setup: func() {
				store := s.getStore()
				store.Set(keeper.MakeIndexKeyAssetToOrder("acorn", 1), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyAssetToOrder("acorn", 2), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyAssetToOrder("acorn", 3), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyAssetToOrder("acorn", 4), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAssetToOrder("acorn", 5), []byte{keeper.OrderKeyTypeBid})
			},
			assetDenom: "acorn",
			expSeen: []orderIterCBArgs{
				newOrderIterCBArgs(1, keeper.OrderKeyTypeBid),
				newOrderIterCBArgs(2, keeper.OrderKeyTypeBid),
				newOrderIterCBArgs(3, keeper.OrderKeyTypeBid),
				newOrderIterCBArgs(4, keeper.OrderKeyTypeAsk),
				newOrderIterCBArgs(5, keeper.OrderKeyTypeBid),
			},
		},
		{
			name: "five entries, 1 through 5: get one",
			setup: func() {
				store := s.getStore()
				store.Set(keeper.MakeIndexKeyAssetToOrder("acorn", 1), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyAssetToOrder("acorn", 2), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAssetToOrder("acorn", 3), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAssetToOrder("acorn", 4), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyAssetToOrder("acorn", 5), []byte{keeper.OrderKeyTypeAsk})
			},
			assetDenom: "acorn",
			cb:         stopAfter(1),
			expSeen:    []orderIterCBArgs{newOrderIterCBArgs(1, keeper.OrderKeyTypeBid)},
		},
		{
			name: "five entries, 1 through 5: get three",
			setup: func() {
				store := s.getStore()
				store.Set(keeper.MakeIndexKeyAssetToOrder("acorn", 1), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAssetToOrder("acorn", 2), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyAssetToOrder("acorn", 3), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAssetToOrder("acorn", 4), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyAssetToOrder("acorn", 5), []byte{keeper.OrderKeyTypeBid})
			},
			assetDenom: "acorn",
			cb:         stopAfter(3),
			expSeen: []orderIterCBArgs{
				newOrderIterCBArgs(1, keeper.OrderKeyTypeAsk),
				newOrderIterCBArgs(2, keeper.OrderKeyTypeBid),
				newOrderIterCBArgs(3, keeper.OrderKeyTypeAsk),
			},
		},
		{
			name: "five entries, random: get all",
			setup: func() {
				store := s.getStore()
				store.Set(keeper.MakeIndexKeyAssetToOrder("raspberry", 44), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAssetToOrder("raspberry", 96), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyAssetToOrder("raspberry", 75), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyAssetToOrder("raspberry", 3), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyAssetToOrder("raspberry", 56), []byte{keeper.OrderKeyTypeAsk})
			},
			assetDenom: "raspberry",
			expSeen: []orderIterCBArgs{
				newOrderIterCBArgs(3, keeper.OrderKeyTypeBid),
				newOrderIterCBArgs(44, keeper.OrderKeyTypeAsk),
				newOrderIterCBArgs(56, keeper.OrderKeyTypeAsk),
				newOrderIterCBArgs(75, keeper.OrderKeyTypeBid),
				newOrderIterCBArgs(96, keeper.OrderKeyTypeBid),
			},
		},
		{
			name: "five entries, random: get one",
			setup: func() {
				store := s.getStore()
				store.Set(keeper.MakeIndexKeyAssetToOrder("raspberry", 44), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyAssetToOrder("raspberry", 96), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAssetToOrder("raspberry", 75), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyAssetToOrder("raspberry", 3), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAssetToOrder("raspberry", 56), []byte{keeper.OrderKeyTypeAsk})
			},
			assetDenom: "raspberry",
			cb:         stopAfter(1),
			expSeen:    []orderIterCBArgs{newOrderIterCBArgs(3, keeper.OrderKeyTypeAsk)},
		},
		{
			name: "five entries, random: get three",
			setup: func() {
				store := s.getStore()
				store.Set(keeper.MakeIndexKeyAssetToOrder("huckleberry", 44), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyAssetToOrder("huckleberry", 96), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAssetToOrder("huckleberry", 75), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAssetToOrder("huckleberry", 3), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyAssetToOrder("huckleberry", 56), []byte{keeper.OrderKeyTypeBid})
			},
			assetDenom: "huckleberry",
			cb:         stopAfter(3),
			expSeen: []orderIterCBArgs{
				newOrderIterCBArgs(3, keeper.OrderKeyTypeBid),
				newOrderIterCBArgs(44, keeper.OrderKeyTypeBid),
				newOrderIterCBArgs(56, keeper.OrderKeyTypeBid),
			},
		},
		{
			name: "five entries: two are bad",
			setup: func() {
				store := s.getStore()
				store.Set(keeper.MakeIndexKeyAssetToOrder("huckleberry", 1), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAssetToOrder("huckleberry", 2), []byte{})
				store.Set(s.badKey(keeper.MakeIndexKeyAssetToOrder("huckleberry", 3)), []byte{keeper.OrderKeyTypeAsk})
				store.Set(keeper.MakeIndexKeyAssetToOrder("huckleberry", 4), []byte{keeper.OrderKeyTypeBid})
				store.Set(keeper.MakeIndexKeyAssetToOrder("huckleberry", 5), []byte{keeper.OrderKeyTypeAsk})
			},
			assetDenom: "huckleberry",
			expSeen: []orderIterCBArgs{
				newOrderIterCBArgs(1, keeper.OrderKeyTypeAsk),
				newOrderIterCBArgs(4, keeper.OrderKeyTypeBid),
				newOrderIterCBArgs(5, keeper.OrderKeyTypeAsk),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			if tc.cb == nil {
				tc.cb = getAll
			}

			seen = nil
			testFunc := func() {
				s.k.IterateAssetOrders(s.ctx, tc.assetDenom, tc.cb)
			}
			s.Require().NotPanics(testFunc, "IterateAssetOrders(%q)", tc.assetDenom)
			s.Assert().Equal(tc.expSeen, seen, "args provided to callback")
			assertEqualSlice(s, tc.expSeen, seen, orderIterCBArgs.orderIDString, "args provided to callback")
		})
	}
}

// TODO[1789]: func (s *TestSuite) TestKeeper_CancelAllOrdersForMarket()
