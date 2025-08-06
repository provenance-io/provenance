package keeper_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/ledger/keeper"
	"github.com/provenance-io/provenance/x/ledger/types"
)

func TestMsgServerTestSuite(t *testing.T) {
	// suite.Run(t, new(MsgServerTestSuite))
}

type MsgServerTestSuite struct {
	suite.Suite
	app *simapp.App
	ctx sdk.Context
}

func (s *MsgServerTestSuite) SetupTest() {
	s.app = simapp.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(true)
}

const authority = "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn" // should equal s.app.LedgerKeeper.GetAuthority()

// MockKeeper is a fake x/ledger Keeper for use in the MsgServer.
type MockKeeper struct {
	ValidateAuthorityErrs  []string
	ValidateAuthorityExp   []string
	ValidateAuthorityArgs  []string
	AddLedgerErrs          []string
	AddLedgerExp           []*types.Ledger
	AddLedgerArgs          []*types.Ledger
	GetLedgerErrs          []string
	GetLedgerExp           []*types.Ledger
	GetLedgerArgs          []*types.LedgerKey
	UpdateLedgerStatusErrs []string
	UpdateLedgerStatusExp  []*types.LedgerKey
	UpdateLedgerStatusArgs []*types.LedgerKey
}

// var _ keeper.Keeper = (*MockKeeper)(nil)

// NewMockKeeper creates a new (ledger) MockKeeper for use in the msg-server tests.
func NewMockKeeper() *MockKeeper {
	return &MockKeeper{}
}

// WithValidateAuthorityErrs adds the provided errs to be returned from ValidateAuthority.
// An empty string indicates no error. This method both updates the receiver and returns it.
func (k *MockKeeper) WithValidateAuthorityErrs(errs ...string) *MockKeeper {
	k.ValidateAuthorityErrs = append(k.ValidateAuthorityErrs, errs...)
	return k
}

// WithAddLedgerErrs adds the provided errs to be returned from AddLedger.
// An empty string indicates no error. This method both updates the receiver and returns it.
func (k *MockKeeper) WithAddLedgerErrs(errs ...string) *MockKeeper {
	k.AddLedgerErrs = append(k.AddLedgerErrs, errs...)
	return k
}

// WithGetLedgerErrs adds the provided errs to be returned from GetLedger.
// An empty string indicates no error. This method both updates the receiver and returns it.
func (k *MockKeeper) WithGetLedgerErrs(errs ...string) *MockKeeper {
	k.GetLedgerErrs = append(k.GetLedgerErrs, errs...)
	return k
}

// WithUpdateLedgerStatusErrs adds the provided errs to be returned from UpdateLedgerStatus.
// An empty string indicates no error. This method both updates the receiver and returns it.
func (k *MockKeeper) WithUpdateLedgerStatusErrs(errs ...string) *MockKeeper {
	k.UpdateLedgerStatusErrs = append(k.UpdateLedgerStatusErrs, errs...)
	return k
}

// WithExpValidateAuthority adds the provided authorities to the list of expected calls to ValidateAuthority.
// This method both updates the receiver and returns it.
func (k *MockKeeper) WithExpValidateAuthority(authorities ...string) *MockKeeper {
	k.ValidateAuthorityExp = append(k.ValidateAuthorityExp, authorities...)
	return k
}

// WithExpAddLedger adds the provided ledgers to the list of expected calls to AddLedger.
// This method both updates the receiver and returns it.
func (k *MockKeeper) WithExpAddLedger(ledgers ...*types.Ledger) *MockKeeper {
	k.AddLedgerExp = append(k.AddLedgerExp, ledgers...)
	return k
}

// WithExpGetLedger adds the provided ledger keys to the list of expected calls to GetLedger.
// This method both updates the receiver and returns it.
func (k *MockKeeper) WithExpGetLedger(ledgerKeys ...*types.Ledger) *MockKeeper {
	k.GetLedgerExp = append(k.GetLedgerExp, ledgerKeys...)
	return k
}

