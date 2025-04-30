package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/gogoproto/proto"

	simapp "github.com/provenance-io/provenance/app"
	attrtypes "github.com/provenance-io/provenance/x/attribute/types"
	markerkeeper "github.com/provenance-io/provenance/x/marker/keeper"
	"github.com/provenance-io/provenance/x/marker/types"
)

type MsgServerTestSuite struct {
	suite.Suite

	app            *simapp.App
	ctx            sdk.Context
	msgServer      types.MsgServer
	blockStartTime time.Time

	privkey1   cryptotypes.PrivKey
	pubkey1    cryptotypes.PubKey
	owner1     string
	owner1Addr sdk.AccAddress
	acct1      sdk.AccountI

	privkey2   cryptotypes.PrivKey
	pubkey2    cryptotypes.PubKey
	owner2     string
	owner2Addr sdk.AccAddress
	acct2      sdk.AccountI

	addresses []sdk.AccAddress
}

func (s *MsgServerTestSuite) SetupTest() {

	s.blockStartTime = time.Now()
	s.app = simapp.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContextLegacy(false, cmtproto.Header{Time: s.blockStartTime})
	s.msgServer = markerkeeper.NewMsgServerImpl(s.app.MarkerKeeper)

	s.privkey1 = secp256k1.GenPrivKey()
	s.pubkey1 = s.privkey1.PubKey()
	s.owner1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.owner1 = s.owner1Addr.String()
	acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.owner1Addr)
	s.app.AccountKeeper.SetAccount(s.ctx, acc)

	s.privkey2 = secp256k1.GenPrivKey()
	s.pubkey2 = s.privkey2.PubKey()
	s.owner2Addr = sdk.AccAddress(s.pubkey2.Address())
	s.owner2 = s.owner2Addr.String()
	acc = s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.owner2Addr)
	s.app.AccountKeeper.SetAccount(s.ctx, acc)
}
func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerTestSuite))
}

func (s *MsgServerTestSuite) TestMsgAddMarkerRequest() {
	denom := "hotdog"
	rdenom := "restrictedhotdog"
	navDenom := "navdenom"
	denomWithDashPeriod := fmt.Sprintf("%s-my.marker", denom)

	cases := []struct {
		name     string
		msg      types.MsgAddMarkerRequest
		expErr   string
		expEvent []proto.Message
	}{
		{
			name: "successfully ADD new marker",
			msg: types.MsgAddMarkerRequest{
				Amount:                 sdk.NewInt64Coin(denom, 100),
				Manager:                s.owner1,
				FromAddress:            s.owner1,
				Status:                 types.StatusProposed,
				MarkerType:             types.MarkerType_Coin,
				SupplyFixed:            true,
				AllowGovernanceControl: true,
				AllowForcedTransfer:    false,
			},
			expEvent: []proto.Message{
				&types.EventMarkerAdd{
					Denom:      denom,
					Address:    types.MustGetMarkerAddress(denom).String(),
					Amount:     "100",
					Status:     "proposed",
					Manager:    s.owner1,
					MarkerType: types.MarkerType_Coin.String(),
				},
			},
		},
		{
			name: "fail to ADD new marker, invalid status",
			msg: types.MsgAddMarkerRequest{
				Amount:                 sdk.NewInt64Coin(denom, 100),
				Manager:                s.owner1,
				FromAddress:            s.owner1,
				Status:                 types.StatusActive,
				MarkerType:             types.MarkerType_Coin,
				SupplyFixed:            true,
				AllowGovernanceControl: true,
				AllowForcedTransfer:    false,
			},
			expErr: "marker can only be created with a Proposed or Finalized status",
		},
		{
			name: "fail to ADD new marker, marker already exists",
			msg: types.MsgAddMarkerRequest{
				Amount:                 sdk.NewInt64Coin(denom, 100),
				Manager:                s.owner1,
				FromAddress:            s.owner1,
				Status:                 types.StatusProposed,
				MarkerType:             types.MarkerType_Coin,
				SupplyFixed:            true,
				AllowGovernanceControl: true,
				AllowForcedTransfer:    false,
			},
			expErr: fmt.Sprintf("marker address already exists for %s: invalid request", types.MustGetMarkerAddress(denom)),
		},
		{
			name: "fail to ADD new marker, incorrect nav config",
			msg: types.MsgAddMarkerRequest{
				Amount:                 sdk.NewInt64Coin("jackthecat", 100),
				Manager:                s.owner1,
				FromAddress:            s.owner1,
				Status:                 types.StatusProposed,
				MarkerType:             types.MarkerType_Coin,
				SupplyFixed:            true,
				AllowGovernanceControl: true,
				AllowForcedTransfer:    false,
				UsdMills:               1,
				Volume:                 0,
			},
			expErr: `cannot set net asset value: marker net asset value volume must be positive value: invalid request`,
		},
		{
			name: "successfully Add marker with nav",
			msg: types.MsgAddMarkerRequest{
				Amount:                 sdk.NewInt64Coin(navDenom, 100),
				Manager:                s.owner1,
				FromAddress:            s.owner1,
				Status:                 types.StatusProposed,
				MarkerType:             types.MarkerType_Coin,
				SupplyFixed:            true,
				AllowGovernanceControl: true,
				AllowForcedTransfer:    false,
				UsdMills:               1,
				Volume:                 10,
			},
			expEvent: []proto.Message{
				&types.EventMarkerAdd{
					Denom:      navDenom,
					Address:    types.MustGetMarkerAddress(navDenom).String(),
					Amount:     "100",
					Status:     "proposed",
					Manager:    s.owner1,
					MarkerType: types.MarkerType_Coin.String(),
				},
				&types.EventSetNetAssetValue{
					Denom:  navDenom,
					Price:  "1usd",
					Volume: "10",
					Source: types.ModuleName,
				},
			},
		},
		{
			name: "successfully add marker with dash and period",
			msg: types.MsgAddMarkerRequest{
				Amount:                 sdk.NewInt64Coin(denomWithDashPeriod, 1000),
				Manager:                s.owner1,
				FromAddress:            s.owner1,
				Status:                 types.StatusProposed,
				MarkerType:             types.MarkerType_Coin,
				SupplyFixed:            true,
				AllowGovernanceControl: true,
				AllowForcedTransfer:    false,
			},
			expEvent: []proto.Message{
				&types.EventMarkerAdd{
					Denom:      denomWithDashPeriod,
					Address:    types.MustGetMarkerAddress(denomWithDashPeriod).String(),
					Amount:     "1000",
					Status:     "proposed",
					Manager:    s.owner1,
					MarkerType: types.MarkerType_Coin.String(),
				},
			},
		},
		{
			name: "successfully ADD new marker with required attributes",
			msg: types.MsgAddMarkerRequest{
				Amount:                 sdk.NewInt64Coin(rdenom, 100),
				Manager:                s.owner1,
				FromAddress:            s.owner1,
				Status:                 types.StatusProposed,
				MarkerType:             types.MarkerType_RestrictedCoin,
				SupplyFixed:            true,
				AllowGovernanceControl: true,
				AllowForcedTransfer:    false,
				RequiredAttributes:     []string{"attribute.one.com", "attribute.two.com"},
			},
			expEvent: []proto.Message{
				&types.EventMarkerAdd{
					Denom:      rdenom,
					Address:    types.MustGetMarkerAddress(rdenom).String(),
					Amount:     "100",
					Status:     "proposed",
					Manager:    s.owner1,
					MarkerType: types.MarkerType_RestrictedCoin.String(),
				},
			},
		},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
			res, err := s.msgServer.AddMarker(s.ctx, &tc.msg)
			if len(tc.expErr) > 0 {
				s.Require().EqualError(err, tc.expErr, "AddMarker(%v) error", tc.msg)
			} else {
				events := s.ctx.EventManager().ABCIEvents()
				s.Require().NoError(err, "AddMarker(%v) error", tc.msg)
				s.Assert().Equal(res, &types.MsgAddMarkerResponse{})
				for _, expEvent := range tc.expEvent {
					s.Assert().True(s.containsMessage(events, expEvent), "AddMarker missing expected event %T", expEvent)
				}
			}
		})
	}
}

func (s *MsgServerTestSuite) containsMessage(events []abci.Event, msg proto.Message) bool {
	for _, event := range events {
		typeEvent, _ := sdk.ParseTypedEvent(event)
		if assert.ObjectsAreEqual(msg, typeEvent) {
			return true
		}
	}
	return false
}

// noAccessErr creates an expected error message for an address not having access on a marker.
func (s *MsgServerTestSuite) noAccessErr(addr string, role types.Access, denom string) string {
	mAddr, err := types.MarkerAddress(denom)
	s.Require().NoError(err, "MarkerAddress(%q)", denom)
	return fmt.Sprintf("%s does not have %s on %s marker (%s)", addr, role, denom, mAddr)
}

