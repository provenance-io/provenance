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

func (s *TestSuite) TestConvertDenomToHash() {
	app, ctx, _ := s.app, s.ctx, s.addrs
	usdDollar := sdk.NewCoin("usd", sdk.NewInt(7_000)) // $7.00 == 100hash
	nhash, err := app.MsgFeesKeeper.ConvertDenomToHash(ctx, usdDollar)
	s.Assert().NoError(err)
	s.Assert().Equal(sdk.NewCoin("nhash", sdk.NewInt(100_000_000_000)), nhash)
	usdDollar = sdk.NewCoin("usd", sdk.NewInt(70)) // $7 == 1hash
	nhash, err = app.MsgFeesKeeper.ConvertDenomToHash(ctx, usdDollar)
	s.Assert().NoError(err)
	s.Assert().Equal(sdk.NewCoin("nhash", sdk.NewInt(1_000_000_000)), nhash)
	usdDollar = sdk.NewCoin("usd", sdk.NewInt(1_000)) // $1 == 14.2hash
	nhash, err = app.MsgFeesKeeper.ConvertDenomToHash(ctx, usdDollar)
	s.Assert().NoError(err)
	s.Assert().Equal(sdk.NewCoin("nhash", sdk.NewInt(14_000_000_000)), nhash)

	usdDollar = sdk.NewCoin("usd", sdk.NewInt(10))
	nhash, err = app.MsgFeesKeeper.ConvertDenomToHash(ctx, usdDollar)
	s.Assert().NoError(err)
	s.Assert().Equal(sdk.NewCoin("nhash", sdk.NewInt(700000000)), nhash)

	jackTheCat := sdk.NewCoin("jackThecat", sdk.NewInt(70))
	nhash, err = app.MsgFeesKeeper.ConvertDenomToHash(ctx, jackTheCat)
	s.Assert().Equal("denom not supported for conversion jackThecat: invalid type", err.Error())
	s.Assert().Equal(sdk.Coin{}, nhash)
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
