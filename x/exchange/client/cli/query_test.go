package cli_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *CmdTestSuite) TestCmdQueryOrderFeeCalc() {
	tests := []queryCmdTestCase{
		{
			name:     "input error",
			args:     []string{"order-fee-calc", "--bid", "--market", "3", "--assets", "99apple"},
			expInErr: []string{"required flag(s) \"price\" not set"},
		},
		{
			name:     "market does not exist",
			args:     []string{"fee-calc", "--market", "69", "--bid", "--price", "1000peach"},
			expInErr: []string{"market 69 does not exist", "invalid request", "InvalidArgument"},
		},
		{
			name:     "only ask fees, ask",
			args:     []string{"order-calc", "--market", "3", "--ask", "--price", "1000peach"},
			expInOut: nil,
			expOut: `creation_fee_options:
- amount: "10"
  denom: peach
settlement_flat_fee_options:
- amount: "50"
  denom: peach
settlement_ratio_fee_options:
- amount: "10"
  denom: peach
`,
		},
		{
			name:     "only ask fees, bid",
			args:     []string{"order-calc", "--market", "3", "--bid", "--price", "1000peach"},
			expInOut: nil,
			expOut: `creation_fee_options: []
settlement_flat_fee_options: []
settlement_ratio_fee_options: []
`,
		},
		{
			name:     "only bid fees, ask",
			args:     []string{"order-calc", "--market", "5", "--ask", "--price", "1000peach"},
			expInOut: nil,
			expOut: `creation_fee_options: []
settlement_flat_fee_options: []
settlement_ratio_fee_options: []
`,
		},
		{
			name: "only bid fees, bid",
			args: []string{"order-calc", "--market", "5", "--bid", "--price", "1000peach", "--output", "--json"},
			expInOut: []string{
				`"creation_fee_options":[{"denom":"peach","amount":"10"}]`,
				`"settlement_flat_fee_options":[{"denom":"peach","amount":"50"}]`,
				`"settlement_ratio_fee_options":[{"denom":"peach","amount":"10"},{"denom":"stake","amount":"30"}]`,
			},
		},
		{
			name: "both fees, ask",
			args: []string{"order-calc", "--market", "420", "--ask", "--price", "1000peach", "--output", "--json"},
			expInOut: []string{
				`"creation_fee_options":[{"denom":"peach","amount":"20"}]`,
				`"settlement_flat_fee_options":[{"denom":"peach","amount":"100"}]`,
				`"settlement_ratio_fee_options":[{"denom":"peach","amount":"14"}]`,
			},
		},
		{
			name:     "both fees, bid",
			args:     []string{"order-calc", "--market", "420", "--bid", "--price", "1000peach"},
			expInOut: nil,
			expOut: `creation_fee_options:
- amount: "25"
  denom: peach
settlement_flat_fee_options:
- amount: "105"
  denom: peach
settlement_ratio_fee_options:
- amount: "20"
  denom: peach
- amount: "60"
  denom: stake
`,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runQueryCmdTestCase(tc)
		})
	}
}

func (s *CmdTestSuite) TestCmdQueryGetOrder() {
	tests := []queryCmdTestCase{
		{
			name:     "no order id",
			args:     []string{"get-order"},
			expInErr: []string{"no <order id> provided"},
		},
		{
			name:     "order does not exist",
			args:     []string{"order", "1234567899"},
			expInErr: []string{"order 1234567899 not found", "invalid request", "InvalidArgument"},
		},
		{
			name: "ask order",
			args: []string{"order", "--order", "42"},
			expOut: fmt.Sprintf(`order:
  ask_order:
    allow_partial: true
    assets:
      amount: "4200"
      denom: acorn
    external_id: my-id-42
    market_id: 420
    price:
      amount: "17640"
      denom: peach
    seller: %s
    seller_settlement_flat_fee: null
  order_id: "42"
`,
				s.accountAddrs[2],
			),
		},
		{
			name: "bid order",
			args: []string{"get-order", "41", "--output", "json"},
			expInOut: []string{
				`"order_id":"41"`,
				`"bid_order":`,
				`"market_id":420`,
				fmt.Sprintf(`"buyer":"%s"`, s.accountAddrs[1]),
				`"assets":{"denom":"apple","amount":"4100"}`,
				`"price":{"denom":"peach","amount":"16810"}`,
				`"buyer_settlement_fees":[]`,
				`"allow_partial":false`,
				`"external_id":"my-id-41`,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runQueryCmdTestCase(tc)
		})
	}
}

func (s *CmdTestSuite) TestCmdQueryGetOrderByExternalID() {
	tests := []queryCmdTestCase{
		{
			name:     "cmd error",
			args:     []string{"order-by-external-id", "--external-id", "my-id-15"},
			expInErr: []string{"required flag(s) \"market\" not set"},
		},
		{
			name: "order does not exist",
			args: []string{"get-order-by-external-id", "--market", "3", "--external-id", "my-id-15"},
			expInErr: []string{
				"order not found in market 3 with external id \"my-id-15\"", "invalid request", "InvalidArgument",
			},
		},
		{
			name: "order exists",
			args: []string{"external-id", "--external-id", "my-id-15", "--market", "420", "--output", "json"},
			expInOut: []string{
				`"order_id":"15"`,
				`"bid_order":`,
				`"market_id":420`,
				fmt.Sprintf(`"buyer":"%s"`, s.accountAddrs[5]),
				`"assets":{"denom":"acorn","amount":"1500"}`,
				`"price":{"denom":"peach","amount":"2250"}`,
				`"buyer_settlement_fees":[]`,
				`"allow_partial":false`,
				`"external_id":"my-id-15"`,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runQueryCmdTestCase(tc)
		})
	}
}

