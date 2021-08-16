package marker_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/golang/protobuf/proto"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256r1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/marker"
	"github.com/provenance-io/provenance/x/marker/keeper"
	"github.com/provenance-io/provenance/x/marker/types"
)

type HandlerTestSuite struct {
	suite.Suite

	app     *app.App
	ctx     sdk.Context
	handler sdk.Handler

	pubkey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress

	pubkey2   cryptotypes.PubKey
	user2     string
	user2Addr sdk.AccAddress
}

func (s *HandlerTestSuite) SetupTest() {
	s.app = app.Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	s.handler = marker.NewHandler(s.app.MarkerKeeper)

	s.pubkey1 = secp256k1.GenPrivKey().PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()

	privKey, _ := secp256r1.GenPrivKey()
	s.pubkey2 = privKey.PubKey()
	s.user2Addr = sdk.AccAddress(s.pubkey2.Address())
	s.user2 = s.user2Addr.String()

	s.app.AccountKeeper.SetAccount(s.ctx, s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user1Addr))
	s.app.AccountKeeper.SetAccount(s.ctx, s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user2Addr))
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

func TestInvalidMsg(t *testing.T) {
	k := keeper.Keeper{}
	h := marker.NewHandler(k)

	res, err := h(sdk.NewContext(nil, tmproto.Header{}, false, nil), testdata.NewTestMsg())
	require.Error(t, err)
	require.Nil(t, res)
	require.Contains(t, err.Error(), "unrecognized marker message type")
	require.Contains(t, err.Error(), "testdata.TestMsg")
}

func TestInvalidProposal(t *testing.T) {
	k := keeper.Keeper{}
	h := marker.NewProposalHandler(k)

	err := h(sdk.NewContext(nil, tmproto.Header{}, false, nil), govtypes.NewTextProposal("Test", "description"))
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "unrecognized marker proposal content type: *types.TextProposal"))
}

func (s HandlerTestSuite) containsMessage(result *sdk.Result, msg proto.Message) bool {
	events := result.GetEvents().ToABCIEvents()
	for _, event := range events {
		typeEvent, _ := sdk.ParseTypedEvent(event)
		if assert.ObjectsAreEqual(msg, typeEvent) {
			return true
		}
	}
	return false
}

type CommonTest struct {
	name          string
	msg           sdk.Msg
	signers       []string
	errorMsg      string
	expectedEvent proto.Message
}

func (s HandlerTestSuite) runTests(cases []CommonTest) {
	for _, tc := range cases {
		s.T().Run(tc.name, func(t *testing.T) {
			response, err := s.handler(s.ctx, tc.msg)

			if len(tc.errorMsg) > 0 {
				assert.EqualError(t, err, tc.errorMsg)
			} else {
				if tc.expectedEvent != nil {
					result := s.containsMessage(response, tc.expectedEvent)
					s.True(result, fmt.Sprintf("Expected typed event was not found: %v", tc.expectedEvent))
				}

			}
		})
	}
}

func (s HandlerTestSuite) TestMsgAddMarkerRequest() {
	denom := "hotdog"
	denomWithDashPeriod := fmt.Sprintf("%s-my.marker", denom)
	activeStatus := types.NewMsgAddMarkerRequest(denom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true)
	activeStatus.Status = types.StatusActive

	undefinedStatus := types.NewMsgAddMarkerRequest(denom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true)
	undefinedStatus.Status = types.StatusUndefined

	cases := []CommonTest{
		{
			"should successfully ADD new marker",
			types.NewMsgAddMarkerRequest(denom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true),
			[]string{s.user1},
			"",
			types.NewEventMarkerAdd(denom, "100", "proposed", s.user1, types.MarkerType_Coin.String()),
		},
		{
			"should fail to ADD new marker, validate basic failure",
			undefinedStatus,
			[]string{s.user1},
			"invalid marker status: invalid request",
			nil,
		},
		{
			"should fail to ADD new marker, invalid status",
			activeStatus,
			[]string{s.user1},
			"marker can only be created with a Proposed or Finalized status: invalid request",
			nil,
		},
		{
			"should fail to ADD new marker, marker already exists",
			types.NewMsgAddMarkerRequest(denom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true),
			[]string{s.user1},
			fmt.Sprintf("marker address already exists for %s: invalid request", types.MustGetMarkerAddress(denom)),
			nil,
		},
		{
			"should successfully add marker with dash and period",
			types.NewMsgAddMarkerRequest(denomWithDashPeriod, sdk.NewInt(1000), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true),
			[]string{s.user1},
			"",
			types.NewEventMarkerAdd(denomWithDashPeriod, "1000", "proposed", s.user1, types.MarkerType_Coin.String()),
		},
	}
	s.runTests(cases)
}

