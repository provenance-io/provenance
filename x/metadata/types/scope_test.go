package types

import (
	"encoding/hex"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	pubHex, _ = hex.DecodeString("85EA54E8598B27EC37EAEEEEA44F1E78A9B5E671")
	addr      = sdk.AccAddress(pubHex)
)

type scopeTestSuite struct {
	suite.Suite
}

// func ownerPartyList is defined in msg_test.go

func TestAddressTestSuite(t *testing.T) {
	suite.Run(t, new(scopeTestSuite))
}

func (s *scopeTestSuite) SetupSuite() {
	s.T().Parallel()
}

func (s *scopeTestSuite) TestScopeValidateBasic() {
	tests := []struct {
		name    string
		scope   *Scope
		want    string
		wantErr bool
	}{
		{
			"valid scope one owner",
			NewScope(ScopeMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), ownerPartyList(addr.String()), []string{}, ""),
			"",
			false,
		},
		{
			"valid scope one owner, one data access",
			NewScope(ScopeMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), ownerPartyList(addr.String()), []string{addr.String()}, ""),
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
			NewScope(ScopeMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), []Party{}, []string{addr.String()}, ""),
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
			NewScope(MetadataAddress(addr), ScopeSpecMetadataAddress(uuid.New()), []Party{}, []string{}, ""),
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
			NewScope(ScopeMetadataAddress(uuid.New()), MetadataAddress(addr), []Party{}, []string{}, ""),
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

