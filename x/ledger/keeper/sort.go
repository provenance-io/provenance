// sort.go contains sorting logic for ledger types
package keeper

import (
	"sort"

	ledger "github.com/provenance-io/provenance/x/ledger/types"
)

// sortLedgerEntries sorts the ledger entries by effective date and then by sequence.
func sortLedgerEntries(entries []*ledger.LedgerEntry) {
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].EffectiveDate == entries[j].EffectiveDate {
			return (entries)[i].Sequence < (entries)[j].Sequence
		}
		return entries[i].EffectiveDate < entries[j].EffectiveDate
	})
}
