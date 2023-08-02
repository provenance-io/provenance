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

// allRequestMsgs defines all the Msg*Request messages.
var allRequestMsgs = []sdk.Msg{
	(*MsgAddMarkerRequest)(nil),
	(*MsgAddAccessRequest)(nil),
	(*MsgDeleteAccessRequest)(nil),
	(*MsgFinalizeRequest)(nil),
	(*MsgActivateRequest)(nil),
	(*MsgCancelRequest)(nil),
	(*MsgDeleteRequest)(nil),
	(*MsgMintRequest)(nil),
	(*MsgBurnRequest)(nil),
	(*MsgWithdrawRequest)(nil),
	(*MsgTransferRequest)(nil),
	(*MsgIbcTransferRequest)(nil),
	(*MsgGrantAllowanceRequest)(nil),
	(*MsgAddFinalizeActivateMarkerRequest)(nil),
	(*MsgSupplyIncreaseProposalRequest)(nil),
	(*MsgUpdateRequiredAttributesRequest)(nil),
	(*MsgUpdateForcedTransferRequest)(nil),
	(*MsgSetAccountDataRequest)(nil),
	(*MsgUpdateSendDenyListRequest)(nil),
}

// NewMsgAddMarkerRequest creates a new marker in a proposed state with a given total supply a denomination
func NewMsgAddMarkerRequest(
	denom string,
	totalSupply sdkmath.Int,
	fromAddress sdk.AccAddress,
	manager sdk.AccAddress,
	markerType MarkerType,
	supplyFixed bool,
	allowGovernanceControl bool,
	allowForcedTransfer bool,
	requiredAttributes []string,
	// markerNetAssetValues []*MarkerNetAssetValue,
) *MsgAddMarkerRequest {
	return &MsgAddMarkerRequest{
		Amount:                 sdk.NewCoin(denom, totalSupply),
		Manager:                manager.String(),
		FromAddress:            fromAddress.String(),
		Status:                 StatusProposed,
		MarkerType:             markerType,
		SupplyFixed:            supplyFixed,
		AllowGovernanceControl: allowGovernanceControl,
		AllowForcedTransfer:    allowForcedTransfer,
		RequiredAttributes:     requiredAttributes,
		// MarkerNetAssetValues:   markerNetAssetValues,
	}
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgAddMarkerRequest) ValidateBasic() error {
	// A proposed marker must have a manager assigned to allow updates to be made by the caller.
	if len(msg.Manager) == 0 && msg.Status == StatusProposed {
		return fmt.Errorf("marker manager cannot be empty when creating a proposed marker")
	}
	testCoin := sdk.Coin{
		Denom:  msg.Amount.Denom,
		Amount: msg.Amount.Amount,
	}
	if !testCoin.IsValid() {
		return fmt.Errorf("invalid marker denom/total supply: %w", sdkerrors.ErrInvalidCoins)
	}

	if msg.AllowForcedTransfer && msg.MarkerType != MarkerType_RestrictedCoin {
		return fmt.Errorf("forced transfer is only available for restricted coins")
	}

	if len(msg.RequiredAttributes) > 0 && msg.MarkerType != MarkerType_RestrictedCoin {
		return fmt.Errorf("required attributes are reserved for restricted markers")
	}

	if len(msg.RequiredAttributes) > 0 {
		seen := make(map[string]bool)
		for _, str := range msg.RequiredAttributes {
			if seen[str] {
				return fmt.Errorf("required attribute list contains duplicate entries")
			}
			seen[str] = true
		}
	}

	if len(msg.MarkerNetAssetValues) > 0 {
		seen := make(map[string]bool)
		for _, nav := range msg.MarkerNetAssetValues {
			if seen[nav.Value.Denom] {
				return fmt.Errorf("net asset values contain duplicate %q denom", nav.Value.Denom)
			}
			seen[nav.Value.Denom] = true
		}
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
func NewMsgAddAccessRequest(denom string, admin sdk.AccAddress, access AccessGrant) *MsgAddAccessRequest {
	return &MsgAddAccessRequest{
		Denom:         denom,
		Administrator: admin.String(),
		Access:        []AccessGrant{access},
	}
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgAddAccessRequest) ValidateBasic() error {
	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return err
	}
	return ValidateGrants(msg.Access...)
}

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgAddAccessRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.Administrator)}
}

