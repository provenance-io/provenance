package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/math"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/x/nft"
	nftkeeper "cosmossdk.io/x/nft/keeper"
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
	nftKeeper  nftkeeper.Keeper

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
	s.pastDate = keeper.DaysSinceEpoch(time.Now().Add(-24 * time.Hour).UTC())

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
		LedgerClassId: "test-ledger-class-id",
		AssetClassId:  s.validNFTClass.Id,
		Denom:         s.bondDenom,
	}
	s.keeper.CreateLedgerClass(s.ctx, s.validLedgerClass)

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

func (s *TestSuite) TestCreateLedgerClass() {

	tests := []struct {
		name        string
		ledgerClass ledger.LedgerClass
		expErr      []string
	}{
		{
			name: "valid ledger class should already exist",
			ledgerClass: ledger.LedgerClass{
				LedgerClassId: s.validLedgerClass.LedgerClassId,
				AssetClassId:  s.validLedgerClass.AssetClassId,
				Denom:         s.bondDenom,
			},
			expErr: []string{"already exists"},
		},
		{
			name: "invalid asset class id",
			ledgerClass: ledger.LedgerClass{
				LedgerClassId: s.validLedgerClass.LedgerClassId,
				AssetClassId:  "non-existent-class-id",
				Denom:         s.bondDenom,
			},
			expErr: []string{"asset_class_id"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := s.keeper.CreateLedgerClass(s.ctx, tc.ledgerClass)
			if len(tc.expErr) > 0 {
				s.assertErrorContents(err, tc.expErr, "CreateLedgerClass error")
			} else {
				s.Require().NoError(err, "CreateLedgerClass error")
			}
		})
	}
}

// TestCreateLedger tests the CreateLedger function
func (s *TestSuite) TestCreateLedger() {
	tests := []struct {
		name     string
		ledger   ledger.Ledger
		expErr   []string
		expEvent bool
	}{
		{
			name: "valid ledger",
			ledger: ledger.Ledger{
				Key: &ledger.LedgerKey{
					AssetClassId: s.validNFTClass.Id,
					NftId:        s.validNFT.Id,
				},
				LedgerClassId: s.validLedgerClass.LedgerClassId,
				StatusTypeId:  1,
			},
			expEvent: true,
		},
		{
			name: "empty nft address",
			ledger: ledger.Ledger{
				Key: &ledger.LedgerKey{
					AssetClassId: s.validNFTClass.Id,
					NftId:        "",
				},
				LedgerClassId: s.validLedgerClass.LedgerClassId,
				StatusTypeId:  1,
			},
			expErr: []string{"nft_id"},
		},
		{
			name: "duplicate ledger",
			ledger: ledger.Ledger{
				Key: &ledger.LedgerKey{
					AssetClassId: s.validNFTClass.Id,
					NftId:        s.validNFT.Id,
				},
				LedgerClassId: s.validLedgerClass.LedgerClassId,
				StatusTypeId:  1,
			},
			expErr: []string{"already exists"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Clear events before each test
			s.ctx.EventManager().Events()

			err := s.keeper.CreateLedger(s.ctx, s.addr1, tc.ledger)

			if len(tc.expErr) > 0 {
				s.assertErrorContents(err, tc.expErr, "CreateLedger error")
			} else {
				s.Require().NoError(err, "CreateLedger error")

				// Verify the ledger was created
				l, err := s.keeper.GetLedger(s.ctx, tc.ledger.Key)
				s.Require().NoError(err, "GetLedger error")
				s.Require().NotNil(l, "GetLedger result")
				s.Require().Equal(tc.ledger.Key.NftId, l.Key.NftId, "ledger nft address")
				s.Require().Equal(tc.ledger.Key.AssetClassId, l.Key.AssetClassId, "ledger asset class id")
				s.Require().Equal(tc.ledger.LedgerClassId, l.LedgerClassId, "ledger asset class id")

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
					s.Require().Equal("asset_class_id", foundEvent.Attributes[0].Key, "event asset class id key")
					s.Require().Equal(tc.ledger.Key.AssetClassId, foundEvent.Attributes[0].Value, "event asset class id value")
					s.Require().Equal("nft_id", foundEvent.Attributes[1].Key, "event nft id key")
					s.Require().Equal(tc.ledger.Key.NftId, foundEvent.Attributes[1].Value, "event nft id value")
				}
			}
		})
	}
}

