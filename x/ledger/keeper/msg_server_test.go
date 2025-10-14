package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/errors"
	"cosmossdk.io/math"
	"cosmossdk.io/x/nft"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/ledger/helper"
	"github.com/provenance-io/provenance/x/ledger/keeper"
	ledger "github.com/provenance-io/provenance/x/ledger/types"
	registrytypes "github.com/provenance-io/provenance/x/registry/types"
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
	now              time.Time
	validLedgerClass ledger.LedgerClass
	validNFTClass    nft.Class
	validNFT         nft.NFT
	validNFT2        nft.NFT
	validAddress1    sdk.AccAddress
	validAddress2    sdk.AccAddress

	existingLedger *ledger.Ledger
}

func (s *MsgServerTestSuite) SetupTest() {
	s.app = app.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(true)
	s.keeper = s.app.LedgerKeeper

	var err error
	s.bondDenom, err = s.app.StakingKeeper.BondDenom(s.ctx)
	s.Require().NoError(err, "app.StakingKeeper.BondDenom(s.ctx)")

	// Create a timestamp 24 hours in the past to avoid future date errors
	s.now = time.Date(2021, 6, 9, 16, 20, 0, 0, time.UTC)
	s.pastDate = helper.DaysSinceEpoch(s.now.Add(-24 * time.Hour).UTC())

	s.validAddress1, err = sdk.AccAddressFromBech32("cosmos156vfr4kpaa0f07y673awf3u87eeemygldsjwu5")
	s.Require().NoError(err, "AccAddressFromBech32 error")

	s.validAddress2, err = sdk.AccAddressFromBech32("cosmos1ze3f954mtj30st8dw2qhylfvvtdv5q6x0e4k4q")
	s.Require().NoError(err, "AccAddressFromBech32 error")

	err = testutil.FundAccount(s.ctx, s.app.BankKeeper, s.validAddress1, sdk.NewCoins(sdk.NewCoin(s.bondDenom, math.NewInt(1000000000000000000))))
	s.Require().NoError(err, "FundAccount error")

	// Load the test ledger class configs
	s.ConfigureTest()
}

func (s *MsgServerTestSuite) ConfigureTest() {
	s.ctx = s.ctx.WithBlockTime(s.now)

	s.validNFTClass = nft.Class{
		Id: "test-nft-class-id",
	}
	err := s.app.NFTKeeper.SaveClass(s.ctx, s.validNFTClass)
	s.Require().NoError(err, "Save NFTClass error")

	s.validNFT = nft.NFT{
		ClassId: s.validNFTClass.Id,
		Id:      "test-nft-id",
	}
	err = s.app.NFTKeeper.Mint(s.ctx, s.validNFT, s.validAddress1)
	s.Require().NoError(err, "Mint NFT error")

	s.validNFT2 = nft.NFT{
		ClassId: s.validNFTClass.Id,
		Id:      "test-nft-id-2",
	}
	err = s.app.NFTKeeper.Mint(s.ctx, s.validNFT2, s.validAddress1)
	s.Require().NoError(err, "Mint NFT error")

	s.validLedgerClass = ledger.LedgerClass{
		LedgerClassId:     "test-ledger-class-id",
		AssetClassId:      s.validNFTClass.Id,
		MaintainerAddress: s.validAddress1.String(),
		Denom:             s.bondDenom,
	}
	err = s.keeper.AddLedgerClass(s.ctx, s.validLedgerClass)
	s.Require().NoError(err, "AddLedgerClass error")

	err = s.keeper.AddClassEntryType(s.ctx, s.validLedgerClass.LedgerClassId, ledger.LedgerClassEntryType{
		Id:          1,
		Code:        "SCHEDULED_PAYMENT",
		Description: "Scheduled Payment",
	})
	s.Require().NoError(err, "AddClassEntryType error")

	err = s.keeper.AddClassEntryType(s.ctx, s.validLedgerClass.LedgerClassId, ledger.LedgerClassEntryType{
		Id:          2,
		Code:        "DISBURSEMENT",
		Description: "Disbursement",
	})
	s.Require().NoError(err, "AddClassEntryType error")

	err = s.keeper.AddClassEntryType(s.ctx, s.validLedgerClass.LedgerClassId, ledger.LedgerClassEntryType{
		Id:          3,
		Code:        "ORIGINATION_FEE",
		Description: "Origination Fee",
	})
	s.Require().NoError(err, "AddClassEntryType error")

	err = s.keeper.AddClassBucketType(s.ctx, s.validLedgerClass.LedgerClassId, ledger.LedgerClassBucketType{
		Id:          1,
		Code:        "PRINCIPAL",
		Description: "Principal",
	})
	s.Require().NoError(err, "AddClassBucketType error")

	err = s.keeper.AddClassBucketType(s.ctx, s.validLedgerClass.LedgerClassId, ledger.LedgerClassBucketType{
		Id:          2,
		Code:        "INTEREST",
		Description: "Interest",
	})
	s.Require().NoError(err, "AddClassBucketType error")

	err = s.keeper.AddClassBucketType(s.ctx, s.validLedgerClass.LedgerClassId, ledger.LedgerClassBucketType{
		Id:          3,
		Code:        "ESCROW",
		Description: "Escrow",
	})
	s.Require().NoError(err, "AddClassBucketType error")

	err = s.keeper.AddClassStatusType(s.ctx, s.validLedgerClass.LedgerClassId, ledger.LedgerClassStatusType{
		Id:          1,
		Code:        "IN_REPAYMENT",
		Description: "In Repayment",
	})
	s.Require().NoError(err, "AddClassStatusType error")

	err = s.keeper.AddClassStatusType(s.ctx, s.validLedgerClass.LedgerClassId, ledger.LedgerClassStatusType{
		Id:          2,
		Code:        "IN_DEFERMENT",
		Description: "In Deferment",
	})
	s.Require().NoError(err, "AddClassStatusType error")

	s.existingLedger = &ledger.Ledger{
		Key: &ledger.LedgerKey{
			AssetClassId: s.validNFTClass.Id,
			NftId:        s.validNFT.Id,
		},
		LedgerClassId: s.validLedgerClass.LedgerClassId,
		StatusTypeId:  1,
	}
	err = s.keeper.AddLedger(s.ctx, *s.existingLedger)
	s.Require().NoError(err, "AddLedger error")
}

