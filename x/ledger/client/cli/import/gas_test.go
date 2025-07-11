package cli

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/provenance-io/provenance/x/ledger/types"
	"github.com/stretchr/testify/require"
)

func TestGetChunkSizeBytes(t *testing.T) {
	// Test with empty chunk
	emptyChunk := &types.GenesisState{
		LedgerToEntries: []types.LedgerToEntries{},
	}
	size := getChunkSizeBytes(emptyChunk)
	require.Greater(t, size, 0, "Empty chunk should have some size due to JSON structure")

	// Test with chunk containing data
	chunk := &types.GenesisState{
		LedgerToEntries: []types.LedgerToEntries{
			{
				LedgerKey: &types.LedgerKey{
					NftId:        "test-nft-1",
					AssetClassId: "test-asset-1",
				},
				Ledger: &types.Ledger{
					LedgerClassId: "test-class-1",
					StatusTypeId:  1,
				},
				Entries: []*types.LedgerEntry{
					{
						CorrelationId: "entry-1",
						Sequence:      1,
						EntryTypeId:   1,
						PostedDate:    20000,
						EffectiveDate: 20000,
						TotalAmt:      math.NewInt(1000000),
					},
				},
			},
		},
	}
	size = getChunkSizeBytes(chunk)
	require.Greater(t, size, 0, "Chunk with data should have positive size")

	// Test that larger chunks have larger sizes
	largerChunk := &types.GenesisState{
		LedgerToEntries: []types.LedgerToEntries{
			{
				LedgerKey: &types.LedgerKey{
					NftId:        "test-nft-1",
					AssetClassId: "test-asset-1",
				},
				Ledger: &types.Ledger{
					LedgerClassId: "test-class-1",
					StatusTypeId:  1,
				},
				Entries: []*types.LedgerEntry{
					{
						CorrelationId: "entry-1",
						Sequence:      1,
						EntryTypeId:   1,
						PostedDate:    20000,
						EffectiveDate: 20000,
						TotalAmt:      math.NewInt(1000000),
					},
					{
						CorrelationId: "entry-2",
						Sequence:      2,
						EntryTypeId:   2,
						PostedDate:    20001,
						EffectiveDate: 20001,
						TotalAmt:      math.NewInt(2000000),
					},
				},
			},
		},
	}
	largerSize := getChunkSizeBytes(largerChunk)
	require.Greater(t, largerSize, size, "Larger chunk should have larger size")
}

func TestEstimateChunkGasFromCosts(t *testing.T) {
	costs := &GasCosts{
		LedgerWithKeyGas: 100000,
		EntryGas:         5000,
	}

	// Test with chunk containing ledger and entries
	chunk := &types.GenesisState{
		LedgerToEntries: []types.LedgerToEntries{
			{
				LedgerKey: &types.LedgerKey{
					NftId:        "test-nft-1",
					AssetClassId: "test-asset-1",
				},
				Ledger: &types.Ledger{
					LedgerClassId: "test-class-1",
					StatusTypeId:  1,
				},
				Entries: []*types.LedgerEntry{
					{
						CorrelationId: "entry-1",
						Sequence:      1,
						EntryTypeId:   1,
						PostedDate:    20000,
						EffectiveDate: 20000,
						TotalAmt:      math.NewInt(1000000),
					},
					{
						CorrelationId: "entry-2",
						Sequence:      2,
						EntryTypeId:   2,
						PostedDate:    20001,
						EffectiveDate: 20001,
						TotalAmt:      math.NewInt(2000000),
					},
				},
			},
		},
	}

	estimatedGas := estimateChunkGasFromCosts(chunk, costs)
	expectedGas := costs.LedgerWithKeyGas + (2 * costs.EntryGas) // 1 ledger + 2 entries
	require.Equal(t, expectedGas, estimatedGas, "Estimated gas should match expected calculation")

	// Test with chunk containing only ledger key (no ledger object)
	chunkWithKeyOnly := &types.GenesisState{
		LedgerToEntries: []types.LedgerToEntries{
			{
				LedgerKey: &types.LedgerKey{
					NftId:        "test-nft-1",
					AssetClassId: "test-asset-1",
				},
				Ledger: nil, // No ledger object
				Entries: []*types.LedgerEntry{
					{
						CorrelationId: "entry-1",
						Sequence:      1,
						EntryTypeId:   1,
						PostedDate:    20000,
						EffectiveDate: 20000,
						TotalAmt:      math.NewInt(1000000),
					},
				},
			},
		},
	}

	estimatedGas = estimateChunkGasFromCosts(chunkWithKeyOnly, costs)
	expectedGas = 1 * costs.EntryGas // Only entry gas (no ledger gas)
	require.Equal(t, expectedGas, estimatedGas, "Estimated gas should only include entry costs")

	// Test with empty chunk
	emptyChunk := &types.GenesisState{
		LedgerToEntries: []types.LedgerToEntries{},
	}
	estimatedGas = estimateChunkGasFromCosts(emptyChunk, costs)
	require.Equal(t, 0, estimatedGas, "Empty chunk should have zero gas cost")

	// Test with chunk containing ledger but no entries
	chunkWithLedgerOnly := &types.GenesisState{
		LedgerToEntries: []types.LedgerToEntries{
			{
				LedgerKey: &types.LedgerKey{
					NftId:        "test-nft-1",
					AssetClassId: "test-asset-1",
				},
				Ledger: &types.Ledger{
					LedgerClassId: "test-class-1",
					StatusTypeId:  1,
				},
				Entries: []*types.LedgerEntry{}, // No entries
			},
		},
	}

	estimatedGas = estimateChunkGasFromCosts(chunkWithLedgerOnly, costs)
	expectedGas = costs.LedgerWithKeyGas // Only ledger gas
	require.Equal(t, expectedGas, estimatedGas, "Estimated gas should only include ledger costs")
}

