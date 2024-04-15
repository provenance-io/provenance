package quarantine_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	"github.com/provenance-io/provenance/x/quarantine"
	"github.com/provenance-io/provenance/x/quarantine/testutil"
)

func TestPrefixValues(t *testing.T) {
	prefixes := []struct {
		name     string
		prefix   []byte
		expected []byte
	}{
		{name: "OptInPrefix", prefix: quarantine.OptInPrefix, expected: []byte{0x00}},
		{name: "AutoResponsePrefix", prefix: quarantine.AutoResponsePrefix, expected: []byte{0x01}},
		{name: "RecordPrefix", prefix: quarantine.RecordPrefix, expected: []byte{0x02}},
		{name: "RecordIndexPrefix", prefix: quarantine.RecordIndexPrefix, expected: []byte{0x03}},
	}

	for _, p := range prefixes {
		t.Run(fmt.Sprintf("%s expected value", p.name), func(t *testing.T) {
			assert.Equal(t, p.prefix, p.expected, p.name)
		})
	}

	for i := 0; i < len(prefixes)-1; i++ {
		for j := i + 1; j < len(prefixes); j++ {
			t.Run(fmt.Sprintf("%s is different from %s", prefixes[i].name, prefixes[j].name), func(t *testing.T) {
				assert.NotEqual(t, prefixes[i].prefix, prefixes[j].prefix, "expected: %s, actual: %s", prefixes[i].name, prefixes[j].name)
			})
		}
	}
}

func TestMakeKey(t *testing.T) {
	tests := []struct {
		name  string
		part1 []byte
		part2 []byte
		exp   []byte
	}{
		{
			name:  "nil + nil",
			part1: nil,
			part2: nil,
			exp:   []byte{},
		},
		{
			name:  "nil + empty",
			part1: nil,
			part2: []byte{},
			exp:   []byte{},
		},
		{
			name:  "empty + nil",
			part1: []byte{},
			part2: nil,
			exp:   []byte{},
		},
		{
			name:  "empty + empty",
			part1: []byte{},
			part2: []byte{},
			exp:   []byte{},
		},
		{
			name:  "nil + one",
			part1: nil,
			part2: []byte{0x70},
			exp:   []byte{0x70},
		},
		{
			name:  "empty + one",
			part1: []byte{},
			part2: []byte{0x70},
			exp:   []byte{0x70},
		},
		{
			name:  "one + one",
			part1: []byte{0x69},
			part2: []byte{0x70},
			exp:   []byte{0x69, 0x70},
		},

		{
			name:  "nil + five",
			part1: nil,
			part2: []byte{0x70, 0x70, 0x70, 0x70, 0x70},
			exp:   []byte{0x70, 0x70, 0x70, 0x70, 0x70},
		},
		{
			name:  "empty + five",
			part1: []byte{},
			part2: []byte{0x70, 0x70, 0x70, 0x70, 0x70},
			exp:   []byte{0x70, 0x70, 0x70, 0x70, 0x70},
		},
		{
			name:  "one + five",
			part1: []byte{0x69},
			part2: []byte{0x70, 0x70, 0x70, 0x70, 0x70},
			exp:   []byte{0x69, 0x70, 0x70, 0x70, 0x70, 0x70},
		},
		{
			name:  "six + five",
			part1: []byte{0x68, 0x68, 0x68, 0x68, 0x68, 0x68},
			part2: []byte{0x70, 0x70, 0x70, 0x70, 0x70},
			exp:   []byte{0x68, 0x68, 0x68, 0x68, 0x68, 0x68, 0x70, 0x70, 0x70, 0x70, 0x70},
		},
		{
			name:  "six + one",
			part1: []byte{0x68, 0x68, 0x68, 0x68, 0x68, 0x68},
			part2: []byte{0x70},
			exp:   []byte{0x68, 0x68, 0x68, 0x68, 0x68, 0x68, 0x70},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			origPart1 := testutil.MakeCopyOfByteSlice(tc.part1)
			origPart2 := testutil.MakeCopyOfByteSlice(tc.part2)
			actual := quarantine.MakeKey(tc.part1, tc.part2)
			actualCopy := testutil.MakeCopyOfByteSlice(actual)
			assert.Equal(t, tc.exp, actual, "MakeKey result")
			assert.Equal(t, origPart1, tc.part1, "part1 before and after MakeKey")
			assert.Equal(t, origPart2, tc.part2, "part2 before and after MakeKey")
			if len(tc.part1) > 0 {
				// Make sure the result doesn't change if part1 is later changed.
				for i := range tc.part1 {
					tc.part1[i]++
				}
				assert.Equal(t, actualCopy, actual, "MakeKey result after changing each byte of part1 slice")
				for i := range tc.part1 {
					tc.part1[i]--
				}
			}
			if len(tc.part2) > 0 {
				// Make sure the result doesn't change if part2 is later changed.
				for i := range tc.part2 {
					tc.part2[i]++
				}
				assert.Equal(t, actualCopy, actual, "MakeKey result after changing each byte of part2 slice")
				for i := range tc.part2 {
					tc.part2[i]--
				}
			}
			if len(actual) > 0 {
				// Make sure the parts don't change if the result is later changed.
				for i := range actual {
					actual[i]++
				}
				assert.Equal(t, origPart1, tc.part1, "part1 after changing each byte of result slice")
				assert.Equal(t, origPart2, tc.part2, "part2 after changing each byte of result slice")
				for i := range actual {
					actual[i]--
				}
			}
		})
	}
}