// TestAppend tests the Append message server method
func (s *MsgServerTestSuite) TestAppend() {
	tests := []struct {
		name    string
		req     *ledger.MsgAppendRequest
		expErr  error
		expResp *ledger.MsgAppendResponse
	}{
		{
			name: "successful append",
			req: &ledger.MsgAppendRequest{
				Key: s.existingLedger.Key,
				Entries: []*ledger.LedgerEntry{
					{
						CorrelationId:  "test-correlation-id-append",
						EntryTypeId:    1,
						PostedDate:     s.pastDate,
						EffectiveDate:  s.pastDate,
						TotalAmt:       math.NewInt(100),
						AppliedAmounts: []*ledger.LedgerBucketAmount{{AppliedAmt: math.NewInt(100), BucketTypeId: 1}},
						BalanceAmounts: []*ledger.BucketBalance{{BucketTypeId: 1, BalanceAmt: math.NewInt(100)}},
					},
				},
				Signer: s.validAddress1.String(),
			},
			expResp: &ledger.MsgAppendResponse{},
		},
		{
			name: "unauthorized append",
			req: &ledger.MsgAppendRequest{
				Key: s.existingLedger.Key,
				Entries: []*ledger.LedgerEntry{
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
						CorrelationId: "test-correlation-id-unauthorized",
					},
				},
				Signer: "cosmos1invalid",
			},
			expErr: ledger.ErrUnauthorized,
		},
		{
			name: "invalid entry type",
			req: &ledger.MsgAppendRequest{
				Key: s.existingLedger.Key,
				Entries: []*ledger.LedgerEntry{
					{
						EntryTypeId:   999, // Invalid entry type
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
				Signer: s.validAddress1.String(),
			},
			expErr: ledger.ErrInvalidField,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			msgServer := keeper.NewMsgServer(s.keeper)
			resp, err := msgServer.Append(s.ctx, tc.req)

			if tc.expErr != nil {
				s.Require().Error(err, "Append should fail")
				s.Require().ErrorIs(err, tc.expErr)
				s.Require().Nil(resp, "response should be nil on error")
			} else {
				s.Require().NoError(err, "Append should succeed")
				s.Require().Equal(tc.expResp, resp, "response should match expected")
			}
		})
	}
}

// TestUpdateBalances tests the UpdateBalances message server method
func (s *MsgServerTestSuite) TestUpdateBalances() {
	// Add an entry first
	entry := ledger.LedgerEntry{
		CorrelationId:  "test-correlation-id-update-balances",
		EntryTypeId:    1,
		PostedDate:     s.pastDate,
		EffectiveDate:  s.pastDate,
		TotalAmt:       math.NewInt(100),
		AppliedAmounts: []*ledger.LedgerBucketAmount{{AppliedAmt: math.NewInt(100), BucketTypeId: 1}},
		BalanceAmounts: []*ledger.BucketBalance{{BucketTypeId: 1, BalanceAmt: math.NewInt(1000000)}},
	}

	err := s.keeper.AppendEntries(s.ctx, s.existingLedger.Key, []*ledger.LedgerEntry{&entry})
	s.Require().NoError(err, "AppendEntry error")

	tests := []struct {
		name    string
		req     *ledger.MsgUpdateBalancesRequest
		expErr  string
		expResp *ledger.MsgUpdateBalancesResponse
	}{
		{
			name: "successful update balances",
			req: &ledger.MsgUpdateBalancesRequest{
				Key:            s.existingLedger.Key,
				Signer:         s.validAddress1.String(),
				CorrelationId:  "test-correlation-id-update-balances",
				TotalAmt:       math.NewInt(100),
				AppliedAmounts: []*ledger.LedgerBucketAmount{{AppliedAmt: math.NewInt(100), BucketTypeId: 1}},
				BalanceAmounts: []*ledger.BucketBalance{{BucketTypeId: 1, BalanceAmt: math.NewInt(200)}},
			},
			expResp: &ledger.MsgUpdateBalancesResponse{},
		},
		{
			name: "unauthorized update balances",
			req: &ledger.MsgUpdateBalancesRequest{
				Key:            s.existingLedger.Key,
				Signer:         s.validAddress2.String(),
				CorrelationId:  "test-correlation-id-update-balances",
				TotalAmt:       math.NewInt(200),
				AppliedAmounts: []*ledger.LedgerBucketAmount{{AppliedAmt: math.NewInt(200), BucketTypeId: 1}},
				BalanceAmounts: []*ledger.BucketBalance{{BucketTypeId: 1, BalanceAmt: math.NewInt(200)}},
			},
			expErr: "unauthorized access: signer is not the nft owner: unauthorized",
		},
		{
			name: "non-existent entry",
			req: &ledger.MsgUpdateBalancesRequest{
				Key:            s.existingLedger.Key,
				Signer:         s.validAddress1.String(),
				CorrelationId:  "non-existent-correlation-id",
				TotalAmt:       math.NewInt(200),
				AppliedAmounts: []*ledger.LedgerBucketAmount{{AppliedAmt: math.NewInt(200), BucketTypeId: 1}},
				BalanceAmounts: []*ledger.BucketBalance{{BucketTypeId: 1, BalanceAmt: math.NewInt(200)}},
			},
			expErr: "entry not found: not found",
		},
		{
			name: "incorrect applied amounts total",
			req: &ledger.MsgUpdateBalancesRequest{
				Key:           s.existingLedger.Key,
				Signer:        s.validAddress1.String(),
				CorrelationId: "test-correlation-id-update-balances",
				TotalAmt:       math.NewInt(100),
				AppliedAmounts: []*ledger.LedgerBucketAmount{
					{AppliedAmt: math.NewInt(50), BucketTypeId: 1},
					{AppliedAmt: math.NewInt(-49), BucketTypeId: 2},
				},
				BalanceAmounts: []*ledger.BucketBalance{{BucketTypeId: 1, BalanceAmt: math.NewInt(200)}},
			},
			expErr: "applied_amounts: total amount must equal sum of abs(applied amounts)",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			msgServer := keeper.NewMsgServer(s.keeper)
			resp, err := msgServer.UpdateBalances(s.ctx, tc.req)

			if len(tc.expErr) != 0 {
				s.Assert().EqualError(err, tc.expErr, "UpdateBalances should fail")
				s.Assert().Nil(resp, "response should be nil on error")
			} else {
				s.Assert().NoError(err, "UpdateBalances should succeed")
				s.Assert().Equal(tc.expResp, resp, "response should match expected")
			}
		})
	}
}

