package v039

// ModuleName is the name of this module.
const ModuleName string = "account"

// Attribute is a typed key-value pair we can store under an account.
type Attribute struct {
	// The attribute name.
	Name string `json:"name"`
	// The attribute value.
	Value []byte `json:"value"`
	// The attribute value type.
	Type string `json:"type"`
	// The attribute address (used only for genesis import/export)
	Address string `json:"address,omitempty"`
	// The block height this attr was added in.
	Height int64 `json:"height,omitempty"`
}

// GenesisState is the module state at genesis.
type GenesisState struct {
	Attributes []Attribute `json:"attributes"`
}
