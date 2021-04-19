package types

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// A sane default for maximum length of an audit message string (memo)
	maxAuditMessageLength = 200
)

// NewScope creates a new instance.
func NewScope(
	scopeID,
	scopeSpecification MetadataAddress,
	owners []Party,
	dataAccess []string,
	valueOwner string,
) *Scope {
	return &Scope{
		ScopeId:           scopeID,
		SpecificationId:   scopeSpecification,
		Owners:            owners,
		DataAccess:        dataAccess,
		ValueOwnerAddress: valueOwner,
	}
}

// ValidateBasic performs basic format checking of data within a scope
func (s *Scope) ValidateBasic() error {
	prefix, err := VerifyMetadataAddressFormat(s.ScopeId)
	if err != nil {
		return err
	}
	if prefix != PrefixScope {
		return fmt.Errorf("invalid scope identifier (expected: %s, got %s)", PrefixScope, prefix)
	}
	if !s.SpecificationId.Empty() {
		prefix, err = VerifyMetadataAddressFormat(s.SpecificationId)
		if err != nil {
			return err
		}
		if prefix != PrefixScopeSpecification {
			return fmt.Errorf("invalid scope specification identifier (expected: %s, got %s)", PrefixScopeSpecification, prefix)
		}
	}
	if len(s.Owners) < 1 {
		return errors.New("scope must have at least one owner")
	}
	for _, o := range s.Owners {
		if _, err = sdk.AccAddressFromBech32(o.Address); err != nil {
			return fmt.Errorf("invalid owner on scope: %w", err)
		}
	}
	for _, d := range s.DataAccess {
		if _, err = sdk.AccAddressFromBech32(d); err != nil {
			return fmt.Errorf("invalid address in data access on scope: %w", err)
		}
	}
	if len(s.ValueOwnerAddress) > 0 {
		if _, err = sdk.AccAddressFromBech32(s.ValueOwnerAddress); err != nil {
			return fmt.Errorf("invalid value owner address on scope: %w", err)
		}
	}
	return nil
}

// String implements stringer interface
func (s Scope) String() string {
	out, _ := yaml.Marshal(s)
	return string(out)
}

// UpdateAudit computes a set of changes to the audit fields based on the existing message.
func (a *AuditFields) UpdateAudit(blocktime time.Time, signers, message string) *AuditFields {
	if a == nil {
		return &AuditFields{
			Version:     1,
			CreatedDate: blocktime,
			CreatedBy:   signers,
			Message:     message,
		}
	}
	return &AuditFields{
		Version:     a.Version + 1,
		CreatedDate: a.CreatedDate,
		CreatedBy:   a.CreatedBy,
		UpdatedDate: blocktime,
		UpdatedBy:   signers,
		Message:     message,
	}
}

// NewSession creates a new instance
func NewSession(name string, sessionID, contractSpecification MetadataAddress, parties []Party, auditFields *AuditFields) *Session {
	return &Session{
		SessionId:       sessionID,
		SpecificationId: contractSpecification,
		Parties:         parties,
		Name:            name,
		Audit:           auditFields,
	}
}

// ValidateBasic performs basic format checking of data within a scope
func (s *Session) ValidateBasic() error {
	prefix, err := VerifyMetadataAddressFormat(s.SessionId)
	if err != nil {
		return err
	}
	if prefix != PrefixSession {
		return fmt.Errorf("invalid session identifier (expected: %s, got %s)", PrefixSession, prefix)
	}
	if len(s.Parties) < 1 {
		return errors.New("session must have at least one party")
	}
	for _, p := range s.Parties {
		if err = p.ValidateBasic(); err != nil {
			return fmt.Errorf("invalid party on session: %w", err)
		}
	}
	prefix, err = VerifyMetadataAddressFormat(s.SpecificationId)
	if err != nil {
		return err
	}
	if prefix != PrefixContractSpecification {
		return fmt.Errorf("invalid contract specification identifier (expected: %s, got %s)", PrefixContractSpecification, prefix)
	}
	if s.Audit != nil && len(s.Audit.Message) > maxAuditMessageLength {
		return fmt.Errorf("session audit message exceeds maximum length (expected < %d got: %d)",
			maxAuditMessageLength, len(s.Audit.Message))
	}
	return nil
}

// String implements stringer interface
func (s Session) String() string {
	out, _ := yaml.Marshal(s)
	return string(out)
}

// NewRecord creates new instance of Record
func NewRecord(
	name string,
	sessionID MetadataAddress,
	process Process,
	inputs []RecordInput,
	outputs []RecordOutput,
	specificationID MetadataAddress,
) *Record {
	return &Record{
		Name:            name,
		SessionId:       sessionID,
		Process:         process,
		Inputs:          inputs,
		Outputs:         outputs,
		SpecificationId: specificationID,
	}
}

// ValidateBasic performs static checking of Record format
func (r Record) ValidateBasic() error {
	prefix, err := VerifyMetadataAddressFormat(r.SessionId)
	if err != nil {
		return err
	}
	if prefix != PrefixSession {
		return fmt.Errorf("invalid record identifier (expected: %s, got %s)", PrefixSession, prefix)
	}
	if !r.SpecificationId.Empty() {
		// For now, we'll allow an empty specification id and set it appropriately during ValidateRecordUpdate if it's missing.
		// But if we've got it, we should make sure it's okay.
		specPrefix, e := VerifyMetadataAddressFormat(r.SpecificationId)
		if e != nil {
			return e
		}
		if specPrefix != PrefixRecordSpecification {
			return fmt.Errorf("invalid record specification identifier (expected: %s, got %s)", PrefixRecordSpecification, specPrefix)
		}
	}
	for _, i := range r.Inputs {
		if err = i.ValidateBasic(); err != nil {
			return fmt.Errorf("invalid record input: %w", err)
		}
	}
	for _, o := range r.Outputs {
		if err = o.ValidateBasic(); err != nil {
			return fmt.Errorf("invalid record output: %w", err)
		}
	}
	if len(r.Name) < 1 {
		return fmt.Errorf("invalid/missing name for record")
	}
	if err = r.Process.ValidateBasic(); err != nil {
		return fmt.Errorf("invalid record process: %w", err)
	}
	return nil
}

