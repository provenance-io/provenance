package types

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestWrapScope(t *testing.T) {
	scopeUUID := uuid.MustParse("AB1D357C-E86D-4102-A532-5E47A707E4AD")
	specUUID := uuid.MustParse("245FBC68-1C47-41AA-A0D8-5CF7CD13F471")
	scope := &Scope{
		ScopeId:           ScopeMetadataAddress(scopeUUID),
		SpecificationId:   ScopeSpecMetadataAddress(specUUID),
		DataAccess:        []string{sdk.AccAddress("data_access_addr____").String()},
		ValueOwnerAddress: sdk.AccAddress("value_owner_addr____").String(),
	}
	t.Logf("scope: %s (%s)", scope.ScopeId.String(), scopeUUID.String())
	t.Logf("spec: %s (%s)", scope.SpecificationId.String(), specUUID.String())

	tests := []struct {
		name   string
		scope  *Scope
		incInf bool
		exp    *ScopeWrapper
	}{
		{
			name:   "nil scope and include info",
			scope:  nil,
			incInf: true,
			exp:    &ScopeWrapper{},
		},
		{
			name:   "nil scope and no info",
			scope:  nil,
			incInf: false,
			exp:    &ScopeWrapper{},
		},
		{
			name:   "not nil scope and include info",
			scope:  scope,
			incInf: true,
			exp: &ScopeWrapper{
				Scope:           scope,
				ScopeIdInfo:     GetScopeIDInfo(scope.ScopeId),
				ScopeSpecIdInfo: GetScopeSpecIDInfo(scope.SpecificationId),
			},
		},
		{
			name:   "not nil scope but no info",
			scope:  scope,
			incInf: false,
			exp:    &ScopeWrapper{Scope: scope},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := WrapScope(tc.scope, tc.incInf)
			assert.Equal(t, tc.exp, w, "WrapScope")
		})
	}
}

func TestWrapScopeNotFound(t *testing.T) {
	zeroAddr := MetadataAddress(append(ScopeKeyPrefix, bytes.Repeat([]byte{0}, 16)...))
	t.Logf("zeroAddr: %s", zeroAddr.String())

	scopeUUID := uuid.MustParse("AB1D357C-E86D-4102-A532-5E47A707E4AD")
	scopeAddr := ScopeMetadataAddress(scopeUUID)
	t.Logf("scopeAddr: %s (%s)", scopeAddr.String(), scopeUUID.String())

	randUUID1 := uuid.New()
	randAddr1 := ScopeMetadataAddress(randUUID1)
	t.Logf("randAddr1: %s (%s)", randAddr1.String(), randUUID1.String())

	randUUID2 := uuid.New()
	randAddr2 := ScopeMetadataAddress(randUUID2)
	t.Logf("randAddr2: %s (%s)", randAddr2.String(), randUUID2.String())

	tests := []struct {
		name string
		addr MetadataAddress
	}{
		{name: "nil", addr: nil},
		{name: "empty", addr: []byte{}},
		{name: "zeros", addr: zeroAddr},
		{name: "constant", addr: scopeAddr},
		{name: "random 1", addr: randAddr1},
		{name: "random 2", addr: randAddr2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			exp := &ScopeWrapper{ScopeIdInfo: GetScopeIDInfo(tc.addr)}
			act := WrapScopeNotFound(tc.addr)
			assert.Equal(t, exp, act, "WrapScopeNotFound")
		})
	}
}