// TestCreate tests the Create message server method
func (s *MsgServerTestSuite) TestCreate() {
	nftOwner := s.validAddress1
	nftServicer := s.validAddress2

	tests := []struct {
		name            string
		mintNFTs        []nft.NFT
		registryEntries []registrytypes.RolesEntry
		req             *ledger.MsgCreateLedgerRequest
		expResp         *ledger.MsgCreateLedgerResponse
		expErr          error
	}{
		{
			name: "successful create no registry",
			mintNFTs: []nft.NFT{
				{
					ClassId: s.validNFTClass.Id,
					Id:      s.validNFT.Id + "1",
				},
			},
			req: &ledger.MsgCreateLedgerRequest{
				// Use a new ledger key to avoid already exist errors
				Ledger: &ledger.Ledger{
					Key: &ledger.LedgerKey{
						AssetClassId: s.validNFTClass.Id,
						NftId:        s.validNFT.Id + "1",
					},
					LedgerClassId: s.validLedgerClass.LedgerClassId,
					StatusTypeId:  1,
				},
				Signer: nftOwner.String(),
			},
			expResp: &ledger.MsgCreateLedgerResponse{},
		},
		{
			name: "successful create with registry",
			mintNFTs: []nft.NFT{
				{
					ClassId: s.validNFTClass.Id,
					Id:      s.validNFT.Id + "2",
				},
			},
			registryEntries: []registrytypes.RolesEntry{
				{
					Role:      registrytypes.RegistryRole_REGISTRY_ROLE_SERVICER,
					Addresses: []string{nftServicer.String()},
				},
			},
			req: &ledger.MsgCreateLedgerRequest{
				// Use a new ledger key to avoid already exist errors
				Ledger: &ledger.Ledger{
					Key: &ledger.LedgerKey{
						AssetClassId: s.validNFTClass.Id,
						NftId:        s.validNFT.Id + "2",
					},
					LedgerClassId: s.validLedgerClass.LedgerClassId,
					StatusTypeId:  1,
				},
				// Note that we authorize with the servicer address, not the owner address
				Signer: nftServicer.String(),
			},
			expResp: &ledger.MsgCreateLedgerResponse{},
		},
		{
			name: "unauthorized create no registry",
			mintNFTs: []nft.NFT{
				{
					ClassId: s.validNFTClass.Id,
					Id:      s.validNFT.Id + "3",
				},
			},
			req: &ledger.MsgCreateLedgerRequest{
				Ledger: &ledger.Ledger{
					Key: &ledger.LedgerKey{
						AssetClassId: s.validNFTClass.Id,
						NftId:        s.validNFT.Id + "3",
					},
					LedgerClassId: s.validLedgerClass.LedgerClassId,
					StatusTypeId:  1,
				},
				// Servicer shouldn't be able to create a ledger because there is no registry/role
				Signer: nftServicer.String(),
			},
			expErr: ledger.ErrUnauthorized,
		},
		{
			name: "unauthorized create with registry",
			mintNFTs: []nft.NFT{
				{
					ClassId: s.validNFTClass.Id,
					Id:      s.validNFT.Id + "4",
				},
			},
			registryEntries: []registrytypes.RolesEntry{
				{
					Role:      registrytypes.RegistryRole_REGISTRY_ROLE_SERVICER,
					Addresses: []string{nftServicer.String()},
				},
			},
			req: &ledger.MsgCreateLedgerRequest{
				Ledger: &ledger.Ledger{
					Key: &ledger.LedgerKey{
						AssetClassId: s.validNFTClass.Id,
						NftId:        s.validNFT.Id + "4",
					},
					LedgerClassId: s.validLedgerClass.LedgerClassId,
					StatusTypeId:  1,
				},
				// Owner has given servicing off to a servicer, so they shouldn't be able to create a ledger
				Signer: nftOwner.String(),
			},
			expErr: ledger.ErrUnauthorized,
		},
		{
			name: "duplicate ledger",
			req: &ledger.MsgCreateLedgerRequest{
				Ledger: s.existingLedger,
				Signer: nftOwner.String(),
			},
			expErr: ledger.ErrAlreadyExists,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Create the NFTs with valid address 1
			for _, nft := range tc.mintNFTs {
				s.app.NFTKeeper.Mint(s.ctx, nft, nftOwner)

				// Associate all roles with the created NFT
				registryKey := registrytypes.RegistryKey{
					AssetClassId: nft.ClassId,
					NftId:        nft.Id,
				}

				// Create a registry if there are roles to grant
				if len(tc.registryEntries) > 0 {
					err := s.app.RegistryKeeper.CreateRegistry(s.ctx, &registryKey, tc.registryEntries)
					s.Require().NoError(err, "CreateRegistry error")
				}
			}

			msgServer := keeper.NewMsgServer(s.keeper)
			resp, err := msgServer.CreateLedger(s.ctx, tc.req)

			if tc.expErr != nil {
				s.Require().ErrorIs(err, tc.expErr)
				s.Require().Nil(resp, "response should be nil on error")
			} else {
				s.Require().NoError(err, "Create should succeed")
				s.Require().Equal(tc.expResp, resp, "response should match expected")
			}
		})
	}
}