// String implements stringer interface
func (r Record) String() string {
	out := fmt.Sprintf("%s (%s) Results [", r.Name, r.SessionId)
	for _, o := range r.Outputs {
		out += fmt.Sprintf("%s - %s, ", o.Status, o.Hash)
	}
	out = strings.TrimRight(out, ", ")
	out += fmt.Sprintf("] (%s/%s)", r.Process.Name, r.Process.Method)
	return out
}

// GetRecordAddress returns the address for this record, or an empty MetadataAddress if it cannot be constructed.
func (r Record) GetRecordAddress() MetadataAddress {
	addr, err := r.SessionId.AsRecordAddress(r.Name)
	if err == nil {
		return addr
	}
	return MetadataAddress{}
}

// NewRecordInput creates new instance of RecordInput
func NewRecordInput(name string, source isRecordInput_Source, typeName string, status RecordInputStatus) *RecordInput {
	return &RecordInput{
		Name:     name,
		Source:   source,
		TypeName: typeName,
		Status:   status,
	}
}

// ValidateBasic performs a static check over the record input format
func (ri RecordInput) ValidateBasic() error {
	if len(ri.Name) < 1 {
		return fmt.Errorf("missing required name")
	}
	if ri.Status == RecordInputStatus_Unknown {
		return fmt.Errorf("invalid record input status, status unknown or missing")
	}
	if ri.Source == nil {
		return fmt.Errorf("missing required record input source")
	}
	switch source := ri.Source.(type) {
	case *RecordInput_Hash:
		if ri.Status != RecordInputStatus_Proposed {
			return fmt.Errorf("hash specifier only applies to proposed inputs")
		}
		if len(source.Hash) < 1 {
			return fmt.Errorf("missing required hash for proposed value")
		}
	case *RecordInput_RecordId:
		if ri.Status != RecordInputStatus_Record {
			return fmt.Errorf("record id must be used with Record type inputs")
		}
		prefix, err := VerifyMetadataAddressFormat(source.RecordId)
		if err != nil {
			return fmt.Errorf("invalid record input recordid %w", err)
		}
		if prefix != PrefixRecord {
			return fmt.Errorf("invalid record id address (found %s, expected record)", prefix)
		}
	}
	if len(ri.TypeName) < 1 {
		return fmt.Errorf("missing type name")
	}
	return nil
}

// String implements stringer interface
func (ri RecordInput) String() string {
	out := fmt.Sprintf("%s (%s) - %s ", ri.Name, ri.TypeName, ri.Status)
	switch source := ri.Source.(type) {
	case *RecordInput_Hash:
		out += source.Hash
	case *RecordInput_RecordId:
		out += source.RecordId.String()
	}
	return out
}

// NewRecordOutput creates a new instance of RecordOutput
func NewRecordOutput(hash string, status ResultStatus) *RecordOutput {
	return &RecordOutput{
		Hash:   hash,
		Status: status,
	}
}

// ValidateBasic performs a static check over the record output format
func (ro RecordOutput) ValidateBasic() error {
	if ro.Status == ResultStatus_RESULT_STATUS_SKIP {
		return nil
	}
	if ro.Status == ResultStatus_RESULT_STATUS_UNSPECIFIED {
		return fmt.Errorf("invalid record output status, status unspecified")
	}
	if len(ro.Hash) < 1 {
		return fmt.Errorf("missing required hash")
	}
	return nil
}

// String implements stringer interface
func (ro RecordOutput) String() string {
	return fmt.Sprintf("%s - %s", ro.Hash, ro.Status)
}

// NewProcess creates a new instance of Process
func NewProcess(name string, processID isProcess_ProcessId, method string) *Process {
	return &Process{
		Name:      name,
		ProcessId: processID,
		Method:    method,
	}
}

// ProcessID is a publicly exposed isProcess_ProcessId
type ProcessID isProcess_ProcessId

// ValidateBasic performs a static check over the process format
func (ps Process) ValidateBasic() error {
	if len(ps.Method) < 1 {
		return fmt.Errorf("missing required method")
	}
	if len(ps.Name) < 1 {
		return fmt.Errorf("missing required name")
	}
	if ps.ProcessId == nil {
		return fmt.Errorf("missing required process id")
	}
	return nil
}

// String implements stringer interface
func (ps Process) String() string {
	return fmt.Sprintf("%s - %s - %s", ps.Name, ps.Method, ps.ProcessId)
}

// ValidateBasic performs static checking of Party format
func (p Party) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(p.Address); err != nil {
		return fmt.Errorf("invalid address: %w", err)
	}
	if p.Role == PartyType_PARTY_TYPE_UNSPECIFIED {
		return fmt.Errorf("invalid party type;  party type not specified")
	}
	return nil
}

// String implements stringer interface
func (p Party) String() string {
	return fmt.Sprintf("%s - %s", p.Address, p.Role)
}
