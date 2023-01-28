package types

import (
	"errors"
	"fmt"

	"github.com/gogo/protobuf/proto"

	sdkmath "cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	feegranttypes "github.com/cosmos/cosmos-sdk/x/feegrant"

	ibctransfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
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
	_ sdk.Msg = &MsgIbcTransferRequest{}
	_ sdk.Msg = &MsgGrantAllowanceRequest{}
	_ sdk.Msg = &MsgAddFinalizeActivateMarkerRequest{}
)

// NewMsgAddMarkerRequest creates a new marker in a proposed state with a given total supply a denomination
func NewMsgAddMarkerRequest(
	denom string, totalSupply sdkmath.Int, fromAddress sdk.AccAddress, manager sdk.AccAddress, markerType MarkerType, supplyFixed bool, allowGovernanceControl bool, //nolint:interfacer
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

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgAddAccessRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.Administrator)}
}

// NewDeleteAccessRequest
func NewDeleteAccessRequest(denom string, admin sdk.AccAddress, removed sdk.AccAddress) *MsgDeleteAccessRequest { //nolint:interfacer
	return &MsgDeleteAccessRequest{
		Denom:          denom,
		Administrator:  admin.String(),
		RemovedAddress: removed.String(),
	}
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgDeleteAccessRequest) ValidateBasic() error {
	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return fmt.Errorf(err.Error())
	}
	_, err := sdk.AccAddressFromBech32(msg.RemovedAddress)
	return err
}

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgDeleteAccessRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.Administrator)}
}

// NewMsgFinalizeRequest
func NewMsgFinalizeRequest(denom string, admin sdk.AccAddress) *MsgFinalizeRequest { //nolint:interfacer
	return &MsgFinalizeRequest{
		Denom:         denom,
		Administrator: admin.String(),
	}
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgFinalizeRequest) ValidateBasic() error {
	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return fmt.Errorf(err.Error())
	}
	return nil
}

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgFinalizeRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.Administrator)}
}

// NewMsgActivateRequest
func NewMsgActivateRequest(denom string, admin sdk.AccAddress) *MsgActivateRequest { //nolint:interfacer
	return &MsgActivateRequest{
		Denom:         denom,
		Administrator: admin.String(),
	}
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgActivateRequest) ValidateBasic() error {
	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return fmt.Errorf(err.Error())
	}
	return nil
}

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgActivateRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.Administrator)}
}

// NewMsgCancelRequest
func NewMsgCancelRequest(denom string, admin sdk.AccAddress) *MsgCancelRequest { //nolint:interfacer
	return &MsgCancelRequest{
		Denom:         denom,
		Administrator: admin.String(),
	}
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgCancelRequest) ValidateBasic() error {
	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return fmt.Errorf(err.Error())
	}
	return nil
}

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgCancelRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.Administrator)}
}

// NewMsgDeleteRequest
func NewMsgDeleteRequest(denom string, admin sdk.AccAddress) *MsgDeleteRequest { //nolint:interfacer
	return &MsgDeleteRequest{
		Denom:         denom,
		Administrator: admin.String(),
	}
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgDeleteRequest) ValidateBasic() error {
	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return fmt.Errorf(err.Error())
	}
	return nil
}

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgDeleteRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.Administrator)}
}

// NewMsgMintRequest creates a mint supply message
func NewMsgMintRequest(admin sdk.AccAddress, amount sdk.Coin) *MsgMintRequest { //nolint:interfacer
	return &MsgMintRequest{
		Administrator: admin.String(),
		Amount:        amount,
	}
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgMintRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Administrator); err != nil {
		return err
	}
	return msg.Amount.Validate()
}

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgMintRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.Administrator)}
}

// NewMsgBurnRequest creates a burn supply message
func NewMsgBurnRequest(admin sdk.AccAddress, amount sdk.Coin) *MsgBurnRequest { //nolint:interfacer
	return &MsgBurnRequest{
		Administrator: admin.String(),
		Amount:        amount,
	}
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgBurnRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Administrator); err != nil {
		return err
	}

	return msg.Amount.Validate()
}

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgBurnRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.Administrator)}
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

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgWithdrawRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.Administrator)}
}

// NewMsgTransferRequest
func NewMsgTransferRequest(
	admin, fromAddress, toAddress sdk.AccAddress, amount sdk.Coin, //nolint:interfacer
) *MsgTransferRequest {
	return &MsgTransferRequest{
		Administrator: admin.String(),
		ToAddress:     toAddress.String(),
		FromAddress:   fromAddress.String(),
		Amount:        amount,
	}
}

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

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgTransferRequest) GetSigners() []sdk.AccAddress {
	adminAddr, err := sdk.AccAddressFromBech32(msg.Administrator)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{adminAddr}
}

// NewIbcMsgTransferRequest
func NewIbcMsgTransferRequest(
	administrator string,
	sourcePort,
	sourceChannel string,
	token sdk.Coin,
	sender,
	receiver string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64, //nolint:interfacer
	memo string,
) *MsgIbcTransferRequest {
	return &MsgIbcTransferRequest{
		Administrator: administrator,
		Transfer: ibctransfertypes.MsgTransfer{
			SourcePort:       sourcePort,
			SourceChannel:    sourceChannel,
			Token:            token,
			Sender:           sender,
			Receiver:         receiver,
			TimeoutHeight:    timeoutHeight,
			TimeoutTimestamp: timeoutTimestamp,
			Memo:             memo,
		},
	}
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgIbcTransferRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Administrator); err != nil {
		return err
	}
	return msg.Transfer.ValidateBasic()
}

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgIbcTransferRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.Administrator)}
}

