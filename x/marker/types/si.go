package types

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

var (
	// SIPrefixSymbol is used to look up the symbol for a SIPrefix enum entry.
	SIPrefixSymbol = map[SIPrefix]string{
		SI_PREFIX_NONE: "",

		SI_PREFIX_DEKA:  "da",
		SI_PREFIX_HECTO: "h",
		SI_PREFIX_KILO:  "k",
		SI_PREFIX_MEGA:  "M",
		SI_PREFIX_GIGA:  "G",
		SI_PREFIX_TERA:  "T",
		SI_PREFIX_PETA:  "P",
		SI_PREFIX_EXA:   "E",
		SI_PREFIX_ZETTA: "Z",
		SI_PREFIX_YOTTA: "Y",

		SI_PREFIX_DECI:  "d",
		SI_PREFIX_CENTI: "c",
		SI_PREFIX_MILLI: "m",
		SI_PREFIX_MICRO: "µ",
		SI_PREFIX_NANO:  "n",
		SI_PREFIX_PICO:  "p",
		SI_PREFIX_FEMTO: "f",
		SI_PREFIX_ATTO:  "a",
		SI_PREFIX_ZEPTO: "z",
		SI_PREFIX_YOCTO: "y",
	}

	// SIPrefixSymbolMap is used to look up the SIPrefix enum entry for a symbol.
	// Some SIPrefix values might appear more than once in this map, e.g. "u" and "µ" are both for SI_PREFIX_MICRO.
	SIPrefixSymbolMap map[string]SIPrefix

	// SIPrefixName is used to look up the name for a SIPrefix enum entry.
	// The values are all lower-case.
	SIPrefixName map[SIPrefix]string

	// SIPrefixSymbolMap is used to look up the SIPrefix enum entry for a name.
	// The keys are all lower-case.
	// Some SIPrefix values might appear more than once in this map, e.g. "" and "none" are both for SI_PREFIX_NONE.
	SIPrefixNameMap map[string]SIPrefix

	// invalidSIPrefix is an int32 that's been converted to a SIPrefix but doesn't have a valid value.
	invalidSIPrefix = SIPrefix(0xff)
)

// Populate the SIPrefixSymbolMap from SIPrefixSymbol, and make sure that there's an entry for every SIPrefix.
// Also populate the SIPrefixName and SIPrefixNameMap maps.
func init() {
	SIPrefixSymbolMap = make(map[string]SIPrefix)
	SIPrefixName = make(map[SIPrefix]string)
	SIPrefixNameMap = make(map[string]SIPrefix)
	for i, str := range SIPrefix_name {
		p := SIPrefix(i)
		if s, ok := SIPrefixSymbol[p]; ok {
			SIPrefixSymbolMap[s] = p
		} else {
			panic(fmt.Errorf("no SIPrefixSymbol entry defined for [%s]", str))
		}
		name := strings.ToLower(str[10:])
		SIPrefixName[p] = name
		SIPrefixNameMap[name] = p
	}
	// SI_PREFIX_NONE is a special case. Override the name for it with "", and add "" as a name mapping back to it.
	SIPrefixName[SI_PREFIX_NONE] = ""
	SIPrefixNameMap[""] = SI_PREFIX_NONE
	// SI_PREFIX_MICRO has a difficult symbol: Greek lowercase mu. To be nice, also allow a "u" there.
	SIPrefixSymbolMap["u"] = SI_PREFIX_MICRO
}

// MustGetSIPrefixFromString turns a string into a SIPrefix enum entry or panics if invalid.
// Example inputs: "SI_PREFIX_PICO", "PICO", "p", "M", "µ".
// Note: An empty string will produce SI_PREFIX_NONE.
func MustGetSIPrefixFromString(str string) SIPrefix {
	s, err := SIPrefixFromString(str)
	if err != nil {
		panic(err)
	}
	return s
}

// SIPrefixFromString attempts to convert a string into a SIPrefix enum entry.
// Example inputs: "SI_PREFIX_GIGA", "GIGA", "G", "m", "µ".
// Note: An empty string will produce SI_PREFIX_NONE.
func SIPrefixFromString(str string) (SIPrefix, error) {
	// Check if it's a name (names are all lower-case).
	if val, ok := SIPrefixNameMap[strings.ToLower(str)]; ok {
		return val, nil
	}
	// Check if it's an enum string value (enum string values are all upper-case).
	if val, ok := SIPrefix_value[strings.ToUpper(str)]; ok {
		return SIPrefix(val), nil
	}
	// Check if it's a symbol (case sensitive here).
	if val, ok := SIPrefixSymbolMap[str]; ok {
		return val, nil
	}
	// Check if it's an enum int32 value that was converted to a string (because why not?).
	if exp, err := strconv.ParseInt(str, 10, 32); err == nil {
		return SIPrefixFromExponent(int(exp))
	}
	// Give up.
	return invalidSIPrefix, fmt.Errorf("could not convert string [%s] to a SIPrefix value", str)
}

