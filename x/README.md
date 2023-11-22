---
sidebar_position: 0
---

# List of Modules

Modules are the code components of the Provenance Blockchain that execute the majority of the business logic for applications. The [Cosmos SDK](https://docs.cosmos.network/v0.47) enables developers to build modules that utilize the core structure of the SDK to allow the modules to function together. To read more about creating modules, refer to the [Cosmos documentation on modules](https://docs.cosmos.network/v0.47/building-modules/intro).

Provenance uses inherited modules from the Cosmos SDK, and has also developed modules that are specific to Provenance.

* [Inherited Cosmos modules](https://docs.cosmos.network/v0.47/build/modules)
* [Attribute](./attribute/README.md) - Functions as a blockchain registry for storing \<Name, Value\> pairs.
* [Exchange](./exchange/README.md) - Facilitates the trading of on-chain assets.
* [Hold](./hold/README.md) - Keeps track of funds in an account that have a hold placed on them.
* [ibchooks](./ibchooks/README.md) - Forked from https://github.com/osmosis-labs/osmosis/tree/main/x/ibchooks
* [Marker](./marker/README.md) - Allows for the creation of fungible tokens.
* [Metadata](./metadata/README.md) - Provides a system for referencing off-chain information.
* [msgfees](./msgfees/README.md) - Manages additional fees that can be applied to tx msgs.
* [Name](./name/README.md) - Provides a system for providing human-readable names as aliases for addresses.
* [Oracle](./oracle/README.md) - Provides the capability to dynamically expose query endpoints.
* [Reward](./reward/README.md) - Provides a system for distributing rewards to accounts.
* [Trigger](./trigger/README.md) - Provides a system for triggering transactions based on predeterminded events.
