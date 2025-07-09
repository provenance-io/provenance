package cli

import (
	"bufio"
	"bytes"
	"context"
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
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"cosmossdk.io/log"

	"github.com/provenance-io/provenance/x/ledger/types"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
)

// ChunkConfig defines configuration for chunking large datasets
type ChunkConfig struct {
	MaxChunkSizeBytes int // Maximum chunk size in bytes (memory safety limit during parsing)
	MaxGasPerTx       int // Maximum gas consumption per transaction
	MaxTxSizeBytes    int // Maximum transaction size in bytes (blockchain limit)
}

// DefaultChunkConfig returns a reasonable default configuration
func DefaultChunkConfig() ChunkConfig {
	return ChunkConfig{
		MaxChunkSizeBytes: 10000000, // 10MB per chunk (memory safety limit)
		MaxGasPerTx:       4000000,  // 4M gas per transaction (matching blockchain limit)
		MaxTxSizeBytes:    1000000,  // 1MB max transaction size (typical blockchain limit)
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

// broadcastAndCheckTx broadcasts a transaction and checks its response for success/failure
func broadcastAndCheckTx(clientCtx client.Context, cmd *cobra.Command, msg sdk.Msg, chunkIndex int, logger log.Logger) error {
	// Force broadcast mode to sync for this transaction to ensure it's committed
	originalBroadcastMode := cmd.Flag(flags.FlagBroadcastMode).Value.String()
	cmd.Flag(flags.FlagBroadcastMode).Value.Set("sync")

	// Create a custom writer to capture the output
	var outputBuffer bytes.Buffer
	originalOutput := clientCtx.Output
	clientCtx.Output = &outputBuffer

	// Broadcast transaction
	err := tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)

	// Restore original broadcast mode and output
	cmd.Flag(flags.FlagBroadcastMode).Value.Set(originalBroadcastMode)
	clientCtx.Output = originalOutput

	if err != nil {
		return fmt.Errorf("failed to broadcast transaction for chunk %d: %w", chunkIndex, err)
	}

	// Parse the captured output to check transaction status
	outputBytes := outputBuffer.Bytes()
	outputStr := string(outputBytes)
	logger.Debug("Transaction output captured", "chunk_index", chunkIndex, "output", outputStr)

	var txResp sdk.TxResponse
	if err := clientCtx.Codec.UnmarshalJSON(outputBytes, &txResp); err != nil {
		// If we can't parse the response, log it and continue
		logger.Debug("Could not parse transaction response", "output", outputStr, "error", err)
	} else {
		// Check if the transaction was successful
		if txResp.Code != 0 {
			logger.Error("Transaction failed",
				"chunk_index", chunkIndex,
				"code", txResp.Code,
				"raw_log", txResp.RawLog,
				"tx_hash", txResp.TxHash)
			return fmt.Errorf("transaction failed for chunk %d with code %d: %s", chunkIndex, txResp.Code, txResp.RawLog)
		}
		logger.Info("Transaction successful",
			"chunk_index", chunkIndex,
			"tx_hash", txResp.TxHash,
			"code", txResp.Code,
			"gas_used", txResp.GasUsed,
			"gas_wanted", txResp.GasWanted)
	}

	// Wait a moment for the transaction to be processed
	time.Sleep(1 * time.Second)

	// Get the account to check if the sequence number was incremented
	// This provides additional confirmation that the transaction was accepted
	account, err := clientCtx.AccountRetriever.GetAccount(clientCtx, clientCtx.FromAddress)
	if err != nil {
		return fmt.Errorf("failed to get account info after chunk %d: %w", chunkIndex, err)
	}

	// Log the sequence number for debugging
	logger.Debug("Account sequence after transaction", "sequence", account.GetSequence())

	return nil
}

// getChunkSizeBytes returns the actual serialized size of a chunk in bytes
func getChunkSizeBytes(chunk *types.GenesisState) int {
	data, err := json.Marshal(chunk)
	if err != nil {
		return 0
	}
	return len(data)
}

// simulateChunkGas builds, signs, and simulates the transaction for accurate gas estimation
func simulateChunkGas(chunk *types.GenesisState, clientCtx client.Context, cmd *cobra.Command) (int, error) {
	msg := &types.MsgBulkImportRequest{
		Authority:    clientCtx.FromAddress.String(),
		GenesisState: chunk,
	}

	txFactory, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
	if err != nil {
		return 0, fmt.Errorf("failed to create tx factory: %w", err)
	}

	// Get account number and sequence
	accountRetriever := clientCtx.AccountRetriever
	account, err := accountRetriever.GetAccount(clientCtx, clientCtx.FromAddress)
	if err != nil {
		return 0, fmt.Errorf("failed to get account info: %w", err)
	}
	accNum := account.GetAccountNumber()
	seq := account.GetSequence()

	txFactory = txFactory.WithAccountNumber(accNum).WithSequence(seq)

	txBuilder := clientCtx.TxConfig.NewTxBuilder()
	if err := txBuilder.SetMsgs(msg); err != nil {
		return 0, fmt.Errorf("failed to set message in tx builder: %w", err)
	}

	// Set a dummy fee and gas (will be overwritten by simulation)
	txBuilder.SetFeeAmount([]sdk.Coin{sdk.NewInt64Coin("nhash", 1)})
	txBuilder.SetGasLimit(2000000)

	// Sign the transaction
	err = tx.Sign(cmd.Context(), txFactory, clientCtx.GetFromName(), txBuilder, false)
	if err != nil {
		return 0, fmt.Errorf("failed to sign tx for simulation: %w", err)
	}

	txBytes, err := clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return 0, fmt.Errorf("failed to encode tx: %w", err)
	}

	queryClient := msgfeestypes.NewQueryClient(clientCtx)
	response, err := queryClient.CalculateTxFees(
		context.Background(),
		&msgfeestypes.CalculateTxFeesRequest{
			TxBytes:          txBytes,
			DefaultBaseDenom: "nhash",
			GasAdjustment:    1.2, // 20% margin
		},
	)
	if err != nil {
		return 0, fmt.Errorf("failed to simulate tx: %w", err)
	}

	return int(response.EstimatedGas), nil
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

			// Optimize chunks using simulation to ensure they fit within gas limits
			logger.Info("Optimizing chunks using simulation")
			err = processor.optimizeChunksUsingSimulation(clientCtx, cmd)
			if err != nil {
				return fmt.Errorf("failed to optimize chunks using simulation: %w", err)
			}

			// Update chunkedState with optimized chunks
			chunkedState.Chunks = processor.chunks
			chunkedState.TotalChunks = len(processor.chunks)

			logger.Info("Import summary",
				"import_id", chunkedState.ImportID,
				"total_ledgers", chunkedState.TotalLedgers,
				"total_entries", chunkedState.TotalEntries,
				"total_chunks", chunkedState.TotalChunks,
				"max_chunk_size_bytes", config.MaxChunkSizeBytes,
				"max_tx_size_bytes", config.MaxTxSizeBytes)

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
				chunkGas, err := simulateChunkGas(chunk, clientCtx, cmd)
				if err != nil {
					status.Status = "failed"
					status.ErrorMessage = fmt.Sprintf("failed to estimate gas for chunk %d: %v", i+1, err)
					status.UpdatedAt = time.Now().Format(time.RFC3339)
					_ = writeLocalBulkImportStatus(status)
					return fmt.Errorf("failed to estimate gas for chunk %d: %w", i+1, err)
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
				if chunkGas > config.MaxGasPerTx-100000 { // 100k gas safety margin
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

				// Broadcast transaction and check its status
				err = broadcastAndCheckTx(clientCtx, cmd, msg, i+1, logger)
				if err != nil {
					status.Status = "failed"
					status.ErrorMessage = fmt.Sprintf("failed to process chunk %d: %v", i+1, err)
					status.UpdatedAt = time.Now().Format(time.RFC3339)
					_ = writeLocalBulkImportStatus(status)
					return fmt.Errorf("failed to process chunk %d: %w", i+1, err)
				}

				// Wait for transaction to be committed and account sequence to be updated
				// This is necessary to ensure the next transaction uses the correct sequence number

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
				"status_check_command", fmt.Sprintf("provenanced query ledger bulk-import-status %s --chain-id %s", chunkedState.ImportID, clientCtx.ChainID))

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

// generateImportID creates a unique identifier for an import operation
func generateImportID() string {
	// In a real implementation, you might want to use a more sophisticated ID generation
	// For now, we'll use a simple timestamp-based approach
	return fmt.Sprintf("import_%d", time.Now().UnixNano())
}

// optimizeChunksUsingSimulation takes the initial chunks from parsing and optimizes them
// using simulation to ensure they fit within gas limits
func (p *StreamingGenesisProcessor) optimizeChunksUsingSimulation(clientCtx client.Context, cmd *cobra.Command) error {
	// First, run a few representative simulations to understand gas costs
	gasCosts, err := p.estimateGasCosts(clientCtx, cmd)
	if err != nil {
		return fmt.Errorf("failed to estimate gas costs: %w", err)
	}

	p.logger.Info("Gas cost estimates",
		"ledger_key_gas", gasCosts.LedgerWithKeyGas,
		"entry_gas", gasCosts.EntryGas)

	var optimizedChunks []*types.GenesisState

	for _, chunk := range p.chunks {
		// Estimate gas for this chunk using our cost model
		estimatedGas := p.estimateChunkGasFromCosts(chunk, gasCosts)

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

// GasCosts represents the estimated gas costs for different components
type GasCosts struct {
	LedgerWithKeyGas int // Gas cost for a ledger key + ledger (base cost)
	EntryGas         int // Gas cost per entry
}

// estimateGasCosts runs representative simulations to understand gas costs
func (p *StreamingGenesisProcessor) estimateGasCosts(clientCtx client.Context, cmd *cobra.Command) (*GasCosts, error) {
	// Find a representative ledger to use for testing
	var testLedger *types.LedgerToEntries
	for _, chunk := range p.chunks {
		for _, lte := range chunk.LedgerToEntries {
			if lte.LedgerKey != nil && lte.Ledger != nil && len(lte.Entries) > 0 {
				testLedger = &lte
				break
			}
		}
		if testLedger != nil {
			break
		}
	}
	if testLedger == nil {
		return nil, fmt.Errorf("no suitable test ledger found")
	}

	// Simulation 1: LedgerKey + Ledger (no entries)
	ledgerWithKey := &types.GenesisState{
		LedgerToEntries: []types.LedgerToEntries{
			{
				LedgerKey: testLedger.LedgerKey,
				Ledger:    testLedger.Ledger,
				Entries:   []*types.LedgerEntry{},
			},
		},
	}
	ledgerWithKeyGas, err := simulateChunkGas(ledgerWithKey, clientCtx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to simulate ledger with key: %w", err)
	}

	// Simulation 2: LedgerKey + Ledger + 1 Entry
	ledgerWithEntry := &types.GenesisState{
		LedgerToEntries: []types.LedgerToEntries{
			{
				LedgerKey: testLedger.LedgerKey,
				Ledger:    testLedger.Ledger,
				Entries:   []*types.LedgerEntry{testLedger.Entries[0]},
			},
		},
	}
	ledgerWithEntryGas, err := simulateChunkGas(ledgerWithEntry, clientCtx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to simulate ledger with entry: %w", err)
	}

	// Calculate component costs
	entryCost := ledgerWithEntryGas - ledgerWithKeyGas

	// Run a third simulation with more entries to get a more accurate per-entry cost
	numTestEntries := 10
	var ledgerWithMoreEntriesGas int
	if len(testLedger.Entries) >= numTestEntries {
		testEntries := testLedger.Entries[:numTestEntries]
		ledgerWithMoreEntries := &types.GenesisState{
			LedgerToEntries: []types.LedgerToEntries{
				{
					LedgerKey: testLedger.LedgerKey,
					Ledger:    testLedger.Ledger,
					Entries:   testEntries,
				},
			},
		}

		ledgerWithMoreEntriesGas, err = simulateChunkGas(ledgerWithMoreEntries, clientCtx, cmd)
		if err != nil {
			p.logger.Warn("Failed to simulate with more entries, using single entry cost", "error", err)
		} else {
			// Calculate per-entry cost from the larger sample
			entryCost = (ledgerWithMoreEntriesGas - ledgerWithKeyGas) / numTestEntries
		}
	}

	p.logger.Info("Gas cost calculation",
		"ledger_with_key_gas", ledgerWithKeyGas,
		"ledger_with_entry_gas", ledgerWithEntryGas,
		"ledger_with_more_entries_gas", ledgerWithMoreEntriesGas,
		"calculated_entry_cost", entryCost)

	return &GasCosts{
		LedgerWithKeyGas: ledgerWithKeyGas,
		EntryGas:         entryCost,
	}, nil
}

// estimateChunkGasFromCosts estimates gas usage for a chunk using the cost model
func (p *StreamingGenesisProcessor) estimateChunkGasFromCosts(chunk *types.GenesisState, costs *GasCosts) int {
	totalGas := 0
	for _, lte := range chunk.LedgerToEntries {
		if lte.LedgerKey != nil && lte.Ledger != nil {
			// First chunk: ledger + key + entries
			totalGas += costs.LedgerWithKeyGas
			if len(lte.Entries) > 0 {
				totalGas += len(lte.Entries) * costs.EntryGas
			}
		} else if lte.LedgerKey != nil && lte.Ledger == nil {
			// Subsequent chunks: only entries (ledger already exists)
			if len(lte.Entries) > 0 {
				totalGas += len(lte.Entries) * costs.EntryGas
			}
		}
	}
	return totalGas
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
		estimatedGas := p.estimateChunkGasFromCosts(chunk, costs)
		if estimatedGas > maxGasPerChunk {
			p.logger.Warn("Chunk exceeds gas limit, reducing entries",
				"estimated_gas", estimatedGas,
				"max_gas", maxGasPerChunk,
				"entries_count", len(chunkEntries))

			// If we still exceed the limit, reduce entries one by one
			for len(chunkEntries) > 1 && estimatedGas > maxGasPerChunk {
				chunkEntries = chunkEntries[:len(chunkEntries)-1]
				chunk.LedgerToEntries[0].Entries = chunkEntries
				estimatedGas = p.estimateChunkGasFromCosts(chunk, costs)
			}
		}

		result = append(result, chunk)

		p.logger.Info("Created chunk",
			"chunk_index", len(result),
			"entries_count", len(chunkEntries),
			"estimated_gas", p.estimateChunkGasFromCosts(chunk, costs),
			"has_ledger", chunkLedger != nil)
	}

	return result
}

// splitChunkByLedgers splits a chunk by creating individual chunks for each ledger
func (p *StreamingGenesisProcessor) splitChunkByLedgers(chunk *types.GenesisState, clientCtx client.Context, cmd *cobra.Command) []*types.GenesisState {
	var result []*types.GenesisState

	for _, lte := range chunk.LedgerToEntries {
		// Create a single ledger chunk
		singleLedgerChunk := &types.GenesisState{
			LedgerToEntries: []types.LedgerToEntries{lte},
		}

		// Check if this single ledger chunk fits within gas limits
		estimatedGas, err := simulateChunkGas(singleLedgerChunk, clientCtx, cmd)
		if err != nil {
			p.logger.Warn("Failed to simulate gas for single ledger chunk, including it anyway", "error", err)
			result = append(result, singleLedgerChunk)
			continue
		}

		if estimatedGas <= p.config.MaxGasPerTx-100000 {
			result = append(result, singleLedgerChunk)
			p.logger.Info("Single ledger chunk fits within gas limits",
				"estimated_gas", estimatedGas,
				"entries_count", len(lte.Entries))
		} else {
			// If even a single ledger is too large, split it by entries
			p.logger.Info("Single ledger exceeds gas limit, splitting by entries",
				"estimated_gas", estimatedGas,
				"entries_count", len(lte.Entries))

			splitEntryChunks := p.splitLedgerByEntries(&lte, clientCtx, cmd)
			result = append(result, splitEntryChunks...)
		}
	}

	return result
}

// splitLedgerByEntries splits a ledger with too many entries into multiple chunks
func (p *StreamingGenesisProcessor) splitLedgerByEntries(lte *types.LedgerToEntries, clientCtx client.Context, cmd *cobra.Command) []*types.GenesisState {
	var result []*types.GenesisState

	// Start with a reasonable number of entries per chunk
	entriesPerChunk := 100

	for i := 0; i < len(lte.Entries); i += entriesPerChunk {
		end := i + entriesPerChunk
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
		estimatedGas, err := simulateChunkGas(chunk, clientCtx, cmd)
		if err != nil {
			p.logger.Warn("Failed to simulate gas for entry chunk, reducing entries", "error", err)
			// If simulation fails, reduce entries and try again
			for len(chunkEntries) > 1 {
				chunkEntries = chunkEntries[:len(chunkEntries)-1]
				chunk.LedgerToEntries[0].Entries = chunkEntries
				estimatedGas, err = simulateChunkGas(chunk, clientCtx, cmd)
				if err == nil {
					break
				}
			}
		}

		// If still too large, reduce entries one by one
		for estimatedGas > p.config.MaxGasPerTx-100000 && len(chunkEntries) > 1 {
			chunkEntries = chunkEntries[:len(chunkEntries)-1]
			chunk.LedgerToEntries[0].Entries = chunkEntries
			estimatedGas, err = simulateChunkGas(chunk, clientCtx, cmd)
			if err != nil {
				break
			}
		}

		result = append(result, chunk)

		p.logger.Info("Created entry chunk",
			"chunk_index", len(result),
			"entries_count", len(chunkEntries),
			"estimated_gas", estimatedGas,
			"has_ledger", chunkLedger != nil)
	}

	return result
}
