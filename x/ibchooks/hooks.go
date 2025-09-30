// Package ibchooks provides hooks for the IBC module lifecycle events.
package ibchooks

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	"github.com/provenance-io/provenance/x/ibchooks/types"
)

// Hooks defines the set of hooks for IBC lifecycle events.
type Hooks interface{}

// SendPacketPreProcessors returns a list of ordered functions to be executed before ibc's SendPacket function in middleware
type SendPacketPreProcessors interface {
	GetSendPacketPreProcessors() []types.PreSendPacketDataProcessingFn
}

// OnChanOpenInitOverrideHooks defines hooks for overriding OnChanOpenInit.
type OnChanOpenInitOverrideHooks interface {
	OnChanOpenInitOverride(im IBCMiddleware, ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID string, channelID string, channelCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, version string) (string, error)
}

// OnChanOpenInitBeforeHooks defines hooks for actions before OnChanOpenInit.
type OnChanOpenInitBeforeHooks interface {
	OnChanOpenInitBeforeHook(ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID string, channelID string, channelCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, version string)
}

// OnChanOpenInitAfterHooks defines hooks for actions after OnChanOpenInit.
type OnChanOpenInitAfterHooks interface {
	OnChanOpenInitAfterHook(ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID string, channelID string, channelCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, version string, finalVersion string, err error)
}

// OnChanOpenTryOverrideHooks defines hooks for overriding OnChanOpenTry.
type OnChanOpenTryOverrideHooks interface {
	OnChanOpenTryOverride(im IBCMiddleware, ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID, channelID string, channelCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, counterpartyVersion string) (string, error)
}

// OnChanOpenTryBeforeHooks defines hooks for actions before OnChanOpenTry
type OnChanOpenTryBeforeHooks interface {
	OnChanOpenTryBeforeHook(ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID, channelID string, channelCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, counterpartyVersion string)
}

// OnChanOpenTryAfterHooks defines hooks for actions after OnChanOpenTry.
type OnChanOpenTryAfterHooks interface {
	OnChanOpenTryAfterHook(ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID, channelID string, channelCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, counterpartyVersion string, version string, err error)
}

// OnChanOpenAckOverrideHooks defines hooks for overriding OnChanOpenAck.
type OnChanOpenAckOverrideHooks interface {
	OnChanOpenAckOverride(im IBCMiddleware, ctx sdk.Context, portID, channelID string, counterpartyChannelID string, counterpartyVersion string) error
}

// OnChanOpenAckBeforeHooks defines hooks for actions before OnChanOpenAck.
type OnChanOpenAckBeforeHooks interface {
	OnChanOpenAckBeforeHook(ctx sdk.Context, portID, channelID string, counterpartyChannelID string, counterpartyVersion string)
}

// OnChanOpenAckAfterHooks defines hooks for actions after OnChanOpenAck.
type OnChanOpenAckAfterHooks interface {
	OnChanOpenAckAfterHook(ctx sdk.Context, portID, channelID string, counterpartyChannelID string, counterpartyVersion string, err error)
}

// OnChanOpenConfirmOverrideHooks defines hooks for overriding OnChanOpenConfirm.
type OnChanOpenConfirmOverrideHooks interface {
	OnChanOpenConfirmOverride(im IBCMiddleware, ctx sdk.Context, portID, channelID string) error
}

// OnChanOpenConfirmBeforeHooks defines hooks for actions before OnChanOpenConfirm.
type OnChanOpenConfirmBeforeHooks interface {
	OnChanOpenConfirmBeforeHook(ctx sdk.Context, portID, channelID string)
}

// OnChanOpenConfirmAfterHooks defines hooks for actions after OnChanOpenConfirm.
type OnChanOpenConfirmAfterHooks interface {
	OnChanOpenConfirmAfterHook(ctx sdk.Context, portID, channelID string, err error)
}

// OnChanCloseInitOverrideHooks defines hooks for overriding OnChanCloseInit.
type OnChanCloseInitOverrideHooks interface {
	OnChanCloseInitOverride(im IBCMiddleware, ctx sdk.Context, portID, channelID string) error
}

// OnChanCloseInitBeforeHooks defines hooks for actions before OnChanCloseInit.
type OnChanCloseInitBeforeHooks interface {
	OnChanCloseInitBeforeHook(ctx sdk.Context, portID, channelID string)
}

// OnChanCloseInitAfterHooks defines hooks for actions after OnChanCloseInit.
type OnChanCloseInitAfterHooks interface {
	OnChanCloseInitAfterHook(ctx sdk.Context, portID, channelID string, err error)
}

// OnChanCloseConfirmOverrideHooks defines hooks for overriding OnChanCloseConfirm.
type OnChanCloseConfirmOverrideHooks interface {
	OnChanCloseConfirmOverride(im IBCMiddleware, ctx sdk.Context, portID, channelID string) error
}

// OnChanCloseConfirmBeforeHooks defines hooks for actions before OnChanCloseConfirm.
type OnChanCloseConfirmBeforeHooks interface {
	OnChanCloseConfirmBeforeHook(ctx sdk.Context, portID, channelID string)
}

