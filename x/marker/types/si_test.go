package types

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SITestSuite struct {
	suite.Suite
}

func (s *SITestSuite) SetupTest() {}

func TestSITestSuite(t *testing.T) {
	suite.Run(t, new(SITestSuite))
}

func (s *SITestSuite) TestInit() {
	s.T().Run("SIPrefixSymbol spot check", func(t *testing.T) {
		assert.Equal(t, "", SIPrefixSymbol[SI_PREFIX_NONE], "SI_PREFIX_NONE")
		assert.Equal(t, "Z", SIPrefixSymbol[SI_PREFIX_ZETTA], "SI_PREFIX_ZETTA")
		assert.Equal(t, "µ", SIPrefixSymbol[SI_PREFIX_MICRO], "SI_PREFIX_MICRO")
		assert.Equal(t, "f", SIPrefixSymbol[SI_PREFIX_FEMTO], "SI_PREFIX_FEMTO")
		_, okBad := SIPrefixSymbol[SIPrefix(100)]
		assert.False(t, okBad, "invalid SIPrefix")
	})

	s.T().Run("SIPrefixSymbolMap spot check", func(t *testing.T) {
		assert.Equal(t, SI_PREFIX_NONE, SIPrefixSymbolMap[""], "SI_PREFIX_NONE")
		assert.Equal(t, SI_PREFIX_DEKA, SIPrefixSymbolMap["da"], "SI_PREFIX_DEKA")
		assert.Equal(t, SI_PREFIX_PICO, SIPrefixSymbolMap["p"], "SI_PREFIX_PICO")
		assert.Equal(t, SI_PREFIX_YOTTA, SIPrefixSymbolMap["Y"], "SI_PREFIX_YOTTA")
		assert.Equal(t, SI_PREFIX_YOCTO, SIPrefixSymbolMap["y"], "SI_PREFIX_YOCTO")
		assert.Equal(t, SI_PREFIX_MICRO, SIPrefixSymbolMap["u"], "SI_PREFIX_MICRO u")
		assert.Equal(t, SI_PREFIX_MICRO, SIPrefixSymbolMap["µ"], "SI_PREFIX_MICRO mu")
		_, okTera := SIPrefixSymbolMap["tera"]
		assert.False(t, okTera, "tera should not exist (not a symbol)")
		_, okDA := SIPrefixSymbolMap["DA"]
		assert.False(t, okDA, "DA should not exist (letter casing)")
		_, okQ := SIPrefixSymbolMap["Q"]
		assert.False(t, okQ, "Q should not exist")
	})

	s.T().Run("SIPrefixName all lowercase", func(t *testing.T) {
		for k, v := range SIPrefixName {
			assert.Equal(t, strings.ToLower(v), v, k.String())
		}
	})

	s.T().Run("SIPrefixName spot check", func(t *testing.T) {
		assert.Equal(t, "", SIPrefixName[SI_PREFIX_NONE], "SI_PREFIX_NONE")
		assert.Equal(t, "deci", SIPrefixName[SI_PREFIX_DECI], "SI_PREFIX_DECI")
		assert.Equal(t, "femto", SIPrefixName[SI_PREFIX_FEMTO], "SI_PREFIX_FEMTO")
		assert.Equal(t, "exa", SIPrefixName[SI_PREFIX_EXA], "SI_PREFIX_EXA")
		assert.Equal(t, "giga", SIPrefixName[SI_PREFIX_GIGA], "SI_PREFIX_GIGA")
		_, okBad := SIPrefixName[SIPrefix(100)]
		assert.False(t, okBad, "invalid SIPrefix")
	})

	s.T().Run("SIPrefixNameMap all lowercase", func(t *testing.T) {
		for k, v := range SIPrefixNameMap {
			assert.Equal(t, strings.ToLower(k), k, v.String())
		}
	})

	s.T().Run("SIPrefixNameMap spot check", func(t *testing.T) {
		assert.Equal(t, SI_PREFIX_NONE, SIPrefixNameMap[""], "SI_PREFIX_NONE")
		assert.Equal(t, SI_PREFIX_NONE, SIPrefixNameMap["none"], "SI_PREFIX_NONE")
		assert.Equal(t, SI_PREFIX_MICRO, SIPrefixNameMap["micro"], "SI_PREFIX_MICRO")
		assert.Equal(t, SI_PREFIX_ATTO, SIPrefixNameMap["atto"], "SI_PREFIX_ATTO")
		assert.Equal(t, SI_PREFIX_TERA, SIPrefixNameMap["tera"], "SI_PREFIX_TERA")
		_, okTera := SIPrefixNameMap["Tera"]
		assert.False(t, okTera, "Tera should not exist (upper-case letter)")
		_, okSpace := SIPrefixNameMap[" "]
		assert.False(t, okSpace, "{space} should not exist")
		_, okZepto := SIPrefixNameMap["ZEPTO"]
		assert.False(t, okZepto, "ZEPTO should not exist (upper-case letters)")
	})
}

