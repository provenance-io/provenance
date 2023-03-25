package keeper_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/internal/pioconfig"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	"github.com/provenance-io/provenance/x/metadata/keeper"
	"github.com/provenance-io/provenance/x/metadata/types"
)

type AuthzTestSuite struct {
	suite.Suite

	app *simapp.App
	ctx sdk.Context

	pubkey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress

	pubkey2   cryptotypes.PubKey
	user2     string
	user2Addr sdk.AccAddress

	pubkey3   cryptotypes.PubKey
	user3     string
	user3Addr sdk.AccAddress
}

func (s *AuthzTestSuite) SetupTest() {
	pioconfig.SetProvenanceConfig("atom", 0)
	s.app = simapp.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{Time: time.Now()})

	s.pubkey1 = secp256k1.GenPrivKey().PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()
	s.app.AccountKeeper.SetAccount(s.ctx, s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user1Addr))

	s.pubkey2 = secp256k1.GenPrivKey().PubKey()
	s.user2Addr = sdk.AccAddress(s.pubkey2.Address())
	s.user2 = s.user2Addr.String()

	s.pubkey3 = secp256k1.GenPrivKey().PubKey()
	s.user3Addr = sdk.AccAddress(s.pubkey3.Address())
	s.user3 = s.user3Addr.String()
}

func TestAuthzTestSuite(t *testing.T) {
	suite.Run(t, new(AuthzTestSuite))
}

func (s *AuthzTestSuite) TestGetAuthzMessageTypeURLs() {
	type testCase struct {
		name     string // defaults to the msg name (from the url) if not defined.
		url      string
		expected []string
	}
	getMsgName := func(url string) string {
		lastDot := strings.LastIndex(url, ".")
		if lastDot < 0 || lastDot+1 >= len(url) {
			return url
		}
		return url[lastDot+1:]
	}
	getName := func(tc testCase) string {
		if tc.name != "" {
			return tc.name
		}
		return getMsgName(tc.url)
	}
	boringCase := func(url string) testCase {
		return testCase{
			name:     "boring " + getMsgName(url),
			url:      url,
			expected: []string{url},
		}
	}

	tests := []testCase{
		{
			name:     "empty",
			url:      "",
			expected: []string{},
		},
		{
			name:     "random",
			url:      "random",
			expected: []string{"random"},
		},
		boringCase(types.TypeURLMsgWriteScopeRequest),
		boringCase(types.TypeURLMsgDeleteScopeRequest),
		{
			url:      types.TypeURLMsgAddScopeDataAccessRequest,
			expected: []string{types.TypeURLMsgAddScopeDataAccessRequest, types.TypeURLMsgWriteScopeRequest},
		},
		{
			url:      types.TypeURLMsgDeleteScopeDataAccessRequest,
			expected: []string{types.TypeURLMsgDeleteScopeDataAccessRequest, types.TypeURLMsgWriteScopeRequest},
		},
		{
			url:      types.TypeURLMsgAddScopeOwnerRequest,
			expected: []string{types.TypeURLMsgAddScopeOwnerRequest, types.TypeURLMsgWriteScopeRequest},
		},
		{
			url:      types.TypeURLMsgDeleteScopeOwnerRequest,
			expected: []string{types.TypeURLMsgDeleteScopeOwnerRequest, types.TypeURLMsgWriteScopeRequest},
		},
		boringCase(types.TypeURLMsgWriteSessionRequest),
		{
			url:      types.TypeURLMsgWriteRecordRequest,
			expected: []string{types.TypeURLMsgWriteRecordRequest, types.TypeURLMsgWriteSessionRequest},
		},
		boringCase(types.TypeURLMsgDeleteRecordRequest),
		boringCase(types.TypeURLMsgWriteScopeSpecificationRequest),
		boringCase(types.TypeURLMsgDeleteScopeSpecificationRequest),
		boringCase(types.TypeURLMsgWriteContractSpecificationRequest),
		boringCase(types.TypeURLMsgDeleteContractSpecificationRequest),
		{
			url:      types.TypeURLMsgAddContractSpecToScopeSpecRequest,
			expected: []string{types.TypeURLMsgAddContractSpecToScopeSpecRequest, types.TypeURLMsgWriteScopeSpecificationRequest},
		},
		{
			url:      types.TypeURLMsgDeleteContractSpecFromScopeSpecRequest,
			expected: []string{types.TypeURLMsgDeleteContractSpecFromScopeSpecRequest, types.TypeURLMsgWriteScopeSpecificationRequest},
		},
		{
			url:      types.TypeURLMsgWriteRecordSpecificationRequest,
			expected: []string{types.TypeURLMsgWriteRecordSpecificationRequest, types.TypeURLMsgWriteContractSpecificationRequest},
		},
		{
			url:      types.TypeURLMsgDeleteRecordSpecificationRequest,
			expected: []string{types.TypeURLMsgDeleteRecordSpecificationRequest, types.TypeURLMsgDeleteContractSpecificationRequest},
		},
		boringCase(types.TypeURLMsgBindOSLocatorRequest),
		boringCase(types.TypeURLMsgDeleteOSLocatorRequest),
		boringCase(types.TypeURLMsgModifyOSLocatorRequest),
	}

	for _, tc := range tests {
		s.Run(getName(tc), func() {
			actual := s.app.MetadataKeeper.GetAuthzMessageTypeURLs(tc.url)
			s.Assert().Equal(tc.expected, actual, "GetAuthzMessageTypeURLs(%q)", tc.url)
		})
	}
}

