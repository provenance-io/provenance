package provutils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/gogoproto/proto"
	"github.com/provenance-io/provenance/testutil/assertions"

	. "github.com/provenance-io/provenance/internal/provutils"
)

// TestEnum is an enum created for these unit tests.
// It's just like one that would be auto-generated from a proto file.
//
//	enum TestEnum {
//	  TEST_ENUM_UNSPECIFIED = 0;
//	  TEST_ENUM_ONE = 1;
//	  TEST_ENUM_TWO = 2;
//	  TEST_ENUM_THREE = 3;
//	  TEST_ENUM_FOUR = 4;
//	  TEST_ENUM_FIVE = 5;
//	}
type TestEnum int32

const (
	TEST_ENUM_UNSPECIFIED TestEnum = 0
	TEST_ENUM_ONE         TestEnum = 1
	TEST_ENUM_TWO         TestEnum = 2
	TEST_ENUM_THREE       TestEnum = 3
	TEST_ENUM_FOUR        TestEnum = 4
	TEST_ENUM_FIVE        TestEnum = 5
)

var TestEnum_name = map[int32]string{
	0: "TEST_ENUM_UNSPECIFIED",
	1: "TEST_ENUM_ONE",
	2: "TEST_ENUM_TWO",
	3: "TEST_ENUM_THREE",
	4: "TEST_ENUM_FOUR",
	5: "TEST_ENUM_FIVE",
}

var TestEnum_value = map[string]int32{
	"TEST_ENUM_UNSPECIFIED": 0,
	"TEST_ENUM_ONE":         1,
	"TEST_ENUM_TWO":         2,
	"TEST_ENUM_THREE":       3,
	"TEST_ENUM_FOUR":        4,
	"TEST_ENUM_FIVE":        5,
}

func (x TestEnum) String() string {
	return proto.EnumName(TestEnum_name, int32(x))
}

func TestEnumUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name   string
		data   string
		exp    int32
		expErr string
	}{
		{
			name: "string: short: unspecified",
			data: `"UNSPECIFIED"`,
			exp:  int32(TEST_ENUM_UNSPECIFIED),
		},
		{
			name: "string: short: specified: uppercase",
			data: `"THREE"`,
			exp:  int32(TEST_ENUM_THREE),
		},
		{
			name: "string: short: specified: lowercase",
			data: `"one"`,
			exp:  int32(TEST_ENUM_ONE),
		},
		{
			name: "string: short: specified: mixed case",
			data: `"fOUr"`,
			exp:  int32(TEST_ENUM_FOUR),
		},
		{
			name: "string full: unspecified",
			data: `"TEST_ENUM_UNSPECIFIED"`,
			exp:  int32(TEST_ENUM_UNSPECIFIED),
		},
		{
			name: "string full: specified: uppercase",
			data: `"TEST_ENUM_FIVE"`,
			exp:  int32(TEST_ENUM_FIVE),
		},
		{
			name: "string full: specified: lowercase",
			data: `"test_enum_two"`,
			exp:  int32(TEST_ENUM_TWO),
		},
		{
			name: "string full: specified: mixed case",
			data: `"tesT_enuM_threE"`,
			exp:  int32(TEST_ENUM_THREE),
		},
		{
			name:   "string: does not exist",
			data:   `"SIX"`,
			expErr: "unknown test_enum string value: \"SIX\"",
		},
		{
			name: "number: unspecified",
			data: "0",
			exp:  int32(TEST_ENUM_UNSPECIFIED),
		},
		{
			name: "number: specified",
			data: "2",
			exp:  int32(TEST_ENUM_TWO),
		},
		{
			name:   "number: does not exist: negative",
			data:   "-1",
			expErr: "unknown test_enum integer value: -1",
		},
		{
			name:   "number: does not exist: positive",
			data:   "6",
			expErr: "unknown test_enum integer value: 6",
		},
		{
			name:   "invalid data",
			data:   "nope",
			exp:    0,
			expErr: "test_enum must be a string or integer, got: \"nope\"",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act int32
			var err error
			testFunc := func() {
				act, err = EnumUnmarshalJSON([]byte(tc.data), TestEnum_value, TestEnum_name)
			}
			require.NotPanics(t, testFunc, "EnumUnmarshalJSON(%q)", tc.data)
			assertions.AssertErrorValue(t, err, tc.expErr, "EnumUnmarshalJSON(%q) error", tc.data)
			assert.Equal(t, int(tc.exp), int(act), "EnumUnmarshalJSON(%q) result", tc.data)
		})
	}
}

func TestEnumValidateExists(t *testing.T) {
	tests := []struct {
		name   string
		value  TestEnum
		expErr string
	}{
		{name: "unspecified", value: TEST_ENUM_UNSPECIFIED},
		{name: "one", value: TEST_ENUM_ONE},
		{name: "two", value: TEST_ENUM_TWO},
		{name: "three", value: TEST_ENUM_THREE},
		{name: "four", value: TEST_ENUM_FOUR},
		{name: "five", value: TEST_ENUM_FIVE},
		{name: "negative one", value: -1, expErr: "unknown test_enum enum value: -1"},
		{name: "six", value: 6, expErr: "unknown test_enum enum value: 6"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = EnumValidateExists(tc.value, TestEnum_name)
			}
			require.NotPanics(t, testFunc, "EnumValidateExists(%s)", tc.value)
			assertions.AssertErrorValue(t, err, tc.expErr, "EnumValidateExists(%s) error", tc.value)
		})
	}
}
