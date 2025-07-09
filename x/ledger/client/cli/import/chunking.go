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

	var optimizedChunks []*types.GenesisState

	for _, chunk := range p.chunks {
		// Estimate gas for this chunk using our cost model
		estimatedGas := estimateChunkGasFromCosts(chunk, gasCosts)

		p.logger.Info("Processing chunk for optimization",
			"ledger_count", len(chunk.LedgerToEntries),
			"estimated_gas", estimatedGas,
			"max_gas", p.config.MaxGasPerTx-100000)

		// If the chunk fits within gas limits, keep it as-is
		if estimatedGas <= p.config.MaxGasPerTx-100000 {
			optimizedChunks = append(optimizedChunks, chunk)
			p.logger.Info("Chunk fits within gas limits, keeping as-is")
			continue
		}

		// If the chunk is too large, split it into smaller chunks
		p.logger.Info("Chunk exceeds gas limit, splitting into smaller chunks",
			"estimated_gas", estimatedGas,
			"max_gas", p.config.MaxGasPerTx-100000,
			"ledger_count", len(chunk.LedgerToEntries))

		// Split the chunk using our cost model
		splitChunks := p.splitChunkByCostModel(chunk, gasCosts)
		optimizedChunks = append(optimizedChunks, splitChunks...)
	}

	p.chunks = optimizedChunks
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

	p.logger.Info("Calculated optimal chunk size",
		"max_gas_per_chunk", maxGasPerChunk,
		"base_gas", baseGas,
		"gas_per_entry", gasPerEntry,
		"max_entries_per_chunk", maxEntriesPerChunk,
		"total_entries", len(lte.Entries))

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

		p.logger.Info("Created chunk",
			"chunk_index", len(result),
			"entries_count", len(chunkEntries),
			"estimated_gas", estimateChunkGasFromCosts(chunk, costs),
			"has_ledger", chunkLedger != nil)
	}

	return result
}
