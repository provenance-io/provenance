package cli_test

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	testnet "github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil"
	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/exchange"
	"github.com/provenance-io/provenance/x/exchange/client/cli"
	"github.com/provenance-io/provenance/x/hold"
)

type CmdTestSuite struct {
	suite.Suite

	cfg          testnet.Config
	testnet      *testnet.Network
	keyring      keyring.Keyring
	keyringDir   string
	accountAddrs []sdk.AccAddress

	addr0 sdk.AccAddress
	addr1 sdk.AccAddress
	addr2 sdk.AccAddress
	addr3 sdk.AccAddress
	addr4 sdk.AccAddress
	addr5 sdk.AccAddress
	addr6 sdk.AccAddress
	addr7 sdk.AccAddress
	addr8 sdk.AccAddress
	addr9 sdk.AccAddress

	authorityAddr sdk.AccAddress

	addrNameLookup map[string]string
}

func TestCmdTestSuite(t *testing.T) {
	suite.Run(t, new(CmdTestSuite))
}

func (s *CmdTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")
	pioconfig.SetProvenanceConfig("", 0)
	s.cfg = testutil.DefaultTestNetworkConfig()
	s.cfg.NumValidators = 1
	s.cfg.ChainID = antewrapper.SimAppChainID
	s.cfg.TimeoutCommit = 500 * time.Millisecond

	s.generateAccountsWithKeyring(10)
	s.addr0 = s.accountAddrs[0]
	s.addr1 = s.accountAddrs[1]
	s.addr2 = s.accountAddrs[2]
	s.addr3 = s.accountAddrs[3]
	s.addr4 = s.accountAddrs[4]
	s.addr5 = s.accountAddrs[5]
	s.addr6 = s.accountAddrs[6]
	s.addr7 = s.accountAddrs[7]
	s.addr8 = s.accountAddrs[8]
	s.addr9 = s.accountAddrs[9]
	s.addrNameLookup = map[string]string{
		s.addr0.String(): "addr0",
		s.addr1.String(): "addr1",
		s.addr2.String(): "addr2",
		s.addr3.String(): "addr3",
		s.addr4.String(): "addr4",
		s.addr5.String(): "addr5",
		s.addr6.String(): "addr6",
		s.addr7.String(): "addr7",
		s.addr8.String(): "addr8",
		s.addr9.String(): "addr9",
	}

	s.authorityAddr = authtypes.NewModuleAddress(govtypes.ModuleName)
	s.addrNameLookup[s.authorityAddr.String()] = "authorityAddr"

	// Add accounts to auth gen state.
	var authGen authtypes.GenesisState
	err := s.cfg.Codec.UnmarshalJSON(s.cfg.GenesisState[authtypes.ModuleName], &authGen)
	s.Require().NoError(err, "UnmarshalJSON auth gen state")
	genAccs := make(authtypes.GenesisAccounts, len(s.accountAddrs))
	for i, addr := range s.accountAddrs {
		genAccs[i] = authtypes.NewBaseAccount(addr, nil, 0, 1)
	}
	newAccounts, err := authtypes.PackAccounts(genAccs)
	s.Require().NoError(err, "PackAccounts")
	authGen.Accounts = append(authGen.Accounts, newAccounts...)
	s.cfg.GenesisState[authtypes.ModuleName], err = s.cfg.Codec.MarshalJSON(&authGen)
	s.Require().NoError(err, "MarshalJSON auth gen state")

	// Add some markets to the exchange gen state.
	var exchangeGen exchange.GenesisState
	err = s.cfg.Codec.UnmarshalJSON(s.cfg.GenesisState[exchange.ModuleName], &exchangeGen)
	s.Require().NoError(err, "UnmarshalJSON exchange gen state")
	exchangeGen.Params = exchange.DefaultParams()
	exchangeGen.Markets = append(exchangeGen.Markets,
		exchange.Market{
			MarketId: 3,
			MarketDetails: exchange.MarketDetails{
				Name:        "Market Three",
				Description: "The third market (or is it?). It only has ask/seller fees.",
			},
			FeeCreateAskFlat:        []sdk.Coin{sdk.NewInt64Coin("peach", 10)},
			FeeSellerSettlementFlat: []sdk.Coin{sdk.NewInt64Coin("peach", 50)},
			FeeSellerSettlementRatios: []exchange.FeeRatio{
				{Price: sdk.NewInt64Coin("peach", 100), Fee: sdk.NewInt64Coin("peach", 1)},
			},
			AcceptingOrders:     true,
			AllowUserSettlement: true,
			AccessGrants: []exchange.AccessGrant{
				{Address: s.addr1.String(), Permissions: exchange.AllPermissions()},
				{Address: s.addr2.String(), Permissions: exchange.AllPermissions()},
				{Address: s.addr3.String(), Permissions: []exchange.Permission{exchange.Permission_cancel, exchange.Permission_attributes}},
			},
		},
		exchange.Market{
			MarketId: 5,
			MarketDetails: exchange.MarketDetails{
				Name:        "Market Five",
				Description: "Market the Fifth. It only has bid/buyer fees.",
			},
			FeeCreateBidFlat:       []sdk.Coin{sdk.NewInt64Coin("peach", 10)},
			FeeBuyerSettlementFlat: []sdk.Coin{sdk.NewInt64Coin("peach", 50)},
			FeeBuyerSettlementRatios: []exchange.FeeRatio{
				{Price: sdk.NewInt64Coin("peach", 100), Fee: sdk.NewInt64Coin("peach", 1)},
				{Price: sdk.NewInt64Coin("peach", 100), Fee: sdk.NewInt64Coin(s.cfg.BondDenom, 3)},
			},
			AcceptingOrders:     true,
			AllowUserSettlement: true,
			AccessGrants: []exchange.AccessGrant{
				{Address: s.addr1.String(), Permissions: exchange.AllPermissions()},
			},
		},
		// Do not make a market 419, lots of tests expect it to not exist.
		exchange.Market{
			// The orders in this market are for the orders queries.
			// Don't use it in other unit tests (e.g. order creation or settlement).
			MarketId: 420,
			MarketDetails: exchange.MarketDetails{
				Name:        "THE Market",
				Description: "It's coming; you know it. It has all the fees.",
			},
			FeeCreateAskFlat:        []sdk.Coin{sdk.NewInt64Coin("peach", 20)},
			FeeCreateBidFlat:        []sdk.Coin{sdk.NewInt64Coin("peach", 25)},
			FeeSellerSettlementFlat: []sdk.Coin{sdk.NewInt64Coin("peach", 100)},
			FeeSellerSettlementRatios: []exchange.FeeRatio{
				{Price: sdk.NewInt64Coin("peach", 75), Fee: sdk.NewInt64Coin("peach", 1)},
			},
			FeeBuyerSettlementFlat: []sdk.Coin{sdk.NewInt64Coin("peach", 105)},
			FeeBuyerSettlementRatios: []exchange.FeeRatio{
				{Price: sdk.NewInt64Coin("peach", 50), Fee: sdk.NewInt64Coin("peach", 1)},
				{Price: sdk.NewInt64Coin("peach", 50), Fee: sdk.NewInt64Coin(s.cfg.BondDenom, 3)},
			},
			AcceptingOrders:     true,
			AllowUserSettlement: true,
			AccessGrants: []exchange.AccessGrant{
				{Address: s.addr1.String(), Permissions: exchange.AllPermissions()},
			},
			ReqAttrCreateAsk: []string{"seller.kyc"},
			ReqAttrCreateBid: []string{"buyer.kyc"},
		},
		exchange.Market{
			// This market has an invalid setup. Don't mess with it.
			MarketId:      421,
			MarketDetails: exchange.MarketDetails{Name: "Broken"},
			FeeSellerSettlementRatios: []exchange.FeeRatio{
				{Price: sdk.NewInt64Coin("peach", 55), Fee: sdk.NewInt64Coin("peach", 1)},
			},
			FeeBuyerSettlementRatios: []exchange.FeeRatio{
				{Price: sdk.NewInt64Coin("peach", 56), Fee: sdk.NewInt64Coin("peach", 1)},
				{Price: sdk.NewInt64Coin("plum", 57), Fee: sdk.NewInt64Coin("plum", 1)},
			},
		},
	)
	toHold := make(map[string]sdk.Coins)
	exchangeGen.Orders = make([]exchange.Order, 60)
	for i := range exchangeGen.Orders {
		order := s.makeInitialOrder(uint64(i + 1))
		exchangeGen.Orders[i] = *order
		toHold[order.GetOwner()] = toHold[order.GetOwner()].Add(order.GetHoldAmount()...)
	}
	exchangeGen.LastOrderId = uint64(100)
	s.cfg.GenesisState[exchange.ModuleName], err = s.cfg.Codec.MarshalJSON(&exchangeGen)
	s.Require().NoError(err, "MarshalJSON exchange gen state")

	// Create all the needed holds.
	var holdGen hold.GenesisState
	err = s.cfg.Codec.UnmarshalJSON(s.cfg.GenesisState[hold.ModuleName], &holdGen)
	s.Require().NoError(err, "UnmarshalJSON hold gen state")
	for _, addr := range s.accountAddrs {
		holdGen.Holds = append(holdGen.Holds, &hold.AccountHold{
			Address: addr.String(),
			Amount:  toHold[addr.String()],
		})
	}
	s.cfg.GenesisState[hold.ModuleName], err = s.cfg.Codec.MarshalJSON(&holdGen)
	s.Require().NoError(err, "MarshalJSON hold gen state")

	// Add balances to bank gen state.
	// Any initial holds for an account are added to this so that
	// this what's available to each at the start of the unit tests.
	balance := sdk.NewCoins(
		sdk.NewInt64Coin(s.cfg.BondDenom, 1_000_000_000),
		sdk.NewInt64Coin("acorn", 1_000_000_000),
		sdk.NewInt64Coin("apple", 1_000_000_000),
		sdk.NewInt64Coin("peach", 1_000_000_000),
	)
	var bankGen banktypes.GenesisState
	err = s.cfg.Codec.UnmarshalJSON(s.cfg.GenesisState[banktypes.ModuleName], &bankGen)
	s.Require().NoError(err, "UnmarshalJSON bank gen state")
	for _, addr := range s.accountAddrs {
		bal := balance.Add(toHold[addr.String()]...)
		bankGen.Balances = append(bankGen.Balances, banktypes.Balance{Address: addr.String(), Coins: bal})
	}
	s.cfg.GenesisState[banktypes.ModuleName], err = s.cfg.Codec.MarshalJSON(&bankGen)
	s.Require().NoError(err, "MarshalJSON bank gen state")

	// And fire it all up!!
	s.testnet, err = testnet.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err, "testnet.New(...)")

	_, err = s.testnet.WaitForHeight(1)
	s.Require().NoError(err, "s.testnet.WaitForHeight(1)")
}

