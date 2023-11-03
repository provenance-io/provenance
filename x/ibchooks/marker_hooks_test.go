package ibchooks_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/provenance-io/provenance/app"
	testutil "github.com/provenance-io/provenance/testutil/ibc"
	"github.com/provenance-io/provenance/x/ibchooks"
	"github.com/provenance-io/provenance/x/marker/types"
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
	suite.chainA.GetProvenanceApp().BankKeeper.MintCoins(suite.chainA.GetContext(), markertypes.CoinPoolName, sdk.NewCoins(sdk.NewInt64Coin("ibc/F7466BCD642C14163B3E67D5B4401FD2B77271C3225FDE0C57ADD61B8046253D", 100)))
	address1 := sdk.AccAddress("address1")
	address2 := sdk.AccAddress("address2")
	markerHooks := ibchooks.NewMarkerHooks(&suite.chainA.GetProvenanceApp().MarkerKeeper)
	testCases := []struct {
		name          string
		denom         string
		memo          string
		expErr        string
		expIbcDenom   string
		expTransAuths []sdk.AccAddress
		expSupply     sdk.Coin
	}{
		{
			name:        "successfully process with empty memo",
			denom:       "fiftyfivehamburgers",
			memo:        "",
			expErr:      "",
			expIbcDenom: "ibc/F3F4565153F3DD64470F075D6D6B1CB183F06EB55B287CCD0D3506277A03DE8E",
			expSupply:   sdk.NewInt64Coin("ibc/F3F4565153F3DD64470F075D6D6B1CB183F06EB55B287CCD0D3506277A03DE8E", 1),
		},
		{
			name:        "successfully process with non json memo",
			denom:       "fiftyfivehamburgers",
			memo:        "55 burger 55 fries...",
			expErr:      "",
			expIbcDenom: "ibc/F3F4565153F3DD64470F075D6D6B1CB183F06EB55B287CCD0D3506277A03DE8E",
			expSupply:   sdk.NewInt64Coin("ibc/F3F4565153F3DD64470F075D6D6B1CB183F06EB55B287CCD0D3506277A03DE8E", 1),
		},
		{
			name:        "successfully process with non json marker part memo",
			denom:       "fiftyfivehamburgers",
			memo:        `{"marker":{random},"wasm":{"contract":"%1234","msg":{"echo":{"msg":"test"}}}}`,
			expErr:      "",
			expIbcDenom: "ibc/F3F4565153F3DD64470F075D6D6B1CB183F06EB55B287CCD0D3506277A03DE8E",
			expSupply:   sdk.NewInt64Coin("ibc/F3F4565153F3DD64470F075D6D6B1CB183F06EB55B287CCD0D3506277A03DE8E", 1),
		},
		{
			name:          "successfully process with transfer auths",
			denom:         "fiftyfivefries",
			memo:          fmt.Sprintf(`{"marker":{"transfer-auths":["%s", "%s"]}}`, address1.String(), address2.String()),
			expErr:        "",
			expIbcDenom:   "ibc/1B3A5773661E8A6B9F6BB407979B5933C2FA792DF24ED2A40B028C90277B0C22",
			expTransAuths: []sdk.AccAddress{address1, address2},
			expSupply:     sdk.NewInt64Coin("ibc/1B3A5773661E8A6B9F6BB407979B5933C2FA792DF24ED2A40B028C90277B0C22", 1),
		},
		{
			name:   "fail invalid json",
			denom:  "fiftyfivetacos",
			memo:   fmt.Sprintf(`{"marker":{"transfer-auths":"%s"}}`, address1.String()),
			expErr: "json: cannot unmarshal string into Go struct field MarkerPayload.transfer-auths of type []string",
		},
		{
			name:        "successfully process with ibc denom that existed before marker correctly adjust supply",
			denom:       "ibcdenombeforemiddleware",
			memo:        "",
			expErr:      "",
			expIbcDenom: "ibc/F7466BCD642C14163B3E67D5B4401FD2B77271C3225FDE0C57ADD61B8046253D",
			expSupply:   sdk.NewInt64Coin("ibc/F7466BCD642C14163B3E67D5B4401FD2B77271C3225FDE0C57ADD61B8046253D", 101),
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
				assert.Equal(t, tc.expSupply, marker.GetSupply(), "Marker Supply should match expected")
				metadata, found := suite.chainA.GetProvenanceApp().BankKeeper.GetDenomMetaData(suite.chainA.GetContext(), tc.expIbcDenom)
				require.True(t, found, "GetDenomMetaData() not found for "+tc.expErr)
				assert.Equal(t, marker.GetDenom(), metadata.Base, "Metadata Base should equal marker denom")
				assert.Equal(t, "testchain2/"+tc.denom, metadata.Name, "Metadata Name should be chainid/denom")
				assert.Equal(t, "testchain2/"+tc.denom, metadata.Display, "Metadata Display should be chainid/denom")
				assert.Equal(t, tc.denom+" from testchain2", metadata.Description, "Metadata Description is incorrect")
				assert.Len(t, marker.GetAccessList(), len(tc.expTransAuths), "Resulting access list does not equal expect length")
				for _, access := range marker.GetAccessList() {
					assert.Len(t, access.GetAccessList(), 1, "Expecting permissions list to only one item")
					assert.Equal(t, access.GetAccessList()[0], markertypes.Access_Transfer, "Expecting permissions to be transfer")
					assert.Contains(t, tc.expTransAuths, sdk.MustAccAddressFromBech32(access.Address), "Actual list does not contain required address")
				}
			}
		})
	}
}

