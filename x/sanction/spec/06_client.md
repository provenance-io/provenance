# Client

A user can interact with the `x/sanction` module using `gRPC`, `CLI`, or `REST`.

## gRPC

A user can interact with and query the `x/sanction` module using `gRPC`.

For details see [Msg Service](03_messages.md) or [gRPC Queries](05_queries.md).

## CLI

The `gRPC` transaction and query endpoints are made available through CLI helpers.

### Transactions

The transaction endpoints are only for use with governance proposals.
As such, the CLI's `tx gov` commands can be used to interact with them.

### Queries

Each of these commands facilitates running a `gRPC` query.
Standard `query` flags are available unless otherwise noted.

#### IsSanctioned

```shell
$ provenanced query sanction is-sanctioned --help
Check if an address is sanctioned.

Examples:
  $ provenanced query sanction is-sanctioned pb1v4uxzmtsd3j4zat9wfu5zerywgc47h6l4dhfq4
  $ provenanced query sanction is pb1v4uxzmtsd3j4zat9wfu5zerywgc47h6l4dhfq4
  $ provenanced query sanction check pb1v4uxzmtsd3j4zat9wfu5zerywgc47h6l4dhfq4

Usage:
  provenanced query sanction is-sanctioned <address> [flags]

Aliases:
  is-sanctioned, is, check, is-sanction
```

#### SanctionedAddresses

```shell
$ provenanced query sanction sanctioned-addresses --help
List all the sanctioned addresses.

Examples:
  $ provenanced query sanction sanctioned-addresses
  $ provenanced query sanction addresses
  $ provenanced query sanction all

Usage:
  provenanced query sanction sanctioned-addresses [flags]

Aliases:
  sanctioned-addresses, addresses, all
```

Standard pagination flags are also available for this command.

#### TemporaryEntries

```shell
$ provenanced query sanction temporary-entries --help
List all temporarily sanctioned/unsanctioned addresses.
If an address is provided, only temporary entries for that address are returned.
Otherwise, all temporary entries are returned.

Examples:
  $ provenanced query sanction temporary-entries
  $ provenanced query sanction temporary-entries pb1v4uxzmtsd3j4zat9wfu5zerywgc47h6l4dhfq4
  $ provenanced query sanction temp-entries
  $ provenanced query sanction temp-entries pb1v4uxzmtsd3j4zat9wfu5zerywgc47h6l4dhfq4
  $ provenanced query sanction temp
  $ provenanced query sanction temp pb1v4uxzmtsd3j4zat9wfu5zerywgc47h6l4dhfq4

Usage:
  provenanced query sanction temporary-entries [<address>] [flags]

Aliases:
  temporary-entries, temp-entries, temp
```

Standard pagination flags are also available for this command.

#### Params

```shell
$ provenanced query sanction params --help
Get the sanction module params.

Example:
  $ provenanced query sanction params

Usage:
  provenanced query sanction params [flags]
```

## REST

Each of the sanction `gRPC` query endpoints is also available through one or more `REST` endpoints.

| Name                        | URL                                              |
|-----------------------------|--------------------------------------------------|
| IsSanctioned                | `/provenance/sanction/v1/check/{address}`        |
| SanctionedAddresses         | `/provenance/sanction/v1/all`                    |
| TemporaryEntries - all      | `/provenance/sanction/v1/temp`                   |
| TemporaryEntries - specific | `/provenance/sanction/v1/temp?address={address}` |
| Params                      | `/provenance/sanction/v1/params`                 |

For `SanctionedAddresses` and `TemporaryEntries`, pagination parameters can be provided using the standard pagination query parameters.