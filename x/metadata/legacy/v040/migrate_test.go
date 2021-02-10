package v040_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"

	simapp "github.com/provenance-io/provenance/app"
	v039metadata "github.com/provenance-io/provenance/x/metadata/legacy/v039"
	v040metadata "github.com/provenance-io/provenance/x/metadata/legacy/v040"
)

func TestMigrate(t *testing.T) {

	encodingConfig := simapp.MakeEncodingConfig()
	clientCtx := client.Context{}.
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithJSONMarshaler(encodingConfig.Marshaler)

	gs := v039metadata.GenesisState{
		// TODO build this out with test case data
		ScopeRecords:   v039metadata.DefaultGenesisState().ScopeRecords,
		Specifications: v039metadata.DefaultGenesisState().Specifications,
	}

	migrated := v040metadata.Migrate(gs)

	expected := `{
  "group_specifications": [],
  "groups": [],
  "params": {},
  "records": [],
  "scope_specifications": [],
  "scopes": []
}`
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
