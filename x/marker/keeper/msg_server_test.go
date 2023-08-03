package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/stretchr/testify/suite"

	simapp "github.com/provenance-io/provenance/app"
	markerkeeper "github.com/provenance-io/provenance/x/marker/keeper"
	"github.com/provenance-io/provenance/x/marker/types"
)

type MsgServerTestSuite struct {
	suite.Suite

	app       *simapp.App
	ctx       sdk.Context
	msgServer types.MsgServer

	privkey1   cryptotypes.PrivKey
	pubkey1    cryptotypes.PubKey
	owner1     string
	owner1Addr sdk.AccAddress
	acct1      authtypes.AccountI

	addresses []sdk.AccAddress
}

func (s *MsgServerTestSuite) SetupTest() {

	s.app = simapp.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
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
