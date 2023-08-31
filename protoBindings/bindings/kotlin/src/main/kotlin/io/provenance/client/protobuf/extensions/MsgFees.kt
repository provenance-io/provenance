package io.provenance.client.protobuf.extensions

import io.provenance.msgfees.v1.MsgFee
import io.provenance.msgfees.v1.QueryAllMsgFeesRequest
import io.provenance.msgfees.v1.QueryGrpc.QueryBlockingStub as BlockingMsgFees
import io.provenance.msgfees.v1.QueryGrpcKt.QueryCoroutineStub as CoroutineMsgFees

/**
 * Get a coin balance in the account at the supplied address.
 *
 * @return A list of [MsgFee]
 */
fun BlockingMsgFees.getAllMsgFees(): List<MsgFee> =
    queryAllMsgFees(QueryAllMsgFeesRequest.getDefaultInstance()).msgFeesList

/**
 * Get a coin balance in the account at the supplied address.
 *
 * @return A list of [MsgFee]
 */
suspend fun CoroutineMsgFees.getAllMsgFees(): List<MsgFee> =
    queryAllMsgFees(QueryAllMsgFeesRequest.getDefaultInstance()).msgFeesList