func (s HandlerTestSuite) TestMsgAddAccessRequest() {

	accessMintGrant := types.AccessGrant{
		Address:     s.user1,
		Permissions: types.AccessListByNames("MINT"),
	}

	accessInvalidGrant := types.AccessGrant{
		Address:     s.user1,
		Permissions: types.AccessListByNames("Invalid"),
	}

	cases := []CommonTest{
		{
			"setup new marker for test",
			types.NewMsgAddMarkerRequest("hotdog", sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true),
			[]string{s.user2},
			"",
			nil,
		},
		{
			"should successfully grant access to marker",
			types.NewMsgAddAccessRequest("hotdog", s.user1Addr, accessMintGrant),

			[]string{s.user2},
			"",
			types.NewEventMarkerAddAccess(&accessMintGrant, "hotdog", s.user1),
		},
		{
			"should fail to ADD access to marker, validate basic fails",
			types.NewMsgAddAccessRequest("hotdog", s.user1Addr, accessInvalidGrant),
			[]string{s.user2},
			"invalid access type: invalid request",
			nil,
		},
		{
			"should fail to ADD access to marker, keeper AddAccess failure",
			types.NewMsgAddAccessRequest("hotdog", s.user2Addr, accessMintGrant),
			[]string{s.user2},
			fmt.Sprintf("updates to pending marker hotdog can only be made by %s: unauthorized", s.user1),
			nil,
		},
	}

	s.runTests(cases)
}

func (s HandlerTestSuite) TestMsgDeleteAccessMarkerRequest() {

	hotdogDenom := "hotdog"
	accessMintGrant := types.AccessGrant{
		Address:     s.user1,
		Permissions: types.AccessListByNames("MINT"),
	}

	cases := []CommonTest{
		{
			"setup new marker for test",
			types.NewMsgAddMarkerRequest(hotdogDenom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true),
			[]string{s.user1},
			"",
			nil,
		},
		{
			"setup grant access to marker",
			types.NewMsgAddAccessRequest(hotdogDenom, s.user1Addr, accessMintGrant),
			[]string{s.user1},
			"",
			nil,
		},
		{
			"should successfully delete grant access to marker",
			types.NewDeleteAccessRequest(hotdogDenom, s.user1Addr, s.user1Addr),
			[]string{s.user1},
			"",
			types.NewEventMarkerDeleteAccess(s.user1, hotdogDenom, s.user1),
		},
	}
	s.runTests(cases)
}

func (s HandlerTestSuite) TestMsgFinalizeMarkerRequest() {

	hotdogDenom := "hotdog"

	cases := []CommonTest{
		{
			"setup new marker for test",
			types.NewMsgAddMarkerRequest(hotdogDenom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true),
			[]string{s.user1},
			"",
			nil,
		},
		{
			"should successfully finalize marker",
			types.NewMsgFinalizeRequest(hotdogDenom, s.user1Addr),
			[]string{s.user1},
			"",
			types.NewEventMarkerFinalize(hotdogDenom, s.user1),
		},
	}
	s.runTests(cases)
}

func (s HandlerTestSuite) TestMsgActivateMarkerRequest() {

	hotdogDenom := "hotdog"

	cases := []CommonTest{
		{
			"setup new marker for test",
			types.NewMsgAddMarkerRequest(hotdogDenom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true),
			[]string{s.user1},
			"",
			nil,
		},
		{
			"setup finalize marker",
			types.NewMsgFinalizeRequest(hotdogDenom, s.user1Addr),
			[]string{s.user1},
			"",
			nil,
		},
		{
			"should successfully activate marker",
			types.NewMsgActivateRequest(hotdogDenom, s.user1Addr),
			[]string{s.user1},
			"",
			types.NewEventMarkerActivate(hotdogDenom, s.user1),
		},
	}
	s.runTests(cases)
}

