package keeper_test

import (
	"errors"
	"fmt"
	"math/rand"
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
	ctx := s.FreshCtx()

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
}

func (s *AuthzTestSuite) FreshCtx() sdk.Context {
	return keeper.AddAuthzCacheToContext(s.app.BaseApp.NewContext(false, tmproto.Header{Time: time.Now()}))
}

// AssertErrorValue asserts that:
//   - If errorString is empty, theError must be nil
//   - If errorString is not empty, theError must equal the errorString.
func AssertErrorValue(t *testing.T, theError error, errorString string, msgAndArgs ...interface{}) bool {
	t.Helper()
	if len(errorString) > 0 {
		return assert.EqualError(t, theError, errorString, msgAndArgs...)
	}
	return assert.NoError(t, theError, msgAndArgs...)
}

// AssertErrorValue asserts that:
//   - If errorString is empty, theError must be nil
//   - If errorString is not empty, theError must equal the errorString.
func (s *AuthzTestSuite) AssertErrorValue(theError error, errorString string, msgAndArgs ...interface{}) bool {
	s.T().Helper()
	return AssertErrorValue(s.T(), theError, errorString, msgAndArgs...)
}

func TestAuthzTestSuite(t *testing.T) {
	suite.Run(t, new(AuthzTestSuite))
}

func (s *AuthzTestSuite) TestValidateSignersWithParties() {
	// These tests are pretty light since all it does is call
	// validateAllRequiredPartiesSigned and validateSmartContractSigners.
	// The assumption is that those are well tested.

	accStr := func(str string) string {
		return sdk.AccAddress(str).String()
	}
	pt := func(addr string, role types.PartyType, opt bool) types.Party {
		return types.Party{
			Address:  accStr(addr),
			Role:     role,
			Optional: opt,
		}
	}
	ptz := func(parties ...types.Party) []types.Party {
		rv := make([]types.Party, 0, len(parties))
		rv = append(rv, parties...)
		return rv
	}
	normalMsg := func(signers ...string) types.MetadataMsg {
		rv := &types.MsgWriteScopeRequest{Signers: make([]string, len(signers))}
		for i, signer := range signers {
			rv.Signers[i] = accStr(signer)
		}
		return rv
	}
	normalMsgType := types.TypeURLMsgWriteScopeRequest
	scGetAccCall := func(addr string) *GetAccountCall {
		return &GetAccountCall{
			Addr:   sdk.AccAddress(addr),
			Result: authtypes.NewBaseAccount(sdk.AccAddress(addr), nil, 0, 0),
		}
	}

	owner := types.PartyType_PARTY_TYPE_OWNER
	provenance := types.PartyType_PARTY_TYPE_PROVENANCE

	tests := []struct {
		name             string
		reqParties       []types.Party
		availableParties []types.Party
		reqRoles         []types.PartyType
		msg              types.MetadataMsg
		authK            *MockAuthKeeper
		authzK           *MockAuthzKeeper
		expErr           string
	}{
		{
			name:             "all nil",
			reqParties:       nil,
			availableParties: nil,
			reqRoles:         nil,
			msg:              normalMsg("signer1"),
			expErr:           "",
		},
		{
			name:             "all empty",
			reqParties:       []types.Party{},
			availableParties: []types.Party{},
			reqRoles:         []types.PartyType{},
			msg:              normalMsg("signer1"),
			expErr:           "",
		},
		{
			name:       "missing sig from required party",
			reqParties: ptz(pt("req1", owner, false)),
			msg:        normalMsg("signer1"),
			expErr:     "missing required signature: " + accStr("req1") + " (OWNER)",
		},
		{
			name:             "missing sig from req role",
			availableParties: ptz(pt("party1", owner, true), pt("party2", owner, true)),
			reqRoles:         []types.PartyType{owner},
			msg:              normalMsg("party3"),
			expErr:           "missing signers for roles required by spec: OWNER need 1 have 0",
		},
		{
			name:       "provenance role in req parties is not smart contract",
			reqParties: ptz(pt("prov", provenance, true)),
			msg:        normalMsg("signer1"),
			expErr:     "",
		},
		{
			name:             "provenance role in available parties is not smart contract",
			availableParties: ptz(pt("prov", provenance, true)),
			msg:              normalMsg("signer1"),
			expErr:           `account "` + accStr("prov") + `" has role PROVENANCE but is not a smart contract`,
		},
		{
			name:       "smart contract in req parties is not provenance role",
			reqParties: ptz(pt("sc1", owner, false), pt("user1", owner, false)),
			msg:        normalMsg("sc1", "user1"),
			authK:      NewMockAuthKeeper().WithGetAccountResults(scGetAccCall("sc1")),
			expErr:     "",
		},
		{
			name:             "smart contract in available parties is not provenance role",
			availableParties: ptz(pt("sc1", owner, false), pt("user1", owner, false)),
			msg:              normalMsg("sc1", "user1"),
			authK:            NewMockAuthKeeper().WithGetAccountResults(scGetAccCall("sc1")),
			expErr:           `account "` + accStr("sc1") + `" is a smart contract but does not have the PROVENANCE role`,
		},
		{
			name:             "smart contract signer not a party",
			reqParties:       ptz(pt("req1", owner, false)),
			availableParties: ptz(pt("req1", owner, false)),
			reqRoles:         []types.PartyType{owner},
			msg:              normalMsg("sc1", "req1"),
			authK:            NewMockAuthKeeper().WithGetAccountResults(scGetAccCall("sc1")),
			expErr:           "smart contract signer " + accStr("sc1") + " is not authorized",
		},
		{
			name:             "smart contract signer is a party",
			reqParties:       ptz(pt("req1", owner, false)),
			availableParties: ptz(pt("req1", owner, false), pt("sc1", provenance, true)),
			reqRoles:         []types.PartyType{owner},
			msg:              normalMsg("sc1", "req1"),
			authK:            NewMockAuthKeeper().WithGetAccountResults(scGetAccCall("sc1")),
			expErr:           "",
		},
		{
			name:             "smart contract signer is last and not a party",
			reqParties:       ptz(pt("req1", owner, false)),
			availableParties: ptz(pt("req1", owner, false)),
			reqRoles:         []types.PartyType{owner},
			msg:              normalMsg("req1", "sc1"),
			authK:            NewMockAuthKeeper().WithGetAccountResults(scGetAccCall("sc1")),
			expErr:           "smart contract signer " + accStr("sc1") + " cannot be the last signer",
		},
		{
			name:             "smart contract and req signed sc not a party but is authorized",
			reqParties:       ptz(pt("req1", owner, false)),
			availableParties: ptz(pt("req1", owner, false)),
			reqRoles:         []types.PartyType{owner},
			msg:              normalMsg("sc1", "req1"),
			authK:            NewMockAuthKeeper().WithGetAccountResults(scGetAccCall("sc1")),
			authzK: NewMockAuthzKeeper().WithGetAuthorizationResults(
				GetAuthorizationCall{
					GrantInfo: GrantInfo{Grantee: sdk.AccAddress("sc1"), Granter: sdk.AccAddress("req1"), MsgType: normalMsgType},
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one", authz.AcceptResponse{Accept: true}, nil),
						Exp:  nil,
					},
				},
			),
			expErr: "",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.authK == nil {
				tc.authK = NewMockAuthKeeper()
			}
			if tc.authzK == nil {
				tc.authzK = NewMockAuthzKeeper()
			}
			k := s.app.MetadataKeeper
			origAuthK := k.SetAuthKeeper(tc.authK)
			origAuthzK := k.SetAuthzKeeper(tc.authzK)
			defer func() {
				k.SetAuthKeeper(origAuthK)
				k.SetAuthzKeeper(origAuthzK)
			}()

			err := k.ValidateSignersWithParties(s.FreshCtx(), tc.reqParties, tc.availableParties, tc.reqRoles, tc.msg)
			s.AssertErrorValue(err, tc.expErr, "ValidateSignersWithParties")
		})
	}
}

func (s *AuthzTestSuite) TestValidateAllRequiredPartiesSigned() {
	acc := func(addr string) sdk.AccAddress {
		if len(addr) == 0 {
			return nil
		}
		return sdk.AccAddress(addr)
	}
	accStr := func(addr string) string {
		if len(addr) == 0 {
			return ""
		}
		return acc(addr).String()
	}

	pt := func(addr string, role types.PartyType, optional bool) types.Party {
		return types.Party{
			Address:  accStr(addr),
			Role:     role,
			Optional: optional,
		}
	}
	ptz := func(parties ...types.Party) []types.Party {
		rv := make([]types.Party, 0, len(parties))
		rv = append(rv, parties...)
		return rv
	}
	// rv is just a shorter way to define a []types.PartyType
	rz := func(roles ...types.PartyType) []types.PartyType {
		rv := make([]types.PartyType, 0, len(roles))
		rv = append(rv, roles...)
		return rv
	}

	scAcct := func(addr string) *authtypes.BaseAccount {
		return authtypes.NewBaseAccount(acc(addr), nil, 0, 0)
	}

	newMsg := func(signers ...string) types.MetadataMsg {
		rv := &types.MsgWriteScopeRequest{Signers: make([]string, len(signers))}
		for i, signer := range signers {
			rv.Signers[i] = accStr(signer)
		}
		return rv
	}
	newMsgType := types.TypeURLMsgWriteScopeRequest

	gi := func(grantee, granter string) GrantInfo {
		return GrantInfo{
			Grantee: acc(grantee),
			Granter: acc(granter),
			MsgType: newMsgType,
		}
	}
	delGetAuthCall := func(grantee, granter, name string) GetAuthorizationCall {
		return GetAuthorizationCall{
			GrantInfo: gi(grantee, granter),
			Result: GetAuthorizationResult{
				Auth: NewMockAuthorization(name, authz.AcceptResponse{Accept: true, Delete: true}, nil),
				Exp:  nil,
			},
		}
	}
	delGrantCall := func(grantee, granter, err string) DeleteGrantCall {
		rv := DeleteGrantCall{
			GrantInfo: gi(grantee, granter),
			Result:    nil,
		}
		if len(err) > 0 {
			rv.Result = errors.New(err)
		}
		return rv
	}

	originator := types.PartyType_PARTY_TYPE_ORIGINATOR
	servicer := types.PartyType_PARTY_TYPE_SERVICER
	investor := types.PartyType_PARTY_TYPE_INVESTOR
	custodian := types.PartyType_PARTY_TYPE_CUSTODIAN
	owner := types.PartyType_PARTY_TYPE_OWNER
	affiliate := types.PartyType_PARTY_TYPE_AFFILIATE
	omnibus := types.PartyType_PARTY_TYPE_OMNIBUS
	provenance := types.PartyType_PARTY_TYPE_PROVENANCE
	controller := types.PartyType_PARTY_TYPE_CONTROLLER
	validator := types.PartyType_PARTY_TYPE_VALIDATOR

	tests := []struct {
		name             string
		reqParties       []types.Party
		availableParties []types.Party
		reqRoles         []types.PartyType
		msg              types.MetadataMsg
		authKeeper       *MockAuthKeeper
		authzKeeper      *MockAuthzKeeper
		expParties       []*keeper.PartyDetails
		expErr           string
	}{
		{
			name:             "nil parties and roles",
			reqParties:       nil,
			availableParties: nil,
			reqRoles:         nil,
			msg:              newMsg("signer1"),
			authKeeper:       NewMockAuthKeeper(),
			authzKeeper:      NewMockAuthzKeeper(),
			expParties:       []*keeper.PartyDetails{},
			expErr:           "",
		},
		{
			name:             "err from associateAuthorizations",
			reqParties:       ptz(pt("party1", originator, false)),
			availableParties: nil,
			reqRoles:         nil,
			msg:              newMsg("signer1"),
			authKeeper:       NewMockAuthKeeper(),
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(
				delGetAuthCall("signer1", "party1", "one"),
			).WithDeleteGrantResults(
				delGrantCall("signer1", "party1", "test_error_from_DeleteGrant"),
			),
			expParties: nil,
			expErr:     "test_error_from_DeleteGrant",
		},
		{
			name:             "required party missing signatures",
			reqParties:       ptz(pt("party1", servicer, false), pt("party2", investor, false)),
			availableParties: nil,
			reqRoles:         rz(owner),
			msg:              newMsg("signer1"),
			authKeeper:       NewMockAuthKeeper(),
			authzKeeper:      NewMockAuthzKeeper(),
			expParties:       nil,
			expErr: fmt.Sprintf("missing required signatures: %s (SERVICER), %s (INVESTOR)",
				accStr("party1"), accStr("party2")),
		},
		{
			name: "associateAuthorizations not called on signed required or optional",
			// This test will have 3 parties and 2 signers.
			// One party will be a signer and the other signer will be an outside address.
			// The second party won't be required.
			// The third party will be required but won't have a signer or authorizations.
			// Other authorizations will be set up to cause an error in order to
			// demonstrate that they are not being looked up/used.
			reqParties: ptz(
				pt("party1", custodian, false),
				pt("party2", owner, true),
				pt("party3", affiliate, false)),
			availableParties: nil,
			reqRoles:         nil,
			msg:              newMsg("party1", "other_signer"),
			authKeeper:       NewMockAuthKeeper(),
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(
				delGetAuthCall("party1", "party2", "one-two"),
				delGetAuthCall("other_signer", "party1", "other-one"),
				delGetAuthCall("other_signer", "party2", "other-two"),
			).WithDeleteGrantResults(
				delGrantCall("party1", "party2", "test_error: delete called for party1 party2"),
				delGrantCall("other_signer", "party1", "test_error: delete called for other_signer party1"),
				delGrantCall("other_signer", "party2", "test_error: delete called for other_signer party2"),
			),
			expParties: nil,
			expErr:     fmt.Sprintf("missing required signature: %s (AFFILIATE)", accStr("party3")),
		},

		{
			name:             "err from associateAuthorizationsForRoles",
			reqParties:       nil,
			availableParties: ptz(pt("party1", omnibus, false), pt("party2", controller, false)),
			reqRoles:         rz(omnibus, controller),
			msg:              newMsg("signer1"),
			authKeeper:       NewMockAuthKeeper(),
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(
				delGetAuthCall("signer1", "party2", "one"),
			).WithDeleteGrantResults(
				delGrantCall("signer1", "party2", "test_error_deleting_grant"),
			),
			expParties: nil,
			expErr:     "test_error_deleting_grant",
		},
		{
			name:             "required roles missing signed parties",
			reqParties:       nil,
			availableParties: ptz(pt("party1", validator, false)),
			reqRoles:         rz(originator, provenance, validator, originator),
			msg:              newMsg("party1"),
			authKeeper:       NewMockAuthKeeper(),
			authzKeeper:      NewMockAuthzKeeper(),
			expParties:       nil,
			expErr:           "missing signers for roles required by spec: ORIGINATOR need 2 have 0, PROVENANCE need 1 have 0",
		},
		{
			name: "associateAuthorizationsForRoles only called as needed",
			// This test will have 3 parties each with unique roles.
			// Only two of the roles will be required.
			// One required role will be signed for directly.
			// The other will end up not having a signer.
			// Authorizations will be set up to error if looked up/used for pairs that shouldn't be.
			reqParties: nil,
			availableParties: ptz(
				pt("party1", owner, true),
				pt("party2", investor, true),
				pt("party3", omnibus, true)),
			reqRoles:   rz(owner, investor),
			msg:        newMsg("party1", "other_signer"),
			authKeeper: NewMockAuthKeeper(),
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(
				delGetAuthCall("party1", "party3", "one-three"),
				delGetAuthCall("other_signer", "party1", "other-one"),
				delGetAuthCall("other_signer", "party3", "other-three"),
			).WithDeleteGrantResults(
				delGrantCall("party1", "party3", "test_error_deleting_grant: party1 party3"),
				delGrantCall("other_signer", "party1", "test_error_deleting_grant: other_signer party1"),
				delGrantCall("other_signer", "party3", "test_error_deleting_grant: other_signer party3"),
			),
			expParties: nil,
			expErr:     "missing signers for roles required by spec: INVESTOR need 1 have 0",
		},
		{
			name:             "required role fulfillment cannot come from reqParties",
			reqParties:       ptz(pt("party1", owner, false)),
			availableParties: nil,
			reqRoles:         rz(owner),
			msg:              newMsg("party1"),
			authKeeper:       NewMockAuthKeeper(),
			authzKeeper:      NewMockAuthzKeeper(),
			expParties:       nil,
			expErr:           "missing signers for roles required by spec: OWNER need 1 have 0",
		},

		{
			name:             "provenance non-smart-contract party ignored in reqParties",
			reqParties:       ptz(pt("party1", provenance, false)),
			availableParties: nil,
			reqRoles:         nil,
			msg:              newMsg("party1"),
			authKeeper:       NewMockAuthKeeper(),
			authzKeeper:      NewMockAuthzKeeper(),
			expParties: []*keeper.PartyDetails{
				keeper.TestablePartyDetails{
					Address:         accStr("party1"),
					Role:            provenance,
					Optional:        false,
					Acc:             nil,
					Signer:          accStr("party1"),
					SignerAcc:       nil,
					CanBeUsedBySpec: false,
					UsedBySpec:      false,
				}.Real(),
			},
			expErr: "",
		},
		{
			name:             "provenance party not smart contract",
			reqParties:       nil,
			availableParties: ptz(pt("party1", provenance, false)),
			reqRoles:         rz(provenance),
			msg:              newMsg("party1"),
			authKeeper:       NewMockAuthKeeper(), // will return nil by default, so no need to mock it specifically.
			authzKeeper:      NewMockAuthzKeeper(),
			expParties:       nil,
			expErr:           fmt.Sprintf("account %q has role PROVENANCE but is not a smart contract", accStr("party1")),
		},
		{
			name:             "non-provenance smart contract account in reqParties ignored",
			reqParties:       ptz(pt("party1", owner, false)),
			availableParties: nil,
			reqRoles:         nil,
			msg:              newMsg("party1"),
			authKeeper: NewMockAuthKeeper().WithGetAccountResults(
				NewGetAccountCall(acc("party1"), scAcct("party1")),
			),
			authzKeeper: NewMockAuthzKeeper(),
			expParties: []*keeper.PartyDetails{
				keeper.TestablePartyDetails{
					Address:         accStr("party1"),
					Role:            owner,
					Optional:        false,
					Acc:             nil,
					Signer:          accStr("party1"),
					SignerAcc:       nil,
					CanBeUsedBySpec: false,
					UsedBySpec:      false,
				}.Real(),
			},
			expErr: "",
		},
		{
			name:             "smart contract not provenance party",
			reqParties:       nil,
			availableParties: ptz(pt("party1", owner, false)),
			reqRoles:         rz(owner),
			msg:              newMsg("party1"),
			authKeeper: NewMockAuthKeeper().WithGetAccountResults(
				NewGetAccountCall(acc("party1"), scAcct("party1")),
			),
			authzKeeper: NewMockAuthzKeeper(),
			expParties:  nil,
			expErr:      fmt.Sprintf("account %q is a smart contract but does not have the PROVENANCE role", accStr("party1")),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			k := s.app.MetadataKeeper
			origAuthKeeper := k.SetAuthKeeper(tc.authKeeper)
			origAuthzKeeper := k.SetAuthzKeeper(tc.authzKeeper)
			defer func() {
				k.SetAuthKeeper(origAuthKeeper)
				k.SetAuthzKeeper(origAuthzKeeper)
			}()

			parties, err := k.ValidateAllRequiredPartiesSigned(s.FreshCtx(), tc.reqParties, tc.availableParties, tc.reqRoles, tc.msg)
			s.AssertErrorValue(err, tc.expErr, "ValidateSignersWithParties error")
			s.Assert().Equal(tc.expParties, parties, "ValidateSignersWithParties parties")
		})
	}
}

func (s *AuthzTestSuite) TestValidateAllRequiredPartiesSigned_CountAuthorizations() {
	// Two addrs in three parties (1, 1, and 2), and one signer (A),
	// Two authz count authorizations:
	//  - granter 1, grantee A, count: 1
	//  - granter 2, grantee A, count: 2
	// Ensure first auth works for both then gets deleted.
	// Ensure second auth is updated.
	// One of the parties for 1 is required.
	// The other has available and has a required role.

	acc := func(addr string) sdk.AccAddress {
		if len(addr) == 0 {
			return nil
		}
		return sdk.AccAddress(addr)
	}
	accStr := func(addr string) string {
		if len(addr) == 0 {
			return ""
		}
		return acc(addr).String()
	}

	party1 := "party1"
	party2 := "party2"
	signer := "signer"
	reqParties := []types.Party{
		{
			Address:  accStr(party1),
			Role:     types.PartyType_PARTY_TYPE_OWNER,
			Optional: false,
		},
		{
			Address:  accStr(party2),
			Role:     types.PartyType_PARTY_TYPE_SERVICER,
			Optional: false,
		},
	}
	availableParties := []types.Party{
		{
			Address:  accStr(party1),
			Role:     types.PartyType_PARTY_TYPE_VALIDATOR,
			Optional: false,
		},
	}
	reqRoles := []types.PartyType{types.PartyType_PARTY_TYPE_VALIDATOR}
	expDetails := []*keeper.PartyDetails{
		keeper.TestablePartyDetails{
			Address:         accStr(party1),
			Role:            types.PartyType_PARTY_TYPE_VALIDATOR,
			Optional:        true,
			Acc:             acc(party1),
			Signer:          "",
			SignerAcc:       acc(signer),
			CanBeUsedBySpec: true,
			UsedBySpec:      true,
		}.Real(),
		keeper.TestablePartyDetails{
			Address:         accStr(party1),
			Role:            types.PartyType_PARTY_TYPE_OWNER,
			Optional:        false,
			Acc:             acc(party1),
			Signer:          "",
			SignerAcc:       acc(signer),
			CanBeUsedBySpec: false,
			UsedBySpec:      false,
		}.Real(),
		keeper.TestablePartyDetails{
			Address:         accStr(party2),
			Role:            types.PartyType_PARTY_TYPE_SERVICER,
			Optional:        false,
			Acc:             acc(party2),
			Signer:          "",
			SignerAcc:       acc(signer),
			CanBeUsedBySpec: false,
			UsedBySpec:      false,
		}.Real(),
	}

	msg := &types.MsgDeleteScopeRequest{Signers: []string{accStr(signer)}}
	msgTypeURL := types.TypeURLMsgDeleteScopeRequest

	ctx := s.FreshCtx()

	// first grant: signer can sign for party1 one time.
	auth1 := authz.NewCountAuthorization(msgTypeURL, 1)
	err := s.app.AuthzKeeper.SaveGrant(ctx, acc(signer), acc(party1), auth1, nil)
	s.Require().NoError(err, "SaveGrant signer can sign for party1: 1 use")

	// second grant: signer can sign for party2 two times.
	auth2 := authz.NewCountAuthorization(msgTypeURL, 2)
	err = s.app.AuthzKeeper.SaveGrant(ctx, acc(signer), acc(party2), auth2, nil)
	s.Require().NoError(err, "SaveGrant signer can sign for party2: 2 uses")

	details, err := s.app.MetadataKeeper.ValidateAllRequiredPartiesSigned(ctx, reqParties, availableParties, reqRoles, msg)
	s.Require().NoError(err, "ValidateSignersWithParties error")
	s.Assert().Equal(expDetails, details, "ValidateSignersWithParties party details")

	auth1Final, _ := s.app.AuthzKeeper.GetAuthorization(ctx, acc(signer), acc(party1), msgTypeURL)
	s.Assert().Nil(auth1Final, "GetAuthorization after only allowed use")

	auth2Final, _ := s.app.AuthzKeeper.GetAuthorization(ctx, acc(signer), acc(party2), msgTypeURL)
	s.Assert().NotNil(auth2Final, "GetAuthorization after first of two uses")
	actual := auth2Final.(*authz.CountAuthorization).AllowedAuthorizations
	s.Assert().Equal(1, int(actual), "number of uses left after first of two uses")
}