// TestUpdateStatus tests the UpdateStatus message server method
func (s *MsgServerTestSuite) TestUpdateStatus() {
	tests := []struct {
		name    string
		req     *ledger.MsgUpdateStatusRequest
		expErr  *errors.Error
		expResp *ledger.MsgUpdateStatusResponse
	}{
		{
			name: "successful update status",
			req: &ledger.MsgUpdateStatusRequest{
				Key:          s.existingLedger.Key,
				Signer:       s.validAddress1.String(),
				StatusTypeId: 2,
			},
			expResp: &ledger.MsgUpdateStatusResponse{},
		},
		{
			name: "unauthorized update status",
			req: &ledger.MsgUpdateStatusRequest{
				Key:          s.existingLedger.Key,
				Signer:       s.validAddress2.String(),
				StatusTypeId: 2,
			},
			expErr: ledger.ErrUnauthorized,
		},
		{
			name: "non-existent ledger",
			req: &ledger.MsgUpdateStatusRequest{
				Key: &ledger.LedgerKey{
					AssetClassId: s.validNFTClass.Id,
					NftId:        "non-existent-nft",
				},
				Signer:       s.validAddress1.String(),
				StatusTypeId: 2,
			},
			expErr: ledger.ErrNotFound,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			msgServer := keeper.NewMsgServer(s.keeper)

			var combinedErr error
			var resp *ledger.MsgUpdateStatusResponse
			if err := tc.req.ValidateBasic(); err != nil {
				combinedErr = err
			} else {
				resp, combinedErr = msgServer.UpdateStatus(s.ctx, tc.req)
			}

			if tc.expErr != nil {
				s.Require().Error(combinedErr, "UpdateStatus should fail")
				s.Require().ErrorIs(combinedErr, tc.expErr)
				s.Require().Nil(resp, "response should be nil on error")
			} else {
				s.Require().NoError(combinedErr, "UpdateStatus should succeed")
				s.Require().Equal(tc.expResp, resp, "response should match expected")
			}
		})
	}
}

// TestUpdateInterestRate tests the UpdateInterestRate message server method
func (s *MsgServerTestSuite) TestUpdateInterestRate() {
	tests := []struct {
		name    string
		req     *ledger.MsgUpdateInterestRateRequest
		expErr  error
		expResp *ledger.MsgUpdateInterestRateResponse
	}{
		{
			name: "successful update interest rate",
			req: &ledger.MsgUpdateInterestRateRequest{
				Key:                        s.existingLedger.Key,
				Signer:                     s.validAddress1.String(),
				InterestRate:               5000000, // 5%
				InterestDayCountConvention: ledger.DAY_COUNT_CONVENTION_ACTUAL_365,
				InterestAccrualMethod:      ledger.INTEREST_ACCRUAL_METHOD_SIMPLE_INTEREST,
			},
			expResp: &ledger.MsgUpdateInterestRateResponse{},
		},
		{
			name: "unauthorized update interest rate",
			req: &ledger.MsgUpdateInterestRateRequest{
				Key:                        s.existingLedger.Key,
				Signer:                     "cosmos1invalid",
				InterestRate:               5000000,
				InterestDayCountConvention: ledger.DAY_COUNT_CONVENTION_ACTUAL_365,
				InterestAccrualMethod:      ledger.INTEREST_ACCRUAL_METHOD_SIMPLE_INTEREST,
			},
			expErr: ledger.ErrUnauthorized,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			msgServer := keeper.NewMsgServer(s.keeper)
			resp, err := msgServer.UpdateInterestRate(s.ctx, tc.req)

			if tc.expErr != nil {
				s.Require().Error(err, "UpdateInterestRate should fail")
				s.Require().ErrorIs(err, tc.expErr)
				s.Require().Nil(resp, "response should be nil on error")
			} else {
				s.Require().NoError(err, "UpdateInterestRate should succeed")
				s.Require().Equal(tc.expResp, resp, "response should match expected")
			}
		})
	}
}

// TestUpdatePayment tests the UpdatePayment message server method
func (s *MsgServerTestSuite) TestUpdatePayment() {
	nextPmtDate, err := helper.ParseYMD("2024-01-15")
	s.Require().NoError(err, "ParseYMD error")

	tests := []struct {
		name    string
		req     *ledger.MsgUpdatePaymentRequest
		expErr  error
		expResp *ledger.MsgUpdatePaymentResponse
	}{
		{
			name: "successful update payment",
			req: &ledger.MsgUpdatePaymentRequest{
				Key:              s.existingLedger.Key,
				Signer:           s.validAddress1.String(),
				NextPmtAmt:       math.NewInt(1000),
				NextPmtDate:      helper.DaysSinceEpoch(nextPmtDate),
				PaymentFrequency: ledger.PAYMENT_FREQUENCY_MONTHLY,
			},
			expResp: &ledger.MsgUpdatePaymentResponse{},
		},
		{
			name: "unauthorized update payment",
			req: &ledger.MsgUpdatePaymentRequest{
				Key:              s.existingLedger.Key,
				Signer:           "cosmos1invalid",
				NextPmtAmt:       math.NewInt(1000),
				NextPmtDate:      helper.DaysSinceEpoch(nextPmtDate),
				PaymentFrequency: ledger.PAYMENT_FREQUENCY_MONTHLY,
			},
			expErr: ledger.ErrUnauthorized,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			msgServer := keeper.NewMsgServer(s.keeper)
			resp, err := msgServer.UpdatePayment(s.ctx, tc.req)

			if tc.expErr != nil {
				s.Require().Error(err, "UpdatePayment should fail")
				s.Require().ErrorIs(err, tc.expErr)
				s.Require().Nil(resp, "response should be nil on error")
			} else {
				s.Require().NoError(err, "UpdatePayment should succeed")
				s.Require().Equal(tc.expResp, resp, "response should match expected")
			}
		})
	}
}

// TestUpdateMaturityDate tests the UpdateMaturityDate message server method
func (s *MsgServerTestSuite) TestUpdateMaturityDate() {
	maturityDate, err := helper.ParseYMD("2025-12-31")
	s.Require().NoError(err, "ParseYMD error")

	tests := []struct {
		name    string
		req     *ledger.MsgUpdateMaturityDateRequest
		expErr  error
		expResp *ledger.MsgUpdateMaturityDateResponse
	}{
		{
			name: "successful update maturity date",
			req: &ledger.MsgUpdateMaturityDateRequest{
				Key:          s.existingLedger.Key,
				Signer:       s.validAddress1.String(),
				MaturityDate: helper.DaysSinceEpoch(maturityDate),
			},
			expResp: &ledger.MsgUpdateMaturityDateResponse{},
		},
		{
			name: "unauthorized update maturity date",
			req: &ledger.MsgUpdateMaturityDateRequest{
				Key:          s.existingLedger.Key,
				Signer:       "cosmos1invalid",
				MaturityDate: helper.DaysSinceEpoch(maturityDate),
			},
			expErr: ledger.ErrUnauthorized,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			msgServer := keeper.NewMsgServer(s.keeper)
			resp, err := msgServer.UpdateMaturityDate(s.ctx, tc.req)

			if tc.expErr != nil {
				s.Require().Error(err, "UpdateMaturityDate should fail")
				s.Require().ErrorIs(err, tc.expErr)
				s.Require().Nil(resp, "response should be nil on error")
			} else {
				s.Require().NoError(err, "UpdateMaturityDate should succeed")
				s.Require().Equal(tc.expResp, resp, "response should match expected")
			}
		})
	}
}

