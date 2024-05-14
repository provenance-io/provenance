package types_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/testutil"

	. "github.com/provenance-io/provenance/x/metadata/types"
)

// getTypeName gets just the type name of the provided thing e.g "MsgWriteScopeRequest".
func getTypeName(thing interface{}) string {
	rv := fmt.Sprintf("%T", thing) // e.g. "*types.MsgWriteScopeRequest"
	lastDot := strings.LastIndex(rv, ".")
	if lastDot < 0 || lastDot+1 >= len(rv) {
		return rv
	}
	return rv[lastDot+1:]
}

// cdc holds a fully built codec.
// Do not use this variable directly, instead use the GetCdc function to get it.
var cdc *codec.ProtoCodec

// GetCdc returns the codec, creating it if needed.
func GetCdc(t *testing.T) *codec.ProtoCodec {
	if cdc == nil {
		encCfg := app.MakeTestEncodingConfig(t)
		cdc = codec.NewProtoCodec(encCfg.InterfaceRegistry)
	}
	return cdc
}

func TestAllMsgsGetSigners(t *testing.T) {
	singleSignerMsgMakers := []testutil.MsgMaker{
		func(signer string) sdk.Msg {
			return &MsgBindOSLocatorRequest{Locator: ObjectStoreLocator{Owner: signer}}
		},
		func(signer string) sdk.Msg {
			return &MsgDeleteOSLocatorRequest{Locator: ObjectStoreLocator{Owner: signer}}
		},
		func(signer string) sdk.Msg {
			return &MsgModifyOSLocatorRequest{Locator: ObjectStoreLocator{Owner: signer}}
		},
	}

	multiSignerMsgMakers := []testutil.MsgMakerMulti{
		func(signers []string) sdk.Msg { return &MsgWriteScopeRequest{Signers: signers} },
		func(signers []string) sdk.Msg { return &MsgDeleteScopeRequest{Signers: signers} },
		func(signers []string) sdk.Msg { return &MsgAddScopeDataAccessRequest{Signers: signers} },
		func(signers []string) sdk.Msg { return &MsgDeleteScopeDataAccessRequest{Signers: signers} },
		func(signers []string) sdk.Msg { return &MsgAddScopeOwnerRequest{Signers: signers} },
		func(signers []string) sdk.Msg { return &MsgDeleteScopeOwnerRequest{Signers: signers} },
		func(signers []string) sdk.Msg { return &MsgUpdateValueOwnersRequest{Signers: signers} },
		func(signers []string) sdk.Msg { return &MsgMigrateValueOwnerRequest{Signers: signers} },
		func(signers []string) sdk.Msg { return &MsgWriteSessionRequest{Signers: signers} },
		func(signers []string) sdk.Msg { return &MsgWriteRecordRequest{Signers: signers} },
		func(signers []string) sdk.Msg { return &MsgDeleteRecordRequest{Signers: signers} },
		func(signers []string) sdk.Msg { return &MsgWriteScopeSpecificationRequest{Signers: signers} },
		func(signers []string) sdk.Msg { return &MsgDeleteScopeSpecificationRequest{Signers: signers} },
		func(signers []string) sdk.Msg { return &MsgWriteContractSpecificationRequest{Signers: signers} },
		func(signers []string) sdk.Msg { return &MsgDeleteContractSpecificationRequest{Signers: signers} },
		func(signers []string) sdk.Msg { return &MsgAddContractSpecToScopeSpecRequest{Signers: signers} },
		func(signers []string) sdk.Msg {
			return &MsgDeleteContractSpecFromScopeSpecRequest{Signers: signers}
		},
		func(signers []string) sdk.Msg { return &MsgWriteRecordSpecificationRequest{Signers: signers} },
		func(signers []string) sdk.Msg { return &MsgDeleteRecordSpecificationRequest{Signers: signers} },
		func(signers []string) sdk.Msg { return &MsgSetAccountDataRequest{Signers: signers} },
		func(signers []string) sdk.Msg { return &MsgAddNetAssetValuesRequest{Signers: signers} },
	}

	testutil.RunGetSignersTests(t, AllRequestMsgs, singleSignerMsgMakers, multiSignerMsgMakers)
}

