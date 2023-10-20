package keeper_test

import (
	"bytes"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/exchange"
	"github.com/provenance-io/provenance/x/exchange/keeper"
)

// assertEqualOrderID asserts that two uint64 values are equal, and if not, includes their decimal form in the log.
// This is nice because .Equal failures output uints in hex, which can make it difficult to identify what's going on.
func (s *TestSuite) assertEqualOrderID(expected, actual uint64, msgAndArgs ...interface{}) bool {
	if s.Assert().Equal(expected, actual, msgAndArgs...) {
		return true
	}
	s.T().Logf("Expected order id: %d", expected)
	s.T().Logf("  Actual order id: %d", actual)
	return false
}

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
			expErr: "invalid market id: must not be zero",
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
			expEvents := sdk.Events{}
			if len(tc.expErr) == 0 {
				expOrder = exchange.NewOrder(tc.expOrderID).WithAsk(&tc.askOrder)
				event, err := sdk.TypedEventToEvent(exchange.NewEventOrderCreated(expOrder))
				s.Require().NoError(err, "TypedEventToEvent")
				expEvents = append(expEvents, event)
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
			s.Assert().Equal(tc.expOrderID, lastOrderID, "last order id")
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
			expErr: "invalid market id: must not be zero",
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
			expEvents := sdk.Events{}
			if len(tc.expErr) == 0 {
				expOrder = exchange.NewOrder(tc.expOrderID).WithBid(&tc.bidOrder)
				event, err := sdk.TypedEventToEvent(exchange.NewEventOrderCreated(expOrder))
				s.Require().NoError(err, "TypedEventToEvent")
				expEvents = append(expEvents, event)
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
			s.Assert().Equal(tc.expOrderID, lastOrderID, "last order id")
		})
	}
}

// TODO[1658]: func (s *TestSuite) TestKeeper_CancelOrder()

// TODO[1658]: func (s *TestSuite) TestKeeper_SetOrderExternalID()

// TODO[1658]: func (s *TestSuite) TestKeeper_IterateOrders()

// TODO[1658]: func (s *TestSuite) TestKeeper_IterateMarketOrders()

// TODO[1658]: func (s *TestSuite) TestKeeper_IterateAddressOrders()

// TODO[1658]: func (s *TestSuite) TestKeeper_IterateAssetOrders()
