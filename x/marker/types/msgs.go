package types

import (
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"
	feegranttypes "cosmossdk.io/x/feegrant"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gogoproto/proto"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
)

// AllRequestMsgs defines all the Msg*Request messages.
var AllRequestMsgs = []sdk.Msg{
	(*MsgFinalizeRequest)(nil),
	(*MsgActivateRequest)(nil),
	(*MsgCancelRequest)(nil),
	(*MsgDeleteAccessRequest)(nil),
	(*MsgMintRequest)(nil),
	(*MsgBurnRequest)(nil),
	(*MsgAddAccessRequest)(nil),
	(*MsgDeleteRequest)(nil),
	(*MsgWithdrawRequest)(nil),
	(*MsgAddMarkerRequest)(nil),
	(*MsgTransferRequest)(nil),
	(*MsgIbcTransferRequest)(nil),
	(*MsgSetDenomMetadataRequest)(nil),
	(*MsgGrantAllowanceRequest)(nil),
	(*MsgRevokeGrantAllowanceRequest)(nil),
	(*MsgAddFinalizeActivateMarkerRequest)(nil),
	(*MsgSupplyIncreaseProposalRequest)(nil),
	(*MsgSupplyDecreaseProposalRequest)(nil),
	(*MsgUpdateRequiredAttributesRequest)(nil),
	(*MsgUpdateForcedTransferRequest)(nil),
	(*MsgSetAccountDataRequest)(nil),
	(*MsgUpdateSendDenyListRequest)(nil),
	(*MsgAddNetAssetValuesRequest)(nil),
	(*MsgSetAdministratorProposalRequest)(nil),
	(*MsgRemoveAdministratorProposalRequest)(nil),
	(*MsgChangeStatusProposalRequest)(nil),
	(*MsgWithdrawEscrowProposalRequest)(nil),
	(*MsgSetDenomMetadataProposalRequest)(nil),
	(*MsgUpdateParamsRequest)(nil),
}

// NewMsgFinalizeRequest creates a new MsgFinalizeRequest instance.
func NewMsgFinalizeRequest(denom string, admin sdk.AccAddress) *MsgFinalizeRequest {
	return &MsgFinalizeRequest{
		Denom:         denom,
		Administrator: admin.String(),
	}
}

// ValidateBasic implements basic validation for MsgFinalizeRequest.
func (msg MsgFinalizeRequest) ValidateBasic() error {
	return sdk.ValidateDenom(msg.Denom)
}

// NewMsgActivateRequest creates a new MsgActivateRequest instance.
func NewMsgActivateRequest(denom string, admin sdk.AccAddress) *MsgActivateRequest {
	return &MsgActivateRequest{
		Denom:         denom,
		Administrator: admin.String(),
	}
}

// ValidateBasic implements basic validation for MsgFinalizeRequest.
func (msg MsgActivateRequest) ValidateBasic() error {
	return sdk.ValidateDenom(msg.Denom)
}

// NewMsgCancelRequest creates a new MsgCancelRequest instance.
func NewMsgCancelRequest(denom string, admin sdk.AccAddress) *MsgCancelRequest {
	return &MsgCancelRequest{
		Denom:         denom,
		Administrator: admin.String(),
	}
}

// ValidateBasic implements basic validation for MsgFinalizeRequest.
func (msg MsgCancelRequest) ValidateBasic() error {
	return sdk.ValidateDenom(msg.Denom)
}

// NewMsgDeleteRequest creates a new MsgDeleteRequest instance.
func NewMsgDeleteRequest(denom string, admin sdk.AccAddress) *MsgDeleteRequest {
	return &MsgDeleteRequest{
		Denom:         denom,
		Administrator: admin.String(),
	}
}

// ValidateBasic implements basic validation for MsgFinalizeRequest.
func (msg MsgDeleteRequest) ValidateBasic() error {
	return sdk.ValidateDenom(msg.Denom)
}

