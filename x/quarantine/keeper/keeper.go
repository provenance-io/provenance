package keeper

import (
	"errors"
	"fmt"
	"strings"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/provenance-io/provenance/x/quarantine"
)

type Keeper struct {
	cdc codec.BinaryCodec

	bankKeeper quarantine.BankKeeper

	fundsHolder sdk.AccAddress
	schema      collections.Schema
	// optIns tracks addresses that have opted into quarantine.
	optIns collections.Map[sdk.AccAddress, bool]
	// autoResponses holds per-sender auto-accept / auto-decline settings.
	autoResponses collections.Map[collections.Pair[sdk.AccAddress, sdk.AccAddress], quarantine.AutoResponse]
	// records holds quarantined fund records.
	records collections.Map[collections.Pair[sdk.AccAddress, sdk.AccAddress], quarantine.QuarantineRecord]
	// recordIndex is a per-fromAddr → []recordSuffix index.
	recordIndex collections.Map[collections.Pair[sdk.AccAddress, sdk.AccAddress], quarantine.QuarantineRecordSuffixIndex]
}

func NewKeeper(cdc codec.BinaryCodec, storeService store.KVStoreService, bankKeeper quarantine.BankKeeper, fundsHolder sdk.AccAddress) Keeper {
	if len(fundsHolder) == 0 {
		fundsHolder = authtypes.NewModuleAddress(quarantine.ModuleName)
	}
	addrKey := quarantine.AddressLengthPrefixedKey
	pairCodec := collections.PairKeyCodec(addrKey, addrKey)
	sb := collections.NewSchemaBuilder(storeService)
	rv := Keeper{
		cdc: cdc,
		optIns: collections.NewMap(
			sb,
			quarantine.OptInKeyPrefix,
			"opt_ins",
			addrKey,
			quarantine.OptInValue,
		),
		autoResponses: collections.NewMap(
			sb,
			quarantine.AutoResponseKeyPrefix,
			"auto_responses",
			pairCodec,
			quarantine.AutoResponseValue,
		),
		records: collections.NewMap(
			sb,
			quarantine.RecordKeyPrefix,
			"records",
			pairCodec,
			codec.CollValue[quarantine.QuarantineRecord](cdc),
		),
		recordIndex: collections.NewMap(
			sb,
			quarantine.RecordIndexKeyPrefix,
			"record_index",
			pairCodec,
			codec.CollValue[quarantine.QuarantineRecordSuffixIndex](cdc),
		),
		bankKeeper:  bankKeeper,
		fundsHolder: fundsHolder,
	}
	schema, err := sb.Build()
	if err != nil {
		panic(fmt.Errorf("quarantine: failed to build collections schema: %w", err))
	}
	rv.schema = schema
	bankKeeper.AppendSendRestriction(rv.SendRestrictionFn)
	return rv
}

// GetSchema returns the collections schema for this keeper.
func (k Keeper) GetSchema() collections.Schema {
	return k.schema
}

// GetFundsHolder returns the account address that holds quarantined funds.
func (k Keeper) GetFundsHolder() sdk.AccAddress {
	return k.fundsHolder
}

// SetOptIn records that an address has opted into quarantine.
func (k Keeper) SetOptIn(ctx sdk.Context, toAddr sdk.AccAddress) error {
	if err := k.optIns.Set(ctx, toAddr, true); err != nil {
		return fmt.Errorf("quarantine: failed to set opt-in for %s: %w", toAddr, err)
	}
	return ctx.EventManager().EmitTypedEvent(&quarantine.EventOptIn{ToAddress: toAddr.String()})
}

// SetOptOut removes an address' quarantine opt-in record.
func (k Keeper) SetOptOut(ctx sdk.Context, toAddr sdk.AccAddress) error {
	if err := k.optIns.Remove(ctx, toAddr); err != nil {
		return fmt.Errorf("quarantine: failed to remove opt-in for %s: %w", toAddr, err)
	}
	return ctx.EventManager().EmitTypedEvent(&quarantine.EventOptOut{ToAddress: toAddr.String()})
}

// IsQuarantinedAddr returns true if the given address has opted into quarantine.
func (k Keeper) IsQuarantinedAddr(ctx sdk.Context, toAddr sdk.AccAddress) bool {
	has, err := k.optIns.Has(ctx, toAddr)
	if err != nil {
		panic(fmt.Errorf("quarantine: failed to check opt-in for %s: %w", toAddr, err))
	}
	return has
}

