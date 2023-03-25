package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"

	"github.com/stretchr/testify/suite"
)

type MsgFeesProposalTestSuite struct {
	suite.Suite
}

func (s *MsgFeesProposalTestSuite) SetupSuite() {
}

func TestMsgFeesProposalTestSuite(t *testing.T) {
	suite.Run(t, new(MsgFeesProposalTestSuite))
}

func (s *MsgFeesProposalTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
}

func (s *MsgFeesProposalTestSuite) TestAddMsgFeeProposalType() {
	msgType := sdk.MsgTypeURL(&metadatatypes.MsgWriteRecordRequest{})

	tests := []struct {
		name        string
		proposal    *AddMsgFeeProposal
		expectedErr string
	}{
		{
			name: "Empty type error",
			proposal: NewAddMsgFeeProposal("title", "description", "", sdk.NewCoin("hotdog", sdk.NewInt(10)), "", ""),
			expectedErr: ErrEmptyMsgType.Error(),
		},
		{
			name: "Invalid fee amounts",
			proposal: NewAddMsgFeeProposal("title", "description", msgType, sdk.NewCoin("hotdog", sdk.NewInt(0)), "", ""),
			expectedErr: ErrInvalidFee.Error(),
		},
		{
			name: "Invalid proposal details",
			proposal: NewAddMsgFeeProposal("title", "", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "", ""),
			expectedErr: "proposal description cannot be blank: invalid proposal content",
		},
		{
			name: "Invalid proposal recipient address",
			proposal: NewAddMsgFeeProposal("title", "", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "invalid", ""),
			expectedErr: "decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			name: "Invalid proposal invalid basis points for address",
			proposal: NewAddMsgFeeProposal("title", "description", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", "10001"),
			expectedErr: "recipient basis points can only be between 0 and 10,000 : 10001",
		},
		{
			name: "Valid proposal without recipient",
			proposal: NewAddMsgFeeProposal("title", "description", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "", ""),
			expectedErr: "",
		},
		{
		name: 	"Valid proposal with recipient",
			proposal: NewAddMsgFeeProposal("title", "description", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", "10000"),
			expectedErr: "",
		},
		{
			name: "Valid proposal with recipient without defined bips",
			proposal: NewAddMsgFeeProposal("title", "description", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", ""),
			expectedErr: "",
		},
		{
			name: "Valid proposal with recipient with defined bips",
			proposal: NewAddMsgFeeProposal("title", "description", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", "10"),
			expectedErr: "",
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			err := tc.proposal.ValidateBasic()
			if len(tc.expectedErr) != 0 {
				s.Require().NotNil(err, "Error should not be nil for test %s", tc.name)
				s.Assert().Equal(tc.expectedErr, err.Error(), "Error messages do not match for test %s", tc.name)
			} else {
				s.Require().Nil(err, "Error should be nil for test %s", tc.name)
			}
		})
	}
}

func (s *MsgFeesProposalTestSuite) TestUpdateMsgFeeProposalType() {
	msgType := sdk.MsgTypeURL(&metadatatypes.MsgWriteRecordRequest{})
	tests := []struct {
		name        string
		proposal    *UpdateMsgFeeProposal
		expectedErr string
	}{
		{
			name: "Empty type error",
			proposal: NewUpdateMsgFeeProposal("title", "description", "", sdk.NewCoin("hotdog", sdk.NewInt(10)), "", ""),
			expectedErr: ErrEmptyMsgType.Error(),
		},
		{
			name: "Invalid fee amounts",
			proposal: NewUpdateMsgFeeProposal("title", "description", msgType, sdk.NewCoin("hotdog", sdk.NewInt(0)), "", ""),
			expectedErr: ErrInvalidFee.Error(),
		},
		{
			name: "Invalid proposal details",
			proposal: NewUpdateMsgFeeProposal("title", "", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "", ""),
			expectedErr: "proposal description cannot be blank: invalid proposal content",
		},
		{
			name: "Invalid proposal recipient address",
			proposal: NewUpdateMsgFeeProposal("title", "", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "invalid", "50"),
			expectedErr: "decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			name: "Invalid proposal invalid basis points for address",
			proposal: NewUpdateMsgFeeProposal("title", "description", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", "10001"),
			expectedErr: "recipient basis points can only be between 0 and 10,000 : 10001",
		},
		{
			name: "Valid proposal without recipient",
			proposal: NewUpdateMsgFeeProposal("title", "description", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "", ""),
			expectedErr: "",
		},
		{
			name: "Valid proposal with recipient without defined bips",
			proposal: NewUpdateMsgFeeProposal("title", "description", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", ""),
			expectedErr: "",
		},
		{
			name: "Valid proposal with recipient with defined bips",
			proposal: NewUpdateMsgFeeProposal("title", "description", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", "10"),
			expectedErr: "",
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			err := tc.proposal.ValidateBasic()
			if len(tc.expectedErr) != 0 {
				s.Require().NotNil(err, "Error should not be nil for test %s", tc.name)
				s.Assert().Equal(tc.expectedErr, err.Error(), "Error messages do not match for test %s", tc.name)
			} else {
				s.Require().Nil(err, "Error should be nil for test %s", tc.name)
			}
		})
	}

}

func (s *MsgFeesProposalTestSuite) TestRemoveMsgFeeProposalType() {
	msgType := sdk.MsgTypeURL(&metadatatypes.MsgWriteRecordRequest{})

	m := NewRemoveMsgFeeProposal("title", "description", msgType)
	err := m.ValidateBasic()
	s.Assert().NoError(err)

	m.MsgTypeUrl = ""
	err = m.ValidateBasic()
	s.Assert().ErrorIs(err, ErrEmptyMsgType)

	m.MsgTypeUrl = msgType
	m.Description = ""
	err = m.ValidateBasic()
	s.Assert().Equal("proposal description cannot be blank: invalid proposal content", err.Error())
}

func (s *MsgFeesProposalTestSuite) TestUpdateUsdConversionRateProposalValidateBasic() {
	tests := []struct {
		name        string
		proposal    *UpdateNhashPerUsdMilProposal
		expectedErr string
	}{
		{
			name: "Empty type error",
			proposal: NewUpdateNhashPerUsdMilProposal("title", "description", 0),
			expectedErr: "nhash per usd mil must be greater than 0",
		},
		{
			name: "Invalid proposal details",
			proposal: NewUpdateNhashPerUsdMilProposal("title", "", 70),
			expectedErr: "proposal description cannot be blank: invalid proposal content",
		},
		{
			name: "Valid proposal",
			proposal: NewUpdateNhashPerUsdMilProposal("title", "description", 70),
			expectedErr: "",
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			err := tc.proposal.ValidateBasic()
			if len(tc.expectedErr) == 0 {
				s.Assert().NoError(err)
			} else {
				s.Assert().Equal(tc.expectedErr, err.Error())
			}
		})
	}

}
