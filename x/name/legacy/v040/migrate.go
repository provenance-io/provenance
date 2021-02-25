package v040

import (
	v039name "github.com/provenance-io/provenance/x/name/legacy/v039"
	v040name "github.com/provenance-io/provenance/x/name/types"
)

// Migrate accepts exported x/name genesis state from v0.39 and migrates it
// to v0.40 x/name genesis state. The migration includes:
//
// - Convert addresses from bytes to bech32 strings.
// - Re-encode in v0.40 GenesisState.
func Migrate(oldGenState v039name.GenesisState) *v040name.GenesisState {
	var nameRecords = make([]v040name.NameRecord, 0, len(oldGenState.Bindings))
	for _,name := range oldGenState.Bindings {
		nameRecords = append(nameRecords, v040name.NameRecord{
			Name:       name.Name,
			Address:    name.Address.String(),
			Restricted: name.Restricted,
		})
	}

	return &v040name.GenesisState{
		Params: v040name.DefaultParams(),
		Bindings: nameRecords,
	}
}
