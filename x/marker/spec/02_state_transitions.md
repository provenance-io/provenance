# State Transitions

This document describes the state transition operations pertaining markers:

<!-- TOC 2 2 -->
  - [Undefined](#undefined)
  - [Proposed](#proposed)
  - [Finalized](#finalized)
  - [Active](#active)
  - [Cancelled](#cancelled)
  - [Destroyed](#destroyed)



## Undefined

The undefined status is not allowed and its use will be flagged as an error condition.

## Proposed

The proposed status is the initial state of a marker.  A marker in the `proposed` status will accept
changes to supply via the `mint`/`burn` methods, updates to the access list, and state transitions when
called by the address set in the `manager` property.

On Transition:
- Proposed is the initial state of a marker by default.  It is not possible to transition to this state from any other.

Next Status:
- **Finalized**
- **Cancelled**

## Finalized

The finalized state of the marker is used to verify the readiness of a marker before activating it.

Requirements:
- Marker must exist
- Caller address must match the `manager` address on the marker
- Current status of marker must be `Proposed`
- Supply of the marker must meet or exceed the amount of any existing coin in circulation on the network of
  the denom of the marker. (This will only apply )

On Transition:
- Marker status is set to `Finalized`
- A marker finalize typed event is dispatched

Next Status:
- **Active**
- **Cancelled**

## Active

An active marker is considered ready for use.

On Transition:
- Marker status is set to `Active`
- Requested coin supply is minted and placed in the marker account
- For markers with a `fixed_supply` the Invariant checks are performed in `begin_block`
- Permissions as assigned in the access list are enforced for any management actions performed
- The `manager` field is cleared.  All management actions require explicit permission grants.
- A marker activate typed event is dispatched

Next Status:
- **Cancelled**

## Cancelled

A cancelled marker will have no coin supply in circulation.  Markers may remain in the Cancelled state long term to
prevent their denom reuse by another future marker. If a marker is no longer needed at all then the **Destroyed** 
status maybe appropriate.

Requirements:
- Caller must have the `delete` permission assigned to their address or
- Caller must be the manager of the marker (applies only to proposed markers that are Cancelled)
- The supply of the coin in circulation outside of the marker account must be zero.

On Transition:
- Marker status is set to `Cancelled`
- A marker Cancelled typed event is dispatched

Next Status:
- **Destroyed**

## Destroyed

A destroyed marker is denoted as available for subsequent removal from the state store by clean up processes.  Markers
in the destroyed status will be removed in the Begin Block ABCI handler at the beginning of the next block (v1.3.0+).

On Transition:
- All supply of the coin denom will be burned.
- Marker status is set to `Destroyed`
- Marker will ultimately be deleted from the KVStore during the next ABCI Begin Block (v1.3.0+)

Next Status:
- **None**
