package io.provenance

import io.provenance.kbech32.Bech32
import io.provenance.kbech32.Bech32Data
import java.io.ByteArrayOutputStream
import java.nio.ByteBuffer
import java.security.MessageDigest
import java.util.UUID

const val PREFIX_SCOPE = "scope"
const val PREFIX_SESSION = "session"
const val PREFIX_RECORD = "record"
const val PREFIX_SCOPE_SPECIFICATION = "scopespec"
const val PREFIX_CONTRACT_SPECIFICATION = "contractspec"
const val PREFIX_RECORD_SPECIFICATION = "recspec"

const val KEY_SCOPE: Byte = 0x00
const val KEY_SESSION: Byte = 0x01
const val KEY_RECORD: Byte = 0x02
const val KEY_SCOPE_SPECIFICATION: Byte = 0x04 // Note that this is not in numerical order.
const val KEY_CONTRACT_SPECIFICATION: Byte = 0x03
const val KEY_RECORD_SPECIFICATION: Byte = 0x05

/**
 * This MetadataAddress class helps create ids for the various types objects stored by the metadata module.
 */
data class MetadataAddress internal constructor(val bytes: ByteArray) {
    companion object {
        /** Create a MetadataAddress for a Scope. */
        fun forScope(scopeUuid: UUID) =
            MetadataAddress(byteArrayOf(KEY_SCOPE).plus(uuidAsByteArray(scopeUuid)))

        /** Create a MetadataAddress for a Session. */
        fun forSession(scopeUuid: UUID, sessionUuid: UUID) =
            MetadataAddress(byteArrayOf(KEY_SESSION).plus(uuidAsByteArray(scopeUuid)).plus(uuidAsByteArray(sessionUuid)))

        /** Create a MetadataAddress for a Record. */
        fun forRecord(scopeUuid: UUID, recordName: String): MetadataAddress {
            if (recordName.isBlank()) {
                throw IllegalArgumentException("Invalid recordName: cannot be empty or blank.")
            }
            return MetadataAddress(byteArrayOf(KEY_RECORD).plus(uuidAsByteArray(scopeUuid)).plus(asHashedBytes(recordName)))
        }

        /** Create a MetadataAddress for a Scope Specification. */
        fun forScopeSpecification(scopeSpecUuid: UUID) =
            MetadataAddress(byteArrayOf(KEY_SCOPE_SPECIFICATION).plus(uuidAsByteArray(scopeSpecUuid)))

        /** Create a MetadataAddress for a Contract Specification. */
        fun forContractSpecification(contractSpecUuid: UUID) =
            MetadataAddress(byteArrayOf(KEY_CONTRACT_SPECIFICATION).plus(uuidAsByteArray(contractSpecUuid)))

        /** Create a MetadataAddress for a Record Specification. */
        fun forRecordSpecification(contractSpecUuid: UUID, recordSpecName: String): MetadataAddress {
            if (recordSpecName.isBlank()) {
                throw IllegalArgumentException("Invalid recordSpecName: cannot be empty or blank.")
            }
            return MetadataAddress(byteArrayOf(KEY_RECORD_SPECIFICATION).plus(uuidAsByteArray(contractSpecUuid)).plus(asHashedBytes(recordSpecName)))
        }

        /** Create a MetadataAddress object from a bech32 address representation of a MetadataAddress. */
        fun fromBech32(bech32Value: String): MetadataAddress {
            val (hrp, data) = decodeAndConvert(bech32Value)
            val prefix = getPrefixFromKey(data[0])
            if (hrp != prefix) {
                throw IllegalArgumentException("Incorrect HRP: Expected ${prefix}, Actual: ${hrp}.")
            }
            validateBytes(data)
            return MetadataAddress(data)
        }

        /** Create a MetadataAddress from a ByteArray. */
        fun fromBytes(bytes: ByteArray): MetadataAddress {
            validateBytes(bytes)
            return MetadataAddress(bytes)
        }

        /** Get the prefix that corresponds to the provided key Byte. */
        private fun getPrefixFromKey(key: Byte) =
            when (key) {
                KEY_SCOPE -> PREFIX_SCOPE
                KEY_SESSION -> PREFIX_SESSION
                KEY_RECORD -> PREFIX_RECORD
                KEY_SCOPE_SPECIFICATION -> PREFIX_SCOPE_SPECIFICATION
                KEY_CONTRACT_SPECIFICATION -> PREFIX_CONTRACT_SPECIFICATION
                KEY_RECORD_SPECIFICATION -> PREFIX_RECORD_SPECIFICATION
                else -> {
                    throw IllegalArgumentException("Invalid key: $key")
                }
            }

        /** Checks that the data has a correct key and length. Throws IllegalArgumentException if not. */
        private fun validateBytes(bytes: ByteArray) {
            val expectedLength = when (bytes[0]) {
                KEY_SCOPE -> 17
                KEY_SESSION -> 33
                KEY_RECORD -> 33
                KEY_SCOPE_SPECIFICATION -> 17
                KEY_CONTRACT_SPECIFICATION -> 17
                KEY_RECORD_SPECIFICATION -> 33
                else -> {
                    throw IllegalArgumentException("Invalid key: ${bytes[0]}")
                }
            }
            if (expectedLength != bytes.size) {
                throw IllegalArgumentException("Incorrect data length for type ${getPrefixFromKey(bytes[0])}: Expected ${expectedLength}, Actual: ${bytes.size}.")
            }
        }

        /** Converts a UUID to a ByteArray. */
        private fun uuidAsByteArray(uuid: UUID): ByteArray {
            val b = ByteBuffer.wrap(ByteArray(16))
            b.putLong(uuid.mostSignificantBits)
            b.putLong(uuid.leastSignificantBits)
            return b.array()
        }

        /** Converts a ByteArray to a UUID. */
        private fun byteArrayAsUuid(data: ByteArray): UUID {
            val uuidBytes = ByteArray(16)
            if (data.size >= 16) {
                data.copyInto(uuidBytes, 0, 0, 16)
            } else if (data.isNotEmpty()) {
                data.copyInto(uuidBytes, 0, 0, data.size)
            }
            val bb = ByteBuffer.wrap(uuidBytes)
            val mostSig = bb.long
            val leastSig = bb.long
            return UUID(mostSig, leastSig)
        }

        /** Hashes a string and gets the bytes desired for a MetadataAddress. */
        private fun asHashedBytes(str: String) =
            MessageDigest.getInstance("SHA-256").digest(str.trim().toLowerCase().toByteArray()).copyOfRange(0, 16)

        /**
         * Converts from a base64 encoded ByteArray to base32 encoded ByteArray and then to bech32 address string.
         */
        private fun convertAndEncode(humanReadablePart: String, dataBase64: ByteArray) =
            Bech32.encode(humanReadablePart, convertBits(dataBase64, 8, 5, true))

        /**
         * Decodes a bech32 encoded string and converts to base64 encoded bytes.
         * Note: Even though this returns an object named "Bech32Data" the data is actual base64 encoded.
         */
        private fun decodeAndConvert(bech32: String): Bech32Data {
            val (hrp, dataBase32) = Bech32.decode(bech32)
            val dataBase64 = convertBits(dataBase32, 5, 8, false)
            return Bech32Data(hrp, dataBase64)
        }

        // convertBits Taken from [Bitcoinj SegwitAddress Java implementation](https://github.com/bitcoinj/bitcoinj/blob/31c7e5fbceb9884cb02d2dabc755009caa2d613e/core/src/main/java/org/bitcoinj/core/SegwitAddress.java)
        /**
         * converts a ByteArray where each byte is encoding fromBits bits,
         * to a ByteArray where each byte is encoding toBits bits.
         */
        private fun convertBits(data: ByteArray, fromBits: Int, toBits: Int, pad: Boolean): ByteArray {
            if (fromBits < 1 || fromBits > 8 || toBits < 1 || toBits > 8) {
                throw IllegalArgumentException("only bit groups between 1 and 8 allowed")
            }

            var acc = 0
            var bits = 0
            val out = ByteArrayOutputStream(64)
            val maxv = (1 shl toBits) - 1
            val maxAcc = (1 shl (fromBits + toBits - 1)) - 1

            for (b in data) {
                val value = b.toInt() and 0xff
                if ((value ushr fromBits) != 0) {
                    throw IllegalArgumentException(String.format("Input value '%X' exceeds '%d' bit size", value, fromBits))
                }
                acc = ((acc shl fromBits) or value) and maxAcc
                bits += fromBits
                while (bits >= toBits) {
                    bits -= toBits
                    out.write((acc ushr bits) and maxv)
                }
            }
            if (pad) {
                if (bits > 0) {
                    out.write((acc shl (toBits - bits)) and maxv)
                }
            } else if (bits >= fromBits || ((acc shl (toBits - bits)) and maxv) != 0) {
                throw IllegalArgumentException("Could not convert bits, invalid padding")
            }
            return out.toByteArray()
        }
    }

    /** Gets the key byte for this MetadataAddress. */
    fun getKey() = this.bytes[0]

    /** Gets the prefix string for this MetadataAddress, e.g. "scope". */
    fun getPrefix() = getPrefixFromKey(this.bytes[0])

    /** Gets the set of bytes for the primary uuid part of this MetadataAddress as a UUID. */
    fun getPrimaryUuid() = byteArrayAsUuid(this.bytes.copyOfRange(1,17))

    /** Gets the set of bytes for the secondary part of this MetadataAddress. */
    fun getSecondaryBytes() = if (this.bytes.size <= 17) byteArrayOf() else bytes.copyOfRange(17, this.bytes.size)

    /** returns this MetadataAddress as a bech32 address string, e.g. "scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel" */
    override fun toString() = convertAndEncode(getPrefixFromKey(this.bytes[0]), this.bytes)

    /** hashCode implementation for a MetadataAddress. */
    override fun hashCode() = this.bytes.contentHashCode()

    /** equals implementation for a MetadataAddress. */
    override fun equals(other: Any?): Boolean {
        if (this === other) {
            return true
        }
        if (other !is MetadataAddress) {
            return false
        }
        return this.bytes.contentEquals(other.bytes)
    }
}