func (s *CmdTestSuite) TearDownSuite() {
	testutil.CleanUp(s.testnet, s.T())
}

// generateAccountsWithKeyring creates a keyring and adds a number of keys to it.
// The s.keyringDir, s.keyring, and s.accountAddrs are all set in here.
// The getClientCtx function returns a context that knows about this keyring.
func (s *CmdTestSuite) generateAccountsWithKeyring(number int) {
	path := hd.CreateHDPath(118, 0, 0).String()
	s.keyringDir = s.T().TempDir()
	var err error
	s.keyring, err = keyring.New(s.T().Name(), "test", s.keyringDir, nil, s.cfg.Codec)
	s.Require().NoError(err, "keyring.New(...)")

	s.accountAddrs = make([]sdk.AccAddress, number)
	for i := range s.accountAddrs {
		keyId := fmt.Sprintf("test_key_%v", i)
		var info *keyring.Record
		info, _, err = s.keyring.NewMnemonic(keyId, keyring.English, path, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
		s.Require().NoError(err, "[%d] s.keyring.NewMnemonic(...)", i)
		s.accountAddrs[i], err = info.GetAddress()
		s.Require().NoError(err, "[%d] getting keyring address", i)
	}
}

// makeInitialOrder makes an order using the order id for various aspects.
func (s *CmdTestSuite) makeInitialOrder(orderID uint64) *exchange.Order {
	addr := s.accountAddrs[int(orderID)%len(s.accountAddrs)]
	assetDenom := "apple"
	if orderID%7 <= 2 {
		assetDenom = "acorn"
	}
	assets := sdk.NewInt64Coin(assetDenom, int64(orderID*100))
	price := sdk.NewInt64Coin("peach", int64(orderID*orderID*10))
	partial := orderID%2 == 0
	order := exchange.NewOrder(orderID)
	switch orderID % 6 {
	case 0, 1, 4:
		order.WithAsk(&exchange.AskOrder{
			MarketId:     420,
			Seller:       addr.String(),
			Assets:       assets,
			Price:        price,
			ExternalId:   fmt.Sprintf("my-id-%d", orderID),
			AllowPartial: partial,
		})
	case 2, 3, 5:
		order.WithBid(&exchange.BidOrder{
			MarketId:     420,
			Buyer:        addr.String(),
			Assets:       assets,
			Price:        price,
			ExternalId:   fmt.Sprintf("my-id-%d", orderID),
			AllowPartial: partial,
		})
	}
	return order
}

// getClientCtx get a client context that knows about the suite's keyring.
func (s *CmdTestSuite) getClientCtx() client.Context {
	return s.testnet.Validators[0].ClientCtx.
		WithKeyringDir(s.keyringDir).
		WithKeyring(s.keyring)
}

// getAddrName tries to get the variable name (in this suite) of the provided address.
func (s *CmdTestSuite) getAddrName(addr string) string {
	if rv, found := s.addrNameLookup[addr]; found {
		return rv
	}
	return addr
}

// txCmdTestCase is a test case for a TX command.
type txCmdTestCase struct {
	// name is a name for this test case.
	name string
	// preRun is a function that is run first.
	// It should return any arguments to append to the args and a function that will
	// run any fallow-up checks to do after the command is run.
	preRun func() ([]string, func(*sdk.TxResponse))
	// args are the arguments to provide to the command.
	args []string
	// expInErr are strings to expect in an error from the cmd.
	// Errors that come from the endpoint will not be here; use expInRawLog for those.
	expInErr []string
	// expInRawLog are strings to expect in the TxResponse.RawLog.
	expInRawLog []string
	// expectedCode is the code expected from the Tx.
	expectedCode uint32
}

// RunTxCmdTestCase runs a txCmdTestCase by executing the command and checking the result.
func (s *CmdTestSuite) runTxCmdTestCase(tc txCmdTestCase) {
	s.T().Helper()
	var extraArgs []string
	var followup func(*sdk.TxResponse)
	var preRunFailed bool
	if tc.preRun != nil {
		s.Run("pre-run: "+tc.name, func() {
			preRunFailed = true
			extraArgs, followup = tc.preRun()
			preRunFailed = s.T().Failed()
		})
	}

	cmd := cli.CmdTx()

	args := append(tc.args, extraArgs...)
	args = append(args,
		"--"+flags.FlagGas, "250000",
		"--"+flags.FlagFees, s.bondCoins(10).String(),
		"--"+flags.FlagBroadcastMode, flags.BroadcastBlock,
		"--"+flags.FlagSkipConfirmation,
	)

	var txResponse *sdk.TxResponse
	var cmdFailed bool
	testRunner := func() {
		if preRunFailed {
			s.T().Skip("Skipping execution due to pre-run failure.")
		}

		cmdName := cmd.Name()
		var outBz []byte
		defer func() {
			if s.T().Failed() {
				s.T().Logf("Command: %s\nArgs: %q\nOutput\n%s", cmdName, args, string(outBz))
				cmdFailed = true
			}
		}()

		clientCtx := s.getClientCtx()
		out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
		outBz = out.Bytes()

		s.assertErrorContents(err, tc.expInErr, "ExecTestCLICmd error")
		for _, exp := range tc.expInErr {
			s.Assert().Contains(string(outBz), exp, "command output should contain:\n%q", exp)
		}

		if len(tc.expInErr) == 0 && err == nil {
			var resp sdk.TxResponse
			err = clientCtx.Codec.UnmarshalJSON(outBz, &resp)
			if s.Assert().NoError(err, "UnmarshalJSON(command output) error") {
				txResponse = &resp
				s.Assert().Equal(int(tc.expectedCode), int(resp.Code), "response code")
				for _, exp := range tc.expInRawLog {
					s.Assert().Contains(resp.RawLog, exp, "TxResponse.RawLog should countain:\n%q", exp)
				}
			}
		}
	}

	if tc.preRun != nil {
		s.Run("execute: "+tc.name, testRunner)
	} else {
		testRunner()
	}

	if followup != nil {
		s.Run("followup: "+tc.name, func() {
			if preRunFailed {
				s.T().Skip("Skipping followup due to pre-run failure.")
			}
			if cmdFailed {
				s.T().Skip("Skipping followup due to failure with command.")
			}
			if s.Assert().NotNil(txResponse, "the TxResponse from the command output") {
				followup(txResponse)
			}
		})
	}
}

// queryCmdTestCase is a test case of a query command.
type queryCmdTestCase struct {
	// name is a name for this test case.
	name string
	// args are the arguments to provide to the command.
	args []string
	// expInErr are strings to expect in an error message (and output).
	expInErr []string
	// expInOut are strings to expect in the output.
	expInOut []string
	// expOut is the expected full output. Leave empty to skip this check.
	expOut string
}

// RunQueryCmdTestCase runs a queryCmdTestCase by executing the command and checking the result.
func (s *CmdTestSuite) runQueryCmdTestCase(tc queryCmdTestCase) {
	s.T().Helper()
	cmd := cli.CmdQuery()

	cmdName := cmd.Name()
	var outStr string
	defer func() {
		if s.T().Failed() {
			s.T().Logf("Command: %s\nArgs: %q\nOutput\n%s", cmdName, tc.args, outStr)
		}
	}()

	clientCtx := s.getClientCtx()
	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
	outStr = out.String()

	s.assertErrorContents(err, tc.expInErr, "ExecTestCLICmd error")
	for _, exp := range tc.expInErr {
		if !s.Assert().Contains(outStr, exp, "command output (error)") {
			s.T().Logf("Not found: %q", exp)
		}
	}

	for _, exp := range tc.expInOut {
		if !s.Assert().Contains(outStr, exp, "command output:\n%q", exp) {
			s.T().Logf("Not found: %q", exp)
		}
	}

	if len(tc.expOut) > 0 {
		s.Assert().Equal(tc.expOut, outStr, "command output string")
	}
}

// getEventAttribute finds the value of an attribute in an event.
// Returns an error if the value is empty, the attribute doesn't exist, or the event doesn't exist.
func (s *CmdTestSuite) getEventAttribute(events []abci.Event, eventType, attribute string) (string, error) {
	for _, event := range events {
		if event.Type == eventType {
			for _, attr := range event.Attributes {
				if string(attr.Key) == attribute {
					val := strings.Trim(string(attr.Value), `"`)
					if len(val) > 0 {
						return val, nil
					}
					return "", fmt.Errorf("the %s.%s value is empty", eventType, attribute)
				}
			}
			return "", fmt.Errorf("no %s attribute found in %s", attribute, eventType)
		}
	}
	return "", fmt.Errorf("no %s found", eventType)
}

// findNewOrderID gets the order id from the EventOrderCreated event.
func (s *CmdTestSuite) findNewOrderID(resp *sdk.TxResponse) (string, error) {
	return s.getEventAttribute(resp.Events, "provenance.exchange.v1.EventOrderCreated", "order_id")
}

// assertOrder uses the GetOrder query to look up an order and make sure it equals the one provided.
// If the provided order is nil, ensures the query returns an order not found error.
func (s *CmdTestSuite) assertGetOrder(orderID string, order *exchange.Order) (okay bool) {
	s.T().Helper()
	if !s.Assert().NotEmpty(orderID, "order id") {
		return false
	}

	var expInErr []string
	if order == nil {
		expInErr = append(expInErr, fmt.Sprintf("order %s not found", orderID))
	}

	var getOrderOutBz []byte
	getOrderArgs := []string{orderID, "--output", "json"}
	defer func() {
		if !okay {
			s.T().Logf("Query GetOrder %s output:\n%s", getOrderArgs, string(getOrderOutBz))
		}
	}()

	clientCtx := s.getClientCtx()
	getOrderCmd := cli.CmdQueryGetOrder()
	getOrderOutBW, err := clitestutil.ExecTestCLICmd(clientCtx, getOrderCmd, getOrderArgs)
	getOrderOutBz = getOrderOutBW.Bytes()
	if !s.assertErrorContents(err, expInErr, "ExecTestCLICmd GetOrder %s error", orderID) {
		return false
	}

	if order == nil {
		return true
	}

	var resp exchange.QueryGetOrderResponse
	err = clientCtx.Codec.UnmarshalJSON(getOrderOutBz, &resp)
	if !s.Assert().NoError(err, "UnmarshalJSON on GetOrder %s response", orderID) {
		return false
	}
	return s.Assert().Equal(order, resp.Order, "order %s", orderID)
}

// getOrderFollowup returns a follow-up function that looks up an order and makes sure it's the one provided.
func (s *CmdTestSuite) getOrderFollowup(orderID string, order *exchange.Order) func(*sdk.TxResponse) {
	return func(*sdk.TxResponse) {
		if order != nil {
			order.OrderId = s.asOrderID(orderID)
		}
		s.assertGetOrder(orderID, order)
	}
}

// createOrderFollowup returns a followup function that identifies the new order id, looks it up,
// and makes sure it is as expected.
func (s *CmdTestSuite) createOrderFollowup(order *exchange.Order) func(*sdk.TxResponse) {
	return func(resp *sdk.TxResponse) {
		orderID, err := s.findNewOrderID(resp)
		if s.Assert().NoError(err, "finding new order id") {
			order.OrderId = s.asOrderID(orderID)
			s.assertGetOrder(orderID, order)
		}
	}
}

// getMarket executes a query to get the given market.
func (s *CmdTestSuite) getMarket(marketID string) *exchange.Market {
	s.T().Helper()
	if !s.Assert().NotEmpty(marketID, "market id") {
		return nil
	}

	okay := false
	var outBz []byte
	args := []string{marketID, "--output", "json"}
	defer func() {
		if !okay {
			s.T().Logf("Query GetMarket\nArgs: %q\nOutput:\n%s", args, string(outBz))
		}
	}()

	clientCtx := s.getClientCtx()
	cmd := cli.CmdQueryGetMarket()
	outBW, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	outBz = outBW.Bytes()

	s.Require().NoError(err, "ExecTestCLICmd error")

	var resp exchange.QueryGetMarketResponse
	err = clientCtx.Codec.UnmarshalJSON(outBz, &resp)
	s.Require().NoError(err, "UnmarshalJSON on GetMarket %s response", marketID)
	s.Require().NotNil(resp.Market, "GetMarket %s response .Market", marketID)
	okay = true
	return resp.Market
}

// getMarketFollowup returns a follow-up function that asserts that the existing market is as expected.
func (s *CmdTestSuite) getMarketFollowup(marketID string, expected *exchange.Market) func(*sdk.TxResponse) {
	return func(_ *sdk.TxResponse) {
		actual := s.getMarket(marketID)
		s.Assert().Equal(expected, actual, "market %s", marketID)
	}
}

// findNewProposalID gets the proposal id from the submit_proposal event.
func (s *CmdTestSuite) findNewProposalID(resp *sdk.TxResponse) (string, error) {
	return s.getEventAttribute(resp.Events, "submit_proposal", "proposal_id")
}

// AssertGovPropMsg queries for the given proposal and makes sure it's got just the provided Msg.
func (s *CmdTestSuite) assertGovPropMsg(propID string, msg sdk.Msg) bool {
	s.T().Helper()
	if msg == nil {
		return true
	}

	if !s.Assert().NotEmpty(propID, "proposal id") {
		return false
	}
	expPropMsgAny, err := codectypes.NewAnyWithValue(msg)
	if !s.Assert().NoError(err, "NewAnyWithValue(%T)", msg) {
		return false
	}

	clientCtx := s.getClientCtx()
	getPropCmd := govcli.GetCmdQueryProposal()
	propOutBW, err := clitestutil.ExecTestCLICmd(clientCtx, getPropCmd, []string{propID, "--output", "json"})
	propOutBz := propOutBW.Bytes()
	s.T().Logf("Query proposal %s output:\n%s", propID, string(propOutBz))
	if !s.Assert().NoError(err, "GetCmdQueryProposal %s error", propID) {
		return false
	}

	var prop govv1.Proposal
	err = clientCtx.Codec.UnmarshalJSON(propOutBz, &prop)
	if !s.Assert().NoError(err, "UnmarshalJSON on proposal %s response", propID) {
		return false
	}
	if !s.Assert().Len(prop.Messages, 1, "number of messages in proposal %s", propID) {
		return false
	}
	return s.Assert().Equal(expPropMsgAny, prop.Messages[0], "the message in proposal %s", propID)
}

// govPropFollowup returns a followup function that identifies the new proposal id, looks it up,
// and makes sure it's got the provided msg.
func (s *CmdTestSuite) govPropFollowup(msg sdk.Msg) func(*sdk.TxResponse) {
	return func(resp *sdk.TxResponse) {
		propID, err := s.findNewProposalID(resp)
		if s.Assert().NoError(err, "finding new proposal id") {
			s.assertGovPropMsg(propID, msg)
		}
	}
}

// assertErrorContents is a wrapper for assertions.AssertErrorContents using this suite's T().
func (s *CmdTestSuite) assertErrorContents(theError error, contains []string, msgAndArgs ...interface{}) bool {
	s.T().Helper()
	return assertions.AssertErrorContents(s.T(), theError, contains, msgAndArgs...)
}

// bondCoins returns a Coins with just an entry with the bond denom and the provided amount.
func (s *CmdTestSuite) bondCoins(amt int64) sdk.Coins {
	return sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, amt))
}

