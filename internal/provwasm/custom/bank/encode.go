package bank

import (
	"encoding/json"
	"fmt"

	"github.com/provenance-io/provenance/internal/provwasm"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RouterKey is the registry key for the provwasm custom bank encoder
const RouterKey = "bank"

// Compile time interface check
var _ provwasm.Encoder = Encoder

// CustomBankMsgParams are params for wrapping bank []sdk.Msg types.
type CustomBankMsgParams struct {
	// Encode a bank send message
	Send *SendParams `json:"send,omitempty"`
}

// SendParams are params for encoding a MsgSend.
type SendParams struct {
	// The sender
	To string `json:"to"`
	// The recipient
	From string `json:"from"`
	// The funds to send
	Coins sdk.Coins `json:"coins"`
}

// Encoder returns a smart contract message encoder for bank sends.
func Encoder(contract sdk.AccAddress, msg json.RawMessage, version string) ([]sdk.Msg, error) {
	wrapper := struct {
		Params *CustomBankMsgParams `json:"bank"`
	}{}
	if err := json.Unmarshal(msg, &wrapper); err != nil {
		return nil, fmt.Errorf("wasm: failed to unmarshal bank encode params: %w", err)
	}
	params := wrapper.Params
	if params == nil {
		return nil, fmt.Errorf("wasm: nil bank encode params")
	}
	switch {
	case params.Send != nil:
		return params.Send.Encode(contract)
	default:
		return nil, fmt.Errorf("wasm: invalid bank encode request: %s", string(msg))
	}
}

// Encode creates a MsgSend.
func (params *SendParams) Encode(contract sdk.AccAddress) ([]sdk.Msg, error) {
	from, err := sdk.AccAddressFromBech32(params.From)
	if err != nil {
		return nil, fmt.Errorf("wasm: invalid from address: %w", err)
	}
	to, err := sdk.AccAddressFromBech32(params.To)
	if err != nil {
		return nil, fmt.Errorf("wasm: invalid to address: %w", err)
	}
	if err := params.Coins.Validate(); err != nil {
		return nil, fmt.Errorf("wasm: invalid coins: %w", err)
	}
	msg := banktypes.NewMsgSend(from, to, params.Coins)
	return []sdk.Msg{msg}, nil
}
