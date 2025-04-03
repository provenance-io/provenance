package cli_test

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	abci "github.com/cometbft/cometbft/abci/types"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	testnet "github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/gogoproto/proto"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil"
	"github.com/provenance-io/provenance/testutil/assertions"
	testcli "github.com/provenance-io/provenance/testutil/cli"
	"github.com/provenance-io/provenance/testutil/queries"
	"github.com/provenance-io/provenance/x/exchange"
	"github.com/provenance-io/provenance/x/exchange/client/cli"
	"github.com/provenance-io/provenance/x/hold"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

type CmdTestSuite struct {
	suite.Suite

	cfg      testnet.Config
	testnet  *testnet.Network
	feeDenom string

	keyring        keyring.Keyring
	keyringEntries []testutil.TestKeyringEntry
	accountAddrs   []sdk.AccAddress

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

	addrNameLookup map[string]string
}

func TestCmdTestSuite(t *testing.T) {
	suite.Run(t, new(CmdTestSuite))
}

func (s *CmdTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")
	pioconfig.SetProvenanceConfig("", 0)
	govv1.DefaultMinDepositRatio = sdkmath.LegacyZeroDec()
	s.cfg = testutil.DefaultTestNetworkConfig()
	s.cfg.NumValidators = 1
	s.cfg.ChainID = antewrapper.SimAppChainID
	s.cfg.TimeoutCommit = 500 * time.Millisecond
	s.feeDenom = pioconfig.GetProvenanceConfig().FeeDenom

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
		s.addr0.String():           "addr0",
		s.addr1.String():           "addr1",
		s.addr2.String():           "addr2",
		s.addr3.String():           "addr3",
		s.addr4.String():           "addr4",
		s.addr5.String():           "addr5",
		s.addr6.String():           "addr6",
		s.addr7.String():           "addr7",
		s.addr8.String():           "addr8",
		s.addr9.String():           "addr9",
		cli.AuthorityAddr.String(): "authorityAddr",
	}

	var allMarkers []*markertypes.MarkerAccount
	newMarker := func(denom string) *markertypes.MarkerAccount {
		rv := &markertypes.MarkerAccount{
			BaseAccount: &authtypes.BaseAccount{
				Address: markertypes.MustGetMarkerAddress(denom).String(),
			},
			Status:                 markertypes.StatusActive,
			Denom:                  "cherry",
			Supply:                 sdkmath.NewInt(0),
			MarkerType:             markertypes.MarkerType_Coin,
			SupplyFixed:            false,
			AllowGovernanceControl: true,
		}
		allMarkers = append(allMarkers, rv)
		return rv
	}
	appleMarker := newMarker("apple")
	acornMarker := newMarker("acorn")
	peachMarker := newMarker("peach")
	cherryMarker := newMarker("cherry")
	newMarker("strawberry")
	newMarker("tangerine")

	// Add accounts to auth gen state.
	testutil.MutateGenesisState(s.T(), &s.cfg, authtypes.ModuleName, &authtypes.GenesisState{}, func(authGen *authtypes.GenesisState) *authtypes.GenesisState {
		genAccs := make(authtypes.GenesisAccounts, 0, len(s.accountAddrs)+len(allMarkers))
		for _, addr := range s.accountAddrs {
			genAccs = append(genAccs, authtypes.NewBaseAccount(addr, nil, 0, 1))
		}
		for _, marker := range allMarkers {
			genAccs = append(genAccs, marker)
		}
		newAccounts, err := authtypes.PackAccounts(genAccs)
		s.Require().NoError(err, "PackAccounts")
		authGen.Accounts = append(authGen.Accounts, newAccounts...)
		return authGen
	})

	// Add some markets to the exchange gen state.
	toHold := make(map[string]sdk.Coins)
	testutil.MutateGenesisState(s.T(), &s.cfg, exchange.ModuleName, &exchange.GenesisState{}, func(exchangeGen *exchange.GenesisState) *exchange.GenesisState {
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
				AcceptingCommitments:     true,
				CommitmentSettlementBips: 50,
				IntermediaryDenom:        "cherry",
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
				AcceptingCommitments:     true,
				CommitmentSettlementBips: 50,
				IntermediaryDenom:        "cherry",
				FeeCreateCommitmentFlat:  []sdk.Coin{sdk.NewInt64Coin("peach", 15)},
			},
			// Do not make a market 419, lots of tests expect it to not exist.
			exchange.Market{
				// The stuff in this market is for the query tests.
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

				AcceptingCommitments:     true,
				FeeCreateCommitmentFlat:  []sdk.Coin{sdk.NewInt64Coin("peach", 5)},
				CommitmentSettlementBips: 50,
				IntermediaryDenom:        "cherry",
				ReqAttrCreateCommitment:  []string{"committer.kyc"},
			},
			exchange.Market{
				// This market has an invalid setup. Don't use it for anything else.
				MarketId:      421,
				MarketDetails: exchange.MarketDetails{Name: "Broken"},
				FeeSellerSettlementRatios: []exchange.FeeRatio{
					{Price: sdk.NewInt64Coin("peach", 55), Fee: sdk.NewInt64Coin("peach", 1)},
				},
				FeeBuyerSettlementRatios: []exchange.FeeRatio{
					{Price: sdk.NewInt64Coin("peach", 56), Fee: sdk.NewInt64Coin("peach", 1)},
					{Price: sdk.NewInt64Coin("plum", 57), Fee: sdk.NewInt64Coin("plum", 1)},
				},
				AccessGrants: []exchange.AccessGrant{
					{Address: s.addr1.String(), Permissions: exchange.AllPermissions()},
				},
			},
		)

		exchangeGen.Orders = make([]exchange.Order, 60)
		for i := range exchangeGen.Orders {
			order := s.makeInitialOrder(uint64(i + 1))
			exchangeGen.Orders[i] = *order
		}
		exchangeGen.LastOrderId = uint64(100)

		for i := range s.accountAddrs {
			com := s.makeInitialCommitment(i)
			if !com.Amount.IsZero() {
				exchangeGen.Commitments = append(exchangeGen.Commitments, *com)
			}
		}

		exchangeGen.Commitments = append(exchangeGen.Commitments, exchange.Commitment{
			Account:  s.addr1.String(),
			MarketId: 421,
			Amount:   sdk.NewCoins(sdk.NewInt64Coin("apple", 4210), sdk.NewInt64Coin("peach", 421)),
		})

		for sourceI := range s.accountAddrs {
			for targetI := range s.accountAddrs {
				payment := s.makeInitialPayment(sourceI, targetI)
				exchangeGen.Payments = append(exchangeGen.Payments, *payment)
			}
			payment := s.makeInitialPayment(sourceI, len(s.accountAddrs))
			exchangeGen.Payments = append(exchangeGen.Payments, *payment)
		}

		for _, order := range exchangeGen.Orders {
			toHold[order.GetOwner()] = toHold[order.GetOwner()].Add(order.GetHoldAmount()...)
		}

		for _, com := range exchangeGen.Commitments {
			toHold[com.Account] = toHold[com.Account].Add(com.Amount...)
		}

		for _, payment := range exchangeGen.Payments {
			if !payment.SourceAmount.IsZero() {
				toHold[payment.Source] = toHold[payment.Source].Add(payment.SourceAmount...)
			}
		}

		return exchangeGen
	})

	// Create all the needed holds.
	testutil.MutateGenesisState(s.T(), &s.cfg, hold.ModuleName, &hold.GenesisState{}, func(holdGen *hold.GenesisState) *hold.GenesisState {
		for _, addr := range s.accountAddrs {
			holdGen.Holds = append(holdGen.Holds, &hold.AccountHold{
				Address: addr.String(),
				Amount:  toHold[addr.String()],
			})
		}
		return holdGen
	})

	// Add the markers.
	testutil.MutateGenesisState(s.T(), &s.cfg, markertypes.ModuleName, &markertypes.GenesisState{}, func(markerGen *markertypes.GenesisState) *markertypes.GenesisState {
		markerGen.NetAssetValues = append(markerGen.NetAssetValues, []markertypes.MarkerNetAssetValues{
			{
				Address:        cherryMarker.Address,
				NetAssetValues: []markertypes.NetAssetValue{{Price: s.feeCoin(100), Volume: 1}},
			},
			{
				Address:        appleMarker.Address,
				NetAssetValues: []markertypes.NetAssetValue{{Price: sdk.NewInt64Coin("cherry", 8), Volume: 1}},
			},
			{
				Address:        acornMarker.Address,
				NetAssetValues: []markertypes.NetAssetValue{{Price: sdk.NewInt64Coin("cherry", 3), Volume: 17}},
			},
			{
				Address:        peachMarker.Address,
				NetAssetValues: []markertypes.NetAssetValue{{Price: sdk.NewInt64Coin("cherry", 778), Volume: 3}},
			},
		}...)
		return markerGen
	})

	// Add balances to bank gen state.
	testutil.MutateGenesisState(s.T(), &s.cfg, banktypes.ModuleName, &banktypes.GenesisState{}, func(bankGen *banktypes.GenesisState) *banktypes.GenesisState {
		// Any initial holds for an account are added to this so that
		// this is what's spendable to each at the start of the unit tests.
		balance := sdk.NewCoins(
			s.bondCoin(1_000_000_000),
			s.feeCoin(1_000_000_000_000),
			sdk.NewInt64Coin("acorn", 1_000_000_000),
			sdk.NewInt64Coin("apple", 1_000_000_000),
			sdk.NewInt64Coin("peach", 1_000_000_000),
			sdk.NewInt64Coin("strawberry", 1_000_000_000),
			sdk.NewInt64Coin("tangerine", 1_000_000_000),
		)
		for _, addr := range s.accountAddrs {
			bal := balance.Add(toHold[addr.String()]...)
			bankGen.Balances = append(bankGen.Balances, banktypes.Balance{Address: addr.String(), Coins: bal})
		}
		return bankGen
	})

	// And fire it all up!!
	var err error
	s.testnet, err = testnet.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err, "testnet.New(...)")

	s.testnet.Validators[0].ClientCtx = s.testnet.Validators[0].ClientCtx.WithKeyring(s.keyring)

	_, err = testutil.WaitForHeight(s.testnet, 1)
	s.Require().NoError(err, "s.testnet.WaitForHeight(1)")
}

