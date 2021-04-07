# State

## Marker Accounts

Markers are represented as a type that extends the base_account type of the `auth` SDK module.  As a valid account a 
marker is able to perform normal functions such as receiving and holding coins, and having a defined address that can
be queried against for balance information from the `bank` module.

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/marker.proto#L28-L58
```go
type MarkerAccount struct {
	// cosmos base_account  including address and account number
    Address       string
    PubKey        *types.Any // NOTE: not used for marker
    AccountNumber uint64
    Sequence      uint64     // NOTE: always zero on marker

    // Address that owns the marker configuration.  This account must sign any requests
	// to change marker config (only valid for statuses prior to finalization)
	Manager string
	// Access control lists
	AccessControl []AccessGrant
	// Indicates the current status of this marker record.
	Status MarkerStatus
	// value denomination and total supply for the token.
	Denom string
	// the total supply expected for a marker.  This is the amount that is minted when a marker is created.
	Supply Int
	// Marker type information
	MarkerType MarkerType
	// A fixed supply will mint additional coin automatically if the total supply decreases below a set value.  This
	// may occur if the coin is burned or an account holding the coin is slashed. (default: true)
	SupplyFixed bool
	// indicates that governance based control is allowed for this marker
	AllowGovernanceControl bool
}
```

## Marker Address Cache

For performance purposes the marker module maintains a KVStore entry with the address of every marker account.  This
allows for cheap iterator operations over all marker accounts without having to filter through the native account
iterator from the auth module.

- `0x01 | Address -> Address`
## Params

Params is a module-wide configuration structure that stores system parameters
and defines overall functioning of the marker module.

- Params: `Paramsspace("marker") -> legacy_amino(params)`

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/marker.proto#L14-L25



