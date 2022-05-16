package wasm

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"

	"github.com/provenance-io/provenance/x/metadata/types"
)

// These are the Go struct that map to the Rust smart contract types defined by the provwasm JSON schema.
// The types in this file were generated using quicktype.io and slightly modified.

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
	// PartyTypeController is a concrete party type.
	PartyTypeController PartyType = "controller"
	// PartyTypeValidator is a concrete party type.
	PartyTypeValidator PartyType = "validator"
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

// A slightly modified, non-panicing version of MetadataAddress.String(). Panics across FFI
// boundaries can crash the chain, so just fail the query.
func bech32Address(ma types.MetadataAddress) (string, error) {
	if ma.Empty() { // cause a query failure for addresses we expect to be non-empty.
		return "", fmt.Errorf("wasm: empty metadata address")
	}
	hrp, err := types.VerifyMetadataAddressFormat(ma)
	if err != nil {
		return "", fmt.Errorf("wasm: %w", err)
	}
	bech32Addr, err := bech32.ConvertAndEncode(hrp, ma.Bytes())
	if err != nil {
		return "", fmt.Errorf("wasm: %w", err)
	}
	return bech32Addr, nil
}

// Convert a provwasm scope into the baseType scope.
func (scope *Scope) convertToBaseType() (*types.Scope, error) {
	scopeID, err := types.MetadataAddressFromBech32(scope.ScopeID)
	if err != nil {
		return nil, fmt.Errorf("wasm: invalid 'scope id': %w", err)
	}
	specificationID, err := types.MetadataAddressFromBech32(scope.SpecificationID)
	if err != nil {
		return nil, fmt.Errorf("wasm: invalid 'specification id': %w", err)
	}
	// verify the data_access addresses are valid
	for _, addr := range scope.DataAccess {
		_, err = sdk.AccAddressFromBech32(addr)
		if err != nil {
			return nil, fmt.Errorf("wasm: invalid 'data_access' address: %v", err)
		}
	}
	baseType := &types.Scope{
		ScopeId:           scopeID,
		SpecificationId:   specificationID,
		Owners:            make([]types.Party, len(scope.Owners)),
		DataAccess:        scope.DataAccess,
		ValueOwnerAddress: scope.ValueOwnerAddress,
	}
	for i, o := range scope.Owners {
		party, err := o.convertToBaseType()
		if err != nil {
			return nil, err
		}
		baseType.Owners[i] = *party
	}

	return baseType, nil
}

// Convert a provwasm party into the baseType party.
func (party *Party) convertToBaseType() (*types.Party, error) {
	_, err := sdk.AccAddressFromBech32(party.Address)
	if err != nil {
		return nil, fmt.Errorf("wasm: invalid 'data_access' address: %v", err)
	}
	return &types.Party{
		Address: party.Address,
		Role:    party.Role.convertToBaseType(),
	}, nil
}

// Convert a provwasm partytype into the baseType partytype.
func (partyType *PartyType) convertToBaseType() types.PartyType {
	switch *partyType {
	case PartyTypeOriginator:
		return types.PartyType_PARTY_TYPE_ORIGINATOR
	case PartyTypeServicer:
		return types.PartyType_PARTY_TYPE_SERVICER
	case PartyTypeInvestor:
		return types.PartyType_PARTY_TYPE_INVESTOR
	case PartyTypeCustodian:
		return types.PartyType_PARTY_TYPE_CUSTODIAN
	case PartyTypeOwner:
		return types.PartyType_PARTY_TYPE_OWNER
	case PartyTypeAffiliate:
		return types.PartyType_PARTY_TYPE_AFFILIATE
	case PartyTypeOmnibus:
		return types.PartyType_PARTY_TYPE_OMNIBUS
	case PartyTypeProvenance:
		return types.PartyType_PARTY_TYPE_PROVENANCE
	case PartyTypeController:
		return types.PartyType_PARTY_TYPE_CONTROLLER
	case PartyTypeValidator:
		return types.PartyType_PARTY_TYPE_VALIDATOR
	default:
		return types.PartyType_PARTY_TYPE_UNSPECIFIED
	}
}

// Convert a scope into provwasm JSON format.
func createScopeResponse(baseType types.Scope) ([]byte, error) {
	scopeID, err := bech32Address(baseType.ScopeId)
	if err != nil {
		return nil, err
	}
	specificationID, err := bech32Address(baseType.SpecificationId)
	if err != nil {
		return nil, err
	}
	scope := &Scope{
		ScopeID:           scopeID,
		SpecificationID:   specificationID,
		ValueOwnerAddress: baseType.ValueOwnerAddress,
		DataAccess:        baseType.DataAccess,
		Owners:            make([]*Party, len(baseType.Owners)),
	}
	for i, o := range baseType.Owners {
		scope.Owners[i] = createParty(o)
	}
	bz, err := json.Marshal(scope)
	if err != nil {
		return nil, fmt.Errorf("wasm: marshal scope failed: %w", err)
	}
	return bz, nil
}

