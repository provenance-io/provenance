package v040_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"

	simapp "github.com/provenance-io/provenance/app"
	v039 "github.com/provenance-io/provenance/x/metadata/legacy/v039"
	v039metadata "github.com/provenance-io/provenance/x/metadata/legacy/v039"
	v040metadata "github.com/provenance-io/provenance/x/metadata/legacy/v040"
)

var (
	exampleV39Spec = `{
  "consideration_specs": [
    {
      "func_name": "additionalParties",
      "input_specs": [
        {
          "name": "perform_input_checks",
          "resource_location": {
            "classname": "io.provenance.Common$BooleanResult",
            "ref": {
              "hash": "Adv+huolGTKofYCR0dw5GHm/R7sUWOwF32XR8r8r9kDy4il5U/LApxOWYHb05jhK4+eY4YzRMRiWcxU3Lx0+Mw=="
            }
          },
          "type": 1
        },
        {
          "name": "additional_parties",
          "resource_location": {
            "classname": "io.provenance.loan.LoanProtos$PartiesList",
            "ref": {
              "hash": "Adv+huolGTKofYCR0dw5GHm/R7sUWOwF32XR8r8r9kDy4il5U/LApxOWYHb05jhK4+eY4YzRMRiWcxU3Lx0+Mw=="
            }
          },
          "type": 1
        }
      ],
      "output_spec": {
        "spec": {
          "name": "additional_parties",
          "resource_location": {
            "classname": "io.provenance.loan.LoanProtos$PartiesList",
            "ref": {
              "hash": "Adv+huolGTKofYCR0dw5GHm/R7sUWOwF32XR8r8r9kDy4il5U/LApxOWYHb05jhK4+eY4YzRMRiWcxU3Lx0+Mw=="
            }
          },
          "type": 1
        }
      },
      "responsible_party": 1
    },
    {
      "func_name": "documents",
      "input_specs": [
        {
          "name": "perform_input_checks",
          "resource_location": {
            "classname": "io.provenance.Common$BooleanResult",
            "ref": {
              "hash": "Adv+huolGTKofYCR0dw5GHm/R7sUWOwF32XR8r8r9kDy4il5U/LApxOWYHb05jhK4+eY4YzRMRiWcxU3Lx0+Mw=="
            }
          },
          "type": 1
        },
        {
          "name": "documents",
          "resource_location": {
            "classname": "io.provenance.common.DocumentProtos$DocumentList",
            "ref": {
              "hash": "Adv+huolGTKofYCR0dw5GHm/R7sUWOwF32XR8r8r9kDy4il5U/LApxOWYHb05jhK4+eY4YzRMRiWcxU3Lx0+Mw=="
            }
          },
          "type": 1
        }
      ],
      "output_spec": {
        "spec": {
          "name": "documents",
          "resource_location": {
            "classname": "io.provenance.common.DocumentProtos$DocumentList",
            "ref": {
              "hash": "Adv+huolGTKofYCR0dw5GHm/R7sUWOwF32XR8r8r9kDy4il5U/LApxOWYHb05jhK4+eY4YzRMRiWcxU3Lx0+Mw=="
            }
          },
          "type": 1
        }
      },
      "responsible_party": 1
    }
  ],
  "definition": {
    "name": "ExampleContract",
    "resource_location": {
      "classname": "io.provenance.contracts.ExampleContract",
      "ref": {
        "hash": "E36eeTUk8GYXGXjIbZTm4s/Dw3G1e42SinH1195t4ekgcXXPhfIpfQaEJ21PTzKhdv6JjhzQJ2kAJXK+TRXmeQ=="
      }
    },
    "type": 2
  },
  "input_specs": [
    {
      "name": "additional_parties",
      "resource_location": {
        "classname": "io.provenance.loan.LoanProtos$PartiesList",
        "ref": {
          "hash": "Adv+huolGTKofYCR0dw5GHm/R7sUWOwF32XR8r8r9kDy4il5U/LApxOWYHb05jhK4+eY4YzRMRiWcxU3Lx0+Mw=="
        }
      },
      "type": 2
    },
    {
      "name": "documents",
      "resource_location": {
        "classname": "io.provenance.common.DocumentProtos$DocumentList",
        "ref": {
          "hash": "Adv+huolGTKofYCR0dw5GHm/R7sUWOwF32XR8r8r9kDy4il5U/LApxOWYHb05jhK4+eY4YzRMRiWcxU3Lx0+Mw=="
        }
      },
      "type": 2
    }
  ],
  "parties_involved": [
    1
  ]
}`
)