func TestAssociateSigners(t *testing.T) {
	// pd is a short way to create a PartyDetails with only what we care about in this test.
	pd := func(address string, acc sdk.AccAddress, signer string, signerAcc sdk.AccAddress) *keeper.PartyDetails {
		return keeper.TestablePartyDetails{
			Address:   address,
			Acc:       acc,
			Signer:    signer,
			SignerAcc: signerAcc,
		}.Real()
	}
	// pdz is a shorter varargs way to define a []*keeper.PartyDetails.
	pdz := func(parties ...*keeper.PartyDetails) []*keeper.PartyDetails {
		rv := make([]*keeper.PartyDetails, 0, len(parties))
		rv = append(rv, parties...)
		return rv
	}
	// sw is a shorter varargs way to define a *keeper.SignersWrapper.
	sw := func(addrs ...string) *keeper.SignersWrapper {
		return keeper.NewSignersWrapper(addrs)
	}
	signersWrapperCopy := func(signers *keeper.SignersWrapper) *keeper.SignersWrapper {
		if signers == nil {
			return nil
		}
		return keeper.NewSignersWrapper(signers.Strings())
	}
	// acc is a shorter way to cast a string to an AccAddress.
	acc := func(acc string) sdk.AccAddress {
		return sdk.AccAddress(acc)
	}
	// accStr is a shorter way to cast a string to an AccAddress and get it's bech32.
	accStr := func(acc string) string {
		return sdk.AccAddress(acc).String()
	}

	// partyStr gets a string of the golang code that would make the provided party for these tests.
	partyStr := func(p *keeper.PartyDetails) string {
		if p == nil {
			return "nil"
		}
		party := p.Testable()
		var addrVal string
		addrAcc, err := sdk.AccAddressFromBech32(party.Address)
		if err == nil {
			addrVal = fmt.Sprintf("accStr(%q)", string(addrAcc))
		} else {
			addrVal = fmt.Sprintf("%q", party.Address)
		}
		accVal := "nil"
		if party.Acc != nil {
			accVal = fmt.Sprintf("acc(%q)", string(party.Acc))
		}
		var sigVal string
		sigAcc, err := sdk.AccAddressFromBech32(party.Signer)
		if err == nil {
			sigVal = fmt.Sprintf("accStr(%q)", string(sigAcc))
		} else {
			sigVal = fmt.Sprintf("%q", party.Signer)
		}
		sigAccVal := "nil"
		if party.SignerAcc != nil {
			sigAccVal = fmt.Sprintf("acc(%q)", string(party.SignerAcc))
		}
		return fmt.Sprintf("pd(%s, %s, %s, %s)", addrVal, accVal, sigVal, sigAccVal)
	}
	// partiesStr gets a string of the golang code that would make the provided parties for these tests.
	partiesStr := func(parties []*keeper.PartyDetails) string {
		if parties == nil {
			return "nil"
		}
		strs := make([]string, len(parties))
		for i, party := range parties {
			strs[i] = partyStr(party)
		}
		if len(strs) <= 2 {
			return fmt.Sprintf("pdz(%s)", strings.Join(strs, ", "))
		}
		return fmt.Sprintf("pdz(\n\t\t%s,\n\t)", strings.Join(strs, ",\n\t\t"))
	}
	// signersStr gets a string of the golang code that would make the provided SignersWrapper for these tests.
	signersStr := func(sw *keeper.SignersWrapper) string {
		if sw == nil {
			return "nil"
		}
		sigs := sw.Strings()
		strs := make([]string, len(sigs))
		for i, sig := range sigs {
			strs[i] = fmt.Sprintf("%q", sig)
		}
		return fmt.Sprintf("sw(%s)", strings.Join(strs, ", "))
	}

	// signersReversed creates a copy of the provided SignersWrapper in reversed order. Nil in = nil out.
	signersReversed := func(signers *keeper.SignersWrapper) *keeper.SignersWrapper {
		if signers == nil {
			return nil
		}
		sigs := signers.Strings()
		revSigs := make([]string, len(sigs))
		for i, sig := range sigs {
			revSigs[len(revSigs)-i-1] = sig
		}
		return keeper.NewSignersWrapper(revSigs)
	}
	// signersShuffled creates a copy of the provided SignersWrapper and shuffles the entries. Nil in = nil out.
	signersShuffled := func(r *rand.Rand, signers *keeper.SignersWrapper) *keeper.SignersWrapper {
		if signers == nil {
			return nil
		}
		sigs := signers.Strings()
		shufSigs := make([]string, len(sigs))
		shufSigs = append(shufSigs, sigs...)
		r.Shuffle(len(shufSigs), func(i, j int) {
			shufSigs[i], shufSigs[j] = shufSigs[j], shufSigs[i]
		})
		return keeper.NewSignersWrapper(shufSigs)
	}
	// partiesShuffled creates a copy of the provided party slices and shuffles the entries.
	// Both parties and expParties must have the same length and if one is nil, the other must be too.
	// The entries are shuffled in tandem. E.g. if parties becomes [1, 0, 2] then expParties will also have the order [1, 0, 2].
	// Nil in = nil out.
	partiesShuffled := func(r *rand.Rand, parties, expParties []*keeper.PartyDetails) ([]*keeper.PartyDetails, []*keeper.PartyDetails) {
		if (parties == nil && expParties != nil) || (parties != nil && expParties == nil) || (len(parties) != len(expParties)) {
			panic("test definition failure: parties and expParties should always have the same number of entries")
		}
		if parties == nil {
			return nil, nil
		}
		rvp := make([]*keeper.PartyDetails, 0, len(parties))
		rvp = append(rvp, parties...)
		rve := make([]*keeper.PartyDetails, 0, len(expParties))
		rve = append(rve, expParties...)
		r.Shuffle(len(rve), func(i, j int) {
			rve[i], rve[j] = rve[j], rve[i]
			rvp[i], rvp[j] = rvp[j], rvp[i]
		})
		return rvp, rve
	}

	type testCase struct {
		name       string
		parties    []*keeper.PartyDetails
		signers    *keeper.SignersWrapper
		expParties []*keeper.PartyDetails
	}

	tests := []testCase{
		{
			name:       "nil nil",
			parties:    nil,
			signers:    nil,
			expParties: nil,
		},
		{
			name:       "empty nil",
			parties:    pdz(),
			signers:    nil,
			expParties: pdz(),
		},
		{
			name:       "nil empty",
			parties:    nil,
			signers:    sw(),
			expParties: nil,
		},
		{
			name:       "empty empty",
			parties:    pdz(),
			signers:    sw(),
			expParties: pdz(),
		},
		{
			name:       "3 parties nil signers",
			parties:    pdz(pd("addr1", nil, "", nil), pd("addr2", nil, "", nil), pd("addr3", nil, "", nil)),
			signers:    nil,
			expParties: pdz(pd("addr1", nil, "", nil), pd("addr2", nil, "", nil), pd("addr3", nil, "", nil)),
		},
		{
			name:       "3 parties empty signers",
			parties:    pdz(pd("addr1", nil, "", nil), pd("addr2", nil, "", nil), pd("addr3", nil, "", nil)),
			signers:    sw(),
			expParties: pdz(pd("addr1", nil, "", nil), pd("addr2", nil, "", nil), pd("addr3", nil, "", nil)),
		},
		{
			name:       "nil parties 3 signers",
			parties:    nil,
			signers:    sw("addr1", "addr2", "addr3"),
			expParties: nil,
		},
		{
			name:       "empty parties 3 signers",
			parties:    pdz(),
			signers:    sw("addr1", "addr2", "addr3"),
			expParties: pdz(),
		},
		{
			name:       "1 party only string is 1 signer",
			parties:    pdz(pd("match", nil, "", acc("stomped"))),
			signers:    sw("match"),
			expParties: pdz(pd("match", nil, "match", nil)),
		},
		{
			name:       "1 party only acc is 1 signer o",
			parties:    pdz(pd("", acc("acc_match"), "", acc("stomped"))),
			signers:    sw(accStr("acc_match")),
			expParties: pdz(pd(accStr("acc_match"), acc("acc_match"), accStr("acc_match"), nil)),
		},
		{
			name:       "1 party conflicting string acc is 1 signer ",
			parties:    pdz(pd("match", acc("acc_match"), "", acc("stomped"))),
			signers:    sw("match"),
			expParties: pdz(pd("match", acc("acc_match"), "match", nil)),
		},
		{
			name:       "1 party conflicting string acc signer matches acc but not string",
			parties:    pdz(pd("match", acc("acc_match"), "", acc("not-stomped"))),
			signers:    sw("acc_match"),
			expParties: pdz(pd("match", acc("acc_match"), "", acc("not-stomped"))),
		},
		{
			name:       "1 party is in 10 signers",
			parties:    pdz(pd("addr6", nil, "", nil)),
			signers:    sw("addr1", "addr2", "addr3", "addr4", "addr5", "addr6", "addr7", "addr8", "addr9", "addr10"),
			expParties: pdz(pd("addr6", nil, "addr6", nil)),
		},
		{
			name:       "1 party is not in 10 signers",
			parties:    pdz(pd("no-match", nil, "", nil)),
			signers:    sw("addr1", "addr2", "addr3", "addr4", "addr5", "addr6", "addr7", "addr8", "addr9", "addr10"),
			expParties: pdz(pd("no-match", nil, "", nil)),
		},
		{
			name:       "1 party already has signer is in signers differently",
			parties:    pdz(pd(accStr("my-addr"), acc("my-other-addr"), accStr("change-me"), acc("don't-change-me-bro"))),
			signers:    sw(accStr("my-addr")),
			expParties: pdz(pd(accStr("my-addr"), acc("my-other-addr"), accStr("my-addr"), nil)),
		},
		{
			name:       "2 parties both in 2 signers same order",
			parties:    pdz(pd("match-1", nil, "", nil), pd("match-2", nil, "", nil)),
			signers:    sw("match-1", "match-2"),
			expParties: pdz(pd("match-1", nil, "match-1", nil), pd("match-2", nil, "match-2", nil)),
		},
		{
			name:       "2 parties both in 2 signers diff order",
			parties:    pdz(pd("match-1", nil, "", nil), pd("match-2", nil, "", nil)),
			signers:    sw("match-2", "match-1"),
			expParties: pdz(pd("match-1", nil, "match-1", nil), pd("match-2", nil, "match-2", nil)),
		},
		{
			name:       "2 parties first is first of 2 signers",
			parties:    pdz(pd("addr1", nil, "", nil), pd("addr2", nil, "", nil)),
			signers:    sw("addr1", "addr3"),
			expParties: pdz(pd("addr1", nil, "addr1", nil), pd("addr2", nil, "", nil)),
		},
		{
			name:       "2 parties first is second of 2 signers",
			parties:    pdz(pd("addr1", nil, "", nil), pd("addr2", nil, "", nil)),
			signers:    sw("addr3", "addr1"),
			expParties: pdz(pd("addr1", nil, "addr1", nil), pd("addr2", nil, "", nil)),
		},
		{
			name:       "2 parties second is first of 2 signers",
			parties:    pdz(pd("addr1", nil, "", nil), pd("addr2", nil, "", nil)),
			signers:    sw("addr2", "addr3"),
			expParties: pdz(pd("addr1", nil, "", nil), pd("addr2", nil, "addr2", nil)),
		},
		{
			name:       "2 parties second is first of 2 signers",
			parties:    pdz(pd("addr1", nil, "", nil), pd("addr2", nil, "", nil)),
			signers:    sw("addr3", "addr2"),
			expParties: pdz(pd("addr1", nil, "", nil), pd("addr2", nil, "addr2", nil)),
		},
		{
			name:       "3 parties all in 10 signers",
			parties:    pdz(pd("addr6", nil, "", nil), pd("addr8", nil, "", nil), pd("addr2", nil, "", nil)),
			signers:    sw("addr1", "addr2", "addr3", "addr4", "addr5", "addr6", "addr7", "addr8", "addr9", "addr10"),
			expParties: pdz(pd("addr6", nil, "addr6", nil), pd("addr8", nil, "addr8", nil), pd("addr2", nil, "addr2", nil)),
		},
		{
			name: "3 parties only accs all in 10 signers",
			parties: pdz(
				pd("", acc("addr6"), "", nil),
				pd("", acc("addr2"), "", nil),
				pd("", acc("addr8"), "", nil),
			),
			signers: sw(
				accStr("addr1"), accStr("addr2"), accStr("addr3"), accStr("addr4"), accStr("addr5"),
				accStr("addr6"), accStr("addr7"), accStr("addr8"), accStr("addr9"), accStr("addr10"),
			),
			expParties: pdz(
				pd(accStr("addr6"), acc("addr6"), accStr("addr6"), nil),
				pd(accStr("addr2"), acc("addr2"), accStr("addr2"), nil),
				pd(accStr("addr8"), acc("addr8"), accStr("addr8"), nil),
			),
		},
		{
			name:       "3 parties first not in 10 signers",
			parties:    pdz(pd("addr6x", nil, "", nil), pd("addr8", nil, "", nil), pd("addr2", nil, "", nil)),
			signers:    sw("addr1", "addr2", "addr3", "addr4", "addr5", "addr6", "addr7", "addr8", "addr9", "addr10"),
			expParties: pdz(pd("addr6x", nil, "", nil), pd("addr8", nil, "addr8", nil), pd("addr2", nil, "addr2", nil)),
		},
		{
			name:       "3 parties second not in 10 signers",
			parties:    pdz(pd("addr6", nil, "", nil), pd("addr8x", nil, "", nil), pd("addr2", nil, "", nil)),
			signers:    sw("addr1", "addr2", "addr3", "addr4", "addr5", "addr6", "addr7", "addr8", "addr9", "addr10"),
			expParties: pdz(pd("addr6", nil, "addr6", nil), pd("addr8x", nil, "", nil), pd("addr2", nil, "addr2", nil)),
		},
		{
			name:       "3 parties third not in 10 signers",
			parties:    pdz(pd("addr6", nil, "", nil), pd("addr8", nil, "", nil), pd("addr2x", nil, "", nil)),
			signers:    sw("addr1", "addr2", "addr3", "addr4", "addr5", "addr6", "addr7", "addr8", "addr9", "addr10"),
			expParties: pdz(pd("addr6", nil, "addr6", nil), pd("addr8", nil, "addr8", nil), pd("addr2x", nil, "", nil)),
		},
		{
			name:       "3 parties only first in 10 signers",
			parties:    pdz(pd("addr6", nil, "", nil), pd("addr8x", nil, "", nil), pd("addr2x", nil, "", nil)),
			signers:    sw("addr1", "addr2", "addr3", "addr4", "addr5", "addr6", "addr7", "addr8", "addr9", "addr10"),
			expParties: pdz(pd("addr6", nil, "addr6", nil), pd("addr8x", nil, "", nil), pd("addr2x", nil, "", nil)),
		},
		{
			name:       "3 parties only second in 10 signers",
			parties:    pdz(pd("addr6x", nil, "", nil), pd("addr8", nil, "", nil), pd("addr2x", nil, "", nil)),
			signers:    sw("addr1", "addr2", "addr3", "addr4", "addr5", "addr6", "addr7", "addr8", "addr9", "addr10"),
			expParties: pdz(pd("addr6x", nil, "", nil), pd("addr8", nil, "addr8", nil), pd("addr2x", nil, "", nil)),
		},
		{
			name:       "3 parties only third in 10 signers",
			parties:    pdz(pd("addr6x", nil, "", nil), pd("addr8x", nil, "", nil), pd("addr2", nil, "", nil)),
			signers:    sw("addr1", "addr2", "addr3", "addr4", "addr5", "addr6", "addr7", "addr8", "addr9", "addr10"),
			expParties: pdz(pd("addr6x", nil, "", nil), pd("addr8x", nil, "", nil), pd("addr2", nil, "addr2", nil)),
		},
		{
			name:       "3 parties none in 10 signers",
			parties:    pdz(pd("addr6x", nil, "", nil), pd("addr8x", nil, "", nil), pd("addr2x", nil, "", nil)),
			signers:    sw("addr1", "addr2", "addr3", "addr4", "addr5", "addr6", "addr7", "addr8", "addr9", "addr10"),
			expParties: pdz(pd("addr6x", nil, "", nil), pd("addr8x", nil, "", nil), pd("addr2x", nil, "", nil)),
		},
		{
			name: "3 same parties 1 other 1 signer for the 3",
			parties: pdz(
				// Since the string versions exist, the acc should be ignored, so I'm using that field as a differentiator.
				pd("addr1", acc("one"), "", nil),
				pd("addr1", acc("two"), "", nil),
				pd("addr2", nil, "", nil),
				pd("addr1", acc("three"), "", nil),
			),
			signers: sw("addr1"),
			expParties: pdz(
				pd("addr1", acc("one"), "addr1", nil),
				pd("addr1", acc("two"), "addr1", nil),
				pd("addr2", nil, "", nil),
				pd("addr1", acc("three"), "addr1", nil),
			),
		},
		{
			name: "3 same parties 1 other both signers",
			parties: pdz(
				// Since the string versions exist, the acc should be ignored, so I'm using that field as a differentiator.
				pd("addr1", acc("one"), "", nil),
				pd("addr1", acc("two"), "", nil),
				pd("addr2", nil, "", nil),
				pd("addr1", acc("three"), "", nil),
			),
			signers: sw("addr1", "addr2"),
			expParties: pdz(
				pd("addr1", acc("one"), "addr1", nil),
				pd("addr1", acc("two"), "addr1", nil),
				pd("addr2", nil, "addr2", nil),
				pd("addr1", acc("three"), "addr1", nil),
			),
		},
		{
			name: "10 parties 8 covered by 3 signers",
			parties: pdz(
				pd("addr1", acc("addr1-one"), "", nil),
				pd("addr1", acc("addr1-two"), "", nil),
				pd("addr1", acc("addr1-three"), "", nil),
				pd("addr1", acc("addr1-four"), "", nil),
				pd("addr2", acc("addr2-one"), "", nil),
				pd("addr2", acc("addr2-two"), "", nil),
				pd("addr2", acc("addr2-three"), "", nil),
				pd("addr3", acc("addr3-one"), "", nil),
				pd("addr4", acc("addr4-one"), "", nil),
				pd("addr5", acc("addr5-one"), "", nil),
			),
			signers: sw("addr1", "addr2", "addr4"),
			expParties: pdz(
				pd("addr1", acc("addr1-one"), "addr1", nil),
				pd("addr1", acc("addr1-two"), "addr1", nil),
				pd("addr1", acc("addr1-three"), "addr1", nil),
				pd("addr1", acc("addr1-four"), "addr1", nil),
				pd("addr2", acc("addr2-one"), "addr2", nil),
				pd("addr2", acc("addr2-two"), "addr2", nil),
				pd("addr2", acc("addr2-three"), "addr2", nil),
				pd("addr3", acc("addr3-one"), "", nil),
				pd("addr4", acc("addr4-one"), "addr4", nil),
				pd("addr5", acc("addr5-one"), "", nil),
			),
		},
	}

	// Copy all tests four times.
	// Once with reversed parties. Once with reversed signers.
	// Once with both reversed. And once with both shuffled.
	revPartiesTests := make([]testCase, len(tests))
	revSigsTests := make([]testCase, len(tests))
	revBothTests := make([]testCase, len(tests))
	shuffledTests := make([]testCase, len(tests))

	for i, tc := range tests {
		revPartiesTests[i] = testCase{
			name:       "rev parties " + tc.name,
			parties:    partiesReversed(tc.parties),
			signers:    signersWrapperCopy(tc.signers),
			expParties: partiesReversed(tc.expParties),
		}
		revSigsTests[i] = testCase{
			name:       "rev sigs " + tc.name,
			parties:    partiesCopy(tc.parties),
			signers:    signersReversed(tc.signers),
			expParties: partiesCopy(tc.expParties),
		}
		revBothTests[i] = testCase{
			name:       "rev both " + tc.name,
			parties:    partiesReversed(tc.parties),
			signers:    signersReversed(tc.signers),
			expParties: partiesReversed(tc.expParties),
		}
		// Using a hard-coded (randomly chosen) seed value here to make life easier if one of these fails.
		// The purpose is to just have them in an order other than as defined (hopefully).
		r := rand.New(rand.NewSource(int64(58720 * i)))
		shufP, shufE := partiesShuffled(r, tc.parties, tc.expParties)
		shuffledTests[i] = testCase{
			name:       "shuffled " + tc.name,
			parties:    shufP,
			signers:    signersShuffled(r, tc.signers),
			expParties: shufE,
		}
	}

	tests = append(tests, revPartiesTests...)
	tests = append(tests, revSigsTests...)
	tests = append(tests, revBothTests...)
	tests = append(tests, shuffledTests...)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			orig := partiesCopy(tc.parties)
			keeper.AssociateSigners(tc.parties, tc.signers)
			if !assert.Equal(t, tc.expParties, tc.parties, "parties after associateSigners") {
				// If the assertion failed, the output will contain the differences.
				// Since some input might not be obvious though, include them now.
				t.Logf("tests = append(tests, {\n\tname: %q,\n\tparties: %s,\n\tsigners: %s,\n\texpParties: %s,\n})",
					tc.name, partiesStr(orig), signersStr(tc.signers), partiesStr(tc.expParties))
			}
		})
	}
}

func TestFindUnsignedRequired(t *testing.T) {
	// pd is a short way to create a PartyDetails with only what we care about in this test.
	pd := func(address string, optional bool, signer string, signerAcc sdk.AccAddress) *keeper.PartyDetails {
		return keeper.TestablePartyDetails{
			Address:   address,
			Optional:  optional,
			Signer:    signer,
			SignerAcc: signerAcc,
		}.Real()
	}
	// pdz is just a shorter way to define a []*keeper.PartyDetails
	pdz := func(parties ...*keeper.PartyDetails) []*keeper.PartyDetails {
		rv := make([]*keeper.PartyDetails, 0, len(parties))
		rv = append(rv, parties...)
		return rv
	}

	addrAcc := sdk.AccAddress("a_signer_address____")
	addr := addrAcc.String()

	tests := []struct {
		name    string
		parties []*keeper.PartyDetails
		exp     []*keeper.PartyDetails
	}{
		{
			name:    "nil",
			parties: nil,
			exp:     nil,
		},
		{
			name:    "empty",
			parties: pdz(),
			exp:     nil,
		},
		{
			name:    "1 party not required not signed",
			parties: pdz(pd("one", true, "", nil)),
			exp:     nil,
		},
		{
			name:    "1 party not required signer only string",
			parties: pdz(pd("one", true, addr, nil)),
			exp:     nil,
		},
		{
			name:    "1 party not required signer only acc",
			parties: pdz(pd("one", true, "", addrAcc)),
			exp:     nil,
		},
		{
			name:    "1 party not required signer both",
			parties: pdz(pd("one", true, addr, addrAcc)),
			exp:     nil,
		},
		{
			name:    "1 party required not signed",
			parties: pdz(pd("one", false, "", nil)),
			exp:     pdz(pd("one", false, "", nil)),
		},
		{
			name:    "1 party required signer only string",
			parties: pdz(pd("one", false, addr, nil)),
			exp:     nil,
		},
		{
			name:    "1 party required signer only acc",
			parties: pdz(pd("one", false, "", addrAcc)),
			exp:     nil,
		},
		{
			name:    "1 party required signer both",
			parties: pdz(pd("one", false, addr, addrAcc)),
			exp:     nil,
		},

		{
			name: "5 parties 2 are req and signed",
			parties: pdz(
				pd("one", true, addr, nil),
				pd("two", false, addr, nil),
				pd("three", true, addr, nil),
				pd("four", true, "", nil),
				pd("five", false, addr, nil),
			),
			exp: nil,
		},
		{
			name: "5 parties 2 are req only first signed",
			parties: pdz(
				pd("one", true, addr, nil),
				pd("two", false, addr, nil),
				pd("three", true, addr, nil),
				pd("four", true, "", nil),
				pd("five", false, "", nil),
			),
			exp: pdz(pd("five", false, "", nil)),
		},
		{
			name: "5 parties 2 are req only second signed",
			parties: pdz(
				pd("one", true, addr, nil),
				pd("two", false, "", nil),
				pd("three", true, addr, nil),
				pd("four", true, "", nil),
				pd("five", false, addr, nil),
			),
			exp: pdz(pd("two", false, "", nil)),
		},
		{
			name: "5 parties 2 are req neither signed",
			parties: pdz(
				pd("one", true, addr, nil),
				pd("two", false, "", nil),
				pd("three", true, addr, nil),
				pd("four", true, "", nil),
				pd("five", false, "", nil),
			),
			exp: pdz(
				pd("two", false, "", nil),
				pd("five", false, "", nil),
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := keeper.FindUnsignedRequired(tc.parties)
			assert.Equal(t, tc.exp, actual, "findUnsignedRequired")
		})
	}

	t.Run("same references are returned", func(t *testing.T) {
		pd1 := pd("one", false, addr, addrAcc)
		pd2 := pd("two", false, "", nil)
		pd3 := pd("three", false, addr, nil)
		pd4 := pd("four", false, "", nil)
		pd5 := pd("five", false, "", nil)
		pd6 := pd("six", false, "", addrAcc)
		parties := pdz(pd1, pd2, pd3, pd4, pd5, pd6)
		exp := pdz(pd2, pd4, pd5)
		actual := keeper.FindUnsignedRequired(parties)
		if assert.Len(t, actual, len(exp), "findUnsignedRequired returned parties") {
			for i := range exp {
				assert.Same(t, exp[i], actual[i], "findUnsignedRequired returned party [%d]", i)
			}
		}
	})
}

func TestAssociateRequiredRoles(t *testing.T) {
	// pd is a short way to create a PartyDetails with only what we care about in this test.
	pd := func(role types.PartyType, canBeUsed, isUsed bool, signer string, signerAcc sdk.AccAddress) *keeper.PartyDetails {
		return keeper.TestablePartyDetails{
			Role:            role,
			Signer:          signer,
			SignerAcc:       signerAcc,
			CanBeUsedBySpec: canBeUsed,
			UsedBySpec:      isUsed,
		}.Real()
	}
	// pdz is just a shorter way to define a []*keeper.PartyDetails
	pdz := func(parties ...*keeper.PartyDetails) []*keeper.PartyDetails {
		rv := make([]*keeper.PartyDetails, 0, len(parties))
		rv = append(rv, parties...)
		return rv
	}
	// rv is just a shorter way to define a []types.PartyType
	rz := func(roles ...types.PartyType) []types.PartyType {
		rv := make([]types.PartyType, 0, len(roles))
		rv = append(rv, roles...)
		return rv
	}

	// Create some aliases that are shorter than their full names.
	unspecified := types.PartyType_PARTY_TYPE_UNSPECIFIED
	originator := types.PartyType_PARTY_TYPE_ORIGINATOR
	servicer := types.PartyType_PARTY_TYPE_SERVICER
	investor := types.PartyType_PARTY_TYPE_INVESTOR
	custodian := types.PartyType_PARTY_TYPE_CUSTODIAN
	owner := types.PartyType_PARTY_TYPE_OWNER
	affiliate := types.PartyType_PARTY_TYPE_AFFILIATE
	omnibus := types.PartyType_PARTY_TYPE_OMNIBUS
	provenance := types.PartyType_PARTY_TYPE_PROVENANCE
	controller := types.PartyType_PARTY_TYPE_CONTROLLER
	validator := types.PartyType_PARTY_TYPE_VALIDATOR

	allRoles := rz(
		unspecified, originator, servicer, investor, custodian, owner,
		affiliate, omnibus, provenance, controller, validator,
	)

	addrAcc := sdk.AccAddress("simple_test_address_")
	addr := addrAcc.String()

	type testCase struct {
		name       string
		parties    []*keeper.PartyDetails
		reqRoles   []types.PartyType
		exp        []types.PartyType
		expParties []*keeper.PartyDetails
	}

	tests := []testCase{
		{
			name:       "nil nil",
			parties:    nil,
			reqRoles:   nil,
			exp:        nil,
			expParties: nil,
		},
		{
			name:       "empty nil",
			parties:    []*keeper.PartyDetails{},
			reqRoles:   nil,
			exp:        nil,
			expParties: []*keeper.PartyDetails{},
		},
		{
			name:       "nil empty",
			parties:    nil,
			reqRoles:   []types.PartyType{},
			exp:        nil,
			expParties: nil,
		},
		{
			name:       "empty empty",
			parties:    []*keeper.PartyDetails{},
			reqRoles:   []types.PartyType{},
			exp:        nil,
			expParties: []*keeper.PartyDetails{},
		},
		{
			name:       "2 req nil parties",
			parties:    nil,
			reqRoles:   rz(validator, investor),
			exp:        rz(validator, investor),
			expParties: nil,
		},
		{
			name:       "2 req empty parties",
			parties:    pdz(),
			reqRoles:   rz(validator, investor),
			exp:        rz(validator, investor),
			expParties: pdz(),
		},
		{
			name:       "2 parties nil req",
			parties:    pdz(pd(owner, true, false, addr, addrAcc), pd(provenance, true, false, addr, addrAcc)),
			reqRoles:   nil,
			exp:        nil,
			expParties: pdz(pd(owner, true, false, addr, addrAcc), pd(provenance, true, false, addr, addrAcc)),
		},
		{
			name:       "2 parties empty req",
			parties:    pdz(pd(owner, true, false, addr, addrAcc), pd(provenance, true, false, addr, addrAcc)),
			reqRoles:   rz(),
			exp:        nil,
			expParties: pdz(pd(owner, true, false, addr, addrAcc), pd(provenance, true, false, addr, addrAcc)),
		},

		// all single req/party combos of usable/unusable, not/already used, right/wrong role,
		// and both signer fields/only string/only acc/neither.
		{
			name:       "usable, not used, right role, both signer string and acc",
			parties:    pdz(pd(originator, true, false, addr, addrAcc)),
			reqRoles:   rz(originator),
			exp:        nil,
			expParties: pdz(pd(originator, true, true, addr, addrAcc)),
		},
		{
			name:       "usable, not used, right role, only signer string",
			parties:    pdz(pd(originator, true, false, addr, nil)),
			reqRoles:   rz(originator),
			exp:        nil,
			expParties: pdz(pd(originator, true, true, addr, nil)),
		},
		{
			name:       "usable, not used, right role, only signer acc",
			parties:    pdz(pd(originator, true, false, "", addrAcc)),
			reqRoles:   rz(originator),
			exp:        nil,
			expParties: pdz(pd(originator, true, true, "", addrAcc)),
		},
		{
			name:       "usable, not used, right role, no signer",
			parties:    pdz(pd(originator, true, false, "", nil)),
			reqRoles:   rz(originator),
			exp:        rz(originator),
			expParties: pdz(pd(originator, true, false, "", nil)),
		},
		{
			name:       "usable, not used, wrong role, both signer string and acc",
			parties:    pdz(pd(originator, true, false, addr, addrAcc)),
			reqRoles:   rz(servicer),
			exp:        rz(servicer),
			expParties: pdz(pd(originator, true, false, addr, addrAcc)),
		},
		{
			name:       "usable, not used, wrong role, only signer string",
			parties:    pdz(pd(originator, true, false, addr, nil)),
			reqRoles:   rz(servicer),
			exp:        rz(servicer),
			expParties: pdz(pd(originator, true, false, addr, nil)),
		},
		{
			name:       "usable, not used, wrong role, only signer acc",
			parties:    pdz(pd(originator, true, false, "", addrAcc)),
			reqRoles:   rz(servicer),
			exp:        rz(servicer),
			expParties: pdz(pd(originator, true, false, "", addrAcc)),
		},
		{
			name:       "usable, not used, wrong role, no signer",
			parties:    pdz(pd(originator, true, false, "", nil)),
			reqRoles:   rz(servicer),
			exp:        rz(servicer),
			expParties: pdz(pd(originator, true, false, "", nil)),
		},
		{
			name:       "usable, already used, right role, both signer string and acc",
			parties:    pdz(pd(investor, true, true, addr, addrAcc)),
			reqRoles:   rz(investor),
			exp:        rz(investor),
			expParties: pdz(pd(investor, true, true, addr, addrAcc)),
		},
		{
			name:       "usable, already used, right role, only signer string",
			parties:    pdz(pd(investor, true, true, addr, nil)),
			reqRoles:   rz(investor),
			exp:        rz(investor),
			expParties: pdz(pd(investor, true, true, addr, nil)),
		},
		{
			name:       "usable, already used, right role, only signer acc",
			parties:    pdz(pd(investor, true, true, "", addrAcc)),
			reqRoles:   rz(investor),
			exp:        rz(investor),
			expParties: pdz(pd(investor, true, true, "", addrAcc)),
		},
		{
			name:       "usable, already used, right role, no signer",
			parties:    pdz(pd(investor, true, true, "", nil)),
			reqRoles:   rz(investor),
			exp:        rz(investor),
			expParties: pdz(pd(investor, true, true, "", nil)),
		},
		{
			name:       "usable, already used, wrong role, both signer string and acc",
			parties:    pdz(pd(investor, true, true, addr, addrAcc)),
			reqRoles:   rz(omnibus),
			exp:        rz(omnibus),
			expParties: pdz(pd(investor, true, true, addr, addrAcc)),
		},
		{
			name:       "usable, already used, wrong role, only signer string",
			parties:    pdz(pd(investor, true, true, addr, nil)),
			reqRoles:   rz(omnibus),
			exp:        rz(omnibus),
			expParties: pdz(pd(investor, true, true, addr, nil)),
		},
		{
			name:       "usable, already used, wrong role, only signer acc",
			parties:    pdz(pd(investor, true, true, "", addrAcc)),
			reqRoles:   rz(omnibus),
			exp:        rz(omnibus),
			expParties: pdz(pd(investor, true, true, "", addrAcc)),
		},
		{
			name:       "usable, already used, wrong role, no signer",
			parties:    pdz(pd(investor, true, true, "", nil)),
			reqRoles:   rz(omnibus),
			exp:        rz(omnibus),
			expParties: pdz(pd(investor, true, true, "", nil)),
		},
		{
			name:       "unusable, not used, right role, both signer string and acc",
			parties:    pdz(pd(originator, false, false, addr, addrAcc)),
			reqRoles:   rz(originator),
			exp:        rz(originator),
			expParties: pdz(pd(originator, false, false, addr, addrAcc)),
		},
		{
			name:       "unusable, not used, right role, only signer string",
			parties:    pdz(pd(originator, false, false, addr, nil)),
			reqRoles:   rz(originator),
			exp:        rz(originator),
			expParties: pdz(pd(originator, false, false, addr, nil)),
		},
		{
			name:       "unusable, not used, right role, only signer acc",
			parties:    pdz(pd(originator, false, false, "", addrAcc)),
			reqRoles:   rz(originator),
			exp:        rz(originator),
			expParties: pdz(pd(originator, false, false, "", addrAcc)),
		},
		{
			name:       "unusable, not used, right role, no signer",
			parties:    pdz(pd(originator, false, false, "", nil)),
			reqRoles:   rz(originator),
			exp:        rz(originator),
			expParties: pdz(pd(originator, false, false, "", nil)),
		},
		{
			name:       "unusable, not used, wrong role, both signer string and acc",
			parties:    pdz(pd(originator, false, false, addr, addrAcc)),
			reqRoles:   rz(servicer),
			exp:        rz(servicer),
			expParties: pdz(pd(originator, false, false, addr, addrAcc)),
		},
		{
			name:       "unusable, not used, wrong role, only signer string",
			parties:    pdz(pd(originator, false, false, addr, nil)),
			reqRoles:   rz(servicer),
			exp:        rz(servicer),
			expParties: pdz(pd(originator, false, false, addr, nil)),
		},
		{
			name:       "unusable, not used, wrong role, only signer acc",
			parties:    pdz(pd(originator, false, false, "", addrAcc)),
			reqRoles:   rz(servicer),
			exp:        rz(servicer),
			expParties: pdz(pd(originator, false, false, "", addrAcc)),
		},
		{
			name:       "unusable, not used, wrong role, no signer",
			parties:    pdz(pd(originator, false, false, "", nil)),
			reqRoles:   rz(servicer),
			exp:        rz(servicer),
			expParties: pdz(pd(originator, false, false, "", nil)),
		},
		{
			name:       "unusable, already used, right role, both signer string and acc",
			parties:    pdz(pd(investor, false, true, addr, addrAcc)),
			reqRoles:   rz(investor),
			exp:        rz(investor),
			expParties: pdz(pd(investor, false, true, addr, addrAcc)),
		},
		{
			name:       "unusable, already used, right role, only signer string",
			parties:    pdz(pd(investor, false, true, addr, nil)),
			reqRoles:   rz(investor),
			exp:        rz(investor),
			expParties: pdz(pd(investor, false, true, addr, nil)),
		},
		{
			name:       "unusable, already used, right role, only signer acc",
			parties:    pdz(pd(investor, false, true, "", addrAcc)),
			reqRoles:   rz(investor),
			exp:        rz(investor),
			expParties: pdz(pd(investor, false, true, "", addrAcc)),
		},
		{
			name:       "unusable, already used, right role, no signer",
			parties:    pdz(pd(investor, false, true, "", nil)),
			reqRoles:   rz(investor),
			exp:        rz(investor),
			expParties: pdz(pd(investor, false, true, "", nil)),
		},
		{
			name:       "unusable, already used, wrong role, both signer string and acc",
			parties:    pdz(pd(investor, false, true, addr, addrAcc)),
			reqRoles:   rz(omnibus),
			exp:        rz(omnibus),
			expParties: pdz(pd(investor, false, true, addr, addrAcc)),
		},
		{
			name:       "unusable, already used, wrong role, only signer string",
			parties:    pdz(pd(investor, false, true, addr, nil)),
			reqRoles:   rz(omnibus),
			exp:        rz(omnibus),
			expParties: pdz(pd(investor, false, true, addr, nil)),
		},
		{
			name:       "unusable, already used, wrong role, only signer acc",
			parties:    pdz(pd(investor, false, true, "", addrAcc)),
			reqRoles:   rz(omnibus),
			exp:        rz(omnibus),
			expParties: pdz(pd(investor, false, true, "", addrAcc)),
		},
		{
			name:       "unusable, already used, wrong role, no signer",
			parties:    pdz(pd(investor, false, true, "", nil)),
			reqRoles:   rz(omnibus),
			exp:        rz(omnibus),
			expParties: pdz(pd(investor, false, true, "", nil)),
		},
	}

	// make sure each role can be associated when there's only a singer string.
	for _, role := range allRoles {
		tests = append(tests, testCase{
			name:       fmt.Sprintf("%s can be associated signer string", strings.ToLower(role.SimpleString())),
			parties:    pdz(pd(role, true, false, addr, nil)),
			reqRoles:   rz(role),
			exp:        nil,
			expParties: pdz(pd(role, true, true, addr, nil)),
		})
	}
	// make sure each role can be associated when there's only a singer acc.
	for _, role := range allRoles {
		tests = append(tests, testCase{
			name:       fmt.Sprintf("%s can be associated signer acc", strings.ToLower(role.SimpleString())),
			parties:    pdz(pd(role, true, false, "", addrAcc)),
			reqRoles:   rz(role),
			exp:        nil,
			expParties: pdz(pd(role, true, true, "", addrAcc)),
		})
	}
	// make sure all roles can come up missing.
	for _, role := range allRoles {
		tests = append(tests, testCase{
			name:       fmt.Sprintf("%s can be be missing", strings.ToLower(role.SimpleString())),
			parties:    nil,
			reqRoles:   rz(role),
			exp:        rz(role),
			expParties: nil,
		})
	}

	tests = append(tests, []testCase{
		{
			name:       "missing ordered by req",
			parties:    pdz(pd(validator, true, false, addr, nil)),
			reqRoles:   rz(validator, owner, validator, originator, owner),
			exp:        rz(owner, validator, originator, owner),
			expParties: pdz(pd(validator, true, true, addr, nil)),
		},
		{
			name: "3 parties 2 req",
			parties: pdz(
				pd(validator, true, false, addr, nil),
				pd(validator, true, false, addr, nil),
				pd(validator, true, false, addr, nil),
			),
			reqRoles: rz(validator, validator),
			exp:      nil,
			expParties: pdz(
				pd(validator, true, true, addr, nil),
				pd(validator, true, true, addr, nil),
				pd(validator, true, false, addr, nil),
			),
		},
		{
			name: "3 parties diff roles all 3 req",
			parties: pdz(
				pd(servicer, true, false, addr, addrAcc),
				pd(owner, true, false, addr, addrAcc),
				pd(validator, true, false, addr, addrAcc),
			),
			reqRoles: rz(validator, servicer, owner),
			exp:      nil,
			expParties: pdz(
				pd(servicer, true, true, addr, addrAcc),
				pd(owner, true, true, addr, addrAcc),
				pd(validator, true, true, addr, addrAcc),
			),
		},
		{
			name: "3 parties diff roles 4 req only 1 filled",
			parties: pdz(
				pd(servicer, true, false, addr, addrAcc),
				pd(owner, true, false, addr, addrAcc),
				pd(validator, true, false, addr, addrAcc),
			),
			reqRoles: rz(originator, affiliate, custodian, owner),
			exp:      rz(originator, affiliate, custodian),
			expParties: pdz(
				pd(servicer, true, false, addr, addrAcc),
				pd(owner, true, true, addr, addrAcc),
				pd(validator, true, false, addr, addrAcc),
			),
		},
		{
			name: "one of each req all there",
			parties: pdz(
				pd(unspecified, true, false, addr, nil),
				pd(originator, true, false, addr, nil),
				pd(servicer, true, false, addr, nil),
				pd(investor, true, false, addr, nil),
				pd(custodian, true, false, addr, nil),
				pd(owner, true, false, addr, nil),
				pd(affiliate, true, false, addr, nil),
				pd(omnibus, true, false, addr, nil),
				pd(provenance, true, false, addr, nil),
				pd(controller, true, false, addr, nil),
				pd(validator, true, false, addr, nil),
			),
			reqRoles: allRoles,
			exp:      nil,
			expParties: pdz(
				pd(unspecified, true, true, addr, nil),
				pd(originator, true, true, addr, nil),
				pd(servicer, true, true, addr, nil),
				pd(investor, true, true, addr, nil),
				pd(custodian, true, true, addr, nil),
				pd(owner, true, true, addr, nil),
				pd(affiliate, true, true, addr, nil),
				pd(omnibus, true, true, addr, nil),
				pd(provenance, true, true, addr, nil),
				pd(controller, true, true, addr, nil),
				pd(validator, true, true, addr, nil),
			),
		},
		{
			name:       "unknown role can be fulfilled",
			parties:    pdz(pd(333, true, false, addr, addrAcc)),
			reqRoles:   rz(333),
			exp:        nil,
			expParties: pdz(pd(333, true, true, addr, addrAcc)),
		},
		{
			name:       "unknown role can be missed",
			parties:    nil,
			reqRoles:   rz(333),
			exp:        rz(333),
			expParties: nil,
		},
	}...)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := keeper.AssociateRequiredRoles(tc.parties, tc.reqRoles)
			assert.Equal(t, tc.exp, actual, "associateRequiredRoles returned roles")
			assert.Equal(t, tc.expParties, tc.parties, "parties after associateRequiredRoles")
		})
	}
}

