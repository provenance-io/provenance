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

func (s *MsgFeesProposalTestSuite) TestAddMsgBasedFeeProposalType() {

	msgType := sdk.MsgTypeURL(&metadatatypes.MsgWriteRecordRequest{})

	m := NewAddMsgBasedFeeProposal("title", "description", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)))
	s.Assert().Equal(
		`Add Msg Based Fee Proposal:
Title:         title
Description:   description
Msg:           /provenance.metadata.v1.MsgWriteRecordRequest
AdditionalFee: 10hotdog
`, m.String())

	tests := []struct {
		name        string
		proposal    *AddMsgBasedFeeProposal
		expectedErr string
	}{
		{
			"Empty type error",
			NewAddMsgBasedFeeProposal("title", "description", "", sdk.NewCoin("hotdog", sdk.NewInt(10))),
			ErrEmptyMsgType.Error(),
		},
		{
			"Invalid fee amounts",
			NewAddMsgBasedFeeProposal("title", "description", msgType, sdk.NewCoin("hotdog", sdk.NewInt(0))),
			ErrInvalidFee.Error(),
		},
		{
			"Invalid proposal details",
			NewAddMsgBasedFeeProposal("title", "", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10))),
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

func (s *MsgFeesProposalTestSuite) TestUpdateMsgBasedFeeProposalType() {

	msgType := sdk.MsgTypeURL(&metadatatypes.MsgWriteRecordRequest{})

	m := NewUpdateMsgBasedFeeProposal("title", "description", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10)))
	s.Assert().Equal(
		`Update Msg Based Fee Proposal:
Title:         title
Description:   description
Msg:           /provenance.metadata.v1.MsgWriteRecordRequest
AdditionalFee: 10hotdog
`, m.String())

	tests := []struct {
		name        string
		proposal    *UpdateMsgBasedFeeProposal
		expectedErr string
	}{
		{
			"Empty type error",
			NewUpdateMsgBasedFeeProposal("title", "description", "", sdk.NewCoin("hotdog", sdk.NewInt(10))),
			ErrEmptyMsgType.Error(),
		},
		{
			"Invalid fee amounts",
			NewUpdateMsgBasedFeeProposal("title", "description", msgType, sdk.NewCoin("hotdog", sdk.NewInt(0))),
			ErrInvalidFee.Error(),
		},
		{
			"Invalid proposal details",
			NewUpdateMsgBasedFeeProposal("title", "", msgType, sdk.NewCoin("hotdog", sdk.NewInt(10))),
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

func (s *MsgFeesProposalTestSuite) TestRemoveMsgBasedFeeProposalType() {
	msgType := sdk.MsgTypeURL(&metadatatypes.MsgWriteRecordRequest{})

	m := NewRemoveMsgBasedFeeProposal("title", "description", msgType)
	s.Assert().Equal(
		`Remove Msg Based Fee Proposal:
  Title:       title
  Description: description
  MsgTypeUrl:         /provenance.metadata.v1.MsgWriteRecordRequest
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
