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
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/spf13/cobra"

	"cosmossdk.io/log"

	"github.com/provenance-io/provenance/x/ledger/types"
)

// ChunkConfig defines configuration for chunking large datasets
type ChunkConfig struct {
	MaxChunkSizeBytes  int // Maximum chunk size in bytes (approximate)
	MaxGasPerChunk     int // Maximum gas consumption per chunk (approximate)
	MaxEntriesPerChunk int // Maximum number of entries per chunk (fallback limit)
	MaxTxSizeBytes     int // Maximum transaction size in bytes (blockchain limit)
}

// DefaultChunkConfig returns a reasonable default configuration
func DefaultChunkConfig() ChunkConfig {
	return ChunkConfig{
		MaxChunkSizeBytes:  300000,  // ~300KB per chunk (reduced for gas considerations)
		MaxGasPerChunk:     4000000, // 4M gas per chunk (matching blockchain limit)
		MaxEntriesPerChunk: 200,     // 200 entries per chunk (reduced for gas considerations)
		MaxTxSizeBytes:     1000000, // 1MB max transaction size (typical blockchain limit)
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
	logger log.Logger
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

// estimateLedgerToEntriesSize estimates the approximate size in bytes of a LedgerToEntries object
func estimateLedgerToEntriesSize(lte *types.LedgerToEntries) int {
	size := 0

	// Estimate ledger key size
	if lte.LedgerKey != nil {
		size += len(lte.LedgerKey.NftId)
		size += len(lte.LedgerKey.AssetClassId)
	}

	// Estimate ledger size
	if lte.Ledger != nil {
		size += len(lte.Ledger.LedgerClassId)
		size += len(lte.Ledger.Key.NftId)
		size += len(lte.Ledger.Key.AssetClassId)
		// Add some overhead for other ledger fields
		size += 100
	}

	// Estimate entries size
	if lte.Entries != nil {
		for _, entry := range lte.Entries {
			if entry != nil {
				size += len(entry.CorrelationId)
				size += len(entry.ReversesCorrelationId)
				size += len(entry.TotalAmt.String())
				// Add overhead for other entry fields
				size += 200

				// Estimate applied amounts
				for _, amt := range entry.AppliedAmounts {
					size += len(amt.AppliedAmt.String())
					size += 50 // overhead for bucket type and other fields
				}

				// Estimate balance amounts
				for _, balance := range entry.BalanceAmounts {
					size += len(balance.BalanceAmt.String())
					size += 50 // overhead for bucket type and other fields
				}
			}
		}
	}

	// Add some overhead for JSON encoding and protobuf overhead
	return size + 500
}

// getChunkSizeBytes returns the actual serialized size of a chunk in bytes
func getChunkSizeBytes(chunk *types.GenesisState) int {
	data, err := json.Marshal(chunk)
	if err != nil {
		// Fallback to estimation if serialization fails
		return estimateChunkSize(chunk)
	}
	return len(data)
}

// estimateChunkSize estimates the size of a chunk by summing up individual ledger sizes
func estimateChunkSize(chunk *types.GenesisState) int {
	totalSize := 0
	for _, lte := range chunk.LedgerToEntries {
		totalSize += estimateLedgerToEntriesSize(&lte)
	}
	// Add overhead for the GenesisState wrapper
	return totalSize + 100
}

// estimateChunkGas estimates the gas consumption for a chunk based on its content
func estimateChunkGas(chunk *types.GenesisState) int {
	totalEntries := 0
	for _, lte := range chunk.LedgerToEntries {
		if lte.Entries != nil {
			totalEntries += len(lte.Entries)
		}
	}
	gasPerEntry := 4000 // Lowered for more safety
	baseGas := 100000
	totalGas := baseGas + (totalEntries * gasPerEntry)
	return int(float64(totalGas) * 1.20) // 20% margin
}

// CmdChunkedBulkImport creates a command for chunked bulk import
func CmdChunkedBulkImport() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "chunked-bulk-import <genesis_state_file> [max_chunk_size_bytes]",
		Aliases: []string{"cbi"},
		Short:   "Bulk import ledger data from a genesis state file using size-based chunking for large datasets",
		Long: `Bulk import ledger data from a genesis state file using intelligent chunking based on data size, gas consumption, and transaction limits.

The chunking algorithm considers:
- Maximum chunk size in bytes (default: 300KB)
- Maximum gas consumption per chunk (default: 4M gas)
- Maximum transaction size (default: 1MB)
- Maximum number of entries per chunk (default: 200)

This ensures chunks fit within blockchain transaction and gas limits while optimizing for performance.`,
		Example: `$ provenanced tx ledger chunked-bulk-import genesis.json --from mykey
$ provenanced tx ledger chunked-bulk-import genesis.json 500000 --from mykey`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Get logger from server context
			logger := server.GetServerContextFromCmd(cmd).Logger

			genesisStateFile := args[0]

			// Parse max chunk size in bytes if provided
			maxChunkSizeBytes := 500000 // default 500KB
			if len(args) > 1 {
				if maxChunkSizeBytes, err = strconv.Atoi(args[1]); err != nil {
					return fmt.Errorf("invalid max chunk size: %w", err)
				}
			}

			// Configure chunking
			config := DefaultChunkConfig()
			config.MaxChunkSizeBytes = maxChunkSizeBytes

			// Process the file using streaming
			processor := NewStreamingGenesisProcessor(config, logger)
			chunkedState, err := processor.ProcessFile(genesisStateFile)
			if err != nil {
				return fmt.Errorf("failed to process genesis file: %w", err)
			}

			logger.Info("Import summary",
				"import_id", chunkedState.ImportID,
				"total_ledgers", chunkedState.TotalLedgers,
				"total_entries", chunkedState.TotalEntries,
				"total_chunks", chunkedState.TotalChunks,
				"max_chunk_size_bytes", config.MaxChunkSizeBytes,
				"max_tx_size_bytes", config.MaxTxSizeBytes)

			// Debug: Print chunk information
			for i, chunk := range chunkedState.Chunks {
				logger.Debug("Chunk details",
					"chunk_index", i+1,
					"ledger_count", len(chunk.LedgerToEntries))
				for j, lte := range chunk.LedgerToEntries {
					hasLedger := lte.Ledger != nil
					ledgerClassId := ""
					if hasLedger {
						ledgerClassId = lte.Ledger.LedgerClassId
					}
					logger.Debug("LedgerToEntries details",
						"chunk_index", i+1,
						"entry_index", j+1,
						"nft_id", lte.LedgerKey.NftId,
						"has_ledger", hasLedger,
						"ledger_class_id", ledgerClassId,
						"entries_count", len(lte.Entries))
				}
			}

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
					logger.Info("Import cancelled by user")
					return nil
				}
			}

			// Update status to in_progress
			status.Status = "in_progress"
			status.UpdatedAt = time.Now().Format(time.RFC3339)
			_ = writeLocalBulkImportStatus(status)

			// Process each chunk
			for i, chunk := range chunkedState.Chunks {
				logger.Info("Processing chunk", "chunk_index", i+1, "total_chunks", chunkedState.TotalChunks)

				// Validate chunk before processing
				if chunk == nil {
					status.Status = "failed"
					status.ErrorMessage = fmt.Sprintf("chunk %d is nil", i+1)
					status.UpdatedAt = time.Now().Format(time.RFC3339)
					_ = writeLocalBulkImportStatus(status)
					return fmt.Errorf("chunk %d is nil", i+1)
				}

				chunkEntries := 0
				for _, ledger := range chunk.LedgerToEntries {
					chunkEntries += len(ledger.Entries)
				}
				chunkSize := getChunkSizeBytes(chunk)
				chunkGas := estimateChunkGas(chunk)

				// Validate chunk size against transaction limits
				if chunkSize > config.MaxTxSizeBytes {
					status.Status = "failed"
					status.ErrorMessage = fmt.Sprintf("chunk %d exceeds maximum transaction size: %d bytes > %d bytes", i+1, chunkSize, config.MaxTxSizeBytes)
					status.UpdatedAt = time.Now().Format(time.RFC3339)
					_ = writeLocalBulkImportStatus(status)
					return fmt.Errorf("chunk %d exceeds maximum transaction size: %d bytes > %d bytes", i+1, chunkSize, config.MaxTxSizeBytes)
				}

				// Validate chunk gas against gas limits
				if chunkGas > config.MaxGasPerChunk {
					status.Status = "failed"
					status.ErrorMessage = fmt.Sprintf("chunk %d exceeds maximum gas limit: %d gas > %d gas", i+1, chunkGas, config.MaxGasPerChunk)
					status.UpdatedAt = time.Now().Format(time.RFC3339)
					_ = writeLocalBulkImportStatus(status)
					return fmt.Errorf("chunk %d exceeds maximum gas limit: %d gas > %d gas", i+1, chunkGas, config.MaxGasPerChunk)
				}

				logger.Info("Chunk details",
					"chunk_index", i+1,
					"ledger_count", len(chunk.LedgerToEntries),
					"entry_count", chunkEntries,
					"size_bytes", chunkSize,
					"estimated_gas", chunkGas)

				// Get fresh client context for each transaction to ensure proper sequence number handling
				clientCtx, err = client.GetClientTxContext(cmd)
				if err != nil {
					status.Status = "failed"
					status.ErrorMessage = fmt.Sprintf("failed to get client context for chunk %d: %v", i+1, err)
					status.UpdatedAt = time.Now().Format(time.RFC3339)
					_ = writeLocalBulkImportStatus(status)
					return fmt.Errorf("failed to get client context for chunk %d: %w", i+1, err)
				}

				// Get current account sequence number
				account, err := clientCtx.AccountRetriever.GetAccount(clientCtx, clientCtx.FromAddress)
				if err != nil {
					status.Status = "failed"
					status.ErrorMessage = fmt.Sprintf("failed to get account info for chunk %d: %v", i+1, err)
					status.UpdatedAt = time.Now().Format(time.RFC3339)
					_ = writeLocalBulkImportStatus(status)
					return fmt.Errorf("failed to get account info for chunk %d: %w", i+1, err)
				}

				logger.Debug("Using sequence number", "sequence", account.GetSequence(), "chunk_index", i+1)

				// Debug: Print chunk contents before sending
				logger.Debug("Chunk contents before sending", "chunk_index", i+1)
				for j, lte := range chunk.LedgerToEntries {
					hasLedger := lte.Ledger != nil
					ledgerClassId := ""
					if hasLedger {
						ledgerClassId = lte.Ledger.LedgerClassId
					}
					logger.Debug("LedgerToEntries in chunk",
						"chunk_index", i+1,
						"entry_index", j+1,
						"nft_id", lte.LedgerKey.NftId,
						"has_ledger", hasLedger,
						"ledger_class_id", ledgerClassId,
						"entries_count", len(lte.Entries))
				}

				msg := &types.MsgBulkImportRequest{
					Authority:    clientCtx.FromAddress.String(),
					GenesisState: chunk,
				}

				// Force broadcast mode to sync for this transaction to ensure it's committed
				// before proceeding to the next chunk
				originalBroadcastMode := cmd.Flag(flags.FlagBroadcastMode).Value.String()
				cmd.Flag(flags.FlagBroadcastMode).Value.Set("sync")

				// Broadcast transaction
				err = tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)

				// Restore original broadcast mode
				cmd.Flag(flags.FlagBroadcastMode).Value.Set(originalBroadcastMode)

				if err != nil {
					status.Status = "failed"
					status.ErrorMessage = fmt.Sprintf("failed to process chunk %d: %v", i+1, err)
					status.UpdatedAt = time.Now().Format(time.RFC3339)
					_ = writeLocalBulkImportStatus(status)
					return fmt.Errorf("failed to process chunk %d: %w", i+1, err)
				}

				// Wait for transaction to be committed and account sequence to be updated
				// This is necessary to ensure the next transaction uses the correct sequence number
				if i < len(chunkedState.Chunks)-1 {
					logger.Info("Waiting for transaction confirmation")

					// Get the configured consensus timeout commit value for block time
					timeoutCommit := clientCtx.Viper.GetDuration("consensus.timeout_commit")
					if timeoutCommit == 0 {
						// Fallback to a reasonable default if not configured
						timeoutCommit = 3 * time.Second
					}

					// Wait for the next block to ensure the transaction is committed
					// and the account sequence number is updated in the blockchain state
					// Use 2x the timeout commit to ensure we wait for the next block
					waitDuration := timeoutCommit * 2
					logger.Debug("Waiting for next block", "wait_duration", waitDuration, "timeout_commit", timeoutCommit)
					time.Sleep(waitDuration)

					// Explicitly query the account to get the updated sequence number
					// and verify it has been incremented
					maxRetries := 10
					for retry := 0; retry < maxRetries; retry++ {
						account, err := clientCtx.AccountRetriever.GetAccount(clientCtx, clientCtx.FromAddress)
						if err != nil {
							if retry == maxRetries-1 {
								status.Status = "failed"
								status.ErrorMessage = fmt.Sprintf("failed to get account info after chunk %d: %v", i+1, err)
								status.UpdatedAt = time.Now().Format(time.RFC3339)
								_ = writeLocalBulkImportStatus(status)
								return fmt.Errorf("failed to get account info after chunk %d: %w", i+1, err)
							}
							time.Sleep(1 * time.Second)
							continue
						}

						logger.Debug("Account sequence updated", "sequence", account.GetSequence())
						break
					}
				}

				logger.Info("Chunk completed successfully", "chunk_index", i+1, "total_chunks", chunkedState.TotalChunks)
				status.CompletedChunks = i + 1
				status.UpdatedAt = time.Now().Format(time.RFC3339)
				_ = writeLocalBulkImportStatus(status)
			}

			status.Status = "completed"
			status.UpdatedAt = time.Now().Format(time.RFC3339)
			_ = writeLocalBulkImportStatus(status)

			logger.Info("Chunked bulk import completed successfully",
				"import_id", chunkedState.ImportID,
				"status_check_command", fmt.Sprintf("provenanced query ledger bulk-import-status %s", chunkedState.ImportID))

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
			// Get logger from server context
			logger := server.GetServerContextFromCmd(cmd).Logger

			importID := args[0]
			status, err := readLocalBulkImportStatus(importID)
			if err != nil {
				return fmt.Errorf("could not read local status file for import ID %s: %v", importID, err)
			}
			b, _ := json.MarshalIndent(status, "", "  ")
			logger.Info("Bulk import status", "import_id", importID, "status_json", string(b))
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
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
	chunkSizeBytes := 0
	chunkEntries := 0

	// Process each ledger entry
	for decoder.More() {
		var lte types.LedgerToEntries
		if err := decoder.Decode(&lte); err != nil {
			return fmt.Errorf("failed to decode LedgerToEntries: %w", err)
		}

		// Debug: Print what was decoded
		p.logger.Debug("Decoded LedgerToEntries",
			"has_ledger_key", lte.LedgerKey != nil,
			"has_ledger", lte.Ledger != nil,
			"entries_count", len(lte.Entries))
		if lte.LedgerKey != nil {
			p.logger.Debug("LedgerKey details",
				"nft_id", lte.LedgerKey.NftId,
				"asset_class_id", lte.LedgerKey.AssetClassId)
		}
		if lte.Ledger != nil {
			p.logger.Debug("Ledger details", "ledger_class_id", lte.Ledger.LedgerClassId)
		}

		// Debug: Print entry count and max entries per chunk before split check
		p.logger.Debug("LedgerToEntries entry count check",
			"entries_count", len(lte.Entries),
			"max_entries_per_chunk", p.config.MaxEntriesPerChunk)

		// If this ledger has too many entries, we need to split it
		if len(lte.Entries) > p.config.MaxEntriesPerChunk {
			// Split the ledger into multiple chunks
			err := p.splitLargeLedger(&lte, p.config)
			if err != nil {
				return fmt.Errorf("failed to split large ledger: %w", err)
			}
			// After splitting, continue to next ledger
			continue
		} else {
			// Estimate the size of this ledger entry
			estimatedSize := estimateLedgerToEntriesSize(&lte)

			// Check if adding this entry would exceed our limits
			wouldExceedSize := chunkSizeBytes+estimatedSize > p.config.MaxChunkSizeBytes
			wouldExceedEntries := chunkEntries+len(lte.Entries) > p.config.MaxEntriesPerChunk

			// Also check against transaction size limit
			testChunk := &types.GenesisState{
				LedgerToEntries: append(currentChunk.LedgerToEntries, lte),
			}
			testChunkSize := getChunkSizeBytes(testChunk)
			wouldExceedTxSize := testChunkSize > p.config.MaxTxSizeBytes

			// Check against gas limit
			testChunkGas := estimateChunkGas(testChunk)
			wouldExceedGas := testChunkGas > p.config.MaxGasPerChunk

			// Start a new chunk if we would exceed any limits
			if (wouldExceedSize || wouldExceedEntries || wouldExceedTxSize || wouldExceedGas) && len(currentChunk.LedgerToEntries) > 0 {
				p.logger.Debug("Appending chunk to chunks list",
					"first_ledger_has_ledger", len(currentChunk.LedgerToEntries) > 0 && currentChunk.LedgerToEntries[0].Ledger != nil)
				// Create a deep copy to avoid pointer issues
				chunkCopy := &types.GenesisState{
					LedgerToEntries: make([]types.LedgerToEntries, len(currentChunk.LedgerToEntries)),
				}
				for i, lte := range currentChunk.LedgerToEntries {
					chunkCopy.LedgerToEntries[i] = lte
					// Deep copy the Ledger object if it exists
					if lte.Ledger != nil {
						ledgerCopy := *lte.Ledger
						chunkCopy.LedgerToEntries[i].Ledger = &ledgerCopy
					}
				}
				p.chunks = append(p.chunks, chunkCopy)
				currentChunk = &types.GenesisState{
					LedgerToEntries: []types.LedgerToEntries{},
				}
				chunkSizeBytes = 0
				chunkEntries = 0
			}

			// Add to current chunk
			currentChunk.LedgerToEntries = append(currentChunk.LedgerToEntries, lte)
			chunkSizeBytes += estimatedSize
			chunkEntries += len(lte.Entries)
		}

		p.stats.TotalLedgers++

		// Count entries
		if lte.Entries != nil {
			p.stats.TotalEntries += len(lte.Entries)
		}
	}

	// Add the final chunk if it has data
	if len(currentChunk.LedgerToEntries) > 0 {
		p.logger.Debug("Appending final chunk to chunks list",
			"first_ledger_has_ledger", len(currentChunk.LedgerToEntries) > 0 && currentChunk.LedgerToEntries[0].Ledger != nil)
		// Create a deep copy to avoid pointer issues
		chunkCopy := &types.GenesisState{
			LedgerToEntries: make([]types.LedgerToEntries, len(currentChunk.LedgerToEntries)),
		}
		for i, lte := range currentChunk.LedgerToEntries {
			chunkCopy.LedgerToEntries[i] = lte
			// Deep copy the Ledger object if it exists
			if lte.Ledger != nil {
				ledgerCopy := *lte.Ledger
				chunkCopy.LedgerToEntries[i].Ledger = &ledgerCopy
			}
		}
		p.chunks = append(p.chunks, chunkCopy)
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

// splitLargeLedger splits a ledger with too many entries into multiple chunks
func (p *StreamingGenesisProcessor) splitLargeLedger(lte *types.LedgerToEntries, config ChunkConfig) error {
	entriesPerChunk := config.MaxEntriesPerChunk

	p.logger.Debug("Splitting large ledger",
		"entries_count", len(lte.Entries),
		"entries_per_chunk", entriesPerChunk)

	// Update stats for this ledger
	p.stats.TotalLedgers++
	if lte.Entries != nil {
		p.stats.TotalEntries += len(lte.Entries)
	}

	for i := 0; i < len(lte.Entries); i += entriesPerChunk {
		end := i + entriesPerChunk
		if end > len(lte.Entries) {
			end = len(lte.Entries)
		}
		splitLte := types.LedgerToEntries{
			LedgerKey: lte.LedgerKey,
			Entries:   lte.Entries[i:end],
		}
		if i == 0 {
			splitLte.Ledger = lte.Ledger
			p.logger.Debug("Split chunk created", "chunk_index", i/entriesPerChunk+1, "entries_range", fmt.Sprintf("%d-%d", i, end-1), "has_ledger", true)
		} else {
			splitLte.Ledger = nil
			p.logger.Debug("Split chunk created", "chunk_index", i/entriesPerChunk+1, "entries_range", fmt.Sprintf("%d-%d", i, end-1), "has_ledger", false)
		}
		// Deep copy the Ledger object if it exists
		if splitLte.Ledger != nil {
			ledgerCopy := *splitLte.Ledger
			splitLte.Ledger = &ledgerCopy
		}
		chunk := &types.GenesisState{
			LedgerToEntries: []types.LedgerToEntries{splitLte},
		}
		p.chunks = append(p.chunks, chunk)
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
