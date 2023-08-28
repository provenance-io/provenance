package exchange

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO[1658]: func TestMarket_Validate(t *testing.T)

// TODO[1658]: func TestMarketDetails_Validate(t *testing.T)

// TODO[1658]: func TestValidateSellerFeeRatios(t *testing.T)

// TODO[1658]: func TestValidateBuyerFeeRatios(t *testing.T)

// TODO[1658]: func TestValidateFeeRatios(t *testing.T)

// TODO[1658]: func TestFeeRatio_String(t *testing.T)

// TODO[1658]: func TestFeeRatio_Validate(t *testing.T)

// TODO[1658]: func TestValidateAccessGrants(t *testing.T)

// TODO[1658]: func TestAccessGrant_Validate(t *testing.T)

// TODO[1658]: func TestPermission_SimpleString(t *testing.T)

// TODO[1658]: func TestPermission_Validate(t *testing.T)

// TODO[1658]: func TestAllPermissions(t *testing.T)

// TODO[1658]: func TestParsePermission(t *testing.T)

// TODO[1658]: func TestParsePermissions(t *testing.T)

// TODO[1658]: func TestValidateReqAttrs(t *testing.T)

func TestIsValidReqAttr(t *testing.T) {
	tests := []struct {
		name    string
		reqAttr string
		exp     bool
	}{
		{name: "already valid and normalized", reqAttr: "x.y.z", exp: true},
		{name: "already valid but not normalized", reqAttr: " x . y . z ", exp: true},
		{name: "invalid character", reqAttr: "x._y.z", exp: false},
		{name: "just the wildcard", reqAttr: " * ", exp: true},
		{name: "just star dot", reqAttr: "*. ", exp: false},
		{name: "star dot valid", reqAttr: "* . x . y . z", exp: true},
		{name: "star dot invalid", reqAttr: "* . x . _y . z", exp: false},
		{name: "empty string", reqAttr: "", exp: false},
		{name: "wildcard in middle", reqAttr: "x.*.y.z", exp: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ok := IsValidReqAttr(tc.reqAttr)
			assert.Equal(t, tc.exp, ok, "IsValidReqAttr(%q)", tc.reqAttr)
		})
	}
}

func TestFindUnmatchedReqAttrs(t *testing.T) {
	tests := []struct {
		name     string
		reqAttrs []string
		accAttrs []string
		exp      []string
	}{
		{
			name:     "nil req attrs",
			reqAttrs: nil,
			accAttrs: []string{"one"},
			exp:      nil,
		},
		{
			name:     "empty req attrs",
			reqAttrs: []string{},
			accAttrs: []string{"one"},
			exp:      nil,
		},
		{
			name:     "one req attr no wildcard: in acc attrs",
			reqAttrs: []string{"one"},
			accAttrs: []string{"one", "two"},
			exp:      nil,
		},
		{
			name:     "one req attr with wildcard: in acc attrs",
			reqAttrs: []string{"*.one"},
			accAttrs: []string{"zero.one", "two"},
			exp:      nil,
		},
		{
			name:     "three req attrs: nil acc attrs",
			reqAttrs: []string{"*.desk", "nickel.dime.quarter", "*.x.y.z"},
			accAttrs: nil,
			exp:      []string{"*.desk", "nickel.dime.quarter", "*.x.y.z"},
		},
		{
			name:     "three req attrs: empty acc attrs",
			reqAttrs: []string{"*.desk", "nickel.dime.quarter", "*.x.y.z"},
			accAttrs: []string{},
			exp:      []string{"*.desk", "nickel.dime.quarter", "*.x.y.z"},
		},
		{
			name:     "three req attrs: only first in acc attrs",
			reqAttrs: []string{"*.desk", "nickel.dime.quarter", "*.x.y.z"},
			accAttrs: []string{"lamp.corner.desk"},
			exp:      []string{"nickel.dime.quarter", "*.x.y.z"},
		},
		{
			name:     "three req attrs: only second in acc attrs",
			reqAttrs: []string{"*.desk", "nickel.dime.quarter", "*.x.y.z"},
			accAttrs: []string{"nickel.dime.quarter"},
			exp:      []string{"*.desk", "*.x.y.z"},
		},
		{
			name:     "three req attrs: only third in acc attrs",
			reqAttrs: []string{"*.desk", "nickel.dime.quarter", "*.x.y.z"},
			accAttrs: []string{"w.x.y.z"},
			exp:      []string{"*.desk", "nickel.dime.quarter"},
		},
		{
			name:     "three req attrs: missing first",
			reqAttrs: []string{"*.desk", "nickel.dime.quarter", "*.x.y.z"},
			accAttrs: []string{"nickel.dime.quarter", "w.x.y.z"},
			exp:      []string{"*.desk"},
		},
		{
			name:     "three req attrs: missing middle",
			reqAttrs: []string{"*.desk", "nickel.dime.quarter", "*.x.y.z"},
			accAttrs: []string{"lamp.corner.desk", "w.x.y.z"},
			exp:      []string{"nickel.dime.quarter"},
		},
		{
			name:     "three req attrs: missing last",
			reqAttrs: []string{"*.desk", "nickel.dime.quarter", "*.x.y.z"},
			accAttrs: []string{"lamp.corner.desk", "nickel.dime.quarter"},
			exp:      []string{"*.x.y.z"},
		},
		{
			name:     "three req attrs: has all",
			reqAttrs: []string{"*.desk", "nickel.dime.quarter", "*.x.y.z"},
			accAttrs: []string{"just.some.first", "w.x.y.z", "other", "lamp.corner.desk", "random.entry", "nickel.dime.quarter", "what.is.this"},
			exp:      nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			unmatched := FindUnmatchedReqAttrs(tc.reqAttrs, tc.accAttrs)
			assert.Equal(t, tc.exp, unmatched, "FindUnmatchedReqAttrs(%q, %q", tc.reqAttrs, tc.accAttrs)
		})
	}
}