func TestMissingRolesString(t *testing.T) {
	// pd is a short way to create a PartyDetails with only what we care about in this test.
	pd := func(role types.PartyType, used bool) *keeper.PartyDetails {
		return keeper.TestablePartyDetails{
			Role:       role,
			UsedBySpec: used,
		}.Real()
	}
	// pdz is just a shorter way to define a []*keeper.PartyDetails
	pdz := func(parties ...*keeper.PartyDetails) []*keeper.PartyDetails {
		rv := make([]*keeper.PartyDetails, 0, len(parties))
		rv = append(rv, parties...)
		return rv
	}
	// rv is just a shorter way to define a []types.PartyType
	rz := func(roles ...types.PartyType) []types.PartyType {
		rv := make([]types.PartyType, 0, len(roles))
		rv = append(rv, roles...)
		return rv
	}
	// rolesCopy returns a copy of the provided slice. Nil in = nil out.
	rolesCopy := func(roles []types.PartyType) []types.PartyType {
		if roles == nil {
			return nil
		}
		return rz(roles...)
	}

	// Create some aliases that are shorter than their full names.
	unspecified := types.PartyType_PARTY_TYPE_UNSPECIFIED
	originator := types.PartyType_PARTY_TYPE_ORIGINATOR
	servicer := types.PartyType_PARTY_TYPE_SERVICER
	investor := types.PartyType_PARTY_TYPE_INVESTOR
	custodian := types.PartyType_PARTY_TYPE_CUSTODIAN
	owner := types.PartyType_PARTY_TYPE_OWNER
	affiliate := types.PartyType_PARTY_TYPE_AFFILIATE
	omnibus := types.PartyType_PARTY_TYPE_OMNIBUS
	provenance := types.PartyType_PARTY_TYPE_PROVENANCE
	controller := types.PartyType_PARTY_TYPE_CONTROLLER
	validator := types.PartyType_PARTY_TYPE_VALIDATOR

	// rolesForDeterministismTests returns two of each PartyType in a random order.
	rolesForDeterministismTests := func() []types.PartyType {
		rv := make([]types.PartyType, 0, 2*len(types.PartyType_name))
		for i := range types.PartyType_name {
			rv = append(rv, types.PartyType(i), types.PartyType(i))
		}
		rand.Shuffle(len(rv), func(i, j int) {
			rv[i], rv[j] = rv[j], rv[i]
		})
		return rv
	}
	// partiesForDeterministismTests returns two parties for each role (1 used, 1 not) in a random order.
	partiesForDeterministismTests := func() []*keeper.PartyDetails {
		var rv []*keeper.PartyDetails
		for i := range types.PartyType_name {
			role := types.PartyType(i)
			rv = append(rv, pd(role, true), pd(role, false))
		}
		rand.Shuffle(len(rv), func(i, j int) {
			rv[i], rv[j] = rv[j], rv[i]
		})
		return rv
	}
	// resultForDeterministismTests is the expected result for all the determinism tests.
	resultForDeterministismTests := "UNSPECIFIED need 2 have 1, ORIGINATOR need 2 have 1, SERVICER need 2 have 1, " +
		"INVESTOR need 2 have 1, CUSTODIAN need 2 have 1, OWNER need 2 have 1, AFFILIATE need 2 have 1, " +
		"OMNIBUS need 2 have 1, PROVENANCE need 2 have 1, CONTROLLER need 2 have 1, VALIDATOR need 2 have 1"

	// roleStr gets a string of the variable name (or value) used in these tests for the roles.
	roleStr := func(role types.PartyType) string {
		return strings.ToLower(role.SimpleString())
	}
	rolesStr := func(roles []types.PartyType) string {
		if roles == nil {
			return "nil"
		}
		strs := make([]string, len(roles))
		for i, role := range roles {
			strs[i] = roleStr(role)
		}
		return fmt.Sprintf("rz(%s)", strings.Join(strs, ", "))
	}
	// partyStr gets a string of the golang code that would make the provided party for these tests.
	partyStr := func(party *keeper.PartyDetails) string {
		return fmt.Sprintf("pd(%s, %t)", roleStr(party.GetRole()), party.IsUsed())
	}
	// partiesStr gets a string of the golang code that would make the provided parties for these tests.
	partiesStr := func(parties []*keeper.PartyDetails) string {
		if parties == nil {
			return "nil"
		}
		strs := make([]string, len(parties))
		for i, party := range parties {
			strs[i] = partyStr(party)
		}
		if len(strs) <= 4 {
			return fmt.Sprintf("pdz(%s)", strings.Join(strs, ", "))
		}
		return fmt.Sprintf("pdz(\n\t\t%s,\n\t)", strings.Join(strs, ",\n\t\t"))
	}

	// rolesReversed copies the roles slice, reversing it. Nil in = nil out.
	rolesReversed := func(roles []types.PartyType) []types.PartyType {
		if roles == nil {
			return nil
		}
		rv := make([]types.PartyType, len(roles))
		for i, role := range roles {
			rv[len(rv)-i-1] = role
		}
		return rv
	}
	// rolesShuffled copies the roles slice and shuffles the entries. Nil in = nil out.
	rolesShuffled := func(r *rand.Rand, roles []types.PartyType) []types.PartyType {
		if roles == nil {
			return nil
		}
		rv := make([]types.PartyType, 0, len(roles))
		rv = append(rv, roles...)
		r.Shuffle(len(rv), func(i, j int) {
			rv[i], rv[j] = rv[j], rv[i]
		})
		return rv
	}
	// partiesShuffled copies each of the provided party and returns them in a random order. Nil in = nil out.
	partiesShuffled := func(r *rand.Rand, parties []*keeper.PartyDetails) []*keeper.PartyDetails {
		if parties == nil {
			return nil
		}
		rv := make([]*keeper.PartyDetails, 0, len(parties))
		rv = append(rv, parties...)
		r.Shuffle(len(rv), func(i, j int) {
			rv[i], rv[j] = rv[j], rv[i]
		})
		return rv
	}

	type testCase struct {
		name     string
		parties  []*keeper.PartyDetails
		reqRoles []types.PartyType
		exp      string
	}

	tests := []testCase{
		// Negative tests for each role.
		{
			name:     "nil parties 2 required unspecified",
			parties:  nil,
			reqRoles: rz(unspecified, unspecified),
			exp:      "UNSPECIFIED need 2 have 0",
		},
		{
			name:     "nil parties 2 required originator",
			parties:  nil,
			reqRoles: rz(originator, originator),
			exp:      "ORIGINATOR need 2 have 0",
		},
		{
			name:     "nil parties 2 required servicer",
			parties:  nil,
			reqRoles: rz(servicer, servicer),
			exp:      "SERVICER need 2 have 0",
		},
		{
			name:     "nil parties 2 required investor",
			parties:  nil,
			reqRoles: rz(investor, investor),
			exp:      "INVESTOR need 2 have 0",
		},
		{
			name:     "nil parties 2 required custodian",
			parties:  nil,
			reqRoles: rz(custodian, custodian),
			exp:      "CUSTODIAN need 2 have 0",
		},
		{
			name:     "nil parties 2 required owner",
			parties:  nil,
			reqRoles: rz(owner, owner),
			exp:      "OWNER need 2 have 0",
		},
		{
			name:     "nil parties 2 required affiliate",
			parties:  nil,
			reqRoles: rz(affiliate, affiliate),
			exp:      "AFFILIATE need 2 have 0",
		},
		{
			name:     "nil parties 2 required omnibus",
			parties:  nil,
			reqRoles: rz(omnibus, omnibus),
			exp:      "OMNIBUS need 2 have 0",
		},
		{
			name:     "nil parties 2 required provenance",
			parties:  nil,
			reqRoles: rz(provenance, provenance),
			exp:      "PROVENANCE need 2 have 0",
		},
		{
			name:     "nil parties 2 required controller",
			parties:  nil,
			reqRoles: rz(controller, controller),
			exp:      "CONTROLLER need 2 have 0",
		},
		{
			name:     "nil parties 2 required validator",
			parties:  nil,
			reqRoles: rz(validator, validator),
			exp:      "VALIDATOR need 2 have 0",
		},

		// Positive tests for each role
		{
			name:     "2 required unspecified satisfied",
			parties:  pdz(pd(unspecified, true), pd(unspecified, true), pd(unspecified, true)),
			reqRoles: rz(unspecified, unspecified),
			exp:      "",
		},
		{
			name:     "2 required originator satisfied",
			parties:  pdz(pd(originator, true), pd(originator, true), pd(originator, true)),
			reqRoles: rz(originator, originator),
			exp:      "",
		},
		{
			name:     "2 required servicer satisfied",
			parties:  pdz(pd(servicer, true), pd(servicer, true), pd(servicer, true)),
			reqRoles: rz(servicer, servicer),
			exp:      "",
		},
		{
			name:     "2 required investor satisfied",
			parties:  pdz(pd(investor, true), pd(investor, true), pd(investor, true)),
			reqRoles: rz(investor, investor),
			exp:      "",
		},
		{
			name:     "2 required custodian satisfied",
			parties:  pdz(pd(custodian, true), pd(custodian, true), pd(custodian, true)),
			reqRoles: rz(custodian, custodian),
			exp:      "",
		},
		{
			name:     "2 required owner satisfied",
			parties:  pdz(pd(owner, true), pd(owner, true), pd(owner, true)),
			reqRoles: rz(owner, owner),
			exp:      "",
		},
		{
			name:     "2 required affiliate satisfied",
			parties:  pdz(pd(affiliate, true), pd(affiliate, true), pd(affiliate, true)),
			reqRoles: rz(affiliate, affiliate),
			exp:      "",
		},
		{
			name:     "2 required omnibus satisfied",
			parties:  pdz(pd(omnibus, true), pd(omnibus, true), pd(omnibus, true)),
			reqRoles: rz(omnibus, omnibus),
			exp:      "",
		},
		{
			name:     "2 required provenance satisfied",
			parties:  pdz(pd(provenance, true), pd(provenance, true), pd(provenance, true)),
			reqRoles: rz(provenance, provenance),
			exp:      "",
		},
		{
			name:     "2 required controller satisfied",
			parties:  pdz(pd(controller, true), pd(controller, true), pd(controller, true)),
			reqRoles: rz(controller, controller),
			exp:      "",
		},
		{
			name:     "2 required validator satisfied",
			parties:  pdz(pd(validator, true), pd(validator, true), pd(validator, true)),
			reqRoles: rz(validator, validator),
			exp:      "",
		},

		// nil/empty handling tests
		{
			name:     "nil nil",
			parties:  nil,
			reqRoles: nil,
			exp:      "",
		},
		{
			name:     "empty nil",
			parties:  pdz(),
			reqRoles: nil,
			exp:      "",
		},
		{
			name:     "nil empty",
			parties:  nil,
			reqRoles: rz(),
			exp:      "",
		},
		{
			name:     "empty empty",
			parties:  pdz(),
			reqRoles: rz(),
			exp:      "",
		},

		// unknown value tests
		{
			name:     "unknown role twice no such parties",
			parties:  pdz(pd(servicer, true), pd(owner, true)),
			reqRoles: rz(owner, 100, servicer, 100),
			exp:      "100 need 2 have 0",
		},
		{
			name:     "unknown role twice 1 such party unused",
			parties:  pdz(pd(100, false), pd(servicer, true), pd(owner, true)),
			reqRoles: rz(owner, 100, servicer, 100),
			exp:      "100 need 2 have 0",
		},
		{
			name:     "unknown role twice 1 such party used",
			parties:  pdz(pd(100, true), pd(servicer, true), pd(owner, true)),
			reqRoles: rz(owner, 100, servicer, 100),
			exp:      "100 need 2 have 1",
		},
		{
			name:     "unknown role twice 2 such parties both unused",
			parties:  pdz(pd(100, false), pd(servicer, true), pd(owner, true), pd(100, false)),
			reqRoles: rz(owner, 100, servicer, 100),
			exp:      "100 need 2 have 0",
		},
		{
			name:     "unknown role twice 2 such parties first unused",
			parties:  pdz(pd(100, false), pd(servicer, true), pd(owner, true), pd(100, true)),
			reqRoles: rz(owner, 100, servicer, 100),
			exp:      "100 need 2 have 1",
		},
		{
			name:     "unknown role twice 2 such parties second unused",
			parties:  pdz(pd(100, true), pd(servicer, true), pd(owner, true), pd(100, false)),
			reqRoles: rz(owner, 100, servicer, 100),
			exp:      "100 need 2 have 1",
		},
		{
			name:     "unknown role twice 2 such parties both used",
			parties:  pdz(pd(100, true), pd(servicer, true), pd(owner, true), pd(100, true)),
			reqRoles: rz(owner, 100, servicer, 100),
			exp:      "",
		},
		{
			name:     "parties with unknown roles",
			parties:  pdz(pd(-55, true), pd(-56, false), pd(9, true), pd(57, true), pd(58, false)),
			reqRoles: rz(owner),
			exp:      "OWNER need 1 have 0",
		},

		// complex tests
		{
			name:     "2 same req have 2 both unused",
			parties:  pdz(pd(owner, false), pd(owner, false)),
			reqRoles: rz(owner, owner),
			exp:      "OWNER need 2 have 0",
		},
		{
			name:     "2 same req have 2 first unused",
			parties:  pdz(pd(owner, false), pd(owner, true)),
			reqRoles: rz(owner, owner),
			exp:      "OWNER need 2 have 1",
		},
		{
			name:     "2 same req have 2 second unused",
			parties:  pdz(pd(owner, true), pd(owner, false)),
			reqRoles: rz(owner, owner),
			exp:      "OWNER need 2 have 1",
		},
		{
			name:     "2 same req have 2 both used",
			parties:  pdz(pd(owner, true), pd(owner, true)),
			reqRoles: rz(owner, owner),
			exp:      "",
		},
		{
			name:     "2 diff req have 2 both unused",
			parties:  pdz(pd(servicer, false), pd(investor, false)),
			reqRoles: rz(servicer, investor),
			exp:      "SERVICER need 1 have 0, INVESTOR need 1 have 0",
		},
		{
			name:     "2 diff req have 2 first unused",
			parties:  pdz(pd(servicer, false), pd(investor, true)),
			reqRoles: rz(servicer, investor),
			exp:      "SERVICER need 1 have 0",
		},
		{
			name:     "2 diff req have 2 second unused",
			parties:  pdz(pd(servicer, true), pd(investor, false)),
			reqRoles: rz(servicer, investor),
			exp:      "INVESTOR need 1 have 0",
		},
		{
			name:     "2 diff req have 2 both used",
			parties:  pdz(pd(servicer, true), pd(investor, true)),
			reqRoles: rz(servicer, investor),
			exp:      "",
		},
		{
			name: "4 req none 7 used parties of other roles plus 1 unused of a req role",
			parties: pdz(
				pd(unspecified, true),
				pd(servicer, true),
				pd(investor, true),
				pd(owner, true),
				pd(affiliate, true),
				pd(omnibus, false),
				pd(controller, true),
				pd(validator, true),
			),
			reqRoles: rz(originator, custodian, omnibus, provenance),
			exp:      "ORIGINATOR need 1 have 0, CUSTODIAN need 1 have 0, OMNIBUS need 1 have 0, PROVENANCE need 1 have 0",
		},
		{
			// For this one, 3 different role types required in amounts of 3, 2, and 1 (6 total required roles).
			// The one with 3 will have 2 used and 1 unused.
			// The one with 2 will have 3 unused.
			// The one with 1 will have 2 used and 1 unused.
			// There will also be parties a 4th role with 1 used and 1 unused.
			name: "10 parties 6 req not all fulfilled",
			parties: pdz(
				pd(custodian, true),
				pd(custodian, true),
				pd(custodian, false),
				pd(owner, false),
				pd(owner, false),
				pd(owner, false),
				pd(controller, true),
				pd(controller, false),
				pd(validator, true),
				pd(validator, true),
				pd(validator, false),
			),
			reqRoles: rz(custodian, custodian, custodian, owner, owner, validator),
			exp:      "CUSTODIAN need 3 have 2, OWNER need 2 have 0",
		},

		// result determinism tests.
		// Three tests that look the same, but end up having different orderings
		// for parties and reqRoles. Three times should be enough to get a nice
		// spread of orderings for the two inputs that sufficiently demonstrates
		// that the result is consistent and deterministic.
		// If a new PartyType is added, these should fail. If that happens, update
		// the resultForDeterministismTests to include the new type.
		{
			name:     "deterministic ordering 1",
			parties:  partiesForDeterministismTests(),
			reqRoles: rolesForDeterministismTests(),
			exp:      resultForDeterministismTests,
		},
		{
			name:     "deterministic ordering 2",
			parties:  partiesForDeterministismTests(),
			reqRoles: rolesForDeterministismTests(),
			exp:      resultForDeterministismTests,
		},
		{
			name:     "deterministic ordering 3",
			parties:  partiesForDeterministismTests(),
			reqRoles: rolesForDeterministismTests(),
			exp:      resultForDeterministismTests,
		},
	}

	// Add three result determinism tests.
	// They all technically have the same inputs and expected result but the input
	// orderings are different for each. Three tests combined with the four extra
	// test variations (added below these) should sufficiently demonstrate that
	// the result is consistent and deterministic.
	// If a new PartyType is added, these should fail. If that happens, update
	// the resultForDeterministismTests to include the new type.
	for i := 1; i <= 3; i++ {
		tests = append(tests, testCase{
			name:     fmt.Sprintf("deterministic ordering %d", i),
			parties:  partiesForDeterministismTests(),
			reqRoles: rolesForDeterministismTests(),
			exp:      resultForDeterministismTests,
		})
	}

	// Copy all tests four times.
	// Once with reversed parties. Once with reversed req roles.
	// Once with both reversed. And once with both shuffled.
	revPartiesTests := make([]testCase, len(tests))
	revRolesTests := make([]testCase, len(tests))
	revBothTests := make([]testCase, len(tests))
	shuffledTests := make([]testCase, len(tests))

	for i, tc := range tests {
		revPartiesTests[i] = testCase{
			name:     "rev parties " + tc.name,
			parties:  partiesReversed(tc.parties),
			reqRoles: rolesCopy(tc.reqRoles),
			exp:      tc.exp,
		}
		revRolesTests[i] = testCase{
			name:     "rev roles " + tc.name,
			parties:  partiesCopy(tc.parties),
			reqRoles: rolesReversed(tc.reqRoles),
			exp:      tc.exp,
		}
		revBothTests[i] = testCase{
			name:     "rev both " + tc.name,
			parties:  partiesReversed(tc.parties),
			reqRoles: rolesReversed(tc.reqRoles),
			exp:      tc.exp,
		}
		// Using a hard-coded (randomly chosen) seed value here to make life easier if one of these fails.
		// The purpose is to just have them in an order other than as defined (hopefully).
		r := rand.New(rand.NewSource(int64(86530 * i)))
		shuffledTests[i] = testCase{
			name:     "shuffled " + tc.name,
			parties:  partiesShuffled(r, tc.parties),
			reqRoles: rolesShuffled(r, tc.reqRoles),
			exp:      tc.exp,
		}
	}
	tests = append(tests, revPartiesTests...)
	tests = append(tests, revRolesTests...)
	tests = append(tests, revBothTests...)
	tests = append(tests, shuffledTests...)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := keeper.MissingRolesString(tc.parties, tc.reqRoles)
			if !assert.Equal(t, tc.exp, actual, "missingRolesString") {
				// The test failed. The expected and actual are in the output.
				// Now, be helpful and output the inputs too.
				t.Logf("tests = append(tests, {\n\tname: %q,\n\tparties: %s,\n\treqRoles: %s,\n\texp: %q\n})",
					tc.name, partiesStr(tc.parties), rolesStr(tc.reqRoles), tc.exp)
			}
		})
	}
}

func TestGetAuthzMessageTypeURLs(t *testing.T) {
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
		t.Run(getName(tc), func(t *testing.T) {
			actual := keeper.GetAuthzMessageTypeURLs(tc.url)
			assert.Equal(t, tc.expected, actual, "getAuthzMessageTypeURLs(%q)", tc.url)
		})
	}
}

