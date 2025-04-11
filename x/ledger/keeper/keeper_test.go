package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/ledger"
	"github.com/provenance-io/provenance/x/ledger/keeper"
)

type TestSuite struct {
	suite.Suite

	app        *app.App
	ctx        sdk.Context
	keeper     keeper.BaseKeeper
	bankKeeper bankkeeper.Keeper

	bondDenom  string
	initBal    sdk.Coins
	initAmount int64

	addr1 sdk.AccAddress
	addr2 sdk.AccAddress
	addr3 sdk.AccAddress
}

func (s *TestSuite) SetupTest() {
	s.app = app.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(false)
	s.keeper = s.app.LedgerKeeper
	s.bankKeeper = s.app.BankKeeper

	var err error
	s.bondDenom, err = s.app.StakingKeeper.BondDenom(s.ctx)
	s.Require().NoError(err, "app.StakingKeeper.BondDenom(s.ctx)")

	s.initAmount = 1_000_000_000
	s.initBal = sdk.NewCoins(sdk.NewCoin(s.bondDenom, sdkmath.NewInt(s.initAmount)))

	addrs := app.AddTestAddrsIncremental(s.app, s.ctx, 3, sdkmath.NewInt(s.initAmount))
	s.addr1 = addrs[0]
	s.addr2 = addrs[1]
	s.addr3 = addrs[2]
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

// TestCreateLedger tests the CreateLedger function
func (s *TestSuite) TestCreateLedger() {
	// Create a valid NFT address for testing
	nftAddr := s.addr1.String()
	denom := s.bondDenom

	tests := []struct {
		name     string
		ledger   ledger.Ledger
		expErr   []string
		expEvent bool
	}{
		{
			name: "valid ledger",
			ledger: ledger.Ledger{
				NftAddress: nftAddr,
				Denom:      denom,
			},
			expEvent: true,
		},
		{
			name: "empty nft address",
			ledger: ledger.Ledger{
				NftAddress: "",
				Denom:      denom,
			},
			expErr: []string{"nft_address"},
		},
		{
			name: "empty denom",
			ledger: ledger.Ledger{
				NftAddress: nftAddr,
				Denom:      "",
			},
			expErr: []string{"denom"},
		},
		{
			name: "invalid nft address",
			ledger: ledger.Ledger{
				NftAddress: "invalid",
				Denom:      denom,
			},
			expErr: []string{"nft_address"},
		},
		{
			name: "duplicate ledger",
			ledger: ledger.Ledger{
				NftAddress: nftAddr,
				Denom:      denom,
			},
			expErr: []string{"already exists"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Clear events before each test
			s.ctx.EventManager().Events()

			err := s.keeper.CreateLedger(s.ctx, tc.ledger)

			if len(tc.expErr) > 0 {
				s.assertErrorContents(err, tc.expErr, "CreateLedger error")
			} else {
				s.Require().NoError(err, "CreateLedger error")

				// Verify the ledger was created
				l, err := s.keeper.GetLedger(s.ctx, tc.ledger.NftAddress)
				s.Require().NoError(err, "GetLedger error")
				s.Require().NotNil(l, "GetLedger result")
				s.Require().Equal(tc.ledger.NftAddress, l.NftAddress, "ledger nft address")
				s.Require().Equal(tc.ledger.Denom, l.Denom, "ledger denom")

				// Verify event emission
				if tc.expEvent {
					// Find the expected event
					var foundEvent *sdk.Event
					for _, e := range s.ctx.EventManager().Events() {
						if e.Type == ledger.EventTypeLedgerCreated {
							foundEvent = &e
							break
						}
					}

					s.Require().NotNil(foundEvent)
					s.Require().Equal(ledger.EventTypeLedgerCreated, foundEvent.Type, "event type")
					s.Require().Len(foundEvent.Attributes, 2, "event attributes length")
					s.Require().Equal("nft_address", foundEvent.Attributes[0].Key, "event nft address key")
					s.Require().Equal(tc.ledger.NftAddress, foundEvent.Attributes[0].Value, "event nft address value")
					s.Require().Equal("denom", foundEvent.Attributes[1].Key, "event denom key")
					s.Require().Equal(tc.ledger.Denom, foundEvent.Attributes[1].Value, "event denom value")
				}
			}
		})
	}
}

// TestGetLedger tests the GetLedger function
func (s *TestSuite) TestGetLedger() {
	// Create a valid NFT address for testing
	nftAddr := s.addr1.String()
	denom := s.bondDenom

	// Create a valid ledger first that we can try to get
	validLedger := ledger.Ledger{
		NftAddress: nftAddr,
		Denom:      denom,
	}
	err := s.keeper.CreateLedger(s.ctx, validLedger)
	s.Require().NoError(err, "CreateLedger error")

	tests := []struct {
		name      string
		nftAddr   string
		expErr    []string
		expLedger *ledger.Ledger
	}{
		{
			name:      "valid ledger retrieval",
			nftAddr:   nftAddr,
			expLedger: &validLedger,
		},
		{
			name:    "empty nft address",
			nftAddr: "",
			expErr:  []string{"nft_address"},
		},
		{
			name:      "non-existent ledger",
			nftAddr:   s.addr2.String(),
			expLedger: nil,
		},
		{
			name:    "invalid nft address",
			nftAddr: "invalid",
			expErr:  []string{"nft_address"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			ledger, err := s.keeper.GetLedger(s.ctx, tc.nftAddr)

			if len(tc.expErr) > 0 {
				s.assertErrorContents(err, tc.expErr, "GetLedger error")
				s.Require().Nil(ledger, "GetLedger result should be nil on error")
			} else {
				s.Require().NoError(err, "GetLedger error")
				if tc.expLedger == nil {
					s.Require().Nil(ledger, "GetLedger result should be nil for non-existent ledger")
				} else {
					s.Require().NotNil(ledger, "GetLedger result")
					s.Require().Equal(tc.expLedger.NftAddress, ledger.NftAddress, "ledger nft address")
					s.Require().Equal(tc.expLedger.Denom, ledger.Denom, "ledger denom")
				}
			}
		})
	}
}

// TestCreateLedgerEntry tests the CreateLedgerEntry function
func (s *TestSuite) TestCreateLedgerEntry() {
	// Create a valid NFT address and ledger for testing
	nftAddr := s.addr1.String()
	denom := s.bondDenom
	validLedger := ledger.Ledger{
		NftAddress: nftAddr,
		Denom:      denom,
	}
	err := s.keeper.CreateLedger(s.ctx, validLedger)
	s.Require().NoError(err, "CreateLedger error")

	// Create test entry data
	entryType := ledger.LedgerEntryType_Disbursement
	postedDate := time.Now()
	effectiveDate := postedDate.Add(24 * time.Hour)
	amount := s.int(1000)

	// Create a valid ledger entry first
	validEntry := ledger.LedgerEntry{
		Type:            entryType,
		PostedDate:      postedDate,
		EffectiveDate:   effectiveDate,
		Amt:             amount,
		PrinAppliedAmt:  amount,
		PrinBalAmt:      amount,
		IntAppliedAmt:   s.int(0),
		IntBalAmt:       s.int(0),
		OtherAppliedAmt: s.int(0),
		OtherBalAmt:     s.int(0),
		Sequence:        1,
	}
	err = s.keeper.AppendEntry(s.ctx, nftAddr, validEntry)
	s.Require().NoError(err, "AppendEntry error")

	tests := []struct {
		name     string
		nftAddr  string
		entry    ledger.LedgerEntry
		expErr   []string
		expEvent bool
	}{
		{
			name:    "valid ledger entry",
			nftAddr: nftAddr,
			entry: ledger.LedgerEntry{
				Type:            entryType,
				PostedDate:      postedDate,
				EffectiveDate:   effectiveDate,
				Amt:             amount,
				PrinAppliedAmt:  amount,
				PrinBalAmt:      amount,
				IntAppliedAmt:   s.int(0),
				IntBalAmt:       s.int(0),
				OtherAppliedAmt: s.int(0),
				OtherBalAmt:     s.int(0),
				Sequence:        1,
				CorrelationId:   "test-correlation-id-1",
			},
			expEvent: true,
		},
		{
			name:    "empty nft address",
			nftAddr: "",
			entry: ledger.LedgerEntry{
				Type:            entryType,
				PostedDate:      postedDate,
				EffectiveDate:   effectiveDate,
				Amt:             amount,
				PrinAppliedAmt:  amount,
				PrinBalAmt:      amount,
				IntAppliedAmt:   s.int(0),
				IntBalAmt:       s.int(0),
				OtherAppliedAmt: s.int(0),
				OtherBalAmt:     s.int(0),
				Sequence:        1,
				CorrelationId:   "test-correlation-id-2",
			},
			expErr: []string{"nft_address"},
		},
		{
			name:    "unspecified entry type",
			nftAddr: nftAddr,
			entry: ledger.LedgerEntry{
				Type:            ledger.LedgerEntryType_Unspecified,
				PostedDate:      postedDate,
				EffectiveDate:   effectiveDate,
				Amt:             amount,
				PrinAppliedAmt:  amount,
				PrinBalAmt:      amount,
				IntAppliedAmt:   s.int(0),
				IntBalAmt:       s.int(0),
				OtherAppliedAmt: s.int(0),
				OtherBalAmt:     s.int(0),
				Sequence:        1,
				CorrelationId:   "test-correlation-id-3",
			},
			expErr: []string{"type"},
		},
		{
			name:    "zero posted date",
			nftAddr: nftAddr,
			entry: ledger.LedgerEntry{
				Type:            entryType,
				PostedDate:      time.Time{},
				EffectiveDate:   effectiveDate,
				Amt:             amount,
				PrinAppliedAmt:  amount,
				PrinBalAmt:      amount,
				IntAppliedAmt:   s.int(0),
				IntBalAmt:       s.int(0),
				OtherAppliedAmt: s.int(0),
				OtherBalAmt:     s.int(0),
				Sequence:        1,
				CorrelationId:   "test-correlation-id-4",
			},
			expErr: []string{"posted_date"},
		},
		{
			name:    "zero effective date",
			nftAddr: nftAddr,
			entry: ledger.LedgerEntry{
				Type:            entryType,
				PostedDate:      postedDate,
				EffectiveDate:   time.Time{},
				Amt:             amount,
				PrinAppliedAmt:  amount,
				PrinBalAmt:      amount,
				IntAppliedAmt:   s.int(0),
				IntBalAmt:       s.int(0),
				OtherAppliedAmt: s.int(0),
				OtherBalAmt:     s.int(0),
				Sequence:        1,
				CorrelationId:   "test-correlation-id-5",
			},
			expErr: []string{"effective_date"},
		},
		{
			name:    "negative amount",
			nftAddr: nftAddr,
			entry: ledger.LedgerEntry{
				Type:            entryType,
				PostedDate:      postedDate,
				EffectiveDate:   effectiveDate,
				Amt:             s.int(-1000),
				PrinAppliedAmt:  amount,
				PrinBalAmt:      amount,
				IntAppliedAmt:   s.int(0),
				IntBalAmt:       s.int(0),
				OtherAppliedAmt: s.int(0),
				OtherBalAmt:     s.int(0),
				Sequence:        1,
				CorrelationId:   "test-correlation-id-6",
			},
			expErr: []string{"amount"},
		},
		{
			name:    "negative principal applied amount",
			nftAddr: nftAddr,
			entry: ledger.LedgerEntry{
				Type:            entryType,
				PostedDate:      postedDate,
				EffectiveDate:   effectiveDate,
				Amt:             amount,
				PrinAppliedAmt:  s.int(-1000),
				PrinBalAmt:      amount,
				IntAppliedAmt:   s.int(0),
				IntBalAmt:       s.int(0),
				OtherAppliedAmt: s.int(0),
				OtherBalAmt:     s.int(0),
				Sequence:        1,
				CorrelationId:   "test-correlation-id-7",
			},
			expErr: []string{"principal applied amount"},
		},
		{
			name:    "non-existent ledger",
			nftAddr: s.addr2.String(),
			entry: ledger.LedgerEntry{
				Type:            entryType,
				PostedDate:      postedDate,
				EffectiveDate:   effectiveDate,
				Amt:             amount,
				PrinAppliedAmt:  amount,
				PrinBalAmt:      amount,
				IntAppliedAmt:   s.int(0),
				IntBalAmt:       s.int(0),
				OtherAppliedAmt: s.int(0),
				OtherBalAmt:     s.int(0),
				Sequence:        1,
				CorrelationId:   "test-correlation-id-8",
			},
			expErr: []string{"not found"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Reset event manager for each test
			s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())

			err := s.keeper.AppendEntry(s.ctx, tc.nftAddr, tc.entry)
			if len(tc.expErr) > 0 {
				s.Require().Error(err)
				for _, expErr := range tc.expErr {
					s.Require().Contains(err.Error(), expErr)
				}
			} else {
				s.Require().NoError(err)
			}

			// Check events
			events := s.ctx.EventManager().Events()
			if tc.expEvent {
				s.Require().Len(events, 1)
				event := events[0]
				s.Require().Equal(ledger.EventTypeLedgerEntryAdded, event.Type)
				s.Require().Len(event.Attributes, 6)
				s.Require().Equal("nft_address", string(event.Attributes[0].Key))
				s.Require().Equal(tc.nftAddr, string(event.Attributes[0].Value))
				s.Require().Equal("correlation_id", string(event.Attributes[1].Key))
				s.Require().Equal(tc.entry.CorrelationId, string(event.Attributes[1].Value))
				s.Require().Equal("entry_type", string(event.Attributes[2].Key))
				s.Require().Equal(tc.entry.Type.String(), string(event.Attributes[2].Value))
				s.Require().Equal("posted_date", string(event.Attributes[3].Key))
				s.Require().Equal(tc.entry.PostedDate.Format(time.RFC3339), string(event.Attributes[3].Value))
				s.Require().Equal("effective_date", string(event.Attributes[4].Key))
				s.Require().Equal(tc.entry.EffectiveDate.Format(time.RFC3339), string(event.Attributes[4].Value))
				s.Require().Equal("amount", string(event.Attributes[5].Key))
				s.Require().Equal(tc.entry.Amt.String(), string(event.Attributes[5].Value))
			} else {
				s.Require().Len(events, 0)
			}
		})
	}
}

