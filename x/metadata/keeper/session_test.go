package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/metadata/keeper"
	"github.com/provenance-io/provenance/x/metadata/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/stretchr/testify/suite"
)

type SessionKeeperTestSuite struct {
	suite.Suite

	app         *simapp.App
	queryClient types.QueryClient

	pubkey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress

	pubkey2   cryptotypes.PubKey
	user2     string
	user2Addr sdk.AccAddress

	scopeUUID uuid.UUID
	scopeID   types.MetadataAddress

	scopeSpecUUID uuid.UUID
	scopeSpecID   types.MetadataAddress

	recordName string
	recordID   types.MetadataAddress

	sessionUUID uuid.UUID
	sessionID   types.MetadataAddress

	contractSpecUUID uuid.UUID
	contractSpecID   types.MetadataAddress
	recordSpecID     types.MetadataAddress
}

func (s *SessionKeeperTestSuite) SetupTest() {
	s.app = simapp.Setup(s.T())
	ctx := s.FreshCtx()
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, s.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, s.app.MetadataKeeper)
	s.queryClient = types.NewQueryClient(queryHelper)

	s.pubkey1 = secp256k1.GenPrivKey().PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()
	user1Acc := s.app.AccountKeeper.NewAccountWithAddress(ctx, s.user1Addr)
	s.Require().NoError(user1Acc.SetSequence(1), "user1Acc.SetSequence(1)")
	s.app.AccountKeeper.SetAccount(ctx, user1Acc)

	s.pubkey2 = secp256k1.GenPrivKey().PubKey()
	s.user2Addr = sdk.AccAddress(s.pubkey2.Address())
	s.user2 = s.user2Addr.String()

	s.scopeUUID = uuid.New()
	s.scopeID = types.ScopeMetadataAddress(s.scopeUUID)

	s.scopeSpecUUID = uuid.New()
	s.scopeSpecID = types.ScopeSpecMetadataAddress(s.scopeSpecUUID)

	s.recordName = "TestRecord"
	s.recordID = types.RecordMetadataAddress(s.scopeUUID, s.recordName)

	s.sessionUUID = uuid.New()
	s.sessionID = types.SessionMetadataAddress(s.scopeUUID, s.sessionUUID)

	s.contractSpecUUID = uuid.New()
	s.contractSpecID = types.ContractSpecMetadataAddress(s.contractSpecUUID)
	s.recordSpecID = types.RecordSpecMetadataAddress(s.contractSpecUUID, s.recordName)
}

func (s *SessionKeeperTestSuite) FreshCtx() sdk.Context {
	return keeper.AddAuthzCacheToContext(s.app.BaseApp.NewContext(false, tmproto.Header{}))
}

func TestSessionKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(SessionKeeperTestSuite))
}

// func ownerPartyList defined in keeper_test.go

func (s *SessionKeeperTestSuite) TestSessionGetSetRemove() {
	ctx := s.FreshCtx()
	r, found := s.app.MetadataKeeper.GetSession(ctx, s.sessionID)
	s.Empty(r)
	s.False(found)

	session := types.NewSession("name", s.sessionID, s.contractSpecID, []types.Party{
		{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: types.PartyType_PARTY_TYPE_AFFILIATE}},
		&types.AuditFields{CreatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", CreatedDate: time.Now(),
			UpdatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", UpdatedDate: time.Now(),
			Message: "message",
		})

	s.NotNil(session)
	s.app.MetadataKeeper.SetSession(ctx, *session)

	sess, found := s.app.MetadataKeeper.GetSession(ctx, s.sessionID)
	s.True(found)
	s.NotEmpty(sess)

	s.app.MetadataKeeper.RemoveSession(ctx, s.sessionID)
	sess, found = s.app.MetadataKeeper.GetSession(ctx, s.sessionID)
	s.False(found)
	s.Empty(sess)

	process := types.NewProcess("processname", &types.Process_Hash{Hash: "HASH"}, "process_method")
	record := types.NewRecord(s.recordName, s.sessionID, *process, []types.RecordInput{}, []types.RecordOutput{}, s.recordSpecID)
	s.app.MetadataKeeper.SetRecord(ctx, *record)
	s.app.MetadataKeeper.SetSession(ctx, *session)

	sess, found = s.app.MetadataKeeper.GetSession(ctx, s.sessionID)
	s.True(found)
	s.NotEmpty(sess)

	s.app.MetadataKeeper.RemoveSession(ctx, s.sessionID)
	sess, found = s.app.MetadataKeeper.GetSession(ctx, s.sessionID)
	s.True(found)
	s.NotEmpty(sess)
}

