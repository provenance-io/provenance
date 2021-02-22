package v040

import (
	"fmt"

	metadatakeeper "github.com/provenance-io/provenance/x/metadata/keeper"
	v039metadata "github.com/provenance-io/provenance/x/metadata/legacy/v039"
	v040metadata "github.com/provenance-io/provenance/x/metadata/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
)

// Migrate accepts exported x/metadata genesis state from v0.39 and migrates it
// to v0.40 x/metadata genesis state. The migration includes:
//
// - Convert addresses from bytes to bech32 strings.
// - Re-encode in v0.40 GenesisState.
func Migrate(oldGenState v039metadata.GenesisState) *v040metadata.GenesisState {
	var scopes = make([]v040metadata.Scope, 0, len(oldGenState.ScopeRecords))
	var groups = make([]v040metadata.RecordGroup, 0)
	var records = make([]v040metadata.Record, 0)
	var groupSpecs = make([]v040metadata.GroupSpecification, 0)

	for i := range oldGenState.ScopeRecords {
		s, g, r := convertScope(&oldGenState.ScopeRecords[i])
		scopes[i] = *s
		groups = append(groups, g...)
		records = append(records, r...)
	}

	// for i, spec := range oldGenState.Specifications {
	// groupSpecs = append(groupSpecs, )
	// todo convert spec v39 to v40
	// }

	return &v040metadata.GenesisState{
		Params: v040metadata.DefaultGenesisState().Params,

		Scopes:              scopes,
		Groups:              groups,
		Records:             records,
		GroupSpecifications: groupSpecs,
	}
}

// convertScope takes a v039 consolidated scope structure and returns the v040 break down into components
func convertScope(
	old *v039metadata.Scope,
) (
	newScope *v040metadata.Scope,
	newGroups []v040metadata.RecordGroup,
	newRecords []v040metadata.Record,
) {
	oldScopeUUID := uuid.MustParse(old.Uuid.Value)
	newScope = &v040metadata.Scope{
		ScopeId:         v040metadata.ScopeMetadataAddress(oldScopeUUID),
		SpecificationId: v040metadata.MetadataAddress{},
		Parties:         convertParties(old.Parties),
		DataAccess:      partyAddresses(convertParties(old.Parties)),
	}
	newGroups, newRecords = convertGroups(oldScopeUUID, old.RecordGroup)
	return
}

// convertGroup uses the old scope uuid to create a collection of updated v040 RecordGroup instances that derived from
// the groups in the old format
func convertGroups(oldScopeUUID uuid.UUID, old []*v039metadata.RecordGroup) (newGroup []v040metadata.RecordGroup, newRecords []v040metadata.Record) {
	newGroup = make([]v040metadata.RecordGroup, 0, len(old))
	newRecords = make([]v040metadata.Record, 0)
	for i, g := range old {
		newGroup[i].Audit.CreatedBy = g.Audit.CreatedBy
		newGroup[i].Audit.CreatedDate = g.Audit.CreatedDate
		newGroup[i].Audit.Message = g.Audit.Message
		newGroup[i].Audit.UpdatedBy = g.Audit.UpdatedBy
		newGroup[i].Audit.UpdatedDate = g.Audit.UpdatedDate
		newGroup[i].Audit.Version = uint32(g.Audit.Version)

		newGroup[i].GroupId = v040metadata.GroupMetadataAddress(oldScopeUUID, uuid.MustParse(g.GroupUuid.Value))
		newGroup[i].Name = g.Classname
		newGroup[i].Parties = convertParties(g.Parties)
		specAddr, err := v040metadata.ConvertHashToAddress(v040metadata.GroupKeyPrefix, g.Specification)
		if err != nil {
			panic(err)
		}
		newGroup[i].SpecificationId = specAddr

		newRecords = append(newRecords, convertRecords(newGroup[i].GroupId, g.Records)...)
	}
	return
}

