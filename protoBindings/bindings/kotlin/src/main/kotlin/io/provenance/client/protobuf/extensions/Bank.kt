package io.provenance.client.protobuf.extensions

import cosmos.bank.v1beta1.QueryOuterClass
import cosmos.base.v1beta1.CoinOuterClass
import cosmos.bank.v1beta1.QueryGrpc.QueryBlockingStub as Bank

/**
 * Get a list of coins in the account at the supplied address.
 *
 * See [Accounts](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Accounts.kt#L34).
 *
 * @param bech32Address The bech32 address to query.
 * @return A list of [CoinOuterClass.Coin] associated with the account.
 */
fun Bank.getAccountCoins(bech32Address: String): List<CoinOuterClass.Coin> =
    allBalances(QueryOuterClass.QueryAllBalancesRequest.newBuilder().setAddress(bech32Address).build()).balancesList

/**
 * Queries the supply of a single coin.
 *
 * See [Accounts](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Accounts.kt#L40).
 *
 * @param denom The denomination of the coin.
 * @return [CoinOuterClass.Coin]
 */
fun Bank.getSupply(denom: String): CoinOuterClass.Coin =
    supplyOf(
        QueryOuterClass.QuerySupplyOfRequest.newBuilder()
            .setDenom(denom)
            .build()
    ).amount