func (s *SITestSuite) TestSIPrefixFromString() {
	tests := []struct {
		name       string
		input      string
		exSIPrefix SIPrefix
		exErr      string
	}{
		{
			name: "empty string",
			input: "",
			exSIPrefix: SI_PREFIX_NONE,
			exErr: "",
		},
		{
			name: "enum name all upper-case",
			input: "SI_PREFIX_GIGA",
			exSIPrefix: SI_PREFIX_GIGA,
			exErr: "",
		},
		{
			name: "enum name all lower-case",
			input: "si_prefix_deci",
			exSIPrefix: SI_PREFIX_DECI,
			exErr: "",
		},
		{
			name: "enum name mixed case",
			input: "Si_PreFix_MiCro",
			exSIPrefix: SI_PREFIX_MICRO,
			exErr: "",
		},
		{
			name: "enum name does not exist",
			input: "SI_PREFIX_NOPE_NOPE",
			exSIPrefix: invalidSIPrefix,
			exErr: "could not convert string [SI_PREFIX_NOPE_NOPE] to a SIPrefix value",
		},
		{
			name: "just name all upper-case",
			input: "ZETTA",
			exSIPrefix: SI_PREFIX_ZETTA,
			exErr: "",
		},
		{
			name: "just name all lower-case",
			input: "kilo",
			exSIPrefix: SI_PREFIX_KILO,
			exErr: "",
		},
		{
			name: "just name mixed case",
			input: "eXa",
			exSIPrefix: SI_PREFIX_EXA,
			exErr: "",
		},
		{
			name: "just name does not exist",
			input: "small",
			exSIPrefix: invalidSIPrefix,
			exErr: "could not convert string [small] to a SIPrefix value",
		},
		{
			name: "symbol lower-case where upper-case also exists",
			input: "m",
			exSIPrefix: SI_PREFIX_MILLI,
			exErr: "",
		},
		{
			name: "symbol upper-case where lower-case also exists",
			input: "M",
			exSIPrefix: SI_PREFIX_MEGA,
			exErr: "",
		},
		{
			name: "symbol da",
			input: "da",
			exSIPrefix: SI_PREFIX_DEKA,
			exErr: "",
		},
		{
			name: "symbol mu",
			input: "µ",
			exSIPrefix: SI_PREFIX_MICRO,
			exErr: "",
		},
		{
			name: "symbol u",
			input: "u",
			exSIPrefix: SI_PREFIX_MICRO,
			exErr: "",
		},
		{
			name: "symbol wrong casing",
			input: "D",
			exSIPrefix: invalidSIPrefix,
			exErr: "could not convert string [D] to a SIPrefix value",
		},
		{
			name: "symbol does not exist",
			input: "l",
			exSIPrefix: invalidSIPrefix,
			exErr: "could not convert string [l] to a SIPrefix value",
		},
		{
			name: "exponent 0",
			input: "0",
			exSIPrefix: SI_PREFIX_NONE,
			exErr: "",
		},
		{
			name: "exponent -3",
			input: "-3",
			exSIPrefix: SI_PREFIX_MILLI,
			exErr: "",
		},
		{
			name: "exponent +24",
			input: "+24",
			exSIPrefix: SI_PREFIX_YOTTA,
			exErr: "",
		},
		{
			name: "exponent 10 does not exist",
			input: "10",
			exSIPrefix: invalidSIPrefix,
			exErr: "could not convert exponent [10] to a SIPrefix value",
		},
	}

	// Run tests on the SIPrefixFromString method.
	for _, tc := range tests {
		s.T().Run(fmt.Sprintf("SIPrefixFromString %s", tc.name), func(t *testing.T) {
			p, err := SIPrefixFromString(tc.input)
			if len(tc.exErr) > 0 {
				assert.EqualError(t, err, tc.exErr, "SIPrefixFromString(\"%s\") expected error", tc.input)
				if p != invalidSIPrefix {
					assert.Fail(t, fmt.Sprintf("SIPrefixFromString(\"%s\") unexpected result [%s]", tc.input, p))
				}
			} else {
				assert.NoError(t, err, "SIPrefixFromString(\"%s\") unexpected error", tc.input)
				assert.Equal(t, tc.exSIPrefix, p, "SIPrefixFromString(\"%s\") expected result", tc.input)
			}
		})
	}

	// Run tests on the MustGetSIPrefixFromString method.
	for _, tc := range tests {
		s.T().Run(fmt.Sprintf("MustGetSIPrefixFromString %s", tc.name), func(t *testing.T) {
			if len(tc.exErr) > 0 {
				assert.PanicsWithError(t, tc.exErr, func() {
					_ = MustGetSIPrefixFromString(tc.input)
				}, "MustGetSIPrefixFromString(\"%s\") expected panic", tc.input)
			} else {
				assert.NotPanics(t, func() {
					p := MustGetSIPrefixFromString(tc.input)
					assert.Equal(t, tc.exSIPrefix, p, "MustGetSIPrefixFromString(\"%s\") expected result", tc.input)
				}, "MustGetSIPrefixFromString(\"%s\") unexpected panic", tc.input)
			}
		})
	}
}

