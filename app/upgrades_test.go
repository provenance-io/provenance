package app

import (
	"fmt"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

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

func (s *IntegrationTestSuite) TestRemoveIsSendEnabledEntries() {
	s.SetupSuite()
	for i := 0; i < 50; i++ {
		denom := fmt.Sprintf("denom%v", i)
		rdenom := fmt.Sprintf("rdenom%v", i)
		s.app.BankKeeper.SetSendEnabled(s.ctx, denom, true)
		s.app.BankKeeper.SetSendEnabled(s.ctx, rdenom, false)
	}
	RemoveIsSendEnabledEntries(s.ctx, s.app)
	sendEnabledItems := s.app.BankKeeper.GetAllSendEnabledEntries(s.ctx)
	s.Assert().Equal(s, sendEnabledItems, 0, "all items should have been removed from table")
	// probably overkill here.
	for i := 0; i < 50; i++ {
		denom := fmt.Sprintf("denom%v", i)
		rdenom := fmt.Sprintf("rdenom%v", i)
		s.Assert().True(s.app.BankKeeper.IsSendEnabledCoin(s.ctx, sdk.NewInt64Coin(denom, 1)))
		s.Assert().True(s.app.BankKeeper.IsSendEnabledCoin(s.ctx, sdk.NewInt64Coin(rdenom, 1)))
	}
}
