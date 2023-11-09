package cli_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	testnet "github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil"
	"github.com/provenance-io/provenance/x/exchange"
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

	// Add balances to bank gen state.
	balance := sdk.NewCoins(
		sdk.NewInt64Coin(s.cfg.BondDenom, 1_000_000_000),
		sdk.NewInt64Coin("apple", 1_000_000_000),
		sdk.NewInt64Coin("peach", 1_000_000_000),
	)
	var bankGen banktypes.GenesisState
	err = s.cfg.Codec.UnmarshalJSON(s.cfg.GenesisState[banktypes.ModuleName], &bankGen)
	s.Require().NoError(err, "UnmarshalJSON bank gen state")
	for _, addr := range s.accountAddrs {
		bankGen.Balances = append(bankGen.Balances, banktypes.Balance{Address: addr.String(), Coins: balance})
	}
	s.cfg.GenesisState[banktypes.ModuleName], err = s.cfg.Codec.MarshalJSON(&bankGen)
	s.Require().NoError(err, "MarshalJSON bank gen state")

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
	s.cfg.GenesisState[exchange.ModuleName], err = s.cfg.Codec.MarshalJSON(&exchangeGen)
	s.Require().NoError(err, "MarshalJSON exchange gen state")

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
// The GetClientCtx function returns a context that knows about this keyring.
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

// GetClientCtx get a client context that knows about the suite's keyring.
func (s *CmdTestSuite) GetClientCtx() client.Context {
	return s.testnet.Validators[0].ClientCtx.
		WithKeyringDir(s.keyringDir).
		WithKeyring(s.keyring)
}
