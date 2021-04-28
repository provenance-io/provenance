package types

import (
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ScopeTestSuite struct {
	suite.Suite

	Addr string
}

// func ownerPartyList is defined in msg_test.go

func TestScopeTestSuite(t *testing.T) {
	suite.Run(t, new(ScopeTestSuite))
}

func (s *ScopeTestSuite) SetupSuite() {
	s.T().Parallel()

	pubHex := "85EA54E8598B27EC37EAEEEEA44F1E78A9B5E671"
	addrHex, err := hex.DecodeString(pubHex)
	s.Require().NoError(err, "DecodeString(%s) error", pubHex)
	s.Addr = sdk.AccAddress(addrHex).String()
}

func (s *ScopeTestSuite) TestScopeValidateBasic() {
	tests := []struct {
		name    string
		scope   *Scope
		want    string
		wantErr bool
	}{
		{
			"valid scope one owner",
			NewScope(ScopeMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), ownerPartyList(s.Addr), []string{}, ""),
			"",
			false,
		},
		{
			"valid scope one owner, one data access",
			NewScope(ScopeMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), ownerPartyList(s.Addr), []string{s.Addr}, ""),
			"",
			false,
		},
		{
			"no owners",
			NewScope(ScopeMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), []Party{}, []string{}, ""),
			"scope must have at least one owner",
			true,
		},
		{
			"no owners, data access",
			NewScope(ScopeMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), []Party{}, []string{s.Addr}, ""),
			"scope must have at least one owner",
			true,
		},
		{
			"invalid scope id",
			NewScope(ScopeSpecMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), []Party{}, []string{}, ""),
			"invalid scope identifier (expected: scope, got scopespec)",
			true,
		},
		{
			"invalid scope id - wrong address type",
			NewScope(MetadataAddress{0x85}, ScopeSpecMetadataAddress(uuid.New()), []Party{}, []string{}, ""),
			"invalid metadata address type: 133",
			true,
		},
		{
			"invalid spec id",
			NewScope(ScopeMetadataAddress(uuid.New()), ScopeMetadataAddress(uuid.New()), []Party{}, []string{}, ""),
			"invalid scope specification identifier (expected: scopespec, got scope)",
			true,
		},
		{
			"invalid spec id - wrong address type",
			NewScope(ScopeMetadataAddress(uuid.New()), MetadataAddress{0x85}, []Party{}, []string{}, ""),
			"invalid metadata address type: 133",
			true,
		},
		{
			"invaid owner on scope",
			NewScope(ScopeMetadataAddress(
				uuid.New()),
				ScopeSpecMetadataAddress(uuid.New()),
				ownerPartyList(":invalid"),
				[]string{},
				"",
			),
			"invalid owner on scope: decoding bech32 failed: invalid index of 1",
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			err := tt.scope.ValidateBasic()
			if (err != nil) != tt.wantErr {
				t.Errorf("Scope ValidateBasic error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				require.Equal(t, tt.want, err.Error())
			}
		})
	}
}

func (s *ScopeTestSuite) TestScopeString() {
	s.T().Run("scope string", func(t *testing.T) {
		scopeUUID := uuid.MustParse("8d80b25a-c089-4446-956e-5d08cfe3e1a5")
		sessionUUID := uuid.MustParse("c25c7bd4-c639-4367-a842-f64fa5fccc19")
		scope := NewScope(ScopeMetadataAddress(
			scopeUUID), ScopeSpecMetadataAddress(sessionUUID),
			ownerPartyList(s.Addr),
			[]string{},
			"")
		require.Equal(t, `scope_id: scope1qzxcpvj6czy5g354dews3nlruxjsahhnsp
specification_id: scopespec1qnp9c775ccu5xeaggtmylf0uesvsqyrkq8
owners:
- address: cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck
  role: 5
data_access: []
value_owner_address: ""
`,
			scope.String())
	})
}

