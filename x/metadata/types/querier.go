package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gopkg.in/yaml.v2"

	legacy "github.com/provenance-io/provenance/x/metadata/legacy/v039"
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
	scope := &legacy.Scope{}
	if err := scope.Unmarshal(qr.Scope); err != nil {
		return nil, err
	}

	// Collect parties and their roles
	formattedParties := make([]formattedParty, len(scope.Parties))
	for i, party := range scope.Parties {
		formattedParties[i] = formattedParty{sdk.AccAddress(party.Address).String(), party.SignerRole.String()}
	}

	// Build up a low fidelity record collection
	var formattedRecords []formattedRecord
	for _, recordGroup := range scope.RecordGroup {
		groupOwners := make([]sdk.AccAddress, len(recordGroup.Parties))
		for j, party := range recordGroup.Parties {
			groupOwners[j] = party.Address
		}
		for _, record := range recordGroup.Records {
			inputArgs := make([]formattedInput, len(record.Inputs))
			for i, input := range record.Inputs {
				inputArgs[i] = formattedInput{
					ClassName: input.Classname,
					Hash:      input.Hash,
					Name:      input.Name,
				}
			}
			formattedRecords = append(formattedRecords, formattedRecord{
				Name:       record.Name,
				Inputs:     inputArgs,
				ClassName:  record.Classname,
				ResultHash: record.ResultHash,
				Owners:     groupOwners,
			})
		}
	}

	// Using abbreviated structures, marshall to yaml and output.
	bs, err := yaml.Marshal(formattedScope{
		ScopeID: scope.Uuid.Value,
		Parties: formattedParties,
		Records: formattedRecords,
	})

	if err != nil {
		return nil, err
	}

	return string(bs), nil
}

type formattedParty struct {
	Address string
	Party   string `yaml:"role"`
}
type formattedRecord struct {
	Name       string
	Owners     []sdk.AccAddress
	ClassName  string
	Inputs     []formattedInput
	ResultHash string
}

type formattedInput struct {
	Name      string
	Hash      string
	ClassName string
}
type formattedScope struct {
	ScopeID     string `yaml:"scope_id"`
	BlockHeight int64  `yaml:"block_height,omitempty"`
	Parties     []formattedParty
	Records     []formattedRecord
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
