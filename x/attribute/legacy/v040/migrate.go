package v040

import (
	v039attribute "github.com/provenance-io/provenance/x/attribute/legacy/v039"
	v040attribute "github.com/provenance-io/provenance/x/attribute/types"
)

// Migrate accepts exported x/attribute genesis state from v0.39 and migrates it
// to v0.40 x/attribute genesis state. The migration includes:
//
// - Convert addresses from bytes to bech32 strings.
// - Re-encode in v0.40 GenesisState.
func Migrate(oldGenState v039attribute.GenesisState) *v040attribute.GenesisState {
	var attributeAccounts = make([]v040attribute.Attribute, 0, len(oldGenState.Attributes))
	for _, at := range oldGenState.Attributes {
		atType, err := v040attribute.AttributeTypeFromString(at.Type)
		if err != nil {
			panic(err)
		}
		attributeAccounts = append(attributeAccounts, v040attribute.Attribute{
			Name:          at.Name,
			Address:       at.Address,
			Value:         at.Value,
			AttributeType: atType,
		})
	}
	return &v040attribute.GenesisState{
		Params:     v040attribute.DefaultParams(),
		Attributes: attributeAccounts,
	}
}
