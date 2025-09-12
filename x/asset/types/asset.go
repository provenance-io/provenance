package types

import (
	"errors"
	"fmt"
)

func (a *Asset) Validate() error {
	if a == nil {
		return errors.New("asset cannot be nil")
	}

	if len(a.ClassId) == 0 {
		return errors.New("class id cannot be empty")
	}

	if len(a.Id) == 0 {
		return errors.New("id cannot be empty")
	}

	if err := validateJSON(a.Data); err != nil {
		return fmt.Errorf("invalid data: %w", err)
	}

	// There is nothing to validate with the Uri, or UriHash.
	return nil
}

func (c *AssetClass) Validate() error {
	if c == nil {
		return errors.New("asset class cannot be nil")
	}

	if len(c.Id) == 0 {
		return errors.New("id cannot be empty")
	}

	if len(c.Name) == 0 {
		return errors.New("name cannot be empty")
	}

	if err := validateJSONSchema(c.Data); err != nil {
		return fmt.Errorf("invalid data: %w", err)
	}

	// There is nothing to validate for the Symbol, Description, Uri, or UriHash.
	return nil
}

func (k *AssetKey) Validate() error {
	if k == nil {
		return errors.New("asset key cannot be nil")
	}

	if len(k.ClassId) == 0 {
		return errors.New("class id cannot be empty")
	}

	if len(k.Id) == 0 {
		return errors.New("id cannot be empty")
	}

	return nil
}
