package types

import (
	"testing"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
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

func (s *MsgFeesProposalTestSuite) TestAddMsgBasedFeesProposalType() {

	msgType, err := cdctypes.NewAnyWithValue(&metadatatypes.MsgWriteRecordRequest{})

	s.Require().NoError(err)
	m := NewAddMsgBasedFeesProposal("title", "description", sdk.NewCoin("hotdog", sdk.NewInt(10)), msgType, sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(10))), sdk.OneDec())
	s.Assert().Equal(
		`Add Msg Based Fees Proposal:
	Title:       title
	Description: description
	Amount:      10hotdog
	Msg:         /provenance.metadata.v1.MsgWriteRecordRequest
	MinFee:      10hotdog
	FeeRate:     1.000000000000000000
`, m.String())

	tests := []struct {
		name        string
		proposal    *AddMsgBasedFeesProposal
		expectedErr string
	}{
		{
			"Empty type error",
			NewAddMsgBasedFeesProposal("title", "description", sdk.NewCoin("hotdog", sdk.NewInt(10)), nil, sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(10))), sdk.OneDec()),
			ErrEmptyMsgType.Error(),
		},
		{
			"Invalid amount",
			NewAddMsgBasedFeesProposal("title", "description", sdk.NewCoin("hotdog", sdk.ZeroInt()), msgType, sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(10))), sdk.OneDec()),
			ErrInvalidCoinAmount.Error(),
		},
		{
			"Invalid fee amounts",
			NewAddMsgBasedFeesProposal("title", "description", sdk.NewCoin("hotdog", sdk.NewInt(10)), msgType, sdk.NewCoins(), sdk.ZeroDec()),
			ErrInvalidFee.Error(),
		},
		{
			"Invalid proposal details",
			NewAddMsgBasedFeesProposal("title", "", sdk.NewCoin("hotdog", sdk.NewInt(10)), msgType, sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(10))), sdk.OneDec()),
			"proposal description cannot be blank: invalid proposal content",
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			err = tc.proposal.ValidateBasic()
			s.Assert().Equal(tc.expectedErr, err.Error())
		})
	}
}

func (s *MsgFeesProposalTestSuite) TestUpdateMsgBasedFeesProposalType() {

	msgType, err := cdctypes.NewAnyWithValue(&metadatatypes.MsgWriteRecordRequest{})

	s.Require().NoError(err)
	m := NewUpdateMsgBasedFeesProposal("title", "description", sdk.NewCoin("hotdog", sdk.NewInt(10)), msgType, sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(10))), sdk.OneDec())
	s.Assert().Equal(
		`Update Msg Based Fees Proposal:
	Title:       title
	Description: description
	Amount:      10hotdog
	Msg:         /provenance.metadata.v1.MsgWriteRecordRequest
	MinFee:      10hotdog
	FeeRate:     1.000000000000000000
`, m.String())

	tests := []struct {
		name        string
		proposal    *UpdateMsgBasedFeesProposal
		expectedErr string
	}{
		{
			"Empty type error",
			NewUpdateMsgBasedFeesProposal("title", "description", sdk.NewCoin("hotdog", sdk.NewInt(10)), nil, sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(10))), sdk.OneDec()),
			ErrEmptyMsgType.Error(),
		},
		{
			"Invalid amount",
			NewUpdateMsgBasedFeesProposal("title", "description", sdk.NewCoin("hotdog", sdk.ZeroInt()), msgType, sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(10))), sdk.OneDec()),
			ErrInvalidCoinAmount.Error(),
		},
		{
			"Invalid fee amounts",
			NewUpdateMsgBasedFeesProposal("title", "description", sdk.NewCoin("hotdog", sdk.NewInt(10)), msgType, sdk.NewCoins(), sdk.ZeroDec()),
			ErrInvalidFee.Error(),
		},
		{
			"Invalid proposal details",
			NewUpdateMsgBasedFeesProposal("title", "", sdk.NewCoin("hotdog", sdk.NewInt(10)), msgType, sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(10))), sdk.OneDec()),
			"proposal description cannot be blank: invalid proposal content",
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			err = tc.proposal.ValidateBasic()
			s.Assert().Equal(tc.expectedErr, err.Error())
		})
	}

}

func (s *MsgFeesProposalTestSuite) TestRemoveMsgBasedFeesProposalType() {

	msgType, err := cdctypes.NewAnyWithValue(&metadatatypes.MsgWriteRecordRequest{})

	s.Require().NoError(err)
	m := NewRemoveMsgBasedFeesProposal("title", "description", msgType)
	s.Assert().Equal(
		`Remove Msg Based Fees Proposal:
	Title:       title
	Description: description
	Msg:         /provenance.metadata.v1.MsgWriteRecordRequest
`, m.String())

	err = m.ValidateBasic()
	s.Assert().NoError(err)

	m.Msg = nil
	err = m.ValidateBasic()
	s.Assert().ErrorIs(err, ErrEmptyMsgType)

	m.Msg = msgType
	m.Description = ""
	err = m.ValidateBasic()
	s.Assert().Equal("", err.Error())
}
