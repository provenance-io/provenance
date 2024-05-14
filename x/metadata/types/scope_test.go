package types

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ScopeTestSuite struct {
	suite.Suite

	Addr string
}

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

func OwnerPartyList(addresses ...string) []Party {
	retval := make([]Party, len(addresses))
	for i, addr := range addresses {
		retval[i] = Party{Address: addr, Role: PartyType_PARTY_TYPE_OWNER}
	}
	return retval
}

func (s *ScopeTestSuite) TestScopeValidateBasic() {
	ns := func(scopeID, scopeSpecification MetadataAddress, owners []Party, dataAccess []string, valueOwner string) *Scope {
		return &Scope{
			ScopeId:            scopeID,
			SpecificationId:    scopeSpecification,
			Owners:             owners,
			DataAccess:         dataAccess,
			ValueOwnerAddress:  valueOwner,
			RequirePartyRollup: false,
		}
	}
	tests := []struct {
		name  string
		scope *Scope
		want  string
	}{
		{
			"valid scope one owner",
			ns(ScopeMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), OwnerPartyList(s.Addr), []string{}, ""),
			"",
		},
		{
			"valid scope one owner, one data access",
			ns(ScopeMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), OwnerPartyList(s.Addr), []string{s.Addr}, ""),
			"",
		},
		{
			"no owners",
			ns(ScopeMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), []Party{}, []string{}, ""),
			"invalid scope owners: at least one party is required",
		},
		{
			"no owners, data access",
			ns(ScopeMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), []Party{}, []string{s.Addr}, ""),
			"invalid scope owners: at least one party is required",
		},
		{
			"invalid scope id",
			ns(ScopeSpecMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), []Party{}, []string{}, ""),
			"invalid scope identifier (expected: scope, got scopespec)",
		},
		{
			"invalid scope id - wrong address type",
			ns(MetadataAddress{0x85}, ScopeSpecMetadataAddress(uuid.New()), []Party{}, []string{}, ""),
			"invalid metadata address type: 133",
		},
		{
			"invalid spec id",
			ns(ScopeMetadataAddress(uuid.New()), ScopeMetadataAddress(uuid.New()), []Party{}, []string{}, ""),
			"invalid scope specification identifier (expected: scopespec, got scope)",
		},
		{
			"invalid spec id - wrong address type",
			ns(ScopeMetadataAddress(uuid.New()), MetadataAddress{0x85}, []Party{}, []string{}, ""),
			"invalid metadata address type: 133",
		},
		{
			"invalid owner on scope",
			ns(ScopeMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), OwnerPartyList(":invalid"), []string{}, ""),
			"invalid scope owners: invalid party address [:invalid]: decoding bech32 failed: invalid separator index -1",
		},
		{
			name: "not rollup with optional party",
			scope: &Scope{
				ScopeId:         ScopeMetadataAddress(uuid.New()),
				SpecificationId: ScopeSpecMetadataAddress(uuid.New()),
				Owners: []Party{
					{
						Address:  sdk.AccAddress("just_a_test_________").String(),
						Role:     PartyType_PARTY_TYPE_SERVICER,
						Optional: true,
					},
				},
				DataAccess:         nil,
				ValueOwnerAddress:  "",
				RequirePartyRollup: false,
			},
			want: "parties can only be optional when require_party_rollup = true",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := tc.scope.ValidateBasic()
			if len(tc.want) > 0 {
				s.Assert().EqualError(err, tc.want, "ValidateBasic")
			} else {
				s.Assert().NoError(err, "ValidateBasic")
			}
		})
	}
}

func (s *ScopeTestSuite) TestScopeAddAccess() {
	ns := func(scopeID, scopeSpecification MetadataAddress, owners []Party, dataAccess []string, valueOwner string) *Scope {
		return &Scope{
			ScopeId:            scopeID,
			SpecificationId:    scopeSpecification,
			Owners:             owners,
			DataAccess:         dataAccess,
			ValueOwnerAddress:  valueOwner,
			RequirePartyRollup: false,
		}
	}

	tests := []struct {
		name       string
		scope      *Scope
		dataAccess []string
		expected   []string
	}{
		{
			"should successfully add new address to scope data access",
			ns(ScopeMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), OwnerPartyList(s.Addr), []string{}, ""),
			[]string{"addr1"},
			[]string{"addr1"},
		},
		{
			"should successfully not add same address twice to data access",
			ns(ScopeMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), OwnerPartyList(s.Addr), []string{"addr1"}, ""),
			[]string{"addr1"},
			[]string{"addr1"},
		},
		{
			"should successfully add new address to data access",
			ns(ScopeMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), OwnerPartyList(s.Addr), []string{"addr1"}, ""),
			[]string{"addr2"},
			[]string{"addr1", "addr2"},
		},
		{
			"should successfully add new address only once to data access",
			ns(ScopeMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), OwnerPartyList(s.Addr), []string{"addr1"}, ""),
			[]string{"addr2", "addr2", "addr2"},
			[]string{"addr1", "addr2"},
		},
	}

	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {

			tt.scope.AddDataAccess(tt.dataAccess)
			require.Equal(t, tt.scope.DataAccess, tt.expected)
		})
	}
}

func (s *ScopeTestSuite) TestScopeRemoveAccess() {
	ns := func(scopeID, scopeSpecification MetadataAddress, owners []Party, dataAccess []string, valueOwner string) *Scope {
		return &Scope{
			ScopeId:            scopeID,
			SpecificationId:    scopeSpecification,
			Owners:             owners,
			DataAccess:         dataAccess,
			ValueOwnerAddress:  valueOwner,
			RequirePartyRollup: false,
		}
	}

	tests := []struct {
		name       string
		scope      *Scope
		dataAccess []string
		expected   []string
	}{
		{
			"should successfully remove address from scope data access",
			ns(ScopeMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), OwnerPartyList(s.Addr), []string{"addr1"}, ""),
			[]string{"addr1"},
			[]string{},
		},
		{
			"should successfully remove from a list more with more than one",
			ns(ScopeMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), OwnerPartyList(s.Addr), []string{"addr1", "addr2"}, ""),
			[]string{"addr2"},
			[]string{"addr1"},
		},
		{
			"should successfully remove nothing",
			ns(ScopeMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), OwnerPartyList(s.Addr), []string{}, ""),
			[]string{"addr2"},
			[]string{},
		},
		{
			"should successfully remove address even when repeated in list",
			ns(ScopeMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), OwnerPartyList(s.Addr), []string{"addr1", "addr2", "addr3"}, ""),
			[]string{"addr2", "addr2", "addr2"},
			[]string{"addr1", "addr3"},
		},
	}

	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {

			tt.scope.RemoveDataAccess(tt.dataAccess)
			require.Equal(t, tt.scope.DataAccess, tt.expected)
		})
	}
}

