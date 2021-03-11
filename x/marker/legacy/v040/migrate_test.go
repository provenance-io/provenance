package v040_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/legacy/v039"

	v039 "github.com/provenance-io/provenance/x/marker/legacy/v039"
	v040 "github.com/provenance-io/provenance/x/marker/legacy/v040"
)

func TestMigrate(t *testing.T) {
	encodingConfig := simapp.MakeTestEncodingConfig()
	clientCtx := client.Context{}.
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithJSONMarshaler(encodingConfig.Marshaler)

	addr1, err := sdk.AccAddressFromBech32("cosmos16xyempempp92x9hyzz9wrgf94r6j9h5f06pxxv")
	require.NoError(t, err)

	gs := v039.GenesisState{
		Markers: []v039.MarkerAccount{
			{
				BaseAccount: &authtypes.BaseAccount{
					Address:       addr1,
					Coins:         sdk.NewCoins(sdk.NewCoin("test", sdk.OneInt())),
					AccountNumber: 5,
					Sequence:      4,
				},
				Manager:        addr1,
				Status:         v039.MustGetMarkerStatus("active"),
				Denom:          "hotdog",
				Supply:         sdk.OneInt(),
				MarkerType:     "COIN",
				AccessControls: []v039.AccessGrant{{Address: addr1, Permissions: []string{"mint", "burn", "grant"}}},
			},
		},
	}

	migrated := v040.Migrate(gs)
	expected := fmt.Sprintf(`{
  "markers": [
    {
      "access_control": [
        {
          "address": "%s",
          "permissions": [
            "ACCESS_MINT",
            "ACCESS_BURN",
            "ACCESS_ADMIN"
          ]
        }
      ],
      "allow_governance_control": false,
      "base_account": {
        "account_number": "5",
        "address": "%s",
        "pub_key": null,
        "sequence": "4"
      },
      "denom": "hotdog",
      "manager": "%s",
      "marker_type": "MARKER_TYPE_COIN",
      "status": "MARKER_STATUS_ACTIVE",
      "supply": "1",
      "supply_fixed": false
    }
  ],
  "params": {
    "enable_governance": true,
    "max_total_supply": "100000000000",
    "unrestricted_denom_regex": "[a-zA-Z][a-zA-Z0-9/]{2,64}"
  }
}`, addr1.String(), addr1.String(), addr1.String())

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
