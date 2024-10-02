package types

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"

	"github.com/provenance-io/provenance/testutil/assertions"
)

// emptySdkContext creates a new sdk.Context that only has an empty background context.
func emptySdkContext() sdk.Context {
	return sdk.Context{}.WithContext(context.Background())
}

func TestWrapRequiredParty(t *testing.T) {
	addr := sdk.AccAddress("just_a_test_address_").String()
	tests := []struct {
		name  string
		party Party
		exp   *PartyDetails
	}{
		{
			name: "control",
			party: Party{
				Address:  addr,
				Role:     PartyType_PARTY_TYPE_OWNER,
				Optional: true,
			},
			exp: &PartyDetails{
				address:  addr,
				role:     PartyType_PARTY_TYPE_OWNER,
				optional: true,
			},
		},
		{
			name:  "zero",
			party: Party{},
			exp:   &PartyDetails{},
		},
		{
			name:  "address only",
			party: Party{Address: addr},
			exp:   &PartyDetails{address: addr},
		},
		{
			name:  "role only",
			party: Party{Role: PartyType_PARTY_TYPE_INVESTOR},
			exp:   &PartyDetails{role: PartyType_PARTY_TYPE_INVESTOR},
		},
		{
			name:  "optional only",
			party: Party{Optional: true},
			exp:   &PartyDetails{optional: true},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := WrapRequiredParty(tc.party)
			assert.Equal(t, tc.exp, actual, "WrapRequiredParty")
		})
	}
}

func TestWrapAvailableParty(t *testing.T) {
	addr := sdk.AccAddress("just_a_test_address_").String()
	tests := []struct {
		name  string
		party Party
		exp   *PartyDetails
	}{
		{
			name: "control",
			party: Party{
				Address:  addr,
				Role:     PartyType_PARTY_TYPE_OWNER,
				Optional: true,
			},
			exp: &PartyDetails{
				address:         addr,
				role:            PartyType_PARTY_TYPE_OWNER,
				optional:        true,
				canBeUsedBySpec: true,
			},
		},
		{
			name:  "zero",
			party: Party{},
			exp: &PartyDetails{
				optional:        true,
				canBeUsedBySpec: true,
			},
		},
		{
			name:  "address only",
			party: Party{Address: addr},
			exp: &PartyDetails{
				address:         addr,
				optional:        true,
				canBeUsedBySpec: true,
			},
		},
		{
			name:  "role only",
			party: Party{Role: PartyType_PARTY_TYPE_INVESTOR},
			exp: &PartyDetails{
				role:            PartyType_PARTY_TYPE_INVESTOR,
				optional:        true,
				canBeUsedBySpec: true,
			},
		},
		{
			name:  "optional only",
			party: Party{Optional: true},
			exp: &PartyDetails{
				optional:        true,
				canBeUsedBySpec: true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := WrapAvailableParty(tc.party)
			assert.Equal(t, tc.exp, actual, "WrapAvailableParty")
		})
	}
}

func TestBuildPartyDetails(t *testing.T) {
	addr1 := sdk.AccAddress("this_is_address_1___").String()
	addr2 := sdk.AccAddress("this_is_address_2___").String()
	addr3 := sdk.AccAddress("this_is_address_3___").String()

	// pz is a short way to create a slice of parties.
	pz := func(parties ...Party) []Party {
		rv := make([]Party, 0, len(parties))
		rv = append(rv, parties...)
		return rv
	}
	// dz is a short way to create a slice of PartyDetails
	pdz := func(parties ...*PartyDetails) []*PartyDetails {
		rv := make([]*PartyDetails, 0, len(parties))
		rv = append(rv, parties...)
		return rv
	}
	tests := []struct {
		name             string
		reqParties       []Party
		availableParties []Party
		exp              []*PartyDetails
	}{
		{
			name:             "nil nil",
			reqParties:       nil,
			availableParties: nil,
			exp:              pdz(),
		},
		{
			name:             "nil empty",
			reqParties:       nil,
			availableParties: pz(),
			exp:              pdz(),
		},
		{
			name:             "nil one",
			reqParties:       nil,
			availableParties: pz(Party{Address: addr1, Role: 3, Optional: false}),
			exp: pdz(&PartyDetails{
				address:         addr1,
				role:            3,
				optional:        true,
				canBeUsedBySpec: true,
			}),
		},
		{
			name:             "empty nil",
			reqParties:       pz(),
			availableParties: nil,
			exp:              pdz(),
		},
		{
			name:             "empty empty",
			reqParties:       pz(),
			availableParties: pz(),
			exp:              pdz(),
		},
		{
			name:             "empty one",
			reqParties:       pz(),
			availableParties: pz(Party{Address: addr1, Role: 3, Optional: false}),
			exp: pdz(&PartyDetails{
				address:         addr1,
				role:            3,
				optional:        true,
				canBeUsedBySpec: true,
			}),
		},
		{
			name:             "one nil",
			reqParties:       pz(Party{Address: addr1, Role: 5, Optional: false}),
			availableParties: nil,
			exp: pdz(&PartyDetails{
				address:  addr1,
				role:     5,
				optional: false,
			}),
		},
		{
			name:             "one empty",
			reqParties:       pz(Party{Address: addr1, Role: 5, Optional: false}),
			availableParties: pz(),
			exp: pdz(&PartyDetails{
				address:  addr1,
				role:     5,
				optional: false,
			}),
		},
		{
			name:             "one one different role and address",
			reqParties:       pz(Party{Address: addr1, Role: 5, Optional: false}),
			availableParties: pz(Party{Address: addr2, Role: 4, Optional: false}),
			exp: pdz(
				&PartyDetails{
					address:         addr2,
					role:            4,
					optional:        true,
					canBeUsedBySpec: true,
				},
				&PartyDetails{
					address:  addr1,
					role:     5,
					optional: false,
				},
			),
		},
		{
			name:             "one one different role same address",
			reqParties:       pz(Party{Address: addr1, Role: 5, Optional: false}),
			availableParties: pz(Party{Address: addr1, Role: 4, Optional: false}),
			exp: pdz(
				&PartyDetails{
					address:         addr1,
					role:            4,
					optional:        true,
					canBeUsedBySpec: true,
				},
				&PartyDetails{
					address:  addr1,
					role:     5,
					optional: false,
				},
			),
		},
		{
			name:             "one one different address same role",
			reqParties:       pz(Party{Address: addr1, Role: 5, Optional: false}),
			availableParties: pz(Party{Address: addr2, Role: 5, Optional: false}),
			exp: pdz(
				&PartyDetails{
					address:         addr2,
					role:            5,
					optional:        true,
					canBeUsedBySpec: true,
				},
				&PartyDetails{
					address:  addr1,
					role:     5,
					optional: false,
				},
			),
		},
		{
			name:             "one one same address and role",
			reqParties:       pz(Party{Address: addr1, Role: 5, Optional: false}),
			availableParties: pz(Party{Address: addr1, Role: 5, Optional: true}),
			exp: pdz(&PartyDetails{
				address:         addr1,
				role:            5,
				optional:        false,
				canBeUsedBySpec: true,
			}),
		},
		{
			name: "two two with one same",
			reqParties: pz(
				Party{Address: addr3, Role: 1, Optional: false},
				Party{Address: addr2, Role: 7, Optional: false},
			),
			availableParties: pz(
				Party{Address: addr1, Role: 5, Optional: true},
				Party{Address: addr2, Role: 7, Optional: true},
			),
			exp: pdz(
				&PartyDetails{
					address:         addr1,
					role:            5,
					optional:        true,
					canBeUsedBySpec: true,
				},
				&PartyDetails{
					address:         addr2,
					role:            7,
					optional:        false,
					canBeUsedBySpec: true,
				},
				&PartyDetails{
					address:  addr3,
					role:     1,
					optional: false,
				},
			),
		},
		{
			name: "duplicate req parties",
			reqParties: pz(
				Party{Address: addr1, Role: 2, Optional: false},
				Party{Address: addr1, Role: 2, Optional: false},
			),
			availableParties: nil,
			exp: pdz(&PartyDetails{
				address:  addr1,
				role:     2,
				optional: false,
			}),
		},
		{
			name:       "duplicate available parties",
			reqParties: nil,
			availableParties: pz(
				Party{Address: addr1, Role: 3, Optional: false},
				Party{Address: addr1, Role: 3, Optional: false},
			),
			exp: pdz(&PartyDetails{
				address:         addr1,
				role:            3,
				optional:        true,
				canBeUsedBySpec: true,
			}),
		},
		{
			name: "two req parties one optional",
			reqParties: pz(
				Party{Address: addr1, Role: 2, Optional: false},
				Party{Address: addr2, Role: 3, Optional: true},
			),
			availableParties: nil,
			exp: pdz(&PartyDetails{
				address:  addr1,
				role:     2,
				optional: false,
			}),
		},
		{
			name: "two req parties one optional also in available",
			reqParties: pz(
				Party{Address: addr1, Role: 2, Optional: false},
				Party{Address: addr2, Role: 3, Optional: true},
			),
			availableParties: pz(Party{Address: addr2, Role: 3, Optional: false}),
			exp: pdz(
				&PartyDetails{
					address:         addr2,
					role:            3,
					optional:        true,
					canBeUsedBySpec: true,
				},
				&PartyDetails{
					address:  addr1,
					role:     2,
					optional: false,
				},
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := BuildPartyDetails(tc.reqParties, tc.availableParties)
			assert.Equal(t, tc.exp, actual, "BuildPartyDetails")
		})
	}
}

