package cli_test

import (
	"fmt"
	"path/filepath"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/x/exchange"
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
			name: "only ask fees, ask",
			args: []string{"order-calc", "--market", "3", "--ask", "--price", "1000peach"},
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
			name: "only ask fees, bid",
			args: []string{"order-calc", "--market", "3", "--bid", "--price", "1000peach"},
			expOut: `creation_fee_options: []
settlement_flat_fee_options: []
settlement_ratio_fee_options: []
`,
		},
		{
			name: "only bid fees, ask",
			args: []string{"order-calc", "--market", "5", "--ask", "--price", "1000peach"},
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
			name: "both fees, bid",
			args: []string{"order-calc", "--market", "420", "--bid", "--price", "1000peach"},
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
			args: []string{"get-market-orders", "420", "--after", "1234567899"},
			expOut: `orders: []
pagination:
  next_key: null
  total: "0"
`,
		},
		{
			name: "several orders",
			args: []string{"market-orders", "--asks", "--market", "420", "--after", "30", "--output", "json", "--count-total"},
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
			name: "no orders",
			args: []string{"asset-orders", "peach"},
			expOut: `orders: []
pagination:
  next_key: null
  total: "0"
`,
		},
		{
			name: "several orders",
			args: []string{"asset-orders", "--denom", "apple", "--limit", "5", "--after", "10", "--output", "json"},
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
			// Hopefully these unit tests don't get up that far.
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

func (s *CmdTestSuite) TestCmdQueryGetCommitment() {
	tests := []queryCmdTestCase{
		{
			name:     "no account or market",
			args:     []string{"commitment"},
			expInErr: []string{"required flag(s) \"account\", \"market\" not set"},
		},
		{
			name: "unknown account and market",
			args: []string{"get-commitment", "--market", "419",
				"--account", sdk.AccAddress("some_account________").String()},
			expOut: "amount: []\n",
		},
		{
			name:   "account with no commitment to market",
			args:   []string{"commitment", "--market", "420", "--account", s.addr9.String()},
			expOut: "amount: []\n",
		},
		{
			name:   "account has commitments in other market",
			args:   []string{"commitment", "--market", "421", "--account", s.addr7.String()},
			expOut: "amount: []\n",
		},
		{
			name: "account has commitment to market: yaml",
			args: []string{"get-commitment", "--market", "420", "--account", s.addr6.String(), "--output", "text"},
			expOut: `amount:
- amount: "10600"
  denom: acorn
- amount: "2200"
  denom: apple
- amount: "4100"
  denom: peach
`,
		},
		{
			name:   "account has commitment to market: json",
			args:   []string{"get-commitment", "--market", "420", "--account", s.addr6.String(), "--output", "json"},
			expOut: `{"amount":[{"denom":"acorn","amount":"10600"},{"denom":"apple","amount":"2200"},{"denom":"peach","amount":"4100"}]}` + "\n",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runQueryCmdTestCase(tc)
		})
	}
}

func (s *CmdTestSuite) TestCmdQueryGetAccountCommitments() {
	tests := []queryCmdTestCase{
		{
			name:     "no account",
			args:     []string{"account-commitments"},
			expInErr: []string{"no <account> provided"},
		},
		{
			name:   "unknown account",
			args:   []string{"get-account-commitments", sdk.AccAddress("unknown_account_____").String()},
			expOut: "commitments: []\n",
		},
		{
			name:   "no commitments",
			args:   []string{"account-commitments", "--account", s.addr8.String()},
			expOut: "commitments: []\n",
		},
		{
			name: "one commitment",
			args: []string{"get-account-commitments", s.addr2.String()},
			expOut: `commitments:
- amount:
  - amount: "900"
    denom: peach
  market_id: 420
`,
		},
		{
			name: "two commitments",
			args: []string{"account-commitments", "--account", s.addr1.String(), "--output", "json"},
			expInOut: []string{
				`{"commitments":[`,
				`{"market_id":420,"amount":[{"denom":"acorn","amount":"10100"}]}`,
				`{"market_id":421,"amount":[{"denom":"apple","amount":"4210"},{"denom":"peach","amount":"421"}]}`,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runQueryCmdTestCase(tc)
		})
	}
}

