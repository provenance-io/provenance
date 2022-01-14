package io.provenance.client.protobuf.extensions.time

import com.google.protobuf.Timestamp
import com.google.protobuf.TimestampOrBuilder
import com.google.protobuf.util.Timestamps
import java.time.Instant
import java.time.OffsetDateTime
import java.time.ZoneId
import java.time.ZonedDateTime

// Extracted from https://github.com/FigureTechnologies/stream-data
// https://github.com/FigureTechnologies/stream-data/blob/master/src/main/kotlin/com/figure/proto/extensions/Extensions.kt

/**
 * Store Instant as Timestamp (UTC)
 */
fun Timestamp.Builder.setValue(instant: Instant): Timestamp.Builder {
    val max = Timestamps.MAX_VALUE
    val min = Timestamps.MIN_VALUE
    when {
        instant.epochSecond > max.seconds -> this.seconds = max.seconds
        instant.epochSecond < min.seconds -> this.seconds = min.seconds
        else -> this.seconds = instant.epochSecond
    }

    when {
        instant.nano > max.nanos -> this.nanos = max.nanos
        instant.nano < min.nanos -> this.nanos = min.nanos
        else -> this.nanos = instant.nano
    }

    return this
}

private fun TimestampOrBuilder.bound(): TimestampOrBuilder {
    val max = Timestamps.MAX_VALUE
    val min = Timestamps.MIN_VALUE
    val new = Timestamp.newBuilder()
    when {
        this.seconds > max.seconds -> new.seconds = max.seconds
        this.seconds < min.seconds -> new.seconds = min.seconds
        else -> new.seconds = this.seconds
    }

    when {
        this.nanos > max.nanos -> new.nanos = max.nanos
        this.nanos < min.nanos -> new.nanos = min.nanos
        else -> new.nanos = this.nanos
    }

    return new.build()
}

/**
 * Get Timestamp as OffsetDateTime (system time zone) if can; otherwise, return null
 */
fun TimestampOrBuilder.toOffsetDateTimeOrNull(): OffsetDateTime? = try {
    toOffsetDateTime(ZoneId.systemDefault())
} catch (t: Throwable) {
    null
}

/**
 * Get Timestamp as OffsetDateTime (system time zone)
 */
fun TimestampOrBuilder.toOffsetDateTime(): OffsetDateTime = toOffsetDateTime(ZoneId.systemDefault())

/**
 * Get Timestamp as OffsetDateTime
 */
fun TimestampOrBuilder.toOffsetDateTimeOrNull(zoneId: ZoneId): OffsetDateTime? = try {
    OffsetDateTime.ofInstant(bound().toInstant(), zoneId)
} catch (t: Throwable) {
    null
}

/**
 * Get Timestamp as OffsetDateTime
 */
fun TimestampOrBuilder.toOffsetDateTime(zoneId: ZoneId) = OffsetDateTime.ofInstant(bound().toInstant(), zoneId)

/**
 * Get Timestamp as ZonedDateTime (system time zone)
 */
fun TimestampOrBuilder.toZonedDateTime(): ZonedDateTime = toZonedDateTime(ZoneId.systemDefault())

/**
 * Get Timestamp as ZonedDateTime
 */
fun TimestampOrBuilder.toZonedDateTime(zoneId: ZoneId): ZonedDateTime = ZonedDateTime.ofInstant(toInstant(), zoneId)

/**
 * Get Timestamp as Instant
 */
fun TimestampOrBuilder.toInstant(): Instant = Instant.ofEpochSecond(seconds, nanos.toLong())

/**
 * Quick convert OffsetDateTime to Timestamp
 */
fun OffsetDateTime.toProtoTimestamp(): Timestamp = Timestamp.newBuilder().setValue(this).build()

/**
 * Quick convert ZonedDateTime to Timestamp
 */
fun ZonedDateTime.toProtoTimestamp(): Timestamp = Timestamp.newBuilder().setValue(this).build()

/**
 * Store OffsetDateTime as Timestamp (UTC)
 */
fun Timestamp.Builder.setValue(odt: OffsetDateTime): Timestamp.Builder = setValue(odt.toInstant())

/**
 * Store ZonedDateTime as Timestamp (UTC)
 */
fun Timestamp.Builder.setValue(odt: ZonedDateTime): Timestamp.Builder = setValue(odt.toInstant())
