package v040_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	v039name "github.com/provenance-io/provenance/x/name/legacy/v039"
	v040 "github.com/provenance-io/provenance/x/name/legacy/v040"
)

func TestMigrate(t *testing.T) {
	encodingConfig := simapp.MakeTestEncodingConfig()
	clientCtx := client.Context{}.
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithJSONMarshaler(encodingConfig.Marshaler)

	addr1, err := sdk.AccAddressFromBech32("cosmos16xyempempp92x9hyzz9wrgf94r6j9h5f06pxxv")
	superBadAddr, err := sdk.AccAddressFromBech32("cosmos14n2lmufqep8qmn98yhtmy4uwsd7msmpfkr8vfd")
	name := "mcvluvin"
	restricted := false
	require.NoError(t, err)

	gs := v039name.GenesisState{
		Bindings: []v039name.NameRecord{
			{
				Name:       name,
				Address:    addr1,
				Restricted: restricted,
				Pointer:    superBadAddr,
			},
		},
	}

	migrated := v040.Migrate(gs)
	expected := fmt.Sprintf(`{
  "bindings": [
    {
      "address": "%s",
      "name": "%s",
      "restricted": %s
    }
  ],
  "params": {
    "allow_unrestricted_names": true,
    "max_name_levels": 16,
    "max_segment_length": 32,
    "min_segment_length": 2
  }
}`, addr1.String(), name, fmt.Sprint(restricted))

	bz, err := clientCtx.JSONMarshaler.MarshalJSON(migrated)
	require.NoError(t, err)

	// Indent the JSON bz correctly.
	var jsonObj map[string]interface{}
	err = json.Unmarshal(bz, &jsonObj)
	require.NoError(t, err)
	indentedBz, err := json.MarshalIndent(jsonObj, "", "  ")
	require.NoError(t, err)

	require.Equal(t, expected, string(indentedBz))
}