func TestCreateOptInKey(t *testing.T) {
	expectedPrefix := quarantine.OptInPrefix
	testAddr0 := testutil.MakeTestAddr("coik", 0)
	testAddr1 := testutil.MakeTestAddr("coik", 1)
	badAddr := testutil.MakeBadAddr("coik", 2)

	t.Run("starts with OptInPrefix", func(t *testing.T) {
		key := quarantine.CreateOptInKey(testAddr0)
		actual := key[:len(expectedPrefix)]
		assert.Equal(t, expectedPrefix, actual, "key prefix")
	})

	makeExpected := func(addrBz []byte) []byte {
		rv := make([]byte, 0, len(expectedPrefix)+1+len(addrBz))
		rv = append(rv, expectedPrefix...)
		rv = append(rv, byte(len(addrBz)))
		rv = append(rv, addrBz...)
		return rv
	}
	tests := []struct {
		name     string
		toAddr   sdk.AccAddress
		expected []byte
		expPanic string
	}{
		{
			name:     "addr 0",
			toAddr:   testAddr0,
			expected: makeExpected(testAddr0),
		},
		{
			name:     "addr 0",
			toAddr:   testAddr1,
			expected: makeExpected(testAddr1),
		},
		{
			name:     "nil",
			toAddr:   nil,
			expected: expectedPrefix,
		},
		{
			name:     "too long",
			toAddr:   badAddr,
			expected: nil,
			expPanic: fmt.Sprintf("address length should be max %d bytes, got %d: unknown address", address.MaxAddrLen, len(badAddr)),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []byte
			testFunc := func() {
				actual = quarantine.CreateOptInKey(tc.toAddr)
			}
			if len(tc.expPanic) == 0 {
				if assert.NotPanics(t, testFunc, "CreateOptInKey") {
					assert.Equal(t, tc.expected, actual, "CreateOptInKey result")
				}
			} else {
				assert.PanicsWithError(t, tc.expPanic, testFunc, "CreateOptInKey")
			}
		})
	}
}

func TestParseOptInKey(t *testing.T) {
	testAddr0 := testutil.MakeTestAddr("poik", 0)
	testAddr1 := testutil.MakeTestAddr("poik", 1)
	testAddr2 := testutil.MakeTestAddr("poik", 2)
	longAddr := testutil.MakeLongAddr("poik", 3)

	makeKey := func(pre []byte, addrLen int, addrBz []byte) []byte {
		rv := make([]byte, 0, len(pre)+1+len(addrBz))
		rv = append(rv, pre...)
		rv = append(rv, byte(addrLen))
		rv = append(rv, addrBz...)
		return rv
	}
	tests := []struct {
		name     string
		key      []byte
		expected sdk.AccAddress
		expPanic string
	}{
		{
			name:     "addr 0",
			key:      makeKey(quarantine.OptInPrefix, len(testAddr0), testAddr0),
			expected: testAddr0,
		},
		{
			name:     "addr 1",
			key:      makeKey(quarantine.OptInPrefix, len(testAddr1), testAddr1),
			expected: testAddr1,
		},
		{
			name:     "addr 2",
			key:      makeKey(quarantine.OptInPrefix, len(testAddr2), testAddr2),
			expected: testAddr2,
		},
		{
			name:     "longer addr",
			key:      makeKey(quarantine.OptInPrefix, len(longAddr), longAddr),
			expected: longAddr,
		},
		{
			name:     "too short",
			key:      makeKey(quarantine.OptInPrefix, len(testAddr0)+1, testAddr0),
			expPanic: fmt.Sprintf("expected key of length at least %d, got %d", len(testAddr0)+1+2, len(testAddr0)+2),
		},
		{
			name:     "from CreateOptInKey addr 0",
			key:      quarantine.CreateOptInKey(testAddr0),
			expected: testAddr0,
		},
		{
			name:     "from CreateOptInKey addr 1",
			key:      quarantine.CreateOptInKey(testAddr1),
			expected: testAddr1,
		},
		{
			name:     "from CreateOptInKey addr 2",
			key:      quarantine.CreateOptInKey(testAddr2),
			expected: testAddr2,
		},
		{
			name:     "from CreateOptInKey longAddr",
			key:      quarantine.CreateOptInKey(longAddr),
			expected: longAddr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sdk.AccAddress
			testFunc := func() {
				actual = quarantine.ParseOptInKey(tc.key)
			}
			if len(tc.expPanic) == 0 {
				if assert.NotPanics(t, testFunc, "ParseOptInKey") {
					assert.Equal(t, tc.expected, actual, "ParseOptInKey result")
				}
			} else {
				assert.PanicsWithValue(t, tc.expPanic, testFunc, "ParseOptInKey")
			}
		})
	}
}

