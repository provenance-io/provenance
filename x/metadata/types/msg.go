package types

import (
	"fmt"
	"net/url"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/metadata/types/p8e"
	yaml "gopkg.in/yaml.v2"
)

const (
	TypeMsgAddScopeRequest                    = "add_scope_request"
	TypeMsgDeleteScopeRequest                 = "delete_scope_request"
	TypeMsgAddSessionRequest                  = "add_session_request"
	TypeMsgAddRecordRequest                   = "add_record_request"
	TypeMsgDeleteRecordRequest                = "delete_record_request"
	TypeMsgAddScopeSpecificationRequest       = "add_scope_specification_request"
	TypeMsgDeleteScopeSpecificationRequest    = "delete_scope_specification_request"
	TypeMsgAddContractSpecificationRequest    = "add_contract_specification_request"
	TypeMsgDeleteContractSpecificationRequest = "delete_contract_specification_request"
	TypeMsgAddRecordSpecificationRequest      = "add_record_specification_request"
	TypeMsgDeleteRecordSpecificationRequest   = "delete_record_specification_request"
	TypeMsgAddP8EContractSpecRequest          = "add_p8e_contract_spec_request"
	TypeMsgP8eMemorializeContractRequest      = "p8e_memorialize_contract_request"
	TypeMsgBindOSLocatorRequest               = "add_os_locator_request"
	TypeMsgDeleteOSLocatorRequest             = "delete_os_locator_request"
	TypeMsgModifyOSLocatorRequest             = "modify_os_locator_request"
)

// Compile time interface checks.
var (
	_ sdk.Msg = &MsgAddScopeRequest{}
	_ sdk.Msg = &MsgDeleteScopeRequest{}
	_ sdk.Msg = &MsgAddSessionRequest{}
	_ sdk.Msg = &MsgAddRecordRequest{}
	_ sdk.Msg = &MsgDeleteRecordRequest{}
	_ sdk.Msg = &MsgAddScopeSpecificationRequest{}
	_ sdk.Msg = &MsgDeleteScopeSpecificationRequest{}
	_ sdk.Msg = &MsgAddContractSpecificationRequest{}
	_ sdk.Msg = &MsgDeleteContractSpecificationRequest{}
	_ sdk.Msg = &MsgAddRecordSpecificationRequest{}
	_ sdk.Msg = &MsgDeleteRecordSpecificationRequest{}
	_ sdk.Msg = &MsgBindOSLocatorRequest{}
	_ sdk.Msg = &MsgDeleteOSLocatorRequest{}
	_ sdk.Msg = &MsgModifyOSLocatorRequest{}
	_ sdk.Msg = &MsgAddP8EContractSpecRequest{}
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

// ------------------  MsgAddSessionRequest  ------------------

// NewMsgAddSessionRequest creates a new msg instance
func NewMsgAddSessionRequest() *MsgAddSessionRequest {
	return &MsgAddSessionRequest{}
}

func (msg MsgAddSessionRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// Route returns the module route
func (msg MsgAddSessionRequest) Route() string {
	return ModuleName
}

// Type returns the type name for this msg
func (msg MsgAddSessionRequest) Type() string {
	return TypeMsgAddSessionRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgAddSessionRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgAddSessionRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic performs a quick validity check
func (msg MsgAddSessionRequest) ValidateBasic() error {
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	return msg.Session.ValidateBasic()
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
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	return msg.Record.ValidateBasic()
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
func (msg MsgAddScopeSpecificationRequest) Type() string {
	return TypeMsgAddScopeSpecificationRequest
}

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

// ------------------  MsgAddContractSpecRequest  ------------------

// NewMsgAddContractSpecRequest creates a new msg instance
func NewMsgAddP8EContractSpecRequest(contractSpec p8e.ContractSpec, signers []string) *MsgAddP8EContractSpecRequest {
	return &MsgAddP8EContractSpecRequest{
		Contractspec: contractSpec,
		Signers:      signers,
	}
}

func (msg MsgAddP8EContractSpecRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// Route returns the module route
func (msg MsgAddP8EContractSpecRequest) Route() string {
	return ModuleName
}

// Type returns the type name for this msg
func (msg MsgAddP8EContractSpecRequest) Type() string {
	return TypeMsgAddP8EContractSpecRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgAddP8EContractSpecRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgAddP8EContractSpecRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic performs a quick validity check
func (msg MsgAddP8EContractSpecRequest) ValidateBasic() error {
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

// ------------------  MsgAddContractSpecificationRequest  ------------------

// NewMsgAddContractSpecificationRequest creates a new msg instance
func NewMsgAddContractSpecificationRequest(specification ContractSpecification, signers []string) *MsgAddContractSpecificationRequest {
	return &MsgAddContractSpecificationRequest{Specification: specification, Signers: signers}
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

// ------------------  MsgAddRecordSpecificationRequest  ------------------

// NewMsgAddRecordSpecificationRequest creates a new msg instance
func NewMsgAddRecordSpecificationRequest(recordSpecification RecordSpecification, signers []string) *MsgAddRecordSpecificationRequest {
	return &MsgAddRecordSpecificationRequest{Specification: recordSpecification, Signers: signers}
}

func (msg MsgAddRecordSpecificationRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// Route returns the module route
func (msg MsgAddRecordSpecificationRequest) Route() string {
	return ModuleName
}

// Type returns the type name for this msg
func (msg MsgAddRecordSpecificationRequest) Type() string {
	return TypeMsgAddRecordSpecificationRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgAddRecordSpecificationRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgAddRecordSpecificationRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic performs a quick validity check
func (msg MsgAddRecordSpecificationRequest) ValidateBasic() error {
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	return msg.Specification.ValidateBasic()
}

// ------------------  MsgDeleteRecordSpecificationRequest  ------------------

// NewMsgDeleteRecordSpecificationRequest creates a new msg instance
func NewMsgDeleteRecordSpecificationRequest(specificationID MetadataAddress, signers []string) *MsgDeleteRecordSpecificationRequest {
	return &MsgDeleteRecordSpecificationRequest{SpecificationId: specificationID, Signers: signers}
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

// --------------------------- OSLocator------------------------------------------

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

// ---- Delete OS locator ------
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

/**
Validates OSLocatorObj
*/
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

//----------Modify OS Locator -----------------

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
