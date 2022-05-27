package types

import (
	"errors"
	"fmt"

	"github.com/gogo/protobuf/proto"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	feegranttypes "github.com/cosmos/cosmos-sdk/x/feegrant"
)

const (
	TypeAddMarkerRequest    = "addmarker"
	TypeAddAccessRequest    = "addaccess"
	TypeDeleteAccessRequest = "deleteaccess"
	TypeFinalizeRequest     = "finalize"
	TypeActivateRequest     = "activate"
	TypeCancelRequest       = "cancel"
	TypeDeleteRequest       = "delete"
	TypeMintRequest         = "mint"
	TypeBurnRequest         = "burn"
	TypeWithdrawRequest     = "withdraw"
	TypeTransferRequest     = "transfer"
	TypeSetMetadataRequest  = "setmetadata"
	TypeGrantAllowance      = "grantallowance"
)

// Compile time interface check.
var (
	_ sdk.Msg = &MsgAddMarkerRequest{}
	_ sdk.Msg = &MsgAddAccessRequest{}
	_ sdk.Msg = &MsgDeleteAccessRequest{}
	_ sdk.Msg = &MsgFinalizeRequest{}
	_ sdk.Msg = &MsgActivateRequest{}
	_ sdk.Msg = &MsgCancelRequest{}
	_ sdk.Msg = &MsgDeleteRequest{}
	_ sdk.Msg = &MsgMintRequest{}
	_ sdk.Msg = &MsgBurnRequest{}
	_ sdk.Msg = &MsgWithdrawRequest{}
	_ sdk.Msg = &MsgTransferRequest{}
	_ sdk.Msg = &MsgGrantAllowanceRequest{}
)

// Type returns the message action.
func (msg MsgAddMarkerRequest) Type() string { return TypeAddMarkerRequest }

// Type returns the message action.
func (msg MsgAddAccessRequest) Type() string { return TypeAddAccessRequest }

// Type returns the message action.
func (msg MsgDeleteAccessRequest) Type() string { return TypeDeleteAccessRequest }

// Type returns the message action.
func (msg MsgFinalizeRequest) Type() string { return TypeFinalizeRequest }

// Type returns the message action.
func (msg MsgActivateRequest) Type() string { return TypeActivateRequest }

// Type returns the message action.
func (msg MsgCancelRequest) Type() string { return TypeCancelRequest }

// Type returns the message action.
func (msg MsgDeleteRequest) Type() string { return TypeDeleteRequest }

// Type returns the message action.
func (msg MsgMintRequest) Type() string { return TypeMintRequest }

// Type returns the message action.
func (msg MsgBurnRequest) Type() string { return TypeBurnRequest }

// Type returns the message action.
func (msg MsgWithdrawRequest) Type() string { return TypeWithdrawRequest }

// Type returns the message action.
func (msg MsgTransferRequest) Type() string { return TypeTransferRequest }

// Type returns the message action.
func (msg MsgSetDenomMetadataRequest) Type() string { return TypeSetMetadataRequest }

// Type returns the message action.
func (msg MsgGrantAllowanceRequest) Type() string { return TypeGrantAllowance }

// NewMsgAddMarkerRequest creates a new marker in a proposed state with a given total supply a denomination
func NewMsgAddMarkerRequest(
	denom string, totalSupply sdk.Int, fromAddress sdk.AccAddress, manager sdk.AccAddress, markerType MarkerType, supplyFixed bool, allowGovernanceControl bool, // nolint:interfacer
) *MsgAddMarkerRequest {
	return &MsgAddMarkerRequest{
		Amount:                 sdk.NewCoin(denom, totalSupply),
		Manager:                manager.String(),
		FromAddress:            fromAddress.String(),
		Status:                 StatusProposed,
		MarkerType:             markerType,
		SupplyFixed:            supplyFixed,
		AllowGovernanceControl: allowGovernanceControl,
	}
}

// Route returns the name of the module.
func (msg MsgAddMarkerRequest) Route() string { return ModuleName }

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgAddMarkerRequest) ValidateBasic() error {
	if msg.Status == StatusUndefined {
		return ErrInvalidMarkerStatus
	}
	// A proposed marker must have a manager assigned to allow updates to be made by the caller.
	if len(msg.Manager) == 0 && msg.Status == StatusProposed {
		return fmt.Errorf("marker manager cannot be empty when creating a proposed marker")
	}
	if msg.Status != StatusFinalized && msg.Status != StatusProposed {
		return fmt.Errorf("marker can only be created with a Proposed or Finalized status")
	}
	testCoin := sdk.Coin{
		Denom:  msg.Amount.Denom,
		Amount: msg.Amount.Amount,
	}
	if !testCoin.IsValid() {
		return fmt.Errorf("invalid marker denom/total supply: %w", sdkerrors.ErrInvalidCoins)
	}

	return nil
}

