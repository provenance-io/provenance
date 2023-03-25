package marker_test

import (
	"fmt"
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
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

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
	s.app = app.Setup(s.T())
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

	err := h(sdk.NewContext(nil, tmproto.Header{}, false, nil), govtypesv1beta1.NewTextProposal("Test", "description"))
	require.ErrorContains(t, err, "unrecognized marker proposal content type: *v1beta1.TextProposal")
}

func (s *HandlerTestSuite) containsMessage(result *sdk.Result, msg proto.Message) bool {
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

func (s *HandlerTestSuite) runTests(cases []CommonTest) {
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

func (s *HandlerTestSuite) TestMsgAddMarkerRequest() {
	denom := "hotdog"
	rdenom := "restrictedhotdog"
	denomWithDashPeriod := fmt.Sprintf("%s-my.marker", denom)
	activeStatus := types.NewMsgAddMarkerRequest(denom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true, false, []string{})
	activeStatus.Status = types.StatusActive

	undefinedStatus := types.NewMsgAddMarkerRequest(denom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true, false, []string{})
	undefinedStatus.Status = types.StatusUndefined

	cases := []CommonTest{
		{
			name: "should successfully ADD new marker",
			msg: types.NewMsgAddMarkerRequest(denom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true, false, []string{}),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: types.NewEventMarkerAdd(denom, "100", "proposed", s.user1, types.MarkerType_Coin.String()),
		},
		{
			name: "should fail to ADD new marker, validate basic failure",
			msg: undefinedStatus,
			signers: []string{s.user1},
			errorMsg: "invalid marker status: invalid request",
			expectedEvent: nil,
		},
		{
			name: "should fail to ADD new marker, invalid status",
			msg: activeStatus,
			signers: []string{s.user1},
			errorMsg: "a marker can not be created in an ACTIVE status: invalid request",
			expectedEvent: nil,
		},
		{
			name: "should fail to ADD new marker, marker already exists",
			msg: types.NewMsgAddMarkerRequest(denom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true, false, []string{}),
			signers: []string{s.user1},
			errorMsg: fmt.Sprintf("marker address already exists for %s: invalid request", types.MustGetMarkerAddress(denom)),
			expectedEvent: nil,
		},
		{
			name: "should successfully add marker with dash and period",
			msg: types.NewMsgAddMarkerRequest(denomWithDashPeriod, sdk.NewInt(1000), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true, false, []string{}),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: types.NewEventMarkerAdd(denomWithDashPeriod, "1000", "proposed", s.user1, types.MarkerType_Coin.String()),
		},
		{
			name: "should successfully ADD new marker with required attributes",
			msg: types.NewMsgAddMarkerRequest(rdenom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_RestrictedCoin, true, true, false, []string{"attribute.one.com", "attribute.two.com"}),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: types.NewEventMarkerAdd(rdenom, "100", "proposed", s.user1, types.MarkerType_RestrictedCoin.String()),
		},
	}
	s.runTests(cases)
}

func (s *HandlerTestSuite) TestMsgAddAccessRequest() {

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
			name: "setup new marker for test",
			msg: types.NewMsgAddMarkerRequest("hotdog", sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true, false, []string{}),
			signers: []string{s.user2},
			errorMsg: "",
			expectedEvent: nil,
		},
		{
			name: "should successfully grant access to marker",
			msg: types.NewMsgAddAccessRequest("hotdog", s.user1Addr, accessMintGrant),

			signers: []string{s.user2},
			errorMsg: "",
			expectedEvent: types.NewEventMarkerAddAccess(&accessMintGrant, "hotdog", s.user1),
		},
		{
			name: "should fail to ADD access to marker, validate basic fails",
			msg: types.NewMsgAddAccessRequest("hotdog", s.user1Addr, accessInvalidGrant),
			signers: []string{s.user2},
			errorMsg: "invalid access type: invalid request",
			expectedEvent: nil,
		},
		{
			name: "should fail to ADD access to marker, keeper AddAccess failure",
			msg: types.NewMsgAddAccessRequest("hotdog", s.user2Addr, accessMintGrant),
			signers: []string{s.user2},
			errorMsg: fmt.Sprintf("updates to pending marker hotdog can only be made by %s: unauthorized", s.user1),
			expectedEvent: nil,
		},
	}

	s.runTests(cases)
}

