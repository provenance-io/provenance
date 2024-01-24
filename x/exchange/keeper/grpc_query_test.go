package keeper_test

import (
	"context"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/provenance-io/provenance/x/exchange"
	"github.com/provenance-io/provenance/x/exchange/keeper"
)

const invalidArgErr = "rpc error: code = InvalidArgument"

// logKeyAsID will log the provided key as either an order id or market id if it's suspected of being one of those.
func (s *TestSuite) logKeyAsID(name string, key []byte) {
	switch len(key) {
	case 8:
		id, ok := keeper.ParseIndexKeySuffixOrderID(key)
		if ok {
			s.T().Logf("%s as order id: %d", name, id)
		}
	case 4:
		id, ok := keeper.ParseKeySuffixKnownMarketID(key)
		if ok {
			s.T().Logf("%s as market id: %d", name, id)
		}
	}
}

// assertEqualPageResponse asserts that two PageResponses are equal.
// If not, some further assertions are made to try to help clarify the differences in the failure messages.
func (s *TestSuite) assertEqualPageResponse(expected, actual *query.PageResponse, msg string, args ...interface{}) bool {
	s.T().Helper()
	if s.Assert().Equalf(expected, actual, msg, args...) {
		return true
	}
	if expected == nil || actual == nil {
		return false
	}
	if !s.Assert().Equalf(expected.NextKey, actual.NextKey, msg+" NextKey", args...) {
		s.logKeyAsID("Expected", expected.NextKey)
		s.logKeyAsID("  Actual", actual.NextKey)
	}
	s.Assert().Equalf(int(expected.Total), int(actual.Total), msg+" Total", args...)
	return false
}

// queryTestDef is the definition of a QueryServer endpoint to be tested.
// R is the request message type. S is the response message type.
type queryTestDef[R any, S any] struct {
	// queryName is the name of the query being tested.
	queryName string
	// query is the query function to invoke.
	query func(goCtx context.Context, req *R) (*S, error)
	// followup is a function that runs any desired followup assertions to help pinpoint
	// differences between the expected and actual. It's only called if they're not equal and neither are nil.
	followup func(expected, actual *S)
}

// queryTestCase is a test case for a QueryServer endpoint.
// R is the request message type. S is the response message type.
type queryTestCase[R any, S any] struct {
	// name is the name of the test case.
	name string
	// setup is a function that does any needed app/state setup.
	// A cached context is used for tests, so this setup will not carry over between test cases.
	setup func()
	// req is the request message to provide to the query.
	req *R
	// expResp is the expected response from the query
	expResp *S
	// expInErr is the strings that are expected to be in the error returned by the endpoint.
	// If empty, that error is expected to be nil.
	expInErr []string
}

// runQueryTestCase runs a unit test on a QueryServer endpoint.
// A cached context is used so each test case won't affect the others.
// R is the request message type. S is the response message type.
func runQueryTestCase[R any, S any](s *TestSuite, td queryTestDef[R, S], tc queryTestCase[R, S]) {
	origCtx := s.ctx
	defer func() {
		s.ctx = origCtx
	}()
	s.ctx, _ = s.ctx.CacheContext()

	if tc.setup != nil {
		tc.setup()
	}

	goCtx := sdk.WrapSDKContext(s.ctx)
	var resp *S
	var err error
	testFunc := func() {
		resp, err = td.query(goCtx, tc.req)
	}
	s.Require().NotPanics(testFunc, td.queryName)
	s.assertErrorContentsf(err, tc.expInErr, "%s error", td.queryName)
	if s.Assert().Equal(tc.expResp, resp, "%s response", td.queryName) {
		return
	}
	if td.followup != nil && tc.expResp != nil && resp != nil {
		td.followup(tc.expResp, resp)
	}
}

func (s *TestSuite) TestQueryServer_OrderFeeCalc() {
	testDef := queryTestDef[exchange.QueryOrderFeeCalcRequest, exchange.QueryOrderFeeCalcResponse]{
		queryName: "OrderFeeCalc",
		query:     keeper.NewQueryServer(s.k).OrderFeeCalc,
		followup: func(expected, actual *exchange.QueryOrderFeeCalcResponse) {
			s.Assert().Equal(s.coinsString(expected.CreationFeeOptions), s.coinsString(actual.CreationFeeOptions),
				"CreationFeeOptions (as strings)")
			s.Assert().Equal(s.coinsString(expected.SettlementFlatFeeOptions), s.coinsString(actual.SettlementFlatFeeOptions),
				"SettlementFlatFeeOptions (as strings)")
			s.Assert().Equal(s.coinsString(expected.SettlementRatioFeeOptions), s.coinsString(actual.SettlementRatioFeeOptions),
				"SettlementRatioFeeOptions (as strings)")
		},
	}

	tests := []queryTestCase[exchange.QueryOrderFeeCalcRequest, exchange.QueryOrderFeeCalcResponse]{
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
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{MarketId: 1})
			},
			req: &exchange.QueryOrderFeeCalcRequest{AskOrder: &exchange.AskOrder{
				Assets: s.coin("1apple"), Price: s.coin("2000plum"), MarketId: 1,
			}},
			expResp: &exchange.QueryOrderFeeCalcResponse{},
		},
		{
			name: "ask: only creation: one option",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
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
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
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
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
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
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
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
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
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
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
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
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
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
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
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
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{MarketId: 1})
			},
			req: &exchange.QueryOrderFeeCalcRequest{BidOrder: &exchange.BidOrder{
				Assets: s.coin("1apple"), Price: s.coin("2000plum"), MarketId: 1,
			}},
			expResp: &exchange.QueryOrderFeeCalcResponse{},
		},
		{
			name: "bid: only creation: one option",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
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
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
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
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
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
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
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
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
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
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
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
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
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
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
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
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
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
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
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
			runQueryTestCase(s, testDef, tc)
		})
	}
}

