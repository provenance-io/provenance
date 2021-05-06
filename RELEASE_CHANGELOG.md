## [v1.3.0](https://github.com/provenance-io/provenance/releases/tag/v1.3.0) - 2021-05-06

### Features

* Add grpc messages and cli command to add/remove addresses from metadata scope data access [#220](https://github.com/provenance-io/provenance/issues/220)
* Add a `context` field to the `Session` [#276](https://github.com/provenance-io/provenance/issues/276)
* Add typed events and telemetry metrics to attribute module [#86](https://github.com/provenance-io/provenance/issues/86)
* Add rpc and cli support for adding/updating/removing owners on a `Scope` [#283](https://github.com/provenance-io/provenance/issues/283)
* Add transaction and query time measurements to marker module [#284](https://github.com/provenance-io/provenance/issues/284)
* Upgrade handler included that sets denom metadata for `hash` bond denom [#294](https://github.com/provenance-io/provenance/issues/294)
* Upgrade wasmd to v0.16.0 [#291](https://github.com/provenance-io/provenance/issues/291)
* Add params query endpoint to the marker module cli [#271](https://github.com/provenance-io/provenance/issues/271)

### Improvements

* Added linkify script for changelog issue links [#107](https://github.com/provenance-io/provenance/issues/107)
* Changed Metadata events to be typed events [#88](https://github.com/provenance-io/provenance/issues/88)
* Updated marker module spec documentation [#93](https://github.com/provenance-io/provenance/issues/93)
* Gas consumption telemetry and tracing [#299](https://github.com/provenance-io/provenance/issues/299)

### Bug Fixes

* More mapping fixes related to `WriteP8EContractSpec` and `P8EMemorializeContract` [#275](https://github.com/provenance-io/provenance/issues/275)
* Fix event manager scope in attribute, name, marker, and metadata modules to prevent event duplication [#289](https://github.com/provenance-io/provenance/issues/289)
* Proposed markers that are cancelled can be deleted without ADMIN role being assigned [#280](https://github.com/provenance-io/provenance/issues/280)
* Fix to ensure markers have no balances in Escrow prior to being deleted. [#303](https://github.com/provenance-io/provenance/issues/303)

### State Machine Breaking

* Add support for purging destroyed markers [#282](https://github.com/provenance-io/provenance/issues/282)
