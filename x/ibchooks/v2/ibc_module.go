package v2

import (
	"encoding/json"
	"fmt"
	"strings"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	channeltypesv2 "github.com/cosmos/ibc-go/v10/modules/core/04-channel/v2/types"
	"github.com/cosmos/ibc-go/v10/modules/core/api"
	ibckeeper "github.com/cosmos/ibc-go/v10/modules/core/keeper"
	tendermintclient "github.com/cosmos/ibc-go/v10/modules/light-clients/07-tendermint"

	ibchooks "github.com/provenance-io/provenance/x/ibchooks"
	"github.com/provenance-io/provenance/x/ibchooks/keeper"
	"github.com/provenance-io/provenance/x/ibchooks/types"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

var _ api.IBCModule = (*IBCModule)(nil)

// IBCModule is the IBC v2 module for ibchooks, providing wasm and marker hook functionality.
type IBCModule struct {
	app                 api.IBCModule
	ibcKeeper           *ibckeeper.Keeper
	ibcHooksKeeper      *keeper.Keeper
	contractKeeper      *wasmkeeper.Keeper
	markerHooks         *ibchooks.MarkerHooks
	bech32PrefixAccAddr string
}

// NewIBCModule creates a new IBC v2 hooks module.
func NewIBCModule(
	app api.IBCModule,
	ibcKeeper *ibckeeper.Keeper,
	ibcHooksKeeper *keeper.Keeper,
	contractKeeper *wasmkeeper.Keeper,
	markerHooks *ibchooks.MarkerHooks,
	bech32PrefixAccAddr string,
) IBCModule {
	return IBCModule{
		app:                 app,
		ibcKeeper:           ibcKeeper,
		ibcHooksKeeper:      ibcHooksKeeper,
		contractKeeper:      contractKeeper,
		markerHooks:         markerHooks,
		bech32PrefixAccAddr: bech32PrefixAccAddr,
	}
}

// properlyConfigured returns true if the module is ready to process hooks.
func (im IBCModule) properlyConfigured() bool {
	return im.ibcKeeper != nil && im.contractKeeper != nil && im.ibcHooksKeeper != nil && im.markerHooks != nil && im.markerHooks.ProperlyConfigured()
}

// OnSendPacket implements api.IBCModule.
func (im IBCModule) OnSendPacket(
	ctx sdk.Context,
	sourceClient string,
	destinationClient string,
	sequence uint64,
	payload channeltypesv2.Payload,
	signer sdk.AccAddress,
) error {
	if !im.properlyConfigured() {
		return im.app.OnSendPacket(ctx, sourceClient, destinationClient, sequence, payload, signer)
	}

	data, err := unmarshalV2PacketData(payload)
	if err != nil {
		// Not an ICS20 packet, pass through.
		return im.app.OnSendPacket(ctx, sourceClient, destinationClient, sequence, payload, signer)
	}

	isCallbackRouted, metadata := ibchooks.JsonStringHasKey(data.GetMemo(), types.IBCCallbackKey)
	if !isCallbackRouted {
		return im.app.OnSendPacket(ctx, sourceClient, destinationClient, sequence, payload, signer)
	}

	// Validate and extract the callback contract address.
	// This matches v1 preprocessor behavior: error on non-string or invalid bech32 values.
	callbackRaw := metadata[types.IBCCallbackKey]
	if callbackRaw != nil {
		contract, ok := callbackRaw.(string)
		if !ok {
			return fmt.Errorf("unable to format callback %v", callbackRaw)
		}
		if _, err := sdk.AccAddressFromBech32(contract); err != nil {
			return fmt.Errorf("invalid bech32 contract address %v: %w", contract, err)
		}
	}

	// Strip the ibc_callback key from the memo before forwarding to the underlying app.
	// This matches v1 preprocessor behavior so receiver chains on older IBC versions
	// can process the packet without choking on the callback metadata.
	delete(metadata, types.IBCCallbackKey)
	strippedMemo, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("ibc_callback marshall error: %w", err)
	}
	if string(strippedMemo) == "{}" {
		data.Memo = ""
	} else {
		data.Memo = string(strippedMemo)
	}

	modifiedPayload, err := marshalV2Payload(data, payload)
	if err != nil {
		return fmt.Errorf("failed to re-encode payload after stripping ibc_callback: %w", err)
	}

	// Pass the modified payload (with ibc_callback stripped) to the underlying app.
	if err := im.app.OnSendPacket(ctx, sourceClient, destinationClient, sequence, modifiedPayload, signer); err != nil {
		return err
	}

	// Store the callback contract for the after-send hook (ack/timeout processing).
	if callbackRaw != nil {
		contract := callbackRaw.(string) // already validated above
		im.ibcHooksKeeper.StorePacketCallback(ctx, sourceClient, sequence, contract)
	}

	return nil
}

