package simulation

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/provenance-io/provenance/x/escrow"
)

const EscrowAccountEscrows = "escrow-account-escrows"

// RandomAccountEscrows randomly selects accounts and escrow amounts for the selected ones.
func RandomAccountEscrows(r *rand.Rand, accounts []simtypes.Account) []*escrow.AccountEscrow {
	if len(accounts) == 0 {
		return nil
	}

	count := r.Intn(len(accounts))
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

	rv := make([]*escrow.AccountEscrow, count)
	for i, addr := range addrs[:count] {
		rv[i] = &escrow.AccountEscrow{
			Address: addr.String(),
			Amount:  sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, r.Int63n(1000)+1)),
		}
	}

	return rv
}

// RandomizedGenState generates a random GenesisState for the escrow module.
func RandomizedGenState(simState *module.SimulationState) {
	genState := &escrow.GenesisState{}

	simState.AppParams.GetOrGenerate(
		simState.Cdc, EscrowAccountEscrows, &genState.Escrows, simState.Rand,
		func(r *rand.Rand) {
			genState.Escrows = RandomAccountEscrows(r, simState.Accounts)
		},
	)

	simState.GenState[escrow.ModuleName] = simState.Cdc.MustMarshalJSON(genState)

	// If we put stuff in escrow, add those funds to the bank accounts.
	if len(genState.Escrows) > 0 {
		bankGenRaw := simState.GenState[banktypes.ModuleName]
		bankGen := banktypes.GenesisState{}
		simState.Cdc.MustUnmarshalJSON(bankGenRaw, &bankGen)

		var totalAdded sdk.Coins
		var newBalances []banktypes.Balance
		for _, ae := range genState.Escrows {
			haveBal := false
			for _, bal := range bankGen.Balances {
				if bal.Address == ae.Address {
					bal.Coins = bal.Coins.Add(ae.Amount...)
					totalAdded = totalAdded.Add(ae.Amount...)
					haveBal = true
					break
				}
			}
			if !haveBal {
				newBalances = append(newBalances, banktypes.Balance{
					Address: ae.Address,
					Coins:   ae.Amount,
				})
				totalAdded = totalAdded.Add(ae.Amount...)
			}
		}
		bankGen.Balances = append(bankGen.Balances, newBalances...)
		bankGen.Supply = bankGen.Supply.Add(totalAdded...)

		simState.GenState[banktypes.ModuleName] = simState.Cdc.MustMarshalJSON(&bankGen)
	}
}
