package keeper_test

import (
	"cosmossdk.io/math"

	"github.com/provenance-io/provenance/x/ledger/keeper"
	ledger "github.com/provenance-io/provenance/x/ledger/types"
)

func (s *TestSuite) TestBulkCreate() {
	// Prepare a fresh ledger key and associated data
	newKey := &ledger.LedgerKey{AssetClassId: s.validNFTClass.Id, NftId: "bulk-nft-1"}
	// Mint the NFT for the new ledger
	nft := s.validNFT
	nft.Id = newKey.NftId
	s.nftKeeper.Mint(s.ctx, nft, s.addr1)

	l := &ledger.Ledger{Key: newKey, LedgerClassId: s.validLedgerClass.LedgerClassId, StatusTypeId: 1}
	entry := &ledger.LedgerEntry{
		CorrelationId: "bc-1",
		EntryTypeId:   1,
		PostedDate:    s.pastDate,
		EffectiveDate: s.pastDate,
		TotalAmt:      math.NewInt(100),
		AppliedAmounts: []*ledger.LedgerBucketAmount{{BucketTypeId: 1, AppliedAmt: math.NewInt(100)}},
		BalanceAmounts: []*ledger.BucketBalance{{BucketTypeId: 1, BalanceAmt: math.NewInt(100)}},
	}
	le := &ledger.LedgerAndEntries{LedgerKey: newKey, Ledger: l, Entries: []*ledger.LedgerEntry{entry}}

	// Success
	s.Run("success", func() {
		msgSrv := keeper.NewMsgServer(s.keeper)
		_, err := msgSrv.BulkCreate(s.ctx, &ledger.MsgBulkCreateRequest{Signer: s.addr1.String(), LedgerAndEntries: []*ledger.LedgerAndEntries{le}})
		s.Require().NoError(err)

		// Verify created ledger
		got, err := s.keeper.GetLedger(s.ctx, newKey)
		s.Require().NoError(err)
		s.Require().NotNil(got)

		// Verify entry exists
		ents, err := s.keeper.ListLedgerEntries(s.ctx, newKey)
		s.Require().NoError(err)
		s.Require().Len(ents, 1)
		s.Require().Equal("bc-1", ents[0].CorrelationId)
	})

	// Failure: unauthorized signer
	s.Run("unauthorized", func() {
		anotherKey := &ledger.LedgerKey{AssetClassId: s.validNFTClass.Id, NftId: "bulk-nft-unauth"}
		nft2 := nft
		nft2.Id = anotherKey.NftId
		s.nftKeeper.Mint(s.ctx, nft2, s.addr1)
		l2 := &ledger.Ledger{Key: anotherKey, LedgerClassId: s.validLedgerClass.LedgerClassId, StatusTypeId: 1}
		le2 := &ledger.LedgerAndEntries{LedgerKey: anotherKey, Ledger: l2, Entries: []*ledger.LedgerEntry{entry}}

		msgSrv := keeper.NewMsgServer(s.keeper)
		_, err := msgSrv.BulkCreate(s.ctx, &ledger.MsgBulkCreateRequest{Signer: s.addr2.String(), LedgerAndEntries: []*ledger.LedgerAndEntries{le2}})
		s.Require().Error(err)
	})

	// Failure: invalid entry type
	s.Run("invalid entry type", func() {
		key := &ledger.LedgerKey{AssetClassId: s.validNFTClass.Id, NftId: "bulk-nft-badtype"}
		nft3 := nft
		nft3.Id = key.NftId
		s.nftKeeper.Mint(s.ctx, nft3, s.addr1)
		bad := &ledger.LedgerEntry{
			EntryTypeId:    999,
			PostedDate:     s.pastDate,
			EffectiveDate:  s.pastDate,
			TotalAmt:       math.NewInt(100),
			AppliedAmounts: []*ledger.LedgerBucketAmount{{BucketTypeId: 1, AppliedAmt: math.NewInt(100)}},
			CorrelationId:  "bc-2",
		}
		leBad := &ledger.LedgerAndEntries{LedgerKey: key, Ledger: &ledger.Ledger{Key: key, LedgerClassId: s.validLedgerClass.LedgerClassId, StatusTypeId: 1}, Entries: []*ledger.LedgerEntry{bad}}
		msgSrv := keeper.NewMsgServer(s.keeper)
		_, err := msgSrv.BulkCreate(s.ctx, &ledger.MsgBulkCreateRequest{Signer: s.addr1.String(), LedgerAndEntries: []*ledger.LedgerAndEntries{leBad}})
		s.Require().Error(err)
	})
}
