package keeper_test

import (
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/app"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/metadata/types"
	"github.com/stretchr/testify/suite"
)

type OSLocatorParamTestSuite struct {
	suite.Suite

	app *app.App
	ctx sdk.Context
}

func (s *OSLocatorParamTestSuite) SetupTest() {
	s.app = simapp.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContextLegacy(false, cmtproto.Header{Time: time.Now()})
}

func TestOSLocatorParamTestSuite(t *testing.T) {
	suite.Run(t, new(OSLocatorParamTestSuite))
}

func (s *OSLocatorParamTestSuite) TestGetSetOSLocatorParams() {
	defaultParams := s.app.MetadataKeeper.GetOSLocatorParams(s.ctx)
	s.Require().Equal(int(types.DefaultMaxURILength), int(defaultParams.MaxUriLength), "GetOSLocatorParams() Default max URI length should match")

	defaultUriLength := s.app.MetadataKeeper.GetMaxURILength(s.ctx)
	s.Require().Equal(int(types.DefaultMaxURILength), int(defaultUriLength), "GetMaxURILength() Default max URI length should match")

	newMaxUriLength := uint32(2048)
	newParams := types.OSLocatorParams{
		MaxUriLength: newMaxUriLength,
	}
	s.app.MetadataKeeper.SetOSLocatorParams(s.ctx, newParams)

	updatedParams := s.app.MetadataKeeper.GetOSLocatorParams(s.ctx)
	s.Require().Equal(int(newMaxUriLength), int(updatedParams.MaxUriLength), "GetOSLocatorParams() Updated max URI length should match")

	updatedUriLength := s.app.MetadataKeeper.GetMaxURILength(s.ctx)
	s.Require().Equal(int(newMaxUriLength), int(updatedUriLength), "GetMaxURILength() Updated max URI length should match")
}
