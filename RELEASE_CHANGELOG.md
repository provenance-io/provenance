## [v1.13.1](https://github.com/provenance-io/provenance/releases/tag/v1.13.1)

This is an in-place upgrade of [v1.13.0](https://github.com/provenance-io/provenance/releases/tag/v1.13.0). Upgrading to this version is recommended at your earliest convenience.

Provenance Blockchain has identified some issues with `cleveldb` as a backend. The issues involve improper closing of the database when the `provenanced` process is killed (e.g. by Cosmovisor). As such, please consider switching to `goleveldb` either with this upgrade or separately.

NOTE: `provenanced` configuration commands are sensitive to the `--home` flags and `PIO_HOME` environment variables of your node setup.

Recommended procedure for switching to `goleveldb` from `cleveldb`:
1. Make sure you're using `cleveldb`. This can be done using the `provenanced config get db_backend` command, or by checking your `config.toml` file. Only proceed with these instructions if you're using `cleveldb`.
1. Stop your node.
2. Make a backup of your `data` directory, e.g. `$PIO_HOME/data`.
3. Change the `db_backend` config value from `cleveldb` to `goleveldb`. This can be done using the `provenanced config set db_backend goleveldb` command, or by updating your `config.toml` directly.
4. Restart your node.

If your node fails to start, change the config back and restore your `data` directory from your backup.
Then please let us know (e.g. in Discord).
Please indicate your OS, and if possible your `cleveldb` library version.

### Improvements

* Updated Cosmos-SDK to `v0.46.6-pio-3` (from `v0.46.6-pio-1`) [PR 1274](https://github.com/provenance-io/provenance/pull/1274).

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.13.0...v1.13.1
