package keeper_test

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256r1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/metadata/keeper"
	"github.com/provenance-io/provenance/x/metadata/types"
)

type MsgServerTestSuite struct {
	suite.Suite

	app       *app.App
	ctx       sdk.Context
	msgServer types.MsgServer

	pubkey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress

	pubkey2   cryptotypes.PubKey
	user2     string
	user2Addr sdk.AccAddress
}

func (s *MsgServerTestSuite) SetupTest() {
	s.app = app.Setup(s.T())
	s.ctx = keeper.AddAuthzCacheToContext(s.app.BaseApp.NewContext(false))
	s.msgServer = keeper.NewMsgServerImpl(s.app.MetadataKeeper)

	s.pubkey1 = secp256k1.GenPrivKey().PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()
	user1Acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user1Addr)
	s.Require().NoError(user1Acc.SetPubKey(s.pubkey1), "SetPubKey user1")

	privKey, _ := secp256r1.GenPrivKey()
	s.pubkey2 = privKey.PubKey()
	s.user2Addr = sdk.AccAddress(s.pubkey2.Address())
	s.user2 = s.user2Addr.String()
	user2Acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user2Addr)
	s.Require().NoError(user2Acc.SetPubKey(s.pubkey1), "SetPubKey user2")

	s.app.AccountKeeper.SetAccount(s.ctx, user1Acc)
	s.app.AccountKeeper.SetAccount(s.ctx, user2Acc)
}

func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerTestSuite))
}

// AssertErrorValue asserts that:
//   - If errorString is empty, theError must be nil
//   - If errorString is not empty, theError must equal the errorString.
func (s *MsgServerTestSuite) AssertErrorValue(theError error, errorString string, msgAndArgs ...interface{}) bool {
	return AssertErrorValue(s.T(), theError, errorString, msgAndArgs...)
}

// newUUID will create a new UUID using the provided name and index to define the bytes.
func (s *MsgServerTestSuite) newUUID(name string, i int) uuid.UUID {
	s.T().Helper()
	str := fmt.Sprintf("%d_%s", i, name)
	if len(str) > 16 {
		s.FailNowf("cannot newUUID(%q, %d): base string %q is longer than 16 bytes", name, i, str)
	}
	if len(str) < 16 {
		str = str + strings.Repeat("_", 16-len(str))
	}
	rv, err := uuid.FromBytes([]byte(str))
	s.Require().NoError(err, "uuid.FromBytes([]byte(%q))", str)
	return rv
}

// namedValue is a way to associate a variable name with its value for use with logNamedValues.
type namedValue struct {
	name  string
	value string
}

// logNamedValues will log the provided entries under the given header.
func (s *MsgServerTestSuite) logNamedValues(header string, entries []namedValue) {
	// Note: This func might not be called by checked-in code, but it's handy when troubleshooting, so don't delete it.
	logNamedValues(s.T(), header, entries)
}

// logNamedValues will log the provided entries under the given header.
func logNamedValues(t *testing.T, header string, entries []namedValue) {
	lines := make([]string, len(entries))
	for i, entry := range entries {
		lines[i] = fmt.Sprintf("%20s = %s", entry.name, entry.value)
	}
	t.Logf("%s:\n%s", header, strings.Join(lines, "\n"))
}

// MakeNonWasmAccount will make sure the account with the provided bech32 string has a sequence of 1.
// This will cause the isWasmAccount test to report that the account is NOT a wasm account.
func (s *MsgServerTestSuite) MakeNonWasmAccounts(bech32s ...string) {
	s.T().Helper()
	for i, bech32 := range bech32s {
		addr, err := sdk.AccAddressFromBech32(bech32)
		s.Require().NoError(err, "[%d]: sdk.AccAddressFromBech32(%q)", i, bech32)
		acct := s.app.AccountKeeper.GetAccount(s.ctx, addr)
		if acct == nil {
			acct = s.app.AccountKeeper.NewAccountWithAddress(s.ctx, addr)
		}
		if acct.GetSequence() == uint64(0) {
			s.Require().NoError(acct.SetSequence(1), "[%d]: %q.SetSequence(1)", i, bech32)
			s.app.AccountKeeper.SetAccount(s.ctx, acct)
		}
	}
}

// TODO: WriteScope tests
// TODO: DeleteScope tests