func (s *ScopeTestSuite) TestScopeAddOwners() {
	ns := func(scopeID, scopeSpecification MetadataAddress, owners []Party, dataAccess []string, valueOwner string) *Scope {
		return &Scope{
			ScopeId:            scopeID,
			SpecificationId:    scopeSpecification,
			Owners:             owners,
			DataAccess:         dataAccess,
			ValueOwnerAddress:  valueOwner,
			RequirePartyRollup: false,
		}
	}

	user1Owner := Party{Address: s.Addr, Role: PartyType_PARTY_TYPE_OWNER}
	user1Investor := Party{Address: s.Addr, Role: PartyType_PARTY_TYPE_INVESTOR}
	user2Affiliate := Party{Address: "addr2", Role: PartyType_PARTY_TYPE_AFFILIATE}
	tests := []struct {
		name     string
		scope    *Scope
		owners   []Party
		expected []Party
		errMsg   string
	}{
		{
			"should successfully update owner address with new role",
			ns(ScopeMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), []Party{user1Owner}, []string{}, ""),
			[]Party{user1Investor},
			[]Party{user1Owner, user1Investor},
			"",
		},
		{
			"should fail to add duplicate owner",
			ns(ScopeMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), []Party{user1Owner}, []string{"addr1"}, ""),
			[]Party{user1Owner},
			[]Party{user1Owner},
			"party already exists with address cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck and role PARTY_TYPE_OWNER",
		},
		{
			"should successfully add new address to owners",
			ns(ScopeMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), []Party{user1Owner}, []string{"addr1"}, ""),
			[]Party{user2Affiliate},
			[]Party{user1Owner, user2Affiliate},
			"",
		},
		{
			"should successfully not change the list",
			ns(ScopeMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), []Party{user1Owner}, []string{"addr1"}, ""),
			[]Party{},
			[]Party{user1Owner},
			"",
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			err := tc.scope.AddOwners(tc.owners)
			if len(tc.errMsg) > 0 {
				require.EqualError(t, err, tc.errMsg, "AddOwners expected error")
			} else {
				require.NoError(t, err, "AddOwners unexpected error")
			}
			require.Equal(t, tc.expected, tc.scope.Owners, "new scope owners value")
		})
	}
}

func (s *ScopeTestSuite) TestScopeRemoveOwners() {
	ns := func(scopeID, scopeSpecification MetadataAddress, owners []Party, dataAccess []string, valueOwner string) *Scope {
		return &Scope{
			ScopeId:            scopeID,
			SpecificationId:    scopeSpecification,
			Owners:             owners,
			DataAccess:         dataAccess,
			ValueOwnerAddress:  valueOwner,
			RequirePartyRollup: false,
		}
	}

	user1Owner := OwnerPartyList(s.Addr)
	user1Investor := Party{Address: s.Addr, Role: PartyType_PARTY_TYPE_INVESTOR}
	user2Affiliate := Party{Address: "addr2", Role: PartyType_PARTY_TYPE_AFFILIATE}
	tests := []struct {
		name     string
		scope    *Scope
		owners   []string
		expected []Party
		errMsg   string
	}{
		{
			"should successfully remove owner by address",
			ns(ScopeMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), user1Owner, []string{}, ""),
			[]string{user1Owner[0].Address},
			[]Party{},
			"",
		},
		{
			"should fail to remove any non-existent owner",
			ns(ScopeMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), user1Owner, []string{"addr1"}, ""),
			[]string{"notanowner"},
			user1Owner,
			"address does not exist in scope owners: notanowner",
		},
		{
			"should successfully remove owner from list of multiple",
			ns(ScopeMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), []Party{user1Investor, user2Affiliate}, []string{"addr1"}, ""),
			[]string{user1Investor.Address},
			[]Party{user2Affiliate},
			"",
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			err := tc.scope.RemoveOwners(tc.owners)
			if len(tc.errMsg) > 0 {
				require.EqualError(t, err, tc.errMsg, "RemoveOwners expected error")
			} else {
				require.NoError(t, err, "RemoveOwners unexpected error")
				require.Equal(t, tc.scope.Owners, tc.expected, "new scope owners value")
			}
		})
	}
}