func (s *AuthzTestSuite) TestValidateAllOwnersAreSigners() {
	// Add a few authorizations

	// User3 can sign for User2 on MsgAddScopeDataAccessRequest.
	// Does not apply to MsgWriteScopeRequest or MsgAddScopeOwnerRequest.
	a := authz.NewGenericAuthorization(types.TypeURLMsgAddScopeDataAccessRequest)
	err := s.app.AuthzKeeper.SaveGrant(s.ctx, s.user3Addr, s.user2Addr, a, nil)
	s.Require().NoError(err, "SaveGrant 2 -> 3 MsgAddScopeDataAccessRequest")

	// User3 can sign for User2 on MsgWriteScopeSpecificationRequest.
	// Applies to MsgDeleteContractSpecFromScopeSpecRequest too.
	a = authz.NewGenericAuthorization(types.TypeURLMsgWriteScopeSpecificationRequest)
	err = s.app.AuthzKeeper.SaveGrant(s.ctx, s.user3Addr, s.user2Addr, a, nil)
	s.Require().NoError(err, "SaveGrant 2 -> 3 MsgWriteScopeSpecificationRequest")

	// User3 can sign for User1 on MsgDeleteContractSpecFromScopeSpecRequest.
	// Does not apply to MsgWriteScopeSpecificationRequest
	a = authz.NewGenericAuthorization(types.TypeURLMsgDeleteContractSpecFromScopeSpecRequest)
	err = s.app.AuthzKeeper.SaveGrant(s.ctx, s.user3Addr, s.user1Addr, a, nil)
	s.Require().NoError(err, "SaveGrant 2 -> 3 MsgWriteScopeSpecificationRequest")

	randAddr1 := sdk.AccAddress("random_address_1____").String()
	randAddr2 := sdk.AccAddress("random_address_2____").String()
	randAddr3 := sdk.AccAddress("random_address_3____").String()

	tests := []struct {
		name     string
		owners   []string
		msg      types.MetadataMsg
		errorMsg string
	}{
		{
			name:     "1 owner no signers",
			owners:   []string{s.user1},
			msg:      &types.MsgWriteSessionRequest{Signers: []string{}},
			errorMsg: fmt.Sprintf("missing signature from existing owner %s; required for update", s.user1),
		},
		{
			name:     "1 owner not in signers list",
			owners:   []string{s.user1},
			msg:      &types.MsgWriteSessionRequest{Signers: []string{randAddr1, randAddr2}},
			errorMsg: fmt.Sprintf("missing signature from existing owner %s; required for update", s.user1),
		},
		{
			name:     "1 owner in signers list with non-owners",
			owners:   []string{s.user1},
			msg:      &types.MsgWriteSessionRequest{Signers: []string{randAddr1, s.user1, randAddr2}},
			errorMsg: "",
		},
		{
			name:     "1 owner only signer in list",
			owners:   []string{s.user1},
			msg:      &types.MsgWriteSessionRequest{Signers: []string{s.user1}},
			errorMsg: "",
		},
		{
			name:   "2 owners no signers",
			owners: []string{s.user1, s.user2},
			msg:    &types.MsgWriteSessionRequest{Signers: []string{}},
			errorMsg: fmt.Sprintf("missing signatures from existing owners %v; required for update",
				[]string{s.user1, s.user2}),
		},
		{
			name:   "2 owners - neither in signers list",
			owners: []string{s.user1, s.user2},
			msg:    &types.MsgWriteSessionRequest{Signers: []string{randAddr1, randAddr2, randAddr3}},
			errorMsg: fmt.Sprintf("missing signatures from existing owners %v; required for update",
				[]string{s.user1, s.user2}),
		},
		{
			name:     "2 owners - first in signers list with non-owners",
			owners:   []string{s.user1, s.user2},
			msg:      &types.MsgWriteSessionRequest{Signers: []string{randAddr1, s.user1, randAddr3}},
			errorMsg: fmt.Sprintf("missing signature from existing owner %s; required for update", s.user2),
		},
		{
			name:     "2 owners - second in signers list with non-owners",
			owners:   []string{s.user1, s.user2},
			msg:      &types.MsgWriteSessionRequest{Signers: []string{randAddr1, s.user2, randAddr3}},
			errorMsg: fmt.Sprintf("missing signature from existing owner %s; required for update", s.user1),
		},
		{
			name:     "2 owners - both in signers list with non-owners",
			owners:   []string{s.user1, s.user2},
			msg:      &types.MsgWriteSessionRequest{Signers: []string{randAddr1, s.user2, randAddr2, s.user1, randAddr3}},
			errorMsg: "",
		},
		{
			name:     "2 owners - both in signers list",
			owners:   []string{s.user1, s.user2},
			msg:      &types.MsgWriteSessionRequest{Signers: []string{s.user1, s.user2}},
			errorMsg: "",
		},
		{
			name:     "2 owners - both in signers list, opposite order",
			owners:   []string{s.user1, s.user2},
			msg:      &types.MsgWriteSessionRequest{Signers: []string{s.user2, s.user1}},
			errorMsg: "",
		},
		// authz test cases
		{
			name: "authz - 2 owners - with grant but both are signers",
			// User3 can sign for User2 on MsgAddScopeDataAccessRequest.
			owners:   []string{s.user2, s.user3},
			msg:      &types.MsgAddScopeDataAccessRequest{Signers: []string{s.user2, s.user3}},
			errorMsg: "",
		},
		{
			name: "authz - 2 owners - 1 signer - no grant",
			// 3 has not granted anything to 2 (it's the other way around).
			owners:   []string{s.user2, s.user3},
			msg:      &types.MsgAddScopeDataAccessRequest{Signers: []string{s.user2}},
			errorMsg: fmt.Sprintf("missing signature from existing owner %s; required for update", s.user3),
		},
		{
			name: "authz - 2 owners - 1 signer - grant on child msg",
			// User3 can sign for User2 on MsgAddScopeDataAccessRequest, but not MsgWriteScopeRequest
			owners:   []string{s.user2, s.user3},
			msg:      &types.MsgWriteScopeRequest{Signers: []string{s.user3}},
			errorMsg: fmt.Sprintf("missing signature from existing owner %s; required for update", s.user2),
		},
		{
			name: "authz - 2 owners - 1 signer - grant on sibling msg",
			// User3 can sign for User2 on MsgAddScopeDataAccessRequest, but not MsgAddScopeOwnerRequest
			owners:   []string{s.user2, s.user3},
			msg:      &types.MsgAddScopeOwnerRequest{Signers: []string{s.user3}},
			errorMsg: fmt.Sprintf("missing signature from existing owner %s; required for update", s.user2),
		},
		{
			name: "authz - 2 owners - 1 signer - with grant",
			// User3 can sign for User2 on MsgAddScopeDataAccessRequest.
			owners:   []string{s.user2, s.user3},
			msg:      &types.MsgAddScopeDataAccessRequest{Signers: []string{s.user3}},
			errorMsg: "",
		},
		{
			name: "authz - 3 owners - 2 signers - with grant",
			// User3 can sign for User2 on MsgAddScopeDataAccessRequest.
			owners:   []string{s.user1, s.user2, s.user3},
			msg:      &types.MsgAddScopeDataAccessRequest{Signers: []string{s.user1, s.user3}},
			errorMsg: "",
		},
		{
			name: "authz - 3 owners - 2 signers - grant from parent msg type",
			// User3 can sign for User2 on MsgWriteScopeSpecificationRequest.
			owners:   []string{s.user1, s.user2, s.user3},
			msg:      &types.MsgDeleteContractSpecFromScopeSpecRequest{Signers: []string{s.user1, s.user3}},
			errorMsg: "",
		},
		{
			name: "authz - 3 owners - 1 signer - 2 grants",
			// User3 can sign for User1 on MsgDeleteContractSpecFromScopeSpecRequest.
			// User3 can sign for User2 on MsgWriteScopeSpecificationRequest.
			owners:   []string{s.user1, s.user2, s.user3},
			msg:      &types.MsgDeleteContractSpecFromScopeSpecRequest{Signers: []string{s.user3}},
			errorMsg: "",
		},
		{
			name: "authz - 3 owners - 1 signer - 1 grant",
			// User3 can sign for User2 on MsgWriteScopeSpecificationRequest, but not user 1.
			owners:   []string{s.user1, s.user2, s.user3},
			msg:      &types.MsgWriteScopeSpecificationRequest{Signers: []string{s.user3}},
			errorMsg: fmt.Sprintf("missing signature from existing owner %s; required for update", s.user1),
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			err := s.app.MetadataKeeper.ValidateAllOwnersAreSignersWithAuthz(s.ctx, tc.owners, tc.msg)
			if len(tc.errorMsg) == 0 {
				assert.NoError(t, err, "ValidateAllOwnersAreSigners unexpected error")
			} else {
				assert.EqualError(t, err, tc.errorMsg, "ValidateAllOwnersAreSigners error")
			}
		})
	}
}

