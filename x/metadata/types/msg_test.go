package types

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func ownerPartyList(addresses ...string) []Party {
	retval := make([]Party, len(addresses))
	for i, addr := range addresses {
		retval[i] = Party{Address: addr, Role: PartyType_PARTY_TYPE_OWNER}
	}
	return retval
}

// getTypeName gets just the type name of the provided thing e.g "MsgWriteScopeRequest".
func getTypeName(thing interface{}) string {
	rv := fmt.Sprintf("%T", thing) // e.g. "*types.MsgWriteScopeRequest"
	lastDot := strings.LastIndex(rv, ".")
	if lastDot < 0 || lastDot+1 >= len(rv) {
		return rv
	}
	return rv[lastDot+1:]
}

func TestAllMsgsGetSigners(t *testing.T) {
	addr1 := sdk.AccAddress("addr1_______________")
	addr2 := sdk.AccAddress("addr2_______________")
	addr3 := sdk.AccAddress("addr3_______________")

	badAddrStr := "badaddr"
	badAddrErr := "decoding bech32 failed: invalid bech32 string length 7"
	emptyAddrErr := "empty address string is not allowed"

	multiSignerMsgMakers := []func(signers []string) MetadataMsg{
		func(signers []string) MetadataMsg { return &MsgWriteScopeRequest{Signers: signers} },
		func(signers []string) MetadataMsg { return &MsgDeleteScopeRequest{Signers: signers} },
		func(signers []string) MetadataMsg { return &MsgAddScopeDataAccessRequest{Signers: signers} },
		func(signers []string) MetadataMsg { return &MsgDeleteScopeDataAccessRequest{Signers: signers} },
		func(signers []string) MetadataMsg { return &MsgAddScopeOwnerRequest{Signers: signers} },
		func(signers []string) MetadataMsg { return &MsgDeleteScopeOwnerRequest{Signers: signers} },
		func(signers []string) MetadataMsg { return &MsgUpdateValueOwnersRequest{Signers: signers} },
		func(signers []string) MetadataMsg { return &MsgMigrateValueOwnerRequest{Signers: signers} },
		func(signers []string) MetadataMsg { return &MsgWriteSessionRequest{Signers: signers} },
		func(signers []string) MetadataMsg { return &MsgWriteRecordRequest{Signers: signers} },
		func(signers []string) MetadataMsg { return &MsgDeleteRecordRequest{Signers: signers} },
		func(signers []string) MetadataMsg { return &MsgWriteScopeSpecificationRequest{Signers: signers} },
		func(signers []string) MetadataMsg { return &MsgDeleteScopeSpecificationRequest{Signers: signers} },
		func(signers []string) MetadataMsg { return &MsgWriteContractSpecificationRequest{Signers: signers} },
		func(signers []string) MetadataMsg { return &MsgDeleteContractSpecificationRequest{Signers: signers} },
		func(signers []string) MetadataMsg { return &MsgAddContractSpecToScopeSpecRequest{Signers: signers} },
		func(signers []string) MetadataMsg {
			return &MsgDeleteContractSpecFromScopeSpecRequest{Signers: signers}
		},
		func(signers []string) MetadataMsg { return &MsgWriteRecordSpecificationRequest{Signers: signers} },
		func(signers []string) MetadataMsg { return &MsgDeleteRecordSpecificationRequest{Signers: signers} },
	}

	singleSignerMsgMakers := []func(signer string) MetadataMsg{
		func(signer string) MetadataMsg {
			return &MsgBindOSLocatorRequest{Locator: ObjectStoreLocator{Owner: signer}}
		},
		func(signer string) MetadataMsg {
			return &MsgDeleteOSLocatorRequest{Locator: ObjectStoreLocator{Owner: signer}}
		},
		func(signer string) MetadataMsg {
			return &MsgModifyOSLocatorRequest{Locator: ObjectStoreLocator{Owner: signer}}
		},
	}

	multiSignerCases := []struct {
		name       string
		msgSigners []string
		expSigners []sdk.AccAddress
		expPanic   string
	}{
		{
			name:       "no signers",
			msgSigners: []string{},
			expSigners: []sdk.AccAddress{},
		},
		{
			name:       "one good signer",
			msgSigners: []string{addr1.String()},
			expSigners: []sdk.AccAddress{addr1},
		},
		{
			name:       "one bad signer",
			msgSigners: []string{badAddrStr},
			expPanic:   badAddrErr,
		},
		{
			name:       "three good signers",
			msgSigners: []string{addr1.String(), addr2.String(), addr3.String()},
			expSigners: []sdk.AccAddress{addr1, addr2, addr3},
		},
		{
			name:       "three signers 1st bad",
			msgSigners: []string{badAddrStr, addr2.String(), addr3.String()},
			expPanic:   badAddrErr,
		},
		{
			name:       "three signers 2nd bad",
			msgSigners: []string{addr1.String(), badAddrStr, addr3.String()},
			expPanic:   badAddrErr,
		},
		{
			name:       "three signers 3rd bad",
			msgSigners: []string{addr1.String(), addr2.String(), badAddrStr},
			expPanic:   badAddrErr,
		},
	}

	singleSignerCases := []struct {
		name       string
		msgSigner  string
		expSigners []sdk.AccAddress
		expPanic   string
	}{
		{
			name:      "no signer",
			msgSigner: "",
			expPanic:  emptyAddrErr,
		},
		{
			name:       "good signer",
			msgSigner:  addr1.String(),
			expSigners: []sdk.AccAddress{addr1},
		},
		{
			name:      "bad signer",
			msgSigner: badAddrStr,
			expPanic:  badAddrErr,
		},
	}

	type testCase struct {
		name           string
		msg            MetadataMsg
		expSigners     []sdk.AccAddress
		expPanic       string
		expSignersStrs []string
	}

	var tests []testCase
	hasMaker := make(map[string]bool)

	for _, msgMaker := range multiSignerMsgMakers {
		typeName := getTypeName(msgMaker(nil))
		hasMaker[typeName] = true
		for _, tc := range multiSignerCases {
			tests = append(tests, testCase{
				name:           typeName + " " + tc.name,
				msg:            msgMaker(tc.msgSigners),
				expSigners:     tc.expSigners,
				expPanic:       tc.expPanic,
				expSignersStrs: tc.msgSigners,
			})
		}
	}

	for _, msgMaker := range singleSignerMsgMakers {
		typeName := getTypeName(msgMaker(""))
		hasMaker[typeName] = true
		for _, tc := range singleSignerCases {
			tests = append(tests, testCase{
				name:           typeName + " " + tc.name,
				msg:            msgMaker(tc.msgSigner),
				expSigners:     tc.expSigners,
				expPanic:       tc.expPanic,
				expSignersStrs: []string{tc.msgSigner},
			})
		}
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var signers []sdk.AccAddress
			testGetSigners := func() {
				signers = tc.msg.GetSigners()
			}
			if len(tc.expPanic) > 0 {
				require.PanicsWithError(t, tc.expPanic, testGetSigners, "GetSigners")
			} else {
				require.NotPanics(t, testGetSigners, "GetSigners")
				assert.Equal(t, tc.expSigners, signers, "GetSigners")
			}

			var signersStrs []string
			testGetSignerStrs := func() {
				signersStrs = tc.msg.GetSignerStrs()
			}
			require.NotPanics(t, testGetSignerStrs, "GetSignerStrs")
			assert.Equal(t, tc.expSignersStrs, signersStrs, "GetSignerStrs")
		})
	}

	// Make sure all of the GetSigners and GetSignerStrs funcs are tested.
	t.Run("all msgs have test case", func(t *testing.T) {
		for _, msg := range allRequestMsgs {
			typeName := getTypeName(msg)
			// If this fails, a maker needs to be defined above for the missing msg type.
			if !assert.True(t, hasMaker[typeName], "hasMaker[%q]", typeName) {
				t.Logf("Also make sure that a TypeURL%s is defined in msg.go = %q", typeName, sdk.MsgTypeURL(msg))
			}
		}
	})
}