func (s *ScopeTestSuite) TestScopeHasOwnerAddress() {
	pt := func(address string) Party {
		return Party{Address: address}
	}
	ptz := func(parties ...Party) []Party {
		rv := make([]Party, 0, len(parties))
		rv = append(rv, parties...)
		return rv
	}
	tests := []struct {
		name    string
		scope   Scope
		address string
		exp     bool
	}{
		{
			name:    "nil owners",
			scope:   Scope{Owners: nil},
			address: "",
			exp:     false,
		},
		{
			name:    "empty owners",
			scope:   Scope{Owners: []Party{}},
			address: "",
			exp:     false,
		},
		{
			name:    "one owner same address",
			scope:   Scope{Owners: ptz(pt("abc"))},
			address: "abc",
			exp:     true,
		},
		{
			name:    "one owner address has extra space at start",
			scope:   Scope{Owners: ptz(pt("abc"))},
			address: " abc",
			exp:     false,
		},
		{
			name:    "one owner address has extra space at end",
			scope:   Scope{Owners: ptz(pt("abc"))},
			address: "abc ",
			exp:     false,
		},
		{
			name:    "one owner address start same but is shorter",
			scope:   Scope{Owners: ptz(pt("abc"))},
			address: "ab",
			exp:     false,
		},
		{
			name:    "one owner address start same but is longer",
			scope:   Scope{Owners: ptz(pt("abc"))},
			address: "abcd",
			exp:     false,
		},
		{
			name:    "one owner address end same but is shorter",
			scope:   Scope{Owners: ptz(pt("abc"))},
			address: "bc",
			exp:     false,
		},
		{
			name:    "one owner address end same but is longer",
			scope:   Scope{Owners: ptz(pt("abc"))},
			address: "aabc",
			exp:     false,
		},
		{
			name:    "one owner address different case",
			scope:   Scope{Owners: ptz(pt("abc"))},
			address: "aBc",
			exp:     false,
		},
		{
			name:    "three owners no match",
			scope:   Scope{Owners: ptz(pt("aaa"), pt("bbb"), pt("ccc"))},
			address: "abc",
			exp:     false,
		},
		{
			name:    "three owners first matches",
			scope:   Scope{Owners: ptz(pt("aaa"), pt("bbb"), pt("ccc"))},
			address: "aaa",
			exp:     true,
		},
		{
			name:    "three owners second matches",
			scope:   Scope{Owners: ptz(pt("aaa"), pt("bbb"), pt("ccc"))},
			address: "bbb",
			exp:     true,
		},
		{
			name:    "three owners third matches",
			scope:   Scope{Owners: ptz(pt("aaa"), pt("bbb"), pt("ccc"))},
			address: "ccc",
			exp:     true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			actual := tc.scope.hasOwnerAddress(tc.address)
			s.Assert().Equal(tc.exp, actual, "hasOwnerAddress(%q) on %#v", tc.address, tc.scope)
		})
	}
}

func (s *ScopeTestSuite) TestScopeString() {
	scopeUUID := uuid.MustParse("8d80b25a-c089-4446-956e-5d08cfe3e1a5")
	specUUID := uuid.MustParse("c25c7bd4-c639-4367-a842-f64fa5fccc19")
	scope := &Scope{
		ScopeId:         ScopeMetadataAddress(scopeUUID),
		SpecificationId: ScopeSpecMetadataAddress(specUUID),
		Owners:          OwnerPartyList(s.Addr),
		DataAccess:      []string{},
	}
	expected := "&Scope{" +
		"ScopeId:scope1qzxcpvj6czy5g354dews3nlruxjsahhnsp," +
		"SpecificationId:scopespec1qnp9c775ccu5xeaggtmylf0uesvsqyrkq8," +
		"Owners:[]Party{cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck - PARTY_TYPE_OWNER,}," +
		"DataAccess:[]," +
		"ValueOwnerAddress:," +
		"RequirePartyRollup:false," +
		"}"
	var actual string
	testFunc := func() {
		actual = scope.String()
	}
	s.Require().NotPanics(testFunc, "scope.String()")
	s.Assert().Equal(expected, actual, "scope.String() result")
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
		errMsg  string
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
		},
		{
			"valid session - no audit",
			NewSession("name", sessionID, contractSpec, []Party{
				{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: PartyType_PARTY_TYPE_AFFILIATE}}, nil),
			"",
		},
		{
			"invalid session, invalid prefix",
			NewSession("my_perfect_session", recordID, contractSpec, []Party{
				{Address: "invalidpartyaddress", Role: PartyType_PARTY_TYPE_CUSTODIAN}}, nil),
			"invalid session identifier (expected: session, got record)",
		},
		{
			"invalid session, invalid party address",
			NewSession("my_perfect_session", sessionID, contractSpec, []Party{
				{Address: "invalidpartyaddress", Role: PartyType_PARTY_TYPE_CUSTODIAN}}, nil),
			"invalid party address [invalidpartyaddress]: decoding bech32 failed: invalid separator index -1",
		},
		{
			"invalid session, invalid party type",
			NewSession("my_perfect_session", sessionID, contractSpec, []Party{
				{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: PartyType_PARTY_TYPE_UNSPECIFIED}}, nil),
			"invalid party type for party cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck",
		},
		{
			"Invalid session, must have at least one party ",
			NewSession("my_perfect_session", sessionID, contractSpec, []Party{}, nil),
			"at least one party is required",
		},
		{
			"invalid session, invalid spec id",
			NewSession("my_perfect_session", sessionID, recordID, []Party{
				{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: PartyType_PARTY_TYPE_AFFILIATE}}, nil),
			"invalid contract specification identifier (expected: contractspec, got record)",
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
		},
	}

	for _, tc := range tests {
		tc := tc
		s.T().Run(tc.name, func(t *testing.T) {
			err := tc.session.ValidateBasic()
			if len(tc.errMsg) > 0 {
				require.EqualError(t, err, tc.errMsg, "Session.ValidateBasic expected error")
			} else {
				require.NoError(t, err, "Session.ValidateBasic unexpected error")
			}
		})
	}
}

func (s *ScopeTestSuite) TestSessionString() {
	scopeUUID := uuid.MustParse("382b6eed-e61e-469b-933e-e4d0210c77f6")
	sessionUUID := uuid.MustParse("c120ba2a-b13d-4678-8e5a-7459d92c3e90")
	sessionID := SessionMetadataAddress(scopeUUID, sessionUUID)
	contractSpecUUID := uuid.MustParse("888B0DAB-61ED-4074-B9B6-0F35CDFA4F0D")
	contractSpecID := ContractSpecMetadataAddress(contractSpecUUID)

	session := &Session{
		SessionId:       sessionID,
		SpecificationId: contractSpecID,
		Parties: []Party{
			{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: PartyType_PARTY_TYPE_AFFILIATE},
		},
		Name:    "whatever",
		Context: []byte("moredata"),
		Audit: &AuditFields{
			CreatedDate: time.Unix(0, 0),
			CreatedBy:   "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck",
			UpdatedDate: time.Unix(0, 0),
			Message:     "message",
		},
	}

	expected := "&Session{" +
		"SessionId:session1qyuzkmhduc0ydxun8mjdqggvwlmvzg9692cn63nc3ed8gkwe9slfq747zd8," +
		"SpecificationId:contractspec1qwygkrdtv8k5qa9ekc8ntn06fuxsj02zcg," +
		"Parties:[]Party{cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck - PARTY_TYPE_AFFILIATE,}," +
		"Name:whatever," +
		"Context:[109 111 114 101 100 97 116 97]," +
		"Audit:created_date:<> created_by:\"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck\" updated_date:<> message:\"message\" ," +
		"}"

	var actual string
	testFunc := func() {
		actual = session.String()
	}
	s.Require().NotPanics(testFunc, "session.String()")
	s.Assert().Equal(expected, actual, "session.String() result")
}

