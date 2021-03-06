package v040

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"time"

	v039metadata "github.com/provenance-io/provenance/x/metadata/legacy/v039"
	v040metadata "github.com/provenance-io/provenance/x/metadata/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/google/uuid"
)

// Migrate accepts exported x/metadata genesis state from v0.39 and migrates it
// to v0.40 x/metadata genesis state. The migration includes:
//
// - Convert addresses from bytes to bech32 strings.
// - Re-encode in v0.40 GenesisState.
func Migrate(oldGenState v039metadata.GenesisState) *v040metadata.GenesisState {
	var scopes = make([]v040metadata.Scope, len(oldGenState.ScopeRecords))
	var sessions = make([]v040metadata.Session, 0)
	var records = make([]v040metadata.Record, 0)
	var contractSpecs = make([]v040metadata.ContractSpecification, 0)
	var recordSpecs = make([]v040metadata.RecordSpecification, 0)

	for i := range oldGenState.ScopeRecords {
		s, g, r := convertScope(&oldGenState.ScopeRecords[i])
		scopes[i] = *s
		sessions = append(sessions, g...)
		records = append(records, r...)
	}

	for i := range oldGenState.Specifications {
		c, r, e := convertContractSpec(&oldGenState.Specifications[i])
		if e != nil {
			panic(e)
		}
		contractSpecs = append(contractSpecs, c)
		recordSpecs = append(recordSpecs, r...)
	}

	return &v040metadata.GenesisState{
		Params:          v040metadata.DefaultGenesisState().Params,
		OSLocatorParams: v040metadata.DefaultGenesisState().OSLocatorParams,

		Scopes:                 scopes,
		Sessions:               sessions,
		Records:                records,
		ContractSpecifications: contractSpecs,
		RecordSpecifications:   recordSpecs,
	}
}

// convertScope takes a v039 consolidated scope structure and returns the v040 break down into components
func convertScope(
	old *v039metadata.Scope,
) (
	newScope *v040metadata.Scope,
	newSessions []v040metadata.Session,
	newRecords []v040metadata.Record,
) {
	oldScopeUUID := uuid.MustParse(old.Uuid.Value)
	newScope = &v040metadata.Scope{
		ScopeId:         v040metadata.ScopeMetadataAddress(oldScopeUUID),
		SpecificationId: v040metadata.MetadataAddress{},
		Owners:          convertParties(old.Parties),
		DataAccess:      partyAddresses(convertParties(old.Parties)),
	}
	newSessions, newRecords = convertGroups(oldScopeUUID, old.RecordGroup)
	return
}

// convertGroup uses the old scope uuid to create a collection of updated v040 RecordGroup instances that derived from
// the groups in the old format
func convertGroups(oldScopeUUID uuid.UUID, old []*v039metadata.RecordGroup) (newSession []v040metadata.Session, newRecords []v040metadata.Record) {
	newSession = make([]v040metadata.Session, len(old))
	newRecords = make([]v040metadata.Record, 0)
	for i, g := range old {
		if g.Audit != nil {
			newSession[i].Audit = &v040metadata.AuditFields{CreatedBy: g.Audit.CreatedBy}
			if g.Audit.CreatedDate != nil {
				newSession[i].Audit.CreatedDate = time.Unix(g.Audit.CreatedDate.Seconds, 0)
			}
			newSession[i].Audit.Message = g.Audit.Message
			newSession[i].Audit.UpdatedBy = g.Audit.UpdatedBy
			if g.Audit.UpdatedDate != nil {
				newSession[i].Audit.UpdatedDate = time.Unix(g.Audit.UpdatedDate.Seconds, 0)
			}
			newSession[i].Audit.Version = uint32(g.Audit.Version)
		}
		newSession[i].SessionId = v040metadata.SessionMetadataAddress(oldScopeUUID, uuid.MustParse(g.GroupUuid.Value))
		newSession[i].Name = g.Classname
		newSession[i].Parties = convertParties(g.Parties)
		specAddr, err := v040metadata.ConvertHashToAddress(v040metadata.ContractSpecificationKeyPrefix, g.Specification)
		if err != nil {
			panic(err)
		}
		newSession[i].SpecificationId = specAddr

		newRecords = append(newRecords, convertRecords(newSession[i].SessionId, specAddr, g.Records)...)
	}
	return
}

// convertRecords converts the v039 Records within a RecordGroup structure to the updated independent record assigning
// each using the groupID provided.
func convertRecords(sessionID v040metadata.MetadataAddress, cSpecAddr v040metadata.MetadataAddress, old []*v039metadata.Record) (new []v040metadata.Record) {
	cSpecUUID, err := cSpecAddr.ContractSpecUUID()
	if err != nil {
		panic(err)
	}
	new = make([]v040metadata.Record, len(old))
	for i, r := range old {
		new[i] = v040metadata.Record{
			Name:      r.ResultName,
			SessionId: sessionID,
			Process: v040metadata.Process{
				Name:   r.Classname,
				Method: r.Name,
				ProcessId: &v040metadata.Process_Hash{
					Hash: r.Hash,
				},
			},
			Inputs: convertRecordInput(r.Inputs),
			Outputs: []v040metadata.RecordOutput{
				{
					Hash:   r.ResultHash,
					Status: v040metadata.ResultStatus(int32(r.Result)),
				},
			},
			SpecificationId: v040metadata.RecordSpecMetadataAddress(cSpecUUID, r.ResultName),
		}
	}
	return
}