// WithExpUpdateLedgerStatus adds the provided ledger keys to the list of expected calls to UpdateLedgerStatus.
// This method both updates the receiver and returns it.
func (k *MockKeeper) WithExpUpdateLedgerStatus(ledgerKeys ...*types.LedgerKey) *MockKeeper {
	k.UpdateLedgerStatusExp = append(k.UpdateLedgerStatusExp, ledgerKeys...)
	return k
}

func shiftErr(errs []string) ([]string, error) {
	var err error
	if len(errs) > 0 {
		errMsg := errs[0]
		errs = errs[1:]
		switch {
		case errMsg == "ErrLedgerDoesNotExist":
			err = types.ErrNotFound
		case errMsg == "ErrLedgerClassDoesNotExist":
			err = types.ErrNotFound
		case len(errMsg) > 0:
			err = errors.New(errMsg)
		}
	}
	return errs, err
}

func (k *MockKeeper) ValidateAuthority(authority string) error {
	k.ValidateAuthorityArgs = append(k.ValidateAuthorityArgs, authority)
	var err error
	k.ValidateAuthorityErrs, err = shiftErr(k.ValidateAuthorityErrs)
	return err
}

func (k *MockKeeper) AddLedger(_ sdk.Context, ledger types.Ledger) error {
	k.AddLedgerArgs = append(k.AddLedgerArgs, &ledger)
	var err error
	k.AddLedgerErrs, err = shiftErr(k.AddLedgerErrs)
	return err
}

func (k *MockKeeper) GetLedger(_ sdk.Context, key *types.LedgerKey) (*types.Ledger, error) {
	k.GetLedgerArgs = append(k.GetLedgerArgs, key)
	var err error
	k.GetLedgerErrs, err = shiftErr(k.GetLedgerErrs)
	if err != nil {
		return nil, err
	}
	// Return the expected ledger if available
	if len(k.GetLedgerExp) > 0 {
		ledger := k.GetLedgerExp[0]
		k.GetLedgerExp = k.GetLedgerExp[1:]
		return ledger, nil
	}
	return nil, nil
}

func (k *MockKeeper) UpdateLedgerStatus(_ sdk.Context, key *types.LedgerKey, statusTypeId int32) error {
	k.UpdateLedgerStatusArgs = append(k.UpdateLedgerStatusArgs, key)
	var err error
	k.UpdateLedgerStatusErrs, err = shiftErr(k.UpdateLedgerStatusErrs)
	return err
}

func (k *MockKeeper) AssertCalls(t testing.TB) bool {
	ok := assert.Equal(t, k.ValidateAuthorityExp, k.ValidateAuthorityArgs, "Calls to ValidateAuthority")
	if assert.Equal(t, len(k.AddLedgerExp), len(k.AddLedgerArgs), "Number of calls to AddLedger") {
		for i, exp := range k.AddLedgerExp {
			if i < len(k.AddLedgerArgs) {
				ok = assert.Equal(t, exp, k.AddLedgerArgs[i], "Call %d to AddLedger", i) && ok
			}
		}
	}
	if assert.Equal(t, len(k.GetLedgerExp), len(k.GetLedgerArgs), "Number of calls to GetLedger") {
		for i, exp := range k.GetLedgerExp {
			if i < len(k.GetLedgerArgs) {
				ok = assert.Equal(t, exp, k.GetLedgerArgs[i], "Call %d to GetLedger", i) && ok
			}
		}
	}
	if assert.Equal(t, len(k.UpdateLedgerStatusExp), len(k.UpdateLedgerStatusArgs), "Number of calls to UpdateLedgerStatus") {
		for i, exp := range k.UpdateLedgerStatusExp {
			if i < len(k.UpdateLedgerStatusArgs) {
				ok = assert.Equal(t, exp, k.UpdateLedgerStatusArgs[i], "Call %d to UpdateLedgerStatus", i) && ok
			}
		}
	}
	return ok
}

