## [v1.1.0](https://github.com/provenance-io/provenance/releases/tag/v1.1.0) - 2021-04-15

## Features
* Add marker cli has two new flags to set SupplyFixed and AllowGovernanceControl #241
* Modify 'enable governance' behavior on marker module #227
* Typed Events and Metric counters in Name Module #85

### Improvements
* Add some extra aliases for the CLI query metadata commands.
* Make p8e contract spec id easier to communicate.

### Bug Fixes
* Add pagination flags to the CLI query metadata commands.
* Fix handling of Metadata Write message id helper fields.
* Fix cli metadata address encoding/decoding command tree #231
* Metadata Module parsing of base64 public key fixed #225
* Fix some conversion pieces in `P8EMemorializeContract`.
* Remove extra Object Store Locator storage.
* Fix input status mapping.
* Add MsgSetDenomMetadataRequest to the marker handler.