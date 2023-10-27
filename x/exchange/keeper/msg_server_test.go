package keeper_test

import (
	"context"
	"fmt"

	"github.com/gogo/protobuf/proto"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/exchange"
	"github.com/provenance-io/provenance/x/exchange/keeper"
	"github.com/provenance-io/provenance/x/hold"
)

// invReqErr is the error added by sdkerrors.ErrInvalidRequest.
const invReqErr = "invalid request"

// msgServerTestDef is the definition of a msg service endoint to be tested.
type msgServerTestDef[R any, S any, F any] struct {
	// endpointName is the name of the endpoint being tested.
	endpointName string
	// endpoint is the endpoint function to invoke.
	endpoint func(goCtx context.Context, msg *R) (*S, error)
	// expResp is the expected response from the endpoint. It's only used if an error is not expected.
	expResp *S
	// followup is a function that runs any needed followup checks.
	// This is only executed if an error neither expected, nor received.
	// The TestSuite's ctx will be the cached context with the results of the setup and endpoint applied.
	followup func(msg *R, fArgs F)
}

// msgServerTestCase is a test case for a msg server endpoint
type msgServerTestCase[R any, F any] struct {
	// name is the name of the test case.
	name string
	// setup is a function that does any needed app/state setup.
	// A cached context is used for tests, so this setup will not carry over between test cases.
	setup func()
	// msg is the sdk.Msg to provide to the endpoint.
	msg R
	// expInErr is the strings that are expected to be in the error returned by the endpoint.
	// If empty, that error is expected to be nil.
	expInErr []string
	// fArgs are any args to provide to the followup function.
	fArgs F
	// expEvents are the typed events that should be emitted.
	// These are only checked if an error is neither expected, nor received.
	expEvents sdk.Events
}

// runMsgServerTestCase runs a unit test on a MsgServer endpoint.
// A cached context is used so each test case won't affect the others.
func runMsgServerTestCase[R any, S any, F any](s *TestSuite, td msgServerTestDef[R, S, F], tc msgServerTestCase[R, F]) {
	s.T().Helper()
	origCtx := s.ctx
	defer func() {
		s.ctx = origCtx
	}()
	s.ctx, _ = s.ctx.CacheContext()

	var expResp *S
	if len(tc.expInErr) == 0 {
		expResp = td.expResp
	}

	if tc.setup != nil {
		tc.setup()
	}

	em := sdk.NewEventManager()
	s.ctx = s.ctx.WithEventManager(em)
	goCtx := sdk.WrapSDKContext(s.ctx)
	var resp *S
	var err error
	testFunc := func() {
		resp, err = td.endpoint(goCtx, &tc.msg)
	}
	s.Require().NotPanicsf(testFunc, td.endpointName)
	s.assertErrorContentsf(err, tc.expInErr, "%s error", td.endpointName)
	s.Assert().Equalf(expResp, resp, "%s response", td.endpointName)

	if len(tc.expInErr) > 0 || err != nil {
		return
	}

	actEvents := em.Events()
	s.assertEqualEvents(tc.expEvents, actEvents, "%s events", td.endpointName)

	td.followup(&tc.msg, tc.fArgs)
}

// untypeEvent returns sdk.TypedEventToEvent(tev) requiring it to not error.
func (s *TestSuite) untypeEvent(tev proto.Message) sdk.Event {
	rv, err := sdk.TypedEventToEvent(tev)
	s.Require().NoError(err, "TypedEventToEvent(%T)", tev)
	return rv
}

// newAttr creates a new EventAttribute with the provided key and the quoted value.
func (s *TestSuite) newAttr(key, value string) abci.EventAttribute {
	return abci.EventAttribute{Key: []byte(key), Value: []byte(fmt.Sprintf("%q", value))}
}

// newAttrNoQ creates a new EventAttribute with the provided key and value (without quoting the value).
func (s *TestSuite) newAttrNoQ(key, value string) abci.EventAttribute {
	return abci.EventAttribute{Key: []byte(key), Value: []byte(value)}
}

// eventCoinSpent creates a new "coin_spent" event.
func (s *TestSuite) eventCoinSpent(spender sdk.AccAddress, amount string) sdk.Event {
	return sdk.Event{
		Type: "coin_spent",
		Attributes: []abci.EventAttribute{
			s.newAttrNoQ("spender", spender.String()),
			s.newAttrNoQ("amount", amount),
		},
	}
}

// eventCoinReceived creates a new "coin_received" event.
func (s *TestSuite) eventCoinReceived(receiver sdk.AccAddress, amount string) sdk.Event {
	return sdk.Event{
		Type: "coin_received",
		Attributes: []abci.EventAttribute{
			s.newAttrNoQ("receiver", receiver.String()),
			s.newAttrNoQ("amount", amount),
		},
	}
}