// convertRecords converts the v039 Records within a RecordGroup structure to the updated independent record assigning
// each using the groupID provided.
func convertRecords(groupID v040metadata.MetadataAddress, old []*v039metadata.Record) (new []v040metadata.Record) {
	new = make([]v040metadata.Record, 0, len(old))
	for i, r := range old {
		new[i] = v040metadata.Record{
			Name:    r.ResultName,
			GroupId: groupID,
			Process: v040metadata.Process{
				Name:   r.Classname,
				Method: r.Name,
				ProcessId: &v040metadata.Process_Hash{
					Hash: r.Hash,
				},
			},
			Inputs: convertRecordInput(r.Inputs),
			Output: []v040metadata.RecordOutput{
				{
					Hash:   r.ResultHash,
					Status: v040metadata.ResultStatus(int32(r.Result)),
				},
			},
		}
	}
	return
}

// convertRecordInput converts the v039 RecordInput structure to the v040 RecordInput by mapping the old enums directly
// to the new ones (codes are preserved) and settings the source using the hash option (address was not supported)
func convertRecordInput(old []*v039metadata.RecordInput) (new []v040metadata.RecordInput) {
	new = make([]v040metadata.RecordInput, 0, len(old))
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
	new = make([]v040metadata.Party, 0, len(old))
	for i, r := range old {
		if len(r.Address) > 0 {
			new[i].Address = sdk.AccAddress(r.Address).String()
		} else {
			// must parse signing key into address
			if r.Signer.SigningPublicKey == nil ||
				r.Signer.SigningPublicKey.Curve != v039metadata.PublicKeyCurve_SECP256K1 {
				panic(fmt.Errorf("unsupported signing publickey type and account address unavailable"))
			}
			_, addr, err := v039metadata.ParsePublicKey(r.Signer.EncryptionPublicKey.PublicKeyBytes)
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

// BackportScope takes a reference to a current scope and backports it to the v039 version by reading the components
// from the keeper to reconstruct it.
func BackportScope(
	ctx sdk.Context,
	k metadatakeeper.Keeper,
	newScope v040metadata.Scope,
) (old v039metadata.Scope, err error) {
	prefix, err := newScope.ScopeId.ScopeGroupIteratorPrefix()
	if err != nil {
		return
	}
	newestGroup := ""
	newestGroupAge := int64(0)
	oldGroups := make([]*v039metadata.RecordGroup, 0)
	err = k.IterateGroups(ctx, prefix, func(t v040metadata.RecordGroup) (stop bool) {
		var groupUUID uuid.UUID
		groupUUID, err = t.GroupId.GroupUUID()
		if err != nil {
			return
		}
		if t.Audit.CreatedDate.UnixNano() > newestGroupAge {
			newestGroupAge = t.Audit.CreatedDate.UnixNano()
			newestGroup = groupUUID.String()
		}
		if t.Audit.UpdatedDate.UnixNano() > newestGroupAge {
			newestGroupAge = t.Audit.UpdatedDate.UnixNano()
			newestGroup = groupUUID.String()
		}

		var parties []*v039metadata.Recital
		parties, err = backportParties(t.Parties)
		if err != nil {
			return
		}

		groupRecords := make([]*v039metadata.Record, 0)
		err = k.IterateRecords(ctx, newScope.ScopeId, func(r v040metadata.Record) (stop bool) {
			if r.GroupId.Equals(t.GroupId) {
				groupRecords = append(groupRecords, &v039metadata.Record{
					Name: r.Process.Method,
					//Hash: r.Process.ProcessId as hash
					ResultHash: r.Output[0].Hash, // v039 only supports a single result hash, use first
					Result:     v039metadata.ExecutionResultType(int32(r.Output[0].Status)),
					ResultName: r.Name,
					Classname:  r.Process.Name,
					Inputs:     backportInputs(r.Inputs),
				})
			}
			return false
		})
		if err != nil {
			return
		}

		executorBech32 := ""
		var executor v039metadata.SigningAndEncryptionPublicKeys

		if len(t.Audit.UpdatedBy) > 0 {
			executorBech32 = t.Audit.UpdatedBy
		} else if len(t.Audit.CreatedBy) > 0 {
			executorBech32 = t.Audit.CreatedBy
		}

		if executorBech32 != "" {
			var addr sdk.AccAddress
			addr, err = sdk.AccAddressFromBech32(executorBech32)
			if err != nil {
				return
			}
			acc := k.GetAccount(ctx, addr)
			if acc != nil && acc.GetPubKey() != nil {
				executor = v039metadata.SigningAndEncryptionPublicKeys{
					SigningPublicKey: &v039metadata.PublicKey{
						PublicKeyBytes: acc.GetPubKey().Bytes(),
						// TODO: only SECP256k1 keys are supported in v40 and v39, this assumption should be checked.
						Type:  v039metadata.PublicKeyType_ELLIPTIC,
						Curve: v039metadata.PublicKeyCurve_SECP256K1,
					},
				}
			}
		}

		specHash := ""
		spec, found := k.GetGroupSpecification(ctx, t.SpecificationId)
		if found {
			specHash = spec.Definition.ResourceLocation.Hash
		}

		oldGroups = append(oldGroups, &v039metadata.RecordGroup{
			Classname: t.Name,
			GroupUuid: &v039metadata.UUID{
				Value: groupUUID.String(),
			},
			Parties:       parties,
			Executor:      &executor,
			Records:       groupRecords,
			Specification: specHash,
			Audit: &v039metadata.AuditFields{
				CreatedBy:   t.Audit.CreatedBy,
				CreatedDate: t.Audit.CreatedDate,
				Message:     t.Audit.Message,
				UpdatedBy:   t.Audit.UpdatedBy,
				UpdatedDate: t.Audit.UpdatedDate,
				Version:     int32(t.Audit.Version),
			},
		})
		return false
	})

	if err != nil {
		return
	}

	oldID, err := newScope.ScopeId.ScopeUUID()
	if err != nil {
		return
	}

	oldParties, err := backportParties(newScope.Parties)
	if err != nil {
		return
	}

	old = v039metadata.Scope{
		Uuid:        &v039metadata.UUID{Value: oldID.String()},
		Parties:     oldParties,
		RecordGroup: oldGroups,
		LastEvent: &v039metadata.Event{
			GroupUuid: &v039metadata.UUID{
				Value: newestGroup,
			},
			// NOTE: there is no concept of execution uuid, this should be the group uuid
		},
	}

	return old, err
}

func backportInputs(new []v040metadata.RecordInput) (old []*v039metadata.RecordInput) {
	old = make([]*v039metadata.RecordInput, 0, len(new))
	for i, ri := range new {
		old[i] = &v039metadata.RecordInput{
			Name:      ri.Name,
			Classname: ri.TypeName,
			// Hash: ri.Source as Hash,
			Type: v039metadata.RecordInputType(int32(ri.Status)),
		}
	}
	return
}

func backportAddressToParties(new []string, role v040metadata.PartyType) (old []*v039metadata.Recital, err error) {
	old = make([]*v039metadata.Recital, 0, len(new))
	for i, n := range new {
		addr, err := sdk.AccAddressFromBech32(n)
		if err != nil {
			return nil, err
		}
		old[i].Address = addr
		old[i].SignerRole = v039metadata.PartyType(int32(role))

		// TODO consider including a context here to bring in the public keys by query of AccountKeeper
		// old[i].Signer.SigningPublicKey
	}
	return
}

func backportParties(new []v040metadata.Party) (old []*v039metadata.Recital, err error) {
	old = make([]*v039metadata.Recital, 0, len(new))
	for i, n := range new {
		addr, err := sdk.AccAddressFromBech32(n.Address)
		if err != nil {
			return nil, err
		}
		old[i].Address = addr
		old[i].SignerRole = v039metadata.PartyType(int32(n.Role))

		// TODO consider including a context here to bring in the public keys by query of AccountKeeper
		// old[i].Signer.SigningPublicKey
	}
	return
}
