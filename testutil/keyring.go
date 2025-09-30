package testutil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TestKeyringEntry is a keyring.Record plus some extra fields that might be needed for unit tests.
type TestKeyringEntry struct {
	keyring.Record
	AccAddress sdk.AccAddress
	Mnemonic   string
}

// GenerateTestKeyring creates an in-memory keyring and adds count entries to it.
// The names of the entries will be incrementing, starting at "test_key_01".
func GenerateTestKeyring(t *testing.T, count int, codec codec.Codec) ([]TestKeyringEntry, keyring.Keyring) {
	kr := NewTestKeyring(t, codec)
	accounts := GenerateTestAccountsInKeyring(t, kr, count)
	return accounts, kr
}

// NewTestKeyring creates a new in-memory keyring for use in tests.
// Use GenerateTestAccountsInKeyring to add entries to the keyring.
func NewTestKeyring(t *testing.T, codec codec.Codec) keyring.Keyring {
	kr, krErr := keyring.New(t.Name(), keyring.BackendMemory, "", nil, codec)
	require.NoError(t, krErr, "keyring.New(...)")
	return kr
}

// GenerateTestAccountsInKeyring adds count entries to the keyring.
// The names of the entries will be like "test_key_01".
//
// Counting starts at 1 + the number of entries already in the keyring.
// E.g. If there aren't any yet, the first one will be "test_key_01".
// E.g. If there are five already in there, the first one added here will be "test_key_06".
//
// Only newly created entries are returned.
func GenerateTestAccountsInKeyring(t *testing.T, kr keyring.Keyring, count int) []TestKeyringEntry {
	curRecs, curErr := kr.List()
	require.NoError(t, curErr, "listing keyring entries")
	i0 := len(curRecs) + 1

	path := hd.CreateHDPath(118, 0, 0).String()
	accounts := make([]TestKeyringEntry, count)
	for i := range accounts {
		name := fmt.Sprintf("test_key_%02d", i+i0)
		info, mnemonic, err := kr.NewMnemonic(name, keyring.English, path, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
		require.NoError(t, err, "[%d] kr.NewMnemonic(%q, ...)", i, name)
		addr, err := info.GetAddress()
		require.NoError(t, err, "[%d] getting keyring address for %q", i, name)
		accounts[i] = TestKeyringEntry{
			Record:     *info,
			AccAddress: addr,
			Mnemonic:   mnemonic,
		}
	}

	return accounts
}

// GetKeyringEntryAddresses gets the AccAddress of each entry.
func GetKeyringEntryAddresses(entries []TestKeyringEntry) []sdk.AccAddress {
	if len(entries) == 0 {
		return nil
	}
	rv := make([]sdk.AccAddress, len(entries))
	for i, entry := range entries {
		rv[i] = entry.AccAddress
	}
	return rv
}
