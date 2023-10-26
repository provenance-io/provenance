package keeper

import (
	"encoding/json"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"

	"github.com/provenance-io/provenance/x/ibcratelimit/types"
)

var (
	msgSend = "send_packet"
	msgRecv = "recv_packet"
)

type UndoSendMsg struct {
	UndoSend UndoPacketMsg `json:"undo_send"`
}

type UndoPacketMsg struct {
	Packet UnwrappedPacket `json:"packet"`
}

type SendPacketMsg struct {
	SendPacket PacketMsg `json:"send_packet"`
}

type RecvPacketMsg struct {
	RecvPacket PacketMsg `json:"recv_packet"`
}

type PacketMsg struct {
	Packet UnwrappedPacket `json:"packet"`
}

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

func (k Keeper) CheckAndUpdateRateLimits(ctx sdk.Context, msgType string, packet exported.PacketI) error {
	contract := k.GetContractAddress(ctx)

	contractAddr, err := sdk.AccAddressFromBech32(contract)
	if err != nil {
		return errorsmod.Wrap(types.ErrContractError, err.Error())
	}

	sendPacketMsg, err := k.BuildWasmExecMsg(
		msgType,
		packet,
	)
	if err != nil {
		return errorsmod.Wrap(types.ErrContractError, err.Error())
	}

	_, err = k.ContractKeeper.Sudo(ctx, contractAddr, sendPacketMsg)

	if err != nil {
		return errorsmod.Wrap(types.ErrRateLimitExceeded, err.Error())
	}

	return nil
}

func (k Keeper) UndoSendRateLimit(ctx sdk.Context, contract string, packet exported.PacketI) error {
	contractAddr, err := sdk.AccAddressFromBech32(contract)
	if err != nil {
		return err
	}

	unwrapped, err := k.unwrapPacket(packet)
	if err != nil {
		return err
	}

	msg := UndoSendMsg{UndoSend: UndoPacketMsg{Packet: unwrapped}}
	asJSON, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = k.ContractKeeper.Sudo(ctx, contractAddr, asJSON)
	if err != nil {
		return errorsmod.Wrap(types.ErrContractError, err.Error())
	}

	return nil
}

func (k Keeper) unwrapPacket(packet exported.PacketI) (UnwrappedPacket, error) {
	var packetData transfertypes.FungibleTokenPacketData
	err := json.Unmarshal(packet.GetData(), &packetData)
	if err != nil {
		return UnwrappedPacket{}, err
	}
	height, ok := packet.GetTimeoutHeight().(clienttypes.Height)
	if !ok {
		return UnwrappedPacket{}, types.ErrBadMessage
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

func (k Keeper) BuildWasmExecMsg(msgType string, packet exported.PacketI) ([]byte, error) {
	unwrapped, err := k.unwrapPacket(packet)
	if err != nil {
		return []byte{}, err
	}

	var asJSON []byte
	switch {
	case msgType == msgSend:
		msg := SendPacketMsg{SendPacket: PacketMsg{
			Packet: unwrapped,
		}}
		asJSON, err = json.Marshal(msg)
	case msgType == msgRecv:
		msg := RecvPacketMsg{RecvPacket: PacketMsg{
			Packet: unwrapped,
		}}
		asJSON, err = json.Marshal(msg)
	default:
		return []byte{}, types.ErrBadMessage
	}

	if err != nil {
		return []byte{}, err
	}

	return asJSON, nil
}

// RevertSentPacket Notifies the contract that a sent packet wasn't properly received
func (k Keeper) RevertSentPacket(
	ctx sdk.Context,
	packet exported.PacketI,
) error {
	contract := k.GetContractAddress(ctx)
	if contract == "" {
		// The contract has not been configured. Continue as usual
		return nil
	}

	return k.UndoSendRateLimit(
		ctx,
		contract,
		packet,
	)
}

func (k Keeper) ValidateReceiverAddress(packet exported.PacketI) error {
	var packetData transfertypes.FungibleTokenPacketData
	if err := json.Unmarshal(packet.GetData(), &packetData); err != nil {
		return err
	}
	if len(packetData.Receiver) >= 4096 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "IBC Receiver address too long. Max supported length is %d", 4096)
	}
	return nil
}
