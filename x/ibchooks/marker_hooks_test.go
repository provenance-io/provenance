package ibchooks_test

import (
	"testing"

	"github.com/provenance-io/provenance/app"
	testutil "github.com/provenance-io/provenance/testutil/ibc"
	"github.com/provenance-io/provenance/x/ibchooks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func (suite *MarkerHooksTestSuite) makeMockPacket(denom, receiver, memo string, prevSequence uint64) channeltypes.Packet {
	packetData := transfertypes.FungibleTokenPacketData{
		Denom:    denom,
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

func (suite *MarkerHooksTestSuite) TestProcessMarkerMemo() {
	markerHooks := ibchooks.NewMarkerHooks(&suite.chainA.GetProvenanceApp().MarkerKeeper)
	testCases := []struct {
		name        string
		denom       string
		memo        string
		expErr      string
		expIbcDenom string
	}{
		{
			"successfully process with empty memo",
			"fiftyfivehamburgers",
			"",
			"",
			"ibc/F3F4565153F3DD64470F075D6D6B1CB183F06EB55B287CCD0D3506277A03DE8E",
		},
	}
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			packet := suite.makeMockPacket(tc.denom, "", tc.memo, 0)
			err := markerHooks.ProcessMarkerMemo(suite.chainA.GetContext(), packet, suite.chainA.GetProvenanceApp().IBCKeeper)
			if len(tc.expErr) > 0 {
				assert.EqualError(t, err, tc.expErr, "ProcessMarkerMemo() error")
			} else {
				assert.NoError(t, err)
				marker, err := suite.chainA.GetProvenanceApp().MarkerKeeper.GetMarkerByDenom(suite.chainA.GetContext(), tc.expIbcDenom)
				require.NoError(t, err, "GetMarkerByDenom should find "+tc.expErr)
				assert.Equal(t, tc.expIbcDenom, marker.GetDenom(), "Marker Denom should be ibc denom")
				metadata, found := suite.chainA.GetProvenanceApp().BankKeeper.GetDenomMetaData(suite.chainA.GetContext(), tc.expIbcDenom)
				require.True(t, found, "GetDenomMetaData() not found for "+tc.expErr)
				assert.Equal(t, marker.GetDenom(), metadata.Base, "Metadata Base should equal marker denom")
				assert.Equal(t, "testchain2/"+tc.denom, metadata.Name, "Metadata Name should be chainid/denom")
				assert.Equal(t, "testchain2/"+tc.denom, metadata.Display, "Metadata Display should be chainid/denom")
				assert.Equal(t, tc.denom+" from chain testchain2", metadata.Description, "Metadata Description is incorrect")
			}
		})
	}
}
