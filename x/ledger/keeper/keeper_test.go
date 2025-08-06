package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/math"
	"cosmossdk.io/x/nft"
	nftkeeper "cosmossdk.io/x/nft/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	registrykeeper "github.com/provenance-io/provenance/x/registry/keeper"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/ledger/helper"
	"github.com/provenance-io/provenance/x/ledger/keeper"
	ledger "github.com/provenance-io/provenance/x/ledger/types"
)

type TestSuite struct {
	suite.Suite

	app            *app.App
	ctx            sdk.Context
	keeper         keeper.Keeper
	bankKeeper     bankkeeper.Keeper
	nftKeeper      nftkeeper.Keeper
	registryKeeper registrykeeper.RegistryKeeper

	bondDenom  string
	initBal    sdk.Coins
	initAmount int64

	addr1 sdk.AccAddress
	addr2 sdk.AccAddress
	addr3 sdk.AccAddress

	pastDate int32

	validLedgerClass ledger.LedgerClass
	validNFTClass    nft.Class
	validNFT         nft.NFT
}

func (s *TestSuite) SetupTest() {
	s.app = app.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(false)
	s.keeper = s.app.LedgerKeeper
	s.bankKeeper = s.app.BankKeeper
	s.nftKeeper = s.app.NFTKeeper
	s.registryKeeper = s.app.RegistryKeeper

	var err error
	s.bondDenom, err = s.app.StakingKeeper.BondDenom(s.ctx)
	s.Require().NoError(err, "app.StakingKeeper.BondDenom(s.ctx)")

	s.initAmount = 1_000_000_000
	s.initBal = sdk.NewCoins(sdk.NewCoin(s.bondDenom, math.NewInt(s.initAmount)))

	addrs := app.AddTestAddrsIncremental(s.app, s.ctx, 3, math.NewInt(s.initAmount))
	s.addr1 = addrs[0]
	s.addr2 = addrs[1]
	s.addr3 = addrs[2]

	// Create a timestamp 24 hours in the past to avoid future date errors
	s.pastDate = helper.DaysSinceEpoch(time.Now().Add(-24 * time.Hour).UTC())

	// Load the test ledger class configs
	s.ConfigureTest()
}

func (s *TestSuite) ConfigureTest() {
	s.ctx = s.ctx.WithBlockTime(time.Now())

	s.validNFTClass = nft.Class{
		Id: "test-nft-class-id",
	}
	s.nftKeeper.SaveClass(s.ctx, s.validNFTClass)

	s.validNFT = nft.NFT{
		ClassId: s.validNFTClass.Id,
		Id:      "test-nft-id",
	}
	s.nftKeeper.Mint(s.ctx, s.validNFT, s.addr1)

	s.validLedgerClass = ledger.LedgerClass{
		LedgerClassId:     "test-ledger-class-id",
		AssetClassId:      s.validNFTClass.Id,
		MaintainerAddress: s.addr1.String(),
		Denom:             s.bondDenom,
	}
	s.keeper.AddLedgerClass(s.ctx, s.validLedgerClass)

	s.keeper.AddClassEntryType(s.ctx, s.validLedgerClass.LedgerClassId, ledger.LedgerClassEntryType{
		Id:          1,
		Code:        "SCHEDULED_PAYMENT",
		Description: "Scheduled Payment",
	})

	s.keeper.AddClassEntryType(s.ctx, s.validLedgerClass.LedgerClassId, ledger.LedgerClassEntryType{
		Id:          2,
		Code:        "DISBURSEMENT",
		Description: "Disbursement",
	})

	s.keeper.AddClassEntryType(s.ctx, s.validLedgerClass.LedgerClassId, ledger.LedgerClassEntryType{
		Id:          3,
		Code:        "ORIGINATION_FEE",
		Description: "Origination Fee",
	})

	s.keeper.AddClassBucketType(s.ctx, s.validLedgerClass.LedgerClassId, ledger.LedgerClassBucketType{
		Id:          1,
		Code:        "PRINCIPAL",
		Description: "Principal",
	})

	s.keeper.AddClassBucketType(s.ctx, s.validLedgerClass.LedgerClassId, ledger.LedgerClassBucketType{
		Id:          2,
		Code:        "INTEREST",
		Description: "Interest",
	})

	s.keeper.AddClassBucketType(s.ctx, s.validLedgerClass.LedgerClassId, ledger.LedgerClassBucketType{
		Id:          3,
		Code:        "ESCROW",
		Description: "Escrow",
	})

	s.keeper.AddClassStatusType(s.ctx, s.validLedgerClass.LedgerClassId, ledger.LedgerClassStatusType{
		Id:          1,
		Code:        "IN_REPAYMENT",
		Description: "In Repayment",
	})
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

// coins creates an sdk.Coins from a string, requiring it to work.
func (s *TestSuite) coins(coins string) sdk.Coins {
	s.T().Helper()
	rv, err := sdk.ParseCoinsNormalized(coins)
	s.Require().NoError(err, "ParseCoinsNormalized(%q)", coins)
	return rv
}

// coin creates a new coin without doing any validation on it.
func (s *TestSuite) coin(amount int64, denom string) sdk.Coin {
	return sdk.Coin{
		Amount: s.int(amount),
		Denom:  denom,
	}
}

// int is a shorter way to call math.NewInt.
func (s *TestSuite) int(amount int64) math.Int {
	return math.NewInt(amount)
}

// intStr creates a math.Int from a string, requiring it to work.
func (s *TestSuite) intStr(amount string) math.Int {
	s.T().Helper()
	rv, ok := math.NewIntFromString(amount)
	s.Require().True(ok, "NewIntFromString(%q) ok bool", amount)
	return rv
}

// assertErrorContents asserts that the provided error is as expected.
func (s *TestSuite) assertErrorContents(theError error, contains []string, msgAndArgs ...interface{}) bool {
	return assertions.AssertErrorContents(s.T(), theError, contains, msgAndArgs...)
}

// assertErrorValue asserts that the provided error equals the expected.
func (s *TestSuite) assertErrorValue(theError error, expected string, msgAndArgs ...interface{}) bool {
	return assertions.AssertErrorValue(s.T(), theError, expected, msgAndArgs...)
}

// requirePanicContents asserts that, if contains is empty, the provided func does not panic
func (s *TestSuite) requirePanicContents(f assertions.PanicTestFunc, contains []string, msgAndArgs ...interface{}) {
	assertions.RequirePanicContents(s.T(), f, contains, msgAndArgs...)
}

// getAddrName returns the name of the variable in this TestSuite holding the provided address.
func (s *TestSuite) getAddrName(addr string) string {
	switch addr {
	case s.addr1.String():
		return "addr1"
	case s.addr2.String():
		return "addr2"
	case s.addr3.String():
		return "addr3"
	default:
		return addr
	}
}

// fundAccount funds an account with the provided coins.
func (s *TestSuite) fundAccount(addr sdk.AccAddress, coins string) {
	s.T().Helper()
	assertions.RequireNotPanicsNoErrorf(s.T(), func() error {
		return testutil.FundAccount(s.ctx, s.app.BankKeeper, addr, s.coins(coins))
	}, "FundAccount(%s, %q)", s.getAddrName(addr.String()), coins)
}

// assertEqualEvents asserts that the expected events equal the actual events.
func (s *TestSuite) assertEqualEvents(expected, actual sdk.Events, msgAndArgs ...interface{}) bool {
	return assertions.AssertEqualEvents(s.T(), expected, actual, msgAndArgs...)
}
