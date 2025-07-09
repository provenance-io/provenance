package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/spf13/cobra"

	"github.com/provenance-io/provenance/x/ledger/types"
)

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
			logger.Info("Starting chunk processing", "total_chunks", len(chunkedState.Chunks))
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

				// Wait for transaction confirmation
				err = waitForTransactionConfirmation(clientCtx, cmd, i, len(chunkedState.Chunks), logger)
				if err != nil {
					status.Status = "failed"
					status.ErrorMessage = fmt.Sprintf("failed to wait for transaction confirmation for chunk %d: %v", i+1, err)
					status.UpdatedAt = time.Now().Format(time.RFC3339)
					_ = writeLocalBulkImportStatus(status)
					return fmt.Errorf("failed to wait for transaction confirmation for chunk %d: %w", i+1, err)
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