// NewMsgMintRequest creates a new MsgMintRequest instance.
func NewMsgMintRequest(admin sdk.AccAddress, amount sdk.Coin, recipient sdk.AccAddress) *MsgMintRequest {
	msg := &MsgMintRequest{
		Administrator: admin.String(),
		Amount:        amount,
	}
	if !recipient.Empty() {
		msg.Recipient = recipient.String()
	}
	return msg
}

// ValidateBasic implements basic validation for MsgFinalizeRequest.
func (msg MsgMintRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Administrator); err != nil {
		return err
	}
	if msg.Recipient != "" {
		if _, err := sdk.AccAddressFromBech32(msg.Recipient); err != nil {
			return sdkerrors.ErrInvalidAddress.Wrapf("invalid recipient address: %s", msg.Recipient)
		}
	}
	return msg.Amount.Validate()
}

// NewMsgBurnRequest creates a new MsgBurnRequest instance.
func NewMsgBurnRequest(admin sdk.AccAddress, amount sdk.Coin) *MsgBurnRequest {
	return &MsgBurnRequest{
		Administrator: admin.String(),
		Amount:        amount,
	}
}

// ValidateBasic implements basic validation for MsgFinalizeRequest.
func (msg MsgBurnRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Administrator); err != nil {
		return err
	}

	return msg.Amount.Validate()
}

// NewMsgAddAccessRequest creates a new MsgAddAccessRequest instance.
func NewMsgAddAccessRequest(denom string, admin sdk.AccAddress, access AccessGrant) *MsgAddAccessRequest {
	return &MsgAddAccessRequest{
		Denom:         denom,
		Administrator: admin.String(),
		Access:        []AccessGrant{access},
	}
}

// ValidateBasic implements basic validation for MsgAddAccessRequest.
func (msg MsgAddAccessRequest) ValidateBasic() error {
	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return err
	}
	return ValidateGrants(msg.Access...)
}

// NewDeleteAccessRequest creates a new DeleteAccessRequest instance.
func NewDeleteAccessRequest(denom string, admin sdk.AccAddress, removed sdk.AccAddress) *MsgDeleteAccessRequest {
	return &MsgDeleteAccessRequest{
		Denom:          denom,
		Administrator:  admin.String(),
		RemovedAddress: removed.String(),
	}
}

// ValidateBasic implements basic validation for MsgDeleteAccessRequest.
func (msg MsgDeleteAccessRequest) ValidateBasic() error {
	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return err
	}
	_, err := sdk.AccAddressFromBech32(msg.RemovedAddress)
	return err
}

// NewMsgWithdrawRequest creates a new MsgWithdrawRequest instance.
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

// ValidateBasic implements basic validation for MsgWithdrawRequest.
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

// NewMsgAddMarkerRequest creates a new MsgAddMarkerRequest instance.
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
	usdMills uint64,
	volume uint64,
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
		UsdMills:               usdMills,
		Volume:                 volume,
	}
}

// ValidateBasic implements basic validation for MsgAddMarkerRequest.
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

	return nil
}

// NewMsgTransferRequest creates a new MsgTransferRequest instance.
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

// ValidateBasic implements basic validation for MsgTransferRequest.
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