// Convert a slice of sessions into provwasm JSON format.
func createSessionsResponse(baseTypeSlice []types.Session) ([]byte, error) {
	sessions := &Sessions{
		Sessions: make([]*Session, len(baseTypeSlice)),
	}
	for i, baseType := range baseTypeSlice {
		session, err := createSession(baseType)
		if err != nil {
			return nil, err
		}
		sessions.Sessions[i] = session
	}
	bz, err := json.Marshal(sessions)
	if err != nil {
		return nil, fmt.Errorf("wasm: marshal sessions failed: %w", err)
	}
	return bz, nil
}

// Convert a session into its provwasm type.
func createSession(baseType types.Session) (*Session, error) {
	sessionID, err := bech32Address(baseType.SessionId)
	if err != nil {
		return nil, err
	}
	specificationID, err := bech32Address(baseType.SpecificationId)
	if err != nil {
		return nil, err
	}
	session := &Session{
		SessionID:       sessionID,
		SpecificationID: specificationID,
		Name:            baseType.Name,
		Context:         baseType.Context,
		Parties:         make([]*Party, len(baseType.Parties)),
	}
	for i, p := range baseType.Parties {
		session.Parties[i] = createParty(p)
	}
	return session, nil
}

// Convert a slice of records into provwasm JSON format.
func createRecordsResponse(baseTypeSlice []*types.Record) ([]byte, error) {
	records := &Records{
		Records: make([]*Record, len(baseTypeSlice)),
	}
	for i, baseType := range baseTypeSlice {
		record, err := createRecord(baseType)
		if err != nil {
			return nil, err
		}
		records.Records[i] = record
	}
	bz, err := json.Marshal(records)
	if err != nil {
		return nil, fmt.Errorf("wasm: marshal records failed: %w", err)
	}
	return bz, nil
}

// Convert a record into its provwasm type.
func createRecord(baseType *types.Record) (*Record, error) {
	sessionID, err := bech32Address(baseType.SessionId)
	if err != nil {
		return nil, err
	}
	specID, err := bech32Address(baseType.SpecificationId)
	if err != nil {
		return nil, err
	}
	process, err := createProcess(baseType.Process)
	if err != nil {
		return nil, err
	}
	record := &Record{
		SessionID:       sessionID,
		SpecificationID: specID,
		Name:            baseType.Name,
		Process:         process,
		Inputs:          make([]*RecordInput, len(baseType.Inputs)),
		Outputs:         make([]*RecordOutput, len(baseType.Outputs)),
	}
	for i, in := range baseType.Inputs {
		input, err := createRecordInput(in)
		if err != nil {
			return nil, err
		}
		record.Inputs[i] = input
	}
	for i, out := range baseType.Outputs {
		record.Outputs[i] = &RecordOutput{
			Hash:   out.Hash,
			Status: createResultStatus(out.Status),
		}
	}
	return record, nil
}

// Convert a process to its provwasm type.
func createProcess(baseType types.Process) (*Process, error) {
	address := baseType.GetAddress()
	hash := baseType.GetHash()
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
		Name:      baseType.Name,
		Method:    baseType.Method,
	}, nil
}

// Convert a record input into its provwasm type.
func createRecordInput(baseType types.RecordInput) (*RecordInput, error) {
	source, err := createSource(baseType)
	if err != nil {
		return nil, err
	}
	return &RecordInput{
		Name:     baseType.Name,
		TypeName: baseType.TypeName,
		Source:   source,
		Status:   createRecordInputStatus(baseType.Status),
	}, nil
}

// Convert a record input source to its provwasm type.
func createSource(baseType types.RecordInput) (*RecordInputSource, error) {
	source := &RecordInputSource{}
	if s, ok := baseType.GetSource().(*types.RecordInput_Hash); ok {
		source.Hash = &RecordInputSourceHash{Hash: s.Hash}
		return source, nil
	}
	if s, ok := baseType.GetSource().(*types.RecordInput_RecordId); ok {
		recordID, err := bech32Address(s.RecordId)
		if err != nil {
			return nil, err
		}
		source.Record = &RecordInputSourceRecord{RecordID: recordID}
		return source, nil
	}
	return nil, fmt.Errorf("wasm: hash or record id must be defined for a source")
}

// Convert a party to its provwasm type.
func createParty(baseType types.Party) *Party {
	return &Party{
		Address: baseType.Address,
		Role:    createRole(baseType.Role),
	}
}

// Convert a party type to its provwasm type.
func createRole(baseType types.PartyType) PartyType {
	switch baseType {
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
	case types.PartyType_PARTY_TYPE_CONTROLLER:
		return PartyTypeController
	case types.PartyType_PARTY_TYPE_VALIDATOR:
		return PartyTypeValidator
	default:
		return PartyTypeUnspecified
	}
}

// Convert a record input status to its provwasm type.
func createRecordInputStatus(baseType types.RecordInputStatus) InputStatus {
	switch baseType {
	case types.RecordInputStatus_Proposed:
		return InputStatusProposed
	case types.RecordInputStatus_Record:
		return InputStatusRecord
	default:
		return InputStatusUnspecified
	}
}

// Convert a result status to its provwasm type.
func createResultStatus(baseType types.ResultStatus) ResultStatus {
	switch baseType {
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
