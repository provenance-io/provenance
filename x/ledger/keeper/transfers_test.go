package keeper_test

import (
	"cosmossdk.io/math"

	"github.com/provenance-io/provenance/x/ledger/helper"
	ledger "github.com/provenance-io/provenance/x/ledger/types"
)

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

	err := s.keeper.AddLedger(s.ctx, l)
	s.Require().NoError(err, "CreateLedger error")

	// Create a test entry
	entries := []*ledger.LedgerEntry{
		{
			// Disbursement
			EntryTypeId:   2,
			PostedDate:    s.pastDate,
			EffectiveDate: s.pastDate,
			Sequence:      3,
			TotalAmt:      math.NewInt(1000),
			AppliedAmounts: []*ledger.LedgerBucketAmount{
				{
					// Principal
					BucketTypeId: 1,
					AppliedAmt:   math.NewInt(1000),
				},
			},
			// This entry has sequence 3, so this balance amount is used for type 1.
			BalanceAmounts: []*ledger.BucketBalance{
				{
					BucketTypeId: 1,
					BalanceAmt:   math.NewInt(1000),
				},
			},
			CorrelationId: "test-correlation-id-3",
		},
		{
			// Origination Fee
			EntryTypeId:   3,
			PostedDate:    s.pastDate,
			EffectiveDate: s.pastDate,
			Sequence:      2,
			TotalAmt:      math.NewInt(10),
			AppliedAmounts: []*ledger.LedgerBucketAmount{
				{
					// Principal
					BucketTypeId: 1,
					AppliedAmt:   math.NewInt(10),
				},
			},
			// This entry has sequence 2, so bucket type 1 here is overwritten by the sequence 3 entry.
			BalanceAmounts: []*ledger.BucketBalance{
				{
					BucketTypeId: 1,
					BalanceAmt:   math.NewInt(10),
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
			// This entry has sequence 1, so bucket type 1 is overwritten, but 2 and 3 are.
			BalanceAmounts: []*ledger.BucketBalance{
				{
					BucketTypeId: 1,
					BalanceAmt:   math.NewInt(910),
				},
				{
					BucketTypeId: 2,
					BalanceAmt:   math.NewInt(-300),
				},
				{
					BucketTypeId: 3,
					BalanceAmt:   math.NewInt(100),
				},
			},
			CorrelationId: "test-correlation-id-1",
		},
	}
	exp := []*ledger.BucketBalance{
		entries[0].BalanceAmounts[0],
		entries[2].BalanceAmounts[1],
		entries[2].BalanceAmounts[2],
	}

	// Add entries to the ledger
	err = s.keeper.AppendEntries(s.ctx, l.Key, entries)
	s.Require().NoError(err, "AppendEntries error")

	entries, err = s.keeper.ListLedgerEntries(s.ctx, l.Key)
	s.Require().NoError(err, "ListLedgerEntries error")
	s.Require().Equal(3, len(entries), "number of entries")

	now := s.curDT
	s.Require().Less(s.pastDate, helper.DaysSinceEpoch(now))

	// Get balances
	balances, err := s.keeper.GetBalancesAsOf(s.ctx, l.Key, s.curDT)
	s.Require().NoError(err, "GetBalancesAsOf error")
	s.Assert().Equal(exp, balances, "GetBalancesAsOf balances")
}
