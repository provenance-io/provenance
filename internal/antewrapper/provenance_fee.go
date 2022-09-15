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

// ProvenanceDeductFeeDecorator identifies the payer (using feegrant funds if appropriate),
// makes sure the payer has enough funds to cover the fees, and deducts the base fee from
// the payer's account. The base fee is the floor gas price * gas.
// If the first signer does not have the funds to pay for the fees, return with InsufficientFunds error
// Call next AnteHandler if fees successfully deducted.
// CONTRACT: In order to use ProvenanceDeductFeeDecorator:
//  1. Tx must implement FeeTx interface.
//  2. GasMeter must be a FeeGasMeter.
type ProvenanceDeductFeeDecorator struct {
	ak             authante.AccountKeeper
	bankKeeper     bankkeeper.Keeper
	feegrantKeeper authante.FeegrantKeeper
	msgFeeKeeper   msgfeestypes.MsgFeesKeeper
}

var (
	AttributeKeyBaseFee       = "basefee"
	AttributeKeyAdditionalFee = "additionalfee"
)

func NewProvenanceDeductFeeDecorator(
	accountKeeper authante.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	feegrantKeeper msgfeestypes.FeegrantKeeper,
	msgfeesKeeper msgfeestypes.MsgFeesKeeper,
) ProvenanceDeductFeeDecorator {
	return ProvenanceDeductFeeDecorator{
		ak:             accountKeeper,
		bankKeeper:     bankKeeper,
		feegrantKeeper: feegrantKeeper,
		msgFeeKeeper:   msgfeesKeeper,
	}
}

func (dfd ProvenanceDeductFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	feeTx, err := GetFeeTx(tx)
	if err != nil {
		return ctx, err
	}

	if !simulate && ctx.BlockHeight() > 0 && feeTx.GetGas() == 0 {
		return ctx, sdkerrors.ErrInvalidGasLimit.Wrap("must provide positive gas")
	}

	if err = dfd.checkDeductBaseFee(ctx, feeTx, simulate); err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

// checkDeductBaseFee does several things:
//  1. Checks for a feegrant and uses the base fees on it if it exists.
//  2. Makes sure the payer has enough funds to cover the base fee + additional fees.
//  3. Deducts the base fee from the payer.
//  4. Emits Tx events: 1. with the full fee and payer, 2. with base fee.
func (dfd ProvenanceDeductFeeDecorator) checkDeductBaseFee(ctx sdk.Context, feeTx sdk.FeeTx, simulate bool) error {
	if addr := dfd.ak.GetModuleAddress(types.FeeCollectorName); addr == nil {
		return sdkerrors.ErrLogic.Wrapf("%s module account has not been set", types.FeeCollectorName)
	}

	feeGasMeter, err := GetFeeGasMeter(ctx)
	if err != nil {
		return err
	}

	// Calculate the base and required fees.
	// Note: The MsgFeesDecorator only checks stuff during IsCheckTx, so we need to do it here too.
	msgs := feeTx.GetMsgs()
	baseFeeToConsume := CalculateBaseFee(ctx, feeTx, dfd.msgFeeKeeper)
	feeDist, err := dfd.msgFeeKeeper.CalculateAdditionalFeesToBePaid(ctx, msgs...)
	if err != nil {
		return sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	requiredFunds := baseFeeToConsume.Add(feeDist.TotalAdditionalFees...)

	deductFeesFrom, err := GetFeePayerUsingFeeGrant(ctx, dfd.feegrantKeeper, feeTx, baseFeeToConsume, msgs)
	if err != nil {
		return err
	}

	deductFeesFromAcc := dfd.ak.GetAccount(ctx, deductFeesFrom)
	if deductFeesFromAcc == nil {
		return sdkerrors.ErrUnknownAddress.Wrapf("fee payer address: %s does not exist", deductFeesFrom)
	}

	// Make sure the payer has enough funds for the whole fee.
	fee := feeTx.GetFee()
	balancePerCoin := sdk.NewCoins()
	for _, fc := range fee {
		balancePerCoin = balancePerCoin.Add(dfd.bankKeeper.GetBalance(ctx, deductFeesFrom, fc.Denom))
	}
	_, hasNeg := balancePerCoin.SafeSub(requiredFunds...)
	if hasNeg && !simulate {
		return sdkerrors.ErrInsufficientFunds.Wrapf("account %s does not have enough balance to pay for %q, balance: %q", deductFeesFrom, requiredFunds, balancePerCoin)
	}

	// deduct minimum amount from fee, remainder will be swept on success
	if !baseFeeToConsume.IsZero() && !simulate {
		err := DeductFees(dfd.bankKeeper, ctx, deductFeesFromAcc, baseFeeToConsume)
		if err != nil {
			return err
		}
		feeGasMeter.ConsumeBaseFee(baseFeeToConsume)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(sdk.EventTypeTx,
			sdk.NewAttribute(sdk.AttributeKeyFee, feeTx.GetFee().String()),
			sdk.NewAttribute(sdk.AttributeKeyFeePayer, deductFeesFrom.String()),
		),
		sdk.NewEvent(sdk.EventTypeTx,
			sdk.NewAttribute(AttributeKeyBaseFee, baseFeeToConsume.String()),
			sdk.NewAttribute(sdk.AttributeKeyFeePayer, deductFeesFrom.String()),
		),
	})

	return nil
}

func GetFeePayerUsingFeeGrant(ctx sdk.Context, feegrantKeeper authante.FeegrantKeeper, feeTx sdk.FeeTx, fee sdk.Coins, msgs []sdk.Msg) (sdk.AccAddress, error) {
	feePayer := feeTx.FeePayer()
	feeGranter := feeTx.FeeGranter()
	deductFeesFrom := feePayer

	// if feegranter set deduct base fee from feegranter account.
	// this works with only when feegrant enabled.
	if feeGranter != nil && !feeGranter.Equals(feePayer) {
		if feegrantKeeper == nil {
			return nil, sdkerrors.ErrInvalidRequest.Wrap("fee grants are not enabled")
		}
		err := feegrantKeeper.UseGrantedFees(ctx, feeGranter, feePayer, fee, msgs)
		if err != nil {
			msgTypes := make([]string, len(msgs))
			for i, msg := range msgs {
				msgTypes[i] = sdk.MsgTypeURL(msg)
			}
			return nil, cerrs.Wrapf(err, "failed to use fee grant: granter: %s, grantee: %s, fee: %q, msgs: %q", feeGranter, feePayer, fee, msgTypes)
		}
		deductFeesFrom = feeGranter
	}

	return deductFeesFrom, nil
}

// CalculateBaseFee calculates the base fee.
// The base fee is floor gas price
func CalculateBaseFee(ctx sdk.Context, feeTx sdk.FeeTx, msgfeekeeper msgfeestypes.MsgFeesKeeper) sdk.Coins {
	if isTestContext(ctx) {
		return DetermineTestBaseFeeAmount(ctx, feeTx)
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
		return sdkerrors.ErrInsufficientFee.Wrapf("invalid fee amount: %s", fee)
	}

	err := bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), types.FeeCollectorName, fee)
	if err != nil {
		return sdkerrors.ErrInsufficientFunds.Wrap(err.Error())
	}
	return nil
}
