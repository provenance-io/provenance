package cli

import (
	"bytes"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"cosmossdk.io/log"
)

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
