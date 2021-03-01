package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	yaml "gopkg.in/yaml.v2"
)

const (
	TypeMsgMemorializeContractRequest      = "memorialize_contract_request"
	TypeMsgChangeOwnershipRequest          = "change_ownership_request"
	TypeMsgAddScopeRequest                 = "add_scope_request"
	TypeMsgDeleteScopeRequest              = "remove_scope_request"
	TypeMsgAddRecordGroupRequest           = "add_recordgroup_request"
	TypeMsgAddRecordRequest                = "add_record_request"
	TypeMsgAddScopeSpecificationRequest    = "add_scope_specification_request"
	TypeMsgRemoveScopeSpecificationRequest = "remove_scope_specification_request"
	TypeMsgAddContractSpecificationRequest = "add_contract_specification_request"
)

// Compile time interface checks.
var (
	_ sdk.Msg = &MsgMemorializeContractRequest{}
	_ sdk.Msg = &MsgChangeOwnershipRequest{}
	_ sdk.Msg = &MsgAddScopeRequest{}
	_ sdk.Msg = &MsgDeleteScopeRequest{}
	_ sdk.Msg = &MsgAddRecordGroupRequest{}
	_ sdk.Msg = &MsgAddRecordRequest{}
	_ sdk.Msg = &MsgAddScopeSpecificationRequest{}
	_ sdk.Msg = &MsgAddContractSpecificationRequest{}
)

// private method to convert an array of strings into an array of Acc Addresses.
func stringsToAccAddresses(strings []string) []sdk.AccAddress {
	retval := make([]sdk.AccAddress, len(strings))

	for i, str := range strings {
		accAddress, err := sdk.AccAddressFromBech32(str)
		if err != nil {
			panic(err)
		}
		retval[i] = accAddress
	}

	return retval
}

// ------------------  MsgMemorializeContractRequest  ------------------

// NewMsgMemorializeContractRequest creates a new msg instance
func NewMsgMemorializeContractRequest() *MsgMemorializeContractRequest {
	return &MsgMemorializeContractRequest{}
}