func (s *SITestSuite) TestSIPrefixFromExponent() {
	valid := []struct {
		input  int
		output SIPrefix
	}{
		{24, SI_PREFIX_YOTTA},
		{21, SI_PREFIX_ZETTA},
		{18, SI_PREFIX_EXA},
		{15, SI_PREFIX_PETA},
		{12, SI_PREFIX_TERA},
		{9, SI_PREFIX_GIGA},
		{6, SI_PREFIX_MEGA},
		{3, SI_PREFIX_KILO},
		{2, SI_PREFIX_HECTO},
		{1, SI_PREFIX_DEKA},
		{0, SI_PREFIX_NONE},
		{-1, SI_PREFIX_DECI},
		{-2, SI_PREFIX_CENTI},
		{-3, SI_PREFIX_MILLI},
		{-6, SI_PREFIX_MICRO},
		{-9, SI_PREFIX_NANO},
		{-12, SI_PREFIX_PICO},
		{-15, SI_PREFIX_FEMTO},
		{-18, SI_PREFIX_ATTO},
		{-21, SI_PREFIX_ZEPTO},
		{-24, SI_PREFIX_YOCTO},
	}
	invalid := []int{4, 13, 25, 27, -5, -27, -100}

	// Run valid tests for SIPrefixFromExponent.
	for _, tc := range valid {
		s.T().Run(fmt.Sprintf("SIPrefixFromExponent valid %s", tc.output), func(t *testing.T) {
			p, err := SIPrefixFromExponent(tc.input)
			assert.NoError(t, err, "SIPrefixFromExponent(%d) unexpected error", tc.input)
			assert.Equal(t, tc.output, p, "SIPrefixFromExponent(%d) result", tc.input)
		})
	}

	// Run valid tests for MustGetSIPrefixFromExponent.
	for _, tc := range valid {
		s.T().Run(fmt.Sprintf("MustGetSIPrefixFromExponent valid %s", tc.output), func(t *testing.T) {
			assert.NotPanics(t, func() {
				p := MustGetSIPrefixFromExponent(tc.input)
				assert.Equal(t, tc.output, p, "MustGetSIPrefixFromExponent(%d) result", tc.input)
			}, "MustGetSIPrefixFromExponent(%d) unexpected panic", tc)
		})
	}

	// Run invalid tests for SIPrefixFromExponent.
	for _, tc := range invalid {
		s.T().Run(fmt.Sprintf("SIPrefixFromExponent invalid %d", tc), func(t *testing.T) {
			p, err := SIPrefixFromExponent(tc)
			assert.EqualError(t, err,
				fmt.Sprintf("could not convert exponent [%d] to a SIPrefix value", tc),
				"SIPrefixFromExponent(%d) expected error", tc)
			if p != invalidSIPrefix {
				assert.Fail(t, fmt.Sprintf("SIPrefixFromExponent(%d) unexpected result [%s]", tc, p))
			}
		})
	}

	// Run valid tests for MustGetSIPrefixFromExponent.
	for _, tc := range invalid {
		s.T().Run(fmt.Sprintf("MustGetSIPrefixFromExponent invalid %d", tc), func(t *testing.T) {
			assert.PanicsWithError(t, fmt.Sprintf("could not convert exponent [%d] to a SIPrefix value", tc), func() {
				_ = MustGetSIPrefixFromExponent(tc)
			}, "MustGetSIPrefixFromExponent(%d) expected panic", tc)
		})
	}
}