// NewDeleteAccessRequest
func NewDeleteAccessRequest(denom string, admin sdk.AccAddress, removed sdk.AccAddress) *MsgDeleteAccessRequest {
	return &MsgDeleteAccessRequest{
		Denom:          denom,
		Administrator:  admin.String(),
		RemovedAddress: removed.String(),
	}
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgDeleteAccessRequest) ValidateBasic() error {
	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return err
	}
	_, err := sdk.AccAddressFromBech32(msg.RemovedAddress)
	return err
}

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgDeleteAccessRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.Administrator)}
}

// NewMsgFinalizeRequest
func NewMsgFinalizeRequest(denom string, admin sdk.AccAddress) *MsgFinalizeRequest {
	return &MsgFinalizeRequest{
		Denom:         denom,
		Administrator: admin.String(),
	}
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgFinalizeRequest) ValidateBasic() error {
	return sdk.ValidateDenom(msg.Denom)
}

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgFinalizeRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.Administrator)}
}

// NewMsgActivateRequest
func NewMsgActivateRequest(denom string, admin sdk.AccAddress) *MsgActivateRequest {
	return &MsgActivateRequest{
		Denom:         denom,
		Administrator: admin.String(),
	}
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgActivateRequest) ValidateBasic() error {
	return sdk.ValidateDenom(msg.Denom)
}

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgActivateRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.Administrator)}
}

// NewMsgCancelRequest
func NewMsgCancelRequest(denom string, admin sdk.AccAddress) *MsgCancelRequest {
	return &MsgCancelRequest{
		Denom:         denom,
		Administrator: admin.String(),
	}
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgCancelRequest) ValidateBasic() error {
	return sdk.ValidateDenom(msg.Denom)
}

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgCancelRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.Administrator)}
}

// NewMsgDeleteRequest
func NewMsgDeleteRequest(denom string, admin sdk.AccAddress) *MsgDeleteRequest {
	return &MsgDeleteRequest{
		Denom:         denom,
		Administrator: admin.String(),
	}
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgDeleteRequest) ValidateBasic() error {
	return sdk.ValidateDenom(msg.Denom)
}

// GetSigners indicates that the message must have been signed by the address provided.
func (msg MsgDeleteRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.Administrator)}
}

