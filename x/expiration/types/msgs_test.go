package types

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ExpirationTestSuite struct {
	suite.Suite

	moduleAssetID string
	owner         string
	blockHeight   int64
	deposit       sdk.Coin
	message       types.Any

	signers      []string
	otherSigners []string

	validExpiration               Expiration
	emptyModuleAssetIdExpiration  Expiration
	emptyOwnerExpiration          Expiration
	negativeBlockHeightExpiration Expiration
	invalidDepositExpiration      Expiration
	negativeDepositExpiration     Expiration
}

func (s *ExpirationTestSuite) SetupTest() {
	s.moduleAssetID = "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"
	s.owner = "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"
	s.blockHeight = 1
	s.deposit = sdk.NewInt64Coin("testcoin", 1905)

	s.signers = []string{s.owner}
	s.otherSigners = []string{"cosmos1tnh2q55v8wyygtt9srz5safamzdengsnqeycj3"}

	s.validExpiration = Expiration{
		ModuleAssetId: s.moduleAssetID,
		Owner:         s.owner,
		BlockHeight:   s.blockHeight,
		Deposit:       s.deposit,
	}
	s.emptyModuleAssetIdExpiration = Expiration{
		Owner:       s.owner,
		BlockHeight: s.blockHeight,
		Deposit:     s.deposit,
	}
	s.emptyOwnerExpiration = Expiration{
		ModuleAssetId: s.moduleAssetID,
		BlockHeight:   s.blockHeight,
		Deposit:       s.deposit,
	}
	s.negativeBlockHeightExpiration = Expiration{
		ModuleAssetId: s.moduleAssetID,
		Owner:         s.owner,
		BlockHeight:   -1,
		Deposit:       s.deposit,
	}
	s.invalidDepositExpiration = Expiration{
		ModuleAssetId: s.moduleAssetID,
		Owner:         s.owner,
		BlockHeight:   s.blockHeight,
	}
	s.negativeDepositExpiration = Expiration{
		ModuleAssetId: s.moduleAssetID,
		Owner:         s.owner,
		BlockHeight:   s.blockHeight,
		Deposit:       sdk.Coin{Denom: "testcoin", Amount: sdk.NewInt(-1)},
	}
}

func TestExpirationTestSuite(t *testing.T) {
	suite.Run(t, new(ExpirationTestSuite))
}

func (s *ExpirationTestSuite) TestMsgAddExpirationRequestValidateBasic() {
	cases := []struct {
		name        string
		msg         *MsgAddExpirationRequest
		wantErr     bool
		expectedErr *errors.Error
	}{
		{
			name:        "should succeed to validate basic",
			msg:         NewMsgAddExpirationRequest(s.validExpiration, s.signers),
			wantErr:     false,
			expectedErr: nil,
		}, {
			name:        "should fail to validate basic - missing module asset id",
			msg:         NewMsgAddExpirationRequest(s.emptyModuleAssetIdExpiration, s.signers),
			wantErr:     true,
			expectedErr: ErrEmptyModuleAssetID,
		}, {
			name:        "should fail to validate basic - missing owner address",
			msg:         NewMsgAddExpirationRequest(s.emptyOwnerExpiration, s.signers),
			wantErr:     true,
			expectedErr: ErrEmptyOwnerAddress,
		}, {
			name:        "should fail to validate basic - negative block height",
			msg:         NewMsgAddExpirationRequest(s.negativeBlockHeightExpiration, s.signers),
			wantErr:     true,
			expectedErr: ErrBlockHeightLteZero,
		}, {
			name:        "should fail to validate basic - invalid deposit",
			msg:         NewMsgAddExpirationRequest(s.invalidDepositExpiration, s.signers),
			wantErr:     true,
			expectedErr: ErrInvalidDeposit,
		}, {
			name:        "should fail to validate basic - negative deposit",
			msg:         NewMsgAddExpirationRequest(s.negativeDepositExpiration, s.signers),
			wantErr:     true,
			expectedErr: ErrInvalidDeposit,
		}, {
			name:        "should fail to validate basic - missing signers",
			msg:         NewMsgAddExpirationRequest(s.validExpiration, []string{}),
			wantErr:     true,
			expectedErr: ErrMissingSigners,
		},
	}

	for _, tc := range cases {
		tc := tc

		s.T().Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.wantErr {
				assert.Error(t, err, "%s expected error", tc.name)
				assert.Equal(t, tc.expectedErr, err, "%s error", tc.name)
			} else {
				assert.NoError(t, err, "%s unexpected error", tc.name)
			}
		})
	}
}