// OnRecvPacket implements api.IBCModule.
func (im IBCModule) OnRecvPacket(
	ctx sdk.Context,
	sourceClient string,
	destinationClient string,
	sequence uint64,
	payload channeltypesv2.Payload,
	relayer sdk.AccAddress,
) channeltypesv2.RecvPacketResult {
	if !im.properlyConfigured() {
		return im.app.OnRecvPacket(ctx, sourceClient, destinationClient, sequence, payload, relayer)
	}

	data, err := unmarshalV2PacketData(payload)
	if err != nil {
		// Not an ICS20 packet, pass through.
		return im.app.OnRecvPacket(ctx, sourceClient, destinationClient, sequence, payload, relayer)
	}

	// Run marker hooks: create marker for IBC denom if needed.
	ibcDenom := extractDenomFromV2PacketOnRecv(data, payload, sourceClient, destinationClient)
	if err := im.addMarkerForDenom(ctx, data, ibcDenom, destinationClient); err != nil {
		return errRecvResult(err, types.ErrMarkerError)
	}

	// Check for wasm routing.
	isWasmRouted, contractAddr, msgBytes, err := ibchooks.ValidateAndParseMemo(data.GetMemo(), data.Receiver)
	if !isWasmRouted {
		return im.app.OnRecvPacket(ctx, sourceClient, destinationClient, sequence, payload, relayer)
	}
	if err != nil {
		return errRecvResult(err, types.ErrMsgValidation)
	}
	if msgBytes == nil || contractAddr == nil {
		return errRecvResult(types.ErrMsgValidation, types.ErrMsgValidation)
	}

	// Derive intermediate sender from the destination client and original sender.
	sender := data.GetSender()
	senderBech32, err := keeper.DeriveIntermediateSender(destinationClient, sender, im.bech32PrefixAccAddr)
	if err != nil {
		return errRecvResult(fmt.Errorf("cannot convert sender address %s/%s to bech32: %w", destinationClient, sender, err), types.ErrBadSender)
	}

	// Hijack the receiver to the intermediate sender so the transfer app sends funds there.
	data.Receiver = senderBech32
	modifiedPayload, err := marshalV2Payload(data, payload)
	if err != nil {
		return errRecvResult(err, types.ErrMarshaling)
	}

	// Execute the receive with the modified payload.
	result := im.app.OnRecvPacket(ctx, sourceClient, destinationClient, sequence, modifiedPayload, relayer)
	if result.Status == channeltypesv2.PacketStatus_Failure {
		return result
	}

	amount, ok := sdkmath.NewIntFromString(data.GetAmount())
	if !ok {
		return errRecvResult(fmt.Errorf("amount %q is not an int", data.GetAmount()), types.ErrInvalidPacket)
	}

	funds := sdk.NewCoins(sdk.NewCoin(ibcDenom, amount))

	// Execute the contract.
	execMsg := wasmtypes.MsgExecuteContract{
		Sender:   senderBech32,
		Contract: contractAddr.String(),
		Msg:      msgBytes,
		Funds:    funds,
	}
	if err := execMsg.ValidateBasic(); err != nil {
		return errRecvResult(fmt.Errorf(types.ErrBadExecutionMsg, err), types.ErrWasmError)
	}
	wasmMsgServer := wasmkeeper.NewMsgServerImpl(im.contractKeeper)
	response, err := wasmMsgServer.ExecuteContract(ctx, &execMsg)
	if err != nil {
		return errRecvResult(err, types.ErrWasmError)
	}

	fullAck := types.ContractAck{ContractResult: response.Data, IbcAck: result.Acknowledgement}
	bz, err := json.Marshal(fullAck)
	if err != nil {
		return errRecvResult(err, types.ErrBadResponse)
	}

	return channeltypesv2.RecvPacketResult{
		Status:          channeltypesv2.PacketStatus_Success,
		Acknowledgement: channeltypes.NewResultAcknowledgement(bz).Acknowledgement(),
	}
}

