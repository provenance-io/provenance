package keeper_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	tests := map[string]struct {
		owners     []string
		signers    []string
		msgTypeURL string
		errorMsg   string
	}{
		"Scope Spec with 1 owner: no signers - error": {
			owners:     []string{s.user1},
			signers:    []string{},
			msgTypeURL: "",
			errorMsg:   fmt.Sprintf("missing signature from existing owner %s; required for update", s.user1),
		},
		"Scope Spec with 1 owner: not in signers list - error": {
			owners:     []string{s.user1},
			signers:    []string{sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()},
			msgTypeURL: "",
			errorMsg:   fmt.Sprintf("missing signature from existing owner %s; required for update", s.user1),
		},
		"Scope Spec with 1 owner: in signers list with non-owners - ok": {
			owners:     []string{s.user1},
			signers:    []string{sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), s.user1, sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()},
			msgTypeURL: "",
			errorMsg:   "",
		},
		"Scope Spec with 1 owner: only signer in list - ok": {
			owners:     []string{s.user1},
			signers:    []string{s.user1},
			msgTypeURL: "",
			errorMsg:   "",
		},
		"Scope Spec with 2 owners: no signers - error": {
			owners:     []string{s.user1, s.user2},
			signers:    []string{},
			msgTypeURL: "",
			errorMsg: fmt.Sprintf("missing signatures from existing owners %v; required for update",
				[]string{s.user1, s.user2}),
		},
		"Scope Spec with 2 owners: neither in signers list - error": {
			owners:     []string{s.user1, s.user2},
			signers:    []string{sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()},
			msgTypeURL: "",
			errorMsg: fmt.Sprintf("missing signatures from existing owners %v; required for update",
				[]string{s.user1, s.user2}),
		},
		"Scope Spec with 2 owners: one in signers list with non-owners - error": {
			owners:     []string{s.user1, s.user2},
			signers:    []string{sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), s.user1, sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()},
			msgTypeURL: "",
			errorMsg:   fmt.Sprintf("missing signature from existing owner %s; required for update", s.user2),
		},
		"Scope Spec with 2 owners: the other in signers list with non-owners - error": {
			owners:     []string{s.user1, s.user2},
			signers:    []string{sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), s.user2, sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()},
			msgTypeURL: "",
			errorMsg:   fmt.Sprintf("missing signature from existing owner %s; required for update", s.user1),
		},
		"Scope Spec with 2 owners: both in signers list with non-owners - ok": {
			owners:     []string{s.user1, s.user2},
			signers:    []string{sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), s.user2, sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), s.user1},
			msgTypeURL: "",
			errorMsg:   "",
		},
		"Scope Spec with 2 owners: only both in signers list - ok": {
			owners:     []string{s.user1, s.user2},
			signers:    []string{s.user1, s.user2},
			msgTypeURL: "",
			errorMsg:   "",
		},
		"Scope Spec with 2 owners: only both in signers list, opposite order - ok": {
			owners:     []string{s.user1, s.user2},
			signers:    []string{s.user2, s.user1},
			msgTypeURL: "",
			errorMsg:   "",
		},
		// authz test cases
		"Scope Spec with 2 owners - both in signers list - authz": {
			owners:     []string{s.user2, s.user3},
			signers:    []string{s.user2, s.user3},
			msgTypeURL: types.TypeURLMsgAddScopeDataAccessRequest,
			errorMsg:   "",
		},
		"Scope Spec with 2 owners - one signer - authz - error": {
			owners:     []string{s.user2, s.user3},
			signers:    []string{s.user2},
			msgTypeURL: types.TypeURLMsgWriteScopeRequest,
			errorMsg:   fmt.Sprintf("missing signature from existing owner %s; required for update", s.user3),
		},
		"Scope Spec with 3 owners - one signer with a special case message type - with grant - authz": {
			owners:     []string{s.user1, s.user2, s.user3},
			signers:    []string{s.user1, s.user3}, // signer 3 is grantee of singer 2
			msgTypeURL: types.TypeURLMsgAddScopeDataAccessRequest,
			errorMsg:   "",
		},
		"Scope Spec with 3 owners - two signers with a special case message type - grant on parent of special case message type - authz": {
			owners:     []string{s.user1, s.user2, s.user3},
			signers:    []string{s.user1, s.user3}, // signer 3 grantee of signer 2
			msgTypeURL: types.TypeURLMsgDeleteContractSpecFromScopeSpecRequest,
			errorMsg:   "",
		},
		"Scope Spec with 2 owners - one signer - no grant - authz - error": {
			owners:     []string{s.user2, s.user3},
			signers:    []string{s.user3},
			msgTypeURL: types.TypeURLMsgDeleteRecordRequest,
			errorMsg:   fmt.Sprintf("missing signature from existing owner %s; required for update", s.user2),
		},
	}

	// Add a few authorizations
	now := s.ctx.BlockHeader().Time
	s.Require().NotNil(now, "now")
	exp1Hour := now.Add(time.Hour)

	// A missing signature with an authz grant on MsgAddScopeOwnerRequest
	granter := s.user1Addr
	grantee := s.user3Addr
	a := authz.NewGenericAuthorization(types.TypeURLMsgAddScopeOwnerRequest)
	err := s.app.AuthzKeeper.SaveGrant(s.ctx, grantee, granter, a, &exp1Hour)
	s.Require().NoError(err)

	// A missing signature on a special case message type with an authz grant on that type.
	// Add (a child msg type) TypeURLMsgAddScopeDataAccessRequest  (of a parent) TypeURLMsgWriteScopeRequest
	granter = s.user2Addr
	grantee = s.user3Addr
	a = authz.NewGenericAuthorization(types.TypeURLMsgAddScopeDataAccessRequest)
	err = s.app.AuthzKeeper.SaveGrant(s.ctx, grantee, granter, a, &exp1Hour)
	s.Require().NoError(err)

	// A missing signature on a special case message type with an authz grant on its parent type.
	// Add grant on the parent type of TypeURLMsgAddContractSpecToScopeSpecRequest.
	granter = s.user2Addr
	grantee = s.user3Addr
	a = authz.NewGenericAuthorization(types.TypeURLMsgWriteScopeSpecificationRequest)
	err = s.app.AuthzKeeper.SaveGrant(s.ctx, grantee, granter, a, &exp1Hour)
	s.Require().NoError(err)

	for n, tc := range tests {
		s.T().Run(n, func(t *testing.T) {
			err := s.app.MetadataKeeper.ValidateAllOwnersAreSignersWithAuthz(s.ctx, tc.owners, tc.signers, tc.msgTypeURL)
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
		name                  string
		owners                []string
		signers               []string
		msgTypeURL            string
		allowedAuthorizations int32
		granter               sdk.AccAddress
		grantee               sdk.AccAddress
		errorMsg              string
	}{
		// count authorization test cases
		{
			name:                  "Scope Spec with 2 owners - one signer - with grant - authz",
			owners:                []string{s.user2, s.user3},
			signers:               []string{s.user3},
			msgTypeURL:            types.TypeURLMsgAddScopeOwnerRequest,
			allowedAuthorizations: oneAllowedAuthorizations,
			granter:               s.user2Addr,
			grantee:               s.user3Addr,
			errorMsg:              "",
		},
		{
			name:                  "Scope Spec with 2 owners - one signer - no grant - authz - error",
			owners:                []string{s.user2, s.user3},
			signers:               []string{s.user2},
			msgTypeURL:            types.TypeURLMsgWriteScopeRequest,
			allowedAuthorizations: manyAllowedAuthorizations,
			granter:               s.user2Addr,
			grantee:               s.user3Addr,
			errorMsg:              fmt.Sprintf("missing signature from existing owner %s; required for update", s.user3),
		},
		{
			name:                  "Scope Spec with 3 owners - one signer with a special case message type - with grant - authz",
			owners:                []string{s.user1, s.user2, s.user3},
			signers:               []string{s.user1, s.user3}, // signer 3 is grantee of singer 2
			msgTypeURL:            types.TypeURLMsgAddScopeDataAccessRequest,
			allowedAuthorizations: manyAllowedAuthorizations,
			granter:               s.user2Addr,
			grantee:               s.user3Addr,
			errorMsg:              "",
		},
		{
			name:                  "Scope Spec with 3 owners - two signers with a special case message type - grant on parent of special case message type - authz",
			owners:                []string{s.user1, s.user2, s.user3},
			signers:               []string{s.user1, s.user3}, // signer 3 grantee of signer 2
			msgTypeURL:            types.TypeURLMsgDeleteContractSpecFromScopeSpecRequest,
			allowedAuthorizations: manyAllowedAuthorizations,
			granter:               s.user2Addr,
			grantee:               s.user3Addr,
			errorMsg:              "",
		},
		{
			name:                  "Scope Spec with 2 owners - one signer - no grant - authz - error",
			owners:                []string{s.user2, s.user3},
			signers:               []string{s.user3},
			msgTypeURL:            types.TypeURLMsgDeleteRecordRequest,
			allowedAuthorizations: manyAllowedAuthorizations,
			granter:               nil,
			grantee:               nil,
			errorMsg:              fmt.Sprintf("missing signature from existing owner %s; required for update", s.user2),
		},
	}

	now := s.ctx.BlockHeader().Time
	s.Require().NotNil(now, "now")
	exp1Hour := now.Add(time.Hour)

	// Test cases
	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			createAuth := tc.grantee != nil && tc.granter != nil
			if createAuth {
				a := authz.NewCountAuthorization(tc.msgTypeURL, tc.allowedAuthorizations)
				err := s.app.AuthzKeeper.SaveGrant(s.ctx, tc.grantee, tc.granter, a, &exp1Hour)
				require.NoError(t, err, "SaveGrant", tc.name)
			}

			err := s.app.MetadataKeeper.ValidateAllOwnersAreSignersWithAuthz(s.ctx, tc.owners, tc.signers, tc.msgTypeURL)
			if len(tc.errorMsg) == 0 {
				assert.NoError(t, err, "ValidateAllOwnersAreSigners unexpected error", tc.name)
			} else {
				assert.EqualError(t, err, tc.errorMsg, "ValidateAllOwnersAreSigners error", tc.name)
			}

			// validate allowedAuthorizations
			if err == nil {
				auth, _ := s.app.AuthzKeeper.GetAuthorization(s.ctx, tc.grantee, tc.granter, tc.msgTypeURL)
				if tc.allowedAuthorizations == 1 {
					// authorization is deleted after one use
					assert.Nil(t, auth)
				} else {
					actual := auth.(*authz.CountAuthorization).AllowedAuthorizations
					assert.Equal(t, tc.allowedAuthorizations-1, actual)
				}
			}
		})
	}

	// Special case test
	//
	// with two owners (1 & 2), and one signer (3),
	// with two authz count authorization
	//	- count grants:
	//		granter: 1, grantee: 3, count: 1
	//		granter: 2, grantee: 3, count: 2

	s.T().Run("test with two owners (1 & 2), and one signer (3)", func(t *testing.T) {
		firstGrantAllowedAuthorizations := int32(1)
		secondGrantAllowedAuthorizations := int32(2)
		specialCaseMsgTypeUrl := types.TypeURLMsgDeleteScopeRequest
		// add first grant - with one use (granter: 1, grantee: 3, count: 1)
		a := authz.NewCountAuthorization(specialCaseMsgTypeUrl, firstGrantAllowedAuthorizations)
		err := s.app.AuthzKeeper.SaveGrant(s.ctx, s.user3Addr, s.user1Addr, a, &exp1Hour)
		assert.NoError(t, err, "special case", "SaveGrant")

		// add second grant - with two uses (granter: 2, grantee: 3, count: 2)
		a = authz.NewCountAuthorization(specialCaseMsgTypeUrl, secondGrantAllowedAuthorizations)
		err = s.app.AuthzKeeper.SaveGrant(s.ctx, s.user3Addr, s.user2Addr, a, &exp1Hour)
		assert.NoError(t, err, "special case", "SaveGrant")

		// test with two owners (1 & 2), and one signer (3)
		owners := []string{s.user1, s.user2}
		signers := []string{s.user3}

		// validate signatures
		err = s.app.MetadataKeeper.ValidateAllOwnersAreSignersWithAuthz(s.ctx, owners, signers, specialCaseMsgTypeUrl)
		assert.NoError(t, err, "special case", "ValidateAllOwnersAreSigners")

		// validate first grant is deleted after one use
		auth, _ := s.app.AuthzKeeper.GetAuthorization(s.ctx, s.user3Addr, s.user1Addr, specialCaseMsgTypeUrl)
		assert.Nil(t, auth, "special case", "DeletedAuthorization")

		// validate second grant count is decremented by one after use
		auth, _ = s.app.AuthzKeeper.GetAuthorization(s.ctx, s.user3Addr, s.user2Addr, specialCaseMsgTypeUrl)
		assert.NotNil(t, auth, "special case", "RemainingAuthorization")
		actual := auth.(*authz.CountAuthorization).AllowedAuthorizations
		assert.Equal(t, secondGrantAllowedAuthorizations-1, actual)
	})
}

