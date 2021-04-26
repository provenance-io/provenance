// Package wasm supports smart contract integration with the provenance marker module.
package wasm

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/provenance-io/provenance/x/marker/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MarkerMsgParams are params for encoding []sdk.Msg types from the marker module.
// Only one field should be set per request.
type MarkerMsgParams struct {
	// Params for encoding a MsgAddMarkerRequest
	Create *CreateMarkerParams `json:"create_marker,omitempty"`
	// Params for encoding a MsgAddAccessRequest
	Grant *GrantAccessParams `json:"grant_marker_access,omitempty"`
	// Params for encoding a MsgRevokeAccess
	Revoke *RevokeAccessParams `json:"revoke_marker_access,omitempty"`
	// Params for encoding a MsgFinalizeRequest
	Finalize *FinalizeMarkerParams `json:"finalize_marker,omitempty"`
	// Params for encoding a MsgActivateMarker
	Activate *ActivateMarkerParams `json:"activate_marker,omitempty"`
	// Params for encoding a MsgCancelRequest
	Cancel *CancelMarkerParams `json:"cancel_marker,omitempty"`
	// Params for encoding a MsgDeleteRequest
	Destroy *DestroyMarkerParams `json:"destroy_marker,omitempty"`
	// Params for encoding a MsgMintRequest
	Mint *MintSupplyParams `json:"mint_marker_supply,omitempty"`
	// Params for encoding a MsgBurnRequest
	Burn *BurnSupplyParams `json:"burn_marker_supply,omitempty"`
	// Params for encoding a MsgWithdrawRequest
	Withdraw *WithdrawParams `json:"withdraw_coins,omitempty"`
	// Params for encoding a MsgTransferRequest
	Transfer *TransferParams `json:"transfer_marker_coins,omitempty"`
}

// CreateMarkerParams are params for encoding a MsgAddMarkerRequest.
type CreateMarkerParams struct {
	// The marker denomination and amount
	Coin sdk.Coin `json:"coin"`
	// The marker type
	Type string `json:"marker_type,omitempty"`
}

// GrantAccessParams are params for encoding a MsgAddAccessRequest.
type GrantAccessParams struct {
	// The marker denomination
	Denom string `json:"denom"`
	// The grant permissions
	Permissions []string `json:"permissions"`
	// The grant address
	Address string `json:"address"`
}

// RevokeAccessParams are params for encoding a MsgDeleteAccessRequest.
type RevokeAccessParams struct {
	// The marker denom
	Denom string `json:"denom"`
	// The address to revoke
	Address string `json:"address"`
}

// FinalizeMarkerParams are params for encoding a MsgFinalizeRequest.
type FinalizeMarkerParams struct {
	// The marker denomination
	Denom string `json:"denom"`
}

// ActivateMarkerParams are params for encoding a MsgActivateRequest.
type ActivateMarkerParams struct {
	// The marker denomination
	Denom string `json:"denom"`
}

// CancelMarkerParams are params for encoding a MsgCancelRequest.
type CancelMarkerParams struct {
	// The marker denomination
	Denom string `json:"denom"`
}

// DestroyMarkerParams are params for encoding a MsgDeleteRequest.
type DestroyMarkerParams struct {
	// The marker denomination
	Denom string `json:"denom"`
}

// MintSupplyParams are params for encoding a MsgMintRequest.
type MintSupplyParams struct {
	// The marker denomination and amount to mint
	Coin sdk.Coin `json:"coin"`
}

// BurnSupplyParams are params for encoding a MsgBurnRequest.
type BurnSupplyParams struct {
	// The marker denomination and amount to burn
	Coin sdk.Coin `json:"coin"`
}

// WithdrawParams are params for encoding a MsgWithdrawRequest.
type WithdrawParams struct {
	// The marker denomination
	Denom string `json:"marker_denom"`
	// The withdrawal denomination and amount
	Coin sdk.Coin `json:"coin"`
	// The recipient of the withdrawal
	Recipient string `json:"recipient"`
}

// TransferParams are params for encoding a MsgTransferRequest.
type TransferParams struct {
	// The denomination and amount to transfer
	Coin sdk.Coin `json:"coin"`
	// The recipient of the transfer
	To string `json:"to"`
	// The sender of the transfer
	From string `json:"from"`
}

