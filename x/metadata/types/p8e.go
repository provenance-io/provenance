package types

import (
	"fmt"

	"github.com/google/uuid"
)

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

	// Set the scope pieces.
	scope.ScopeId, err = parseScopeID(msg.ScopeId)
	if err != nil {
		return
	}
	// TODO: Set scope.SpecificationId
	//       Not sure where to get this from.
	// TODO: Add scope.Owners.
	//       Get from contract.Recitals
	// TODO: Add scope.DataAccess entries.
	//       This is a new field. Leave it empty?
	// TODO: Set scope.ValueOwnerAddress.
	//       If the contract.Invoker public key matches one in contract.Recitals,
	//       then use that as the scope.ValueOwnerAddress.
	//       Otherwise, use scope.Parties looking for roles in this order: Marker, Owner, Originator.
	//       Otherwise, just use the first party.

	// Set the session pieces.
	session.SpecificationId, err = parseSessionID(scope.ScopeId, msg.GroupId)
	if err != nil {
		return
	}
	//       Parse it from the msg.GroupId
	// TODO: Set session.SpecificationId
	//       old way comes from contract.Spec.DataLocation.Ref.Hash string
	//       Might need to communicate a value change here?
	// TODO: Add session.Parties.
	//       Sae as the scope Owners.
	// TODO: Set session.Name
	//       Old way: From the contract spec, .Definition.ResourceLocation.Classname
	//       New way: From the contract spec, ClassName

	// Create the records.
	// TODO: Add records.
	//       Loop through the considerations.
	//       See old repo types/apply.go func considerationsAsRecords for clues.

	// Get the signers.
	if msg.Signatures != nil {
		for _, sig := range msg.Signatures.Signatures {
			if sig != nil && len(sig.Signature) > 0 {
				// TODO: verify that the sig.Signature value is what's desired here.
				//       other data piece: sig.Signer.SigningPublicKey.PublicKeyBytes []byte
				//       See old repo types/apply.go func OnChainRecitals for clues.
				signers = append(signers, sig.Signature)
			}
		}
	}
	return
}

func parseScopeID(input string) (MetadataAddress, error) {
	scopeID, maErr := MetadataAddressFromBech32(input)
	if maErr == nil {
		if !scopeID.IsScopeAddress() {
			return scopeID, fmt.Errorf("metadata address %s is not for a scope", scopeID)
		}
		return scopeID, nil
	}
	scopeUUID, uuidErr := uuid.Parse(input)
	if uuidErr == nil {
		return ScopeMetadataAddress(scopeUUID), nil
	}
	return MetadataAddress{}, fmt.Errorf("could not convert %s into either a scope metadata address (%s) or uuid (%s)",
		input, maErr.Error(), uuidErr.Error())
}

func parseSessionID(scopeID MetadataAddress, input string) (MetadataAddress, error) {
	sessionID, maErr := MetadataAddressFromBech32(input)
	if maErr == nil {
		if !sessionID.IsScopeAddress() {
			return sessionID, fmt.Errorf("metadata address %s is not for a session", sessionID)
		}
		return sessionID, nil
	}
	sessionUUID, uuidErr := uuid.Parse(input)
	if uuidErr == nil {
		return scopeID.AsSessionAddress(sessionUUID)
	}
	return MetadataAddress{}, fmt.Errorf("could not convert %s into either session metadata address (%s) or uuid (%s)",
		input, maErr.Error(), uuidErr.Error())
}
