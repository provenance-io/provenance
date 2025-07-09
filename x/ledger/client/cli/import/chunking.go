package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"

	"github.com/provenance-io/provenance/x/ledger/types"
)

// optimizeChunksUsingSimulation takes the initial chunks from parsing and optimizes them
// using simulation to ensure they fit within gas limits
func (p *StreamingGenesisProcessor) optimizeChunksUsingSimulation(clientCtx client.Context, cmd *cobra.Command) error {
	// First, run a few representative simulations to understand gas costs
	gasCosts, err := estimateGasCosts(p.chunks, clientCtx, cmd, p.logger)
	if err != nil {
		return fmt.Errorf("failed to estimate gas costs: %w", err)
	}

	p.logger.Info("Gas cost estimates",
		"ledger_key_gas", gasCosts.LedgerWithKeyGas,
		"entry_gas", gasCosts.EntryGas)

	return p.optimizeChunksWithCosts(gasCosts)
}

// optimizeChunksUsingStoredCosts optimizes chunks using pre-stored gas costs for deterministic behavior
func (p *StreamingGenesisProcessor) optimizeChunksUsingStoredCosts(storedCosts *GasCosts) error {
	p.logger.Info("Using stored gas costs for optimization",
		"ledger_key_gas", storedCosts.LedgerWithKeyGas,
		"entry_gas", storedCosts.EntryGas)

	return p.optimizeChunksWithCosts(storedCosts)
}

// optimizeChunksWithCosts is the common optimization logic using provided gas costs
func (p *StreamingGenesisProcessor) optimizeChunksWithCosts(gasCosts *GasCosts) error {
	var optimizedChunks []*types.GenesisState
	inputChunkCount := len(p.chunks)

	p.logger.Info("Starting chunk optimization", "input_chunks", inputChunkCount)

	// Log progress every 10 chunks to avoid excessive logging
	logProgressEvery := 10
	if inputChunkCount > 100 {
		logProgressEvery = 50 // Less frequent logging for very large datasets
	}

	for i, chunk := range p.chunks {
		// Estimate gas for this chunk using our cost model
		estimatedGas := estimateChunkGasFromCosts(chunk, gasCosts)

		// Only log detailed info for every Nth chunk or when splitting is needed
		shouldLogDetails := (i+1)%logProgressEvery == 0 || estimatedGas > p.config.MaxGasPerTx-100000

		if shouldLogDetails {
			p.logger.Info("Processing chunk for optimization",
				"chunk_index", i+1,
				"ledger_count", len(chunk.LedgerToEntries),
				"estimated_gas", estimatedGas,
				"max_gas", p.config.MaxGasPerTx-100000)
		}

		// If the chunk fits within gas limits, keep it as-is
		if estimatedGas <= p.config.MaxGasPerTx-100000 {
			optimizedChunks = append(optimizedChunks, chunk)
			if shouldLogDetails {
				p.logger.Info("Chunk fits within gas limits, keeping as-is")
			}
			continue
		}

		// If the chunk is too large, split it into smaller chunks
		p.logger.Info("Chunk exceeds gas limit, splitting into smaller chunks",
			"chunk_index", i+1,
			"estimated_gas", estimatedGas,
			"max_gas", p.config.MaxGasPerTx-100000,
			"ledger_count", len(chunk.LedgerToEntries))

		// Split the chunk using our cost model
		splitChunks := p.splitChunkByCostModel(chunk, gasCosts)
		optimizedChunks = append(optimizedChunks, splitChunks...)
	}

	p.chunks = optimizedChunks
	p.logger.Info("Chunk optimization completed",
		"input_chunks", inputChunkCount,
		"output_chunks", len(optimizedChunks))
	return nil
}

// splitChunkByCostModel splits a chunk using the cost model to maximize utilization
func (p *StreamingGenesisProcessor) splitChunkByCostModel(chunk *types.GenesisState, costs *GasCosts) []*types.GenesisState {
	var result []*types.GenesisState

	// Group ledgers by whether they have the ledger object (first chunk vs subsequent chunks)
	var ledgersWithData []types.LedgerToEntries
	var ledgersWithoutData []types.LedgerToEntries

	for _, lte := range chunk.LedgerToEntries {
		if lte.Ledger != nil {
			ledgersWithData = append(ledgersWithData, lte)
		} else {
			ledgersWithoutData = append(ledgersWithoutData, lte)
		}
	}

	// Process ledgers with data first (they have higher gas cost)
	if len(ledgersWithData) > 0 {
		chunks := p.splitLedgersByCostModel(ledgersWithData, costs, true)
		result = append(result, chunks...)
	}

	// Process ledgers without data (lower gas cost, can fit more)
	if len(ledgersWithoutData) > 0 {
		chunks := p.splitLedgersByCostModel(ledgersWithoutData, costs, false)
		result = append(result, chunks...)
	}

	return result
}

