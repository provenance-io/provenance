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
	entryUUID := "b64596bd-76d5-4a04-86db-7568bd295b33"
	entryType := ledger.LedgerEntryType_Disbursement
	postedDate := time.Now()
	effectiveDate := postedDate.Add(24 * time.Hour)
	amount := s.int(1000)

	// Create a valid ledger entry first
	validEntry := ledger.LedgerEntry{
		Uuid:            entryUUID,
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
				Uuid:            entryUUID,
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
			},
			expEvent: true,
		},
		{
			name:    "empty nft address",
			nftAddr: "",
			entry: ledger.LedgerEntry{
				Uuid:            entryUUID,
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
			},
			expErr: []string{"nft_address"},
		},
		{
			name:    "empty entry uuid",
			nftAddr: nftAddr,
			entry: ledger.LedgerEntry{
				Uuid:            "",
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
			},
			expErr: []string{"uuid"},
		},
		{
			name:    "invalid entry uuid format",
			nftAddr: nftAddr,
			entry: ledger.LedgerEntry{
				Uuid:            "invalid-uuid",
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
			},
			expErr: []string{"uuid"},
		},
		{
			name:    "unspecified entry type",
			nftAddr: nftAddr,
			entry: ledger.LedgerEntry{
				Uuid:            entryUUID,
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
			},
			expErr: []string{"type"},
		},
		{
			name:    "zero posted date",
			nftAddr: nftAddr,
			entry: ledger.LedgerEntry{
				Uuid:            entryUUID,
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
			},
			expErr: []string{"posted_date"},
		},
		{
			name:    "zero effective date",
			nftAddr: nftAddr,
			entry: ledger.LedgerEntry{
				Uuid:            entryUUID,
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
			},
			expErr: []string{"effective_date"},
		},
		{
			name:    "negative amount",
			nftAddr: nftAddr,
			entry: ledger.LedgerEntry{
				Uuid:            entryUUID,
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
			},
			expErr: []string{"amount"},
		},
		{
			name:    "negative principal applied amount",
			nftAddr: nftAddr,
			entry: ledger.LedgerEntry{
				Uuid:            entryUUID,
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
			},
			expErr: []string{"principal_applied_amount"},
		},
		{
			name:    "non-existent ledger",
			nftAddr: s.addr2.String(),
			entry: ledger.LedgerEntry{
				Uuid:            entryUUID,
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
			},
			expErr: []string{"not found"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Clear events before each test
			s.ctx.EventManager().Events()

			err := s.keeper.AppendEntry(s.ctx, tc.nftAddr, tc.entry)

			if len(tc.expErr) > 0 {
				s.assertErrorContents(err, tc.expErr, "AppendEntry error")
			} else {
				s.Require().NoError(err, "AppendEntry error")

				// Verify the entry was created
				entry, err := s.keeper.GetLedgerEntry(s.ctx, tc.nftAddr, tc.entry.Uuid)
				s.Require().NoError(err, "GetLedgerEntry error")
				s.Require().NotNil(entry, "GetLedgerEntry result")
				s.Require().Equal(tc.entry.Uuid, entry.Uuid, "entry uuid")
				s.Require().Equal(tc.entry.Type, entry.Type, "entry type")
				s.Require().True(tc.entry.PostedDate.Equal(entry.PostedDate), "entry posted date")
				s.Require().True(tc.entry.EffectiveDate.Equal(entry.EffectiveDate), "entry effective date")
				s.Require().Equal(tc.entry.Amt, entry.Amt, "entry amount")
				s.Require().Equal(tc.entry.PrinAppliedAmt, entry.PrinAppliedAmt, "entry principal applied amount")
				s.Require().Equal(tc.entry.PrinBalAmt, entry.PrinBalAmt, "entry principal balance amount")
				s.Require().Equal(tc.entry.IntAppliedAmt, entry.IntAppliedAmt, "entry interest applied amount")
				s.Require().Equal(tc.entry.IntBalAmt, entry.IntBalAmt, "entry interest balance amount")
				s.Require().Equal(tc.entry.OtherAppliedAmt, entry.OtherAppliedAmt, "entry other applied amount")
				s.Require().Equal(tc.entry.OtherBalAmt, entry.OtherBalAmt, "entry other balance amount")

				// Verify event emission
				if tc.expEvent {
					// Find the expected event
					var foundEvent *sdk.Event
					for _, e := range s.ctx.EventManager().Events() {
						if e.Type == ledger.EventTypeLedgerEntryAdded {
							foundEvent = &e
							break
						}
					}

					s.Require().NotNil(foundEvent)
					s.Require().Equal(ledger.EventTypeLedgerEntryAdded, foundEvent.Type, "event type")
					s.Require().Len(foundEvent.Attributes, 6, "event attributes length")
					s.Require().Equal("nft_address", string(foundEvent.Attributes[0].Key), "event nft address key")
					s.Require().Equal(tc.nftAddr, string(foundEvent.Attributes[0].Value), "event nft address value")
					s.Require().Equal("entry_uuid", string(foundEvent.Attributes[1].Key), "event entry uuid key")
					s.Require().Equal(tc.entry.Uuid, string(foundEvent.Attributes[1].Value), "event entry uuid value")
					s.Require().Equal("entry_type", string(foundEvent.Attributes[2].Key), "event entry type key")
					s.Require().Equal(tc.entry.Type.String(), string(foundEvent.Attributes[2].Value), "event entry type value")
					s.Require().Equal("posted_date", string(foundEvent.Attributes[3].Key), "event posted date key")
					s.Require().Equal(tc.entry.PostedDate.Format(time.RFC3339), string(foundEvent.Attributes[3].Value), "event posted date value")
					s.Require().Equal("effective_date", string(foundEvent.Attributes[4].Key), "event effective date key")
					s.Require().Equal(tc.entry.EffectiveDate.Format(time.RFC3339), string(foundEvent.Attributes[4].Value), "event effective date value")
					s.Require().Equal("amount", string(foundEvent.Attributes[5].Key), "event amount key")
					s.Require().Equal(tc.entry.Amt.String(), string(foundEvent.Attributes[5].Value), "event amount value")
				}
			}
		})
	}
}

