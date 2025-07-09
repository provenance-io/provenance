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

	"github.com/provenance-io/provenance/x/ledger/types"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
)

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

// estimateGasCosts runs representative simulations to understand gas costs
func estimateGasCosts(chunks []*types.GenesisState, clientCtx client.Context, cmd *cobra.Command, logger log.Logger) (*GasCosts, error) {
	// Find a representative ledger to use for testing
	var testLedger *types.LedgerToEntries
	for _, chunk := range chunks {
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
			logger.Warn("Failed to simulate with more entries, using single entry cost", "error", err)
		} else {
			// Calculate per-entry cost from the larger sample
			entryCost = (ledgerWithMoreEntriesGas - ledgerWithKeyGas) / numTestEntries
		}
	}

	logger.Info("Gas cost calculation",
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
			// Subsequent chunks: only entries (ledger already exists)
			if len(lte.Entries) > 0 {
				totalGas += len(lte.Entries) * costs.EntryGas
			}
		}
	}
	return totalGas
}
