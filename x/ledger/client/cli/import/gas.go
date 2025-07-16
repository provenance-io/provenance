package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"cosmossdk.io/log"

	flatfeestypes "github.com/provenance-io/provenance/x/flatfees/types"
	"github.com/provenance-io/provenance/x/ledger/types"
)

// getChunkSizeBytes returns the actual serialized size of a chunk in bytes
func getChunkSizeBytes(chunk *types.GenesisState) int {
	data, err := json.Marshal(chunk)
	if err != nil {
		return 0
	}
	return len(data)
}

// simulateChunkGas builds, signs, and simulates the transaction for gas estimation (for validation only)
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

	// Set a dummy fee (will be overwritten by simulation)
	txBuilder.SetFeeAmount([]sdk.Coin{sdk.NewInt64Coin("nhash", 1)})

	// Sign the transaction
	err = tx.Sign(cmd.Context(), txFactory, clientCtx.GetFromName(), txBuilder, false)
	if err != nil {
		return 0, fmt.Errorf("failed to sign tx for simulation: %w", err)
	}

	txBytes, err := clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return 0, fmt.Errorf("failed to encode tx: %w", err)
	}

	// Use flat fees CalculateTxFees query for simulation
	queryClient := flatfeestypes.NewQueryClient(clientCtx)
	response, err := queryClient.CalculateTxFees(
		context.Background(),
		&flatfeestypes.QueryCalculateTxFeesRequest{
			TxBytes:       txBytes,
			GasAdjustment: 1.3, // 30% margin for gas estimation to balance accuracy and efficiency
		},
	)
	if err != nil {
		return 0, fmt.Errorf("failed to simulate tx: %w", err)
	}

	return int(response.EstimatedGas), nil
}

// estimateChunkGasFromCosts estimates gas usage for a chunk using the cost model (for validation only)
func estimateChunkGasFromCosts(chunk *types.GenesisState, costs *GasCosts) int {
	totalGas := 0
	for _, lte := range chunk.LedgerToEntries {
		if lte.LedgerKey != nil && lte.Ledger != nil {
			// First chunk: ledger + key + entries
			totalGas += costs.LedgerWithKeyGas
			if len(lte.Entries) > 0 {
				totalGas += len(lte.Entries) * costs.EntryGas
			}
		} else if lte.LedgerKey != nil && lte.Ledger == nil {
			// Subsequent chunks: only key + entries (ledger already exists)
			if len(lte.Entries) > 0 {
				totalGas += len(lte.Entries) * costs.EntryGas
			}
		}
	}
	return totalGas
}

