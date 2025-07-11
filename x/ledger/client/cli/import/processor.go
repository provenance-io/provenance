package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"cosmossdk.io/log"

	"github.com/provenance-io/provenance/x/ledger/types"
)

// StreamingGenesisProcessor handles streaming JSON parsing for large genesis files
type StreamingGenesisProcessor struct {
	config ChunkConfig
	chunks []*types.GenesisState
	stats  *ImportStats
	logger log.Logger
}

// NewStreamingGenesisProcessor creates a new streaming processor
func NewStreamingGenesisProcessor(config ChunkConfig, logger log.Logger) *StreamingGenesisProcessor {
	return &StreamingGenesisProcessor{
		config: config,
		chunks: []*types.GenesisState{},
		stats:  &ImportStats{},
		logger: logger,
	}
}

// ProcessFile processes a genesis file using streaming JSON parsing
func (p *StreamingGenesisProcessor) ProcessFile(filename string) (*ChunkedGenesisState, error) {
	return p.ProcessFileFromCorrelationID(filename, "")
}

// ProcessFileFromCorrelationID processes a genesis file starting from a specific correlation ID
func (p *StreamingGenesisProcessor) ProcessFileFromCorrelationID(filename string, startFromCorrelationID string) (*ChunkedGenesisState, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open genesis state file: %w", err)
	}
	defer file.Close()

	// Get file size for progress reporting
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	if startFromCorrelationID != "" {
		p.logger.Info("Processing genesis file from correlation ID", "filename", filename, "size_bytes", fileInfo.Size(), "start_correlation_id", startFromCorrelationID)
	} else {
		p.logger.Info("Processing genesis file", "filename", filename, "size_bytes", fileInfo.Size())
	}

	// Use buffered reader for efficient reading
	reader := bufio.NewReader(file)

	// Parse the JSON structure using streaming
	err = p.parseStreamingJSON(reader, startFromCorrelationID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Generate import ID
	importID := generateImportID()

	return &ChunkedGenesisState{
		ImportID:     importID,
		TotalChunks:  len(p.chunks),
		Chunks:       p.chunks,
		TotalLedgers: p.stats.TotalLedgers,
		TotalEntries: p.stats.TotalEntries,
	}, nil
}

// ProcessFileWithStoredCosts processes a genesis file using stored gas costs for deterministic chunking
func (p *StreamingGenesisProcessor) ProcessFileWithStoredCosts(filename string, storedCosts *GasCosts) (*ChunkedGenesisState, error) {
	// First process the file normally
	chunkedState, err := p.ProcessFile(filename)
	if err != nil {
		return nil, err
	}

	// Then optimize chunks using stored gas costs instead of simulation
	p.logger.Info("Optimizing chunks using stored gas costs for deterministic chunking")
	err = p.optimizeChunksUsingStoredCosts(storedCosts)
	if err != nil {
		return nil, fmt.Errorf("failed to optimize chunks using stored costs: %w", err)
	}

	// Update chunkedState with optimized chunks
	chunkedState.Chunks = p.chunks
	chunkedState.TotalChunks = len(p.chunks)

	return chunkedState, nil
}

// parseStreamingJSON parses the JSON file using a streaming approach
func (p *StreamingGenesisProcessor) parseStreamingJSON(reader *bufio.Reader, startFromCorrelationID string) error {
	decoder := json.NewDecoder(reader)

	// Expect the root object
	token, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("failed to read JSON token: %w", err)
	}

	if delim, ok := token.(json.Delim); !ok || delim != '{' {
		return fmt.Errorf("expected JSON object, got %v", token)
	}

	// Process the root object
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return fmt.Errorf("failed to read field name: %w", err)
		}

		fieldName, ok := token.(string)
		if !ok {
			return fmt.Errorf("expected field name, got %v", token)
		}

		if fieldName == "ledgerToEntries" || fieldName == "ledger_to_entries" {
			err = p.parseLedgerToEntriesArray(decoder, startFromCorrelationID)
			if err != nil {
				return fmt.Errorf("failed to parse ledgerToEntries: %w", err)
			}
		} else {
			// Skip unknown fields
			if err := p.skipValue(decoder); err != nil {
				return fmt.Errorf("failed to skip field %s: %w", fieldName, err)
			}
		}
	}

	// Expect closing brace
	token, err = decoder.Token()
	if err != nil {
		return fmt.Errorf("failed to read closing brace: %w", err)
	}

	if delim, ok := token.(json.Delim); !ok || delim != '}' {
		return fmt.Errorf("expected closing brace, got %v", token)
	}

	return nil
}

