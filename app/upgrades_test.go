package app

import (
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	icatypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	app *App
	ctx sdk.Context
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.app = Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
}

func (s *IntegrationTestSuite) TestUpgradeICA() {
	versionMap := s.app.UpgradeKeeper.GetModuleVersionMap(s.ctx)
	UpgradeICA(s.ctx, s.app, &versionMap)
	s.Assert().Equal(s.app.mm.Modules[icatypes.ModuleName].ConsensusVersion(), versionMap[icatypes.ModuleName], "consensus version should be set to skip init genesis")
	s.Assert().Equal([]string{"*"}, s.app.ICAHostKeeper.GetAllowMessages(s.ctx), "ica host should accept all messages")
	s.Assert().True(s.app.ICAHostKeeper.IsHostEnabled(s.ctx), "ica host should be enabled")
	s.Assert().Fail("it failed")
}
