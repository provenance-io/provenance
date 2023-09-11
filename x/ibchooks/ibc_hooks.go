package ibchooks

import (
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v6/modules/core/exported"

	"github.com/provenance-io/provenance/x/ibchooks/keeper"
	"github.com/provenance-io/provenance/x/ibchooks/types"
)

type IbcHooks struct {
	ibcHooksKeeper *keeper.Keeper
	wasmHooks      *WasmHooks
	markerHooks    *MarkerHooks
}

func NewIbcHooks(ibcHooksKeeper *keeper.Keeper, wasmHooks *WasmHooks, markerHooks *MarkerHooks) IbcHooks {
	return IbcHooks{
		ibcHooksKeeper: ibcHooksKeeper,
		wasmHooks:      wasmHooks,
		markerHooks:    markerHooks,
	}
}

// ProperlyConfigured returns false if either wasm or marker hooks are configured properly
func (h IbcHooks) ProperlyConfigured() bool {
	return h.wasmHooks.ProperlyConfigured() && h.markerHooks.ProperlyConfigured()
}

// OnRecvPacketOverride executes wasm or marker hooks for Ics20 packets, if not ics20 packet it will continue to process packet with no override
func (h IbcHooks) OnRecvPacketOverride(im IBCMiddleware, ctx sdktypes.Context, packet channeltypes.Packet, relayer sdktypes.AccAddress) ibcexported.Acknowledgement {
	if !h.ProperlyConfigured() {
		return im.App.OnRecvPacket(ctx, packet, relayer)
	}

	isIcs20, data := isIcs20Packet(packet.GetData())
	if !isIcs20 {
		return im.App.OnRecvPacket(ctx, packet, relayer)
	}

	isWasmRouted, _ := jsonStringHasKey(data.GetMemo(), "wasm")
	if isWasmRouted {
		return h.wasmHooks.OnRecvPacketOverride(im, ctx, packet, relayer)
	}
	return h.markerHooks.OnRecvPacketOverride(im, ctx, packet, relayer)
}

func (h IbcHooks) SendPacketOverride(
	i ICS4Middleware,
	ctx sdktypes.Context,
	chanCap *capabilitytypes.Capability,
	sourcePort string,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (uint64, error) {
	isIcs20, ics20Packet := isIcs20Packet(data)
	if !isIcs20 {
		return i.channel.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
	}

	isCallbackRouted, _ := jsonStringHasKey(ics20Packet.GetMemo(), types.IBCCallbackKey)
	if isCallbackRouted {
		return h.wasmHooks.SendPacketOverride(i, ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
	}

	// find marker, create memo struct, send ...

	return i.channel.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
}

func (h IbcHooks) OnTimeoutPacketOverride(im IBCMiddleware, ctx sdktypes.Context, packet channeltypes.Packet, relayer sdktypes.AccAddress) error {
	return h.wasmHooks.OnTimeoutPacketOverride(im, ctx, packet, relayer)
}

func (h IbcHooks) OnAcknowledgementPacketOverride(im IBCMiddleware, ctx sdktypes.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdktypes.AccAddress) error {
	return h.wasmHooks.OnAcknowledgementPacketOverride(im, ctx, packet, acknowledgement, relayer)
}
