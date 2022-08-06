package types

const (
	// EventTypeExpirationAdd is the type of event generated when a module asset expiration is added.
	EventTypeExpirationAdd string = "expiration_add"

	// EventTypeExpirationExtend is the type of event generated when a module asset expiration is extended.
	EventTypeExpirationExtend string = "expiration_extend"

	// EventTypeExpirationDelete is the type of event generated when a module asset expiration is deleted.
	EventTypeExpirationDelete string = "expiration_delete"

	// todo: do we need some event attributes here?
)

func NewEventExpirationAdd(moduleAssetId string) *EventExpirationAdd {
	return &EventExpirationAdd{moduleAssetId}
}

func NewEventExpirationExtend(moduleAssetId string) *EventExpirationExtend {
	return &EventExpirationExtend{moduleAssetId}
}

func NewEventExpirationDelete(moduleAssetId string) *EventExpirationDelete {
	return &EventExpirationDelete{moduleAssetId}
}