func TestWriteScopeRoute(t *testing.T) {
	var scope = &Scope{
		ScopeId:           ScopeMetadataAddress(uuid.MustParse("8d80b25a-c089-4446-956e-5d08cfe3e1a5")),
		SpecificationId:   ScopeSpecMetadataAddress(uuid.MustParse("22fc17a6-40dd-4d68-a95b-ec94e7572a09")),
		Owners:            OwnerPartyList("data_owner"),
		DataAccess:        []string{"data_accessor"},
		ValueOwnerAddress: "value_owner",
	}
	var msg = NewMsgWriteScopeRequest(*scope, []string{}, 0)

	require.Equal(t, sdk.MsgTypeURL(msg), "/provenance.metadata.v1.MsgWriteScopeRequest")

	expectedJSON := "{" +
		"\"scope\":{" +
		"\"scope_id\":\"scope1qzxcpvj6czy5g354dews3nlruxjsahhnsp\"," +
		"\"specification_id\":\"scopespec1qs30c9axgrw5669ft0kffe6h9gysfe58v3\"," +
		"\"owners\":[{\"address\":\"data_owner\",\"role\":\"PARTY_TYPE_OWNER\",\"optional\":false}]," +
		"\"data_access\":[\"data_accessor\"]," +
		"\"value_owner_address\":\"value_owner\"," +
		"\"require_party_rollup\":false" +
		"}," +
		"\"signers\":[]," +
		"\"scope_uuid\":\"\"," +
		"\"spec_uuid\":\"\"," +
		"\"usd_mills\":\"0\"" +
		"}"

	expectedYaml := `scope:
  data_access:
  - data_accessor
  owners:
  - address: data_owner
    optional: false
    role: PARTY_TYPE_OWNER
  require_party_rollup: false
  scope_id: scope1qzxcpvj6czy5g354dews3nlruxjsahhnsp
  specification_id: scopespec1qs30c9axgrw5669ft0kffe6h9gysfe58v3
  value_owner_address: value_owner
scope_uuid: ""
signers: []
spec_uuid: ""
usd_mills: "0"
`
	jsonBZ, err := GetCdc(t).MarshalJSON(msg)
	require.NoError(t, err, "MarshalJSON(msg)")
	assert.Equal(t, expectedJSON, string(jsonBZ), "scope as json")

	yamlBZ, err := yaml.JSONToYAML(jsonBZ)
	require.NoError(t, err, "yaml.JSONToYAML(jsonBZ)")
	assert.Equal(t, expectedYaml, string(yamlBZ), "scope as yaml")
}

func TestWriteScopeValidation(t *testing.T) {
	var scope = &Scope{
		ScopeId:           ScopeMetadataAddress(uuid.MustParse("8d80b25a-c089-4446-956e-5d08cfe3e1a5")),
		SpecificationId:   ScopeSpecMetadataAddress(uuid.MustParse("22fc17a6-40dd-4d68-a95b-ec94e7572a09")),
		Owners:            OwnerPartyList("data_owner"),
		DataAccess:        []string{"data_accessor"},
		ValueOwnerAddress: "value_owner",
	}
	var msg = NewMsgWriteScopeRequest(*scope, []string{"invalid"}, 0)
	err := msg.ValidateBasic()
	require.EqualError(t, err, "invalid scope owners: invalid party address [data_owner]: decoding bech32 failed: invalid separator index -1")

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
		Owners:          OwnerPartyList("cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"),
		DataAccess:      []string{},
	}
	msg.Signers = []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}
	err = msg.Scope.ValidateBasic()
	require.NoError(t, err, "valid add scope request")
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

	require.Equal(t, "/provenance.metadata.v1.MsgBindOSLocatorRequest", sdk.MsgTypeURL(bindRequestMsg))

	bz, _ := GetCdc(t).MarshalJSON(bindRequestMsg)
	require.Equal(t, "{\"locator\":{\"owner\":\"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck\",\"locator_uri\":\"http://foo.com\",\"encryption_key\":\"\"}}", string(bz))
}

