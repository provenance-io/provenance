package wasm

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/provenance-io/provenance/x/metadata/types"
)

// Scope is a root reference for a collection of records owned by one or more parties. This is the Go struct that maps
// to the Rust smart contract type defined by the provwasm JSON schema. This type was generated.
type Scope struct {
	ScopeID           string   `json:"scope_id"`
	SpecificationID   string   `json:"specification_id"`
	DataAccess        []string `json:"data_access,omitempty"`
	Owners            []*Party `json:"owners,omitempty"`
	ValueOwnerAddress string   `json:"value_owner_address"`
}

// PartyType defines roles that can be associated to a party.
type PartyType string

const (
	// PartyTypeAffiliate is a concrete party type.
	PartyTypeAffiliate PartyType = "affiliate"
	// PartyTypeCustodian is a concrete party type.
	PartyTypeCustodian PartyType = "custodian"
	// PartyTypeInvestor is a concrete party type.
	PartyTypeInvestor PartyType = "investor"
	// PartyTypeOmnibus is a concrete party type.
	PartyTypeOmnibus PartyType = "omnibus"
	// PartyTypeOriginator is a concrete party type.
	PartyTypeOriginator PartyType = "originator"
	// PartyTypeOwner is a concrete party type.
	PartyTypeOwner PartyType = "owner"
	// PartyTypeProvenance is a concrete party type.
	PartyTypeProvenance PartyType = "provenance"
	// PartyTypeServicer is a concrete party type.
	PartyTypeServicer PartyType = "servicer"
	// PartyTypeUnspecified is a concrete party type.
	PartyTypeUnspecified PartyType = "unspecified"
)

// Party is an address with an associated role.
type Party struct {
	Address string    `json:"address"`
	Role    PartyType `json:"role"`
}

// A helper function for converting metadata module scope types into provwasm query response types.
func createScopeResponse(input types.Scope) ([]byte, error) {
	scopeID, err := stringifyAddress(input.ScopeId)
	if err != nil {
		return nil, err
	}
	specID, err := stringifyAddress(input.SpecificationId)
	if err != nil {
		return nil, err
	}
	scope := &Scope{
		ScopeID:           scopeID,
		SpecificationID:   specID,
		ValueOwnerAddress: input.ValueOwnerAddress,
	}
	for _, da := range input.DataAccess {
		scope.DataAccess = append(scope.DataAccess, da)
	}
	for _, o := range input.Owners {
		scope.Owners = append(scope.Owners, createOwner(o))
	}
	bz, err := json.Marshal(scope)
	if err != nil {
		return nil, fmt.Errorf("wasm: marshal response failed: %w", err)
	}
	return bz, nil
}

// A non-panicing version of ma.String(). We don't want query panics in smart contracts.
// Just return the error, providing more info than a panic.
func stringifyAddress(ma types.MetadataAddress) (string, error) {
	if ma.Empty() {
		return "", fmt.Errorf("empty addresses are not supported")
	}
	hrp, err := types.VerifyMetadataAddressFormat(ma)
	if err != nil {
		return "", err
	}
	bech32Addr, err := bech32.ConvertAndEncode(hrp, ma.Bytes())
	if err != nil {
		return "", err
	}
	return bech32Addr, nil
}

// Convert a party to its provwasm type.
func createOwner(input types.Party) *Party {
	return &Party{
		Address: input.Address,
		Role:    createRole(input.Role),
	}
}

// Convert a party type to its provwasm type.
func createRole(input types.PartyType) PartyType {
	switch input {
	case types.PartyType_PARTY_TYPE_ORIGINATOR:
		return PartyTypeOriginator
	case types.PartyType_PARTY_TYPE_SERVICER:
		return PartyTypeServicer
	case types.PartyType_PARTY_TYPE_INVESTOR:
		return PartyTypeInvestor
	case types.PartyType_PARTY_TYPE_CUSTODIAN:
		return PartyTypeCustodian
	case types.PartyType_PARTY_TYPE_OWNER:
		return PartyTypeOwner
	case types.PartyType_PARTY_TYPE_AFFILIATE:
		return PartyTypeAffiliate
	case types.PartyType_PARTY_TYPE_OMNIBUS:
		return PartyTypeOmnibus
	case types.PartyType_PARTY_TYPE_PROVENANCE:
		return PartyTypeProvenance
	default:
		return PartyTypeUnspecified
	}
}