// IterateQuarantinedAccounts iterates over all quarantine account addresses.
// The callback function should accept the to address (that has quarantine enabled).
// It should return whether to stop iteration early. I.e. false will allow iteration to continue, true will stop iteration.
func (k Keeper) IterateQuarantinedAccounts(ctx sdk.Context, cb func(toAddr sdk.AccAddress) (stop bool)) {
	err := k.optIns.Walk(ctx, nil, func(toAddr sdk.AccAddress, _ bool) (stop bool, err error) {
		return cb(toAddr), nil
	})
	if err != nil {
		panic(fmt.Errorf("quarantine: failed to iterate quarantined accounts: %w", err))
	}
}

// SetAutoResponse sets the auto response of sends to toAddr from fromAddr.
// If the response is AUTO_RESPONSE_UNSPECIFIED, the auto-response record is deleted,
// otherwise it is created/updated with the given setting.
func (k Keeper) SetAutoResponse(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress, response quarantine.AutoResponse) {
	key := collections.Join(toAddr, fromAddr)
	if quarantine.ToAutoB(response) == quarantine.NoAutoB {
		if err := k.autoResponses.Remove(ctx, key); err != nil {
			panic(fmt.Errorf("quarantine: failed to remove auto-response for %s from %s: %w", toAddr, fromAddr, err))
		}
	} else {
		if err := k.autoResponses.Set(ctx, key, response); err != nil {
			panic(fmt.Errorf("quarantine: failed to set auto-response for %s from %s: %w", toAddr, fromAddr, err))
		}
	}
}

// GetAutoResponse returns the quarantine auto-response for the given to/from addresses.
func (k Keeper) GetAutoResponse(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress) quarantine.AutoResponse {
	if toAddr.Equals(fromAddr) {
		return quarantine.AUTO_RESPONSE_ACCEPT
	}
	resp, err := k.autoResponses.Get(ctx, collections.Join(toAddr, fromAddr))
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return quarantine.AUTO_RESPONSE_UNSPECIFIED
		}
		panic(fmt.Errorf("quarantine: failed to get auto-response for %s from %s: %w", toAddr, fromAddr, err))
	}
	return resp
}

// IsAutoAccept returns true if the to address has enabled auto-accept for ALL the from address.
func (k Keeper) IsAutoAccept(ctx sdk.Context, toAddr sdk.AccAddress, fromAddrs ...sdk.AccAddress) bool {
	for _, fromAddr := range fromAddrs {
		if !k.GetAutoResponse(ctx, toAddr, fromAddr).IsAccept() {
			return false
		}
	}
	return true
}

// IsAutoDecline returns true if the to address has enabled auto-decline for ANY of the from address.
func (k Keeper) IsAutoDecline(ctx sdk.Context, toAddr sdk.AccAddress, fromAddrs ...sdk.AccAddress) bool {
	for _, fromAddr := range fromAddrs {
		if k.GetAutoResponse(ctx, toAddr, fromAddr).IsDecline() {
			return true
		}
	}
	return false
}

// IterateAutoResponses iterates over the auto-responses for a given recipient address,
// or if no address is provided, iterates over all auto-response entries.
// The callback function should accept a to address, from address, and auto-response setting (in that order).
// It should return whether to stop iteration early. I.e. false will allow iteration to continue, true will stop iteration.
func (k Keeper) IterateAutoResponses(ctx sdk.Context, toAddr sdk.AccAddress, cb func(toAddr, fromAddr sdk.AccAddress, response quarantine.AutoResponse) (stop bool)) {
	var ranger collections.Ranger[collections.Pair[sdk.AccAddress, sdk.AccAddress]]
	if len(toAddr) > 0 {
		ranger = collections.NewPrefixedPairRange[sdk.AccAddress, sdk.AccAddress](toAddr)
	}
	err := k.autoResponses.Walk(ctx, ranger, func(
		key collections.Pair[sdk.AccAddress, sdk.AccAddress],
		response quarantine.AutoResponse,
	) (stop bool, err error) {
		return cb(key.K1(), key.K2(), response), nil
	})
	if err != nil {
		panic(fmt.Errorf("quarantine: failed to iterate auto-responses: %w", err))
	}
}

