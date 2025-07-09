package cli

import (
	"os"
	"testing"

	"cosmossdk.io/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessFileFromCorrelationID(t *testing.T) {
	// Create a temporary test file with sample data
	testData := `{
		"ledgerToEntries": [
			{
				"ledgerKey": {
					"nftId": "nft1",
					"assetClassId": "asset1"
				},
				"ledger": {
					"ledgerClassId": "class1",
					"statusTypeId": 1
				},
				"entries": [
					{
						"correlationId": "corr1",
						"sequence": 1,
						"entryTypeId": 1,
						"postedDate": 20000,
						"effectiveDate": 20000,
						"totalAmt": "1000000"
					},
					{
						"correlationId": "corr2",
						"sequence": 2,
						"entryTypeId": 1,
						"postedDate": 20000,
						"effectiveDate": 20000,
						"totalAmt": "2000000"
					}
				]
			},
			{
				"ledgerKey": {
					"nftId": "nft2",
					"assetClassId": "asset2"
				},
				"ledger": {
					"ledgerClassId": "class2",
					"statusTypeId": 1
				},
				"entries": [
					{
						"correlationId": "corr3",
						"sequence": 1,
						"entryTypeId": 1,
						"postedDate": 20000,
						"effectiveDate": 20000,
						"totalAmt": "3000000"
					},
					{
						"correlationId": "corr4",
						"sequence": 2,
						"entryTypeId": 1,
						"postedDate": 20000,
						"effectiveDate": 20000,
						"totalAmt": "4000000"
					}
				]
			}
		]
	}`

	tmpFile, err := os.CreateTemp("", "test_genesis_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(testData)
	require.NoError(t, err)
	tmpFile.Close()

	config := ChunkConfig{
		MaxChunkSizeBytes: 1000,
		MaxTxSizeBytes:    1000,
		MaxGasPerTx:       1000000,
	}
	logger := log.NewNopLogger()

	processor := NewStreamingGenesisProcessor(config, logger)

	// Test 1: Process entire file
	chunkedState, err := processor.ProcessFile(tmpFile.Name())
	require.NoError(t, err)
	assert.Equal(t, 1, chunkedState.TotalChunks)
	assert.Equal(t, 2, chunkedState.TotalLedgers)
	assert.Equal(t, 4, chunkedState.TotalEntries)

	// Test 2: Process from specific correlation ID (resume behavior - starts AFTER the specified ID)
	processor2 := NewStreamingGenesisProcessor(config, logger)
	chunkedState2, err := processor2.ProcessFileFromCorrelationID(tmpFile.Name(), "corr3")
	require.NoError(t, err)
	assert.Equal(t, 1, chunkedState2.TotalChunks)
	assert.Equal(t, 1, chunkedState2.TotalLedgers) // Should only have ledger2
	assert.Equal(t, 1, chunkedState2.TotalEntries) // Should only have corr4 (after corr3)

	// Verify the entries are filtered correctly
	entries := chunkedState2.Chunks[0].LedgerToEntries[0].Entries
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, "corr4", entries[0].CorrelationId)

	// Test 3: Process from non-existent correlation ID
	processor3 := NewStreamingGenesisProcessor(config, logger)
	chunkedState3, err := processor3.ProcessFileFromCorrelationID(tmpFile.Name(), "nonexistent")
	require.NoError(t, err)
	assert.Equal(t, 0, chunkedState3.TotalChunks)
	assert.Equal(t, 0, chunkedState3.TotalLedgers)
	assert.Equal(t, 0, chunkedState3.TotalEntries)
}

func TestProcessFileFromCorrelationIDWithPartialLedger(t *testing.T) {
	// Test case where a chunk contains the remainder of one ledger and part of another
	testData := `{
		"ledgerToEntries": [
			{
				"ledgerKey": {
					"nftId": "nft1",
					"assetClassId": "asset1"
				},
				"ledger": {
					"ledgerClassId": "class1",
					"statusTypeId": 1
				},
				"entries": [
					{
						"correlationId": "corr1",
						"sequence": 1,
						"entryTypeId": 1,
						"postedDate": 20000,
						"effectiveDate": 20000,
						"totalAmt": "1000000"
					},
					{
						"correlationId": "corr2",
						"sequence": 2,
						"entryTypeId": 1,
						"postedDate": 20000,
						"effectiveDate": 20000,
						"totalAmt": "2000000"
					},
					{
						"correlationId": "corr3",
						"sequence": 3,
						"entryTypeId": 1,
						"postedDate": 20000,
						"effectiveDate": 20000,
						"totalAmt": "3000000"
					}
				]
			},
			{
				"ledgerKey": {
					"nftId": "nft2",
					"assetClassId": "asset2"
				},
				"ledger": {
					"ledgerClassId": "class2",
					"statusTypeId": 1
				},
				"entries": [
					{
						"correlationId": "corr4",
						"sequence": 1,
						"entryTypeId": 1,
						"postedDate": 20000,
						"effectiveDate": 20000,
						"totalAmt": "4000000"
					},
					{
						"correlationId": "corr5",
						"sequence": 2,
						"entryTypeId": 1,
						"postedDate": 20000,
						"effectiveDate": 20000,
						"totalAmt": "5000000"
					}
				]
			}
		]
	}`

	tmpFile, err := os.CreateTemp("", "test_genesis_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(testData)
	require.NoError(t, err)
	tmpFile.Close()

	config := ChunkConfig{
		MaxChunkSizeBytes: 1000,
		MaxTxSizeBytes:    1000,
		MaxGasPerTx:       1000000,
	}
	logger := log.NewNopLogger()

	processor := NewStreamingGenesisProcessor(config, logger)

	// Test resuming from middle of ledger1 (resume behavior - starts AFTER the specified ID)
	chunkedState, err := processor.ProcessFileFromCorrelationID(tmpFile.Name(), "corr2")
	require.NoError(t, err)
	assert.Equal(t, 1, chunkedState.TotalChunks)
	assert.Equal(t, 2, chunkedState.TotalLedgers) // Should have both ledgers
	assert.Equal(t, 3, chunkedState.TotalEntries) // Should have corr3, corr4, corr5 (after corr2)

	// Verify the entries are filtered correctly
	entries := chunkedState.Chunks[0].LedgerToEntries[0].Entries
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, "corr3", entries[0].CorrelationId)

	entries2 := chunkedState.Chunks[0].LedgerToEntries[1].Entries
	assert.Equal(t, 2, len(entries2))
	assert.Equal(t, "corr4", entries2[0].CorrelationId)
	assert.Equal(t, "corr5", entries2[1].CorrelationId)
}
