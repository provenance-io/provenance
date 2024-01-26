package keeper_test

import (
	"context"
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"

	"github.com/provenance-io/provenance/testutil/assertions"
	attrtypes "github.com/provenance-io/provenance/x/attribute/types"
	"github.com/provenance-io/provenance/x/exchange"
	"github.com/provenance-io/provenance/x/exchange/keeper"
	"github.com/provenance-io/provenance/x/hold"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

// All of the msg_server endpoints are merely wrappers on other keeper functions, which
// are (hopefully) extensively tested. So, in here, it's some superficial testing, but
// without the mocks so that actual interaction with the other modules can be checked.

// invReqErr is the error added by sdkerrors.ErrInvalidRequest.
const invReqErr = "invalid request"

// msgServerTestDef is the definition of a MsgServer endpoint to be tested.
// R is the request Msg type. S is the response message type.
// F is a type that holds arguments to provide to the followup function.
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

// msgServerTestCase is a test case for a MsgServer endpoint
// R is the request Msg type.
// F is a type that holds arguments to provide to the followup function.
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
// R is the request Msg type. S is the response Msg type.
// F is a type that holds arguments to provide to the td.followup function.
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
	s.logBuffer.Reset()
	var resp *S
	var err error
	testFunc := func() {
		resp, err = td.endpoint(goCtx, &tc.msg)
	}
	s.Require().NotPanicsf(testFunc, td.endpointName)
	_ = s.getLogOutput(td.endpointName)
	s.assertErrorContentsf(err, tc.expInErr, "%s error", td.endpointName)
	s.Assert().Equalf(expResp, resp, "%s response", td.endpointName)

	if len(tc.expInErr) > 0 || err != nil {
		return
	}

	actEvents := em.Events()
	s.assertEqualEvents(tc.expEvents, actEvents, "%s events", td.endpointName)

	td.followup(&tc.msg, tc.fArgs)
}

// newAttr creates a new EventAttribute with the provided key and value.
func (s *TestSuite) newAttr(key, value string) abci.EventAttribute {
	return abci.EventAttribute{Key: []byte(key), Value: []byte(value)}
}

// eventCoinSpent creates a new "coin_spent" event (emitted by the bank module).
func (s *TestSuite) eventCoinSpent(spender sdk.AccAddress, amount string) sdk.Event {
	return sdk.Event{
		Type: "coin_spent",
		Attributes: []abci.EventAttribute{
			s.newAttr("spender", spender.String()),
			s.newAttr("amount", amount),
		},
	}
}

// eventCoinReceived creates a new "coin_received" event (emitted by the bank module).
func (s *TestSuite) eventCoinReceived(receiver sdk.AccAddress, amount string) sdk.Event {
	return sdk.Event{
		Type: "coin_received",
		Attributes: []abci.EventAttribute{
			s.newAttr("receiver", receiver.String()),
			s.newAttr("amount", amount),
		},
	}
}

// eventTransfer creates a new "transfer" event (emitted by the bank module).
func (s *TestSuite) eventTransfer(recipient, sender sdk.AccAddress, amount string) sdk.Event {
	rv := sdk.Event{Type: "transfer"}
	if len(recipient) > 0 {
		rv.Attributes = append(rv.Attributes, s.newAttr("recipient", recipient.String()))
	}
	if len(sender) > 0 {
		rv.Attributes = append(rv.Attributes, s.newAttr("sender", sender.String()))
	}
	rv.Attributes = append(rv.Attributes, s.newAttr("amount", amount))
	return rv
}

// eventMessageSender creates a new "message" event with a "sender" attr (emitted by the bank module).
func (s *TestSuite) eventMessageSender(sender sdk.AccAddress) sdk.Event {
	return sdk.Event{
		Type:       "message",
		Attributes: []abci.EventAttribute{s.newAttr("sender", sender.String())},
	}
}

// eventHoldAdded creates a new event emitted when a hold is added (emitted by the hold module).
func (s *TestSuite) eventHoldAdded(addr sdk.AccAddress, amount string, orderID uint64) sdk.Event {
	return s.untypeEvent(&hold.EventHoldAdded{
		Address: addr.String(), Amount: amount, Reason: fmt.Sprintf("x/exchange: order %d", orderID),
	})
}

// eventHoldAdded creates a new event emitted when a hold is released (emitted by the hold module).
func (s *TestSuite) eventHoldReleased(addr sdk.AccAddress, amount string) sdk.Event {
	return s.untypeEvent(&hold.EventHoldReleased{Address: addr.String(), Amount: amount})
}

// requireFundAccount calls testutil.FundAccount, making sure it doesn't panic or return an error.
func (s *TestSuite) requireFundAccount(addr sdk.AccAddress, coins string) {
	assertions.RequireNotPanicsNoErrorf(s.T(), func() error {
		return testutil.FundAccount(s.app.BankKeeper, s.ctx, addr, s.coins(coins))
	}, "FundAccount(%s, %q)", s.getAddrName(addr), coins)
}

// requireAddHold calls s.app.HoldKeeper.AddHold, making sure it doesn't panic or return an error.
func (s *TestSuite) requireAddHold(addr sdk.AccAddress, holdCoins string, orderID uint64) {
	coins := s.coins(holdCoins)
	reason := fmt.Sprintf("test hold on order %d", orderID)
	assertions.RequireNotPanicsNoErrorf(s.T(), func() error {
		return s.app.HoldKeeper.AddHold(s.ctx, addr, coins, reason)
	}, "AddHold(%s, %q, %q)", s.getAddrName(addr), holdCoins, reason)
}

// requireSetNameRecord creates a name record, requiring it to not error.
func (s *TestSuite) requireSetNameRecord(name string, owner sdk.AccAddress) {
	err := s.app.NameKeeper.SetNameRecord(s.ctx, name, owner, true)
	s.Require().NoError(err, "NameKeeper.SetNameRecord(%q, %s, true)", name, s.getAddrName(owner))
}

// requireSetAttr creates an attribute with the given name on the given addr, requiring it to not error.
func (s *TestSuite) requireSetAttr(addr sdk.AccAddress, name string, owner sdk.AccAddress) {
	attr := attrtypes.Attribute{
		Name:          name,
		Value:         []byte("value of " + name),
		AttributeType: attrtypes.AttributeType_String,
		Address:       addr.String(),
	}
	err := s.app.AttributeKeeper.SetAttribute(s.ctx, attr, owner)
	s.Require().NoError(err, "SetAttribute(%s, %s)", name, s.getAddrName(owner))
}

// requireQuarantineOptIn opts an address into quarantine, requiring it to not error.
func (s *TestSuite) requireQuarantineOptIn(addr sdk.AccAddress) {
	err := s.app.QuarantineKeeper.SetOptIn(s.ctx, addr)
	s.Require().NoError(err, "QuarantineKeeper.SetOptIn(%s)", s.getAddrName(addr))
}

// requireSanctionAddress sanctions an address, requiring it to not error.
func (s *TestSuite) requireSanctionAddress(addr sdk.AccAddress) {
	err := s.app.SanctionKeeper.SanctionAddresses(s.ctx, addr)
	s.Require().NoError(err, "SanctionAddresses(%s)", s.getAddrName(addr))
}

// requireAddFinalizeAndActivateMarker creates a marker, requiring it to not error.
func (s *TestSuite) requireAddFinalizeAndActivateMarker(coin sdk.Coin, manager sdk.AccAddress, reqAttrs ...string) {
	markerAddr := s.markerAddr(coin.Denom)
	marker := &markertypes.MarkerAccount{
		BaseAccount: &authtypes.BaseAccount{Address: markerAddr.String()},
		Manager:     manager.String(),
		AccessControl: []markertypes.AccessGrant{
			{
				Address: manager.String(),
				Permissions: markertypes.AccessList{
					markertypes.Access_Mint, markertypes.Access_Burn,
					markertypes.Access_Deposit, markertypes.Access_Withdraw, markertypes.Access_Delete,
					markertypes.Access_Admin, markertypes.Access_Transfer,
				},
			},
		},
		Status:                 markertypes.StatusProposed,
		Denom:                  coin.Denom,
		Supply:                 coin.Amount,
		MarkerType:             markertypes.MarkerType_RestrictedCoin,
		SupplyFixed:            true,
		AllowGovernanceControl: true,
		AllowForcedTransfer:    true,
		RequiredAttributes:     reqAttrs,
	}
	nav := markertypes.NewNetAssetValue(s.coin("5navcoin"), 1)
	err := s.app.MarkerKeeper.SetNetAssetValue(s.ctx, marker, nav, "testing")
	s.Require().NoError(err, "SetNetAssetValue(%d)", coin.Denom)
	err = s.app.MarkerKeeper.AddFinalizeAndActivateMarker(s.ctx, marker)
	s.Require().NoError(err, "AddFinalizeAndActivateMarker(%s)", coin.Denom)
}

// expBalances is the definition of an account's expected balance, hold, and spendable.
// Only the denoms provided are checked in each type.
type expBalances struct {
	addr     sdk.AccAddress
	expBal   []sdk.Coin
	expHold  []sdk.Coin
	expSpend []sdk.Coin
}

// checkBalances looks up the actual balances and asserts that they're the same as provided.
func (s *TestSuite) checkBalances(eb expBalances) bool {
	addrName := s.getAddrName(eb.addr)
	rv := true

	for _, expBal := range eb.expBal {
		actBal := s.app.BankKeeper.GetBalance(s.ctx, eb.addr, expBal.Denom)
		rv = s.Assert().Equalf(expBal.String(), actBal.String(), "actual balance of %s for %s", expBal.Denom, addrName) && rv
	}

	for _, expHold := range eb.expHold {
		actHold, err := s.app.HoldKeeper.GetHoldCoin(s.ctx, eb.addr, expHold.Denom)
		if s.Assert().NoError(err, "GetHoldCoin(%s, %q)", addrName, expHold.Denom) {
			rv = s.Assert().Equalf(expHold.String(), actHold.String(), "amount on hold of %s for %s", expHold.Denom, addrName) && rv
		} else {
			rv = false
		}
	}

	actSpendBal := s.app.BankKeeper.SpendableCoins(s.ctx, eb.addr)
	for _, expSpend := range eb.expSpend {
		actSpend := sdk.Coin{Denom: expSpend.Denom, Amount: actSpendBal.AmountOf(expSpend.Denom)}
		rv = s.Assert().Equalf(expSpend.String(), actSpend.String(), "spendable balance of %s for %s", expSpend.Denom, addrName) && rv
	}

	return rv
}

// zeroCoin creates a coin in the given denom with a zero amount.
// Handy for putting in an expBalances to check that a denom is zero.
func (s *TestSuite) zeroCoin(denom string) sdk.Coin {
	return sdk.Coin{Denom: denom, Amount: sdkmath.ZeroInt()}
}

// zeroCoins creates a coin for each denom, each with a zero amount.
// Handy for putting in an expBalances to check that several denoms are zero.
func (s *TestSuite) zeroCoins(denoms ...string) []sdk.Coin {
	rv := make([]sdk.Coin, len(denoms))
	for i, denom := range denoms {
		rv[i] = s.zeroCoin(denom)
	}
	return rv
}

