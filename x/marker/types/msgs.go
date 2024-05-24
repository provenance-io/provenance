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

func NewMsgFinalizeRequest(denom string, admin sdk.AccAddress) *MsgFinalizeRequest {
	return &MsgFinalizeRequest{
		Denom:         denom,
		Administrator: admin.String(),
	}
}

func (msg MsgFinalizeRequest) ValidateBasic() error {
	return sdk.ValidateDenom(msg.Denom)
}

func NewMsgActivateRequest(denom string, admin sdk.AccAddress) *MsgActivateRequest {
	return &MsgActivateRequest{
		Denom:         denom,
		Administrator: admin.String(),
	}
}

func (msg MsgActivateRequest) ValidateBasic() error {
	return sdk.ValidateDenom(msg.Denom)
}

func NewMsgCancelRequest(denom string, admin sdk.AccAddress) *MsgCancelRequest {
	return &MsgCancelRequest{
		Denom:         denom,
		Administrator: admin.String(),
	}
}

func (msg MsgCancelRequest) ValidateBasic() error {
	return sdk.ValidateDenom(msg.Denom)
}

func NewMsgDeleteRequest(denom string, admin sdk.AccAddress) *MsgDeleteRequest {
	return &MsgDeleteRequest{
		Denom:         denom,
		Administrator: admin.String(),
	}
}

func (msg MsgDeleteRequest) ValidateBasic() error {
	return sdk.ValidateDenom(msg.Denom)
}

func NewMsgMintRequest(admin sdk.AccAddress, amount sdk.Coin) *MsgMintRequest {
	return &MsgMintRequest{
		Administrator: admin.String(),
		Amount:        amount,
	}
}

func (msg MsgMintRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Administrator); err != nil {
		return err
	}
	return msg.Amount.Validate()
}

func NewMsgBurnRequest(admin sdk.AccAddress, amount sdk.Coin) *MsgBurnRequest {
	return &MsgBurnRequest{
		Administrator: admin.String(),
		Amount:        amount,
	}
}

func (msg MsgBurnRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Administrator); err != nil {
		return err
	}

	return msg.Amount.Validate()
}

func NewMsgAddAccessRequest(denom string, admin sdk.AccAddress, access AccessGrant) *MsgAddAccessRequest {
	return &MsgAddAccessRequest{
		Denom:         denom,
		Administrator: admin.String(),
		Access:        []AccessGrant{access},
	}
}

func (msg MsgAddAccessRequest) ValidateBasic() error {
	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return err
	}
	return ValidateGrants(msg.Access...)
}

func NewDeleteAccessRequest(denom string, admin sdk.AccAddress, removed sdk.AccAddress) *MsgDeleteAccessRequest {
	return &MsgDeleteAccessRequest{
		Denom:          denom,
		Administrator:  admin.String(),
		RemovedAddress: removed.String(),
	}
}

func (msg MsgDeleteAccessRequest) ValidateBasic() error {
	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return err
	}
	_, err := sdk.AccAddressFromBech32(msg.RemovedAddress)
	return err
}

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

func NewSetDenomMetadataRequest(
	metadata banktypes.Metadata, admin sdk.AccAddress,
) *MsgSetDenomMetadataRequest {
	return &MsgSetDenomMetadataRequest{
		Metadata:      metadata,
		Administrator: admin.String(),
	}
}

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

func NewMsgSetAccountDataRequest(denom, value string, signer sdk.AccAddress) *MsgSetAccountDataRequest {
	return &MsgSetAccountDataRequest{
		Denom:  denom,
		Value:  value,
		Signer: signer.String(),
	}
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

func NewMsgAddNetAssetValuesRequest(denom, administrator string, netAssetValues []NetAssetValue) *MsgAddNetAssetValuesRequest {
	return &MsgAddNetAssetValuesRequest{
		Denom:          denom,
		NetAssetValues: netAssetValues,
		Administrator:  administrator,
	}
}

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

func NewMsgSupplyDecreaseProposalRequest(amount sdk.Coin, authority string) *MsgSupplyDecreaseProposalRequest {
	return &MsgSupplyDecreaseProposalRequest{
		Amount:    amount,
		Authority: authority,
	}
}

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

func NewMsgSetAdministratorProposalRequest(denom string, accessGrant []AccessGrant, authority string) *MsgSetAdministratorProposalRequest {
	return &MsgSetAdministratorProposalRequest{
		Denom:     denom,
		Access:    accessGrant,
		Authority: authority,
	}
}

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

func NewMsgRemoveAdministratorProposalRequest(denom string, removedAddress []string, authority string) *MsgRemoveAdministratorProposalRequest {
	return &MsgRemoveAdministratorProposalRequest{
		Denom:          denom,
		RemovedAddress: removedAddress,
		Authority:      authority,
	}
}

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

func NewMsgChangeStatusProposalRequest(denom string, status MarkerStatus, authority string) *MsgChangeStatusProposalRequest {
	return &MsgChangeStatusProposalRequest{
		Denom:     denom,
		NewStatus: status,
		Authority: authority,
	}
}

func (msg MsgChangeStatusProposalRequest) ValidateBasic() error {
	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return err
	}
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	return err
}

func NewMsgWithdrawEscrowProposalRequest(denom string, amount sdk.Coins, targetAddress, authority string) *MsgWithdrawEscrowProposalRequest {
	return &MsgWithdrawEscrowProposalRequest{
		Denom:         denom,
		Amount:        amount,
		TargetAddress: targetAddress,
		Authority:     authority,
	}
}

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

func NewMsgSetDenomMetadataProposalRequest(metadata banktypes.Metadata, authority string) *MsgSetDenomMetadataProposalRequest {
	return &MsgSetDenomMetadataProposalRequest{
		Metadata:  metadata,
		Authority: authority,
	}
}

func (msg MsgSetDenomMetadataProposalRequest) ValidateBasic() error {
	if err := msg.Metadata.Validate(); err != nil {
		return err
	}
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	return err
}

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

func (msg MsgUpdateParamsRequest) ValidateBasic() error {
	if err := msg.Params.Validate(); err != nil {
		return err
	}
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	return err
}
