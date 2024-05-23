package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/testutil"

	. "github.com/provenance-io/provenance/x/attribute/types"
)

var (
	priv1 = ed25519.GenPrivKey()
	priv2 = ed25519.GenPrivKey()
	addrs = []sdk.AccAddress{
		sdk.AccAddress(priv1.PubKey().Address()),
		sdk.AccAddress(priv2.PubKey().Address()),
	}
)

func TestAllMsgsGetSigners(t *testing.T) {
	msgMakers := []testutil.MsgMaker{
		func(signer string) sdk.Msg { return &MsgAddAttributeRequest{Owner: signer} },
		func(signer string) sdk.Msg { return &MsgUpdateAttributeRequest{Owner: signer} },
		func(signer string) sdk.Msg { return &MsgUpdateAttributeExpirationRequest{Owner: signer} },
		func(signer string) sdk.Msg { return &MsgDeleteAttributeRequest{Owner: signer} },
		func(signer string) sdk.Msg { return &MsgDeleteDistinctAttributeRequest{Owner: signer} },
		func(signer string) sdk.Msg { return &MsgSetAccountDataRequest{Account: signer} },
		func(signer string) sdk.Msg { return &MsgUpdateParamsRequest{Authority: signer} },
	}

	testutil.RunGetSignersTests(t, AllRequestMsgs, msgMakers, nil)
}

// test ValidateBasic for TestMsgAddAttribute
func TestMsgAddAttribute(t *testing.T) {
	tests := []struct {
		account            string
		owner              sdk.AccAddress
		name, proposalType string
		proposalValue      string
		expectPass         bool
	}{
		{"", addrs[1], "test", "string", "nil owner", false},
		{addrs[0].String(), nil, "test", "string", "nil account", false},
		{"", nil, "test", "string", "nil owner and account", false},
		{addrs[0].String(), addrs[1], "test", "string", "valid attribute", true},
	}

	for i, tc := range tests {
		at, err := AttributeTypeFromString(tc.proposalType)
		require.NoError(t, err)
		msg := NewMsgAddAttributeRequest(
			tc.account,
			tc.owner,
			tc.name,
			at,
			[]byte(tc.proposalValue),
		)

		if tc.expectPass {
			require.NoError(t, msg.ValidateBasic(), "test: %v", i)
		} else {
			require.Error(t, msg.ValidateBasic(), "test: %v", i)
		}
	}
}

// test ValidateBasic for TestMsgUpdateAttribute
func TestMsgUpdateAttribute(t *testing.T) {
	tests := []struct {
		account       string
		owner         sdk.AccAddress
		name          string
		originalValue []byte
		originalType  AttributeType
		updateValue   []byte
		updateType    AttributeType
		expectPass    bool
		expectedError string
	}{
		{addrs[0].String(), addrs[1], "example", []byte("original"), AttributeType_String, []byte("update"), AttributeType_Bytes, true, ""},
		{"", addrs[1], "example", []byte("original"), AttributeType_String, []byte("update"), AttributeType_Bytes, false, ""},
		{addrs[0].String(), nil, "example", []byte(""), AttributeType_String, []byte("update"), AttributeType_Bytes, false, ""},
	}

	for _, tc := range tests {
		msg := NewMsgUpdateAttributeRequest(tc.account, tc.owner, tc.name, tc.originalValue, tc.updateValue, tc.originalType, tc.updateType)

		if tc.expectPass {
			require.NoError(t, msg.ValidateBasic(), "test: %v", tc)
		} else {
			require.Error(t, msg.ValidateBasic(), "test: %v", tc)
		}
	}
}

// test ValidateBasic for TestMsgDeleteDistinctAttribute
func TestMsgDeleteDistinctAttribute(t *testing.T) {
	tests := []struct {
		account    string
		owner      sdk.AccAddress
		name       string
		value      []byte
		attrType   AttributeType
		expectPass bool
	}{
		{addrs[0].String(), addrs[1], "example", []byte("original"), AttributeType_String, true},
		{"", addrs[1], "example", []byte("original"), AttributeType_String, false},
		{addrs[0].String(), nil, "example", []byte(""), AttributeType_String, false},
	}

	for _, tc := range tests {
		msg := NewMsgDeleteDistinctAttributeRequest(tc.account, tc.owner, tc.name, tc.value)

		if tc.expectPass {
			require.NoError(t, msg.ValidateBasic(), "test: %v", tc)
		} else {
			require.Error(t, msg.ValidateBasic(), "test: %v", tc)
		}
	}
}

func TestMsgSetAccountDataRequest_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgSetAccountDataRequest
		exp  string
	}{
		{
			name: "control",
			msg:  MsgSetAccountDataRequest{Account: sdk.AccAddress("control").String(), Value: "some value"},
			exp:  "",
		},
		{
			name: "bad account",
			msg:  MsgSetAccountDataRequest{Account: "notabech32", Value: "some value"},
			exp:  "invalid account: decoding bech32 failed: invalid separator index -1",
		},
		{
			name: "no value",
			msg:  MsgSetAccountDataRequest{Account: sdk.AccAddress("no value").String(), Value: ""},
			exp:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if len(tc.exp) > 0 {
				assert.EqualError(t, err, tc.exp, "ValidateBasic error")
			} else {
				assert.NoError(t, err, "ValidateBasic error")
			}
		})
	}
}

func TestMsgUpdateParamsRequest(t *testing.T) {
	tests := []struct {
		name           string
		authority      string
		maxValueLength uint32
		expectPass     bool
		expectedError  string
	}{
		{
			name:           "valid authority",
			authority:      sdk.AccAddress(priv1.PubKey().Address()).String(),
			maxValueLength: 100,
			expectPass:     true,
		},
		{
			name:           "invalid authority",
			authority:      "invalid-authority",
			maxValueLength: 100,
			expectPass:     false,
			expectedError:  "invalid authority: decoding bech32 failed: invalid separator index -1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := NewMsgUpdateParamsRequest(tc.authority, tc.maxValueLength)

			err := msg.ValidateBasic()
			if tc.expectPass {
				require.NoError(t, err, "test: %v", tc.name)
			} else {
				require.Error(t, err, "test: %v", tc.name)
				assert.EqualError(t, err, tc.expectedError, "test: %v", tc.name)
			}
		})
	}
}
