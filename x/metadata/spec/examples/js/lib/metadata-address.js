"use strict";
const bech32 = require('bech32').bech32;
const sha256 = require('crypto-js/sha256')

// A looser UUID regex (than spec) since all we care about are having 16 bytes.
const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

/**
 * Checks if the provided thing is a array or typed array
 * @param thing  the thing to check.
 * @returns true if either an array or typed array, false otherwise.
 */
function isAnArray(thing) {
    if (thing == null) {
        return false;
    }
    // Object.prototype.toString.call(thing) will return something like "[object Uint8Array]".
    // If the last thing ends in "Array" then close enough. Gotta include the ']' too though.
    return Array.isArray(thing) || /Array]$/.test(Object.prototype.toString.call(thing));

}

// Taken and tweaked from https://github.com/uuidjs/uuid/blob/master/src/parse.js
/**
 * Parse a UUID string into an array of bytes.
 * @param uuidStr the UUID string to parse, e.g. "2CD73ED5-54BF-4C5B-A2B1-30860B8FD21E"
 * @returns A Uint8Array with 16 elements.
 */
function parseUuid(uuidStr) {
    if (typeof uuidStr !== 'string' || !uuidRegex.test(uuidStr)) {
        throw 'Invalid uuidStr.';
    }

    const retval = new Uint8Array(16);
    let v;

    // Parse ########-....-....-....-............
    retval[0] = (v = parseInt(uuidStr.slice(0, 8), 16)) >>> 24;
    retval[1] = v >>> 16 & 0xff;
    retval[2] = v >>> 8 & 0xff;
    retval[3] = v & 0xff;

    // Parse ........-####-....-....-............
    retval[4] = (v = parseInt(uuidStr.slice(9, 13), 16)) >>> 8;
    retval[5] = v & 0xff;

    // Parse ........-....-####-....-............
    retval[6] = (v = parseInt(uuidStr.slice(14, 18), 16)) >>> 8;
    retval[7] = v & 0xff;

    // Parse ........-....-....-####-............
    retval[8] = (v = parseInt(uuidStr.slice(19, 23), 16)) >>> 8;
    retval[9] = v & 0xff;

    // Parse ........-....-....-....-############
    // (Use "/" to avoid 32-bit truncation when bit-shifting high-order bytes)
    retval[10] = (v = parseInt(uuidStr.slice(24, 36), 16)) / 0x10000000000 & 0xff;
    retval[11] = v / 0x100000000 & 0xff;
    retval[12] = v >>> 24 & 0xff;
    retval[13] = v >>> 16 & 0xff;
    retval[14] = v >>> 8 & 0xff;
    retval[15] = v & 0xff;
    return retval;
}

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
        if (isAnArray(bytes)) {
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
    return retval.toLowerCase();
}

/**
 * Hashes the provided string and gets the bytes we care about for a MetadataAddress.
 * @param string the string to hash.
 * @return A Uint8Array with 16 elements.
 */
function getHashedBytes(string) {
    let sha256Sum = sha256(string.trim().toLowerCase());
    // A sha256 sum is 32 bytes.
    // That sha256 function returns 8 words that are 4 bytes each.
    // We want the info in 1 byte chunks, though.
    // For MetadataAddress purposes, we also only care about the first 16 bytes.
    // 16 bytes / 4 bytes/word = 4 words and each word has 4 bytes.
    let bytes = [];
    for (let i = 0; i < 4; i++) {
        bytes.push(
            sha256Sum.words[i] >>> 24,
            sha256Sum.words[i] >>> 16 & 0xff,
            sha256Sum.words[i] >>> 8 & 0xff,
            sha256Sum.words[i] & 0xff,
        )
    }
    return Uint8Array.from(bytes);
}

/**
 * Get everything that this MetadataAddress library should export.
 * @return an object meant for module.exports.
 */
