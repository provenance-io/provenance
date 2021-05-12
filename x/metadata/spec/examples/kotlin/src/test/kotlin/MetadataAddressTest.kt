package io.provenance

import org.junit.jupiter.api.Assertions.assertArrayEquals
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Test
import java.nio.ByteBuffer
import java.util.UUID

class MetadataAddressTest {

    // These strings come from the output of x/metadata/types/address_test.go TestGenerateExamples().

    // Pre-selected UUID strings that go with ID strings generated from the Go code.
    private val scopeUuidString = "91978ba2-5f35-459a-86a7-feca1b0512e0"
    private val sessionUuidString = "5803f8bc-6067-4eb5-951f-2121671c2ec0"
    private val scopeSpecUuidString = "dc83ea70-eacd-40fe-9adf-1cf6148bf8a2"
    private val contractSpecUuidString = "def6bc0a-c9dd-4874-948f-5206e6060a84"
    private val recordName = "recordname"
    private val recordNameHashedBytes = byteArrayOfInts(234, 169, 160, 84, 154, 205, 183, 162, 227, 133, 142, 181, 183, 185, 209, 190)

    // Pre-generated ID strings created using Go code and providing the above strings.
    private val scopeIdString = "scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel"
    private val sessionIdString = "session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr"
    private val recordIdString = "record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3"
    private val scopeSpecIdString = "scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m"
    private val contractSpecIdString = "contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn"
    private val recordSpecIdString = "recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44"

    // UUID versions of the UUID strings.
    private val scopeUuid: UUID = UUID.fromString(scopeUuidString)
    private val sessionUuid: UUID = UUID.fromString(sessionUuidString)
    private val scopeSpecUuid: UUID = UUID.fromString(scopeSpecUuidString)
    private val contractSpecUuid: UUID = UUID.fromString(contractSpecUuidString)

    private fun byteArrayOfInts(vararg ints: Int) = ByteArray(ints.size) { pos -> ints[pos].toByte() }

    // Copy/Pasted out of MetadataAddress.kt so that it can be private in there.
    /** Converts a UUID to a ByteArray. */
    private fun uuidAsByteArray(uuid: UUID): ByteArray {
        val b = ByteBuffer.wrap(ByteArray(16))
        b.putLong(uuid.mostSignificantBits)
        b.putLong(uuid.leastSignificantBits)
        return b.array()
    }

    @Test
    fun scopeId() {
        val expectedAddr = MetadataAddress.fromBech32(scopeIdString)
        val expectedId = scopeIdString
        val expectedKey = KEY_SCOPE
        val expectedPrefix = PREFIX_SCOPE
        val expectedPrimaryUuid = scopeUuid
        val expectedSecondaryBytes = byteArrayOf()

        val actualAddr = MetadataAddress.forScope(scopeUuid)
        val actualId = actualAddr.toString()
        val actualKey = actualAddr.getKey()
        val actualPrefix = actualAddr.getPrefix()
        val actualPrimaryUuid = actualAddr.getPrimaryUuid()
        val actualSecondaryBytes = actualAddr.getSecondaryBytes()

        val addrFromBytes = MetadataAddress.fromBytes(actualAddr.bytes)

        assertEquals(expectedKey, actualKey, "key")
        assertEquals(expectedPrefix, actualPrefix, "prefix")
        assertEquals(expectedPrimaryUuid, actualPrimaryUuid, "primary uuid")
        assertArrayEquals(expectedSecondaryBytes, actualSecondaryBytes, "secondary bytes")
        assertEquals(expectedId, actualId, "as bech32 strings")
        assertEquals(expectedAddr, actualAddr, "whole metadata address")
        assertEquals(expectedAddr, addrFromBytes, "address from bytes")
        assertEquals(expectedAddr.hashCode(), actualAddr.hashCode(), "hash codes")
    }

    @Test
    fun sessionId() {
        val expectedAddr = MetadataAddress.fromBech32(sessionIdString)
        val expectedId = sessionIdString
        val expectedKey = KEY_SESSION
        val expectedPrefix = PREFIX_SESSION
        val expectedPrimaryUuid = scopeUuid
        val expectedSecondaryBytes = uuidAsByteArray(sessionUuid)

        val actualAddr = MetadataAddress.forSession(scopeUuid, sessionUuid)
        val actualId = actualAddr.toString()
        val actualKey = actualAddr.getKey()
        val actualPrefix = actualAddr.getPrefix()
        val actualPrimaryUuid = actualAddr.getPrimaryUuid()
        val actualSecondaryBytes = actualAddr.getSecondaryBytes()

        val addrFromBytes = MetadataAddress.fromBytes(actualAddr.bytes)

        assertEquals(expectedKey, actualKey, "key")
        assertEquals(expectedPrefix, actualPrefix, "prefix")
        assertEquals(expectedPrimaryUuid, actualPrimaryUuid, "primary uuid")
        assertArrayEquals(expectedSecondaryBytes, actualSecondaryBytes, "secondary bytes")
        assertEquals(expectedId, actualId, "as bech32 strings")
        assertEquals(expectedAddr, actualAddr, "whole metadata address")
        assertEquals(expectedAddr, addrFromBytes, "address from bytes")
        assertEquals(expectedAddr.hashCode(), actualAddr.hashCode(), "hash codes")
    }

