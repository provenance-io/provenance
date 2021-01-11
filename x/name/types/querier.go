package types

// querier keys
const (
	// The query base for getting the module params
	QueryParams = "params"
	// The query base for resolving names.
	QueryResolve = "resolve"
	// The query base for reverse lookup.
	QueryLookup = "lookup"
)

// QueryNameResult contains the address from a name query.
type QueryNameResult struct {
	Name       string `json:"name"`
	Address    string `json:"address"`
	Restricted bool   `json:"restricted"`
}

// String implements fmt.Stringer
func (r QueryNameResult) String() string {
	return r.Name + "->" + r.Address // bech32 string
}

// QueryNameResults contains a sequence of name records.
type QueryNameResults struct {
	Records []QueryNameResult `json:"records"`
}
