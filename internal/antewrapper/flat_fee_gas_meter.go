package antewrapper

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

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
	// the tx fees to pay extra for stuff.

	// upFrontCost is the amount that should be collected before trying to execute the Msgs.
	// This is set by the SetCosts method.
	upFrontCost sdk.Coins
	// onSuccessCost is the amount that should be collected iff all the Msgs are executed successfully.
	// This is set by the SetCosts method.
	onSuccessCost sdk.Coins
	// extraMsgsCost is the amount that should be collected for msgs that weren't known when SetCosts was called.
	// This is set by the Finalize method based on the contents of the extraMsgs field.
	extraMsgsCost sdk.Coins
	// addedFees is the amount that might have been consumed manually during the processing of the tx.
	// This is the sum of the amounts provided to the ConsumeAddedFee method, and/or ConsumeAdditionalFee helper function.
	addedFees sdk.Coins

	// knownMsgs is a map of msg type url to a count of msgs of that type that have been accounted for, but not yet seen.
	// As msgs are consumed, the values will be decremented or, if at zero, the msg is added to extraMsgs.
	knownMsgs map[string]int
	// extraMsgs is a list of msgs that have been consumed, but weren't in the knownMsgs map. I.e. these are
	// messages that have been processed, but not paid for. Once all msgs in a tx are processed, the Finalize
	// method is used to calculate the costs for these messages and populate the extraMsgsCost field.
	extraMsgs []sdk.Msg

	// ffk is the x/flatfees keeper.
	ffk FlatFeesKeeper

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
		ffk:       ffk,
	}
}

// SetCosts identifies the costs for the provided msgs and updates this FlatFeeGasMeter accordingly.
func (g *FlatFeeGasMeter) SetCosts(ctx sdk.Context, msgs []sdk.Msg) error {
	// Make sure we're starting fresh.
	g.upFrontCost, g.onSuccessCost = nil, nil
	g.extraMsgsCost, g.addedFees = nil, nil
	if g.knownMsgs == nil || len(g.knownMsgs) > 0 {
		g.knownMsgs = make(map[string]int)
	}
	g.extraMsgs, g.msgTypeURLs = nil, nil
	// We don't reset .used or .counts because those have info related to the
	// underlying gas meter. Since that's not changing, we leave them alone here.

	var err error
	msgs, err = g.ffk.ExpandMsgs(msgs)
	if err != nil {
		return err
	}

	g.upFrontCost, g.onSuccessCost, err = g.ffk.CalculateMsgCost(ctx, msgs...)
	if err != nil {
		return err
	}

	g.msgTypeURLs = msgTypeURLs(msgs)
	for _, url := range g.msgTypeURLs {
		g.knownMsgs[url]++
	}

	return nil
}

// Finalize calculates the cost for any extra msgs and sets extraMsgsCost.
func (g *FlatFeeGasMeter) Finalize(ctx sdk.Context) error {
	g.extraMsgsCost = nil
	if len(g.extraMsgs) == 0 {
		return nil
	}

	upFront, onSuccess, err := g.ffk.CalculateMsgCost(ctx, g.extraMsgs...)
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
// This is the amount owed due to the processing of msgs that weren't directly in the Tx, e.g. a smart contract issuing a MsgSend.
func (g *FlatFeeGasMeter) GetExtraMsgsCost() sdk.Coins {
	if g == nil {
		return nil
	}
	return g.extraMsgsCost
}

// GetAddedFees is a getter for the addedFees field.
// This is the sum of the amounts provided to either the ConsumeAddedFee method, or ConsumeAdditionalFee helper function.
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

// adjustCostsForUnitTests will examine the chain id. If it indicates that we're running in a unit test,
// it will adjust the required costs accordingly. This exists so that we didn't have to redo a
// lot of unit tests when we switched to flat fees.
func (g *FlatFeeGasMeter) adjustCostsForUnitTests(chainID string, feeProvided sdk.Coins) {
	switch {
	case len(chainID) == 0:
		// Probably a pretty simple unit test. Use what was provided, all charged up-front.
		g.upFrontCost = feeProvided
		g.onSuccessCost = nil
		g.logger.Debug("adjustCostsForUnitTests: Using provided fee for test tx cost.", "fee_provided", lazyCzStr(feeProvided))
	case isTestChainID(chainID):
		// One of the more complex unit tests, possibly involving actually running a chain. No fee.
		g.upFrontCost = nil
		g.onSuccessCost = nil
		g.logger.Debug("adjustCostsForUnitTests: Using zero for test tx cost.")
	default:
		g.logger.Debug("adjustCostsForUnitTests: Not a unit test. Not adjusting tx cost.")
	}
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
