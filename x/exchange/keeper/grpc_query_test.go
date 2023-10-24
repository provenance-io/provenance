package keeper_test

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/exchange"
	"github.com/provenance-io/provenance/x/exchange/keeper"
)

const invalidArgErr = "rpc error: code = InvalidArgument"

// querySetupFunc is a function that will set up a test case.
type querySetupFunc func(ctx sdk.Context)

// queryRunner is a function that will call a query endpoint, returning its response and error.
type queryRunner func(goCtx context.Context) (interface{}, error)

// doQueryTest runs the provided setup and runner, requiring the runner to not panic and asserting its error's content
// to contain the provided strings (or be nil if none are provided).
// A cache context is used for this, so the setup function won't affect other test cases.
// The query's response is returned.
func (s *TestSuite) doQueryTest(setup querySetupFunc, runner queryRunner, expInErr []string, msg string, args ...interface{}) interface{} {
	s.T().Helper()
	ctx, _ := s.ctx.CacheContext()
	if setup != nil {
		setup(ctx)
	}
	goCtx := sdk.WrapSDKContext(ctx)
	var rv interface{}
	var err error
	testFunc := func() {
		rv, err = runner(goCtx)
	}
	s.Require().NotPanicsf(testFunc, msg, args...)
	s.assertErrorContentsf(err, expInErr, msg+" error", args...)
	return rv
}