func (s *SessionKeeperTestSuite) TestSessionIterator() {
	ctx := s.FreshCtx()
	for i := 1; i <= 10; i++ {
		sessionID := types.SessionMetadataAddress(s.scopeUUID, uuid.New())
		session := types.NewSession("name", sessionID, s.contractSpecID, []types.Party{
			{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: types.PartyType_PARTY_TYPE_AFFILIATE}},
			&types.AuditFields{CreatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", CreatedDate: time.Now(),
				UpdatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", UpdatedDate: time.Now(),
				Message: "message",
			})
		s.app.MetadataKeeper.SetSession(ctx, *session)
	}
	count := 0
	err := s.app.MetadataKeeper.IterateSessions(ctx, s.scopeID, func(s types.Session) (stop bool) {
		count++
		return false
	})
	s.Require().NoError(err, "IterateSessions")
	s.Equal(10, count, "iterator should return a full list of sessions")
}

func (s *SessionKeeperTestSuite) TestValidateWriteSession() {
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
	scope := types.NewScope(s.scopeID, s.scopeSpecID, ownerPartyList(s.user1), []string{s.user1}, s.user1, false)
	s.app.MetadataKeeper.SetScope(ctx, *scope)

	auditTime := time.Now()

	invalidScopeUUID := uuid.New()
	parties := []types.Party{{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: types.PartyType_PARTY_TYPE_AFFILIATE}}
	validSession := types.NewSession("processname", s.sessionID, s.contractSpecID, parties, nil)
	validSessionWithAudit := types.NewSession("processname", s.sessionID, s.contractSpecID, parties, &types.AuditFields{CreatedDate: auditTime, CreatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Version: 1})
	invalidIDSession := types.NewSession("processname", types.SessionMetadataAddress(invalidScopeUUID, uuid.New()), s.contractSpecID, parties, nil)
	invalidContractID := types.NewSession("processname", s.sessionID, types.ContractSpecMetadataAddress(uuid.New()), parties, nil)
	invalidParties := []types.Party{{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: types.PartyType_PARTY_TYPE_CUSTODIAN}}
	invalidPartiesSession := types.NewSession("processname", s.sessionID, s.contractSpecID, invalidParties, nil)

	owner := types.PartyType_PARTY_TYPE_OWNER
	affiliate := types.PartyType_PARTY_TYPE_AFFILIATE
	provenance := types.PartyType_PARTY_TYPE_PROVENANCE

	partiesInvolved := []types.PartyType{types.PartyType_PARTY_TYPE_AFFILIATE}
	contractSpec := types.NewContractSpecification(s.contractSpecID, types.NewDescription("name", "desc", "url", "icon"), []string{s.user1}, partiesInvolved, &types.ContractSpecification_Hash{"hash"}, "processname")
	s.app.MetadataKeeper.SetContractSpecification(ctx, *contractSpec)
	otherContractSpec := types.NewContractSpecification(types.ContractSpecMetadataAddress(uuid.New()), types.NewDescription("other name", "other desc", "other url", "other icon"), []string{s.user1}, partiesInvolved, &types.ContractSpecification_Hash{"other_hash"}, "other_processname")
	s.app.MetadataKeeper.SetContractSpecification(ctx, *otherContractSpec)
	scopeSpec := types.NewScopeSpecification(s.scopeSpecID, nil, []string{s.user1}, partiesInvolved, []types.MetadataAddress{s.contractSpecID})
	s.app.MetadataKeeper.SetScopeSpecification(ctx, *scopeSpec)

	scopeWOSpec := types.Scope{
		ScopeId:         types.ScopeMetadataAddress(uuid.New()),
		SpecificationId: types.ScopeSpecMetadataAddress(uuid.New()),
		Owners:          ownerPartyList(s.user1),
	}
	s.app.MetadataKeeper.SetScope(ctx, scopeWOSpec)
	sessIDInScopeWOSpec := scopeWOSpec.ScopeId.MustGetAsSessionAddress(uuid.New())

	rollupScopeSpec := types.ScopeSpecification{
		SpecificationId: types.ScopeSpecMetadataAddress(uuid.New()),
		Description:     types.NewDescription("tester", "tester desc", "", ""),
		OwnerAddresses:  []string{s.user1},
		PartiesInvolved: []types.PartyType{owner, affiliate},
		ContractSpecIds: []types.MetadataAddress{s.contractSpecID},
	}
	s.app.MetadataKeeper.SetScopeSpecification(ctx, rollupScopeSpec)
	rollupScopeSpecID := rollupScopeSpec.SpecificationId
	user3 := sdk.AccAddress("user_3_address______").String()
	otherUser := sdk.AccAddress("other_user_address__").String()

	rollupScope := types.Scope{
		ScopeId:            types.ScopeMetadataAddress(uuid.New()),
		SpecificationId:    rollupScopeSpecID,
		Owners:             ptz(pt(s.user1, owner, false), pt(s.user2, affiliate, true), pt(user3, affiliate, true)),
		RequirePartyRollup: true,
	}
	s.app.MetadataKeeper.SetScope(ctx, rollupScope)
	rollupScopeWProv := types.Scope{
		ScopeId:            types.ScopeMetadataAddress(uuid.New()),
		SpecificationId:    rollupScopeSpecID,
		Owners:             ptz(pt(s.user1, owner, false), pt(s.user1, provenance, true), pt(s.user2, affiliate, true)),
		RequirePartyRollup: true,
	}
	s.app.MetadataKeeper.SetScope(ctx, rollupScopeWProv)

	rollupSessionID := rollupScope.ScopeId.MustGetAsSessionAddress(uuid.New())
	newRollupSession := func(parties ...types.Party) *types.Session {
		return &types.Session{
			SessionId:       rollupSessionID,
			SpecificationId: validSession.SpecificationId,
			Parties:         parties,
			Name:            validSession.Name,
			Context:         validSession.Context,
			Audit:           validSession.Audit,
		}
	}
	rollupSessionWProvID := rollupScopeWProv.ScopeId.MustGetAsSessionAddress(uuid.New())
	newRollupSessionWProv := func(parties ...types.Party) *types.Session {
		return &types.Session{
			SessionId:       rollupSessionWProvID,
			SpecificationId: validSession.SpecificationId,
			Parties:         parties,
			Name:            validSession.Name,
			Context:         validSession.Context,
			Audit:           validSession.Audit,
		}
	}

	cases := map[string]struct {
		existing *types.Session
		proposed *types.Session
		signers  []string
		errorMsg string
	}{
		"invalid session update, proposed does not have scope": {
			existing: nil,
			proposed: &types.Session{},
			signers:  []string{s.user1},
			errorMsg: "address is empty",
		},
		"valid session update, existing and proposed satisfy validation": {
			existing: validSession,
			proposed: validSession,
			signers:  []string{s.user1},
			errorMsg: "",
		},
		"valid session update without audit, existing has audit": {
			existing: validSessionWithAudit,
			proposed: validSession,
			signers:  []string{s.user1},
			errorMsg: "",
		},
		"invalid session update has audit, existing has no audit": {
			existing: validSession,
			proposed: validSessionWithAudit,
			signers:  []string{s.user1},
			errorMsg: "attempt to modify audit fields, modification not allowed",
		},
		"valid session update existing and new have matching audits": {
			existing: validSessionWithAudit,
			proposed: validSessionWithAudit,
			signers:  []string{s.user1},
			errorMsg: "",
		},
		"invalid session update, existing id does not match proposed": {
			existing: validSession,
			proposed: invalidIDSession,
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("cannot update session identifier. expected %s, got %s", validSession.SessionId, invalidIDSession.SessionId),
		},
		"invalid session update, scope does not exist": {
			existing: invalidIDSession,
			proposed: invalidIDSession,
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("scope not found for scope id %s", types.ScopeMetadataAddress(invalidScopeUUID)),
		},
		"invalid session update, cannot change contract spec": {
			existing: validSession,
			proposed: invalidContractID,
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("cannot update specification identifier. expected %s, got %s",
				validSession.SpecificationId, invalidContractID.SpecificationId),
		},
		"invalid session update, contract spec does not exist": {
			existing: nil,
			proposed: invalidContractID,
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("cannot find contract specification %s", invalidContractID.SpecificationId),
		},
		"invalid session update, involved parties do not match": {
			existing: validSession,
			proposed: invalidPartiesSession,
			signers:  []string{s.user1},
			errorMsg: "missing roles required by spec: AFFILIATE need 1 have 0",
		},
		"scope spec not found": {
			existing: nil,
			proposed: &types.Session{
				SessionId:       sessIDInScopeWOSpec,
				SpecificationId: s.contractSpecID,
				Parties:         ownerPartyList(s.user1),
				Name:            validSession.Name,
				Context:         validSession.Context,
				Audit:           validSession.Audit,
			},
			signers:  []string{s.user1},
			errorMsg: "scope spec not found with id " + scopeWOSpec.SpecificationId.String(),
		},
		"contract spec not allowed in scope spec": {
			existing: nil,
			proposed: &types.Session{
				SessionId:       validSession.SessionId,
				SpecificationId: otherContractSpec.SpecificationId,
				Parties:         ownerPartyList(s.user1),
				Name:            validSession.Name,
				Context:         validSession.Context,
				Audit:           validSession.Audit,
			},
			signers:  []string{s.user1},
			errorMsg: "contract spec " + otherContractSpec.SpecificationId.String() + " not listed in scope spec " + scopeSpec.SpecificationId.String(),
		},
		"invalid session update, missing required signers": {
			existing: validSession,
			proposed: validSession,
			signers:  []string{},
			errorMsg: fmt.Sprintf("missing signature: %s", s.user1),
		},
		"invalid session update, invalid proposed name of empty to existing session": {
			existing: validSessionWithAudit,
			proposed: types.NewSession("", s.sessionID, s.contractSpecID, parties, &types.AuditFields{CreatedDate: auditTime, CreatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Version: 1, Message: "fault"}),
			signers:  []string{s.user1},
			errorMsg: "proposed name to existing session must not be empty",
		},
		"invalid session update, modified audit message": {
			existing: validSessionWithAudit,
			proposed: types.NewSession("processname", s.sessionID, s.contractSpecID, parties, &types.AuditFields{CreatedDate: auditTime, CreatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Version: 1, Message: "fault"}),
			signers:  []string{s.user1},
			errorMsg: "attempt to modify message audit field, modification not allowed",
		},
		"invalid session update, modified audit version": {
			existing: validSessionWithAudit,
			proposed: types.NewSession("processname", s.sessionID, s.contractSpecID, parties, &types.AuditFields{CreatedDate: auditTime, CreatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Version: 2}),
			signers:  []string{s.user1},
			errorMsg: "attempt to modify version audit field, modification not allowed",
		},
		"invalid session update, modified audit update date": {
			existing: validSessionWithAudit,
			proposed: types.NewSession("processname", s.sessionID, s.contractSpecID, parties, &types.AuditFields{CreatedDate: auditTime, CreatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Version: 1, UpdatedDate: time.Now()}),
			signers:  []string{s.user1},
			errorMsg: "attempt to modify updated-date audit field, modification not allowed",
		},
		"invalid session update, modified audit update by": {
			existing: validSessionWithAudit,
			proposed: types.NewSession("processname", s.sessionID, s.contractSpecID, parties, &types.AuditFields{CreatedDate: auditTime, CreatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Version: 1, UpdatedBy: "fault"}),
			signers:  []string{s.user1},
			errorMsg: "attempt to modify updated-by audit field, modification not allowed",
		},
		"invalid session update, modified audit created by": {
			existing: validSessionWithAudit,
			proposed: types.NewSession("processname", s.sessionID, s.contractSpecID, parties, &types.AuditFields{CreatedDate: auditTime, CreatedBy: "fault", Version: 1, UpdatedBy: "fault"}),
			signers:  []string{s.user1},
			errorMsg: "attempt to modify created-by audit field, modification not allowed",
		},
		"invalid session update, modified audit created date": {
			existing: validSessionWithAudit,
			proposed: types.NewSession("processname", s.sessionID, s.contractSpecID, parties, &types.AuditFields{CreatedDate: time.Now(), CreatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Version: 1}),
			signers:  []string{s.user1},
			errorMsg: "attempt to modify created-date audit field, modification not allowed",
		},
		"optional parties only allowed with rollup": {
			existing: nil,
			proposed: &types.Session{
				SessionId:       s.sessionID,
				SpecificationId: s.contractSpecID,
				Parties: []types.Party{
					{
						Address:  sdk.AccAddress("just_some_address___").String(),
						Role:     types.PartyType_PARTY_TYPE_CUSTODIAN,
						Optional: true,
					},
				},
				Name:    "some_name",
				Context: nil,
				Audit:   nil,
			},
			signers:  nil,
			errorMsg: "parties can only be optional when require_party_rollup = true",
		},
		"with rollup party not in scope": {
			existing: nil,
			proposed: newRollupSession(pt(otherUser, affiliate, true)),
			signers:  []string{s.user1, s.user2},
			errorMsg: "not all session parties in scope owners: missing party: " + otherUser + " (AFFILIATE)",
		},
		"with rollup missing signature from required scope party": {
			existing: nil,
			proposed: newRollupSession(pt(s.user2, affiliate, true)),
			signers:  []string{s.user2},
			errorMsg: "missing required signature: " + s.user1 + " (OWNER)",
		},
		"with rollup not new missing required role": {
			existing: newRollupSession(pt(s.user2, affiliate, true)),
			proposed: newRollupSession(pt(s.user1, owner, true)),
			signers:  []string{s.user1, s.user2},
			errorMsg: "missing roles required by spec: AFFILIATE need 1 have 0",
		},
		"with rollup not new missing signature from required role": {
			existing: newRollupSession(pt(s.user2, affiliate, true)),
			proposed: newRollupSession(pt(user3, affiliate, true)),
			signers:  []string{s.user1},
			errorMsg: "missing signers for roles required by spec: AFFILIATE need 1 have 0",
		},
		"with rollup not new missing signature from required session party": {
			existing: newRollupSession(pt(s.user2, affiliate, false)),
			proposed: newRollupSession(pt(user3, affiliate, false)),
			signers:  []string{s.user1, user3},
			errorMsg: "missing required signature: " + s.user2 + " (AFFILIATE)",
		},
		"with rollup new missing signature from required role": {
			existing: nil,
			proposed: newRollupSession(pt(s.user2, affiliate, true)),
			signers:  []string{s.user1},
			errorMsg: "missing signers for roles required by spec: AFFILIATE need 1 have 0",
		},
		"new with provenance non-smart-contract party": {
			existing: nil,
			proposed: &types.Session{
				SessionId:       types.SessionMetadataAddress(s.scopeUUID, uuid.New()),
				SpecificationId: s.contractSpecID,
				Parties:         ptz(pt(s.user1, owner, false), pt(s.user1, provenance, false), pt(s.user2, affiliate, false)),
				Name:            "withprov",
			},
			signers:  []string{s.user1},
			errorMsg: `account "` + s.user1 + `" has role PROVENANCE but is not a smart contract`,
		},
		"updating with provenance non-smart-contract party": {
			existing: &types.Session{
				SessionId:       s.sessionID,
				SpecificationId: s.contractSpecID,
				Parties:         ptz(pt(s.user1, owner, false), pt(s.user2, affiliate, false)),
				Name:            "withprov",
			},
			proposed: &types.Session{
				SessionId:       s.sessionID,
				SpecificationId: s.contractSpecID,
				Parties:         ptz(pt(s.user1, owner, false), pt(s.user1, provenance, false), pt(s.user2, affiliate, false)),
				Name:            "withprov",
			},
			signers:  []string{s.user1},
			errorMsg: `account "` + s.user1 + `" has role PROVENANCE but is not a smart contract`,
		},
		"with rollup new with provenance non-smart-contract party": {
			existing: nil,
			proposed: newRollupSessionWProv(pt(s.user1, owner, false), pt(s.user1, provenance, true), pt(s.user2, affiliate, false)),
			signers:  []string{s.user1, s.user2},
			errorMsg: `account "` + s.user1 + `" has role PROVENANCE but is not a smart contract`,
		},
		"with rollup updating with provenance non-smart-contract party": {
			existing: newRollupSessionWProv(pt(s.user1, owner, false), pt(s.user2, affiliate, false)),
			proposed: newRollupSessionWProv(pt(s.user1, owner, false), pt(s.user1, provenance, true), pt(s.user2, affiliate, false)),
			signers:  []string{s.user1, s.user2},
			errorMsg: `account "` + s.user1 + `" has role PROVENANCE but is not a smart contract`,
		},
	}

	for name, tc := range cases {
		s.Run(name, func() {
			msg := &types.MsgWriteSessionRequest{
				Session: *tc.proposed,
				Signers: tc.signers,
			}
			err := s.app.MetadataKeeper.ValidateWriteSession(s.FreshCtx(), tc.existing, msg)
			if len(tc.errorMsg) > 0 {
				s.Assert().EqualError(err, tc.errorMsg, "ValidateWriteSession")
			} else {
				s.Assert().NoError(err, "ValidateWriteSession")
			}
		})
	}
}

// TODO: ValidateAuditUpdate tests
