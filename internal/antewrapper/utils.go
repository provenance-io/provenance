package antewrapper

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"strings"

	cerrs "cosmossdk.io/errors"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/x/feegrant"

	cflags "github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/internal/provutils"
)

const (
	// AttributeKeyBaseFee is the amount of fee charged up-front, even if the tx fails.
	AttributeKeyBaseFee = "basefee"
	// AttributeKeyAdditionalFee is the amount of fee required upon success.
	AttributeKeyAdditionalFee = "additionalfee"
	// AttributeKeyFeeOverage is the amount paid on top of the required amounts.
	AttributeKeyFeeOverage = "fee_overage"
	// AttributeKeyFeeTotal is the total amount paid in fees.
	AttributeKeyFeeTotal = "total"

	// AttributeKeyMinFeeCharged is also the up-front cost, but used in a different Tx event.
	// If a transaction fails, this value is copied to the "fee" attribute in the SDK's standard fee event.
	AttributeKeyMinFeeCharged = "min_fee_charged"

	// nilStr is a string to use to indicate something is nil.
	nilStr = "<nil>"

	SimAppChainID = "simapp-unit-testing"

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

// lazyCzStr is an alias to provutils.NewLazyStringer(val) that needs fewer characters to type, and is typed to Coins.
var lazyCzStr = provutils.NewLazyStringer[sdk.Coins]

type (
	// FlatFeesKeeper has the methods needed from a x/flatfees keeper that are needed for fee checking and collection.
	FlatFeesKeeper interface {
		CalculateMsgCost(ctx sdk.Context, msgs ...sdk.Msg) (upFront sdk.Coins, onSuccess sdk.Coins, err error)
		ExpandMsgs(msgs []sdk.Msg) ([]sdk.Msg, error)
	}

	// FeegrantKeeper defines the expected feegrant keeper.
	FeegrantKeeper interface {
		GetAllowance(ctx context.Context, granter sdk.AccAddress, grantee sdk.AccAddress) (feegrant.FeeAllowanceI, error)
		UseGrantedFees(ctx context.Context, granter, grantee sdk.AccAddress, fee sdk.Coins, msgs []sdk.Msg) error
	}

	// BankKeeper has the methods needed for a Bank keeper that are needed for fee checking and collection.
	BankKeeper interface {
		SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
		GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	}
)

// GetGasWanted returns the amount of gas that this Tx wants.
// Our Simulate method returns the amount of fee as the gas wanted and we tell people to use gas-prices 1nhash.
// That causes all the clients to provide that amount of fee as the gas wanted, though.
// In order to stay compatible with all those clients/wallets, we handle that case here.
//
// E.g. Say a msg costs $0.50 and 1 hash costs $0.10. The msg will cost 5 hash or 5,000,000,000 nhash.
// Our max block gas s 60,000,000, though (set in consensus params). So if we only relied on feeTx.GetGas(),
// the tx wouldn't fit in a block (not to mention be more than the 4,000,000 tx gas limit).
//
// Also factoring into this is that, prior to flat fees, we told everyone to use gas-prices 1905nhash (or 19050nhash on testnet).
// Anyone still using that will end up with a huge fee. So we return max uint64 and an error if that is detected.
func GetGasWanted(logger log.Logger, feeTx sdk.FeeTx) (uint64, error) {
	txGas := feeTx.GetGas()
	logger = logger.With("method", "GetGasWanted", "feeTx.GetGas()", txGas)
	if txGas == 0 {
		// If no gas was provided, use the default instead.
		// This allows users to skip the simulation and just provide the fee without worrying about gas.
		// This could also happen during a free Tx that was simulated first.
		logger.Debug("No gas limit provided. Using default.", "returning", DefaultGasLimit)
		return DefaultGasLimit, nil
	}

	fee := feeTx.GetFee()
	logger = logger.With("feeTx.GetFee()", fee.String())
	hasNhash, feeNhash := fee.Find(pioconfig.GetProvConfig().FeeDenom)
	if !hasNhash {
		// If no nhash is in the fee, but gas was provided, they probably didn't simulate the tx,
		// and instead set the values from previously known amounts. So use the gas they provided.
		// Basically, we can't identify it as a special case, so we keep old behavior.
		logger.Debug("No nhash in fee. Using provided gas limit.", "returning", txGas)
		return txGas, nil
	}

	txGasInt := sdkmath.NewIntFromUint64(txGas)
	if feeNhash.Amount.Equal(txGasInt) {
		// The gas wanted is equal to the amount of nhash provided in the fee.
		// Assume they simulated with --gas-prices 1nhash, and use the default gas.
		logger.Debug("Gas limit equals fee amount. Using default gas limit.", "returning", DefaultGasLimit)
		return DefaultGasLimit, nil
	}

	// Prior to flat-fees, we told everyone to use gas-prices of 1905nhash (or 19050nhash on testnet).
	// If they're still using that, they're providing way too much as a fee and need to update their client settings.
	// To prevent charging for what would be a pretty costly mistake, we return max uint in such cases.
	// This gives users have a chance to update their clients without paying 1905 times what's needed.
	if isOldGasPrices(feeNhash.Amount, txGasInt) {
		// There's a very small chance that this catches a legitimate situation where the tx was not simulated;
		// e.g. the user defined the gas and fee on their own and the numbers worked out just wrong.
		// They can get around this by bumping the gas by 1.
		logger.Debug("Gas limit indicates old gas-prices value. Using max uint64.", "returning", uint64(math.MaxUint64))
		return math.MaxUint64, fmt.Errorf("old gas-prices value detected; always use 1nhash")
	}

	// It's not a known special case, so keep old behavior.
	logger.Debug("Using provided gas limit.", "returning", txGas)
	return txGas, nil
}

var (
	oldMainnetGasPricesAmt = sdkmath.NewInt(1905)
	oldTestnetGasPricesAmt = sdkmath.NewInt(19050)
)

// isOldGasPrices returns true if the nhash and gas amounts indicate that a tx had one of our old gas prices.
// Prior to flat-fees, we told everyone to use gas-prices of 1905nhash (or 19050nhash on testnet).
func isOldGasPrices(nhash, gas sdkmath.Int) bool {
	return nhash.Equal(gas.Mul(oldMainnetGasPricesAmt)) || nhash.Equal(gas.Mul(oldTestnetGasPricesAmt))
}

// txGasLimitShouldApply returns true iff the tx gas limit should be applied.
func txGasLimitShouldApply(chainID string, msgs []sdk.Msg) bool {
	// Skip the tx gas limit for unit tests and simulations; this way, we didn't have
	// to update all the existing unit tests when we introduced this limit.
	// Also, skip the limit for gov props so that they can be used for Txs that require a lot of gas.
	// One of the primary reasons for the tx gas limit is to restrict WASM code submission.
	// There's so much data in those that they always require more gas than the tx gas limit, but
	// if submitted as part of a gov prop, it should be allowed.
	return !isTestChainID(chainID) && !isOnlyGovProps(msgs)
}

// isTestChainID returns true if the chain id is one of the special ones used for unit tests.
func isTestChainID(chainID string) bool {
	return len(chainID) == 0 || chainID == SimAppChainID || chainID == pioconfig.SimAppChainID || strings.HasPrefix(chainID, "testchain")
}

// isOnlyGovProps returns true if there's at least one msg, and all msgs are a MsgSubmitProposal.
func isOnlyGovProps(msgs []sdk.Msg) bool {
	// If there are no messages, there are no gov messages, so return false.
	if len(msgs) == 0 {
		return false
	}
	for _, msg := range msgs {
		if !isGovProp(msg) {
			return false
		}
	}
	return true
}

// isGovProp returns true if the provided message is a governance module MsgSubmitProposal.
func isGovProp(msg sdk.Msg) bool {
	t := sdk.MsgTypeURL(msg)
	// Needs to return true for "/cosmos.gov.v1.MsgSubmitProposal" and "/cosmos.gov.v1beta1.MsgSubmitProposal".
	// Since the types of messages are limited, there's only a limited set of possible msg-type URLs, so we're
	// okay with a bit looser of a test here that allows for new versions to be added later, and still work.
	return strings.HasPrefix(t, "/cosmos.gov.") && strings.HasSuffix(t, ".MsgSubmitProposal")
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

// GetFeeTx coverts the provided Tx to a FeeTx if possible.
func GetFeeTx(tx sdk.Tx) (sdk.FeeTx, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return nil, sdkerrors.ErrTxDecode.Wrapf("Tx must be a FeeTx: %T", tx)
	}
	return feeTx, nil
}

// IsInitGenesis returns true if the context indicates we're in InitGenesis.
func IsInitGenesis(ctx sdk.Context) bool {
	// Note: This isn't fully accurate since you can initialize a chain at a height other than zero.
	// But it should be good enough for our stuff. Ideally we'd want something specifically set in
	// the context during InitGenesis to check, but that'd probably involve some SDK work.
	return ctx.BlockHeight() <= 0
}