func (s *MsgServerTestSuite) TestMsgFinalizeMarkerRequest() {
	authUser := testUserAddress("test")
	noNavMarker := types.NewEmptyMarkerAccount(
		"nonav",
		authUser.String(),
		[]types.AccessGrant{})

	s.Require().NoError(s.app.MarkerKeeper.AddMarkerAccount(s.ctx, noNavMarker))

	validMarker := types.NewEmptyMarkerAccount(
		"hotdog",
		authUser.String(),
		[]types.AccessGrant{
			{Address: authUser.String(), Permissions: types.AccessList{types.Access_Admin, types.Access_Mint}},
		},
	)
	validMarker.Supply = sdkmath.NewInt(1)
	s.Require().NoError(s.app.MarkerKeeper.AddMarkerAccount(s.ctx, validMarker))
	s.Require().NoError(s.app.MarkerKeeper.SetNetAssetValue(s.ctx, validMarker, types.NetAssetValue{Price: sdk.NewInt64Coin(types.UsdDenom, 1), Volume: 1}, "test"))

	testCases := []struct {
		name   string
		msg    types.MsgFinalizeRequest
		expErr string
	}{
		{
			name: "successfully finalize",
			msg:  types.MsgFinalizeRequest{Denom: validMarker.Denom, Administrator: authUser.String()},
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			res, err := s.msgServer.Finalize(s.ctx,
				&tc.msg)

			if len(tc.expErr) > 0 {
				s.Assert().Nil(res)
				s.Assert().EqualError(err, tc.expErr)

			} else {
				s.Assert().NoError(err)
				s.Assert().Equal(res, &types.MsgFinalizeResponse{})
			}
		})
	}
}

func (s *MsgServerTestSuite) TestUpdateForcedTransfer() {
	authority := s.app.MarkerKeeper.GetAuthority()
	otherAddr := sdk.AccAddress("otherAccAddr________").String()

	proposed := types.StatusProposed
	active := types.StatusActive
	finalized := types.StatusFinalized

	newMarker := func(denom string, status types.MarkerStatus, allowForcedTransfer bool) *types.MarkerAccount {
		rv := &types.MarkerAccount{
			BaseAccount: authtypes.NewBaseAccountWithAddress(types.MustGetMarkerAddress(denom)),
			AccessControl: []types.AccessGrant{
				{
					Address: sdk.AccAddress("allAccessAddr_______").String(),
					Permissions: types.AccessList{
						types.Access_Mint, types.Access_Burn,
						types.Access_Deposit, types.Access_Withdraw,
						types.Access_Delete, types.Access_Admin, types.Access_Transfer,
					},
				},
			},
			Status:                 status,
			Denom:                  denom,
			Supply:                 sdkmath.NewInt(1000),
			MarkerType:             types.MarkerType_RestrictedCoin,
			AllowGovernanceControl: true,
			AllowForcedTransfer:    allowForcedTransfer,
		}
		s.app.AccountKeeper.NewAccount(s.ctx, rv.BaseAccount)
		return rv
	}
	newUnMarker := func(denom string) *types.MarkerAccount {
		rv := newMarker(denom, active, false)
		rv.AccessControl = nil
		rv.MarkerType = types.MarkerType_Coin
		return rv
	}
	newNoGovMarker := func(denom string) *types.MarkerAccount {
		rv := newMarker(denom, active, false)
		rv.AllowGovernanceControl = false
		return rv
	}
	newMsg := func(denom string, allowForcedTransfer bool) *types.MsgUpdateForcedTransferRequest {
		return &types.MsgUpdateForcedTransferRequest{
			Denom:               denom,
			AllowForcedTransfer: allowForcedTransfer,
			Authority:           authority,
		}
	}
	markerAddr := func(denom string) string {
		return types.MustGetMarkerAddress(denom).String()
	}

	tests := []struct {
		name       string
		origMarker types.MarkerAccountI
		msg        *types.MsgUpdateForcedTransferRequest
		expErr     string
	}{
		{
			name: "wrong authority",
			msg: &types.MsgUpdateForcedTransferRequest{
				Denom:               "somedenom",
				AllowForcedTransfer: false,
				Authority:           otherAddr,
			},
			expErr: "expected " + authority + " got " + otherAddr + ": expected gov account as only signer for proposal message",
		},
		{
			name:   "marker does not exist",
			msg:    newMsg("nosuchmarker", false),
			expErr: "could not get marker for nosuchmarker: marker nosuchmarker not found for address: " + markerAddr("nosuchmarker"),
		},
		{
			name:       "unrestricted coin",
			origMarker: newUnMarker("unrestrictedcoin"),
			msg:        newMsg("unrestrictedcoin", true),
			expErr:     "cannot update forced transfer on unrestricted marker unrestrictedcoin",
		},
		{
			name:       "gov not enabled",
			origMarker: newNoGovMarker("nogovallowed"),
			msg:        newMsg("nogovallowed", true),
			expErr:     "nogovallowed marker does not allow governance control",
		},
		{
			name:       "false not changing",
			origMarker: newMarker("activefalse", active, false),
			msg:        newMsg("activefalse", false),
			expErr:     "marker activefalse already has allow_forced_transfer = false",
		},
		{
			name:       "true not changing",
			origMarker: newMarker("activetrue", active, true),
			msg:        newMsg("activetrue", true),
			expErr:     "marker activetrue already has allow_forced_transfer = true",
		},
		{
			name:       "active true to false",
			origMarker: newMarker("activetf", active, true),
			msg:        newMsg("activetf", false),
			expErr:     "",
		},
		{
			name:       "active false to true",
			origMarker: newMarker("activeft", active, false),
			msg:        newMsg("activeft", true),
			expErr:     "",
		},
		{
			name:       "proposed true to false",
			origMarker: newMarker("proposedtf", proposed, true),
			msg:        newMsg("proposedtf", false),
			expErr:     "",
		},
		{
			name:       "proposed false to true",
			origMarker: newMarker("proposedft", proposed, false),
			msg:        newMsg("proposedft", true),
			expErr:     "",
		},
		{
			name:       "finalized true to false",
			origMarker: newMarker("finalizedtf", finalized, true),
			msg:        newMsg("finalizedtf", false),
			expErr:     "",
		},
		{
			name:       "finalized false to true",
			origMarker: newMarker("finalizedft", finalized, false),
			msg:        newMsg("finalizedft", true),
			expErr:     "",
		},
	}

	markerLastSet := make(map[string]string)
	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.origMarker != nil {
				denom := tc.origMarker.GetDenom()
				if len(markerLastSet[denom]) > 0 {
					s.T().Logf("WARNING: overwriting %q marker previously defined in test %q.", denom, markerLastSet[denom])
				}
				markerLastSet[denom] = tc.name
				s.app.MarkerKeeper.SetMarker(s.ctx, tc.origMarker)
			}

			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)
			var res *types.MsgUpdateForcedTransferResponse
			var err error
			testFunc := func() {
				res, err = s.msgServer.UpdateForcedTransfer(ctx, tc.msg)
			}

			s.Require().NotPanics(testFunc, "UpdateForcedTransfer")
			if len(tc.expErr) > 0 {
				s.Assert().EqualError(err, tc.expErr, "UpdateForcedTransfer error")
				s.Assert().Nil(res, "UpdateForcedTransfer response")

				events := em.Events()
				s.Assert().Empty(events, "events emitted during failed UpdateForcedTransfer")
			} else {
				s.Require().NoError(err, "UpdateForcedTransfer error")
				s.Assert().Equal(res, &types.MsgUpdateForcedTransferResponse{}, "UpdateForcedTransfer response")

				markerNow, err := s.app.MarkerKeeper.GetMarkerByDenom(s.ctx, tc.msg.Denom)
				if s.Assert().NoError(err, "GetMarkerByDenom(%q)", tc.msg.Denom) {
					allowsForcedTransfer := markerNow.AllowsForcedTransfer()
					s.Assert().Equal(tc.msg.AllowForcedTransfer, allowsForcedTransfer, "AllowsForcedTransfer after UpdateForcedTransfer")
				}
			}
		})
	}
}