func (s *AuthzTestSuite) TestFindAuthzGrantee() {
	acc := func(addr string) sdk.AccAddress {
		return sdk.AccAddress(addr)
	}
	accz := func(addrs ...string) []sdk.AccAddress {
		rv := make([]sdk.AccAddress, len(addrs))
		for i, addr := range addrs {
			rv[i] = acc(addr)
		}
		return rv
	}
	gi := func(grantee, granter, msgType string) GrantInfo {
		return GrantInfo{
			Grantee: acc(grantee),
			Granter: acc(granter),
			MsgType: msgType,
		}
	}
	newErr := func(msg string) error {
		if len(msg) == 0 {
			return nil
		}
		return errors.New(msg)
	}

	normalMsg := &types.MsgWriteScopeRequest{}
	normalMsgType := types.TypeURLMsgWriteScopeRequest
	specialMsg := &types.MsgAddScopeDataAccessRequest{}
	specialMsgType1 := types.TypeURLMsgAddScopeDataAccessRequest
	specialMsgType2 := types.TypeURLMsgWriteScopeRequest

	sometimeVal := time.Unix(1234567, 0)
	sometime := &sometimeVal

	tests := []struct {
		name         string
		granter      sdk.AccAddress
		grantees     []sdk.AccAddress
		msg          types.MetadataMsg
		authzKeeper  *MockAuthzKeeper
		expGrantee   sdk.AccAddress
		expErr       string
		expGetAuth   []*GetAuthorizationCall
		expDelGrant  []*DeleteGrantCall
		expSaveGrant []*SaveGrantCall
	}{
		{
			name:        "nil granter",
			granter:     nil,
			grantees:    accz("grantee_________addr"),
			msg:         normalMsg,
			authzKeeper: NewMockAuthzKeeper(),
			expGrantee:  nil,
			expErr:      "",
		},
		{
			name:        "empty granter",
			granter:     sdk.AccAddress{},
			grantees:    accz("grantee_________addr"),
			msg:         normalMsg,
			authzKeeper: NewMockAuthzKeeper(),
			expGrantee:  nil,
			expErr:      "",
		},
		{
			name:        "nil grantees",
			granter:     acc("granter_addr________"),
			grantees:    nil,
			msg:         normalMsg,
			authzKeeper: NewMockAuthzKeeper(),
			expGrantee:  nil,
			expErr:      "",
		},
		{
			name:        "empty grantees",
			granter:     acc("granter_addr________"),
			grantees:    accz(),
			msg:         normalMsg,
			authzKeeper: NewMockAuthzKeeper(),
			expGrantee:  nil,
			expErr:      "",
		},
		{
			name:        "one grantee no auth",
			granter:     acc("granter_addr________"),
			grantees:    accz("grantee_________addr"),
			msg:         normalMsg,
			authzKeeper: NewMockAuthzKeeper(),
			expGrantee:  nil,
			expErr:      "",
			expGetAuth: []*GetAuthorizationCall{{
				GrantInfo: gi("grantee_________addr", "granter_addr________", normalMsgType),
				Result:    GetAuthorizationResult{Auth: nil, Exp: nil},
			}},
		},
		{
			name:        "one grantee no auth special msg type",
			granter:     acc("granter_addr________"),
			grantees:    accz("grantee_________addr"),
			msg:         specialMsg,
			authzKeeper: NewMockAuthzKeeper(),
			expGrantee:  nil,
			expErr:      "",
			expGetAuth: []*GetAuthorizationCall{
				{
					GrantInfo: gi("grantee_________addr", "granter_addr________", specialMsgType1),
					Result:    GetAuthorizationResult{Auth: nil, Exp: nil},
				},
				{
					GrantInfo: gi("grantee_________addr", "granter_addr________", specialMsgType2),
					Result:    GetAuthorizationResult{Auth: nil, Exp: nil},
				},
			},
		},
		{
			name:        "two grantees no auths",
			granter:     acc("granter_addr________"),
			grantees:    accz("grantee_1_______addr", "grantee_2_______addr"),
			msg:         normalMsg,
			authzKeeper: NewMockAuthzKeeper(),
			expGrantee:  nil,
			expErr:      "",
			expGetAuth: []*GetAuthorizationCall{
				{
					GrantInfo: gi("grantee_1_______addr", "granter_addr________", normalMsgType),
					Result:    GetAuthorizationResult{Auth: nil, Exp: nil},
				},
				{
					GrantInfo: gi("grantee_2_______addr", "granter_addr________", normalMsgType),
					Result:    GetAuthorizationResult{Auth: nil, Exp: nil},
				},
			},
		},
		{
			name:        "two grantees no auths special msg type",
			granter:     acc("granter_addr________"),
			grantees:    accz("grantee_1_______addr", "grantee_2_______addr"),
			msg:         specialMsg,
			authzKeeper: NewMockAuthzKeeper(),
			expGrantee:  nil,
			expErr:      "",
			expGetAuth: []*GetAuthorizationCall{
				{
					GrantInfo: gi("grantee_1_______addr", "granter_addr________", specialMsgType1),
					Result:    GetAuthorizationResult{Auth: nil, Exp: nil},
				},
				{
					GrantInfo: gi("grantee_1_______addr", "granter_addr________", specialMsgType2),
					Result:    GetAuthorizationResult{Auth: nil, Exp: nil},
				},
				{
					GrantInfo: gi("grantee_2_______addr", "granter_addr________", specialMsgType1),
					Result:    GetAuthorizationResult{Auth: nil, Exp: nil},
				},
				{
					GrantInfo: gi("grantee_2_______addr", "granter_addr________", specialMsgType2),
					Result:    GetAuthorizationResult{Auth: nil, Exp: nil},
				},
			},
		},
		{
			name:     "two grantees first with acceptable auth",
			granter:  acc("granter_addr________"),
			grantees: accz("grantee_1_______addr", "grantee_2_______addr"),
			msg:      normalMsg,
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(
				GetAuthorizationCall{
					GrantInfo: gi("grantee_1_______addr", "granter_addr________", normalMsgType),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one", authz.AcceptResponse{Accept: true}, nil),
						Exp:  nil},
				},
			),
			expGrantee: acc("grantee_1_______addr"),
			expErr:     "",
			expGetAuth: []*GetAuthorizationCall{
				{
					GrantInfo: gi("grantee_1_______addr", "granter_addr________", normalMsgType),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one", authz.AcceptResponse{Accept: true}, nil).WithAcceptCalls(normalMsg),
						Exp:  nil},
				},
			},
		},
		{
			name:     "two grantees second with acceptable auth",
			granter:  acc("granter_addr________"),
			grantees: accz("grantee_1_______addr", "grantee_2_______addr"),
			msg:      normalMsg,
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(
				GetAuthorizationCall{
					GrantInfo: gi("grantee_2_______addr", "granter_addr________", normalMsgType),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("two", authz.AcceptResponse{Accept: true}, nil),
						Exp:  nil},
				},
			),
			expGrantee: acc("grantee_2_______addr"),
			expErr:     "",
			expGetAuth: []*GetAuthorizationCall{
				{
					GrantInfo: gi("grantee_1_______addr", "granter_addr________", normalMsgType),
					Result:    GetAuthorizationResult{Auth: nil, Exp: nil},
				},
				{
					GrantInfo: gi("grantee_2_______addr", "granter_addr________", normalMsgType),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("two", authz.AcceptResponse{Accept: true}, nil).WithAcceptCalls(normalMsg),
						Exp:  nil},
				},
			},
		},
		{
			name:     "two grantees special msg first with acceptable auth on first type",
			granter:  acc("granter_addr________"),
			grantees: accz("grantee_1_______addr", "grantee_2_______addr"),
			msg:      specialMsg,
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(
				GetAuthorizationCall{
					GrantInfo: gi("grantee_1_______addr", "granter_addr________", specialMsgType1),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one", authz.AcceptResponse{Accept: true}, nil),
						Exp:  nil},
				},
			),
			expGrantee: acc("grantee_1_______addr"),
			expErr:     "",
			expGetAuth: []*GetAuthorizationCall{
				{
					GrantInfo: gi("grantee_1_______addr", "granter_addr________", specialMsgType1),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one", authz.AcceptResponse{Accept: true}, nil).WithAcceptCalls(specialMsg),
						Exp:  nil},
				},
			},
		},
		{
			name:     "two grantees special msg first with acceptable auth on second type",
			granter:  acc("granter_addr________"),
			grantees: accz("grantee_1_______addr", "grantee_2_______addr"),
			msg:      specialMsg,
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(
				GetAuthorizationCall{
					GrantInfo: gi("grantee_1_______addr", "granter_addr________", specialMsgType2),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one", authz.AcceptResponse{Accept: true}, nil),
						Exp:  nil},
				},
			),
			expGrantee: acc("grantee_1_______addr"),
			expErr:     "",
			expGetAuth: []*GetAuthorizationCall{
				{
					GrantInfo: gi("grantee_1_______addr", "granter_addr________", specialMsgType1),
					Result:    GetAuthorizationResult{Auth: nil, Exp: nil},
				},
				{
					GrantInfo: gi("grantee_1_______addr", "granter_addr________", specialMsgType2),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one", authz.AcceptResponse{Accept: true}, nil).WithAcceptCalls(specialMsg),
						Exp:  nil},
				},
			},
		},
		{
			name:     "two grantees special msg second with acceptable auth on first type",
			granter:  acc("granter_addr________"),
			grantees: accz("grantee_1_______addr", "grantee_2_______addr"),
			msg:      specialMsg,
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(
				GetAuthorizationCall{
					GrantInfo: gi("grantee_2_______addr", "granter_addr________", specialMsgType1),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one", authz.AcceptResponse{Accept: true}, nil),
						Exp:  nil},
				},
			),
			expGrantee: acc("grantee_2_______addr"),
			expErr:     "",
			expGetAuth: []*GetAuthorizationCall{
				{
					GrantInfo: gi("grantee_1_______addr", "granter_addr________", specialMsgType1),
					Result:    GetAuthorizationResult{Auth: nil, Exp: nil},
				},
				{
					GrantInfo: gi("grantee_1_______addr", "granter_addr________", specialMsgType2),
					Result:    GetAuthorizationResult{Auth: nil, Exp: nil},
				},
				{
					GrantInfo: gi("grantee_2_______addr", "granter_addr________", specialMsgType1),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one", authz.AcceptResponse{Accept: true}, nil).WithAcceptCalls(specialMsg),
						Exp:  nil},
				},
			},
		},
		{
			name:     "two grantees special msg second with acceptable auth on second type",
			granter:  acc("granter_addr________"),
			grantees: accz("grantee_1_______addr", "grantee_2_______addr"),
			msg:      specialMsg,
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(
				GetAuthorizationCall{
					GrantInfo: gi("grantee_2_______addr", "granter_addr________", specialMsgType2),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one", authz.AcceptResponse{Accept: true}, nil),
						Exp:  nil},
				},
			),
			expGrantee: acc("grantee_2_______addr"),
			expErr:     "",
			expGetAuth: []*GetAuthorizationCall{
				{
					GrantInfo: gi("grantee_1_______addr", "granter_addr________", specialMsgType1),
					Result:    GetAuthorizationResult{Auth: nil, Exp: nil},
				},
				{
					GrantInfo: gi("grantee_1_______addr", "granter_addr________", specialMsgType2),
					Result:    GetAuthorizationResult{Auth: nil, Exp: nil},
				},
				{
					GrantInfo: gi("grantee_2_______addr", "granter_addr________", specialMsgType1),
					Result:    GetAuthorizationResult{Auth: nil, Exp: nil},
				},
				{
					GrantInfo: gi("grantee_2_______addr", "granter_addr________", specialMsgType2),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one", authz.AcceptResponse{Accept: true}, nil).WithAcceptCalls(specialMsg),
						Exp:  nil},
				},
			},
		},
		{
			name:     "two grantees special message first get auth errors on accept second is acceptable",
			granter:  acc("granter_addr________"),
			grantees: accz("grantee_1_______addr", "grantee_2_______addr"),
			msg:      specialMsg,
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(
				GetAuthorizationCall{
					GrantInfo: gi("grantee_1_______addr", "granter_addr________", specialMsgType1),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one", authz.AcceptResponse{Accept: true}, newErr("error from one")),
						Exp:  nil},
				},
				GetAuthorizationCall{
					GrantInfo: gi("grantee_1_______addr", "granter_addr________", specialMsgType2),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("two", authz.AcceptResponse{Accept: true}, nil),
						Exp:  nil},
				},
			),
			expGrantee: acc("grantee_1_______addr"),
			expErr:     "",
			expGetAuth: []*GetAuthorizationCall{
				{
					GrantInfo: gi("grantee_1_______addr", "granter_addr________", specialMsgType1),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one", authz.AcceptResponse{Accept: true}, newErr("error from one")).WithAcceptCalls(specialMsg),
						Exp:  nil},
				},
				{
					GrantInfo: gi("grantee_1_______addr", "granter_addr________", specialMsgType2),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("two", authz.AcceptResponse{Accept: true}, nil).WithAcceptCalls(specialMsg),
						Exp:  nil},
				},
			},
		},
		{
			name:     "authorization should be deleted",
			granter:  acc("granter_addr________"),
			grantees: accz("grantee_________addr"),
			msg:      normalMsg,
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(
				GetAuthorizationCall{
					GrantInfo: gi("grantee_________addr", "granter_addr________", normalMsgType),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one", authz.AcceptResponse{Accept: true, Delete: true}, nil),
						Exp:  nil},
				},
			),
			expGrantee: acc("grantee_________addr"),
			expErr:     "",
			expGetAuth: []*GetAuthorizationCall{
				{
					GrantInfo: gi("grantee_________addr", "granter_addr________", normalMsgType),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one", authz.AcceptResponse{Accept: true, Delete: true}, nil).WithAcceptCalls(normalMsg),
						Exp:  nil},
				},
			},
			expDelGrant: []*DeleteGrantCall{
				{
					GrantInfo: gi("grantee_________addr", "granter_addr________", normalMsgType),
					Result:    nil,
				},
			},
		},
		{
			name:     "error deleting authorization",
			granter:  acc("granter_addr________"),
			grantees: accz("grantee_________addr"),
			msg:      normalMsg,
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(
				GetAuthorizationCall{
					GrantInfo: gi("grantee_________addr", "granter_addr________", normalMsgType),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one", authz.AcceptResponse{Accept: true, Delete: true}, nil),
						Exp:  nil},
				},
			).WithDeleteGrantResults(
				DeleteGrantCall{
					GrantInfo: gi("grantee_________addr", "granter_addr________", normalMsgType),
					Result:    newErr("test delete error"),
				},
			),
			expGrantee: nil,
			expErr:     "test delete error",
			expGetAuth: []*GetAuthorizationCall{
				{
					GrantInfo: gi("grantee_________addr", "granter_addr________", normalMsgType),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one", authz.AcceptResponse{Accept: true, Delete: true}, nil).WithAcceptCalls(normalMsg),
						Exp:  nil},
				},
			},
			expDelGrant: []*DeleteGrantCall{
				{
					GrantInfo: gi("grantee_________addr", "granter_addr________", normalMsgType),
					Result:    newErr("test delete error"),
				},
			},
		},
		{
			name:     "authorization should be saved",
			granter:  acc("granter_addr________"),
			grantees: accz("grantee_________addr"),
			msg:      normalMsg,
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(
				GetAuthorizationCall{
					GrantInfo: gi("grantee_________addr", "granter_addr________", normalMsgType),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one",
							authz.AcceptResponse{
								Accept:  true,
								Updated: NewMockAuthorization("two", authz.AcceptResponse{}, nil),
							}, nil),
						Exp: sometime},
				},
			),
			expGrantee: acc("grantee_________addr"),
			expErr:     "",
			expGetAuth: []*GetAuthorizationCall{
				{
					GrantInfo: gi("grantee_________addr", "granter_addr________", normalMsgType),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one",
							authz.AcceptResponse{
								Accept:  true,
								Updated: NewMockAuthorization("two", authz.AcceptResponse{}, nil),
							}, nil).WithAcceptCalls(normalMsg),
						Exp: sometime},
				},
			},
			expSaveGrant: []*SaveGrantCall{
				{
					Grantee: acc("grantee_________addr"),
					Granter: acc("granter_addr________"),
					Auth:    NewMockAuthorization("two", authz.AcceptResponse{}, nil),
					Exp:     sometime,
					Result:  nil,
				},
			},
		},
		{
			name:     "error saving authorization",
			granter:  acc("granter_addr________"),
			grantees: accz("grantee_________addr"),
			msg:      normalMsg,
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(
				GetAuthorizationCall{
					GrantInfo: gi("grantee_________addr", "granter_addr________", normalMsgType),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one",
							authz.AcceptResponse{
								Accept:  true,
								Updated: NewMockAuthorization("two", authz.AcceptResponse{}, nil),
							}, nil),
						Exp: sometime},
				},
			).WithSaveGrantResults(
				SaveGrantCall{
					Grantee: acc("grantee_________addr"),
					Granter: acc("granter_addr________"),
					Auth:    NewMockAuthorization("two", authz.AcceptResponse{}, nil),
					Exp:     sometime,
					Result:  newErr("test update error message"),
				},
			),
			expGrantee: nil,
			expErr:     "test update error message",
			expGetAuth: []*GetAuthorizationCall{
				{
					GrantInfo: gi("grantee_________addr", "granter_addr________", normalMsgType),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one",
							authz.AcceptResponse{
								Accept:  true,
								Updated: NewMockAuthorization("two", authz.AcceptResponse{}, nil),
							}, nil).WithAcceptCalls(normalMsg),
						Exp: sometime},
				},
			},
			expSaveGrant: []*SaveGrantCall{
				{
					Grantee: acc("grantee_________addr"),
					Granter: acc("granter_addr________"),
					Auth:    NewMockAuthorization("two", authz.AcceptResponse{}, nil),
					Exp:     sometime,
					Result:  newErr("test update error message"),
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			k := s.app.MetadataKeeper
			origAuthzKeeper := k.SetAuthzKeeper(tc.authzKeeper)
			defer k.SetAuthzKeeper(origAuthzKeeper)

			grantee, err := k.FindAuthzGrantee(s.FreshCtx(), tc.granter, tc.grantees, tc.msg)
			s.AssertErrorValue(err, tc.expErr, "findAuthzGrantee error")
			s.Assert().Equal(tc.expGrantee, grantee, "findAuthzGrantee grantee")

			getAuthorizationCalls := tc.authzKeeper.GetAuthorizationCalls
			s.Assert().Equal(tc.expGetAuth, getAuthorizationCalls, "calls to GetAuthorization")
			deleteGrantCalls := tc.authzKeeper.DeleteGrantCalls
			s.Assert().Equal(tc.expDelGrant, deleteGrantCalls, "calls to DeleteGrant")
			saveGrantCalls := tc.authzKeeper.SaveGrantCalls
			s.Assert().Equal(tc.expSaveGrant, saveGrantCalls, "calls to SaveGrant")
		})
	}

	s.Run("used authorizations are cached", func() {
		granter := acc("granter")
		grantees := accz("grantee1", "grantee2")

		authzKeepr := NewMockAuthzKeeper().WithGetAuthorizationResults(
			GetAuthorizationCall{
				GrantInfo: gi("grantee2", "granter", normalMsgType),
				Result: GetAuthorizationResult{
					Auth: NewMockAuthorization("one", authz.AcceptResponse{Accept: true}, nil),
					Exp:  nil,
				},
			},
		)

		expGrantee := acc("grantee2")
		expGetAuth1 := []*GetAuthorizationCall{
			{
				GrantInfo: gi("grantee1", "granter", normalMsgType),
				Result:    GetAuthorizationResult{Auth: nil, Exp: nil},
			},
			{
				GrantInfo: gi("grantee2", "granter", normalMsgType),
				Result: GetAuthorizationResult{
					Auth: NewMockAuthorization("one", authz.AcceptResponse{Accept: true}, nil).WithAcceptCalls(normalMsg),
					Exp:  nil,
				},
			},
		}
		expGetAuth2 := []*GetAuthorizationCall{
			{
				GrantInfo: gi("grantee1", "granter", normalMsgType),
				Result:    GetAuthorizationResult{Auth: nil, Exp: nil},
			},
		}

		k := s.app.MetadataKeeper
		origAuthzK := k.SetAuthzKeeper(authzKeepr)
		defer k.SetAuthzKeeper(origAuthzK)

		ctx := s.FreshCtx()
		grantee1, err1 := k.FindAuthzGrantee(ctx, granter, grantees, normalMsg)
		s.Assert().NoError(err1, "FindAuthzGrantee error first time")
		s.Assert().Equal(expGrantee, grantee1, "FindAuthzGrantee grantee first time")

		getAuthCalls1 := authzKeepr.GetAuthorizationCalls
		s.Assert().Equal(expGetAuth1, getAuthCalls1, "GetAuthorization first time")

		authzKeepr.GetAuthorizationCalls = nil
		grantee2, err2 := k.FindAuthzGrantee(ctx, granter, grantees, normalMsg)
		s.Assert().NoError(err2, "FindAuthzGrantee error second time")
		s.Assert().Equal(expGrantee, grantee2, "FindAuthzGrantee grantee second time")

		getAuthCalls2 := authzKeepr.GetAuthorizationCalls
		s.Assert().Equal(expGetAuth2, getAuthCalls2, "GetAuthorization second time")
	})
}

func (s *AuthzTestSuite) TestAssociateAuthorizations() {
	acc := func(addr string) sdk.AccAddress {
		if len(addr) == 0 {
			return nil
		}
		return sdk.AccAddress(addr)
	}
	accStr := func(addr string) string {
		if len(addr) == 0 {
			return ""
		}
		return acc(addr).String()
	}
	sw := func(addrs ...string) *keeper.SignersWrapper {
		accs := make([]string, len(addrs))
		for i, addr := range addrs {
			accs[i] = accStr(addr)
		}
		return keeper.NewSignersWrapper(accs)
	}
	// pd is a short way to create a *keeper.PartyDetails with the info needed in these tests.
	// The provided strings are passed through accStr.
	pd := func(address, signer string) *keeper.PartyDetails {
		return keeper.TestablePartyDetails{
			Address: accStr(address),
			Signer:  accStr(signer),
		}.Real()
	}
	// pde is pd "expected". It allows setting the addrAcc and signerAcc values too.
	pde := func(address, addrAcc, signer, signerAcc string) *keeper.PartyDetails {
		return keeper.TestablePartyDetails{
			Address:   accStr(address),
			Acc:       acc(addrAcc),
			Signer:    accStr(signer),
			SignerAcc: acc(signerAcc),
		}.Real()
	}
	pdz := func(parties ...*keeper.PartyDetails) []*keeper.PartyDetails {
		rv := make([]*keeper.PartyDetails, 0, len(parties))
		rv = append(rv, parties...)
		return rv
	}

	theMsg := &types.MsgWriteScopeRequest{}
	theMsgType := types.TypeURLMsgWriteScopeRequest
	authzAccept := authz.AcceptResponse{Accept: true}

	gi := func(grantee, granter string) GrantInfo {
		return GrantInfo{
			Grantee: acc(grantee),
			Granter: acc(granter),
			MsgType: theMsgType,
		}
	}
	noResCall := func(grantee, granter string) *GetAuthorizationCall {
		return &GetAuthorizationCall{
			GrantInfo: gi(grantee, granter),
			Result:    GetAuthorizationResult{Auth: nil, Exp: nil},
		}
	}

	sometimeVal := time.Unix(1324576, 0)
	sometime := &sometimeVal

	tests := []struct {
		name        string
		parties     []*keeper.PartyDetails
		signers     *keeper.SignersWrapper
		authzKeeper *MockAuthzKeeper
		expErr      string
		expParties  []*keeper.PartyDetails
		expGetAuth  []*GetAuthorizationCall
	}{
		{
			name:        "nil parties",
			parties:     nil,
			signers:     sw("ignored"),
			authzKeeper: NewMockAuthzKeeper(),
			expErr:      "",
			expParties:  nil,
		},
		{
			name:        "empty parties",
			parties:     pdz(),
			signers:     sw("ignored"),
			authzKeeper: NewMockAuthzKeeper(),
			expErr:      "",
			expParties:  pdz(),
			expGetAuth:  nil,
		},
		{
			name:        "no signers",
			parties:     pdz(pd("party1", "")),
			signers:     sw(),
			authzKeeper: NewMockAuthzKeeper(),
			expErr:      "",
			expParties:  pdz(pde("party1", "party1", "", "")),
			expGetAuth:  nil,
		},
		{
			name:        "1 party not bech32",
			parties:     pdz(keeper.TestablePartyDetails{Address: "not-correct"}.Real()),
			signers:     sw("signer1"),
			authzKeeper: NewMockAuthzKeeper(),
			expErr:      "",
			expParties:  pdz(keeper.TestablePartyDetails{Address: "not-correct"}.Real()),
			expGetAuth:  nil,
		},
		{
			name:        "1 party already has signer",
			parties:     pdz(pd("party1", "some_signer")),
			signers:     sw("signer1"),
			authzKeeper: NewMockAuthzKeeper(),
			expErr:      "",
			expParties:  pdz(pd("party1", "some_signer")),
			expGetAuth:  nil,
		},
		{
			name:        "1 party 2 signers no authorizations",
			parties:     pdz(pd("party1", "")),
			signers:     sw("signer1", "signer2"),
			authzKeeper: NewMockAuthzKeeper(),
			expErr:      "",
			expParties:  pdz(pde("party1", "party1", "", "")),
			expGetAuth:  []*GetAuthorizationCall{noResCall("signer1", "party1"), noResCall("signer2", "party1")},
		},
		{
			name:    "1 party 2 signers auth from first",
			parties: pdz(pd("party1", "")),
			signers: sw("signer1", "signer2"),
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(
				GetAuthorizationCall{
					GrantInfo: gi("signer1", "party1"),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one", authzAccept, nil),
						Exp:  sometime,
					},
				},
			),
			expErr:     "",
			expParties: pdz(pde("party1", "party1", "", "signer1")),
			expGetAuth: []*GetAuthorizationCall{
				{
					GrantInfo: gi("signer1", "party1"),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one", authzAccept, nil).WithAcceptCalls(theMsg),
						Exp:  sometime,
					},
				},
			},
		},
		{
			name:    "1 party 2 signers auth from second",
			parties: pdz(pd("party1", "")),
			signers: sw("signer1", "signer2"),
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(
				GetAuthorizationCall{
					GrantInfo: gi("signer2", "party1"),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("two", authzAccept, nil),
						Exp:  sometime,
					},
				},
			),
			expErr:     "",
			expParties: pdz(pde("party1", "party1", "", "signer2")),
			expGetAuth: []*GetAuthorizationCall{
				noResCall("signer1", "party1"),
				{
					GrantInfo: gi("signer2", "party1"),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("two", authzAccept, nil).WithAcceptCalls(theMsg),
						Exp:  sometime,
					},
				},
			},
		},
		{
			name:    "1 party 2 signers auth from both",
			parties: pdz(pd("party1", "")),
			signers: sw("signer1", "signer2"),
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(
				GetAuthorizationCall{
					GrantInfo: gi("signer1", "party1"),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one", authzAccept, nil),
						Exp:  sometime,
					},
				},
				GetAuthorizationCall{
					GrantInfo: gi("signer2", "party1"),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("two", authzAccept, nil),
						Exp:  sometime,
					},
				},
			),
			expErr:     "",
			expParties: pdz(pde("party1", "party1", "", "signer1")),
			expGetAuth: []*GetAuthorizationCall{
				{
					GrantInfo: gi("signer1", "party1"),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one", authzAccept, nil).WithAcceptCalls(theMsg),
						Exp:  sometime,
					},
				},
			},
		},
		{
			name:    "1 party 1 signer with authorization but save grant errors",
			parties: pdz(pd("party1", "")),
			signers: sw("signer1"),
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(
				GetAuthorizationCall{
					GrantInfo: gi("signer1", "party1"),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one",
							authz.AcceptResponse{
								Accept:  true,
								Updated: NewMockAuthorization("two", authzAccept, nil),
							}, nil),
						Exp: sometime,
					},
				},
			).WithSaveGrantResults(
				SaveGrantCall{
					Grantee: acc("signer1"),
					Granter: acc("party1"),
					Auth:    NewMockAuthorization("two", authzAccept, nil),
					Exp:     sometime,
					Result:  errors.New("just_some_test_error_from_SaveGrant"),
				},
			),
			expErr:     "just_some_test_error_from_SaveGrant",
			expParties: pdz(pde("party1", "party1", "", "")),
			expGetAuth: []*GetAuthorizationCall{
				{
					GrantInfo: gi("signer1", "party1"),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one",
							authz.AcceptResponse{
								Accept:  true,
								Updated: NewMockAuthorization("two", authzAccept, nil),
							}, nil).WithAcceptCalls(theMsg),
						Exp: sometime,
					},
				},
			},
		},
		{
			name: "4 parties 9 signers 3 parties already signed 4th no auth",
			parties: pdz(
				pd("party1", "party1"), pd("party2", "party2"),
				pd("party3", ""), pd("party4", "party4"),
			),
			signers: sw("signer1", "signer2", "signer3", "signer4",
				"signer5", "signer6", "signer7", "signer8", "signer9"),
			authzKeeper: NewMockAuthzKeeper(),
			expErr:      "",
			expParties: pdz(
				pd("party1", "party1"), pd("party2", "party2"),
				pde("party3", "party3", "", ""), pd("party4", "party4"),
			),
			expGetAuth: []*GetAuthorizationCall{
				noResCall("signer1", "party3"), noResCall("signer2", "party3"), noResCall("signer3", "party3"),
				noResCall("signer4", "party3"), noResCall("signer5", "party3"), noResCall("signer6", "party3"),
				noResCall("signer7", "party3"), noResCall("signer8", "party3"), noResCall("signer9", "party3"),
			},
		},
		{
			name: "4 parties 9 signers 3 parties already signed 4th with auth",
			parties: pdz(
				pd("party1", "party1"), pd("party2", "party2"),
				pd("party3", ""), pd("party4", "party4"),
			),
			signers: sw("signer1", "signer2", "signer3", "signer4",
				"signer5", "signer6", "signer7", "signer8", "signer9"),
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(
				GetAuthorizationCall{
					GrantInfo: gi("signer5", "party3"),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("three_five", authzAccept, nil),
						Exp:  sometime,
					},
				},
			),
			expErr: "",
			expParties: pdz(
				pd("party1", "party1"), pd("party2", "party2"),
				pde("party3", "party3", "", "signer5"), pd("party4", "party4"),
			),
			expGetAuth: []*GetAuthorizationCall{
				noResCall("signer1", "party3"), noResCall("signer2", "party3"),
				noResCall("signer3", "party3"), noResCall("signer4", "party3"),
				{
					GrantInfo: gi("signer5", "party3"),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("three_five", authzAccept, nil).WithAcceptCalls(theMsg),
						Exp:  sometime,
					},
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			k := s.app.MetadataKeeper
			origAuthzKeeper := k.SetAuthzKeeper(tc.authzKeeper)
			defer k.SetAuthzKeeper(origAuthzKeeper)

			err := k.AssociateAuthorizations(s.FreshCtx(), tc.parties, tc.signers, theMsg, nil)
			s.AssertErrorValue(err, tc.expErr, "associateAuthorizations")
			s.Assert().Equal(tc.expParties, tc.parties, "parties after associateAuthorizations")

			getAuthCalls := tc.authzKeeper.GetAuthorizationCalls
			s.Assert().Equal(tc.expGetAuth, getAuthCalls, "calls made to GetAuthorization")
		})
	}

	s.Run("onAssociation with counter", func() {
		counter := 0
		var partiesAssociated []*keeper.PartyDetails
		onAssoc := func(party *keeper.PartyDetails) bool {
			counter++
			partiesAssociated = append(partiesAssociated, party)
			return false
		}

		parties := pdz(
			pd("party_with_signer_1", "party_with_signer_1"),
			pd("party_without_signer_or_auth_1", ""),
			pd("party_with_auth_1", ""),
			pd("party_with_signer_2", "party_with_signer_2"),
			pd("party_with_auth_2", ""),
			pd("party_without_signer_or_auth_2", ""),
		)

		signers := sw("signer")

		expCounter := 2
		expPartiesAssociated := pdz(
			pde("party_with_auth_1", "party_with_auth_1", "", "signer"),
			pde("party_with_auth_2", "party_with_auth_2", "", "signer"),
		)
		expParties := pdz(
			pd("party_with_signer_1", "party_with_signer_1"),
			pde("party_without_signer_or_auth_1", "party_without_signer_or_auth_1", "", ""),
			pde("party_with_auth_1", "party_with_auth_1", "", "signer"),
			pd("party_with_signer_2", "party_with_signer_2"),
			pde("party_with_auth_2", "party_with_auth_2", "", "signer"),
			pde("party_without_signer_or_auth_2", "party_without_signer_or_auth_2", "", ""),
		)

		authzK := NewMockAuthzKeeper().WithGetAuthorizationResults(
			GetAuthorizationCall{
				GrantInfo: gi("signer", "party_with_auth_1"),
				Result: GetAuthorizationResult{
					Auth: NewMockAuthorization("one", authzAccept, nil),
					Exp:  sometime,
				},
			},
			GetAuthorizationCall{
				GrantInfo: gi("signer", "party_with_auth_2"),
				Result: GetAuthorizationResult{
					Auth: NewMockAuthorization("two", authzAccept, nil),
					Exp:  sometime,
				},
			},
		)

		k := s.app.MetadataKeeper
		origAuthzK := k.SetAuthzKeeper(authzK)
		defer k.SetAuthzKeeper(origAuthzK)

		err := k.AssociateAuthorizations(s.FreshCtx(), parties, signers, theMsg, onAssoc)
		s.Require().NoError(err, "associateAuthorizations")

		s.Assert().Equal(expCounter, counter, "number of times onAssociation was called")
		s.Assert().Equal(expPartiesAssociated, partiesAssociated, "parties provided to onAssociation")
		s.Assert().Equal(expParties, parties, "parties after associateAuthorizations")
	})

	s.Run("onAssociation stop early", func() {
		counter := 0
		stopAt := 3
		var partiesAssociated []*keeper.PartyDetails
		onAssoc := func(party *keeper.PartyDetails) bool {
			counter++
			partiesAssociated = append(partiesAssociated, party)
			return counter >= stopAt
		}

		parties := pdz(
			pd("party1", ""), pd("party2", ""), pd("party3", ""),
			pd("party4", ""), pd("party5", ""), pd("party6", ""),
		)

		signers := sw("signer")

		expCounter := stopAt
		expPartiesAssociated := pdz(
			pde("party1", "party1", "", "signer"),
			pde("party2", "party2", "", "signer"),
			pde("party3", "party3", "", "signer"),
		)
		expParties := pdz(
			pde("party1", "party1", "", "signer"),
			pde("party2", "party2", "", "signer"),
			pde("party3", "party3", "", "signer"),
			pd("party4", ""), pd("party5", ""), pd("party6", ""),
		)

		mockAuthzK := NewMockAuthzKeeper()
		for _, party := range parties {
			mockAuthzK.WithGetAuthorizationResults(
				GetAuthorizationCall{
					GrantInfo: GrantInfo{
						Grantee: acc("signer"),
						Granter: sdk.MustAccAddressFromBech32(party.Testable().Address),
						MsgType: theMsgType,
					},
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization(party.GetAddress(), authzAccept, nil),
						Exp:  sometime,
					},
				},
			)
		}

		k := s.app.MetadataKeeper
		origAuthzK := k.SetAuthzKeeper(mockAuthzK)
		defer k.SetAuthzKeeper(origAuthzK)

		err := k.AssociateAuthorizations(s.FreshCtx(), parties, signers, theMsg, onAssoc)
		s.Require().NoError(err, "associateAuthorizations")

		s.Assert().Equal(expCounter, counter, "number of times onAssociation was called")
		s.Assert().Equal(expPartiesAssociated, partiesAssociated, "parties provided to onAssociation")
		s.Assert().Equal(expParties, parties, "parties after associateAuthorizations")
	})
}

