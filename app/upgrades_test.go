package app

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	app *App
	ctx sdk.Context
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.app = Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
}

func (s *IntegrationTestSuite) TestRemoveIsSendEnabledEntries() {
	s.SetupSuite()
	markerAddr := markertypes.MustGetMarkerAddress("testcoin").String()

	err := s.app.MarkerKeeper.AddMarkerAccount(s.ctx, &markertypes.MarkerAccount{
		BaseAccount: &authtypes.BaseAccount{
			Address:       markerAddr,
			AccountNumber: 23,
		},
		AccessControl: []markertypes.AccessGrant{
			{
				Address:     "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
				Permissions: markertypes.AccessListByNames("deposit,withdraw"),
			},
		},
		Denom:      "testcoin",
		Supply:     sdk.NewInt(1000),
		MarkerType: markertypes.MarkerType_Coin,
		Status:     markertypes.StatusActive,
	})
	s.Assert().NoError(err, "should have added marker")
	s.app.BankKeeper.SetSendEnabled(s.ctx, "testcoin", false)
	s.app.BankKeeper.SetSendEnabled(s.ctx, "nonmarkercoin", false)

	sendEnabledItems := s.app.BankKeeper.GetAllSendEnabledEntries(s.ctx)
	s.Assert().Equal(len(sendEnabledItems), 2, "should have 2 items before removal")

	RemoveIsSendEnabledEntries(s.ctx, s.app)
	sendEnabledItems = s.app.BankKeeper.GetAllSendEnabledEntries(s.ctx)
	s.Assert().Equal(len(sendEnabledItems), 1, "denom without a marker should only exist")
	s.Assert().Equal(sendEnabledItems[0].Denom, "nonmarkercoin", "denom without a marker should only exist")
	s.Assert().False(s.app.BankKeeper.IsSendEnabledDenom(s.ctx, "nonmarkercoin"), "should be in table as false since there is not a marker associated with it")
	s.Assert().True(s.app.BankKeeper.IsSendEnabledDenom(s.ctx, "testcoin"), "should not exist in table therefore default to true")

}
