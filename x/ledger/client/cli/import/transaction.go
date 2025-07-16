package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"encoding/hex"

	"cosmossdk.io/log"
	"sigs.k8s.io/yaml"
)

// extractTxHashFromOutput attempts to extract a transaction hash from the output string
func extractTxHashFromOutput(output string) string {
	// Common patterns for transaction hashes in output
	patterns := []string{
		`"txhash":\s*"([a-fA-F0-9]{64})"`,       // JSON format: "txhash": "ABC123..."
		`"hash":\s*"([a-fA-F0-9]{64})"`,         // JSON format: "hash": "ABC123..."
		`txhash:\s*([a-fA-F0-9]{64})`,           // Plain text: txhash: ABC123...
		`hash:\s*([a-fA-F0-9]{64})`,             // Plain text: hash: ABC123...
		`transaction hash:\s*([a-fA-F0-9]{64})`, // Plain text: transaction hash: ABC123...
		`([a-fA-F0-9]{64})`,                     // Any 64-character hex string
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(output)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

// parseTransactionResponse attempts to parse transaction response in multiple formats
func parseTransactionResponse(outputBytes []byte, outputStr string, logger log.Logger) (*sdk.TxResponse, string, error) {
	var txResp sdk.TxResponse
	var txHash string

	// Try JSON parsing first
	if err := json.Unmarshal(outputBytes, &txResp); err == nil {
		// JSON parsing successful
		logger.Debug("JSON parsing successful", "tx_hash", txResp.TxHash, "tx_code", txResp.Code)
		return &txResp, txResp.TxHash, nil
	}

	// Try YAML parsing
	if err := yaml.Unmarshal(outputBytes, &txResp); err == nil {
		// YAML parsing successful
		logger.Debug("YAML parsing successful", "tx_hash", txResp.TxHash, "tx_code", txResp.Code)
		return &txResp, txResp.TxHash, nil
	}

	// If both JSON and YAML parsing fail, try to extract hash from the output string
	txHash = extractTxHashFromOutput(outputStr)
	if txHash != "" {
		// Try to extract additional information from the output string
		// Look for code and raw_log in the YAML-like output
		code := uint32(0)
		rawLog := ""

		// Extract code if present
		if codeMatch := regexp.MustCompile(`code:\s*(\d+)`).FindStringSubmatch(outputStr); len(codeMatch) > 1 {
			if codeVal, err := strconv.ParseUint(codeMatch[1], 10, 32); err == nil {
				code = uint32(codeVal)
			}
		}

		// Extract raw_log if present
		if logMatch := regexp.MustCompile(`raw_log:\s*['"]([^'"]*)['"]`).FindStringSubmatch(outputStr); len(logMatch) > 1 {
			rawLog = logMatch[1]
		}

		logger.Debug("Extracted hash and info from output", "tx_hash", txHash, "tx_code", code, "raw_log", rawLog)
		return &sdk.TxResponse{
			TxHash: txHash,
			Code:   code,
			RawLog: rawLog,
		}, txHash, nil
	}

	// All parsing methods failed
	return nil, "", fmt.Errorf("could not parse transaction response in JSON, YAML, or extract hash from output")
}

// broadcastAndCheckTx broadcasts a transaction and checks its response for success/failure
func broadcastAndCheckTx(clientCtx client.Context, cmd *cobra.Command, msg sdk.Msg, chunkIndex int, logger log.Logger, flatFeeInfo *FlatFeeInfo) (string, error) {
	// Force broadcast mode to sync for this transaction to ensure it's committed
	originalBroadcastMode := cmd.Flag(flags.FlagBroadcastMode).Value.String()
	cmd.Flag(flags.FlagBroadcastMode).Value.Set("sync")

	// Force JSON output format for better parsing
	originalOutput := clientCtx.Output
	clientCtx.Output = &bytes.Buffer{}

	// Create a custom writer to capture the output
	var outputBuffer bytes.Buffer
	clientCtx.Output = &outputBuffer

	// Set flat fee if available
	if flatFeeInfo != nil && !flatFeeInfo.FeeAmount.IsZero() {
		// Clear gas-prices to avoid conflict with fees
		if cmd.Flag("gas-prices") != nil {
			cmd.Flag("gas-prices").Value.Set("")
		}
		// Set the fee amount in the command flags
		cmd.Flag(flags.FlagFees).Value.Set(flatFeeInfo.FeeAmount.String())
		logger.Debug("Set flat fee for transaction", "fee", flatFeeInfo.FeeAmount.String())
	}

	// Force JSON output format
	cmd.Flag("output").Value.Set("json")

	// Use GenerateOrBroadcastTxCLI which provides detailed response information
	err := tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
	if err != nil {
		// Restore original broadcast mode and output
		cmd.Flag(flags.FlagBroadcastMode).Value.Set(originalBroadcastMode)
		clientCtx.Output = originalOutput
		return "", fmt.Errorf("failed to broadcast transaction for chunk %d: %w", chunkIndex, err)
	}

	// Restore original broadcast mode and output
	cmd.Flag(flags.FlagBroadcastMode).Value.Set(originalBroadcastMode)
	clientCtx.Output = originalOutput

	// Parse the captured output to check transaction status
	outputBytes := outputBuffer.Bytes()
	outputStr := string(outputBytes)
	logger.Debug("Transaction output captured", "chunk_index", chunkIndex, "output", outputStr)

	// Parse transaction response in multiple formats
	txResp, txHash, err := parseTransactionResponse(outputBytes, outputStr, logger)
	if err != nil {
		logger.Warn("Could not parse transaction response, but transaction may have succeeded", "error", err)
		// Try to extract hash as last resort
		txHash = extractTxHashFromOutput(outputStr)
		if txHash == "" {
			return "", fmt.Errorf("failed to parse transaction response or extract hash for chunk %d: %w", chunkIndex, err)
		}
		logger.Info("Extracted transaction hash from output", "tx_hash", txHash)
	} else {
		// Successfully parsed the response
		// Check if the transaction was successful
		if txResp.Code != 0 {
			logger.Error("Transaction failed",
				"chunk_index", chunkIndex,
				"code", txResp.Code,
				"raw_log", txResp.RawLog,
				"tx_hash", txHash)
			return txHash, fmt.Errorf("transaction failed for chunk %d with code %d: %s", chunkIndex, txResp.Code, txResp.RawLog)
		}
		logger.Info("Transaction successful",
			"chunk_index", chunkIndex,
			"tx_hash", txHash,
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
		return txHash, fmt.Errorf("failed to get account info after chunk %d: %w", chunkIndex, err)
	}

	// Log the sequence number for debugging
	logger.Debug("Account sequence after transaction", "sequence", account.GetSequence())

	return txHash, nil
}

// waitForTransactionConfirmation waits for transaction confirmation and validates success
func waitForTransactionConfirmation(clientCtx client.Context, cmd *cobra.Command, chunkIndex int, totalChunks int, logger log.Logger, txHash string) error {
	logger.Info("Waiting for transaction confirmation", "chunk_index", chunkIndex+1, "total_chunks", totalChunks, "tx_hash", txHash)

	if txHash == "" {
		return fmt.Errorf("cannot validate transaction confirmation: no transaction hash provided")
	}

	// Use the unified validation function with retry and block wait options
	success, err := validateTransactionByHash(clientCtx, txHash, logger,
		WithRetries(3, 2*time.Second),
		WithBlockWait(true),
		WithChunkContext(chunkIndex, totalChunks),
	)

	if err != nil {
		return fmt.Errorf("transaction confirmation failed for chunk %d: %w", chunkIndex+1, err)
	}

	if !success {
		return fmt.Errorf("transaction for chunk %d was not successful", chunkIndex+1)
	}

	return nil
}

// validateTransactionByHash validates a transaction by hash with optional retry logic
// This function unifies the transaction validation logic used in both checkTransactionAlreadyProcessed
// and waitForTransactionConfirmation
func validateTransactionByHash(clientCtx client.Context, txHash string, logger log.Logger, options ...ValidationOption) (bool, error) {
	if txHash == "" {
		return false, fmt.Errorf("transaction hash is empty")
	}

	// Parse options
	opts := &ValidationOptions{
		MaxRetries:   0, // No retries by default (for immediate checks)
		RetryDelay:   2 * time.Second,
		WaitForBlock: false,
		ChunkIndex:   -1,
		TotalChunks:  0,
	}

	for _, opt := range options {
		opt(opts)
	}

	// Wait for block if requested (for confirmation scenarios)
	if opts.WaitForBlock {
		timeoutCommit := clientCtx.Viper.GetDuration("consensus.timeout_commit")
		if timeoutCommit == 0 {
			timeoutCommit = 3 * time.Second
		}
		waitDuration := timeoutCommit * 2
		logger.Debug("Waiting for next block", "wait_duration", waitDuration, "timeout_commit", timeoutCommit)
		time.Sleep(waitDuration)
	}

	// Convert hash to lowercase and decode from hex string to bytes
	hashLower := strings.ToLower(txHash)
	hashBytes, err := hex.DecodeString(hashLower)
	if err != nil {
		return false, fmt.Errorf("failed to decode transaction hash %s: %w", txHash, err)
	}

	var lastErr error
	maxRetries := opts.MaxRetries
	if maxRetries == 0 {
		maxRetries = 1 // At least try once
	}

	for retry := 0; retry < maxRetries; retry++ {
		// Query the transaction by hash
		txResp, err := clientCtx.Client.Tx(context.Background(), hashBytes, false)
		if err != nil {
			lastErr = err
			if strings.Contains(err.Error(), "not found") {
				if retry < maxRetries-1 {
					logger.Debug("Transaction not yet found in blockchain, retrying",
						"retry", retry+1, "max_retries", maxRetries, "tx_hash", txHash)
					time.Sleep(opts.RetryDelay)
					continue
				}
				// Final retry failed - transaction not found
				return false, fmt.Errorf("transaction not found in blockchain: %w", err)
			}
			// For other errors, wait a bit longer on retries
			if retry < maxRetries-1 {
				logger.Debug("Error querying transaction, retrying",
					"retry", retry+1, "max_retries", maxRetries, "error", err, "tx_hash", txHash)
				time.Sleep(opts.RetryDelay * 2)
				continue
			}
			return false, fmt.Errorf("failed to query transaction %s: %w", txHash, err)
		}

		// Transaction found - check if it was successful
		if txResp.TxResult.Code == 0 {
			// Success case - log appropriate message based on context
			if opts.ChunkIndex >= 0 {
				logger.Info("Transaction confirmed and successful",
					"chunk_index", opts.ChunkIndex+1,
					"total_chunks", opts.TotalChunks,
					"tx_hash", txHash,
					"gas_used", txResp.TxResult.GasUsed,
					"gas_wanted", txResp.TxResult.GasWanted)
			} else {
				logger.Info("Transaction already processed successfully", "tx_hash", txHash)
			}
			return true, nil
		} else {
			// Transaction was committed but failed
			if opts.ChunkIndex >= 0 {
				logger.Error("Transaction confirmed but failed",
					"chunk_index", opts.ChunkIndex+1,
					"total_chunks", opts.TotalChunks,
					"tx_hash", txHash,
					"code", txResp.TxResult.Code,
					"gas_used", txResp.TxResult.GasUsed,
					"gas_wanted", txResp.TxResult.GasWanted,
					"log", txResp.TxResult.Log)
				return false, fmt.Errorf("transaction for chunk %d failed with code %d: %s",
					opts.ChunkIndex+1, txResp.TxResult.Code, txResp.TxResult.Log)
			} else {
				logger.Info("Transaction found but failed", "tx_hash", txHash, "code", txResp.TxResult.Code)
				return false, nil
			}
		}
	}

	// This should never be reached, but just in case
	return false, fmt.Errorf("transaction validation failed after %d retries: %w", maxRetries, lastErr)
}

// ValidationOptions configures the behavior of validateTransactionByHash
type ValidationOptions struct {
	MaxRetries   int           // Number of retries (0 = no retries, immediate check)
	RetryDelay   time.Duration // Delay between retries
	WaitForBlock bool          // Whether to wait for next block before checking
	ChunkIndex   int           // Chunk index for logging (-1 = not in chunk context)
	TotalChunks  int           // Total chunks for logging
}

// ValidationOption is a function that configures ValidationOptions
type ValidationOption func(*ValidationOptions)

// WithRetries configures retry behavior
func WithRetries(maxRetries int, retryDelay time.Duration) ValidationOption {
	return func(opts *ValidationOptions) {
		opts.MaxRetries = maxRetries
		opts.RetryDelay = retryDelay
	}
}

// WithBlockWait configures whether to wait for next block
func WithBlockWait(wait bool) ValidationOption {
	return func(opts *ValidationOptions) {
		opts.WaitForBlock = wait
	}
}

// WithChunkContext configures chunk context for logging
func WithChunkContext(chunkIndex, totalChunks int) ValidationOption {
	return func(opts *ValidationOptions) {
		opts.ChunkIndex = chunkIndex
		opts.TotalChunks = totalChunks
	}
}
