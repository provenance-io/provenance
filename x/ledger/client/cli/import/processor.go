package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"cosmossdk.io/log"

	"github.com/provenance-io/provenance/x/ledger/types"
)

// StreamingChunkProcessor handles streaming JSON parsing and yields chunks one at a time
type StreamingChunkProcessor struct {
	config ChunkConfig
	stats  *ImportStats
	logger log.Logger

	// Streaming state
	file    *os.File
	decoder *json.Decoder
	reader  *bufio.Reader

	// Current chunk state
	currentChunk *types.GenesisState
	ledgerCount  int
	entryCount   int

	// Resume state
	skipUntilFound          bool
	foundStartCorrelationID bool
	startFromCorrelationID  string

	// End of stream flag
	atEnd bool
}

// NewStreamingChunkProcessor creates a new streaming chunk processor
func NewStreamingChunkProcessor(config ChunkConfig, logger log.Logger) *StreamingChunkProcessor {
	return &StreamingChunkProcessor{
		config: config,
		stats:  &ImportStats{},
		logger: logger,
	}
}

// OpenFile opens a genesis file for streaming processing
func (p *StreamingChunkProcessor) OpenFile(filename string, startFromCorrelationID string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open genesis state file: %w", err)
	}

	p.file = file
	p.reader = bufio.NewReader(file)
	p.decoder = json.NewDecoder(p.reader)
	p.startFromCorrelationID = startFromCorrelationID
	p.skipUntilFound = startFromCorrelationID != ""
	p.foundStartCorrelationID = false
	p.atEnd = false

	// Get file size for progress reporting
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	if startFromCorrelationID != "" {
		p.logger.Info("Opening genesis file for streaming from correlation ID", "filename", filename, "size_bytes", fileInfo.Size(), "start_correlation_id", startFromCorrelationID)
	} else {
		p.logger.Info("Opening genesis file for streaming", "filename", filename, "size_bytes", fileInfo.Size())
	}

	// Initialize JSON parsing
	return p.initializeJSONParsing()
}

// Close closes the file and cleans up resources
func (p *StreamingChunkProcessor) Close() error {
	if p.file != nil {
		return p.file.Close()
	}
	return nil
}

// initializeJSONParsing sets up the JSON decoder to the ledgerToEntries array
func (p *StreamingChunkProcessor) initializeJSONParsing() error {
	// Expect the root object
	token, err := p.decoder.Token()
	if err != nil {
		return fmt.Errorf("failed to read JSON token: %w", err)
	}

	if delim, ok := token.(json.Delim); !ok || delim != '{' {
		return fmt.Errorf("expected JSON object, got %v", token)
	}

	// Process the root object to find ledgerToEntries
	for p.decoder.More() {
		token, err := p.decoder.Token()
		if err != nil {
			return fmt.Errorf("failed to read field name: %w", err)
		}

		fieldName, ok := token.(string)
		if !ok {
			return fmt.Errorf("expected field name, got %v", token)
		}

		if fieldName == "ledgerToEntries" || fieldName == "ledger_to_entries" {
			// Expect array start
			token, err := p.decoder.Token()
			if err != nil {
				return fmt.Errorf("failed to read array start: %w", err)
			}

			if delim, ok := token.(json.Delim); !ok || delim != '[' {
				return fmt.Errorf("expected array start, got %v", token)
			}

			// Initialize current chunk
			p.currentChunk = &types.GenesisState{
				LedgerToEntries: []types.LedgerToEntries{},
			}

			return nil
		} else {
			// Skip unknown fields
			if err := p.skipValue(p.decoder); err != nil {
				return fmt.Errorf("failed to skip field %s: %w", fieldName, err)
			}
		}
	}

	return fmt.Errorf("ledgerToEntries field not found in JSON")
}

