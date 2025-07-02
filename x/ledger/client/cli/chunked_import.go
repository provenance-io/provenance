package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/provenance-io/provenance/x/ledger/types"
)

// ChunkConfig defines configuration for chunking large datasets
type ChunkConfig struct {
	MaxLedgersPerChunk int // Maximum number of ledgers per chunk
	MaxEntriesPerChunk int // Maximum number of entries per chunk
	MaxChunkSizeBytes  int // Maximum chunk size in bytes (approximate)
	MaxGasPerChunk     int // Maximum gas consumption per chunk (approximate)
}

// DefaultChunkConfig returns a reasonable default configuration
func DefaultChunkConfig() ChunkConfig {
	return ChunkConfig{
		MaxLedgersPerChunk: 100,     // 100 ledgers per chunk
		MaxEntriesPerChunk: 1000,    // 1000 entries per chunk
		MaxChunkSizeBytes:  500000,  // ~500KB per chunk
		MaxGasPerChunk:     2000000, // 2M gas per chunk (leaving room for overhead)
	}
}

// ChunkedGenesisState represents a chunked version of GenesisState
type ChunkedGenesisState struct {
	ImportID     string
	TotalChunks  int
	Chunks       []*types.GenesisState
	TotalLedgers int
	TotalEntries int
}

