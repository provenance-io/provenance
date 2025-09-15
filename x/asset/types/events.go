package types

// NewEventAssetBurned creates a new EventAssetBurned event.
func NewEventAssetBurned(assetClassID, assetID, owner string) *EventAssetBurned {
	return &EventAssetBurned{
		AssetClassId: assetClassID,
		AssetId:      assetID,
		Owner:        owner,
	}
}

// NewEventAssetClassCreated creates a new EventAssetClassCreated event.
func NewEventAssetClassCreated(assetClassID, assetName, assetSymbol string) *EventAssetClassCreated {
	return &EventAssetClassCreated{
		AssetClassId: assetClassID,
		AssetName:    assetName,
		AssetSymbol:  assetSymbol,
	}
}

// NewEventAssetCreated creates a new EventAssetCreated event.
func NewEventAssetCreated(assetClassID, assetID, owner string) *EventAssetCreated {
	return &EventAssetCreated{
		AssetClassId: assetClassID,
		AssetId:      assetID,
		Owner:        owner,
	}
}

// NewEventPoolCreated creates a new EventPoolCreated event.
func NewEventPoolCreated(pool string, assetCount uint32, owner string) *EventPoolCreated {
	return &EventPoolCreated{
		Pool:       pool,
		AssetCount: assetCount,
		Owner:      owner,
	}
}

// NewEventTokenizationCreated creates a new EventTokenizationCreated event.
func NewEventTokenizationCreated(tokenization, assetClassID, assetID, owner string) *EventTokenizationCreated {
	return &EventTokenizationCreated{
		Tokenization: tokenization,
		AssetClassId: assetClassID,
		AssetId:      assetID,
		Owner:        owner,
	}
}

// NewEventSecuritizationCreated creates a new EventSecuritizationCreated event.
func NewEventSecuritizationCreated(securitizationID string, trancheCount, poolCount uint32, owner string) *EventSecuritizationCreated {
	return &EventSecuritizationCreated{
		SecuritizationId: securitizationID,
		TrancheCount:     trancheCount,
		PoolCount:        poolCount,
		Owner:            owner,
	}
}