// SetQuarantineRecord sets a quarantine record.
// Panics if the record is nil.
// If the record is fully accepted, it is deleted.
// Otherwise, it is saved.
func (k Keeper) SetQuarantineRecord(ctx sdk.Context, toAddr sdk.AccAddress, record *quarantine.QuarantineRecord) {
	if record == nil {
		panic("record cannot be nil")
	}
	fromAddrs := record.GetAllFromAddrs()
	suffix := quarantine.CreateRecordSuffix(fromAddrs)
	key := collections.Join(toAddr, sdk.AccAddress(suffix))
	if record.IsFullyAccepted() {
		if err := k.records.Remove(ctx, key); err != nil {
			panic(fmt.Errorf("quarantine: failed to remove record for %s: %w", toAddr, err))
		}
		// Only multi-sender records have suffix-index entries to clean up.
		if len(fromAddrs) > 1 {
			k.deleteQuarantineRecordSuffixIndexes(ctx, toAddr, fromAddrs, suffix)
		}
	} else {
		if err := k.records.Set(ctx, key, *record); err != nil {
			panic(fmt.Errorf("quarantine: failed to set record for %s: %w", toAddr, err))
		}
		// Only multi-sender records require suffix-index maintenance.
		if len(fromAddrs) > 1 {
			k.addQuarantineRecordSuffixIndexes(ctx, toAddr, fromAddrs, suffix)
		}
	}
}

// bzToQuarantineRecord converts the given byte slice into a QuarantineRecord or returns an error.
// If the byte slice is nil or empty, a default QuarantineRecord is returned with zero coins.
//
//nolint:unused // used in tests
func (k Keeper) bzToQuarantineRecord(bz []byte) (*quarantine.QuarantineRecord, error) {
	qf := quarantine.QuarantineRecord{
		Coins: sdk.Coins{},
	}
	if len(bz) > 0 {
		err := k.cdc.Unmarshal(bz, &qf)
		if err != nil {
			return &qf, err
		}
	}
	return &qf, nil
}

// mustBzToQuarantineRecord returns bzToQuarantineRecord but panics on error.
//
//nolint:unused // used in tests
func (k Keeper) mustBzToQuarantineRecord(bz []byte) *quarantine.QuarantineRecord {
	qf, err := k.bzToQuarantineRecord(bz)
	if err != nil {
		panic(err)
	}
	return qf
}

// GetQuarantineRecord gets the single quarantine record to toAddr from all the fromAddrs.
// If the record doesn't exist, nil is returned.
//
// If you want all records from any of the fromAddrs, use GetQuarantineRecords.
func (k Keeper) GetQuarantineRecord(ctx sdk.Context, toAddr sdk.AccAddress, fromAddrs ...sdk.AccAddress) *quarantine.QuarantineRecord {
	suffix := quarantine.CreateRecordSuffix(fromAddrs)
	record, err := k.records.Get(ctx, collections.Join(toAddr, sdk.AccAddress(suffix)))
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil
		}
		panic(fmt.Errorf("quarantine: failed to get record for %s: %w", toAddr, err))
	}
	return &record
}

// GetQuarantineRecords gets all the quarantine records to toAddr that involved any of the fromAddrs.
//
// If you want a single record from all the fromAddrs, use GetQuarantineRecord.
func (k Keeper) GetQuarantineRecords(ctx sdk.Context, toAddr sdk.AccAddress, fromAddrs ...sdk.AccAddress) []*quarantine.QuarantineRecord {
	allSuffixes := k.getQuarantineRecordSuffixes(ctx, toAddr, fromAddrs)
	var rv []*quarantine.QuarantineRecord //nolint:prealloc // dynamic allocation is acceptable here.
	for _, suffix := range allSuffixes {
		record, err := k.records.Get(ctx, collections.Join(toAddr, sdk.AccAddress(suffix)))
		if err != nil {
			if errors.Is(err, collections.ErrNotFound) {
				continue
			}
			panic(fmt.Errorf("quarantine: failed to get record for suffix: %w", err))
		}
		rv = append(rv, &record)
	}
	return rv
}

