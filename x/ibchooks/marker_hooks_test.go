package ibchooks_test

import (
	"fmt"
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

	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v10/testing"
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

func (suite *MarkerHooksTestSuite) SetupTest() {
	ibctesting.DefaultTestingAppInit = SetupSimAppFn(suite.T())
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

func (suite *MarkerHooksTestSuite) TestAddMarker() {
	// Compute IBC denoms dynamically since channel IDs vary between test runs.
	hamburgerIBCDenom := ibchooks.MustExtractDenomFromPacketOnRecv(suite.makeMockPacket("fiftyfivehamburgers", "", "", 0))
	friesIBCDenom := ibchooks.MustExtractDenomFromPacketOnRecv(suite.makeMockPacket("fiftyfivefries", "", "", 0))
	tacosIBCDenom := ibchooks.MustExtractDenomFromPacketOnRecv(suite.makeMockPacket("fiftyfivetacos", "", "", 0))
	ibcBeforeIBCDenom := ibchooks.MustExtractDenomFromPacketOnRecv(suite.makeMockPacket("ibcdenombeforemiddleware", "", "", 0))

	suite.chainA.GetProvenanceApp().BankKeeper.MintCoins(suite.chainA.GetContext(), markertypes.CoinPoolName, sdk.NewCoins(sdk.NewInt64Coin(ibcBeforeIBCDenom, 100)))
	address1 := sdk.AccAddress("address1")
	address2 := sdk.AccAddress("address2")
	markerHooks := ibchooks.NewMarkerHooks(&suite.chainA.GetProvenanceApp().MarkerKeeper)
	testCases := []struct {
		name        string
		denom       string
		memo        string
		expErr      string
		expIbcDenom string
		expSupply   sdk.Coin
	}{
		{
			name:        "successfully process with empty memo",
			denom:       "fiftyfivehamburgers",
			memo:        "",
			expErr:      "",
			expIbcDenom: hamburgerIBCDenom,
			expSupply:   sdk.NewInt64Coin(hamburgerIBCDenom, 1),
		},
		{
			name:        "successfully process with non json memo",
			denom:       "fiftyfivehamburgers",
			memo:        "55 burger 55 fries...",
			expErr:      "",
			expIbcDenom: hamburgerIBCDenom,
			expSupply:   sdk.NewInt64Coin(hamburgerIBCDenom, 1),
		},
		{
			name:        "successfully process with non json marker part memo",
			denom:       "fiftyfivehamburgers",
			memo:        `{"marker":{random},"wasm":{"contract":"%1234","msg":{"echo":{"msg":"test"}}}}`,
			expErr:      "",
			expIbcDenom: hamburgerIBCDenom,
			expSupply:   sdk.NewInt64Coin(hamburgerIBCDenom, 1),
		},
		{
			name:        "old memo style with transfer auths",
			denom:       "fiftyfivefries",
			memo:        fmt.Sprintf(`{"marker":{"transfer-auths":["%s", "%s"]}}`, address1.String(), address2.String()),
			expErr:      "",
			expIbcDenom: friesIBCDenom,
			expSupply:   sdk.NewInt64Coin(friesIBCDenom, 1),
		},
		{
			name:        "invalid json is ignored",
			denom:       "fiftyfivetacos",
			memo:        fmt.Sprintf(`{"marker":{"transfer-auths":"%s"}}`, address1.String()),
			expIbcDenom: tacosIBCDenom,
			expSupply:   sdk.NewInt64Coin(tacosIBCDenom, 1),
		},
		{
			name:        "successfully process with ibc denom that existed before marker correctly adjust supply",
			denom:       "ibcdenombeforemiddleware",
			memo:        "",
			expErr:      "",
			expIbcDenom: ibcBeforeIBCDenom,
			expSupply:   sdk.NewInt64Coin(ibcBeforeIBCDenom, 101),
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			packet := suite.makeMockPacket(tc.denom, "", tc.memo, 0)
			err := markerHooks.AddMarker(suite.chainA.GetContext(), packet, suite.chainA.GetProvenanceApp().IBCKeeper)
			if len(tc.expErr) > 0 {
				assert.EqualError(t, err, tc.expErr, "ProcessMarkerMemo() error")
			} else {
				assert.NoError(t, err)
				marker, err := suite.chainA.GetProvenanceApp().MarkerKeeper.GetMarkerByDenom(suite.chainA.GetContext(), tc.expIbcDenom)
				require.NoError(t, err, "GetMarkerByDenom should find "+tc.expErr)
				assert.Equal(t, tc.expIbcDenom, marker.GetDenom(), "Marker Denom should be ibc denom")
				assert.Equal(t, tc.expSupply, marker.GetSupply(), "Marker Supply should match expected")
				metadata, found := suite.chainA.GetProvenanceApp().BankKeeper.GetDenomMetaData(suite.chainA.GetContext(), tc.expIbcDenom)
				require.True(t, found, "GetDenomMetaData() not found for "+tc.expErr)
				assert.Equal(t, marker.GetDenom(), metadata.Base, "Metadata Base should equal marker denom")
				assert.Equal(t, "testchain2-1/"+tc.denom, metadata.Name, "Metadata Name should be chainid/denom")
				assert.Equal(t, "testchain2-1/"+tc.denom, metadata.Display, "Metadata Display should be chainid/denom")
				assert.Equal(t, tc.denom+" from testchain2-1", metadata.Description, "Metadata Description is incorrect")
				assert.Len(t, marker.GetAccessList(), 0, "Resulting access list does not equal expect length")
			}
		})
	}
}
