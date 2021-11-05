package handlers

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmosante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	msgbasedfeetypes "github.com/provenance-io/provenance/x/msgfees/types"
)

type MsgBasedFeeInvoker struct {
	msgBasedFeeKeeper msgbasedfeetypes.MsgBasedFeeKeeper
	bankKeeper        banktypes.Keeper
	accountKeeper     cosmosante.AccountKeeper
	feegrantKeeper    msgbasedfeetypes.FeegrantKeeper
}

// concrete impl of how to charge Msg Based Fees
func NewMsgBasedFeeInvoker(bankKeeper banktypes.Keeper, accountKeeper cosmosante.AccountKeeper, feegrantKeeper msgbasedfeetypes.FeegrantKeeper, msgBasedFeeKeeper msgbasedfeetypes.MsgBasedFeeKeeper) MsgBasedFeeInvoker {
	return MsgBasedFeeInvoker{
		msgBasedFeeKeeper,
		bankKeeper,
		accountKeeper,
		feegrantKeeper,
	}
}

func (afd MsgBasedFeeInvoker) Invoke(ctx sdk.Context, simulate bool) (coins sdk.Coins, err error) {
	if ctx.TxBytes() != nil && len(ctx.TxBytes()) != 0 {
		ctx.Logger().Debug("NOTICE: In chargeFees()")
		originalGasMeter := ctx.GasMeter()
		// eat up the gas cost for charging fees.
		ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

		var msgs []sdk.Msg
		tx, err := app.txDecoder(ctx.TxBytes())
		if err != nil {
			panic(fmt.Errorf("error in chargeFees() while getting txBytes: %w", err))
		}
		msgs = tx.GetMsgs()
		// cast to FeeTx
		feeTx, ok := tx.(sdk.FeeTx)
		// only charge additional fee if of type FeeTx since it should give fee payer.
		// for provenace should be a FeeTx since antehandler should enforce it, but
		// not adding complexity here
		if ok {
			feePayer := feeTx.FeePayer()
			deductFeesFrom := feePayer

			// TODO if feegranter set deduct fee from feegranter account.
			// this works with only when feegrant enabled.

			deductFeesFromAcc := afd.accountKeeper.GetAccount(ctx, deductFeesFrom)
			if deductFeesFromAcc == nil {
				panic("fee payer address: %s does not exist")
			}

			for _, msg := range msgs {
				ctx.Logger().Info(fmt.Sprintf("The message type in defer block for fee charging : %s", sdk.MsgTypeURL(msg)))
				msgFees, err := afd.msgBasedFeeKeeper.GetMsgBasedFee(ctx, sdk.MsgTypeURL(msg))
				if err != nil {
					// do nothing for now
					ctx.Logger().Error("unable to get message fees", "err", err)
				}
				if msgFees != nil {
					ctx.Logger().Info("Retrieved a msg based fee.")
					afd.msgBasedFeeKeeper.DeductFees(afd.bankKeeper, ctx, deductFeesFromAcc, sdk.Coins{msgFees.AdditionalFee})
				}
				// TODO remove this but just for testing $$$$$$$$$$
				afd.msgBasedFeeKeeper.DeductFees(afd.bankKeeper, ctx, deductFeesFromAcc, sdk.Coins{sdk.NewInt64Coin("nhash", 55555)})
			}
			//set back the original gasMeter
			ctx = ctx.WithGasMeter(originalGasMeter)
		}
	}
}