// NewSetDenomMetadataRequest  creates a new marker in a proposed state with a given total supply a denomination
func NewSetDenomMetadataRequest(
	metadata banktypes.Metadata, admin sdk.AccAddress, //nolint:interfacer
) *MsgSetDenomMetadataRequest {
	return &MsgSetDenomMetadataRequest{
		Metadata:      metadata,
		Administrator: admin.String(),
	}
}

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

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgSetDenomMetadataRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.Administrator)}
}

// GetFeeAllowanceI returns unpacked FeeAllowance
func (msg MsgGrantAllowanceRequest) GetFeeAllowanceI() (feegranttypes.FeeAllowanceI, error) {
	allowance, ok := msg.Allowance.GetCachedValue().(feegranttypes.FeeAllowanceI)
	if !ok {
		return nil, feegranttypes.ErrNoAllowance.Wrap("failed to get allowance")
	}

	return allowance, nil
}

// NewMsgAddMarkerRequest creates a new marker in a proposed state with a given total supply a denomination
func NewMsgGrantAllowance(
	denom string, admin sdk.AccAddress, grantee sdk.AccAddress, allowance feegranttypes.FeeAllowanceI, //nolint:interfacer
) (*MsgGrantAllowanceRequest, error) {
	msg, ok := allowance.(proto.Message)
	if !ok {
		return nil, sdkerrors.ErrPackAny.Wrapf("cannot proto marshal %T", msg)
	}
	anyMsg, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}

	return &MsgGrantAllowanceRequest{
		Denom:         denom,
		Administrator: admin.String(),
		Grantee:       grantee.String(),
		Allowance:     anyMsg,
	}, nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgGrantAllowanceRequest) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var allowance feegranttypes.FeeAllowanceI
	return unpacker.UnpackAny(msg.Allowance, &allowance)
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgGrantAllowanceRequest) ValidateBasic() error {
	if msg.Denom == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("missing marker denom")
	}
	if msg.Administrator == "" {
		return sdkerrors.ErrInvalidAddress.Wrap("missing administrator address")
	}
	if msg.Grantee == "" {
		return sdkerrors.ErrInvalidAddress.Wrap("missing grantee address")
	}

	allowance, err := msg.GetFeeAllowanceI()
	if err != nil {
		return err
	}

	return allowance.ValidateBasic()
}

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgGrantAllowanceRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.Administrator)}
}

func NewMsgAddFinalizeActivateMarkerRequest(denom string, totalSupply sdkmath.Int, fromAddress sdk.AccAddress, manager sdk.AccAddress, markerType MarkerType, supplyFixed bool, allowGovernanceControl bool, accessGrants []AccessGrant) *MsgAddFinalizeActivateMarkerRequest {
	return &MsgAddFinalizeActivateMarkerRequest{
		Amount:                 sdk.NewCoin(denom, totalSupply),
		Manager:                manager.String(),
		FromAddress:            fromAddress.String(),
		MarkerType:             markerType,
		SupplyFixed:            supplyFixed,
		AllowGovernanceControl: allowGovernanceControl,
		AccessList:             accessGrants,
	}
}

// Route returns the name of the module.
func (msg MsgAddFinalizeActivateMarkerRequest) Route() string { return ModuleName }

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgAddFinalizeActivateMarkerRequest) ValidateBasic() error {
	markerCoin := sdk.Coin{
		Denom:  msg.Amount.Denom,
		Amount: msg.Amount.Amount,
	}
	// IsValid validates denom and amount is not negative.
	if !markerCoin.IsValid() {
		return fmt.Errorf("invalid marker denom/total supply: %w", sdkerrors.ErrInvalidCoins)
	}

	_, err := sdk.AccAddressFromBech32(msg.Manager)
	if err != nil {
		return err
	}

	// since this is a one shot process should have 1 access list member, to have any value for a marker
	if msg.AccessList == nil || len(msg.AccessList) == 0 {
		return fmt.Errorf("since this will activate the marker, must have access list defined")
	}
	return nil
}

// GetSignBytes encodes the message for signing.
func (msg MsgAddFinalizeActivateMarkerRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgAddFinalizeActivateMarkerRequest) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(msg.FromAddress)
	return []sdk.AccAddress{addr}
}

// NewMsgAddMarkerProposalRequest creates a new proposal request to add marker
func NewMsgAddMarkerProposalRequest(
	denom string,
	totalSupply sdkmath.Int,
	manager sdk.AccAddress,
	status MarkerStatus,
	markerType MarkerType,
	access []AccessGrant,
	fixed bool,
	allowGov bool,
	authority string, //nolint:interfacer
) *MsgAddMarkerProposalRequest {
	return &MsgAddMarkerProposalRequest{
		Amount:                 sdk.NewCoin(denom, totalSupply),
		Manager:                manager.String(),
		Status:                 status,
		MarkerType:             markerType,
		AccessList:             access,
		SupplyFixed:            fixed,
		AllowGovernanceControl: allowGov,
		Authority:              authority,
	}
}

func (amp MsgAddMarkerProposalRequest) ValidateBasic() error {
	if amp.Status == StatusUndefined {
		return ErrInvalidMarkerStatus
	}
	// A proposed marker must have a manager assigned to allow updates to be made by the caller.
	if len(amp.Manager) == 0 && amp.Status == StatusProposed {
		return fmt.Errorf("marker manager cannot be empty when creating a proposed marker")
	}
	testCoin := sdk.Coin{
		Denom:  amp.Amount.Denom,
		Amount: amp.Amount.Amount,
	}
	if !testCoin.IsValid() {
		return fmt.Errorf("invalid marker denom/total supply: %w", sdkerrors.ErrInvalidCoins)
	}

	return nil
}

// GetSigners indicates that the message must have been signed by the address provided.
func (amp MsgAddMarkerProposalRequest) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(amp.Authority)
	return []sdk.AccAddress{addr}
}