func (s *MsgServerTestSuite) TestCreate() {
	tests := []struct {
		name   string
		kpr    *MockKeeper
		req    *types.MsgCreateRequest
		expErr string
	}{
		{
			name: "invalid authority",
			kpr:  NewMockKeeper().WithValidateAuthorityErrs("invalid authority address"),
			req: &types.MsgCreateRequest{
				Ledger: &types.Ledger{
					Key: &types.LedgerKey{
						NftId:        "test-nft-id",
						AssetClassId: "test-asset-class-id",
					},
					LedgerClassId: "test-ledger-class-id",
					StatusTypeId:  1,
				},
				Authority: "invalid-authority",
			},
			expErr: "rpc error: code = InvalidArgument desc = invalid authority address",
		},
		{
			name: "successful create",
			kpr: NewMockKeeper().
				WithExpAddLedger(&types.Ledger{
					Key: &types.LedgerKey{
						NftId:        "test-nft-id",
						AssetClassId: "test-asset-class-id",
					},
					LedgerClassId: "test-ledger-class-id",
					StatusTypeId:  1,
				}),
			req: &types.MsgCreateRequest{
				Ledger: &types.Ledger{
					Key: &types.LedgerKey{
						NftId:        "test-nft-id",
						AssetClassId: "test-asset-class-id",
					},
					LedgerClassId: "test-ledger-class-id",
					StatusTypeId:  1,
				},
				Authority: authority,
			},
		},
		{
			name: "add ledger error",
			kpr: NewMockKeeper().
				WithAddLedgerErrs("ledger class does not exist"),
			req: &types.MsgCreateRequest{
				Ledger: &types.Ledger{
					Key: &types.LedgerKey{
						NftId:        "test-nft-id",
						AssetClassId: "test-asset-class-id",
					},
					LedgerClassId: "non-existent-ledger-class",
					StatusTypeId:  1,
				},
				Authority: authority,
			},
			expErr: "rpc error: code = InvalidArgument desc = ledger class does not exist",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.kpr == nil {
				tc.kpr = NewMockKeeper()
			}
			tc.kpr = tc.kpr.WithExpValidateAuthority(tc.req.Authority)

			var expResp, actResp *types.MsgCreateResponse
			if len(tc.expErr) == 0 {
				expResp = &types.MsgCreateResponse{}
			}

			msgServer := keeper.NewMsgServer(s.app.LedgerKeeper)
			var err error
			testFunc := func() {
				actResp, err = msgServer.Create(s.ctx, tc.req)
			}
			s.Require().NotPanics(testFunc, "Create(...)")
			assertions.AssertErrorValue(s.T(), err, tc.expErr, "Create(...) error")
			s.Assert().Equal(expResp, actResp, "Create(...) response")
			tc.kpr.AssertCalls(s.T())
		})
	}
}