// GetSignBytes encodes the message for signing.
func (msg MsgAddMarkerRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgAddMarkerRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// NewAddAccessRequest
func NewMsgAddAccessRequest(denom string, admin sdk.AccAddress, access AccessGrant) *MsgAddAccessRequest { //nolint:interfacer
	return &MsgAddAccessRequest{
		Denom:         denom,
		Administrator: admin.String(),
		Access:        []AccessGrant{access},
	}
}

// Route returns the name of the module.
func (msg MsgAddAccessRequest) Route() string { return ModuleName }

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgAddAccessRequest) ValidateBasic() error {
	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return fmt.Errorf(err.Error())
	}
	if err := ValidateGrants(msg.Access...); err != nil {
		return fmt.Errorf(err.Error())
	}
	return nil
}

// GetSignBytes encodes the message for signing.
func (msg MsgAddAccessRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgAddAccessRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Administrator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// NewDeleteAccessRequest
func NewDeleteAccessRequest(denom string, admin sdk.AccAddress, removed sdk.AccAddress) *MsgDeleteAccessRequest { // nolint:interfacer
	return &MsgDeleteAccessRequest{
		Denom:          denom,
		Administrator:  admin.String(),
		RemovedAddress: removed.String(),
	}
}

// Route returns the name of the module.
func (msg MsgDeleteAccessRequest) Route() string { return ModuleName }

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgDeleteAccessRequest) ValidateBasic() error {
	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return fmt.Errorf(err.Error())
	}
	_, err := sdk.AccAddressFromBech32(msg.RemovedAddress)
	return err
}

// GetSignBytes encodes the message for signing.
func (msg MsgDeleteAccessRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgDeleteAccessRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Administrator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// NewMsgFinalizeRequest
func NewMsgFinalizeRequest(denom string, admin sdk.AccAddress) *MsgFinalizeRequest { // nolint:interfacer
	return &MsgFinalizeRequest{
		Denom:         denom,
		Administrator: admin.String(),
	}
}

// Route returns the name of the module.
func (msg MsgFinalizeRequest) Route() string { return ModuleName }

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgFinalizeRequest) ValidateBasic() error {
	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return fmt.Errorf(err.Error())
	}
	return nil
}

// GetSignBytes encodes the message for signing.
func (msg MsgFinalizeRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgFinalizeRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Administrator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// NewMsgActivateRequest
func NewMsgActivateRequest(denom string, admin sdk.AccAddress) *MsgActivateRequest { // nolint:interfacer
	return &MsgActivateRequest{
		Denom:         denom,
		Administrator: admin.String(),
	}
}

// Route returns the name of the module.
func (msg MsgActivateRequest) Route() string { return ModuleName }

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgActivateRequest) ValidateBasic() error {
	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return fmt.Errorf(err.Error())
	}
	return nil
}

// GetSignBytes encodes the message for signing.
func (msg MsgActivateRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgActivateRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Administrator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// NewMsgCancelRequest
func NewMsgCancelRequest(denom string, admin sdk.AccAddress) *MsgCancelRequest { // nolint:interfacer
	return &MsgCancelRequest{
		Denom:         denom,
		Administrator: admin.String(),
	}
}

// Route returns the name of the module.
func (msg MsgCancelRequest) Route() string { return ModuleName }

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgCancelRequest) ValidateBasic() error {
	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return fmt.Errorf(err.Error())
	}
	return nil
}

// GetSignBytes encodes the message for signing.
func (msg MsgCancelRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgCancelRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Administrator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// NewMsgDeleteRequest
func NewMsgDeleteRequest(denom string, admin sdk.AccAddress) *MsgDeleteRequest { // nolint:interfacer
	return &MsgDeleteRequest{
		Denom:         denom,
		Administrator: admin.String(),
	}
}

// Route returns the name of the module.
func (msg MsgDeleteRequest) Route() string { return ModuleName }

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgDeleteRequest) ValidateBasic() error {
	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return fmt.Errorf(err.Error())
	}
	return nil
}

// GetSignBytes encodes the message for signing.
func (msg MsgDeleteRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgDeleteRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Administrator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// NewMsgMintRequest creates a mint supply message
func NewMsgMintRequest(admin sdk.AccAddress, amount sdk.Coin) *MsgMintRequest { // nolint:interfacer
	return &MsgMintRequest{
		Administrator: admin.String(),
		Amount:        amount,
	}
}

// Route returns the name of the module.
func (msg MsgMintRequest) Route() string { return ModuleName }

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgMintRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Administrator); err != nil {
		return err
	}
	return msg.Amount.Validate()
}

// GetSignBytes encodes the message for signing.
func (msg MsgMintRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgMintRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Administrator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// NewMsgBurnRequest creates a burn supply message
func NewMsgBurnRequest(admin sdk.AccAddress, amount sdk.Coin) *MsgBurnRequest { // nolint:interfacer
	return &MsgBurnRequest{
		Administrator: admin.String(),
		Amount:        amount,
	}
}

// Route returns the name of the module.
func (msg MsgBurnRequest) Route() string { return ModuleName }

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgBurnRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Administrator); err != nil {
		return err
	}

	return msg.Amount.Validate()
}

