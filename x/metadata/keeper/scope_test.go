package keeper_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	simapp "github.com/provenance-io/provenance/app"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	"github.com/provenance-io/provenance/x/metadata/keeper"
	"github.com/provenance-io/provenance/x/metadata/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type ScopeKeeperTestSuite struct {
	suite.Suite

	app         *simapp.App
	queryClient types.QueryClient

	pubkey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress

	pubkey2   cryptotypes.PubKey
	user2     string
	user2Addr sdk.AccAddress

	pubkey3   cryptotypes.PubKey
	user3     string
	user3Addr sdk.AccAddress

	scopeUUID uuid.UUID
	scopeID   types.MetadataAddress

	scopeSpecUUID uuid.UUID
	scopeSpecID   types.MetadataAddress
}

func (s *ScopeKeeperTestSuite) SetupTest() {
	s.app = simapp.Setup(s.T())
	ctx := s.FreshCtx()
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, s.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, s.app.MetadataKeeper)
	s.queryClient = types.NewQueryClient(queryHelper)

	s.pubkey1 = secp256k1.GenPrivKey().PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()
	user1Acc := s.app.AccountKeeper.NewAccountWithAddress(ctx, s.user1Addr)
	s.Require().NoError(user1Acc.SetPubKey(s.pubkey1), "SetPubKey user1")
	s.app.AccountKeeper.SetAccount(ctx, user1Acc)

	s.pubkey2 = secp256k1.GenPrivKey().PubKey()
	s.user2Addr = sdk.AccAddress(s.pubkey2.Address())
	s.user2 = s.user2Addr.String()

	s.pubkey3 = secp256k1.GenPrivKey().PubKey()
	s.user3Addr = sdk.AccAddress(s.pubkey3.Address())
	s.user3 = s.user3Addr.String()

	s.scopeUUID = uuid.New()
	s.scopeID = types.ScopeMetadataAddress(s.scopeUUID)

	s.scopeSpecUUID = uuid.New()
	s.scopeSpecID = types.ScopeSpecMetadataAddress(s.scopeSpecUUID)
}

func (s *ScopeKeeperTestSuite) FreshCtx() sdk.Context {
	return keeper.AddAuthzCacheToContext(s.app.BaseApp.NewContext(false, tmproto.Header{}))
}

func TestScopeKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(ScopeKeeperTestSuite))
}

// func ownerPartyList defined in keeper_test.go

type testUser struct {
	PrivKey cryptotypes.PrivKey
	PubKey  cryptotypes.PubKey
	Addr    sdk.AccAddress
	Bech32  string
}

func randomUser() testUser {
	rv := testUser{}
	rv.PrivKey = secp256k1.GenPrivKey()
	rv.PubKey = rv.PrivKey.PubKey()
	rv.Addr = sdk.AccAddress(rv.PubKey.Address())
	rv.Bech32 = rv.Addr.String()
	return rv
}

func (s *ScopeKeeperTestSuite) TestMetadataScopeGetSet() {
	ctx := s.FreshCtx()
	scope, found := s.app.MetadataKeeper.GetScope(ctx, s.scopeID)
	s.Assert().NotNil(scope)
	s.False(found)

	ns := *types.NewScope(s.scopeID, s.scopeSpecID, ownerPartyList(s.user1), []string{s.user1}, s.user1, false)
	s.Assert().NotNil(ns)
	s.app.MetadataKeeper.SetScope(ctx, ns)

	scope, found = s.app.MetadataKeeper.GetScope(ctx, s.scopeID)
	s.Assert().True(found)
	s.Assert().NotNil(scope)

	s.app.MetadataKeeper.RemoveScope(ctx, ns.ScopeId)
	scope, found = s.app.MetadataKeeper.GetScope(ctx, s.scopeID)
	s.Assert().False(found)
	s.Assert().NotNil(scope)
}

func (s *ScopeKeeperTestSuite) TestMetadataScopeIterator() {
	ctx := s.FreshCtx()
	for i := 1; i <= 10; i++ {
		valueOwner := ""
		if i == 5 {
			valueOwner = s.user2
		}
		ns := types.NewScope(types.ScopeMetadataAddress(uuid.New()), nil, ownerPartyList(s.user1), []string{s.user1}, valueOwner, false)
		s.app.MetadataKeeper.SetScope(ctx, *ns)
	}
	count := 0
	err := s.app.MetadataKeeper.IterateScopes(ctx, func(s types.Scope) (stop bool) {
		count++
		return false
	})
	s.Require().NoError(err, "IterateScopes")
	s.Assert().Equal(10, count, "number of scopes iterated")

	count = 0
	err = s.app.MetadataKeeper.IterateScopesForAddress(ctx, s.user1Addr, func(scopeID types.MetadataAddress) (stop bool) {
		count++
		s.True(scopeID.IsScopeAddress())
		return false
	})
	s.Require().NoError(err, "IterateScopesForAddress user1")
	s.Assert().Equal(10, count, "number of scope ids iterated for user1")

	count = 0
	err = s.app.MetadataKeeper.IterateScopesForAddress(ctx, s.user2Addr, func(scopeID types.MetadataAddress) (stop bool) {
		count++
		s.True(scopeID.IsScopeAddress())
		return false
	})
	s.Require().NoError(err, "IterateScopesForAddress user2")
	s.Assert().Equal(1, count, "number of scope ids iterated for user2")

	count = 0
	err = s.app.MetadataKeeper.IterateScopes(ctx, func(s types.Scope) (stop bool) {
		count++
		return count >= 5
	})
	s.Require().NoError(err, "IterateScopes with early stop")
	s.Assert().Equal(5, count, "number of scopes iterated with early stop")
}

