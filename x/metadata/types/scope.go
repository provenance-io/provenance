package types

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// A sane default for maximum length of an audit message string (memo)
	maxAuditMessageLength = 200
)

// NewScope creates a new instance.
func NewScope(
	scopeID, scopeSpecification MetadataAddress,
	owners, dataAccess []string,
	valueOwner string,
) *Scope {
	return &Scope{
		ScopeId:           scopeID,
		SpecificationId:   scopeSpecification,
		OwnerAddress:      owners,
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
	if len(s.OwnerAddress) < 1 {
		return errors.New("scope must have at least one owner")
	}
	for _, o := range s.OwnerAddress {
		if _, err = sdk.AccAddressFromBech32(o); err != nil {
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

// NewRecordGroup creates a new instance
func NewRecordGroup(name string, groupID, groupSpecification MetadataAddress, parties []Party) *RecordGroup {
	return &RecordGroup{
		GroupId:         groupID,
		SpecificationId: groupSpecification,
		Parties:         parties,
		Name:            name,
		Audit:           AuditFields{},
	}
}

// ValidateBasic performs basic format checking of data within a scope
func (rg *RecordGroup) ValidateBasic() error {
	prefix, err := VerifyMetadataAddressFormat(rg.GroupId)
	if err != nil {
		return err
	}
	if prefix != PrefixGroup {
		return fmt.Errorf("invalid group identifier (expected: %s, got %s)", PrefixGroup, prefix)
	}
	if len(rg.Parties) < 1 {
		return errors.New("record group must have at least one party")
	}
	for _, p := range rg.Parties {
		if err = p.ValidateBasic(); err != nil {
			return fmt.Errorf("invalid party on record group: %w", err)
		}
	}
	prefix, err = VerifyMetadataAddressFormat(rg.SpecificationId)
	if err != nil {
		return err
	}
	if prefix != PrefixGroupSpecification {
		return fmt.Errorf("invalid group specification identifier (expected: %s, got %s)", PrefixGroupSpecification, prefix)
	}
	if len(rg.Name) == 0 {
		return errors.New("record group name can not be empty")
	}
	if len(rg.Audit.CreatedBy) > 0 {
		if _, err := sdk.AccAddressFromBech32(rg.Audit.CreatedBy); err != nil {
			return fmt.Errorf("invalid record group audit:createdby %w", err)
		}
		if rg.Audit.CreatedDate.IsZero() {
			return fmt.Errorf("invalid/null record group audit created date")
		}
	}
	if len(rg.Audit.UpdatedBy) > 0 {
		if _, err := sdk.AccAddressFromBech32(rg.Audit.UpdatedBy); err != nil {
			return fmt.Errorf("invalid record group audit:updatedby %w", err)
		}
		if rg.Audit.UpdatedDate.IsZero() {
			return fmt.Errorf("invalid/null record group audit updated date")
		}
	}
	if len(rg.Audit.Message) > maxAuditMessageLength {
		return fmt.Errorf("record group audit message exceeds maximum length (expected < %d got: %d",
			maxAuditMessageLength, len(rg.Audit.Message))
	}
	return nil
}

// String implements stringer interface
func (rg RecordGroup) String() string {
	out := fmt.Sprintf("%s (%s) [", rg.Name, rg.GroupId)
	for _, p := range rg.Parties {
		out += fmt.Sprintf("%s - %s, ", p.Address, p.Role)
	}
	out = strings.TrimRight(out, ", ")
	out += fmt.Sprintf("] (%s)", rg.SpecificationId)
	return out
}

func (r Record) NewRecord(name string, groupId MetadataAddress, process Process, inputs []RecordInput, outputs []RecordOutput) *Record {
	return &Record{
		Name:    name,
		GroupId: groupId,
		Process: process,
		Inputs:  inputs,
		Outputs: outputs,
	}
}

// ValidateBasic performs static checking of Record format
func (r Record) ValidateBasic() error {
	prefix, err := VerifyMetadataAddressFormat(r.GroupId)
	if err != nil {
		return err
	}
	for _, i := range r.Inputs {
		if err = i.ValidateBasic(); err != nil {
			return fmt.Errorf("invalid record input: %w", err)
		}
	}
	if prefix != PrefixGroup {
		return fmt.Errorf("invalid group identifier (expected: %s, got %s)", PrefixGroup, prefix)
	}
	if len(r.Name) < 1 {
		return fmt.Errorf("invalid/missing name for record")
	}
	return nil
}

// String implements stringer interface
func (r Record) String() string {
	out := fmt.Sprintf("%s (%s) Results [", r.Name, r.GroupId)
	for _, o := range r.Outputs {
		out += fmt.Sprintf("%s - %s, ", o.Status, o.Hash)
	}
	out = strings.TrimRight(out, ", ")
	out += fmt.Sprintf("] (%s/%s)", r.Process.Name, r.Process.Method)
	return out
}

// ValidateBasic performs a static check over the record input format
func (ri RecordInput) ValidateBasic() error {
	if len(ri.Name) < 1 {
		return fmt.Errorf("missing required name")
	}
	if ri.Status == RecordInputStatus_Unknown {
		return fmt.Errorf("invalid record input status, status unknown or missing")
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

// ValidateBasic performs a static check over the record output format
func (ro RecordOutput) ValidateBasic() error {
	if len(ro.Hash) < 1 {
		return fmt.Errorf("missing required hash")
	}
	if ro.Status == ResultStatus_Unspecified {
		return fmt.Errorf("invalid record output status, status unspecified")
	}
	return nil
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
