package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/msgfees/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
	"testing"
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
	types.RegisterQueryServer(queryHelper, app.MsgBasedFeeKeeper)
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
	msgFee, err := app.MsgBasedFeeKeeper.GetMsgBasedFeeSchedule(ctx, bankSendAuthMsgType)
	s.Require().Nil(msgFee)
	s.Require().Nil(err)
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
