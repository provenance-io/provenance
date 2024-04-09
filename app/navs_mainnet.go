package app

import (
	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

// GetPioMainnet1DenomToNav are net asset values for the pio-mainnet-1 taken at blockheight 13631650
// Source: https://figure.tech/service-pricing-engine/external/api/v1/pricing/marker/new?time=2023-11-07T17:59:59.999722Z
// NOTE: These should not be ran against any other network but pio-mainnet-1
// TODO: Remove with the saffron handlers.
func GetPioMainnet1DenomToNav() map[string]markertypes.NetAssetValue {
	return map[string]markertypes.NetAssetValue{}
}