func TestPartyDetails_Copy(t *testing.T) {
	tests := []struct {
		name string
		pd   *PartyDetails
	}{
		{
			name: "nil",
			pd:   nil,
		},
		{
			name: "empty",
			pd:   &PartyDetails{},
		},
		{
			name: "just address",
			pd:   &PartyDetails{address: "address_value"},
		},
		{
			name: "just role",
			pd:   &PartyDetails{role: PartyType_PARTY_TYPE_OMNIBUS},
		},
		{
			name: "just optional",
			pd:   &PartyDetails{optional: true},
		},
		{
			name: "just acc",
			pd:   &PartyDetails{acc: sdk.AccAddress("acc_value")},
		},
		{
			name: "just signer",
			pd:   &PartyDetails{signer: "signer_value"},
		},
		{
			name: "just signerAcc",
			pd:   &PartyDetails{signerAcc: sdk.AccAddress("signerAcc_value")},
		},
		{
			name: "just canBeUsedBySpec",
			pd:   &PartyDetails{canBeUsedBySpec: true},
		},
		{
			name: "just usedBySpec",
			pd:   &PartyDetails{usedBySpec: true},
		},
		{
			name: "required party",
			pd: WrapRequiredParty(Party{
				Address:  "required_address",
				Role:     PartyType_PARTY_TYPE_CUSTODIAN,
				Optional: false,
			}),
		},
		{
			name: "available party",
			pd: WrapAvailableParty(Party{
				Address:  "available_address",
				Role:     PartyType_PARTY_TYPE_ORIGINATOR,
				Optional: true,
			}),
		},
		{
			name: "everything populated",
			pd: &PartyDetails{
				address:         "another_address",
				role:            PartyType_PARTY_TYPE_AFFILIATE,
				optional:        true,
				acc:             sdk.AccAddress("the_acc_field"),
				signer:          "another_signer",
				signerAcc:       sdk.AccAddress("the_signerAcc_field"),
				canBeUsedBySpec: true,
				usedBySpec:      true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual *PartyDetails
			testFunc := func() {
				actual = tc.pd.Copy()
			}
			require.NotPanics(t, testFunc, "Copy()")
			require.Equal(t, tc.pd, actual, "result of Copy()")
			if tc.pd != nil && actual != nil {
				assert.NotSame(t, tc.pd, actual, "result of Copy()")
				assert.NotSame(t, tc.pd.acc, actual.acc, "acc field")
				if len(actual.acc) > 0 {
					// Change the first byte in the copy to make sure it doesn't also change in the original.
					actual.acc[0] = actual.acc[0] + 1
					assert.NotEqual(t, tc.pd.acc, actual.acc, "the acc field after changing it in the copy")
					// And put it back so we don't mess up anything else.
					actual.acc[0] = actual.acc[0] - 1
				}
				assert.NotSame(t, tc.pd.signerAcc, actual.signerAcc, "signerAcc field")
				if len(actual.signerAcc) > 0 {
					// Change the first byte in the copy to make sure it doesn't also change in the original.
					actual.signerAcc[0] = actual.signerAcc[0] + 1
					assert.NotEqual(t, tc.pd.signerAcc, actual.signerAcc, "the signerAcc field after changing it in the copy")
					actual.signerAcc[0] = actual.signerAcc[0] - 1
				}
			}
		})
	}
}

func TestPartyDetails_SetAddress(t *testing.T) {
	// pd is a short way to create a PartyDetails with only what we care about in this test.
	pd := func(address string, acc sdk.AccAddress) *PartyDetails {
		return &PartyDetails{
			address: address,
			acc:     acc,
		}
	}

	addrAcc := sdk.AccAddress("settable_tst_address")
	addr := addrAcc.String()

	tests := []struct {
		name     string
		party    *PartyDetails
		addr     string
		expParty *PartyDetails
	}{
		{
			name:     "unset to set",
			party:    pd("", nil),
			addr:     addr,
			expParty: pd(addr, nil),
		},
		{
			name:     "set to unset",
			party:    pd(addr, addrAcc),
			addr:     "",
			expParty: pd("", nil),
		},
		{
			name:     "changing to non-acc",
			party:    pd(addr, addrAcc),
			addr:     "new-address",
			expParty: pd("new-address", nil),
		},
		{
			name:     "changing from non-acc",
			party:    pd("not-an-acc", addrAcc),
			addr:     addr,
			expParty: pd(addr, nil),
		},
		{
			name:     "not changing",
			party:    pd(addr, sdk.AccAddress("something else")),
			addr:     addr,
			expParty: pd(addr, sdk.AccAddress("something else")),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.party.SetAddress(tc.addr)
			assert.Equal(t, tc.expParty, tc.party, "party after SetAddress")
		})
	}
}

func TestPartyDetails_GetAddress(t *testing.T) {
	// pd is a short way to create a PartyDetails with only what we care about in this test.
	pd := func(address string, acc sdk.AccAddress) *PartyDetails {
		return &PartyDetails{
			address: address,
			acc:     acc,
		}
	}

	addrAcc := sdk.AccAddress("settable_tst_address")
	addr := addrAcc.String()

	tests := []struct {
		name     string
		party    *PartyDetails
		exp      string
		expParty *PartyDetails
	}{
		{
			name:     "no address no acc",
			party:    pd("", nil),
			exp:      "",
			expParty: pd("", nil),
		},
		{
			name:     "address without acc",
			party:    pd(addr, nil),
			exp:      addr,
			expParty: pd(addr, nil),
		},
		{
			name:     "invalid address without acc",
			party:    pd("invalid", nil),
			exp:      "invalid",
			expParty: pd("invalid", nil),
		},
		{
			name:     "invalid address with acc",
			party:    pd("invalid", addrAcc),
			exp:      "invalid",
			expParty: pd("invalid", addrAcc),
		},
		{
			name:     "acc without address",
			party:    pd("", addrAcc),
			exp:      addr,
			expParty: pd(addr, addrAcc),
		},
		{
			name:     "address with different acc",
			party:    pd(addr, sdk.AccAddress("different_acc_______")),
			exp:      addr,
			expParty: pd(addr, sdk.AccAddress("different_acc_______")),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.party.GetAddress()
			assert.Equal(t, tc.exp, actual, "GetAddress")
			assert.Equal(t, tc.expParty, tc.party, "party after GetAddress")
		})
	}
}