// eventTransfer creates a new "transfer" event.
func (s *TestSuite) eventTransfer(recipient, sender sdk.AccAddress, amount string) sdk.Event {
	return sdk.Event{
		Type: "transfer",
		Attributes: []abci.EventAttribute{
			s.newAttrNoQ("recipient", recipient.String()),
			s.newAttrNoQ("sender", sender.String()),
			s.newAttrNoQ("amount", amount),
		},
	}
}

// eventMessage creates a new "message" event.
func (s *TestSuite) eventMessage(sender sdk.AccAddress) sdk.Event {
	return sdk.Event{
		Type:       "message",
		Attributes: []abci.EventAttribute{s.newAttrNoQ("sender", sender.String())},
	}
}

// eventHoldAdded creates a new event emitted when a hold is added.
func (s *TestSuite) eventHoldAdded(addr sdk.AccAddress, amount string, orderID uint64) sdk.Event {
	return s.untypeEvent(&hold.EventHoldAdded{
		Address: addr.String(), Amount: amount, Reason: fmt.Sprintf("x/exchange: order %d", orderID),
	})
}

// eventHoldAdded creates a new event emitted when a hold is released.
func (s *TestSuite) eventHoldReleased(addr sdk.AccAddress, amount string) sdk.Event {
	return s.untypeEvent(&hold.EventHoldReleased{Address: addr.String(), Amount: amount})
}