// OnChanCloseConfirmAfterHooks defines hooks for actions after OnChanCloseConfirm.
type OnChanCloseConfirmAfterHooks interface {
	OnChanCloseConfirmAfterHook(ctx sdk.Context, portID, channelID string, err error)
}

// OnRecvPacketOverrideHooks defines hooks for overriding OnRecvPacket.
type OnRecvPacketOverrideHooks interface {
	OnRecvPacketOverride(im IBCMiddleware, ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) ibcexported.Acknowledgement
}

// OnRecvPacketBeforeHooks defines hooks for actions before OnRecvPacket.
type OnRecvPacketBeforeHooks interface {
	OnRecvPacketBeforeHook(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress)
}

// OnRecvPacketAfterHooks defines hooks for actions after OnRecvPacket.
type OnRecvPacketAfterHooks interface {
	OnRecvPacketAfterHook(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress, ack ibcexported.Acknowledgement)
}

// OnAcknowledgementPacketOverrideHooks defines hooks for overriding OnAcknowledgementPacket.
type OnAcknowledgementPacketOverrideHooks interface {
	OnAcknowledgementPacketOverride(im IBCMiddleware, ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress) error
}

// OnAcknowledgementPacketBeforeHooks defines hooks for actions before OnAcknowledgementPacket.
type OnAcknowledgementPacketBeforeHooks interface {
	OnAcknowledgementPacketBeforeHook(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress)
}

// OnAcknowledgementPacketAfterHooks defines hooks for actions after OnAcknowledgementPacket.
type OnAcknowledgementPacketAfterHooks interface {
	OnAcknowledgementPacketAfterHook(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress, err error)
}

// OnTimeoutPacketOverrideHooks defines hooks for overriding OnTimeoutPacket.
type OnTimeoutPacketOverrideHooks interface {
	OnTimeoutPacketOverride(im IBCMiddleware, ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error
}

// OnTimeoutPacketBeforeHooks defines hooks for actions before OnTimeoutPacket.
type OnTimeoutPacketBeforeHooks interface {
	OnTimeoutPacketBeforeHook(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress)
}

// OnTimeoutPacketAfterHooks defines hooks for actions after OnTimeoutPacket.
type OnTimeoutPacketAfterHooks interface {
	OnTimeoutPacketAfterHook(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress, err error)
}

// SendPacketOverrideHooks defines hooks for overriding SendPacket.
type SendPacketOverrideHooks interface {
	SendPacketOverride(
		i ICS4Middleware,
		ctx sdk.Context,
		chanCap *capabilitytypes.Capability,
		sourcePort string,
		sourceChannel string,
		timeoutHeight clienttypes.Height,
		timeoutTimestamp uint64,
		data []byte,
	) (sequence uint64, err error)
}

// SendPacketBeforeHooks defines hooks for actions before SendPacket.
type SendPacketBeforeHooks interface {
	SendPacketBeforeHook(ctx sdk.Context,
		chanCap *capabilitytypes.Capability,
		sourcePort string,
		sourceChannel string,
		timeoutHeight clienttypes.Height,
		timeoutTimestamp uint64,
		data []byte,
	)
}

// SendPacketAfterHooks defines hooks for actions after SendPacket.
type SendPacketAfterHooks interface {
	SendPacketAfterHook(ctx sdk.Context,
		chanCap *capabilitytypes.Capability,
		sourcePort string,
		sourceChannel string,
		timeoutHeight clienttypes.Height,
		timeoutTimestamp uint64,
		data []byte,
		sequence uint64,
		err error,
		processData map[string]interface{},
	)
}

// WriteAcknowledgementOverrideHooks defines hooks for overriding WriteAcknowledgement.
type WriteAcknowledgementOverrideHooks interface {
	WriteAcknowledgementOverride(i ICS4Middleware, ctx sdk.Context, chanCap *capabilitytypes.Capability, packet ibcexported.PacketI, ack ibcexported.Acknowledgement) error
}

// WriteAcknowledgementBeforeHooks defines hooks for actions before WriteAcknowledgement.
type WriteAcknowledgementBeforeHooks interface {
	WriteAcknowledgementBeforeHook(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet ibcexported.PacketI, ack ibcexported.Acknowledgement)
}

// WriteAcknowledgementAfterHooks defines hooks for actions after WriteAcknowledgement.
type WriteAcknowledgementAfterHooks interface {
	WriteAcknowledgementAfterHook(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet ibcexported.PacketI, ack ibcexported.Acknowledgement, err error)
}

// GetAppVersionOverrideHooks defines hooks for overriding GetAppVersion.
type GetAppVersionOverrideHooks interface {
	GetAppVersionOverride(i ICS4Middleware, ctx sdk.Context, portID, channelID string) (string, bool)
}

// GetAppVersionBeforeHooks defines hooks for actions before GetAppVersion.
type GetAppVersionBeforeHooks interface {
	GetAppVersionBeforeHook(ctx sdk.Context, portID, channelID string)
}

// GetAppVersionAfterHooks defines hooks for actions before GetAppVersion.
type GetAppVersionAfterHooks interface {
	GetAppVersionAfterHook(ctx sdk.Context, portID, channelID string, result string, success bool)
}