func (s *MsgServerTestSuite) TestUpdateSendDenyList() {
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)

	authUser := testUserAddress("test")
	notAuthUser := testUserAddress("test1")

	notRestrictedMarker := types.NewEmptyMarkerAccount(
		"not-restricted-marker",
		authUser.String(),
		[]types.AccessGrant{})

	err := s.app.MarkerKeeper.AddMarkerAccount(s.ctx, notRestrictedMarker)
	s.Require().NoError(err)

	rMarkerDenom := "restricted-marker"
	rMarkerAcct := authtypes.NewBaseAccount(types.MustGetMarkerAddress(rMarkerDenom), nil, 0, 0)
	s.app.MarkerKeeper.SetNewMarker(s.ctx, types.NewMarkerAccount(rMarkerAcct, sdk.NewInt64Coin(rMarkerDenom, 1000), authUser, []types.AccessGrant{{Address: authUser.String(), Permissions: []types.Access{types.Access_Transfer}}}, types.StatusFinalized, types.MarkerType_RestrictedCoin, true, false, false, []string{}))

	rMarkerGovDenom := "restricted-marker-gov"
	rMarkerGovAcct := authtypes.NewBaseAccount(types.MustGetMarkerAddress(rMarkerGovDenom), nil, 0, 0)
	s.app.MarkerKeeper.SetNewMarker(s.ctx, types.NewMarkerAccount(rMarkerGovAcct, sdk.NewInt64Coin(rMarkerGovDenom, 1000), authUser, []types.AccessGrant{{Address: authUser.String(), Permissions: []types.Access{}}}, types.StatusFinalized, types.MarkerType_RestrictedCoin, true, true, false, []string{}))

	denyAddrToRemove := testUserAddress("denyAddrToRemove")
	s.app.MarkerKeeper.AddSendDeny(s.ctx, rMarkerAcct.GetAddress(), denyAddrToRemove)
	s.Require().True(s.app.MarkerKeeper.IsSendDeny(s.ctx, rMarkerAcct.GetAddress(), denyAddrToRemove), rMarkerDenom+" should have added address to deny list "+denyAddrToRemove.String())

	denyAddrToAdd := testUserAddress("denyAddrToAdd")

	denyAddrToAddGov := testUserAddress("denyAddrToAddGov")

	testCases := []struct {
		name   string
		msg    types.MsgUpdateSendDenyListRequest
		expErr string
	}{
		{
			name:   "should fail, cannot find marker",
			msg:    types.MsgUpdateSendDenyListRequest{Denom: "blah", Authority: authUser.String(), RemoveDeniedAddresses: []string{}, AddDeniedAddresses: []string{}},
			expErr: "marker not found for blah: marker blah not found for address: cosmos1psw3a97ywtr595qa4295lw07cz9665hynnfpee",
		},
		{
			name:   "should fail, not a restricted marker",
			msg:    types.MsgUpdateSendDenyListRequest{Denom: notRestrictedMarker.Denom, Authority: authUser.String(), RemoveDeniedAddresses: []string{}, AddDeniedAddresses: []string{}},
			expErr: "marker not-restricted-marker is not a restricted marker",
		},
		{
			name:   "should fail, signer does not have admin access",
			msg:    types.MsgUpdateSendDenyListRequest{Denom: rMarkerDenom, Authority: notAuthUser.String(), RemoveDeniedAddresses: []string{}, AddDeniedAddresses: []string{}},
			expErr: fmt.Sprintf("%s does not have %s on %s marker (%s)", notAuthUser, types.Access_Transfer, rMarkerDenom, rMarkerAcct.Address),
		},
		{
			name:   "should fail, gov not enabled for restricted marker",
			msg:    types.MsgUpdateSendDenyListRequest{Denom: rMarkerDenom, Authority: authority.String(), RemoveDeniedAddresses: []string{}, AddDeniedAddresses: []string{}},
			expErr: "restricted-marker marker does not allow governance control",
		},
		{
			name:   "should fail, address is already on deny list",
			msg:    types.MsgUpdateSendDenyListRequest{Denom: rMarkerDenom, Authority: authUser.String(), RemoveDeniedAddresses: []string{}, AddDeniedAddresses: []string{denyAddrToRemove.String()}},
			expErr: denyAddrToRemove.String() + " is already on deny list cannot add address",
		},
		{
			name:   "should fail, address can not be removed not in deny list",
			msg:    types.MsgUpdateSendDenyListRequest{Denom: rMarkerDenom, Authority: authUser.String(), RemoveDeniedAddresses: []string{denyAddrToAdd.String()}, AddDeniedAddresses: []string{}},
			expErr: denyAddrToAdd.String() + " is not on deny list cannot remove address",
		},
		{
			name:   "should fail, invalid address on add list",
			msg:    types.MsgUpdateSendDenyListRequest{Denom: rMarkerDenom, Authority: authUser.String(), RemoveDeniedAddresses: []string{}, AddDeniedAddresses: []string{"invalid-add-address"}},
			expErr: "decoding bech32 failed: invalid separator index -1",
		},
		{
			name:   "should fail, invalid address on remove list",
			msg:    types.MsgUpdateSendDenyListRequest{Denom: rMarkerDenom, Authority: authUser.String(), RemoveDeniedAddresses: []string{"invalid-remove-address"}, AddDeniedAddresses: []string{}},
			expErr: "decoding bech32 failed: invalid separator index -1",
		},
		{
			name: "should succeed to add to deny list",
			msg:  types.MsgUpdateSendDenyListRequest{Denom: rMarkerDenom, Authority: authUser.String(), RemoveDeniedAddresses: []string{}, AddDeniedAddresses: []string{denyAddrToAdd.String()}},
		},
		{
			name: "should succeed to remove from deny list",
			msg:  types.MsgUpdateSendDenyListRequest{Denom: rMarkerDenom, Authority: authUser.String(), RemoveDeniedAddresses: []string{denyAddrToRemove.String()}, AddDeniedAddresses: []string{}},
		},
		{
			name: "should succeed gov allowed for marker",
			msg:  types.MsgUpdateSendDenyListRequest{Denom: rMarkerGovDenom, Authority: authority.String(), RemoveDeniedAddresses: []string{}, AddDeniedAddresses: []string{denyAddrToAddGov.String()}},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			res, err := s.msgServer.UpdateSendDenyList(s.ctx,
				&tc.msg)

			if len(tc.expErr) > 0 {
				s.Assert().Nil(res)
				s.Assert().EqualError(err, tc.expErr)

			} else {
				s.Assert().NoError(err)
				s.Assert().Equal(res, &types.MsgUpdateSendDenyListResponse{})
			}
		})
	}
}

func (s *MsgServerTestSuite) TestAddNetAssetValue() {
	authUser := testUserAddress("test")
	notAuthUser := testUserAddress("blah")

	markerDenom := "jackthecat"
	markerAcct := authtypes.NewBaseAccount(types.MustGetMarkerAddress(markerDenom), nil, 0, 0)
	s.app.MarkerKeeper.SetNewMarker(s.ctx, types.NewMarkerAccount(markerAcct, sdk.NewInt64Coin(markerDenom, 1000), authUser, []types.AccessGrant{{Address: authUser.String(), Permissions: []types.Access{types.Access_Transfer}}}, types.StatusProposed, types.MarkerType_RestrictedCoin, true, false, false, []string{}))

	valueAcct := authtypes.NewBaseAccount(types.MustGetMarkerAddress(types.UsdDenom), nil, 0, 0)
	s.app.MarkerKeeper.SetNewMarker(s.ctx, types.NewMarkerAccount(valueAcct, sdk.NewInt64Coin(types.UsdDenom, 1000), authUser, []types.AccessGrant{{Address: authUser.String(), Permissions: []types.Access{types.Access_Transfer}}}, types.StatusProposed, types.MarkerType_RestrictedCoin, true, false, false, []string{}))

	finalizedMarkerDenom := "finalizedjackthecat"
	finalizedMarkerAcct := authtypes.NewBaseAccount(types.MustGetMarkerAddress(finalizedMarkerDenom), nil, 1, 0)
	s.app.MarkerKeeper.SetNewMarker(s.ctx, types.NewMarkerAccount(finalizedMarkerAcct, sdk.NewInt64Coin(finalizedMarkerDenom, 1000), authUser, []types.AccessGrant{{Address: authUser.String(), Permissions: []types.Access{types.Access_Transfer}}}, types.StatusFinalized, types.MarkerType_RestrictedCoin, true, false, false, []string{}))

	testCases := []struct {
		name   string
		msg    types.MsgAddNetAssetValuesRequest
		expErr string
	}{
		{
			name: "no marker found",
			msg: types.MsgAddNetAssetValuesRequest{
				Denom: "cantfindme",
				NetAssetValues: []types.NetAssetValue{
					{
						Price:  sdk.NewInt64Coin("navcoin", 1),
						Volume: 1,
					}},
				Administrator: authUser.String()},
			expErr: "marker cantfindme not found for address: cosmos17l2yneua2mdfqaycgyhqag8t20asnjwf6adpmt: invalid request",
		},
		{
			name: "nav denom matches marker denom",
			msg: types.MsgAddNetAssetValuesRequest{
				Denom: markerDenom,
				NetAssetValues: []types.NetAssetValue{
					{
						Price:              sdk.NewInt64Coin(markerDenom, 100),
						Volume:             uint64(100),
						UpdatedBlockHeight: 1,
					},
				},
				Administrator: authUser.String(),
			},
			expErr: `net asset value denom cannot match marker denom "jackthecat": invalid request`,
		},
		{
			name: "value denom does not exist",
			msg: types.MsgAddNetAssetValuesRequest{
				Denom: markerDenom,
				NetAssetValues: []types.NetAssetValue{
					{
						Price:              sdk.NewInt64Coin("hotdog", 100),
						Volume:             uint64(100),
						UpdatedBlockHeight: 1,
					},
				},
				Administrator: authUser.String(),
			},
			expErr: `net asset value denom does not exist: marker hotdog not found for address: cosmos1p6l3annxy35gm5mfm6m0jz2mdj8peheuzf9alh: invalid request`,
		},
		{
			name: "not authorize user",
			msg: types.MsgAddNetAssetValuesRequest{
				Denom: markerDenom,
				NetAssetValues: []types.NetAssetValue{
					{
						Price:              sdk.NewInt64Coin(types.UsdDenom, 100),
						Volume:             uint64(100),
						UpdatedBlockHeight: 1,
					},
				},
				Administrator: notAuthUser.String(),
			},
			expErr: `signer cosmos1psw3a97ywtr595qa4295lw07cz9665hynnfpee does not have permission to add net asset value for "jackthecat"`,
		},
		{
			name: "successfully set nav",
			msg: types.MsgAddNetAssetValuesRequest{
				Denom: markerDenom,
				NetAssetValues: []types.NetAssetValue{
					{
						Price:              sdk.NewInt64Coin(types.UsdDenom, 100),
						Volume:             uint64(100),
						UpdatedBlockHeight: 1,
					},
				},
				Administrator: authUser.String(),
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			res, err := s.msgServer.AddNetAssetValues(s.ctx,
				&tc.msg)

			if len(tc.expErr) > 0 {
				s.Assert().Nil(res)
				s.Assert().EqualError(err, tc.expErr)

			} else {
				s.Assert().NoError(err)
				s.Assert().Equal(res, &types.MsgAddNetAssetValuesResponse{})
			}
		})
	}
}

