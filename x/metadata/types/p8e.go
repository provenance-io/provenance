package types

// EmptyScope creates a new empty Scope.
func EmptyScope() *Scope {
	return &Scope{
		ScopeId:           MetadataAddress{},
		SpecificationId:   MetadataAddress{},
		Owners:            []Party{},
		DataAccess:        []string{},
		ValueOwnerAddress: "",
	}
}

// EmptySession creates a new empty Session.
func EmptySession() *Session {
	return &Session{
		SessionId:       MetadataAddress{},
		SpecificationId: MetadataAddress{},
		Parties:         []Party{},
		Name:            "",
		Audit:           nil,
	}
}

// EmptyRecord creates a new empty Record.
func EmptyRecord() *Record {
	return &Record{
		Name:      "",
		SessionId: MetadataAddress{},
		Process:   *EmptyProcess(),
		Inputs:    []RecordInput{},
		Outputs:   []RecordOutput{},
	}
}

// EmptyProcess creates a new empty Process.
func EmptyProcess() *Process {
	return &Process{
		ProcessId: nil,
		Name:      "",
		Method:    "",
	}
}

// Migrate Converts a MsgP8EMemorializeContractRequest object into the new objects.
func ConvertP8eMemorializeContractRequest(msg *MsgP8EMemorializeContractRequest) (scope Scope, session Session, records []Record, signers []string, err error) {
	scope = *EmptyScope()
	session = *EmptySession()
	records = []Record{}
	signers = []string{}
	err = nil

	// TODO: Set scope.ScopeId
	// TODO: Set scope.SpecificationId
	// TODO: Add scope.Owners.
	// TODO: Add scope.DataAccess entries.
	// TODO: Set scope.ValueOwnerAddress.

	// TODO: Set session.SessionId
	// TODO: Set session.SpecificationId
	// TODO: Add session.Parties.
	// TODO: Set session.Name

	// TODO: Add records.

	// TODO: Add signers.

	return
}
