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

// holdsString creates a JSON object string of address -> amount for each hold.
func holdsString(holds []*hold.AccountHold) string {
	if len(holds) == 0 {
		return "{}"
	}
	lines := make([]string, len(holds))
	for i, ah := range holds {
		lines[i] = fmt.Sprintf("%q:%q", ah.Address, ah.Amount)
	}
	return fmt.Sprintf("{\n %s\n}", strings.Join(lines, ",\n "))
}

// RandomizedGenState generates a random GenesisState for the hold module.
func RandomizedGenState(simState *module.SimulationState) {
	holdGenState := hold.DefaultGenesisState()

	simState.AppParams.GetOrGenerate(
		HoldAccountHolds, &holdGenState.Holds, simState.Rand,
		func(r *rand.Rand) {
			holdGenState.Holds = RandomAccountHolds(r, simState.Accounts)
		},
	)

	simState.GenState[hold.ModuleName] = simState.Cdc.MustMarshalJSON(holdGenState)
	fmt.Printf("Selected %d randomly generated holds:\n%s\n", len(holdGenState.Holds), holdsString(holdGenState.Holds))

	// If we put stuff in hold, add those funds to the bank balances.
	if len(holdGenState.Holds) > 0 {
		bankGenState := banktypes.GetGenesisStateFromAppState(simState.Cdc, simState.GenState)
		UpdateBankGenStateForHolds(bankGenState, holdGenState)
		simState.GenState[banktypes.ModuleName] = simState.Cdc.MustMarshalJSON(bankGenState)
	}
}
