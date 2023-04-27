package keeper_test

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"

	"github.com/provenance-io/provenance/x/metadata/keeper"
	"github.com/provenance-io/provenance/x/metadata/types"
)

// stringSame is a string with an IsSameAs(stringSame) function.
type stringSame string

// IsSameAs satisfies the sameable interface.
func (s stringSame) IsSameAs(c stringSame) bool {
	return string(s) == string(c)
}

// newStringSames converts a slice of strings to a slice of stringEqs.
// nil in => nil out. empty in => empty out.
func newStringSames(strs []string) []stringSame {
	if strs == nil {
		return nil
	}
	rv := make([]stringSame, len(strs), cap(strs))
	for i, str := range strs {
		rv[i] = stringSame(str)
	}
	return rv
}

// stringSameR is a string with an Equals(stringSameC) function that satisfies the sameable interface using
// different types for the receiver and argument.
type stringSameR string

// stringSameC is a string that can be provided to the stringSameR IsSameAs function.
type stringSameC string

// IsSameAs satisfies the sameable interface.
func (s stringSameR) IsSameAs(c stringSameC) bool {
	return string(s) == string(c)
}

// newStringSameRs converts a slice of strings to a slice of stringEqRs.
// nil in => nil out. empty in => empty out.
func newStringSameRs(strs []string) []stringSameR {
	if strs == nil {
		return nil
	}
	rv := make([]stringSameR, len(strs), cap(strs))
	for i, str := range strs {
		rv[i] = stringSameR(str)
	}
	return rv
}

// newStringSameCs converts a slice of strings to a slice of stringEqCs.
// nil in => nil out. empty in => empty out.
func newStringSameCs(strs []string) []stringSameC {
	if strs == nil {
		return nil
	}
	rv := make([]stringSameC, len(strs), cap(strs))
	for i, str := range strs {
		rv[i] = stringSameC(str)
	}
	return rv
}

// partiesCopy creates a new []*keeper.PartyDetails with copies of each provided entry.
// Nil in = nil out.
func partiesCopy(parties []*keeper.PartyDetails) []*keeper.PartyDetails {
	if parties == nil {
		return nil
	}
	rv := make([]*keeper.PartyDetails, len(parties))
	for i, party := range parties {
		rv[i] = party.Copy()
	}
	return rv
}

// partiesReversed creates a new []*keeper.PartyDetails with copies of each provided entry
// in the opposite order as provided. Nil in = nil out.
func partiesReversed(parties []*keeper.PartyDetails) []*keeper.PartyDetails {
	if parties == nil {
		return nil
	}
	rv := make([]*keeper.PartyDetails, len(parties))
	for i, party := range parties {
		rv[len(rv)-i-1] = party.Copy()
	}
	return rv
}

func emptySdkContext() sdk.Context {
	return sdk.Context{}.WithContext(context.Background())
}

func TestWrapRequiredParty(t *testing.T) {
	addr := sdk.AccAddress("just_a_test_address_").String()
	tests := []struct {
		name  string
		party types.Party
		exp   *keeper.PartyDetails
	}{
		{
			name: "control",
			party: types.Party{
				Address:  addr,
				Role:     types.PartyType_PARTY_TYPE_OWNER,
				Optional: true,
			},
			exp: keeper.TestablePartyDetails{
				Address:  addr,
				Role:     types.PartyType_PARTY_TYPE_OWNER,
				Optional: true,
			}.Real(),
		},
		{
			name:  "zero",
			party: types.Party{},
			exp:   keeper.TestablePartyDetails{}.Real(),
		},
		{
			name:  "address only",
			party: types.Party{Address: addr},
			exp:   keeper.TestablePartyDetails{Address: addr}.Real(),
		},
		{
			name:  "role only",
			party: types.Party{Role: types.PartyType_PARTY_TYPE_INVESTOR},
			exp:   keeper.TestablePartyDetails{Role: types.PartyType_PARTY_TYPE_INVESTOR}.Real(),
		},
		{
			name:  "optional only",
			party: types.Party{Optional: true},
			exp:   keeper.TestablePartyDetails{Optional: true}.Real(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := keeper.WrapRequiredParty(tc.party)
			assert.Equal(t, tc.exp, actual, "WrapRequiredParty")
		})
	}
}

func TestWrapAvailableParty(t *testing.T) {
	addr := sdk.AccAddress("just_a_test_address_").String()
	tests := []struct {
		name  string
		party types.Party
		exp   *keeper.PartyDetails
	}{
		{
			name: "control",
			party: types.Party{
				Address:  addr,
				Role:     types.PartyType_PARTY_TYPE_OWNER,
				Optional: true,
			},
			exp: keeper.TestablePartyDetails{
				Address:         addr,
				Role:            types.PartyType_PARTY_TYPE_OWNER,
				Optional:        true,
				CanBeUsedBySpec: true,
			}.Real(),
		},
		{
			name:  "zero",
			party: types.Party{},
			exp: keeper.TestablePartyDetails{
				Optional:        true,
				CanBeUsedBySpec: true,
			}.Real(),
		},
		{
			name:  "address only",
			party: types.Party{Address: addr},
			exp: keeper.TestablePartyDetails{
				Address:         addr,
				Optional:        true,
				CanBeUsedBySpec: true,
			}.Real(),
		},
		{
			name:  "role only",
			party: types.Party{Role: types.PartyType_PARTY_TYPE_INVESTOR},
			exp: keeper.TestablePartyDetails{
				Role:            types.PartyType_PARTY_TYPE_INVESTOR,
				Optional:        true,
				CanBeUsedBySpec: true,
			}.Real(),
		},
		{
			name:  "optional only",
			party: types.Party{Optional: true},
			exp: keeper.TestablePartyDetails{
				Optional:        true,
				CanBeUsedBySpec: true,
			}.Real(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := keeper.WrapAvailableParty(tc.party)
			assert.Equal(t, tc.exp, actual, "WrapAvailableParty")
		})
	}
}