func TestCreateAutoResponseToAddrPrefix(t *testing.T) {
	expectedPrefix := quarantine.AutoResponsePrefix
	testAddr0 := testutil.MakeTestAddr("cartap", 0)
	testAddr1 := testutil.MakeTestAddr("cartap", 1)
	badAddr := testutil.MakeBadAddr("cartap", 2)

	t.Run("starts with AutoResponsePrefix", func(t *testing.T) {
		key := quarantine.CreateAutoResponseToAddrPrefix(testAddr0)
		actual := key[:len(expectedPrefix)]
		assert.Equal(t, expectedPrefix, actual, "key prefix")
	})

	makeExpected := func(addrBz []byte) []byte {
		rv := make([]byte, 0, len(expectedPrefix)+1+len(addrBz))
		rv = append(rv, expectedPrefix...)
		rv = append(rv, byte(len(addrBz)))
		rv = append(rv, addrBz...)
		return rv
	}

	tests := []struct {
		name     string
		toAddr   sdk.AccAddress
		expected []byte
		expPanic string
	}{
		{
			name:     "addr 0",
			toAddr:   testAddr0,
			expected: makeExpected(testAddr0),
		},
		{
			name:     "addr 1",
			toAddr:   testAddr1,
			expected: makeExpected(testAddr1),
		},
		{
			name:     "nil",
			toAddr:   nil,
			expected: expectedPrefix,
		},
		{
			name:     "too long",
			toAddr:   badAddr,
			expected: nil,
			expPanic: fmt.Sprintf("address length should be max %d bytes, got %d: unknown address", address.MaxAddrLen, len(badAddr)),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []byte
			testFunc := func() {
				actual = quarantine.CreateAutoResponseToAddrPrefix(tc.toAddr)
			}
			if len(tc.expPanic) == 0 {
				if assert.NotPanics(t, testFunc, "CreateAutoResponseToAddrPrefix") {
					assert.Equal(t, tc.expected, actual, "CreateAutoResponseToAddrPrefix result")
				}
			} else {
				assert.PanicsWithError(t, tc.expPanic, testFunc, "CreateAutoResponseToAddrPrefix")
			}
		})
	}
}

func TestCreateAutoResponseKey(t *testing.T) {
	expectedPrefix := quarantine.AutoResponsePrefix
	testAddr0 := testutil.MakeTestAddr("cark", 0)
	testAddr1 := testutil.MakeTestAddr("cark", 1)
	badAddr := testutil.MakeBadAddr("cark", 2)
	longAddr := testutil.MakeLongAddr("cark", 3)

	t.Run("starts with AutoResponsePrefix", func(t *testing.T) {
		key := quarantine.CreateAutoResponseKey(testAddr0, testAddr1)
		actual := key[:len(expectedPrefix)]
		assert.Equal(t, expectedPrefix, actual, "key prefix")
	})

	makeExpected := func(toAddrBz, fromAddrBz []byte) []byte {
		rv := make([]byte, 0, len(expectedPrefix)+1+len(toAddrBz)+1+len(fromAddrBz))
		rv = append(rv, expectedPrefix...)
		rv = append(rv, byte(len(toAddrBz)))
		rv = append(rv, toAddrBz...)
		rv = append(rv, byte(len(fromAddrBz)))
		rv = append(rv, fromAddrBz...)
		return rv
	}

	tests := []struct {
		name     string
		toAddr   sdk.AccAddress
		fromAddr sdk.AccAddress
		expected []byte
		expPanic string
	}{
		{
			name:     "addr 0 addr 1",
			toAddr:   testAddr0,
			fromAddr: testAddr1,
			expected: makeExpected(testAddr0, testAddr1),
		},
		{
			name:     "addr 1 long addr",
			toAddr:   testAddr1,
			fromAddr: longAddr,
			expected: makeExpected(testAddr1, longAddr),
		},
		{
			name:     "long addr addr 0",
			toAddr:   longAddr,
			fromAddr: testAddr0,
			expected: makeExpected(longAddr, testAddr0),
		},
		{
			name:     "long addr long addr",
			toAddr:   longAddr,
			fromAddr: longAddr,
			expected: makeExpected(longAddr, longAddr),
		},
		{
			name:     "bad toAddr",
			toAddr:   badAddr,
			fromAddr: testAddr0,
			expPanic: fmt.Sprintf("address length should be max %d bytes, got %d: unknown address", address.MaxAddrLen, len(badAddr)),
		},
		{
			name:     "bad fromAddr",
			toAddr:   testAddr0,
			fromAddr: badAddr,
			expPanic: fmt.Sprintf("address length should be max %d bytes, got %d: unknown address", address.MaxAddrLen, len(badAddr)),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []byte
			testFunc := func() {
				actual = quarantine.CreateAutoResponseKey(tc.toAddr, tc.fromAddr)
			}
			if len(tc.expPanic) == 0 {
				if assert.NotPanics(t, testFunc, "CreateAutoResponseKey") {
					assert.Equal(t, tc.expected, actual, "CreateAutoResponseKey result")
				}
			} else {
				assert.PanicsWithError(t, tc.expPanic, testFunc, "CreateAutoResponseKey")
			}
		})
	}
}

