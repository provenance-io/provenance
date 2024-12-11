# grpcurl

Interact with Provenance blockchain via `grpcurl`.

The examples below are run on a `localnet` docker nodes. For instructions on running your own `localnet` see the [quick start guide](https://github.com/provenance-io/provenance#quick-start) on the provenance repo.

<!-- TOC -->
  - [Installing gRPCurl](#installing-grpcurl)
  - [Clone Provenance](#clone-provenance)
  - [Querying Services](#querying-services)
    - [Bank Service](#bank-service)
    - [Authz Service](#authz-service)
    - [Marker Service](#marker-service)
    - [Metadata Service](#metadata-service)

## Installing gRPCurl

Homebrew (macOS)

```shell
brew install grpcurl
```

- For other methods of installing see  [docs](https://github.com/fullstorydev/grpcurl#installation).

## Clone Provenance

To query services we need to manually reference to relevant `.proto` files. Let's go ahead and clone the `provenance` project.

```shell
git clone https://github.com/provenance-io/provenance.git
```

```shell
cd ./provenance
make build
make localnet-start # requires docker
```


## Querying Services

This section includes query examples for the `bank`, `authz`, `marker`,  and `metadata` modules.

For a full list of all available services, run:

```shell
grpcurl -plaintext localhost:9090 list
```

### Bank Service

#### Describe service methods

Describe available methods of the `bank` service.

```shell
grpcurl \
    -import-path ./proto \
    -import-path ./third_party/proto \
    -proto ./third_party/proto/cosmos/bank/v1beta1/query.proto \
    localhost:9090 \
    describe cosmos.bank.v1beta1.Query
```

Running the command above will result in something like the below out.

```shell
cosmos.bank.v1beta1.Query is a service:
// Query defines the gRPC querier service.
service Query {
  // AllBalances queries the balance of all coins for a single account.
  rpc AllBalances ( .cosmos.bank.v1beta1.QueryAllBalancesRequest ) returns ( .cosmos.bank.v1beta1.QueryAllBalancesResponse ) {
    option (.google.api.http) = { get:"/cosmos/bank/v1beta1/balances/{address}"  };
  }
  // Balance queries the balance of a single coin for a single account.
  rpc Balance ( .cosmos.bank.v1beta1.QueryBalanceRequest ) returns ( .cosmos.bank.v1beta1.QueryBalanceResponse ) {
    option (.google.api.http) = { get:"/cosmos/bank/v1beta1/balances/{address}/{denom}"  };
  }
  // DenomsMetadata queries the client metadata of a given coin denomination.
  rpc DenomMetadata ( .cosmos.bank.v1beta1.QueryDenomMetadataRequest ) returns ( .cosmos.bank.v1beta1.QueryDenomMetadataResponse ) {
    option (.google.api.http) = { get:"/cosmos/bank/v1beta1/denoms_metadata/{denom}"  };
  }
  // DenomOwners queries for all account addresses that own a particular token
  // denomination.
  rpc DenomOwners ( .cosmos.bank.v1beta1.QueryDenomOwnersRequest ) returns ( .cosmos.bank.v1beta1.QueryDenomOwnersResponse ) {
    option (.google.api.http) = { get:"/cosmos/bank/v1beta1/denom_owners/{denom}"  };
  }
  // DenomsMetadata queries the client metadata for all registered coin
  // denominations.
  rpc DenomsMetadata ( .cosmos.bank.v1beta1.QueryDenomsMetadataRequest ) returns ( .cosmos.bank.v1beta1.QueryDenomsMetadataResponse ) {
    option (.google.api.http) = { get:"/cosmos/bank/v1beta1/denoms_metadata"  };
  }
  // Params queries the parameters of x/bank module.
  rpc Params ( .cosmos.bank.v1beta1.QueryParamsRequest ) returns ( .cosmos.bank.v1beta1.QueryParamsResponse ) {
    option (.google.api.http) = { get:"/cosmos/bank/v1beta1/params"  };
  }
  // SupplyOf queries the supply of a single coin.
  rpc SupplyOf ( .cosmos.bank.v1beta1.QuerySupplyOfRequest ) returns ( .cosmos.bank.v1beta1.QuerySupplyOfResponse ) {
    option (.google.api.http) = { get:"/cosmos/bank/v1beta1/supply/{denom}"  };
  }
  // TotalSupply queries the total supply of all coins.
  rpc TotalSupply ( .cosmos.bank.v1beta1.QueryTotalSupplyRequest ) returns ( .cosmos.bank.v1beta1.QueryTotalSupplyResponse ) {
    option (.google.api.http) = { get:"/cosmos/bank/v1beta1/supply"  };
  }
}
```

#### Query service methods

In the below example we use the `AllBalances` method to query a particular `address`.

Note the use of `-plaintext` and `-d {"address": "<address>"}`

```shell
grpcurl \
    -plaintext \
    -import-path ./proto \
    -import-path ./third_party/proto \
    -proto ./third_party/proto/cosmos/bank/v1beta1/query.proto \
    -d '{"address": "tp1cw2vkz4h4zhc6mtfxrh67q9v2k0d76970nvwrg"}' \
    localhost:9090 \
    cosmos.bank.v1beta1.Query/AllBalances
```

```shell
{
  "balances": [
    {
      "denom": "nhash",
      "amount": "24999997751000000000"
    }
  ],
  "pagination": {
    "total": "1"
  }
}
```

### Authz Service

#### Describe service methods

List available methods of the `authz` service.

```shell
grpcurl \
    -import-path ./proto \
    -import-path ./third_party/proto \
    -proto ./third_party/proto/cosmos/authz/v1beta1/query.proto \
    localhost:9090 \
    describe cosmos.authz.v1beta1.Query
```

```shell
cosmos.authz.v1beta1.Query is a service:
// Query defines the gRPC querier service.
service Query {
  // Returns list of `Authorization`, granted to the grantee by the granter.
  rpc Grants ( .cosmos.authz.v1beta1.QueryGrantsRequest ) returns ( .cosmos.authz.v1beta1.QueryGrantsResponse ) {
    option (.google.api.http) = { get:"/cosmos/authz/v1beta1/grants"  };
  }
}
```

#### Query service methods

Query for grants:

```shell
grpcurl \
    -plaintext \
    -import-path ./proto \
    -import-path ./third_party/proto \
    -proto ./third_party/proto/cosmos/authz/v1beta1/query.proto \
    -d '{"granter": "<address>", "grantee":"<address>"}' \
    localhost:9090 \
    cosmos.authz.v1beta1.Query/Grants
```

You should see the below when no grants are found:
```shell
{
  "pagination": {
    
  }
}
```

To see a response with grants in it, run the below command to `add` a grant:
```shell
./build/provenanced tx authz grant <grantee-address> send \
    --spend-limit 100000nhash \
    --from <granter-address> \
    --fees 1000000000nhash \
    -t \
    --home ./build/node0 \
    --chain-id chain-local \
    --keyring-backend test
```

### Marker Service

#### Describe service methods

List available methods of the `marker` module.

```shell
grpcurl \
    -import-path ./proto \
    -import-path ./third_party/proto \
    -proto ./proto/provenance/marker/v1/query.proto \
    localhost:9090 \
    describe provenance.marker.v1.Query
```

You should see something like the below.
```shell
provenance.marker.v1.Query is a service:
// Query defines the gRPC querier service for marker module.
service Query {
  // query for access records on an account
  rpc Access ( .provenance.marker.v1.QueryAccessRequest ) returns ( .provenance.marker.v1.QueryAccessResponse ) {
    option (.google.api.http) = { get:"/provenance/marker/v1/accesscontrol/{id}"  };
  }
  // Returns a list of all markers on the blockchain
  rpc AllMarkers ( .provenance.marker.v1.QueryAllMarkersRequest ) returns ( .provenance.marker.v1.QueryAllMarkersResponse ) {
    option (.google.api.http) = { get:"/provenance/marker/v1/all"  };
  }
  // query for access records on an account
  rpc DenomMetadata ( .provenance.marker.v1.QueryDenomMetadataRequest ) returns ( .provenance.marker.v1.QueryDenomMetadataResponse ) {
    option (.google.api.http) = { get:"/provenance/marker/v1/getdenommetadata/{denom}"  };
  }
  // query for coins on a marker account
  rpc Escrow ( .provenance.marker.v1.QueryEscrowRequest ) returns ( .provenance.marker.v1.QueryEscrowResponse ) {
    option (.google.api.http) = { get:"/provenance/marker/v1/escrow/{id}"  };
  }
  // query for all accounts holding the given marker coins
  rpc Holding ( .provenance.marker.v1.QueryHoldingRequest ) returns ( .provenance.marker.v1.QueryHoldingResponse ) {
    option (.google.api.http) = { get:"/provenance/marker/v1/holding/{id}"  };
  }
  // query for a single marker by denom or address
  rpc Marker ( .provenance.marker.v1.QueryMarkerRequest ) returns ( .provenance.marker.v1.QueryMarkerResponse ) {
    option (.google.api.http) = { get:"/provenance/marker/v1/detail/{id}"  };
  }
  // Params queries the parameters of x/bank module.
  rpc Params ( .provenance.marker.v1.QueryParamsRequest ) returns ( .provenance.marker.v1.QueryParamsResponse ) {
    option (.google.api.http) = { get:"/provenance/marker/v1/params"  };
  }
  // query for supply of coin on a marker account
  rpc Supply ( .provenance.marker.v1.QuerySupplyRequest ) returns ( .provenance.marker.v1.QuerySupplyResponse ) {
    option (.google.api.http) = { get:"/provenance/marker/v1/supply/{id}"  };
  }
}
```

#### Query service methods

Lets query for the current `nhash` supply.

```shell
grpcurl \
    -plaintext \
    -import-path ./proto \
    -import-path ./third_party/proto \
    -proto ./proto/provenance/marker/v1/query.proto \
    -d '{"id": "nhash"}' \
    localhost:9090 \
    provenance.marker.v1.Query/Supply
```

```shell
{
  "amount": {
    "denom": "nhash",
    "amount": "100000000000000000000"
  }
}
```

### Metadata Service

#### Describe servie methods

List available methods of the `metadata` module.

```shell
grpcurl \
    -import-path ./proto \
    -import-path ./third_party/proto \
    -proto ./proto/provenance/metadata/v1/query.proto \
    localhost:9090 \
    describe provenance.metadata.v1.Query
```

```shell
provenance.metadata.v1.Query is a service:
// Query defines the Metadata Query service.
service Query {
  // ContractSpecification returns a contract specification for the given specification id.
  //
  // The specification_id can either be a uuid, e.g. def6bc0a-c9dd-4874-948f-5206e6060a84, a bech32 contract
  // specification address, e.g. contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn, or a bech32 record specification
  // address, e.g. recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44. If it is a record specification
  // address, then the contract specification that contains that record specification is looked up.
  //
  // By default, the record specifications for this contract specification are not included.
  // Set include_record_specs to true to include them in the result.
  rpc ContractSpecification ( .provenance.metadata.v1.ContractSpecificationRequest ) returns ( .provenance.metadata.v1.ContractSpecificationResponse ) {
    option (.google.api.http) = { get:"/provenance/metadata/v1/contractspec/{specification_id}"  };
  }
  // ContractSpecificationsAll retrieves all contract specifications.
  rpc ContractSpecificationsAll ( .provenance.metadata.v1.ContractSpecificationsAllRequest ) returns ( .provenance.metadata.v1.ContractSpecificationsAllResponse ) {
    option (.google.api.http) = { get:"/provenance/metadata/v1/contractspecs/all"  };
  }
  // OSAllLocators returns all ObjectStoreLocator entries.
  rpc OSAllLocators ( .provenance.metadata.v1.OSAllLocatorsRequest ) returns ( .provenance.metadata.v1.OSAllLocatorsResponse ) {
    option (.google.api.http) = { get:"/provenance/metadata/v1/locators/all"  };
  }
  // OSLocator returns an ObjectStoreLocator by its owner's address.
  rpc OSLocator ( .provenance.metadata.v1.OSLocatorRequest ) returns ( .provenance.metadata.v1.OSLocatorResponse ) {
    option (.google.api.http) = { get:"/provenance/metadata/v1/locator/{owner}"  };
  }
  // OSLocatorParams returns all parameters for the object store locator sub module.
  rpc OSLocatorParams ( .provenance.metadata.v1.OSLocatorParamsRequest ) returns ( .provenance.metadata.v1.OSLocatorParamsResponse ) {
    option (.google.api.http) = { get:"/provenance/metadata/v1/locator/params"  };
  }
  // OSLocatorsByScope returns all ObjectStoreLocator entries for a for all signer's present in the specified scope.
  rpc OSLocatorsByScope ( .provenance.metadata.v1.OSLocatorsByScopeRequest ) returns ( .provenance.metadata.v1.OSLocatorsByScopeResponse ) {
    option (.google.api.http) = { get:"/provenance/metadata/v1/locator/scope/{scope_id}"  };
  }
  // OSLocatorsByURI returns all ObjectStoreLocator entries for a locator uri.
  rpc OSLocatorsByURI ( .provenance.metadata.v1.OSLocatorsByURIRequest ) returns ( .provenance.metadata.v1.OSLocatorsByURIResponse ) {
    option (.google.api.http) = { get:"/provenance/metadata/v1/locator/uri/{uri}"  };
  }
  // Ownership returns the scope identifiers that list the given address as either a data or value owner.
  rpc Ownership ( .provenance.metadata.v1.OwnershipRequest ) returns ( .provenance.metadata.v1.OwnershipResponse ) {
    option (.google.api.http) = { get:"/provenance/metadata/v1/ownership/{address}"  };
  }
  // Params queries the parameters of x/metadata module.
  rpc Params ( .provenance.metadata.v1.QueryParamsRequest ) returns ( .provenance.metadata.v1.QueryParamsResponse ) {
    option (.google.api.http) = { get:"/provenance/metadata/v1/params"  };
  }
  // RecordSpecification returns a record specification for the given input.
  rpc RecordSpecification ( .provenance.metadata.v1.RecordSpecificationRequest ) returns ( .provenance.metadata.v1.RecordSpecificationResponse ) {
    option (.google.api.http) = { get:"/provenance/metadata/v1/recordspec/{specification_id}" additional_bindings:<get:"/provenance/metadata/v1/contractspec/{specification_id}/recordspec/{name}" >  };
  }
  // RecordSpecificationsAll retrieves all record specifications.
  rpc RecordSpecificationsAll ( .provenance.metadata.v1.RecordSpecificationsAllRequest ) returns ( .provenance.metadata.v1.RecordSpecificationsAllResponse ) {
    option (.google.api.http) = { get:"/provenance/metadata/v1/recordspecs/all"  };
  }
  // RecordSpecificationsForContractSpecification returns the record specifications for the given input.
  //
  // The specification_id can either be a uuid, e.g. def6bc0a-c9dd-4874-948f-5206e6060a84, a bech32 contract
  // specification address, e.g. contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn, or a bech32 record specification
  // address, e.g. recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44. If it is a record specification
  // address, then the contract specification that contains that record specification is used.
  rpc RecordSpecificationsForContractSpecification ( .provenance.metadata.v1.RecordSpecificationsForContractSpecificationRequest ) returns ( .provenance.metadata.v1.RecordSpecificationsForContractSpecificationResponse ) {
    option (.google.api.http) = { get:"/provenance/metadata/v1/contractspec/{specification_id}/recordspecs"  };
  }
  // Records searches for records.
  //
  // The record_addr, if provided, must be a bech32 record address, e.g.
  // record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3. The scope-id can either be scope uuid, e.g.
  // 91978ba2-5f35-459a-86a7-feca1b0512e0 or a scope address, e.g. scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel. Similarly,
  // the session_id can either be a uuid or session address, e.g.
  // session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr. The name is the name of the record you're
  // interested in.
  //
  // * If only a record_addr is provided, that single record will be returned.
  // * If only a scope_id is provided, all records in that scope will be returned.
  // * If only a session_id (or scope_id/session_id), all records in that session will be returned.
  // * If a name is provided with a scope_id and/or session_id, that single record will be returned.
  //
  // A bad request is returned if:
  // * The session_id is a uuid and no scope_id is provided.
  // * There are two or more of record_addr, session_id, and scope_id, and they don't all refer to the same scope.
  // * A name is provided, but not a scope_id and/or a session_id.
  // * A name and record_addr are provided and the name doesn't match the record_addr.
  //
  // By default, the scope and sessions are not included.
  // Set include_scope and/or include_sessions to true to include the scope and/or sessions.
  rpc Records ( .provenance.metadata.v1.RecordsRequest ) returns ( .provenance.metadata.v1.RecordsResponse ) {
    option (.google.api.http) = { get:"/provenance/metadata/v1/record/{record_addr}" additional_bindings:<get:"/provenance/metadata/v1/scope/{scope_id}/records" > additional_bindings:<get:"/provenance/metadata/v1/scope/{scope_id}/record/{name}" > additional_bindings:<get:"/provenance/metadata/v1/scope/{scope_id}/session/{session_id}/records" > additional_bindings:<get:"/provenance/metadata/v1/scope/{scope_id}/session/{session_id}/record/{name}" > additional_bindings:<get:"/provenance/metadata/v1/session/{session_id}/records" > additional_bindings:<get:"/provenance/metadata/v1/session/{session_id}/record/{name}" >  };
  }
  // RecordsAll retrieves all records.
  rpc RecordsAll ( .provenance.metadata.v1.RecordsAllRequest ) returns ( .provenance.metadata.v1.RecordsAllResponse ) {
    option (.google.api.http) = { get:"/provenance/metadata/v1/records/all"  };
  }
  // Scope searches for a scope.
  //
  // The scope id, if provided, must either be scope uuid, e.g. 91978ba2-5f35-459a-86a7-feca1b0512e0 or a scope address,
  // e.g. scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel. The session addr, if provided, must be a bech32 session address,
  // e.g. session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr. The record_addr, if provided, must be a
  // bech32 record address, e.g. record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3.
  //
  // * If only a scope_id is provided, that scope is returned.
  // * If only a session_addr is provided, the scope containing that session is returned.
  // * If only a record_addr is provided, the scope containing that record is returned.
  // * If more than one of scope_id, session_addr, and record_addr are provided, and they don't refer to the same scope,
  // a bad request is returned.
  //
  // Providing a session addr or record addr does not limit the sessions and records returned (if requested).
  // Those parameters are only used to find the scope.
  //
  // By default, sessions and records are not included.
  // Set include_sessions and/or include_records to true to include sessions and/or records.
  rpc Scope ( .provenance.metadata.v1.ScopeRequest ) returns ( .provenance.metadata.v1.ScopeResponse ) {
    option (.google.api.http) = { get:"/provenance/metadata/v1/scope/{scope_id}" additional_bindings:<get:"/provenance/metadata/v1/session/{session_addr}/scope" > additional_bindings:<get:"/provenance/metadata/v1/record/{record_addr}/scope" >  };
  }
  // ScopeSpecification returns a scope specification for the given specification id.
  //
  // The specification_id can either be a uuid, e.g. dc83ea70-eacd-40fe-9adf-1cf6148bf8a2 or a bech32 scope
  // specification address, e.g. scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m.
  rpc ScopeSpecification ( .provenance.metadata.v1.ScopeSpecificationRequest ) returns ( .provenance.metadata.v1.ScopeSpecificationResponse ) {
    option (.google.api.http) = { get:"/provenance/metadata/v1/scopespec/{specification_id}"  };
  }
  // ScopeSpecificationsAll retrieves all scope specifications.
  rpc ScopeSpecificationsAll ( .provenance.metadata.v1.ScopeSpecificationsAllRequest ) returns ( .provenance.metadata.v1.ScopeSpecificationsAllResponse ) {
    option (.google.api.http) = { get:"/provenance/metadata/v1/scopespecs/all"  };
  }
  // ScopesAll retrieves all scopes.
  rpc ScopesAll ( .provenance.metadata.v1.ScopesAllRequest ) returns ( .provenance.metadata.v1.ScopesAllResponse ) {
    option (.google.api.http) = { get:"/provenance/metadata/v1/scopes/all"  };
  }
  // Sessions searches for sessions.
  //
  // The scope_id can either be scope uuid, e.g. 91978ba2-5f35-459a-86a7-feca1b0512e0 or a scope address, e.g.
  // scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel. Similarly, the session_id can either be a uuid or session address, e.g.
  // session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr. The record_addr, if provided, must be a
  // bech32 record address, e.g. record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3.
  //
  // * If only a scope_id is provided, all sessions in that scope are returned.
  // * If only a session_id is provided, it must be an address, and that single session is returned.
  // * If the session_id is a uuid, then either a scope_id or record_addr must also be provided, and that single session
  // is returned.
  // * If only a record_addr is provided, the session containing that record will be returned.
  // * If a record_name is provided then either a scope_id, session_id as an address, or record_addr must also be
  // provided, and the session containing that record will be returned.
  //
  // A bad request is returned if:
  // * The session_id is a uuid and is provided without a scope_id or record_addr.
  // * A record_name is provided without any way to identify the scope (e.g. a scope_id, a session_id as an address, or
  // a record_addr).
  // * Two or more of scope_id, session_id as an address, and record_addr are provided and don't all refer to the same
  // scope.
  // * A record_addr (or scope_id and record_name) is provided with a session_id and that session does not contain such
  // a record.
  // * A record_addr and record_name are both provided, but reference different records.
  //
  // By default, the scope and records are not included.
  // Set include_scope and/or include_records to true to include the scope and/or records.
  rpc Sessions ( .provenance.metadata.v1.SessionsRequest ) returns ( .provenance.metadata.v1.SessionsResponse ) {
    option (.google.api.http) = { get:"/provenance/metadata/v1/session/{session_id}" additional_bindings:<get:"/provenance/metadata/v1/scope/{scope_id}/sessions" > additional_bindings:<get:"/provenance/metadata/v1/scope/{scope_id}/session/{session_id}" > additional_bindings:<get:"/provenance/metadata/v1/record/{record_addr}/session" > additional_bindings:<get:"/provenance/metadata/v1/scope/{scope_id}/record/{record_name}/session" >  };
  }
  // SessionsAll retrieves all sessions.
  rpc SessionsAll ( .provenance.metadata.v1.SessionsAllRequest ) returns ( .provenance.metadata.v1.SessionsAllResponse ) {
    option (.google.api.http) = { get:"/provenance/metadata/v1/sessions/all"  };
  }
  // ValueOwnership returns the scope identifiers that list the given address as the value owner.
  rpc ValueOwnership ( .provenance.metadata.v1.ValueOwnershipRequest ) returns ( .provenance.metadata.v1.ValueOwnershipResponse ) {
    option (.google.api.http) = { get:"/provenance/metadata/v1/valueownership/{address}"  };
  }
}
```

#### Query service methods

Lets query for any locators. To do this will query the `OSAllLocators` method.


```shell
grpcurl \
    -plaintext \
    -import-path ./proto \
    -import-path ./third_party/proto \
    -proto ./proto/provenance/metadata/v1/query.proto \
    localhost:9090 \
    provenance.metadata.v1.Query/OSAllLocators
```

```shell
OSAllLocators
{
  "locators": [
    {
      "owner": "tp18jjjggrwvyxqtrmk5l4wv2scav29y4sa3ckzvd",
      "locatorUri": "http://test-2.com"
    },
    {
      "owner": "tp1cw2vkz4h4zhc6mtfxrh67q9v2k0d76970nvwrg",
      "locatorUri": "http://test.com"
    }
  ],
  "request": {
    
  },
  "pagination": {
    "total": "2"
  }
}
```
