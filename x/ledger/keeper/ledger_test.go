package keeper_test

import (
	"cosmossdk.io/math"
	"cosmossdk.io/x/nft"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/provenance-io/provenance/x/ledger/keeper"
	"github.com/provenance-io/provenance/x/ledger/types"
	registrytypes "github.com/provenance-io/provenance/x/registry/types"
)

func (s *TestSuite) TestNonExistentDenom() {
	// Mint a new nft.
	nft := nft.NFT{
		ClassId: s.validNFTClass.Id,
		Id:      "test-nft-id-2",
	}
	s.nftKeeper.Mint(s.ctx, nft, s.addr1)

	// Attempt to attach a denom that doesn't exist.
	ledgerClass := types.LedgerClass{
		LedgerClassId:     "test-ledger-class-id-2",
		AssetClassId:      s.validNFTClass.Id,
		MaintainerAddress: s.addr1.String(),
		Denom:             "non-existent-denom",
	}
	err := s.keeper.AddLedgerClass(s.ctx, ledgerClass)
	s.Require().Error(err, "CreateLedgerClass error")
	s.Require().Contains(err.Error(), "denom doesn't have a supply", "CreateLedgerClass error")
}

func (s *TestSuite) TestCreateLedgerClassMaintainerNotOwner() {
	s.T().Skip("Skipping test - authorization logic moved out of keeper")
	err := s.keeper.AddClassStatusType(s.ctx, s.validLedgerClass.LedgerClassId, types.LedgerClassStatusType{
		Id:          1,
		Code:        "IN_REPAYMENT",
		Description: "In Repayment",
	})
	s.Require().Error(err, "AddClassStatusType error")
	s.Require().Contains(err.Error(), types.ErrCodeUnauthorized, "AddClassStatusType error")
}

// Test to ensure only the registered servicer or owner can create a ledger.
func (s *TestSuite) TestCreateLedgerNotOwnerOrServicer() {
	s.T().Skip("Skipping test - authorization logic moved out of keeper")
	ledger := types.Ledger{
		Key: &types.LedgerKey{
			AssetClassId: s.validNFTClass.Id,
			NftId:        s.validNFT.Id,
		},
		LedgerClassId: s.validLedgerClass.LedgerClassId,
		StatusTypeId:  1,
	}

	err := s.keeper.AddLedger(s.ctx, ledger)
	s.Require().Error(err, "CreateLedger error")
	s.Require().Contains(err.Error(), "unauthorized", "CreateLedger error")

	registryKey := &registrytypes.RegistryKey{
		AssetClassId: s.validNFTClass.Id,
		NftId:        s.validNFT.Id,
	}

	// Create a no role registry entry for the nft
	err = s.registryKeeper.CreateRegistry(s.ctx, registryKey, []registrytypes.RolesEntry{})
	s.Require().NoError(err, "CreateRegistry error")

	err = s.keeper.AddLedger(s.ctx, ledger)
	s.Require().Error(err, "CreateLedger error")

	// Grant a role of servicer to the s.addr2 so that it can create the ledger
	err = s.registryKeeper.GrantRole(s.ctx, registryKey, registrytypes.RegistryRole_REGISTRY_ROLE_SERVICER, []string{s.addr2.String()})
	s.Require().NoError(err, "GrantRole error")

	// Verify that the registry granted the role to the s.addr2
	hasRole, err := s.registryKeeper.HasRole(s.ctx, registryKey, registrytypes.RegistryRole_REGISTRY_ROLE_SERVICER, s.addr2.String())
	s.Require().NoError(err, "HasRole error")
	s.Require().True(hasRole, "HasRole error")

	// Verify that the s.addr2 can create the ledger as the servicer
	err = s.keeper.AddLedger(s.ctx, ledger)
	s.Require().NoError(err, "CreateLedger error")
}

