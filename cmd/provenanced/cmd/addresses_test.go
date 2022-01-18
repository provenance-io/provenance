package cmd_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/provenance-io/provenance/cmd/provenanced/cmd"
)

type MetaaddressTestSuite struct {
	suite.Suite

	// Pre-selected UUID strings that go with ID strings generated from the Go code.
	scopeUUIDStr          string
	sessionUUIDStr        string
	scopeSpecUUIDStr      string
	contractSpecUUIDStr   string
	recordName            string
	recordNameHashedBytes []byte
	recordNameHashedHex   string

	// Pre-generated ID strings created using Go code and providing the above strings.
	scopeIDStr        string
	sessionIDStr      string
	recordIDStr       string
	scopeSpecIDStr    string
	contractSpecIDStr string
	recordSpecIDStr   string

	// UUID versions of the UUID strings.
	scopeUUID        uuid.UUID
	sessionUUID      uuid.UUID
	scopeSpecUUID    uuid.UUID
	contractSpecUUID uuid.UUID
}

func (s *MetaaddressTestSuite) SetupTest() {
	// These strings come from the output of x/metadata/types/address_test.go TestGenerateExamples().

	s.scopeUUIDStr = "91978ba2-5f35-459a-86a7-feca1b0512e0"
	s.sessionUUIDStr = "5803f8bc-6067-4eb5-951f-2121671c2ec0"
	s.scopeSpecUUIDStr = "dc83ea70-eacd-40fe-9adf-1cf6148bf8a2"
	s.contractSpecUUIDStr = "def6bc0a-c9dd-4874-948f-5206e6060a84"
	s.recordName = "recordname"
	s.recordNameHashedBytes = []byte{234, 169, 160, 84, 154, 205, 183, 162, 227, 133, 142, 181, 183, 185, 209, 190}
	s.recordNameHashedHex = hex.EncodeToString(s.recordNameHashedBytes)

	s.scopeIDStr = "scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel"
	s.sessionIDStr = "session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr"
	s.recordIDStr = "record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3"
	s.scopeSpecIDStr = "scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m"
	s.contractSpecIDStr = "contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn"
	s.recordSpecIDStr = "recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44"

	s.scopeUUID = uuid.MustParse(s.scopeUUIDStr)
	s.sessionUUID = uuid.MustParse(s.sessionUUIDStr)
	s.scopeSpecUUID = uuid.MustParse(s.scopeSpecUUIDStr)
	s.contractSpecUUID = uuid.MustParse(s.contractSpecUUIDStr)
}

func TestMetaaddressTestSuite(t *testing.T) {
	suite.Run(t, new(MetaaddressTestSuite))
}

func (s MetaaddressTestSuite) TestAddMetaAddressDecoder() {
	command := cmd.AddMetaAddressDecoder()

	tests := []struct {
		name     string
		args     []string
		inResult []string
		err      string
	}{
		{
			name: "valid scope",
			args: []string{s.scopeIDStr},
			inResult: []string{
				"Type: Scope",
				fmt.Sprintf("Scope UUID: %s", s.scopeUUIDStr),
			},
		},
		{
			name: "valid session",
			args: []string{s.sessionIDStr},
			inResult: []string{
				"Type: Session",
				fmt.Sprintf("Scope Id: %s", s.scopeIDStr),
				fmt.Sprintf("Scope UUID: %s", s.scopeUUIDStr),
				fmt.Sprintf("Session UUID: %s", s.sessionUUIDStr),
			},
		},
		{
			name: "valid record",
			args: []string{s.recordIDStr},
			inResult: []string{
				"Type: Record",
				fmt.Sprintf("Scope Id: %s", s.scopeIDStr),
				fmt.Sprintf("Scope UUID: %s", s.scopeUUIDStr),
				fmt.Sprintf("Name Hash (hex): %s", s.recordNameHashedHex),
			},
		},
		{
			name: "valid scope specification",
			args: []string{s.scopeSpecIDStr},
			inResult: []string{
				"Type: Scope Specification",
				fmt.Sprintf("Scope Specification UUID: %s", s.scopeSpecUUIDStr),
			},
		},
		{
			name: "valid contract specification",
			args: []string{s.contractSpecIDStr},
			inResult: []string{
				"Type: Contract Specification",
				fmt.Sprintf("Contract Specification UUID: %s", s.contractSpecUUIDStr),
			},
		},
		{
			name: "valid record specification",
			args: []string{s.recordSpecIDStr},
			inResult: []string{
				"Type: Record Specification",
				fmt.Sprintf("Contract Specification Id: %s", s.contractSpecIDStr),
				fmt.Sprintf("Contract Specification UUID: %s", s.contractSpecUUIDStr),
				fmt.Sprintf("Name Hash (hex): %s", s.recordNameHashedHex),
			},
		},
		{
			name: "no args",
			args: []string{},
			err:  "accepts 1 arg(s), received 0",
		},
		{
			name: "two args",
			args: []string{s.scopeIDStr, s.sessionIDStr},
			err:  "accepts 1 arg(s), received 2",
		},
		{
			name: "invalid address",
			args: []string{s.scopeIDStr + "bad"},
			err:  "decoding bech32 failed: invalid character not part of charset: 98",
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			command.SetArgs(tc.args)
			b := bytes.NewBufferString("")
			command.SetOut(b)
			err := command.Execute()
			if len(tc.err) > 0 {
				require.EqualErrorf(t, err, tc.err, "%s - expected error", command.Name())
			} else {
				require.NoErrorf(t, err, "%s - unexpected error", command.Name())
				out, err := ioutil.ReadAll(b)
				require.NoError(t, err, "%s - unexpected buffer read error", command.Name())
				outStr := string(out)
				for _, str := range tc.inResult {
					assert.Containsf(t, outStr, str, "%s - expected value to be in output", command.Name())
				}
			}
		})
	}
}