func (s *ScopeKeeperTestSuite) TestValidateWriteScope() {
	ns := func(scopeID, scopeSpecification types.MetadataAddress, owners []types.Party, dataAccess []string, valueOwner string) *types.Scope {
		return &types.Scope{
			ScopeId:           scopeID,
			SpecificationId:   scopeSpecification,
			Owners:            owners,
			DataAccess:        dataAccess,
			ValueOwnerAddress: valueOwner,
		}
	}
	rollupScope := func(scopeID, specID types.MetadataAddress, owners []types.Party, valueOwner string) *types.Scope {
		return &types.Scope{
			ScopeId:            scopeID,
			SpecificationId:    specID,
			Owners:             owners,
			DataAccess:         nil,
			ValueOwnerAddress:  valueOwner,
			RequirePartyRollup: true,
		}
	}
	pt := func(addr string, role types.PartyType, opt bool) types.Party {
		return types.Party{
			Address:  addr,
			Role:     role,
			Optional: opt,
		}
	}
	ptz := func(parties ...types.Party) []types.Party {
		rv := make([]types.Party, 0, len(parties))
		rv = append(rv, parties...)
		return rv
	}

	owner := types.PartyType_PARTY_TYPE_OWNER
	affiliate := types.PartyType_PARTY_TYPE_AFFILIATE

	ctx := s.FreshCtx()
	markerAddr := markertypes.MustGetMarkerAddress("testcoin").String()
	err := s.app.MarkerKeeper.AddMarkerAccount(ctx, &markertypes.MarkerAccount{
		BaseAccount: &authtypes.BaseAccount{
			Address:       markerAddr,
			AccountNumber: 23,
		},
		AccessControl: []markertypes.AccessGrant{
			{
				Address:     s.user1,
				Permissions: markertypes.AccessListByNames("deposit,withdraw"),
			},
		},
		Denom:      "testcoin",
		Supply:     sdk.NewInt(1000),
		MarkerType: markertypes.MarkerType_Coin,
		Status:     markertypes.StatusActive,
	})
	s.Require().NoError(err, "AddMarkerAccount")

	scopeSpecID := types.ScopeSpecMetadataAddress(uuid.New())
	scopeSpec := types.NewScopeSpecification(scopeSpecID, nil, []string{s.user1}, []types.PartyType{types.PartyType_PARTY_TYPE_OWNER}, []types.MetadataAddress{})
	s.app.MetadataKeeper.SetScopeSpecification(ctx, *scopeSpec)

	scopeID := types.ScopeMetadataAddress(uuid.New())
	scopeID2 := types.ScopeMetadataAddress(uuid.New())

	// Give user 3 authority to sign for user 1 for scope updates.
	a := authz.NewGenericAuthorization(types.TypeURLMsgWriteScopeRequest)
	s.Require().NoError(s.app.AuthzKeeper.SaveGrant(ctx, s.user3Addr, s.user1Addr, a, nil), "authz SaveGrant user1 to user3")

	otherAddr := sdk.AccAddress("other_address_______").String()

	cases := []struct {
		name     string
		existing *types.Scope
		proposed types.Scope
		signers  []string
		errorMsg string
	}{
		{
			name:     "nil previous, proposed throws address error",
			existing: nil,
			proposed: types.Scope{},
			signers:  []string{s.user1},
			errorMsg: "address is empty",
		},
		{
			name:     "valid proposed with nil existing doesn't error",
			existing: nil,
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			signers:  []string{s.user1},
			errorMsg: "",
		},
		{
			name:     "can't change scope id in update",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			proposed: *ns(scopeID2, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("cannot update scope identifier. expected %s, got %s", scopeID.String(), scopeID2.String()),
		},
		{
			name:     "missing existing owner signer on update fails",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{s.user1}, ""),
			signers:  []string{s.user2},
			errorMsg: fmt.Sprintf("missing signature: %s", s.user1),
		},
		{
			name:     "missing existing owner signer on update fails",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user2), []string{}, ""),
			signers:  []string{s.user2},
			errorMsg: fmt.Sprintf("missing signature: %s", s.user1),
		},
		{
			name:     "no error when update includes existing owner signer",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{s.user1}, ""),
			signers:  []string{s.user1},
			errorMsg: "",
		},
		{
			name:     "no error when there are no updates regardless of signatures",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			signers:  []string{},
			errorMsg: "",
		},
		{
			name:     "setting value owner when unset does not error",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			signers:  []string{s.user1},
			errorMsg: "",
		},
		{
			name:     "setting value owner when unset requires current owner signature",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			signers:  []string{},
			errorMsg: fmt.Sprintf("missing signature: %s", s.user1),
		},
		{
			name:     "setting value owner to user does not require their signature",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user2),
			signers:  []string{s.user1},
			errorMsg: "",
		},
		{
			name:     "setting value owner to new user does not require their signature",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user2),
			signers:  []string{s.user1},
			errorMsg: "",
		},
		{
			name:     "no change to value owner should not error",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			signers:  []string{s.user1},
			errorMsg: "",
		},
		{
			name:     "setting a new value owner should not error with withdraw permission",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, markerAddr),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			signers:  []string{s.user1},
			errorMsg: "",
		},
		{
			name:     "with rollup setting a new value owner should not error with withdraw permission",
			existing: rollupScope(scopeID, scopeSpecID, ownerPartyList(s.user1), markerAddr),
			proposed: *rollupScope(scopeID, scopeSpecID, ownerPartyList(s.user1), s.user1),
			signers:  []string{s.user1},
			errorMsg: "",
		},
		{
			name:     "setting a new value owner fails if missing withdraw permission",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user2), []string{}, markerAddr),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user2), []string{}, s.user2),
			signers:  []string{s.user2},
			errorMsg: fmt.Sprintf("missing signature for %s (testcoin) with authority to withdraw/remove it as scope value owner", markerAddr),
		},
		{
			name:     "with rollup setting a new value owner fails if missing withdraw permission",
			existing: rollupScope(scopeID, scopeSpecID, ownerPartyList(s.user2), markerAddr),
			proposed: *rollupScope(scopeID, scopeSpecID, ownerPartyList(s.user2), s.user2),
			signers:  []string{s.user2},
			errorMsg: fmt.Sprintf("missing signature for %s (testcoin) with authority to withdraw/remove it as scope value owner", markerAddr),
		},
		{
			name:     "setting a new value owner fails if missing deposit permission",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user2), []string{}, ""),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user2), []string{}, markerAddr),
			signers:  []string{s.user2},
			errorMsg: fmt.Sprintf("missing signature for %s (testcoin) with authority to deposit/add it as scope value owner", markerAddr),
		},
		{
			name:     "with rollup setting a new value owner fails if missing deposit permission",
			existing: rollupScope(scopeID, scopeSpecID, ownerPartyList(s.user2), ""),
			proposed: *rollupScope(scopeID, scopeSpecID, ownerPartyList(s.user2), markerAddr),
			signers:  []string{s.user2},
			errorMsg: fmt.Sprintf("missing signature for %s (testcoin) with authority to deposit/add it as scope value owner", markerAddr),
		},
		{
			name:     "setting a new value owner fails for scope owner when value owner signature is missing",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user2),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("missing signature from existing value owner %s", s.user2),
		},
		{
			name:     "with rollup setting a new value owner fails for scope owner when value owner signature is missing",
			existing: rollupScope(scopeID, scopeSpecID, ownerPartyList(s.user1), s.user2),
			proposed: *rollupScope(scopeID, scopeSpecID, ownerPartyList(s.user1), s.user1),
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("missing signature from existing value owner %s", s.user2),
		},
		{
			name:     "changing only value owner only requires value owner sig",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1, s.user2), []string{}, otherAddr),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1, s.user2), []string{}, s.user1),
			signers:  []string{otherAddr},
			errorMsg: "",
		},
		{
			name:     "with rollup changing only value owner only requires value owner sig",
			existing: rollupScope(scopeID, scopeSpecID, ownerPartyList(s.user1, s.user2), otherAddr),
			proposed: *rollupScope(scopeID, scopeSpecID, ownerPartyList(s.user1, s.user2), s.user1),
			signers:  []string{otherAddr},
			errorMsg: "",
		},
		{
			name:     "unsetting all fields on a scope should be successful",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			proposed: types.Scope{ScopeId: scopeID, SpecificationId: scopeSpecID, Owners: ownerPartyList(s.user1)},
			signers:  []string{s.user1},
			errorMsg: "",
		},
		{
			name:     "setting specification id to nil should fail",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			proposed: *ns(scopeID, nil, ownerPartyList(s.user1), []string{}, s.user1),
			signers:  []string{s.user1},
			errorMsg: "invalid specification id: address is empty",
		},
		{
			name:     "setting unknown specification id should fail",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			proposed: *ns(scopeID, types.ScopeSpecMetadataAddress(s.scopeUUID), ownerPartyList(s.user1), []string{}, s.user1),
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("scope specification %s not found", types.ScopeSpecMetadataAddress(s.scopeUUID)),
		},
		{
			name:     "adding data access with authz grant should be successful",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{s.user2}, s.user1),
			signers:  []string{s.user3}, // user 1 has granted scope-write to user 3
			errorMsg: "",
		},
		{
			name:     "multi owner adding data access with authz grant should be successful",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1, s.user2), []string{}, s.user1),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1, s.user2), []string{s.user2}, s.user1),
			signers:  []string{s.user2, s.user3}, // user 1 has granted scope-write to user 3
			errorMsg: "",
		},
		{
			name:     "changing value owner with authz grant should be successful",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user2),
			signers:  []string{s.user3}, // user 1 has granted scope-write to user 3
			errorMsg: "",
		},
		{
			name:     "changing value owner by authz granter should be successful",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user2),
			signers:  []string{s.user1},
			errorMsg: "",
		},
		{
			name:     "changing value owner by non-authz grantee should fail",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user2),
			signers:  []string{s.user2},
			errorMsg: fmt.Sprintf("missing signature from existing value owner %s", s.user1),
		},
		{
			name:     "changing value owner from non-authz granter with different signer should fail",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user2),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			signers:  []string{s.user3},
			errorMsg: fmt.Sprintf("missing signature from existing value owner %s", s.user2),
		},
		{
			name:     "setting value owner from nothing to non-owner only signed by non-owner should fail",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user2),
			signers:  []string{s.user2},
			errorMsg: fmt.Sprintf("missing signature: %s", s.user1),
		},
		{
			name:     "with rollup scope missing req role",
			existing: nil,
			proposed: *rollupScope(scopeID, scopeSpecID, ptz(pt(s.user1, affiliate, false)), ""),
			signers:  nil,
			errorMsg: "missing roles required by spec: OWNER need 1 have 0",
		},
		{
			name:     "with rollup without existing but has req role and signer not involved in scope",
			existing: nil,
			proposed: *rollupScope(scopeID, scopeSpecID, ownerPartyList(s.user1), ""),
			signers:  []string{otherAddr},
			errorMsg: "",
		},
		{
			name:     "with rollup existing required owner is not signer",
			existing: rollupScope(scopeID, scopeSpecID, ownerPartyList(s.user1), ""),
			proposed: *rollupScope(scopeID, scopeSpecID, ownerPartyList(s.user1, s.user2), ""),
			signers:  []string{s.user2},
			errorMsg: "missing required signature: " + s.user1 + " (OWNER)",
		},
		{
			name:     "with rollup missing signer from required role",
			existing: rollupScope(scopeID, scopeSpecID, ptz(pt(s.user1, owner, true), pt(s.user2, affiliate, false)), ""),
			proposed: *rollupScope(scopeID, scopeSpecID, ptz(pt(s.user1, owner, true), pt(s.user2, affiliate, false), pt(s.user2, types.PartyType_PARTY_TYPE_OWNER, true)), ""),
			signers:  []string{s.user2},
			errorMsg: "missing signers for roles required by spec: OWNER need 1 have 0",
		},
		{
			name:     "with rollup two optional owners one signs",
			existing: rollupScope(scopeID, scopeSpecID, ptz(pt(s.user1, owner, true), pt(s.user2, owner, true)), ""),
			proposed: *rollupScope(scopeID, scopeSpecID, ptz(pt(s.user2, owner, true)), ""),
			signers:  []string{s.user2},
			errorMsg: "",
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			msg := &types.MsgWriteScopeRequest{
				Scope:   tc.proposed,
				Signers: tc.signers,
			}
			err = s.app.MetadataKeeper.ValidateWriteScope(s.FreshCtx(), tc.existing, msg)
			if len(tc.errorMsg) > 0 {
				s.Assert().EqualError(err, tc.errorMsg, "ValidateWriteScope expected error")
			} else {
				s.Assert().NoError(err, "ValidateWriteScope unexpected error")
			}
		})
	}
}

