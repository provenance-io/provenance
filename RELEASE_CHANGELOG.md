## [v1.12.2](https://github.com/provenance-io/provenance/releases/tag/v1.12.1) - 2022-10-25

Provenance v1.12.2 enables the ability to upgrade your IAVL state store to be faster and handle errors better. This upgrade is recommended and should be done at your convenience prior to the v1.13 chain upgrade.

The IAVL store upgrade is expected to take 30 to 90 minutes. During that time, your node will be down. There will be some log output (info level), but it is sparce and there may be long periods (25+ minutes) without any new log output. Once it has started, it's best to not interrupt the process.

It is highly recommended that you do one of these two prior to the v1.13 chain upgrade:

Either

- Upgrade your node's IAVL store:
  1. Stop your node.
  2. Upgrade `provenanced` to v1.12.2.
  3. Run the command: `provenanced config set iavl-disable-fastnode false`.
  4. Restart your node. Once the upgrade has finished, your node will automatically run as normal.

Or

- Explicitly define that you don't want to upgrade your node's IAVL store:
   1. Ensure that you have `provenanced` v1.12.1 (or higher), e.g. Run the command: `provenanced version`. If you are on 1.12.0, upgrade to at least v1.12.1.
   2. Run the command: `provenanced config set iavl-disable-fastnode true`.

---

You can manually update your `app.toml` file, but using the `config set` command is the recommended method. The `iavl-disable-fastnode` field was added in v1.12.1 and most likely does not yet exist in your `app.toml` file. There are other new sections and fields too. Using the command will add them all (using defaults) as well as their descriptions. If you want to update your `app.toml` manually, the `iavl-disable-fastnode` entry should go below the `index-events` entry and before the `[telemetry]` section.

If you do nothing before the v1.13 chain upgrade, your node will most likely upgrade the IAVL store when v1.13 first runs. The v1.13 chain upgrade and migrations are expected to only take a minute. If your node is also upgrading the IAVL store at that time, it will take 30-90+ minutes.

Note: The command `provenanced config get iavl-disable-fastnode` will report a value regardless of whether the field exists in `app.toml`. As such, that command is insufficient for determining whether the value exists in the `app.toml` file.

### Improvements

* Bump Cosmos-SDK to v0.45.10-pio-2 (from v0.45.9-pio-1) [PR 1193](https://github.com/provenance-io/provenance/pull/1193).
* Allow the IAVL store to be upgraded [PR 1193](https://github.com/provenance-io/provenance/pull/1193).

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.12.1...v1.12.2

