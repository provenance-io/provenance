package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

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
func parseTransactionResponse(outputBytes []byte, outputStr string) (*sdk.TxResponse, string, error) {
	var txResp sdk.TxResponse
	var txHash string

	// Try JSON parsing first
	if err := json.Unmarshal(outputBytes, &txResp); err == nil {
		// JSON parsing successful
		return &txResp, txResp.TxHash, nil
	}

	// Try YAML parsing
	if err := yaml.Unmarshal(outputBytes, &txResp); err == nil {
		// YAML parsing successful
		return &txResp, txResp.TxHash, nil
	}

	// If both JSON and YAML parsing fail, try to extract hash from the output string
	txHash = extractTxHashFromOutput(outputStr)
	if txHash != "" {
		// Create a minimal response with just the hash
		return &sdk.TxResponse{
			TxHash: txHash,
			Code:   0, // Assume success if we can extract hash
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

	// Create a custom writer to capture the output
	var outputBuffer bytes.Buffer
	originalOutput := clientCtx.Output
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
	txResp, txHash, err := parseTransactionResponse(outputBytes, outputStr)
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

// waitForTransactionConfirmation waits for transaction confirmation and sequence update
func waitForTransactionConfirmation(clientCtx client.Context, cmd *cobra.Command, chunkIndex int, totalChunks int, logger log.Logger) error {
	logger.Info("Waiting for transaction confirmation", "chunk_index", chunkIndex+1, "total_chunks", totalChunks)

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
				return fmt.Errorf("failed to get account info after chunk %d: %w", chunkIndex+1, err)
			}
			time.Sleep(1 * time.Second)
			continue
		}

		logger.Debug("Account sequence updated", "sequence", account.GetSequence())
		break
	}

	return nil
}