    @Test
    fun recordId() {
        val expectedAddr = MetadataAddress.fromBech32(recordIdString)
        val expectedId = recordIdString
        val expectedKey = KEY_RECORD
        val expectedPrefix = PREFIX_RECORD
        val expectedPrimaryUuid = scopeUuid
        val expectedSecondaryBytes = recordNameHashedBytes

        val actualAddr = MetadataAddress.forRecord(scopeUuid, recordName)
        val actualId = actualAddr.toString()
        val actualKey = actualAddr.getKey()
        val actualPrefix = actualAddr.getPrefix()
        val actualPrimaryUuid = actualAddr.getPrimaryUuid()
        val actualSecondaryBytes = actualAddr.getSecondaryBytes()

        val addrFromBytes = MetadataAddress.fromBytes(actualAddr.bytes)

        assertEquals(expectedKey, actualKey, "key")
        assertEquals(expectedPrefix, actualPrefix, "prefix")
        assertEquals(expectedPrimaryUuid, actualPrimaryUuid, "primary uuid")
        assertArrayEquals(expectedSecondaryBytes, actualSecondaryBytes, "secondary bytes")
        assertEquals(expectedId, actualId, "as bech32 strings")
        assertEquals(expectedAddr, actualAddr, "whole metadata address")
        assertEquals(expectedAddr, addrFromBytes, "address from bytes")
        assertEquals(expectedAddr.hashCode(), actualAddr.hashCode(), "hash codes")
    }

    @Test
    fun scopeSpecId() {
        val expectedAddr = MetadataAddress.fromBech32(scopeSpecIdString)
        val expectedId = scopeSpecIdString
        val expectedKey = KEY_SCOPE_SPECIFICATION
        val expectedPrefix = PREFIX_SCOPE_SPECIFICATION
        val expectedPrimaryUuid = scopeSpecUuid
        val expectedSecondaryBytes = byteArrayOf()

        val actualAddr = MetadataAddress.forScopeSpecification(scopeSpecUuid)
        val actualId = actualAddr.toString()
        val actualKey = actualAddr.getKey()
        val actualPrefix = actualAddr.getPrefix()
        val actualPrimaryUuid = actualAddr.getPrimaryUuid()
        val actualSecondaryBytes = actualAddr.getSecondaryBytes()

        val addrFromBytes = MetadataAddress.fromBytes(actualAddr.bytes)

        assertEquals(expectedKey, actualKey, "key")
        assertEquals(expectedPrefix, actualPrefix, "prefix")
        assertEquals(expectedPrimaryUuid, actualPrimaryUuid, "primary uuid")
        assertArrayEquals(expectedSecondaryBytes, actualSecondaryBytes, "secondary bytes")
        assertEquals(expectedId, actualId, "as bech32 strings")
        assertEquals(expectedAddr, actualAddr, "whole metadata address")
        assertEquals(expectedAddr, addrFromBytes, "address from bytes")
        assertEquals(expectedAddr.hashCode(), actualAddr.hashCode(), "hash codes")
    }

    @Test
    fun contractSpecId() {
        val expectedAddr = MetadataAddress.fromBech32(contractSpecIdString)
        val expectedId = contractSpecIdString
        val expectedKey = KEY_CONTRACT_SPECIFICATION
        val expectedPrefix = PREFIX_CONTRACT_SPECIFICATION
        val expectedPrimaryUuid = contractSpecUuid
        val expectedSecondaryBytes = byteArrayOf()

        val actualAddr = MetadataAddress.forContractSpecification(contractSpecUuid)
        val actualId = actualAddr.toString()
        val actualKey = actualAddr.getKey()
        val actualPrefix = actualAddr.getPrefix()
        val actualPrimaryUuid = actualAddr.getPrimaryUuid()
        val actualSecondaryBytes = actualAddr.getSecondaryBytes()

        val addrFromBytes = MetadataAddress.fromBytes(actualAddr.bytes)

        assertEquals(expectedKey, actualKey, "key")
        assertEquals(expectedPrefix, actualPrefix, "prefix")
        assertEquals(expectedPrimaryUuid, actualPrimaryUuid, "primary uuid")
        assertArrayEquals(expectedSecondaryBytes, actualSecondaryBytes, "secondary bytes")
        assertEquals(expectedId, actualId, "as bech32 strings")
        assertEquals(expectedAddr, actualAddr, "whole metadata address")
        assertEquals(expectedAddr, addrFromBytes, "address from bytes")
        assertEquals(expectedAddr.hashCode(), actualAddr.hashCode(), "hash codes")
    }

    @Test
    fun recordSpecId() {
        val expectedAddr = MetadataAddress.fromBech32(recordSpecIdString)
        val expectedId = recordSpecIdString
        val expectedKey = KEY_RECORD_SPECIFICATION
        val expectedPrefix = PREFIX_RECORD_SPECIFICATION
        val expectedPrimaryUuid = contractSpecUuid
        val expectedSecondaryBytes = recordNameHashedBytes

        val actualAddr = MetadataAddress.forRecordSpecification(contractSpecUuid, recordName)
        val actualId = actualAddr.toString()
        val actualKey = actualAddr.getKey()
        val actualPrefix = actualAddr.getPrefix()
        val actualPrimaryUuid = actualAddr.getPrimaryUuid()
        val actualSecondaryBytes = actualAddr.getSecondaryBytes()

        val addrFromBytes = MetadataAddress.fromBytes(actualAddr.bytes)

        assertEquals(expectedKey, actualKey, "key")
        assertEquals(expectedPrefix, actualPrefix, "prefix")
        assertEquals(expectedPrimaryUuid, actualPrimaryUuid, "primary uuid")
        assertArrayEquals(expectedSecondaryBytes, actualSecondaryBytes, "secondary bytes")
        assertEquals(expectedId, actualId, "as bech32 strings")
        assertEquals(expectedAddr, actualAddr, "whole metadata address")
        assertEquals(expectedAddr, addrFromBytes, "address from bytes")
        assertEquals(expectedAddr.hashCode(), actualAddr.hashCode(), "hash codes")
    }
}