func (s MetaaddressTestSuite) TestAddMetaAddressEncoder() {
	command := cmd.AddMetaAddressEncoder()

	tests := []struct {
		name     string
		args     []string
		inResult []string
		err      string
	}{
		// Generic invalid cases
		{
			name: "no args",
			args: []string{},
			err:  "accepts between 2 and 3 arg(s), received 0",
		},
		{
			name: "one arg",
			args: []string{"one"},
			err:  "accepts between 2 and 3 arg(s), received 1",
		},
		{
			name: "four args",
			args: []string{"one", "two", "three", "four"},
			err:  "accepts between 2 and 3 arg(s), received 4",
		},
		{
			name: "invalid primary uuid",
			args: []string{"scope", "not-a-uuid"},
			err:  "invalid UUID length: 10",
		},
		{
			name: "invalid type",
			args: []string{"not-a-type", s.scopeUUIDStr},
			err:  fmt.Sprintf("unknown type: %s, Supported types: scope session record scope-specification contract-specification record-specification", "not-a-type"),
		},
		// Scope cases
		{
			name:     "scope valid",
			args:     []string{"scope", s.scopeUUIDStr},
			inResult: []string{s.scopeIDStr},
		},
		{
			name:     "Scope valid",
			args:     []string{"Scope", s.scopeUUIDStr},
			inResult: []string{s.scopeIDStr},
		},
		{
			name: "scope invalid has extra param",
			args: []string{"scope", s.scopeUUIDStr, "bad-arg"},
			err:  "too many arguments for scope address encoder",
		},
		// Session cases
		{
			name:     "session valid",
			args:     []string{"session", s.scopeUUIDStr, s.sessionUUIDStr},
			inResult: []string{s.sessionIDStr},
		},
		{
			name:     "Session valid",
			args:     []string{"Session", s.scopeUUIDStr, s.sessionUUIDStr},
			inResult: []string{s.sessionIDStr},
		},
		{
			name: "session invalid missing param",
			args: []string{"session", s.scopeUUIDStr},
			err:  "not enough arguments for session address encoder",
		},
		{
			name: "session invalid second uuid",
			args: []string{"session", s.scopeUUIDStr, "bad-arg"},
			err:  "invalid UUID length: 7",
		},
		// Record cases
		{
			name:     "record valid",
			args:     []string{"record", s.scopeUUIDStr, s.recordName},
			inResult: []string{s.recordIDStr},
		},
		{
			name:     "Record valid",
			args:     []string{"Record", s.scopeUUIDStr, s.recordName},
			inResult: []string{s.recordIDStr},
		},
		{
			name: "record invalid missing param",
			args: []string{"record", s.scopeUUIDStr},
			err:  "not enough arguments for record address encoder",
		},
		{
			name: "record invalid empty name param",
			args: []string{"record", s.scopeUUIDStr, ""},
			err:  "not enough arguments for record address encoder",
		},
		// Scope Specification cases
		{
			name:     "scope-specification valid",
			args:     []string{"scope-specification", s.scopeSpecUUIDStr},
			inResult: []string{s.scopeSpecIDStr},
		},
		{
			name:     "ScopeSpecification valid",
			args:     []string{"ScopeSpecification", s.scopeSpecUUIDStr},
			inResult: []string{s.scopeSpecIDStr},
		},
		{
			name:     "scope-spec valid",
			args:     []string{"scope-spec", s.scopeSpecUUIDStr},
			inResult: []string{s.scopeSpecIDStr},
		},
		{
			name:     "ScopeSpec valid",
			args:     []string{"ScopeSpec", s.scopeSpecUUIDStr},
			inResult: []string{s.scopeSpecIDStr},
		},
		{
			name: "scope-specification invalid has extra param",
			args: []string{"scope-specification", s.scopeSpecUUIDStr, "bad-arg"},
			err:  "too many arguments for scope-specification address encoder",
		},
		// Contract Specification cases
		{
			name:     "contract-specification valid",
			args:     []string{"contract-specification", s.contractSpecUUIDStr},
			inResult: []string{s.contractSpecIDStr},
		},
		{
			name:     "ContractSpecification valid",
			args:     []string{"ContractSpecification", s.contractSpecUUIDStr},
			inResult: []string{s.contractSpecIDStr},
		},
		{
			name:     "contract-spec valid",
			args:     []string{"contract-spec", s.contractSpecUUIDStr},
			inResult: []string{s.contractSpecIDStr},
		},
		{
			name:     "ContractSpec valid",
			args:     []string{"ContractSpec", s.contractSpecUUIDStr},
			inResult: []string{s.contractSpecIDStr},
		},
		{
			name:     "cspec valid",
			args:     []string{"cspec", s.contractSpecUUIDStr},
			inResult: []string{s.contractSpecIDStr},
		},
		{
			name: "contract-specification invalid has extra param",
			args: []string{"contract-specification", s.contractSpecUUIDStr, "bad-arg"},
			err:  "too many arguments for contract-specification address encoder",
		},
		// Record Specification cases
		{
			name:     "record-specification valid",
			args:     []string{"record-specification", s.contractSpecUUIDStr, s.recordName},
			inResult: []string{s.recordSpecIDStr},
		},
		{
			name:     "RecordSpecification valid",
			args:     []string{"RecordSpecification", s.contractSpecUUIDStr, s.recordName},
			inResult: []string{s.recordSpecIDStr},
		},
		{
			name:     "record-spec valid",
			args:     []string{"record-spec", s.contractSpecUUIDStr, s.recordName},
			inResult: []string{s.recordSpecIDStr},
		},
		{
			name:     "RecordSpec valid",
			args:     []string{"RecordSpec", s.contractSpecUUIDStr, s.recordName},
			inResult: []string{s.recordSpecIDStr},
		},
		{
			name:     "rec-spec valid",
			args:     []string{"rec-spec", s.contractSpecUUIDStr, s.recordName},
			inResult: []string{s.recordSpecIDStr},
		},
		{
			name:     "RecSpec valid",
			args:     []string{"RecSpec", s.contractSpecUUIDStr, s.recordName},
			inResult: []string{s.recordSpecIDStr},
		},
		{
			name: "record-specification invalid missing param",
			args: []string{"record-specification", s.contractSpecUUIDStr},
			err:  "not enough arguments for record-specification address encoder",
		},
		{
			name: "record-specification invalid empty name param",
			args: []string{"record-specification", s.contractSpecUUIDStr, ""},
			err:  "not enough arguments for record-specification address encoder",
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			command.SetArgs(tc.args)
			b := bytes.NewBufferString("")
			command.SetOut(b)
			err := command.Execute()
			if len(tc.err) > 0 {
				require.EqualErrorf(t, err, tc.err, "%s - expected error", command.Name())
			} else {
				require.NoErrorf(t, err, "%s - unexpected error", command.Name())
				out, err := ioutil.ReadAll(b)
				require.NoError(t, err, "%s - unexpected buffer read error", command.Name())
				outStr := string(out)
				for _, str := range tc.inResult {
					assert.Containsf(t, outStr, str, "%s - expected value to be in output", command.Name())
				}
			}
		})
	}
}
