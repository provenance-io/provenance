// Package wasm supports smart contract integration with the provenance marker module.
package wasm

import (
	"github.com/provenance-io/provenance/x/marker/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Types in this file were generated using JSON schema:
// https://github.com/provenance-io/provwasm/blob/main/packages/bindings/schema/marker.json
// Naming has been tweaked slightly to remove dup names across types.

// Marker represents a marker account in provwasm supported format.
type Marker struct {
	AccountNumber uint64         `json:"account_number"`
	Address       string         `json:"address"`
	Coins         sdk.Coins      `json:"coins"`
	Denom         string         `json:"denom"`
	Manager       string         `json:"manager"`
	MarkerType    MarkerType     `json:"marker_type"`
	Permissions   []*AccessGrant `json:"permissions,omitempty"`
	Sequence      uint64         `json:"sequence"`
	Status        MarkerStatus   `json:"status"`
	TotalSupply   string         `json:"total_supply"`
	SupplyFixed   bool           `json:"supply_fixed"`
}

// AccessGrant are marker permissions granted to an account.
type AccessGrant struct {
	Address     string             `json:"address"`
	Permissions []MarkerPermission `json:"permissions,omitempty"`
}

// MarkerType defines types of markers.
type MarkerType string

const (
	// MarkerTypeCoin is a concrete marker type
	MarkerTypeCoin MarkerType = "coin"
	// MarkerTypeRestricted is a concrete marker type
	MarkerTypeRestricted MarkerType = "restricted"
	// MarkerTypeUnspecified is a concrete marker type
	MarkerTypeUnspecified MarkerType = "unspecified"
)

// MarkerPermission defines marker permission types.
type MarkerPermission string

const (
	// MarkerPermissionAdmin is a concrete marker permission type
	MarkerPermissionAdmin MarkerPermission = "admin"
	// MarkerPermissionBurn is a concrete marker permission type
	MarkerPermissionBurn MarkerPermission = "burn"
	// MarkerPermissionDelete is a concrete marker permission type
	MarkerPermissionDelete MarkerPermission = "delete"
	// MarkerPermissionDeposit is a concrete marker permission type
	MarkerPermissionDeposit MarkerPermission = "deposit"
	// MarkerPermissionMint is a concrete marker permission type
	MarkerPermissionMint MarkerPermission = "mint"
	// MarkerPermissionTransfer is a concrete marker permission type
	MarkerPermissionTransfer MarkerPermission = "transfer"
	// MarkerPermissionUnspecified is a concrete marker permission type
	MarkerPermissionUnspecified MarkerPermission = "unspecified"
	// MarkerPermissionWithdraw is a concrete marker permission type
	MarkerPermissionWithdraw MarkerPermission = "withdraw"
)

// MarkerStatus defines marker status types.
type MarkerStatus string

const (
	// MarkerStatusActive is a concrete marker status type
	MarkerStatusActive MarkerStatus = "active"
	// MarkerStatusCancelled is a concrete marker status type
	MarkerStatusCancelled MarkerStatus = "cancelled"
	// MarkerStatusDestroyed is a concrete marker status type
	MarkerStatusDestroyed MarkerStatus = "destroyed"
	// MarkerStatusFinalized is a concrete marker status type
	MarkerStatusFinalized MarkerStatus = "finalized"
	// MarkerStatusProposed is a concrete marker status type
	MarkerStatusProposed MarkerStatus = "proposed"
	// MarkerStatusUnspecified is a concrete marker status type
	MarkerStatusUnspecified MarkerStatus = "unspecified"
)

// Convert a core marker type to provwasm supported format.
func createResponseType(input *types.MarkerAccount, balance sdk.Coins) *Marker {
	marker := &Marker{
		AccountNumber: input.GetAccountNumber(),
		Address:       input.GetAddress().String(),
		Coins:         balance,
		Denom:         input.GetDenom(),
		Manager:       input.GetManager().String(),
		MarkerType:    markerTypeFor(input.GetMarkerType()),
		Sequence:      input.GetSequence(),
		Status:        markerStatusFor(input.GetStatus()),
		TotalSupply:   input.GetSupply().Amount.String(),
		SupplyFixed:   input.SupplyFixed,
	}
	for _, ag := range input.GetAccessList() {
		marker.Permissions = append(marker.Permissions, accessGrantFor(ag))
	}
	return marker
}

// Adapt the core marker type to provwasm format.
func markerTypeFor(input types.MarkerType) MarkerType {
	switch input {
	case types.MarkerType_Coin:
		return MarkerTypeCoin
	case types.MarkerType_RestrictedCoin:
		return MarkerTypeRestricted
	default:
		return MarkerTypeUnspecified
	}
}

// Adapt the core marker status to provwasm format.
func markerStatusFor(input types.MarkerStatus) MarkerStatus {
	switch input {
	case types.StatusActive:
		return MarkerStatusActive
	case types.StatusCancelled:
		return MarkerStatusCancelled
	case types.StatusDestroyed:
		return MarkerStatusDestroyed
	case types.StatusFinalized:
		return MarkerStatusFinalized
	case types.StatusProposed:
		return MarkerStatusProposed
	default:
		return MarkerStatusUnspecified
	}
}

// Adapt the core marker access grant type to provwasm format.
func accessGrantFor(input types.AccessGrant) *AccessGrant {
	grant := &AccessGrant{
		Address: input.Address,
	}
	for _, a := range input.GetAccessList() {
		grant.Permissions = append(grant.Permissions, permissionFor(a))
	}
	return grant
}

// Adapt the core marker access type to provwasm format.
func permissionFor(input types.Access) MarkerPermission {
	switch input {
	case types.Access_Admin:
		return MarkerPermissionAdmin
	case types.Access_Burn:
		return MarkerPermissionBurn
	case types.Access_Delete:
		return MarkerPermissionDelete
	case types.Access_Deposit:
		return MarkerPermissionDeposit
	case types.Access_Mint:
		return MarkerPermissionMint
	case types.Access_Transfer:
		return MarkerPermissionTransfer
	case types.Access_Withdraw:
		return MarkerPermissionWithdraw
	default:
		return MarkerPermissionUnspecified
	}
}
