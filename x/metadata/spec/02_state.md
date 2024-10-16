# Metadata State

The Metadata module manages the state of several types of entries related to off-chain information.

<!-- TOC -->
  - [Entries](#entries)
    - [Scopes](#scopes)
    - [Sessions](#sessions)
    - [Records](#records)
  - [Specifications](#specifications)
    - [Scope Specifications](#scope-specifications)
    - [Contract Specifications](#contract-specifications)
    - [Record Specifications](#record-specifications)
  - [Object Store Locators](#object-store-locators)



## Entries

The term "entries" refers to scopes, sessions, and records.
They group and identify information.

### Scopes

A scope is a high-level grouping of information combined with some access control.

* A scope must conform to a pre-determined scope specification.
* A scope is used to group together many sessions and records.

#### Scope Keys (Metadata Addresses)

Byte Array Length: `17`

| Byte range | Description         |
|------------|---------------------|
| 0          | `0x00`              |
| 1-16       | UUID of this scope. |

* Field Name: `Scope.scope_id`
* Bech32 HRP: `"scope"`
* Bech32 Example: `"scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel"`

#### Scope Values
<!-- link message: Scope -->

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/metadata/v1/scope.proto#L70-L102

```protobuf
// Scope defines a root reference for a collection of records owned by one or more parties.
message Scope {
  option (gogoproto.goproto_stringer) = false;

  // Unique ID for this scope.  Implements sdk.Address interface for use where addresses are required in Cosmos
  bytes scope_id = 1 [(gogoproto.nullable) = false, (gogoproto.customtype) = "MetadataAddress"];
  // the scope specification that contains the specifications for data elements allowed within this scope
  bytes specification_id = 2 [(gogoproto.nullable) = false, (gogoproto.customtype) = "MetadataAddress"];
  // These parties represent top level owners of the records within.  These parties must sign any requests that modify
  // the data within the scope.  These addresses are in union with parties listed on the sessions.
  repeated Party owners = 3 [(gogoproto.nullable) = false];
  // Addresses in this list are authorized to receive off-chain data associated with this scope.
  repeated string data_access = 4;
  // The address that controls the value associated with this scope.
  //
  // The value owner is actually tracked by the bank module using a coin with the denom "nft/<scope_id>".
  // The value owner can be changed using WriteScope or anything that transfers funds, e.g. MsgSend.
  //
  // During WriteScope:
  //  - If this field is empty, it indicates that there should not be a change to the value owner.
  //    I.e. Once a scope has a value owner, it will always have one (until it's deleted).
  //  - If this field has a value, the existing value owner will be looked up, and
  //    - If there's already an existing value owner, they must be a signer,
  //      and the coin will be transferred to the new value owner.
  //    - If there isn't yet a value owner, the coin will be minted and sent to the new value owner.
  //      If the scope already exists, the owners must be signers (just like changing other fields).
  //      If it's a new scope, there's no special signer limitations related to the value owner.
  string value_owner_address = 5;
  // Whether all parties in this scope and its sessions must be present in this scope's owners field.
  // This also enables use of optional=true scope owners and session parties.
  bool require_party_rollup = 6;
}
```

Before a scope is stored in state, the `value_owner_address` is cleared out (set to an empty string).
The scope is then protobuf encoded, and those bytes are the value stored in state.

#### Scope Value Owners

The `value_owner_address` is tracked using the `x/bank` module. When a scope first gets a value owner (either upon scope
creation, or later with an update), a single coin with the denom `nft/<scope_id>` is minted and placed in the value
owner's account. That coin can be transferred or traded the same ways as any other on-chain funds, e.g. via `MsgSend`.

#### Scope Indexes

Scopes by owner:
* Type byte: `0x17`
* Part 1: The owner address (length byte then value bytes)
* Part 2: All bytes of the scope key

<!-- This index also appears in the section for scope specification indexes. They must stay the same. -->
Scopes by Scope Specification:
* Type byte: `0x11`
* Part 1: All bytes of the scope specification key
* Part 2: All bytes of the scope key



### Sessions

A session is a grouping of records and the parties in charge of those records.

* A session must conform to a pre-determined contract specification.
* A session groups together a collection of records.
* A session is part of exactly one scope.

#### Session Keys (Metadata Addresses)

Byte Array Length: `33`

| Byte range | Description                                    |
|------------|------------------------------------------------|
| 0          | `0x01`                                         |
| 1-16       | UUID of the scope that this session is part of |
| 17-32      | UUID of this session                           |

* Field Name: `Session.session_id`
* Bech32 HRP: `"session"`
* Bech32 Example: `"session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr"`

#### Session Values
<!-- link message: Session -->

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/metadata/v1/scope.proto#L104-L123

```protobuf
// Session defines an execution context against a specific specification instance.
// The context will have a specification and set of parties involved.
//
// NOTE: When there are no more Records within a Scope that reference a Session, the Session is removed.
message Session {
  option (gogoproto.goproto_stringer) = false;

  bytes session_id = 1 [(gogoproto.nullable) = false, (gogoproto.customtype) = "MetadataAddress"];
  // unique id of the contract specification that was used to create this session.
  bytes specification_id = 2 [(gogoproto.nullable) = false, (gogoproto.customtype) = "MetadataAddress"];
  // parties is the set of identities that signed this contract
  repeated Party parties = 3 [(gogoproto.nullable) = false];
  // name to associate with this session execution context, typically classname
  string name = 4;
  // context is a field for storing client specific data associated with a session.
  bytes context = 5;
  // Created by, updated by, timestamps, version number, and related info.
  AuditFields audit = 99;
}
```

#### Session Indexes

There are no extra indexes involving sessions.
Note, though, that the session key is constructed in a way that automatically indexes sessions by scope.



### Records

A record identifies the inputs and outputs of a process.
It is conceptually similar to the values involved in a method call.

* A record must conform to a pre-determined record specification.
* A record is part of exactly one scope.
* A record is part of exactly one session.

#### Record Keys (Metadata Addresses)

Byte Array Length: `33`

| Byte range | Description                                                 |
|------------|-------------------------------------------------------------|
| 0          | `0x02`                                                      |
| 1-16       | UUID of the scope that this record is part of               |
| 17-32      | First 16 bytes of the SHA256 checksum of this record's name |

* Field Name: `Record.record_id`
* Bech32 HRP: `"record"`
* Bech32 Example: `"record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3"`

#### Record Values
<!-- link message: Record -->

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/metadata/v1/scope.proto#L125-L142

```protobuf
// A record (of fact) is attached to a session or each consideration output from a contract
message Record {
  option (gogoproto.goproto_stringer) = false;

  // name/identifier for this record.  Value must be unique within the scope.  Also known as a Fact name
  string name = 1;
  // id of the session context that was used to create this record (use with filtered kvprefix iterator)
  bytes session_id = 2 [(gogoproto.nullable) = false, (gogoproto.customtype) = "MetadataAddress"];
  // process contain information used to uniquely identify an execution on or off chain that generated this record
  Process process = 3 [(gogoproto.nullable) = false];
  // inputs used with the process to achieve the output on this record
  repeated RecordInput inputs = 4 [(gogoproto.nullable) = false];
  // output(s) is the results of executing the process on the given process indicated in this record
  repeated RecordOutput outputs = 5 [(gogoproto.nullable) = false];
  // specification_id is the id of the record specification that was used to create this record.
  bytes specification_id = 6 [(gogoproto.nullable) = false, (gogoproto.customtype) = "MetadataAddress"];
}
```

#### Record Indexes

There are no extra indexes involving records.
Note, though, that the record key is constructed in a way that automatically indexes records by scope.



## Specifications

The term "specifications" refers to scope specifications, contract specifications, and record specifications.
They define validation parameters for the various entries.
Ideally, specifications will be used for multiple entries.

### Scope Specifications

A scope specification defines validation parameters for scopes.
They group together contract specifications and define roles that must be involved in a scope.

#### Scope Specification Keys (Metadata Addresses)

Byte Array Length: `17`

| Byte range | Description                      |
|------------|----------------------------------|
| 0          | `0x04`                           |
| 1-16       | UUID of this scope specification |

* Field Name: `ScopeSpecification.specification_id`
* Bech32 HRP: `"scopespec"`
* Bech32 Example: `"scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m"`

#### Scope Specification Values
<!-- link message: ScopeSpecification -->

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/metadata/v1/specification.proto#L36-L51

```protobuf
// ScopeSpecification defines the required parties, resources, conditions, and consideration outputs for a contract
message ScopeSpecification {
  option (gogoproto.goproto_stringer) = false;

  // unique identifier for this specification on chain
  bytes specification_id = 1 [(gogoproto.nullable) = false, (gogoproto.customtype) = "MetadataAddress"];
  // General information about this scope specification.
  Description description = 2;
  // Addresses of the owners of this scope specification.
  repeated string owner_addresses = 3;
  // A list of parties that must be present on a scope (and their associated roles)
  repeated PartyType parties_involved = 4;
  // A list of contract specification ids allowed for a scope based on this specification.
  repeated bytes contract_spec_ids = 5 [(gogoproto.nullable) = false, (gogoproto.customtype) = "MetadataAddress"];
}
```

#### Scope Specification Indexes

Scope specifications by owner:
* Type byte: `0x19`
* Part 1: The owner address (length byte then value bytes)
* Part 2: All bytes of the scope specification key

<!-- This index also appears in the section for contract specification indexes.  They must stay the same. -->
Scope Specifications by contract specification:
* Type byte: `0x14`
* Part 1: All bytes of the contract specification key
* Part 2: All bytes of the scope specification key

<!-- This index also appears in the section for scope indexes. They must stay the same. -->
Scopes by Scope Specification:
* Type byte: `0x11`
* Part 1: All bytes of the scope specification key
* Part 2: All bytes of the scope key



### Contract Specifications

A contract specification defines validation parameters for sessions.
They contain source information and roles that must be involved in a session.
They also group together record specifications.

A contract specification can be part of multiple scope specifications.

#### Contract Specification Keys (Metadata Addresses)

Byte Array Length: `17`

| Byte range | Description                         |
|------------|-------------------------------------|
| 0          | `0x03`                              |
| 1-16       | UUID of this contract specification |

* Field Name: `ContractSpecification.specification_id`
* Bech32 HRP: `"contractspec"`
* Bech32 Example: `"contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn"`

#### Contract Specification Values
<!-- link message: ContractSpecification -->

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/metadata/v1/specification.proto#L53-L76

```protobuf
// ContractSpecification defines the required parties, resources, conditions, and consideration outputs for a contract
message ContractSpecification {
  option (gogoproto.goproto_stringer) = false;

  // unique identifier for this specification on chain
  bytes specification_id = 1 [(gogoproto.nullable) = false, (gogoproto.customtype) = "MetadataAddress"];
  // Description information for this contract specification
  Description description = 2;
  // Address of the account that owns this specificaiton
  repeated string owner_addresses = 3;
  // a list of party roles that must be fullfilled when signing a transaction for this contract specification
  repeated PartyType parties_involved = 4;
  // Reference to a metadata record with a hash and type information for the instance of code that will process this
  // contract
  oneof source {
    // the address of a record on chain that represents this contract
    bytes resource_id = 5 [(gogoproto.casttype) = "github.com/cosmos/cosmos-sdk/types.AccAddress"];
    // the hash of contract binary (off-chain instance)
    string hash = 6;
  }
  // name of the class/type of this contract executable
  string class_name = 7;
}
```

#### Contract Specification Indexes

Contract specifications by owner:
* Type byte: `0x20`
* Part 1: The owner address (length byte then value bytes)
* Part 2: All bytes of the contract specification key

<!-- This index also appears in the section for scope specification indexes. They must stay the same. -->
Scope Specifications by contract specification:
* Type byte: `0x14`
* Part 1: All bytes of the contract specification key
* Part 2: All bytes of the scope specification key



### Record Specifications

A record specification defines validation parameters for records.
They contain expected inputs and outputs and parties that must be involved in a record.

A record specification is part of exactly one contract specification.

#### Record Specification Keys (Metadata Addresses)

Byte Array Length: `33`

| Byte range | Description                                                                  |
|------------|------------------------------------------------------------------------------|
| 0          | `0x05`                                                                       |
| 1-16       | UUID of the contract specification that this record specification is part of |
| 17-32      | First 16 bytes of the SHA256 checksum of this record specification's name    |

* Field Name: `RecordSpecification.specification_id`
* Bech32 HRP: `"recspec"`
* Bech32 Example: `"recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44"`

#### Record Specification Values
<!-- link message: RecordSpecification -->

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/metadata/v1/specification.proto#L78-L95

```protobuf
// RecordSpecification defines the specification for a Record including allowed/required inputs/outputs
message RecordSpecification {
  option (gogoproto.goproto_stringer) = false;

  // unique identifier for this specification on chain
  bytes specification_id = 1 [(gogoproto.nullable) = false, (gogoproto.customtype) = "MetadataAddress"];
  // Name of Record that will be created when this specification is used
  string name = 2;
  // A set of inputs that must be satisified to apply this RecordSpecification and create a Record
  repeated InputSpecification inputs = 3;
  // A type name for data associated with this record (typically a class or proto name)
  string type_name = 4;
  // Type of result for this record specification (must be RECORD or RECORD_LIST)
  DefinitionType result_type = 5;
  // Type of party responsible for this record
  repeated PartyType responsible_parties = 6;
}
```

#### Record Specification Indexes

There are no extra indexes involving record specifications.
Note, though, that the record key is constructed in a way that automatically indexes record specifications by contract specification.



## Object Store Locators

An object store locator indicates the location of off-chain data.

#### Object Store Locator Keys

Byte Array Length: `21`

| Byte range   | Description                                             |
|--------------|---------------------------------------------------------|
| 0            | `0x21`                                                  |
| 1            | Owner address length, either `0x14` (20) or `0x20` (32) |
| 2-(21 or 33) | The bytes of the owner address.                         |

#### Object Store Locator Values
<!-- link message: ObjectStoreLocator -->

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/metadata/v1/objectstore.proto#L12-L23

```protobuf
// Defines an Locator object stored on chain, which represents a owner( blockchain address) associated with a endpoint
// uri for it's associated object store.
message ObjectStoreLocator {
  option (cosmos.msg.v1.signer) = "owner";

  // account address the endpoint is owned by
  string owner = 1;
  // locator endpoint uri
  string locator_uri = 2;
  // owners encryption key address
  string encryption_key = 3;
}
```

#### Object Store Locator Indexes

There are no extra indexes involving object store locators.