func (s *ScopeTestSuite) TestRecordValidateBasic() {
	scopeUUID := uuid.New()
	sessionUUID := uuid.New()
	sessionID := SessionMetadataAddress(scopeUUID, sessionUUID)
	recordID := RecordMetadataAddress(scopeUUID, "test_record")
	validRI := NewRecordInput("ri_name", &RecordInput_Hash{"hash"}, "ri_type", RecordInputStatus_Proposed)
	validRO := NewRecordOutput("ro_hash", ResultStatus_RESULT_STATUS_PASS)
	validPs := NewProcess("process_name", &Process_Hash{"address"}, "method")
	tests := []struct {
		name    string
		record  *Record
		want    string
		wantErr bool
	}{
		{
			"valid record",
			NewRecord("name", sessionID, *validPs, []RecordInput{*validRI}, []RecordOutput{*validRO}, nil),
			"",
			false,
		},
		{
			"invalid record, invalid/missing name for record",
			NewRecord("", sessionID, *validPs, []RecordInput{*validRI}, []RecordOutput{*validRO}, nil),
			"invalid/missing name for record",
			true,
		},
		{
			"invalid record, missing sessionid",
			NewRecord("name", nil, *validPs, []RecordInput{*validRI}, []RecordOutput{*validRO}, nil),
			"address is empty",
			true,
		},
		{
			"invalid record, missing process name",
			NewRecord("name", sessionID, *NewProcess("", &Process_Address{"address"}, "method"), []RecordInput{*validRI}, []RecordOutput{*validRO}, nil),
			"invalid record process: missing required name",
			true,
		},
		{
			"invalid record, missing process process id",
			NewRecord("name", sessionID, *NewProcess("process_name", nil, "method"), []RecordInput{*validRI}, []RecordOutput{*validRO}, nil),
			"invalid record process: missing required process id",
			true,
		},
		{
			"invalid record, missing process method",
			NewRecord("name", sessionID, *NewProcess("process_name", &Process_Address{"address"}, ""), []RecordInput{*validRI}, []RecordOutput{*validRO}, nil),
			"invalid record process: missing required method",
			true,
		},
		{
			"invalid record, missing record input name",
			NewRecord("name", sessionID, *validPs,
				[]RecordInput{*NewRecordInput("", &RecordInput_Hash{"hash"}, "type_name", RecordInputStatus_Proposed)},
				[]RecordOutput{*validRO}, nil),
			"invalid record input: missing required name",
			true,
		},
		{
			"invalid record, missing record input source",
			NewRecord("name", sessionID, *validPs,
				[]RecordInput{*NewRecordInput("ri_name", nil, "type_name", RecordInputStatus_Proposed)},
				[]RecordOutput{*validRO}, nil),
			"invalid record input: missing required record input source",
			true,
		},
		{
			"invalid record, missing record input type name",
			NewRecord("name", sessionID, *validPs,
				[]RecordInput{*NewRecordInput("ri_name", &RecordInput_Hash{"hash"}, "", RecordInputStatus_Proposed)},
				[]RecordOutput{*validRO}, nil),
			"invalid record input: missing type name",
			true,
		},
		{
			"Invalid record, unknown record input status",
			NewRecord("name", sessionID, *validPs,
				[]RecordInput{*NewRecordInput("ri_name", &RecordInput_Hash{"hash"}, "type_name", RecordInputStatus_Unknown)},
				[]RecordOutput{*validRO}, nil),
			"invalid record input: invalid record input status, status unknown or missing",
			true,
		},
		{
			"Invalid record, missing record input hash",
			NewRecord("name", sessionID, *validPs,
				[]RecordInput{*NewRecordInput("ri_name", &RecordInput_Hash{""}, "type_name", RecordInputStatus_Proposed)},
				[]RecordOutput{*validRO}, nil),
			"invalid record input: missing required hash for proposed value",
			true,
		},
		{
			"Invalid record, incorrect status of record for record input source hash",
			NewRecord("name", sessionID, *validPs,
				[]RecordInput{*NewRecordInput("ri_name", &RecordInput_Hash{"hash"}, "type_name", RecordInputStatus_Record)},
				[]RecordOutput{*validRO}, nil),
			"invalid record input: hash specifier only applies to proposed inputs",
			true,
		},
		{
			"Invalid record, incorrect status of proposed for record input source record id",
			NewRecord("name", sessionID, *validPs,
				[]RecordInput{*NewRecordInput("name", &RecordInput_RecordId{recordID}, "type_name", RecordInputStatus_Proposed)},
				[]RecordOutput{*validRO}, nil),
			"invalid record input: record id must be used with Record type inputs",
			true,
		},
		{
			"Invalid record, incorrect status of unknown for record input source record id",
			NewRecord("name", sessionID, *validPs,
				[]RecordInput{*NewRecordInput("name", &RecordInput_RecordId{recordID}, "type_name", RecordInputStatus_Unknown)},
				[]RecordOutput{*validRO}, nil),
			"invalid record input: invalid record input status, status unknown or missing",
			true,
		},
		{
			"Invalid record, incorrect record id format of length 0 for record input",
			NewRecord("name", sessionID, *validPs,
				[]RecordInput{*NewRecordInput("name", &RecordInput_RecordId{}, "type_name", RecordInputStatus_Record)},
				[]RecordOutput{*validRO}, nil),
			"invalid record input: invalid record input recordid address is empty",
			true,
		},
		{
			"Invalid record, incorrect record id prefix for record input",
			NewRecord("name", sessionID, *validPs,
				[]RecordInput{*NewRecordInput("name", &RecordInput_RecordId{sessionID}, "type_name", RecordInputStatus_Record)},
				[]RecordOutput{*validRO}, nil),
			"invalid record input: invalid record id address (found session, expected record)",
			true,
		},
		{
			"Valid record, record input record id with proper prefix",
			NewRecord("name", sessionID, *validPs,
				[]RecordInput{*NewRecordInput("name", &RecordInput_RecordId{recordID}, "type_name", RecordInputStatus_Record)},
				[]RecordOutput{*validRO}, nil),
			"",
			false,
		},
		{
			"Invalid record, incorrect result status for record output",
			NewRecord("name", sessionID, *validPs,
				[]RecordInput{*validRI},
				[]RecordOutput{*NewRecordOutput("hash", ResultStatus_RESULT_STATUS_UNSPECIFIED)}, nil),
			"invalid record output: invalid record output status, status unspecified",
			true,
		},
		{
			"Invalid record, missing hash for record output",
			NewRecord("name", sessionID, *validPs,
				[]RecordInput{*validRI},
				[]RecordOutput{*NewRecordOutput("", ResultStatus_RESULT_STATUS_PASS)}, nil),
			"invalid record output: missing required hash",
			true,
		},
		{
			"Valid record, record output skip",
			NewRecord("name", sessionID, *validPs,
				[]RecordInput{*validRI},
				[]RecordOutput{*NewRecordOutput("", ResultStatus_RESULT_STATUS_SKIP)}, nil),
			"",
			false,
		},
	}

	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			err := tt.record.ValidateBasic()
			if (err != nil) != tt.wantErr {
				t.Errorf("Record ValidateBasic error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				require.Equal(t, tt.want, err.Error())
			}

		})
	}
}

