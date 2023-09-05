package keeper

// This file is in the keeper package (not keeper_test) so that it can expose
// some private keeper stuff for unit testing.

// ParseLengthPrefixedAddr is a test-only exposure of parseLengthPrefixedAddr.
var ParseLengthPrefixedAddr = parseLengthPrefixedAddr
