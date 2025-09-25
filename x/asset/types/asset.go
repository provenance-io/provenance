package types

import (
	"encoding/json"
	"errors"
	"fmt"

	registrytypes "github.com/provenance-io/provenance/x/registry/types"
)

const (
	MaxLenAssetClassID = 128             // 128 characters
	MaxLenAssetID      = 128             // 128 characters
	MaxLenName         = 128             // 128 characters
	MaxLenURI          = 512             // 256 characters
	MaxLenURIHash      = 128             // 128 characters
	MaxLenData         = 5 * 1024 * 1024 // 5MB
)

// Validate validates the Asset type fields.
func (a *Asset) Validate() error {
	if a == nil {
		return fmt.Errorf("asset cannot be nil")
	}

	var errs []error
	if err := registrytypes.ValidateClassID(a.ClassId); err != nil {
		errs = append(errs, fmt.Errorf("invalid class_id: %w", err))
	}

	if err := registrytypes.ValidateNftID(a.Id); err != nil {
		errs = append(errs, fmt.Errorf("invalid id: %w", err))
	}

	if err := registrytypes.ValidateStringLength(a.Uri, 0, MaxLenURI); err != nil {
		errs = append(errs, fmt.Errorf("invalid uri: %w", err))
	}

	if err := registrytypes.ValidateStringLength(a.UriHash, 0, MaxLenURIHash); err != nil {
		errs = append(errs, fmt.Errorf("invalid uri_hash: %w", err))
	}

	if err := registrytypes.ValidateStringLength(a.Data, 0, MaxLenData); err != nil {
		errs = append(errs, fmt.Errorf("invalid data: %w", err))
	}

	if err := validateJSON(a.Data); err != nil {
		errs = append(errs, fmt.Errorf("invalid data: %w", err))
	}

	return errors.Join(errs...)
}

// Validate validates the AssetClass type fields.
func (c *AssetClass) Validate() error {
	if c == nil {
		return fmt.Errorf("asset_class cannot be nil")
	}

	var errs []error
	if err := registrytypes.ValidateClassID(c.Id); err != nil {
		errs = append(errs, fmt.Errorf("invalid id: %w", err))
	}

	if err := registrytypes.ValidateStringLength(c.Name, 1, MaxLenName); err != nil {
		errs = append(errs, fmt.Errorf("invalid name: %w", err))
	}

	if err := registrytypes.ValidateStringLength(c.Uri, 0, MaxLenURI); err != nil {
		errs = append(errs, fmt.Errorf("invalid uri: %w", err))
	}

	if err := registrytypes.ValidateStringLength(c.UriHash, 0, MaxLenURIHash); err != nil {
		errs = append(errs, fmt.Errorf("invalid uri_hash: %w", err))
	}

	if err := registrytypes.ValidateStringLength(c.Data, 0, MaxLenData); err != nil {
		errs = append(errs, fmt.Errorf("invalid data: %w", err))
	}

	if err := validateJSONSchema(c.Data); err != nil {
		errs = append(errs, fmt.Errorf("invalid data: %w", err))
	}

	return errors.Join(errs...)
}

// Validate validates the AssetKey type fields.
func (k *AssetKey) Validate() error {
	if k == nil {
		return fmt.Errorf("asset_key cannot be nil")
	}

	var errs []error
	if err := registrytypes.ValidateClassID(k.ClassId); err != nil {
		errs = append(errs, fmt.Errorf("invalid class_id: %w", err))
	}

	if err := registrytypes.ValidateNftID(k.Id); err != nil {
		errs = append(errs, fmt.Errorf("invalid id: %w", err))
	}

	return errors.Join(errs...)
}

func validateJSON(data string) error {
	if data == "" {
		return nil // Empty data is valid
	}

	var jsonData any
	if err := json.Unmarshal([]byte(data), &jsonData); err != nil {
		return fmt.Errorf("invalid JSON data: %w", err)
	}

	return nil
}

// validateJSONSchema validates that the provided string is a well-formed JSON schema
func validateJSONSchema(data string) error {
	if data == "" {
		return nil // Empty data is valid
	}

	// Try to parse the data as JSON
	var jsonData any
	if err := json.Unmarshal([]byte(data), &jsonData); err != nil {
		return fmt.Errorf("invalid JSON data: %w", err)
	}

	// Check if it's a JSON schema by looking for required schema properties
	schemaMap, ok := jsonData.(map[string]any)
	if !ok {
		return fmt.Errorf("not a JSON object")
	}

	// Check for type property which is required in JSON schemas
	if _, hasType := schemaMap["type"]; !hasType {
		return fmt.Errorf("type is required")
	}

	return nil
}