func TestHasReqAttrMatch(t *testing.T) {
	tests := []struct {
		name     string
		reqAttr  string
		accAttrs []string
		exp      bool
	}{
		{
			name:     "nil acc attrs",
			reqAttr:  "nickel.dime.quarter",
			accAttrs: nil,
			exp:      false,
		},
		{
			name:     "empty acc attrs",
			reqAttr:  "nickel.dime.quarter",
			accAttrs: []string{},
			exp:      false,
		},
		{
			name:    "no wildcard: not in acc attrs",
			reqAttr: "nickel.dime.quarter",
			accAttrs: []string{
				"xnickel.dime.quarter",
				"nickelx.dime.quarter",
				"nickel.xdime.quarter",
				"nickel.dimex.quarter",
				"nickel.dime.xquarter",
				"nickel.dime.quarterx",
				"penny.nickel.dime.quarter",
				"nickel.dime.quarter.dollar",
			},
			exp: false,
		},
		{
			name:     "no wildcard: only one in acc attrs",
			reqAttr:  "nickel.dime.quarter",
			accAttrs: []string{"nickel.dime.quarter"},
			exp:      true,
		},
		{
			name:    "no wildcard: first in acc attrs",
			reqAttr: "nickel.dime.quarter",
			accAttrs: []string{
				"nickel.dime.quarter",
				"xnickel.dime.quarter",
				"nickelx.dime.quarter",
				"nickel.xdime.quarter",
				"nickel.dimex.quarter",
				"nickel.dime.xquarter",
				"nickel.dime.quarterx",
				"penny.nickel.dime.quarter",
				"nickel.dime.quarter.dollar",
			},
			exp: true,
		},
		{
			name:    "no wildcard: in middle of acc attrs",
			reqAttr: "nickel.dime.quarter",
			accAttrs: []string{
				"xnickel.dime.quarter",
				"nickelx.dime.quarter",
				"nickel.xdime.quarter",
				"nickel.dimex.quarter",
				"nickel.dime.quarter",
				"nickel.dime.xquarter",
				"nickel.dime.quarterx",
				"penny.nickel.dime.quarter",
				"nickel.dime.quarter.dollar",
			},
			exp: true,
		},
		{
			name:    "no wildcard: at end of acc attrs",
			reqAttr: "nickel.dime.quarter",
			accAttrs: []string{
				"xnickel.dime.quarter",
				"nickelx.dime.quarter",
				"nickel.xdime.quarter",
				"nickel.dimex.quarter",
				"nickel.dime.xquarter",
				"nickel.dime.quarterx",
				"penny.nickel.dime.quarter",
				"nickel.dime.quarter.dollar",
				"nickel.dime.quarter",
			},
			exp: true,
		},

		{
			name:    "with wildcard: no match",
			reqAttr: "*.dime.quarter",
			accAttrs: []string{
				"dime.quarter",
				"penny.xdime.quarter",
				"penny.dimex.quarter",
				"penny.dime.xquarter",
				"penny.dime.quarterx",
				"penny.quarter",
				"penny.dime",
			},
			exp: false,
		},
		{
			name:    "with wildcard: matches only entry",
			reqAttr: "*.dime.quarter",
			accAttrs: []string{
				"penny.dime.quarter",
			},
			exp: true,
		},
		{
			name:    "with wildcard: matches first entry",
			reqAttr: "*.dime.quarter",
			accAttrs: []string{
				"penny.dime.quarter",
				"dime.quarter",
				"penny.xdime.quarter",
				"penny.dimex.quarter",
				"penny.dime.xquarter",
				"penny.dime.quarterx",
				"penny.quarter",
				"penny.dime",
			},
			exp: true,
		},
		{
			name:    "with wildcard: matches middle entry",
			reqAttr: "*.dime.quarter",
			accAttrs: []string{
				"dime.quarter",
				"penny.xdime.quarter",
				"penny.dimex.quarter",
				"penny.dime.xquarter",
				"penny.dime.quarter",
				"penny.dime.quarterx",
				"penny.quarter",
				"penny.dime",
			},
			exp: true,
		},
		{
			name:    "with wildcard: matches last entry",
			reqAttr: "*.dime.quarter",
			accAttrs: []string{
				"dime.quarter",
				"penny.xdime.quarter",
				"penny.dimex.quarter",
				"penny.dime.xquarter",
				"penny.dime.quarterx",
				"penny.quarter",
				"penny.dime",
				"penny.dime.quarter",
			},
			exp: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hasMatch := HasReqAttrMatch(tc.reqAttr, tc.accAttrs)
			assert.Equal(t, tc.exp, hasMatch, "HasReqAttrMatch(%q, %q)", tc.reqAttr, tc.accAttrs)
		})
	}
}

