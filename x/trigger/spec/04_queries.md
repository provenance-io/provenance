<!--
order: 4
-->

# Trigger Queries

In this section we describe the queries available for looking up trigger information.

<!-- TOC 2 -->
  - [Query Trigger By ID](#query-trigger-by-id)
  - [Query Triggers](#query-triggers)


---
## Query Trigger By ID

The `QueryTriggerByID` query is used to obtain the content of a specific Trigger.

### Request

+++ https://github.com/provenance-io/provenance/blob/bda28e5f58a4a58e8fef21141400ad362b84518b/proto/provenance/trigger/v1/query.proto#L25-L29

The `id` is the unique identifier for the Trigger.

### Response

+++ https://github.com/provenance-io/provenance/blob/bda28e5f58a4a58e8fef21141400ad362b84518b/proto/provenance/trigger/v1/query.proto#L31-L35


---
## Query Triggers

The `QueryTriggers` query is used to obtain all Triggers.

### Request

+++ https://github.com/provenance-io/provenance/blob/bda28e5f58a4a58e8fef21141400ad362b84518b/proto/provenance/trigger/v1/query.proto#L37-L41

### Response

+++ https://github.com/provenance-io/provenance/blob/bda28e5f58a4a58e8fef21141400ad362b84518b/proto/provenance/trigger/v1/query.proto#L43-L49