func (s *AuthzTestSuite) TestValidateAllOwnersAreSignersWithCountAuthorization() {

	oneAllowedAuthorizations := int32(1)
	manyAllowedAuthorizations := int32(10)

	tests := []struct {
		name     string
		owners   []string
		msg      types.MetadataMsg
		count    int32
		granter  sdk.AccAddress
		grantee  sdk.AccAddress
		errorMsg string
	}{
		// count authorization test cases
		{
			name:     "Scope Spec with 2 owners - one signer - with grant - authz",
			owners:   []string{s.user2, s.user3},
			msg:      &types.MsgAddScopeOwnerRequest{Signers: []string{s.user3}},
			count:    oneAllowedAuthorizations,
			granter:  s.user2Addr,
			grantee:  s.user3Addr,
			errorMsg: "",
		},
		{
			name:     "Scope Spec with 2 owners - one signer - no grant - authz - error",
			owners:   []string{s.user2, s.user3},
			msg:      &types.MsgWriteScopeRequest{Signers: []string{s.user2}},
			count:    manyAllowedAuthorizations,
			granter:  s.user2Addr,
			grantee:  s.user3Addr,
			errorMsg: fmt.Sprintf("missing signature from existing owner %s; required for update", s.user3),
		},
		{
			name:     "Scope Spec with 3 owners - one signer with a special case message type - with grant - authz",
			owners:   []string{s.user1, s.user2, s.user3},
			msg:      &types.MsgAddScopeDataAccessRequest{Signers: []string{s.user1, s.user3}}, // signer 3 is grantee of singer 2
			count:    manyAllowedAuthorizations,
			granter:  s.user2Addr,
			grantee:  s.user3Addr,
			errorMsg: "",
		},
		{
			name:     "Scope Spec with 3 owners - two signers with a special case message type - grant on parent of special case message type - authz",
			owners:   []string{s.user1, s.user2, s.user3},
			msg:      &types.MsgDeleteContractSpecFromScopeSpecRequest{Signers: []string{s.user1, s.user3}}, // signer 3 grantee of signer 2
			count:    manyAllowedAuthorizations,
			granter:  s.user2Addr,
			grantee:  s.user3Addr,
			errorMsg: "",
		},
		{
			name:     "Scope Spec with 2 owners - one signer - no grant - authz - error",
			owners:   []string{s.user2, s.user3},
			msg:      &types.MsgDeleteRecordRequest{Signers: []string{s.user3}},
			count:    manyAllowedAuthorizations,
			granter:  nil,
			grantee:  nil,
			errorMsg: fmt.Sprintf("missing signature from existing owner %s; required for update", s.user2),
		},
	}

	// Test cases
	for _, tc := range tests {
		s.Run(tc.name, func() {
			msgTypeURL := sdk.MsgTypeURL(tc.msg)
			if tc.grantee != nil && tc.granter != nil {
				a := authz.NewCountAuthorization(msgTypeURL, tc.count)
				err := s.app.AuthzKeeper.SaveGrant(s.ctx, tc.grantee, tc.granter, a, nil)
				s.Require().NoError(err, "SaveGrant")
			}

			err := s.app.MetadataKeeper.ValidateAllOwnersAreSignersWithAuthz(s.ctx, tc.owners, tc.msg)
			if len(tc.errorMsg) == 0 {
				s.Assert().NoError(err, "ValidateAllOwnersAreSignersWithAuthz error")
			} else {
				s.Assert().EqualError(err, tc.errorMsg, "ValidateAllOwnersAreSignersWithAuthz error")
			}

			// validate allowedAuthorizations
			if err == nil {
				auth, _ := s.app.AuthzKeeper.GetAuthorization(s.ctx, tc.grantee, tc.granter, msgTypeURL)
				if tc.count == 1 {
					// authorization is deleted after one use
					s.Assert().Nil(auth, "GetAuthorization after only allowed use")
				} else {
					actual := auth.(*authz.CountAuthorization).AllowedAuthorizations
					s.Assert().Equal(tc.count-1, actual, "uses left on authorization")
				}
			}
		})
	}

	s.Run("ensure authorizations are updated", func() {
		// Two owners (1 & 2), and one signer (3),
		// Two authz count authorization
		//	- count grants:
		//		granter: 1, grantee: 3, count: 1
		//		granter: 2, grantee: 3, count: 2
		// Require signatures from 1 and 2, but sign with 3.
		// Ensure both authorizations are applied and updated.

		msg := &types.MsgDeleteScopeRequest{}
		msgTypeUrl := sdk.MsgTypeURL(msg)

		// first grant: 3 can sign for 1 one time.
		a := authz.NewCountAuthorization(msgTypeUrl, 1)
		err := s.app.AuthzKeeper.SaveGrant(s.ctx, s.user3Addr, s.user1Addr, a, nil)
		s.Assert().NoError(err, "SaveGrant 1 -> 3, 1 use")

		// second grant: 3 can sign for 2 two times.
		a = authz.NewCountAuthorization(msgTypeUrl, 2)
		err = s.app.AuthzKeeper.SaveGrant(s.ctx, s.user3Addr, s.user2Addr, a, nil)
		s.Assert().NoError(err, "SaveGrant 2 -> 3, 2 uses")

		// two owners (1 & 2), and one signer (3)
		owners := []string{s.user1, s.user2}
		msg.Signers = []string{s.user3}

		// Validate signatures. This should also use both count authorizations.
		err = s.app.MetadataKeeper.ValidateAllOwnersAreSignersWithAuthz(s.ctx, owners, msg)
		s.Assert().NoError(err, "ValidateAllOwnersAreSigners")

		// first grant should be deleted because it used its last use.
		auth, _ := s.app.AuthzKeeper.GetAuthorization(s.ctx, s.user3Addr, s.user1Addr, msgTypeUrl)
		s.Assert().Nil(auth, "GetAuthorization 1 -> 3 after only allowed use")

		// second grant should still exist, but only have one use left.
		auth, _ = s.app.AuthzKeeper.GetAuthorization(s.ctx, s.user3Addr, s.user2Addr, msgTypeUrl)
		s.Assert().NotNil(auth, "GetAuthorization 2 -> 3 after one use")
		actual := auth.(*authz.CountAuthorization).AllowedAuthorizations
		s.Assert().Equal(1, int(actual), "number of uses left on 2 -> 3 authorization")
	})
}

