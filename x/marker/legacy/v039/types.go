package v039

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v038auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v038"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/legacy/v039"
	"github.com/tendermint/tendermint/crypto"
	"gopkg.in/yaml.v2"
)

const (
	AccessMint     = "mint"
	AccessBurn     = "burn"
	AccessDeposit  = "deposit"
	AccessWithdraw = "withdraw"
	AccessDelete   = "delete"
	AccessAdmin    = "grant"

	AllPermissions    = "mint,burn,deposit,withdraw,delete,grant"
	SupplyPermissions = "mint,burn"
	AssetPermissions  = "deposit,withdraw"
	ModuleName        = "marker"
)

// MarkerAccount defines a marker account interface for modules that interact with markers
type MarkerAccountI interface {
	v038auth.Account

	Validate() error

	GetDenom() string
	GetManager() sdk.AccAddress
	GetMarkerType() string

	GetStatus() string
	SetStatus(string) error

	GetSupply() sdk.Coin
	SetSupply(sdk.Coin) error

	GrantAccess(AccessGrant) error
	RevokeAccess(sdk.AccAddress) error

	AddressHasPermission(sdk.AccAddress, string) bool
	AddressListForPermission(string) []sdk.AccAddress
}

// AccessGrant defines an interface for interacting with roles assigned to a given address.
type AccessGrantI interface {
	Validate() error
	GetAddress() sdk.AccAddress

	HasPermission(string) bool
	GetPermissions() []string

	AddPermission(string) error
	RemovePermission(string) error

	MergeAdd(AccessGrant) error
	MergeRemove(AccessGrant) error
}

// GenesisState is the initial marker module state.
type GenesisState struct {
	Markers []MarkerAccount `json:"markers"`
}

// AccessGrant is a structure for assigning a set of marker management permissions to an address
type AccessGrant struct {
	Permissions []string       `json:"permissions" yaml:"permissions"`
	Address     sdk.AccAddress `json:"address" yaml:"address"`
}

// MarkerAssets is a list of scope ids that a given address (of a marker) is associated with
type MarkerAssets struct {
	// Address of the marker
	Address sdk.AccAddress `json:"address" yaml:"address"`
	// List of scope uuids that have the marker as a party member
	ScopeID []string `json:"scope_id" yaml:"scope_id"`
}

// MarkerAccount is a configuration structure that defines a Token and the resulting supply of coins.
type MarkerAccount struct {
	*authtypes.BaseAccount

	// Address that owns the marker configuration.  This account must sign any requests
	// to change marker config (only valid for statuses prior to finalization)
	Manager sdk.AccAddress `json:"manager,omitempty" yaml:"manager"`
	// Access control lists
	AccessControls []AccessGrant `json:"accesscontrol,omitempty" yaml:"accesscontrol"`

	// Indicates the current status of this marker record.
	Status MarkerStatus `json:"status" yaml:"status"`

	// value denomination and total supply for the token.
	Denom  string  `json:"denom" yaml:"denom"`
	Supply sdk.Int `json:"total_supply" yaml:"total_supply"`
	// Marker type information
	MarkerType string `json:"type,omitempty" yaml:"type"`
}

// NewEmptyMarkerAccount creates a new empty marker account in a Proposed state
func NewEmptyMarkerAccount(denom string, grants []AccessGrant) *MarkerAccount {
	baseAcc := authtypes.NewBaseAccountWithAddress(MustGetMarkerAddress(denom))
	return &MarkerAccount{
		BaseAccount:    &baseAcc,
		AccessControls: grants,
		Denom:          denom,
		Supply:         sdk.ZeroInt(),
		Status:         StatusProposed,
		MarkerType:     "COIN",
	}
}

// NewMarkerAccount creates a marker account initialized over a given base account.
func NewMarkerAccount(
	baseAcc *authtypes.BaseAccount,
	totalSupply sdk.Coin,
	accessControls []AccessGrant,
	status MarkerStatus,
	markerType string,
) *MarkerAccount {
	return &MarkerAccount{
		BaseAccount:    baseAcc,
		Denom:          totalSupply.Denom,
		Supply:         totalSupply.Amount,
		AccessControls: accessControls,
		Status:         status,
		MarkerType:     markerType,
	}
}

// GetDenom the denomination of the coin associated with this marker
func (ma MarkerAccount) GetDenom() string { return ma.Denom }

// AddressHasPermission returns true if the provided address has been assigned the provided
// role within the current MarkerAccount AccessControls
func (ma *MarkerAccount) AddressHasPermission(addr sdk.AccAddress, role string) bool {
	for _, g := range ma.AccessControls {
		if g.Address.Equals(addr) && g.HasPermission(role) {
			return true
		}
	}
	return false
}

