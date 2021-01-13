package types

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/stretchr/testify/require"

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
		account, owner     sdk.AccAddress
		name, proposalType string
		proposalValue      string
		expectPass         bool
	}{
		{nil, addrs[1], "test", "string", "nil owner", false},
		{addrs[0], nil, "test", "string", "nil account", false},
		{nil, nil, "test", "string", "nil owner and account", false},
		{addrs[0], addrs[1], "test", "string", "valid attribute", true},
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
