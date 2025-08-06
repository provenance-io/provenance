package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/math"
	"cosmossdk.io/x/nft"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/ledger/helper"
	"github.com/provenance-io/provenance/x/ledger/keeper"
	ledger "github.com/provenance-io/provenance/x/ledger/types"
)

func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerTestSuite))
}

type MsgServerTestSuite struct {
	suite.Suite
	app *app.App
	ctx sdk.Context

	keeper           keeper.Keeper
	bondDenom        string
	pastDate         int32
	validLedgerClass ledger.LedgerClass
	validNFTClass    nft.Class
	validNFT         nft.NFT
}

func (s *MsgServerTestSuite) SetupTest() {
	s.app = app.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(true)
	s.keeper = s.app.LedgerKeeper

	var err error
	s.bondDenom, err = s.app.StakingKeeper.BondDenom(s.ctx)
	s.Require().NoError(err, "app.StakingKeeper.BondDenom(s.ctx)")

	// Create a timestamp 24 hours in the past to avoid future date errors
	s.pastDate = helper.DaysSinceEpoch(time.Now().Add(-24 * time.Hour).UTC())

	// Load the test ledger class configs
	s.ConfigureTest()
}

func (s *MsgServerTestSuite) ConfigureTest() {
	s.ctx = s.ctx.WithBlockTime(time.Now())

	s.validNFTClass = nft.Class{
		Id: "test-nft-class-id",
	}
	s.app.NFTKeeper.SaveClass(s.ctx, s.validNFTClass)

	s.validNFT = nft.NFT{
		ClassId: s.validNFTClass.Id,
		Id:      "test-nft-id",
	}
	s.app.NFTKeeper.Mint(s.ctx, s.validNFT, s.app.AccountKeeper.GetModuleAddress("distribution"))

	s.validLedgerClass = ledger.LedgerClass{
		LedgerClassId:     "test-ledger-class-id",
		AssetClassId:      s.validNFTClass.Id,
		MaintainerAddress: s.app.AccountKeeper.GetModuleAddress("distribution").String(),
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

// TestAppendEntriesValidation tests validation logic for AppendEntries
func (s *MsgServerTestSuite) TestAppendEntriesValidation() {
	// Create a test ledger first
	l := ledger.Ledger{
		Key: &ledger.LedgerKey{
			AssetClassId: s.validNFTClass.Id,
			NftId:        s.validNFT.Id,
		},
		LedgerClassId: s.validLedgerClass.LedgerClassId,
		StatusTypeId:  1,
	}

	err := s.keeper.AddLedger(s.ctx, l)
	s.Require().NoError(err, "CreateLedger error")

	tests := []struct {
		name    string
		entries []*ledger.LedgerEntry
		expErr  string
	}{
		{
			name: "future posted date",
			entries: []*ledger.LedgerEntry{
				{
					EntryTypeId:   1,
					PostedDate:    helper.DaysSinceEpoch(time.Now().Add(24 * time.Hour).UTC()),
					EffectiveDate: s.pastDate,
					TotalAmt:      math.NewInt(100),
					AppliedAmounts: []*ledger.LedgerBucketAmount{
						{
							AppliedAmt:   math.NewInt(100),
							BucketTypeId: 1,
						},
					},
					CorrelationId: "test-correlation-id-future",
				},
			},
			expErr: "posted_date cannot be in the future",
		},
		{
			name: "invalid entry type id",
			entries: []*ledger.LedgerEntry{
				{
					EntryTypeId:   999, // Non-existent entry type
					PostedDate:    s.pastDate,
					EffectiveDate: s.pastDate,
					TotalAmt:      math.NewInt(100),
					AppliedAmounts: []*ledger.LedgerBucketAmount{
						{
							AppliedAmt:   math.NewInt(100),
							BucketTypeId: 1,
						},
					},
					CorrelationId: "test-correlation-id-invalid-type",
				},
			},
			expErr: "entry_type_id entry type doesn't exist",
		},
		{
			name: "non-existent ledger",
			entries: []*ledger.LedgerEntry{
				{
					EntryTypeId:   1,
					PostedDate:    s.pastDate,
					EffectiveDate: s.pastDate,
					TotalAmt:      math.NewInt(100),
					AppliedAmounts: []*ledger.LedgerBucketAmount{
						{
							AppliedAmt:   math.NewInt(100),
							BucketTypeId: 1,
						},
					},
					CorrelationId: "test-correlation-id-non-existent",
				},
			},
			expErr: "ledger not found",
		},
		{
			name: "valid entry",
			entries: []*ledger.LedgerEntry{
				{
					EntryTypeId:   1,
					PostedDate:    s.pastDate,
					EffectiveDate: s.pastDate,
					TotalAmt:      math.NewInt(100),
					AppliedAmounts: []*ledger.LedgerBucketAmount{
						{
							AppliedAmt:   math.NewInt(100),
							BucketTypeId: 1,
						},
					},
					CorrelationId: "test-correlation-id-valid",
				},
			},
			expErr: "",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			ledgerKey := l.Key
			if tc.name == "non-existent ledger" {
				ledgerKey = &ledger.LedgerKey{
					AssetClassId: s.validNFTClass.Id,
					NftId:        "non-existent-nft",
				}
			}

			err := s.keeper.AppendEntries(s.ctx, ledgerKey, tc.entries)
			if tc.expErr != "" {
				s.Require().Error(err, "AppendEntries should fail")
				s.Require().Contains(err.Error(), tc.expErr, "error message")
			} else {
				s.Require().NoError(err, "AppendEntries should succeed")
			}
		})
	}
}

// TestUpdateEntryBalancesValidation tests validation logic for UpdateEntryBalances
func (s *MsgServerTestSuite) TestUpdateEntryBalancesValidation() {
	// Create a test ledger and entry
	l := ledger.Ledger{
		Key: &ledger.LedgerKey{
			AssetClassId: s.validNFTClass.Id,
			NftId:        s.validNFT.Id,
		},
		LedgerClassId: s.validLedgerClass.LedgerClassId,
		StatusTypeId:  1,
	}

	err := s.keeper.AddLedger(s.ctx, l)
	s.Require().NoError(err, "CreateLedger error")

	// Add an entry first
	entry := ledger.LedgerEntry{
		EntryTypeId:   1,
		PostedDate:    s.pastDate,
		EffectiveDate: s.pastDate,
		TotalAmt:      math.NewInt(100),
		AppliedAmounts: []*ledger.LedgerBucketAmount{
			{
				AppliedAmt:   math.NewInt(100),
				BucketTypeId: 1,
			},
		},
		CorrelationId: "test-correlation-id-update",
	}

	err = s.keeper.AppendEntries(s.ctx, l.Key, []*ledger.LedgerEntry{&entry})
	s.Require().NoError(err, "AppendEntry error")

	tests := []struct {
		name           string
		ledgerKey      *ledger.LedgerKey
		correlationId  string
		balanceAmounts []*ledger.BucketBalance
		appliedAmounts []*ledger.LedgerBucketAmount
		expErr         string
	}{
		{
			name: "non-existent entry",
			ledgerKey: &ledger.LedgerKey{
				AssetClassId: s.validNFTClass.Id,
				NftId:        s.validNFT.Id,
			},
			correlationId: "non-existent-correlation-id",
			balanceAmounts: []*ledger.BucketBalance{
				{
					BucketTypeId: 1,
					BalanceAmt:   math.NewInt(200),
				},
			},
			appliedAmounts: []*ledger.LedgerBucketAmount{
				{
					AppliedAmt:   math.NewInt(200),
					BucketTypeId: 1,
				},
			},
			expErr: "entry not found",
		},
		{
			name:          "valid update",
			ledgerKey:     l.Key,
			correlationId: "test-correlation-id-update",
			balanceAmounts: []*ledger.BucketBalance{
				{
					BucketTypeId: 1,
					BalanceAmt:   math.NewInt(200),
				},
			},
			appliedAmounts: []*ledger.LedgerBucketAmount{
				{
					AppliedAmt:   math.NewInt(200),
					BucketTypeId: 1,
				},
			},
			expErr: "",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := s.keeper.UpdateEntryBalances(s.ctx, tc.ledgerKey, tc.correlationId, tc.balanceAmounts, tc.appliedAmounts)
			if tc.expErr != "" {
				s.Require().Error(err, "UpdateEntryBalances should fail")
				s.Require().Contains(err.Error(), tc.expErr, "error message")
			} else {
				s.Require().NoError(err, "UpdateEntryBalances should succeed")
			}
		})
	}
}

