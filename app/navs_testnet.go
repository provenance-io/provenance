package app

import (
	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

// GetPioTestnet1DenomToNav are net asset values for the pio-testnet-1
// Source: https://test.figure.tech/service-pricing-engine/external/api/v1/pricing/marker/new?time=2024-03-19T00:00:00.000-05:00
// NOTE: These should not be ran against any other network but pio-testnet-1
func GetPioTestnet1DenomToNav() map[string]markertypes.NetAssetValue {
	return map[string]markertypes.NetAssetValue{}
}