func (s HandlerTestSuite) TestMsgCancelMarkerRequest() {

	hotdogDenom := "hotdog"
	accessDeleteGrant := types.AccessGrant{
		Address:     s.user1,
		Permissions: types.AccessListByNames("DELETE"),
	}

	cases := []CommonTest{
		{
			"setup new marker for test",
			types.NewMsgAddMarkerRequest(hotdogDenom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true),
			[]string{s.user1},
			"",
			nil,
		},
		{
			"setup grant delete access to marker",
			types.NewMsgAddAccessRequest(hotdogDenom, s.user1Addr, accessDeleteGrant),
			[]string{s.user1},
			"",
			nil,
		},
		{
			"should successfully cancel marker",
			types.NewMsgCancelRequest(hotdogDenom, s.user1Addr),
			[]string{s.user1},
			"",
			types.NewEventMarkerCancel(hotdogDenom, s.user1),
		},
	}
	s.runTests(cases)
}

func (s HandlerTestSuite) TestMsgDeleteMarkerRequest() {

	hotdogDenom := "hotdog"
	accessDeleteMintGrant := types.AccessGrant{
		Address:     s.user1,
		Permissions: types.AccessListByNames("DELETE,MINT"),
	}

	cases := []CommonTest{
		{
			"setup new marker for test",
			types.NewMsgAddMarkerRequest(hotdogDenom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true),
			[]string{s.user1},
			"",
			nil,
		},
		{
			"setup grant delete access to marker",
			types.NewMsgAddAccessRequest(hotdogDenom, s.user1Addr, accessDeleteMintGrant),
			[]string{s.user1},
			"",
			nil,
		},
		{
			"setup cancel marker",
			types.NewMsgCancelRequest(hotdogDenom, s.user1Addr),
			[]string{s.user1},
			"",
			nil,
		},
		{
			"should successfully delete marker",
			types.NewMsgDeleteRequest(hotdogDenom, s.user1Addr),
			[]string{s.user1},
			"",
			types.NewEventMarkerDelete(hotdogDenom, s.user1),
		},
	}
	s.runTests(cases)
}

func (s HandlerTestSuite) TestMsgMintMarkerRequest() {

	hotdogDenom := "hotdog"
	access := types.AccessGrant{
		Address:     s.user1,
		Permissions: types.AccessListByNames("MINT,BURN"),
	}

	cases := []CommonTest{
		{
			"setup new marker for test",
			types.NewMsgAddMarkerRequest(hotdogDenom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true),
			[]string{s.user1},
			"",
			nil,
		},
		{
			"setup grant mint access to marker",
			types.NewMsgAddAccessRequest(hotdogDenom, s.user1Addr, access),
			[]string{s.user1},
			"",
			nil,
		},
		{
			"should successfully mint marker",
			types.NewMsgMintRequest(s.user1Addr, sdk.NewCoin(hotdogDenom, sdk.NewInt(100))),
			[]string{s.user1},
			"",
			types.NewEventMarkerMint("100", hotdogDenom, s.user1),
		},
	}
	s.runTests(cases)
}

func (s HandlerTestSuite) TestMsgBurnMarkerRequest() {

	hotdogDenom := "hotdog"
	access := types.AccessGrant{
		Address:     s.user1,
		Permissions: types.AccessListByNames("DELETE,MINT,BURN"),
	}

	cases := []CommonTest{
		{
			"setup new marker for test",
			types.NewMsgAddMarkerRequest(hotdogDenom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true),
			[]string{s.user1},
			"",
			nil,
		},
		{
			"setup grant mint access to marker",
			types.NewMsgAddAccessRequest(hotdogDenom, s.user1Addr, access),
			[]string{s.user1},
			"",
			nil,
		},
		{
			"should successfully burn marker",
			types.NewMsgBurnRequest(s.user1Addr, sdk.NewCoin(hotdogDenom, sdk.NewInt(100))),
			[]string{s.user1},
			"",
			types.NewEventMarkerBurn("100", hotdogDenom, s.user1),
		},
	}
	s.runTests(cases)
}

