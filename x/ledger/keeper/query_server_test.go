package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/provenance-io/provenance/x/ledger/keeper"
	ledger "github.com/provenance-io/provenance/x/ledger/types"
)

func (s *TestSuite) TestLedgerQueryServer_Basics() {
	// Ensure there's at least one ledger class from SetupTest
	qs := keeper.NewLedgerQueryServer(s.keeper)

	// LedgerClasses
	respLCs, err := qs.LedgerClasses(s.ctx, &ledger.QueryLedgerClassesRequest{Pagination: &query.PageRequest{CountTotal: true}})
	s.Require().NoError(err)
	s.Require().NotNil(respLCs)
	s.Require().NotNil(respLCs.Pagination)

	// LedgerClass
	respLC, err := qs.LedgerClass(s.ctx, &ledger.QueryLedgerClassRequest{LedgerClassId: s.validLedgerClass.LedgerClassId})
	s.Require().NoError(err)
	s.Require().NotNil(respLC)
	s.Require().NotNil(respLC.LedgerClass)

	// LedgerClassEntryTypes
	respETs, err := qs.LedgerClassEntryTypes(s.ctx, &ledger.QueryLedgerClassEntryTypesRequest{LedgerClassId: s.validLedgerClass.LedgerClassId})
	s.Require().NoError(err)
	s.Require().NotNil(respETs)

	// LedgerClassStatusTypes
	respSTs, err := qs.LedgerClassStatusTypes(s.ctx, &ledger.QueryLedgerClassStatusTypesRequest{LedgerClassId: s.validLedgerClass.LedgerClassId})
	s.Require().NoError(err)
	s.Require().NotNil(respSTs)

	// LedgerClassBucketTypes
	respBTs, err := qs.LedgerClassBucketTypes(s.ctx, &ledger.QueryLedgerClassBucketTypesRequest{LedgerClassId: s.validLedgerClass.LedgerClassId})
	s.Require().NoError(err)
	s.Require().NotNil(respBTs)
}

func (s *TestSuite) TestLedgerQueryServer_LedgerAndEntries() {
	qs := keeper.NewLedgerQueryServer(s.keeper)

	// Create a ledger and entry
	key := &ledger.LedgerKey{AssetClassId: s.validNFTClass.Id, NftId: "qs-nft"}
	nft := s.validNFT
	nft.Id = key.NftId
	s.nftKeeper.Mint(s.ctx, nft, s.addr1)
	s.Require().NoError(s.keeper.AddLedger(s.ctx, ledger.Ledger{Key: key, LedgerClassId: s.validLedgerClass.LedgerClassId, StatusTypeId: 1}))

	// Query ledger
	lresp, err := qs.Ledger(s.ctx, &ledger.QueryLedgerRequest{Key: key})
	s.Require().NoError(err)
	s.Require().NotNil(lresp)
	s.Require().NotNil(lresp.Ledger)

	// Query Ledgers
	lsresp, err := qs.Ledgers(s.ctx, &ledger.QueryLedgersRequest{Pagination: &query.PageRequest{CountTotal: true}})
	s.Require().NoError(err)
	s.Require().NotNil(lsresp)

	// Add an entry and query entries
	entry := &ledger.LedgerEntry{EntryTypeId: 1, PostedDate: s.pastDate, EffectiveDate: s.pastDate, TotalAmt: s.int(100), AppliedAmounts: []*ledger.LedgerBucketAmount{{BucketTypeId: 1, AppliedAmt: s.int(100)}}, CorrelationId: "qs-1"}
	s.Require().NoError(s.keeper.AppendEntries(s.ctx, key, []*ledger.LedgerEntry{entry}))

	ersp, err := qs.LedgerEntries(s.ctx, &ledger.QueryLedgerEntriesRequest{Key: key})
	s.Require().NoError(err)
	s.Require().NotNil(ersp)
	s.Require().GreaterOrEqual(len(ersp.Entries), 1)

	// Query single entry
	eres, err := qs.LedgerEntry(s.ctx, &ledger.QueryLedgerEntryRequest{Key: key, CorrelationId: "qs-1"})
	s.Require().NoError(err)
	s.Require().NotNil(eres)
	s.Require().NotNil(eres.Entry)
}

func (s *TestSuite) TestLedgerQueryServer_BalancesAsOf() {
	qs := keeper.NewLedgerQueryServer(s.keeper)
	key := &ledger.LedgerKey{AssetClassId: s.validNFTClass.Id, NftId: "qs-balances"}
	nft := s.validNFT
	nft.Id = key.NftId
	s.nftKeeper.Mint(s.ctx, nft, s.addr1)
	s.Require().NoError(s.keeper.AddLedger(s.ctx, ledger.Ledger{Key: key, LedgerClassId: s.validLedgerClass.LedgerClassId, StatusTypeId: 1}))

	// add one entry
	entry := &ledger.LedgerEntry{EntryTypeId: 1, PostedDate: s.pastDate, EffectiveDate: s.pastDate, TotalAmt: s.int(100), AppliedAmounts: []*ledger.LedgerBucketAmount{{BucketTypeId: 1, AppliedAmt: s.int(100)}}, CorrelationId: "qs-bal"}
	s.Require().NoError(s.keeper.AppendEntries(s.ctx, key, []*ledger.LedgerEntry{entry}))

	bresp, err := qs.LedgerBalancesAsOf(s.ctx, &ledger.QueryLedgerBalancesAsOfRequest{Key: key, AsOfDate: s.curDateStr})
	s.Require().NoError(err)
	s.Require().NotNil(bresp)
	s.Require().GreaterOrEqual(len(bresp.BucketBalances), 1)
}
