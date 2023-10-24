package keeper_test

import (
	"github.com/gogo/protobuf/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/provenance-io/provenance/x/exchange"
)

func (s *TestSuite) TestKeeper_FillBids() {
	tests := []struct {
		name         string
		attrKeeper   *MockAttributeKeeper
		bankKeeper   *MockBankKeeper
		holdKeeper   *MockHoldKeeper
		setup        func()
		msg          exchange.MsgFillBidsRequest
		expErr       string
		expEvents    []*exchange.EventOrderFilled
		expAttrCalls AttributeCalls
		expHoldCalls HoldCalls
		expBankCalls BankCalls
	}{
		// Tests on error conditions.
		{
			name: "invalid msg",
			msg: exchange.MsgFillBidsRequest{
				Seller:      s.addr1.String(),
				MarketId:    0,
				TotalAssets: s.coins("1apple"),
				BidOrderIds: []uint64{1},
			},
			expErr: "invalid market id: cannot be zero",
		},
		{
			name: "market does not exist",
			msg: exchange.MsgFillBidsRequest{
				Seller:      s.addr1.String(),
				MarketId:    1,
				TotalAssets: s.coins("1apple"),
				BidOrderIds: []uint64{1},
			},
			expErr: "market 1 does not exist",
		},
		{
			name: "market not accepting orders",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:        1,
					AcceptingOrders: false,
				})
			},
			msg: exchange.MsgFillBidsRequest{
				Seller:      s.addr1.String(),
				MarketId:    1,
				TotalAssets: s.coins("1apple"),
				BidOrderIds: []uint64{1},
			},
			expErr: "market 1 is not accepting orders",
		},
		{
			name: "market does not allow user-settle",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId: 1, AcceptingOrders: true,
					AllowUserSettlement: false,
				})
			},
			msg: exchange.MsgFillBidsRequest{
				Seller:      s.addr1.String(),
				MarketId:    1,
				TotalAssets: s.coins("1apple"),
				BidOrderIds: []uint64{1},
			},
			expErr: "market 1 does not allow user settlement",
		},
		{
			name: "seller cannot create ask",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId: 1, AcceptingOrders: true, AllowUserSettlement: true,
					ReqAttrCreateAsk: []string{"some.attr.no.one.has"},
				})
			},
			msg: exchange.MsgFillBidsRequest{
				Seller:      s.addr1.String(),
				MarketId:    1,
				TotalAssets: s.coins("1apple"),
				BidOrderIds: []uint64{1},
			},
			expErr:       "account " + s.addr1.String() + " is not allowed to create ask orders in market 1",
			expAttrCalls: AttributeCalls{GetAllAttributesAddr: [][]byte{s.addr1}},
		},
		{
			name: "not enough creation fee",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId: 1, AcceptingOrders: true, AllowUserSettlement: true,
					FeeCreateAskFlat: s.coins("5fig"),
				})
			},
			msg: exchange.MsgFillBidsRequest{
				Seller:              s.addr1.String(),
				MarketId:            1,
				TotalAssets:         s.coins("1apple"),
				BidOrderIds:         []uint64{1},
				AskOrderCreationFee: s.coinP("4fig"),
			},
			expErr: "insufficient ask order creation fee: \"4fig\" is less than required amount \"5fig\"",
		},
		{
			name: "not enough seller settlement flat fee",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId: 1, AcceptingOrders: true, AllowUserSettlement: true,
					FeeSellerSettlementFlat: s.coins("5fig"),
				})
			},
			msg: exchange.MsgFillBidsRequest{
				Seller:                  s.addr1.String(),
				MarketId:                1,
				TotalAssets:             s.coins("1apple"),
				BidOrderIds:             []uint64{1},
				SellerSettlementFlatFee: s.coinP("4fig"),
			},
			expErr: "insufficient seller settlement flat fee: \"4fig\" is less than required amount \"5fig\"",
		},
		{
			name: "bid order does not exist",
			setup: func() {
				s.requireCreateMarket(exchange.Market{MarketId: 1, AcceptingOrders: true, AllowUserSettlement: true})
			},
			msg: exchange.MsgFillBidsRequest{
				Seller:      s.addr1.String(),
				MarketId:    1,
				TotalAssets: s.coins("1apple"),
				BidOrderIds: []uint64{8},
			},
			expErr: "order 8 not found",
		},
		{
			name: "ask order id provided",
			setup: func() {
				s.requireCreateMarket(exchange.Market{MarketId: 1, AcceptingOrders: true, AllowUserSettlement: true})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(8).WithAsk(&exchange.AskOrder{
					MarketId: 1,
					Seller:   s.addr2.String(),
					Assets:   s.coin("1apple"),
					Price:    s.coin("1plum"),
				}))
			},
			msg: exchange.MsgFillBidsRequest{
				Seller:      s.addr1.String(),
				MarketId:    1,
				TotalAssets: s.coins("1apple"),
				BidOrderIds: []uint64{8},
			},
			expErr: "order 8 is type ask: expected bid",
		},
		{
			name: "order in wrong market",
			setup: func() {
				s.requireCreateMarket(exchange.Market{MarketId: 1, AcceptingOrders: true, AllowUserSettlement: true})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(8).WithBid(&exchange.BidOrder{
					MarketId: 2,
					Buyer:    s.addr2.String(),
					Assets:   s.coin("1apple"),
					Price:    s.coin("1plum"),
				}))
			},
			msg: exchange.MsgFillBidsRequest{
				Seller:      s.addr1.String(),
				MarketId:    1,
				TotalAssets: s.coins("1apple"),
				BidOrderIds: []uint64{8},
			},
			expErr: "order 8 market id 2 does not equal requested market id 1",
		},
		{
			name: "order has same buyer as provided seller",
			setup: func() {
				s.requireCreateMarket(exchange.Market{MarketId: 1, AcceptingOrders: true, AllowUserSettlement: true})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(8).WithBid(&exchange.BidOrder{
					MarketId: 1,
					Buyer:    s.addr1.String(),
					Assets:   s.coin("1apple"),
					Price:    s.coin("1plum"),
				}))
			},
			msg: exchange.MsgFillBidsRequest{
				Seller:      s.addr1.String(),
				MarketId:    1,
				TotalAssets: s.coins("1apple"),
				BidOrderIds: []uint64{8},
			},
			expErr: "order 8 has the same buyer " + s.addr1.String() + " as the requested seller",
		},
		{
			name: "multiple problems with orders",
			setup: func() {
				s.requireCreateMarket(exchange.Market{MarketId: 1, AcceptingOrders: true, AllowUserSettlement: true})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(3).WithBid(&exchange.BidOrder{
					MarketId: 2,
					Buyer:    s.addr2.String(),
					Assets:   s.coin("1apple"),
					Price:    s.coin("1plum"),
				}))
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(17).WithAsk(&exchange.AskOrder{
					MarketId: 1,
					Seller:   s.addr1.String(),
					Assets:   s.coin("1apple"),
					Price:    s.coin("1plum"),
				}))
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(11).WithBid(&exchange.BidOrder{
					MarketId: 1,
					Buyer:    s.addr1.String(),
					Assets:   s.coin("1apple"),
					Price:    s.coin("1plum"),
				}))
			},
			msg: exchange.MsgFillBidsRequest{
				Seller:      s.addr1.String(),
				MarketId:    1,
				TotalAssets: s.coins("1apple"),
				BidOrderIds: []uint64{8, 3, 17, 11},
			},
			expErr: s.joinErrs(
				"order 8 not found",
				"order 3 market id 2 does not equal requested market id 1",
				"order 17 is type ask: expected bid",
				"order 11 has the same buyer "+s.addr1.String()+" as the requested seller",
			),
		},
		{
			name: "provided total assets less than actual total assets",
			setup: func() {
				s.requireCreateMarket(exchange.Market{MarketId: 2, AcceptingOrders: true, AllowUserSettlement: true})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(1).WithBid(&exchange.BidOrder{
					Assets: s.coin("1apple"), Price: s.coin("1plum"), MarketId: 2, Buyer: s.addr1.String(),
				}))
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(2).WithBid(&exchange.BidOrder{
					Assets: s.coin("2apple"), Price: s.coin("2plum"), MarketId: 2, Buyer: s.addr2.String(),
				}))
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(3).WithBid(&exchange.BidOrder{
					Assets: s.coin("3apple"), Price: s.coin("3plum"), MarketId: 2, Buyer: s.addr3.String(),
				}))
			},
			msg: exchange.MsgFillBidsRequest{
				Seller:      s.addr4.String(),
				MarketId:    2,
				TotalAssets: s.coins("5apple"),
				BidOrderIds: []uint64{1, 2, 3},
			},
			expErr: "total assets \"5apple\" does not equal sum of bid order assets \"6apple\"",
		},
		{
			name: "provided total assets more than actual total assets",
			setup: func() {
				s.requireCreateMarket(exchange.Market{MarketId: 2, AcceptingOrders: true, AllowUserSettlement: true})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(1).WithBid(&exchange.BidOrder{
					Assets: s.coin("1apple"), Price: s.coin("1plum"), MarketId: 2, Buyer: s.addr1.String(),
				}))
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(2).WithBid(&exchange.BidOrder{
					Assets: s.coin("2apple"), Price: s.coin("2plum"), MarketId: 2, Buyer: s.addr2.String(),
				}))
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(3).WithBid(&exchange.BidOrder{
					Assets: s.coin("3apple"), Price: s.coin("3plum"), MarketId: 2, Buyer: s.addr3.String(),
				}))
			},
			msg: exchange.MsgFillBidsRequest{
				Seller:      s.addr4.String(),
				MarketId:    2,
				TotalAssets: s.coins("7apple"),
				BidOrderIds: []uint64{1, 2, 3},
			},
			expErr: "total assets \"7apple\" does not equal sum of bid order assets \"6apple\"",
		},
		{
			name: "ratio fee calc error",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId: 2, AcceptingOrders: true, AllowUserSettlement: true,
					FeeSellerSettlementRatios: s.ratios("20prune:1prune"),
				})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(1).WithBid(&exchange.BidOrder{
					Assets: s.coin("6apple"), Price: s.coin("6plum"), MarketId: 2, Buyer: s.addr1.String(),
				}))
			},
			msg: exchange.MsgFillBidsRequest{
				Seller:      s.addr4.String(),
				MarketId:    2,
				TotalAssets: s.coins("6apple"),
				BidOrderIds: []uint64{1},
			},
			expErr: "error calculating seller settlement ratio fee for order 1: no seller " +
				"settlement fee ratio found for denom \"plum\"",
		},
		{
			name: "invalid bid order owner",
			setup: func() {
				s.requireCreateMarket(exchange.Market{MarketId: 2, AcceptingOrders: true, AllowUserSettlement: true})
				key, value, err := s.k.GetOrderStoreKeyValue(*exchange.NewOrder(1).WithBid(&exchange.BidOrder{
					Assets: s.coin("6apple"), Price: s.coin("6plum"), MarketId: 2, Buyer: "badbuyer",
				}))
				s.Require().NoError(err, "GetOrderStoreKeyValue 1")
				s.getStore().Set(key, value)
			},
			msg: exchange.MsgFillBidsRequest{
				Seller:      s.addr4.String(),
				MarketId:    2,
				TotalAssets: s.coins("6apple"),
				BidOrderIds: []uint64{1},
			},
			expErr: "invalid bid order 1 owner \"badbuyer\": decoding bech32 failed: invalid separator index -1",
		},
		{
			name:       "error releasing hold",
			holdKeeper: NewMockHoldKeeper().WithReleaseHoldResults("no plum for you"),
			setup: func() {
				s.requireCreateMarket(exchange.Market{MarketId: 2, AcceptingOrders: true, AllowUserSettlement: true})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(1).WithBid(&exchange.BidOrder{
					Assets: s.coin("6apple"), Price: s.coin("6plum"), MarketId: 2, Buyer: s.addr1.String(),
				}))
			},
			msg: exchange.MsgFillBidsRequest{
				Seller:      s.addr4.String(),
				MarketId:    2,
				TotalAssets: s.coins("6apple"),
				BidOrderIds: []uint64{1},
			},
			expErr:       "error releasing hold for bid order 1: no plum for you",
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{{addr: s.addr1, funds: s.coins("6plum")}}},
		},
		{
			name:       "error transferring assets",
			bankKeeper: NewMockBankKeeper().WithSendCoinsResults("first transfer error"),
			setup: func() {
				s.requireCreateMarket(exchange.Market{MarketId: 2, AcceptingOrders: true, AllowUserSettlement: true})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(1).WithBid(&exchange.BidOrder{
					Assets: s.coin("1apple"), Price: s.coin("6plum"), MarketId: 2, Buyer: s.addr1.String(),
				}))
			},
			msg: exchange.MsgFillBidsRequest{
				Seller:      s.addr4.String(),
				MarketId:    2,
				TotalAssets: s.coins("1apple"),
				BidOrderIds: []uint64{1},
			},
			expErr:       "first transfer error",
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{{addr: s.addr1, funds: s.coins("6plum")}}},
			expBankCalls: BankCalls{SendCoins: []*SendCoinsArgs{
				{ctxHasQuarantineBypass: true, fromAddr: s.addr4, toAddr: s.addr1, amt: s.coins("1apple")},
				{ctxHasQuarantineBypass: true, fromAddr: s.addr1, toAddr: s.addr4, amt: s.coins("6plum")},
			}},
		},
		{
			name:       "error transferring price",
			bankKeeper: NewMockBankKeeper().WithSendCoinsResults("", "second transfer error"),
			setup: func() {
				s.k.SetParams(s.ctx, &exchange.Params{})
				s.requireCreateMarket(exchange.Market{MarketId: 2, AcceptingOrders: true, AllowUserSettlement: true})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(1).WithBid(&exchange.BidOrder{
					Assets: s.coin("1apple"), Price: s.coin("6plum"), MarketId: 2, Buyer: s.addr1.String(),
				}))
			},
			msg: exchange.MsgFillBidsRequest{
				Seller:      s.addr4.String(),
				MarketId:    2,
				TotalAssets: s.coins("1apple"),
				BidOrderIds: []uint64{1},
			},
			expErr:       "second transfer error",
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{{addr: s.addr1, funds: s.coins("6plum")}}},
			expBankCalls: BankCalls{SendCoins: []*SendCoinsArgs{
				{ctxHasQuarantineBypass: true, fromAddr: s.addr4, toAddr: s.addr1, amt: s.coins("1apple")},
				{ctxHasQuarantineBypass: true, fromAddr: s.addr1, toAddr: s.addr4, amt: s.coins("6plum")},
			}},
		},
		{
			name:       "error collecting settlement fees",
			bankKeeper: NewMockBankKeeper().WithInputOutputCoinsResults("first fake error"),
			setup: func() {
				s.k.SetParams(s.ctx, &exchange.Params{})
				s.requireCreateMarket(exchange.Market{
					MarketId: 2, AcceptingOrders: true, AllowUserSettlement: true,
					FeeSellerSettlementRatios: s.ratios("6plum:1plum"),
					FeeBuyerSettlementRatios:  s.ratios("6plum:2fig"),
				})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(99).WithBid(&exchange.BidOrder{
					Assets: s.coin("1apple"), Price: s.coin("6plum"), MarketId: 2, Buyer: s.addr1.String(),
					BuyerSettlementFees: s.coins("2fig"),
				}))
			},
			msg: exchange.MsgFillBidsRequest{
				Seller:      s.addr4.String(),
				MarketId:    2,
				TotalAssets: s.coins("1apple"),
				BidOrderIds: []uint64{99},
			},
			expErr:       "error collecting fees for market 2: first fake error",
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{{addr: s.addr1, funds: s.coins("2fig,6plum")}}},
			expBankCalls: BankCalls{
				SendCoins: []*SendCoinsArgs{
					{ctxHasQuarantineBypass: true, fromAddr: s.addr4, toAddr: s.addr1, amt: s.coins("1apple")},
					{ctxHasQuarantineBypass: true, fromAddr: s.addr1, toAddr: s.addr4, amt: s.coins("6plum")},
				},
				InputOutputCoins: []*InputOutputCoinsArgs{
					{
						ctxHasQuarantineBypass: false,
						inputs: []banktypes.Input{
							{Address: s.addr1.String(), Coins: s.coins("2fig")},
							{Address: s.addr4.String(), Coins: s.coins("1plum")},
						},
						outputs: []banktypes.Output{{Address: s.marketAddr2.String(), Coins: s.coins("2fig,1plum")}},
					},
				},
			},
		},
		{
			name:       "error collecting creation fee",
			bankKeeper: NewMockBankKeeper().WithSendCoinsResults("", "", "another error for testing"),
			setup: func() {
				s.k.SetParams(s.ctx, &exchange.Params{})
				s.requireCreateMarket(exchange.Market{MarketId: 2, AcceptingOrders: true, AllowUserSettlement: true})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(99).WithBid(&exchange.BidOrder{
					Assets: s.coin("1apple"), Price: s.coin("6plum"), MarketId: 2, Buyer: s.addr1.String(),
				}))
			},
			msg: exchange.MsgFillBidsRequest{
				Seller:              s.addr4.String(),
				MarketId:            2,
				TotalAssets:         s.coins("1apple"),
				BidOrderIds:         []uint64{99},
				AskOrderCreationFee: s.coinP("2fig"),
			},
			expErr: "error collecting create-ask fee \"2fig\": error transferring 2fig from " + s.addr4.String() +
				" to market 2: another error for testing",
			expEvents: []*exchange.EventOrderFilled{
				{OrderId: 99, Assets: "1apple", Price: "6plum", MarketId: 2},
			},
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{{addr: s.addr1, funds: s.coins("6plum")}}},
			expBankCalls: BankCalls{
				SendCoins: []*SendCoinsArgs{
					{ctxHasQuarantineBypass: true, fromAddr: s.addr4, toAddr: s.addr1, amt: s.coins("1apple")},
					{ctxHasQuarantineBypass: true, fromAddr: s.addr1, toAddr: s.addr4, amt: s.coins("6plum")},
					{ctxHasQuarantineBypass: false, fromAddr: s.addr4, toAddr: s.marketAddr2, amt: s.coins("2fig")},
				},
			},
		},

		// Tests on successes.
		{
			name: "one order: no fees",
			setup: func() {
				s.requireCreateMarket(exchange.Market{MarketId: 6, AcceptingOrders: true, AllowUserSettlement: true})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(13).WithBid(&exchange.BidOrder{
					Assets: s.coin("12apple"), Price: s.coin("60plum"), MarketId: 6, Buyer: s.addr2.String(),
				}))
			},
			msg: exchange.MsgFillBidsRequest{
				Seller:      s.addr5.String(),
				MarketId:    6,
				TotalAssets: s.coins("12apple"),
				BidOrderIds: []uint64{13},
			},
			expEvents: []*exchange.EventOrderFilled{
				{OrderId: 13, Assets: "12apple", Price: "60plum", MarketId: 6},
			},
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{{addr: s.addr2, funds: s.coins("60plum")}}},
			expBankCalls: BankCalls{
				SendCoins: []*SendCoinsArgs{
					{ctxHasQuarantineBypass: true, fromAddr: s.addr5, toAddr: s.addr2, amt: s.coins("12apple")},
					{ctxHasQuarantineBypass: true, fromAddr: s.addr2, toAddr: s.addr5, amt: s.coins("60plum")},
				},
			},
		},
		{
			name: "one order: all the fees",
			setup: func() {
				s.k.SetParams(s.ctx, &exchange.Params{
					DefaultSplit: 1000,
					DenomSplits:  []exchange.DenomSplit{{Denom: "fig", Split: 2000}},
				})
				s.requireCreateMarket(exchange.Market{
					MarketId: 3, AcceptingOrders: true, AllowUserSettlement: true,
					FeeSellerSettlementRatios: s.ratios("30plum:1plum"),
				})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(13).WithBid(&exchange.BidOrder{
					Assets: s.coin("12apple"), Price: s.coin("60plum"), MarketId: 3, Buyer: s.addr2.String(),
					BuyerSettlementFees: s.coins("10fig"),
					ExternalId:          "thirteen",
				}))
			},
			msg: exchange.MsgFillBidsRequest{
				Seller:                  s.addr5.String(),
				MarketId:                3,
				TotalAssets:             s.coins("12apple"),
				BidOrderIds:             []uint64{13},
				SellerSettlementFlatFee: s.coinP("8plum"),
				AskOrderCreationFee:     s.coinP("15fig"),
			},
			expEvents: []*exchange.EventOrderFilled{
				{OrderId: 13, Assets: "12apple", Price: "60plum", Fees: "10fig", MarketId: 3, ExternalId: "thirteen"},
			},
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{{addr: s.addr2, funds: s.coins("10fig,60plum")}}},
			expBankCalls: BankCalls{
				SendCoins: []*SendCoinsArgs{
					{ctxHasQuarantineBypass: true, fromAddr: s.addr5, toAddr: s.addr2, amt: s.coins("12apple")},
					{ctxHasQuarantineBypass: true, fromAddr: s.addr2, toAddr: s.addr5, amt: s.coins("60plum")},
					{ctxHasQuarantineBypass: false, fromAddr: s.addr5, toAddr: s.marketAddr3, amt: s.coins("15fig")},
				},
				InputOutputCoins: []*InputOutputCoinsArgs{
					{
						inputs: []banktypes.Input{
							{Address: s.addr2.String(), Coins: s.coins("10fig")},
							{Address: s.addr5.String(), Coins: s.coins("10plum")},
						},
						outputs: []banktypes.Output{{Address: s.marketAddr3.String(), Coins: s.coins("10fig,10plum")}},
					},
				},
				SendCoinsFromAccountToModule: []*SendCoinsFromAccountToModuleArgs{
					{senderAddr: s.marketAddr3, recipientModule: s.feeCollector, amt: s.coins("2fig,1plum")},
					{senderAddr: s.marketAddr3, recipientModule: s.feeCollector, amt: s.coins("3fig")},
				},
			},
		},
		{
			name: "three orders",
			setup: func() {
				s.k.SetParams(s.ctx, &exchange.Params{DefaultSplit: 1000})
				s.requireCreateMarket(exchange.Market{
					MarketId: 3, AcceptingOrders: true, AllowUserSettlement: true,
					FeeSellerSettlementRatios: s.ratios("30plum:1plum,88prune:5prune"),
				})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(17).WithBid(&exchange.BidOrder{
					Assets: s.coin("12apple"), Price: s.coin("60plum"), MarketId: 3, Buyer: s.addr2.String(),
				}))
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(55).WithBid(&exchange.BidOrder{
					Assets: s.coin("5acorn"), Price: s.coin("50prune"), MarketId: 3, Buyer: s.addr2.String(),
					BuyerSettlementFees: s.coins("22fig"),
				}))
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(121).WithBid(&exchange.BidOrder{
					Assets: s.coin("6apple"), Price: s.coin("33prune"), MarketId: 3, Buyer: s.addr3.String(),
				}))
			},
			msg: exchange.MsgFillBidsRequest{
				Seller:      s.addr1.String(),
				MarketId:    3,
				TotalAssets: s.coins("5acorn,18apple"),
				BidOrderIds: []uint64{55, 121, 17},
			},
			expEvents: []*exchange.EventOrderFilled{
				{OrderId: 55, Assets: "5acorn", Price: "50prune", MarketId: 3, Fees: "22fig"},
				{OrderId: 121, Assets: "6apple", Price: "33prune", MarketId: 3},
				{OrderId: 17, Assets: "12apple", Price: "60plum", MarketId: 3},
			},
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{
				{addr: s.addr2, funds: s.coins("22fig,50prune")},
				{addr: s.addr3, funds: s.coins("33prune")},
				{addr: s.addr2, funds: s.coins("60plum")},
			}},
			expBankCalls: BankCalls{
				InputOutputCoins: []*InputOutputCoinsArgs{
					{
						ctxHasQuarantineBypass: true,
						inputs:                 []banktypes.Input{{Address: s.addr1.String(), Coins: s.coins("5acorn,18apple")}},
						outputs: []banktypes.Output{
							{Address: s.addr2.String(), Coins: s.coins("5acorn,12apple")},
							{Address: s.addr3.String(), Coins: s.coins("6apple")},
						},
					},
					{
						ctxHasQuarantineBypass: true,
						inputs: []banktypes.Input{
							{Address: s.addr2.String(), Coins: s.coins("60plum,50prune")},
							{Address: s.addr3.String(), Coins: s.coins("33prune")},
						},
						outputs: []banktypes.Output{{Address: s.addr1.String(), Coins: s.coins("60plum,83prune")}},
					},
					{
						ctxHasQuarantineBypass: false,
						inputs: []banktypes.Input{
							{Address: s.addr2.String(), Coins: s.coins("22fig")},
							{Address: s.addr1.String(), Coins: s.coins("2plum,5prune")},
						},
						outputs: []banktypes.Output{{Address: s.marketAddr3.String(), Coins: s.coins("22fig,2plum,5prune")}},
					},
				},
				SendCoinsFromAccountToModule: []*SendCoinsFromAccountToModuleArgs{
					{senderAddr: s.marketAddr3, recipientModule: s.feeCollector, amt: s.coins("3fig,1plum,1prune")},
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			if tc.attrKeeper == nil {
				tc.attrKeeper = NewMockAttributeKeeper()
			}
			if s.accKeeper == nil {
				s.accKeeper = NewMockAccountKeeper()
			}
			if tc.bankKeeper == nil {
				tc.bankKeeper = NewMockBankKeeper()
			}
			if tc.holdKeeper == nil {
				tc.holdKeeper = NewMockHoldKeeper()
			}

			expEvents := make(sdk.Events, len(tc.expEvents))
			for i, expEvent := range tc.expEvents {
				event, err := sdk.TypedEventToEvent(expEvent)
				s.Require().NoError(err, "[%d]TypedEventToEvent")
				expEvents[i] = event
			}

			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)
			kpr := s.k.WithAttributeKeeper(tc.attrKeeper).
				WithAccountKeeper(s.accKeeper).
				WithBankKeeper(tc.bankKeeper).
				WithHoldKeeper(tc.holdKeeper)
			var err error
			testFunc := func() {
				err = kpr.FillBids(ctx, &tc.msg)
			}
			s.Require().NotPanics(testFunc, "FillBids")
			s.assertErrorValue(err, tc.expErr, "FillBids error")
			actEvents := em.Events()
			s.assertEqualEvents(expEvents, actEvents, "FillBids events")
			s.assertAttributeKeeperCalls(tc.attrKeeper, tc.expAttrCalls, "FillBids")
			s.assertHoldKeeperCalls(tc.holdKeeper, tc.expHoldCalls, "FillBids")
			s.assertBankKeeperCalls(tc.bankKeeper, tc.expBankCalls, "FillBids")

			if len(actEvents) == 0 {
				return
			}

			// Make sure all the orders have been deleted.
			for _, orderID := range tc.msg.BidOrderIds {
				order, oerr := s.k.GetOrder(s.ctx, orderID)
				s.Assert().NoError(oerr, "GetOrder(%d) after FillBids", orderID)
				s.Assert().Nil(order, "GetOrder(%d) after FillBids", orderID)
			}
		})
	}
}

