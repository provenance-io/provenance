package keeper_test

import (
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/app"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/ibchooks/types"
	"github.com/stretchr/testify/suite"
)

type IbcHooksParamTestSuite struct {
	suite.Suite

	app *app.App
	ctx sdk.Context
}

func (s *IbcHooksParamTestSuite) SetupTest() {
	s.app = simapp.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContextLegacy(false, cmtproto.Header{Time: time.Now()})
}

func TestIbcHooksParamTestSuite(t *testing.T) {
	suite.Run(t, new(IbcHooksParamTestSuite))
}

func (s *IbcHooksParamTestSuite) TestGetSetParams() {
	defaultParams := s.app.IBCHooksKeeper.GetParams(s.ctx)
	s.Require().Len(defaultParams.AllowedAsyncAckContracts, 0, "Default AllowedAsyncAckContracts should be empty")

	newAllowedAsyncAckContracts := []string{"contract1", "contract2", "contract3"}
	newParams := types.Params{
		AllowedAsyncAckContracts: newAllowedAsyncAckContracts,
	}

	s.app.IBCHooksKeeper.SetParams(s.ctx, newParams)

	updatedParams := s.app.IBCHooksKeeper.GetParams(s.ctx)
	s.Require().Equal(newAllowedAsyncAckContracts, updatedParams.AllowedAsyncAckContracts, "Updated AllowedAsyncAckContracts should match")
}