// AddressListForPermission returns a list of all addresses with the provided rule within the
// current MarkerAccount AccessControls list
func (ma *MarkerAccount) AddressListForPermission(role string) []sdk.AccAddress {
	var addressList []sdk.AccAddress

	for _, g := range ma.AccessControls {
		if g.HasPermission(role) {
			addressList = append(addressList, g.GetAddress())
		}
	}
	return addressList
}

// Validate performs minimal sanity checking over the current MarkerAccount instance
func (ma MarkerAccount) Validate() error {
	if !ValidMarkerStatus(ma.Status) {
		return fmt.Errorf("invalid marker status")
	}
	// unlikely as this is set using Coin which prohibits negative values.
	if ma.Supply.IsNegative() {
		return fmt.Errorf("total supply must be greater than or equal to zero")
	}
	if ma.Status != StatusProposed && len(ma.AddressListForPermission(AccessMint)) == 0 && ma.Supply.IsZero() {
		return fmt.Errorf("cannot create a marker with zero total supply and no authorization for minting more")
	}
	// unlikely as this is set using a Coin which prohibits this value.
	if strings.TrimSpace(ma.Denom) == "" {
		return fmt.Errorf("marker denom cannot be empty")
	}
	markerAddress, err := MarkerAddress(ma.Denom)
	if err != nil {
		return fmt.Errorf("marker denom is invalid, only 3-16 lowercase letters or numbers are allowed")
	}
	if !ma.BaseAccount.Address.Equals(markerAddress) {
		return fmt.Errorf("address %s cannot be derived from the marker denom '%s'", ma.Address, ma.Denom)
	}
	if err := ValidateGrants(ma.AccessControls...); err != nil {
		return fmt.Errorf("invalid access privileges granted: %v", err)
	}
	selfGrant := GrantsForAddress(ma.Address, ma.AccessControls...).GetPermissions()
	if len(selfGrant) > 0 {
		return fmt.Errorf("permissions cannot be granted to '%s' marker account: %v", ma.Denom, selfGrant)
	}
	return ma.BaseAccount.Validate()
}

// SetPubKey implements authtypes.Account (but there are no public keys associated with the account for signing)
func (ma *MarkerAccount) SetPubKey(pubKey cryptotypes.PubKey) error {
	return fmt.Errorf("not supported for marker accounts")
}

// SetSequence implements authtypes.Account (but you can't set a sequence as you can't sign tx for this account)
func (ma *MarkerAccount) SetSequence(seq uint64) error {
	return fmt.Errorf("not supported for marker accounts")
}

// GetStatus returns the status of the marker account.
func (ma MarkerAccount) GetStatus() string {
	return ma.Status.String()
}

// SetStatus sets the status of the marker to the provided value.
func (ma *MarkerAccount) SetStatus(newStatus string) error {
	status, err := MarkerStatusFromString(newStatus)
	if err != nil {
		return fmt.Errorf("error invalid marker status %s is not a known status type", newStatus)
	}
	if status == StatusActive {
		// When activated the manager property is no longer valid so clear it
		ma.Manager = sdk.AccAddress([]byte{})
	}

	ma.Status = status
	return nil
}

// GetMarkerType returns the type of the marker account.
func (ma MarkerAccount) GetMarkerType() string {
	return ma.MarkerType
}

// GetManager returns the address of the account that is responsible for the proposed marker.
func (ma MarkerAccount) GetManager() sdk.AccAddress {
	return ma.Manager
}

// SetManager sets the manager/owner address for proposed marker accounts
func (ma *MarkerAccount) SetManager(manager sdk.AccAddress) error {
	if !manager.Empty() && ma.Status != StatusProposed {
		return fmt.Errorf("manager address is only valid for proposed markers, use access grants instead")
	}
	ma.Manager = manager
	return nil
}

// SetSupply sets the total supply amount to track
func (ma *MarkerAccount) SetSupply(total sdk.Coin) error {
	if total.Denom != ma.Denom {
		return fmt.Errorf("supply coin denom must match marker denom")
	}
	ma.Supply = total.Amount
	return nil
}

// GetSupply implements authtypes.Account
func (ma MarkerAccount) GetSupply() sdk.Coin {
	return sdk.NewCoin(ma.Denom, ma.Supply)
}

