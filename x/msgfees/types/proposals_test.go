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
			"Empty type error",
			NewAddMsgFeeProposal("title", "description", "", sdk.NewCoin("hotdog", sdk.NewInt(10)), "", ""),
			ErrEmptyMsgType.Error(),
		},
		{
			"Invalid fee amounts",
			NewAddMsgFeeProposal("title", "description", msgType, sdk.NewCoin("hotdog", sdk.NewInt(0)), "", ""),
			ErrInvalidFee.Error(),
		},
		{
			"Invalid proposal details",
			NewAddMsgFeeProposal("title", "", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "", ""),
			"proposal description cannot be blank: invalid proposal content",
		},
		{
			"Invalid proposal recipient address",
			NewAddMsgFeeProposal("title", "", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "invalid", ""),
			"decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			"Invalid proposal invalid basis points for address",
			NewAddMsgFeeProposal("title", "description", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", "10_001"),
			"recipient basis points can only be between 0 and 10,000 : 10001",
		},
		{
			"Valid proposal without recipient",
			NewAddMsgFeeProposal("title", "description", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "", ""),
			"",
		},
		{
			"Valid proposal with recipient",
			NewAddMsgFeeProposal("title", "description", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", "10_000"),
			"",
		},
		{
			"Valid proposal with recipient without defined bips",
			NewAddMsgFeeProposal("title", "description", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", ""),
			"",
		},
		{
			"Valid proposal with recipient with defined bips",
			NewAddMsgFeeProposal("title", "description", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", "10"),
			"",
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
			"Empty type error",
			NewUpdateMsgFeeProposal("title", "description", "", sdk.NewCoin("hotdog", sdk.NewInt(10)), "", ""),
			ErrEmptyMsgType.Error(),
		},
		{
			"Invalid fee amounts",
			NewUpdateMsgFeeProposal("title", "description", msgType, sdk.NewCoin("hotdog", sdk.NewInt(0)), "", ""),
			ErrInvalidFee.Error(),
		},
		{
			"Invalid proposal details",
			NewUpdateMsgFeeProposal("title", "", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "", ""),
			"proposal description cannot be blank: invalid proposal content",
		},
		{
			"Invalid proposal recipient address",
			NewUpdateMsgFeeProposal("title", "", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "invalid", "50"),
			"decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			"Invalid proposal invalid basis points for address",
			NewUpdateMsgFeeProposal("title", "description", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", "10_001"),
			"recipient basis points can only be between 0 and 10,000 : 10001",
		},
		{
			"Valid proposal without recipient",
			NewUpdateMsgFeeProposal("title", "description", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "", ""),
			"",
		},
		{
			"Valid proposal with recipient without defined bips",
			NewUpdateMsgFeeProposal("title", "description", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", ""),
			"",
		},
		{
			"Valid proposal with recipient with defined bips",
			NewUpdateMsgFeeProposal("title", "description", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)), "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", "10"),
			"",
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
			"Empty type error",
			NewUpdateNhashPerUsdMilProposal("title", "description", 0),
			"nhash per usd mil must be greater than 0",
		},
		{
			"Invalid proposal details",
			NewUpdateNhashPerUsdMilProposal("title", "", 70),
			"proposal description cannot be blank: invalid proposal content",
		},
		{
			"Valid proposal",
			NewUpdateNhashPerUsdMilProposal("title", "description", 70),
			"",
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
