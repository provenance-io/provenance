package io.provenance.client.protobuf.extensions

import cosmos.bank.v1beta1.QueryOuterClass
import cosmos.base.v1beta1.CoinOuterClass
import cosmos.bank.v1beta1.QueryGrpc.QueryBlockingStub as BlockingBank
import cosmos.bank.v1beta1.QueryGrpcKt.QueryCoroutineStub as CoroutineBank

/**
 * Get a coin balance in the account at the supplied address.
 *
 * See [Accounts](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Accounts.kt#L34).
 *
 * @param bech32Address The bech32 address to query.
 * @param denom The denomination of the coin.
 * @return [CoinOuterClass.Coin]
 */
fun BlockingBank.getAccountBalance(bech32Address: String, denom: String): CoinOuterClass.Coin =
    balance(QueryOuterClass.QueryBalanceRequest.newBuilder().setAddress(bech32Address).setDenom(denom).build()).balance

/**
 * Get a coin balance in the account at the supplied address.
 *
 * See [Accounts](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Accounts.kt#L34).
 *
 * @param bech32Address The bech32 address to query.
 * @param denom The denomination of the coin.
 * @return [CoinOuterClass.Coin]
 */
suspend fun CoroutineBank.getAccountBalance(bech32Address: String, denom: String): CoinOuterClass.Coin =
    balance(QueryOuterClass.QueryBalanceRequest.newBuilder().setAddress(bech32Address).setDenom(denom).build()).balance

/**
 * Get a list of coin balances in the account at the supplied address.
 *
 * See [Accounts](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Accounts.kt#L34).
 *
 * @param bech32Address The bech32 address to query.
 * @return A list of [CoinOuterClass.Coin] associated with the account.
 */
fun BlockingBank.getAccountCoins(bech32Address: String): List<CoinOuterClass.Coin> =
    allBalances(QueryOuterClass.QueryAllBalancesRequest.newBuilder().setAddress(bech32Address).build()).balancesList

/**
 * Get a list of coin balances in the account at the supplied address.
 *
 * See [Accounts](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Accounts.kt#L34).
 *
 * @param bech32Address The bech32 address to query.
 * @return A list of [CoinOuterClass.Coin] associated with the account.
 */
suspend fun CoroutineBank.getAccountCoins(bech32Address: String): List<CoinOuterClass.Coin> =
    allBalances(QueryOuterClass.QueryAllBalancesRequest.newBuilder().setAddress(bech32Address).build()).balancesList

/**
 * Queries the supply of a single coin.
 *
 * See [Accounts](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Accounts.kt#L40).
 *
 * @param denom The denomination of the coin.
 * @return [CoinOuterClass.Coin]
 */
fun BlockingBank.getSupply(denom: String): CoinOuterClass.Coin =
    supplyOf(
        QueryOuterClass.QuerySupplyOfRequest.newBuilder()
            .setDenom(denom)
            .build()
    ).amount

/**
 * Queries the supply of a single coin.
 *
 * See [Accounts](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Accounts.kt#L40).
 *
 * @param denom The denomination of the coin.
 * @return [CoinOuterClass.Coin]
 */
suspend fun CoroutineBank.getSupply(denom: String): CoinOuterClass.Coin =
    supplyOf(
        QueryOuterClass.QuerySupplyOfRequest.newBuilder()
            .setDenom(denom)
            .build()
    ).amount
