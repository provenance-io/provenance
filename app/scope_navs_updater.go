package app

import (
	"encoding/csv"
	"os"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
)

type NetAssetValueWithHeight struct {
	ScopeUUID     string
	NetAssetValue metadatatypes.NetAssetValue
	Height        uint64
}

// parseValueToUsdMills parses and converts cents amount into usd mills as int64 $1.24 = 1240
func parseValueToUsdMills(navStr string) (int64, error) {
	navValue, err := strconv.ParseFloat(navStr, 64)
	if err != nil {
		return 0, err
	}
	return int64(navValue * 1000), nil
}

// ReadNetAssetValues reads a CSV file and parses its contents into a slice of NetAssetValueWithHeight
func ReadNetAssetValues(fileName string) ([]NetAssetValueWithHeight, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Skip the header line
	if _, err := reader.Read(); err != nil {
		return nil, err
	}

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	assets := make([]NetAssetValueWithHeight, 0, len(records))
	for _, record := range records {
		if len(record) < 3 {
			continue
		}

		scopeUUID := record[0]

		navInt64, err := parseValueToUsdMills(record[1])
		if err != nil {
			return nil, err
		}

		heightIndex := len(record) - 1
		height, err := strconv.ParseUint(record[heightIndex], 10, 64)
		if err != nil {
			return nil, err
		}

		price := sdk.NewInt64Coin(metadatatypes.UsdDenom, navInt64)

		asset := NetAssetValueWithHeight{
			ScopeUUID:     scopeUUID,
			NetAssetValue: metadatatypes.NewNetAssetValue(price),
			Height:        height,
		}

		assets = append(assets, asset)
	}

	return assets, nil
}
