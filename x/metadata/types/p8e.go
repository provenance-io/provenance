package types

import (
	"encoding/base64"
	fmt "fmt"

	"github.com/google/uuid"
	p8e "github.com/provenance-io/provenance/x/metadata/types/p8e"
)

// ConvertP8eContractSpec converts a v39 ContractSpec to a v40 ContractSpecification
func ConvertP8eContractSpec(old *p8e.ContractSpec, owners []string) (
	newSpec ContractSpecification,
	newRecords []RecordSpecification,
	err error,
) {
	raw, err := base64.StdEncoding.DecodeString(old.Definition.ResourceLocation.Ref.Hash)
	if err != nil {
		return newSpec, nil, err
	}
	specUUID, err := uuid.FromBytes(raw[0:16])
	if err != nil {
		return newSpec, nil, err
	}
	id := ContractSpecMetadataAddress(specUUID)

	parties := make([]PartyType, len(old.PartiesInvolved))
	for i := range old.PartiesInvolved {
		parties[i] = PartyType(int32(old.PartiesInvolved[i]))
	}

	newSpec = ContractSpecification{
		SpecificationId: id,
		Description: &Description{
			Name:        old.Definition.Name,
			Description: old.Definition.ResourceLocation.Classname,
		},
		PartiesInvolved: parties,
		OwnerAddresses:  owners,
		Source: &ContractSpecification_Hash{
			Hash: old.Definition.ResourceLocation.Ref.Hash,
		},
		ClassName: old.Definition.ResourceLocation.Classname,
	}
	err = newSpec.ValidateBasic()
	if err != nil {
		return ContractSpecification{}, nil, err
	}

	newRecords = make([]RecordSpecification, len(old.ConsiderationSpecs))
	for i := range old.ConsiderationSpecs {
		recordInputs, err := convertP8eInputSpecs(old.ConsiderationSpecs[i].InputSpecs)
		if err != nil {
			return newSpec, nil, err
		}
		specUUID, err := newSpec.SpecificationId.ContractSpecUUID()
		if err != nil {
			return newSpec, nil, err
		}
		recordSpecID := RecordSpecMetadataAddress(specUUID, old.ConsiderationSpecs[i].OutputSpec.Spec.Name)
		newRecords[i] = RecordSpecification{
			SpecificationId:    recordSpecID,
			Name:               old.ConsiderationSpecs[i].OutputSpec.Spec.Name,
			TypeName:           old.ConsiderationSpecs[i].OutputSpec.Spec.ResourceLocation.Classname,
			ResultType:         DefinitionType(old.ConsiderationSpecs[i].OutputSpec.Spec.Type),
			Inputs:             recordInputs,
			ResponsibleParties: []PartyType{PartyType(old.ConsiderationSpecs[i].ResponsibleParty)},
		}
		err = newRecords[i].ValidateBasic()
		if err != nil {
			return newSpec, nil, err
		}
	}

	return newSpec, newRecords, nil
}

// convertP8eInputSpecs a v39 DefinitionSpec used for inputs into a v40 InputSpecification
func convertP8eInputSpecs(old []*p8e.DefinitionSpec) (inputs []*InputSpecification, err error) {
	inputs = make([]*InputSpecification, len(old))
	for i, oldInput := range old {
		if oldInput.ResourceLocation.Ref.ScopeUuid != nil &&
			len(oldInput.ResourceLocation.Ref.ScopeUuid.Value) > 0 {
			scopeUUID, err := uuid.Parse(oldInput.ResourceLocation.Ref.ScopeUuid.Value)
			if err != nil {
				return nil, err
			}
			if len(oldInput.ResourceLocation.Ref.Name) < 1 {
				return nil, fmt.Errorf("must have a value for record name")
			}
			inputs[i] = &InputSpecification{
				Name:     oldInput.Name,
				TypeName: oldInput.ResourceLocation.Classname,
				Source: &InputSpecification_RecordId{
					RecordId: RecordMetadataAddress(scopeUUID, oldInput.ResourceLocation.Ref.Name),
				},
			}
		} else {
			inputs[i] = &InputSpecification{
				Name:     oldInput.Name,
				TypeName: oldInput.ResourceLocation.Classname,
				Source: &InputSpecification_Hash{
					Hash: oldInput.ResourceLocation.Ref.Hash,
				},
			}
		}
		err := inputs[i].ValidateBasic()
		if err != nil {
			return nil, err
		}
	}
	return inputs, nil
}
