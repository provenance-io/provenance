package ibchooks

import (
	"encoding/json"
	"fmt"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"

	"github.com/provenance-io/provenance/x/ibchooks/keeper"
	"github.com/provenance-io/provenance/x/ibchooks/types"
)

type WasmHooks struct {
	ContractKeeper      *wasmkeeper.Keeper
	ibcHooksKeeper      *keeper.Keeper
	bech32PrefixAccAddr string
}

func NewWasmHooks(ibcHooksKeeper *keeper.Keeper, contractKeeper *wasmkeeper.Keeper, bech32PrefixAccAddr string) WasmHooks {
	return WasmHooks{
		ContractKeeper:      contractKeeper,
		ibcHooksKeeper:      ibcHooksKeeper,
		bech32PrefixAccAddr: bech32PrefixAccAddr,
	}
}

// ProperlyConfigured returns false when wasm hooks are configured incorrectly
func (h WasmHooks) ProperlyConfigured() bool {
	return h.ContractKeeper != nil && h.ibcHooksKeeper != nil
}

func (h WasmHooks) OnRecvPacketOverride(im IBCMiddleware, ctx sdk.Context, channelVersion string, packet channeltypes.Packet, relayer sdk.AccAddress) ibcexported.Acknowledgement {
	if !h.ProperlyConfigured() {
		return im.App.OnRecvPacket(ctx, channelVersion, packet, relayer)
	}
	isIcs20, data := isIcs20Packet(packet.GetData())
	if !isIcs20 {
		return im.App.OnRecvPacket(ctx, channelVersion, packet, relayer)
	}

	isWasmRouted, contractAddr, msgBytes, err := ValidateAndParseMemo(data.GetMemo(), data.Receiver)
	if !isWasmRouted {
		return im.App.OnRecvPacket(ctx, channelVersion, packet, relayer)
	}
	if err != nil {
		return NewEmitErrorAcknowledgement(ctx, types.ErrMsgValidation, err.Error())
	}
	if msgBytes == nil || contractAddr == nil { // This should never happen
		return NewEmitErrorAcknowledgement(ctx, types.ErrMsgValidation)
	}

	// Calculate the receiver / contract caller based on the packet's channel and sender
	channel := packet.GetDestChannel()
	sender := data.GetSender()
	senderBech32, err := keeper.DeriveIntermediateSender(channel, sender, h.bech32PrefixAccAddr)
	if err != nil {
		return NewEmitErrorAcknowledgement(ctx, types.ErrBadSender, fmt.Sprintf("cannot convert sender address %s/%s to bech32: %s", channel, sender, err.Error()))
	}

	// The funds sent on this packet need to be transferred to the intermediary account for the sender.
	// For this, we override the ICS20 packet's Receiver (essentially hijacking the funds to this new address)
	// and execute the underlying OnRecvPacket() call (which should eventually land on the transfer app's
	// relay.go and send the funds to the intermediary account.
	//
	// If that succeeds, we make the contract call
	data.Receiver = senderBech32
	bz, err := json.Marshal(data)
	if err != nil {
		return NewEmitErrorAcknowledgement(ctx, types.ErrMarshaling, err.Error())
	}
	packet.Data = bz

	// Execute the receive
	ack := im.App.OnRecvPacket(ctx, channelVersion, packet, relayer)
	if !ack.Success() {
		return ack
	}

	amount, ok := sdkmath.NewIntFromString(data.GetAmount())
	if !ok {
		// This should never happen, as it should've been caught in the underlaying call to OnRecvPacket,
		// but returning here for completeness
		return NewEmitErrorAcknowledgement(ctx, types.ErrInvalidPacket, "Amount is not an int")
	}

	// The packet's denom is the denom in the sender chain. This needs to be converted to the local denom.
	denom, err := ExtractDenomFromPacketOnRecv(packet)
	if err != nil {
		return NewEmitErrorAcknowledgement(ctx, types.ErrInvalidPacket, err.Error())
	}
	funds := sdk.NewCoins(sdk.NewCoin(denom, amount))

	// Execute the contract
	execMsg := wasmtypes.MsgExecuteContract{
		Sender:   senderBech32,
		Contract: contractAddr.String(),
		Msg:      msgBytes,
		Funds:    funds,
	}
	response, err := h.execWasmMsg(ctx, &execMsg)
	if err != nil {
		return NewEmitErrorAcknowledgement(ctx, types.ErrWasmError, err.Error())
	}

	fullAck := types.ContractAck{ContractResult: response.Data, IbcAck: ack.Acknowledgement()}
	bz, err = json.Marshal(fullAck)
	if err != nil {
		return NewEmitErrorAcknowledgement(ctx, types.ErrBadResponse, err.Error())
	}

	return channeltypes.NewResultAcknowledgement(bz)
}

