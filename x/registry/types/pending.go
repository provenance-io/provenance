package types

import (
	"crypto/sha256"
	"encoding/hex"
	"slices"
	"strconv"
	"strings"
)

// NewPendingRoleChangeID computes the deterministic identifier for a pending role change.
//
// The id is a hash of the registry key and the (order-independent) set of role updates, each
// reduced to its role and the order-independent set of desired addresses. Two proposals describing
// the same batch of role updates therefore collapse onto the same pending change, so their
// approvals accumulate together.
func NewPendingRoleChangeID(key *RegistryKey, roleUpdates []RoleUpdate) string {
	// Reduce each update to a canonical "role|sortedAddrs" line, then sort the lines so the id is
	// independent of both address order within a role and role order across the batch.
	lines := make([]string, len(roleUpdates))
	for i, update := range roleUpdates {
		sorted := slices.Clone(update.Addresses)
		slices.Sort(sorted)
		lines[i] = strconv.Itoa(int(update.Role)) + "|" + strings.Join(sorted, ",")
	}
	slices.Sort(lines)

	var b strings.Builder
	b.WriteString(key.AssetClassId)
	b.WriteByte(0)
	b.WriteString(key.NftId)
	b.WriteByte(0)
	b.WriteString(strings.Join(lines, "\n"))

	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}
