package keeper_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/provenance-io/provenance/x/hold"
)

var _ hold.BankKeeper = (*MockBankKeeper)(nil)

type MockBankKeeper struct {
	// Locked is a map of sdk.AccAddress (cast to string) to
	// what the result of LockedCoins should be for it.
	Locked map[string]sdk.Coins
	// Balance is a map of sdk.AccAddress (cast to string) to
	// what the result of GetBalance should be for it.
	Balance map[string]sdk.Coins
}

func NewMockBankKeeper() *MockBankKeeper {
	return &MockBankKeeper{
		Locked:  make(map[string]sdk.Coins),
		Balance: make(map[string]sdk.Coins),
	}
}

func (k *MockBankKeeper) WithLocked(addr sdk.AccAddress, amount sdk.Coins) *MockBankKeeper {
	k.Locked[string(addr)] = amount
	return k
}

func (k *MockBankKeeper) WithBalance(addr sdk.AccAddress, amount sdk.Coins) *MockBankKeeper {
	k.Balance[string(addr)] = amount
	return k
}

func (k *MockBankKeeper) AppendLockedCoinsGetter(_ banktypes.GetLockedCoinsFn) {
	// Do nothing.
}

func (k *MockBankKeeper) LockedCoins(_ context.Context, addr sdk.AccAddress) sdk.Coins {
	return k.Locked[string(addr)]
}

func (k *MockBankKeeper) GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return sdk.Coin{Denom: denom, Amount: k.Balance[string(addr)].AmountOf(denom)}
}

type MockAccountKeeper struct {
	GetAccountResults map[string]sdk.AccountI
	GetAccountCalls   []string

	SetAccountCalls []sdk.AccountI
}

var _ hold.AccountKeeper = (*MockAccountKeeper)(nil)

func NewMockAccountKeeper() *MockAccountKeeper {
	return &MockAccountKeeper{GetAccountResults: make(map[string]sdk.AccountI)}
}

func (k *MockAccountKeeper) WithAccount(addr sdk.AccAddress, acct sdk.AccountI) *MockAccountKeeper {
	k.GetAccountResults[string(addr)] = acct
	return k
}

func (k *MockAccountKeeper) WithAccounts(t *testing.T, accts ...sdk.AccountI) *MockAccountKeeper {
	for _, acct := range accts {
		addr := acct.GetAddress()
		require.NotEmpty(t, addr, "%#v.GetAddress() result", acct)
		k.GetAccountResults[string(addr)] = acct
	}
	return k
}

func (k *MockAccountKeeper) Reset() {
	k.GetAccountCalls = nil
	k.SetAccountCalls = nil
}

func (k *MockAccountKeeper) GetAccount(_ context.Context, addr sdk.AccAddress) sdk.AccountI {
	k.GetAccountCalls = append(k.GetAccountCalls, addr.String())
	return k.GetAccountResults[string(addr)]
}

func (k *MockAccountKeeper) SetAccount(_ context.Context, acc sdk.AccountI) {
	k.SetAccountCalls = append(k.SetAccountCalls, acc)
}
