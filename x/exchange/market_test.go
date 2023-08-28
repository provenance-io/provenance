package exchange

import (
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/maps"
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

func TestPermission_SimpleString(t *testing.T) {
	tests := []struct {
		name string
		p    Permission
		exp  string
	}{
		{
			name: "unspecified",
			p:    Permission_unspecified,
			exp:  "unspecified",
		},
		{
			name: "settle",
			p:    Permission_settle,
			exp:  "settle",
		},
		{
			name: "cancel",
			p:    Permission_cancel,
			exp:  "cancel",
		},
		{
			name: "withdraw",
			p:    Permission_withdraw,
			exp:  "withdraw",
		},
		{
			name: "update",
			p:    Permission_update,
			exp:  "update",
		},
		{
			name: "permissions",
			p:    Permission_permissions,
			exp:  "permissions",
		},
		{
			name: "attributes",
			p:    Permission_attributes,
			exp:  "attributes",
		},
		{
			name: "negative 1",
			p:    -1,
			exp:  "-1",
		},
		{
			name: "unknown value",
			p:    99,
			exp:  "99",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.p.SimpleString()
			assert.Equal(t, tc.exp, actual, "%s.SimpleString()", tc.p)
		})
	}
}

func TestPermission_Validate(t *testing.T) {
	tests := []struct {
		name string
		p    Permission
		exp  string
	}{
		{
			name: "unspecified",
			p:    Permission_unspecified,
			exp:  "permission is unspecified",
		},
		{
			name: "settle",
			p:    Permission_settle,
			exp:  "",
		},
		{
			name: "cancel",
			p:    Permission_cancel,
			exp:  "",
		},
		{
			name: "withdraw",
			p:    Permission_withdraw,
			exp:  "",
		},
		{
			name: "update",
			p:    Permission_update,
			exp:  "",
		},
		{
			name: "permissions",
			p:    Permission_permissions,
			exp:  "",
		},
		{
			name: "attributes",
			p:    Permission_attributes,
			exp:  "",
		},
		{
			name: "negative 1",
			p:    -1,
			exp:  "permission -1 does not exist",
		},
		{
			name: "unknown value",
			p:    99,
			exp:  "permission 99 does not exist",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.p.Validate()

			// TODO: Refactor this to testutils.AssertErrorValue(t, err, tc.exp, "Validate()")
			if len(tc.exp) > 0 {
				assert.EqualError(t, err, tc.exp, "Validate()")
			} else {
				assert.NoError(t, err, "Validate()")
			}
		})
	}

	t.Run("all values have a test case", func(t *testing.T) {
		allVals := maps.Keys(Permission_name)
		sort.Slice(allVals, func(i, j int) bool {
			return allVals[i] < allVals[j]
		})

		for _, val := range allVals {
			perm := Permission(val)
			hasTest := false
			for _, tc := range tests {
				if tc.p == perm {
					hasTest = true
					break
				}
			}
			assert.True(t, hasTest, "No test case found that expects the %s permission", perm)
		}
	})
}

func TestAllPermissions(t *testing.T) {
	expected := []Permission{
		Permission_settle,
		Permission_cancel,
		Permission_withdraw,
		Permission_update,
		Permission_permissions,
		Permission_attributes,
	}

	actual := AllPermissions()
	assert.Equal(t, expected, actual, "AllPermissions()")
}

