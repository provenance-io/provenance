package ibchooks

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v6/modules/core/exported"
)

var _ porttypes.ICS4Wrapper = &ICS4Middleware{}

type ICS4Middleware struct {
	channel porttypes.ICS4Wrapper

	// Hooks
	Hooks Hooks
}

func NewICS4Middleware(channel porttypes.ICS4Wrapper, hooks Hooks) ICS4Middleware {
	return ICS4Middleware{
		channel: channel,
		Hooks:   hooks,
	}
}

func (i ICS4Middleware) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	sourcePort string,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (sequence uint64, err error) {
	if hook, ok := i.Hooks.(SendPacketOverrideHooks); ok {
		return hook.SendPacketOverride(i, ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
	}

	if hook, ok := i.Hooks.(SendPacketBeforeHooks); ok {
		hook.SendPacketBeforeHook(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
	}

	processStateData := make(map[string]interface{})

	// Go Through all chained send function here that alter the data package and do other current chain operation
	if hook, ok := i.Hooks.(GetSendPacketFns); ok {
		fns := hook.GetSendPacketFns()
		for i := 0; i < len(fns); i++ {
			data, err = fns[i](ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data, processStateData)
			if err != nil {
				return 0, err
			}
		}
	}

	seq, err := i.channel.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)

	if hook, ok := i.Hooks.(SendPacketAfterHooks); ok {
		hook.SendPacketAfterHook(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data, seq, err, processStateData)
	}

	return seq, err
}

func (i ICS4Middleware) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet ibcexported.PacketI, ack ibcexported.Acknowledgement) error {
	if hook, ok := i.Hooks.(WriteAcknowledgementOverrideHooks); ok {
		return hook.WriteAcknowledgementOverride(i, ctx, chanCap, packet, ack)
	}

	if hook, ok := i.Hooks.(WriteAcknowledgementBeforeHooks); ok {
		hook.WriteAcknowledgementBeforeHook(ctx, chanCap, packet, ack)
	}
	err := i.channel.WriteAcknowledgement(ctx, chanCap, packet, ack)
	if hook, ok := i.Hooks.(WriteAcknowledgementAfterHooks); ok {
		hook.WriteAcknowledgementAfterHook(ctx, chanCap, packet, ack, err)
	}

	return err
}

func (i ICS4Middleware) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	if hook, ok := i.Hooks.(GetAppVersionOverrideHooks); ok {
		return hook.GetAppVersionOverride(i, ctx, portID, channelID)
	}

	if hook, ok := i.Hooks.(GetAppVersionBeforeHooks); ok {
		hook.GetAppVersionBeforeHook(ctx, portID, channelID)
	}
	version, err := i.channel.GetAppVersion(ctx, portID, channelID)
	if hook, ok := i.Hooks.(GetAppVersionAfterHooks); ok {
		hook.GetAppVersionAfterHook(ctx, portID, channelID, version, err)
	}

	return version, err
}
