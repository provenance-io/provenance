# Exchange Params

The exchange module params dictate how much of the fees (collected by markets) go to the exchange/chain.
The split values are in basis points and are limited to between `0` and `10,000` (both inclusive).
The `default_split` is used when a specific `DenomSplit` does not exist for a given denom.

* A split of `0` is 0% and would mean that the exchange receives none of the fees (of the applicable denom), and the market keeps all of it.
* A split of `500` is 5%, and would mean that the exchange receives 5% of the fees (of the applicable denom) collected by any market, and the market keeps 95%.
* A split of `10,000` is 100% and would mean that the exchange receives all of the fees (of the applicable denom) and the market gets nothing.

The exchange module params also dictate the fees associated with payments.
The `fee_create_payment_flat` is assessed as a msg fee when creating a payment (paid by the caller/source).
The `fee_accept_payment_flat` is assessed as a msg fee when accepting a payment (paid by the caller/target).

The default `Params` have a `default_split` of `500` and no `DenomSplit`s.
The default `fee_create_payment_flat` and `fee_accept_payment_flat` are each 100,000,000 `nhash` (0.1 `hash`).

Params are set using the [UpdateParams](03_messages.md#updateparams) governance proposal endpoint.

The current params can be looked up using the [Params](05_queries.md#params) query.

See also: [Exchange Fees](01_concepts.md#exchange-fees).

## Params

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/params.proto#L13-L31

## DenomSplit

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/params.proto#L33-L40