func (s *TestSuite) TestQueryServer_OrderFeeCalc() {
	queryName := "OrderFeeCalc"
	runner := func(req *exchange.QueryOrderFeeCalcRequest) queryRunner {
		return func(goCtx context.Context) (interface{}, error) {
			return keeper.NewQueryServer(s.k).OrderFeeCalc(goCtx, req)
		}
	}

	tests := []struct {
		name     string
		setup    querySetupFunc
		req      *exchange.QueryOrderFeeCalcRequest
		expResp  *exchange.QueryOrderFeeCalcResponse
		expInErr []string
	}{
		// Bad request tests.
		{
			name:     "nil req",
			req:      nil,
			expInErr: []string{invalidArgErr, "empty request"},
		},
		{
			name:     "empty req",
			req:      &exchange.QueryOrderFeeCalcRequest{},
			expInErr: []string{invalidArgErr, "empty request"},
		},
		{
			name: "both order types",
			req: &exchange.QueryOrderFeeCalcRequest{
				AskOrder: &exchange.AskOrder{MarketId: 1},
				BidOrder: &exchange.BidOrder{MarketId: 1},
			},
			expInErr: []string{invalidArgErr, "invalid request"},
		},

		// AskOrder tests.
		{
			name: "ask: unknown market",
			req: &exchange.QueryOrderFeeCalcRequest{AskOrder: &exchange.AskOrder{
				Assets: s.coin("1apple"), Price: s.coin("2plum"), MarketId: 99,
			}},
			expInErr: []string{invalidArgErr, "market 99 does not exist"},
		},
		{
			name: "ask: no fees",
			setup: func(ctx sdk.Context) {
				s.requireCreateMarket(exchange.Market{MarketId: 1})
			},
			req: &exchange.QueryOrderFeeCalcRequest{AskOrder: &exchange.AskOrder{
				Assets: s.coin("1apple"), Price: s.coin("2000plum"), MarketId: 1,
			}},
			expResp: &exchange.QueryOrderFeeCalcResponse{},
		},
		{
			name: "ask: only creation: one option",
			setup: func(ctx sdk.Context) {
				s.requireCreateMarket(exchange.Market{
					MarketId:         1,
					FeeCreateAskFlat: s.coins("3fig"),
				})
			},
			req: &exchange.QueryOrderFeeCalcRequest{AskOrder: &exchange.AskOrder{
				Assets: s.coin("1apple"), Price: s.coin("2000plum"), MarketId: 1,
			}},
			expResp: &exchange.QueryOrderFeeCalcResponse{
				CreationFeeOptions: s.coins("3fig"),
			},
		},
		{
			name: "ask: only creation: three options",
			setup: func(ctx sdk.Context) {
				s.requireCreateMarket(exchange.Market{
					MarketId:         8,
					FeeCreateAskFlat: s.coins("3fig,52grape,1honeydew"),
				})
			},
			req: &exchange.QueryOrderFeeCalcRequest{AskOrder: &exchange.AskOrder{
				Assets: s.coin("1apple"), Price: s.coin("2000plum"), MarketId: 8,
			}},
			expResp: &exchange.QueryOrderFeeCalcResponse{
				CreationFeeOptions: s.coins("3fig,52grape,1honeydew"),
			},
		},
		{
			name: "ask: only settlement flat: one option",
			setup: func(ctx sdk.Context) {
				s.requireCreateMarket(exchange.Market{
					MarketId:                8,
					FeeSellerSettlementFlat: s.coins("8grape"),
				})
			},
			req: &exchange.QueryOrderFeeCalcRequest{AskOrder: &exchange.AskOrder{
				Assets: s.coin("1apple"), Price: s.coin("2000plum"), MarketId: 8,
			}},
			expResp: &exchange.QueryOrderFeeCalcResponse{
				SettlementFlatFeeOptions: s.coins("8grape"),
			},
		},
		{
			name: "ask: only settlement flat: three option",
			setup: func(ctx sdk.Context) {
				s.requireCreateMarket(exchange.Market{
					MarketId:                8,
					FeeSellerSettlementFlat: s.coins("23fig,6grape,15pineapple"),
				})
			},
			req: &exchange.QueryOrderFeeCalcRequest{AskOrder: &exchange.AskOrder{
				Assets: s.coin("1apple"), Price: s.coin("2000plum"), MarketId: 8,
			}},
			expResp: &exchange.QueryOrderFeeCalcResponse{
				SettlementFlatFeeOptions: s.coins("23fig,6grape,15pineapple"),
			},
		},
		{
			name: "ask: price denom without ratio",
			setup: func(ctx sdk.Context) {
				s.requireCreateMarket(exchange.Market{
					MarketId:                  4,
					FeeSellerSettlementRatios: s.ratios("500plum:3plum"),
				})
			},
			req: &exchange.QueryOrderFeeCalcRequest{AskOrder: &exchange.AskOrder{
				Assets: s.coin("1apple"), Price: s.coin("2000peach"), MarketId: 4,
			}},
			expInErr: []string{invalidArgErr, "failed to calculate seller ratio fee option",
				"no seller settlement fee ratio found for denom \"peach\""},
		},
		{
			name: "ask: only settlement ratio",
			setup: func(ctx sdk.Context) {
				s.requireCreateMarket(exchange.Market{
					MarketId:                  8,
					FeeSellerSettlementRatios: s.ratios("500plum:3plum"),
				})
			},
			req: &exchange.QueryOrderFeeCalcRequest{AskOrder: &exchange.AskOrder{
				Assets: s.coin("1apple"), Price: s.coin("2000plum"), MarketId: 8,
			}},
			expResp: &exchange.QueryOrderFeeCalcResponse{
				SettlementRatioFeeOptions: s.coins("12plum"),
			},
		},
		{
			name: "ask: both settlement",
			setup: func(ctx sdk.Context) {
				s.requireCreateMarket(exchange.Market{
					MarketId:                  8,
					FeeSellerSettlementFlat:   s.coins("23fig,6grape,15pineapple"),
					FeeSellerSettlementRatios: s.ratios("500plum:3plum"),
				})
			},
			req: &exchange.QueryOrderFeeCalcRequest{AskOrder: &exchange.AskOrder{
				Assets: s.coin("1apple"), Price: s.coin("2000plum"), MarketId: 8,
			}},
			expResp: &exchange.QueryOrderFeeCalcResponse{
				SettlementFlatFeeOptions:  s.coins("23fig,6grape,15pineapple"),
				SettlementRatioFeeOptions: s.coins("12plum"),
			},
		},
		{
			name: "ask: all fees",
			setup: func(ctx sdk.Context) {
				s.requireCreateMarket(exchange.Market{
					MarketId:                  1,
					FeeCreateAskFlat:          s.coins("3fig,52grape,1honeydew"),
					FeeSellerSettlementFlat:   s.coins("23fig,6grape,15pineapple"),
					FeeSellerSettlementRatios: s.ratios("500plum:3plum"),
				})
			},
			req: &exchange.QueryOrderFeeCalcRequest{AskOrder: &exchange.AskOrder{
				Assets: s.coin("1apple"), Price: s.coin("2000plum"), MarketId: 1,
			}},
			expResp: &exchange.QueryOrderFeeCalcResponse{
				CreationFeeOptions:        s.coins("3fig,52grape,1honeydew"),
				SettlementFlatFeeOptions:  s.coins("23fig,6grape,15pineapple"),
				SettlementRatioFeeOptions: s.coins("12plum"),
			},
		},

		// BidOrder tests.
		{
			name: "bid: unknown market",
			req: &exchange.QueryOrderFeeCalcRequest{BidOrder: &exchange.BidOrder{
				Assets: s.coin("1apple"), Price: s.coin("2000plum"), MarketId: 33,
			}},
			expInErr: []string{invalidArgErr, "market 33 does not exist"},
		},
		{
			name: "bid: no fees",
			setup: func(ctx sdk.Context) {
				s.requireCreateMarket(exchange.Market{MarketId: 1})
			},
			req: &exchange.QueryOrderFeeCalcRequest{BidOrder: &exchange.BidOrder{
				Assets: s.coin("1apple"), Price: s.coin("2000plum"), MarketId: 1,
			}},
			expResp: &exchange.QueryOrderFeeCalcResponse{},
		},
		{
			name: "bid: only creation: one option",
			setup: func(ctx sdk.Context) {
				s.requireCreateMarket(exchange.Market{
					MarketId:         33,
					FeeCreateBidFlat: s.coins("7honeydew"),
				})
			},
			req: &exchange.QueryOrderFeeCalcRequest{BidOrder: &exchange.BidOrder{
				Assets: s.coin("1apple"), Price: s.coin("2000plum"), MarketId: 33,
			}},
			expResp: &exchange.QueryOrderFeeCalcResponse{
				CreationFeeOptions: s.coins("7honeydew"),
			},
		},
		{
			name: "bid: only creation: three options",
			setup: func(ctx sdk.Context) {
				s.requireCreateMarket(exchange.Market{
					MarketId:         33,
					FeeCreateBidFlat: s.coins("57fig,6honeydew,223jackfruit"),
				})
			},
			req: &exchange.QueryOrderFeeCalcRequest{BidOrder: &exchange.BidOrder{
				Assets: s.coin("1apple"), Price: s.coin("2000plum"), MarketId: 33,
			}},
			expResp: &exchange.QueryOrderFeeCalcResponse{
				CreationFeeOptions: s.coins("57fig,6honeydew,223jackfruit"),
			},
		},
		{
			name: "bid: only settlement flat: one option",
			setup: func(ctx sdk.Context) {
				s.requireCreateMarket(exchange.Market{
					MarketId:               3,
					FeeBuyerSettlementFlat: s.coins("12pineapple"),
				})
			},
			req: &exchange.QueryOrderFeeCalcRequest{BidOrder: &exchange.BidOrder{
				Assets: s.coin("1apple"), Price: s.coin("2000plum"), MarketId: 3,
			}},
			expResp: &exchange.QueryOrderFeeCalcResponse{
				SettlementFlatFeeOptions: s.coins("12pineapple"),
			},
		},
		{
			name: "bid: only settlement flat: three options",
			setup: func(ctx sdk.Context) {
				s.requireCreateMarket(exchange.Market{
					MarketId:               3,
					FeeBuyerSettlementFlat: s.coins("7peach,12pineapple,66plum"),
				})
			},
			req: &exchange.QueryOrderFeeCalcRequest{BidOrder: &exchange.BidOrder{
				Assets: s.coin("1apple"), Price: s.coin("2000plum"), MarketId: 3,
			}},
			expResp: &exchange.QueryOrderFeeCalcResponse{
				SettlementFlatFeeOptions: s.coins("7peach,12pineapple,66plum"),
			},
		},
		{
			name: "bid: price denom without ratio",
			setup: func(ctx sdk.Context) {
				s.requireCreateMarket(exchange.Market{
					MarketId:                 1,
					FeeBuyerSettlementRatios: s.ratios("1000peach:3fig"),
				})
			},
			req: &exchange.QueryOrderFeeCalcRequest{BidOrder: &exchange.BidOrder{
				Assets: s.coin("1apple"), Price: s.coin("2000plum"), MarketId: 1,
			}},
			expInErr: []string{invalidArgErr, "failed to calculate buyer ratio fee options",
				"no buyer settlement fee ratios found for denom \"plum\""},
		},
		{
			name: "bid: no applicable ratios for price amount",
			setup: func(ctx sdk.Context) {
				s.requireCreateMarket(exchange.Market{
					MarketId:                 7,
					FeeBuyerSettlementRatios: s.ratios("1000plum:3fig,750plum:66grape,500plum:1honeydew"),
				})
			},
			req: &exchange.QueryOrderFeeCalcRequest{BidOrder: &exchange.BidOrder{
				Assets: s.coin("1apple"), Price: s.coin("5737plum"), MarketId: 7,
			}},
			expInErr: []string{invalidArgErr, "failed to calculate buyer ratio fee options",
				"cannot apply ratio 1000plum:3fig to price 5737plum",
				"cannot apply ratio 750plum:66grape to price 5737plum",
				"cannot apply ratio 500plum:1honeydew to price 5737plum",
				"no applicable buyer settlement fee ratios found for price 5737plum",
			},
		},
		{
			name: "bid: only settlement ratio: one option",
			setup: func(ctx sdk.Context) {
				s.requireCreateMarket(exchange.Market{
					MarketId:                 1,
					FeeBuyerSettlementRatios: s.ratios("1000plum:3fig"),
				})
			},
			req: &exchange.QueryOrderFeeCalcRequest{BidOrder: &exchange.BidOrder{
				Assets: s.coin("1apple"), Price: s.coin("2000plum"), MarketId: 1,
			}},
			expResp: &exchange.QueryOrderFeeCalcResponse{
				SettlementRatioFeeOptions: s.coins("6fig"),
			},
		},
		{
			name: "bid: only settlement ratio: three options",
			setup: func(ctx sdk.Context) {
				s.requireCreateMarket(exchange.Market{
					MarketId: 1,
					FeeBuyerSettlementRatios: []exchange.FeeRatio{
						s.ratio("1000plum:3fig"),
						s.ratio("751plum:33grape"),   // cannot be applied to given price.
						s.ratio("1apple:10honeydew"), // cannot be applied to given price.
						s.ratio("2000plum:5peach"),
						s.ratio("500plum:1plum"),
					},
				})
			},
			req: &exchange.QueryOrderFeeCalcRequest{BidOrder: &exchange.BidOrder{
				Assets: s.coin("1apple"), Price: s.coin("2000plum"), MarketId: 1,
			}},
			expResp: &exchange.QueryOrderFeeCalcResponse{
				SettlementRatioFeeOptions: s.coins("6fig,5peach,4plum"),
			},
		},
		{
			name: "bid: both settlement",
			setup: func(ctx sdk.Context) {
				s.requireCreateMarket(exchange.Market{
					MarketId:                 2,
					FeeBuyerSettlementFlat:   s.coins("12fig,15grape"),
					FeeBuyerSettlementRatios: s.ratios("1000plum:3fig,1000plum:4grape"),
				})
			},
			req: &exchange.QueryOrderFeeCalcRequest{BidOrder: &exchange.BidOrder{
				Assets: s.coin("1apple"), Price: s.coin("2000plum"), MarketId: 2,
			}},
			expResp: &exchange.QueryOrderFeeCalcResponse{
				SettlementFlatFeeOptions:  s.coins("12fig,15grape"),
				SettlementRatioFeeOptions: s.coins("6fig,8grape"),
			},
		},
		{
			name: "bid: all fees",
			setup: func(ctx sdk.Context) {
				s.requireCreateMarket(exchange.Market{
					MarketId:                 3,
					FeeCreateBidFlat:         s.coins("77fig,88grape"),
					FeeBuyerSettlementFlat:   s.coins("12fig,15grape"),
					FeeBuyerSettlementRatios: s.ratios("1000plum:3fig,1000plum:4grape"),
				})
			},
			req: &exchange.QueryOrderFeeCalcRequest{BidOrder: &exchange.BidOrder{
				Assets: s.coin("1apple"), Price: s.coin("2000plum"), MarketId: 3,
			}},
			expResp: &exchange.QueryOrderFeeCalcResponse{
				CreationFeeOptions:        s.coins("77fig,88grape"),
				SettlementFlatFeeOptions:  s.coins("12fig,15grape"),
				SettlementRatioFeeOptions: s.coins("6fig,8grape"),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			respRaw := s.doQueryTest(tc.setup, runner(tc.req), tc.expInErr, queryName)
			if s.Assert().Equal(tc.expResp, respRaw, queryName+" result") {
				return
			}
			resp, ok := respRaw.(*exchange.QueryOrderFeeCalcResponse)
			s.Require().True(ok, queryName+" response is of type %T and could not be cast to %T", respRaw, tc.expResp)
			if tc.expResp == nil || resp == nil {
				return
			}
			s.Assert().Equal(s.coinsString(tc.expResp.CreationFeeOptions), s.coinsString(resp.CreationFeeOptions),
				queryName+" CreationFeeOptions (as strings)")
			s.Assert().Equal(s.coinsString(tc.expResp.SettlementFlatFeeOptions), s.coinsString(resp.SettlementFlatFeeOptions),
				queryName+" SettlementFlatFeeOptions (as strings)")
			s.Assert().Equal(s.coinsString(tc.expResp.SettlementRatioFeeOptions), s.coinsString(resp.SettlementRatioFeeOptions),
				queryName+" SettlementRatioFeeOptions (as strings)")
		})
	}
}

