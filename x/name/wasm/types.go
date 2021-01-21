// Package wasm supports smart contract integration with the provenance name module.
package wasm

// QueryResName contains the address from a name query.
type QueryResName struct {
	Name       string `json:"name"`
	Address    string `json:"address"`
	Restricted bool   `json:"restricted"`
}

// QueryResNames contains a sequence of name records.
type QueryResNames struct {
	Records []QueryResName `json:"records,omitempty"`
}