func TestParseAutoResponseKey(t *testing.T) {
	testAddr0 := testutil.MakeTestAddr("park", 0)
	testAddr1 := testutil.MakeTestAddr("park", 1)
	longAddr := testutil.MakeLongAddr("park", 2)

	makeKey := func(pre []byte, toAddrLen int, toAddrBz []byte, fromAddrLen int, fromAddrBz []byte) []byte {
		rv := make([]byte, 0, len(pre)+1+len(toAddrBz)+1+len(fromAddrBz))
		rv = append(rv, pre...)
		rv = append(rv, byte(toAddrLen))
		rv = append(rv, toAddrBz...)
		rv = append(rv, byte(fromAddrLen))
		rv = append(rv, fromAddrBz...)
		return rv
	}

	tests := []struct {
		name        string
		key         []byte
		expToAddr   sdk.AccAddress
		expFromAddr sdk.AccAddress
		expPanic    string
	}{
		{
			name:        "addr 0 addr 1",
			key:         quarantine.CreateAutoResponseKey(testAddr0, testAddr1),
			expToAddr:   testAddr0,
			expFromAddr: testAddr1,
		},
		{
			name:        "addr 1 addr 0",
			key:         quarantine.CreateAutoResponseKey(testAddr1, testAddr0),
			expToAddr:   testAddr1,
			expFromAddr: testAddr0,
		},
		{
			name:        "long addr addr 1",
			key:         quarantine.CreateAutoResponseKey(longAddr, testAddr1),
			expToAddr:   longAddr,
			expFromAddr: testAddr1,
		},
		{
			name:        "addr 0 long addr",
			key:         quarantine.CreateAutoResponseKey(testAddr0, longAddr),
			expToAddr:   testAddr0,
			expFromAddr: longAddr,
		},
		{
			name:     "bad toAddr len",
			key:      makeKey(quarantine.AutoResponsePrefix, 200, testAddr0, 20, testAddr1),
			expPanic: fmt.Sprintf("expected key of length at least %d, got %d", 202, 43),
		},
		{
			name:     "bad fromAddr len",
			key:      makeKey(quarantine.AutoResponsePrefix, len(testAddr1), testAddr1, len(testAddr0)+1, testAddr0),
			expPanic: fmt.Sprintf("expected key of length at least %d, got %d", 44, 43),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actualToAddr, actualFromAddr sdk.AccAddress
			testFunc := func() {
				actualToAddr, actualFromAddr = quarantine.ParseAutoResponseKey(tc.key)
			}
			if len(tc.expPanic) == 0 {
				if assert.NotPanics(t, testFunc, "ParseAutoResponseKey") {
					assert.Equal(t, tc.expToAddr, actualToAddr, "ParseAutoResponseKey toAddr")
					assert.Equal(t, tc.expFromAddr, actualFromAddr, "ParseAutoResponseKey fromAddr")
				}
			} else {
				assert.PanicsWithValue(t, tc.expPanic, testFunc, "ParseAutoResponseKey")
			}
		})
	}
}

func TestCreateRecordToAddrPrefix(t *testing.T) {
	expectedPrefix := quarantine.RecordPrefix
	testAddr0 := testutil.MakeTestAddr("crtap", 0)
	testAddr1 := testutil.MakeTestAddr("crtap", 1)
	badAddr := testutil.MakeBadAddr("crtap", 2)

	t.Run("starts with RecordPrefix", func(t *testing.T) {
		key := quarantine.CreateRecordToAddrPrefix(testAddr0)
		actual := key[:len(expectedPrefix)]
		assert.Equal(t, expectedPrefix, actual, "key prefix")
	})

	makeExpected := func(addrBz []byte) []byte {
		rv := make([]byte, 0, len(expectedPrefix)+1+len(addrBz))
		rv = append(rv, expectedPrefix...)
		rv = append(rv, byte(len(addrBz)))
		rv = append(rv, addrBz...)
		return rv
	}

	tests := []struct {
		name     string
		toAddr   sdk.AccAddress
		expected []byte
		expPanic string
	}{
		{
			name:     "addr 0",
			toAddr:   testAddr0,
			expected: makeExpected(testAddr0),
		},
		{
			name:     "addr 1",
			toAddr:   testAddr1,
			expected: makeExpected(testAddr1),
		},
		{
			name:     "nil",
			toAddr:   nil,
			expected: expectedPrefix,
		},
		{
			name:     "too long",
			toAddr:   badAddr,
			expected: nil,
			expPanic: fmt.Sprintf("address length should be max %d bytes, got %d: unknown address", address.MaxAddrLen, len(badAddr)),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []byte
			testFunc := func() {
				actual = quarantine.CreateRecordToAddrPrefix(tc.toAddr)
			}
			if len(tc.expPanic) == 0 {
				if assert.NotPanics(t, testFunc, "CreateRecordToAddrPrefix") {
					assert.Equal(t, tc.expected, actual, "CreateRecordToAddrPrefix result")
				}
			} else {
				assert.PanicsWithError(t, tc.expPanic, testFunc, "CreateRecordToAddrPrefix")
			}
		})
	}
}

