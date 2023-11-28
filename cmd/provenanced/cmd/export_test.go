package cmd

// This file is in the cmd package (not cmd_test) so that it can expose
// some private keeper stuff for unit testing.

var (
	// MakeDefaultMarket is a test-only exposure of makeDefaultMarket.
	MakeDefaultMarket = makeDefaultMarket
	// AddMarketsToAppState is a test-only exposure of addMarketsToAppState.
	AddMarketsToAppState = addMarketsToAppState
	// GetNextAvailableMarketID is a test-only exposure of getNextAvailableMarketID.
	GetNextAvailableMarketID = getNextAvailableMarketID
)
