package cli

import (
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
	sdk "github.com/cosmos/cosmos-sdk/types"
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

// scanLedgerAndEntryCounts scans the genesis file for total ledgers and entries efficiently.
func scanLedgerAndEntryCounts(filename string) (int, int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)

	// Expect root object
	token, err := decoder.Token()
	if err != nil {
		return 0, 0, err
	}
	if delim, ok := token.(json.Delim); !ok || delim != '{' {
		return 0, 0, fmt.Errorf("expected root object, got %v", token)
	}

	var numLedgers, numEntries int

	// Find ledgerToEntries array
	for decoder.More() {
		t, err := decoder.Token()
		if err != nil {
			return numLedgers, numEntries, err
		}
		fieldName, ok := t.(string)
		if !ok {
			return numLedgers, numEntries, fmt.Errorf("expected field name, got %v", t)
		}
		if fieldName == "ledgerToEntries" || fieldName == "ledger_to_entries" {
			// Expect array start
			token, err = decoder.Token()
			if err != nil {
				return numLedgers, numEntries, err
			}
			if delim, ok := token.(json.Delim); !ok || delim != '[' {
				return numLedgers, numEntries, fmt.Errorf("expected array start, got %v", token)
			}
			// Iterate through array
			for decoder.More() {
				var lte types.LedgerToEntries
				if err := decoder.Decode(&lte); err != nil {
					return numLedgers, numEntries, err
				}
				numLedgers++
				numEntries += len(lte.Entries)
			}
			break
		} else {
			// Skip unknown fields
			_ = skipJSONValue(decoder)
		}
	}
	return numLedgers, numEntries, nil
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

			// // Get validate-tx flag
			// validateTxEnabled, _ := cmd.Flags().GetBool("validate-tx")

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
			status, err := readLocalBulkImportStatus(importID)
			isResume := false
			lastSuccessfulCorrelationID := ""

			if err != nil {
				logger.Info("No existing status file found, starting fresh import", "error", err)
			} else if status != nil {
				logger.Info("Existing status file found",
					"status", status.Status,
					"completed_chunks", status.CompletedChunks,
					"total_ledgers", status.TotalLedgers,
					"total_entries", status.TotalEntries,
					"file_hash", status.FileHash,
					"current_file_hash", fileHash)
			}

			// After determining importID and fileHash, and before the scan for total ledgers/entries:
			if status == nil {
				logger.Info("No existing status file found, creating new status file", "import_id", importID)
				status = &LocalBulkImportStatus{
					ImportID:  importID,
					FileHash:  fileHash,
					Status:    "pending",
					CreatedAt: time.Now().Format(time.RFC3339),
					UpdatedAt: time.Now().Format(time.RFC3339),
				}
				_ = writeLocalBulkImportStatus(status)
			}

			// Determine resume mode and starting position
			if err == nil && status != nil && status.LastAttemptedChunk != nil {
				// Status file exists with LastAttemptedChunk - check if we can resume
				switch status.Status {
				case "completed":
					logger.Info("Import already completed", "import_id", importID)
					return nil
				case "failed", "in_progress":
					// Validate that the existing status matches our current import
					logger.Info("Checking if existing status matches current import",
						"existing_ledgers", status.TotalLedgers,
						"existing_entries", status.TotalEntries,
						"existing_file_hash", status.FileHash,
						"current_file_hash", fileHash)

					// For resume, we only check file hash since we'll reprocess the file
					if status.FileHash == fileHash {

						isResume = true
						lastSuccessfulCorrelationID = status.LastSuccessfulCorrelationID

						logger.Info("Resuming import from previous state",
							"import_id", importID,
							"completed_chunks", status.CompletedChunks,
							"status", status.Status,
							"last_correlation_id", lastSuccessfulCorrelationID,
							"file_hash", fileHash)

						// Log resume details
						if status.LastAttemptedChunk != nil {
							logger.Info("Resume details",
								"last_attempted_first_correlation_id", status.LastAttemptedChunk.FirstCorrelationID,
								"last_attempted_last_correlation_id", status.LastAttemptedChunk.LastCorrelationID,
								"last_attempted_confirmed", status.LastAttemptedChunk.Confirmed)
						}

						// Handle unconfirmed last attempted chunk
						if !status.LastAttemptedChunk.Confirmed {
							logger.Info("Last chunk was unconfirmed, checking if it was actually processed",
								"last_tx_hash", status.LastAttemptedChunk.TransactionHash)

							processed, err := validateTransactionByHash(clientCtx, status.LastAttemptedChunk.TransactionHash, logger)
							if err != nil {
								logger.Warn("Failed to check transaction status, proceeding anyway", "error", err)
							}
							if processed {
								// The transaction actually succeeded, mark it as confirmed
								logger.Info("Last chunk was actually processed successfully, marking as confirmed",
									"last_tx_hash", status.LastAttemptedChunk.TransactionHash)
								status.LastAttemptedChunk.Confirmed = true
								status.UpdatedAt = time.Now().Format(time.RFC3339)
								_ = writeLocalBulkImportStatus(status)

								// Start from the next correlation ID (after the last one)
								lastSuccessfulCorrelationID = status.LastAttemptedChunk.LastCorrelationID
							} else {
								// The transaction really failed, we need to find the next correlation ID after the last successfully processed one
								// Since we don't have a transaction hash, we need to determine where to resume from
								if status.LastSuccessfulCorrelationID != "" {
									// We have a last successful correlation ID, so we should start from the next one
									logger.Info("Last chunk was not processed, scanning file to find next correlation ID after",
										"last_successful_correlation_id", status.LastSuccessfulCorrelationID)

									nextCorrelationID, err := findNextCorrelationIDAfter(genesisStateFile, status.LastSuccessfulCorrelationID, logger)
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
										"first_correlation_id", status.LastAttemptedChunk.FirstCorrelationID)

									nextCorrelationID, err := findNextCorrelationIDAfter(genesisStateFile, status.LastAttemptedChunk.FirstCorrelationID, logger)
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
							lastSuccessfulCorrelationID = status.LastAttemptedChunk.LastCorrelationID
						}

						if status.Status == "failed" {
							logger.Info("Previous import failed", "error", status.ErrorMessage)
						}
					} else {
						logger.Warn("Existing status file found but file hash doesn't match - starting fresh import",
							"existing_file_hash", status.FileHash,
							"current_file_hash", fileHash)
					}
				}
			}

			// Scan for total ledgers and entries if not already set in status
			if (status.TotalLedgers == 0 || status.TotalEntries == 0) && genesisStateFile != "" {
				logger.Info("Scanning file for total ledgers and entries for progress tracking")
				numLedgers, numEntries, err := scanLedgerAndEntryCounts(genesisStateFile)
				if err != nil {
					logger.Warn("Failed to scan file for total ledgers/entries", "error", err)
				} else {
					logger.Info("File scan complete", "total_ledgers", numLedgers, "total_entries", numEntries)
					if status != nil {
						status.TotalLedgers = numLedgers
						status.TotalEntries = numEntries
						status.UpdatedAt = time.Now().Format(time.RFC3339)
						_ = writeLocalBulkImportStatus(status)
					}
				}
			}

			// For gas estimation and flat fee querying, we still need to process a sample of the file
			// Use the original processor to get representative chunks for gas estimation
			var gasCosts *GasCosts
			var flatFeeInfo *FlatFeeInfo

			// Query flat fee for bulk import
			logger.Info("Querying flat fee for bulk import")
			flatFeeInfo, err = getFlatFeeForBulkImport(clientCtx)
			if err != nil {
				logger.Warn("Failed to query flat fee, using default fee", "error", err)
				// Fallback to default fee if query fails
				flatFeeInfo = &FlatFeeInfo{
					FeeAmount: sdk.NewCoins(sdk.NewInt64Coin("nhash", 1000000)), // 1 nhash default
					MsgType:   "/provenance.ledger.v1.MsgBulkImportRequest",
				}
			} else {
				logFlatFeeInfo(flatFeeInfo, logger)
			}

			if isResume && status != nil && status.GasCosts != nil {
				// Use stored gas costs for resume
				logger.Info("Using stored gas costs for resume")
				gasCosts = status.GasCosts
				// Use stored flat fee info if available
				if status.FlatFeeInfo != nil {
					flatFeeInfo = status.FlatFeeInfo
					logger.Info("Using stored flat fee info for resume",
						"msg_type", status.FlatFeeInfo.MsgType,
						"fee_amount", status.FlatFeeInfo.FeeAmount.String())
				}
			} else {
				// --- Initial gas estimation by streaming in just enough data for CalculateTxFees ---
				logger.Info("Performing initial gas estimation by streaming minimal chunk for simulation")
				streamingProcessor := NewStreamingChunkProcessor(config, logger)
				err := streamingProcessor.OpenFile(genesisStateFile, "")
				if err != nil {
					return fmt.Errorf("failed to open file for initial gas estimation: %w", err)
				}
				defer streamingProcessor.Close()

				// Use fallback gas costs for initial chunking to ensure we get a reasonable minimal chunk
				fallbackGasCosts := &GasCosts{
					LedgerWithKeyGas: 70000,
					EntryGas:         100000,
				}

				// Get the first chunk using gas limits to ensure it's not too large
				minimalChunk, err := streamingProcessor.NextChunkWithGasLimit(fallbackGasCosts)
				if err != nil {
					return fmt.Errorf("failed to get minimal chunk for gas estimation: %w", err)
				}

				// Continue getting chunks until we find one with 1 ledger entry that has 1 ledger and 50+ entries
				for {
					ledgerCount := len(minimalChunk.LedgerToEntries)
					entryCount := 0
					for _, lte := range minimalChunk.LedgerToEntries {
						entryCount += len(lte.Entries)
					}

					// Check if this chunk meets our criteria: 1 ledger entry with 1 ledger and 50+ entries
					if ledgerCount == 1 && entryCount >= 50 {
						logger.Info("Found suitable chunk for gas estimation",
							"ledger_count", ledgerCount,
							"entry_count", entryCount)
						break
					}

					// Get the next chunk
					nextChunk, err := streamingProcessor.NextChunkWithGasLimit(fallbackGasCosts)
					if err == io.EOF {
						logger.Warn("Reached end of file without finding suitable chunk, using current chunk")
						break
					}
					if err != nil {
						logger.Warn("Failed to get next chunk for gas estimation, using current chunk", "error", err)
						break
					}

					minimalChunk = nextChunk
					logger.Info("Trying next chunk for gas estimation",
						"ledger_count", ledgerCount,
						"entry_count", entryCount)
				}

				// Simulate gas for this minimal chunk
				gasCosts, err = EstimateGasCostsAccurately(minimalChunk, clientCtx, cmd, logger)
				if err != nil {
					logger.Warn("Failed to estimate gas costs accurately, using fallback values", "error", err)
					gasCosts = fallbackGasCosts
				}
				logger.Info("Initial gas costs estimated from minimal chunk",
					"ledger_with_key_gas", gasCosts.LedgerWithKeyGas,
					"entry_gas", gasCosts.EntryGas)

				// Validate gas costs against actual chunks if possible
				logger.Info("Validating gas cost estimates")
				validationProcessor := NewStreamingChunkProcessor(config, logger)
				err = validationProcessor.OpenFile(genesisStateFile, "")
				if err == nil {
					defer validationProcessor.Close()

					// Collect a few chunks for validation
					var validationChunks []*types.GenesisState
					for i := 0; i < 3; i++ {
						chunk, err := validationProcessor.NextChunkWithGasLimit(gasCosts)
						if err == io.EOF {
							break
						}
						if err != nil {
							logger.Warn("Failed to get chunk for validation", "error", err)
							break
						}
						validationChunks = append(validationChunks, chunk)
					}

					if len(validationChunks) > 0 {
						err = validateGasCosts(gasCosts, validationChunks, clientCtx, cmd, logger)
						if err != nil {
							logger.Warn("Gas cost validation failed, but continuing with estimates", "error", err)
						}
					}
				}

				// Optionally, store these in the status file for future resumes
				if status != nil {
					status.GasCosts = gasCosts
					status.FlatFeeInfo = flatFeeInfo
					status.UpdatedAt = time.Now().Format(time.RFC3339)
					_ = writeLocalBulkImportStatus(status)
				}
			}

			// Log the final resume decision for debugging
			if isResume {
				logger.Info("Resume decision made",
					"is_resume", isResume,
					"last_successful_correlation_id", lastSuccessfulCorrelationID,
					"will_process_file", lastSuccessfulCorrelationID != "")
			}

			// Handle case where no next correlation ID found during resume
			if isResume && lastSuccessfulCorrelationID == "" {
				logger.Info("No next correlation ID found, checking if import is complete")

				// Mark the import as completed since there's nothing more to process
				if status != nil {
					status.Status = "completed"
					status.UpdatedAt = time.Now().Format(time.RFC3339)
					_ = writeLocalBulkImportStatus(status)
					logger.Info("Import marked as completed - no more data to process")
				}
				return nil
			}

			logger.Info("Import summary",
				"import_id", importID,
				"max_chunk_size_bytes", config.MaxChunkSizeBytes,
				"max_tx_size_bytes", config.MaxTxSizeBytes,
				"status_check_command", fmt.Sprintf("provenanced query ledger bulk-import-status %s --chain-id %s", importID, clientCtx.ChainID))

			// Initialize or update local status file
			if isResume {
				// Use existing status but update for resume
				status.Status = "in_progress"
				status.UpdatedAt = time.Now().Format(time.RFC3339)
				status.ErrorMessage = "" // Clear any previous error message
				// Preserve stored gas costs for reuse
				if status.GasCosts != nil {
					logger.Info("Preserving stored gas costs for resume",
						"ledger_with_key_gas", status.GasCosts.LedgerWithKeyGas,
						"entry_gas", status.GasCosts.EntryGas)
				}
				// Preserve stored flat fee info for reuse
				if status.FlatFeeInfo != nil {
					logger.Info("Preserving stored flat fee info for resume",
						"msg_type", status.FlatFeeInfo.MsgType,
						"fee_amount", status.FlatFeeInfo.FeeAmount.String())
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
					fmt.Printf("Resume import from correlation ID %s? (y/N): ", lastSuccessfulCorrelationID)
				} else {
					fmt.Print("Proceed with streaming import? (y/N): ")
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

			// Initialize streaming processor
			streamingProcessor := NewStreamingChunkProcessor(config, logger)
			defer streamingProcessor.Close()

			// Open file for streaming
			err = streamingProcessor.OpenFile(genesisStateFile, lastSuccessfulCorrelationID)
			if err != nil {
				return fmt.Errorf("failed to open file for streaming: %w", err)
			}

			// Process chunks using streaming
			logger.Info("Starting streaming chunk processing",
				"is_resume", isResume,
				"last_correlation_id", lastSuccessfulCorrelationID)

			chunkIndex := 0
			for {
				chunk, err := streamingProcessor.NextChunkWithGasLimit(gasCosts)
				if err == io.EOF {
					logger.Info("Reached end of file")
					break
				}
				if err != nil {
					status.Status = "failed"
					status.ErrorMessage = fmt.Sprintf("failed to read chunk %d: %v", chunkIndex+1, err)
					status.UpdatedAt = time.Now().Format(time.RFC3339)
					_ = writeLocalBulkImportStatus(status)
					return fmt.Errorf("failed to read chunk %d: %w", chunkIndex+1, err)
				}

				chunkIndex++
				logger.Info("Processing chunk", "chunk_index", chunkIndex)

				// Validate chunk before processing
				if chunk == nil {
					status.Status = "failed"
					status.ErrorMessage = fmt.Sprintf("chunk %d is nil", chunkIndex)
					status.UpdatedAt = time.Now().Format(time.RFC3339)
					_ = writeLocalBulkImportStatus(status)
					return fmt.Errorf("chunk %d is nil", chunkIndex)
				}

				chunkEntries := 0
				for _, ledger := range chunk.LedgerToEntries {
					chunkEntries += len(ledger.Entries)
				}
				chunkSize := getChunkSizeBytes(chunk)

				// Get gas estimate for this chunk using the pre-estimated gas costs
				gas := estimateChunkGasFromCosts(chunk, gasCosts)
				chunkGas := uint64(gas)
				effectiveGasLimit := config.GetEffectiveGasLimit()
				logger.Info("Using estimated gas costs for chunk validation",
					"chunk_index", chunkIndex,
					"gas", chunkGas,
					"effective_gas_limit", effectiveGasLimit,
					"gas_margin", config.MaxGasPerTx-effectiveGasLimit)

				// Store gas costs and flat fee info in status for future use (only on first chunk)
				if status.GasCosts == nil {
					status.GasCosts = gasCosts
					logger.Info("Stored gas costs for future use",
						"ledger_with_key_gas", gasCosts.LedgerWithKeyGas,
						"entry_gas", gasCosts.EntryGas)
				}
				if status.FlatFeeInfo == nil {
					status.FlatFeeInfo = flatFeeInfo
					logger.Info("Stored flat fee info for future use",
						"msg_type", flatFeeInfo.MsgType,
						"fee_amount", flatFeeInfo.FeeAmount.String())
					_ = writeLocalBulkImportStatus(status)
				}

				// Validate chunk using flat fees and gas limits
				err = validateChunkWithFlatFees(chunk, int(chunkGas), config, logger)
				if err != nil {
					status.Status = "failed"
					status.ErrorMessage = fmt.Sprintf("chunk %d validation failed: %v", chunkIndex, err)
					status.UpdatedAt = time.Now().Format(time.RFC3339)
					_ = writeLocalBulkImportStatus(status)
					return fmt.Errorf("chunk %d validation failed: %w", chunkIndex, err)
				}

				logger.Info("Chunk details",
					"chunk_index", chunkIndex,
					"ledger_count", len(chunk.LedgerToEntries),
					"entry_count", chunkEntries,
					"size_bytes", chunkSize,
					"estimated_gas", chunkGas,
					"flat_fee", flatFeeInfo.FeeAmount.String())

				// Get fresh client context for each transaction to ensure proper sequence number handling
				clientCtx, err = client.GetClientTxContext(cmd)
				if err != nil {
					status.Status = "failed"
					status.ErrorMessage = fmt.Sprintf("failed to get client context for chunk %d: %v", chunkIndex, err)
					status.UpdatedAt = time.Now().Format(time.RFC3339)
					_ = writeLocalBulkImportStatus(status)
					return fmt.Errorf("failed to get client context for chunk %d: %w", chunkIndex, err)
				}

				msg := &types.MsgBulkImportRequest{
					Authority:    clientCtx.FromAddress.String(),
					GenesisState: chunk,
				}

				// // Check if this chunk was already processed (for resume safety)
				// if isResume && status.LastAttemptedChunk != nil {
				// 	lastChunk := status.LastAttemptedChunk
				// 	if lastChunk.Confirmed {
				// 		// Last chunk was confirmed, we can continue from the next chunk
				// 		logger.Info("Last chunk was confirmed, continuing from next chunk",
				// 			"last_first_correlation_id", lastChunk.FirstCorrelationID,
				// 			"last_last_correlation_id", lastChunk.LastCorrelationID,
				// 			"last_tx_hash", lastChunk.TransactionHash)
				// 	} else if lastChunk.TransactionHash != "" {
				// 		// Last chunk was sent but not confirmed - check if it actually succeeded
				// 		logger.Info("Checking if last chunk was actually processed",
				// 			"last_first_correlation_id", lastChunk.FirstCorrelationID,
				// 			"last_last_correlation_id", lastChunk.LastCorrelationID,
				// 			"last_tx_hash", lastChunk.TransactionHash)

				// 		processed, err := validateTransactionByHash(clientCtx, lastChunk.TransactionHash, logger)
				// 		if err != nil {
				// 			if validateTxEnabled {
				// 				return fmt.Errorf("failed to validate transaction: %w", err)
				// 			}
				// 			logger.Info("Transaction indexing disabled or unavailable, proceeding with resume", "tx_hash", lastChunk.TransactionHash)
				// 		} else if processed {
				// 			// The transaction actually succeeded, mark it as confirmed
				// 			logger.Info("Last chunk was actually processed successfully, marking as confirmed",
				// 				"last_tx_hash", lastChunk.TransactionHash)
				// 			status.LastAttemptedChunk.Confirmed = true
				// 			status.UpdatedAt = time.Now().Format(time.RFC3339)
				// 			_ = writeLocalBulkImportStatus(status)
				// 		} else {
				// 			// The transaction really failed, we can retry this chunk
				// 			logger.Info("Last chunk was not processed, will retry",
				// 				"last_tx_hash", lastChunk.TransactionHash)
				// 		}
				// 	}
				// }

				// Broadcast transaction and check its status
				txHash, err := broadcastAndCheckTx(clientCtx, cmd, msg, chunkIndex, logger, flatFeeInfo)

				if err != nil {
					status.Status = "failed"
					status.ErrorMessage = fmt.Sprintf("failed to process chunk %d: %v", chunkIndex, err)
					status.UpdatedAt = time.Now().Format(time.RFC3339)
					_ = writeLocalBulkImportStatus(status)
					return fmt.Errorf("failed to process chunk %d: %w", chunkIndex, err)
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
						"chunk_index", chunkIndex,
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
							"chunk_index", chunkIndex,
							"tx_hash", txHash,
							"status_file", statusFileName(status.ImportID))
					}
				}

				// Wait for transaction confirmation
				err = waitForTransactionConfirmation(clientCtx, cmd, chunkIndex-1, 0, logger, txHash) // chunkIndex-1 because waitForTransactionConfirmation expects 0-based index
				if err != nil {
					status.Status = "failed"
					status.ErrorMessage = fmt.Sprintf("failed to wait for transaction confirmation for chunk %d: %v", chunkIndex, err)
					status.UpdatedAt = time.Now().Format(time.RFC3339)
					_ = writeLocalBulkImportStatus(status)
					return fmt.Errorf("failed to wait for transaction confirmation for chunk %d: %w", chunkIndex, err)
				}

				// Mark last attempted chunk as confirmed after successful transaction confirmation
				if status.LastAttemptedChunk != nil {
					status.LastAttemptedChunk.Confirmed = true
					logger.Info("Marked last attempted chunk as confirmed",
						"chunk_index", chunkIndex,
						"first_correlation_id", status.LastAttemptedChunk.FirstCorrelationID,
						"last_correlation_id", status.LastAttemptedChunk.LastCorrelationID)
					// Update completed ledgers and entries
					stats := streamingProcessor.GetStats()
					status.CompletedLedgers = stats.TotalLedgers
					status.CompletedEntries = stats.TotalEntries
					// Log the current status after confirmation
					logger.Info("Status after tx confirmation",
						"chunk_index", chunkIndex,
						"completed_chunks", status.CompletedChunks,
						"completed_ledgers", status.CompletedLedgers,
						"completed_entries", status.CompletedEntries,
						"total_ledgers", status.TotalLedgers,
						"total_entries", status.TotalEntries,
						"status", status.Status)
				}

				// Update status with last successful correlation ID
				status.LastSuccessfulCorrelationID = lastCorrelationID
				if lastCorrelationID != "" {
					logger.Info("Updated last successful correlation ID", "correlation_id", lastCorrelationID)
				}

				// Update totals from streaming processor stats
				stats := streamingProcessor.GetStats()
				status.CompletedLedgers = stats.TotalLedgers
				status.CompletedEntries = stats.TotalEntries

				logger.Info("Chunk completed successfully", "chunk_index", chunkIndex)
				status.CompletedChunks = chunkIndex
				status.UpdatedAt = time.Now().Format(time.RFC3339)
				_ = writeLocalBulkImportStatus(status)
			}

			logger.Info("All chunks processed successfully", "total_chunks_processed", chunkIndex)
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
	// cmd.Flags().Bool("validate-tx", false, "Enable transaction validation during resume (checks if transactions were actually processed)")
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
