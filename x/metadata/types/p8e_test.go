package types

import (


)
import (
	"fmt"
	"reflect"
	"testing"
	"unicode"

	"github.com/provenance-io/provenance/x/metadata/types/p8e"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
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
	return retval;
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
	for  thingType.Kind() == reflect.Ptr {
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
			fields  = append(fields, thingType.Field(i).Name)
		}
	}
	return
}

// assertSetsAreEqual asserts that all of the elements in expected are in actual.
// Assumes elements in a set are unique.
func assertSetsAreEqual(t *testing.T, expected []string, actual []string) bool {
	allPass := assert.Equal(t, len(expected), len(actual), "lengths")
	if (allPass) {
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
	return p8e.ContractSpec{ConsiderationSpecs: []*p8e.ConsiderationSpec{
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
	invalidDefSpecHash := createDefinitionSpec("ExampleContract", "io.provenance.contracts.ExampleContract", p8e.ProvenanceReference{Hash: "should fail to decode this"}, 1)

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
		"should fail to decode resource location": {
			createContractSpec([]*p8e.DefinitionSpec{&validDefSpec}, p8e.OutputSpec{Spec: &validDefSpec}, invalidDefSpecHash),
			[]string{s.user1},
			true,
			"illegal base64 data at input byte 6",
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
		name string
		test func(scope *Scope, t *testing.T)
	} {
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
		name string
		test func(session *Session, t *testing.T)
	} {
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
		name string
		test func(record *Record, t *testing.T)
	} {
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
		name string
		test func(record *Process, t *testing.T)
	} {
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

// TODO: ConvertP8eMemorializeContractRequest tests
