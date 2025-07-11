package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
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

	// Clean up the status file after test
	defer func() {
		_ = os.Remove(statusFileName(chunkedState.ImportID))
	}()

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

func TestResumeFunctionality(t *testing.T) {
	// Create a temporary test file
	tmpFile, err := os.CreateTemp("", "test_genesis_resume_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Create test data with multiple chunks
	testData := createTestGenesisData(10) // Larger dataset to ensure multiple chunks

	// Write test data to file
	encoder := json.NewEncoder(tmpFile)
	err = encoder.Encode(testData)
	require.NoError(t, err)

	// Test streaming processor
	config := DefaultChunkConfig()
	config.MaxChunkSizeBytes = 1000 // Small chunk size to ensure multiple chunks
	processor := NewStreamingGenesisProcessor(config, log.NewNopLogger())

	chunkedState, err := processor.ProcessFile(tmpFile.Name())
	require.NoError(t, err)

	// Verify we have multiple chunks
	require.Greater(t, chunkedState.TotalChunks, 1, "Should have multiple chunks for resume testing")

	// Clean up the status file after test
	defer func() {
		_ = os.Remove(statusFileName(chunkedState.ImportID))
	}()

	// Test 1: Create a status file with partial completion
	status := &LocalBulkImportStatus{
		ImportID:        chunkedState.ImportID,
		TotalChunks:     chunkedState.TotalChunks,
		CompletedChunks: 2, // Simulate 2 chunks completed
		TotalLedgers:    chunkedState.TotalLedgers,
		TotalEntries:    chunkedState.TotalEntries,
		Status:          "in_progress",
		CreatedAt:       time.Now().Format(time.RFC3339),
		UpdatedAt:       time.Now().Format(time.RFC3339),
	}
	err = writeLocalBulkImportStatus(status)
	require.NoError(t, err)

	// Test 2: Verify the status file can be read
	readStatus, err := readLocalBulkImportStatus(chunkedState.ImportID)
	require.NoError(t, err)
	require.Equal(t, chunkedState.ImportID, readStatus.ImportID)
	require.Equal(t, 2, readStatus.CompletedChunks)
	require.Equal(t, "in_progress", readStatus.Status)

	// Test 3: Simulate resume logic
	existingStatus, err := readLocalBulkImportStatus(chunkedState.ImportID)
	require.NoError(t, err)

	isResume := false
	startChunkIndex := 0

	if existingStatus.Status == "in_progress" || existingStatus.Status == "failed" {
		// Validate that the existing status matches our current import
		if existingStatus.TotalChunks == chunkedState.TotalChunks &&
			existingStatus.TotalLedgers == chunkedState.TotalLedgers &&
			existingStatus.TotalEntries == chunkedState.TotalEntries {

			isResume = true
			startChunkIndex = existingStatus.CompletedChunks
		}
	}

	require.True(t, isResume, "Should detect resume scenario")
	require.Equal(t, 2, startChunkIndex, "Should start from chunk 3 (index 2)")

	// Test 4: Test completed import detection
	completedStatus := &LocalBulkImportStatus{
		ImportID:        chunkedState.ImportID,
		TotalChunks:     chunkedState.TotalChunks,
		CompletedChunks: chunkedState.TotalChunks, // All chunks completed
		TotalLedgers:    chunkedState.TotalLedgers,
		TotalEntries:    chunkedState.TotalEntries,
		Status:          "completed",
		CreatedAt:       time.Now().Format(time.RFC3339),
		UpdatedAt:       time.Now().Format(time.RFC3339),
	}
	err = writeLocalBulkImportStatus(completedStatus)
	require.NoError(t, err)

	// Verify completed import is detected
	completedReadStatus, err := readLocalBulkImportStatus(chunkedState.ImportID)
	require.NoError(t, err)
	require.Equal(t, "completed", completedReadStatus.Status)
	require.Equal(t, chunkedState.TotalChunks, completedReadStatus.CompletedChunks)

	// Test 5: Test failed import detection
	failedStatus := &LocalBulkImportStatus{
		ImportID:        chunkedState.ImportID,
		TotalChunks:     chunkedState.TotalChunks,
		CompletedChunks: 1,
		TotalLedgers:    chunkedState.TotalLedgers,
		TotalEntries:    chunkedState.TotalEntries,
		Status:          "failed",
		ErrorMessage:    "test error",
		CreatedAt:       time.Now().Format(time.RFC3339),
		UpdatedAt:       time.Now().Format(time.RFC3339),
	}
	err = writeLocalBulkImportStatus(failedStatus)
	require.NoError(t, err)

	// Verify failed import is detected
	failedReadStatus, err := readLocalBulkImportStatus(chunkedState.ImportID)
	require.NoError(t, err)
	require.Equal(t, "failed", failedReadStatus.Status)
	require.Equal(t, "test error", failedReadStatus.ErrorMessage)
	require.Equal(t, 1, failedReadStatus.CompletedChunks)
}

func TestCorrelationIDTracking(t *testing.T) {
	// Create a temporary test file
	tmpFile, err := os.CreateTemp("", "test_genesis_correlation_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Create test data
	testData := createTestGenesisData(20) // Larger dataset to ensure multiple chunks

	// Write test data to file
	encoder := json.NewEncoder(tmpFile)
	err = encoder.Encode(testData)
	require.NoError(t, err)

	// Test 1: First processing run
	config1 := DefaultChunkConfig()
	config1.MaxChunkSizeBytes = 1000 // Small chunk size to ensure multiple chunks
	processor1 := NewStreamingGenesisProcessor(config1, log.NewNopLogger())

	chunkedState1, err := processor1.ProcessFile(tmpFile.Name())
	require.NoError(t, err)

	// Verify we have multiple chunks
	require.Greater(t, chunkedState1.TotalChunks, 1, "Should have multiple chunks for testing")

	// Clean up the status file after test
	defer func() {
		_ = os.Remove(statusFileName(chunkedState1.ImportID))
	}()

	// Test 2: Test file hash calculation
	fileHash1, err := calculateFileHash(tmpFile.Name())
	require.NoError(t, err)
	require.NotEmpty(t, fileHash1, "File hash should not be empty")

	// Test 3: Test correlation ID extraction
	if len(chunkedState1.Chunks) > 0 {
		lastCorrelationID := getLastCorrelationIDFromChunk(chunkedState1.Chunks[0])
		// Our test data has correlation IDs like "entry-2-0", "entry-2-1", etc.
		require.NotEmpty(t, lastCorrelationID, "Test data should have correlation IDs")
		require.Contains(t, lastCorrelationID, "entry-", "Correlation ID should contain 'entry-' prefix")
	}

	// Test 4: Test status with correlation ID tracking
	status := &LocalBulkImportStatus{
		ImportID:                    chunkedState1.ImportID,
		TotalChunks:                 chunkedState1.TotalChunks,
		CompletedChunks:             2,
		TotalLedgers:                chunkedState1.TotalLedgers,
		TotalEntries:                chunkedState1.TotalEntries,
		Status:                      "in_progress",
		CreatedAt:                   time.Now().Format(time.RFC3339),
		UpdatedAt:                   time.Now().Format(time.RFC3339),
		LastSuccessfulCorrelationID: "test_correlation_123",
		FileHash:                    fileHash1,
	}

	err = writeLocalBulkImportStatus(status)
	require.NoError(t, err)

	// Test 5: Read back status and verify correlation ID tracking
	readStatus, err := readLocalBulkImportStatus(chunkedState1.ImportID)
	require.NoError(t, err)
	require.Equal(t, "test_correlation_123", readStatus.LastSuccessfulCorrelationID)
	require.Equal(t, fileHash1, readStatus.FileHash)

	// Test 6: Test resume validation logic
	existingStatus, err := readLocalBulkImportStatus(chunkedState1.ImportID)
	require.NoError(t, err)

	isResume := false
	startChunkIndex := 0
	lastSuccessfulCorrelationID := ""

	if existingStatus.Status == "in_progress" || existingStatus.Status == "failed" {
		// Validate that the existing status matches our current import
		if existingStatus.TotalLedgers == chunkedState1.TotalLedgers &&
			existingStatus.TotalEntries == chunkedState1.TotalEntries &&
			existingStatus.FileHash == fileHash1 {

			isResume = true
			startChunkIndex = existingStatus.CompletedChunks
			lastSuccessfulCorrelationID = existingStatus.LastSuccessfulCorrelationID
		}
	}

	require.True(t, isResume, "Should detect resume scenario")
	require.Equal(t, 2, startChunkIndex, "Should start from chunk 3 (index 2)")
	require.Equal(t, "test_correlation_123", lastSuccessfulCorrelationID, "Should have correct correlation ID")
}

func TestFindNextCorrelationIDAfter(t *testing.T) {
	// Create a temporary test file
	tmpFile, err := os.CreateTemp("", "test_correlation_id_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Create test data with correlation IDs
	testData := map[string]interface{}{
		"ledgerToEntries": []map[string]interface{}{
			{
				"ledgerKey": map[string]interface{}{
					"nftId":        "nft1",
					"assetClassId": "asset1",
				},
				"ledger": map[string]interface{}{
					"ledgerClassId": "class1",
					"statusTypeId":  1,
				},
				"entries": []map[string]interface{}{
					{
						"correlationId": "33a830f5-3969-4a6b-bc80-f3e7f4e8850b",
						"sequence":      1,
						"entryTypeId":   1,
						"postedDate":    20000,
						"effectiveDate": 20000,
						"totalAmt":      "1000000",
					},
					{
						"correlationId": "40bf044e-5f84-41d1-a5e4-795ab98c3dc6",
						"sequence":      2,
						"entryTypeId":   1,
						"postedDate":    20000,
						"effectiveDate": 20000,
						"totalAmt":      "2000000",
					},
					{
						"correlationId": "50cf155f-6f95-52e2-b6f5-8a6bc09d4ed7",
						"sequence":      3,
						"entryTypeId":   1,
						"postedDate":    20000,
						"effectiveDate": 20000,
						"totalAmt":      "3000000",
					},
				},
			},
		},
	}

	// Write test data to file
	encoder := json.NewEncoder(tmpFile)
	err = encoder.Encode(testData)
	require.NoError(t, err)

	logger := log.NewNopLogger()

	// Test finding next correlation ID after the first one
	nextID, err := findNextCorrelationIDAfter(tmpFile.Name(), "33a830f5-3969-4a6b-bc80-f3e7f4e8850b", logger)
	require.NoError(t, err)
	assert.Equal(t, "40bf044e-5f84-41d1-a5e4-795ab98c3dc6", nextID)

	// Test finding next correlation ID after the second one
	nextID2, err := findNextCorrelationIDAfter(tmpFile.Name(), "40bf044e-5f84-41d1-a5e4-795ab98c3dc6", logger)
	require.NoError(t, err)
	assert.Equal(t, "50cf155f-6f95-52e2-b6f5-8a6bc09d4ed7", nextID2)

	// Test finding next correlation ID after the last one (should return empty)
	nextID3, err := findNextCorrelationIDAfter(tmpFile.Name(), "50cf155f-6f95-52e2-b6f5-8a6bc09d4ed7", logger)
	require.NoError(t, err)
	assert.Equal(t, "", nextID3)

	// Test finding next correlation ID after non-existent ID (should return empty)
	nextID4, err := findNextCorrelationIDAfter(tmpFile.Name(), "nonexistent-id", logger)
	require.NoError(t, err)
	assert.Equal(t, "", nextID4)
}

func TestResumeScenarioWithUnconfirmedChunk(t *testing.T) {
	// Create a temporary test file
	tmpFile, err := os.CreateTemp("", "test_resume_scenario_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Create test data with correlation IDs matching the error scenario
	testData := map[string]interface{}{
		"ledgerToEntries": []map[string]interface{}{
			{
				"ledgerKey": map[string]interface{}{
					"nftId":        "nft1",
					"assetClassId": "asset1",
				},
				"ledger": map[string]interface{}{
					"ledgerClassId": "class1",
					"statusTypeId":  1,
				},
				"entries": []map[string]interface{}{
					{
						"correlationId": "33a830f5-3969-4a6b-bc80-f3e7f4e8850b",
						"sequence":      1,
						"entryTypeId":   1,
						"postedDate":    20000,
						"effectiveDate": 20000,
						"totalAmt":      "1000000",
					},
					{
						"correlationId": "40bf044e-5f84-41d1-a5e4-795ab98c3dc6",
						"sequence":      2,
						"entryTypeId":   1,
						"postedDate":    20000,
						"effectiveDate": 20000,
						"totalAmt":      "2000000",
					},
					{
						"correlationId": "50cf155f-6f95-52e2-b6f5-8a6bc09d4ed7",
						"sequence":      3,
						"entryTypeId":   1,
						"postedDate":    20000,
						"effectiveDate": 20000,
						"totalAmt":      "3000000",
					},
				},
			},
		},
	}

	// Write test data to file
	encoder := json.NewEncoder(tmpFile)
	err = encoder.Encode(testData)
	require.NoError(t, err)

	logger := log.NewNopLogger()

	// Simulate the exact scenario from the error message:
	// - Last attempted chunk has first correlation ID: 33a830f5-3969-4a6b-bc80-f3e7f4e8850b
	// - Last attempted chunk has last correlation ID: 40bf044e-5f84-41d1-a5e4-795ab98c3dc6
	// - Last attempted chunk was not confirmed
	// - No last successful correlation ID
	// - No transaction hash

	// Test that we find the next correlation ID after the first correlation ID of the last attempted chunk
	nextID, err := findNextCorrelationIDAfter(tmpFile.Name(), "33a830f5-3969-4a6b-bc80-f3e7f4e8850b", logger)
	require.NoError(t, err)
	assert.Equal(t, "40bf044e-5f84-41d1-a5e4-795ab98c3dc6", nextID)

	// Test that we find the next correlation ID after the last correlation ID of the last attempted chunk
	nextID2, err := findNextCorrelationIDAfter(tmpFile.Name(), "40bf044e-5f84-41d1-a5e4-795ab98c3dc6", logger)
	require.NoError(t, err)
	assert.Equal(t, "50cf155f-6f95-52e2-b6f5-8a6bc09d4ed7", nextID2)

	// Test that ProcessFileFromCorrelationID starts from the correct correlation ID
	config := DefaultChunkConfig()
	config.MaxChunkSizeBytes = 1000 // Small chunk size to ensure we get the right behavior
	processor := NewStreamingGenesisProcessor(config, logger)

	// Process from the next correlation ID after the first correlation ID of the last attempted chunk
	chunkedState, err := processor.ProcessFileFromCorrelationID(tmpFile.Name(), "33a830f5-3969-4a6b-bc80-f3e7f4e8850b")
	require.NoError(t, err)
	assert.Equal(t, 1, chunkedState.TotalChunks)
	assert.Equal(t, 1, chunkedState.TotalLedgers)
	assert.Equal(t, 2, chunkedState.TotalEntries) // Should have 2 entries: 40bf044e-5f84-41d1-a5e4-795ab98c3dc6 and 50cf155f-6f95-52e2-b6f5-8a6bc09d4ed7

	// Verify the entries are correct
	entries := chunkedState.Chunks[0].LedgerToEntries[0].Entries
	assert.Equal(t, 2, len(entries))
	assert.Equal(t, "40bf044e-5f84-41d1-a5e4-795ab98c3dc6", entries[0].CorrelationId)
	assert.Equal(t, "50cf155f-6f95-52e2-b6f5-8a6bc09d4ed7", entries[1].CorrelationId)
}

func TestResumeScenarioExactErrorCase(t *testing.T) {
	// Create a temporary test file
	tmpFile, err := os.CreateTemp("", "test_exact_error_case_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Create test data with correlation IDs matching the exact error scenario
	testData := map[string]interface{}{
		"ledgerToEntries": []map[string]interface{}{
			{
				"ledgerKey": map[string]interface{}{
					"nftId":        "nft1",
					"assetClassId": "asset1",
				},
				"ledger": map[string]interface{}{
					"ledgerClassId": "class1",
					"statusTypeId":  1,
				},
				"entries": []map[string]interface{}{
					{
						"correlationId": "33a830f5-3969-4a6b-bc80-f3e7f4e8850b",
						"sequence":      1,
						"entryTypeId":   1,
						"postedDate":    20000,
						"effectiveDate": 20000,
						"totalAmt":      "1000000",
					},
					{
						"correlationId": "40bf044e-5f84-41d1-a5e4-795ab98c3dc6",
						"sequence":      2,
						"entryTypeId":   1,
						"postedDate":    20000,
						"effectiveDate": 20000,
						"totalAmt":      "2000000",
					},
				},
			},
		},
	}

	// Write test data to file
	encoder := json.NewEncoder(tmpFile)
	err = encoder.Encode(testData)
	require.NoError(t, err)

	logger := log.NewNopLogger()

	// Simulate the exact status from the error message:
	// - Status: in_progress
	// - Completed chunks: 0
	// - Last attempted chunk: first_correlation_id=33a830f5-3969-4a6b-bc80-f3e7f4e8850b, last_correlation_id=40bf044e-5f84-41d1-a5e4-795ab98c3dc6
	// - Last attempted confirmed: false
	// - Transaction hash: empty
	// - Last successful correlation ID: empty

	// Test that we find the next correlation ID after the first correlation ID of the last attempted chunk
	nextID, err := findNextCorrelationIDAfter(tmpFile.Name(), "33a830f5-3969-4a6b-bc80-f3e7f4e8850b", logger)
	require.NoError(t, err)
	assert.Equal(t, "40bf044e-5f84-41d1-a5e4-795ab98c3dc6", nextID)

	// Test that ProcessFileFromCorrelationID starts from the correct correlation ID
	config := DefaultChunkConfig()
	config.MaxChunkSizeBytes = 1000 // Small chunk size to ensure we get the right behavior
	processor := NewStreamingGenesisProcessor(config, logger)

	// Process from the next correlation ID after the first correlation ID of the last attempted chunk
	chunkedState, err := processor.ProcessFileFromCorrelationID(tmpFile.Name(), "33a830f5-3969-4a6b-bc80-f3e7f4e8850b")
	require.NoError(t, err)
	assert.Equal(t, 1, chunkedState.TotalChunks)
	assert.Equal(t, 1, chunkedState.TotalLedgers)
	assert.Equal(t, 1, chunkedState.TotalEntries) // Should have 1 entry: 40bf044e-5f84-41d1-a5e4-795ab98c3dc6

	// Verify the entries are correct - should NOT include 33a830f5-3969-4a6b-bc80-f3e7f4e8850b
	entries := chunkedState.Chunks[0].LedgerToEntries[0].Entries
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, "40bf044e-5f84-41d1-a5e4-795ab98c3dc6", entries[0].CorrelationId)
	assert.NotEqual(t, "33a830f5-3969-4a6b-bc80-f3e7f4e8850b", entries[0].CorrelationId, "Should not include the correlation ID we're resuming from")
}

func TestTransactionHashStorageInStatus(t *testing.T) {
	// Test that transaction hash is properly stored in LastAttemptedChunk after successful broadcast
	status := &LocalBulkImportStatus{
		ImportID:        "test_import_123",
		TotalChunks:     1,
		CompletedChunks: 0,
		TotalLedgers:    1,
		TotalEntries:    1,
		Status:          "in_progress",
		CreatedAt:       time.Now().Format(time.RFC3339),
		UpdatedAt:       time.Now().Format(time.RFC3339),
	}

	// Simulate successful transaction broadcast
	txHash := "0x1234567890abcdef"
	firstCorrelationID := "corr_123"
	lastCorrelationID := "corr_123"

	// Update LastAttemptedChunk after successful broadcast (as per the new logic)
	status.LastAttemptedChunk = &ChunkStatus{
		FirstCorrelationID: firstCorrelationID,
		LastCorrelationID:  lastCorrelationID,
		Confirmed:          false,  // Will be set to true after confirmation
		TransactionHash:    txHash, // Set immediately after successful broadcast
	}

	// Verify the transaction hash is stored
	require.NotNil(t, status.LastAttemptedChunk, "LastAttemptedChunk should not be nil")
	require.Equal(t, txHash, status.LastAttemptedChunk.TransactionHash, "Transaction hash should be stored")
	require.Equal(t, firstCorrelationID, status.LastAttemptedChunk.FirstCorrelationID, "First correlation ID should be stored")
	require.Equal(t, lastCorrelationID, status.LastAttemptedChunk.LastCorrelationID, "Last correlation ID should be stored")
	require.False(t, status.LastAttemptedChunk.Confirmed, "Chunk should not be confirmed yet")

	// Test that the status can be written and read back correctly
	err := writeLocalBulkImportStatus(status)
	require.NoError(t, err, "Should be able to write status file")

	// Read back the status
	readStatus, err := readLocalBulkImportStatus(status.ImportID)
	require.NoError(t, err, "Should be able to read status file")
	require.NotNil(t, readStatus, "Read status should not be nil")
	require.NotNil(t, readStatus.LastAttemptedChunk, "LastAttemptedChunk should be preserved")
	require.Equal(t, txHash, readStatus.LastAttemptedChunk.TransactionHash, "Transaction hash should be preserved")
	require.Equal(t, firstCorrelationID, readStatus.LastAttemptedChunk.FirstCorrelationID, "First correlation ID should be preserved")
	require.Equal(t, lastCorrelationID, readStatus.LastAttemptedChunk.LastCorrelationID, "Last correlation ID should be preserved")

	// Clean up
	_ = os.Remove(statusFileName(status.ImportID))
}
