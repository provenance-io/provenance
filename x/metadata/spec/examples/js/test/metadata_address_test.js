"use strict";
const test = require('ava');
const MetadataAddress = require('../lib/metadata-address');

// IntelliJ can't yet handle ava tests, and will tell you to run them on the command line.
// Additionally, if run from the root of the repo, undesired js files might get attention.
// It's best to run ava from the root of this js example.
// From the root of this repo:
//   $ cd x/metadata/spec/examples/js
//   $ node_modules/.bin/ava
// If that doesn't exist, you might need to:
//   $ npm install

// These strings come from the output of x/metadata/types/address_test.go TestGenerateExamples().

// Pre-selected UUID strings that go with ID strings generated from the Go code.
const SCOPE_UUID = "91978ba2-5f35-459a-86a7-feca1b0512e0";
const SESSION_UUID = "5803f8bc-6067-4eb5-951f-2121671c2ec0";
const SCOPE_SPEC_UUID = "dc83ea70-eacd-40fe-9adf-1cf6148bf8a2";
const CONTRACT_SPEC_UUID = "def6bc0a-c9dd-4874-948f-5206e6060a84";
const RECORD_NAME = "recordname";
const RECORD_NAME_HASHED_BYTES = Uint8Array.of(234, 169, 160, 84, 154, 205, 183, 162, 227, 133, 142, 181, 183, 185, 209, 190);

// Pre-generated ID strings created using Go code and providing the above strings.
const SCOPE_ID = "scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel";
const SESSION_ID = "session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr";
const RECORD_ID = "record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3";
const SCOPE_SPEC_ID = "scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m";
const CONTRACT_SPEC_ID = "contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn";
const RECORD_SPEC_ID = "recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44";

// Copied from metadata-address.js
/**
 * Convert an array of bytes into a UUID string (lowercase).
 * @param bytes the array containing the bytes to convert.
 * @returns A lowercase string in the format "2cd73ed5-54bf-4c5b-a2b1-30860b8fd21e".
 */
function uuidString(bytes) {
    // We just want 16 bytes, so if bytes is longer than that, just get the first 16.
    // If it's shorter, leave the rest of them 0.
    // This isn't really a standard thing, but I'm favoring this over extra validation and errors.
    const uuidBytes = new Uint8Array(16);
    if (bytes != null) {
        if (Number.isInteger(bytes.length)) {
            for (let i = 0; i < bytes.length && i < 16; i++) {
                // converts bytes[i] to a unsigned 8-bit integer.
                // Overflows are wrapped, e.g. -1 becomes 255, and 256 becomes 0, decimals are truncated.
                // Strings are converted to numbers as expected, then the same overflow stuff can happen.
                uuidBytes[i] = bytes[i];
            }
        } else {
            console.log('ignoring bytes argument provided to uuidString because it is not an array or typed array.');
        }
    }

    let retval = "";
    for (let i = 0; i < 16; i++) {
        retval = retval + (uuidBytes[i] + 0x100).toString(16).substr(1);
        if (i === 3 || i === 5 || i === 7 || i === 9) {
            retval = retval + "-";
        }
    }
    return retval.toLocaleLowerCase();
}

test('scopeId', t => {
    let expectedAddr = MetadataAddress.fromBech32(SCOPE_ID);
    let expectedId = SCOPE_ID;
    let expectedKey = MetadataAddress.KEY_SCOPE;
    let expectedPrefix = MetadataAddress.PREFIX_SCOPE;
    let expectedPrimaryUuid = SCOPE_UUID;
    let expectedSecondaryBytes = new Uint8Array(0);

    let actualAddr = MetadataAddress.forScope(SCOPE_UUID);
    let actualId = actualAddr.toString();
    let actualKey = actualAddr.key;
    let actualPrefix = actualAddr.prefix;
    let actualPrimaryUuid = actualAddr.primaryUuid;
    let actualSecondaryBytes = actualAddr.secondaryBytes;

    let addrFromBytes = MetadataAddress.fromBytes(actualAddr.bytes);

    t.deepEqual(expectedKey, actualKey, "key")
    t.deepEqual(expectedPrefix, actualPrefix, "prefix")
    t.deepEqual(expectedPrimaryUuid, actualPrimaryUuid, "primary UUID")
    t.deepEqual(expectedSecondaryBytes, actualSecondaryBytes, "secondary bytes")
    t.deepEqual(expectedId, actualId, "as bech32 string")
    t.assert(expectedAddr.equals(actualAddr), "whole metadata address")
    t.assert(expectedAddr.equals(addrFromBytes), "address from bytes")
});

