# `Marker`

## Abstract

This document specifies the marker module of the Provenance blockchain.

The marker module provides the capability for creation and management of
fungible tokens on the Provenance blockchain.  Various types of tokens can
be represented including standard coins and restricted coins (securities).

Further the marker module allows for coins to be fixed upon creation or
managed by an identified list of accounts, or through the governance
proposal process.
## Context

Using the blockchain as a ledger requires the ability to track fungible and non-fungible resources on chain with
fractional ownership.  Each of these resources requires rules governing supply and exchange.  Examples of resources
include fractional ownership in the network itself (stake), credits for network resources (gas/fees), fractional
ownership of an arbitrary asset (metadata/scope), and omnibus account balances (stable coins).  The rules governing the
asset must be enforced by the blockchain itself such that the entity controlling the asset must abide by these
rules and is not able to invalidate these processes.  These enforced constraints are what provide the value and
support trust in the platform itself.

## Overview

The marker module provides various tools for defining fractional ownership and control.  Markers can be created and
managed by normal Msg requests or through the governance process.  A marker can have many users with explicit control
or none at all.  A marker can be used to create a coin that can be freely exchange or one that requires facilitated
transfer by the marker itself when invoked by a user/process with appropriate permissions.

## Contents

1. **[State](01_state.md)**
1. **[State_transitions](02_state_transitions.md)**
1. **[Messages](03_messages.md)**
1. **[Begin Block](04_begin_block.md)**
1. **[End Block](05_end_block.md)**
1. **[Hooks](06_hooks.md)**
1. **[Events](07_events.md)**
1. **[Telemetry](08_telemetry.md)**
1. **[Params](09_params.md)**
1. **[Governance](10_governance.md)**