func TestWriteScopeRoute(t *testing.T) {
	var scope = &Scope{
		ScopeId:           ScopeMetadataAddress(uuid.MustParse("8d80b25a-c089-4446-956e-5d08cfe3e1a5")),
		SpecificationId:   ScopeSpecMetadataAddress(uuid.MustParse("22fc17a6-40dd-4d68-a95b-ec94e7572a09")),
		Owners:            ownerPartyList("data_owner"),
		DataAccess:        []string{"data_accessor"},
		ValueOwnerAddress: "value_owner",
	}
	var msg = NewMsgWriteScopeRequest(*scope, []string{})

	require.Equal(t, sdk.MsgTypeURL(msg), "/provenance.metadata.v1.MsgWriteScopeRequest")
	expectedYaml := `scope:
  scope_id: scope1qzxcpvj6czy5g354dews3nlruxjsahhnsp
  specification_id: scopespec1qs30c9axgrw5669ft0kffe6h9gysfe58v3
  owners:
  - address: data_owner
    role: 5
    optional: false
  data_access:
  - data_accessor
  value_owner_address: value_owner
  require_party_rollup: false
signers: []
scope_uuid: ""
spec_uuid: ""
`
	bz, err := yaml.Marshal(msg)
	require.NoError(t, err, "yaml.Marshal(msg)")
	assert.Equal(t, expectedYaml, string(bz), "scope as yaml")

	bz, err = ModuleCdc.MarshalJSON(msg)
	require.NoError(t, err, "ModuleCdc.MarshalJSON(msg)")
	assert.Equal(t, "{\"scope\":{\"scope_id\":\"scope1qzxcpvj6czy5g354dews3nlruxjsahhnsp\",\"specification_id\":\"scopespec1qs30c9axgrw5669ft0kffe6h9gysfe58v3\",\"owners\":[{\"address\":\"data_owner\",\"role\":\"PARTY_TYPE_OWNER\",\"optional\":false}],\"data_access\":[\"data_accessor\"],\"value_owner_address\":\"value_owner\",\"require_party_rollup\":false},\"signers\":[],\"scope_uuid\":\"\",\"spec_uuid\":\"\"}", string(bz))
}

