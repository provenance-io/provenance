package keeper

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/exchange"
)

// This file is in the keeper package (not keeper_test) so that it can expose
// some private keeper stuff for unit testing.

// WithAccountKeeper is a test-only method that returns a new Keeper that uses the provided AccountKeeper.
func (k Keeper) WithAccountKeeper(accountKeeper exchange.AccountKeeper) Keeper {
	k.accountKeeper = accountKeeper
	return k
}

// WithAttributeKeeper is a test-only method that returns a new Keeper that uses the provided AttributeKeeper.
func (k Keeper) WithAttributeKeeper(attrKeeper exchange.AttributeKeeper) Keeper {
	k.attrKeeper = attrKeeper
	return k
}

// WithBankKeeper is a test-only method that returns a new Keeper that uses the provided BankKeeper.
func (k Keeper) WithBankKeeper(bankKeeper exchange.BankKeeper) Keeper {
	k.bankKeeper = bankKeeper
	return k
}

// WithHoldKeeper is a test-only method that returns a new Keeper that uses the provided HoldKeeper.
func (k Keeper) WithHoldKeeper(holdKeeper exchange.HoldKeeper) Keeper {
	k.holdKeeper = holdKeeper
	return k
}

// WithMarkerKeeper is a test-only method that returns a new Keeper that uses the provided MarkerKeeper.
func (k Keeper) WithMarkerKeeper(markerKeeper exchange.MarkerKeeper) Keeper {
	k.markerKeeper = markerKeeper
	return k
}

// WithMetadataKeeper is a test-only method that returns a new Keeper that uses the provided MetadataKeeper.
func (k Keeper) WithMetadataKeeper(metadataKeeper exchange.MetadataKeeper) Keeper {
	k.metadataKeeper = metadataKeeper
	return k
}

// GetStore is a test-only exposure of getStore.
func (k Keeper) GetStore(ctx sdk.Context) storetypes.KVStore {
	return k.getStore(ctx)
}

// SetOrderInStore is a test-only exposure of setOrderInStore.
func (k Keeper) SetOrderInStore(store storetypes.KVStore, order exchange.Order) error {
	return k.setOrderInStore(store, order)
}

// GetOrderStoreKeyValue is a test-only exposure of getOrderStoreKeyValue.
func (k Keeper) GetOrderStoreKeyValue(order exchange.Order) ([]byte, []byte, error) {
	return k.getOrderStoreKeyValue(order)
}

// AddCommitmentsUnsafe is a test-only exposure of addCommitmentsUnsafe.
func (k Keeper) AddCommitmentsUnsafe(ctx sdk.Context, marketID uint32, toAdd []exchange.AccountAmount, eventTag string) error {
	return k.addCommitmentsUnsafe(ctx, marketID, toAdd, eventTag)
}

// SetPaymentInStore is a test-only exposure of setPaymentInStore.
func (k Keeper) SetPaymentInStore(store storetypes.KVStore, payment *exchange.Payment) error {
	return k.setPaymentInStore(store, payment)
}

// GetCodec is a test-only exposure of this keeper's cdc.
func (k Keeper) GetCodec() codec.BinaryCodec {
	return k.cdc
}