func (s *TestSuite) TestQueryServer_GetOrder() {
	testDef := queryTestDef[exchange.QueryGetOrderRequest, exchange.QueryGetOrderResponse]{
		queryName: "GetOrder",
		query:     keeper.NewQueryServer(s.k).GetOrder,
	}

	tests := []queryTestCase[exchange.QueryGetOrderRequest, exchange.QueryGetOrderResponse]{
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
			setup: func() {
				key, value, err := s.k.GetOrderStoreKeyValue(*exchange.NewOrder(4).WithAsk(&exchange.AskOrder{
					MarketId: 1,
					Seller:   s.addr1.String(),
					Assets:   s.coin("55apple"),
					Price:    s.coin("99prune"),
				}))
				s.Require().NoError(err, "GetOrderStoreKeyValue 4")
				value[0] = 9
				s.getStore().Set(key, value)
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
			setup: func() {
				store := s.getStore()
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
			setup: func() {
				store := s.getStore()
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
			setup: func() {
				store := s.getStore()
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
			runQueryTestCase(s, testDef, tc)
		})
	}
}

func (s *TestSuite) TestQueryServer_GetOrderByExternalID() {
	testDef := queryTestDef[exchange.QueryGetOrderByExternalIDRequest, exchange.QueryGetOrderByExternalIDResponse]{
		queryName: "GetOrderByExternalID",
		query:     keeper.NewQueryServer(s.k).GetOrderByExternalID,
	}

	tests := []queryTestCase[exchange.QueryGetOrderByExternalIDRequest, exchange.QueryGetOrderByExternalIDResponse]{
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
			setup: func() {
				order5 := exchange.NewOrder(5).WithBid(&exchange.BidOrder{
					MarketId:            1,
					Buyer:               s.addr3.String(),
					Assets:              s.coin("1apple"),
					Price:               s.coin("1plum"),
					BuyerSettlementFees: nil,
					AllowPartial:        false,
					ExternalId:          "babbaderr",
				})
				store := s.getStore()
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
			setup: func() {
				store := s.getStore()
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
			setup: func() {
				store := s.getStore()
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
			setup: func() {
				store := s.getStore()
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
			setup: func() {
				store := s.getStore()
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
			setup: func() {
				store := s.getStore()
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
			setup: func() {
				store := s.getStore()
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
			runQueryTestCase(s, testDef, tc)
		})
	}
}

func (s *TestSuite) TestQueryServer_GetMarketOrders() {
	testDef := queryTestDef[exchange.QueryGetMarketOrdersRequest, exchange.QueryGetMarketOrdersResponse]{
		queryName: "GetMarketOrders",
		query:     keeper.NewQueryServer(s.k).GetMarketOrders,
		followup: func(expected, actual *exchange.QueryGetMarketOrdersResponse) {
			s.assertEqualOrders(expected.Orders, actual.Orders, "Orders")
			s.assertEqualPageResponse(expected.Pagination, actual.Pagination, "Pagination")
		},
	}
	makeKey := func(order *exchange.Order) []byte {
		return keeper.Uint64Bz(order.OrderId)
	}

	marketCount, ordersPerMarket := 3, 20
	marketOrders := make(map[uint32][]*exchange.Order, marketCount)
	marketAskOrders := make(map[uint32][]*exchange.Order, marketCount)
	marketBidOrders := make(map[uint32][]*exchange.Order, marketCount)
	for m := uint32(1); m <= uint32(marketCount); m++ {
		marketOrders[m] = make([]*exchange.Order, 0, ordersPerMarket)
		marketAskOrders[m] = make([]*exchange.Order, 0, ordersPerMarket/2)
		marketBidOrders[m] = make([]*exchange.Order, 0, ordersPerMarket/2)
	}
	mainStore := s.getStore()
	for i := 1; i <= marketCount*ordersPerMarket; i++ {
		orderID := uint64(i)
		marketID := uint32((i-1)%marketCount) + 1
		order := exchange.NewOrder(orderID)
		if orderID%2 == 0 {
			order.WithAsk(&exchange.AskOrder{
				MarketId:     marketID,
				Seller:       sdk.AccAddress(fmt.Sprintf("seller_%d____________", orderID)[:20]).String(),
				Assets:       sdk.NewInt64Coin("apple", int64(i)),
				Price:        sdk.NewInt64Coin("plum", int64(i)),
				AllowPartial: orderID%4 < 2,
				ExternalId:   fmt.Sprintf("external-id-%d", i),
			})
			marketAskOrders[marketID] = append(marketAskOrders[marketID], order)
		} else {
			order.WithBid(&exchange.BidOrder{
				MarketId:     marketID,
				Buyer:        sdk.AccAddress(fmt.Sprintf("buyer_%d_____________", orderID)[:20]).String(),
				Assets:       sdk.NewInt64Coin("apple", int64(i)),
				Price:        sdk.NewInt64Coin("plum", int64(i)),
				AllowPartial: orderID%4 < 2,
				ExternalId:   fmt.Sprintf("external-id-%d", i),
			})
			marketBidOrders[marketID] = append(marketBidOrders[marketID], order)
		}
		marketOrders[marketID] = append(marketOrders[marketID], order)
		s.requireSetOrderInStore(mainStore, order)
	}

	// OrderIDs in each market:
	//   0  1  2   3   4   5   6   7   8   9  10  11  12  13  14  15  16  17  18  19
	//1: 1, 4, 7, 10, 13, 16, 19, 22, 25, 28, 31, 34, 37, 40, 43, 46, 49, 52, 55, 58
	//2: 2, 5, 8, 11, 14, 17, 20, 23, 26, 29, 32, 35, 38, 41, 44, 47, 50, 53, 56, 59
	//3: 3, 6, 9, 12, 15, 18, 21, 24, 27, 30, 33, 36, 39, 42, 45, 48, 51, 54, 57, 60

	tests := []queryTestCase[exchange.QueryGetMarketOrdersRequest, exchange.QueryGetMarketOrdersResponse]{
		// Tests on errors and non-normal conditions.
		{
			name:     "nil req",
			req:      nil,
			expInErr: []string{invalidArgErr, "empty request"},
		},
		{
			name:     "market 0",
			req:      &exchange.QueryGetMarketOrdersRequest{MarketId: 0},
			expInErr: []string{invalidArgErr, "empty request"},
		},
		{
			name:     "unknown order type",
			req:      &exchange.QueryGetMarketOrdersRequest{MarketId: 4, OrderType: "burger and fries"},
			expInErr: []string{invalidArgErr, "error iterating orders for market 4: unknown order type \"burger and fries\""},
		},
		{
			name:    "no orders",
			req:     &exchange.QueryGetMarketOrdersRequest{MarketId: 8},
			expResp: &exchange.QueryGetMarketOrdersResponse{Orders: nil, Pagination: &query.PageResponse{}},
		},
		{
			name: "bad index entry",
			setup: func() {
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(98).WithAsk(&exchange.AskOrder{
					MarketId: 8, Seller: s.addr1.String(), Assets: s.coin("98apple"), Price: s.coin("98prune"),
				}))
				key99, value99, err := s.k.GetOrderStoreKeyValue(*exchange.NewOrder(99).WithAsk(&exchange.AskOrder{
					MarketId: 8, Seller: s.addr2.String(), Assets: s.coin("99apple"), Price: s.coin("99prune"),
				}))
				s.Require().NoError(err, "GetOrderStoreKeyValue 99")
				store.Set(key99, value99)
				store.Set(s.badKey(keeper.MakeIndexKeyMarketToOrder(8, 99)), []byte{keeper.OrderKeyTypeAsk})
				s.requireSetOrderInStore(store, exchange.NewOrder(100).WithAsk(&exchange.AskOrder{
					MarketId: 8, Seller: s.addr3.String(), Assets: s.coin("100apple"), Price: s.coin("100prune"),
				}))
			},
			req: &exchange.QueryGetMarketOrdersRequest{MarketId: 8},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders: []*exchange.Order{
					exchange.NewOrder(98).WithAsk(&exchange.AskOrder{
						MarketId: 8, Seller: s.addr1.String(), Assets: s.coin("98apple"), Price: s.coin("98prune"),
					}),
					exchange.NewOrder(100).WithAsk(&exchange.AskOrder{
						MarketId: 8, Seller: s.addr3.String(), Assets: s.coin("100apple"), Price: s.coin("100prune"),
					}),
				},
				Pagination: &query.PageResponse{Total: 2},
			},
		},
		{
			name: "index entry to order that does not exist",
			setup: func() {
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(98).WithAsk(&exchange.AskOrder{
					MarketId: 8, Seller: s.addr1.String(), Assets: s.coin("98apple"), Price: s.coin("98prune"),
				}))
				key := keeper.MakeIndexKeyMarketToOrder(8, 99)
				store.Set(key, []byte{keeper.OrderKeyTypeAsk})
				s.requireSetOrderInStore(store, exchange.NewOrder(100).WithAsk(&exchange.AskOrder{
					MarketId: 8, Seller: s.addr3.String(), Assets: s.coin("100apple"), Price: s.coin("100prune"),
				}))
			},
			req: &exchange.QueryGetMarketOrdersRequest{MarketId: 8},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders: []*exchange.Order{
					exchange.NewOrder(98).WithAsk(&exchange.AskOrder{
						MarketId: 8, Seller: s.addr1.String(), Assets: s.coin("98apple"), Price: s.coin("98prune"),
					}),
					exchange.NewOrder(100).WithAsk(&exchange.AskOrder{
						MarketId: 8, Seller: s.addr3.String(), Assets: s.coin("100apple"), Price: s.coin("100prune"),
					}),
				},
				Pagination: &query.PageResponse{Total: 3},
			},
		},
		{
			name: "error reading an order",
			setup: func() {
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(98).WithAsk(&exchange.AskOrder{
					MarketId: 8, Seller: s.addr1.String(), Assets: s.coin("98apple"), Price: s.coin("98prune"),
				}))
				key99, value99, err := s.k.GetOrderStoreKeyValue(*exchange.NewOrder(99).WithAsk(&exchange.AskOrder{
					MarketId: 8, Seller: s.addr2.String(), Assets: s.coin("99apple"), Price: s.coin("99prune"),
				}))
				s.Require().NoError(err, "GetOrderStoreKeyValue 99")
				value99[0] = 8
				store.Set(key99, value99)
				idxKey := keeper.MakeIndexKeyMarketToOrder(8, 99)
				store.Set(idxKey, []byte{8})
				s.requireSetOrderInStore(store, exchange.NewOrder(100).WithAsk(&exchange.AskOrder{
					MarketId: 8, Seller: s.addr3.String(), Assets: s.coin("100apple"), Price: s.coin("100prune"),
				}))
			},
			req: &exchange.QueryGetMarketOrdersRequest{MarketId: 8},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders: []*exchange.Order{
					exchange.NewOrder(98).WithAsk(&exchange.AskOrder{
						MarketId: 8, Seller: s.addr1.String(), Assets: s.coin("98apple"), Price: s.coin("98prune"),
					}),
					exchange.NewOrder(100).WithAsk(&exchange.AskOrder{
						MarketId: 8, Seller: s.addr3.String(), Assets: s.coin("100apple"), Price: s.coin("100prune"),
					}),
				},
				Pagination: &query.PageResponse{Total: 3},
			},
		},
		{
			name: "both offset and key provided",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId:   1,
				Pagination: &query.PageRequest{Offset: 2, Key: makeKey(marketOrders[1][2])},
			},
			expInErr: []string{invalidArgErr, "error iterating orders for market 1",
				"invalid request, either offset or key is expected, got both"},
		},

		// Forward, no order type.
		{
			name: "forward, no order type, no after order, get all",
			req:  &exchange.QueryGetMarketOrdersRequest{MarketId: 1},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     marketOrders[1],
				Pagination: &query.PageResponse{Total: 20},
			},
		},
		{
			name: "forward, no order type, no after order, limit with key",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId:   2,
				Pagination: &query.PageRequest{Limit: 3, Key: makeKey(marketOrders[2][2])},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     marketOrders[2][2:5],
				Pagination: &query.PageResponse{NextKey: makeKey(marketOrders[2][5])},
			},
		},
		{
			name: "forward, no order type, no after order, limit with offset, no count",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId:   3,
				Pagination: &query.PageRequest{Limit: 5, Offset: 8, CountTotal: false},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     marketOrders[3][8:13],
				Pagination: &query.PageResponse{NextKey: makeKey(marketOrders[3][13])},
			},
		},
		{
			name: "forward, no order type, no after order, limit with offset and count",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId:   1,
				Pagination: &query.PageRequest{Limit: 5, Offset: 6, CountTotal: true},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     marketOrders[1][6:11],
				Pagination: &query.PageResponse{NextKey: makeKey(marketOrders[1][11]), Total: 20},
			},
		},
		{
			name: "forward, no order type, after order 30, get all",
			req:  &exchange.QueryGetMarketOrdersRequest{MarketId: 2, AfterOrderId: 30},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     marketOrders[2][10:],
				Pagination: &query.PageResponse{Total: 10},
			},
		},
		{
			name: "forward, no order type, after order 30, limit with key",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 1, AfterOrderId: 30,
				Pagination: &query.PageRequest{Limit: 2, Key: makeKey(marketOrders[1][15])},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     marketOrders[1][15:17],
				Pagination: &query.PageResponse{NextKey: makeKey(marketOrders[1][17])},
			},
		},
		{
			name: "forward, no order type, after order 30, limit with offset, no count",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 1, AfterOrderId: 30,
				Pagination: &query.PageRequest{Limit: 3, Offset: 2, CountTotal: false},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     marketOrders[1][12:15],
				Pagination: &query.PageResponse{NextKey: makeKey(marketOrders[1][15])},
			},
		},
		{
			name: "forward, no order type, after order 30, limit with offset and count",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 3, AfterOrderId: 30,
				Pagination: &query.PageRequest{Limit: 1, Offset: 7, CountTotal: true},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     marketOrders[3][17:18],
				Pagination: &query.PageResponse{NextKey: makeKey(marketOrders[3][18]), Total: 10},
			},
		},

		// Forward, only ask orders
		{
			name: "forward, ask orders, no after order, get all",
			req:  &exchange.QueryGetMarketOrdersRequest{MarketId: 3, OrderType: "ask"},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     marketAskOrders[3],
				Pagination: &query.PageResponse{Total: 10},
			},
		},
		{
			name: "forward, ask orders, no after order, limit with key",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 1, OrderType: "asks",
				Pagination: &query.PageRequest{Limit: 3, Key: makeKey(marketAskOrders[1][4])},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     marketAskOrders[1][4:7],
				Pagination: &query.PageResponse{NextKey: makeKey(marketAskOrders[1][7])},
			},
		},
		{
			name: "forward, ask orders, no after order, limit with offset, no count",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 2, OrderType: "ASK",
				Pagination: &query.PageRequest{Limit: 3, Offset: 8, CountTotal: false},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     marketAskOrders[2][8:],
				Pagination: &query.PageResponse{},
			},
		},
		{
			name: "forward, ask orders, no after order, limit with offset and count",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 2, OrderType: "ASKS",
				Pagination: &query.PageRequest{Limit: 3, Offset: 6, CountTotal: true},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     marketAskOrders[2][6:9],
				Pagination: &query.PageResponse{NextKey: makeKey(marketAskOrders[2][9]), Total: 10},
			},
		},
		{
			name: "forward, ask orders, after order 30, get all",
			req:  &exchange.QueryGetMarketOrdersRequest{MarketId: 3, OrderType: "AskOrders", AfterOrderId: 30},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     marketAskOrders[3][5:],
				Pagination: &query.PageResponse{Total: 5},
			},
		},
		{
			name: "forward, ask orders, after order 30, limit with key",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 2, OrderType: "ask orders", AfterOrderId: 30,
				Pagination: &query.PageRequest{Limit: 1, Key: makeKey(marketAskOrders[2][7])},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     marketAskOrders[2][7:8],
				Pagination: &query.PageResponse{NextKey: makeKey(marketAskOrders[2][8])},
			},
		},
		{
			name: "forward, ask orders, after order 30, limit with offset, no count",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 1, OrderType: "askOrder", AfterOrderId: 30,
				Pagination: &query.PageRequest{Limit: 2, Offset: 2, CountTotal: false},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     marketAskOrders[1][7:9],
				Pagination: &query.PageResponse{NextKey: makeKey(marketAskOrders[1][9])},
			},
		},
		{
			name: "forward, ask orders, after order 30, limit with offset and count",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 1, OrderType: "aSKs", AfterOrderId: 30,
				Pagination: &query.PageRequest{Limit: 2, Offset: 1, CountTotal: true},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     marketAskOrders[1][6:8],
				Pagination: &query.PageResponse{NextKey: makeKey(marketAskOrders[1][8]), Total: 5},
			},
		},

		// Forward, only bid orders
		{
			name: "forward, bid orders, no after order, get all",
			req:  &exchange.QueryGetMarketOrdersRequest{MarketId: 3, OrderType: "bid"},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     marketBidOrders[3],
				Pagination: &query.PageResponse{Total: 10},
			},
		},
		{
			name: "forward, bid orders, no after order, limit with key",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 1, OrderType: "bids",
				Pagination: &query.PageRequest{Limit: 3, Key: makeKey(marketBidOrders[1][4])},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     marketBidOrders[1][4:7],
				Pagination: &query.PageResponse{NextKey: makeKey(marketBidOrders[1][7])},
			},
		},
		{
			name: "forward, bid orders, no after order, limit with offset, no count",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 2, OrderType: "BID",
				Pagination: &query.PageRequest{Limit: 3, Offset: 8, CountTotal: false},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     marketBidOrders[2][8:],
				Pagination: &query.PageResponse{},
			},
		},
		{
			name: "forward, bid orders, no after order, limit with offset and count",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 2, OrderType: "BIDS",
				Pagination: &query.PageRequest{Limit: 3, Offset: 6, CountTotal: true},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     marketBidOrders[2][6:9],
				Pagination: &query.PageResponse{NextKey: makeKey(marketBidOrders[2][9]), Total: 10},
			},
		},
		{
			name: "forward, bid orders, after order 30, get all",
			req:  &exchange.QueryGetMarketOrdersRequest{MarketId: 3, OrderType: "BidOrders", AfterOrderId: 30},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     marketBidOrders[3][5:],
				Pagination: &query.PageResponse{Total: 5},
			},
		},
		{
			name: "forward, bid orders, after order 30, limit with key",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 2, OrderType: "bid orders", AfterOrderId: 30,
				Pagination: &query.PageRequest{Limit: 1, Key: makeKey(marketBidOrders[2][7])},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     marketBidOrders[2][7:8],
				Pagination: &query.PageResponse{NextKey: makeKey(marketBidOrders[2][8])},
			},
		},
		{
			name: "forward, bid orders, after order 30, limit with offset, no count",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 1, OrderType: "bidOrder", AfterOrderId: 30,
				Pagination: &query.PageRequest{Limit: 2, Offset: 2, CountTotal: false},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     marketBidOrders[1][7:9],
				Pagination: &query.PageResponse{NextKey: makeKey(marketBidOrders[1][9])},
			},
		},
		{
			name: "forward, bid orders, after order 30, limit with offset and count",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 1, OrderType: "bIDs", AfterOrderId: 30,
				Pagination: &query.PageRequest{Limit: 2, Offset: 1, CountTotal: true},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     marketBidOrders[1][6:8],
				Pagination: &query.PageResponse{NextKey: makeKey(marketBidOrders[1][8]), Total: 5},
			},
		},

		// Reverse, no order type.
		{
			name: "reverse, no order type, no after order, get all",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId:   1,
				Pagination: &query.PageRequest{Reverse: true},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     reverseSlice(marketOrders[1]),
				Pagination: &query.PageResponse{Total: 20},
			},
		},
		{
			name: "reverse, no order type, no after order, limit with key",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId:   2,
				Pagination: &query.PageRequest{Reverse: true, Limit: 3, Key: makeKey(marketOrders[2][12])},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     reverseSlice(marketOrders[2][10:13]),
				Pagination: &query.PageResponse{NextKey: makeKey(marketOrders[2][9])},
			},
		},
		{
			name: "reverse, no order type, no after order, limit with offset, no count",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId:   3,
				Pagination: &query.PageRequest{Reverse: true, Limit: 5, Offset: 8, CountTotal: false},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     reverseSlice(marketOrders[3][7:12]),
				Pagination: &query.PageResponse{NextKey: makeKey(marketOrders[3][6])},
			},
		},
		{
			name: "reverse, no order type, no after order, limit with offset and count",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId:   1,
				Pagination: &query.PageRequest{Reverse: true, Limit: 5, Offset: 6, CountTotal: true},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     reverseSlice(marketOrders[1][9:14]),
				Pagination: &query.PageResponse{NextKey: makeKey(marketOrders[1][8]), Total: 20},
			},
		},
		{
			name: "reverse, no order type, after order 30, get all",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 2, AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     reverseSlice(marketOrders[2][10:]),
				Pagination: &query.PageResponse{Total: 10},
			},
		},
		{
			name: "reverse, no order type, after order 30, limit with key",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 1, AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true, Limit: 2, Key: makeKey(marketOrders[1][15])},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     reverseSlice(marketOrders[1][14:16]),
				Pagination: &query.PageResponse{NextKey: makeKey(marketOrders[1][13])},
			},
		},
		{
			name: "reverse, no order type, after order 30, limit with offset, no count",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 1, AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true, Limit: 3, Offset: 2, CountTotal: false},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     reverseSlice(marketOrders[1][15:18]),
				Pagination: &query.PageResponse{NextKey: makeKey(marketOrders[1][14])},
			},
		},
		{
			name: "reverse, no order type, after order 30, limit with offset and count",
			// A key point of this test is that order 30 is in market 3. The AfterOrderID order
			// should NOT be included in results, though, so there should still only be 10 results here.
			// This validates that the "afterOrderID + 1" is correct in the getOrderIterator reverse block.
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 3, AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true, Limit: 1, Offset: 7, CountTotal: true},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     reverseSlice(marketOrders[3][12:13]),
				Pagination: &query.PageResponse{NextKey: makeKey(marketOrders[3][11]), Total: 10},
			},
		},

		// Reverse, only ask orders
		{
			name: "reverse, ask orders, no after order, get all",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 3, OrderType: "ask",
				Pagination: &query.PageRequest{Reverse: true},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     reverseSlice(marketAskOrders[3]),
				Pagination: &query.PageResponse{Total: 10},
			},
		},
		{
			name: "reverse, ask orders, no after order, limit with key",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 1, OrderType: "asks",
				Pagination: &query.PageRequest{Reverse: true, Limit: 3, Key: makeKey(marketAskOrders[1][4])},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     reverseSlice(marketAskOrders[1][2:5]),
				Pagination: &query.PageResponse{NextKey: makeKey(marketAskOrders[1][1])},
			},
		},
		{
			name: "reverse, ask orders, no after order, limit with offset, no count",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 2, OrderType: "ASK",
				Pagination: &query.PageRequest{Reverse: true, Limit: 3, Offset: 8, CountTotal: false},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     reverseSlice(marketAskOrders[2][:2]),
				Pagination: &query.PageResponse{},
			},
		},
		{
			name: "reverse, ask orders, no after order, limit with offset and count",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 2, OrderType: "ASKS",
				Pagination: &query.PageRequest{Reverse: true, Limit: 3, Offset: 1, CountTotal: true},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     reverseSlice(marketAskOrders[2][6:9]),
				Pagination: &query.PageResponse{NextKey: makeKey(marketAskOrders[2][5]), Total: 10},
			},
		},
		{
			name: "reverse, ask orders, after order 30, get all",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 3, OrderType: "AskOrders", AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     reverseSlice(marketAskOrders[3][5:]),
				Pagination: &query.PageResponse{Total: 5},
			},
		},
		{
			name: "reverse, ask orders, after order 30, limit with key",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 2, OrderType: "ask orders", AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true, Limit: 1, Key: makeKey(marketAskOrders[2][7])},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     reverseSlice(marketAskOrders[2][7:8]),
				Pagination: &query.PageResponse{NextKey: makeKey(marketAskOrders[2][6])},
			},
		},
		{
			name: "reverse, ask orders, after order 30, limit with offset, no count",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 1, OrderType: "askOrder", AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true, Limit: 2, Offset: 2, CountTotal: false},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     reverseSlice(marketAskOrders[1][6:8]),
				Pagination: &query.PageResponse{NextKey: makeKey(marketAskOrders[1][5])},
			},
		},
		{
			name: "reverse, ask orders, after order 30, limit with offset and count",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 1, OrderType: "aSKs", AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true, Limit: 3, Offset: 1, CountTotal: true},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     reverseSlice(marketAskOrders[1][6:9]),
				Pagination: &query.PageResponse{NextKey: makeKey(marketAskOrders[1][5]), Total: 5},
			},
		},

		// Reverse, only bid orders
		{
			name: "reverse, bid orders, no after order, get all",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 3, OrderType: "bid",
				Pagination: &query.PageRequest{Reverse: true},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     reverseSlice(marketBidOrders[3]),
				Pagination: &query.PageResponse{Total: 10},
			},
		},
		{
			name: "reverse, bid orders, no after order, limit with key",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 1, OrderType: "bids",
				Pagination: &query.PageRequest{Reverse: true, Limit: 3, Key: makeKey(marketBidOrders[1][4])},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     reverseSlice(marketBidOrders[1][2:5]),
				Pagination: &query.PageResponse{NextKey: makeKey(marketBidOrders[1][1])},
			},
		},
		{
			name: "reverse, bid orders, no after order, limit with offset, no count",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 2, OrderType: "BID",
				Pagination: &query.PageRequest{Reverse: true, Limit: 3, Offset: 8, CountTotal: false},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     reverseSlice(marketBidOrders[2][:2]),
				Pagination: &query.PageResponse{},
			},
		},
		{
			name: "reverse, bid orders, no after order, limit with offset and count",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 2, OrderType: "BIDS",
				Pagination: &query.PageRequest{Reverse: true, Limit: 3, Offset: 1, CountTotal: true},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     reverseSlice(marketBidOrders[2][6:9]),
				Pagination: &query.PageResponse{NextKey: makeKey(marketBidOrders[2][5]), Total: 10},
			},
		},
		{
			name: "reverse, bid orders, after order 30, get all",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 3, OrderType: "BidOrders", AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     reverseSlice(marketBidOrders[3][5:]),
				Pagination: &query.PageResponse{Total: 5},
			},
		},
		{
			name: "reverse, bid orders, after order 30, limit with key",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 2, OrderType: "bid orders", AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true, Limit: 1, Key: makeKey(marketBidOrders[2][7])},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     reverseSlice(marketBidOrders[2][7:8]),
				Pagination: &query.PageResponse{NextKey: makeKey(marketBidOrders[2][6])},
			},
		},
		{
			name: "reverse, bid orders, after order 30, limit with offset, no count",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 1, OrderType: "bidOrder", AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true, Limit: 2, Offset: 2, CountTotal: false},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     reverseSlice(marketBidOrders[1][6:8]),
				Pagination: &query.PageResponse{NextKey: makeKey(marketBidOrders[1][5])},
			},
		},
		{
			name: "reverse, bid orders, after order 30, limit with offset and count",
			req: &exchange.QueryGetMarketOrdersRequest{
				MarketId: 1, OrderType: "bIDs", AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true, Limit: 3, Offset: 1, CountTotal: true},
			},
			expResp: &exchange.QueryGetMarketOrdersResponse{
				Orders:     reverseSlice(marketBidOrders[1][6:9]),
				Pagination: &query.PageResponse{NextKey: makeKey(marketBidOrders[1][5]), Total: 5},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			runQueryTestCase(s, testDef, tc)
		})
	}
}

