package ibchooks_test

import (
	"testing"

	"github.com/provenance-io/provenance/app"
	testutil "github.com/provenance-io/provenance/testutil/ibc"
	"github.com/provenance-io/provenance/x/ibchooks"
	markertypes "github.com/provenance-io/provenance/x/marker/types"

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

func (suite *MarkerHooksTestSuite) TestAddUpdateMarker() {
	markerHooks := ibchooks.NewMarkerHooks(&suite.chainA.GetProvenanceApp().MarkerKeeper)
	testCases := []struct {
		name        string
		denom       string
		memo        string
		expErr      string
		expIbcDenom string
	}{
		// {
		// 	name:        "successfully process with empty memo",
		// 	denom:       "fiftyfivehamburgers",
		// 	memo:        "",
		// 	expErr:      "",
		// 	expIbcDenom: "ibc/F3F4565153F3DD64470F075D6D6B1CB183F06EB55B287CCD0D3506277A03DE8E",
		// },
		// {
		// 	name:        "successfully process with non json memo",
		// 	denom:       "fiftyfivehamburgers",
		// 	memo:        "55 burger 55 fries...",
		// 	expErr:      "",
		// 	expIbcDenom: "ibc/F3F4565153F3DD64470F075D6D6B1CB183F06EB55B287CCD0D3506277A03DE8E",
		// },
		{
			name:        "successfully process with non json memo",
			denom:       "fiftyfivehamburgers",
			memo:        `{"marker":{random},"wasm":{"contract":"%1234","msg":{"echo":{"msg":"test"}}}}`,
			expErr:      "",
			expIbcDenom: "ibc/F3F4565153F3DD64470F075D6D6B1CB183F06EB55B287CCD0D3506277A03DE8E",
		},
	}
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			packet := suite.makeMockPacket(tc.denom, "", tc.memo, 0)
			err := markerHooks.AddUpdateMarker(suite.chainA.GetContext(), packet, suite.chainA.GetProvenanceApp().IBCKeeper)
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

func (suite *MarkerHooksTestSuite) TestProcessMarkerMemo() {
	testCases := []struct {
		name          string
		memo          string
		expAddresses  []sdk.AccAddress
		expMarkerType markertypes.MarkerType
		expErr        string
	}{
		// {
		// 	name:          "successfully process with non json memo",
		// 	memo:          `{"marker":{random},"wasm":{"contract":"%1234","msg":{"echo":{"msg":"test"}}}}`,
		// 	expAddresses:  []sdk.AccAddress{},
		// 	expMarkerType: markertypes.MarkerType_Coin,
		// 	expErr:        "",
		// },
		{
			name:          "successfully process no marker part",
			memo:          `{"marker":{"test":"test"},"wasm":{"contract":"%1234","msg":{"echo":{"msg":"test"}}}}`,
			expAddresses:  []sdk.AccAddress{},
			expMarkerType: markertypes.MarkerType_Coin,
			expErr:        "",
		},
	}
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			actualAddrs, actualMarkerType, err := ibchooks.ProcessMarkerMemo(tc.memo)
			if len(tc.expErr) > 0 {
				assert.EqualError(t, err, tc.expErr, "ProcessMarkerMemo() error")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expMarkerType, actualMarkerType, "Actual Marker type is incorrect")
				assert.Len(t, actualAddrs, len(tc.expAddresses), "Actual and expect address list must have same amount of elements")
				for _, addr := range tc.expAddresses {
					assert.Contains(t, actualAddrs, addr, "Actual list does not contain required address")

				}
			}
		})
	}
}