func TestPartyDetails_SetAcc(t *testing.T) {
	// pd is a short way to create a PartyDetails with only what we care about in this test.
	pd := func(address string, acc sdk.AccAddress) *PartyDetails {
		return &PartyDetails{
			address: address,
			acc:     acc,
		}
	}

	addrAcc := sdk.AccAddress("settable_tst_address")
	addr := addrAcc.String()

	tests := []struct {
		name     string
		party    *PartyDetails
		addr     sdk.AccAddress
		expParty *PartyDetails
	}{
		{
			name:     "unset to set",
			party:    pd("", nil),
			addr:     addrAcc,
			expParty: pd("", addrAcc),
		},
		{
			name:     "set to unset",
			party:    pd(addr, addrAcc),
			addr:     nil,
			expParty: pd("", nil),
		},
		{
			name:     "changing no address",
			party:    pd("", addrAcc),
			addr:     sdk.AccAddress("new_address_________"),
			expParty: pd("", sdk.AccAddress("new_address_________")),
		},
		{
			name:     "changing have address",
			party:    pd(addr, addrAcc),
			addr:     sdk.AccAddress("new_address_________"),
			expParty: pd("", sdk.AccAddress("new_address_________")),
		},
		{
			name:     "not changing",
			party:    pd("something else", addrAcc),
			addr:     addrAcc,
			expParty: pd("something else", addrAcc),
		},
		{
			name:     "nil to empty",
			party:    pd("foo", nil),
			addr:     sdk.AccAddress{},
			expParty: pd("foo", sdk.AccAddress{}),
		},
		{
			name:     "empty to nil",
			party:    pd("foo", sdk.AccAddress{}),
			addr:     nil,
			expParty: pd("foo", nil),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.party.SetAcc(tc.addr)
			assert.Equal(t, tc.expParty, tc.party, "party after SetAcc")
		})
	}
}

func TestPartyDetails_GetAcc(t *testing.T) {
	// pd is a short way to create a PartyDetails with only what we care about in this test.
	pd := func(address string, acc sdk.AccAddress) *PartyDetails {
		return &PartyDetails{
			address: address,
			acc:     acc,
		}
	}

	addrAcc := sdk.AccAddress("settable_tst_address")
	addr := addrAcc.String()

	tests := []struct {
		name     string
		party    *PartyDetails
		exp      sdk.AccAddress
		expParty *PartyDetails
	}{
		{
			name:     "no address nil acc",
			party:    pd("", nil),
			exp:      nil,
			expParty: pd("", nil),
		},
		{
			name:     "no address empty acc",
			party:    pd("", sdk.AccAddress{}),
			exp:      sdk.AccAddress{},
			expParty: pd("", sdk.AccAddress{}),
		},
		{
			name:     "address without acc",
			party:    pd(addr, nil),
			exp:      addrAcc,
			expParty: pd(addr, addrAcc),
		},
		{
			name:     "invalid address without acc",
			party:    pd("invalid", nil),
			exp:      nil,
			expParty: pd("invalid", nil),
		},
		{
			name:     "invalid address with acc",
			party:    pd("invalid", addrAcc),
			exp:      addrAcc,
			expParty: pd("invalid", addrAcc),
		},
		{
			name:     "acc without address",
			party:    pd("", addrAcc),
			exp:      addrAcc,
			expParty: pd("", addrAcc),
		},
		{
			name:     "address with different acc",
			party:    pd(addr, sdk.AccAddress("different_acc_______")),
			exp:      sdk.AccAddress("different_acc_______"),
			expParty: pd(addr, sdk.AccAddress("different_acc_______")),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.party.GetAcc()
			assert.Equal(t, tc.exp, actual, "GetAcc")
			assert.Equal(t, tc.expParty, tc.party, "party after GetAcc")
		})
	}
}

func TestPartyDetails_SetRole(t *testing.T) {
	// pd is a short way to create a PartyDetails with only what we care about in this test.
	pd := func(role PartyType) *PartyDetails {
		return &PartyDetails{role: role}
	}

	tests := []struct {
		name     string
		party    *PartyDetails
		role     PartyType
		expParty *PartyDetails
	}{
		{
			name:     "unset to set",
			party:    pd(0),
			role:     1,
			expParty: pd(1),
		},
		{
			name:     "set to unset",
			party:    pd(2),
			role:     0,
			expParty: pd(0),
		},
		{
			name:     "changing",
			party:    pd(3),
			role:     8,
			expParty: pd(8),
		},
		{
			name:     "not changing",
			party:    pd(4),
			role:     4,
			expParty: pd(4),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.party.SetRole(tc.role)
			assert.Equal(t, tc.expParty, tc.party, "party after SetRole")
		})
	}
}

func TestPartyDetails_GetRole(t *testing.T) {
	// pd is a short way to create a PartyDetails with only what we care about in this test.
	pd := func(role PartyType) *PartyDetails {
		return &PartyDetails{role: role}
	}

	type testCase struct {
		name  string
		party *PartyDetails
		exp   PartyType
	}

	var tests []testCase
	for r := range PartyType_name {
		role := PartyType(r)
		tests = append(tests, testCase{
			name:  role.SimpleString(),
			party: pd(role),
			exp:   role,
		})
	}
	sort.Slice(tests, func(i, j int) bool {
		return tests[i].party.GetRole() < tests[j].party.GetRole()
	})
	tests = append(tests, testCase{
		name:  "invalid role",
		party: pd(-8),
		exp:   -8,
	})

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.party.GetRole()
			assert.Equal(t, tc.exp.SimpleString(), actual.SimpleString(), "GetRole")
		})
	}
}

func TestPartyDetails_SetOptional(t *testing.T) {
	// pd is a short way to create a PartyDetails with only what we care about in this test.
	pd := func(optional bool) *PartyDetails {
		return &PartyDetails{optional: optional}
	}

	tests := []struct {
		name     string
		party    *PartyDetails
		optional bool
		expParty *PartyDetails
	}{
		{
			name:     "true to true",
			party:    pd(true),
			optional: true,
			expParty: pd(true),
		},
		{
			name:     "true to false",
			party:    pd(true),
			optional: false,
			expParty: pd(false),
		},
		{
			name:     "false to true",
			party:    pd(false),
			optional: true,
			expParty: pd(true),
		},
		{
			name:     "false to false",
			party:    pd(false),
			optional: false,
			expParty: pd(false),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.party.SetOptional(tc.optional)
			assert.Equal(t, tc.expParty, tc.party, "party after SetOptional")
		})
	}
}

func TestPartyDetails_MakeRequired(t *testing.T) {
	// pd is a short way to create a PartyDetails with only what we care about in this test.
	pd := func(optional bool) *PartyDetails {
		return &PartyDetails{optional: optional}
	}

	tests := []struct {
		name     string
		party    *PartyDetails
		expParty *PartyDetails
	}{
		{
			name:     "from optional",
			party:    pd(true),
			expParty: pd(false),
		},
		{
			name:     "from required",
			party:    pd(false),
			expParty: pd(false),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.party.MakeRequired()
			assert.Equal(t, tc.expParty, tc.party, "party after MakeRequired")
		})
	}
}

func TestPartyDetails_GetOptional(t *testing.T) {
	// pd is a short way to create a PartyDetails with only what we care about in this test.
	pd := func(optional bool) *PartyDetails {
		return &PartyDetails{optional: optional}
	}

	tests := []struct {
		name     string
		party    *PartyDetails
		exp      bool
		expParty *PartyDetails
	}{
		{
			name:     "optional",
			party:    pd(true),
			exp:      true,
			expParty: pd(true),
		},
		{
			name:     "required",
			party:    pd(false),
			exp:      false,
			expParty: pd(false),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.party.GetOptional()
			assert.Equal(t, tc.exp, actual, "GetOptional")
			assert.Equal(t, tc.expParty, tc.party, "party after GetOptional")
		})
	}
}

func TestPartyDetails_IsRequired(t *testing.T) {
	// pd is a short way to create a PartyDetails with only what we care about in this test.
	pd := func(optional bool) *PartyDetails {
		return &PartyDetails{optional: optional}
	}

	tests := []struct {
		name     string
		party    *PartyDetails
		exp      bool
		expParty *PartyDetails
	}{
		{
			name:     "optional",
			party:    pd(true),
			exp:      false,
			expParty: pd(true),
		},
		{
			name:     "required",
			party:    pd(false),
			exp:      true,
			expParty: pd(false),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.party.IsRequired()
			assert.Equal(t, tc.exp, actual, "IsRequired")
			assert.Equal(t, tc.expParty, tc.party, "party after IsRequired")
		})
	}
}