func (s *TestSuite) TestQueryServer_GetOwnerOrders() {
	testDef := queryTestDef[exchange.QueryGetOwnerOrdersRequest, exchange.QueryGetOwnerOrdersResponse]{
		queryName: "GetOwnerOrders",
		query:     keeper.NewQueryServer(s.k).GetOwnerOrders,
		followup: func(expected, actual *exchange.QueryGetOwnerOrdersResponse) {
			s.assertEqualOrders(expected.Orders, actual.Orders, "Orders")
			s.assertEqualPageResponse(expected.Pagination, actual.Pagination, "Pagination")
		},
	}
	makeKey := func(order *exchange.Order) []byte {
		return keeper.Uint64Bz(order.OrderId)
	}

	addr1, addr2, addr3 := s.addr1.String(), s.addr2.String(), s.addr3.String()
	owners := []string{addr1, addr2, addr3}
	ownerCount := len(owners)
	ordersPerOwner := 20
	ownerOrders := make(map[string][]*exchange.Order, ownerCount)
	ownerAskOrders := make(map[string][]*exchange.Order, ownerCount)
	ownerBidOrders := make(map[string][]*exchange.Order, ownerCount)
	for _, owner := range owners {
		ownerOrders[owner] = make([]*exchange.Order, 0, ordersPerOwner)
		ownerAskOrders[owner] = make([]*exchange.Order, 0, ordersPerOwner/2)
		ownerBidOrders[owner] = make([]*exchange.Order, 0, ordersPerOwner/2)
	}
	mainStore := s.getStore()
	for i := 1; i <= ownerCount*ordersPerOwner; i++ {
		orderID := uint64(i)
		owner := owners[i%ownerCount]
		order := exchange.NewOrder(orderID)
		if orderID%2 == 0 {
			order.WithAsk(&exchange.AskOrder{
				MarketId:     1,
				Seller:       owner,
				Assets:       sdk.NewInt64Coin("apple", int64(i)),
				Price:        sdk.NewInt64Coin("plum", int64(i)),
				AllowPartial: orderID%4 < 2,
				ExternalId:   fmt.Sprintf("external-id-%d", i),
			})
			ownerAskOrders[owner] = append(ownerAskOrders[owner], order)
		} else {
			order.WithBid(&exchange.BidOrder{
				MarketId:     1,
				Buyer:        owner,
				Assets:       sdk.NewInt64Coin("apple", int64(i)),
				Price:        sdk.NewInt64Coin("plum", int64(i)),
				AllowPartial: orderID%4 < 2,
				ExternalId:   fmt.Sprintf("external-id-%d", i),
			})
			ownerBidOrders[owner] = append(ownerBidOrders[owner], order)
		}
		ownerOrders[owner] = append(ownerOrders[owner], order)
		s.requireSetOrderInStore(mainStore, order)
	}

	// OrderIDs for each owner:
	//       0  1  2   3   4   5   6   7   8   9  10  11  12  13  14  15  16  17  18  19
	//addr1: 1, 4, 7, 10, 13, 16, 19, 22, 25, 28, 31, 34, 37, 40, 43, 46, 49, 52, 55, 58
	//addr2: 2, 5, 8, 11, 14, 17, 20, 23, 26, 29, 32, 35, 38, 41, 44, 47, 50, 53, 56, 59
	//addr3: 3, 6, 9, 12, 15, 18, 21, 24, 27, 30, 33, 36, 39, 42, 45, 48, 51, 54, 57, 60

	tests := []queryTestCase[exchange.QueryGetOwnerOrdersRequest, exchange.QueryGetOwnerOrdersResponse]{
		// Tests on errors and non-normal conditions.
		{
			name:     "nil req",
			req:      nil,
			expInErr: []string{invalidArgErr, "empty request"},
		},
		{
			name:     "empty owner",
			req:      &exchange.QueryGetOwnerOrdersRequest{Owner: ""},
			expInErr: []string{invalidArgErr, "empty request"},
		},
		{
			name:     "invalid owner",
			req:      &exchange.QueryGetOwnerOrdersRequest{Owner: "notgonnawork"},
			expInErr: []string{invalidArgErr, "invalid owner \"notgonnawork\"", "decoding bech32 failed"},
		},
		{
			name:     "unknown order type",
			req:      &exchange.QueryGetOwnerOrdersRequest{Owner: addr1, OrderType: "burger and fries"},
			expInErr: []string{invalidArgErr, "error iterating orders for owner " + addr1 + ": unknown order type \"burger and fries\""},
		},
		{
			name:    "no orders",
			req:     &exchange.QueryGetOwnerOrdersRequest{Owner: s.addr4.String()},
			expResp: &exchange.QueryGetOwnerOrdersResponse{Orders: nil, Pagination: &query.PageResponse{}},
		},
		{
			name: "bad index entry",
			setup: func() {
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(98).WithAsk(&exchange.AskOrder{
					MarketId: 8, Seller: s.addr4.String(), Assets: s.coin("98apple"), Price: s.coin("98prune"),
				}))
				key99, value99, err := s.k.GetOrderStoreKeyValue(*exchange.NewOrder(99).WithAsk(&exchange.AskOrder{
					MarketId: 8, Seller: s.addr4.String(), Assets: s.coin("99apple"), Price: s.coin("99prune"),
				}))
				s.Require().NoError(err, "GetOrderStoreKeyValue 99")
				store.Set(key99, value99)
				store.Set(s.badKey(keeper.MakeIndexKeyAddressToOrder(s.addr4, 99)), []byte{keeper.OrderKeyTypeAsk})
				s.requireSetOrderInStore(store, exchange.NewOrder(100).WithAsk(&exchange.AskOrder{
					MarketId: 8, Seller: s.addr4.String(), Assets: s.coin("100apple"), Price: s.coin("100prune"),
				}))
			},
			req: &exchange.QueryGetOwnerOrdersRequest{Owner: s.addr4.String()},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders: []*exchange.Order{
					exchange.NewOrder(98).WithAsk(&exchange.AskOrder{
						MarketId: 8, Seller: s.addr4.String(), Assets: s.coin("98apple"), Price: s.coin("98prune"),
					}),
					exchange.NewOrder(100).WithAsk(&exchange.AskOrder{
						MarketId: 8, Seller: s.addr4.String(), Assets: s.coin("100apple"), Price: s.coin("100prune"),
					}),
				},
				Pagination: &query.PageResponse{Total: 2},
			},
		},
		{
			name: "index entry to order that does not exist",
			setup: func() {
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(98).WithAsk(&exchange.AskOrder{
					MarketId: 8, Seller: s.addr4.String(), Assets: s.coin("98apple"), Price: s.coin("98prune"),
				}))
				key := keeper.MakeIndexKeyAddressToOrder(s.addr4, 99)
				store.Set(key, []byte{keeper.OrderKeyTypeAsk})
				s.requireSetOrderInStore(store, exchange.NewOrder(100).WithAsk(&exchange.AskOrder{
					MarketId: 8, Seller: s.addr4.String(), Assets: s.coin("100apple"), Price: s.coin("100prune"),
				}))
			},
			req: &exchange.QueryGetOwnerOrdersRequest{Owner: s.addr4.String()},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders: []*exchange.Order{
					exchange.NewOrder(98).WithAsk(&exchange.AskOrder{
						MarketId: 8, Seller: s.addr4.String(), Assets: s.coin("98apple"), Price: s.coin("98prune"),
					}),
					exchange.NewOrder(100).WithAsk(&exchange.AskOrder{
						MarketId: 8, Seller: s.addr4.String(), Assets: s.coin("100apple"), Price: s.coin("100prune"),
					}),
				},
				Pagination: &query.PageResponse{Total: 3},
			},
		},
		{
			name: "error reading an order",
			setup: func() {
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(98).WithAsk(&exchange.AskOrder{
					MarketId: 8, Seller: s.addr5.String(), Assets: s.coin("98apple"), Price: s.coin("98prune"),
				}))
				key99, value99, err := s.k.GetOrderStoreKeyValue(*exchange.NewOrder(99).WithAsk(&exchange.AskOrder{
					MarketId: 8, Seller: s.addr5.String(), Assets: s.coin("99apple"), Price: s.coin("99prune"),
				}))
				s.Require().NoError(err, "GetOrderStoreKeyValue 99")
				value99[0] = 8
				store.Set(key99, value99)
				idxKey := keeper.MakeIndexKeyAddressToOrder(s.addr5, 99)
				store.Set(idxKey, []byte{8})
				s.requireSetOrderInStore(store, exchange.NewOrder(100).WithAsk(&exchange.AskOrder{
					MarketId: 8, Seller: s.addr5.String(), Assets: s.coin("100apple"), Price: s.coin("100prune"),
				}))
			},
			req: &exchange.QueryGetOwnerOrdersRequest{Owner: s.addr5.String()},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders: []*exchange.Order{
					exchange.NewOrder(98).WithAsk(&exchange.AskOrder{
						MarketId: 8, Seller: s.addr5.String(), Assets: s.coin("98apple"), Price: s.coin("98prune"),
					}),
					exchange.NewOrder(100).WithAsk(&exchange.AskOrder{
						MarketId: 8, Seller: s.addr5.String(), Assets: s.coin("100apple"), Price: s.coin("100prune"),
					}),
				},
				Pagination: &query.PageResponse{Total: 3},
			},
		},
		{
			name: "both offset and key provided",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner:      addr1,
				Pagination: &query.PageRequest{Offset: 2, Key: makeKey(ownerOrders[addr1][2])},
			},
			expInErr: []string{invalidArgErr, "error iterating orders for owner " + addr1,
				"invalid request, either offset or key is expected, got both"},
		},

		// Forward, no order type.
		{
			name: "forward, no order type, no after order, get all",
			req:  &exchange.QueryGetOwnerOrdersRequest{Owner: addr1},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     ownerOrders[addr1],
				Pagination: &query.PageResponse{Total: 20},
			},
		},
		{
			name: "forward, no order type, no after order, limit with key",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner:      addr2,
				Pagination: &query.PageRequest{Limit: 3, Key: makeKey(ownerOrders[addr2][2])},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     ownerOrders[addr2][2:5],
				Pagination: &query.PageResponse{NextKey: makeKey(ownerOrders[addr2][5])},
			},
		},
		{
			name: "forward, no order type, no after order, limit with offset, no count",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner:      addr3,
				Pagination: &query.PageRequest{Limit: 5, Offset: 8, CountTotal: false},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     ownerOrders[addr3][8:13],
				Pagination: &query.PageResponse{NextKey: makeKey(ownerOrders[addr3][13])},
			},
		},
		{
			name: "forward, no order type, no after order, limit with offset and count",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner:      addr1,
				Pagination: &query.PageRequest{Limit: 5, Offset: 6, CountTotal: true},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     ownerOrders[addr1][6:11],
				Pagination: &query.PageResponse{NextKey: makeKey(ownerOrders[addr1][11]), Total: 20},
			},
		},
		{
			name: "forward, no order type, after order 30, get all",
			req:  &exchange.QueryGetOwnerOrdersRequest{Owner: addr2, AfterOrderId: 30},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     ownerOrders[addr2][10:],
				Pagination: &query.PageResponse{Total: 10},
			},
		},
		{
			name: "forward, no order type, after order 30, limit with key",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr1, AfterOrderId: 30,
				Pagination: &query.PageRequest{Limit: 2, Key: makeKey(ownerOrders[addr1][15])},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     ownerOrders[addr1][15:17],
				Pagination: &query.PageResponse{NextKey: makeKey(ownerOrders[addr1][17])},
			},
		},
		{
			name: "forward, no order type, after order 30, limit with offset, no count",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr1, AfterOrderId: 30,
				Pagination: &query.PageRequest{Limit: 3, Offset: 2, CountTotal: false},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     ownerOrders[addr1][12:15],
				Pagination: &query.PageResponse{NextKey: makeKey(ownerOrders[addr1][15])},
			},
		},
		{
			name: "forward, no order type, after order 30, limit with offset and count",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr3, AfterOrderId: 30,
				Pagination: &query.PageRequest{Limit: 1, Offset: 7, CountTotal: true},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     ownerOrders[addr3][17:18],
				Pagination: &query.PageResponse{NextKey: makeKey(ownerOrders[addr3][18]), Total: 10},
			},
		},

		// Forward, only ask orders
		{
			name: "forward, ask orders, no after order, get all",
			req:  &exchange.QueryGetOwnerOrdersRequest{Owner: addr3, OrderType: "ask"},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     ownerAskOrders[addr3],
				Pagination: &query.PageResponse{Total: 10},
			},
		},
		{
			name: "forward, ask orders, no after order, limit with key",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr1, OrderType: "asks",
				Pagination: &query.PageRequest{Limit: 3, Key: makeKey(ownerAskOrders[addr1][4])},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     ownerAskOrders[addr1][4:7],
				Pagination: &query.PageResponse{NextKey: makeKey(ownerAskOrders[addr1][7])},
			},
		},
		{
			name: "forward, ask orders, no after order, limit with offset, no count",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr2, OrderType: "ASK",
				Pagination: &query.PageRequest{Limit: 3, Offset: 8, CountTotal: false},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     ownerAskOrders[addr2][8:],
				Pagination: &query.PageResponse{},
			},
		},
		{
			name: "forward, ask orders, no after order, limit with offset and count",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr2, OrderType: "ASKS",
				Pagination: &query.PageRequest{Limit: 3, Offset: 6, CountTotal: true},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     ownerAskOrders[addr2][6:9],
				Pagination: &query.PageResponse{NextKey: makeKey(ownerAskOrders[addr2][9]), Total: 10},
			},
		},
		{
			name: "forward, ask orders, after order 30, get all",
			req:  &exchange.QueryGetOwnerOrdersRequest{Owner: addr3, OrderType: "AskOrders", AfterOrderId: 30},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     ownerAskOrders[addr3][5:],
				Pagination: &query.PageResponse{Total: 5},
			},
		},
		{
			name: "forward, ask orders, after order 30, limit with key",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr2, OrderType: "ask orders", AfterOrderId: 30,
				Pagination: &query.PageRequest{Limit: 1, Key: makeKey(ownerAskOrders[addr2][7])},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     ownerAskOrders[addr2][7:8],
				Pagination: &query.PageResponse{NextKey: makeKey(ownerAskOrders[addr2][8])},
			},
		},
		{
			name: "forward, ask orders, after order 30, limit with offset, no count",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr1, OrderType: "askOrder", AfterOrderId: 30,
				Pagination: &query.PageRequest{Limit: 2, Offset: 2, CountTotal: false},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     ownerAskOrders[addr1][7:9],
				Pagination: &query.PageResponse{NextKey: makeKey(ownerAskOrders[addr1][9])},
			},
		},
		{
			name: "forward, ask orders, after order 30, limit with offset and count",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr1, OrderType: "aSKs", AfterOrderId: 30,
				Pagination: &query.PageRequest{Limit: 2, Offset: 1, CountTotal: true},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     ownerAskOrders[addr1][6:8],
				Pagination: &query.PageResponse{NextKey: makeKey(ownerAskOrders[addr1][8]), Total: 5},
			},
		},

		// Forward, only bid orders
		{
			name: "forward, bid orders, no after order, get all",
			req:  &exchange.QueryGetOwnerOrdersRequest{Owner: addr3, OrderType: "bid"},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     ownerBidOrders[addr3],
				Pagination: &query.PageResponse{Total: 10},
			},
		},
		{
			name: "forward, bid orders, no after order, limit with key",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr1, OrderType: "bids",
				Pagination: &query.PageRequest{Limit: 3, Key: makeKey(ownerBidOrders[addr1][4])},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     ownerBidOrders[addr1][4:7],
				Pagination: &query.PageResponse{NextKey: makeKey(ownerBidOrders[addr1][7])},
			},
		},
		{
			name: "forward, bid orders, no after order, limit with offset, no count",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr2, OrderType: "BID",
				Pagination: &query.PageRequest{Limit: 3, Offset: 8, CountTotal: false},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     ownerBidOrders[addr2][8:],
				Pagination: &query.PageResponse{},
			},
		},
		{
			name: "forward, bid orders, no after order, limit with offset and count",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr2, OrderType: "BIDS",
				Pagination: &query.PageRequest{Limit: 3, Offset: 6, CountTotal: true},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     ownerBidOrders[addr2][6:9],
				Pagination: &query.PageResponse{NextKey: makeKey(ownerBidOrders[addr2][9]), Total: 10},
			},
		},
		{
			name: "forward, bid orders, after order 30, get all",
			req:  &exchange.QueryGetOwnerOrdersRequest{Owner: addr3, OrderType: "BidOrders", AfterOrderId: 30},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     ownerBidOrders[addr3][5:],
				Pagination: &query.PageResponse{Total: 5},
			},
		},
		{
			name: "forward, bid orders, after order 30, limit with key",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr2, OrderType: "bid orders", AfterOrderId: 30,
				Pagination: &query.PageRequest{Limit: 1, Key: makeKey(ownerBidOrders[addr2][7])},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     ownerBidOrders[addr2][7:8],
				Pagination: &query.PageResponse{NextKey: makeKey(ownerBidOrders[addr2][8])},
			},
		},
		{
			name: "forward, bid orders, after order 30, limit with offset, no count",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr1, OrderType: "bidOrder", AfterOrderId: 30,
				Pagination: &query.PageRequest{Limit: 2, Offset: 2, CountTotal: false},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     ownerBidOrders[addr1][7:9],
				Pagination: &query.PageResponse{NextKey: makeKey(ownerBidOrders[addr1][9])},
			},
		},
		{
			name: "forward, bid orders, after order 30, limit with offset and count",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr1, OrderType: "bIDs", AfterOrderId: 30,
				Pagination: &query.PageRequest{Limit: 2, Offset: 1, CountTotal: true},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     ownerBidOrders[addr1][6:8],
				Pagination: &query.PageResponse{NextKey: makeKey(ownerBidOrders[addr1][8]), Total: 5},
			},
		},

		// Reverse, no order type.
		{
			name: "reverse, no order type, no after order, get all",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner:      addr1,
				Pagination: &query.PageRequest{Reverse: true},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     reverseSlice(ownerOrders[addr1]),
				Pagination: &query.PageResponse{Total: 20},
			},
		},
		{
			name: "reverse, no order type, no after order, limit with key",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner:      addr2,
				Pagination: &query.PageRequest{Reverse: true, Limit: 3, Key: makeKey(ownerOrders[addr2][12])},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     reverseSlice(ownerOrders[addr2][10:13]),
				Pagination: &query.PageResponse{NextKey: makeKey(ownerOrders[addr2][9])},
			},
		},
		{
			name: "reverse, no order type, no after order, limit with offset, no count",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner:      addr3,
				Pagination: &query.PageRequest{Reverse: true, Limit: 5, Offset: 8, CountTotal: false},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     reverseSlice(ownerOrders[addr3][7:12]),
				Pagination: &query.PageResponse{NextKey: makeKey(ownerOrders[addr3][6])},
			},
		},
		{
			name: "reverse, no order type, no after order, limit with offset and count",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner:      addr1,
				Pagination: &query.PageRequest{Reverse: true, Limit: 5, Offset: 6, CountTotal: true},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     reverseSlice(ownerOrders[addr1][9:14]),
				Pagination: &query.PageResponse{NextKey: makeKey(ownerOrders[addr1][8]), Total: 20},
			},
		},
		{
			name: "reverse, no order type, after order 30, get all",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr2, AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     reverseSlice(ownerOrders[addr2][10:]),
				Pagination: &query.PageResponse{Total: 10},
			},
		},
		{
			name: "reverse, no order type, after order 30, limit with key",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr1, AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true, Limit: 2, Key: makeKey(ownerOrders[addr1][15])},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     reverseSlice(ownerOrders[addr1][14:16]),
				Pagination: &query.PageResponse{NextKey: makeKey(ownerOrders[addr1][13])},
			},
		},
		{
			name: "reverse, no order type, after order 30, limit with offset, no count",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr1, AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true, Limit: 3, Offset: 2, CountTotal: false},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     reverseSlice(ownerOrders[addr1][15:18]),
				Pagination: &query.PageResponse{NextKey: makeKey(ownerOrders[addr1][14])},
			},
		},
		{
			name: "reverse, no order type, after order 30, limit with offset and count",
			// A key point of this test is that order 30 is in market 3. The AfterOrderID order
			// should NOT be included in results, though, so there should still only be 10 results here.
			// This validates that the "afterOrderID + 1" is correct in the getOrderIterator reverse block.
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr3, AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true, Limit: 1, Offset: 7, CountTotal: true},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     reverseSlice(ownerOrders[addr3][12:13]),
				Pagination: &query.PageResponse{NextKey: makeKey(ownerOrders[addr3][11]), Total: 10},
			},
		},

		// Reverse, only ask orders
		{
			name: "reverse, ask orders, no after order, get all",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr3, OrderType: "ask",
				Pagination: &query.PageRequest{Reverse: true},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     reverseSlice(ownerAskOrders[addr3]),
				Pagination: &query.PageResponse{Total: 10},
			},
		},
		{
			name: "reverse, ask orders, no after order, limit with key",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr1, OrderType: "asks",
				Pagination: &query.PageRequest{Reverse: true, Limit: 3, Key: makeKey(ownerAskOrders[addr1][4])},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     reverseSlice(ownerAskOrders[addr1][2:5]),
				Pagination: &query.PageResponse{NextKey: makeKey(ownerAskOrders[addr1][1])},
			},
		},
		{
			name: "reverse, ask orders, no after order, limit with offset, no count",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr2, OrderType: "ASK",
				Pagination: &query.PageRequest{Reverse: true, Limit: 3, Offset: 8, CountTotal: false},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     reverseSlice(ownerAskOrders[addr2][:2]),
				Pagination: &query.PageResponse{},
			},
		},
		{
			name: "reverse, ask orders, no after order, limit with offset and count",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr2, OrderType: "ASKS",
				Pagination: &query.PageRequest{Reverse: true, Limit: 3, Offset: 1, CountTotal: true},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     reverseSlice(ownerAskOrders[addr2][6:9]),
				Pagination: &query.PageResponse{NextKey: makeKey(ownerAskOrders[addr2][5]), Total: 10},
			},
		},
		{
			name: "reverse, ask orders, after order 30, get all",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr3, OrderType: "AskOrders", AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     reverseSlice(ownerAskOrders[addr3][5:]),
				Pagination: &query.PageResponse{Total: 5},
			},
		},
		{
			name: "reverse, ask orders, after order 30, limit with key",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr2, OrderType: "ask orders", AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true, Limit: 1, Key: makeKey(ownerAskOrders[addr2][7])},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     reverseSlice(ownerAskOrders[addr2][7:8]),
				Pagination: &query.PageResponse{NextKey: makeKey(ownerAskOrders[addr2][6])},
			},
		},
		{
			name: "reverse, ask orders, after order 30, limit with offset, no count",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr1, OrderType: "askOrder", AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true, Limit: 2, Offset: 2, CountTotal: false},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     reverseSlice(ownerAskOrders[addr1][6:8]),
				Pagination: &query.PageResponse{NextKey: makeKey(ownerAskOrders[addr1][5])},
			},
		},
		{
			name: "reverse, ask orders, after order 30, limit with offset and count",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr1, OrderType: "aSKs", AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true, Limit: 3, Offset: 1, CountTotal: true},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     reverseSlice(ownerAskOrders[addr1][6:9]),
				Pagination: &query.PageResponse{NextKey: makeKey(ownerAskOrders[addr1][5]), Total: 5},
			},
		},

		// Reverse, only bid orders
		{
			name: "reverse, bid orders, no after order, get all",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr3, OrderType: "bid",
				Pagination: &query.PageRequest{Reverse: true},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     reverseSlice(ownerBidOrders[addr3]),
				Pagination: &query.PageResponse{Total: 10},
			},
		},
		{
			name: "reverse, bid orders, no after order, limit with key",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr1, OrderType: "bids",
				Pagination: &query.PageRequest{Reverse: true, Limit: 3, Key: makeKey(ownerBidOrders[addr1][4])},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     reverseSlice(ownerBidOrders[addr1][2:5]),
				Pagination: &query.PageResponse{NextKey: makeKey(ownerBidOrders[addr1][1])},
			},
		},
		{
			name: "reverse, bid orders, no after order, limit with offset, no count",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr2, OrderType: "BID",
				Pagination: &query.PageRequest{Reverse: true, Limit: 3, Offset: 8, CountTotal: false},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     reverseSlice(ownerBidOrders[addr2][:2]),
				Pagination: &query.PageResponse{},
			},
		},
		{
			name: "reverse, bid orders, no after order, limit with offset and count",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr2, OrderType: "BIDS",
				Pagination: &query.PageRequest{Reverse: true, Limit: 3, Offset: 1, CountTotal: true},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     reverseSlice(ownerBidOrders[addr2][6:9]),
				Pagination: &query.PageResponse{NextKey: makeKey(ownerBidOrders[addr2][5]), Total: 10},
			},
		},
		{
			name: "reverse, bid orders, after order 30, get all",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr3, OrderType: "BidOrders", AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     reverseSlice(ownerBidOrders[addr3][5:]),
				Pagination: &query.PageResponse{Total: 5},
			},
		},
		{
			name: "reverse, bid orders, after order 30, limit with key",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr2, OrderType: "bid orders", AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true, Limit: 1, Key: makeKey(ownerBidOrders[addr2][7])},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     reverseSlice(ownerBidOrders[addr2][7:8]),
				Pagination: &query.PageResponse{NextKey: makeKey(ownerBidOrders[addr2][6])},
			},
		},
		{
			name: "reverse, bid orders, after order 30, limit with offset, no count",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr1, OrderType: "bidOrder", AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true, Limit: 2, Offset: 2, CountTotal: false},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     reverseSlice(ownerBidOrders[addr1][6:8]),
				Pagination: &query.PageResponse{NextKey: makeKey(ownerBidOrders[addr1][5])},
			},
		},
		{
			name: "reverse, bid orders, after order 30, limit with offset and count",
			req: &exchange.QueryGetOwnerOrdersRequest{
				Owner: addr1, OrderType: "bIDs", AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true, Limit: 3, Offset: 1, CountTotal: true},
			},
			expResp: &exchange.QueryGetOwnerOrdersResponse{
				Orders:     reverseSlice(ownerBidOrders[addr1][6:9]),
				Pagination: &query.PageResponse{NextKey: makeKey(ownerBidOrders[addr1][5]), Total: 5},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			runQueryTestCase(s, testDef, tc)
		})
	}
}