// AddQuarantinedCoins records that some new funds have been quarantined.
func (k Keeper) AddQuarantinedCoins(ctx sdk.Context, coins sdk.Coins, toAddr sdk.AccAddress, fromAddrs ...sdk.AccAddress) error {
	qr := k.GetQuarantineRecord(ctx, toAddr, fromAddrs...)
	if qr != nil {
		qr.AddCoins(coins...)
	} else {
		qr = &quarantine.QuarantineRecord{
			Coins: coins,
		}
		for _, fromAddr := range fromAddrs {
			if k.IsAutoAccept(ctx, toAddr, fromAddr) {
				qr.AcceptedFromAddresses = append(qr.AcceptedFromAddresses, fromAddr)
			} else {
				qr.UnacceptedFromAddresses = append(qr.UnacceptedFromAddresses, fromAddr)
			}
		}
	}
	if qr.IsFullyAccepted() {
		fromAddrStrs := make([]string, len(fromAddrs))
		for i, addr := range fromAddrs {
			fromAddrStrs[i] = addr.String()
		}
		return fmt.Errorf("cannot add quarantined funds %q to %s from %s: already fully accepted",
			coins.String(), toAddr.String(), strings.Join(fromAddrStrs, ", "))
	}
	// Regardless of if its new or existing, set declined based on current auto-decline info.
	qr.Declined = k.IsAutoDecline(ctx, toAddr, fromAddrs...)
	k.SetQuarantineRecord(ctx, toAddr, qr)
	return ctx.EventManager().EmitTypedEvent(&quarantine.EventFundsQuarantined{
		ToAddress: toAddr.String(),
		Coins:     coins,
	})
}

// AcceptQuarantinedFunds looks up all quarantined funds to toAddr from any of the fromAddrs.
// It marks and saves each as accepted and, if fully accepted, releases (sends) the funds to toAddr.
// Returns total funds released and possibly an error.
func (k Keeper) AcceptQuarantinedFunds(ctx sdk.Context, toAddr sdk.AccAddress, fromAddrs ...sdk.AccAddress) (sdk.Coins, error) {
	fundsReleased := sdk.Coins{}
	for _, record := range k.GetQuarantineRecords(ctx, toAddr, fromAddrs...) {
		if record.AcceptFrom(fromAddrs) {
			if record.IsFullyAccepted() {
				err := k.bankKeeper.SendCoins(quarantine.WithBypass(ctx), k.fundsHolder, toAddr, record.Coins)
				if err != nil {
					return nil, err
				}
				fundsReleased = fundsReleased.Add(record.Coins...)

				err = ctx.EventManager().EmitTypedEvent(&quarantine.EventFundsReleased{
					ToAddress: toAddr.String(),
					Coins:     record.Coins,
				})
				if err != nil {
					return nil, err
				}
			} else {
				// update declined to false unless one of the unaccepted from addresses is set to auto-decline.
				record.Declined = k.IsAutoDecline(ctx, toAddr, record.UnacceptedFromAddresses...)
			}
			k.SetQuarantineRecord(ctx, toAddr, record)
		}
	}

	return fundsReleased, nil
}

// DeclineQuarantinedFunds marks as declined, all quarantined funds to toAddr where any fromAddr is a sender.
func (k Keeper) DeclineQuarantinedFunds(ctx sdk.Context, toAddr sdk.AccAddress, fromAddrs ...sdk.AccAddress) {
	for _, record := range k.GetQuarantineRecords(ctx, toAddr, fromAddrs...) {
		if record.DeclineFrom(fromAddrs) {
			k.SetQuarantineRecord(ctx, toAddr, record)
		}
	}
}

// IterateQuarantineRecords iterates over the quarantine records for a given recipient address,
// or if no address is provided, iterates over all quarantine records.
// The callback function should accept a to address, record suffix, and QuarantineRecord (in that order).
// It should return whether to stop iteration early. I.e. false will allow iteration to continue, true will stop iteration.
func (k Keeper) IterateQuarantineRecords(ctx sdk.Context, toAddr sdk.AccAddress, cb func(toAddr, recordSuffix sdk.AccAddress, record *quarantine.QuarantineRecord) (stop bool)) {
	var ranger collections.Ranger[collections.Pair[sdk.AccAddress, sdk.AccAddress]]
	if len(toAddr) > 0 {
		ranger = collections.NewPrefixedPairRange[sdk.AccAddress, sdk.AccAddress](toAddr)
	}
	err := k.records.Walk(ctx, ranger, func(
		key collections.Pair[sdk.AccAddress, sdk.AccAddress],
		record quarantine.QuarantineRecord,
	) (stop bool, err error) {
		return cb(key.K1(), key.K2(), &record), nil
	})
	if err != nil {
		panic(fmt.Errorf("quarantine: failed to iterate quarantine records: %w", err))
	}
}

