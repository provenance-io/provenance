package types

// NewEventAssetBurned creates a new EventAssetBurned event.
func NewEventAssetBurned(classID, id, owner string) *EventAssetBurned {
	return &EventAssetBurned{
		ClassId: classID,
		Id:      id,
		Owner:   owner,
	}
}

// NewEventAssetClassCreated creates a new EventAssetClassCreated event.
func NewEventAssetClassCreated(classID, className, classSymbol string) *EventAssetClassCreated {
	return &EventAssetClassCreated{
		ClassId:     classID,
		ClassName:   className,
		ClassSymbol: classSymbol,
	}
}

// NewEventAssetCreated creates a new EventAssetCreated event.
func NewEventAssetCreated(classID, id, owner string) *EventAssetCreated {
	return &EventAssetCreated{
		ClassId: classID,
		Id:      id,
		Owner:   owner,
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
func NewEventTokenizationCreated(tokenization, classID, id, owner string) *EventTokenizationCreated {
	return &EventTokenizationCreated{
		Tokenization: tokenization,
		ClassId:      classID,
		Id:           id,
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
