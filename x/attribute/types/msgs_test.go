package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	priv1 = ed25519.GenPrivKey()
	priv2 = ed25519.GenPrivKey()
	addrs = []sdk.AccAddress{
		sdk.AccAddress(priv1.PubKey().Address()),
		sdk.AccAddress(priv2.PubKey().Address()),
	}
)

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

func TestMsgSetAccountDataRequest_GetSigners(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgSetAccountDataRequest
		exp  []sdk.AccAddress
		expP string
	}{
		{
			name: "good account",
			msg:  MsgSetAccountDataRequest{Account: sdk.AccAddress("good account").String()},
			exp:  []sdk.AccAddress{sdk.AccAddress("good account")},
		},
		{
			name: "bad account",
			msg:  MsgSetAccountDataRequest{Account: "bad account"},
			expP: "decoding bech32 failed: invalid character in string: ' '",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var signers []sdk.AccAddress
			testFunc := func() {
				signers = tc.msg.GetSigners()
			}
			if len(tc.expP) > 0 {
				require.PanicsWithError(t, tc.expP, testFunc, "GetSigners")
			} else {
				require.NotPanics(t, testFunc, "GetSigners")
			}
			assert.Equal(t, tc.exp, signers, "GetSigners result")
		})
	}
}
