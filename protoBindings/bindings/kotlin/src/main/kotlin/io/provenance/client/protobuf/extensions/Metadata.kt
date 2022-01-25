package io.provenance.client.protobuf.extensions

import io.provenance.metadata.v1.ValueOwnershipRequest
import io.provenance.metadata.v1.QueryGrpc.QueryBlockingStub as Metadata

/**
 * Get scope IDs for a given address.
 *
 * See [Metadata](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/Metadata.kt#L11).
 *
 * @param bech32Address The bech32 address to fetch the scope IDs of.
 * @return A list of scope IDs.
 */
fun Metadata.getScopeIds(bech32Address: String): List<String> =
    valueOwnership(ValueOwnershipRequest.newBuilder().setAddress(bech32Address).build())
        .scopeUuidsList
