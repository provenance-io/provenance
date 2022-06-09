# ADR 200: On-chain Asset Expiration

## Changelog

* 2022-06-03: Initial version

## Status

Draft (WORK IN PROGRESS)

## Abstract

This ADR introduces an approach for expiring long-lived assets to prevent the cost of maintaining their state when no longer needed. The expiration process is designed to recover gas costs of the expiration operations and not require any per-block processing overhead in handling asset expiration. 

## Context

When storing assets on-chain, there is an expense in terms of node memory and processing that is associated with each added asset. Without some method of expiration, the active state of the system would grow until it eventually becomes too large to be effiently managed and system performance would degrade. 

In many cases, asset lifespan may be estimated at creation so that the asset may be expired and potentially removed from active state. The expiration process uses node processing resources to prune the state, so the solution should offset gas costs if possible. Also, since a large number of assets may be monitored for expiration over a long lifespan, care must be taken to prevent expiration processing from adding too much overhead to standard processing loop for nodes.

## Decision

Based on the above requirements we will introduce a new system module that is reponsible for handling asset expiration. The module will support adding expiration metadata for assets from other system modules and will be flexible in the way it approaches expiration. The assumption is that each asset type will likely require its own customizable series of operations in order to be removed from active state.

### Adding Expiration Metadata
Expiration metadata is added to the system as part of the asset creation process and is customized on a per-asset-type basis. An example of the message used to add expiration metadata for an asset might look like the following:

```protobuf
message MsgAddExpirationMetadataRequest {
  string module_asset_identifier = 1;
  string owner_address = 2;
  int64  expiration_height = 3;
  cosmos.base.v1beta1.Coin deposit = 4;
  repeated google.protobuf.Any expiration_messages = 5;
}
```

After the expiration metadata is added to the system, funds will be moved from the owner address to a module account corresponding to the module the asset is associated with. This deposit can be returned if the owner later executes the expiration logic which frees up the resources for the underlying asset. Note that if the expiration logic is not triggered and the expiration time passes, external actors may take action to execute the expiration logic and collect the deposit (which offsets some of the gas fees required for processing the expiration logic).

#### Questions
* How is `module_asset_identifier` built so that we can effectively perform queries (e.g. do we support query of expiring assets by module)?
* Is there a single `owner_address` or a list of addresses?
* Is `expiration_height` simply a block height or should this tie into epoch module?
* Which approach is used to collect the deposit? Is this accomplished using the message fees module or with a more direct approach?

#### Observations
* The `expiration_messages` field is based on the one for the Cosmos governance module. We may be able to leverage common behavior with that module when executing the list of messages to achieve expiration.

### Extending Expiration
Asset expiration may be extended at any time by the owner(s) by issuing an extension request. An example of the message used to extend expiration for an asset might look like the following:

```protobuf
message MsgExtendExpirationRequest {
  string module_asset_identifier = 1;
  int64  new_expiration_height = 2;
  cosmos.base.v1beta1.Coin additional_deposit = 3;
}
```

#### Questions
* Are there any limitations on `new_expiration_height` (e.g. max distance in the future)?
* Should we include `additional_deposit` to allow for requiring more funds for extending the expiration? Would this be tied in some way to the length of the expiration period?

#### Observations
* The message signer for the expiration extension should be the owner (or in the list of owners if multiple are supported).

### Extend Expiration on Asset Update
The current approach assumes that assets which have expiration metadata assigned would have the expiration period automatically pushed further into the future when the underlying assets are updated. This will require additional code to be added to the update methods of the assets since there is not a generic way to execute this behavior out-of-band.

#### Questions
* Does the automatic update functionality fire a message to extend the expiration in a manner similar to an external request? This seems like the right approach so that there is an audit trail.

### Query Expiration Metadata
There will be various queries which allow expiration metadata to be listed based on criteria. Possible criteria include:
* Expiring assets by owner (assets about to expire within x blocks)
* Expiring assets by module/type (assets about to expire within x blocks)
* Expiring assets across all modules/owners? (assets about to expire within x blocks)
* Expired assets by owner?
* Expired assets by module/type
* All expired assets?

#### Questions
* Which queries above do we plan to support? Are there others?

### Owner Enforced Expiration
A common use case will be for an owner to execute expiration logic for an asset they registered for expiration. When the expiration logic is executed, the list of messages in the expiration metadata will be processed in order, after which the deposit will be refunded to the owner account.

#### Questions
* What if one or more queries in the expiration logic fail?

### Externally Enforced Expiration
If expiration logic is not executed by an owner within the expiration period, the expiration logic may be executed by an external actor in order to redeem the deposit. The deposit helps to offset gas costs required to execute the logic and provides an incentive for dangling assets to be pruned quickly.

#### Questions
* Are there any limitations on which external actors can execute the expiration logic?

## Consequences

### Backwards Compatibility

This is new functionality so there should not be issues with backward compatibility. Existing on-chain assets will not have expiration data associated and will therefore be immune to expiration. If the preference is to add expiration metadata for existing assets, that data could be loaded in bulk via an upgrade handler or similar mechanism.

### Positive

Expiration of on-chain assets will positively affect the system by removing state data that is no longer required. This will improve system perforance and allow for more assets to be onboarded without as much stress on the system.

### Negative

Expiration adds extra metadata to the system for tracking when assets expire and also adds processing to each asset creation since expiration metadata is created at the same time. These consequenses should be offset by the savings realized by removing the corresponding assets from active state.

## Further Discussions

More discussions go here.

## Test Cases [optional]

Test cases for an implementation are mandatory for ADRs that are affecting consensus changes. Other ADRs can choose to include links to test cases if applicable.

## References

* {reference link}