func (s HandlerTestSuite) TestMsgWithdrawMarkerRequest() {

	hotdogDenom := "hotdog"
	access := types.AccessGrant{
		Address:     s.user1,
		Permissions: types.AccessListByNames("DELETE,MINT,WITHDRAW"),
	}

	cases := []CommonTest{
		{
			"setup new marker for test",
			types.NewMsgAddMarkerRequest(hotdogDenom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true),
			[]string{s.user1},
			"",
			nil,
		},
		{
			"setup grant access to marker",
			types.NewMsgAddAccessRequest(hotdogDenom, s.user1Addr, access),
			[]string{s.user1},
			"",
			nil,
		},
		{
			"setup finalize marker",
			types.NewMsgFinalizeRequest(hotdogDenom, s.user1Addr),
			[]string{s.user1},
			"",
			nil,
		},
		{
			"setup activate marker",
			types.NewMsgActivateRequest(hotdogDenom, s.user1Addr),
			[]string{s.user1},
			"",
			nil,
		},
		{
			"should successfully withdraw marker",
			types.NewMsgWithdrawRequest(s.user1Addr, s.user1Addr, hotdogDenom, sdk.NewCoins(sdk.NewCoin(hotdogDenom, sdk.NewInt(100)))),
			[]string{s.user1},
			"",
			types.NewEventMarkerWithdraw("100hotdog", hotdogDenom, s.user1, s.user1),
		},
	}
	s.runTests(cases)
}

func (s HandlerTestSuite) TestMsgTransferMarkerRequest() {

	hotdogDenom := "hotdog"
	access := types.AccessGrant{
		Address:     s.user1,
		Permissions: types.AccessListByNames("DELETE,MINT,WITHDRAW,TRANSFER"),
	}

	cases := []CommonTest{
		{
			"setup new marker for test",
			types.NewMsgAddMarkerRequest(hotdogDenom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_RestrictedCoin, true, true),
			[]string{s.user1},
			"",
			nil,
		},
		{
			"setup grant access to marker",
			types.NewMsgAddAccessRequest(hotdogDenom, s.user1Addr, access),
			[]string{s.user1},
			"",
			nil,
		},
		{
			"setup finalize marker",
			types.NewMsgFinalizeRequest(hotdogDenom, s.user1Addr),
			[]string{s.user1},
			"",
			nil,
		},
		{
			"setup activate marker",
			types.NewMsgActivateRequest(hotdogDenom, s.user1Addr),
			[]string{s.user1},
			"",
			nil,
		},
		{
			"should successfully mint marker",
			types.NewMsgMintRequest(s.user1Addr, sdk.NewCoin(hotdogDenom, sdk.NewInt(1000))),
			[]string{s.user1},
			"",
			nil,
		},
		{
			"should successfully transfer marker",
			types.NewMsgTransferRequest(s.user1Addr, s.user1Addr, s.user2Addr, sdk.NewCoin(hotdogDenom, sdk.NewInt(0))),
			[]string{s.user1},
			"",
			types.NewEventMarkerTransfer("0", hotdogDenom, s.user1, s.user2, s.user1),
		},
	}
	s.runTests(cases)
}

func (s HandlerTestSuite) TestMsgSetDenomMetadataRequest() {

	hotdogDenom := "hotdog"
	hotdogName := "Jason"
	hotdogSymbol := "WIFI"
	access := types.AccessGrant{
		Address:     s.user1,
		Permissions: types.AccessListByNames("DELETE,MINT,WITHDRAW,TRANSFER"),
	}

	hotdogMetadata := banktypes.Metadata{
		Description: "a description",
		DenomUnits: []*banktypes.DenomUnit{
			{Denom: fmt.Sprintf("n%s", hotdogDenom), Exponent: 0, Aliases: []string{fmt.Sprintf("nano%s", hotdogDenom)}},
			{Denom: fmt.Sprintf("u%s", hotdogDenom), Exponent: 3, Aliases: []string{}},
			{Denom: hotdogDenom, Exponent: 9, Aliases: []string{}},
			{Denom: fmt.Sprintf("mega%s", hotdogDenom), Exponent: 15, Aliases: []string{}},
		},
		Base:    fmt.Sprintf("n%s", hotdogDenom),
		Display: hotdogDenom,
		Name:    hotdogName,
		Symbol:  hotdogSymbol,
	}

	cases := []CommonTest{
		{
			"setup new marker for test",
			types.NewMsgAddMarkerRequest(fmt.Sprintf("n%s", hotdogDenom), sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_RestrictedCoin, true, true),
			[]string{s.user1},
			"",
			nil,
		},
		{
			"setup grant access to marker",
			types.NewMsgAddAccessRequest(fmt.Sprintf("n%s", hotdogDenom), s.user1Addr, access),
			[]string{s.user1},
			"",
			nil,
		},
		{
			"should successfully set denom metadata on marker",
			types.NewSetDenomMetadataRequest(hotdogMetadata, s.user1Addr),
			[]string{s.user1},
			"",
			types.NewEventMarkerSetDenomMetadata(hotdogMetadata, s.user1),
		},
	}
	s.runTests(cases)
}