func (s *AuthzTestSuite) TestValidateAllOwnerPartiesAreSigners() {

	cases := map[string]struct {
		owners     []types.Party
		signers    []string
		msgTypeURL string
		errorMsg   string
	}{
		"no owners - no signers": {
			owners:     []types.Party{},
			signers:    []string{},
			msgTypeURL: "",
			errorMsg:   "",
		},
		"one owner - is signer": {
			owners:     []types.Party{{Address: "signer1", Role: types.PartyType_PARTY_TYPE_OWNER}},
			signers:    []string{"signer1"},
			msgTypeURL: "",
			errorMsg:   "",
		},
		"one owner - is one of two signers": {
			owners:     []types.Party{{Address: "signer1", Role: types.PartyType_PARTY_TYPE_OWNER}},
			signers:    []string{"signer1", "signer2"},
			msgTypeURL: "",
			errorMsg:   "",
		},
		"one owner - is not one of two signers": {
			owners:     []types.Party{{Address: "missingowner", Role: types.PartyType_PARTY_TYPE_OWNER}},
			signers:    []string{"signer1", "signer2"},
			msgTypeURL: "",
			errorMsg:   "missing signature from [missingowner (PARTY_TYPE_OWNER)]",
		},
		"two owners - both are signers": {
			owners: []types.Party{
				{Address: "owner1", Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: "owner2", Role: types.PartyType_PARTY_TYPE_OWNER}},
			msgTypeURL: "",
			signers:    []string{"owner2", "owner1"},
			errorMsg:   "",
		},
		"two owners - only one is signer": {
			owners: []types.Party{
				{Address: "owner1", Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: "missingowner", Role: types.PartyType_PARTY_TYPE_OWNER}},
			signers:    []string{"owner2", "owner1"},
			msgTypeURL: "",
			errorMsg:   "missing signature from [missingowner (PARTY_TYPE_OWNER)]",
		},
		"two parties - one owner one other - only owner is signer": {
			owners: []types.Party{
				{Address: "owner", Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: "affiliate", Role: types.PartyType_PARTY_TYPE_AFFILIATE}},
			signers:    []string{"owner"},
			msgTypeURL: "",
			errorMsg:   "missing signature from [affiliate (PARTY_TYPE_AFFILIATE)]",
		},
		"two parties - one owner one other - only other is signer": {
			owners: []types.Party{
				{Address: "owner", Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: "affiliate", Role: types.PartyType_PARTY_TYPE_AFFILIATE}},
			signers:    []string{"affiliate"},
			msgTypeURL: "",
			errorMsg:   "missing signature from [owner (PARTY_TYPE_OWNER)]",
		},
		// authz test cases
		"two parties - one missing signature with authz grant - two signers": {
			owners: []types.Party{
				{Address: s.user1, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user3, Role: types.PartyType_PARTY_TYPE_OWNER}}, // grantee of singer 1
			signers:    []string{s.user2, s.user3},
			msgTypeURL: types.TypeURLMsgWriteScopeRequest,
			errorMsg:   "",
		},
		"two parties - one missing signature without authz grant - one signer": {
			owners: []types.Party{
				{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user3, Role: types.PartyType_PARTY_TYPE_OWNER}},
			signers:    []string{s.user2},
			msgTypeURL: types.TypeURLMsgWriteScopeRequest,
			errorMsg:   fmt.Sprintf("missing signature from [%s (PARTY_TYPE_OWNER)]", s.user3),
		},
		"two parties - one missing signature with a special case message type - authz grant - two signers": {
			owners: []types.Party{
				{Address: s.user1, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user3, Role: types.PartyType_PARTY_TYPE_OWNER}}, // grantee of singer 2
			signers:    []string{s.user1, s.user3},
			msgTypeURL: types.TypeURLMsgAddScopeDataAccessRequest,
			errorMsg:   "",
		},
		"two parties - one missing signature with a special case message type - authz grant on parent message type - two signers": {
			owners: []types.Party{
				{Address: s.user1, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user3, Role: types.PartyType_PARTY_TYPE_OWNER}}, // grantee of singer 2
			signers:    []string{s.user1, s.user3},
			msgTypeURL: types.TypeURLMsgAddContractSpecToScopeSpecRequest,
			errorMsg:   "",
		},
		"two parties - one missing signature with a special case message type without authz grant - one signer": {
			owners: []types.Party{
				{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user3, Role: types.PartyType_PARTY_TYPE_OWNER}},
			signers:    []string{s.user3},
			msgTypeURL: types.TypeURLMsgDeleteRecordRequest,
			errorMsg:   fmt.Sprintf("missing signature from [%s (PARTY_TYPE_OWNER)]", s.user2),
		},
	}

	// Add a few authorizations
	now := s.ctx.BlockHeader().Time
	s.Require().NotNil(now, "now")
	exp1Hour := now.Add(time.Hour)

	// A missing signature with an authz grant on MsgAddScopeOwnerRequest
	granter := s.user1Addr
	grantee := s.user3Addr
	a := authz.NewGenericAuthorization(types.TypeURLMsgWriteScopeRequest)
	err := s.app.AuthzKeeper.SaveGrant(s.ctx, grantee, granter, a, &exp1Hour)
	s.Require().NoError(err)

	// A missing signature on a special case message type with an authz grant on that type.
	// Add (a child msg type) TypeURLMsgAddScopeDataAccessRequest  (of a parent) TypeURLMsgWriteScopeRequest
	granter = s.user2Addr
	grantee = s.user3Addr
	a = authz.NewGenericAuthorization(types.TypeURLMsgAddScopeDataAccessRequest)
	err = s.app.AuthzKeeper.SaveGrant(s.ctx, grantee, granter, a, &exp1Hour)
	s.Require().NoError(err)

	// A missing signature on a special case message type with an authz grant on its parent type.
	// Add grant on the parent type of TypeURLMsgAddContractSpecToScopeSpecRequest.
	granter = s.user2Addr
	grantee = s.user3Addr
	a = authz.NewGenericAuthorization(types.TypeURLMsgWriteScopeSpecificationRequest)
	err = s.app.AuthzKeeper.SaveGrant(s.ctx, grantee, granter, a, &exp1Hour)
	s.Require().NoError(err)

	// Test cases
	for n, tc := range cases {
		s.T().Run(n, func(t *testing.T) {
			err := s.app.MetadataKeeper.ValidateAllPartiesAreSignersWithAuthz(s.ctx, tc.owners, tc.signers, tc.msgTypeURL)
			if len(tc.errorMsg) == 0 {
				assert.NoError(t, err, "%s unexpected error", n)
			} else {
				assert.EqualError(t, err, tc.errorMsg, "%s error", n)
			}
		})
	}
}

