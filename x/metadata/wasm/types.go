package wasm

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/provenance-io/provenance/x/metadata/types"
)

// These are the Go struct that map to the Rust smart contract types defined by the provwasm JSON schema.
// The types in this file were generated using quicktype.io.

// Scope is a root reference for a collection of records owned by one or more parties.
type Scope struct {
	ScopeID           string   `json:"scope_id"`
	SpecificationID   string   `json:"specification_id"`
	DataAccess        []string `json:"data_access,omitempty"`
	Owners            []*Party `json:"owners,omitempty"`
	ValueOwnerAddress string   `json:"value_owner_address"`
}

// Sessions is a group of sessions.
type Sessions struct {
	Sessions []*Session `json:"sessions"`
}

// Session is the state of an execution context for a specification instance.
type Session struct {
	SessionID       string   `json:"session_id"`
	SpecificationID string   `json:"specification_id"`
	Name            string   `json:"name"`
	Context         []byte   `json:"context"`
	Parties         []*Party `json:"parties,omitempty"`
}

// PartyType defines roles that can be associated to a party.
type PartyType string

const (
	// PartyTypeAffiliate is a concrete party type.
	PartyTypeAffiliate PartyType = "affiliate"
	// PartyTypeCustodian is a concrete party type.
	PartyTypeCustodian PartyType = "custodian"
	// PartyTypeInvestor is a concrete party type.
	PartyTypeInvestor PartyType = "investor"
	// PartyTypeOmnibus is a concrete party type.
	PartyTypeOmnibus PartyType = "omnibus"
	// PartyTypeOriginator is a concrete party type.
	PartyTypeOriginator PartyType = "originator"
	// PartyTypeOwner is a concrete party type.
	PartyTypeOwner PartyType = "owner"
	// PartyTypeProvenance is a concrete party type.
	PartyTypeProvenance PartyType = "provenance"
	// PartyTypeServicer is a concrete party type.
	PartyTypeServicer PartyType = "servicer"
	// PartyTypeUnspecified is a concrete party type.
	PartyTypeUnspecified PartyType = "unspecified"
)

// Party is an address with an associated role.
type Party struct {
	Address string    `json:"address"`
	Role    PartyType `json:"role"`
}

// Records is a group of records.
type Records struct {
	Records []*Record `json:"records"`
}

// Record is a record of fact for a session.
type Record struct {
	SessionID       string          `json:"session_id"`
	SpecificationID string          `json:"specification_id"`
	Name            string          `json:"name"`
	Process         *Process        `json:"process"`
	Inputs          []*RecordInput  `json:"inputs,omitempty"`
	Outputs         []*RecordOutput `json:"outputs,omitempty"`
}

// RecordInput is an input used to produce a record.
type RecordInput struct {
	Name     string             `json:"name"`
	TypeName string             `json:"type_name"`
	Source   *RecordInputSource `json:"source"`
	Status   InputStatus        `json:"status"`
}

// RecordInputSource is a record input source. Either record or hash should be set, but not both.
type RecordInputSource struct {
	Record *RecordInputSourceRecord `json:"record,omitempty"`
	Hash   *RecordInputSourceHash   `json:"hash,omitempty"`
}

// RecordInputSourceRecord is the address of a record on chain (established records).
type RecordInputSourceRecord struct {
	RecordID string `json:"record_id"`
}

// RecordInputSourceHash is the hash of an off-chain piece of information (proposed records).
type RecordInputSourceHash struct {
	Hash string `json:"hash"`
}

// RecordOutput is the output of a process.
type RecordOutput struct {
	Hash   string       `json:"hash"`
	Status ResultStatus `json:"status"`
}

// Process is the entity that generated a record.
type Process struct {
	ProcessID *ProcessID `json:"process_id"`
	Name      string     `json:"name"`
	Method    string     `json:"method"`
}

// ProcessID is a process identifier. Either address or hash should be set, but not both.
type ProcessID struct {
	Address *ProcessIDAddress `json:"address,omitempty"`
	Hash    *ProcessIDHash    `json:"hash,omitempty"`
}