func (s *MsgServerTestSuite) TestMsgAddAccessRequest() {
	accessMintGrant := types.AccessGrant{
		Address:     s.owner1,
		Permissions: types.AccessListByNames("MINT"),
	}

	accessInvalidGrant := types.AccessGrant{
		Address:     s.owner1,
		Permissions: types.AccessListByNames("Invalid"),
	}

	addMarkerMsg := types.NewMsgAddMarkerRequest("hotdog", sdkmath.NewInt(100), s.owner1Addr, s.owner1Addr, types.MarkerType_Coin, true, true, false, []string{}, 0, 0)
	_, err := s.msgServer.AddMarker(s.ctx, addMarkerMsg)
	s.Assert().NoError(err, "should successfully add marker")

	testCases := []struct {
		name          string
		msg           *types.MsgAddAccessRequest
		errorMsg      string
		expectedEvent proto.Message
	}{
		{
			name:          "should successfully grant access to marker",
			msg:           types.NewMsgAddAccessRequest("hotdog", s.owner1Addr, accessMintGrant),
			expectedEvent: types.NewEventMarkerAddAccess(&accessMintGrant, "hotdog", s.owner1),
		},
		{
			name:     "should fail to ADD access to marker, validate basic fails",
			msg:      types.NewMsgAddAccessRequest("hotdog", s.owner1Addr, accessInvalidGrant),
			errorMsg: "invalid access type: invalid request",
		},
		{

			name:     "should fail to ADD access to marker, keeper AddAccess failure",
			msg:      types.NewMsgAddAccessRequest("hotdog", s.owner2Addr, accessMintGrant),
			errorMsg: fmt.Sprintf("updates to pending marker hotdog can only be made by %s: unauthorized", s.owner1),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
			response, err := s.msgServer.AddAccess(s.ctx, tc.msg)
			if len(tc.errorMsg) > 0 {
				s.Require().EqualError(err, tc.errorMsg, "handler(%T) error", tc.msg)
			} else {
				s.Require().NoError(err, "handler(%T) error", tc.msg)
				if tc.expectedEvent != nil {
					result := s.containsMessage(s.ctx.EventManager().ABCIEvents(), tc.expectedEvent)
					s.Assert().True(result, "Expected typed event was not found in response.\n    Expected: %+v\n    Response: %+v", tc.expectedEvent, response)
				}
			}
		})
	}
}

func (s *MsgServerTestSuite) TestMsgDeleteAccessMarkerRequest() {
	hotdogDenom := "hotdog"
	accessMintGrant := types.AccessGrant{
		Address:     s.owner1,
		Permissions: types.AccessListByNames("MINT"),
	}

	addMarkerMsg := types.NewMsgAddMarkerRequest(hotdogDenom, sdkmath.NewInt(100), s.owner1Addr, s.owner1Addr, types.MarkerType_Coin, true, true, false, []string{}, 0, 0)
	_, err := s.msgServer.AddMarker(s.ctx, addMarkerMsg)
	s.Assert().NoError(err, "should successfully add marker")

	addAccessMsg := types.NewMsgAddAccessRequest(hotdogDenom, s.owner1Addr, accessMintGrant)
	_, err = s.msgServer.AddAccess(s.ctx, addAccessMsg)
	s.Assert().NoError(err, "should add access to newly added marker")

	testcases := []struct {
		name          string
		msg           *types.MsgDeleteAccessRequest
		expectedEvent proto.Message
	}{
		{
			name:          "should successfully delete grant access to marker",
			msg:           types.NewDeleteAccessRequest(hotdogDenom, s.owner1Addr, s.owner1Addr),
			expectedEvent: types.NewEventMarkerDeleteAccess(s.owner1, hotdogDenom, s.owner1),
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
			response, err := s.msgServer.DeleteAccess(s.ctx, tc.msg)
			s.Require().NoError(err, "handler(%T) error", tc.msg)
			if tc.expectedEvent != nil {
				result := s.containsMessage(s.ctx.EventManager().ABCIEvents(), tc.expectedEvent)
				s.Assert().True(result, "Expected typed event was not found in response.\n    Expected: %+v\n    Response: %+v", tc.expectedEvent, response)
			}
		})
	}
}

func (s *MsgServerTestSuite) TestMsgActivateMarkerRequest() {
	hotdogDenom := "hotdog"

	addMarkerMsg := types.NewMsgAddMarkerRequest(hotdogDenom, sdkmath.NewInt(100), s.owner1Addr, s.owner1Addr, types.MarkerType_Coin, true, true, false, []string{}, 0, 0)
	_, err := s.msgServer.AddMarker(s.ctx, addMarkerMsg)
	s.Assert().NoError(err, "should successfully add marker")

	finalizeMarkerMsg := types.NewMsgFinalizeRequest(hotdogDenom, s.owner1Addr)
	_, err = s.msgServer.Finalize(s.ctx, finalizeMarkerMsg)
	s.Assert().NoError(err, "should not throw error when finalizing request")

	testcases := []struct {
		name          string
		msg           *types.MsgActivateRequest
		expectedEvent proto.Message
	}{
		{
			name:          "should successfully activate marker",
			msg:           types.NewMsgActivateRequest(hotdogDenom, s.owner1Addr),
			expectedEvent: types.NewEventMarkerActivate(hotdogDenom, s.owner1),
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
			response, err := s.msgServer.Activate(s.ctx, tc.msg)
			s.Require().NoError(err, "handler(%T) error", tc.msg)
			if tc.expectedEvent != nil {
				result := s.containsMessage(s.ctx.EventManager().ABCIEvents(), tc.expectedEvent)
				s.Assert().True(result, "Expected typed event was not found in response.\n    Expected: %+v\n    Response: %+v", tc.expectedEvent, response)
			}
		})
	}
}

func (s *MsgServerTestSuite) TestMsgCancelMarkerRequest() {
	hotdogDenom := "hotdog"
	accessDeleteGrant := types.AccessGrant{
		Address:     s.owner1,
		Permissions: types.AccessListByNames("DELETE"),
	}

	addMarkerMsg := types.NewMsgAddMarkerRequest(hotdogDenom, sdkmath.NewInt(100), s.owner1Addr, s.owner1Addr, types.MarkerType_Coin, true, true, false, []string{}, 0, 0)
	_, err := s.msgServer.AddMarker(s.ctx, addMarkerMsg)
	s.Assert().NoError(err, "should successfully add marker")

	addAccessMsg := types.NewMsgAddAccessRequest(hotdogDenom, s.owner1Addr, accessDeleteGrant)
	_, err = s.msgServer.AddAccess(s.ctx, addAccessMsg)
	s.Assert().NoError(err, "should not throw error when adding access to marker")

	testcases := []struct {
		name          string
		msg           *types.MsgCancelRequest
		expectedEvent proto.Message
	}{
		{
			name:          "should successfully cancel marker",
			msg:           types.NewMsgCancelRequest(hotdogDenom, s.owner1Addr),
			expectedEvent: types.NewEventMarkerCancel(hotdogDenom, s.owner1),
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
			response, err := s.msgServer.Cancel(s.ctx, tc.msg)
			s.Require().NoError(err, "handler(%T) error", tc.msg)
			if tc.expectedEvent != nil {
				result := s.containsMessage(s.ctx.EventManager().ABCIEvents(), tc.expectedEvent)
				s.Assert().True(result, "Expected typed event was not found in response.\n    Expected: %+v\n    Response: %+v", tc.expectedEvent, response)
			}
		})
	}
}

func (s *MsgServerTestSuite) TestMsgDeleteMarkerRequest() {
	hotdogDenom := "hotdog"
	accessDeleteMintGrant := types.AccessGrant{
		Address:     s.owner1,
		Permissions: types.AccessListByNames("DELETE,MINT"),
	}

	addMarkerMsg := types.NewMsgAddMarkerRequest(hotdogDenom, sdkmath.NewInt(100), s.owner1Addr, s.owner1Addr, types.MarkerType_Coin, true, true, false, []string{}, 0, 0)
	_, err := s.msgServer.AddMarker(s.ctx, addMarkerMsg)
	s.Assert().NoError(err, "should successfully add marker")

	addAccessMsg := types.NewMsgAddAccessRequest(hotdogDenom, s.owner1Addr, accessDeleteMintGrant)
	_, err = s.msgServer.AddAccess(s.ctx, addAccessMsg)
	s.Assert().NoError(err, "should not throw error when adding access to marker")

	cancelMsg := types.NewMsgCancelRequest(hotdogDenom, s.owner1Addr)
	_, err = s.msgServer.Cancel(s.ctx, cancelMsg)
	s.Assert().NoError(err, "should not throw error when canceling marker")

	testcases := []struct {
		name          string
		msg           *types.MsgDeleteRequest
		expectedEvent proto.Message
	}{
		{
			name:          "should successfully delete marker",
			msg:           types.NewMsgDeleteRequest(hotdogDenom, s.owner1Addr),
			expectedEvent: types.NewEventMarkerDelete(hotdogDenom, s.owner1),
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
			response, err := s.msgServer.Delete(s.ctx, tc.msg)
			s.Require().NoError(err, "handler(%T) error", tc.msg)
			if tc.expectedEvent != nil {
				result := s.containsMessage(s.ctx.EventManager().ABCIEvents(), tc.expectedEvent)
				s.Assert().True(result, "Expected typed event was not found in response.\n    Expected: %+v\n    Response: %+v", tc.expectedEvent, response)
			}
		})
	}
}

