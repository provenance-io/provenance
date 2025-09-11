package io.provenance.client.protobuf.extensions

import io.provenance.flatfees.v1.MsgFee
import io.provenance.flatfees.v1.QueryAllMsgFeesRequest
import io.provenance.flatfees.v1.QueryGrpc.QueryBlockingStub as BlockingMsgFees
import io.provenance.flatfees.v1.QueryGrpcKt.QueryCoroutineStub as CoroutineMsgFees

/**
 * Get a coin balance in the account at the supplied address.
 *
 * @return A list of [MsgFee]
 */
fun BlockingMsgFees.getAllMsgFees(): List<MsgFee> =
    allMsgFees(QueryAllMsgFeesRequest.getDefaultInstance()).msgFeesList

/**
 * Get a coin balance in the account at the supplied address.
 *
 * @return A list of [MsgFee]
 */
suspend fun CoroutineMsgFees.getAllMsgFees(): List<MsgFee> =
    allMsgFees(QueryAllMsgFeesRequest.getDefaultInstance()).msgFeesList