func (s *CmdTestSuite) TearDownSuite() {
	testutil.Cleanup(s.testnet, s.T())
}

// generateAccountsWithKeyring creates a keyring and adds a number of keys to it.
// The s.keyringDir, s.keyring, and s.accountAddrs are all set in here.
// The getClientCtx function returns a context that knows about this keyring.
func (s *CmdTestSuite) generateAccountsWithKeyring(number int) {
	s.keyringEntries, s.keyring = testutil.GenerateTestKeyring(s.T(), number, s.cfg.Codec)
	s.accountAddrs = testutil.GetKeyringEntryAddresses(s.keyringEntries)
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
	externalID := fmt.Sprintf("my-id-%d", orderID)
	order := exchange.NewOrder(orderID)
	switch orderID % 6 {
	case 0, 1, 4:
		order.WithAsk(&exchange.AskOrder{
			MarketId:     420,
			Seller:       addr.String(),
			Assets:       assets,
			Price:        price,
			ExternalId:   externalID,
			AllowPartial: partial,
		})
	case 2, 3, 5:
		order.WithBid(&exchange.BidOrder{
			MarketId:     420,
			Buyer:        addr.String(),
			Assets:       assets,
			Price:        price,
			ExternalId:   externalID,
			AllowPartial: partial,
		})
	}
	return order
}

