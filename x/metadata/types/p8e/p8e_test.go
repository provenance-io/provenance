package p8e

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/provenance-io/provenance/x/metadata/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type p8eTestSuite struct {
	suite.Suite
}

func TestP8eTestSuite(t *testing.T) {
	suite.Run(t, new(p8eTestSuite))
}

func (s *p8eTestSuite) SetupSuite() {
	s.T().Parallel()
}

// getFieldNames gets all the field names in the provided struct.
// Panics if the provided thing isn't a struct.
func getFieldNames(thing interface{}) []string {
	retval := []string{}
	t := reflect.TypeOf(thing)
	fieldCount := t.NumField()
	for i := 0; i < fieldCount; i++ {
		retval = append(retval, t.Field(i).Name)
	}
	return retval;
}

// assertSetsAreEqual asserts that all of the elements in expected are in actual.
// Assumes elements in the sets are unique.
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

func (s *p8eTestSuite) TestEmptyScope() {
	testedFields := []string{}
	tests := []struct {
		field string
		name string
		test func(scope *types.Scope, t *testing.T)
	} {
		{
			"",
			"Does not return nil",
			func(scope *types.Scope, t *testing.T) {
				assert.NotNil(t, scope)
			},
		},
		{
			"ScopeId",
			"is empty",
			func(scope *types.Scope, t *testing.T) {
				assert.True(t, scope.ScopeId.Empty())
			},
		},
		{
			"SpecificationId",
			"is empty",
			func(scope *types.Scope, t *testing.T) {
				assert.True(t, scope.SpecificationId.Empty())
			},
		},
		{
			"Owners",
			"is not nil",
			func(scope *types.Scope, t *testing.T) {
				assert.NotNil(t, scope.Owners)
			},
		},
		{
			"Owners",
			"length is 0",
			func(scope *types.Scope, t *testing.T) {
				assert.Equal(t, 0, len(scope.Owners))
			},
		},
		{
			"DataAccess",
			"is not nil",
			func(scope *types.Scope, t *testing.T) {
				assert.NotNil(t, scope.DataAccess)
			},
		},
		{
			"DataAccess",
			"length is 0",
			func(scope *types.Scope, t *testing.T) {
				assert.Equal(t, 0, len(scope.DataAccess))
			},
		},
		{
			"ValueOwnerAddress",
			"is empty string",
			func(scope *types.Scope, t *testing.T) {
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
			scope := EmptyScope()
			tc.test(scope, t)
			if len(tc.field) > 0 {
				testedFields = appendUnique(testedFields, tc.field)
			}
		})
	}

	s.T().Run("All fields are tested", func(t *testing.T) {
		allStructFields := getFieldNames(types.Scope{})
		assertSetsAreEqual(t, testedFields, allStructFields)
	})
}

func (s *p8eTestSuite) TestEmptySession() {
	testedFields := []string{}
	tests := []struct {
		field string
		name string
		test func(session *types.Session, t *testing.T)
	} {
		{
			"",
			"Does not return nil",
			func(session *types.Session, t *testing.T) {
				assert.NotNil(t, session)
			},
		},
		{
			"SessionId",
			"is empty",
			func(session *types.Session, t *testing.T) {
				assert.True(t, session.SessionId.Empty())
			},
		},
		{
			"SpecificationId",
			"is empty",
			func(session *types.Session, t *testing.T) {
				assert.True(t, session.SpecificationId.Empty())
			},
		},
		{
			"Parties",
			"is not nil",
			func(session *types.Session, t *testing.T) {
				assert.NotNil(t, session.Parties)
			},
		},
		{
			"Parties",
			"length is 0",
			func(session *types.Session, t *testing.T) {
				assert.Equal(t, 0, len(session.Parties))
			},
		},
		{
			"Name",
			"is empty string",
			func(session *types.Session, t *testing.T) {
				assert.Equal(t, "", session.Name)
			},
		},
		{
			"Audit",
			"is nil",
			func(session *types.Session, t *testing.T) {
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
			session := EmptySession()
			tc.test(session, t)
			if len(tc.field) > 0 {
				testedFields = appendUnique(testedFields, tc.field)
			}
		})
	}

	s.T().Run("All fields are tested", func(t *testing.T) {
		allStructFields := getFieldNames(types.Session{})
		assertSetsAreEqual(t, testedFields, allStructFields)
	})
}

func (s *p8eTestSuite) TestEmptyRecord() {
	testedFields := []string{}
	tests := []struct {
		field string
		name string
		test func(record *types.Record, t *testing.T)
	} {
		{
			"",
			"Does not return nil",
			func(record *types.Record, t *testing.T) {
				assert.NotNil(t, record)
			},
		},
		{
			"Name",
			"is empty string",
			func(record *types.Record, t *testing.T) {
				assert.Equal(t, "", record.Name, "Name")
			},
		},
		{
			"SessionId",
			"is empty",
			func(record *types.Record, t *testing.T) {
				assert.True(t, record.SessionId.Empty(), "SessionId")
			},
		},
		{
			"Process",
			"is empty",
			func(record *types.Record, t *testing.T) {
				assert.Equal(t, *EmptyProcess(), record.Process, "Process")
			},
		},
		{
			"Inputs",
			"is not nil",
			func(record *types.Record, t *testing.T) {
				assert.NotNil(t, record.Inputs, "Inputs")
			},
		},
		{
			"Inputs",
			"length is 0",
			func(record *types.Record, t *testing.T) {
				assert.Equal(t, 0, len(record.Inputs), "Inputs")
			},
		},
		{
			"Outputs",
			"is not nil",
			func(record *types.Record, t *testing.T) {
				assert.NotNil(t, record.Outputs, "Outputs")
			},
		},
		{
			"Outputs",
			"length is 0",
			func(record *types.Record, t *testing.T) {
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
			record := EmptyRecord()
			tc.test(record, t)
			if len(tc.field) > 0 {
				testedFields = appendUnique(testedFields, tc.field)
			}
		})
	}

	s.T().Run("All fields are tested", func(t *testing.T) {
		allStructFields := getFieldNames(types.Record{})
		assertSetsAreEqual(t, testedFields, allStructFields)
	})
}

func (s *p8eTestSuite) TestEmptyProcess() {
	testedFields := []string{}
	tests := []struct {
		field string
		name string
		test func(record *types.Process, t *testing.T)
	} {
		{
			"",
			"Does not return nil",
			func(process *types.Process, t *testing.T) {
				assert.NotNil(t, process)
			},
		},
		{
			"ProcessId",
			"is nil",
			func(process *types.Process, t *testing.T) {
				assert.Nil(t, process.ProcessId, "ProcessId")
			},
		},
		{
			"Name",
			"is empty string",
			func(process *types.Process, t *testing.T) {
				assert.Equal(t, "", process.Name, "Name")
			},
		},
		{
			"Method",
			"is empty string",
			func(process *types.Process, t *testing.T) {
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
			process := EmptyProcess()
			tc.test(process, t)
			if len(tc.field) > 0 {
				testedFields = appendUnique(testedFields, tc.field)
			}
		})
	}

	s.T().Run("All fields are tested", func(t *testing.T) {
		allStructFields := getFieldNames(types.Process{})
		assertSetsAreEqual(t, testedFields, allStructFields)
	})
}