func TestIsReqAttrMatch(t *testing.T) {
	tests := []struct {
		name    string
		reqAttr string
		accAttr string
		exp     bool
	}{
		{
			name:    "empty req attr",
			reqAttr: "",
			accAttr: "foo",
			exp:     false,
		},
		{
			name:    "empty acc attr",
			reqAttr: "foo",
			accAttr: "",
			exp:     false,
		},
		{
			name:    "both empty",
			reqAttr: "",
			accAttr: "",
			exp:     false,
		},
		{
			name:    "no wildcard: exact match",
			reqAttr: "penny.dime.quarter",
			accAttr: "penny.dime.quarter",
			exp:     true,
		},
		{
			name:    "no wildcard: opposite order",
			reqAttr: "penny.dime.quarter",
			accAttr: "quarter.dime.penny",
			exp:     false,
		},
		{
			name:    "no wildcard: missing 1st char from 1st name",
			reqAttr: "penny.dime.quarter",
			accAttr: "enny.dime.quarter",
			exp:     false,
		},
		{
			name:    "no wildcard: missing last char from 1st name",
			reqAttr: "penny.dime.quarter",
			accAttr: "penn.dime.quarter",
			exp:     false,
		},
		{
			name:    "no wildcard: missing 1st char from middle name",
			reqAttr: "penny.dime.quarter",
			accAttr: "penny.ime.quarter",
			exp:     false,
		},
		{
			name:    "no wildcard: missing last char from middle name",
			reqAttr: "penny.dime.quarter",
			accAttr: "penny.dim.quarter",
			exp:     false,
		},
		{
			name:    "no wildcard: missing 1st char from last name",
			reqAttr: "penny.dime.quarter",
			accAttr: "penny.dime.uarter",
			exp:     false,
		},
		{
			name:    "no wildcard: missing last char from last name",
			reqAttr: "penny.dime.quarter",
			accAttr: "penny.dime.quarte",
			exp:     false,
		},
		{
			name:    "no wildcard: extra char at start of first name",
			reqAttr: "penny.dime.quarter",
			accAttr: "xpenny.dime.quarter",
			exp:     false,
		},
		{
			name:    "no wildcard: extra char at end of first name",
			reqAttr: "penny.dime.quarter",
			accAttr: "pennyx.dime.quarter",
			exp:     false,
		},
		{
			name:    "no wildcard: extra char at start of middle name",
			reqAttr: "penny.dime.quarter",
			accAttr: "penny.xdime.quarter",
			exp:     false,
		},
		{
			name:    "no wildcard: extra char at end of middle name",
			reqAttr: "penny.dime.quarter",
			accAttr: "penny.dimex.quarter",
			exp:     false,
		},
		{
			name:    "no wildcard: extra char at start of last name",
			reqAttr: "penny.dime.quarter",
			accAttr: "penny.dime.xquarter",
			exp:     false,
		},
		{
			name:    "no wildcard: extra char at end of first name",
			reqAttr: "penny.dime.quarter",
			accAttr: "penny.dime.quarterx",
			exp:     false,
		},
		{
			name:    "no wildcard: extra name at start",
			reqAttr: "penny.dime.quarter",
			accAttr: "mil.penny.dime.quarter",
			exp:     false,
		},
		{
			name:    "no wildcard: extra name at end",
			reqAttr: "penny.dime.quarter",
			accAttr: "penny.dime.quarter.dollar",
			exp:     false,
		},
		{
			name:    "with wildcard: missing 1st char from 1st name",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "enny.dime.quarter",
			exp:     false,
		},
		{
			name:    "with wildcard: missing last char from 1st name",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "penn.dime.quarter",
			exp:     false,
		},
		{
			name:    "with wildcard: missing 1st char from middle name",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "penny.ime.quarter",
			exp:     false,
		},
		{
			name:    "with wildcard: missing last char from middle name",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "penny.dim.quarter",
			exp:     false,
		},
		{
			name:    "with wildcard: missing 1st char from last name",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "penny.dime.uarter",
			exp:     false,
		},
		{
			name:    "with wildcard: missing last char from last name",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "penny.dime.quarte",
			exp:     false,
		},
		{
			name:    "with wildcard: extra char at start of first name",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "xpenny.dime.quarter",
			exp:     false,
		},
		{
			name:    "with wildcard: extra char at end of first name",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "pennyx.dime.quarter",
			exp:     false,
		},
		{
			name:    "with wildcard: extra char at start of middle name",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "penny.xdime.quarter",
			exp:     false,
		},
		{
			name:    "with wildcard: extra char at end of middle name",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "penny.dimex.quarter",
			exp:     false,
		},
		{
			name:    "with wildcard: extra char at start of last name",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "penny.dime.xquarter",
			exp:     false,
		},
		{
			name:    "with wildcard: extra char at end of first name",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "penny.dime.quarterx",
			exp:     false,
		},
		{
			name:    "with wildcard: extra name at start",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "mil.penny.dime.quarter",
			exp:     true,
		},
		{
			name:    "with wildcard: two extra names at start",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "scraps.mil.penny.dime.quarter",
			exp:     true,
		},
		{
			name:    "with wildcard: extra name at start but wrong 1st req name",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "mil.xpenny.dime.quarter",
			exp:     false,
		},
		{
			name:    "with wildcard: only base name",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "penny.dime.quarter",
			exp:     false,
		},
		{
			name:    "with wildcard: extra name at end",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "penny.dime.quarter.dollar",
			exp:     false,
		},
		{
			name:    "with wildcard: extra name at start but wrong base order",
			reqAttr: "*.penny.dime.quarter",
			accAttr: "dollar.quarter.dime.penny",
			exp:     false,
		},
		{
			name:    "star in middle",
			reqAttr: "penny.*.quarter",
			accAttr: "penny.dime.quarter",
			exp:     false,
		},
		{
			name:    "just a star: empty account attribute",
			reqAttr: "*",
			accAttr: "",
			exp:     false,
		},
		{
			name:    "just wildcard: empty account attribute",
			reqAttr: "*.",
			accAttr: "",
			exp:     false,
		},
		{
			name:    "just a star: account attribute has value",
			reqAttr: "*",
			accAttr: "penny.dime.quarter",
			exp:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isMatch := IsReqAttrMatch(tc.reqAttr, tc.accAttr)
			assert.Equal(t, tc.exp, isMatch, "IsReqAttrMatch(%q, %q)", tc.reqAttr, tc.accAttr)
		})
	}
}
