package types

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/types/bech32"
)

const (
	ledgerKeyHrp = "ledger"
)

// Combine the asset class id and nft id into a bech32 string.
// Using bech32 here just allows us a readable identifier for the ledger.
func (lk LedgerKey) String() string {
	joined := strings.Join([]string{lk.AssetClassId, lk.NftId}, ":")

	b32, err := bech32.ConvertAndEncode(ledgerKeyHrp, []byte(joined))
	if err != nil {
		panic(err)
	}

	return b32
}

func StringToLedgerKey(s string) (*LedgerKey, error) {
	hrp, b, err := bech32.DecodeAndConvert(s)
	if err != nil {
		return nil, err
	}

	if hrp != ledgerKeyHrp {
		return nil, fmt.Errorf("invalid hrp: %s", hrp)
	}

	parts := strings.Split(string(b), ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid key: %s", s)
	}

	return &LedgerKey{
		AssetClassId: parts[0],
		NftId:        parts[1],
	}, nil
}