func TestWrapSession(t *testing.T) {
	scopeUUID := uuid.MustParse("AB1D357C-E86D-4102-A532-5E47A707E4AD")
	sessionUUID := uuid.MustParse("618A5BE9-10E6-40AD-8925-8E06D368D2E6")
	specUUID := uuid.MustParse("245FBC68-1C47-41AA-A0D8-5CF7CD13F471")
	session := &Session{
		SessionId:       SessionMetadataAddress(scopeUUID, sessionUUID),
		SpecificationId: ContractSpecMetadataAddress(specUUID),
		Name:            "testSession",
		Context:         []byte("test session context"),
	}
	t.Logf("session: %s (%s, %s)", session.SessionId.String(), scopeUUID.String(), sessionUUID.String())
	t.Logf("spec: %s (%s)", session.SpecificationId.String(), specUUID.String())

	tests := []struct {
		name    string
		session *Session
		incInf  bool
		exp     *SessionWrapper
	}{
		{
			name:    "nil session and include info",
			session: nil,
			incInf:  true,
			exp:     &SessionWrapper{},
		},
		{
			name:    "nil session and no info",
			session: nil,
			incInf:  false,
			exp:     &SessionWrapper{},
		},
		{
			name:    "not nil session and include info",
			session: session,
			incInf:  true,
			exp: &SessionWrapper{
				Session:            session,
				SessionIdInfo:      GetSessionIDInfo(session.SessionId),
				ContractSpecIdInfo: GetContractSpecIDInfo(session.SpecificationId),
			},
		},
		{
			name:    "not nil session but no info",
			session: session,
			incInf:  false,
			exp:     &SessionWrapper{Session: session},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := WrapSession(tc.session, tc.incInf)
			assert.Equal(t, tc.exp, w, "WrapSession")
		})
	}
}

func TestWrapSessionNotFound(t *testing.T) {
	zeroAddr := MetadataAddress(append(SessionKeyPrefix, bytes.Repeat([]byte{0}, 32)...))
	t.Logf("zeroAddr: %s", zeroAddr.String())

	scopeUUID := uuid.MustParse("AB1D357C-E86D-4102-A532-5E47A707E4AD")
	sessionUUID := uuid.MustParse("618A5BE9-10E6-40AD-8925-8E06D368D2E6")
	sessionAddr := SessionMetadataAddress(scopeUUID, sessionUUID)
	t.Logf("sessionAddr: %s (%s, %s)", sessionAddr.String(), scopeUUID.String(), sessionUUID.String())

	randUUID1A, randUUID1B := uuid.New(), uuid.New()
	randAddr1 := SessionMetadataAddress(randUUID1A, randUUID1B)
	t.Logf("randAddr1: %s (%s, %s)", randAddr1.String(), randUUID1A.String(), randUUID1B.String())

	randUUID2A, randUUID2B := uuid.New(), uuid.New()
	randAddr2 := SessionMetadataAddress(randUUID2A, randUUID2B)
	t.Logf("randAddr2: %s (%s, %s)", randAddr2.String(), randUUID2A.String(), randUUID2B.String())

	tests := []struct {
		name string
		addr MetadataAddress
	}{
		{name: "nil", addr: nil},
		{name: "empty", addr: []byte{}},
		{name: "zeros", addr: zeroAddr},
		{name: "constant", addr: sessionAddr},
		{name: "random 1", addr: randAddr1},
		{name: "random 2", addr: randAddr2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			exp := &SessionWrapper{SessionIdInfo: GetSessionIDInfo(tc.addr)}
			act := WrapSessionNotFound(tc.addr)
			assert.Equal(t, exp, act, "WrapSessionNotFound")
		})
	}
}