// EstimateGasCostsAccurately simulates two separate transactions to accurately calculate gas costs:
// 1. X = A transaction with only ledger + key (no entries)
// 2. Y = A transaction with ledger + key + entries
// Then calculates: LedgerWithKeyGas = X, EntryGas = (Y - X) / num_entries
// This is a more accurate way to calculate gas costs because it takes into account the fact that
// the ledger + key is already in the database and only the entries need to be added.
func EstimateGasCostsAccurately(chunk *types.GenesisState, clientCtx client.Context, cmd *cobra.Command, logger log.Logger) (*GasCosts, error) {
	
	if len(chunk.LedgerToEntries) == 0 {
		return &GasCosts{
			LedgerWithKeyGas: 70000,
			EntryGas:         35000, // fallback
		}, nil
	}

	// Create a chunk with only ledger + key (no entries)
	ledgerOnlyChunk := &types.GenesisState{
		LedgerToEntries: make([]types.LedgerToEntries, len(chunk.LedgerToEntries)),
	}
	for i, lte := range chunk.LedgerToEntries {
		ledgerOnlyChunk.LedgerToEntries[i] = types.LedgerToEntries{
			LedgerKey: lte.LedgerKey,
			Ledger:    lte.Ledger,
			Entries:   []*types.LedgerEntry{}, // Empty entries
		}
	}

	// Simulate transaction with only ledger + key
	logger.Info("Simulating gas for ledger + key only")
	ledgerOnlyGas, err := simulateChunkGas(ledgerOnlyChunk, clientCtx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to simulate ledger-only gas: %w", err)
	}
	logger.Info("Simulated gas for ledger + key only", "gas", ledgerOnlyGas)

	// Simulate transaction with ledger + key + entries
	logger.Info("Simulating gas for ledger + key + entries")
	fullChunkGas, err := simulateChunkGas(chunk, clientCtx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to simulate full chunk gas: %w", err)
	}
	logger.Info("Simulated gas for ledger + key + entries", "gas", fullChunkGas)

	// Calculate total entries in the chunk
	totalEntries := 0
	for _, lte := range chunk.LedgerToEntries {
		totalEntries += len(lte.Entries)
	}

	if totalEntries == 0 {
		// No entries, use ledger-only gas for both
		return &GasCosts{
			LedgerWithKeyGas: ledgerOnlyGas,
			EntryGas:         35000, // fallback
		}, nil
	}

	// Calculate per-entry gas cost
	entryGas := (fullChunkGas - ledgerOnlyGas) / totalEntries

	logger.Info("Accurate gas cost calculation",
		"ledger_only_gas", ledgerOnlyGas,
		"full_chunk_gas", fullChunkGas,
		"total_entries", totalEntries,
		"calculated_entry_gas", entryGas,
		"calculated_ledger_with_key_gas", ledgerOnlyGas)

	return &GasCosts{
		LedgerWithKeyGas: ledgerOnlyGas,
		EntryGas:         entryGas,
	}, nil
}

// validateGasCosts tests the estimated gas costs against actual chunks to ensure accuracy
func validateGasCosts(gasCosts *GasCosts, chunks []*types.GenesisState, clientCtx client.Context, cmd *cobra.Command, logger log.Logger) error {
	logger.Info("Validating gas cost estimates against actual chunks")

	// Test against a few representative chunks
	maxValidationChunks := 3
	validationErrors := 0

	for i, chunk := range chunks {
		if i >= maxValidationChunks {
			break
		}

		// Skip empty chunks
		if len(chunk.LedgerToEntries) == 0 {
			continue
		}

		// Estimate gas using our cost model
		estimatedGas := estimateChunkGasFromCosts(chunk, gasCosts)

		// Simulate actual gas usage
		actualGas, err := simulateChunkGas(chunk, clientCtx, cmd)
		if err != nil {
			logger.Warn("Failed to validate chunk", "chunk_index", i+1, "error", err)
			continue
		}

		// Calculate accuracy
		accuracy := float64(estimatedGas) / float64(actualGas)
		errorPercent := (float64(estimatedGas-actualGas) / float64(actualGas)) * 100

		logger.Info("Gas cost validation result",
			"chunk_index", i+1,
			"estimated_gas", estimatedGas,
			"actual_gas", actualGas,
			"accuracy", accuracy,
			"error_percent", errorPercent,
			"ledger_count", len(chunk.LedgerToEntries),
			"total_entries", countTotalEntries(chunk))

		// Check if estimate is significantly off (more than 20% error)
		if accuracy < 0.8 || accuracy > 1.2 {
			validationErrors++
			logger.Warn("Gas estimate significantly off",
				"chunk_index", i+1,
				"estimated_gas", estimatedGas,
				"actual_gas", actualGas,
				"error_percent", errorPercent)
		}
	}

	if validationErrors > 0 {
		logger.Warn("Gas cost validation found significant errors",
			"validation_errors", validationErrors,
			"chunks_tested", maxValidationChunks)
		return fmt.Errorf("gas cost estimates have significant errors in %d/%d test chunks", validationErrors, maxValidationChunks)
	}

	logger.Info("Gas cost validation passed", "chunks_tested", maxValidationChunks)
	return nil
}

// countTotalEntries counts the total number of entries in a chunk
func countTotalEntries(chunk *types.GenesisState) int {
	total := 0
	for _, lte := range chunk.LedgerToEntries {
		total += len(lte.Entries)
	}
	return total
}
