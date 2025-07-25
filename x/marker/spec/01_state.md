# State

<!-- TOC 2 3 -->
  - [Marker Accounts](#marker-accounts)
    - [Marker Types](#marker-types)
    - [Access Grants](#access-grants)
    - [Fixed Supply vs Floating](#fixed-supply-vs-floating)
    - [Forced Transfers](#forced-transfers)
    - [Required Attributes](#required-attributes)
  - [Marker Address Cache](#marker-address-cache)
    - [Marker Net Asset Value](#marker-net-asset-value)
  - [Params](#params)



## Marker Accounts

Markers are represented as a type that extends the `base_account` type of the `auth` SDK module.  As a valid account a
marker is able to perform normal functions such as receiving and holding coins, and having a defined address that can
be queried against for balance information from the `bank` module.
<!-- link message: MarkerAccount -->

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/marker/v1/marker.proto#L28-L59

```go
type MarkerAccount struct {
	// cosmos base_account  including address and account number
    Address       string
    AccountNumber uint64

    PubKey        *types.Any // NOTE: not used for marker, it is not possible to sign for a marker account directly
    Sequence      uint64     // NOTE: always zero on marker

    // Address that owns the marker configuration.  This account must sign any requests
	// to change marker config (only valid for statuses prior to finalization)
	Manager string

	// Access control lists.  Account addresses are assigned control of the marker using these entries
	AccessControl []AccessGrant

	// Indicates the current status of this marker record. (Pending, Active, Cancelled, etc)
	Status MarkerStatus

	// value denomination.
	Denom string

	// the total supply expected for a marker.  This is the amount that is minted when a marker is created.  For
	// SupplyFixed markers this value will be enforced through an invariant that mints/burns from this account to
	// maintain a match between this value and the supply on the chain (maintained by bank module).  For all non-fixed
	// supply markers this value will be set to zero when the marker is activated.
	Supply Int

	// Marker type information.  The type of marker controls behavior of its account.
	MarkerType MarkerType

	// A fixed supply will mint additional coin automatically if the total supply decreases below a set value.  This
	// may occur if the coin is burned or an account holding the coin is slashed. (default: true)
	SupplyFixed bool

	// indicates that governance based control is allowed for this marker
	AllowGovernanceControl bool

	// Whether an admin can transfer restricted coins from a 3rd-party account without their signature.
	AllowForcedTransfer bool

	// list of required attributes on restricted marker in order to send and receive transfers if sender does not have
	// transfer authority
	RequiredAttributes []string
}
```

### Marker Types

There are currently two basic types of markers.

- **Coin** - A marker with a type of coin represents a standard fungible token with zero or more coins in circulation
- **Restricted Coin** - Restricted Coins work just like a regular coin with one important difference--the bank module
  "send_enabled" status for the coin is set to false.  This means that a user account that holds the coin can not send
  it to another account directly using the bank module.  In order to facilitate exchange there must be an address set
  on the marker with the "Transfer" permission grant.  This address must sign calls to the marker module to move these
  coins between accounts using the `transfer` method on the api.

### Access Grants

Control of a marker account is configured through a list of access grants assigned to the marker when it is created
or applied afterwards through the API calls to add or remove access.

```go
const (
	// ACCESS_UNSPECIFIED defines a no-op vote option.
	Access_Unknown Access = 0
	// ACCESS_MINT is the ability to increase the supply of a marker.
	Access_Mint Access = 1
	// ACCESS_BURN is the ability to decrease the supply of the marker using coin held by the marker.
	Access_Burn Access = 2
	// ACCESS_DEPOSIT is the ability to transfer funds from another account to this marker account
	// or to set a reference to this marker in the metadata/scopes module.
	Access_Deposit Access = 3
	// ACCESS_WITHDRAW is the ability to transfer funds from this marker account to another account
	// or to remove a reference to this marker in the metadata/scopes module.
	Access_Withdraw Access = 4
	// ACCESS_DELETE is the ability to move a proposed, finalized or active marker into the cancelled state.
	// This access also allows cancelled markers to be marked for deletion.
	Access_Delete Access = 5
	// ACCESS_ADMIN is the ability to add access grants for accounts to the list of marker permissions.
	// This access also gives the ability to update the marker's denom metadata.
	Access_Admin Access = 6
	// ACCESS_TRANSFER is the ability to manage transfer settings and broker transfers of the marker.
	// Accounts with this access can:
	//  - Update the marker's required attributes.
	//  - Update the send-deny list.
	//  - Use the transfer or bank send endpoints to move marker funds out of their own account.
	// This access right is only supported on RESTRICTED markers.
	Access_Transfer Access = 7
	// ACCESS_FORCE_TRANSFER is the ability to transfer restricted coins from a 3rd-party account without their signature.
	// This access right is only supported on RESTRICTED markers and only has meaning when allow_forced_transfer is true.
	Access_ForceTransfer Access = 8
)

// A structure associating a list of access permissions for a given account identified by is address
type AccessGrant struct {
	// A bech32 encoded address string of the account the permissions are assigned to
	Address     string
	 // An array of enum values as defined above
	Permissions AccessList
}
```

An admin with `Access_ForceTransfer` can use the `Transfer` endpoint to move marker funds (forced or not). However, an
admin with `Access_ForceTransfer`, but without `Access_Transfer`, cannot move marker funds by other means (e.g. a bank
`Send`). I.e. `Access_ForceTransfer` only has meaning with the `Transfer` endpoint.

### Fixed Supply vs Floating

A marker can be configured to have a fixed supply or one that is allowed to float.  A marker will always mint an amount
of coin indicated in its `supply` field when it is activated.  For markers that have a fixed supply an invariant check
is enforced that ensures the supply of the marker always matches the configured value.  For a floating supply no
additional checks or adjustments are performed and the supply value is set to zero when activated.

#### When a Marker has a Fixed Supply that does not match target

Under certain conditions a marker may begin a block with a total supply in circulation less than its configured amount.
When this occurs the marker will take action to correct the balance of coin supply.

**A fixed supply marker will attempt to automatically correct a supply imbalance at the start of the next block**

This means that if the supply in circulation exceeds the configured amount the attempted fix is to burn a required
amount from the marker's account itself.  If this fails an invariant will be broken and the chain will halt.

If the requested supply is greater than the amount in circulation (as occurs when a coin is burned in a slash) the
marker module will mint the difference between expected supply and circulation and place the created coin in the marker's
account.

A supply imbalance typically occurs during the genesis of a blockchain when a fixed supply for a marker is less than
the initial balances assigned to accounts.  It may also occur if the marker is associated with the bind denom of the
chain and a slash penalty is assessed resulting in the burning of a portion of coins.

### Forced Transfers

A marker with the **Restricted Coin** type can be configured to allow forced transfer of funds for that marker's denom.
A forced transfer is one where the `admin` (with `ACCESS_FORCE_TRANSFER`) is different than the `from` address. In such
cases, if the marker allows forced transfers, the transfer is allowed. If forced transfers are not allowed, an `admin`
cannot transfer the marker's coins from another account unless granted permission to do so via `authz`.
Forced transfers can only be made using the marker module's `Transfer` endpoint.

Markers with **Coin** type cannot be configured to allow forced transfers.

### Required Attributes

A marker with the **Restricted Coin** type can be configured to allow transfers with a normal `MsgSend` to address that have defined attributes.
This can be configured by setting the `required_attributes` array on the Marker.  When a `MsgSend` transaction is executed and the coin type is `restricted`, the `required_attributes` are checked. If the `ToAddress` associated with the `MsgSend` command has **all** the required attributes, the transfer will be executed.

A single wildcard can only be used for the starting name of the required attribute. For example, `*.provenance.io` is a valid wildcard attribute. Invalid wildcard usages include forms such as `*kyc.provenance.io` or `kyc.*.provenance.io`.  Matching will be accepted for any number of child level names, i.e. `one.two.three.provenance.io` and `one.provenance.io` will be accepted for `*.provenance.io`.

## Marker Address Cache

For performance purposes the marker module maintains a KVStore entry with the address of every marker account.  This
allows for cheap iterator operations over all marker accounts without having to filter through the native account
iterator from the auth module.

- `0x01 | Address -> Address`

### Marker Net Asset Value

A marker can support multiple distinct net asset values assigned to track settlement pricing information on-chain. The `price` attribute denotes the value assigned to the marker for a specific asset's associated `volume`. For instance, when considering a scenario where 10 billion `nhash` holds a value of 15¢, the corresponding `volume` should reflect the quantity of 10,000,000,000. The `update_block_height` attribute captures the block height when the update occurred.
<!-- link message: NetAssetValue -->

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/marker/v1/marker.proto#L91-L99

## Params

Params is a module-wide configuration structure that stores system parameters
and defines overall functioning of the marker module.

- Params: `Paramsspace("marker") -> legacy_amino(params)`

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/marker/v1/marker.proto#L14-L26