func (s *AuthzTestSuite) TestValidateAllOwnerPartiesAreSigners() {
	// A missing signature with an authz grant on MsgAddScopeOwnerRequest
	a := authz.NewGenericAuthorization(types.TypeURLMsgWriteScopeRequest)
	err := s.app.AuthzKeeper.SaveGrant(s.ctx, s.user3Addr, s.user1Addr, a, nil)
	s.Require().NoError(err)

	// A missing signature on a special case message type with an authz grant on that type.
	// Add (a child msg type) TypeURLMsgAddScopeDataAccessRequest  (of a parent) TypeURLMsgWriteScopeRequest
	a = authz.NewGenericAuthorization(types.TypeURLMsgAddScopeDataAccessRequest)
	err = s.app.AuthzKeeper.SaveGrant(s.ctx, s.user3Addr, s.user2Addr, a, nil)
	s.Require().NoError(err)

	// A missing signature on a special case message type with an authz grant on its parent type.
	// Add grant on the parent type of TypeURLMsgAddContractSpecToScopeSpecRequest.
	a = authz.NewGenericAuthorization(types.TypeURLMsgWriteScopeSpecificationRequest)
	err = s.app.AuthzKeeper.SaveGrant(s.ctx, s.user3Addr, s.user2Addr, a, nil)
	s.Require().NoError(err)

	cases := []struct {
		name     string
		owners   []types.Party
		msg      types.MetadataMsg
		errorMsg string
	}{
		{
			name:     "no owners - no signers",
			owners:   []types.Party{},
			msg:      &types.MsgDeleteRecordRequest{Signers: []string{}},
			errorMsg: "",
		},
		{
			name:     "one owner - is signer",
			owners:   []types.Party{{Address: "signer1", Role: types.PartyType_PARTY_TYPE_OWNER}},
			msg:      &types.MsgDeleteRecordRequest{Signers: []string{"signer1"}},
			errorMsg: "",
		},
		{
			name:     "one owner - is one of two signers",
			owners:   []types.Party{{Address: "signer1", Role: types.PartyType_PARTY_TYPE_OWNER}},
			msg:      &types.MsgDeleteRecordRequest{Signers: []string{"signer1", "signer2"}},
			errorMsg: "",
		},
		{
			name:     "one owner - is not one of two signers",
			owners:   []types.Party{{Address: "missingowner", Role: types.PartyType_PARTY_TYPE_OWNER}},
			msg:      &types.MsgDeleteRecordRequest{Signers: []string{"signer1", "signer2"}},
			errorMsg: "missing signature from [missingowner (PARTY_TYPE_OWNER)]",
		},
		{
			name: "two owners - both are signers",
			owners: []types.Party{
				{Address: "owner1", Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: "owner2", Role: types.PartyType_PARTY_TYPE_OWNER}},
			msg:      &types.MsgDeleteRecordRequest{Signers: []string{"owner2", "owner1"}},
			errorMsg: "",
		},
		{
			name: "two owners - only one is signer",
			owners: []types.Party{
				{Address: "owner1", Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: "missingowner", Role: types.PartyType_PARTY_TYPE_OWNER}},
			msg:      &types.MsgDeleteRecordRequest{Signers: []string{"owner2", "owner1"}},
			errorMsg: "missing signature from [missingowner (PARTY_TYPE_OWNER)]",
		},
		{
			name: "two parties - one owner one other - only owner is signer",
			owners: []types.Party{
				{Address: "owner", Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: "affiliate", Role: types.PartyType_PARTY_TYPE_AFFILIATE}},
			msg:      &types.MsgDeleteRecordRequest{Signers: []string{"owner"}},
			errorMsg: "missing signature from [affiliate (PARTY_TYPE_AFFILIATE)]",
		},
		{
			name: "two parties - one owner one other - only other is signer",
			owners: []types.Party{
				{Address: "owner", Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: "affiliate", Role: types.PartyType_PARTY_TYPE_AFFILIATE}},
			msg:      &types.MsgDeleteRecordRequest{Signers: []string{"affiliate"}},
			errorMsg: "missing signature from [owner (PARTY_TYPE_OWNER)]",
		},
		// authz test cases
		{
			name: "two parties - one missing signature with authz grant - two signers",
			owners: []types.Party{
				{Address: s.user1, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user3, Role: types.PartyType_PARTY_TYPE_OWNER}}, // grantee of singer 1
			msg:      &types.MsgWriteScopeRequest{Signers: []string{s.user2, s.user3}},
			errorMsg: "",
		},
		{
			name: "two parties - one missing signature without authz grant - one signer",
			owners: []types.Party{
				{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user3, Role: types.PartyType_PARTY_TYPE_OWNER}},
			msg:      &types.MsgWriteScopeRequest{Signers: []string{s.user2}},
			errorMsg: fmt.Sprintf("missing signature from [%s (PARTY_TYPE_OWNER)]", s.user3),
		},
		{
			name: "two parties - one missing signature with a special case message type - authz grant - two signers",
			owners: []types.Party{
				{Address: s.user1, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user3, Role: types.PartyType_PARTY_TYPE_OWNER}}, // grantee of singer 2
			msg:      &types.MsgAddScopeDataAccessRequest{Signers: []string{s.user1, s.user3}},
			errorMsg: "",
		},
		{
			name: "two parties - one missing signature with a special case message type - authz grant on parent message type - two signers",
			owners: []types.Party{
				{Address: s.user1, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user3, Role: types.PartyType_PARTY_TYPE_OWNER}}, // grantee of singer 2
			msg:      &types.MsgAddContractSpecToScopeSpecRequest{Signers: []string{s.user1, s.user3}},
			errorMsg: "",
		},
		{
			name: "two parties - one missing signature with a special case message type without authz grant - one signer",
			owners: []types.Party{
				{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user3, Role: types.PartyType_PARTY_TYPE_OWNER}},
			msg:      &types.MsgDeleteRecordRequest{Signers: []string{s.user3}},
			errorMsg: fmt.Sprintf("missing signature from [%s (PARTY_TYPE_OWNER)]", s.user2),
		},
	}

	// Test cases
	for _, tc := range cases {
		s.Run(tc.name, func() {
			err = s.app.MetadataKeeper.ValidateAllPartiesAreSignersWithAuthz(s.ctx, tc.owners, tc.msg)
			if len(tc.errorMsg) == 0 {
				s.Assert().NoError(err, "ValidateAllPartiesAreSignersWithAuthz")
			} else {
				s.Assert().EqualError(err, tc.errorMsg, "ValidateAllPartiesAreSignersWithAuthz")
			}
		})
	}
}

