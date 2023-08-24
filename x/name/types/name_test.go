package types

import (
	"fmt"
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type NameRecordTestSuite struct {
	addr sdk.AccAddress
	suite.Suite
}

func TestNameRecordSuite(t *testing.T) {
	s := new(NameRecordTestSuite)
	s.addr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	suite.Run(t, s)
}

func (s *NameRecordTestSuite) TestNameRecordString() {
	nr := NewNameRecord("example", s.addr, true)
	s.Require().Equal(fmt.Sprintf("example: %s [restricted]", s.addr.String()), nr.String())
	nr = NewNameRecord("example", s.addr, false)
	s.Require().Equal(fmt.Sprintf("example: %s", s.addr.String()), nr.String())
}

func (s *NameRecordTestSuite) TestNameRecordValidateBasic() {
	cases := map[string]struct {
		name      NameRecord
		expectErr bool
		errValue  string
	}{
		"valid name": {
			NewNameRecord("example", s.addr, true),
			false,
			"",
		},
		"should fail to validate basic empty name": {
			NewNameRecord("", s.addr, true),
			true,
			"segment of name is too short",
		},
		"should fail to validate basic empty addr": {
			NewNameRecord("example", sdk.AccAddress{}, true),
			true,
			"invalid account address",
		},
	}
	for n, tc := range cases {
		tc := tc

		s.Run(n, func() {
			err := tc.name.Validate()
			if tc.expectErr {
				s.Error(err)
				if s != nil {
					s.Equal(tc.errValue, err.Error())
				}
			} else {
				s.NoError(err)
			}

		})
	}
}

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		exp   string
	}{
		{name: "empty string", input: "", exp: ""},
		{name: "two spaces", input: "  ", exp: ""},
		{name: "with middle space", input: " a b ", exp: "a b"},
		{name: "upper case", input: "ABcDE", exp: "abcde"},
		{name: "space around first segment", input: " ghi .def.abc", exp: "ghi.def.abc"},
		{name: "space around middle segment", input: "ghi. def .abc", exp: "ghi.def.abc"},
		{name: "space around third segment", input: "ghi.def. abc ", exp: "ghi.def.abc"},
		{name: "middle segment has upper case", input: "ghi.DeF.abc", exp: "ghi.def.abc"},
		{name: "empty first segment", input: ".def.abc", exp: ".def.abc"},
		{name: "first segment is a space", input: " .def.abc", exp: ".def.abc"},
		{name: "empty last segment", input: "ghi.def.", exp: "ghi.def."},
		{name: "last segment is a space", input: "ghi.def. ", exp: "ghi.def."},
		{name: "empty middle segment", input: "ghi..abc", exp: "ghi..abc"},
		{name: "middle segment is a space", input: "ghi. .abc", exp: "ghi..abc"},
		{name: "middle segment is a bell", input: "ghi.\a.abc", exp: "ghi.\a.abc"},
		{name: "middle segment is three dashes", input: "ghi.---.abc", exp: "ghi.---.abc"},
		{
			name:  "middle segment is upper-case uuid",
			input: "ghi.CFE48CAA-223E-44E1-8AD7-9181A30D4D91.abc",
			exp:   "ghi.cfe48caa-223e-44e1-8ad7-9181a30d4d91.abc",
		},
		{
			name:  "middle segment is lower-case uuid with extra spaces around it",
			input: "ghi.   cfe48caa-223e-44e1-8ad7-9181a30d4d91 .abc",
			exp:   "ghi.cfe48caa-223e-44e1-8ad7-9181a30d4d91.abc",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			normalized := NormalizeName(tc.input)
			assert.Equal(t, tc.exp, normalized, "NormalizeName(%q)", tc.input)
		})
	}
}

func TestIsValidName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		exp   bool
	}{
		{name: "empty string", input: "", exp: true},
		{name: "two spaces", input: "  ", exp: true},
		{name: "with middle space", input: " a b ", exp: false},
		{name: "upper case", input: "ABcDE", exp: true},
		{name: "space around first segment", input: " ghi .def.abc", exp: true},
		{name: "space around middle segment", input: "ghi. def .abc", exp: true},
		{name: "space around third segment", input: "ghi.def. abc ", exp: true},
		{name: "middle segment has upper case", input: "ghi.DeF.abc", exp: true},
		{name: "empty first segment", input: ".def.abc", exp: true},
		{name: "first segment is a space", input: " .def.abc", exp: true},
		{name: "empty last segment", input: "ghi.def.", exp: true},
		{name: "last segment is a space", input: "ghi.def. ", exp: true},
		{name: "empty middle segment", input: "ghi..abc", exp: true},
		{name: "middle segment is a space", input: "ghi. .abc", exp: true},
		{name: "one segment two dashes", input: "a-b-c", exp: false},
		{name: "two segments one dash each", input: "a-1.b-2", exp: true},
		{name: "comma in first segment", input: "a,1.b-2", exp: false},
		{name: "comma in second segment", input: "a-1.b,2", exp: false},
		{name: "space in middle of first segment", input: "a 1.b-2", exp: false},
		{name: "space in middle of second segment", input: "a-1.b 2", exp: false},
		{name: "newline in middle of second segment", input: "a-1.b\n2", exp: false},
		{name: "middle segment has middlespace", input: "a. b c .d", exp: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ok := IsValidName(tc.input)
			assert.Equal(t, tc.exp, ok, "IsValidName(%q)", tc.input)
		})
	}
}

