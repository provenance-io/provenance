package io.provenance.client.protobuf.extensions

import com.google.protobuf.Any
import com.google.protobuf.GeneratedMessageV3
import com.google.protobuf.Message
import cosmos.tx.v1beta1.TxOuterClass

/*
 * Check if a protobuf field is set
 */
fun GeneratedMessageV3?.isSet() = when {
    this != null && this != this.defaultInstanceForType -> true
    else -> false
}

fun GeneratedMessageV3.isNotSet() = !this.isSet()

fun <T : GeneratedMessageV3> T.whenSet(block: (T) -> Unit) =
    this.takeIf { it.isSet() }?.also { block(it) }

fun <T : GeneratedMessageV3> T.whenNotSet(block: (T) -> Unit) =
    this.takeIf { it.isNotSet() }?.also { block(it) }

fun <T : GeneratedMessageV3, K> T.whenSetLet(block: (T) -> K) =
    this.takeIf { it.isSet() }?.let { block(it) }

fun <T : GeneratedMessageV3, K> T.whenNotSetLet(block: (T) -> K) =
    this.takeIf { it.isNotSet() }?.let { block(it) }

fun Message.toAny(typeUrlPrefix: String = ""): Any = Any.pack(this, typeUrlPrefix)

fun Iterable<Any>.toTxBody(memo: String? = null, timeoutHeight: Long? = null): TxOuterClass.TxBody =
    TxOuterClass.TxBody.newBuilder()
        .addAllMessages(this)
        .also { builder ->
            memo?.run { builder.memo = this }
            timeoutHeight?.run { builder.timeoutHeight = this }
        }
        .build()

fun Any.toTxBody(memo: String? = null, timeoutHeight: Long? = null): TxOuterClass.TxBody =
    listOf(this).toTxBody(memo, timeoutHeight)
