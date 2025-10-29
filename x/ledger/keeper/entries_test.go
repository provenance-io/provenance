package keeper_test

import (
	"errors"

	"cosmossdk.io/math"

	ledger "github.com/provenance-io/provenance/x/ledger/types"
)

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

	err := s.keeper.AddLedger(s.ctx, l)
	s.Require().NoError(err, "CreateLedger error")

	// Test cases
	tests := []struct {
		name   string
		key    *ledger.LedgerKey
		entry  ledger.LedgerEntry
		expErr string // empty string = no error expected
	}{
		{
			name: "valid amounts and balances",
			key: &ledger.LedgerKey{
				AssetClassId: s.validNFTClass.Id,
				NftId:        s.validNFT.Id,
			},
			entry: ledger.LedgerEntry{
				CorrelationId: "test-correlation-id-12",
				EntryTypeId:   1,
				PostedDate:    s.pastDate,
				EffectiveDate: s.pastDate,
				TotalAmt:      math.NewInt(100),
				AppliedAmounts: []*ledger.LedgerBucketAmount{
					{AppliedAmt: math.NewInt(50), BucketTypeId: 1},
					{AppliedAmt: math.NewInt(50), BucketTypeId: 2},
				},
				BalanceAmounts: []*ledger.BucketBalance{{BucketTypeId: 1, BalanceAmt: math.NewInt(9950)}},
			},
			expErr: "",
		},
		{
			name: "valid amounts and balances with negative applied amount",
			key: &ledger.LedgerKey{
				AssetClassId: s.validNFTClass.Id,
				NftId:        s.validNFT.Id,
			},
			entry: ledger.LedgerEntry{
				CorrelationId: "test-correlation-id-13",
				EntryTypeId:   1,
				PostedDate:    s.pastDate,
				EffectiveDate: s.pastDate,
				TotalAmt:      math.NewInt(0),
				AppliedAmounts: []*ledger.LedgerBucketAmount{
					{AppliedAmt: math.NewInt(50), BucketTypeId: 1},
					{AppliedAmt: math.NewInt(-50), BucketTypeId: 2},
				},
				BalanceAmounts: []*ledger.BucketBalance{{BucketTypeId: 1, BalanceAmt: math.NewInt(10000)}},
			},
			expErr: "",
		},
		{
			name: "allow negative principal applied amount",
			key: &ledger.LedgerKey{
				AssetClassId: s.validNFTClass.Id,
				NftId:        s.validNFT.Id,
			},
			entry: ledger.LedgerEntry{
				CorrelationId: "test-correlation-id-15",
				EntryTypeId:   1,
				PostedDate:    s.pastDate,
				EffectiveDate: s.pastDate,
				TotalAmt:      math.NewInt(100),
				AppliedAmounts: []*ledger.LedgerBucketAmount{
					{AppliedAmt: math.NewInt(150), BucketTypeId: 1},
					{AppliedAmt: math.NewInt(-50), BucketTypeId: 2},
				},
				BalanceAmounts: []*ledger.BucketBalance{{BucketTypeId: 1, BalanceAmt: math.NewInt(10000)}},
			},
			expErr: "",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err = s.keeper.AppendEntries(s.ctx, tc.key, []*ledger.LedgerEntry{&tc.entry})
			s.assertErrorValue(err, tc.expErr, "AppendEntry")
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

	err := s.keeper.AddLedger(s.ctx, l)
	s.Require().NoError(err, "CreateLedger error")

	// Create test entries with the same effective date but different sequence numbers
	entries := []*ledger.LedgerEntry{
		{
			CorrelationId: "test-correlation-id-1",
			EntryTypeId:   1,
			PostedDate:    s.pastDate,
			EffectiveDate: s.pastDate,
			Sequence:      1,
			TotalAmt:      math.NewInt(100),
			AppliedAmounts: []*ledger.LedgerBucketAmount{
				{AppliedAmt: math.NewInt(50), BucketTypeId: 1},
				{AppliedAmt: math.NewInt(50), BucketTypeId: 2},
			},
			BalanceAmounts: []*ledger.BucketBalance{{BucketTypeId: 1, BalanceAmt: math.NewInt(10000)}},
		},
		{
			CorrelationId: "test-correlation-id-2",
			EntryTypeId:   1,
			PostedDate:    s.pastDate,
			EffectiveDate: s.pastDate,
			Sequence:      2,
			TotalAmt:      math.NewInt(100),
			AppliedAmounts: []*ledger.LedgerBucketAmount{
				{AppliedAmt: math.NewInt(50), BucketTypeId: 1},
				{AppliedAmt: math.NewInt(50), BucketTypeId: 2},
			},
			BalanceAmounts: []*ledger.BucketBalance{{BucketTypeId: 1, BalanceAmt: math.NewInt(12345)}},
		},
		{
			CorrelationId: "test-correlation-id-3",
			EntryTypeId:   1,
			PostedDate:    s.pastDate,
			EffectiveDate: s.pastDate,
			Sequence:      3,
			TotalAmt:      math.NewInt(100),
			AppliedAmounts: []*ledger.LedgerBucketAmount{
				{AppliedAmt: math.NewInt(50), BucketTypeId: 1},
				{AppliedAmt: math.NewInt(50), BucketTypeId: 2},
			},
			BalanceAmounts: []*ledger.BucketBalance{{BucketTypeId: 1, BalanceAmt: math.NewInt(99950)}},
		},
	}

	// Add entries in a specific order to test sequence number adjustment
	// First add entry with sequence 2
	err = s.keeper.AppendEntries(s.ctx, l.Key, []*ledger.LedgerEntry{entries[1]})
	s.Require().NoError(err, "AppendEntry error for sequence 2")
	allEntries, err := s.keeper.ListLedgerEntries(s.ctx, l.Key)
	s.Require().NoError(err, "ListLedgerEntries error")
	s.Require().Equal(uint32(2), allEntries[0].Sequence, "sequence number for correlation-id-2")

	// Then add entry with sequence 1
	err = s.keeper.AppendEntries(s.ctx, l.Key, []*ledger.LedgerEntry{entries[0]})
	s.Require().NoError(err, "AppendEntry error for sequence 1")
	allEntries, err = s.keeper.ListLedgerEntries(s.ctx, l.Key)
	s.Require().NoError(err, "ListLedgerEntries error")
	s.Require().Equal(uint32(1), allEntries[0].Sequence, "sequence number for correlation-id-2")

	// Finally add entry with sequence 3
	err = s.keeper.AppendEntries(s.ctx, l.Key, []*ledger.LedgerEntry{entries[2]})
	s.Require().NoError(err, "AppendEntry error for sequence 3")
	allEntries, err = s.keeper.ListLedgerEntries(s.ctx, l.Key)
	s.Require().NoError(err, "ListLedgerEntries error")
	s.Require().Equal(uint32(3), allEntries[2].Sequence, "sequence number for correlation-id-2")

	// Get all entries and verify their sequence numbers
	allEntries, err = s.keeper.ListLedgerEntries(s.ctx, l.Key)
	s.Require().NoError(err, "ListLedgerEntries error")

	// Verify sequence numbers
	s.Require().Len(allEntries, 3, "number of entries")
	s.Assert().Equal(1, int(allEntries[0].Sequence), "sequence number for correlation-id-1")
	s.Assert().Equal(2, int(allEntries[1].Sequence), "sequence number for correlation-id-2")
	s.Assert().Equal(3, int(allEntries[2].Sequence), "sequence number for correlation-id-3")

	// Add another entry with sequence 2 to test sequence number adjustment
	newEntry := ledger.LedgerEntry{
		CorrelationId: "test-correlation-id-4",
		EntryTypeId:   1,
		PostedDate:    s.pastDate,
		EffectiveDate: s.pastDate,
		Sequence:      2,
		TotalAmt:      math.NewInt(100),
		AppliedAmounts: []*ledger.LedgerBucketAmount{
			{AppliedAmt: math.NewInt(50), BucketTypeId: 1},
			{AppliedAmt: math.NewInt(50), BucketTypeId: 2},
		},
		BalanceAmounts: []*ledger.BucketBalance{{BucketTypeId: 1, BalanceAmt: math.NewInt(99950)}},
	}

	err = s.keeper.AppendEntries(s.ctx, l.Key, []*ledger.LedgerEntry{&newEntry})
	s.Require().NoError(err, "AppendEntry error for new entry with sequence 2")

	// Get all entries again and verify updated sequence numbers
	allEntries, err = s.keeper.ListLedgerEntries(s.ctx, l.Key)
	s.Require().NoError(err, "ListLedgerEntries error")

	// Verify updated sequence numbers
	s.Require().Len(allEntries, 4, "number of entries after adding new entry")
	s.Assert().Equal(1, int(allEntries[0].Sequence), "sequence number for correlation-id-1")
	s.Assert().Equal(2, int(allEntries[1].Sequence), "sequence number for correlation-id-4 (new entry)")
	s.Assert().Equal(3, int(allEntries[2].Sequence), "sequence number for correlation-id-2 (shifted)")
	s.Assert().Equal(4, int(allEntries[3].Sequence), "sequence number for correlation-id-3 (shifted)")
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

	err := s.keeper.AddLedger(s.ctx, l)
	s.Require().NoError(err, "CreateLedger error")

	// Create a test entry
	entry := ledger.LedgerEntry{
		CorrelationId: "test-correlation-id-1",
		EntryTypeId:   1,
		PostedDate:    s.pastDate,
		EffectiveDate: s.pastDate,
		Sequence:      1,
		TotalAmt:      math.NewInt(100),
		AppliedAmounts: []*ledger.LedgerBucketAmount{
			{AppliedAmt: math.NewInt(50), BucketTypeId: 1},
			{AppliedAmt: math.NewInt(50), BucketTypeId: 2},
		},
		BalanceAmounts: []*ledger.BucketBalance{{BucketTypeId: 1, BalanceAmt: math.NewInt(10000)}},
	}

	// Add the entry successfully
	err = s.keeper.AppendEntries(s.ctx, l.Key, []*ledger.LedgerEntry{&entry})
	s.Require().NoError(err, "AppendEntry error for first entry")

	// Try to add the same entry again with the same correlation ID
	err = s.keeper.AppendEntries(s.ctx, l.Key, []*ledger.LedgerEntry{&entry})
	s.Require().Error(err, "AppendEntry should fail for duplicate correlation ID")
	s.Require().True(errors.Is(err, ledger.ErrAlreadyExists), "error should be ErrAlreadyExists")

	// Verify that only one entry exists
	allEntries, err := s.keeper.ListLedgerEntries(s.ctx, l.Key)
	s.Require().NoError(err, "ListLedgerEntries error")
	s.Require().Len(allEntries, 1, "should only have one entry")

	// Try to add a different entry with the same correlation ID
	entry2 := ledger.LedgerEntry{
		CorrelationId: "test-correlation-id-1",
		EntryTypeId:   1,
		PostedDate:    s.pastDate,
		EffectiveDate: s.pastDate,
		Sequence:      2,
		TotalAmt:      math.NewInt(200),
		AppliedAmounts: []*ledger.LedgerBucketAmount{
			{AppliedAmt: math.NewInt(100), BucketTypeId: 1},
			{AppliedAmt: math.NewInt(100), BucketTypeId: 2},
		},
		BalanceAmounts: []*ledger.BucketBalance{{BucketTypeId: 1, BalanceAmt: math.NewInt(99950)}},
	}

	err = s.keeper.AppendEntries(s.ctx, l.Key, []*ledger.LedgerEntry{&entry2})
	s.Require().Error(err, "AppendEntry should fail for duplicate correlation ID")
	s.Require().True(errors.Is(err, ledger.ErrAlreadyExists), "error should be ErrAlreadyExists")

	// Verify that still only one entry exists
	allEntries, err = s.keeper.ListLedgerEntries(s.ctx, l.Key)
	s.Require().NoError(err, "ListLedgerEntries error")
	s.Require().Len(allEntries, 1, "should still only have one entry")
	s.Require().Equal(entry.TotalAmt, allEntries[0].TotalAmt, "entry amount should match original entry")
}

// TestRequireGetLedgerEntry tests the RequireGetLedgerEntry function
// This function should return the ledger entry if it exists, or an error if it doesn't exist
func (s *TestSuite) TestRequireGetLedgerEntry() {
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
	entry := ledger.LedgerEntry{
		CorrelationId: "test-correlation-id-require",
		EntryTypeId:   1,
		PostedDate:    s.pastDate,
		EffectiveDate: s.pastDate,
		Sequence:      1,
		TotalAmt:      math.NewInt(100),
		AppliedAmounts: []*ledger.LedgerBucketAmount{
			{AppliedAmt: math.NewInt(50), BucketTypeId: 1},
			{AppliedAmt: math.NewInt(50), BucketTypeId: 2},
		},
		BalanceAmounts: []*ledger.BucketBalance{{BucketTypeId: 1, BalanceAmt: math.NewInt(10000)}},
	}

	// Add the entry to the ledger
	err = s.keeper.AppendEntries(s.ctx, l.Key, []*ledger.LedgerEntry{&entry})
	s.Require().NoError(err, "AppendEntry error")

	tests := []struct {
		name          string
		key           *ledger.LedgerKey
		correlationId string
		expEntry      *ledger.LedgerEntry
		expErr        error
	}{
		{
			name: "existing entry should be retrieved successfully",
			key: &ledger.LedgerKey{
				AssetClassId: s.validNFTClass.Id,
				NftId:        s.validNFT.Id,
			},
			correlationId: "test-correlation-id-require",
			expEntry:      &entry,
			expErr:        nil,
		},
		{
			name: "non-existent entry should return error",
			key: &ledger.LedgerKey{
				AssetClassId: s.validNFTClass.Id,
				NftId:        s.validNFT.Id,
			},
			correlationId: "non-existent-correlation-id",
			expEntry:      nil,
			expErr:        ledger.ErrNotFound,
		},
		{
			name: "non-existent ledger should return error",
			key: &ledger.LedgerKey{
				AssetClassId: s.validNFTClass.Id,
				NftId:        "non-existent-nft-id",
			},
			correlationId: "test-correlation-id-require",
			expEntry:      nil,
			expErr:        ledger.ErrNotFound,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			retrievedEntry, err := s.keeper.RequireGetLedgerEntry(s.ctx, tc.key, tc.correlationId)

			if tc.expErr != nil {
				s.Require().Error(err, "RequireGetLedgerEntry should return error")
				s.Require().ErrorIs(err, tc.expErr, "error should contain expected error code")
				s.Require().Nil(retrievedEntry, "retrieved entry should be nil on error")
			} else {
				s.Require().NoError(err, "RequireGetLedgerEntry should not return error")
				s.Require().NotNil(retrievedEntry, "retrieved entry should not be nil")
				s.Require().Equal(tc.expEntry.CorrelationId, retrievedEntry.CorrelationId, "correlation ID should match")
				s.Require().Equal(tc.expEntry.TotalAmt, retrievedEntry.TotalAmt, "total amount should match")
				s.Require().Equal(tc.expEntry.EntryTypeId, retrievedEntry.EntryTypeId, "entry type ID should match")
				s.Require().Equal(tc.expEntry.PostedDate, retrievedEntry.PostedDate, "posted date should match")
				s.Require().Equal(tc.expEntry.EffectiveDate, retrievedEntry.EffectiveDate, "effective date should match")
				s.Require().Equal(tc.expEntry.Sequence, retrievedEntry.Sequence, "sequence should match")
				s.Require().Len(retrievedEntry.AppliedAmounts, len(tc.expEntry.AppliedAmounts), "applied amounts length should match")
			}
		})
	}
}
