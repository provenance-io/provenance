package antewrapper

import (
	"bytes"
	"context"
	"fmt"
	"maps"
	"runtime/debug"
	"slices"
	"strings"

	cerrs "cosmossdk.io/errors"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/feegrant"

	cflags "github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/provenance-io/provenance/internal/pioconfig"
	internalsdk "github.com/provenance-io/provenance/internal/sdk"
)

const (
	// AttributeKeyBaseFee is the amount of fee charged up-front, even if the tx fails.
	AttributeKeyBaseFee = "basefee"
	// AttributeKeyAdditionalFee is the amount of fee required upon success.
	AttributeKeyAdditionalFee = "additionalfee"
	// AttributeKeyFeeOverage is the amount paid on top of the required amounts.
	AttributeKeyFeeOverage = "fee_overage"
	// AttributeKeyMinFeeCharged is
	AttributeKeyMinFeeCharged = "min_fee_charged"

	// nilStr is a string to use to indicate something is nil.
	nilStr = "<nil>"

	// TxGasLimit is the maximum amount of gas we allow in a single Tx.
	TxGasLimit uint64 = 4_000_000
	// DefaultGasLimit is the default gas to give a tx.
	// We want this to be low enough that we're not limiting Tx per block too much, but high enough to handle most.
	DefaultGasLimit uint64 = 500_000
	// For reference, consensus params on mainnet and testnet have max block gas at 60,000,000
)

func init() {
	cflags.DefaultGasLimit = DefaultGasLimit
}

// FlatFeesKeeper has the methods needed from a x/flatfees keeper that are needed for fee checking and collection.
type FlatFeesKeeper interface {
	CalculateMsgCost(ctx sdk.Context, msgs ...sdk.Msg) (upFront sdk.Coins, onSuccess sdk.Coins, err error)
	ExpandMsgs(msgs []sdk.Msg) ([]sdk.Msg, error)
}

// FeegrantKeeper defines the expected feegrant keeper.
type FeegrantKeeper interface {
	GetAllowance(ctx context.Context, granter sdk.AccAddress, grantee sdk.AccAddress) (feegrant.FeeAllowanceI, error)
	UseGrantedFees(ctx context.Context, granter, grantee sdk.AccAddress, fee sdk.Coins, msgs []sdk.Msg) error
}

