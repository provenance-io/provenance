package io.provenance.client.protobuf.extensions

import cosmos.tx.v1beta1.ServiceOuterClass
import cosmos.tx.v1beta1.ServiceGrpc.ServiceBlockingStub as Transactions

/**
 * Get a transaction by its hash.
 *
 * See [Transactions](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Transactions.kt#L13).
 *
 * @param hash The hash of the transaction to look up.
 * @return [ServiceOuterClass.GetTxResponse].
 */
fun Transactions.getTx(hash: String): ServiceOuterClass.GetTxResponse =
    getTx(ServiceOuterClass.GetTxRequest.newBuilder().setHash(hash).build())

/**
 * Get transactions sent by address.
 *
 * See [Transactions](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Transactions.kt#L15).
 *
 * @param address The source bech32 address.
 * @return [ServiceOuterClass.GetTxsEventResponse].
 */
fun Transactions.getSentTxsByAddress(address: String): ServiceOuterClass.GetTxsEventResponse =
    getTxsEvent(
        ServiceOuterClass.GetTxsEventRequest.newBuilder()
            .addEvents("message.sender='$address'")
            .build()
    )

/**
 * Get transactions received by address.
 *
 * See [Transactions](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Transactions.kt#L21).
 *
 * @param address The destination bech32 address.
 * @return [ServiceOuterClass.GetTxsEventResponse].
 */
fun Transactions.getReceivedTxsByAddress(address: String): ServiceOuterClass.GetTxsEventResponse =
    getTxsEvent(
        ServiceOuterClass.GetTxsEventRequest.newBuilder()
            .addEvents("transfer.recipient='$address'")
            .build()
    )