// makeInitialCommitment makes a commitment for the s.accountAddrs with the provided addrI.
// The amount is based off the addrI and might be zero.
func (s *CmdTestSuite) makeInitialCommitment(addrI int) *exchange.Commitment {
	var addr sdk.AccAddress
	s.Require().NotPanics(func() {
		addr = s.accountAddrs[addrI]
	}, "s.accountAddrs[%d]", addrI)
	rv := &exchange.Commitment{
		Account:  addr.String(),
		MarketId: 420,
	}

	i := addrI % 10 // in case the number of addresses changes later.
	// One denom  (3): 0 = apple, 1 = acorn, 2 = peach
	// Two denoms (3): 3 = apple+acorn, 4 = apple+peach, 5 = acorn+peach
	// All denoms (2): 6, 7
	// Nothing    (2): 8, 9
	// apple <= 0, 3, 4, 6, 7
	// acorn <= 1, 3, 5, 6, 7
	// peach <= 2, 4, 5, 6, 7
	if contains([]int{0, 3, 4, 6, 7}, i) {
		rv.Amount = rv.Amount.Add(sdk.NewInt64Coin("apple", int64(1000+addrI*200)))
	}
	if contains([]int{1, 3, 5, 6, 7}, i) {
		rv.Amount = rv.Amount.Add(sdk.NewInt64Coin("acorn", int64(10_000+addrI*100)))
	}
	if contains([]int{2, 4, 5, 6, 7}, i) {
		rv.Amount = rv.Amount.Add(sdk.NewInt64Coin("peach", int64(500+addrI*addrI*100)))
	}
	return rv
}