func (s *ExpirationTestSuite) TestMsgExtendExpirationRequestValidateBasic() {
	cases := []struct {
		name        string
		msg         *MsgExtendExpirationRequest
		wantErr     bool
		expectedErr *errors.Error
	}{
		{
			name:        "should succeed to validate basic",
			msg:         NewMsgExtendExpirationRequest(s.validExpiration, s.signers),
			wantErr:     false,
			expectedErr: nil,
		}, {
			name:        "should fail to validate basic - missing module asset id",
			msg:         NewMsgExtendExpirationRequest(s.emptyModuleAssetIdExpiration, s.signers),
			wantErr:     true,
			expectedErr: ErrEmptyModuleAssetID,
		}, {
			name:        "should fail to validate basic - missing owner address",
			msg:         NewMsgExtendExpirationRequest(s.emptyOwnerExpiration, s.signers),
			wantErr:     true,
			expectedErr: ErrEmptyOwnerAddress,
		}, {
			name:        "should fail to validate basic - negative block height",
			msg:         NewMsgExtendExpirationRequest(s.negativeBlockHeightExpiration, s.signers),
			wantErr:     true,
			expectedErr: ErrBlockHeightLteZero,
		}, {
			name:        "should fail to validate basic - invalid deposit",
			msg:         NewMsgExtendExpirationRequest(s.invalidDepositExpiration, s.signers),
			wantErr:     true,
			expectedErr: ErrInvalidDeposit,
		}, {
			name:        "should fail to validate basic - negative deposit",
			msg:         NewMsgExtendExpirationRequest(s.negativeDepositExpiration, s.signers),
			wantErr:     true,
			expectedErr: ErrInvalidDeposit,
		}, {
			name:        "should fail to validate basic - missing signers",
			msg:         NewMsgExtendExpirationRequest(s.validExpiration, []string{}),
			wantErr:     true,
			expectedErr: ErrMissingSigners,
		},
	}

	for _, tc := range cases {
		tc := tc

		s.T().Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.wantErr {
				assert.Error(t, err, "%s expected error", tc.name)
				assert.Equal(t, tc.expectedErr, err, "%s error", tc.name)
			} else {
				assert.NoError(t, err, "%s unexpected error", tc.name)
			}
		})
	}
}

func (s *ExpirationTestSuite) TestMsgDeleteExpirationRequestValidateBasic() {
	cases := []struct {
		name        string
		msg         *MsgDeleteExpirationRequest
		wantErr     bool
		expectedErr *errors.Error
	}{
		{
			name:        "should succeed to validate basic",
			msg:         NewMsgDeleteExpirationRequest(s.moduleAssetID, s.signers),
			wantErr:     false,
			expectedErr: nil,
		}, {
			name:        "should fail to validate basic - empty module asset id",
			msg:         NewMsgDeleteExpirationRequest("", s.signers),
			wantErr:     true,
			expectedErr: ErrEmptyModuleAssetID,
		}, {
			name:        "should fail to validate basic - missing signers",
			msg:         NewMsgDeleteExpirationRequest(s.moduleAssetID, []string{}),
			wantErr:     true,
			expectedErr: ErrMissingSigners,
		},
	}

	for _, tc := range cases {
		tc := tc

		s.T().Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.wantErr {
				assert.Error(t, err, "%s expected error", tc.name)
				assert.Equal(t, tc.expectedErr, err, "%s error", tc.name)
			} else {
				assert.NoError(t, err, "%s unexpected error", tc.name)
			}
		})
	}
}
