package cli

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/spf13/cobra"

	"github.com/provenance-io/provenance/x/ledger/types"
)

// calculateFileHash calculates SHA256 hash of a file for validation
func calculateFileHash(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("failed to open file for hashing: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to read file for hashing: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// getLastCorrelationIDFromChunk extracts the last correlation ID from a chunk
func getLastCorrelationIDFromChunk(chunk *types.GenesisState) string {
	var lastCorrelationID string

	for _, lte := range chunk.LedgerToEntries {
		for _, entry := range lte.Entries {
			if entry.CorrelationId != "" {
				lastCorrelationID = entry.CorrelationId
			}
		}
	}

	return lastCorrelationID
}

// getFirstCorrelationIDFromChunk extracts the first correlation ID from a chunk
func getFirstCorrelationIDFromChunk(chunk *types.GenesisState) string {
	if chunk == nil || len(chunk.LedgerToEntries) == 0 {
		return ""
	}

	for _, lte := range chunk.LedgerToEntries {
		for _, entry := range lte.Entries {
			if entry.CorrelationId != "" {
				return entry.CorrelationId
			}
		}
	}
	return ""
}

// getCorrelationIDRangeFromChunk extracts the first and last correlation IDs from a chunk
func getCorrelationIDRangeFromChunk(chunk *types.GenesisState) (string, string) {
	if chunk == nil || len(chunk.LedgerToEntries) == 0 {
		return "", ""
	}

	var firstID, lastID string
	for _, lte := range chunk.LedgerToEntries {
		for _, entry := range lte.Entries {
			if entry.CorrelationId != "" {
				if firstID == "" {
					firstID = entry.CorrelationId
				}
				lastID = entry.CorrelationId
			}
		}
	}
	return firstID, lastID
}

// checkCorrelationIDExists checks if a correlation ID exists on-chain by querying the ledger

// checkTransactionAlreadyProcessed checks if a transaction was already processed by querying the blockchain
func checkTransactionAlreadyProcessed(clientCtx client.Context, txHash string, logger log.Logger) (bool, error) {
	if txHash == "" {
		return false, nil
	}

	// Query the transaction by hash using the correct API
	resp, err := clientCtx.Client.Tx(context.Background(), []byte(txHash), false)
	if err != nil {
		// If transaction not found, it wasn't processed
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, fmt.Errorf("failed to query transaction %s: %w", txHash, err)
	}

	// Check if transaction was successful
	if resp.TxResult.Code == 0 {
		logger.Info("Transaction already processed successfully", "tx_hash", txHash)
		return true, nil
	}

	logger.Info("Transaction found but failed", "tx_hash", txHash, "code", resp.TxResult.Code)
	return false, nil
}

// findNextCorrelationIDAfter scans a genesis file to find the next correlation ID after the given one
func findNextCorrelationIDAfter(filename string, afterCorrelationID string, logger log.Logger) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("failed to open file for correlation ID scanning: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	// Expect the root object
	token, err := decoder.Token()
	if err != nil {
		return "", fmt.Errorf("failed to read JSON token: %w", err)
	}

	if delim, ok := token.(json.Delim); !ok || delim != '{' {
		return "", fmt.Errorf("expected JSON object, got %v", token)
	}

	// Process the root object
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return "", fmt.Errorf("failed to read field name: %w", err)
		}

		fieldName, ok := token.(string)
		if !ok {
			return "", fmt.Errorf("expected field name, got %v", token)
		}

		if fieldName == "ledgerToEntries" || fieldName == "ledger_to_entries" {
			nextID, err := findNextCorrelationIDInArray(decoder, afterCorrelationID)
			if err != nil {
				return "", fmt.Errorf("failed to find next correlation ID in array: %w", err)
			}
			return nextID, nil
		} else {
			// Skip unknown fields
			if err := skipJSONValue(decoder); err != nil {
				return "", fmt.Errorf("failed to skip field %s: %w", fieldName, err)
			}
		}
	}

	// Expect closing brace
	token, err = decoder.Token()
	if err != nil {
		return "", fmt.Errorf("failed to read closing brace: %w", err)
	}

	if delim, ok := token.(json.Delim); !ok || delim != '}' {
		return "", fmt.Errorf("expected closing brace, got %v", token)
	}

	return "", nil
}