// NextChunkWithGasLimit reads the next chunk from the stream with gas limit enforcement
func (p *StreamingChunkProcessor) NextChunkWithGasLimit(gasCosts *GasCosts) (*types.GenesisState, error) {
	if p.atEnd {
		return nil, io.EOF
	}

	// Process ledger entries until we have a complete chunk
	for p.decoder.More() {
		var lte types.LedgerToEntries
		if err := p.decoder.Decode(&lte); err != nil {
			return nil, fmt.Errorf("failed to decode LedgerToEntries: %w", err)
		}

		// Validate the ledger entry
		if err := p.validateLedgerToEntries(&lte); err != nil {
			return nil, fmt.Errorf("validation failed: %w", err)
		}

		// Handle resume functionality - skip until we find the start correlation ID
		if p.skipUntilFound {
			// Check if any entry in this ledger has the start correlation ID
			for _, entry := range lte.Entries {
				if entry.CorrelationId == p.startFromCorrelationID {
					p.foundStartCorrelationID = true
					p.skipUntilFound = false
					break
				}
			}

			if !p.foundStartCorrelationID {
				// Skip this entire ledger entry
				continue
			} else {
				// We found the start correlation ID, but we need to filter out entries before and including it
				var filteredEntries []*types.LedgerEntry
				startFound := false

				for _, entry := range lte.Entries {
					if !startFound && entry.CorrelationId == p.startFromCorrelationID {
						startFound = true
						// Skip this entry (the last processed one) and continue to the next
						continue
					}
					if startFound {
						filteredEntries = append(filteredEntries, entry)
					}
				}

				// Update the ledger entry with filtered entries
				lte.Entries = filteredEntries
			}
		}

		// Check gas limit BEFORE adding the entry to the chunk
		if gasCosts != nil {
			// Create a temporary chunk with the new entry to estimate gas
			tempChunk := &types.GenesisState{
				LedgerToEntries: append(p.currentChunk.LedgerToEntries, lte),
			}
			estimatedGas := estimateChunkGasFromCosts(tempChunk, gasCosts)
			maxGasPerChunk := p.config.GetEffectiveGasLimit()

			if estimatedGas > maxGasPerChunk {
				// Return the current chunk without adding the new entry
				if len(p.currentChunk.LedgerToEntries) > 0 {
					chunkToReturn := p.currentChunk
					// Start a new chunk with the current entry
					p.currentChunk = &types.GenesisState{
						LedgerToEntries: []types.LedgerToEntries{lte},
					}
					return chunkToReturn, nil
				}
			}
		}

		// Add to current chunk
		p.currentChunk.LedgerToEntries = append(p.currentChunk.LedgerToEntries, lte)

		// Check if we should start a new chunk based on transaction size limit
		chunkSize := getChunkSizeBytes(p.currentChunk)
		if chunkSize > p.config.MaxTxSizeBytes {
			// Remove the last entry and return the current chunk
			p.currentChunk.LedgerToEntries = p.currentChunk.LedgerToEntries[:len(p.currentChunk.LedgerToEntries)-1]

			// Create the chunk to return
			chunkToReturn := p.currentChunk

			// Start a new chunk with the current entry
			p.currentChunk = &types.GenesisState{
				LedgerToEntries: []types.LedgerToEntries{lte},
			}

			// Update stats
			p.stats.TotalLedgers++
			p.ledgerCount++
			if lte.Entries != nil {
				p.stats.TotalEntries += len(lte.Entries)
				p.entryCount += len(lte.Entries)
			}

			return chunkToReturn, nil
		}

		// Update stats
		p.stats.TotalLedgers++
		p.ledgerCount++
		if lte.Entries != nil {
			p.stats.TotalEntries += len(lte.Entries)
			p.entryCount += len(lte.Entries)
		}

		// Periodic progress log
		if p.ledgerCount%1000 == 0 {
			p.logger.Debug("StreamingChunkProcessor: progress", "ledger_count", p.ledgerCount, "entry_count", p.entryCount)
		}
	}

	// We've reached the end of the array
	p.atEnd = true

	// Return the final chunk if it has data
	if len(p.currentChunk.LedgerToEntries) > 0 {
		finalChunk := p.currentChunk
		p.currentChunk = nil // Clear to prevent reuse
		return finalChunk, nil
	}

	return nil, io.EOF
}

// GetStats returns the current import statistics
func (p *StreamingChunkProcessor) GetStats() *ImportStats {
	return p.stats
}

// validateLedgerToEntries validates a single LedgerToEntries object
func (p *StreamingChunkProcessor) validateLedgerToEntries(lte *types.LedgerToEntries) error {
	if lte.LedgerKey == nil {
		return fmt.Errorf("ledger key is nil")
	}

	if lte.LedgerKey.NftId == "" {
		return fmt.Errorf("ledger key NftId is empty")
	}

	if lte.LedgerKey.AssetClassId == "" {
		return fmt.Errorf("ledger key AssetClassId is empty")
	}

	if lte.Entries != nil {
		for i, entry := range lte.Entries {
			if entry == nil {
				return fmt.Errorf("entry at index %d is nil", i)
			}
		}
	}

	return nil
}

// skipValue skips a JSON value (object, array, string, number, boolean, null)
func (p *StreamingChunkProcessor) skipValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}

	switch token {
	case json.Delim('{'):
		// Skip object
		for decoder.More() {
			// Skip key
			if _, err := decoder.Token(); err != nil {
				return err
			}
			// Skip value
			if err := p.skipValue(decoder); err != nil {
				return err
			}
		}
		// Skip closing brace
		if _, err := decoder.Token(); err != nil {
			return err
		}
	case json.Delim('['):
		// Skip array
		for decoder.More() {
			if err := p.skipValue(decoder); err != nil {
				return err
			}
		}
		// Skip closing bracket
		if _, err := decoder.Token(); err != nil {
			return err
		}
	case json.Delim('}'), json.Delim(']'):
		// These should not be encountered here
		return fmt.Errorf("unexpected delimiter: %v", token)
	default:
		// String, number, boolean, null - already consumed
	}

	return nil
}