func (s *TestSuite) TestMsgServer_CreateAsk() {
	type followupArgs struct {
		expOrderID uint64
		expBal     expBalances
	}
	testDef := msgServerTestDef[exchange.MsgCreateAskRequest, exchange.MsgCreateAskResponse, followupArgs]{
		endpointName: "CreateAsk",
		endpoint:     keeper.NewMsgServer(s.k).CreateAsk,
		followup: func(_ *exchange.MsgCreateAskRequest, fargs followupArgs) {
			s.checkBalances(fargs.expBal)
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
			expInErr: []string{invReqErr, "invalid market id: cannot be zero"},
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
				expOrderID: 84,
				expBal: expBalances{
					addr:     s.addr2,
					expBal:   s.coins("100apple,100fig,100pear"),
					expHold:  []sdk.Coin{s.coin("60apple"), s.zeroCoin("fig"), s.zeroCoin("pear")},
					expSpend: s.coins("40apple,100fig,100pear"),
				},
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
				expOrderID: 7,
				expBal: expBalances{
					addr:     s.addr2,
					expBal:   s.coins("100apple,100fig,92pear"),
					expHold:  []sdk.Coin{s.coin("75apple"), s.zeroCoin("fig"), s.zeroCoin("pear")},
					expSpend: s.coins("25apple,100fig,92pear"),
				},
			},
			expEvents: sdk.Events{
				s.eventCoinSpent(s.addr2, "8pear"),
				s.eventCoinReceived(s.marketAddr2, "8pear"),
				s.eventTransfer(s.marketAddr2, s.addr2, "8pear"),
				s.eventMessageSender(s.addr2),
				s.eventCoinSpent(s.marketAddr2, "1pear"),
				s.eventCoinReceived(s.feeCollectorAddr, "1pear"),
				s.eventTransfer(s.feeCollectorAddr, s.marketAddr2, "1pear"),
				s.eventMessageSender(s.marketAddr2),
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
				expOrderID: 12345,
				expBal: expBalances{
					addr:     s.addr2,
					expBal:   s.coins("100apple,92fig,100pear"),
					expHold:  []sdk.Coin{s.coin("75apple"), s.coin("12fig"), s.zeroCoin("pear")},
					expSpend: s.coins("25apple,80fig,100pear"),
				},
			},
			expEvents: sdk.Events{
				s.eventCoinSpent(s.addr2, "8fig"),
				s.eventCoinReceived(s.marketAddr3, "8fig"),
				s.eventTransfer(s.marketAddr3, s.addr2, "8fig"),
				s.eventMessageSender(s.addr2),
				s.eventCoinSpent(s.marketAddr3, "1fig"),
				s.eventCoinReceived(s.feeCollectorAddr, "1fig"),
				s.eventTransfer(s.feeCollectorAddr, s.marketAddr3, "1fig"),
				s.eventMessageSender(s.marketAddr3),
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
		expOrderID uint64
		expBal     expBalances
	}
	testDef := msgServerTestDef[exchange.MsgCreateBidRequest, exchange.MsgCreateBidResponse, followupArgs]{
		endpointName: "CreateBid",
		endpoint:     keeper.NewMsgServer(s.k).CreateBid,
		followup: func(_ *exchange.MsgCreateBidRequest, fargs followupArgs) {
			s.checkBalances(fargs.expBal)
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
			expInErr: []string{invReqErr, "invalid market id: cannot be zero"},
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
				expOrderID: 84,
				expBal: expBalances{
					addr:     s.addr2,
					expBal:   s.coins("100apple,100fig,100pear"),
					expHold:  []sdk.Coin{s.zeroCoin("apple"), s.zeroCoin("fig"), s.coin("45pear")},
					expSpend: s.coins("100apple,100fig,55pear"),
				},
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
				expOrderID: 7,
				expBal: expBalances{
					addr:     s.addr2,
					expBal:   s.coins("100apple,100fig,92pear"),
					expHold:  []sdk.Coin{s.zeroCoin("apple"), s.zeroCoin("fig"), s.coin("87pear")},
					expSpend: s.coins("100apple,100fig,5pear"),
				},
			},
			expEvents: sdk.Events{
				s.eventCoinSpent(s.addr2, "8pear"),
				s.eventCoinReceived(s.marketAddr2, "8pear"),
				s.eventTransfer(s.marketAddr2, s.addr2, "8pear"),
				s.eventMessageSender(s.addr2),
				s.eventCoinSpent(s.marketAddr2, "1pear"),
				s.eventCoinReceived(s.feeCollectorAddr, "1pear"),
				s.eventTransfer(s.feeCollectorAddr, s.marketAddr2, "1pear"),
				s.eventMessageSender(s.marketAddr2),
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

// TODO[1789]: func (s *TestSuite) TestMsgServer_CommitFunds()

func (s *TestSuite) TestMsgServer_CancelOrder() {
	testDef := msgServerTestDef[exchange.MsgCancelOrderRequest, exchange.MsgCancelOrderResponse, expBalances]{
		endpointName: "CancelOrder",
		endpoint:     keeper.NewMsgServer(s.k).CancelOrder,
		expResp:      &exchange.MsgCancelOrderResponse{},
		followup: func(msg *exchange.MsgCancelOrderRequest, eb expBalances) {
			order, err := s.k.GetOrder(s.ctx, msg.OrderId)
			s.Assert().NoError(err, "GetOrder(%d) error", msg.OrderId)
			s.Assert().Nil(order, "GetOrder(%d) order", msg.OrderId)
			s.checkBalances(eb)
		},
	}

	tests := []msgServerTestCase[exchange.MsgCancelOrderRequest, expBalances]{
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
					MarketId:     2,
					AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_cancel)},
				})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(44).WithAsk(&exchange.AskOrder{
					MarketId: 2, Seller: s.addr1.String(), Assets: s.coin("1apple"), Price: s.coin("1pear"),
				}))
				s.requireFundAccount(s.addr1, "10apple")
				s.requireAddHold(s.addr1, "2apple", 44)
			},
			msg: exchange.MsgCancelOrderRequest{Signer: s.addr5.String(), OrderId: 44},
			fArgs: expBalances{
				addr:     s.addr1,
				expBal:   s.coins("10apple"),
				expHold:  s.coins("1apple"),
				expSpend: s.coins("9apple"),
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
					MarketId:     2,
					AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_cancel)},
				})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(44).WithBid(&exchange.BidOrder{
					MarketId: 2, Buyer: s.addr1.String(), Assets: s.coin("1apple"), Price: s.coin("1pear"),
				}))
				s.requireFundAccount(s.addr1, "10pear")
				s.requireAddHold(s.addr1, "1pear", 44)
			},
			msg: exchange.MsgCancelOrderRequest{Signer: s.addr5.String(), OrderId: 44},
			fArgs: expBalances{
				addr:     s.addr1,
				expBal:   s.coins("10pear"),
				expHold:  []sdk.Coin{s.zeroCoin("pear")},
				expSpend: s.coins("10pear"),
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
			fArgs: expBalances{
				addr:     s.addr1,
				expBal:   s.coins("15apple,5fig"),
				expHold:  s.zeroCoins("apple", "fig"),
				expSpend: s.coins("15apple,5fig"),
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
			fArgs: expBalances{
				addr:     s.addr2,
				expBal:   s.coins("15apple,5fig"),
				expHold:  []sdk.Coin{s.zeroCoin("apple"), s.coin("1fig")},
				expSpend: s.coins("15apple,4fig"),
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
			fArgs: expBalances{
				addr:     s.addr1,
				expBal:   s.coins("15pear,5fig"),
				expHold:  s.zeroCoins("pear", "fig"),
				expSpend: s.coins("15pear,5fig"),
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
				s.requireAddHold(s.addr2, "6pear", 98765)
			},
			msg: exchange.MsgCancelOrderRequest{Signer: s.addr2.String(), OrderId: 98765},
			fArgs: expBalances{
				addr:     s.addr2,
				expBal:   s.coins("15pear,5fig"),
				expHold:  s.zeroCoins("pear", "fig"),
				expSpend: s.coins("15pear,5fig"),
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

func (s *TestSuite) TestMsgServer_FillBids() {
	testDef := msgServerTestDef[exchange.MsgFillBidsRequest, exchange.MsgFillBidsResponse, []expBalances]{
		endpointName: "FillBids",
		endpoint:     keeper.NewMsgServer(s.k).FillBids,
		expResp:      &exchange.MsgFillBidsResponse{},
		followup: func(msg *exchange.MsgFillBidsRequest, ebs []expBalances) {
			for _, orderID := range msg.BidOrderIds {
				order, err := s.k.GetOrder(s.ctx, orderID)
				s.Assert().NoError(err, "GetOrder(%d) error", orderID)
				s.Assert().Nil(order, "GetOrder(%d) order", orderID)
			}

			for _, eb := range ebs {
				s.checkBalances(eb)
			}
		},
	}

	tests := []msgServerTestCase[exchange.MsgFillBidsRequest, []expBalances]{
		{
			name: "user can't create ask",
			setup: func() {
				s.requireSetNameRecord("almost.gonna.have.it", s.addr5)
				s.requireSetAttr(s.addr1, "almost.gonna.have.it", s.addr5)

				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 1, AcceptingOrders: true, AllowUserSettlement: true,
					ReqAttrCreateAsk: []string{"not.gonna.have.it"},
				})
			},
			msg: exchange.MsgFillBidsRequest{
				Seller:      s.addr1.String(),
				MarketId:    1,
				TotalAssets: s.coins("1apple"),
				BidOrderIds: []uint64{1},
			},
			expInErr: []string{invReqErr, "account " + s.addr1.String() + " is not allowed to create ask orders in market 1"},
		},
		{
			name: "one bid, both quarantined, no markers",
			setup: func() {
				s.requireFundAccount(s.addr1, "50pear")
				s.requireFundAccount(s.addr2, "10apple")

				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 3, AcceptingOrders: true, AllowUserSettlement: true,
				})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(54).WithBid(&exchange.BidOrder{
					MarketId: 3, Buyer: s.addr1.String(), Assets: s.coin("10apple"), Price: s.coin("50pear"),
				}))
				s.requireAddHold(s.addr1, "50pear", 54)

				s.requireQuarantineOptIn(s.addr1)
				s.requireQuarantineOptIn(s.addr2)
			},
			msg: exchange.MsgFillBidsRequest{
				Seller:      s.addr2.String(),
				MarketId:    3,
				TotalAssets: s.coins("10apple"),
				BidOrderIds: []uint64{54},
			},
			fArgs: []expBalances{
				{
					addr:     s.addr1,
					expBal:   []sdk.Coin{s.coin("10apple"), s.zeroCoin("pear")},
					expHold:  s.zeroCoins("apple", "pear"),
					expSpend: []sdk.Coin{s.coin("10apple"), s.zeroCoin("pear")},
				},
				{
					addr:     s.addr2,
					expBal:   []sdk.Coin{s.zeroCoin("apple"), s.coin("50pear")},
					expHold:  s.zeroCoins("apple", "pear"),
					expSpend: []sdk.Coin{s.zeroCoin("apple"), s.coin("50pear")},
				},
			},
			expEvents: sdk.Events{
				s.eventHoldReleased(s.addr1, "50pear"),
				s.eventCoinSpent(s.addr2, "10apple"),
				s.eventCoinReceived(s.addr1, "10apple"),
				s.eventTransfer(s.addr1, s.addr2, "10apple"),
				s.eventMessageSender(s.addr2),
				s.eventCoinSpent(s.addr1, "50pear"),
				s.eventCoinReceived(s.addr2, "50pear"),
				s.eventTransfer(s.addr2, s.addr1, "50pear"),
				s.eventMessageSender(s.addr1),
				s.untypeEvent(&exchange.EventOrderFilled{
					OrderId: 54, Assets: "10apple", Price: "50pear", MarketId: 3,
				}),
				s.navSetEvent("10apple", "50pear", 3),
			},
		},
		{
			name: "one bid, both quarantined, with markers",
			setup: func() {
				s.requireAddFinalizeAndActivateMarker(s.coin("10apple"), s.addr5, "got.it")
				s.requireAddFinalizeAndActivateMarker(s.coin("50pear"), s.addr5, "got.it")
				s.requireSetNameRecord("got.it", s.addr5)
				s.requireSetAttr(s.addr1, "got.it", s.addr5)
				s.requireSetAttr(s.addr2, "got.it", s.addr5)
				s.requireFundAccount(s.addr1, "50pear")
				s.requireFundAccount(s.addr2, "10apple")

				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 3, AcceptingOrders: true, AllowUserSettlement: true,
				})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(54).WithBid(&exchange.BidOrder{
					MarketId: 3, Buyer: s.addr1.String(), Assets: s.coin("10apple"), Price: s.coin("50pear"),
				}))
				s.requireAddHold(s.addr1, "50pear", 54)

				s.requireQuarantineOptIn(s.addr1)
				s.requireQuarantineOptIn(s.addr2)
			},
			msg: exchange.MsgFillBidsRequest{
				Seller:      s.addr2.String(),
				MarketId:    3,
				TotalAssets: s.coins("10apple"),
				BidOrderIds: []uint64{54},
			},
			fArgs: []expBalances{
				{
					addr:     s.addr1,
					expBal:   []sdk.Coin{s.coin("10apple"), s.zeroCoin("pear")},
					expHold:  s.zeroCoins("apple", "pear"),
					expSpend: []sdk.Coin{s.coin("10apple"), s.zeroCoin("pear")},
				},
				{
					addr:     s.addr2,
					expBal:   []sdk.Coin{s.zeroCoin("apple"), s.coin("50pear")},
					expHold:  s.zeroCoins("apple", "pear"),
					expSpend: []sdk.Coin{s.zeroCoin("apple"), s.coin("50pear")},
				},
			},
			expEvents: sdk.Events{
				s.eventHoldReleased(s.addr1, "50pear"),
				s.eventCoinSpent(s.addr2, "10apple"),
				s.eventCoinReceived(s.addr1, "10apple"),
				s.eventTransfer(s.addr1, s.addr2, "10apple"),
				s.eventMessageSender(s.addr2),
				s.eventCoinSpent(s.addr1, "50pear"),
				s.eventCoinReceived(s.addr2, "50pear"),
				s.eventTransfer(s.addr2, s.addr1, "50pear"),
				s.eventMessageSender(s.addr1),
				s.untypeEvent(&exchange.EventOrderFilled{
					OrderId: 54, Assets: "10apple", Price: "50pear", MarketId: 3,
				}),
				s.navSetEvent("10apple", "50pear", 3),
			},
		},
		{
			name: "one bid, buyer sanctioned",
			setup: func() {
				s.requireFundAccount(s.addr1, "50pear")
				s.requireFundAccount(s.addr4, "10apple")

				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 2, AcceptingOrders: true, AllowUserSettlement: true,
				})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(77).WithBid(&exchange.BidOrder{
					MarketId: 2, Buyer: s.addr1.String(), Assets: s.coin("10apple"), Price: s.coin("50pear"),
				}))
				s.requireAddHold(s.addr1, "50pear", 77)

				s.requireSanctionAddress(s.addr1)
			},
			msg: exchange.MsgFillBidsRequest{
				Seller:      s.addr4.String(),
				MarketId:    2,
				TotalAssets: s.coins("10apple"),
				BidOrderIds: []uint64{77},
			},
			expInErr: []string{invReqErr, "cannot send from " + s.addr1.String(), "account is sanctioned"},
		},
		{
			name: "one bid, seller sanctioned",
			setup: func() {
				s.requireFundAccount(s.addr1, "50pear")
				s.requireFundAccount(s.addr4, "10apple")

				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 2, AcceptingOrders: true, AllowUserSettlement: true,
				})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(77).WithBid(&exchange.BidOrder{
					MarketId: 2, Buyer: s.addr1.String(), Assets: s.coin("10apple"), Price: s.coin("50pear"),
				}))

				s.requireAddHold(s.addr1, "50pear", 77)

				s.requireSanctionAddress(s.addr4)
			},
			msg: exchange.MsgFillBidsRequest{
				Seller:      s.addr4.String(),
				MarketId:    2,
				TotalAssets: s.coins("10apple"),
				BidOrderIds: []uint64{77},
			},
			expInErr: []string{invReqErr, "cannot send from " + s.addr4.String(), "account is sanctioned"},
		},
		{
			name: "one bid, buyer does not have asset marker's req attrs",
			setup: func() {
				s.requireAddFinalizeAndActivateMarker(s.coin("10apple"), s.addr5, "not.gonna.have.it")
				s.requireSetNameRecord("not.gonna.have.it", s.addr5)
				s.requireSetAttr(s.addr1, "not.gonna.have.it", s.addr5)
				s.requireFundAccount(s.addr2, "50pear")
				s.requireFundAccount(s.addr1, "10apple")

				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 2, AcceptingOrders: true, AllowUserSettlement: true,
				})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(4).WithBid(&exchange.BidOrder{
					MarketId: 2, Buyer: s.addr2.String(), Assets: s.coin("10apple"), Price: s.coin("50pear"),
				}))
				s.requireAddHold(s.addr2, "50pear", 4)
			},
			msg: exchange.MsgFillBidsRequest{
				Seller:      s.addr1.String(),
				MarketId:    2,
				TotalAssets: s.coins("10apple"),
				BidOrderIds: []uint64{4},
			},
			expInErr: []string{invReqErr,
				"address " + s.addr2.String() + " does not contain the \"apple\" required attribute: \"not.gonna.have.it\"",
			},
		},
		{
			name: "one bid, seller does not have price marker's req attrs",
			setup: func() {
				s.requireAddFinalizeAndActivateMarker(s.coin("50pear"), s.addr5, "not.gonna.have.it")
				s.requireSetNameRecord("not.gonna.have.it", s.addr5)
				s.requireSetAttr(s.addr2, "not.gonna.have.it", s.addr5)
				s.requireFundAccount(s.addr2, "50pear")
				s.requireFundAccount(s.addr1, "10apple")

				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 2, AcceptingOrders: true, AllowUserSettlement: true,
				})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(4).WithBid(&exchange.BidOrder{
					MarketId: 2, Buyer: s.addr2.String(), Assets: s.coin("10apple"), Price: s.coin("50pear"),
				}))
				s.requireAddHold(s.addr2, "50pear", 4)
			},
			msg: exchange.MsgFillBidsRequest{
				Seller:      s.addr1.String(),
				MarketId:    2,
				TotalAssets: s.coins("10apple"),
				BidOrderIds: []uint64{4},
			},
			expInErr: []string{invReqErr,
				"address " + s.addr1.String() + " does not contain the \"pear\" required attribute: \"not.gonna.have.it\"",
			},
		},
		{
			name: "market does not have req attr for fee denom",
			setup: func() {
				s.requireAddFinalizeAndActivateMarker(s.coin("200fig"), s.addr5, "not.gonna.have.it")
				s.requireSetNameRecord("not.gonna.have.it", s.addr5)
				s.requireSetAttr(s.addr1, "not.gonna.have.it", s.addr5)
				s.requireSetAttr(s.addr2, "not.gonna.have.it", s.addr5)
				s.requireFundAccount(s.addr2, "50pear,100fig")
				s.requireFundAccount(s.addr1, "10apple,100fig")

				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 1, AcceptingOrders: true, AllowUserSettlement: true,
				})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(12345).WithBid(&exchange.BidOrder{
					MarketId: 1, Buyer: s.addr2.String(), Assets: s.coin("10apple"), Price: s.coin("50pear"),
					BuyerSettlementFees: s.coins("100fig"),
				}))
				s.requireAddHold(s.addr2, "50pear,100fig", 12345)
			},
			msg: exchange.MsgFillBidsRequest{
				Seller:                  s.addr1.String(),
				MarketId:                1,
				TotalAssets:             s.coins("10apple"),
				BidOrderIds:             []uint64{12345},
				SellerSettlementFlatFee: s.coinP("100fig"),
			},
			expInErr: []string{invReqErr,
				"address " + s.marketAddr1.String() + " does not contain the \"fig\" required attribute: \"not.gonna.have.it\"",
			},
		},
		{
			name: "okay: two bids, all req attrs and fees",
			setup: func() {
				s.requireAddFinalizeAndActivateMarker(s.coin("13apple"), s.addr5, "*.gonna.have.it")
				s.requireAddFinalizeAndActivateMarker(s.coin("70pear"), s.addr5, "*.gonna.have.it")
				s.requireAddFinalizeAndActivateMarker(s.coin("300fig"), s.addr5, "*.gonna.have.it")
				s.requireSetNameRecord("buyer.gonna.have.it", s.addr5)
				s.requireSetNameRecord("seller.gonna.have.it", s.addr5)
				s.requireSetNameRecord("market.gonna.have.it", s.addr5)
				s.requireSetAttr(s.addr1, "seller.gonna.have.it", s.addr5)
				s.requireSetAttr(s.addr2, "buyer.gonna.have.it", s.addr5)
				s.requireSetAttr(s.addr3, "buyer.gonna.have.it", s.addr5)
				s.requireSetAttr(s.marketAddr1, "market.gonna.have.it", s.addr5)
				s.requireFundAccount(s.addr1, "13apple,100fig")
				s.requireFundAccount(s.addr2, "50pear,100fig")
				s.requireFundAccount(s.addr3, "20pear,100fig")

				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 1, AcceptingOrders: true, AllowUserSettlement: true,
					FeeCreateAskFlat:          s.coins("10fig"),
					FeeCreateBidFlat:          s.coins("200fig"),
					FeeSellerSettlementFlat:   s.coins("5pear"),
					FeeSellerSettlementRatios: s.ratios("35pear:2pear"),
					FeeBuyerSettlementFlat:    s.coins("30fig"),
					FeeBuyerSettlementRatios:  s.ratios("10pear:1fig"),
					ReqAttrCreateAsk:          []string{"*.gonna.have.it"},
					ReqAttrCreateBid:          []string{"not.gonna.have.it"},
				})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(12345).WithBid(&exchange.BidOrder{
					MarketId: 1, Buyer: s.addr2.String(), Assets: s.coin("10apple"), Price: s.coin("50pear"),
					BuyerSettlementFees: s.coins("35fig"), ExternalId: "first order",
				}))
				s.requireAddHold(s.addr2, "50pear,35fig", 12345)
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(98765).WithBid(&exchange.BidOrder{
					MarketId: 1, Buyer: s.addr3.String(), Assets: s.coin("3apple"), Price: s.coin("20pear"),
					BuyerSettlementFees: s.coins("32fig"), ExternalId: "second order",
				}))
				s.requireAddHold(s.addr3, "20pear,32fig", 98765)
			},
			msg: exchange.MsgFillBidsRequest{
				Seller:                  s.addr1.String(),
				MarketId:                1,
				TotalAssets:             s.coins("13apple"),
				BidOrderIds:             []uint64{12345, 98765},
				SellerSettlementFlatFee: s.coinP("5pear"),
				AskOrderCreationFee:     s.coinP("10fig"),
			},
			fArgs: []expBalances{
				{
					addr:    s.addr1,
					expBal:  []sdk.Coin{s.zeroCoin("apple"), s.coin("61pear"), s.coin("90fig")},
					expHold: s.zeroCoins("apple", "pear", "fig"),
				},
				{
					addr:    s.addr2,
					expBal:  []sdk.Coin{s.coin("10apple"), s.zeroCoin("pear"), s.coin("65fig")},
					expHold: s.zeroCoins("apple", "pear", "fig"),
				},
				{
					addr:    s.addr3,
					expBal:  []sdk.Coin{s.coin("3apple"), s.zeroCoin("pear"), s.coin("68fig")},
					expHold: s.zeroCoins("apple", "pear", "fig"),
				},
				{
					addr:    s.marketAddr1,
					expBal:  []sdk.Coin{s.zeroCoin("apple"), s.coin("8pear"), s.coin("72fig")},
					expHold: s.zeroCoins("apple", "pear", "fig"),
				},
				{
					addr:   s.feeCollectorAddr,
					expBal: []sdk.Coin{s.zeroCoin("apple"), s.coin("1pear"), s.coin("5fig")},
				},
			},
			expEvents: sdk.Events{
				// Hold release events.
				s.eventHoldReleased(s.addr2, "35fig,50pear"),
				s.eventHoldReleased(s.addr3, "32fig,20pear"),

				// Asset transfer events.
				s.eventCoinSpent(s.addr1, "13apple"),
				s.eventMessageSender(s.addr1),
				s.eventCoinReceived(s.addr2, "10apple"),
				s.eventTransfer(s.addr2, nil, "10apple"),
				s.eventCoinReceived(s.addr3, "3apple"),
				s.eventTransfer(s.addr3, nil, "3apple"),

				// Price transfer events.
				s.eventCoinSpent(s.addr2, "50pear"),
				s.eventMessageSender(s.addr2),
				s.eventCoinSpent(s.addr3, "20pear"),
				s.eventMessageSender(s.addr3),
				s.eventCoinReceived(s.addr1, "70pear"),
				s.eventTransfer(s.addr1, nil, "70pear"),

				// Settlement fee transfer events.
				s.eventCoinSpent(s.addr2, "35fig"),
				s.eventMessageSender(s.addr2),
				s.eventCoinSpent(s.addr3, "32fig"),
				s.eventMessageSender(s.addr3),
				s.eventCoinSpent(s.addr1, "9pear"),
				s.eventMessageSender(s.addr1),
				s.eventCoinReceived(s.marketAddr1, "67fig,9pear"),
				s.eventTransfer(s.marketAddr1, nil, "67fig,9pear"),

				// Transfer of exchange portion of settlement fee.
				s.eventCoinSpent(s.marketAddr1, "4fig,1pear"),
				s.eventCoinReceived(s.feeCollectorAddr, "4fig,1pear"),
				s.eventTransfer(s.feeCollectorAddr, s.marketAddr1, "4fig,1pear"),
				s.eventMessageSender(s.marketAddr1),

				// Order filled events.
				s.untypeEvent(&exchange.EventOrderFilled{
					OrderId:    12345,
					Assets:     "10apple",
					Price:      "50pear",
					Fees:       "35fig",
					MarketId:   1,
					ExternalId: "first order",
				}),
				s.untypeEvent(&exchange.EventOrderFilled{
					OrderId:    98765,
					Assets:     "3apple",
					Price:      "20pear",
					Fees:       "32fig",
					MarketId:   1,
					ExternalId: "second order",
				}),

				// The net-asset-value event.
				s.navSetEvent("13apple", "70pear", 1),

				// Order creation fee events.
				s.eventCoinSpent(s.addr1, "10fig"),
				s.eventCoinReceived(s.marketAddr1, "10fig"),
				s.eventTransfer(s.marketAddr1, s.addr1, "10fig"),
				s.eventMessageSender(s.addr1),

				// Transfer of exchange portion of order creation fees.
				s.eventCoinSpent(s.marketAddr1, "1fig"),
				s.eventCoinReceived(s.feeCollectorAddr, "1fig"),
				s.eventTransfer(s.feeCollectorAddr, s.marketAddr1, "1fig"),
				s.eventMessageSender(s.marketAddr1),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			runMsgServerTestCase(s, testDef, tc)
		})
	}
}

