package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/msgfees/types"
)

type TestSuite struct {
	suite.Suite

	app         *simapp.App
	ctx         sdk.Context
	addrs       []sdk.AccAddress
	queryClient types.QueryClient
}

var bankSendAuthMsgType = banktypes.SendAuthorization{}.MsgTypeURL()

func (s *TestSuite) SetupTest() {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: now})
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.MsgFeesKeeper)
	queryClient := types.NewQueryClient(queryHelper)
	s.queryClient = queryClient

	s.app = app
	s.ctx = ctx
	s.queryClient = queryClient
	s.addrs = simapp.AddTestAddrsIncremental(app, ctx, 3, sdk.NewInt(30000000))
}

func (s *TestSuite) TestKeeper() {
	app, ctx, _ := s.app, s.ctx, s.addrs

	s.T().Log("verify that creating msg fee for type works")
	msgFee, err := app.MsgFeesKeeper.GetMsgFee(ctx, bankSendAuthMsgType)
	s.Require().Nil(msgFee)
	s.Require().Nil(err)
	newCoin := sdk.NewInt64Coin("steak", 100)
	msgFeeToCreate := types.NewMsgFee(bankSendAuthMsgType, newCoin)
	app.MsgFeesKeeper.SetMsgFee(ctx, msgFeeToCreate)

	msgFee, err = app.MsgFeesKeeper.GetMsgFee(ctx, bankSendAuthMsgType)
	s.Require().NotNil(msgFee)
	s.Require().Nil(err)
	s.Require().Equal(bankSendAuthMsgType, msgFee.MsgTypeUrl)

	msgFee, err = app.MsgFeesKeeper.GetMsgFee(ctx, "does-not-exist")
	s.Require().Nil(err)
	s.Require().Nil(msgFee)

	err = app.MsgFeesKeeper.RemoveMsgFee(ctx, bankSendAuthMsgType)
	s.Require().Nil(err)
	msgFee, err = app.MsgFeesKeeper.GetMsgFee(ctx, bankSendAuthMsgType)
	s.Require().Nil(msgFee)
	s.Require().Nil(err)

	err = app.MsgFeesKeeper.RemoveMsgFee(ctx, "does-not-exist")
	s.Require().ErrorIs(err, types.ErrMsgFeeDoesNotExist)

}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