func (msg MsgMemorializeContractRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// Route returns the module route
func (msg MsgMemorializeContractRequest) Route() string {
	return ModuleName
}

// Type returns the type name for this msg
func (msg MsgMemorializeContractRequest) Type() string {
	return TypeMsgMemorializeContractRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgMemorializeContractRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Notary)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgMemorializeContractRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

//
// ValidateBasic quick validity check
func (msg MsgMemorializeContractRequest) ValidateBasic() error {
	if strings.TrimSpace(msg.ScopeId) == "" {
		return fmt.Errorf("scope ID is empty")
	}
	if strings.TrimSpace(msg.GroupId) == "" {
		return fmt.Errorf("group ID is empty")
	}
	if strings.TrimSpace(msg.ExecutionId) == "" {
		return fmt.Errorf("execution ID is empty")
	}
	if strings.TrimSpace(msg.Notary) == "" {
		return fmt.Errorf("notary address is empty")
	}
	if err := msg.Contract.ValidateBasic(); err != nil {
		return err
	}

	return msg.Contract.ValidateBasic()
}

// ------------------  MsgChangeOwnershipRequest  ------------------

// NewMsgChangeOwnershipRequest creates a new msg instance
func NewMsgChangeOwnershipRequest() *MsgChangeOwnershipRequest {
	return &MsgChangeOwnershipRequest{}
}

func (msg MsgChangeOwnershipRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// Route returns the module route
func (msg MsgChangeOwnershipRequest) Route() string {
	return ModuleName
}

// Type returns the type name for this msg
func (msg MsgChangeOwnershipRequest) Type() string {
	return TypeMsgChangeOwnershipRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgChangeOwnershipRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Notary)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgChangeOwnershipRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic performs a quick validity check
func (msg MsgChangeOwnershipRequest) ValidateBasic() error {
	if strings.TrimSpace(msg.ScopeId) == "" {
		return fmt.Errorf("scope ID is empty")
	}
	if strings.TrimSpace(msg.GroupId) == "" {
		return fmt.Errorf("group ID is empty")
	}
	if strings.TrimSpace(msg.ExecutionId) == "" {
		return fmt.Errorf("execution ID is empty")
	}
	if strings.TrimSpace(msg.Notary) == "" {
		return fmt.Errorf("notary address is empty")
	}

	// Must have one of contract, recitals but not both.
	if msg.Contract == nil && msg.Recitals == nil {
		return fmt.Errorf("one of contract or recitals is required")
	}
	if msg.Contract != nil && msg.Recitals != nil {
		return fmt.Errorf("only one of contract or recitals is allowed")
	}
	return msg.Contract.ValidateBasic()
}

// ------------------  MsgAddScopeRequest  ------------------

// NewMsgAddScopeRequest creates a new msg instance
func NewMsgAddScopeRequest(scope Scope, signers []string) *MsgAddScopeRequest {
	return &MsgAddScopeRequest{
		Scope:   scope,
		Signers: signers,
	}
}

func (msg MsgAddScopeRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// Route returns the module route
func (msg MsgAddScopeRequest) Route() string {
	return ModuleName
}

// Type returns the type name for this msg
func (msg MsgAddScopeRequest) Type() string {
	return TypeMsgAddScopeRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgAddScopeRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgAddScopeRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic performs a quick validity check
func (msg MsgAddScopeRequest) ValidateBasic() error {
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	return msg.Scope.ValidateBasic()
}

// ------------------  MsgRemoveScopeRequest  ------------------

// NewMsgDeleteScopeRequest creates a new msg instance
func NewMsgDeleteScopeRequest(scopeID MetadataAddress, signers []string) *MsgDeleteScopeRequest {
	return &MsgDeleteScopeRequest{
		ScopeId: scopeID,
		Signers: signers,
	}
}

func (msg MsgDeleteScopeRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// Route returns the module route
func (msg MsgDeleteScopeRequest) Route() string {
	return ModuleName
}

// Type returns the type name for this msg
func (msg MsgDeleteScopeRequest) Type() string {
	return TypeMsgDeleteScopeRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgDeleteScopeRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgDeleteScopeRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic performs a quick validity check
func (msg MsgDeleteScopeRequest) ValidateBasic() error {
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	if !msg.ScopeId.IsScopeAddress() {
		return fmt.Errorf("invalid scope address")
	}
	return nil
}

// ------------------  MsgAddRecordGroupRequest  ------------------

// NewMsgAddRecordGroupRequest creates a new msg instance
func NewMsgAddRecordGroupRequest() *MsgAddRecordGroupRequest {
	return &MsgAddRecordGroupRequest{}
}

func (msg MsgAddRecordGroupRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// Route returns the module route
func (msg MsgAddRecordGroupRequest) Route() string {
	return ModuleName
}

// Type returns the type name for this msg
func (msg MsgAddRecordGroupRequest) Type() string {
	return TypeMsgAddRecordGroupRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgAddRecordGroupRequest) GetSigners() []sdk.AccAddress {
	delAddr, err := sdk.AccAddressFromBech32(msg.Notary)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{delAddr}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgAddRecordGroupRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic performs a quick validity check
func (msg MsgAddRecordGroupRequest) ValidateBasic() error {
	return nil
}

// ------------------  MsgAddRecordRequest  ------------------

// NewMsgAddRecordRequest creates a new msg instance
func NewMsgAddRecordRequest() *MsgAddRecordRequest {
	return &MsgAddRecordRequest{}
}

func (msg MsgAddRecordRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// Route returns the module route
func (msg MsgAddRecordRequest) Route() string {
	return ModuleName
}

// Type returns the type name for this msg
func (msg MsgAddRecordRequest) Type() string {
	return TypeMsgAddRecordRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgAddRecordRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgAddRecordRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic performs a quick validity check
func (msg MsgAddRecordRequest) ValidateBasic() error {
	return nil
}

// ------------------  MsgAddScopeSpecificationRequest  ------------------

// NewMsgAddScopeSpecificationRequest creates a new msg instance
func NewMsgAddScopeSpecificationRequest() *MsgAddScopeSpecificationRequest {
	return &MsgAddScopeSpecificationRequest{}
}

func (msg MsgAddScopeSpecificationRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// Route returns the module route
func (msg MsgAddScopeSpecificationRequest) Route() string {
	return ModuleName
}

// Type returns the type name for this msg
func (msg MsgAddScopeSpecificationRequest) Type() string { return TypeMsgAddScopeSpecificationRequest }

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgAddScopeSpecificationRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgAddScopeSpecificationRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic performs a quick validity check
func (msg MsgAddScopeSpecificationRequest) ValidateBasic() error {
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	return msg.Specification.ValidateBasic()
}

// ------------------  MsgDeleteScopeSpecificationRequest  ------------------

// NewMsgRemoveScopeSpecificationRequest creates a new msg instance
func NewMsgRemoveScopeSpecificationRequest() *MsgDeleteScopeSpecificationRequest {
	return &MsgDeleteScopeSpecificationRequest{}
}

func (msg MsgDeleteScopeSpecificationRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// Route returns the module route
func (msg MsgDeleteScopeSpecificationRequest) Route() string {
	return ModuleName
}

// Type returns the type name for this msg
func (msg MsgDeleteScopeSpecificationRequest) Type() string {
	return TypeMsgRemoveScopeSpecificationRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgDeleteScopeSpecificationRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgDeleteScopeSpecificationRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic performs a quick validity check
func (msg MsgDeleteScopeSpecificationRequest) ValidateBasic() error {
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	return nil
}

// ------------------  MsgAddContractSpecificationRequest  ------------------

// NewMsgAddContractSpecificationRequest creates a new msg instance
func NewMsgAddContractSpecificationRequest() *MsgAddContractSpecificationRequest {
	return &MsgAddContractSpecificationRequest{}
}

func (msg MsgAddContractSpecificationRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// Route returns the module route
func (msg MsgAddContractSpecificationRequest) Route() string {
	return ModuleName
}

// Type returns the type name for this msg
func (msg MsgAddContractSpecificationRequest) Type() string {
	return TypeMsgAddContractSpecificationRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgAddContractSpecificationRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgAddContractSpecificationRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic performs a quick validity check
func (msg MsgAddContractSpecificationRequest) ValidateBasic() error {
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	return msg.Specification.ValidateBasic()
}

// ------------------  MsgDeleteContractSpecificationRequest  ------------------

// NewMsgRemoveContractSpecificationRequest creates a new msg instance
func NewMsgRemoveContractSpecificationRequest() *MsgDeleteContractSpecificationRequest {
	return &MsgDeleteContractSpecificationRequest{}
}

func (msg MsgDeleteContractSpecificationRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// Route returns the module route
func (msg MsgDeleteContractSpecificationRequest) Route() string {
	return ModuleName
}

// Type returns the type name for this msg
func (msg MsgDeleteContractSpecificationRequest) Type() string {
	return TypeMsgRemoveScopeSpecificationRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgDeleteContractSpecificationRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgDeleteContractSpecificationRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic performs a quick validity check
func (msg MsgDeleteContractSpecificationRequest) ValidateBasic() error {
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	return nil
}