// LocalBulkImportStatus tracks the status of a chunked bulk import on the client side.
type LocalBulkImportStatus struct {
	ImportID        string `json:"import_id"`
	TotalChunks     int    `json:"total_chunks"`
	CompletedChunks int    `json:"completed_chunks"`
	TotalLedgers    int    `json:"total_ledgers"`
	TotalEntries    int    `json:"total_entries"`
	Status          string `json:"status"` // "pending", "in_progress", "completed", "failed"
	ErrorMessage    string `json:"error_message,omitempty"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

// StreamingGenesisProcessor handles streaming JSON parsing for large genesis files
type StreamingGenesisProcessor struct {
	config ChunkConfig
	chunks []*types.GenesisState
	stats  *ImportStats
}

// ImportStats tracks import statistics
type ImportStats struct {
	TotalLedgers int
	TotalEntries int
	TotalChunks  int
}

func statusFileName(importID string) string {
	return ".bulk_import_status." + importID + ".json"
}

func readLocalBulkImportStatus(importID string) (*LocalBulkImportStatus, error) {
	file := statusFileName(importID)
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var status LocalBulkImportStatus
	if err := json.Unmarshal(b, &status); err != nil {
		return nil, err
	}
	return &status, nil
}

func writeLocalBulkImportStatus(status *LocalBulkImportStatus) error {
	file := statusFileName(status.ImportID)
	b, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(file, b, 0644)
}

// CmdChunkedBulkImport creates a command for chunked bulk import
func CmdChunkedBulkImport() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "chunked-bulk-import <genesis_state_file> [chunk_size]",
		Aliases: []string{"cbi"},
		Short:   "Bulk import ledger data from a genesis state file using chunking for large datasets",
		Example: `$ provenanced tx ledger chunked-bulk-import genesis.json --from mykey\n$ provenanced tx ledger chunked-bulk-import genesis.json 50 --from mykey`,
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			genesisStateFile := args[0]

			// Parse chunk size if provided
			chunkSize := 100 // default
			if len(args) > 1 {
				if chunkSize, err = strconv.Atoi(args[1]); err != nil {
					return fmt.Errorf("invalid chunk size: %w", err)
				}
			}

			// Configure chunking
			config := DefaultChunkConfig()
			config.MaxLedgersPerChunk = chunkSize

			// Process the file using streaming
			processor := NewStreamingGenesisProcessor(config)
			chunkedState, err := processor.ProcessFile(genesisStateFile)
			if err != nil {
				return fmt.Errorf("failed to process genesis file: %w", err)
			}

			fmt.Printf("Import ID: %s\n", chunkedState.ImportID)
			fmt.Printf("Total ledgers: %d\n", chunkedState.TotalLedgers)
			fmt.Printf("Total entries: %d\n", chunkedState.TotalEntries)
			fmt.Printf("Number of chunks: %d\n", chunkedState.TotalChunks)

			// Initialize local status file
			status := &LocalBulkImportStatus{
				ImportID:        chunkedState.ImportID,
				TotalChunks:     chunkedState.TotalChunks,
				CompletedChunks: 0,
				TotalLedgers:    chunkedState.TotalLedgers,
				TotalEntries:    chunkedState.TotalEntries,
				Status:          "pending",
				CreatedAt:       time.Now().Format(time.RFC3339),
				UpdatedAt:       time.Now().Format(time.RFC3339),
			}
			_ = writeLocalBulkImportStatus(status)

			// Ask for confirmation
			confirm, err := cmd.Flags().GetBool("yes")
			if err != nil {
				return err
			}

			if !confirm {
				fmt.Print("Proceed with chunked import? (y/N): ")
				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
					fmt.Println("Import cancelled")
					return nil
				}
			}

			// Update status to in_progress
			status.Status = "in_progress"
			status.UpdatedAt = time.Now().Format(time.RFC3339)
			_ = writeLocalBulkImportStatus(status)

			// Process each chunk
			for i, chunk := range chunkedState.Chunks {
				fmt.Printf("Processing chunk %d/%d...\n", i+1, chunkedState.TotalChunks)

				// Validate chunk before processing
				if chunk == nil {
					status.Status = "failed"
					status.ErrorMessage = fmt.Sprintf("chunk %d is nil", i+1)
					status.UpdatedAt = time.Now().Format(time.RFC3339)
					_ = writeLocalBulkImportStatus(status)
					return fmt.Errorf("chunk %d is nil", i+1)
				}

				fmt.Printf("Chunk %d contains %d ledger to entries\n", i+1, len(chunk.LedgerToEntries))

				msg := &types.MsgBulkImportRequest{
					Authority:    clientCtx.FromAddress.String(),
					GenesisState: chunk,
				}

				// Broadcast transaction
				err = tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
				if err != nil {
					status.Status = "failed"
					status.ErrorMessage = fmt.Sprintf("failed to process chunk %d: %v", i+1, err)
					status.UpdatedAt = time.Now().Format(time.RFC3339)
					_ = writeLocalBulkImportStatus(status)
					return fmt.Errorf("failed to process chunk %d: %w", i+1, err)
				}

				fmt.Printf("Chunk %d/%d completed successfully\n", i+1, chunkedState.TotalChunks)
				status.CompletedChunks = i + 1
				status.UpdatedAt = time.Now().Format(time.RFC3339)
				_ = writeLocalBulkImportStatus(status)
			}

			status.Status = "completed"
			status.UpdatedAt = time.Now().Format(time.RFC3339)
			_ = writeLocalBulkImportStatus(status)

			fmt.Printf("Chunked bulk import completed successfully!\n")
			fmt.Printf("Import ID: %s\n", chunkedState.ImportID)
			fmt.Printf("You can check status with: provenanced query ledger bulk-import-status %s\n", chunkedState.ImportID)

			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdBulkImportStatus creates a command to check bulk import status
func CmdBulkImportStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bulk-import-status <import_id>",
		Aliases: []string{"bis"},
		Short:   "Check the status of a bulk import operation",
		Example: `$ provenanced query ledger bulk-import-status import_1234567890`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			importID := args[0]
			status, err := readLocalBulkImportStatus(importID)
			if err != nil {
				return fmt.Errorf("could not read local status file for import ID %s: %v", importID, err)
			}
			b, _ := json.MarshalIndent(status, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// NewStreamingGenesisProcessor creates a new streaming processor
func NewStreamingGenesisProcessor(config ChunkConfig) *StreamingGenesisProcessor {
	return &StreamingGenesisProcessor{
		config: config,
		chunks: []*types.GenesisState{},
		stats:  &ImportStats{},
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

	fmt.Printf("Processing file: %s (size: %d bytes)\n", filename, fileInfo.Size())

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

		if fieldName == "ledgerToEntries" {
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
	ledgerCount := 0

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
		ledgerCount++
		p.stats.TotalLedgers++

		// Count entries
		if lte.Entries != nil {
			p.stats.TotalEntries += len(lte.Entries)
		}

		// Check if we need to start a new chunk
		if ledgerCount >= p.config.MaxLedgersPerChunk {
			p.chunks = append(p.chunks, currentChunk)
			currentChunk = &types.GenesisState{
				LedgerToEntries: []types.LedgerToEntries{},
			}
			ledgerCount = 0
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

// generateImportID creates a unique identifier for an import operation
func generateImportID() string {
	// In a real implementation, you might want to use a more sophisticated ID generation
	// For now, we'll use a simple timestamp-based approach
	return fmt.Sprintf("import_%d", time.Now().UnixNano())
}