// NewMsgIbcTransferRequest creates a new MsgIbcTransferRequest instance.
func NewMsgIbcTransferRequest(
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

// NewSetDenomMetadataRequest creates a new SetDenomMetadataRequest instance.
func NewSetDenomMetadataRequest(
	metadata banktypes.Metadata, admin sdk.AccAddress,
) *MsgSetDenomMetadataRequest {
	return &MsgSetDenomMetadataRequest{
		Metadata:      metadata,
		Administrator: admin.String(),
	}
}

// ValidateBasic implements basic validation for MsgSetDenomMetadataRequest.
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

// NewMsgGrantAllowance creates a new MsgGrantAllowance instance.
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

// ValidateBasic implements basic validation for MsgGrantAllowanceRequest.
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

// NewMsgRevokeGrantAllowance creates a new MsgRevokeGrantAllowance instance.
func NewMsgRevokeGrantAllowance(denom string, admin sdk.AccAddress, grantee sdk.AccAddress) *MsgRevokeGrantAllowanceRequest {
	return &MsgRevokeGrantAllowanceRequest{
		Denom:         denom,
		Administrator: admin.String(),
		Grantee:       grantee.String(),
	}
}

// ValidateBasic implements basic validation for MsgRevokeGrantAllowanceRequest.
func (msg MsgRevokeGrantAllowanceRequest) ValidateBasic() error {
	if msg.Denom == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("missing marker denom")
	}
	if msg.Administrator == "" {
		return sdkerrors.ErrInvalidAddress.Wrap("missing administrator address")
	}
	if msg.Grantee == "" {
		return sdkerrors.ErrInvalidAddress.Wrap("missing grantee address")
	}

	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces for this MsgGrantAllowanceRequest.
func (msg MsgGrantAllowanceRequest) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var allowance feegranttypes.FeeAllowanceI
	return unpacker.UnpackAny(msg.Allowance, &allowance)
}

// GetFeeAllowanceI returns the unpacked FeeAllowance.
func (msg MsgGrantAllowanceRequest) GetFeeAllowanceI() (feegranttypes.FeeAllowanceI, error) {
	allowance, ok := msg.Allowance.GetCachedValue().(feegranttypes.FeeAllowanceI)
	if !ok {
		return nil, feegranttypes.ErrNoAllowance.Wrap("failed to get allowance")
	}

	return allowance, nil
}

// NewMsgAddFinalizeActivateMarkerRequest creates a new MsgAddFinalizeActivateMarkerRequest instance.
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
	usdMills uint64,
	netAssetVolume uint64,
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
		UsdMills:               usdMills,
		Volume:                 netAssetVolume,
	}
}

// ValidateBasic implements basic validation for MsgAddFinalizeActivateMarkerRequest.
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
	if len(msg.AccessList) == 0 {
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

// NewMsgSupplyIncreaseProposalRequest creates a new MsgSupplyIncreaseProposalRequest instance.
func NewMsgSupplyIncreaseProposalRequest(amount sdk.Coin, targetAddress string, authority string) *MsgSupplyIncreaseProposalRequest {
	return &MsgSupplyIncreaseProposalRequest{
		Amount:        amount,
		TargetAddress: targetAddress,
		Authority:     authority,
	}
}

// ValidateBasic implements basic validation for MsgSupplyIncreaseProposalRequest.
func (msg *MsgSupplyIncreaseProposalRequest) ValidateBasic() error {
	err := msg.Amount.Validate()
	if err != nil {
		return err
	}

	if len(msg.TargetAddress) > 0 {
		_, err = sdk.AccAddressFromBech32(msg.TargetAddress)
		if err != nil {
			return err
		}
	}

	_, err = sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return err
	}

	return nil
}

// NewMsgUpdateRequiredAttributesRequest creates a new MsgUpdateRequiredAttributesRequest instance.
func NewMsgUpdateRequiredAttributesRequest(denom string, transferAuthority sdk.AccAddress, removeRequiredAttributes, addRequiredAttributes []string) *MsgUpdateRequiredAttributesRequest {
	return &MsgUpdateRequiredAttributesRequest{
		Denom:                    denom,
		TransferAuthority:        transferAuthority.String(),
		RemoveRequiredAttributes: removeRequiredAttributes,
		AddRequiredAttributes:    addRequiredAttributes,
	}
}

// ValidateBasic implements basic validation for MsgUpdateRequiredAttributesRequest.
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

// NewMsgUpdateForcedTransferRequest creates a new MsgUpdateForcedTransferRequest instance.
func NewMsgUpdateForcedTransferRequest(denom string, allowForcedTransfer bool, authority sdk.AccAddress) *MsgUpdateForcedTransferRequest {
	return &MsgUpdateForcedTransferRequest{
		Denom:               denom,
		AllowForcedTransfer: allowForcedTransfer,
		Authority:           authority.String(),
	}
}

