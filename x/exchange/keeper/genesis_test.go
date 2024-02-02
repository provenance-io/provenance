package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/provenance-io/provenance/x/exchange"
	"github.com/provenance-io/provenance/x/exchange/keeper"
)

// assertEqualGenState asserts that the provided gen states are equal and returns true if they are.
// If they are different, extra comparisons are done on each field to help identify what's actually different.
func (s *TestSuite) assertEqualGenState(expected, actual *exchange.GenesisState, msg string, args ...interface{}) bool {
	if s.Assert().Equalf(expected, actual, msg, args...) {
		return true
	}

	// If either are nil, that'll be obvious in the failure output, so we don't need to dig deeper.
	if expected == nil || actual == nil {
		return false
	}

	// Run assertions on individual pieces to hopefully help identify what's actually different.
	if !s.Assert().Equalf(expected.Params, actual.Params, msg+" Params", args...) && expected.Params != nil && actual.Params != nil {
		s.Assert().Equalf(int(expected.Params.DefaultSplit), int(actual.Params.DefaultSplit), msg+" Params.DefaultSplit", args...)
		assertEqualSlice(s, expected.Params.DenomSplits, actual.Params.DenomSplits, s.getGenStateDenomSplitStr, msg+" Params.DenomSplits", args...)
	}
	assertEqualSlice(s, expected.Markets, actual.Markets, s.getGenStateMarketStr, msg+" Markets", args...)
	assertEqualSlice(s, expected.Orders, actual.Orders, s.getGenStateOrderStr, msg+" Orders", args...)
	s.Assert().Equalf(int(expected.LastMarketId), int(actual.LastMarketId), msg+" LastMarketId", args...)
	s.Assert().Equalf(fmt.Sprintf("%d", expected.LastOrderId), fmt.Sprintf("%d", actual.LastOrderId), msg+" LastMarketId", args...)
	s.assertEqualCommitments(expected.Commitments, actual.Commitments, msg+" Commitments", args...)
	return false
}

// getGenStateMarketStr returns a string representing the market to help identify slice entries.
func (s *TestSuite) getGenStateDenomSplitStr(split exchange.DenomSplit) string {
	return fmt.Sprintf("%s=%d", split.Denom, split.Split)
}

// getGenStateMarketStr returns a string representing the market to help identify slice entries.
func (s *TestSuite) getGenStateMarketStr(market exchange.Market) string {
	return fmt.Sprintf("%d: %s", market.MarketId, market.MarketDetails.Name)
}

// getGenStateOrderStr returns a string representing the order to help identify slice entries.
func (s *TestSuite) getGenStateOrderStr(order exchange.Order) string {
	return fmt.Sprintf("%s order %d: %s %s at %s",
		order.GetOrderType(), order.OrderId, order.GetOwner(), order.GetAssets(), order.GetPrice())
}