func (s *TestSuite) TestMsgServer_FillAsks() {
	testDef := msgServerTestDef[exchange.MsgFillAsksRequest, exchange.MsgFillAsksResponse, []expBalances]{
		endpointName: "FillAsks",
		endpoint:     keeper.NewMsgServer(s.k).FillAsks,
		expResp:      &exchange.MsgFillAsksResponse{},
		followup: func(msg *exchange.MsgFillAsksRequest, ebs []expBalances) {
			for _, orderID := range msg.AskOrderIds {
				order, err := s.k.GetOrder(s.ctx, orderID)
				s.Assert().NoError(err, "GetOrder(%d) error", orderID)
				s.Assert().Nil(order, "GetOrder(%d) order", orderID)
			}

			for _, eb := range ebs {
				s.checkBalances(eb)
			}
		},
	}

	tests := []msgServerTestCase[exchange.MsgFillAsksRequest, []expBalances]{
		{
			name: "user can't create bid",
			setup: func() {
				s.requireSetNameRecord("almost.gonna.have.it", s.addr5)
				s.requireSetAttr(s.addr1, "almost.gonna.have.it", s.addr5)

				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 1, AcceptingOrders: true, AllowUserSettlement: true,
					ReqAttrCreateBid: []string{"not.gonna.have.it"},
				})
			},
			msg: exchange.MsgFillAsksRequest{
				Buyer:       s.addr1.String(),
				MarketId:    1,
				TotalPrice:  s.coin("1pear"),
				AskOrderIds: []uint64{1},
			},
			expInErr: []string{invReqErr, "account " + s.addr1.String() + " is not allowed to create bid orders in market 1"},
		},
		{
			name: "one ask, both quarantined, no markers",
			setup: func() {
				s.requireFundAccount(s.addr1, "50pear")
				s.requireFundAccount(s.addr2, "10apple")

				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 3, AcceptingOrders: true, AllowUserSettlement: true,
				})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(54).WithAsk(&exchange.AskOrder{
					MarketId: 3, Seller: s.addr2.String(), Assets: s.coin("10apple"), Price: s.coin("50pear"),
				}))
				s.requireAddHold(s.addr2, "10apple", 54)

				s.requireQuarantineOptIn(s.addr1)
				s.requireQuarantineOptIn(s.addr2)
			},
			msg: exchange.MsgFillAsksRequest{
				Buyer:       s.addr1.String(),
				MarketId:    3,
				TotalPrice:  s.coin("50pear"),
				AskOrderIds: []uint64{54},
			},
			fArgs: []expBalances{
				{
					addr:     s.addr1,
					expBal:   []sdk.Coin{s.coin("10apple"), s.zeroCoin("pear")},
					expHold:  s.zeroCoins("apple", "pear"),
					expSpend: []sdk.Coin{s.coin("10apple"), s.zeroCoin("pear")},
				},
				{
					addr:     s.addr2,
					expBal:   []sdk.Coin{s.zeroCoin("apple"), s.coin("50pear")},
					expHold:  s.zeroCoins("apple", "pear"),
					expSpend: []sdk.Coin{s.zeroCoin("apple"), s.coin("50pear")},
				},
			},
			expEvents: sdk.Events{
				s.eventHoldReleased(s.addr2, "10apple"),
				s.eventCoinSpent(s.addr2, "10apple"),
				s.eventCoinReceived(s.addr1, "10apple"),
				s.eventTransfer(s.addr1, s.addr2, "10apple"),
				s.eventMessageSender(s.addr2),
				s.eventCoinSpent(s.addr1, "50pear"),
				s.eventCoinReceived(s.addr2, "50pear"),
				s.eventTransfer(s.addr2, s.addr1, "50pear"),
				s.eventMessageSender(s.addr1),
				s.untypeEvent(&exchange.EventOrderFilled{
					OrderId: 54, Assets: "10apple", Price: "50pear", MarketId: 3,
				}),
				s.navSetEvent("10apple", "50pear", 3),
			},
		},
		{
			name: "one ask, both quarantined, with markers",
			setup: func() {
				s.requireAddFinalizeAndActivateMarker(s.coin("10apple"), s.addr5, "got.it")
				s.requireAddFinalizeAndActivateMarker(s.coin("50pear"), s.addr5, "got.it")
				s.requireSetNameRecord("got.it", s.addr5)
				s.requireSetAttr(s.addr1, "got.it", s.addr5)
				s.requireSetAttr(s.addr2, "got.it", s.addr5)
				s.requireFundAccount(s.addr1, "50pear")
				s.requireFundAccount(s.addr2, "10apple")

				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 3, AcceptingOrders: true, AllowUserSettlement: true,
				})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(54).WithAsk(&exchange.AskOrder{
					MarketId: 3, Seller: s.addr2.String(), Assets: s.coin("10apple"), Price: s.coin("50pear"),
				}))
				s.requireAddHold(s.addr2, "10apple", 54)

				s.requireQuarantineOptIn(s.addr1)
				s.requireQuarantineOptIn(s.addr2)
			},
			msg: exchange.MsgFillAsksRequest{
				Buyer:       s.addr1.String(),
				MarketId:    3,
				TotalPrice:  s.coin("50pear"),
				AskOrderIds: []uint64{54},
			},
			fArgs: []expBalances{
				{
					addr:     s.addr1,
					expBal:   []sdk.Coin{s.coin("10apple"), s.zeroCoin("pear")},
					expHold:  s.zeroCoins("apple", "pear"),
					expSpend: []sdk.Coin{s.coin("10apple"), s.zeroCoin("pear")},
				},
				{
					addr:     s.addr2,
					expBal:   []sdk.Coin{s.zeroCoin("apple"), s.coin("50pear")},
					expHold:  s.zeroCoins("apple", "pear"),
					expSpend: []sdk.Coin{s.zeroCoin("apple"), s.coin("50pear")},
				},
			},
			expEvents: sdk.Events{
				s.eventHoldReleased(s.addr2, "10apple"),
				s.eventCoinSpent(s.addr2, "10apple"),
				s.eventCoinReceived(s.addr1, "10apple"),
				s.eventTransfer(s.addr1, s.addr2, "10apple"),
				s.eventMessageSender(s.addr2),
				s.eventCoinSpent(s.addr1, "50pear"),
				s.eventCoinReceived(s.addr2, "50pear"),
				s.eventTransfer(s.addr2, s.addr1, "50pear"),
				s.eventMessageSender(s.addr1),
				s.untypeEvent(&exchange.EventOrderFilled{
					OrderId: 54, Assets: "10apple", Price: "50pear", MarketId: 3,
				}),
				s.navSetEvent("10apple", "50pear", 3),
			},
		},
		{
			name: "one ask, buyer sanctioned",
			setup: func() {
				s.requireFundAccount(s.addr1, "50pear")
				s.requireFundAccount(s.addr4, "10apple")

				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 2, AcceptingOrders: true, AllowUserSettlement: true,
				})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(77).WithAsk(&exchange.AskOrder{
					MarketId: 2, Seller: s.addr4.String(), Assets: s.coin("10apple"), Price: s.coin("50pear"),
				}))
				s.requireAddHold(s.addr4, "10apple", 77)

				s.requireSanctionAddress(s.addr1)
			},
			msg: exchange.MsgFillAsksRequest{
				Buyer:       s.addr1.String(),
				MarketId:    2,
				TotalPrice:  s.coin("50pear"),
				AskOrderIds: []uint64{77},
			},
			expInErr: []string{invReqErr, "cannot send from " + s.addr1.String(), "account is sanctioned"},
		},
		{
			name: "one ask, seller sanctioned",
			setup: func() {
				s.requireFundAccount(s.addr1, "50pear")
				s.requireFundAccount(s.addr4, "10apple")

				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 2, AcceptingOrders: true, AllowUserSettlement: true,
				})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(77).WithAsk(&exchange.AskOrder{
					MarketId: 2, Seller: s.addr4.String(), Assets: s.coin("10apple"), Price: s.coin("50pear"),
				}))
				s.requireAddHold(s.addr4, "10apple", 77)

				s.requireSanctionAddress(s.addr4)
			},
			msg: exchange.MsgFillAsksRequest{
				Buyer:       s.addr1.String(),
				MarketId:    2,
				TotalPrice:  s.coin("50pear"),
				AskOrderIds: []uint64{77},
			},
			expInErr: []string{invReqErr, "cannot send from " + s.addr4.String(), "account is sanctioned"},
		},
		{
			name: "one ask, buyer does not have asset marker's req attrs",
			setup: func() {
				s.requireAddFinalizeAndActivateMarker(s.coin("10apple"), s.addr5, "not.gonna.have.it")
				s.requireSetNameRecord("not.gonna.have.it", s.addr5)
				s.requireSetAttr(s.addr1, "not.gonna.have.it", s.addr5)
				s.requireFundAccount(s.addr2, "50pear")
				s.requireFundAccount(s.addr1, "10apple")

				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 2, AcceptingOrders: true, AllowUserSettlement: true,
				})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(4).WithAsk(&exchange.AskOrder{
					MarketId: 2, Seller: s.addr1.String(), Assets: s.coin("10apple"), Price: s.coin("50pear"),
				}))
				s.requireAddHold(s.addr1, "10apple", 4)
			},
			msg: exchange.MsgFillAsksRequest{
				Buyer:       s.addr2.String(),
				MarketId:    2,
				TotalPrice:  s.coin("50pear"),
				AskOrderIds: []uint64{4},
			},
			expInErr: []string{invReqErr,
				"address " + s.addr2.String() + " does not contain the \"apple\" required attribute: \"not.gonna.have.it\"",
			},
		},
		{
			name: "one ask, seller does not have price marker's req attrs",
			setup: func() {
				s.requireAddFinalizeAndActivateMarker(s.coin("50pear"), s.addr5, "not.gonna.have.it")
				s.requireSetNameRecord("not.gonna.have.it", s.addr5)
				s.requireSetAttr(s.addr2, "not.gonna.have.it", s.addr5)
				s.requireFundAccount(s.addr2, "50pear")
				s.requireFundAccount(s.addr1, "10apple")

				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 2, AcceptingOrders: true, AllowUserSettlement: true,
				})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(4).WithAsk(&exchange.AskOrder{
					MarketId: 2, Seller: s.addr1.String(), Assets: s.coin("10apple"), Price: s.coin("50pear"),
				}))
				s.requireAddHold(s.addr1, "10apple", 4)
			},
			msg: exchange.MsgFillAsksRequest{
				Buyer:       s.addr2.String(),
				MarketId:    2,
				TotalPrice:  s.coin("50pear"),
				AskOrderIds: []uint64{4},
			},
			expInErr: []string{invReqErr,
				"address " + s.addr1.String() + " does not contain the \"pear\" required attribute: \"not.gonna.have.it\"",
			},
		},
		{
			name: "market does not have req attr for fee denom",
			setup: func() {
				s.requireAddFinalizeAndActivateMarker(s.coin("200fig"), s.addr5, "not.gonna.have.it")
				s.requireSetNameRecord("not.gonna.have.it", s.addr5)
				s.requireSetAttr(s.addr1, "not.gonna.have.it", s.addr5)
				s.requireSetAttr(s.addr2, "not.gonna.have.it", s.addr5)
				s.requireFundAccount(s.addr2, "50pear,100fig")
				s.requireFundAccount(s.addr1, "10apple,100fig")

				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 1, AcceptingOrders: true, AllowUserSettlement: true,
				})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(12345).WithAsk(&exchange.AskOrder{
					MarketId: 1, Seller: s.addr1.String(), Assets: s.coin("10apple"), Price: s.coin("50pear"),
					SellerSettlementFlatFee: s.coinP("100fig"),
				}))
				s.requireAddHold(s.addr1, "10apple,100fig", 12345)
			},
			msg: exchange.MsgFillAsksRequest{
				Buyer:               s.addr2.String(),
				MarketId:            1,
				TotalPrice:          s.coin("50pear"),
				AskOrderIds:         []uint64{12345},
				BuyerSettlementFees: s.coins("100fig"),
			},
			expInErr: []string{invReqErr,
				"address " + s.marketAddr1.String() + " does not contain the \"fig\" required attribute: \"not.gonna.have.it\"",
			},
		},
		{
			name: "okay: two asks, all req attrs and fees",
			setup: func() {
				s.requireSetNameRecord("buyer.gonna.have.it", s.addr5)
				s.requireSetNameRecord("seller.gonna.have.it", s.addr5)
				s.requireSetNameRecord("market.gonna.have.it", s.addr5)
				s.requireSetAttr(s.addr1, "buyer.gonna.have.it", s.addr5)
				s.requireSetAttr(s.addr2, "seller.gonna.have.it", s.addr5)
				s.requireSetAttr(s.addr3, "seller.gonna.have.it", s.addr5)
				s.requireSetAttr(s.marketAddr1, "market.gonna.have.it", s.addr5)
				s.requireFundAccount(s.addr1, "70pear,100fig")
				s.requireFundAccount(s.addr2, "10apple,100fig")
				s.requireFundAccount(s.addr3, "3apple,100fig")
				s.requireAddFinalizeAndActivateMarker(s.coin("13apple"), s.addr5, "*.gonna.have.it")
				s.requireAddFinalizeAndActivateMarker(s.coin("70pear"), s.addr5, "*.gonna.have.it")
				s.requireAddFinalizeAndActivateMarker(s.coin("300fig"), s.addr5, "*.gonna.have.it")

				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 1, AcceptingOrders: true, AllowUserSettlement: true,
					FeeCreateAskFlat:          s.coins("200fig"),
					FeeCreateBidFlat:          s.coins("10fig"),
					FeeSellerSettlementFlat:   s.coins("5pear,12fig"),
					FeeSellerSettlementRatios: s.ratios("35pear:2pear"),
					FeeBuyerSettlementFlat:    s.coins("30fig"),
					FeeBuyerSettlementRatios:  s.ratios("10pear:1fig"),
					ReqAttrCreateAsk:          []string{"not.gonna.have.it"},
					ReqAttrCreateBid:          []string{"*.gonna.have.it"},
				})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(12345).WithAsk(&exchange.AskOrder{
					MarketId: 1, Seller: s.addr2.String(), Assets: s.coin("10apple"), Price: s.coin("50pear"),
					SellerSettlementFlatFee: s.coinP("5pear"), ExternalId: "first order",
				}))
				s.requireAddHold(s.addr2, "10apple", 12345)
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(98765).WithAsk(&exchange.AskOrder{
					MarketId: 1, Seller: s.addr3.String(), Assets: s.coin("3apple"), Price: s.coin("20pear"),
					SellerSettlementFlatFee: s.coinP("12fig"), ExternalId: "second order",
				}))
				s.requireAddHold(s.addr3, "3apple,12fig", 98765)
			},
			msg: exchange.MsgFillAsksRequest{
				Buyer:               s.addr1.String(),
				MarketId:            1,
				TotalPrice:          s.coin("70pear"),
				AskOrderIds:         []uint64{12345, 98765},
				BuyerSettlementFees: s.coins("37fig"),
				BidOrderCreationFee: s.coinP("10fig"),
			},
			fArgs: []expBalances{
				{
					addr:    s.addr1,
					expBal:  []sdk.Coin{s.coin("13apple"), s.zeroCoin("pear"), s.coin("53fig")},
					expHold: s.zeroCoins("apple", "pear", "fig"),
				},
				{
					addr:    s.addr2,
					expBal:  []sdk.Coin{s.zeroCoin("apple"), s.coin("42pear"), s.coin("100fig")},
					expHold: s.zeroCoins("apple", "pear", "fig"),
				},
				{
					addr:    s.addr3,
					expBal:  []sdk.Coin{s.zeroCoin("apple"), s.coin("18pear"), s.coin("88fig")},
					expHold: s.zeroCoins("apple", "pear", "fig"),
				},
				{
					addr:    s.marketAddr1,
					expBal:  []sdk.Coin{s.zeroCoin("apple"), s.coin("9pear"), s.coin("55fig")},
					expHold: s.zeroCoins("apple", "pear", "fig"),
				},
				{
					addr:   s.feeCollectorAddr,
					expBal: []sdk.Coin{s.zeroCoin("apple"), s.coin("1pear"), s.coin("4fig")},
				},
			},
			expEvents: sdk.Events{
				// Hold release events.
				s.eventHoldReleased(s.addr2, "10apple"),
				s.eventHoldReleased(s.addr3, "3apple,12fig"),

				// Asset transfer events.
				s.eventCoinSpent(s.addr2, "10apple"),
				s.eventMessageSender(s.addr2),
				s.eventCoinSpent(s.addr3, "3apple"),
				s.eventMessageSender(s.addr3),
				s.eventCoinReceived(s.addr1, "13apple"),
				s.eventTransfer(s.addr1, nil, "13apple"),

				// Price transfer events.
				s.eventCoinSpent(s.addr1, "70pear"),
				s.eventMessageSender(s.addr1),
				s.eventCoinReceived(s.addr2, "50pear"),
				s.eventTransfer(s.addr2, nil, "50pear"),
				s.eventCoinReceived(s.addr3, "20pear"),
				s.eventTransfer(s.addr3, nil, "20pear"),

				// Settlement fee transfer events.
				s.eventCoinSpent(s.addr2, "8pear"),
				s.eventMessageSender(s.addr2),
				s.eventCoinSpent(s.addr3, "12fig,2pear"),
				s.eventMessageSender(s.addr3),
				s.eventCoinSpent(s.addr1, "37fig"),
				s.eventMessageSender(s.addr1),
				s.eventCoinReceived(s.marketAddr1, "49fig,10pear"),
				s.eventTransfer(s.marketAddr1, nil, "49fig,10pear"),

				// Transfer of exchange portion of settlement fee.
				s.eventCoinSpent(s.marketAddr1, "3fig,1pear"),
				s.eventCoinReceived(s.feeCollectorAddr, "3fig,1pear"),
				s.eventTransfer(s.feeCollectorAddr, s.marketAddr1, "3fig,1pear"),
				s.eventMessageSender(s.marketAddr1),

				// Order filled events.
				s.untypeEvent(&exchange.EventOrderFilled{
					OrderId:    12345,
					Assets:     "10apple",
					Price:      "50pear",
					Fees:       "8pear",
					MarketId:   1,
					ExternalId: "first order",
				}),
				s.untypeEvent(&exchange.EventOrderFilled{
					OrderId:    98765,
					Assets:     "3apple",
					Price:      "20pear",
					Fees:       "12fig,2pear",
					MarketId:   1,
					ExternalId: "second order",
				}),

				// The net-asset-value event.
				s.navSetEvent("13apple", "70pear", 1),

				// Order creation fee events.
				s.eventCoinSpent(s.addr1, "10fig"),
				s.eventCoinReceived(s.marketAddr1, "10fig"),
				s.eventTransfer(s.marketAddr1, s.addr1, "10fig"),
				s.eventMessageSender(s.addr1),

				// Transfer of exchange portion of order creation fees.
				s.eventCoinSpent(s.marketAddr1, "1fig"),
				s.eventCoinReceived(s.feeCollectorAddr, "1fig"),
				s.eventTransfer(s.feeCollectorAddr, s.marketAddr1, "1fig"),
				s.eventMessageSender(s.marketAddr1),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			runMsgServerTestCase(s, testDef, tc)
		})
	}
}

