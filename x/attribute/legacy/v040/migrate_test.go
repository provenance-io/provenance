package v040_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	v039 "github.com/provenance-io/provenance/x/attribute/legacy/v039"
	v040 "github.com/provenance-io/provenance/x/attribute/legacy/v040"
)

func TestMigrate(t *testing.T) {
	encodingConfig := simapp.MakeTestEncodingConfig()
	clientCtx := client.Context{}.
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithJSONCodec(encodingConfig.Marshaler)

	addr1, err := sdk.AccAddressFromBech32("cosmos16xyempempp92x9hyzz9wrgf94r6j9h5f06pxxv")
	require.NoError(t, err)

	gs := v039.GenesisState{
		Attributes: []v039.Attribute{
			{
				Name:    "test",
				Value:   []byte("test-value"),
				Type:    "String",
				Address: addr1.String(),
				Height:  2,
			},
		},
	}

	migrated := v040.Migrate(gs)
	expected := fmt.Sprintf(`{
  "attributes": [
    {
      "address": "%s",
      "attribute_type": "ATTRIBUTE_TYPE_STRING",
      "name": "test",
      "value": "dGVzdC12YWx1ZQ=="
    }
  ],
  "params": {
    "max_value_length": 10000
  }
}`, addr1.String())

	bz, err := clientCtx.JSONCodec.MarshalJSON(migrated)
	require.NoError(t, err)

	// Indent the JSON bz correctly.
	var jsonObj map[string]interface{}
	err = json.Unmarshal(bz, &jsonObj)
	require.NoError(t, err)
	indentedBz, err := json.MarshalIndent(jsonObj, "", "  ")
	require.NoError(t, err)

	require.Equal(t, expected, string(indentedBz))
}
