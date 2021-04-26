## [v1.2.0](https://github.com/provenance-io/provenance/releases/tag/v1.2.0) - 2021-04-26

### Improvements

* Add spec documentation for the metadata module #224

### Features

* Add typed events and telemetry metrics to marker module #247

### Bug Fixes

* Wired recovery flag into `init` command #254
* Always anchor unrestricted denom validation expressions, Do not allow slashes in marker denom expressions #258
* Mapping and validation fixes found while trying to use `P8EMemorializeContract` #256

### Client Breaking

* Update marker transfer request signing behavior #246
