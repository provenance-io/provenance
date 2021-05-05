# Begin-Block


## Supply Checks

Each ABCI begin block call, all markers that are active and have a fixed supply
are evaluated to ensure configured supply level matches actual supply levels.

- For markers that have a configured supply exceeding the amount in circulation the difference is minted and placed
  within the marker account.
- For markers that have a configured supply less than the amount in circulation, an attempt to burn sufficient coin
  to balance the circulation against the the supply will be performed.  If the marker does not hold enough coin to
  perform this action an invariant constraint violation is thrown and the chain will halt.

## Destroyed Markers
In addition to supply checks the ABCI begin block call is used to purge markers that have been selected for deletion.

- Markers in the `destroyed` status are deleted from the KVStore.