// TestGetLedgerEntry tests the GetLedgerEntry function
func (s *TestSuite) TestGetLedgerEntry() {
	// Create a test ledger
	l := ledger.Ledger{
		NftAddress: s.addr1.String(),
		Denom:      s.bondDenom,
	}

	err := s.keeper.CreateLedger(s.ctx, l)
	s.Require().NoError(err, "CreateLedger error")

	// Test cases
	tests := []struct {
		name     string
		nftAddr  string
		expEntry *ledger.LedgerEntry
		expErr   error
	}{
		{
			name:     "invalid nft address",
			nftAddr:  "invalid",
			expEntry: nil,
			expErr:   fmt.Errorf("invalid nft_address"),
		},
		{
			name:     "not found",
			nftAddr:  s.addr2.String(),
			expEntry: nil,
			expErr:   nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			entry, err := s.keeper.GetLedgerEntry(s.ctx, tc.nftAddr, "")
			if tc.expErr != nil {
				s.Require().Error(err, "GetLedgerEntry error")
				s.Require().Equal(tc.expErr.Error(), err.Error(), "GetLedgerEntry error type")
			} else {
				s.Require().NoError(err, "GetLedgerEntry error")
				s.Require().Equal(tc.expEntry, entry, "GetLedgerEntry result")
			}
		})
	}
}