func (s *AuthzTestSuite) TestValidateAllOwnerPartiesAreSignersWithCountAuthorization() {

	oneAllowedAuthorizations := int32(1)
	manyAllowedAuthorizations := int32(10)

	cases := []struct {
		name     string
		owners   []types.Party
		msg      types.MetadataMsg
		count    int32
		granter  sdk.AccAddress
		grantee  sdk.AccAddress
		errorMsg string
	}{
		// count authorization test cases
		{
			name: "three parties - one missing signature with one authz grant - two signers",
			owners: []types.Party{
				{Address: s.user1, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user3, Role: types.PartyType_PARTY_TYPE_OWNER}}, // grantee of singer 1
			msg:      &types.MsgWriteScopeRequest{Signers: []string{s.user2, s.user3}},
			count:    oneAllowedAuthorizations,
			granter:  s.user1Addr,
			grantee:  s.user3Addr,
			errorMsg: "",
		},
		{
			name: "three parties - one missing signature with a special case message type - authz grant - two signers",
			owners: []types.Party{
				{Address: s.user1, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user3, Role: types.PartyType_PARTY_TYPE_OWNER}}, // grantee of singer 2
			msg:      &types.MsgAddScopeDataAccessRequest{Signers: []string{s.user1, s.user3}},
			count:    manyAllowedAuthorizations,
			granter:  s.user2Addr,
			grantee:  s.user3Addr,
			errorMsg: "",
		},
		{
			name: "three parties - one missing signature with a special case message type - authz grant on parent message type - two signers",
			owners: []types.Party{
				{Address: s.user1, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user3, Role: types.PartyType_PARTY_TYPE_OWNER}}, // grantee of singer 2
			msg:      &types.MsgAddContractSpecToScopeSpecRequest{Signers: []string{s.user1, s.user3}},
			count:    manyAllowedAuthorizations,
			granter:  s.user2Addr,
			grantee:  s.user3Addr,
			errorMsg: "",
		},
		{
			name: "two parties - one missing signature with a special case message type without authz grant - one signer",
			owners: []types.Party{
				{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user3, Role: types.PartyType_PARTY_TYPE_OWNER}},
			msg:      &types.MsgDeleteRecordRequest{Signers: []string{s.user3}},
			count:    manyAllowedAuthorizations,
			granter:  nil,
			grantee:  nil,
			errorMsg: fmt.Sprintf("missing signature from [%s (PARTY_TYPE_OWNER)]", s.user2),
		},
	}

	// Test cases
	for _, tc := range cases {
		s.Run(tc.name, func() {
			msgTypeURL := sdk.MsgTypeURL(tc.msg)
			if tc.grantee != nil && tc.granter != nil {
				a := authz.NewCountAuthorization(msgTypeURL, tc.count)
				err := s.app.AuthzKeeper.SaveGrant(s.ctx, tc.grantee, tc.granter, a, nil)
				s.Require().NoError(err, "SaveGrant")
			}

			err := s.app.MetadataKeeper.ValidateAllPartiesAreSignersWithAuthz(s.ctx, tc.owners, tc.msg)
			if len(tc.errorMsg) == 0 {
				s.Assert().NoError(err, "ValidateAllPartiesAreSignersWithAuthz error")
			} else {
				s.Assert().EqualError(err, tc.errorMsg, "ValidateAllPartiesAreSignersWithAuthz error")
			}

			// validate allowedAuthorizations
			if err == nil {
				auth, _ := s.app.AuthzKeeper.GetAuthorization(s.ctx, tc.grantee, tc.granter, msgTypeURL)
				if tc.count == 1 {
					// authorization is deleted after one use
					s.Assert().Nil(auth, "GetAuthorization after only allowed use")
				} else {
					actual := auth.(*authz.CountAuthorization).AllowedAuthorizations
					s.Assert().Equal(tc.count-1, actual, "uses left on authorization")
				}
			}
		})
	}

	s.Run("ensure authorizations are updated", func() {
		// Two owners (1 & 2), and one signer (3),
		// Two authz count authorization
		//	- count grants:
		//		granter: 1, grantee: 3, count: 1
		//		granter: 2, grantee: 3, count: 2
		// Require signatures from 1 and 2, but sign with 3.
		// Ensure both authorizations are applied and updated.

		msg := &types.MsgDeleteScopeRequest{}
		msgTypeUrl := sdk.MsgTypeURL(msg)

		// first grant: 3 can sign for 1 one time.
		a := authz.NewCountAuthorization(msgTypeUrl, 1)
		err := s.app.AuthzKeeper.SaveGrant(s.ctx, s.user3Addr, s.user1Addr, a, nil)
		s.Require().NoError(err, "SaveGrant 1 -> 3, 1 use")

		// second grant: 3 can sign for 2 two times.
		a = authz.NewCountAuthorization(msgTypeUrl, 2)
		err = s.app.AuthzKeeper.SaveGrant(s.ctx, s.user3Addr, s.user2Addr, a, nil)
		s.Require().NoError(err, "SaveGrant 2 -> 3, 2 uses")

		// two parties (1 & 2), and one signer (3)
		parties := []types.Party{
			{Address: s.user1, Role: types.PartyType_PARTY_TYPE_OWNER},
			{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER}}
		msg.Signers = []string{s.user3}

		// validate signatures
		err = s.app.MetadataKeeper.ValidateAllPartiesAreSignersWithAuthz(s.ctx, parties, msg)
		s.Assert().NoError(err, "ValidateAllPartiesAreSigners")

		// first grant should be deleted because it used its last use.
		auth, _ := s.app.AuthzKeeper.GetAuthorization(s.ctx, s.user3Addr, s.user1Addr, msgTypeUrl)
		s.Assert().Nil(auth, "GetAuthorization 1 -> 3 after only allowed use")

		// second grant should still exist, but only have one use left.
		auth, _ = s.app.AuthzKeeper.GetAuthorization(s.ctx, s.user3Addr, s.user2Addr, msgTypeUrl)
		s.Assert().NotNil(auth, "GetAuthorization 2 -> 3 after one use")
		actual := auth.(*authz.CountAuthorization).AllowedAuthorizations
		s.Assert().Equal(1, int(actual), "number of uses left on 2 -> 3 authorization")
	})
}