// TestTransferFundsWithSettlement tests the TransferFundsWithSettlement message server method
func (s *MsgServerTestSuite) TestTransferFundsWithSettlement() {
	s.keeper.AppendEntries(s.ctx, s.existingLedger.Key, []*ledger.LedgerEntry{
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
			BalanceAmounts: []*ledger.BucketBalance{
				{
					BucketTypeId: 1,
					BalanceAmt:   math.NewInt(100),
				},
			},
			CorrelationId: "1",
		},
	})

	tests := []struct {
		name    string
		req     *ledger.MsgTransferFundsWithSettlementRequest
		expErr  error
		expResp *ledger.MsgTransferFundsWithSettlementResponse
	}{
		{
			name: "successful transfer funds with settlement",
			req: &ledger.MsgTransferFundsWithSettlementRequest{
				Signer: s.validAddress1.String(),
				Transfers: []*ledger.FundTransferWithSettlement{
					{
						Key:                      s.existingLedger.Key,
						LedgerEntryCorrelationId: "1",
						SettlementInstructions: []*ledger.SettlementInstruction{
							{
								Amount: sdk.Coin{
									Denom:  s.bondDenom,
									Amount: math.NewInt(1000),
								},
								RecipientAddress: s.validAddress2.String(),
								Status:           ledger.FundingTransferStatus_FUNDING_TRANSFER_STATUS_PENDING,
								Memo:             "test transfer",
							},
						},
					},
				},
			},
			expResp: &ledger.MsgTransferFundsWithSettlementResponse{},
		},
		{
			name: "unauthorized transfer funds",
			req: &ledger.MsgTransferFundsWithSettlementRequest{
				Signer: s.validAddress2.String(),
				Transfers: []*ledger.FundTransferWithSettlement{
					{
						Key:                      s.existingLedger.Key,
						LedgerEntryCorrelationId: "1",
						SettlementInstructions: []*ledger.SettlementInstruction{
							{
								Amount: sdk.Coin{
									Denom:  s.bondDenom,
									Amount: math.NewInt(100),
								},
								RecipientAddress: s.validAddress2.String(),
								Status:           ledger.FundingTransferStatus_FUNDING_TRANSFER_STATUS_PENDING,
								Memo:             "test transfer",
							},
						},
					},
				},
			},
			expErr: ledger.ErrUnauthorized,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			msgServer := keeper.NewMsgServer(s.keeper)
			resp, err := msgServer.TransferFundsWithSettlement(s.ctx, tc.req)

			if tc.expErr != nil {
				s.Require().Error(err, "TransferFundsWithSettlement should fail")
				s.Require().ErrorIs(err, tc.expErr)
				s.Require().Nil(resp, "response should be nil on error")
			} else {
				s.Require().NoError(err, "TransferFundsWithSettlement should succeed")
				s.Require().Equal(tc.expResp, resp, "response should match expected")
			}
		})
	}
}

// TestDestroy tests the Destroy message server method
func (s *MsgServerTestSuite) TestDestroy() {
	tests := []struct {
		name    string
		req     *ledger.MsgDestroyRequest
		expErr  error
		expResp *ledger.MsgDestroyResponse
	}{
		{
			name: "unauthorized destroy",
			req: &ledger.MsgDestroyRequest{
				Key:    s.existingLedger.Key,
				Signer: s.validAddress2.String(),
			},
			expErr: ledger.ErrUnauthorized,
		},
		{
			name: "successful destroy",
			req: &ledger.MsgDestroyRequest{
				Key:    s.existingLedger.Key,
				Signer: s.validAddress1.String(),
			},
			expResp: &ledger.MsgDestroyResponse{},
		},
		{
			name: "non-existent ledger",
			req: &ledger.MsgDestroyRequest{
				Key: &ledger.LedgerKey{
					AssetClassId: s.validNFTClass.Id,
					NftId:        "non-existent-nft",
				},
				Signer: s.validAddress1.String(),
			},
			expErr: ledger.ErrNotFound,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			msgServer := keeper.NewMsgServer(s.keeper)
			resp, err := msgServer.Destroy(s.ctx, tc.req)

			if tc.expErr != nil {
				s.Require().Error(err, "Destroy should fail")
				s.Require().ErrorIs(err, tc.expErr)
				s.Require().Nil(resp, "response should be nil on error")
			} else {
				s.Require().NoError(err, "Destroy should succeed")
				s.Require().Equal(tc.expResp, resp, "response should match expected")
			}
		})
	}
}

// TestCreateLedgerClass tests the CreateLedgerClass message server method
func (s *MsgServerTestSuite) TestCreateLedgerClass() {
	authorizedAddr := s.validAddress1

	tests := []struct {
		name    string
		req     *ledger.MsgCreateLedgerClassRequest
		expErr  error
		expResp *ledger.MsgCreateLedgerClassResponse
	}{
		{
			name: "successful create ledger class",
			req: &ledger.MsgCreateLedgerClassRequest{
				LedgerClass: &ledger.LedgerClass{
					LedgerClassId:     "test-ledger-class-new",
					AssetClassId:      s.validNFTClass.Id,
					MaintainerAddress: authorizedAddr.String(),
					Denom:             s.bondDenom,
				},
				Signer: authorizedAddr.String(),
			},
			expResp: &ledger.MsgCreateLedgerClassResponse{},
		},
		{
			name: "duplicate ledger class",
			req: &ledger.MsgCreateLedgerClassRequest{
				LedgerClass: &ledger.LedgerClass{
					LedgerClassId:     s.validLedgerClass.LedgerClassId,
					AssetClassId:      s.validNFTClass.Id,
					MaintainerAddress: authorizedAddr.String(),
					Denom:             s.bondDenom,
				},
				Signer: authorizedAddr.String(),
			},
			expErr: ledger.ErrAlreadyExists,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			msgServer := keeper.NewMsgServer(s.keeper)
			resp, err := msgServer.CreateLedgerClass(s.ctx, tc.req)

			if tc.expErr != nil {
				s.Require().Error(err, "CreateLedgerClass should fail")
				s.Require().ErrorIs(err, tc.expErr)
				s.Require().Nil(resp, "response should be nil on error")
			} else {
				s.Require().NoError(err, "CreateLedgerClass should succeed")
				s.Require().Equal(tc.expResp, resp, "response should match expected")
			}
		})
	}
}

