## [v1.0.0](https://github.com/provenance-io/provenance/releases/tag/v1.0.0) - 2021-03-31

### Bug Fixes

* Resolves an issue where Gov Proposals to Create a new name would fail for new root domains #192
* Remove deprecated ModuleCdc amino encoding from Metadata Locator records #187
* Update Cosmos SDK to 0.42.3
* Remove deprecated ModuleCdc amino encoding from name module #189
* Remove deprecated ModuleCdc amino encoding from attribute module #188

### Features

* Allow withdrawals of any coin type from a marker account in WASM smart contracts. #151
* Added cli tx commands `write-contract-specification` `remove-contract-specification` for updating/adding/removing metadata `ContractSpecification`s. #195
* Added cli tx commands `write-record-specification` `remove-record-specification` for updating/adding/removing metadata `RecordSpecification`s. #176
* Added cli tx commands `write-scope-specification` `remove-scope-specification` for updating/adding/removing metadata `ScopeSpecification`s. #202
* Added cli tx commands `write-scope` `remove-scope` for updating/adding/removing metadata `Scope`s. #199
* Added cli tx commands `write-record` `remove-record` for updating/adding/removing metadata `Record`s. #205
* Simulation testing support #95
* Name module simulation testing #24
* Added default IBC parameters for v039 chain genesis migration script #102
* Expand and simplify querying. #169
  * Added endpoints for getting all entries of a type, e.g. `RecordsAll`.
  * Combined some endpoints (see notesin "API Breaking" section).
  * Allow searching for related entries. E.g. you can provide a record id to the scope search.
  * Add ability to return related entries. E.g. the `Sessions` endpoint has a `include_records` flag that will cause the response to contain the records that are part of the sessions.
* Add optional identification fields in tx `Write...` messages. #169
* The `Write` endpoints now return information about the written entries. #169
* Added a CLI command for getting all entries of a type, `query metadata all <type>`, or `query metadata <type> all`. #169
* Restrict denom metadata. #208

### API Breaking

* Change `Add...` metadata tx endpoints to `Write...` (e.g. `AddScope` is now `WriteScope`). #169
* Expand and simplify metadata querying. #169
  * Removed the `SessionContextByID` and `SessionContextByUUID` endponts. Replaced with the `Sessions` endpoint.
  * Removed the `RecordsByScopeID` and `RecordsByScopeUUID` endpoints. Replaced with the `Records` endpoint.
  * Removed the `ContractSpecificationExtended` endpoint. Use `ContractSpecification` now with the `include_record_specs` flag.
  * Removed the `RecordSpecificationByID` endpoint. Use the `RecordSpecification` endpoint.
  * Change the `_uuid` fields in the queries to `_id` to allow for either address or uuid input.
  * The `Scope` query no longer returns `Sessions` and `Records` by default. Use the `include_sessions` and `include_records` if you want them.
  * Query result entries are now wrapped to include extra id information alongside an entry.
    E.g. Where a `Scope` used to be returned, now a `ScopeWrapper` is returned containing a `Scope` and its `ScopeIdInfo`.
    So where you previously had `resp.Scope` you will now want `resp.Scope.Scope`.
  * Pluralized both the message name and field name of locator queries that return multiple entries.
    * `OSLocatorByScopeUUIDRequest` and `OSLocatorByScopeUUIDResponse` changed to `OSLocatorsByScopeUUIDRequest` and `OSLocatorsByScopeUUIDResponse`.
    * `OSLocatorByURIRequest` and `OSLocatorByURIResponse` changed to `OSLocatorsByURIRequest` and `OSLocatorsByURIResponse`.
    * Field name `locator` changed to `locators` in `OSLocatorsByURIResponse`, `OSLocatorsByScopeUUIDResponse`, `OSAllLocatorsResponse`.

### Client Breaking

* The paths for querying metadata have changed. See API Breaking section for an overview, and the proto file for details. #169
* The CLI has been updated for metadata querying. #169
  * Removed the `fullscope` command. Use `query metadata scope --include-sessions --include-records` now.
  * Combined the `locator-by-addr`, `locator-by-uri`, `locator-by-scope`, and `locator-all` into a single `locator` command.
* Changed the CLI metadata tx `add-...` commands to `write-...`. #166

