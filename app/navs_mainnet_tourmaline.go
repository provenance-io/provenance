package app

import (
	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

type NetAssetValueWithHeight struct {
	Denom         string
	NetAssetValue markertypes.NetAssetValue
	Height        uint64
}

// GetPioMainnet1NavsTourmaline are net asset values for the pio-mainnet-1
// https://figure.tech/service-pricing-engine/external/api/v1/pricing/marker/new?time=2024-03-21T00:00:00.000000Z
// NOTE: These should not be ran against any other network but pio-mainnet-1 and Tourmaline v1.18.0 upgrade
// TODO: Remove with the tourmaline handlers.
func GetPioMainnet1NavsTourmaline() []NetAssetValueWithHeight {
	return []NetAssetValueWithHeight{}
}
