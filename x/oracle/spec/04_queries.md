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

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/oracle/v1/query.proto#L27-L28

### Response

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/oracle/v1/query.proto#L30-L34


---
## Query/Oracle
The `QueryOracle` query forwards a query to the module's oracle.

### Request

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/oracle/v1/query.proto#L36-L40

### Response

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/oracle/v1/query.proto#L42-L46

The data from the `query` field is a `CosmWasm query` forwarded to the `oracle`. 