func (s *AuthzTestSuite) TestAssociateAuthorizationsForRoles() {
	acc := func(addr string) sdk.AccAddress {
		if len(addr) == 0 {
			return nil
		}
		return sdk.AccAddress(addr)
	}
	accStr := func(addr string) string {
		if len(addr) == 0 {
			return ""
		}
		return acc(addr).String()
	}
	sw := func(addrs ...string) *keeper.SignersWrapper {
		accs := make([]string, len(addrs))
		for i, addr := range addrs {
			accs[i] = accStr(addr)
		}
		return keeper.NewSignersWrapper(accs)
	}
	// pdu creates a usable, unsigned *keeper.PartyDetails.
	// The provided strings are passed through accStr.
	pdu := func(address string, role types.PartyType) *keeper.PartyDetails {
		return keeper.TestablePartyDetails{
			Address:         accStr(address),
			Acc:             acc(address),
			Role:            role,
			CanBeUsedBySpec: true,
			UsedBySpec:      false,
		}.Real()
	}
	// pdx creates a *keeper.PartyDetails that isn't usable.
	pdx := func(address string, role types.PartyType) *keeper.PartyDetails {
		return keeper.TestablePartyDetails{
			Address:         accStr(address),
			Acc:             acc(address),
			Role:            role,
			CanBeUsedBySpec: false,
			UsedBySpec:      false,
		}.Real()
	}
	// pdus creates a *keeper.PartyDetails that was usable but now has a signer and is used.
	pdus := func(address string, role types.PartyType, signer string) *keeper.PartyDetails {
		return keeper.TestablePartyDetails{
			Address:         accStr(address),
			Acc:             acc(address),
			Role:            role,
			CanBeUsedBySpec: true,
			UsedBySpec:      true,
			SignerAcc:       acc(signer),
		}.Real()
	}
	pdz := func(parties ...*keeper.PartyDetails) []*keeper.PartyDetails {
		rv := make([]*keeper.PartyDetails, 0, len(parties))
		rv = append(rv, parties...)
		return rv
	}
	rz := func(roles ...types.PartyType) []types.PartyType {
		rv := make([]types.PartyType, 0, len(roles))
		rv = append(rv, roles...)
		return rv
	}

	theMsg := &types.MsgWriteScopeRequest{}
	theMsgType := types.TypeURLMsgWriteScopeRequest
	authzAccept := authz.AcceptResponse{Accept: true}

	sometimeVal := time.Unix(2134567, 0)
	sometime := &sometimeVal

	gi := func(grantee, granter string) GrantInfo {
		return GrantInfo{
			Grantee: acc(grantee),
			Granter: acc(granter),
			MsgType: theMsgType,
		}
	}
	noResCall := func(grantee, granter string) *GetAuthorizationCall {
		return &GetAuthorizationCall{
			GrantInfo: gi(grantee, granter),
			Result:    GetAuthorizationResult{Auth: nil, Exp: nil},
		}
	}
	// getAuthCall creates a acceptable GetAuthorizationCall.
	getAuthCall := func(grantee, granter, name string) GetAuthorizationCall {
		return GetAuthorizationCall{
			GrantInfo: gi(grantee, granter),
			Result: GetAuthorizationResult{
				Auth: NewMockAuthorization(name, authzAccept, nil),
				Exp:  sometime,
			},
		}
	}
	// getAuthCallExp creates a acceptable GetAuthorizationCall with an AcceptCall expected.
	// This is the "expected" entry from the same args provided to getAuthCall.
	getAuthCallExp := func(grantee, granter, name string) *GetAuthorizationCall {
		return &GetAuthorizationCall{
			GrantInfo: gi(grantee, granter),
			Result: GetAuthorizationResult{
				Auth: NewMockAuthorization(name, authzAccept, nil).WithAcceptCalls(theMsg),
				Exp:  sometime,
			},
		}
	}

	unspecified := types.PartyType_PARTY_TYPE_UNSPECIFIED
	originator := types.PartyType_PARTY_TYPE_ORIGINATOR
	servicer := types.PartyType_PARTY_TYPE_SERVICER
	investor := types.PartyType_PARTY_TYPE_INVESTOR
	custodian := types.PartyType_PARTY_TYPE_CUSTODIAN
	owner := types.PartyType_PARTY_TYPE_OWNER
	affiliate := types.PartyType_PARTY_TYPE_AFFILIATE
	omnibus := types.PartyType_PARTY_TYPE_OMNIBUS
	provenance := types.PartyType_PARTY_TYPE_PROVENANCE
	controller := types.PartyType_PARTY_TYPE_CONTROLLER
	validator := types.PartyType_PARTY_TYPE_VALIDATOR

	tests := []struct {
		name        string
		roles       []types.PartyType
		parties     []*keeper.PartyDetails
		signers     *keeper.SignersWrapper
		authzKeeper *MockAuthzKeeper
		expMissing  bool
		expErr      string
		expParties  []*keeper.PartyDetails
		expGetAuth  []*GetAuthorizationCall
	}{
		{
			name:        "nil roles",
			roles:       nil,
			parties:     pdz(pdu("party1", owner)),
			signers:     sw("signer1"),
			authzKeeper: NewMockAuthzKeeper(),
			expMissing:  false,
			expErr:      "",
			expParties:  pdz(pdu("party1", owner)),
			expGetAuth:  nil,
		},
		{
			name:        "empty roles",
			roles:       rz(),
			parties:     pdz(pdu("party1", owner)),
			signers:     sw("signer1"),
			authzKeeper: NewMockAuthzKeeper(),
			expMissing:  false,
			expErr:      "",
			expParties:  pdz(pdu("party1", owner)),
			expGetAuth:  nil,
		},
		{
			name:        "1 role nil parties",
			roles:       rz(owner),
			parties:     nil,
			signers:     sw("signer1"),
			authzKeeper: NewMockAuthzKeeper(),
			expMissing:  true,
			expErr:      "",
			expParties:  nil,
			expGetAuth:  nil,
		},
		{
			name:        "1 role empty parties",
			roles:       rz(owner),
			parties:     pdz(),
			signers:     sw("signer1"),
			authzKeeper: NewMockAuthzKeeper(),
			expMissing:  true,
			expErr:      "",
			expParties:  pdz(),
			expGetAuth:  nil,
		},
		{
			name:        "empty signers",
			roles:       rz(originator),
			parties:     pdz(pdu("part1", originator)),
			signers:     sw(),
			authzKeeper: NewMockAuthzKeeper(),
			expMissing:  true,
			expErr:      "",
			expParties:  pdz(pdu("part1", originator)),
			expGetAuth:  nil,
		},

		{
			name:        "1 role unspecified with auth",
			roles:       rz(unspecified),
			parties:     pdz(pdu("party1", unspecified)),
			signers:     sw("signer1"),
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(getAuthCall("signer1", "party1", "one")),
			expMissing:  false,
			expErr:      "",
			expParties:  pdz(pdus("party1", unspecified, "signer1")),
			expGetAuth:  []*GetAuthorizationCall{getAuthCallExp("signer1", "party1", "one")},
		},
		{
			name:        "1 role unspecified no auth",
			roles:       rz(unspecified),
			parties:     pdz(pdu("party1", unspecified)),
			signers:     sw("signer1"),
			authzKeeper: NewMockAuthzKeeper(),
			expMissing:  true,
			expErr:      "",
			expParties:  pdz(pdu("party1", unspecified)),
			expGetAuth:  []*GetAuthorizationCall{noResCall("signer1", "party1")},
		},
		{
			name:        "1 role originator with auth",
			roles:       rz(originator),
			parties:     pdz(pdu("party1", originator)),
			signers:     sw("signer1"),
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(getAuthCall("signer1", "party1", "one")),
			expMissing:  false,
			expErr:      "",
			expParties:  pdz(pdus("party1", originator, "signer1")),
			expGetAuth:  []*GetAuthorizationCall{getAuthCallExp("signer1", "party1", "one")},
		},
		{
			name:        "1 role originator no auth",
			roles:       rz(originator),
			parties:     pdz(pdu("party1", originator)),
			signers:     sw("signer1"),
			authzKeeper: NewMockAuthzKeeper(),
			expMissing:  true,
			expErr:      "",
			expParties:  pdz(pdu("party1", originator)),
			expGetAuth:  []*GetAuthorizationCall{noResCall("signer1", "party1")},
		},
		{
			name:        "1 role servicer with auth",
			roles:       rz(servicer),
			parties:     pdz(pdu("party1", servicer)),
			signers:     sw("signer1"),
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(getAuthCall("signer1", "party1", "one")),
			expMissing:  false,
			expErr:      "",
			expParties:  pdz(pdus("party1", servicer, "signer1")),
			expGetAuth:  []*GetAuthorizationCall{getAuthCallExp("signer1", "party1", "one")},
		},
		{
			name:        "1 role servicer no auth",
			roles:       rz(servicer),
			parties:     pdz(pdu("party1", servicer)),
			signers:     sw("signer1"),
			authzKeeper: NewMockAuthzKeeper(),
			expMissing:  true,
			expErr:      "",
			expParties:  pdz(pdu("party1", servicer)),
			expGetAuth:  []*GetAuthorizationCall{noResCall("signer1", "party1")},
		},
		{
			name:        "1 role investor with auth",
			roles:       rz(investor),
			parties:     pdz(pdu("party1", investor)),
			signers:     sw("signer1"),
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(getAuthCall("signer1", "party1", "one")),
			expMissing:  false,
			expErr:      "",
			expParties:  pdz(pdus("party1", investor, "signer1")),
			expGetAuth:  []*GetAuthorizationCall{getAuthCallExp("signer1", "party1", "one")},
		},
		{
			name:        "1 role investor no auth",
			roles:       rz(investor),
			parties:     pdz(pdu("party1", investor)),
			signers:     sw("signer1"),
			authzKeeper: NewMockAuthzKeeper(),
			expMissing:  true,
			expErr:      "",
			expParties:  pdz(pdu("party1", investor)),
			expGetAuth:  []*GetAuthorizationCall{noResCall("signer1", "party1")},
		},
		{
			name:        "1 role custodian with auth",
			roles:       rz(custodian),
			parties:     pdz(pdu("party1", custodian)),
			signers:     sw("signer1"),
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(getAuthCall("signer1", "party1", "one")),
			expMissing:  false,
			expErr:      "",
			expParties:  pdz(pdus("party1", custodian, "signer1")),
			expGetAuth:  []*GetAuthorizationCall{getAuthCallExp("signer1", "party1", "one")},
		},
		{
			name:        "1 role custodian no auth",
			roles:       rz(custodian),
			parties:     pdz(pdu("party1", custodian)),
			signers:     sw("signer1"),
			authzKeeper: NewMockAuthzKeeper(),
			expMissing:  true,
			expErr:      "",
			expParties:  pdz(pdu("party1", custodian)),
			expGetAuth:  []*GetAuthorizationCall{noResCall("signer1", "party1")},
		},
		{
			name:        "1 role owner with auth",
			roles:       rz(owner),
			parties:     pdz(pdu("party1", owner)),
			signers:     sw("signer1"),
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(getAuthCall("signer1", "party1", "one")),
			expMissing:  false,
			expErr:      "",
			expParties:  pdz(pdus("party1", owner, "signer1")),
			expGetAuth:  []*GetAuthorizationCall{getAuthCallExp("signer1", "party1", "one")},
		},
		{
			name:        "1 role owner no auth",
			roles:       rz(owner),
			parties:     pdz(pdu("party1", owner)),
			signers:     sw("signer1"),
			authzKeeper: NewMockAuthzKeeper(),
			expMissing:  true,
			expErr:      "",
			expParties:  pdz(pdu("party1", owner)),
			expGetAuth:  []*GetAuthorizationCall{noResCall("signer1", "party1")},
		},
		{
			name:        "1 role affiliate with auth",
			roles:       rz(affiliate),
			parties:     pdz(pdu("party1", affiliate)),
			signers:     sw("signer1"),
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(getAuthCall("signer1", "party1", "one")),
			expMissing:  false,
			expErr:      "",
			expParties:  pdz(pdus("party1", affiliate, "signer1")),
			expGetAuth:  []*GetAuthorizationCall{getAuthCallExp("signer1", "party1", "one")},
		},
		{
			name:        "1 role affiliate no auth",
			roles:       rz(affiliate),
			parties:     pdz(pdu("party1", affiliate)),
			signers:     sw("signer1"),
			authzKeeper: NewMockAuthzKeeper(),
			expMissing:  true,
			expErr:      "",
			expParties:  pdz(pdu("party1", affiliate)),
			expGetAuth:  []*GetAuthorizationCall{noResCall("signer1", "party1")},
		},
		{
			name:        "1 role omnibus with auth",
			roles:       rz(omnibus),
			parties:     pdz(pdu("party1", omnibus)),
			signers:     sw("signer1"),
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(getAuthCall("signer1", "party1", "one")),
			expMissing:  false,
			expErr:      "",
			expParties:  pdz(pdus("party1", omnibus, "signer1")),
			expGetAuth:  []*GetAuthorizationCall{getAuthCallExp("signer1", "party1", "one")},
		},
		{
			name:        "1 role omnibus no auth",
			roles:       rz(omnibus),
			parties:     pdz(pdu("party1", omnibus)),
			signers:     sw("signer1"),
			authzKeeper: NewMockAuthzKeeper(),
			expMissing:  true,
			expErr:      "",
			expParties:  pdz(pdu("party1", omnibus)),
			expGetAuth:  []*GetAuthorizationCall{noResCall("signer1", "party1")},
		},
		{
			name:        "1 role provenance with auth",
			roles:       rz(provenance),
			parties:     pdz(pdu("party1", provenance)),
			signers:     sw("signer1"),
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(getAuthCall("signer1", "party1", "one")),
			expMissing:  false,
			expErr:      "",
			expParties:  pdz(pdus("party1", provenance, "signer1")),
			expGetAuth:  []*GetAuthorizationCall{getAuthCallExp("signer1", "party1", "one")},
		},
		{
			name:        "1 role provenance no auth",
			roles:       rz(provenance),
			parties:     pdz(pdu("party1", provenance)),
			signers:     sw("signer1"),
			authzKeeper: NewMockAuthzKeeper(),
			expMissing:  true,
			expErr:      "",
			expParties:  pdz(pdu("party1", provenance)),
			expGetAuth:  []*GetAuthorizationCall{noResCall("signer1", "party1")},
		},
		{
			name:        "1 role controller with auth",
			roles:       rz(controller),
			parties:     pdz(pdu("party1", controller)),
			signers:     sw("signer1"),
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(getAuthCall("signer1", "party1", "one")),
			expMissing:  false,
			expErr:      "",
			expParties:  pdz(pdus("party1", controller, "signer1")),
			expGetAuth:  []*GetAuthorizationCall{getAuthCallExp("signer1", "party1", "one")},
		},
		{
			name:        "1 role controller no auth",
			roles:       rz(controller),
			parties:     pdz(pdu("party1", controller)),
			signers:     sw("signer1"),
			authzKeeper: NewMockAuthzKeeper(),
			expMissing:  true,
			expErr:      "",
			expParties:  pdz(pdu("party1", controller)),
			expGetAuth:  []*GetAuthorizationCall{noResCall("signer1", "party1")},
		},
		{
			name:        "1 role validator with auth",
			roles:       rz(validator),
			parties:     pdz(pdu("party1", validator)),
			signers:     sw("signer1"),
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(getAuthCall("signer1", "party1", "one")),
			expMissing:  false,
			expErr:      "",
			expParties:  pdz(pdus("party1", validator, "signer1")),
			expGetAuth:  []*GetAuthorizationCall{getAuthCallExp("signer1", "party1", "one")},
		},
		{
			name:        "1 role validator no auth",
			roles:       rz(validator),
			parties:     pdz(pdu("party1", validator)),
			signers:     sw("signer1"),
			authzKeeper: NewMockAuthzKeeper(),
			expMissing:  true,
			expErr:      "",
			expParties:  pdz(pdu("party1", validator)),
			expGetAuth:  []*GetAuthorizationCall{noResCall("signer1", "party1")},
		},

		{
			name:        "1 role 3 parties none with role",
			roles:       rz(validator),
			parties:     pdz(pdu("party1", owner), pdu("party2", servicer), pdu("party3", omnibus)),
			signers:     sw("signer1"),
			authzKeeper: NewMockAuthzKeeper(),
			expMissing:  true,
			expErr:      "",
			expParties:  pdz(pdu("party1", owner), pdu("party2", servicer), pdu("party3", omnibus)),
			expGetAuth:  nil,
		},
		{
			name:        "1 role 3 parties all unusable",
			roles:       rz(investor),
			parties:     pdz(pdx("party1", investor), pdx("party2", investor), pdx("party3", investor)),
			signers:     sw("signer1"),
			authzKeeper: NewMockAuthzKeeper(),
			expMissing:  true,
			expErr:      "",
			expParties:  pdz(pdx("party1", investor), pdx("party2", investor), pdx("party3", investor)),
			expGetAuth:  nil,
		},
		{
			name:    "2 same roles 3 same parties all authed",
			roles:   rz(custodian, custodian),
			parties: pdz(pdu("party1", custodian), pdu("party2", custodian), pdu("party3", custodian)),
			signers: sw("signer1"),
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(
				getAuthCall("signer1", "party1", "one"),
				getAuthCall("signer1", "party2", "two"),
				getAuthCall("signer1", "party3", "three"),
			),
			expMissing: false,
			expErr:     "",
			expParties: pdz(
				pdus("party1", custodian, "signer1"),
				pdus("party2", custodian, "signer1"),
				pdu("party3", custodian)),
			expGetAuth: []*GetAuthorizationCall{
				getAuthCallExp("signer1", "party1", "one"),
				getAuthCallExp("signer1", "party2", "two"),
			},
		},
		{
			name:    "3 same roles 3 same parties all authed diff signers",
			roles:   rz(affiliate, affiliate, affiliate),
			parties: pdz(pdu("party1", affiliate), pdu("party2", affiliate), pdu("party3", affiliate)),
			signers: sw("signer1", "signer2", "signer3"),
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(
				getAuthCall("signer1", "party1", "one"),
				getAuthCall("signer2", "party2", "two"),
				getAuthCall("signer3", "party3", "three"),
			),
			expMissing: false,
			expErr:     "",
			expParties: pdz(
				pdus("party1", affiliate, "signer1"),
				pdus("party2", affiliate, "signer2"),
				pdus("party3", affiliate, "signer3")),
			expGetAuth: []*GetAuthorizationCall{
				getAuthCallExp("signer1", "party1", "one"),
				noResCall("signer1", "party2"),
				getAuthCallExp("signer2", "party2", "two"),
				noResCall("signer1", "party3"),
				noResCall("signer2", "party3"),
				getAuthCallExp("signer3", "party3", "three"),
			},
		},
		{
			name:    "4 same roles 3 same parties all authed diff signers",
			roles:   rz(affiliate, affiliate, affiliate, affiliate),
			parties: pdz(pdu("party1", affiliate), pdu("party2", affiliate), pdu("party3", affiliate)),
			signers: sw("signer1", "signer2", "signer3"),
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(
				getAuthCall("signer1", "party1", "one"),
				getAuthCall("signer2", "party2", "two"),
				getAuthCall("signer3", "party3", "three"),
			),
			expMissing: true,
			expErr:     "",
			expParties: pdz(
				pdus("party1", affiliate, "signer1"),
				pdus("party2", affiliate, "signer2"),
				pdus("party3", affiliate, "signer3")),
			expGetAuth: []*GetAuthorizationCall{
				getAuthCallExp("signer1", "party1", "one"),
				noResCall("signer1", "party2"),
				getAuthCallExp("signer2", "party2", "two"),
				noResCall("signer1", "party3"),
				noResCall("signer2", "party3"),
				getAuthCallExp("signer3", "party3", "three"),
			},
		},
		{
			name:    "error from associateAuthorizations",
			roles:   rz(controller),
			parties: pdz(pdu("party1", controller)),
			signers: sw("signer1"),
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(
				GetAuthorizationCall{
					GrantInfo: gi("signer1", "party1"),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one", authz.AcceptResponse{
							Accept: true,
							Delete: true,
						}, nil),
						Exp: nil,
					},
				},
			).WithDeleteGrantResults(DeleteGrantCall{
				GrantInfo: gi("signer1", "party1"),
				Result:    errors.New("test_error_from_DeleteGrant"),
			}),
			expMissing: true,
			expErr:     "test_error_from_DeleteGrant",
			expParties: pdz(pdu("party1", controller)),
			expGetAuth: []*GetAuthorizationCall{
				{
					GrantInfo: gi("signer1", "party1"),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one", authz.AcceptResponse{
							Accept: true,
							Delete: true,
						}, nil).WithAcceptCalls(theMsg),
						Exp: nil,
					},
				},
			},
		},

		{
			name:        "2 roles both missing",
			roles:       rz(omnibus, provenance),
			parties:     pdz(pdu("party1", omnibus), pdu("party2", provenance)),
			signers:     sw("signer1", "signer2"),
			authzKeeper: NewMockAuthzKeeper(),
			expMissing:  true,
			expErr:      "",
			expParties:  pdz(pdu("party1", omnibus), pdu("party2", provenance)),
			expGetAuth: []*GetAuthorizationCall{
				noResCall("signer1", "party1"),
				noResCall("signer2", "party1"),
				noResCall("signer1", "party2"),
				noResCall("signer2", "party2"),
			},
		},
		{
			name:        "2 roles missing first",
			roles:       rz(servicer, controller),
			parties:     pdz(pdu("party1", servicer), pdu("party2", controller)),
			signers:     sw("signer1", "signer2"),
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(getAuthCall("signer2", "party1", "one")),
			expMissing:  true,
			expErr:      "",
			expParties:  pdz(pdus("party1", servicer, "signer2"), pdu("party2", controller)),
			expGetAuth: []*GetAuthorizationCall{
				noResCall("signer1", "party1"),
				getAuthCallExp("signer2", "party1", "one"),
				noResCall("signer1", "party2"),
				noResCall("signer2", "party2"),
			},
		},
		{
			name:        "2 roles missing second",
			roles:       rz(servicer, controller),
			parties:     pdz(pdu("party1", servicer), pdu("party2", controller)),
			signers:     sw("signer1", "signer2"),
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(getAuthCall("signer1", "party2", "one")),
			expMissing:  true,
			expErr:      "",
			expParties:  pdz(pdu("party1", servicer), pdus("party2", controller, "signer1")),
			expGetAuth: []*GetAuthorizationCall{
				noResCall("signer1", "party1"),
				noResCall("signer2", "party1"),
				getAuthCallExp("signer1", "party2", "one"),
			},
		},
		{
			name:    "2 roles both authed",
			roles:   rz(owner, servicer),
			parties: pdz(pdu("party1", servicer), pdu("party2", owner)),
			signers: sw("signer1", "signer2"),
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(
				getAuthCall("signer1", "party1", "one"),
				getAuthCall("signer2", "party2", "two"),
			),
			expMissing: false,
			expErr:     "",
			expParties: pdz(pdus("party1", servicer, "signer1"), pdus("party2", owner, "signer2")),
			expGetAuth: []*GetAuthorizationCall{
				noResCall("signer1", "party2"),
				getAuthCallExp("signer2", "party2", "two"),
				getAuthCallExp("signer1", "party1", "one"),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			k := s.app.MetadataKeeper
			origAuthzK := k.SetAuthzKeeper(tc.authzKeeper)
			defer k.SetAuthzKeeper(origAuthzK)

			missing, err := k.AssociateAuthorizationsForRoles(s.FreshCtx(), tc.roles, tc.parties, tc.signers, theMsg)
			s.AssertErrorValue(err, tc.expErr, "associateAuthorizationsForRoles error")
			s.Assert().Equal(tc.expMissing, missing, "associateAuthorizationsForRoles missing roles bool")
			s.Assert().Equal(tc.expParties, tc.parties, "parties after associateAuthorizationsForRoles")

			getAuthCalls := tc.authzKeeper.GetAuthorizationCalls
			s.Assert().Equal(tc.expGetAuth, getAuthCalls, "calls made to GetAuthorization")
		})
	}
}

