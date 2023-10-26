package keeper_test

import (
	"testing"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/ibcratelimit/keeper"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite

	app    *app.App
	ctx    sdk.Context
	keeper keeper.Keeper
}

func (s *TestSuite) SetupTest() {
	s.app = app.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	s.ctx = s.ctx.WithBlockHeight(0)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestNewKeeper() {
	tests := []struct {
		name string
	}{}

	for _, tc := range tests {
		s.Run(tc.name, func() {

		})
	}
}

func (s *TestSuite) TestGetSetParams() {
	tests := []struct {
		name string
	}{}

	for _, tc := range tests {
		s.Run(tc.name, func() {

		})
	}
}

func (s *TestSuite) TestGetContractAddress() {
	tests := []struct {
		name string
	}{}

	for _, tc := range tests {
		s.Run(tc.name, func() {

		})
	}
}

func (s *TestSuite) TestContractConfigured() {
	tests := []struct {
		name string
	}{}

	for _, tc := range tests {
		s.Run(tc.name, func() {

		})
	}
}
