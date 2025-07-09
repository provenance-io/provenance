package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"cosmossdk.io/log"
	"github.com/stretchr/testify/require"
)

func TestStreamingGenesisProcessor(t *testing.T) {
	// Create a temporary test file
	tmpFile, err := os.CreateTemp("", "test_genesis_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Create test data
	testData := createTestGenesisData(150) // More than the default chunk size

	// Write test data to file
	encoder := json.NewEncoder(tmpFile)
	err = encoder.Encode(testData)
	require.NoError(t, err)

	// Test streaming processor
	config := DefaultChunkConfig()
	config.MaxChunkSizeBytes = 50000 // Small chunk size for testing
	processor := NewStreamingGenesisProcessor(config, log.NewNopLogger())

	chunkedState, err := processor.ProcessFile(tmpFile.Name())
	require.NoError(t, err)

	// Verify results
	require.NotEmpty(t, chunkedState.ImportID)
	require.Equal(t, 150, chunkedState.TotalLedgers)
	require.Equal(t, 300, chunkedState.TotalEntries)       // 2 entries per ledger
	require.GreaterOrEqual(t, chunkedState.TotalChunks, 1) // Should have at least one chunk

	// Verify chunks are not empty
	require.NotEmpty(t, chunkedState.Chunks)

	// Verify total ledgers across all chunks matches expected
	totalLedgersInChunks := 0
	for _, chunk := range chunkedState.Chunks {
		totalLedgersInChunks += len(chunk.LedgerToEntries)
	}
	require.Equal(t, 150, totalLedgersInChunks)
}

func TestStreamingGenesisProcessorSmallFile(t *testing.T) {
	// Create a temporary test file
	tmpFile, err := os.CreateTemp("", "test_genesis_small_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Create test data smaller than chunk size
	testData := createTestGenesisData(25)

	// Write test data to file
	encoder := json.NewEncoder(tmpFile)
	err = encoder.Encode(testData)
	require.NoError(t, err)

	// Test streaming processor
	config := DefaultChunkConfig()
	config.MaxChunkSizeBytes = 1000000 // Large enough to fit all data
	processor := NewStreamingGenesisProcessor(config, log.NewNopLogger())

	chunkedState, err := processor.ProcessFile(tmpFile.Name())
	require.NoError(t, err)

	// Verify results - should fit in one chunk
	require.NotEmpty(t, chunkedState.ImportID)
	require.Equal(t, 25, chunkedState.TotalLedgers)
	require.Equal(t, 50, chunkedState.TotalEntries)
	require.Equal(t, 1, chunkedState.TotalChunks)
	require.Len(t, chunkedState.Chunks, 1)
	require.Len(t, chunkedState.Chunks[0].LedgerToEntries, 25)
}

func TestStreamingGenesisProcessorEmptyFile(t *testing.T) {
	// Create a temporary test file
	tmpFile, err := os.CreateTemp("", "test_genesis_empty_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Create empty test data
	testData := createTestGenesisData(0)

	// Write test data to file
	encoder := json.NewEncoder(tmpFile)
	err = encoder.Encode(testData)
	require.NoError(t, err)

	// Test streaming processor
	config := DefaultChunkConfig()
	processor := NewStreamingGenesisProcessor(config, log.NewNopLogger())

	chunkedState, err := processor.ProcessFile(tmpFile.Name())
	require.NoError(t, err)

	// Verify results
	require.NotEmpty(t, chunkedState.ImportID)
	require.Equal(t, 0, chunkedState.TotalLedgers)
	require.Equal(t, 0, chunkedState.TotalEntries)
	require.Equal(t, 0, chunkedState.TotalChunks)
	require.Len(t, chunkedState.Chunks, 0)
}

func TestStreamingGenesisProcessorInvalidJSON(t *testing.T) {
	// Create a temporary test file with invalid JSON
	tmpFile, err := os.CreateTemp("", "test_genesis_invalid_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write invalid JSON
	_, err = tmpFile.WriteString(`{"ledgerToEntries": [invalid json]}`)
	require.NoError(t, err)

	// Test streaming processor
	config := DefaultChunkConfig()
	processor := NewStreamingGenesisProcessor(config, log.NewNopLogger())

	_, err = processor.ProcessFile(tmpFile.Name())
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode LedgerToEntries")
}

func TestStreamingGenesisProcessorMissingField(t *testing.T) {
	// Create a temporary test file with missing required fields
	tmpFile, err := os.CreateTemp("", "test_genesis_missing_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Create test data with missing fields
	testData := map[string]interface{}{
		"ledgerToEntries": []map[string]interface{}{
			{
				"ledgerKey": map[string]interface{}{
					"nftId":        "", // Empty NFT ID should cause validation error
					"assetClassId": "test-asset-1",
				},
				"ledger": map[string]interface{}{
					"ledgerClassId": "test-class-1",
					"statusTypeId":  1,
				},
				"entries": []interface{}{},
			},
		},
	}

	// Write test data to file
	encoder := json.NewEncoder(tmpFile)
	err = encoder.Encode(testData)
	require.NoError(t, err)

	// Test streaming processor
	config := DefaultChunkConfig()
	processor := NewStreamingGenesisProcessor(config, log.NewNopLogger())

	_, err = processor.ProcessFile(tmpFile.Name())
	require.Error(t, err)
	require.Contains(t, err.Error(), "ledger key NftId is empty")
}

func TestStreamingGenesisProcessorLargeLedger(t *testing.T) {
	// Create a temporary test file
	tmpFile, err := os.CreateTemp("", "test_genesis_large_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Create test data with one large ledger (many entries)
	testData := createTestGenesisDataWithLargeLedger(1, 2000) // 1 ledger with 2000 entries

	// Write test data to file
	encoder := json.NewEncoder(tmpFile)
	err = encoder.Encode(testData)
	require.NoError(t, err)

	// Test streaming processor
	config := DefaultChunkConfig()
	config.MaxChunkSizeBytes = 100000 // Small chunk size to force splitting
	processor := NewStreamingGenesisProcessor(config, log.NewNopLogger())

	chunkedState, err := processor.ProcessFile(tmpFile.Name())
	require.NoError(t, err)

	// Verify results
	require.NotEmpty(t, chunkedState.ImportID)
	require.Equal(t, 1, chunkedState.TotalLedgers)
	require.Equal(t, 2000, chunkedState.TotalEntries)
	require.GreaterOrEqual(t, chunkedState.TotalChunks, 1) // Should have at least one chunk

	// Verify that the data was processed correctly
	totalEntriesInChunks := 0
	for _, chunk := range chunkedState.Chunks {
		for _, lte := range chunk.LedgerToEntries {
			totalEntriesInChunks += len(lte.Entries)
		}
	}
	require.Equal(t, 2000, totalEntriesInChunks)
}

func TestStreamingGenesisProcessorMixedData(t *testing.T) {
	// Create a temporary test file
	tmpFile, err := os.CreateTemp("", "test_genesis_mixed_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Create test data with mixed ledger sizes
	testData := createTestGenesisDataMixed(10) // 10 ledgers with varying entry counts

	// Write test data to file
	encoder := json.NewEncoder(tmpFile)
	err = encoder.Encode(testData)
	require.NoError(t, err)

	// Test streaming processor
	config := DefaultChunkConfig()
	config.MaxChunkSizeBytes = 50000 // Small chunk size to force chunking
	processor := NewStreamingGenesisProcessor(config, log.NewNopLogger())

	chunkedState, err := processor.ProcessFile(tmpFile.Name())
	require.NoError(t, err)

	// Verify results
	require.NotEmpty(t, chunkedState.ImportID)
	require.Equal(t, 10, chunkedState.TotalLedgers)
	require.Greater(t, chunkedState.TotalEntries, 0)
	require.GreaterOrEqual(t, chunkedState.TotalChunks, 1)
}

func TestStreamingGenesisProcessorFileNotFound(t *testing.T) {
	// Test streaming processor with non-existent file
	config := DefaultChunkConfig()
	processor := NewStreamingGenesisProcessor(config, log.NewNopLogger())

	_, err := processor.ProcessFile("non_existent_file.json")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to open genesis state file")
}

func TestStreamingGenesisProcessorMalformedJSON(t *testing.T) {
	// Create a temporary test file with malformed JSON
	tmpFile, err := os.CreateTemp("", "test_genesis_malformed_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write malformed JSON (missing closing brace)
	_, err = tmpFile.WriteString(`{"ledgerToEntries": [{"ledgerKey": {"nftId": "test"}}`)
	require.NoError(t, err)

	// Test streaming processor
	config := DefaultChunkConfig()
	processor := NewStreamingGenesisProcessor(config, log.NewNopLogger())

	_, err = processor.ProcessFile(tmpFile.Name())
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse JSON")
}

func TestStreamingGenesisProcessorNilLedgerKey(t *testing.T) {
	// Create a temporary test file with nil ledger key
	tmpFile, err := os.CreateTemp("", "test_genesis_nil_key_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Create test data with nil ledger key
	testData := map[string]interface{}{
		"ledgerToEntries": []map[string]interface{}{
			{
				"ledgerKey": nil, // Nil ledger key should cause validation error
				"ledger": map[string]interface{}{
					"ledgerClassId": "test-class-1",
					"statusTypeId":  1,
				},
				"entries": []interface{}{},
			},
		},
	}

	// Write test data to file
	encoder := json.NewEncoder(tmpFile)
	err = encoder.Encode(testData)
	require.NoError(t, err)

	// Test streaming processor
	config := DefaultChunkConfig()
	processor := NewStreamingGenesisProcessor(config, log.NewNopLogger())

	_, err = processor.ProcessFile(tmpFile.Name())
	require.Error(t, err)
	require.Contains(t, err.Error(), "ledger key is nil")
}

// createTestGenesisData creates test genesis data with the specified number of ledgers
func createTestGenesisData(numLedgers int) map[string]interface{} {
	ledgerToEntries := make([]map[string]interface{}, numLedgers)

	for i := 0; i < numLedgers; i++ {
		ledgerToEntries[i] = map[string]interface{}{
			"ledgerKey": map[string]interface{}{
				"nftId":        fmt.Sprintf("test-nft-%d", i),
				"assetClassId": "test-asset-1",
			},
			"ledger": map[string]interface{}{
				"ledgerClassId": "test-class-1",
				"statusTypeId":  1,
			},
			"entries": []map[string]interface{}{
				{
					"correlationId": fmt.Sprintf("entry-1-%d", i),
					"sequence":      1,
					"entryTypeId":   1,
					"postedDate":    20000,
					"effectiveDate": 20000,
					"totalAmt":      "1000000",
					"appliedAmounts": []map[string]interface{}{
						{
							"bucketTypeId": 1,
							"appliedAmt":   "1000000",
						},
					},
				},
				{
					"correlationId": fmt.Sprintf("entry-2-%d", i),
					"sequence":      2,
					"entryTypeId":   2,
					"postedDate":    20001,
					"effectiveDate": 20001,
					"totalAmt":      "500000",
					"appliedAmounts": []map[string]interface{}{
						{
							"bucketTypeId": 2,
							"appliedAmt":   "500000",
						},
					},
				},
			},
		}
	}

	return map[string]interface{}{
		"ledgerToEntries": ledgerToEntries,
	}
}

// createTestGenesisDataWithLargeLedger creates test data with one ledger containing many entries
func createTestGenesisDataWithLargeLedger(numLedgers, entriesPerLedger int) map[string]interface{} {
	ledgerToEntries := make([]map[string]interface{}, numLedgers)

	for i := 0; i < numLedgers; i++ {
		entries := make([]map[string]interface{}, entriesPerLedger)
		for j := 0; j < entriesPerLedger; j++ {
			entries[j] = map[string]interface{}{
				"correlationId": fmt.Sprintf("entry-%d-%d", i, j),
				"sequence":      j + 1,
				"entryTypeId":   1,
				"postedDate":    20000 + j,
				"effectiveDate": 20000 + j,
				"totalAmt":      "1000000",
				"appliedAmounts": []map[string]interface{}{
					{
						"bucketTypeId": 1,
						"appliedAmt":   "1000000",
					},
				},
			}
		}

		ledgerToEntries[i] = map[string]interface{}{
			"ledgerKey": map[string]interface{}{
				"nftId":        fmt.Sprintf("test-nft-%d", i),
				"assetClassId": "test-asset-1",
			},
			"ledger": map[string]interface{}{
				"ledgerClassId": "test-class-1",
				"statusTypeId":  1,
			},
			"entries": entries,
		}
	}

	return map[string]interface{}{
		"ledgerToEntries": ledgerToEntries,
	}
}

// createTestGenesisDataMixed creates test data with ledgers of varying sizes
func createTestGenesisDataMixed(numLedgers int) map[string]interface{} {
	ledgerToEntries := make([]map[string]interface{}, numLedgers)

	for i := 0; i < numLedgers; i++ {
		// Vary the number of entries per ledger (1 to 100)
		entriesCount := (i % 100) + 1
		entries := make([]map[string]interface{}, entriesCount)

		for j := 0; j < entriesCount; j++ {
			entries[j] = map[string]interface{}{
				"correlationId": fmt.Sprintf("entry-%d-%d", i, j),
				"sequence":      j + 1,
				"entryTypeId":   1,
				"postedDate":    20000 + j,
				"effectiveDate": 20000 + j,
				"totalAmt":      "1000000",
				"appliedAmounts": []map[string]interface{}{
					{
						"bucketTypeId": 1,
						"appliedAmt":   "1000000",
					},
				},
			}
		}

		ledgerToEntries[i] = map[string]interface{}{
			"ledgerKey": map[string]interface{}{
				"nftId":        fmt.Sprintf("test-nft-%d", i),
				"assetClassId": "test-asset-1",
			},
			"ledger": map[string]interface{}{
				"ledgerClassId": "test-class-1",
				"statusTypeId":  1,
			},
			"entries": entries,
		}
	}

	return map[string]interface{}{
		"ledgerToEntries": ledgerToEntries,
	}
}
