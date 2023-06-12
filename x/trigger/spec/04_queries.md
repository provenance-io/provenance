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

+++ https://github.com/provenance-io/provenance/blob/1b12a2c9115cbff9c9868d3aec9a671776c74976/proto/provenance/trigger/v1/query.proto#L26-L29

The `id` is the unique identifier for the Trigger.

### Response

+++ https://github.com/provenance-io/provenance/blob/1b12a2c9115cbff9c9868d3aec9a671776c74976/proto/provenance/trigger/v1/query.proto#L32-L35


---
## Query Triggers

The `QueryTriggers` query is used to obtain all Triggers.

### Request

+++ https://github.com/provenance-io/provenance/blob/1b12a2c9115cbff9c9868d3aec9a671776c74976/proto/provenance/trigger/v1/query.proto#L38-L41

### Response

+++ https://github.com/provenance-io/provenance/blob/1b12a2c9115cbff9c9868d3aec9a671776c74976/proto/provenance/trigger/v1/query.proto#L44-L49