func TestWriteScopeValidation(t *testing.T) {
	var scope = &Scope{
		ScopeId:           ScopeMetadataAddress(uuid.MustParse("8d80b25a-c089-4446-956e-5d08cfe3e1a5")),
		SpecificationId:   ScopeSpecMetadataAddress(uuid.MustParse("22fc17a6-40dd-4d68-a95b-ec94e7572a09")),
		Owners:            ownerPartyList("data_owner"),
		DataAccess:        []string{"data_accessor"},
		ValueOwnerAddress: "value_owner",
	}
	var msg = NewMsgWriteScopeRequest(*scope, []string{"invalid"})
	err := msg.ValidateBasic()
	require.EqualError(t, err, "invalid scope owners: invalid party address [data_owner]: decoding bech32 failed: invalid separator index -1")
	require.Panics(t, func() { msg.GetSigners() }, "panics due to invalid addresses")

	err = msg.Scope.ValidateBasic()
	require.Error(t, err, "invalid addresses")
	require.Equal(t, "invalid scope owners: invalid party address [data_owner]: decoding bech32 failed: invalid separator index -1", err.Error())

	msg.Scope = Scope{
		ScopeId:         ScopeMetadataAddress(uuid.MustParse("8d80b25a-c089-4446-956e-5d08cfe3e1a5")),
		SpecificationId: ScopeSpecMetadataAddress(uuid.MustParse("22fc17a6-40dd-4d68-a95b-ec94e7572a09")),
		Owners:          []Party{},
		DataAccess:      []string{},
	}
	err = msg.Scope.ValidateBasic()
	require.Error(t, err, "no owners")
	require.Equal(t, "invalid scope owners: at least one party is required", err.Error())

	msg.Scope = Scope{
		ScopeId:         ScopeMetadataAddress(uuid.MustParse("8d80b25a-c089-4446-956e-5d08cfe3e1a5")),
		SpecificationId: ScopeSpecMetadataAddress(uuid.MustParse("22fc17a6-40dd-4d68-a95b-ec94e7572a09")),
		Owners:          ownerPartyList("cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"),
		DataAccess:      []string{},
	}
	msg.Signers = []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}
	err = msg.Scope.ValidateBasic()
	require.NoError(t, err, "valid add scope request")
	requiredSigners := msg.GetSigners()
	require.Equal(t, 1, len(requiredSigners))
	x, err := hex.DecodeString("85EA54E8598B27EC37EAEEEEA44F1E78A9B5E671")
	require.NoError(t, err)
	require.Equal(t, sdk.AccAddress(x), requiredSigners[0])
}

