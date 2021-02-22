package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	yaml "gopkg.in/yaml.v2"
)

const (
	TypeMsgMemorializeContractRequest      = "memorialize_contract_request"
	TypeMsgChangeOwnershipRequest          = "change_ownershipr_equest"
	TypeMsgAddScopeRequest                 = "add_scope_request"
	TypeMsgRemoveScopeRequest              = "remove_scope_request"
	TypeMsgAddRecordGroupRequest           = "add_recordgroup_request"
	TypeMsgAddRecordRequest                = "add_record_request"
	TypeMsgAddScopeSpecificationRequest    = "add_scope_specification_request"
	TypeMsgRemoveScopeSpecificationRequest = "remove_scope_specification_request"
	TypeMsgAddGroupSpecificationRequest    = "add_group_specification_request"
)

// Compile time interface checks.
var (
	_ sdk.Msg = &MsgMemorializeContractRequest{}
	_ sdk.Msg = &MsgChangeOwnershipRequest{}
	_ sdk.Msg = &MsgAddScopeRequest{}
	_ sdk.Msg = &MsgAddRecordGroupRequest{}
	_ sdk.Msg = &MsgAddRecordRequest{}
	_ sdk.Msg = &MsgAddScopeSpecificationRequest{}
	_ sdk.Msg = &MsgAddGroupSpecificationRequest{}
	_ sdk.Msg = &MsgRemoveScopeRequest{}
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

// ----------------------------------------------------------------------

// NewMsgMemorializeContractRequest creates a new msg instance
func NewMsgMemorializeContractRequest() *MsgMemorializeContractRequest {
	return &MsgMemorializeContractRequest{}
}

func (msg MsgMemorializeContractRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// Route returns the module route
func (msg MsgMemorializeContractRequest) Route() string { return ModuleName }

// Type returns the type name for this msg
func (msg MsgMemorializeContractRequest) Type() string { return TypeMsgMemorializeContractRequest }

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

// ----------------------------------------------------------------------

// NewMsgChangeOwnershipRequest creates a new msg instance
func NewMsgChangeOwnershipRequest() *MsgChangeOwnershipRequest {
	return &MsgChangeOwnershipRequest{}
}

func (msg MsgChangeOwnershipRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// Route returns the module route
func (msg MsgChangeOwnershipRequest) Route() string { return ModuleName }

// Type returns the type name for this msg
func (msg MsgChangeOwnershipRequest) Type() string { return TypeMsgChangeOwnershipRequest }

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

// ----------------------------------------------------------------------

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
func (msg MsgAddScopeRequest) Route() string { return ModuleName }

// Type returns the type name for this msg
func (msg MsgAddScopeRequest) Type() string { return TypeMsgAddScopeRequest }

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgAddScopeRequest) GetSigners() []sdk.AccAddress {
	signers := make([]sdk.AccAddress, len(msg.Signers))

	for i, signerAddress := range msg.Signers {
		signAddr, err := sdk.AccAddressFromBech32(signerAddress)
		if err != nil {
			panic(err)
		}
		signers[i] = signAddr
	}

	return signers
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

// ----------------------------------------------------------------------

// NewMsgRemoveScopeRequest creates a new msg instance
func NewMsgRemoveScopeRequest(scopeID MetadataAddress, signers []string) *MsgRemoveScopeRequest {
	return &MsgRemoveScopeRequest{
		ScopeId: scopeID,
		Signers: signers,
	}
}

func (msg MsgRemoveScopeRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// Route returns the module route
func (msg MsgRemoveScopeRequest) Route() string { return ModuleName }

// Type returns the type name for this msg
func (msg MsgRemoveScopeRequest) Type() string { return TypeMsgRemoveScopeRequest }

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgRemoveScopeRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgRemoveScopeRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic performs a quick validity check
func (msg MsgRemoveScopeRequest) ValidateBasic() error {
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	if !msg.ScopeId.IsScopeAddress() {
		return fmt.Errorf("invalid scope address")
	}
	return nil
}

// ----------------------------------------------------------------------

// NewMsgAddRecordGroupRequest creates a new msg instance
func NewMsgAddRecordGroupRequest() *MsgAddRecordGroupRequest {
	return &MsgAddRecordGroupRequest{}
}

func (msg MsgAddRecordGroupRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// Route returns the module route
func (msg MsgAddRecordGroupRequest) Route() string { return ModuleName }

// Type returns the type name for this msg
func (msg MsgAddRecordGroupRequest) Type() string { return TypeMsgAddRecordGroupRequest }

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

// ----------------------------------------------------------------------

// NewMsgAddRecordRequest creates a new msg instance
func NewMsgAddRecordRequest() *MsgAddRecordRequest {
	return &MsgAddRecordRequest{}
}

func (msg MsgAddRecordRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// Route returns the module route
func (msg MsgAddRecordRequest) Route() string { return ModuleName }

// Type returns the type name for this msg
func (msg MsgAddRecordRequest) Type() string { return TypeMsgAddRecordRequest }

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgAddRecordRequest) GetSigners() []sdk.AccAddress {
	delAddr, err := sdk.AccAddressFromBech32(msg.Notary)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{delAddr}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgAddRecordRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic performs a quick validity check
func (msg MsgAddRecordRequest) ValidateBasic() error {
	return nil
}

// ----------------------------------------------------------------------

// NewMsgAddScopeSpecificationRequest creates a new msg instance
func NewMsgAddScopeSpecificationRequest() *MsgAddScopeSpecificationRequest {
	return &MsgAddScopeSpecificationRequest{}
}

func (msg MsgAddScopeSpecificationRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// Route returns the module route
func (msg MsgAddScopeSpecificationRequest) Route() string { return ModuleName }

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

// ----------------------------------------------------------------------

// NewMsgRemoveScopeSpecificationRequest creates a new msg instance
func NewMsgRemoveScopeSpecificationRequest() *MsgRemoveScopeSpecificationRequest {
	return &MsgRemoveScopeSpecificationRequest{}
}

func (msg MsgRemoveScopeSpecificationRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// Route returns the module route
func (msg MsgRemoveScopeSpecificationRequest) Route() string { return ModuleName }

// Type returns the type name for this msg
func (msg MsgRemoveScopeSpecificationRequest) Type() string {
	return TypeMsgRemoveScopeSpecificationRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgRemoveScopeSpecificationRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgRemoveScopeSpecificationRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic performs a quick validity check
func (msg MsgRemoveScopeSpecificationRequest) ValidateBasic() error {
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	return nil
}

// ----------------------------------------------------------------------

// NewMsgAddGroupSpecificationRequest creates a new msg instance
func NewMsgAddGroupSpecificationRequest() *MsgAddGroupSpecificationRequest {
	return &MsgAddGroupSpecificationRequest{}
}

func (msg MsgAddGroupSpecificationRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// Route returns the module route
func (msg MsgAddGroupSpecificationRequest) Route() string { return ModuleName }

// Type returns the type name for this msg
func (msg MsgAddGroupSpecificationRequest) Type() string { return TypeMsgAddGroupSpecificationRequest }

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgAddGroupSpecificationRequest) GetSigners() []sdk.AccAddress {
	delAddr, err := sdk.AccAddressFromBech32(msg.Notary)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{delAddr}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgAddGroupSpecificationRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic performs a quick validity check
func (msg MsgAddGroupSpecificationRequest) ValidateBasic() error {
	return nil
}
