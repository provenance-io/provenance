package types

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"

	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ExpirationTestSuite struct {
	suite.Suite

	moduleAssetID string
	owner         string
	time          time.Time
	deposit       sdk.Coin
	message       types.Any

	signers      []string
	otherSigners []string

	scopeID metadatatypes.MetadataAddress

	validExpiration              Expiration
	emptyModuleAssetIdExpiration Expiration
	emptyOwnerExpiration         Expiration
	expireTimeInPastExpiration   Expiration
	invalidDepositExpiration     Expiration
	negativeDepositExpiration    Expiration
	invalidMessageExpiration     Expiration
}

func (s *ExpirationTestSuite) SetupTest() {
	expirationTime := time.Now().AddDate(0, 0, 2)
	s.moduleAssetID = "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"
	s.owner = "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"
	s.time = expirationTime
	s.deposit = sdk.NewInt64Coin("testcoin", 1905)

	s.signers = []string{s.owner}
	s.otherSigners = []string{"cosmos1tnh2q55v8wyygtt9srz5safamzdengsnqeycj3"}

	s.scopeID = metadatatypes.ScopeMetadataAddress(uuid.New())

	past := expirationTime.AddDate(-1, 0, 0)

	s.validExpiration = Expiration{
		ModuleAssetId: s.moduleAssetID,
		Owner:         s.owner,
		Time:          s.time,
		Deposit:       s.deposit,
		Message:       s.anyMsg(s.owner),
	}
	s.emptyModuleAssetIdExpiration = Expiration{
		Owner:   s.owner,
		Time:    s.time,
		Deposit: s.deposit,
		Message: s.anyMsg(s.owner),
	}
	s.emptyOwnerExpiration = Expiration{
		ModuleAssetId: s.moduleAssetID,
		Time:          s.time,
		Deposit:       s.deposit,
		Message:       s.anyMsg(s.owner),
	}
	s.expireTimeInPastExpiration = Expiration{
		ModuleAssetId: s.moduleAssetID,
		Owner:         s.owner,
		Time:          past,
		Deposit:       s.deposit,
		Message:       s.anyMsg(s.owner),
	}
	s.invalidDepositExpiration = Expiration{
		ModuleAssetId: s.moduleAssetID,
		Owner:         s.owner,
		Time:          s.time,
		Message:       s.anyMsg(s.owner),
	}
	s.negativeDepositExpiration = Expiration{
		ModuleAssetId: s.moduleAssetID,
		Owner:         s.owner,
		Time:          s.time,
		Deposit:       sdk.Coin{Denom: "testcoin", Amount: sdk.NewInt(-1)},
		Message:       s.anyMsg(s.owner),
	}
	s.invalidMessageExpiration = Expiration{
		ModuleAssetId: s.moduleAssetID,
		Owner:         s.owner,
		Time:          s.time,
		Deposit:       s.deposit,
		Message:       types.Any{}, // will fail validation
	}
}

func TestExpirationTestSuite(t *testing.T) {
	suite.Run(t, new(ExpirationTestSuite))
}

func (s *ExpirationTestSuite) anyMsg(owner string) types.Any {
	msg := &metadatatypes.MsgDeleteScopeRequest{
		ScopeId: s.scopeID,
		Signers: []string{owner},
	}
	anyMsg, err := types.NewAnyWithValue(msg)
	if err != nil {
		panic(err)
	}
	return *anyMsg
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
			name:        "should fail to validate basic - expiration time in past",
			msg:         NewMsgAddExpirationRequest(s.expireTimeInPastExpiration, s.signers),
			wantErr:     true,
			expectedErr: ErrTimeInPast,
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
			name:        "should fail to validate basic - invalid message",
			msg:         NewMsgAddExpirationRequest(s.invalidMessageExpiration, []string{}),
			wantErr:     true,
			expectedErr: ErrInvalidMessage,
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
				e, ok := err.(*errors.Error)
				require.True(t, ok, "%s failed error type check", tc.name)
				assert.Error(t, err, "%s expected error", tc.name)
				assert.Equal(t, tc.expectedErr.ABCICode(), e.ABCICode(), "%s error", tc.name)
			} else {
				assert.NoError(t, err, "%s unexpected error", tc.name)
			}
		})
	}
}

func (s *ExpirationTestSuite) TestMsgExtendExpirationRequestValidateBasic() {
	duration := "11h"
	cases := []struct {
		name        string
		msg         *MsgExtendExpirationRequest
		wantErr     bool
		expectedErr *errors.Error
	}{
		{
			name:        "should succeed to validate basic",
			msg:         NewMsgExtendExpirationRequest(s.moduleAssetID, duration, s.signers),
			wantErr:     false,
			expectedErr: nil,
		}, {
			name:        "should fail to validate basic - missing module asset id",
			msg:         NewMsgExtendExpirationRequest("", duration, s.signers),
			wantErr:     true,
			expectedErr: ErrEmptyModuleAssetID,
		}, {
			name:        "should fail to validate basic - invalid duration format",
			msg:         NewMsgExtendExpirationRequest(s.moduleAssetID, "1s", s.signers),
			wantErr:     true,
			expectedErr: ErrDurationValue,
		},
		{
			name:        "should fail to validate basic - negative duration period",
			msg:         NewMsgExtendExpirationRequest(s.moduleAssetID, "-1y", s.signers),
			wantErr:     true,
			expectedErr: ErrDurationValue,
		},
		{
			name:        "should fail to validate basic - missing signers",
			msg:         NewMsgExtendExpirationRequest(s.moduleAssetID, duration, []string{}),
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

func (s *ExpirationTestSuite) TestMsgInvokeExpirationRequestValidateBasic() {
	cases := []struct {
		name        string
		msg         *MsgInvokeExpirationRequest
		wantErr     bool
		expectedErr *errors.Error
	}{
		{
			name:        "should succeed to validate basic",
			msg:         NewMsgInvokeExpirationRequest(s.moduleAssetID, s.signers),
			wantErr:     false,
			expectedErr: nil,
		}, {
			name:        "should fail to validate basic - empty module asset id",
			msg:         NewMsgInvokeExpirationRequest("", s.signers),
			wantErr:     true,
			expectedErr: ErrEmptyModuleAssetID,
		}, {
			name:        "should fail to validate basic - missing signers",
			msg:         NewMsgInvokeExpirationRequest(s.moduleAssetID, []string{}),
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
