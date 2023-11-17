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
			expOut: `order:
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
    seller: ` + s.accountAddrs[2].String() + `
    seller_settlement_flat_fee: null
  order_id: "42"
`,
		},
		{
			name: "bid order",
			args: []string{"get-order", "41", "--output", "json"},
			expInOut: []string{
				`"order_id":"41"`,
				`"bid_order":`,
				`"market_id":420,`,
				fmt.Sprintf(`"buyer":"%s"`, s.accountAddrs[1]),
				`"assets":{"denom":"apple","amount":"4100"}`,
				`"price":{"denom":"peach","amount":"16810"}`,
				`"buyer_settlement_fees":[]`,
				`"allow_partial":false`,
				`"external_id":"my-id-41"`,
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
				`"market_id":420,`,
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
				`"market_id":420,`,
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
				`"market_id":420,`, `"order_id":"11"`, `"order_id":"12"`,
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

func (s *CmdTestSuite) TestCmdQueryGetMarket() {
	tests := []queryCmdTestCase{
		{
			name:     "no market id",
			args:     []string{"market"},
			expInErr: []string{"no <market id> provided"},
		},
		{
			name:     "market does not exist",
			args:     []string{"get-market", "419"},
			expInErr: []string{"market 419 not found", "invalid request", "InvalidArgument"},
		},
		{
			name: "market exists",
			args: []string{"market", "420"},
			expOut: `address: cosmos1dmk5hcws5xfue8rd6pl5lu6uh8jyt9fpqs0kf6
market:
  accepting_orders: true
  access_grants:
  - address: ` + s.addr1.String() + `
    permissions:
    - PERMISSION_SETTLE
    - PERMISSION_SET_IDS
    - PERMISSION_CANCEL
    - PERMISSION_WITHDRAW
    - PERMISSION_UPDATE
    - PERMISSION_PERMISSIONS
    - PERMISSION_ATTRIBUTES
  allow_user_settlement: true
  fee_buyer_settlement_flat:
  - amount: "105"
    denom: peach
  fee_buyer_settlement_ratios:
  - fee:
      amount: "1"
      denom: peach
    price:
      amount: "50"
      denom: peach
  - fee:
      amount: "3"
      denom: stake
    price:
      amount: "50"
      denom: peach
  fee_create_ask_flat:
  - amount: "20"
    denom: peach
  fee_create_bid_flat:
  - amount: "25"
    denom: peach
  fee_seller_settlement_flat:
  - amount: "100"
    denom: peach
  fee_seller_settlement_ratios:
  - fee:
      amount: "1"
      denom: peach
    price:
      amount: "75"
      denom: peach
  market_details:
    description: It's coming; you know it. It has all the fees.
    icon_uri: ""
    name: THE Market
    website_url: ""
  market_id: 420
  req_attr_create_ask:
  - seller.kyc
  req_attr_create_bid:
  - buyer.kyc
`,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runQueryCmdTestCase(tc)
		})
	}
}

func (s *CmdTestSuite) TestCmdQueryGetAllMarkets() {
	tests := []queryCmdTestCase{
		{
			name:     "cmd error",
			args:     []string{"get-all-markets", "--unexpectedflag"},
			expInErr: []string{"unknown flag: --unexpectedflag"},
		},
		{
			name:     "get all",
			args:     []string{"all-markets"},
			expInOut: []string{`market_id: 3`, `market_id: 5`, `market_id: 420`, `market_id: 421`},
		},
		{
			name: "no markets",
			// this page key is the base64 encoded max uint32 -1, aka "the next to last possible market id."
			// Hopefully these unit tests get up that far.
			args:   []string{"get-all-markets", "--page-key", "/////g==", "--output", "json"},
			expOut: `{"markets":[],"pagination":{"next_key":null,"total":"0"}}` + "\n",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runQueryCmdTestCase(tc)
		})
	}
}

func (s *CmdTestSuite) TestCmdQueryParams() {
	tests := []queryCmdTestCase{
		{
			name:     "cmd error",
			args:     []string{"params", "--unexpectedflag"},
			expInErr: []string{"unknown flag: --unexpectedflag"},
		},
		{
			name: "as text",
			args: []string{"params", "--output", "text"},
			expOut: `params:
  default_split: 500
  denom_splits: []
`,
		},
		{
			name:   "as json",
			args:   []string{"get-params", "--output", "json"},
			expOut: `{"params":{"default_split":500,"denom_splits":[]}}` + "\n",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runQueryCmdTestCase(tc)
		})
	}
}

func (s *CmdTestSuite) TestCmdQueryValidateCreateMarket() {
	tests := []queryCmdTestCase{
		{
			name:     "cmd error",
			args:     []string{"validate-create-market", "--create-ask", "orange"},
			expInErr: []string{"invalid coin expression: \"orange\""},
		},
		{
			name:   "problem with proposal",
			args:   []string{"create-market-validate", "--market", "420", "--name", "Other Name", "--output", "json"},
			expOut: `{"error":"market id 420 account cosmos1dmk5hcws5xfue8rd6pl5lu6uh8jyt9fpqs0kf6 already exists","gov_prop_will_pass":false}` + "\n",
		},
		{
			name: "okay",
			args: []string{"validate-create-market",
				"--name", "New Market", "--create-ask", "50nhash", "--create-bid", "50nhash",
				"--accepting-orders",
			},
			expOut: `error: ""
gov_prop_will_pass: true
`,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runQueryCmdTestCase(tc)
		})
	}
}

func (s *CmdTestSuite) TestCmdQueryValidateMarket() {
	tests := []queryCmdTestCase{
		{
			name:     "no market id",
			args:     []string{"validate-market"},
			expInErr: []string{"no <market id> provided"},
		},
		{
			name:   "invalid market",
			args:   []string{"validate-market", "--output", "json", "--market", "421"},
			expOut: `{"error":"buyer settlement fee ratios have price denom \"plum\" but there is not a seller settlement fee ratio with that price denom"}` + "\n",
		},
		{
			name: "valid market",
			args: []string{"market-validate", "420"},
			expOut: `error: ""
`,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runQueryCmdTestCase(tc)
		})
	}
}

func (s *CmdTestSuite) TestCmdQueryValidateManageFees() {
	tests := []queryCmdTestCase{
		{
			name:     "cmd error",
			args:     []string{"validate-manage-fees", "--seller-flat-add", "orange", "--market", "419"},
			expInErr: []string{"invalid coin expression: \"orange\""},
		},
		{
			name: "problem with proposal",
			args: []string{"manage-fees-validate", "--market", "420",
				"--seller-ratios-add", "123plum:5plum",
				"--buyer-ratios-add", "123pear:5pear",
			},
			expOut: `error: |-
  seller settlement fee ratios have price denom "plum" but there are no buyer settlement fee ratios with that price denom
  buyer settlement fee ratios have price denom "pear" but there is not a seller settlement fee ratio with that price denom
gov_prop_will_pass: true
`,
		},
		{
			name: "fixes existing problem",
			args: []string{"validate-manage-fees", "--market", "421",
				"--seller-ratios-add", "123plum:5plum", "--output", "json"},
			expOut: `{"error":"","gov_prop_will_pass":true}` + "\n",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runQueryCmdTestCase(tc)
		})
	}
}