func (s *ScopeTestSuite) TestSessionValidateBasic() {
	scopeUUID := uuid.New()
	sessionUUID := uuid.New()
	sessionID := SessionMetadataAddress(scopeUUID, sessionUUID)
	recordID := RecordMetadataAddress(scopeUUID, "test_record")
	contractSpec := ContractSpecMetadataAddress(uuid.New())
	tests := []struct {
		name    string
		session *Session
		want    string
		wantErr bool
	}{
		{
			"valid session",
			NewSession("name", sessionID, contractSpec, []Party{
				{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: PartyType_PARTY_TYPE_AFFILIATE}},
				&AuditFields{CreatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", CreatedDate: time.Now(),
					UpdatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", UpdatedDate: time.Now(),
					Message: "message",
				}),
			"",
			false,
		},
		{
			"valid session - no audit",
			NewSession("name", sessionID, contractSpec, []Party{
				{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: PartyType_PARTY_TYPE_AFFILIATE}}, nil),
			"",
			false,
		},
		{
			"invalid session, invalid prefix",
			NewSession("my_perfect_session", recordID, contractSpec, []Party{
				{Address: "invalidpartyaddress", Role: PartyType_PARTY_TYPE_CUSTODIAN}}, nil),
			"invalid session identifier (expected: session, got record)",
			true,
		},
		{
			"invalid session, invalid party address",
			NewSession("my_perfect_session", sessionID, contractSpec, []Party{
				{Address: "invalidpartyaddress", Role: PartyType_PARTY_TYPE_CUSTODIAN}}, nil),
			"invalid party on session: invalid address: decoding bech32 failed: invalid index of 1",
			true,
		},
		{
			"invalid session, invalid party type",
			NewSession("my_perfect_session", sessionID, contractSpec, []Party{
				{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: PartyType_PARTY_TYPE_UNSPECIFIED}}, nil),
			"invalid party on session: invalid party type;  party type not specified",
			true,
		},
		{
			"Invalid session, must have at least one party ",
			NewSession("my_perfect_session", sessionID, contractSpec, []Party{}, nil),
			"session must have at least one party",
			true,
		},
		{
			"invalid session, invalid spec id",
			NewSession("my_perfect_session", sessionID, recordID, []Party{
				{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: PartyType_PARTY_TYPE_AFFILIATE}}, nil),
			"invalid contract specification identifier (expected: contractspec, got record)",
			true,
		},
		{
			"Invalid session, max audit message length too long",
			NewSession("name", sessionID, contractSpec, []Party{
				{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: PartyType_PARTY_TYPE_AFFILIATE}},
				&AuditFields{CreatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", CreatedDate: time.Now(),
					UpdatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", UpdatedDate: time.Now(),
					Message: "messssssssssaaaaaaaaaage messssssssssaaaaaaaaaage messssssssssaaaaaaaaaage messssssssssaaaaaaaaaage messssssssssaaaaaaaaaage messssssssssaaaaaaaaaage messssssssssaaaaaaaaaage messssssssssaaaaaaaaaage  1",
				}),
			"session audit message exceeds maximum length (expected < 200 got: 202)",
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			err := tt.session.ValidateBasic()
			if (err != nil) != tt.wantErr {
				t.Errorf("Session ValidateBasic error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				require.Equal(t, tt.want, err.Error())
			}

		})
	}
}

