
<!--
order: 6
-->

# Parameters

The MsgFee module contains the following parameter:

| Key                    | Type     | Example                           |
|------------------------|----------|-----------------------------------|
| FloorGasPrice          | `uint32` | `"1905"`                          |
| UsdConversionRate      | `uint64` | `"70"`                            |



FloorGasPrice is the value of base denom that is charged for calculating base fees, for when base fee and additional fee are charged in the base denom.

UsdConversionRate is the conversion rate of usd mils to 1 hash.