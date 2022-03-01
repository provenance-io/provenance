package io.provenance.client.protobuf.extensions

import cosmos.feegrant.v1beta1.Feegrant
import cosmos.feegrant.v1beta1.QueryOuterClass
import cosmos.feegrant.v1beta1.QueryGrpc.QueryBlockingStub as FeeGrants

/**
 * Get the fee grants for (granter, grantee) addresses.
 *
 * See [FeeGrants](https://github.com/FigureTechnologies/service-wallet/blob/v45/pb-client/src/main/kotlin/com/figure/wallet/pbclient/client/grpc/FeeGrants.kt#L11).
 *
 * @param granterAddress The bech32 address of the fee granter.
 * @param granteeAddress The bech32 address of the fee grantee.
 * @return [Feegrant.Grant]
 */
fun FeeGrants.getFeeGrant(granterAddress: String, granteeAddress: String): Feegrant.Grant =
    allowance(
        QueryOuterClass.QueryAllowanceRequest
            .newBuilder()
            .setGranter(granterAddress)
            .setGrantee(granteeAddress)
            .build()
    )
        .allowance
