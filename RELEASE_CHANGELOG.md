## [v1.17.1](https://github.com/provenance-io/provenance/releases/tag/v1.17.1) - 2024-01-11

Users should upgrade to v1.17.1 at their earliest convenience.

Release v1.17.1 addresses [CWA-2023-004](https://github.com/CosmWasm/advisories/blob/main/CWAs/CWA-2023-004.md) and also adds some command-line functionality.

### New Configuration Recommendations

Provenance has a few new recommendations reguarding node configuration.

#### Goleveldb

Provenance now recommends that nodes use `goleveldb` as their db backend. Support for `cleveldb` and `badgerdb` will be removed in a future upgrade; it would be better to do this migration outside of an upgrade. Nodes currently using those database backends should migrate to `goleveldb` at your leisure prior to that upgrade. If your node is using `cleveldb`, a warning will be issued when your node starts.

To migrate to `goleveldb` from `cleveldb`:

1. Stop your node.
2. Back-up your `data` and `config` directories.
3. Update your `config.toml` to have `db_backend = "goleveldb"`.
4. Update your `app.toml` to have either `app-db-backedn = ""` or `app-db-backend = "goleveldb"`.
5. Restart your node.

In some cases, that process might not work and your node will fail to restart. If that happens, or if you are using `badgerdb`:

1. Stop your node.
2. If the above process failed, restore your `data` and `config` directories from your backups.
3. Use the [dbmigrate](https://github.com/provenance-io/provenance/releases/download/v1.17.0/dbmigrate-linux-amd64-v1.17.0.zip) utility to migrate your node's database to `goleveldb`. This can take 3 hours or more to complete and should not be interrupted.
4. Restart your node.

#### IAVL-Fastsync

Provenance also recommends that nodes enable iavl-fastsync. In a future upgrade, nodes will be required to use iavl-fastnode, and it would be better to do this migration outside of an upgrade. If your node's `app.toml` has `iavl-disable-fastnode = true`, you should migrate your store at your leisure prior to the next upgrade. If your node is has iavl-fastsync disabled, a warning will be issued when your node starts.

It might take 3 hours or more for the migration to finish. Do not stop or restart your node during this process. Your node will be unavailable during this process.

To migrate to iavl-fastsync, follow these steps:

1. Stop your node.
2. Back up your data directory.
3. Update your node's `app.toml` to have `iavl-disable-fastnode = false`.
4. Restart your node.

#### Pruning Interval

Provenance recommends, too, that validators use a pruning interval of at most `10`. This can help prevent missed blocks. This is configured in `app.toml`. If you have a `pruning` value of `"default"`, `"nothing"`, or `"everything"`, you are okay. If you have a `pruning` value of `"custom"` and a `pruning-interval` of `1000` or more, a warning will be issued when your node starts.

#### Indexer

Lastly, Provenance recommends that validators do not enable tx indexing. This should also help prevent missed blocks. This is configured in `config.toml` in the `indexer` field of the `tx_index` section (aka `tx_index.indexer`). If you have an indexer defined, a warning will be issued when your node starts.

---

### Features

* Add CLI commands for the exchange module endpoints and queries [#1701](https://github.com/provenance-io/provenance/issues/1701).
* Create CLI commands for adding a market to a genesis file [#1757](https://github.com/provenance-io/provenance/issues/1757).
* Add CLI command to generate autocomplete shell scripts [#1762](https://github.com/provenance-io/provenance/pull/1762).

### Improvements

* Add StoreLoader wrapper to check configuration settings [#1792](https://github.com/provenance-io/provenance/pull/1792).
* Create a default market in `make run`, `localnet`, `devnet` and the `provenanced testnet` command [#1757](https://github.com/provenance-io/provenance/issues/1757).
* Updated documentation for each module to work with docusaurus [PR 1763](https://github.com/provenance-io/provenance/pull/1763)

### Bug Fixes

* Deprecate marker proposal transaction [#1797](https://github.com/provenance-io/provenance/issues/1797).

### Dependencies

- Bump `github.com/spf13/cobra` from 1.7.0 to 1.8.0 ([#1733](https://github.com/provenance-io/provenance/pull/1733))
- Bump `github.com/CosmWasm/wasmvm` from 1.2.4 to 1.2.6 ([#1799](https://github.com/provenance-io/provenance/issues/1799))
- Bump `github.com/CosmWasm/wasmd` from v0.30.0-pio-5 to v0.30.0-pio-6 ([#1799](https://github.com/provenance-io/provenance/issues/1799))

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.17.0...v1.17.1

