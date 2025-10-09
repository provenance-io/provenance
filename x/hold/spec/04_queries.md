# Queries

The `x/hold` module provides some queries for looking up hold-related data.

<!-- TOC -->
  - [GetHolds](#getholds)
  - [GetAllHolds](#getallholds)

## GetHolds

To look up the funds on hold for an account, use the `GetHolds` query.
The query takes in an `address` and returns a coins `amount`.

Request:

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/hold/v1/query.proto#L31-L38

Response:

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/hold/v1/query.proto#L40-L52

It is expected to fail if the `address` is invalid or missing.

If the account doesn't exist, or no coins are on hold for the account, the amount will be empty.

## GetAllHolds

To get all funds on hold for all accounts, use the `GetAllHolds` query.
The query takes in pagination parameters and returns a list of `address`/`amount` pairs.

Request:

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/hold/v1/query.proto#L54-L61

Response:

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/hold/v1/query.proto#L63-L69

<!-- link message: AccountHold -->

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/hold/v1/hold.proto#L12-L23

It is expected to fail if the pagination parameters are invalid.