func (s *TestSuite) TestMsgServer_MarketSettle() {
	type followupArgs struct {
		expBals     []expBalances
		partialLeft *exchange.Order
	}
	testDef := msgServerTestDef[exchange.MsgMarketSettleRequest, exchange.MsgMarketSettleResponse, followupArgs]{
		endpointName: "MarketSettle",
		endpoint:     keeper.NewMsgServer(s.k).MarketSettle,
		expResp:      &exchange.MsgMarketSettleResponse{},
		followup: func(msg *exchange.MsgMarketSettleRequest, fArgs followupArgs) {
			for _, orderID := range msg.AskOrderIds {
				var expOrder *exchange.Order
				if fArgs.partialLeft != nil && fArgs.partialLeft.OrderId == orderID {
					expOrder = fArgs.partialLeft
				}
				order, err := s.k.GetOrder(s.ctx, orderID)
				s.Assert().NoError(err, "GetOrder(%d) error", orderID)
				s.Assert().Equal(expOrder, order, "GetOrder(%d) order", orderID)
			}

			for _, eb := range fArgs.expBals {
				s.checkBalances(eb)
			}
		},
	}

	tests := []msgServerTestCase[exchange.MsgMarketSettleRequest, followupArgs]{
		{
			name: "admin does not have settle permission",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:     1,
					AccessGrants: []exchange.AccessGrant{s.agCanAllBut(s.addr5, exchange.Permission_settle)},
				})
			},
			msg: exchange.MsgMarketSettleRequest{
				Admin:         s.addr5.String(),
				MarketId:      1,
				AskOrderIds:   []uint64{1},
				BidOrderIds:   []uint64{2},
				ExpectPartial: false,
			},
			expInErr: []string{invReqErr,
				"account " + s.addr5.String() + " does not have permission to settle orders for market 1"},
		},
		{
			name: "an address is sanctioned",
			setup: func() {
				s.requireFundAccount(s.addr1, "7apple")
				s.requireFundAccount(s.addr2, "100pear")
				s.requireFundAccount(s.addr3, "11apple")
				s.requireFundAccount(s.addr4, "85pear")

				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 1, AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_settle)},
				})

				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
					MarketId: 1, Seller: s.addr1.String(), Assets: s.coin("7apple"), Price: s.coin("75pear"),
				}))
				s.requireAddHold(s.addr1, "7apple", 1)
				s.requireSetOrderInStore(store, exchange.NewOrder(22).WithBid(&exchange.BidOrder{
					MarketId: 1, Buyer: s.addr2.String(), Assets: s.coin("10apple"), Price: s.coin("100pear"),
				}))
				s.requireAddHold(s.addr2, "100pear", 22)
				s.requireSetOrderInStore(store, exchange.NewOrder(333).WithAsk(&exchange.AskOrder{
					MarketId: 1, Seller: s.addr3.String(), Assets: s.coin("11apple"), Price: s.coin("105pear"),
				}))
				s.requireAddHold(s.addr3, "11apple", 333)
				s.requireSetOrderInStore(store, exchange.NewOrder(4444).WithBid(&exchange.BidOrder{
					MarketId: 1, Buyer: s.addr4.String(), Assets: s.coin("8apple"), Price: s.coin("85pear"),
				}))
				s.requireAddHold(s.addr4, "85pear", 4444)

				s.requireSanctionAddress(s.addr2)
			},
			msg: exchange.MsgMarketSettleRequest{
				Admin:       s.addr5.String(),
				MarketId:    1,
				AskOrderIds: []uint64{333, 1},
				BidOrderIds: []uint64{22, 4444},
			},
			expInErr: []string{invReqErr, "cannot send from " + s.addr2.String(), "account is sanctioned"},
		},
		{
			name: "a buyer does not have asset req attr",
			setup: func() {
				s.requireAddFinalizeAndActivateMarker(s.coin("18apple"), s.addr5, "*.have.it")
				s.requireSetNameRecord("buyer.have.it", s.addr5)
				s.requireSetNameRecord("seller.have.it", s.addr5)
				s.requireSetNameRecord("doesnot-have.it", s.addr5)
				s.requireSetAttr(s.addr1, "seller.have.it", s.addr5)
				s.requireSetAttr(s.addr2, "buyer.have.it", s.addr5)
				s.requireSetAttr(s.addr3, "seller.have.it", s.addr5)
				s.requireSetAttr(s.addr4, "doesnot-have.it", s.addr5)
				s.requireFundAccount(s.addr1, "7apple")
				s.requireFundAccount(s.addr2, "100pear")
				s.requireFundAccount(s.addr3, "11apple")
				s.requireFundAccount(s.addr4, "85pear")

				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 1, AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_settle)},
				})

				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
					MarketId: 1, Seller: s.addr1.String(), Assets: s.coin("7apple"), Price: s.coin("75pear"),
				}))
				s.requireAddHold(s.addr1, "7apple", 1)
				s.requireSetOrderInStore(store, exchange.NewOrder(22).WithBid(&exchange.BidOrder{
					MarketId: 1, Buyer: s.addr2.String(), Assets: s.coin("10apple"), Price: s.coin("100pear"),
				}))
				s.requireAddHold(s.addr2, "100pear", 22)
				s.requireSetOrderInStore(store, exchange.NewOrder(333).WithAsk(&exchange.AskOrder{
					MarketId: 1, Seller: s.addr3.String(), Assets: s.coin("11apple"), Price: s.coin("105pear"),
				}))
				s.requireAddHold(s.addr3, "11apple", 333)
				s.requireSetOrderInStore(store, exchange.NewOrder(4444).WithBid(&exchange.BidOrder{
					MarketId: 1, Buyer: s.addr4.String(), Assets: s.coin("8apple"), Price: s.coin("85pear"),
				}))
				s.requireAddHold(s.addr4, "85pear", 4444)
			},
			msg: exchange.MsgMarketSettleRequest{
				Admin:       s.addr5.String(),
				MarketId:    1,
				AskOrderIds: []uint64{333, 1},
				BidOrderIds: []uint64{22, 4444},
			},
			expInErr: []string{invReqErr,
				"address " + s.addr4.String() + " does not contain the \"apple\" required attribute: \"*.have.it\""},
		},
		{
			name: "a seller does not have price req attr",
			setup: func() {
				s.requireAddFinalizeAndActivateMarker(s.coin("185pear"), s.addr5, "*.have.it")
				s.requireSetNameRecord("buyer.have.it", s.addr5)
				s.requireSetNameRecord("seller.have.it", s.addr5)
				s.requireSetNameRecord("doesnot-have.it", s.addr5)
				s.requireSetAttr(s.addr1, "doesnot-have.it", s.addr5)
				s.requireSetAttr(s.addr2, "buyer.have.it", s.addr5)
				s.requireSetAttr(s.addr3, "seller.have.it", s.addr5)
				s.requireSetAttr(s.addr4, "buyer.have.it", s.addr5)
				s.requireFundAccount(s.addr1, "7apple")
				s.requireFundAccount(s.addr2, "100pear")
				s.requireFundAccount(s.addr3, "11apple")
				s.requireFundAccount(s.addr4, "85pear")

				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 1, AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_settle)},
				})

				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
					MarketId: 1, Seller: s.addr1.String(), Assets: s.coin("7apple"), Price: s.coin("75pear"),
				}))
				s.requireAddHold(s.addr1, "7apple", 1)
				s.requireSetOrderInStore(store, exchange.NewOrder(22).WithBid(&exchange.BidOrder{
					MarketId: 1, Buyer: s.addr2.String(), Assets: s.coin("10apple"), Price: s.coin("100pear"),
				}))
				s.requireAddHold(s.addr2, "100pear", 22)
				s.requireSetOrderInStore(store, exchange.NewOrder(333).WithAsk(&exchange.AskOrder{
					MarketId: 1, Seller: s.addr3.String(), Assets: s.coin("11apple"), Price: s.coin("105pear"),
				}))
				s.requireAddHold(s.addr3, "11apple", 333)
				s.requireSetOrderInStore(store, exchange.NewOrder(4444).WithBid(&exchange.BidOrder{
					MarketId: 1, Buyer: s.addr4.String(), Assets: s.coin("8apple"), Price: s.coin("85pear"),
				}))
				s.requireAddHold(s.addr4, "85pear", 4444)
			},
			msg: exchange.MsgMarketSettleRequest{
				Admin:       s.addr5.String(),
				MarketId:    1,
				AskOrderIds: []uint64{333, 1},
				BidOrderIds: []uint64{22, 4444},
			},
			expInErr: []string{invReqErr,
				"address " + s.addr1.String() + " does not contain the \"pear\" required attribute: \"*.have.it\""},
		},
		{
			name: "all addresses quarantined",
			setup: func() {
				s.requireFundAccount(s.addr1, "7apple")
				s.requireFundAccount(s.addr2, "100pear")
				s.requireFundAccount(s.addr3, "11apple")
				s.requireFundAccount(s.addr4, "85pear")

				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 1, AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_settle)},
				})

				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
					MarketId: 1, Seller: s.addr1.String(), Assets: s.coin("7apple"), Price: s.coin("75pear"),
				}))
				s.requireAddHold(s.addr1, "7apple", 1)
				s.requireSetOrderInStore(store, exchange.NewOrder(22).WithBid(&exchange.BidOrder{
					MarketId: 1, Buyer: s.addr2.String(), Assets: s.coin("10apple"), Price: s.coin("100pear"),
				}))
				s.requireAddHold(s.addr2, "100pear", 22)
				s.requireSetOrderInStore(store, exchange.NewOrder(333).WithAsk(&exchange.AskOrder{
					MarketId: 1, Seller: s.addr3.String(), Assets: s.coin("11apple"), Price: s.coin("105pear"),
				}))
				s.requireAddHold(s.addr3, "11apple", 333)
				s.requireSetOrderInStore(store, exchange.NewOrder(4444).WithBid(&exchange.BidOrder{
					MarketId: 1, Buyer: s.addr4.String(), Assets: s.coin("8apple"), Price: s.coin("85pear"),
				}))
				s.requireAddHold(s.addr4, "85pear", 4444)

				s.requireQuarantineOptIn(s.addr1)
				s.requireQuarantineOptIn(s.addr2)
				s.requireQuarantineOptIn(s.addr3)
				s.requireQuarantineOptIn(s.addr4)
				s.requireQuarantineOptIn(s.addr5)
			},
			msg: exchange.MsgMarketSettleRequest{
				Admin:       s.addr5.String(),
				MarketId:    1,
				AskOrderIds: []uint64{1, 333},
				BidOrderIds: []uint64{4444, 22},
			},
			fArgs: followupArgs{
				expBals: []expBalances{
					{
						addr:    s.addr1,
						expBal:  []sdk.Coin{s.zeroCoin("apple"), s.coin("77pear")},
						expHold: s.zeroCoins("apple", "pear"),
					},
					{
						addr:    s.addr2,
						expBal:  []sdk.Coin{s.coin("10apple"), s.zeroCoin("pear")},
						expHold: s.zeroCoins("apple", "pear"),
					},
					{
						addr:    s.addr3,
						expBal:  []sdk.Coin{s.zeroCoin("apple"), s.coin("108pear")},
						expHold: s.zeroCoins("apple", "pear"),
					},
					{
						addr:    s.addr4,
						expBal:  []sdk.Coin{s.coin("8apple"), s.zeroCoin("pear")},
						expHold: s.zeroCoins("apple", "pear"),
					},
				},
			},
			expEvents: sdk.Events{
				// Hold releases
				s.eventHoldReleased(s.addr1, "7apple"),
				s.eventHoldReleased(s.addr3, "11apple"),
				s.eventHoldReleased(s.addr4, "85pear"),
				s.eventHoldReleased(s.addr2, "100pear"),

				// Asset transfers
				s.eventCoinSpent(s.addr1, "7apple"),
				s.eventCoinReceived(s.addr4, "7apple"),
				s.eventTransfer(s.addr4, s.addr1, "7apple"),
				s.eventMessageSender(s.addr1),
				s.eventCoinSpent(s.addr3, "11apple"),
				s.eventMessageSender(s.addr3),
				s.eventCoinReceived(s.addr4, "1apple"),
				s.eventTransfer(s.addr4, nil, "1apple"),
				s.eventCoinReceived(s.addr2, "10apple"),
				s.eventTransfer(s.addr2, nil, "10apple"),

				// Price transfers
				s.eventCoinSpent(s.addr4, "85pear"),
				s.eventMessageSender(s.addr4),
				s.eventCoinReceived(s.addr1, "75pear"),
				s.eventTransfer(s.addr1, nil, "75pear"),
				s.eventCoinReceived(s.addr3, "10pear"),
				s.eventTransfer(s.addr3, nil, "10pear"),
				s.eventCoinSpent(s.addr2, "100pear"),
				s.eventMessageSender(s.addr2),
				s.eventCoinReceived(s.addr3, "98pear"),
				s.eventTransfer(s.addr3, nil, "98pear"),
				s.eventCoinReceived(s.addr1, "2pear"),
				s.eventTransfer(s.addr1, nil, "2pear"),

				// Orders filled
				s.untypeEvent(&exchange.EventOrderFilled{
					OrderId: 1, Assets: "7apple", Price: "77pear", MarketId: 1,
				}),
				s.untypeEvent(&exchange.EventOrderFilled{
					OrderId: 333, Assets: "11apple", Price: "108pear", MarketId: 1,
				}),
				s.untypeEvent(&exchange.EventOrderFilled{
					OrderId: 4444, Assets: "8apple", Price: "85pear", MarketId: 1,
				}),
				s.untypeEvent(&exchange.EventOrderFilled{
					OrderId: 22, Assets: "10apple", Price: "100pear", MarketId: 1,
				}),

				// The net-asset-value event.
				s.navSetEvent("18apple", "185pear", 1),
			},
		},
		{
			name: "one ask, one bid, partial ask",
			setup: func() {
				s.requireFundAccount(s.addr1, "10apple")
				s.requireFundAccount(s.addr2, "75pear")

				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 3, AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_settle)},
				})

				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
					MarketId: 3, Seller: s.addr1.String(), Assets: s.coin("10apple"), Price: s.coin("100pear"),
					AllowPartial: true,
				}))
				s.requireAddHold(s.addr1, "10apple", 1)
				s.requireSetOrderInStore(store, exchange.NewOrder(22).WithBid(&exchange.BidOrder{
					MarketId: 3, Buyer: s.addr2.String(), Assets: s.coin("7apple"), Price: s.coin("75pear"),
				}))
				s.requireAddHold(s.addr2, "75pear", 22)
			},
			msg: exchange.MsgMarketSettleRequest{
				Admin:         s.addr5.String(),
				MarketId:      3,
				AskOrderIds:   []uint64{1},
				BidOrderIds:   []uint64{22},
				ExpectPartial: true,
			},
			fArgs: followupArgs{
				expBals: []expBalances{
					{
						addr:    s.addr1,
						expBal:  s.coins("3apple,75pear"),
						expHold: []sdk.Coin{s.coin("3apple"), s.zeroCoin("pear")},
					},
					{
						addr:    s.addr2,
						expBal:  []sdk.Coin{s.coin("7apple"), s.zeroCoin("pear")},
						expHold: s.zeroCoins("apple", "pear"),
					},
				},
				partialLeft: exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
					MarketId: 3, Seller: s.addr1.String(), Assets: s.coin("3apple"), Price: s.coin("30pear"),
					AllowPartial: true,
				}),
			},
			expEvents: sdk.Events{
				// Hold releases
				s.eventHoldReleased(s.addr2, "75pear"),
				s.eventHoldReleased(s.addr1, "7apple"),

				// Asset transfer
				s.eventCoinSpent(s.addr1, "7apple"),
				s.eventCoinReceived(s.addr2, "7apple"),
				s.eventTransfer(s.addr2, s.addr1, "7apple"),
				s.eventMessageSender(s.addr1),

				// Price transfer
				s.eventCoinSpent(s.addr2, "75pear"),
				s.eventCoinReceived(s.addr1, "75pear"),
				s.eventTransfer(s.addr1, s.addr2, "75pear"),
				s.eventMessageSender(s.addr2),

				// Orders filled
				s.untypeEvent(&exchange.EventOrderFilled{
					OrderId: 22, Assets: "7apple", Price: "75pear", MarketId: 3,
				}),
				// Partial fill
				s.untypeEvent(&exchange.EventOrderPartiallyFilled{
					OrderId: 1, Assets: "7apple", Price: "75pear", MarketId: 3,
				}),

				// The net-asset-value event.
				s.navSetEvent("7apple", "75pear", 3),
			},
		},
		{
			name: "one ask, one bid, partial bid",
			setup: func() {
				s.requireFundAccount(s.addr1, "7apple")
				s.requireFundAccount(s.addr2, "100pear")

				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 3, AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_settle)},
				})

				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
					MarketId: 3, Seller: s.addr1.String(), Assets: s.coin("7apple"), Price: s.coin("65pear"),
				}))
				s.requireAddHold(s.addr1, "7apple", 1)
				s.requireSetOrderInStore(store, exchange.NewOrder(22).WithBid(&exchange.BidOrder{
					MarketId: 3, Buyer: s.addr2.String(), Assets: s.coin("10apple"), Price: s.coin("100pear"),
					AllowPartial: true,
				}))
				s.requireAddHold(s.addr2, "100pear", 22)
			},
			msg: exchange.MsgMarketSettleRequest{
				Admin:         s.addr5.String(),
				MarketId:      3,
				AskOrderIds:   []uint64{1},
				BidOrderIds:   []uint64{22},
				ExpectPartial: true,
			},
			fArgs: followupArgs{
				expBals: []expBalances{
					{
						addr:    s.addr1,
						expBal:  []sdk.Coin{s.zeroCoin("apple"), s.coin("70pear")},
						expHold: s.zeroCoins("apple", "pear"),
					},
					{
						addr:    s.addr2,
						expBal:  []sdk.Coin{s.coin("7apple"), s.coin("30pear")},
						expHold: []sdk.Coin{s.zeroCoin("apple"), s.coin("30pear")},
					},
				},
				partialLeft: exchange.NewOrder(22).WithBid(&exchange.BidOrder{
					MarketId: 3, Buyer: s.addr2.String(), Assets: s.coin("3apple"), Price: s.coin("30pear"),
					AllowPartial: true,
				}),
			},
			expEvents: sdk.Events{
				// Hold releases
				s.eventHoldReleased(s.addr1, "7apple"),
				s.eventHoldReleased(s.addr2, "70pear"),

				// Asset transfer
				s.eventCoinSpent(s.addr1, "7apple"),
				s.eventCoinReceived(s.addr2, "7apple"),
				s.eventTransfer(s.addr2, s.addr1, "7apple"),
				s.eventMessageSender(s.addr1),

				// Price transfer
				s.eventCoinSpent(s.addr2, "70pear"),
				s.eventCoinReceived(s.addr1, "70pear"),
				s.eventTransfer(s.addr1, s.addr2, "70pear"),
				s.eventMessageSender(s.addr2),

				// Orders filled
				s.untypeEvent(&exchange.EventOrderFilled{
					OrderId: 1, Assets: "7apple", Price: "70pear", MarketId: 3,
				}),
				// Partial fill
				s.untypeEvent(&exchange.EventOrderPartiallyFilled{
					OrderId: 22, Assets: "7apple", Price: "70pear", MarketId: 3,
				}),

				// The net-asset-value event.
				s.navSetEvent("7apple", "70pear", 3),
			},
		},
		{
			name: "two of each with fees and req attrs",
			setup: func() {
				s.requireAddFinalizeAndActivateMarker(s.coin("185pear"), s.addr5, "*.have.it")
				s.requireAddFinalizeAndActivateMarker(s.coin("18apple"), s.addr5, "*.have.it")
				s.requireSetNameRecord("buyer.have.it", s.addr5)
				s.requireSetNameRecord("seller.have.it", s.addr5)
				s.requireSetNameRecord("market.have.it", s.addr5)
				s.requireSetAttr(s.addr1, "seller.have.it", s.addr5)
				s.requireSetAttr(s.addr2, "buyer.have.it", s.addr5)
				s.requireSetAttr(s.addr3, "seller.have.it", s.addr5)
				s.requireSetAttr(s.addr4, "buyer.have.it", s.addr5)
				s.requireSetAttr(s.marketAddr2, "market.have.it", s.addr5)
				s.requireFundAccount(s.addr1, "20apple,100pear,100fig")
				s.requireFundAccount(s.addr2, "20apple,100pear,100fig")
				s.requireFundAccount(s.addr3, "20apple,100pear,100fig")
				s.requireFundAccount(s.addr4, "20apple,100pear,100fig")

				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 2, AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_settle)},
					FeeSellerSettlementRatios: s.ratios("10pear:1pear"),
				})

				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
					MarketId: 2, Seller: s.addr1.String(), Assets: s.coin("7apple"), Price: s.coin("75pear"),
					SellerSettlementFlatFee: s.coinP("10fig"),
				}))
				s.requireAddHold(s.addr1, "7apple,10fig", 1)
				s.requireSetOrderInStore(store, exchange.NewOrder(22).WithBid(&exchange.BidOrder{
					MarketId: 2, Buyer: s.addr2.String(), Assets: s.coin("10apple"), Price: s.coin("100pear"),
					BuyerSettlementFees: s.coins("20fig"),
				}))
				s.requireAddHold(s.addr2, "100pear,20fig", 22)
				s.requireSetOrderInStore(store, exchange.NewOrder(333).WithAsk(&exchange.AskOrder{
					MarketId: 2, Seller: s.addr3.String(), Assets: s.coin("11apple"), Price: s.coin("105pear"),
					SellerSettlementFlatFee: s.coinP("5pear"),
				}))
				s.requireAddHold(s.addr3, "11apple", 333)
				s.requireSetOrderInStore(store, exchange.NewOrder(4444).WithBid(&exchange.BidOrder{
					MarketId: 2, Buyer: s.addr4.String(), Assets: s.coin("8apple"), Price: s.coin("85pear"),
					BuyerSettlementFees: s.coins("10pear"),
				}))
				s.requireAddHold(s.addr4, "95pear", 4444)
			},
			msg: exchange.MsgMarketSettleRequest{
				Admin:       s.addr5.String(),
				MarketId:    2,
				AskOrderIds: []uint64{1, 333},
				BidOrderIds: []uint64{22, 4444},
			},
			fArgs: followupArgs{
				expBals: []expBalances{
					{
						addr:    s.addr1,
						expBal:  s.coins("13apple,169pear,90fig"),
						expHold: s.zeroCoins("apple", "pear", "fig"),
					},
					{
						addr:    s.addr2,
						expBal:  []sdk.Coin{s.coin("30apple"), s.zeroCoin("pear"), s.coin("80fig")},
						expHold: s.zeroCoins("apple", "pear", "fig"),
					},
					{
						addr:    s.addr3,
						expBal:  s.coins("9apple,192pear,100fig"),
						expHold: s.zeroCoins("apple", "pear", "fig"),
					},
					{
						addr:    s.addr4,
						expBal:  s.coins("28apple,5pear,100fig"),
						expHold: s.zeroCoins("apple", "pear", "fig"),
					},
					{
						addr:   s.marketAddr2,
						expBal: []sdk.Coin{s.zeroCoin("apple"), s.coin("32pear"), s.coin("28fig")},
					},
					{
						addr:   s.feeCollectorAddr,
						expBal: []sdk.Coin{s.zeroCoin("apple"), s.coin("2pear"), s.coin("2fig")},
					},
				},
			},
			expEvents: sdk.Events{
				// Hold releases
				s.eventHoldReleased(s.addr1, "7apple,10fig"),
				s.eventHoldReleased(s.addr3, "11apple"),
				s.eventHoldReleased(s.addr2, "20fig,100pear"),
				s.eventHoldReleased(s.addr4, "95pear"),

				// Asset transfers
				s.eventCoinSpent(s.addr1, "7apple"),
				s.eventCoinReceived(s.addr2, "7apple"),
				s.eventTransfer(s.addr2, s.addr1, "7apple"),
				s.eventMessageSender(s.addr1),
				s.eventCoinSpent(s.addr3, "11apple"),
				s.eventMessageSender(s.addr3),
				s.eventCoinReceived(s.addr2, "3apple"),
				s.eventTransfer(s.addr2, nil, "3apple"),
				s.eventCoinReceived(s.addr4, "8apple"),
				s.eventTransfer(s.addr4, nil, "8apple"),

				// Price transfers
				s.eventCoinSpent(s.addr2, "100pear"),
				s.eventMessageSender(s.addr2),
				s.eventCoinReceived(s.addr1, "75pear"),
				s.eventTransfer(s.addr1, nil, "75pear"),
				s.eventCoinReceived(s.addr3, "25pear"),
				s.eventTransfer(s.addr3, nil, "25pear"),
				s.eventCoinSpent(s.addr4, "85pear"),
				s.eventMessageSender(s.addr4),
				s.eventCoinReceived(s.addr3, "83pear"),
				s.eventTransfer(s.addr3, nil, "83pear"),
				s.eventCoinReceived(s.addr1, "2pear"),
				s.eventTransfer(s.addr1, nil, "2pear"),

				// Fee transfers to market
				s.eventCoinSpent(s.addr1, "10fig,8pear"),
				s.eventMessageSender(s.addr1),
				s.eventCoinSpent(s.addr3, "16pear"),
				s.eventMessageSender(s.addr3),
				s.eventCoinSpent(s.addr2, "20fig"),
				s.eventMessageSender(s.addr2),
				s.eventCoinSpent(s.addr4, "10pear"),
				s.eventMessageSender(s.addr4),
				s.eventCoinReceived(s.marketAddr2, "30fig,34pear"),
				s.eventTransfer(s.marketAddr2, nil, "30fig,34pear"),

				// Transfers of exchange portion of fees
				s.eventCoinSpent(s.marketAddr2, "2fig,2pear"),
				s.eventCoinReceived(s.feeCollectorAddr, "2fig,2pear"),
				s.eventTransfer(s.feeCollectorAddr, s.marketAddr2, "2fig,2pear"),
				s.eventMessageSender(s.marketAddr2),

				// Orders filled
				s.untypeEvent(&exchange.EventOrderFilled{
					OrderId: 1, Assets: "7apple", Price: "77pear", MarketId: 2, Fees: "10fig,8pear",
				}),
				s.untypeEvent(&exchange.EventOrderFilled{
					OrderId: 333, Assets: "11apple", Price: "108pear", MarketId: 2, Fees: "16pear",
				}),
				s.untypeEvent(&exchange.EventOrderFilled{
					OrderId: 22, Assets: "10apple", Price: "100pear", MarketId: 2, Fees: "20fig",
				}),
				s.untypeEvent(&exchange.EventOrderFilled{
					OrderId: 4444, Assets: "8apple", Price: "85pear", MarketId: 2, Fees: "10pear",
				}),

				// The net-asset-value event.
				s.navSetEvent("18apple", "185pear", 2),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			runMsgServerTestCase(s, testDef, tc)
		})
	}
}