// ValidateBasic implements basic validation for MsgUpdateForcedTransferRequest.
func (msg MsgUpdateForcedTransferRequest) ValidateBasic() error {
	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return err
	}
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return fmt.Errorf("invalid authority: %w", err)
	}
	return nil
}

// NewMsgSetAccountDataRequest creates a new MsgSetAccountDataRequest instance.
func NewMsgSetAccountDataRequest(denom, value string, signer sdk.AccAddress) *MsgSetAccountDataRequest {
	return &MsgSetAccountDataRequest{
		Denom:  denom,
		Value:  value,
		Signer: signer.String(),
	}
}

// ValidateBasic implements basic validation for MsgSetAccountDataRequest.
func (msg MsgSetAccountDataRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		return fmt.Errorf("invalid signer: %w", err)
	}
	if len(msg.Denom) == 0 {
		return errors.New("invalid denom: empty denom string is not allowed")
	}
	return sdk.ValidateDenom(msg.Denom)
}

// NewMsgUpdateSendDenyListRequest creates a new MsgUpdateSendDenyListRequest instance.
func NewMsgUpdateSendDenyListRequest(denom string, authority sdk.AccAddress, removeDenyAddresses, addDenyAddresses []string) *MsgUpdateSendDenyListRequest {
	return &MsgUpdateSendDenyListRequest{
		Denom:                 denom,
		Authority:             authority.String(),
		RemoveDeniedAddresses: removeDenyAddresses,
		AddDeniedAddresses:    addDenyAddresses,
	}
}

// ValidateBasic implements basic validation for MsgUpdateSendDenyListRequest.
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

// NewMsgAddNetAssetValuesRequest creates a new MsgAddNetAssetValuesRequest with the given parameters.
func NewMsgAddNetAssetValuesRequest(denom, administrator string, netAssetValues []NetAssetValue) *MsgAddNetAssetValuesRequest {
	return &MsgAddNetAssetValuesRequest{
		Denom:          denom,
		NetAssetValues: netAssetValues,
		Administrator:  administrator,
	}
}

// ValidateBasic implements basic validation for MsgAddNetAssetValuesRequest.
func (msg MsgAddNetAssetValuesRequest) ValidateBasic() error {
	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return err
	}

	if len(msg.NetAssetValues) == 0 {
		return fmt.Errorf("net asset value list cannot be empty")
	}

	seen := make(map[string]bool)
	for _, nav := range msg.NetAssetValues {
		if err := nav.Validate(); err != nil {
			return err
		}

		if nav.UpdatedBlockHeight != 0 {
			return fmt.Errorf("marker net asset value must not have update height set")
		}

		if seen[nav.Price.Denom] {
			return fmt.Errorf("list of net asset values contains duplicates")
		}
		seen[nav.Price.Denom] = true
	}

	_, err := sdk.AccAddressFromBech32(msg.Administrator)
	return err
}

// NewMsgSupplyDecreaseProposalRequest creates a new MsgSupplyDecreaseProposalRequest with the given parameters.
func NewMsgSupplyDecreaseProposalRequest(amount sdk.Coin, authority string) *MsgSupplyDecreaseProposalRequest {
	return &MsgSupplyDecreaseProposalRequest{
		Amount:    amount,
		Authority: authority,
	}
}

// ValidateBasic implements basic validation for MsgSupplyDecreaseProposalRequest.
func (msg MsgSupplyDecreaseProposalRequest) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return err
	}
	if msg.Amount.IsNegative() {
		return fmt.Errorf("amount to decrease must be greater than zero")
	}
	return nil
}

// NewMsgSetAdministratorProposalRequest creates a new MsgSetAdministratorProposalRequest with the given parameters.
func NewMsgSetAdministratorProposalRequest(denom string, accessGrant []AccessGrant, authority string) *MsgSetAdministratorProposalRequest {
	return &MsgSetAdministratorProposalRequest{
		Denom:     denom,
		Access:    accessGrant,
		Authority: authority,
	}
}

