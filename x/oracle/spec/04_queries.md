<!--
order: 4
-->

# Queries

In this section we describe the queries available for looking up oracle information.

---
## Query Oracle Address
The `QueryOracleAddress` query is used to obtain the address of the module's oracle.

### Request

+++ https://github.com/provenance-io/provenance/blob/65865991f93e2c1a7647e29be11f6527f49616e6/proto/provenance/oracle/v1/query.proto#L42-L43

### Response

+++ https://github.com/provenance-io/provenance/blob/65865991f93e2c1a7647e29be11f6527f49616e6/proto/provenance/oracle/v1/query.proto#L45-L49


---
## Query Oracle
The `QueryOracle` query forwards a query to the module's oracle.

### Request

+++ https://github.com/provenance-io/provenance/blob/65865991f93e2c1a7647e29be11f6527f49616e6/proto/provenance/oracle/v1/query.proto#L51-L55

### Response

+++ https://github.com/provenance-io/provenance/blob/65865991f93e2c1a7647e29be11f6527f49616e6/proto/provenance/oracle/v1/query.proto#L57-L61

The data from the `query` field is a `CosmWasm query` forwarded to the `oracle`. 

---
## Query Query State
The `QueryQueryState` query is used to obtain the state of an existing query.

### Request

+++ https://github.com/provenance-io/provenance/blob/65865991f93e2c1a7647e29be11f6527f49616e6/proto/provenance/oracle/v1/query.proto#L31-L34

### Response

+++ https://github.com/provenance-io/provenance/blob/65865991f93e2c1a7647e29be11f6527f49616e6/proto/provenance/oracle/v1/query.proto#L36-L40