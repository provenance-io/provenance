package keeper_test

import (
	"cosmossdk.io/x/nft"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/ledger/types"
	ledger "github.com/provenance-io/provenance/x/ledger/types"
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
	ledgerClass := ledger.LedgerClass{
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
	err := s.keeper.AddClassEntryType(s.ctx, s.validLedgerClass.LedgerClassId, ledger.LedgerClassEntryType{
		Id:          1,
		Code:        "SCHEDULED_PAYMENT",
		Description: "Scheduled Payment",
	})
	s.Require().Error(err, "AddClassEntryType error")
	s.Require().ErrorIs(err, ledger.ErrUnauthorized, "AddClassEntryType error")

	err = s.keeper.AddClassBucketType(s.ctx, s.validLedgerClass.LedgerClassId, ledger.LedgerClassBucketType{
		Id:          1,
		Code:        "PRINCIPAL",
		Description: "Principal",
	})
	s.Require().Error(err, "AddClassBucketType error")
	s.Require().Contains(err.Error(), ledger.ErrCodeUnauthorized, "AddClassBucketType error")

	err = s.keeper.AddClassStatusType(s.ctx, s.validLedgerClass.LedgerClassId, ledger.LedgerClassStatusType{
		Id:          1,
		Code:        "IN_REPAYMENT",
		Description: "In Repayment",
	})
	s.Require().Error(err, "AddClassStatusType error")
	s.Require().Contains(err.Error(), ledger.ErrCodeUnauthorized, "AddClassStatusType error")
}

// Test to ensure only the registered servicer or owner can create a ledger.
func (s *TestSuite) TestCreateLedgerNotOwnerOrServicer() {
	s.T().Skip("Skipping test - authorization logic moved out of keeper")
	ledger := ledger.Ledger{
		Key: &ledger.LedgerKey{
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
		ledgerClass ledger.LedgerClass
		expErr      []string
	}{
		{
			name: "valid ledger class should already exist",
			ledgerClass: ledger.LedgerClass{
				LedgerClassId:     s.validLedgerClass.LedgerClassId,
				AssetClassId:      s.validLedgerClass.AssetClassId,
				MaintainerAddress: s.addr1.String(),
				Denom:             s.bondDenom,
			},
			expErr: []string{"already exists"},
		},
		{
			name: "invalid asset class id",
			ledgerClass: ledger.LedgerClass{
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
		ledger   ledger.Ledger
		expErr   []error
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
			name: "duplicate ledger",
			ledger: ledger.Ledger{
				Key: &ledger.LedgerKey{
					AssetClassId: s.validNFTClass.Id,
					NftId:        s.validNFT.Id,
				},
				LedgerClassId: s.validLedgerClass.LedgerClassId,
				StatusTypeId:  1,
			},
			expErr: []error{ledger.ErrAlreadyExists},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Clear events before each test
			s.ctx.EventManager().Events()

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
	validLedger := ledger.Ledger{
		Key: &ledger.LedgerKey{
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
		expLedger *ledger.Ledger
	}{
		{
			name:      "valid ledger retrieval",
			nftId:     s.validNFT.Id,
			expLedger: &validLedger,
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

	err := s.keeper.AddLedger(s.ctx, l)
	s.Require().NoError(err, "CreateLedger error")

	// Test cases
	tests := []struct {
		name          string
		key           *ledger.LedgerKey
		correlationId string
		expEntry      *ledger.LedgerEntry
		expErr        *types.ErrCode
	}{
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

func (s *TestSuite) TestBech32() {
	ledgerKey := &ledger.LedgerKey{
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
	validLedger := ledger.Ledger{
		Key: &ledger.LedgerKey{
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
			s.ctx.EventManager().Events()

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
	validLedger := ledger.Ledger{
		Key: &ledger.LedgerKey{
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
			s.ctx.EventManager().Events()

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
	validLedger := ledger.Ledger{
		Key: &ledger.LedgerKey{
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
		nextPmtAmt       int64
		nextPmtDate      int32
		paymentFrequency types.PaymentFrequency
		expErr           []string
		expEvent         bool
	}{
		{
			name:             "valid payment update",
			nextPmtAmt:       1000000,  // 1000 tokens
			nextPmtDate:      20241201, // Dec 1, 2024
			paymentFrequency: types.PAYMENT_FREQUENCY_MONTHLY,
			expEvent:         true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Clear events before each test
			s.ctx.EventManager().Events()

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
	validLedger := ledger.Ledger{
		Key: &ledger.LedgerKey{
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
			s.ctx.EventManager().Events()

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
