package exchange

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/testutil/assertions"
)

func TestGenesisState_Validate(t *testing.T) {
	addr1 := sdk.AccAddress("addr1_______________").String()
	coins := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "sdk.ParseCoinsNormalized(%q)", coins)
		return rv
	}
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	askOrder := func(orderID uint64, marketID uint32, assets string, price string) Order {
		priceCoin, err := sdk.ParseCoinNormalized(price)
		require.NoError(t, err, "sdk.ParseCoinNormalized(%q)", price)
		return *NewOrder(orderID).WithAsk(&AskOrder{
			MarketId: marketID,
			Seller:   addr1,
			Assets:   coins(assets),
			Price:    priceCoin,
		})
	}
	bidOrder := func(orderID uint64, marketID uint32, assets string, price string) Order {
		priceCoin, err := sdk.ParseCoinNormalized(price)
		require.NoError(t, err, "sdk.ParseCoinNormalized(%q)", price)
		return *NewOrder(orderID).WithBid(&BidOrder{
			MarketId: marketID,
			Buyer:    addr1,
			Assets:   coins(assets),
			Price:    priceCoin,
		})
	}

	tests := []struct {
		name     string
		genState GenesisState
		expErr   []string
	}{
		{
			name:     "zero state",
			genState: GenesisState{},
			expErr:   nil,
		},
		{
			name:     "default state",
			genState: *DefaultGenesisState(),
			expErr:   nil,
		},
		{
			name:     "empty params",
			genState: GenesisState{Params: &Params{}},
			expErr:   nil,
		},
		{
			name:     "invalid params",
			genState: GenesisState{Params: &Params{DefaultSplit: 10_001}},
			expErr:   []string{"invalid params: default split 10001 cannot be greater than 10000"},
		},
		{
			name: "three markets, 3 orders each all valid",
			genState: GenesisState{
				Params: DefaultParams(),
				Markets: []Market{
					{MarketId: 1},
					{MarketId: 2},
					{MarketId: 3},
				},
				Orders: []Order{
					askOrder(1, 1, "99fry", "9leela"),
					bidOrder(2, 1, "100fry", "10leela"),
					bidOrder(3, 1, "101fry", "11leela"),
					askOrder(4, 2, "12zapp", "1farnsworth"),
					askOrder(5, 2, "14zapp", "1farnsworth"),
					bidOrder(6, 2, "80zapp", "10farnsworth"),
					bidOrder(7, 3, "5wong", "50nibbler"),
					bidOrder(8, 3, "5wong", "50nibbler"),
					bidOrder(9, 3, "5wong", "50nibbler"),
				},
			},
			expErr: nil,
		},
		{
			name: "three markets: all invalid",
			genState: GenesisState{
				Markets: []Market{
					{MarketId: 1, FeeCreateAskFlat: sdk.Coins{coin(-1, "kif")}},
					{MarketId: 2, FeeCreateAskFlat: sdk.Coins{coin(-2, "kif")}},
					{MarketId: 3, FeeCreateAskFlat: sdk.Coins{coin(-3, "kif")}},
				},
			},
			expErr: []string{
				`invalid market[0]: invalid create-ask flat fee option "-1kif": negative coin amount: -1`,
				`invalid market[1]: invalid create-ask flat fee option "-2kif": negative coin amount: -2`,
				`invalid market[2]: invalid create-ask flat fee option "-3kif": negative coin amount: -3`,
			},
		},
		{
			name: "three markets: all market id zero",
			genState: GenesisState{
				Markets: []Market{
					{MarketId: 0},
					{MarketId: 0},
					{MarketId: 0},
				},
			},
			expErr: []string{
				`invalid market[0]: market id cannot be zero`,
				`invalid market[1]: market id cannot be zero`,
				`invalid market[2]: market id cannot be zero`,
			},
		},
		{
			name: "three markets: all market id one",
			genState: GenesisState{
				Markets: []Market{
					{MarketId: 1},
					{MarketId: 1},
					{MarketId: 1},
				},
			},
			expErr: []string{
				`invalid market[1]: duplicate market id 1 seen at [0]`,
				`invalid market[2]: duplicate market id 1 seen at [0]`,
			},
		},
		{
			name: "three orders: all invalid",
			genState: GenesisState{
				Markets: []Market{{MarketId: 4}, {MarketId: 5}, {MarketId: 6}},
				Orders: []Order{
					askOrder(0, 4, "28fry", "2bender"),
					bidOrder(0, 5, "28fry", "2bender"),
					askOrder(0, 6, "28fry", "2bender"),
				},
			},
			expErr: []string{
				`invalid order[0]: invalid order id: must not be zero`,
				`invalid order[1]: invalid order id: must not be zero`,
				`invalid order[2]: invalid order id: must not be zero`,
			},
		},
		{
			name: "three orders: all unknown markets",
			genState: GenesisState{
				Markets: []Market{{MarketId: 4}, {MarketId: 5}, {MarketId: 6}},
				Orders: []Order{
					askOrder(1, 1, "28fry", "2bender"),
					bidOrder(2, 2, "28fry", "2bender"),
					askOrder(3, 3, "28fry", "2bender"),
				},
			},
			expErr: []string{
				`invalid order[0]: unknown market id 1`,
				`invalid order[1]: unknown market id 2`,
				`invalid order[2]: unknown market id 3`,
			},
		},
		{
			name: "three orders: all id one",
			genState: GenesisState{
				Markets: []Market{{MarketId: 4}, {MarketId: 5}, {MarketId: 6}},
				Orders: []Order{
					askOrder(1, 4, "28fry", "2bender"),
					bidOrder(1, 5, "28fry", "2bender"),
					askOrder(1, 6, "28fry", "2bender"),
				},
			},
			expErr: []string{
				`invalid order[1]: duplicate order id 1 seen at [0]`,
				`invalid order[2]: duplicate order id 1 seen at [0]`,
			},
		},
		{
			name: "multiple errors",
			genState: GenesisState{
				Params: &Params{DefaultSplit: 10_001},
				Markets: []Market{
					{MarketId: 1},
					{MarketId: 2, FeeCreateBidFlat: sdk.Coins{coin(-1, "zapp")}},
					{MarketId: 3},
				},
				Orders: []Order{
					askOrder(1, 1, "28fry", "2bender"),
					bidOrder(2, 4, "28fry", "2bender"),
					askOrder(3, 3, "28fry", "2bender"),
				},
			},
			expErr: []string{
				"invalid params: default split 10001 cannot be greater than 10000",
				`invalid market[1]: invalid create-bid flat fee option "-1zapp": negative coin amount: -1`,
				`invalid order[1]: unknown market id 4`,
			},
		},
		{
			name:     "last market id 1",
			genState: GenesisState{LastMarketId: 1},
			expErr:   nil,
		},
		{
			name:     "last market id 256",
			genState: GenesisState{LastMarketId: 256},
			expErr:   nil,
		},
		{
			name:     "last market id 65,536",
			genState: GenesisState{LastMarketId: 65_536},
			expErr:   nil,
		},
		{
			name:     "last market id 16,777,216",
			genState: GenesisState{LastMarketId: 16_777_216},
			expErr:   nil,
		},
		{
			name:     "last market id max uint32",
			genState: GenesisState{LastMarketId: 4_294_967_295},
			expErr:   nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.genState.Validate()
			}
			require.NotPanics(t, testFunc, "GenesisState.Validate")

			assertions.AssertErrorContents(t, err, tc.expErr, "GenesisState.Validate")
		})
	}
}