// TODO[1789]: func (s *TestSuite) TestMsgServer_MarketCommitmentSettle()

// TODO[1789]: func (s *TestSuite) TestMsgServer_MarketReleaseCommitments()

func (s *TestSuite) TestMsgServer_MarketSetOrderExternalID() {
	type followupArgs struct{}
	testDef := msgServerTestDef[exchange.MsgMarketSetOrderExternalIDRequest, exchange.MsgMarketSetOrderExternalIDResponse, followupArgs]{
		endpointName: "MarketSetOrderExternalID",
		endpoint:     keeper.NewMsgServer(s.k).MarketSetOrderExternalID,
		expResp:      &exchange.MsgMarketSetOrderExternalIDResponse{},
		followup: func(msg *exchange.MsgMarketSetOrderExternalIDRequest, _ followupArgs) {
			order, err := s.k.GetOrder(s.ctx, msg.OrderId)
			s.Assert().NoError(err, "GetOrder(%d) error", msg.OrderId)
			if s.Assert().NotNil(order, "GetOrder(%d) order", msg.OrderId) {
				s.Assert().Equal(msg.ExternalId, order.GetExternalID(), "GetOrder(%d) order ExternalID", msg.OrderId)
			}
		},
	}

	tests := []msgServerTestCase[exchange.MsgMarketSetOrderExternalIDRequest, followupArgs]{
		{
			name: "admin does not have permission",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:     1,
					AccessGrants: []exchange.AccessGrant{s.agCanAllBut(s.addr5, exchange.Permission_set_ids)},
				})
			},
			msg: exchange.MsgMarketSetOrderExternalIDRequest{
				Admin: s.addr5.String(), MarketId: 1, OrderId: 1, ExternalId: "bananas",
			},
			expInErr: []string{invReqErr,
				"account " + s.addr5.String() + " does not have permission to set external ids on orders for market 1"},
		},
		{
			name: "order does not exist",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 1, AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_set_ids)},
				})
			},
			msg: exchange.MsgMarketSetOrderExternalIDRequest{
				Admin: s.addr5.String(), MarketId: 1, OrderId: 1, ExternalId: "bananas",
			},
			expInErr: []string{invReqErr, "order 1 not found"},
		},
		{
			name: "okay: nothing to something",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 1, AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_set_ids)},
				})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(7).WithAsk(&exchange.AskOrder{
					MarketId: 1, Seller: s.addr1.String(), Assets: s.coin("1apple"), Price: s.coin("1pear"),
					ExternalId: "",
				}))
			},
			msg: exchange.MsgMarketSetOrderExternalIDRequest{
				Admin: s.addr5.String(), MarketId: 1, OrderId: 7, ExternalId: "bananas",
			},
			expEvents: sdk.Events{
				s.untypeEvent(&exchange.EventOrderExternalIDUpdated{
					OrderId:    7,
					MarketId:   1,
					ExternalId: "bananas",
				}),
			},
		},
		{
			name: "okay: something to something else",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 1, AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_set_ids)},
				})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(7).WithBid(&exchange.BidOrder{
					MarketId: 1, Buyer: s.addr1.String(), Assets: s.coin("1apple"), Price: s.coin("1pear"),
					ExternalId: "something",
				}))
			},
			msg: exchange.MsgMarketSetOrderExternalIDRequest{
				Admin: s.addr5.String(), MarketId: 1, OrderId: 7, ExternalId: "bananas",
			},
			expEvents: sdk.Events{
				s.untypeEvent(&exchange.EventOrderExternalIDUpdated{
					OrderId:    7,
					MarketId:   1,
					ExternalId: "bananas",
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

func (s *TestSuite) TestMsgServer_MarketWithdraw() {
	testDef := msgServerTestDef[exchange.MsgMarketWithdrawRequest, exchange.MsgMarketWithdrawResponse, []expBalances]{
		endpointName: "MarketWithdraw",
		endpoint:     keeper.NewMsgServer(s.k).MarketWithdraw,
		expResp:      &exchange.MsgMarketWithdrawResponse{},
		followup: func(_ *exchange.MsgMarketWithdrawRequest, fArgs []expBalances) {
			for _, eb := range fArgs {
				s.checkBalances(eb)
			}
		},
	}

	tests := []msgServerTestCase[exchange.MsgMarketWithdrawRequest, []expBalances]{
		{
			name: "admin does not have permission to withdraw",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:     1,
					AccessGrants: []exchange.AccessGrant{s.agCanAllBut(s.addr5, exchange.Permission_withdraw)},
				})
			},
			msg: exchange.MsgMarketWithdrawRequest{
				Admin: s.addr5.String(), MarketId: 1, ToAddress: s.addr1.String(), Amount: s.coins("100fig"),
			},
			expInErr: []string{invReqErr,
				"account " + s.addr5.String() + " does not have permission to withdraw from market 1"},
		},
		{
			name: "insufficient funds in market",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:     1,
					AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_withdraw)},
				})
				s.requireFundAccount(s.marketAddr1, "100apple,99pear,100fig")
			},
			msg: exchange.MsgMarketWithdrawRequest{
				Admin: s.addr5.String(), MarketId: 1, ToAddress: s.addr1.String(), Amount: s.coins("3apple,100pear,50fig"),
			},
			expInErr: []string{invReqErr, "spendable balance 99pear is smaller than 100pear", "insufficient funds"},
		},
		{
			name: "destination does not have req attrs",
			setup: func() {
				s.requireAddFinalizeAndActivateMarker(s.coin("105apple"), s.addr5, "*.apple.what.what")
				s.requireAddFinalizeAndActivateMarker(s.coin("105pear"), s.addr5, "*.pear.what.what")
				s.requireSetNameRecord("nut.apple.what.what", s.addr5)
				s.requireSetNameRecord("nut.pear.what.what", s.addr5)
				s.requireSetAttr(s.marketAddr1, "nut.apple.what.what", s.addr5)
				s.requireSetAttr(s.marketAddr1, "nut.pear.what.what", s.addr5)
				s.requireSetAttr(s.addr1, "nut.apple.what.what", s.addr5)

				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:     1,
					AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_withdraw)},
				})
				s.requireFundAccount(s.marketAddr1, "100apple,100pear,100fig")
			},
			msg: exchange.MsgMarketWithdrawRequest{
				Admin: s.addr5.String(), MarketId: 1, ToAddress: s.addr1.String(), Amount: s.coins("3apple,100pear,50fig"),
			},
			expInErr: []string{invReqErr, "failed to withdraw 3apple,50fig,100pear from market 1",
				"address " + s.addr1.String() + " does not contain the \"pear\" required attribute: \"*.pear.what.what\""},
		},
		{
			name: "okay",
			setup: func() {
				s.requireAddFinalizeAndActivateMarker(s.coin("100apple"), s.addr5, "*.apple.what.what")
				s.requireAddFinalizeAndActivateMarker(s.coin("100pear"), s.addr5, "*.pear.what.what")
				s.requireSetNameRecord("nut.apple.what.what", s.addr5)
				s.requireSetNameRecord("nut.pear.what.what", s.addr5)
				s.requireSetAttr(s.marketAddr1, "nut.apple.what.what", s.addr5)
				s.requireSetAttr(s.marketAddr1, "nut.pear.what.what", s.addr5)
				s.requireSetAttr(s.addr1, "nut.apple.what.what", s.addr5)
				s.requireSetAttr(s.addr1, "nut.pear.what.what", s.addr5)

				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:     1,
					AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_withdraw)},
				})
				s.requireFundAccount(s.marketAddr1, "100apple,100pear,100fig")
				s.requireFundAccount(s.addr1, "5apple,5pear")
			},
			msg: exchange.MsgMarketWithdrawRequest{
				Admin: s.addr5.String(), MarketId: 1, ToAddress: s.addr1.String(), Amount: s.coins("3apple,100pear,50fig"),
			},
			fArgs: []expBalances{
				{
					addr:   s.marketAddr1,
					expBal: []sdk.Coin{s.coin("97apple"), s.zeroCoin("pear"), s.coin("50fig")},
				},
				{
					addr:   s.addr1,
					expBal: s.coins("8apple,105pear,50fig"),
				},
			},
			expEvents: sdk.Events{
				s.eventCoinSpent(s.marketAddr1, "3apple,50fig,100pear"),
				s.eventCoinReceived(s.addr1, "3apple,50fig,100pear"),
				s.eventTransfer(s.addr1, s.marketAddr1, "3apple,50fig,100pear"),
				s.eventMessageSender(s.marketAddr1),
				s.untypeEvent(&exchange.EventMarketWithdraw{
					MarketId:    1,
					Amount:      "3apple,50fig,100pear",
					Destination: s.addr1.String(),
					WithdrawnBy: s.addr5.String(),
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

func (s *TestSuite) TestMsgServer_MarketUpdateDetails() {
	testDef := msgServerTestDef[exchange.MsgMarketUpdateDetailsRequest, exchange.MsgMarketUpdateDetailsResponse, struct{}]{
		endpointName: "MarketUpdateDetails",
		endpoint:     keeper.NewMsgServer(s.k).MarketUpdateDetails,
		expResp:      &exchange.MsgMarketUpdateDetailsResponse{},
		followup: func(msg *exchange.MsgMarketUpdateDetailsRequest, _ struct{}) {
			market := s.k.GetMarket(s.ctx, msg.MarketId)
			if s.Assert().NotNil(market, "GetMarket(%d)", msg.MarketId) {
				s.Assert().Equal(msg.MarketDetails, market.MarketDetails, "market %d details", msg.MarketId)
			}
		},
	}

	tests := []msgServerTestCase[exchange.MsgMarketUpdateDetailsRequest, struct{}]{
		{
			name: "admin does not have permission to update market",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:     2,
					AccessGrants: []exchange.AccessGrant{s.agCanAllBut(s.addr5, exchange.Permission_update)},
				})
			},
			msg: exchange.MsgMarketUpdateDetailsRequest{
				Admin:         s.addr5.String(),
				MarketId:      2,
				MarketDetails: exchange.MarketDetails{Name: "new name"},
			},
			expInErr: []string{invReqErr,
				"account " + s.addr5.String() + " does not have permission to update market 2"},
		},
		{
			name: "error updating details",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 2, AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_update)},
				})
				ma := s.k.GetMarketAccount(s.ctx, 2)
				s.app.AccountKeeper.SetAccount(s.ctx, ma.BaseAccount)
			},
			msg: exchange.MsgMarketUpdateDetailsRequest{
				Admin:         s.addr5.String(),
				MarketId:      2,
				MarketDetails: exchange.MarketDetails{Name: "new name"},
			},
			expInErr: []string{invReqErr, "market 2 account not found"},
		},
		{
			name: "all good",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 2, AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_update)},
					MarketDetails: exchange.MarketDetails{
						Name:        "Market 2 Old Name",
						Description: "The old description of market 2.",
						WebsiteUrl:  "http://example.com/old/market/2",
						IconUri:     "http://oops.example.com/old/market/2",
					},
				})
			},
			msg: exchange.MsgMarketUpdateDetailsRequest{
				Admin:    s.addr5.String(),
				MarketId: 2,
				MarketDetails: exchange.MarketDetails{
					Name:        "Market Two",
					Description: "This is the new, better, stronger description of Market Two!",
					WebsiteUrl:  "http://example.com/new/market/2",
					IconUri:     "http://example.com/new/market/2/icon",
				},
			},
			expEvents: sdk.Events{
				s.untypeEvent(&exchange.EventMarketDetailsUpdated{MarketId: 2, UpdatedBy: s.addr5.String()}),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			runMsgServerTestCase(s, testDef, tc)
		})
	}
}

