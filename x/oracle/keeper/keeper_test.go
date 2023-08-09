package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/99designs/keyring"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/app"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/oracle/keeper"
	"github.com/provenance-io/provenance/x/oracle/types"
	"github.com/stretchr/testify/suite"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app         *simapp.App
	ctx         sdk.Context
	queryClient types.QueryClient
	msgServer   types.MsgServer

	accountAddr      sdk.AccAddress
	accountKey       *secp256k1.PrivKey
	keyring          keyring.Keyring
	keyringDir       string
	accountAddresses []sdk.AccAddress
}

func (s *KeeperTestSuite) CreateAccounts(number int) {
	for i := 0; i < number; i++ {
		accountKey := secp256k1.GenPrivKeyFromSecret([]byte(fmt.Sprintf("acc%d", i+2)))
		addr, err := sdk.AccAddressFromHexUnsafe(accountKey.PubKey().Address().String())
		s.Require().NoError(err)
		s.accountAddr = addr
		s.accountAddresses = append(s.accountAddresses, addr)
	}
}

func (s *KeeperTestSuite) SetupTest() {
	s.app = app.Setup(s.T())
	s.CreateAccounts(4)
	s.msgServer = keeper.NewMsgServerImpl(&s.app.OracleKeeper)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{Time: time.Now().UTC()})
	s.ctx = s.ctx.WithBlockHeight(100)

	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, s.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, s.app.OracleKeeper)
	s.queryClient = types.NewQueryClient(queryHelper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