func (s *TestSuite) TestQueryServer_GetOrder() {
	queryName := "GetOrder"
	runner := func(req *exchange.QueryGetOrderRequest) queryRunner {
		return func(goCtx context.Context) (interface{}, error) {
			return keeper.NewQueryServer(s.k).GetOrder(goCtx, req)
		}
	}

	tests := []struct {
		name     string
		setup    querySetupFunc
		req      *exchange.QueryGetOrderRequest
		expResp  *exchange.QueryGetOrderResponse
		expInErr []string
	}{
		{
			name:     "nil req",
			req:      nil,
			expInErr: []string{invalidArgErr, "empty request"},
		},
		{
			name:     "order 0",
			req:      &exchange.QueryGetOrderRequest{OrderId: 0},
			expInErr: []string{invalidArgErr, "empty request"},
		},
		{
			name: "error getting order",
			setup: func(ctx sdk.Context) {
				key, value, err := s.k.GetOrderStoreKeyValue(*exchange.NewOrder(4).WithAsk(&exchange.AskOrder{
					MarketId: 1,
					Seller:   s.addr1.String(),
					Assets:   s.coin("55apple"),
					Price:    s.coin("99prune"),
				}))
				s.Require().NoError(err, "GetOrderStoreKeyValue 4")
				value[0] = 9
				s.k.GetStore(ctx).Set(key, value)
			},
			req:      &exchange.QueryGetOrderRequest{OrderId: 4},
			expInErr: []string{invalidArgErr, "failed to read order 4: unknown type byte 0x9"},
		},
		{
			name:     "order not found",
			req:      &exchange.QueryGetOrderRequest{OrderId: 4},
			expInErr: []string{invalidArgErr, "order 4 not found"},
		},
		{
			name: "order 1: ask",
			setup: func(ctx sdk.Context) {
				store := s.k.GetStore(ctx)
				s.requireSetOrderInStore(store, exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
					MarketId:                1,
					Seller:                  s.addr1.String(),
					Assets:                  s.coin("20apple"),
					Price:                   s.coin("3pineapple"),
					SellerSettlementFlatFee: s.coinP("15fig"),
					AllowPartial:            true,
					ExternalId:              "ask-order-1-id",
				}))
			},
			req: &exchange.QueryGetOrderRequest{OrderId: 1},
			expResp: &exchange.QueryGetOrderResponse{Order: exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
				MarketId:                1,
				Seller:                  s.addr1.String(),
				Assets:                  s.coin("20apple"),
				Price:                   s.coin("3pineapple"),
				SellerSettlementFlatFee: s.coinP("15fig"),
				AllowPartial:            true,
				ExternalId:              "ask-order-1-id",
			})},
		},
		{
			name: "order 1: bid",
			setup: func(ctx sdk.Context) {
				store := s.k.GetStore(ctx)
				s.requireSetOrderInStore(store, exchange.NewOrder(1).WithBid(&exchange.BidOrder{
					MarketId:            1,
					Buyer:               s.addr1.String(),
					Assets:              s.coin("20apple"),
					Price:               s.coin("3pineapple"),
					BuyerSettlementFees: s.coins("15fig,10grape"),
					AllowPartial:        true,
					ExternalId:          "ask-order-1-id",
				}))
			},
			req: &exchange.QueryGetOrderRequest{OrderId: 1},
			expResp: &exchange.QueryGetOrderResponse{Order: exchange.NewOrder(1).WithBid(&exchange.BidOrder{
				MarketId:            1,
				Buyer:               s.addr1.String(),
				Assets:              s.coin("20apple"),
				Price:               s.coin("3pineapple"),
				BuyerSettlementFees: s.coins("15fig,10grape"),
				AllowPartial:        true,
				ExternalId:          "ask-order-1-id",
			})},
		},
		{
			name: "order 5555",
			setup: func(ctx sdk.Context) {
				store := s.k.GetStore(ctx)
				s.requireSetOrderInStore(store, exchange.NewOrder(5554).WithBid(&exchange.BidOrder{
					MarketId: 1, Buyer: s.addr1.String(),
					Assets: s.coin("20apple"), Price: s.coin("3pineapple"),
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(5555).WithAsk(&exchange.AskOrder{
					MarketId: 1, Seller: s.addr2.String(),
					Assets: s.coin("77acorn"), Price: s.coin("453prune"),
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(5556).WithBid(&exchange.BidOrder{
					MarketId: 1, Buyer: s.addr3.String(),
					Assets: s.coin("55acai"), Price: s.coin("77peach"),
				}))
			},
			req: &exchange.QueryGetOrderRequest{OrderId: 5555},
			expResp: &exchange.QueryGetOrderResponse{Order: exchange.NewOrder(5555).WithAsk(&exchange.AskOrder{
				MarketId: 1, Seller: s.addr2.String(),
				Assets: s.coin("77acorn"), Price: s.coin("453prune"),
			})},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			respRaw := s.doQueryTest(tc.setup, runner(tc.req), tc.expInErr, queryName)
			s.Assert().Equal(tc.expResp, respRaw, queryName+" result")
		})
	}
}

func (s *TestSuite) TestQueryServer_GetOrderByExternalID() {
	queryName := "GetOrderByExternalID"
	runner := func(req *exchange.QueryGetOrderByExternalIDRequest) queryRunner {
		return func(goCtx context.Context) (interface{}, error) {
			return keeper.NewQueryServer(s.k).GetOrderByExternalID(goCtx, req)
		}
	}

	tests := []struct {
		name     string
		setup    querySetupFunc
		req      *exchange.QueryGetOrderByExternalIDRequest
		expResp  *exchange.QueryGetOrderByExternalIDResponse
		expInErr []string
	}{
		{
			name:     "nil request",
			req:      nil,
			expInErr: []string{invalidArgErr, "empty request"},
		},
		{
			name:     "market 0",
			req:      &exchange.QueryGetOrderByExternalIDRequest{MarketId: 0, ExternalId: "something"},
			expInErr: []string{invalidArgErr, "invalid request"},
		},
		{
			name:     "no external id",
			req:      &exchange.QueryGetOrderByExternalIDRequest{MarketId: 1, ExternalId: ""},
			expInErr: []string{invalidArgErr, "invalid request"},
		},
		{
			name: "error getting order",
			setup: func(ctx sdk.Context) {
				order5 := exchange.NewOrder(5).WithBid(&exchange.BidOrder{
					MarketId:            1,
					Buyer:               s.addr3.String(),
					Assets:              s.coin("1apple"),
					Price:               s.coin("1plum"),
					BuyerSettlementFees: nil,
					AllowPartial:        false,
					ExternalId:          "babbaderr",
				})
				store := s.k.GetStore(ctx)
				// Save it normally to get the indexes with it, then overwite the value with a bad one.
				s.requireSetOrderInStore(store, order5)
				key5, value5, err := s.k.GetOrderStoreKeyValue(*order5)
				s.Require().NoError(err, "GetOrderStoreKeyValue 5")
				value5[0] = 9
				store.Set(key5, value5)
			},
			req:      &exchange.QueryGetOrderByExternalIDRequest{MarketId: 1, ExternalId: "babbaderr"},
			expInErr: []string{invalidArgErr, "failed to read order 5: unknown type byte 0x9"},
		},
		{
			name: "no such order",
			setup: func(ctx sdk.Context) {
				store := s.k.GetStore(ctx)
				s.requireSetOrderInStore(store, exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
					MarketId: 1, Seller: s.addr2.String(),
					Assets: s.coin("1apple"), Price: s.coin("1plum"),
					ExternalId: "nosuchorder",
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(3).WithAsk(&exchange.AskOrder{
					MarketId: 3, Seller: s.addr2.String(),
					Assets: s.coin("1apple"), Price: s.coin("1plum"),
					ExternalId: "nosuchorder",
				}))
			},
			req:      &exchange.QueryGetOrderByExternalIDRequest{MarketId: 2, ExternalId: "nosuchorder"},
			expInErr: []string{invalidArgErr, "order not found in market 2 with external id \"nosuchorder\""},
		},
		{
			name: "only one order with the id: ask",
			setup: func(ctx sdk.Context) {
				store := s.k.GetStore(ctx)
				s.requireSetOrderInStore(store, exchange.NewOrder(77).WithAsk(&exchange.AskOrder{
					MarketId: 3, Seller: s.addr2.String(), Assets: s.coin("1apple"), Price: s.coin("1plum"),
					ExternalId: "myspecialid",
				}))
			},
			req: &exchange.QueryGetOrderByExternalIDRequest{MarketId: 3, ExternalId: "myspecialid"},
			expResp: &exchange.QueryGetOrderByExternalIDResponse{Order: exchange.NewOrder(77).WithAsk(&exchange.AskOrder{
				MarketId: 3, Seller: s.addr2.String(), Assets: s.coin("1apple"), Price: s.coin("1plum"),
				ExternalId: "myspecialid",
			})},
		},
		{
			name: "only one order with the id: bid",
			setup: func(ctx sdk.Context) {
				store := s.k.GetStore(ctx)
				s.requireSetOrderInStore(store, exchange.NewOrder(54).WithBid(&exchange.BidOrder{
					MarketId: 999, Buyer: s.addr2.String(), Assets: s.coin("1apple"), Price: s.coin("1plum"),
					ExternalId: "mycoolid",
				}))
			},
			req: &exchange.QueryGetOrderByExternalIDRequest{MarketId: 999, ExternalId: "mycoolid"},
			expResp: &exchange.QueryGetOrderByExternalIDResponse{Order: exchange.NewOrder(54).WithBid(&exchange.BidOrder{
				MarketId: 999, Buyer: s.addr2.String(), Assets: s.coin("1apple"), Price: s.coin("1plum"),
				ExternalId: "mycoolid",
			})},
		},
		{
			name: "three markets with same id: first",
			setup: func(ctx sdk.Context) {
				store := s.k.GetStore(ctx)
				s.requireSetOrderInStore(store, exchange.NewOrder(54).WithBid(&exchange.BidOrder{
					MarketId: 88, Buyer: s.addr2.String(), Assets: s.coin("1apple"), Price: s.coin("1plum"),
					ExternalId: "commonid",
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(55).WithBid(&exchange.BidOrder{
					MarketId: 2, Buyer: s.addr2.String(), Assets: s.coin("2apple"), Price: s.coin("2plum"),
					ExternalId: "commonid",
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(56).WithBid(&exchange.BidOrder{
					MarketId: 9001, Buyer: s.addr2.String(), Assets: s.coin("3apple"), Price: s.coin("3plum"),
					ExternalId: "commonid",
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(57).WithBid(&exchange.BidOrder{
					MarketId: 88, Buyer: s.addr2.String(), Assets: s.coin("1apple"), Price: s.coin("1plum"),
					ExternalId: "otherid1",
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(58).WithBid(&exchange.BidOrder{
					MarketId: 88, Buyer: s.addr2.String(), Assets: s.coin("1apple"), Price: s.coin("1plum"),
					ExternalId: "otherid2",
				}))
			},
			req: &exchange.QueryGetOrderByExternalIDRequest{MarketId: 88, ExternalId: "commonid"},
			expResp: &exchange.QueryGetOrderByExternalIDResponse{Order: exchange.NewOrder(54).WithBid(&exchange.BidOrder{
				MarketId: 88, Buyer: s.addr2.String(), Assets: s.coin("1apple"), Price: s.coin("1plum"),
				ExternalId: "commonid",
			})},
		},
		{
			name: "three markets with same id: second",
			setup: func(ctx sdk.Context) {
				store := s.k.GetStore(ctx)
				s.requireSetOrderInStore(store, exchange.NewOrder(54).WithBid(&exchange.BidOrder{
					MarketId: 88, Buyer: s.addr2.String(), Assets: s.coin("1apple"), Price: s.coin("1plum"),
					ExternalId: "commonid",
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(55).WithBid(&exchange.BidOrder{
					MarketId: 2, Buyer: s.addr2.String(), Assets: s.coin("2apple"), Price: s.coin("2plum"),
					ExternalId: "commonid",
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(56).WithBid(&exchange.BidOrder{
					MarketId: 9001, Buyer: s.addr2.String(), Assets: s.coin("3apple"), Price: s.coin("3plum"),
					ExternalId: "commonid",
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(57).WithBid(&exchange.BidOrder{
					MarketId: 2, Buyer: s.addr2.String(), Assets: s.coin("1apple"), Price: s.coin("1plum"),
					ExternalId: "otherid1",
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(58).WithBid(&exchange.BidOrder{
					MarketId: 2, Buyer: s.addr2.String(), Assets: s.coin("1apple"), Price: s.coin("1plum"),
					ExternalId: "otherid2",
				}))
			},
			req: &exchange.QueryGetOrderByExternalIDRequest{MarketId: 2, ExternalId: "commonid"},
			expResp: &exchange.QueryGetOrderByExternalIDResponse{Order: exchange.NewOrder(55).WithBid(&exchange.BidOrder{
				MarketId: 2, Buyer: s.addr2.String(), Assets: s.coin("2apple"), Price: s.coin("2plum"),
				ExternalId: "commonid",
			})},
		},
		{
			name: "three markets with same id: second",
			setup: func(ctx sdk.Context) {
				store := s.k.GetStore(ctx)
				s.requireSetOrderInStore(store, exchange.NewOrder(54).WithBid(&exchange.BidOrder{
					MarketId: 88, Buyer: s.addr2.String(), Assets: s.coin("1apple"), Price: s.coin("1plum"),
					ExternalId: "commonid",
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(55).WithBid(&exchange.BidOrder{
					MarketId: 2, Buyer: s.addr2.String(), Assets: s.coin("2apple"), Price: s.coin("2plum"),
					ExternalId: "commonid",
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(56).WithBid(&exchange.BidOrder{
					MarketId: 9001, Buyer: s.addr2.String(), Assets: s.coin("3apple"), Price: s.coin("3plum"),
					ExternalId: "commonid",
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(57).WithBid(&exchange.BidOrder{
					MarketId: 9001, Buyer: s.addr2.String(), Assets: s.coin("1apple"), Price: s.coin("1plum"),
					ExternalId: "otherid1",
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(58).WithBid(&exchange.BidOrder{
					MarketId: 9001, Buyer: s.addr2.String(), Assets: s.coin("1apple"), Price: s.coin("1plum"),
					ExternalId: "otherid2",
				}))
			},
			req: &exchange.QueryGetOrderByExternalIDRequest{MarketId: 9001, ExternalId: "commonid"},
			expResp: &exchange.QueryGetOrderByExternalIDResponse{Order: exchange.NewOrder(56).WithBid(&exchange.BidOrder{
				MarketId: 9001, Buyer: s.addr2.String(), Assets: s.coin("3apple"), Price: s.coin("3plum"),
				ExternalId: "commonid",
			})},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			respRaw := s.doQueryTest(tc.setup, runner(tc.req), tc.expInErr, queryName)
			s.Assert().Equal(tc.expResp, respRaw, queryName+" result")
		})
	}
}

// TODO[1658]: func (s *TestSuite) TestQueryServer_GetMarketOrders()

// TODO[1658]: func (s *TestSuite) TestQueryServer_GetOwnerOrders()

// TODO[1658]: func (s *TestSuite) TestQueryServer_GetAssetOrders()

// TODO[1658]: func (s *TestSuite) TestQueryServer_GetAllOrders()

// TODO[1658]: func (s *TestSuite) TestQueryServer_GetMarket()

// TODO[1658]: func (s *TestSuite) TestQueryServer_GetAllMarkets()

// TODO[1658]: func (s *TestSuite) TestQueryServer_Params()

// TODO[1658]: func (s *TestSuite) TestQueryServer_ValidateCreateMarket()

// TODO[1658]: func (s *TestSuite) TestQueryServer_ValidateMarket()

// TODO[1658]: func (s *TestSuite) TestQueryServer_ValidateManageFees()
