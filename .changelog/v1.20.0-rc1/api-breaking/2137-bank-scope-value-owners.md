* The `Ownership` query in the `x/metadata` module now only returns scopes that have the provided address in the `owners` list [#2137](https://github.com/provenance-io/provenance/issues/2137).
  Previously, if an address was the value owner of a scope, but not in the `owners` list, the scope would be returned
  by the `Ownership` query when given that address.  That is no longer the case.
  The `ValueOwnership` query can be to identify scopes with a specific value owner (like before).
  If a scope has a value owner that is also in its `owners` list, it will still be returned by both queries.
* The `WriteScope` endpoint now uses the `scope.value_owner_address` differently [#2137](https://github.com/provenance-io/provenance/issues/2137).
  If it is empty, it indicates that there is no change to the value owner of the scope and the releated lookups and validation
  are skipped. If it isn't empty, the current value owner will be looked up and the coin for the scope will be transferred to
  the provided address (assuming signer validation passed).
* An authz grant on `MsgWriteScope` no longer also applies to the `UpdateValueOwners` or `MigrateValueOwner` endpoints [#2137](https://github.com/provenance-io/provenance/issues/2137).
