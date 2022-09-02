package antewrapper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	txgas "github.com/provenance-io/provenance/x/txgas/types"
)

// TxGasLimitDecorator will check if the transaction's gas amount is higher than
// 5% of the maximum gas allowed per block.
// If gas is too high, decorator returns error and tx is rejected from mempool.
// If gas is below the limit, then call next AnteHandler
// CONTRACT: Tx must implement FeeTx to use TxGasLimitDecorator
type TxGasLimitDecorator struct{
	tk txgas.TxKeeper
}

func NewTxGasLimitDecorator(tk txgas.TxKeeper) TxGasLimitDecorator {
	return TxGasLimitDecorator{
		tk: tk,
	}
}

// Checks whether the given message is related to governance.
func isGovernanceMessage(msg sdk.Msg) bool {
	_, isSubmitPropMsg := msg.(*govtypes.MsgSubmitProposal)
	_, isVoteMsg := msg.(*govtypes.MsgVote)
	_, isVoteWeightedMsg := msg.(*govtypes.MsgVoteWeighted)
	_, isDepositMsg := msg.(*govtypes.MsgDeposit)
	return isSubmitPropMsg || isVoteMsg || isVoteWeightedMsg || isDepositMsg
}

func (mfd TxGasLimitDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}
	// Ensure that the requested gas does not exceed the configured block maximum
	gas := feeTx.GetGas()
	gasTxLimit := mfd.tk.GetParams(ctx).TxGasLimit

	// Skip gas limit check for txs with MsgSubmitProposal
	hasGovMsg := false
	for _, msg := range tx.GetMsgs() {
		isGovMsg := isGovernanceMessage(msg)
		if isGovMsg {
			hasGovMsg = true
			break
		}
	}

	// TODO - remove "gasTxLimit > 0" with SDK 0.46 which fixes the infinite gas meter to use max int vs zero for the limit.
	if !isTestContext(ctx) && gasTxLimit > 0 && gas > gasTxLimit && !hasGovMsg {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrTxTooLarge, "transaction gas exceeds maximum allowed; got: %d max allowed: %d", gas, gasTxLimit)
	}
	return next(ctx, tx, simulate)
}