func (s *CmdTestSuite) TestCmdQueryGetMarketCommitments() {
	coinJSON := func(coin sdk.Coin) string {
		return fmt.Sprintf(`{"denom":"%s","amount":"%s"}`, coin.Denom, coin.Amount)
	}
	coinsJSON := func(coins sdk.Coins) string {
		strs := make([]string, len(coins))
		for i, c := range coins {
			strs[i] = coinJSON(c)
		}
		return "[" + strings.Join(strs, ",") + "]"
	}
	comJSON := func(addr sdk.AccAddress, coins ...sdk.Coin) string {
		return fmt.Sprintf(`{"account":"%s","amount":%s}`, addr.String(), coinsJSON(sdk.NewCoins(coins...)))
	}

	tests := []queryCmdTestCase{
		{
			name:     "no market given",
			args:     []string{"market-commitments"},
			expInErr: []string{"no <market id> provided"},
		},
		{
			name: "market does not exist",
			args: []string{"get-market-commitments", "419"},
			expOut: `commitments: []
pagination:
  next_key: null
  total: "0"
`,
		},
		{
			name: "one commitment",
			args: []string{"market-commitments", "--market", "421"},
			expOut: `commitments:
- account: ` + s.addr1.String() + `
  amount:
  - amount: "4210"
    denom: apple
  - amount: "421"
    denom: peach
pagination:
  next_key: null
  total: "0"
`,
		},
		{
			name: "several commitments",
			args: []string{"market-commitments", "420", "--output", "json", "--count-total"},
			expInOut: []string{
				// Note: The actual ordering is different since it depends on the address bytes,
				// and the addresses in this suite aren't in actual order.
				comJSON(s.addr0, sdk.NewInt64Coin("apple", 1000)),
				comJSON(s.addr1, sdk.NewInt64Coin("acorn", 10100)),
				comJSON(s.addr2, sdk.NewInt64Coin("peach", 900)),
				comJSON(s.addr3, sdk.NewInt64Coin("acorn", 10300), sdk.NewInt64Coin("apple", 1600)),
				comJSON(s.addr4, sdk.NewInt64Coin("apple", 1800), sdk.NewInt64Coin("peach", 2100)),
				comJSON(s.addr5, sdk.NewInt64Coin("acorn", 10500), sdk.NewInt64Coin("peach", 3000)),
				comJSON(s.addr6, sdk.NewInt64Coin("acorn", 10600), sdk.NewInt64Coin("apple", 2200), sdk.NewInt64Coin("peach", 4100)),
				comJSON(s.addr7, sdk.NewInt64Coin("acorn", 10700), sdk.NewInt64Coin("apple", 2400), sdk.NewInt64Coin("peach", 5400)),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.runQueryCmdTestCase(tc)
		})
	}
}