func TestCreateRecordKey(t *testing.T) {
	expectedPrefix := quarantine.RecordPrefix
	testAddr0 := testutil.MakeTestAddr("crk", 0)
	testAddr1 := testutil.MakeTestAddr("crk", 1)
	testAddr2 := testutil.MakeTestAddr("crk", 2)
	testAddr3 := testutil.MakeTestAddr("crk", 3)
	badAddr := testutil.MakeBadAddr("crk", 4)
	longAddr := testutil.MakeLongAddr("crk", 5)

	t.Run("starts with RecordPrefix", func(t *testing.T) {
		key := quarantine.CreateRecordKey(testAddr0, testAddr1)
		actual := key[:len(expectedPrefix)]
		assert.Equal(t, expectedPrefix, actual, "key prefix")
	})

	makeExpected := func(toAddrBz []byte, fromAddrs ...sdk.AccAddress) []byte {
		recordId := quarantine.CreateRecordSuffix(fromAddrs)
		rv := make([]byte, 0, len(expectedPrefix)+1+len(toAddrBz)+1+len(recordId))
		rv = append(rv, expectedPrefix...)
		rv = append(rv, byte(len(toAddrBz)))
		rv = append(rv, toAddrBz...)
		rv = append(rv, byte(len(recordId)))
		rv = append(rv, recordId...)
		return rv
	}

	tests := []struct {
		name      string
		toAddr    sdk.AccAddress
		fromAddrs []sdk.AccAddress
		expected  []byte
		expPanic  string
	}{
		{
			name:      "addr 0 addr 1",
			toAddr:    testAddr0,
			fromAddrs: []sdk.AccAddress{testAddr1},
			expected:  makeExpected(testAddr0, testAddr1),
		},
		{
			name:      "addr 1 long addr",
			toAddr:    testAddr1,
			fromAddrs: []sdk.AccAddress{longAddr},
			expected:  makeExpected(testAddr1, longAddr),
		},
		{
			name:      "long addr addr 0",
			toAddr:    longAddr,
			fromAddrs: []sdk.AccAddress{testAddr0},
			expected:  makeExpected(longAddr, testAddr0),
		},
		{
			name:      "long addr long addr",
			toAddr:    longAddr,
			fromAddrs: []sdk.AccAddress{longAddr},
			expected:  makeExpected(longAddr, longAddr),
		},
		{
			name:      "to addr 3 from addrs 0 1 2 and long",
			toAddr:    testAddr3,
			fromAddrs: []sdk.AccAddress{testAddr0, testAddr1, testAddr2, longAddr},
			expected:  makeExpected(testAddr3, testAddr0, testAddr1, testAddr2, longAddr),
		},
		{
			name:      "to addr 2 from addrs 1 0 diff order",
			toAddr:    testAddr2,
			fromAddrs: []sdk.AccAddress{testAddr1, testAddr0},
			expected:  makeExpected(testAddr2, testAddr0, testAddr1),
		},
		{
			name:      "bad toAddr panics",
			toAddr:    badAddr,
			fromAddrs: []sdk.AccAddress{testAddr0},
			expPanic:  fmt.Sprintf("address length should be max %d bytes, got %d: unknown address", address.MaxAddrLen, len(badAddr)),
		},
		{
			name:      "bad fromAddr ok",
			toAddr:    testAddr0,
			fromAddrs: []sdk.AccAddress{badAddr},
			expected:  makeExpected(testAddr0, badAddr),
		},
		{
			name:      "no fromAddrs panics",
			toAddr:    testAddr2,
			fromAddrs: []sdk.AccAddress{},
			expPanic:  "at least one fromAddr is required: internal logic error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []byte
			testFunc := func() {
				actual = quarantine.CreateRecordKey(tc.toAddr, tc.fromAddrs...)
			}
			if len(tc.expPanic) == 0 {
				if assert.NotPanics(t, testFunc, "CreateRecordKey") {
					assert.Equal(t, tc.expected, actual, "CreateRecordKey result")
				}
			} else {
				assert.PanicsWithError(t, tc.expPanic, testFunc, "CreateRecordKey")
			}
		})
	}
}