test('sessionId', t => {
    let expectedAddr = MetadataAddress.fromBech32(SESSION_ID);
    let expectedId = SESSION_ID;
    let expectedKey = MetadataAddress.KEY_SESSION;
    let expectedPrefix = MetadataAddress.PREFIX_SESSION;
    let expectedPrimaryUuid = SCOPE_UUID;
    let expectedSecondaryBytes = SESSION_UUID;

    let actualAddr = MetadataAddress.forSession(SCOPE_UUID, SESSION_UUID);
    let actualId = actualAddr.toString();
    let actualKey = actualAddr.key;
    let actualPrefix = actualAddr.prefix;
    let actualPrimaryUuid = actualAddr.primaryUuid;
    let actualSecondaryBytes = uuidString(actualAddr.secondaryBytes);

    let addrFromBytes = MetadataAddress.fromBytes(actualAddr.bytes);

    t.deepEqual(expectedKey, actualKey, "key")
    t.deepEqual(expectedPrefix, actualPrefix, "prefix")
    t.deepEqual(expectedPrimaryUuid, actualPrimaryUuid, "primary UUID")
    t.deepEqual(expectedSecondaryBytes, actualSecondaryBytes, "secondary bytes")
    t.deepEqual(expectedId, actualId, "as bech32 string")
    t.assert(expectedAddr.equals(actualAddr), "whole metadata address")
    t.assert(expectedAddr.equals(addrFromBytes), "address from bytes")
});

test('recordId', t => {
    let expectedAddr = MetadataAddress.fromBech32(RECORD_ID);
    let expectedId = RECORD_ID;
    let expectedKey = MetadataAddress.KEY_RECORD;
    let expectedPrefix = MetadataAddress.PREFIX_RECORD;
    let expectedPrimaryUuid = SCOPE_UUID;
    let expectedSecondaryBytes = RECORD_NAME_HASHED_BYTES;

    let actualAddr = MetadataAddress.forRecord(SCOPE_UUID, RECORD_NAME);
    let actualId = actualAddr.toString();
    let actualKey = actualAddr.key;
    let actualPrefix = actualAddr.prefix;
    let actualPrimaryUuid = actualAddr.primaryUuid;
    let actualSecondaryBytes = actualAddr.secondaryBytes;

    let addrFromBytes = MetadataAddress.fromBytes(actualAddr.bytes);

    t.deepEqual(expectedKey, actualKey, "key")
    t.deepEqual(expectedPrefix, actualPrefix, "prefix")
    t.deepEqual(expectedPrimaryUuid, actualPrimaryUuid, "primary UUID")
    t.deepEqual(expectedSecondaryBytes, actualSecondaryBytes, "secondary bytes")
    t.deepEqual(expectedId, actualId, "as bech32 string")
    t.assert(expectedAddr.equals(actualAddr), "whole metadata address")
    t.assert(expectedAddr.equals(addrFromBytes), "address from bytes")
});

