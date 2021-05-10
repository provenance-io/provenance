package io.provenance.kbech32

// From https://github.com/komputing/KBech32
// https://github.com/komputing/KBech32/blob/master/core/common/src/org/komputing/kbech32/Bech32Data.kt

data class Bech32Data(
    val humanReadablePart: String,
    val data: ByteArray
)