func (h WasmHooks) execWasmMsg(ctx sdk.Context, execMsg *wasmtypes.MsgExecuteContract) (*wasmtypes.MsgExecuteContractResponse, error) {
	if err := execMsg.ValidateBasic(); err != nil {
		return nil, fmt.Errorf(types.ErrBadExecutionMsg, err)
	}
	wasmMsgServer := wasmkeeper.NewMsgServerImpl(h.ContractKeeper)
	return wasmMsgServer.ExecuteContract(ctx, execMsg)
}

func isIcs20Packet(data []byte) (isIcs20 bool, ics20data transfertypes.FungibleTokenPacketData) {
	var packetdata transfertypes.FungibleTokenPacketData
	if err := json.Unmarshal(data, &packetdata); err != nil {
		return false, packetdata
	}
	return true, packetdata
}

// JsonStringHasKey parses the memo as a json object and checks if it contains the key.
func JsonStringHasKey(memo, key string) (found bool, jsonObject map[string]interface{}) {
	jsonObject = make(map[string]interface{})

	// If there is no memo, the packet was either sent with an earlier version of IBC, or the memo was
	// intentionally left blank. Nothing to do here. Ignore the packet and pass it down the stack.
	if len(memo) == 0 {
		return false, jsonObject
	}

	// the jsonObject must be a valid JSON object
	err := json.Unmarshal([]byte(memo), &jsonObject)
	if err != nil {
		return false, jsonObject
	}

	// If the key doesn't exist, there's nothing to do on this hook. Continue by passing the packet
	// down the stack
	_, ok := jsonObject[key]
	if !ok {
		return false, jsonObject
	}

	return true, jsonObject
}

func ValidateAndParseMemo(memo string, receiver string) (isWasmRouted bool, contractAddr sdk.AccAddress, msgBytes []byte, err error) {
	isWasmRouted, metadata := JsonStringHasKey(memo, "wasm")
	if !isWasmRouted {
		return isWasmRouted, sdk.AccAddress{}, nil, nil
	}

	wasmRaw := metadata["wasm"]

	// Make sure the wasm key is a map. If it isn't, ignore this packet
	wasm, ok := wasmRaw.(map[string]interface{})
	if !ok {
		return isWasmRouted, sdk.AccAddress{}, nil,
			fmt.Errorf(types.ErrBadMetadataFormatMsg, memo, "wasm metadata is not a valid JSON map object")
	}

	// Get the contract
	contract, ok := wasm["contract"].(string)
	if !ok {
		// The tokens will be returned
		return isWasmRouted, sdk.AccAddress{}, nil,
			fmt.Errorf(types.ErrBadMetadataFormatMsg, memo, `Could not find key wasm["contract"]`)
	}

	contractAddr, err = sdk.AccAddressFromBech32(contract)
	if err != nil {
		return isWasmRouted, sdk.AccAddress{}, nil,
			fmt.Errorf(types.ErrBadMetadataFormatMsg, memo, `wasm["contract"] is not a valid bech32 address`)
	}

	// The contract and the receiver should be the same for the packet to be valid
	if contract != receiver {
		return isWasmRouted, sdk.AccAddress{}, nil,
			fmt.Errorf(types.ErrBadMetadataFormatMsg, memo, `wasm["contract"] should be the same as the receiver of the packet`)
	}

	// Ensure the message key is provided
	if wasm["msg"] == nil {
		return isWasmRouted, sdk.AccAddress{}, nil,
			fmt.Errorf(types.ErrBadMetadataFormatMsg, memo, `Could not find key wasm["msg"]`)
	}

	// Make sure the msg key is a map. If it isn't, return an error
	_, ok = wasm["msg"].(map[string]interface{})
	if !ok {
		return isWasmRouted, sdk.AccAddress{}, nil,
			fmt.Errorf(types.ErrBadMetadataFormatMsg, memo, `wasm["msg"] is not a map object`)
	}

	// Get the message string by serializing the map
	msgBytes, err = json.Marshal(wasm["msg"])
	if err != nil {
		// The tokens will be returned
		return isWasmRouted, sdk.AccAddress{}, nil,
			fmt.Errorf(types.ErrBadMetadataFormatMsg, memo, err.Error())
	}

	return isWasmRouted, contractAddr, msgBytes, nil
}

