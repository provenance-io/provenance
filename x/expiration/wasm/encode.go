// Package wasm supports smart contract integration with the provenance expiration module.
package wasm

//import (
//	"encoding/json"
//	"fmt"
//
//	sdk "github.com/cosmos/cosmos-sdk/types"
//
//	"github.com/provenance-io/provenance/internal/provwasm"
//	"github.com/provenance-io/provenance/x/expiration/types"
//)
//
//// Compile time interface check
//var _ provwasm.Encoder = Encoder
//
//// ExpirationMsgParams are params for encoding []sdk.Msg types from the expiration module.
//// Only one field should be set per request.
//type ExpirationMsgParams struct {
//	AddExpiration *AddExpiration `json:"add_expiration,omitempty"`
//}
//
//// AddExpiration are params for encoding a MsgAddExpirationRequest
//type AddExpiration struct {
//	Expiration types.Expiration `json:"expiration"`
//	Signers    []string         `json:"signers"`
//}
//
//// Encoder returns a smart contract message encoder for the expiration module.
//func Encoder(_ sdk.AccAddress, msg json.RawMessage, _ string) ([]sdk.Msg, error) {
//	wrapper := struct {
//		Params *ExpirationMsgParams `json:"expiration"`
//	}{}
//	if err := json.Unmarshal(msg, &wrapper); err != nil {
//		return nil, fmt.Errorf("wasm: failed to unmarshal %s encode params: %w", types.ModuleName, err)
//	}
//	params := wrapper.Params
//	if params == nil {
//		return nil, fmt.Errorf("wasm: nil %s encode params", types.ModuleName)
//	}
//	switch {
//	case params != nil:
//		return params.AddExpiration.Encode()
//	default:
//		return nil, fmt.Errorf("wasm: invalid %s encode request: %s", types.ModuleName, string(msg))
//	}
//}
//
//// Encode creates a MsgAddExpirationRequest
//func (params *AddExpiration) Encode() ([]sdk.Msg, error) {
//	// verify the signer addresses are valid
//	for _, addr := range params.Signers {
//		if _, err := sdk.AccAddressFromBech32(addr); err != nil {
//			return nil, fmt.Errorf("wasm: signer address must be a Bech32 string: %v", err)
//		}
//	}
//	// Create message request
//	msg := types.NewMsgAddExpirationRequest(params.Expiration, params.Signers)
//	if err := msg.ValidateBasic(); err != nil {
//		return nil, err
//	}
//	return []sdk.Msg{msg}, nil
//}