test('scopeSpecId', t => {
    let expectedAddr = MetadataAddress.fromBech32(SCOPE_SPEC_ID);
    let expectedId = SCOPE_SPEC_ID;
    let expectedKey = MetadataAddress.KEY_SCOPE_SPECIFICATION;
    let expectedPrefix = MetadataAddress.PREFIX_SCOPE_SPECIFICATION;
    let expectedPrimaryUuid = SCOPE_SPEC_UUID;
    let expectedSecondaryBytes = new Uint8Array(0);

    let actualAddr = MetadataAddress.forScopeSpecification(SCOPE_SPEC_UUID);
    let actualId = actualAddr.toString();
    let actualKey = actualAddr.key;
    let actualPrefix = actualAddr.prefix;
    let actualPrimaryUuid = actualAddr.primaryUuid;
    let actualSecondaryBytes = actualAddr.secondaryBytes;

    let addrFromBytes = MetadataAddress.fromBytes(actualAddr.bytes);

    t.deepEqual(expectedKey, actualKey, "key")
    t.deepEqual(expectedPrefix, actualPrefix, "prefix")
    t.deepEqual(expectedPrimaryUuid, actualPrimaryUuid, "primary UUID")
    t.deepEqual(expectedSecondaryBytes, actualSecondaryBytes, "secondary bytes")
    t.deepEqual(expectedId, actualId, "as bech32 string")
    t.assert(expectedAddr.equals(actualAddr), "whole metadata address")
    t.assert(expectedAddr.equals(addrFromBytes), "address from bytes")
});

test('contractSpecId', t => {
    let expectedAddr = MetadataAddress.fromBech32(CONTRACT_SPEC_ID);
    let expectedId = CONTRACT_SPEC_ID;
    let expectedKey = MetadataAddress.KEY_CONTRACT_SPECIFICATION;
    let expectedPrefix = MetadataAddress.PREFIX_CONTRACT_SPECIFICATION;
    let expectedPrimaryUuid = CONTRACT_SPEC_UUID;
    let expectedSecondaryBytes = new Uint8Array(0);

    let actualAddr = MetadataAddress.forContractSpecification(CONTRACT_SPEC_UUID);
    let actualId = actualAddr.toString();
    let actualKey = actualAddr.key;
    let actualPrefix = actualAddr.prefix;
    let actualPrimaryUuid = actualAddr.primaryUuid;
    let actualSecondaryBytes = actualAddr.secondaryBytes;

    let addrFromBytes = MetadataAddress.fromBytes(actualAddr.bytes);

    t.deepEqual(expectedKey, actualKey, "key")
    t.deepEqual(expectedPrefix, actualPrefix, "prefix")
    t.deepEqual(expectedPrimaryUuid, actualPrimaryUuid, "primary UUID")
    t.deepEqual(expectedSecondaryBytes, actualSecondaryBytes, "secondary bytes")
    t.deepEqual(expectedId, actualId, "as bech32 string")
    t.assert(expectedAddr.equals(actualAddr), "whole metadata address")
    t.assert(expectedAddr.equals(addrFromBytes), "address from bytes")
});

test('recordSpecId', t => {
    let expectedAddr = MetadataAddress.fromBech32(RECORD_SPEC_ID);
    let expectedId = RECORD_SPEC_ID;
    let expectedKey = MetadataAddress.KEY_RECORD_SPECIFICATION;
    let expectedPrefix = MetadataAddress.PREFIX_RECORD_SPECIFICATION;
    let expectedPrimaryUuid = CONTRACT_SPEC_UUID;
    let expectedSecondaryBytes = RECORD_NAME_HASHED_BYTES;

    let actualAddr = MetadataAddress.forRecordSpecification(CONTRACT_SPEC_UUID, RECORD_NAME);
    let actualId = actualAddr.toString();
    let actualKey = actualAddr.key;
    let actualPrefix = actualAddr.prefix;
    let actualPrimaryUuid = actualAddr.primaryUuid;
    let actualSecondaryBytes = actualAddr.secondaryBytes;

    let addrFromBytes = MetadataAddress.fromBytes(actualAddr.bytes);

    t.deepEqual(expectedKey, actualKey, "key")
    t.deepEqual(expectedPrefix, actualPrefix, "prefix")
    t.deepEqual(expectedPrimaryUuid, actualPrimaryUuid, "primary UUID")
    t.deepEqual(expectedSecondaryBytes, actualSecondaryBytes, "secondary bytes")
    t.deepEqual(expectedId, actualId, "as bech32 string")
    t.assert(expectedAddr.equals(actualAddr), "whole metadata address")
    t.assert(expectedAddr.equals(addrFromBytes), "address from bytes")
});