func (s *TestSuite) TestMsgServer_CreateAsk() {
	type followupArgs struct {
		expOrderID  uint64
		addr        sdk.AccAddress
		expBal      sdk.Coins
		expHoldAmt  sdk.Coins
		expSpendBal sdk.Coins
	}
	testDef := msgServerTestDef[exchange.MsgCreateAskRequest, exchange.MsgCreateAskResponse, followupArgs]{
		endpointName: "CreateAsk",
		endpoint:     keeper.NewMsgServer(s.k).CreateAsk,
		followup: func(_ *exchange.MsgCreateAskRequest, fargs followupArgs) {
			for _, expBal := range fargs.expBal {
				actBal := s.app.BankKeeper.GetBalance(s.ctx, fargs.addr, expBal.Denom)
				s.Assert().Equalf(expBal.String(), actBal.String(), "actual balance of %s", expBal.Denom)
			}
			holdAmt, err := s.app.HoldKeeper.GetHoldCoins(s.ctx, fargs.addr)
			if s.Assert().NoError(err, "GetHoldCoins(%s) error", s.getAddrName(fargs.addr)) {
				s.Assert().Equalf(fargs.expHoldAmt.String(), holdAmt.String(), "amount on hold for %s", s.getAddrName(fargs.addr))
			}
			actSpendBal := s.app.BankKeeper.SpendableCoins(s.ctx, fargs.addr)
			for _, expBal := range fargs.expSpendBal {
				actBal := actSpendBal.AmountOf(expBal.Denom)
				s.Assert().Equalf(expBal.Amount.String(), actBal.String(), "spendable balance of %s", expBal.Denom)
			}
		},
	}

	tests := []msgServerTestCase[exchange.MsgCreateAskRequest, followupArgs]{
		{
			name:  "invalid msg",
			setup: nil,
			msg: exchange.MsgCreateAskRequest{
				AskOrder: exchange.AskOrder{
					MarketId: 0, Seller: s.addr1.String(), Assets: s.coin("1apple"), Price: s.coin("1peach"),
				},
			},
			expInErr: []string{invReqErr, "invalid market id: must not be zero"},
		},
		{
			name: "market does not exist",
			msg: exchange.MsgCreateAskRequest{
				AskOrder: exchange.AskOrder{
					MarketId: 7, Seller: s.addr1.String(), Assets: s.coin("1apple"), Price: s.coin("1peach"),
				},
			},
			expInErr: []string{invReqErr, "market 7 does not exist"},
		},
		{
			name: "cannot collect creation fee",
			setup: func() {
				s.requireFundAccount(s.addr1, "9fig")
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 1, AcceptingOrders: true,
					FeeCreateAskFlat: s.coins("10fig"),
				})
			},
			msg: exchange.MsgCreateAskRequest{
				AskOrder: exchange.AskOrder{
					MarketId: 1, Seller: s.addr1.String(), Assets: s.coin("1apple"), Price: s.coin("1peach"),
				},
				OrderCreationFee: s.coinP("10fig"),
			},
			expInErr: []string{
				invReqErr, "error collecting ask order creation fee",
				"error transferring 10fig from " + s.addr1.String() + " to market 1",
				"spendable balance 9fig is smaller than 10fig",
				"insufficient funds",
			},
		},
		{
			name: "duplicate external id",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{MarketId: 1, AcceptingOrders: true})
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(8).WithBid(&exchange.BidOrder{
					MarketId:   1,
					Buyer:      s.addr5.String(),
					Assets:     s.coin("1apple"),
					Price:      s.coin("1peach"),
					ExternalId: "dupeid",
				}))
				keeper.SetLastOrderID(store, 10)
			},
			msg: exchange.MsgCreateAskRequest{
				AskOrder: exchange.AskOrder{
					MarketId:   1,
					Seller:     s.addr1.String(),
					Assets:     s.coin("1apple"),
					Price:      s.coin("1peach"),
					ExternalId: "dupeid",
				},
			},
			expInErr: []string{
				invReqErr, "error storing ask order",
				"external id \"dupeid\" is already in use by order 8: cannot be used for order 11",
			},
		},
		{
			name: "assets not in account",
			setup: func() {
				s.requireFundAccount(s.addr1, "9apple")
				s.requireCreateMarketUnmocked(exchange.Market{MarketId: 3, AcceptingOrders: true})
			},
			msg: exchange.MsgCreateAskRequest{
				AskOrder: exchange.AskOrder{
					MarketId: 3, Seller: s.addr1.String(), Assets: s.coin("10apple"), Price: s.coin("10peach"),
				},
			},
			expInErr: []string{
				invReqErr, "error placing hold for ask order 1",
				"account " + s.addr1.String() + " spendable balance 9apple is less than hold amount 10apple",
			},
		},
		{
			name: "settlement fee not in account",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 3, AcceptingOrders: true,
					FeeCreateAskFlat:        s.coins("10fig"),
					FeeSellerSettlementFlat: s.coins("5fig"),
				})
				s.requireFundAccount(s.addr1, "100apple,20fig")
				s.requireAddHold(s.addr1, "6fig", 0)
			},
			msg: exchange.MsgCreateAskRequest{
				AskOrder: exchange.AskOrder{
					MarketId: 3, Seller: s.addr1.String(),
					Assets: s.coin("100apple"), Price: s.coin("10peach"),
					SellerSettlementFlatFee: s.coinP("5fig"),
				},
				OrderCreationFee: s.coinP("10fig"),
			},
			expInErr: []string{
				invReqErr, "error placing hold for ask order 1",
				"account " + s.addr1.String() + " spendable balance 4fig is less than hold amount 5fig",
			},
		},
		{
			name: "okay: no settlement fee",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{MarketId: 5, AcceptingOrders: true})
				s.requireFundAccount(s.addr2, "100apple,100fig,100pear")
				keeper.SetLastOrderID(s.getStore(), 83)
			},
			msg: exchange.MsgCreateAskRequest{
				AskOrder: exchange.AskOrder{
					MarketId: 5, Seller: s.addr2.String(),
					Assets: s.coin("60apple"), Price: s.coin("45pear"),
				},
			},
			fArgs: followupArgs{
				expOrderID:  84,
				addr:        s.addr2,
				expBal:      s.coins("100apple,100fig,100pear"),
				expHoldAmt:  s.coins("60apple"),
				expSpendBal: s.coins("40apple,100fig,100pear"),
			},
			expEvents: sdk.Events{
				s.eventHoldAdded(s.addr2, "60apple", 84),
				s.untypeEvent(&exchange.EventOrderCreated{
					OrderId: 84, OrderType: "ask", MarketId: 5, ExternalId: "",
				}),
			},
		},
		{
			name: "okay: settlement fee same denom as price",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 2, AcceptingOrders: true,
					FeeCreateAskFlat:        s.coins("8pear"),
					FeeSellerSettlementFlat: s.coins("12pear"),
				})
				s.requireFundAccount(s.addr2, "100apple,100fig,100pear")
				keeper.SetLastOrderID(s.getStore(), 6)
			},
			msg: exchange.MsgCreateAskRequest{
				AskOrder: exchange.AskOrder{
					MarketId: 2, Seller: s.addr2.String(),
					Assets: s.coin("75apple"), Price: s.coin("45pear"),
					SellerSettlementFlatFee: s.coinP("12pear"),
					ExternalId:              "just-an-id",
				},
				OrderCreationFee: s.coinP("8pear"),
			},
			fArgs: followupArgs{
				expOrderID:  7,
				addr:        s.addr2,
				expBal:      s.coins("100apple,100fig,92pear"),
				expHoldAmt:  s.coins("75apple"),
				expSpendBal: s.coins("25apple,100fig,92pear"),
			},
			expEvents: sdk.Events{
				s.eventCoinSpent(s.addr2, "8pear"),
				s.eventCoinReceived(s.marketAddr2, "8pear"),
				s.eventTransfer(s.marketAddr2, s.addr2, "8pear"),
				s.eventMessage(s.addr2),
				s.eventCoinSpent(s.marketAddr2, "1pear"),
				s.eventCoinReceived(s.feeCollectorAddr, "1pear"),
				s.eventTransfer(s.feeCollectorAddr, s.marketAddr2, "1pear"),
				s.eventMessage(s.marketAddr2),
				s.eventHoldAdded(s.addr2, "75apple", 7),
				s.untypeEvent(&exchange.EventOrderCreated{
					OrderId: 7, OrderType: "ask", MarketId: 2, ExternalId: "just-an-id",
				}),
			},
		},
		{
			name: "okay: settlement fee diff denom from price",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 3, AcceptingOrders: true,
					FeeCreateAskFlat:        s.coins("8fig"),
					FeeSellerSettlementFlat: s.coins("12fig"),
				})
				s.requireFundAccount(s.addr2, "100apple,100fig,100pear")
				keeper.SetLastOrderID(s.getStore(), 12344)
			},
			msg: exchange.MsgCreateAskRequest{
				AskOrder: exchange.AskOrder{
					MarketId: 3, Seller: s.addr2.String(),
					Assets: s.coin("75apple"), Price: s.coin("45pear"),
					SellerSettlementFlatFee: s.coinP("12fig"),
				},
				OrderCreationFee: s.coinP("8fig"),
			},
			fArgs: followupArgs{
				expOrderID:  12345,
				addr:        s.addr2,
				expBal:      s.coins("100apple,92fig,100pear"),
				expHoldAmt:  s.coins("75apple,12fig"),
				expSpendBal: s.coins("25apple,80fig,100pear"),
			},
			expEvents: sdk.Events{
				s.eventCoinSpent(s.addr2, "8fig"),
				s.eventCoinReceived(s.marketAddr3, "8fig"),
				s.eventTransfer(s.marketAddr3, s.addr2, "8fig"),
				s.eventMessage(s.addr2),
				s.eventCoinSpent(s.marketAddr3, "1fig"),
				s.eventCoinReceived(s.feeCollectorAddr, "1fig"),
				s.eventTransfer(s.feeCollectorAddr, s.marketAddr3, "1fig"),
				s.eventMessage(s.marketAddr3),
				s.eventHoldAdded(s.addr2, "75apple,12fig", 12345),
				s.untypeEvent(&exchange.EventOrderCreated{
					OrderId: 12345, OrderType: "ask", MarketId: 3, ExternalId: "",
				}),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			td := testDef
			td.expResp = &exchange.MsgCreateAskResponse{OrderId: tc.fArgs.expOrderID}
			runMsgServerTestCase(s, td, tc)
		})
	}
}

