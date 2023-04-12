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
	requirePartyRollup bool,
) *Scope {
	return &Scope{
		ScopeId:            scopeID,
		SpecificationId:    scopeSpecification,
		Owners:             owners,
		DataAccess:         dataAccess,
		ValueOwnerAddress:  valueOwner,
		RequirePartyRollup: requirePartyRollup,
	}
}

func (s Scope) Equals(t Scope) bool {
	return s.ScopeId.Equals(t.ScopeId) &&
		s.SpecificationId.Equals(t.SpecificationId) &&
		EqualParties(s.Owners, t.Owners) &&
		equivalentDataAssessors(s.DataAccess, t.DataAccess) &&
		s.ValueOwnerAddress == t.ValueOwnerAddress &&
		s.RequirePartyRollup == t.RequirePartyRollup
}

// ValidateBasic performs basic format checking of data within a scope
func (s Scope) ValidateBasic() error {
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
	if err = s.ValidateOwnersBasic(); err != nil {
		return err
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

func (s Scope) ValidateOwnersBasic() error {
	if err := ValidatePartiesBasic(s.Owners); err != nil {
		return fmt.Errorf("invalid scope owners: %w", err)
	}
	if err := ValidateOptionalParties(s.RequirePartyRollup, s.Owners); err != nil {
		return err
	}
	return nil
}

// String implements stringer interface
func (s Scope) String() string {
	out, _ := yaml.Marshal(s)
	return string(out)
}

func (s *Scope) RemoveDataAccess(addresses []string) {
	newDataAccess := []string{}
	for _, da := range s.DataAccess {
		found := false
		for _, addr := range addresses {
			if addr == da {
				found = true
				break
			}
		}
		if !found {
			newDataAccess = append(newDataAccess, da)
		}
	}

	s.DataAccess = newDataAccess
}

func (s *Scope) AddDataAccess(addresses []string) {
	for _, addr := range addresses {
		found := false
		for _, da := range s.DataAccess {
			if addr == da {
				found = true
				break
			}
		}
		if !found {
			s.DataAccess = append(s.DataAccess, addr)
		}
	}
}

// AddOwners will add new owners to this scope.
// If an owner already exists that's equal to a provided owner, an error is returned.
func (s *Scope) AddOwners(owners []Party) error {
	if len(owners) == 0 {
		return nil
	}
	for _, newOwner := range owners {
		for _, scopeOwner := range s.Owners {
			//nolint:gosec // G601: Implicit memory aliasing okay here since we're not storing the reference anywhere.
			if newOwner.IsSameAs(&scopeOwner) {
				return fmt.Errorf("party already exists with address %s and role %s", newOwner.Address, newOwner.Role)
			}
		}
	}
	s.Owners = append(s.Owners, owners...)
	return nil
}

// RemoveOwners will remove owners with the given addresses.
// If an address is provided that is not an owner, an error is returned.
func (s *Scope) RemoveOwners(addressesToRemove []string) error {
	if len(addressesToRemove) == 0 {
		return nil
	}
	for _, addr := range addressesToRemove {
		if !s.hasOwnerAddress(addr) {
			return fmt.Errorf("address does not exist in scope owners: %s", addr)
		}
	}
	newOwners := make([]Party, 0, len(s.Owners))
ownersLoop:
	for _, owner := range s.Owners {
		for _, addr := range addressesToRemove {
			if owner.Address == addr {
				continue ownersLoop
			}
		}
		newOwners = append(newOwners, owner)
	}
	s.Owners = newOwners
	return nil
}

// hasOwnerAddress returns true if this scope has an owner party with the provided address.
func (s *Scope) hasOwnerAddress(address string) bool {
	for _, party := range s.Owners {
		if address == party.Address {
			return true
		}
	}
	return false
}

// GetAllOwnerAddresses gets the addresses of all of the owners. Each address can only appear once in the return value.
func (s Scope) GetAllOwnerAddresses() []string {
	return GetPartyAddresses(s.Owners)
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
func (s Session) ValidateBasic() error {
	prefix, err := VerifyMetadataAddressFormat(s.SessionId)
	if err != nil {
		return err
	}
	if prefix != PrefixSession {
		return fmt.Errorf("invalid session identifier (expected: %s, got %s)", PrefixSession, prefix)
	}
	err = ValidatePartiesBasic(s.Parties)
	if err != nil {
		return err
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

// GetAllPartyAddresses gets the addresses of all of the parties. Each address can only appear once in the return value.
func (s Session) GetAllPartyAddresses() []string {
	return GetPartyAddresses(s.Parties)
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
		// For now, we'll allow an empty specification id and set it appropriately during ValidateWriteRecord if it's missing.
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
	if len(p.Address) == 0 {
		return errors.New("missing party address")
	}
	if _, err := sdk.AccAddressFromBech32(p.Address); err != nil {
		return fmt.Errorf("invalid party address [%s]: %w", p.Address, err)
	}
	if !p.Role.IsValid() || p.Role == PartyType_PARTY_TYPE_UNSPECIFIED {
		return fmt.Errorf("invalid party type for party %s", p.Address)
	}
	return nil
}

// ValidatePartiesAreUnique makes sure that no two provided parties are the same.
func ValidatePartiesAreUnique(parties []Party) error {
	for i := 0; i < len(parties)-1; i++ {
		for j := i + 1; j < len(parties); j++ {
			if parties[i].IsSameAs(&parties[j]) {
				return fmt.Errorf("duplicate parties not allowed: address = %s, role = %s, indexes: %d, %d",
					parties[i].Address, parties[i].Role.SimpleString(), i, j)
			}
		}
	}
	return nil
}

// ValidatePartiesBasic validates a required list of parties.
func ValidatePartiesBasic(parties []Party) error {
	if len(parties) < 1 {
		return errors.New("at least one party is required")
	}
	for _, p := range parties {
		if err := p.ValidateBasic(); err != nil {
			return err
		}
	}
	return ValidatePartiesAreUnique(parties)
}

// ValidateOptionalParties validates that, if optional parties aren't allowed, none are provided.
func ValidateOptionalParties(optAllowed bool, parties []Party) error {
	if !optAllowed {
		for _, party := range parties {
			if party.Optional {
				return fmt.Errorf("parties can only be optional when require_party_rollup = true")
			}
		}
	}
	return nil
}

// String implements stringer interface
func (p Party) String() string {
	return fmt.Sprintf("%s - %s", p.Address, p.Role)
}

// Partier is an interface with the getter methods of a Party.
type Partier interface {
	GetAddress() string
	GetRole() PartyType
	GetOptional() bool
}

// EqualPartiers returns true if p1 and p2 are equivalent.
func EqualPartiers(p1, p2 Partier) bool {
	if p1 == p2 {
		return true
	}
	if p1 == nil || p2 == nil {
		return false
	}
	return p1.GetAddress() == p2.GetAddress() && p1.GetRole() == p2.GetRole() && p1.GetOptional() == p2.GetOptional()
}

// SamePartiers returns true if p1 and p2 are have the same address and role.
func SamePartiers(p1, p2 Partier) bool {
	if p1 == p2 {
		return true
	}
	if p1 == nil || p2 == nil {
		return false
	}
	return p1.GetAddress() == p2.GetAddress() && p1.GetRole() == p2.GetRole()
}

// Equals returns true if this party is equal to the provided party.
// See also: IsSameAs for a comparison that ignores the Optional field.
func (p Party) Equals(p2 Partier) bool {
	return EqualPartiers(&p, p2)
}

// IsSameAs returns true if this party's address and role are the same as the provided party's.
// See also: Equals for a more thorough comparison.
func (p Party) IsSameAs(p2 Partier) bool {
	return SamePartiers(&p, p2)
}

// EqualParties returns true if the two provided sets of parties contain the same entries.
// This assumes that duplicates are not allowed in a set of parties.
func EqualParties(p1, p2 []Party) bool {
	if len(p1) != len(p2) {
		return false
	}
p1Loop:
	for _, p1p := range p1 {
		for _, p2p := range p2 {
			//nolint:gosec // G601: Implicit memory aliasing okay here since we're not storing the reference anywhere.
			if p1p.Equals(&p2p) {
				continue p1Loop
			}
		}
		return false
	}
	return true
}

// GetPartyAddresses gets the addresses of all of the parties. Each address can only appear once in the return value.
func GetPartyAddresses(parties []Party) []string {
	var rv []string
	have := make(map[string]bool)
	for _, party := range parties {
		if !have[party.Address] {
			rv = append(rv, party.Address)
			have[party.Address] = true
		}
	}
	return rv
}

// GetRequiredPartyAddresses gets the addresses of all the required parties.
// Each address can only appear once in the return value.
func GetRequiredPartyAddresses(parties []Party) []string {
	var req []Party
	for _, party := range parties {
		if !party.Optional {
			req = append(req, party)
		}
	}
	return GetPartyAddresses(req)
}

// equivalentDataAssessors returns true if all the entries in s1 are in s2, and vice versa.
func equivalentDataAssessors(s1, s2 []string) bool {
s1Loop:
	for _, s1s := range s1 {
		for _, s2s := range s2 {
			if s1s == s2s {
				continue s1Loop
			}
		}
		return false
	}
s2Loop:
	for _, s2s := range s2 {
		for _, s1s := range s1 {
			if s1s == s2s {
				continue s2Loop
			}
		}
		return false
	}
	return true
}
