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

	msg := types.UndoSendMsg{UndoSend: types.UndoPacketMsg{Packet: unwrapped}}
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

func (k Keeper) unwrapPacket(packet exported.PacketI) (types.UnwrappedPacket, error) {
	var packetData transfertypes.FungibleTokenPacketData
	err := json.Unmarshal(packet.GetData(), &packetData)
	if err != nil {
		return types.UnwrappedPacket{}, err
	}
	height, ok := packet.GetTimeoutHeight().(clienttypes.Height)
	if !ok {
		return types.UnwrappedPacket{}, types.ErrBadMessage
	}
	return types.UnwrappedPacket{
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
	case msgType == types.MsgSendPacket:
		msg := types.SendPacketMsg{SendPacket: types.PacketMsg{
			Packet: unwrapped,
		}}
		asJSON, err = json.Marshal(msg)
	case msgType == types.MsgRecvPacket:
		msg := types.RecvPacketMsg{RecvPacket: types.PacketMsg{
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
	if !k.ContractConfigured(ctx) {
		// The contract has not been configured. Continue as usual
		return nil
	}

	contract := k.GetContractAddress(ctx)

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