func TestCreateRecordSuffix(t *testing.T) {
	testAddr0 := testutil.MakeTestAddr("crs", 0)
	testAddr1 := testutil.MakeTestAddr("crs", 1)
	testAddr2 := testutil.MakeTestAddr("crs", 2)
	testAddrs := []sdk.AccAddress{testAddr0, testAddr1, testAddr2}
	badAddr := testutil.MakeBadAddr("crs", 3)

	t.Run("panics if no addrs", func(t *testing.T) {
		assert.PanicsWithError(t, "at least one fromAddr is required: internal logic error",
			func() { quarantine.CreateRecordSuffix([]sdk.AccAddress{}) },
			"createRecordSuffix([]sdk.AccAddress{})",
		)
	})

	t.Run("panics with nil addrs", func(t *testing.T) {
		assert.PanicsWithError(t, "at least one fromAddr is required: internal logic error",
			func() { quarantine.CreateRecordSuffix(nil) },
			"createRecordSuffix(nil)",
		)
	})

	createRecordSuffixAndAssertInputUnchanged := func(t *testing.T, input []sdk.AccAddress, msg string, args ...interface{}) []byte {
		t.Helper()
		msgAndArgs := []interface{}{msg + " input before and after"}
		msgAndArgs = append(msgAndArgs, args...)
		var orig []sdk.AccAddress
		if input != nil {
			orig = make([]sdk.AccAddress, len(input))
			for i, addr := range input {
				orig[i] = make(sdk.AccAddress, len(addr))
				copy(orig[i], addr)
			}
		}
		actual := quarantine.CreateRecordSuffix(input)
		assert.Equal(t, orig, input, msgAndArgs...)
		return actual
	}

	t.Run("single addrs are unchanged", func(t *testing.T) {
		for i, addr := range testAddrs {
			expected := make([]byte, len(addr))
			copy(expected, addr)

			actual := createRecordSuffixAndAssertInputUnchanged(t, []sdk.AccAddress{addr}, "addr %d", i)
			assert.Equal(t, expected, actual, "addr %d", i)
		}
	})

	t.Run("long addr is truncated", func(t *testing.T) {
		expected := make([]byte, 32)
		copy(expected, badAddr[:32])

		actual := createRecordSuffixAndAssertInputUnchanged(t, []sdk.AccAddress{badAddr}, "bad addr")
		assert.Equal(t, expected, actual, "bad addr as suffix")
	})

	t.Run("two addrs order does not matter", func(t *testing.T) {
		input1 := []sdk.AccAddress{testAddr0, testAddr1}
		input2 := []sdk.AccAddress{testAddr1, testAddr0}
		expected := createRecordSuffixAndAssertInputUnchanged(t, input1, "addrs 0 then 1")
		actual := createRecordSuffixAndAssertInputUnchanged(t, input2, "addrs 1 then 0")
		assert.Equal(t, expected, actual, "addrs 0 then 1, vs 1 then 0")
	})

	t.Run("three addrs order does not matter", func(t *testing.T) {
		inputTestAddrsIndexes := [][]int{
			{0, 1, 2},
			{0, 2, 1},
			{1, 0, 2},
			{1, 2, 0},
			{2, 0, 1},
			{2, 1, 0},
		}
		inputs := make([][]sdk.AccAddress, len(inputTestAddrsIndexes))
		outputs := make([][]byte, len(inputTestAddrsIndexes))
		for i, taIndexes := range inputTestAddrsIndexes {
			inputs[i] = make([]sdk.AccAddress, len(taIndexes))
			for j, ind := range taIndexes {
				inputs[i][j] = testAddrs[ind]
			}
			outputs[i] = createRecordSuffixAndAssertInputUnchanged(t, inputs[i], "addrs %v", taIndexes)
		}
		for i := 0; i < len(outputs)-1; i++ {
			for j := i + 1; j < len(outputs); j++ {
				assert.Equal(t, outputs[i], outputs[j], "test addrs %v vs %v", inputTestAddrsIndexes[i], inputTestAddrsIndexes[j])
			}
		}
	})

	t.Run("two addrs different alone vs together", func(t *testing.T) {
		input1 := []sdk.AccAddress{testAddr1}
		input2 := []sdk.AccAddress{testAddr2}
		inputBoth := []sdk.AccAddress{testAddr1, testAddr2}
		actual1 := createRecordSuffixAndAssertInputUnchanged(t, input1, "addr 1")
		actual2 := createRecordSuffixAndAssertInputUnchanged(t, input2, "addr 2")
		actualBoth := createRecordSuffixAndAssertInputUnchanged(t, inputBoth, "both")

		assert.NotEqual(t, actual1, actual2, "addr 1 vs addr 2")
		assert.NotEqual(t, actual1, actualBoth, "addr 1 vs both")
		assert.NotEqual(t, actual2, actualBoth, "addr 2 vs both")
		assert.NotContains(t, actualBoth, actual1, "both vs addr 1")
		assert.NotContains(t, actualBoth, actual2, "both vs addr 2")
	})
}