// createOrder issues a command to create the provided order and returns its order id.
func (s *CmdTestSuite) createOrder(order *exchange.Order, creationFee *sdk.Coin) uint64 {
	cmd := cli.CmdTx()
	args := []string{
		order.GetOrderType(),
		"--market", fmt.Sprintf("%d", order.GetMarketID()),
		"--from", order.GetOwner(),
		"--assets", order.GetAssets().String(),
		"--price", order.GetPrice().String(),
	}
	settleFee := order.GetSettlementFees()
	if !settleFee.IsZero() {
		args = append(args, "--settlement-fee", settleFee.String())
	}
	if order.PartialFillAllowed() {
		args = append(args, "--partial")
	}
	eid := order.GetExternalID()
	if len(eid) > 0 {
		args = append(args, "--external-id", eid)
	}
	if creationFee != nil {
		args = append(args, "--creation-fee", creationFee.String())
	}
	args = append(args,
		"--"+flags.FlagFees, s.bondCoins(10).String(),
		"--"+flags.FlagBroadcastMode, flags.BroadcastBlock,
		"--"+flags.FlagSkipConfirmation,
	)

	cmdName := cmd.Name()
	var outBz []byte
	defer func() {
		if s.T().Failed() {
			s.T().Logf("Command: %s\nArgs: %q\nOutput\n%s", cmdName, args, string(outBz))
		}
	}()

	clientCtx := s.getClientCtx()
	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	outBz = out.Bytes()

	s.Require().NoError(err, "ExecTestCLICmd error")

	var resp sdk.TxResponse
	err = clientCtx.Codec.UnmarshalJSON(outBz, &resp)
	s.Require().NoError(err, "UnmarshalJSON(command output) error")
	orderIDStr, err := s.findNewOrderID(&resp)
	s.Require().NoError(err, "findNewOrderID")
	return s.asOrderID(orderIDStr)
}