func TestParsePermission(t *testing.T) {
	tests := []struct {
		permission string
		expected   Permission
		expErr     string
	}{
		// Permission_settle
		{permission: "settle", expected: Permission_settle},
		{permission: " settle", expected: Permission_settle},
		{permission: "settle ", expected: Permission_settle},
		{permission: "SETTLE", expected: Permission_settle},
		{permission: "SeTTle", expected: Permission_settle},
		{permission: "permission_settle", expected: Permission_settle},
		{permission: "PERMISSION_SETTLE", expected: Permission_settle},
		{permission: "pERmiSSion_seTTle", expected: Permission_settle},

		// Permission_cancel
		{permission: "cancel", expected: Permission_cancel},
		{permission: " cancel", expected: Permission_cancel},
		{permission: "cancel ", expected: Permission_cancel},
		{permission: "CANCEL", expected: Permission_cancel},
		{permission: "caNCel", expected: Permission_cancel},
		{permission: "permission_cancel", expected: Permission_cancel},
		{permission: "PERMISSION_CANCEL", expected: Permission_cancel},
		{permission: "pERmiSSion_CanCEl", expected: Permission_cancel},

		// Permission_withdraw
		{permission: "withdraw", expected: Permission_withdraw},
		{permission: " withdraw", expected: Permission_withdraw},
		{permission: "withdraw ", expected: Permission_withdraw},
		{permission: "WITHDRAW", expected: Permission_withdraw},
		{permission: "wiTHdRaw", expected: Permission_withdraw},
		{permission: "permission_withdraw", expected: Permission_withdraw},
		{permission: "PERMISSION_WITHDRAW", expected: Permission_withdraw},
		{permission: "pERmiSSion_wIThdrAw", expected: Permission_withdraw},

		// Permission_update
		{permission: "update", expected: Permission_update},
		{permission: " update", expected: Permission_update},
		{permission: "update ", expected: Permission_update},
		{permission: "UPDATE", expected: Permission_update},
		{permission: "uPDaTe", expected: Permission_update},
		{permission: "permission_update", expected: Permission_update},
		{permission: "PERMISSION_UPDATE", expected: Permission_update},
		{permission: "pERmiSSion_UpdAtE", expected: Permission_update},

		// Permission_permissions
		{permission: "permissions", expected: Permission_permissions},
		{permission: " permissions", expected: Permission_permissions},
		{permission: "permissions ", expected: Permission_permissions},
		{permission: "PERMISSIONS", expected: Permission_permissions},
		{permission: "pErmiSSions", expected: Permission_permissions},
		{permission: "permission_permissions", expected: Permission_permissions},
		{permission: "PERMISSION_PERMISSIONS", expected: Permission_permissions},
		{permission: "pERmiSSion_perMIssIons", expected: Permission_permissions},

		// Permission_attributes
		{permission: "attributes", expected: Permission_attributes},
		{permission: " attributes", expected: Permission_attributes},
		{permission: "attributes ", expected: Permission_attributes},
		{permission: "ATTRIBUTES", expected: Permission_attributes},
		{permission: "aTTribuTes", expected: Permission_attributes},
		{permission: "permission_attributes", expected: Permission_attributes},
		{permission: "PERMISSION_ATTRIBUTES", expected: Permission_attributes},
		{permission: "pERmiSSion_attRiButes", expected: Permission_attributes},

		// Permission_unspecified
		{permission: "unspecified", expErr: `invalid permission: "unspecified"`},
		{permission: " unspecified", expErr: `invalid permission: " unspecified"`},
		{permission: "unspecified ", expErr: `invalid permission: "unspecified "`},
		{permission: "UNSPECIFIED", expErr: `invalid permission: "UNSPECIFIED"`},
		{permission: "unsPeCiFied", expErr: `invalid permission: "unsPeCiFied"`},
		{permission: "permission_unspecified", expErr: `invalid permission: "permission_unspecified"`},
		{permission: "PERMISSION_UNSPECIFIED", expErr: `invalid permission: "PERMISSION_UNSPECIFIED"`},
		{permission: "pERmiSSion_uNSpEcifiEd", expErr: `invalid permission: "pERmiSSion_uNSpEcifiEd"`},

		// Invalid
		{permission: "ettle", expErr: `invalid permission: "ettle"`},
		{permission: "settl", expErr: `invalid permission: "settl"`},
		{permission: "set tle", expErr: `invalid permission: "set tle"`},

		{permission: "ancel", expErr: `invalid permission: "ancel"`},
		{permission: "cance", expErr: `invalid permission: "cance"`},
		{permission: "can cel", expErr: `invalid permission: "can cel"`},

		{permission: "ithdraw", expErr: `invalid permission: "ithdraw"`},
		{permission: "withdra", expErr: `invalid permission: "withdra"`},
		{permission: "with draw", expErr: `invalid permission: "with draw"`},

		{permission: "pdate", expErr: `invalid permission: "pdate"`},
		{permission: "updat", expErr: `invalid permission: "updat"`},
		{permission: "upd ate", expErr: `invalid permission: "upd ate"`},

		{permission: "ermissions", expErr: `invalid permission: "ermissions"`},
		{permission: "permission", expErr: `invalid permission: "permission"`},
		{permission: "permission_permission", expErr: `invalid permission: "permission_permission"`},
		{permission: "permis sions", expErr: `invalid permission: "permis sions"`},

		{permission: "ttributes", expErr: `invalid permission: "ttributes"`},
		{permission: "attribute", expErr: `invalid permission: "attribute"`},
		{permission: "attr ibutes", expErr: `invalid permission: "attr ibutes"`},

		{permission: "", expErr: `invalid permission: ""`},
	}

	for _, tc := range tests {
		name := tc.permission
		if len(tc.permission) == 0 {
			name = "empty"
		}
		t.Run(name, func(t *testing.T) {
			perm, err := ParsePermission(tc.permission)

			if len(tc.expErr) > 0 {
				assert.EqualError(t, err, tc.expErr, "ParsePermission(%q) error", tc.permission)
			} else {
				assert.NoError(t, err, "ParsePermission(%q) error", tc.permission)
			}

			assert.Equal(t, tc.expected, perm, "ParsePermission(%q) result", tc.permission)
		})
	}

	t.Run("all values have a test case", func(t *testing.T) {
		allVals := maps.Keys(Permission_name)
		sort.Slice(allVals, func(i, j int) bool {
			return allVals[i] < allVals[j]
		})

		for _, val := range allVals {
			perm := Permission(val)
			hasTest := false
			for _, tc := range tests {
				if tc.expected == perm {
					hasTest = true
					break
				}
			}
			assert.True(t, hasTest, "No test case found that expects the %s permission", perm)
		}
	})
}

