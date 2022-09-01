package antewrapper

import (
	cerrs "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
)

// ProvenanceDeductFeeDecorator deducts fees from the first signer of the tx
// If the first signer does not have the funds to pay for the fees, return with InsufficientFunds error
// Call next AnteHandler if fees successfully deducted
// CONTRACT: Tx must implement FeeTx interface to use ProvenanceDeductFeeDecorator
type ProvenanceDeductFeeDecorator struct {
	ak             authante.AccountKeeper
	bankKeeper     bankkeeper.Keeper
	feegrantKeeper authante.FeegrantKeeper
	msgFeeKeeper   msgfeestypes.MsgFeesKeeper
}

var (
	AttributeKeyBaseFee       = "basefee"
	AttributeKeyAdditionalFee = "additionalfee"
	AttributeKeyMinFeeCharged = "min_fee_charged"
)

func NewProvenanceDeductFeeDecorator(ak authante.AccountKeeper, bk bankkeeper.Keeper, fk msgfeestypes.FeegrantKeeper, mbfk msgfeestypes.MsgFeesKeeper) ProvenanceDeductFeeDecorator {
	return ProvenanceDeductFeeDecorator{
		ak:             ak,
		bankKeeper:     bk,
		feegrantKeeper: fk,
		msgFeeKeeper:   mbfk,
	}
}

func (dfd ProvenanceDeductFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, err := GetFeeTx(tx)
	if err != nil {
		return ctx, err
	}

	if !simulate && ctx.BlockHeight() > 0 && feeTx.GetGas() == 0 {
		return ctx, sdkerrors.ErrInvalidGasLimit.Wrap("must provide positive gas")
	}

	if err = dfd.checkDeductFee(ctx, tx, simulate); err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

func (dfd ProvenanceDeductFeeDecorator) checkDeductFee(ctx sdk.Context, tx sdk.Tx, simulate bool) error {
	feeTx, err := GetFeeTx(tx)
	if err != nil {
		return err
	}
	feeGasMeter, ok := ctx.GasMeter().(*FeeGasMeter)
	if !ok {
		return sdkerrors.ErrTxDecode.Wrap("GasMeter not a FeeGasMeter")
	}
	if addr := dfd.ak.GetModuleAddress(types.FeeCollectorName); addr == nil {
		return sdkerrors.ErrNotFound.Wrapf("%s module account has not been set", types.FeeCollectorName)
	}

	fee := feeTx.GetFee()
	feePayer := feeTx.FeePayer()
	feeGranter := feeTx.FeeGranter()
	deductFeesFrom := feePayer

	// if feegranter set deduct fee from feegranter account.
	// this works with only when feegrant enabled.
	if feeGranter != nil && !feeGranter.Equals(feePayer) {
		if dfd.feegrantKeeper == nil {
			return sdkerrors.ErrInvalidRequest.Wrap("fee grants are not enabled")
		}
		err = dfd.feegrantKeeper.UseGrantedFees(ctx, feeGranter, feePayer, fee, tx.GetMsgs())
		if err != nil {
			return cerrs.Wrapf(err, "%s not allowed to pay fees from %s", feeGranter, feePayer)
		}

		deductFeesFrom = feeGranter
	}

	deductFeesFromAcc := dfd.ak.GetAccount(ctx, deductFeesFrom)
	if deductFeesFromAcc == nil {
		return sdkerrors.ErrUnknownAddress.Wrapf("fee payer address: %s does not exist", deductFeesFrom)
	}

	// Make sure there are enough fees provided to cover any msg-based fees.
	feeDist, errFromCalculateAdditionalFeesToBePaid := CalculateAdditionalFeesToBePaid(ctx, dfd.msgFeeKeeper, feeTx.GetMsgs()...)
	if errFromCalculateAdditionalFeesToBePaid != nil {
		return sdkerrors.ErrInvalidRequest.Wrap(errFromCalculateAdditionalFeesToBePaid.Error())
	}
	if feeDist != nil && len(feeDist.TotalAdditionalFees) > 0 {
		_, hasNeg := fee.SafeSub(feeDist.TotalAdditionalFees...)
		if hasNeg && !simulate {
			return sdkerrors.ErrInsufficientFee.Wrapf("invalid fee amount: %s", fee)
		}
	}

	// deduct minimum fee amount, remainder will be swept on success
	baseGasFee := CalculateFloorGasFee(ctx, feeTx, dfd.msgFeeKeeper)
	if !baseGasFee.IsZero() && !simulate {
		err = DeductFees(dfd.bankKeeper, ctx, deductFeesFromAcc, baseGasFee)
		if err != nil {
			return err
		}
		feeGasMeter.ConsumeBaseFee(baseGasFee)
	}

	events := sdk.Events{
		sdk.NewEvent(sdk.EventTypeTx,
			sdk.NewAttribute(sdk.AttributeKeyFee, feeTx.GetFee().String()),
			sdk.NewAttribute(sdk.AttributeKeyFeePayer, deductFeesFrom.String()),
		),
		sdk.NewEvent(sdk.EventTypeTx,
			sdk.NewAttribute(AttributeKeyMinFeeCharged, baseGasFee.String()),
		),
	}
	ctx.EventManager().EmitEvents(events)

	return nil
}

// CalculateFloorGasFee gets the minimum gas fee for the provided FeeTx.
func CalculateFloorGasFee(ctx sdk.Context, feeTx sdk.FeeTx, msgfeekeeper msgfeestypes.MsgFeesKeeper) sdk.Coins {
	if isTestContext(ctx) {
		DetermineTestBaseFeeAmount(ctx, feeTx)
	}
	gasWanted := feeTx.GetGas()
	floorPrice := msgfeekeeper.GetFloorGasPrice(ctx)
	amount := floorPrice.Amount.Mul(sdk.NewIntFromUint64(gasWanted))
	baseFeeToDeduct := sdk.NewCoins(sdk.NewCoin(floorPrice.Denom, amount))
	return baseFeeToDeduct
}

// DetermineTestBaseFeeAmount determines the type of test that is running.  ChainID = "" is a simple unit
// We need this because of how tests are setup using atom and we have nhash specific code for msgfees
func DetermineTestBaseFeeAmount(ctx sdk.Context, feeTx sdk.FeeTx) sdk.Coins {
	if len(ctx.ChainID()) == 0 {
		return feeTx.GetFee()
	}
	return sdk.NewCoins()
}

// DeductFees deducts fees from the given account.
func DeductFees(bankKeeper bankkeeper.Keeper, ctx sdk.Context, acc types.AccountI, fee sdk.Coins) error {
	if !fee.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "invalid fee amount: %s", fee)
	}

	err := bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), types.FeeCollectorName, fee)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
	}
	return nil
}