// queryBankBalances executes a bank query to get an account's balances.
func (s *CmdTestSuite) queryBankBalances(addr string) sdk.Coins {
	clientCtx := s.getClientCtx()
	cmd := bankcli.GetBalancesCmd()
	args := []string{addr, "--output", "json"}
	outBW, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	s.Require().NoError(err, "ExecTestCLICmd %s %q", cmd.Name(), args)
	outBz := outBW.Bytes()

	var resp banktypes.QueryAllBalancesResponse
	err = clientCtx.Codec.UnmarshalJSON(outBz, &resp)
	s.Require().NoError(err, "UnmarshalJSON(%q, %T)", string(outBz), &resp)
	return resp.Balances
}

// execBankSend executes a bank send command.
func (s *CmdTestSuite) execBankSend(fromAddr, toAddr, amount string) {
	clientCtx := s.getClientCtx()
	cmd := bankcli.NewSendTxCmd()
	cmdName := cmd.Name()
	args := []string{
		fromAddr, toAddr, amount,
		"--" + flags.FlagFees, s.bondCoins(10).String(),
		"--" + flags.FlagBroadcastMode, flags.BroadcastBlock,
		"--" + flags.FlagSkipConfirmation,
	}
	failed := true
	var outStr string
	defer func() {
		if failed {
			s.T().Logf("Command: %s\nArgs: %q\nOutput\n%s", cmdName, args, outStr)
		}
	}()

	outBW, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	outStr = outBW.String()
	s.Require().NoError(err, "ExecTestCLICmd %s %q", cmdName, args)
	failed = false
}