func (s *ScopeKeeperTestSuite) TestValidateDeleteScope() {
	pt := func(addr string, role types.PartyType, opt bool) types.Party {
		return types.Party{
			Address:  addr,
			Role:     role,
			Optional: opt,
		}
	}
	ptz := func(parties ...types.Party) []types.Party {
		rv := make([]types.Party, 0, len(parties))
		rv = append(rv, parties...)
		return rv
	}

	ctx := s.FreshCtx()
	markerDenom := "testcoins2"
	markerAddr := markertypes.MustGetMarkerAddress(markerDenom).String()
	err := s.app.MarkerKeeper.AddMarkerAccount(ctx, &markertypes.MarkerAccount{
		BaseAccount: &authtypes.BaseAccount{
			Address:       markerAddr,
			AccountNumber: 24,
		},
		AccessControl: []markertypes.AccessGrant{
			{
				Address:     s.user1,
				Permissions: markertypes.AccessListByNames("deposit,withdraw"),
			},
		},
		Denom:      markerDenom,
		Supply:     sdk.NewInt(1000),
		MarkerType: markertypes.MarkerType_Coin,
		Status:     markertypes.StatusActive,
	})
	s.Require().NoError(err, "AddMarkerAccount")

	scopeNoValueOwner := types.Scope{
		ScopeId:           types.ScopeMetadataAddress(uuid.New()),
		SpecificationId:   types.ScopeSpecMetadataAddress(uuid.New()),
		Owners:            ownerPartyList(s.user1, s.user2),
		DataAccess:        nil,
		ValueOwnerAddress: "",
	}
	s.app.MetadataKeeper.SetScope(ctx, scopeNoValueOwner)

	scopeMarkerValueOwner := types.Scope{
		ScopeId:           types.ScopeMetadataAddress(uuid.New()),
		SpecificationId:   types.ScopeSpecMetadataAddress(uuid.New()),
		Owners:            ownerPartyList(s.user2),
		DataAccess:        nil,
		ValueOwnerAddress: markerAddr,
	}
	s.app.MetadataKeeper.SetScope(ctx, scopeMarkerValueOwner)

	scopeUserValueOwner := types.Scope{
		ScopeId:           types.ScopeMetadataAddress(uuid.New()),
		SpecificationId:   types.ScopeSpecMetadataAddress(uuid.New()),
		Owners:            ownerPartyList(s.user2),
		DataAccess:        nil,
		ValueOwnerAddress: s.user1,
	}
	s.app.MetadataKeeper.SetScope(ctx, scopeUserValueOwner)

	owner := types.PartyType_PARTY_TYPE_OWNER
	servicer := types.PartyType_PARTY_TYPE_SERVICER

	scopeSpec := types.ScopeSpecification{
		SpecificationId: types.ScopeSpecMetadataAddress(uuid.New()),
		Description:     types.NewDescription("tester", "test scope spec", "", ""),
		OwnerAddresses:  []string{s.user1},
		PartiesInvolved: []types.PartyType{owner, servicer},
		ContractSpecIds: []types.MetadataAddress{types.ContractSpecMetadataAddress(uuid.New())},
	}
	s.app.MetadataKeeper.SetScopeSpecification(ctx, scopeSpec)

	otherUser := sdk.AccAddress("some_other_user_____").String()

	// with rollup no scope spec req party not signed
	scopeRollupNoSpecReq := types.Scope{
		ScopeId:            types.ScopeMetadataAddress(uuid.New()),
		SpecificationId:    types.ScopeSpecMetadataAddress(uuid.New()),
		Owners:             ptz(pt(s.user1, owner, false), pt(s.user2, servicer, false), pt(otherUser, owner, true)),
		DataAccess:         nil,
		ValueOwnerAddress:  "",
		RequirePartyRollup: true,
	}
	s.app.MetadataKeeper.SetScope(ctx, scopeRollupNoSpecReq)

	// with rollup no scope spec all optional parties signer not involved
	scopeRollupNoSpecAllOpt := types.Scope{
		ScopeId:            types.ScopeMetadataAddress(uuid.New()),
		SpecificationId:    types.ScopeSpecMetadataAddress(uuid.New()),
		Owners:             ptz(pt(s.user1, owner, true), pt(s.user2, servicer, true)),
		DataAccess:         nil,
		ValueOwnerAddress:  "",
		RequirePartyRollup: true,
	}
	s.app.MetadataKeeper.SetScope(ctx, scopeRollupNoSpecAllOpt)

	// with rollup req scope owner not signed
	// with rollup req role not signed
	// with rollup req scope owner and req role signed.
	scopeRollup := types.Scope{
		ScopeId:            types.ScopeMetadataAddress(uuid.New()),
		SpecificationId:    scopeSpec.SpecificationId,
		Owners:             ptz(pt(s.user1, owner, false), pt(s.user2, servicer, true), pt(otherUser, owner, true)),
		DataAccess:         nil,
		ValueOwnerAddress:  "",
		RequirePartyRollup: true,
	}
	s.app.MetadataKeeper.SetScope(ctx, scopeRollup)

	// with rollup marker value owner no signer has withdraw
	scopeRollupMarkerValueOwner := types.Scope{
		ScopeId:           types.ScopeMetadataAddress(uuid.New()),
		SpecificationId:   types.ScopeSpecMetadataAddress(uuid.New()),
		Owners:            ownerPartyList(s.user2),
		DataAccess:        nil,
		ValueOwnerAddress: markerAddr,
	}
	s.app.MetadataKeeper.SetScope(ctx, scopeRollupMarkerValueOwner)

	// with rollup value owner not signed
	scopeRollupUserValueOwner := types.Scope{
		ScopeId:           types.ScopeMetadataAddress(uuid.New()),
		SpecificationId:   types.ScopeSpecMetadataAddress(uuid.New()),
		Owners:            ownerPartyList(s.user2),
		DataAccess:        nil,
		ValueOwnerAddress: s.user1,
	}
	s.app.MetadataKeeper.SetScope(ctx, scopeRollupUserValueOwner)

	dneScopeID := types.ScopeMetadataAddress(uuid.New())

	missing1Sig := func(addr string) string {
		return fmt.Sprintf("missing signature: %s", addr)
	}

	missing2Sigs := func(addr1, addr2 string) string {
		return fmt.Sprintf("missing signatures: %s, %s", addr1, addr2)
	}

	tests := []struct {
		name     string
		scope    types.Scope
		signers  []string
		expected string
	}{
		{
			name:     "no value owner all signers",
			scope:    scopeNoValueOwner,
			signers:  []string{s.user1, s.user2},
			expected: "",
		},
		{
			name:     "no value owner all signers reversed",
			scope:    scopeNoValueOwner,
			signers:  []string{s.user1, s.user2},
			expected: "",
		},
		{
			name:     "no value owner extra signer",
			scope:    scopeNoValueOwner,
			signers:  []string{s.user1, s.user2, s.user3},
			expected: "",
		},
		{
			name:     "no value owner missing signer 1",
			scope:    scopeNoValueOwner,
			signers:  []string{s.user2},
			expected: missing1Sig(s.user1),
		},
		{
			name:     "no value owner missing signer 2",
			scope:    scopeNoValueOwner,
			signers:  []string{s.user1},
			expected: missing1Sig(s.user2),
		},
		{
			name:     "no value owner no signers",
			scope:    scopeNoValueOwner,
			signers:  []string{},
			expected: missing2Sigs(s.user1, s.user2),
		},
		{
			name:     "no value owner wrong signer",
			scope:    scopeNoValueOwner,
			signers:  []string{s.user3},
			expected: missing2Sigs(s.user1, s.user2),
		},
		{
			name:     "marker value owner signed by owner and user with auth",
			scope:    scopeMarkerValueOwner,
			signers:  []string{s.user1, s.user2},
			expected: "",
		},
		{
			name:     "marker value owner signed by owner and user with auth reversed",
			scope:    scopeMarkerValueOwner,
			signers:  []string{s.user2, s.user1},
			expected: "",
		},
		{
			name:     "marker value owner not signed by owner",
			scope:    scopeMarkerValueOwner,
			signers:  []string{s.user1},
			expected: missing1Sig(s.user2),
		},
		{
			name:     "marker value owner not signed by user with auth",
			scope:    scopeMarkerValueOwner,
			signers:  []string{s.user2},
			expected: fmt.Sprintf("missing signature for %s (testcoins2) with authority to withdraw/remove it as scope value owner", markerAddr),
		},
		{
			name:     "user value owner signed by owner and value owner",
			scope:    scopeUserValueOwner,
			signers:  []string{s.user1, s.user2},
			expected: "",
		},
		{
			name:     "user value owner signed by owner and value owner reversed",
			scope:    scopeUserValueOwner,
			signers:  []string{s.user2, s.user1},
			expected: "",
		},
		{
			name:     "user value owner not signed by owner",
			scope:    scopeUserValueOwner,
			signers:  []string{s.user1},
			expected: missing1Sig(s.user2),
		},
		{
			name:     "user value owner not signed by value owner",
			scope:    scopeUserValueOwner,
			signers:  []string{s.user2},
			expected: fmt.Sprintf("missing signature from existing value owner %s", s.user1),
		},
		{
			name:     "scope does not exist",
			scope:    types.Scope{ScopeId: dneScopeID},
			signers:  []string{},
			expected: fmt.Sprintf("scope not found with id %s", dneScopeID),
		},
		{
			name:     "with rollup no scope spec neither req party signed",
			scope:    scopeRollupNoSpecReq,
			signers:  []string{otherUser},
			expected: "missing signatures: " + s.user1 + ", " + s.user2 + "",
		},
		{
			name:     "with rollup no scope spec req party 1 not signed",
			scope:    scopeRollupNoSpecReq,
			signers:  []string{s.user2},
			expected: "missing signature: " + s.user1,
		},
		{
			name:     "with rollup no scope spec req party 2 not signed",
			scope:    scopeRollupNoSpecReq,
			signers:  []string{s.user1},
			expected: "missing signature: " + s.user2,
		},
		{
			name:     "with rollup no scope spec both req parties signed",
			scope:    scopeRollupNoSpecReq,
			signers:  []string{s.user1, s.user2},
			expected: "",
		},
		{
			name:     "with rollup no scope spec all optional parties signer not involved",
			scope:    scopeRollupNoSpecAllOpt,
			signers:  []string{otherUser},
			expected: "",
		},
		{
			name:     "with rollup req scope owner not signed",
			scope:    scopeRollup,
			signers:  []string{s.user2, otherUser},
			expected: "missing required signature: " + s.user1 + " (OWNER)",
		},
		{
			name:     "with rollup req role not signed",
			scope:    scopeRollup,
			signers:  []string{s.user1},
			expected: "missing signers for roles required by spec: SERVICER need 1 have 0",
		},
		{
			name:     "with rollup req scope owner and req roles signed",
			scope:    scopeRollup,
			signers:  []string{s.user1, s.user2},
			expected: "",
		},
		{
			name:     "with rollup marker value owner no signer has withdraw",
			scope:    scopeRollupMarkerValueOwner,
			signers:  []string{s.user2},
			expected: "missing signature for " + markerAddr + " (testcoins2) with authority to withdraw/remove it as scope value owner",
		},
		{
			name:     "with rollup marker value owner signer has withdraw",
			scope:    scopeRollupMarkerValueOwner,
			signers:  []string{s.user1, s.user2},
			expected: "",
		},
		{
			name:     "with rollup value owner not signed",
			scope:    scopeRollupUserValueOwner,
			signers:  []string{s.user2},
			expected: "missing signature from existing value owner " + s.user1,
		},
		{
			name:     "with rollup value owner signed",
			scope:    scopeRollupUserValueOwner,
			signers:  []string{s.user1, s.user2},
			expected: "",
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			msg := &types.MsgDeleteScopeRequest{
				ScopeId: tc.scope.ScopeId,
				Signers: tc.signers,
			}
			actual := s.app.MetadataKeeper.ValidateDeleteScope(s.FreshCtx(), msg)
			if len(tc.expected) > 0 {
				require.EqualError(t, actual, tc.expected)
			} else {
				require.NoError(t, actual)
			}
		})
	}
}

