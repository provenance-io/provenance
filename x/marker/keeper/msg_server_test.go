package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/stretchr/testify/suite"

	simapp "github.com/provenance-io/provenance/app"
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
	acct1      authtypes.AccountI

	addresses []sdk.AccAddress
}

func (s *MsgServerTestSuite) SetupTest() {

	s.blockStartTime = time.Now()
	s.app = simapp.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{
		Time: s.blockStartTime,
	})
	s.msgServer = markerkeeper.NewMsgServerImpl(s.app.MarkerKeeper)

	s.privkey1 = secp256k1.GenPrivKey()
	s.pubkey1 = s.privkey1.PubKey()
	s.owner1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.owner1 = s.owner1Addr.String()
	acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.owner1Addr)
	s.app.AccountKeeper.SetAccount(s.ctx, acc)
}
func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerTestSuite))
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
			Supply:                 sdk.NewInt(1000),
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
			goCtx := sdk.WrapSDKContext(s.ctx.WithEventManager(em))
			var res *types.MsgUpdateForcedTransferResponse
			var err error
			testFunc := func() {
				res, err = s.msgServer.UpdateForcedTransfer(goCtx, tc.msg)
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

				expEvents := sdk.Events{
					{
						Type: sdk.EventTypeMessage,
						Attributes: []abci.EventAttribute{
							{Key: []byte(sdk.AttributeKeyModule), Value: []byte(types.ModuleName)},
						},
					},
				}
				events := em.Events()
				s.Assert().Equal(expEvents, events, "events emitted during UpdateForcedTransfer")
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
	s.app.MarkerKeeper.SetMarker(s.ctx, types.NewMarkerAccount(rMarkerAcct, sdk.NewInt64Coin(rMarkerDenom, 1000), authUser, []types.AccessGrant{{Address: authUser.String(), Permissions: []types.Access{types.Access_Transfer}}}, types.StatusFinalized, types.MarkerType_RestrictedCoin, true, false, false, []string{}))

	rMarkerGovDenom := "restricted-marker-gov"
	rMarkerGovAcct := authtypes.NewBaseAccount(types.MustGetMarkerAddress(rMarkerGovDenom), nil, 0, 0)
	s.app.MarkerKeeper.SetMarker(s.ctx, types.NewMarkerAccount(rMarkerGovAcct, sdk.NewInt64Coin(rMarkerGovDenom, 1000), authUser, []types.AccessGrant{{Address: authUser.String(), Permissions: []types.Access{}}}, types.StatusFinalized, types.MarkerType_RestrictedCoin, true, true, false, []string{}))

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
			expErr: "cosmos1ku2jzvpkt4ffxxaajyk2r88axk9cr5jqlthcm4 does not have transfer authority for restricted-marker marker",
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
			res, err := s.msgServer.UpdateSendDenyList(sdk.WrapSDKContext(s.ctx),
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

	markerDenom := "jackthecat"
	markerAcct := authtypes.NewBaseAccount(types.MustGetMarkerAddress(markerDenom), nil, 0, 0)
	s.app.MarkerKeeper.SetMarker(s.ctx, types.NewMarkerAccount(markerAcct, sdk.NewInt64Coin(markerDenom, 1000), authUser, []types.AccessGrant{{Address: authUser.String(), Permissions: []types.Access{types.Access_Transfer}}}, types.StatusProposed, types.MarkerType_RestrictedCoin, true, false, false, []string{}))

	valueDenom := "usd"
	valueAcct := authtypes.NewBaseAccount(types.MustGetMarkerAddress(valueDenom), nil, 0, 0)
	s.app.MarkerKeeper.SetMarker(s.ctx, types.NewMarkerAccount(valueAcct, sdk.NewInt64Coin(valueDenom, 1000), authUser, []types.AccessGrant{{Address: authUser.String(), Permissions: []types.Access{types.Access_Transfer}}}, types.StatusProposed, types.MarkerType_RestrictedCoin, true, false, false, []string{}))

	finalizedMarkerDenom := "finalizedjackthecat"
	finalizedMarkerAcct := authtypes.NewBaseAccount(types.MustGetMarkerAddress(finalizedMarkerDenom), nil, 1, 0)
	s.app.MarkerKeeper.SetMarker(s.ctx, types.NewMarkerAccount(finalizedMarkerAcct, sdk.NewInt64Coin(finalizedMarkerDenom, 1000), authUser, []types.AccessGrant{{Address: authUser.String(), Permissions: []types.Access{types.Access_Transfer}}}, types.StatusFinalized, types.MarkerType_RestrictedCoin, true, false, false, []string{}))

	testCases := []struct {
		name   string
		msg    types.MsgAddNetAssetValueRequest
		expErr string
	}{
		{
			name:   "no marker found",
			msg:    types.MsgAddNetAssetValueRequest{Denom: "cantfindme", NetAssetValues: []types.NetAssetValue{}, Administrator: authUser.String()},
			expErr: "marker cantfindme not found for address: cosmos17l2yneua2mdfqaycgyhqag8t20asnjwf6adpmt: invalid request",
		},
		{
			name:   "marker is not in proposed state",
			msg:    types.MsgAddNetAssetValueRequest{Denom: finalizedMarkerDenom, NetAssetValues: []types.NetAssetValue{}, Administrator: authUser.String()},
			expErr: "can only add net asset values to markers in the Proposed status: invalid request",
		},
		{
			name: "nav denom matches marker denom",
			msg: types.MsgAddNetAssetValueRequest{
				Denom: markerDenom,
				NetAssetValues: []types.NetAssetValue{
					{
						Value:      sdk.NewInt64Coin(markerDenom, 100),
						Volume:     uint64(100),
						Source:     "exchange",
						UpdateTime: s.blockStartTime,
					},
				},
				Administrator: authUser.String(),
			},
			expErr: `net asset value denom cannot match marker denom "jackthecat": invalid request`,
		},
		{
			name: "value denom does not exist",
			msg: types.MsgAddNetAssetValueRequest{
				Denom: markerDenom,
				NetAssetValues: []types.NetAssetValue{
					{
						Value:      sdk.NewInt64Coin("hotdog", 100),
						Volume:     uint64(100),
						Source:     "exchange",
						UpdateTime: s.blockStartTime,
					},
				},
				Administrator: authUser.String(),
			},
			expErr: `net asset value denom does not exist: marker hotdog not found for address: cosmos1p6l3annxy35gm5mfm6m0jz2mdj8peheuzf9alh: invalid request`,
		},
		{
			name: "time is in the future of current block",
			msg: types.MsgAddNetAssetValueRequest{
				Denom: markerDenom,
				NetAssetValues: []types.NetAssetValue{
					{
						Value:      sdk.NewInt64Coin(valueDenom, 100),
						Volume:     uint64(100),
						Source:     "exchange",
						UpdateTime: s.blockStartTime.Add(10 * time.Hour),
					},
				},
				Administrator: authUser.String(),
			},
			expErr: fmt.Sprintf("net asset value update time (%v) is later than current block time (%v): invalid request", s.blockStartTime.Add(10*time.Hour).UTC(), s.blockStartTime.UTC()),
		},
		{
			name: "successfully set nav",
			msg: types.MsgAddNetAssetValueRequest{
				Denom: markerDenom,
				NetAssetValues: []types.NetAssetValue{
					{
						Value:      sdk.NewInt64Coin(valueDenom, 100),
						Volume:     uint64(100),
						Source:     "exchange",
						UpdateTime: s.blockStartTime.UTC(),
					},
				},
				Administrator: authUser.String(),
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			res, err := s.msgServer.AddNetAssetValue(sdk.WrapSDKContext(s.ctx),
				&tc.msg)

			if len(tc.expErr) > 0 {
				s.Assert().Nil(res)
				s.Assert().EqualError(err, tc.expErr)

			} else {
				s.Assert().NoError(err)
				s.Assert().Equal(res, &types.MsgAddNetAssetValueResponse{})
			}
		})
	}
}

func (s *MsgServerTestSuite) TestDeleteNetAssetValue() {
	authUser := testUserAddress("test")

	markerDenom := "jackthecat"
	markerAcct := authtypes.NewBaseAccount(types.MustGetMarkerAddress(markerDenom), nil, 0, 0)
	s.app.MarkerKeeper.SetMarker(s.ctx, types.NewMarkerAccount(markerAcct, sdk.NewInt64Coin(markerDenom, 1000), authUser, []types.AccessGrant{{Address: authUser.String(), Permissions: []types.Access{types.Access_Transfer}}}, types.StatusProposed, types.MarkerType_RestrictedCoin, true, false, false, []string{}))

	valueDenom := "usd"
	valueAcct := authtypes.NewBaseAccount(types.MustGetMarkerAddress(valueDenom), nil, 0, 0)
	s.app.MarkerKeeper.SetMarker(s.ctx, types.NewMarkerAccount(valueAcct, sdk.NewInt64Coin(valueDenom, 1000), authUser, []types.AccessGrant{{Address: authUser.String(), Permissions: []types.Access{types.Access_Transfer}}}, types.StatusProposed, types.MarkerType_RestrictedCoin, true, false, false, []string{}))

	nav := types.NetAssetValue{
		Value:      sdk.NewInt64Coin(valueDenom, 100),
		Volume:     uint64(100),
		Source:     "exchange",
		UpdateTime: s.blockStartTime,
	}
	s.app.MarkerKeeper.SetNetAssetValue(s.ctx, markerAcct.GetAddress(), nav)

	finalizedMarkerDenom := "finalizedjackthecat"
	finalizedMarkerAcct := authtypes.NewBaseAccount(types.MustGetMarkerAddress(finalizedMarkerDenom), nil, 1, 0)
	s.app.MarkerKeeper.SetMarker(s.ctx, types.NewMarkerAccount(finalizedMarkerAcct, sdk.NewInt64Coin(finalizedMarkerDenom, 1000), authUser, []types.AccessGrant{{Address: authUser.String(), Permissions: []types.Access{types.Access_Transfer}}}, types.StatusFinalized, types.MarkerType_RestrictedCoin, true, false, false, []string{}))

	testCases := []struct {
		name   string
		msg    types.MsgDeleteNetAssetValueRequest
		expErr string
	}{
		{
			name:   "no marker found",
			msg:    types.MsgDeleteNetAssetValueRequest{Denom: "cantfindme", ValueDenoms: []string{valueDenom}, Administrator: authUser.String()},
			expErr: "marker cantfindme not found for address: cosmos17l2yneua2mdfqaycgyhqag8t20asnjwf6adpmt: invalid request",
		},
		{
			name:   "marker is not in proposed state",
			msg:    types.MsgDeleteNetAssetValueRequest{Denom: finalizedMarkerDenom, ValueDenoms: []string{valueDenom}, Administrator: authUser.String()},
			expErr: "can only remove net asset values to markers in the Proposed status: invalid request",
		},
		{
			name:   "marker is not in proposed state",
			msg:    types.MsgDeleteNetAssetValueRequest{Denom: markerDenom, ValueDenoms: []string{"dne"}, Administrator: authUser.String()},
			expErr: "could not remove net asset value : net asset value for cosmos157rf76qwxlttnjyncsaxvelc96m9e5eepfat5g marker address not found for value dne: invalid request",
		},
		{
			name: "successfully remove",
			msg:  types.MsgDeleteNetAssetValueRequest{Denom: markerDenom, ValueDenoms: []string{valueDenom}, Administrator: authUser.String()},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			res, err := s.msgServer.DeleteNetAssetValue(sdk.WrapSDKContext(s.ctx),
				&tc.msg)

			if len(tc.expErr) > 0 {
				s.Assert().Nil(res)
				s.Assert().EqualError(err, tc.expErr)

			} else {
				s.Assert().NoError(err)
				s.Assert().Equal(res, &types.MsgDeleteNetAssetValueResponse{})
			}
		})
	}
}