func (s *ScopeTestSuite) TestRecordString() {
	scopeUUID := uuid.MustParse("382b6eed-e61e-469b-933e-e4d0210c77f6")
	sessionUUID := uuid.MustParse("c120ba2a-b13d-4678-8e5a-7459d92c3e90")
	sessionID := SessionMetadataAddress(scopeUUID, sessionUUID)
	contractSpecUUID := uuid.MustParse("888B0DAB-61ED-4074-B9B6-0F35CDFA4F0D")
	name := "something"
	recSpecID := RecordSpecMetadataAddress(contractSpecUUID, name)
	otherUUID := uuid.MustParse("37b553b7-c36e-4e4a-9e72-aef0b3689ce8")

	record := &Record{
		Name:      name,
		SessionId: sessionID,
		Process: Process{
			ProcessId: &Process_Hash{Hash: "5075140835d0bc504791c76b04c33d2b"},
			Name:      "runner",
			Method:    "UsainBolt",
		},
		Inputs: []RecordInput{
			{
				Name:     "incoming",
				Source:   &RecordInput_Hash{Hash: "d48f944ac6c78b97d544f98b89b506ca"},
				TypeName: "thingy",
				Status:   RecordInputStatus_Record,
			},
			{
				Name:     "notacomeback",
				Source:   &RecordInput_RecordId{RecordId: RecordMetadataAddress(otherUUID, "beenhere")},
				TypeName: "widget",
				Status:   RecordInputStatus_Proposed,
			},
		},
		Outputs: []RecordOutput{
			{
				Hash:   "fba028e9ebb6ae55787d676995306e06",
				Status: ResultStatus_RESULT_STATUS_PASS,
			},
		},
		SpecificationId: recSpecID,
	}

	expected := "&Record{" +
		"Name:something," +
		"SessionId:session1qyuzkmhduc0ydxun8mjdqggvwlmvzg9692cn63nc3ed8gkwe9slfq747zd8," +
		"Process:runner - UsainBolt - {5075140835d0bc504791c76b04c33d2b}," +
		"Inputs:[]RecordInput{" +
		"incoming (thingy) - RECORD_INPUT_STATUS_RECORD d48f944ac6c78b97d544f98b89b506ca," +
		"notacomeback (widget) - RECORD_INPUT_STATUS_PROPOSED record1qgmm25ahcdhyuj57w2h0pvmgnn5v78sxmz4r9998fpk9zhc9yzexw8p9exn," +
		"}," +
		"Outputs:[]RecordOutput{fba028e9ebb6ae55787d676995306e06 - RESULT_STATUS_PASS,}," +
		"SpecificationId:recspec1qkygkrdtv8k5qa9ekc8ntn06fuxnljdk39ze6uu03jy28fy2483n25phf7m," +
		"}"

	var actual string
	testFunc := func() {
		actual = record.String()
	}
	s.Require().NotPanics(testFunc, "record.String()")
	s.Assert().Equal(expected, actual, "record.String() result")
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

func (s *ScopeTestSuite) TestParty_ValidateBasic() {
	addr := sdk.AccAddress("test").String()

	tests := []struct {
		name   string
		party  Party
		expErr string
	}{
		{
			name:   "good addr good party optional",
			party:  Party{Address: addr, Role: 3, Optional: true},
			expErr: "",
		},
		{
			name:   "good addr good party not optional",
			party:  Party{Address: addr, Role: 3, Optional: false},
			expErr: "",
		},
		{
			name:   "no address",
			party:  Party{Address: ""},
			expErr: "missing party address",
		},
		{
			name:   "bad address",
			party:  Party{Address: "badbad"},
			expErr: "invalid party address [badbad]: decoding bech32 failed: invalid bech32 string length 6",
		},
		{
			name:   "negative party type",
			party:  Party{Address: addr, Role: -10},
			expErr: "invalid party type for party " + addr,
		},
		{
			name:   "large party type",
			party:  Party{Address: addr, Role: 123},
			expErr: "invalid party type for party " + addr,
		},
		{
			name:   "unspecified party type",
			party:  Party{Address: addr, Role: PartyType_PARTY_TYPE_UNSPECIFIED},
			expErr: "invalid party type for party " + addr,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			orig := tc.party
			var err error
			testFunc := func() {
				err = tc.party.ValidateBasic()
			}
			s.Require().NotPanics(testFunc, "ValidateBasic")
			if len(tc.expErr) > 0 {
				s.Assert().EqualError(err, tc.expErr, "ValidateBasic")
			} else {
				s.Assert().NoError(err, "ValidateBasic")
			}
			s.Assert().Equal(orig, tc.party, "party after ValidateBasic")
		})
	}
}