// OnAcknowledgementPacket implements api.IBCModule.
func (im IBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	sourceClient string,
	destinationClient string,
	sequence uint64,
	acknowledgement []byte,
	payload channeltypesv2.Payload,
	relayer sdk.AccAddress,
) error {
	if err := im.app.OnAcknowledgementPacket(ctx, sourceClient, destinationClient, sequence, acknowledgement, payload, relayer); err != nil {
		return err
	}

	if !im.properlyConfigured() {
		return nil
	}

	contract := im.ibcHooksKeeper.GetPacketCallback(ctx, sourceClient, sequence)
	if contract == "" {
		return nil
	}

	contractAddr, err := sdk.AccAddressFromBech32(contract)
	if err != nil {
		return fmt.Errorf("ack callback error: %w", err)
	}

	success := "false"
	if !ibchooks.IsJSONAckError(acknowledgement) {
		success = "true"
	}

	ackAsJSON, err := json.Marshal(acknowledgement)
	if err != nil {
		return err
	}

	sudoMsg := []byte(fmt.Sprintf(
		`{"ibc_lifecycle_complete": {"ibc_ack": {"channel": "%s", "sequence": %d, "ack": %s, "success": %s}}}`,
		sourceClient, sequence, ackAsJSON, success))
	_, err = im.contractKeeper.Sudo(ctx, contractAddr, sudoMsg)
	// Delete the callback regardless of outcome.
	im.ibcHooksKeeper.DeletePacketCallback(ctx, sourceClient, sequence)
	if err != nil {
		return fmt.Errorf("ack callback error: %w", err)
	}
	return nil
}

// OnTimeoutPacket implements api.IBCModule.
func (im IBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	sourceClient string,
	destinationClient string,
	sequence uint64,
	payload channeltypesv2.Payload,
	relayer sdk.AccAddress,
) error {
	if err := im.app.OnTimeoutPacket(ctx, sourceClient, destinationClient, sequence, payload, relayer); err != nil {
		return err
	}

	if !im.properlyConfigured() {
		return nil
	}

	contract := im.ibcHooksKeeper.GetPacketCallback(ctx, sourceClient, sequence)
	if contract == "" {
		return nil
	}

	contractAddr, err := sdk.AccAddressFromBech32(contract)
	if err != nil {
		return fmt.Errorf("timeout callback error: %w", err)
	}

	sudoMsg := []byte(fmt.Sprintf(
		`{"ibc_lifecycle_complete": {"ibc_timeout": {"channel": "%s", "sequence": %d}}}`,
		sourceClient, sequence))
	_, err = im.contractKeeper.Sudo(ctx, contractAddr, sudoMsg)
	if err != nil {
		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				"ibc-timeout-callback-error",
				sdk.NewAttribute("contract", contractAddr.String()),
				sdk.NewAttribute("message", string(sudoMsg)),
				sdk.NewAttribute("error", err.Error()),
			),
		})
	}
	im.ibcHooksKeeper.DeletePacketCallback(ctx, sourceClient, sequence)
	return nil
}

// errRecvResult creates a failure RecvPacketResult with an error acknowledgement.
func errRecvResult(err error, wrapErr error) channeltypesv2.RecvPacketResult {
	return channeltypesv2.RecvPacketResult{
		Status:          channeltypesv2.PacketStatus_Failure,
		Acknowledgement: channeltypes.NewErrorAcknowledgement(fmt.Errorf("%w: %w", wrapErr, err)).Acknowledgement(),
	}
}

// unmarshalV2PacketData extracts FungibleTokenPacketData from a v2 Payload.
func unmarshalV2PacketData(payload channeltypesv2.Payload) (transfertypes.FungibleTokenPacketData, error) {
	transfer, err := transfertypes.UnmarshalPacketData(payload.Value, payload.Version, payload.Encoding)
	if err != nil {
		return transfertypes.FungibleTokenPacketData{}, err
	}

	return transfertypes.FungibleTokenPacketData{
		Denom:    transfer.Token.Denom.Path(),
		Amount:   transfer.Token.Amount,
		Sender:   transfer.Sender,
		Receiver: transfer.Receiver,
		Memo:     transfer.Memo,
	}, nil
}

