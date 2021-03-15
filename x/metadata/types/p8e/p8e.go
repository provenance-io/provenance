package p8e

import "github.com/provenance-io/provenance/x/metadata/types"

// Migrate Converts a P8eMsgMemorializeP8eContractRequest object into the new objects.
func (p *P8EMsgMemorializeP8EContractRequest) Migrate() (scope types.Scope, session types.Session, records []types.Record) {
	scope = *EmptyScope()
	// TODO: Set scope.ScopeId
	// TODO: Set scope.SpecificationId
	// TODO: Add scope.Owners.
	// TODO: Add scope.DataAccess entries.
	// TODO: Set scope.ValueOwnerAddress.

	// TODO: Set proper session id and set proper specification id.
	session = *EmptySession()
	// TODO: Set session.SessionId
	// TODO: Set session.SpecificationId
	// TODO: Add session.Parties.
	// TODO: Set session.Name

	records = []types.Record{}
	// TODO: Add records.

	return
}

// EmptyScope creates a new empty Scope.
func EmptyScope() *types.Scope {
	return &types.Scope{
		ScopeId:           types.MetadataAddress{},
		SpecificationId:   types.MetadataAddress{},
		Owners:            []types.Party{},
		DataAccess:        []string{},
		ValueOwnerAddress: "",
	}
}

// EmptySession creates a new empty Session.
func EmptySession() *types.Session {
	return &types.Session{
		SessionId:       types.MetadataAddress{},
		SpecificationId: types.MetadataAddress{},
		Parties:         []types.Party{},
		Name:            "",
		Audit:           nil,
	}
}

// EmptyRecord creates a new empty Record.
func EmptyRecord() *types.Record {
	return &types.Record {
		Name:      "",
		SessionId: types.MetadataAddress{},
		Process:   *EmptyProcess(),
		Inputs:    []types.RecordInput{},
		Outputs:   []types.RecordOutput{},
	}
}

// EmptyProcess creates a new empty Process.
func EmptyProcess() *types.Process {
	return &types.Process{
		ProcessId: nil,
		Name:      "",
		Method:    "",
	}
}