// ProcessIDAddress is the on-chain address of a process.
type ProcessIDAddress struct {
	Address string `json:"address"`
}

// ProcessIDHash is the hash of an off-chain process.
type ProcessIDHash struct {
	Hash string `json:"hash"`
}

// InputStatus defines record input status types.
type InputStatus string

const (
	// InputStatusRecord is a concrete record input status type.
	InputStatusRecord InputStatus = "record"
	// InputStatusProposed is a concrete record input status type.
	InputStatusProposed InputStatus = "proposed"
	// InputStatusUnspecified is a concrete record input status type.
	InputStatusUnspecified InputStatus = "unspecified"
)

// Result status types.
type ResultStatus string

const (
	// ResultStatusPass is a concrete result status type.
	ResultStatusPass ResultStatus = "pass"
	// ResultStatusFail is a concrete result status type.
	ResultStatusFail ResultStatus = "fail"
	// ResultStatusSkip is a concrete result status type.
	ResultStatusSkip ResultStatus = "skip"
	// ResultStatusUnspecified is a concrete result status type.
	ResultStatusUnspecified ResultStatus = "unspecified"
)

// A helper function for converting metadata module scope types into provwasm query response types.
func createScopeResponse(input types.Scope) ([]byte, error) {
	scopeID, err := stringifyAddress(input.ScopeId)
	if err != nil {
		return nil, err
	}
	specID, err := stringifyAddress(input.SpecificationId)
	if err != nil {
		return nil, err
	}
	scope := &Scope{
		ScopeID:           scopeID,
		SpecificationID:   specID,
		ValueOwnerAddress: input.ValueOwnerAddress,
	}
	for _, da := range input.DataAccess {
		scope.DataAccess = append(scope.DataAccess, da)
	}
	for _, o := range input.Owners {
		scope.Owners = append(scope.Owners, createParty(o))
	}
	bz, err := json.Marshal(scope)
	if err != nil {
		return nil, fmt.Errorf("wasm: marshal response failed: %w", err)
	}
	return bz, nil
}

// A helper function for converting metadata module session types into provwasm query response types.
func createSessionsResponse(input []types.Session) ([]byte, error) {
	ss := &Sessions{}
	for _, s := range input {
		sessionID, err := stringifyAddress(s.SessionId)
		if err != nil {
			return nil, err
		}
		specID, err := stringifyAddress(s.SpecificationId)
		if err != nil {
			return nil, err
		}
		session := &Session{
			SessionID:       sessionID,
			SpecificationID: specID,
			Name:            s.Name,
			Context:         append([]byte{}, s.Context...),
		}
		for _, p := range s.Parties {
			session.Parties = append(session.Parties, createParty(p))
		}
		ss.Sessions = append(ss.Sessions, session)
	}
	bz, err := json.Marshal(ss)
	if err != nil {
		return nil, fmt.Errorf("wasm: marshal response failed: %w", err)
	}
	return bz, nil
}

// A helper function for converting metadata module record types into provwasm query response types.
func createRecordsResponse(input []*types.Record) ([]byte, error) {
	rs := &Records{}
	for _, r := range input {
		sessionID, err := stringifyAddress(r.SessionId)
		if err != nil {
			return nil, err
		}
		specID, err := stringifyAddress(r.SpecificationId)
		if err != nil {
			return nil, err
		}
		process, err := createProcess(r.Process)
		if err != nil {
			return nil, err
		}
		record := &Record{
			SessionID:       sessionID,
			SpecificationID: specID,
			Name:            r.Name,
			Process:         process,
		}
		for _, i := range r.Inputs {
			source, err := createSource(i)
			if err != nil {
				return nil, err
			}
			ri := &RecordInput{
				Name:     i.Name,
				TypeName: i.TypeName,
				Source:   source,
				Status:   createRecordInputStatus(i.Status),
			}
			record.Inputs = append(record.Inputs, ri)
		}
		for _, o := range r.Outputs {
			ro := &RecordOutput{
				Hash:   o.Hash,
				Status: createResultStatus(o.Status),
			}
			record.Outputs = append(record.Outputs, ro)
		}
		rs.Records = append(rs.Records, record)
	}
	bz, err := json.Marshal(rs)
	if err != nil {
		return nil, fmt.Errorf("wasm: marshal response failed: %w", err)
	}
	return bz, nil
}