func (s *HandlerTestSuite) TestMsgDeleteAccessMarkerRequest() {

	hotdogDenom := "hotdog"
	accessMintGrant := types.AccessGrant{
		Address:     s.user1,
		Permissions: types.AccessListByNames("MINT"),
	}

	cases := []CommonTest{
		{
			name: "setup new marker for test",
			msg: types.NewMsgAddMarkerRequest(hotdogDenom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true, false, []string{}),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: nil,
		},
		{
			name: "setup grant access to marker",
			msg: types.NewMsgAddAccessRequest(hotdogDenom, s.user1Addr, accessMintGrant),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: nil,
		},
		{
			name: "should successfully delete grant access to marker",
			msg: types.NewDeleteAccessRequest(hotdogDenom, s.user1Addr, s.user1Addr),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: types.NewEventMarkerDeleteAccess(s.user1, hotdogDenom, s.user1),
		},
	}
	s.runTests(cases)
}

func (s *HandlerTestSuite) TestMsgFinalizeMarkerRequest() {

	hotdogDenom := "hotdog"

	cases := []CommonTest{
		{
			name: "setup new marker for test",
			msg: types.NewMsgAddMarkerRequest(hotdogDenom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true, false, []string{}),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: nil,
		},
		{
			name: "should successfully finalize marker",
			msg: types.NewMsgFinalizeRequest(hotdogDenom, s.user1Addr),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: types.NewEventMarkerFinalize(hotdogDenom, s.user1),
		},
	}
	s.runTests(cases)
}

func (s *HandlerTestSuite) TestMsgActivateMarkerRequest() {

	hotdogDenom := "hotdog"

	cases := []CommonTest{
		{
			name: "setup new marker for test",
			msg: types.NewMsgAddMarkerRequest(hotdogDenom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true, false, []string{}),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: nil,
		},
		{
			name: "setup finalize marker",
			msg: types.NewMsgFinalizeRequest(hotdogDenom, s.user1Addr),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: nil,
		},
		{
			name: "should successfully activate marker",
			msg: types.NewMsgActivateRequest(hotdogDenom, s.user1Addr),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: types.NewEventMarkerActivate(hotdogDenom, s.user1),
		},
	}
	s.runTests(cases)
}

func (s *HandlerTestSuite) TestMsgCancelMarkerRequest() {

	hotdogDenom := "hotdog"
	accessDeleteGrant := types.AccessGrant{
		Address:     s.user1,
		Permissions: types.AccessListByNames("DELETE"),
	}

	cases := []CommonTest{
		{
			name: "setup new marker for test",
			msg: types.NewMsgAddMarkerRequest(hotdogDenom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true, false, []string{}),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: nil,
		},
		{
			name: "setup grant delete access to marker",
			msg: types.NewMsgAddAccessRequest(hotdogDenom, s.user1Addr, accessDeleteGrant),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: nil,
		},
		{
			name: "should successfully cancel marker",
			msg: types.NewMsgCancelRequest(hotdogDenom, s.user1Addr),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: types.NewEventMarkerCancel(hotdogDenom, s.user1),
		},
	}
	s.runTests(cases)
}

func (s *HandlerTestSuite) TestMsgDeleteMarkerRequest() {

	hotdogDenom := "hotdog"
	accessDeleteMintGrant := types.AccessGrant{
		Address:     s.user1,
		Permissions: types.AccessListByNames("DELETE,MINT"),
	}

	cases := []CommonTest{
		{
			name: "setup new marker for test",
			msg: types.NewMsgAddMarkerRequest(hotdogDenom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true, false, []string{}),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: nil,
		},
		{
			name: "setup grant delete access to marker",
			msg: types.NewMsgAddAccessRequest(hotdogDenom, s.user1Addr, accessDeleteMintGrant),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: nil,
		},
		{
			name: "setup cancel marker",
			msg: types.NewMsgCancelRequest(hotdogDenom, s.user1Addr),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: nil,
		},
		{
			name: "should successfully delete marker",
			msg: types.NewMsgDeleteRequest(hotdogDenom, s.user1Addr),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: types.NewEventMarkerDelete(hotdogDenom, s.user1),
		},
	}
	s.runTests(cases)
}

