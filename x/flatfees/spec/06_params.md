# FlatFees Parameters

The flatfees module params define the default cost and the conversion factor used to convert Msg costs into the fee denom.

Costs, including the default, should be defined using a stable coin denom.
The conversion factor dictates an equivalent amount of fee coin (converted amount) and stable coin (base amount).
The `conversion_factor.definition_amount` should have the same denom as the `default_cost`.
The `conversion_factor.converted_amount` should have the fee denom.

This setup allows us to define the costs in, e.g. `cusd`, and charge them in `nhash`.
Later, when the price of nhash changes (externally), we can update the conversion factor to match; this keeps Msg costs roughly constant in terms of USD (or whatever denom is used). 

Params are set using the [UpdateParams](03_messages.md#updateparams) governance proposal endpoint.

The current params can be looked up using the [Params](05_queries.md#params) query.

## Params

+++ https://github.com/provenance-io/provenance/blob/v1.24.0/proto/provenance/flatfees/v1/flatfees.proto#L12-L20

## ConversionFactor

+++ https://github.com/provenance-io/provenance/blob/v1.24.0/proto/provenance/flatfees/v1/flatfees.proto#L22-L36
