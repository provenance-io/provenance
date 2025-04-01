
<!--
order: 6
-->

# Parameters

The MsgFee module contains the following parameter:

| Key                    | Type     | Example                           |
|------------------------|----------|-----------------------------------|
| FloorGasPrice          | `uint32` | `"1905"`                          |
| NhashPerUsdMil         | `uint64` | `"14285714"`                      |



FloorGasPrice is the value of base denom that is charged for calculating base fees, for when base fee and additional fee are charged in the base denom.

NhashPerUsdMil is the number of nhash per usd mil 