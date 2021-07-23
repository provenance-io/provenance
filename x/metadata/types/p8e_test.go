package types

import (
	b64 "encoding/base64"
	"fmt"
	"reflect"
	"testing"
	"unicode"

	"github.com/provenance-io/provenance/x/metadata/types/p8e"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/google/uuid"
)

type P8eTestSuite struct {
	suite.Suite

	pubkey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress
}

func (s *P8eTestSuite) SetupTest() {
	s.pubkey1 = secp256k1.GenPrivKey().PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()
}

func TestP8eTestSuite(t *testing.T) {
	suite.Run(t, new(P8eTestSuite))
}

// func ownerPartyList is defined in msg_test.go

// getPublicFieldNames gets all the public field names in the provided struct.
// Panics if the provided thing isn't a struct.
func getPublicFieldNames(thing interface{}) []string {
	fields, err := getAllFieldNames(thing)
	if err != nil {
		panic(err)
	}
	retval := []string{}
	for _, field := range fields {
		if unicode.IsUpper(rune(field[0])) {
			retval = append(retval, field)
		}
	}
	return retval
}

// getAllFieldNames attempts to get the field names of the provided thing.
// An empty list and an error is returned if the thing is nil or not a struct (after following all pointers).
// If the thing is a struct (or a struct after following all pointers), all of its field names will be returned.
func getAllFieldNames(thing interface{}) (fields []string, err error) {
	fields = []string{}
	err = nil
	// Initial nil check.
	if thing == nil {
		err = fmt.Errorf("cannot get field names from nil")
		return
	}
	// Get the thing type and follow it until it's not a pointer anymore.
	thingValue := reflect.ValueOf(thing)
	thingType := thingValue.Type()
	for thingType.Kind() == reflect.Ptr {
		thingValue = thingValue.Elem()
		thingType = thingValue.Type()
	}
	// Make sure we've got a struct.
	if thingType.Kind() != reflect.Struct {
		err = fmt.Errorf("cannot get fields names from %s", thingType.Kind())
		return
	}
	// Get all the field names!
	if thingType.Kind() == reflect.Struct {
		fieldCount := thingType.NumField()
		for i := 0; i < fieldCount; i++ {
			fields = append(fields, thingType.Field(i).Name)
		}
	}
	return
}

// assertSetsAreEqual asserts that all of the elements in expected are in actual.
// Assumes elements in a set are unique.
func assertSetsAreEqual(t *testing.T, expected []string, actual []string) bool {
	allPass := assert.Equal(t, len(expected), len(actual), "lengths")
	if allPass {
		for _, e := range expected {
			found := false
			for _, a := range actual {
				if e == a {
					found = true
					break
				}
			}
			if !found {
				assert.Fail(t, "entry missing from actual set", "missing entry: %s", e)
				allPass = false
			}
		}
	}
	return allPass
}

// appendUnique appends newval to list if it isn't already in list.
// Use it the same way you'd use the append built-in function.
func appendUnique(list []string, newval string) []string {
	isNew := true
	for _, val := range list {
		if val == newval {
			isNew = false
			break
		}
	}
	if isNew {
		return append(list, newval)
	}
	return list
}

func createContractSpec(inputSpecs []*p8e.DefinitionSpec, outputSpec p8e.OutputSpec, definitionSpec p8e.DefinitionSpec) p8e.ContractSpec {
	return p8e.ContractSpec{
		ConsiderationSpecs: []*p8e.ConsiderationSpec{
			{FuncName: "additionalParties",
				InputSpecs:       inputSpecs,
				OutputSpec:       &outputSpec,
				ResponsibleParty: 1,
			},
		},
		Definition:      &definitionSpec,
		InputSpecs:      inputSpecs,
		PartiesInvolved: []p8e.PartyType{p8e.PartyType_PARTY_TYPE_AFFILIATE},
	}
}

func createDefinitionSpec(name string, classname string, reference p8e.ProvenanceReference, defType int) p8e.DefinitionSpec {
	return p8e.DefinitionSpec{
		Name: name,
		ResourceLocation: &p8e.Location{Classname: classname,
			Ref: &reference,
		},
		Type: 1,
	}
}

