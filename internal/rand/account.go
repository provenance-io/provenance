package rand

import (
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// SelectAccounts selects count accounts from the ones provided.
func SelectAccounts(r *rand.Rand, accs []simtypes.Account, count int) ([]simtypes.Account, error) {
	return SelectEntries(r, accs, count, "accounts")
}
