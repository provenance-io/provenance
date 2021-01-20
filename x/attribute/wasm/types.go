// Package wasm supports smart contract integration with the provenance attribute module.
package wasm

// Attribute is a typed key-value pair attached to a cosmos account.
type Attribute struct {
	// The attribute name.
	Name string `json:"name"`
	// The attribute value.
	Value []byte `json:"value"`
	// The attribute value type.
	Type string `json:"type"`
}

// AttributeResponse returns attributes attached to a cosmos account.
type AttributeResponse struct {
	// The account account address in Bech32 formt
	Address string `json:"address"`
	// The attributes queried for the account.
	Attributes []Attribute `json:"attributes,omitempty"`
}