func TestWrapRecord(t *testing.T) {
	scopeUUID := uuid.MustParse("AB1D357C-E86D-4102-A532-5E47A707E4AD")
	sessionUUID := uuid.MustParse("618A5BE9-10E6-40AD-8925-8E06D368D2E6")
	specUUID := uuid.MustParse("245FBC68-1C47-41AA-A0D8-5CF7CD13F471")
	record := &Record{
		Name:      "testrecord",
		SessionId: SessionMetadataAddress(scopeUUID, sessionUUID),
		Process:   Process{Name: "myproc"},
	}
	record.SpecificationId = RecordSpecMetadataAddress(specUUID, record.Name)
	t.Logf("record: %s (%s, %q)", record.GetRecordAddress().String(), scopeUUID.String(), record.Name)
	t.Logf("spec: %s (%s, %q)", record.SpecificationId.String(), specUUID.String(), record.Name)

	tests := []struct {
		name   string
		record *Record
		incInf bool
		exp    *RecordWrapper
	}{
		{
			name:   "nil record and include info",
			record: nil,
			incInf: true,
			exp:    &RecordWrapper{},
		},
		{
			name:   "nil record and no info",
			record: nil,
			incInf: false,
			exp:    &RecordWrapper{},
		},
		{
			name:   "not nil record and include info",
			record: record,
			incInf: true,
			exp: &RecordWrapper{
				Record:           record,
				RecordIdInfo:     GetRecordIDInfo(record.GetRecordAddress()),
				RecordSpecIdInfo: GetRecordSpecIDInfo(record.SpecificationId),
			},
		},
		{
			name:   "not nil record but no info",
			record: record,
			incInf: false,
			exp:    &RecordWrapper{Record: record},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := WrapRecord(tc.record, tc.incInf)
			assert.Equal(t, tc.exp, w, "WrapRecord")
		})
	}
}

func TestWrapRecordNotFound(t *testing.T) {
	zeroAddr := MetadataAddress(append(RecordKeyPrefix, bytes.Repeat([]byte{0}, 32)...))
	t.Logf("zeroAddr: %s", zeroAddr.String())

	scopeUUID := uuid.MustParse("AB1D357C-E86D-4102-A532-5E47A707E4AD")
	recordName := "testrecord"
	recordAddr := RecordMetadataAddress(scopeUUID, recordName)
	t.Logf("recordAddr: %s (%s, %q)", recordAddr.String(), scopeUUID.String(), recordName)

	randUUID1 := uuid.New()
	randName1 := "test.record." + strings.ReplaceAll(randUUID1.String(), "-", "")
	randAddr1 := RecordMetadataAddress(randUUID1, randName1)
	t.Logf("randAddr1: %s (%s, %q)", randAddr1.String(), randUUID1.String(), randName1)

	randUUID2 := uuid.New()
	randName2 := "test.record." + strings.ReplaceAll(randUUID2.String(), "-", "")
	randAddr2 := RecordMetadataAddress(randUUID2, randName2)
	t.Logf("randAddr2: %s (%s, %q)", randAddr2.String(), randUUID2.String(), randName2)

	tests := []struct {
		name string
		addr MetadataAddress
	}{
		{name: "nil", addr: nil},
		{name: "empty", addr: []byte{}},
		{name: "zeros", addr: zeroAddr},
		{name: "constant", addr: recordAddr},
		{name: "random 1", addr: randAddr1},
		{name: "random 2", addr: randAddr2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			exp := &RecordWrapper{RecordIdInfo: GetRecordIDInfo(tc.addr)}
			act := WrapRecordNotFound(tc.addr)
			assert.Equal(t, exp, act, "WrapRecordNotFound")
		})
	}
}

func TestWrapScopeSpec(t *testing.T) {
	specUUID := uuid.MustParse("245FBC68-1C47-41AA-A0D8-5CF7CD13F471")
	spec := &ScopeSpecification{
		SpecificationId: ScopeSpecMetadataAddress(specUUID),
		OwnerAddresses:  []string{sdk.AccAddress("owner_address_0_____").String()},
	}
	t.Logf("spec: %s (%s)", spec.SpecificationId.String(), specUUID.String())

	tests := []struct {
		name   string
		spec   *ScopeSpecification
		incInf bool
		exp    *ScopeSpecificationWrapper
	}{
		{
			name:   "nil spec and include info",
			spec:   nil,
			incInf: true,
			exp:    &ScopeSpecificationWrapper{},
		},
		{
			name:   "nil scope and no info",
			spec:   nil,
			incInf: false,
			exp:    &ScopeSpecificationWrapper{},
		},
		{
			name:   "not nil scope and include info",
			spec:   spec,
			incInf: true,
			exp: &ScopeSpecificationWrapper{
				Specification:   spec,
				ScopeSpecIdInfo: GetScopeSpecIDInfo(spec.SpecificationId),
			},
		},
		{
			name:   "not nil scope but no info",
			spec:   spec,
			incInf: false,
			exp:    &ScopeSpecificationWrapper{Specification: spec},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := WrapScopeSpec(tc.spec, tc.incInf)
			assert.Equal(t, tc.exp, w, "WrapScopeSpec")
		})
	}
}