func TestPartyDetails_SetSigner(t *testing.T) {
	// pd is a short way to create a PartyDetails with only what we care about in this test.
	pd := func(signer string, signerAcc sdk.AccAddress) *PartyDetails {
		return &PartyDetails{
			signer:    signer,
			signerAcc: signerAcc,
		}
	}

	addrAcc := sdk.AccAddress("settable_tst_address")
	addr := addrAcc.String()

	tests := []struct {
		name     string
		party    *PartyDetails
		signer   string
		expParty *PartyDetails
	}{
		{
			name:     "unset to set",
			party:    pd("", nil),
			signer:   addr,
			expParty: pd(addr, nil),
		},
		{
			name:     "set to unset",
			party:    pd(addr, addrAcc),
			signer:   "",
			expParty: pd("", nil),
		},
		{
			name:     "changing to non-acc",
			party:    pd(addr, addrAcc),
			signer:   "new-address",
			expParty: pd("new-address", nil),
		},
		{
			name:     "changing from non-acc",
			party:    pd("not-an-acc", addrAcc),
			signer:   addr,
			expParty: pd(addr, nil),
		},
		{
			name:     "not changing",
			party:    pd(addr, sdk.AccAddress("something else")),
			signer:   addr,
			expParty: pd(addr, sdk.AccAddress("something else")),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.party.SetSigner(tc.signer)
			assert.Equal(t, tc.expParty, tc.party, "party after SetSigner")
		})
	}
}

func TestPartyDetails_GetSigner(t *testing.T) {
	// pd is a short way to create a PartyDetails with only what we care about in this test.
	pd := func(signer string, signerAcc sdk.AccAddress) *PartyDetails {
		return &PartyDetails{
			signer:    signer,
			signerAcc: signerAcc,
		}
	}

	addrAcc := sdk.AccAddress("settable_tst_address")
	addr := addrAcc.String()

	tests := []struct {
		name     string
		party    *PartyDetails
		exp      string
		expParty *PartyDetails
	}{
		{
			name:     "no address no acc",
			party:    pd("", nil),
			exp:      "",
			expParty: pd("", nil),
		},
		{
			name:     "address without acc",
			party:    pd(addr, nil),
			exp:      addr,
			expParty: pd(addr, nil),
		},
		{
			name:     "invalid address without acc",
			party:    pd("invalid", nil),
			exp:      "invalid",
			expParty: pd("invalid", nil),
		},
		{
			name:     "invalid address with acc",
			party:    pd("invalid", addrAcc),
			exp:      "invalid",
			expParty: pd("invalid", addrAcc),
		},
		{
			name:     "acc without address",
			party:    pd("", addrAcc),
			exp:      addr,
			expParty: pd(addr, addrAcc),
		},
		{
			name:     "address with different acc",
			party:    pd(addr, sdk.AccAddress("different_acc_______")),
			exp:      addr,
			expParty: pd(addr, sdk.AccAddress("different_acc_______")),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.party.GetSigner()
			assert.Equal(t, tc.exp, actual, "GetSigner")
			assert.Equal(t, tc.expParty, tc.party, "party after GetSigner")
		})
	}
}

func TestPartyDetails_SetSignerAcc(t *testing.T) {
	// pd is a short way to create a PartyDetails with only what we care about in this test.
	pd := func(signer string, signerAcc sdk.AccAddress) *PartyDetails {
		return &PartyDetails{
			signer:    signer,
			signerAcc: signerAcc,
		}
	}

	addrAcc := sdk.AccAddress("settable_tst_address")
	addr := addrAcc.String()

	tests := []struct {
		name     string
		party    *PartyDetails
		signer   sdk.AccAddress
		expParty *PartyDetails
	}{
		{
			name:     "unset to set",
			party:    pd("", nil),
			signer:   addrAcc,
			expParty: pd("", addrAcc),
		},
		{
			name:     "set to unset",
			party:    pd(addr, addrAcc),
			signer:   nil,
			expParty: pd("", nil),
		},
		{
			name:     "changing no address",
			party:    pd("", addrAcc),
			signer:   sdk.AccAddress("new_address_________"),
			expParty: pd("", sdk.AccAddress("new_address_________")),
		},
		{
			name:     "changing have address",
			party:    pd(addr, addrAcc),
			signer:   sdk.AccAddress("new_address_________"),
			expParty: pd("", sdk.AccAddress("new_address_________")),
		},
		{
			name:     "not changing",
			party:    pd("something else", addrAcc),
			signer:   addrAcc,
			expParty: pd("something else", addrAcc),
		},
		{
			name:     "nil to empty",
			party:    pd("foo", nil),
			signer:   sdk.AccAddress{},
			expParty: pd("foo", sdk.AccAddress{}),
		},
		{
			name:     "empty to nil",
			party:    pd("foo", sdk.AccAddress{}),
			signer:   nil,
			expParty: pd("foo", nil),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.party.SetSignerAcc(tc.signer)
			assert.Equal(t, tc.expParty, tc.party, "party after SetSignerAcc")
		})
	}
}

func TestPartyDetails_GetSignerAcc(t *testing.T) {
	// pd is a short way to create a PartyDetails with only what we care about in this test.
	pd := func(signer string, signerAcc sdk.AccAddress) *PartyDetails {
		return &PartyDetails{
			signer:    signer,
			signerAcc: signerAcc,
		}
	}

	addrAcc := sdk.AccAddress("settable_tst_address")
	addr := addrAcc.String()

	tests := []struct {
		name     string
		party    *PartyDetails
		exp      sdk.AccAddress
		expParty *PartyDetails
	}{
		{
			name:     "no address nil acc",
			party:    pd("", nil),
			exp:      nil,
			expParty: pd("", nil),
		},
		{
			name:     "no address empty acc",
			party:    pd("", sdk.AccAddress{}),
			exp:      sdk.AccAddress{},
			expParty: pd("", sdk.AccAddress{}),
		},
		{
			name:     "address without acc",
			party:    pd(addr, nil),
			exp:      addrAcc,
			expParty: pd(addr, addrAcc),
		},
		{
			name:     "invalid address without acc",
			party:    pd("invalid", nil),
			exp:      nil,
			expParty: pd("invalid", nil),
		},
		{
			name:     "invalid address with acc",
			party:    pd("invalid", addrAcc),
			exp:      addrAcc,
			expParty: pd("invalid", addrAcc),
		},
		{
			name:     "acc without address",
			party:    pd("", addrAcc),
			exp:      addrAcc,
			expParty: pd("", addrAcc),
		},
		{
			name:     "address with different acc",
			party:    pd(addr, sdk.AccAddress("different_acc_______")),
			exp:      sdk.AccAddress("different_acc_______"),
			expParty: pd(addr, sdk.AccAddress("different_acc_______")),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.party.GetSignerAcc()
			assert.Equal(t, tc.exp, actual, "GetSignerAcc")
			assert.Equal(t, tc.expParty, tc.party, "party after GetSignerAcc")
		})
	}
}

func TestPartyDetails_HasSigner(t *testing.T) {
	// pd is a short way to create a PartyDetails with only what we care about in this test.
	pd := func(signer string, signerAcc sdk.AccAddress) *PartyDetails {
		return &PartyDetails{
			signer:    signer,
			signerAcc: signerAcc,
		}
	}

	tests := []struct {
		name     string
		party    *PartyDetails
		exp      bool
		expParty *PartyDetails
	}{
		{
			name:     "no string or acc",
			party:    pd("", nil),
			exp:      false,
			expParty: pd("", nil),
		},
		{
			name:     "string no acc",
			party:    pd("a", nil),
			exp:      true,
			expParty: pd("a", nil),
		},
		{
			name:     "acc no string",
			party:    pd("", sdk.AccAddress("b")),
			exp:      true,
			expParty: pd("", sdk.AccAddress("b")),
		},
		{
			name:     "string and acc",
			party:    pd("a", sdk.AccAddress("b")),
			exp:      true,
			expParty: pd("a", sdk.AccAddress("b")),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.party.HasSigner()
			assert.Equal(t, tc.exp, actual, "HasSigner")
			assert.Equal(t, tc.expParty, tc.party, "party after HasSigner")
		})
	}
}

