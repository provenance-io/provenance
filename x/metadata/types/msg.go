package types

import (
	"crypto/sha512"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	yaml "gopkg.in/yaml.v2"
)

const (
	TypeMsgMemorializeContractRequest   = "memorialize_contract_request"
	TypeMsgChangeOwnershipRequest       = "change_ownershipr_equest"
	TypeMsgAddScopeRequest              = "add_scope_request"
	TypeMsgAddRecordGroupRequest        = "add_recordgroup_request"
	TypeMsgAddRecordRequest             = "add_record_request"
	TypeMsgAddScopeSpecificationRequest = "add_scope_specification_request"
	TypeMsgAddGroupSpecificationRequest = "add_group_specification_request"
)

// Compile time interface checks.
var (
// _ sdk.Msg = MsgMemorializeContractRequest{}
// _ sdk.Msg = MsgChangeOwnershipRequest{}
// _ sdk.Msg = MsgAddScopeRequest{}
// _ sdk.Msg = MsgAddRecordGroupRequest{}
// _ sdk.Msg = MsgAddRecordRequest{}
// _ sdk.Msg = MsgAddScopeSpecificationRequest{}
// _ sdk.Msg = MsgAddGroupSpecificationRequest{}
)

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
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
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
	contractBytes, err := msg.Contract.Marshal()
	if err != nil {
		return err
	}
	digest := sha512.Sum512(contractBytes)

	signers := make(map[string]byte)
	for _, s := range msg.Signatures.Signatures {
		addr, err := ValidateContractSignature(*s, digest[:])
		if err != nil {
			return err
		}
		signers[addr.String()] = 1
	}

	requiredSigners := msg.Contract.GetSigners()
	for _, s := range requiredSigners {
		if signers[s.String()] != 1 {
			return fmt.Errorf("missing required signer: %s", s)
		}
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
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
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

	var err error
	var payloadBytes []byte
	if msg.Contract != nil {
		payloadBytes, err = msg.Contract.Marshal()
	} else {
		payloadBytes, err = msg.Recitals.Marshal()
	}
	if err != nil {
		return err
	}
	digest := sha512.Sum512(payloadBytes)

	signers := make(map[string]byte)
	for _, s := range msg.Signatures.Signatures {
		addr, err := ValidateContractSignature(*s, digest[:])
		if err != nil {
			return err
		}
		signers[addr.String()] = 1
	}

	requiredSigners := msg.Contract.GetSigners()
	for _, s := range requiredSigners {
		if signers[s.String()] != 1 {
			return fmt.Errorf("missing required signer: %s", s)
		}
	}

	return msg.Contract.ValidateBasic()
}

// ----------------------------------------------------------------------

// NewMsgAddScopeRequest creates a new msg instance
func NewMsgAddScopeRequest() *MsgAddScopeRequest {
	return &MsgAddScopeRequest{}
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
	delAddr, err := sdk.AccAddressFromBech32(msg.Notary)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{delAddr}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgAddScopeRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic performs a quick validity check
func (msg MsgAddScopeRequest) ValidateBasic() error {
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
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
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
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
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
	delAddr, err := sdk.AccAddressFromBech32(msg.Notary)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{delAddr}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgAddScopeSpecificationRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic performs a quick validity check
func (msg MsgAddScopeSpecificationRequest) ValidateBasic() error {
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
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic performs a quick validity check
func (msg MsgAddGroupSpecificationRequest) ValidateBasic() error {
	return nil
}