function getMetadataAddressLibrary() {
    // The name value of a MetadataAddress object.
    const METADATA_ADDRESS_NAME = "MetadataAddress";

    // Prefix strings for the various types of Metadata Addresses.
    const PREFIX_SCOPE = "scope";
    const PREFIX_SESSION = "session";
    const PREFIX_RECORD = "record";
    const PREFIX_SCOPE_SPECIFICATION = "scopespec";
    const PREFIX_CONTRACT_SPECIFICATION = "contractspec";
    const PREFIX_RECORD_SPECIFICATION = "recspec";

    // Key bytes for the various types of Metadata Addresses.
    const KEY_SCOPE = 0;
    const KEY_SESSION = 1;
    const KEY_RECORD = 2;
    const KEY_SCOPE_SPECIFICATION = 4; // Note that this is not in numerical order.
    const KEY_CONTRACT_SPECIFICATION = 3;
    const KEY_RECORD_SPECIFICATION = 5;

    /**
     * Get the prefix for a key byte.
     * @param key the byte in question.
     * @returns a string prefix, e.g. "scope".
     */
    function getPrefixFromKey(key) {
        let prefix = key === KEY_SCOPE ? PREFIX_SCOPE
                   : key === KEY_SESSION ? PREFIX_SESSION
                   : key === KEY_RECORD ? PREFIX_RECORD
                   : key === KEY_SCOPE_SPECIFICATION ? PREFIX_SCOPE_SPECIFICATION
                   : key === KEY_CONTRACT_SPECIFICATION ? PREFIX_CONTRACT_SPECIFICATION
                   : key === KEY_RECORD_SPECIFICATION ? PREFIX_RECORD_SPECIFICATION
                   : undefined;
        if (prefix === undefined) {
            throw 'Invalid key: [' + key + ']';
        }
        return prefix;
    }

    /**
     * Makes sure the bytes have a valid key and correct length.
     * @param bytes the array of bytes to validate.
     * @returns nothing, but might throw an exception.
     */
    function validateBytes(bytes) {
        if (bytes == null || bytes.length === 0) {
            throw 'Invalid bytes: undefined, null, or empty.';
        }
        let expectedLength = bytes[0] === KEY_SCOPE ? 17
                           : bytes[0] === KEY_SESSION ? 33
                           : bytes[0] === KEY_RECORD ? 33
                           : bytes[0] === KEY_SCOPE_SPECIFICATION ? 17
                           : bytes[0] === KEY_CONTRACT_SPECIFICATION ? 17
                           : bytes[0] === KEY_RECORD_SPECIFICATION ? 33
                           : undefined;
        if (expectedLength === undefined) {
            throw 'Invalid key: [' + key + ']';
        }
        if (expectedLength !== bytes.length) {
            throw 'Incorrect data length for type [' + getPrefixFromKey(bytes[0]) + ']: expected [' + expectedLength + '], actual [' + bytes.length + ']';
        }
    }

    /**
     * Private constructor for a MetadataAddress.
     * @param key the key byte for this MetadataAddress.
     * @param primaryUuid either a UUID string or an array with the 16 bytes of the primary UUID.
     * @param secondary either a string to be hashed or an array of bytes.
     */
    function newMetadataAddress(key, primaryUuid, secondary) {
        if (!Number.isInteger(key) || key < 0 || key > 5) {
            throw 'Invalid key: expected integer between 0 and 5 (inclusive), actual: [' + key + ']';
        }
        if (primaryUuid == null) {
            throw 'Invalid primaryUuid: null or undefined.';
        }
        let primaryUuidBytes = (typeof primaryUuid === "string") ? parseUuid(primaryUuid)
                             : Uint8Array.from(primaryUuid);
        if (primaryUuidBytes.length !== 16) {
            throw 'Invalid primaryUuid: expected byte length [16], actual [' + primaryUuid.length + '].';
        }
        let secondaryBytes = secondary == null ? new Uint8Array(0)
                           : (typeof secondary === "string") ? getHashedBytes(secondary)
                           : Uint8Array.from(secondary);

        // Create the private array of bytes representing this address.
        const bytes = new Uint8Array(17 + secondaryBytes.length);
        bytes[0] = key;
        for (let i = 0; i < 16; i++) {
            bytes[i+1] = primaryUuidBytes[i];
        }
        for (let i = 0; i < secondaryBytes.length; i++) {
            bytes[i+17] = secondaryBytes[i];
        }

        // Pre-compute the bech32 to flush out any final issues (and prevent extra work later).
        let bytesAsBech32 = bech32.encode(getPrefixFromKey(bytes[0]), bech32.toWords(bytes));

        let retval = {
            /** The name of this object: "MetadataAddress". */
            name: METADATA_ADDRESS_NAME,
            /** The key byte (integer) for this MetadataAddress. */
            key: bytes[0],
            /** The prefix string for this MetadataAddress, e.g. "scope". */
            prefix: getPrefixFromKey(bytes[0]),
            /** The lowercase UUID string of the bytes of the primary UUID in this MetadataAddress. */
            primaryUuid: uuidString(bytes.slice(1,17)),
            /** The secondary bytes of this MetadataAddress (may be empty). */
            secondaryBytes: bytes.slice(17),
            /** The bech32 address string of this MetadataAddress. */
            bech32: bytesAsBech32,
            /** Returns the bech32 address string for this MetadataAddress. */
            toString: function() {
                return bytesAsBech32;
            },
            equals: function(other) {
                return other != null && other.name === METADATA_ADDRESS_NAME && bytesAsBech32 === other.toString();
            }
        };

        // Make some of the retval properties read-only and show up during object enumeration.
        ['name', 'key', 'prefix', 'primaryUuid', 'secondaryBytes', 'bech32'].forEach(function(field) {
            Object.defineProperty(retval, field, {
                value: retval[field],
                writable: false,
                enumerable: true
            });
        })

        // Create a getter property for the bytes that always returns a copy of the bytes array.
        // This helps prevent this MetadataAddress from being altered while still providing its information.
        Object.defineProperty(retval, 'bytes', {
            get: function() {
                return bytes.slice(0);
            },
            enumerable: true
        });

        return retval;
    }

    /** Creates a MetadataAddress for a scope. */
    function forScope(scopeUuid) {
        return newMetadataAddress(KEY_SCOPE, scopeUuid);
    }

    /** Creates a MetadataAddress for a session. */
    function forSession(scopeUuid, sessionUuid) {
        if (typeof sessionUuid === 'string') {
            sessionUuid = parseUuid(sessionUuid);
        }
        return newMetadataAddress(KEY_SESSION, scopeUuid, sessionUuid);
    }

    /** Creates a MetadataAddress for a record. */
    function forRecord(scopeUuid, recordName) {
        return newMetadataAddress(KEY_RECORD, scopeUuid, recordName);
    }

    /** Creates a MetadataAddress for a scope specification. */
    function forScopeSpecification(scopeSpecUuid) {
        return newMetadataAddress(KEY_SCOPE_SPECIFICATION, scopeSpecUuid);
    }

    /** Creates a MetadataAddress for a contract specification. */
    function forContractSpecification(contractSpecUuid) {
        return newMetadataAddress(KEY_CONTRACT_SPECIFICATION, contractSpecUuid);
    }

    /** Creates a MetadataAddress for a record specification. */
    function forRecordSpecification(contractSpecUuid, recordSpecName) {
        return newMetadataAddress(KEY_RECORD_SPECIFICATION, contractSpecUuid, recordSpecName);
    }

    /** Creates a MetadataAddress from a bech32 string. */
    function fromBech32(bech32Str) {
        let b32 = bech32.decode(bech32Str);
        let hrp = b32.prefix;
        let bytes = bech32.fromWords(b32.words);
        validateBytes(bytes);
        let prefix = getPrefixFromKey(bytes[0]);
        if (prefix !== hrp) {
            throw 'Incorrect HRP: expected [' + prefix + '], actual [' + hrp + '].';
        }
        return newMetadataAddress(bytes[0], bytes.slice(1,17), bytes.slice(17));
    }

    /** Creates a MetadataAddress from an array of bytes. */
    function fromBytes(bytes) {
        validateBytes(bytes);
        return newMetadataAddress(bytes[0], bytes.slice(1,17), bytes.slice(17));
    }

    return {
        forScope,
        forSession,
        forRecord,
        forScopeSpecification,
        forContractSpecification,
        forRecordSpecification,
        fromBech32,
        fromBytes,

        PREFIX_SCOPE,
        PREFIX_SESSION,
        PREFIX_RECORD,
        PREFIX_SCOPE_SPECIFICATION,
        PREFIX_CONTRACT_SPECIFICATION,
        PREFIX_RECORD_SPECIFICATION,

        KEY_SCOPE,
        KEY_SESSION,
        KEY_RECORD,
        KEY_SCOPE_SPECIFICATION,
        KEY_CONTRACT_SPECIFICATION,
        KEY_RECORD_SPECIFICATION,
    };
}

module.exports = getMetadataAddressLibrary();