// findNextCorrelationIDInArray scans the ledgerToEntries array to find the next correlation ID
func findNextCorrelationIDInArray(decoder *json.Decoder, afterCorrelationID string) (string, error) {
	// Expect array start
	token, err := decoder.Token()
	if err != nil {
		return "", fmt.Errorf("failed to read array start: %w", err)
	}

	if delim, ok := token.(json.Delim); !ok || delim != '[' {
		return "", fmt.Errorf("expected array start, got %v", token)
	}

	foundTarget := false

	// Process each ledger entry
	for decoder.More() {
		var lte types.LedgerToEntries
		if err := decoder.Decode(&lte); err != nil {
			return "", fmt.Errorf("failed to decode LedgerToEntries: %w", err)
		}

		// Check entries in this ledger
		for _, entry := range lte.Entries {
			if !foundTarget && entry.CorrelationId == afterCorrelationID {
				foundTarget = true
				continue
			}
			if foundTarget && entry.CorrelationId != "" {
				// Found the next correlation ID after the target
				return entry.CorrelationId, nil
			}
		}
	}

	// Expect array end
	token, err = decoder.Token()
	if err != nil {
		return "", fmt.Errorf("failed to read array end: %w", err)
	}

	if delim, ok := token.(json.Delim); !ok || delim != ']' {
		return "", fmt.Errorf("expected array end, got %v", token)
	}

	return "", nil
}