// setQuarantineRecordSuffixIndex writes the provided suffix index.
// If it is nil or there are no record suffixes, the entry is instead deleted.
func (k Keeper) setQuarantineRecordSuffixIndex(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress, value *quarantine.QuarantineRecordSuffixIndex) {
	key := collections.Join(toAddr, fromAddr)
	if value == nil || len(value.RecordSuffixes) == 0 {
		if err := k.recordIndex.Remove(ctx, key); err != nil {
			panic(fmt.Errorf("quarantine: failed to remove record suffix index for %s from %s: %w", toAddr, fromAddr, err))
		}
	} else {
		if err := k.recordIndex.Set(ctx, key, *value); err != nil {
			panic(fmt.Errorf("quarantine: failed to set record suffix index for %s from %s: %w", toAddr, fromAddr, err))
		}
	}
}

// bzToQuarantineRecordSuffixIndex converts the given byte slice into a QuarantineRecordSuffixIndex or returns an error.
// If the byte slice is nil or empty, a default QuarantineRecordSuffixIndex is returned with no suffixes.
//
//nolint:unused // used in tests
func (k Keeper) bzToQuarantineRecordSuffixIndex(bz []byte) (*quarantine.QuarantineRecordSuffixIndex, error) {
	var si quarantine.QuarantineRecordSuffixIndex
	if len(bz) > 0 {
		err := k.cdc.Unmarshal(bz, &si)
		if err != nil {
			return &si, err
		}
	}
	return &si, nil
}

// mustBzToQuarantineRecordSuffixIndex returns bzToQuarantineRecordSuffixIndex but panics on error.
//
//nolint:unused // used in tests
func (k Keeper) mustBzToQuarantineRecordSuffixIndex(bz []byte) *quarantine.QuarantineRecordSuffixIndex {
	si, err := k.bzToQuarantineRecordSuffixIndex(bz)
	if err != nil {
		panic(err)
	}
	return si
}

// getQuarantineRecordSuffixIndex gets a quarantine record suffix entry and it's key.
func (k Keeper) getQuarantineRecordSuffixIndex(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress) *quarantine.QuarantineRecordSuffixIndex {
	si, err := k.recordIndex.Get(ctx, collections.Join(toAddr, fromAddr))
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return &quarantine.QuarantineRecordSuffixIndex{}
		}
		panic(fmt.Errorf("quarantine: failed to get record suffix index for %s from %s: %w", toAddr, fromAddr, err))
	}
	return &si
}

// getQuarantineRecordSuffixes gets a sorted list of known record suffixes of quarantine records to toAddr
// from any of the fromAddrs. The list will not contain duplicates, but may contain suffixes that don't point to records.
func (k Keeper) getQuarantineRecordSuffixes(ctx sdk.Context, toAddr sdk.AccAddress, fromAddrs []sdk.AccAddress) [][]byte {
	rv := &quarantine.QuarantineRecordSuffixIndex{}
	for _, fromAddr := range fromAddrs {
		suffixes := k.getQuarantineRecordSuffixIndex(ctx, toAddr, fromAddr)
		rv.AddSuffixes(suffixes.RecordSuffixes...)
		rv.AddSuffixes(fromAddr)
	}
	rv.Simplify()
	return rv.RecordSuffixes
}

// addQuarantineRecordSuffixIndexes adds the provided suffix to all to/from suffix index entries.
func (k Keeper) addQuarantineRecordSuffixIndexes(ctx sdk.Context, toAddr sdk.AccAddress, fromAddrs []sdk.AccAddress, suffix []byte) {
	for _, fromAddr := range fromAddrs {
		ind := k.getQuarantineRecordSuffixIndex(ctx, toAddr, fromAddr)
		ind.AddSuffixes(suffix)
		ind.Simplify(fromAddr)
		k.setQuarantineRecordSuffixIndex(ctx, toAddr, fromAddr, ind)
	}
}

// deleteQuarantineRecordSuffixIndexes removes the provided suffix from all to/from suffix index entries and either saves
// the updated list or deletes it if it's now empty.
func (k Keeper) deleteQuarantineRecordSuffixIndexes(ctx sdk.Context, toAddr sdk.AccAddress, fromAddrs []sdk.AccAddress, suffix []byte) {
	for _, fromAddr := range fromAddrs {
		ind := k.getQuarantineRecordSuffixIndex(ctx, toAddr, fromAddr)
		ind.Simplify(fromAddr, suffix)
		k.setQuarantineRecordSuffixIndex(ctx, toAddr, fromAddr, ind)
	}
}
