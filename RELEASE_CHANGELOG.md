## [v1.4.0](https://github.com/provenance-io/provenance/releases/tag/v1.4.0) - 2021-06-02

### Features

* ENV config support, SDK v0.42.5 update [#320](https://github.com/provenance-io/provenance/issues/320)
* Upgrade handler set version name to `citrine` [#339](https://github.com/provenance-io/provenance/issues/339)

### Bug Fixes

* P8EMemorializeContract: preserve some Scope fields if the scope already exists [PR 336](https://github.com/provenance-io/provenance/pull/336)
* Set default standard err/out for `provenanced` commands [PR 337](https://github.com/provenance-io/provenance/pull/337)
* Fix for invalid help text permissions list on marker access grant command [PR 337](https://github.com/provenance-io/provenance/pull/337)
* When writing a session, make sure the scope spec of the containing scope, contains the session's contract spec. [#322](https://github.com/provenance-io/provenance/issues/322)

###  Improvements

* Informative error message for `min-gas-prices` invalid config panic on startup [#333](https://github.com/provenance-io/provenance/issues/333)
* Update marker event documentation to match typed event namespaces [#304](https://github.com/provenance-io/provenance/issues/304)

