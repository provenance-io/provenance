## [v1.9.0](https://github.com/provenance-io/provenance/releases/tag/v1.9.0) - 2022-04-25

### Summary

Provenance 1.9.0 brings some minor features and security improvements.

### Features

* Add `add-genesis-msg-fee` command to add msg fees to genesis.json and update Makefile to have pre-defined msg fees [#667](https://github.com/provenance-io/provenance/issues/667)
* Add msgfees summary event to be emitted when there are txs that have fees [#678](https://github.com/provenance-io/provenance/issues/678)
* Adds home subcommand to the cli's config command [#620] (https://github.com/provenance-io/provenance/issues/620)
* Add support for rocksdb and badgerdb [#702](https://github.com/provenance-io/provenance/issues/702)
* Create `dbmigrate` utility for migrating a data folder to use a different db backend [#696](https://github.com/provenance-io/provenance/issues/696)

### Improvements

* When the `start` command encounters an error, it no longer outputs command usage [#670](https://github.com/provenance-io/provenance/issues/670)
* Change max length on marker unresticted denom from 64 to 83 [#719](https://github.com/provenance-io/provenance/issues/719)
* Set prerelease to `true` for release candidates. [#666](https://github.com/provenance-io/provenance/issues/666)
* Allow authz grants to work on scope value owners [#755](https://github.com/provenance-io/provenance/issues/755)
* Bump wasmd to v0.26 (from v0.24). [#799](https://github.com/provenance-io/provenance/pull/799)