func (s *P8eTestSuite) TestConvertP8eContractSpec() {
	validDefSpec := createDefinitionSpec("perform_input_checks", "io.provenance.loan.LoanProtos$PartiesList", p8e.ProvenanceReference{Hash: "Adv+huolGTKofYCR0dw5GHm/R7sUWOwF32XR8r8r9kDy4il5U/LApxOWYHb05jhK4+eY4YzRMRiWcxU3Lx0+Mw=="}, 1)
	validDefSpecUUID := createDefinitionSpec("perform_input_checks", "io.provenance.loan.LoanProtos$PartiesList", p8e.ProvenanceReference{ScopeUuid: &p8e.UUID{Value: uuid.New().String()}, Name: "recordname"}, 1)

	invalidDefSpecNoName := createDefinitionSpec("", "io.provenance.loan.LoanProtos$PartiesList", p8e.ProvenanceReference{Hash: "Adv+huolGTKofYCR0dw5GHm/R7sUWOwF32XR8r8r9kDy4il5U/LApxOWYHb05jhK4+eY4YzRMRiWcxU3Lx0+Mw=="}, 1)
	invalidDefSpecNoClass := createDefinitionSpec("perform_action", "", p8e.ProvenanceReference{Hash: "Adv+huolGTKofYCR0dw5GHm/R7sUWOwF32XR8r8r9kDy4il5U/LApxOWYHb05jhK4+eY4YzRMRiWcxU3Lx0+Mw=="}, 1)
	invalidDefSpecUUID := createDefinitionSpec("perform_input_checks", "io.provenance.loan.LoanProtos$PartiesList", p8e.ProvenanceReference{ScopeUuid: &p8e.UUID{Value: "not-a-uuid"}, Name: "recordname"}, 1)
	invalidDefSpecUUIDNoName := createDefinitionSpec("perform_input_checks", "io.provenance.loan.LoanProtos$PartiesList", p8e.ProvenanceReference{ScopeUuid: &p8e.UUID{Value: uuid.New().String()}}, 1)

	cases := map[string]struct {
		v39CSpec p8e.ContractSpec
		signers  []string
		wantErr  bool
		errorMsg string
	}{
		"should convert a contract spec successfully": {
			createContractSpec([]*p8e.DefinitionSpec{&validDefSpec}, p8e.OutputSpec{Spec: &validDefSpec}, validDefSpec),
			[]string{s.user1},
			false,
			"",
		},
		"should convert a contract spec successfully with uuid input spec": {
			createContractSpec([]*p8e.DefinitionSpec{&validDefSpecUUID}, p8e.OutputSpec{Spec: &validDefSpec}, validDefSpec),
			[]string{s.user1},
			false,
			"",
		},
		"should fail to validate basic on contract specification": {
			createContractSpec([]*p8e.DefinitionSpec{&validDefSpec}, p8e.OutputSpec{Spec: &validDefSpec}, validDefSpec),
			[]string{},
			true,
			"invalid owner addresses count (expected > 0 got: 0)",
		},
		"should fail to validatebasic on input specification": {
			createContractSpec([]*p8e.DefinitionSpec{&invalidDefSpecNoName}, p8e.OutputSpec{Spec: &validDefSpec}, validDefSpec),
			[]string{s.user1},
			true,
			"input specification name cannot be empty",
		},
		"should fail to validatebasic on record specification": {
			createContractSpec([]*p8e.DefinitionSpec{&validDefSpec}, p8e.OutputSpec{Spec: &invalidDefSpecNoClass}, validDefSpec),
			[]string{s.user1},
			true,
			"record specification type name cannot be empty",
		},
		"should fail to decode input spec uuid": {
			createContractSpec([]*p8e.DefinitionSpec{&invalidDefSpecUUID}, p8e.OutputSpec{Spec: &validDefSpec}, validDefSpec),
			[]string{s.user1},
			true,
			"invalid UUID length: 10",
		},
		"should fail to find name on record with uuid": {
			createContractSpec([]*p8e.DefinitionSpec{&invalidDefSpecUUIDNoName}, p8e.OutputSpec{Spec: &validDefSpec}, validDefSpec),
			[]string{s.user1},
			true,
			"must have a value for record name",
		},
	}

	for n, tc := range cases {
		tc := tc

		s.Run(n, func() {
			_, _, err := ConvertP8eContractSpec(&tc.v39CSpec, tc.signers)
			if tc.wantErr {
				s.Error(err)
				s.Equal(tc.errorMsg, err.Error())
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *P8eTestSuite) TestEmptyScope() {
	testedFields := []string{}
	tests := []struct {
		field string
		name  string
		test  func(scope *Scope, t *testing.T)
	}{
		{
			"",
			"Does not return nil",
			func(scope *Scope, t *testing.T) {
				assert.NotNil(t, scope)
			},
		},
		{
			"ScopeId",
			"is empty",
			func(scope *Scope, t *testing.T) {
				assert.True(t, scope.ScopeId.Empty())
			},
		},
		{
			"SpecificationId",
			"is empty",
			func(scope *Scope, t *testing.T) {
				assert.True(t, scope.SpecificationId.Empty())
			},
		},
		{
			"Owners",
			"is not nil",
			func(scope *Scope, t *testing.T) {
				assert.NotNil(t, scope.Owners)
			},
		},
		{
			"Owners",
			"length is 0",
			func(scope *Scope, t *testing.T) {
				assert.Equal(t, 0, len(scope.Owners))
			},
		},
		{
			"DataAccess",
			"is not nil",
			func(scope *Scope, t *testing.T) {
				assert.NotNil(t, scope.DataAccess)
			},
		},
		{
			"DataAccess",
			"length is 0",
			func(scope *Scope, t *testing.T) {
				assert.Equal(t, 0, len(scope.DataAccess))
			},
		},
		{
			"ValueOwnerAddress",
			"is empty string",
			func(scope *Scope, t *testing.T) {
				assert.Equal(t, "", scope.ValueOwnerAddress)
			},
		},
	}

	for i, tc := range tests {
		name := tc.name
		if len(tc.field) > 0 {
			name = fmt.Sprintf("%s %s", tc.field, tc.name)
		}
		s.T().Run(fmt.Sprintf("%d %s", i, name), func(t *testing.T) {
			scope := emptyScope()
			tc.test(scope, t)
			if len(tc.field) > 0 {
				testedFields = appendUnique(testedFields, tc.field)
			}
		})
	}

	s.T().Run("All fields are tested", func(t *testing.T) {
		allStructFields := getPublicFieldNames(Scope{})
		assertSetsAreEqual(t, testedFields, allStructFields)
	})
}

func (s *P8eTestSuite) TestEmptySession() {
	testedFields := []string{}
	tests := []struct {
		field string
		name  string
		test  func(session *Session, t *testing.T)
	}{
		{
			"",
			"Does not return nil",
			func(session *Session, t *testing.T) {
				assert.NotNil(t, session)
			},
		},
		{
			"SessionId",
			"is empty",
			func(session *Session, t *testing.T) {
				assert.True(t, session.SessionId.Empty())
			},
		},
		{
			"SpecificationId",
			"is empty",
			func(session *Session, t *testing.T) {
				assert.True(t, session.SpecificationId.Empty())
			},
		},
		{
			"Parties",
			"is not nil",
			func(session *Session, t *testing.T) {
				assert.NotNil(t, session.Parties)
			},
		},
		{
			"Parties",
			"length is 0",
			func(session *Session, t *testing.T) {
				assert.Equal(t, 0, len(session.Parties))
			},
		},
		{
			"Name",
			"is empty string",
			func(session *Session, t *testing.T) {
				assert.Equal(t, "", session.Name)
			},
		},
		{
			"Audit",
			"is nil",
			func(session *Session, t *testing.T) {
				assert.Nil(t, session.Audit)
			},
		},
		{
			"Context",
			"is nil",
			func(session *Session, t *testing.T) {
				assert.Nil(t, session.Context)
			},
		},
	}

	for i, tc := range tests {
		name := tc.name
		if len(tc.field) > 0 {
			name = fmt.Sprintf("%s %s", tc.field, tc.name)
		}
		s.T().Run(fmt.Sprintf("%d %s", i, name), func(t *testing.T) {
			session := emptySession()
			tc.test(session, t)
			if len(tc.field) > 0 {
				testedFields = appendUnique(testedFields, tc.field)
			}
		})
	}

	s.T().Run("All fields are tested", func(t *testing.T) {
		allStructFields := getPublicFieldNames(Session{})
		assertSetsAreEqual(t, testedFields, allStructFields)
	})
}

func (s *P8eTestSuite) TestEmptyRecord() {
	testedFields := []string{}
	tests := []struct {
		field string
		name  string
		test  func(record *Record, t *testing.T)
	}{
		{
			"",
			"Does not return nil",
			func(record *Record, t *testing.T) {
				assert.NotNil(t, record)
			},
		},
		{
			"Name",
			"is empty string",
			func(record *Record, t *testing.T) {
				assert.Equal(t, "", record.Name, "Name")
			},
		},
		{
			"SessionId",
			"is empty",
			func(record *Record, t *testing.T) {
				assert.True(t, record.SessionId.Empty(), "SessionId")
			},
		},
		{
			"Process",
			"is empty",
			func(record *Record, t *testing.T) {
				assert.Equal(t, *emptyProcess(), record.Process, "Process")
			},
		},
		{
			"Inputs",
			"is not nil",
			func(record *Record, t *testing.T) {
				assert.NotNil(t, record.Inputs, "Inputs")
			},
		},
		{
			"Inputs",
			"length is 0",
			func(record *Record, t *testing.T) {
				assert.Equal(t, 0, len(record.Inputs), "Inputs")
			},
		},
		{
			"Outputs",
			"is not nil",
			func(record *Record, t *testing.T) {
				assert.NotNil(t, record.Outputs, "Outputs")
			},
		},
		{
			"Outputs",
			"length is 0",
			func(record *Record, t *testing.T) {
				assert.Equal(t, 0, len(record.Outputs), "Outputs")
			},
		},
		{
			"SpecificationId",
			"is empty",
			func(record *Record, t *testing.T) {
				assert.True(t, record.SpecificationId.Empty(), "SpecificationId")
			},
		},
	}

	for i, tc := range tests {
		name := tc.name
		if len(tc.field) > 0 {
			name = fmt.Sprintf("%s %s", tc.field, tc.name)
		}
		s.T().Run(fmt.Sprintf("%d %s", i, name), func(t *testing.T) {
			record := emptyRecord()
			tc.test(record, t)
			if len(tc.field) > 0 {
				testedFields = appendUnique(testedFields, tc.field)
			}
		})
	}

	s.T().Run("All fields are tested", func(t *testing.T) {
		allStructFields := getPublicFieldNames(Record{})
		assertSetsAreEqual(t, testedFields, allStructFields)
	})
}

func (s *P8eTestSuite) TestEmptyProcess() {
	testedFields := []string{}
	tests := []struct {
		field string
		name  string
		test  func(record *Process, t *testing.T)
	}{
		{
			"",
			"Does not return nil",
			func(process *Process, t *testing.T) {
				assert.NotNil(t, process)
			},
		},
		{
			"ProcessId",
			"is nil",
			func(process *Process, t *testing.T) {
				assert.Nil(t, process.ProcessId, "ProcessId")
			},
		},
		{
			"Name",
			"is empty string",
			func(process *Process, t *testing.T) {
				assert.Equal(t, "", process.Name, "Name")
			},
		},
		{
			"Method",
			"is empty string",
			func(process *Process, t *testing.T) {
				assert.Equal(t, "", process.Method, "Method")
			},
		},
	}

	for i, tc := range tests {
		name := tc.name
		if len(tc.field) > 0 {
			name = fmt.Sprintf("%s %s", tc.field, tc.name)
		}
		s.T().Run(fmt.Sprintf("%d %s", i, name), func(t *testing.T) {
			process := emptyProcess()
			tc.test(process, t)
			if len(tc.field) > 0 {
				testedFields = appendUnique(testedFields, tc.field)
			}
		})
	}

	s.T().Run("All fields are tested", func(t *testing.T) {
		allStructFields := getPublicFieldNames(Process{})
		assertSetsAreEqual(t, testedFields, allStructFields)
	})
}

func (s *P8eTestSuite) TestConvertP8eMemorializeContractRequest() {
	scopeUUID := uuid.New()
	scopeID := ScopeMetadataAddress(scopeUUID)
	sessionUUID := uuid.New()
	sessionID := SessionMetadataAddress(scopeUUID, sessionUUID)

	tests := []struct {
		name     string
		req      MsgP8EMemorializeContractRequest
		p8EData  P8EData
		errorMsg string
	}{
		{
			"valid conversion",
			MsgP8EMemorializeContractRequest{
				ScopeId:              scopeID.String(),
				GroupId:              sessionID.String(),
				ScopeSpecificationId: "", // TODO
				Recitals: &p8e.Recitals{
					Parties: []*p8e.Recital{
						{
							SignerRole: p8e.PartyType_PARTY_TYPE_OWNER,
							Signer: &p8e.SigningAndEncryptionPublicKeys{
								SigningPublicKey: &p8e.PublicKey{
									PublicKeyBytes: s.pubkey1.Bytes(),
									Type:           p8e.PublicKeyType_ELLIPTIC,
									Curve:          p8e.PublicKeyCurve_SECP256K1,
								},
								EncryptionPublicKey: nil,
							},
							Address: s.user1Addr,
						},
					},
				},
				Contract: &p8e.Contract{
					Definition: nil,
					Spec: &p8e.Fact{
						Name: "", // TODO
						DataLocation: &p8e.Location{
							Ref:       nil, // TODO
							Classname: "",
						},
					},
					Invoker: &p8e.SigningAndEncryptionPublicKeys{
						SigningPublicKey: &p8e.PublicKey{
							PublicKeyBytes: s.pubkey1.Bytes(),
							Type:           p8e.PublicKeyType_ELLIPTIC,
							Curve:          p8e.PublicKeyCurve_SECP256K1,
						},
						EncryptionPublicKey: nil,
					},
					Inputs:         []*p8e.Fact{},          // TODO
					Conditions:     []*p8e.Condition{},     // TODO
					Considerations: []*p8e.Consideration{}, // TODO
					Recitals:       []*p8e.Recital{},       // TODO
					TimesExecuted:  1,
					StartTime:      nil,
				},
				Signatures: &p8e.SignatureSet{
					Signatures: []*p8e.Signature{
						{
							Algo:      "",
							Provider:  "",
							Signature: "",
							Signer: &p8e.SigningAndEncryptionPublicKeys{
								SigningPublicKey: &p8e.PublicKey{
									PublicKeyBytes: s.pubkey1.Bytes(),
									Type:           p8e.PublicKeyType_ELLIPTIC,
									Curve:          p8e.PublicKeyCurve_SECP256K1,
								},
								EncryptionPublicKey: nil,
							},
						},
					},
				},
				Invoker: s.user1,
			},
			P8EData{
				Scope:      nil,        // TODO
				Session:    nil,        // TODO
				RecordReqs: nil,        // TODO
				Signers:    []string{}, // TODO
			},
			"",
		},
		// TODO: Valid call
		// TODO: convertParties fails
		// TODO: parseScopeID fails
		// TODO: getValueOwner fails
		// TODO: getValueOwner fails
		// TODO: parseSessionID fails
		// TODO: getContractSpecID fails
		// TODO: convertSigners fails
	}

	tests = tests[:0] // TODO: Remove this line once tests are actually ready.

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			p8eData, err := ConvertP8eMemorializeContractRequest(&tc.req)
			if len(tc.errorMsg) > 0 {
				assert.EqualError(t, err, tc.errorMsg, "expected error")
			} else {
				require.NoError(t, err, "unexpected error")
				assert.Equal(t, tc.p8EData.Scope, p8eData.Scope, "scope")
				assert.Equal(t, tc.p8EData.Session, p8eData.Session, "session")
				assert.Equal(t, tc.p8EData.RecordReqs, p8eData.RecordReqs, "recordReqs")
				assert.Equal(t, tc.p8EData.Signers, p8eData.Signers, "signers")
			}
		})
	}
}

func (s *P8eTestSuite) TestParsePublicKey() {
	tests := []struct {
		name      string
		key       string
		expectErr bool
	}{
		{
			"should parse uncompressed public key",
			"BGxX6eJRAdXlU64APi95Al44m1FJVgfHlrTpXAqUAB+8JNhM0HgIGWElKbgD6K0KOX9HTJZdlX0z3WTmQrdW+8Q=",
			false,
		},
		{
			"should parse compressed public key",
			"AmxX6eJRAdXlU64APi95Al44m1FJVgfHlrTpXAqUAB+8=",
			false,
		},
		{
			"should fail to parse incorrect public key size",
			"4m1FJVgfHlrTpXAqUAB+8",
			true,
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			b, _ := b64.StdEncoding.DecodeString(tc.key)
			pubKey, addr, err := parsePublicKey(b)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NotNil(t, pubKey, "should have successfully created public key")
				assert.NotNil(t, addr, "should have successfully account address")
			}
		})
	}

}