func (s *TestSuite) TestMsgServer_MarketUpdateEnabled() {
	testDef := msgServerTestDef[exchange.MsgMarketUpdateEnabledRequest, exchange.MsgMarketUpdateEnabledResponse, struct{}]{
		endpointName: "MarketUpdateEnabled",
		endpoint:     keeper.NewMsgServer(s.k).MarketUpdateEnabled,
		expResp:      &exchange.MsgMarketUpdateEnabledResponse{},
	}

	tc := msgServerTestCase[exchange.MsgMarketUpdateEnabledRequest, struct{}]{
		name:     "always error",
		msg:      exchange.MsgMarketUpdateEnabledRequest{},
		expInErr: []string{"the MarketUpdateEnabled endpoint has been replaced by the MarketUpdateAcceptingOrders endpoint"},
	}

	runMsgServerTestCase(s, testDef, tc)
}

func (s *TestSuite) TestMsgServer_MarketUpdateAcceptingOrders() {
	testDef := msgServerTestDef[exchange.MsgMarketUpdateAcceptingOrdersRequest, exchange.MsgMarketUpdateAcceptingOrdersResponse, struct{}]{
		endpointName: "MarketUpdateAcceptingOrders",
		endpoint:     keeper.NewMsgServer(s.k).MarketUpdateAcceptingOrders,
		expResp:      &exchange.MsgMarketUpdateAcceptingOrdersResponse{},
		followup: func(msg *exchange.MsgMarketUpdateAcceptingOrdersRequest, _ struct{}) {
			isEnabled := s.k.IsMarketAcceptingOrders(s.ctx, msg.MarketId)
			s.Assert().Equal(msg.AcceptingOrders, isEnabled, "IsMarketAcceptingOrders(%d)", msg.MarketId)
		},
	}

	tests := []msgServerTestCase[exchange.MsgMarketUpdateAcceptingOrdersRequest, struct{}]{
		{
			name: "admin does not have permission to update market",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:     3,
					AccessGrants: []exchange.AccessGrant{s.agCanAllBut(s.addr5, exchange.Permission_update)},
				})
			},
			msg: exchange.MsgMarketUpdateAcceptingOrdersRequest{
				Admin:           s.addr5.String(),
				MarketId:        3,
				AcceptingOrders: true,
			},
			expInErr: []string{invReqErr,
				"account " + s.addr5.String() + " does not have permission to update market 3"},
		},
		{
			name: "false to false",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 3, AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_update)},
					AcceptingOrders: false,
				})
			},
			msg: exchange.MsgMarketUpdateAcceptingOrdersRequest{
				Admin:           s.addr5.String(),
				MarketId:        3,
				AcceptingOrders: false,
			},
			expInErr: []string{invReqErr, "market 3 already has accepting-orders false"},
		},
		{
			name: "true to true",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 3, AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_update)},
					AcceptingOrders: true,
				})
			},
			msg: exchange.MsgMarketUpdateAcceptingOrdersRequest{
				Admin:           s.addr5.String(),
				MarketId:        3,
				AcceptingOrders: true,
			},
			expInErr: []string{invReqErr, "market 3 already has accepting-orders true"},
		},
		{
			name: "false to true",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 3, AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_update)},
					AcceptingOrders: false,
				})
			},
			msg: exchange.MsgMarketUpdateAcceptingOrdersRequest{
				Admin:           s.addr5.String(),
				MarketId:        3,
				AcceptingOrders: true,
			},
			expEvents: sdk.Events{
				s.untypeEvent(&exchange.EventMarketOrdersEnabled{MarketId: 3, UpdatedBy: s.addr5.String()}),
			},
		},
		{
			name: "true to false",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 3, AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_update)},
					AcceptingOrders: true,
				})
			},
			msg: exchange.MsgMarketUpdateAcceptingOrdersRequest{
				Admin:           s.addr5.String(),
				MarketId:        3,
				AcceptingOrders: false,
			},
			expEvents: sdk.Events{
				s.untypeEvent(&exchange.EventMarketOrdersDisabled{MarketId: 3, UpdatedBy: s.addr5.String()}),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			runMsgServerTestCase(s, testDef, tc)
		})
	}
}