func TestBuildPartyDetails(t *testing.T) {
	addr1 := sdk.AccAddress("this_is_address_1___").String()
	addr2 := sdk.AccAddress("this_is_address_2___").String()
	addr3 := sdk.AccAddress("this_is_address_3___").String()

	// pz is a short way to create a slice of parties.
	pz := func(parties ...types.Party) []types.Party {
		rv := make([]types.Party, 0, len(parties))
		rv = append(rv, parties...)
		return rv
	}
	// dz is a short way to create a slice of PartyDetails
	pdz := func(parties ...*keeper.PartyDetails) []*keeper.PartyDetails {
		rv := make([]*keeper.PartyDetails, 0, len(parties))
		rv = append(rv, parties...)
		return rv
	}
	tests := []struct {
		name             string
		reqParties       []types.Party
		availableParties []types.Party
		exp              []*keeper.PartyDetails
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
			availableParties: pz(types.Party{Address: addr1, Role: 3, Optional: false}),
			exp: pdz(keeper.TestablePartyDetails{
				Address:         addr1,
				Role:            3,
				Optional:        true,
				CanBeUsedBySpec: true,
			}.Real()),
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
			availableParties: pz(types.Party{Address: addr1, Role: 3, Optional: false}),
			exp: pdz(keeper.TestablePartyDetails{
				Address:         addr1,
				Role:            3,
				Optional:        true,
				CanBeUsedBySpec: true,
			}.Real()),
		},
		{
			name:             "one nil",
			reqParties:       pz(types.Party{Address: addr1, Role: 5, Optional: false}),
			availableParties: nil,
			exp: pdz(keeper.TestablePartyDetails{
				Address:  addr1,
				Role:     5,
				Optional: false,
			}.Real()),
		},
		{
			name:             "one empty",
			reqParties:       pz(types.Party{Address: addr1, Role: 5, Optional: false}),
			availableParties: pz(),
			exp: pdz(keeper.TestablePartyDetails{
				Address:  addr1,
				Role:     5,
				Optional: false,
			}.Real()),
		},
		{
			name:             "one one different role and address",
			reqParties:       pz(types.Party{Address: addr1, Role: 5, Optional: false}),
			availableParties: pz(types.Party{Address: addr2, Role: 4, Optional: false}),
			exp: pdz(
				keeper.TestablePartyDetails{
					Address:         addr2,
					Role:            4,
					Optional:        true,
					CanBeUsedBySpec: true,
				}.Real(),
				keeper.TestablePartyDetails{
					Address:  addr1,
					Role:     5,
					Optional: false,
				}.Real(),
			),
		},
		{
			name:             "one one different role same address",
			reqParties:       pz(types.Party{Address: addr1, Role: 5, Optional: false}),
			availableParties: pz(types.Party{Address: addr1, Role: 4, Optional: false}),
			exp: pdz(
				keeper.TestablePartyDetails{
					Address:         addr1,
					Role:            4,
					Optional:        true,
					CanBeUsedBySpec: true,
				}.Real(),
				keeper.TestablePartyDetails{
					Address:  addr1,
					Role:     5,
					Optional: false,
				}.Real(),
			),
		},
		{
			name:             "one one different address same role",
			reqParties:       pz(types.Party{Address: addr1, Role: 5, Optional: false}),
			availableParties: pz(types.Party{Address: addr2, Role: 5, Optional: false}),
			exp: pdz(
				keeper.TestablePartyDetails{
					Address:         addr2,
					Role:            5,
					Optional:        true,
					CanBeUsedBySpec: true,
				}.Real(),
				keeper.TestablePartyDetails{
					Address:  addr1,
					Role:     5,
					Optional: false,
				}.Real(),
			),
		},
		{
			name:             "one one same address and role",
			reqParties:       pz(types.Party{Address: addr1, Role: 5, Optional: false}),
			availableParties: pz(types.Party{Address: addr1, Role: 5, Optional: true}),
			exp: pdz(keeper.TestablePartyDetails{
				Address:         addr1,
				Role:            5,
				Optional:        false,
				CanBeUsedBySpec: true,
			}.Real()),
		},
		{
			name: "two two with one same",
			reqParties: pz(
				types.Party{Address: addr3, Role: 1, Optional: false},
				types.Party{Address: addr2, Role: 7, Optional: false},
			),
			availableParties: pz(
				types.Party{Address: addr1, Role: 5, Optional: true},
				types.Party{Address: addr2, Role: 7, Optional: true},
			),
			exp: pdz(
				keeper.TestablePartyDetails{
					Address:         addr1,
					Role:            5,
					Optional:        true,
					CanBeUsedBySpec: true,
				}.Real(),
				keeper.TestablePartyDetails{
					Address:         addr2,
					Role:            7,
					Optional:        false,
					CanBeUsedBySpec: true,
				}.Real(),
				keeper.TestablePartyDetails{
					Address:  addr3,
					Role:     1,
					Optional: false,
				}.Real(),
			),
		},
		{
			name: "duplicate req parties",
			reqParties: pz(
				types.Party{Address: addr1, Role: 2, Optional: false},
				types.Party{Address: addr1, Role: 2, Optional: false},
			),
			availableParties: nil,
			exp: pdz(keeper.TestablePartyDetails{
				Address:  addr1,
				Role:     2,
				Optional: false,
			}.Real()),
		},
		{
			name:       "duplicate available parties",
			reqParties: nil,
			availableParties: pz(
				types.Party{Address: addr1, Role: 3, Optional: false},
				types.Party{Address: addr1, Role: 3, Optional: false},
			),
			exp: pdz(keeper.TestablePartyDetails{
				Address:         addr1,
				Role:            3,
				Optional:        true,
				CanBeUsedBySpec: true,
			}.Real()),
		},
		{
			name: "two req parties one optional",
			reqParties: pz(
				types.Party{Address: addr1, Role: 2, Optional: false},
				types.Party{Address: addr2, Role: 3, Optional: true},
			),
			availableParties: nil,
			exp: pdz(keeper.TestablePartyDetails{
				Address:  addr1,
				Role:     2,
				Optional: false,
			}.Real()),
		},
		{
			name: "two req parties one optional also in available",
			reqParties: pz(
				types.Party{Address: addr1, Role: 2, Optional: false},
				types.Party{Address: addr2, Role: 3, Optional: true},
			),
			availableParties: pz(types.Party{Address: addr2, Role: 3, Optional: false}),
			exp: pdz(
				keeper.TestablePartyDetails{
					Address:         addr2,
					Role:            3,
					Optional:        true,
					CanBeUsedBySpec: true,
				}.Real(),
				keeper.TestablePartyDetails{
					Address:  addr1,
					Role:     2,
					Optional: false,
				}.Real(),
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := keeper.BuildPartyDetails(tc.reqParties, tc.availableParties)
			assert.Equal(t, tc.exp, actual, "BuildPartyDetails")
		})
	}
}

