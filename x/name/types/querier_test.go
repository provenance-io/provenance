package types

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestQuerierNameResult(t *testing.T) {
	addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	nr := NewNameRecord("example.name", addr, true)
	// check that result can wrap a name record
	qnr := QueryNameResult(nr)
	require.NotNil(t, qnr)
	require.Equal(t, qnr.Name, "example.name")
	require.Equal(t, qnr.Address, addr.String())
	require.Equal(t, qnr.Restricted, true)
	require.Equal(t, fmt.Sprintf("example.name->%s", addr.String()), qnr.String())
}