// adjustBalance creates a new Balance with the order owner's Address and a Coins that's
// the result of the order and fees applied to the provided current balance.
func (s *CmdTestSuite) adjustBalance(curBal sdk.Coins, order *exchange.Order, creationFees ...sdk.Coin) banktypes.Balance {
	rv := banktypes.Balance{
		Address: order.GetOwner(),
	}

	price := order.GetPrice()
	assets := order.GetAssets()
	var hasNeg bool
	if order.IsAskOrder() {
		rv.Coins, hasNeg = curBal.Add(price).SafeSub(assets)
		s.Require().False(hasNeg, "hasNeg: %s + %s - %s", curBal, price, assets)
	}
	if order.IsBidOrder() {
		rv.Coins, hasNeg = curBal.Add(assets).SafeSub(price)
		s.Require().False(hasNeg, "hasNeg: %s + %s - %s", curBal, assets, price)
	}

	settleFees := order.GetSettlementFees()
	if !settleFees.IsZero() {
		orig := rv.Coins
		rv.Coins, hasNeg = rv.Coins.SafeSub(settleFees...)
		s.Require().False(hasNeg, "hasNeg (settlement fees): %s - %s", orig, settleFees)
	}

	for _, fee := range creationFees {
		orig := rv.Coins
		rv.Coins, hasNeg = rv.Coins.SafeSub(fee)
		s.Require().False(hasNeg, "hasNeg (creation fee): %s - %s", orig, fee)
	}

	return rv
}

