package cli

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/metadata/types"
)

func nameFromInput(arg string) string {
	if len(arg) > 0 {
		return arg
	}
	return "empty"
}

// AssertErrorValue asserts that:
//   - If errorString is empty, theError must be nil
//   - If errorString is not empty, theError must equal the errorString.
func AssertErrorValue(t *testing.T, theError error, errorString string, msgAndArgs ...interface{}) bool {
	t.Helper()
	if len(errorString) > 0 {
		return assert.EqualError(t, theError, errorString, msgAndArgs...)
	}
	return assert.NoError(t, theError, msgAndArgs...)
}

func TestParseProcess(t *testing.T) {
	mdAddr := types.RecordMetadataAddress(uuid.New(), "record name").String()
	tests := []struct {
		name    string
		arg     string
		expProc types.Process
		expErr  string
	}{
		{
			name: "control with hash",
			arg:  "name,hashhash,method",
			expProc: types.Process{
				ProcessId: &types.Process_Hash{Hash: "hashhash"},
				Name:      "name",
				Method:    "method",
			},
			expErr: "",
		},
		{
			name: "control with address",
			arg:  "name," + mdAddr + ",method",
			expProc: types.Process{
				ProcessId: &types.Process_Address{Address: mdAddr},
				Name:      "name",
				Method:    "method",
			},
			expErr: "",
		},
		{
			name: "empty name",
			arg:  ",hash,method",
			expProc: types.Process{
				ProcessId: &types.Process_Hash{Hash: "hash"},
				Name:      "",
				Method:    "method",
			},
			expErr: "",
		},
		{
			name: "empty method",
			arg:  "name,hash,",
			expProc: types.Process{
				ProcessId: &types.Process_Hash{Hash: "hash"},
				Name:      "name",
				Method:    "",
			},
			expErr: "",
		},
		{
			name:    "no commas",
			arg:     "name hash method",
			expProc: types.Process{},
			expErr:  `invalid process "name hash method": expected 3 parts, have: 1`,
		},
		{
			name:    "only 2 parts",
			arg:     "parta,partb",
			expProc: types.Process{},
			expErr:  `invalid process "parta,partb": expected 3 parts, have: 2`,
		},
		{
			name:    "4 parts",
			arg:     "partv,partw,partx,party",
			expProc: types.Process{},
			expErr:  `invalid process "partv,partw,partx,party": expected 3 parts, have: 4`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			proc, err := parseProcess(tc.arg)
			AssertErrorValue(t, err, tc.expErr, "parseProcess(%q) error", tc.arg)
			assert.Equal(t, tc.expProc, proc, "parseProcess(%q) process", tc.arg)
		})
	}
}

func TestParseParties(t *testing.T) {
	pt := func(addr string, role types.PartyType, optional bool) types.Party {
		return types.Party{Address: addr, Role: role, Optional: optional}
	}
	ptz := func(parties ...types.Party) []types.Party {
		rv := make([]types.Party, 0, len(parties))
		rv = append(rv, parties...)
		return rv
	}
	addr1 := sdk.AccAddress("address_1___________").String()
	addr2 := sdk.AccAddress("address_2___________").String()
	addr3 := sdk.AccAddress("address_3___________").String()

	tests := []struct {
		name   string
		arg    string
		exp    []types.Party
		expErr string
	}{
		{
			name:   "empty string",
			arg:    "",
			exp:    nil,
			expErr: "",
		},
		{
			name:   "one party: good",
			arg:    addr1 + ",investor",
			exp:    ptz(pt(addr1, types.PartyType_PARTY_TYPE_INVESTOR, false)),
			expErr: "",
		},
		{
			name:   "one party: bad",
			arg:    addr1 + ",bad",
			exp:    nil,
			expErr: fmt.Sprintf(`invalid party "%s,bad": unknown party type: "bad"`, addr1),
		},
		{
			name: "three parties: all good",
			arg:  addr1 + ";" + addr2 + ",provenance;" + addr3 + ",controller,opt",
			exp: ptz(
				pt(addr1, types.PartyType_PARTY_TYPE_OWNER, false),
				pt(addr2, types.PartyType_PARTY_TYPE_PROVENANCE, false),
				pt(addr3, types.PartyType_PARTY_TYPE_CONTROLLER, true),
			),
			expErr: "",
		},
		{
			name:   "three parties: last one bad",
			arg:    addr1 + ";" + addr2 + ",provenance;" + addr3 + ", controller,opt",
			exp:    nil,
			expErr: fmt.Sprintf(`invalid party "%s, controller,opt": unknown party type: " controller"`, addr3),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := parseParties(tc.arg)
			AssertErrorValue(t, err, tc.expErr, "parseParties(%q)", tc.arg)
			assert.Equal(t, tc.exp, actual, "parseParties(%q)", tc.arg)
		})
	}
}

