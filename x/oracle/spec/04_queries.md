<!--
order: 4
-->

# Queries

In this section we describe the queries available for looking up oracle information.

<!-- TOC 2 -->
  - [Query Oracle Address](#query-oracle-address)
  - [Query Oracle](#query-oracle)

---
## Query Oracle Address
The `QueryOracleAddress` query is used to obtain the address of the module's oracle.

### Request

+++ https://github.com/provenance-io/provenance/blob/5afab1b1797b0071cf6a19ea5928c5b8f8831329/proto/provenance/oracle/v1/query.proto#L26-L27

### Response

+++ https://github.com/provenance-io/provenance/blob/5afab1b1797b0071cf6a19ea5928c5b8f8831329/proto/provenance/oracle/v1/query.proto#L29-L33


---
## Query Oracle
The `QueryOracle` query forwards a query to the module's oracle.

### Request

+++ https://github.com/provenance-io/provenance/blob/5afab1b1797b0071cf6a19ea5928c5b8f8831329/proto/provenance/oracle/v1/query.proto#L35-L39

### Response

+++ https://github.com/provenance-io/provenance/blob/5afab1b1797b0071cf6a19ea5928c5b8f8831329/proto/provenance/oracle/v1/query.proto#L41-L45

The data from the `query` field is a `CosmWasm query` forwarded to the `oracle`. 
