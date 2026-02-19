package ibcratelimit

import (
	"encoding/json"

	errorsmod "cosmossdk.io/errors"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
)

const (
	// MsgSendPacket is the operation used for tracking a sent packet.
	MsgSendPacket = "send_packet"
	// MsgRecvPacket is the operation used for tracking a received packet.
	MsgRecvPacket = "recv_packet"
)

// UndoSendMsg is an ibcratelimit contract message meant to undo tracked sends.
type UndoSendMsg struct {
	UndoSend UndoPacketMsg `json:"undo_send"`
}

// UndoPacketMsg is an operation done by the UndoSendMsg.
type UndoPacketMsg struct {
	Packet UnwrappedPacket `json:"packet"`
}

// SendPacketMsg is an ibcratelimit contract message meant to track sends.
type SendPacketMsg struct {
	SendPacket PacketMsg `json:"send_packet"`
}

// RecvPacketMsg is an ibcratelimit contract message meant to track receives.
type RecvPacketMsg struct {
	RecvPacket PacketMsg `json:"recv_packet"`
}

// PacketMsg contains
type PacketMsg struct {
	Packet UnwrappedPacket `json:"packet"`
}

// UnwrappedPacket is a FungibleTokenPacket.
type UnwrappedPacket struct {
	Sequence           uint64                                `json:"sequence"`
	SourcePort         string                                `json:"source_port"`
	SourceChannel      string                                `json:"source_channel"`
	DestinationPort    string                                `json:"destination_port"`
	DestinationChannel string                                `json:"destination_channel"`
	Data               transfertypes.FungibleTokenPacketData `json:"data"`
	TimeoutHeight      clienttypes.Height                    `json:"timeout_height"`
	TimeoutTimestamp   uint64                                `json:"timeout_timestamp,omitempty"`
}

// ValidateReceiverAddress Checks if the receiver is valid for the transfer data.
func ValidateReceiverAddress(packet exported.PacketI) error {
	var packetData transfertypes.FungibleTokenPacketData
	if err := json.Unmarshal(packet.GetData(), &packetData); err != nil {
		return err
	}
	if len(packetData.Receiver) >= 4096 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "IBC Receiver address too long. Max supported length is %d", 4096)
	}
	return nil
}

// UnwrapPacket Converts a PacketI into an UnwrappedPacket structure.
func UnwrapPacket(packet exported.PacketI) (UnwrappedPacket, error) {
	if packet == nil {
		return UnwrappedPacket{}, ErrBadMessage
	}
	var packetData transfertypes.FungibleTokenPacketData
	err := json.Unmarshal(packet.GetData(), &packetData)
	if err != nil {
		return UnwrappedPacket{}, err
	}
	height, ok := packet.GetTimeoutHeight().(clienttypes.Height)
	if !ok {
		return UnwrappedPacket{}, ErrBadMessage
	}
	return UnwrappedPacket{
		Sequence:           packet.GetSequence(),
		SourcePort:         packet.GetSourcePort(),
		SourceChannel:      packet.GetSourceChannel(),
		DestinationPort:    packet.GetDestPort(),
		DestinationChannel: packet.GetDestChannel(),
		Data:               packetData,
		TimeoutHeight:      height,
		TimeoutTimestamp:   packet.GetTimeoutTimestamp(),
	}, nil
}