// MustGetSIPrefixFromExponent turns an integer exponent into a SIPrefix enum entry or panics if invalid.
// Example inputs: 15, -6
func MustGetSIPrefixFromExponent(exp int) SIPrefix {
	s, err := SIPrefixFromExponent(exp)
	if err != nil {
		panic(err)
	}
	return s
}

// SIPrefixFromExponent attempts to convert an integer exponent into a SIPrefix enum entry.
// Example inputs: 12, -9
func SIPrefixFromExponent(exp int) (SIPrefix, error) {
	if exp < math.MinInt32 || exp > math.MaxInt32 {
		return invalidSIPrefix, fmt.Errorf("exponent [%d] out of bounds for int32", exp)
	}
	p := SIPrefix(exp)
	if p.IsValid() {
		return p, nil
	}
	return invalidSIPrefix, fmt.Errorf("could not convert exponent [%d] to a SIPrefix value", exp)
}

// ParseSIPrefixedString extracts the prefix from the provided val using root as the base.
// Returns the SI prefix, and a boolean to indicate that it was successful.
// Possible reasons for it to be unsuccessful:
//  - The provided val is shorter than the root.
//  - The right-most characters in val are not equal to root (case insensitive).
//  - No SI Prefix can be found matching the left-most portion of val (that isn't root).
func ParseSIPrefixedString(val string, root string) (SIPrefix, bool) {
	if len(val) < len(root) {
		return invalidSIPrefix, false
	}
	if strings.EqualFold(val, root) {
		return SI_PREFIX_NONE, true
	}
	valRoot := val[len(val)-len(root):]
	if !strings.EqualFold(root, valRoot) {
		return invalidSIPrefix, false
	}
	prefix := val[:len(val)-len(root)]
	if p, ok := SIPrefixNameMap[strings.ToLower(prefix)]; ok {
		return p, true
	}
	if p, ok := SIPrefixSymbolMap[prefix]; ok {
		return p, true
	}
	return invalidSIPrefix, false
}

// IsValid checks that this SIPrefix is a valid value.
func (p SIPrefix) IsValid() bool {
	_, ok := SIPrefix_name[int32(p)]
	return ok
}

// Format implements the fmt.Formatter interface.
func (p SIPrefix) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		// The enum name.
		s.Write([]byte(p.String()))
	case 'd':
		// The decimal form of 1 * 10 ^ p, e.g. 1000000000 or 0.000000000000001
		s.Write([]byte(p.GetDecimalString()))
	case 'e':
		// Exponential form with a lower-case "e", e.g. 1e+9 or 1e-15
		s.Write([]byte(p.GetExponentString()))
	case 'E':
		// Exponential form with an upper-case "E", e.g. 1E+9 or 1E-15
		s.Write([]byte(strings.ToUpper(p.GetExponentString())))
	default:
		// Anything else should just use the int value, e.g. 9 or -15.
		s.Write([]byte(fmt.Sprint(int(p))))
	}
}

// GetName gets the lower-case name of this prefix, e.g. "nano", "exa", or "micro".
func (p SIPrefix) GetName() string {
	if str, ok := SIPrefixName[p]; ok {
		return str
	}
	// should only happen for invalid SIPrefix values.
	return "invalid"
}

// GetSymbol gets the symbol for this prefix, e.g. "n", "E", or "µ".
// Case matters here. E.g. "m" is "milli" while "M" is "mega".
func (p SIPrefix) GetSymbol() string {
	if s, ok := SIPrefixSymbol[p]; ok {
		return s
	}
	// should only happen for invalid SIPrefix values.
	return "INV"
}

// GetDecimalString gets a string with the decimal representation of the multiplier for this SI Prefix.
// Examples: SI_PREFIX_ZETTA becomes "1000000000000000000000", and SI_PREFIX_ATTO becomes "0.000000000000000001".
// It is a string because a float64 has rounding issues on the longer numbers.
func (p SIPrefix) GetDecimalString() string {
	if p == 0 {
		return "1"
	}
	if p > 0 {
		return "1" + strings.Repeat("0", int(p))
	}
	return "0." + strings.Repeat("0", -1*int(p)-1) + "1"
}

// GetExponentString gets an exponent representation of the multiplier for this SI Prefix.
// Examples: "1e+6", "1e-12".
func (p SIPrefix) GetExponentString() string {
	return fmt.Sprintf("1e%+d", int(p))
}

// Get the Exponent value of this SIPrefix.
// Examples, SI_PREFIX_PETA is 15 and SI_PREFIX_MICRO is -6.
func (p SIPrefix) GetExponent() int {
	return int(p)
}