func (s *SITestSuite) TestParseSIPrefixedString() {
	tests := []struct {
		name     string
		val      string
		root     string
		exPrefix SIPrefix
	}{
		{
			name: "val shorter than root",
			val: "abc",
			root: "abcd",
			exPrefix: invalidSIPrefix,
		},
		{
			name:"val equals root same case",
			val:"abcd",
			root:"abcd",
			exPrefix:SI_PREFIX_NONE,
		},
		{
			name:"val equals root different case",
			val:"abcd",
			root:"ABCD",
			exPrefix:SI_PREFIX_NONE,
		},
		{
			name:"val does not end in root",
			val:"Mbce",
			root:"bcd",
			exPrefix:invalidSIPrefix,
		},
		{
			name:"val has name prefix",
			val:"megabcd",
			root:"bcd",
			exPrefix:SI_PREFIX_MEGA,
		},
		{
			name:"val has name prefix upper case",
			val:"MICROBCD",
			root:"bcd",
			exPrefix:SI_PREFIX_MICRO,
		},
		{
			name:"val has symbol prefix val root is lower case",
			val:"nbcd",
			root:"bcd",
			exPrefix:SI_PREFIX_NANO,
		},
		{
			name:"val has symbol prefix val root is upper case",
			val:"kBCD",
			root:"bcd",
			exPrefix:SI_PREFIX_KILO,
		},
		{
			name:"val has unknown prefix",
			val:"unknownbcd",
			root:"bcd",
			exPrefix:invalidSIPrefix,
		},
		{
			name:"root is empty string val is a prefix name",
			val:"tera",
			root:"",
			exPrefix:SI_PREFIX_TERA,
		},
		{
			name:"root is empty string val is a prefix symbol",
			val:"u",
			root:"",
			exPrefix:SI_PREFIX_MICRO,
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			p, ok := ParseSIPrefixedString(tc.val, tc.root)
			assert.Equal(t, tc.exPrefix, p, "ParseSIPrefixedString(\"%s\", \"%s\") result prefix", tc.val, tc.root)
			assert.Equal(t, tc.exPrefix != invalidSIPrefix, ok, "ParseSIPrefixedString(\"%s\", \"%s\") result bool", tc.val, tc.root)
		})
	}
}

func (s *SITestSuite) TestIsValid() {
	tests := []struct {
		name     string
		prefix   SIPrefix
		expected bool
	}{
		{"SI_PREFIX_NONE", SI_PREFIX_NONE, true},
		{"SI_PREFIX_ZETTA", SI_PREFIX_ZETTA, true},
		{"SI_PREFIX_ZEPTO", SI_PREFIX_ZEPTO, true},
		{"invalidSIPrefix", invalidSIPrefix, false},
		{"invalid from cast 13", SIPrefix(13), false},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			actual := tc.prefix.IsValid()
			assert.Equal(t, tc.expected, actual, "%s.IsValid() result", tc.prefix)
		})
	}
}

