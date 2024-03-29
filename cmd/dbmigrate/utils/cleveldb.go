//go:build cleveldb
// +build cleveldb

package utils

import (
	tmdb "github.com/cometbft/cometbft-db"
)

// This file is included when built with the cleveldb tag (which matches the tag Tendermint looks for).
// Tendermint does all the heavy lifting, but doesn't expose a way to identify which DB types are available.
// That list would also have MemDB, which we don't want in here anyway.
// That's all this is doing, just identifying that it was built with that tag and that this DB type is available.

func init() {
	AddPossibleDBType(tmdb.CLevelDBBackend)
}
