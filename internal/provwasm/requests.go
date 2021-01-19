// Package provwasm allows CosmWasm smart contracts to communicate with custom provenance modules.
package provwasm

import (
	"encoding/json"
)

// RequestFields contains fields shared between query requests and encode requests.
// Version should be the semantic data format version of the provenance rust bindings (eg 1.2.3).
type RequestFields struct {
	// The router key of the module
	Route string `json:"route"`
	// The module-specific inputs represented as JSON.
	Params json.RawMessage `json:"params"`
	// Enables smart contract backwards compatibility.
	Version string `json:"version,omitempty"`
}

// QueryRequest is the top-level type for provenance query support in CosmWasm smart contracts.
type QueryRequest struct {
	RequestFields
}

// EncodeRequest is the top-level type for provenance message encoding support in CosmWasm smart contracts.
type EncodeRequest struct {
	RequestFields
}
