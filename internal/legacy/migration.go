package legacy

import (
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	v040 "github.com/provenance-io/provenance/internal/legacy/v040"
	"sort"
)

var migrationMap = types.MigrationMap{
	"v1.0.0": v040.Migrate, // provenance 1.0 blockchain is based on v40 sdk
}

// GetMigrationCallback returns a MigrationCallback for a given version.
func GetMigrationCallback(version string) types.MigrationCallback {
	return migrationMap[version]
}

// GetMigrationVersions get all migration version in a sorted slice.
func GetMigrationVersions() []string {
	versions := make([]string, len(migrationMap))

	var i int

	for version := range migrationMap {
		versions[i] = version
		i++
	}

	sort.Strings(versions)

	return versions
}