func (s *AuthzTestSuite) TestValidateProvenanceRole() {
	acc := func(addr string) sdk.AccAddress {
		return sdk.AccAddress(addr)
	}
	accStr := func(addr string) string {
		return acc(addr).String()
	}
	pd := func(canBeUsed bool, role types.PartyType, address string) *keeper.PartyDetails {
		return keeper.TestablePartyDetails{
			CanBeUsedBySpec: canBeUsed,
			Role:            role,
			Address:         address,
		}.Real()
	}
	pdz := func(parties ...*keeper.PartyDetails) []*keeper.PartyDetails {
		rv := make([]*keeper.PartyDetails, 0, len(parties))
		rv = append(rv, parties...)
		return rv
	}

	unspecified := types.PartyType_PARTY_TYPE_UNSPECIFIED
	originator := types.PartyType_PARTY_TYPE_ORIGINATOR
	servicer := types.PartyType_PARTY_TYPE_SERVICER
	investor := types.PartyType_PARTY_TYPE_INVESTOR
	custodian := types.PartyType_PARTY_TYPE_CUSTODIAN
	owner := types.PartyType_PARTY_TYPE_OWNER
	affiliate := types.PartyType_PARTY_TYPE_AFFILIATE
	omnibus := types.PartyType_PARTY_TYPE_OMNIBUS
	provenance := types.PartyType_PARTY_TYPE_PROVENANCE
	controller := types.PartyType_PARTY_TYPE_CONTROLLER
	validator := types.PartyType_PARTY_TYPE_VALIDATOR

	errNotSC := func(addr string) string {
		return fmt.Sprintf("account %q has role PROVENANCE but is not a smart contract", accStr(addr))
	}
	errNotProv := func(addr string) string {
		return fmt.Sprintf("account %q is a smart contract but does not have the PROVENANCE role", accStr(addr))
	}

	baNoKey := func(addr string, sequence uint64) *authtypes.BaseAccount {
		return &authtypes.BaseAccount{
			Address:       accStr(addr),
			PubKey:        nil,
			AccountNumber: 0,
			Sequence:      sequence,
		}
	}
	pubKey := secp256k1.GenPrivKey().PubKey()
	baWithKey := func(addr string, sequence uint64) *authtypes.BaseAccount {
		rv := baNoKey(addr, sequence)
		s.Require().NoError(rv.SetPubKey(pubKey), "SetPubKey for addr %s", addr)
		return rv
	}
	scCall := func(addr string) *GetAccountCall {
		return NewGetAccountCall(acc(addr), baNoKey(addr, 0))
	}
	nonSCCall := func(addr string) *GetAccountCall {
		return NewGetAccountCall(acc(addr), baWithKey(addr, 1))
	}
	nilCall := func(addr string) *GetAccountCall {
		return NewGetAccountCall(acc(addr), nil)
	}

	tests := []struct {
		name       string
		parties    []*keeper.PartyDetails
		authKeeper *MockAuthKeeper
		expErr     string
		expGetAcc  []*GetAccountCall
	}{
		{
			name:       "nil parties",
			parties:    nil,
			authKeeper: NewMockAuthKeeper(),
			expErr:     "",
			expGetAcc:  nil,
		},
		{
			name:       "empty parties",
			parties:    pdz(),
			authKeeper: NewMockAuthKeeper(),
			expErr:     "",
			expGetAcc:  nil,
		},
		{
			name:       "one party provenance not usable",
			parties:    pdz(pd(false, provenance, accStr("addr"))),
			authKeeper: NewMockAuthKeeper(),
			expErr:     "",
			expGetAcc:  nil,
		},
		{
			name:       "one party provenance not bech32",
			parties:    pdz(pd(true, provenance, "not_an_address")),
			authKeeper: NewMockAuthKeeper(),
			expErr:     "",
			expGetAcc:  nil,
		},
		{
			name:       "one party provenance no account",
			parties:    pdz(pd(true, provenance, accStr("no_account"))),
			authKeeper: NewMockAuthKeeper(),
			expErr:     errNotSC("no_account"),
			expGetAcc:  []*GetAccountCall{NewGetAccountCall(acc("no_account"), nil)},
		},
		{
			name:    "one party provenance not base account",
			parties: pdz(pd(true, provenance, accStr("marker_account"))),
			authKeeper: NewMockAuthKeeper().WithGetAccountResults(
				&GetAccountCall{
					Addr:   acc("marker_account"),
					Result: &markertypes.MarkerAccount{BaseAccount: baNoKey("marker_account", 0)},
				}),
			expErr: errNotSC("marker_account"),
			expGetAcc: []*GetAccountCall{
				{
					Addr:   acc("marker_account"),
					Result: &markertypes.MarkerAccount{BaseAccount: baNoKey("marker_account", 0)},
				},
			},
		},
		{
			name:    "one party provenance sequence 1",
			parties: pdz(pd(true, provenance, accStr("account_with_seq____"))),
			authKeeper: NewMockAuthKeeper().WithGetAccountResults(
				&GetAccountCall{
					Addr:   acc("account_with_seq____"),
					Result: baNoKey("account_with_seq____", 1),
				}),
			expErr: errNotSC("account_with_seq____"),
			expGetAcc: []*GetAccountCall{
				{
					Addr:   acc("account_with_seq____"),
					Result: baNoKey("account_with_seq____", 1),
				},
			},
		},
		{
			name:    "one party provenance has pub key",
			parties: pdz(pd(true, provenance, accStr("account_with_key____"))),
			authKeeper: NewMockAuthKeeper().WithGetAccountResults(
				&GetAccountCall{
					Addr:   acc("account_with_key____"),
					Result: baWithKey("account_with_key____", 0),
				}),
			expErr: errNotSC("account_with_key____"),
			expGetAcc: []*GetAccountCall{
				{
					Addr:   acc("account_with_key____"),
					Result: baWithKey("account_with_key____", 0),
				},
			},
		},
		{
			name:       "one party provenance is smart contract",
			parties:    pdz(pd(true, provenance, accStr("smart_______contract"))),
			authKeeper: NewMockAuthKeeper().WithGetAccountResults(scCall("smart_______contract")),
			expErr:     "",
			expGetAcc:  []*GetAccountCall{scCall("smart_______contract")},
		},
		{
			name:       "one party unusable not provenance is smart contract",
			parties:    pdz(pd(false, owner, accStr("smart_______contract"))),
			authKeeper: NewMockAuthKeeper(),
			expErr:     "",
			expGetAcc:  nil,
		},

		{
			name:       "smart contract account unspecified",
			parties:    pdz(pd(true, unspecified, accStr("unspecified"))),
			authKeeper: NewMockAuthKeeper().WithGetAccountResults(scCall("unspecified")),
			expErr:     errNotProv("unspecified"),
			expGetAcc:  []*GetAccountCall{scCall("unspecified")},
		},
		{
			name:       "smart contract account originator",
			parties:    pdz(pd(true, originator, accStr("originator"))),
			authKeeper: NewMockAuthKeeper().WithGetAccountResults(scCall("originator")),
			expErr:     errNotProv("originator"),
			expGetAcc:  []*GetAccountCall{scCall("originator")},
		},
		{
			name:       "smart contract account servicer",
			parties:    pdz(pd(true, servicer, accStr("servicer"))),
			authKeeper: NewMockAuthKeeper().WithGetAccountResults(scCall("servicer")),
			expErr:     errNotProv("servicer"),
			expGetAcc:  []*GetAccountCall{scCall("servicer")},
		},
		{
			name:       "smart contract account investor",
			parties:    pdz(pd(true, investor, accStr("investor"))),
			authKeeper: NewMockAuthKeeper().WithGetAccountResults(scCall("investor")),
			expErr:     errNotProv("investor"),
			expGetAcc:  []*GetAccountCall{scCall("investor")},
		},
		{
			name:       "smart contract account custodian",
			parties:    pdz(pd(true, custodian, accStr("custodian"))),
			authKeeper: NewMockAuthKeeper().WithGetAccountResults(scCall("custodian")),
			expErr:     errNotProv("custodian"),
			expGetAcc:  []*GetAccountCall{scCall("custodian")},
		},
		{
			name:       "smart contract account owner",
			parties:    pdz(pd(true, owner, accStr("owner"))),
			authKeeper: NewMockAuthKeeper().WithGetAccountResults(scCall("owner")),
			expErr:     errNotProv("owner"),
			expGetAcc:  []*GetAccountCall{scCall("owner")},
		},
		{
			name:       "smart contract account affiliate",
			parties:    pdz(pd(true, affiliate, accStr("affiliate"))),
			authKeeper: NewMockAuthKeeper().WithGetAccountResults(scCall("affiliate")),
			expErr:     errNotProv("affiliate"),
			expGetAcc:  []*GetAccountCall{scCall("affiliate")},
		},
		{
			name:       "smart contract account omnibus",
			parties:    pdz(pd(true, omnibus, accStr("omnibus"))),
			authKeeper: NewMockAuthKeeper().WithGetAccountResults(scCall("omnibus")),
			expErr:     errNotProv("omnibus"),
			expGetAcc:  []*GetAccountCall{scCall("omnibus")},
		},
		{
			name:       "smart contract account controller",
			parties:    pdz(pd(true, controller, accStr("controller"))),
			authKeeper: NewMockAuthKeeper().WithGetAccountResults(scCall("controller")),
			expErr:     errNotProv("controller"),
			expGetAcc:  []*GetAccountCall{scCall("controller")},
		},
		{
			name:       "smart contract account validator",
			parties:    pdz(pd(true, validator, accStr("validator"))),
			authKeeper: NewMockAuthKeeper().WithGetAccountResults(scCall("validator")),
			expErr:     errNotProv("validator"),
			expGetAcc:  []*GetAccountCall{scCall("validator")},
		},

		{
			name:       "normal account unspecified",
			parties:    pdz(pd(true, unspecified, accStr("unspecified"))),
			authKeeper: NewMockAuthKeeper().WithGetAccountResults(nonSCCall("unspecified")),
			expErr:     "",
			expGetAcc:  []*GetAccountCall{nonSCCall("unspecified")},
		},
		{
			name:       "normal account originator",
			parties:    pdz(pd(true, originator, accStr("originator"))),
			authKeeper: NewMockAuthKeeper().WithGetAccountResults(nonSCCall("originator")),
			expErr:     "",
			expGetAcc:  []*GetAccountCall{nonSCCall("originator")},
		},
		{
			name:       "normal account servicer",
			parties:    pdz(pd(true, servicer, accStr("servicer"))),
			authKeeper: NewMockAuthKeeper().WithGetAccountResults(nonSCCall("servicer")),
			expErr:     "",
			expGetAcc:  []*GetAccountCall{nonSCCall("servicer")},
		},
		{
			name:       "normal account investor",
			parties:    pdz(pd(true, investor, accStr("investor"))),
			authKeeper: NewMockAuthKeeper().WithGetAccountResults(nonSCCall("investor")),
			expErr:     "",
			expGetAcc:  []*GetAccountCall{nonSCCall("investor")},
		},
		{
			name:       "normal account custodian",
			parties:    pdz(pd(true, custodian, accStr("custodian"))),
			authKeeper: NewMockAuthKeeper().WithGetAccountResults(nonSCCall("custodian")),
			expErr:     "",
			expGetAcc:  []*GetAccountCall{nonSCCall("custodian")},
		},
		{
			name:       "normal account owner",
			parties:    pdz(pd(true, owner, accStr("owner"))),
			authKeeper: NewMockAuthKeeper().WithGetAccountResults(nonSCCall("owner")),
			expErr:     "",
			expGetAcc:  []*GetAccountCall{nonSCCall("owner")},
		},
		{
			name:       "normal account affiliate",
			parties:    pdz(pd(true, affiliate, accStr("affiliate"))),
			authKeeper: NewMockAuthKeeper().WithGetAccountResults(nonSCCall("affiliate")),
			expErr:     "",
			expGetAcc:  []*GetAccountCall{nonSCCall("affiliate")},
		},
		{
			name:       "normal account omnibus",
			parties:    pdz(pd(true, omnibus, accStr("omnibus"))),
			authKeeper: NewMockAuthKeeper().WithGetAccountResults(nonSCCall("omnibus")),
			expErr:     "",
			expGetAcc:  []*GetAccountCall{nonSCCall("omnibus")},
		},
		{
			name:       "normal account controller",
			parties:    pdz(pd(true, controller, accStr("controller"))),
			authKeeper: NewMockAuthKeeper().WithGetAccountResults(nonSCCall("controller")),
			expErr:     "",
			expGetAcc:  []*GetAccountCall{nonSCCall("controller")},
		},
		{
			name:       "normal account validator",
			parties:    pdz(pd(true, validator, accStr("validator"))),
			authKeeper: NewMockAuthKeeper().WithGetAccountResults(nonSCCall("validator")),
			expErr:     "",
			expGetAcc:  []*GetAccountCall{nonSCCall("validator")},
		},
		{
			name: "one of each role no accounts except smart contract",
			parties: pdz(
				pd(true, servicer, accStr("servicer")),
				pd(true, omnibus, accStr("omnibus")),
				pd(true, unspecified, accStr("unspecified")),
				pd(true, custodian, accStr("custodian")),
				pd(true, validator, accStr("validator")),
				pd(true, controller, accStr("controller")),
				pd(true, owner, accStr("owner")),
				pd(true, originator, accStr("originator")),
				pd(true, affiliate, accStr("affiliate")),
				pd(true, provenance, accStr("provenance")),
				pd(true, investor, accStr("investor")),
			),
			authKeeper: NewMockAuthKeeper().WithGetAccountResults(scCall("provenance")),
			expErr:     "",
			expGetAcc: []*GetAccountCall{
				nilCall("servicer"),
				nilCall("omnibus"),
				nilCall("unspecified"),
				nilCall("custodian"),
				nilCall("validator"),
				nilCall("controller"),
				nilCall("owner"),
				nilCall("originator"),
				nilCall("affiliate"),
				scCall("provenance"),
				nilCall("investor"),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			k := s.app.MetadataKeeper
			origAuthKeeper := k.SetAuthKeeper(tc.authKeeper)
			defer k.SetAuthKeeper(origAuthKeeper)

			err := k.ValidateProvenanceRole(s.FreshCtx(), tc.parties)
			s.AssertErrorValue(err, tc.expErr, "validateProvenanceRole")

			getAccountCalls := tc.authKeeper.GetAccountCalls
			s.Assert().Equal(tc.expGetAcc, getAccountCalls, "calls made to GetAccount")
		})
	}
}

func (s *AuthzTestSuite) TestIsWasmAccount() {
	tests := []struct {
		name  string
		authK *MockAuthKeeper
		addr  sdk.AccAddress
		exp   bool
	}{
		{
			name:  "nil addr",
			authK: NewMockAuthKeeper(),
			addr:  nil,
			exp:   false,
		},
		{
			name:  "empty addr",
			authK: NewMockAuthKeeper(),
			addr:  sdk.AccAddress{},
			exp:   false,
		},
		{
			name: "base account sequence 0 no pub key",
			authK: NewMockAuthKeeper().WithGetAccountResults(
				NewGetAccountCall(sdk.AccAddress("wasm_account"),
					authtypes.NewBaseAccount(sdk.AccAddress("wasm_account"), nil, 0, 0)),
			),
			addr: sdk.AccAddress("wasm_account"),
			exp:  true,
		},
		{
			name:  "account does not exist",
			authK: NewMockAuthKeeper(),
			addr:  sdk.AccAddress("account_doesnt_exist"),
			exp:   false,
		},
		{
			name: "marker account with sequence 0 and no pub key",
			authK: NewMockAuthKeeper().WithGetAccountResults(
				NewGetAccountCall(markertypes.MustGetMarkerAddress("bananas"),
					&markertypes.MarkerAccount{
						BaseAccount: authtypes.NewBaseAccount(markertypes.MustGetMarkerAddress("bananas"), nil, 0, 0),
					})),
			addr: markertypes.MustGetMarkerAddress("bananas"),
			exp:  false,
		},
		{
			name: "base account sequence 1 no pub key",
			authK: NewMockAuthKeeper().WithGetAccountResults(
				NewGetAccountCall(sdk.AccAddress("sequence_1"),
					authtypes.NewBaseAccount(sdk.AccAddress("sequence_1"), nil, 0, 1)),
			),
			addr: sdk.AccAddress("sequence_1"),
			exp:  false,
		},
		{
			name: "base account sequence 0 with pub key",
			authK: NewMockAuthKeeper().WithGetAccountResults(
				NewGetAccountCall(sdk.AccAddress("with_pub_key"),
					authtypes.NewBaseAccount(sdk.AccAddress("with_pub_key"), secp256k1.GenPrivKey().PubKey(), 0, 0)),
			),
			addr: sdk.AccAddress("with_pub_key"),
			exp:  false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			k := s.app.MetadataKeeper
			origAuthK := k.SetAuthKeeper(tc.authK)
			defer k.SetAuthKeeper(origAuthK)
			actual := k.IsWasmAccount(s.FreshCtx(), tc.addr)
			s.Assert().Equal(tc.exp, actual, "IsWasmAccount")
		})
	}
}

func (s *AuthzTestSuite) TestValidateSmartContractSigners() {
	acc := func(str string) sdk.AccAddress {
		return sdk.AccAddress(str)
	}
	accStr := func(str string) string {
		return acc(str).String()
	}
	normalMsg := func(signers ...string) types.MetadataMsg {
		rv := &types.MsgWriteScopeRequest{Signers: make([]string, len(signers))}
		for i, signer := range signers {
			rv.Signers[i] = accStr(signer)
		}
		return rv
	}
	normalMsgType := types.TypeURLMsgWriteScopeRequest
	gi := func(grantee, granter string) GrantInfo {
		return GrantInfo{
			Grantee: acc(grantee),
			Granter: acc(granter),
			MsgType: normalMsgType,
		}
	}
	smartAcc := func(addr string) authtypes.AccountI {
		return authtypes.NewBaseAccount(acc(addr), nil, 0, 0)
	}
	smartAccCall := func(addr string) *GetAccountCall {
		return &GetAccountCall{
			Addr:   acc(addr),
			Result: smartAcc(addr),
		}
	}
	userAcc := func(addr string) authtypes.AccountI {
		return authtypes.NewBaseAccount(acc(addr), nil, 0, 1)
	}
	userAccCall := func(addr string) *GetAccountCall {
		return &GetAccountCall{
			Addr:   acc(addr),
			Result: userAcc(addr),
		}
	}
	noAccCall := func(addr string) *GetAccountCall {
		return &GetAccountCall{Addr: acc(addr)}
	}
	authCallRes := func(name, grantee, granter string) *GetAuthorizationCall {
		return &GetAuthorizationCall{
			GrantInfo: gi(grantee, granter),
			Result: GetAuthorizationResult{
				Auth: NewMockAuthorization(name, authz.AcceptResponse{Accept: true}, nil),
				Exp:  nil,
			},
		}
	}
	authCall := func(name, grantee, granter string) GetAuthorizationCall {
		return *authCallRes(name, grantee, granter)
	}
	noAuthCallRes := func(grantee, granter string) *GetAuthorizationCall {
		return &GetAuthorizationCall{
			GrantInfo: gi(grantee, granter),
			Result:    GetAuthorizationResult{Auth: nil, Exp: nil},
		}
	}

	tests := []struct {
		name        string
		usedSigners map[string]bool
		msg         types.MetadataMsg
		authK       *MockAuthKeeper
		authzK      *MockAuthzKeeper
		expErr      string
		expGetAcc   []*GetAccountCall
		expGetAuth  []*GetAuthorizationCall
	}{
		{
			name:   "no signers",
			msg:    normalMsg(),
			expErr: "",
		},
		{
			name:      "one signer no account",
			msg:       normalMsg("signer1"),
			expErr:    "",
			expGetAcc: []*GetAccountCall{noAccCall("signer1")},
		},
		{
			name:        "one signer smart contract but in used",
			usedSigners: map[string]bool{accStr("smart_contract"): true},
			msg:         normalMsg("smart_contract"),
			authK:       NewMockAuthKeeper().WithGetAccountResults(smartAccCall("smart_contract")),
			expErr:      "",
			expGetAcc:   nil,
		},
		{
			name:      "one smart contract no used",
			msg:       normalMsg("smart_contract"),
			authK:     NewMockAuthKeeper().WithGetAccountResults(smartAccCall("smart_contract")),
			expErr:    "smart contract signer " + accStr("smart_contract") + " cannot be the last signer",
			expGetAcc: []*GetAccountCall{smartAccCall("smart_contract")},
		},
		{
			name:       "two smart contracts no used",
			msg:        normalMsg("sc1", "sc2"),
			authK:      NewMockAuthKeeper().WithGetAccountResults(smartAccCall("sc1"), smartAccCall("sc2")),
			expErr:     "smart contract signer " + accStr("sc1") + " is not authorized",
			expGetAcc:  []*GetAccountCall{smartAccCall("sc1")},
			expGetAuth: []*GetAuthorizationCall{noAuthCallRes("sc1", "sc2")},
		},
		{
			name: "smart contract then two users both with authorizations",
			msg:  normalMsg("smart_contract", "user1", "user2"),
			authK: NewMockAuthKeeper().WithGetAccountResults(
				smartAccCall("smart_contract"), userAccCall("user1"), userAccCall("user2"),
			),
			authzK: NewMockAuthzKeeper().WithGetAuthorizationResults(
				authCall("one", "smart_contract", "user1"),
				authCall("two", "smart_contract", "user2"),
			),
			expErr: "",
			expGetAcc: []*GetAccountCall{
				smartAccCall("smart_contract"), userAccCall("user1"), userAccCall("user2"),
			},
			expGetAuth: []*GetAuthorizationCall{
				authCallRes("one", "smart_contract", "user1"),
				authCallRes("two", "smart_contract", "user2"),
			},
		},
		{
			name: "smart contract then two users only 1st with authorizations",
			msg:  normalMsg("smart_contract", "user1", "user2"),
			authK: NewMockAuthKeeper().WithGetAccountResults(
				smartAccCall("smart_contract"), userAccCall("user1"), userAccCall("user2"),
			),
			authzK: NewMockAuthzKeeper().WithGetAuthorizationResults(
				authCall("one", "smart_contract", "user1"),
			),
			expErr:    "smart contract signer " + accStr("smart_contract") + " is not authorized",
			expGetAcc: []*GetAccountCall{smartAccCall("smart_contract")},
			expGetAuth: []*GetAuthorizationCall{
				authCallRes("one", "smart_contract", "user1"),
				noAuthCallRes("smart_contract", "user2"),
			},
		},
		{
			name:        "smart contract in used then two users neither with authorizations",
			usedSigners: map[string]bool{accStr("smart_contract"): true},
			msg:         normalMsg("smart_contract", "user1", "user2"),
			authK: NewMockAuthKeeper().WithGetAccountResults(
				smartAccCall("smart_contract"), userAccCall("user1"), userAccCall("user2"),
			),
			expErr: "",
			expGetAcc: []*GetAccountCall{
				userAccCall("user1"), userAccCall("user2"),
			},
		},
		{
			name:        "contract user contract user, first in used second with auth",
			usedSigners: map[string]bool{accStr("sc1"): true},
			msg:         normalMsg("sc1", "user1", "sc2", "user2"),
			authK: NewMockAuthKeeper().WithGetAccountResults(
				smartAccCall("sc1"), smartAccCall("sc2"), userAccCall("user1"), userAccCall("user2"),
			),
			authzK: NewMockAuthzKeeper().WithGetAuthorizationResults(
				authCall("two", "sc2", "user2"),
			),
			expErr: "",
			expGetAcc: []*GetAccountCall{
				userAccCall("user1"), smartAccCall("sc2"), userAccCall("user2"),
			},
			expGetAuth: []*GetAuthorizationCall{
				authCallRes("two", "sc2", "user2"),
			},
		},
		{
			name: "contract user contract user, first has auth from users but not other contract",
			msg:  normalMsg("sc1", "user1", "sc2", "user2"),
			authK: NewMockAuthKeeper().WithGetAccountResults(
				smartAccCall("sc1"), smartAccCall("sc2"), userAccCall("user1"), userAccCall("user2"),
			),
			authzK: NewMockAuthzKeeper().WithGetAuthorizationResults(
				authCall("one", "sc1", "user1"),
				authCall("two", "sc1", "user2"),
			),
			expErr:    "smart contract signer " + accStr("sc1") + " is not authorized",
			expGetAcc: []*GetAccountCall{smartAccCall("sc1")},
			expGetAuth: []*GetAuthorizationCall{
				authCallRes("one", "sc1", "user1"),
				noAuthCallRes("sc1", "sc2"),
			},
		},
		{
			name:  "error from authorization",
			msg:   normalMsg("sc1", "user1"),
			authK: NewMockAuthKeeper().WithGetAccountResults(smartAccCall("sc1")),
			authzK: NewMockAuthzKeeper().WithGetAuthorizationResults(
				GetAuthorizationCall{
					GrantInfo: gi("sc1", "user1"),
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("bad", authz.AcceptResponse{Accept: true, Delete: true}, nil),
						Exp:  nil,
					},
				},
			).WithDeleteGrantResults(DeleteGrantCall{
				GrantInfo: gi("sc1", "user1"),
				Result:    errors.New("I'm Sorry Dave, I'm Afraid I Can't Do That."),
			}),
			expErr:    "I'm Sorry Dave, I'm Afraid I Can't Do That.",
			expGetAcc: []*GetAccountCall{smartAccCall("sc1")},
			expGetAuth: []*GetAuthorizationCall{{
				GrantInfo: gi("sc1", "user1"),
				Result: GetAuthorizationResult{
					Auth: NewMockAuthorization("bad", authz.AcceptResponse{Accept: true, Delete: true}, nil),
					Exp:  nil,
				},
			}},
		},
		{
			name:        "user contrac user contract authed only to 1st user",
			usedSigners: map[string]bool{accStr("user1"): true},
			msg:         normalMsg("user1", "sc1", "user2"),
			authK:       NewMockAuthKeeper().WithGetAccountResults(smartAccCall("sc1")),
			authzK:      NewMockAuthzKeeper().WithGetAuthorizationResults(authCall("one", "sc1", "user1")),
			expErr:      "smart contract signer " + accStr("sc1") + " is not authorized",
			expGetAcc:   []*GetAccountCall{smartAccCall("sc1")},
			expGetAuth:  []*GetAuthorizationCall{noAuthCallRes("sc1", "user2")},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.authK == nil {
				tc.authK = NewMockAuthKeeper()
			}
			if tc.authzK == nil {
				tc.authzK = NewMockAuthzKeeper()
			}
			if tc.expGetAuth != nil {
				for _, auth := range tc.expGetAuth {
					if auth.Result.Auth != nil {
						mockAuth, ok := auth.Result.Auth.(*MockAuthorization)
						if ok && len(mockAuth.AcceptCalls) == 0 {
							auth.Result.Auth = mockAuth.WithAcceptCalls(tc.msg)
						}
					}
				}
			}
			k := s.app.MetadataKeeper
			origAuthK := k.SetAuthKeeper(tc.authK)
			origAuthzK := k.SetAuthzKeeper(tc.authzK)
			defer func() {
				k.SetAuthKeeper(origAuthK)
				k.SetAuthzKeeper(origAuthzK)
			}()

			err := k.ValidateSmartContractSigners(s.FreshCtx(), tc.usedSigners, tc.msg)
			s.AssertErrorValue(err, tc.expErr, "ValidateSmartContractSigners")

			getAccountCalls := tc.authK.GetAccountCalls
			s.Assert().Equal(tc.expGetAcc, getAccountCalls, "calls made to GetAccount")

			getAuthCalls := tc.authzK.GetAuthorizationCalls
			s.Assert().Equal(tc.expGetAuth, getAuthCalls, "calls made to GetAuthorization")
		})
	}
}

