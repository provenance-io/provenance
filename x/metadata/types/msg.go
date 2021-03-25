package types

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
	"github.com/provenance-io/provenance/x/metadata/types/p8e"
	yaml "gopkg.in/yaml.v2"
)

const (
	TypeMsgWriteScopeRequest                  = "write_scope_request"
	TypeMsgDeleteScopeRequest                 = "delete_scope_request"
	TypeMsgWriteSessionRequest                = "write_session_request"
	TypeMsgWriteRecordRequest                 = "write_record_request"
	TypeMsgDeleteRecordRequest                = "delete_record_request"
	TypeMsgWriteScopeSpecificationRequest     = "write_scope_specification_request"
	TypeMsgDeleteScopeSpecificationRequest    = "delete_scope_specification_request"
	TypeMsgWriteContractSpecificationRequest  = "write_contract_specification_request"
	TypeMsgDeleteContractSpecificationRequest = "delete_contract_specification_request"
	TypeMsgWriteRecordSpecificationRequest    = "write_record_specification_request"
	TypeMsgDeleteRecordSpecificationRequest   = "delete_record_specification_request"
	TypeMsgWriteP8EContractSpecRequest        = "write_p8e_contract_spec_request"
	TypeMsgP8eMemorializeContractRequest      = "p8e_memorialize_contract_request"
	TypeMsgBindOSLocatorRequest               = "write_os_locator_request"
	TypeMsgDeleteOSLocatorRequest             = "delete_os_locator_request"
	TypeMsgModifyOSLocatorRequest             = "modify_os_locator_request"
)

// Compile time interface checks.
var (
	_ sdk.Msg = &MsgWriteScopeRequest{}
	_ sdk.Msg = &MsgDeleteScopeRequest{}
	_ sdk.Msg = &MsgWriteSessionRequest{}
	_ sdk.Msg = &MsgWriteRecordRequest{}
	_ sdk.Msg = &MsgDeleteRecordRequest{}
	_ sdk.Msg = &MsgWriteScopeSpecificationRequest{}
	_ sdk.Msg = &MsgDeleteScopeSpecificationRequest{}
	_ sdk.Msg = &MsgWriteContractSpecificationRequest{}
	_ sdk.Msg = &MsgDeleteContractSpecificationRequest{}
	_ sdk.Msg = &MsgWriteRecordSpecificationRequest{}
	_ sdk.Msg = &MsgDeleteRecordSpecificationRequest{}
	_ sdk.Msg = &MsgBindOSLocatorRequest{}
	_ sdk.Msg = &MsgDeleteOSLocatorRequest{}
	_ sdk.Msg = &MsgModifyOSLocatorRequest{}
	_ sdk.Msg = &MsgWriteP8EContractSpecRequest{}
	_ sdk.Msg = &MsgP8EMemorializeContractRequest{}
)

// private method to convert an array of strings into an array of Acc Addresses.
func stringsToAccAddresses(strings []string) []sdk.AccAddress {
	retval := make([]sdk.AccAddress, len(strings))

	for i, str := range strings {
		retval[i] = stringToAccAddress(str)
	}

	return retval
}

func stringToAccAddress(s string) sdk.AccAddress {
	accAddress, err := sdk.AccAddressFromBech32(s)
	if err != nil {
		panic(err)
	}
	return accAddress
}

// ------------------  MsgWriteScopeRequest  ------------------

// NewMsgWriteScopeRequest creates a new msg instance
func NewMsgWriteScopeRequest(scope Scope, signers []string) *MsgWriteScopeRequest {
	return &MsgWriteScopeRequest{
		Scope:   scope,
		Signers: signers,
	}
}