func (s *TestSuite) TestCreateLedgerClass() {
	tests := []struct {
		name        string
		ledgerClass types.LedgerClass
		expErr      []string
	}{
		{
			name: "valid ledger class should already exist",
			ledgerClass: types.LedgerClass{
				LedgerClassId:     s.validLedgerClass.LedgerClassId,
				AssetClassId:      s.validLedgerClass.AssetClassId,
				MaintainerAddress: s.addr1.String(),
				Denom:             s.bondDenom,
			},
			expErr: []string{"already exists"},
		},
		{
			name: "invalid asset class id",
			ledgerClass: types.LedgerClass{
				LedgerClassId:     s.validLedgerClass.LedgerClassId,
				AssetClassId:      "non-existent-class-id",
				MaintainerAddress: s.addr1.String(),
				Denom:             s.bondDenom,
			},
			expErr: []string{"asset_class_id"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := s.keeper.AddLedgerClass(s.ctx, tc.ledgerClass)
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
		ledger   types.Ledger
		expErr   []error
		expEvent bool
	}{
		{
			name: "valid ledger",
			ledger: types.Ledger{
				Key: &types.LedgerKey{
					AssetClassId: s.validNFTClass.Id,
					NftId:        s.validNFT.Id,
				},
				LedgerClassId: s.validLedgerClass.LedgerClassId,
				StatusTypeId:  1,
			},
			expEvent: true,
		},

		{
			name: "duplicate ledger",
			ledger: types.Ledger{
				Key: &types.LedgerKey{
					AssetClassId: s.validNFTClass.Id,
					NftId:        s.validNFT.Id,
				},
				LedgerClassId: s.validLedgerClass.LedgerClassId,
				StatusTypeId:  1,
			},
			expErr: []error{types.ErrAlreadyExists},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Clear events before each test
			s.ctx.WithEventManager(sdk.NewEventManager())

			err := s.keeper.AddLedger(s.ctx, tc.ledger)

			if len(tc.expErr) > 0 {
				s.Require().ErrorIs(err, tc.expErr[0], "CreateLedger error")
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
						// The event type should be the proto message name with package prefix
						// For EventLedgerCreated in package provenance.ledger.v1, it should be:
						// "provenance.ledger.v1.EventLedgerCreated"
						if e.Type == "provenance.ledger.v1.EventLedgerCreated" {
							foundEvent = &e
							break
						}
					}
					s.Require().NotNil(foundEvent, "EventLedgerCreated event should be found")
					s.Require().Equal("provenance.ledger.v1.EventLedgerCreated", foundEvent.Type, "event type")
					s.Require().Len(foundEvent.Attributes, 2, "event attributes length")
					s.Require().Equal("asset_class_id", foundEvent.Attributes[0].Key, "event asset class id key")
					s.Require().Contains(foundEvent.Attributes[0].Value, tc.ledger.Key.AssetClassId, "event asset class id value")
					s.Require().Equal("nft_id", foundEvent.Attributes[1].Key, "event nft id key")
					s.Require().Contains(foundEvent.Attributes[1].Value, tc.ledger.Key.NftId, "event nft id value")
				}
			}
		})
	}
}