// TestAppendEntriesMultipleValidation tests validation for multiple entries
func (s *MsgServerTestSuite) TestAppendEntriesMultipleValidation() {
	// Create a test ledger
	l := ledger.Ledger{
		Key: &ledger.LedgerKey{
			AssetClassId: s.validNFTClass.Id,
			NftId:        s.validNFT.Id,
		},
		LedgerClassId: s.validLedgerClass.LedgerClassId,
		StatusTypeId:  1,
	}

	err := s.keeper.AddLedger(s.ctx, l)
	s.Require().NoError(err, "CreateLedger error")

	tests := []struct {
		name    string
		entries []*ledger.LedgerEntry
		expErr  string
	}{
		{
			name: "mixed valid and invalid entries",
			entries: []*ledger.LedgerEntry{
				{
					EntryTypeId:   1,
					PostedDate:    s.pastDate,
					EffectiveDate: s.pastDate,
					TotalAmt:      math.NewInt(100),
					AppliedAmounts: []*ledger.LedgerBucketAmount{
						{
							AppliedAmt:   math.NewInt(100),
							BucketTypeId: 1,
						},
					},
					CorrelationId: "test-correlation-id-valid-1",
				},
				{
					EntryTypeId:   999, // Invalid entry type
					PostedDate:    s.pastDate,
					EffectiveDate: s.pastDate,
					TotalAmt:      math.NewInt(200),
					AppliedAmounts: []*ledger.LedgerBucketAmount{
						{
							AppliedAmt:   math.NewInt(200),
							BucketTypeId: 1,
						},
					},
					CorrelationId: "test-correlation-id-invalid-1",
				},
			},
			expErr: "entry_type_id entry type doesn't exist",
		},
		{
			name: "all valid entries",
			entries: []*ledger.LedgerEntry{
				{
					EntryTypeId:   1,
					PostedDate:    s.pastDate,
					EffectiveDate: s.pastDate,
					TotalAmt:      math.NewInt(100),
					AppliedAmounts: []*ledger.LedgerBucketAmount{
						{
							AppliedAmt:   math.NewInt(100),
							BucketTypeId: 1,
						},
					},
					CorrelationId: "test-correlation-id-valid-2",
				},
				{
					EntryTypeId:   2,
					PostedDate:    s.pastDate,
					EffectiveDate: s.pastDate,
					TotalAmt:      math.NewInt(200),
					AppliedAmounts: []*ledger.LedgerBucketAmount{
						{
							AppliedAmt:   math.NewInt(200),
							BucketTypeId: 1,
						},
					},
					CorrelationId: "test-correlation-id-valid-3",
				},
			},
			expErr: "",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := s.keeper.AppendEntries(s.ctx, l.Key, tc.entries)
			if tc.expErr != "" {
				s.Require().Error(err, "AppendEntries should fail")
				s.Require().Contains(err.Error(), tc.expErr, "error message")
			} else {
				s.Require().NoError(err, "AppendEntries should succeed")
			}
		})
	}
}