func (h WasmHooks) SendPacketOverride(
	i ICS4Middleware,
	ctx sdk.Context,
	sourcePort string,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (uint64, error) {
	isIcs20, ics20Packet := isIcs20Packet(data)
	if !isIcs20 {
		return i.channel.SendPacket(ctx, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data) // continue
	}

	isCallbackRouted, metadata := JsonStringHasKey(ics20Packet.GetMemo(), types.IBCCallbackKey)
	if !isCallbackRouted {
		return i.channel.SendPacket(ctx, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data) // continue
	}

	// We remove the callback metadata from the memo as it has already been processed.

	// If the only available key in the memo is the callback, we should remove the memo
	// from the data completely so the packet is sent without it.
	// This way receiver chains that are on old versions of IBC will be able to process the packet

	callbackRaw := metadata[types.IBCCallbackKey] // This will be used later.
	delete(metadata, types.IBCCallbackKey)
	bzMetadata, err := json.Marshal(metadata)
	if err != nil {
		return 0, sdkerrors.Wrap(err, "ibc_callback marshall error")
	}
	stringMetadata := string(bzMetadata)
	if stringMetadata == "{}" {
		ics20Packet.Memo = ""
	} else {
		ics20Packet.Memo = stringMetadata
	}
	dataBytes, err := json.Marshal(ics20Packet)
	if err != nil {
		return 0, sdkerrors.Wrap(err, "ics20data marshall error")
	}

	seq, err := i.channel.SendPacket(ctx, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, dataBytes)
	if err != nil {
		return 0, err
	}

	// Make sure the callback contract is a string and a valid bech32 addr. If it isn't, ignore this packet
	contract, ok := callbackRaw.(string)
	if !ok {
		return seq, nil
	}

	if _, err := sdk.AccAddressFromBech32(contract); err != nil {
		return seq, nil
	}

	h.ibcHooksKeeper.StorePacketCallback(ctx, sourceChannel, seq, contract)
	return seq, nil
}

func (h WasmHooks) GetWasmSendPacketPreProcessor(
	_ sdk.Context,
	data []byte,
	processData map[string]interface{},
) ([]byte, error) {
	isIcs20, ics20Packet := isIcs20Packet(data)
	if !isIcs20 {
		return data, nil
	}

	isCallbackRouted, metadata := JsonStringHasKey(ics20Packet.GetMemo(), types.IBCCallbackKey)
	if !isCallbackRouted {
		return data, nil
	}

	// We remove the callback metadata from the memo as it has already been processed.

	// If the only available key in the memo is the callback, we should remove the memo
	// from the data completely so the packet is sent without it.
	// This way receiver chains that are on old versions of IBC will be able to process the packet

	callbackRaw := metadata[types.IBCCallbackKey]
	if callbackRaw != nil {
		contract, ok := callbackRaw.(string)
		if !ok {
			return nil, fmt.Errorf("unable to format callback %v", callbackRaw)
		}

		if _, err := sdk.AccAddressFromBech32(contract); err != nil {
			return nil, fmt.Errorf("invalid bech32 contract address %v: %w", contract, err)
		}
	}

	processData[types.IBCCallbackKey] = callbackRaw

	delete(metadata, types.IBCCallbackKey)
	bzMetadata, err := json.Marshal(metadata)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "ibc_callback marshall error")
	}
	stringMetadata := string(bzMetadata)
	if stringMetadata == "{}" {
		ics20Packet.Memo = ""
	} else {
		ics20Packet.Memo = stringMetadata
	}
	dataBytes, err := json.Marshal(ics20Packet)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "ics20data marshall error")
	}

	return dataBytes, nil
}

func (h WasmHooks) SendPacketAfterHook(ctx sdk.Context,
	_ string,
	sourceChannel string,
	_ clienttypes.Height,
	_ uint64,
	_ []byte,
	sequence uint64,
	err error,
	processData map[string]interface{},
) {
	if err != nil {
		return
	}
	callbackRaw := processData[types.IBCCallbackKey]
	if callbackRaw == nil {
		return
	}
	// Make sure the callback contract is a string and a valid bech32 addr. If it isn't, ignore this packet
	contract, ok := callbackRaw.(string)
	if !ok {
		return
	}

	if _, err := sdk.AccAddressFromBech32(contract); err != nil {
		return
	}

	h.ibcHooksKeeper.StorePacketCallback(ctx, sourceChannel, sequence, contract)
}

func (h WasmHooks) OnAcknowledgementPacketOverride(im IBCMiddleware, ctx sdk.Context, channelVersion string, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress) error {
	err := im.App.OnAcknowledgementPacket(ctx, channelVersion, packet, acknowledgement, relayer)
	if err != nil {
		return err
	}

	if !h.ProperlyConfigured() {
		// Not configured. Return from the underlying implementation
		return nil
	}

	contract := h.ibcHooksKeeper.GetPacketCallback(ctx, packet.GetSourceChannel(), packet.GetSequence())
	if contract == "" {
		// No callback configured
		return nil
	}

	contractAddr, err := sdk.AccAddressFromBech32(contract)
	if err != nil {
		return sdkerrors.Wrap(err, "Ack callback error")
	}

	success := !IsJSONAckError(acknowledgement)

	// Notify the sender that the ack has been received
	ackAsJSON, err := json.Marshal(acknowledgement)
	if err != nil {
		return err
	}

	sudoMsg, err := json.Marshal(types.NewIbcLifecycleCompleteAck(packet.SourceChannel, packet.Sequence, ackAsJSON, success))
	if err != nil {
		return err
	}
	_, err = h.ContractKeeper.Sudo(ctx, contractAddr, sudoMsg)
	if err != nil {
		// No need to delete the packet callback on error because it would just get reverted.
		return sdkerrors.Wrap(err, "Ack callback error")
	}
	h.ibcHooksKeeper.DeletePacketCallback(ctx, packet.GetSourceChannel(), packet.GetSequence())
	return nil
}