func TestParseParty(t *testing.T) {
	pt := func(addr string, role types.PartyType, optional bool) types.Party {
		return types.Party{Address: addr, Role: role, Optional: optional}
	}
	addr := sdk.AccAddress("test_address________").String()
	owner := types.PartyType_PARTY_TYPE_OWNER

	tests := []struct {
		name   string
		arg    string
		exp    types.Party
		expErr string
	}{
		{
			name:   "empty string",
			arg:    "",
			exp:    pt("", owner, false),
			expErr: "cannot be empty",
		},
		{
			name:   "four parts",
			arg:    "a,b,c,d",
			exp:    pt("", owner, false),
			expErr: "expected 1, 2, or 3 parts, have: 4",
		},
		{
			name:   "one arg: good",
			arg:    addr,
			exp:    pt(addr, owner, false),
			expErr: "",
		},
		{
			name:   "one arg: bad",
			arg:    "not-an-address",
			exp:    pt("", owner, false),
			expErr: `invalid address "not-an-address": decoding bech32 failed: invalid separator index -1`,
		},
		{
			name:   "two args: addr,role",
			arg:    addr + ",servicer",
			exp:    pt(addr, types.PartyType_PARTY_TYPE_SERVICER, false),
			expErr: "",
		},
		{
			name:   "two args: addr,opt",
			arg:    addr + ",opt",
			exp:    pt(addr, owner, true),
			expErr: "",
		},
		{
			name:   "two args: addr,req",
			arg:    addr + ",req",
			exp:    pt(addr, owner, false),
			expErr: "",
		},
		{
			name:   "two args: bad,role",
			arg:    "bad,validator",
			exp:    pt("", owner, false),
			expErr: `invalid address "bad": decoding bech32 failed: invalid bech32 string length 3`,
		},
		{
			name:   "two args: addr,bad",
			arg:    addr + ",bad",
			exp:    pt(addr, owner, false),
			expErr: `unknown party type: "bad"`,
		},
		{
			name:   "three args: addr,role,opt",
			arg:    addr + ",omnibus,opt",
			exp:    pt(addr, types.PartyType_PARTY_TYPE_OMNIBUS, true),
			expErr: "",
		},
		{
			name:   "three args: addr,role,req",
			arg:    addr + ",originator,req",
			exp:    pt(addr, types.PartyType_PARTY_TYPE_ORIGINATOR, false),
			expErr: "",
		},
		{
			name:   "three args: bad,role,opt",
			arg:    "badbad,investor,opt",
			exp:    pt("", owner, false),
			expErr: `invalid address "badbad": decoding bech32 failed: invalid bech32 string length 6`,
		},
		{
			name:   "three args: addr,bad,opt",
			arg:    addr + ",badbad,opt",
			exp:    pt(addr, owner, false),
			expErr: `unknown party type: "badbad"`,
		},
		{
			name:   "three args: addr,role,bad",
			arg:    addr + ",affiliate,badbad",
			exp:    pt(addr, types.PartyType_PARTY_TYPE_AFFILIATE, false),
			expErr: `unknown optional value: "badbad", expected "opt" or "req"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.expErr) > 0 {
				tc.expErr = fmt.Sprintf("invalid party %q: %s", tc.arg, tc.expErr)
			}
			actual, err := parseParty(tc.arg)
			AssertErrorValue(t, err, tc.expErr, "parseParty(%q)", tc.arg)
			assert.Equal(t, tc.exp, actual, "parseParty(%q)", tc.arg)
		})
	}
}

func TestParsePartyType(t *testing.T) {
	unspecified := types.PartyType_PARTY_TYPE_UNSPECIFIED
	originator := types.PartyType_PARTY_TYPE_ORIGINATOR
	servicer := types.PartyType_PARTY_TYPE_SERVICER
	investor := types.PartyType_PARTY_TYPE_INVESTOR
	custodian := types.PartyType_PARTY_TYPE_CUSTODIAN
	owner := types.PartyType_PARTY_TYPE_OWNER
	affiliate := types.PartyType_PARTY_TYPE_AFFILIATE
	omnibus := types.PartyType_PARTY_TYPE_OMNIBUS
	provenance := types.PartyType_PARTY_TYPE_PROVENANCE
	controller := types.PartyType_PARTY_TYPE_CONTROLLER
	validator := types.PartyType_PARTY_TYPE_VALIDATOR

	// If this fails, add some unit tests for the new value(s), then update the expected length here.
	assert.Len(t, types.PartyType_name, 11, "types.PartyType_name")

	tests := []struct {
		input  string
		exp    types.PartyType
		expErr string
	}{
		{input: "originator", exp: originator, expErr: ""},
		{input: "ORIGINATOR", exp: originator, expErr: ""},
		{input: "Originator", exp: originator, expErr: ""},
		{input: "servicer", exp: servicer, expErr: ""},
		{input: "SERVICER", exp: servicer, expErr: ""},
		{input: "Servicer", exp: servicer, expErr: ""},
		{input: "investor", exp: investor, expErr: ""},
		{input: "INVESTOR", exp: investor, expErr: ""},
		{input: "Investor", exp: investor, expErr: ""},
		{input: "custodian", exp: custodian, expErr: ""},
		{input: "CUSTODIAN", exp: custodian, expErr: ""},
		{input: "Custodian", exp: custodian, expErr: ""},
		{input: "owner", exp: owner, expErr: ""},
		{input: "OWNER", exp: owner, expErr: ""},
		{input: "Owner", exp: owner, expErr: ""},
		{input: "affiliate", exp: affiliate, expErr: ""},
		{input: "AFFILIATE", exp: affiliate, expErr: ""},
		{input: "Affiliate", exp: affiliate, expErr: ""},
		{input: "omnibus", exp: omnibus, expErr: ""},
		{input: "OMNIBUS", exp: omnibus, expErr: ""},
		{input: "Omnibus", exp: omnibus, expErr: ""},
		{input: "provenance", exp: provenance, expErr: ""},
		{input: "PROVENANCE", exp: provenance, expErr: ""},
		{input: "Provenance", exp: provenance, expErr: ""},
		{input: "controller", exp: controller, expErr: ""},
		{input: "CONTROLLER", exp: controller, expErr: ""},
		{input: "Controller", exp: controller, expErr: ""},
		{input: "validator", exp: validator, expErr: ""},
		{input: "VALIDATOR", exp: validator, expErr: ""},
		{input: "Validator", exp: validator, expErr: ""},

		{input: "party_type_originator", exp: originator, expErr: ""},
		{input: "PARTY_TYPE_ORIGINATOR", exp: originator, expErr: ""},
		{input: "Party_Type_Originator", exp: originator, expErr: ""},
		{input: "party_type_servicer", exp: servicer, expErr: ""},
		{input: "PARTY_TYPE_SERVICER", exp: servicer, expErr: ""},
		{input: "Party_Type_Servicer", exp: servicer, expErr: ""},
		{input: "party_type_investor", exp: investor, expErr: ""},
		{input: "PARTY_TYPE_INVESTOR", exp: investor, expErr: ""},
		{input: "Party_Type_Investor", exp: investor, expErr: ""},
		{input: "party_type_custodian", exp: custodian, expErr: ""},
		{input: "PARTY_TYPE_CUSTODIAN", exp: custodian, expErr: ""},
		{input: "Party_Type_Custodian", exp: custodian, expErr: ""},
		{input: "party_type_owner", exp: owner, expErr: ""},
		{input: "PARTY_TYPE_OWNER", exp: owner, expErr: ""},
		{input: "Party_Type_Owner", exp: owner, expErr: ""},
		{input: "party_type_affiliate", exp: affiliate, expErr: ""},
		{input: "PARTY_TYPE_AFFILIATE", exp: affiliate, expErr: ""},
		{input: "Party_Type_Affiliate", exp: affiliate, expErr: ""},
		{input: "party_type_omnibus", exp: omnibus, expErr: ""},
		{input: "PARTY_TYPE_OMNIBUS", exp: omnibus, expErr: ""},
		{input: "Party_Type_Omnibus", exp: omnibus, expErr: ""},
		{input: "party_type_provenance", exp: provenance, expErr: ""},
		{input: "PARTY_TYPE_PROVENANCE", exp: provenance, expErr: ""},
		{input: "Party_Type_Provenance", exp: provenance, expErr: ""},
		{input: "party_type_controller", exp: controller, expErr: ""},
		{input: "PARTY_TYPE_CONTROLLER", exp: controller, expErr: ""},
		{input: "Party_Type_Controller", exp: controller, expErr: ""},
		{input: "party_type_validator", exp: validator, expErr: ""},
		{input: "PARTY_TYPE_VALIDATOR", exp: validator, expErr: ""},
		{input: "Party_Type_Validator", exp: validator, expErr: ""},

		{input: "unspecified", exp: unspecified, expErr: `unknown party type: "unspecified"`},
		{input: "UNSPECIFIED", exp: unspecified, expErr: `unknown party type: "UNSPECIFIED"`},
		{input: "Unspecified", exp: unspecified, expErr: `unknown party type: "Unspecified"`},

		{input: "party_type_unspecified", exp: unspecified, expErr: `unknown party type: "party_type_unspecified"`},
		{input: "PARTY_TYPE_UNSPECIFIED", exp: unspecified, expErr: `unknown party type: "PARTY_TYPE_UNSPECIFIED"`},
		{input: "Party_Type_Unspecified", exp: unspecified, expErr: `unknown party type: "Party_Type_Unspecified"`},

		{input: "", exp: unspecified, expErr: `unknown party type: ""`},
		{input: " ", exp: unspecified, expErr: `unknown party type: " "`},
		{input: "owner ", exp: unspecified, expErr: `unknown party type: "owner "`},
		{input: " owner", exp: unspecified, expErr: `unknown party type: " owner"`},
		{input: "party_type_owner ", exp: unspecified, expErr: `unknown party type: "party_type_owner "`},
		{input: " party_type_owner", exp: unspecified, expErr: `unknown party type: " party_type_owner"`},
		{input: "owner,controller", exp: unspecified, expErr: `unknown party type: "owner,controller"`},
	}

	for _, tc := range tests {
		t.Run(nameFromInput(tc.input), func(t *testing.T) {
			actual, err := parsePartyType(tc.input)
			AssertErrorValue(t, err, tc.expErr, "parsePartyType(%q) error", tc.input)
			assert.Equal(t, tc.exp.String(), actual.String(), "parsePartyType(%q) value", tc.input)
		})
	}
}

func TestParseOptional(t *testing.T) {
	tests := []struct {
		input  string
		expOpt bool
		expErr string
	}{
		{input: "o", expOpt: true, expErr: ""},
		{input: "O", expOpt: true, expErr: ""},
		{input: "opt", expOpt: true, expErr: ""},
		{input: "OPT", expOpt: true, expErr: ""},
		{input: "Opt", expOpt: true, expErr: ""},
		{input: "optional", expOpt: true, expErr: ""},
		{input: "OPTIONAL", expOpt: true, expErr: ""},
		{input: "Optional", expOpt: true, expErr: ""},

		{input: "r", expOpt: false, expErr: ""},
		{input: "R", expOpt: false, expErr: ""},
		{input: "req", expOpt: false, expErr: ""},
		{input: "REQ", expOpt: false, expErr: ""},
		{input: "Req", expOpt: false, expErr: ""},
		{input: "required", expOpt: false, expErr: ""},
		{input: "REQUIRED", expOpt: false, expErr: ""},
		{input: "Required", expOpt: false, expErr: ""},

		{input: "", expOpt: false, expErr: `unknown optional value: "", expected "opt" or "req"`},
		{input: "  ", expOpt: false, expErr: `unknown optional value: "  ", expected "opt" or "req"`},
		{input: "oo", expOpt: false, expErr: `unknown optional value: "oo", expected "opt" or "req"`},
		{input: "rr", expOpt: false, expErr: `unknown optional value: "rr", expected "opt" or "req"`},
		{input: " o", expOpt: false, expErr: `unknown optional value: " o", expected "opt" or "req"`},
		{input: "o ", expOpt: false, expErr: `unknown optional value: "o ", expected "opt" or "req"`},
	}

	for _, tc := range tests {
		t.Run(nameFromInput(tc.input), func(t *testing.T) {
			opt, err := parseOptional(tc.input)
			AssertErrorValue(t, err, tc.expErr, "parseOptional(%q) error", tc.input)
			assert.Equal(t, tc.expOpt, opt, "parseOptional(%q) bool", tc.input)
		})
	}
}

func TestParseRecordInputs(t *testing.T) {
	hashSource := func(hash string) *types.RecordInput_Hash {
		return &types.RecordInput_Hash{Hash: hash}
	}
	tests := []struct {
		name   string
		arg    string
		exp    []types.RecordInput
		expErr string
	}{
		{name: "empty", arg: "", exp: nil, expErr: ""},
		{name: "space", arg: " ", expErr: `invalid record input " ": expected 4 parts, have 1`},
		{
			name: "one good",
			arg:  "name1,hash1,type1,record",
			exp: []types.RecordInput{
				{Name: "name1", Source: hashSource("hash1"), TypeName: "type1", Status: types.RecordInputStatus_Record},
			},
		},
		{
			name:   "one bad",
			arg:    "name1,hash1,type1,foo",
			expErr: `invalid record input "name1,hash1,type1,foo": unknown record input status: "foo"`,
		},
		{
			name: "three good",
			arg:  "name1,hash1,type1,record;name2,hash2,type2,record;name3,hash3,type3,proposed",
			exp: []types.RecordInput{
				{Name: "name1", Source: hashSource("hash1"), TypeName: "type1", Status: types.RecordInputStatus_Record},
				{Name: "name2", Source: hashSource("hash2"), TypeName: "type2", Status: types.RecordInputStatus_Record},
				{Name: "name3", Source: hashSource("hash3"), TypeName: "type3", Status: types.RecordInputStatus_Proposed},
			},
		},
		{
			name:   "three with bad last",
			arg:    "name1,hash1,type1,record;name2,hash2,type2,record;not,enough,parts",
			expErr: `invalid record input "not,enough,parts": expected 4 parts, have 3`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := parseRecordInputs(tc.arg)
			AssertErrorValue(t, err, tc.expErr, "parseRecordInputs(%q)", tc.arg)
			assert.Equal(t, tc.exp, actual, "parseRecordInputs(%q)", tc.arg)
		})
	}
}

func TestParseRecordInput(t *testing.T) {
	addr := types.RecordMetadataAddress(uuid.New(), "record name")

	tests := []struct {
		name   string
		arg    string
		exp    types.RecordInput
		expErr string
	}{
		{name: "empty", arg: "", expErr: "expected 4 parts, have 1"},
		{name: "1 part", arg: "one", expErr: "expected 4 parts, have 1"},
		{name: "2 parts", arg: "one,two", expErr: "expected 4 parts, have 2"},
		{name: "3 parts", arg: "one,two,three", expErr: "expected 4 parts, have 3"},
		{name: "5 parts", arg: "one,two,three,four,five", expErr: "expected 4 parts, have 5"},
		{
			name: "bad status",
			arg:  "name0,hash0,type0,nope",
			exp: types.RecordInput{
				Name:     "name0",
				Source:   &types.RecordInput_Hash{Hash: "hash0"},
				TypeName: "type0",
				Status:   types.RecordInputStatus_Unknown,
			},
			expErr: `unknown record input status: "nope"`},
		{
			name: "name,hash,type,status",
			arg:  "name1,hash1,type1,proposed",
			exp: types.RecordInput{
				Name:     "name1",
				Source:   &types.RecordInput_Hash{Hash: "hash1"},
				TypeName: "type1",
				Status:   types.RecordInputStatus_Proposed,
			},
		},
		{
			name: "name,record,type,status",
			arg:  "name2," + addr.String() + ",type2,record",
			exp: types.RecordInput{
				Name:     "name2",
				Source:   &types.RecordInput_RecordId{RecordId: addr},
				TypeName: "type2",
				Status:   types.RecordInputStatus_Record,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.expErr) > 0 {
				tc.expErr = fmt.Sprintf(`invalid record input "%s": %s`, tc.arg, tc.expErr)
			}
			actual, err := parseRecordInput(tc.arg)
			AssertErrorValue(t, err, tc.expErr, "parseRecordInput(%q)", tc.arg)
			assert.Equal(t, tc.exp, actual, "parseRecordInput(%q)", tc.arg)
		})
	}
}

func TestParseRecordInputStatus(t *testing.T) {
	unspecified := types.RecordInputStatus_Unknown
	proposed := types.RecordInputStatus_Proposed
	record := types.RecordInputStatus_Record

	// If this fails, add some unit tests for the new value(s), then update the expected length here.
	assert.Len(t, types.RecordInputStatus_name, 3, "types.RecordInputStatus_name")

	tests := []struct {
		input  string
		exp    types.RecordInputStatus
		expErr string
	}{
		{input: "proposed", exp: proposed, expErr: ""},
		{input: "PROPOSED", exp: proposed, expErr: ""},
		{input: "Proposed", exp: proposed, expErr: ""},
		{input: "record", exp: record, expErr: ""},
		{input: "RECORD", exp: record, expErr: ""},
		{input: "Record", exp: record, expErr: ""},

		{input: "record_input_status_proposed", exp: proposed, expErr: ""},
		{input: "RECORD_INPUT_STATUS_PROPOSED", exp: proposed, expErr: ""},
		{input: "Record_Input_Status_Proposed", exp: proposed, expErr: ""},
		{input: "record_input_status_record", exp: record, expErr: ""},
		{input: "RECORD_INPUT_STATUS_RECORD", exp: record, expErr: ""},
		{input: "Record_Input_Status_Record", exp: record, expErr: ""},

		{input: "unspecified", exp: unspecified, expErr: `unknown record input status: "unspecified"`},
		{input: "UNSPECIFIED", exp: unspecified, expErr: `unknown record input status: "UNSPECIFIED"`},
		{input: "Unspecified", exp: unspecified, expErr: `unknown record input status: "Unspecified"`},
		{input: "unknown", exp: unspecified, expErr: `unknown record input status: "unknown"`},
		{input: "UNKNOWN", exp: unspecified, expErr: `unknown record input status: "UNKNOWN"`},
		{input: "Unknown", exp: unspecified, expErr: `unknown record input status: "Unknown"`},

		{input: "record_input_status_unspecified", exp: unspecified, expErr: `unknown record input status: "record_input_status_unspecified"`},
		{input: "RECORD_INPUT_STATUS_UNSPECIFIED", exp: unspecified, expErr: `unknown record input status: "RECORD_INPUT_STATUS_UNSPECIFIED"`},
		{input: "Record_Input_Status_Unspecified", exp: unspecified, expErr: `unknown record input status: "Record_Input_Status_Unspecified"`},
		{input: "record_input_status_unknown", exp: unspecified, expErr: `unknown record input status: "record_input_status_unknown"`},
		{input: "RECORD_INPUT_STATUS_UNKNOWN", exp: unspecified, expErr: `unknown record input status: "RECORD_INPUT_STATUS_UNKNOWN"`},
		{input: "Record_Input_Status_Unknown", exp: unspecified, expErr: `unknown record input status: "Record_Input_Status_Unknown"`},

		{input: "", exp: unspecified, expErr: `unknown record input status: ""`},
		{input: " ", exp: unspecified, expErr: `unknown record input status: " "`},
		{input: " record", exp: unspecified, expErr: `unknown record input status: " record"`},
		{input: "record ", exp: unspecified, expErr: `unknown record input status: "record "`},
		{input: " record_input_status_record", exp: unspecified, expErr: `unknown record input status: " record_input_status_record"`},
		{input: "record_input_status_record ", exp: unspecified, expErr: `unknown record input status: "record_input_status_record "`},
		{input: "record,record", exp: unspecified, expErr: `unknown record input status: "record,record"`},
	}

	for _, tc := range tests {
		t.Run(nameFromInput(tc.input), func(t *testing.T) {
			actual, err := parseRecordInputStatus(tc.input)
			AssertErrorValue(t, err, tc.expErr, "parseRecordInputStatus(%q)", tc.input)
			assert.Equal(t, tc.exp, actual, "parseRecordInputStatus(%q)", tc.input)
		})
	}
}

func TestParseRecordOutputs(t *testing.T) {
	tests := []struct {
		name   string
		arg    string
		exp    []types.RecordOutput
		expErr string
	}{
		{
			name:   "empty",
			arg:    "",
			exp:    nil,
			expErr: "",
		},
		{
			name: "one good",
			arg:  "hash1,skip",
			exp: []types.RecordOutput{
				{Hash: "hash1", Status: types.ResultStatus_RESULT_STATUS_SKIP},
			},
			expErr: "",
		},
		{
			name:   "one bad",
			arg:    "bad",
			exp:    nil,
			expErr: `invalid record output "bad": expected 2 parts, have 1`,
		},
		{
			name: "three good",
			arg:  "hash1,pass;hash2,skip;hash3,fail",
			exp: []types.RecordOutput{
				{Hash: "hash1", Status: types.ResultStatus_RESULT_STATUS_PASS},
				{Hash: "hash2", Status: types.ResultStatus_RESULT_STATUS_SKIP},
				{Hash: "hash3", Status: types.ResultStatus_RESULT_STATUS_FAIL},
			},
			expErr: "",
		},
		{
			name:   "three with bad last",
			arg:    "hash1,pass;hash2,skip;hash3",
			exp:    nil,
			expErr: `invalid record output "hash3": expected 2 parts, have 1`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := parseRecordOutputs(tc.arg)
			AssertErrorValue(t, err, tc.expErr, "parseRecordOutputs(%q)", tc.arg)
			assert.Equal(t, tc.exp, actual, "parseRecordOutputs(%q)", tc.arg)
		})
	}
}

func TestParseRecordOutput(t *testing.T) {
	tests := []struct {
		name   string
		arg    string
		exp    types.RecordOutput
		expErr string
	}{
		{
			name:   "empty",
			arg:    "",
			exp:    types.RecordOutput{},
			expErr: "expected 2 parts, have 1",
		},
		{
			name:   "hash",
			arg:    "hash1",
			exp:    types.RecordOutput{},
			expErr: "expected 2 parts, have 1",
		},
		{
			name:   "hash,status",
			arg:    "hash1,fail",
			exp:    types.RecordOutput{Hash: "hash1", Status: types.ResultStatus_RESULT_STATUS_FAIL},
			expErr: "",
		},
		{
			name:   "empty,status",
			arg:    ",pass",
			exp:    types.RecordOutput{Hash: "", Status: types.ResultStatus_RESULT_STATUS_PASS},
			expErr: "",
		},
		{
			name:   "hash,empty",
			arg:    "hash2,",
			exp:    types.RecordOutput{Hash: "hash2"},
			expErr: `unknown result status: ""`,
		},
		{
			name:   "hash,bad",
			arg:    "hash3,what",
			exp:    types.RecordOutput{Hash: "hash3"},
			expErr: `unknown result status: "what"`,
		},
		{
			name:   "three parts",
			arg:    "part1,part2,part3",
			exp:    types.RecordOutput{},
			expErr: "expected 2 parts, have 3",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.expErr) > 0 {
				tc.expErr = fmt.Sprintf(`invalid record output "%s": %s`, tc.arg, tc.expErr)
			}
			actual, err := parseRecordOutput(tc.arg)
			AssertErrorValue(t, err, tc.expErr, "parseRecordOutput(%q)", tc.arg)
			assert.Equal(t, tc.exp, actual, "parseRecordOutput(%q)", tc.arg)
		})
	}
}

func TestParseResultStatus(t *testing.T) {
	unspecified := types.ResultStatus_RESULT_STATUS_UNSPECIFIED
	pass := types.ResultStatus_RESULT_STATUS_PASS
	skip := types.ResultStatus_RESULT_STATUS_SKIP
	fail := types.ResultStatus_RESULT_STATUS_FAIL

	// If this fails, add some unit tests for the new value(s), then update the expected length here.
	assert.Len(t, types.ResultStatus_name, 4, "types.ResultStatus_name")

	tests := []struct {
		input  string
		exp    types.ResultStatus
		expErr string
	}{
		{input: "pass", exp: pass, expErr: ""},
		{input: "PASS", exp: pass, expErr: ""},
		{input: "Pass", exp: pass, expErr: ""},
		{input: "skip", exp: skip, expErr: ""},
		{input: "SKIP", exp: skip, expErr: ""},
		{input: "Skip", exp: skip, expErr: ""},
		{input: "fail", exp: fail, expErr: ""},
		{input: "FAIL", exp: fail, expErr: ""},
		{input: "Fail", exp: fail, expErr: ""},

		{input: "result_status_pass", exp: pass, expErr: ""},
		{input: "RESULT_STATUS_PASS", exp: pass, expErr: ""},
		{input: "Result_Status_Pass", exp: pass, expErr: ""},
		{input: "result_status_skip", exp: skip, expErr: ""},
		{input: "RESULT_STATUS_SKIP", exp: skip, expErr: ""},
		{input: "Result_Status_Skip", exp: skip, expErr: ""},
		{input: "result_status_fail", exp: fail, expErr: ""},
		{input: "RESULT_STATUS_FAIL", exp: fail, expErr: ""},
		{input: "Result_Status_Fail", exp: fail, expErr: ""},

		{input: "", exp: unspecified, expErr: `unknown result status: ""`},
		{input: " pass", exp: unspecified, expErr: `unknown result status: " pass"`},
		{input: "pass ", exp: unspecified, expErr: `unknown result status: "pass "`},
		{input: " result_status_pass", exp: unspecified, expErr: `unknown result status: " result_status_pass"`},
		{input: "result_status_pass ", exp: unspecified, expErr: `unknown result status: "result_status_pass "`},
		{input: "pass,pass", exp: unspecified, expErr: `unknown result status: "pass,pass"`},
	}

	for _, tc := range tests {
		t.Run(nameFromInput(tc.input), func(t *testing.T) {
			actual, err := parseResultStatus(tc.input)
			AssertErrorValue(t, err, tc.expErr, "parseResultStatus(%q)", tc.input)
			assert.Equal(t, tc.exp.String(), actual.String(), "parseResultStatus(%q)", tc.input)
		})
	}
}

func TestParseInputSpecifications(t *testing.T) {
	addr := types.RecordMetadataAddress(uuid.New(), "record name")

	tests := []struct {
		name   string
		arg    string
		exp    []*types.InputSpecification
		expErr string
	}{
		{
			name:   "empty string",
			arg:    "",
			exp:    nil,
			expErr: "",
		},
		{
			name:   "one bad entry",
			arg:    "bad",
			exp:    nil,
			expErr: `invalid input specification "bad": expected 3 parts, have 1`,
		},
		{
			name: "one good entry",
			arg:  "name1,type1,hash1",
			exp: []*types.InputSpecification{
				{Name: "name1", TypeName: "type1", Source: &types.InputSpecification_Hash{Hash: "hash1"}},
			},
			expErr: "",
		},
		{
			name: "three good entries",
			arg:  "name1,type1,hash1;name2,type2," + addr.String() + ";name3,type3,hash3",
			exp: []*types.InputSpecification{
				{Name: "name1", TypeName: "type1", Source: &types.InputSpecification_Hash{Hash: "hash1"}},
				{Name: "name2", TypeName: "type2", Source: &types.InputSpecification_RecordId{RecordId: addr}},
				{Name: "name3", TypeName: "type3", Source: &types.InputSpecification_Hash{Hash: "hash3"}},
			},
			expErr: "",
		},
		{
			name:   "three entries last bad",
			arg:    "name1,type1,hash1;name2,type2," + addr.String() + ";name3,type3",
			exp:    nil,
			expErr: `invalid input specification "name3,type3": expected 3 parts, have 2`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := parseInputSpecifications(tc.arg)
			AssertErrorValue(t, err, tc.expErr, "parseInputSpecifications(%q)", tc.arg)
			assert.Equal(t, tc.exp, actual, "parseInputSpecifications(%q)", tc.arg)
		})
	}
}

func TestParseInputSpecification(t *testing.T) {
	addr := types.RecordMetadataAddress(uuid.New(), "record name")

	tests := []struct {
		name   string
		arg    string
		exp    *types.InputSpecification
		expErr string
	}{
		{
			name:   "empty string",
			arg:    "",
			exp:    nil,
			expErr: `invalid input specification "": expected 3 parts, have 1`,
		},
		{
			name:   "one part",
			arg:    "a",
			exp:    nil,
			expErr: `invalid input specification "a": expected 3 parts, have 1`,
		},
		{
			name:   "two parts",
			arg:    "a,b",
			exp:    nil,
			expErr: `invalid input specification "a,b": expected 3 parts, have 2`,
		},
		{
			name:   "four parts",
			arg:    "a,b,c,d",
			exp:    nil,
			expErr: `invalid input specification "a,b,c,d": expected 3 parts, have 4`,
		},
		{
			name: "name,type,hash",
			arg:  "name,type,hash",
			exp: &types.InputSpecification{
				Name:     "name",
				TypeName: "type",
				Source:   &types.InputSpecification_Hash{Hash: "hash"},
			},
			expErr: "",
		},
		{
			name: "name,type,record",
			arg:  "NAME,TYPE," + addr.String(),
			exp: &types.InputSpecification{
				Name:     "NAME",
				TypeName: "TYPE",
				Source:   &types.InputSpecification_RecordId{RecordId: addr},
			},
			expErr: "",
		},
		{
			name: "empty,type,hash",
			arg:  ",type,hash",
			exp: &types.InputSpecification{
				Name:     "",
				TypeName: "type",
				Source:   &types.InputSpecification_Hash{Hash: "hash"},
			},
			expErr: "",
		},
		{
			name: "empty,type,record",
			arg:  ",type," + addr.String(),
			exp: &types.InputSpecification{
				Name:     "",
				TypeName: "type",
				Source:   &types.InputSpecification_RecordId{RecordId: addr},
			},
			expErr: "",
		},
		{
			name: "name,empty,hash",
			arg:  "name,,hash",
			exp: &types.InputSpecification{
				Name:     "name",
				TypeName: "",
				Source:   &types.InputSpecification_Hash{Hash: "hash"},
			},
			expErr: "",
		},
		{
			name: "name,empty,record",
			arg:  "name,," + addr.String(),
			exp: &types.InputSpecification{
				Name:     "name",
				TypeName: "",
				Source:   &types.InputSpecification_RecordId{RecordId: addr},
			},
			expErr: "",
		},
		{
			name: "name,type,empty",
			arg:  "name,type,",
			exp: &types.InputSpecification{
				Name:     "name",
				TypeName: "type",
				Source:   &types.InputSpecification_Hash{Hash: ""},
			},
			expErr: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := parseInputSpecification(tc.arg)
			AssertErrorValue(t, err, tc.expErr, "parseInputSpecification(%q)", tc.arg)
			assert.Equal(t, tc.exp, actual, "parseInputSpecification(%q)", tc.arg)
		})
	}
}

func TestParsePartyTypes(t *testing.T) {
	originator := types.PartyType_PARTY_TYPE_ORIGINATOR
	servicer := types.PartyType_PARTY_TYPE_SERVICER
	investor := types.PartyType_PARTY_TYPE_INVESTOR
	custodian := types.PartyType_PARTY_TYPE_CUSTODIAN
	owner := types.PartyType_PARTY_TYPE_OWNER
	affiliate := types.PartyType_PARTY_TYPE_AFFILIATE
	omnibus := types.PartyType_PARTY_TYPE_OMNIBUS
	provenance := types.PartyType_PARTY_TYPE_PROVENANCE
	controller := types.PartyType_PARTY_TYPE_CONTROLLER
	validator := types.PartyType_PARTY_TYPE_VALIDATOR

	// If this fails, update the "one of each" test, then update the expected length here.
	assert.Len(t, types.PartyType_name, 11, "types.PartyType_name")

	tests := []struct {
		name   string
		arg    string
		exp    []types.PartyType
		expErr string
	}{
		{name: "empty string", arg: "", exp: nil, expErr: ""},
		{name: "a space", arg: " ", exp: nil, expErr: `unknown party type: " "`},
		{name: "single bad entry", arg: "bad", exp: nil, expErr: `unknown party type: "bad"`},
		{
			name:   "single good entry",
			arg:    "custodian",
			exp:    []types.PartyType{custodian},
			expErr: "",
		},
		{
			name:   "two good entries",
			arg:    "validator,owner",
			exp:    []types.PartyType{validator, owner},
			expErr: "",
		},
		{name: "two entries first bad", arg: "bad,owner", exp: nil, expErr: `unknown party type: "bad"`},
		{name: "two entries second bad", arg: "omnibus,bad", exp: nil, expErr: `unknown party type: "bad"`},
		{
			name: "one of each",
			arg:  "controller,investor,provenance,custodian,owner,affiliate,validator,servicer,originator,omnibus",
			exp: []types.PartyType{
				controller,
				investor,
				provenance,
				custodian,
				owner,
				affiliate,
				validator,
				servicer,
				originator,
				omnibus,
			},
			expErr: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := parsePartyTypes(tc.arg)
			AssertErrorValue(t, err, tc.expErr, "parsePartyTypes(%q)", tc.arg)
			assert.Equal(t, tc.exp, actual, "parsePartyTypes(%q)", tc.arg)
		})
	}
}

func TestParseDescription(t *testing.T) {
	tests := []struct {
		name string
		args []string
		exp  *types.Description
	}{
		{name: "nil args", args: nil, exp: nil},
		{name: "empty args", args: []string{}, exp: nil},
		{
			name: "one arg",
			args: []string{"arg 1"},
			exp:  &types.Description{Name: "arg 1", Description: "", WebsiteUrl: "", IconUrl: ""},
		},
		{
			name: "two args",
			args: []string{"arg 1", "arg 2"},
			exp:  &types.Description{Name: "arg 1", Description: "arg 2", WebsiteUrl: "", IconUrl: ""},
		},
		{
			name: "three args",
			args: []string{"arg 1", "arg 2", "arg 3"},
			exp:  &types.Description{Name: "arg 1", Description: "arg 2", WebsiteUrl: "arg 3", IconUrl: ""},
		},
		{
			name: "three args middle one empty",
			args: []string{"arg 1", "", "arg 3"},
			exp:  &types.Description{Name: "arg 1", Description: "", WebsiteUrl: "arg 3", IconUrl: ""},
		},
		{
			name: "four args",
			args: []string{"arg 1", "arg 2", "arg 3", "arg 4"},
			exp:  &types.Description{Name: "arg 1", Description: "arg 2", WebsiteUrl: "arg 3", IconUrl: "arg 4"},
		},
		{
			name: "five args",
			args: []string{"arg 1", "arg 2", "arg 3", "arg 4", "arg 5"},
			exp:  &types.Description{Name: "arg 1", Description: "arg 2", WebsiteUrl: "arg 3", IconUrl: "arg 4"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var origArgs []string
			if tc.args != nil {
				origArgs = make([]string, 0, len(tc.args))
			}
			origArgs = append(origArgs, tc.args...)
			actual := parseDescription(tc.args)
			assert.Equal(t, tc.exp, actual, "parseDescription")
			assert.Equal(t, origArgs, tc.args, "args provided to parseDescription after calling it")
		})
	}
}