// TestAddLedgerClassStatusType tests the AddLedgerClassStatusType message server method
func (s *MsgServerTestSuite) TestAddLedgerClassStatusType() {
	authorizedAddr := s.validAddress1

	tests := []struct {
		name    string
		req     *ledger.MsgAddLedgerClassStatusTypeRequest
		expErr  error
		expResp *ledger.MsgAddLedgerClassStatusTypeResponse
	}{
		{
			name: "successful add status type",
			req: &ledger.MsgAddLedgerClassStatusTypeRequest{
				LedgerClassId: s.validLedgerClass.LedgerClassId,
				StatusType: &ledger.LedgerClassStatusType{
					Id:          4,
					Code:        "COMPLETED",
					Description: "Completed",
				},
				Signer: authorizedAddr.String(),
			},
			expResp: &ledger.MsgAddLedgerClassStatusTypeResponse{},
		},
		{
			name: "unauthorized add status type",
			req: &ledger.MsgAddLedgerClassStatusTypeRequest{
				LedgerClassId: s.validLedgerClass.LedgerClassId,
				StatusType: &ledger.LedgerClassStatusType{
					Id:          5,
					Code:        "CANCELLED",
					Description: "Cancelled",
				},
				Signer: "cosmos1invalid",
			},
			expErr: ledger.ErrUnauthorized,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			msgServer := keeper.NewMsgServer(s.keeper)
			resp, err := msgServer.AddLedgerClassStatusType(s.ctx, tc.req)

			if tc.expErr != nil {
				s.Require().Error(err, "AddLedgerClassStatusType should fail")
				s.Require().ErrorIs(err, tc.expErr)
				s.Require().Nil(resp, "response should be nil on error")
			} else {
				s.Require().NoError(err, "AddLedgerClassStatusType should succeed")
				s.Require().Equal(tc.expResp, resp, "response should match expected")
			}
		})
	}
}

// TestAddLedgerClassEntryType tests the AddLedgerClassEntryType message server method
func (s *MsgServerTestSuite) TestAddLedgerClassEntryType() {
	authorizedAddr := s.validAddress1

	tests := []struct {
		name    string
		req     *ledger.MsgAddLedgerClassEntryTypeRequest
		expErr  error
		expResp *ledger.MsgAddLedgerClassEntryTypeResponse
	}{
		{
			name: "successful add entry type",
			req: &ledger.MsgAddLedgerClassEntryTypeRequest{
				LedgerClassId: s.validLedgerClass.LedgerClassId,
				EntryType: &ledger.LedgerClassEntryType{
					Id:          4,
					Code:        "ADJUSTMENT",
					Description: "Adjustment",
				},
				Signer: authorizedAddr.String(),
			},
			expResp: &ledger.MsgAddLedgerClassEntryTypeResponse{},
		},
		{
			name: "unauthorized add entry type",
			req: &ledger.MsgAddLedgerClassEntryTypeRequest{
				LedgerClassId: s.validLedgerClass.LedgerClassId,
				EntryType: &ledger.LedgerClassEntryType{
					Id:          5,
					Code:        "FEE",
					Description: "Fee",
				},
				Signer: "cosmos1invalid",
			},
			expErr: ledger.ErrUnauthorized,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			msgServer := keeper.NewMsgServer(s.keeper)
			resp, err := msgServer.AddLedgerClassEntryType(s.ctx, tc.req)

			if tc.expErr != nil {
				s.Require().Error(err, "AddLedgerClassEntryType should fail")
				s.Require().ErrorIs(err, tc.expErr)
				s.Require().Nil(resp, "response should be nil on error")
			} else {
				s.Require().NoError(err, "AddLedgerClassEntryType should succeed")
				s.Require().Equal(tc.expResp, resp, "response should match expected")
			}
		})
	}
}

// TestAddLedgerClassBucketType tests the AddLedgerClassBucketType message server method
func (s *MsgServerTestSuite) TestAddLedgerClassBucketType() {
	authorizedAddr := s.validAddress1

	tests := []struct {
		name    string
		req     *ledger.MsgAddLedgerClassBucketTypeRequest
		expErr  error
		expResp *ledger.MsgAddLedgerClassBucketTypeResponse
	}{
		{
			name: "successful add bucket type",
			req: &ledger.MsgAddLedgerClassBucketTypeRequest{
				LedgerClassId: s.validLedgerClass.LedgerClassId,
				BucketType: &ledger.LedgerClassBucketType{
					Id:          4,
					Code:        "FEES",
					Description: "Fees",
				},
				Signer: authorizedAddr.String(),
			},
			expResp: &ledger.MsgAddLedgerClassBucketTypeResponse{},
		},
		{
			name: "unauthorized add bucket type",
			req: &ledger.MsgAddLedgerClassBucketTypeRequest{
				LedgerClassId: s.validLedgerClass.LedgerClassId,
				BucketType: &ledger.LedgerClassBucketType{
					Id:          5,
					Code:        "PENALTIES",
					Description: "Penalties",
				},
				Signer: "cosmos1invalid",
			},
			expErr: ledger.ErrUnauthorized,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			msgServer := keeper.NewMsgServer(s.keeper)
			resp, err := msgServer.AddLedgerClassBucketType(s.ctx, tc.req)

			if tc.expErr != nil {
				s.Require().Error(err, "AddLedgerClassBucketType should fail")
				s.Require().ErrorIs(err, tc.expErr)
				s.Require().Nil(resp, "response should be nil on error")
			} else {
				s.Require().NoError(err, "AddLedgerClassBucketType should succeed")
				s.Require().Equal(tc.expResp, resp, "response should match expected")
			}
		})
	}
}

