package io.provenance

import java.io.ByteArrayOutputStream

/** Data involved with a Bech32 address */
data class Bech32Data(val hrp: String, val data: ByteArray) {

    /**
     * The encapsulated data returned as a Hexadecimal string
     */
    val hexData = this.data.joinToString("") { "%02x".format(it) }

    /**
     * The Bech32 encoded value of the data prefixed with the human readable portion and
     * protected by an appended checksum.
     */
    val address = Bech32.encode(hrp, data)

    /**
     * The Bech32 Address toString prints state information for debugging purposes.
     * @see address() for the bech32 encoded address string output.
     */
    override fun toString(): String {
        return "bech32 : ${this.address}\nhuman: ${this.hrp} \nbytes: ${this.hexData}"
        /*
        bech32 : provenance1gx58vp8pryh3jkvxnkvzmd0hqmqqnyqxrtvheq
        human: provenance
        bytes: 41A87604E1192F1959869D982DB5F706C0099006
         */
    }

    /** equals implementation for a Bech32Data object. */
    override fun equals(other: Any?): Boolean {
        if (this === other) return true
        if (javaClass != other?.javaClass) return false
        other as Bech32Data
        return this.hrp == other.hrp &&
                this.data.contentEquals(other.data)
    }

    /** equals implementation for a Bech32Data object. */
    override fun hashCode(): Int {
        var result = hrp.hashCode()
        result = 31 * result + this.data.contentHashCode()
        return result
    }
}

/** BIP173 compliant processing functions for handling Bech32 encoding and decoding. */
class Bech32 {
    companion object {
        private const val CHECKSUM_SIZE = 6
        private const val MIN_VALID_LENGTH = 8
        private const val MAX_VALID_LENGTH = 90
        private const val MIN_VALID_CODEPOINT = 33
        private const val MAX_VALID_CODEPOINT = 126

        private const val charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"
        private val gen = intArrayOf(0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3)

        /** Decodes a Bech32 String */
        fun decode(bech32: String): Bech32Data {
            require(bech32.length in MIN_VALID_LENGTH..MAX_VALID_LENGTH) { "invalid bech32 string length" }
            require(bech32.toCharArray().none { c -> c.toInt() < MIN_VALID_CODEPOINT || c.toInt() > MAX_VALID_CODEPOINT })
            { "invalid character in bech32: ${bech32.toCharArray().map { c -> c.toInt() }
                .filter { c -> c < MIN_VALID_CODEPOINT || c > MAX_VALID_CODEPOINT }}" }

            require(bech32 == bech32.toLowerCase() || bech32 == bech32.toUpperCase())
            { "bech32 must be either all upper or lower case" }
            require(bech32.substring(1).dropLast(CHECKSUM_SIZE).contains('1')) { "invalid index of '1'" }

            val hrp = bech32.substringBeforeLast('1').toLowerCase()
            val dataString = bech32.substringAfterLast('1').toLowerCase()

            require(dataString.toCharArray().all { c -> charset.contains(c) }) { "invalid data encoding character in bech32"}

            val dataBytes = dataString.map { c -> charset.indexOf(c).toByte() }.toByteArray()
            val checkBytes = dataString.takeLast(CHECKSUM_SIZE).map { c -> charset.indexOf(c).toByte() }.toByteArray()

            val actualSum = checksum(hrp, dataBytes.dropLast(CHECKSUM_SIZE).toTypedArray())
            require(1 == polymod(expandHrp(hrp).plus(dataBytes.map { d -> d.toInt() }))) { "checksum failed: $checkBytes != $actualSum" }

            return Bech32Data(hrp, convertBits(dataBytes.dropLast(CHECKSUM_SIZE).toByteArray(), 5, 8, false))
        }

        /**
         * Encodes the provided hrp and data to a Bech32 address string.
         * @param hrp the human readable portion (prefix) to use.
         * @param eightBitData an array of 8-bit encoded bytes.
         */
        fun encode(hrp: String, eightBitData: ByteArray) =
            encodeFiveBitData(hrp, convertBits(eightBitData, 8, 5, true))

        /** Encodes 5-bit bytes (fiveBitData) with a given human readable portion (hrp) into a bech32 string. */
        private fun encodeFiveBitData(hrp: String, fiveBitData: ByteArray): String {
            return (fiveBitData.plus(checksum(hrp, fiveBitData.toTypedArray()))
                .map { b -> charset[b.toInt()] }).joinToString("", hrp + "1")
        }

        /**
         * ConvertBits regroups bytes with toBits set based on reading groups of bits as a continuous stream group by fromBits.
         * This process is used to convert from base64 (from 8) to base32 (to 5) or the inverse.
         */
        private fun convertBits(data: ByteArray, fromBits: Int, toBits: Int, pad: Boolean): ByteArray {
            require (fromBits in 1..8 && toBits in 1..8) { "only bit groups between 1 and 8 are supported"}

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

        /** Calculates a bech32 checksum based on BIP 173 specification */
        private fun checksum(hrp: String, data: Array<Byte>): ByteArray {
            val values = expandHrp(hrp)
                .plus(data.map { d -> d.toInt() })
                .plus(Array(6){ 0 }.toIntArray())

            val poly = polymod(values) xor 1

            return (0..5).map {
                ((poly shr (5 * (5-it))) and 31).toByte()
            }.toByteArray()
        }

        /** Expands the human readable prefix per BIP173 for Checksum encoding */
        private fun expandHrp(hrp: String) =
            hrp.map { c -> c.toInt() shr 5 }
                .plus(0)
                .plus(hrp.map { c -> c.toInt() and 31 })
                .toIntArray()

        /** Polynomial division function for checksum calculation.  For details see BIP173 */
        private fun polymod(values: IntArray): Int {
            var chk = 1
            return values.map { v ->
                val b = chk shr 25
                chk = ((chk and 0x1ffffff) shl 5) xor v
                (0..4).map {
                    if (((b shr it) and 1) == 1) {
                        chk = chk xor gen[it]
                    }
                }
            }.let { chk }
        }
    }
}