func (s *TestSuite) TestKeeper_FillAsks() {
	tests := []struct {
		name         string
		attrKeeper   *MockAttributeKeeper
		bankKeeper   *MockBankKeeper
		holdKeeper   *MockHoldKeeper
		setup        func()
		msg          exchange.MsgFillAsksRequest
		expErr       string
		expEvents    []*exchange.EventOrderFilled
		expAttrCalls AttributeCalls
		expHoldCalls HoldCalls
		expBankCalls BankCalls
	}{
		// Tests on error conditions.
		{
			name: "invalid msg",
			msg: exchange.MsgFillAsksRequest{
				Buyer:       s.addr1.String(),
				MarketId:    0,
				TotalPrice:  s.coin("1prune"),
				AskOrderIds: []uint64{1},
			},
			expErr: "invalid market id: cannot be zero",
		},
		{
			name: "market does not exist",
			msg: exchange.MsgFillAsksRequest{
				Buyer:       s.addr1.String(),
				MarketId:    1,
				TotalPrice:  s.coin("1prune"),
				AskOrderIds: []uint64{1},
			},
			expErr: "market 1 does not exist",
		},
		{
			name: "market not accepting orders",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:        1,
					AcceptingOrders: false,
				})
			},
			msg: exchange.MsgFillAsksRequest{
				Buyer:       s.addr1.String(),
				MarketId:    1,
				TotalPrice:  s.coin("1prune"),
				AskOrderIds: []uint64{1},
			},
			expErr: "market 1 is not accepting orders",
		},
		{
			name: "market does not allow user-settle",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId: 1, AcceptingOrders: true,
					AllowUserSettlement: false,
				})
			},
			msg: exchange.MsgFillAsksRequest{
				Buyer:       s.addr1.String(),
				MarketId:    1,
				TotalPrice:  s.coin("1prune"),
				AskOrderIds: []uint64{1},
			},
			expErr: "market 1 does not allow user settlement",
		},
		{
			name: "buyer cannot create bid",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId: 1, AcceptingOrders: true, AllowUserSettlement: true,
					ReqAttrCreateBid: []string{"some.attr.no.one.has"},
				})
			},
			msg: exchange.MsgFillAsksRequest{
				Buyer:       s.addr1.String(),
				MarketId:    1,
				TotalPrice:  s.coin("1prune"),
				AskOrderIds: []uint64{1},
			},
			expErr:       "account " + s.addr1.String() + " is not allowed to create bid orders in market 1",
			expAttrCalls: AttributeCalls{GetAllAttributesAddr: [][]byte{s.addr1}},
		},
		{
			name: "not enough creation fee",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId: 1, AcceptingOrders: true, AllowUserSettlement: true,
					FeeCreateBidFlat: s.coins("5fig"),
				})
			},
			msg: exchange.MsgFillAsksRequest{
				Buyer:               s.addr1.String(),
				MarketId:            1,
				TotalPrice:          s.coin("1prune"),
				AskOrderIds:         []uint64{1},
				BidOrderCreationFee: s.coinP("4fig"),
			},
			expErr: "insufficient bid order creation fee: \"4fig\" is less than required amount \"5fig\"",
		},
		{
			name: "not enough buyer settlement fee",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId: 1, AcceptingOrders: true, AllowUserSettlement: true,
					FeeBuyerSettlementFlat: s.coins("5fig"),
				})
			},
			msg: exchange.MsgFillAsksRequest{
				Buyer:               s.addr1.String(),
				MarketId:            1,
				TotalPrice:          s.coin("1prune"),
				AskOrderIds:         []uint64{1},
				BuyerSettlementFees: s.coins("4fig"),
			},
			expErr: s.joinErrs(
				"4fig is less than required flat fee 5fig",
				"required flat fee not satisfied, valid options: 5fig",
				"insufficient buyer settlement fee 4fig",
			),
		},
		{
			name: "ask order does not exist",
			setup: func() {
				s.requireCreateMarket(exchange.Market{MarketId: 1, AcceptingOrders: true, AllowUserSettlement: true})
			},
			msg: exchange.MsgFillAsksRequest{
				Buyer:       s.addr1.String(),
				MarketId:    1,
				TotalPrice:  s.coin("1prune"),
				AskOrderIds: []uint64{8},
			},
			expErr: "order 8 not found",
		},
		{
			name: "bid order id provided",
			setup: func() {
				s.requireCreateMarket(exchange.Market{MarketId: 1, AcceptingOrders: true, AllowUserSettlement: true})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(8).WithBid(&exchange.BidOrder{
					MarketId: 1,
					Buyer:    s.addr2.String(),
					Assets:   s.coin("1apple"),
					Price:    s.coin("1plum"),
				}))
			},
			msg: exchange.MsgFillAsksRequest{
				Buyer:       s.addr1.String(),
				MarketId:    1,
				TotalPrice:  s.coin("1plum"),
				AskOrderIds: []uint64{8},
			},
			expErr: "order 8 is type bid: expected ask",
		},
		{
			name: "order in wrong market",
			setup: func() {
				s.requireCreateMarket(exchange.Market{MarketId: 1, AcceptingOrders: true, AllowUserSettlement: true})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(8).WithAsk(&exchange.AskOrder{
					MarketId: 2,
					Seller:   s.addr2.String(),
					Assets:   s.coin("1apple"),
					Price:    s.coin("1plum"),
				}))
			},
			msg: exchange.MsgFillAsksRequest{
				Buyer:       s.addr1.String(),
				MarketId:    1,
				TotalPrice:  s.coin("1plum"),
				AskOrderIds: []uint64{8},
			},
			expErr: "order 8 market id 2 does not equal requested market id 1",
		},
		{
			name: "order has same seller as provided buyer",
			setup: func() {
				s.requireCreateMarket(exchange.Market{MarketId: 1, AcceptingOrders: true, AllowUserSettlement: true})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(8).WithAsk(&exchange.AskOrder{
					MarketId: 1,
					Seller:   s.addr1.String(),
					Assets:   s.coin("1apple"),
					Price:    s.coin("1plum"),
				}))
			},
			msg: exchange.MsgFillAsksRequest{
				Buyer:       s.addr1.String(),
				MarketId:    1,
				TotalPrice:  s.coin("1plum"),
				AskOrderIds: []uint64{8},
			},
			expErr: "order 8 has the same seller " + s.addr1.String() + " as the requested buyer",
		},
		{
			name: "multiple problems with orders",
			setup: func() {
				s.requireCreateMarket(exchange.Market{MarketId: 1, AcceptingOrders: true, AllowUserSettlement: true})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(3).WithAsk(&exchange.AskOrder{
					MarketId: 2,
					Seller:   s.addr2.String(),
					Assets:   s.coin("1apple"),
					Price:    s.coin("1plum"),
				}))
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(17).WithBid(&exchange.BidOrder{
					MarketId: 1,
					Buyer:    s.addr1.String(),
					Assets:   s.coin("1apple"),
					Price:    s.coin("1plum"),
				}))
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(11).WithAsk(&exchange.AskOrder{
					MarketId: 1,
					Seller:   s.addr1.String(),
					Assets:   s.coin("1apple"),
					Price:    s.coin("1plum"),
				}))
			},
			msg: exchange.MsgFillAsksRequest{
				Buyer:       s.addr1.String(),
				MarketId:    1,
				TotalPrice:  s.coin("1plum"),
				AskOrderIds: []uint64{8, 3, 17, 11},
			},
			expErr: s.joinErrs(
				"order 8 not found",
				"order 3 market id 2 does not equal requested market id 1",
				"order 17 is type bid: expected ask",
				"order 11 has the same seller "+s.addr1.String()+" as the requested buyer",
			),
		},
		{
			name: "provided total price less than actual total price",
			setup: func() {
				s.requireCreateMarket(exchange.Market{MarketId: 2, AcceptingOrders: true, AllowUserSettlement: true})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
					Assets: s.coin("1apple"), Price: s.coin("1plum"), MarketId: 2, Seller: s.addr1.String(),
				}))
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(2).WithAsk(&exchange.AskOrder{
					Assets: s.coin("2apple"), Price: s.coin("2plum"), MarketId: 2, Seller: s.addr2.String(),
				}))
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(3).WithAsk(&exchange.AskOrder{
					Assets: s.coin("3apple"), Price: s.coin("3plum"), MarketId: 2, Seller: s.addr3.String(),
				}))
			},
			msg: exchange.MsgFillAsksRequest{
				Buyer:       s.addr4.String(),
				MarketId:    2,
				TotalPrice:  s.coin("5plum"),
				AskOrderIds: []uint64{1, 2, 3},
			},
			expErr: "total price \"5plum\" does not equal sum of ask order prices \"6plum\"",
		},
		{
			name: "provided total price more than actual total price",
			setup: func() {
				s.requireCreateMarket(exchange.Market{MarketId: 2, AcceptingOrders: true, AllowUserSettlement: true})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
					Assets: s.coin("1apple"), Price: s.coin("1plum"), MarketId: 2, Seller: s.addr1.String(),
				}))
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(2).WithAsk(&exchange.AskOrder{
					Assets: s.coin("2apple"), Price: s.coin("2plum"), MarketId: 2, Seller: s.addr2.String(),
				}))
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(3).WithAsk(&exchange.AskOrder{
					Assets: s.coin("3apple"), Price: s.coin("3plum"), MarketId: 2, Seller: s.addr3.String(),
				}))
			},
			msg: exchange.MsgFillAsksRequest{
				Buyer:       s.addr4.String(),
				MarketId:    2,
				TotalPrice:  s.coin("7plum"),
				AskOrderIds: []uint64{1, 2, 3},
			},
			expErr: "total price \"7plum\" does not equal sum of ask order prices \"6plum\"",
		},
		{
			name: "ratio fee calc error",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId: 2, AcceptingOrders: true, AllowUserSettlement: true,
					FeeSellerSettlementRatios: s.ratios("20prune:1prune"),
				})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
					Assets: s.coin("6apple"), Price: s.coin("6plum"), MarketId: 2, Seller: s.addr1.String(),
				}))
			},
			msg: exchange.MsgFillAsksRequest{
				Buyer:       s.addr4.String(),
				MarketId:    2,
				TotalPrice:  s.coin("6plum"),
				AskOrderIds: []uint64{1},
			},
			expErr: "error calculating seller settlement ratio fee for order 1: no seller " +
				"settlement fee ratio found for denom \"plum\"",
		},
		{
			name: "invalid bid order owner",
			setup: func() {
				s.requireCreateMarket(exchange.Market{MarketId: 2, AcceptingOrders: true, AllowUserSettlement: true})
				key, value, err := s.k.GetOrderStoreKeyValue(*exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
					Assets: s.coin("6apple"), Price: s.coin("6plum"), MarketId: 2, Seller: "badseller",
				}))
				s.Require().NoError(err, "GetOrderStoreKeyValue 1")
				s.getStore().Set(key, value)
			},
			msg: exchange.MsgFillAsksRequest{
				Buyer:       s.addr4.String(),
				MarketId:    2,
				TotalPrice:  s.coin("6plum"),
				AskOrderIds: []uint64{1},
			},
			expErr: "invalid ask order 1 owner \"badseller\": decoding bech32 failed: invalid separator index -1",
		},
		{
			name:       "error releasing hold",
			holdKeeper: NewMockHoldKeeper().WithReleaseHoldResults("no apple for you"),
			setup: func() {
				s.requireCreateMarket(exchange.Market{MarketId: 2, AcceptingOrders: true, AllowUserSettlement: true})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
					Assets: s.coin("6apple"), Price: s.coin("6plum"), MarketId: 2, Seller: s.addr1.String(),
				}))
			},
			msg: exchange.MsgFillAsksRequest{
				Buyer:       s.addr4.String(),
				MarketId:    2,
				TotalPrice:  s.coin("6plum"),
				AskOrderIds: []uint64{1},
			},
			expErr:       "error releasing hold for ask order 1: no apple for you",
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{{addr: s.addr1, funds: s.coins("6apple")}}},
		},
		{
			name:       "error transferring assets",
			bankKeeper: NewMockBankKeeper().WithSendCoinsResults("first transfer error"),
			setup: func() {
				s.requireCreateMarket(exchange.Market{MarketId: 2, AcceptingOrders: true, AllowUserSettlement: true})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
					Assets: s.coin("1apple"), Price: s.coin("6plum"), MarketId: 2, Seller: s.addr1.String(),
				}))
			},
			msg: exchange.MsgFillAsksRequest{
				Buyer:       s.addr4.String(),
				MarketId:    2,
				TotalPrice:  s.coin("6plum"),
				AskOrderIds: []uint64{1},
			},
			expErr:       "first transfer error",
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{{addr: s.addr1, funds: s.coins("1apple")}}},
			expBankCalls: BankCalls{SendCoins: []*SendCoinsArgs{
				{ctxHasQuarantineBypass: true, fromAddr: s.addr1, toAddr: s.addr4, amt: s.coins("1apple")},
				{ctxHasQuarantineBypass: true, fromAddr: s.addr4, toAddr: s.addr1, amt: s.coins("6plum")},
			}},
		},
		{
			name:       "error transferring price",
			bankKeeper: NewMockBankKeeper().WithSendCoinsResults("", "second transfer error"),
			setup: func() {
				s.k.SetParams(s.ctx, &exchange.Params{})
				s.requireCreateMarket(exchange.Market{MarketId: 2, AcceptingOrders: true, AllowUserSettlement: true})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
					Assets: s.coin("1apple"), Price: s.coin("6plum"), MarketId: 2, Seller: s.addr1.String(),
				}))
			},
			msg: exchange.MsgFillAsksRequest{
				Buyer:       s.addr4.String(),
				MarketId:    2,
				TotalPrice:  s.coin("6plum"),
				AskOrderIds: []uint64{1},
			},
			expErr:       "second transfer error",
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{{addr: s.addr1, funds: s.coins("1apple")}}},
			expBankCalls: BankCalls{SendCoins: []*SendCoinsArgs{
				{ctxHasQuarantineBypass: true, fromAddr: s.addr1, toAddr: s.addr4, amt: s.coins("1apple")},
				{ctxHasQuarantineBypass: true, fromAddr: s.addr4, toAddr: s.addr1, amt: s.coins("6plum")},
			}},
		},
		{
			name:       "error collecting settlement fees",
			bankKeeper: NewMockBankKeeper().WithInputOutputCoinsResults("first fake error"),
			setup: func() {
				s.k.SetParams(s.ctx, &exchange.Params{})
				s.requireCreateMarket(exchange.Market{
					MarketId: 2, AcceptingOrders: true, AllowUserSettlement: true,
					FeeSellerSettlementRatios: s.ratios("6plum:1plum"),
					FeeBuyerSettlementRatios:  s.ratios("6plum:2fig"),
				})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(99).WithAsk(&exchange.AskOrder{
					Assets: s.coin("1apple"), Price: s.coin("6plum"), MarketId: 2, Seller: s.addr1.String(),
					SellerSettlementFlatFee: s.coinP("2fig"),
				}))
			},
			msg: exchange.MsgFillAsksRequest{
				Buyer:               s.addr4.String(),
				MarketId:            2,
				TotalPrice:          s.coin("6plum"),
				AskOrderIds:         []uint64{99},
				BuyerSettlementFees: s.coins("2fig"),
			},
			expErr:       "error collecting fees for market 2: first fake error",
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{{addr: s.addr1, funds: s.coins("2fig,1apple")}}},
			expBankCalls: BankCalls{
				SendCoins: []*SendCoinsArgs{
					{ctxHasQuarantineBypass: true, fromAddr: s.addr1, toAddr: s.addr4, amt: s.coins("1apple")},
					{ctxHasQuarantineBypass: true, fromAddr: s.addr4, toAddr: s.addr1, amt: s.coins("6plum")},
				},
				InputOutputCoins: []*InputOutputCoinsArgs{
					{
						ctxHasQuarantineBypass: false,
						inputs: []banktypes.Input{
							{Address: s.addr1.String(), Coins: s.coins("2fig,1plum")},
							{Address: s.addr4.String(), Coins: s.coins("2fig")},
						},
						outputs: []banktypes.Output{{Address: s.marketAddr2.String(), Coins: s.coins("4fig,1plum")}},
					},
				},
			},
		},
		{
			name:       "error collecting creation fee",
			bankKeeper: NewMockBankKeeper().WithSendCoinsResults("", "", "another error for testing"),
			setup: func() {
				s.k.SetParams(s.ctx, &exchange.Params{})
				s.requireCreateMarket(exchange.Market{MarketId: 2, AcceptingOrders: true, AllowUserSettlement: true})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(99).WithAsk(&exchange.AskOrder{
					Assets: s.coin("1apple"), Price: s.coin("6plum"), MarketId: 2, Seller: s.addr1.String(),
				}))
			},
			msg: exchange.MsgFillAsksRequest{
				Buyer:               s.addr4.String(),
				MarketId:            2,
				TotalPrice:          s.coin("6plum"),
				AskOrderIds:         []uint64{99},
				BidOrderCreationFee: s.coinP("2fig"),
			},
			expErr: "error collecting create-ask fee \"2fig\": error transferring 2fig from " + s.addr4.String() +
				" to market 2: another error for testing",
			expEvents: []*exchange.EventOrderFilled{
				{OrderId: 99, Assets: "1apple", Price: "6plum", MarketId: 2},
			},
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{{addr: s.addr1, funds: s.coins("1apple")}}},
			expBankCalls: BankCalls{
				SendCoins: []*SendCoinsArgs{
					{ctxHasQuarantineBypass: true, fromAddr: s.addr1, toAddr: s.addr4, amt: s.coins("1apple")},
					{ctxHasQuarantineBypass: true, fromAddr: s.addr4, toAddr: s.addr1, amt: s.coins("6plum")},
					{ctxHasQuarantineBypass: false, fromAddr: s.addr4, toAddr: s.marketAddr2, amt: s.coins("2fig")},
				},
			},
		},

		// Tests on successes.
		{
			name: "one order: no fees",
			setup: func() {
				s.requireCreateMarket(exchange.Market{MarketId: 6, AcceptingOrders: true, AllowUserSettlement: true})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(13).WithAsk(&exchange.AskOrder{
					Assets: s.coin("12apple"), Price: s.coin("60plum"), MarketId: 6, Seller: s.addr2.String(),
				}))
			},
			msg: exchange.MsgFillAsksRequest{
				Buyer:       s.addr5.String(),
				MarketId:    6,
				TotalPrice:  s.coin("60plum"),
				AskOrderIds: []uint64{13},
			},
			expEvents: []*exchange.EventOrderFilled{
				{OrderId: 13, Assets: "12apple", Price: "60plum", MarketId: 6},
			},
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{{addr: s.addr2, funds: s.coins("12apple")}}},
			expBankCalls: BankCalls{
				SendCoins: []*SendCoinsArgs{
					{ctxHasQuarantineBypass: true, fromAddr: s.addr2, toAddr: s.addr5, amt: s.coins("12apple")},
					{ctxHasQuarantineBypass: true, fromAddr: s.addr5, toAddr: s.addr2, amt: s.coins("60plum")},
				},
			},
		},
		{
			name: "one order: all the fees",
			setup: func() {
				s.k.SetParams(s.ctx, &exchange.Params{
					DefaultSplit: 1000,
					DenomSplits:  []exchange.DenomSplit{{Denom: "fig", Split: 2000}},
				})
				s.requireCreateMarket(exchange.Market{
					MarketId: 3, AcceptingOrders: true, AllowUserSettlement: true,
					FeeSellerSettlementRatios: s.ratios("30plum:1plum"),
				})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(13).WithAsk(&exchange.AskOrder{
					Assets: s.coin("12apple"), Price: s.coin("60plum"), MarketId: 3, Seller: s.addr2.String(),
					SellerSettlementFlatFee: s.coinP("8fig"),
					ExternalId:              "thirteen",
				}))
			},
			msg: exchange.MsgFillAsksRequest{
				Buyer:               s.addr5.String(),
				MarketId:            3,
				TotalPrice:          s.coin("60plum"),
				AskOrderIds:         []uint64{13},
				BuyerSettlementFees: s.coins("10plum"),
				BidOrderCreationFee: s.coinP("15fig"),
			},
			expEvents: []*exchange.EventOrderFilled{
				{OrderId: 13, Assets: "12apple", Price: "60plum", Fees: "8fig,2plum", MarketId: 3, ExternalId: "thirteen"},
			},
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{{addr: s.addr2, funds: s.coins("12apple,8fig")}}},
			expBankCalls: BankCalls{
				SendCoins: []*SendCoinsArgs{
					{ctxHasQuarantineBypass: true, fromAddr: s.addr2, toAddr: s.addr5, amt: s.coins("12apple")},
					{ctxHasQuarantineBypass: true, fromAddr: s.addr5, toAddr: s.addr2, amt: s.coins("60plum")},
					{ctxHasQuarantineBypass: false, fromAddr: s.addr5, toAddr: s.marketAddr3, amt: s.coins("15fig")},
				},
				InputOutputCoins: []*InputOutputCoinsArgs{
					{
						inputs: []banktypes.Input{
							{Address: s.addr2.String(), Coins: s.coins("8fig,2plum")},
							{Address: s.addr5.String(), Coins: s.coins("10plum")},
						},
						outputs: []banktypes.Output{{Address: s.marketAddr3.String(), Coins: s.coins("8fig,12plum")}},
					},
				},
				SendCoinsFromAccountToModule: []*SendCoinsFromAccountToModuleArgs{
					{senderAddr: s.marketAddr3, recipientModule: s.feeCollector, amt: s.coins("2fig,2plum")},
					{senderAddr: s.marketAddr3, recipientModule: s.feeCollector, amt: s.coins("3fig")},
				},
			},
		},
		{
			name: "three orders",
			setup: func() {
				s.k.SetParams(s.ctx, &exchange.Params{DefaultSplit: 1000})
				s.requireCreateMarket(exchange.Market{
					MarketId: 3, AcceptingOrders: true, AllowUserSettlement: true,
					FeeSellerSettlementRatios: s.ratios("143prune:5prune"),
				})
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(17).WithAsk(&exchange.AskOrder{
					Assets: s.coin("12apple"), Price: s.coin("60prune"), MarketId: 3, Seller: s.addr2.String(),
				}))
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(55).WithAsk(&exchange.AskOrder{
					Assets: s.coin("5acorn"), Price: s.coin("50prune"), MarketId: 3, Seller: s.addr2.String(),
					SellerSettlementFlatFee: s.coinP("22fig"),
				}))
				s.requireSetOrderInStore(s.getStore(), exchange.NewOrder(121).WithAsk(&exchange.AskOrder{
					Assets: s.coin("6apple"), Price: s.coin("33prune"), MarketId: 3, Seller: s.addr3.String(),
				}))
			},
			msg: exchange.MsgFillAsksRequest{
				Buyer:       s.addr1.String(),
				MarketId:    3,
				TotalPrice:  s.coin("143prune"),
				AskOrderIds: []uint64{55, 121, 17},
			},
			expEvents: []*exchange.EventOrderFilled{
				{OrderId: 55, Assets: "5acorn", Price: "50prune", MarketId: 3, Fees: "22fig,2prune"},
				{OrderId: 121, Assets: "6apple", Price: "33prune", MarketId: 3, Fees: "2prune"},
				{OrderId: 17, Assets: "12apple", Price: "60prune", MarketId: 3, Fees: "3prune"},
			},
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{
				{addr: s.addr2, funds: s.coins("5acorn,22fig")},
				{addr: s.addr3, funds: s.coins("6apple")},
				{addr: s.addr2, funds: s.coins("12apple")},
			}},
			expBankCalls: BankCalls{
				InputOutputCoins: []*InputOutputCoinsArgs{
					{
						ctxHasQuarantineBypass: true,
						inputs: []banktypes.Input{
							{Address: s.addr2.String(), Coins: s.coins("5acorn,12apple")},
							{Address: s.addr3.String(), Coins: s.coins("6apple")},
						},
						outputs: []banktypes.Output{{Address: s.addr1.String(), Coins: s.coins("5acorn,18apple")}},
					},
					{
						ctxHasQuarantineBypass: true,
						inputs: []banktypes.Input{
							{Address: s.addr1.String(), Coins: s.coins("143prune")},
						},
						outputs: []banktypes.Output{
							{Address: s.addr2.String(), Coins: s.coins("110prune")},
							{Address: s.addr3.String(), Coins: s.coins("33prune")},
						},
					},
					{
						ctxHasQuarantineBypass: false,
						inputs: []banktypes.Input{
							{Address: s.addr2.String(), Coins: s.coins("22fig,5prune")},
							{Address: s.addr3.String(), Coins: s.coins("2prune")},
						},
						outputs: []banktypes.Output{{Address: s.marketAddr3.String(), Coins: s.coins("22fig,7prune")}},
					},
				},
				SendCoinsFromAccountToModule: []*SendCoinsFromAccountToModuleArgs{
					{senderAddr: s.marketAddr3, recipientModule: s.feeCollector, amt: s.coins("3fig,1prune")},
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			if tc.attrKeeper == nil {
				tc.attrKeeper = NewMockAttributeKeeper()
			}
			if s.accKeeper == nil {
				s.accKeeper = NewMockAccountKeeper()
			}
			if tc.bankKeeper == nil {
				tc.bankKeeper = NewMockBankKeeper()
			}
			if tc.holdKeeper == nil {
				tc.holdKeeper = NewMockHoldKeeper()
			}

			expEvents := make(sdk.Events, len(tc.expEvents))
			for i, expEvent := range tc.expEvents {
				event, err := sdk.TypedEventToEvent(expEvent)
				s.Require().NoError(err, "[%d]TypedEventToEvent")
				expEvents[i] = event
			}

			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)
			kpr := s.k.WithAttributeKeeper(tc.attrKeeper).
				WithAccountKeeper(s.accKeeper).
				WithBankKeeper(tc.bankKeeper).
				WithHoldKeeper(tc.holdKeeper)
			var err error
			testFunc := func() {
				err = kpr.FillAsks(ctx, &tc.msg)
			}
			s.Require().NotPanics(testFunc, "FillAsks")
			s.assertErrorValue(err, tc.expErr, "FillAsks error")
			actEvents := em.Events()
			s.assertEqualEvents(expEvents, actEvents, "FillAsks events")
			s.assertAttributeKeeperCalls(tc.attrKeeper, tc.expAttrCalls, "FillAsks")
			s.assertHoldKeeperCalls(tc.holdKeeper, tc.expHoldCalls, "FillAsks")
			s.assertBankKeeperCalls(tc.bankKeeper, tc.expBankCalls, "FillAsks")

			if len(actEvents) == 0 {
				return
			}

			// Make sure all the orders have been deleted.
			for _, orderID := range tc.msg.AskOrderIds {
				order, oerr := s.k.GetOrder(s.ctx, orderID)
				s.Assert().NoError(oerr, "GetOrder(%d) after FillAsks", orderID)
				s.Assert().Nil(order, "GetOrder(%d) after FillAsks", orderID)
			}
		})
	}
}