func (s *TestSuite) TestKeeper_InitAndExportGenesis() {
	marketAcc := func(marketID uint32, name string) *exchange.MarketAccount {
		return &exchange.MarketAccount{
			BaseAccount:   &authtypes.BaseAccount{Address: exchange.GetMarketAddress(marketID).String()},
			MarketId:      marketID,
			MarketDetails: exchange.MarketDetails{Name: name},
		}
	}
	accAddr := func(prefix string, orderID uint64) sdk.AccAddress {
		return sdk.AccAddress(fmt.Sprintf("%s%d____________________", prefix, orderID)[:20])
	}
	assetDenom, priceDenom, feeDenom := "apple", "pear", "fig"
	askOrder := func(orderID uint64, marketID uint32, seller string) exchange.Order {
		if len(seller) == 0 {
			seller = accAddr("seller", orderID).String()
		}
		return *exchange.NewOrder(orderID).WithAsk(&exchange.AskOrder{
			MarketId:                marketID,
			Seller:                  seller,
			Assets:                  s.coin(fmt.Sprintf("%d%s", orderID, assetDenom)),
			Price:                   s.coin(fmt.Sprintf("%d%s", orderID, priceDenom)),
			SellerSettlementFlatFee: s.coinP(fmt.Sprintf("%d%s", orderID, feeDenom)),
			AllowPartial:            true,
			ExternalId:              fmt.Sprintf("ExtId%d", orderID),
		})
	}
	bidOrder := func(orderID uint64, marketID uint32, buyer string) exchange.Order {
		if len(buyer) == 0 {
			buyer = accAddr("buyer", orderID).String()
		}
		return *exchange.NewOrder(orderID).WithBid(&exchange.BidOrder{
			MarketId:            marketID,
			Buyer:               buyer,
			Assets:              s.coin(fmt.Sprintf("%d%s", orderID, assetDenom)),
			Price:               s.coin(fmt.Sprintf("%d%s", orderID, priceDenom)),
			BuyerSettlementFees: s.coins(fmt.Sprintf("%d%s", orderID, feeDenom)),
			AllowPartial:        true,
			ExternalId:          fmt.Sprintf("ExtId%d", orderID),
		})
	}
	askHoldCoins := func(orderID uint64) sdk.Coins {
		return s.coins(fmt.Sprintf("%d%s,%d%s", orderID, assetDenom, orderID, feeDenom))
	}
	bidHoldCoins := func(orderID uint64) sdk.Coins {
		return s.coins(fmt.Sprintf("%d%s,%d%s", orderID, priceDenom, orderID, feeDenom))
	}
	commitment := func(addr sdk.AccAddress, marketID uint32, amount string) exchange.Commitment {
		return exchange.Commitment{
			Account:  addr.String(),
			MarketId: marketID,
			Amount:   s.coins(amount),
		}
	}

	tests := []struct {
		name         string
		accKeeper    *MockAccountKeeper
		holdKeeper   *MockHoldKeeper
		setup        func()
		genState     *exchange.GenesisState
		expGenState  *exchange.GenesisState
		expInitPanic string
		expExportLog string
		expAccCalls  AccountCalls
		expHoldCalls HoldCalls
	}{
		{
			name:     "nil gen state",
			genState: nil,
		},
		{
			name:     "empty gen state",
			genState: &exchange.GenesisState{},
		},
		{
			name: "just params: no splits",
			genState: &exchange.GenesisState{
				Params: &exchange.Params{
					DefaultSplit: 777,
					DenomSplits:  nil,
				},
			},
		},
		{
			name: "just params: one split",
			genState: &exchange.GenesisState{
				Params: &exchange.Params{
					DefaultSplit: 777,
					DenomSplits: []exchange.DenomSplit{
						{Denom: "yam", Split: 333},
					},
				},
			},
		},
		{
			name: "just params: three splits",
			genState: &exchange.GenesisState{
				Params: &exchange.Params{
					DefaultSplit: 777,
					DenomSplits: []exchange.DenomSplit{
						{Denom: "green", Split: 999},
						{Denom: "orange", Split: 100},
						{Denom: "yellow", Split: 543},
					},
				},
			},
		},
		{
			name:      "one market: account already exists with same details",
			accKeeper: NewMockAccountKeeper().WithGetAccountResult(s.marketAddr1, marketAcc(1, "some name")),
			genState: &exchange.GenesisState{
				Markets: []exchange.Market{
					{
						MarketId:                  1,
						MarketDetails:             exchange.MarketDetails{Name: "some name"},
						FeeCreateAskFlat:          s.coins("1apple"),
						FeeCreateBidFlat:          s.coins("2banana"),
						FeeSellerSettlementFlat:   s.coins("3cactus"),
						FeeSellerSettlementRatios: s.ratios("4damson:5elderberry"),
						FeeBuyerSettlementFlat:    s.coins("6fig"),
						FeeBuyerSettlementRatios:  s.ratios("7grape:8honeydew"),
						AcceptingOrders:           true,
						AllowUserSettlement:       true,
						AccessGrants: []exchange.AccessGrant{
							{Address: s.addr1.String(), Permissions: exchange.AllPermissions()},
						},
						ReqAttrCreateAsk:         []string{"ask.create.req"},
						ReqAttrCreateBid:         []string{"bid.create.req"},
						AcceptingCommitments:     true,
						FeeCreateCommitmentFlat:  s.coins("9cherry"),
						CommitmentSettlementBips: 50,
						IntermediaryDenom:        "lemon",
						ReqAttrCreateCommitment:  []string{"commitment.create.req"},
					},
				},
			},
			expAccCalls: AccountCalls{GetAccount: []sdk.AccAddress{s.marketAddr1}},
		},
		{
			name:      "one market: account already exists with different details",
			accKeeper: NewMockAccountKeeper().WithGetAccountResult(s.marketAddr2, marketAcc(2, "existing name")),
			genState: &exchange.GenesisState{
				Markets: []exchange.Market{
					{
						MarketId:                  2,
						MarketDetails:             exchange.MarketDetails{Name: "new name"},
						FeeCreateAskFlat:          s.coins("1apple"),
						FeeSellerSettlementFlat:   s.coins("3cactus"),
						FeeSellerSettlementRatios: s.ratios("4damson:5elderberry"),
						ReqAttrCreateAsk:          []string{"ask.create.req"},
					},
				},
			},
			expAccCalls: AccountCalls{
				GetAccount: []sdk.AccAddress{s.marketAddr2},
				SetAccount: []authtypes.AccountI{marketAcc(2, "new name")},
			},
		},
		{
			name: "one market: account does not yet exist",
			genState: &exchange.GenesisState{
				Markets: []exchange.Market{{MarketId: 3, MarketDetails: exchange.MarketDetails{Name: "Name Three"}}},
			},
			expAccCalls: AccountCalls{
				GetAccount: []sdk.AccAddress{s.marketAddr3},
				NewAccount: []authtypes.AccountI{marketAcc(3, "Name Three")},
				SetAccount: []authtypes.AccountI{marketAcc(3, "Name Three")},
			},
		},
		{
			name: "three markets",
			// First will not yet have an account
			// Second will have an account with different details
			// Third will have an account with the same details
			accKeeper: NewMockAccountKeeper().
				WithGetAccountResult(exchange.GetMarketAddress(75), marketAcc(75, "Original Second")).
				WithGetAccountResult(s.marketAddr3, marketAcc(3, "Third")),
			genState: &exchange.GenesisState{
				Markets: []exchange.Market{
					{
						MarketId:                  1,
						MarketDetails:             exchange.MarketDetails{Name: "First"},
						FeeCreateAskFlat:          s.coins("1apple"),
						FeeSellerSettlementFlat:   s.coins("3cactus"),
						FeeSellerSettlementRatios: s.ratios("4damson:5elderberry"),
						AcceptingOrders:           true,
						ReqAttrCreateAsk:          []string{"ask.create.req"},
					},
					{
						MarketId:                 75,
						MarketDetails:            exchange.MarketDetails{Name: "New Second Wave"},
						FeeCreateBidFlat:         s.coins("2banana"),
						FeeBuyerSettlementFlat:   s.coins("6fig"),
						FeeBuyerSettlementRatios: s.ratios("7grape:8honeydew"),
						AllowUserSettlement:      true,
						AccessGrants: []exchange.AccessGrant{
							{Address: s.addr1.String(), Permissions: []exchange.Permission{1}},
							{Address: s.addr2.String(), Permissions: exchange.AllPermissions()},
						},
						ReqAttrCreateBid: []string{"bid.create.req"},
					},
					{
						MarketId:                  3,
						MarketDetails:             exchange.MarketDetails{Name: "Third"},
						FeeSellerSettlementRatios: nil,
						FeeBuyerSettlementRatios:  nil,
						AcceptingOrders:           true,
						AllowUserSettlement:       true,
						AccessGrants: []exchange.AccessGrant{
							{Address: s.addr1.String(), Permissions: []exchange.Permission{1, 2}},
							{Address: s.addr2.String(), Permissions: []exchange.Permission{3, 4}},
							{Address: s.addr3.String(), Permissions: []exchange.Permission{5, 6}},
							{Address: s.addr4.String(), Permissions: exchange.AllPermissions()},
							{Address: s.addr5.String(), Permissions: []exchange.Permission{7, 8}},
						},
					},
				},
			},
			expAccCalls: AccountCalls{
				GetAccount: []sdk.AccAddress{s.marketAddr1, exchange.GetMarketAddress(75), s.marketAddr3},
				NewAccount: []authtypes.AccountI{marketAcc(1, "First")},
				SetAccount: []authtypes.AccountI{marketAcc(1, "First"), marketAcc(75, "New Second Wave")},
			},
		},
		{
			name:       "one order: ask",
			holdKeeper: NewMockHoldKeeper().WithGetHoldCoinResult(accAddr("seller", 7), askHoldCoins(7)...),
			genState: &exchange.GenesisState{
				Orders:      []exchange.Order{askOrder(7, 2, "")},
				LastOrderId: 7,
			},
			expHoldCalls: HoldCalls{
				GetHoldCoin: []*GetHoldCoinArgs{
					{addr: accAddr("seller", 7), denom: assetDenom},
					{addr: accAddr("seller", 7), denom: feeDenom},
				},
			},
		},
		{
			name:       "one order: bid",
			holdKeeper: NewMockHoldKeeper().WithGetHoldCoinResult(accAddr("buyer", 4), bidHoldCoins(4)...),
			genState: &exchange.GenesisState{
				Orders:      []exchange.Order{bidOrder(4, 1, "")},
				LastOrderId: 4,
			},
			expHoldCalls: HoldCalls{
				GetHoldCoin: []*GetHoldCoinArgs{
					{addr: accAddr("buyer", 4), denom: feeDenom},
					{addr: accAddr("buyer", 4), denom: priceDenom},
				},
			},
		},
		{
			name: "several orders",
			holdKeeper: NewMockHoldKeeper().
				WithGetHoldCoinResult(accAddr("buyer", 70), bidHoldCoins(100)...). // extra should be okay.
				WithGetHoldCoinResult(accAddr("seller", 55), askHoldCoins(55)...).
				WithGetHoldCoinResult(s.addr1, bidHoldCoins(2).Add(askHoldCoins(44)...)...).
				WithGetHoldCoinResult(accAddr("buyer", 25), bidHoldCoins(25)...),
			genState: &exchange.GenesisState{
				Orders: []exchange.Order{
					bidOrder(70, 95, ""),
					askOrder(55, 8, ""),
					bidOrder(2, 8, s.addr1.String()),
					bidOrder(25, 36, ""),
					askOrder(33, 95, s.addr1.String()),
					askOrder(11, 95, s.addr1.String()),
				},
				LastOrderId: 100,
			},
			expHoldCalls: HoldCalls{
				GetHoldCoin: []*GetHoldCoinArgs{
					{addr: accAddr("buyer", 70), denom: feeDenom}, {addr: accAddr("buyer", 70), denom: priceDenom},
					{addr: accAddr("seller", 55), denom: assetDenom}, {addr: accAddr("seller", 55), denom: feeDenom},
					{addr: s.addr1, denom: assetDenom}, {addr: s.addr1, denom: feeDenom}, {addr: s.addr1, denom: priceDenom},
					{addr: accAddr("buyer", 25), denom: feeDenom}, {addr: accAddr("buyer", 25), denom: priceDenom},
				},
			},
		},
		{
			name: "error setting order",
			genState: &exchange.GenesisState{
				Orders: []exchange.Order{
					*exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
						MarketId:   1,
						Seller:     accAddr("seller", 1).String(),
						Assets:     s.coin("1" + assetDenom),
						Price:      s.coin("1" + priceDenom),
						ExternalId: "duplicate external id",
					}),
					*exchange.NewOrder(2).WithAsk(&exchange.AskOrder{
						MarketId:   1,
						Seller:     accAddr("seller", 2).String(),
						Assets:     s.coin("2" + assetDenom),
						Price:      s.coin("2" + priceDenom),
						ExternalId: "duplicate external id",
					}),
				},
			},
			expInitPanic: "failed to store Orders[1]: external id \"duplicate external id\" is already " +
				"in use by order 1: cannot be used for order 2",
		},
		{
			name: "error checking holds",
			holdKeeper: NewMockHoldKeeper().WithGetHoldCoinErrorResult(accAddr("buyer", 1), feeDenom,
				"this is an error that has been injected for testing"),
			genState: &exchange.GenesisState{
				Orders:      []exchange.Order{bidOrder(1, 1, "")},
				LastOrderId: 1,
			},
			expInitPanic: "failed to look up amount of \"" + feeDenom + "\" on hold for " +
				accAddr("buyer", 1).String() + ": this is an error that has been injected for testing",
			expHoldCalls: HoldCalls{
				GetHoldCoin: []*GetHoldCoinArgs{{addr: accAddr("buyer", 1), denom: feeDenom}},
			},
		},
		{
			name:       "not enough hold on account: ask",
			holdKeeper: NewMockHoldKeeper().WithGetHoldCoinResult(accAddr("seller", 7), askHoldCoins(6)...),
			genState: &exchange.GenesisState{
				Orders:      []exchange.Order{askOrder(7, 2, "")},
				LastOrderId: 7,
			},
			expHoldCalls: HoldCalls{
				GetHoldCoin: []*GetHoldCoinArgs{{addr: accAddr("seller", 7), denom: assetDenom}},
			},
			expInitPanic: "account " + accAddr("seller", 7).String() + " should have at least \"7" + assetDenom + "\" on hold " +
				"(due to the exchange module), but only has \"6" + assetDenom + "\"",
		},
		{
			name:       "not enough hold on account: bid",
			holdKeeper: NewMockHoldKeeper().WithGetHoldCoinResult(accAddr("buyer", 777), bidHoldCoins(776)...),
			genState: &exchange.GenesisState{
				Orders:      []exchange.Order{bidOrder(777, 1, "")},
				LastOrderId: 1000,
			},
			expHoldCalls: HoldCalls{
				GetHoldCoin: []*GetHoldCoinArgs{{addr: accAddr("buyer", 777), denom: feeDenom}},
			},
			expInitPanic: "account " + accAddr("buyer", 777).String() + " should have at least \"777" + feeDenom + "\" on hold " +
				"(due to the exchange module), but only has \"776" + feeDenom + "\"",
		},
		{
			name: "last order id too low",
			holdKeeper: NewMockHoldKeeper().
				WithGetHoldCoinResult(accAddr("buyer", 70), bidHoldCoins(100)...). // extra should be okay.
				WithGetHoldCoinResult(accAddr("seller", 55), askHoldCoins(55)...).
				WithGetHoldCoinResult(s.addr1, bidHoldCoins(2).Add(askHoldCoins(44)...)...).
				WithGetHoldCoinResult(accAddr("buyer", 25), bidHoldCoins(25)...),
			genState: &exchange.GenesisState{
				Orders: []exchange.Order{
					bidOrder(70, 95, ""),
					askOrder(55, 8, ""),
					bidOrder(2, 8, s.addr1.String()),
					bidOrder(25, 36, ""),
					askOrder(33, 95, s.addr1.String()),
					askOrder(11, 95, s.addr1.String()),
				},
				LastOrderId: 69,
			},
			expInitPanic: "last order id 69 is less than largest order id 70",
		},
		{
			name:     "just last market id",
			genState: &exchange.GenesisState{LastMarketId: 8},
		},
		{
			name:     "just last order id",
			genState: &exchange.GenesisState{LastOrderId: 9},
		},
		{
			name: "error reading orders",
			setup: func() {
				store := s.getStore()
				order1 := askOrder(1, 1, "")
				s.requireSetOrderInStore(store, &order1)
				key2, value2, err := s.k.GetOrderStoreKeyValue(askOrder(2, 1, ""))
				s.Require().NoError(err, "GetOrderStoreKeyValue 2")
				value2[0] = 8
				store.Set(key2, value2)
				key3, value3, err := s.k.GetOrderStoreKeyValue(bidOrder(3, 1, ""))
				s.Require().NoError(err, "GetOrderStoreKeyValue 3")
				value3[0] = 8
				store.Set(key3, value3)
				order4 := bidOrder(4, 1, "")
				s.requireSetOrderInStore(store, &order4)
				keeper.SetLastOrderID(store, 4)
			},
			expGenState: &exchange.GenesisState{
				Orders:      []exchange.Order{askOrder(1, 1, ""), bidOrder(4, 1, "")},
				LastOrderId: 4,
			},
			expExportLog: "ERR error (ignored) while reading orders: failed to read order 2: unknown type byte 0x8\n" +
				"failed to read order 3: unknown type byte 0x8 module=x/exchange\n",
		},
		{
			name:       "one commitment",
			holdKeeper: NewMockHoldKeeper().WithGetHoldCoinResult(s.addr2, s.coins("25cherry")...),
			genState: &exchange.GenesisState{
				Commitments: []exchange.Commitment{commitment(s.addr2, 1, "25cherry")},
			},
			expHoldCalls: HoldCalls{GetHoldCoin: []*GetHoldCoinArgs{{addr: s.addr2, denom: "cherry"}}},
		},
		{
			name: "three commitments",
			holdKeeper: NewMockHoldKeeper().
				WithGetHoldCoinResult(s.addr2, s.coins("16apple,16cherry")...).
				WithGetHoldCoinResult(s.addr3, s.coins("17apple,32cherry,15pear")...),
			genState: &exchange.GenesisState{
				Commitments: []exchange.Commitment{
					commitment(s.addr2, 1, "16apple,16cherry"),
					commitment(s.addr3, 1, "15cherry,15pear"),
					commitment(s.addr3, 3, "17apple,17cherry"),
				},
			},
			expHoldCalls: HoldCalls{
				GetHoldCoin: []*GetHoldCoinArgs{
					{addr: s.addr2, denom: "apple"}, {addr: s.addr2, denom: "cherry"},
					{addr: s.addr3, denom: "apple"}, {addr: s.addr3, denom: "cherry"}, {addr: s.addr3, denom: "pear"},
				},
			},
		},
		{
			name: "commitment with bad address",
			genState: &exchange.GenesisState{
				Commitments: []exchange.Commitment{
					commitment(s.addr2, 1, "22cherry"),
					commitment(s.addr3, 1, "23cherry"),
					{Account: "badbadaddr", MarketId: 1, Amount: s.coins("99cherry")},
					commitment(s.addr4, 1, "24cherry"),
				},
			},
			expInitPanic: "failed to convert commitments[2].Account=\"badbadaddr\" to AccAddress: " +
				"decoding bech32 failed: invalid separator index -1",
		},
		{
			name:       "not enough hold on account: commitment",
			holdKeeper: NewMockHoldKeeper().WithGetHoldCoinResult(s.addr2, s.coins("24cherry")...),
			genState: &exchange.GenesisState{
				Commitments: []exchange.Commitment{commitment(s.addr2, 1, "25cherry")},
			},
			expHoldCalls: HoldCalls{GetHoldCoin: []*GetHoldCoinArgs{{addr: s.addr2, denom: "cherry"}}},
			expInitPanic: "account " + s.addr2.String() + " should have at least \"25cherry\" on hold " +
				"(due to the exchange module), but only has \"24cherry\"",
		},
		{
			name:       "bad commitment entry in state",
			holdKeeper: NewMockHoldKeeper().WithGetHoldCoinResult(s.addr3, s.coins("25cherry")...),
			setup: func() {
				s.getStore().Set(keeper.MakeKeyCommitment(8, s.addr3), []byte("x")) // Entry should just get ignored.
			},
			genState: &exchange.GenesisState{
				Commitments: []exchange.Commitment{commitment(s.addr3, 1, "25cherry")},
			},
			expHoldCalls: HoldCalls{GetHoldCoin: []*GetHoldCoinArgs{{addr: s.addr3, denom: "cherry"}}},
		},
		{
			name: "not enough hold on account: multiple sources",
			holdKeeper: NewMockHoldKeeper().
				WithGetHoldCoinResult(s.addr3, s.coins("109apple,54cherry,164fig,110pear")...), // 54+55+56=165 req fig.
			genState: &exchange.GenesisState{
				Orders: []exchange.Order{
					askOrder(55, 1, s.addr3.String()),
					bidOrder(56, 1, s.addr3.String()),
				},
				LastOrderId: 100,
				Commitments: []exchange.Commitment{commitment(s.addr3, 1, "54apple,54cherry,54fig,54pear")},
			},
			expHoldCalls: HoldCalls{
				GetHoldCoin: []*GetHoldCoinArgs{
					{addr: s.addr3, denom: "apple"}, {addr: s.addr3, denom: "cherry"}, {addr: s.addr3, denom: "fig"},
				},
			},
			expInitPanic: "account " + s.addr3.String() + " should have at least \"165fig\" on hold " +
				"(due to the exchange module), but only has \"164fig\"",
		},
		{
			name: "a little of everything",
			holdKeeper: NewMockHoldKeeper().
				WithGetHoldCoinResult(s.addr1, askHoldCoins(1)...).
				WithGetHoldCoinResult(s.addr2, bidHoldCoins(10)...).
				WithGetHoldCoinResult(s.addr3, bidHoldCoins(77).Add(askHoldCoins(79)...).Add(s.coins("25cherry,25fig")...)...).
				WithGetHoldCoinResult(s.addr4, askHoldCoins(1101)...).
				WithGetHoldCoinResult(s.addr5, s.coins("53cherry,27grape")...),
			genState: &exchange.GenesisState{
				Params: &exchange.Params{DefaultSplit: 333},
				Markets: []exchange.Market{
					{
						MarketId:            1,
						MarketDetails:       exchange.MarketDetails{Name: "First Market"},
						AcceptingOrders:     true,
						AllowUserSettlement: true,
						AccessGrants: []exchange.AccessGrant{
							{Address: s.addr1.String(), Permissions: exchange.AllPermissions()},
						},
						AcceptingCommitments:     true,
						CommitmentSettlementBips: 3,
						IntermediaryDenom:        "lemon",
					},
					{
						MarketId:            420,
						MarketDetails:       exchange.MarketDetails{Name: "THE Market"},
						AcceptingOrders:     true,
						AllowUserSettlement: true,
						AccessGrants: []exchange.AccessGrant{
							{Address: s.addr4.String(), Permissions: exchange.AllPermissions()},
						},
					},
				},
				Orders: []exchange.Order{
					askOrder(1, 1, s.addr1.String()),
					bidOrder(2, 1, s.addr2.String()),
					bidOrder(8, 420, s.addr2.String()),
					bidOrder(77, 1, s.addr3.String()),
					askOrder(79, 420, s.addr3.String()),
					askOrder(1101, 1, s.addr4.String()),
				},
				LastMarketId: 66,
				LastOrderId:  5555,
				Commitments: []exchange.Commitment{
					commitment(s.addr3, 1, "25cherry,25fig"),
					commitment(s.addr5, 1, "26cherry"),
					commitment(s.addr5, 420, "27cherry,27grape"),
				},
			},
			expAccCalls: AccountCalls{
				GetAccount: []sdk.AccAddress{s.marketAddr1, exchange.GetMarketAddress(420)},
				SetAccount: []authtypes.AccountI{marketAcc(1, "First Market"), marketAcc(420, "THE Market")},
				NewAccount: []authtypes.AccountI{marketAcc(1, "First Market"), marketAcc(420, "THE Market")},
			},
			expHoldCalls: HoldCalls{
				GetHoldCoin: []*GetHoldCoinArgs{
					{addr: s.addr1, denom: assetDenom}, {addr: s.addr1, denom: feeDenom},
					{addr: s.addr2, denom: feeDenom}, {addr: s.addr2, denom: priceDenom},
					{addr: s.addr3, denom: assetDenom}, {addr: s.addr3, denom: "cherry"},
					{addr: s.addr3, denom: feeDenom}, {addr: s.addr3, denom: priceDenom},
					{addr: s.addr4, denom: assetDenom}, {addr: s.addr4, denom: feeDenom},
					{addr: s.addr5, denom: "cherry"}, {addr: s.addr5, denom: "grape"},
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			origGenState := s.copyGenState(tc.genState)

			if tc.accKeeper == nil {
				tc.accKeeper = NewMockAccountKeeper()
			}
			if tc.holdKeeper == nil {
				tc.holdKeeper = NewMockHoldKeeper()
			}
			if tc.expGenState == nil && len(tc.expInitPanic) == 0 {
				tc.expGenState = s.sortGenState(s.copyGenState(tc.genState))
			}
			if tc.expGenState == nil {
				tc.expGenState = &exchange.GenesisState{}
			}

			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			kpr := s.k.WithAccountKeeper(tc.accKeeper).WithHoldKeeper(tc.holdKeeper)
			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)
			testInit := func() {
				kpr.InitGenesis(ctx, tc.genState)
			}
			s.requirePanicEquals(testInit, tc.expInitPanic, "InitGenesis")
			s.assertEqualGenState(origGenState, tc.genState, "GenState before (expected) and after (actual) InitGenesis")
			events := em.Events()
			s.assertEqualEvents(nil, events, "events emitted during InitGenesis")
			s.assertAccountKeeperCalls(tc.accKeeper, tc.expAccCalls, "InitGenesis")
			s.assertHoldKeeperCalls(tc.holdKeeper, tc.expHoldCalls, "InitGenesis")
			if len(tc.expInitPanic) > 0 {
				return
			}

			s.logBuffer.Reset()
			var actGenState *exchange.GenesisState
			testExport := func() {
				actGenState = kpr.ExportGenesis(s.ctx)
			}
			s.Require().NotPanics(testExport, "ExportGenesis")
			s.assertEqualGenState(tc.expGenState, actGenState, "ExportGenesis")
			actExportLog := s.getLogOutput("ExportGenesis")
			s.Assert().Equal(tc.expExportLog, actExportLog, "things logged during ExportGenesis")
		})
	}
}