func (s *HandlerTestSuite) TestMsgMintMarkerRequest() {

	hotdogDenom := "hotdog"
	access := types.AccessGrant{
		Address:     s.user1,
		Permissions: types.AccessListByNames("MINT,BURN"),
	}

	cases := []CommonTest{
		{
			name: "setup new marker for test",
			msg: types.NewMsgAddMarkerRequest(hotdogDenom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true, false, []string{}),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: nil,
		},
		{
			name: "setup grant mint access to marker",
			msg: types.NewMsgAddAccessRequest(hotdogDenom, s.user1Addr, access),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: nil,
		},
		{
			name: "should successfully mint marker",
			msg: types.NewMsgMintRequest(s.user1Addr, sdk.NewCoin(hotdogDenom, sdk.NewInt(100))),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: types.NewEventMarkerMint("100", hotdogDenom, s.user1),
		},
	}
	s.runTests(cases)
}

func (s *HandlerTestSuite) TestMsgBurnMarkerRequest() {

	hotdogDenom := "hotdog"
	access := types.AccessGrant{
		Address:     s.user1,
		Permissions: types.AccessListByNames("DELETE,MINT,BURN"),
	}

	cases := []CommonTest{
		{
			name: "setup new marker for test",
			msg: types.NewMsgAddMarkerRequest(hotdogDenom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true, false, []string{}),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: nil,
		},
		{
			name: "setup grant mint access to marker",
			msg: types.NewMsgAddAccessRequest(hotdogDenom, s.user1Addr, access),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: nil,
		},
		{
			name: "should successfully burn marker",
			msg: types.NewMsgBurnRequest(s.user1Addr, sdk.NewCoin(hotdogDenom, sdk.NewInt(100))),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: types.NewEventMarkerBurn("100", hotdogDenom, s.user1),
		},
	}
	s.runTests(cases)
}

func (s *HandlerTestSuite) TestMsgWithdrawMarkerRequest() {

	hotdogDenom := "hotdog"
	access := types.AccessGrant{
		Address:     s.user1,
		Permissions: types.AccessListByNames("DELETE,MINT,WITHDRAW"),
	}

	cases := []CommonTest{
		{
			name: "setup new marker for test",
			msg: types.NewMsgAddMarkerRequest(hotdogDenom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true, false, []string{}),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: nil,
		},
		{
			name: "setup grant access to marker",
			msg: types.NewMsgAddAccessRequest(hotdogDenom, s.user1Addr, access),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: nil,
		},
		{
			name: "setup finalize marker",
			msg: types.NewMsgFinalizeRequest(hotdogDenom, s.user1Addr),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: nil,
		},
		{
			name: "setup activate marker",
			msg: types.NewMsgActivateRequest(hotdogDenom, s.user1Addr),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: nil,
		},
		{
			name: "should successfully withdraw marker",
			msg: types.NewMsgWithdrawRequest(s.user1Addr, s.user1Addr, hotdogDenom, sdk.NewCoins(sdk.NewCoin(hotdogDenom, sdk.NewInt(100)))),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: types.NewEventMarkerWithdraw("100hotdog", hotdogDenom, s.user1, s.user1),
		},
	}
	s.runTests(cases)
}

func (s *HandlerTestSuite) TestMsgTransferMarkerRequest() {

	hotdogDenom := "hotdog"
	access := types.AccessGrant{
		Address:     s.user1,
		Permissions: types.AccessListByNames("DELETE,MINT,WITHDRAW,TRANSFER"),
	}

	cases := []CommonTest{
		{
			name: "setup new marker for test",
			msg: types.NewMsgAddMarkerRequest(hotdogDenom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_RestrictedCoin, true, true, false, []string{}),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: nil,
		},
		{
			name: "setup grant access to marker",
			msg: types.NewMsgAddAccessRequest(hotdogDenom, s.user1Addr, access),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: nil,
		},
		{
			name: "setup finalize marker",
			msg: types.NewMsgFinalizeRequest(hotdogDenom, s.user1Addr),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: nil,
		},
		{
			name: "setup activate marker",
			msg: types.NewMsgActivateRequest(hotdogDenom, s.user1Addr),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: nil,
		},
		{
			name: "should successfully mint marker",
			msg: types.NewMsgMintRequest(s.user1Addr, sdk.NewCoin(hotdogDenom, sdk.NewInt(1000))),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: nil,
		},
		{
			name: "should successfully transfer marker",
			msg: types.NewMsgTransferRequest(s.user1Addr, s.user1Addr, s.user2Addr, sdk.NewCoin(hotdogDenom, sdk.NewInt(0))),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: types.NewEventMarkerTransfer("0", hotdogDenom, s.user1, s.user2, s.user1),
		},
	}
	s.runTests(cases)
}

