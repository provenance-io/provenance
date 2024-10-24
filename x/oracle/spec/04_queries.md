<!--
order: 4
-->

# Queries

In this section we describe the queries available for looking up oracle information.

<!-- TOC 2 -->
  - [Query/OracleAddress](#queryoracleaddress)
  - [Query/Oracle](#queryoracle)

---
## Query/OracleAddress
The `QueryOracleAddress` query is used to obtain the address of the module's oracle.

### Request

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/oracle/v1/query.proto#L25-L26

### Response

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/oracle/v1/query.proto#L28-L32


---
## Query/Oracle
The `QueryOracle` query forwards a query to the module's oracle.

### Request

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/oracle/v1/query.proto#L34-L38

### Response

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/oracle/v1/query.proto#L40-L44

The data from the `query` field is a `CosmWasm query` forwarded to the `oracle`. 