func TestMigrate(t *testing.T) {

	encodingConfig := simapp.MakeEncodingConfig()
	clientCtx := client.Context{}.
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithJSONMarshaler(encodingConfig.Marshaler)

	var v39ContractSpec v039metadata.ContractSpec
	clientCtx.JSONMarshaler.UnmarshalJSON([]byte(exampleV39Spec), &v39ContractSpec)

	gs := v039metadata.GenesisState{
		// TODO build this out with test case data
		ScopeRecords:   v039metadata.DefaultGenesisState().ScopeRecords,
		Specifications: []v039.ContractSpec{v39ContractSpec},
	}

	migrated := v040metadata.Migrate(gs)

	expected := `{
  "contract_specifications": [
    {
      "class_name": "io.provenance.contracts.ExampleContract",
      "description": {
        "description": "io.provenance.contracts.ExampleContract",
        "icon_url": "",
        "name": "ExampleContract",
        "website_url": ""
      },
      "hash": "E36eeTUk8GYXGXjIbZTm4s/Dw3G1e42SinH1195t4ekgcXXPhfIpfQaEJ21PTzKhdv6JjhzQJ2kAJXK+TRXmeQ==",
      "owner_addresses": [],
      "parties_involved": [
        "PARTY_TYPE_ORIGINATOR"
      ],
      "specification_id": "contractspec1qvfha8nex5j0qeshr9uvsmv5um3q7sghss"
    }
  ],
  "groups": [],
  "params": {},
  "record_specifications": [
    {
      "inputs": [
        {
          "hash": "Adv+huolGTKofYCR0dw5GHm/R7sUWOwF32XR8r8r9kDy4il5U/LApxOWYHb05jhK4+eY4YzRMRiWcxU3Lx0+Mw==",
          "name": "perform_input_checks",
          "type_name": "io.provenance.Common$BooleanResult"
        },
        {
          "hash": "Adv+huolGTKofYCR0dw5GHm/R7sUWOwF32XR8r8r9kDy4il5U/LApxOWYHb05jhK4+eY4YzRMRiWcxU3Lx0+Mw==",
          "name": "additional_parties",
          "type_name": "io.provenance.loan.LoanProtos$PartiesList"
        }
      ],
      "name": "additional_parties",
      "responsible_parties": [
        "PARTY_TYPE_ORIGINATOR"
      ],
      "result_type": "DEFINITION_TYPE_PROPOSED",
      "specification_id": "contractspec1qvfha8nex5j0qeshr9uvsmv5um3q7sghss",
      "type_name": "io.provenance.loan.LoanProtos$PartiesList"
    },
    {
      "inputs": [
        {
          "hash": "Adv+huolGTKofYCR0dw5GHm/R7sUWOwF32XR8r8r9kDy4il5U/LApxOWYHb05jhK4+eY4YzRMRiWcxU3Lx0+Mw==",
          "name": "perform_input_checks",
          "type_name": "io.provenance.Common$BooleanResult"
        },
        {
          "hash": "Adv+huolGTKofYCR0dw5GHm/R7sUWOwF32XR8r8r9kDy4il5U/LApxOWYHb05jhK4+eY4YzRMRiWcxU3Lx0+Mw==",
          "name": "documents",
          "type_name": "io.provenance.common.DocumentProtos$DocumentList"
        }
      ],
      "name": "documents",
      "responsible_parties": [
        "PARTY_TYPE_ORIGINATOR"
      ],
      "result_type": "DEFINITION_TYPE_PROPOSED",
      "specification_id": "contractspec1qvfha8nex5j0qeshr9uvsmv5um3q7sghss",
      "type_name": "io.provenance.common.DocumentProtos$DocumentList"
    }
  ],
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