func (s *TestSuite) TestMsgServer_MarketUpdateUserSettle() {
	testDef := msgServerTestDef[exchange.MsgMarketUpdateUserSettleRequest, exchange.MsgMarketUpdateUserSettleResponse, struct{}]{
		endpointName: "MarketUpdateUserSettle",
		endpoint:     keeper.NewMsgServer(s.k).MarketUpdateUserSettle,
		expResp:      &exchange.MsgMarketUpdateUserSettleResponse{},
		followup: func(msg *exchange.MsgMarketUpdateUserSettleRequest, _ struct{}) {
			allowed := s.k.IsUserSettlementAllowed(s.ctx, msg.MarketId)
			s.Assert().Equal(msg.AllowUserSettlement, allowed, "IsUserSettlementAllowed(%d)", msg.MarketId)
		},
	}

	tests := []msgServerTestCase[exchange.MsgMarketUpdateUserSettleRequest, struct{}]{
		{
			name: "admin does not have permission to update market",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:     3,
					AccessGrants: []exchange.AccessGrant{s.agCanAllBut(s.addr5, exchange.Permission_update)},
				})
			},
			msg: exchange.MsgMarketUpdateUserSettleRequest{
				Admin:               s.addr5.String(),
				MarketId:            3,
				AllowUserSettlement: true,
			},
			expInErr: []string{invReqErr,
				"account " + s.addr5.String() + " does not have permission to update market 3"},
		},
		{
			name: "false to false",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 3, AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_update)},
					AllowUserSettlement: false,
				})
			},
			msg: exchange.MsgMarketUpdateUserSettleRequest{
				Admin:               s.addr5.String(),
				MarketId:            3,
				AllowUserSettlement: false,
			},
			expInErr: []string{invReqErr, "market 3 already has allow-user-settlement false"},
		},
		{
			name: "true to true",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 3, AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_update)},
					AllowUserSettlement: true,
				})
			},
			msg: exchange.MsgMarketUpdateUserSettleRequest{
				Admin:               s.addr5.String(),
				MarketId:            3,
				AllowUserSettlement: true,
			},
			expInErr: []string{invReqErr, "market 3 already has allow-user-settlement true"},
		},
		{
			name: "false to true",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 3, AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_update)},
					AllowUserSettlement: false,
				})
			},
			msg: exchange.MsgMarketUpdateUserSettleRequest{
				Admin:               s.addr5.String(),
				MarketId:            3,
				AllowUserSettlement: true,
			},
			expEvents: sdk.Events{
				s.untypeEvent(&exchange.EventMarketUserSettleEnabled{MarketId: 3, UpdatedBy: s.addr5.String()}),
			},
		},
		{
			name: "true to false",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 3, AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_update)},
					AllowUserSettlement: true,
				})
			},
			msg: exchange.MsgMarketUpdateUserSettleRequest{
				Admin:               s.addr5.String(),
				MarketId:            3,
				AllowUserSettlement: false,
			},
			expEvents: sdk.Events{
				s.untypeEvent(&exchange.EventMarketUserSettleDisabled{MarketId: 3, UpdatedBy: s.addr5.String()}),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			runMsgServerTestCase(s, testDef, tc)
		})
	}
}

func (s *TestSuite) TestMsgServer_MarketUpdateAcceptingCommitments() {
	testDef := msgServerTestDef[exchange.MsgMarketUpdateAcceptingCommitmentsRequest, exchange.MsgMarketUpdateAcceptingCommitmentsResponse, struct{}]{
		endpointName: "MarketUpdateAcceptingCommitments",
		endpoint:     keeper.NewMsgServer(s.k).MarketUpdateAcceptingCommitments,
		expResp:      &exchange.MsgMarketUpdateAcceptingCommitmentsResponse{},
		followup: func(msg *exchange.MsgMarketUpdateAcceptingCommitmentsRequest, _ struct{}) {
			isEnabled := s.k.IsMarketAcceptingCommitments(s.ctx, msg.MarketId)
			s.Assert().Equal(msg.AcceptingCommitments, isEnabled, "IsMarketAcceptingCommitments(%d)", msg.MarketId)
		},
	}

	tests := []msgServerTestCase[exchange.MsgMarketUpdateAcceptingCommitmentsRequest, struct{}]{
		{
			name: "admin does not have permission to update market",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:     3,
					AccessGrants: []exchange.AccessGrant{s.agCanAllBut(s.addr5, exchange.Permission_update)},
				})
			},
			msg: exchange.MsgMarketUpdateAcceptingCommitmentsRequest{
				Admin:                s.addr5.String(),
				MarketId:             3,
				AcceptingCommitments: true,
			},
			expInErr: []string{invReqErr, "account " + s.addr5.String() + " does not have permission to update market 3"},
		},
		{
			name: "false to false",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 3, AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_update)},
					AcceptingCommitments: false,
				})
			},
			msg: exchange.MsgMarketUpdateAcceptingCommitmentsRequest{
				Admin:                s.addr5.String(),
				MarketId:             3,
				AcceptingCommitments: false,
			},
			expInErr: []string{invReqErr, "market 3 already has accepting-commitments false"},
		},
		{
			name: "true to true",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 3, AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_update)},
					AcceptingCommitments: true,
				})
			},
			msg: exchange.MsgMarketUpdateAcceptingCommitmentsRequest{
				Admin:                s.addr5.String(),
				MarketId:             3,
				AcceptingCommitments: true,
			},
			expInErr: []string{invReqErr, "market 3 already has accepting-commitments true"},
		},
		{
			name: "false to true: no fees: authority",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 3, AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_update)},
					AcceptingCommitments: false,
				})
			},
			msg: exchange.MsgMarketUpdateAcceptingCommitmentsRequest{
				Admin:                s.k.GetAuthority(),
				MarketId:             3,
				AcceptingCommitments: true,
			},
			expEvents: sdk.Events{
				s.untypeEvent(&exchange.EventMarketCommitmentsEnabled{MarketId: 3, UpdatedBy: s.k.GetAuthority()}),
			},
		},
		{
			name: "false to true: no fees: addr",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 3, AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_update)},
					AcceptingCommitments: false,
				})
			},
			msg: exchange.MsgMarketUpdateAcceptingCommitmentsRequest{
				Admin:                s.addr5.String(),
				MarketId:             3,
				AcceptingCommitments: true,
			},
			expInErr: []string{invReqErr, "market 3 does not have any commitment fees defined"},
		},
		{
			name: "false to true: with fees: addr",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 3, AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_update)},
					AcceptingCommitments:     false,
					CommitmentSettlementBips: 50,
					IntermediaryDenom:        "cherry",
				})
			},
			msg: exchange.MsgMarketUpdateAcceptingCommitmentsRequest{
				Admin:                s.addr5.String(),
				MarketId:             3,
				AcceptingCommitments: true,
			},
			expEvents: sdk.Events{
				s.untypeEvent(&exchange.EventMarketCommitmentsEnabled{MarketId: 3, UpdatedBy: s.addr5.String()}),
			},
		},
		{
			name: "true to false",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 3, AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_update)},
					AcceptingCommitments: true,
				})
			},
			msg: exchange.MsgMarketUpdateAcceptingCommitmentsRequest{
				Admin:                s.addr5.String(),
				MarketId:             3,
				AcceptingCommitments: false,
			},
			expEvents: sdk.Events{
				s.untypeEvent(&exchange.EventMarketCommitmentsDisabled{MarketId: 3, UpdatedBy: s.addr5.String()}),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			runMsgServerTestCase(s, testDef, tc)
		})
	}
}

func (s *TestSuite) TestMsgServer_MarketUpdateIntermediaryDenom() {
	testDef := msgServerTestDef[exchange.MsgMarketUpdateIntermediaryDenomRequest, exchange.MsgMarketUpdateIntermediaryDenomResponse, struct{}]{
		endpointName: "MarketUpdateIntermediaryDenom",
		endpoint:     keeper.NewMsgServer(s.k).MarketUpdateIntermediaryDenom,
		expResp:      &exchange.MsgMarketUpdateIntermediaryDenomResponse{},
		followup: func(msg *exchange.MsgMarketUpdateIntermediaryDenomRequest, _ struct{}) {
			denom := s.k.GetIntermediaryDenom(s.ctx, msg.MarketId)
			s.Assert().Equal(msg.IntermediaryDenom, denom, "GetIntermediaryDenom(%d)", msg.MarketId)
		},
	}

	tests := []msgServerTestCase[exchange.MsgMarketUpdateIntermediaryDenomRequest, struct{}]{
		{
			name: "admin does not have permission to update market",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:                 3,
					AccessGrants:             []exchange.AccessGrant{s.agCanAllBut(s.addr5, exchange.Permission_update)},
					AcceptingCommitments:     true,
					CommitmentSettlementBips: 50,
					IntermediaryDenom:        "banana",
				})
			},
			msg: exchange.MsgMarketUpdateIntermediaryDenomRequest{
				Admin:             s.addr5.String(),
				MarketId:          3,
				IntermediaryDenom: "cherry",
			},
			expInErr: []string{invReqErr, "account " + s.addr5.String() + " does not have permission to update market 3"},
		},
		{
			name: "admin has permission",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:                 3,
					AccessGrants:             []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_update)},
					AcceptingCommitments:     true,
					CommitmentSettlementBips: 50,
					IntermediaryDenom:        "banana",
				})
			},
			msg: exchange.MsgMarketUpdateIntermediaryDenomRequest{
				Admin:             s.addr5.String(),
				MarketId:          3,
				IntermediaryDenom: "cherry",
			},
			expEvents: sdk.Events{
				s.untypeEvent(&exchange.EventMarketIntermediaryDenomUpdated{MarketId: 3, UpdatedBy: s.addr5.String()}),
			},
		},
		{
			name: "authority",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:                 7,
					AcceptingCommitments:     true,
					CommitmentSettlementBips: 50,
					IntermediaryDenom:        "banana",
				})
			},
			msg: exchange.MsgMarketUpdateIntermediaryDenomRequest{
				Admin:             s.k.GetAuthority(),
				MarketId:          7,
				IntermediaryDenom: "cherry",
			},
			expEvents: sdk.Events{
				s.untypeEvent(&exchange.EventMarketIntermediaryDenomUpdated{MarketId: 7, UpdatedBy: s.k.GetAuthority()}),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			runMsgServerTestCase(s, testDef, tc)
		})
	}
}

func (s *TestSuite) TestMsgServer_MarketManagePermissions() {
	testDef := msgServerTestDef[exchange.MsgMarketManagePermissionsRequest, exchange.MsgMarketManagePermissionsResponse, []exchange.AccessGrant]{
		endpointName: "MarketManagePermissions",
		endpoint:     keeper.NewMsgServer(s.k).MarketManagePermissions,
		expResp:      &exchange.MsgMarketManagePermissionsResponse{},
		followup: func(msg *exchange.MsgMarketManagePermissionsRequest, expAGs []exchange.AccessGrant) {
			for _, expAG := range expAGs {
				addr, err := sdk.AccAddressFromBech32(expAG.Address)
				if s.Assert().NoError(err, "AccAddressFromBech32(%q)", expAG.Address) {
					actPerms := s.k.GetUserPermissions(s.ctx, msg.MarketId, addr)
					s.Assert().Equal(expAG.Permissions, actPerms, "market %d permissions for %s", msg.MarketId, s.getAddrName(addr))
				}

			}
		},
	}

	tests := []msgServerTestCase[exchange.MsgMarketManagePermissionsRequest, []exchange.AccessGrant]{
		{
			name: "admin does not have permission to manage permissions",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:     1,
					AccessGrants: []exchange.AccessGrant{s.agCanAllBut(s.addr5, exchange.Permission_permissions)},
				})
			},
			msg: exchange.MsgMarketManagePermissionsRequest{
				Admin:     s.addr5.String(),
				MarketId:  1,
				RevokeAll: []string{s.addr1.String()},
			},
			expInErr: []string{invReqErr,
				"account " + s.addr5.String() + " does not have permission to manage permissions for market 1"},
		},
		{
			name: "error updating permissions",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 1, AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_permissions)},
				})
			},
			msg: exchange.MsgMarketManagePermissionsRequest{
				Admin:     s.addr5.String(),
				MarketId:  1,
				RevokeAll: []string{s.addr1.String()},
			},
			expInErr: []string{invReqErr, "account " + s.addr1.String() + " does not have any permissions for market 1"},
		},
		{
			name: "okay",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 1,
					AccessGrants: []exchange.AccessGrant{
						s.agCanEverything(s.addr1),
						s.agCanEverything(s.addr2),
						s.agCanOnly(s.addr3, exchange.Permission_withdraw),
						s.agCanOnly(s.addr5, exchange.Permission_permissions),
					},
				})
			},
			msg: exchange.MsgMarketManagePermissionsRequest{
				Admin:     s.addr5.String(),
				MarketId:  1,
				RevokeAll: []string{s.addr1.String()},
				ToRevoke: []exchange.AccessGrant{
					s.agCanAllBut(s.addr2, exchange.Permission_cancel),
					s.agCanOnly(s.addr3, exchange.Permission_withdraw),
				},
				ToGrant: []exchange.AccessGrant{
					s.agCanOnly(s.addr4, exchange.Permission_withdraw),
				},
			},
			fArgs: []exchange.AccessGrant{
				{Address: s.addr1.String(), Permissions: nil},
				s.agCanOnly(s.addr2, exchange.Permission_cancel),
				{Address: s.addr3.String(), Permissions: nil},
				s.agCanOnly(s.addr4, exchange.Permission_withdraw),
				s.agCanOnly(s.addr5, exchange.Permission_permissions),
			},
			expEvents: sdk.Events{
				s.untypeEvent(&exchange.EventMarketPermissionsUpdated{MarketId: 1, UpdatedBy: s.addr5.String()}),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			runMsgServerTestCase(s, testDef, tc)
		})
	}
}

func (s *TestSuite) TestMsgServer_MarketManageReqAttrs() {
	type followupArgs struct {
		expAsk []string
		expBid []string
	}
	testDef := msgServerTestDef[exchange.MsgMarketManageReqAttrsRequest, exchange.MsgMarketManageReqAttrsResponse, followupArgs]{
		endpointName: "MarketManageReqAttrs",
		endpoint:     keeper.NewMsgServer(s.k).MarketManageReqAttrs,
		expResp:      &exchange.MsgMarketManageReqAttrsResponse{},
		followup: func(msg *exchange.MsgMarketManageReqAttrsRequest, fArgs followupArgs) {
			actAsk := s.k.GetReqAttrsAsk(s.ctx, msg.MarketId)
			actBid := s.k.GetReqAttrsBid(s.ctx, msg.MarketId)
			s.Assert().Equal(fArgs.expAsk, actAsk, "market %d req attrs ask", msg.MarketId)
			s.Assert().Equal(fArgs.expBid, actBid, "market %d req attrs bid", msg.MarketId)
		},
	}

	tests := []msgServerTestCase[exchange.MsgMarketManageReqAttrsRequest, followupArgs]{
		{
			name: "admin does not have permission to manage req attrs",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:     1,
					AccessGrants: []exchange.AccessGrant{s.agCanAllBut(s.addr5, exchange.Permission_attributes)},
				})
			},
			msg: exchange.MsgMarketManageReqAttrsRequest{
				Admin: s.addr5.String(), MarketId: 1, CreateAskToAdd: []string{"nope"},
			},
			expInErr: []string{invReqErr,
				"account " + s.addr5.String() + " does not have permission to manage required attributes for market 1"},
		},
		{
			name: "error updating attrs",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:     1,
					AccessGrants: []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_attributes)},
				})
			},
			msg: exchange.MsgMarketManageReqAttrsRequest{
				Admin:             s.addr5.String(),
				MarketId:          1,
				CreateAskToRemove: []string{"nope"},
			},
			expInErr: []string{invReqErr,
				"cannot remove create-ask required attribute \"nope\": attribute not currently required"},
		},
		{
			name: "okay",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:         1,
					AccessGrants:     []exchange.AccessGrant{s.agCanOnly(s.addr5, exchange.Permission_attributes)},
					ReqAttrCreateAsk: []string{"ask.base", "*.other"},
					ReqAttrCreateBid: []string{"bid.base", "*.fresh"},
				})
			},
			msg: exchange.MsgMarketManageReqAttrsRequest{
				Admin:             s.addr5.String(),
				MarketId:          1,
				CreateAskToAdd:    []string{"ask.deeper.base"},
				CreateAskToRemove: []string{"ask.base"},
				CreateBidToAdd:    []string{"bid.deeper.base"},
				CreateBidToRemove: []string{"bid.base"},
			},
			fArgs: followupArgs{
				expAsk: []string{"*.other", "ask.deeper.base"},
				expBid: []string{"*.fresh", "bid.deeper.base"},
			},
			expEvents: sdk.Events{
				s.untypeEvent(&exchange.EventMarketReqAttrUpdated{MarketId: 1, UpdatedBy: s.addr5.String()}),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			runMsgServerTestCase(s, testDef, tc)
		})
	}
}