// GetSignBytes encodes the message for signing.
func (msg MsgBurnRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgBurnRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Administrator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// NewMsgWithdrawRequest
func NewMsgWithdrawRequest(
	admin sdk.AccAddress, toAddress sdk.AccAddress, denom string, amount sdk.Coins,
) *MsgWithdrawRequest {
	if toAddress.Empty() {
		toAddress = admin
	}
	return &MsgWithdrawRequest{
		Denom:         denom,
		Administrator: admin.String(),
		ToAddress:     toAddress.String(),
		Amount:        amount,
	}
}

// Route returns the name of the module.
func (msg MsgWithdrawRequest) Route() string { return ModuleName }

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgWithdrawRequest) ValidateBasic() error {
	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return err
	}
	if _, err := sdk.AccAddressFromBech32(msg.Administrator); err != nil {
		return err
	}
	if msg.ToAddress != "" {
		if _, err := sdk.AccAddressFromBech32(msg.ToAddress); err != nil {
			return err
		}
	}

	return msg.Amount.Validate()
}

// GetSignBytes encodes the message for signing.
func (msg MsgWithdrawRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgWithdrawRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Administrator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// NewMsgTransferRequest
func NewMsgTransferRequest(
	admin, fromAddress, toAddress sdk.AccAddress, amount sdk.Coin, // nolint:interfacer
) *MsgTransferRequest {
	return &MsgTransferRequest{
		Administrator: admin.String(),
		ToAddress:     toAddress.String(),
		FromAddress:   fromAddress.String(),
		Amount:        amount,
	}
}

// Route returns the name of the module.
func (msg MsgTransferRequest) Route() string { return ModuleName }

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgTransferRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Administrator); err != nil {
		return err
	}
	if _, err := sdk.AccAddressFromBech32(msg.ToAddress); err != nil {
		return err
	}
	if _, err := sdk.AccAddressFromBech32(msg.FromAddress); err != nil {
		return err
	}
	return msg.Amount.Validate()
}

// GetSignBytes encodes the message for signing.
func (msg MsgTransferRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgTransferRequest) GetSigners() []sdk.AccAddress {
	adminAddr, err := sdk.AccAddressFromBech32(msg.Administrator)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{adminAddr}
}

// NewSetDenomMetadataRequest  creates a new marker in a proposed state with a given total supply a denomination
func NewSetDenomMetadataRequest(
	metadata banktypes.Metadata, admin sdk.AccAddress, // nolint:interfacer
) *MsgSetDenomMetadataRequest {
	return &MsgSetDenomMetadataRequest{
		Metadata:      metadata,
		Administrator: admin.String(),
	}
}

// Route returns the name of the module.
func (msg MsgSetDenomMetadataRequest) Route() string { return ModuleName }

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgSetDenomMetadataRequest) ValidateBasic() error {
	if len(msg.Administrator) == 0 {
		return errors.New("invalid set denom metadata request: administrator cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Administrator); err != nil {
		return fmt.Errorf("invalid set denom metadata request: administrator must be a bech32 address string: %w", err)
	}
	if err := ValidateDenomMetadataBasic(msg.Metadata); err != nil {
		return fmt.Errorf("invalid set denom metadata request: %w", err)
	}
	return nil
}

// GetSignBytes encodes the message for signing.
func (msg MsgSetDenomMetadataRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgSetDenomMetadataRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Administrator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetFeeAllowanceI returns unpacked FeeAllowance
func (msg MsgGrantAllowanceRequest) GetFeeAllowanceI() (feegranttypes.FeeAllowanceI, error) {
	allowance, ok := msg.Allowance.GetCachedValue().(feegranttypes.FeeAllowanceI)
	if !ok {
		return nil, sdkerrors.Wrap(feegranttypes.ErrNoAllowance, "failed to get allowance")
	}

	return allowance, nil
}

// NewMsgAddMarkerRequest creates a new marker in a proposed state with a given total supply a denomination
func NewMsgGrantAllowance(
	denom string, admin sdk.AccAddress, grantee sdk.AccAddress, allowance feegranttypes.FeeAllowanceI, // nolint:interfacer
) (*MsgGrantAllowanceRequest, error) {
	msg, ok := allowance.(proto.Message)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrPackAny, "cannot proto marshal %T", msg)
	}
	any, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}

	return &MsgGrantAllowanceRequest{
		Denom:         denom,
		Administrator: admin.String(),
		Grantee:       grantee.String(),
		Allowance:     any,
	}, nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgGrantAllowanceRequest) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var allowance feegranttypes.FeeAllowanceI
	return unpacker.UnpackAny(msg.Allowance, &allowance)
}

// Route returns the name of the module.
func (msg MsgGrantAllowanceRequest) Route() string { return ModuleName }

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgGrantAllowanceRequest) ValidateBasic() error {
	if msg.Denom == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "missing marker denom")
	}
	if msg.Administrator == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing administrator address")
	}
	if msg.Grantee == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing grantee address")
	}

	allowance, err := msg.GetFeeAllowanceI()
	if err != nil {
		return err
	}

	return allowance.ValidateBasic()
}

// GetSignBytes encodes the message for signing.
func (msg MsgGrantAllowanceRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgGrantAllowanceRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Administrator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}
