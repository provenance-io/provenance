package io.provenance.client.protobuf.extensions

import io.provenance.msgfees.v1.MsgFee
import io.provenance.msgfees.v1.QueryAllMsgFeesRequest
import io.provenance.msgfees.v1.QueryGrpc.QueryBlockingStub as MsgFees

/**
 * Get a coin balance in the account at the supplied address.
 *
 * @return A list of [MsgFee]
 */
fun MsgFees.getAllMsgFees(): List<MsgFee> =
    queryAllMsgFees(QueryAllMsgFeesRequest.getDefaultInstance()).msgFeesList
