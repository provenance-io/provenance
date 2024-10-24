<!--
order: 4
-->

# Trigger Queries

In this section we describe the queries available for looking up trigger information.

<!-- TOC 2 -->
  - [Query/TriggerByID](#querytriggerbyid)
  - [Query/Triggers](#querytriggers)


---
## Query/TriggerByID

The `QueryTriggerByID` query is used to obtain the content of a specific Trigger.

### Request

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/trigger/v1/query.proto#L25-L29

The `id` is the unique identifier for the Trigger.

### Response

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/trigger/v1/query.proto#L31-L35


---
## Query/Triggers

The `QueryTriggers` query is used to obtain all Triggers.

### Request

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/trigger/v1/query.proto#L37-L41

### Response

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/trigger/v1/query.proto#L43-L49