func (s *ScopeKeeperTestSuite) TestValidateScopeAddDataAccess() {
	pt := func(addr string, role types.PartyType, opt bool) types.Party {
		return types.Party{
			Address:  addr,
			Role:     role,
			Optional: opt,
		}
	}
	ptz := func(parties ...types.Party) []types.Party {
		rv := make([]types.Party, 0, len(parties))
		rv = append(rv, parties...)
		return rv
	}

	scope := types.Scope{
		ScopeId:           s.scopeID,
		SpecificationId:   types.ScopeSpecMetadataAddress(s.scopeUUID),
		Owners:            ownerPartyList(s.user1),
		DataAccess:        []string{s.user1},
		ValueOwnerAddress: s.user1,
	}

	owner := types.PartyType_PARTY_TYPE_OWNER
	controller := types.PartyType_PARTY_TYPE_CONTROLLER

	scopeSpec := types.ScopeSpecification{
		SpecificationId: types.ScopeSpecMetadataAddress(uuid.New()),
		Description:     types.NewDescription("tester", "test description", "", ""),
		OwnerAddresses:  []string{s.user1},
		PartiesInvolved: []types.PartyType{owner, controller},
		ContractSpecIds: []types.MetadataAddress{types.ContractSpecMetadataAddress(uuid.New())},
	}
	s.app.MetadataKeeper.SetScopeSpecification(s.FreshCtx(), scopeSpec)

	dneSpecID := types.ScopeSpecMetadataAddress(uuid.New())

	otherAddr := sdk.AccAddress("blah_blah_blah_blah_").String()

	cases := []struct {
		name            string
		dataAccessAddrs []string
		existing        types.Scope
		signers         []string
		errorMsg        string
	}{
		{
			name:            "should fail to validate add scope data access, does not have any users",
			dataAccessAddrs: []string{},
			existing:        scope,
			signers:         []string{s.user1},
			errorMsg:        "data access list cannot be empty",
		},
		{
			name:            "should fail to validate add scope data access, user is already on the data access list",
			dataAccessAddrs: []string{s.user1},
			existing:        scope,
			signers:         []string{s.user1},
			errorMsg:        fmt.Sprintf("address already exists for data access %s", s.user1),
		},
		{
			name:            "should fail to validate add scope data access, incorrect signer for scope",
			dataAccessAddrs: []string{s.user2},
			existing:        scope,
			signers:         []string{s.user2},
			errorMsg:        fmt.Sprintf("missing signature: %s", s.user1),
		},
		{
			name:            "should fail to validate add scope data access, incorrect address type",
			dataAccessAddrs: []string{"invalidaddr"},
			existing:        scope,
			signers:         []string{s.user1},
			errorMsg:        "failed to decode data access address invalidaddr : decoding bech32 failed: invalid separator index -1",
		},
		{
			name:            "should successfully validate add scope data access",
			dataAccessAddrs: []string{s.user2},
			existing:        scope,
			signers:         []string{s.user1},
			errorMsg:        "",
		},
		{
			name:            "with rollup spec found signed correctly with opt addr 1",
			dataAccessAddrs: []string{otherAddr},
			existing: types.Scope{
				ScopeId:            types.ScopeMetadataAddress(uuid.New()),
				SpecificationId:    scopeSpec.SpecificationId,
				Owners:             ptz(pt(s.user1, owner, false), pt(s.user2, controller, true), pt(s.user3, controller, true)),
				DataAccess:         nil,
				RequirePartyRollup: true,
			},
			signers:  []string{s.user1, s.user2},
			errorMsg: "",
		},
		{
			name:            "with rollup spec found signed correctly with opt addr 2",
			dataAccessAddrs: []string{otherAddr},
			existing: types.Scope{
				ScopeId:            types.ScopeMetadataAddress(uuid.New()),
				SpecificationId:    scopeSpec.SpecificationId,
				Owners:             ptz(pt(s.user1, owner, false), pt(s.user2, controller, true), pt(s.user3, controller, true)),
				DataAccess:         nil,
				RequirePartyRollup: true,
			},
			signers:  []string{s.user1, s.user3},
			errorMsg: "",
		},
		{
			name:            "with rollup spec not found",
			dataAccessAddrs: []string{otherAddr},
			existing: types.Scope{
				ScopeId:            types.ScopeMetadataAddress(uuid.New()),
				SpecificationId:    dneSpecID,
				Owners:             ptz(pt(s.user1, owner, false), pt(s.user2, controller, true), pt(s.user3, controller, true)),
				DataAccess:         nil,
				RequirePartyRollup: true,
			},
			signers:  []string{s.user1, s.user2},
			errorMsg: "scope specification " + dneSpecID.String() + " not found",
		},
		{
			name:            "with rollup req party not signed",
			dataAccessAddrs: []string{otherAddr},
			existing: types.Scope{
				ScopeId:         types.ScopeMetadataAddress(uuid.New()),
				SpecificationId: scopeSpec.SpecificationId,
				Owners: ptz(pt(s.user1, owner, false), pt(s.user2, controller, true),
					pt(s.user3, owner, true), pt(s.user3, controller, true)),
				DataAccess:         nil,
				RequirePartyRollup: true,
			},
			signers:  []string{s.user2, s.user3},
			errorMsg: "missing required signature: " + s.user1 + " (OWNER)",
		},
		{
			name:            "with rollup req role not signed",
			dataAccessAddrs: []string{otherAddr},
			existing: types.Scope{
				ScopeId:         types.ScopeMetadataAddress(uuid.New()),
				SpecificationId: scopeSpec.SpecificationId,
				Owners: ptz(pt(s.user1, owner, false), pt(s.user2, controller, true),
					pt(s.user3, owner, true), pt(s.user3, controller, true)),
				DataAccess:         nil,
				RequirePartyRollup: true,
			},
			signers:  []string{s.user1},
			errorMsg: "missing signers for roles required by spec: CONTROLLER need 1 have 0",
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			msg := &types.MsgAddScopeDataAccessRequest{
				DataAccess: tc.dataAccessAddrs,
				Signers:    tc.signers,
			}
			err := s.app.MetadataKeeper.ValidateAddScopeDataAccess(s.FreshCtx(), tc.existing, msg)
			if len(tc.errorMsg) > 0 {
				s.Assert().EqualError(err, tc.errorMsg, "ValidateAddScopeDataAccess")
			} else {
				s.Assert().NoError(err, tc.errorMsg, "ValidateAddScopeDataAccess")
			}
		})
	}
}

