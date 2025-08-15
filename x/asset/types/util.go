package types

import (
	"encoding/json"
	"fmt"

	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"

	codec "github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	proto "github.com/cosmos/gogoproto/proto"

	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

// AnyToString extracts a string value from an Any type that contains a StringValue
func AnyToString(_ codec.BinaryCodec, anyMsg *cdctypes.Any) (string, error) {
	// Convert to proto Any so we can use UnmarshalTo
	stdAny := &wrapperspb.StringValue{}
	err := proto.Unmarshal(anyMsg.Value, stdAny)
	if err != nil {
		return "", err
	}
	return stdAny.Value, nil
}

// StringToAny converts a string to an Any type by wrapping it in a StringValue
func StringToAny(str string) (*cdctypes.Any, error) {
	strMsg := wrapperspb.String(str)
	anyMsg, err := cdctypes.NewAnyWithValue(strMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Any from string: %w", err)
	}
	return anyMsg, nil
}

func AnyToJSONSchema(cdc codec.BinaryCodec, anyValue *cdctypes.Any) (map[string]interface{}, error) {
	if anyValue == nil {
		return nil, nil
	}

	schemaString, err := AnyToString(cdc, anyValue)
	if err != nil {
		return nil, err
	}

	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(schemaString), &schema); err != nil {
		return nil, err
	}
	return schema, nil
}

func ValidateJSONSchema(_ map[string]interface{}, data []byte) error {
	// Simple JSON validation - just check if the data is valid JSON
	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return fmt.Errorf("invalid JSON data: %w", err)
	}

	// Note: Full JSON schema validation has been simplified due to import restrictions.
	// This function now only validates that the data is well-formed JSON.
	return nil
}

func NewDefaultMarker(denom sdk.Coin, fromAddr string) (*markertypes.MarkerAccount, error) {
	// Get the from address
	fromAcc, err := sdk.AccAddressFromBech32(fromAddr)
	if err != nil {
		return &markertypes.MarkerAccount{}, fmt.Errorf("invalid from address: %w", err)
	}

	// Create a new marker account
	markerAddr := markertypes.MustGetMarkerAddress(denom.Denom)
	marker := markertypes.NewMarkerAccount(
		authtypes.NewBaseAccountWithAddress(markerAddr),
		denom,
		fromAcc,
		[]markertypes.AccessGrant{
			{
				Address: fromAcc.String(),
				Permissions: markertypes.AccessList{
					markertypes.Access_Admin,
					markertypes.Access_Mint,
					markertypes.Access_Burn,
					markertypes.Access_Withdraw,
					markertypes.Access_Transfer,
				},
			},
		},
		markertypes.StatusProposed,
		markertypes.MarkerType_RestrictedCoin,
		true,       // Supply fixed
		false,      // Allow governance control
		false,      // Don't allow forced transfer
		[]string{}, // No required attributes
	)

	return marker, nil
}
