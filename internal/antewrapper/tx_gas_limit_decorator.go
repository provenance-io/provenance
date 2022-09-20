package antewrapper

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const gasTxLimit uint64 = 4_000_000

// TxGasLimitDecorator will check if the transaction's gas amount is higher than
// 5% of the maximum gas allowed per block.
// If gas is too high, decorator returns error and tx is rejected from mempool.
// If gas is below the limit, then call next AnteHandler
// CONTRACT: Tx must implement FeeTx to use TxGasLimitDecorator
type TxGasLimitDecorator struct{}

func NewTxGasLimitDecorator() TxGasLimitDecorator {
	return TxGasLimitDecorator{}
}

// isGovMessage returns true if the provided message is a governance module message.
func isGovMessage(msg sdk.Msg) bool {
	return strings.HasPrefix(sdk.MsgTypeURL(msg), "/cosmos.gov.")
}

func isOnlyGovMsgs(msgs []sdk.Msg) bool {
	// If there are no messages, there are no gov messages, so return false.
	if len(msgs) == 0 {
		return false
	}
	for _, msg := range msgs {
		if !isGovMessage(msg) {
			return false
		}
	}
	return true
}

func (mfd TxGasLimitDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	feeTx, err := GetFeeTx(tx)
	if err != nil {
		return ctx, err
	}
	// Skip gas limit check for test contexts.
	// Skip gas limit check for txs with only gov msgs.
	if !isTestContext(ctx) && !isOnlyGovMsgs(tx.GetMsgs()) {
		// Ensure that the requested gas does not exceed the configured block maximum
		// If consensus_params.block.max_gas is set to -1, ignore gasTxLimit. This is to allow for testing on local nodes
		// since mainnet and testnet have block level limit set.
		maxGasLimit := ctx.ConsensusParams().Block.GetMaxGas()
		gas := feeTx.GetGas()
		if maxGasLimit > -1 && gas > gasTxLimit {
			return ctx, sdkerrors.ErrTxTooLarge.Wrapf("transaction gas exceeds maximum allowed; got: %d max allowed: %d", gas, gasTxLimit)
		}
	}

	return next(ctx, tx, simulate)
}