func TestIsValidNameSegment(t *testing.T) {
	tests := []struct {
		name string
		seg  string
		exp  bool
	}{
		{name: "empty", seg: "", exp: true},
		{name: "uuid with dashes", seg: "01234567-8909-8765-4321-012345678901", exp: true},
		{name: "uuid without dashes", seg: "01234567890987654321012345678901", exp: true},
		{name: "one dash", seg: "-", exp: true},
		{name: "two dashes", seg: "--", exp: false},
		{name: "all english lower-case letters, a dash, and all arabic digits", seg: "abcdefghijklmnopqrstuvwxyz-0123456789", exp: true},
		{name: "with a newline", seg: "ab\nde", exp: false},
		{name: "with a space", seg: "ab de", exp: false},
		{name: "with an underscore", seg: "ab_de", exp: false},
		{name: "single quoted", seg: "'abcde'", exp: false},
		{name: "double quoted", seg: `"abcde"`, exp: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ok := IsValidNameSegment(tc.seg)
			assert.Equal(t, tc.exp, ok, "IsValidNameSegment(%q)", tc.seg)
		})
	}
}

func TestIsValidNameSegmentChars(t *testing.T) {
	type testerFunc func(t *testing.T, r rune, tableName string, lo uint32, hi uint32, stride uint32) bool
	assertRuneIsOkay := func(t *testing.T, r rune, tableName string, lo uint32, hi uint32, stride uint32) bool {
		isGraphic := unicode.IsGraphic(r)
		if !assert.True(t, isGraphic, "IsGraphic(%q = %u) %s{%u, %u, %d}", r, r, tableName, lo, hi, stride) {
			return false
		}
		isValid := IsValidNameSegment(string(r))
		return assert.True(t, isValid, "IsValidNameSegment(%q = %u) %s{%u, %u, %d}", r, r, tableName, lo, hi, stride)
	}
	assertRuneIsInvalid := func(t *testing.T, r rune, tableName string, lo uint32, hi uint32, stride uint32) bool {
		isValid := IsValidNameSegment(string(r))
		return assert.False(t, isValid, "IsValidNameSegment(%q = %u) %s{%u, %u, %d}", r, r, tableName, lo, hi, stride)
	}
	containsRune := func(r rune, rz []rune) bool {
		for _, z := range rz {
			if r == z {
				return true
			}
		}
		return false
	}

	tests := []struct {
		name      string
		table     *unicode.RangeTable
		tableName string
		tester    testerFunc
		skips     []rune
	}{
		{
			name:      "all lower-case letters are okay",
			table:     unicode.Lower,
			tableName: "Lower",
			tester:    assertRuneIsOkay,
		},
		{
			name:      "all digits are okay",
			table:     unicode.Digit,
			tableName: "Digit",
			tester:    assertRuneIsOkay,
		},
		{
			name:      "no upper-case letters are allowed",
			table:     unicode.Upper,
			tableName: "Upper",
			tester:    assertRuneIsInvalid,
		},
		{
			name:      "most punctuation is not okay",
			table:     unicode.Punct,
			tableName: "Punct",
			tester:    assertRuneIsInvalid,
			skips:     []rune{'-'},
		},
		{
			name:      "no control/other characters are allowed",
			table:     unicode.Other,
			tableName: "Other",
			tester:    assertRuneIsInvalid,
		},
		{
			name:      "no space characters are allowed",
			table:     unicode.Space,
			tableName: "Space",
			tester:    assertRuneIsInvalid,
		},
		{
			name:      "no marks are allowed",
			table:     unicode.Mark,
			tableName: "Mark",
			tester:    assertRuneIsInvalid,
		},
		{
			name:      "no symbols are allowed",
			table:     unicode.Symbol,
			tableName: "Symbol",
			tester:    assertRuneIsInvalid,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, row := range tc.table.R16 {
				rv := row.Lo
				for rv <= row.Hi {
					r := rune(rv)
					if !containsRune(r, tc.skips) {
						tc.tester(t, r, tc.tableName, uint32(row.Lo), uint32(row.Hi), uint32(row.Stride))
					}
					if rv+row.Stride <= rv {
						break
					}
					rv += row.Stride
				}
			}
			for _, row := range tc.table.R32 {
				rv := row.Lo
				for rv <= row.Hi {
					r := rune(rv)
					if !containsRune(r, tc.skips) {
						tc.tester(t, r, tc.tableName, row.Lo, row.Hi, row.Stride)
					}
					if rv+row.Stride <= rv {
						break
					}
					rv += row.Stride
				}
			}
		})
	}
}

func TestIsValidUUID(t *testing.T) {
	tests := []struct {
		name string
		str  string
		exp  bool
	}{
		{name: "upper case", str: "4FEBAF0F-1BA7-473E-B62A-B1F3C44067C0", exp: true},
		{name: "lower case ", str: "4febaf0f-1ba7-473e-b62a-b1f3c44067c0", exp: true},
		{name: "no dashes", str: "4FEBAF0F1BA7473EB62AB1F3C44067C0", exp: true},
		{name: "empty", str: "", exp: false},
		{name: "whitespace", str: strings.Repeat(" ", 32), exp: false},
		{name: "one short ", str: "4febaf0f-1ba7-473e-b62a-b1f3c44067c", exp: false},
		{name: "bad char ", str: "4febaf0f-1ba7-473e-b62a-b1f3c44067cg", exp: false},
		{name: "missing a couple dashes", str: "4febaf0f-1ba7473eb62a-b1f3c44067c0", exp: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ok := IsValidUUID(tc.str)
			assert.Equal(t, tc.exp, ok, "IsValidUUID(%q)", tc.str)
		})
	}
}
