package app

import (
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
	s.app = Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
}