// GrantAccess appends the access grant to the marker account.
func (ma *MarkerAccount) GrantAccess(access AccessGrant) error {
	if err := access.Validate(); err != nil {
		return fmt.Errorf(err.Error())
	}
	// Find any existing permissions and append specified permissions
	for _, ac := range ma.AccessControls {
		if ac.Address.Equals(access.GetAddress()) {
			if err := access.MergeAdd(*NewAccessGrant(ac.GetAddress(), ac.GetPermissions())); err != nil {
				return err
			}
		}
	}
	// Revoke existing (no errors from this as we have validated above)
	if err := ma.RevokeAccess(access.GetAddress()); err != nil {
		return err
	}
	// Append the new record
	ma.AccessControls = append(ma.AccessControls, *NewAccessGrant(access.GetAddress(), access.GetPermissions()))
	return nil
}

// RevokeAccess removes any AccessGrant for the given address on this marker.
func (ma *MarkerAccount) RevokeAccess(addr sdk.AccAddress) error {
	if err := sdk.VerifyAddressFormat(addr); err != nil {
		return fmt.Errorf("can not revoke access for invalid address")
	}

	var accessList []AccessGrant
	for _, ac := range ma.AccessControls {
		if !ac.Address.Equals(addr) {
			accessList = append(accessList, ac)
		}
	}

	ma.AccessControls = accessList
	return nil
}

// markerAccountPretty suppresses the public key value during json serialization
type markerAccountPretty struct {
	Address        sdk.AccAddress `json:"address" yaml:"address"`
	Coins          sdk.Coins      `json:"coins" yaml:"coins"`
	PubKey         string         `json:"public_key,omitempty" yaml:"public_key,omitempty"`
	AccountNumber  uint64         `json:"account_number" yaml:"account_number"`
	Sequence       uint64         `json:"sequence" yaml:"sequence"`
	Manager        sdk.AccAddress `json:"manager,omitempty" yaml:"manager,omitempty"`
	AccessControls []AccessGrant  `json:"permissions" yaml:"permissions"`
	Status         MarkerStatus   `json:"status" yaml:"status"`
	Denom          string         `json:"denom" yaml:"denom"`
	Supply         sdk.Int        `json:"total_supply" yaml:"total_supply"`
	MarkerType     string         `json:"marker_type" yaml:"marker_type"`
}

func (ma MarkerAccount) String() string {
	out, _ := ma.MarshalYAML()
	return out.(string)
}

// MarshalYAML returns the YAML representation of a MarkerAccount.
func (ma MarkerAccount) MarshalYAML() (interface{}, error) {
	bs, err := yaml.Marshal(markerAccountPretty{
		Address:        ma.Address,
		Coins:          ma.Coins,
		PubKey:         "",
		AccountNumber:  ma.AccountNumber,
		Sequence:       ma.Sequence,
		Manager:        ma.Manager,
		AccessControls: ma.AccessControls,
		Status:         ma.Status,
		Denom:          ma.Denom,
		Supply:         ma.Supply,
		MarkerType:     ma.MarkerType,
	})

	if err != nil {
		return nil, err
	}

	return string(bs), nil
}

// MarshalJSON returns the JSON representation of a MarkerAccount.
func (ma MarkerAccount) MarshalJSON() ([]byte, error) {
	return json.Marshal(markerAccountPretty{
		Address:        ma.Address,
		Coins:          ma.Coins,
		PubKey:         "",
		AccountNumber:  ma.AccountNumber,
		Sequence:       ma.Sequence,
		Manager:        ma.Manager,
		AccessControls: ma.AccessControls,
		Status:         ma.Status,
		Denom:          ma.Denom,
		Supply:         ma.Supply,
		MarkerType:     ma.MarkerType,
	})
}

// UnmarshalJSON un-marshals raw JSON bytes into a MarkerAccount.
func (ma *MarkerAccount) UnmarshalJSON(bz []byte) error {
	var alias markerAccountPretty
	if err := json.Unmarshal(bz, &alias); err != nil {
		return err
	}

	ma.BaseAccount = authtypes.NewBaseAccount(alias.Address, alias.Coins, nil, alias.AccountNumber, alias.Sequence)
	ma.Manager = alias.Manager
	ma.AccessControls = alias.AccessControls
	ma.Status = alias.Status
	ma.Denom = alias.Denom
	ma.Supply = alias.Supply
	ma.MarkerType = alias.MarkerType
	return nil
}

