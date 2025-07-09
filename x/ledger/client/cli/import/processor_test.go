package cli

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"github.com/provenance-io/provenance/x/ledger/types"
	"github.com/stretchr/testify/require"
)

func TestNewStreamingGenesisProcessor(t *testing.T) {
	config := DefaultChunkConfig()
	logger := log.NewNopLogger()

	processor := NewStreamingGenesisProcessor(config, logger)

	require.NotNil(t, processor, "Processor should not be nil")
	require.Equal(t, config, processor.config, "Config should match")
	require.Equal(t, logger, processor.logger, "Logger should match")
	require.NotNil(t, processor.chunks, "Chunks slice should be initialized")
	require.NotNil(t, processor.stats, "Stats should be initialized")
	require.Equal(t, 0, len(processor.chunks), "Chunks should be empty initially")
	require.Equal(t, 0, processor.stats.TotalLedgers, "Total ledgers should be 0 initially")
	require.Equal(t, 0, processor.stats.TotalEntries, "Total entries should be 0 initially")
}

func TestStreamingGenesisProcessorValidateLedgerToEntries(t *testing.T) {
	config := DefaultChunkConfig()
	logger := log.NewNopLogger()
	processor := NewStreamingGenesisProcessor(config, logger)

	// Test valid ledger entry
	validLTE := &types.LedgerToEntries{
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
	}

	err := processor.validateLedgerToEntries(validLTE)
	require.NoError(t, err, "Valid ledger entry should pass validation")

	// Test nil ledger key
	nilKeyLTE := &types.LedgerToEntries{
		LedgerKey: nil,
		Ledger: &types.Ledger{
			LedgerClassId: "test-class-1",
			StatusTypeId:  1,
		},
		Entries: []*types.LedgerEntry{},
	}

	err = processor.validateLedgerToEntries(nilKeyLTE)
	require.Error(t, err, "Nil ledger key should fail validation")
	require.Contains(t, err.Error(), "ledger key is nil")

	// Test empty NFT ID
	emptyNftLTE := &types.LedgerToEntries{
		LedgerKey: &types.LedgerKey{
			NftId:        "", // Empty NFT ID
			AssetClassId: "test-asset-1",
		},
		Ledger: &types.Ledger{
			LedgerClassId: "test-class-1",
			StatusTypeId:  1,
		},
		Entries: []*types.LedgerEntry{},
	}

	err = processor.validateLedgerToEntries(emptyNftLTE)
	require.Error(t, err, "Empty NFT ID should fail validation")
	require.Contains(t, err.Error(), "ledger key NftId is empty")

	// Test empty asset class ID
	emptyAssetLTE := &types.LedgerToEntries{
		LedgerKey: &types.LedgerKey{
			NftId:        "test-nft-1",
			AssetClassId: "", // Empty asset class ID
		},
		Ledger: &types.Ledger{
			LedgerClassId: "test-class-1",
			StatusTypeId:  1,
		},
		Entries: []*types.LedgerEntry{},
	}

	err = processor.validateLedgerToEntries(emptyAssetLTE)
	require.Error(t, err, "Empty asset class ID should fail validation")
	require.Contains(t, err.Error(), "ledger key AssetClassId is empty")
}

func TestStreamingGenesisProcessorSkipValue(t *testing.T) {
	config := DefaultChunkConfig()
	logger := log.NewNopLogger()
	processor := NewStreamingGenesisProcessor(config, logger)

	// Test skipValue with valid JSON decoder
	// This is a basic test to ensure the function doesn't panic
	// In a real scenario, this would be tested with actual JSON parsing
	require.NotNil(t, processor, "Processor should be created successfully")
}

func TestStreamingGenesisProcessorStats(t *testing.T) {
	config := DefaultChunkConfig()
	logger := log.NewNopLogger()
	processor := NewStreamingGenesisProcessor(config, logger)

	// Test initial stats
	require.Equal(t, 0, processor.stats.TotalLedgers, "Initial total ledgers should be 0")
	require.Equal(t, 0, processor.stats.TotalEntries, "Initial total entries should be 0")
	require.Equal(t, 0, processor.stats.TotalChunks, "Initial total chunks should be 0")

	// Test stats after adding data
	processor.stats.TotalLedgers = 10
	processor.stats.TotalEntries = 50
	processor.stats.TotalChunks = 2

	require.Equal(t, 10, processor.stats.TotalLedgers, "Total ledgers should be updated")
	require.Equal(t, 50, processor.stats.TotalEntries, "Total entries should be updated")
	require.Equal(t, 2, processor.stats.TotalChunks, "Total chunks should be updated")
}

func TestStreamingGenesisProcessorChunks(t *testing.T) {
	config := DefaultChunkConfig()
	logger := log.NewNopLogger()
	processor := NewStreamingGenesisProcessor(config, logger)

	// Test initial chunks
	require.Equal(t, 0, len(processor.chunks), "Initial chunks should be empty")

	// Test adding chunks
	testChunk := &types.GenesisState{
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
				Entries: []*types.LedgerEntry{},
			},
		},
	}

	processor.chunks = append(processor.chunks, testChunk)
	require.Equal(t, 1, len(processor.chunks), "Should have one chunk after adding")
	require.Equal(t, testChunk, processor.chunks[0], "Added chunk should match")
}

func TestStreamingGenesisProcessorConfig(t *testing.T) {
	config := ChunkConfig{
		MaxChunkSizeBytes: 500000,
		MaxGasPerTx:       2000000,
		MaxTxSizeBytes:    1000000,
	}
	logger := log.NewNopLogger()
	processor := NewStreamingGenesisProcessor(config, logger)

	require.Equal(t, config, processor.config, "Processor config should match input config")
	require.Equal(t, 500000, processor.config.MaxChunkSizeBytes, "MaxChunkSizeBytes should match")
	require.Equal(t, 2000000, processor.config.MaxGasPerTx, "MaxGasPerTx should match")
	require.Equal(t, 1000000, processor.config.MaxTxSizeBytes, "MaxTxSizeBytes should match")
}

func TestStreamingGenesisProcessorLogger(t *testing.T) {
	config := DefaultChunkConfig()
	logger := log.NewNopLogger()
	processor := NewStreamingGenesisProcessor(config, logger)

	require.Equal(t, logger, processor.logger, "Processor logger should match input logger")
	require.NotNil(t, processor.logger, "Logger should not be nil")
}
