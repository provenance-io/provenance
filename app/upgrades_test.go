package app

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	msgfeetypes "github.com/provenance-io/provenance/x/msgfees/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	app *App
	ctx sdk.Context
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupTest() {
	s.app = Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
}

func (s *IntegrationTestSuite) TestRemoveLeaveGroupMsgFee() {
	typeURL := sdk.MsgTypeURL(&group.MsgLeaveGroup{})
	s.Run("fee does not exist", func() {
		// Make sure there isn't one already.
		_ = s.app.MsgFeesKeeper.RemoveMsgFee(s.ctx, typeURL)
		err := RemoveLeaveGroupMsgFee(s.ctx, s.app)
		s.Require().NoError(err, "RemoveLeaveGroupMsgFee error")
		msgFee, err := s.app.MsgFeesKeeper.GetMsgFee(s.ctx, typeURL)
		s.Assert().NoError(err, "GetMsgFee error")
		s.Assert().Nil(msgFee, "GetMsgFee value")
	})

	s.Run("fee exists", func() {
		newMsgFee := msgfeetypes.MsgFee{
			MsgTypeUrl:           typeURL,
			AdditionalFee:        sdk.NewInt64Coin("feecoin", 8),
			Recipient:            "",
			RecipientBasisPoints: 0,
		}
		err := s.app.MsgFeesKeeper.SetMsgFee(s.ctx, newMsgFee)
		s.Require().NoError(err, "SetMsgFee error")
		err = RemoveLeaveGroupMsgFee(s.ctx, s.app)
		s.Require().NoError(err, "RemoveLeaveGroupMsgFee error")
		msgFee, err := s.app.MsgFeesKeeper.GetMsgFee(s.ctx, typeURL)
		s.Assert().NoError(err, "GetMsgFee error")
		s.Assert().Nil(msgFee, "GetMsgFee value")
	})
}