func (s *TestSuite) TestKeeper_SettleOrders() {
	tests := []struct {
		name           string
		bankKeeper     *MockBankKeeper
		holdKeeper     *MockHoldKeeper
		setup          func()
		marketID       uint32
		askOrderIDs    []uint64
		bidOrderIDs    []uint64
		expectPartial  bool
		expErr         string
		expEvents      []proto.Message
		expPartialLeft *exchange.Order
		expHoldCalls   HoldCalls
		expBankCalls   BankCalls
	}{
		// Tests on error conditions.
		{
			name:          "market does not exist",
			marketID:      1,
			askOrderIDs:   []uint64{1},
			bidOrderIDs:   []uint64{2},
			expectPartial: false,
			expErr:        "market 1 does not exist",
		},
		{
			name: "errors getting orders",
			setup: func() {
				s.requireCreateMarket(exchange.Market{MarketId: 1})
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(2).WithBid(&exchange.BidOrder{
					Assets: s.coin("1apple"), Price: s.coin("6peach"), MarketId: 1, Buyer: s.addr1.String(),
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(3).WithAsk(&exchange.AskOrder{
					Assets: s.coin("1apple"), Price: s.coin("6peach"), MarketId: 2, Seller: s.addr2.String(),
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(5).WithAsk(&exchange.AskOrder{
					Assets: s.coin("1apple"), Price: s.coin("6peach"), MarketId: 1, Seller: s.addr3.String(),
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(6).WithBid(&exchange.BidOrder{
					Assets: s.coin("1apple"), Price: s.coin("6peach"), MarketId: 3, Buyer: s.addr4.String(),
				}))
			},
			marketID:      1,
			askOrderIDs:   []uint64{1, 2, 3},
			bidOrderIDs:   []uint64{4, 5, 6},
			expectPartial: false,
			expErr: s.joinErrs(
				"order 1 not found",
				"order 2 is type bid: expected ask",
				"order 3 market id 2 does not equal requested market id 1",
				"order 4 not found",
				"order 5 is type ask: expected bid",
				"order 6 market id 3 does not equal requested market id 1",
			),
		},
		{
			name: "errors building settlement",
			setup: func() {
				s.requireCreateMarket(exchange.Market{MarketId: 1})
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(3).WithAsk(&exchange.AskOrder{
					Assets: s.coin("1apple"), Price: s.coin("6peach"), MarketId: 1, Seller: s.addr1.String(),
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(2).WithBid(&exchange.BidOrder{
					Assets: s.coin("2acorn"), Price: s.coin("5plum"), MarketId: 1, Buyer: s.addr2.String(),
				}))
			},
			marketID:      1,
			askOrderIDs:   []uint64{3},
			bidOrderIDs:   []uint64{2},
			expectPartial: false,
			expErr: s.joinErrs(
				"cannot settle different ask \"1apple\" and bid \"2acorn\" asset denoms",
				"cannot settle different ask \"6peach\" and bid \"5plum\" price denoms",
			),
		},
		{
			name: "expect partial, full",
			setup: func() {
				s.requireCreateMarket(exchange.Market{MarketId: 1})
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(3).WithAsk(&exchange.AskOrder{
					Assets: s.coin("1apple"), Price: s.coin("6peach"), MarketId: 1, Seller: s.addr1.String(),
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(2).WithBid(&exchange.BidOrder{
					Assets: s.coin("1apple"), Price: s.coin("6peach"), MarketId: 1, Buyer: s.addr2.String(),
				}))
			},
			marketID:      1,
			askOrderIDs:   []uint64{3},
			bidOrderIDs:   []uint64{2},
			expectPartial: true,
			expErr:        "settlement unexpectedly resulted in all orders fully filled",
		},
		{
			name: "expect full, partial",
			setup: func() {
				s.requireCreateMarket(exchange.Market{MarketId: 1})
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(3).WithAsk(&exchange.AskOrder{
					Assets: s.coin("1apple"), Price: s.coin("6peach"), MarketId: 1, Seller: s.addr1.String(),
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(2).WithBid(&exchange.BidOrder{
					Assets: s.coin("2apple"), Price: s.coin("12peach"), MarketId: 1, Buyer: s.addr2.String(),
					AllowPartial: true,
				}))
			},
			marketID:      1,
			askOrderIDs:   []uint64{3},
			bidOrderIDs:   []uint64{2},
			expectPartial: false,
			expErr:        "settlement resulted in unexpected partial order 2",
		},
		{
			name: "errors releasing holds",
			holdKeeper: NewMockHoldKeeper().
				WithReleaseHoldResults("first hold error", "second error releasing hold", "hold error the third"),
			setup: func() {
				s.requireCreateMarket(exchange.Market{MarketId: 1})
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(3).WithAsk(&exchange.AskOrder{
					Assets: s.coin("4apple"), Price: s.coin("16peach"), MarketId: 1, Seller: s.addr1.String(),
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(2).WithBid(&exchange.BidOrder{
					Assets: s.coin("2apple"), Price: s.coin("8peach"), MarketId: 1, Buyer: s.addr2.String(),
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(4).WithBid(&exchange.BidOrder{
					Assets: s.coin("3apple"), Price: s.coin("12peach"), MarketId: 1, Buyer: s.addr3.String(),
					AllowPartial: true,
				}))
			},
			marketID:      1,
			askOrderIDs:   []uint64{3},
			bidOrderIDs:   []uint64{2, 4},
			expectPartial: true,
			expErr: s.joinErrs(
				"error releasing hold for ask order 3: first hold error",
				"error releasing hold for bid order 2: second error releasing hold",
				"error releasing hold for bid order 4: hold error the third",
			),
			expHoldCalls: HoldCalls{
				ReleaseHold: []*ReleaseHoldArgs{
					{addr: s.addr1, funds: s.coins("4apple")},
					{addr: s.addr2, funds: s.coins("8peach")},
					{addr: s.addr3, funds: s.coins("8peach")},
				},
			},
		},
		{
			name: "errors transferring stuff",
			bankKeeper: NewMockBankKeeper().
				WithSendCoinsResults("first send error", "second send error").
				WithSendCoinsFromAccountToModuleResults("and a fee collection error too"),
			setup: func() {
				s.k.SetParams(s.ctx, &exchange.Params{
					DefaultSplit: 1000,
					DenomSplits:  []exchange.DenomSplit{{Denom: "grape", Split: 5000}},
				})
				s.requireCreateMarket(exchange.Market{MarketId: 1})
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(3).WithAsk(&exchange.AskOrder{
					Assets: s.coin("4apple"), Price: s.coin("16peach"), MarketId: 1, Seller: s.addr1.String(),
					SellerSettlementFlatFee: s.coinP("100fig"),
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(2).WithBid(&exchange.BidOrder{
					Assets: s.coin("4apple"), Price: s.coin("16peach"), MarketId: 1, Buyer: s.addr2.String(),
					BuyerSettlementFees: s.coins("50grape"),
				}))
			},
			marketID:      1,
			askOrderIDs:   []uint64{3},
			bidOrderIDs:   []uint64{2},
			expectPartial: false,
			expErr: s.joinErrs(
				"first send error",
				"second send error",
				"error collecting exchange fee 10fig,25grape (based off 100fig,50grape) from market 1: "+
					"and a fee collection error too",
			),
			expHoldCalls: HoldCalls{
				ReleaseHold: []*ReleaseHoldArgs{
					{addr: s.addr1, funds: s.coins("100fig,4apple")},
					{addr: s.addr2, funds: s.coins("50grape,16peach")},
				},
			},
			expBankCalls: BankCalls{
				SendCoins: []*SendCoinsArgs{
					{ctxHasQuarantineBypass: true, fromAddr: s.addr1, toAddr: s.addr2, amt: s.coins("4apple")},
					{ctxHasQuarantineBypass: true, fromAddr: s.addr2, toAddr: s.addr1, amt: s.coins("16peach")},
				},
				InputOutputCoins: []*InputOutputCoinsArgs{
					{
						ctxHasQuarantineBypass: false,
						inputs: []banktypes.Input{
							{Address: s.addr1.String(), Coins: s.coins("100fig")},
							{Address: s.addr2.String(), Coins: s.coins("50grape")},
						},
						outputs: []banktypes.Output{{Address: s.marketAddr1.String(), Coins: s.coins("100fig,50grape")}},
					},
				},
				SendCoinsFromAccountToModule: []*SendCoinsFromAccountToModuleArgs{
					{
						senderAddr:      s.marketAddr1,
						recipientModule: s.feeCollector,
						amt:             s.coins("10fig,25grape"),
					},
				},
			},
		},
		{
			name: "error updating partial",
			setup: func() {
				s.k.SetParams(s.ctx, &exchange.Params{})
				s.requireCreateMarket(exchange.Market{MarketId: 1})
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(5).WithBid(&exchange.BidOrder{
					Assets: s.coin("1apple"), Price: s.coin("5peach"), MarketId: 1, Buyer: s.addr4.String(),
					ExternalId: "oops-dup-id",
				}))
				key8, value8, err := s.k.GetOrderStoreKeyValue(*exchange.NewOrder(8).WithAsk(&exchange.AskOrder{
					MarketId:     1,
					Seller:       s.addr5.String(),
					Assets:       s.coin("10apple"),
					Price:        s.coin("50peach"),
					AllowPartial: true,
					ExternalId:   "oops-dup-id",
				}))
				s.Require().NoError(err, "GetOrderStoreKeyValue 8")
				store.Set(key8, value8)
			},
			marketID:      1,
			askOrderIDs:   []uint64{8},
			bidOrderIDs:   []uint64{5},
			expectPartial: true,
			expErr: "could not update partial ask order 8: external id \"oops-dup-id\" is " +
				"already in use by order 5: cannot be used for order 8",
			expHoldCalls: HoldCalls{
				ReleaseHold: []*ReleaseHoldArgs{
					{addr: s.addr4, funds: s.coins("5peach")},
					{addr: s.addr5, funds: s.coins("1apple")},
				},
			},
			expBankCalls: BankCalls{
				SendCoins: []*SendCoinsArgs{
					{ctxHasQuarantineBypass: true, fromAddr: s.addr5, toAddr: s.addr4, amt: s.coins("1apple")},
					{ctxHasQuarantineBypass: true, fromAddr: s.addr4, toAddr: s.addr5, amt: s.coins("5peach")},
				},
			},
		},

		// Tests on successes.
		{
			name: "one ask one bid: both full, no fees",
			setup: func() {
				s.requireCreateMarket(exchange.Market{MarketId: 1})
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
					Assets: s.coin("1apple"), Price: s.coin("5peach"), MarketId: 1, Seller: s.addr3.String(),
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(5).WithBid(&exchange.BidOrder{
					Assets: s.coin("1apple"), Price: s.coin("5peach"), MarketId: 1, Buyer: s.addr4.String(),
				}))
			},
			marketID:      1,
			askOrderIDs:   []uint64{1},
			bidOrderIDs:   []uint64{5},
			expectPartial: false,
			expEvents: []proto.Message{
				&exchange.EventOrderFilled{OrderId: 1, Assets: "1apple", Price: "5peach", MarketId: 1},
				&exchange.EventOrderFilled{OrderId: 5, Assets: "1apple", Price: "5peach", MarketId: 1},
			},
			expHoldCalls: HoldCalls{
				ReleaseHold: []*ReleaseHoldArgs{
					{addr: s.addr3, funds: s.coins("1apple")},
					{addr: s.addr4, funds: s.coins("5peach")},
				},
			},
			expBankCalls: BankCalls{
				SendCoins: []*SendCoinsArgs{
					{ctxHasQuarantineBypass: true, fromAddr: s.addr3, toAddr: s.addr4, amt: s.coins("1apple")},
					{ctxHasQuarantineBypass: true, fromAddr: s.addr4, toAddr: s.addr3, amt: s.coins("5peach")},
				},
			},
		},
		{
			name: "one ask one bid: both full, all the fees",
			setup: func() {
				s.k.SetParams(s.ctx, &exchange.Params{DefaultSplit: 1000})
				s.requireCreateMarket(exchange.Market{
					MarketId:                  1,
					FeeSellerSettlementRatios: s.ratios("25peach:1peach"),
				})
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
					Assets: s.coin("10apple"), Price: s.coin("50peach"), MarketId: 1, Seller: s.addr3.String(),
					SellerSettlementFlatFee: s.coinP("3peach"),
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(5).WithBid(&exchange.BidOrder{
					Assets: s.coin("10apple"), Price: s.coin("50peach"), MarketId: 1, Buyer: s.addr4.String(),
					BuyerSettlementFees: s.coins("15peach"),
				}))
			},
			marketID:      1,
			askOrderIDs:   []uint64{1},
			bidOrderIDs:   []uint64{5},
			expectPartial: false,
			expEvents: []proto.Message{
				&exchange.EventOrderFilled{OrderId: 1, Assets: "10apple", Price: "50peach", MarketId: 1, Fees: "5peach"},
				&exchange.EventOrderFilled{OrderId: 5, Assets: "10apple", Price: "50peach", MarketId: 1, Fees: "15peach"},
			},
			expHoldCalls: HoldCalls{
				ReleaseHold: []*ReleaseHoldArgs{
					{addr: s.addr3, funds: s.coins("10apple")},
					{addr: s.addr4, funds: s.coins("65peach")},
				},
			},
			expBankCalls: BankCalls{
				SendCoins: []*SendCoinsArgs{
					{ctxHasQuarantineBypass: true, fromAddr: s.addr3, toAddr: s.addr4, amt: s.coins("10apple")},
					{ctxHasQuarantineBypass: true, fromAddr: s.addr4, toAddr: s.addr3, amt: s.coins("50peach")},
				},
				InputOutputCoins: []*InputOutputCoinsArgs{
					{
						inputs: []banktypes.Input{
							{Address: s.addr3.String(), Coins: s.coins("5peach")},
							{Address: s.addr4.String(), Coins: s.coins("15peach")},
						},
						outputs: []banktypes.Output{{Address: s.marketAddr1.String(), Coins: s.coins("20peach")}},
					},
				},
				SendCoinsFromAccountToModule: []*SendCoinsFromAccountToModuleArgs{
					{senderAddr: s.marketAddr1, recipientModule: s.feeCollector, amt: s.coins("2peach")},
				},
			},
		},
		{
			name: "one ask one bid: partial ask",
			setup: func() {
				s.k.SetParams(s.ctx, &exchange.Params{})
				s.requireCreateMarket(exchange.Market{MarketId: 1})
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
					Assets: s.coin("10apple"), Price: s.coin("50peach"), MarketId: 1, Seller: s.addr5.String(),
					SellerSettlementFlatFee: s.coinP("20fig"),
					ExternalId:              "the-ask-order",
					AllowPartial:            true,
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(2).WithBid(&exchange.BidOrder{
					Assets: s.coin("7apple"), Price: s.coin("40peach"), MarketId: 1, Buyer: s.addr3.String(),
					ExternalId: "the-bid-order",
				}))
			},
			marketID:      1,
			askOrderIDs:   []uint64{1},
			bidOrderIDs:   []uint64{2},
			expectPartial: true,
			expEvents: []proto.Message{
				&exchange.EventOrderFilled{
					OrderId: 2, Assets: "7apple", Price: "40peach",
					MarketId: 1, ExternalId: "the-bid-order",
				},
				&exchange.EventOrderPartiallyFilled{
					OrderId: 1, Assets: "7apple", Price: "40peach", Fees: "14fig",
					MarketId: 1, ExternalId: "the-ask-order",
				},
			},
			expPartialLeft: exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
				Assets: s.coin("3apple"), Price: s.coin("15peach"), MarketId: 1, Seller: s.addr5.String(),
				SellerSettlementFlatFee: s.coinP("6fig"),
				ExternalId:              "the-ask-order",
				AllowPartial:            true,
			}),
			expHoldCalls: HoldCalls{
				ReleaseHold: []*ReleaseHoldArgs{
					{addr: s.addr3, funds: s.coins("40peach")},
					{addr: s.addr5, funds: s.coins("14fig,7apple")},
				},
			},
			expBankCalls: BankCalls{
				SendCoins: []*SendCoinsArgs{
					{ctxHasQuarantineBypass: true, fromAddr: s.addr5, toAddr: s.addr3, amt: s.coins("7apple")},
					{ctxHasQuarantineBypass: true, fromAddr: s.addr3, toAddr: s.addr5, amt: s.coins("40peach")},
					{fromAddr: s.addr5, toAddr: s.marketAddr1, amt: s.coins("14fig")},
				},
			},
		},
		{
			name: "one ask one bid: partial bid",
			setup: func() {
				s.k.SetParams(s.ctx, &exchange.Params{})
				s.requireCreateMarket(exchange.Market{MarketId: 1})
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
					Assets: s.coin("7apple"), Price: s.coin("30peach"), MarketId: 1, Seller: s.addr5.String(),
					ExternalId: "the-ask-order",
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(2).WithBid(&exchange.BidOrder{
					Assets: s.coin("10apple"), Price: s.coin("50peach"), MarketId: 1, Buyer: s.addr3.String(),
					BuyerSettlementFees: s.coins("20fig"),
					ExternalId:          "the-bid-order",
					AllowPartial:        true,
				}))
			},
			marketID:      1,
			askOrderIDs:   []uint64{1},
			bidOrderIDs:   []uint64{2},
			expectPartial: true,
			expEvents: []proto.Message{
				&exchange.EventOrderFilled{
					OrderId: 1, Assets: "7apple", Price: "35peach",
					MarketId: 1, ExternalId: "the-ask-order",
				},
				&exchange.EventOrderPartiallyFilled{
					OrderId: 2, Assets: "7apple", Price: "35peach", Fees: "14fig",
					MarketId: 1, ExternalId: "the-bid-order",
				},
			},
			expPartialLeft: exchange.NewOrder(2).WithBid(&exchange.BidOrder{
				Assets: s.coin("3apple"), Price: s.coin("15peach"), MarketId: 1, Buyer: s.addr3.String(),
				BuyerSettlementFees: s.coins("6fig"),
				ExternalId:          "the-bid-order",
				AllowPartial:        true,
			}),
			expHoldCalls: HoldCalls{
				ReleaseHold: []*ReleaseHoldArgs{
					{addr: s.addr5, funds: s.coins("7apple")},
					{addr: s.addr3, funds: s.coins("14fig,35peach")},
				},
			},
			expBankCalls: BankCalls{
				SendCoins: []*SendCoinsArgs{
					{ctxHasQuarantineBypass: true, fromAddr: s.addr5, toAddr: s.addr3, amt: s.coins("7apple")},
					{ctxHasQuarantineBypass: true, fromAddr: s.addr3, toAddr: s.addr5, amt: s.coins("35peach")},
					{fromAddr: s.addr3, toAddr: s.marketAddr1, amt: s.coins("14fig")},
				},
			},
		},
		{
			name: "two asks, three bids",
			setup: func() {
				s.k.SetParams(s.ctx, &exchange.Params{})
				s.requireCreateMarket(exchange.Market{MarketId: 2})
				store := s.getStore()
				s.requireSetOrderInStore(store, exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
					Assets: s.coin("25apple"), Price: s.coin("100peach"), MarketId: 2, Seller: s.addr1.String(),
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(6).WithBid(&exchange.BidOrder{
					Assets: s.coin("20apple"), Price: s.coin("40peach"), MarketId: 2, Buyer: s.addr2.String(),
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(7).WithBid(&exchange.BidOrder{
					Assets: s.coin("30apple"), Price: s.coin("60peach"), MarketId: 2, Buyer: s.addr3.String(),
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(77).WithAsk(&exchange.AskOrder{
					Assets: s.coin("75apple"), Price: s.coin("50peach"), MarketId: 2, Seller: s.addr4.String(),
				}))
				s.requireSetOrderInStore(store, exchange.NewOrder(88).WithBid(&exchange.BidOrder{
					Assets: s.coin("50apple"), Price: s.coin("50peach"), MarketId: 2, Buyer: s.addr5.String(),
				}))
			},
			marketID:      2,
			askOrderIDs:   []uint64{77, 1},
			bidOrderIDs:   []uint64{7, 6, 88},
			expectPartial: false,
			expEvents: []proto.Message{
				&exchange.EventOrderFilled{OrderId: 77, Assets: "75apple", Price: "50peach", MarketId: 2},
				&exchange.EventOrderFilled{OrderId: 1, Assets: "25apple", Price: "100peach", MarketId: 2},
				&exchange.EventOrderFilled{OrderId: 7, Assets: "30apple", Price: "60peach", MarketId: 2},
				&exchange.EventOrderFilled{OrderId: 6, Assets: "20apple", Price: "40peach", MarketId: 2},
				&exchange.EventOrderFilled{OrderId: 88, Assets: "50apple", Price: "50peach", MarketId: 2},
			},
			expHoldCalls: HoldCalls{
				ReleaseHold: []*ReleaseHoldArgs{
					{addr: s.addr4, funds: s.coins("75apple")},
					{addr: s.addr1, funds: s.coins("25apple")},
					{addr: s.addr3, funds: s.coins("60peach")},
					{addr: s.addr2, funds: s.coins("40peach")},
					{addr: s.addr5, funds: s.coins("50peach")},
				},
			},
			expBankCalls: BankCalls{
				SendCoins: []*SendCoinsArgs{
					{ctxHasQuarantineBypass: true, fromAddr: s.addr1, toAddr: s.addr5, amt: s.coins("25apple")},
					{ctxHasQuarantineBypass: true, fromAddr: s.addr2, toAddr: s.addr1, amt: s.coins("40peach")},
					{ctxHasQuarantineBypass: true, fromAddr: s.addr5, toAddr: s.addr1, amt: s.coins("50peach")},
				},
				InputOutputCoins: []*InputOutputCoinsArgs{
					{
						ctxHasQuarantineBypass: true,
						inputs: []banktypes.Input{
							{Address: s.addr4.String(), Coins: s.coins("75apple")},
						},
						outputs: []banktypes.Output{
							{Address: s.addr3.String(), Coins: s.coins("30apple")},
							{Address: s.addr2.String(), Coins: s.coins("20apple")},
							{Address: s.addr5.String(), Coins: s.coins("25apple")},
						},
					},
					{
						ctxHasQuarantineBypass: true,
						inputs: []banktypes.Input{
							{Address: s.addr3.String(), Coins: s.coins("60peach")},
						},
						outputs: []banktypes.Output{
							{Address: s.addr4.String(), Coins: s.coins("50peach")},
							{Address: s.addr1.String(), Coins: s.coins("10peach")},
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			if s.accKeeper == nil {
				s.accKeeper = NewMockAccountKeeper()
			}
			if tc.bankKeeper == nil {
				tc.bankKeeper = NewMockBankKeeper()
			}
			if tc.holdKeeper == nil {
				tc.holdKeeper = NewMockHoldKeeper()
			}

			expEvents := make(sdk.Events, len(tc.expEvents))
			for i, expEvent := range tc.expEvents {
				event, err := sdk.TypedEventToEvent(expEvent)
				s.Require().NoError(err, "[%d]TypedEventToEvent(%T)", expEvent)
				expEvents[i] = event
			}

			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)
			kpr := s.k.WithAccountKeeper(s.accKeeper).WithBankKeeper(tc.bankKeeper).WithHoldKeeper(tc.holdKeeper)
			var err error
			testFunc := func() {
				err = kpr.SettleOrders(ctx, tc.marketID, tc.askOrderIDs, tc.bidOrderIDs, tc.expectPartial)
			}
			s.Require().NotPanics(testFunc, "SettleOrders")
			s.assertErrorValue(err, tc.expErr, "SettleOrders error")
			actEvents := em.Events()
			s.assertEqualEvents(expEvents, actEvents, "SettleOrders events")
			s.assertHoldKeeperCalls(tc.holdKeeper, tc.expHoldCalls, "SettleOrders")
			s.assertBankKeeperCalls(tc.bankKeeper, tc.expBankCalls, "SettleOrders")

			if len(actEvents) == 0 {
				return
			}

			for _, orderID := range tc.askOrderIDs {
				if tc.expPartialLeft == nil || tc.expPartialLeft.OrderId != orderID {
					order, oerr := s.k.GetOrder(s.ctx, orderID)
					s.Assert().NoError(oerr, "GetOrder(%d) (ask) after SettleOrders", orderID)
					s.Assert().Nil(order, "GetOrder(%d) (ask) after SettleOrders", orderID)
				}
			}
			for _, orderID := range tc.bidOrderIDs {
				if tc.expPartialLeft == nil || tc.expPartialLeft.OrderId != orderID {
					order, oerr := s.k.GetOrder(s.ctx, orderID)
					s.Assert().NoError(oerr, "GetOrder(%d) (bid) after SettleOrders", orderID)
					s.Assert().Nil(order, "GetOrder(%d) (bid) after SettleOrders", orderID)
				}
			}
			if tc.expPartialLeft != nil {
				order, oerr := s.k.GetOrder(s.ctx, tc.expPartialLeft.OrderId)
				s.Assert().NoError(oerr, "GetOrder(%d) (partial) after SettleOrders", tc.expPartialLeft.OrderId)
				s.Assert().Equal(tc.expPartialLeft, order, "GetOrder(%d) (partial) after SettleOrders", tc.expPartialLeft.OrderId)
			}
		})
	}
}