func TestPartyDetails_SetAddress(t *testing.T) {
	// pd is a short way to create a PartyDetails with only what we care about in this test.
	pd := func(address string, acc sdk.AccAddress) *keeper.PartyDetails {
		return keeper.TestablePartyDetails{
			Address: address,
			Acc:     acc,
		}.Real()
	}

	addrAcc := sdk.AccAddress("settable_tst_address")
	addr := addrAcc.String()

	tests := []struct {
		name     string
		party    *keeper.PartyDetails
		addr     string
		expParty *keeper.PartyDetails
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
	pd := func(address string, acc sdk.AccAddress) *keeper.PartyDetails {
		return keeper.TestablePartyDetails{
			Address: address,
			Acc:     acc,
		}.Real()
	}

	addrAcc := sdk.AccAddress("settable_tst_address")
	addr := addrAcc.String()

	tests := []struct {
		name     string
		party    *keeper.PartyDetails
		exp      string
		expParty *keeper.PartyDetails
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
	pd := func(address string, acc sdk.AccAddress) *keeper.PartyDetails {
		return keeper.TestablePartyDetails{
			Address: address,
			Acc:     acc,
		}.Real()
	}

	addrAcc := sdk.AccAddress("settable_tst_address")
	addr := addrAcc.String()

	tests := []struct {
		name     string
		party    *keeper.PartyDetails
		addr     sdk.AccAddress
		expParty *keeper.PartyDetails
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
	pd := func(address string, acc sdk.AccAddress) *keeper.PartyDetails {
		return keeper.TestablePartyDetails{
			Address: address,
			Acc:     acc,
		}.Real()
	}

	addrAcc := sdk.AccAddress("settable_tst_address")
	addr := addrAcc.String()

	tests := []struct {
		name     string
		party    *keeper.PartyDetails
		exp      sdk.AccAddress
		expParty *keeper.PartyDetails
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
	pd := func(role types.PartyType) *keeper.PartyDetails {
		return keeper.TestablePartyDetails{Role: role}.Real()
	}

	tests := []struct {
		name     string
		party    *keeper.PartyDetails
		role     types.PartyType
		expParty *keeper.PartyDetails
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
	pd := func(role types.PartyType) *keeper.PartyDetails {
		return keeper.TestablePartyDetails{Role: role}.Real()
	}

	type testCase struct {
		name  string
		party *keeper.PartyDetails
		exp   types.PartyType
	}

	var tests []testCase
	for r := range types.PartyType_name {
		role := types.PartyType(r)
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
	pd := func(optional bool) *keeper.PartyDetails {
		return keeper.TestablePartyDetails{Optional: optional}.Real()
	}

	tests := []struct {
		name     string
		party    *keeper.PartyDetails
		optional bool
		expParty *keeper.PartyDetails
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
	pd := func(optional bool) *keeper.PartyDetails {
		return keeper.TestablePartyDetails{Optional: optional}.Real()
	}

	tests := []struct {
		name     string
		party    *keeper.PartyDetails
		expParty *keeper.PartyDetails
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
	pd := func(optional bool) *keeper.PartyDetails {
		return keeper.TestablePartyDetails{Optional: optional}.Real()
	}

	tests := []struct {
		name     string
		party    *keeper.PartyDetails
		exp      bool
		expParty *keeper.PartyDetails
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
	pd := func(optional bool) *keeper.PartyDetails {
		return keeper.TestablePartyDetails{Optional: optional}.Real()
	}

	tests := []struct {
		name     string
		party    *keeper.PartyDetails
		exp      bool
		expParty *keeper.PartyDetails
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
	pd := func(signer string, signerAcc sdk.AccAddress) *keeper.PartyDetails {
		return keeper.TestablePartyDetails{
			Signer:    signer,
			SignerAcc: signerAcc,
		}.Real()
	}

	addrAcc := sdk.AccAddress("settable_tst_address")
	addr := addrAcc.String()

	tests := []struct {
		name     string
		party    *keeper.PartyDetails
		signer   string
		expParty *keeper.PartyDetails
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
	pd := func(signer string, signerAcc sdk.AccAddress) *keeper.PartyDetails {
		return keeper.TestablePartyDetails{
			Signer:    signer,
			SignerAcc: signerAcc,
		}.Real()
	}

	addrAcc := sdk.AccAddress("settable_tst_address")
	addr := addrAcc.String()

	tests := []struct {
		name     string
		party    *keeper.PartyDetails
		exp      string
		expParty *keeper.PartyDetails
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
	pd := func(signer string, signerAcc sdk.AccAddress) *keeper.PartyDetails {
		return keeper.TestablePartyDetails{
			Signer:    signer,
			SignerAcc: signerAcc,
		}.Real()
	}

	addrAcc := sdk.AccAddress("settable_tst_address")
	addr := addrAcc.String()

	tests := []struct {
		name     string
		party    *keeper.PartyDetails
		signer   sdk.AccAddress
		expParty *keeper.PartyDetails
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
	pd := func(signer string, signerAcc sdk.AccAddress) *keeper.PartyDetails {
		return keeper.TestablePartyDetails{
			Signer:    signer,
			SignerAcc: signerAcc,
		}.Real()
	}

	addrAcc := sdk.AccAddress("settable_tst_address")
	addr := addrAcc.String()

	tests := []struct {
		name     string
		party    *keeper.PartyDetails
		exp      sdk.AccAddress
		expParty *keeper.PartyDetails
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
	pd := func(signer string, signerAcc sdk.AccAddress) *keeper.PartyDetails {
		return keeper.TestablePartyDetails{
			Signer:    signer,
			SignerAcc: signerAcc,
		}.Real()
	}

	tests := []struct {
		name     string
		party    *keeper.PartyDetails
		exp      bool
		expParty *keeper.PartyDetails
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
	pd := func(canBeUsedBySpec bool) *keeper.PartyDetails {
		return keeper.TestablePartyDetails{CanBeUsedBySpec: canBeUsedBySpec}.Real()
	}

	tests := []struct {
		name     string
		party    *keeper.PartyDetails
		exp      bool
		expParty *keeper.PartyDetails
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
	pd := func(usedBySpec bool) *keeper.PartyDetails {
		return keeper.TestablePartyDetails{UsedBySpec: usedBySpec}.Real()
	}

	tests := []struct {
		name     string
		party    *keeper.PartyDetails
		expParty *keeper.PartyDetails
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
	pd := func(usedBySpec bool) *keeper.PartyDetails {
		return keeper.TestablePartyDetails{UsedBySpec: usedBySpec}.Real()
	}

	tests := []struct {
		name     string
		party    *keeper.PartyDetails
		exp      bool
		expParty *keeper.PartyDetails
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
	pd := func(role types.PartyType, canBeUsedBySpec, usedBySpec bool) *keeper.PartyDetails {
		return keeper.TestablePartyDetails{
			Role:            role,
			CanBeUsedBySpec: canBeUsedBySpec,
			UsedBySpec:      usedBySpec,
		}.Real()
	}

	tests := []struct {
		name     string
		party    *keeper.PartyDetails
		role     types.PartyType
		exp      bool
		expParty *keeper.PartyDetails
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
		party    *keeper.PartyDetails
		p2       types.Partier
		exp      bool
		expParty *keeper.PartyDetails
	}{
		{
			name: "party details same addr and role all others different",
			party: keeper.TestablePartyDetails{
				Address:         "same",
				Role:            1,
				Optional:        false,
				Acc:             sdk.AccAddress("one_________________"),
				Signer:          "signer1",
				SignerAcc:       sdk.AccAddress("signer1_____________"),
				CanBeUsedBySpec: false,
				UsedBySpec:      false,
			}.Real(),
			p2: keeper.TestablePartyDetails{
				Address:         "same",
				Role:            1,
				Optional:        true,
				Acc:             sdk.AccAddress("two_________________"),
				Signer:          "signer2",
				SignerAcc:       sdk.AccAddress("signer2_____________"),
				CanBeUsedBySpec: true,
				UsedBySpec:      true,
			}.Real(),
			exp: true,
			expParty: keeper.TestablePartyDetails{
				Address:         "same",
				Role:            1,
				Optional:        false,
				Acc:             sdk.AccAddress("one_________________"),
				Signer:          "signer1",
				SignerAcc:       sdk.AccAddress("signer1_____________"),
				CanBeUsedBySpec: false,
				UsedBySpec:      false,
			}.Real(),
		},
		{
			name: "party same addr and role different optional",
			party: keeper.TestablePartyDetails{
				Address:  "same",
				Role:     1,
				Optional: false,
			}.Real(),
			p2: &types.Party{
				Address:  "same",
				Role:     1,
				Optional: true,
			},
			exp: true,
			expParty: keeper.TestablePartyDetails{
				Address:  "same",
				Role:     1,
				Optional: false,
			}.Real(),
		},
		{
			name: "same but only have acc",
			party: keeper.TestablePartyDetails{
				Acc:      sdk.AccAddress("same_acc_address____"),
				Role:     1,
				Optional: false,
			}.Real(),
			p2: &types.Party{
				Address:  sdk.AccAddress("same_acc_address____").String(),
				Role:     1,
				Optional: true,
			},
			exp: true,
			expParty: keeper.TestablePartyDetails{
				Address:  sdk.AccAddress("same_acc_address____").String(),
				Acc:      sdk.AccAddress("same_acc_address____"),
				Role:     1,
				Optional: false,
			}.Real(),
		},
		{
			name: "same but both only have acc",
			party: keeper.TestablePartyDetails{
				Acc:      sdk.AccAddress("same_acc_address____"),
				Role:     1,
				Optional: false,
			}.Real(),
			p2: keeper.TestablePartyDetails{
				Acc:      sdk.AccAddress("same_acc_address____"),
				Role:     1,
				Optional: false,
			}.Real(),
			exp: true,
			expParty: keeper.TestablePartyDetails{
				Address:  sdk.AccAddress("same_acc_address____").String(),
				Acc:      sdk.AccAddress("same_acc_address____"),
				Role:     1,
				Optional: false,
			}.Real(),
		},
		{
			name: "party details different address",
			party: keeper.TestablePartyDetails{
				Address:  "same",
				Role:     1,
				Optional: false,
			}.Real(),
			p2: keeper.TestablePartyDetails{
				Address:  "not same",
				Role:     1,
				Optional: true,
			}.Real(),
			exp: false,
			expParty: keeper.TestablePartyDetails{
				Address:  "same",
				Role:     1,
				Optional: false,
			}.Real(),
		},
		{
			name: "party details different role",
			party: keeper.TestablePartyDetails{
				Address:  "same",
				Role:     1,
				Optional: false,
			}.Real(),
			p2: keeper.TestablePartyDetails{
				Address:  "same",
				Role:     2,
				Optional: true,
			}.Real(),
			exp: false,
			expParty: keeper.TestablePartyDetails{
				Address:  "same",
				Role:     1,
				Optional: false,
			}.Real(),
		},
		{
			name: "party different address",
			party: keeper.TestablePartyDetails{
				Address:  "same",
				Role:     1,
				Optional: false,
			}.Real(),
			p2: &types.Party{
				Address:  "not same",
				Role:     1,
				Optional: true,
			},
			exp: false,
			expParty: keeper.TestablePartyDetails{
				Address:  "same",
				Role:     1,
				Optional: false,
			}.Real(),
		},
		{
			name: "party different role",
			party: keeper.TestablePartyDetails{
				Address:  "same",
				Role:     1,
				Optional: false,
			}.Real(),
			p2: &types.Party{
				Address:  "same",
				Role:     2,
				Optional: true,
			},
			exp: false,
			expParty: keeper.TestablePartyDetails{
				Address:  "same",
				Role:     1,
				Optional: false,
			}.Real(),
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
	pd := func(address, signer, signerAcc string) *keeper.PartyDetails {
		return keeper.TestablePartyDetails{
			Address:   addrStr(address),
			Signer:    addrStr(signer),
			SignerAcc: addr(signerAcc),
		}.Real()
	}
	pdz := func(parties ...*keeper.PartyDetails) []*keeper.PartyDetails {
		rv := make([]*keeper.PartyDetails, 0, len(parties))
		rv = append(rv, parties...)
		return parties
	}

	tests := []struct {
		name    string
		parties []*keeper.PartyDetails
		exp     keeper.UsedSignersMap
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
			actual := keeper.GetUsedSigners(tc.parties)
			assert.Equal(t, tc.exp, actual, "GetAllSigners")
		})
	}
}

func TestSignersWrapper(t *testing.T) {
	addr1Acc := sdk.AccAddress("address_one_________")
	addr2Acc := sdk.AccAddress("address_one_________")
	addr1 := addr1Acc.String()
	addr2 := addr2Acc.String()

	strz := func(strings ...string) []string {
		rv := make([]string, 0, len(strings))
		rv = append(rv, strings...)
		return rv
	}
	accz := func(accs ...sdk.AccAddress) []sdk.AccAddress {
		rv := make([]sdk.AccAddress, 0, len(accs))
		rv = append(rv, accs...)
		return rv
	}

	tests := []struct {
		name       string
		wrapper    *keeper.SignersWrapper
		expStrings []string
		expAccs    []sdk.AccAddress
	}{
		{
			name:       "nil strings",
			wrapper:    keeper.NewSignersWrapper(nil),
			expStrings: nil,
			expAccs:    accz(),
		},
		{
			name:       "empty strings",
			wrapper:    keeper.NewSignersWrapper(strz()),
			expStrings: strz(),
			expAccs:    accz(),
		},
		{
			name:       "two valid address",
			wrapper:    keeper.NewSignersWrapper(strz(addr1, addr2)),
			expStrings: strz(addr1, addr2),
			expAccs:    accz(addr1Acc, addr2Acc),
		},
		{
			name:       "two invalid addresses",
			wrapper:    keeper.NewSignersWrapper(strz("bad1", "bad2")),
			expStrings: strz("bad1", "bad2"),
			expAccs:    accz(),
		},
		{
			name:       "three addresses first invalid",
			wrapper:    keeper.NewSignersWrapper(strz("bad1", addr1, addr2)),
			expStrings: strz("bad1", addr1, addr2),
			expAccs:    accz(addr1Acc, addr2Acc),
		},
		{
			name:       "three addresses second invalid",
			wrapper:    keeper.NewSignersWrapper(strz(addr1, "bad2", addr2)),
			expStrings: strz(addr1, "bad2", addr2),
			expAccs:    accz(addr1Acc, addr2Acc),
		},
		{
			name:       "three addresses third invalid",
			wrapper:    keeper.NewSignersWrapper(strz(addr1, addr2, "bad3")),
			expStrings: strz(addr1, addr2, "bad3"),
			expAccs:    accz(addr1Acc, addr2Acc),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actualStrings := tc.wrapper.Strings()
			assert.Equal(t, tc.expStrings, actualStrings, ".String()")
			actualAccs := tc.wrapper.Accs()
			assert.Equal(t, tc.expAccs, actualAccs, ".Accs()")

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

	actual := keeper.AuthzCacheAcceptableKey(grantee, granter, msgTypeURL)

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

func TestNewAuthzCache(t *testing.T) {
	c1 := keeper.NewAuthzCache()
	c1Type := fmt.Sprintf("%T", c1)
	c2 := keeper.NewAuthzCache()
	assert.NotNil(t, c1, "NewAuthzCache result")
	assert.Empty(t, c1.AcceptableMap(), "acceptable map")
	assert.Equal(t, "*keeper.AuthzCache", c1Type, "type returned by NewAuthzCache")
	assert.NotSame(t, c1, c2, "NewAuthzCache twice")
	assert.NotSame(t, c1.AcceptableMap(), c2.AcceptableMap(), "acceptable maps of two NewAuthzCache")
}

func TestAuthzCache_Clear(t *testing.T) {
	c := keeper.NewAuthzCache()
	c.AcceptableMap()["key1"] = &authz.CountAuthorization{}
	c.AcceptableMap()["key2"] = &authz.GenericAuthorization{}
	assert.NotEmpty(t, c.AcceptableMap(), "AuthzCache acceptable map before clear")
	c.Clear()
	assert.Empty(t, c.AcceptableMap(), "AuthzCache acceptable map after clear")
}

func TestAuthzCache_SetAcceptable(t *testing.T) {
	c := keeper.NewAuthzCache()
	grantee := sdk.AccAddress("grantee")
	granter := sdk.AccAddress("granter")
	msgTypeURL := "msgTypeURL"
	authorization := &authz.CountAuthorization{
		Msg:                   msgTypeURL,
		AllowedAuthorizations: 77,
	}

	c.SetAcceptable(grantee, granter, msgTypeURL, authorization)
	actual := c.AcceptableMap()[keeper.AuthzCacheAcceptableKey(grantee, granter, msgTypeURL)]
	assert.Equal(t, authorization, actual, "the authorization stored by SetAcceptable")
}

func TestAuthzCache_GetAcceptable(t *testing.T) {
	c := keeper.NewAuthzCache()
	grantee := sdk.AccAddress("grantee")
	granter := sdk.AccAddress("granter")
	msgTypeURL := "msgTypeURL"
	key := keeper.AuthzCacheAcceptableKey(grantee, granter, msgTypeURL)
	authorization := &authz.CountAuthorization{
		Msg:                   msgTypeURL,
		AllowedAuthorizations: 8,
	}
	c.AcceptableMap()[key] = authorization

	actual := c.GetAcceptable(grantee, granter, msgTypeURL)
	assert.Equal(t, authorization, actual, "GetAcceptable result")

	notThere := c.GetAcceptable(granter, grantee, msgTypeURL)
	assert.Nil(t, notThere, "GetAcceptable on an entry that should not exist")
}

func TestAddAuthzCacheToContext(t *testing.T) {
	t.Run("context does not already have the key", func(t *testing.T) {
		origCtx := emptySdkContext()
		newCtx := keeper.AddAuthzCacheToContext(origCtx)

		cacheOrig := origCtx.Value(keeper.AuthzCacheContextKey)
		assert.Nil(t, cacheOrig, "original context %q value", keeper.AuthzCacheContextKey)

		cacheV := newCtx.Value(keeper.AuthzCacheContextKey)
		require.NotNil(t, cacheV, "new context %q value", keeper.AuthzCacheContextKey)
		cache, ok := cacheV.(*keeper.AuthzCache)
		require.True(t, ok, "can cast %q value to *keeper.AuthzCache", keeper.AuthzCacheContextKey)
		require.NotNil(t, cache, "the %q value cast to a *keeper.AuthzCache", keeper.AuthzCacheContextKey)
		assert.Empty(t, cache.AcceptableMap(), "the acceptable map of the newly added *keeper.AuthzCache")
	})

	t.Run("context already has an AuthzCache", func(t *testing.T) {
		grantee := sdk.AccAddress("grantee")
		granter := sdk.AccAddress("granter")
		msgTypeURL := "msgTypeURL"
		authorization := &authz.CountAuthorization{
			Msg:                   msgTypeURL,
			AllowedAuthorizations: 8,
		}
		origCache := keeper.NewAuthzCache()
		origCache.SetAcceptable(grantee, granter, msgTypeURL, authorization)

		origCtx := emptySdkContext().WithValue(keeper.AuthzCacheContextKey, origCache)
		newCtx := keeper.AddAuthzCacheToContext(origCtx)

		var newCache *keeper.AuthzCache
		testFunc := func() {
			newCache = keeper.GetAuthzCache(newCtx)
		}
		require.NotPanics(t, testFunc, "GetAuthzCache")
		assert.Same(t, origCache, newCache, "cache from new context")
		assert.Empty(t, newCache.AcceptableMap(), "cache acceptable map")
	})

	t.Run("context has something else", func(t *testing.T) {
		origCtx := emptySdkContext().WithValue(keeper.AuthzCacheContextKey, "something else")

		expErr := "context value \"authzCacheContextKey\" is a string, expected *keeper.AuthzCache"
		testFunc := func() {
			_ = keeper.AddAuthzCacheToContext(origCtx)
		}
		require.PanicsWithError(t, expErr, testFunc, "AddAuthzCacheToContext")
	})
}

func TestGetAuthzCache(t *testing.T) {
	t.Run("context does not have it", func(t *testing.T) {
		ctx := emptySdkContext()
		expErr := "context does not contain a \"authzCacheContextKey\" value"
		testFunc := func() {
			_ = keeper.GetAuthzCache(ctx)
		}
		require.PanicsWithError(t, expErr, testFunc, "GetAuthzCache")
	})

	t.Run("context has something else", func(t *testing.T) {
		ctx := emptySdkContext().WithValue(keeper.AuthzCacheContextKey, "something else")
		expErr := "context value \"authzCacheContextKey\" is a string, expected *keeper.AuthzCache"
		testFunc := func() {
			_ = keeper.GetAuthzCache(ctx)
		}
		require.PanicsWithError(t, expErr, testFunc, "GetAuthzCache")
	})

	t.Run("context has it", func(t *testing.T) {
		origCache := keeper.NewAuthzCache()
		origCache.AcceptableMap()["key1"] = &authz.GenericAuthorization{Msg: "msg"}
		ctx := emptySdkContext().WithValue(keeper.AuthzCacheContextKey, origCache)
		var cache *keeper.AuthzCache
		testFunc := func() {
			cache = keeper.GetAuthzCache(ctx)
		}
		require.NotPanics(t, testFunc, "GetAuthzCache")
		assert.Same(t, origCache, cache, "cache returned by GetAuthzCache")
	})
}

func TestUnwrapMetadataContext(t *testing.T) {
	origCtx := emptySdkContext()
	goCtx := sdk.WrapSDKContext(origCtx)
	var ctx sdk.Context
	testUnwrap := func() {
		ctx = keeper.UnwrapMetadataContext(goCtx)
	}
	require.NotPanics(t, testUnwrap, "UnwrapMetadataContext")
	var cache *keeper.AuthzCache
	testGet := func() {
		cache = keeper.GetAuthzCache(ctx)
	}
	require.NotPanics(t, testGet, "GetAuthzCache")
	assert.NotNil(t, cache, "cache returned by GetAuthzCache")
	assert.Empty(t, cache.AcceptableMap(), "cache acceptable map")
}

func TestUsedSignersMap_AlsoUse(t *testing.T) {
	// TODO[1329]: Write TestUsedSignersMap_AlsoUse
	t.Fatalf("not yet written")
}

type testCaseFindMissing struct {
	name     string
	required []string
	toCheck  []string
	expected []string
}

func testCasesForFindMissing() []testCaseFindMissing {
	return []testCaseFindMissing{
		{
			name:     "nil required - nil toCheck - nil out",
			required: nil,
			toCheck:  nil,
			expected: nil,
		},
		{
			name:     "empty required - nil toCheck - nil out",
			required: []string{},
			toCheck:  nil,
			expected: nil,
		},
		{
			name:     "nil required - empty toCheck - nil out",
			required: nil,
			toCheck:  []string{},
			expected: nil,
		},
		{
			name:     "empty required - empty toCheck - nil out",
			required: []string{},
			toCheck:  []string{},
			expected: nil,
		},
		{
			name:     "nil required - 2 toCheck - nil out",
			required: nil,
			toCheck:  []string{"one", "two"},
			expected: nil,
		},
		{
			name:     "empty required - 2 toCheck - nil out",
			required: []string{},
			toCheck:  []string{"one", "two"},
			expected: nil,
		},
		{
			name:     "1 required - is only toCheck - nil out",
			required: []string{"one"},
			toCheck:  []string{"one"},
			expected: nil,
		},
		{
			name:     "1 required - is 1st of 2 toCheck - nil out",
			required: []string{"one"},
			toCheck:  []string{"one", "two"},
			expected: nil,
		},
		{
			name:     "1 required - is 2nd of 2 toCheck - nil out",
			required: []string{"one"},
			toCheck:  []string{"two", "one"},
			expected: nil,
		},
		{
			name:     "1 required -  nil toCheck - required out",
			required: []string{"one"},
			toCheck:  nil,
			expected: []string{"one"},
		},
		{
			name:     "1 required - empty toCheck - required out",
			required: []string{"one"},
			toCheck:  []string{},
			expected: []string{"one"},
		},
		{
			name:     "1 required - 1 other in toCheck - required out",
			required: []string{"one"},
			toCheck:  []string{"two"},
			expected: []string{"one"},
		},
		{
			name:     "1 required - 2 other in toCheck - required out",
			required: []string{"one"},
			toCheck:  []string{"two", "three"},
			expected: []string{"one"},
		},
		{
			name:     "2 required - both in toCheck - nil out",
			required: []string{"one", "two"},
			toCheck:  []string{"one", "two"},
			expected: nil,
		},
		{
			name:     "2 required - reversed in toCheck - nil out",
			required: []string{"one", "two"},
			toCheck:  []string{"two", "one"},
			expected: nil,
		},
		{
			name:     "2 required - only 1st in toCheck - 2nd out",
			required: []string{"one", "two"},
			toCheck:  []string{"one"},
			expected: []string{"two"},
		},
		{
			name:     "2 required - only 2nd in toCheck - 1st out",
			required: []string{"one", "two"},
			toCheck:  []string{"two"},
			expected: []string{"one"},
		},
		{
			name:     "2 required - 1st and other in toCheck - 2nd out",
			required: []string{"one", "two"},
			toCheck:  []string{"one", "other"},
			expected: []string{"two"},
		},
		{
			name:     "2 required - 2nd and other in toCheck - 1st out",
			required: []string{"one", "two"},
			toCheck:  []string{"two", "other"},
			expected: []string{"one"},
		},
		{
			name:     "2 required - nil toCheck - required out",
			required: []string{"one", "two"},
			toCheck:  nil,
			expected: []string{"one", "two"},
		},
		{
			name:     "2 required - empty toCheck - required out",
			required: []string{"one", "two"},
			toCheck:  []string{},
			expected: []string{"one", "two"},
		},
		{
			name:     "2 required - neither in 1 toCheck - required out",
			required: []string{"one", "two"},
			toCheck:  []string{"neither"},
			expected: []string{"one", "two"},
		},
		{
			name:     "2 required - neither in 3 toCheck - required out",
			required: []string{"one", "two"},
			toCheck:  []string{"neither", "nor", "nothing"},
			expected: []string{"one", "two"},
		},
		{
			name:     "2 required - 1st not in 3 toCheck 2nd at 0 - 1st out",
			required: []string{"one", "two"},
			toCheck:  []string{"two", "nor", "nothing"},
			expected: []string{"one"},
		},
		{
			name:     "2 required - 1st not in 3 toCheck 2nd at 1 - 1st out",
			required: []string{"one", "two"},
			toCheck:  []string{"neither", "two", "nothing"},
			expected: []string{"one"},
		},
		{
			name:     "2 required - 1s5 not in 3 toCheck 2nd at 2 - 1st out",
			required: []string{"one", "two"},
			toCheck:  []string{"neither", "nor", "two"},
			expected: []string{"one"},
		},
		{
			name:     "2 required - 2nd not in 3 toCheck 1st at 0 - 2nd out",
			required: []string{"one", "two"},
			toCheck:  []string{"one", "nor", "nothing"},
			expected: []string{"two"},
		},
		{
			name:     "2 required - 2nd not in 3 toCheck 1st at 1 - 2nd out",
			required: []string{"one", "two"},
			toCheck:  []string{"neither", "one", "nothing"},
			expected: []string{"two"},
		},
		{
			name:     "2 required - 2nd not in 3 toCheck 1st at 2 - 2nd out",
			required: []string{"one", "two"},
			toCheck:  []string{"neither", "nor", "one"},
			expected: []string{"two"},
		},

		{
			name:     "3 required - none in 5 toCheck - required out",
			required: []string{"one", "two", "three"},
			toCheck:  []string{"other1", "other2", "other3", "other4", "other5"},
			expected: []string{"one", "two", "three"},
		},
		{
			name:     "3 required - only 1st in 5 toCheck - 2nd 3rd out",
			required: []string{"one", "two", "three"},
			toCheck:  []string{"other1", "other2", "one", "other4", "other5"},
			expected: []string{"two", "three"},
		},
		{
			name:     "3 required - only 2nd in 5 toCheck - 1st 3rd out",
			required: []string{"one", "two", "three"},
			toCheck:  []string{"other1", "two", "other3", "other4", "other5"},
			expected: []string{"one", "three"},
		},
		{
			name:     "3 required - only 3rd in 5 toCheck - 1st 2nd out",
			required: []string{"one", "two", "three"},
			toCheck:  []string{"other1", "other2", "other3", "three", "other5"},
			expected: []string{"one", "two"},
		},
		{
			name:     "3 required - 1st 2nd in 5 toCheck - 3rd out",
			required: []string{"one", "two", "three"},
			toCheck:  []string{"other1", "two", "other3", "one", "other5"},
			expected: []string{"three"},
		},
		{
			name:     "3 required - 1st 3nd in 5 toCheck - 2nd out",
			required: []string{"one", "two", "three"},
			toCheck:  []string{"three", "other2", "other3", "other4", "one"},
			expected: []string{"two"},
		},
		{
			name:     "3 required - 2nd 3rd in 5 toCheck - 1st out",
			required: []string{"one", "two", "three"},
			toCheck:  []string{"other1", "other2", "two", "three", "other5"},
			expected: []string{"one"},
		},
		{
			name:     "3 required - all in 5 toCheck - nil out",
			required: []string{"one", "two", "three"},
			toCheck:  []string{"two", "other2", "one", "three", "other5"},
			expected: nil,
		},
		{
			name:     "3 required with dup - all in toCheck - nil out",
			required: []string{"one", "two", "one"},
			toCheck:  []string{"one", "two"},
			expected: nil,
		},
		{
			name:     "3 required with dup - dup not in toCheck - dups out",
			required: []string{"one", "two", "one"},
			toCheck:  []string{"two"},
			expected: []string{"one", "one"},
		},
		{
			name:     "3 required with dup - other not in toCheck - other out",
			required: []string{"one", "two", "one"},
			toCheck:  []string{"one"},
			expected: []string{"two"},
		},
		{
			name:     "3 required all dup - in toCheck - nil out",
			required: []string{"one", "one", "one"},
			toCheck:  []string{"one"},
			expected: nil,
		},
		{
			name:     "3 required all dup - not in toCheck - all 3 out",
			required: []string{"one", "one", "one"},
			toCheck:  []string{"two"},
			expected: []string{"one", "one", "one"},
		},
	}
}

func TestFindMissing(t *testing.T) {
	for _, tc := range testCasesForFindMissing() {
		t.Run(tc.name, func(t *testing.T) {
			actual := keeper.FindMissing(tc.required, tc.toCheck)
			assert.Equal(t, tc.expected, actual, "findMissing")
		})
	}
}

func TestFindMissingParties(t *testing.T) {
	// pz is just a shorter way to define a []types.Party
	pz := func(parties ...types.Party) []types.Party {
		return parties
	}

	pOne3Req := types.Party{Address: "one", Role: 3, Optional: false}
	pOne3Opt := types.Party{Address: "one", Role: 3, Optional: true}
	pOne4Req := types.Party{Address: "one", Role: 4, Optional: false}
	pOne4Opt := types.Party{Address: "one", Role: 4, Optional: true}
	pTwo3Req := types.Party{Address: "two", Role: 3, Optional: false}
	pTwo3Opt := types.Party{Address: "two", Role: 3, Optional: true}
	pTwo4Req := types.Party{Address: "two", Role: 4, Optional: false}
	pTwo4Opt := types.Party{Address: "two", Role: 4, Optional: true}

	// Note: types.PartyType_PARTY_TYPE_INVESTOR = 3, types.PartyType_PARTY_TYPE_CUSTODIAN = 4

	tests := []struct {
		name     string
		required []types.Party
		toCheck  []types.Party
		expected []types.Party
	}{
		{
			name:     "nil nil",
			required: nil,
			toCheck:  nil,
			expected: nil,
		},
		{
			name:     "empty nil",
			required: pz(),
			toCheck:  nil,
			expected: nil,
		},
		{
			name:     "nil empty",
			required: nil,
			toCheck:  pz(),
			expected: nil,
		},
		{
			name:     "empty empty",
			required: pz(),
			toCheck:  pz(),
			expected: nil,
		},

		{
			name:     "nil VS one3",
			required: nil,
			toCheck:  pz(pOne3Req),
			expected: nil,
		},
		{
			name:     "empty VS one3",
			required: pz(),
			toCheck:  pz(pOne3Req),
			expected: nil,
		},

		{
			name:     "one3req VS one3req",
			required: pz(pOne3Req),
			toCheck:  pz(pOne3Req),
			expected: nil,
		},
		{
			name:     "one3req VS one3opt",
			required: pz(pOne3Req),
			toCheck:  pz(pOne3Opt),
			expected: nil,
		},
		{
			name:     "one3opt VS one3req",
			required: pz(pOne3Opt),
			toCheck:  pz(pOne3Req),
			expected: nil,
		},
		{
			name:     "one3opt VS one3opt",
			required: pz(pOne3Opt),
			toCheck:  pz(pOne3Opt),
			expected: nil,
		},

		{
			name:     "one3 one4 two3 two4 req VS one4 one3 two4 two3 req",
			required: pz(pOne3Req, pOne4Req, pTwo3Req, pTwo4Req),
			toCheck:  pz(pOne4Req, pOne3Req, pTwo4Req, pTwo3Req),
			expected: nil,
		},
		{
			name:     "one3 one4 two3 two4 req VS one4 one3 two4 two3 opt",
			required: pz(pOne3Req, pOne4Req, pTwo3Req, pTwo4Req),
			toCheck:  pz(pOne4Opt, pOne3Opt, pTwo4Opt, pTwo3Opt),
			expected: nil,
		},
		{
			name:     "one3 one4 two3 two4 opt vs one4 one3 two4 two3 req",
			required: pz(pOne3Opt, pOne4Opt, pTwo3Opt, pTwo4Opt),
			toCheck:  pz(pOne4Req, pOne3Req, pTwo4Req, pTwo3Req),
			expected: nil,
		},
		{
			name:     "one3 one4 two3 two4 opt vs one4 one3 two4 two3 opt",
			required: pz(pOne3Opt, pOne4Opt, pTwo3Opt, pTwo4Opt),
			toCheck:  pz(pOne4Opt, pOne3Opt, pTwo4Opt, pTwo3Opt),
			expected: nil,
		},

		{
			name:     "one3 two4 VS nil",
			required: pz(pOne3Opt, pTwo4Req),
			toCheck:  nil,
			expected: pz(pOne3Opt, pTwo4Req),
		},
		{
			name:     "one3 two4 VS empty",
			required: pz(pOne3Opt, pTwo4Req),
			toCheck:  pz(),
			expected: pz(pOne3Opt, pTwo4Req),
		},
		{
			name:     "one3 two4 VS one3",
			required: pz(pOne3Opt, pTwo4Req),
			toCheck:  pz(pOne3Req),
			expected: pz(pTwo4Req),
		},
		{
			name:     "one3 two4 VS one4",
			required: pz(pOne3Opt, pTwo4Req),
			toCheck:  pz(pOne4Opt),
			expected: pz(pOne3Opt, pTwo4Req),
		},
		{
			name:     "one3 two4 VS two3",
			required: pz(pOne3Opt, pTwo4Req),
			toCheck:  pz(pTwo3Opt),
			expected: pz(pOne3Opt, pTwo4Req),
		},
		{
			name:     "one3 two4 VS two4",
			required: pz(pOne3Opt, pTwo4Req),
			toCheck:  pz(pTwo4Opt),
			expected: pz(pOne3Opt),
		},

		{
			name:     "one3req two4opt VS two4req one3opt",
			required: pz(pOne3Req, pTwo4Opt),
			toCheck:  pz(pTwo4Req, pOne3Opt),
			expected: nil,
		},
		{
			name:     "one3opt two4req VS two4opt one3req",
			required: pz(pOne3Opt, pTwo4Req),
			toCheck:  pz(pTwo4Opt, pOne3Req),
			expected: nil,
		},

		{
			name:     "one3opt VS all others req",
			required: pz(pOne3Opt),
			toCheck:  pz(pOne3Req, pOne4Req, pTwo3Req, pTwo4Req),
			expected: nil,
		},
		{
			name:     "one3req VS all others opt",
			required: pz(pOne3Req),
			toCheck:  pz(pOne3Opt, pOne4Opt, pTwo3Opt, pTwo4Opt),
			expected: nil,
		},
		{
			name:     "all req VS two3Opt",
			required: pz(pOne4Req, pTwo3Req, pOne3Req, pTwo4Req),
			toCheck:  pz(pTwo3Opt),
			expected: pz(pOne4Req, pOne3Req, pTwo4Req),
		},
		{
			name:     "all opt VS two3Req",
			required: pz(pOne4Opt, pOne3Opt, pTwo3Opt, pTwo4Opt),
			toCheck:  pz(pTwo3Req),
			expected: pz(pOne4Opt, pOne3Opt, pTwo4Opt),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := keeper.FindMissingParties(tc.required, tc.toCheck)
			assert.Equal(t, tc.expected, actual, "findMissingParties")
		})
	}
}

func TestFindMissingComp(t *testing.T) {
	t.Run("equals equals", func(t *testing.T) {
		comp := func(r, c string) bool {
			return r == c
		}
		for _, tc := range testCasesForFindMissing() {
			t.Run(tc.name, func(t *testing.T) {
				actual := keeper.FindMissingComp(tc.required, tc.toCheck, comp)
				assert.Equal(t, tc.expected, actual, "findMissingComp")
			})
		}
	})

	t.Run("is same as same types", func(t *testing.T) {
		comp := func(r, c stringSame) bool {
			return r.IsSameAs(c)
		}
		for _, tc := range testCasesForFindMissing() {
			t.Run(tc.name, func(t *testing.T) {
				required := newStringSames(tc.required)
				toCheck := newStringSames(tc.toCheck)
				expected := newStringSames(tc.expected)
				actual := keeper.FindMissingComp(required, toCheck, comp)
				assert.Equal(t, expected, actual, "findMissingComp")
			})
		}
	})

	t.Run("is same as different types", func(t *testing.T) {
		comp := func(r stringSameR, c stringSameC) bool {
			return r.IsSameAs(c)
		}
		for _, tc := range testCasesForFindMissing() {
			t.Run(tc.name, func(t *testing.T) {
				required := newStringSameRs(tc.required)
				toCheck := newStringSameCs(tc.toCheck)
				expected := newStringSameRs(tc.expected)
				actual := keeper.FindMissingComp(required, toCheck, comp)
				assert.Equal(t, expected, actual, "findMissingComp")
			})
		}
	})

	t.Run("string lengths", func(t *testing.T) {
		comp := func(r string, c int) bool {
			return len(r) == c
		}
		req := []string{"a", "bb", "ccc", "dddd", "eeeee"}
		checks := []struct {
			name     string
			toCheck  []int
			expected []string
		}{
			{name: "all there", toCheck: []int{1, 2, 3, 4, 5}, expected: nil},
			{name: "missing len 1", toCheck: []int{2, 3, 4, 5}, expected: []string{"a"}},
			{name: "missing len 2", toCheck: []int{1, 3, 4, 5}, expected: []string{"bb"}},
			{name: "missing len 3", toCheck: []int{1, 2, 4, 5}, expected: []string{"ccc"}},
			{name: "missing len 4", toCheck: []int{1, 2, 3, 5}, expected: []string{"dddd"}},
			{name: "missing len 5", toCheck: []int{1, 2, 3, 4}, expected: []string{"eeeee"}},
			{name: "none there", toCheck: []int{0, 6}, expected: req},
		}
		for _, tc := range checks {
			t.Run(tc.name, func(t *testing.T) {
				actual := keeper.FindMissingComp(req, tc.toCheck, comp)
				assert.Equal(t, tc.expected, actual, "findMissingComp")
			})
		}
	})

	t.Run("div two", func(t *testing.T) {
		comp := func(r int, c int) bool {
			return r/2 == c
		}
		req := []int{1, 2, 3, 4, 5}
		checks := []struct {
			name     string
			toCheck  []int
			expected []int
		}{
			{name: "all there", toCheck: []int{0, 1, 2}, expected: nil},
			{name: "missing 0", toCheck: []int{1, 2}, expected: []int{1}},
			{name: "missing 1", toCheck: []int{0, 2}, expected: []int{2, 3}},
			{name: "missing 2", toCheck: []int{0, 1}, expected: []int{4, 5}},
			{name: "none there", toCheck: []int{-1, 3}, expected: req},
		}
		for _, tc := range checks {
			t.Run(tc.name, func(t *testing.T) {
				actual := keeper.FindMissingComp(req, tc.toCheck, comp)
				assert.Equal(t, tc.expected, actual, "findMissingComp")
			})
		}
	})

	t.Run("all true", func(t *testing.T) {
		comp := func(r, c string) bool {
			return true
		}
		for _, tc := range testCasesForFindMissing() {
			t.Run(tc.name, func(t *testing.T) {
				var expected []string
				// required entries are only marked as found after being compared to something.
				// So if there's nothing in the toCheck list, all the required will be returned.
				// But if tc.required is an empty slice, we still expect to get nil back, so we don't
				// set expected = tc.required in that case.
				if len(tc.toCheck) == 0 && len(tc.required) > 0 {
					expected = tc.required
				}
				actual := keeper.FindMissingComp(tc.required, tc.toCheck, comp)
				assert.Equal(t, expected, actual, "findMissingComp comp always returns true")
			})
		}
	})

	t.Run("all false", func(t *testing.T) {
		comp := func(r, c string) bool {
			return false
		}
		for _, tc := range testCasesForFindMissing() {
			t.Run(tc.name, func(t *testing.T) {
				// If tc.required is nil, or an empty slice, we expect nil, otherwise, we always expect tc.required back.
				var expected []string
				if len(tc.required) > 0 {
					expected = tc.required
				}
				actual := keeper.FindMissingComp(tc.required, tc.toCheck, comp)
				assert.Equal(t, expected, actual, "findMissingComp comp always returns false")
			})
		}
	})
}

func TestPluralEnding(t *testing.T) {
	tests := []struct {
		i   int
		exp string
	}{
		{i: 0, exp: "s"},
		{i: 1, exp: ""},
		{i: -1, exp: "s"},
		{i: 2, exp: "s"},
		{i: 3, exp: "s"},
		{i: 5, exp: "s"},
		{i: 50, exp: "s"},
		{i: -100, exp: "s"},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%d", tc.i), func(t *testing.T) {
			actual := keeper.PluralEnding(tc.i)
			assert.Equal(t, tc.exp, actual, "pluralEnding(%d)", tc.i)
		})
	}
}

func TestSafeBech32ToAccAddresses(t *testing.T) {
	tests := []struct {
		name    string
		bech32s []string
		exp     []sdk.AccAddress
	}{
		{
			name:    "nil",
			bech32s: nil,
			exp:     []sdk.AccAddress{},
		},
		{
			name:    "empty",
			bech32s: []string{},
			exp:     []sdk.AccAddress{},
		},
		{
			name:    "one good",
			bech32s: []string{sdk.AccAddress("one_good_one________").String()},
			exp:     []sdk.AccAddress{sdk.AccAddress("one_good_one________")},
		},
		{
			name:    "one bad",
			bech32s: []string{"one_bad_one_________"},
			exp:     []sdk.AccAddress{},
		},
		{
			name:    "one empty",
			bech32s: []string{""},
			exp:     []sdk.AccAddress{},
		},
		{
			name: "three good",
			bech32s: []string{
				sdk.AccAddress("first_is_good_______").String(),
				sdk.AccAddress("second_is_good______").String(),
				sdk.AccAddress("third_is_good_______").String(),
			},
			exp: []sdk.AccAddress{
				sdk.AccAddress("first_is_good_______"),
				sdk.AccAddress("second_is_good______"),
				sdk.AccAddress("third_is_good_______"),
			},
		},
		{
			name: "three with first bad",
			bech32s: []string{
				"bad_first___________",
				sdk.AccAddress("second_is_good______").String(),
				sdk.AccAddress("third_is_good_______").String(),
			},
			exp: []sdk.AccAddress{
				sdk.AccAddress("second_is_good______"),
				sdk.AccAddress("third_is_good_______"),
			},
		},
		{
			name: "three with bad second",
			bech32s: []string{
				sdk.AccAddress("first_is_good_______").String(),
				"bad_second__________",
				sdk.AccAddress("third_is_good_______").String(),
			},
			exp: []sdk.AccAddress{
				sdk.AccAddress("first_is_good_______"),
				sdk.AccAddress("third_is_good_______"),
			},
		},
		{
			name: "three with bad third",
			bech32s: []string{
				sdk.AccAddress("first_is_good_______").String(),
				sdk.AccAddress("second_is_good______").String(),
				"bad_third___________",
			},
			exp: []sdk.AccAddress{
				sdk.AccAddress("first_is_good_______"),
				sdk.AccAddress("second_is_good______"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := keeper.SafeBech32ToAccAddresses(tc.bech32s)
			assert.Equal(t, tc.exp, actual, "safeBech32ToAccAddresses")
		})
	}
}
