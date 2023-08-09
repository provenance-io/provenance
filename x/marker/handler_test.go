package marker_test

import (
	"fmt"
	"testing"

	"github.com/golang/protobuf/proto"

	attrtypes "github.com/provenance-io/provenance/x/attribute/types"
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

	s.app.MarkerKeeper.AddMarkerAccount(s.ctx, types.NewEmptyMarkerAccount("navcoin", s.user1, []types.AccessGrant{}))
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
	errorMsg      string
	expectedEvent proto.Message
}

func (s *HandlerTestSuite) runTests(cases []CommonTest) {
	for _, tc := range cases {
		s.Run(tc.name, func() {
			response, err := s.handler(s.ctx, tc.msg)
			if len(tc.errorMsg) > 0 {
				s.Require().EqualError(err, tc.errorMsg, "handler(%T) error", tc.msg)
			} else {
				s.Require().NoError(err, "handler(%T) error", tc.msg)
				if tc.expectedEvent != nil {
					result := s.containsMessage(response, tc.expectedEvent)
					s.Assert().True(result, "Expected typed event was not found in response.\n    Expected: %+v\n    Response: %+v", tc.expectedEvent, response)
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

	cases := []CommonTest{
		{
			name:          "should successfully ADD new marker",
			msg:           types.NewMsgAddMarkerRequest(denom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true, false, []string{}),
			expectedEvent: types.NewEventMarkerAdd(denom, types.MustGetMarkerAddress(denom).String(), "100", "proposed", s.user1, types.MarkerType_Coin.String()),
		},
		{
			name:     "should fail to ADD new marker, invalid status",
			msg:      activeStatus,
			errorMsg: "marker can only be created with a Proposed or Finalized status",
		},
		{
			name:     "should fail to ADD new marker, marker already exists",
			msg:      types.NewMsgAddMarkerRequest(denom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true, false, []string{}),
			errorMsg: fmt.Sprintf("marker address already exists for %s: invalid request", types.MustGetMarkerAddress(denom)),
		},
		{

			name:          "should successfully add marker with dash and period",
			msg:           types.NewMsgAddMarkerRequest(denomWithDashPeriod, sdk.NewInt(1000), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true, false, []string{}),
			expectedEvent: types.NewEventMarkerAdd(denomWithDashPeriod, types.MustGetMarkerAddress(denomWithDashPeriod).String(), "1000", "proposed", s.user1, types.MarkerType_Coin.String()),
		},
		{

			name:          "should successfully ADD new marker with required attributes",
			msg:           types.NewMsgAddMarkerRequest(rdenom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_RestrictedCoin, true, true, false, []string{"attribute.one.com", "attribute.two.com"}),
			expectedEvent: types.NewEventMarkerAdd(rdenom, types.MustGetMarkerAddress(rdenom).String(), "100", "proposed", s.user1, types.MarkerType_RestrictedCoin.String()),
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
			msg:  types.NewMsgAddMarkerRequest("hotdog", sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true, false, []string{}),
		},
		{
			name:          "should successfully grant access to marker",
			msg:           types.NewMsgAddAccessRequest("hotdog", s.user1Addr, accessMintGrant),
			expectedEvent: types.NewEventMarkerAddAccess(&accessMintGrant, "hotdog", s.user1),
		},
		{
			name:     "should fail to ADD access to marker, validate basic fails",
			msg:      types.NewMsgAddAccessRequest("hotdog", s.user1Addr, accessInvalidGrant),
			errorMsg: "invalid access type: invalid request",
		},
		{

			name:     "should fail to ADD access to marker, keeper AddAccess failure",
			msg:      types.NewMsgAddAccessRequest("hotdog", s.user2Addr, accessMintGrant),
			errorMsg: fmt.Sprintf("updates to pending marker hotdog can only be made by %s: unauthorized", s.user1),
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
			msg:  types.NewMsgAddMarkerRequest(hotdogDenom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true, false, []string{}),
		},
		{
			name: "setup grant access to marker",
			msg:  types.NewMsgAddAccessRequest(hotdogDenom, s.user1Addr, accessMintGrant),
		},
		{
			name:          "should successfully delete grant access to marker",
			msg:           types.NewDeleteAccessRequest(hotdogDenom, s.user1Addr, s.user1Addr),
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
			msg:  types.NewMsgAddMarkerRequest(hotdogDenom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true, false, []string{}),
		},
		{
			name:          "should successfully finalize marker",
			msg:           types.NewMsgFinalizeRequest(hotdogDenom, s.user1Addr),
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
			msg:  types.NewMsgAddMarkerRequest(hotdogDenom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true, false, []string{}),
		},
		{
			name: "setup finalize marker",
			msg:  types.NewMsgFinalizeRequest(hotdogDenom, s.user1Addr),
		},
		{
			name:          "should successfully activate marker",
			msg:           types.NewMsgActivateRequest(hotdogDenom, s.user1Addr),
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
			msg:  types.NewMsgAddMarkerRequest(hotdogDenom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true, false, []string{}),
		},
		{
			name: "setup grant delete access to marker",
			msg:  types.NewMsgAddAccessRequest(hotdogDenom, s.user1Addr, accessDeleteGrant),
		},
		{
			name:          "should successfully cancel marker",
			msg:           types.NewMsgCancelRequest(hotdogDenom, s.user1Addr),
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
			msg:  types.NewMsgAddMarkerRequest(hotdogDenom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true, false, []string{}),
		},
		{
			name: "setup grant delete access to marker",
			msg:  types.NewMsgAddAccessRequest(hotdogDenom, s.user1Addr, accessDeleteMintGrant),
		},
		{
			name: "setup cancel marker",
			msg:  types.NewMsgCancelRequest(hotdogDenom, s.user1Addr),
		},
		{
			name:          "should successfully delete marker",
			msg:           types.NewMsgDeleteRequest(hotdogDenom, s.user1Addr),
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
			msg:  types.NewMsgAddMarkerRequest(hotdogDenom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true, false, []string{}),
		},
		{
			name: "setup grant mint access to marker",
			msg:  types.NewMsgAddAccessRequest(hotdogDenom, s.user1Addr, access),
		},
		{
			name:          "should successfully mint marker",
			msg:           types.NewMsgMintRequest(s.user1Addr, sdk.NewCoin(hotdogDenom, sdk.NewInt(100))),
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
			msg:  types.NewMsgAddMarkerRequest(hotdogDenom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true, false, []string{}),
		},
		{
			name: "setup grant mint access to marker",
			msg:  types.NewMsgAddAccessRequest(hotdogDenom, s.user1Addr, access),
		},
		{
			name:          "should successfully burn marker",
			msg:           types.NewMsgBurnRequest(s.user1Addr, sdk.NewCoin(hotdogDenom, sdk.NewInt(100))),
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
			msg:  types.NewMsgAddMarkerRequest(hotdogDenom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true, false, []string{}),
		},
		{
			name: "setup grant access to marker",
			msg:  types.NewMsgAddAccessRequest(hotdogDenom, s.user1Addr, access),
		},
		{
			name: "setup finalize marker",
			msg:  types.NewMsgFinalizeRequest(hotdogDenom, s.user1Addr),
		},
		{
			name: "setup activate marker",
			msg:  types.NewMsgActivateRequest(hotdogDenom, s.user1Addr),
		},
		{
			name:          "should successfully withdraw marker",
			msg:           types.NewMsgWithdrawRequest(s.user1Addr, s.user1Addr, hotdogDenom, sdk.NewCoins(sdk.NewCoin(hotdogDenom, sdk.NewInt(100)))),
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
			msg:  types.NewMsgAddMarkerRequest(hotdogDenom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_RestrictedCoin, true, true, false, []string{}),
		},
		{
			name: "setup grant access to marker",
			msg:  types.NewMsgAddAccessRequest(hotdogDenom, s.user1Addr, access),
		},
		{
			name: "setup finalize marker",
			msg:  types.NewMsgFinalizeRequest(hotdogDenom, s.user1Addr),
		},
		{
			name: "setup activate marker",
			msg:  types.NewMsgActivateRequest(hotdogDenom, s.user1Addr),
		},
		{
			name: "should successfully mint marker",
			msg:  types.NewMsgMintRequest(s.user1Addr, sdk.NewCoin(hotdogDenom, sdk.NewInt(1000))),
		},
		{
			name:          "should successfully transfer marker",
			msg:           types.NewMsgTransferRequest(s.user1Addr, s.user1Addr, s.user2Addr, sdk.NewCoin(hotdogDenom, sdk.NewInt(0))),
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
			msg:  types.NewMsgAddMarkerRequest(fmt.Sprintf("n%s", hotdogDenom), sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_RestrictedCoin, true, true, false, []string{}),
		},
		{
			name: "setup grant access to marker",
			msg:  types.NewMsgAddAccessRequest(fmt.Sprintf("n%s", hotdogDenom), s.user1Addr, access),
		},
		{
			name:          "should successfully set denom metadata on marker",
			msg:           types.NewSetDenomMetadataRequest(hotdogMetadata, s.user1Addr),
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
			name:          "should successfully ADD,FINALIZE,ACTIVATE new marker",
			msg:           msgWithActiveStatus,
			expectedEvent: types.NewEventMarkerAdd(denom, types.MustGetMarkerAddress(denom).String(), "100", "proposed", s.user1, types.MarkerType_Coin.String()),
		},
		{
			name:          "should successfully ADD,FINALIZE,ACTIVATE new marker with attributes",
			msg:           msgWithActiveStatusAttr,
			expectedEvent: types.NewEventMarkerAdd(rdenom, types.MustGetMarkerAddress(rdenom).String(), "100", "proposed", s.user1, types.MarkerType_RestrictedCoin.String()),
		},
		{
			name:     "should fail to ADD,FINALIZE,ACTIVATE new marker, validate basic failure",
			msg:      accessGrantWrongStatus,
			errorMsg: "since this will activate the marker, must have at least one access list defined: invalid request",
		},
		{
			name:     "should fail to ADD,FINALIZE,ACTIVATE new marker, marker already exists",
			msg:      types.NewMsgAddMarkerRequest(denom, sdk.NewInt(100), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true, false, []string{}),
			errorMsg: fmt.Sprintf("marker address already exists for %s: invalid request", types.MustGetMarkerAddress(denom)),
		},
		{
			name:          "should successfully add marker with dash and period",
			msg:           types.NewMsgAddMarkerRequest(denomWithDashPeriod, sdk.NewInt(1000), s.user1Addr, s.user1Addr, types.MarkerType_Coin, true, true, false, []string{}),
			expectedEvent: types.NewEventMarkerAdd(denomWithDashPeriod, types.MustGetMarkerAddress(denomWithDashPeriod).String(), "1000", "proposed", s.user1, types.MarkerType_Coin.String()),
		},
		{
			name:          "should successfully mint denom",
			msg:           types.NewMsgMintRequest(s.user1Addr, sdk.NewInt64Coin(denom, 1000)),
			expectedEvent: types.NewEventMarkerMint("1000", denom, s.user1),
		},
		{
			name:     "should fail to  burn denom, user doesn't have permissions",
			msg:      types.NewMsgBurnRequest(s.user1Addr, sdk.NewInt64Coin(denom, 50)),
			errorMsg: fmt.Sprintf("%s does not have ACCESS_BURN on hotdog markeraccount: invalid request", s.user1),
		},
	}
	s.runTests(cases)
}

func (s *HandlerTestSuite) TestMsgSetAccountDataRequest() {
	denomU := "aducoin"
	denomR := "adrcoin"

	denomUAddr := types.MustGetMarkerAddress(denomU).String()
	denomRAddr := types.MustGetMarkerAddress(denomR).String()

	authority := s.app.MarkerKeeper.GetAuthority()

	s.T().Logf("%s: %s", denomU, denomUAddr)
	s.T().Logf("%s: %s", denomR, denomRAddr)
	s.T().Logf("authority: %s", authority)

	tests := []CommonTest{
		{
			name: "should successfully add/finalize/active unrestricted marker",
			msg: types.NewMsgAddFinalizeActivateMarkerRequest(
				denomU, sdk.NewInt(100),
				s.user1Addr, s.user1Addr, // From and Manager.
				types.MarkerType_Coin,
				true,       // Supply fixed
				true,       // Allow gov
				false,      // don't allow forced transfer
				[]string{}, // No required attributes.
				[]types.AccessGrant{
					{Address: s.user1, Permissions: []types.Access{types.Access_Mint, types.Access_Admin}},
					{Address: s.user2, Permissions: []types.Access{types.Access_Deposit}},
				},
			),
		},
		{
			name: "should successfully add/finalize/active restricted marker",
			msg: types.NewMsgAddFinalizeActivateMarkerRequest(
				denomR, sdk.NewInt(100),
				s.user1Addr, s.user1Addr, // From and Manager.
				types.MarkerType_RestrictedCoin,
				true,       // Supply fixed
				true,       // Allow gov
				false,      // don't allow forced transfer
				[]string{}, // No required attributes.
				[]types.AccessGrant{
					{Address: s.user1, Permissions: []types.Access{types.Access_Mint, types.Access_Admin}},
					{Address: s.user2, Permissions: []types.Access{types.Access_Deposit}},
				},
			),
		},
		{
			name: "should successfully set account data on unrestricted marker via gov prop",
			msg: &types.MsgSetAccountDataRequest{
				Denom:  denomU,
				Value:  "This is some unrestricted coin data.",
				Signer: authority,
			},
			expectedEvent: &attrtypes.EventAccountDataUpdated{Account: denomUAddr},
		},
		{
			name: "should successfully set account data on unrestricted marker by signer with deposit",
			msg: &types.MsgSetAccountDataRequest{
				Denom:  denomU,
				Value:  "This is some different unrestricted coin data.",
				Signer: s.user2,
			},
			expectedEvent: &attrtypes.EventAccountDataUpdated{Account: denomUAddr},
		},
		{
			name: "should fail to set account data on unrestricted marker because signer does not have deposit",
			msg: &types.MsgSetAccountDataRequest{
				Denom:  denomU,
				Value:  "This is some unrestricted coin data. This won't get used though.",
				Signer: s.user1,
			},
			errorMsg: s.user1 + " does not have deposit access for " + denomU + " marker",
		},
		{
			name: "should successfully set account data on restricted marker via gov prop",
			msg: &types.MsgSetAccountDataRequest{
				Denom:  denomR,
				Value:  "This is some restricted coin data.",
				Signer: authority,
			},
			expectedEvent: &attrtypes.EventAccountDataUpdated{Account: denomRAddr},
		},
		{
			name: "should successfully set account data on restricted marker by signer with deposit",
			msg: &types.MsgSetAccountDataRequest{
				Denom:  denomR,
				Value:  "This is some different restricted coin data.",
				Signer: s.user2,
			},
			expectedEvent: &attrtypes.EventAccountDataUpdated{Account: denomRAddr},
		},
		{
			name: "should fail to set account data on restricted marker because signer does not have deposit",
			msg: &types.MsgSetAccountDataRequest{
				Denom:  denomR,
				Value:  "This is some restricted coin data. This won't get used though.",
				Signer: s.user1,
			},
			errorMsg: s.user1 + " does not have deposit access for " + denomR + " marker",
		},
	}
	s.runTests(tests)
}
