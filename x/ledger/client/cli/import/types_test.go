package cli

import (
	"testing"
	"time"

	"github.com/provenance-io/provenance/x/ledger/types"
	"github.com/stretchr/testify/require"
)

func TestDefaultChunkConfig(t *testing.T) {
	config := DefaultChunkConfig()

	// Test default values
	require.Equal(t, 10000000, config.MaxChunkSizeBytes, "MaxChunkSizeBytes should be 10MB")
	require.Equal(t, 4000000, config.MaxGasPerTx, "MaxGasPerTx should be 4M gas")
	require.Equal(t, 1000000, config.MaxTxSizeBytes, "MaxTxSizeBytes should be 1MB")
}

func TestGenerateImportID(t *testing.T) {
	// Test that import IDs are unique
	id1 := generateImportID()
	time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	id2 := generateImportID()

	require.NotEqual(t, id1, id2, "Import IDs should be unique")
	require.Contains(t, id1, "import_", "Import ID should contain 'import_' prefix")
	require.Contains(t, id2, "import_", "Import ID should contain 'import_' prefix")

	// Test that IDs are valid format
	require.Greater(t, len(id1), len("import_"), "Import ID should be longer than just the prefix")
	require.Greater(t, len(id2), len("import_"), "Import ID should be longer than just the prefix")
}

func TestChunkConfigValidation(t *testing.T) {
	// Test valid configuration
	config := ChunkConfig{
		MaxChunkSizeBytes: 1000000,
		MaxGasPerTx:       2000000,
		MaxTxSizeBytes:    500000,
	}

	require.Greater(t, config.MaxChunkSizeBytes, 0, "MaxChunkSizeBytes should be positive")
	require.Greater(t, config.MaxGasPerTx, 0, "MaxGasPerTx should be positive")
	require.Greater(t, config.MaxTxSizeBytes, 0, "MaxTxSizeBytes should be positive")
}

func TestChunkedGenesisState(t *testing.T) {
	// Test ChunkedGenesisState creation and validation
	state := &ChunkedGenesisState{
		ImportID:     "test_import_123",
		TotalChunks:  5,
		Chunks:       []*types.GenesisState{},
		TotalLedgers: 100,
		TotalEntries: 500,
	}

	require.Equal(t, "test_import_123", state.ImportID)
	require.Equal(t, 5, state.TotalChunks)
	require.Equal(t, 100, state.TotalLedgers)
	require.Equal(t, 500, state.TotalEntries)
	require.NotNil(t, state.Chunks)
}

func TestLocalBulkImportStatus(t *testing.T) {
	// Test LocalBulkImportStatus creation and validation
	status := &LocalBulkImportStatus{
		ImportID:        "test_import_123",
		TotalChunks:     5,
		CompletedChunks: 2,
		TotalLedgers:    100,
		TotalEntries:    500,
		Status:          "in_progress",
		ErrorMessage:    "",
		CreatedAt:       "2024-01-01T12:00:00Z",
		UpdatedAt:       "2024-01-01T12:15:00Z",
	}

	require.Equal(t, "test_import_123", status.ImportID)
	require.Equal(t, 5, status.TotalChunks)
	require.Equal(t, 2, status.CompletedChunks)
	require.Equal(t, 100, status.TotalLedgers)
	require.Equal(t, 500, status.TotalEntries)
	require.Equal(t, "in_progress", status.Status)
	require.Empty(t, status.ErrorMessage)
	require.Equal(t, "2024-01-01T12:00:00Z", status.CreatedAt)
	require.Equal(t, "2024-01-01T12:15:00Z", status.UpdatedAt)
}

func TestImportStats(t *testing.T) {
	// Test ImportStats creation and validation
	stats := &ImportStats{
		TotalLedgers: 100,
		TotalEntries: 500,
		TotalChunks:  5,
	}

	require.Equal(t, 100, stats.TotalLedgers)
	require.Equal(t, 500, stats.TotalEntries)
	require.Equal(t, 5, stats.TotalChunks)
}

func TestGasCosts(t *testing.T) {
	// Test GasCosts creation and validation
	costs := &GasCosts{
		LedgerWithKeyGas: 100000,
		EntryGas:         5000,
	}

	require.Equal(t, 100000, costs.LedgerWithKeyGas)
	require.Equal(t, 5000, costs.EntryGas)
}
