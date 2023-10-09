package keeper

import (
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

// ParseLengthPrefixedAddr is a test-only exposure of parseLengthPrefixedAddr.
var ParseLengthPrefixedAddr = parseLengthPrefixedAddr

// GetStore is a test-only exposure of getStore.
func (k Keeper) GetStore(ctx sdk.Context) sdk.KVStore {
	return k.getStore(ctx)
}

var (
	// DeleteAll is a test-only exposure of deleteAll.
	DeleteAll = deleteAll
	// Iterate is a test-only exposure of iterate.
	Iterate = iterate
)