// assertBalancesFollowup returns a follow-up function that asserts that the balances are now as expected.
func (s *CmdTestSuite) assertBalancesFollowup(expBals []banktypes.Balance) func(*sdk.TxResponse) {
	return func(_ *sdk.TxResponse) {
		for _, expBal := range expBals {
			actBal := s.queryBankBalances(expBal.Address)
			s.Assert().Equal(expBal.Coins.String(), actBal.String(), "%s balances", s.getAddrName(expBal.Address))
		}
	}
}

// joinErrs joins the provided error strings matching to how errors.Join does.
func joinErrs(errs ...string) string {
	return strings.Join(errs, "\n")
}

// toStringSlice applies the stringer to each value and returns a slice with the results.
func toStringSlice[T any](vals []T, stringer func(T) string) []string {
	if vals == nil {
		return nil
	}
	rv := make([]string, len(vals))
	for i, val := range vals {
		rv[i] = stringer(val)
	}
	return rv
}

// assertEqualSlices asserts that the two slices are equal; returns true if so.
// If not, the stringer is applied to each entry and the comparison is redone
// using the strings for a more helpful failure message.
func assertEqualSlices[T any](t *testing.T, expected []T, actual []T, stringer func(T) string, message string, args ...interface{}) bool {
	t.Helper()
	if assert.Equalf(t, expected, actual, message, args...) {
		return true
	}
	expStrs := toStringSlice(expected, stringer)
	actStrs := toStringSlice(actual, stringer)
	assert.Equalf(t, expStrs, actStrs, message+" as strings", args...)
	return false
}