func (s *CmdTestSuite) TestCmdQueryGetAllCommitments() {
	coinJSON := func(coin sdk.Coin) string {
		return fmt.Sprintf(`{"denom":"%s","amount":"%s"}`, coin.Denom, coin.Amount)
	}
	coinsJSON := func(coins sdk.Coins) string {
		strs := make([]string, len(coins))
		for i, c := range coins {
			strs[i] = coinJSON(c)
		}
		return "[" + strings.Join(strs, ",") + "]"
	}
	comJSON := func(addr sdk.AccAddress, marketID uint32, coins ...sdk.Coin) string {
		return fmt.Sprintf(`{"account":"%s","market_id":%d,"amount":%s}`,
			addr.String(), marketID, coinsJSON(sdk.NewCoins(coins...)))
	}

	tests := []queryCmdTestCase{
		{
			name:     "cmd error",
			args:     []string{"all-commitments", "--unexpectedflag"},
			expInErr: []string{"unknown flag: --unexpectedflag"},
		},
		{
			name: "get all",
			args: []string{"get-all-commitments", "--output", "json", "--limit", "10000"},
			expInOut: []string{
				// Note: The actual ordering is different since it depends on the address bytes,
				// and the addresses in this suite aren't in actual order.
				// There might also be more depending on when/how other tests have run.
				comJSON(s.addr0, 420, sdk.NewInt64Coin("apple", 1000)),
				comJSON(s.addr1, 420, sdk.NewInt64Coin("acorn", 10100)),
				comJSON(s.addr2, 420, sdk.NewInt64Coin("peach", 900)),
				comJSON(s.addr3, 420, sdk.NewInt64Coin("acorn", 10300), sdk.NewInt64Coin("apple", 1600)),
				comJSON(s.addr4, 420, sdk.NewInt64Coin("apple", 1800), sdk.NewInt64Coin("peach", 2100)),
				comJSON(s.addr5, 420, sdk.NewInt64Coin("acorn", 10500), sdk.NewInt64Coin("peach", 3000)),
				comJSON(s.addr6, 420, sdk.NewInt64Coin("acorn", 10600), sdk.NewInt64Coin("apple", 2200), sdk.NewInt64Coin("peach", 4100)),
				comJSON(s.addr7, 420, sdk.NewInt64Coin("acorn", 10700), sdk.NewInt64Coin("apple", 2400), sdk.NewInt64Coin("peach", 5400)),
				comJSON(s.addr1, 421, sdk.NewInt64Coin("apple", 4210), sdk.NewInt64Coin("peach", 421)),
			},
		},
		{
			name: "no commitments",
			// The keys here are <market id (4 bytes)><addr length (1 byte)><addr bytes>
			// This page key is the max market id, 32 byte address length, and max address -1.
			// I.e. 37 bytes, all are 255 except the 5th = 32 and last = 254.
			// Hopefully these unit tests don't add something after that.
			args: []string{"all-commitments", "--page-key", "/////yD//////////////////////////////////////////g=="},
			expOut: `commitments: []
pagination:
  next_key: null
  total: "0"
`,
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
  accepting_commitments: true
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
  commitment_settlement_bips: 50
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
  fee_create_commitment_flat:
  - amount: "5"
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
  intermediary_denom: cherry
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
  req_attr_create_commitment:
  - committer.kyc
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
			// Hopefully these unit tests don't get up that far.
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

func (s *CmdTestSuite) TestCmdQueryCommitmentSettlementFeeCalc() {
	tdir := s.T().TempDir()
	filename := filepath.Join(tdir, "commitment-settle.json")
	fileMsg := &exchange.MsgMarketCommitmentSettleRequest{
		Admin:    sdk.AccAddress("msg_admin___________").String(),
		MarketId: 5,
		Inputs: []exchange.AccountAmount{
			{Account: s.addr8.String(), Amount: sdk.NewCoins(sdk.NewInt64Coin("apple", 10), sdk.NewInt64Coin("banana", 15))},
			{Account: s.addr9.String(), Amount: sdk.NewCoins(sdk.NewInt64Coin("peach", 50))},
		},
		Outputs: []exchange.AccountAmount{
			{Account: s.addr9.String(), Amount: sdk.NewCoins(sdk.NewInt64Coin("apple", 10), sdk.NewInt64Coin("banana", 15))},
			{Account: s.addr8.String(), Amount: sdk.NewCoins(sdk.NewInt64Coin("peach", 50))},
		},
		Fees: []exchange.AccountAmount{
			{Account: s.addr8.String(), Amount: sdk.NewCoins(sdk.NewInt64Coin("fig", 4))},
		},
		Navs: []exchange.NetAssetPrice{
			{Assets: sdk.NewInt64Coin("banana", 15), Price: sdk.NewInt64Coin("cherry", 35)},
		},
		EventTag: "the-msg-event-tag",
	}
	tx := newTx(s.T(), fileMsg)
	writeFileAsJson(s.T(), filename, tx)

	feeDenom := pioconfig.GetProvenanceConfig().FeeDenom

	// existing navs:
	// 1cherry => 100<bond denom>
	// 1apple => 8cherry
	// 17acorn => 3cherry
	// 3peach => 778cherry

	tests := []queryCmdTestCase{
		{
			name: "cmd error",
			args: []string{"commitment-settlement-fee-calc",
				"--inputs", s.addr1.String() + ":10apple",
				"--outputs", s.addr2.String() + ":10apple",
			},
			expInErr: []string{"at least one of the flags in the group [file from admin authority] is required"},
		},
		{
			name: "error from endpoint",
			args: []string{"settle-commitments-fee-calc",
				"--admin", s.addr1.String(),
				"--market", "5",
				"--inputs", s.addr2.String() + ":10banana",
				"--outputs", s.addr3.String() + ":10banana",
			},
			expInErr: []string{"InvalidArgument", "no nav found from assets denom \"banana\" to intermediary denom \"cherry\""},
		},
		{
			name: "result without details",
			args: []string{"fee-calc-commitment-settlement",
				"--admin", s.addr9.String(),
				"--market", "5",
				"--inputs", s.addr2.String() + ":30peach",
				"--inputs", s.addr3.String() + ":1945apple",
				"--outputs", s.addr4.String() + ":1945apple,30peach",
				"--navs", "1apple:4cherry",
			},
			// 30peach * 778cherry/3peach = 7780cherry
			// 1945apple * 4cherry/1apple = 7780cherry
			// total = 15560cherry
			// 15560cherry * 100nhash / 1 cherry = 1556000nhash
			// 1556000nhash * 50 / 20000 = 3890nhash
			expOut: `conversion_navs: []
converted_total: []
exchange_fees:
- amount: "3890"
  denom: ` + feeDenom + `
input_total: []
to_fee_nav: null
`,
		},
		{
			name: "result with details",
			args: []string{"fee-calc-settle-commitments", "--file", filename, "--details", "--output", "json"},
			// 10apple * 8cherry/1apple = 80cherry
			// 15banana * 35cherry/15banana = 35cherry
			// 50peach * 778cherry/3peach = 12966.67cherry
			// sum = 13081.67 => 13082cherry
			// 13082cherry * 100nhash/1cherry = 1308200cherry
			// 1308200 * 50 / 20000 = 3270.5 => 3271nhash
			expInOut: []string{
				`"exchange_fees":[{"denom":"nhash","amount":"3271"}]`,
				`"input_total":[{"denom":"apple","amount":"10"},{"denom":"banana","amount":"15"},{"denom":"peach","amount":"50"}]`,
				`"converted_total":[{"denom":"cherry","amount":"13082"}]`,
				`"conversion_navs":[`,
				`{"assets":{"denom":"apple","amount":"1"},"price":{"denom":"cherry","amount":"8"}}`,
				`{"assets":{"denom":"banana","amount":"15"},"price":{"denom":"cherry","amount":"35"}}`,
				`{"assets":{"denom":"peach","amount":"3"},"price":{"denom":"cherry","amount":"778"}}`,
				`"to_fee_nav":{"assets":{"denom":"cherry","amount":"1"},"price":{"denom":"nhash","amount":"100"}}`,
			},
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
