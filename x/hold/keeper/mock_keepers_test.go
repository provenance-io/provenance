package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/provenance-io/provenance/x/hold"
)

var _ hold.BankKeeper = (*MockBankKeeper)(nil)

type MockBankKeeper struct {
	// Spendable is a map of sdk.AccAddress (cast to string) to
	// what the result of SpendableCoins should be for it.
	Spendable map[string]sdk.Coins
}

func NewMockBankKeeper() *MockBankKeeper {
	return &MockBankKeeper{
		Spendable: make(map[string]sdk.Coins),
	}
}

func (k *MockBankKeeper) WithSpendable(addr sdk.AccAddress, amount sdk.Coins) *MockBankKeeper {
	k.Spendable[string(addr)] = amount
	return k
}

func (k *MockBankKeeper) AppendLockedCoinsGetter(_ banktypes.GetLockedCoinsFn) {
	// Do nothing.
}

func (k *MockBankKeeper) SpendableCoins(_ sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return k.Spendable[string(addr)]
}