func (s *AuthzTestSuite) TestValidatePartiesInvolved() {

	cases := map[string]struct {
		parties         []types.Party
		requiredParties []types.PartyType
		wantErr         bool
		errorMsg        string
	}{
		"valid, matching no parties involved": {
			parties:         []types.Party{},
			requiredParties: []types.PartyType{},
			wantErr:         false,
			errorMsg:        "",
		},
		"invalid, parties contain no required parties": {
			parties:         []types.Party{},
			requiredParties: []types.PartyType{types.PartyType_PARTY_TYPE_AFFILIATE},
			wantErr:         true,
			errorMsg:        "missing required party type [PARTY_TYPE_AFFILIATE] from parties",
		},
		"invalid, missing one required party": {
			parties:         []types.Party{{Address: "address", Role: types.PartyType_PARTY_TYPE_CUSTODIAN}},
			requiredParties: []types.PartyType{types.PartyType_PARTY_TYPE_AFFILIATE},
			wantErr:         true,
			errorMsg:        "missing required party type [PARTY_TYPE_AFFILIATE] from parties",
		},
		"invalid, missing twp required parties": {
			parties:         []types.Party{{Address: "address", Role: types.PartyType_PARTY_TYPE_CUSTODIAN}},
			requiredParties: []types.PartyType{types.PartyType_PARTY_TYPE_AFFILIATE, types.PartyType_PARTY_TYPE_INVESTOR},
			wantErr:         true,
			errorMsg:        "missing required party types [PARTY_TYPE_AFFILIATE PARTY_TYPE_INVESTOR] from parties",
		},
		"valid, required parties fulfilled": {
			parties:         []types.Party{{Address: "address", Role: types.PartyType_PARTY_TYPE_CUSTODIAN}},
			requiredParties: []types.PartyType{types.PartyType_PARTY_TYPE_CUSTODIAN},
			wantErr:         false,
			errorMsg:        "",
		},
	}

	for n, tc := range cases {
		tc := tc

		s.T().Run(n, func(t *testing.T) {
			err := s.app.MetadataKeeper.ValidatePartiesInvolved(tc.parties, tc.requiredParties)
			if tc.wantErr {
				assert.EqualError(t, err, tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s *AuthzTestSuite) TestFindMissing() {
	tests := map[string]struct {
		required []string
		entries  []string
		expected []string
	}{
		"empty required - empty entries - empty out": {
			[]string{},
			[]string{},
			[]string{},
		},
		"empty required - 2 entries - empty out": {
			[]string{},
			[]string{"one", "two"},
			[]string{},
		},
		"one required - is only entry - empty out": {
			[]string{"one"},
			[]string{"one"},
			[]string{},
		},
		"one required - is first of two entries - empty out": {
			[]string{"one"},
			[]string{"one", "two"},
			[]string{},
		},
		"one required - is second of two entries - empty out": {
			[]string{"one"},
			[]string{"two", "one"},
			[]string{},
		},
		"one required - empty entries - required out": {
			[]string{"one"},
			[]string{},
			[]string{"one"},
		},
		"one required - one other in entries - required out": {
			[]string{"one"},
			[]string{"two"},
			[]string{"one"},
		},
		"one required - two other in entries - required out": {
			[]string{"one"},
			[]string{"two", "three"},
			[]string{"one"},
		},
		"two required - both in entries - empty out": {
			[]string{"one", "two"},
			[]string{"one", "two"},
			[]string{},
		},
		"two required - reversed in entries - empty out": {
			[]string{"one", "two"},
			[]string{"two", "one"},
			[]string{},
		},
		"two required - only first in entries - second out": {
			[]string{"one", "two"},
			[]string{"one"},
			[]string{"two"},
		},
		"two required - only second in entries - first out": {
			[]string{"one", "two"},
			[]string{"two"},
			[]string{"one"},
		},
		"two required - first and other in entries - second out": {
			[]string{"one", "two"},
			[]string{"one", "other"},
			[]string{"two"},
		},
		"two required - second and other in entries - first out": {
			[]string{"one", "two"},
			[]string{"two", "other"},
			[]string{"one"},
		},
		"two required - empty entries - required out": {
			[]string{"one", "two"},
			[]string{},
			[]string{"one", "two"},
		},
		"two required - neither in one entries - required out": {
			[]string{"one", "two"},
			[]string{"neither"},
			[]string{"one", "two"},
		},
		"two required - neither in three entries - required out": {
			[]string{"one", "two"},
			[]string{"neither", "nor", "nothing"},
			[]string{"one", "two"},
		},
		"two required - first not in three entries 0 - first out": {
			[]string{"one", "two"},
			[]string{"two", "nor", "nothing"},
			[]string{"one"},
		},
		"two required - first not in three entries 1 - first out": {
			[]string{"one", "two"},
			[]string{"neither", "two", "nothing"},
			[]string{"one"},
		},
		"two required - first not in three entries 2 - first out": {
			[]string{"one", "two"},
			[]string{"neither", "nor", "two"},
			[]string{"one"},
		},
		"two required - second not in three entries 0 - second out": {
			[]string{"one", "two"},
			[]string{"one", "nor", "nothing"},
			[]string{"two"},
		},
		"two required - second not in three entries 1 - second out": {
			[]string{"one", "two"},
			[]string{"neither", "one", "nothing"},
			[]string{"two"},
		},
		"two required - second not in three entries 2 - second out": {
			[]string{"one", "two"},
			[]string{"neither", "nor", "one"},
			[]string{"two"},
		},
	}

	for n, tc := range tests {
		s.T().Run(n, func(t *testing.T) {
			actual := keeper.FindMissing(tc.required, tc.entries)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func (s *AuthzTestSuite) TestIsMarkerAndHasAuthority_IsMarker() {
	markerAddr := markertypes.MustGetMarkerAddress("testcoin").String()
	err := s.app.MarkerKeeper.AddMarkerAccount(s.ctx, &markertypes.MarkerAccount{
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
	s.app.AccountKeeper.SetAccount(s.ctx, s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user1Addr))

	tests := []struct {
		name     string
		addr     string
		expected bool
	}{
		{name: "is a marker", addr: markerAddr, expected: true},
		{name: "exists but is a user not a marker", addr: s.user1, expected: false},
		{name: "does not exist", addr: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", expected: false},
		{name: "invalid address", addr: "invalid", expected: false},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			isMarker, _ := s.app.MetadataKeeper.IsMarkerAndHasAuthority(s.ctx, tc.addr, []string{}, markertypes.Access_Unknown)
			s.Assert().Equal(tc.expected, isMarker, "IsMarkerAndHasAuthority first result bool")
		})
	}
}

func (s *AuthzTestSuite) TestIsMarkerAndHasAuthority_HasAuth() {
	markerAddr := markertypes.MustGetMarkerAddress("testcoin").String()
	err := s.app.MarkerKeeper.AddMarkerAccount(s.ctx, &markertypes.MarkerAccount{
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
	s.app.AccountKeeper.SetAccount(s.ctx, s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user1Addr))

	tests := []struct {
		name     string
		addr     string
		signers  []string
		role     markertypes.Access
		expected bool
	}{
		{
			name:     "invalid value owner",
			addr:     "invalid",
			signers:  []string{s.user1},
			role:     markertypes.Access_Deposit,
			expected: false,
		},
		{
			name:     "value owner does not exist",
			addr:     "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck",
			signers:  []string{s.user1},
			role:     markertypes.Access_Deposit,
			expected: false,
		},
		{
			name:     "addr is not a marker",
			addr:     s.user1,
			signers:  []string{s.user1},
			role:     markertypes.Access_Deposit,
			expected: false,
		},
		{
			name:     "user has access",
			addr:     markerAddr,
			signers:  []string{s.user1},
			role:     markertypes.Access_Deposit,
			expected: true,
		},
		{
			name:     "user has access even with invalid signer",
			addr:     markerAddr,
			signers:  []string{"invalidaddress", s.user1},
			role:     markertypes.Access_Deposit,
			expected: true,
		},
		{
			name:     "user does not have this access",
			addr:     markerAddr,
			signers:  []string{s.user1},
			role:     markertypes.Access_Burn,
			expected: false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			_, hasAuth := s.app.MetadataKeeper.IsMarkerAndHasAuthority(s.ctx, tc.addr, tc.signers, tc.role)
			s.Assert().Equal(tc.expected, hasAuth, "IsMarkerAndHasAuthority second result bool")
		})
	}
}
