package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gopkg.in/yaml.v2"
)

// query endpoints supported by the auth Querier
const (
	QueryScope     = "account"
	QueryOwnership = "ownership"
	QueryParams    = "params"
)

// QueryMetadataParams defines the params for queries that support paging (get by scope)
type QueryMetadataParams struct {
	Page, Limit int
}

// NewQueryMetadataParams object
func NewQueryMetadataParams(page, limit int) QueryMetadataParams {
	return QueryMetadataParams{page, limit}
}

// QueryResScope is the result for legacy scope queries.
type QueryResScope struct {
	Scope []byte `json:"scope"`
}

// String implements fmt.Stringer
func (qr QueryResScope) String() string {
	return fmt.Sprintf("%X", qr.Scope)
}

// MarshalYAML returns the YAML representation of a QueryResScope.
func (qr QueryResScope) MarshalYAML() (interface{}, error) {
	// Unmarshal protobuf scope
	scope := &Scope{}
	if err := scope.Unmarshal(qr.Scope); err != nil {
		return nil, err
	}
	bs, err := yaml.Marshal(qr.Scope)
	if err != nil {
		return nil, err
	}

	return string(bs), nil
}

// QueryResOwnership is the result of a query for scopes associated with a given address
// query: 'custom/metadata/ownership/{address}
type QueryResOwnership struct {
	Address sdk.AccAddress `json:"address" yaml:"address"`
	ScopeID []string       `json:"scope_id" yaml:"scope_id"`
}

// String representation of the query for address ownership scope IDs.
func (qor QueryResOwnership) String() string {
	return fmt.Sprintf("%s - %s", qor.Address.String(), strings.Join(qor.ScopeID, ", "))
}