// TestGetLedgerEntry tests the GetLedgerEntry function
func (s *TestSuite) TestGetLedgerEntry() {
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
	entryUUID := "b64596bd-76d5-4a04-86db-7568bd295b33"
	entryType := ledger.LedgerEntryType_Disbursement
	postedDate := time.Now()
	effectiveDate := postedDate.Add(24 * time.Hour)
	amount := s.int(1000)

	// Create a valid ledger entry first
	validEntry := ledger.LedgerEntry{
		Uuid:            entryUUID,
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
	}
	err = s.keeper.AppendEntry(s.ctx, nftAddr, validEntry)
	s.Require().NoError(err, "AppendEntry error")

	tests := []struct {
		name     string
		nftAddr  string
		uuid     string
		expErr   []string
		expEntry *ledger.LedgerEntry
		expEvent bool
	}{
		{
			name:     "valid ledger entry retrieval",
			nftAddr:  nftAddr,
			uuid:     entryUUID,
			expEntry: &validEntry,
			expEvent: true,
		},
		{
			name:     "empty nft address",
			nftAddr:  "",
			uuid:     entryUUID,
			expErr:   []string{"nft_address"},
			expEvent: false,
		},
		{
			name:     "empty uuid",
			nftAddr:  nftAddr,
			uuid:     "",
			expErr:   []string{"uuid"},
			expEvent: false,
		},
		{
			name:     "non-existent ledger",
			nftAddr:  s.addr2.String(),
			uuid:     entryUUID,
			expEntry: nil,
			expEvent: false,
		},
		{
			name:     "non-existent entry",
			nftAddr:  nftAddr,
			uuid:     "b64596bd-76d5-4a04-86db-7568bd295b31",
			expEntry: nil,
			expEvent: false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Clear events before each test
			s.ctx.EventManager().Events()

			entry, err := s.keeper.GetLedgerEntry(s.ctx, tc.nftAddr, tc.uuid)

			if len(tc.expErr) > 0 {
				s.assertErrorContents(err, tc.expErr, "GetLedgerEntry error")
				s.Require().Nil(entry, "GetLedgerEntry result should be nil on error")
			} else {
				s.Require().NoError(err, "GetLedgerEntry error")
				if tc.expEntry == nil {
					s.Require().Nil(entry, "GetLedgerEntry result should be nil for non-existent entry")
				} else {
					s.Require().NotNil(entry, "GetLedgerEntry result")
					s.Require().Equal(tc.expEntry.Uuid, entry.Uuid, "entry uuid")
					s.Require().Equal(tc.expEntry.Type, entry.Type, "entry type")
					s.Require().True(tc.expEntry.PostedDate.Equal(entry.PostedDate), "entry posted date")
					s.Require().True(tc.expEntry.EffectiveDate.Equal(entry.EffectiveDate), "entry effective date")
					s.Require().Equal(tc.expEntry.Amt, entry.Amt, "entry amount")
					s.Require().Equal(tc.expEntry.PrinAppliedAmt, entry.PrinAppliedAmt, "entry principal applied amount")
					s.Require().Equal(tc.expEntry.PrinBalAmt, entry.PrinBalAmt, "entry principal balance amount")
					s.Require().Equal(tc.expEntry.IntAppliedAmt, entry.IntAppliedAmt, "entry interest applied amount")
					s.Require().Equal(tc.expEntry.IntBalAmt, entry.IntBalAmt, "entry interest balance amount")
					s.Require().Equal(tc.expEntry.OtherAppliedAmt, entry.OtherAppliedAmt, "entry other applied amount")
					s.Require().Equal(tc.expEntry.OtherBalAmt, entry.OtherBalAmt, "entry other balance amount")
				}
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
	// Create test ledgers
	ledger1 := ledger.Ledger{
		NftAddress: s.addr1.String(),
		Denom:      s.bondDenom,
	}
	ledger2 := ledger.Ledger{
		NftAddress: s.addr2.String(),
		Denom:      s.bondDenom,
	}

	tests := []struct {
		name      string
		genState  *ledger.GenesisState
		expPanic  bool
		expLedger *ledger.Ledger
	}{
		{
			name: "valid genesis state",
			genState: &ledger.GenesisState{
				Ledgers: []ledger.Ledger{ledger1, ledger2},
			},
		},
		{
			name: "empty genesis state",
			genState: &ledger.GenesisState{
				Ledgers: []ledger.Ledger{},
			},
		},
		{
			name: "invalid ledger in genesis state",
			genState: &ledger.GenesisState{
				Ledgers: []ledger.Ledger{
					{
						NftAddress: "invalid",
						Denom:      "testinvaliddenom",
					},
				},
			},
			expPanic: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Clear events before each test
			s.ctx.EventManager().Events()

			if tc.expPanic {
				s.Require().Panics(func() {
					s.keeper.InitGenesis(s.ctx, tc.genState)
				}, "InitGenesis should panic with invalid ledger")
				return
			}

			// Initialize genesis state
			s.keeper.InitGenesis(s.ctx, tc.genState)

			// Verify ledgers were created
			for _, l := range tc.genState.Ledgers {
				// Get the ledger and verify it exists
				foundLedger, err := s.keeper.GetLedger(s.ctx, l.NftAddress)
				s.Require().NoError(err, "GetLedger error")
				s.Require().NotNil(foundLedger, "GetLedger result")
				s.Require().Equal(l.NftAddress, foundLedger.NftAddress, "ledger nft address")
				s.Require().Equal(l.Denom, foundLedger.Denom, "ledger denom")

				// Verify event emission
				events := s.ctx.EventManager().Events()
				var foundEvents []*sdk.Event
				for _, e := range events {
					if e.Type == ledger.EventTypeLedgerCreated {
						foundEvents = append(foundEvents, &e)
					}
				}

				// Find the event for this specific ledger
				var foundEvent *sdk.Event
				for _, e := range foundEvents {
					if string(e.Attributes[0].Value) == l.NftAddress {
						foundEvent = e
						break
					}
				}

				s.Require().NotNil(foundEvent, "Event should be emitted for ledger creation")
				s.Require().Equal(ledger.EventTypeLedgerCreated, foundEvent.Type, "event type")
				s.Require().Len(foundEvent.Attributes, 2, "event attributes length")
				s.Require().Equal("nft_address", string(foundEvent.Attributes[0].Key), "event nft address key")
				s.Require().Equal(l.NftAddress, string(foundEvent.Attributes[0].Value), "event nft address value")
				s.Require().Equal("denom", string(foundEvent.Attributes[1].Key), "event denom key")
				s.Require().Equal(l.Denom, string(foundEvent.Attributes[1].Value), "event denom value")
			}
		})
	}
}

// TestExportGenesis tests the ExportGenesis function
func (s *TestSuite) TestExportGenesis() {
	// Create test ledgers
	ledger1 := ledger.Ledger{
		NftAddress: s.addr1.String(),
		Denom:      s.bondDenom,
	}
	ledger2 := ledger.Ledger{
		NftAddress: s.addr2.String(),
		Denom:      s.bondDenom,
	}

	// Create ledgers first
	err := s.keeper.CreateLedger(s.ctx, ledger1)
	s.Require().NoError(err, "CreateLedger error")

	// Create second ledger
	err = s.keeper.CreateLedger(s.ctx, ledger2)
	s.Require().NoError(err, "CreateLedger error")

	// Verify first ledger creation event
	ledgerCreateEvents := make(map[string]*sdk.Event)
	events := s.ctx.EventManager().Events()
	for _, e := range events {
		if e.Type == ledger.EventTypeLedgerCreated {
			ledgerCreateEvents[string(e.Attributes[0].Value)] = &e
		}
	}

	l1Event := ledgerCreateEvents[ledger1.NftAddress]
	s.Require().NotNil(l1Event, "ledger1 event should be emitted")

	l2Event := ledgerCreateEvents[ledger2.NftAddress]
	s.Require().NotNil(l2Event, "ledger2 event should be emitted")

	eventCheck := func(e sdk.Event, l ledger.Ledger) {
		s.Require().Equal(ledger.EventTypeLedgerCreated, e.Type, "event type")
		s.Require().Len(e.Attributes, 2, "event attributes length")
		s.Require().Equal("nft_address", string(e.Attributes[0].Key), "event nft address key")
		s.Require().Equal(l.NftAddress, string(e.Attributes[0].Value), "event nft address value")
		s.Require().Equal("denom", string(e.Attributes[1].Key), "event denom key")
		s.Require().Equal(l.Denom, string(e.Attributes[1].Value), "event denom value")
	}

	eventCheck(*l1Event, ledger1)
	eventCheck(*l2Event, ledger2)

	// Export genesis state
	genState := s.keeper.ExportGenesis(s.ctx)
	s.Require().Len(genState.Ledgers, 2, "number of ledgers in exported state")

	// map the ledgers to an nft address for easier comparison
	ledgers := genState.Ledgers
	exportedLedgers := make(map[string]ledger.Ledger)
	for _, l := range ledgers {
		exportedLedgers[l.NftAddress] = l
	}

	genStateCheck := func(expected ledger.Ledger, actual ledger.Ledger) {
		s.Require().Equal(expected.NftAddress, actual.NftAddress, "ledger nft address")
		s.Require().Equal(expected.Denom, actual.Denom, "ledger denom")
	}
	genStateCheck(ledger1, exportedLedgers[ledger1.NftAddress])
	genStateCheck(ledger2, exportedLedgers[ledger2.NftAddress])
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
func (s *TestSuite) getAddrName(addr sdk.AccAddress) string {
	switch string(addr) {
	case string(s.addr1):
		return "addr1"
	case string(s.addr2):
		return "addr2"
	case string(s.addr3):
		return "addr3"
	default:
		return addr.String()
	}
}

// requireFundAccount calls testutil.FundAccount, making sure it doesn't panic or return an error.
func (s *TestSuite) requireFundAccount(addr sdk.AccAddress, coins string) {
	assertions.RequireNotPanicsNoErrorf(s.T(), func() error {
		return testutil.FundAccount(s.ctx, s.app.BankKeeper, addr, s.coins(coins))
	}, "FundAccount(%s, %q)", s.getAddrName(addr), coins)
}

// assertEqualEvents asserts that the expected events equal the actual events.
func (s *TestSuite) assertEqualEvents(expected, actual sdk.Events, msgAndArgs ...interface{}) bool {
	return assertions.AssertEqualEvents(s.T(), expected, actual, msgAndArgs...)
}
