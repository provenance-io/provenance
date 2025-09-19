package types

func (a *Asset) Validate() error {
	if a == nil {
		return NewErrCodeInvalidField("asset", "asset cannot be nil")
	}

	if len(a.ClassId) == 0 {
		return NewErrCodeMissingField("class_id")
	}

	if len(a.Id) == 0 {
		return NewErrCodeMissingField("id")
	}

	if err := validateJSON(a.Data); err != nil {
		return NewErrCodeInvalidField("data", err.Error())
	}

	// There is nothing to validate with the Uri, or UriHash.
	return nil
}

func (c *AssetClass) Validate() error {
	if c == nil {
		return NewErrCodeInvalidField("asset_class", "asset class cannot be nil")
	}

	if len(c.Id) == 0 {
		return NewErrCodeMissingField("id")
	}

	if len(c.Name) == 0 {
		return NewErrCodeMissingField("name")
	}

	if err := validateJSONSchema(c.Data); err != nil {
		return NewErrCodeInvalidField("data", err.Error())
	}

	// There is nothing to validate for the Symbol, Description, Uri, or UriHash.
	return nil
}

func (k *AssetKey) Validate() error {
	if k == nil {
		return NewErrCodeInvalidField("asset_key", "asset key cannot be nil")
	}

	if len(k.ClassId) == 0 {
		return NewErrCodeMissingField("class_id")
	}

	if len(k.Id) == 0 {
		return NewErrCodeMissingField("id")
	}

	return nil
}
