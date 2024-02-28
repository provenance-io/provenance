package keeper

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/quarantine"

	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/x/exchange"
)

// paymentExists returns true if there's a payment in the store with the given source and external id.
func paymentExists(store sdk.KVStore, source, externalID string) bool {
	sourceAddr, err := sdk.AccAddressFromBech32(source)
	if err != nil {
		return false
	}
	return store.Has(MakeKeyPayment(sourceAddr, externalID))
}

// parsePaymentStoreValue converts a payment store value into the Payment object.
func (k Keeper) parsePaymentStoreValue(value []byte) (*exchange.Payment, error) {
	if len(value) == 0 {
		return nil, nil
	}

	var payment exchange.Payment
	err := k.cdc.Unmarshal(value, &payment)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal payment: %w", err)
	}
	return &payment, nil
}

// getPaymentFromStore gets a Payment from the store.
func (k Keeper) getPaymentFromStore(store sdk.KVStore, source sdk.AccAddress, externalID string) (*exchange.Payment, error) {
	key := MakeKeyPayment(source, externalID)
	value := store.Get(key)
	return k.parsePaymentStoreValue(value)
}

// getPaymentsForTargetAndSourceFromStore gets all the payments with the given target and source from the state store.
func (k Keeper) getPaymentsForTargetAndSourceFromStore(store sdk.KVStore, target, source sdk.AccAddress) []*exchange.Payment {
	var rv []*exchange.Payment
	keyPrefix := GetIndexKeyPrefixTargetToPaymentsForSource(target, source)
	iterate(store, keyPrefix, func(keySuffix, _ []byte) bool {
		externalID := string(keySuffix)
		payment, err := k.getPaymentFromStore(store, source, externalID)
		if err == nil && payment != nil {
			rv = append(rv, payment)
		}
		return false
	})
	return rv
}

// requirePaymentFromStore is like getPaymentFromStore but returns with an error if the payment does not exist.
func (k Keeper) requirePaymentFromStore(store sdk.KVStore, source sdk.AccAddress, externalID string) (*exchange.Payment, error) {
	payment, err := k.getPaymentFromStore(store, source, externalID)
	if err != nil {
		return nil, fmt.Errorf("error getting existing payment with source %s and external id %q: %w",
			source, externalID, err)
	}
	if payment == nil {
		return nil, fmt.Errorf("no payment found with source %s and external id %q", source, externalID)
	}
	return payment, nil
}

// setPaymentInStore sets a payment in the store making sure the index entry stays up to date.
func (k Keeper) setPaymentInStore(store sdk.KVStore, payment *exchange.Payment) error {
	source, err := sdk.AccAddressFromBech32(payment.Source)
	if err != nil {
		return fmt.Errorf("invalid source %q: %w", payment.Source, err)
	}
	pKey := MakeKeyPayment(source, payment.ExternalId)
	pVal, err := k.cdc.Marshal(payment)
	if err != nil {
		return fmt.Errorf("error marshaling payment: %w", err)
	}

	var target sdk.AccAddress
	var iKey []byte
	if len(payment.Target) > 0 {
		target, err = sdk.AccAddressFromBech32(payment.Target)
		if err != nil {
			return fmt.Errorf("invalid target %q: %w", payment.Target, err)
		}
		iKey = MakeIndexKeyTargetToPayment(target, source, payment.ExternalId)
	}

	var oldIKey []byte
	if existing, _ := k.getPaymentFromStore(store, source, payment.ExternalId); existing != nil {
		switch existing.Target {
		case "":
			// There isn't an entry yet, so there's nothing to delete.
		case payment.Target:
			// The existing entry has the same target. No need to delete it and rewrite the same index entry.
			iKey = nil
		default:
			// The target's changing, delete the old index entry.
			var oldTarget sdk.AccAddress
			oldTarget, err = sdk.AccAddressFromBech32(existing.Target)
			if err == nil && len(oldTarget) > 0 {
				oldIKey = MakeIndexKeyTargetToPayment(oldTarget, source, payment.ExternalId)
			}
		}
	}

	store.Set(pKey, pVal)
	if len(oldIKey) > 0 {
		store.Delete(oldIKey)
	}
	if len(iKey) > 0 {
		store.Set(iKey, []byte{})
	}

	return nil
}