func TestPartyDetails_CanBeUsed(t *testing.T) {
	// pd is a short way to create a PartyDetails with only what we care about in this test.
	pd := func(canBeUsedBySpec bool) *PartyDetails {
		return &PartyDetails{canBeUsedBySpec: canBeUsedBySpec}
	}

	tests := []struct {
		name     string
		party    *PartyDetails
		exp      bool
		expParty *PartyDetails
	}{
		{
			name:     "can be used",
			party:    pd(true),
			exp:      true,
			expParty: pd(true),
		},
		{
			name:     "cannot be used",
			party:    pd(false),
			exp:      false,
			expParty: pd(false),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.party.CanBeUsed()
			assert.Equal(t, tc.exp, actual, "CanBeUsed")
			assert.Equal(t, tc.expParty, tc.party, "party after CanBeUsed")
		})
	}
}

func TestPartyDetails_MarkAsUsed(t *testing.T) {
	// pd is a short way to create a PartyDetails with only what we care about in this test.
	pd := func(usedBySpec bool) *PartyDetails {
		return &PartyDetails{usedBySpec: usedBySpec}
	}

	tests := []struct {
		name     string
		party    *PartyDetails
		expParty *PartyDetails
	}{
		{
			name:     "from not used",
			party:    pd(false),
			expParty: pd(true),
		},
		{
			name:     "from used",
			party:    pd(true),
			expParty: pd(true),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.party.MarkAsUsed()
			assert.Equal(t, tc.expParty, tc.party, "party after MarkAsUsed")
		})
	}
}

func TestPartyDetails_IsUsed(t *testing.T) {
	// pd is a short way to create a PartyDetails with only what we care about in this test.
	pd := func(usedBySpec bool) *PartyDetails {
		return &PartyDetails{usedBySpec: usedBySpec}
	}

	tests := []struct {
		name     string
		party    *PartyDetails
		exp      bool
		expParty *PartyDetails
	}{
		{
			name:     "used",
			party:    pd(true),
			exp:      true,
			expParty: pd(true),
		},
		{
			name:     "not used",
			party:    pd(false),
			exp:      false,
			expParty: pd(false),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.party.IsUsed()
			assert.Equal(t, tc.exp, actual, "IsUsed")
			assert.Equal(t, tc.expParty, tc.party, "party after IsUsed")
		})
	}
}

func TestPartyDetails_IsStillUsableAs(t *testing.T) {
	// pd is a short way to create a PartyDetails with only what we care about in this test.
	pd := func(role PartyType, canBeUsedBySpec, usedBySpec bool) *PartyDetails {
		return &PartyDetails{
			role:            role,
			canBeUsedBySpec: canBeUsedBySpec,
			usedBySpec:      usedBySpec,
		}
	}

	tests := []struct {
		name     string
		party    *PartyDetails
		role     PartyType
		exp      bool
		expParty *PartyDetails
	}{
		{
			name:     "same role can be used is not used",
			party:    pd(1, true, false),
			role:     1,
			exp:      true,
			expParty: pd(1, true, false),
		},
		{
			name:     "same role can be used is used",
			party:    pd(1, true, true),
			role:     1,
			exp:      false,
			expParty: pd(1, true, true),
		},
		{
			name:     "same role cannot be used is not used",
			party:    pd(1, false, false),
			role:     1,
			exp:      false,
			expParty: pd(1, false, false),
		},
		{
			name:     "same role cannot be used is used",
			party:    pd(1, false, true),
			role:     1,
			exp:      false,
			expParty: pd(1, false, true),
		},
		{
			name:     "diff role can be used is not used",
			party:    pd(1, true, false),
			role:     2,
			exp:      false,
			expParty: pd(1, true, false),
		},
		{
			name:     "diff role can be used is used",
			party:    pd(1, true, true),
			role:     2,
			exp:      false,
			expParty: pd(1, true, true),
		},
		{
			name:     "diff role cannot be used is not used",
			party:    pd(1, false, false),
			role:     2,
			exp:      false,
			expParty: pd(1, false, false),
		},
		{
			name:     "diff role cannot be used is used",
			party:    pd(1, false, true),
			role:     2,
			exp:      false,
			expParty: pd(1, false, true),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.party.IsStillUsableAs(tc.role)
			assert.Equal(t, tc.exp, actual, "IsStillUsableAs(%s)", tc.role.SimpleString())
			assert.Equal(t, tc.expParty, tc.party, "party after IsStillUsableAs")
		})
	}
}

func TestPartyDetails_IsSameAs(t *testing.T) {
	tests := []struct {
		name     string
		party    *PartyDetails
		p2       Partier
		exp      bool
		expParty *PartyDetails
	}{
		{
			name: "party details same addr and role all others different",
			party: &PartyDetails{
				address:         "same",
				role:            1,
				optional:        false,
				acc:             sdk.AccAddress("one_________________"),
				signer:          "signer1",
				signerAcc:       sdk.AccAddress("signer1_____________"),
				canBeUsedBySpec: false,
				usedBySpec:      false,
			},
			p2: &PartyDetails{
				address:         "same",
				role:            1,
				optional:        true,
				acc:             sdk.AccAddress("two_________________"),
				signer:          "signer2",
				signerAcc:       sdk.AccAddress("signer2_____________"),
				canBeUsedBySpec: true,
				usedBySpec:      true,
			},
			exp: true,
			expParty: &PartyDetails{
				address:         "same",
				role:            1,
				optional:        false,
				acc:             sdk.AccAddress("one_________________"),
				signer:          "signer1",
				signerAcc:       sdk.AccAddress("signer1_____________"),
				canBeUsedBySpec: false,
				usedBySpec:      false,
			},
		},
		{
			name: "party same addr and role different optional",
			party: &PartyDetails{
				address:  "same",
				role:     1,
				optional: false,
			},
			p2: &Party{
				Address:  "same",
				Role:     1,
				Optional: true,
			},
			exp: true,
			expParty: &PartyDetails{
				address:  "same",
				role:     1,
				optional: false,
			},
		},
		{
			name: "same but only have acc",
			party: &PartyDetails{
				acc:      sdk.AccAddress("same_acc_address____"),
				role:     1,
				optional: false,
			},
			p2: &Party{
				Address:  sdk.AccAddress("same_acc_address____").String(),
				Role:     1,
				Optional: true,
			},
			exp: true,
			expParty: &PartyDetails{
				address:  sdk.AccAddress("same_acc_address____").String(),
				acc:      sdk.AccAddress("same_acc_address____"),
				role:     1,
				optional: false,
			},
		},
		{
			name: "same but both only have acc",
			party: &PartyDetails{
				acc:      sdk.AccAddress("same_acc_address____"),
				role:     1,
				optional: false,
			},
			p2: &PartyDetails{
				acc:      sdk.AccAddress("same_acc_address____"),
				role:     1,
				optional: false,
			},
			exp: true,
			expParty: &PartyDetails{
				address:  sdk.AccAddress("same_acc_address____").String(),
				acc:      sdk.AccAddress("same_acc_address____"),
				role:     1,
				optional: false,
			},
		},
		{
			name: "party details different address",
			party: &PartyDetails{
				address:  "same",
				role:     1,
				optional: false,
			},
			p2: &PartyDetails{
				address:  "not same",
				role:     1,
				optional: true,
			},
			exp: false,
			expParty: &PartyDetails{
				address:  "same",
				role:     1,
				optional: false,
			},
		},
		{
			name: "party details different role",
			party: &PartyDetails{
				address:  "same",
				role:     1,
				optional: false,
			},
			p2: &PartyDetails{
				address:  "same",
				role:     2,
				optional: true,
			},
			exp: false,
			expParty: &PartyDetails{
				address:  "same",
				role:     1,
				optional: false,
			},
		},
		{
			name: "party different address",
			party: &PartyDetails{
				address:  "same",
				role:     1,
				optional: false,
			},
			p2: &Party{
				Address:  "not same",
				Role:     1,
				Optional: true,
			},
			exp: false,
			expParty: &PartyDetails{
				address:  "same",
				role:     1,
				optional: false,
			},
		},
		{
			name: "party different role",
			party: &PartyDetails{
				address:  "same",
				role:     1,
				optional: false,
			},
			p2: &Party{
				Address:  "same",
				Role:     2,
				Optional: true,
			},
			exp: false,
			expParty: &PartyDetails{
				address:  "same",
				role:     1,
				optional: false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.party.IsSameAs(tc.p2)
			assert.Equal(t, tc.exp, actual, "IsSameAs")
			assert.Equal(t, tc.expParty, tc.party, "party after IsSameAs")
		})
	}
}