func (s *TestSuite) TestQueryServer_GetAssetOrders() {
	testDef := queryTestDef[exchange.QueryGetAssetOrdersRequest, exchange.QueryGetAssetOrdersResponse]{
		queryName: "GetAssetOrders",
		query:     keeper.NewQueryServer(s.k).GetAssetOrders,
		followup: func(expected, actual *exchange.QueryGetAssetOrdersResponse) {
			s.assertEqualOrders(expected.Orders, actual.Orders, "Orders")
			s.assertEqualPageResponse(expected.Pagination, actual.Pagination, "Pagination")
		},
	}
	makeKey := func(order *exchange.Order) []byte {
		return keeper.Uint64Bz(order.OrderId)
	}

	denom1, denom2, denom3 := "one", "two", "three"
	denoms := []string{denom1, denom2, denom3}
	denomCount := len(denoms)
	ordersPerDenom := 20
	denomOrders := make(map[string][]*exchange.Order, denomCount)
	denomAskOrders := make(map[string][]*exchange.Order, denomCount)
	denomBidOrders := make(map[string][]*exchange.Order, denomCount)
	for _, denom := range denoms {
		denomOrders[denom] = make([]*exchange.Order, 0, ordersPerDenom)
		denomAskOrders[denom] = make([]*exchange.Order, 0, ordersPerDenom/2)
		denomBidOrders[denom] = make([]*exchange.Order, 0, ordersPerDenom/2)
	}
	mainStore := s.getStore()
	for i := 1; i <= denomCount*ordersPerDenom; i++ {
		orderID := uint64(i)
		denom := denoms[i%denomCount]
		order := exchange.NewOrder(orderID)
		if orderID%2 == 0 {
			order.WithAsk(&exchange.AskOrder{
				MarketId:     uint32(5000 + i),
				Seller:       sdk.AccAddress(fmt.Sprintf("seller_%d____________", orderID)[:20]).String(),
				Assets:       sdk.NewInt64Coin(denom, int64(i)),
				Price:        sdk.NewInt64Coin("plum", int64(i)),
				AllowPartial: orderID%4 < 2,
				ExternalId:   fmt.Sprintf("external-id-%d", i),
			})
			denomAskOrders[denom] = append(denomAskOrders[denom], order)
		} else {
			order.WithBid(&exchange.BidOrder{
				MarketId:     uint32(5000 + i),
				Buyer:        sdk.AccAddress(fmt.Sprintf("buyer_%d_____________", orderID)[:20]).String(),
				Assets:       sdk.NewInt64Coin(denom, int64(i)),
				Price:        sdk.NewInt64Coin("plum", int64(i)),
				AllowPartial: orderID%4 < 2,
				ExternalId:   fmt.Sprintf("external-id-%d", i),
			})
			denomBidOrders[denom] = append(denomBidOrders[denom], order)
		}
		denomOrders[denom] = append(denomOrders[denom], order)
		s.requireSetOrderInStore(mainStore, order)
	}

	// OrderIDs for each denom:
	//        0  1  2   3   4   5   6   7   8   9  10  11  12  13  14  15  16  17  18  19
	//denom1: 1, 4, 7, 10, 13, 16, 19, 22, 25, 28, 31, 34, 37, 40, 43, 46, 49, 52, 55, 58
	//denom2: 2, 5, 8, 11, 14, 17, 20, 23, 26, 29, 32, 35, 38, 41, 44, 47, 50, 53, 56, 59
	//denom3: 3, 6, 9, 12, 15, 18, 21, 24, 27, 30, 33, 36, 39, 42, 45, 48, 51, 54, 57, 60

	tests := []queryTestCase[exchange.QueryGetAssetOrdersRequest, exchange.QueryGetAssetOrdersResponse]{
		// Tests on errors and non-normal conditions.
		{
			name:     "nil req",
			req:      nil,
			expInErr: []string{invalidArgErr, "empty request"},
		},
		{
			name:     "empty asset",
			req:      &exchange.QueryGetAssetOrdersRequest{Asset: ""},
			expInErr: []string{invalidArgErr, "empty request"},
		},
		{
			name:     "unknown order type",
			req:      &exchange.QueryGetAssetOrdersRequest{Asset: denom1, OrderType: "burger and fries"},
			expInErr: []string{invalidArgErr, "error iterating orders for asset " + denom1 + ": unknown order type \"burger and fries\""},
		},
		{
			name:    "no orders",
			req:     &exchange.QueryGetAssetOrdersRequest{Asset: "four"},
			expResp: &exchange.QueryGetAssetOrdersResponse{Orders: nil, Pagination: &query.PageResponse{}},
		},
		{
			name: "bad index entry",
			setup: func() {
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(98).WithAsk(&exchange.AskOrder{
					MarketId: 7, Seller: s.addr1.String(), Assets: s.coin("98apple"), Price: s.coin("98prune"),
				}))
				key99, value99, err := s.k.GetOrderStoreKeyValue(*exchange.NewOrder(99).WithAsk(&exchange.AskOrder{
					MarketId: 8, Seller: s.addr2.String(), Assets: s.coin("99apple"), Price: s.coin("99prune"),
				}))
				s.Require().NoError(err, "GetOrderStoreKeyValue 99")
				store.Set(key99, value99)
				store.Set(s.badKey(keeper.MakeIndexKeyMarketToOrder(8, 99)), []byte{keeper.OrderKeyTypeAsk})
				s.requireSetOrderInStore(store, exchange.NewOrder(100).WithAsk(&exchange.AskOrder{
					MarketId: 9, Seller: s.addr3.String(), Assets: s.coin("100apple"), Price: s.coin("100prune"),
				}))
			},
			req: &exchange.QueryGetAssetOrdersRequest{Asset: "apple"},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders: []*exchange.Order{
					exchange.NewOrder(98).WithAsk(&exchange.AskOrder{
						MarketId: 7, Seller: s.addr1.String(), Assets: s.coin("98apple"), Price: s.coin("98prune"),
					}),
					exchange.NewOrder(100).WithAsk(&exchange.AskOrder{
						MarketId: 9, Seller: s.addr3.String(), Assets: s.coin("100apple"), Price: s.coin("100prune"),
					}),
				},
				Pagination: &query.PageResponse{Total: 2},
			},
		},
		{
			name: "index entry to order that does not exist",
			setup: func() {
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(98).WithAsk(&exchange.AskOrder{
					MarketId: 7, Seller: s.addr1.String(), Assets: s.coin("98acorn"), Price: s.coin("98prune"),
				}))
				key := keeper.MakeIndexKeyAssetToOrder("acorn", 99)
				store.Set(key, []byte{keeper.OrderKeyTypeAsk})
				s.requireSetOrderInStore(store, exchange.NewOrder(100).WithAsk(&exchange.AskOrder{
					MarketId: 9, Seller: s.addr3.String(), Assets: s.coin("100acorn"), Price: s.coin("100prune"),
				}))
			},
			req: &exchange.QueryGetAssetOrdersRequest{Asset: "acorn"},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders: []*exchange.Order{
					exchange.NewOrder(98).WithAsk(&exchange.AskOrder{
						MarketId: 7, Seller: s.addr1.String(), Assets: s.coin("98acorn"), Price: s.coin("98prune"),
					}),
					exchange.NewOrder(100).WithAsk(&exchange.AskOrder{
						MarketId: 9, Seller: s.addr3.String(), Assets: s.coin("100acorn"), Price: s.coin("100prune"),
					}),
				},
				Pagination: &query.PageResponse{Total: 3},
			},
		},
		{
			name: "error reading an order",
			setup: func() {
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(98).WithAsk(&exchange.AskOrder{
					MarketId: 7, Seller: s.addr1.String(), Assets: s.coin("98acorn"), Price: s.coin("98prune"),
				}))
				key99, value99, err := s.k.GetOrderStoreKeyValue(*exchange.NewOrder(99).WithAsk(&exchange.AskOrder{
					MarketId: 8, Seller: s.addr2.String(), Assets: s.coin("99acorn"), Price: s.coin("99prune"),
				}))
				s.Require().NoError(err, "GetOrderStoreKeyValue 99")
				value99[0] = 8
				store.Set(key99, value99)
				idxKey := keeper.MakeIndexKeyAssetToOrder("acorn", 99)
				store.Set(idxKey, []byte{8})
				s.requireSetOrderInStore(store, exchange.NewOrder(100).WithAsk(&exchange.AskOrder{
					MarketId: 9, Seller: s.addr3.String(), Assets: s.coin("100acorn"), Price: s.coin("100prune"),
				}))
			},
			req: &exchange.QueryGetAssetOrdersRequest{Asset: "acorn"},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders: []*exchange.Order{
					exchange.NewOrder(98).WithAsk(&exchange.AskOrder{
						MarketId: 7, Seller: s.addr1.String(), Assets: s.coin("98acorn"), Price: s.coin("98prune"),
					}),
					exchange.NewOrder(100).WithAsk(&exchange.AskOrder{
						MarketId: 9, Seller: s.addr3.String(), Assets: s.coin("100acorn"), Price: s.coin("100prune"),
					}),
				},
				Pagination: &query.PageResponse{Total: 3},
			},
		},
		{
			name: "both offset and key provided",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset:      denom1,
				Pagination: &query.PageRequest{Offset: 2, Key: makeKey(denomOrders[denom1][2])},
			},
			expInErr: []string{invalidArgErr, "error iterating orders for asset " + denom1,
				"invalid request, either offset or key is expected, got both"},
		},

		// Forward, no order type.
		{
			name: "forward, no order type, no after order, get all",
			req:  &exchange.QueryGetAssetOrdersRequest{Asset: denom1},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     denomOrders[denom1],
				Pagination: &query.PageResponse{Total: 20},
			},
		},
		{
			name: "forward, no order type, no after order, limit with key",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset:      denom2,
				Pagination: &query.PageRequest{Limit: 3, Key: makeKey(denomOrders[denom2][2])},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     denomOrders[denom2][2:5],
				Pagination: &query.PageResponse{NextKey: makeKey(denomOrders[denom2][5])},
			},
		},
		{
			name: "forward, no order type, no after order, limit with offset, no count",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset:      denom3,
				Pagination: &query.PageRequest{Limit: 5, Offset: 8, CountTotal: false},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     denomOrders[denom3][8:13],
				Pagination: &query.PageResponse{NextKey: makeKey(denomOrders[denom3][13])},
			},
		},
		{
			name: "forward, no order type, no after order, limit with offset and count",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset:      denom1,
				Pagination: &query.PageRequest{Limit: 5, Offset: 6, CountTotal: true},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     denomOrders[denom1][6:11],
				Pagination: &query.PageResponse{NextKey: makeKey(denomOrders[denom1][11]), Total: 20},
			},
		},
		{
			name: "forward, no order type, after order 30, get all",
			req:  &exchange.QueryGetAssetOrdersRequest{Asset: denom2, AfterOrderId: 30},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     denomOrders[denom2][10:],
				Pagination: &query.PageResponse{Total: 10},
			},
		},
		{
			name: "forward, no order type, after order 30, limit with key",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom1, AfterOrderId: 30,
				Pagination: &query.PageRequest{Limit: 2, Key: makeKey(denomOrders[denom1][15])},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     denomOrders[denom1][15:17],
				Pagination: &query.PageResponse{NextKey: makeKey(denomOrders[denom1][17])},
			},
		},
		{
			name: "forward, no order type, after order 30, limit with offset, no count",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom1, AfterOrderId: 30,
				Pagination: &query.PageRequest{Limit: 3, Offset: 2, CountTotal: false},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     denomOrders[denom1][12:15],
				Pagination: &query.PageResponse{NextKey: makeKey(denomOrders[denom1][15])},
			},
		},
		{
			name: "forward, no order type, after order 30, limit with offset and count",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom3, AfterOrderId: 30,
				Pagination: &query.PageRequest{Limit: 1, Offset: 7, CountTotal: true},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     denomOrders[denom3][17:18],
				Pagination: &query.PageResponse{NextKey: makeKey(denomOrders[denom3][18]), Total: 10},
			},
		},

		// Forward, only ask orders
		{
			name: "forward, ask orders, no after order, get all",
			req:  &exchange.QueryGetAssetOrdersRequest{Asset: denom3, OrderType: "ask"},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     denomAskOrders[denom3],
				Pagination: &query.PageResponse{Total: 10},
			},
		},
		{
			name: "forward, ask orders, no after order, limit with key",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom1, OrderType: "asks",
				Pagination: &query.PageRequest{Limit: 3, Key: makeKey(denomAskOrders[denom1][4])},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     denomAskOrders[denom1][4:7],
				Pagination: &query.PageResponse{NextKey: makeKey(denomAskOrders[denom1][7])},
			},
		},
		{
			name: "forward, ask orders, no after order, limit with offset, no count",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom2, OrderType: "ASK",
				Pagination: &query.PageRequest{Limit: 3, Offset: 8, CountTotal: false},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     denomAskOrders[denom2][8:],
				Pagination: &query.PageResponse{},
			},
		},
		{
			name: "forward, ask orders, no after order, limit with offset and count",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom2, OrderType: "ASKS",
				Pagination: &query.PageRequest{Limit: 3, Offset: 6, CountTotal: true},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     denomAskOrders[denom2][6:9],
				Pagination: &query.PageResponse{NextKey: makeKey(denomAskOrders[denom2][9]), Total: 10},
			},
		},
		{
			name: "forward, ask orders, after order 30, get all",
			req:  &exchange.QueryGetAssetOrdersRequest{Asset: denom3, OrderType: "AskOrders", AfterOrderId: 30},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     denomAskOrders[denom3][5:],
				Pagination: &query.PageResponse{Total: 5},
			},
		},
		{
			name: "forward, ask orders, after order 30, limit with key",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom2, OrderType: "ask orders", AfterOrderId: 30,
				Pagination: &query.PageRequest{Limit: 1, Key: makeKey(denomAskOrders[denom2][7])},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     denomAskOrders[denom2][7:8],
				Pagination: &query.PageResponse{NextKey: makeKey(denomAskOrders[denom2][8])},
			},
		},
		{
			name: "forward, ask orders, after order 30, limit with offset, no count",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom1, OrderType: "askOrder", AfterOrderId: 30,
				Pagination: &query.PageRequest{Limit: 2, Offset: 2, CountTotal: false},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     denomAskOrders[denom1][7:9],
				Pagination: &query.PageResponse{NextKey: makeKey(denomAskOrders[denom1][9])},
			},
		},
		{
			name: "forward, ask orders, after order 30, limit with offset and count",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom1, OrderType: "aSKs", AfterOrderId: 30,
				Pagination: &query.PageRequest{Limit: 2, Offset: 1, CountTotal: true},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     denomAskOrders[denom1][6:8],
				Pagination: &query.PageResponse{NextKey: makeKey(denomAskOrders[denom1][8]), Total: 5},
			},
		},

		// Forward, only bid orders
		{
			name: "forward, bid orders, no after order, get all",
			req:  &exchange.QueryGetAssetOrdersRequest{Asset: denom3, OrderType: "bid"},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     denomBidOrders[denom3],
				Pagination: &query.PageResponse{Total: 10},
			},
		},
		{
			name: "forward, bid orders, no after order, limit with key",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom1, OrderType: "bids",
				Pagination: &query.PageRequest{Limit: 3, Key: makeKey(denomBidOrders[denom1][4])},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     denomBidOrders[denom1][4:7],
				Pagination: &query.PageResponse{NextKey: makeKey(denomBidOrders[denom1][7])},
			},
		},
		{
			name: "forward, bid orders, no after order, limit with offset, no count",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom2, OrderType: "BID",
				Pagination: &query.PageRequest{Limit: 3, Offset: 8, CountTotal: false},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     denomBidOrders[denom2][8:],
				Pagination: &query.PageResponse{},
			},
		},
		{
			name: "forward, bid orders, no after order, limit with offset and count",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom2, OrderType: "BIDS",
				Pagination: &query.PageRequest{Limit: 3, Offset: 6, CountTotal: true},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     denomBidOrders[denom2][6:9],
				Pagination: &query.PageResponse{NextKey: makeKey(denomBidOrders[denom2][9]), Total: 10},
			},
		},
		{
			name: "forward, bid orders, after order 30, get all",
			req:  &exchange.QueryGetAssetOrdersRequest{Asset: denom3, OrderType: "BidOrders", AfterOrderId: 30},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     denomBidOrders[denom3][5:],
				Pagination: &query.PageResponse{Total: 5},
			},
		},
		{
			name: "forward, bid orders, after order 30, limit with key",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom2, OrderType: "bid orders", AfterOrderId: 30,
				Pagination: &query.PageRequest{Limit: 1, Key: makeKey(denomBidOrders[denom2][7])},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     denomBidOrders[denom2][7:8],
				Pagination: &query.PageResponse{NextKey: makeKey(denomBidOrders[denom2][8])},
			},
		},
		{
			name: "forward, bid orders, after order 30, limit with offset, no count",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom1, OrderType: "bidOrder", AfterOrderId: 30,
				Pagination: &query.PageRequest{Limit: 2, Offset: 2, CountTotal: false},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     denomBidOrders[denom1][7:9],
				Pagination: &query.PageResponse{NextKey: makeKey(denomBidOrders[denom1][9])},
			},
		},
		{
			name: "forward, bid orders, after order 30, limit with offset and count",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom1, OrderType: "bIDs", AfterOrderId: 30,
				Pagination: &query.PageRequest{Limit: 2, Offset: 1, CountTotal: true},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     denomBidOrders[denom1][6:8],
				Pagination: &query.PageResponse{NextKey: makeKey(denomBidOrders[denom1][8]), Total: 5},
			},
		},

		// Reverse, no order type.
		{
			name: "reverse, no order type, no after order, get all",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset:      denom1,
				Pagination: &query.PageRequest{Reverse: true},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     reverseSlice(denomOrders[denom1]),
				Pagination: &query.PageResponse{Total: 20},
			},
		},
		{
			name: "reverse, no order type, no after order, limit with key",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset:      denom2,
				Pagination: &query.PageRequest{Reverse: true, Limit: 3, Key: makeKey(denomOrders[denom2][12])},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     reverseSlice(denomOrders[denom2][10:13]),
				Pagination: &query.PageResponse{NextKey: makeKey(denomOrders[denom2][9])},
			},
		},
		{
			name: "reverse, no order type, no after order, limit with offset, no count",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset:      denom3,
				Pagination: &query.PageRequest{Reverse: true, Limit: 5, Offset: 8, CountTotal: false},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     reverseSlice(denomOrders[denom3][7:12]),
				Pagination: &query.PageResponse{NextKey: makeKey(denomOrders[denom3][6])},
			},
		},
		{
			name: "reverse, no order type, no after order, limit with offset and count",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset:      denom1,
				Pagination: &query.PageRequest{Reverse: true, Limit: 5, Offset: 6, CountTotal: true},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     reverseSlice(denomOrders[denom1][9:14]),
				Pagination: &query.PageResponse{NextKey: makeKey(denomOrders[denom1][8]), Total: 20},
			},
		},
		{
			name: "reverse, no order type, after order 30, get all",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom2, AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     reverseSlice(denomOrders[denom2][10:]),
				Pagination: &query.PageResponse{Total: 10},
			},
		},
		{
			name: "reverse, no order type, after order 30, limit with key",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom1, AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true, Limit: 2, Key: makeKey(denomOrders[denom1][15])},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     reverseSlice(denomOrders[denom1][14:16]),
				Pagination: &query.PageResponse{NextKey: makeKey(denomOrders[denom1][13])},
			},
		},
		{
			name: "reverse, no order type, after order 30, limit with offset, no count",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom1, AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true, Limit: 3, Offset: 2, CountTotal: false},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     reverseSlice(denomOrders[denom1][15:18]),
				Pagination: &query.PageResponse{NextKey: makeKey(denomOrders[denom1][14])},
			},
		},
		{
			name: "reverse, no order type, after order 30, limit with offset and count",
			// A key point of this test is that order 30 is in market 3. The AfterOrderID order
			// should NOT be included in results, though, so there should still only be 10 results here.
			// This validates that the "afterOrderID + 1" is correct in the getOrderIterator reverse block.
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom3, AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true, Limit: 1, Offset: 7, CountTotal: true},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     reverseSlice(denomOrders[denom3][12:13]),
				Pagination: &query.PageResponse{NextKey: makeKey(denomOrders[denom3][11]), Total: 10},
			},
		},

		// Reverse, only ask orders
		{
			name: "reverse, ask orders, no after order, get all",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom3, OrderType: "ask",
				Pagination: &query.PageRequest{Reverse: true},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     reverseSlice(denomAskOrders[denom3]),
				Pagination: &query.PageResponse{Total: 10},
			},
		},
		{
			name: "reverse, ask orders, no after order, limit with key",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom1, OrderType: "asks",
				Pagination: &query.PageRequest{Reverse: true, Limit: 3, Key: makeKey(denomAskOrders[denom1][4])},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     reverseSlice(denomAskOrders[denom1][2:5]),
				Pagination: &query.PageResponse{NextKey: makeKey(denomAskOrders[denom1][1])},
			},
		},
		{
			name: "reverse, ask orders, no after order, limit with offset, no count",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom2, OrderType: "ASK",
				Pagination: &query.PageRequest{Reverse: true, Limit: 3, Offset: 8, CountTotal: false},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     reverseSlice(denomAskOrders[denom2][:2]),
				Pagination: &query.PageResponse{},
			},
		},
		{
			name: "reverse, ask orders, no after order, limit with offset and count",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom2, OrderType: "ASKS",
				Pagination: &query.PageRequest{Reverse: true, Limit: 3, Offset: 1, CountTotal: true},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     reverseSlice(denomAskOrders[denom2][6:9]),
				Pagination: &query.PageResponse{NextKey: makeKey(denomAskOrders[denom2][5]), Total: 10},
			},
		},
		{
			name: "reverse, ask orders, after order 30, get all",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom3, OrderType: "AskOrders", AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     reverseSlice(denomAskOrders[denom3][5:]),
				Pagination: &query.PageResponse{Total: 5},
			},
		},
		{
			name: "reverse, ask orders, after order 30, limit with key",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom2, OrderType: "ask orders", AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true, Limit: 1, Key: makeKey(denomAskOrders[denom2][7])},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     reverseSlice(denomAskOrders[denom2][7:8]),
				Pagination: &query.PageResponse{NextKey: makeKey(denomAskOrders[denom2][6])},
			},
		},
		{
			name: "reverse, ask orders, after order 30, limit with offset, no count",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom1, OrderType: "askOrder", AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true, Limit: 2, Offset: 2, CountTotal: false},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     reverseSlice(denomAskOrders[denom1][6:8]),
				Pagination: &query.PageResponse{NextKey: makeKey(denomAskOrders[denom1][5])},
			},
		},
		{
			name: "reverse, ask orders, after order 30, limit with offset and count",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom1, OrderType: "aSKs", AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true, Limit: 3, Offset: 1, CountTotal: true},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     reverseSlice(denomAskOrders[denom1][6:9]),
				Pagination: &query.PageResponse{NextKey: makeKey(denomAskOrders[denom1][5]), Total: 5},
			},
		},

		// Reverse, only bid orders
		{
			name: "reverse, bid orders, no after order, get all",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom3, OrderType: "bid",
				Pagination: &query.PageRequest{Reverse: true},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     reverseSlice(denomBidOrders[denom3]),
				Pagination: &query.PageResponse{Total: 10},
			},
		},
		{
			name: "reverse, bid orders, no after order, limit with key",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom1, OrderType: "bids",
				Pagination: &query.PageRequest{Reverse: true, Limit: 3, Key: makeKey(denomBidOrders[denom1][4])},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     reverseSlice(denomBidOrders[denom1][2:5]),
				Pagination: &query.PageResponse{NextKey: makeKey(denomBidOrders[denom1][1])},
			},
		},
		{
			name: "reverse, bid orders, no after order, limit with offset, no count",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom2, OrderType: "BID",
				Pagination: &query.PageRequest{Reverse: true, Limit: 3, Offset: 8, CountTotal: false},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     reverseSlice(denomBidOrders[denom2][:2]),
				Pagination: &query.PageResponse{},
			},
		},
		{
			name: "reverse, bid orders, no after order, limit with offset and count",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom2, OrderType: "BIDS",
				Pagination: &query.PageRequest{Reverse: true, Limit: 3, Offset: 1, CountTotal: true},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     reverseSlice(denomBidOrders[denom2][6:9]),
				Pagination: &query.PageResponse{NextKey: makeKey(denomBidOrders[denom2][5]), Total: 10},
			},
		},
		{
			name: "reverse, bid orders, after order 30, get all",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom3, OrderType: "BidOrders", AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     reverseSlice(denomBidOrders[denom3][5:]),
				Pagination: &query.PageResponse{Total: 5},
			},
		},
		{
			name: "reverse, bid orders, after order 30, limit with key",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom2, OrderType: "bid orders", AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true, Limit: 1, Key: makeKey(denomBidOrders[denom2][7])},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     reverseSlice(denomBidOrders[denom2][7:8]),
				Pagination: &query.PageResponse{NextKey: makeKey(denomBidOrders[denom2][6])},
			},
		},
		{
			name: "reverse, bid orders, after order 30, limit with offset, no count",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom1, OrderType: "bidOrder", AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true, Limit: 2, Offset: 2, CountTotal: false},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     reverseSlice(denomBidOrders[denom1][6:8]),
				Pagination: &query.PageResponse{NextKey: makeKey(denomBidOrders[denom1][5])},
			},
		},
		{
			name: "reverse, bid orders, after order 30, limit with offset and count",
			req: &exchange.QueryGetAssetOrdersRequest{
				Asset: denom1, OrderType: "bIDs", AfterOrderId: 30,
				Pagination: &query.PageRequest{Reverse: true, Limit: 3, Offset: 1, CountTotal: true},
			},
			expResp: &exchange.QueryGetAssetOrdersResponse{
				Orders:     reverseSlice(denomBidOrders[denom1][6:9]),
				Pagination: &query.PageResponse{NextKey: makeKey(denomBidOrders[denom1][5]), Total: 5},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			runQueryTestCase(s, testDef, tc)
		})
	}
}