// TestProcessFundTransfer tests the ProcessFundTransfer function
func (s *TestSuite) TestProcessFundTransfer() {
	// TODO: Implement test cases for ProcessFundTransfer
}

// TestInitGenesis tests the InitGenesis function
func (s *TestSuite) TestInitGenesis() {
	tests := []struct {
		name     string
		genState *ledger.GenesisState
	}{
		{
			name:     "empty genesis state",
			genState: &ledger.GenesisState{},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Clear events before each test
			s.ctx.EventManager().Events()

			// Initialize genesis state
			s.keeper.InitGenesis(s.ctx, tc.genState)
		})
	}
}

// TestExportGenesis tests the ExportGenesis function
func (s *TestSuite) TestExportGenesis() {
	// Export genesis state
	genState := s.keeper.ExportGenesis(s.ctx)
	s.Require().NotNil(genState, "exported genesis state should not be nil")
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

// int is a shorter way to call sdkmath.NewInt.
func (s *TestSuite) int(amount int64) sdkmath.Int {
	return sdkmath.NewInt(amount)
}

// intStr creates an sdkmath.Int from a string, requiring it to work.
func (s *TestSuite) intStr(amount string) sdkmath.Int {
	s.T().Helper()
	rv, ok := sdkmath.NewIntFromString(amount)
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

func (s *TestSuite) TestAppendEntry() {
	// Create a test ledger
	l := ledger.Ledger{
		NftAddress: s.addr1.String(),
		Denom:      s.bondDenom,
	}

	err := s.keeper.CreateLedger(s.ctx, l)
	s.Require().NoError(err, "CreateLedger error")

	// Test cases
	tests := []struct {
		name    string
		nftAddr string
		entry   ledger.LedgerEntry
		expErr  error
	}{
		{
			name:    "invalid nft address",
			nftAddr: "invalid",
			entry: ledger.LedgerEntry{
				Type:          ledger.LedgerEntryType_Payment,
				PostedDate:    time.Now(),
				EffectiveDate: time.Now(),
				Amt:           s.int(100),
				CorrelationId: "test-correlation-id-9",
			},
			expErr: fmt.Errorf("invalid nft_address"),
		},
		{
			name:    "not found",
			nftAddr: s.addr2.String(),
			entry: ledger.LedgerEntry{
				Type:          ledger.LedgerEntryType_Payment,
				PostedDate:    time.Now(),
				EffectiveDate: time.Now(),
				Amt:           s.int(100),
				CorrelationId: "test-correlation-id-10",
			},
			expErr: fmt.Errorf("ledger not found"),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := s.keeper.AppendEntry(s.ctx, tc.nftAddr, tc.entry)
			if tc.expErr != nil {
				s.Require().Error(err, "AppendEntry error")
				s.Require().Equal(tc.expErr.Error(), err.Error(), "AppendEntry error type")
			} else {
				s.Require().NoError(err, "AppendEntry error")
			}
		})
	}
}