func (suite *MarkerHooksTestSuite) TestProcessMarkerMemo() {
	address1 := sdk.AccAddress("address1")
	address2 := sdk.AccAddress("address2")
	testCases := []struct {
		name                  string
		memo                  string
		expAddresses          []sdk.AccAddress
		expMarkerType         markertypes.MarkerType
		expAllowForceTransfer bool
		expErr                string
	}{
		{
			name:          "successfully process with non json memo",
			memo:          `{"marker":{random},"wasm":{"contract":"%1234","msg":{"echo":{"msg":"test"}}}}`,
			expAddresses:  []sdk.AccAddress{},
			expMarkerType: markertypes.MarkerType_Coin,
			expErr:        "",
		},
		{
			name:          "successfully process marker ignore unknown property",
			memo:          `{"marker":{"test":"test"},"wasm":{"contract":"%1234","msg":{"echo":{"msg":"test"}}}}`,
			expAddresses:  []sdk.AccAddress{},
			expMarkerType: markertypes.MarkerType_Coin,
			expErr:        "",
		},
		{
			name:          "transfer auth in correct type",
			memo:          `{"marker":{"transfer-auths":"incorrect data type"}}`,
			expAddresses:  []sdk.AccAddress{},
			expMarkerType: markertypes.MarkerType_Coin,
			expErr:        "json: cannot unmarshal string into Go struct field MarkerPayload.transfer-auths of type []string",
		},
		{
			name:          "transfer auth in correct address bech32 value",
			memo:          `{"marker":{"transfer-auths":["invalidbech32"]}}`,
			expAddresses:  []sdk.AccAddress{},
			expMarkerType: markertypes.MarkerType_Coin,
			expErr:        "decoding bech32 failed: invalid separator index -1",
		},
		{
			name:          "transfer auth in correct type",
			memo:          fmt.Sprintf(`{"marker":{"transfer-auths":["%s", "%s"]}}`, address1.String(), address2.String()),
			expAddresses:  []sdk.AccAddress{address2, address1},
			expMarkerType: markertypes.MarkerType_RestrictedCoin,
			expErr:        "",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			actualAddrs, actualMarkerType, actualAllowForceTransfer, err := ibchooks.ProcessMarkerMemo(tc.memo)
			if len(tc.expErr) > 0 {
				assert.EqualError(t, err, tc.expErr, "ProcessMarkerMemo() error")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expMarkerType, actualMarkerType, "Actual Marker type is incorrect")
				assert.Len(t, actualAddrs, len(tc.expAddresses), "Actual and expect address list must have same amount of elements")
				for _, addr := range tc.expAddresses {
					assert.Contains(t, actualAddrs, addr, "Actual list does not contain required address")

				}
				assert.Equal(t, tc.expAllowForceTransfer, actualAllowForceTransfer, "Actual allow force transfer is incorrect")
			}
		})
	}
}