func TestWrapScopeSpecNotFound(t *testing.T) {
	zeroAddr := MetadataAddress(append(ScopeSpecificationKeyPrefix, bytes.Repeat([]byte{0}, 16)...))
	t.Logf("zeroAddr: %s", zeroAddr.String())

	specUUID := uuid.MustParse("AB1D357C-E86D-4102-A532-5E47A707E4AD")
	specAddr := ScopeSpecMetadataAddress(specUUID)
	t.Logf("specAddr: %s (%s)", specAddr.String(), specUUID.String())

	randUUID1 := uuid.New()
	randAddr1 := ScopeSpecMetadataAddress(randUUID1)
	t.Logf("randAddr1: %s (%s)", randAddr1.String(), randUUID1.String())

	randUUID2 := uuid.New()
	randAddr2 := ScopeSpecMetadataAddress(randUUID2)
	t.Logf("randAddr2: %s (%s)", randAddr2.String(), randUUID2.String())

	tests := []struct {
		name string
		addr MetadataAddress
	}{
		{name: "nil", addr: nil},
		{name: "empty", addr: []byte{}},
		{name: "zeros", addr: zeroAddr},
		{name: "constant", addr: specAddr},
		{name: "random 1", addr: randAddr1},
		{name: "random 2", addr: randAddr2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			exp := &ScopeSpecificationWrapper{ScopeSpecIdInfo: GetScopeSpecIDInfo(tc.addr)}
			act := WrapScopeSpecNotFound(tc.addr)
			assert.Equal(t, exp, act, "WrapScopeSpecNotFound")
		})
	}
}

func TestWrapContractSpec(t *testing.T) {
	specUUID := uuid.MustParse("245FBC68-1C47-41AA-A0D8-5CF7CD13F471")
	spec := &ContractSpecification{
		SpecificationId: ContractSpecMetadataAddress(specUUID),
		OwnerAddresses:  []string{sdk.AccAddress("owner_address_0_____").String()},
	}
	t.Logf("spec: %s (%s)", spec.SpecificationId.String(), specUUID.String())

	tests := []struct {
		name   string
		spec   *ContractSpecification
		incInf bool
		exp    *ContractSpecificationWrapper
	}{
		{
			name:   "nil spec and include info",
			spec:   nil,
			incInf: true,
			exp:    &ContractSpecificationWrapper{},
		},
		{
			name:   "nil scope and no info",
			spec:   nil,
			incInf: false,
			exp:    &ContractSpecificationWrapper{},
		},
		{
			name:   "not nil scope and include info",
			spec:   spec,
			incInf: true,
			exp: &ContractSpecificationWrapper{
				Specification:      spec,
				ContractSpecIdInfo: GetContractSpecIDInfo(spec.SpecificationId),
			},
		},
		{
			name:   "not nil scope but no info",
			spec:   spec,
			incInf: false,
			exp:    &ContractSpecificationWrapper{Specification: spec},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := WrapContractSpec(tc.spec, tc.incInf)
			assert.Equal(t, tc.exp, w, "WrapContractSpec")
		})
	}
}

