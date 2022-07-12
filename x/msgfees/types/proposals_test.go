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

	m := NewAddMsgFeeProposal("title", "description", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)))
	s.Assert().Equal(
		`Add Msg Fee Proposal:
Title:         title
Description:   description
Msg:           /provenance.metadata.v1.MsgWriteRecordRequest
AdditionalFee: 10hotdog
`, m.String())

	tests := []struct {
		name        string
		proposal    *AddMsgFeeProposal
		expectedErr string
	}{
		{
			"Empty type error",
			NewAddMsgFeeProposal("title", "description", "", sdk.NewCoin("hotdog", sdk.NewInt(10))),
			ErrEmptyMsgType.Error(),
		},
		{
			"Invalid fee amounts",
			NewAddMsgFeeProposal("title", "description", msgType, sdk.NewCoin("hotdog", sdk.NewInt(0))),
			ErrInvalidFee.Error(),
		},
		{
			"Invalid proposal details",
			NewAddMsgFeeProposal("title", "", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10))),
			"proposal description cannot be blank: invalid proposal content",
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			err := tc.proposal.ValidateBasic()
			s.Require().NotNil(err)
			s.Assert().Equal(tc.expectedErr, err.Error())
		})
	}
}

func (s *MsgFeesProposalTestSuite) TestUpdateMsgFeeProposalType() {
	msgType := sdk.MsgTypeURL(&metadatatypes.MsgWriteRecordRequest{})

	m := NewUpdateMsgFeeProposal("title", "description", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)))
	s.Assert().Equal(
		`Update Msg Fee Proposal:
Title:         title
Description:   description
Msg:           /provenance.metadata.v1.MsgWriteRecordRequest
AdditionalFee: 10hotdog
`, m.String())

	tests := []struct {
		name        string
		proposal    *UpdateMsgFeeProposal
		expectedErr string
	}{
		{
			"Empty type error",
			NewUpdateMsgFeeProposal("title", "description", "", sdk.NewCoin("hotdog", sdk.NewInt(10))),
			ErrEmptyMsgType.Error(),
		},
		{
			"Invalid fee amounts",
			NewUpdateMsgFeeProposal("title", "description", msgType, sdk.NewCoin("hotdog", sdk.NewInt(0))),
			ErrInvalidFee.Error(),
		},
		{
			"Invalid proposal details",
			NewUpdateMsgFeeProposal("title", "", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10))),
			"proposal description cannot be blank: invalid proposal content",
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			err := tc.proposal.ValidateBasic()
			s.Assert().Equal(tc.expectedErr, err.Error())
		})
	}

}

func (s *MsgFeesProposalTestSuite) TestRemoveMsgFeeProposalType() {
	msgType := sdk.MsgTypeURL(&metadatatypes.MsgWriteRecordRequest{})

	m := NewRemoveMsgFeeProposal("title", "description", msgType)
	s.Assert().Equal(
		`Remove Msg Fee Proposal:
  Title:       title
  Description: description
  MsgTypeUrl:  /provenance.metadata.v1.MsgWriteRecordRequest
`, m.String())

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
	m := NewUpdateNhashPerUsdMilProposal("title", "description", 70)
	s.Assert().Equal(
		`Update Nhash to Usd Mil Proposal:
  Title:             title
  Description:       description
  NhashPerUsdMil:    70
`, m.String())

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
