package simulation

import (
	"fmt"
	"math/rand"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	hold "github.com/provenance-io/provenance/x/hold"
)

const HoldAccountHolds = "hold-account-holds"

// RandomAccountHolds randomly selects accounts with an existing balance to place a hold of a random amount.
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

// RandomAccountHolds2 randomly selects accounts with an existing balance to place a hold of a random amount.
func RandomAccountHolds2(r *rand.Rand, balances []banktypes.Balance) []*hold.AccountHold {
	if len(balances) == 0 {
		return nil
	}

	count := r.Intn(len(balances) + 1)
	if count == 0 {
		return nil
	}

	randBals := make([]banktypes.Balance, 0, len(balances))
	for _, bal := range balances {
		if !bal.Coins.IsZero() {
			randBals = append(randBals, bal)
		}
	}
	r.Shuffle(len(randBals), func(i, j int) {
		randBals[i], randBals[j] = randBals[j], randBals[i]
	})

	rv := make([]*hold.AccountHold, count)
	for i, bal := range randBals[:count] {
		rv[i] = &hold.AccountHold{Address: bal.Address}
		// First, add 0 to 1000 of each denom.
		for _, coin := range bal.Coins {
			amt := r.Int63n(1001)
			holdCoin := sdk.NewInt64Coin(coin.Denom, amt)
			if !holdCoin.IsZero() {
				rv[i].Amount = append(rv[i].Amount, holdCoin)
			}
		}
		// If we still don't have a hold amount, add 1 to 1000 of a randomly selected denom.
		if rv[i].Amount.IsZero() {
			ind := r.Intn(len(bal.Coins))
			amt := r.Int63n(1000) + 1
			rv[i].Amount = append(rv[i].Amount, sdk.NewInt64Coin(bal.Coins[ind].Denom, amt))
		}
	}

	return rv
}

// UpdateBankGenStateForHolds adds all hold funds to the bank balances.
// Panics if there's an address with a hold that doesn't already have a balance.
func UpdateBankGenStateForHolds(bankGenState *banktypes.GenesisState, holdGenState *hold.GenesisState) {
	if len(holdGenState.Holds) == 0 {
		return
	}

	var totalAdded sdk.Coins
HoldsLoop:
	for _, ah := range holdGenState.Holds {
		for i, bal := range bankGenState.Balances {
			if ah.Address == bal.Address {
				bankGenState.Balances[i].Coins = bal.Coins.Add(ah.Amount...)
				totalAdded = totalAdded.Add(ah.Amount...)
				continue HoldsLoop
			}
		}
		panic(fmt.Errorf("no bank genesis balance found for %s that should have a hold on %s", ah.Address, ah.Amount))
	}

	bankGenState.Supply = bankGenState.Supply.Add(totalAdded...)
}

// addrCoinsStringsObjJSON creates a JSON object string of address -> amount fields for each provided entry.
func addrCoinsStringsObjJSON[T any](entries []T, getAddr func(T) string, getAmt func(T) sdk.Coins) string {
	if len(entries) == 0 {
		return "{}"
	}
	strs := make([]string, len(entries))
	for i, entry := range entries {
		strs[i] = fmt.Sprintf("%q:%q", getAddr(entry), getAmt(entry))
	}
	return fmt.Sprintf("{\n %s\n}", strings.Join(strs, ",\n "))
}

// holdsString creates a JSON object string of address -> amount for each hold.
func holdsString(holds []*hold.AccountHold) string {
	return addrCoinsStringsObjJSON(holds,
		func(ah *hold.AccountHold) string {
			return ah.Address
		},
		func(ah *hold.AccountHold) sdk.Coins {
			return ah.Amount
		},
	)
}

// balancesString creates a JSON object string of address -> amount for each balance.
func balancesString(balances []banktypes.Balance) string {
	return addrCoinsStringsObjJSON(balances,
		func(bal banktypes.Balance) string {
			return bal.Address
		},
		banktypes.Balance.GetCoins,
	)
}

// RandomizedGenState generates a random GenesisState for the hold module.
func RandomizedGenState(simState *module.SimulationState) {
	holdGenState := hold.DefaultGenesisState()

	simState.AppParams.GetOrGenerate(
		simState.Cdc, HoldAccountHolds, &holdGenState.Holds, simState.Rand,
		func(r *rand.Rand) {
			holdGenState.Holds = RandomAccountHolds(r, simState.Accounts)
		},
	)

	simState.GenState[hold.ModuleName] = simState.Cdc.MustMarshalJSON(holdGenState)
	fmt.Printf("Selected randomly generated holds:\n%s\n", holdsString(holdGenState.Holds))

	// If we put stuff in hold, add those funds to the bank balances.
	if len(holdGenState.Holds) > 0 {
		bankGenState := banktypes.GetGenesisStateFromAppState(simState.Cdc, simState.GenState)
		// TODO[1607]: Remove this next Printf (but keep the bottom one).
		fmt.Printf("Bank balances before update due to randomly generated holds:\n%s\n", balancesString(bankGenState.Balances))
		UpdateBankGenStateForHolds(bankGenState, holdGenState)
		simState.GenState[banktypes.ModuleName] = simState.Cdc.MustMarshalJSON(bankGenState)
		fmt.Printf("Bank balances after update due to randomly generated holds:\n%s\n", balancesString(bankGenState.Balances))
	}
}