func (s *ScopeTestSuite) TestSessionString() {

	scopeUUID := uuid.New()
	sessionUUID := uuid.New()
	sessionID := SessionMetadataAddress(scopeUUID, sessionUUID)
	contractSpec := ContractSpecMetadataAddress(uuid.New())
	session := NewSession("name", sessionID, contractSpec, []Party{
		{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: PartyType_PARTY_TYPE_AFFILIATE}},
		&AuditFields{CreatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck",
			Message: "message",
		})

	// println(session.String())
	s.T().Run("session string", func(t *testing.T) {
		require.Equal(t, fmt.Sprintf(`session_id: %s
specification_id: %s
parties:
- address: cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck
  role: 6
type: name
context: null
audit:
  created_by: cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck
  message: message
`, session.SessionId.String(), session.SpecificationId.String()),
			session.String())
	})
}

func (s *ScopeTestSuite) TestMetadataAuditUpdate() {
	blockTime := time.Now()
	var initial *AuditFields
	result := initial.UpdateAudit(blockTime, "creator", "initial")
	s.Equal(uint32(1), result.Version)
	s.Equal(blockTime, result.CreatedDate)
	s.Equal("creator", result.CreatedBy)
	s.Equal(time.Time{}, result.UpdatedDate)
	s.Equal("", result.UpdatedBy)
	s.Equal("initial", result.Message)

	auditTime := time.Now()
	initial = &AuditFields{CreatedDate: auditTime, CreatedBy: "creator", Version: 1}
	result = initial.UpdateAudit(blockTime, "updater", "")
	s.Equal(uint32(2), result.Version)
	s.Equal(auditTime, result.CreatedDate)
	s.Equal("creator", result.CreatedBy)
	s.Equal(blockTime, result.UpdatedDate)
	s.Equal("updater", result.UpdatedBy)
	s.Equal("", result.Message)
}
