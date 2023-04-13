// Package wasm supports smart contract integration with the provenance metadata module.
package wasm

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/metadata/types"
)

// MetadataMsgParams are params for encoding []sdk.Msg types from the scope module.
// Only one field should be set per request.
type MetadataMsgParams struct {
	// Params for encoding a MsgWriteScopeRequest
	WriteScope *WriteScope `json:"write_scope,omitempty"`
}

// WriteScope are params for encoding a MsgWriteScopeRequest.
type WriteScope struct {
	// The scope we want to create/update.
	Scope Scope `json:"scope"`
	// The signers' addresses.
	Signers []string `json:"signers"`
}

// Encoder returns a smart contract message encoder for the metadata module.
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
	case params.WriteScope != nil:
		return params.WriteScope.Encode(contract)
	default:
		return nil, fmt.Errorf("wasm: invalid metadata encode request: %s", string(msg))
	}
}

// Encode creates a MsgAddScopeDataAccessRequest.
func (params *WriteScope) Encode(contract sdk.AccAddress) ([]sdk.Msg, error) {
	// Promote the contract address to first signer
	signers := promoteSigner(*params, contract)

	// verify the signer addresses are valid
	for _, addr := range signers {
		_, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			return nil, fmt.Errorf("wasm: signer address must be a Bech32 string: %w", err)
		}
	}
	scope, err := params.Scope.convertToBaseType()
	if err != nil {
		return nil, err
	}

	msg := types.NewMsgWriteScopeRequest(*scope, signers)

	return []sdk.Msg{msg}, nil
}

func promoteSigner(params WriteScope, address sdk.AccAddress) []string {
	signers := []string{}
	signers = append(signers, address.String())
	for _, addr := range params.Signers {
		if addr != address.String() {
			signers = append(signers, addr)
		}
	}
	return signers
}