func TestGetUsedSigners(t *testing.T) {
	addr := func(str string) sdk.AccAddress {
		if len(str) == 0 {
			return nil
		}
		return sdk.AccAddress(str)
	}
	addrStr := func(str string) string {
		if len(str) == 0 {
			return ""
		}
		return addr(str).String()
	}
	pd := func(address, signer, signerAcc string) *PartyDetails {
		return &PartyDetails{
			address:   addrStr(address),
			signer:    addrStr(signer),
			signerAcc: addr(signerAcc),
		}
	}
	pdz := func(parties ...*PartyDetails) []*PartyDetails {
		rv := make([]*PartyDetails, 0, len(parties))
		rv = append(rv, parties...)
		return parties
	}

	tests := []struct {
		name    string
		parties []*PartyDetails
		exp     UsedSignersMap
	}{
		{
			name:    "nil parties",
			parties: nil,
			exp:     map[string]bool{},
		},
		{
			name:    "empty parties",
			parties: pdz(),
			exp:     map[string]bool{},
		},
		{
			name:    "one party no signer",
			parties: pdz(pd("addr1", "", "")),
			exp:     map[string]bool{},
		},
		{
			name:    "one party signer string",
			parties: pdz(pd("addr1", "signer_string", "")),
			exp:     map[string]bool{addrStr("signer_string"): true},
		},
		{
			name:    "one party signer acc",
			parties: pdz(pd("addr1", "", "signer_acc")),
			exp:     map[string]bool{addrStr("signer_acc"): true},
		},
		{
			name:    "one party both signer string and acc",
			parties: pdz(pd("addr1", "signer_string", "signer_acc")),
			exp:     map[string]bool{addrStr("signer_string"): true},
		},
		{
			name:    "two parties neither have signer",
			parties: pdz(pd("addr1", "", ""), pd("addr2", "", "")),
			exp:     map[string]bool{},
		},
		{
			name:    "two parties 1st has signer",
			parties: pdz(pd("addr1", "signer1", ""), pd("addr2", "", "")),
			exp:     map[string]bool{addrStr("signer1"): true},
		},
		{
			name:    "two parties 2nd has signer",
			parties: pdz(pd("addr1", "", ""), pd("addr2", "signer2", "")),
			exp:     map[string]bool{addrStr("signer2"): true},
		},
		{
			name:    "two parties both have different signer",
			parties: pdz(pd("addr1", "signer1", ""), pd("addr2", "signer2", "")),
			exp:     map[string]bool{addrStr("signer1"): true, addrStr("signer2"): true},
		},
		{
			name:    "two parties both have same signer",
			parties: pdz(pd("addr1", "signer1", ""), pd("addr2", "signer1", "")),
			exp:     map[string]bool{addrStr("signer1"): true},
		},
		{
			name: "five parties, 1 without a signer, 1 with signer str, 1 with same signer acc, 2 with unique signers",
			parties: pdz(
				pd("addr1", "signer1", ""),
				pd("addr2", "", ""),
				pd("addr3", "", "signer2"),
				pd("addr4", "", "signer1"),
				pd("addr5", "signer3", ""),
			),
			exp: map[string]bool{
				addrStr("signer1"): true,
				addrStr("signer2"): true,
				addrStr("signer3"): true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := GetUsedSigners(tc.parties)
			assert.Equal(t, tc.exp, actual, "GetAllSigners")
		})
	}
}

func TestNewTestablePartyDetails(t *testing.T) {
	t.Run("nil panics", func(t *testing.T) {
		expPanic := "runtime error: invalid memory address or nil pointer dereference"
		testFunc := func() {
			_ = NewTestablePartyDetails(nil)
		}
		assertions.RequirePanicEquals(t, testFunc, expPanic, "NewTestablePartyDetails")
	})

	t.Run("normal", func(t *testing.T) {
		expected := TestablePartyDetails{
			Address:         "address",
			Role:            10,
			Optional:        true,
			Acc:             sdk.AccAddress("acc"),
			Signer:          "signer",
			SignerAcc:       sdk.AccAddress("signer_acc"),
			CanBeUsedBySpec: true,
			UsedBySpec:      true,
		}
		pd := &PartyDetails{
			address:         "address",
			role:            10,
			optional:        true,
			acc:             sdk.AccAddress("acc"),
			signer:          "signer",
			signerAcc:       sdk.AccAddress("signer_acc"),
			canBeUsedBySpec: true,
			usedBySpec:      true,
		}
		var actual TestablePartyDetails
		testFunc := func() {
			actual = NewTestablePartyDetails(pd)
		}
		require.NotPanics(t, testFunc, "NewTestablePartyDetails")
		assert.Equal(t, expected, actual, "result of NewTestablePartyDetails")
		assert.NotSame(t, pd.acc, actual.Acc, "the acc field")
		actual.Acc[0] = actual.Acc[0] + 1
		assert.NotEqual(t, pd.acc, actual.Acc, "the acc field after a change to it in the result")
		assert.NotSame(t, pd.signerAcc, actual.SignerAcc, "the signerAcc field")
		actual.SignerAcc[0] = actual.SignerAcc[0] + 1
		assert.NotEqual(t, pd.signerAcc, actual.SignerAcc, "the signerAcc field after a change to it in the result")
	})
}

func TestUsedSignersMap(t *testing.T) {
	tests := []struct {
		name     string
		actual   UsedSignersMap
		expected UsedSignersMap
		isUsed   []string
	}{
		{
			name:     "NewUsedSignersMap",
			actual:   NewUsedSignersMap(),
			expected: UsedSignersMap{},
		},
		{
			name:     "Use with two different addrs",
			actual:   NewUsedSignersMap().Use("addr1", "addr2"),
			expected: UsedSignersMap{"addr1": true, "addr2": true},
			isUsed:   []string{"addr1", "addr2"},
		},
		{
			name:     "Use with two same addrs",
			actual:   NewUsedSignersMap().Use("addr", "addr"),
			expected: UsedSignersMap{"addr": true},
			isUsed:   []string{"addr"},
		},
		{
			name:     "Use without any addrs",
			actual:   NewUsedSignersMap().Use(),
			expected: UsedSignersMap{},
		},
		{
			name:     "Use twice different addrs",
			actual:   NewUsedSignersMap().Use("addr1").Use("addr2"),
			expected: UsedSignersMap{"addr1": true, "addr2": true},
			isUsed:   []string{"addr1", "addr2"},
		},
		{
			name:     "Use twice same addr",
			actual:   NewUsedSignersMap().Use("addr").Use("addr"),
			expected: UsedSignersMap{"addr": true},
			isUsed:   []string{"addr"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.actual)
			for _, addr := range tc.isUsed {
				isUsed := tc.actual.IsUsed(addr)
				assert.True(t, isUsed, "IsUsed(%q)", addr)
			}
		})
	}
}

func TestUsedSignersMap_AlsoUse(t *testing.T) {
	tests := []struct {
		name string
		base UsedSignersMap
		m2   UsedSignersMap
		exp  UsedSignersMap
	}{
		{
			name: "two different addrs",
			base: NewUsedSignersMap().Use("addr1"),
			m2:   NewUsedSignersMap().Use("addr2"),
			exp:  UsedSignersMap{"addr1": true, "addr2": true},
		},
		{
			name: "two same addrs",
			base: NewUsedSignersMap().Use("addr"),
			m2:   NewUsedSignersMap().Use("addr"),
			exp:  UsedSignersMap{"addr": true},
		},
		{
			name: "both empty",
			base: NewUsedSignersMap(),
			m2:   NewUsedSignersMap(),
			exp:  UsedSignersMap{},
		},
		{
			name: "base empty",
			base: NewUsedSignersMap(),
			m2:   NewUsedSignersMap().Use("addr"),
			exp:  UsedSignersMap{"addr": true},
		},
		{
			name: "m2 empty",
			base: NewUsedSignersMap().Use("addr"),
			m2:   NewUsedSignersMap(),
			exp:  UsedSignersMap{"addr": true},
		},
		{
			name: "m2 nil",
			base: NewUsedSignersMap().Use("addr"),
			m2:   nil,
			exp:  UsedSignersMap{"addr": true},
		},
		{
			name: "each have 3 with 1 common",
			base: NewUsedSignersMap().Use("addr1", "addr2", "addr3"),
			m2:   NewUsedSignersMap().Use("addr3", "addr4", "addr5"),
			exp: UsedSignersMap{
				"addr1": true,
				"addr2": true,
				"addr3": true,
				"addr4": true,
				"addr5": true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual UsedSignersMap
			testFunc := func() {
				actual = tc.base.AlsoUse(tc.m2)
			}
			require.NotPanics(t, testFunc, "AlsoUse")
			assert.Equal(t, tc.exp, actual, "AlsoUse return value")
			assert.Equal(t, tc.exp, tc.base, "base after AlsoUse")
		})
	}
}