// TestGetLedger tests the GetLedger function
func (s *TestSuite) TestGetLedger() {
	// Create a valid ledger first that we can try to get
	validLedger := types.Ledger{
		Key: &types.LedgerKey{
			AssetClassId: s.validNFTClass.Id,
			NftId:        s.validNFT.Id,
		},
		LedgerClassId: s.validLedgerClass.LedgerClassId,
		StatusTypeId:  1,
	}
	err := s.keeper.AddLedger(s.ctx, validLedger)
	s.Require().NoError(err, "CreateLedger error")

	tests := []struct {
		name      string
		nftId     string
		expErr    []string
		expLedger *types.Ledger
	}{
		{
			name:      "valid ledger retrieval",
			nftId:     s.validNFT.Id,
			expLedger: &validLedger,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			ledger, err := s.keeper.GetLedger(s.ctx, &types.LedgerKey{
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
	l := types.Ledger{
		Key: &types.LedgerKey{
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
		name          string
		key           *types.LedgerKey
		correlationId string
		expEntry      *types.LedgerEntry
		expErr        *types.ErrCode
	}{
		{
			name: "not found",
			key: &types.LedgerKey{
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
		genState *types.GenesisState
	}{
		{
			name:     "empty genesis state",
			genState: &types.GenesisState{},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Clear events before each test
			s.ctx.WithEventManager(sdk.NewEventManager())

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

func (s *TestSuite) TestBech32() {
	ledgerKey := &types.LedgerKey{
		AssetClassId: s.validNFTClass.Id,
		NftId:        s.addr1.String(),
	}

	expectedBech32Str := "ledger1w3jhxapdden8gttrd3shxuedd9jqqcm0wdkk7ue3x44hjwtyw5uxzvnhd3ehg73kvec8svmsx3khzur209ex6dtrvack5amv8pehzl09ezy"

	bech32Id := ledgerKey.String()
	s.Require().Equal(expectedBech32Str, bech32Id)

	ledgerKey2, err := types.StringToLedgerKey(bech32Id)
	s.Require().NoError(err, "StringToLedgerKey error")
	s.Require().Equal(ledgerKey, ledgerKey2, "ledger keys should be equal")

	_, err = types.StringToLedgerKey("ledgerasdf1w3jhxapdden8gttrd3shxuedd9jr5cm0wdkk7ue3x44hjwtyw5uxzvnhd3ehg73kvec8svmsx3khzur209ex6dtrvack5amv8pehzv7wxj4")
	s.Require().Error(err, "StringToLedgerKey error")
}

// TestUpdateLedgerStatus tests the UpdateLedgerStatus function and event emission
func (s *TestSuite) TestUpdateLedgerStatus() {
	// Create a valid ledger first
	validLedger := types.Ledger{
		Key: &types.LedgerKey{
			AssetClassId: s.validNFTClass.Id,
			NftId:        s.validNFT.Id,
		},
		LedgerClassId: s.validLedgerClass.LedgerClassId,
		StatusTypeId:  1,
	}
	err := s.keeper.AddLedger(s.ctx, validLedger)
	s.Require().NoError(err, "CreateLedger error")

	tests := []struct {
		name         string
		statusTypeId int32
		expErr       []string
		expEvent     bool
	}{
		{
			name:         "valid status update",
			statusTypeId: 2,
			expEvent:     true,
		},
		{
			name:         "invalid status type id",
			statusTypeId: 999,
			expErr:       []string{"status type doesn't exist"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Clear events before each test
			s.ctx.WithEventManager(sdk.NewEventManager())

			err := s.keeper.UpdateLedgerStatus(s.ctx, validLedger.Key, tc.statusTypeId)

			if len(tc.expErr) > 0 {
				s.assertErrorContents(err, tc.expErr, "UpdateLedgerStatus error")
			} else {
				s.Require().NoError(err, "UpdateLedgerStatus error")

				// Verify the ledger was updated
				l, err := s.keeper.GetLedger(s.ctx, validLedger.Key)
				s.Require().NoError(err, "GetLedger error")
				s.Require().NotNil(l, "GetLedger result")
				s.Require().Equal(tc.statusTypeId, l.StatusTypeId, "ledger status type id")

				// Verify event emission
				if tc.expEvent {
					var foundEvent *sdk.Event
					for _, e := range s.ctx.EventManager().Events() {
						if e.Type == "provenance.ledger.v1.EventLedgerUpdated" {
							foundEvent = &e
							break
						}
					}
					s.Require().NotNil(foundEvent, "EventLedgerUpdated event should be found")
					s.Require().Equal("provenance.ledger.v1.EventLedgerUpdated", foundEvent.Type, "event type")
					s.Require().Len(foundEvent.Attributes, 3, "event attributes length")
					s.Require().Equal("asset_class_id", foundEvent.Attributes[0].Key, "event asset class id key")
					s.Require().Equal("nft_id", foundEvent.Attributes[1].Key, "event nft id key")
					s.Require().Equal("update_type", foundEvent.Attributes[2].Key, "event update type key")
					s.Require().Contains(foundEvent.Attributes[2].Value, "UPDATE_TYPE_STATUS", "event update type value")
				}
			}
		})
	}
}

// TestUpdateLedgerInterestRate tests the UpdateLedgerInterestRate function and event emission
func (s *TestSuite) TestUpdateLedgerInterestRate() {
	// Create a valid ledger first
	validLedger := types.Ledger{
		Key: &types.LedgerKey{
			AssetClassId: s.validNFTClass.Id,
			NftId:        s.validNFT.Id,
		},
		LedgerClassId: s.validLedgerClass.LedgerClassId,
		StatusTypeId:  1,
	}
	err := s.keeper.AddLedger(s.ctx, validLedger)
	s.Require().NoError(err, "CreateLedger error")

	tests := []struct {
		name                  string
		interestRate          int32
		dayCountConvention    types.DayCountConvention
		interestAccrualMethod types.InterestAccrualMethod
		expErr                []string
		expEvent              bool
	}{
		{
			name:                  "valid interest rate update",
			interestRate:          500, // 5%
			dayCountConvention:    types.DAY_COUNT_CONVENTION_THIRTY_360,
			interestAccrualMethod: types.INTEREST_ACCRUAL_METHOD_SIMPLE_INTEREST,
			expEvent:              true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Clear events before each test
			s.ctx.WithEventManager(sdk.NewEventManager())

			err := s.keeper.UpdateLedgerInterestRate(s.ctx, validLedger.Key, tc.interestRate, tc.dayCountConvention, tc.interestAccrualMethod)

			if len(tc.expErr) > 0 {
				s.assertErrorContents(err, tc.expErr, "UpdateLedgerInterestRate error")
			} else {
				s.Require().NoError(err, "UpdateLedgerInterestRate error")

				// Verify the ledger was updated
				l, err := s.keeper.GetLedger(s.ctx, validLedger.Key)
				s.Require().NoError(err, "GetLedger error")
				s.Require().NotNil(l, "GetLedger result")
				s.Require().Equal(tc.interestRate, l.InterestRate, "ledger interest rate")
				s.Require().Equal(tc.dayCountConvention, l.InterestDayCountConvention, "ledger day count convention")
				s.Require().Equal(tc.interestAccrualMethod, l.InterestAccrualMethod, "ledger interest accrual method")

				// Verify event emission
				if tc.expEvent {
					var foundEvent *sdk.Event
					for _, e := range s.ctx.EventManager().Events() {
						if e.Type == "provenance.ledger.v1.EventLedgerUpdated" {
							foundEvent = &e
							break
						}
					}
					s.Require().NotNil(foundEvent, "EventLedgerUpdated event should be found")
					s.Require().Equal("provenance.ledger.v1.EventLedgerUpdated", foundEvent.Type, "event type")
					s.Require().Len(foundEvent.Attributes, 3, "event attributes length")
					s.Require().Equal("update_type", foundEvent.Attributes[2].Key, "event update type key")
					s.Require().Contains(foundEvent.Attributes[2].Value, "UPDATE_TYPE_INTEREST_RATE", "event update type value")
				}
			}
		})
	}
}

// TestUpdateLedgerPayment tests the UpdateLedgerPayment function and event emission
func (s *TestSuite) TestUpdateLedgerPayment() {
	// Create a valid ledger first
	validLedger := types.Ledger{
		Key: &types.LedgerKey{
			AssetClassId: s.validNFTClass.Id,
			NftId:        s.validNFT.Id,
		},
		LedgerClassId: s.validLedgerClass.LedgerClassId,
		StatusTypeId:  1,
	}
	err := s.keeper.AddLedger(s.ctx, validLedger)
	s.Require().NoError(err, "CreateLedger error")

	tests := []struct {
		name             string
		nextPmtAmt       math.Int
		nextPmtDate      int32
		paymentFrequency types.PaymentFrequency
		expErr           []string
		expEvent         bool
	}{
		{
			name:             "valid payment update",
			nextPmtAmt:       math.NewInt(1000000), // 1000 tokens
			nextPmtDate:      20241201,             // Dec 1, 2024
			paymentFrequency: types.PAYMENT_FREQUENCY_MONTHLY,
			expEvent:         true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Clear events before each test
			s.ctx.WithEventManager(sdk.NewEventManager())

			err := s.keeper.UpdateLedgerPayment(s.ctx, validLedger.Key, tc.nextPmtAmt, tc.nextPmtDate, tc.paymentFrequency)

			if len(tc.expErr) > 0 {
				s.assertErrorContents(err, tc.expErr, "UpdateLedgerPayment error")
			} else {
				s.Require().NoError(err, "UpdateLedgerPayment error")

				// Verify the ledger was updated
				l, err := s.keeper.GetLedger(s.ctx, validLedger.Key)
				s.Require().NoError(err, "GetLedger error")
				s.Require().NotNil(l, "GetLedger result")
				s.Require().Equal(tc.nextPmtAmt, l.NextPmtAmt, "ledger next payment amount")
				s.Require().Equal(tc.nextPmtDate, l.NextPmtDate, "ledger next payment date")
				s.Require().Equal(tc.paymentFrequency, l.PaymentFrequency, "ledger payment frequency")

				// Verify event emission
				if tc.expEvent {
					var foundEvent *sdk.Event
					for _, e := range s.ctx.EventManager().Events() {
						if e.Type == "provenance.ledger.v1.EventLedgerUpdated" {
							foundEvent = &e
							break
						}
					}
					s.Require().NotNil(foundEvent, "EventLedgerUpdated event should be found")
					s.Require().Equal("provenance.ledger.v1.EventLedgerUpdated", foundEvent.Type, "event type")
					s.Require().Len(foundEvent.Attributes, 3, "event attributes length")
					s.Require().Equal("update_type", foundEvent.Attributes[2].Key, "event update type key")
					s.Require().Contains(foundEvent.Attributes[2].Value, "UPDATE_TYPE_PAYMENT", "event update type value")
				}
			}
		})
	}
}

// TestUpdateLedgerMaturityDate tests the UpdateLedgerMaturityDate function and event emission
func (s *TestSuite) TestUpdateLedgerMaturityDate() {
	// Create a valid ledger first
	validLedger := types.Ledger{
		Key: &types.LedgerKey{
			AssetClassId: s.validNFTClass.Id,
			NftId:        s.validNFT.Id,
		},
		LedgerClassId: s.validLedgerClass.LedgerClassId,
		StatusTypeId:  1,
	}
	err := s.keeper.AddLedger(s.ctx, validLedger)
	s.Require().NoError(err, "CreateLedger error")

	tests := []struct {
		name         string
		maturityDate int32
		expErr       []string
		expEvent     bool
	}{
		{
			name:         "valid maturity date update",
			maturityDate: 20251231, // Dec 31, 2025
			expEvent:     true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Clear events before each test
			s.ctx.WithEventManager(sdk.NewEventManager())

			err := s.keeper.UpdateLedgerMaturityDate(s.ctx, validLedger.Key, tc.maturityDate)

			if len(tc.expErr) > 0 {
				s.assertErrorContents(err, tc.expErr, "UpdateLedgerMaturityDate error")
			} else {
				s.Require().NoError(err, "UpdateLedgerMaturityDate error")

				// Verify the ledger was updated
				l, err := s.keeper.GetLedger(s.ctx, validLedger.Key)
				s.Require().NoError(err, "GetLedger error")
				s.Require().NotNil(l, "GetLedger result")
				s.Require().Equal(tc.maturityDate, l.MaturityDate, "ledger maturity date")

				// Verify event emission
				if tc.expEvent {
					var foundEvent *sdk.Event
					for _, e := range s.ctx.EventManager().Events() {
						if e.Type == "provenance.ledger.v1.EventLedgerUpdated" {
							foundEvent = &e
							break
						}
					}
					s.Require().NotNil(foundEvent, "EventLedgerUpdated event should be found")
					s.Require().Equal("provenance.ledger.v1.EventLedgerUpdated", foundEvent.Type, "event type")
					s.Require().Len(foundEvent.Attributes, 3, "event attributes length")
					s.Require().Equal("update_type", foundEvent.Attributes[2].Key, "event update type key")
					s.Require().Contains(foundEvent.Attributes[2].Value, "UPDATE_TYPE_MATURITY_DATE", "event update type value")
				}
			}
		})
	}
}

func (s *TestSuite) TestGetAllLedgerClasses() {
	// Create additional ledger classes for testing pagination
	nftClass2 := nft.Class{Id: "test-nft-class-id-2"}
	s.nftKeeper.SaveClass(s.ctx, nftClass2)

	nft2 := nft.NFT{ClassId: nftClass2.Id, Id: "test-nft-id-2"}
	s.nftKeeper.Mint(s.ctx, nft2, s.addr1)

	ledgerClass2 := types.LedgerClass{
		LedgerClassId:     "test-ledger-class-id-2",
		AssetClassId:      nftClass2.Id,
		MaintainerAddress: s.addr1.String(),
		Denom:             s.bondDenom,
	}
	err := s.keeper.AddLedgerClass(s.ctx, ledgerClass2)
	s.Require().NoError(err, "AddLedgerClass error")

	nftClass3 := nft.Class{Id: "test-nft-class-id-3"}
	s.nftKeeper.SaveClass(s.ctx, nftClass3)

	nft3 := nft.NFT{ClassId: nftClass3.Id, Id: "test-nft-id-3"}
	s.nftKeeper.Mint(s.ctx, nft3, s.addr1)

	ledgerClass3 := types.LedgerClass{
		LedgerClassId:     "test-ledger-class-id-3",
		AssetClassId:      nftClass3.Id,
		MaintainerAddress: s.addr1.String(),
		Denom:             s.bondDenom,
	}
	err = s.keeper.AddLedgerClass(s.ctx, ledgerClass3)
	s.Require().NoError(err, "AddLedgerClass error")

	tests := []struct {
		name          string
		pageRequest   *query.PageRequest
		expectedCount int
		expectedTotal uint64
		expectError   bool
	}{
		{
			name:          "get all ledger classes without pagination",
			pageRequest:   nil,
			expectedCount: 3, // Including the one from SetupTest
			expectedTotal: 3,
			expectError:   false,
		},
		{
			name:          "get all ledger classes with limit 2",
			pageRequest:   &query.PageRequest{Limit: 2, CountTotal: true},
			expectedCount: 2,
			expectedTotal: 3,
			expectError:   false,
		},
		{
			name:          "get all ledger classes with offset 1 and limit 2",
			pageRequest:   &query.PageRequest{Offset: 1, Limit: 2, CountTotal: true},
			expectedCount: 2,
			expectedTotal: 3,
			expectError:   false,
		},
		{
			name:          "get all ledger classes with offset 2 and limit 5",
			pageRequest:   &query.PageRequest{Offset: 2, Limit: 5, CountTotal: true},
			expectedCount: 1,
			expectedTotal: 3,
			expectError:   false,
		},
		{
			name:          "get all ledger classes with offset beyond total count",
			pageRequest:   &query.PageRequest{Offset: 10, Limit: 5, CountTotal: true},
			expectedCount: 0,
			expectedTotal: 3,
			expectError:   false,
		},
		{
			name:          "get all ledger classes with large limit",
			pageRequest:   &query.PageRequest{Limit: 100, CountTotal: true},
			expectedCount: 3,
			expectedTotal: 3,
			expectError:   false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Set default pagination if nil
			pageReq := tc.pageRequest
			if pageReq == nil {
				pageReq = &query.PageRequest{CountTotal: true}
			}

			ledgerClasses, pageRes, err := s.keeper.GetAllLedgerClasses(s.ctx, pageReq)

			if tc.expectError {
				s.Require().Error(err)
				return
			}

			s.Require().NoError(err)
			s.Require().NotNil(ledgerClasses)
			s.Require().Len(ledgerClasses, tc.expectedCount)
			if pageRes != nil && pageRes.Total > 0 {
				s.Require().Equal(tc.expectedTotal, pageRes.Total)
			}

			// Verify all returned ledger classes are valid
			for _, lc := range ledgerClasses {
				s.Require().NotNil(lc)
				s.Require().NotEmpty(lc.LedgerClassId)
				s.Require().NotEmpty(lc.AssetClassId)
				s.Require().NotEmpty(lc.MaintainerAddress)
				s.Require().NotEmpty(lc.Denom)
			}
		})
	}
}

func (s *TestSuite) TestLedgerClassesQueryServer() {
	// Create additional ledger classes for testing
	nftClass2 := nft.Class{Id: "test-nft-class-id-2"}
	s.nftKeeper.SaveClass(s.ctx, nftClass2)

	nft2 := nft.NFT{ClassId: nftClass2.Id, Id: "test-nft-id-2"}
	s.nftKeeper.Mint(s.ctx, nft2, s.addr1)

	ledgerClass2 := types.LedgerClass{
		LedgerClassId:     "test-ledger-class-id-2",
		AssetClassId:      nftClass2.Id,
		MaintainerAddress: s.addr1.String(),
		Denom:             s.bondDenom,
	}
	err := s.keeper.AddLedgerClass(s.ctx, ledgerClass2)
	s.Require().NoError(err, "AddLedgerClass error")

	// Create query server
	queryServer := keeper.NewLedgerQueryServer(s.keeper)

	tests := []struct {
		name        string
		request     *types.QueryLedgerClassesRequest
		expectError bool
	}{
		{
			name: "valid request without pagination",
			request: &types.QueryLedgerClassesRequest{
				Pagination: &query.PageRequest{CountTotal: true},
			},
			expectError: false,
		},
		{
			name: "valid request with pagination limit",
			request: &types.QueryLedgerClassesRequest{
				Pagination: &query.PageRequest{Limit: 1, CountTotal: true},
			},
			expectError: false,
		},
		{
			name: "valid request with nil pagination",
			request: &types.QueryLedgerClassesRequest{
				Pagination: nil,
			},
			expectError: false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			resp, err := queryServer.LedgerClasses(sdk.WrapSDKContext(s.ctx), tc.request)

			if tc.expectError {
				s.Require().Error(err)
				return
			}

			s.Require().NoError(err)
			s.Require().NotNil(resp)
			s.Require().NotNil(resp.LedgerClasses)
			s.Require().NotNil(resp.Pagination)

			// Verify we have at least the original ledger class from SetupTest plus the new one
			s.Require().GreaterOrEqual(len(resp.LedgerClasses), 1)

			// If pagination was set with limit 1, should return exactly 1
			if tc.request.Pagination != nil && tc.request.Pagination.Limit == 1 {
				s.Require().Len(resp.LedgerClasses, 1)
			}

			// Verify all returned ledger classes are valid
			for _, lc := range resp.LedgerClasses {
				s.Require().NotNil(lc)
				s.Require().NotEmpty(lc.LedgerClassId)
				s.Require().NotEmpty(lc.AssetClassId)
				s.Require().NotEmpty(lc.MaintainerAddress)
				s.Require().NotEmpty(lc.Denom)
			}
		})
	}
}

func (s *TestSuite) TestGetAllLedgers() {
	// Create additional ledgers for testing pagination
	// First, create a second ledger class and NFT
	nftClass2 := nft.Class{Id: "test-nft-class-id-2"}
	s.nftKeeper.SaveClass(s.ctx, nftClass2)

	nft2 := nft.NFT{ClassId: nftClass2.Id, Id: "test-nft-id-2"}
	s.nftKeeper.Mint(s.ctx, nft2, s.addr1)

	ledgerClass2 := types.LedgerClass{
		LedgerClassId:     "test-ledger-class-id-2",
		AssetClassId:      nftClass2.Id,
		MaintainerAddress: s.addr1.String(),
		Denom:             s.bondDenom,
	}
	err := s.keeper.AddLedgerClass(s.ctx, ledgerClass2)
	s.Require().NoError(err, "AddLedgerClass error")

	// Add status types for the new ledger class
	err = s.keeper.AddClassStatusType(s.ctx, ledgerClass2.LedgerClassId, types.LedgerClassStatusType{
		Id:          1,
		Code:        "ACTIVE",
		Description: "Active",
	})
	s.Require().NoError(err, "AddClassStatusType error")

	// Create multiple ledgers for testing
	ledger1 := types.Ledger{
		Key: &types.LedgerKey{
			AssetClassId: s.validNFTClass.Id,
			NftId:        "test-ledger-nft-1",
		},
		LedgerClassId: s.validLedgerClass.LedgerClassId,
		StatusTypeId:  1,
	}

	// Mint NFT for ledger1
	nft1 := nft.NFT{ClassId: s.validNFTClass.Id, Id: "test-ledger-nft-1"}
	s.nftKeeper.Mint(s.ctx, nft1, s.addr1)
	err = s.keeper.AddLedger(s.ctx, ledger1)
	s.Require().NoError(err, "AddLedger error")

	ledger2 := types.Ledger{
		Key: &types.LedgerKey{
			AssetClassId: nftClass2.Id,
			NftId:        nft2.Id,
		},
		LedgerClassId: ledgerClass2.LedgerClassId,
		StatusTypeId:  1,
	}
	err = s.keeper.AddLedger(s.ctx, ledger2)
	s.Require().NoError(err, "AddLedger error")

	ledger3 := types.Ledger{
		Key: &types.LedgerKey{
			AssetClassId: s.validNFTClass.Id,
			NftId:        "test-ledger-nft-3",
		},
		LedgerClassId: s.validLedgerClass.LedgerClassId,
		StatusTypeId:  1,
	}

	// Mint NFT for ledger3
	nft3 := nft.NFT{ClassId: s.validNFTClass.Id, Id: "test-ledger-nft-3"}
	s.nftKeeper.Mint(s.ctx, nft3, s.addr1)
	err = s.keeper.AddLedger(s.ctx, ledger3)
	s.Require().NoError(err, "AddLedger error")

	tests := []struct {
		name          string
		pageRequest   *query.PageRequest
		expectedCount int
		expectedTotal uint64
		expectError   bool
	}{
		{
			name:          "get all ledgers without pagination",
			pageRequest:   nil,
			expectedCount: 3,
			expectedTotal: 3,
			expectError:   false,
		},
		{
			name:          "get all ledgers with limit 2",
			pageRequest:   &query.PageRequest{Limit: 2, CountTotal: true},
			expectedCount: 2,
			expectedTotal: 3,
			expectError:   false,
		},
		{
			name:          "get all ledgers with offset 1 and limit 2",
			pageRequest:   &query.PageRequest{Offset: 1, Limit: 2, CountTotal: true},
			expectedCount: 2,
			expectedTotal: 3,
			expectError:   false,
		},
		{
			name:          "get all ledgers with offset 2 and limit 5",
			pageRequest:   &query.PageRequest{Offset: 2, Limit: 5, CountTotal: true},
			expectedCount: 1,
			expectedTotal: 3,
			expectError:   false,
		},
		{
			name:          "get all ledgers with offset beyond total count",
			pageRequest:   &query.PageRequest{Offset: 10, Limit: 5, CountTotal: true},
			expectedCount: 0,
			expectedTotal: 3,
			expectError:   false,
		},
		{
			name:          "get all ledgers with large limit",
			pageRequest:   &query.PageRequest{Limit: 100, CountTotal: true},
			expectedCount: 3,
			expectedTotal: 3,
			expectError:   false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Set default pagination if nil
			pageReq := tc.pageRequest
			if pageReq == nil {
				pageReq = &query.PageRequest{CountTotal: true}
			}

			ledgers, pageRes, err := s.keeper.GetAllLedgers(s.ctx, pageReq)

			if tc.expectError {
				s.Require().Error(err)
				return
			}

			s.Require().NoError(err)
			s.Require().NotNil(ledgers)
			s.Require().Len(ledgers, tc.expectedCount)
			if pageRes != nil && pageRes.Total > 0 {
				s.Require().Equal(tc.expectedTotal, pageRes.Total)
			}

			// Verify all returned ledgers are valid and have their keys set
			for _, l := range ledgers {
				s.Require().NotNil(l)
				s.Require().NotNil(l.Key)
				s.Require().NotEmpty(l.Key.AssetClassId)
				s.Require().NotEmpty(l.Key.NftId)
				s.Require().NotEmpty(l.LedgerClassId)
				s.Require().NotZero(l.StatusTypeId)
			}
		})
	}
}

func (s *TestSuite) TestLedgersQueryServer() {
	// Create additional ledger for testing
	ledger1 := types.Ledger{
		Key: &types.LedgerKey{
			AssetClassId: s.validNFTClass.Id,
			NftId:        "test-ledger-nft-server",
		},
		LedgerClassId: s.validLedgerClass.LedgerClassId,
		StatusTypeId:  1,
	}

	// Mint NFT for ledger1
	nft1 := nft.NFT{ClassId: s.validNFTClass.Id, Id: "test-ledger-nft-server"}
	s.nftKeeper.Mint(s.ctx, nft1, s.addr1)
	err := s.keeper.AddLedger(s.ctx, ledger1)
	s.Require().NoError(err, "AddLedger error")

	// Create query server
	queryServer := keeper.NewLedgerQueryServer(s.keeper)

	tests := []struct {
		name        string
		request     *types.QueryLedgersRequest
		expectError bool
	}{
		{
			name: "valid request without pagination",
			request: &types.QueryLedgersRequest{
				Pagination: &query.PageRequest{CountTotal: true},
			},
			expectError: false,
		},
		{
			name: "valid request with pagination limit",
			request: &types.QueryLedgersRequest{
				Pagination: &query.PageRequest{Limit: 1, CountTotal: true},
			},
			expectError: false,
		},
		{
			name: "valid request with nil pagination",
			request: &types.QueryLedgersRequest{
				Pagination: nil,
			},
			expectError: false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			resp, err := queryServer.Ledgers(sdk.WrapSDKContext(s.ctx), tc.request)

			if tc.expectError {
				s.Require().Error(err)
				return
			}

			s.Require().NoError(err)
			s.Require().NotNil(resp)
			s.Require().NotNil(resp.Ledgers)
			s.Require().NotNil(resp.Pagination)

			// Verify we have at least one ledger
			s.Require().GreaterOrEqual(len(resp.Ledgers), 1)

			// If pagination was set with limit 1, should return exactly 1
			if tc.request.Pagination != nil && tc.request.Pagination.Limit == 1 {
				s.Require().Len(resp.Ledgers, 1)
			}

			// Verify all returned ledgers are valid and have their keys set
			for _, l := range resp.Ledgers {
				s.Require().NotNil(l)
				s.Require().NotNil(l.Key)
				s.Require().NotEmpty(l.Key.AssetClassId)
				s.Require().NotEmpty(l.Key.NftId)
				s.Require().NotEmpty(l.LedgerClassId)
				s.Require().NotZero(l.StatusTypeId)
			}
		})
	}
}
