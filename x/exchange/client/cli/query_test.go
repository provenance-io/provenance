package cli_test

import "fmt"

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
      denom: apple
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

// TODO[1701]: func (s *CmdTestSuite) TestCmdQueryGetOrderByExternalID()

// TODO[1701]: func (s *CmdTestSuite) TestCmdQueryGetMarketOrders()

// TODO[1701]: func (s *CmdTestSuite) TestCmdQueryGetOwnerOrders()

// TODO[1701]: func (s *CmdTestSuite) TestCmdQueryGetAssetOrders()

// TODO[1701]: func (s *CmdTestSuite) TestCmdQueryGetAllOrders()

// TODO[1701]: func (s *CmdTestSuite) TestCmdQueryGetMarket()

// TODO[1701]: func (s *CmdTestSuite) TestCmdQueryGetAllMarkets()

// TODO[1701]: func (s *CmdTestSuite) TestCmdQueryParams()

// TODO[1701]: func (s *CmdTestSuite) TestCmdQueryValidateCreateMarket()

// TODO[1701]: func (s *CmdTestSuite) TestCmdQueryValidateMarket()

// TODO[1701]: func (s *CmdTestSuite) TestCmdQueryValidateManageFees()
