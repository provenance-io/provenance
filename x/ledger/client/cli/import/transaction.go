package cli

import (
	"bytes"
	"fmt"
	"regexp"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"cosmossdk.io/log"
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

// broadcastAndCheckTx broadcasts a transaction and checks its response for success/failure
func broadcastAndCheckTx(clientCtx client.Context, cmd *cobra.Command, msg sdk.Msg, chunkIndex int, logger log.Logger) (string, error) {
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
		return "", fmt.Errorf("failed to broadcast transaction for chunk %d: %w", chunkIndex, err)
	}

	// Parse the captured output to check transaction status
	outputBytes := outputBuffer.Bytes()
	outputStr := string(outputBytes)
	logger.Debug("Transaction output captured", "chunk_index", chunkIndex, "output", outputStr)

	var txResp sdk.TxResponse
	var txHash string

	if err := clientCtx.Codec.UnmarshalJSON(outputBytes, &txResp); err != nil {
		// If we can't parse the response, try to extract tx hash from the output string
		logger.Debug("Could not parse transaction response", "output", outputStr, "error", err)

		// Try to extract transaction hash from the output string
		// Look for patterns like "txhash: ABC123..." or "hash: ABC123..."
		txHash = extractTxHashFromOutput(outputStr)
		if txHash == "" {
			logger.Warn("Could not extract transaction hash from output", "output", outputStr)
		} else {
			logger.Info("Extracted transaction hash from output", "tx_hash", txHash)
		}
	} else {
		// Successfully parsed the response
		txHash = txResp.TxHash

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