func (s *ScopeTestSuite) TestValidateOptionalParties() {
	tests := []struct {
		name       string
		optAllowed bool
		parties    []Party
		expErr     string
	}{
		{
			name:       "opt allowed nil parties",
			optAllowed: true,
			parties:    nil,
			expErr:     "",
		},
		{
			name:       "opt allowed empty parties",
			optAllowed: true,
			parties:    []Party{},
			expErr:     "",
		},
		{
			name:       "opt not allowed nil parties",
			optAllowed: false,
			parties:    nil,
			expErr:     "",
		},
		{
			name:       "opt not allowed empty parties",
			optAllowed: false,
			parties:    []Party{},
			expErr:     "",
		},
		{
			name:       "opt allowed 1 party required",
			optAllowed: true,
			parties:    []Party{{Optional: false}},
			expErr:     "",
		},
		{
			name:       "opt allowed 1 party optional",
			optAllowed: true,
			parties:    []Party{{Optional: true}},
			expErr:     "",
		},
		{
			name:       "opt not allowed 1 party required",
			optAllowed: false,
			parties:    []Party{{Optional: false}},
			expErr:     "",
		},
		{
			name:       "opt not allowed 1 party optional",
			optAllowed: false,
			parties:    []Party{{Optional: true}},
			expErr:     "parties can only be optional when require_party_rollup = true",
		},
		{
			name:       "opt allowed 2 parties req req",
			optAllowed: true,
			parties:    []Party{{Optional: false}, {Optional: false}},
			expErr:     "",
		},
		{
			name:       "opt allowed 2 parties opt req",
			optAllowed: true,
			parties:    []Party{{Optional: true}, {Optional: false}},
			expErr:     "",
		},
		{
			name:       "opt allowed 2 parties req opt",
			optAllowed: true,
			parties:    []Party{{Optional: false}, {Optional: true}},
			expErr:     "",
		},
		{
			name:       "opt allowed 2 parties opt opt",
			optAllowed: true,
			parties:    []Party{{Optional: true}, {Optional: true}},
			expErr:     "",
		},
		{
			name:       "opt not allowed 2 parties req req",
			optAllowed: false,
			parties:    []Party{{Optional: false}, {Optional: false}},
			expErr:     "",
		},
		{
			name:       "opt not allowed 2 parties opt req",
			optAllowed: false,
			parties:    []Party{{Optional: true}, {Optional: false}},
			expErr:     "parties can only be optional when require_party_rollup = true",
		},
		{
			name:       "opt not allowed 2 parties req opt",
			optAllowed: false,
			parties:    []Party{{Optional: false}, {Optional: true}},
			expErr:     "parties can only be optional when require_party_rollup = true",
		},
		{
			name:       "opt not allowed 2 parties opt opt",
			optAllowed: false,
			parties:    []Party{{Optional: true}, {Optional: true}},
			expErr:     "parties can only be optional when require_party_rollup = true",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := ValidateOptionalParties(tc.optAllowed, tc.parties)
			if len(tc.expErr) > 0 {
				s.Assert().EqualError(err, tc.expErr, "ValidateOptionalParties")
			} else {
				s.Assert().NoError(err, "ValidateOptionalParties")
			}
		})
	}
}

func (s *ScopeTestSuite) TestValidatePartiesAreUnique() {
	oneCapParty := make([]Party, 1, 1)
	oneCapParty[0] = Party{Address: "one", Role: 1}

	tests := []struct {
		name    string
		parties []Party
		expErr  string
	}{
		{
			name:    "nil",
			parties: nil,
			expErr:  "",
		},
		{
			name:    "empty",
			parties: []Party{},
			expErr:  "",
		},
		{
			name:    "one party",
			parties: []Party{{Address: "abc", Role: 7}},
			expErr:  "",
		},
		{
			name:    "two parties diff addr",
			parties: []Party{{Address: "abc", Role: 3}, {Address: "def", Role: 3}},
			expErr:  "",
		},
		{
			name:    "two parties diff role",
			parties: []Party{{Address: "abc", Role: 3}, {Address: "abc", Role: 4}},
			expErr:  "",
		},
		{
			name:    "two parties diff optional",
			parties: []Party{{Address: "abc", Role: 3, Optional: false}, {Address: "abc", Role: 3, Optional: true}},
			expErr:  "duplicate parties not allowed: address = abc, role = INVESTOR, indexes: 0, 1",
		},
		{
			name: "five parties unique",
			parties: []Party{
				{Address: "abc", Role: 1},
				{Address: "def", Role: 2},
				{Address: "ghi", Role: 3},
				{Address: "abc", Role: 4},
				{Address: "jkl", Role: 5},
			},
			expErr: "",
		},
		{
			name: "five parties not unique",
			parties: []Party{
				{Address: "abc", Role: 1},
				{Address: "def", Role: 2},
				{Address: "ghi", Role: 3},
				{Address: "def", Role: 2},
				{Address: "jkl", Role: 5},
			},
			expErr: "duplicate parties not allowed: address = def, role = SERVICER, indexes: 1, 3",
		},
		{
			name:    "one party with capacity = 1",
			parties: oneCapParty,
			expErr:  "",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var orig []Party
			if tc.parties != nil {
				orig = make([]Party, 0, len(tc.parties))
				orig = append(orig, tc.parties...)
			}
			var err error
			testFunc := func() {
				err = ValidatePartiesAreUnique(tc.parties)
			}
			s.Require().NotPanics(testFunc, "ValidatePartiesAreUnique")
			if len(tc.expErr) > 0 {
				s.Assert().EqualError(err, tc.expErr, "ValidatePartiesAreUnique")
			} else {
				s.Assert().NoError(err, tc.expErr, "ValidatePartiesAreUnique")
			}
			s.Assert().Equal(orig, tc.parties, "parties slice after providing it to ValidatePartiesAreUnique")
		})
	}
}