func (suite *MarkerHooksTestSuite) TestResetMarkerAccessGrants() {
	address1 := sdk.AccAddress("address1")
	address2 := sdk.AccAddress("address2")
	address3 := sdk.AccAddress("address3")
	testCases := []struct {
		name          string
		transferAuths []sdk.AccAddress
		markerAcct    markertypes.MarkerAccount
		expErr        string
	}{
		{
			name:          "successfully reset marker access grants and remove all others",
			transferAuths: []sdk.AccAddress{address1, address2},
			markerAcct:    *markertypes.NewEmptyMarkerAccount("jackthecat", address1.String(), []types.AccessGrant{*types.NewAccessGrant(address1, []types.Access{types.Access_Burn}), *types.NewAccessGrant(address1, []types.Access{types.Access_Admin})}),
			expErr:        "",
		},
		{
			name:          "successfully reset marker access grants and remove all others",
			transferAuths: []sdk.AccAddress{address1, address2},
			markerAcct:    *markertypes.NewEmptyMarkerAccount("jackthecat", address1.String(), []types.AccessGrant{}),
			expErr:        "",
		},
		{
			name:          "successfully reset marker access grants and remove other transfer grant",
			transferAuths: []sdk.AccAddress{address1, address2},
			markerAcct:    *markertypes.NewEmptyMarkerAccount("jackthecat", address1.String(), []types.AccessGrant{*types.NewAccessGrant(address3, []types.Access{types.Access_Transfer})}),
			expErr:        "",
		},
		{
			name:          "successful with empty transfer auths",
			transferAuths: []sdk.AccAddress{},
			markerAcct:    *markertypes.NewEmptyMarkerAccount("jackthecat", address1.String(), []types.AccessGrant{}),
			expErr:        "",
		},
		{
			name:          "successful with nil transfer auths",
			transferAuths: nil,
			markerAcct:    *markertypes.NewEmptyMarkerAccount("jackthecat", address1.String(), []types.AccessGrant{}),
			expErr:        "",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			err := ibchooks.ResetMarkerAccessGrants(tc.transferAuths, &tc.markerAcct)
			if len(tc.expErr) > 0 {
				assert.EqualError(t, err, tc.expErr, "ResetMarkerAccessGrants() error")
			} else {
				require.NoError(t, err)
				assert.Len(t, tc.markerAcct.GetAccessList(), len(tc.transferAuths), "Resulting access list does not equal expect length")
				for _, access := range tc.markerAcct.GetAccessList() {
					assert.Len(t, access.GetAccessList(), 1, "Expecting permissions list to only one item")
					assert.Equal(t, access.GetAccessList()[0], markertypes.Access_Transfer, "Expecting permissions to be transfer")
					assert.Contains(t, tc.transferAuths, sdk.MustAccAddressFromBech32(access.Address), "Actual list does not contain required address")
				}
			}
		})
	}
}

func (suite *MarkerHooksTestSuite) TestSanitizeMemo() {
	testCases := []struct {
		name    string
		memo    string
		expMemo string
	}{
		{
			name:    "plain text memo",
			memo:    "plain text user memo",
			expMemo: `{"memo":"plain text user memo"}`,
		},
		{
			name:    "mal-formed json should be moved to memo",
			memo:    `{"marker":{"transfer-auths":["123", "345"]}`,
			expMemo: "{\"memo\":\"{\\\"marker\\\":{\\\"transfer-auths\\\":[\\\"123\\\", \\\"345\\\"]}\"}",
		},
		{
			name:    "correct json should not modify memo",
			memo:    `{"marker":{"transfer-auths":["address"]}}`,
			expMemo: `{"marker":{"transfer-auths":["address"]}}`,
		},
		{
			name:    "empty memo",
			memo:    "",
			expMemo: `{}`,
		},
	}
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			actualMemoJSON := ibchooks.SanitizeMemo(tc.memo)
			actualMemo, err := json.Marshal(actualMemoJSON)
			require.NoError(t, err, "json.Mashal() failed to mashal memo")
			assert.Equal(t, tc.expMemo, string(actualMemo), "SanitizeMemo() should have transformed memo to expected memo")
		})
	}
}

