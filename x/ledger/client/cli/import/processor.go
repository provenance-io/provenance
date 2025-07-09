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

	p.logger.Info("Processing genesis file", "filename", filename, "size_bytes", fileInfo.Size())

	// Use buffered reader for efficient reading
	reader := bufio.NewReader(file)

	// Parse the JSON structure using streaming
	err = p.parseStreamingJSON(reader)
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

// parseStreamingJSON parses the JSON file using a streaming approach
func (p *StreamingGenesisProcessor) parseStreamingJSON(reader *bufio.Reader) error {
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
			err = p.parseLedgerToEntriesArray(decoder)
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
func (p *StreamingGenesisProcessor) parseLedgerToEntriesArray(decoder *json.Decoder) error {
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
