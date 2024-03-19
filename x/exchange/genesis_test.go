package exchange

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil/assertions"
)

func TestDefaultGenesisState(t *testing.T) {
	pioconfig.SetProvenanceConfig("", 0)
	expected := &GenesisState{
		Params: &Params{
			DefaultSplit:         DefaultDefaultSplit,
			DenomSplits:          nil,
			FeeCreatePaymentFlat: []sdk.Coin{{Denom: "nhash", Amount: sdkmath.NewInt(DefaultFeeCreatePaymentFlatAmount)}},
			FeeAcceptPaymentFlat: []sdk.Coin{{Denom: "nhash", Amount: sdkmath.NewInt(DefaultFeeAcceptPaymentFlatAmount)}},
		},
		Markets:      nil,
		Orders:       nil,
		LastMarketId: 0,
		Commitments:  nil,
		Payments:     nil,
	}
	var actual *GenesisState
	testFunc := func() {
		actual = DefaultGenesisState()
	}
	require.NotPanics(t, testFunc, "DefaultGenesisState()")
	assert.Equal(t, expected, actual, "DefaultGenesisState() result")
}

func TestGenesisState_Validate(t *testing.T) {
	pioconfig.SetProvenanceConfig("", 0)
	addr1 := sdk.AccAddress("addr1_______________").String()
	addr2 := sdk.AccAddress("addr2_______________").String()
	addr3 := sdk.AccAddress("addr3_______________").String()
	addr4 := sdk.AccAddress("addr4_______________").String()
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	askOrder := func(orderID uint64, marketID uint32, assets string, price string) Order {
		priceCoin, err := sdk.ParseCoinNormalized(price)
		require.NoError(t, err, "ask price sdk.ParseCoinNormalized(%q)", price)
		assetsCoin, err := sdk.ParseCoinNormalized(assets)
		require.NoError(t, err, "ask assets sdk.ParseCoinNormalized(%q)", assets)
		return *NewOrder(orderID).WithAsk(&AskOrder{
			MarketId: marketID,
			Seller:   addr1,
			Assets:   assetsCoin,
			Price:    priceCoin,
		})
	}
	bidOrder := func(orderID uint64, marketID uint32, assets string, price string) Order {
		priceCoin, err := sdk.ParseCoinNormalized(price)
		require.NoError(t, err, "bid price sdk.ParseCoinNormalized(%q)", price)
		assetsCoin, err := sdk.ParseCoinNormalized(assets)
		require.NoError(t, err, "bid assets sdk.ParseCoinNormalized(%q)", assets)
		return *NewOrder(orderID).WithBid(&BidOrder{
			MarketId: marketID,
			Buyer:    addr1,
			Assets:   assetsCoin,
			Price:    priceCoin,
		})
	}
	payment := func(source, sourceAmount, target, targetAmount, externalID string) Payment {
		rv := Payment{
			Source:     source,
			Target:     target,
			ExternalId: externalID,
		}
		var err error
		if len(sourceAmount) > 0 {
			rv.SourceAmount, err = sdk.ParseCoinsNormalized(sourceAmount)
			require.NoError(t, err, "source ParseCoinsNormalized(%q)", sourceAmount)
		}
		if len(targetAmount) > 0 {
			rv.TargetAmount, err = sdk.ParseCoinsNormalized(targetAmount)
			require.NoError(t, err, "target ParseCoinsNormalized(%q)", targetAmount)
		}
		return rv
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
				LastOrderId: 9,
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
				`invalid order[0]: invalid order id: cannot be zero`,
				`invalid order[1]: invalid order id: cannot be zero`,
				`invalid order[2]: invalid order id: cannot be zero`,
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
				LastOrderId: 3,
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
				LastOrderId: 1,
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
				LastOrderId: 1,
				Commitments: []Commitment{
					{Account: addr1, MarketId: 1, Amount: sdk.Coins{coin(15, "apple")}},
					{Account: addr1, MarketId: 4, Amount: sdk.Coins{coin(45, "apple")}},
					{Account: addr1, MarketId: 3, Amount: sdk.Coins{coin(35, "apple")}},
				},
				Payments: []Payment{
					payment(addr1, "3strawberry", addr2, "5tangerine", ""),
					payment(addr1, "5strawberry", addr2, "6tangerine", "v2"),
					payment(addr1, "20strawberry", addr3, "4tangerine", ""),
				},
			},
			expErr: []string{
				"invalid params: default split 10001 cannot be greater than 10000",
				`invalid market[1]: invalid create-bid flat fee option "-1zapp": negative coin amount: -1`,
				`invalid order[1]: unknown market id 4`,
				"last order id 1 is less than the largest id in the provided orders 3",
				"invalid commitment[1]: unknown market id 4",
				"invalid payment[2]: duplicate payment, source " + addr1 + " and external id \"\" seen at [0]",
			},
		},
		{
			name:     "last market id 1",
			genState: GenesisState{LastMarketId: 1},
			expErr:   nil,
		},
		{
			name: "last market id less than largest market id",
			genState: GenesisState{
				Markets:      []Market{{MarketId: 3}, {MarketId: 1}},
				LastMarketId: 1,
			},
			expErr: nil,
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
		{
			name: "last order id less than largest order id",
			genState: GenesisState{
				Markets: []Market{{MarketId: 1}},
				Orders: []Order{
					askOrder(1, 1, "28fry", "2bender"),
					bidOrder(88, 1, "28fry", "2bender"),
					bidOrder(2, 1, "28fry", "2bender"),
					askOrder(3, 1, "28fry", "2bender"),
				},
				LastOrderId: 87,
			},
			expErr: []string{"last order id 87 is less than the largest id in the provided orders 88"},
		},
		{
			name: "last order id equals largest order id",
			genState: GenesisState{
				Markets: []Market{{MarketId: 1}},
				Orders: []Order{
					askOrder(1, 1, "28fry", "2bender"),
					bidOrder(88, 1, "28fry", "2bender"),
					bidOrder(2, 1, "28fry", "2bender"),
					askOrder(3, 1, "28fry", "2bender"),
				},
				LastOrderId: 88,
			},
			expErr: nil,
		},
		{
			name: "last order id more than largest order id",
			genState: GenesisState{
				Markets: []Market{{MarketId: 1}},
				Orders: []Order{
					askOrder(1, 1, "28fry", "2bender"),
					bidOrder(88, 1, "28fry", "2bender"),
					bidOrder(2, 1, "28fry", "2bender"),
					askOrder(3, 1, "28fry", "2bender"),
				},
				LastOrderId: 89,
			},
			expErr: nil,
		},
		{
			name: "one commitment: bad account",
			genState: GenesisState{
				Markets: []Market{{MarketId: 1}},
				Commitments: []Commitment{
					{Account: "notanaccountstring", MarketId: 1, Amount: sdk.Coins{coin(58, "cherry")}},
				},
			},
			expErr: []string{`invalid commitment[0]: invalid account "notanaccountstring": decoding bech32 failed: invalid separator index -1`},
		},
		{
			name: "one commitment: market zero",
			genState: GenesisState{
				Markets: []Market{{MarketId: 1}},
				Commitments: []Commitment{
					{Account: addr1, MarketId: 0, Amount: sdk.Coins{coin(58, "cherry")}},
				},
			},
			expErr: []string{`invalid commitment[0]: invalid market id: cannot be zero`},
		},
		{
			name: "one commitment: unknown market",
			genState: GenesisState{
				Markets: []Market{{MarketId: 1}, {MarketId: 3}},
				Commitments: []Commitment{
					{Account: addr1, MarketId: 2, Amount: sdk.Coins{coin(58, "cherry")}},
				},
			},
			expErr: []string{`invalid commitment[0]: unknown market id 2`},
		},
		{
			name: "one commitment: bad amount denom",
			genState: GenesisState{
				Markets: []Market{{MarketId: 1}},
				Commitments: []Commitment{
					{Account: addr1, MarketId: 1, Amount: sdk.Coins{coin(58, "c")}},
				},
			},
			expErr: []string{`invalid commitment[0]: invalid amount "58c": invalid denom: c`},
		},
		{
			name: "one commitment: negative amount",
			genState: GenesisState{
				Markets: []Market{{MarketId: 1}},
				Commitments: []Commitment{
					{Account: addr1, MarketId: 1, Amount: sdk.Coins{coin(-1, "cherry")}},
				},
			},
			expErr: []string{`invalid commitment[0]: invalid amount "-1cherry": coin -1cherry amount is not positive`},
		},
		{
			name: "four commitments: three invalid",
			genState: GenesisState{
				Markets: []Market{{MarketId: 1}, {MarketId: 8}},
				Commitments: []Commitment{
					{Account: addr1, MarketId: 4, Amount: sdk.Coins{coin(58, "cherry")}},
					{Account: "whatanaddr", MarketId: 8, Amount: sdk.Coins{coin(12, "grape")}},
					{Account: addr1, MarketId: 8, Amount: sdk.Coins{coin(19, "apple")}},
					{Account: addr1, MarketId: 1, Amount: sdk.Coins{coin(-6, "banana")}},
				},
			},
			expErr: []string{
				`invalid commitment[0]: unknown market id 4`,
				`invalid commitment[1]: invalid account "whatanaddr": decoding bech32 failed: invalid separator index -1`,
				`invalid commitment[3]: invalid amount "-6banana": coin -6banana amount is not positive`,
			},
		},
		{
			name:     "one payment: okay",
			genState: GenesisState{Payments: []Payment{payment(addr4, "8starfruit", addr3, "2tomato", "p-zero")}},
			expErr:   nil,
		},
		{
			name:     "one payment: no source",
			genState: GenesisState{Payments: []Payment{payment("", "66strawberry", addr3, "888tomato", "p-zero")}},
			expErr:   []string{"invalid payment[0]: invalid source \"\": empty address string is not allowed"},
		},
		{
			name:     "one payment: no amounts",
			genState: GenesisState{Payments: []Payment{payment(addr2, "", addr1, "", "f-zero")}},
			expErr:   []string{"invalid payment[0]: source amount and target amount cannot both be zero"},
		},
		{
			name: "three payments: all invalid",
			genState: GenesisState{
				Payments: []Payment{
					payment("", "12strawberry", addr1, "", ""),
					payment(addr3, "", addr4, "", "there's two of me"),
					payment(addr3, "", addr2, "15tomato", "there's two of me"),
				},
			},
			expErr: []string{
				"invalid payment[0]: invalid source \"\": empty address string is not allowed",
				"invalid payment[1]: source amount and target amount cannot both be zero",
				"invalid payment[2]: duplicate payment, source " + addr3 + " and external id \"there's two of me\" seen at [1]",
			},
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