func (s *TestSuite) TestMsgServer_CreateBid() {
	type followupArgs struct {
		expOrderID  uint64
		addr        sdk.AccAddress
		expBal      sdk.Coins
		expHoldAmt  sdk.Coins
		expSpendBal sdk.Coins
	}
	testDef := msgServerTestDef[exchange.MsgCreateBidRequest, exchange.MsgCreateBidResponse, followupArgs]{
		endpointName: "CreateBid",
		endpoint:     keeper.NewMsgServer(s.k).CreateBid,
		followup: func(_ *exchange.MsgCreateBidRequest, fargs followupArgs) {
			for _, expBal := range fargs.expBal {
				actBal := s.app.BankKeeper.GetBalance(s.ctx, fargs.addr, expBal.Denom)
				s.Assert().Equalf(expBal.String(), actBal.String(), "actual balance of %s", expBal.Denom)
			}
			holdAmt, err := s.app.HoldKeeper.GetHoldCoins(s.ctx, fargs.addr)
			if s.Assert().NoError(err, "GetHoldCoins(%s) error", s.getAddrName(fargs.addr)) {
				s.Assert().Equalf(fargs.expHoldAmt.String(), holdAmt.String(), "amount on hold for %s", s.getAddrName(fargs.addr))
			}
			actSpendBal := s.app.BankKeeper.SpendableCoins(s.ctx, fargs.addr)
			for _, expBal := range fargs.expSpendBal {
				actBal := actSpendBal.AmountOf(expBal.Denom)
				s.Assert().Equalf(expBal.Amount.String(), actBal.String(), "spendable balance of %s", expBal.Denom)
			}
		},
	}

	tests := []msgServerTestCase[exchange.MsgCreateBidRequest, followupArgs]{
		{
			name:  "invalid msg",
			setup: nil,
			msg: exchange.MsgCreateBidRequest{
				BidOrder: exchange.BidOrder{
					MarketId: 0, Buyer: s.addr1.String(), Assets: s.coin("1apple"), Price: s.coin("1peach"),
				},
			},
			expInErr: []string{invReqErr, "invalid market id: must not be zero"},
		},
		{
			name: "market does not exist",
			msg: exchange.MsgCreateBidRequest{
				BidOrder: exchange.BidOrder{
					MarketId: 7, Buyer: s.addr1.String(), Assets: s.coin("1apple"), Price: s.coin("1peach"),
				},
			},
			expInErr: []string{invReqErr, "market 7 does not exist"},
		},
		{
			name: "cannot collect creation fee",
			setup: func() {
				s.requireFundAccount(s.addr1, "9fig")
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 1, AcceptingOrders: true,
					FeeCreateAskFlat: s.coins("10fig"),
				})
			},
			msg: exchange.MsgCreateBidRequest{
				BidOrder: exchange.BidOrder{
					MarketId: 1, Buyer: s.addr1.String(), Assets: s.coin("1apple"), Price: s.coin("1peach"),
				},
				OrderCreationFee: s.coinP("10fig"),
			},
			expInErr: []string{
				invReqErr, "error collecting bid order creation fee",
				"error transferring 10fig from " + s.addr1.String() + " to market 1",
				"spendable balance 9fig is smaller than 10fig",
				"insufficient funds",
			},
		},
		{
			name: "duplicate external id",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{MarketId: 1, AcceptingOrders: true})
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(8).WithAsk(&exchange.AskOrder{
					MarketId:   1,
					Seller:     s.addr5.String(),
					Assets:     s.coin("1apple"),
					Price:      s.coin("1peach"),
					ExternalId: "dupeid",
				}))
				keeper.SetLastOrderID(store, 10)
			},
			msg: exchange.MsgCreateBidRequest{
				BidOrder: exchange.BidOrder{
					MarketId:   1,
					Buyer:      s.addr1.String(),
					Assets:     s.coin("1apple"),
					Price:      s.coin("1peach"),
					ExternalId: "dupeid",
				},
			},
			expInErr: []string{
				invReqErr, "error storing bid order",
				"external id \"dupeid\" is already in use by order 8: cannot be used for order 11",
			},
		},
		{
			name: "price not in account",
			setup: func() {
				s.requireFundAccount(s.addr1, "9peach")
				s.requireCreateMarketUnmocked(exchange.Market{MarketId: 3, AcceptingOrders: true})
			},
			msg: exchange.MsgCreateBidRequest{
				BidOrder: exchange.BidOrder{
					MarketId: 3, Buyer: s.addr1.String(), Assets: s.coin("10apple"), Price: s.coin("10peach"),
				},
			},
			expInErr: []string{
				invReqErr, "error placing hold for bid order 1",
				"account " + s.addr1.String() + " spendable balance 9peach is less than hold amount 10peach",
			},
		},
		{
			name: "settlement fee not in account",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 3, AcceptingOrders: true,
					FeeCreateAskFlat:        s.coins("10fig"),
					FeeSellerSettlementFlat: s.coins("5fig"),
				})
				s.requireFundAccount(s.addr1, "100peach,20fig")
				s.requireAddHold(s.addr1, "6fig", 0)
			},
			msg: exchange.MsgCreateBidRequest{
				BidOrder: exchange.BidOrder{
					MarketId: 3, Buyer: s.addr1.String(),
					Assets: s.coin("10apple"), Price: s.coin("100peach"),
					BuyerSettlementFees: s.coins("5fig"),
				},
				OrderCreationFee: s.coinP("10fig"),
			},
			expInErr: []string{
				invReqErr, "error placing hold for bid order 1",
				"account " + s.addr1.String() + " spendable balance 4fig is less than hold amount 5fig",
			},
		},
		{
			name: "okay: no settlement fee",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{MarketId: 2, AcceptingOrders: true})
				s.requireFundAccount(s.addr2, "100apple,100fig,100pear")
				keeper.SetLastOrderID(s.getStore(), 83)
			},
			msg: exchange.MsgCreateBidRequest{
				BidOrder: exchange.BidOrder{
					MarketId: 2, Buyer: s.addr2.String(),
					Assets: s.coin("60apple"), Price: s.coin("45pear"),
				},
			},
			fArgs: followupArgs{
				expOrderID:  84,
				addr:        s.addr2,
				expBal:      s.coins("100apple,100fig,100pear"),
				expHoldAmt:  s.coins("45pear"),
				expSpendBal: s.coins("100apple,100fig,55pear"),
			},
			expEvents: sdk.Events{
				s.eventHoldAdded(s.addr2, "45pear", 84),
				s.untypeEvent(&exchange.EventOrderCreated{
					OrderId: 84, OrderType: "bid", MarketId: 2, ExternalId: "",
				}),
			},
		},
		{
			name: "okay: with settlement fee",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 2, AcceptingOrders: true,
					FeeCreateAskFlat:        s.coins("8pear"),
					FeeSellerSettlementFlat: s.coins("12pear"),
				})
				s.requireFundAccount(s.addr2, "100apple,100fig,100pear")
				keeper.SetLastOrderID(s.getStore(), 6)
			},
			msg: exchange.MsgCreateBidRequest{
				BidOrder: exchange.BidOrder{
					MarketId: 2, Buyer: s.addr2.String(),
					Assets: s.coin("60apple"), Price: s.coin("75pear"),
					BuyerSettlementFees: s.coins("12pear"),
					ExternalId:          "some-random-id",
				},
				OrderCreationFee: s.coinP("8pear"),
			},
			fArgs: followupArgs{
				expOrderID:  7,
				addr:        s.addr2,
				expBal:      s.coins("100apple,100fig,92pear"),
				expHoldAmt:  s.coins("87pear"),
				expSpendBal: s.coins("100apple,100fig,5pear"),
			},
			expEvents: sdk.Events{
				s.eventCoinSpent(s.addr2, "8pear"),
				s.eventCoinReceived(s.marketAddr2, "8pear"),
				s.eventTransfer(s.marketAddr2, s.addr2, "8pear"),
				s.eventMessage(s.addr2),
				s.eventCoinSpent(s.marketAddr2, "1pear"),
				s.eventCoinReceived(s.feeCollectorAddr, "1pear"),
				s.eventTransfer(s.feeCollectorAddr, s.marketAddr2, "1pear"),
				s.eventMessage(s.marketAddr2),
				s.eventHoldAdded(s.addr2, "87pear", 7),
				s.untypeEvent(&exchange.EventOrderCreated{
					OrderId: 7, OrderType: "bid", MarketId: 2, ExternalId: "some-random-id",
				}),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			td := testDef
			td.expResp = &exchange.MsgCreateBidResponse{OrderId: tc.fArgs.expOrderID}
			runMsgServerTestCase(s, td, tc)
		})
	}
}

