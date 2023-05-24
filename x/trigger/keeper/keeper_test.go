package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"

	"github.com/provenance-io/provenance/app"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/trigger"
	"github.com/provenance-io/provenance/x/trigger/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

var (
	PKs = simapp.CreateTestPubKeys(500)
)

type KeeperTestSuite struct {
	suite.Suite

	app         *simapp.App
	ctx         sdk.Context
	queryClient types.QueryClient
	handler     sdk.Handler

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
	s.handler = trigger.NewHandler(s.app.TriggerKeeper)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{Time: time.Now().UTC()})

	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, s.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, s.app.TriggerKeeper)
	s.queryClient = types.NewQueryClient(queryHelper)

	s.SetupEventHistory()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupEventHistory() {
	attributes1 := []sdk.Attribute{
		sdk.NewAttribute("key1", "value1"),
		sdk.NewAttribute("key2", "value2"),
		sdk.NewAttribute("key3", "value3"),
	}
	attributes2 := []sdk.Attribute{
		sdk.NewAttribute("key1", "value1"),
		sdk.NewAttribute("key3", "value2"),
		sdk.NewAttribute("key4", "value3"),
	}
	event1 := sdk.NewEvent("event1", attributes1...)
	event2 := sdk.NewEvent("event2", attributes2...)
	event3 := sdk.NewEvent("event1", attributes1...)
	loggedEvents := sdk.Events{
		event1,
		event2,
		event3,
	}
	eventManagerStub := sdk.NewEventManagerWithHistory(loggedEvents.ToABCIEvents())
	s.ctx = s.ctx.WithEventManager(eventManagerStub)
}
