package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
)

func TestReadScopeNAVs(t *testing.T) {
	tests := []struct {
		fileName string
		expCount int
		expFirst ScopeNAV
		expLast  ScopeNAV
	}{
		{
			fileName: umberTestnetScopeNAVsFN,
			expCount: 1101,
			expFirst: ScopeNAV{
				ScopeUUID:     "2f389a9f-873d-4920-85a6-7734f27e1738",
				NetAssetValue: metadatatypes.NewNetAssetValue(sdk.NewInt64Coin(metadatatypes.UsdDenom, 398820670)),
				Height:        23056719,
			},
			expLast: ScopeNAV{
				ScopeUUID:     "65939db0-6d7a-42ef-9443-378304d33225",
				NetAssetValue: metadatatypes.NewNetAssetValue(sdk.NewInt64Coin(metadatatypes.UsdDenom, 93661920)),
				Height:        23056719,
			},
		},
	}

	assertEqualEntry := func(t *testing.T, expected, actual ScopeNAV, msg string, args ...interface{}) bool {
		t.Helper()
		if assert.Equalf(t, expected, actual, msg, args...) {
			return true
		}

		assert.Equalf(t, expected.ScopeUUID, actual.ScopeUUID, msg+" ScopeUUID", args...)
		if !assert.Equalf(t, expected.NetAssetValue, actual.NetAssetValue, msg+" NetAssetValue", args...) {
			assert.Equalf(t, expected.NetAssetValue.Price.String(), actual.NetAssetValue.Price.String(), msg+" NetAssetValue.Price", args...)
			assert.Equalf(t, int(expected.NetAssetValue.UpdatedBlockHeight), int(actual.NetAssetValue.UpdatedBlockHeight), msg+" NetAssetValue.UpdatedBlockHeight", args...)
		}
		assert.Equalf(t, expected.Height, actual.Height, msg+" Height", args...)

		return false
	}

	for _, tc := range tests {
		t.Run(tc.fileName, func(t *testing.T) {
			assets, err := ReadScopeNAVs(tc.fileName)
			require.NoError(t, err, "Failed to read net asset values")
			assert.Len(t, assets, tc.expCount, "The number of assets should be 1101")

			assertEqualEntry(t, tc.expFirst, assets[0], "First entry")
			assertEqualEntry(t, tc.expLast, assets[len(assets)-1], "Last entry")
		})
	}
}

func TestParseValueToUsdMills(t *testing.T) {
	tests := []struct {
		input          string
		expectedOutput int64
		expectError    bool
	}{
		{"1.24", 1240, false},
		{"0.99", 990, false},
		{"100.5", 100500, false},
		{"100.3456", 100345, false},
		{"172755", 172755000, false},
		{"abc", 0, true},
		{"", 0, true},
	}

	for _, tc := range tests {
		name := tc.input
		if len(name) == 0 {
			name = "empty"
		}
		t.Run(name, func(t *testing.T) {
			output, err := parseValueToUsdMills(tc.input)
			if tc.expectError {
				assert.Error(t, err, "parseValueToUsdMills error")
			} else {
				assert.NoError(t, err, "parseValueToUsdMills error")
			}
			assert.Equal(t, tc.expectedOutput, output, "parseValueToUsdMills output")
		})
	}
}