func TestAuthzCacheAcceptableKey(t *testing.T) {
	grantee := sdk.AccAddress("y_grantee_z")
	granter := sdk.AccAddress("Y_GRANTER_Z")
	msgTypeURL := "1_msg_type_url_9"

	firstChar := func(str string) string {
		return str[0:1]
	}
	lastChar := func(str string) string {
		return str[len(str)-1:]
	}

	tests := []struct {
		name     string
		subStr   string
		contains bool
	}{
		{
			name:     "grantee",
			subStr:   string(grantee),
			contains: true,
		},
		{
			name:     "granter",
			subStr:   string(granter),
			contains: true,
		},
		{
			name:     "msgTypeURL",
			subStr:   msgTypeURL,
			contains: true,
		},
		{
			name:     "grantee last granter first",
			subStr:   lastChar(string(grantee)) + firstChar(string(granter)),
			contains: false,
		},
		{
			name:     "granter last grantee first",
			subStr:   lastChar(string(granter)) + firstChar(string(grantee)),
			contains: false,
		},
		{
			name:     "grantee last msgTypeURL first",
			subStr:   lastChar(string(grantee)) + firstChar(msgTypeURL),
			contains: false,
		},
		{
			name:     "msgTypeURL last grantee first",
			subStr:   lastChar(msgTypeURL) + firstChar(string(grantee)),
			contains: false,
		},
		{
			name:     "granter last msgTypeURL first",
			subStr:   lastChar(string(granter)) + firstChar(msgTypeURL),
			contains: false,
		},
		{
			name:     "msgTypeURL last granter first",
			subStr:   lastChar(msgTypeURL) + firstChar(string(granter)),
			contains: false,
		},
	}

	actual := authzCacheAcceptableKey(grantee, granter, msgTypeURL)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.contains {
				assert.Contains(t, actual, tc.subStr, "expected substring of authzCacheAcceptableKey result")
			} else {
				assert.NotContains(t, actual, tc.subStr, "unexpected substring of authzCacheAcceptableKey result")
			}
		})
	}
}

func TestAuthzCacheIsWasmKey(t *testing.T) {
	tests := []struct {
		name string
		str  string
	}{
		{name: "20 char addr", str: "20_character_address"},
		{name: "32 char addr", str: "thirty_two___character___address"},
		{name: "a space", str: " "},
		{name: "empty", str: ""},
		{name: "bytes 0 to 10", str: string([]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10})},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			addr := sdk.AccAddress(tc.str)
			actual := authzCacheIsWasmKey(addr)
			assert.Equal(t, tc.str, actual, "authzCacheIsWasmKey")
		})
	}
}

func TestNewAuthzCache(t *testing.T) {
	c1 := NewAuthzCache()
	c1Type := fmt.Sprintf("%T", c1)
	c2 := NewAuthzCache()

	assert.NotNil(t, c1, "NewAuthzCache result")
	assert.Equal(t, "*types.AuthzCache", c1Type, "type returned by NewAuthzCache")
	assert.Empty(t, c1.acceptable, "acceptable map")
	assert.Empty(t, c1.isWasm, "isWasm map")

	assert.NotSame(t, c1, c2, "NewAuthzCache twice")
	assert.NotSame(t, c1.acceptable, c2.acceptable, "acceptable maps of two NewAuthzCache")
	assert.NotSame(t, c1.isWasm, c2.isWasm, "isWasm maps of two NewAuthzCache")
}

func TestAuthzCache_Clear(t *testing.T) {
	c := NewAuthzCache()
	c.acceptable["key1"] = &authz.CountAuthorization{}
	c.acceptable["key2"] = &authz.GenericAuthorization{}
	c.isWasm["key3"] = true
	c.isWasm["key4"] = false
	assert.NotEmpty(t, c.acceptable, "AuthzCache acceptable map before clear")
	assert.NotEmpty(t, c.isWasm, "AuthzCache isWasm map before clear")
	c.Clear()
	assert.Empty(t, c.acceptable, "AuthzCache acceptable map after clear")
	assert.Empty(t, c.isWasm, "AuthzCache isWasm map after clear")
}

func TestAuthzCache_SetAcceptable(t *testing.T) {
	c := NewAuthzCache()
	grantee := sdk.AccAddress("grantee")
	granter := sdk.AccAddress("granter")
	msgTypeURL := "msgTypeURL"
	authorization := &authz.CountAuthorization{
		Msg:                   msgTypeURL,
		AllowedAuthorizations: 77,
	}

	c.SetAcceptable(grantee, granter, msgTypeURL, authorization)
	actual := c.acceptable[authzCacheAcceptableKey(grantee, granter, msgTypeURL)]
	assert.Equal(t, authorization, actual, "the authorization stored by SetAcceptable")
}

func TestAuthzCache_GetAcceptable(t *testing.T) {
	c := NewAuthzCache()
	grantee := sdk.AccAddress("grantee")
	granter := sdk.AccAddress("granter")
	msgTypeURL := "msgTypeURL"
	key := authzCacheAcceptableKey(grantee, granter, msgTypeURL)

	authorization := &authz.CountAuthorization{
		Msg:                   msgTypeURL,
		AllowedAuthorizations: 8,
	}
	c.acceptable[key] = authorization

	actual := c.GetAcceptable(grantee, granter, msgTypeURL)
	assert.Equal(t, authorization, actual, "GetAcceptable result")

	notThere := c.GetAcceptable(granter, grantee, msgTypeURL)
	assert.Nil(t, notThere, "GetAcceptable on an entry that should not exist")
}

func TestAuthzCache_SetIsWasm(t *testing.T) {
	c := NewAuthzCache()

	// These tests will build on eachother using the same AuthzCache.
	tests := []struct {
		name  string
		addr  sdk.AccAddress
		value bool
		exp   map[string]bool
	}{
		{
			name:  "new entry true",
			addr:  sdk.AccAddress("addr_true"),
			value: true,
			exp:   map[string]bool{"addr_true": true},
		},
		{
			name:  "new entry false",
			addr:  sdk.AccAddress("addr_false"),
			value: false,
			exp:   map[string]bool{"addr_true": true, "addr_false": false},
		},
		{
			name:  "change true to false",
			addr:  sdk.AccAddress("addr_true"),
			value: false,
			exp:   map[string]bool{"addr_true": false, "addr_false": false},
		},
		{
			name:  "change false to true",
			addr:  sdk.AccAddress("addr_false"),
			value: true,
			exp:   map[string]bool{"addr_true": false, "addr_false": true},
		},
		{
			name:  "nil address",
			addr:  nil,
			value: true,
			exp:   map[string]bool{"addr_true": false, "addr_false": true, "": true},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c.SetIsWasm(tc.addr, tc.value)
			m := c.isWasm
			assert.Equal(t, tc.exp, m, "isWasm map after SetIsWasm")
		})
	}
}

func TestAuthzCache_HasIsWasm(t *testing.T) {
	c := NewAuthzCache()
	addrTrue := sdk.AccAddress("addrTrue")
	addrFalse := sdk.AccAddress("addrFalse")
	addrUnknown := sdk.AccAddress("addrUnknown")
	c.SetIsWasm(addrTrue, true)
	c.SetIsWasm(addrFalse, false)

	tests := []struct {
		name string
		addr sdk.AccAddress
		exp  bool
	}{
		{name: "known true", addr: addrTrue, exp: true},
		{name: "known false", addr: addrFalse, exp: true},
		{name: "unknown", addr: addrUnknown, exp: false},
		{name: "nil", addr: nil, exp: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := c.HasIsWasm(tc.addr)
			assert.Equal(t, tc.exp, actual, "HasIsWasm")
		})
	}
}

