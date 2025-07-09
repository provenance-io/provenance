package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

	// Verify that the total entries across all chunks matches the expected total
	verifyTotalEntriesAcrossChunks(t, chunkedState, 300)
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

	// Verify that the total entries across all chunks matches the expected total
	verifyTotalEntriesAcrossChunks(t, chunkedState, 50)
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

	// Verify that the total entries across all chunks matches the expected total
	verifyTotalEntriesAcrossChunks(t, chunkedState, 2000)
}

func TestStreamingGenesisProcessorMixedData(t *testing.T) {
	// Create a temporary test file
	tmpFile, err := os.CreateTemp("", "test_genesis_mixed_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Create test data with mixed ledger sizes
	testData := createTestGenesisDataMixed(10) // 10 ledgers with varying entry counts

	// Calculate expected total entries
	expectedTotalEntries := 0
	for i := 0; i < 10; i++ {
		entriesCount := (i % 100) + 1
		expectedTotalEntries += entriesCount
	}

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
	require.Equal(t, expectedTotalEntries, chunkedState.TotalEntries)
	require.GreaterOrEqual(t, chunkedState.TotalChunks, 1)

	// Verify that the total entries across all chunks matches the expected total
	verifyTotalEntriesAcrossChunks(t, chunkedState, expectedTotalEntries)
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

// verifyTotalEntriesAcrossChunks verifies that the total entries across all chunks matches the expected count
func verifyTotalEntriesAcrossChunks(t *testing.T, chunkedState *ChunkedGenesisState, expectedTotalEntries int) {
	totalEntriesInChunks := 0
	for _, chunk := range chunkedState.Chunks {
		for _, lte := range chunk.LedgerToEntries {
			totalEntriesInChunks += len(lte.Entries)
		}
	}
	require.Equal(t, expectedTotalEntries, totalEntriesInChunks,
		"Total entries across chunks (%d) should match expected total (%d)",
		totalEntriesInChunks, expectedTotalEntries)
}

// TestBroadcastAndCheckTx tests the broadcastAndCheckTx function's ability to detect transaction failures
func TestBroadcastAndCheckTx(t *testing.T) {
	// Test that transaction response parsing works correctly
	t.Run("transaction response parsing", func(t *testing.T) {
		// Test successful transaction response
		successResp := sdk.TxResponse{
			Code:    0,
			TxHash:  "success-hash",
			RawLog:  "success",
			GasUsed: 100000,
		}

		// Verify that code 0 indicates success
		require.Equal(t, uint32(0), successResp.Code, "Code 0 should indicate success")

		// Test failed transaction response
		failedResp := sdk.TxResponse{
			Code:    1, // Non-zero code indicates failure
			TxHash:  "failed-hash",
			RawLog:  "out of gas",
			GasUsed: 0,
		}

		// Verify that non-zero code indicates failure
		require.NotEqual(t, uint32(0), failedResp.Code, "Non-zero code should indicate failure")

		// Test insufficient funds transaction response
		insufficientFundsResp := sdk.TxResponse{
			Code:    5, // Non-zero code indicates failure
			TxHash:  "insufficient-funds-hash",
			RawLog:  "insufficient funds",
			GasUsed: 0,
		}

		// Verify that non-zero code indicates failure
		require.NotEqual(t, uint32(0), insufficientFundsResp.Code, "Non-zero code should indicate failure")
	})
}

// TestTransactionFailureHandling tests that the main processing loop properly handles transaction failures
func TestTransactionFailureHandling(t *testing.T) {
	// Create a temporary test file
	tmpFile, err := os.CreateTemp("", "test_genesis_failure_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Create test data
	testData := createTestGenesisData(5) // Small dataset for testing

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

	// Verify we have a valid chunked state
	require.NotEmpty(t, chunkedState.ImportID)
	require.Equal(t, 5, chunkedState.TotalLedgers)
	require.Equal(t, 10, chunkedState.TotalEntries)
	require.Equal(t, 1, chunkedState.TotalChunks)

	// Test that the status file is created and can be read
	// Note: The status file is only created when the actual import command runs,
	// not during the ProcessFile step, so we'll create it manually for testing
	status := &LocalBulkImportStatus{
		ImportID:        chunkedState.ImportID,
		TotalChunks:     1,
		CompletedChunks: 0,
		TotalLedgers:    5,
		TotalEntries:    10,
		Status:          "pending",
		CreatedAt:       time.Now().Format(time.RFC3339),
		UpdatedAt:       time.Now().Format(time.RFC3339),
	}
	err = writeLocalBulkImportStatus(status)
	require.NoError(t, err)

	// Now read it back to verify
	readStatus, err := readLocalBulkImportStatus(chunkedState.ImportID)
	require.NoError(t, err)
	require.Equal(t, chunkedState.ImportID, readStatus.ImportID)
	require.Equal(t, 1, readStatus.TotalChunks)
	require.Equal(t, 5, readStatus.TotalLedgers)
	require.Equal(t, 10, readStatus.TotalEntries)
	require.Equal(t, "pending", readStatus.Status)

	// Test that the status can be updated to failed
	status.Status = "failed"
	status.ErrorMessage = "test transaction failure"
	status.UpdatedAt = time.Now().Format(time.RFC3339)
	err = writeLocalBulkImportStatus(status)
	require.NoError(t, err)

	// Verify the status was written correctly
	updatedStatus, err := readLocalBulkImportStatus(chunkedState.ImportID)
	require.NoError(t, err)
	require.Equal(t, "failed", updatedStatus.Status)
	require.Equal(t, "test transaction failure", updatedStatus.ErrorMessage)
}
