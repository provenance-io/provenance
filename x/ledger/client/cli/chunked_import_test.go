package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStreamingGenesisProcessor(t *testing.T) {
	// Create a temporary test file
	tmpFile, err := os.CreateTemp("", "test_genesis_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Create test data
	testData := createTestGenesisData(150) // More than the default chunk size of 100

	// Write test data to file
	encoder := json.NewEncoder(tmpFile)
	err = encoder.Encode(testData)
	require.NoError(t, err)

	// Test streaming processor
	config := DefaultChunkConfig()
	config.MaxLedgersPerChunk = 50 // Small chunk size for testing
	processor := NewStreamingGenesisProcessor(config)

	chunkedState, err := processor.ProcessFile(tmpFile.Name())
	require.NoError(t, err)

	// Verify results
	require.NotEmpty(t, chunkedState.ImportID)
	require.Equal(t, 150, chunkedState.TotalLedgers)
	require.Equal(t, 300, chunkedState.TotalEntries) // 2 entries per ledger
	require.Equal(t, 3, chunkedState.TotalChunks)    // 150 ledgers / 50 per chunk = 3 chunks

	// Verify chunk sizes
	require.Len(t, chunkedState.Chunks, 3)
	require.Len(t, chunkedState.Chunks[0].LedgerToEntries, 50)
	require.Len(t, chunkedState.Chunks[1].LedgerToEntries, 50)
	require.Len(t, chunkedState.Chunks[2].LedgerToEntries, 50)
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
	config.MaxLedgersPerChunk = 100
	processor := NewStreamingGenesisProcessor(config)

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
	processor := NewStreamingGenesisProcessor(config)

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
	processor := NewStreamingGenesisProcessor(config)

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
	processor := NewStreamingGenesisProcessor(config)

	_, err = processor.ProcessFile(tmpFile.Name())
	require.Error(t, err)
	require.Contains(t, err.Error(), "ledger key NftId is empty")
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