func (h WasmHooks) OnTimeoutPacketOverride(im IBCMiddleware, ctx sdk.Context, channelVersion string, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	err := im.App.OnTimeoutPacket(ctx, channelVersion, packet, relayer)
	if err != nil {
		return err
	}

	if !h.ProperlyConfigured() {
		return nil
	}

	contract := h.ibcHooksKeeper.GetPacketCallback(ctx, packet.GetSourceChannel(), packet.GetSequence())
	if contract == "" {
		return nil
	}

	contractAddr, err := sdk.AccAddressFromBech32(contract)
	if err != nil {
		return sdkerrors.Wrap(err, "Timeout callback error")
	}

	sudoMsg, err := json.Marshal(types.NewIbcLifecycleCompleteTimeout(packet.SourceChannel, packet.Sequence))
	if err != nil {
		return err
	}
	_, err = h.ContractKeeper.Sudo(ctx, contractAddr, sudoMsg)
	if err != nil {
		// error processing the callback. This could be because the contract doesn't implement the message type to
		// process the callback. Retrying this will not help, so we can delete the callback from storage.
		// Since the packet has timed out, we don't expect any other responses that may trigger the callback.
		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				"ibc-timeout-callback-error",
				sdk.NewAttribute("contract", contractAddr.String()),
				sdk.NewAttribute("message", string(sudoMsg)),
				sdk.NewAttribute("error", err.Error()),
			),
		})
	}
	h.ibcHooksKeeper.DeletePacketCallback(ctx, packet.GetSourceChannel(), packet.GetSequence())
	return nil
}

// NewEmitErrorAcknowledgement creates a new error acknowledgement after having emitted an event with the
// details of the error.
func NewEmitErrorAcknowledgement(ctx sdk.Context, err error, errorContexts ...string) channeltypes.Acknowledgement {
	errorType := "ibc-acknowledgement-error"
	logger := ctx.Logger().With("module", errorType)

	attributes := make([]sdk.Attribute, len(errorContexts)+1)
	attributes[0] = sdk.NewAttribute("error", err.Error())
	for i, s := range errorContexts {
		attributes[i+1] = sdk.NewAttribute("error-context", s)
		logger.Error(fmt.Sprintf("error-context: %v", s))
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			errorType,
			attributes...,
		),
	})

	return channeltypes.NewErrorAcknowledgement(err)
}

// IsJSONAckError checks an IBC acknowledgement to see if it's an error.
// This is a replacement for ack.Success() which is currently not working on some circumstances
func IsJSONAckError(acknowledgement []byte) bool {
	var ackErr channeltypes.Acknowledgement_Error
	if err := json.Unmarshal(acknowledgement, &ackErr); err == nil && len(ackErr.Error) > 0 {
		return true
	}
	return false
}

// MustExtractDenomFromPacketOnRecv takes a packet with a valid ICS20 token data in the Data field and returns the
// denom as represented in the local chain.
// If the data cannot be unmarshalled this function will panic.
func ExtractDenomFromPacketOnRecv(packet ibcexported.PacketI) (string, error) {
	var data transfertypes.FungibleTokenPacketData
	if err := json.Unmarshal(packet.GetData(), &data); err != nil {
		return "", fmt.Errorf("unable to unmarshal ICS20 packet data: %w", err)
	}

	denom := transfertypes.ExtractDenomFromPath(data.Denom)
	if denom.HasPrefix(packet.GetSourcePort(), packet.GetSourceChannel()) {
		// Token originally came from this chain; strip the source hop to recover the local denom.
		denom.Trace = denom.Trace[1:]
		if denom.IsNative() {
			return denom.Base, nil
		}
		return denom.IBCDenom(), nil
	}
	// Token came from the source chain; prepend the dest port/channel hop.
	return transfertypes.NewDenom(denom.Base,
		append([]transfertypes.Hop{transfertypes.NewHop(packet.GetDestPort(), packet.GetDestChannel())}, denom.Trace...)...,
	).IBCDenom(), nil
}