// TestBulkImport tests the BulkImport message server method
// func (s *MsgServerTestSuite) TestBulkImport() {
// 	tests := []struct {
// 		name    string
// 		req     *ledger.MsgBulkImportRequest
// 		expErr  string
// 		expResp *ledger.MsgBulkImportResponse
// 	}{
// 		{
// 			name: "successful bulk import",
// 			req: &ledger.MsgBulkImportRequest{
// 				GenesisState: &ledger.GenesisState{
// 					LedgerToEntries: []*ledger.LedgerToEntries{
// 						{
// 							Ledger: &ledger.Ledger{
// 								Key: &ledger.LedgerKey{
// 									AssetClassId: s.validNFTClass.Id,
// 									NftId:        "bulk-import-nft",
// 								},
// 								LedgerClassId: s.validLedgerClass.LedgerClassId,
// 								StatusTypeId:  1,
// 							},
// 							Entries: []*ledger.LedgerEntry{
// 								{
// 									EntryTypeId:   1,
// 									PostedDate:    s.pastDate,
// 									EffectiveDate: s.pastDate,
// 									TotalAmt:      math.NewInt(100),
// 									AppliedAmounts: []*ledger.LedgerBucketAmount{
// 										{
// 											AppliedAmt:   math.NewInt(100),
// 											BucketTypeId: 1,
// 										},
// 									},
// 									CorrelationId: "bulk-import-correlation-id",
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			expResp: &ledger.MsgBulkImportResponse{},
// 		},
// 		{
// 			name: "empty genesis state",
// 			req: &ledger.MsgBulkImportRequest{
// 				GenesisState: &ledger.GenesisState{},
// 			},
// 			expResp: &ledger.MsgBulkImportResponse{},
// 		},
// 	}

// 	for _, tc := range tests {
// 		s.Run(tc.name, func() {
// 			msgServer := keeper.NewMsgServer(s.keeper)
// 			resp, err := msgServer.BulkImport(s.ctx, tc.req)

// 			if tc.expErr != "" {
// 				s.Require().Error(err, "BulkImport should fail")
// 				s.Require().Contains(err.Error(), tc.expErr, "error message")
// 				s.Require().Nil(resp, "response should be nil on error")
// 			} else {
// 				s.Require().NoError(err, "BulkImport should succeed")
// 				s.Require().Equal(tc.expResp, resp, "response should match expected")
// 			}
// 		})
// 	}
// }

// TestAppendEntriesValidation tests validation logic for AppendEntries
func (s *MsgServerTestSuite) TestAppendEntriesValidation() {
	tests := []struct {
		name    string
		entries []*ledger.LedgerEntry
		expErr  error
	}{
		{
			name: "future posted date",
			entries: []*ledger.LedgerEntry{
				{
					EntryTypeId:   1,
					PostedDate:    helper.DaysSinceEpoch(s.now.Add(24 * time.Hour).UTC()),
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
			expErr: ledger.ErrInvalidField,
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
			expErr: ledger.ErrInvalidField,
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
			expErr: ledger.ErrNotFound,
		},
		{
			name: "valid entry",
			entries: []*ledger.LedgerEntry{
				{
					CorrelationId:  "test-correlation-id-valid",
					EntryTypeId:    1,
					PostedDate:     s.pastDate,
					EffectiveDate:  s.pastDate,
					TotalAmt:       math.NewInt(100),
					AppliedAmounts: []*ledger.LedgerBucketAmount{{AppliedAmt: math.NewInt(100), BucketTypeId: 1}},
					BalanceAmounts: []*ledger.BucketBalance{{BucketTypeId: 1, BalanceAmt: math.NewInt(100)}},
				},
			},
			expErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			ledgerKey := s.existingLedger.Key
			if tc.name == "non-existent ledger" {
				ledgerKey = &ledger.LedgerKey{
					AssetClassId: s.validNFTClass.Id,
					NftId:        "non-existent-nft",
				}
			}

			err := s.keeper.AppendEntries(s.ctx, ledgerKey, tc.entries)
			if tc.expErr != nil {
				s.Require().Error(err, "AppendEntries should fail")
				s.Require().ErrorIs(err, tc.expErr)
			} else {
				s.Require().NoError(err, "AppendEntries should succeed")
			}
		})
	}
}

// TestUpdateEntryBalancesValidation tests validation logic for UpdateEntryBalances
func (s *MsgServerTestSuite) TestUpdateEntryBalancesValidation() {
	// Add an entry first
	entry := ledger.LedgerEntry{
		CorrelationId:  "test-correlation-id-update",
		EntryTypeId:    1,
		PostedDate:     s.pastDate,
		EffectiveDate:  s.pastDate,
		TotalAmt:       math.NewInt(100),
		AppliedAmounts: []*ledger.LedgerBucketAmount{{AppliedAmt: math.NewInt(100), BucketTypeId: 1}},
		BalanceAmounts: []*ledger.BucketBalance{{BucketTypeId: 1, BalanceAmt: math.NewInt(10000000)}},
	}

	err := s.keeper.AppendEntries(s.ctx, s.existingLedger.Key, []*ledger.LedgerEntry{&entry})
	s.Require().NoError(err, "AppendEntry error")

	tests := []struct {
		name           string
		ledgerKey      *ledger.LedgerKey
		correlationId  string
		totalAmt       math.Int
		balanceAmounts []*ledger.BucketBalance
		appliedAmounts []*ledger.LedgerBucketAmount
		expErr         string
	}{
		{
			name:           "non-existent entry",
			ledgerKey:      s.existingLedger.Key,
			correlationId:  "non-existent-correlation-id",
			totalAmt:       math.NewInt(100),
			appliedAmounts: []*ledger.LedgerBucketAmount{{AppliedAmt: math.NewInt(100), BucketTypeId: 1}},
			balanceAmounts: []*ledger.BucketBalance{{BucketTypeId: 1, BalanceAmt: math.NewInt(200)}},
			expErr:         "entry not found: not found",
		},
		{
			name:           "incorrect applied amounts",
			ledgerKey:      s.existingLedger.Key,
			correlationId:  "test-correlation-id-update",
			totalAmt:       math.NewInt(199),
			appliedAmounts: []*ledger.LedgerBucketAmount{{AppliedAmt: math.NewInt(200), BucketTypeId: 1}},
			balanceAmounts: []*ledger.BucketBalance{{BucketTypeId: 1, BalanceAmt: math.NewInt(200)}},
			expErr:         "applied_amounts: total amount must equal sum of abs(applied amounts)",
		},
		{
			name:          "valid update: same total",
			ledgerKey:     s.existingLedger.Key,
			correlationId: "test-correlation-id-update",
			totalAmt:      math.NewInt(100),
			appliedAmounts: []*ledger.LedgerBucketAmount{
				{AppliedAmt: math.NewInt(75), BucketTypeId: 1},
				{AppliedAmt: math.NewInt(-25), BucketTypeId: 2},
			},
			balanceAmounts: []*ledger.BucketBalance{{BucketTypeId: 1, BalanceAmt: math.NewInt(200)}},
		},
		{
			name:          "valid update: new total",
			ledgerKey:     s.existingLedger.Key,
			correlationId: "test-correlation-id-update",
			totalAmt:      math.NewInt(150),
			appliedAmounts: []*ledger.LedgerBucketAmount{
				{AppliedAmt: math.NewInt(80), BucketTypeId: 1},
				{AppliedAmt: math.NewInt(-30), BucketTypeId: 2},
				{AppliedAmt: math.NewInt(40), BucketTypeId: 3},
			},
			balanceAmounts: []*ledger.BucketBalance{{BucketTypeId: 1, BalanceAmt: math.NewInt(9876543)}},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err = s.keeper.UpdateEntryBalances(s.ctx, tc.ledgerKey, tc.correlationId, tc.totalAmt, tc.appliedAmounts, tc.balanceAmounts)
			assertions.AssertErrorValue(s.T(), err, tc.expErr, "UpdateEntryBalances")

			// If we weren't expecting an error, check that the entry was actually updated.
			if len(tc.expErr) != 0 {
				return
			}
			curEntry, err := s.keeper.GetLedgerEntry(s.ctx, tc.ledgerKey, tc.correlationId)
			s.Require().NoError(err, "GetLedgerEntry error")
			s.Require().NotNil(curEntry, "GetLedgerEntry result")
			s.Assert().Equal(tc.totalAmt, curEntry.TotalAmt, "total amount")
			s.Assert().Equal(tc.appliedAmounts, curEntry.AppliedAmounts, "applied amounts")
			s.Assert().Equal(tc.balanceAmounts, curEntry.BalanceAmounts, "balance amounts")
		})
	}
}