func (s *HandlerTestSuite) TestMsgSetDenomMetadataRequest() {

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
			name: "setup new marker for test",
			msg: types.NewMsgAddMarkerRequest(fmt.Sprintf("n%s", hotdogDenom), sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_RestrictedCoin, true, true, false, []string{}),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: nil,
		},
		{
			name: "setup grant access to marker",
			msg: types.NewMsgAddAccessRequest(fmt.Sprintf("n%s", hotdogDenom), s.user1Addr, access),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: nil,
		},
		{
			name: "should successfully set denom metadata on marker",
			msg: types.NewSetDenomMetadataRequest(hotdogMetadata, s.user1Addr),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: types.NewEventMarkerSetDenomMetadata(hotdogMetadata, s.user1),
		},
	}
	s.runTests(cases)
}

func (s *HandlerTestSuite) TestMsgAddFinalizeActivateMarkerRequest() {
	denom := "hotdog"
	rdenom := "restrictedhotdog"
	denomWithDashPeriod := fmt.Sprintf("%s-my.marker", denom)
	msgWithActiveStatus := types.NewMsgAddFinalizeActivateMarkerRequest(denom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true, false, []string{}, []types.AccessGrant{*types.NewAccessGrant(s.user1Addr, []types.Access{types.Access_Mint, types.Access_Admin})})
	msgWithActiveStatusAttr := types.NewMsgAddFinalizeActivateMarkerRequest(rdenom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_RestrictedCoin, true, true, false, []string{"attributes.one.com", "attributes.two.com"}, []types.AccessGrant{*types.NewAccessGrant(s.user1Addr, []types.Access{types.Access_Mint, types.Access_Admin})})

	accessGrantWrongStatus := types.NewMsgAddFinalizeActivateMarkerRequest(denom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true, false, []string{}, nil)

	cases := []CommonTest{
		{
			name: "should successfully ADD,FINALIZE,ACTIVATE new marker",
			msg: msgWithActiveStatus,
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: types.NewEventMarkerAdd(denom, "100", "proposed", s.user1, types.MarkerType_Coin.String()),
		},
		{
			name: "should successfully ADD,FINALIZE,ACTIVATE new marker with attributes",
			msg: msgWithActiveStatusAttr,
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: types.NewEventMarkerAdd(rdenom, "100", "proposed", s.user1, types.MarkerType_RestrictedCoin.String()),
		},
		{
			name: "should fail to ADD,FINALIZE,ACTIVATE new marker, validate basic failure",
			msg: accessGrantWrongStatus,
			signers: []string{s.user1},
			errorMsg: "since this will activate the marker, must have at least one access list defined: invalid request",
			expectedEvent: nil,
		},
		{
			name: "should fail to ADD,FINALIZE,ACTIVATE new marker, marker already exists",
			msg: types.NewMsgAddMarkerRequest(denom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true, false, []string{}),
			signers: []string{s.user1},
			errorMsg: fmt.Sprintf("marker address already exists for %s: invalid request", types.MustGetMarkerAddress(denom)),
			expectedEvent: nil,
		},
		{
			name: "should successfully add marker with dash and period",
			msg: types.NewMsgAddMarkerRequest(denomWithDashPeriod, sdk.NewInt(1000), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true, false, []string{}),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: types.NewEventMarkerAdd(denomWithDashPeriod, "1000", "proposed", s.user1, types.MarkerType_Coin.String()),
		},
		{
			name: "should successfully mint denom",
			msg: types.NewMsgMintRequest(s.user1Addr, sdk.NewInt64Coin(denom, 1000)),
			signers: []string{s.user1},
			errorMsg: "",
			expectedEvent: types.NewEventMarkerMint("1000", denom, s.user1),
		},
		{
			name: "should fail to  burn denom, user doesn't have permissions",
			msg: types.NewMsgBurnRequest(s.user1Addr, sdk.NewInt64Coin(denom, 50)),
			signers: []string{s.user1},
			errorMsg: fmt.Sprintf("%s does not have ACCESS_BURN on hotdog markeraccount: invalid request", s.user1),
			expectedEvent: nil,
		},
	}
	s.runTests(cases)
}