// convertRecordInput converts the v039 RecordInput structure to the v040 RecordInput by mapping the old enums directly
// to the new ones (codes are preserved) and settings the source using the hash option (address was not supported)
func convertRecordInput(old []*v039metadata.RecordInput) (new []v040metadata.RecordInput) {
	new = make([]v040metadata.RecordInput, len(old))
	for i, input := range old {
		new[i] = v040metadata.RecordInput{
			Name:     input.Name,
			TypeName: input.Classname,
			Source: &v040metadata.RecordInput_Hash{
				Hash: input.Hash,
			},
			Status: v040metadata.RecordInputStatus(int32(input.Type)),
		}
	}
	return
}

// partyAddresses returns an array of addresses from an array of parties
func partyAddresses(parties []v040metadata.Party) (addresses []string) {
	for _, p := range parties {
		if len(p.Address) > 0 {
			addresses = append(addresses, p.Address)
		}
	}
	return
}

// convertParties converts the v039 Recital structure into a v040 Party by calculating the address (as required) and
// copying over the existing party role value into the new structure
func convertParties(old []*v039metadata.Recital) (new []v040metadata.Party) {
	new = make([]v040metadata.Party, len(old))
	for i, r := range old {
		if len(r.Address) > 0 {
			new[i].Address = sdk.AccAddress(r.Address).String()
		} else {
			// must parse signing key into address
			if r.Signer.SigningPublicKey == nil ||
				r.Signer.SigningPublicKey.Curve != v039metadata.PublicKeyCurve_SECP256K1 {
				panic(fmt.Errorf("unsupported signing publickey type and account address unavailable"))
			}
			_, addr, err := v039metadata.ParsePublicKey(r.Signer.SigningPublicKey.PublicKeyBytes)
			if err != nil {
				panic(err)
			}
			new[i].Address = addr.String()
		}
		// old v39 and new v40 enum codes are the same
		new[i].Role = v040metadata.PartyType(int32(r.SignerRole))
	}
	return
}

func convertContractSpec(old *v039metadata.ContractSpec) (
	newSpec v040metadata.ContractSpecification,
	newRecords []v040metadata.RecordSpecification,
	err error,
) {
	rawProtoOld, err := proto.Marshal(old)
	if err != nil {
		return newSpec, nil, err
	}
	sha512Old := sha512.Sum512(rawProtoOld)
	specUUID, err := uuid.FromBytes(sha512Old[0:16])
	if err != nil {
		return newSpec, nil, err
	}
	id := v040metadata.ContractSpecMetadataAddress(specUUID)

	parties := make([]v040metadata.PartyType, len(old.PartiesInvolved))
	for i := range old.PartiesInvolved {
		parties[i] = v040metadata.PartyType(int32(old.PartiesInvolved[i]))
	}

	newSpec = v040metadata.ContractSpecification{
		SpecificationId: id,
		Description: &v040metadata.Description{
			Name:        old.Definition.Name,
			Description: old.Definition.ResourceLocation.Classname,
		},
		PartiesInvolved: parties,
		// OwnerAddresses: -- TODO: there were no owners set on the v39 chain, maybe trace one from a group that used this spec?
		Source: &v040metadata.ContractSpecification_Hash{
			Hash: base64.StdEncoding.EncodeToString(sha512Old[:]),
		},
		ClassName: old.Definition.ResourceLocation.Classname,
	}

	newRecords = make([]v040metadata.RecordSpecification, len(old.ConsiderationSpecs))
	for i := range old.ConsiderationSpecs {
		recordInputs, err := convertInputSpecs(old.ConsiderationSpecs[i].InputSpecs)
		if err != nil {
			return newSpec, nil, err
		}
		specUUID, err := newSpec.SpecificationId.ContractSpecUUID()
		if err != nil {
			return newSpec, nil, err
		}
		recordSpecID := v040metadata.RecordSpecMetadataAddress(specUUID, old.ConsiderationSpecs[i].OutputSpec.Spec.Name)
		newRecords[i] = v040metadata.RecordSpecification{
			SpecificationId:    recordSpecID,
			Name:               old.ConsiderationSpecs[i].OutputSpec.Spec.Name,
			TypeName:           old.ConsiderationSpecs[i].OutputSpec.Spec.ResourceLocation.Classname,
			ResultType:         v040metadata.DefinitionType(old.ConsiderationSpecs[i].OutputSpec.Spec.Type),
			Inputs:             recordInputs,
			ResponsibleParties: []v040metadata.PartyType{v040metadata.PartyType(old.ConsiderationSpecs[i].ResponsibleParty)},
		}
	}

	return newSpec, newRecords, nil
}

// converts a v39 DefinitionSpec used for inputs into a v40 input specification.
func convertInputSpecs(old []*v039metadata.DefinitionSpec) (inputs []*v040metadata.InputSpecification, err error) {
	inputs = make([]*v040metadata.InputSpecification, len(old))
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
			inputs[i] = &v040metadata.InputSpecification{
				Name:     oldInput.Name,
				TypeName: oldInput.ResourceLocation.Classname,
				Source: &v040metadata.InputSpecification_RecordId{
					RecordId: v040metadata.RecordMetadataAddress(scopeUUID, oldInput.ResourceLocation.Ref.Name),
				},
			}
		} else {
			inputs[i] = &v040metadata.InputSpecification{
				Name:     oldInput.Name,
				TypeName: oldInput.ResourceLocation.Classname,
				Source: &v040metadata.InputSpecification_Hash{
					Hash: oldInput.ResourceLocation.Ref.Hash,
				},
			}
		}
	}
	return
}