func (s *ScopeKeeperTestSuite) TestValidateScopeDeleteDataAccess() {
	pt := func(addr string, role types.PartyType, opt bool) types.Party {
		return types.Party{
			Address:  addr,
			Role:     role,
			Optional: opt,
		}
	}
	ptz := func(parties ...types.Party) []types.Party {
		rv := make([]types.Party, 0, len(parties))
		rv = append(rv, parties...)
		return rv
	}

	scope := types.Scope{
		ScopeId:           s.scopeID,
		SpecificationId:   types.ScopeSpecMetadataAddress(s.scopeUUID),
		Owners:            ownerPartyList(s.user1),
		DataAccess:        []string{s.user1, s.user2},
		ValueOwnerAddress: s.user1,
	}

	owner := types.PartyType_PARTY_TYPE_OWNER
	controller := types.PartyType_PARTY_TYPE_CONTROLLER

	scopeSpec := types.ScopeSpecification{
		SpecificationId: types.ScopeSpecMetadataAddress(uuid.New()),
		Description:     types.NewDescription("tester", "test description", "", ""),
		OwnerAddresses:  []string{s.user1},
		PartiesInvolved: []types.PartyType{owner, controller},
		ContractSpecIds: []types.MetadataAddress{types.ContractSpecMetadataAddress(uuid.New())},
	}
	s.app.MetadataKeeper.SetScopeSpecification(s.FreshCtx(), scopeSpec)

	dneSpecID := types.ScopeSpecMetadataAddress(uuid.New())

	otherAddr := sdk.AccAddress("blah_blah_blah_blah_").String()

	cases := []struct {
		name            string
		dataAccessAddrs []string
		existing        types.Scope
		signers         []string
		errorMsg        string
	}{
		{
			name:            "should fail to validate delete scope data access, does not have any users",
			dataAccessAddrs: []string{},
			existing:        scope,
			signers:         []string{s.user1},
			errorMsg:        "data access list cannot be empty",
		},
		{
			name:            "should fail to validate delete scope data access, address is not in data access list",
			dataAccessAddrs: []string{s.user2, s.user3},
			existing:        scope,
			signers:         []string{s.user1},
			errorMsg:        fmt.Sprintf("address does not exist in scope data access: %s", s.user3),
		},
		{
			name:            "should fail to validate delete scope data access, incorrect signer for scope",
			dataAccessAddrs: []string{s.user2},
			existing:        scope,
			signers:         []string{s.user2},
			errorMsg:        fmt.Sprintf("missing signature: %s", s.user1),
		},
		{
			name:            "should successfully validate delete scope data access",
			dataAccessAddrs: []string{s.user1, s.user2},
			existing:        scope,
			signers:         []string{s.user1},
			errorMsg:        "",
		},
		{
			name:            "with rollup spec found signed correctly with opt addr 1",
			dataAccessAddrs: []string{otherAddr},
			existing: types.Scope{
				ScopeId:            types.ScopeMetadataAddress(uuid.New()),
				SpecificationId:    scopeSpec.SpecificationId,
				Owners:             ptz(pt(s.user1, owner, false), pt(s.user2, controller, true), pt(s.user3, controller, true)),
				DataAccess:         []string{otherAddr},
				RequirePartyRollup: true,
			},
			signers:  []string{s.user1, s.user2},
			errorMsg: "",
		},
		{
			name:            "with rollup spec found signed correctly with opt addr 2",
			dataAccessAddrs: []string{otherAddr},
			existing: types.Scope{
				ScopeId:            types.ScopeMetadataAddress(uuid.New()),
				SpecificationId:    scopeSpec.SpecificationId,
				Owners:             ptz(pt(s.user1, owner, false), pt(s.user2, controller, true), pt(s.user3, controller, true)),
				DataAccess:         []string{otherAddr},
				RequirePartyRollup: true,
			},
			signers:  []string{s.user1, s.user3},
			errorMsg: "",
		},
		{
			name:            "with rollup spec not found",
			dataAccessAddrs: []string{otherAddr},
			existing: types.Scope{
				ScopeId:            types.ScopeMetadataAddress(uuid.New()),
				SpecificationId:    dneSpecID,
				Owners:             ptz(pt(s.user1, owner, false), pt(s.user2, controller, true), pt(s.user3, controller, true)),
				DataAccess:         []string{otherAddr},
				RequirePartyRollup: true,
			},
			signers:  []string{s.user1, s.user2},
			errorMsg: "scope specification " + dneSpecID.String() + " not found",
		},
		{
			name:            "with rollup req party not signed",
			dataAccessAddrs: []string{otherAddr},
			existing: types.Scope{
				ScopeId:         types.ScopeMetadataAddress(uuid.New()),
				SpecificationId: scopeSpec.SpecificationId,
				Owners: ptz(pt(s.user1, owner, false), pt(s.user2, controller, true),
					pt(s.user3, owner, true), pt(s.user3, controller, true)),
				DataAccess:         []string{otherAddr},
				RequirePartyRollup: true,
			},
			signers:  []string{s.user2, s.user3},
			errorMsg: "missing required signature: " + s.user1 + " (OWNER)",
		},
		{
			name:            "with rollup req role not signed",
			dataAccessAddrs: []string{otherAddr},
			existing: types.Scope{
				ScopeId:         types.ScopeMetadataAddress(uuid.New()),
				SpecificationId: scopeSpec.SpecificationId,
				Owners: ptz(pt(s.user1, owner, false), pt(s.user2, controller, true),
					pt(s.user3, owner, true), pt(s.user3, controller, true)),
				DataAccess:         []string{otherAddr},
				RequirePartyRollup: true,
			},
			signers:  []string{s.user1},
			errorMsg: "missing signers for roles required by spec: CONTROLLER need 1 have 0",
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			msg := &types.MsgDeleteScopeDataAccessRequest{
				DataAccess: tc.dataAccessAddrs,
				Signers:    tc.signers,
			}
			err := s.app.MetadataKeeper.ValidateDeleteScopeDataAccess(s.FreshCtx(), tc.existing, msg)
			if len(tc.errorMsg) > 0 {
				s.Assert().EqualError(err, tc.errorMsg, "ValidateDeleteScopeDataAccess")
			} else {
				s.Assert().NoError(err, "ValidateDeleteScopeDataAccess")
			}
		})
	}
}