func (s *scopeTestSuite) TestScopeString() {
	s.T().Run("scope string", func(t *testing.T) {
		scopeUUID := uuid.MustParse("8d80b25a-c089-4446-956e-5d08cfe3e1a5")
		groupUUID := uuid.MustParse("c25c7bd4-c639-4367-a842-f64fa5fccc19")
		scope := NewScope(ScopeMetadataAddress(
			scopeUUID), ScopeSpecMetadataAddress(groupUUID),
			ownerPartyList(addr.String()),
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

func (s *scopeTestSuite) TestRecordValidateBasic() {
	scopeUUID := uuid.New()
	groupUUID := uuid.New()
	groupId := GroupMetadataAddress(scopeUUID, groupUUID)
	recordId := RecordMetadataAddress(scopeUUID, "test_record")
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
			NewRecord("name", groupId, *validPs, []RecordInput{*validRI}, []RecordOutput{*validRO}),
			"",
			false,
		},
		{
			"invalid record, invalid/missing name for record",
			NewRecord("", groupId, *validPs, []RecordInput{*validRI}, []RecordOutput{*validRO}),
			"invalid/missing name for record",
			true,
		},
		{
			"invalid record, missing groupid",
			NewRecord("name", nil, *validPs, []RecordInput{*validRI}, []RecordOutput{*validRO}),
			"incorrect address length (must be at least 17, actual: 0)",
			true,
		},
		{
			"invalid record, missing process name",
			NewRecord("name", groupId, *NewProcess("", &Process_Address{"address"}, "method"), []RecordInput{*validRI}, []RecordOutput{*validRO}),
			"invalid record process: missing required name",
			true,
		},
		{
			"invalid record, missing process process id",
			NewRecord("name", groupId, *NewProcess("process_name", nil, "method"), []RecordInput{*validRI}, []RecordOutput{*validRO}),
			"invalid record process: missing required process id",
			true,
		},
		{
			"invalid record, missing process method",
			NewRecord("name", groupId, *NewProcess("process_name", &Process_Address{"address"}, ""), []RecordInput{*validRI}, []RecordOutput{*validRO}),
			"invalid record process: missing required method",
			true,
		},
		{
			"invalid record, missing record input name",
			NewRecord("name", groupId, *validPs,
				[]RecordInput{*NewRecordInput("", &RecordInput_Hash{"hash"}, "type_name", RecordInputStatus_Proposed)},
				[]RecordOutput{*validRO}),
			"invalid record input: missing required name",
			true,
		},
		{
			"invalid record, missing record input source",
			NewRecord("name", groupId, *validPs,
				[]RecordInput{*NewRecordInput("ri_name", nil, "type_name", RecordInputStatus_Proposed)},
				[]RecordOutput{*validRO}),
			"invalid record input: missing required record input source",
			true,
		},
		{
			"invalid record, missing record input type name",
			NewRecord("name", groupId, *validPs,
				[]RecordInput{*NewRecordInput("ri_name", &RecordInput_Hash{"hash"}, "", RecordInputStatus_Proposed)},
				[]RecordOutput{*validRO}),
			"invalid record input: missing type name",
			true,
		},
		{
			"Invalid record, unknown record input status",
			NewRecord("name", groupId, *validPs,
				[]RecordInput{*NewRecordInput("ri_name", &RecordInput_Hash{"hash"}, "type_name", RecordInputStatus_Unknown)},
				[]RecordOutput{*validRO}),
			"invalid record input: invalid record input status, status unknown or missing",
			true,
		},
		{
			"Invalid record, missing record input hash",
			NewRecord("name", groupId, *validPs,
				[]RecordInput{*NewRecordInput("ri_name", &RecordInput_Hash{""}, "type_name", RecordInputStatus_Proposed)},
				[]RecordOutput{*validRO}),
			"invalid record input: missing required hash for proposed value",
			true,
		},
		{
			"Invalid record, incorrect status of record for record input source hash",
			NewRecord("name", groupId, *validPs,
				[]RecordInput{*NewRecordInput("ri_name", &RecordInput_Hash{"hash"}, "type_name", RecordInputStatus_Record)},
				[]RecordOutput{*validRO}),
			"invalid record input: hash specifier only applies to proposed inputs",
			true,
		},
		{
			"Invalid record, incorrect status of proposed for record input source record id",
			NewRecord("name", groupId, *validPs,
				[]RecordInput{*NewRecordInput("name", &RecordInput_RecordId{recordId}, "type_name", RecordInputStatus_Proposed)},
				[]RecordOutput{*validRO}),
			"invalid record input: record id must be used with Record type inputs",
			true,
		},
		{
			"Invalid record, incorrect status of unknown for record input source record id",
			NewRecord("name", groupId, *validPs,
				[]RecordInput{*NewRecordInput("name", &RecordInput_RecordId{recordId}, "type_name", RecordInputStatus_Unknown)},
				[]RecordOutput{*validRO}),
			"invalid record input: invalid record input status, status unknown or missing",
			true,
		},
		{
			"Invalid record, incorrect record id format of length 0 for record input",
			NewRecord("name", groupId, *validPs,
				[]RecordInput{*NewRecordInput("name", &RecordInput_RecordId{}, "type_name", RecordInputStatus_Record)},
				[]RecordOutput{*validRO}),
			"invalid record input: invalid record input recordid incorrect address length (must be at least 17, actual: 0)",
			true,
		},
		{
			"Invalid record, incorrect record id prefix for record input",
			NewRecord("name", groupId, *validPs,
				[]RecordInput{*NewRecordInput("name", &RecordInput_RecordId{groupId}, "type_name", RecordInputStatus_Record)},
				[]RecordOutput{*validRO}),
			"invalid record input: invalid record id address (found group, expected record)",
			true,
		},
		{
			"Valid record, record input record id with proper prefix",
			NewRecord("name", groupId, *validPs,
				[]RecordInput{*NewRecordInput("name", &RecordInput_RecordId{recordId}, "type_name", RecordInputStatus_Record)},
				[]RecordOutput{*validRO}),
			"",
			false,
		},
		{
			"Invalid record, incorrect result status for record output",
			NewRecord("name", groupId, *validPs,
				[]RecordInput{*validRI},
				[]RecordOutput{*NewRecordOutput("hash", ResultStatus_RESULT_STATUS_UNSPECIFIED)}),
			"invalid record output: invalid record output status, status unspecified",
			true,
		},
		{
			"Invalid record, missing hash for record output",
			NewRecord("name", groupId, *validPs,
				[]RecordInput{*validRI},
				[]RecordOutput{*NewRecordOutput("", ResultStatus_RESULT_STATUS_PASS)}),
			"invalid record output: missing required hash",
			true,
		},
		{
			"Valid record, record output skip",
			NewRecord("name", groupId, *validPs,
				[]RecordInput{*validRI},
				[]RecordOutput{*NewRecordOutput("", ResultStatus_RESULT_STATUS_SKIP)}),
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
