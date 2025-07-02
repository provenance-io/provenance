# FlatFees Parameters

The flatfees module params define the default cost and the conversion factor used to convert Msg costs into the fee denom.

Costs, including the default, should be defined using a stable coin, e.g. `musd` (`1musd` = `$0.001)`.
The conversion factor dictates an equivalent amount of fee coin (converted amount) and stable coin (base amount).
The `conversion_factor.definition_amount` must have the same denom as the `default_cost`.
The `conversion_factor.converted_amount` must have the fee denom.

This setup allows us to define the costs in, e.g. `musd`, and charge them in `nhash`.
Later, when the price of nhash changes (externally) with respects to `musd`, we can update the conversion factor to match; this keeps Msg costs roughly constant in terms of USD (or whatever denom is used).

Params are set using the [UpdateParams](03_messages.md#updateparams) governance proposal endpoint.

The current params can be looked up using the [Params](05_queries.md#params) query.

## Params

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/flatfees/v1/flatfees.proto#L12-L24

## ConversionFactor

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/flatfees/v1/flatfees.proto#L26-L42
