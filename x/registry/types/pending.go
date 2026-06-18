package types

import (
	"crypto/sha256"
	"encoding/hex"
	"slices"
	"strconv"
	"strings"
)

// ShortString returns a compact human-readable form of the operation (e.g. "GRANT").
func (o RoleChangeOperation) ShortString() string {
	return strings.TrimPrefix(o.String(), "ROLE_CHANGE_OPERATION_")
}

// NewPendingRoleChangeID computes the deterministic identifier for a pending role change.
//
// The id is a hash of the registry key, role, operation, and the (order-independent) set of
// target addresses. Two proposals describing the same change therefore collapse onto the same
// pending change, so their approvals accumulate together.
func NewPendingRoleChangeID(key *RegistryKey, role RegistryRole, op RoleChangeOperation, addresses []string) string {
	sorted := slices.Clone(addresses)
	slices.Sort(sorted)

	var b strings.Builder
	b.WriteString(key.AssetClassId)
	b.WriteByte(0)
	b.WriteString(key.NftId)
	b.WriteByte(0)
	b.WriteString(strconv.Itoa(int(role)))
	b.WriteByte(0)
	b.WriteString(strconv.Itoa(int(op)))
	b.WriteByte(0)
	b.WriteString(strings.Join(sorted, ","))

	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}
