package antewrapper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// TxGasLimitDecorator will check if the transaction's gas amount is higher than
// 5% of the maximum gas allowed per block.
// If gas is too high, decorator returns error and tx is rejected from mempool.
// If gas is below the limit, then call next AnteHandler
// CONTRACT: Tx must implement FeeTx to use TxGasLimitDecorator
type TxGasLimitDecorator struct{}

func NewTxGasLimitDecorator() TxGasLimitDecorator {
	return TxGasLimitDecorator{}
}

// govMsgURLs the MsgURLs of all the governance module's messages.
// Use getGovMsgURLs() instead of using this variable directly.
var govMsgURLs []string

// getGovMsgURLs returns govVoteMsgUrls, but first sets it if it hasn't yet been set.
func getGovMsgURLs() []string {
	// Checking for nil here (as opposed to len == 0) because we only want to set it
	// if it hasn't been set yet.
	if govMsgURLs == nil {
		// sdk.MsgTypeURL sometimes uses reflection and/or proto registration.
		// So govMsgURLs is only set when it's finally needed in the hopes
		// that everything's wired up as needed by then.
		govMsgURLs = []string{
			sdk.MsgTypeURL(&govtypesv1.MsgSubmitProposal{}),
			sdk.MsgTypeURL(&govtypesv1.MsgVote{}),
			sdk.MsgTypeURL(&govtypesv1.MsgVoteWeighted{}),
			sdk.MsgTypeURL(&govtypesv1.MsgDeposit{}),
			sdk.MsgTypeURL(&govtypesv1beta1.MsgSubmitProposal{}),
			sdk.MsgTypeURL(&govtypesv1beta1.MsgVote{}),
			sdk.MsgTypeURL(&govtypesv1beta1.MsgVoteWeighted{}),
			sdk.MsgTypeURL(&govtypesv1beta1.MsgDeposit{}),
		}
	}
	return govMsgURLs
}

// Checks whether the given message is related to governance.
func isGovernanceMessage(msg sdk.Msg) bool {
	msgURL := sdk.MsgTypeURL(msg)
	for _, govMsgURL := range getGovMsgURLs() {
		if msgURL == govMsgURL {
			return true
		}
	}
	return false
}

func (mfd TxGasLimitDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}
	// Ensure that the requested gas does not exceed the configured block maximum
	gas := feeTx.GetGas()
	gasTxLimit := uint64(4_000_000)

	// If consensus_params.block.max_gas is set to -1, ignore gasTxLimit. This is to allow for testing on local nodes
	// since mainnet and testnet have block level limit set.
	maxGasLimit := ctx.ConsensusParams().Block.GetMaxGas()

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
	if !isTestContext(ctx) && maxGasLimit > -1 && gasTxLimit > 0 && gas > gasTxLimit && !hasGovMsg {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrTxTooLarge, "transaction gas exceeds maximum allowed; got: %d max allowed: %d", gas, gasTxLimit)
	}
	return next(ctx, tx, simulate)
}