func (s *MsgServerTestSuite) TestAddAndDeleteScopeDataAccess() {
	scopeSpecID := types.ScopeSpecMetadataAddress(uuid.New())
	scopeSpec := types.NewScopeSpecification(scopeSpecID, nil, []string{s.user1}, []types.PartyType{types.PartyType_PARTY_TYPE_OWNER}, []types.MetadataAddress{})
	scopeID := types.ScopeMetadataAddress(uuid.New())
	scope := types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{s.user1}, "", false)
	dneScopeID := types.ScopeMetadataAddress(uuid.New())
	user3 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()

	scopeSpecMsg := types.NewMsgWriteScopeSpecificationRequest(*scopeSpec, []string{s.user1})
	_, err := s.msgServer.WriteScopeSpecification(s.ctx, scopeSpecMsg)
	s.Assert().NoError(err, "setup test with new scope specification")

	writeScopeMsg := types.NewMsgWriteScopeRequest(*scope, []string{s.user1}, 0)
	_, err = s.msgServer.WriteScope(s.ctx, writeScopeMsg)
	s.Assert().NoError(err, "setup test with new scope")

	cases := []struct {
		name     string
		addMsg   *types.MsgAddScopeDataAccessRequest
		delMsg   *types.MsgDeleteScopeDataAccessRequest
		signers  []string
		errorMsg string
	}{

		{
			name:     "should fail to ADD address to data access, msg validate basic failure",
			addMsg:   types.NewMsgAddScopeDataAccessRequest(scopeID, []string{}, []string{s.user1}),
			signers:  []string{s.user1},
			errorMsg: "data access list cannot be empty: invalid request",
		},
		{
			name:     "should fail to ADD address to data access, validate add failure",
			addMsg:   types.NewMsgAddScopeDataAccessRequest(dneScopeID, []string{s.user1}, []string{s.user1}),
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("scope not found with id %s: not found", dneScopeID),
		},
		{
			name:     "should fail to ADD address to data access, validate add failure",
			addMsg:   types.NewMsgAddScopeDataAccessRequest(scopeID, []string{s.user1}, []string{s.user1}),
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("address already exists for data access %s: invalid request", s.user1),
		},
		{
			name:    "should successfully ADD address to data access",
			addMsg:  types.NewMsgAddScopeDataAccessRequest(scopeID, []string{s.user2}, []string{s.user1}),
			signers: []string{s.user1},
		},
		{
			name:     "should fail to DELETE address from data access, msg validate basic failure",
			delMsg:   types.NewMsgDeleteScopeDataAccessRequest(scopeID, []string{}, []string{s.user1}),
			signers:  []string{s.user1},
			errorMsg: "data access list cannot be empty: invalid request",
		},
		{
			name:     "should fail to DELETE address from data access, validate add failure",
			delMsg:   types.NewMsgDeleteScopeDataAccessRequest(dneScopeID, []string{s.user1}, []string{s.user1}),
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("scope not found with id %s: not found", dneScopeID),
		},
		{
			name:     "should fail to DELETE address from data access, validate add failure",
			delMsg:   types.NewMsgDeleteScopeDataAccessRequest(scopeID, []string{user3}, []string{s.user1}),
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("address does not exist in scope data access: %s: invalid request", user3),
		},
		{
			name:    "should successfully DELETE address from data access",
			delMsg:  types.NewMsgDeleteScopeDataAccessRequest(scopeID, []string{s.user2}, []string{s.user1}),
			signers: []string{s.user1},
		},
	}

	for _, tc := range cases {
		s.T().Run(tc.name, func(t *testing.T) {
			var err error
			if tc.delMsg != nil {
				_, err = s.msgServer.DeleteScopeDataAccess(s.ctx, tc.delMsg)
			}
			if tc.addMsg != nil {
				_, err = s.msgServer.AddScopeDataAccess(s.ctx, tc.addMsg)
			}

			if len(tc.errorMsg) > 0 {
				assert.EqualError(t, err, tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	s.T().Run("data access actually deleted and added", func(t *testing.T) {
		addrOriginator := "cosmos1rr4d0eu62pgt4edw38d2ev27798pfhdhm39zct"
		addrServicer := "cosmos1a7mmtar5ke5fxk5gn00dlag0zfmdkmhapmugk7"
		scopeA := types.Scope{
			ScopeId:           types.ScopeMetadataAddress(uuid.New()),
			SpecificationId:   types.ScopeSpecMetadataAddress(uuid.New()),
			DataAccess:        []string{addrOriginator, addrServicer},
			ValueOwnerAddress: addrServicer,
			Owners: []types.Party{
				{
					Address: addrOriginator,
					Role:    types.PartyType_PARTY_TYPE_ORIGINATOR,
				},
			},
		}

		scopeSpecA := types.ScopeSpecification{
			SpecificationId: scopeA.SpecificationId,
			Description: &types.Description{
				Name:        "com.figure.origination.loan",
				Description: "Figure loan origination",
			},
			OwnerAddresses:  []string{"cosmos1q8n4v4m0hm8v0a7n697nwtpzhfsz3f4d40lnsu"},
			PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_ORIGINATOR},
			ContractSpecIds: nil,
		}

		s.app.MetadataKeeper.SetScope(s.ctx, scopeA)
		s.app.MetadataKeeper.SetScopeSpecification(s.ctx, scopeSpecA)

		msgDel := types.NewMsgDeleteScopeDataAccessRequest(
			scopeA.ScopeId,
			[]string{addrServicer},
			[]string{addrOriginator},
		)

		_, errDel := s.msgServer.DeleteScopeDataAccess(s.ctx, msgDel)
		require.NoError(t, errDel, "Failed to make DeleteScopeDataAccessRequest call")

		scopeB, foundB := s.app.MetadataKeeper.GetScope(s.ctx, scopeA.ScopeId)
		require.Truef(t, foundB, "Scope %s not found after DeleteScopeOwnerRequest call.", scopeA.ScopeId)

		assert.Equal(t, scopeA.ScopeId, scopeB.ScopeId, "del ScopeId")
		assert.Equal(t, scopeA.SpecificationId, scopeB.SpecificationId, "del SpecificationId")
		assert.Equal(t, scopeA.DataAccess[0:1], scopeB.DataAccess, "del DataAccess")
		assert.Equal(t, scopeA.ValueOwnerAddress, scopeB.ValueOwnerAddress, "del ValueOwnerAddress")
		assert.Equal(t, scopeA.Owners, scopeB.Owners, "del Owners")

		// Stop test if it's already failed.
		if t.Failed() {
			t.FailNow()
		}

		msgAdd := types.NewMsgAddScopeDataAccessRequest(
			scopeA.ScopeId,
			[]string{addrServicer},
			[]string{addrOriginator},
		)

		_, errAdd := s.msgServer.AddScopeDataAccess(s.ctx, msgAdd)
		require.NoError(t, errAdd, "Failed to make AddScopeDataAccessRequest call")

		scopeC, foundC := s.app.MetadataKeeper.GetScope(s.ctx, scopeA.ScopeId)
		require.Truef(t, foundC, "Scope %s not found after AddScopeOwnerRequest call.", scopeA.ScopeId)

		assert.Equal(t, scopeA.ScopeId, scopeC.ScopeId, "add ScopeId")
		assert.Equal(t, scopeA.SpecificationId, scopeC.SpecificationId, "add SpecificationId")
		assert.Equal(t, scopeA.DataAccess, scopeC.DataAccess, "add DataAccess")
		assert.Equal(t, scopeA.ValueOwnerAddress, scopeC.ValueOwnerAddress, "add ValueOwnerAddress")
		assert.Equal(t, scopeA.Owners, scopeC.Owners, "add Owners")
	})
}

func (s *MsgServerTestSuite) TestAddAndDeleteScopeOwners() {
	scopeSpecID := types.ScopeSpecMetadataAddress(uuid.New())
	scopeSpec := types.NewScopeSpecification(scopeSpecID, nil, []string{s.user1}, []types.PartyType{types.PartyType_PARTY_TYPE_OWNER}, []types.MetadataAddress{})
	scopeID := types.ScopeMetadataAddress(uuid.New())
	scope := types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{s.user1}, "", false)
	dneScopeID := types.ScopeMetadataAddress(uuid.New())
	user3 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()

	scopeSpecMsg := types.NewMsgWriteScopeSpecificationRequest(*scopeSpec, []string{s.user1})
	_, err := s.msgServer.WriteScopeSpecification(s.ctx, scopeSpecMsg)
	s.Assert().NoError(err, "setup test with new scope specification")

	writeScopeMsg := types.NewMsgWriteScopeRequest(*scope, []string{s.user1}, 0)
	_, err = s.msgServer.WriteScope(s.ctx, writeScopeMsg)
	s.Assert().NoError(err, "setup test with new scope")

	cases := []struct {
		name     string
		addMsg   *types.MsgAddScopeOwnerRequest
		delMsg   *types.MsgDeleteScopeOwnerRequest
		signers  []string
		errorMsg string
	}{
		{
			name:     "should fail to ADD owners, msg validate basic failure",
			addMsg:   types.NewMsgAddScopeOwnerRequest(scopeID, []types.Party{}, []string{s.user1}),
			signers:  []string{s.user1},
			errorMsg: "invalid owners: at least one party is required: invalid request",
		},
		{
			name:     "should fail to ADD owners, can not find scope",
			addMsg:   types.NewMsgAddScopeOwnerRequest(dneScopeID, []types.Party{{Address: s.user1, Role: types.PartyType_PARTY_TYPE_OWNER}}, []string{s.user1}),
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("scope not found with id %s: not found", dneScopeID),
		},
		{
			name:     "should fail to ADD owners, validate add failure",
			addMsg:   types.NewMsgAddScopeOwnerRequest(scopeID, []types.Party{{Address: s.user1, Role: types.PartyType_PARTY_TYPE_OWNER}}, []string{s.user1}),
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("party already exists with address %s and role %s", s.user1, types.PartyType_PARTY_TYPE_OWNER),
		},
		{
			name:    "should successfully ADD owners",
			addMsg:  types.NewMsgAddScopeOwnerRequest(scopeID, []types.Party{{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER}}, []string{s.user1}),
			signers: []string{s.user1},
		},
		{
			name:     "should fail to DELETE owners, msg validate basic failure",
			delMsg:   types.NewMsgDeleteScopeOwnerRequest(scopeID, []string{}, []string{s.user1, s.user2}),
			signers:  []string{s.user1},
			errorMsg: "at least one owner address is required: invalid request",
		},
		{
			name:     "should fail to DELETE owners, validate add failure",
			delMsg:   types.NewMsgDeleteScopeOwnerRequest(dneScopeID, []string{s.user1}, []string{s.user1, s.user2}),
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("scope not found with id %s: not found", dneScopeID),
		},
		{
			name:     "should fail to DELETE owners, validate add failure",
			delMsg:   types.NewMsgDeleteScopeOwnerRequest(scopeID, []string{user3}, []string{s.user1, s.user2}),
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("address does not exist in scope owners: %s", user3),
		},
		{
			name:    "should successfully DELETE owners",
			delMsg:  types.NewMsgDeleteScopeOwnerRequest(scopeID, []string{s.user2}, []string{s.user1, s.user2}),
			signers: []string{s.user1},
		},
	}

	for _, tc := range cases {
		s.T().Run(tc.name, func(t *testing.T) {
			var err error
			if tc.delMsg != nil {
				_, err = s.msgServer.DeleteScopeOwner(s.ctx, tc.delMsg)
			}
			if tc.addMsg != nil {
				_, err = s.msgServer.AddScopeOwner(s.ctx, tc.addMsg)
			}
			if len(tc.errorMsg) > 0 {
				assert.EqualError(t, err, tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	s.T().Run("owner actually deleted and added", func(t *testing.T) {
		addrOriginator := "cosmos1rr4d0eu62pgt4edw38d2ev27798pfhdhm39zct"
		addrServicer := "cosmos1a7mmtar5ke5fxk5gn00dlag0zfmdkmhapmugk7"
		scopeA := types.Scope{
			ScopeId:           types.ScopeMetadataAddress(uuid.New()),
			SpecificationId:   types.ScopeSpecMetadataAddress(uuid.New()),
			DataAccess:        []string{addrOriginator, addrServicer},
			ValueOwnerAddress: addrServicer,
			Owners: []types.Party{
				{
					Address: addrOriginator,
					Role:    types.PartyType_PARTY_TYPE_ORIGINATOR,
				},
				{
					Address: addrServicer,
					Role:    types.PartyType_PARTY_TYPE_SERVICER,
				},
			},
		}

		scopeSpecA := types.ScopeSpecification{
			SpecificationId: scopeA.SpecificationId,
			Description: &types.Description{
				Name:        "com.figure.origination.loan",
				Description: "Figure loan origination",
			},
			OwnerAddresses:  []string{"cosmos1q8n4v4m0hm8v0a7n697nwtpzhfsz3f4d40lnsu"},
			PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_ORIGINATOR},
			ContractSpecIds: nil,
		}

		s.app.MetadataKeeper.SetScope(s.ctx, scopeA)
		s.app.MetadataKeeper.SetScopeSpecification(s.ctx, scopeSpecA)

		msgDel := types.NewMsgDeleteScopeOwnerRequest(
			scopeA.ScopeId,
			[]string{addrServicer},
			[]string{addrOriginator, addrServicer},
		)

		_, errDel := s.msgServer.DeleteScopeOwner(s.ctx, msgDel)
		require.NoError(t, errDel, "Failed to make DeleteScopeOwnerRequest call")

		scopeB, foundB := s.app.MetadataKeeper.GetScope(s.ctx, scopeA.ScopeId)
		require.Truef(t, foundB, "Scope %s not found after DeleteScopeOwnerRequest call.", scopeA.ScopeId)

		assert.Equal(t, scopeA.ScopeId, scopeB.ScopeId, "del ScopeId")
		assert.Equal(t, scopeA.SpecificationId, scopeB.SpecificationId, "del SpecificationId")
		assert.Equal(t, scopeA.DataAccess, scopeB.DataAccess, "del DataAccess")
		assert.Equal(t, scopeA.ValueOwnerAddress, scopeB.ValueOwnerAddress, "del ValueOwnerAddress")
		assert.Equal(t, scopeA.Owners[0:1], scopeB.Owners, "del Owners")

		// Stop test if it's already failed.
		if t.Failed() {
			t.FailNow()
		}

		msgAdd := types.NewMsgAddScopeOwnerRequest(
			scopeA.ScopeId,
			[]types.Party{{Address: addrServicer, Role: types.PartyType_PARTY_TYPE_SERVICER}},
			[]string{addrOriginator},
		)

		_, errAdd := s.msgServer.AddScopeOwner(s.ctx, msgAdd)
		require.NoError(t, errAdd, "Failed to make DeleteScopeOwnerRequest call")

		scopeC, foundC := s.app.MetadataKeeper.GetScope(s.ctx, scopeA.ScopeId)
		require.Truef(t, foundC, "Scope %s not found after AddScopeOwnerRequest call.", scopeA.ScopeId)

		assert.Equal(t, scopeA.ScopeId, scopeC.ScopeId, "add ScopeId")
		assert.Equal(t, scopeA.SpecificationId, scopeC.SpecificationId, "add SpecificationId")
		assert.Equal(t, scopeA.DataAccess, scopeC.DataAccess, "add DataAccess")
		assert.Equal(t, scopeA.ValueOwnerAddress, scopeC.ValueOwnerAddress, "add ValueOwnerAddress")
		assert.Equal(t, scopeA.Owners, scopeC.Owners, "add Owners")
	})
}

func (s *MsgServerTestSuite) TestUpdateValueOwners() {
	scopeID1 := types.ScopeMetadataAddress(s.newUUID("scope", 1))             // scope1qqc47umrdacx2h6lta047h6lta0sfyvr90
	scopeID2 := types.ScopeMetadataAddress(s.newUUID("scope", 2))             // scope1qqe97umrdacx2h6lta047h6lta0sk6uj0g
	scopeIDNotFound := types.ScopeMetadataAddress(s.newUUID("notfound", 0))   // scope1qqc97mn0w3nx7atwv3047h6lta0sylrdee
	scopeID3Diff1 := types.ScopeMetadataAddress(s.newUUID("scope_3_diff", 1)) // scope1qqc47umrdacx2hentajxjenxta0sp32qg7
	scopeID3Diff2 := types.ScopeMetadataAddress(s.newUUID("scope_3_diff", 2)) // scope1qqe97umrdacx2hentajxjenxta0s7063ze
	scopeID3Diff3 := types.ScopeMetadataAddress(s.newUUID("scope_3_diff", 3)) // scope1qqe47umrdacx2hentajxjenxta0szju7u4
	scopeID3Same1 := types.ScopeMetadataAddress(s.newUUID("scope_3_same", 1)) // scope1qqc47umrdacx2hentaekzmt9ta0sc9gw9t
	scopeID3Same2 := types.ScopeMetadataAddress(s.newUUID("scope_3_same", 2)) // scope1qqe97umrdacx2hentaekzmt9ta0s8mcl0v
	scopeID3Same3 := types.ScopeMetadataAddress(s.newUUID("scope_3_same", 3)) // scope1qqe47umrdacx2hentaekzmt9ta0smx7s3q

	owner1 := sdk.AccAddress("owner1______________").String()      // cosmos1damkuetjx9047h6lta047h6lta047h6lccgedl
	owner2 := sdk.AccAddress("owner2______________").String()      // cosmos1damkuetjxf047h6lta047h6lta047h6lsp5nql
	owner3Diff1 := sdk.AccAddress("owner3Diff1_________").String() // cosmos1damkuetjxdzxjenxx9047h6lta047h6lx6slvt
	owner3Diff2 := sdk.AccAddress("owner3Diff2_________").String() // cosmos1damkuetjxdzxjenxxf047h6lta047h6l955tqt
	owner3Diff3 := sdk.AccAddress("owner3Diff3_________").String() // cosmos1damkuetjxdzxjenxxd047h6lta047h6lyf08yt
	owner3Same1 := sdk.AccAddress("owner3Same1_________").String() // cosmos1damkuetjxdfkzmt9x9047h6lta047h6l4tvcmn
	owner3Same2 := sdk.AccAddress("owner3Same2_________").String() // cosmos1damkuetjxdfkzmt9xf047h6lta047h6lk9gvhn
	owner3Same3 := sdk.AccAddress("owner3Same3_________").String() // cosmos1damkuetjxdfkzmt9xd047h6lta047h6lhcnqnn

	dataAccess1 := sdk.AccAddress("dataAccess1_________").String()      // cosmos1v3shgc2pvd3k2umnx9047h6lta047h6lvp7hkw
	dataAccess2 := sdk.AccAddress("dataAccess2_________").String()      // cosmos1v3shgc2pvd3k2umnxf047h6lta047h6l006r6w
	dataAccess3Diff1 := sdk.AccAddress("dataAccess3Diff1____").String() // cosmos1v3shgc2pvd3k2umnxdzxjenxx9047h6lfpj2sv
	dataAccess3Diff2 := sdk.AccAddress("dataAccess3Diff2____").String() // cosmos1v3shgc2pvd3k2umnxdzxjenxxf047h6lngt850
	dataAccess3Diff3 := sdk.AccAddress("dataAccess3Diff3____").String() // cosmos1v3shgc2pvd3k2umnxdzxjenxxd047h6lz0mm0w
	dataAccess3Same1 := sdk.AccAddress("dataAccess3Same1____").String() // cosmos1v3shgc2pvd3k2umnxdfkzmt9x9047h6l84c2fh
	dataAccess3Same2 := sdk.AccAddress("dataAccess3Same2____").String() // cosmos1v3shgc2pvd3k2umnxdfkzmt9xf047h6laup8d5
	dataAccess3Same3 := sdk.AccAddress("dataAccess3Same3____").String() // cosmos1v3shgc2pvd3k2umnxdfkzmt9xd047h6lvm3mk4

	valueOwner1 := sdk.AccAddress("valueOwner1_________").String()      // cosmos1weskcat9famkuetjx9047h6lta047h6l7yqwad
	valueOwner2 := sdk.AccAddress("valueOwner2_________").String()      // cosmos1weskcat9famkuetjxf047h6lta047h6la2y63d
	valueOwner3Diff1 := sdk.AccAddress("valueOwner3Diff1____").String() // cosmos1weskcat9famkuetjxdzxjenxx9047h6lmyvnm0
	valueOwner3Diff2 := sdk.AccAddress("valueOwner3Diff2____").String() // cosmos1weskcat9famkuetjxdzxjenxxf047h6lpd47lv
	valueOwner3Diff3 := sdk.AccAddress("valueOwner3Diff3____").String() // cosmos1weskcat9famkuetjxdzxjenxxd047h6ls29zyd
	valueOwner3Same := sdk.AccAddress("valueOwner3Same_____").String()  // cosmos1weskcat9famkuetjxdfkzmt9ta047h6ly5atem
	s.MakeNonWasmAccounts(valueOwner1, valueOwner2, valueOwner3Diff1)

	scopeSpecID := types.ScopeSpecMetadataAddress(s.newUUID("scopespec", 1)) // scopespec1qsc47umrdacx2umsv4347h6lta0s56jv59
	ns := func(scopeID types.MetadataAddress, owner, dataAccess, valueOwner string) types.Scope {
		return types.Scope{
			ScopeId:           scopeID,
			SpecificationId:   scopeSpecID,
			Owners:            []types.Party{{Address: owner, Role: types.PartyType_PARTY_TYPE_OWNER}},
			DataAccess:        []string{dataAccess},
			ValueOwnerAddress: valueOwner,
		}
	}
	ids := func(scopeIDs ...types.MetadataAddress) []types.MetadataAddress {
		return scopeIDs
	}

	newValueOwner := sdk.AccAddress("newValueOwner_______").String() // cosmos1dejhw4npd36k2nmhdejhyh6lta047h6lvuu8z5

	tests := []struct {
		name     string
		starters []types.Scope
		scopeIDs []types.MetadataAddress
		signers  []string
		expErr   string
	}{
		{
			name: "scope 1 of 3 not found",
			starters: []types.Scope{
				ns(scopeID1, owner1, dataAccess1, valueOwner1),
				ns(scopeID2, owner2, dataAccess2, valueOwner2),
			},
			scopeIDs: ids(scopeIDNotFound, scopeID1, scopeID2),
			signers:  []string{valueOwner1, valueOwner2},
			expErr:   "no account address associated with metadata address \"" + scopeIDNotFound.String() + "\": invalid request",
		},
		{
			name: "scope 2 of 3 not found",
			starters: []types.Scope{
				ns(scopeID1, owner1, dataAccess1, valueOwner1),
				ns(scopeID2, owner2, dataAccess2, valueOwner2),
			},
			scopeIDs: ids(scopeID1, scopeIDNotFound, scopeID2),
			signers:  []string{valueOwner1, valueOwner2},
			expErr:   "no account address associated with metadata address \"" + scopeIDNotFound.String() + "\": invalid request",
		},
		{
			name: "scope 3 of 3 not found",
			starters: []types.Scope{
				ns(scopeID1, owner1, dataAccess1, valueOwner1),
				ns(scopeID2, owner2, dataAccess2, valueOwner2),
			},
			scopeIDs: ids(scopeID1, scopeID2, scopeIDNotFound),
			signers:  []string{valueOwner1, valueOwner2},
			expErr:   "no account address associated with metadata address \"" + scopeIDNotFound.String() + "\": invalid request",
		},
		{
			name: "not properly signed",
			starters: []types.Scope{
				ns(scopeID1, owner1, dataAccess1, valueOwner1),
				ns(scopeID2, owner2, dataAccess2, valueOwner2),
			},
			scopeIDs: ids(scopeID1, scopeID2),
			signers:  []string{valueOwner1},
			expErr:   "missing signature from existing value owner \"" + valueOwner2 + "\": invalid request",
		},
		{
			name: "1 scope without value owner",
			starters: []types.Scope{
				ns(scopeID1, owner1, dataAccess1, ""),
			},
			scopeIDs: ids(scopeID1),
			signers:  []string{owner1},
			expErr:   "no account address associated with metadata address \"" + scopeID1.String() + "\": invalid request",
		},
		{
			name: "1 scope updated",
			starters: []types.Scope{
				ns(scopeID1, owner1, dataAccess1, valueOwner1),
			},
			scopeIDs: ids(scopeID1),
			signers:  []string{valueOwner1},
			expErr:   "",
		},
		{
			name: "3 scopes updated all different",
			starters: []types.Scope{
				ns(scopeID3Diff1, owner3Diff1, dataAccess3Diff1, valueOwner3Diff1),
				ns(scopeID3Diff2, owner3Diff2, dataAccess3Diff2, valueOwner3Diff2),
				ns(scopeID3Diff3, owner3Diff3, dataAccess3Diff3, valueOwner3Diff3),
			},
			scopeIDs: ids(scopeID3Diff1, scopeID3Diff2, scopeID3Diff3),
			signers:  []string{valueOwner3Diff1, valueOwner3Diff2, valueOwner3Diff3},
			expErr:   "",
		},
		{
			name: "3 scopes updated all same",
			starters: []types.Scope{
				ns(scopeID3Same1, owner3Same1, dataAccess3Same1, valueOwner3Same),
				ns(scopeID3Same2, owner3Same2, dataAccess3Same2, valueOwner3Same),
				ns(scopeID3Same3, owner3Same3, dataAccess3Same3, valueOwner3Same),
			},
			scopeIDs: ids(scopeID3Same1, scopeID3Same2, scopeID3Same3),
			signers:  []string{valueOwner3Same},
			expErr:   "",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			for _, scope := range tc.starters {
				defer WriteTempScope(s.T(), s.app.MetadataKeeper, s.ctx, scope)()
			}
			msg := types.MsgUpdateValueOwnersRequest{
				ScopeIds:          tc.scopeIDs,
				ValueOwnerAddress: newValueOwner,
				Signers:           tc.signers,
			}
			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)

			_, err := s.msgServer.UpdateValueOwners(ctx, &msg)
			if len(tc.expErr) > 0 {
				s.Assert().EqualError(err, tc.expErr, "error from UpdateValueOwners")
			} else {
				s.Require().NoError(err, "error from UpdateValueOwners")

				for i, scopeID := range tc.scopeIDs {
					actVO, err2 := s.app.MetadataKeeper.GetScopeValueOwner(s.ctx, scopeID)
					if s.Assert().NoError(err2, "[%d]: error from GetScopeValueOwner(%q)", i, scopeID) {
						s.Assert().Equal(msg.ValueOwnerAddress, actVO.String(), "[%d]: addr from GetScopeValueOwner(%q)", i, scopeID)
					}
				}
			}
		})
	}
}

func (s *MsgServerTestSuite) TestMigrateValueOwner() {
	scopeSpecID := types.ScopeSpecMetadataAddress(uuid.New())
	storeScope := func(valueOwner string, scopeID types.MetadataAddress) {
		scope := types.Scope{
			ScopeId:           scopeID,
			SpecificationId:   scopeSpecID,
			Owners:            []types.Party{{Address: s.user1, Role: types.PartyType_PARTY_TYPE_OWNER}},
			ValueOwnerAddress: valueOwner,
		}
		s.app.MetadataKeeper.SetScope(s.ctx, scope)
	}
	addr := func(str string) string {
		return sdk.AccAddress(str).String()
	}

	addrW1 := addr("addrW1______________")
	addrW3 := addr("addrW3______________")

	scopeID1 := types.ScopeMetadataAddress(uuid.New())
	scopeID31 := types.ScopeMetadataAddress(uuid.New())
	scopeID32 := types.ScopeMetadataAddress(uuid.New())
	scopeID33 := types.ScopeMetadataAddress(uuid.New())

	storeScope(addrW1, scopeID1)
	storeScope(addrW3, scopeID31)
	storeScope(addrW3, scopeID32)
	storeScope(addrW3, scopeID33)

	tests := []struct {
		name     string
		msg      *types.MsgMigrateValueOwnerRequest
		expErr   string
		scopeIDs []types.MetadataAddress
	}{
		{
			name: "err from IterateScopesForValueOwner",
			msg: &types.MsgMigrateValueOwnerRequest{
				Existing: "",
				Proposed: "doesn't matter",
				Signers:  []string{"who cares"},
			},
			expErr: "invalid existing address \"\": empty address string is not allowed: invalid request",
		},
		{
			name: "no scopes",
			msg: &types.MsgMigrateValueOwnerRequest{
				Existing: addr("unknown_value_owner_"),
				Proposed: addr("does_not_matter_____"),
				Signers:  []string{addr("signer______________")},
			},
			expErr: "no scopes found with value owner \"" + addr("unknown_value_owner_") + "\": not found",
		},
		{
			name: "err from ValidateUpdateValueOwners",
			msg: &types.MsgMigrateValueOwnerRequest{
				Existing: addrW1,
				Proposed: addr("not_for_public_use__"),
				Signers:  []string{addr("incorrect_signer____")},
			},
			expErr: "missing signature from existing value owner \"" + addrW1 + "\": invalid request",
		},
		{
			name: "1 scope updated",
			msg: &types.MsgMigrateValueOwnerRequest{
				Existing: addrW1,
				Proposed: addr("proposed_value_owner"),
				Signers:  []string{addrW1},
			},
			scopeIDs: []types.MetadataAddress{scopeID1},
		},
		{
			name: "3 scopes updated",
			msg: &types.MsgMigrateValueOwnerRequest{
				Existing: addrW3,
				Proposed: addr("a_longer_proposed_value_owner___"),
				Signers:  []string{addrW3},
			},
			scopeIDs: []types.MetadataAddress{scopeID31, scopeID32, scopeID33},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			_, err := s.msgServer.MigrateValueOwner(s.ctx, tc.msg)
			if len(tc.expErr) > 0 {
				s.Assert().EqualError(err, tc.expErr, "Metadata hander(%T)", tc.msg)
			} else {
				s.Require().NoError(err, tc.expErr, "Metadata hander(%T)", tc.msg)
				for i, scopeID := range tc.scopeIDs {
					actVO, err2 := s.app.MetadataKeeper.GetScopeValueOwner(s.ctx, scopeID)
					if s.Assert().NoError(err2, "[%d]: error from GetScopeValueOwner(%q)", i, scopeID) {
						s.Assert().Equal(tc.msg.Proposed, actVO.String(), "[%d]: addr from GetScopeValueOwner(%q)", i, scopeID)
					}
				}
			}
		})
	}
}

func (s *MsgServerTestSuite) TestWriteSession() {
	cSpec := types.ContractSpecification{
		SpecificationId: types.ContractSpecMetadataAddress(uuid.New()),
		Description:     nil,
		OwnerAddresses:  []string{s.user1},
		PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		Source:          types.NewContractSpecificationSourceHash("somesource"),
		ClassName:       "someclass",
	}
	s.app.MetadataKeeper.SetContractSpecification(s.ctx, cSpec)
	sSpec := types.ScopeSpecification{
		SpecificationId: types.ScopeSpecMetadataAddress(uuid.New()),
		Description:     nil,
		OwnerAddresses:  []string{s.user1},
		PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		ContractSpecIds: []types.MetadataAddress{cSpec.SpecificationId},
	}
	s.app.MetadataKeeper.SetScopeSpecification(s.ctx, sSpec)

	scopeUUID := uuid.New()
	scope := types.Scope{
		ScopeId:         types.ScopeMetadataAddress(scopeUUID),
		SpecificationId: sSpec.SpecificationId,
		Owners: []types.Party{{
			Address: s.user1,
			Role:    types.PartyType_PARTY_TYPE_OWNER,
		}},
		DataAccess:        nil,
		ValueOwnerAddress: "",
	}
	s.app.MetadataKeeper.SetScope(s.ctx, scope)

	someBytes, err := base64.StdEncoding.DecodeString("ChFIRUxMTyBQUk9WRU5BTkNFIQ==")
	require.NoError(s.T(), err, "trying to create someBytes")

	cases := []struct {
		name     string
		session  types.Session
		signers  []string
		errorMsg string
	}{
		{
			"valid without context",
			types.Session{
				SessionId:       types.SessionMetadataAddress(scopeUUID, uuid.New()),
				SpecificationId: cSpec.SpecificationId,
				Parties:         scope.Owners,
				Name:            "someclass",
				Context:         nil,
				Audit:           nil,
			},
			[]string{s.user1},
			"",
		},
		{
			"valid with context",
			types.Session{
				SessionId:       types.SessionMetadataAddress(scopeUUID, uuid.New()),
				SpecificationId: cSpec.SpecificationId,
				Parties:         scope.Owners,
				Name:            "someclass",
				Context:         someBytes,
				Audit:           nil,
			},
			[]string{s.user1},
			"",
		},
	}

	for _, tc := range cases {
		s.T().Run(tc.name, func(t *testing.T) {
			msg := types.MsgWriteSessionRequest{
				Session:             tc.session,
				Signers:             tc.signers,
				SessionIdComponents: nil,
				SpecUuid:            "",
			}
			_, err := s.msgServer.WriteSession(s.ctx, &msg)
			if len(tc.errorMsg) > 0 {
				assert.EqualError(t, err, tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s *MsgServerTestSuite) TestWriteDeleteRecord() {
	cSpecUUID := uuid.New()
	cSpec := types.ContractSpecification{
		SpecificationId: types.ContractSpecMetadataAddress(cSpecUUID),
		Description:     nil,
		OwnerAddresses:  []string{s.user1},
		PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		Source:          types.NewContractSpecificationSourceHash("somesource1"),
		ClassName:       "someclass1",
	}
	s.app.MetadataKeeper.SetContractSpecification(s.ctx, cSpec)
	defer func() {
		s.Assert().NoError(s.app.MetadataKeeper.RemoveContractSpecification(s.ctx, cSpec.SpecificationId), "removing contract spec")
	}()

	sSpecUUID := uuid.New()
	sSpec := types.ScopeSpecification{
		SpecificationId: types.ScopeSpecMetadataAddress(sSpecUUID),
		Description:     nil,
		OwnerAddresses:  []string{s.user1},
		PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		ContractSpecIds: []types.MetadataAddress{cSpec.SpecificationId},
	}
	s.app.MetadataKeeper.SetScopeSpecification(s.ctx, sSpec)
	defer func() {
		s.Assert().NoError(s.app.MetadataKeeper.RemoveScopeSpecification(s.ctx, sSpec.SpecificationId), "removing scope spec")
	}()

	rSpec := types.RecordSpecification{
		SpecificationId: types.RecordSpecMetadataAddress(cSpecUUID, "record"),
		Name:            "record",
		Inputs: []*types.InputSpecification{
			{
				Name:     "ri1",
				TypeName: "string",
				Source:   types.NewInputSpecificationSourceHash("ri1hash"),
			},
		},
		TypeName:           "string",
		ResultType:         types.DefinitionType_DEFINITION_TYPE_RECORD_LIST,
		ResponsibleParties: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
	}
	s.app.MetadataKeeper.SetRecordSpecification(s.ctx, rSpec)
	defer func() {
		s.Assert().NoError(s.app.MetadataKeeper.RemoveRecordSpecification(s.ctx, rSpec.SpecificationId), "removing record spec 1")
	}()

	scopeUUID := uuid.New()
	scope := types.Scope{
		ScopeId:         types.ScopeMetadataAddress(scopeUUID),
		SpecificationId: sSpec.SpecificationId,
		Owners: []types.Party{{
			Address: s.user1,
			Role:    types.PartyType_PARTY_TYPE_OWNER,
		}},
		DataAccess:        nil,
		ValueOwnerAddress: "",
	}
	defer WriteTempScope(s.T(), s.app.MetadataKeeper, s.ctx, scope)()

	session1UUID := uuid.New()
	session1 := types.Session{
		SessionId:       types.SessionMetadataAddress(scopeUUID, session1UUID),
		SpecificationId: cSpec.SpecificationId,
		Parties:         ownerPartyList(s.user1),
		Name:            "someclass1",
	}
	s.app.MetadataKeeper.SetSession(s.ctx, session1)
	defer s.app.MetadataKeeper.RemoveSession(s.ctx, session1.SessionId)

	session2UUID := uuid.New()
	session2 := types.Session{
		SessionId:       types.SessionMetadataAddress(scopeUUID, session2UUID),
		SpecificationId: cSpec.SpecificationId,
		Parties:         ownerPartyList(s.user1),
		Name:            "someclass1",
	}
	s.app.MetadataKeeper.SetSession(s.ctx, session2)
	defer s.app.MetadataKeeper.RemoveSession(s.ctx, session2.SessionId)

	record := types.Record{
		Name:      rSpec.Name,
		SessionId: session1.SessionId,
		Process: types.Process{
			ProcessId: &types.Process_Hash{Hash: "rprochash"},
			Name:      "rproc",
			Method:    "rprocmethod",
		},
		Inputs: []types.RecordInput{
			{
				Name:     rSpec.Inputs[0].Name,
				Source:   &types.RecordInput_Hash{Hash: "rhash"},
				TypeName: rSpec.Inputs[0].TypeName,
				Status:   types.RecordInputStatus_Proposed,
			},
		},
		Outputs: []types.RecordOutput{
			{
				Hash:   "rout1",
				Status: types.ResultStatus_RESULT_STATUS_PASS,
			},
			{
				Hash:   "rout2",
				Status: types.ResultStatus_RESULT_STATUS_PASS,
			},
		},
		SpecificationId: rSpec.SpecificationId,
	}
	recordID := types.RecordMetadataAddress(scopeUUID, rSpec.Name)
	// Not adding the record here because we're testing that stuff.

	s.T().Run("write invalid record", func(t *testing.T) {
		// Make a record with an unknown spec id. Try to write it and expect an error.
		badRecord := types.Record{
			Name:      rSpec.Name,
			SessionId: session1.SessionId,
			Process: types.Process{
				ProcessId: &types.Process_Hash{Hash: "badrprochash"},
				Name:      "badrproc",
				Method:    "badrprocmethod",
			},
			Inputs: []types.RecordInput{
				{
					Name:     rSpec.Inputs[0].Name,
					Source:   &types.RecordInput_Hash{Hash: "badrhash"},
					TypeName: rSpec.Inputs[0].TypeName,
					Status:   types.RecordInputStatus_Proposed,
				},
			},
			Outputs: []types.RecordOutput{
				{
					Hash:   "badrout1",
					Status: types.ResultStatus_RESULT_STATUS_PASS,
				},
				{
					Hash:   "badrout2",
					Status: types.ResultStatus_RESULT_STATUS_PASS,
				},
			},
			SpecificationId: types.RecordSpecMetadataAddress(uuid.New(), rSpec.Name),
		}
		msg := types.MsgWriteRecordRequest{
			Record:              badRecord,
			Signers:             []string{s.user1},
			SessionIdComponents: nil,
			ContractSpecUuid:    "",
			Parties:             ownerPartyList(s.user1),
		}
		_, err := s.msgServer.WriteRecord(s.ctx, &msg)
		require.Error(t, err, "sending bad MsgWriteRecordRequest")
		require.Contains(t, err.Error(), "proposed specification id")
		require.Contains(t, err.Error(), "does not match expected")
	})

	s.T().Run("write record to session 1", func(t *testing.T) {
		msg := types.MsgWriteRecordRequest{
			Record:              record,
			Signers:             []string{s.user1},
			SessionIdComponents: nil,
			ContractSpecUuid:    "",
			Parties:             ownerPartyList(s.user1),
		}
		_, err := s.msgServer.WriteRecord(s.ctx, &msg)
		require.NoError(t, err, "sending MsgWriteRecordRequest")
		r, rok := s.app.MetadataKeeper.GetRecord(s.ctx, recordID)
		if assert.True(t, rok, "GetRecord bool") {
			assert.Equal(t, record, r, "GetRecord record")
		}
	})

	s.T().Run("Update record to other session", func(t *testing.T) {
		record.SessionId = session2.SessionId
		msg := types.MsgWriteRecordRequest{
			Record:              record,
			Signers:             []string{s.user1},
			SessionIdComponents: nil,
			ContractSpecUuid:    "",
			Parties:             ownerPartyList(s.user1),
		}
		_, err := s.msgServer.WriteRecord(s.ctx, &msg)
		require.NoError(t, err, "sending MsgWriteRecordRequest")
		r, rok := s.app.MetadataKeeper.GetRecord(s.ctx, recordID)
		if assert.True(t, rok, "GetRecord bool") {
			assert.Equal(t, record, r, "GetRecord record")
		}
		// Make sure the session was deleted since it's now empty.
		_, sok := s.app.MetadataKeeper.GetSession(s.ctx, session1.SessionId)
		assert.False(t, sok, "GetSession session 1 bool")
	})

	s.T().Run("delete the record", func(t *testing.T) {
		msg := types.MsgDeleteRecordRequest{
			RecordId: recordID,
			Signers:  []string{s.user1},
		}
		_, err := s.msgServer.DeleteRecord(s.ctx, &msg)
		require.NoError(t, err, "sending MsgDeleteRecordRequest")
		_, rok := s.app.MetadataKeeper.GetRecord(s.ctx, recordID)
		assert.False(t, rok, "GetRecord bool")
		// Make sure the session was deleted since it's now empty.
		_, sok := s.app.MetadataKeeper.GetSession(s.ctx, session2.SessionId)
		assert.False(t, sok, "GetSession session 2 bool")
	})
}

// TODO: WriteScopeSpecification tests
// TODO: DeleteScopeSpecification tests
// TODO: WriteContractSpecification tests
// TODO: DeleteContractSpecification tests

func (s *MsgServerTestSuite) TestAddContractSpecToScopeSpec() {
	cSpec := types.ContractSpecification{
		SpecificationId: types.ContractSpecMetadataAddress(uuid.New()),
		Description:     nil,
		OwnerAddresses:  []string{s.user1},
		PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		Source:          types.NewContractSpecificationSourceHash("somesource"),
		ClassName:       "someclass",
	}
	s.app.MetadataKeeper.SetContractSpecification(s.ctx, cSpec)
	sSpec := types.ScopeSpecification{
		SpecificationId: types.ScopeSpecMetadataAddress(uuid.New()),
		Description:     nil,
		OwnerAddresses:  []string{s.user1},
		PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		ContractSpecIds: []types.MetadataAddress{cSpec.SpecificationId},
	}
	s.app.MetadataKeeper.SetScopeSpecification(s.ctx, sSpec)

	cSpec2 := types.ContractSpecification{
		SpecificationId: types.ContractSpecMetadataAddress(uuid.New()),
		Description:     nil,
		OwnerAddresses:  []string{s.user1},
		PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		Source:          types.NewContractSpecificationSourceHash("somesource"),
		ClassName:       "someclass",
	}
	s.app.MetadataKeeper.SetContractSpecification(s.ctx, cSpec2)

	unknownContractSpecId := types.ContractSpecMetadataAddress(uuid.New())
	unknownScopeSpecId := types.ScopeSpecMetadataAddress(uuid.New())

	cases := []struct {
		name           string
		contractSpecId types.MetadataAddress
		scopeSpecId    types.MetadataAddress
		signers        []string
		errorMsg       string
	}{
		{
			"fail to add contract spec, cannot find contract spec",
			unknownContractSpecId,
			sSpec.SpecificationId,
			[]string{s.user1},
			fmt.Sprintf("contract specification not found with id %s: not found", unknownContractSpecId),
		},
		{
			"fail to add contract spec, cannot find scope spec",
			cSpec2.SpecificationId,
			unknownScopeSpecId,
			[]string{s.user1},
			fmt.Sprintf("scope specification not found with id %s: not found", unknownScopeSpecId),
		},
		{
			"fail to add contract spec, scope spec already has contract spec",
			cSpec.SpecificationId,
			sSpec.SpecificationId,
			[]string{s.user1},
			fmt.Sprintf("scope spec %s already contains contract spec %s: invalid request", sSpec.SpecificationId, cSpec.SpecificationId),
		},
		{
			"should successfully add contract spec to scope spec",
			cSpec2.SpecificationId,
			sSpec.SpecificationId,
			[]string{s.user1},
			"",
		},
	}

	for _, tc := range cases {
		s.T().Run(tc.name, func(t *testing.T) {
			msg := types.MsgAddContractSpecToScopeSpecRequest{
				ContractSpecificationId: tc.contractSpecId,
				ScopeSpecificationId:    tc.scopeSpecId,
				Signers:                 tc.signers,
			}
			_, err := s.msgServer.AddContractSpecToScopeSpec(s.ctx, &msg)
			if len(tc.errorMsg) > 0 {
				assert.EqualError(t, err, tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s *MsgServerTestSuite) TestDeleteContractSpecFromScopeSpec() {
	cSpec := types.ContractSpecification{
		SpecificationId: types.ContractSpecMetadataAddress(uuid.New()),
		Description:     nil,
		OwnerAddresses:  []string{s.user1},
		PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		Source:          types.NewContractSpecificationSourceHash("somesource"),
		ClassName:       "someclass",
	}
	s.app.MetadataKeeper.SetContractSpecification(s.ctx, cSpec)
	cSpec2 := types.ContractSpecification{
		SpecificationId: types.ContractSpecMetadataAddress(uuid.New()),
		Description:     nil,
		OwnerAddresses:  []string{s.user1},
		PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		Source:          types.NewContractSpecificationSourceHash("somesource"),
		ClassName:       "someclass",
	}
	s.app.MetadataKeeper.SetContractSpecification(s.ctx, cSpec2)
	cSpecDNE := types.ContractSpecMetadataAddress(uuid.New()) // Does Not Exist.
	sSpec := types.ScopeSpecification{
		SpecificationId: types.ScopeSpecMetadataAddress(uuid.New()),
		Description:     nil,
		OwnerAddresses:  []string{s.user1},
		PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		ContractSpecIds: []types.MetadataAddress{cSpec.SpecificationId, cSpec2.SpecificationId, cSpecDNE},
	}
	s.app.MetadataKeeper.SetScopeSpecification(s.ctx, sSpec)

	unknownScopeSpecId := types.ScopeSpecMetadataAddress(uuid.New())

	cases := []struct {
		name           string
		contractSpecId types.MetadataAddress
		scopeSpecId    types.MetadataAddress
		signers        []string
		errorMsg       string
	}{
		{
			"cannot find contract spec",
			cSpecDNE,
			sSpec.SpecificationId,
			[]string{s.user1},
			"",
		},
		{
			"fail to delete contract spec from scope spec, cannot find scope spec",
			cSpec2.SpecificationId,
			unknownScopeSpecId,
			[]string{s.user1},
			fmt.Sprintf("scope specification not found with id %s: not found", unknownScopeSpecId),
		},
		{
			"should succeed to add contract spec to scope spec",
			cSpec2.SpecificationId,
			sSpec.SpecificationId,
			[]string{s.user1},
			"",
		},
		{
			"fail to delete contract spec from scope spec, scope spec does not contain contract spec",
			cSpec2.SpecificationId,
			sSpec.SpecificationId,
			[]string{s.user1},
			fmt.Sprintf("contract specification %s not found in scope specification %s: not found", cSpec2.SpecificationId, sSpec.SpecificationId),
		},
	}

	for _, tc := range cases {
		s.T().Run(tc.name, func(t *testing.T) {
			msg := types.MsgDeleteContractSpecFromScopeSpecRequest{
				ContractSpecificationId: tc.contractSpecId,
				ScopeSpecificationId:    tc.scopeSpecId,
				Signers:                 tc.signers,
			}
			_, err := s.msgServer.DeleteContractSpecFromScopeSpec(s.ctx, &msg)
			if len(tc.errorMsg) > 0 {
				assert.EqualError(t, err, tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TODO: WriteRecordSpecification tests
// TODO: DeleteRecordSpecification tests

// TODO: BindOSLocator tests
// TODO: DeleteOSLocator tests
// TODO: ModifyOSLocator tests

func (s *MsgServerTestSuite) TestSetAccountData() {
	scopeSpec := types.ScopeSpecification{
		SpecificationId: types.ScopeSpecMetadataAddress(uuid.New()),
		OwnerAddresses:  []string{s.user1},
		PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
	}
	s.app.MetadataKeeper.SetScopeSpecification(s.ctx, scopeSpec)

	scope := types.Scope{
		ScopeId:         types.ScopeMetadataAddress(uuid.New()),
		SpecificationId: scopeSpec.SpecificationId,
		Owners:          []types.Party{{Address: s.user1, Role: types.PartyType_PARTY_TYPE_OWNER}},
	}
	s.app.MetadataKeeper.SetScope(s.ctx, scope)

	tests := []struct {
		name   string
		msg    *types.MsgSetAccountDataRequest
		exp    *types.MsgSetAccountDataResponse
		expErr string
	}{
		{
			name: "incorrect signer",
			msg: &types.MsgSetAccountDataRequest{
				MetadataAddr: scope.ScopeId,
				Value:        "This won't work.",
				Signers:      []string{s.user2},
			},
			expErr: "missing signature: " + s.user1 + ": invalid request",
		},
		{
			name: "value too long",
			msg: &types.MsgSetAccountDataRequest{
				MetadataAddr: scope.ScopeId,
				Value:        strings.Repeat("This won't work. ", 1000),
				Signers:      []string{s.user1},
			},
			expErr: "could not set accountdata for \"" + scope.ScopeId.String() + "\": attribute value length of 17000 exceeds max length 10000: invalid request",
		},
		{
			name: "all good",
			msg: &types.MsgSetAccountDataRequest{
				MetadataAddr: scope.ScopeId,
				Value:        "This value is a good value for this scope.",
				Signers:      []string{s.user1},
			},
			exp: &types.MsgSetAccountDataResponse{},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var err error
			var result *types.MsgSetAccountDataResponse
			testFunc := func() {
				result, err = s.msgServer.SetAccountData(s.ctx, tc.msg)
			}
			s.Require().NotPanics(testFunc, "%T hander", tc.msg)
			s.AssertErrorValue(err, tc.expErr, "%T handler error", tc.msg)
			if tc.exp == nil {
				s.Assert().Nil(result, "%T handler result", tc.msg)
			} else {
				s.Assert().Equal(tc.exp, result, "%T handler msg response", tc.msg)
			}
		})
	}
}

func (s *MsgServerTestSuite) TestIssue412WriteScopeOptionalField() {
	ownerAddress := "cosmos1vz99nyd2er8myeugsr4xm5duwhulhp5ae4dvpa"
	specIDStr := "scopespec1qjkyp28sldx5r9ueaxqc5adrc5wszy6nsh"
	specUUIDStr := "ac40a8f0-fb4d-4197-99e9-818a75a3c51d"
	specID, specIDErr := types.MetadataAddressFromBech32(specIDStr)
	require.NoError(s.T(), specIDErr, "converting scopeIDStr to a metadata address")

	s.T().Run("Ensure ID and UUID strings match", func(t *testing.T) {
		specIDFromID, e1 := types.MetadataAddressFromBech32(specIDStr)
		require.NoError(t, e1, "specIDFromIDStr")
		specUUIDFromID, e2 := specIDFromID.ScopeSpecUUID()
		require.NoError(t, e2, "specUUIDActualStr")
		specUUIDStrActual := specUUIDFromID.String()
		assert.Equal(t, specUUIDStr, specUUIDStrActual, "UUID strings")

		specIDFFromUUID := types.ScopeSpecMetadataAddress(uuid.MustParse(specUUIDStr))
		specIDStrActual := specIDFFromUUID.String()
		assert.Equal(t, specIDStr, specIDStrActual, "ID strings")

		assert.Equal(t, specIDFromID, specIDFFromUUID, "scope spec ids")
	})

	s.T().Run("Setting scope spec with just a spec ID", func(t *testing.T) {
		msg := types.MsgWriteScopeSpecificationRequest{
			Specification: types.ScopeSpecification{
				SpecificationId: specID,
				Description: &types.Description{
					Name:        "io.prov.contracts.examplekotlin.helloWorld",
					Description: "A generic scope that allows for a lot of example hello world contracts.",
				},
				OwnerAddresses:  []string{ownerAddress},
				PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				ContractSpecIds: nil,
			},
			Signers:  []string{ownerAddress},
			SpecUuid: "",
		}
		res, err := s.msgServer.WriteScopeSpecification(s.ctx, &msg)
		assert.NoError(t, err)
		assert.NotNil(t, 0, res)
	})

	s.T().Run("Setting scope spec with just a UUID", func(t *testing.T) {
		msg := types.MsgWriteScopeSpecificationRequest{
			Specification: types.ScopeSpecification{
				SpecificationId: nil,
				Description: &types.Description{
					Name:        "io.prov.contracts.examplekotlin.helloWorld",
					Description: "A generic scope that allows for a lot of example hello world contracts.",
				},
				OwnerAddresses:  []string{ownerAddress},
				PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				ContractSpecIds: nil,
			},
			Signers:  []string{ownerAddress},
			SpecUuid: specUUIDStr,
		}
		res, err := s.msgServer.WriteScopeSpecification(s.ctx, &msg)
		assert.NoError(t, err)
		assert.NotNil(t, 0, res)
	})

	s.T().Run("Setting scope spec with matching ID and UUID", func(t *testing.T) {
		msg := types.MsgWriteScopeSpecificationRequest{
			Specification: types.ScopeSpecification{
				SpecificationId: specID,
				Description: &types.Description{
					Name:        "io.prov.contracts.examplekotlin.helloWorld",
					Description: "A generic scope that allows for a lot of example hello world contracts.",
				},
				OwnerAddresses:  []string{ownerAddress},
				PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				ContractSpecIds: nil,
			},
			Signers:  []string{ownerAddress},
			SpecUuid: specUUIDStr,
		}
		res, err := s.msgServer.WriteScopeSpecification(s.ctx, &msg)
		assert.NoError(t, err)
		assert.NotNil(t, 0, res)
	})
}

func (s *MsgServerTestSuite) TestAddNetAssetValue() {
	scopeSpecUUIDNF := uuid.New()
	scopeSpecIDNF := types.ScopeSpecMetadataAddress(scopeSpecUUIDNF)

	scopeUUID := uuid.New()
	scopeID := types.ScopeMetadataAddress(scopeUUID)
	scopeSpecUUID := uuid.New()
	scopeSpecID := types.ScopeSpecMetadataAddress(scopeSpecUUID)
	pubkey1 := secp256k1.GenPrivKey().PubKey()
	user1Addr := sdk.AccAddress(pubkey1.Address())
	user1 := user1Addr.String()
	pubkey2 := secp256k1.GenPrivKey().PubKey()
	user2Addr := sdk.AccAddress(pubkey2.Address())
	user2 := user2Addr.String()

	ns := *types.NewScope(scopeID, scopeSpecID, ownerPartyList(user1), []string{user1}, user1, false)

	s.app.MetadataKeeper.SetScope(s.ctx, ns)

	testCases := []struct {
		name   string
		msg    types.MsgAddNetAssetValuesRequest
		expErr string
	}{
		{
			name: "no marker found",
			msg: types.MsgAddNetAssetValuesRequest{
				ScopeId: scopeSpecIDNF.String(),
				NetAssetValues: []types.NetAssetValue{
					{
						Price: sdk.NewInt64Coin("navcoin", 1),
					}},
				Signers: []string{user1},
			},
			expErr: fmt.Sprintf("scope not found: %v: not found", scopeSpecIDNF.String()),
		},
		{
			name: "value denom does not exist",
			msg: types.MsgAddNetAssetValuesRequest{
				ScopeId: scopeID.String(),
				NetAssetValues: []types.NetAssetValue{
					{
						Price:              sdk.NewInt64Coin("hotdog", 100),
						UpdatedBlockHeight: 1,
					},
				},
				Signers: []string{user1},
			},
			expErr: `net asset value denom does not exist: marker hotdog not found for address: cosmos1p6l3annxy35gm5mfm6m0jz2mdj8peheuzf9alh: invalid request`,
		},
		{
			name: "not authorize user",
			msg: types.MsgAddNetAssetValuesRequest{
				ScopeId: scopeID.String(),
				NetAssetValues: []types.NetAssetValue{
					{
						Price:              sdk.NewInt64Coin(types.UsdDenom, 100),
						UpdatedBlockHeight: 1,
					},
				},
				Signers: []string{user2},
			},
			expErr: fmt.Sprintf("missing signature: %v: unauthorized", user1),
		},
		{
			name: "successfully set nav",
			msg: types.MsgAddNetAssetValuesRequest{
				ScopeId: scopeID.String(),
				NetAssetValues: []types.NetAssetValue{
					{
						Price:              sdk.NewInt64Coin(types.UsdDenom, 100),
						UpdatedBlockHeight: 1,
					},
				},
				Signers: []string{user1},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			res, err := s.msgServer.AddNetAssetValues(sdk.WrapSDKContext(s.ctx),
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