func (s *ScopeKeeperTestSuite) TestValidateScopeUpdateOwners() {
	pt := func(addr string, role types.PartyType, opt bool) types.Party {
		return types.Party{Address: addr, Role: role, Optional: opt}
	}
	ctx := s.FreshCtx()
	scopeSpecID := types.ScopeSpecMetadataAddress(uuid.New())
	scopeSpec := types.NewScopeSpecification(scopeSpecID, nil, []string{s.user1}, []types.PartyType{types.PartyType_PARTY_TYPE_OWNER}, []types.MetadataAddress{})
	s.app.MetadataKeeper.SetScopeSpecification(ctx, *scopeSpec)

	scopeWithOwners := func(owners []types.Party) types.Scope {
		return types.Scope{
			ScopeId:           s.scopeID,
			SpecificationId:   scopeSpecID,
			Owners:            owners,
			DataAccess:        []string{s.user1},
			ValueOwnerAddress: s.user1,
		}
	}
	rollupScopeWithOwners := func(owners ...types.Party) types.Scope {
		return types.Scope{
			ScopeId:            s.scopeID,
			SpecificationId:    scopeSpecID,
			Owners:             owners,
			DataAccess:         []string{s.user1},
			ValueOwnerAddress:  s.user1,
			RequirePartyRollup: true,
		}
	}

	originalOwners := ownerPartyList(s.user1)
	dneScopeSpecID := types.ScopeSpecMetadataAddress(uuid.New())

	owner := types.PartyType_PARTY_TYPE_OWNER
	omnibus := types.PartyType_PARTY_TYPE_OMNIBUS

	testCases := []struct {
		name     string
		existing types.Scope
		proposed types.Scope
		signers  []string
		errorMsg string
	}{
		{
			name:     "should fail to validate update scope owners, fail to decode address",
			existing: scopeWithOwners(originalOwners),
			proposed: scopeWithOwners([]types.Party{{Address: "shoulderror", Role: types.PartyType_PARTY_TYPE_AFFILIATE}}),
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("invalid scope owners: invalid party address [%s]: %s", "shoulderror", "decoding bech32 failed: invalid separator index -1"),
		},
		{
			name:     "should fail to validate update scope owners, role cannot be unspecified",
			existing: scopeWithOwners(originalOwners),
			proposed: scopeWithOwners([]types.Party{{Address: s.user1, Role: types.PartyType_PARTY_TYPE_UNSPECIFIED}}),
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("invalid scope owners: invalid party type for party %s", s.user1),
		},
		{
			name:     "should fail to validate update scope owner, wrong signer new owner",
			existing: scopeWithOwners(originalOwners),
			proposed: scopeWithOwners([]types.Party{{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER}}),
			signers:  []string{s.user2},
			errorMsg: fmt.Sprintf("missing signature: %s", s.user1),
		},
		{
			name:     "should successfully validate update scope owner, same owner different role",
			existing: scopeWithOwners(ownerPartyList(s.user1, s.user2)),
			proposed: scopeWithOwners([]types.Party{
				{Address: s.user1, Role: types.PartyType_PARTY_TYPE_CUSTODIAN},
				{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER},
			}),
			signers:  []string{s.user1, s.user2},
			errorMsg: "",
		},
		{
			name:     "should successfully validate update scope owner, new owner",
			existing: scopeWithOwners(originalOwners),
			proposed: scopeWithOwners([]types.Party{{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER}}),
			signers:  []string{s.user1},
			errorMsg: "",
		},
		{
			name:     "should fail to validate update scope owner, missing role",
			existing: scopeWithOwners(originalOwners),
			proposed: scopeWithOwners([]types.Party{{Address: s.user1, Role: types.PartyType_PARTY_TYPE_CUSTODIAN}}),
			signers:  []string{s.user1},
			errorMsg: "missing roles required by spec: OWNER need 1 have 0",
		},
		{
			name:     "should fail to validate update scope owner, empty list",
			existing: scopeWithOwners(originalOwners),
			proposed: scopeWithOwners([]types.Party{}),
			signers:  []string{s.user1},
			errorMsg: "invalid scope owners: at least one party is required",
		},
		{
			name:     "should successfully validate update scope owner, 1st owner removed",
			existing: scopeWithOwners(ownerPartyList(s.user1, s.user2)),
			proposed: scopeWithOwners(ownerPartyList(s.user2)),
			signers:  []string{s.user1, s.user2},
			errorMsg: "",
		},
		{
			name:     "should successfully validate update scope owner, 2nd owner removed",
			existing: scopeWithOwners(ownerPartyList(s.user1, s.user2)),
			proposed: scopeWithOwners(ownerPartyList(s.user1)),
			signers:  []string{s.user1, s.user2},
			errorMsg: "",
		},
		{
			name:     "should fail to add optional owner to a non-rollup scope",
			existing: scopeWithOwners(ownerPartyList(s.user1)),
			proposed: scopeWithOwners([]types.Party{
				{Address: s.user1, Role: types.PartyType_PARTY_TYPE_OWNER, Optional: false},
				{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER, Optional: true},
			}),
			signers:  []string{s.user1},
			errorMsg: "parties can only be optional when require_party_rollup = true",
		},
		{
			name: "scope spec not found",
			existing: types.Scope{
				ScopeId:           s.scopeID,
				SpecificationId:   dneScopeSpecID,
				Owners:            originalOwners,
				DataAccess:        []string{s.user1},
				ValueOwnerAddress: s.user1,
			},
			proposed: types.Scope{
				ScopeId:           s.scopeID,
				SpecificationId:   dneScopeSpecID,
				Owners:            ownerPartyList(s.user1, s.user2),
				DataAccess:        []string{s.user1},
				ValueOwnerAddress: s.user1,
			},
			signers:  []string{s.user1},
			errorMsg: "scope specification " + dneScopeSpecID.String() + " not found",
		},
		{
			name: "with rollup, scope spec not found",
			existing: types.Scope{
				ScopeId:            s.scopeID,
				SpecificationId:    dneScopeSpecID,
				Owners:             originalOwners,
				DataAccess:         []string{s.user1},
				ValueOwnerAddress:  s.user1,
				RequirePartyRollup: true,
			},
			proposed: types.Scope{
				ScopeId:            s.scopeID,
				SpecificationId:    dneScopeSpecID,
				Owners:             ownerPartyList(s.user1, s.user2),
				DataAccess:         []string{s.user1},
				ValueOwnerAddress:  s.user1,
				RequirePartyRollup: true,
			},
			signers:  []string{s.user1},
			errorMsg: "scope specification " + dneScopeSpecID.String() + " not found",
		},
		{
			name:     "with rollup new owners do not have required roles",
			existing: rollupScopeWithOwners(pt(s.user1, owner, false)),
			proposed: rollupScopeWithOwners(pt(s.user1, omnibus, false)),
			signers:  []string{s.user1},
			errorMsg: "missing roles required by spec: OWNER need 1 have 0",
		},
		{
			name:     "with rollup neither optional party signed for required role",
			existing: rollupScopeWithOwners(pt(s.user1, owner, true), pt(s.user2, owner, true)),
			proposed: rollupScopeWithOwners(pt(s.user1, owner, true), pt(s.user2, omnibus, true)),
			signers:  []string{s.user3},
			errorMsg: "missing signers for roles required by spec: OWNER need 1 have 0",
		},
		{
			name:     "with rollup one optional party signed for required role",
			existing: rollupScopeWithOwners(pt(s.user1, owner, true), pt(s.user2, owner, true)),
			proposed: rollupScopeWithOwners(pt(s.user1, owner, true), pt(s.user2, omnibus, true)),
			signers:  []string{s.user2},
			errorMsg: "",
		},
		{
			name:     "with rollup required party not signed",
			existing: rollupScopeWithOwners(pt(s.user1, owner, true), pt(s.user2, omnibus, false)),
			proposed: rollupScopeWithOwners(pt(s.user1, owner, true), pt(s.user3, omnibus, false)),
			signers:  []string{s.user1},
			errorMsg: "missing required signature: " + s.user2 + " (OMNIBUS)",
		},
		{
			name:     "with rollup all good",
			existing: rollupScopeWithOwners(pt(s.user1, owner, true), pt(s.user2, omnibus, false)),
			proposed: rollupScopeWithOwners(pt(s.user1, owner, true), pt(s.user3, omnibus, false)),
			signers:  []string{s.user1, s.user2},
			errorMsg: "",
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			var msg types.MetadataMsg
			if len(tc.proposed.Owners) > len(tc.existing.Owners) {
				msg = &types.MsgAddScopeOwnerRequest{Signers: tc.signers}
			} else {
				msg = &types.MsgDeleteScopeOwnerRequest{Signers: tc.signers}
			}
			err := s.app.MetadataKeeper.ValidateUpdateScopeOwners(s.FreshCtx(), tc.existing, tc.proposed, msg)
			if len(tc.errorMsg) > 0 {
				assert.EqualError(t, err, tc.errorMsg, "ValidateUpdateScopeOwners expected error")
			} else {
				assert.NoError(t, err, "ValidateUpdateScopeOwners unexpected error")
			}
		})
	}
}