// BankKeeper has the methods needed for a Bank keeper that are needed for fee checking and collection.
type BankKeeper interface {
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// FlatFeeGasMeter extends a GasMeter to have info about the flat fees associated with a Tx.
// It also logs gas and fee info for each Tx.
type FlatFeeGasMeter struct {
	storetypes.GasMeter

	// The upFrontCost and onSuccessCost in here covers all the Msgs provided in the tx and any sub-Msgs
	// that can be extracted from them. E.g. an authz Exec has sub-Msgs directly in it, so those are accounted for.
	// But there are some Msg types that cause other Msgs to be handled but don't contain all those sub-msgs,
	// e.g. a group Exec or smart contract ExecuteContract. To handle those, we will call ConsumeMsg from each msg
	// handler in order to identify any Msgs that we didn't know about at the start of the Tx. Later, when
	// collecting the remainder of the fee, we add in the costs of those extra msgs.
	// If the tx fails, we'll miss out on the up-front costs for the extra Msgs, but I'm not sure that
	// there's a reasonable way around that.
	// The addedFees is anything provided to ConsumeAddedFee by any endpoints that want to use
	// the tx fees to pay for stuff.

	// upFrontCost is the amount that should be collected before trying to execute the Msgs.
	upFrontCost sdk.Coins
	// onSuccessCost is the amount that should be collected iff all the Msgs are executed successfully.
	onSuccessCost sdk.Coins
	// extraMsgsCost is the amount that should be collected for msgs that weren't known when SetCosts was called.
	extraMsgsCost sdk.Coins
	// addedFees is the amount that might have been consumed manually during the processing of the tx.
	addedFees sdk.Coins

	// knownMsgs is a map of msg type url to a count of msgs of that type that have been accounted for, but not yet seen.
	// As msgs are consumed, the values will be decremented or, if at zero, the msg is added to extraMsgs.
	knownMsgs map[string]int
	// extraMsgs is a list of msgs that have been consumed, but weren't in the knownMsgs map.
	extraMsgs []sdk.Msg

	// fk is the x/flatfees keeper.
	fk FlatFeesKeeper

	// logger is a context logger to use to output gas info.
	logger log.Logger
	// msgTypeURLs is a list of all of the msg type urls of the Msgs in this tx. Only here for logging.
	msgTypeURLs []string
	// used is a map of description (e.g. "ReadFlat") to the amount of gas consumed. Only here for logging.
	used map[string]storetypes.Gas
	// counts is a map of description (e.g. "ReadFlat") to the number of times gas is consumed. Only here for logging.
	counts map[string]uint64
}

var _ storetypes.GasMeter = (*FlatFeeGasMeter)(nil)

func NewFlatFeeGasMeter(base storetypes.GasMeter, logger log.Logger, ffk FlatFeesKeeper) *FlatFeeGasMeter {
	return &FlatFeeGasMeter{
		GasMeter:  base,
		knownMsgs: make(map[string]int),
		logger:    logger,
		used:      make(map[string]storetypes.Gas),
		counts:    make(map[string]uint64),
		fk:        ffk,
	}
}

// SetCosts identifies the costs for the provided msgs and updates this FlatFeeGasMeter accordingly.
func (g *FlatFeeGasMeter) SetCosts(ctx sdk.Context, msgs []sdk.Msg) error {
	var err error
	msgs, err = g.fk.ExpandMsgs(msgs)
	if err != nil {
		return err
	}

	g.upFrontCost, g.onSuccessCost, err = g.fk.CalculateMsgCost(ctx, msgs...)
	if err != nil {
		return err
	}
	g.msgTypeURLs = msgTypeURLs(msgs)

	// Make sure we're starting with an empty knownMsgs map since everything else is fresh too.
	if len(g.knownMsgs) > 0 {
		g.knownMsgs = make(map[string]int)
	}
	for _, url := range g.msgTypeURLs {
		g.knownMsgs[url]++
	}
	g.extraMsgs = nil
	g.addedFees = nil

	return nil
}

// Finalize calculates the cost for any extra msgs and sets extraMsgsCost.
func (g *FlatFeeGasMeter) Finalize(ctx sdk.Context) error {
	if len(g.extraMsgs) == 0 {
		return nil
	}

	upFront, onSuccess, err := g.fk.CalculateMsgCost(ctx, g.extraMsgs...)
	if err != nil {
		return err
	}
	g.extraMsgsCost = upFront.Add(onSuccess...)
	return nil
}

// GetUpFrontCost is a getter for the upFrontCost field.
func (g *FlatFeeGasMeter) GetUpFrontCost() sdk.Coins {
	if g == nil {
		return nil
	}
	return g.upFrontCost
}

// GetOnSuccessCost is a getter for the onSuccessCost field.
func (g *FlatFeeGasMeter) GetOnSuccessCost() sdk.Coins {
	if g == nil {
		return nil
	}
	return g.onSuccessCost
}

// GetExtraMsgsCost is a getter for the extraMsgsCost field.
func (g *FlatFeeGasMeter) GetExtraMsgsCost() sdk.Coins {
	if g == nil {
		return nil
	}
	return g.extraMsgsCost
}

// GetAddedFees is a getter for the addedFees field.
func (g *FlatFeeGasMeter) GetAddedFees() sdk.Coins {
	if g == nil {
		return nil
	}
	return g.addedFees
}

// GetRequiredFee gets the total cost plus the additional fees.
func (g *FlatFeeGasMeter) GetRequiredFee() sdk.Coins {
	if g == nil {
		return nil
	}

	rv := g.upFrontCost
	rv = rv.Add(g.onSuccessCost...)
	rv = rv.Add(g.extraMsgsCost...)
	rv = rv.Add(g.addedFees...)

	return rv
}

// String implements stringer interface.
func (g *FlatFeeGasMeter) String() string {
	if g == nil {
		return nilStr
	}
	partFmt := "%17s: %s" // 17 = 15 (max length of desc) + 2 (extra for spacing).
	parts := []string{
		fmt.Sprintf(partFmt, "msg type urls", g.MsgCountsString()),
		fmt.Sprintf(partFmt, "up-front cost", g.upFrontCost.String()),
		fmt.Sprintf(partFmt, "on-success cost", g.onSuccessCost.String()),
		fmt.Sprintf(partFmt, "extra msgs cost", g.extraMsgsCost.String()),
		fmt.Sprintf(partFmt, "added fees", g.addedFees.String()),
		fmt.Sprintf(partFmt, "gas meter type", g.GasMeter.String()), // Should expand to multiple lines.
	}
	return fmt.Sprintf("FlatFeeGasMeter:\n%s", strings.Join(parts, "\n"))
}

// MsgCountsString returns a string that lists the msg type urls in this.
// Multiple msgs with the same type are grouped and the entry is prefixed with the count.
// Example output: "/provenance.name.v1.MsgBindNameRequest, 3x/provenance.attribute.v1.MsgAddAttributeRequest"
func (g *FlatFeeGasMeter) MsgCountsString() string {
	if g == nil {
		return nilStr
	}
	// Handle the most common (and easy) cases first.
	switch len(g.msgTypeURLs) {
	case 1:
		return g.msgTypeURLs[0]
	case 0:
		return "<none>"
	}

	counts := make(map[string]uint)
	parts := make([]string, 0, len(g.msgTypeURLs))
	for _, url := range g.msgTypeURLs {
		if _, ok := counts[url]; !ok {
			parts = append(parts, url)
		}
		counts[url]++
	}

	for i, url := range parts {
		if counts[url] > 1 {
			parts[i] = fmt.Sprintf("%dx%s", counts[url], url)
		}
	}

	return strings.Join(parts, ", ")
}

// GasUseString returns a multi-line string with details about the gas used for various things and also the total.
func (g *FlatFeeGasMeter) GasUseString() string {
	if g == nil {
		return nilStr
	}
	parts := make([]string, len(g.used))
	var total uint64
	for i, desc := range slices.Sorted(maps.Keys(g.used)) {
		parts[i] = fmt.Sprintf("%10d = %3dx %s", g.used[desc], g.counts[desc], desc)
		total += g.used[desc]
	}
	return fmt.Sprintf("%s\n%s\n%10d = Total gas", strings.Join(parts, "\n"), strings.Repeat("-", 30), total)
}

// RequiredFeeString returns a string representing the cost and possibly the breakdown.
// Examples:
//   - There's only an up-front cost: "123nhash".
//   - There's an on-success cost: "579nhash = 123nhash (up-front) + 456nhash (on success)".
//   - This FlatFeeGasMeter is nil or costs aren't set: "<nil>".
func (g *FlatFeeGasMeter) RequiredFeeString() string {
	if g == nil {
		return nilStr
	}

	// Coin{}.String() returns "<nil>". Coin{}.IsZero() panics.
	var parts []string
	if !g.upFrontCost.IsZero() {
		parts = append(parts, fmt.Sprintf("%s (up-front)", g.upFrontCost))
	}
	if !g.onSuccessCost.IsZero() {
		parts = append(parts, fmt.Sprintf("%s (on success)", g.onSuccessCost))
	}
	if !g.extraMsgsCost.IsZero() {
		parts = append(parts, fmt.Sprintf("%s (extra msgs)", g.extraMsgsCost))
	}
	if !g.addedFees.IsZero() {
		parts = append(parts, fmt.Sprintf("%s (added fees)", g.addedFees))
	}

	total := g.GetRequiredFee()
	if len(parts) > 1 {
		return fmt.Sprintf("%s = %s", total, strings.Join(parts, " + "))
	}
	return total.String()
}

// DetailsString returns a string with all the details about this FlatFeeGasMeter.
func (g *FlatFeeGasMeter) DetailsString() string {
	if g == nil {
		return nilStr
	}
	return fmt.Sprintf("FlatFeeGasMeter:\n  Msgs: %s\n  Cost: %s\n   Gas:\n%s",
		g.MsgCountsString(), g.RequiredFeeString(), g.GasUseString())
}

// ConsumeGas keeps track of amounts and counts and reports the info to the base GasMeter.
func (g *FlatFeeGasMeter) ConsumeGas(amount storetypes.Gas, descriptor string) {
	g.used[descriptor] += amount
	g.counts[descriptor]++
	g.GasMeter.ConsumeGas(amount, descriptor)
}

// GasConsumed reports the amount of gas consumed at Log.Info level and returns the base GasMeter's gas consumed.
func (g *FlatFeeGasMeter) GasConsumed() storetypes.Gas {
	g.logger.Debug("Stack trace:\n" + string(debug.Stack()))
	g.logger.Info(g.DetailsString())
	return g.GasMeter.GasConsumed()
}

// ConsumeMsg updates this gas meter to indicate that the provided msg has been run.
func (g *FlatFeeGasMeter) ConsumeMsg(msg sdk.Msg) {
	url := sdk.MsgTypeURL(msg)
	if (g.knownMsgs[url]) > 0 {
		g.knownMsgs[url]--
		return
	}
	g.extraMsgs = append(g.extraMsgs, msg)
	g.msgTypeURLs = append(g.msgTypeURLs, url)
}

// ConsumeAddedFee adds the provided fee to this gas meter so that it can be
// collected once all the msgs in the tx have run successfully.
func (g *FlatFeeGasMeter) ConsumeAddedFee(fee sdk.Coins) {
	g.addedFees = g.addedFees.Add(fee...)
}

// adjustCostsForUnitTests will the chain id. If it indicates that we're running in a unit test,
// it will adjust the required costs accordingly. This exists so that we didn't have to redo a
// lot of unit tests when we switched to flat fees.
func (g *FlatFeeGasMeter) adjustCostsForUnitTests(logger log.Logger, chainID string, feeProvided sdk.Coins) {
	switch {
	case len(chainID) == 0:
		// Probably a pretty simple unit test. Use what was provided, all charged up-front.
		g.upFrontCost = feeProvided
		g.onSuccessCost = nil
		logger.Debug(fmt.Sprintf("adjustCostsForUnitTests: Using provided fee for test tx cost: %q.", feeProvided))
	case isTestChainID(chainID):
		// One of the more complex unit tests, possibly involving actually running a chain. No fee.
		g.upFrontCost = nil
		g.onSuccessCost = nil
		logger.Debug("adjustCostsForUnitTests: Using zero for test tx cost.")
	default:
		logger.Debug("adjustCostsForUnitTests: Not a unit test. Not adjusting tx cost.")
	}
}

// isTestChainID returns true if the chain id is one of the special ones used for unit tests.
func isTestChainID(chainID string) bool {
	return chainID == SimAppChainID || chainID == pioconfig.SimAppChainID || strings.HasPrefix(chainID, "testchain")
}

// GetFlatFeeGasMeter will extract the flat fee gas meter from the ctx.
func GetFlatFeeGasMeter(ctx sdk.Context) (*FlatFeeGasMeter, error) {
	rv, ok := ctx.GasMeter().(*FlatFeeGasMeter)
	if !ok {
		return nil, sdkerrors.ErrLogic.Wrapf("gas meter is not a FlatFeeGasMeter: %T", ctx.GasMeter())
	}
	return rv, nil
}

// ConsumeAdditionalFee will get the FlatFeeGasMeter from the context and call ConsumeAddedFee with the provided fee.
// Does nothing if the fee is zero, or the context doesn't have a FlatFeeGasMeter.
func ConsumeAdditionalFee(ctx sdk.Context, fee sdk.Coins) {
	if fee.IsZero() {
		return
	}

	// There are some legitimate reasons why we might not get a flat fee gas meter here
	// (e.g. during a gov prop). In those cases, we just skip consuming this fee and move on.
	feeGasMeter, err := GetFlatFeeGasMeter(ctx)
	if err == nil && feeGasMeter != nil {
		feeGasMeter.ConsumeAddedFee(fee)
	}
}

// ProvSetUpContextDecorator creates and sets the flat-fee GasMeter in the Context and wraps the
// next AnteHandler with a defer clause to recover from any downstream OutOfGas panics in the
// AnteHandler chain to return an error with information on gas provided and gas used.
// CONTRACT: Must be first decorator in the chain.
// CONTRACT: Tx must implement GasTx interface.
// This is similar to "github.com/cosmos/cosmos-sdk/x/auth/ante".SetUpContextDecorator
// except we set and check the gas limits a little differently.
type ProvSetUpContextDecorator struct {
	ffk FlatFeesKeeper
}

func NewProvSetUpContextDecorator(ffk FlatFeesKeeper) ProvSetUpContextDecorator {
	return ProvSetUpContextDecorator{ffk: ffk}
}

func (d ProvSetUpContextDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	defer func() {
		if err != nil {
			ctx.Logger().Error("Error being returned from ProvSetUpContextDecorator.",
				"error", err, "simulate", simulate, "IsCheckTx", ctx.IsCheckTx())
		}
	}()

	ctx.Logger().Debug("Starting ProvSetUpContextDecorator.", "simulate", simulate, "IsCheckTx", ctx.IsCheckTx(), "IsInitGenesis", IsInitGenesis(ctx))

	// All transactions must implement FeeTx.
	feeTx, err := GetFeeTx(tx)
	if err != nil {
		// Set a gas meter with limit 0 as to prevent an infinite gas meter attack during runTx.
		newCtx = ante.SetGasMeter(simulate, ctx, 0)
		return newCtx, err
	}

	// Our app.Simulate returns the gas used amount equal to the amount of nhash required for the fee, and we tell
	// folks to use gas-prices of 1nhash. This results in the client including the correct amount of fee, but means
	// the client will send that amount as the gas wanted too. In such cases, we want to use the default.
	// We also want to use the default if no gas wanted was provided.
	// But if another amount of gas is wanted, we want to use that.
	gasWanted := feeTx.GetGas()
	switch {
	case gasWanted == DefaultGasLimit: // Do nothing.
		ctx.Logger().Debug("Using provided gas limit.", "gas wanted", gasWanted)
	case gasWanted <= 0 || gasWanted > TxGasLimit:
		// If no gas wanted was given, use the default.
		// If gas wanted is strictly more than the Tx gas limit, assume it's from Simulate, and use the default.
		// If it's equal to the TxGasLimit, we still want to use it as provided, though.
		ctx.Logger().Debug("Gas limit out of bounds. Using default gas limit.", "original gas wanted", gasWanted, "actual gas wanted", DefaultGasLimit)
		gasWanted = DefaultGasLimit
	default:
		fee := feeTx.GetFee()
		if len(fee) == 1 && fee[0].Denom == pioconfig.GetProvConfig().FeeDenom && fee[0].Amount.Equal(sdkmath.NewIntFromUint64(gasWanted)) {
			// The gas wanted is equal to the amount of nhash provided in the fee.
			// Assume they simulated with gas-prices 1nhash, and switch to the default gas.
			ctx.Logger().Debug("Gas limit equals fee amount. Using default gas limit.",
				"original gas wanted", gasWanted,
				"actual gas wanted", DefaultGasLimit,
				"fee", fee.String(),
			)
			gasWanted = DefaultGasLimit
		}
	}

	// Set a generic gas meter in the context with the appropriate amount of gas.
	// Note that SetGasMeter uses an infinite gas meter if simulating or at height 0 (init genesis).
	newCtx = ante.SetGasMeter(simulate, ctx, gasWanted)
	// Now wrap that gas meter in our flat-fee gas meter.
	newCtx = ctx.WithGasMeter(NewFlatFeeGasMeter(newCtx.GasMeter(), newCtx.Logger(), d.ffk))
	// Note: We don't set the costs yet, because we want to check with the circuit breaker (another antehandler) first.

	// Ensure that the requested gas does not exceed the configured block maximum.
	// We skip this if there are no block consensus params because that indicates there shouldn't be any gas limits.
	if bp := ctx.ConsensusParams().Block; bp != nil {
		maxBlockGas := bp.GetMaxGas()
		newCtx.Logger().Debug("Consensus Params available. Checking max block gas.", "Block Params", bp, "maxBlockGas", maxBlockGas, "gas wanted", gasWanted)
		// If there exists a maximum block gas limit, we must ensure that the tx does not exceed it.
		if maxBlockGas > 0 && gasWanted > uint64(maxBlockGas) {
			return newCtx, sdkerrors.ErrInvalidGasLimit.Wrapf("tx gas limit %d exceeds block max gas %d", gasWanted, maxBlockGas)
		}
	} else {
		newCtx.Logger().Debug("No consensus params. Skipping max block gas check.")
	}

	// Decorator will catch an OutOfGasPanic caused in the next antehandler
	// AnteHandlers must have their own defer/recover in order for the BaseApp
	// to know how much gas was used! This is because the GasMeter is created in
	// the AnteHandler, but if it panics the context won't be set properly in
	// runTx's recover call.
	defer func() {
		if r := recover(); r != nil {
			switch rType := r.(type) {
			case storetypes.ErrorOutOfGas:
				err = sdkerrors.ErrOutOfGas.Wrapf("out of gas in location: %v; gasWanted: %d, gasUsed: %d",
					rType.Descriptor, newCtx.GasMeter().Limit(), newCtx.GasMeter().GasConsumed())
			default:
				panic(r)
			}
		}
	}()

	return next(newCtx, tx, simulate)
}

// FlatFeeSetupDecorator is an AnteHandler that calculates costs for the msgs, and ensures a sufficient fee is provided.
type FlatFeeSetupDecorator struct{}

func NewFlatFeeSetupDecorator() FlatFeeSetupDecorator {
	return FlatFeeSetupDecorator{}
}

func (d FlatFeeSetupDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	defer func() {
		if err != nil {
			ctx.Logger().Error("Error being returned from FlatFeeSetupDecorator.",
				"error", err, "simulate", simulate, "IsCheckTx", ctx.IsCheckTx())
		}
	}()

	feeTx, err := GetFeeTx(tx)
	if err != nil {
		return ctx, err
	}

	gasMeter, err := GetFlatFeeGasMeter(ctx)
	if err != nil {
		return ctx, err
	}

	// Calculate and set the costs in the gas meter.
	err = gasMeter.SetCosts(ctx, feeTx.GetMsgs())
	if err != nil {
		return ctx, fmt.Errorf("could not calculate msg costs: %w", err)
	}
	feeProvided := feeTx.GetFee()
	gasMeter.adjustCostsForUnitTests(ctx.Logger(), ctx.ChainID(), feeProvided)

	// Make sure the fee provided is enough. There's a chance that costs/fees are added during the execution
	// of the tx, so we'll check this again later. We check it here too, though, in order to skip processing
	// the tx msgs if we know now that there isn't enough fee for what's been provided.
	// Skip if simulating since the fee is probably what they're trying to find out.
	// Skip during init genesis too since those should be free (and there's no one to pay).
	if !simulate && !IsInitGenesis(ctx) {
		reqFee := gasMeter.GetRequiredFee()
		ctx.Logger().Debug("Validating fee", "required", reqFee.String(), "provided", feeProvided.String())
		err = validateFeeAmount(reqFee, feeProvided)
		if err != nil {
			return ctx, err
		}
	}

	return next(ctx, tx, simulate)
}

// validateFeeAmount returns an error if the required fee is more than the provided fee.
func validateFeeAmount(required sdk.Coins, provided sdk.Coins) error {
	// sdk.Coins.Validate() doesn't allow for coins with a zero amount, but we want to allow that here.
	var nonZero sdk.Coins
	for _, coin := range provided {
		if coin.IsNil() || !coin.IsZero() {
			// Coin{}.IsZero() will panic if the amount is nil, so we have to check for that first.
			// We include the nil coins so that we get the correct error message from them.
			nonZero = append(nonZero, coin)
		}
	}
	if err := nonZero.Validate(); err != nil {
		return sdkerrors.ErrInsufficientFee.Wrapf("fee provided %q is invalid: %v", provided, err)
	}

	_, hasNeg := provided.SafeSub(required...)
	if hasNeg {
		return sdkerrors.ErrInsufficientFee.Wrapf("fee required: %q, fee provided: %q", required, provided)
	}
	return nil
}

// DeductFeeDecorator is an AnteHandler that collects the up-front cost and ensures the fee
// payer account has enough in it to cover the entire fee that has been provided.
type DeductFeeDecorator struct {
	ak ante.AccountKeeper
	bk BankKeeper
	fk ante.FeegrantKeeper
}

func NewDeductFeeDecorator(ak ante.AccountKeeper, bk BankKeeper, fk ante.FeegrantKeeper) DeductFeeDecorator {
	return DeductFeeDecorator{ak: ak, bk: bk, fk: fk}
}

func (d DeductFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	defer func() {
		if err != nil {
			ctx.Logger().Error("Error being returned from DeductFeeDecorator.",
				"error", err, "simulate", simulate, "IsCheckTx", ctx.IsCheckTx())
		}
	}()

	if err = d.checkDeductUpFrontCost(ctx, tx, simulate); err != nil {
		return ctx, err
	}
	return next(ctx, tx, simulate)
}

// checkDeductUpFrontCost identifies the fee payer (possibly using a fee grant), makes sure they have enough
// in their account to cover the entire fee that was provided, and collects the up-front cost from them.
func (d DeductFeeDecorator) checkDeductUpFrontCost(ctx sdk.Context, tx sdk.Tx, simulate bool) error {
	ctx.Logger().Debug("Starting checkDeductUpFrontCost.")
	if addr := d.ak.GetModuleAddress(authtypes.FeeCollectorName); addr == nil {
		return sdkerrors.ErrLogic.Wrapf("%s module account has not been set", authtypes.FeeCollectorName)
	}

	feeTx, err := GetFeeTx(tx)
	if err != nil {
		return err
	}

	gasMeter, err := GetFlatFeeGasMeter(ctx)
	if err != nil {
		return err
	}

	upFrontCost := gasMeter.GetUpFrontCost()
	deductFeesFrom, usedFeeGrant, err := GetFeePayerUsingFeeGrant(ctx, d.fk, feeTx, upFrontCost, feeTx.GetMsgs())
	if err != nil {
		return err
	}

	// Make sure the account has enough to cover the whole cost. We do this check in order to prevent the futile
	// execution of the tx msgs just to find out at the end that the account doesn't have enough to pay for it.
	// When simulating, we don't care about the fees being paid.
	fullFee := feeTx.GetFee()
	if !simulate {
		// Make sure the paying account exists before we check their balance.
		if d.ak.GetAccount(ctx, deductFeesFrom) == nil {
			return sdkerrors.ErrUnknownAddress.Wrapf("fee payer address %q does not exist", deductFeesFrom.String())
		}
		// Now we can check their balance for the full fee they're expected to pay.
		// We check the full balance here (instead of just the up-front cost) in order to prevent
		// the extra work of running all the Msgs when it's unlikely they'll be able to pay.
		// This means that a tx cannot be paid for using funds received during that tx.
		// I feel like it will be really rare that a tx sends funds to the fee payer.
		if err = validateHasBalance(ctx, d.bk, deductFeesFrom, fullFee); err != nil {
			return err
		} else {
			ctx.Logger().Debug("Fee payer has enough in their account to pay.", "fee provided", fullFee.String())
		}
	}

	// Pay the up-front cost.
	// When simulating, we don't care about the fees being paid.
	// During InitGenesis, there's no fees to pay (and no one to pay them).
	if !simulate && !IsInitGenesis(ctx) && !upFrontCost.IsZero() {
		ctx2 := ctx
		if usedFeeGrant {
			ctx2 = internalsdk.WithFeeGrantInUse(ctx)
		}
		if err = PayFee(ctx2, d.bk, deductFeesFrom, upFrontCost); err != nil {
			return cerrs.Wrapf(err, "could not collect up-front fee of %q", upFrontCost.String())
		}
		ctx.Logger().Debug("Up Front cost collected.", "up-front cost", upFrontCost.String())
	} else {
		ctx.Logger().Debug("Skipping collection of up-front cost.", "up-front cost", upFrontCost.String(), "simulate", simulate)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(sdk.EventTypeTx,
			sdk.NewAttribute(sdk.AttributeKeyFee, fullFee.String()),
			sdk.NewAttribute(sdk.AttributeKeyFeePayer, deductFeesFrom.String()),
		),
		sdk.NewEvent(sdk.EventTypeTx,
			sdk.NewAttribute(AttributeKeyMinFeeCharged, upFrontCost.String()),
			sdk.NewAttribute(sdk.AttributeKeyFeePayer, deductFeesFrom.String()),
		),
	})

	return nil
}

