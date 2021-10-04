package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	simapp "github.com/provenance-io/provenance/app"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/provenance-io/provenance/x/msgfees/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
	"testing"
	"time"
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
	app := simapp.Setup( false)
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
	app, ctx, addrs := s.app, s.ctx, s.addrs

	granterAddr := addrs[0]
	granteeAddr := addrs[1]
	//recipientAddr := addrs[2]

	s.T().Log("verify that creating msg fee for type works")
	authorization, expiration := app.AuthzKeeper.GetCleanAuthorization(ctx, granteeAddr, granterAddr, bankSendAuthMsgType)
	s.Require().Nil(authorization)
	s.Require().Equal(expiration, time.Time{})
	now := s.ctx.BlockHeader().Time
	s.Require().NotNil(now)
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