// makeInitialPayment makes a payment with the source and target having the s.accountAddrs with the given indexes.
// If sourceI or targetI is not in s.accountAddrs, the payment won't have a source or target (respectively).
// The amounts are based off of the sourceI and targetI, and might each be zero (but not both).
func (s *CmdTestSuite) makeInitialPayment(sourceI, targetI int) *exchange.Payment {
	rv := &exchange.Payment{
		ExternalId: fmt.Sprintf("initial-payment-%02d-%02d", sourceI, targetI),
	}
	if sourceI < len(s.accountAddrs) {
		rv.Source = s.accountAddrs[sourceI].String()
	}
	if targetI < len(s.accountAddrs) {
		rv.Target = s.accountAddrs[targetI].String()
	}

	switch sourceI % 3 {
	case 1:
		rv.SourceAmount = sdk.NewCoins(sdk.NewInt64Coin("strawberry", int64(300+50*targetI)))
	case 2:
		rv.SourceAmount = sdk.NewCoins(sdk.NewInt64Coin("strawberry", int64(500+100*targetI)),
			sdk.NewInt64Coin("peach", int64(10+50*targetI*targetI)))
	}

	switch targetI % 3 {
	case 1:
		rv.TargetAmount = sdk.NewCoins(sdk.NewInt64Coin("tangerine", int64(100+200*sourceI)))
	case 2:
		rv.TargetAmount = sdk.NewCoins(sdk.NewInt64Coin("tangerine", int64(1000+50*sourceI)),
			sdk.NewInt64Coin("acorn", int64(10_000+10*sourceI*sourceI)))
	}

	// Don't allow both the amounts to be zero
	if rv.SourceAmount.IsZero() && rv.TargetAmount.IsZero() {
		rv.SourceAmount = sdk.NewCoins(sdk.NewInt64Coin("peach", int64(targetI+1)))
		rv.TargetAmount = sdk.NewCoins(sdk.NewInt64Coin("acorn", int64(sourceI+1)))
	}

	return rv
}

// contains reports whether v is present in s.
func contains[S ~[]E, E comparable](s S, v E) bool {
	for i := range s {
		if v == s[i] {
			return true
		}
	}
	return false
}

// getClientCtx get a client context that knows about the suite's keyring.
func (s *CmdTestSuite) getClientCtx() client.Context {
	return s.testnet.Validators[0].ClientCtx
}

// getAddrName tries to get the variable name (in this suite) of the provided address.
func (s *CmdTestSuite) getAddrName(addr string) string {
	if len(addr) == 0 {
		return "<empty>"
	}
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
	// run any follow-up checks to do after the command is run.
	preRun func() ([]string, func(*sdk.TxResponse))
	// args are the arguments to provide to the command.
	args []string
	// addedFees is any fees to add to the default 10<bond> amount.
	addedFees sdk.Coins
	// gas is the amount of gas to include. Default is 250,000.
	gas int
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

	fees := s.bondCoins(10)
	if !tc.addedFees.IsZero() {
		fees = fees.Add(tc.addedFees...)
	}

	gas := "300000"
	if tc.gas > 0 {
		gas = fmt.Sprintf("%d", tc.gas)
	}

	args := append(tc.args, extraArgs...)
	args = append(args,
		"--"+flags.FlagGas, gas,
		"--"+flags.FlagFees, fees.String(),
		"--"+flags.FlagBroadcastMode, flags.BroadcastSync,
		"--"+flags.FlagSkipConfirmation,
	)

	var txResponse *sdk.TxResponse
	var cmdFailed bool
	testRunner := func() {
		if preRunFailed {
			s.T().Skip("Skipping execution due to pre-run failure.")
		}

		var cmdOk bool
		txResponse, cmdOk = testcli.NewTxExecutor(cmd, args).
			WithExpInErrMsg(tc.expInErr).
			WithExpCode(tc.expectedCode).
			WithExpInRawLog(tc.expInRawLog).
			AssertExecute(s.T(), s.testnet)
		cmdFailed = !cmdOk
	}

	if tc.preRun != nil || followup != nil {
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
				s.T().Skip("Skipping followup due to execute failure.")
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
		if !s.Assert().Contains(outStr, exp, "command output") {
			s.T().Logf("Not found: %q", exp)
		}
	}

	if len(tc.expOut) > 0 {
		s.Assert().Equal(tc.expOut, outStr, "command output string")
	}
}