func (s *TestSuite) TestQueryServer_GetAllOrders() {
	testDef := queryTestDef[exchange.QueryGetAllOrdersRequest, exchange.QueryGetAllOrdersResponse]{
		queryName: "GetAllOrders",
		query:     keeper.NewQueryServer(s.k).GetAllOrders,
		followup: func(expected, actual *exchange.QueryGetAllOrdersResponse) {
			s.assertEqualOrders(expected.Orders, actual.Orders, "Orders")
			s.assertEqualPageResponse(expected.Pagination, actual.Pagination, "Pagination")
		},
	}
	makeKey := func(order *exchange.Order) []byte {
		return keeper.Uint64Bz(order.OrderId)
	}

	fiveOrders := []*exchange.Order{
		exchange.NewOrder(14).WithAsk(&exchange.AskOrder{
			MarketId: 8, Seller: s.addr1.String(), Assets: s.coin("14apple"), Price: s.coin("14prune"),
			SellerSettlementFlatFee: s.coinP("14fig"), AllowPartial: false, ExternalId: "external-id-5",
		}),
		exchange.NewOrder(38).WithBid(&exchange.BidOrder{
			MarketId: 6, Buyer: s.addr1.String(), Assets: s.coin("38apple"), Price: s.coin("38prune"),
			BuyerSettlementFees: s.coins("38fig"), AllowPartial: true, ExternalId: "external-id-4",
		}),
		exchange.NewOrder(39).WithBid(&exchange.BidOrder{
			MarketId: 5, Buyer: s.addr1.String(), Assets: s.coin("39apple"), Price: s.coin("39prune"),
			BuyerSettlementFees: s.coins("39fig"), AllowPartial: false, ExternalId: "external-id-1",
		}),
		exchange.NewOrder(71).WithAsk(&exchange.AskOrder{
			MarketId: 5, Seller: s.addr3.String(), Assets: s.coin("71apple"), Price: s.coin("71prune"),
			SellerSettlementFlatFee: s.coinP("71fig"), AllowPartial: true, ExternalId: "external-id-3",
		}),
		exchange.NewOrder(73).WithBid(&exchange.BidOrder{
			MarketId: 5, Buyer: s.addr2.String(), Assets: s.coin("73apple"), Price: s.coin("73prune"),
			BuyerSettlementFees: s.coins("73fig"), AllowPartial: false, ExternalId: "external-id-2",
		}),
	}
	fiveOrderSetup := func() {
		store := s.getStore()
		s.requireSetOrderInStore(store, fiveOrders[2])
		s.requireSetOrderInStore(store, fiveOrders[4])
		s.requireSetOrderInStore(store, fiveOrders[3])
		s.requireSetOrderInStore(store, fiveOrders[1])
		s.requireSetOrderInStore(store, fiveOrders[0])
	}

	tests := []queryTestCase[exchange.QueryGetAllOrdersRequest, exchange.QueryGetAllOrdersResponse]{
		{
			name: "bad key in store",
			setup: func() {
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(1).WithBid(&exchange.BidOrder{
					MarketId: 1, Buyer: s.addr1.String(), Assets: s.coin("1apple"), Price: s.coin("1prune"),
					BuyerSettlementFees: s.coins("1fig"), AllowPartial: false, ExternalId: "external-id-1",
				}))

				key2, value2, err := s.k.GetOrderStoreKeyValue(*exchange.NewOrder(2).WithBid(&exchange.BidOrder{
					MarketId: 2, Buyer: s.addr2.String(), Assets: s.coin("2apple"), Price: s.coin("2prune"),
					BuyerSettlementFees: s.coins("2fig"), AllowPartial: false, ExternalId: "external-id-2",
				}))
				s.Require().NoError(err, "GetOrderStoreKeyValue 2")
				store.Set(s.badKey(key2), value2)

				s.requireSetOrderInStore(store, exchange.NewOrder(3).WithAsk(&exchange.AskOrder{
					MarketId: 3, Seller: s.addr3.String(), Assets: s.coin("3apple"), Price: s.coin("3prune"),
					SellerSettlementFlatFee: s.coinP("3fig"), AllowPartial: false, ExternalId: "external-id-3",
				}))
			},
			expResp: &exchange.QueryGetAllOrdersResponse{
				Orders: []*exchange.Order{
					exchange.NewOrder(1).WithBid(&exchange.BidOrder{
						MarketId: 1, Buyer: s.addr1.String(), Assets: s.coin("1apple"), Price: s.coin("1prune"),
						BuyerSettlementFees: s.coins("1fig"), AllowPartial: false, ExternalId: "external-id-1",
					}),
					exchange.NewOrder(3).WithAsk(&exchange.AskOrder{
						MarketId: 3, Seller: s.addr3.String(), Assets: s.coin("3apple"), Price: s.coin("3prune"),
						SellerSettlementFlatFee: s.coinP("3fig"), AllowPartial: false, ExternalId: "external-id-3",
					}),
				},
				Pagination: &query.PageResponse{Total: 2},
			},
		},
		{
			name: "bad order in store",
			setup: func() {
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(1).WithBid(&exchange.BidOrder{
					MarketId: 1, Buyer: s.addr1.String(), Assets: s.coin("1apple"), Price: s.coin("1prune"),
					BuyerSettlementFees: s.coins("1fig"), AllowPartial: false, ExternalId: "external-id-1",
				}))

				key2, value2, err := s.k.GetOrderStoreKeyValue(*exchange.NewOrder(2).WithBid(&exchange.BidOrder{
					MarketId: 2, Buyer: s.addr2.String(), Assets: s.coin("2apple"), Price: s.coin("2prune"),
					BuyerSettlementFees: s.coins("2fig"), AllowPartial: false, ExternalId: "external-id-2",
				}))
				s.Require().NoError(err, "GetOrderStoreKeyValue 2")
				value2[0] = 9
				store.Set(key2, value2)

				s.requireSetOrderInStore(store, exchange.NewOrder(3).WithAsk(&exchange.AskOrder{
					MarketId: 3, Seller: s.addr3.String(), Assets: s.coin("3apple"), Price: s.coin("3prune"),
					SellerSettlementFlatFee: s.coinP("3fig"), AllowPartial: false, ExternalId: "external-id-3",
				}))
			},
			expResp: &exchange.QueryGetAllOrdersResponse{
				Orders: []*exchange.Order{
					exchange.NewOrder(1).WithBid(&exchange.BidOrder{
						MarketId: 1, Buyer: s.addr1.String(), Assets: s.coin("1apple"), Price: s.coin("1prune"),
						BuyerSettlementFees: s.coins("1fig"), AllowPartial: false, ExternalId: "external-id-1",
					}),
					exchange.NewOrder(3).WithAsk(&exchange.AskOrder{
						MarketId: 3, Seller: s.addr3.String(), Assets: s.coin("3apple"), Price: s.coin("3prune"),
						SellerSettlementFlatFee: s.coinP("3fig"), AllowPartial: false, ExternalId: "external-id-3",
					}),
				},
				Pagination: &query.PageResponse{Total: 3},
			},
		},
		{
			name: "both offset and key provided",
			req:  &exchange.QueryGetAllOrdersRequest{Pagination: &query.PageRequest{Offset: 2, Key: makeKey(fiveOrders[0])}},
			expInErr: []string{invalidArgErr, "error iterating all orders",
				"invalid request, either offset or key is expected, got both"},
		},
		{
			name:    "no orders in state",
			expResp: &exchange.QueryGetAllOrdersResponse{Pagination: &query.PageResponse{}},
		},
		{
			name:    "5 orders: get all: nil req",
			setup:   fiveOrderSetup,
			req:     nil,
			expResp: &exchange.QueryGetAllOrdersResponse{Orders: fiveOrders, Pagination: &query.PageResponse{Total: 5}},
		},
		{
			name:    "5 orders: get all: empty req",
			setup:   fiveOrderSetup,
			req:     &exchange.QueryGetAllOrdersRequest{},
			expResp: &exchange.QueryGetAllOrdersResponse{Orders: fiveOrders, Pagination: &query.PageResponse{Total: 5}},
		},
		{
			name:    "5 orders: get all: empty pagination",
			setup:   fiveOrderSetup,
			req:     &exchange.QueryGetAllOrdersRequest{Pagination: &query.PageRequest{}},
			expResp: &exchange.QueryGetAllOrdersResponse{Orders: fiveOrders, Pagination: &query.PageResponse{Total: 5}},
		},
		{
			name:  "5 orders: limit 2",
			setup: fiveOrderSetup,
			req:   &exchange.QueryGetAllOrdersRequest{Pagination: &query.PageRequest{Limit: 2}},
			expResp: &exchange.QueryGetAllOrdersResponse{
				Orders:     fiveOrders[0:2],
				Pagination: &query.PageResponse{NextKey: makeKey(fiveOrders[2])},
			},
		},
		{
			name:  "5 orders: get second using key",
			setup: fiveOrderSetup,
			req:   &exchange.QueryGetAllOrdersRequest{Pagination: &query.PageRequest{Limit: 1, Key: makeKey(fiveOrders[1])}},
			expResp: &exchange.QueryGetAllOrdersResponse{
				Orders:     fiveOrders[1:2],
				Pagination: &query.PageResponse{NextKey: makeKey(fiveOrders[2])},
			},
		},
		{
			name:  "5 orders: get third and fourth using offset",
			setup: fiveOrderSetup,
			req:   &exchange.QueryGetAllOrdersRequest{Pagination: &query.PageRequest{Limit: 2, Offset: 2}},
			expResp: &exchange.QueryGetAllOrdersResponse{
				Orders:     fiveOrders[2:4],
				Pagination: &query.PageResponse{NextKey: makeKey(fiveOrders[4])},
			},
		},
		{
			name:  "5 orders: get all: reversed",
			setup: fiveOrderSetup,
			req:   &exchange.QueryGetAllOrdersRequest{Pagination: &query.PageRequest{Reverse: true}},
			expResp: &exchange.QueryGetAllOrdersResponse{
				Orders:     reverseSlice(fiveOrders),
				Pagination: &query.PageResponse{Total: 5}},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			runQueryTestCase(s, testDef, tc)
		})
	}
}