func TestAddScopeDataAccessValidateBasic(t *testing.T) {
	notAScopeId := RecordMetadataAddress(uuid.New(), "recordname")
	actualScopeId := ScopeMetadataAddress(uuid.New())

	cases := map[string]struct {
		msg      *MsgAddScopeDataAccessRequest
		wantErr  bool
		errorMsg string
	}{
		"should fail to validate basic, incorrect scope id type": {
			NewMsgAddScopeDataAccessRequest(notAScopeId, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			true,
			fmt.Sprintf("address is not a scope id: %v", notAScopeId.String()),
		},
		"should fail to validate basic, requires at least one data access address": {
			NewMsgAddScopeDataAccessRequest(actualScopeId, []string{}, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			true,
			"at least one data access address is required",
		},
		"should fail to validate basic, incorrect data access address format": {
			NewMsgAddScopeDataAccessRequest(actualScopeId, []string{"notabech32address"}, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			true,
			"data access address is invalid: notabech32address",
		},
		"should fail to validate basic, requires at least one signer": {
			NewMsgAddScopeDataAccessRequest(actualScopeId, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}, []string{}),
			true,
			"at least one signer is required",
		},
		"should successfully validate basic": {
			NewMsgAddScopeDataAccessRequest(actualScopeId, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			false,
			"",
		},
	}

	for n, tc := range cases {
		tc := tc

		t.Run(n, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.wantErr {
				require.Error(t, err)
				require.Equal(t, tc.errorMsg, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDeleteScopeDataAccessValidateBasic(t *testing.T) {
	notAScopeId := RecordMetadataAddress(uuid.New(), "recordname")
	actualScopeId := ScopeMetadataAddress(uuid.New())

	cases := map[string]struct {
		msg      *MsgDeleteScopeDataAccessRequest
		wantErr  bool
		errorMsg string
	}{
		"should fail to validate basic, incorrect scope id type": {
			NewMsgDeleteScopeDataAccessRequest(notAScopeId, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			true,
			fmt.Sprintf("address is not a scope id: %v", notAScopeId.String()),
		},
		"should fail to validate basic, requires at least one data access address": {
			NewMsgDeleteScopeDataAccessRequest(actualScopeId, []string{}, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			true,
			"at least one data access address is required",
		},
		"should fail to validate basic, incorrect data access address format": {
			NewMsgDeleteScopeDataAccessRequest(actualScopeId, []string{"notabech32address"}, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			true,
			"data access address is invalid: notabech32address",
		},
		"should fail to validate basic, requires at least one signer": {
			NewMsgDeleteScopeDataAccessRequest(actualScopeId, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}, []string{}),
			true,
			"at least one signer is required",
		},
		"should successfully validate basic": {
			NewMsgDeleteScopeDataAccessRequest(actualScopeId, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			false,
			"",
		},
	}

	for n, tc := range cases {
		tc := tc

		t.Run(n, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.wantErr {
				require.Error(t, err)
				require.Equal(t, tc.errorMsg, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAddScopeOwnersValidateBasic(t *testing.T) {
	notAScopeId := RecordMetadataAddress(uuid.New(), "recordname")
	actualScopeId := ScopeMetadataAddress(uuid.New())

	cases := []struct {
		name     string
		msg      *MsgAddScopeOwnerRequest
		errorMsg string
	}{
		{
			"should fail to validate basic, incorrect scope id type",
			NewMsgAddScopeOwnerRequest(
				notAScopeId,
				[]Party{{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: PartyType_PARTY_TYPE_OWNER}},
				[]string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			fmt.Sprintf("address is not a scope id: %v", notAScopeId.String()),
		},
		{
			"should fail to validate basic, requires at least one owner address",
			NewMsgAddScopeOwnerRequest(
				actualScopeId,
				[]Party{},
				[]string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"},
			),
			"invalid owners: at least one party is required",
		},
		{
			"should fail to validate basic, incorrect owner address format",
			NewMsgAddScopeOwnerRequest(
				actualScopeId,
				[]Party{{Address: "notabech32address", Role: PartyType_PARTY_TYPE_OWNER}},
				[]string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"},
			),
			"invalid owners: invalid party address [notabech32address]: decoding bech32 failed: invalid separator index -1",
		},
		{
			"should fail to validate basic, incorrect party type",
			NewMsgAddScopeOwnerRequest(
				actualScopeId,
				[]Party{{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: PartyType_PARTY_TYPE_UNSPECIFIED}},
				[]string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"},
			),
			"invalid owners: invalid party type for party cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck",
		},
		{
			"should fail to validate basic, requires at least one signer",
			NewMsgAddScopeOwnerRequest(
				actualScopeId,
				[]Party{{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: PartyType_PARTY_TYPE_OWNER}},
				[]string{},
			),
			"at least one signer is required",
		},
		{
			"should successfully validate basic",
			NewMsgAddScopeOwnerRequest(
				actualScopeId,
				[]Party{{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: PartyType_PARTY_TYPE_OWNER}},
				[]string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"},
			),
			"",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if len(tc.errorMsg) > 0 {
				require.EqualError(t, err, tc.errorMsg, "MsgAddScopeOwnerRequest.ValidateBasic expected error")
			} else {
				require.NoError(t, err, "MsgAddScopeOwnerRequest.ValidateBasic unexpected error")
			}
		})
	}
}

func TestDeleteScopeOwnerValidateBasic(t *testing.T) {
	notAScopeId := RecordMetadataAddress(uuid.New(), "recordname")
	actualScopeId := ScopeMetadataAddress(uuid.New())

	cases := map[string]struct {
		msg      *MsgDeleteScopeOwnerRequest
		wantErr  bool
		errorMsg string
	}{
		"should fail to validate basic, incorrect scope id type": {
			NewMsgDeleteScopeOwnerRequest(notAScopeId, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			true,
			fmt.Sprintf("address is not a scope id: %v", notAScopeId.String()),
		},
		"should fail to validate basic, requires at least one owner address": {
			NewMsgDeleteScopeOwnerRequest(actualScopeId, []string{}, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			true,
			"at least one owner address is required",
		},
		"should fail to validate basic, incorrect data access address format": {
			NewMsgDeleteScopeOwnerRequest(actualScopeId, []string{"notabech32address"}, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			true,
			"owner address is invalid: notabech32address",
		},
		"should fail to validate basic, requires at least one signer": {
			NewMsgDeleteScopeOwnerRequest(actualScopeId, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}, []string{}),
			true,
			"at least one signer is required",
		},
		"should successfully validate basic": {
			NewMsgDeleteScopeOwnerRequest(actualScopeId, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			false,
			"",
		},
	}

	for n, tc := range cases {
		tc := tc

		t.Run(n, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.wantErr {
				require.Error(t, err)
				require.Equal(t, tc.errorMsg, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgUpdateValueOwnersRequest_ValidateBasic(t *testing.T) {
	badScopeIDErr := func(i int, id MetadataAddress) string {
		return fmt.Sprintf("scope id[%d]: %q: invalid scope id", i, id.String())
	}
	notAScopeID := ScopeSpecMetadataAddress(uuid.New())

	tests := []struct {
		name string
		msg  MsgUpdateValueOwnersRequest
		exp  string
	}{
		{
			name: "control",
			msg: MsgUpdateValueOwnersRequest{
				ScopeIds:          []MetadataAddress{ScopeMetadataAddress(uuid.New())},
				ValueOwnerAddress: sdk.AccAddress("ValueOwnerAddress___").String(),
				Signers:           []string{sdk.AccAddress("signer______________").String()},
			},
			exp: "",
		},
		{
			name: "no scope ids",
			msg: MsgUpdateValueOwnersRequest{
				ScopeIds:          []MetadataAddress{},
				ValueOwnerAddress: sdk.AccAddress("ValueOwnerAddress___").String(),
				Signers:           []string{sdk.AccAddress("signer______________").String()},
			},
			exp: "at least one scope id is required",
		},
		{
			name: "one bad scope id",
			msg: MsgUpdateValueOwnersRequest{
				ScopeIds:          []MetadataAddress{notAScopeID},
				ValueOwnerAddress: sdk.AccAddress("ValueOwnerAddress___").String(),
				Signers:           []string{sdk.AccAddress("signer______________").String()},
			},
			exp: badScopeIDErr(0, notAScopeID),
		},
		{
			name: "two valid scope ids",
			msg: MsgUpdateValueOwnersRequest{
				ScopeIds: []MetadataAddress{
					ScopeMetadataAddress(uuid.New()), ScopeMetadataAddress(uuid.New()),
				},
				ValueOwnerAddress: sdk.AccAddress("ValueOwnerAddress___").String(),
				Signers:           []string{sdk.AccAddress("signer______________").String()},
			},
			exp: "",
		},
		{
			name: "two scope ids 1st bad ",
			msg: MsgUpdateValueOwnersRequest{
				ScopeIds: []MetadataAddress{
					notAScopeID, ScopeMetadataAddress(uuid.New()),
				},
				ValueOwnerAddress: sdk.AccAddress("ValueOwnerAddress___").String(),
				Signers:           []string{sdk.AccAddress("signer______________").String()},
			},
			exp: badScopeIDErr(0, notAScopeID),
		},
		{
			name: "two scope ids 2nd bad ",
			msg: MsgUpdateValueOwnersRequest{
				ScopeIds: []MetadataAddress{
					ScopeMetadataAddress(uuid.New()), notAScopeID,
				},
				ValueOwnerAddress: sdk.AccAddress("ValueOwnerAddress___").String(),
				Signers:           []string{sdk.AccAddress("signer______________").String()},
			},
			exp: badScopeIDErr(1, notAScopeID),
		},
		{
			name: "empty value owner",
			msg: MsgUpdateValueOwnersRequest{
				ScopeIds:          []MetadataAddress{ScopeMetadataAddress(uuid.New())},
				ValueOwnerAddress: "",
				Signers:           []string{sdk.AccAddress("signer______________").String()},
			},
			exp: "invalid value owner address: empty address string is not allowed",
		},
		{
			name: "bad value owner",
			msg: MsgUpdateValueOwnersRequest{
				ScopeIds:          []MetadataAddress{ScopeMetadataAddress(uuid.New())},
				ValueOwnerAddress: "badaddr",
				Signers:           []string{sdk.AccAddress("signer______________").String()},
			},
			exp: "invalid value owner address: decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			name: "no signers",
			msg: MsgUpdateValueOwnersRequest{
				ScopeIds:          []MetadataAddress{ScopeMetadataAddress(uuid.New())},
				ValueOwnerAddress: sdk.AccAddress("ValueOwnerAddress___").String(),
				Signers:           []string{},
			},
			exp: "at least one signer is required",
		},
		{
			name: "one bad signer",
			msg: MsgUpdateValueOwnersRequest{
				ScopeIds:          []MetadataAddress{ScopeMetadataAddress(uuid.New())},
				ValueOwnerAddress: sdk.AccAddress("ValueOwnerAddress___").String(),
				Signers:           []string{"badsigner"},
			},
			// Not expecting an error here. This check is not part of ValidateBasic
			// because of the assumption that GetSigners() is called ealier and will
			// panic on a bad signer.
			exp: "",
		},
		{
			name: "lots of scope ids and signers",
			msg: MsgUpdateValueOwnersRequest{
				ScopeIds: []MetadataAddress{
					ScopeMetadataAddress(uuid.New()), ScopeMetadataAddress(uuid.New()),
					ScopeMetadataAddress(uuid.New()), ScopeMetadataAddress(uuid.New()),
					ScopeMetadataAddress(uuid.New()), ScopeMetadataAddress(uuid.New()),
					ScopeMetadataAddress(uuid.New()), ScopeMetadataAddress(uuid.New()),
					ScopeMetadataAddress(uuid.New()), ScopeMetadataAddress(uuid.New()),
					ScopeMetadataAddress(uuid.New()), ScopeMetadataAddress(uuid.New()),
				},
				ValueOwnerAddress: sdk.AccAddress("value_owner_address_").String(),
				Signers: []string{
					sdk.AccAddress("signer__0___________").String(),
					sdk.AccAddress("signer__1___________").String(),
					sdk.AccAddress("signer__2___________").String(),
					sdk.AccAddress("signer__3___________").String(),
					sdk.AccAddress("signer__4___________").String(),
					sdk.AccAddress("signer__5___________").String(),
					sdk.AccAddress("signer__6___________").String(),
					sdk.AccAddress("signer__7___________").String(),
					sdk.AccAddress("signer__8___________").String(),
					sdk.AccAddress("signer__9___________").String(),
					sdk.AccAddress("signer_10___________").String(),
					sdk.AccAddress("signer_11___________").String(),
					sdk.AccAddress("signer_12___________").String(),
					sdk.AccAddress("signer_13___________").String(),
					sdk.AccAddress("signer_14___________").String(),
				},
			},
			exp: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if len(tc.exp) > 0 {
				assert.EqualError(t, err, tc.exp, "ValidateBasic")
			} else {
				assert.NoError(t, err, "ValidateBasic")
			}
		})
	}
}

func TestMsgMigrateValueOwnerRequest_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgMigrateValueOwnerRequest
		exp  string
	}{
		{
			name: "control",
			msg: MsgMigrateValueOwnerRequest{
				Existing: sdk.AccAddress("existing_value_owner").String(),
				Proposed: sdk.AccAddress("proposed_value_owner").String(),
				Signers:  []string{"signer1"},
			},
			exp: "",
		},
		{
			name: "missing existing",
			msg: MsgMigrateValueOwnerRequest{
				Existing: "",
				Proposed: sdk.AccAddress("proposed_value_owner").String(),
				Signers:  []string{"signer1"},
			},
			exp: "invalid existing value owner address: empty address string is not allowed",
		},
		{
			name: "invalid existing",
			msg: MsgMigrateValueOwnerRequest{
				Existing: "notanaddress",
				Proposed: sdk.AccAddress("proposed_value_owner").String(),
				Signers:  []string{"signer1"},
			},
			exp: "invalid existing value owner address: decoding bech32 failed: invalid separator index -1",
		},
		{
			name: "empty proposed",
			msg: MsgMigrateValueOwnerRequest{
				Existing: sdk.AccAddress("existing_value_owner").String(),
				Proposed: "",
				Signers:  []string{"signer1"},
			},
			exp: "invalid proposed value owner address: empty address string is not allowed",
		},
		{
			name: "invalid proposed",
			msg: MsgMigrateValueOwnerRequest{
				Existing: sdk.AccAddress("existing_value_owner").String(),
				Proposed: "notanaddress",
				Signers:  []string{"signer1"},
			},
			exp: "invalid proposed value owner address: decoding bech32 failed: invalid separator index -1",
		},
		{
			name: "nil signers",
			msg: MsgMigrateValueOwnerRequest{
				Existing: sdk.AccAddress("existing_value_owner").String(),
				Proposed: sdk.AccAddress("proposed_value_owner").String(),
				Signers:  nil,
			},
			exp: "at least one signer is required",
		},
		{
			name: "empty signers",
			msg: MsgMigrateValueOwnerRequest{
				Existing: sdk.AccAddress("existing_value_owner").String(),
				Proposed: sdk.AccAddress("proposed_value_owner").String(),
				Signers:  []string{},
			},
			exp: "at least one signer is required",
		},
		{
			name: "five signers",
			msg: MsgMigrateValueOwnerRequest{
				Existing: sdk.AccAddress("existing_value_owner").String(),
				Proposed: sdk.AccAddress("proposed_value_owner").String(),
				Signers:  []string{"signer1", "signer2", "signer3", "signer4", "signer5"},
			},
			exp: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if len(tc.exp) > 0 {
				assert.EqualError(t, err, tc.exp, "ValidateBasic")
			} else {
				assert.NoError(t, err, "ValidateBasic")
			}
		})
	}
}

func TestMsgAddContractSpecToScopeSpecRequestValidateBasic(t *testing.T) {
	contractSpecID := ContractSpecMetadataAddress(uuid.New())
	scopeSpecID := ScopeSpecMetadataAddress(uuid.New())

	cases := map[string]struct {
		msg      *MsgAddContractSpecToScopeSpecRequest
		wantErr  bool
		errorMsg string
	}{
		"should fail to validate basic, incorrect contract spec id type": {
			NewMsgAddContractSpecToScopeSpecRequest(scopeSpecID, scopeSpecID, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			true,
			fmt.Sprintf("address is not a contract specification id: %v", scopeSpecID.String()),
		},
		"should fail to validate basic, incorrect scope spec id type": {
			NewMsgAddContractSpecToScopeSpecRequest(contractSpecID, contractSpecID, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			true,
			fmt.Sprintf("address is not a scope specification id: %v", contractSpecID.String()),
		},
		"should fail to validate basic, requires at least one signer": {
			NewMsgAddContractSpecToScopeSpecRequest(contractSpecID, scopeSpecID, []string{}),
			true,
			"at least one signer is required",
		},
		"should successfully validate basic": {
			NewMsgAddContractSpecToScopeSpecRequest(contractSpecID, scopeSpecID, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			false,
			"",
		},
	}

	for n, tc := range cases {
		tc := tc

		t.Run(n, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.wantErr {
				require.Error(t, err)
				require.Equal(t, tc.errorMsg, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgDeleteContractSpecFromScopeSpecRequestValidateBasic(t *testing.T) {
	contractSpecID := ContractSpecMetadataAddress(uuid.New())
	scopeSpecID := ScopeSpecMetadataAddress(uuid.New())

	cases := map[string]struct {
		msg      *MsgDeleteContractSpecFromScopeSpecRequest
		wantErr  bool
		errorMsg string
	}{
		"should fail to validate basic, incorrect contract spec id type": {
			NewMsgDeleteContractSpecFromScopeSpecRequest(scopeSpecID, scopeSpecID, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			true,
			fmt.Sprintf("address is not a contract specification id: %v", scopeSpecID.String()),
		},
		"should fail to validate basic, incorrect scope spec id type": {
			NewMsgDeleteContractSpecFromScopeSpecRequest(contractSpecID, contractSpecID, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			true,
			fmt.Sprintf("address is not a scope specification id: %v", contractSpecID.String()),
		},
		"should fail to validate basic, requires at least one signer": {
			NewMsgDeleteContractSpecFromScopeSpecRequest(contractSpecID, scopeSpecID, []string{}),
			true,
			"at least one signer is required",
		},
		"should successfully validate basic": {
			NewMsgDeleteContractSpecFromScopeSpecRequest(contractSpecID, scopeSpecID, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			false,
			"",
		},
	}

	for n, tc := range cases {
		tc := tc

		t.Run(n, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.wantErr {
				require.Error(t, err)
				require.Equal(t, tc.errorMsg, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBindOSLocator(t *testing.T) {
	var bindRequestMsg = NewMsgBindOSLocatorRequest(ObjectStoreLocator{Owner: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", LocatorUri: "http://foo.com"})

	err := bindRequestMsg.ValidateBasic()
	require.NoError(t, err)
	signers := bindRequestMsg.GetSigners()
	require.Equal(t, "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", signers[0].String())
	require.Equal(t, "/provenance.metadata.v1.MsgBindOSLocatorRequest", sdk.MsgTypeURL(bindRequestMsg))

	bz, _ := ModuleCdc.MarshalJSON(bindRequestMsg)
	require.Equal(t, "{\"locator\":{\"owner\":\"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck\",\"locator_uri\":\"http://foo.com\",\"encryption_key\":\"\"}}", string(bz))
}

func TestModifyOSLocator(t *testing.T) {
	var modifyRequest = NewMsgModifyOSLocatorRequest(ObjectStoreLocator{Owner: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", LocatorUri: "http://foo.com"})

	err := modifyRequest.ValidateBasic()
	require.NoError(t, err)
	signers := modifyRequest.GetSigners()
	require.Equal(t, "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", signers[0].String())
	require.Equal(t, "/provenance.metadata.v1.MsgModifyOSLocatorRequest", sdk.MsgTypeURL(modifyRequest))

	bz, _ := ModuleCdc.MarshalJSON(modifyRequest)
	require.Equal(t, "{\"locator\":{\"owner\":\"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck\",\"locator_uri\":\"http://foo.com\",\"encryption_key\":\"\"}}", string(bz))
}

func TestDeleteOSLocator(t *testing.T) {
	var deleteRequest = NewMsgDeleteOSLocatorRequest(ObjectStoreLocator{Owner: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", LocatorUri: "http://foo.com"})

	err := deleteRequest.ValidateBasic()
	require.NoError(t, err)

	signers := deleteRequest.GetSigners()
	require.Equal(t, "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", signers[0].String())
	require.Equal(t, "/provenance.metadata.v1.MsgDeleteOSLocatorRequest", sdk.MsgTypeURL(deleteRequest))

	bz, _ := ModuleCdc.MarshalJSON(deleteRequest)
	require.Equal(t, "{\"locator\":{\"owner\":\"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck\",\"locator_uri\":\"http://foo.com\",\"encryption_key\":\"\"}}", string(bz))
}

func TestBindOSLocatorInvalid(t *testing.T) {
	var bindRequestMsg = NewMsgBindOSLocatorRequest(ObjectStoreLocator{Owner: "vamonos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", LocatorUri: "http://foo.com"})

	err := bindRequestMsg.ValidateBasic()
	require.Error(t, err)
}

func TestBindOSLocatorInvalidAddr(t *testing.T) {
	var bindRequestMsg = NewMsgBindOSLocatorRequest(ObjectStoreLocator{Owner: "", LocatorUri: "http://foo.com"})

	err := bindRequestMsg.ValidateBasic()
	require.Error(t, err)
}

func TestBindOSLocatorInvalidURI(t *testing.T) {
	var bindRequestMsg = NewMsgBindOSLocatorRequest(ObjectStoreLocator{Owner: "", LocatorUri: "foo://foo.com"})

	err := bindRequestMsg.ValidateBasic()
	require.Error(t, err)
}

// TestPrintMessageTypeStrings just prints out all the MsgTypeURLs.
// The output can be copy/pasted into the const area in msg.go
func TestPrintMessageTypeStrings(t *testing.T) {
	var sb strings.Builder
	sb.WriteString("const (\n")
	for _, msg := range allRequestMsgs {
		typeName := getTypeName(msg) // e.g. "*types.MsgWriteScopeRequest"
		typeURL := sdk.MsgTypeURL(msg)
		// Note: 41 comes from the length of the longest Msg name: MsgDeleteContractSpecFromScopeSpecRequest
		// That aligns the = in the way gofmt wants.
		line := fmt.Sprintf("\tTypeURL%-41s = %q\n", typeName, typeURL)
		sb.WriteString(line)
	}
	sb.WriteString(")\n")

	section := sb.String()
	fmt.Printf("%s", section)
}

func TestRegisterInterfaces(t *testing.T) {
	// This test is mostly just a demonstration that the entries in allRequestMsgs can be
	// defined using the (*MsgWriteScopeRequest)(nil) pattern instead of &MsgWriteScopeRequest{}.
	// That's why this is in msg_test.go instead of codec_test.go.

	registry := cdctypes.NewInterfaceRegistry()
	RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	for _, msg := range allRequestMsgs {
		msgT := fmt.Sprintf("%T", msg)
		t.Run(msgT, func(t *testing.T) {
			msgTypeURL := sdk.MsgTypeURL(msg)
			msgActual, err := registry.Resolve(msgTypeURL)
			require.NoError(t, err, "registry.Resolve(%q) value", msgTypeURL)
			assert.NotNil(t, msgActual, "registry.Resolve(%q) message")
			actualT := fmt.Sprintf("%T", msgActual)
			assert.Equal(t, msgT, actualT, "resolved message type")

			msgToJSON, _ := registry.Resolve(msgTypeURL)
			err = cdc.UnmarshalJSON([]byte("{}"), msgToJSON)
			assert.NoError(t, err, "UnmarshalJSON(\"{}\")")
			assert.NotNil(t, msgToJSON, "message after UnmarshalJSON")

			msgToAny, _ := registry.Resolve(msgTypeURL)
			asAny, err := cdctypes.NewAnyWithValue(msgToAny)
			if assert.NoError(t, err, "NewAnyWithValue") {
				if assert.NotNil(t, asAny, "message wrapped as any") {
					err = cdc.UnpackAny(asAny, &msgToAny)
					if assert.NoError(t, err, "UnpackAny") {
						assert.Equal(t, msgActual, msgToAny, "message after unpacking it from the any")
					}
				}
			}
		})
	}
}