func (s *CmdTestSuite) composeFollowups(followups ...func(*sdk.TxResponse)) func(*sdk.TxResponse) {
	return func(resp *sdk.TxResponse) {
		for _, followup := range followups {
			followup(resp)
		}
	}
}

// getEventAttribute finds the value of an attribute in an event.
// Returns an error if the value is empty, the attribute doesn't exist, or the event doesn't exist.
func (s *CmdTestSuite) getEventAttribute(events []abci.Event, eventType, attribute string) (string, error) {
	for _, event := range events {
		if event.Type == eventType {
			for _, attr := range event.Attributes {
				if attr.Key == attribute {
					val := strings.Trim(attr.Value, `"`)
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

// assertEventTagFollowup asserts that the "tag" attribute of an event with the given name is
func (s *CmdTestSuite) assertEventsContains(expected sdk.Events) func(*sdk.TxResponse) {
	return func(resp *sdk.TxResponse) {
		actual := s.abciEventsToSDKEvents(resp.Events)
		assertions.AssertEventsContains(s.T(), expected, actual, "TxResponse.Events")
	}
}

// abciEventsToSDKEvents converts the provided events into sdk.Events.
func (s *CmdTestSuite) abciEventsToSDKEvents(fromResp []abci.Event) sdk.Events {
	if fromResp == nil {
		return nil
	}
	events := make(sdk.Events, len(fromResp))
	for i, v := range fromResp {
		events[i] = sdk.Event(v)
	}
	return events
}

// markAttrsIndexed sets Index = true on all attributes in the provided events.
func (s *CmdTestSuite) markAttrsIndexed(events sdk.Events) {
	for e, event := range events {
		for a := range event.Attributes {
			events[e].Attributes[a].Index = true
		}
	}
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

// assertGetPayment uses the GetPayment query to look up a payment and make sure it equals the one provided.
// If the provided payment is nil, ensures the query returns a payment not found error.
func (s *CmdTestSuite) assertGetPayment(source, externalID string, payment *exchange.Payment) (okay bool) {
	s.T().Helper()
	if !s.Assert().NotEmpty(source, "source") {
		return false
	}

	var expInErr []string
	if payment == nil {
		expInErr = append(expInErr, fmt.Sprintf("no payment found with source %s and external id %q", source, externalID))
	}

	var getPaymentOutBz []byte
	getPaymentArgs := []string{source, externalID, "--output", "json"}
	defer func() {
		if !okay {
			s.T().Logf("Query GetPayment %q output:\n%s", getPaymentArgs, string(getPaymentOutBz))
		}
	}()

	clientCtx := s.getClientCtx()
	getPaymentCmd := cli.CmdQueryGetPayment()
	getPaymentOutBW, err := clitestutil.ExecTestCLICmd(clientCtx, getPaymentCmd, getPaymentArgs)
	getPaymentOutBz = getPaymentOutBW.Bytes()
	if !s.assertErrorContents(err, expInErr, "ExecTestCLICmd GetPayment %q error", getPaymentArgs) {
		return false
	}

	if payment == nil {
		return true
	}

	var resp exchange.QueryGetPaymentResponse
	err = clientCtx.Codec.UnmarshalJSON(getPaymentOutBz, &resp)
	if !s.Assert().NoError(err, "UnmarshalJSON on GetPayment %q response", getPaymentArgs) {
		return false
	}
	return s.Assert().Equal(payment, resp.Payment, "payment %s %q", source, externalID)
}

// getOrderFollowup returns a follow-up function that looks up an order and makes sure it's the one provided.
func (s *CmdTestSuite) getPaymentFollowup(source, externalID string, payment *exchange.Payment) func(*sdk.TxResponse) {
	return func(*sdk.TxResponse) {
		s.assertGetPayment(source, externalID, payment)
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

	prop := queries.GetGovProp(s.T(), s.testnet, propID)
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

// bondCoin returns a Coin with the bond denom and the provided amount.
func (s *CmdTestSuite) bondCoin(amt int64) sdk.Coin {
	return sdk.NewInt64Coin(s.cfg.BondDenom, amt)
}

// bondCoins returns a Coins with just an entry with the bond denom and the provided amount.
func (s *CmdTestSuite) bondCoins(amt int64) sdk.Coins {
	return sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, amt))
}

// feeCoin returns a Coin with the fee denom and the provided amount.
func (s *CmdTestSuite) feeCoin(amt int64) sdk.Coin {
	return sdk.NewInt64Coin(s.feeDenom, amt)
}

// feeCoins returns a Coins with just an entry with the fee denom and the provided amount.
func (s *CmdTestSuite) feeCoins(amt int64) sdk.Coins {
	return sdk.NewCoins(sdk.NewInt64Coin(s.feeDenom, amt))
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

// assertBalancesFollowup returns a follow-up function that asserts that the spendable balances are now as expected.
func (s *CmdTestSuite) assertSpendableBalancesFollowup(expSpendable []banktypes.Balance) func(*sdk.TxResponse) {
	return func(_ *sdk.TxResponse) {
		for _, expBal := range expSpendable {
			actBal := s.queryBankSpendableBalances(expBal.Address)
			s.Assert().Equal(expBal.Coins.String(), actBal.String(), "%s spendable balances", s.getAddrName(expBal.Address))
		}
	}
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
		"--"+flags.FlagBroadcastMode, flags.BroadcastSync,
		"--"+flags.FlagSkipConfirmation,
	)

	resp := testcli.NewTxExecutor(cmd, args).Execute(s.T(), s.testnet)
	s.Require().NotNil(resp, "TxResponse from creating order")
	orderIDStr, err := s.findNewOrderID(resp)
	s.Require().NoError(err, "findNewOrderID")

	return s.asOrderID(orderIDStr)
}

// commitFunds issues a command to commit funds.
func (s *CmdTestSuite) commitFunds(addr sdk.AccAddress, marketID uint32, amount sdk.Coins, creationFee sdk.Coins) {
	cmd := cli.CmdTx()
	args := []string{
		"commit",
		"--from", addr.String(),
		"--market", fmt.Sprintf("%d", marketID),
		"--amount", amount.String(),
	}
	if !creationFee.IsZero() {
		args = append(args, "--creation-fee", creationFee.String())
	}

	args = append(args,
		"--"+flags.FlagFees, s.bondCoins(10).String(),
		"--"+flags.FlagBroadcastMode, flags.BroadcastSync,
		"--"+flags.FlagSkipConfirmation,
	)

	testcli.NewTxExecutor(cmd, args).Execute(s.T(), s.testnet)
}

// createPayment issues a command to create a payment.
func (s *CmdTestSuite) createPayment(payment *exchange.Payment) {
	cmd := cli.CmdTx()
	args := []string{
		"create-payment", "--from", payment.Source,
	}
	if !payment.SourceAmount.IsZero() {
		args = append(args, "--source-amount", payment.SourceAmount.String())
	}
	if len(payment.Target) > 0 {
		args = append(args, "--target", payment.Target)
	}
	if !payment.TargetAmount.IsZero() {
		args = append(args, "--target-amount", payment.TargetAmount.String())
	}
	if len(payment.ExternalId) > 0 {
		args = append(args, "--external-id", payment.ExternalId)
	}

	fees := s.bondCoins(10).Add(s.feeCoin(exchange.DefaultFeeCreatePaymentFlatAmount))
	args = append(args,
		"--"+flags.FlagFees, fees.String(),
		"--"+flags.FlagBroadcastMode, flags.BroadcastSync,
		"--"+flags.FlagSkipConfirmation,
	)

	testcli.NewTxExecutor(cmd, args).Execute(s.T(), s.testnet)
}

// queryBankBalances executes a bank query to get an account's balances.
func (s *CmdTestSuite) queryBankBalances(addr string) sdk.Coins {
	return queries.GetAllBalances(s.T(), s.testnet, addr)
}

// queryBankSpendableBalances executes a bank query to get an account's spendable balances.
func (s *CmdTestSuite) queryBankSpendableBalances(addr string) sdk.Coins {
	return queries.GetSpendableBalances(s.T(), s.testnet, addr)
}

// execBankSend executes a bank send command.
func (s *CmdTestSuite) execBankSend(fromAddr, toAddr, amount string) {
	addrCdc := s.cfg.Codec.InterfaceRegistry().SigningContext().AddressCodec()
	cmd := bankcli.NewSendTxCmd(addrCdc)
	args := []string{
		fromAddr, toAddr, amount,
		"--" + flags.FlagFees, s.bondCoins(10).String(),
		"--" + flags.FlagBroadcastMode, flags.BroadcastSync,
		"--" + flags.FlagSkipConfirmation,
	}
	testcli.NewTxExecutor(cmd, args).Execute(s.T(), s.testnet)
}

// untypeEvent calls untypeEvent and requires it to not return an error.
func (s *CmdTestSuite) untypeEvent(tev proto.Message) sdk.Event {
	rv, err := sdk.TypedEventToEvent(tev)
	s.Require().NoError(err, "TypedEventToEvent(%T)", tev)
	rv.Attributes = append(rv.Attributes, abci.EventAttribute{Key: "msg_index", Value: "0", Index: true})
	return rv
}

// joinErrs joins the provided error strings matching to how errors.Join does.
func joinErrs(errs ...string) string {
	return strings.Join(errs, "\n")
}

// toStringSlice applies the stringer to each value and returns a slice with the results.
//
// T is the type of things being converted to strings.
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
//
// T is the type of things being compared (and possibly converted to strings).
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
	// Annotations are only checked on the flags listed in expFlags.
	expAnnotations map[string]map[string][]string
	// expInUse is a set of strings that are expected to be in the command's Use string.
	// Each entry that does not start with a "[" is also checked to not be in the Use wrapped in [].
	expInUse []string
	// expExamples is a set of examples to ensure are on the command.
	// There must be a full line in the command's Example that matches each entry.
	expExamples []string
	// skipArgsCheck true causes the runner to skip the check ensuring that the command's Args func has been set.
	skipArgsCheck bool
	// skipAddingFromFlag true causes the runner to not add the from flag to the dummy command.
	skipAddingFromFlag bool
	// skipFlagInUseCheck true causes the runner to skip checking that each entry in expFlags also appears in the cmd.Use.
	skipFlagInUseCheck bool
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
	if !tc.skipAddingFromFlag {
		cmd.Flags().String(flags.FlagFrom, "", "The from flag")
	}

	testFunc := func() {
		tc.setup(cmd)
	}
	require.NotPanics(t, testFunc, tc.name)

	pageFlags := []string{
		flags.FlagPage, flags.FlagPageKey, flags.FlagOffset,
		flags.FlagLimit, flags.FlagCountTotal, flags.FlagReverse,
	}
	for i, flagName := range tc.expFlags {
		t.Run(fmt.Sprintf("flag[%d]: --%s", i, flagName), func(t *testing.T) {
			flag := cmd.Flags().Lookup(flagName)
			if assert.NotNil(t, flag, "--%s", flagName) {
				expAnnotations, _ := tc.expAnnotations[flagName]
				actAnnotations := flag.Annotations
				assert.Equal(t, expAnnotations, actAnnotations, "--%s annotations", flagName)
				if !tc.skipFlagInUseCheck {
					expInUse := "--" + flagName
					if exchange.ContainsString(pageFlags, flagName) {
						expInUse = cli.PageFlagsUse
					}
					assert.Contains(t, cmd.Use, expInUse, "cmd.Use should have something about the %s flag", flagName)
				}
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

// addOneReqAnnotations adds the expected annotations to a setupTestCase that indicate that one of a set of flags is required.
// The flags should be provided in the same order that they're provided to cmd.MarkFlagsOneRequired.
func addOneReqAnnotations(tc *setupTestCase, oneReqFlags ...string) {
	oneReqVal := strings.Join(oneReqFlags, " ")
	if tc.expAnnotations == nil {
		tc.expAnnotations = make(map[string]map[string][]string)
	}
	for _, name := range oneReqFlags {
		if tc.expAnnotations[name] == nil {
			tc.expAnnotations[name] = make(map[string][]string)
		}
		tc.expAnnotations[name][oneReq] = append(tc.expAnnotations[name][oneReq], oneReqVal)
	}
}

// encodingConfig is an encoding config that can be used by these tests.
// Do not use this variable directly. Instead, use the getEncodingConfig function.
var encodingConfig *params.EncodingConfig

// getEncodingConfig gets the encoding config, creating it if it hasn't been made yet.
func getEncodingConfig(t *testing.T) params.EncodingConfig {
	t.Helper()
	if encodingConfig == nil {
		encCfg := app.MakeTestEncodingConfig(t)
		encodingConfig = &encCfg
	}
	return *encodingConfig
}

// newClientContext returns a new client.Context that has a codec and keyring.
func newClientContext(t *testing.T) client.Context {
	ctx := client.Context{}
	ctx = clientContextWithCodec(t, ctx)
	return clientContextWithKeyring(t, ctx)
}

// clientContextWithCodec adds a useful Codec to the provided client context.
func clientContextWithCodec(t *testing.T, clientCtx client.Context) client.Context {
	encCfg := getEncodingConfig(t)
	return clientCtx.
		WithCodec(encCfg.Marshaler).
		WithInterfaceRegistry(encCfg.InterfaceRegistry).
		WithTxConfig(encCfg.TxConfig)
}

const (
	keyringName     = "keyringaddr"
	keyringMnemonic = "pave kit dust loop forest symptom lobster tape note attitude aim cloth cat welcome basic head wet pistol hair funny library dove oak drift"
	keyringAddr     = "cosmos177zjzs7mh79j0cx6wcxq8dayetkftkx2crt4u6"
)

// clientContextWithKeyring returns a client context that has a keyring and an entry for keyringName=>keyringAddr.
func clientContextWithKeyring(t *testing.T, clientCtx client.Context) client.Context {
	kr, err := client.NewKeyringFromBackend(clientCtx, keyring.BackendMemory)
	require.NoError(t, err, "NewKeyringFromBackend")
	clientCtx = clientCtx.WithKeyring(kr)

	// This mnemonic was generated by using Keyring.NewAccount().
	// I then logged it (and the address) with t.Logf(...) and copy/pasted it into the constants above.
	_, err = clientCtx.Keyring.NewAccount(keyringName, keyringMnemonic, keyring.DefaultBIP39Passphrase, hd.CreateHDPath(118, 0, 0).String(), hd.Secp256k1)
	require.NoError(t, err, "adding %s to the keyring using NewAccount(...)", keyringName)

	return clientCtx
}

// newGovProp creates a new MsgSubmitProposal containing the provided messages, requiring it to not error.
func newGovProp(t *testing.T, msgs ...sdk.Msg) *govv1.MsgSubmitProposal {
	rv := &govv1.MsgSubmitProposal{}
	for _, msg := range msgs {
		msgAny, err := codectypes.NewAnyWithValue(msg)
		require.NoError(t, err, "NewAnyWithValue(%T)", msg)
		rv.Messages = append(rv.Messages, msgAny)
	}
	return rv
}

// newTx creates a new Tx containing the provided messages, requiring it to not error.
func newTx(t *testing.T, msgs ...sdk.Msg) *txtypes.Tx {
	rv := &txtypes.Tx{
		Body:       &txtypes.TxBody{},
		AuthInfo:   &txtypes.AuthInfo{},
		Signatures: make([][]byte, 0),
	}
	for _, msg := range msgs {
		msgAny, err := codectypes.NewAnyWithValue(msg)
		require.NoError(t, err, "NewAnyWithValue(%T)", msg)
		rv.Body.Messages = append(rv.Body.Messages, msgAny)
	}
	return rv
}

// writeFileAsJson writes the provided proto message as a json file, requiring it to not error.
func writeFileAsJson(t *testing.T, filename string, content proto.Message) {
	clientCtx := newClientContext(t)
	bz, err := clientCtx.Codec.MarshalJSON(content)
	require.NoError(t, err, "MarshalJSON(%T)", content)
	writeFile(t, filename, bz)
}

// writeFile writes a file requiring it to not error.
func writeFile(t *testing.T, filename string, bz []byte) {
	err := os.WriteFile(filename, bz, 0o644)
	require.NoError(t, err, "WriteFile(%q)", filename)
}

// getAnyTypes gets the TypeURL field of each of the provided anys.
func getAnyTypes(anys []*codectypes.Any) []string {
	rv := make([]string, len(anys))
	for i, a := range anys {
		rv[i] = a.GetTypeUrl()
	}
	return rv
}