func (s *MsgServerTestSuite) TestMsgMintMarkerRequest() {
	hotdogDenom := "hotdog"
	access := types.AccessGrant{
		Address:     s.owner1,
		Permissions: types.AccessListByNames("MINT,BURN,WITHDRAW"),
	}
	addMarkerMsg := types.NewMsgAddMarkerRequest(hotdogDenom, sdkmath.NewInt(100), s.owner1Addr, s.owner1Addr, types.MarkerType_Coin, true, true, false, []string{}, 0, 0)
	_, err := s.msgServer.AddMarker(s.ctx, addMarkerMsg)
	s.Assert().NoError(err, "should successfully add marker")

	addAccessMsg := types.NewMsgAddAccessRequest(hotdogDenom, s.owner1Addr, access)
	_, err = s.msgServer.AddAccess(s.ctx, addAccessMsg)
	s.Assert().NoError(err, "should not throw error when adding access to marker")

	updateStatusProposal := types.NewMsgChangeStatusProposalRequest(hotdogDenom, types.StatusActive, s.app.MarkerKeeper.GetAuthority())
	_, err = s.msgServer.ChangeStatusProposal(s.ctx, updateStatusProposal)
	s.Assert().NoError(err, "should not throw error when status proposal changed")

	testcases := []struct {
		name          string
		msg           *types.MsgMintRequest
		expectedEvent proto.Message
	}{
		{
			name:          "should successfully mint marker",
			msg:           types.NewMsgMintRequest(s.owner1Addr, sdk.NewInt64Coin(hotdogDenom, 100), s.owner2Addr),
			expectedEvent: types.NewEventMarkerMint("100", hotdogDenom, s.owner1),
		},
		{
			name:          "should successfully mint marker; recipient is empty, withdrawal skipped",
			msg:           types.NewMsgMintRequest(s.owner1Addr, sdk.NewInt64Coin(hotdogDenom, 100), sdk.AccAddress{}),
			expectedEvent: types.NewEventMarkerMint("100", hotdogDenom, s.owner1),
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
			response, err := s.msgServer.Mint(s.ctx, tc.msg)
			s.Require().NoError(err, "handler(%T) error", tc.msg)
			if tc.expectedEvent != nil {
				result := s.containsMessage(s.ctx.EventManager().ABCIEvents(), tc.expectedEvent)
				s.Assert().True(result, "Expected typed event was not found in response.\n    Expected: %+v\n    Response: %+v", tc.expectedEvent, response)
			}
		})
	}
}

func (s *MsgServerTestSuite) TestMsgBurnMarkerRequest() {
	hotdogDenom := "hotdog"
	access := types.AccessGrant{
		Address:     s.owner1,
		Permissions: types.AccessListByNames("DELETE,MINT,BURN"),
	}

	addMarkerMsg := types.NewMsgAddMarkerRequest(hotdogDenom, sdkmath.NewInt(100), s.owner1Addr, s.owner1Addr, types.MarkerType_Coin, true, true, false, []string{}, 0, 0)
	_, err := s.msgServer.AddMarker(s.ctx, addMarkerMsg)
	s.Assert().NoError(err, "should successfully add marker")

	addAccessMsg := types.NewMsgAddAccessRequest(hotdogDenom, s.owner1Addr, access)
	_, err = s.msgServer.AddAccess(s.ctx, addAccessMsg)
	s.Assert().NoError(err, "should not throw error when adding access to marker")

	testcases := []struct {
		name          string
		msg           *types.MsgBurnRequest
		expectedEvent proto.Message
	}{
		{
			name:          "should successfully burn marker",
			msg:           types.NewMsgBurnRequest(s.owner1Addr, sdk.NewInt64Coin(hotdogDenom, 100)),
			expectedEvent: types.NewEventMarkerBurn("100", hotdogDenom, s.owner1),
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
			response, err := s.msgServer.Burn(s.ctx, tc.msg)
			s.Require().NoError(err, "handler(%T) error", tc.msg)
			if tc.expectedEvent != nil {
				result := s.containsMessage(s.ctx.EventManager().ABCIEvents(), tc.expectedEvent)
				s.Assert().True(result, "Expected typed event was not found in response.\n    Expected: %+v\n    Response: %+v", tc.expectedEvent, response)
			}
		})
	}
}

func (s *MsgServerTestSuite) TestMsgWithdrawMarkerRequest() {
	hotdogDenom := "hotdog"
	access := types.AccessGrant{
		Address:     s.owner1,
		Permissions: types.AccessListByNames("DELETE,MINT,WITHDRAW"),
	}

	addMarkerMsg := types.NewMsgAddMarkerRequest(hotdogDenom, sdkmath.NewInt(100), s.owner1Addr, s.owner1Addr, types.MarkerType_RestrictedCoin, true, true, false, []string{}, 0, 0)
	_, err := s.msgServer.AddMarker(s.ctx, addMarkerMsg)
	s.Assert().NoError(err, "should successfully add marker")

	addAccessMsg := types.NewMsgAddAccessRequest(hotdogDenom, s.owner1Addr, access)
	_, err = s.msgServer.AddAccess(s.ctx, addAccessMsg)
	s.Assert().NoError(err, "should not throw error when adding access to marker")

	finalizeMsg := types.NewMsgFinalizeRequest(hotdogDenom, s.owner1Addr)
	_, err = s.msgServer.Finalize(s.ctx, finalizeMsg)
	s.Assert().NoError(err, "should not throw error when finalizing marker")

	activateMsg := types.NewMsgActivateRequest(hotdogDenom, s.owner1Addr)
	_, err = s.msgServer.Activate(s.ctx, activateMsg)
	s.Assert().NoError(err, "should not throw error when activating marker message")

	testcases := []struct {
		name          string
		msg           *types.MsgWithdrawRequest
		expectedEvent proto.Message
	}{
		{
			name:          "should successfully withdraw marker",
			msg:           types.NewMsgWithdrawRequest(s.owner1Addr, s.owner1Addr, hotdogDenom, sdk.NewCoins(sdk.NewInt64Coin(hotdogDenom, 100))),
			expectedEvent: types.NewEventMarkerWithdraw("100hotdog", hotdogDenom, s.owner1, s.owner1),
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
			response, err := s.msgServer.Withdraw(s.ctx, tc.msg)
			s.Require().NoError(err, "handler(%T) error", tc.msg)
			if tc.expectedEvent != nil {
				result := s.containsMessage(s.ctx.EventManager().ABCIEvents(), tc.expectedEvent)
				s.Assert().True(result, "Expected typed event was not found in response.\n    Expected: %+v\n    Response: %+v", tc.expectedEvent, response)
			}
		})
	}
}

func (s *MsgServerTestSuite) TestMsgTransferMarkerRequest() {
	hotdogDenom := "hotdog"
	access := types.AccessGrant{
		Address:     s.owner1,
		Permissions: types.AccessListByNames("DELETE,MINT,WITHDRAW,TRANSFER"),
	}

	addMarkerMsg := types.NewMsgAddMarkerRequest(hotdogDenom, sdkmath.NewInt(100), s.owner1Addr, s.owner1Addr, types.MarkerType_RestrictedCoin, true, true, false, []string{}, 0, 0)
	_, err := s.msgServer.AddMarker(s.ctx, addMarkerMsg)
	s.Assert().NoError(err, "should sucessfully add marker")

	addAccessMsg := types.NewMsgAddAccessRequest(hotdogDenom, s.owner1Addr, access)
	_, err = s.msgServer.AddAccess(s.ctx, addAccessMsg)
	s.Assert().NoError(err, "should not throw error when adding access to marker")

	finalizeMsg := types.NewMsgFinalizeRequest(hotdogDenom, s.owner1Addr)
	_, err = s.msgServer.Finalize(s.ctx, finalizeMsg)
	s.Assert().NoError(err, "should not throw error when finalizing marker")

	activateMsg := types.NewMsgActivateRequest(hotdogDenom, s.owner1Addr)
	_, err = s.msgServer.Activate(s.ctx, activateMsg)
	s.Assert().NoError(err, "should not throw error when activating marker message")

	mintMsg := types.NewMsgMintRequest(s.owner1Addr, sdk.NewInt64Coin(hotdogDenom, 1000), s.owner2Addr)
	_, err = s.msgServer.Mint(s.ctx, mintMsg)
	s.Assert().NoError(err, "should not throw error when minting marker")

	testcases := []struct {
		name          string
		msg           *types.MsgTransferRequest
		expectedEvent proto.Message
	}{
		{
			name:          "should successfully transfer marker",
			msg:           types.NewMsgTransferRequest(s.owner1Addr, s.owner1Addr, s.owner2Addr, sdk.NewInt64Coin(hotdogDenom, 0)),
			expectedEvent: types.NewEventMarkerTransfer("0", hotdogDenom, s.owner1, s.owner2, s.owner1),
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
			response, err := s.msgServer.Transfer(s.ctx, tc.msg)
			s.Require().NoError(err, "handler(%T) error", tc.msg)
			if tc.expectedEvent != nil {
				result := s.containsMessage(s.ctx.EventManager().ABCIEvents(), tc.expectedEvent)
				s.Assert().True(result, "Expected typed event was not found in response.\n    Expected: %+v\n    Response: %+v", tc.expectedEvent, response)
			}
		})
	}
}