func TestWrapContractSpecNotFound(t *testing.T) {
	zeroAddr := MetadataAddress(append(ContractSpecificationKeyPrefix, bytes.Repeat([]byte{0}, 16)...))
	t.Logf("zeroAddr: %s", zeroAddr.String())

	specUUID := uuid.MustParse("AB1D357C-E86D-4102-A532-5E47A707E4AD")
	specAddr := ContractSpecMetadataAddress(specUUID)
	t.Logf("specAddr: %s (%s)", specAddr.String(), specUUID.String())

	randUUID1 := uuid.New()
	randAddr1 := ContractSpecMetadataAddress(randUUID1)
	t.Logf("randAddr1: %s (%s)", randAddr1.String(), randUUID1.String())

	randUUID2 := uuid.New()
	randAddr2 := ContractSpecMetadataAddress(randUUID2)
	t.Logf("randAddr2: %s (%s)", randAddr2.String(), randUUID2.String())

	tests := []struct {
		name string
		addr MetadataAddress
	}{
		{name: "nil", addr: nil},
		{name: "empty", addr: []byte{}},
		{name: "zeros", addr: zeroAddr},
		{name: "constant", addr: specAddr},
		{name: "random 1", addr: randAddr1},
		{name: "random 2", addr: randAddr2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			exp := &ContractSpecificationWrapper{ContractSpecIdInfo: GetContractSpecIDInfo(tc.addr)}
			act := WrapContractSpecNotFound(tc.addr)
			assert.Equal(t, exp, act, "WrapContractSpecNotFound")
		})
	}
}

func TestWrapRecordSpec(t *testing.T) {
	specUUID := uuid.MustParse("245FBC68-1C47-41AA-A0D8-5CF7CD13F471")
	spec := &RecordSpecification{
		Name:     "wrapmeup",
		TypeName: "testype",
	}
	spec.SpecificationId = RecordSpecMetadataAddress(specUUID, spec.Name)
	t.Logf("spec: %s (%s, %q)", spec.SpecificationId.String(), specUUID.String(), spec.Name)

	tests := []struct {
		name   string
		spec   *RecordSpecification
		incInf bool
		exp    *RecordSpecificationWrapper
	}{
		{
			name:   "nil spec and include info",
			spec:   nil,
			incInf: true,
			exp:    &RecordSpecificationWrapper{},
		},
		{
			name:   "nil scope and no info",
			spec:   nil,
			incInf: false,
			exp:    &RecordSpecificationWrapper{},
		},
		{
			name:   "not nil scope and include info",
			spec:   spec,
			incInf: true,
			exp: &RecordSpecificationWrapper{
				Specification:    spec,
				RecordSpecIdInfo: GetRecordSpecIDInfo(spec.SpecificationId),
			},
		},
		{
			name:   "not nil scope but no info",
			spec:   spec,
			incInf: false,
			exp:    &RecordSpecificationWrapper{Specification: spec},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := WrapRecordSpec(tc.spec, tc.incInf)
			assert.Equal(t, tc.exp, w, "WrapRecordSpec")
		})
	}
}

