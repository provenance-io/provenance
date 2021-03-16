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
	exampleV39Scope = `{
  "last_event": {
    "execution_uuid": {
      "value": "d60098e7-153c-492e-bcb3-37ce293ffb17"
    },
    "group_uuid": {
      "value": "0f2c6db3-8b47-4d98-ae5f-647a7e511545"
    }
  },
  "parties": [
    {
      "address": "GOrX55pQULrlronarLFe8U4U3bc=",
      "signer": {
        "encryption_public_key": {
          "public_key_bytes": "BOMGXlqove6huk+stReazUD43ANdehXQbewNHw/mv8vzRWsrxK/II+0wulZKG08458ykUnhSHHKpiv4EyYT8XzM="
        },
        "signing_public_key": {
          "public_key_bytes": "BOMGXlqove6huk+stReazUD43ANdehXQbewNHw/mv8vzRWsrxK/II+0wulZKG08458ykUnhSHHKpiv4EyYT8XzM="
        }
      },
      "signer_role": 1
    }
  ],
  "record_group": [
    {
      "audit": {
        "created_by": "tp1rr4d0eu62pgt4edw38d2ev27798pfhdhp5ttha",
        "created_date": {
          "seconds": 1608242483
        },
        "version": 1
      },
      "classname": "io.p8e.contracts.origination.ETLTouch",
      "executor": {
        "encryption_public_key": {
          "public_key_bytes": "BOMGXlqove6huk+stReazUD43ANdehXQbewNHw/mv8vzRWsrxK/II+0wulZKG08458ykUnhSHHKpiv4EyYT8XzM="
        }
      },
      "group_uuid": {
        "value": "0f2c6db3-8b47-4d98-ae5f-647a7e511545"
      },
      "parties": [
        {
          "address": "GOrX55pQULrlronarLFe8U4U3bc=",
          "signer": {
            "encryption_public_key": {
              "public_key_bytes": "BOMGXlqove6huk+stReazUD43ANdehXQbewNHw/mv8vzRWsrxK/II+0wulZKG08458ykUnhSHHKpiv4EyYT8XzM="
            },
            "signing_public_key": {
              "public_key_bytes": "BOMGXlqove6huk+stReazUD43ANdehXQbewNHw/mv8vzRWsrxK/II+0wulZKG08458ykUnhSHHKpiv4EyYT8XzM="
            }
          },
          "signer_role": 1
        }
      ],
      "records": [
        {
          "classname": "io.provenance.proto.loan.LoanProtos$Loan",
          "hash": "VQUlaiUsAJB6XVegaZ3vD+vMdqfR+/jHp6UjNSW/RUdbJOTUrKn50/HkDGzQyUmniNYYZ/YxjwrdCkTB9mm6qw==",
          "name": "noop",
          "result": 1,
          "result_hash": "poYoiYr8gi22vyBlo09YkSSnGHRY0jQW9DZAvaTPT5slTbt2SV8KgGSKoK72PYVL/yLrCgnrEDaRn08byB/JHQ==",
          "result_name": "loan"
        }
      ],
      "specification": "2Sl7QeL/Zn37Md9w0Rncl9/qG8UvfLbsxOpwUUUi3pel4P8xa3sdEzSjwVvENpYGwjHY51jpdvlSX7PbntrjOg=="
    }
  ],
  "uuid": {
    "value": "000029c9-8e2a-45d7-95af-b670bb54061a"
  }
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

	var v39Scope v039metadata.Scope
	clientCtx.JSONMarshaler.UnmarshalJSON([]byte(exampleV39Scope), &v39Scope)

	gs := v039metadata.GenesisState{
		ScopeRecords:   []v039metadata.Scope{v39Scope},
		Specifications: []v039metadata.ContractSpec{v39ContractSpec},
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
  "o_s_locator_params": {
    "max_uri_length": 2048
  },
  "object_store_locators": [],
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
  "records": [
    {
      "inputs": [],
      "name": "loan",
      "outputs": [
        {
          "hash": "poYoiYr8gi22vyBlo09YkSSnGHRY0jQW9DZAvaTPT5slTbt2SV8KgGSKoK72PYVL/yLrCgnrEDaRn08byB/JHQ==",
          "status": "RESULT_STATUS_PASS"
        }
      ],
      "process": {
        "hash": "VQUlaiUsAJB6XVegaZ3vD+vMdqfR+/jHp6UjNSW/RUdbJOTUrKn50/HkDGzQyUmniNYYZ/YxjwrdCkTB9mm6qw==",
        "method": "noop",
        "name": "io.provenance.proto.loan.LoanProtos$Loan"
      },
      "session_id": "session1qyqqq2wf3c4yt4u447m8pw65qcdq7trdkw95wnvc4e0kg7n72y252q9j0yu"
    }
  ],
  "scope_specifications": [],
  "scopes": [
    {
      "data_access": [
        "cosmos1rr4d0eu62pgt4edw38d2ev27798pfhdhm39zct"
      ],
      "owners": [
        {
          "address": "cosmos1rr4d0eu62pgt4edw38d2ev27798pfhdhm39zct",
          "role": "PARTY_TYPE_ORIGINATOR"
        }
      ],
      "scope_id": "scope1qqqqq2wf3c4yt4u447m8pw65qcdqrre82d",
      "specification_id": "",
      "value_owner_address": ""
    }
  ],
  "sessions": [
    {
      "audit": {
        "created_by": "tp1rr4d0eu62pgt4edw38d2ev27798pfhdhp5ttha",
        "created_date": "2020-12-17T22:01:23Z",
        "message": "",
        "updated_by": "",
        "updated_date": "0001-01-01T00:00:00Z",
        "version": 1
      },
      "name": "io.p8e.contracts.origination.ETLTouch",
      "parties": [
        {
          "address": "cosmos1rr4d0eu62pgt4edw38d2ev27798pfhdhm39zct",
          "role": "PARTY_TYPE_ORIGINATOR"
        }
      ],
      "session_id": "session1qyqqq2wf3c4yt4u447m8pw65qcdq7trdkw95wnvc4e0kg7n72y252q9j0yu",
      "specification_id": "contractspec1q0vjj76putlkvl0mx80hp5gemjts9mzggk"
    }
  ]
}`
	bz, err := clientCtx.JSONMarshaler.MarshalJSON(migrated)
	require.NoError(t, err)

	// Indent the JSON bz correctly.
	var jsonObj map[string]interface{}
	err = json.Unmarshal(bz, &jsonObj)
	require.NoError(t, err)
	indentedBz, err := json.MarshalIndent(jsonObj, "", "  ")
	require.NoError(t, err)

	println(string(indentedBz))
	require.Equal(t, expected, string(indentedBz))
}
