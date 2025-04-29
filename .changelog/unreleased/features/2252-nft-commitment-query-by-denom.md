* Enhanced `GetAccountCommitments` to support querying NFT coin commitments with optional asset filter in the exchange module [#2252](https://github.com/provenance-io/provenance/issues/2252).
* The `GetAccountCommitments` query now lets you filter commitments by a specific asset, thanks to an optional asset parameter.
* The CLI command `SetupCmdQueryGetAccountCommitments` was updated to include a --asset flag, making it easier to filter commitments by asset directly from the command * line.
* New test cases were added to ensure the `GetAccountCommitments` query works correctly when filtering by asset.
* Documentation was updated to explain the changes in both the query and the command.