// splitLedgersByCostModel splits ledgers using the cost model to maximize gas utilization
func (p *StreamingGenesisProcessor) splitLedgersByCostModel(ledgers []types.LedgerToEntries, costs *GasCosts, hasLedger bool) []*types.GenesisState {
	var result []*types.GenesisState
	maxGasPerChunk := p.config.MaxGasPerTx - 100000 // 100k safety margin

	var currentChunk []types.LedgerToEntries
	currentGas := 0

	for _, lte := range ledgers {
		// Calculate gas for this ledger
		ledgerGas := 0
		if hasLedger {
			ledgerGas = costs.LedgerWithKeyGas
		}
		if len(lte.Entries) > 0 {
			ledgerGas += len(lte.Entries) * costs.EntryGas
		}

		// Check if adding this ledger would exceed gas limit
		if currentGas+ledgerGas > maxGasPerChunk && len(currentChunk) > 0 {
			// Create chunk with current ledgers
			chunk := &types.GenesisState{
				LedgerToEntries: make([]types.LedgerToEntries, len(currentChunk)),
			}
			copy(chunk.LedgerToEntries, currentChunk)
			result = append(result, chunk)

			// Start new chunk
			currentChunk = []types.LedgerToEntries{}
			currentGas = 0
		}

		// If even a single ledger exceeds the limit, split it by entries
		if ledgerGas > maxGasPerChunk {
			// Add any existing ledgers to result first
			if len(currentChunk) > 0 {
				chunk := &types.GenesisState{
					LedgerToEntries: make([]types.LedgerToEntries, len(currentChunk)),
				}
				copy(chunk.LedgerToEntries, currentChunk)
				result = append(result, chunk)
				currentChunk = []types.LedgerToEntries{}
				currentGas = 0
			}

			// Split this large ledger by entries
			splitChunks := p.splitLargeLedgerByCostModel(&lte, costs)
			result = append(result, splitChunks...)
		} else {
			// Add to current chunk
			currentChunk = append(currentChunk, lte)
			currentGas += ledgerGas
		}
	}

	// Add remaining ledgers
	if len(currentChunk) > 0 {
		chunk := &types.GenesisState{
			LedgerToEntries: make([]types.LedgerToEntries, len(currentChunk)),
		}
		copy(chunk.LedgerToEntries, currentChunk)
		result = append(result, chunk)
	}

	return result
}

// splitLargeLedgerByCostModel splits a large ledger using the cost model
func (p *StreamingGenesisProcessor) splitLargeLedgerByCostModel(lte *types.LedgerToEntries, costs *GasCosts) []*types.GenesisState {
	var result []*types.GenesisState

	// Calculate how many entries we can fit in each chunk
	maxGasPerChunk := p.config.MaxGasPerTx - 100000 // 100k safety margin
	baseGas := costs.LedgerWithKeyGas
	gasPerEntry := costs.EntryGas

	// Calculate optimal entries per chunk
	// Formula: baseGas + (entries * gasPerEntry) <= maxGasPerChunk
	// So: entries <= (maxGasPerChunk - baseGas) / gasPerEntry
	maxEntriesPerChunk := (maxGasPerChunk - baseGas) / gasPerEntry

	// Ensure we don't have negative entries
	if maxEntriesPerChunk <= 0 {
		maxEntriesPerChunk = 1
	}

	p.logger.Info("Splitting large ledger",
		"total_entries", len(lte.Entries),
		"max_entries_per_chunk", maxEntriesPerChunk,
		"estimated_chunks", (len(lte.Entries)+maxEntriesPerChunk-1)/maxEntriesPerChunk)

	// Split entries into optimal chunks
	for i := 0; i < len(lte.Entries); i += maxEntriesPerChunk {
		end := i + maxEntriesPerChunk
		if end > len(lte.Entries) {
			end = len(lte.Entries)
		}

		chunkEntries := lte.Entries[i:end]
		isFirstChunk := i == 0

		// Only include ledger in the first chunk
		var chunkLedger *types.Ledger
		if isFirstChunk && lte.Ledger != nil {
			chunkLedger = lte.Ledger
		}

		chunk := &types.GenesisState{
			LedgerToEntries: []types.LedgerToEntries{
				{
					LedgerKey: lte.LedgerKey,
					Entries:   chunkEntries,
					Ledger:    chunkLedger,
				},
			},
		}

		// Verify the chunk fits within gas limits
		estimatedGas := estimateChunkGasFromCosts(chunk, costs)
		if estimatedGas > maxGasPerChunk {
			p.logger.Warn("Chunk exceeds gas limit, reducing entries",
				"estimated_gas", estimatedGas,
				"max_gas", maxGasPerChunk,
				"entries_count", len(chunkEntries))

			// If we still exceed the limit, reduce entries one by one
			for len(chunkEntries) > 1 && estimatedGas > maxGasPerChunk {
				chunkEntries = chunkEntries[:len(chunkEntries)-1]
				chunk.LedgerToEntries[0].Entries = chunkEntries
				estimatedGas = estimateChunkGasFromCosts(chunk, costs)
			}
		}

		result = append(result, chunk)

		// Only log for first and last chunks, or every 10th chunk for very large ledgers
		shouldLog := len(result) == 1 || i+maxEntriesPerChunk >= len(lte.Entries) || len(result)%10 == 0
		if shouldLog {
			p.logger.Info("Created split chunk",
				"chunk_index", len(result),
				"entries_count", len(chunkEntries),
				"estimated_gas", estimateChunkGasFromCosts(chunk, costs),
				"has_ledger", chunkLedger != nil)
		}
	}

	p.logger.Info("Finished splitting large ledger",
		"total_chunks_created", len(result),
		"total_entries", len(lte.Entries))

	return result
}