// TODO[1789] func (s *TestSuite) TestQueryServer_GetCommitment()

// TODO[1789] func (s *TestSuite) TestQueryServer_GetAccountCommitments()

// TODO[1789] func (s *TestSuite) TestQueryServer_GetMarketCommitments()

// TODO[1789] func (s *TestSuite) TestQueryServer_GetAllCommitments()

func (s *TestSuite) TestQueryServer_GetMarket() {
	testDef := queryTestDef[exchange.QueryGetMarketRequest, exchange.QueryGetMarketResponse]{
		queryName: "GetMarket",
		query:     keeper.NewQueryServer(s.k).GetMarket,
	}

	tests := []queryTestCase[exchange.QueryGetMarketRequest, exchange.QueryGetMarketResponse]{
		{
			name:     "nil request",
			req:      nil,
			expInErr: []string{invalidArgErr, "empty request"},
		},
		{
			name:     "market 0",
			req:      &exchange.QueryGetMarketRequest{MarketId: 0},
			expInErr: []string{invalidArgErr, "empty request"},
		},
		{
			name:     "empty state",
			req:      &exchange.QueryGetMarketRequest{MarketId: 1},
			expInErr: []string{invalidArgErr, "market 1 not found"},
		},
		{
			name: "market does not exist",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{MarketId: 1})
				s.requireCreateMarketUnmocked(exchange.Market{MarketId: 2})
				s.requireCreateMarketUnmocked(exchange.Market{MarketId: 4})
				s.requireCreateMarketUnmocked(exchange.Market{MarketId: 5})
			},
			req:      &exchange.QueryGetMarketRequest{MarketId: 3},
			expInErr: []string{invalidArgErr, "market 3 not found"},
		},
		{
			name: "market exists",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{MarketId: 1})
				s.requireCreateMarketUnmocked(exchange.Market{MarketId: 2})
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 3,
					MarketDetails: exchange.MarketDetails{
						Name:        "Market Three",
						Description: "This is the third market. Not the first or second. And fourth is just too far.",
						WebsiteUrl:  "not actually a websute url for market 3",
						IconUri:     "https://www.example.com/market/3/icon",
					},
					FeeCreateAskFlat:          s.coins("10fig,100grape"),
					FeeCreateBidFlat:          s.coins("20fig,200grape"),
					FeeSellerSettlementFlat:   s.coins("10pineapple,50prune"),
					FeeSellerSettlementRatios: s.ratios("1000pineapple:1pineapple,100prune:1prune"),
					FeeBuyerSettlementFlat:    s.coins("12pineapple60prune"),
					FeeBuyerSettlementRatios:  s.ratios("1000pineapple:3pineapple,100prune:3prune"),
					AcceptingOrders:           true,
					AllowUserSettlement:       true,
					AccessGrants: []exchange.AccessGrant{
						{Address: s.addr1.String(), Permissions: exchange.AllPermissions()},
						{Address: s.addr2.String(), Permissions: []exchange.Permission{1, 2}},
						{Address: s.addr3.String(), Permissions: []exchange.Permission{3, 4}},
						{Address: s.addr4.String(), Permissions: []exchange.Permission{5, 6}},
						{Address: s.addr5.String(), Permissions: []exchange.Permission{2, 4, 6, 7}},
					},
					ReqAttrCreateAsk: []string{"ask.good.kyc", "*.my.custom"},
					ReqAttrCreateBid: []string{"bid.good.kyc", "*.my.custom"},

					AllowCommitments:         true,
					FeeCreateCommitmentFlat:  s.coins("30fig,300grape"),
					CommitmentSettlementBips: 23,
					IntermediaryDenom:        "cherry",
					ReqAttrCreateCommitment:  []string{"commitment.good.kyc", "*.my.custom"},
				})
				s.requireCreateMarketUnmocked(exchange.Market{MarketId: 4})
				s.requireCreateMarketUnmocked(exchange.Market{MarketId: 5})
			},
			req: &exchange.QueryGetMarketRequest{MarketId: 3},
			expResp: &exchange.QueryGetMarketResponse{
				Address: exchange.GetMarketAddress(3).String(),
				Market: &exchange.Market{
					MarketId: 3,
					MarketDetails: exchange.MarketDetails{
						Name:        "Market Three",
						Description: "This is the third market. Not the first or second. And fourth is just too far.",
						WebsiteUrl:  "not actually a websute url for market 3",
						IconUri:     "https://www.example.com/market/3/icon",
					},
					FeeCreateAskFlat:          s.coins("10fig,100grape"),
					FeeCreateBidFlat:          s.coins("20fig,200grape"),
					FeeSellerSettlementFlat:   s.coins("10pineapple,50prune"),
					FeeSellerSettlementRatios: s.ratios("1000pineapple:1pineapple,100prune:1prune"),
					FeeBuyerSettlementFlat:    s.coins("12pineapple60prune"),
					FeeBuyerSettlementRatios:  s.ratios("1000pineapple:3pineapple,100prune:3prune"),
					AcceptingOrders:           true,
					AllowUserSettlement:       true,
					AccessGrants: []exchange.AccessGrant{
						{Address: s.addr1.String(), Permissions: exchange.AllPermissions()},
						{Address: s.addr2.String(), Permissions: []exchange.Permission{1, 2}},
						{Address: s.addr3.String(), Permissions: []exchange.Permission{3, 4}},
						{Address: s.addr4.String(), Permissions: []exchange.Permission{5, 6}},
						{Address: s.addr5.String(), Permissions: []exchange.Permission{2, 4, 6, 7}},
					},
					ReqAttrCreateAsk: []string{"ask.good.kyc", "*.my.custom"},
					ReqAttrCreateBid: []string{"bid.good.kyc", "*.my.custom"},

					AllowCommitments:         true,
					FeeCreateCommitmentFlat:  s.coins("30fig,300grape"),
					CommitmentSettlementBips: 23,
					IntermediaryDenom:        "cherry",
					ReqAttrCreateCommitment:  []string{"commitment.good.kyc", "*.my.custom"},
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			runQueryTestCase(s, testDef, tc)
		})
	}
}