// createPaymentInStore verifies that the provided payment does not yet exist, then writes it to the state store.
func (k Keeper) createPaymentInStore(store sdk.KVStore, payment *exchange.Payment) error {
	if paymentExists(store, payment.Source, payment.ExternalId) {
		return fmt.Errorf("a payment already exists with source %s and external id %q",
			payment.Source, payment.ExternalId)
	}
	return k.setPaymentInStore(store, payment)
}

// deletePaymentFromStore deletes a payment (and its index) from the state store.
func deletePaymentFromStore(store sdk.KVStore, payment *exchange.Payment) error {
	if payment == nil {
		return errors.New("cannot delete nil payment")
	}

	source, err := sdk.AccAddressFromBech32(payment.Source)
	if err != nil {
		return fmt.Errorf("invalid source %q: %w", payment.Source, err)
	}
	pKey := MakeKeyPayment(source, payment.ExternalId)

	var target sdk.AccAddress
	var iKey []byte
	if len(payment.Target) > 0 {
		target, err = sdk.AccAddressFromBech32(payment.Target)
		if err != nil {
			return fmt.Errorf("invalid target %q: %w", payment.Target, err)
		}
		iKey = MakeIndexKeyTargetToPayment(target, source, payment.ExternalId)
	}

	store.Delete(pKey)
	if len(iKey) > 0 {
		store.Delete(iKey)
	}

	return nil
}

// deletePaymentAndReleaseHold deletes a payment from the state store and releases its hold.
func (k Keeper) deletePaymentAndReleaseHold(ctx sdk.Context, store sdk.KVStore, payment *exchange.Payment) error {
	err := deletePaymentFromStore(store, payment)
	if err != nil {
		return err
	}

	source, _ := sdk.AccAddressFromBech32(payment.Source)
	err = k.holdKeeper.ReleaseHold(ctx, source, payment.SourceAmount)
	if err != nil {
		return fmt.Errorf("error releasing hold on payment source: %w", err)
	}

	return nil
}

func (k Keeper) deletePaymentsAndReleaseHolds(ctx sdk.Context, store sdk.KVStore, payments []*exchange.Payment) error {
	for _, payment := range payments {
		err := k.deletePaymentAndReleaseHold(ctx, store, payment)
		if err != nil {
			return fmt.Errorf("failed to delete payment with source %s and external id %q: %w",
				payment.Source, payment.ExternalId, err)
		}
	}

	return nil
}

// GetPayment gets a payment from the state store. If it doesn't exist, nil, nil is returned.
func (k Keeper) GetPayment(ctx sdk.Context, source sdk.AccAddress, externalID string) (*exchange.Payment, error) {
	return k.getPaymentFromStore(k.getStore(ctx), source, externalID)
}

// CreatePayment stores the provided payment in the state store and places a hold on the source funds.
func (k Keeper) CreatePayment(ctx sdk.Context, payment *exchange.Payment) error {
	if payment == nil {
		return errors.New("cannot create nil payment")
	}
	if err := payment.Validate(); err != nil {
		return fmt.Errorf("cannot create invalid payment: %w", err)
	}

	err := k.createPaymentInStore(k.getStore(ctx), payment)
	if err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}

	source, _ := sdk.AccAddressFromBech32(payment.Source)
	err = k.holdKeeper.AddHold(ctx, source, payment.SourceAmount, fmt.Sprintf("x/exchange: payment %q", payment.ExternalId))
	if err != nil {
		return fmt.Errorf("error placing hold on payment source: %w", err)
	}

	return nil
}

