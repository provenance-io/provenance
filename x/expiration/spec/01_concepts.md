# Concepts

The expiration module manages asset expirations.

<!-- TOC -->
  - [Assets](#assets)
  - [Expiration Metadata](#expiration-metadata)
    - [Module Asset ID](#module-asset-id)
    - [Owner](#owner)
    - [Expiration Time](#expiration-time)
    - [Deposit](#deposit)
    - [Message](#message)
  - [Creating Expirations on Assets](#creating-expirations-on-assets)
  - [Extending Expirations on Assets](#extending-expirations-on-assets)
  - [Owner Enforced Expiration](#owner-enforced-expiration)
  - [Externally Enforced Expiration](#externally-enforced-expiration)
  - [Authz](#authz)


---
## Assets

When storing assets on-chain, there is an expense in terms of node memory and processing 
that is associated with each added asset. Without some method of expiration, the active state 
of the system would grow until it eventually becomes too large to be efficiently managed and 
system performance would degrade.

## Expiration Metadata

Expiration metadata is added to the system as part of the asset creation process and 
is customized on a per-asset-type basis.

### Module Asset ID

the module asset ID is a bech32 address string or a [MetadataAddress](../../metadata/spec/01_concepts.md) 
that assigned to each asset on chain.

### Owner

The owner is a bech32 address string that determines the rightful owner to a particular expiring asset on chain.

### Expiration Time

The expiration time determines when an asset on chain will expire and is up for removal.

### Deposit

When an expiration record is added to the system, funds, in the form of a deposit, will be moved from the owner address
associated with the module asset and stored in the account associated with the expiration module. The deposit can be returned
when the owner later executes the expiration logic which frees up the funds held for the underlying asset.Note that if the
expiration logic is not triggered and the expiration time passes, external actors may take action to execute the expiration
logic and collect the deposit (which offsets some gas fees required for processing the expiration logic).

### Message

When expiration logic is executed, the module asset message stored in the expiration metadata is dispatched to the associated module
through the [Msg Service Router](https://docs.cosmos.network/main/core/baseapp.html#msg-service-router).

## Creating Expirations on Assets

Creating an expiration is done through the module that owns the asset. When a new asset is created for a particular module,
the user creating the asset has the option to specify if the asset in question will expire at some time in the future.
Note that certain module assets can never expire. Therefore, asset expiration is an opt-in feature that asset owners can
leverage for their on-chain assets.

## Extending Expirations on Assets

Extending an expiration can be done through the CLI or through auto-updates from the modules that own them. 
Note that auto-updates are left up to the module that owns the asset.

from the cmd line:
```shell
provenanced tx expiration extend <required params>
```

## Owner Enforced Expiration

A common use case is for an owner to execute expiration logic for an asset they registered for expiration.
When the expiration logic is executed, the message in the expiration metadata will be processed, after which 
the deposit will be refunded to the owner's account.

from the cmd line:
```shell
provenanced tx expiration invoke <required params>
```

## Externally Enforced Expiration

If expiration logic is not executed by an owner within the expiration period, the expiration logic may be executed
by an external actor in order to redeem the deposit. The deposit helps to offset gas costs required to execute the
logic and provides an incentive for dangling assets to be pruned quickly.

from the cmd line:
```shell
provenanced tx expiration invoke <required params>
```

## Authz

Expirations can be managed by third parties through authorization grants. Owners of an asset expiration will need
to grant permissions to third parties through the [authz](https://docs.cosmos.network/main/modules/authz) module.