func (msg MsgWriteScopeRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// Route returns the module route
func (msg MsgWriteScopeRequest) Route() string {
	return ModuleName
}

// Type returns the type name for this msg
func (msg MsgWriteScopeRequest) Type() string {
	return TypeMsgWriteScopeRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgWriteScopeRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgWriteScopeRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic performs a quick validity check
func (msg MsgWriteScopeRequest) ValidateBasic() error {
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	return nil
}

// ConvertOptionalFields will look at the ScopeUuid and SpecUuid fields in the message.
// For each, if present, it will be converted to a MetadataAddress and set in the Scope appropriately.
// Once used, those uuid fields will be set to empty strings so that calling this again has no effect.
func (msg *MsgWriteScopeRequest) ConvertOptionalFields() error {
	if len(msg.ScopeUuid) > 0 {
		uid, err := uuid.Parse(msg.ScopeUuid)
		if err != nil {
			return fmt.Errorf("invalid scope uuid: %w", err)
		}
		msg.Scope.ScopeId = ScopeMetadataAddress(uid)
		msg.ScopeUuid = ""
	}
	if len(msg.SpecUuid) > 0 {
		uid, err := uuid.Parse(msg.SpecUuid)
		if err != nil {
			return fmt.Errorf("invalid spec uuid: %w", err)
		}
		msg.Scope.SpecificationId = ScopeSpecMetadataAddress(uid)
		msg.SpecUuid = ""
	}
	return nil
}

// ------------------  NewMsgDeleteScopeRequest  ------------------

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

// ------------------  MsgWriteSessionRequest  ------------------

// NewMsgWriteSessionRequest creates a new msg instance
func NewMsgWriteSessionRequest() *MsgWriteSessionRequest {
	return &MsgWriteSessionRequest{}
}

func (msg MsgWriteSessionRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// Route returns the module route
func (msg MsgWriteSessionRequest) Route() string {
	return ModuleName
}

// Type returns the type name for this msg
func (msg MsgWriteSessionRequest) Type() string {
	return TypeMsgWriteSessionRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgWriteSessionRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgWriteSessionRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic performs a quick validity check
func (msg MsgWriteSessionRequest) ValidateBasic() error {
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	return nil
}

// ConvertOptionalFields will look at the SessionIdComponents and SpecUuid fields in the message.
// For each, if present, it will be converted to a MetadataAddress and set in the Session appropriately.
// Once used, those fields will be emptied so that calling this again has no effect.
func (msg *MsgWriteSessionRequest) ConvertOptionalFields() error {
	if msg.SessionIdComponents != nil {
		sessionID, err := msg.SessionIdComponents.GetSessionID()
		if err != nil {
			return fmt.Errorf("invalid session id components: %w", err)
		}
		if sessionID != nil {
			msg.Session.SessionId = *sessionID
		}
		msg.SessionIdComponents = nil
	}
	if len(msg.SpecUuid) > 0 {
		uid, err := uuid.Parse(msg.SpecUuid)
		if err != nil {
			return fmt.Errorf("invalid spec uuid: %w", err)
		}
		msg.Session.SpecificationId = ContractSpecMetadataAddress(uid)
		msg.SpecUuid = ""
	}
	return nil
}

// ------------------  MsgWriteRecordRequest  ------------------

// NewMsgWriteRecordRequest creates a new msg instance
func NewMsgWriteRecordRequest() *MsgWriteRecordRequest {
	return &MsgWriteRecordRequest{}
}

func (msg MsgWriteRecordRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// Route returns the module route
func (msg MsgWriteRecordRequest) Route() string {
	return ModuleName
}

// Type returns the type name for this msg
func (msg MsgWriteRecordRequest) Type() string {
	return TypeMsgWriteRecordRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgWriteRecordRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgWriteRecordRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic performs a quick validity check
func (msg MsgWriteRecordRequest) ValidateBasic() error {
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	return nil
}

// ConvertOptionalFields will look at the SessionIdComponents and ContractSpecUuid fields in the message.
// For each, if present, it will be converted to a MetadataAddress and set in the Record appropriately.
// Once used, those fields will be emptied so that calling this again has no effect.
func (msg *MsgWriteRecordRequest) ConvertOptionalFields() error {
	if msg.SessionIdComponents != nil {
		sessionID, err := msg.SessionIdComponents.GetSessionID()
		if err != nil {
			return fmt.Errorf("invalid session id components: %w", err)
		}
		if sessionID != nil {
			msg.Record.SessionId = *sessionID
		}
		msg.SessionIdComponents = nil
	}
	if len(msg.ContractSpecUuid) > 0 {
		uid, err := uuid.Parse(msg.ContractSpecUuid)
		if err != nil {
			return fmt.Errorf("invalid contract spec uuid: %w", err)
		}
		if len(strings.TrimSpace(msg.Record.Name)) == 0 {
			return errors.New("empty record name")
		}
		msg.Record.SpecificationId = RecordSpecMetadataAddress(uid, msg.Record.Name)
		msg.ContractSpecUuid = ""
	}
	return nil
}

// ------------------  MsgDeleteRecordRequest  ------------------

// NewMsgDeleteScopeSpecificationRequest creates a new msg instance
func NewMsgDeleteRecordRequest() *MsgDeleteRecordRequest {
	return &MsgDeleteRecordRequest{}
}

func (msg MsgDeleteRecordRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// Route returns the module route
func (msg MsgDeleteRecordRequest) Route() string {
	return ModuleName
}

// Type returns the type name for this msg
func (msg MsgDeleteRecordRequest) Type() string {
	return TypeMsgDeleteRecordRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgDeleteRecordRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgDeleteRecordRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic performs a quick validity check
func (msg MsgDeleteRecordRequest) ValidateBasic() error {
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	return nil
}

// ------------------  MsgWriteScopeSpecificationRequest  ------------------

// NewMsgWriteScopeSpecificationRequest creates a new msg instance
func NewMsgWriteScopeSpecificationRequest() *MsgWriteScopeSpecificationRequest {
	return &MsgWriteScopeSpecificationRequest{}
}

func (msg MsgWriteScopeSpecificationRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// Route returns the module route
func (msg MsgWriteScopeSpecificationRequest) Route() string {
	return ModuleName
}

// Type returns the type name for this msg
func (msg MsgWriteScopeSpecificationRequest) Type() string {
	return TypeMsgWriteScopeSpecificationRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgWriteScopeSpecificationRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgWriteScopeSpecificationRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic performs a quick validity check
func (msg MsgWriteScopeSpecificationRequest) ValidateBasic() error {
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	return nil
}

// ConvertOptionalFields will look at the SpecUuid field in the message.
// If present, it will be converted to a MetadataAddress and set in the Specification appropriately.
// Once used, it will be emptied so that calling this again has no effect.
func (msg *MsgWriteScopeSpecificationRequest) ConvertOptionalFields() error {
	if len(msg.SpecUuid) > 0 {
		uid, err := uuid.Parse(msg.SpecUuid)
		if err != nil {
			return fmt.Errorf("invalid spec uuid: %w", err)
		}
		msg.Specification.SpecificationId = ScopeSpecMetadataAddress(uid)
		msg.SpecUuid = ""
	}
	return nil
}

// ------------------  MsgWriteP8EContractSpecRequest  ------------------

// NewMsgWriteContractSpecRequest creates a new msg instance
func NewMsgWriteP8EContractSpecRequest(contractSpec p8e.ContractSpec, signers []string) *MsgWriteP8EContractSpecRequest {
	return &MsgWriteP8EContractSpecRequest{
		Contractspec: contractSpec,
		Signers:      signers,
	}
}

func (msg MsgWriteP8EContractSpecRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// Route returns the module route
func (msg MsgWriteP8EContractSpecRequest) Route() string {
	return ModuleName
}

// Type returns the type name for this msg
func (msg MsgWriteP8EContractSpecRequest) Type() string {
	return TypeMsgWriteP8EContractSpecRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgWriteP8EContractSpecRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgWriteP8EContractSpecRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic performs a quick validity check
func (msg MsgWriteP8EContractSpecRequest) ValidateBasic() error {
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	_, _, err := ConvertP8eContractSpec(&msg.Contractspec, msg.Signers)
	if err != nil {
		return fmt.Errorf("failed to convert p8e ContractSpec %s", err)
	}
	return nil
}

// ------------------  MsgDeleteScopeSpecificationRequest  ------------------

// NewMsgDeleteScopeSpecificationRequest creates a new msg instance
func NewMsgDeleteScopeSpecificationRequest() *MsgDeleteScopeSpecificationRequest {
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
	return TypeMsgDeleteScopeSpecificationRequest
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

// ------------------  MsgWriteContractSpecificationRequest  ------------------

// NewMsgWriteContractSpecificationRequest creates a new msg instance
func NewMsgWriteContractSpecificationRequest(specification ContractSpecification, signers []string) *MsgWriteContractSpecificationRequest {
	return &MsgWriteContractSpecificationRequest{Specification: specification, Signers: signers}
}

func (msg MsgWriteContractSpecificationRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// Route returns the module route
func (msg MsgWriteContractSpecificationRequest) Route() string {
	return ModuleName
}

// Type returns the type name for this msg
func (msg MsgWriteContractSpecificationRequest) Type() string {
	return TypeMsgWriteContractSpecificationRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgWriteContractSpecificationRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgWriteContractSpecificationRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic performs a quick validity check
func (msg MsgWriteContractSpecificationRequest) ValidateBasic() error {
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	return nil
}

// ConvertOptionalFields will look at the SpecUuid field in the message.
// If present, it will be converted to a MetadataAddress and set in the Specification appropriately.
// Once used, it will be emptied so that calling this again has no effect.
func (msg *MsgWriteContractSpecificationRequest) ConvertOptionalFields() error {
	if len(msg.SpecUuid) > 0 {
		uid, err := uuid.Parse(msg.SpecUuid)
		if err != nil {
			return fmt.Errorf("invalid spec uuid: %w", err)
		}
		msg.Specification.SpecificationId = ContractSpecMetadataAddress(uid)
		msg.SpecUuid = ""
	}
	return nil
}

// ------------------  MsgDeleteContractSpecificationRequest  ------------------

// NewMsgDeleteContractSpecificationRequest creates a new msg instance
func NewMsgDeleteContractSpecificationRequest(specificationID MetadataAddress, signers []string) *MsgDeleteContractSpecificationRequest {
	return &MsgDeleteContractSpecificationRequest{SpecificationId: specificationID, Signers: signers}
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
	return TypeMsgDeleteContractSpecificationRequest
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

// ------------------  MsgWriteRecordSpecificationRequest  ------------------

// NewMsgWriteRecordSpecificationRequest creates a new msg instance
func NewMsgWriteRecordSpecificationRequest() *MsgWriteRecordSpecificationRequest {
	return &MsgWriteRecordSpecificationRequest{}
}

func (msg MsgWriteRecordSpecificationRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// Route returns the module route
func (msg MsgWriteRecordSpecificationRequest) Route() string {
	return ModuleName
}

// Type returns the type name for this msg
func (msg MsgWriteRecordSpecificationRequest) Type() string {
	return TypeMsgWriteRecordSpecificationRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgWriteRecordSpecificationRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgWriteRecordSpecificationRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic performs a quick validity check
func (msg MsgWriteRecordSpecificationRequest) ValidateBasic() error {
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	return nil
}

// ConvertOptionalFields will look at the ContractSpecUuid field in the message.
// If present, it will be converted to a MetadataAddress and set in the Specification appropriately.
// Once used, it will be emptied so that calling this again has no effect.
func (msg *MsgWriteRecordSpecificationRequest) ConvertOptionalFields() error {
	if len(msg.ContractSpecUuid) > 0 {
		uid, err := uuid.Parse(msg.ContractSpecUuid)
		if err != nil {
			return fmt.Errorf("invalid spec uuid: %w", err)
		}
		if len(strings.TrimSpace(msg.Specification.Name)) == 0 {
			return errors.New("empty specification name")
		}
		msg.Specification.SpecificationId = RecordSpecMetadataAddress(uid, msg.Specification.Name)
		msg.ContractSpecUuid = ""
	}
	return nil
}

// ------------------  MsgDeleteRecordSpecificationRequest  ------------------

// NewMsgDeleteRecordSpecificationRequest creates a new msg instance
func NewMsgDeleteRecordSpecificationRequest() *MsgDeleteRecordSpecificationRequest {
	return &MsgDeleteRecordSpecificationRequest{}
}

func (msg MsgDeleteRecordSpecificationRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// Route returns the module route
func (msg MsgDeleteRecordSpecificationRequest) Route() string {
	return ModuleName
}

// Type returns the type name for this msg
func (msg MsgDeleteRecordSpecificationRequest) Type() string {
	return TypeMsgDeleteRecordSpecificationRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgDeleteRecordSpecificationRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgDeleteRecordSpecificationRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic performs a quick validity check
func (msg MsgDeleteRecordSpecificationRequest) ValidateBasic() error {
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	return nil
}

// ------------------  MsgP8EMemorializeContractRequest  ------------------

// NewMsgP8EMemorializeContractRequest creates a new msg instance
func NewMsgP8EMemorializeContractRequest() *MsgP8EMemorializeContractRequest {
	return &MsgP8EMemorializeContractRequest{}
}

func (msg MsgP8EMemorializeContractRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// Route returns the module route
func (msg MsgP8EMemorializeContractRequest) Route() string {
	return ModuleName
}

// Type returns the type name for this msg
func (msg MsgP8EMemorializeContractRequest) Type() string {
	return TypeMsgP8eMemorializeContractRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgP8EMemorializeContractRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{stringToAccAddress(msg.Invoker)}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgP8EMemorializeContractRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic performs a quick validity check
func (msg MsgP8EMemorializeContractRequest) ValidateBasic() error {
	_, _, err := ConvertP8eMemorializeContractRequest(&msg)
	return err
}

// ------------------  MsgBindOSLocatorRequest  ------------------

// NewMsgBindOSLocatorRequest creates a new msg instance
func NewMsgBindOSLocatorRequest(obj ObjectStoreLocator) *MsgBindOSLocatorRequest {
	return &MsgBindOSLocatorRequest{
		Locator: obj,
	}
}

func (msg MsgBindOSLocatorRequest) Route() string {
	return ModuleName
}

func (msg MsgBindOSLocatorRequest) Type() string {
	return TypeMsgBindOSLocatorRequest
}

func (msg MsgBindOSLocatorRequest) ValidateBasic() error {
	err := ValidateOSLocatorObj(msg.Locator.Owner, msg.Locator.LocatorUri)
	if err != nil {
		return err
	}
	return nil
}

func (msg MsgBindOSLocatorRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgBindOSLocatorRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{stringToAccAddress(msg.Locator.Owner)}
}

// ------------------  MsgDeleteOSLocatorRequest  ------------------

func NewMsgDeleteOSLocatorRequest(obj ObjectStoreLocator) *MsgDeleteOSLocatorRequest {
	return &MsgDeleteOSLocatorRequest{
		Locator: obj,
	}
}
func (msg MsgDeleteOSLocatorRequest) Route() string {
	return ModuleName
}

func (msg MsgDeleteOSLocatorRequest) Type() string {
	return TypeMsgDeleteOSLocatorRequest
}

func (msg MsgDeleteOSLocatorRequest) ValidateBasic() error {
	err := ValidateOSLocatorObj(msg.Locator.Owner, msg.Locator.LocatorUri)
	if err != nil {
		return err
	}

	return nil
}

func (msg MsgDeleteOSLocatorRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// Signers returns the addrs of signers that must sign.
// CONTRACT: All signatures must be present to be valid.
// CONTRACT: Returns addrs in some deterministic order.
// here we assume msg for delete request has the right address
// should be verified later in the keeper?
func (msg MsgDeleteOSLocatorRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{stringToAccAddress(msg.Locator.Owner)}
}

// ValidateOSLocatorObj Validates OSLocatorObj data
func ValidateOSLocatorObj(ownerAddr string, uri string) error {
	if strings.TrimSpace(ownerAddr) == "" {
		return fmt.Errorf("owner address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(ownerAddr); err != nil {
		return fmt.Errorf("failed to add locator for a given owner address,"+
			" invalid address: %s", ownerAddr)
	}

	if strings.TrimSpace(uri) == "" {
		return fmt.Errorf("uri cannot be empty")
	}

	if _, err := url.Parse(uri); err != nil {
		return fmt.Errorf("failed to add locator for a given"+
			" owner address, invalid uri: %s", uri)
	}
	return nil
}

// ------------------  MsgModifyOSLocatorRequest  ------------------

func NewMsgModifyOSLocatorRequest(obj ObjectStoreLocator) *MsgModifyOSLocatorRequest {
	return &MsgModifyOSLocatorRequest{
		Locator: obj,
	}
}

func (msg MsgModifyOSLocatorRequest) Route() string {
	return ModuleName
}

func (msg MsgModifyOSLocatorRequest) Type() string {
	return TypeMsgModifyOSLocatorRequest
}

func (msg MsgModifyOSLocatorRequest) ValidateBasic() error {
	err := ValidateOSLocatorObj(msg.Locator.Owner, msg.Locator.LocatorUri)
	if err != nil {
		return err
	}

	return nil
}

func (msg MsgModifyOSLocatorRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgModifyOSLocatorRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{stringToAccAddress(msg.Locator.Owner)}
}

// ------------------  SessionIdComponents  ------------------

func (msg *SessionIdComponents) GetSessionID() (*MetadataAddress, error) {
	var scopeUUID, sessionUUID *uuid.UUID
	if len(msg.SessionUuid) > 0 {
		uid, err := uuid.Parse(msg.SessionUuid)
		if err != nil {
			return nil, fmt.Errorf("invalid session uuid: %w", err)
		}
		scopeUUID = &uid
	}
	if msgScopeUUID := msg.GetScopeUuid(); len(msgScopeUUID) > 0 {
		uid, err := uuid.Parse(msgScopeUUID)
		if err != nil {
			return nil, fmt.Errorf("invalid scope uuid: %w", err)
		}
		sessionUUID = &uid
	} else if msgScopeAddr := msg.GetScopeAddr(); len(msgScopeAddr) > 0 {
		addr, addrErr := MetadataAddressFromBech32(msgScopeAddr)
		if addrErr != nil {
			return nil, fmt.Errorf("invalid scope addr: %w", addrErr)
		}
		uid, err := addr.ScopeUUID()
		if err != nil {
			return nil, fmt.Errorf("invalid scope addr: %w", err)
		}
		sessionUUID = &uid
	}

	if scopeUUID == nil && sessionUUID == nil {
		return nil, nil
	}
	if scopeUUID == nil {
		return nil, errors.New("session uuid provided but missing scope uuid or addr")
	}
	if sessionUUID == nil {
		return nil, errors.New("scope uuid or addr provided but missing session uuid")
	}
	ma := SessionMetadataAddress(*scopeUUID, *sessionUUID)
	return &ma, nil
}