func (s *ScopeTestSuite) TestValidatePartiesBasic() {
	tests := []struct {
		name    string
		parties []Party
		expErr  string
	}{
		{
			name:    "nil parties",
			parties: nil,
			expErr:  "at least one party is required",
		},
		{
			name:    "empty parties",
			parties: []Party{},
			expErr:  "at least one party is required",
		},
		{
			name:    "one bad party",
			parties: []Party{{Address: "", Role: 3}},
			expErr:  "missing party address",
		},
		{
			name: "not unique",
			parties: []Party{
				{Address: sdk.AccAddress("abc").String(), Role: 1},
				{Address: sdk.AccAddress("def").String(), Role: 2},
				{Address: sdk.AccAddress("ghi").String(), Role: 3},
				{Address: sdk.AccAddress("def").String(), Role: 2},
				{Address: sdk.AccAddress("jkl").String(), Role: 5},
			},
			expErr: "duplicate parties not allowed: address = " + sdk.AccAddress("def").String() + ", role = SERVICER, indexes: 1, 3",
		},
		{
			name: "five good parties",
			parties: []Party{
				{Address: sdk.AccAddress("abc").String(), Role: 1, Optional: true},
				{Address: sdk.AccAddress("def").String(), Role: 2, Optional: false},
				{Address: sdk.AccAddress("ghi").String(), Role: 3, Optional: true},
				{Address: sdk.AccAddress("abc").String(), Role: 2, Optional: false},
				{Address: sdk.AccAddress("jkl").String(), Role: 5, Optional: true},
			},
			expErr: "",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var orig []Party
			if tc.parties != nil {
				orig = make([]Party, 0, len(tc.parties))
				orig = append(orig, tc.parties...)
			}
			var err error
			testFunc := func() {
				err = ValidatePartiesBasic(tc.parties)
			}
			s.Require().NotPanics(testFunc, "ValidatePartiesBasic")
			if len(tc.expErr) > 0 {
				s.Assert().EqualError(err, tc.expErr, "ValidatePartiesBasic")
			} else {
				s.Assert().NoError(err, tc.expErr, "ValidatePartiesBasic")
			}
			s.Assert().Equal(orig, tc.parties, "parties slice after providing it to ValidatePartiesBasic")
		})
	}
}

func (s *ScopeTestSuite) TestEqualParties() {
	tests := []struct {
		name     string
		p1       []Party
		p2       []Party
		expected bool
	}{
		{
			name:     "empty sets",
			p1:       []Party{},
			p2:       []Party{},
			expected: true,
		},
		{
			name:     "one party in each that is equal",
			p1:       []Party{{Address: "abc", Role: 3}},
			p2:       []Party{{Address: "abc", Role: 3}},
			expected: true,
		},
		{
			name:     "one party in each different addresses",
			p1:       []Party{{Address: "abc", Role: 3}},
			p2:       []Party{{Address: "abcd", Role: 3}},
			expected: false,
		},
		{
			name:     "one party in each different roles",
			p1:       []Party{{Address: "abc", Role: 3}},
			p2:       []Party{{Address: "abc", Role: 4}},
			expected: false,
		},
		{
			name:     "both have 3 equal elements in same order",
			p1:       []Party{{Address: "abc", Role: 3}, {Address: "def", Role: 4}, {Address: "ghi", Role: 5}},
			p2:       []Party{{Address: "abc", Role: 3}, {Address: "def", Role: 4}, {Address: "ghi", Role: 5}},
			expected: true,
		},
		{
			name:     "both have 3 equal elements in different order",
			p1:       []Party{{Address: "abc", Role: 3}, {Address: "def", Role: 4}, {Address: "ghi", Role: 5}},
			p2:       []Party{{Address: "def", Role: 4}, {Address: "ghi", Role: 5}, {Address: "abc", Role: 3}},
			expected: true,
		},
		{
			name:     "one missing from p1",
			p1:       []Party{{Address: "abc", Role: 3}, {Address: "ghi", Role: 5}},
			p2:       []Party{{Address: "abc", Role: 3}, {Address: "def", Role: 4}, {Address: "ghi", Role: 5}},
			expected: false,
		},
		{
			name:     "one missing from p2",
			p1:       []Party{{Address: "abc", Role: 3}, {Address: "def", Role: 4}, {Address: "ghi", Role: 5}},
			p2:       []Party{{Address: "abc", Role: 3}, {Address: "ghi", Role: 5}},
			expected: false,
		},
		{
			name:     "two equal parties reverse order",
			p1:       []Party{{Address: "aaa", Role: 3}, {Address: "bbb", Role: 5}},
			p2:       []Party{{Address: "bbb", Role: 5}, {Address: "aaa", Role: 3}},
			expected: true,
		},
		{
			name:     "three equal parties random order",
			p1:       []Party{{Address: "aaa", Role: 3}, {Address: "bbb", Role: 5}, {Address: "ccc", Role: 4}},
			p2:       []Party{{Address: "bbb", Role: 5}, {Address: "ccc", Role: 4}, {Address: "aaa", Role: 3}},
			expected: true,
		},
		{
			name:     "1 equal and 1 diff addr",
			p1:       []Party{{Address: "aaa", Role: 3}, {Address: "bba", Role: 4}},
			p2:       []Party{{Address: "aaa", Role: 3}, {Address: "bbb", Role: 4}},
			expected: false,
		},
		{
			name:     "1 equal and 1 diff role",
			p1:       []Party{{Address: "aaa", Role: 3}, {Address: "bbb", Role: 3}},
			p2:       []Party{{Address: "aaa", Role: 4}, {Address: "bbb", Role: 3}},
			expected: false,
		},
		{
			name:     "one party different optional",
			p1:       []Party{{Address: "AAA", Role: 1, Optional: true}},
			p2:       []Party{{Address: "AAA", Role: 1, Optional: false}},
			expected: false,
		},
		{
			name:     "two parties first with different optional",
			p1:       []Party{{Address: "AAA", Role: 1, Optional: true}, {Address: "BBB", Role: 2, Optional: true}},
			p2:       []Party{{Address: "AAA", Role: 1, Optional: false}, {Address: "BBB", Role: 2, Optional: true}},
			expected: false,
		},
		{
			name:     "two parties second with different optional",
			p1:       []Party{{Address: "AAA", Role: 1, Optional: true}, {Address: "BBB", Role: 2, Optional: false}},
			p2:       []Party{{Address: "AAA", Role: 1, Optional: true}, {Address: "BBB", Role: 2, Optional: true}},
			expected: false,
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			// copy the two slices so we can later make sure the ones provided didn't change.
			op1 := append(make([]Party, 0, len(tc.p1)), tc.p1...)
			op2 := append(make([]Party, 0, len(tc.p2)), tc.p2...)
			// Do the thing.
			actual := EqualParties(tc.p1, tc.p2)
			assert.Equal(t, tc.expected, actual, "result")
			assert.Equal(t, op1, tc.p1, "p1")
			assert.Equal(t, op2, tc.p2, "p2")
		})
	}
}