func (s *MsgServerTestSuite) TestMsgSetDenomMetadataRequest() {
	hotdogDenom := "hotdog"
	hotdogName := "Jason"
	hotdogSymbol := "WIFI"
	access := types.AccessGrant{
		Address:     s.owner1,
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

	addMarkerMsg := types.NewMsgAddMarkerRequest(fmt.Sprintf("n%s", hotdogDenom), sdkmath.NewInt(100), s.owner1Addr, s.owner1Addr, types.MarkerType_RestrictedCoin, true, true, false, []string{}, 0, 0)
	_, err := s.msgServer.AddMarker(s.ctx, addMarkerMsg)
	s.Assert().NoError(err, "should successfully add marker")

	addAccessMsg := types.NewMsgAddAccessRequest(fmt.Sprintf("n%s", hotdogDenom), s.owner1Addr, access)
	_, err = s.msgServer.AddAccess(s.ctx, addAccessMsg)
	s.Assert().NoError(err, "should not throw error when adding access to marker")

	testcases := []struct {
		name          string
		msg           *types.MsgSetDenomMetadataRequest
		expectedEvent proto.Message
	}{
		{
			name:          "should successfully set denom metadata on marker",
			msg:           types.NewSetDenomMetadataRequest(hotdogMetadata, s.owner1Addr),
			expectedEvent: types.NewEventMarkerSetDenomMetadata(hotdogMetadata, s.owner1),
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
			response, err := s.msgServer.SetDenomMetadata(s.ctx, tc.msg)
			s.Require().NoError(err, "handler(%T) error", tc.msg)
			if tc.expectedEvent != nil {
				result := s.containsMessage(s.ctx.EventManager().ABCIEvents(), tc.expectedEvent)
				s.Assert().True(result, "Expected typed event was not found in response.\n    Expected: %+v\n    Response: %+v", tc.expectedEvent, response)
			}
		})
	}
}

func (s *MsgServerTestSuite) TestMsgAddFinalizeActivateMarkerRequest() {
	denom := "hotdog"
	rdenom := "restrictedhotdog"
	denomWithDashPeriod := fmt.Sprintf("%s-my.marker", denom)
	access_owner_1 := types.AccessGrant{
		Address:     s.owner1,
		Permissions: types.AccessList{types.Access_Mint, types.Access_Withdraw, types.Access_Admin},
	}
	testcases := []struct {
		name          string
		handler       func(sdk.Context) error
		expectedEvent proto.Message
		errorMsg      string
	}{

		{
			name: "should successfully ADD,FINALIZE,ACTIVATE new marker",
			handler: func(ctx sdk.Context) error {
				msg := types.NewMsgAddFinalizeActivateMarkerRequest(denom, sdkmath.NewInt(100), s.owner1Addr, s.owner1Addr, types.MarkerType_Coin, true, true, false, []string{}, []types.AccessGrant{*types.NewAccessGrant(s.owner1Addr, []types.Access{types.Access_Mint, types.Access_Admin})}, 0, 0)
				_, err := s.msgServer.AddFinalizeActivateMarker(s.ctx, msg)
				return err
			},
			expectedEvent: types.NewEventMarkerAdd(denom, types.MustGetMarkerAddress(denom).String(), "100", "proposed", s.owner1, types.MarkerType_Coin.String()),
		},
		{
			name: "should successfully ADD,FINALIZE,ACTIVATE new marker with attributes",
			handler: func(ctx sdk.Context) error {
				msg := types.NewMsgAddFinalizeActivateMarkerRequest(rdenom, sdkmath.NewInt(100), s.owner1Addr, s.owner1Addr, types.MarkerType_RestrictedCoin, true, true, false, []string{"attributes.one.com", "attributes.two.com"}, []types.AccessGrant{*types.NewAccessGrant(s.owner1Addr, []types.Access{types.Access_Mint, types.Access_Admin})}, 0, 0)
				_, err := s.msgServer.AddFinalizeActivateMarker(s.ctx, msg)
				return err
			},
			expectedEvent: types.NewEventMarkerAdd(rdenom, types.MustGetMarkerAddress(rdenom).String(), "100", "proposed", s.owner1, types.MarkerType_RestrictedCoin.String()),
		},
		{
			name: "should fail to ADD,FINALIZE,ACTIVATE new marker, validate basic failure",
			handler: func(ctx sdk.Context) error {
				msg := types.NewMsgAddFinalizeActivateMarkerRequest(denom, sdkmath.NewInt(100), s.owner1Addr, s.owner1Addr, types.MarkerType_Coin, true, true, false, []string{}, nil, 0, 0)
				_, err := s.msgServer.AddFinalizeActivateMarker(s.ctx, msg)
				return err
			},
			errorMsg: "since this will activate the marker, must have at least one access list defined: invalid request",
		},
		{
			name: "should fail to ADD,FINALIZE,ACTIVATE new marker, marker already exists",
			handler: func(ctx sdk.Context) error {
				msg := types.NewMsgAddMarkerRequest(denom, sdkmath.NewInt(100), s.owner1Addr, s.owner1Addr, types.MarkerType_Coin, true, true, false, []string{}, 0, 0)
				_, err := s.msgServer.AddMarker(s.ctx, msg)
				return err
			},
			errorMsg: fmt.Sprintf("marker address already exists for %s: invalid request", types.MustGetMarkerAddress(denom)),
		},
		{
			name: "should successfully add marker with dash and period",
			handler: func(ctx sdk.Context) error {
				msg := types.NewMsgAddMarkerRequest(denomWithDashPeriod, sdkmath.NewInt(1000), s.owner1Addr, s.owner1Addr, types.MarkerType_Coin, true, true, false, []string{}, 0, 0)
				_, err := s.msgServer.AddMarker(s.ctx, msg)
				return err
			},
			expectedEvent: types.NewEventMarkerAdd(denomWithDashPeriod, types.MustGetMarkerAddress(denomWithDashPeriod).String(), "1000", "proposed", s.owner1, types.MarkerType_Coin.String()),
		},
		{
			name: "should not throw error when adding access to marker",
			handler: func(ctx sdk.Context) error {
				msg := types.NewMsgAddAccessRequest(denom, s.owner1Addr, access_owner_1)
				_, err := s.msgServer.AddAccess(s.ctx, msg)
				return err
			},
			expectedEvent: types.NewEventMarkerAddAccess(&access_owner_1, denom, s.owner1Addr.String()),
		},

		{
			name: "should successfully mint denom",
			handler: func(ctx sdk.Context) error {
				msg := types.NewMsgMintRequest(s.owner1Addr, sdk.NewInt64Coin(denom, 1000), s.owner1Addr)
				_, err := s.msgServer.Mint(s.ctx, msg)
				return err
			},
			expectedEvent: types.NewEventMarkerMint("1000", denom, s.owner1),
		},
		{
			name: "should fail to burn denom, user doesn't have permissions",
			handler: func(ctx sdk.Context) error {
				msg := types.NewMsgBurnRequest(s.owner1Addr, sdk.NewInt64Coin(denom, 50))
				_, err := s.msgServer.Burn(s.ctx, msg)
				return err
			},
			errorMsg: s.noAccessErr(s.owner1, types.Access_Burn, denom) + ": invalid request",
		},
	}
	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
			err := tc.handler(s.ctx)
			if len(tc.errorMsg) > 0 {
				s.Require().EqualError(err, tc.errorMsg, "should have the correct error")
			} else {
				s.Require().NoError(err, "should have no error")
				if tc.expectedEvent != nil {
					result := s.containsMessage(s.ctx.EventManager().ABCIEvents(), tc.expectedEvent)
					s.Assert().True(result, "Expected typed event was not found in response.\n    Expected: %+v\n", tc.expectedEvent)
				}
			}
		})
	}
}

func (s *MsgServerTestSuite) TestMsgSetAccountDataRequest() {
	denomU := "aducoin"
	denomR := "adrcoin"

	denomUAddr := types.MustGetMarkerAddress(denomU).String()
	denomRAddr := types.MustGetMarkerAddress(denomR).String()

	authority := s.app.MarkerKeeper.GetAuthority()

	msg := types.NewMsgAddFinalizeActivateMarkerRequest(
		denomU, sdkmath.NewInt(100),
		s.owner1Addr, s.owner1Addr, // From and Manager.
		types.MarkerType_Coin,
		true,       // Supply fixed
		true,       // Allow gov
		false,      // don't allow forced transfer
		[]string{}, // No required attributes.
		[]types.AccessGrant{
			{Address: s.owner1, Permissions: []types.Access{types.Access_Mint, types.Access_Admin}},
			{Address: s.owner2, Permissions: []types.Access{types.Access_Deposit}},
		},
		0,
		0,
	)
	_, err := s.msgServer.AddFinalizeActivateMarker(s.ctx, msg)
	s.Assert().NoError(err, "should successfully add/finalize/active unrestricted marker")

	msg = types.NewMsgAddFinalizeActivateMarkerRequest(
		denomR, sdkmath.NewInt(100),
		s.owner1Addr, s.owner1Addr, // From and Manager.
		types.MarkerType_RestrictedCoin,
		true,       // Supply fixed
		true,       // Allow gov
		false,      // don't allow forced transfer
		[]string{}, // No required attributes.
		[]types.AccessGrant{
			{Address: s.owner1, Permissions: []types.Access{types.Access_Mint, types.Access_Admin}},
			{Address: s.owner2, Permissions: []types.Access{types.Access_Deposit}},
		},
		0,
		0,
	)
	_, err = s.msgServer.AddFinalizeActivateMarker(s.ctx, msg)
	s.Assert().NoError(err, "should successfully add/finalize/active restricted marker")

	testcases := []struct {
		name          string
		msg           *types.MsgSetAccountDataRequest
		expectedEvent proto.Message
		errorMsg      string
	}{
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
				Signer: s.owner2,
			},
			expectedEvent: &attrtypes.EventAccountDataUpdated{Account: denomUAddr},
		},
		{
			name: "should fail to set account data on unrestricted marker because signer does not have deposit",
			msg: &types.MsgSetAccountDataRequest{
				Denom:  denomU,
				Value:  "This is some unrestricted coin data. This won't get used though.",
				Signer: s.owner1,
			},
			errorMsg: s.noAccessErr(s.owner1, types.Access_Deposit, denomU),
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
				Signer: s.owner2,
			},
			expectedEvent: &attrtypes.EventAccountDataUpdated{Account: denomRAddr},
		},
		{
			name: "should fail to set account data on restricted marker because signer does not have deposit",
			msg: &types.MsgSetAccountDataRequest{
				Denom:  denomR,
				Value:  "This is some restricted coin data. This won't get used though.",
				Signer: s.owner1,
			},
			errorMsg: s.noAccessErr(s.owner1, types.Access_Deposit, denomR),
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
			response, err := s.msgServer.SetAccountData(s.ctx, tc.msg)
			if len(tc.errorMsg) > 0 {
				s.Require().EqualError(err, tc.errorMsg, "handler(%T) error", tc.msg)
			} else {
				s.Require().NoError(err, "handler(%T) error", tc.msg)
				if tc.expectedEvent != nil {
					result := s.containsMessage(s.ctx.EventManager().ABCIEvents(), tc.expectedEvent)
					s.Assert().True(result, "Expected typed event was not found in response.\n    Expected: %+v\n    Response: %+v", tc.expectedEvent, response)
				}
			}
		})
	}
}

