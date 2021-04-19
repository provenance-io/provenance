package types

const (
	// EventAttributeMarkerKey is the attribute key for a marker.
	EventAttributeMarkerKey string = "marker"
	// EventAttributeDenomKey is the attribute key for a marker.
	EventAttributeDenomKey string = "denom"
	// EventAttributeAmountKey is the attribute key for a marker.
	EventAttributeAmountKey string = "amount"
	// EventAttributeAdministratorKey is the attribute for the admin invoking the event action
	EventAttributeAdministratorKey string = "administrator"
	// EventAttributeMarkerStatusKey is the attribute that holds the current status of the marker
	EventAttributeMarkerStatusKey string = "status"
	// EventAttributeMarkerTypeKey is the attribute that holds the type of the marker
	EventAttributeMarkerTypeKey string = "type"

	// EventAttributeGrantKey is the attribute key for an access grant.
	EventAttributeGrantKey string = "marker_AccessGrant"
	// EventAttributeRevokeKey is the attribute key for a revoke event
	EventAttributeRevokeKey string = "marker_access_revoked"

	// EventAttributeModuleNameKey is the attribute key for the entire marker module
	EventAttributeModuleNameKey string = "module"

	// EventTypeMarkerAdded emitted when marker added
	EventTypeMarkerAdded string = EventAttributeMarkerKey + "_added"
	// EventTypeMarkerUpdated emitted when marker updated
	EventTypeMarkerUpdated string = EventAttributeMarkerKey + "_updated"

	// EventTypeGrantAccess emitted when access grant made for user against marker
	EventTypeGrantAccess string = EventAttributeMarkerKey + "_AccessGranted"
	// EventTypeRevokeAccess emitted when access grant removed for user against marker
	EventTypeRevokeAccess string = EventAttributeMarkerKey + "_access_revoked"

	// EventTypeFinalize emitted when a marker configuration is finalized
	EventTypeFinalize string = EventAttributeMarkerKey + "_finalized"
	// EventTypeActivate emitted when a marker configuration is finalized
	EventTypeActivate string = EventAttributeMarkerKey + "_activated"
	// EventTypeCancel emitted when a marker configuration is finalized
	EventTypeCancel string = EventAttributeMarkerKey + "_cancelled"
	// EventTypeDestroy emitted when a marker is destroyed and marked for deletion
	EventTypeDestroy string = EventAttributeMarkerKey + "_destroyed"

	// EventTypeMint emitted when a marker has coins minted against it
	EventTypeMint string = EventAttributeMarkerKey + "_minted_coins"
	// EventTypeBurn emitted when a marker has coins burned from it
	EventTypeBurn string = EventAttributeMarkerKey + "_burned_coins"
	// EventTypeWithdraw emitted when an administrator withdraws coins from marker
	EventTypeWithdraw string = EventAttributeMarkerKey + "_withdraw_coins"
	// EventTypeTransfer emitted when a restricted coin marker transfer occurs
	EventTypeTransfer string = EventAttributeMarkerKey + "_tranfer_coin"

	// EventTypeDepositAsset emitted when assets are assigned as marker collateral
	EventTypeDepositAsset string = EventAttributeMarkerKey + "_asset_deposited"
	// EventTypeWithdrawAsset emitted when assets are removed from marker collateral
	EventTypeWithdrawAsset string = EventAttributeMarkerKey + "_asset_withdrawn"

	EventTelemetryAddress            string = "address"
	EventTelemetryLabelAmount        string = "amount"
	EventTelemetryLabelDenom         string = "denom"
	EventTelemetryLabelStatus        string = "status"
	EventTelemetryLabelManager       string = "manager"
	EventTelemetryLabelAdministrator string = "administrator"
	EventTelemetryLabelMarkerType    string = "marker-type"
	EventTelemetryLabelAccess        string = "access"
)

func NewEventMarkerAdd(denom string, amount string, status string, manager string, markerType string) *EventMarkerAdd {
	return &EventMarkerAdd{
		Denom:      denom,
		Amount:     amount,
		Status:     status,
		Manager:    manager,
		MarkerType: markerType,
	}
}

func NewEventMarkerAddAccess(accessGrant AccessGrant, denom string, administrator string) *EventMarkerAddAccess {

	permissions := make([]string, len(accessGrant.Permissions))
	for i, permission := range accessGrant.Permissions {
		permissions[i] = permission.String()
	}

	access := EventMarkerAccess{
		Address:     accessGrant.Address,
		Permissions: permissions,
	}

	return &EventMarkerAddAccess{
		Access:        access,
		Denom:         denom,
		Administrator: administrator,
	}
}

func NewEventMarkerDeleteAccess(removeAddress string, denom string, administrator string) *EventMarkerDeleteAccess {
	return &EventMarkerDeleteAccess{
		RemoveAddress: removeAddress,
		Denom:         denom,
		Administrator: administrator,
	}
}
