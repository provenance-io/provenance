package ibchooks_test

import (
	"testing"

	"github.com/provenance-io/provenance/app"
	testutil "github.com/provenance-io/provenance/testutil/ibc"
	"github.com/provenance-io/provenance/x/ibchooks"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v6/testing"
)

type MarkerHooksTestSuite struct {
	suite.Suite

	App         *app.App
	Ctx         sdk.Context
	QueryHelper *baseapp.QueryServiceTestHelper
	TestAccs    []sdk.AccAddress

	coordinator *ibctesting.Coordinator

	chainA *testutil.TestChain
	chainB *testutil.TestChain

	path *ibctesting.Path
}

func init() {
	ibctesting.DefaultTestingAppInit = SetupSimApp
}

func (suite *MarkerHooksTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	suite.chainA = &testutil.TestChain{
		TestChain: suite.coordinator.GetChain(ibctesting.GetChainID(1)),
	}
	suite.chainB = &testutil.TestChain{
		TestChain: suite.coordinator.GetChain(ibctesting.GetChainID(2)),
	}
	suite.path = NewTransferPath(suite.chainA, suite.chainB)
	suite.coordinator.Setup(suite.path)
}

func TestMarkerHooksTestSuite(t *testing.T) {
	suite.Run(t, new(MarkerHooksTestSuite))
}

func (suite *MarkerHooksTestSuite) makeMockPacket(receiver, memo string, prevSequence uint64) channeltypes.Packet {
	packetData := transfertypes.FungibleTokenPacketData{
		Denom:    sdk.DefaultBondDenom,
		Amount:   "1",
		Sender:   suite.chainB.SenderAccount.GetAddress().String(),
		Receiver: receiver,
		Memo:     memo,
	}

	return channeltypes.NewPacket(
		packetData.GetBytes(),
		prevSequence+1,
		suite.path.EndpointB.ChannelConfig.PortID,
		suite.path.EndpointB.ChannelID,
		suite.path.EndpointA.ChannelConfig.PortID,
		suite.path.EndpointA.ChannelID,
		clienttypes.NewHeight(0, 100),
		0,
	)
}

func (suite *MarkerHooksTestSuite) TestProcessMarkerMemoHooks() {
	packet := suite.makeMockPacket("", "", 0)
	markerHooks := ibchooks.NewMarkerHooks(&suite.chainA.GetProvenanceApp().MarkerKeeper)
	markerHooks.ProcessMarkerMemo(suite.chainA.GetContext(), packet, suite.chainA.GetProvenanceApp().IBCKeeper)
}