// A non-panicing version of MetadataAddress.String(). We don't want query panics in smart contracts.
// Just return the error, providing more info to the provwasm side.
func stringifyAddress(ma types.MetadataAddress) (string, error) {
	if ma.Empty() {
		return "", fmt.Errorf("wasm: empty metadata address")
	}
	hrp, err := types.VerifyMetadataAddressFormat(ma)
	if err != nil {
		return "", err
	}
	bech32Addr, err := bech32.ConvertAndEncode(hrp, ma.Bytes())
	if err != nil {
		return "", err
	}
	return bech32Addr, nil
}

// Convert a party to its provwasm type.
func createParty(input types.Party) *Party {
	return &Party{
		Address: input.Address,
		Role:    createRole(input.Role),
	}
}

// Convert a party type to its provwasm type.
func createRole(input types.PartyType) PartyType {
	switch input {
	case types.PartyType_PARTY_TYPE_ORIGINATOR:
		return PartyTypeOriginator
	case types.PartyType_PARTY_TYPE_SERVICER:
		return PartyTypeServicer
	case types.PartyType_PARTY_TYPE_INVESTOR:
		return PartyTypeInvestor
	case types.PartyType_PARTY_TYPE_CUSTODIAN:
		return PartyTypeCustodian
	case types.PartyType_PARTY_TYPE_OWNER:
		return PartyTypeOwner
	case types.PartyType_PARTY_TYPE_AFFILIATE:
		return PartyTypeAffiliate
	case types.PartyType_PARTY_TYPE_OMNIBUS:
		return PartyTypeOmnibus
	case types.PartyType_PARTY_TYPE_PROVENANCE:
		return PartyTypeProvenance
	default:
		return PartyTypeUnspecified
	}
}

// Convert a process to its provwasm type.
func createProcess(input types.Process) (*Process, error) {
	address := input.GetAddress()
	hash := input.GetHash()
	processID := &ProcessID{}
	switch {
	case address != "":
		processID.Address = &ProcessIDAddress{Address: address}
	case hash != "":
		processID.Hash = &ProcessIDHash{Hash: hash}
	default:
		return nil, fmt.Errorf("wasm: address or hash must be defined for a process id")
	}
	return &Process{
		ProcessID: processID,
		Name:      input.Name,
		Method:    input.Method,
	}, nil
}

// Convert a source to its provwasm type.
func createSource(input types.RecordInput) (*RecordInputSource, error) {
	source := &RecordInputSource{}
	if s, ok := input.GetSource().(*types.RecordInput_Hash); ok {
		source.Hash = &RecordInputSourceHash{Hash: s.Hash}
		return source, nil
	}
	if s, ok := input.GetSource().(*types.RecordInput_RecordId); ok {
		recordID, err := stringifyAddress(s.RecordId)
		if err != nil {
			return nil, err
		}
		source.Record = &RecordInputSourceRecord{RecordID: recordID}
		return source, nil
	}
	return nil, fmt.Errorf("wasm: hash or record id must be defined for a source")
}

// Convert a record input status to its provwasm type.
func createRecordInputStatus(input types.RecordInputStatus) InputStatus {
	switch input {
	case types.RecordInputStatus_Proposed:
		return InputStatusProposed
	case types.RecordInputStatus_Record:
		return InputStatusRecord
	default:
		return InputStatusUnspecified
	}
}

// Convert a result status to its provwasm type.
func createResultStatus(input types.ResultStatus) ResultStatus {
	switch input {
	case types.ResultStatus_RESULT_STATUS_PASS:
		return ResultStatusPass
	case types.ResultStatus_RESULT_STATUS_FAIL:
		return ResultStatusFail
	case types.ResultStatus_RESULT_STATUS_SKIP:
		return ResultStatusSkip
	default:
		return ResultStatusUnspecified
	}
}