func (suite *MarkerHooksTestSuite) TestPreSendPacketDataProcessingFn() {
	address1 := sdk.AccAddress("address1")
	markerHooks := ibchooks.NewMarkerHooks(&suite.chainA.GetProvenanceApp().MarkerKeeper)
	marker1 := *markertypes.NewEmptyMarkerAccount("jackthecat", address1.String(), []types.AccessGrant{*types.NewAccessGrant(address1, []types.Access{types.Access_Transfer}), *types.NewAccessGrant(address1, []types.Access{types.Access_Admin})})
	marker1.MarkerType = markertypes.MarkerType_RestrictedCoin
	require.NoError(suite.T(), suite.chainA.GetProvenanceApp().MarkerKeeper.AddMarkerAccount(suite.chainA.GetContext(), &marker1), "AddMarkerAccount() in test setup")
	testCases := []struct {
		name    string
		data    []byte
		expData []byte
		expErr  string
	}{
		{
			name:    "not a ics20 packet",
			data:    []byte{0, 12},
			expData: []byte{0, 12},
		},
		{
			name:    "packet with plain non json memo",
			data:    suite.makeMockPacket("hotdogs", "recieverAddr", "my memo", 0).Data,
			expData: suite.makeMockPacket("hotdogs", "recieverAddr", `{"marker":{},"memo":"my memo"}`, 0).Data,
		},
		{
			name:    "packet with marker json",
			data:    suite.makeMockPacket("hotdogs", "recieverAddr", `{"marker":{"test":"test"},"wasm":{"contract":"%1234","msg":{"echo":{"msg":"test"}}}}`, 0).Data,
			expData: suite.makeMockPacket("hotdogs", "recieverAddr", `{"marker":{},"wasm":{"contract":"%1234","msg":{"echo":{"msg":"test"}}}}`, 0).Data,
		},
		{
			name:    "packet with marker json with non-transfer auth addresses ",
			data:    suite.makeMockPacket("hotdogs", "recieverAddr", `{"marker":{"transfer-auths":["test"]}}`, 0).Data,
			expData: suite.makeMockPacket("hotdogs", "recieverAddr", `{"marker":{}}`, 0).Data,
		},
		{
			name:    "packet with marker json replace transfer auths with marker transfer auths",
			data:    suite.makeMockPacket("jackthecat", "recieverAddr", `{"marker":{"transfer-auths":["test"]}}`, 0).Data,
			expData: suite.makeMockPacket("jackthecat", "recieverAddr", fmt.Sprintf(`{"marker":{"transfer-auths":["%s"],"allow-force-transfer":false}}`, address1.String()), 0).Data,
		},
		{
			name:   "invalid denom should error",
			data:   suite.makeMockPacket("~~", "recieverAddr", "my memo", 0).Data,
			expErr: "invalid denom: ~~",
		},
	}
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			actualData, err := markerHooks.SetupMarkerMemoFn(suite.chainA.GetContext(), tc.data, nil)
			if len(tc.expErr) > 0 {
				assert.EqualError(t, err, tc.expErr, "PreSendPacketDataProcessingFn() error")
				assert.Nil(t, actualData, "PreSendPacketDataProcessingFn() return `data` should be nil")
			} else {
				require.NoError(t, err, "PreSendPacketDataProcessingFn() was expecting no error")
				assert.Equal(t, tc.expData, actualData, "PreSendPacketDataProcessingFn() `data` not expected")
			}
		})
	}
}
