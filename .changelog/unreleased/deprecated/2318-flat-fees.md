* The `CalculateTxFees` query in the `x/msgfees` module is deprecated [PR 2318](https://github.com/provenance-io/provenance/pull/2318).
  Users should switch to query in the `x/flatfees` module with the same name.
* The `minimum-gas-prices` config field (in app.toml) is now ignored and is always treated as `1nhash` [PR 2318](https://github.com/provenance-io/provenance/pull/2318).