func TestWrapRecordSpecs(t *testing.T) {
	spec1UUID := uuid.MustParse("31E634BC-50F6-4AB2-8A29-AD911ABE9AC8")
	spec1 := &RecordSpecification{Name: "testspec1"}
	spec1.SpecificationId = RecordSpecMetadataAddress(spec1UUID, spec1.Name)
	spec1WInf := &RecordSpecificationWrapper{
		Specification:    spec1,
		RecordSpecIdInfo: GetRecordSpecIDInfo(spec1.SpecificationId),
	}
	spec1NoInf := &RecordSpecificationWrapper{Specification: spec1}
	t.Logf("spec1: %s (%s, %q)", spec1.SpecificationId.String(), spec1UUID.String(), spec1.Name)

	spec2UUID := uuid.MustParse("275F2CE0-7C2F-4F9B-A1CE-6294E43E9B6F")
	spec2 := &RecordSpecification{Name: "testspec2"}
	spec2.SpecificationId = RecordSpecMetadataAddress(spec2UUID, spec2.Name)
	spec2WInf := &RecordSpecificationWrapper{
		Specification:    spec2,
		RecordSpecIdInfo: GetRecordSpecIDInfo(spec2.SpecificationId),
	}
	spec2NoInf := &RecordSpecificationWrapper{Specification: spec2}
	t.Logf("spec2: %s (%s, %q)", spec2.SpecificationId.String(), spec2UUID.String(), spec2.Name)

	spec3UUID := uuid.MustParse("5A42DA0C-A0E2-4986-A680-F12A4101B5BC")
	spec3 := &RecordSpecification{Name: "testspec3"}
	spec3.SpecificationId = RecordSpecMetadataAddress(spec3UUID, spec3.Name)
	spec3WInf := &RecordSpecificationWrapper{
		Specification:    spec3,
		RecordSpecIdInfo: GetRecordSpecIDInfo(spec3.SpecificationId),
	}
	spec3NoInf := &RecordSpecificationWrapper{Specification: spec3}
	t.Logf("spec3: %s (%s, %q)", spec3.SpecificationId.String(), spec3UUID.String(), spec3.Name)

	specs := func(sz ...*RecordSpecification) []*RecordSpecification {
		return sz
	}
	wrappers := func(wz ...*RecordSpecificationWrapper) []*RecordSpecificationWrapper {
		return wz
	}

	tests := []struct {
		name   string
		specs  []*RecordSpecification
		incInf bool
		exp    []*RecordSpecificationWrapper
	}{
		{
			name:  "nil",
			specs: nil,
			exp:   []*RecordSpecificationWrapper{},
		},
		{
			name:  "empty",
			specs: []*RecordSpecification{},
			exp:   []*RecordSpecificationWrapper{},
		},
		{
			name:   "one spec with info",
			specs:  specs(spec1),
			incInf: true,
			exp:    wrappers(spec1WInf),
		},
		{
			name:   "one spec no info",
			specs:  specs(spec1),
			incInf: false,
			exp:    wrappers(spec1NoInf),
		},
		{
			name:   "three specs with info",
			specs:  specs(spec1, spec2, spec3),
			incInf: true,
			exp:    wrappers(spec1WInf, spec2WInf, spec3WInf),
		},
		{
			name:   "three specs no info",
			specs:  specs(spec1, spec2, spec3),
			incInf: false,
			exp:    wrappers(spec1NoInf, spec2NoInf, spec3NoInf),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			act := WrapRecordSpecs(tc.specs, tc.incInf)
			assert.Equal(t, tc.exp, act, "WrapRecordSpecs")
		})
	}
}

func TestWrapRecordSpecNotFound(t *testing.T) {
	zeroAddr := MetadataAddress(append(RecordSpecificationKeyPrefix, bytes.Repeat([]byte{0}, 32)...))
	t.Logf("zeroAddr: %s", zeroAddr.String())

	specUUID := uuid.MustParse("AB1D357C-E86D-4102-A532-5E47A707E4AD")
	specName := "testrecordspec"
	specAddr := RecordSpecMetadataAddress(specUUID, specName)
	t.Logf("specAddr: %s (%s, %q)", specAddr.String(), specUUID.String(), specName)

	randUUID1 := uuid.New()
	randName1 := "test.record.spec." + strings.ReplaceAll(randUUID1.String(), "-", "")
	randAddr1 := RecordSpecMetadataAddress(randUUID1, randName1)
	t.Logf("randAddr1: %s (%s, %q)", randAddr1.String(), randUUID1.String(), randName1)

	randUUID2 := uuid.New()
	randName2 := "test.record.spec." + strings.ReplaceAll(randUUID2.String(), "-", "")
	randAddr2 := RecordSpecMetadataAddress(randUUID2, randName2)
	t.Logf("randAddr2: %s (%s, %q)", randAddr2.String(), randUUID2.String(), randName2)

	tests := []struct {
		name string
		addr MetadataAddress
	}{
		{name: "nil", addr: nil},
		{name: "empty", addr: []byte{}},
		{name: "zeros", addr: zeroAddr},
		{name: "constant", addr: specAddr},
		{name: "random 1", addr: randAddr1},
		{name: "random 2", addr: randAddr2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			exp := &RecordSpecificationWrapper{RecordSpecIdInfo: GetRecordSpecIDInfo(tc.addr)}
			act := WrapRecordSpecNotFound(tc.addr)
			assert.Equal(t, exp, act, "WrapRecordSpecNotFound")
		})
	}
}