// AcceptPayment verifies that all the payment data matches what's in state, then deletes the payment and
// sends the source funds to the target and target funds to the source.
func (k Keeper) AcceptPayment(ctx sdk.Context, payment *exchange.Payment) error {
	if payment == nil {
		return errors.New("cannot accept nil payment")
	}
	if err := payment.Validate(); err != nil {
		return fmt.Errorf("cannot accept invalid payment: %w", err)
	}
	if len(payment.Target) == 0 {
		return errors.New("cannot accept a payment without a target")
	}

	store := k.getStore(ctx)
	source, _ := sdk.AccAddressFromBech32(payment.Source)
	target, _ := sdk.AccAddressFromBech32(payment.Target)
	existing, err := k.requirePaymentFromStore(store, source, payment.ExternalId)
	if err != nil {
		return err
	}

	if payment.Source != existing.Source {
		return fmt.Errorf("provided source %q does not equal existing source %q",
			payment.Source, existing.Source)
	}
	if !exchange.CoinsEquals(payment.SourceAmount, existing.SourceAmount) {
		return fmt.Errorf("provided source amount %q does not equal existing source amount %q",
			payment.SourceAmount, existing.SourceAmount)
	}
	if payment.Target != existing.Target {
		return fmt.Errorf("provided target %q does not equal existing target %q",
			payment.Target, existing.Target)
	}
	if !exchange.CoinsEquals(payment.TargetAmount, existing.TargetAmount) {
		return fmt.Errorf("provided target amount %q does not equal existing target amount %q",
			payment.TargetAmount, existing.TargetAmount)
	}
	if payment.ExternalId != existing.ExternalId {
		return fmt.Errorf("provided external id %q does not equal existing external id %q",
			payment.Source, existing.Source)
	}

	err = deletePaymentFromStore(store, existing)
	if err != nil {
		return fmt.Errorf("error removing payment from store: %w", err)
	}

	ctx = quarantine.WithBypass(ctx)
	if !payment.SourceAmount.IsZero() {
		err = k.bankKeeper.SendCoins(ctx, source, target, payment.SourceAmount)
		if err != nil {
			return fmt.Errorf("error sending %q from source %s to target %s: %w",
				payment.SourceAmount, source, target, err)
		}
	}
	if !payment.TargetAmount.IsZero() {
		err = k.bankKeeper.SendCoins(ctx, target, source, payment.TargetAmount)
		if err != nil {
			return fmt.Errorf("error sending %q from target %s to source %s: %w",
				payment.TargetAmount, target, source, err)
		}
	}

	return nil
}

// RejectPayment deletes a payment and releases the hold on it.
// An error is returned if a payment can't be found for the source + external id,
// or if that payment has a different target than the one provided.
func (k Keeper) RejectPayment(ctx sdk.Context, target, source sdk.AccAddress, externalID string) error {
	store := k.getStore(ctx)
	payment, err := k.requirePaymentFromStore(store, source, externalID)
	if err != nil {
		return err
	}

	if payment.Target != target.String() {
		return fmt.Errorf("target %s cannot reject payment with target %s", target, payment.Target)
	}

	return k.deletePaymentAndReleaseHold(ctx, store, payment)
}

// RejectPayments deletes some payments and releases their holds.
// Each source must have at least one payment for the target.
func (k Keeper) RejectPayments(ctx sdk.Context, target sdk.AccAddress, sources []sdk.AccAddress) error {
	var payments []*exchange.Payment
	store := k.getStore(ctx)
	for _, source := range sources {
		sPayments := k.getPaymentsForTargetAndSourceFromStore(store, target, source)
		if len(sPayments) == 0 {
			return fmt.Errorf("source %s does not have any payments for target %s", source, target)
		}
		payments = append(payments, sPayments...)
	}

	if len(payments) == 0 {
		return errors.New("no payments found to reject")
	}

	return k.deletePaymentsAndReleaseHolds(ctx, store, payments)
}

