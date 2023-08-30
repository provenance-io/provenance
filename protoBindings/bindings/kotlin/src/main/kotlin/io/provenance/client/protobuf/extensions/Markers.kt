package io.provenance.client.protobuf.extensions

import cosmos.bank.v1beta1.Bank
import cosmos.base.v1beta1.CoinOuterClass
import io.provenance.client.protobuf.paginationBuilder
import io.provenance.marker.v1.MarkerAccount
import io.provenance.marker.v1.QueryDenomMetadataRequest
import io.provenance.marker.v1.QueryEscrowRequest
import io.provenance.marker.v1.QueryHoldingRequest
import io.provenance.marker.v1.QueryHoldingResponse
import io.provenance.marker.v1.QueryMarkerRequest
import io.provenance.marker.v1.QueryGrpc.QueryBlockingStub as BlockingMarkers
import io.provenance.marker.v1.QueryGrpcKt.QueryCoroutineStub as CoroutineMarkers

/**
 * Get a marker account by ID.
 *
 * See [Markers](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Markers.kt#L17).
 *
 * @param id The identifier of the marker account.
 * @return [MarkerAccount]
 */
fun BlockingMarkers.getMarkerAccount(id: String): MarkerAccount =
    marker(QueryMarkerRequest.newBuilder().setId(id).build()).marker.run {
        when {
            this.`is`(MarkerAccount::class.java) -> unpack(MarkerAccount::class.java)
            else -> throw IllegalArgumentException("Marker type not handled:$typeUrl")
        }
    }

/**
 * Get a marker account by ID.
 *
 * See [Markers](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Markers.kt#L17).
 *
 * @param id The identifier of the marker account.
 * @return [MarkerAccount]
 */
suspend fun CoroutineMarkers.getMarkerAccount(id: String): MarkerAccount =
    marker(QueryMarkerRequest.newBuilder().setId(id).build()).marker.run {
        when {
            this.`is`(MarkerAccount::class.java) -> unpack(MarkerAccount::class.java)
            else -> throw IllegalArgumentException("Marker type not handled:$typeUrl")
        }
    }

/**
 * Get the metadata associated with a given marker.
 *
 * See [Markers](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Markers.kt#L25).
 *
 * @param denom The denomination.
 * @return [Bank.Metadata]
 */
fun BlockingMarkers.getMarkerMetadata(denom: String): Bank.Metadata =
    denomMetadata(QueryDenomMetadataRequest.newBuilder().setDenom(denom).build()).metadata

/**
 * Get the metadata associated with a given marker.
 *
 * See [Markers](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Markers.kt#L25).
 *
 * @param denom The denomination.
 * @return [Bank.Metadata]
 */
suspend fun CoroutineMarkers.getMarkerMetadata(denom: String): Bank.Metadata =
    denomMetadata(QueryDenomMetadataRequest.newBuilder().setDenom(denom).build()).metadata

/**
 * List all accounts holding the given marker.
 *
 * See [Markers](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Markers.kt#L31).
 *
 * @param denom The denomination.
 * @param offset
 * @param limit
 * @return [QueryHoldingResponse]
 */
fun BlockingMarkers.getMarkerHolders(denom: String, offset: Int = 0, limit: Int = 200): QueryHoldingResponse =
    holding(
        QueryHoldingRequest.newBuilder()
            .setId(denom)
            .setPagination(paginationBuilder(offset, limit))
            .build()
    )

/**
 * List all accounts holding the given marker.
 *
 * See [Markers](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Markers.kt#L31).
 *
 * @param denom The denomination.
 * @param offset
 * @param limit
 * @return [QueryHoldingResponse]
 */
suspend fun CoroutineMarkers.getMarkerHolders(denom: String, offset: Int = 0, limit: Int = 200): QueryHoldingResponse =
    holding(
        QueryHoldingRequest.newBuilder()
            .setId(denom)
            .setPagination(paginationBuilder(offset, limit))
            .build()
    )

/**
 * TODO: Description for getMarkerEscrow
 *
 * @param id The identifier of the marker.
 * @param escrowDenom The escrow denomination.
 * @return [CoinOuterClass.Coin?]
 */
fun BlockingMarkers.getMarkerEscrow(id: String, escrowDenom: String): CoinOuterClass.Coin? =
    escrow(
        QueryEscrowRequest
            .newBuilder()
            .setId(id)
            .build()
    ).escrowList
        .firstOrNull { it.denom == escrowDenom }

/**
 * TODO: Description for getMarkerEscrow
 *
 * @param id The identifier of the marker.
 * @param escrowDenom The escrow denomination.
 * @return [CoinOuterClass.Coin?]
 */
suspend fun CoroutineMarkers.getMarkerEscrow(id: String, escrowDenom: String): CoinOuterClass.Coin? =
    escrow(
        QueryEscrowRequest
            .newBuilder()
            .setId(id)
            .build()
    ).escrowList
        .firstOrNull { it.denom == escrowDenom }