// TestAppendEntriesMultipleValidation tests validation for multiple entries
func (s *MsgServerTestSuite) TestAppendEntriesMultipleValidation() {
	tests := []struct {
		name    string
		entries []*ledger.LedgerEntry
		expErr  error
	}{
		{
			name: "mixed valid and invalid entries",
			entries: []*ledger.LedgerEntry{
				{
					CorrelationId:  "test-correlation-id-valid-1",
					EntryTypeId:    1,
					PostedDate:     s.pastDate,
					EffectiveDate:  s.pastDate,
					TotalAmt:       math.NewInt(100),
					AppliedAmounts: []*ledger.LedgerBucketAmount{{AppliedAmt: math.NewInt(100), BucketTypeId: 1}},
					BalanceAmounts: []*ledger.BucketBalance{{BucketTypeId: 1, BalanceAmt: math.NewInt(100)}},
				},
				{
					CorrelationId:  "test-correlation-id-invalid-1",
					EntryTypeId:    999, // Invalid entry type
					PostedDate:     s.pastDate,
					EffectiveDate:  s.pastDate,
					TotalAmt:       math.NewInt(200),
					AppliedAmounts: []*ledger.LedgerBucketAmount{{AppliedAmt: math.NewInt(200), BucketTypeId: 1}},
					BalanceAmounts: []*ledger.BucketBalance{{BucketTypeId: 1, BalanceAmt: math.NewInt(100)}},
				},
			},
			expErr: ledger.ErrInvalidField,
		},
		{
			name: "all valid entries",
			entries: []*ledger.LedgerEntry{
				{
					CorrelationId:  "test-correlation-id-valid-2",
					EntryTypeId:    1,
					PostedDate:     s.pastDate,
					EffectiveDate:  s.pastDate,
					TotalAmt:       math.NewInt(100),
					AppliedAmounts: []*ledger.LedgerBucketAmount{{AppliedAmt: math.NewInt(100), BucketTypeId: 1}},
					BalanceAmounts: []*ledger.BucketBalance{{BucketTypeId: 1, BalanceAmt: math.NewInt(100)}},
				},
				{
					CorrelationId:  "test-correlation-id-valid-3",
					EntryTypeId:    2,
					PostedDate:     s.pastDate,
					EffectiveDate:  s.pastDate,
					TotalAmt:       math.NewInt(200),
					AppliedAmounts: []*ledger.LedgerBucketAmount{{AppliedAmt: math.NewInt(200), BucketTypeId: 1}},
					BalanceAmounts: []*ledger.BucketBalance{{BucketTypeId: 1, BalanceAmt: math.NewInt(100)}},
				},
			},
			expErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := s.keeper.AppendEntries(s.ctx, s.existingLedger.Key, tc.entries)
			if tc.expErr != nil {
				s.Require().Error(err, "AppendEntries should fail")
				s.Require().ErrorIs(err, tc.expErr)
			} else {
				s.Require().NoError(err, "AppendEntries should succeed")
			}
		})
	}
}

// TestAppendEntriesEmptyArray tests edge case of empty entries array
func (s *MsgServerTestSuite) TestAppendEntriesEmptyArray() {
	err := s.keeper.AppendEntries(s.ctx, s.existingLedger.Key, []*ledger.LedgerEntry{})
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

	// Test with empty ledger key
	emptyKey := &ledger.LedgerKey{
		AssetClassId: "",
		NftId:        "",
	}
	err := s.keeper.AppendEntries(s.ctx, emptyKey, []*ledger.LedgerEntry{&entry})
	s.Require().Error(err, "AppendEntries with empty ledger key should fail")
}
