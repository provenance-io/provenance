package types

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/provenance-io/provenance/x/registry"
)

const (
	ledgerKeyHrp = "ledger"
)

// Combine the asset class id and nft id into a bech32 string.
// Using bech32 here just allows us a readable identifier for the ledger.
func (lk LedgerKey) String() string {
	// Use null byte as delimiter
	joined := lk.AssetClassId + "\x00" + lk.NftId

	b32, err := bech32.ConvertAndEncode(ledgerKeyHrp, []byte(joined))
	if err != nil {
		panic(err)
	}

	return b32
}

// Convert a bech32 string to a LedgerKey.
func StringToLedgerKey(s string) (*LedgerKey, error) {
	hrp, b, err := bech32.DecodeAndConvert(s)
	if err != nil {
		return nil, err
	}

	if hrp != ledgerKeyHrp {
		return nil, fmt.Errorf("invalid hrp: %s", hrp)
	}

	// Split by null byte delimiter
	parts := strings.Split(string(b), "\x00")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid key: %s", s)
	}

	return &LedgerKey{
		AssetClassId: parts[0],
		NftId:        parts[1],
	}, nil
}

// Implement Compare() for LedgerEntry
func (a *LedgerEntry) Compare(b *LedgerEntry) int {
	// First compare effective date (ISO8601 string)
	if a.EffectiveDate < b.EffectiveDate {
		return -1
	}
	if a.EffectiveDate > b.EffectiveDate {
		return 1
	}

	// Then compare sequence number
	if a.Sequence < b.Sequence {
		return -1
	}
	if a.Sequence > b.Sequence {
		return 1
	}

	// Equal
	return 0
}

func (lk LedgerKey) ToRegistryKey() *registry.RegistryKey {
	return &registry.RegistryKey{
		AssetClassId: lk.AssetClassId,
		NftId:        lk.NftId,
	}
}