func TestEstimateChunkGasFromCostsNilLedgerKey(t *testing.T) {
	costs := &GasCosts{
		LedgerWithKeyGas: 100000,
		EntryGas:         5000,
	}

	// Test with chunk containing nil ledger key
	chunk := &types.GenesisState{
		LedgerToEntries: []types.LedgerToEntries{
			{
				LedgerKey: nil, // Nil ledger key
				Ledger: &types.Ledger{
					LedgerClassId: "test-class-1",
					StatusTypeId:  1,
				},
				Entries: []*types.LedgerEntry{
					{
						CorrelationId: "entry-1",
						Sequence:      1,
						EntryTypeId:   1,
						PostedDate:    20000,
						EffectiveDate: 20000,
						TotalAmt:      math.NewInt(1000000),
					},
				},
			},
		},
	}

	estimatedGas := estimateChunkGasFromCosts(chunk, costs)
	require.Equal(t, 0, estimatedGas, "Should return 0 gas when ledger key is nil")
}

func TestGasCostsValidation(t *testing.T) {
	// Test valid gas costs
	costs := &GasCosts{
		LedgerWithKeyGas: 100000,
		EntryGas:         5000,
	}

	require.GreaterOrEqual(t, costs.LedgerWithKeyGas, 0, "LedgerWithKeyGas should be non-negative")
	require.GreaterOrEqual(t, costs.EntryGas, 0, "EntryGas should be non-negative")

	// Test zero costs
	zeroCosts := &GasCosts{
		LedgerWithKeyGas: 0,
		EntryGas:         0,
	}

	require.Equal(t, 0, zeroCosts.LedgerWithKeyGas, "LedgerWithKeyGas can be zero")
	require.Equal(t, 0, zeroCosts.EntryGas, "EntryGas can be zero")
}

func TestGasCostsStorageAndReuse(t *testing.T) {
	// Test that gas costs can be stored and reused
	costs := &GasCosts{
		LedgerWithKeyGas: 150000,
		EntryGas:         7500,
	}

	// Create a test chunk
	chunk := &types.GenesisState{
		LedgerToEntries: []types.LedgerToEntries{
			{
				LedgerKey: &types.LedgerKey{
					NftId:        "test-nft-1",
					AssetClassId: "test-asset-1",
				},
				Ledger: &types.Ledger{
					LedgerClassId: "test-class-1",
					StatusTypeId:  1,
				},
				Entries: []*types.LedgerEntry{
					{
						CorrelationId: "entry-1",
						Sequence:      1,
						EntryTypeId:   1,
						PostedDate:    20000,
						EffectiveDate: 20000,
						TotalAmt:      math.NewInt(1000000),
					},
					{
						CorrelationId: "entry-2",
						Sequence:      2,
						EntryTypeId:   2,
						PostedDate:    20001,
						EffectiveDate: 20001,
						TotalAmt:      math.NewInt(2000000),
					},
				},
			},
		},
	}

	// Estimate gas using the costs
	estimatedGas := estimateChunkGasFromCosts(chunk, costs)
	expectedGas := costs.LedgerWithKeyGas + (2 * costs.EntryGas) // 1 ledger + 2 entries
	require.Equal(t, expectedGas, estimatedGas, "Estimated gas should match expected calculation")

	// Test that the same costs produce consistent results
	estimatedGas2 := estimateChunkGasFromCosts(chunk, costs)
	require.Equal(t, estimatedGas, estimatedGas2, "Gas estimation should be consistent with same costs")

	// Test with different costs
	differentCosts := &GasCosts{
		LedgerWithKeyGas: 200000,
		EntryGas:         10000,
	}
	estimatedGas3 := estimateChunkGasFromCosts(chunk, differentCosts)
	expectedGas3 := differentCosts.LedgerWithKeyGas + (2 * differentCosts.EntryGas)
	require.Equal(t, expectedGas3, estimatedGas3, "Different costs should produce different estimates")
	require.NotEqual(t, estimatedGas, estimatedGas3, "Different costs should produce different estimates")
}