// Encoder returns a smart contract message encoder for the name module.
func Encoder(contract sdk.AccAddress, msg json.RawMessage, version string) ([]sdk.Msg, error) {
	wrapper := struct {
		Params *MarkerMsgParams `json:"marker"`
	}{}
	if err := json.Unmarshal(msg, &wrapper); err != nil {
		return nil, fmt.Errorf("wasm: failed to unmarshal marker encode params: %w", err)
	}
	params := wrapper.Params
	if params == nil {
		return nil, fmt.Errorf("wasm: nil marker encode params")
	}
	switch {
	case params.Create != nil:
		return params.Create.Encode(contract)
	case params.Grant != nil:
		return params.Grant.Encode(contract)
	case params.Revoke != nil:
		return params.Revoke.Encode(contract)
	case params.Finalize != nil:
		return params.Finalize.Encode(contract)
	case params.Activate != nil:
		return params.Activate.Encode(contract)
	case params.Cancel != nil:
		return params.Cancel.Encode(contract)
	case params.Destroy != nil:
		return params.Destroy.Encode(contract)
	case params.Mint != nil:
		return params.Mint.Encode(contract)
	case params.Burn != nil:
		return params.Burn.Encode(contract)
	case params.Withdraw != nil:
		return params.Withdraw.Encode(contract)
	case params.Transfer != nil:
		return params.Transfer.Encode(contract)
	default:
		return nil, fmt.Errorf("wasm: invalid marker encode request: %s", string(msg))
	}
}

// Encode creates a MsgAddMarkerRequest.
// The contract must be the signer (from address) and manager of the marker.
func (params *CreateMarkerParams) Encode(contract sdk.AccAddress) ([]sdk.Msg, error) {
	if !params.Coin.IsValid() {
		return nil, fmt.Errorf("wasm: invalid marker supply in CreateMarkerParams: coin is invalid")
	}
	if strings.TrimSpace(params.Type) == "" {
		return nil, fmt.Errorf("wasm: missing marker type in CreateMarkerParams")
	}
	markerType, err := types.MarkerTypeFromString(params.Type)
	if err != nil {
		return nil, fmt.Errorf("wasm: invalid marker type in CreateMarkerParams: %w", err)
	}
	msg := types.NewMsgAddMarkerRequest(
		params.Coin.Denom, params.Coin.Amount, contract, contract, markerType, false, false,
	)

	return []sdk.Msg{msg}, nil
}

// Encode creates a MsgAddAccessRequest.
// The contract must be the administrator of the marker.
func (params *GrantAccessParams) Encode(contract sdk.AccAddress) ([]sdk.Msg, error) {
	address, err := sdk.AccAddressFromBech32(params.Address)
	if err != nil {
		return nil, fmt.Errorf("wasm: invalid address in GrantAccessParams: %w", err)
	}
	if strings.TrimSpace(params.Denom) == "" {
		return nil, fmt.Errorf("wasm: empty denomination in GrantAccessParams")
	}
	access := make([]types.Access, len(params.Permissions))
	for i, perm := range params.Permissions {
		access[i] = types.AccessByName(perm)
	}
	msg := types.NewMsgAddAccessRequest(
		params.Denom,
		contract,
		*types.NewAccessGrant(address, access),
	)
	return []sdk.Msg{msg}, nil
}

// Encode creates a MsgDeleteAccessRequest.
// The contract must be the administrator of the marker.
func (params *RevokeAccessParams) Encode(contract sdk.AccAddress) ([]sdk.Msg, error) {
	if err := sdk.ValidateDenom(params.Denom); err != nil {
		return nil, fmt.Errorf("wasm: invalid denomination in RevokeAccessParams: %w", err)
	}
	address, err := sdk.AccAddressFromBech32(params.Address)
	if err != nil {
		return nil, fmt.Errorf("wasm: invalid address in RevokeAccessParams: %w", err)
	}
	msg := types.NewDeleteAccessRequest(params.Denom, contract, address)
	return []sdk.Msg{msg}, nil
}

// Encode creates a MsgFinalizeRequest.
// The contract must be the administrator of the marker.
func (params *FinalizeMarkerParams) Encode(contract sdk.AccAddress) ([]sdk.Msg, error) {
	if err := sdk.ValidateDenom(params.Denom); err != nil {
		return nil, fmt.Errorf("wasm: invalid denomination in FinalizeMarkerParams: %w", err)
	}
	msg := types.NewMsgFinalizeRequest(params.Denom, contract)
	return []sdk.Msg{msg}, nil
}

