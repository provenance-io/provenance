package keeper_test

import (
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/app"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/msgfees/types"
	"github.com/stretchr/testify/suite"
)

type MsgFeesParamTestSuite struct {
	suite.Suite

	app *app.App
	ctx sdk.Context
}

func (s *MsgFeesParamTestSuite) SetupTest() {
	s.app = simapp.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContextLegacy(false, cmtproto.Header{Time: time.Now()})
}

func TestMsgFeesParamTestSuite(t *testing.T) {
	suite.Run(t, new(MsgFeesParamTestSuite))
}

func (s *MsgFeesParamTestSuite) TestGetSetParams() {
	defaultParams := s.app.MsgFeesKeeper.GetParams(s.ctx)
	s.Require().Equal(types.DefaultFloorGasPrice(), defaultParams.FloorGasPrice, "Default FloorGasPrice should match")
	s.Require().Equal(types.DefaultParams().NhashPerUsdMil, defaultParams.NhashPerUsdMil, "Default NhashPerUsdMil should match")
	s.Require().Equal(types.DefaultParams().ConversionFeeDenom, defaultParams.ConversionFeeDenom, "Default ConversionFeeDenom should match")

	newFloorGasPrice := sdk.NewInt64Coin("nhash", 100)
	newNhashPerUsdMil := uint64(25000000)
	newConversionFeeDenom := "usd"

	newParams := types.Params{
		FloorGasPrice:      newFloorGasPrice,
		NhashPerUsdMil:     newNhashPerUsdMil,
		ConversionFeeDenom: newConversionFeeDenom,
	}

	s.app.MsgFeesKeeper.SetParams(s.ctx, newParams)

	updatedParams := s.app.MsgFeesKeeper.GetParams(s.ctx)
	s.Require().Equal(newFloorGasPrice, updatedParams.FloorGasPrice, "Updated FloorGasPrice should match")
	s.Require().Equal(newNhashPerUsdMil, updatedParams.NhashPerUsdMil, "Updated NhashPerUsdMil should match")
	s.Require().Equal(newConversionFeeDenom, updatedParams.ConversionFeeDenom, "Updated ConversionFeeDenom should match")
}