// Equals returns true if this MarkerAccount is equal to other MarkerAccount in all properties
func (ma MarkerAccount) Equals(other MarkerAccount) bool {
	grantsEqual := true
	for _, g := range ma.AccessControls {
		if !g.Equals(GrantsForAddress(g.Address, other.AccessControls...)) {
			grantsEqual = false
		}
	}
	return grantsEqual &&
		other.Address.Equals(ma.Address) &&
		other.Coins.IsEqual(ma.Coins) &&
		other.AccountNumber == ma.AccountNumber &&
		other.Sequence == ma.Sequence &&
		other.Manager.Equals(ma.Manager) &&
		other.Status.String() == ma.Status.String() &&
		other.Denom == ma.Denom &&
		other.Supply.Equal(ma.Supply) &&
		other.MarkerType == ma.MarkerType
}

// NewAccessGrant creates a new AccessGrant object
func NewAccessGrant(address sdk.AccAddress, permissions []string) *AccessGrant {
	return &AccessGrant{
		Permissions: permissions,
		Address:     address,
	}
}

// ValidateGrants checks a collection of grants and returns any errors encountered or nil
func ValidateGrants(grants ...AccessGrant) error {
	for _, grant := range grants {
		if err := grant.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// GrantsForAddress return
func GrantsForAddress(account sdk.AccAddress, grants ...AccessGrant) AccessGrant {
	for _, grant := range grants {
		if grant.Address.Equals(account) {
			return grant
		}
	}
	return AccessGrant{nil, account}
}

// Validate performs checks to ensure this acccess grant is properly formed.
func (ag AccessGrant) Validate() error {
	if ag.Address.Empty() {
		return fmt.Errorf("address can not be empty")
	}
	return validateGranted(ag.Permissions...)
}

// HasPermission returns true if the AccessGrant allows the given permission
func (ag AccessGrant) HasPermission(permission string) bool {
	// Empty addresses can have no permissions.
	if ag.Address.Empty() {
		return false
	}
	for _, p := range ag.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// AddPermission adds the given permission to this grant if not included
func (ag *AccessGrant) AddPermission(permission string) error {
	if err := validateGranted(permission); err != nil {
		return err
	}
	if !ag.HasPermission(permission) {
		ag.Permissions = append(ag.Permissions, permission)
	}
	return nil
}

// RemovePermission removes the given permission from this grant (if included)
func (ag *AccessGrant) RemovePermission(permission string) error {
	if err := validateGranted(permission); err != nil {
		return err
	}
	if ag.HasPermission(permission) {
		var newPerms []string
		for _, p := range ag.Permissions {
			if p != permission {
				newPerms = append(newPerms, p)
			}
		}
		ag.Permissions = newPerms
	}
	return nil
}

// MergeAdd looks for any missing permissions in the given grant and adds them to this instance.
func (ag *AccessGrant) MergeAdd(other AccessGrant) error {
	if err := other.Validate(); err != nil {
		return err
	}
	if !other.GetAddress().Equals(ag.Address) {
		return fmt.Errorf("cannot merge in AccessGrant for different address")
	}
	for _, p := range other.GetPermissions() {
		if !ag.HasPermission(p) {
			ag.Permissions = append(ag.Permissions, p)
		}
	}
	return nil
}

// MergeRemove looks for permissions in this instance that exist in the given grant and removes them.
func (ag *AccessGrant) MergeRemove(other AccessGrant) error {
	if err := other.Validate(); err != nil {
		return err
	}
	if !other.GetAddress().Equals(ag.Address) {
		return fmt.Errorf("cannot merge in AccessGrant for different address")
	}
	var newPerms []string
	for _, p := range ag.Permissions {
		if !other.HasPermission(p) {
			newPerms = append(newPerms, p)
		}
	}
	ag.Permissions = newPerms
	return nil
}

// GetAddress returns the address the AccessGrant is for
func (ag AccessGrant) GetAddress() sdk.AccAddress {
	return ag.Address
}

// GetPermissions returns the permissions granted to the address
func (ag AccessGrant) GetPermissions() []string {
	return ag.Permissions
}

// Equals returns true if both AccessGrants has the same address and list of permissions.
func (ag AccessGrant) Equals(other AccessGrant) bool {
	// same address and same number of permissions...
	areEqual := ag.Address.Equals(other.Address) &&
		len(ag.Permissions) == len(other.Permissions)
	// and other has all the same permissiosn that we have...
	for _, p := range ag.Permissions {
		if !other.HasPermission(p) {
			areEqual = false
		}
	}
	return areEqual
}

// performs basic permission validation
func validateGranted(permissions ...string) error {
	for _, permission := range permissions {
		for _, c := range permission {
			// this check protects against injecting commas (or anything else unexpected) which could break our list processing
			if c < 'a' || c > 'z' {
				return fmt.Errorf("invalid permission only lowercase letters are supported: '%s'", permission)
			}
		}

		if strings.TrimSpace(permission) == "" {
			return fmt.Errorf("access permission is empty")
		}
		if !strings.Contains(AllPermissions, permission) {
			return fmt.Errorf("access permission [%s] is not a valid permission", permission)
		}
	}
	return nil
}

// MarkerStatus defines the status type of the marker record
type MarkerStatus byte

// Marker state types
const (
	// Invalid/uninitialized
	StatusUndefined MarkerStatus = 0x00

	// Initial configuration period, updates allowed, token supply not created.
	StatusProposed MarkerStatus = 0x01

	// Configuration finalized, ready for supply creation
	StatusFinalized MarkerStatus = 0x02

	// Supply is created, rules are in force.
	StatusActive MarkerStatus = 0x03

	// Marker has been cancelled, pending destroy
	StatusCancelled MarkerStatus = 0x04

	// Marker supply has all been recalled, marker is considered destroyed and no further actions allowed.
	StatusDestroyed MarkerStatus = 0x05
)

// MustGetMarkerStatus turns the string into a MarkerStatus typed value ... panics if invalid.
func MustGetMarkerStatus(str string) MarkerStatus {
	s, err := MarkerStatusFromString(str)
	if err != nil {
		panic(err)
	}
	return s
}

// MarkerStatusFromString returns a MarkerStatus from a string. It returns an error
// if the string is invalid.
func MarkerStatusFromString(str string) (MarkerStatus, error) {
	switch strings.ToLower(str) {
	case "undefined":
		return StatusUndefined, nil
	case "proposed":
		return StatusProposed, nil
	case "finalized":
		return StatusFinalized, nil
	case "active":
		return StatusActive, nil
	case "cancelled":
		return StatusCancelled, nil
	case "destroyed":
		return StatusDestroyed, nil

	default:
		return MarkerStatus(0xff), fmt.Errorf("'%s' is not a valid marker status", str)
	}
}

// ValidMarkerStatus returns true if the marker status is valid and false otherwise.
func ValidMarkerStatus(markerStatus MarkerStatus) bool {
	if markerStatus == StatusProposed ||
		markerStatus == StatusFinalized ||
		markerStatus == StatusActive ||
		markerStatus == StatusCancelled ||
		markerStatus == StatusDestroyed {
		return true
	}
	return false
}

// Marshal needed for protobuf compatibility.
func (rt MarkerStatus) Marshal() ([]byte, error) {
	return []byte{byte(rt)}, nil
}

// Unmarshal needed for protobuf compatibility.
func (rt *MarkerStatus) Unmarshal(data []byte) error {
	*rt = MarkerStatus(data[0])
	return nil
}

// MarshalJSON using string.
func (rt MarkerStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(rt.String())
}

// UnmarshalJSON decodes from JSON string version of this status
func (rt *MarkerStatus) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	bz2, err := MarkerStatusFromString(s)
	if err != nil {
		return err
	}

	*rt = bz2
	return nil
}

// String implements the Stringer interface.
func (rt MarkerStatus) String() string {
	switch rt {
	case StatusUndefined:
		return "undefined"
	case StatusProposed:
		return "proposed"
	case StatusFinalized:
		return "finalized"
	case StatusActive:
		return "active"
	case StatusCancelled:
		return "cancelled"
	case StatusDestroyed:
		return "destroyed"

	default:
		return ""
	}
}

// Format implements the fmt.Formatter interface.
func (rt MarkerStatus) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(rt.String()))
	default:
		s.Write([]byte(fmt.Sprintf("%v", byte(rt))))
	}
}

// MarkerAddress returns the module account address for the given denomination
func MarkerAddress(denom string) (sdk.AccAddress, error) {
	if err := sdk.ValidateDenom(denom); err != nil {
		return nil, err
	}
	return sdk.AccAddress(crypto.AddressHash([]byte(denom))), nil
}

func MustGetMarkerAddress(denom string) sdk.AccAddress {
	addr, err := MarkerAddress(denom)
	if err != nil {
		panic(err)
	}
	return addr
}

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterInterface((*AccessGrantI)(nil), nil)
	cdc.RegisterInterface((*MarkerAccountI)(nil), nil)
	cdc.RegisterConcrete(&MarkerAccount{}, "provenance/marker/Account", nil)
	cdc.RegisterConcrete(&AccessGrant{}, "provenance/marker/AcccessGrant", nil)
}