func (s *AuthzTestSuite) TestValidateScopeValueOwnerUpdate() {
	acc := func(addr string) sdk.AccAddress {
		return sdk.AccAddress(addr)
	}
	accStr := func(addr string) string {
		return acc(addr).String()
	}
	pd := func(address string, acc sdk.AccAddress, signer string, signerAcc sdk.AccAddress) *keeper.PartyDetails {
		return keeper.TestablePartyDetails{
			Address:   address,
			Acc:       acc,
			Signer:    signer,
			SignerAcc: signerAcc,
		}.Real()
	}

	withdrawAddrAcc := acc("withdraw_address____")
	noWithdrawAddrAcc := acc("no_withdraw_address_")
	depositAddrAcc := acc("deposit_address_____")
	noDepositAddrAcc := acc("no_deposit_address__")
	allAddrAcc := acc("all_address_________")
	noneAddrAcc := acc("none_address________")

	withdrawAddr := withdrawAddrAcc.String()
	noWithdrawAddr := noWithdrawAddrAcc.String()
	depositAddr := depositAddrAcc.String()
	noDepositAddr := noDepositAddrAcc.String()
	allAddr := allAddrAcc.String()
	noneAddr := noneAddrAcc.String()

	marker1 := &markertypes.MarkerAccount{
		BaseAccount: &authtypes.BaseAccount{},
		Manager:     "",
		AccessControl: []markertypes.AccessGrant{
			{Address: withdrawAddr, Permissions: markertypes.AccessList{markertypes.Access_Withdraw}},
			{
				Address: noWithdrawAddr,
				Permissions: markertypes.AccessList{
					markertypes.Access_Mint, markertypes.Access_Burn,
					markertypes.Access_Deposit,
					markertypes.Access_Delete, markertypes.Access_Admin, markertypes.Access_Transfer,
				},
			},
			{Address: depositAddr, Permissions: markertypes.AccessList{markertypes.Access_Deposit}},
			{
				Address: noDepositAddr,
				Permissions: markertypes.AccessList{
					markertypes.Access_Mint, markertypes.Access_Burn,
					markertypes.Access_Withdraw,
					markertypes.Access_Delete, markertypes.Access_Admin, markertypes.Access_Transfer,
				},
			},
			{
				Address: allAddr,
				Permissions: markertypes.AccessList{
					markertypes.Access_Mint, markertypes.Access_Burn,
					markertypes.Access_Deposit, markertypes.Access_Withdraw,
					markertypes.Access_Delete, markertypes.Access_Admin, markertypes.Access_Transfer,
				},
			},
		},
		Status:     markertypes.StatusActive,
		Denom:      "onecoin",
		Supply:     sdk.OneInt(),
		MarkerType: markertypes.MarkerType_RestrictedCoin,
	}
	marker1AddrAcc, marker1AddrErr := markertypes.MarkerAddress(marker1.Denom)
	s.Require().NoError(marker1AddrErr, "MarkerAddress(%q)", marker1.Denom)
	marker1.BaseAccount.Address = marker1AddrAcc.String()
	marker1Addr := marker1AddrAcc.String()

	marker2 := &markertypes.MarkerAccount{
		BaseAccount:   &authtypes.BaseAccount{},
		Manager:       "",
		AccessControl: marker1.AccessControl,
		Status:        markertypes.StatusActive,
		Denom:         "twocoin",
		Supply:        sdk.OneInt(),
		MarkerType:    markertypes.MarkerType_RestrictedCoin,
	}
	marker2AddrAcc, marker2AddrErr := markertypes.MarkerAddress(marker2.Denom)
	s.Require().NoError(marker2AddrErr, "MarkerAddress(%q)", marker2.Denom)
	marker2.BaseAccount.Address = marker2AddrAcc.String()
	marker2Addr := marker2AddrAcc.String()

	mockAuthWithMarkers := func() *MockAuthKeeper {
		return NewMockAuthKeeper().WithGetAccountResults(
			NewGetAccountCall(marker1AddrAcc, marker1),
			NewGetAccountCall(marker2AddrAcc, marker2),
		)
	}

	normalMsg := func(signers ...string) types.MetadataMsg {
		rv := &types.MsgWriteScopeRequest{
			Signers: make([]string, 0, len(signers)),
		}
		rv.Signers = append(rv.Signers, signers...)
		return rv
	}
	normalMsgType := types.TypeURLMsgWriteScopeRequest

	errMissingSigRem := func(marker *markertypes.MarkerAccount) string {
		return fmt.Sprintf("missing signature for %s (%s) with authority to withdraw/remove it as scope value owner", marker.Address, marker.Denom)
	}
	errMissingSigAdd := func(marker *markertypes.MarkerAccount) string {
		return fmt.Sprintf("missing signature for %s (%s) with authority to deposit/add it as scope value owner", marker.Address, marker.Denom)
	}
	errMissingSig := func(addr string) string {
		return fmt.Sprintf("missing signature from existing value owner %s", addr)
	}

	tests := []struct {
		name             string
		existing         string
		proposed         string
		validatedParties []*keeper.PartyDetails
		msg              types.MetadataMsg
		authKeeper       *MockAuthKeeper
		authzKeeper      *MockAuthzKeeper
		expErr           string
		expUsed          map[string]bool
		expGetAccount    []*GetAccountCall
		expGetAuth       []*GetAuthorizationCall
	}{
		{
			name:     "both empty",
			existing: "",
			proposed: "",
			expErr:   "",
		},
		{
			name:     "existing equals proposed",
			existing: "same",
			proposed: "same",
			expErr:   "",
		},
		{
			name:          "empty to non-marker",
			existing:      "",
			proposed:      accStr("new-proposed"),
			msg:           normalMsg(),
			authKeeper:    mockAuthWithMarkers(),
			expErr:        "",
			expGetAccount: []*GetAccountCall{NewGetAccountCall(acc("new-proposed"), nil)},
		},
		{
			name:          "empty to non-bech32",
			existing:      "",
			proposed:      "proposed value owner string",
			msg:           normalMsg(),
			authKeeper:    mockAuthWithMarkers(),
			expErr:        "",
			expGetAccount: []*GetAccountCall{},
		},
		{
			name:          "empty to marker no signers",
			existing:      "",
			proposed:      marker2Addr,
			msg:           normalMsg(),
			authKeeper:    mockAuthWithMarkers(),
			expErr:        errMissingSigAdd(marker2),
			expGetAccount: []*GetAccountCall{NewGetAccountCall(marker2AddrAcc, marker2)},
		},
		{
			name:          "empty to marker 1 signer only withdraw permission",
			existing:      "",
			proposed:      marker2Addr,
			msg:           normalMsg(withdrawAddr),
			authKeeper:    mockAuthWithMarkers(),
			expErr:        errMissingSigAdd(marker2),
			expGetAccount: []*GetAccountCall{NewGetAccountCall(marker2AddrAcc, marker2)},
		},
		{
			name:          "empty to marker 1 signer only deposit permission",
			existing:      "",
			proposed:      marker2Addr,
			msg:           normalMsg(depositAddr),
			authKeeper:    mockAuthWithMarkers(),
			expErr:        "",
			expUsed:       map[string]bool{depositAddr: true},
			expGetAccount: []*GetAccountCall{NewGetAccountCall(marker2AddrAcc, marker2)},
		},
		{
			name:          "empty to marker 1 signer all permissions except deposit",
			existing:      "",
			proposed:      marker2Addr,
			msg:           normalMsg(noDepositAddr),
			authKeeper:    mockAuthWithMarkers(),
			expErr:        errMissingSigAdd(marker2),
			expGetAccount: []*GetAccountCall{NewGetAccountCall(marker2AddrAcc, marker2)},
		},
		{
			name:          "empty to marker 1 signer all permissions",
			existing:      "",
			proposed:      marker2Addr,
			msg:           normalMsg(allAddr),
			authKeeper:    mockAuthWithMarkers(),
			expErr:        "",
			expUsed:       map[string]bool{allAddr: true},
			expGetAccount: []*GetAccountCall{NewGetAccountCall(marker2AddrAcc, marker2)},
		},
		{
			name:          "empty to marker three signers none with deposit",
			existing:      "",
			proposed:      marker2Addr,
			msg:           normalMsg(noneAddr, accStr("some_other_addr"), noDepositAddr),
			authKeeper:    mockAuthWithMarkers(),
			expErr:        errMissingSigAdd(marker2),
			expGetAccount: []*GetAccountCall{NewGetAccountCall(marker2AddrAcc, marker2)},
		},
		{
			name:          "empty to marker three signers one with deposit",
			existing:      "",
			proposed:      marker2Addr,
			msg:           normalMsg(noneAddr, noDepositAddr, depositAddr),
			authKeeper:    mockAuthWithMarkers(),
			expErr:        "",
			expUsed:       map[string]bool{depositAddr: true},
			expGetAccount: []*GetAccountCall{NewGetAccountCall(marker2AddrAcc, marker2)},
		},
		{
			name:          "marker to empty no signers",
			existing:      marker1Addr,
			proposed:      "",
			msg:           normalMsg(),
			authKeeper:    mockAuthWithMarkers(),
			expErr:        errMissingSigRem(marker1),
			expGetAccount: []*GetAccountCall{NewGetAccountCall(marker1AddrAcc, marker1)},
		},
		{
			name:       "marker to empty 1 signer only withdraw permission",
			existing:   marker1Addr,
			proposed:   "",
			msg:        normalMsg(withdrawAddr),
			authKeeper: mockAuthWithMarkers(),
			expErr:     "",
			expUsed:    map[string]bool{withdrawAddr: true},

			expGetAccount: []*GetAccountCall{NewGetAccountCall(marker1AddrAcc, marker1)},
		},
		{
			name:          "marker to empty 1 signer only deposit permission",
			existing:      marker1Addr,
			proposed:      "",
			msg:           normalMsg(depositAddr),
			authKeeper:    mockAuthWithMarkers(),
			expErr:        errMissingSigRem(marker1),
			expGetAccount: []*GetAccountCall{NewGetAccountCall(marker1AddrAcc, marker1)},
		},
		{
			name:          "marker to empty 1 signer all permissions except withdraw",
			existing:      marker1Addr,
			proposed:      "",
			msg:           normalMsg(noWithdrawAddr),
			authKeeper:    mockAuthWithMarkers(),
			expErr:        errMissingSigRem(marker1),
			expGetAccount: []*GetAccountCall{NewGetAccountCall(marker1AddrAcc, marker1)},
		},
		{
			name:          "marker to empty 1 signer all permissions",
			existing:      marker1Addr,
			proposed:      "",
			msg:           normalMsg(allAddr),
			authKeeper:    mockAuthWithMarkers(),
			expErr:        "",
			expUsed:       map[string]bool{allAddr: true},
			expGetAccount: []*GetAccountCall{NewGetAccountCall(marker1AddrAcc, marker1)},
		},
		{
			name:          "marker to empty three signers none with withdraw",
			existing:      marker1Addr,
			proposed:      "",
			msg:           normalMsg(noneAddr, accStr("some_other_addr"), noWithdrawAddr),
			authKeeper:    mockAuthWithMarkers(),
			expErr:        errMissingSigRem(marker1),
			expGetAccount: []*GetAccountCall{NewGetAccountCall(marker1AddrAcc, marker1)},
		},
		{
			name:          "marker to empty three signers one with withdraw",
			existing:      marker1Addr,
			proposed:      "",
			msg:           normalMsg(noneAddr, noWithdrawAddr, withdrawAddr),
			authKeeper:    mockAuthWithMarkers(),
			expErr:        "",
			expUsed:       map[string]bool{withdrawAddr: true},
			expGetAccount: []*GetAccountCall{NewGetAccountCall(marker1AddrAcc, marker1)},
		},
		{
			name:          "marker to marker no signers",
			existing:      marker1Addr,
			proposed:      marker2Addr,
			msg:           normalMsg(),
			authKeeper:    mockAuthWithMarkers(),
			expErr:        errMissingSigRem(marker1),
			expGetAccount: []*GetAccountCall{NewGetAccountCall(marker1AddrAcc, marker1)},
		},
		{
			name:          "marker to marker 1 signer no permissions",
			existing:      marker1Addr,
			proposed:      marker2Addr,
			msg:           normalMsg(noneAddr),
			authKeeper:    mockAuthWithMarkers(),
			expErr:        errMissingSigRem(marker1),
			expGetAccount: []*GetAccountCall{NewGetAccountCall(marker1AddrAcc, marker1)},
		},
		{
			name:          "marker to marker 1 signer only deposit permission",
			existing:      marker1Addr,
			proposed:      marker2Addr,
			msg:           normalMsg(depositAddr),
			authKeeper:    mockAuthWithMarkers(),
			expErr:        errMissingSigRem(marker1),
			expGetAccount: []*GetAccountCall{NewGetAccountCall(marker1AddrAcc, marker1)},
		},
		{
			name:       "marker to marker 1 signer only withdraw permission",
			existing:   marker1Addr,
			proposed:   marker2Addr,
			msg:        normalMsg(withdrawAddr),
			authKeeper: mockAuthWithMarkers(),
			expErr:     errMissingSigAdd(marker2),
			expGetAccount: []*GetAccountCall{
				NewGetAccountCall(marker1AddrAcc, marker1),
				NewGetAccountCall(marker2AddrAcc, marker2),
			},
		},
		{
			name:       "marker to marker 1 signer all permissions",
			existing:   marker1Addr,
			proposed:   marker2Addr,
			msg:        normalMsg(allAddr),
			authKeeper: mockAuthWithMarkers(),
			expErr:     "",
			expUsed:    map[string]bool{allAddr: true},
			expGetAccount: []*GetAccountCall{
				NewGetAccountCall(marker1AddrAcc, marker1),
				NewGetAccountCall(marker2AddrAcc, marker2),
			},
		},
		{
			name:       "marker to marker 2 signers only deposit then only withdraw",
			existing:   marker1Addr,
			proposed:   marker2Addr,
			msg:        normalMsg(depositAddr, withdrawAddr),
			authKeeper: mockAuthWithMarkers(),
			expErr:     "",
			expUsed:    map[string]bool{depositAddr: true, withdrawAddr: true},
			expGetAccount: []*GetAccountCall{
				NewGetAccountCall(marker1AddrAcc, marker1),
				NewGetAccountCall(marker2AddrAcc, marker2),
			},
		},
		{
			name:       "marker to marker 2 signers only withdraw then only deposit",
			existing:   marker1Addr,
			proposed:   marker2Addr,
			msg:        normalMsg(withdrawAddr, depositAddr),
			authKeeper: mockAuthWithMarkers(),
			expErr:     "",
			expUsed:    map[string]bool{withdrawAddr: true, depositAddr: true},
			expGetAccount: []*GetAccountCall{
				NewGetAccountCall(marker1AddrAcc, marker1),
				NewGetAccountCall(marker2AddrAcc, marker2),
			},
		},
		{
			name:       "marker to marker 3 signers one with withdraw one with deposit one with nothing",
			existing:   marker1Addr,
			proposed:   marker2Addr,
			msg:        normalMsg(withdrawAddr, noneAddr, depositAddr),
			authKeeper: mockAuthWithMarkers(),
			expErr:     "",
			expUsed:    map[string]bool{withdrawAddr: true, depositAddr: true},
			expGetAccount: []*GetAccountCall{
				NewGetAccountCall(marker1AddrAcc, marker1),
				NewGetAccountCall(marker2AddrAcc, marker2),
			},
		},
		{
			name:       "marker to non-marker 1 signer only withdraw",
			existing:   marker2Addr,
			proposed:   accStr("something_else"),
			msg:        normalMsg(withdrawAddr),
			authKeeper: mockAuthWithMarkers(),
			expErr:     "",
			expUsed:    map[string]bool{withdrawAddr: true},
			expGetAccount: []*GetAccountCall{
				NewGetAccountCall(marker2AddrAcc, marker2),
				NewGetAccountCall(acc("something_else"), nil),
			},
		},
		{
			name:          "marker to non-marker 1 signer no withdraw",
			existing:      marker2Addr,
			proposed:      accStr("something_else"),
			msg:           normalMsg(noWithdrawAddr),
			authKeeper:    mockAuthWithMarkers(),
			expErr:        errMissingSigRem(marker2),
			expGetAccount: []*GetAccountCall{NewGetAccountCall(marker2AddrAcc, marker2)},
		},
		{
			name:          "non-bech32 to empty in signers somehow",
			existing:      "existing_value_owner_string",
			proposed:      "",
			msg:           normalMsg(noneAddr, allAddr, "existing_value_owner_string", depositAddr),
			authKeeper:    NewMockAuthKeeper(),
			expErr:        "",
			expUsed:       map[string]bool{"existing_value_owner_string": true},
			expGetAccount: []*GetAccountCall{},
		},
		{
			name:     "non-bech32 to empty in validated parties string somehow",
			existing: "existing_value_owner_string",
			proposed: "",
			msg:      normalMsg(noneAddr, allAddr, depositAddr),
			validatedParties: []*keeper.PartyDetails{
				pd("existing_value_owner_string", nil, "existing_value_owner_string", nil),
			},
			authKeeper:    NewMockAuthKeeper(),
			expErr:        "",
			expUsed:       map[string]bool{"existing_value_owner_string": true},
			expGetAccount: []*GetAccountCall{},
		},
		{
			name:     "non-bech32 to empty not in signers or validated parties",
			existing: "existing_value_owner_string",
			proposed: "",
			validatedParties: []*keeper.PartyDetails{
				pd(noneAddr, nil, noneAddr, nil),
				pd("", allAddrAcc, "", noneAddrAcc),
				pd(depositAddr, depositAddrAcc, "", noneAddrAcc),
			},
			msg:           normalMsg(noneAddr, allAddr, depositAddr),
			authKeeper:    NewMockAuthKeeper(),
			authzKeeper:   NewMockAuthzKeeper(),
			expErr:        errMissingSig("existing_value_owner_string"),
			expGetAccount: []*GetAccountCall{},
			expGetAuth:    []*GetAuthorizationCall{},
		},
		{
			name:          "addr to empty in signers",
			existing:      accStr("existing"),
			proposed:      "",
			msg:           normalMsg(accStr("existing")),
			authKeeper:    NewMockAuthKeeper(),
			authzKeeper:   NewMockAuthzKeeper(),
			expErr:        "",
			expUsed:       map[string]bool{accStr("existing"): true},
			expGetAccount: []*GetAccountCall{NewGetAccountCall(acc("existing"), nil)},
			expGetAuth:    []*GetAuthorizationCall{},
		},
		{
			name:        "addr to other in signers",
			existing:    accStr("existing"),
			proposed:    accStr("proposed"),
			msg:         normalMsg(accStr("existing")),
			authKeeper:  NewMockAuthKeeper(),
			authzKeeper: NewMockAuthzKeeper(),
			expErr:      "",
			expUsed:     map[string]bool{accStr("existing"): true},
			expGetAccount: []*GetAccountCall{
				NewGetAccountCall(acc("existing"), nil),
				NewGetAccountCall(acc("proposed"), nil),
			},
			expGetAuth: []*GetAuthorizationCall{},
		},
		{
			name:     "addr to empty in validated parties string string",
			existing: accStr("existing"),
			proposed: "",
			validatedParties: []*keeper.PartyDetails{
				pd(noneAddr, nil, noneAddr, nil),
				pd("", allAddrAcc, "", noneAddrAcc),
				pd(depositAddr, depositAddrAcc, "", noneAddrAcc),
				pd(accStr("existing"), nil, accStr("existing"), nil),
			},
			msg:           normalMsg(noneAddr, allAddr, depositAddr),
			authKeeper:    NewMockAuthKeeper(),
			authzKeeper:   NewMockAuthzKeeper(),
			expErr:        "",
			expUsed:       map[string]bool{noneAddr: true, accStr("existing"): true},
			expGetAccount: []*GetAccountCall{NewGetAccountCall(acc("existing"), nil)},
			expGetAuth:    []*GetAuthorizationCall{},
		},
		{
			name:     "addr to empty in validated parties string acc",
			existing: accStr("existing"),
			proposed: "",
			validatedParties: []*keeper.PartyDetails{
				pd(noneAddr, nil, noneAddr, nil),
				pd("", allAddrAcc, "", noneAddrAcc),
				pd(depositAddr, depositAddrAcc, "", noneAddrAcc),
				pd(accStr("existing"), nil, "", acc("existing")),
			},
			msg:           normalMsg(noneAddr, allAddr, depositAddr),
			authKeeper:    NewMockAuthKeeper(),
			authzKeeper:   NewMockAuthzKeeper(),
			expErr:        "",
			expUsed:       map[string]bool{noneAddr: true, accStr("existing"): true},
			expGetAccount: []*GetAccountCall{NewGetAccountCall(acc("existing"), nil)},
			expGetAuth:    []*GetAuthorizationCall{},
		},
		{
			name:     "addr to empty in validated parties acc string",
			existing: accStr("existing"),
			proposed: "",
			validatedParties: []*keeper.PartyDetails{
				pd(noneAddr, nil, noneAddr, nil),
				pd("", allAddrAcc, "", noneAddrAcc),
				pd(depositAddr, depositAddrAcc, "", noneAddrAcc),
				pd("", acc("existing"), accStr("existing"), nil),
			},
			msg:           normalMsg(noneAddr, allAddr, depositAddr),
			authKeeper:    NewMockAuthKeeper(),
			authzKeeper:   NewMockAuthzKeeper(),
			expErr:        "",
			expUsed:       map[string]bool{noneAddr: true, accStr("existing"): true},
			expGetAccount: []*GetAccountCall{NewGetAccountCall(acc("existing"), nil)},
			expGetAuth:    []*GetAuthorizationCall{},
		},
		{
			name:     "addr to empty in validated parties acc acc",
			existing: accStr("existing"),
			proposed: "",
			validatedParties: []*keeper.PartyDetails{
				pd(noneAddr, nil, noneAddr, nil),
				pd("", allAddrAcc, "", noneAddrAcc),
				pd(depositAddr, depositAddrAcc, "", noneAddrAcc),
				pd("", acc("existing"), "", acc("existing")),
			},
			msg:           normalMsg(noneAddr, allAddr, depositAddr),
			authKeeper:    NewMockAuthKeeper(),
			authzKeeper:   NewMockAuthzKeeper(),
			expErr:        "",
			expUsed:       map[string]bool{noneAddr: true, accStr("existing"): true},
			expGetAccount: []*GetAccountCall{NewGetAccountCall(acc("existing"), nil)},
			expGetAuth:    []*GetAuthorizationCall{},
		},
		{
			name:     "addr to empty in validated parties other signer string",
			existing: accStr("existing"),
			proposed: "",
			validatedParties: []*keeper.PartyDetails{
				pd(noneAddr, nil, noneAddr, nil),
				pd("", allAddrAcc, "", noneAddrAcc),
				pd(depositAddr, depositAddrAcc, "", noneAddrAcc),
				pd(accStr("existing"), nil, accStr("other"), nil),
			},
			msg:           normalMsg(noneAddr, allAddr, depositAddr),
			authKeeper:    NewMockAuthKeeper(),
			authzKeeper:   NewMockAuthzKeeper(),
			expErr:        "",
			expUsed:       map[string]bool{noneAddr: true, accStr("other"): true},
			expGetAccount: []*GetAccountCall{NewGetAccountCall(acc("existing"), nil)},
			expGetAuth:    []*GetAuthorizationCall{},
		},
		{
			name:     "addr to empty in validated parties other signer acc",
			existing: accStr("existing"),
			proposed: "",
			validatedParties: []*keeper.PartyDetails{
				pd(noneAddr, nil, noneAddr, nil),
				pd("", allAddrAcc, "", noneAddrAcc),
				pd(depositAddr, depositAddrAcc, "", noneAddrAcc),
				pd(accStr("existing"), nil, "", acc("other")),
			},
			msg:           normalMsg(noneAddr, allAddr, depositAddr),
			authKeeper:    NewMockAuthKeeper(),
			authzKeeper:   NewMockAuthzKeeper(),
			expErr:        "",
			expUsed:       map[string]bool{noneAddr: true, accStr("other"): true},
			expGetAccount: []*GetAccountCall{NewGetAccountCall(acc("existing"), nil)},
			expGetAuth:    []*GetAuthorizationCall{},
		},
		{
			name:     "addr to other in validated parties",
			existing: accStr("existing"),
			proposed: accStr("proposed"),
			validatedParties: []*keeper.PartyDetails{
				pd(noneAddr, nil, noneAddr, nil),
				pd("", allAddrAcc, "", noneAddrAcc),
				pd(depositAddr, depositAddrAcc, "", noneAddrAcc),
				pd(accStr("existing"), nil, accStr("other"), nil),
			},
			msg:         normalMsg(noneAddr, allAddr, depositAddr),
			authKeeper:  NewMockAuthKeeper(),
			authzKeeper: NewMockAuthzKeeper(),
			expErr:      "",
			expUsed:     map[string]bool{noneAddr: true, accStr("other"): true},
			expGetAccount: []*GetAccountCall{
				NewGetAccountCall(acc("existing"), nil),
				NewGetAccountCall(acc("proposed"), nil),
			},
			expGetAuth: []*GetAuthorizationCall{},
		},
		{
			name:     "addr to empty with authz",
			existing: accStr("existing"),
			proposed: "",
			validatedParties: []*keeper.PartyDetails{
				pd(noneAddr, nil, noneAddr, nil),
				pd("", allAddrAcc, "", noneAddrAcc),
				pd(depositAddr, depositAddrAcc, "", noneAddrAcc),
				pd(accStr("existing"), nil, "", nil),
			},
			msg:        normalMsg(allAddr, noneAddr, depositAddr),
			authKeeper: NewMockAuthKeeper(),
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(
				GetAuthorizationCall{
					GrantInfo: GrantInfo{Grantee: noneAddrAcc, Granter: acc("existing"), MsgType: normalMsgType},
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one", authz.AcceptResponse{Accept: true}, nil),
						Exp:  nil,
					},
				},
			),
			expErr:        "",
			expUsed:       map[string]bool{noneAddr: true},
			expGetAccount: []*GetAccountCall{NewGetAccountCall(acc("existing"), nil)},
			expGetAuth: []*GetAuthorizationCall{
				{
					GrantInfo: GrantInfo{Grantee: allAddrAcc, Granter: acc("existing"), MsgType: normalMsgType},
					Result:    GetAuthorizationResult{Auth: nil, Exp: nil},
				},
				{
					GrantInfo: GrantInfo{Grantee: noneAddrAcc, Granter: acc("existing"), MsgType: normalMsgType},
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one",
							authz.AcceptResponse{Accept: true},
							nil).WithAcceptCalls(normalMsg(allAddr, noneAddr, depositAddr)),
						Exp: nil,
					},
				},
			},
		},
		{
			name:     "addr to empty not authorized",
			existing: accStr("existing"),
			proposed: "",
			validatedParties: []*keeper.PartyDetails{
				pd(noneAddr, nil, noneAddr, nil),
				pd("", allAddrAcc, "", noneAddrAcc),
				pd(depositAddr, depositAddrAcc, "", noneAddrAcc),
				pd(accStr("existing"), nil, "", nil),
			},
			msg:           normalMsg(allAddr, withdrawAddr, depositAddr),
			authKeeper:    NewMockAuthKeeper(),
			authzKeeper:   NewMockAuthzKeeper(),
			expErr:        errMissingSig(accStr("existing")),
			expGetAccount: []*GetAccountCall{NewGetAccountCall(acc("existing"), nil)},
			expGetAuth: []*GetAuthorizationCall{
				{
					GrantInfo: GrantInfo{Grantee: allAddrAcc, Granter: acc("existing"), MsgType: normalMsgType},
					Result:    GetAuthorizationResult{Auth: nil, Exp: nil},
				},
				{
					GrantInfo: GrantInfo{Grantee: withdrawAddrAcc, Granter: acc("existing"), MsgType: normalMsgType},
					Result:    GetAuthorizationResult{Auth: nil, Exp: nil},
				},
				{
					GrantInfo: GrantInfo{Grantee: depositAddrAcc, Granter: acc("existing"), MsgType: normalMsgType},
					Result:    GetAuthorizationResult{Auth: nil, Exp: nil},
				},
			},
		},
		{
			name:       "addr to empty authz error",
			existing:   accStr("existing"),
			proposed:   "",
			msg:        normalMsg(noneAddr),
			authKeeper: NewMockAuthKeeper(),
			authzKeeper: NewMockAuthzKeeper().WithGetAuthorizationResults(
				GetAuthorizationCall{
					GrantInfo: GrantInfo{Grantee: noneAddrAcc, Granter: acc("existing"), MsgType: normalMsgType},
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one",
							authz.AcceptResponse{
								Accept:  true,
								Updated: NewMockAuthorization("two", authz.AcceptResponse{}, nil),
							},
							nil),
						Exp: nil,
					},
				},
			).WithSaveGrantResults(
				SaveGrantCall{
					Grantee: noneAddrAcc,
					Granter: acc("existing"),
					Auth:    NewMockAuthorization("two", authz.AcceptResponse{}, nil),
					Exp:     nil,
					Result:  errors.New("test error from SaveGrant"),
				},
			),
			expErr:        fmt.Sprintf("authz error with existing value owner %q: %s", accStr("existing"), "test error from SaveGrant"),
			expGetAccount: []*GetAccountCall{NewGetAccountCall(acc("existing"), nil)},
			expGetAuth: []*GetAuthorizationCall{
				{
					GrantInfo: GrantInfo{Grantee: noneAddrAcc, Granter: acc("existing"), MsgType: normalMsgType},
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one",
							authz.AcceptResponse{
								Accept:  true,
								Updated: NewMockAuthorization("two", authz.AcceptResponse{}, nil),
							},
							nil).WithAcceptCalls(normalMsg(noneAddr)),
						Exp: nil,
					},
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			k := s.app.MetadataKeeper
			origAuthKeeper := k.SetAuthKeeper(tc.authKeeper)
			origAuthzKeeper := k.SetAuthzKeeper(tc.authzKeeper)
			defer func() {
				k.SetAuthKeeper(origAuthKeeper)
				k.SetAuthzKeeper(origAuthzKeeper)
			}()
			if tc.expGetAccount != nil {
				s.Require().NotNil(tc.authKeeper, "expGetAccount defined but test case does not have an authKeeper defined")
				tc.authKeeper.GetAccountCalls = make([]*GetAccountCall, 0, len(tc.expGetAccount))
			}
			if tc.expGetAuth != nil {
				s.Require().NotNil(tc.authzKeeper, "expGetAuth defined but test case does not have an authzKeeper defined")
				tc.authzKeeper.GetAuthorizationCalls = make([]*GetAuthorizationCall, 0, len(tc.expGetAuth))
			}
			if tc.expUsed == nil && len(tc.expErr) == 0 {
				tc.expUsed = make(map[string]bool)
			}

			used, err := k.ValidateScopeValueOwnerUpdate(s.FreshCtx(), tc.existing, tc.proposed, tc.validatedParties, tc.msg)
			s.AssertErrorValue(err, tc.expErr, "ValidateScopeValueOwnerUpdate")
			s.Assert().Equal(tc.expUsed, used, "ValidateScopeValueOwnerUpdate used signatures map")

			if tc.expGetAccount != nil {
				getAccountCalls := tc.authKeeper.GetAccountCalls
				s.Assert().Equal(tc.expGetAccount, getAccountCalls, "calls made to GetAccount")
			}
			if tc.expGetAuth != nil {
				getAuthCalls := tc.authzKeeper.GetAuthorizationCalls
				s.Assert().Equal(tc.expGetAuth, getAuthCalls, "calls made to GetAuthorization")
			}
		})
	}
}

