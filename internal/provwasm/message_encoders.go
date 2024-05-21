// Package provwasm allows CosmWasm smart contracts to communicate with custom provenance modules.
package provwasm

import (
	"encoding/json"
	"fmt"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	"cosmossdk.io/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/provenance-io/provenance/internal/helpers"
)

// Encoder describes behavior for provenance smart contract message encoding.
// The contract address must ALWAYS be set as the Msg signer.
type Encoder func(contract sdk.AccAddress, data json.RawMessage, version string) ([]sdk.Msg, error)

// EncoderRegistry maps routes to encoders.
type EncoderRegistry struct {
	encoders map[string]Encoder
}

// NewEncoderRegistry creates a new registry for message encoders.
func NewEncoderRegistry() *EncoderRegistry {
	return &EncoderRegistry{
		encoders: make(map[string]Encoder),
	}
}

// RegisterEncoder adds a message encoder for the given route.
func (qr *EncoderRegistry) RegisterEncoder(route string, encoder Encoder) {
	if _, exists := qr.encoders[route]; exists {
		panic(fmt.Sprintf("wasm: encoder already registered for route: %s", route))
	}
	qr.encoders[route] = encoder
}

// MessageEncoders provides provenance message encoding support for smart contracts.
func MessageEncoders(registry *EncoderRegistry, logger log.Logger) *wasmkeeper.MessageEncoders {
	return &wasmkeeper.MessageEncoders{
		Custom: customEncoders(registry, logger),
	}
}

// Custom provenance encoders for CosmWasm integration.
func customEncoders(registry *EncoderRegistry, logger log.Logger) wasmkeeper.CustomEncoder {
	return func(contract sdk.AccAddress, msg json.RawMessage) ([]sdk.Msg, error) {
		req := EncodeRequest{}
		if err := json.Unmarshal(msg, &req); err != nil {
			logger.Error("failed to unmarshal encode request", "err", err)
			return nil, sdkerrors.ErrJSONUnmarshal.Wrap(err.Error())
		}
		encode, exists := registry.encoders[req.Route]
		if !exists {
			logger.Error("encoder not found", "route", req.Route)
			return nil, sdkerrors.ErrInvalidRequest.Wrapf("encoder not found for route: %s", req.Route)
		}
		msgs, err := encode(contract, req.Params, req.Version)
		if err != nil {
			logger.Error("failed to encode message", "err", err)
			return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
		}
		for _, msg := range msgs {
			if err := helpers.ValidateBasic(msg); err != nil {
				logger.Error("message validation failed", "err", err)
				return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
			}
		}
		return msgs, nil
	}
}