type otherParty struct {
	address  string
	role     PartyType
	optional bool
}

var _ Partier = (*otherParty)(nil)

func (p otherParty) GetAddress() string {
	return p.address
}

func (p otherParty) GetRole() PartyType {
	return p.role
}

func (p otherParty) GetOptional() bool {
	return p.optional
}

func (s *ScopeTestSuite) TestEqualPartiers() {
	aParty := Party{
		Address:  "123",
		Role:     88,
		Optional: true,
	}

	tests := []struct {
		name string
		p1   Partier
		p2   Partier
		exp  bool
	}{
		{
			name: "nil nil",
			p1:   nil,
			p2:   nil,
			exp:  true,
		},
		{
			name: "nil party",
			p1:   nil,
			p2:   &aParty,
			exp:  false,
		},
		{
			name: "party nil",
			p1:   &aParty,
			p2:   nil,
			exp:  false,
		},
		{
			name: "same references",
			p1:   &aParty,
			p2:   &aParty,
			exp:  true,
		},
		{
			name: "equal parties same type",
			p1:   &Party{Address: "333", Role: 3, Optional: false},
			p2:   &Party{Address: "333", Role: 3, Optional: false},
			exp:  true,
		},
		{
			name: "equal parties different types",
			p1:   &Party{Address: "333", Role: 3, Optional: false},
			p2:   &otherParty{address: "333", role: 3, optional: false},
			exp:  true,
		},
		{
			name: "different addresses same types",
			p1:   &Party{Address: "333", Role: 3, Optional: false},
			p2:   &Party{Address: "444", Role: 3, Optional: false},
			exp:  false,
		},
		{
			name: "different roles same types",
			p1:   &Party{Address: "333", Role: 4, Optional: false},
			p2:   &Party{Address: "333", Role: 3, Optional: false},
			exp:  false,
		},
		{
			name: "different optionals same types",
			p1:   &Party{Address: "333", Role: 3, Optional: false},
			p2:   &Party{Address: "333", Role: 3, Optional: true},
			exp:  false,
		},
		{
			name: "different addresses different types",
			p1:   &Party{Address: "333", Role: 3, Optional: false},
			p2:   &otherParty{address: "444", role: 3, optional: false},
			exp:  false,
		},
		{
			name: "different roles different types",
			p1:   &Party{Address: "333", Role: 4, Optional: false},
			p2:   &otherParty{address: "333", role: 3, optional: false},
			exp:  false,
		},
		{
			name: "different optionals different types",
			p1:   &Party{Address: "333", Role: 3, Optional: false},
			p2:   &otherParty{address: "333", role: 3, optional: true},
			exp:  false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			actual := EqualPartiers(tc.p1, tc.p2)
			s.Assert().Equal(tc.exp, actual, "EqualPartiers")
		})
	}
}

func (s *ScopeTestSuite) TestSamePartiers() {
	aParty := Party{
		Address:  "123",
		Role:     88,
		Optional: true,
	}

	tests := []struct {
		name string
		p1   Partier
		p2   Partier
		exp  bool
	}{
		{
			name: "nil nil",
			p1:   nil,
			p2:   nil,
			exp:  true,
		},
		{
			name: "nil party",
			p1:   nil,
			p2:   &aParty,
			exp:  false,
		},
		{
			name: "party nil",
			p1:   &aParty,
			p2:   nil,
			exp:  false,
		},
		{
			name: "same references",
			p1:   &aParty,
			p2:   &aParty,
			exp:  true,
		},
		{
			name: "equal parties same type",
			p1:   &Party{Address: "333", Role: 3, Optional: false},
			p2:   &Party{Address: "333", Role: 3, Optional: false},
			exp:  true,
		},
		{
			name: "equal parties different types",
			p1:   &Party{Address: "333", Role: 3, Optional: false},
			p2:   &otherParty{address: "333", role: 3, optional: false},
			exp:  true,
		},
		{
			name: "different addresses same types",
			p1:   &Party{Address: "333", Role: 3, Optional: false},
			p2:   &Party{Address: "444", Role: 3, Optional: false},
			exp:  false,
		},
		{
			name: "different roles same types",
			p1:   &Party{Address: "333", Role: 4, Optional: false},
			p2:   &Party{Address: "333", Role: 3, Optional: false},
			exp:  false,
		},
		{
			name: "different optionals same types",
			p1:   &Party{Address: "333", Role: 3, Optional: false},
			p2:   &Party{Address: "333", Role: 3, Optional: true},
			exp:  true,
		},
		{
			name: "different addresses different types",
			p1:   &Party{Address: "333", Role: 3, Optional: false},
			p2:   &otherParty{address: "444", role: 3, optional: false},
			exp:  false,
		},
		{
			name: "different roles different types",
			p1:   &Party{Address: "333", Role: 4, Optional: false},
			p2:   &otherParty{address: "333", role: 3, optional: false},
			exp:  false,
		},
		{
			name: "different optionals different types",
			p1:   &Party{Address: "333", Role: 3, Optional: false},
			p2:   &otherParty{address: "333", role: 3, optional: true},
			exp:  true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			actual := SamePartiers(tc.p1, tc.p2)
			s.Assert().Equal(tc.exp, actual, "EqualPartiers")
		})
	}
}

func (s *ScopeTestSuite) TestParty_Equals() {
	tests := []struct {
		name string
		p1   Party
		p2   Partier
		exp  bool
	}{
		{name: "different addresses", p1: Party{Address: "123"}, p2: &Party{Address: "456"}, exp: false},
		{name: "different roles", p1: Party{Role: 1}, p2: &Party{Role: 2}, exp: false},
		{name: "different optional", p1: Party{Optional: true}, p2: &Party{Optional: false}, exp: false},
		{
			name: "all same",
			p1:   Party{Address: "1", Role: 1, Optional: true},
			p2:   &Party{Address: "1", Role: 1, Optional: true},
			exp:  true,
		},

		{name: "other type different addresses", p1: Party{Address: "123"}, p2: &otherParty{address: "456"}, exp: false},
		{name: "other type different roles", p1: Party{Role: 1}, p2: &otherParty{role: 2}, exp: false},
		{name: "other type different optional", p1: Party{Optional: true}, p2: &otherParty{optional: false}, exp: false},
		{
			name: "other type all same",
			p1:   Party{Address: "1", Role: 1, Optional: true},
			p2:   &otherParty{address: "1", role: 1, optional: true},
			exp:  true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			actual := tc.p1.Equals(tc.p2)
			s.Assert().Equal(tc.exp, actual, "%v.Equals(%v)", tc.p1, tc.p2)
		})
	}
}