// TestAppendEntriesEmptyArray tests edge case of empty entries array
func (s *MsgServerTestSuite) TestAppendEntriesEmptyArray() {
	// Create a test ledger
	l := ledger.Ledger{
		Key: &ledger.LedgerKey{
			AssetClassId: s.validNFTClass.Id,
			NftId:        s.validNFT.Id,
		},
		LedgerClassId: s.validLedgerClass.LedgerClassId,
		StatusTypeId:  1,
	}

	err := s.keeper.AddLedger(s.ctx, l)
	s.Require().NoError(err, "CreateLedger error")

	// Test appending empty array
	err = s.keeper.AppendEntries(s.ctx, l.Key, []*ledger.LedgerEntry{})
	s.Require().NoError(err, "AppendEntries with empty array should succeed")
}

// TestAppendEntriesInvalidLedgerKey tests with malformed ledger key
func (s *MsgServerTestSuite) TestAppendEntriesInvalidLedgerKey() {
	entry := ledger.LedgerEntry{
		EntryTypeId:   1,
		PostedDate:    s.pastDate,
		EffectiveDate: s.pastDate,
		TotalAmt:      math.NewInt(100),
		AppliedAmounts: []*ledger.LedgerBucketAmount{
			{
				AppliedAmt:   math.NewInt(100),
				BucketTypeId: 1,
			},
		},
		CorrelationId: "test-correlation-id-invalid-key",
	}

	// Test with nil ledger key
	err := s.keeper.AppendEntries(s.ctx, nil, []*ledger.LedgerEntry{&entry})
	s.Require().Error(err, "AppendEntries with nil ledger key should fail")

	// Test with empty ledger key
	emptyKey := &ledger.LedgerKey{
		AssetClassId: "",
		NftId:        "",
	}
	err = s.keeper.AppendEntries(s.ctx, emptyKey, []*ledger.LedgerEntry{&entry})
	s.Require().Error(err, "AppendEntries with empty ledger key should fail")
}
