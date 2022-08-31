package types

func NewEventExpirationAdd(moduleAssetID string) *EventExpirationAdd {
	return &EventExpirationAdd{moduleAssetID}
}

func NewEventExpirationExtend(moduleAssetID string) *EventExpirationExtend {
	return &EventExpirationExtend{moduleAssetID}
}

func NewEventExpirationDelete(moduleAssetID string) *EventExpirationDelete {
	return &EventExpirationDelete{moduleAssetID}
}