func (s *AuthzTestSuite) TestValidateAllOwnerPartiesAreSignersWithCountAuthorization() {

	oneAllowedAuthorizations := int32(1)
	manyAllowedAuthorizations := int32(10)

	cases := []struct {
		name                  string
		owners                []types.Party
		signers               []string
		msgTypeURL            string
		allowedAuthorizations int32
		granter               sdk.AccAddress
		grantee               sdk.AccAddress
		errorMsg              string
	}{
		// count authorization test cases
		{
			name: "three parties - one missing signature with one authz grant - two signers",
			owners: []types.Party{
				{Address: s.user1, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user3, Role: types.PartyType_PARTY_TYPE_OWNER}}, // grantee of singer 1
			signers:               []string{s.user2, s.user3},
			msgTypeURL:            types.TypeURLMsgWriteScopeRequest,
			allowedAuthorizations: oneAllowedAuthorizations,
			granter:               s.user1Addr,
			grantee:               s.user3Addr,
			errorMsg:              "",
		},
		{
			name: "three parties - one missing signature with a special case message type - authz grant - two signers",
			owners: []types.Party{
				{Address: s.user1, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user3, Role: types.PartyType_PARTY_TYPE_OWNER}}, // grantee of singer 2
			signers:               []string{s.user1, s.user3},
			msgTypeURL:            types.TypeURLMsgAddScopeDataAccessRequest,
			allowedAuthorizations: manyAllowedAuthorizations,
			granter:               s.user2Addr,
			grantee:               s.user3Addr,
			errorMsg:              "",
		},
		{
			name: "three parties - one missing signature with a special case message type - authz grant on parent message type - two signers",
			owners: []types.Party{
				{Address: s.user1, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user3, Role: types.PartyType_PARTY_TYPE_OWNER}}, // grantee of singer 2
			signers:               []string{s.user1, s.user3},
			msgTypeURL:            types.TypeURLMsgAddContractSpecToScopeSpecRequest,
			allowedAuthorizations: manyAllowedAuthorizations,
			granter:               s.user2Addr,
			grantee:               s.user3Addr,
			errorMsg:              "",
		},
		{
			name: "two parties - one missing signature with a special case message type without authz grant - one signer",
			owners: []types.Party{
				{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user3, Role: types.PartyType_PARTY_TYPE_OWNER}},
			signers:               []string{s.user3},
			msgTypeURL:            types.TypeURLMsgDeleteRecordRequest,
			allowedAuthorizations: manyAllowedAuthorizations,
			granter:               nil,
			grantee:               nil,
			errorMsg:              fmt.Sprintf("missing signature from [%s (PARTY_TYPE_OWNER)]", s.user2),
		},
	}

	now := s.ctx.BlockHeader().Time
	s.Require().NotNil(now, "now")
	exp1Hour := now.Add(time.Hour)

	// Test cases
	for _, tc := range cases {
		s.T().Run(tc.name, func(t *testing.T) {
			createAuth := tc.grantee != nil && tc.granter != nil
			if createAuth {
				a := authz.NewCountAuthorization(tc.msgTypeURL, tc.allowedAuthorizations)
				err := s.app.AuthzKeeper.SaveGrant(s.ctx, tc.grantee, tc.granter, a, &exp1Hour)
				s.Require().NoError(err)
			}

			err := s.app.MetadataKeeper.ValidateAllPartiesAreSignersWithAuthz(s.ctx, tc.owners, tc.signers, tc.msgTypeURL)
			if len(tc.errorMsg) == 0 {
				assert.NoError(t, err, "%s unexpected error", tc.name)
			} else {
				assert.EqualError(t, err, tc.errorMsg, "%s error", tc.name)
			}

			// validate allowedAuthorizations
			if err == nil {
				auth, _ := s.app.AuthzKeeper.GetAuthorization(s.ctx, tc.grantee, tc.granter, tc.msgTypeURL)
				if tc.allowedAuthorizations == 1 {
					// authorization is deleted after one use
					assert.Nil(t, auth)
				} else {
					actual := auth.(*authz.CountAuthorization).AllowedAuthorizations
					assert.Equal(t, tc.allowedAuthorizations-1, actual)
				}
			}
		})
	}

	// Special case test
	//
	// with two parties (1 & 2), and one signer (3),
	// with two authz count authorization
	//	- count grants:
	//		granter: 1, grantee: 3, count: 1
	//		granter: 2, grantee: 3, count: 2

	s.T().Run("test with two owners (1 & 2), and one signer (3)", func(t *testing.T) {
		firstGrantAllowedAuthorizations := int32(1)
		secondGrantAllowedAuthorizations := int32(2)
		specialCaseMsgTypeUrl := types.TypeURLMsgDeleteScopeRequest
		// add first grant - with one use (granter: 1, grantee: 3, count: 1)
		a := authz.NewCountAuthorization(specialCaseMsgTypeUrl, firstGrantAllowedAuthorizations)
		err := s.app.AuthzKeeper.SaveGrant(s.ctx, s.user3Addr, s.user1Addr, a, &exp1Hour)
		assert.NoError(t, err, "special case", "SaveGrant")

		// add second grant - with two uses (granter: 2, grantee: 3, count: 2)
		a = authz.NewCountAuthorization(specialCaseMsgTypeUrl, secondGrantAllowedAuthorizations)
		err = s.app.AuthzKeeper.SaveGrant(s.ctx, s.user3Addr, s.user2Addr, a, &exp1Hour)
		assert.NoError(t, err, "special case", "SaveGrant")

		// test with two parties (1 & 2), and one signer (3)
		parties := []types.Party{
			{Address: s.user1, Role: types.PartyType_PARTY_TYPE_OWNER},
			{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER}}
		signers := []string{s.user3}

		// validate signatures
		err = s.app.MetadataKeeper.ValidateAllPartiesAreSignersWithAuthz(s.ctx, parties, signers, specialCaseMsgTypeUrl)
		assert.NoError(t, err, "special case", "ValidateAllPartiesAreSigners")

		// validate first grant is deleted after one use
		auth, _ := s.app.AuthzKeeper.GetAuthorization(s.ctx, s.user3Addr, s.user1Addr, specialCaseMsgTypeUrl)
		assert.Nil(t, auth, "special case", "DeletedAuthorization")

		// validate second grant count is decremented by one after use
		auth, _ = s.app.AuthzKeeper.GetAuthorization(s.ctx, s.user3Addr, s.user2Addr, specialCaseMsgTypeUrl)
		assert.NotNil(t, auth, "special case", "RemainingAuthorization")
		actual := auth.(*authz.CountAuthorization).AllowedAuthorizations
		assert.Equal(t, secondGrantAllowedAuthorizations-1, actual)
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