func (s *ScopeTestSuite) TestParty_IsSameAs() {
	tests := []struct {
		name string
		p1   Party
		p2   Partier
		exp  bool
	}{
		{name: "different addresses", p1: Party{Address: "123"}, p2: &Party{Address: "456"}, exp: false},
		{name: "different roles", p1: Party{Role: 1}, p2: &Party{Role: 2}, exp: false},
		{name: "different optional", p1: Party{Optional: true}, p2: &Party{Optional: false}, exp: true},
		{
			name: "all same",
			p1:   Party{Address: "1", Role: 1, Optional: true},
			p2:   &Party{Address: "1", Role: 1, Optional: true},
			exp:  true,
		},

		{name: "other type different addresses", p1: Party{Address: "123"}, p2: &otherParty{address: "456"}, exp: false},
		{name: "other type different roles", p1: Party{Role: 1}, p2: &otherParty{role: 2}, exp: false},
		{name: "other type different optional", p1: Party{Optional: true}, p2: &otherParty{optional: false}, exp: true},
		{
			name: "other type all same",
			p1:   Party{Address: "1", Role: 1, Optional: true},
			p2:   &otherParty{address: "1", role: 1, optional: true},
			exp:  true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			actual := tc.p1.IsSameAs(tc.p2)
			s.Assert().Equal(tc.exp, actual, "%v.IsSameAs(%v)", tc.p1, tc.p2)
		})
	}
}

func (s *ScopeTestSuite) TestEquivalentDataAssessors() {
	tests := []struct {
		name     string
		p1       []string
		p2       []string
		expected bool
	}{
		{
			name:     "empty sets",
			p1:       []string{},
			p2:       []string{},
			expected: true,
		},
		{
			name:     "one in each same",
			p1:       []string{"abc"},
			p2:       []string{"abc"},
			expected: true,
		},
		{
			name:     "one in each different",
			p1:       []string{"abc"},
			p2:       []string{"abcd"},
			expected: false,
		},
		{
			name:     "both have 3 equal elements in same order",
			p1:       []string{"abc", "def", "ghi"},
			p2:       []string{"abc", "def", "ghi"},
			expected: true,
		},
		{
			name:     "both have 3 equal elements in different order",
			p1:       []string{"abc", "def", "ghi"},
			p2:       []string{"def", "ghi", "abc"},
			expected: true,
		},
		{
			name:     "one missing from p1",
			p1:       []string{"abc", "ghi"},
			p2:       []string{"abc", "def", "ghi"},
			expected: false,
		},
		{
			name:     "one missing from p2",
			p1:       []string{"abc", "def", "ghi"},
			p2:       []string{"abc", "ghi"},
			expected: false,
		},
		{
			name:     "aab vs abb",
			p1:       []string{"aaa", "aaa", "bbb"},
			p2:       []string{"aaa", "bbb", "bbb"},
			expected: true,
		},
		{
			name:     "aab vs ab",
			p1:       []string{"aaa", "aaa", "bbb"},
			p2:       []string{"aaa", "bbb"},
			expected: true,
		},
		{
			name:     "abb vs ab",
			p1:       []string{"aaa", "bbb", "bbb"},
			p2:       []string{"aaa", "bbb"},
			expected: true,
		},
		{
			name:     "aab vs aba",
			p1:       []string{"aaa", "aaa", "bbb"},
			p2:       []string{"aaa", "bbb", "aaa"},
			expected: true,
		},
		{
			name:     "aaa vs bbb",
			p1:       []string{"aaa", "aaa", "aaa"},
			p2:       []string{"bbb", "bbb", "bbb"},
			expected: false,
		},
		{
			name:     "2 equal and 1 diff",
			p1:       []string{"aaa", "aaa", "aaa"},
			p2:       []string{"aaa", "aaa", "bbb"},
			expected: false,
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			// copy the two slices so we can later make sure the ones provided didn't change.
			op1 := append(make([]string, 0, len(tc.p1)), tc.p1...)
			op2 := append(make([]string, 0, len(tc.p2)), tc.p2...)
			// Do the thing.
			actual := equivalentDataAssessors(tc.p1, tc.p2)
			assert.Equal(t, tc.expected, actual, "result")
			assert.Equal(t, op1, tc.p1, "p1")
			assert.Equal(t, op2, tc.p2, "p2")
		})
	}
}

func TestNetAssetValueValidate(t *testing.T) {
	tests := []struct {
		name   string
		nav    NetAssetValue
		expErr string
	}{
		{
			name: "price validation fails",
			nav: NetAssetValue{
				Price: sdk.Coin{Denom: "invalidDenom", Amount: sdkmath.NewInt(-100)},
			},
			expErr: "negative coin amount: -100", // Replace with the actual error message from Price.Validate()
		},
		{
			name:   "invalid denom",
			nav:    NetAssetValue{},
			expErr: "invalid denom: ",
		},
		{
			name: "successful with 0 volume and coin",
			nav: NetAssetValue{
				Price: sdk.NewInt64Coin("usdcents", 0),
			},
		},
		{
			name: "successful",
			nav: NetAssetValue{
				Price: sdk.NewInt64Coin("jackthecat", 420),
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := tt.nav.Validate()
			if len(tt.expErr) > 0 {
				assert.EqualErrorf(t, err, tt.expErr, "NetAssetValue validate expected error")
			} else {
				assert.NoError(t, err, "NetAssetValue validate should have passed")
			}
		})
	}
}