var (
	// DeleteAll is a test-only exposure of deleteAll.
	DeleteAll = deleteAll
	// Iterate is a test-only exposure of iterate.
	Iterate = iterate
	// ParseLengthPrefixedAddr is a test-only exposure of parseLengthPrefixedAddr.
	ParseLengthPrefixedAddr = parseLengthPrefixedAddr
	// Uint16Bz is a test-only exposure of uint16Bz.
	Uint16Bz = uint16Bz
	// Uint32Bz is a test-only exposure of uint32Bz.
	Uint32Bz = uint32Bz
	// Uint64Bz is a test-only exposure of uint64Bz.
	Uint64Bz = uint64Bz

	// SetParamsSplit is a test-only exposure of setParamsSplit.
	SetParamsSplit = setParamsSplit
	// SetParamsFeeCreatePaymentFlat is a test-only exposure of setParamsFeeCreatePaymentFlat.
	SetParamsFeeCreatePaymentFlat = setParamsFeeCreatePaymentFlat
	// SetParamsFeeAcceptPaymentFlat is a test-only exposure of setParamsFeeAcceptPaymentFlat.
	SetParamsFeeAcceptPaymentFlat = setParamsFeeAcceptPaymentFlat

	// GetLastAutoMarketID is a test-only exposure of getLastAutoMarketID.
	GetLastAutoMarketID = getLastAutoMarketID
	// SetLastAutoMarketID is a test-only exposure of setLastAutoMarketID.
	SetLastAutoMarketID = setLastAutoMarketID
	// SetMarketKnown is a test-only exposure of setMarketKnown.
	SetMarketKnown = setMarketKnown
	// SetCreateAskFlatFees is a test-only exposure of setCreateAskFlatFees.
	SetCreateAskFlatFees = setCreateAskFlatFees
	// SetCreateBidFlatFees is a test-only exposure of setCreateBidFlatFees.
	SetCreateBidFlatFees = setCreateBidFlatFees
	// SetCreateCommitmentFlatFees is a test-only exposure of setCreateCommitmentFlatFees.
	SetCreateCommitmentFlatFees = setCreateCommitmentFlatFees
	// SetSellerSettlementFlatFees is a test-only exposure of setSellerSettlementFlatFees.
	SetSellerSettlementFlatFees = setSellerSettlementFlatFees
	// SetBuyerSettlementFlatFees is a test-only exposure of setBuyerSettlementFlatFees.
	SetBuyerSettlementFlatFees = setBuyerSettlementFlatFees
	// SetSellerSettlementRatios is a test-only exposure of setSellerSettlementRatios.
	SetSellerSettlementRatios = setSellerSettlementRatios
	// SetBuyerSettlementRatios is a test-only exposure of setBuyerSettlementRatios.
	SetBuyerSettlementRatios = setBuyerSettlementRatios
	// SetCommitmentSettlementBips is a test-only exposure of setCommitmentSettlementBips.
	SetCommitmentSettlementBips = setCommitmentSettlementBips
	// SetIntermediaryDenom is a test-only exposure of setIntermediaryDenom.
	SetIntermediaryDenom = setIntermediaryDenom
	// SetMarketAcceptingOrders is a test-only exposure of setMarketAcceptingOrders.
	SetMarketAcceptingOrders = setMarketAcceptingOrders
	// SetUserSettlementAllowed is a test-only exposure of setUserSettlementAllowed.
	SetUserSettlementAllowed = setUserSettlementAllowed
	// SetMarketAcceptingCommitments is a test-only exposure of setMarketAcceptingCommitments.
	SetMarketAcceptingCommitments = setMarketAcceptingCommitments
	// GrantPermissions is a test-only exposure of grantPermissions.
	GrantPermissions = grantPermissions
	// SetReqAttrsAsk is a test-only exposure of setReqAttrsAsk.
	SetReqAttrsAsk = setReqAttrsAsk
	// SetReqAttrsBid is a test-only exposure of setReqAttrsBid.
	SetReqAttrsBid = setReqAttrsBid
	// SetReqAttrsCommitment is a test-only exposure of setReqAttrsCommitment.
	SetReqAttrsCommitment = setReqAttrsCommitment
	// StoreMarket is a test-only exposure of storeMarket.
	StoreMarket = storeMarket

	// GetLastOrderID is a test-only exposure of getLastOrderID.
	GetLastOrderID = getLastOrderID
	// SetLastOrderID is a test-only exposure of setLastOrderID.
	SetLastOrderID = setLastOrderID
	// CreateConstantIndexEntries is a test-only exposure of createConstantIndexEntries.
	CreateConstantIndexEntries = createConstantIndexEntries
	// CreateMarketExternalIDToOrderEntry is a test-only exposure of createMarketExternalIDToOrderEntry.
	CreateMarketExternalIDToOrderEntry = createMarketExternalIDToOrderEntry

	// SetCommitmentAmount is a test-only exposure of setCommitmentAmount.
	SetCommitmentAmount = setCommitmentAmount
)