func (s *MsgServerTestSuite) TestUpdateStatus() {
	tests := []struct {
		name   string
		kpr    *MockKeeper
		req    *types.MsgUpdateStatusRequest
		expErr string
	}{
		{
			name: "invalid authority",
			kpr:  NewMockKeeper().WithValidateAuthorityErrs("invalid authority address"),
			req: &types.MsgUpdateStatusRequest{
				Key: &types.LedgerKey{
					NftId:        "test-nft-id",
					AssetClassId: "test-asset-class-id",
				},
				Authority:    "invalid-authority",
				StatusTypeId: 2,
			},
			expErr: "rpc error: code = InvalidArgument desc = invalid authority address",
		},
		{
			name: "successful update",
			kpr: NewMockKeeper().
				WithExpUpdateLedgerStatus(&types.LedgerKey{
					NftId:        "test-nft-id",
					AssetClassId: "test-asset-class-id",
				}),
			req: &types.MsgUpdateStatusRequest{
				Key: &types.LedgerKey{
					NftId:        "test-nft-id",
					AssetClassId: "test-asset-class-id",
				},
				Authority:    authority,
				StatusTypeId: 2,
			},
		},
		{
			name: "update status error",
			kpr: NewMockKeeper().
				WithUpdateLedgerStatusErrs("ledger does not exist"),
			req: &types.MsgUpdateStatusRequest{
				Key: &types.LedgerKey{
					NftId:        "non-existent-nft-id",
					AssetClassId: "test-asset-class-id",
				},
				Authority:    authority,
				StatusTypeId: 2,
			},
			expErr: "rpc error: code = InvalidArgument desc = ledger does not exist",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.kpr == nil {
				tc.kpr = NewMockKeeper()
			}
			tc.kpr = tc.kpr.WithExpValidateAuthority(tc.req.Authority)

			var expResp, actResp *types.MsgUpdateStatusResponse
			if len(tc.expErr) == 0 {
				expResp = &types.MsgUpdateStatusResponse{}
			}

			msgServer := keeper.NewMsgServer(s.app.LedgerKeeper)
			var err error
			testFunc := func() {
				actResp, err = msgServer.UpdateStatus(s.ctx, tc.req)
			}
			s.Require().NotPanics(testFunc, "UpdateStatus(...)")
			assertions.AssertErrorValue(s.T(), err, tc.expErr, "UpdateStatus(...) error")
			s.Assert().Equal(expResp, actResp, "UpdateStatus(...) response")
			tc.kpr.AssertCalls(s.T())
		})
	}
}

func (s *MsgServerTestSuite) TestUpdateInterestRate() {
	// TODO: Implement test for MsgUpdateInterestRateRequest
	s.T().Skip("Test not implemented yet")
}

func (s *MsgServerTestSuite) TestUpdatePayment() {
	// TODO: Implement test for MsgUpdatePaymentRequest
	s.T().Skip("Test not implemented yet")
}

func (s *MsgServerTestSuite) TestUpdateMaturityDate() {
	// TODO: Implement test for MsgUpdateMaturityDateRequest
	s.T().Skip("Test not implemented yet")
}

func (s *MsgServerTestSuite) TestAppend() {
	// TODO: Implement test for MsgAppendRequest
	s.T().Skip("Test not implemented yet")
}

func (s *MsgServerTestSuite) TestUpdateBalances() {
	// TODO: Implement test for MsgUpdateBalancesRequest
	s.T().Skip("Test not implemented yet")
}

func (s *MsgServerTestSuite) TestTransferFundsWithSettlement() {
	// TODO: Implement test for MsgTransferFundsWithSettlementRequest
	s.T().Skip("Test not implemented yet")
}

func (s *MsgServerTestSuite) TestDestroy() {
	// TODO: Implement test for MsgDestroyRequest
	s.T().Skip("Test not implemented yet")
}

func (s *MsgServerTestSuite) TestCreateLedgerClass() {
	// TODO: Implement test for MsgCreateLedgerClassRequest
	s.T().Skip("Test not implemented yet")
}

func (s *MsgServerTestSuite) TestAddLedgerClassStatusType() {
	// TODO: Implement test for MsgAddLedgerClassStatusTypeRequest
	s.T().Skip("Test not implemented yet")
}

func (s *MsgServerTestSuite) TestAddLedgerClassEntryType() {
	// TODO: Implement test for MsgAddLedgerClassEntryTypeRequest
	s.T().Skip("Test not implemented yet")
}

func (s *MsgServerTestSuite) TestAddLedgerClassBucketType() {
	// TODO: Implement test for MsgAddLedgerClassBucketTypeRequest
	s.T().Skip("Test not implemented yet")
}

func (s *MsgServerTestSuite) TestBulkImport() {
	// TODO: Implement test for MsgBulkImportRequest
	s.T().Skip("Test not implemented yet")
}