// skipJSONValue skips a JSON value (object, array, string, number, boolean, null)
func skipJSONValue(decoder *json.Decoder) error {
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
			if err := skipJSONValue(decoder); err != nil {
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
			if err := skipJSONValue(decoder); err != nil {
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

// CmdChunkedBulkImport creates a command for chunked bulk import
func CmdChunkedBulkImport() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "chunked-bulk-import <genesis_state_file> [max_chunk_size_bytes]",
		Aliases: []string{"cbi"},
		Short:   "Bulk import ledger data from a genesis state file using size-based chunking for large datasets",
		Long: `Bulk import ledger data from a genesis state file using intelligent chunking based on data size, gas consumption, and transaction limits.

The chunking algorithm considers:
- Maximum chunk size in bytes (default: 10MB, memory safety limit)
- Maximum gas consumption per transaction (default: 4M gas)
- Maximum transaction size (default: 1MB)

This ensures chunks fit within blockchain transaction and gas limits while optimizing for performance.

IMPORT ID BEHAVIOR:
- Without --import-id: Always starts a fresh import with a new auto-generated import ID
- With --import-id: Uses the specified import ID (resume if status exists, fresh start if not)

RESUME FUNCTIONALITY:
When --import-id is provided, if an import was previously interrupted, the command will automatically 
detect the existing status file and resume from where it left off. The resume mechanism tracks the 
last attempted chunk's first and last correlation IDs, transaction hash, and whether it was confirmed on-chain. 
The chunk status is stored before sending the transaction to ensure all attempts are captured, 
even if the import is cancelled during confirmation. On resume, if the last chunk was sent but 
not confirmed, the system queries the blockchain using the transaction hash to check if the transaction 
succeeded, ensuring accurate resume behavior regardless of interruption timing.`,
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
			maxChunkSizeBytes := 5000000 // default 5MB (5x larger than max tx size)
			if len(args) > 1 {
				if maxChunkSizeBytes, err = strconv.Atoi(args[1]); err != nil {
					return fmt.Errorf("invalid max chunk size: %w", err)
				}
			}

			// Configure chunking
			config := DefaultChunkConfig()
			config.MaxChunkSizeBytes = maxChunkSizeBytes

			// Add import-id flag support
			userImportID, _ := cmd.Flags().GetString("import-id")

			// Determine import ID behavior:
			// - If --import-id is provided: use it (resume if status exists, fresh start if not)
			// - If no --import-id: always start fresh with auto-generated import ID
			var importID string
			if userImportID != "" {
				importID = userImportID
			} else {
				// Generate a new import ID for fresh start
				importID = generateImportID()
			}

			// Calculate file hash for validation
			fileHash, err := calculateFileHash(genesisStateFile)
			if err != nil {
				return fmt.Errorf("failed to calculate file hash: %w", err)
			}

			// Check for existing status file to determine if this is a resume operation
			logger.Info("Checking for existing status file", "import_id", importID, "status_file", statusFileName(importID))
			existingStatus, err := readLocalBulkImportStatus(importID)
			isResume := false
			startChunkIndex := 0
			lastSuccessfulCorrelationID := ""

			if err != nil {
				logger.Info("No existing status file found, starting fresh import", "error", err)
			} else if existingStatus != nil {
				logger.Info("Existing status file found",
					"status", existingStatus.Status,
					"completed_chunks", existingStatus.CompletedChunks,
					"total_chunks", existingStatus.TotalChunks,
					"total_ledgers", existingStatus.TotalLedgers,
					"total_entries", existingStatus.TotalEntries,
					"file_hash", existingStatus.FileHash,
					"current_file_hash", fileHash)
			}

			// Determine resume mode and starting position
			if err == nil && existingStatus != nil && existingStatus.LastAttemptedChunk != nil {
				// Status file exists with LastAttemptedChunk - check if we can resume
				if existingStatus.Status == "completed" {
					logger.Info("Import already completed", "import_id", importID)
					return nil
				} else if existingStatus.Status == "failed" || existingStatus.Status == "in_progress" {
					// Validate that the existing status matches our current import
					logger.Info("Checking if existing status matches current import",
						"existing_ledgers", existingStatus.TotalLedgers,
						"existing_entries", existingStatus.TotalEntries,
						"existing_file_hash", existingStatus.FileHash,
						"current_file_hash", fileHash)

					// For resume, we only check file hash since we'll reprocess the file
					if existingStatus.FileHash == fileHash {

						isResume = true
						startChunkIndex = existingStatus.CompletedChunks
						lastSuccessfulCorrelationID = existingStatus.LastSuccessfulCorrelationID

						logger.Info("Resuming import from previous state",
							"import_id", importID,
							"completed_chunks", existingStatus.CompletedChunks,
							"total_chunks", existingStatus.TotalChunks,
							"status", existingStatus.Status,
							"last_correlation_id", lastSuccessfulCorrelationID,
							"file_hash", fileHash)

						// Log resume details
						if existingStatus.LastAttemptedChunk != nil {
							logger.Info("Resume details",
								"last_attempted_first_correlation_id", existingStatus.LastAttemptedChunk.FirstCorrelationID,
								"last_attempted_last_correlation_id", existingStatus.LastAttemptedChunk.LastCorrelationID,
								"last_attempted_confirmed", existingStatus.LastAttemptedChunk.Confirmed)
						}

						// Handle unconfirmed last attempted chunk
						if !existingStatus.LastAttemptedChunk.Confirmed {
							logger.Info("Last chunk was unconfirmed, checking if it was actually processed",
								"last_tx_hash", existingStatus.LastAttemptedChunk.TransactionHash)

							processed, err := checkTransactionAlreadyProcessed(clientCtx, existingStatus.LastAttemptedChunk.TransactionHash, logger)
							if err != nil {
								logger.Warn("Failed to check transaction status, proceeding anyway", "error", err)
							} else if processed {
								// The transaction actually succeeded, mark it as confirmed
								logger.Info("Last chunk was actually processed successfully, marking as confirmed",
									"last_tx_hash", existingStatus.LastAttemptedChunk.TransactionHash)
								existingStatus.LastAttemptedChunk.Confirmed = true
								existingStatus.UpdatedAt = time.Now().Format(time.RFC3339)
								_ = writeLocalBulkImportStatus(existingStatus)

								// Start from the next correlation ID (after the last one)
								lastSuccessfulCorrelationID = existingStatus.LastAttemptedChunk.LastCorrelationID
							} else {
								// The transaction really failed, we need to find the next correlation ID after the last successfully processed one
								// Since we don't have a transaction hash, we need to determine where to resume from
								if existingStatus.LastSuccessfulCorrelationID != "" {
									// We have a last successful correlation ID, so we should start from the next one
									logger.Info("Last chunk was not processed, scanning file to find next correlation ID after",
										"last_successful_correlation_id", existingStatus.LastSuccessfulCorrelationID)

									nextCorrelationID, err := findNextCorrelationIDAfter(genesisStateFile, existingStatus.LastSuccessfulCorrelationID, logger)
									if err != nil {
										logger.Warn("Failed to find next correlation ID, import may be complete or file may be corrupted", "error", err)
										lastSuccessfulCorrelationID = ""
									} else if nextCorrelationID != "" {
										logger.Info("Found next correlation ID to resume from", "next_correlation_id", nextCorrelationID)
										lastSuccessfulCorrelationID = nextCorrelationID
									} else {
										logger.Info("No next correlation ID found, import may be complete")
										lastSuccessfulCorrelationID = ""
									}
								} else {
									// No last successful correlation ID, we need to find the next correlation ID after the first correlation ID of the last attempted chunk
									// This is because the first correlation ID might have already been processed
									logger.Info("Last chunk was not processed and no last successful correlation ID, scanning file to find next correlation ID after",
										"first_correlation_id", existingStatus.LastAttemptedChunk.FirstCorrelationID)

									nextCorrelationID, err := findNextCorrelationIDAfter(genesisStateFile, existingStatus.LastAttemptedChunk.FirstCorrelationID, logger)
									if err != nil {
										logger.Warn("Failed to find next correlation ID, import may be complete or file may be corrupted", "error", err)
										lastSuccessfulCorrelationID = ""
									} else if nextCorrelationID != "" {
										logger.Info("Found next correlation ID to resume from", "next_correlation_id", nextCorrelationID)
										lastSuccessfulCorrelationID = nextCorrelationID
									} else {
										logger.Info("No next correlation ID found, import may be complete")
										lastSuccessfulCorrelationID = ""
									}
								}
							}
						} else {
							// Last chunk was confirmed, start from the next correlation ID
							lastSuccessfulCorrelationID = existingStatus.LastAttemptedChunk.LastCorrelationID
						}

						if existingStatus.Status == "failed" {
							logger.Info("Previous import failed", "error", existingStatus.ErrorMessage)
						}
					} else {
						logger.Warn("Existing status file found but file hash doesn't match - starting fresh import",
							"existing_file_hash", existingStatus.FileHash,
							"current_file_hash", fileHash)
					}
				}
			}

			// Process the file using streaming, with resume-aware positioning
			processor := NewStreamingGenesisProcessor(config, logger)
			var chunkedState *ChunkedGenesisState

			if isResume && lastSuccessfulCorrelationID != "" {
				// Resume mode: process file starting from the appropriate correlation ID
				logger.Info("Processing file for resume", "start_correlation_id", lastSuccessfulCorrelationID)
				chunkedState, err = processor.ProcessFileFromCorrelationID(genesisStateFile, lastSuccessfulCorrelationID)
			} else if isResume && lastSuccessfulCorrelationID == "" {
				// Resume mode but no next correlation ID found - import may be complete
				logger.Info("No next correlation ID found, checking if import is complete")

				// Mark the import as completed since there's nothing more to process
				if existingStatus != nil {
					existingStatus.Status = "completed"
					existingStatus.UpdatedAt = time.Now().Format(time.RFC3339)
					_ = writeLocalBulkImportStatus(existingStatus)
					logger.Info("Import marked as completed - no more data to process")
				}
				return nil
			} else {
				// Fresh start: process entire file
				logger.Info("Processing file for fresh import")
				chunkedState, err = processor.ProcessFile(genesisStateFile)
			}

			// Log the final resume decision for debugging
			if isResume {
				logger.Info("Resume decision made",
					"is_resume", isResume,
					"last_successful_correlation_id", lastSuccessfulCorrelationID,
					"will_process_file", lastSuccessfulCorrelationID != "")
			}

			if err != nil {
				return fmt.Errorf("failed to process genesis file: %w", err)
			}

			// Optimize chunks using simulation AFTER file processing to ensure accurate gas estimation
			// This ensures we estimate gas based on the actual data that will be processed (after resume filtering)
			logger.Info("Optimizing chunks using simulation")
			logger.Info("Pre-optimization chunk info",
				"total_chunks", len(processor.chunks),
				"total_ledgers", chunkedState.TotalLedgers,
				"total_entries", chunkedState.TotalEntries)

			// Use stored gas costs if available (for resume), otherwise run simulations
			if isResume && existingStatus != nil && existingStatus.GasCosts != nil {
				logger.Info("Using stored gas costs for chunk optimization during resume")
				err = processor.optimizeChunksUsingStoredCosts(existingStatus.GasCosts)
			} else {
				logger.Info("Running gas cost estimation for fresh import")
				err = processor.optimizeChunksUsingSimulation(clientCtx, cmd)
			}
			if err != nil {
				return fmt.Errorf("failed to optimize chunks: %w", err)
			}

			// Update chunkedState with optimized chunks
			chunkedState.Chunks = processor.chunks
			chunkedState.TotalChunks = len(processor.chunks)

			logger.Info("Post-optimization chunk info",
				"total_chunks", len(processor.chunks),
				"total_ledgers", chunkedState.TotalLedgers,
				"total_entries", chunkedState.TotalEntries)

			logger.Info("Import summary",
				"import_id", importID,
				"total_ledgers", chunkedState.TotalLedgers,
				"total_entries", chunkedState.TotalEntries,
				"total_chunks", chunkedState.TotalChunks,
				"max_chunk_size_bytes", config.MaxChunkSizeBytes,
				"max_tx_size_bytes", config.MaxTxSizeBytes,
				"status_check_command", fmt.Sprintf("provenanced query ledger bulk-import-status %s --chain-id %s", importID, clientCtx.ChainID))

			// Initialize or update local status file
			var status *LocalBulkImportStatus
			if isResume {
				// Use existing status but update for resume
				status = existingStatus
				status.Status = "in_progress"
				status.UpdatedAt = time.Now().Format(time.RFC3339)
				status.ErrorMessage = "" // Clear any previous error message
				// Preserve stored gas costs for reuse
				if status.GasCosts != nil {
					logger.Info("Preserving stored gas costs for resume",
						"ledger_with_key_gas", status.GasCosts.LedgerWithKeyGas,
						"entry_gas", status.GasCosts.EntryGas)
				}
			} else {
				// Create new status
				status = &LocalBulkImportStatus{
					ImportID:        importID,
					TotalChunks:     chunkedState.TotalChunks,
					CompletedChunks: 0,
					TotalLedgers:    chunkedState.TotalLedgers,
					TotalEntries:    chunkedState.TotalEntries,
					Status:          "pending",
					CreatedAt:       time.Now().Format(time.RFC3339),
					UpdatedAt:       time.Now().Format(time.RFC3339),
					FileHash:        fileHash,
				}
			}
			_ = writeLocalBulkImportStatus(status)

			// Ask for confirmation (only if not resuming or if user wants to confirm resume)
			confirm, err := cmd.Flags().GetBool("yes")
			if err != nil {
				return err
			}

			if !confirm {
				if isResume {
					fmt.Printf("Resume import from chunk %d/%d? (y/N): ", startChunkIndex+1, chunkedState.TotalChunks)
				} else {
					fmt.Print("Proceed with chunked import? (y/N): ")
				}
				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
					logger.Info("Import cancelled by user")
					return nil
				}
			}

			// Update status to in_progress if not already
			if status.Status != "in_progress" {
				status.Status = "in_progress"
				status.UpdatedAt = time.Now().Format(time.RFC3339)
				_ = writeLocalBulkImportStatus(status)
			}

			// Process chunks starting from the resume point
			logger.Info("Starting chunk processing",
				"total_chunks", len(chunkedState.Chunks),
				"start_chunk", startChunkIndex+1,
				"is_resume", isResume,
				"last_correlation_id", lastSuccessfulCorrelationID)

			for i := startChunkIndex; i < len(chunkedState.Chunks); i++ {
				chunk := chunkedState.Chunks[i]
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

				// Get gas estimate for this chunk
				var chunkGas uint64
				if status.GasCosts != nil {
					// Use stored gas costs to estimate gas for this chunk
					gas := estimateChunkGasFromCosts(chunk, status.GasCosts)
					chunkGas = uint64(gas)
					logger.Info("Using stored gas costs for chunk", "chunk_index", i+1, "gas", chunkGas)
				} else {
					// Simulate to get gas estimate and store the costs for future use
					gas, err := simulateChunkGas(chunk, clientCtx, cmd)
					chunkGas = uint64(gas)
					if err != nil {
						status.Status = "failed"
						status.ErrorMessage = fmt.Sprintf("failed to estimate gas for chunk %d: %v", i+1, err)
						status.UpdatedAt = time.Now().Format(time.RFC3339)
						_ = writeLocalBulkImportStatus(status)
						return fmt.Errorf("failed to estimate gas for chunk %d: %w", i+1, err)
					}

					// Store gas costs for future use (only on first chunk or if not already stored)
					if status.GasCosts == nil {
						// Extract gas costs from representative simulations
						logger.Info("Estimating gas costs from representative simulations")
						gasCosts, err := estimateGasCosts(chunkedState.Chunks, clientCtx, cmd, logger)
						if err != nil {
							logger.Warn("Failed to estimate gas costs, using fallback values", "error", err)
							// Fallback to reasonable defaults if estimation fails
							status.GasCosts = &GasCosts{
								LedgerWithKeyGas: 100000,
								EntryGas:         5000,
							}
						} else {
							status.GasCosts = gasCosts
							logger.Info("Stored gas costs for future use",
								"ledger_with_key_gas", gasCosts.LedgerWithKeyGas,
								"entry_gas", gasCosts.EntryGas)
						}
						_ = writeLocalBulkImportStatus(status)
					}
				}

				// Validate chunk size against transaction limits
				if chunkSize > config.MaxTxSizeBytes {
					status.Status = "failed"
					status.ErrorMessage = fmt.Sprintf("chunk %d exceeds maximum transaction size: %d bytes > %d bytes", i+1, chunkSize, config.MaxTxSizeBytes)
					status.UpdatedAt = time.Now().Format(time.RFC3339)
					_ = writeLocalBulkImportStatus(status)
					return fmt.Errorf("chunk %d exceeds maximum transaction size: %d bytes > %d bytes", i+1, chunkSize, config.MaxTxSizeBytes)
				}

				// Validate chunk gas against gas limits
				if chunkGas > uint64(config.MaxGasPerTx-100000) { // 100k gas safety margin
					status.Status = "failed"
					status.ErrorMessage = fmt.Sprintf("chunk %d exceeds maximum gas limit: %d gas > %d gas", i+1, chunkGas, config.MaxGasPerTx-100000)
					status.UpdatedAt = time.Now().Format(time.RFC3339)
					_ = writeLocalBulkImportStatus(status)
					return fmt.Errorf("chunk %d exceeds maximum gas limit: %d gas > %d gas", i+1, chunkGas, config.MaxGasPerTx-100000)
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

				msg := &types.MsgBulkImportRequest{
					Authority:    clientCtx.FromAddress.String(),
					GenesisState: chunk,
				}

				// Check if this chunk was already processed (for resume safety)
				if isResume && status.LastAttemptedChunk != nil {
					lastChunk := status.LastAttemptedChunk
					if lastChunk.Confirmed {
						// Last chunk was confirmed, we can continue from the next chunk
						logger.Info("Last chunk was confirmed, continuing from next chunk",
							"last_first_correlation_id", lastChunk.FirstCorrelationID,
							"last_last_correlation_id", lastChunk.LastCorrelationID,
							"last_tx_hash", lastChunk.TransactionHash)
					} else if lastChunk.TransactionHash != "" {
						// Last chunk was sent but not confirmed - check if it actually succeeded
						logger.Info("Checking if last chunk was actually processed",
							"last_first_correlation_id", lastChunk.FirstCorrelationID,
							"last_last_correlation_id", lastChunk.LastCorrelationID,
							"last_tx_hash", lastChunk.TransactionHash)

						processed, err := checkTransactionAlreadyProcessed(clientCtx, lastChunk.TransactionHash, logger)
						if err != nil {
							logger.Warn("Failed to check transaction status, proceeding anyway", "error", err)
						} else if processed {
							// The transaction actually succeeded, mark it as confirmed
							logger.Info("Last chunk was actually processed successfully, marking as confirmed",
								"last_tx_hash", lastChunk.TransactionHash)
							status.LastAttemptedChunk.Confirmed = true
							status.UpdatedAt = time.Now().Format(time.RFC3339)
							_ = writeLocalBulkImportStatus(status)
						} else {
							// The transaction really failed, we can retry this chunk
							logger.Info("Last chunk was not processed, will retry",
								"last_tx_hash", lastChunk.TransactionHash)
						}
					}
				}

				// Broadcast transaction and check its status
				txHash, err := broadcastAndCheckTx(clientCtx, cmd, msg, i+1, logger)

				if err != nil {
					status.Status = "failed"
					status.ErrorMessage = fmt.Sprintf("failed to process chunk %d: %v", i+1, err)
					status.UpdatedAt = time.Now().Format(time.RFC3339)
					_ = writeLocalBulkImportStatus(status)
					return fmt.Errorf("failed to process chunk %d: %w", i+1, err)
				}

				// Update LastAttemptedChunk immediately after successful broadcast
				firstCorrelationID, lastCorrelationID := getCorrelationIDRangeFromChunk(chunk)
				if firstCorrelationID != "" && txHash != "" {
					status.LastAttemptedChunk = &ChunkStatus{
						FirstCorrelationID: firstCorrelationID,
						LastCorrelationID:  lastCorrelationID,
						Confirmed:          false,  // Will be set to true after confirmation
						TransactionHash:    txHash, // Set immediately after successful broadcast
					}
					logger.Info("Updated last attempted chunk status after successful broadcast",
						"chunk_index", i+1,
						"first_correlation_id", firstCorrelationID,
						"last_correlation_id", lastCorrelationID,
						"tx_hash", txHash)

					// Write status immediately to capture the transaction hash BEFORE waiting for confirmation
					// This ensures the tx hash is saved even if the import is canceled during confirmation wait
					err = writeLocalBulkImportStatus(status)
					if err != nil {
						logger.Warn("Failed to write status file after broadcast", "error", err)
					} else {
						logger.Info("Status file written with transaction hash before confirmation wait",
							"chunk_index", i+1,
							"tx_hash", txHash,
							"status_file", statusFileName(status.ImportID))
					}
				}

				// Wait for transaction confirmation
				err = waitForTransactionConfirmation(clientCtx, cmd, i, len(chunkedState.Chunks), logger)
				if err != nil {
					status.Status = "failed"
					status.ErrorMessage = fmt.Sprintf("failed to wait for transaction confirmation for chunk %d: %v", i+1, err)
					status.UpdatedAt = time.Now().Format(time.RFC3339)
					_ = writeLocalBulkImportStatus(status)
					return fmt.Errorf("failed to wait for transaction confirmation for chunk %d: %w", i+1, err)
				}

				// Mark last attempted chunk as confirmed after successful transaction confirmation
				if status.LastAttemptedChunk != nil {
					status.LastAttemptedChunk.Confirmed = true
					logger.Info("Marked last attempted chunk as confirmed",
						"chunk_index", i+1,
						"first_correlation_id", status.LastAttemptedChunk.FirstCorrelationID,
						"last_correlation_id", status.LastAttemptedChunk.LastCorrelationID)
				}

				// Update status with last successful correlation ID
				status.LastSuccessfulCorrelationID = lastCorrelationID
				if lastCorrelationID != "" {
					logger.Info("Updated last successful correlation ID", "correlation_id", lastCorrelationID)
				}

				logger.Info("Chunk completed successfully", "chunk_index", i+1, "total_chunks", chunkedState.TotalChunks)
				status.CompletedChunks = i + 1
				status.UpdatedAt = time.Now().Format(time.RFC3339)
				_ = writeLocalBulkImportStatus(status)
			}

			logger.Info("All chunks processed successfully", "total_chunks_processed", len(chunkedState.Chunks))
			status.Status = "completed"
			status.UpdatedAt = time.Now().Format(time.RFC3339)
			_ = writeLocalBulkImportStatus(status)

			logger.Info("Chunked bulk import completed successfully",
				"import_id", importID,
				"status_check_command", fmt.Sprintf("provenanced query ledger bulk-import-status %s --chain-id %s", importID, clientCtx.ChainID))

			return nil
		},
	}

	cmd.Flags().String("import-id", "", "Explicit import ID to use for status tracking (advanced)")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdBulkImportStatus creates a command to check bulk import status
func CmdBulkImportStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bulk-import-status <import_id>",
		Aliases: []string{"bis"},
		Short:   "Check the status of a bulk import operation",
		Example: `$ provenanced query ledger bulk-import-status import_1234567890 --chain-id testing`,
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