func (s *CmdTestSuite) TestCmdQueryGetMarketOrders() {
	tests := []queryCmdTestCase{
		{
			name:     "cmd error",
			args:     []string{"market-orders", "420", "--asks", "--bids"},
			expInErr: []string{"if any flags in the group [asks bids] are set none of the others can be; [asks bids] were all set"},
		},
		{
			name: "no orders",
			args: []string{"get-market-orders", "420", "--order", "1234567899"},
			expOut: `orders: []
pagination:
  next_key: null
  total: "0"
`,
		},
		{
			name: "several orders",
			args: []string{"market-orders", "--asks", "--market", "420", "--order", "30", "--output", "json", "--count-total"},
			expInOut: []string{
				`"market_id":420`,
				`"order_id":"31"`, `"order_id":"34"`, `"order_id":"36"`,
				`"order_id":"37"`, `"order_id":"40"`, `"order_id":"42"`,
				`"order_id":"43"`, `"order_id":"46"`, `"order_id":"48"`,
				`"order_id":"49"`, `"order_id":"52"`, `"order_id":"54"`,
				`"order_id":"55"`, `"order_id":"58"`, `"order_id":"60"`,
				`"total":"15"`,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runQueryCmdTestCase(tc)
		})
	}
}

func (s *CmdTestSuite) TestCmdQueryGetOwnerOrders() {
	tests := []queryCmdTestCase{
		{
			name:     "cmd error",
			args:     []string{"owner-orders", "--limit", "10"},
			expInErr: []string{"no <owner> provided"},
		},
		{
			name:   "no orders",
			args:   []string{"get-owner-orders", sdk.AccAddress("not_gonna_have_it___").String(), "--output", "json"},
			expOut: `{"orders":[],"pagination":{"next_key":null,"total":"0"}}` + "\n",
		},
		{
			name: "several orders",
			args: []string{"owner-orders", "--owner", s.accountAddrs[9].String()},
			expInOut: []string{
				`market_id: 420`,
				`order_id: "9"`, `order_id: "19"`, `order_id: "29"`,
				`order_id: "39"`, `order_id: "49"`, `order_id: "59"`,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runQueryCmdTestCase(tc)
		})
	}
}

func (s *CmdTestSuite) TestCmdQueryGetAssetOrders() {
	tests := []queryCmdTestCase{
		{
			name:     "cmd error",
			args:     []string{"asset-orders", "--asks"},
			expInErr: []string{"no <asset> provided"},
		},
		{
			name:     "no orders",
			args:     []string{"asset-orders", "peach"},
			expInOut: nil,
			expOut: `orders: []
pagination:
  next_key: null
  total: "0"
`,
		},
		{
			name: "several orders",
			args: []string{"asset-orders", "--denom", "apple", "--limit", "5", "--order", "10", "--output", "json"},
			expInOut: []string{
				`"market_id":420`, `"order_id":"11"`, `"order_id":"12"`,
				`"order_id":"13"`, `"order_id":"17"`, `"order_id":"18"`,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runQueryCmdTestCase(tc)
		})
	}
}

func (s *CmdTestSuite) TestCmdQueryGetAllOrders() {
	tests := []queryCmdTestCase{
		{
			name:     "cmd error",
			args:     []string{"all-orders", "extraarg"},
			expInErr: []string{"unknown command \"extraarg\" for \"exchange all-orders\""},
		},
		{
			name: "no orders",
			// this page key is the base64 encoded max uint64 -1, aka "the next to last possible order id."
			// Hopefully these unit tests get up that far.
			args:   []string{"get-all-orders", "--page-key", "//////////4=", "--output", "json"},
			expOut: `{"orders":[],"pagination":{"next_key":null,"total":"0"}}` + "\n",
		},
		{
			name: "some orders",
			args: []string{"all-orders", "--limit", "3", "--offset", "20"},
			expInOut: []string{
				`order_id: "21"`, `order_id: "22"`, `order_id: "23"`,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runQueryCmdTestCase(tc)
		})
	}
}

// TODO[1701]: func (s *CmdTestSuite) TestCmdQueryGetMarket()

// TODO[1701]: func (s *CmdTestSuite) TestCmdQueryGetAllMarkets()

// TODO[1701]: func (s *CmdTestSuite) TestCmdQueryParams()

// TODO[1701]: func (s *CmdTestSuite) TestCmdQueryValidateCreateMarket()

// TODO[1701]: func (s *CmdTestSuite) TestCmdQueryValidateMarket()

// TODO[1701]: func (s *CmdTestSuite) TestCmdQueryValidateManageFees()
