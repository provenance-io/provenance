package keeper_test

import (
	"fmt"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"testing"

	"github.com/provenance-io/provenance/app"
	simapp "github.com/provenance-io/provenance/app"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/provenance-io/provenance/x/metadata/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app         *app.App
	ctx         sdk.Context
	queryClient types.QueryClient

	pubkey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress

	pubkey2   cryptotypes.PubKey
	user2     string
	user2Addr sdk.AccAddress
}

func (s *KeeperTestSuite) SetupTest() {
	s.app = simapp.Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, s.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, s.app.MetadataKeeper)
	s.queryClient = types.NewQueryClient(queryHelper)

	s.pubkey1 = secp256k1.GenPrivKey().PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()
	s.app.AccountKeeper.SetAccount(s.ctx, s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user1Addr))

	s.pubkey2 = secp256k1.GenPrivKey().PubKey()
	s.user2Addr = sdk.AccAddress(s.pubkey2.Address())
	s.user2 = s.user2Addr.String()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// ownerPartyList returns a party with role OWNER for each address provided.
// This func is used in other keeper test files.
func ownerPartyList(addresses ...string) []types.Party {
	retval := make([]types.Party, len(addresses))
	for i, addr := range addresses {
		retval[i] = types.Party{Address: addr, Role: types.PartyType_PARTY_TYPE_OWNER}
	}
	return retval
}

func (s *KeeperTestSuite) TestValidatePartiesInvolved() {

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
			errorMsg:        "missing required party type PARTY_TYPE_AFFILIATE from parties",
		},
		"invalid, missing required parties": {
			parties:         []types.Party{{Address: "address", Role: types.PartyType_PARTY_TYPE_CUSTODIAN}},
			requiredParties: []types.PartyType{types.PartyType_PARTY_TYPE_AFFILIATE},
			wantErr:         true,
			errorMsg:        "missing required party type PARTY_TYPE_AFFILIATE from parties",
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

func (s *KeeperTestSuite) TestValidateAllOwnerPartiesAreSigners() {

	cases := map[string]struct {
		owners   []types.Party
		signers  []string
		errorMsg string
	}{
		"no owners - no signers": {
			owners:   []types.Party{},
			signers:  []string{},
			errorMsg: "",
		},
		"one owner - is signer": {
			owners:   []types.Party{{Address: "signer1", Role: types.PartyType_PARTY_TYPE_OWNER}},
			signers:  []string{"signer1"},
			errorMsg: "",
		},
		"one owner - is one of two signers": {
			owners:   []types.Party{{Address: "signer1", Role: types.PartyType_PARTY_TYPE_OWNER}},
			signers:  []string{"signer1", "signer2"},
			errorMsg: "",
		},
		"one owner - is not one of two signers": {
			owners:   []types.Party{{Address: "missingowner", Role: types.PartyType_PARTY_TYPE_OWNER}},
			signers:  []string{"signer1", "signer2"},
			errorMsg: "missing signature from missingowner (PARTY_TYPE_OWNER)",
		},
		"two owners - both are signers": {
			owners:   []types.Party{
				{Address: "owner1", Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: "owner2", Role: types.PartyType_PARTY_TYPE_OWNER}},
			signers:  []string{"owner2", "owner1"},
			errorMsg: "",
		},
		"two owners - only one is signer": {
			owners:   []types.Party{
				{Address: "owner1", Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: "missingowner", Role: types.PartyType_PARTY_TYPE_OWNER}},
			signers:  []string{"owner2", "owner1"},
			errorMsg: "missing signature from missingowner (PARTY_TYPE_OWNER)",
		},
		"two parties - one owner one other - only owner is signer": {
			owners:   []types.Party{
				{Address: "owner", Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: "affiliate", Role: types.PartyType_PARTY_TYPE_AFFILIATE}},
			signers:  []string{"owner"},
			errorMsg: "missing signature from affiliate (PARTY_TYPE_AFFILIATE)",
		},
		"two parties - one owner one other - only other is signer": {
			owners:   []types.Party{
				{Address: "owner", Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: "affiliate", Role: types.PartyType_PARTY_TYPE_AFFILIATE}},
			signers:  []string{"affiliate"},
			errorMsg: "missing signature from owner (PARTY_TYPE_OWNER)",
		},
	}

	for n, tc := range cases {
		s.T().Run(n, func(t *testing.T) {
			err := s.app.MetadataKeeper.ValidateAllPartiesAreSigners(tc.owners, tc.signers)
			if len(tc.errorMsg) == 0 {
				assert.NoError(t, err, "%s unexpected error", n)
			} else {
				assert.EqualError(t, err, tc.errorMsg, "%s error", n)
			}
		})
	}
}

func (s *KeeperTestSuite) TestValidateAllOwnersAreSigners() {
	tests := map[string]struct {
		owners   []string
		signers  []string
		errorMsg string
	}{
		"Scope Spec with 1 owner: no signers - error": {
			[]string{s.user1},
			[]string{},
			fmt.Sprintf("missing signature from existing owner %s; required for update", s.user1),
		},
		"Scope Spec with 1 owner: not in signers list - error": {
			[]string{s.user1},
			[]string{sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()},
			fmt.Sprintf("missing signature from existing owner %s; required for update", s.user1),
		},
		"Scope Spec with 1 owner: in signers list with non-owners - ok": {
			[]string{s.user1},
			[]string{sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), s.user1, sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()},
			"",
		},
		"Scope Spec with 1 owner: only signer in list - ok": {
			[]string{s.user1},
			[]string{s.user1},
			"",
		},
		"Scope Spec with 2 owners: no signers - error": {
			[]string{s.user1, s.user2},
			[]string{},
			fmt.Sprintf("missing signatures from existing owners %v; required for update",
				[]string{s.user1, s.user2}),
		},
		"Scope Spec with 2 owners: neither in signers list - error": {
			[]string{s.user1, s.user2},
			[]string{sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()},
			fmt.Sprintf("missing signatures from existing owners %v; required for update",
				[]string{s.user1, s.user2}),
		},
		"Scope Spec with 2 owners: one in signers list with non-owners - error": {
			[]string{s.user1, s.user2},
			[]string{sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), s.user1, sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()},
			fmt.Sprintf("missing signature from existing owner %s; required for update", s.user2),
		},
		"Scope Spec with 2 owners: the other in signers list with non-owners - error": {
			[]string{s.user1, s.user2},
			[]string{sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), s.user2, sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()},
			fmt.Sprintf("missing signature from existing owner %s; required for update", s.user1),
		},
		"Scope Spec with 2 owners: both in signers list with non-owners - ok": {
			[]string{s.user1, s.user2},
			[]string{sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), s.user2, sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), s.user1},
			"",
		},
		"Scope Spec with 2 owners: only both in signers list - ok": {
			[]string{s.user1, s.user2},
			[]string{s.user1, s.user2},
			"",
		},
		"Scope Spec with 2 owners: only both in signers list, opposite order - ok": {
			[]string{s.user1, s.user2},
			[]string{s.user2, s.user1},
			"",
		},
	}

	for n, tc := range tests {
		s.T().Run(n, func(t *testing.T) {
			err := s.app.MetadataKeeper.ValidateAllOwnersAreSigners(tc.owners, tc.signers)
			if len(tc.errorMsg) == 0 {
				assert.NoError(t, err, "ValidateAllOwnersAreSigners unexpected error")
			} else {
				assert.EqualError(t, err, tc.errorMsg, "ValidateAllOwnersAreSigners error")
			}
		})
	}
}
