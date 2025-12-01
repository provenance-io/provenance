Provenance Blockchain version `v1.26.0` contains some exciting new features.

**Important**: All users should now use `1nhash` for `gas-prices` and a multiplier of `1.0` (the default).

Fees on Provenance Blockchain are now based on msg type instead of gas.
The standard Tx simulation process now returns the fee amount as gas wanted (i.e. it no longer reflects an actual gas amount).
By using `1nhash` for gas prices when simulating, existing client(s) will properly set the fee for the Tx.

The new `x/flatfees` module manages the costs of each Msg type.
Costs are defined in milli-US-dollars (musd).
A conversion factor (defined in module params) is used to determine the amount of nhash equivalent to the musd cost of a Msg.
This conversion factor will be constant to start, but can be updated manually (via governance proposal).
In the future, the conversion factor might be updated automatically based on market conditions.
By keeping the conversion factor in-line with the cost of hash, fees will remain constant in terms of how much they cost in US dollars (even though the required amounts of hash might change).

---

This upgrade also adds several other modules:

* The `x/ledger` module manages financial tracking for NFTs and metadata scopes.
* The `x/registry` module assigns roles to addresses associated NFTs (including scopes).
* The `x/asset` module provides a comprehensive digital asset management system on the Provenance blockchain.
* The `x/nft` module is the Cosmos-SDK's standard way of recording and managing NFTs.
* The `x/vault` module provides a system for tokenized vaults built on Provenance’s marker and account model.

The Asset, Ledger, and Registry modules, together, provide mechanisms for tracking the worth of the scopes and nfts on Provenance.
