package app

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
	"github.com/stretchr/testify/assert"
)

func TestReadNetAssetValues(t *testing.T) {
	fileName := "upgrade_files/testnet_scope_navs.csv"

	assets, err := ReadNetAssetValues(fileName)
	assert.NoError(t, err, "Failed to read net asset values")
	assert.Len(t, assets, 1101, "The number of assets should be 1101")

	expectedFirst := NetAssetValueWithHeight{
		ScopeUUID:     "2f389a9f-873d-4920-85a6-7734f27e1738",
		NetAssetValue: metadatatypes.NewNetAssetValue(sdk.NewInt64Coin(metadatatypes.UsdDenom, 398820670)),
		Height:        406,
	}
	assert.Equal(t, expectedFirst.ScopeUUID, assets[0].ScopeUUID, "The first ScopeUUID should match")
	assert.True(t, assets[0].NetAssetValue.Price.Equal(expectedFirst.NetAssetValue.Price), "The first NetAssetValue should match")
	assert.Equal(t, expectedFirst.Height, assets[0].Height, "The first Height should match")

	expectedLast := NetAssetValueWithHeight{
		ScopeUUID:     "65939db0-6d7a-42ef-9443-378304d33225",
		NetAssetValue: metadatatypes.NewNetAssetValue(sdk.NewInt64Coin(metadatatypes.UsdDenom, 93661920)),
		Height:        406,
	}
	assert.Equal(t, expectedLast.ScopeUUID, assets[len(assets)-1].ScopeUUID, "The last ScopeUUID should match")
	assert.True(t, assets[len(assets)-1].NetAssetValue.Price.Equal(expectedLast.NetAssetValue.Price), "The last NetAssetValue should match")
	assert.Equal(t, expectedLast.Height, assets[len(assets)-1].Height, "The last Height should match")
}