func (s *TestSuite) TestMsgServer_GovCreateMarket() {
	testDef := msgServerTestDef[exchange.MsgGovCreateMarketRequest, exchange.MsgGovCreateMarketResponse, uint32]{
		endpointName: "GovCreateMarket",
		endpoint:     keeper.NewMsgServer(s.k).GovCreateMarket,
		expResp:      &exchange.MsgGovCreateMarketResponse{},
		followup: func(msg *exchange.MsgGovCreateMarketRequest, marketID uint32) {
			expMarket := msg.Market
			expMarket.MarketId = marketID
			actMarket := s.k.GetMarket(s.ctx, marketID)
			s.Assert().Equal(expMarket, *actMarket, "GetMarket(%d)", marketID)
		},
	}

	tests := []msgServerTestCase[exchange.MsgGovCreateMarketRequest, uint32]{
		{
			name: "wrong authority",
			msg: exchange.MsgGovCreateMarketRequest{
				Authority: s.addr5.String(),
				Market:    exchange.Market{MarketDetails: exchange.MarketDetails{Name: "Market 5"}},
			},
			expInErr: []string{
				"expected \"" + s.k.GetAuthority() + "\" got \"" + s.addr5.String() + "\"",
				"expected gov account as only signer for proposal message"},
		},
		{
			name: "error creating market",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId: 1, AccessGrants: []exchange.AccessGrant{s.agCanEverything(s.addr5)},
				})
			},
			msg: exchange.MsgGovCreateMarketRequest{
				Authority: s.k.GetAuthority(),
				Market: exchange.Market{
					MarketId: 1, MarketDetails: exchange.MarketDetails{Name: "Muwahahahaha"},
				},
			},
			expInErr: []string{invReqErr, "market id 1 account " + exchange.GetMarketAddress(1).String() + " already exists"},
		},
		{
			name: "okay: market 0",
			setup: func() {
				keeper.SetLastAutoMarketID(s.getStore(), 54)
			},
			msg: exchange.MsgGovCreateMarketRequest{
				Authority: s.k.GetAuthority(),
				Market: exchange.Market{
					MarketId: 0,
					MarketDetails: exchange.MarketDetails{
						Name:        "Next Market Please",
						Description: "A description!!",
						WebsiteUrl:  "WeBsItEuRl",
						IconUri:     "iCoNuRi",
					},
					FeeCreateBidFlat:          s.coins("10fig"),
					FeeSellerSettlementRatios: s.ratios("100apple:1apple"),
					FeeBuyerSettlementFlat:    s.coins("33fig"),
					AcceptingOrders:           true,
					AllowUserSettlement:       true,
					AccessGrants: []exchange.AccessGrant{
						s.agCanEverything(s.addr1),
						s.agCanEverything(s.addr5),
					},
					ReqAttrCreateAsk: []string{"*.some.thing"},
				},
			},
			fArgs: 55,
			expEvents: sdk.Events{
				s.untypeEvent(&exchange.EventMarketCreated{MarketId: 55}),
			},
		},
		{
			name: "okay: market 420",
			setup: func() {
				keeper.SetLastAutoMarketID(s.getStore(), 68)
			},
			msg: exchange.MsgGovCreateMarketRequest{
				Authority: s.k.GetAuthority(),
				Market: exchange.Market{
					MarketId: 420,
					MarketDetails: exchange.MarketDetails{
						Name:        "Second Day",
						Description: "It's Tuesday!",
						WebsiteUrl:  "websiteurl",
						IconUri:     "ICONURI",
					},
					FeeCreateAskFlat:         s.coins("10fig"),
					FeeBuyerSettlementRatios: s.ratios("100apple:1apple"),
					FeeSellerSettlementFlat:  s.coins("33fig"),
					AccessGrants: []exchange.AccessGrant{
						s.agCanEverything(s.addr4),
						s.agCanOnly(s.addr5, exchange.Permission_settle),
					},
					ReqAttrCreateBid: []string{"*.other.thing"},
				},
			},
			fArgs: 420,
			expEvents: sdk.Events{
				s.untypeEvent(&exchange.EventMarketCreated{MarketId: 420}),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			runMsgServerTestCase(s, testDef, tc)
		})
	}
}

func (s *TestSuite) TestMsgServer_GovManageFees() {
	testDef := msgServerTestDef[exchange.MsgGovManageFeesRequest, exchange.MsgGovManageFeesResponse, exchange.Market]{
		endpointName: "GovManageFees",
		endpoint:     keeper.NewMsgServer(s.k).GovManageFees,
		expResp:      &exchange.MsgGovManageFeesResponse{},
		followup: func(msg *exchange.MsgGovManageFeesRequest, expMarket exchange.Market) {
			actMarket := s.k.GetMarket(s.ctx, msg.MarketId)
			s.Assert().Equal(exchange.Market(expMarket), *actMarket, "GetMarket(%d)", msg.MarketId)
		},
	}

	tests := []msgServerTestCase[exchange.MsgGovManageFeesRequest, exchange.Market]{
		{
			name: "wrong authority",
			msg: exchange.MsgGovManageFeesRequest{
				Authority:           s.addr5.String(),
				AddFeeCreateAskFlat: s.coins("10fig"),
			},
			expInErr: []string{
				"expected \"" + s.k.GetAuthority() + "\" got \"" + s.addr5.String() + "\"",
				"expected gov account as only signer for proposal message"},
		},
		{
			name: "okay",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:                  2,
					MarketDetails:             exchange.MarketDetails{Name: "Market Too"},
					FeeCreateAskFlat:          s.coins("9apple,5tomato"),
					FeeCreateBidFlat:          s.coins("9avocado,6tomato"),
					FeeSellerSettlementFlat:   s.coins("10apple,2tomato"),
					FeeSellerSettlementRatios: s.ratios("100apple:33apple,100tomato:7tomato"),
					FeeBuyerSettlementFlat:    s.coins("9aubergine,1tomato"),
					FeeBuyerSettlementRatios:  s.ratios("100cherry:1cherry,100tomato:7tomato"),
				})
			},
			msg: exchange.MsgGovManageFeesRequest{
				Authority:                       s.k.GetAuthority(),
				MarketId:                        2,
				RemoveFeeCreateAskFlat:          s.coins("9apple"),
				AddFeeCreateAskFlat:             s.coins("10apple"),
				RemoveFeeCreateBidFlat:          s.coins("9avocado"),
				AddFeeCreateBidFlat:             s.coins("10avocado"),
				RemoveFeeSellerSettlementFlat:   s.coins("10apple"),
				AddFeeSellerSettlementFlat:      s.coins("10acai"),
				RemoveFeeSellerSettlementRatios: s.ratios("100apple:33apple"),
				AddFeeSellerSettlementRatios:    s.ratios("100acai:3acai"),
				RemoveFeeBuyerSettlementFlat:    s.coins("9aubergine"),
				AddFeeBuyerSettlementFlat:       s.coins("10aubergine"),
				RemoveFeeBuyerSettlementRatios:  s.ratios("100cherry:1cherry"),
				AddFeeBuyerSettlementRatios:     s.ratios("80cherry:3cherry"),
			},
			fArgs: exchange.Market{
				MarketId:                  2,
				MarketDetails:             exchange.MarketDetails{Name: "Market Too"},
				FeeCreateAskFlat:          s.coins("10apple,5tomato"),
				FeeCreateBidFlat:          s.coins("10avocado,6tomato"),
				FeeSellerSettlementFlat:   s.coins("10acai,2tomato"),
				FeeSellerSettlementRatios: s.ratios("100acai:3acai,100tomato:7tomato"),
				FeeBuyerSettlementFlat:    s.coins("10aubergine,1tomato"),
				FeeBuyerSettlementRatios:  s.ratios("80cherry:3cherry,100tomato:7tomato"),
			},
			expEvents: sdk.Events{
				s.untypeEvent(&exchange.EventMarketFeesUpdated{MarketId: 2}),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			runMsgServerTestCase(s, testDef, tc)
		})
	}
}

func (s *TestSuite) TestMsgServer_GovCloseMarket() {
	testDef := msgServerTestDef[exchange.MsgGovCloseMarketRequest, exchange.MsgGovCloseMarketResponse, exchange.Market]{
		endpointName: "GovCloseMarket",
		endpoint:     keeper.NewMsgServer(s.k).GovCloseMarket,
		expResp:      &exchange.MsgGovCloseMarketResponse{},
		followup: func(msg *exchange.MsgGovCloseMarketRequest, expMarket exchange.Market) {
			actMarket := s.k.GetMarket(s.ctx, msg.MarketId)
			s.Assert().Equal(expMarket, *actMarket, "GetMarket(%d)", msg.MarketId)

			var marketOrders []*exchange.Order
			s.k.IterateMarketOrders(s.ctx, expMarket.MarketId, func(orderID uint64, _ byte) bool {
				order, err := s.k.GetOrder(s.ctx, orderID)
				s.Require().NoError(err, "GetOrder(%d)", orderID)
				marketOrders = append(marketOrders, order)
				return false
			})
			s.Assert().Empty(marketOrders, "orders in market %d", msg.MarketId)

			var marketCommitments []exchange.Commitment
			s.k.IterateCommitments(s.ctx, func(commitment exchange.Commitment) bool {
				if commitment.MarketId == msg.MarketId {
					marketCommitments = append(marketCommitments, commitment)
				}
				return false
			})
			s.Assert().Empty(marketCommitments, "commitments in market %d", msg.MarketId)
		},
	}

	tests := []msgServerTestCase[exchange.MsgGovCloseMarketRequest, exchange.Market]{
		{
			name: "wrong authority",
			msg: exchange.MsgGovCloseMarketRequest{
				Authority: s.addr5.String(),
				MarketId:  3,
			},
			expInErr: []string{
				"expected \"" + s.k.GetAuthority() + "\" got \"" + s.addr5.String() + "\"",
				"expected gov account as only signer for proposal message"},
		},
		{
			name: "okay",
			setup: func() {
				s.requireCreateMarketUnmocked(exchange.Market{
					MarketId:                 2,
					AcceptingOrders:          true,
					AcceptingCommitments:     true,
					CommitmentSettlementBips: 51,
					IntermediaryDenom:        "cherry",
				})

				s.requireFundAccount(s.addr1, "10apple")
				s.requireFundAccount(s.addr2, "20peach")
				askOrder := exchange.NewOrder(18).WithAsk(&exchange.AskOrder{
					MarketId: 2, Seller: s.addr1.String(), Assets: s.coin("10apple"), Price: s.coin("20peach"),
				})
				bidOrder := exchange.NewOrder(19).WithBid(&exchange.BidOrder{
					MarketId: 2, Buyer: s.addr2.String(), Assets: s.coin("10apple"), Price: s.coin("20peach"),
				})
				s.requireSetOrdersInStore(s.getStore(), askOrder, bidOrder)
				err := s.app.HoldKeeper.AddHold(s.ctx, s.addr1, s.coins("10apple"), "testing")
				s.Require().NoError(err, "AddHold addr1 10apple")
				err = s.app.HoldKeeper.AddHold(s.ctx, s.addr2, s.coins("20peach"), "testing")
				s.Require().NoError(err, "AddHold addr2 20peach")

				s.requireFundAccount(s.addr3, "30banana")
				err = s.k.AddCommitment(s.ctx, 2, s.addr3, s.coins("30banana"), "testing")
				s.Require().NoError(err, "AddCommitment addr3")
			},
			msg: exchange.MsgGovCloseMarketRequest{
				Authority: s.k.GetAuthority(),
				MarketId:  2,
			},
			fArgs: exchange.Market{
				MarketId:                 2,
				AcceptingOrders:          false,
				AcceptingCommitments:     false,
				CommitmentSettlementBips: 51,
				IntermediaryDenom:        "cherry",
			},
			expEvents: sdk.Events{
				s.untypeEvent(exchange.NewEventMarketOrdersDisabled(2, s.k.GetAuthority())),
				s.untypeEvent(exchange.NewEventMarketCommitmentsDisabled(2, s.k.GetAuthority())),
				s.untypeEvent(hold.NewEventHoldReleased(s.addr1, s.coins("10apple"))),
				s.untypeEvent(&exchange.EventOrderCancelled{OrderId: 18, MarketId: 2, CancelledBy: s.k.GetAuthority()}),
				s.untypeEvent(hold.NewEventHoldReleased(s.addr2, s.coins("20peach"))),
				s.untypeEvent(&exchange.EventOrderCancelled{OrderId: 19, MarketId: 2, CancelledBy: s.k.GetAuthority()}),
				s.untypeEvent(hold.NewEventHoldReleased(s.addr3, s.coins("30banana"))),
				s.untypeEvent(exchange.NewEventCommitmentReleased(s.addr3.String(), 2, s.coins("30banana"), "GovCloseMarket")),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			runMsgServerTestCase(s, testDef, tc)
		})
	}
}

func (s *TestSuite) TestMsgServer_GovUpdateParams() {
	testDef := msgServerTestDef[exchange.MsgGovUpdateParamsRequest, exchange.MsgGovUpdateParamsResponse, struct{}]{
		endpointName: "GovUpdateParams",
		endpoint:     keeper.NewMsgServer(s.k).GovUpdateParams,
		expResp:      &exchange.MsgGovUpdateParamsResponse{},
		followup: func(msg *exchange.MsgGovUpdateParamsRequest, _ struct{}) {
			actParams := s.k.GetParams(s.ctx)
			s.Assert().Equal(&msg.Params, actParams, "GetParams")
		},
	}

	tests := []msgServerTestCase[exchange.MsgGovUpdateParamsRequest, struct{}]{
		{
			name: "wrong authority",
			msg: exchange.MsgGovUpdateParamsRequest{
				Authority: s.addr5.String(),
				Params:    exchange.Params{},
			},
			expInErr: []string{
				"expected \"" + s.k.GetAuthority() + "\" got \"" + s.addr5.String() + "\"",
				"expected gov account as only signer for proposal message"},
		},
		{
			name: "okay: was not previously set",
			setup: func() {
				s.k.SetParams(s.ctx, nil)
			},
			msg: exchange.MsgGovUpdateParamsRequest{
				Authority: s.k.GetAuthority(),
				Params: exchange.Params{
					DefaultSplit: 333,
					DenomSplits:  []exchange.DenomSplit{{Denom: "banana", Split: 99}},
				},
			},
			expEvents: sdk.Events{
				s.untypeEvent(&exchange.EventParamsUpdated{}),
			},
		},
		{
			name: "okay: no change",
			setup: func() {
				s.k.SetParams(s.ctx, exchange.DefaultParams())
			},
			msg: exchange.MsgGovUpdateParamsRequest{
				Authority: s.k.GetAuthority(),
				Params:    *exchange.DefaultParams(),
			},
			expEvents: sdk.Events{
				s.untypeEvent(&exchange.EventParamsUpdated{}),
			},
		},
		{
			name: "okay: was previously defaults",
			setup: func() {
				s.k.SetParams(s.ctx, exchange.DefaultParams())
			},
			msg: exchange.MsgGovUpdateParamsRequest{
				Authority: s.k.GetAuthority(),
				Params: exchange.Params{
					DefaultSplit: 333,
					DenomSplits:  []exchange.DenomSplit{{Denom: "banana", Split: 99}},
				},
			},
			expEvents: sdk.Events{
				s.untypeEvent(&exchange.EventParamsUpdated{}),
			},
		},
		{
			name: "okay: was previously set",
			setup: func() {
				s.k.SetParams(s.ctx, &exchange.Params{
					DefaultSplit: 987,
					DenomSplits:  []exchange.DenomSplit{{Denom: "cherry", Split: 4}},
				})
			},
			msg: exchange.MsgGovUpdateParamsRequest{
				Authority: s.k.GetAuthority(),
				Params: exchange.Params{
					DefaultSplit: 345,
					DenomSplits:  []exchange.DenomSplit{{Denom: "banana", Split: 99}},
				},
			},
			expEvents: sdk.Events{
				s.untypeEvent(&exchange.EventParamsUpdated{}),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			runMsgServerTestCase(s, testDef, tc)
		})
	}
}