// parseLedgerToEntriesArray parses the ledgerToEntries array
func (p *StreamingGenesisProcessor) parseLedgerToEntriesArray(decoder *json.Decoder, startFromCorrelationID string) error {
	// Expect array start
	token, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("failed to read array start: %w", err)
	}

	if delim, ok := token.(json.Delim); !ok || delim != '[' {
		return fmt.Errorf("expected array start, got %v", token)
	}

	currentChunk := &types.GenesisState{
		LedgerToEntries: []types.LedgerToEntries{},
	}

	// Track if we should start processing (for resume functionality)
	skipUntilFound := startFromCorrelationID != ""
	foundStartCorrelationID := false

	// Process each ledger entry
	for decoder.More() {
		var lte types.LedgerToEntries
		if err := decoder.Decode(&lte); err != nil {
			return fmt.Errorf("failed to decode LedgerToEntries: %w", err)
		}

		// Validate the ledger entry
		if err := p.validateLedgerToEntries(&lte); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		// Handle resume functionality - skip until we find the start correlation ID
		if skipUntilFound {
			// Check if any entry in this ledger has the start correlation ID
			for _, entry := range lte.Entries {
				if entry.CorrelationId == startFromCorrelationID {
					foundStartCorrelationID = true
					skipUntilFound = false
					break
				}
			}

			if !foundStartCorrelationID {
				// Skip this entire ledger entry
				continue
			} else {
				// We found the start correlation ID, but we need to filter out entries before and including it
				// (since we want to start from the NEXT correlation ID after the last processed one)
				var filteredEntries []*types.LedgerEntry
				startFound := false

				for _, entry := range lte.Entries {
					if !startFound && entry.CorrelationId == startFromCorrelationID {
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

		// Add to current chunk
		currentChunk.LedgerToEntries = append(currentChunk.LedgerToEntries, lte)

		// Check if we should start a new chunk based on size
		chunkSize := getChunkSizeBytes(currentChunk)
		if chunkSize > p.config.MaxChunkSizeBytes {
			// Remove the last entry and create a new chunk
			currentChunk.LedgerToEntries = currentChunk.LedgerToEntries[:len(currentChunk.LedgerToEntries)-1]

			// Add the current chunk to chunks list
			p.chunks = append(p.chunks, currentChunk)

			// Start a new chunk with the current entry
			currentChunk = &types.GenesisState{
				LedgerToEntries: []types.LedgerToEntries{lte},
			}
		}

		p.stats.TotalLedgers++

		// Count entries
		if lte.Entries != nil {
			p.stats.TotalEntries += len(lte.Entries)
		}
	}

	// Add the final chunk if it has data
	if len(currentChunk.LedgerToEntries) > 0 {
		p.chunks = append(p.chunks, currentChunk)
	}

	// Expect array end
	token, err = decoder.Token()
	if err != nil {
		return fmt.Errorf("failed to read array end: %w", err)
	}

	if delim, ok := token.(json.Delim); !ok || delim != ']' {
		return fmt.Errorf("expected array end, got %v", token)
	}

	return nil
}

// validateLedgerToEntries validates a single LedgerToEntries object
func (p *StreamingGenesisProcessor) validateLedgerToEntries(lte *types.LedgerToEntries) error {
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
func (p *StreamingGenesisProcessor) skipValue(decoder *json.Decoder) error {
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
