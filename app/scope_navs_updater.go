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

func ReadNetAssetValues(fileName string) ([]NetAssetValueWithHeight, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// skips the header line
	if _, err := reader.Read(); err != nil {
		return nil, err
	}

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var assets []NetAssetValueWithHeight
	for _, record := range records {
		if len(record) < 3 {
			continue
		}

		scopeUUID := record[0]

		navValue, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			return nil, err
		}
		navValue *= 1000

		heightIndex := len(record) - 1
		height, err := strconv.ParseUint(record[heightIndex], 10, 64)
		if err != nil {
			return nil, err
		}

		price := sdk.NewInt64Coin(metadatatypes.UsdDenom, int64(navValue))

		asset := NetAssetValueWithHeight{
			ScopeUUID:     scopeUUID,
			NetAssetValue: metadatatypes.NewNetAssetValue(price),
			Height:        height,
		}

		assets = append(assets, asset)
	}

	return assets, nil
}