// TestGetLedger tests the GetLedger function
func (s *TestSuite) TestGetLedger() {
	// Create a valid ledger first that we can try to get
	validLedger := ledger.Ledger{
		Key: &ledger.LedgerKey{
			AssetClassId: s.validNFTClass.Id,
			NftId:        s.validNFT.Id,
		},
		LedgerClassId: s.validLedgerClass.LedgerClassId,
		StatusTypeId:  1,
	}
	err := s.keeper.CreateLedger(s.ctx, s.addr1, validLedger)
	s.Require().NoError(err, "CreateLedger error")

	tests := []struct {
		name      string
		nftId     string
		expErr    []string
		expLedger *ledger.Ledger
	}{
		{
			name:      "valid ledger retrieval",
			nftId:     s.validNFT.Id,
			expLedger: &validLedger,
		},
		{
			name:   "empty nft address",
			nftId:  "",
			expErr: []string{"nft_id"},
		},
		{
			name:      "non-existent ledger",
			nftId:     s.addr2.String(),
			expLedger: nil,
		},
		{
			name:   "invalid nft address",
			nftId:  "invalid",
			expErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			ledger, err := s.keeper.GetLedger(s.ctx, &ledger.LedgerKey{
				AssetClassId: s.validNFTClass.Id,
				NftId:        tc.nftId,
			})

			if len(tc.expErr) > 0 {
				s.assertErrorContents(err, tc.expErr, "GetLedger error")
				s.Require().Nil(ledger, "GetLedger result should be nil on error")
			} else {
				s.Require().NoError(err, "GetLedger error")
				if tc.expLedger == nil {
					s.Require().Nil(ledger, "GetLedger result should be nil for non-existent ledger")
				} else {
					s.Require().NotNil(ledger, "GetLedger result")
					s.Require().Equal(tc.expLedger.Key.NftId, ledger.Key.NftId, "ledger nft address")
					s.Require().Equal(tc.expLedger.Key.AssetClassId, ledger.Key.AssetClassId, "ledger asset class id")
					s.Require().Equal(tc.expLedger.LedgerClassId, ledger.LedgerClassId, "ledger class id")
				}
			}
		})
	}
}