func (s *ScopeKeeperTestSuite) TestScopeIndexing() {
	scopeID := types.ScopeMetadataAddress(uuid.New())

	specIDOrig := types.ScopeSpecMetadataAddress(uuid.New())
	specIDNew := types.ScopeSpecMetadataAddress(uuid.New())

	ownerConstant := randomUser()
	ownerToAdd := randomUser()
	ownerToRemove := randomUser()
	valueOwnerOrig := randomUser()
	valueOwnerNew := randomUser()

	scopeV1 := types.Scope{
		ScopeId:           scopeID,
		SpecificationId:   specIDOrig,
		Owners:            ownerPartyList(ownerConstant.Bech32, ownerToRemove.Bech32),
		DataAccess:        nil,
		ValueOwnerAddress: valueOwnerOrig.Bech32,
	}
	scopeV2 := types.Scope{
		ScopeId:           scopeID,
		SpecificationId:   specIDNew,
		Owners:            ownerPartyList(ownerConstant.Bech32, ownerToAdd.Bech32),
		DataAccess:        nil,
		ValueOwnerAddress: valueOwnerNew.Bech32,
	}

	ctx := s.FreshCtx()
	store := ctx.KVStore(s.app.GetKey(types.ModuleName))

	s.T().Run("1 write new scope", func(t *testing.T) {
		expectedIndexes := []struct {
			key  []byte
			name string
		}{
			{types.GetAddressScopeCacheKey(ownerConstant.Addr, scopeID), "ownerConstant address index"},
			{types.GetAddressScopeCacheKey(ownerToRemove.Addr, scopeID), "ownerToRemove address index"},
			{types.GetAddressScopeCacheKey(valueOwnerOrig.Addr, scopeID), "valueOwnerOrig address index"},

			{types.GetValueOwnerScopeCacheKey(valueOwnerOrig.Addr, scopeID), "valueOwnerOrig value owner index"},

			{types.GetScopeSpecScopeCacheKey(specIDOrig, scopeID), "specIDOrig spec index"},
		}

		s.app.MetadataKeeper.SetScope(ctx, scopeV1)

		for _, expected := range expectedIndexes {
			assert.True(t, store.Has(expected.key), expected.name)
		}
	})

	s.T().Run("2 update scope", func(t *testing.T) {
		expectedIndexes := []struct {
			key  []byte
			name string
		}{
			{types.GetAddressScopeCacheKey(ownerConstant.Addr, scopeID), "ownerConstant address index"},
			{types.GetAddressScopeCacheKey(ownerToAdd.Addr, scopeID), "ownerToAdd address index"},
			{types.GetAddressScopeCacheKey(valueOwnerNew.Addr, scopeID), "valueOwnerNew address index"},

			{types.GetValueOwnerScopeCacheKey(valueOwnerNew.Addr, scopeID), "valueOwnerNew value owner index"},

			{types.GetScopeSpecScopeCacheKey(specIDNew, scopeID), "specIDNew spec index"},
		}
		unexpectedIndexes := []struct {
			key  []byte
			name string
		}{
			{types.GetAddressScopeCacheKey(ownerToRemove.Addr, scopeID), "ownerToRemove address index"},
			{types.GetAddressScopeCacheKey(valueOwnerOrig.Addr, scopeID), "valueOwnerOrig address index"},

			{types.GetValueOwnerScopeCacheKey(valueOwnerOrig.Addr, scopeID), "valueOwnerOrig value owner index"},

			{types.GetScopeSpecScopeCacheKey(specIDOrig, scopeID), "specIDOrig spec index"},
		}

		s.app.MetadataKeeper.SetScope(ctx, scopeV2)

		for _, expected := range expectedIndexes {
			assert.True(t, store.Has(expected.key), expected.name)
		}
		for _, unexpected := range unexpectedIndexes {
			assert.False(t, store.Has(unexpected.key), unexpected.name)
		}
	})

	s.T().Run("3 delete scope", func(t *testing.T) {
		unexpectedIndexes := []struct {
			key  []byte
			name string
		}{
			{types.GetAddressScopeCacheKey(ownerConstant.Addr, scopeID), "ownerConstant address index"},
			{types.GetAddressScopeCacheKey(ownerToRemove.Addr, scopeID), "ownerToRemove address index"},
			{types.GetAddressScopeCacheKey(ownerToAdd.Addr, scopeID), "ownerToAdd address index"},
			{types.GetAddressScopeCacheKey(valueOwnerOrig.Addr, scopeID), "valueOwnerOrig address index"},
			{types.GetAddressScopeCacheKey(valueOwnerNew.Addr, scopeID), "valueOwnerNew address index"},

			{types.GetValueOwnerScopeCacheKey(valueOwnerOrig.Addr, scopeID), "valueOwnerOrig value owner index"},
			{types.GetValueOwnerScopeCacheKey(valueOwnerNew.Addr, scopeID), "valueOwnerNew value owner index"},

			{types.GetScopeSpecScopeCacheKey(specIDOrig, scopeID), "specIDOrig spec index"},
			{types.GetScopeSpecScopeCacheKey(specIDNew, scopeID), "specIDNew spec index"},
		}

		s.app.MetadataKeeper.RemoveScope(ctx, scopeID)

		for _, unexpected := range unexpectedIndexes {
			assert.False(t, store.Has(unexpected.key), unexpected.name)
		}
	})
}