// ValidateBasic implements basic validation for MsgSetAdministratorProposalRequest.
func (msg MsgSetAdministratorProposalRequest) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return err
	}
	for _, a := range msg.Access {
		if err := a.Validate(); err != nil {
			return fmt.Errorf("invalid access grant for administrator: %w", err)
		}
	}
	return nil
}

// NewMsgRemoveAdministratorProposalRequest creates a new MsgRemoveAdministratorProposalRequest with the given parameters.
func NewMsgRemoveAdministratorProposalRequest(denom string, removedAddress []string, authority string) *MsgRemoveAdministratorProposalRequest {
	return &MsgRemoveAdministratorProposalRequest{
		Denom:          denom,
		RemovedAddress: removedAddress,
		Authority:      authority,
	}
}

// ValidateBasic implements basic validation for MsgRemoveAdministratorProposalRequest.
func (msg MsgRemoveAdministratorProposalRequest) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return err
	}
	for _, ra := range msg.RemovedAddress {
		if _, err := sdk.AccAddressFromBech32(ra); err != nil {
			return fmt.Errorf("administrator account address is invalid: %w", err)
		}
	}
	return nil
}

// NewMsgChangeStatusProposalRequest creates a new MsgChangeStatusProposalRequest with the given parameters.
func NewMsgChangeStatusProposalRequest(denom string, status MarkerStatus, authority string) *MsgChangeStatusProposalRequest {
	return &MsgChangeStatusProposalRequest{
		Denom:     denom,
		NewStatus: status,
		Authority: authority,
	}
}

// ValidateBasic implements basic validation for MsgChangeStatusProposalRequest.
func (msg MsgChangeStatusProposalRequest) ValidateBasic() error {
	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return err
	}
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	return err
}

// NewMsgWithdrawEscrowProposalRequest creates a new MsgWithdrawEscrowProposalRequest with the given parameters.
func NewMsgWithdrawEscrowProposalRequest(denom string, amount sdk.Coins, targetAddress, authority string) *MsgWithdrawEscrowProposalRequest {
	return &MsgWithdrawEscrowProposalRequest{
		Denom:         denom,
		Amount:        amount,
		TargetAddress: targetAddress,
		Authority:     authority,
	}
}

// ValidateBasic implements basic validation for MsgWithdrawEscrowProposalRequest.
func (msg MsgWithdrawEscrowProposalRequest) ValidateBasic() error {
	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return err
	}
	if !msg.Amount.IsValid() {
		return fmt.Errorf("amount is invalid: %v", msg.Amount)
	}
	_, err := sdk.AccAddressFromBech32(msg.TargetAddress)
	if err != nil {
		return err
	}

	_, err = sdk.AccAddressFromBech32(msg.Authority)
	return err
}

// NewMsgSetDenomMetadataProposalRequest creates a new MsgSetDenomMetadataProposalRequest with the given parameters.
func NewMsgSetDenomMetadataProposalRequest(metadata banktypes.Metadata, authority string) *MsgSetDenomMetadataProposalRequest {
	return &MsgSetDenomMetadataProposalRequest{
		Metadata:  metadata,
		Authority: authority,
	}
}

// ValidateBasic implements basic validation for MsgSetDenomMetadataProposalRequest.
func (msg MsgSetDenomMetadataProposalRequest) ValidateBasic() error {
	if err := msg.Metadata.Validate(); err != nil {
		return err
	}
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	return err
}

// NewMsgUpdateParamsRequest creates a new MsgUpdateParamsRequest with the given parameters.
func NewMsgUpdateParamsRequest(
	enableGovernance bool,
	unrestrictedDenomRegex string,
	maxSupply sdkmath.Int,
	authority string,
) *MsgUpdateParamsRequest {
	return &MsgUpdateParamsRequest{
		Authority: authority,
		Params: NewParams(
			enableGovernance,
			unrestrictedDenomRegex,
			maxSupply,
		),
	}
}

// ValidateBasic implements basic validation for MsgUpdateParamsRequest.
func (msg MsgUpdateParamsRequest) ValidateBasic() error {
	if err := msg.Params.Validate(); err != nil {
		return err
	}
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	return err
}
