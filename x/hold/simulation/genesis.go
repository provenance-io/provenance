package simulation

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	hold "github.com/provenance-io/provenance/x/hold"
)

const HoldAccountHolds = "hold-account-holds"

// RandomAccountHolds randomly selects accounts and hold amounts for the selected ones.
func RandomAccountHolds(r *rand.Rand, accounts []simtypes.Account) []*hold.AccountHold {
	if len(accounts) == 0 {
		return nil
	}

	count := r.Intn(len(accounts) + 1)
	if count == 0 {
		return nil
	}

	addrs := make([]sdk.AccAddress, len(accounts))
	for i, acc := range accounts {
		addrs[i] = acc.Address
	}
	r.Shuffle(len(addrs), func(i, j int) {
		addrs[i], addrs[j] = addrs[j], addrs[i]
	})

	rv := make([]*hold.AccountHold, count)
	for i, addr := range addrs[:count] {
		rv[i] = &hold.AccountHold{
			Address: addr.String(),
			Amount:  sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, r.Int63n(1000)+1)),
		}
	}

	return rv
}

// RandomizedGenState generates a random GenesisState for the hold module.
func RandomizedGenState(simState *module.SimulationState) {
	genState := &hold.GenesisState{}

	simState.AppParams.GetOrGenerate(
		simState.Cdc, HoldAccountHolds, &genState.Holds, simState.Rand,
		func(r *rand.Rand) {
			genState.Holds = RandomAccountHolds(r, simState.Accounts)
		},
	)

	simState.GenState[hold.ModuleName] = simState.Cdc.MustMarshalJSON(genState)

	// If we put stuff in hold, add those funds to the bank accounts.
	if len(genState.Holds) > 0 {
		bankGenRaw := simState.GenState[banktypes.ModuleName]
		bankGen := banktypes.GenesisState{}
		simState.Cdc.MustUnmarshalJSON(bankGenRaw, &bankGen)

		var totalAdded sdk.Coins
		var newBalances []banktypes.Balance
		for _, ah := range genState.Holds {
			haveBal := false
			for i, bal := range bankGen.Balances {
				if bal.Address == ah.Address {
					bankGen.Balances[i].Coins = bal.Coins.Add(ah.Amount...)
					totalAdded = totalAdded.Add(ah.Amount...)
					haveBal = true
					break
				}
			}
			if !haveBal {
				newBalances = append(newBalances, banktypes.Balance{
					Address: ah.Address,
					Coins:   ah.Amount,
				})
				totalAdded = totalAdded.Add(ah.Amount...)
			}
		}
		bankGen.Balances = append(bankGen.Balances, newBalances...)
		bankGen.Supply = bankGen.Supply.Add(totalAdded...)

		simState.GenState[banktypes.ModuleName] = simState.Cdc.MustMarshalJSON(&bankGen)
	}
}