func (s *TestSuite) TestQueryServer_GetAllMarkets() {
	briefIDStringer := func(brief *exchange.MarketBrief) string {
		if brief == nil {
			return "<nil>"
		}
		return fmt.Sprintf("%d", brief.MarketId)
	}
	testDef := queryTestDef[exchange.QueryGetAllMarketsRequest, exchange.QueryGetAllMarketsResponse]{
		queryName: "GetAllMarkets",
		query:     keeper.NewQueryServer(s.k).GetAllMarkets,
		followup: func(expected, actual *exchange.QueryGetAllMarketsResponse) {
			assertEqualSlice(s, expected.Markets, actual.Markets, briefIDStringer, "Markets")
			s.assertEqualPageResponse(expected.Pagination, actual.Pagination, "Pagination")
		},
	}
	makeKey := func(market *exchange.Market) []byte {
		return keeper.Uint32Bz(market.MarketId)
	}

	newMarket := func(marketID uint32) *exchange.Market {
		return &exchange.Market{
			MarketId: marketID,
			MarketDetails: exchange.MarketDetails{
				Name:        fmt.Sprintf("Market %d", marketID),
				Description: fmt.Sprintf("This is the description of market %d.", marketID),
				WebsiteUrl:  fmt.Sprintf("https://example.com/market/%d/info", marketID),
				IconUri:     fmt.Sprintf("https://example.com/market/%d/icon", marketID),
			},
		}
	}
	fiveMarkets := []*exchange.Market{
		newMarket(6),
		newMarket(34),
		newMarket(53),
		newMarket(81),
		newMarket(98),
	}
	fiveMarketsSetup := func() {
		s.requireCreateMarketUnmocked(*fiveMarkets[1])
		s.requireCreateMarketUnmocked(*fiveMarkets[0])
		s.requireCreateMarketUnmocked(*fiveMarkets[3])
		s.requireCreateMarketUnmocked(*fiveMarkets[2])
		s.requireCreateMarketUnmocked(*fiveMarkets[4])
	}

	newBrief := func(marketID uint32) *exchange.MarketBrief {
		market := newMarket(marketID)
		return &exchange.MarketBrief{
			MarketId:      market.MarketId,
			MarketAddress: exchange.GetMarketAddress(market.MarketId).String(),
			MarketDetails: market.MarketDetails,
		}
	}
	fiveBriefs := make([]*exchange.MarketBrief, len(fiveMarkets))
	for i, market := range fiveMarkets {
		fiveBriefs[i] = newBrief(market.MarketId)
	}

	tests := []queryTestCase[exchange.QueryGetAllMarketsRequest, exchange.QueryGetAllMarketsResponse]{
		{
			name: "both key and offset provided",
			req: &exchange.QueryGetAllMarketsRequest{
				Pagination: &query.PageRequest{Key: makeKey(fiveMarkets[1]), Offset: 3},
			},
			expInErr: []string{invalidArgErr, "error iterating all known markets",
				"invalid request, either offset or key is expected, got both"},
		},
		{
			name: "bad market key",
			setup: func() {
				s.requireCreateMarketUnmocked(*newMarket(1))
				s.getStore().Set(s.badKey(keeper.MakeKeyKnownMarketID(2)), []byte{})
				s.requireCreateMarketUnmocked(*newMarket(3))
			},
			expResp: &exchange.QueryGetAllMarketsResponse{
				Markets:    []*exchange.MarketBrief{newBrief(1), newBrief(3)},
				Pagination: &query.PageResponse{Total: 2},
			},
		},
		{
			name: "market account does not exist",
			setup: func() {
				s.requireCreateMarketUnmocked(*newMarket(1))
				keeper.StoreMarket(s.getStore(), *newMarket(2))
				s.requireCreateMarketUnmocked(*newMarket(3))
			},
			expResp: &exchange.QueryGetAllMarketsResponse{
				Markets:    []*exchange.MarketBrief{newBrief(1), newBrief(3)},
				Pagination: &query.PageResponse{Total: 3},
			},
		},
		{
			name:    "no markets in state",
			expResp: &exchange.QueryGetAllMarketsResponse{Pagination: &query.PageResponse{Total: 0}},
		},
		{
			name:  "five markets: nil req",
			setup: fiveMarketsSetup,
			req:   nil,
			expResp: &exchange.QueryGetAllMarketsResponse{
				Markets:    fiveBriefs,
				Pagination: &query.PageResponse{Total: 5},
			},
		},
		{
			name:  "five markets: empty req",
			setup: fiveMarketsSetup,
			req:   &exchange.QueryGetAllMarketsRequest{},
			expResp: &exchange.QueryGetAllMarketsResponse{
				Markets:    fiveBriefs,
				Pagination: &query.PageResponse{Total: 5},
			},
		},
		{
			name:  "five markets: empty pagination",
			setup: fiveMarketsSetup,
			req:   &exchange.QueryGetAllMarketsRequest{Pagination: &query.PageRequest{}},
			expResp: &exchange.QueryGetAllMarketsResponse{
				Markets:    fiveBriefs,
				Pagination: &query.PageResponse{Total: 5},
			},
		},
		{
			name:  "five markets: reversed",
			setup: fiveMarketsSetup,
			req:   &exchange.QueryGetAllMarketsRequest{Pagination: &query.PageRequest{Reverse: true}},
			expResp: &exchange.QueryGetAllMarketsResponse{
				Markets:    reverseSlice(fiveBriefs),
				Pagination: &query.PageResponse{Total: 5},
			},
		},
		{
			name:  "five markets: limit 3",
			setup: fiveMarketsSetup,
			req:   &exchange.QueryGetAllMarketsRequest{Pagination: &query.PageRequest{Limit: 3}},
			expResp: &exchange.QueryGetAllMarketsResponse{
				Markets:    fiveBriefs[0:3],
				Pagination: &query.PageResponse{NextKey: makeKey(fiveMarkets[3])},
			},
		},
		{
			name:  "five markets: limit 3, reversed",
			setup: fiveMarketsSetup,
			req:   &exchange.QueryGetAllMarketsRequest{Pagination: &query.PageRequest{Limit: 3, Reverse: true}},
			expResp: &exchange.QueryGetAllMarketsResponse{
				Markets:    reverseSlice(fiveBriefs[2:]),
				Pagination: &query.PageResponse{NextKey: makeKey(fiveMarkets[1])},
			},
		},
		{
			name:  "five markets: just second using key",
			setup: fiveMarketsSetup,
			req:   &exchange.QueryGetAllMarketsRequest{Pagination: &query.PageRequest{Limit: 1, Key: makeKey(fiveMarkets[1])}},
			expResp: &exchange.QueryGetAllMarketsResponse{
				Markets:    fiveBriefs[1:2],
				Pagination: &query.PageResponse{NextKey: makeKey(fiveMarkets[2])},
			},
		},
		{
			name:  "five markets: just third and fourth using offset",
			setup: fiveMarketsSetup,
			req:   &exchange.QueryGetAllMarketsRequest{Pagination: &query.PageRequest{Limit: 2, Offset: 2}},
			expResp: &exchange.QueryGetAllMarketsResponse{
				Markets:    fiveBriefs[2:4],
				Pagination: &query.PageResponse{NextKey: makeKey(fiveMarkets[4])},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			runQueryTestCase(s, testDef, tc)
		})
	}
}

func (s *TestSuite) TestQueryServer_Params() {
	testDef := queryTestDef[exchange.QueryParamsRequest, exchange.QueryParamsResponse]{
		queryName: "Params",
		query:     keeper.NewQueryServer(s.k).Params,
	}

	tests := []queryTestCase[exchange.QueryParamsRequest, exchange.QueryParamsResponse]{
		{
			name: "no params in state, nil req",
			setup: func() {
				s.k.SetParams(s.ctx, nil)
			},
			req:     nil,
			expResp: &exchange.QueryParamsResponse{Params: exchange.DefaultParams()},
		},
		{
			name: "no params in state, empty req",
			setup: func() {
				s.k.SetParams(s.ctx, nil)
			},
			req:     &exchange.QueryParamsRequest{},
			expResp: &exchange.QueryParamsResponse{Params: exchange.DefaultParams()},
		},
		{
			name: "default params in state, nil req",
			setup: func() {
				s.k.SetParams(s.ctx, exchange.DefaultParams())
			},
			req:     nil,
			expResp: &exchange.QueryParamsResponse{Params: exchange.DefaultParams()},
		},
		{
			name: "default params in state, empty req",
			setup: func() {
				s.k.SetParams(s.ctx, exchange.DefaultParams())
			},
			req:     nil,
			expResp: &exchange.QueryParamsResponse{Params: exchange.DefaultParams()},
		},
		{
			name: "just the default split changed",
			setup: func() {
				s.k.SetParams(s.ctx, &exchange.Params{DefaultSplit: 987})
			},
			expResp: &exchange.QueryParamsResponse{Params: &exchange.Params{DefaultSplit: 987}},
		},
		{
			name: "with denom splits, nil req",
			setup: func() {
				s.k.SetParams(s.ctx, &exchange.Params{
					DefaultSplit: 0,
					DenomSplits: []exchange.DenomSplit{
						{Denom: "apple", Split: 500},
						{Denom: "banana", Split: 333}, // mmmmmmmm
						{Denom: "cactus", Split: 777},
					},
				})
			},
			expResp: &exchange.QueryParamsResponse{Params: &exchange.Params{
				DefaultSplit: 0,
				DenomSplits: []exchange.DenomSplit{
					{Denom: "apple", Split: 500},
					{Denom: "banana", Split: 333},
					{Denom: "cactus", Split: 777},
				},
			}},
		},
		{
			name: "with denom splits, empty req",
			setup: func() {
				s.k.SetParams(s.ctx, &exchange.Params{
					DefaultSplit: 1000,
					DenomSplits: []exchange.DenomSplit{
						{Denom: "acorn", Split: 600},
						{Denom: "blueberry", Split: 55},
						{Denom: "cherry", Split: 1234},
						{Denom: "date", Split: 1000},
					},
				})
			},
			req: &exchange.QueryParamsRequest{},
			expResp: &exchange.QueryParamsResponse{Params: &exchange.Params{
				DefaultSplit: 1000,
				DenomSplits: []exchange.DenomSplit{
					{Denom: "acorn", Split: 600},
					{Denom: "blueberry", Split: 55},
					{Denom: "cherry", Split: 1234},
					{Denom: "date", Split: 1000},
				},
			}},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			runQueryTestCase(s, testDef, tc)
		})
	}
}

// TODO[1789] func (s *TestSuite) TestQueryServer_CommitmentSettlementFeeCalc()

func (s *TestSuite) TestQueryServer_ValidateCreateMarket() {
	testDef := queryTestDef[exchange.QueryValidateCreateMarketRequest, exchange.QueryValidateCreateMarketResponse]{
		queryName: "ValidateCreateMarket",
		query:     keeper.NewQueryServer(s.k).ValidateCreateMarket,
	}

	tests := []queryTestCase[exchange.QueryValidateCreateMarketRequest, exchange.QueryValidateCreateMarketResponse]{
		{
			name:     "nil req",
			req:      nil,
			expInErr: []string{invalidArgErr, "empty request"},
		},
		{
			name:     "empty req",
			req:      &exchange.QueryValidateCreateMarketRequest{},
			expInErr: []string{invalidArgErr, "empty request"},
		},
		{
			name: "invalid market",
			req: &exchange.QueryValidateCreateMarketRequest{CreateMarketRequest: &exchange.MsgGovCreateMarketRequest{
				Authority: s.k.GetAuthority(),
				Market: exchange.Market{
					MarketDetails: exchange.MarketDetails{Name: strings.Repeat("s", exchange.MaxName+1)},
				},
			}},
			expResp: &exchange.QueryValidateCreateMarketResponse{
				Error: fmt.Sprintf("name length %d exceeds maximum length of %d",
					exchange.MaxName+1, exchange.MaxName),
			},
		},
		{
			name: "no authority",
			req: &exchange.QueryValidateCreateMarketRequest{CreateMarketRequest: &exchange.MsgGovCreateMarketRequest{
				Authority: "",
			}},
			expResp: &exchange.QueryValidateCreateMarketResponse{
				Error: "invalid authority: empty address string is not allowed",
			},
		},
		{
			name: "bad authority",
			req: &exchange.QueryValidateCreateMarketRequest{CreateMarketRequest: &exchange.MsgGovCreateMarketRequest{
				Authority: "bad",
			}},
			expResp: &exchange.QueryValidateCreateMarketResponse{
				Error: "invalid authority: decoding bech32 failed: invalid bech32 string length 3",
			},
		},
		{
			name: "wrong authority",
			req: &exchange.QueryValidateCreateMarketRequest{CreateMarketRequest: &exchange.MsgGovCreateMarketRequest{
				Authority: s.addr1.String(),
			}},
			expResp: &exchange.QueryValidateCreateMarketResponse{
				Error: "expected \"" + s.k.GetAuthority() + "\" got \"" + s.addr1.String() + "\": " +
					"expected gov account as only signer for proposal message",
			},
		},
		{
			name: "market already exists",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{MarketId: 1})
			},
			req: &exchange.QueryValidateCreateMarketRequest{CreateMarketRequest: &exchange.MsgGovCreateMarketRequest{
				Authority: s.k.GetAuthority(),
				Market:    exchange.Market{MarketId: 1},
			}},
			expResp: &exchange.QueryValidateCreateMarketResponse{
				Error: "market id 1 account " + exchange.GetMarketAddress(1).String() + " already exists",
			},
		},
		{
			name: "problems with market definition",
			req: &exchange.QueryValidateCreateMarketRequest{CreateMarketRequest: &exchange.MsgGovCreateMarketRequest{
				Authority: s.k.GetAuthority(),
				Market: exchange.Market{
					ReqAttrCreateAsk: []string{" ask .bb.cc"},
					ReqAttrCreateBid: []string{" bid .bb.cc"},
				},
			}},
			expResp: &exchange.QueryValidateCreateMarketResponse{
				GovPropWillPass: true,
				Error: s.joinErrs(
					"create ask required attribute \" ask .bb.cc\" is not normalized, expected \"ask.bb.cc\"",
					"create bid required attribute \" bid .bb.cc\" is not normalized, expected \"bid.bb.cc\"",
				),
			},
		},
		{
			name: "all good",
			req: &exchange.QueryValidateCreateMarketRequest{CreateMarketRequest: &exchange.MsgGovCreateMarketRequest{
				Authority: s.k.GetAuthority(),
				Market: exchange.Market{
					ReqAttrCreateAsk: []string{"ask.bb.cc"},
					ReqAttrCreateBid: []string{"bid.bb.cc"},
				},
			}},
			expResp: &exchange.QueryValidateCreateMarketResponse{GovPropWillPass: true, Error: ""},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			runQueryTestCase(s, testDef, tc)
		})
	}
}

func (s *TestSuite) TestQueryServer_ValidateMarket() {
	testDef := queryTestDef[exchange.QueryValidateMarketRequest, exchange.QueryValidateMarketResponse]{
		queryName: "ValidateMarket",
		query:     keeper.NewQueryServer(s.k).ValidateMarket,
	}

	tests := []queryTestCase[exchange.QueryValidateMarketRequest, exchange.QueryValidateMarketResponse]{
		{
			name:     "nil req",
			req:      nil,
			expInErr: []string{invalidArgErr, "empty request"},
		},
		{
			name:     "market 0",
			req:      &exchange.QueryValidateMarketRequest{MarketId: 0},
			expInErr: []string{invalidArgErr, "empty request"},
		},
		{
			name:    "market does not exist",
			req:     &exchange.QueryValidateMarketRequest{MarketId: 66},
			expResp: &exchange.QueryValidateMarketResponse{Error: "market 66 does not exist"},
		},
		{
			name: "bad ratios",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:                  2,
					FeeSellerSettlementRatios: s.ratios("100peach:1peach,100plum:3plum"),
					FeeBuyerSettlementRatios:  s.ratios("100plum:1plum,100prune:7prune"),
				})
			},
			req: &exchange.QueryValidateMarketRequest{MarketId: 2},
			expResp: &exchange.QueryValidateMarketResponse{Error: s.joinErrs(
				"seller settlement fee ratios have price denom \"peach\" but there are no "+
					"buyer settlement fee ratios with that price denom",
				"buyer settlement fee ratios have price denom \"prune\" but there is not a "+
					"seller settlement fee ratio with that price denom",
			)},
		},
		{
			name: "bad commitment setup",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:                 2,
					CommitmentSettlementBips: 3,
					IntermediaryDenom:        "",
				})
			},
			req: &exchange.QueryValidateMarketRequest{MarketId: 2},
			expResp: &exchange.QueryValidateMarketResponse{
				Error: "no intermediary denom defined, but commitment settlement bips 3 is not zero",
			},
		},
		{
			name: "all good",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:                  2,
					FeeSellerSettlementRatios: s.ratios("100peach:1peach,100plum:3plum,100prune:7prune"),
					FeeBuyerSettlementRatios:  s.ratios("100peach:3peach,100plum:7plum,100prune:1prune"),
				})
			},
			req:     &exchange.QueryValidateMarketRequest{MarketId: 2},
			expResp: &exchange.QueryValidateMarketResponse{Error: ""},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			runQueryTestCase(s, testDef, tc)
		})
	}
}

