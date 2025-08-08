package keeper_test

import (
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/asset/keeper"
	"github.com/provenance-io/provenance/x/asset/types"
)

type KeeperTestSuite struct {
	suite.Suite
	app       *app.App
	ctx       sdk.Context
	user1Addr sdk.AccAddress
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.app = app.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContextLegacy(false, cmtproto.Header{Time: time.Now()})
	priv := secp256k1.GenPrivKey()
	s.user1Addr = sdk.AccAddress(priv.PubKey().Address())
}

func (s *KeeperTestSuite) TestNewKeeper() {
	k := keeper.NewKeeper(
		s.app.AppCodec(),
		s.app.GetKey(types.ModuleName),
		s.app.NFTKeeper,
		s.app.BaseApp.MsgServiceRouter(),
		s.app.LedgerKeeper,
		s.app.RegistryKeeper,
		s.app.MarkerKeeper,
	)
	s.Require().NotNil(k)
}

func (s *KeeperTestSuite) TestLogger() {
	k := keeper.NewKeeper(
		s.app.AppCodec(),
		s.app.GetKey(types.ModuleName),
		s.app.NFTKeeper,
		s.app.BaseApp.MsgServiceRouter(),
		s.app.LedgerKeeper,
		s.app.RegistryKeeper,
		s.app.MarkerKeeper,
	)
	logger := k.Logger(s.ctx)
	s.Require().NotNil(logger)
}

func (s *KeeperTestSuite) TestGetModuleAddress() {
	k := keeper.NewKeeper(
		s.app.AppCodec(),
		s.app.GetKey(types.ModuleName),
		s.app.NFTKeeper,
		s.app.BaseApp.MsgServiceRouter(),
		s.app.LedgerKeeper,
		s.app.RegistryKeeper,
		s.app.MarkerKeeper,
	)
	addr := k.GetModuleAddress()
	s.Require().NotNil(addr)
}