// TestGetLedgerEntry tests the GetLedgerEntry function
func (s *TestSuite) TestGetLedgerEntry() {
	// Create a test ledger
	l := ledger.Ledger{
		Key: &ledger.LedgerKey{
			AssetClassId: s.validNFTClass.Id,
			NftId:        s.validNFT.Id,
		},
		LedgerClassId: s.validLedgerClass.LedgerClassId,
		StatusTypeId:  1,
	}

	err := s.keeper.CreateLedger(s.ctx, s.addr1, l)
	s.Require().NoError(err, "CreateLedger error")

	expErr := keeper.ErrCodeNotFound

	// Test cases
	tests := []struct {
		name          string
		key           *ledger.LedgerKey
		correlationId string
		expEntry      *ledger.LedgerEntry
		expErr        *string
	}{
		{
			name: "invalid nft address",
			key: &ledger.LedgerKey{
				AssetClassId: s.validNFTClass.Id,
				NftId:        "invalid",
			},
			correlationId: "test-correlation-id",
			expEntry:      nil,
			expErr:        &expErr,
		},
		{
			name: "not found",
			key: &ledger.LedgerKey{
				AssetClassId: s.validNFTClass.Id,
				NftId:        s.validNFT.Id,
			},
			correlationId: "test-correlation-id-222",
			expEntry:      nil,
			expErr:        nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			entry, err := s.keeper.GetLedgerEntry(s.ctx, tc.key, tc.correlationId)
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

func (s *TestSuite) TestAppendEntry() {
	// Create a test ledger
	l := ledger.Ledger{
		Key: &ledger.LedgerKey{
			AssetClassId: s.validNFTClass.Id,
			NftId:        s.validNFT.Id,
		},
		LedgerClassId: s.validLedgerClass.LedgerClassId,
		StatusTypeId:  1,
	}

	err := s.keeper.CreateLedger(s.ctx, s.addr1, l)
	s.Require().NoError(err, "CreateLedger error")

	// Test cases
	tests := []struct {
		name   string
		key    *ledger.LedgerKey
		entry  ledger.LedgerEntry
		expErr *string
	}{
		{
			name: "invalid nft address",
			key: &ledger.LedgerKey{
				AssetClassId: s.validNFTClass.Id,
				NftId:        "invalid",
			},
			entry: ledger.LedgerEntry{
				EntryTypeId:   1,
				PostedDate:    s.pastDate,
				EffectiveDate: s.pastDate,
				TotalAmt:      math.NewInt(100),
				AppliedAmounts: []*ledger.LedgerBucketAmount{
					{
						AppliedAmt:   math.NewInt(50),
						BucketTypeId: 1,
					},
					{
						AppliedAmt:   math.NewInt(50),
						BucketTypeId: 2,
					},
				},
				CorrelationId: "test-correlation-id-9",
			},
			expErr: keeper.StrPtr(keeper.ErrCodeNotFound),
		},
		{
			name: "not found",
			key: &ledger.LedgerKey{
				AssetClassId: s.validNFTClass.Id,
				NftId:        "unknown nft",
			},
			entry: ledger.LedgerEntry{
				EntryTypeId:   1,
				PostedDate:    s.pastDate,
				EffectiveDate: s.pastDate,
				TotalAmt:      math.NewInt(100),
				AppliedAmounts: []*ledger.LedgerBucketAmount{
					{
						AppliedAmt:   math.NewInt(50),
						BucketTypeId: 1,
					},
					{
						AppliedAmt:   math.NewInt(50),
						BucketTypeId: 2,
					},
				},
				CorrelationId: "test-correlation-id-10",
			},
			expErr: keeper.StrPtr(keeper.ErrCodeNotFound),
		},
		{
			name: "amounts_do_not_sum_to_total",
			key: &ledger.LedgerKey{
				AssetClassId: s.validNFTClass.Id,
				NftId:        s.validNFT.Id,
			},
			entry: ledger.LedgerEntry{
				EntryTypeId:   1,
				PostedDate:    s.pastDate,
				EffectiveDate: s.pastDate,
				TotalAmt:      math.NewInt(100),
				AppliedAmounts: []*ledger.LedgerBucketAmount{
					{
						AppliedAmt:   math.NewInt(50),
						BucketTypeId: 1,
					},
					{
						AppliedAmt:   math.NewInt(25),
						BucketTypeId: 2,
					},
				},
				CorrelationId: "test-correlation-id-11",
			},
			expErr: keeper.StrPtr(keeper.ErrCodeInvalidField),
		},
		{
			name: "valid amounts and balances",
			key: &ledger.LedgerKey{
				AssetClassId: s.validNFTClass.Id,
				NftId:        s.validNFT.Id,
			},
			entry: ledger.LedgerEntry{
				EntryTypeId:   1,
				PostedDate:    s.pastDate,
				EffectiveDate: s.pastDate,
				TotalAmt:      math.NewInt(100),
				AppliedAmounts: []*ledger.LedgerBucketAmount{
					{
						AppliedAmt:   math.NewInt(50),
						BucketTypeId: 1,
					},
					{
						AppliedAmt:   math.NewInt(50),
						BucketTypeId: 2,
					},
				},
				CorrelationId: "test-correlation-id-12",
			},
			expErr: nil,
		},
		{
			name: "valid amounts and balances with negative applied amount",
			key: &ledger.LedgerKey{
				AssetClassId: s.validNFTClass.Id,
				NftId:        s.validNFT.Id,
			},
			entry: ledger.LedgerEntry{
				EntryTypeId:   1,
				PostedDate:    s.pastDate,
				EffectiveDate: s.pastDate,
				TotalAmt:      math.NewInt(100),
				AppliedAmounts: []*ledger.LedgerBucketAmount{
					{
						AppliedAmt:   math.NewInt(50),
						BucketTypeId: 1,
					},
					{
						AppliedAmt:   math.NewInt(-50),
						BucketTypeId: 2,
					},
				},
				CorrelationId: "test-correlation-id-13",
			},
			expErr: nil,
		},
		{
			name: "negative amount",
			key: &ledger.LedgerKey{
				AssetClassId: s.validNFTClass.Id,
				NftId:        s.validNFT.Id,
			},
			entry: ledger.LedgerEntry{
				EntryTypeId:   1,
				PostedDate:    s.pastDate,
				EffectiveDate: s.pastDate,
				TotalAmt:      math.NewInt(-100),
				AppliedAmounts: []*ledger.LedgerBucketAmount{
					{
						AppliedAmt:   math.NewInt(50),
						BucketTypeId: 1,
					},
					{
						AppliedAmt:   math.NewInt(50),
						BucketTypeId: 2,
					},
				},
				CorrelationId: "test-correlation-id-14",
			},
			expErr: keeper.StrPtr(keeper.ErrCodeInvalidField),
		},
		{
			name: "allow negative principal applied amount",
			key: &ledger.LedgerKey{
				AssetClassId: s.validNFTClass.Id,
				NftId:        s.validNFT.Id,
			},
			entry: ledger.LedgerEntry{
				EntryTypeId:   1,
				PostedDate:    s.pastDate,
				EffectiveDate: s.pastDate,
				TotalAmt:      math.NewInt(100),
				AppliedAmounts: []*ledger.LedgerBucketAmount{
					{
						AppliedAmt:   math.NewInt(50),
						BucketTypeId: 1,
					},
					{
						AppliedAmt:   math.NewInt(-50),
						BucketTypeId: 2,
					},
				},
				CorrelationId: "test-correlation-id-15",
			},
			expErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := s.keeper.AppendEntries(s.ctx, s.addr1, tc.key, []*ledger.LedgerEntry{&tc.entry})
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
	// Create a test ledger
	l := ledger.Ledger{
		Key: &ledger.LedgerKey{
			AssetClassId: s.validNFTClass.Id,
			NftId:        s.validNFT.Id,
		},
		LedgerClassId: s.validLedgerClass.LedgerClassId,
		StatusTypeId:  1,
	}

	err := s.keeper.CreateLedger(s.ctx, s.addr1, l)
	s.Require().NoError(err, "CreateLedger error")

	// Create test entries with the same effective date but different sequence numbers
	entries := []*ledger.LedgerEntry{
		{
			EntryTypeId:   1,
			PostedDate:    s.pastDate,
			EffectiveDate: s.pastDate,
			Sequence:      1,
			TotalAmt:      math.NewInt(100),
			AppliedAmounts: []*ledger.LedgerBucketAmount{
				{
					AppliedAmt:   math.NewInt(50),
					BucketTypeId: 1,
				},
				{
					AppliedAmt:   math.NewInt(50),
					BucketTypeId: 2,
				},
			},
			CorrelationId: "test-correlation-id-1",
		},
		{
			EntryTypeId:   1,
			PostedDate:    s.pastDate,
			EffectiveDate: s.pastDate,
			Sequence:      2,
			TotalAmt:      math.NewInt(100),
			AppliedAmounts: []*ledger.LedgerBucketAmount{
				{
					AppliedAmt:   math.NewInt(50),
					BucketTypeId: 1,
				},
				{
					AppliedAmt:   math.NewInt(50),
					BucketTypeId: 2,
				},
			},
			CorrelationId: "test-correlation-id-2",
		},
		{
			EntryTypeId:   1,
			PostedDate:    s.pastDate,
			EffectiveDate: s.pastDate,
			Sequence:      3,
			TotalAmt:      math.NewInt(100),
			AppliedAmounts: []*ledger.LedgerBucketAmount{
				{
					AppliedAmt:   math.NewInt(50),
					BucketTypeId: 1,
				},
				{
					AppliedAmt:   math.NewInt(50),
					BucketTypeId: 2,
				},
			},
			CorrelationId: "test-correlation-id-3",
		},
	}

	// Add entries in a specific order to test sequence number adjustment
	// First add entry with sequence 2
	err = s.keeper.AppendEntries(s.ctx, s.addr2, l.Key, []*ledger.LedgerEntry{entries[1]})
	s.Require().NoError(err, "AppendEntry error for sequence 2")
	allEntries, err := s.keeper.ListLedgerEntries(s.ctx, l.Key)
	s.Require().NoError(err, "ListLedgerEntries error")
	s.Require().Equal(uint32(2), allEntries[0].Sequence, "sequence number for correlation-id-2")

	// Then add entry with sequence 1
	err = s.keeper.AppendEntries(s.ctx, s.addr2, l.Key, []*ledger.LedgerEntry{entries[0]})
	s.Require().NoError(err, "AppendEntry error for sequence 1")
	allEntries, err = s.keeper.ListLedgerEntries(s.ctx, l.Key)
	s.Require().NoError(err, "ListLedgerEntries error")
	s.Require().Equal(uint32(1), allEntries[0].Sequence, "sequence number for correlation-id-2")

	// Finally add entry with sequence 3
	err = s.keeper.AppendEntries(s.ctx, s.addr2, l.Key, []*ledger.LedgerEntry{entries[2]})
	s.Require().NoError(err, "AppendEntry error for sequence 3")
	allEntries, err = s.keeper.ListLedgerEntries(s.ctx, l.Key)
	s.Require().NoError(err, "ListLedgerEntries error")
	s.Require().Equal(uint32(3), allEntries[2].Sequence, "sequence number for correlation-id-2")

	// Get all entries and verify their sequence numbers
	allEntries, err = s.keeper.ListLedgerEntries(s.ctx, l.Key)
	s.Require().NoError(err, "ListLedgerEntries error")

	// Verify sequence numbers
	s.Require().Len(allEntries, 3, "number of entries")
	s.Require().Equal(uint32(1), allEntries[0].Sequence, "sequence number for correlation-id-1")
	s.Require().Equal(uint32(2), allEntries[1].Sequence, "sequence number for correlation-id-2")
	s.Require().Equal(uint32(3), allEntries[2].Sequence, "sequence number for correlation-id-3")

	// Add another entry with sequence 2 to test sequence number adjustment
	newEntry := ledger.LedgerEntry{
		EntryTypeId:   1,
		PostedDate:    s.pastDate,
		EffectiveDate: s.pastDate,
		Sequence:      2,
		TotalAmt:      math.NewInt(100),
		AppliedAmounts: []*ledger.LedgerBucketAmount{
			{
				AppliedAmt:   math.NewInt(50),
				BucketTypeId: 1,
			},
			{
				AppliedAmt:   math.NewInt(50),
				BucketTypeId: 2,
			},
		},
		CorrelationId: "test-correlation-id-4",
	}

	err = s.keeper.AppendEntries(s.ctx, s.addr2, l.Key, []*ledger.LedgerEntry{&newEntry})
	s.Require().NoError(err, "AppendEntry error for new entry with sequence 2")

	// Get all entries again and verify updated sequence numbers
	allEntries, err = s.keeper.ListLedgerEntries(s.ctx, l.Key)
	s.Require().NoError(err, "ListLedgerEntries error")

	// Verify updated sequence numbers
	s.Require().Len(allEntries, 4, "number of entries after adding new entry")
	s.Require().Equal(uint32(1), allEntries[0].Sequence, "sequence number for correlation-id-1")
	s.Require().Equal(uint32(2), allEntries[1].Sequence, "sequence number for correlation-id-4 (new entry)")
	s.Require().Equal(uint32(3), allEntries[2].Sequence, "sequence number for correlation-id-2 (shifted)")
	s.Require().Equal(uint32(4), allEntries[3].Sequence, "sequence number for correlation-id-3 (shifted)")
}

func (s *TestSuite) TestAppendEntryDuplicateCorrelationId() {
	// Create a test ledger
	l := ledger.Ledger{
		Key: &ledger.LedgerKey{
			AssetClassId: s.validNFTClass.Id,
			NftId:        s.validNFT.Id,
		},
		LedgerClassId: s.validLedgerClass.LedgerClassId,
		StatusTypeId:  1,
	}

	err := s.keeper.CreateLedger(s.ctx, s.addr1, l)
	s.Require().NoError(err, "CreateLedger error")

	// Create a test entry
	entry := ledger.LedgerEntry{
		EntryTypeId:   1,
		PostedDate:    s.pastDate,
		EffectiveDate: s.pastDate,
		Sequence:      1,
		TotalAmt:      math.NewInt(100),
		AppliedAmounts: []*ledger.LedgerBucketAmount{
			{
				AppliedAmt:   math.NewInt(50),
				BucketTypeId: 1,
			},
			{
				AppliedAmt:   math.NewInt(50),
				BucketTypeId: 2,
			},
		},
		CorrelationId: "test-correlation-id-1",
	}

	// Add the entry successfully
	err = s.keeper.AppendEntries(s.ctx, s.addr1, l.Key, []*ledger.LedgerEntry{&entry})
	s.Require().NoError(err, "AppendEntry error for first entry")

	// Try to add the same entry again with the same correlation ID
	err = s.keeper.AppendEntries(s.ctx, s.addr1, l.Key, []*ledger.LedgerEntry{&entry})
	s.Require().Error(err, "AppendEntry should fail for duplicate correlation ID")
	s.Require().Contains(err.Error(), keeper.ErrCodeAlreadyExists, "error should be ErrCodeAlreadyExists")

	// Verify that only one entry exists
	allEntries, err := s.keeper.ListLedgerEntries(s.ctx, l.Key)
	s.Require().NoError(err, "ListLedgerEntries error")
	s.Require().Len(allEntries, 1, "should only have one entry")

	// Try to add a different entry with the same correlation ID
	entry2 := ledger.LedgerEntry{
		EntryTypeId:   1,
		PostedDate:    s.pastDate,
		EffectiveDate: s.pastDate,
		Sequence:      2,
		TotalAmt:      math.NewInt(200),
		AppliedAmounts: []*ledger.LedgerBucketAmount{
			{
				AppliedAmt:   math.NewInt(100),
				BucketTypeId: 1,
			},
			{
				AppliedAmt:   math.NewInt(100),
				BucketTypeId: 2,
			},
		},
		CorrelationId: "test-correlation-id-1",
	}

	err = s.keeper.AppendEntries(s.ctx, s.addr1, l.Key, []*ledger.LedgerEntry{&entry2})
	s.Require().Error(err, "AppendEntry should fail for duplicate correlation ID")
	s.Require().Contains(err.Error(), keeper.ErrCodeAlreadyExists, "error should be ErrCodeAlreadyExists")

	// Verify that still only one entry exists
	allEntries, err = s.keeper.ListLedgerEntries(s.ctx, l.Key)
	s.Require().NoError(err, "ListLedgerEntries error")
	s.Require().Len(allEntries, 1, "should still only have one entry")
	s.Require().Equal(entry.TotalAmt, allEntries[0].TotalAmt, "entry amount should match original entry")
}

func (s *TestSuite) TestGetBalances() {
	// Create a test ledger
	l := ledger.Ledger{
		Key: &ledger.LedgerKey{
			AssetClassId: s.validNFTClass.Id,
			NftId:        s.validNFT.Id,
		},
		LedgerClassId: s.validLedgerClass.LedgerClassId,
		StatusTypeId:  1,
	}

	err := s.keeper.CreateLedger(s.ctx, s.addr1, l)
	s.Require().NoError(err, "CreateLedger error")

	// Create a test entry
	entries := []*ledger.LedgerEntry{
		{
			// Disbursement
			EntryTypeId:   2,
			PostedDate:    s.pastDate,
			EffectiveDate: s.pastDate,
			Sequence:      1,
			TotalAmt:      math.NewInt(1000),
			AppliedAmounts: []*ledger.LedgerBucketAmount{
				{
					// Principal
					BucketTypeId: 1,
					AppliedAmt:   math.NewInt(1000),
				},
			},
			BucketBalances: map[int32]*ledger.BucketBalance{
				1: {
					BucketTypeId: 1,
					Balance:      math.NewInt(1000),
				},
			},
			CorrelationId: "test-correlation-id-1",
		},
		{
			// Origination Fee
			EntryTypeId:   3,
			PostedDate:    s.pastDate,
			EffectiveDate: s.pastDate,
			Sequence:      1,
			TotalAmt:      math.NewInt(10),
			AppliedAmounts: []*ledger.LedgerBucketAmount{
				{
					// Principal
					BucketTypeId: 1,
					AppliedAmt:   math.NewInt(10),
				},
			},
			BucketBalances: map[int32]*ledger.BucketBalance{
				1: {
					BucketTypeId: 1,
					Balance:      math.NewInt(10),
				},
			},
			CorrelationId: "test-correlation-id-2",
		},

		{
			// Payment
			EntryTypeId:   1,
			PostedDate:    s.pastDate,
			EffectiveDate: s.pastDate,
			Sequence:      1,
			TotalAmt:      math.NewInt(500),
			AppliedAmounts: []*ledger.LedgerBucketAmount{
				{
					// Principal
					BucketTypeId: 1,
					AppliedAmt:   math.NewInt(-100),
				},
				{
					// Interest
					BucketTypeId: 2,
					AppliedAmt:   math.NewInt(-300),
				},
				{
					// Escrow
					BucketTypeId: 3,
					AppliedAmt:   math.NewInt(100),
				},
			},
			BucketBalances: map[int32]*ledger.BucketBalance{
				1: {
					BucketTypeId: 1,
					Balance:      math.NewInt(910),
				},
				2: {
					BucketTypeId: 2,
					Balance:      math.NewInt(-300),
				},
				3: {
					BucketTypeId: 3,
					Balance:      math.NewInt(100),
				},
			},
			CorrelationId: "test-correlation-id-3",
		},
	}

	// Add entries to the ledger
	err = s.keeper.AppendEntries(s.ctx, s.addr1, l.Key, entries)
	s.Require().NoError(err, "AppendEntries error")

	entries, err = s.keeper.ListLedgerEntries(s.ctx, l.Key)
	s.Require().NoError(err, "ListLedgerEntries error")
	s.Require().Equal(3, len(entries), "number of entries")

	s.Require().Less(s.pastDate, keeper.DaysSinceEpoch(time.Now().UTC()))

	// Get balances
	balances, err := s.keeper.GetBalancesAsOf(s.ctx, l.Key, time.Now().UTC())
	s.Require().NoError(err, "GetBalances error")
	s.Require().Equal(3, len(balances.BucketBalances), "number of bucket balances")

	s.Require().Equal(math.NewInt(910), balances.BucketBalances[0].Balance)
	s.Require().Equal(math.NewInt(-300), balances.BucketBalances[1].Balance)
	s.Require().Equal(math.NewInt(100), balances.BucketBalances[2].Balance)
}

func (s *TestSuite) TestBech32() {
	ledgerKey := &ledger.LedgerKey{
		AssetClassId: s.validNFTClass.Id,
		NftId:        s.addr1.String(),
	}

	expectedBech32Str := "ledger1w3jhxapdden8gttrd3shxuedd9jr5cm0wdkk7ue3x44hjwtyw5uxzvnhd3ehg73kvec8svmsx3khzur209ex6dtrvack5amv8pehzv7wxj4"

	bech32Id, err := keeper.LedgerKeyToString(ledgerKey)
	s.Require().NoError(err, "LedgerKeyToString error")
	s.Require().Equal(expectedBech32Str, *bech32Id)

	ledgerKey2, err := keeper.StringToLedgerKey(*bech32Id)
	s.Require().NoError(err, "StringToLedgerKey error")
	s.Require().Equal(ledgerKey, ledgerKey2, "ledger keys should be equal")

	_, err = keeper.StringToLedgerKey("ledgerasdf1w3jhxapdden8gttrd3shxuedd9jr5cm0wdkk7ue3x44hjwtyw5uxzvnhd3ehg73kvec8svmsx3khzur209ex6dtrvack5amv8pehzv7wxj4")
	s.Require().Error(err, "StringToLedgerKey error")
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
