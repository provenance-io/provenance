package types

import (
	"fmt"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	// EventAttributeDenomKey is the attribute key for a marker.
	EventAttributeDenomKey string = "denom"

	// EventTypeDestroy emitted when a marker is deleted during abci-begin_event
	EventTypeDestroy string = "marker_destroyed"

	// EventTelemetryLabelAddress address label for telemetry metrics
	EventTelemetryLabelAddress string = "address"
	// EventTelemetryLabelToAddress to address label for telemetry metrics
	EventTelemetryLabelToAddress string = "to_address"
	// EventTelemetryLabelFromAddress from address label for telemetry metrics
	EventTelemetryLabelFromAddress string = "from_address"
	// EventTelemetryLabelCoins coins label for telemetry metrics
	EventTelemetryLabelCoins string = "coins"
	// EventTelemetryLabelDenom denom label for telemetry metrics
	EventTelemetryLabelDenom string = "denom"
	// EventTelemetryLabelManager manager label for telemetry metrics
	EventTelemetryLabelManager string = "manager"
	// EventTelemetryLabelAdministrator administrator label for telemetry metrics
	EventTelemetryLabelAdministrator string = "administrator"
	// EventTelemetryKeyBurn burn telemetry metrics key
	EventTelemetryKeyBurn string = "burn"
	// EventTelemetryKeyMint mint telemetry metrics key
	EventTelemetryKeyMint string = "mint"
	// EventTelemetryKeyTransfer transfer telemetry metrics key
	EventTelemetryKeyTransfer string = "transfer"
	// EventTelemetryKeyWithdraw withdraw telemetry metrics key
	EventTelemetryKeyWithdraw string = "withdraw"
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

func NewEventMarkerAddAccess(accessGrant AccessGrantI, denom string, administrator string) *EventMarkerAddAccess {
	accessList := accessGrant.GetAccessList()
	permissions := make([]string, len(accessList))
	for i, permission := range accessList {
		permissions[i] = permission.String()
	}

	access := EventMarkerAccess{
		Address:     accessGrant.GetAddress().String(),
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

func NewEventMarkerFinalize(denom string, administrator string) *EventMarkerFinalize {
	return &EventMarkerFinalize{
		Denom:         denom,
		Administrator: administrator,
	}
}

func NewEventMarkerActivate(denom string, administrator string) *EventMarkerActivate {
	return &EventMarkerActivate{
		Denom:         denom,
		Administrator: administrator,
	}
}

func NewEventMarkerCancel(denom string, administrator string) *EventMarkerCancel {
	return &EventMarkerCancel{
		Denom:         denom,
		Administrator: administrator,
	}
}

func NewEventMarkerDelete(denom string, administrator string) *EventMarkerDelete {
	return &EventMarkerDelete{
		Denom:         denom,
		Administrator: administrator,
	}
}

func NewEventMarkerMint(amount string, denom string, administrator string) *EventMarkerMint {
	return &EventMarkerMint{
		Amount:        amount,
		Denom:         denom,
		Administrator: administrator,
	}
}

func NewEventMarkerBurn(amount string, denom string, administrator string) *EventMarkerBurn {
	return &EventMarkerBurn{
		Amount:        amount,
		Denom:         denom,
		Administrator: administrator,
	}
}

func NewEventMarkerWithdraw(coins string, denom string, administrator string, toAddress string) *EventMarkerWithdraw {
	return &EventMarkerWithdraw{
		Coins:         coins,
		Denom:         denom,
		Administrator: administrator,
		ToAddress:     toAddress,
	}
}

func NewEventMarkerTransfer(amount string, denom string, administrator string, toAddress string, fromAddress string) *EventMarkerTransfer {
	return &EventMarkerTransfer{
		Amount:        amount,
		Denom:         denom,
		Administrator: administrator,
		ToAddress:     toAddress,
		FromAddress:   fromAddress,
	}
}

func NewEventMarkerSetDenomMetadata(metadata banktypes.Metadata, administrator string) *EventMarkerSetDenomMetadata {
	metadataDenomUnits := make([]*EventDenomUnit, len(metadata.DenomUnits))
	for i, du := range metadata.DenomUnits {
		denomUnit := EventDenomUnit{
			Denom:    du.Denom,
			Exponent: fmt.Sprint(du.Exponent),
			Aliases:  du.Aliases,
		}
		metadataDenomUnits[i] = &denomUnit
	}
	return &EventMarkerSetDenomMetadata{
		MetadataBase:        metadata.Base,
		MetadataDescription: metadata.Description,
		MetadataDisplay:     metadata.Display,
		MetadataDenomUnits:  metadataDenomUnits,
		MetadataName:        metadata.Name,
		MetadataSymbol:      metadata.Symbol,
		Administrator:       administrator,
	}
}