func (s *SITestSuite) TestFormat() {
	tests := []struct {
		name     string
		verb     string
		value    SIPrefix
		expected string
	}{
		{
			name: "s SI_PREFIX_NONE",
			verb: "s",
			value: SI_PREFIX_NONE,
			expected: "SI_PREFIX_NONE",
		},
		{
			name: "s SI_PREFIX_YOTTA",
			verb: "s",
			value: SI_PREFIX_YOTTA,
			expected: "SI_PREFIX_YOTTA",
		},
		{
			name: "s SI_PREFIX_NANO",
			verb: "s",
			value: SI_PREFIX_NANO,
			expected: "SI_PREFIX_NANO",
		},
		{
			name: "s invalid int -20",
			verb: "s",
			value: SIPrefix(-20),
			expected: "-20",
		},
		{
			name: "d SI_PREFIX_EXA",
			verb: "d",
			value: SI_PREFIX_EXA,
			expected: "1000000000000000000",
		},
		{
			name: "d SI_PREFIX_NANO",
			verb: "d",
			value: SI_PREFIX_NANO,
			expected: "0.000000001",
		},
		{
			name: "d invalid int 4",
			verb: "d",
			value: SIPrefix(4),
			expected: "10000",
		},
		{
			name: "d invalid int -5",
			verb: "d",
			value: SIPrefix(-5),
			expected: "0.00001",
		},
		{
			name: "e SI_PREFIX_GIGA",
			verb: "e",
			value: SI_PREFIX_GIGA,
			expected: "1e+9",
		},
		{
			name: "e SI_PREFIX_CENTI",
			verb: "e",
			value: SI_PREFIX_CENTI,
			expected: "1e-2",
		},
		{
			name: "E SI_PREFIX_ZETTA",
			verb: "E",
			value: SI_PREFIX_ZETTA,
			expected: "1E+21",
		},
		{
			name: "E SI_PREFIX_FEMTO",
			verb: "E",
			value: SI_PREFIX_FEMTO,
			expected: "1E-15",
		},
		{
			name: "v SI_PREFIX_KILO",
			verb: "v",
			value: SI_PREFIX_KILO,
			expected: "3",
		},
		{
			name: "v SI_PREFIX_DECI",
			verb: "v",
			value: SI_PREFIX_DECI,
			expected: "-1",
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			f := fmt.Sprintf("%%%s", tc.verb)
			actual := fmt.Sprintf(f, tc.value)
			assert.Equal(t, tc.expected, actual, "fmt.Sprintf(\"%s\", %s)", f, tc.value)
		})
	}
}

func (s *SITestSuite) TestGetName() {
	tests := []struct {
		name     string
		prefix   SIPrefix
		expected string
	}{
		{"SI_PREFIX_YOTTA", SI_PREFIX_YOTTA, "yotta"},
		{"SI_PREFIX_ZETTA", SI_PREFIX_ZETTA, "zetta"},
		{"SI_PREFIX_EXA", SI_PREFIX_EXA, "exa"},
		{"SI_PREFIX_PETA", SI_PREFIX_PETA, "peta"},
		{"SI_PREFIX_TERA", SI_PREFIX_TERA, "tera"},
		{"SI_PREFIX_GIGA", SI_PREFIX_GIGA, "giga"},
		{"SI_PREFIX_MEGA", SI_PREFIX_MEGA, "mega"},
		{"SI_PREFIX_KILO", SI_PREFIX_KILO, "kilo"},
		{"SI_PREFIX_HECTO", SI_PREFIX_HECTO, "hecto"},
		{"SI_PREFIX_DEKA", SI_PREFIX_DEKA, "deka"},

		{"SI_PREFIX_NONE", SI_PREFIX_NONE, ""},

		{"SI_PREFIX_DECI", SI_PREFIX_DECI, "deci"},
		{"SI_PREFIX_CENTI", SI_PREFIX_CENTI, "centi"},
		{"SI_PREFIX_MILLI", SI_PREFIX_MILLI, "milli"},
		{"SI_PREFIX_MICRO", SI_PREFIX_MICRO, "micro"},
		{"SI_PREFIX_NANO", SI_PREFIX_NANO, "nano"},
		{"SI_PREFIX_PICO", SI_PREFIX_PICO, "pico"},
		{"SI_PREFIX_FEMTO", SI_PREFIX_FEMTO, "femto"},
		{"SI_PREFIX_ATTO", SI_PREFIX_ATTO, "atto"},
		{"SI_PREFIX_ZEPTO", SI_PREFIX_ZEPTO, "zepto"},
		{"SI_PREFIX_YOCTO", SI_PREFIX_YOCTO, "yocto"},

		{"invalid 20", SIPrefix(20), "invalid"},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			actual := tc.prefix.GetName()
			assert.Equal(t, tc.expected, actual, "%s.GetName()", tc.prefix)
		})
	}
}

