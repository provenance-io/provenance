package keeper_test

import (
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/provenance-io/provenance/app"

	. "github.com/provenance-io/provenance/x/asset/keeper"
)

type KeeperTestSuite struct {
	suite.Suite
	app       *app.App
	ctx       sdk.Context
	user1Addr sdk.AccAddress
	k         Keeper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.app = app.Setup(s.T())
	s.k = s.app.AssetKeeper
	s.ctx = s.app.BaseApp.NewContextLegacy(false, cmtproto.Header{Time: time.Now()})
	priv := secp256k1.GenPrivKey()
	s.user1Addr = sdk.AccAddress(priv.PubKey().Address())
}

func (s *KeeperTestSuite) TestLogger() {
	logger := s.k.Logger(s.ctx)
	s.Require().NotNil(logger)
}

func (s *KeeperTestSuite) TestGetModuleAddress() {
	exp := authtypes.NewModuleAddress("asset")
	act := s.k.GetModuleAddress()
	s.Require().NotEmpty(act, "GetModuleAddress() result")
	s.Assert().Equal(exp, act, "GetModuleAddress() result")
}