func TestAuthzCache_GetIsWasm(t *testing.T) {
	c := NewAuthzCache()
	addrTrue := sdk.AccAddress("addrTrue")
	addrFalse := sdk.AccAddress("addrFalse")
	addrUnknown := sdk.AccAddress("addrUnknown")
	c.SetIsWasm(addrTrue, true)
	c.SetIsWasm(addrFalse, false)

	tests := []struct {
		name string
		addr sdk.AccAddress
		exp  bool
	}{
		{name: "known true", addr: addrTrue, exp: true},
		{name: "known false", addr: addrFalse, exp: false},
		{name: "unknown", addr: addrUnknown, exp: false},
		{name: "nil", addr: nil, exp: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := c.GetIsWasm(tc.addr)
			assert.Equal(t, tc.exp, actual, "GetIsWasm")
		})
	}
}

func TestAuthzCache_GetAcceptableMap(t *testing.T) {
	makeCache := func(counts ...int32) *AuthzCache {
		rv := NewAuthzCache()
		for i, count := range counts {
			rv.acceptable[fmt.Sprintf("key_%d__%d", i, count)] = &authz.CountAuthorization{
				Msg:                   fmt.Sprintf("msgTypeURL%d", i),
				AllowedAuthorizations: count,
			}
		}
		return rv
	}

	tests := []struct {
		name  string
		cache *AuthzCache
	}{
		{
			name:  "nil",
			cache: nil,
		},
		{
			name:  "nil map",
			cache: &AuthzCache{},
		},
		{
			name:  "empty map",
			cache: makeCache(),
		},
		{
			name:  "one entry",
			cache: makeCache(5),
		},
		{
			name:  "three entries",
			cache: makeCache(52, 1, 12),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var expected map[string]authz.Authorization
			if tc.cache != nil {
				expected = tc.cache.acceptable
			}
			var actual map[string]authz.Authorization
			testFunc := func() {
				actual = tc.cache.GetAcceptableMap()
			}
			require.NotPanics(t, testFunc, "GetAcceptableMap")
			if expected != nil && actual != nil {
				require.NotSame(t, tc.cache.acceptable, actual, "result from GetAcceptableMap")
			}
			require.Equal(t, expected, actual, "result from GetAcceptableMap")
			if len(actual) > 0 {
				key := ""
				for k := range actual {
					if len(key) == 0 || k < key {
						key = k
					}
				}
				actual[key] = &authz.GenericAuthorization{Msg: "changed"}
				assert.NotEqual(t, tc.cache.acceptable, actual, "after change to result from GetAcceptableMap")
			}
		})
	}
}

func TestAuthzCache_GetIsWasmMap(t *testing.T) {
	makeCache := func(bools ...bool) *AuthzCache {
		rv := NewAuthzCache()
		for i, b := range bools {
			rv.isWasm[fmt.Sprintf("key_%d__%t", i, b)] = b
		}
		return rv
	}

	tests := []struct {
		name  string
		cache *AuthzCache
	}{
		{
			name:  "nil",
			cache: nil,
		},
		{
			name:  "nil map",
			cache: &AuthzCache{},
		},
		{
			name:  "empty map",
			cache: makeCache(),
		},
		{
			name:  "one true entry",
			cache: makeCache(true),
		},
		{
			name:  "one false entry",
			cache: makeCache(false),
		},
		{
			name:  "three entries: true true true",
			cache: makeCache(true, true, true),
		},
		{
			name:  "three entries: true true false",
			cache: makeCache(true, true, false),
		},
		{
			name:  "three entries: true false true",
			cache: makeCache(true, false, true),
		},
		{
			name:  "three entries: false true true",
			cache: makeCache(false, true, true),
		},
		{
			name:  "three entries: true false false",
			cache: makeCache(true, false, false),
		},
		{
			name:  "three entries: false true false",
			cache: makeCache(false, true, false),
		},
		{
			name:  "three entries: false false true",
			cache: makeCache(false, false, true),
		},
		{
			name:  "three entries: false false false",
			cache: makeCache(false, false, false),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var expected map[string]bool
			if tc.cache != nil {
				expected = tc.cache.isWasm
			}
			var actual map[string]bool
			testFunc := func() {
				actual = tc.cache.GetIsWasmMap()
			}
			require.NotPanics(t, testFunc, "GetIsWasmMap")
			if expected != nil && actual != nil {
				require.NotSame(t, tc.cache.isWasm, actual, "result from GetIsWasmMap")
			}
			require.Equal(t, expected, actual, "result from GetIsWasmMap")
			if len(actual) > 0 {
				key := ""
				for k := range actual {
					if len(key) == 0 || k < key {
						key = k
					}
				}
				actual[key] = !actual[key]
				assert.NotEqual(t, tc.cache.isWasm, actual, "after change to result from GetIsWasmMap")
			}
		})
	}
}

func TestAddAuthzCacheToContext(t *testing.T) {
	t.Run("context does not already have the key", func(t *testing.T) {
		origCtx := emptySdkContext()
		newCtx := AddAuthzCacheToContext(origCtx)

		cacheOrig := origCtx.Value(authzCacheContextKey)
		assert.Nil(t, cacheOrig, "original context %q value", authzCacheContextKey)

		cacheV := newCtx.Value(authzCacheContextKey)
		require.NotNil(t, cacheV, "new context %q value", authzCacheContextKey)
		cache, ok := cacheV.(*AuthzCache)
		require.True(t, ok, "can cast %q value to *AuthzCache", authzCacheContextKey)
		require.NotNil(t, cache, "the %q value cast to a *AuthzCache", authzCacheContextKey)
		assert.Empty(t, cache.acceptable, "the acceptable map of the newly added *AuthzCache")
	})

	t.Run("context already has an AuthzCache", func(t *testing.T) {
		grantee := sdk.AccAddress("grantee")
		granter := sdk.AccAddress("granter")
		msgTypeURL := "msgTypeURL"
		authorization := &authz.CountAuthorization{
			Msg:                   msgTypeURL,
			AllowedAuthorizations: 8,
		}
		origCache := NewAuthzCache()
		origCache.SetAcceptable(grantee, granter, msgTypeURL, authorization)

		origCtx := emptySdkContext().WithValue(authzCacheContextKey, origCache)
		newCtx := AddAuthzCacheToContext(origCtx)

		var newCache *AuthzCache
		testFunc := func() {
			newCache = GetAuthzCache(newCtx)
		}
		require.NotPanics(t, testFunc, "GetAuthzCache")
		assert.Same(t, origCache, newCache, "cache from new context")
		assert.Empty(t, newCache.acceptable, "cache acceptable map")
	})

	t.Run("context has something else", func(t *testing.T) {
		origCtx := emptySdkContext().WithValue(authzCacheContextKey, "something else")

		expErr := "context value \"authzCacheContextKey\" is a string, expected *types.AuthzCache"
		testFunc := func() {
			_ = AddAuthzCacheToContext(origCtx)
		}
		require.PanicsWithError(t, expErr, testFunc, "AddAuthzCacheToContext")
	})
}

func TestGetAuthzCache(t *testing.T) {
	t.Run("context does not have it", func(t *testing.T) {
		ctx := emptySdkContext()
		expErr := "context does not contain a \"authzCacheContextKey\" value"
		testFunc := func() {
			_ = GetAuthzCache(ctx)
		}
		require.PanicsWithError(t, expErr, testFunc, "GetAuthzCache")
	})

	t.Run("context has something else", func(t *testing.T) {
		ctx := emptySdkContext().WithValue(authzCacheContextKey, "something else")
		expErr := "context value \"authzCacheContextKey\" is a string, expected *types.AuthzCache"
		testFunc := func() {
			_ = GetAuthzCache(ctx)
		}
		require.PanicsWithError(t, expErr, testFunc, "GetAuthzCache")
	})

	t.Run("context has it", func(t *testing.T) {
		origCache := NewAuthzCache()
		origCache.acceptable["key1"] = &authz.GenericAuthorization{Msg: "msg"}
		ctx := emptySdkContext().WithValue(authzCacheContextKey, origCache)
		var cache *AuthzCache
		testFunc := func() {
			cache = GetAuthzCache(ctx)
		}
		require.NotPanics(t, testFunc, "GetAuthzCache")
		assert.Same(t, origCache, cache, "cache returned by GetAuthzCache")
	})
}