func (s *SITestSuite) TestGetSymbol() {
	tests := []struct {
		name     string
		prefix   SIPrefix
		expected string
	}{
		{"SI_PREFIX_YOTTA", SI_PREFIX_YOTTA, "Y"},
		{"SI_PREFIX_ZETTA", SI_PREFIX_ZETTA, "Z"},
		{"SI_PREFIX_EXA", SI_PREFIX_EXA, "E"},
		{"SI_PREFIX_PETA", SI_PREFIX_PETA, "P"},
		{"SI_PREFIX_TERA", SI_PREFIX_TERA, "T"},
		{"SI_PREFIX_GIGA", SI_PREFIX_GIGA, "G"},
		{"SI_PREFIX_MEGA", SI_PREFIX_MEGA, "M"},
		{"SI_PREFIX_KILO", SI_PREFIX_KILO, "k"},
		{"SI_PREFIX_HECTO", SI_PREFIX_HECTO, "h"},
		{"SI_PREFIX_DEKA", SI_PREFIX_DEKA, "da"},

		{"SI_PREFIX_NONE", SI_PREFIX_NONE, ""},

		{"SI_PREFIX_DECI", SI_PREFIX_DECI, "d"},
		{"SI_PREFIX_CENTI", SI_PREFIX_CENTI, "c"},
		{"SI_PREFIX_MILLI", SI_PREFIX_MILLI, "m"},
		{"SI_PREFIX_MICRO", SI_PREFIX_MICRO, "µ"},
		{"SI_PREFIX_NANO", SI_PREFIX_NANO, "n"},
		{"SI_PREFIX_PICO", SI_PREFIX_PICO, "p"},
		{"SI_PREFIX_FEMTO", SI_PREFIX_FEMTO, "f"},
		{"SI_PREFIX_ATTO", SI_PREFIX_ATTO, "a"},
		{"SI_PREFIX_ZEPTO", SI_PREFIX_ZEPTO, "z"},
		{"SI_PREFIX_YOCTO", SI_PREFIX_YOCTO, "y"},

		{"invalid -20", SIPrefix(-20), "INV"},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			actual := tc.prefix.GetSymbol()
			assert.Equal(t, tc.expected, actual, "%s.GetSymbol()", tc.prefix)
		})
	}
}

func (s *SITestSuite) TestGetDecimalString() {
	tests := []struct {
		name     string
		prefix   SIPrefix
		expected string
	}{
		{"SI_PREFIX_YOTTA", SI_PREFIX_YOTTA, "1000000000000000000000000"},
		{"SI_PREFIX_ZETTA", SI_PREFIX_ZETTA, "1000000000000000000000"},
		{"SI_PREFIX_EXA", SI_PREFIX_EXA, "1000000000000000000"},
		{"SI_PREFIX_PETA", SI_PREFIX_PETA, "1000000000000000"},
		{"SI_PREFIX_TERA", SI_PREFIX_TERA, "1000000000000"},
		{"SI_PREFIX_GIGA", SI_PREFIX_GIGA, "1000000000"},
		{"SI_PREFIX_MEGA", SI_PREFIX_MEGA, "1000000"},
		{"SI_PREFIX_KILO", SI_PREFIX_KILO, "1000"},
		{"SI_PREFIX_HECTO", SI_PREFIX_HECTO, "100"},
		{"SI_PREFIX_DEKA", SI_PREFIX_DEKA, "10"},

		{"SI_PREFIX_NONE", SI_PREFIX_NONE, "1"},

		{"SI_PREFIX_DECI", SI_PREFIX_DECI, "0.1"},
		{"SI_PREFIX_CENTI", SI_PREFIX_CENTI, "0.01"},
		{"SI_PREFIX_MILLI", SI_PREFIX_MILLI, "0.001"},
		{"SI_PREFIX_MICRO", SI_PREFIX_MICRO, "0.000001"},
		{"SI_PREFIX_NANO", SI_PREFIX_NANO, "0.000000001"},
		{"SI_PREFIX_PICO", SI_PREFIX_PICO, "0.000000000001"},
		{"SI_PREFIX_FEMTO", SI_PREFIX_FEMTO, "0.000000000000001"},
		{"SI_PREFIX_ATTO", SI_PREFIX_ATTO, "0.000000000000000001"},
		{"SI_PREFIX_ZEPTO", SI_PREFIX_ZEPTO, "0.000000000000000000001"},
		{"SI_PREFIX_YOCTO", SI_PREFIX_YOCTO, "0.000000000000000000000001"},

		{"invalid -7", SIPrefix(-7), "0.0000001"},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			actual := tc.prefix.GetDecimalString()
			assert.Equal(t, tc.expected, actual, "%s.GetDecimalString()", tc.prefix)
		})
	}
}

