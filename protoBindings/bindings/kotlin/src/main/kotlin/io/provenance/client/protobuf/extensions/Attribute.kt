package io.provenance.client.protobuf.extensions

import io.provenance.attribute.v1.QueryAttributesRequest
import io.provenance.client.protobuf.paginationBuilder
import java.util.concurrent.TimeUnit
import io.provenance.attribute.v1.QueryGrpc.QueryBlockingStub as Attribute

/**
 * Given an address, fetch all [io.provenance.attribute.v1.Attribute] instances associated with it.
 *
 * See [Attributes](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Attributes.kt#L14).
 *
 * @param address
 * @param deadlineInSeconds
 * @return A list of [io.provenance.attribute.v1.Attribute]
 */
fun Attribute.getAllAttributes(
    address: String,
    deadlineInSeconds: Long? = null
): List<io.provenance.attribute.v1.Attribute> {
    var offset = 0
    var total: Long
    val limit = 100
    val attributes = mutableListOf<io.provenance.attribute.v1.Attribute>()

    do {
        val request = QueryAttributesRequest.newBuilder()
            .setAccount(address)
            .setPagination(paginationBuilder(offset, limit))
            .build()

        val client = if (deadlineInSeconds == null) {
            this
        } else {
            this.withDeadlineAfter(deadlineInSeconds, TimeUnit.SECONDS)
        }

        val results = client.attributes(request)

        total = results.pagination?.total ?: results.attributesCount.toLong()
        offset += limit
        attributes.addAll(results.attributesList)
    } while (attributes.count() < total)

    return attributes
}