func TestParseRecordKey(t *testing.T) {
	testAddr0 := testutil.MakeTestAddr("prk", 0)
	testAddr1 := testutil.MakeTestAddr("prk", 1)
	testAddr2 := testutil.MakeTestAddr("prk", 2)
	longAddr := testutil.MakeLongAddr("prk", 3)

	makeKey := func(pre []byte, toAddrLen int, toAddrBz []byte, fromAddrLen int, fromAddrBz []byte) []byte {
		rv := make([]byte, 0, len(pre)+1+len(toAddrBz)+1+len(fromAddrBz))
		rv = append(rv, pre...)
		rv = append(rv, byte(toAddrLen))
		rv = append(rv, toAddrBz...)
		rv = append(rv, byte(fromAddrLen))
		rv = append(rv, fromAddrBz...)
		return rv
	}

	tests := []struct {
		name        string
		key         []byte
		expToAddr   sdk.AccAddress
		expFromAddr sdk.AccAddress
		expPanic    string
	}{
		{
			name:        "addr 0 addr 1",
			key:         quarantine.CreateRecordKey(testAddr0, testAddr1),
			expToAddr:   testAddr0,
			expFromAddr: testAddr1,
		},
		{
			name:        "addr 1 addr 0",
			key:         quarantine.CreateRecordKey(testAddr1, testAddr0),
			expToAddr:   testAddr1,
			expFromAddr: testAddr0,
		},
		{
			name:        "long addr addr 1",
			key:         quarantine.CreateRecordKey(longAddr, testAddr1),
			expToAddr:   longAddr,
			expFromAddr: testAddr1,
		},
		{
			name:        "addr 0 long addr",
			key:         quarantine.CreateRecordKey(testAddr0, longAddr),
			expToAddr:   testAddr0,
			expFromAddr: longAddr,
		},
		{
			name:        "multiple from addrs",
			key:         quarantine.CreateRecordKey(testAddr0, testAddr1, testAddr2),
			expToAddr:   testAddr0,
			expFromAddr: quarantine.CreateRecordSuffix([]sdk.AccAddress{testAddr1, testAddr2}),
		},
		{
			name:        "multiple from addrs diff order",
			key:         quarantine.CreateRecordKey(testAddr0, testAddr2, testAddr1),
			expToAddr:   testAddr0,
			expFromAddr: quarantine.CreateRecordSuffix([]sdk.AccAddress{testAddr1, testAddr2}),
		},
		{
			name:     "bad toAddr len",
			key:      makeKey(quarantine.RecordPrefix, 200, testAddr0, 20, testAddr1),
			expPanic: fmt.Sprintf("expected key of length at least %d, got %d", 202, 43),
		},
		{
			name:     "bad fromAddr len",
			key:      makeKey(quarantine.RecordPrefix, len(testAddr1), testAddr1, len(testAddr0)+1, testAddr0),
			expPanic: fmt.Sprintf("expected key of length at least %d, got %d", 44, 43),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actualToAddr, actualFromAddr sdk.AccAddress
			testFunc := func() {
				actualToAddr, actualFromAddr = quarantine.ParseRecordKey(tc.key)
			}
			if len(tc.expPanic) == 0 {
				if assert.NotPanics(t, testFunc, "ParseRecordKey") {
					assert.Equal(t, tc.expToAddr, actualToAddr, "ParseRecordKey toAddr")
					assert.Equal(t, tc.expFromAddr, actualFromAddr, "ParseRecordKey fromAddr")
				}
			} else {
				assert.PanicsWithValue(t, tc.expPanic, testFunc, "ParseRecordKey")
			}
		})
	}
}

func TestCreateRecordIndexToAddrPrefix(t *testing.T) {
	expectedPrefix := quarantine.RecordIndexPrefix
	testAddr0 := testutil.MakeTestAddr("critap", 0)
	testAddr1 := testutil.MakeTestAddr("critap", 1)
	badAddr := testutil.MakeBadAddr("critap", 2)

	t.Run("starts with RecordIndexPrefix", func(t *testing.T) {
		key := quarantine.CreateRecordIndexToAddrPrefix(testAddr0)
		actual := key[:len(expectedPrefix)]
		assert.Equal(t, expectedPrefix, actual, "key prefix")
	})

	makeExpected := func(addrBz []byte) []byte {
		rv := make([]byte, 0, len(expectedPrefix)+1+len(addrBz))
		rv = append(rv, expectedPrefix...)
		rv = append(rv, byte(len(addrBz)))
		rv = append(rv, addrBz...)
		return rv
	}

	tests := []struct {
		name     string
		toAddr   sdk.AccAddress
		expected []byte
		expPanic string
	}{
		{
			name:     "addr 0",
			toAddr:   testAddr0,
			expected: makeExpected(testAddr0),
		},
		{
			name:     "addr 1",
			toAddr:   testAddr1,
			expected: makeExpected(testAddr1),
		},
		{
			name:     "nil",
			toAddr:   nil,
			expected: expectedPrefix,
		},
		{
			name:     "too long",
			toAddr:   badAddr,
			expected: nil,
			expPanic: fmt.Sprintf("address length should be max %d bytes, got %d: unknown address", address.MaxAddrLen, len(badAddr)),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []byte
			testFunc := func() {
				actual = quarantine.CreateRecordIndexToAddrPrefix(tc.toAddr)
			}
			if len(tc.expPanic) == 0 {
				if assert.NotPanics(t, testFunc, "CreateRecordIndexToAddrPrefix") {
					assert.Equal(t, tc.expected, actual, "CreateRecordIndexToAddrPrefix result")
				}
			} else {
				assert.PanicsWithError(t, tc.expPanic, testFunc, "CreateRecordIndexToAddrPrefix")
			}
		})
	}
}

