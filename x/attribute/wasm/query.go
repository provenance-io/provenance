// Package wasm supports smart contract integration with the provenance attribute module.
package wasm

import (
	"encoding/json"
	"fmt"

	"github.com/provenance-io/provenance/internal/provwasm"
	"github.com/provenance-io/provenance/x/attribute/keeper"
	"github.com/provenance-io/provenance/x/attribute/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AttributeQueryParams represents the request type for the attribute module sent by a smart contracts.
// Only one field should be set.
type AttributeQueryParams struct {
	// Get account attributes by name.
	Get *GetAttributesParams `json:"get_attributes,omitempty"`
	// Get all account attributes.
	GetAll *GetAllAttributesParams `json:"get_all_attributes,omitempty"`
}

// GetAttributesParams are params for querying an account attributes by address and name.
type GetAttributesParams struct {
	// The account address
	Address string `json:"address"`
	// The name of the attributes to query
	Name string `json:"name"`
}

// GetAllAttributesParams are params for querying account attributes by address.
type GetAllAttributesParams struct {
	// The account to query
	Address string `json:"address"`
}

// Querier returns a smart contract querier for the attribute module.
func Querier(keeper keeper.Keeper) provwasm.Querier {
	return func(ctx sdk.Context, query json.RawMessage, version string) ([]byte, error) {
		wrapper := struct {
			Params *AttributeQueryParams `json:"attribute"`
		}{}
		if err := json.Unmarshal(query, &wrapper); err != nil {
			return nil, fmt.Errorf("wasm: invalid query: %w", err)
		}
		params := wrapper.Params
		if params == nil {
			return nil, fmt.Errorf("wasm: nil account query params")
		}
		switch {
		case params.Get != nil:
			return params.Get.Run(ctx, keeper)
		case params.GetAll != nil:
			return params.GetAll.Run(ctx, keeper)
		default:
			return nil, fmt.Errorf("wasm: invalid account attribute query: %s", string(query))
		}
	}
}

// Run queries for account attributes by address and name.
func (params *GetAttributesParams) Run(ctx sdk.Context, keeper keeper.Keeper) ([]byte, error) {
	err := types.ValidateAttributeAddress(params.Address)
	if err != nil {
		return nil, fmt.Errorf("wasm: invalid address: %w", err)
	}
	attrs, err := keeper.GetAttributes(ctx, params.Address, params.Name)
	if err != nil {
		return nil, fmt.Errorf("wasm: attribute query failed: %w", err)
	}
	return createResponse(params.Address, attrs)
}

// Run queries for account attributes by address.
func (params *GetAllAttributesParams) Run(ctx sdk.Context, keeper keeper.Keeper) ([]byte, error) {
	err := types.ValidateAttributeAddress(params.Address)
	if err != nil {
		return nil, fmt.Errorf("wasm: invalid address: %w", err)
	}
	attrs, err := keeper.GetAllAttributes(ctx, params.Address)
	if err != nil {
		return nil, fmt.Errorf("wasm: attribute query failed: %w", err)
	}
	return createResponse(params.Address, attrs)
}

// Create a JSON response from the results of a account attribute query.
func createResponse(address string, attrs []types.Attribute) ([]byte, error) {
	res := AttributeResponse{Address: address}
	for _, a := range attrs {
		attr := Attribute{
			Name:  a.Name,
			Value: append([]byte{}, a.Value...),
			Type:  decodeType(a.AttributeType),
		}
		res.Attributes = append(res.Attributes, attr)
	}
	bz, err := json.Marshal(res)
	if err != nil {
		return nil, fmt.Errorf("wasm: marshal response failed: %w", err)
	}
	return bz, nil
}

// Adapt the attribute type to a string that will deserialize to the correct rust enum type on
// the smart contract side of the query.
func decodeType(attributeType types.AttributeType) string {
	switch attributeType {
	case types.AttributeType_Bytes:
		return "bytes"
	case types.AttributeType_Float:
		return "float"
	case types.AttributeType_Int:
		return "int"
	case types.AttributeType_JSON:
		return "json"
	case types.AttributeType_String:
		return "string"
	case types.AttributeType_UUID:
		return "uuid"
	case types.AttributeType_Uri:
		return "uri"
	case types.AttributeType_Proto:
		return "proto"
	default:
		return "unspecified"
	}
}