func (s *TestSuite) TestMsgServer_CancelOrder() {
	type followupArgs struct {
		addr        sdk.AccAddress
		expBal      sdk.Coins
		expHoldAmt  sdk.Coins
		expSpendBal sdk.Coins
	}
	testDef := msgServerTestDef[exchange.MsgCancelOrderRequest, exchange.MsgCancelOrderResponse, followupArgs]{
		endpointName: "CancelOrder",
		endpoint:     keeper.NewMsgServer(s.k).CancelOrder,
		expResp:      &exchange.MsgCancelOrderResponse{},
		followup: func(msg *exchange.MsgCancelOrderRequest, fargs followupArgs) {
			order, err := s.k.GetOrder(s.ctx, msg.OrderId)
			s.Assert().NoError(err, "GetOrder(%d) error", msg.OrderId)
			s.Assert().Nil(order, "GetOrder(%d) order", msg.OrderId)

			for _, expBal := range fargs.expBal {
				actBal := s.app.BankKeeper.GetBalance(s.ctx, fargs.addr, expBal.Denom)
				s.Assert().Equalf(expBal.String(), actBal.String(), "actual balance of %s", expBal.Denom)
			}

			holdAmt, err := s.app.HoldKeeper.GetHoldCoins(s.ctx, fargs.addr)
			if s.Assert().NoError(err, "GetHoldCoins(%s) error", s.getAddrName(fargs.addr)) {
				s.Assert().Equalf(fargs.expHoldAmt.String(), holdAmt.String(), "amount on hold for %s", s.getAddrName(fargs.addr))
			}

			actSpendBal := s.app.BankKeeper.SpendableCoins(s.ctx, fargs.addr)
			for _, expBal := range fargs.expSpendBal {
				actBal := actSpendBal.AmountOf(expBal.Denom)
				s.Assert().Equalf(expBal.Amount.String(), actBal.String(), "spendable balance of %s", expBal.Denom)
			}
		},
	}

	tests := []msgServerTestCase[exchange.MsgCancelOrderRequest, followupArgs]{
		{
			name:     "order 0",
			setup:    nil,
			msg:      exchange.MsgCancelOrderRequest{Signer: s.addr1.String(), OrderId: 0},
			expInErr: []string{invReqErr, "order 0 does not exist"},
		},
		{
			name: "order does not exist",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{MarketId: 3})
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(5).WithAsk(&exchange.AskOrder{
					MarketId: 3, Seller: s.addr1.String(), Assets: s.coin("1apple"), Price: s.coin("1pear"),
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(7).WithBid(&exchange.BidOrder{
					MarketId: 3, Buyer: s.addr3.String(), Assets: s.coin("1apple"), Price: s.coin("1pear"),
				}))
			},
			msg:      exchange.MsgCancelOrderRequest{Signer: s.addr2.String(), OrderId: 6},
			expInErr: []string{invReqErr, "order 6 does not exist"},
		},
		{
			name: "wrong signer",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{MarketId: 2})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(83).WithAsk(&exchange.AskOrder{
					MarketId: 2, Seller: s.addr1.String(), Assets: s.coin("1apple"), Price: s.coin("1pear"),
				}))
			},
			msg:      exchange.MsgCancelOrderRequest{Signer: s.addr2.String(), OrderId: 83},
			expInErr: []string{invReqErr, "account " + s.addr2.String() + " does not have permission to cancel order 83"},
		},
		{
			name: "market signer: ask",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 2,
					AccessGrants: []exchange.AccessGrant{
						{Address: s.addr5.String(), Permissions: []exchange.Permission{exchange.Permission_cancel}},
					},
				})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(44).WithAsk(&exchange.AskOrder{
					MarketId: 2, Seller: s.addr1.String(), Assets: s.coin("1apple"), Price: s.coin("1pear"),
				}))
				s.requireFundAccount(s.addr1, "10apple")
				s.requireAddHold(s.addr1, "2apple", 44)
			},
			msg: exchange.MsgCancelOrderRequest{Signer: s.addr5.String(), OrderId: 44},
			fArgs: followupArgs{
				addr:        s.addr1,
				expBal:      s.coins("10apple"),
				expHoldAmt:  s.coins("1apple"),
				expSpendBal: s.coins("9apple"),
			},
			expEvents: sdk.Events{
				s.eventHoldReleased(s.addr1, "1apple"),
				s.untypeEvent(&exchange.EventOrderCancelled{
					OrderId: 44, CancelledBy: s.addr5.String(), MarketId: 2, ExternalId: "",
				}),
			},
		},
		{
			name: "market signer: bid",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 2,
					AccessGrants: []exchange.AccessGrant{
						{Address: s.addr5.String(), Permissions: []exchange.Permission{exchange.Permission_cancel}},
					},
				})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(44).WithBid(&exchange.BidOrder{
					MarketId: 2, Buyer: s.addr1.String(), Assets: s.coin("1apple"), Price: s.coin("1pear"),
				}))
				s.requireFundAccount(s.addr1, "10pear")
				s.requireAddHold(s.addr1, "2pear", 44)
			},
			msg: exchange.MsgCancelOrderRequest{Signer: s.addr5.String(), OrderId: 44},
			fArgs: followupArgs{
				addr:        s.addr1,
				expBal:      s.coins("10pear"),
				expHoldAmt:  s.coins("1pear"),
				expSpendBal: s.coins("9pear"),
			},
			expEvents: sdk.Events{
				s.eventHoldReleased(s.addr1, "1pear"),
				s.untypeEvent(&exchange.EventOrderCancelled{
					OrderId: 44, CancelledBy: s.addr5.String(), MarketId: 2, ExternalId: "",
				}),
			},
		},
		{
			name: "ask with diff fee denom from price",
			setup: func() {
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(5555).WithAsk(&exchange.AskOrder{
					MarketId:                1,
					Seller:                  s.addr1.String(),
					Assets:                  s.coin("10apple"),
					Price:                   s.coin("5pear"),
					SellerSettlementFlatFee: s.coinP("1fig"),
					ExternalId:              "ext-id-5555",
				}))
				s.requireFundAccount(s.addr1, "15apple,5fig")
				s.requireAddHold(s.addr1, "10apple,1fig", 5555)
			},
			msg: exchange.MsgCancelOrderRequest{Signer: s.addr1.String(), OrderId: 5555},
			fArgs: followupArgs{
				addr:        s.addr1,
				expBal:      s.coins("15apple,5fig"),
				expHoldAmt:  nil,
				expSpendBal: s.coins("15apple,5fig"),
			},
			expEvents: sdk.Events{
				s.eventHoldReleased(s.addr1, "10apple,1fig"),
				s.untypeEvent(&exchange.EventOrderCancelled{
					OrderId: 5555, CancelledBy: s.addr1.String(), MarketId: 1, ExternalId: "ext-id-5555",
				}),
			},
		},
		{
			name: "ask with same fee denom as price",
			setup: func() {
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(98765).WithAsk(&exchange.AskOrder{
					MarketId:                3,
					Seller:                  s.addr2.String(),
					Assets:                  s.coin("10apple"),
					Price:                   s.coin("5pear"),
					SellerSettlementFlatFee: s.coinP("1pear"),
					ExternalId:              "whatever",
				}))
				s.requireFundAccount(s.addr2, "15apple,5fig")
				s.requireAddHold(s.addr2, "10apple,1fig", 98765)
			},
			msg: exchange.MsgCancelOrderRequest{Signer: s.addr2.String(), OrderId: 98765},
			fArgs: followupArgs{
				addr:        s.addr2,
				expBal:      s.coins("15apple,5fig"),
				expHoldAmt:  s.coins("1fig"),
				expSpendBal: s.coins("15apple,4fig"),
			},
			expEvents: sdk.Events{
				s.eventHoldReleased(s.addr2, "10apple"),
				s.untypeEvent(&exchange.EventOrderCancelled{
					OrderId: 98765, CancelledBy: s.addr2.String(), MarketId: 3, ExternalId: "whatever",
				}),
			},
		},
		{
			name: "bid with diff fee denom from price",
			setup: func() {
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(5555).WithBid(&exchange.BidOrder{
					MarketId:            1,
					Buyer:               s.addr1.String(),
					Assets:              s.coin("10apple"),
					Price:               s.coin("5pear"),
					BuyerSettlementFees: s.coins("1fig"),
					ExternalId:          "ext-id-5555",
				}))
				s.requireFundAccount(s.addr1, "15pear,5fig")
				s.requireAddHold(s.addr1, "5pear,1fig", 5555)
			},
			msg: exchange.MsgCancelOrderRequest{Signer: s.addr1.String(), OrderId: 5555},
			fArgs: followupArgs{
				addr:        s.addr1,
				expBal:      s.coins("15pear,5fig"),
				expHoldAmt:  nil,
				expSpendBal: s.coins("15pear,5fig"),
			},
			expEvents: sdk.Events{
				s.eventHoldReleased(s.addr1, "1fig,5pear"),
				s.untypeEvent(&exchange.EventOrderCancelled{
					OrderId: 5555, CancelledBy: s.addr1.String(), MarketId: 1, ExternalId: "ext-id-5555",
				}),
			},
		},
		{
			name: "bid with same fee denom as price",
			setup: func() {
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(98765).WithBid(&exchange.BidOrder{
					MarketId:            3,
					Buyer:               s.addr2.String(),
					Assets:              s.coin("10apple"),
					Price:               s.coin("5pear"),
					BuyerSettlementFees: s.coins("1pear"),
					ExternalId:          "whatever",
				}))
				s.requireFundAccount(s.addr2, "15pear,5fig")
				s.requireAddHold(s.addr2, "1fig,6pear", 98765)
			},
			msg: exchange.MsgCancelOrderRequest{Signer: s.addr2.String(), OrderId: 98765},
			fArgs: followupArgs{
				addr:        s.addr2,
				expBal:      s.coins("15pear,5fig"),
				expHoldAmt:  s.coins("1fig"),
				expSpendBal: s.coins("15pear,4fig"),
			},
			expEvents: sdk.Events{
				s.eventHoldReleased(s.addr2, "6pear"),
				s.untypeEvent(&exchange.EventOrderCancelled{
					OrderId: 98765, CancelledBy: s.addr2.String(), MarketId: 3, ExternalId: "whatever",
				}),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			runMsgServerTestCase(s, testDef, tc)
		})
	}
}

// TODO[1658]: func (s *TestSuite) TestMsgServer_FillBids()

// TODO[1658]: func (s *TestSuite) TestMsgServer_FillAsks()

// TODO[1658]: func (s *TestSuite) TestMsgServer_MarketSettle()

// TODO[1658]: func (s *TestSuite) TestMsgServer_MarketSetOrderExternalID()

// TODO[1658]: func (s *TestSuite) TestMsgServer_MarketWithdraw()

// TODO[1658]: func (s *TestSuite) TestMsgServer_MarketUpdateDetails()

// TODO[1658]: func (s *TestSuite) TestMsgServer_MarketUpdateEnabled()

// TODO[1658]: func (s *TestSuite) TestMsgServer_MarketUpdateUserSettle()

// TODO[1658]: func (s *TestSuite) TestMsgServer_MarketManagePermissions()

// TODO[1658]: func (s *TestSuite) TestMsgServer_MarketManageReqAttrs()

// TODO[1658]: func (s *TestSuite) TestMsgServer_GovCreateMarket()

// TODO[1658]: func (s *TestSuite) TestMsgServer_GovManageFees()

// TODO[1658]: func (s *TestSuite) TestMsgServer_GovUpdateParams()