// validateHasBalance returns an error of the provided account does not have at least the required funds.
func validateHasBalance(ctx sdk.Context, bk BankKeeper, addr sdk.AccAddress, required sdk.Coins) error {
	if required.IsZero() {
		return nil
	}

	var bal sdk.Coins
	for _, coin := range required {
		bal = append(bal, bk.GetBalance(ctx, addr, coin.Denom))
	}
	_, hasNeg := bal.SafeSub(required...)
	if hasNeg {
		return sdkerrors.ErrInsufficientFunds.Wrapf("account %q balance: %q, required: %q",
			addr.String(), bal.String(), required.String())
	}

	return nil
}

// GetFeePayerUsingFeeGrant identifies the fee payer, updating the applicable feegrant if appropriate.
// Returns the address responsible for paying the fees, and whether a feegrant was used.
func GetFeePayerUsingFeeGrant(ctx sdk.Context, feegrantKeeper ante.FeegrantKeeper, feeTx sdk.FeeTx, amount sdk.Coins, msgs []sdk.Msg) (sdk.AccAddress, bool, error) {
	feePayer := sdk.AccAddress(feeTx.FeePayer())
	feeGranter := sdk.AccAddress(feeTx.FeeGranter())
	deductFeesFrom := feePayer
	usedFeeGrant := false

	// if feegranter set deduct base fee from feegranter account.
	// this works with only when feegrant enabled.
	if feeGranter != nil && !bytes.Equal(feeGranter, feePayer) {
		if feegrantKeeper == nil {
			return nil, false, sdkerrors.ErrInvalidRequest.Wrap("fee grants are not enabled")
		}
		if !amount.IsZero() {
			err := feegrantKeeper.UseGrantedFees(ctx, feeGranter, feePayer, amount, msgs)
			if err != nil {
				return nil, false, cerrs.Wrapf(err, "failed to use fee grant: granter: %s, grantee: %s, fee: %q, msgs: %q",
					feeGranter, feePayer, amount, msgTypeURLs(msgs))
			}
		}
		deductFeesFrom = feeGranter
		usedFeeGrant = true
	}

	return deductFeesFrom, usedFeeGrant, nil
}