// NewMsgMintRequest creates a mint supply message
func NewMsgMintRequest(admin sdk.AccAddress, amount sdk.Coin) *MsgMintRequest {
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
func NewMsgBurnRequest(admin sdk.AccAddress, amount sdk.Coin) *MsgBurnRequest {
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
	admin, fromAddress, toAddress sdk.AccAddress, amount sdk.Coin,
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
	timeoutTimestamp uint64,
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
	metadata banktypes.Metadata, admin sdk.AccAddress,
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
	denom string, admin sdk.AccAddress, grantee sdk.AccAddress, allowance feegranttypes.FeeAllowanceI,
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

func NewMsgAddFinalizeActivateMarkerRequest(
	denom string,
	totalSupply sdkmath.Int,
	fromAddress sdk.AccAddress,
	manager sdk.AccAddress,
	markerType MarkerType,
	supplyFixed bool,
	allowGovernanceControl bool,
	allowForcedTransfer bool,
	requiredAttributes []string,
	accessGrants []AccessGrant,
) *MsgAddFinalizeActivateMarkerRequest {
	return &MsgAddFinalizeActivateMarkerRequest{
		Amount:                 sdk.NewCoin(denom, totalSupply),
		Manager:                manager.String(),
		FromAddress:            fromAddress.String(),
		MarkerType:             markerType,
		SupplyFixed:            supplyFixed,
		AllowGovernanceControl: allowGovernanceControl,
		AccessList:             accessGrants,
		AllowForcedTransfer:    allowForcedTransfer,
		RequiredAttributes:     requiredAttributes,
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

	if msg.AllowForcedTransfer && msg.MarkerType != MarkerType_RestrictedCoin {
		return fmt.Errorf("forced transfer is only available for restricted coins")
	}

	if len(msg.RequiredAttributes) > 0 && msg.MarkerType != MarkerType_RestrictedCoin {
		return fmt.Errorf("required attributes are reserved for restricted markers")
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

func NewMsgSupplyIncreaseProposalRequest(amount sdk.Coin, targetAddress string, authority string) *MsgSupplyIncreaseProposalRequest {
	return &MsgSupplyIncreaseProposalRequest{
		Amount:        amount,
		TargetAddress: targetAddress,
		Authority:     authority,
	}
}

func (msg *MsgSupplyIncreaseProposalRequest) ValidateBasic() error {
	err := msg.Amount.Validate()
	if err != nil {
		return err
	}

	_, err = sdk.AccAddressFromBech32(msg.TargetAddress)
	if err != nil {
		return err
	}

	_, err = sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return err
	}

	return nil
}

func (msg *MsgSupplyIncreaseProposalRequest) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

// NewMsgUpdateRequiredAttributesRequest creates a MsgUpdateRequiredAttributesRequest
func NewMsgUpdateRequiredAttributesRequest(denom string, transferAuthority sdk.AccAddress, removeRequiredAttributes, addRequiredAttributes []string) *MsgUpdateRequiredAttributesRequest {
	return &MsgUpdateRequiredAttributesRequest{
		Denom:                    denom,
		TransferAuthority:        transferAuthority.String(),
		RemoveRequiredAttributes: removeRequiredAttributes,
		AddRequiredAttributes:    addRequiredAttributes,
	}
}

func (msg MsgUpdateRequiredAttributesRequest) ValidateBasic() error {
	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return err
	}
	if len(msg.AddRequiredAttributes) == 0 && len(msg.RemoveRequiredAttributes) == 0 {
		return fmt.Errorf("both add and remove lists cannot be empty")
	}

	combined := []string{}
	combined = append(combined, msg.AddRequiredAttributes...)
	combined = append(combined, msg.RemoveRequiredAttributes...)
	seen := make(map[string]bool)
	for _, str := range combined {
		if seen[str] {
			return fmt.Errorf("required attribute lists contain duplicate entries")
		}
		seen[str] = true
	}

	_, err := sdk.AccAddressFromBech32(msg.TransferAuthority)
	return err
}

func (msg *MsgUpdateRequiredAttributesRequest) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(msg.TransferAuthority)
	return []sdk.AccAddress{addr}
}

func NewMsgUpdateForcedTransferRequest(denom string, allowForcedTransfer bool, authority sdk.AccAddress) *MsgUpdateForcedTransferRequest {
	return &MsgUpdateForcedTransferRequest{
		Denom:               denom,
		AllowForcedTransfer: allowForcedTransfer,
		Authority:           authority.String(),
	}
}

func (msg MsgUpdateForcedTransferRequest) ValidateBasic() error {
	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return err
	}
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return fmt.Errorf("invalid authority: %w", err)
	}
	return nil
}

func (msg MsgUpdateForcedTransferRequest) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

func (msg MsgSetAccountDataRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		return fmt.Errorf("invalid signer: %w", err)
	}
	if len(msg.Denom) == 0 {
		return errors.New("invalid denom: empty denom string is not allowed")
	}
	return sdk.ValidateDenom(msg.Denom)
}

func (msg MsgSetAccountDataRequest) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(msg.Signer)
	return []sdk.AccAddress{addr}
}

// NewMsgUpdateSendDenyListRequest creates a NewMsgUpdateSendDenyListRequest
func NewMsgUpdateSendDenyListRequest(denom string, authority sdk.AccAddress, removeDenyAddresses, addDenyAddresses []string) *MsgUpdateSendDenyListRequest {
	return &MsgUpdateSendDenyListRequest{
		Denom:                 denom,
		Authority:             authority.String(),
		RemoveDeniedAddresses: removeDenyAddresses,
		AddDeniedAddresses:    addDenyAddresses,
	}
}

func (msg MsgUpdateSendDenyListRequest) ValidateBasic() error {
	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return err
	}
	if len(msg.AddDeniedAddresses) == 0 && len(msg.RemoveDeniedAddresses) == 0 {
		return fmt.Errorf("both add and remove lists cannot be empty")
	}

	combined := []string{}
	combined = append(combined, msg.AddDeniedAddresses...)
	combined = append(combined, msg.RemoveDeniedAddresses...)
	seen := make(map[string]bool)
	for _, addr := range combined {
		if _, err := sdk.AccAddressFromBech32(addr); err != nil {
			return err
		}
		if seen[addr] {
			return fmt.Errorf("denied address lists contain duplicate entries")
		}
		seen[addr] = true
	}

	_, err := sdk.AccAddressFromBech32(msg.Authority)
	return err
}

func (msg *MsgUpdateSendDenyListRequest) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}