// CancelPayments deletes the payments (and releases their holds) for a source and set of external ids.
// There must be at least one external id and there must be a payment for each external id (and source).
func (k Keeper) CancelPayments(ctx sdk.Context, source sdk.AccAddress, externalIDs []string) error {
	if len(externalIDs) == 0 {
		return errors.New("at least one external id is required")
	}

	store := k.getStore(ctx)
	payments := make([]*exchange.Payment, 0, len(externalIDs))
	for _, externalID := range externalIDs {
		payment, err := k.requirePaymentFromStore(store, source, externalID)
		if err != nil {
			return err
		}
		payments = append(payments, payment)
	}

	return k.deletePaymentsAndReleaseHolds(ctx, store, payments)
}

// UpdatePaymentTarget changes the target of a payment.
func (k Keeper) UpdatePaymentTarget(ctx sdk.Context, source sdk.AccAddress, externalID string, newTarget sdk.AccAddress) error {
	store := k.getStore(ctx)
	existing, err := k.getPaymentFromStore(store, source, externalID)
	if err != nil {
		return fmt.Errorf("error getting existing payment with source %s and external id %q: %w",
			source, externalID, err)
	}
	if existing == nil {
		return fmt.Errorf("no payment found with source %s and external id %q", source, externalID)
	}

	newTargetStr := newTarget.String()
	if existing.Target == newTargetStr {
		return fmt.Errorf("payment with source %s and external id %q already has target %s",
			source, externalID, newTarget)
	}

	existing.Target = newTargetStr
	return k.setPaymentInStore(store, existing)
}

// GetPaymentsForTargetAndSource gets all the payments with the given target and source.
func (k Keeper) GetPaymentsForTargetAndSource(ctx sdk.Context, target, source sdk.AccAddress) []*exchange.Payment {
	return k.getPaymentsForTargetAndSourceFromStore(k.getStore(ctx), target, source)
}

// IteratePayments iterates over all payments.
// The callback takes in the payment and should return whether to stop iterating.
func (k Keeper) IteratePayments(ctx sdk.Context, cb func(payment *exchange.Payment) bool) {
	k.iterate(ctx, GetKeyPrefixAllPayments(), func(_, value []byte) bool {
		payment, err := k.parsePaymentStoreValue(value)
		if err != nil || payment == nil {
			return false
		}
		return cb(payment)
	})
}

// CalculatePaymentFees calculates the fees required for the provided payment.
func (k Keeper) CalculatePaymentFees(ctx sdk.Context, payment *exchange.Payment) *exchange.QueryPaymentFeeCalcResponse {
	resp := &exchange.QueryPaymentFeeCalcResponse{}
	if payment == nil {
		return resp
	}

	store := k.getStore(ctx)
	if !payment.SourceAmount.IsZero() {
		opts := getParamsFeeCreatePaymentFlat(store)
		if len(opts) > 0 {
			resp.FeeCreate = sdk.Coins{opts[0]}
		}
	}
	if !payment.TargetAmount.IsZero() {
		opts := getParamsFeeAcceptPaymentFlat(store)
		if len(opts) > 0 {
			resp.FeeAccept = sdk.Coins{opts[0]}
		}
	}

	return resp
}

// consumePaymentFee consumes the first entry in opts (if there is one) as a msg fee.
func consumePaymentFee(ctx sdk.Context, opts []sdk.Coin, msg sdk.Msg) {
	if len(opts) == 0 || opts[0].IsZero() {
		return
	}
	antewrapper.ConsumeMsgFee(ctx, sdk.Coins{opts[0]}, msg, "")
}

// consumeCreatePaymentFee looks up and consumes the create-payment fee.
func (k Keeper) consumeCreatePaymentFee(ctx sdk.Context, msg sdk.Msg) {
	consumePaymentFee(ctx, getParamsFeeCreatePaymentFlat(k.getStore(ctx)), msg)
}

// consumeAcceptPaymentFee looks up and consumes the accept-payment fee.
func (k Keeper) consumeAcceptPaymentFee(ctx sdk.Context, msg sdk.Msg) {
	consumePaymentFee(ctx, getParamsFeeAcceptPaymentFlat(k.getStore(ctx)), msg)
}