// PayFee sends the fee from the addr to the fee collector.
func PayFee(ctx sdk.Context, bankKeeper BankKeeper, addr sdk.AccAddress, fee sdk.Coins) error {
	if fee.IsZero() {
		return nil
	}
	if !fee.IsValid() {
		return sdkerrors.ErrInsufficientFee.Wrapf("invalid fee amount: %s", fee)
	}

	err := bankKeeper.SendCoinsFromAccountToModule(ctx, addr, authtypes.FeeCollectorName, fee)
	if err != nil {
		return sdkerrors.ErrInsufficientFunds.Wrapf("%v: account: %s:", err, addr)
	}
	return nil
}

// msgTypeURLs returns a slice of MsgTypeURL that correspond to the provided Msgs.
func msgTypeURLs(msgs []sdk.Msg) []string {
	if msgs == nil {
		return nil
	}
	rv := make([]string, len(msgs))
	for i, msg := range msgs {
		rv[i] = sdk.MsgTypeURL(msg)
	}
	return rv
}

// FlatFeePostHandler is an sdk.PostDecorator that collects the fee remainder once all the msgs have run.
type FlatFeePostHandler struct {
	bk BankKeeper
	fk ante.FeegrantKeeper
}

var _ sdk.PostDecorator = (*FlatFeePostHandler)(nil)

