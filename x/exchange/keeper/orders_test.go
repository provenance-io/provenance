package keeper_test

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/exchange"
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

// TODO[1658]: func (s *TestSuite) TestKeeper_GetOrderByExternalID()

// TODO[1658]: func (s *TestSuite) TestKeeper_CreateAskOrder()

// TODO[1658]: func (s *TestSuite) TestKeeper_CreateBidOrder()

// TODO[1658]: func (s *TestSuite) TestKeeper_CancelOrder()

// TODO[1658]: func (s *TestSuite) TestKeeper_SetOrderExternalID()

// TODO[1658]: func (s *TestSuite) TestKeeper_IterateOrders()

// TODO[1658]: func (s *TestSuite) TestKeeper_IterateMarketOrders()

// TODO[1658]: func (s *TestSuite) TestKeeper_IterateAddressOrders()

// TODO[1658]: func (s *TestSuite) TestKeeper_IterateAssetOrders()