func (s *SITestSuite) TestGetExponentString() {
	tests := []struct {
		name     string
		prefix   SIPrefix
		expected string
	}{
		{"SI_PREFIX_YOTTA", SI_PREFIX_YOTTA, "1e+24"},
		{"SI_PREFIX_ZETTA", SI_PREFIX_ZETTA, "1e+21"},
		{"SI_PREFIX_EXA", SI_PREFIX_EXA, "1e+18"},
		{"SI_PREFIX_PETA", SI_PREFIX_PETA, "1e+15"},
		{"SI_PREFIX_TERA", SI_PREFIX_TERA, "1e+12"},
		{"SI_PREFIX_GIGA", SI_PREFIX_GIGA, "1e+9"},
		{"SI_PREFIX_MEGA", SI_PREFIX_MEGA, "1e+6"},
		{"SI_PREFIX_KILO", SI_PREFIX_KILO, "1e+3"},
		{"SI_PREFIX_HECTO", SI_PREFIX_HECTO, "1e+2"},
		{"SI_PREFIX_DEKA", SI_PREFIX_DEKA, "1e+1"},

		{"SI_PREFIX_NONE", SI_PREFIX_NONE, "1e+0"},

		{"SI_PREFIX_DECI", SI_PREFIX_DECI, "1e-1"},
		{"SI_PREFIX_CENTI", SI_PREFIX_CENTI, "1e-2"},
		{"SI_PREFIX_MILLI", SI_PREFIX_MILLI, "1e-3"},
		{"SI_PREFIX_MICRO", SI_PREFIX_MICRO, "1e-6"},
		{"SI_PREFIX_NANO", SI_PREFIX_NANO, "1e-9"},
		{"SI_PREFIX_PICO", SI_PREFIX_PICO, "1e-12"},
		{"SI_PREFIX_FEMTO", SI_PREFIX_FEMTO, "1e-15"},
		{"SI_PREFIX_ATTO", SI_PREFIX_ATTO, "1e-18"},
		{"SI_PREFIX_ZEPTO", SI_PREFIX_ZEPTO, "1e-21"},
		{"SI_PREFIX_YOCTO", SI_PREFIX_YOCTO, "1e-24"},

		{"invalid 11", SIPrefix(11), "1e+11"},
		{"invalid -8", SIPrefix(-8), "1e-8"},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			actual := tc.prefix.GetExponentString()
			assert.Equal(t, tc.expected, actual, "%s.GetExponentString()", tc.prefix)
		})
	}
}

func (s *SITestSuite) TestGetExponent() {
	tests := []struct {
		name     string
		prefix   SIPrefix
		expected int
	}{
		{"SI_PREFIX_YOTTA", SI_PREFIX_YOTTA, 24},
		{"SI_PREFIX_ZETTA", SI_PREFIX_ZETTA, 21},
		{"SI_PREFIX_EXA", SI_PREFIX_EXA, 18},
		{"SI_PREFIX_PETA", SI_PREFIX_PETA, 15},
		{"SI_PREFIX_TERA", SI_PREFIX_TERA, 12},
		{"SI_PREFIX_GIGA", SI_PREFIX_GIGA, 9},
		{"SI_PREFIX_MEGA", SI_PREFIX_MEGA, 6},
		{"SI_PREFIX_KILO", SI_PREFIX_KILO, 3},
		{"SI_PREFIX_HECTO", SI_PREFIX_HECTO, 2},
		{"SI_PREFIX_DEKA", SI_PREFIX_DEKA, 1},

		{"SI_PREFIX_NONE", SI_PREFIX_NONE, 0},

		{"SI_PREFIX_DECI", SI_PREFIX_DECI, -1},
		{"SI_PREFIX_CENTI", SI_PREFIX_CENTI, -2},
		{"SI_PREFIX_MILLI", SI_PREFIX_MILLI, -3},
		{"SI_PREFIX_MICRO", SI_PREFIX_MICRO, -6},
		{"SI_PREFIX_NANO", SI_PREFIX_NANO, -9},
		{"SI_PREFIX_PICO", SI_PREFIX_PICO, -12},
		{"SI_PREFIX_FEMTO", SI_PREFIX_FEMTO, -15},
		{"SI_PREFIX_ATTO", SI_PREFIX_ATTO, -18},
		{"SI_PREFIX_ZEPTO", SI_PREFIX_ZEPTO, -21},
		{"SI_PREFIX_YOCTO", SI_PREFIX_YOCTO, -24},

		{"invalid 11", SIPrefix(11), 11},
		{"invalid -8", SIPrefix(-8), -8},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			actual := tc.prefix.GetExponent()
			assert.Equal(t, tc.expected, actual, "%s.GetExponent()", tc.prefix)
		})
	}
}
