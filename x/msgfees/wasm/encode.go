// Package wasm supports smart contract integration with the provenance name module.
package wasm

import (
	"encoding/json"
	"fmt"

	"github.com/provenance-io/provenance/internal/provwasm"
	"github.com/provenance-io/provenance/x/msgfees/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Compile time interface check
var _ provwasm.Encoder = Encoder

// MsgFeesMsgParams are params for encoding []sdk.Msg types from the msgfees module.
type MsgFeesMsgParams struct {
	// Params for encoding a MsgAddMarkerRequest
	AssessCustomFee *AssessCustomFeeParams `json:"assess_custom_fee,omitempty"`
}

// AssessCustomFeeParams are params for encoding a MsgAssessCustomMsgFeeRequest.
type AssessCustomFeeParams struct {
	// The fee amount to assess
	Amount sdk.Coin `json:"amount"`
	// The signer of the message
	From string `json:"from"`
	// An optional short name
	Name string `json:"name,omitempty"`
	// An optional address to receive the fees. if present, fees are split 50/50 between the account and fee module,
	// otherwise the fee module receives the full amount
	Recipient string `json:"recipient,omitempty"`
}

// Encoder returns a smart contract message encoder for the name module.
func Encoder(contract sdk.AccAddress, msg json.RawMessage, version string) ([]sdk.Msg, error) {
	wrapper := struct {
		Params *MsgFeesMsgParams `json:"msgfees"`
	}{}
	if err := json.Unmarshal(msg, &wrapper); err != nil {
		return nil, fmt.Errorf("wasm: failed to unmarshal name encode params: %w", err)
	}
	params := wrapper.Params
	if params == nil {
		return nil, fmt.Errorf("wasm: nil msgfees encode params")
	}
	switch {
	case params != nil:
		return params.AssessCustomFee.Encode(contract)
	default:
		return nil, fmt.Errorf("wasm: invalid msgfees encode request: %s", string(msg))
	}
}

// Encode creates a MsgAssessCustomMsgFeeRequest.
func (params *AssessCustomFeeParams) Encode(contract sdk.AccAddress) ([]sdk.Msg, error) {
	// Create message request
	msg := types.NewMsgAssessCustomMsgFeeRequest(params.Name, params.Amount, params.Recipient, params.From)
	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}
	return []sdk.Msg{&msg}, nil
}