// marshalV2Payload re-encodes modified FungibleTokenPacketData back into a v2 Payload,
// preserving the original payload's ports, version, and encoding.
func marshalV2Payload(data transfertypes.FungibleTokenPacketData, original channeltypesv2.Payload) (channeltypesv2.Payload, error) {
	bz, err := transfertypes.MarshalPacketData(data, original.Version, original.Encoding)
	if err != nil {
		return channeltypesv2.Payload{}, err
	}
	return channeltypesv2.Payload{
		SourcePort:      original.SourcePort,
		DestinationPort: original.DestinationPort,
		Version:         original.Version,
		Encoding:        original.Encoding,
		Value:           bz,
	}, nil
}

// extractDenomFromV2PacketOnRecv computes the local IBC denom for a received v2 packet.
// sourceClient is the client ID on the sending chain; destinationClient is the client ID on the receiving chain.
// This mirrors the v1 logic where HasPrefix checks sourcePort/sourceChannel and the prepend uses destPort/destChannel.
func extractDenomFromV2PacketOnRecv(data transfertypes.FungibleTokenPacketData, payload channeltypesv2.Payload, sourceClient, destinationClient string) string {
	denom := transfertypes.ExtractDenomFromPath(data.Denom)
	if denom.HasPrefix(payload.SourcePort, sourceClient) {
		// Token originally came from this chain; strip the source hop.
		denom.Trace = denom.Trace[1:]
		if denom.IsNative() {
			return denom.Base
		}
		return denom.IBCDenom()
	}
	// Token came from the source chain; prepend the dest port/client hop.
	return transfertypes.NewDenom(denom.Base,
		append([]transfertypes.Hop{transfertypes.NewHop(payload.DestinationPort, destinationClient)}, denom.Trace...)...,
	).IBCDenom()
}

// addMarkerForDenom creates a marker for an IBC denom if one doesn't already exist.
func (im IBCModule) addMarkerForDenom(ctx sdk.Context, data transfertypes.FungibleTokenPacketData, ibcDenom, destinationClient string) error {
	if !im.markerHooks.ProperlyConfigured() {
		return nil
	}
	if !strings.HasPrefix(ibcDenom, "ibc/") {
		return nil
	}

	markerAddress, err := markertypes.MarkerAddress(ibcDenom)
	if err != nil {
		return err
	}
	marker, err := im.markerHooks.MarkerKeeper.GetMarker(ctx, markerAddress)
	if err != nil {
		return err
	}
	if marker != nil {
		return nil
	}

	amount, ok := sdkmath.NewIntFromString(data.Amount)
	if !ok {
		return fmt.Errorf("invalid amount %q", data.Amount)
	}
	newMarker := markertypes.NewMarkerAccount(
		authtypes.NewBaseAccountWithAddress(markertypes.MustGetMarkerAddress(ibcDenom)),
		sdk.NewCoin(ibcDenom, amount),
		nil,
		nil,
		markertypes.StatusActive,
		markertypes.MarkerType_Coin,
		false, // supply fixed
		false, // allow gov
		false, // force transfer not allowed.
		[]string{},
	)
	existingSupply := sdk.NewCoin(newMarker.Denom, im.markerHooks.MarkerKeeper.CurrentCirculation(ctx, newMarker))
	_ = newMarker.SetSupply(newMarker.GetSupply().Add(existingSupply))
	if err = im.markerHooks.MarkerKeeper.AddMarkerAccount(ctx, newMarker); err != nil {
		return err
	}

	// Look up chain ID directly from the client (v2 doesn't use channels).
	chainID := im.getChainIDFromClient(ctx, destinationClient)
	markerMetadata := banktypes.Metadata{
		Base:        ibcDenom,
		Name:        chainID + "/" + data.Denom,
		Display:     chainID + "/" + data.Denom,
		Description: data.Denom + " from " + chainID,
	}
	return im.markerHooks.MarkerKeeper.SetDenomMetaData(ctx, markerMetadata, authtypes.NewModuleAddress(types.ModuleName))
}

// getChainIDFromClient returns the chain ID for a tendermint client, or "unknown" if not available.
func (im IBCModule) getChainIDFromClient(ctx sdk.Context, clientID string) string {
	clientState, found := im.ibcKeeper.ClientKeeper.GetClientState(ctx, clientID)
	if !found {
		return "unknown"
	}
	tmClientState, ok := clientState.(*tendermintclient.ClientState)
	if ok {
		return tmClientState.ChainId
	}
	return "unknown"
}