func (s *MsgServerTestSuite) TestSetAdministratorProposal() {
	hotdogMarker := types.NewMarkerAccount(
		authtypes.NewBaseAccountWithAddress(types.MustGetMarkerAddress("hotdog")),
		sdk.NewInt64Coin("hotdog", 1000),
		s.owner1Addr,
		[]types.AccessGrant{
			{Address: s.owner1Addr.String(), Permissions: types.AccessList{types.Access_Admin, types.Access_Mint}},
		},
		types.StatusActive,
		types.MarkerType_Coin,
		true,
		true,
		false,
		[]string{},
	)
	s.Require().NoError(s.app.MarkerKeeper.AddMarkerAccount(s.ctx, hotdogMarker), "Failed to add 'hotdog' marker for tests")

	testCases := []struct {
		name   string
		msg    *types.MsgSetAdministratorProposalRequest
		expErr string
	}{
		{
			name: "success case",
			msg: &types.MsgSetAdministratorProposalRequest{
				Denom:     "hotdog",
				Access:    []types.AccessGrant{{Address: s.owner1, Permissions: types.AccessList{types.Access_Mint}}},
				Authority: s.app.MarkerKeeper.GetAuthority(),
			},
		},
		{
			name: "failed authority",
			msg: &types.MsgSetAdministratorProposalRequest{
				Denom:     "hotdog",
				Access:    []types.AccessGrant{{Address: s.owner1, Permissions: types.AccessList{types.Access_Mint}}},
				Authority: "wrongauthority",
			},
			expErr: "expected gov account as only signer for proposal message",
		},
		{
			name: "failed handler",
			msg: &types.MsgSetAdministratorProposalRequest{
				Denom:     "nonexistent",
				Access:    []types.AccessGrant{{Address: s.owner1, Permissions: types.AccessList{types.Access_Mint}}},
				Authority: s.app.MarkerKeeper.GetAuthority(),
			},
			expErr: "nonexistent marker does not exist",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			_, err := s.msgServer.SetAdministratorProposal(s.ctx, tc.msg)
			if len(tc.expErr) > 0 {
				s.Assert().Error(err, tc.expErr, "SetAdministratorProposal() error incorrect.")
			} else {
				s.Assert().NoError(err, "SetAdministratorProposal() should have no error for valid request.")
			}
		})
	}
}

func (s *MsgServerTestSuite) TestSupplyIncreaseProposal() {
	hotdogMarker := types.NewMarkerAccount(
		authtypes.NewBaseAccountWithAddress(types.MustGetMarkerAddress("hotdog")),
		sdk.NewInt64Coin("hotdog", 1000),
		s.owner1Addr,
		[]types.AccessGrant{
			{Address: s.owner1Addr.String(), Permissions: types.AccessList{types.Access_Admin, types.Access_Mint}},
		},
		types.StatusActive,
		types.MarkerType_Coin,
		true,
		true,
		false,
		[]string{},
	)
	s.Require().NoError(s.app.MarkerKeeper.AddMarkerAccount(s.ctx, hotdogMarker), "Failed to add 'hotdog' marker for tests")

	testCases := []struct {
		name   string
		msg    *types.MsgSupplyIncreaseProposalRequest
		expErr string
	}{
		{
			name: "success case",
			msg: &types.MsgSupplyIncreaseProposalRequest{
				Amount:        sdk.NewInt64Coin("hotdog", 1000),
				TargetAddress: s.owner1,
				Authority:     s.app.MarkerKeeper.GetAuthority(),
			},
		},
		{
			name: "failed authority",
			msg: &types.MsgSupplyIncreaseProposalRequest{
				Amount:        sdk.NewInt64Coin("hotdog", 1000),
				TargetAddress: s.owner1,
				Authority:     "wrongauthority",
			},
			expErr: "expected gov account as only signer for proposal message",
		},
		{
			name: "failed handler",
			msg: &types.MsgSupplyIncreaseProposalRequest{
				Amount:        sdk.NewInt64Coin("nonexistent", 1000),
				TargetAddress: s.owner1,
				Authority:     s.app.MarkerKeeper.GetAuthority(),
			},
			expErr: "nonexistent marker does not exist",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			_, err := s.msgServer.SupplyIncreaseProposal(s.ctx, tc.msg)
			if len(tc.expErr) > 0 {
				s.Assert().Error(err, tc.expErr, "SupplyIncreaseProposal() error incorrect.")
			} else {
				s.Assert().NoError(err, "SupplyIncreaseProposal() should have no error for valid request.")
			}
		})
	}
}

func (s *MsgServerTestSuite) TestSupplyDecreaseProposal() {
	hotdogMarker := types.NewMarkerAccount(
		authtypes.NewBaseAccountWithAddress(types.MustGetMarkerAddress("hotdog")),
		sdk.NewInt64Coin("hotdog", 1000),
		s.owner1Addr,
		[]types.AccessGrant{
			{Address: s.owner1Addr.String(), Permissions: types.AccessList{types.Access_Admin, types.Access_Mint, types.Access_Burn}},
		},
		types.StatusProposed,
		types.MarkerType_Coin,
		true,
		true,
		false,
		[]string{},
	)
	s.Require().NoError(s.app.MarkerKeeper.AddSetNetAssetValues(s.ctx, hotdogMarker, []types.NetAssetValue{types.NewNetAssetValue(sdk.NewInt64Coin(types.UsdDenom, 1), 1)}, types.ModuleName), "Failed to add navs to 'hotdog' marker for tests")
	s.Require().NoError(s.app.MarkerKeeper.AddFinalizeAndActivateMarker(s.ctx, hotdogMarker), "Failed to add 'hotdog' marker for tests")

	testCases := []struct {
		name   string
		msg    *types.MsgSupplyDecreaseProposalRequest
		expErr string
	}{
		{
			name: "success case",
			msg: &types.MsgSupplyDecreaseProposalRequest{
				Amount:    sdk.NewInt64Coin("hotdog", 100),
				Authority: s.app.MarkerKeeper.GetAuthority(),
			},
		},
		{
			name: "failed authority",
			msg: &types.MsgSupplyDecreaseProposalRequest{
				Amount:    sdk.NewInt64Coin("hotdog", 500),
				Authority: "wrongauthority",
			},
			expErr: "expected gov account as only signer for proposal message",
		},
		{
			name: "failed handler",
			msg: &types.MsgSupplyDecreaseProposalRequest{
				Amount:    sdk.NewInt64Coin("nonexistent", 500),
				Authority: s.app.MarkerKeeper.GetAuthority(),
			},
			expErr: "nonexistent marker does not exist",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			_, err := s.msgServer.SupplyDecreaseProposal(s.ctx, tc.msg)
			if len(tc.expErr) > 0 {
				s.Assert().Error(err, tc.expErr, "SupplyDecreaseProposal() error incorrect.")
			} else {
				s.Assert().NoError(err, "SupplyDecreaseProposal() should have no error for valid request.")
			}
		})
	}
}

func (s *MsgServerTestSuite) TestRemoveAdministratorProposal() {
	hotdogMarker := types.NewMarkerAccount(
		authtypes.NewBaseAccountWithAddress(types.MustGetMarkerAddress("hotdog")),
		sdk.NewInt64Coin("hotdog", 1000),
		s.owner1Addr,
		[]types.AccessGrant{
			{Address: s.owner1Addr.String(), Permissions: types.AccessList{types.Access_Admin, types.Access_Mint}},
		},
		types.StatusActive,
		types.MarkerType_Coin,
		true,
		true,
		false,
		[]string{},
	)
	s.Require().NoError(s.app.MarkerKeeper.AddMarkerAccount(s.ctx, hotdogMarker), "Failed to add 'hotdog' marker for tests")

	testCases := []struct {
		name   string
		msg    *types.MsgRemoveAdministratorProposalRequest
		expErr string
	}{
		{
			name: "success case",
			msg: &types.MsgRemoveAdministratorProposalRequest{
				Denom:          "hotdog",
				RemovedAddress: []string{s.owner1},
				Authority:      s.app.MarkerKeeper.GetAuthority(),
			},
		},
		{
			name: "failed authority",
			msg: &types.MsgRemoveAdministratorProposalRequest{
				Denom:          "hotdog",
				RemovedAddress: []string{s.owner1},
				Authority:      "wrongauthority",
			},
			expErr: "expected gov account as only signer for proposal message",
		},
		{
			name: "failed handler",
			msg: &types.MsgRemoveAdministratorProposalRequest{
				Denom:          "nonexistent",
				RemovedAddress: []string{s.owner1},
				Authority:      s.app.MarkerKeeper.GetAuthority(),
			},
			expErr: "nonexistent marker does not exist",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			_, err := s.msgServer.RemoveAdministratorProposal(s.ctx, tc.msg)
			if len(tc.expErr) > 0 {
				s.Assert().Error(err, "RemoveAdministratorProposal() error incorrect.")
			} else {
				s.Assert().NoError(err, "RemoveAdministratorProposal() should have no error for valid request.")
			}
		})
	}
}