func TestCreateRecordIndexKey(t *testing.T) {
	expectedPrefix := quarantine.RecordIndexPrefix
	testAddr0 := testutil.MakeTestAddr("crik", 0)
	testAddr1 := testutil.MakeTestAddr("crik", 1)
	badAddr := testutil.MakeBadAddr("crik", 2)
	longAddr := testutil.MakeLongAddr("crik", 3)

	t.Run("starts with RecordIndexPrefix", func(t *testing.T) {
		key := quarantine.CreateRecordIndexKey(testAddr0, testAddr1)
		actual := key[:len(expectedPrefix)]
		assert.Equal(t, expectedPrefix, actual, "key prefix")
	})

	makeExpected := func(toAddrBz, fromAddrBz []byte) []byte {
		rv := make([]byte, 0, len(expectedPrefix)+1+len(toAddrBz)+1+len(fromAddrBz))
		rv = append(rv, expectedPrefix...)
		rv = append(rv, byte(len(toAddrBz)))
		rv = append(rv, toAddrBz...)
		rv = append(rv, byte(len(fromAddrBz)))
		rv = append(rv, fromAddrBz...)
		return rv
	}

	tests := []struct {
		name     string
		toAddr   sdk.AccAddress
		fromAddr sdk.AccAddress
		expected []byte
		expPanic string
	}{
		{
			name:     "addr 0 addr 1",
			toAddr:   testAddr0,
			fromAddr: testAddr1,
			expected: makeExpected(testAddr0, testAddr1),
		},
		{
			name:     "addr 1 long addr",
			toAddr:   testAddr1,
			fromAddr: longAddr,
			expected: makeExpected(testAddr1, longAddr),
		},
		{
			name:     "long addr addr 0",
			toAddr:   longAddr,
			fromAddr: testAddr0,
			expected: makeExpected(longAddr, testAddr0),
		},
		{
			name:     "long addr long addr",
			toAddr:   longAddr,
			fromAddr: longAddr,
			expected: makeExpected(longAddr, longAddr),
		},
		{
			name:     "bad toAddr",
			toAddr:   badAddr,
			fromAddr: testAddr0,
			expPanic: fmt.Sprintf("address length should be max %d bytes, got %d: unknown address", address.MaxAddrLen, len(badAddr)),
		},
		{
			name:     "bad fromAddr",
			toAddr:   testAddr0,
			fromAddr: badAddr,
			expPanic: fmt.Sprintf("address length should be max %d bytes, got %d: unknown address", address.MaxAddrLen, len(badAddr)),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []byte
			testFunc := func() {
				actual = quarantine.CreateRecordIndexKey(tc.toAddr, tc.fromAddr)
			}
			if len(tc.expPanic) == 0 {
				if assert.NotPanics(t, testFunc, "CreateRecordIndexKey") {
					assert.Equal(t, tc.expected, actual, "CreateRecordIndexKey result")
				}
			} else {
				assert.PanicsWithError(t, tc.expPanic, testFunc, "CreateRecordIndexKey")
			}
		})
	}
}

func TestParseRecordIndexKey(t *testing.T) {
	testAddr0 := testutil.MakeTestAddr("prik", 0)
	testAddr1 := testutil.MakeTestAddr("prik", 1)
	longAddr := testutil.MakeLongAddr("prik", 2)

	makeKey := func(pre []byte, toAddrLen int, toAddrBz []byte, fromAddrLen int, fromAddrBz []byte) []byte {
		rv := make([]byte, 0, len(pre)+1+len(toAddrBz)+1+len(fromAddrBz))
		rv = append(rv, pre...)
		rv = append(rv, byte(toAddrLen))
		rv = append(rv, toAddrBz...)
		rv = append(rv, byte(fromAddrLen))
		rv = append(rv, fromAddrBz...)
		return rv
	}

	tests := []struct {
		name        string
		key         []byte
		expToAddr   sdk.AccAddress
		expFromAddr sdk.AccAddress
		expPanic    string
	}{
		{
			name:        "addr 0 addr 1",
			key:         quarantine.CreateRecordIndexKey(testAddr0, testAddr1),
			expToAddr:   testAddr0,
			expFromAddr: testAddr1,
		},
		{
			name:        "addr 1 addr 0",
			key:         quarantine.CreateRecordIndexKey(testAddr1, testAddr0),
			expToAddr:   testAddr1,
			expFromAddr: testAddr0,
		},
		{
			name:        "long addr addr 1",
			key:         quarantine.CreateRecordIndexKey(longAddr, testAddr1),
			expToAddr:   longAddr,
			expFromAddr: testAddr1,
		},
		{
			name:        "addr 0 long addr",
			key:         quarantine.CreateRecordIndexKey(testAddr0, longAddr),
			expToAddr:   testAddr0,
			expFromAddr: longAddr,
		},
		{
			name:     "bad toAddr len",
			key:      makeKey(quarantine.RecordIndexPrefix, 200, testAddr0, 20, testAddr1),
			expPanic: fmt.Sprintf("expected key of length at least %d, got %d", 202, 43),
		},
		{
			name:     "bad fromAddr len",
			key:      makeKey(quarantine.RecordIndexPrefix, len(testAddr1), testAddr1, len(testAddr0)+1, testAddr0),
			expPanic: fmt.Sprintf("expected key of length at least %d, got %d", 44, 43),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actualToAddr, actualFromAddr sdk.AccAddress
			testFunc := func() {
				actualToAddr, actualFromAddr = quarantine.ParseRecordIndexKey(tc.key)
			}
			if len(tc.expPanic) == 0 {
				if assert.NotPanics(t, testFunc, "ParseRecordIndexKey") {
					assert.Equal(t, tc.expToAddr, actualToAddr, "ParseRecordIndexKey toAddr")
					assert.Equal(t, tc.expFromAddr, actualFromAddr, "ParseRecordIndexKey fromAddr")
				}
			} else {
				assert.PanicsWithValue(t, tc.expPanic, testFunc, "ParseRecordIndexKey")
			}
		})
	}
}
