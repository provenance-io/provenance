package wasm

import (
	"encoding/json"
	"fmt"

	types2 "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/expiration/types"
)

// These are the Go struct that map to the Rust smart contract types defined by the provwasm JSON schema:
// https://github.com/provenance-io/provwasm/tree/main/packages/bindings/schema/expiration.json

// Expiration holds a typed key/value structure for data associated with an expiring module asset
type Expiration struct {
	ModuleAssetID string          `json:"module_asset_id"`
	Owner         string          `json:"owner"`
	BlockHeight   int64           `json:"block_height"`
	Deposit       sdk.Coin        `json:"deposit"`
	Message       json.RawMessage `json:"message"`
}

type Expirations struct {
	Expirations []*Expiration `json:"expirations"`
}

func (expiration *Expiration) convertToBaseType() (*types.Expiration, error) {
	anyMsg, err := rawMessageToAny(expiration.Message)
	if err != nil {
		return nil, err
	}

	baseType := &types.Expiration{
		ModuleAssetId: expiration.ModuleAssetID,
		Owner:         expiration.Owner,
		BlockHeight:   expiration.BlockHeight,
		Deposit:       expiration.Deposit,
		Message:       anyMsg,
	}
	if err := baseType.ValidateBasic(); err != nil {
		return nil, err
	}

	return baseType, nil
}

// Covert a json.RawMessage to `Any` making sure that it's making sure that the
// message type conforms to the `sdk.Msg` interface and that it is also registered
// with the InterfaceRegistry.
func rawMessageToAny(message json.RawMessage) (*types2.Any, error) {
	var msg sdk.Msg
	if err := types.ModuleCdc.UnmarshalInterfaceJSON(message, &msg); err != nil {
		return nil, err
	}
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	anyMsg, err := types2.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}
	return anyMsg, nil
}

// Convert an expiration into its provwasm type
func createExpiration(baseType *types.Expiration) (*Expiration, error) {
	var msg sdk.Msg
	if err := types.ModuleCdc.UnmarshalInterface(baseType.Message.Value, &msg); err != nil {
		return nil, fmt.Errorf("wasm: unmarshal %s message failed: %w", types.ModuleName, err)
	}

	rawMsg, err := types.ModuleCdc.MarshalInterfaceJSON(msg)
	if err != nil {
		return nil, fmt.Errorf("wasm: marshal %s message to JSON failed: %w", types.ModuleName, err)
	}

	expiration := &Expiration{
		ModuleAssetID: baseType.ModuleAssetId,
		Owner:         baseType.Owner,
		BlockHeight:   baseType.BlockHeight,
		Deposit:       baseType.Deposit,
		Message:       rawMsg,
	}

	return expiration, nil
}

// Convert an expiration into provwasm JSON format
func createExpirationResponse(baseType *types.Expiration) ([]byte, error) {
	expiration, err := createExpiration(baseType)
	if err != nil {
		return nil, err
	}

	bz, err := json.Marshal(expiration)
	if err != nil {
		return nil, fmt.Errorf("wasm: marshal %s failed: %w", types.ModuleName, err)
	}

	return bz, nil
}

// Convert a slice of expirations into provwasm JSON format
func createExpirationsResponse(baseTypes []*types.Expiration) ([]byte, error) {
	expirations := &Expirations{
		Expirations: make([]*Expiration, len(baseTypes)),
	}

	for i, baseType := range baseTypes {
		expiration, err := createExpiration(baseType)
		if err != nil {
			return nil, err
		}
		expirations.Expirations[i] = expiration
	}

	bz, err := json.Marshal(expirations)
	if err != nil {
		return nil, fmt.Errorf("wasm: marshal %ss failed: %w", types.ModuleName, err)
	}

	return bz, nil
}