// Encode creates a MsgActivateRequest.
// The contract must be the administrator of the marker.
func (params *ActivateMarkerParams) Encode(contract sdk.AccAddress) ([]sdk.Msg, error) {
	if err := sdk.ValidateDenom(params.Denom); err != nil {
		return nil, fmt.Errorf("wasm: invalid denomination in ActivateMarkerParams: %w", err)
	}
	msg := types.NewMsgActivateRequest(params.Denom, contract)
	return []sdk.Msg{msg}, nil
}

// Encode creates a MsgCancelRequest.
// The contract must be the administrator of the marker.
func (params *CancelMarkerParams) Encode(contract sdk.AccAddress) ([]sdk.Msg, error) {
	if err := sdk.ValidateDenom(params.Denom); err != nil {
		return nil, fmt.Errorf("wasm: invalid denomination in CancelMarkerParams: %w", err)
	}
	msg := types.NewMsgCancelRequest(params.Denom, contract)
	return []sdk.Msg{msg}, nil
}

// Encode creates a MsgDeleteRequest.
// The contract must be the administrator of the marker.
func (params *DestroyMarkerParams) Encode(contract sdk.AccAddress) ([]sdk.Msg, error) {
	if err := sdk.ValidateDenom(params.Denom); err != nil {
		return nil, fmt.Errorf("wasm: invalid denomination in DestroyMarkerParams: %w", err)
	}
	msg := types.NewMsgDeleteRequest(params.Denom, contract)
	return []sdk.Msg{msg}, nil
}

// Encode creates a MsgMintRequest.
// The contract must be the administrator of the marker.
func (params *MintSupplyParams) Encode(contract sdk.AccAddress) ([]sdk.Msg, error) {
	if !params.Coin.IsValid() {
		return nil, fmt.Errorf("wasm: invalid MintSupplyParams: coin is invalid")
	}
	msg := types.NewMsgMintRequest(contract, params.Coin)
	return []sdk.Msg{msg}, nil
}

// Encode creates a MsgBurnRequest.
// The contract must be the administrator of the marker.
func (params *BurnSupplyParams) Encode(contract sdk.AccAddress) ([]sdk.Msg, error) {
	if !params.Coin.IsValid() {
		return nil, fmt.Errorf("wasm: invalid BurnSupplyParams: coin is invalid")
	}
	msg := types.NewMsgBurnRequest(contract, params.Coin)
	return []sdk.Msg{msg}, nil
}

// Encode creates a MsgWithdrawRequest.
// The contract must be the administrator of the marker.
func (params *WithdrawParams) Encode(contract sdk.AccAddress) ([]sdk.Msg, error) {
	if err := sdk.ValidateDenom(params.Denom); err != nil {
		return nil, fmt.Errorf("wasm: invalid marker denom in WithdrawParams: %w", err)
	}
	if !params.Coin.IsValid() {
		return nil, fmt.Errorf("wasm: invalid WithdrawParams: coin is invalid")
	}
	recipient, err := sdk.AccAddressFromBech32(params.Recipient)
	if err != nil {
		return nil, fmt.Errorf("wasm: invalid recipient address: %w", err)
	}
	msg := types.NewMsgWithdrawRequest(
		contract, recipient, params.Denom, sdk.NewCoins(params.Coin))
	return []sdk.Msg{msg}, nil
}

// Encode creates a MsgTransferRequest.
// The contract must be the administrator of the marker.
func (params *TransferParams) Encode(contract sdk.AccAddress) ([]sdk.Msg, error) {
	if !params.Coin.IsValid() {
		return nil, fmt.Errorf("wasm: invalid TransferParams: coin is invalid")
	}
	to, err := sdk.AccAddressFromBech32(params.To)
	if err != nil {
		return nil, fmt.Errorf("wasm: invalid 'to' address in TransferParams: %w", err)
	}
	from, err := sdk.AccAddressFromBech32(params.From)
	if err != nil {
		return nil, fmt.Errorf("wasm: invalid 'from' address in TransferParams: %w", err)
	}
	msg := types.NewMsgTransferRequest(contract, from, to, params.Coin)
	return []sdk.Msg{msg}, nil
}