func (s *MsgServerTestSuite) TestChangeStatusProposal() {
	hotdogMarker := types.NewMarkerAccount(
		authtypes.NewBaseAccountWithAddress(types.MustGetMarkerAddress("hotdog")),
		sdk.NewInt64Coin("hotdog", 1000),
		s.owner1Addr,
		[]types.AccessGrant{
			{Address: s.owner1Addr.String(), Permissions: types.AccessList{types.Access_Admin, types.Access_Mint}},
		},
		types.StatusFinalized,
		types.MarkerType_Coin,
		true,
		true,
		false,
		[]string{},
	)
	s.Require().NoError(s.app.MarkerKeeper.AddMarkerAccount(s.ctx, hotdogMarker), "Failed to add 'hotdog' marker for tests")

	testCases := []struct {
		name   string
		msg    *types.MsgChangeStatusProposalRequest
		expErr string
	}{
		{
			name: "success case",
			msg: &types.MsgChangeStatusProposalRequest{
				Denom:     "hotdog",
				NewStatus: types.StatusActive,
				Authority: s.app.MarkerKeeper.GetAuthority(),
			},
		},
		{
			name: "failed authority",
			msg: &types.MsgChangeStatusProposalRequest{
				Denom:     "hotdog",
				NewStatus: types.StatusActive,
				Authority: "wrongauthority",
			},
			expErr: "expected gov account as only signer for proposal message",
		},
		{
			name: "failed handler",
			msg: &types.MsgChangeStatusProposalRequest{
				Denom:     "nonexistent",
				NewStatus: types.StatusActive,
				Authority: s.app.MarkerKeeper.GetAuthority(),
			},
			expErr: "nonexistent marker does not exist",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			_, err := s.msgServer.ChangeStatusProposal(s.ctx, tc.msg)
			if len(tc.expErr) > 0 {
				s.Assert().Error(err, "ChangeStatusProposal() error incorrect.")
			} else {
				s.Assert().NoError(err, "ChangeStatusProposal() should have no error for valid request.")
			}
		})
	}
}

func (s *MsgServerTestSuite) TestWithdrawEscrowProposal() {
	hotdogMarker := types.NewMarkerAccount(
		authtypes.NewBaseAccountWithAddress(types.MustGetMarkerAddress("hotdog")),
		sdk.NewInt64Coin("hotdog", 1000),
		s.owner1Addr,
		[]types.AccessGrant{
			{Address: s.owner1Addr.String(), Permissions: types.AccessList{types.Access_Admin, types.Access_Mint}},
		},
		types.StatusProposed,
		types.MarkerType_Coin,
		true,
		true,
		false,
		[]string{},
	)
	s.Require().NoError(s.app.MarkerKeeper.AddSetNetAssetValues(s.ctx, hotdogMarker, []types.NetAssetValue{types.NewNetAssetValue(sdk.NewInt64Coin(types.UsdDenom, 1), 1)}, types.ModuleName), "Failed to add navs to 'hotdog' marker for tests")
	s.Require().NoError(s.app.MarkerKeeper.AddFinalizeAndActivateMarker(s.ctx, hotdogMarker), "Failed to add 'hotdog' marker for tests")

	testCases := []struct {
		name   string
		msg    *types.MsgWithdrawEscrowProposalRequest
		expErr string
	}{
		{
			name: "success case",
			msg: &types.MsgWithdrawEscrowProposalRequest{
				Denom:         "hotdog",
				TargetAddress: s.owner2Addr.String(),
				Amount:        sdk.NewCoins(sdk.NewInt64Coin("hotdog", 500)),
				Authority:     s.app.MarkerKeeper.GetAuthority(),
			},
		},
		{
			name: "failed authority",
			msg: &types.MsgWithdrawEscrowProposalRequest{
				Denom:         "hotdog",
				TargetAddress: s.owner2Addr.String(),
				Amount:        sdk.NewCoins(sdk.NewInt64Coin("hotdog", 500)),
				Authority:     "wrongauthority",
			},
			expErr: "expected gov account as only signer for proposal message",
		},
		{
			name: "failed handler",
			msg: &types.MsgWithdrawEscrowProposalRequest{
				Denom:         "nonexistent",
				TargetAddress: s.owner2Addr.String(),
				Amount:        sdk.NewCoins(sdk.NewInt64Coin("nonexistent", 500)),
				Authority:     s.app.MarkerKeeper.GetAuthority(),
			},
			expErr: "nonexistent marker does not exist",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			_, err := s.msgServer.WithdrawEscrowProposal(s.ctx, tc.msg)
			if len(tc.expErr) > 0 {
				s.Assert().Error(err, tc.expErr, "WithdrawEscrowProposal() error incorrect.")
			} else {
				s.Assert().NoError(err, "WithdrawEscrowProposal() should have no error for valid request.")
			}
		})
	}
}

func (s *MsgServerTestSuite) TestSetDenomMetadataProposal() {
	hotdogMarker := types.NewMarkerAccount(
		authtypes.NewBaseAccountWithAddress(types.MustGetMarkerAddress("hotdog")),
		sdk.NewInt64Coin("hotdog", 1000),
		s.owner1Addr,
		[]types.AccessGrant{
			{Address: s.owner1Addr.String(), Permissions: types.AccessList{types.Access_Admin, types.Access_Mint}},
		},
		types.StatusProposed,
		types.MarkerType_Coin,
		true,
		true,
		false,
		[]string{},
	)
	s.Require().NoError(s.app.MarkerKeeper.AddSetNetAssetValues(s.ctx, hotdogMarker, []types.NetAssetValue{types.NewNetAssetValue(sdk.NewInt64Coin(types.UsdDenom, 1), 1)}, types.ModuleName), "Failed to add navs to 'hotdog' marker for tests")
	s.Require().NoError(s.app.MarkerKeeper.AddFinalizeAndActivateMarker(s.ctx, hotdogMarker), "Failed to add 'hotdog' marker for tests")

	hotdogMetadata := banktypes.Metadata{
		Description: "Hotdog Coin",
		DenomUnits: []*banktypes.DenomUnit{
			{Denom: "uhotdog", Exponent: 0, Aliases: []string{"microhotdog"}},
			{Denom: "mhotdog", Exponent: 3, Aliases: []string{}},
			{Denom: "hotdog", Exponent: 6, Aliases: []string{}},
		},
		Base:    "hotdog",
		Display: "hotdog",
		Name:    "Hotdog",
		Symbol:  "HOTDOG",
	}

	testCases := []struct {
		name   string
		msg    *types.MsgSetDenomMetadataProposalRequest
		expErr string
	}{
		{
			name: "success case",
			msg: &types.MsgSetDenomMetadataProposalRequest{
				Metadata:  hotdogMetadata,
				Authority: s.app.MarkerKeeper.GetAuthority(),
			},
		},
		{
			name: "failed authority",
			msg: &types.MsgSetDenomMetadataProposalRequest{
				Metadata:  hotdogMetadata,
				Authority: "wrongauthority",
			},
			expErr: "expected gov account as only signer for proposal message",
		},
		{
			name: "failed handler",
			msg: &types.MsgSetDenomMetadataProposalRequest{
				Metadata:  banktypes.Metadata{Base: "nonexistent"},
				Authority: s.app.MarkerKeeper.GetAuthority(),
			},
			expErr: "nonexistent marker does not exist",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			_, err := s.msgServer.SetDenomMetadataProposal(s.ctx, tc.msg)
			if len(tc.expErr) > 0 {
				s.Assert().Error(err, tc.expErr, "SetDenomMetadataProposal() error incorrect.")
			} else {
				s.Assert().NoError(err, "SetDenomMetadataProposal() should have no error for valid request.")
			}
		})
	}
}

func (s *MsgServerTestSuite) TestMsgUpdateParamsRequest() {
	authority := s.app.MarkerKeeper.GetAuthority()

	testCases := []struct {
		name   string
		msg    *types.MsgUpdateParamsRequest
		expErr string
	}{
		{
			name: "successfully update params",
			msg: &types.MsgUpdateParamsRequest{
				Authority: authority,
				Params: types.NewParams(
					true,
					"[a-zA-Z][a-zA-Z0-9\\-\\.]{2,83}",
					sdkmath.NewInt(1000000000000),
				),
			},
		},
		{
			name: "fail to update params, invalid authority",
			msg: &types.MsgUpdateParamsRequest{
				Authority: "invalidAuthority",
				Params: types.NewParams(
					true,
					"[a-zA-Z][a-zA-Z0-9\\-\\.]{2,83}",
					sdkmath.NewInt(1000000000000),
				),
			},
			expErr: `expected "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn" got "invalidAuthority": expected gov account as only signer for proposal message`,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			ctx := sdk.WrapSDKContext(s.ctx)
			res, err := s.msgServer.UpdateParams(ctx, tc.msg)
			if len(tc.expErr) > 0 {
				s.Require().EqualError(err, tc.expErr, "UpdateParams error")
				s.Require().Nil(res, "UpdateParams response")
			} else {
				s.Require().NoError(err, "UpdateParams error")
				s.Require().NotNil(res, "UpdateParams response")
			}
		})
	}
}
