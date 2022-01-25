// Package wasm supports smart contract integration with the provenance metadata module.
package wasm

import (
	"encoding/json"
	"fmt"
	"github.com/provenance-io/provenance/x/metadata/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MetadataMsgParams are params for encoding []sdk.Msg types from the scope module.
// Only one field should be set per request.
type MetadataMsgParams struct {
	// Params for encoding a MsgAddMetadataRequest
	AddDataAccess    *AddDataAccess `json:"add_data_access,omitempty"`
	DeleteDataAccess *AddDataAccess `json:"delete_data_access,omitempty"`
}

// AddDataAccess are params for encoding a MsgAddScopeDataAccessRequest.
type AddDataAccess struct {
	// The bech32 address of the scope we want to update.
	ScopeID string `json:"scope_id"`
	// The data access addresses we want to add to the scope.
	DataAccess []string `json:"data_access"`
	// The signers' addresses.
	Signers []string `json:"signers"`
}

// DeleteDataAccess are params for encoding a MsgDeleteScopeDataAccessRequest.
type DeleteDataAccess struct {
	// The bech32 address of the scope we want to update.
	ScopeID string `json:"scope_id"`
	// The data access addresses we want to delete from the scope.
	DataAccess []string `json:"data_access"`
	// The signers' addresses.
	Signers []string `json:"signers"`
}

// SetValueOwnerAddress are params for encoding a .
type SetValueOwnerAddress struct {
	// The bech32 address of the scope we want to update.
	ScopeID string `json:"scope_id"`
	// The bech32 address of the value owner to set on the scope.
	ValueOwnerAddress string `json:"value_owner_address"`
}

// Encoder returns a smart contract message encoder for the name module.
func Encoder(contract sdk.AccAddress, msg json.RawMessage, version string) ([]sdk.Msg, error) {
	wrapper := struct {
		Params *MetadataMsgParams `json:"metadata"`
	}{}
	if err := json.Unmarshal(msg, &wrapper); err != nil {
		return nil, fmt.Errorf("wasm: failed to unmarshal metadata encode params: %w", err)
	}
	params := wrapper.Params
	if params == nil {
		return nil, fmt.Errorf("wasm: nil metadata encode params")
	}
	switch {
	case params.AddDataAccess != nil:
		return params.AddDataAccess.Encode(contract)
	case params.DeleteDataAccess != nil:
		return params.DeleteDataAccess.Encode(contract)
	default:
		return nil, fmt.Errorf("wasm: invalid metadata encode request: %s", string(msg))
	}
}

// Encode creates a MsgAddScopeDataAccessRequest.
func (params *AddDataAccess) Encode(contract sdk.AccAddress) ([]sdk.Msg, error) {
	scopeID, err := types.MetadataAddressFromBech32(params.ScopeID)
	if err != nil {
		return nil, fmt.Errorf("wasm: invalid scope ID: %w", err)
	}
	// verify the data_access addresses are valid
	for _, addr := range params.DataAccess {
		_, err = sdk.AccAddressFromBech32(addr)
		if err != nil {
			return nil, fmt.Errorf("wasm: invalid 'data_access' address: %v", err)
		}
	}
	// verify the signer addresses are valid
	for _, addr := range params.Signers {
		_, err = sdk.AccAddressFromBech32(addr)
		if err != nil {
			return nil, fmt.Errorf("wasm: signer address must be a Bech32 string: %v", err)
		}
	}
	msg := types.NewMsgAddScopeDataAccessRequest(
		scopeID, params.DataAccess, params.Signers,
	)

	return []sdk.Msg{msg}, nil
}

// Encode creates a MsgDeleteScopeDataAccessRequest.
func (params *DeleteDataAccess) Encode(contract sdk.AccAddress) ([]sdk.Msg, error) {
	scopeID, err := types.MetadataAddressFromBech32(params.ScopeID)
	if err != nil {
		return nil, fmt.Errorf("wasm: invalid scope ID: %v", err)
	}
	// verify the data_access addresses are valid
	for _, addr := range params.DataAccess {
		_, err = sdk.AccAddressFromBech32(addr)
		if err != nil {
			return nil, fmt.Errorf("wasm: invalid 'data_access' address: %v", err)
		}
	}
	// verify the signer addresses are valid
	for _, addr := range params.Signers {
		_, err = sdk.AccAddressFromBech32(addr)
		if err != nil {
			return nil, fmt.Errorf("wasm: signer address must be a Bech32 string: %v", err)
		}
	}
	msg := types.NewMsgDeleteScopeDataAccessRequest(
		scopeID, params.DataAccess, params.Signers,
	)

	return []sdk.Msg{msg}, nil
}

//// Encode creates a MsgDeleteScopeDataAccessRequest.
//func (params *SetValueOwnerAddress) Encode(contract sdk.AccAddress) ([]sdk.Msg, error) {
//	scopeID, err := types.MetadataAddressFromBech32(params.ScopeID)
//	if err != nil {
//		return nil, fmt.Errorf("wasm: invalid scope ID: %v", err)
//	}
//	valueOwnerAddress, err := sdk.AccAddressFromBech32(params.ValueOwnerAddress)
//	if err != nil {
//		return nil, fmt.Errorf("wasm: value owner address must be a Bech32 string: %v", err)
//	}
//	msg := types.(
//		scopeID, params.DataAccess, params.Signers,
//	)
//
//	return []sdk.Msg{msg}, nil
//}