func NewFlatFeePostHandler(bk BankKeeper, fk ante.FeegrantKeeper) FlatFeePostHandler {
	return FlatFeePostHandler{bk: bk, fk: fk}
}

// PostHandle collects the rest of the fees for a tx.
func (h FlatFeePostHandler) PostHandle(ctx sdk.Context, tx sdk.Tx, simulate, success bool, next sdk.PostHandler) (newCtx sdk.Context, err error) {
	defer func() {
		if err != nil {
			ctx.Logger().Error("Error being returned from FlatFeePostHandler.",
				"error", err, "simulate", simulate, "IsCheckTx", ctx.IsCheckTx())
		}
	}()

	// If it wasn't successful, there's nothing to do in here.
	if !success {
		ctx.Logger().Debug("Skipping FlatFeePostHandler because tx was not successful.")
		return next(ctx, tx, simulate, success)
	}

	feeTx, err := GetFeeTx(tx)
	if err != nil {
		return ctx, err
	}

	gasMeter, err := GetFlatFeeGasMeter(ctx)
	if err != nil {
		return ctx, err
	}
	err = gasMeter.Finalize(ctx)
	if err != nil {
		return ctx, fmt.Errorf("could not finalize gas meter: %w", err)
	}

	reqFee := gasMeter.GetRequiredFee()
	feeProvided := feeTx.GetFee()

	extraMsgsCost := gasMeter.GetExtraMsgsCost()
	addedFees := gasMeter.GetAddedFees()
	newCharges := addedFees.Add(extraMsgsCost...)
	if !newCharges.IsZero() && !simulate && !IsInitGenesis(ctx) {
		// There were extra msg costs added since we set up the gas meter. Re-check that
		// the full fee is enough to cover everything (now that we know what everything is).
		err = validateFeeAmount(reqFee, feeProvided)
		if err != nil {
			return ctx, err
		}
	}

	// We collect the entirety of the fee provided, even if it's more than what's required.
	// We've already collected the up-front cost, though, so we take that out of the full fee and collect what's left.
	upFrontCost := gasMeter.GetUpFrontCost()
	var uncharged sdk.Coins
	if !simulate && !IsInitGenesis(ctx) {
		uncharged = feeProvided.Sub(upFrontCost...)
		ctx.Logger().Debug("Uncharged amount calculated using fee provided.",
			"fee provided", feeProvided, "up-front cost", upFrontCost.String(), "uncharged", uncharged.String())
	} else {
		// If simulating, pretend the reqFee is what was provided since there might not have been a fee provided.
		uncharged = reqFee.Sub(upFrontCost...)
		ctx.Logger().Debug("Uncharged amount calculated using required fee.",
			"required fee", reqFee.String(), "up-front cost", upFrontCost.String(), "uncharged", uncharged.String())
	}
	deductFeesFrom, usedFeeGrant, err := GetFeePayerUsingFeeGrant(ctx, h.fk, feeTx, uncharged, feeTx.GetMsgs())
	if err != nil {
		return ctx, err
	}

	// Skip checking that the deductFeesFrom account exists since we checked that earlier and it shouldn't change.
	// We skip checking the balance because the bank module will do that for us during the transfer.

	// Pay whatever is left.
	// When simulating, we don't care about the fees being paid.
	// During InitGenesis, there's no fees to pay (and no one to pay them).
	if !simulate && !IsInitGenesis(ctx) && !uncharged.IsZero() {
		ctx2 := ctx
		if usedFeeGrant {
			ctx2 = internalsdk.WithFeeGrantInUse(ctx)
		}
		if err = PayFee(ctx2, h.bk, deductFeesFrom, uncharged); err != nil {
			return ctx, cerrs.Wrapf(err, "could not collect fee remainder %q upon success", uncharged.String())
		}
		ctx.Logger().Debug("Collected remaining amount.", "remaining cost", uncharged.String())
	} else {
		ctx.Logger().Debug("Skipping collection of remaining cost.", "remaining cost", uncharged.String(), "simulate", simulate)
	}

	var overage sdk.Coins
	if !simulate && !IsInitGenesis(ctx) {
		overage = feeProvided.Sub(reqFee...)
	}
	onSuccessCost := reqFee.Sub(upFrontCost...)
	ctx.EventManager().EmitEvent(CreateFeeEvent(deductFeesFrom, upFrontCost, onSuccessCost, overage))
	ctx.Logger().Debug("Fee event created.",
		sdk.AttributeKeyFeePayer, deductFeesFrom.String(),
		AttributeKeyBaseFee, upFrontCost.String(),
		AttributeKeyAdditionalFee, onSuccessCost.String(),
		AttributeKeyFeeOverage, overage.String(),
	)

	return next(ctx, tx, simulate, success)
}

// CreateFeeEvent creates the event we emit containing tx fee information.
func CreateFeeEvent(feePayer sdk.AccAddress, baseFee, additionalFee, overage sdk.Coins) sdk.Event {
	rv := sdk.NewEvent(sdk.EventTypeTx,
		sdk.NewAttribute(sdk.AttributeKeyFeePayer, feePayer.String()),
		sdk.NewAttribute(AttributeKeyBaseFee, baseFee.String()),
	)
	if !additionalFee.IsZero() {
		rv = rv.AppendAttributes(sdk.NewAttribute(AttributeKeyAdditionalFee, additionalFee.String()))
	}
	if !overage.IsZero() {
		rv = rv.AppendAttributes(sdk.NewAttribute(AttributeKeyFeeOverage, overage.String()))
	}
	return rv
}
