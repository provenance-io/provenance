package keeper_test

import (
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

var (
	pastDate        time.Time
	validNftAddress string
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

	// Create a timestamp 24 hours in the past to avoid future date errors
	pastDate = time.Now().Add(-24 * time.Hour).UTC()
	validNftAddress = s.addr1.String()
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

// TestGetLedgerEntry tests the GetLedgerEntry function
func (s *TestSuite) TestGetLedgerEntry() {
	// Create a test ledger
	l := ledger.Ledger{
		NftAddress: s.addr1.String(),
		Denom:      s.bondDenom,
	}

	err := s.keeper.CreateLedger(s.ctx, l)
	s.Require().NoError(err, "CreateLedger error")

	expErr := keeper.ErrCodeInvalidField

	// Test cases
	tests := []struct {
		name          string
		nftAddr       string
		correlationId string
		expEntry      *ledger.LedgerEntry
		expErr        *string
	}{
		{
			name:          "invalid nft address",
			nftAddr:       "invalid",
			correlationId: "test-correlation-id",
			expEntry:      nil,
			expErr:        &expErr,
		},
		{
			name:          "not found",
			nftAddr:       s.addr2.String(),
			correlationId: "test-correlation-id",
			expEntry:      nil,
			expErr:        nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			entry, err := s.keeper.GetLedgerEntry(s.ctx, tc.nftAddr, tc.correlationId)
			if tc.expErr != nil {
				s.Require().Error(err, "expected GetLedgerEntry error")
				s.Require().Contains(err.Error(), *tc.expErr, "expected INVALID_FIELD error")
			} else {
				s.Require().NoError(err, "expected no GetLedgerEntry error")
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
	s.ctx = s.ctx.WithBlockTime(time.Now())

	// Create a test ledger
	l := ledger.Ledger{
		NftAddress: s.addr1.String(),
		Denom:      s.bondDenom,
	}

	err := s.keeper.CreateLedger(s.ctx, l)
	s.Require().NoError(err, "CreateLedger error")

	// Use a past date for testing
	pastDate := time.Now().Add(-24 * time.Hour).Format("2006-01-02")

	// Test cases
	tests := []struct {
		name    string
		nftAddr string
		entry   ledger.LedgerEntry
		expErr  *string
	}{
		{
			name:    "invalid nft address",
			nftAddr: "invalid",
			entry: ledger.LedgerEntry{
				Type:          ledger.LedgerEntryType_Scheduled_Payment,
				PostedDate:    pastDate,
				EffectiveDate: pastDate,
				TotalAmt:      s.int(100),
				AppliedAmounts: []*ledger.LedgerBucketAmount{
					{
						AppliedAmt: s.int(50),
						BalanceAmt: s.int(50),
						Bucket:     "principal",
					},
					{
						AppliedAmt: s.int(25),
						BalanceAmt: s.int(25),
						Bucket:     "interest",
					},
				},
				CorrelationId: "test-correlation-id-9",
			},
			expErr: keeper.StrPtr(keeper.ErrCodeNotFound),
		},
		{
			name:    "not found",
			nftAddr: s.addr2.String(),
			entry: ledger.LedgerEntry{
				Type:          ledger.LedgerEntryType_Scheduled_Payment,
				PostedDate:    pastDate,
				EffectiveDate: pastDate,
				TotalAmt:      s.int(100),
				AppliedAmounts: []*ledger.LedgerBucketAmount{
					{
						AppliedAmt: s.int(50),
						BalanceAmt: s.int(50),
						Bucket:     "principal",
					},
					{
						AppliedAmt: s.int(25),
						BalanceAmt: s.int(25),
						Bucket:     "interest",
					},
				},
				CorrelationId: "test-correlation-id-10",
			},
			expErr: keeper.StrPtr(keeper.ErrCodeNotFound),
		},
		{
			name:    "amounts_do_not_sum_to_total",
			nftAddr: s.addr1.String(),
			entry: ledger.LedgerEntry{
				Type:          ledger.LedgerEntryType_Scheduled_Payment,
				PostedDate:    pastDate,
				EffectiveDate: pastDate,
				TotalAmt:      s.int(100),
				AppliedAmounts: []*ledger.LedgerBucketAmount{
					{
						AppliedAmt: s.int(50),
						BalanceAmt: s.int(50),
						Bucket:     "principal",
					},
					{
						AppliedAmt: s.int(25),
						BalanceAmt: s.int(25),
						Bucket:     "interest",
					},
				},
				CorrelationId: "test-correlation-id-11",
			},
			expErr: keeper.StrPtr(keeper.ErrCodeInvalidField),
		},
		{
			name:    "missing_balance_fields",
			nftAddr: s.addr1.String(),
			entry: ledger.LedgerEntry{
				Type:            ledger.LedgerEntryType_Scheduled_Payment,
				PostedDate:      pastDate,
				EffectiveDate: pastDate,
				TotalAmt:      s.int(100),
				AppliedAmounts: []*ledger.LedgerBucketAmount{
					{
						AppliedAmt: s.int(50),
						BalanceAmt: s.int(50),
						Bucket:     "principal",
					},
					{
						AppliedAmt: s.int(25),
						BalanceAmt: s.int(25),
					},
				},
				CorrelationId: "test-correlation-id-13",
			},
			expErr: keeper.StrPtr(keeper.ErrCodeInvalidField),
		},
		{
			name:    "valid amounts and balances",
			nftAddr: s.addr1.String(),
			entry: ledger.LedgerEntry{
				Type:            ledger.LedgerEntryType_Scheduled_Payment,
				PostedDate:      pastDate,
				EffectiveDate:   pastDate,
				TotalAmt:      s.int(100),
				AppliedAmounts: []*ledger.LedgerBucketAmount{
					{
						AppliedAmt: s.int(50),
						BalanceAmt: s.int(50),
						Bucket:     "principal",
					},
					{
						AppliedAmt: s.int(25),
						BalanceAmt: s.int(25),
						Bucket:     "interest",
					},
				},
				CorrelationId: "test-correlation-id-12",
			},
			expErr: nil,
		},
		{
			name:    "negative amount",
			nftAddr: s.addr1.String(),
			entry: ledger.LedgerEntry{
				Type:            ledger.LedgerEntryType_Scheduled_Payment,
				PostedDate:      pastDate,
				EffectiveDate:   pastDate,
				TotalAmt:      s.int(-100),
				AppliedAmounts: []*ledger.LedgerBucketAmount{
					{
						AppliedAmt: s.int(50),
						BalanceAmt: s.int(50),
						Bucket:     "principal",
					},
					{
						AppliedAmt: s.int(25),
						BalanceAmt: s.int(25),
						Bucket:     "interest",
					},
				},
				CorrelationId:   "test-correlation-id-13",
			},
			expErr: keeper.StrPtr(keeper.ErrCodeInvalidField),
		},
		{
			name:    "allow negative principal applied amount",
			nftAddr: s.addr1.String(),
			entry: ledger.LedgerEntry{
				Type:            ledger.LedgerEntryType_Scheduled_Payment,
				PostedDate:      pastDate,
				EffectiveDate:   pastDate,
				TotalAmt:      s.int(100),
				AppliedAmounts: []*ledger.LedgerBucketAmount{
					{
						AppliedAmt: s.int(50),
						BalanceAmt: s.int(50),
						Bucket:     "principal",
					},
					{
						AppliedAmt: s.int(25),
						BalanceAmt: s.int(25),
						Bucket:     "interest",
					},
				CorrelationId:   "test-correlation-id-14",
			},
			expErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := s.keeper.AppendEntries(s.ctx, tc.nftAddr, []*ledger.LedgerEntry{&tc.entry})
			if tc.expErr != nil {
				s.Require().Error(err, "AppendEntry error")
				s.Require().Contains(err.Error(), *tc.expErr, "AppendEntry error type")
			} else {
				s.Require().NoError(err, "AppendEntry error")
			}
		})
	}
}

func (s *TestSuite) TestAppendEntrySequenceNumbers() {
	s.ctx = s.ctx.WithBlockTime(time.Now())

	// Create a test ledger
	l := ledger.Ledger{
		NftAddress: s.addr2.String(),
		Denom:      s.bondDenom,
	}

	err := s.keeper.CreateLedger(s.ctx, l)
	s.Require().NoError(err, "CreateLedger error")

	// Use a past date for testing
	pastDate := time.Now().Add(-24 * time.Hour).Format("2006-01-02")

	// Create test entries with the same effective date but different sequence numbers
	entries := []*ledger.LedgerEntry{
		{
			Type:            ledger.LedgerEntryType_Scheduled_Payment,
			PostedDate:      pastDate,
			EffectiveDate:   pastDate,
			Sequence:        1,
			Amt:             s.int(100),
			PrinAppliedAmt:  s.int(50),
			PrinBalAmt:      s.int(50),
			IntAppliedAmt:   s.int(25),
			IntBalAmt:       s.int(25),
			OtherAppliedAmt: s.int(25),
			OtherBalAmt:     s.int(25),
			CorrelationId:   "test-correlation-id-1",
		},
		{
			Type:            ledger.LedgerEntryType_Scheduled_Payment,
			PostedDate:      pastDate,
			EffectiveDate:   pastDate,
			Sequence:        2,
			Amt:             s.int(100),
			PrinAppliedAmt:  s.int(50),
			PrinBalAmt:      s.int(50),
			IntAppliedAmt:   s.int(25),
			IntBalAmt:       s.int(25),
			OtherAppliedAmt: s.int(25),
			OtherBalAmt:     s.int(25),
			CorrelationId:   "test-correlation-id-2",
		},
		{
			Type:            ledger.LedgerEntryType_Scheduled_Payment,
			PostedDate:      pastDate,
			EffectiveDate:   pastDate,
			Sequence:        3,
			Amt:             s.int(100),
			PrinAppliedAmt:  s.int(50),
			PrinBalAmt:      s.int(50),
			IntAppliedAmt:   s.int(25),
			IntBalAmt:       s.int(25),
			OtherAppliedAmt: s.int(25),
			OtherBalAmt:     s.int(25),
			CorrelationId:   "test-correlation-id-3",
		},
	}

	// Add entries in a specific order to test sequence number adjustment
	// First add entry with sequence 2
	err = s.keeper.AppendEntries(s.ctx, s.addr2.String(), []*ledger.LedgerEntry{entries[1]})
	s.Require().NoError(err, "AppendEntry error for sequence 2")
	allEntries, err := s.keeper.ListLedgerEntries(s.ctx, s.addr2.String())
	s.Require().NoError(err, "ListLedgerEntries error")
	s.Require().Equal(uint32(2), allEntries[0].Sequence, "sequence number for correlation-id-2")

	// Then add entry with sequence 1
	err = s.keeper.AppendEntries(s.ctx, s.addr2.String(), []*ledger.LedgerEntry{entries[0]})
	s.Require().NoError(err, "AppendEntry error for sequence 1")
	allEntries, err = s.keeper.ListLedgerEntries(s.ctx, s.addr2.String())
	s.Require().NoError(err, "ListLedgerEntries error")
	s.Require().Equal(uint32(1), allEntries[0].Sequence, "sequence number for correlation-id-2")

	// Finally add entry with sequence 3
	err = s.keeper.AppendEntries(s.ctx, s.addr2.String(), []*ledger.LedgerEntry{entries[2]})
	s.Require().NoError(err, "AppendEntry error for sequence 3")
	allEntries, err = s.keeper.ListLedgerEntries(s.ctx, s.addr2.String())
	s.Require().NoError(err, "ListLedgerEntries error")
	s.Require().Equal(uint32(3), allEntries[2].Sequence, "sequence number for correlation-id-2")

	// Get all entries and verify their sequence numbers
	allEntries, err = s.keeper.ListLedgerEntries(s.ctx, s.addr2.String())
	s.Require().NoError(err, "ListLedgerEntries error")

	// Verify sequence numbers
	s.Require().Len(allEntries, 3, "number of entries")
	s.Require().Equal(uint32(1), allEntries[0].Sequence, "sequence number for correlation-id-1")
	s.Require().Equal(uint32(2), allEntries[1].Sequence, "sequence number for correlation-id-2")
	s.Require().Equal(uint32(3), allEntries[2].Sequence, "sequence number for correlation-id-3")

	// Add another entry with sequence 2 to test sequence number adjustment
	newEntry := ledger.LedgerEntry{
		Type:            ledger.LedgerEntryType_Scheduled_Payment,
		PostedDate:      pastDate,
		EffectiveDate:   pastDate,
		Sequence:        2,
		Amt:             s.int(100),
		PrinAppliedAmt:  s.int(50),
		PrinBalAmt:      s.int(50),
		IntAppliedAmt:   s.int(25),
		IntBalAmt:       s.int(25),
		OtherAppliedAmt: s.int(25),
		OtherBalAmt:     s.int(25),
		CorrelationId:   "test-correlation-id-4",
	}

	err = s.keeper.AppendEntries(s.ctx, s.addr2.String(), []*ledger.LedgerEntry{&newEntry})
	s.Require().NoError(err, "AppendEntry error for new entry with sequence 2")

	// Get all entries again and verify updated sequence numbers
	allEntries, err = s.keeper.ListLedgerEntries(s.ctx, s.addr2.String())
	s.Require().NoError(err, "ListLedgerEntries error")

	// Verify updated sequence numbers
	s.Require().Len(allEntries, 4, "number of entries after adding new entry")
	s.Require().Equal(uint32(1), allEntries[0].Sequence, "sequence number for correlation-id-1")
	s.Require().Equal(uint32(2), allEntries[1].Sequence, "sequence number for correlation-id-4 (new entry)")
	s.Require().Equal(uint32(3), allEntries[2].Sequence, "sequence number for correlation-id-2 (shifted)")
	s.Require().Equal(uint32(4), allEntries[3].Sequence, "sequence number for correlation-id-3 (shifted)")
}

func (s *TestSuite) TestAppendEntryDuplicateCorrelationId() {
	s.ctx = s.ctx.WithBlockTime(time.Now())

	// Create a test ledger
	l := ledger.Ledger{
		NftAddress: s.addr1.String(),
		Denom:      s.bondDenom,
	}

	err := s.keeper.CreateLedger(s.ctx, l)
	s.Require().NoError(err, "CreateLedger error")

	// Use a past date for testing
	pastDate := time.Now().Add(-24 * time.Hour).Format("2006-01-02")

	// Create a test entry
	entry := ledger.LedgerEntry{
		Type:            ledger.LedgerEntryType_Scheduled_Payment,
		PostedDate:      pastDate,
		EffectiveDate:   pastDate,
		Sequence:        1,
		Amt:             s.int(100),
		PrinAppliedAmt:  s.int(50),
		PrinBalAmt:      s.int(50),
		IntAppliedAmt:   s.int(25),
		IntBalAmt:       s.int(25),
		OtherAppliedAmt: s.int(25),
		OtherBalAmt:     s.int(25),
		CorrelationId:   "test-correlation-id-1",
	}

	// Add the entry successfully
	err = s.keeper.AppendEntries(s.ctx, s.addr1.String(), []*ledger.LedgerEntry{&entry})
	s.Require().NoError(err, "AppendEntry error for first entry")

	// Try to add the same entry again with the same correlation ID
	err = s.keeper.AppendEntries(s.ctx, s.addr1.String(), []*ledger.LedgerEntry{&entry})
	s.Require().Error(err, "AppendEntry should fail for duplicate correlation ID")
	s.Require().Contains(err.Error(), keeper.ErrCodeAlreadyExists, "error should be ErrCodeAlreadyExists")

	// Verify that only one entry exists
	allEntries, err := s.keeper.ListLedgerEntries(s.ctx, s.addr1.String())
	s.Require().NoError(err, "ListLedgerEntries error")
	s.Require().Len(allEntries, 1, "should only have one entry")

	// Try to add a different entry with the same correlation ID
	entry2 := ledger.LedgerEntry{
		Type:            ledger.LedgerEntryType_Scheduled_Payment,
		PostedDate:      pastDate,
		EffectiveDate:   pastDate,
		Sequence:        2,
		Amt:             s.int(200),
		PrinAppliedAmt:  s.int(100),
		PrinBalAmt:      s.int(100),
		IntAppliedAmt:   s.int(50),
		IntBalAmt:       s.int(50),
		OtherAppliedAmt: s.int(50),
		OtherBalAmt:     s.int(50),
		CorrelationId:   "test-correlation-id-1", // Same correlation ID
	}

	err = s.keeper.AppendEntries(s.ctx, s.addr1.String(), []*ledger.LedgerEntry{&entry2})
	s.Require().Error(err, "AppendEntry should fail for duplicate correlation ID")
	s.Require().Contains(err.Error(), keeper.ErrCodeAlreadyExists, "error should be ErrCodeAlreadyExists")

	// Verify that still only one entry exists
	allEntries, err = s.keeper.ListLedgerEntries(s.ctx, s.addr1.String())
	s.Require().NoError(err, "ListLedgerEntries error")
	s.Require().Len(allEntries, 1, "should still only have one entry")
	s.Require().Equal(entry.Amt, allEntries[0].Amt, "entry amount should match original entry")
}
