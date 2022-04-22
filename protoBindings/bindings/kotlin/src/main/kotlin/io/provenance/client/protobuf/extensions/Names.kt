package io.provenance.client.protobuf.extensions

import io.provenance.client.protobuf.paginationBuilder
import io.provenance.name.v1.QueryResolveRequest
import io.provenance.name.v1.QueryReverseLookupRequest
import io.provenance.name.v1.QueryGrpc.QueryBlockingStub as Names

/**
 * Resolve a name to an address.
 *
 * See [Names](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Names.kt#L13).
 *
 * @param name The name to look up.
 * @return The bech32 address of the name.
 */
fun Names.resolveAddressForName(name: String): String =
    resolve(QueryResolveRequest.newBuilder().setName(name).build()).address

/**
 * Get all names associated with the given address.
 *
 * See [Names](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Names.kt#L16).
 *
 * @param address The bech32 address to list the names of.
 * @return The names associated with the address.
 */
fun Names.getAllNames(address: String): List<String> {
    var offset = 0
    var total: Long
    val limit = 100
    val names = mutableListOf<String>()

    do {
        val results =
            reverseLookup(
                QueryReverseLookupRequest.newBuilder()
                    .setAddress(address)
                    .setPagination(paginationBuilder(offset, limit))
                    .build()
            )
        total = results.pagination?.total ?: results.nameCount.toLong()
        offset += limit
        names.addAll(results.nameList)
    } while (names.count() < total)

    return names
}