func (s *AuthzTestSuite) TestValidateSignersWithoutParties() {
	// These tests are pretty light since all it does is call
	// validateAllRequiredSigned and validateSmartContractSigners.
	// The assumption is that those are well tested.

	accStr := func(str string) string {
		return sdk.AccAddress(str).String()
	}
	normalMsg := func(signers ...string) types.MetadataMsg {
		rv := &types.MsgWriteScopeRequest{Signers: make([]string, len(signers))}
		for i, signer := range signers {
			rv.Signers[i] = accStr(signer)
		}
		return rv
	}

	tests := []struct {
		name  string
		req   []string
		msg   types.MetadataMsg
		authK *MockAuthKeeper
		exp   string
	}{
		{
			name: "nil req",
			req:  nil,
			msg:  normalMsg("signer1"),
			exp:  "",
		},
		{
			name: "empty req",
			req:  []string{},
			msg:  normalMsg("signer1"),
			exp:  "",
		},
		{
			name: "one req is signer not sc",
			req:  []string{accStr("signer1")},
			msg:  normalMsg("signer1"),
			exp:  "",
		},
		{
			name: "one req is not signer",
			req:  []string{accStr("req1")},
			msg:  normalMsg("signer1"),
			exp:  "missing signature: " + accStr("req1"),
		},
		{
			name: "no req one signer is smart contract",
			msg:  normalMsg("signer1"),
			authK: NewMockAuthKeeper().WithGetAccountResults(&GetAccountCall{
				Addr:   sdk.AccAddress("signer1"),
				Result: authtypes.NewBaseAccount(sdk.AccAddress("signer1"), nil, 0, 0),
			}),
			exp: "smart contract signer " + accStr("signer1") + " cannot be the last signer",
		},
		{
			name: "one req is signer and smart contract",
			req:  []string{accStr("signer1")},
			msg:  normalMsg("signer1"),
			authK: NewMockAuthKeeper().WithGetAccountResults(&GetAccountCall{
				Addr:   sdk.AccAddress("signer1"),
				Result: authtypes.NewBaseAccount(sdk.AccAddress("signer1"), nil, 0, 0),
			}),
			exp: "",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.authK == nil {
				tc.authK = NewMockAuthKeeper()
			}
			k := s.app.MetadataKeeper
			origAuthK := k.SetAuthKeeper(tc.authK)
			defer k.SetAuthKeeper(origAuthK)

			err := k.ValidateSignersWithoutParties(s.FreshCtx(), tc.req, tc.msg)
			s.AssertErrorValue(err, tc.exp, "ValidateSignersWithoutParties")
		})
	}
}

func (s *AuthzTestSuite) TestValidateAllRequiredSigned() {
	ctx := s.FreshCtx()

	// Add a few authorizations

	// User3 can sign for User2 on MsgAddScopeDataAccessRequest.
	// Does not apply to MsgWriteScopeRequest or MsgAddScopeOwnerRequest.
	a := authz.NewGenericAuthorization(types.TypeURLMsgAddScopeDataAccessRequest)
	err := s.app.AuthzKeeper.SaveGrant(ctx, s.user3Addr, s.user2Addr, a, nil)
	s.Require().NoError(err, "SaveGrant 2 -> 3 MsgAddScopeDataAccessRequest")

	// User3 can sign for User2 on MsgWriteScopeSpecificationRequest.
	// Applies to MsgDeleteContractSpecFromScopeSpecRequest too.
	a = authz.NewGenericAuthorization(types.TypeURLMsgWriteScopeSpecificationRequest)
	err = s.app.AuthzKeeper.SaveGrant(ctx, s.user3Addr, s.user2Addr, a, nil)
	s.Require().NoError(err, "SaveGrant 2 -> 3 MsgWriteScopeSpecificationRequest")

	// User3 can sign for User1 on MsgDeleteContractSpecFromScopeSpecRequest.
	// Does not apply to MsgWriteScopeSpecificationRequest
	a = authz.NewGenericAuthorization(types.TypeURLMsgDeleteContractSpecFromScopeSpecRequest)
	err = s.app.AuthzKeeper.SaveGrant(ctx, s.user3Addr, s.user1Addr, a, nil)
	s.Require().NoError(err, "SaveGrant 2 -> 3 MsgWriteScopeSpecificationRequest")

	randAddr1 := sdk.AccAddress("random_address_1____").String()
	randAddr2 := sdk.AccAddress("random_address_2____").String()
	randAddr3 := sdk.AccAddress("random_address_3____").String()

	// expFoundSigner creates a PartyDetails for a party found as a signer.
	expFoundSigner := func(addr string) *keeper.PartyDetails {
		return keeper.TestablePartyDetails{
			Address:         addr,
			Role:            types.PartyType_PARTY_TYPE_UNSPECIFIED,
			Optional:        false,
			Acc:             nil,
			Signer:          addr,
			SignerAcc:       nil,
			CanBeUsedBySpec: false,
			UsedBySpec:      false,
		}.Real()
	}
	// expFoundAuthz creates a PartyDetails for a party found via authz.
	expFoundAuthz := func(addr string, signer sdk.AccAddress) *keeper.PartyDetails {
		rv := keeper.TestablePartyDetails{
			Address:         addr,
			Role:            types.PartyType_PARTY_TYPE_UNSPECIFIED,
			Optional:        false,
			Acc:             nil,
			Signer:          "",
			SignerAcc:       signer,
			CanBeUsedBySpec: false,
			UsedBySpec:      false,
		}.Real()
		rv.GetAcc() // need the acc of the provided addr to be set.
		return rv
	}
	// pdz is just a shorter way of creating a slice of PartyDetails.
	pdz := func(details ...*keeper.PartyDetails) []*keeper.PartyDetails {
		return details
	}

	tests := []struct {
		name     string
		owners   []string
		msg      types.MetadataMsg
		exp      []*keeper.PartyDetails
		errorMsg string
	}{
		{
			name:     "1 owner no signers",
			owners:   []string{s.user1},
			msg:      &types.MsgWriteSessionRequest{Signers: []string{}},
			errorMsg: fmt.Sprintf("missing signature: %s", s.user1),
		},
		{
			name:     "1 owner not in signers list",
			owners:   []string{s.user1},
			msg:      &types.MsgWriteSessionRequest{Signers: []string{randAddr1, randAddr2}},
			errorMsg: fmt.Sprintf("missing signature: %s", s.user1),
		},
		{
			name:     "1 owner in signers list with non-owners",
			owners:   []string{s.user1},
			msg:      &types.MsgWriteSessionRequest{Signers: []string{randAddr1, s.user1, randAddr2}},
			exp:      pdz(expFoundSigner(s.user1)),
			errorMsg: "",
		},
		{
			name:     "1 owner only signer in list",
			owners:   []string{s.user1},
			msg:      &types.MsgWriteSessionRequest{Signers: []string{s.user1}},
			exp:      pdz(expFoundSigner(s.user1)),
			errorMsg: "",
		},
		{
			name:     "2 owners no signers",
			owners:   []string{s.user1, s.user2},
			msg:      &types.MsgWriteSessionRequest{Signers: []string{}},
			errorMsg: fmt.Sprintf("missing signatures: %s, %s", s.user1, s.user2),
		},
		{
			name:     "2 owners - neither in signers list",
			owners:   []string{s.user1, s.user2},
			msg:      &types.MsgWriteSessionRequest{Signers: []string{randAddr1, randAddr2, randAddr3}},
			errorMsg: fmt.Sprintf("missing signatures: %s, %s", s.user1, s.user2),
		},
		{
			name:     "2 owners - first in signers list with non-owners",
			owners:   []string{s.user1, s.user2},
			msg:      &types.MsgWriteSessionRequest{Signers: []string{randAddr1, s.user1, randAddr3}},
			errorMsg: fmt.Sprintf("missing signature: %s", s.user2),
		},
		{
			name:     "2 owners - second in signers list with non-owners",
			owners:   []string{s.user1, s.user2},
			msg:      &types.MsgWriteSessionRequest{Signers: []string{randAddr1, s.user2, randAddr3}},
			errorMsg: fmt.Sprintf("missing signature: %s", s.user1),
		},
		{
			name:     "2 owners - both in signers list with non-owners",
			owners:   []string{s.user1, s.user2},
			msg:      &types.MsgWriteSessionRequest{Signers: []string{randAddr1, s.user2, randAddr2, s.user1, randAddr3}},
			exp:      pdz(expFoundSigner(s.user1), expFoundSigner(s.user2)),
			errorMsg: "",
		},
		{
			name:     "2 owners - both in signers list",
			owners:   []string{s.user1, s.user2},
			msg:      &types.MsgWriteSessionRequest{Signers: []string{s.user1, s.user2}},
			exp:      pdz(expFoundSigner(s.user1), expFoundSigner(s.user2)),
			errorMsg: "",
		},
		{
			name:     "2 owners - both in signers list, opposite order",
			owners:   []string{s.user1, s.user2},
			msg:      &types.MsgWriteSessionRequest{Signers: []string{s.user2, s.user1}},
			exp:      pdz(expFoundSigner(s.user1), expFoundSigner(s.user2)),
			errorMsg: "",
		},
		// authz test cases
		{
			name: "authz - 2 owners - with grant but both are signers",
			// User3 can sign for User2 on MsgAddScopeDataAccessRequest.
			owners:   []string{s.user2, s.user3},
			msg:      &types.MsgAddScopeDataAccessRequest{Signers: []string{s.user2, s.user3}},
			exp:      pdz(expFoundSigner(s.user2), expFoundSigner(s.user3)),
			errorMsg: "",
		},
		{
			name: "authz - 2 owners - 1 signer - no grant",
			// 3 has not granted anything to 2 (it's the other way around).
			owners:   []string{s.user2, s.user3},
			msg:      &types.MsgAddScopeDataAccessRequest{Signers: []string{s.user2}},
			errorMsg: fmt.Sprintf("missing signature: %s", s.user3),
		},
		{
			name: "authz - 2 owners - 1 signer - grant on child msg",
			// User3 can sign for User2 on MsgAddScopeDataAccessRequest, but not MsgWriteScopeRequest
			owners:   []string{s.user2, s.user3},
			msg:      &types.MsgWriteScopeRequest{Signers: []string{s.user3}},
			errorMsg: fmt.Sprintf("missing signature: %s", s.user2),
		},
		{
			name: "authz - 2 owners - 1 signer - grant on sibling msg",
			// User3 can sign for User2 on MsgAddScopeDataAccessRequest, but not MsgAddScopeOwnerRequest
			owners:   []string{s.user2, s.user3},
			msg:      &types.MsgAddScopeOwnerRequest{Signers: []string{s.user3}},
			errorMsg: fmt.Sprintf("missing signature: %s", s.user2),
		},
		{
			name: "authz - 2 owners - 1 signer - with grant",
			// User3 can sign for User2 on MsgAddScopeDataAccessRequest.
			owners:   []string{s.user2, s.user3},
			msg:      &types.MsgAddScopeDataAccessRequest{Signers: []string{s.user3}},
			exp:      pdz(expFoundAuthz(s.user2, s.user3Addr), expFoundSigner(s.user3)),
			errorMsg: "",
		},
		{
			name: "authz - 3 owners - 2 signers - with grant",
			// User3 can sign for User2 on MsgAddScopeDataAccessRequest.
			owners:   []string{s.user1, s.user2, s.user3},
			msg:      &types.MsgAddScopeDataAccessRequest{Signers: []string{s.user1, s.user3}},
			exp:      pdz(expFoundSigner(s.user1), expFoundAuthz(s.user2, s.user3Addr), expFoundSigner(s.user3)),
			errorMsg: "",
		},
		{
			name: "authz - 3 owners - 2 signers - grant from parent msg type",
			// User3 can sign for User2 on MsgWriteScopeSpecificationRequest.
			owners:   []string{s.user1, s.user2, s.user3},
			msg:      &types.MsgDeleteContractSpecFromScopeSpecRequest{Signers: []string{s.user1, s.user3}},
			exp:      pdz(expFoundSigner(s.user1), expFoundAuthz(s.user2, s.user3Addr), expFoundSigner(s.user3)),
			errorMsg: "",
		},
		{
			name: "authz - 3 owners - 1 signer - 2 grants",
			// User3 can sign for User1 on MsgDeleteContractSpecFromScopeSpecRequest.
			// User3 can sign for User2 on MsgWriteScopeSpecificationRequest.
			owners:   []string{s.user1, s.user2, s.user3},
			msg:      &types.MsgDeleteContractSpecFromScopeSpecRequest{Signers: []string{s.user3}},
			exp:      pdz(expFoundAuthz(s.user1, s.user3Addr), expFoundAuthz(s.user2, s.user3Addr), expFoundSigner(s.user3)),
			errorMsg: "",
		},
		{
			name: "authz - 3 owners - 1 signer - 1 grant",
			// User3 can sign for User2 on MsgWriteScopeSpecificationRequest, but not user 1.
			owners:   []string{s.user1, s.user2, s.user3},
			msg:      &types.MsgWriteScopeSpecificationRequest{Signers: []string{s.user3}},
			errorMsg: fmt.Sprintf("missing signature: %s", s.user1),
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			actual, err := s.app.MetadataKeeper.ValidateAllRequiredSigned(s.FreshCtx(), tc.owners, tc.msg)
			AssertErrorValue(t, err, tc.errorMsg, "ValidateSignersWithoutParties unexpected error")
			assert.Equal(t, tc.exp, actual, "ValidateSignersWithoutParties validated parties")
		})
	}
}

func (s *AuthzTestSuite) TestValidateAllRequiredSigned_CountAuthorizations() {

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
			errorMsg: fmt.Sprintf("missing signature: %s", s.user3),
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
			errorMsg: fmt.Sprintf("missing signature: %s", s.user2),
		},
	}

	// Test cases
	for _, tc := range tests {
		s.Run(tc.name, func() {
			ctx := s.FreshCtx()
			msgTypeURL := sdk.MsgTypeURL(tc.msg)
			if tc.grantee != nil && tc.granter != nil {
				a := authz.NewCountAuthorization(msgTypeURL, tc.count)
				err := s.app.AuthzKeeper.SaveGrant(ctx, tc.grantee, tc.granter, a, nil)
				s.Require().NoError(err, "SaveGrant")
			}

			_, err := s.app.MetadataKeeper.ValidateAllRequiredSigned(ctx, tc.owners, tc.msg)
			s.AssertErrorValue(err, tc.errorMsg, "ValidateSignersWithoutParties error")

			// validate allowedAuthorizations
			if err == nil {
				auth, _ := s.app.AuthzKeeper.GetAuthorization(ctx, tc.grantee, tc.granter, msgTypeURL)
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
		ctx := s.FreshCtx()
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
		err := s.app.AuthzKeeper.SaveGrant(ctx, s.user3Addr, s.user1Addr, a, nil)
		s.Assert().NoError(err, "SaveGrant 1 -> 3, 1 use")

		// second grant: 3 can sign for 2 two times.
		a = authz.NewCountAuthorization(msgTypeUrl, 2)
		err = s.app.AuthzKeeper.SaveGrant(ctx, s.user3Addr, s.user2Addr, a, nil)
		s.Assert().NoError(err, "SaveGrant 2 -> 3, 2 uses")

		// two owners (1 & 2), and one signer (3)
		owners := []string{s.user1, s.user2}
		msg.Signers = []string{s.user3}

		// Validate signatures. This should also use both count authorizations.
		_, err = s.app.MetadataKeeper.ValidateAllRequiredSigned(ctx, owners, msg)
		s.Assert().NoError(err, "ValidateSignersWithoutParties")

		// first grant should be deleted because it used its last use.
		auth, _ := s.app.AuthzKeeper.GetAuthorization(ctx, s.user3Addr, s.user1Addr, msgTypeUrl)
		s.Assert().Nil(auth, "GetAuthorization 1 -> 3 after only allowed use")

		// second grant should still exist, but only have one use left.
		auth, _ = s.app.AuthzKeeper.GetAuthorization(ctx, s.user3Addr, s.user2Addr, msgTypeUrl)
		s.Assert().NotNil(auth, "GetAuthorization 2 -> 3 after one use")
		actual := auth.(*authz.CountAuthorization).AllowedAuthorizations
		s.Assert().Equal(1, int(actual), "number of uses left on 2 -> 3 authorization")
	})
}

func TestValidateRolesPresent(t *testing.T) {
	// p is a short way to create a Party.
	p := func(addr string, role types.PartyType) types.Party {
		return types.Party{
			Address:  addr,
			Role:     role,
			Optional: false,
		}
	}

	// pz is a short way to create a slice of Parties.
	pz := func(parties ...types.Party) []types.Party {
		rv := make([]types.Party, 0, len(parties))
		rv = append(rv, parties...)
		return rv
	}

	// ptz is a short way to create a slice of PartyTypes.
	ptz := func(roles ...types.PartyType) []types.PartyType {
		rv := make([]types.PartyType, 0, len(roles))
		rv = append(rv, roles...)
		return rv
	}

	tests := []struct {
		name     string
		parties  []types.Party
		reqRoles []types.PartyType
		exp      string
	}{
		{
			name:     "nil nil",
			parties:  nil,
			reqRoles: nil,
			exp:      "",
		},
		{
			name:     "nil empty",
			parties:  nil,
			reqRoles: ptz(),
			exp:      "",
		},
		{
			name:     "empty nil",
			parties:  pz(),
			reqRoles: nil,
			exp:      "",
		},
		{
			name:     "empty empty",
			parties:  pz(),
			reqRoles: ptz(),
			exp:      "",
		},
		{
			name:     "one req two parties present in both",
			parties:  pz(p("addr1", 1), p("addr2", 1)),
			reqRoles: ptz(1),
			exp:      "",
		},
		{
			name:     "one req two parties present in first",
			parties:  pz(p("addr1", 1), p("addr2", 2)),
			reqRoles: ptz(1),
			exp:      "",
		},
		{
			name:     "one req two parties present in second",
			parties:  pz(p("addr1", 1), p("addr2", 2)),
			reqRoles: ptz(2),
			exp:      "",
		},
		{
			name:     "one req two parties not present",
			parties:  pz(p("addr1", 1), p("addr2", 2)),
			reqRoles: ptz(3),
			exp:      "missing roles required by spec: INVESTOR need 1 have 0",
		},
		{
			name:     "two diff req two parties present",
			parties:  pz(p("addr1", 1), p("addr2", 2)),
			reqRoles: ptz(2, 1),
			exp:      "",
		},
		{
			name:     "two diff req two parties first not present",
			parties:  pz(p("addr1", 1), p("addr2", 2)),
			reqRoles: ptz(3, 1),
			exp:      "missing roles required by spec: INVESTOR need 1 have 0",
		},
		{
			name:     "two diff req two parties second not present",
			parties:  pz(p("addr1", 1), p("addr2", 2)),
			reqRoles: ptz(2, 3),
			exp:      "missing roles required by spec: INVESTOR need 1 have 0",
		},
		{
			name:     "two diff req two parties neither present",
			parties:  pz(p("addr1", 1), p("addr2", 2)),
			reqRoles: ptz(3, 4),
			exp:      "missing roles required by spec: INVESTOR need 1 have 0, CUSTODIAN need 1 have 0",
		},
		{
			name:     "two same req two parties present",
			parties:  pz(p("addr1", 1), p("addr2", 1)),
			reqRoles: ptz(1, 1),
			exp:      "",
		},
		{
			name:     "two same req two parties only one",
			parties:  pz(p("addr1", 1), p("addr2", 2)),
			reqRoles: ptz(1, 1),
			exp:      "missing roles required by spec: ORIGINATOR need 2 have 1",
		},
		{
			name:     "two same req two parties none",
			parties:  pz(p("addr1", 1), p("addr2", 2)),
			reqRoles: ptz(3, 3),
			exp:      "missing roles required by spec: INVESTOR need 2 have 0",
		},
		{
			name: "crazy but ok",
			parties: pz(
				p("addr1", 1), p("addr1", 2), p("addr1", 3), p("addr1", 4),
				p("addr2", 1), p("addr2", 2), p("addr2", 3), p("addr2", 4),
				p("addr3", 1), p("addr3", 2), p("addr3", 3), p("addr3", 4),
				p("addr4", 1), p("addr4", 2), p("addr4", 3), p("addr4", 4),
			),
			reqRoles: ptz(1, 1, 1, 2, 2, 2, 2, 3, 3, 3, 3, 4, 4, 4),
			exp:      "",
		},
		{
			name: "crazy not okay",
			parties: pz(
				p("addr1", 1), p("addr1", 2), p("addr1", 3), p("addr1", 4),
				p("addr2", 1), p("addr2", 2), p("addr2", 3), p("addr2", 4),
				p("addr3", 1), p("addr3", 2), p("addr3", 3), p("addr3", 4),
				p("addr4", 1), p("addr4", 2), p("addr4", 3), p("addr4", 4),
				p("addr5", 11),
			),
			reqRoles: ptz(1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2, 3, 5, 5, 5, 11, 11),
			exp:      "missing roles required by spec: SERVICER need 7 have 4, OWNER need 3 have 0, VALIDATOR need 2 have 1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := keeper.ValidateRolesPresent(tc.parties, tc.reqRoles)
			AssertErrorValue(t, err, tc.exp, "validateRolesPresent")
		})
	}
}

func TestValidatePartiesArePresent(t *testing.T) {
	// p is a short way to create a Party.
	p := func(addr string, role types.PartyType, optional bool) types.Party {
		return types.Party{
			Address:  addr,
			Role:     role,
			Optional: optional,
		}
	}

	// pz is a short way to create a slice of parties.
	pz := func(parties ...types.Party) []types.Party {
		rv := make([]types.Party, 0, len(parties))
		rv = append(rv, parties...)
		return rv
	}

	tests := []struct {
		name      string
		required  []types.Party
		available []types.Party
		exp       string
	}{
		{
			name:      "nil nil",
			required:  nil,
			available: nil,
			exp:       "",
		},
		{
			name:      "empty nil",
			required:  pz(),
			available: nil,
			exp:       "",
		},
		{
			name:      "nil empty",
			required:  nil,
			available: pz(),
			exp:       "",
		},
		{
			name:      "empty empty",
			required:  pz(),
			available: pz(),
			exp:       "",
		},
		{
			name:      "no req some available",
			required:  pz(),
			available: pz(p("a", 1, false)),
			exp:       "",
		},
		{
			name:      "one req is available same optional",
			required:  pz(p("a", 1, false)),
			available: pz(p("a", 1, false)),
			exp:       "",
		},
		{
			name:      "one req one available diff optional",
			required:  pz(p("a", 1, false)),
			available: pz(p("a", 1, false)),
			exp:       "",
		},
		{
			name:      "one req one avail diff addr",
			required:  pz(p("addr1", 1, false)),
			available: pz(p("b", 1, false)),
			exp:       "missing party: addr1 (ORIGINATOR)",
		},
		{
			name:      "one req one avail diff role",
			required:  pz(p("addr1", 1, false)),
			available: pz(p("addr1", 2, false)),
			exp:       "missing party: addr1 (ORIGINATOR)",
		},
		{
			name:     "three req five avail all present",
			required: pz(p("addr1", 1, false), p("addr2", 2, true), p("addr3", 3, false)),
			available: pz(p("addr2", 2, false), p("addr3", 3, true), p("addrX", 8, true),
				p("addrY", 9, false), p("addr1", 1, true)),
			exp: "",
		},
		{
			name:     "three req five avail none present",
			required: pz(p("addr1", 1, false), p("addr2", 2, true), p("addr3", 3, false)),
			available: pz(p("addrV", 4, false), p("addrW", 5, false),
				p("addrX", 6, false), p("addrY", 7, false), p("addrZ", 8, false)),
			exp: "missing parties: addr1 (ORIGINATOR), addr2 (SERVICER), addr3 (INVESTOR)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := keeper.ValidatePartiesArePresent(tc.required, tc.available)
			AssertErrorValue(t, err, tc.exp, "validatePartiesArePresent")
		})
	}
}

func (s *AuthzTestSuite) TestGetMarkerAndCheckAuthority() {
	markerAddr := markertypes.MustGetMarkerAddress("testcoin").String()
	marker := markertypes.MarkerAccount{
		BaseAccount: &authtypes.BaseAccount{
			Address:       markerAddr,
			AccountNumber: 23,
		},
		AccessControl: []markertypes.AccessGrant{
			{
				Address:     s.user1,
				Permissions: markertypes.AccessListByNames("deposit,withdraw"),
			},
			{
				Address:     s.user2,
				Permissions: markertypes.AccessListByNames("burn,mint"),
			},
		},
		Denom:      "testcoin",
		Supply:     sdk.NewInt(1000),
		MarkerType: markertypes.MarkerType_Coin,
		Status:     markertypes.StatusActive,
	}
	s.Require().NoError(s.app.MarkerKeeper.AddMarkerAccount(s.FreshCtx(), &marker), "AddMarkerAccount")
	// s.user1 has an account created in TestSetup.

	tests := []struct {
		name      string
		addr      string
		signers   []string
		role      markertypes.Access
		expMarker markertypes.MarkerAccountI
		expHasAcc bool
		expSig    string
	}{
		{name: "invalid address", addr: "invalid", expMarker: nil},
		{name: "account does not exist", addr: sdk.AccAddress("does-not-exist").String(), expMarker: nil},
		{name: "account exists but is not marker", addr: s.user1, expMarker: nil},
		{
			name:      "is marker does not have signer",
			addr:      markerAddr,
			signers:   []string{s.user3},
			expMarker: &marker,
			expHasAcc: false,
		},
		{
			name:      "is marker with signer 1 but not role",
			addr:      markerAddr,
			signers:   []string{s.user1},
			role:      markertypes.Access_Transfer,
			expMarker: &marker,
			expHasAcc: false,
		},
		{
			name:      "is marker with signer 1 with role other user has",
			addr:      markerAddr,
			signers:   []string{s.user1},
			role:      markertypes.Access_Burn,
			expMarker: &marker,
			expHasAcc: false,
		},
		{
			name:      "is marker with signer 1 and role 1",
			addr:      markerAddr,
			signers:   []string{s.user1},
			role:      markertypes.Access_Deposit,
			expMarker: &marker,
			expHasAcc: true,
			expSig:    s.user1,
		},
		{
			name:      "is marker with signer 1 and role 2",
			addr:      markerAddr,
			signers:   []string{s.user1},
			role:      markertypes.Access_Withdraw,
			expMarker: &marker,
			expHasAcc: true,
			expSig:    s.user1,
		},
		{
			name:      "is marker with signer 2 but not role",
			addr:      markerAddr,
			signers:   []string{s.user2},
			role:      markertypes.Access_Transfer,
			expMarker: &marker,
			expHasAcc: false,
		},
		{
			name:      "is marker with signer 2 with role other user has",
			addr:      markerAddr,
			signers:   []string{s.user2},
			role:      markertypes.Access_Deposit,
			expMarker: &marker,
			expHasAcc: false,
		},
		{
			name:      "is marker with signer 2 and role 1",
			addr:      markerAddr,
			signers:   []string{s.user2},
			role:      markertypes.Access_Burn,
			expMarker: &marker,
			expHasAcc: true,
			expSig:    s.user2,
		},
		{
			name:      "is marker with signer 2 and role 2",
			addr:      markerAddr,
			signers:   []string{s.user2},
			role:      markertypes.Access_Mint,
			expMarker: &marker,
			expHasAcc: true,
			expSig:    s.user2,
		},
		{
			name:      "is marker both signers role from first",
			addr:      markerAddr,
			signers:   []string{s.user1, s.user2},
			role:      markertypes.Access_Withdraw,
			expMarker: &marker,
			expHasAcc: true,
			expSig:    s.user1,
		},
		{
			name:      "is marker both signers role from second",
			addr:      markerAddr,
			signers:   []string{s.user1, s.user2},
			role:      markertypes.Access_Mint,
			expMarker: &marker,
			expHasAcc: true,
			expSig:    s.user2,
		},
		{
			name:      "is marker both signers neither have role",
			addr:      markerAddr,
			signers:   []string{s.user1, s.user2},
			role:      markertypes.Access_Transfer,
			expMarker: &marker,
			expHasAcc: false,
		},
		{
			name:      "is marker two signers first has role",
			addr:      markerAddr,
			signers:   []string{s.user1, s.user3},
			role:      markertypes.Access_Withdraw,
			expMarker: &marker,
			expHasAcc: true,
			expSig:    s.user1,
		},
		{
			name:      "is marker two signers second has role",
			addr:      markerAddr,
			signers:   []string{s.user3, s.user2},
			role:      markertypes.Access_Burn,
			expMarker: &marker,
			expHasAcc: true,
			expSig:    s.user2,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			actualMarker, actualHasAcc, actualSig := s.app.MetadataKeeper.GetMarkerAndCheckAuthority(s.FreshCtx(), tc.addr, tc.signers, tc.role)
			s.Assert().Equal(tc.expMarker, actualMarker, "GetMarkerAndCheckAuthority marker")
			s.Assert().Equal(tc.expHasAcc, actualHasAcc, "GetMarkerAndCheckAuthority has access")
			s.Assert().Equal(tc.expSig, actualSig, "GetMarkerAndCheckAuthority signer")
		})
	}
}
