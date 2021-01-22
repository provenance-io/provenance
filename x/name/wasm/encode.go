// Package wasm supports smart contract integration with the provenance name module.
package wasm

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/provenance-io/provenance/internal/provwasm"
	"github.com/provenance-io/provenance/x/name/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Compile time interface check
var _ provwasm.Encoder = Encoder

// NameMsgParams are params for encoding []sdk.Msg types from the name module.
// Only one field should be set.
type NameMsgParams struct {
	// Encode a MsgBindName
	Bind *BindNameParams `json:"bind_name,omitempty"`
	// Encode a MsgUnBindName
	Delete *DeleteNameParams `json:"delete_name,omitempty"`
}

// BindNameParams are params for encoding a MsgBindName.
type BindNameParams struct {
	// The combined name and root name (eg in x.y.z : name = x, root_name = y.z)
	Name string `json:"name"`
	// The address to bind
	Address string `json:"address"`
	// Whether to restrict binding child names to the owner
	Restrict bool `json:"restrict"`
}

// DeleteNameParams are params for encoding a MsgDeleteNameRequest.
type DeleteNameParams struct {
	// The name to unbind from the contract address.
	Name string `json:"name"`
}

// Encoder returns a smart contract message encoder for the name module.
func Encoder(contract sdk.AccAddress, msg json.RawMessage, version string) ([]sdk.Msg, error) {
	wrapper := struct {
		Params *NameMsgParams `json:"name"`
	}{}
	if err := json.Unmarshal(msg, &wrapper); err != nil {
		return nil, fmt.Errorf("wasm: failed to unmarshal name encode params: %w", err)
	}
	params := wrapper.Params
	if params == nil {
		return nil, fmt.Errorf("wasm: nil name encode params")
	}
	switch {
	case params.Bind != nil:
		return params.Bind.Encode(contract)
	case params.Delete != nil:
		return params.Delete.Encode(contract)
	default:
		return nil, fmt.Errorf("wasm: invalid name encode request: %s", string(msg))
	}
}

// Encode creates a MsgBindNameRequest.
// The parent address is required to be the signer. But, in x/wasm the contract address must be the signer.
// This means that contract instances should have a parent name they own, or the parent name must be unrestricted.
func (params *BindNameParams) Encode(contract sdk.AccAddress) ([]sdk.Msg, error) {
	address, err := sdk.AccAddressFromBech32(params.Address)
	if err != nil {
		return nil, fmt.Errorf("wasm: invalid bind address: %w", err)
	}
	// Extract the name and parent name (eg in x.y.z : name = x, parent_name = y.z)
	names := strings.SplitN(params.Name, ".", 2)
	if len(names) != 2 {
		return nil, fmt.Errorf("wasm: invalid name: %s", params.Name)
	}
	// Create message request
	record := types.NewNameRecord(names[0], address, params.Restrict)
	parent := types.NewNameRecord(names[1], contract, false)
	msg := types.NewMsgBindNameRequest(record, parent)
	return []sdk.Msg{msg}, nil
}

// Encode creates a MsgDeleteNameRequest.
func (params *DeleteNameParams) Encode(contract sdk.AccAddress) ([]sdk.Msg, error) {
	record := types.NewNameRecord(params.Name, contract, false)
	msg := types.NewMsgDeleteNameRequest(record)
	return []sdk.Msg{msg}, nil
}