// splitStringer makes a string from the provided DenomSplit.
func splitStringer(split exchange.DenomSplit) string {
	return fmt.Sprintf("%s:%d", split.Denom, split.Split)
}

// orderIDStringer converts an order id to a string.
func orderIDStringer(orderID uint64) string {
	return fmt.Sprintf("%d", orderID)
}

// asOrderID converts the provided string into an order id.
func (s *CmdTestSuite) asOrderID(str string) uint64 {
	rv, err := strconv.ParseUint(str, 10, 64)
	s.Require().NoError(err, "ParseUint(%q, 10, 64)", str)
	return rv
}

// truncate truncates the provided string returning at most length characters.
func truncate(str string, length int) string {
	if len(str) < length-3 {
		return str
	}
	return str[:length-3] + "..."
}

const (
	// mutExc is the annotation type for "mutually exclusive".
	// It equals the cobra.Command.mutuallyExclusive variable.
	mutExc = "cobra_annotation_mutually_exclusive"
	// oneReq is the annotation type for "one required".
	// It equals the cobra.Command.oneRequired variable.
	oneReq = "cobra_annotation_one_required"
	// mutExc is the annotation type for "required".
	required = cobra.BashCompOneRequiredFlag
)

// setupTestCase contains the stuff that runSetupTestCase should check.
type setupTestCase struct {
	// name is the name of the setup func being tested.
	name string
	// setup is the function being tested.
	setup func(cmd *cobra.Command)
	// expFlags is the list of flags expected to be added to the command after setup.
	// The flags.FlagFrom flag is added to the command prior to calling the setup func;
	// it should be included in this list if you want to check its annotations.
	expFlags []string
	// expAnnotations is the annotations expected for each of the expFlags.
	// The map is "flag name" -> "annotation type" -> values
	// The following variables have the annotation type strings: mutExc, oneReq, required.
	// Annotions are only checked on the flags listed in expFlags.
	expAnnotations map[string]map[string][]string
	// expInUse is a set of strings that are expected to be in the command's Use string.
	// Each entry that does not start with a "[" is also checked to not be in the Use wrapped in [].
	expInUse []string
	// expExamples is a set of examples to ensure are on the command.
	// There must be a full line in the command's Example that matches each entry.
	expExamples []string
	// skipArgsCheck true causes the runner to skip the check ensuring that the command's Args func has been set.
	skipArgsCheck bool
}

