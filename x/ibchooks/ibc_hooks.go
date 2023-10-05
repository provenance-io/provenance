package ibchooks

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v6/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v6/modules/core/keeper"

	"github.com/provenance-io/provenance/x/ibchooks/keeper"
	"github.com/provenance-io/provenance/x/ibchooks/types"
)

type IbcHooks struct {
	cdc                            codec.BinaryCodec
	ibcKeeper                      *ibckeeper.Keeper
	ibcHooksKeeper                 *keeper.Keeper
	wasmHooks                      *WasmHooks
	markerHooks                    *MarkerHooks
	PreSendPacketDataProcessingFns []types.PreSendPacketDataProcessingFn
}

func NewIbcHooks(cdc codec.BinaryCodec, ibcHooksKeeper *keeper.Keeper, ibcKeeper *ibckeeper.Keeper, wasmHooks *WasmHooks, markerHooks *MarkerHooks, preSendPacketDataProcessingFns []types.PreSendPacketDataProcessingFn) IbcHooks {
	return IbcHooks{
		cdc:                            cdc,
		ibcKeeper:                      ibcKeeper,
		ibcHooksKeeper:                 ibcHooksKeeper,
		wasmHooks:                      wasmHooks,
		markerHooks:                    markerHooks,
		PreSendPacketDataProcessingFns: preSendPacketDataProcessingFns,
	}
}

// ProperlyConfigured returns false if either wasm or marker hooks are configured properly
func (h IbcHooks) ProperlyConfigured() bool {
	return h.wasmHooks.ProperlyConfigured() && h.markerHooks.ProperlyConfigured() && h.ibcHooksKeeper != nil
}

func (h IbcHooks) GetPreSendPacketDataProcessingFns() []types.PreSendPacketDataProcessingFn {
	return h.PreSendPacketDataProcessingFns
}

// OnRecvPacketOverride executes wasm or marker hooks for Ics20 packets, if not ics20 packet it will continue to process packet with no override
func (h IbcHooks) OnRecvPacketOverride(im IBCMiddleware, ctx sdktypes.Context, packet channeltypes.Packet, relayer sdktypes.AccAddress) ibcexported.Acknowledgement {
	if !h.ProperlyConfigured() {
		return im.App.OnRecvPacket(ctx, packet, relayer)
	}

	isIcs20, _ := isIcs20Packet(packet.GetData())
	if !isIcs20 {
		return im.App.OnRecvPacket(ctx, packet, relayer)
	}

	if err := h.markerHooks.AddUpdateMarker(ctx, packet, h.ibcKeeper); err != nil {
		return NewEmitErrorAcknowledgement(ctx, types.ErrMarkerError, err.Error())
	}
	return h.wasmHooks.OnRecvPacketOverride(im, ctx, packet, relayer)
}

func (h IbcHooks) SendPacketAfterHook(ctx sdktypes.Context,
	chanCap *capabilitytypes.Capability,
	sourcePort string,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
	sequence uint64,
	err error,
	processData map[string]interface{},
) {
	h.wasmHooks.SendPacketAfterHook(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data, sequence, err, processData)
}

func (h IbcHooks) OnTimeoutPacketOverride(im IBCMiddleware, ctx sdktypes.Context, packet channeltypes.Packet, relayer sdktypes.AccAddress) error {
	return h.wasmHooks.OnTimeoutPacketOverride(im, ctx, packet, relayer)
}

func (h IbcHooks) OnAcknowledgementPacketOverride(im IBCMiddleware, ctx sdktypes.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdktypes.AccAddress) error {
	return h.wasmHooks.OnAcknowledgementPacketOverride(im, ctx, packet, acknowledgement, relayer)
}
