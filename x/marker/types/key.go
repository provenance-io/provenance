package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
)

const (
	// ModuleName is the name of the module
	ModuleName = "marker"

	// StoreKey is string representation of the store key for marker
	StoreKey = ModuleName

	// RouterKey to be used for routing msgs
	RouterKey = ModuleName

	// QuerierRoute to be used for queries
	QuerierRoute = ModuleName

	// CoinPoolName to be used for coin pool associated with mint/burn activities.
	CoinPoolName = ModuleName

	// DefaultParamspace is the name used for the parameter subspace for this module.
	DefaultParamspace = ModuleName

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

	// EventTypeMarkerAdded emitted when marker addded
	EventTypeMarkerAdded string = EventAttributeMarkerKey + "_added"

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
)

var (
	// MarkerStoreKeyPrefix prefix for marker-address reference (improves iterator performance over auth accounts)
	MarkerStoreKeyPrefix = []byte{0x01}
)

// MarkerAddress returns the module account address for the given denomination
func MarkerAddress(denom string) (sdk.AccAddress, error) {
	if err := sdk.ValidateDenom(denom); err != nil {
		return nil, err
	}
	return sdk.AccAddress(crypto.AddressHash([]byte(fmt.Sprintf("%s/%s", ModuleName, denom)))), nil
}

// MustGetMarkerAddress returns the module account address for the given denomination, panics on error
func MustGetMarkerAddress(denom string) sdk.AccAddress {
	addr, err := MarkerAddress(denom)
	if err != nil {
		panic(err)
	}
	return addr
}

// MarkerStoreKey turn an address to key used to get it from the account store
func MarkerStoreKey(addr sdk.AccAddress) []byte {
	return append(MarkerStoreKeyPrefix, addr.Bytes()...)
}

// SplitMarkerStoreKey returns an account address given a store key
func SplitMarkerStoreKey(key []byte) sdk.AccAddress {
	return sdk.AccAddress(key[1 : sdk.AddrLen+1])
}