func (s *TestSuite) TestQueryServer_ValidateManageFees() {
	testDef := queryTestDef[exchange.QueryValidateManageFeesRequest, exchange.QueryValidateManageFeesResponse]{
		queryName: "ValidateManageFees",
		query:     keeper.NewQueryServer(s.k).ValidateManageFees,
	}

	tests := []queryTestCase[exchange.QueryValidateManageFeesRequest, exchange.QueryValidateManageFeesResponse]{
		{
			name:     "nil req",
			req:      nil,
			expInErr: []string{invalidArgErr, "empty request"},
		},
		{
			name:     "empty req",
			req:      &exchange.QueryValidateManageFeesRequest{},
			expInErr: []string{invalidArgErr, "empty request"},
		},
		{
			name: "invalid msg",
			req: &exchange.QueryValidateManageFeesRequest{ManageFeesRequest: &exchange.MsgGovManageFeesRequest{
				Authority: "", MarketId: 1,
			}},
			expResp: &exchange.QueryValidateManageFeesResponse{
				Error: s.joinErrs(
					"invalid authority: empty address string is not allowed",
					"no updates",
				),
			},
		},
		{
			name: "wrong authority",
			req: &exchange.QueryValidateManageFeesRequest{ManageFeesRequest: &exchange.MsgGovManageFeesRequest{
				Authority: s.addr1.String(), MarketId: 1,
				AddFeeCreateAskFlat: s.coins("100plum"),
			}},
			expResp: &exchange.QueryValidateManageFeesResponse{
				Error: "expected \"" + s.k.GetAuthority() + "\" got \"" + s.addr1.String() + "\": " +
					"expected gov account as only signer for proposal message",
			},
		},
		{
			name: "market does not exist",
			req: &exchange.QueryValidateManageFeesRequest{ManageFeesRequest: &exchange.MsgGovManageFeesRequest{
				Authority: s.k.GetAuthority(), MarketId: 1,
				AddFeeCreateAskFlat: s.coins("100plum"),
			}},
			expResp: &exchange.QueryValidateManageFeesResponse{
				Error: "market 1 does not exist",
			},
		},
		{
			name: "add/rem create-ask errors",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:         7,
					FeeCreateAskFlat: s.coins("100peach"),
				})
			},
			req: &exchange.QueryValidateManageFeesRequest{ManageFeesRequest: &exchange.MsgGovManageFeesRequest{
				Authority: s.k.GetAuthority(), MarketId: 7,
				RemoveFeeCreateAskFlat: s.coins("100plum"),
				AddFeeCreateAskFlat:    s.coins("90peach"),
			}},
			expResp: &exchange.QueryValidateManageFeesResponse{
				GovPropWillPass: true,
				Error: s.joinErrs(
					"cannot remove create-ask flat fee \"100plum\": no such fee exists",
					"cannot add create-ask flat fee \"90peach\": fee with that denom already exists",
				),
			},
		},
		{
			name: "add/rem create-bid errors",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:         7,
					FeeCreateBidFlat: s.coins("100apple"),
				})
			},
			req: &exchange.QueryValidateManageFeesRequest{ManageFeesRequest: &exchange.MsgGovManageFeesRequest{
				Authority: s.k.GetAuthority(), MarketId: 7,
				RemoveFeeCreateBidFlat: s.coins("100acorn"),
				AddFeeCreateBidFlat:    s.coins("90apple"),
			}},
			expResp: &exchange.QueryValidateManageFeesResponse{
				GovPropWillPass: true,
				Error: s.joinErrs(
					"cannot remove create-bid flat fee \"100acorn\": no such fee exists",
					"cannot add create-bid flat fee \"90apple\": fee with that denom already exists",
				),
			},
		},
		{
			name: "add/rem create-commitment errors",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:                7,
					FeeCreateCommitmentFlat: s.coins("100orange"),
				})
			},
			req: &exchange.QueryValidateManageFeesRequest{ManageFeesRequest: &exchange.MsgGovManageFeesRequest{
				Authority: s.k.GetAuthority(), MarketId: 7,
				RemoveFeeCreateCommitmentFlat: s.coins("100olive"),
				AddFeeCreateCommitmentFlat:    s.coins("90orange"),
			}},
			expResp: &exchange.QueryValidateManageFeesResponse{
				GovPropWillPass: true,
				Error: s.joinErrs(
					"cannot remove create-commitment flat fee \"100olive\": no such fee exists",
					"cannot add create-commitment flat fee \"90orange\": fee with that denom already exists",
				),
			},
		},
		{
			name: "add/rem seller flat errors",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:                7,
					FeeSellerSettlementFlat: s.coins("100cherry"),
				})
			},
			req: &exchange.QueryValidateManageFeesRequest{ManageFeesRequest: &exchange.MsgGovManageFeesRequest{
				Authority: s.k.GetAuthority(), MarketId: 7,
				RemoveFeeSellerSettlementFlat: s.coins("100cactus"),
				AddFeeSellerSettlementFlat:    s.coins("90cherry"),
			}},
			expResp: &exchange.QueryValidateManageFeesResponse{
				GovPropWillPass: true,
				Error: s.joinErrs(
					"cannot remove seller settlement flat fee \"100cactus\": no such fee exists",
					"cannot add seller settlement flat fee \"90cherry\": fee with that denom already exists",
				),
			},
		},
		{
			name: "add/rem seller ratio errors",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:                  7,
					FeeSellerSettlementRatios: s.ratios("100pear:1pear"),
				})
			},
			req: &exchange.QueryValidateManageFeesRequest{ManageFeesRequest: &exchange.MsgGovManageFeesRequest{
				Authority: s.k.GetAuthority(), MarketId: 7,
				RemoveFeeSellerSettlementRatios: s.ratios("100prune:1prune"),
				AddFeeSellerSettlementRatios:    s.ratios("90pear:1pear"),
			}},
			expResp: &exchange.QueryValidateManageFeesResponse{
				GovPropWillPass: true,
				Error: s.joinErrs(
					"cannot remove seller settlement ratio fee \"100prune:1prune\": no such ratio exists",
					"cannot add seller settlement ratio fee \"90pear:1pear\": ratio with those denoms already exists",
				),
			},
		},
		{
			name: "add/rem buyer flat errors",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:               7,
					FeeBuyerSettlementFlat: s.coins("100date"),
				})
			},
			req: &exchange.QueryValidateManageFeesRequest{ManageFeesRequest: &exchange.MsgGovManageFeesRequest{
				Authority: s.k.GetAuthority(), MarketId: 7,
				RemoveFeeBuyerSettlementFlat: s.coins("100durian"),
				AddFeeBuyerSettlementFlat:    s.coins("90date"),
			}},
			expResp: &exchange.QueryValidateManageFeesResponse{
				GovPropWillPass: true,
				Error: s.joinErrs(
					"cannot remove buyer settlement flat fee \"100durian\": no such fee exists",
					"cannot add buyer settlement flat fee \"90date\": fee with that denom already exists",
				),
			},
		},
		{
			name: "add/rem buyer ratio errors",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:                 7,
					FeeBuyerSettlementRatios: s.ratios("100banana:1banana"),
				})
			},
			req: &exchange.QueryValidateManageFeesRequest{ManageFeesRequest: &exchange.MsgGovManageFeesRequest{
				Authority: s.k.GetAuthority(), MarketId: 7,
				RemoveFeeBuyerSettlementRatios: s.ratios("100blueberry:1blueberry"),
				AddFeeBuyerSettlementRatios:    s.ratios("90banana:1banana"),
			}},
			expResp: &exchange.QueryValidateManageFeesResponse{
				GovPropWillPass: true,
				Error: s.joinErrs(
					"cannot remove buyer settlement ratio fee \"100blueberry:1blueberry\": no such ratio exists",
					"cannot add buyer settlement ratio fee \"90banana:1banana\": ratio with those denoms already exists",
				),
			},
		},
		{
			name: "seller ratio problems after add",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:                  4,
					FeeSellerSettlementRatios: s.ratios("100banana:1banana"),
					FeeBuyerSettlementRatios:  s.ratios("100banana:3banana"),
				})
			},
			req: &exchange.QueryValidateManageFeesRequest{ManageFeesRequest: &exchange.MsgGovManageFeesRequest{
				Authority: s.k.GetAuthority(), MarketId: 4,
				AddFeeSellerSettlementRatios: s.ratios("90apple:1apple"),
			}},
			expResp: &exchange.QueryValidateManageFeesResponse{
				GovPropWillPass: true,
				Error: "seller settlement fee ratios have price denom \"apple\" but there are no " +
					"buyer settlement fee ratios with that price denom",
			},
		},
		{
			name: "seller ratio problems after remove",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:                  4,
					FeeSellerSettlementRatios: s.ratios("100banana:1banana,90apple:1apple"),
					FeeBuyerSettlementRatios:  s.ratios("100banana:3banana,90apple:7apple"),
				})
			},
			req: &exchange.QueryValidateManageFeesRequest{ManageFeesRequest: &exchange.MsgGovManageFeesRequest{
				Authority: s.k.GetAuthority(), MarketId: 4,
				RemoveFeeSellerSettlementRatios: s.ratios("90apple:1apple"),
			}},
			expResp: &exchange.QueryValidateManageFeesResponse{
				GovPropWillPass: true,
				Error: "buyer settlement fee ratios have price denom \"apple\" but there is not a " +
					"seller settlement fee ratio with that price denom",
			},
		},
		{
			name: "buyer ratio problems after add",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:                  4,
					FeeSellerSettlementRatios: s.ratios("100banana:1banana"),
					FeeBuyerSettlementRatios:  s.ratios("100banana:3banana"),
				})
			},
			req: &exchange.QueryValidateManageFeesRequest{ManageFeesRequest: &exchange.MsgGovManageFeesRequest{
				Authority: s.k.GetAuthority(), MarketId: 4,
				AddFeeBuyerSettlementRatios: s.ratios("90apple:7apple"),
			}},
			expResp: &exchange.QueryValidateManageFeesResponse{
				GovPropWillPass: true,
				Error: "buyer settlement fee ratios have price denom \"apple\" but there is not a " +
					"seller settlement fee ratio with that price denom",
			},
		},
		{
			name: "buyer ratio problems after remove",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:                  4,
					FeeSellerSettlementRatios: s.ratios("100banana:1banana,90apple:1apple"),
					FeeBuyerSettlementRatios:  s.ratios("100banana:3banana,90apple:7apple"),
				})
			},
			req: &exchange.QueryValidateManageFeesRequest{ManageFeesRequest: &exchange.MsgGovManageFeesRequest{
				Authority: s.k.GetAuthority(), MarketId: 4,
				RemoveFeeBuyerSettlementRatios: s.ratios("90apple:7apple"),
			}},
			expResp: &exchange.QueryValidateManageFeesResponse{
				GovPropWillPass: true,
				Error: "seller settlement fee ratios have price denom \"apple\" but there are no " +
					"buyer settlement fee ratios with that price denom",
			},
		},
		{
			name: "setting commitment bips without intermediary denom",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{MarketId: 4})
			},
			req: &exchange.QueryValidateManageFeesRequest{ManageFeesRequest: &exchange.MsgGovManageFeesRequest{
				Authority: s.k.GetAuthority(), MarketId: 4,
				SetFeeCommitmentSettlementBips: 35,
			}},
			expResp: &exchange.QueryValidateManageFeesResponse{
				GovPropWillPass: true,
				Error:           "no intermediary denom defined, but commitment settlement bips 35 is not zero",
			},
		},
		{
			name: "all the problems",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:                  7,
					FeeCreateAskFlat:          s.coins("100peach"),
					FeeCreateBidFlat:          s.coins("100apple"),
					FeeCreateCommitmentFlat:   s.coins("100orange"),
					FeeSellerSettlementFlat:   s.coins("100cherry"),
					FeeSellerSettlementRatios: s.ratios("100pear:1pear"),
					FeeBuyerSettlementFlat:    s.coins("100date"),
					FeeBuyerSettlementRatios:  s.ratios("100banana:1banana"),
				})
			},
			req: &exchange.QueryValidateManageFeesRequest{ManageFeesRequest: &exchange.MsgGovManageFeesRequest{
				Authority: s.k.GetAuthority(), MarketId: 7,
				RemoveFeeCreateAskFlat:          s.coins("100plum"),
				AddFeeCreateAskFlat:             s.coins("90peach"),
				RemoveFeeCreateBidFlat:          s.coins("100acorn"),
				AddFeeCreateBidFlat:             s.coins("90apple"),
				RemoveFeeSellerSettlementFlat:   s.coins("100cactus"),
				AddFeeSellerSettlementFlat:      s.coins("90cherry"),
				RemoveFeeSellerSettlementRatios: s.ratios("100prune:1prune"),
				AddFeeSellerSettlementRatios:    s.ratios("90pear:1pear"),
				RemoveFeeBuyerSettlementFlat:    s.coins("100durian"),
				AddFeeBuyerSettlementFlat:       s.coins("90date"),
				RemoveFeeBuyerSettlementRatios:  s.ratios("100blueberry:1blueberry"),
				AddFeeBuyerSettlementRatios:     s.ratios("90banana:1banana"),
				RemoveFeeCreateCommitmentFlat:   s.coins("100olive"),
				AddFeeCreateCommitmentFlat:      s.coins("90orange"),
				SetFeeCommitmentSettlementBips:  35,
			}},
			expResp: &exchange.QueryValidateManageFeesResponse{
				GovPropWillPass: true,
				Error: s.joinErrs(
					"cannot remove create-ask flat fee \"100plum\": no such fee exists",
					"cannot add create-ask flat fee \"90peach\": fee with that denom already exists",
					"cannot remove create-bid flat fee \"100acorn\": no such fee exists",
					"cannot add create-bid flat fee \"90apple\": fee with that denom already exists",
					"cannot remove create-commitment flat fee \"100olive\": no such fee exists",
					"cannot add create-commitment flat fee \"90orange\": fee with that denom already exists",
					"cannot remove seller settlement flat fee \"100cactus\": no such fee exists",
					"cannot add seller settlement flat fee \"90cherry\": fee with that denom already exists",
					"cannot remove seller settlement ratio fee \"100prune:1prune\": no such ratio exists",
					"cannot add seller settlement ratio fee \"90pear:1pear\": ratio with those denoms already exists",
					"cannot remove buyer settlement flat fee \"100durian\": no such fee exists",
					"cannot add buyer settlement flat fee \"90date\": fee with that denom already exists",
					"cannot remove buyer settlement ratio fee \"100blueberry:1blueberry\": no such ratio exists",
					"cannot add buyer settlement ratio fee \"90banana:1banana\": ratio with those denoms already exists",
					"seller settlement fee ratios have price denom \"pear\" but there are no "+
						"buyer settlement fee ratios with that price denom",
					"buyer settlement fee ratios have price denom \"banana\" but there is not a "+
						"seller settlement fee ratio with that price denom",
					"no intermediary denom defined, but commitment settlement bips 35 is not zero",
				),
			},
		},
		{
			name: "all good",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:                  7,
					FeeCreateAskFlat:          s.coins("100peach"),
					FeeCreateBidFlat:          s.coins("100apple"),
					FeeCreateCommitmentFlat:   s.coins("100orange"),
					FeeSellerSettlementFlat:   s.coins("100cherry"),
					FeeSellerSettlementRatios: s.ratios("100pear:1pear"),
					FeeBuyerSettlementFlat:    s.coins("100date"),
					FeeBuyerSettlementRatios:  s.ratios("100banana:1banana"),
				})
			},
			req: &exchange.QueryValidateManageFeesRequest{ManageFeesRequest: &exchange.MsgGovManageFeesRequest{
				Authority: s.k.GetAuthority(), MarketId: 7,
				RemoveFeeCreateAskFlat:          s.coins("100peach"),
				AddFeeCreateAskFlat:             s.coins("90peach"),
				RemoveFeeCreateBidFlat:          s.coins("100apple"),
				AddFeeCreateBidFlat:             s.coins("90apple"),
				RemoveFeeSellerSettlementFlat:   s.coins("100cherry"),
				AddFeeSellerSettlementFlat:      s.coins("90cherry"),
				RemoveFeeSellerSettlementRatios: s.ratios("100pear:1pear"),
				AddFeeSellerSettlementRatios:    s.ratios("90pear:1pear,100banana:1banana"),
				RemoveFeeBuyerSettlementFlat:    s.coins("100date"),
				AddFeeBuyerSettlementFlat:       s.coins("90date"),
				RemoveFeeBuyerSettlementRatios:  s.ratios("100banana:1banana"),
				AddFeeBuyerSettlementRatios:     s.ratios("90banana:1banana,100pear:1pear"),
			}},
			expResp: &exchange.QueryValidateManageFeesResponse{GovPropWillPass: true, Error: ""},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			runQueryTestCase(s, testDef, tc)
		})
	}
}