// runSetupTestCase runs the provided setup func and checks that everything is set up as expected.
func runSetupTestCase(t *testing.T, tc setupTestCase) {
	if tc.expAnnotations == nil {
		tc.expAnnotations = make(map[string]map[string][]string)
	}
	cmd := &cobra.Command{
		Use: "dummy",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("the dummy command should not have been executed")
		},
	}
	cmd.Flags().String(flags.FlagFrom, "", "The from flag")

	testFunc := func() {
		tc.setup(cmd)
	}
	require.NotPanics(t, testFunc, tc.name)

	for i, flagName := range tc.expFlags {
		t.Run(fmt.Sprintf("flag[%d]: --%s", i, flagName), func(t *testing.T) {
			flag := cmd.Flags().Lookup(flagName)
			if assert.NotNil(t, flag, "--%s", flagName) {
				expAnnotations, _ := tc.expAnnotations[flagName]
				actAnnotations := flag.Annotations
				assert.Equal(t, expAnnotations, actAnnotations, "--%s annotations", flagName)
			}
		})
	}

	for i, exp := range tc.expInUse {
		t.Run(fmt.Sprintf("use[%d]: %s", i, truncate(exp, 20)), func(t *testing.T) {
			assert.Contains(t, cmd.Use, exp, "command use after %s", tc.name)
			if exp[0] != '[' {
				assert.NotContains(t, cmd.Use, "["+exp+"]", "command use after %s", tc.name)
			}
		})
	}

	examples := strings.Split(cmd.Example, "\n")
	for i, exp := range tc.expExamples {
		t.Run(fmt.Sprintf("examples[%d]", i), func(t *testing.T) {
			assert.Contains(t, examples, exp, "command examples after %s", tc.name)
		})
	}

	if !tc.skipArgsCheck {
		t.Run("args", func(t *testing.T) {
			assert.NotNil(t, cmd.Args, "command args after %s", tc.name)
		})
	}
}
