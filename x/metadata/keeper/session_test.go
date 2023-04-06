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
	s.app.AccountKeeper.SetAccount(ctx, s.app.AccountKeeper.NewAccountWithAddress(ctx, s.user1Addr))

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
	ctx := s.FreshCtx()
	scope := types.NewScope(s.scopeID, s.scopeSpecID, ownerPartyList(s.user1), []string{s.user1}, s.user1)
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

	partiesInvolved := []types.PartyType{types.PartyType_PARTY_TYPE_AFFILIATE}
	contractSpec := types.NewContractSpecification(s.contractSpecID, types.NewDescription("name", "desc", "url", "icon"), []string{s.user1}, partiesInvolved, &types.ContractSpecification_Hash{"hash"}, "processname")
	s.app.MetadataKeeper.SetContractSpecification(ctx, *contractSpec)
	scopeSpec := types.NewScopeSpecification(s.scopeSpecID, nil, []string{s.user1}, partiesInvolved, []types.MetadataAddress{s.contractSpecID})
	s.app.MetadataKeeper.SetScopeSpecification(ctx, *scopeSpec)

	cases := map[string]struct {
		existing *types.Session
		proposed *types.Session
		signers  []string
		errorMsg string
	}{
		"invalid session update, existing record does not have scope": {
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
			errorMsg: "parties can only be optional when the scope has require_party_rollup = true",
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
