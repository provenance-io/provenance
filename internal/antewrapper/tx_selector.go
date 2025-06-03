package antewrapper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// provTxSelector implements the baseapp.TxSelector interface, and is our version of the SDK's defaultTxSelector.
// The only difference between ours and theirs is that ours uses our GetGasWanted function instead of tx.GetGas().
//
// Since our Simulate returns the fee amount as the gas, the tx.GetGas() method will often return an amount of
// gas that exceeds the block limit. Defining this custom TxSelector fixes the ABCI PrepareProposal process.
// Without this, Txs (that were simulated with gas-prices 1nhash) get put into the mempool, but never get selected
// because the defaultTxSelector ends up thinking the tx doesn't fit in a block. The behavior the user sees is
// that the Tx is submitted, never to be heard from again. The node that it was sent to, though, will have it in
// mempool where it'll get re-checked until it either fails re-check or the node is restarted.
//
// At the time of writing this, our default ABCI ProcessProposal function is a no-op; however, if the mempool
// is NOT a no-op (ours is a no-op), the DefaultProposalHandler.ProcessProposalHandler method also uses tx.GetGas()
// to track the block's gas. If Tx processing and block creation stops working (e.g. after an SDK bump), check that.
type provTxSelector struct {
	totalTxBytes uint64
	totalTxGas   uint64
	selectedTxs  [][]byte
}

// NewProvTxSelector creates the new baseapp.TxSelector that we use.
func NewProvTxSelector() baseapp.TxSelector {
	return &provTxSelector{}
}

// SelectedTxs should return a copy of the selected transactions.
func (ts *provTxSelector) SelectedTxs(_ context.Context) [][]byte {
	txs := make([][]byte, len(ts.selectedTxs))
	copy(txs, ts.selectedTxs)
	return txs
}

// Clear should clear the TxSelector, nulling out all relevant fields.
func (ts *provTxSelector) Clear() {
	ts.totalTxBytes = 0
	ts.totalTxGas = 0
	ts.selectedTxs = nil
}

// SelectTxForProposal should attempt to select a transaction for inclusion in
// a proposal based on inclusion criteria defined by the TxSelector. It must
// return <true> if the caller should halt the transaction selection loop
// (typically over a mempool) or <false> otherwise.
func (ts *provTxSelector) SelectTxForProposal(ctx context.Context, maxTxBytes, maxBlockGas uint64, memTx sdk.Tx, txBz []byte) bool {
	txSize := uint64(len(txBz))

	var txGasLimit uint64
	if memTx != nil {
		// This little chunk is the only thing changed from the defaultTxSelector.
		if feeTx, err := GetFeeTx(memTx); err != nil {
			txGasLimit, _ = GetGasWanted(sdk.UnwrapSDKContext(ctx).Logger(), feeTx)
		}
	}

	// only add the transaction to the proposal if we have enough capacity
	if (txSize + ts.totalTxBytes) <= maxTxBytes {
		// If there is a max block gas limit, add the tx only if the limit has
		// not been met.
		if maxBlockGas > 0 {
			if (txGasLimit + ts.totalTxGas) <= maxBlockGas {
				ts.totalTxGas += txGasLimit
				ts.totalTxBytes += txSize
				ts.selectedTxs = append(ts.selectedTxs, txBz)
			}
		} else {
			ts.totalTxBytes += txSize
			ts.selectedTxs = append(ts.selectedTxs, txBz)
		}
	}

	// check if we've reached capacity; if so, we cannot select any more transactions
	return ts.totalTxBytes >= maxTxBytes || (maxBlockGas > 0 && (ts.totalTxGas >= maxBlockGas))
}
