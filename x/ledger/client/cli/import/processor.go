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

	// Resume state
	skipUntilFound          bool
	foundStartCorrelationID bool
	startFromCorrelationID  string

	// End of stream flag
	atEnd bool

	// New state for partial lte processing
	inProgressLTE        *types.LedgerToEntries		// If non-nil, we're in the middle of splitting an lte
	inProgressLTEEntryIndex int						// Index of the next entry to process in partialLTE
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

	maxGasPerChunk := p.config.GetEffectiveGasLimit()
	maxTxSize := p.config.MaxTxSizeBytes
	candidateChunk := &types.GenesisState{
		LedgerToEntries: []types.LedgerToEntries{},
	}
	
	for {
		startEntry := 0
		// Determine processingLTE
		// If we have an LTE from last time, continue processing it
		// Otherwise, get the next LTE	
		if p.inProgressLTE != nil {
			startEntry = p.inProgressLTEEntryIndex
		} else {
			// Get the next LTE and store it in inProgressLTE
			if !p.decoder.More() {
				p.atEnd = true
				break
			}
			if err := p.decoder.Decode(&p.inProgressLTE); err != nil {
				return nil, fmt.Errorf("failed to decode LedgerToEntries: %w", err)
			}

			// Validate the ledger entry
			if err := p.validateLedgerToEntries(p.inProgressLTE); err != nil {
				return nil, fmt.Errorf("validation failed: %w", err)
			}

			// Handle resume functionality - skip until we find the start correlation ID
			if p.skipUntilFound {
				for i, entry := range p.inProgressLTE.Entries {
					if entry.CorrelationId == p.startFromCorrelationID {
						p.foundStartCorrelationID = true
						p.skipUntilFound = false
						startEntry = i + 1 // skip the found entry, start at next
						break
					}
				}
				if !p.foundStartCorrelationID {
					continue // skip this entire lte
				}
			}
		}

		candidateLTE := types.LedgerToEntries{
			LedgerKey: p.inProgressLTE.LedgerKey,
			Ledger:    nil, // assume not the first LTE so no ledger object for now
			Entries:   []*types.LedgerEntry{},
		}
		
		// If this is a new LTE, try to add set the ledger object
		if startEntry == 0 {
			candidateLTE.Ledger = p.inProgressLTE.Ledger // Only the first LTE has the ledger object
		}

		// Add the LTE (with no entries) to the chunk first to see if it fits
		tempChunk := &types.GenesisState{
			LedgerToEntries: append(append([]types.LedgerToEntries{}, candidateChunk.LedgerToEntries...), candidateLTE),
		}
		gas := 0
		if gasCosts != nil {
			gas = estimateChunkGasFromCosts(tempChunk, gasCosts)
		}
		size := getChunkSizeBytes(tempChunk)
		if (gasCosts != nil && gas > maxGasPerChunk) || size > maxTxSize {
			// Ledger+key (no entries) do not fit, return current chunk and start a new one
			if len(candidateChunk.LedgerToEntries) > 0 {
				p.inProgressLTEEntryIndex = startEntry
				return candidateChunk, nil
			} else {
				return nil, fmt.Errorf("ledger object too large to fit in chunk (ledgerKey=%v)", candidateLTE.LedgerKey)
			}
		}
		
		entries := p.inProgressLTE.Entries
		for i := startEntry; i < len(entries); i++ {
			tempLTE := candidateLTE
			tempLTE.Entries = append(tempLTE.Entries, entries[i])
			
			testChunk := &types.GenesisState{
				LedgerToEntries: append(append([]types.LedgerToEntries{}, candidateChunk.LedgerToEntries...), tempLTE),
			}
			gas := 0
			if gasCosts != nil {
				gas = estimateChunkGasFromCosts(testChunk, gasCosts)
			}
			size := getChunkSizeBytes(testChunk)
			// Chunk is too large, return candidate chunk and save state for next time
			if (gasCosts != nil && gas > maxGasPerChunk) || size > maxTxSize {
				p.inProgressLTEEntryIndex = i
				return tempChunk, nil
			}
			candidateLTE = tempLTE
			tempChunk = testChunk
			p.stats.TotalEntries++
		}

		// Finished this lte
		candidateChunk = tempChunk
		p.inProgressLTE = nil
		p.inProgressLTEEntryIndex = 0
		p.stats.TotalLedgers++
	}

	if len(candidateChunk.LedgerToEntries) == 0 {
		return nil, io.EOF
	}
	return candidateChunk, nil
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