func TestModifyOSLocator(t *testing.T) {
	var modifyRequest = NewMsgModifyOSLocatorRequest(ObjectStoreLocator{Owner: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", LocatorUri: "http://foo.com"})

	err := modifyRequest.ValidateBasic()
	require.NoError(t, err)

	require.Equal(t, "/provenance.metadata.v1.MsgModifyOSLocatorRequest", sdk.MsgTypeURL(modifyRequest))

	bz, _ := GetCdc(t).MarshalJSON(modifyRequest)
	require.Equal(t, "{\"locator\":{\"owner\":\"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck\",\"locator_uri\":\"http://foo.com\",\"encryption_key\":\"\"}}", string(bz))
}

func TestDeleteOSLocator(t *testing.T) {
	var deleteRequest = NewMsgDeleteOSLocatorRequest(ObjectStoreLocator{Owner: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", LocatorUri: "http://foo.com"})

	err := deleteRequest.ValidateBasic()
	require.NoError(t, err)

	require.Equal(t, "/provenance.metadata.v1.MsgDeleteOSLocatorRequest", sdk.MsgTypeURL(deleteRequest))

	bz, _ := GetCdc(t).MarshalJSON(deleteRequest)
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

func TestMsgSetAccountDataRequest_ValidateBasic(t *testing.T) {
	addr1 := sdk.AccAddress("addr1_______________").String()
	addr2 := sdk.AccAddress("addr2_______________").String()

	uuid1 := uuid.MustParse("D0FE5658-1A5A-4428-BBEC-7034476C990B")
	uuid2 := uuid.MustParse("C132796F-B2C1-4804-A77E-DD34F46E22F4")
	recordName := "somerecord"

	errNotScopeID := "invalid metadata address: only scope ids are supported"

	msgWAddr := func(addr MetadataAddress) MsgSetAccountDataRequest {
		return MsgSetAccountDataRequest{
			MetadataAddr: addr,
			Value:        "Some test value.",
			Signers:      []string{addr1},
		}
	}

	msgWSigners := func(signers ...string) MsgSetAccountDataRequest {
		return MsgSetAccountDataRequest{
			MetadataAddr: ScopeMetadataAddress(uuid1),
			Value:        "Some other test value.",
			Signers:      signers,
		}
	}

	tests := []struct {
		name string
		msg  MsgSetAccountDataRequest
		exp  string
	}{
		{name: "nil metadata address", msg: msgWAddr(nil), exp: "invalid metadata address: address is empty"},
		{name: "empty metadata address", msg: msgWAddr(MetadataAddress{}), exp: "invalid metadata address: address is empty"},
		{name: "scope id", msg: msgWAddr(ScopeMetadataAddress(uuid1))},
		{name: "session id", msg: msgWAddr(SessionMetadataAddress(uuid1, uuid2)), exp: errNotScopeID},
		{name: "record id", msg: msgWAddr(RecordMetadataAddress(uuid1, recordName)), exp: errNotScopeID},
		{name: "scope spec id", msg: msgWAddr(ScopeSpecMetadataAddress(uuid1)), exp: errNotScopeID},
		{name: "contract spec id", msg: msgWAddr(ContractSpecMetadataAddress(uuid1)), exp: errNotScopeID},
		{name: "record spec id", msg: msgWAddr(RecordSpecMetadataAddress(uuid1, recordName)), exp: errNotScopeID},
		{name: "no signers", msg: msgWSigners(), exp: "at least one signer is required"},
		{name: "one signer", msg: msgWSigners(addr1)},
		{name: "two signers", msg: msgWSigners(addr1, addr2)},
		{
			name: "empty value",
			msg: MsgSetAccountDataRequest{
				MetadataAddr: ScopeMetadataAddress(uuid1),
				Value:        "",
				Signers:      []string{addr1},
			},
		},
		{
			name: "super long value",
			msg: MsgSetAccountDataRequest{
				MetadataAddr: ScopeMetadataAddress(uuid1),
				Value:        strings.Repeat("long-", 10000),
				Signers:      []string{addr1},
			},
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

// TestPrintMessageTypeStrings just prints out all the MsgTypeURLs.
// The output can be copy/pasted into the const area in msgs.go
func TestPrintMessageTypeStrings(t *testing.T) {
	var sb strings.Builder
	sb.WriteString("const (\n")
	for _, msg := range AllRequestMsgs {
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
	// This test is mostly just a demonstration that the entries in AllRequestMsgs can be
	// defined using the (*MsgWriteScopeRequest)(nil) pattern instead of &MsgWriteScopeRequest{}.
	// That's why this is in msgs_test.go instead of codec_test.go.

	registry := cdctypes.NewInterfaceRegistry()
	RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	for _, msg := range AllRequestMsgs {
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

func TestMsgAddNetAssetValueValidateBasic(t *testing.T) {
	addr := sdk.AccAddress("addr________________").String()
	scopeID := "scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel"
	sessionID := "session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr"
	netAssetValue1 := NetAssetValue{Price: sdk.NewInt64Coin("jackthecat", 100)}
	netAssetValue2 := NetAssetValue{Price: sdk.NewInt64Coin("hotdog", 100)}
	invalidNetAssetValue2 := NetAssetValue{Price: sdk.NewInt64Coin("hotdog", 100), UpdatedBlockHeight: 1}

	tests := []struct {
		name   string
		msg    MsgAddNetAssetValuesRequest
		expErr string
	}{
		{
			name: "should succeed",
			msg:  MsgAddNetAssetValuesRequest{ScopeId: scopeID, NetAssetValues: []NetAssetValue{netAssetValue1, netAssetValue2}, Signers: []string{addr}},
		},
		{
			name:   "block height is set",
			msg:    MsgAddNetAssetValuesRequest{ScopeId: scopeID, NetAssetValues: []NetAssetValue{invalidNetAssetValue2}, Signers: []string{addr}},
			expErr: "scope net asset value must not have update height set",
		},
		{
			name:   "duplicate net asset values",
			msg:    MsgAddNetAssetValuesRequest{ScopeId: scopeID, NetAssetValues: []NetAssetValue{netAssetValue1, netAssetValue2, netAssetValue2}, Signers: []string{addr}},
			expErr: "list of net asset values contains duplicates",
		},
		{
			name:   "incorrect meta address",
			msg:    MsgAddNetAssetValuesRequest{ScopeId: "", NetAssetValues: []NetAssetValue{netAssetValue1, netAssetValue2, netAssetValue2}, Signers: []string{addr}},
			expErr: `invalid metadata address "": empty address string is not allowed`,
		},
		{
			name:   "not scope meta address",
			msg:    MsgAddNetAssetValuesRequest{ScopeId: sessionID, NetAssetValues: []NetAssetValue{netAssetValue1, netAssetValue2, netAssetValue2}, Signers: []string{addr}},
			expErr: "metadata address is not scope address: session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr",
		},
		{
			name:   "invalid administrator address",
			msg:    MsgAddNetAssetValuesRequest{ScopeId: scopeID, NetAssetValues: []NetAssetValue{netAssetValue1, netAssetValue2}, Signers: []string{"invalid"}},
			expErr: "decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			name:   "empty net asset list",
			msg:    MsgAddNetAssetValuesRequest{ScopeId: scopeID, NetAssetValues: []NetAssetValue{}, Signers: []string{addr}},
			expErr: "net asset value list cannot be empty",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if len(tc.expErr) > 0 {
				require.EqualErrorf(t, err, tc.expErr, "ValidateBasic error")
			} else {
				require.NoError(t, err, "ValidateBasic error")
			}
		})
	}
}