func TestParsePermissions(t *testing.T) {
	tests := []struct {
		name        string
		permissions []string
		expected    []Permission
		expErr      string
	}{
		{
			name:        "nil permissions",
			permissions: nil,
			expected:    nil,
			expErr:      "",
		},
		{
			name:        "empty permissions",
			permissions: nil,
			expected:    nil,
			expErr:      "",
		},
		{
			name:        "one of each permission",
			permissions: []string{"settle", "cancel", "PERMISSION_WITHDRAW", "permission_update", "permissions", "attributes"},
			expected: []Permission{
				Permission_settle,
				Permission_cancel,
				Permission_withdraw,
				Permission_update,
				Permission_permissions,
				Permission_attributes,
			},
		},
		{
			name:        "one bad entry",
			permissions: []string{"settle", "what", "cancel"},
			expected: []Permission{
				Permission_settle,
				Permission_unspecified,
				Permission_cancel,
			},
			expErr: `invalid permission: "what"`,
		},
		{
			name:        "two bad entries",
			permissions: []string{"nope", "withdraw", "notgood"},
			expected: []Permission{
				Permission_unspecified,
				Permission_withdraw,
				Permission_unspecified,
			},
			expErr: `invalid permission: "nope"` + "\n" +
				`invalid permission: "notgood"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			perms, err := ParsePermissions(tc.permissions...)

			// TODO: Refactor to use testutils.AssertErrorValue(t, err, tc.expErr, "ParsePermissions(%q) error", tc.permissions)
			if len(tc.expErr) > 0 {
				assert.EqualError(t, err, tc.expErr, "ParsePermissions(%q) error", tc.permissions)
			} else {
				assert.NoError(t, err, "ParsePermissions(%q) error", tc.permissions)
			}
			assert.Equal(t, tc.expected, perms, "ParsePermissions(%q) result", tc.permissions)
		})
	}
}

func TestValidateReqAttrs(t *testing.T) {
	joinErrs := func(errs ...string) string {
		return strings.Join(errs, "\n")
	}

	tests := []struct {
		name      string
		attrLists [][]string
		exp       string
	}{
		{
			name:      "nil lists",
			attrLists: nil,
			exp:       "",
		},
		{
			name:      "no lists",
			attrLists: [][]string{},
			exp:       "",
		},
		{
			name:      "two empty lists",
			attrLists: [][]string{{}, {}},
			exp:       "",
		},
		{
			name: "one list: three valid entries: normalized",
			attrLists: [][]string{
				{"*.wildcard", "penny.nickel.dime", "*.example.pb"},
			},
			exp: "",
		},
		{
			name: "one list: three valid entries: not normalized",
			attrLists: [][]string{
				{" * . wildcard ", " penny  . nickel .   dime ", " * . example . pb        "},
			},
			exp: "",
		},
		{
			name: "one list: three entries: first invalid",
			attrLists: [][]string{
				{"x.*.wildcard", "penny.nickel.dime", "*.example.pb"},
			},
			exp: `invalid required attribute "x.*.wildcard"`,
		},
		{
			name: "one list: three entries: second invalid",
			attrLists: [][]string{
				{"*.wildcard", "penny.nic kel.dime", "*.example.pb"},
			},
			exp: `invalid required attribute "penny.nic kel.dime"`,
		},
		{
			name: "one list: three entries: third invalid",
			attrLists: [][]string{
				{"*.wildcard", "penny.nickel.dime", "*.ex-am-ple.pb"},
			},
			exp: `invalid required attribute "*.ex-am-ple.pb"`,
		},
		{
			name: "one list: duplicate entries",
			attrLists: [][]string{
				{"*.multi", "*.multi", "*.multi"},
			},
			exp: `duplicate required attribute entry: "*.multi"`,
		},
		{
			name: "one list: duplicate bad entries",
			attrLists: [][]string{
				{"bad.*.example", "bad. * .example"},
			},
			exp: `invalid required attribute "bad.*.example"`,
		},
		{
			name: "one list: multiple problems",
			attrLists: [][]string{
				{
					"one.multi", "x.*.wildcard", "x.*.wildcard", "one.multi", "two.multi",
					"penny.nic kel.dime", "one.multi", "two.multi", "*.ex-am-ple.pb", "two.multi",
				},
			},
			exp: joinErrs(
				`invalid required attribute "x.*.wildcard"`,
				`duplicate required attribute entry: "one.multi"`,
				`invalid required attribute "penny.nic kel.dime"`,
				`duplicate required attribute entry: "two.multi"`,
				`invalid required attribute "*.ex-am-ple.pb"`,
			),
		},
		{
			name: "two lists: second has invalid first",
			attrLists: [][]string{
				{"*.ok", "also.okay.by.me", "this.makes.me.happy"},
				{"x.*.wildcard", "penny.nickel.dime", "*.example.pb"},
			},
			exp: `invalid required attribute "x.*.wildcard"`,
		},
		{
			name: "two lists: second has invalid middle",
			attrLists: [][]string{
				{"*.ok", "also.okay.by.me", "this.makes.me.happy"},
				{"*.wildcard", "penny.nic kel.dime", "*.example.pb"},
			},
			exp: `invalid required attribute "penny.nic kel.dime"`,
		},
		{
			name: "two lists: second has invalid last",
			attrLists: [][]string{
				{"*.ok", "also.okay.by.me", "this.makes.me.happy"},
				{"*.wildcard", "penny.nickel.dime", "*.ex-am-ple.pb"},
			},
			exp: `invalid required attribute "*.ex-am-ple.pb"`,
		},
		{
			name: "two lists: same entry in both but one is not normalized",
			attrLists: [][]string{
				{"this.attr.is.twice"},
				{" This .    Attr . Is . TWice"},
			},
			exp: `duplicate required attribute entry: " This .    Attr . Is . TWice"`,
		},
		{
			name: "two lists: multiple problems",
			attrLists: [][]string{
				{"one.multi", "x.*.wildcard", "x.*.wildcard", "one.multi", "two.multi"},
				{"penny.nic kel.dime", "one.multi", "two.multi", "*.ex-am-ple.pb", "two.multi"},
			},
			exp: joinErrs(
				`invalid required attribute "x.*.wildcard"`,
				`duplicate required attribute entry: "one.multi"`,
				`invalid required attribute "penny.nic kel.dime"`,
				`duplicate required attribute entry: "two.multi"`,
				`invalid required attribute "*.ex-am-ple.pb"`,
			),
		},
		{
			name: "many lists: multiple problems",
			attrLists: [][]string{
				{" one . multi "}, {"x.*.wildcard"}, {"x.*.wildcard"}, {"one.multi"}, {"   two.multi       "},
				{"penny.nic kel.dime"}, {"one.multi"}, {"two.multi"}, {"*.ex-am-ple.pb"}, {"two.multi"},
			},
			exp: joinErrs(
				`invalid required attribute "x.*.wildcard"`,
				`duplicate required attribute entry: "one.multi"`,
				`invalid required attribute "penny.nic kel.dime"`,
				`duplicate required attribute entry: "two.multi"`,
				`invalid required attribute "*.ex-am-ple.pb"`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateReqAttrs(tc.attrLists...)
			// TODO[1658]: Replace this with testutils.AssertErrorValue(t, err, tc.exp, "ValidateReqAttrs")
			if len(tc.exp) > 0 {
				assert.EqualError(t, err, tc.exp, "ValidateReqAttrs")
			} else {
				assert.NoError(t, err, "ValidateReqAttrs")
			}
		})
	}
}

func TestIsValidReqAttr(t *testing.T) {
	tests := []struct {
		name    string
		reqAttr string
		exp     bool
	}{
		{name: "already valid and normalized", reqAttr: "x.y.z", exp: true},
		{name: "already valid but not normalized", reqAttr: " x . y . z ", exp: false},
		{name: "invalid character", reqAttr: "x._y.z", exp: false},
		{name: "just the wildcard", reqAttr: "*", exp: true},
		{name: "just the wildcard not normalized", reqAttr: " * ", exp: false},
		{name: "just star dot", reqAttr: "*.", exp: false},
		{name: "star dot valid", reqAttr: "*.x.y.z", exp: true},
		{name: "star dot valid not normalized", reqAttr: "* . x . y . z", exp: false},
		{name: "star dot invalid", reqAttr: "*.x._y.z", exp: false},
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
