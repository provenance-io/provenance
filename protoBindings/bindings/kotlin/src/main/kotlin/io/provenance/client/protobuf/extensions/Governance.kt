package io.provenance.client.protobuf.extensions

import cosmos.gov.v1beta1.Gov
import cosmos.gov.v1beta1.QueryGrpc.QueryBlockingStub as Governance

/**
 * Get a coin balance in the account at the supplied address.
 *
 * @return A list of [Gov.Proposal]
 */
fun Governance.getAllProposals(): List<Gov.Proposal> =
    proposals(cosmos.gov.v1beta1.QueryOuterClass.QueryProposalsRequest.getDefaultInstance()).proposalsList
