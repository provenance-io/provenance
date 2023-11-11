package cli_test

import (
	"errors"
	"fmt"
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
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
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
		},
	)
	toHold := make(map[string]sdk.Coins)
	exchangeGen.Orders = make([]exchange.Order, 60)
	for i := range exchangeGen.Orders {
		order := s.createInitialOrder(uint64(i + 1))
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

// createInitialOrder creates an order using the order id for various aspects.
func (s *CmdTestSuite) createInitialOrder(orderID uint64) *exchange.Order {
	addr := s.accountAddrs[int(orderID)%len(s.accountAddrs)]
	assets := sdk.NewInt64Coin("apple", int64(orderID*100))
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

// txCmdTestCase is a test case for a TX command.
type txCmdTestCase struct {
	// name is a name for this test case.
	name string
	// cmdGen is a generator for the command to execute. If not set, CmdTx is used.
	cmdGen func() *cobra.Command
	// args are the arguments to provide to the command.
	args []string
	// expInErr are strings to expect in an error message (and the output).
	expInErr []string
	// expectedCode is the code expected from the Tx.
	expectedCode uint32
	// followup is any further checks to do.
	followup func(txResponse *sdk.TxResponse)
}

// RunTxCmdTestCase runs a txCmdTestCase by executing the command and checking the result.
func (s *CmdTestSuite) runTxCmdTestCase(tc txCmdTestCase) {
	s.T().Helper()
	var cmd *cobra.Command
	if tc.cmdGen != nil {
		cmd = tc.cmdGen()
	} else {
		cmd = cli.CmdTx()
	}

	cmdName := cmd.Name()
	var outBz []byte
	defer func() {
		if s.T().Failed() {
			s.T().Logf("Command: %s\nArgs: %q\nOutput\n%s", cmdName, tc.args, string(outBz))
		}
	}()

	clientCtx := s.getClientCtx()
	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
	outBz = out.Bytes()

	s.assertErrorContents(err, tc.expInErr, "ExecTestCLICmd error")
	for _, exp := range tc.expInErr {
		s.Assert().Contains(string(outBz), exp, "command output should contain:\n%q", exp)
	}

	var txResponse *sdk.TxResponse
	if len(tc.expInErr) == 0 && err == nil {
		var resp sdk.TxResponse
		err = clientCtx.Codec.UnmarshalJSON(outBz, &resp)
		if s.Assert().NoError(err, "UnmarshalJSON(command output) error") {
			txResponse = &resp
			if s.Assert().Equalf(int(tc.expectedCode), int(resp.Code), "response code") {
				s.T().Logf("TxResponse:\n%v", resp)
			}
		}
	}

	if tc.followup != nil {
		err = s.testnet.WaitForNextBlock()
		if s.Assert().NoError(err, "waiting for next block before followup") {
			if s.Assert().NotNil(txResponse, "the TxResponse from the command output") {
				tc.followup(txResponse)
			}
		}
	}
}

// queryCmdTestCase is a test case of a query command.
type queryCmdTestCase struct {
	// name is a name for this test case.
	name string
	// cmdGen is a generator for the command to execute. If not set, CmdQuery is used.
	cmdGen func() *cobra.Command
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
	var cmd *cobra.Command
	if tc.cmdGen != nil {
		cmd = tc.cmdGen()
	} else {
		cmd = cli.CmdQuery()
	}

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
func (s *CmdTestSuite) assertGetOrder(orderID string, order *exchange.Order) bool {
	s.T().Helper()
	if !s.Assert().NotEmptyf(orderID, "order id") {
		return false
	}

	var expInErr []string
	if order == nil {
		expInErr = append(expInErr, fmt.Sprintf("order %s not found", orderID))
	}

	clientCtx := s.getClientCtx()
	getOrderCmd := cli.CmdQueryGetOrder()
	getOrderOutBW, err := clitestutil.ExecTestCLICmd(clientCtx, getOrderCmd, []string{orderID, "--output", "json"})
	getOrderOutBz := getOrderOutBW.Bytes()
	s.T().Logf("Query GetOrder %s output:\n%s", orderID, string(getOrderOutBz))
	if !s.assertErrorContents(err, expInErr, "CmdQueryGetOrder %s error", orderID) {
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

// getOrderFollowup returns a followup function that identifies the new order id, looks it up,
// and makes sure it is as expected.
func (s *CmdTestSuite) getOrderFollowup(order *exchange.Order) func(resp *sdk.TxResponse) {
	return func(resp *sdk.TxResponse) {
		orderID, err := s.findNewOrderID(resp)
		if s.Assert().NoError(err, "finding new order id") {
			s.assertGetOrder(orderID, order)
		}
	}
}

// assertOrder uses the GetOrder query to look up an order and make sure it equals the one provided.
// If the provided order is nil, ensures the query returns an order not found error.
func (s *CmdTestSuite) assertGetOrderByExternalID(marketID, externalID string, order *exchange.Order) bool {
	s.T().Helper()
	if !s.Assert().NotEmpty(externalID, "external id") {
		return false
	}

	var expInErr []string
	if order == nil {
		expInErr = append(expInErr, fmt.Sprintf("order not found in market %s with external id %q", marketID, externalID))
	}

	clientCtx := s.getClientCtx()
	getOrderCmd := cli.CmdQueryGetOrderByExternalID()
	args := []string{"--market", marketID, "--external-id", externalID, "--output", "json"}
	getOrderOutBW, err := clitestutil.ExecTestCLICmd(clientCtx, getOrderCmd, args)
	getOrderOutBz := getOrderOutBW.Bytes()
	s.T().Logf("Query GetOrder %s %q output:\n%s", marketID, externalID, string(getOrderOutBz))
	if !s.assertErrorContents(err, expInErr, "CmdQueryGetOrderByExternalID %s %q error", marketID, externalID) {
		return false
	}

	if order == nil {
		return true
	}

	var resp exchange.QueryGetOrderByExternalIDResponse
	err = clientCtx.Codec.UnmarshalJSON(getOrderOutBz, &resp)
	if !s.Assert().NoError(err, "UnmarshalJSON on CmdQueryGetOrderByExternalID %s %q response", marketID, externalID) {
		return false
	}
	return s.Assert().Equal(order, resp.Order, "order in %s with %q", marketID, externalID)
}

// getOrderByExternalIDFollowup returns a followup function that looks up an order by external id.
func (s *CmdTestSuite) getOrderByExternalIDFollowup(marketID, externalID string, order *exchange.Order) func(resp *sdk.TxResponse) {
	return func(_ *sdk.TxResponse) {
		s.assertGetOrderByExternalID(marketID, externalID, order)
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
func (s *CmdTestSuite) govPropFollowup(msg sdk.Msg) func(resp *sdk.TxResponse) {
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
